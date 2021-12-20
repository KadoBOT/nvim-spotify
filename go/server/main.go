package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
	logger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/KadoBOT/spotify/v2"
	spotifyauth "github.com/KadoBOT/spotify/v2/auth"
)

const BASE_URL = "https://api.spotify.com/v1"
const REDIRECT_URI = "http://192.168.178.101:3000/callback"
const SCOPE = "user-read-playback-state user-read-currently-playing user-modify-playback-state streaming user-read-playback-position user-library-modify"

var (
	state  = uuid.New().String()
	client = &spotify.Client{}
	auth   = &spotifyauth.Authenticator{}
)

func NewAuth() {
	auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(REDIRECT_URI),
		spotifyauth.WithScopes(SCOPE),
		spotifyauth.WithClientID(os.Getenv("SPOTIFY_ID")),
		spotifyauth.WithClientSecret(os.Getenv("SPOTIFY_SECRET")),
	)
}

func configAuthRoutes(app *fiber.App) {
	NewAuth()

	login := func(c *fiber.Ctx) error {
		c.Redirect(auth.AuthURL(state))
		return nil
	}

	callback := func(c *fiber.Ctx) error {
		q := spotifyauth.Query{
			Error: c.Query("error"),
			Code:  c.Query("code"),
			State: c.Query("state"),
		}
		if st := c.Query("state"); st != state {
			log.Fatalf("state mismatch")
			c.Context().NotFound()
		}

		token, err := auth.Token(c.Context(), state, q, oauth2.AccessTypeOffline)
		if err != nil {
			log.Fatalf("failed to retrieve token")
			c.Context().Error("Couldn't get token", http.StatusForbidden)
		}

		c.JSON(token)

		return nil
	}

	app.Get("/", login)
	app.Get("/callback", callback)
}

func configAppRoutes(app *fiber.App) {
	log.Printf("configuring app routes")

	app.Use(func(c *fiber.Ctx) error {
		log.Printf("client middleware")

		refresh_token := c.Get("refresh_token")
		if refresh_token == "" {
			c.Context().Error("Please login", 403)
			return nil
		}
		token := &oauth2.Token{RefreshToken: refresh_token}
		client = spotify.New(auth.Client(c.Context(), token))

		return c.Next()
	})

	search := func(c *fiber.Ctx) error {
		query := c.Params("query")
		query, _ = url.QueryUnescape(query)

		t := c.Params("type")
		limit, _ := strconv.Atoi(c.Query("limit", "20"))

		var searchType spotify.SearchType
		switch strings.ToLower(t) {
		case "album":
			searchType = spotify.SearchTypeAlbum
		case "artist":
			searchType = spotify.SearchTypeArtist
		case "playlist":
			searchType = spotify.SearchTypePlaylist
		default:
			searchType = spotify.SearchTypeTrack
		}

		res, err := client.Search(c.Context(), query, searchType, spotify.Limit(limit))
		if err != nil {
			c.Context().Error(err.Error(), fiber.StatusBadRequest)
			return err
		}

		c.JSON(res)
		return nil
	}

	play := func(c *fiber.Ctx) error {
		uri := c.Params("uri")
		client.PlayOpt(c.Context(), &spotify.PlayOptions{
			PlaybackContext: (*spotify.URI)(&uri),
		})
		return nil
	}

	currentlyPlaying := func(c *fiber.Ctx) error {
		res, err := client.PlayerCurrentlyPlaying(c.Context())
		if err != nil {
			c.Context().Error(err.Error(), fiber.StatusBadRequest)
			return err
		}

		c.JSON(res)
		return nil
	}

	app.Get("/search/:type/:query", search)
	app.Get("/currently-playing", currentlyPlaying)
	app.Get("/play/:uri", play)
}

func configServer(app *fiber.App) {
	log.Println("creating server")
	l, _ := os.Create("/tmp/nvim-spotify-server.log")
	log.SetOutput(l)

	log.Println("configuring logger")
	app.Use(logger.New(logger.Config{Output: l}))
}

func main() {
	os.Setenv("SPOTIFY_SECRET", "2d8fbdee1aa2445aa48758dff38f50c4")
	os.Setenv("SPOTIFY_ID", "29a97ceeeb7d488fb045b7ed60e3e0c7")

	app := fiber.New()
	configServer(app)

	log.Println("configuring routes")
	configAuthRoutes(app)
	configAppRoutes(app)

	log.Fatal(app.Listen(":3000"))
}
