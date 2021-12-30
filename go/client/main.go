package client

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
	oauthspotify "golang.org/x/oauth2/spotify"
	"gopkg.in/yaml.v3"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    int    `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type Client struct {
	ClientId     string `yml:"client_id"`
	ClientSecret string `yml:"client_secret"`
	DeviceId     string `yml:"device_id"`
}

func NewClient() *spotify.Client {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Println("cannot read home dir")
	}

	folder := filepath.Join(home, ".config", "spotify-tui")
	clientFile, err := os.ReadFile(filepath.Join(folder, "client.yml"))
	if err != nil {
		log.Println("cannot read client file")
	}

	var clientInfo Client
	if err := yaml.Unmarshal(clientFile, &clientInfo); err != nil {
		log.Println("cannot unmarshal client")
	}

	tokenFile, err := os.ReadFile(filepath.Join(folder, ".spotify_token_cache.json"))
	if err != nil {
		log.Println("cannot read token file")
	}

	var token Token
	if err := json.Unmarshal(tokenFile, &token); err != nil {
		log.Println("cannot unmarshal token")
	}

	cfg := oauth2.Config{
		ClientID:     clientInfo.ClientId,
		ClientSecret: clientInfo.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthspotify.Endpoint.AuthURL,
			TokenURL: oauthspotify.Endpoint.TokenURL,
		},
		Scopes: strings.Split(token.Scope, " "),
	}

	oauthToken := oauth2.Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       time.Unix(int64(token.ExpiresAt), 0),
	}

	ctx := context.Background()
	tokenSource := cfg.TokenSource(ctx, &oauthToken)
	oauthClient := oauth2.NewClient(ctx, tokenSource)

	client := spotify.New(oauthClient)
	return client
}
