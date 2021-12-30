package command

import (
	"context"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/kadobot/nvim-spotify/utils"
	"github.com/neovim/go-client/nvim"
	"github.com/zmb3/spotify/v2"
)

func (p *Command) GetCurrentlyPlayingTrack() error {
	log.Println("cur playing")
	ctx := context.Background()
	curPlaying, err := p.client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		log.Println(err.Error())
	}

	if curPlaying.Playing {
		log.Printf("Creating CurrentlyPlaying")
		buf, err := p.CreateBuffer(false, true)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}

		fullName := fmt.Sprintf("%s by %s", curPlaying.Item.Name, utils.FormatArtistsName(curPlaying.Item.Artists))
		playingName := utils.SafeString(fullName, WIDTH-10)

		log.Println(playingName)

		top_border := []byte("╭" + strings.Repeat("─", (WIDTH-19)/2) + " Currently Playing " + strings.Repeat("─", (WIDTH-21)/2) + "╮")
		empty_line := []byte("│ 墳" + strings.Repeat(" ", WIDTH-5) + "│")
		bot_border := []byte("╰" + strings.Repeat("─", WIDTH-2) + "╯")

		replacement := [][]byte{top_border, empty_line, bot_border}

		opts := nvim.WindowConfig{
			Relative:  "win",
			Win:       *p.anchor,
			Width:     WIDTH,
			Height:    HEIGHT,
			BufPos:    [2]int{0, 0},
			Row:       -3,
			Col:       -2,
			Style:     "minimal",
			ZIndex:    1,
			Focusable: false,
		}

		if err := p.SetBufferLines(buf, 0, -1, true, replacement); err != nil {
			log.Fatalf(err.Error())
			return err
		}

		p.SetBufferText(buf, 1, 7, 1, utf8.RuneCountInString(playingName)+7, [][]byte{[]byte(playingName)})
		p.SetBufferOption(buf, "modifiable", false)
		p.SetBufferOption(buf, "bufhidden", "wipe")
		p.SetBufferOption(buf, "buftype", "nofile")

		win, err := p.OpenWindow(buf, false, &opts)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		p.wins[&win] = true

		p.SetWindowOption(win, "winhl", "Normal:SpotifyBorder")
		p.SetWindowOption(win, "winblend", 0)
		p.SetWindowOption(win, "foldlevel", 100)
	}

	return nil
}

func (p *Command) setKeyMaps(keys [][3]string) {
	log.Printf("Setting Keymaps")

	opts := map[string]bool{"noremap": true, "silent": true, "nowait": true}
	for _, k := range keys {
		p.SetBufferKeyMap(*p.Buffer, k[0], k[1], k[2], opts)
	}
}

func (p *Command) SearchFn(args []string) {
	searchType := args[0]
	input := args[1]
	log.Println("searchtype: ", searchType)
	log.Println("searchinput: ", input)
	p.SetVar("spotify_type", searchType)

	var spotifySearch [][4]string
	ctx := context.Background()
	switch searchType {
	case "tracks":
		searchResult, _ := p.client.Search(ctx, input, spotify.SearchTypeTrack)
		for _, l := range searchResult.Tracks.Tracks {
			artistName := utils.FormatArtistsName(l.Artists)
			spotifySearch = append(spotifySearch, [4]string{l.Name, artistName, string(l.URI), l.ID.String()})
		}
	case "playlists":
		searchResult, _ := p.client.Search(ctx, input, spotify.SearchTypePlaylist)
		for _, l := range searchResult.Playlists.Playlists {
			spotifySearch = append(spotifySearch, [4]string{l.Name, l.Owner.DisplayName, string(l.URI), l.ID.String()})
		}
	case "artists":
		searchResult, _ := p.client.Search(ctx, input, spotify.SearchTypeArtist)
		for _, l := range searchResult.Artists.Artists {
			spotifySearch = append(spotifySearch, [4]string{l.Name, "", string(l.URI), l.ID.String()})
		}
	case "albums":
		searchResult, _ := p.client.Search(ctx, input, spotify.SearchTypeAlbum)
		for _, l := range searchResult.Albums.Albums {
			artistName := utils.FormatArtistsName(l.Artists)
			spotifySearch = append(spotifySearch, [4]string{l.Name, artistName, string(l.URI), l.ID.String()})
		}
	case "artistsalbums":
		p.SetVar("spotify_type", "albums")

		artistAlbums, err := p.client.GetArtistAlbums(ctx, spotify.ID(input), []spotify.AlbumType{spotify.AlbumTypeAlbum})
		if err != nil {
			log.Println(err.Error())
		}
		for _, l := range artistAlbums.Albums {
			artistName := utils.FormatArtistsName(l.Artists)
			spotifySearch = append(spotifySearch, [4]string{l.Name, artistName, string(l.URI), l.ID.String()})
		}
	case "playliststracks":
		p.SetVar("spotify_type", "tracks")
		playlistTracks, err := p.client.GetPlaylistTracks(ctx, spotify.ID(input))
		if err != nil {
			log.Println(err.Error())
		}
		for _, l := range playlistTracks.Tracks {
			artistName := utils.FormatArtistsName(l.Track.Artists)
			spotifySearch = append(spotifySearch, [4]string{l.Track.Name, artistName, string(l.Track.URI), l.Track.ID.String()})
		}
	case "albumstracks":
		p.SetVar("spotify_type", "tracks")
		albumTracks, err := p.client.GetAlbumTracks(ctx, spotify.ID(input))
		if err != nil {
			log.Println(err.Error())
		}
		for _, l := range albumTracks.Tracks {
			artistName := utils.FormatArtistsName(l.Artists)
			spotifySearch = append(spotifySearch, [4]string{l.Name, artistName, string(l.URI), l.ID.String()})
		}
	}

	p.SetVar("spotify_search", spotifySearch)
	p.Command("lua require'nvim-spotify'.init()")
}

func (p *Command) Search(args []string) {
	log.Printf("starting search...")
	searchType := args[0]
	b, err := p.CurrentLine()
	if err != nil {
		log.Fatalf("Input cannot be empty")
	}
	input := string(b)
	p.SetVar("spotify_title", input)

	p.SearchFn([]string{searchType, input})
}

func (p *Command) Play(args []string) {
	var selected []string
	p.Var("spotify_device", &selected)
	log.Println("selected device: ", selected)
	if len(selected) != 0 {
		utils.ExecCommand("spt", "play", "-u", args[0], "-d", selected[0])
	} else {
		utils.ExecCommand("spt", "play", "-u", args[0])
	}
}

func (p *Command) Playback(args []string) {
	switch args[0] {
	case "next":
		utils.ExecCommand("spt", "playback", "--next")
	case "pause":
		utils.ExecCommand("spt", "playback", "--toggle")
	}
}

func (p *Command) Save() {
	utils.ExecCommand("spt", "playback", "--like")
}

func (p *Command) GetDevices() ([]string, bool) {
	log.Println("getting devices")
	res, ok := utils.ExecCommand("spt", "list", "-d")

	if ok {
		devices := strings.Split(res, "\n")
		return devices, ok
	}

	return nil, ok
}
