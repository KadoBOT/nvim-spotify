package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/KadoBOT/spotify/v2"
	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

const WIDTH = 70
const HEIGHT = 3

type Command struct {
	*nvim.Nvim
	*nvim.Buffer
	wins  map[*nvim.Window]bool
	input string
}

func (p *Command) call(url string) *http.Response {
	log.Printf(url)
	refreshToken := p.getRefreshToken()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf(err.Error())
	}
	req.Header.Add("refresh_token", refreshToken)

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func NewCommand(v *nvim.Nvim) *Command {
	return &Command{Nvim: v, wins: make(map[*nvim.Window]bool)}
}

func (p *Command) getRefreshToken() string {
	var refreshToken string
	log.Printf("getting refresh token")
	if err := p.Nvim.Var("spotify_refresh_token", &refreshToken); err != nil {
		log.Fatalf("cannot get refreshToken %s", err.Error())
	}
	log.Printf(refreshToken)
	return refreshToken
}

func (p *Command) createPlaceholder() error {
	log.Printf("Creating Placeholder")
	buf, err := p.CreateBuffer(false, true)
	if err != nil {
		log.Fatalf(err.Error())
		return err
	}

	top_border := []byte("╭" + strings.Repeat("─", (WIDTH-17)/2) + " SpotifySearch " + strings.Repeat("─", (WIDTH-16)/2) + "╮")
	empty_line := []byte("│ › " + strings.Repeat(" ", WIDTH-5) + "│")
	bot_border := []byte("╰" + strings.Repeat("─", WIDTH-2) + "╯")

	replacement := [][]byte{top_border, empty_line, bot_border}
	uis, err := p.UIs()
	if err != nil {
		log.Fatalf(err.Error())
		return err
	}

	opts := nvim.WindowConfig{
		Relative:  "editor",
		Anchor:    "NW",
		Width:     WIDTH,
		Height:    HEIGHT,
		Row:       (float64(uis[0].Height) / 2) - (float64(HEIGHT) / 2) - 2,
		Col:       (float64(uis[0].Width) / 2) - (float64(WIDTH) / 2),
		Style:     "minimal",
		ZIndex:    1,
		Focusable: false,
	}

	if err := p.SetBufferLines(buf, 0, -1, true, replacement); err != nil {
		log.Fatalf(err.Error())
		return err
	}

	if err := p.SetBufferOption(buf, "modifiable", false); err != nil {
		log.Fatalf(err.Error())
		return err
	}

	if err := p.SetBufferOption(buf, "bufhidden", "wipe"); err != nil {
		log.Fatalf(err.Error())
		return err
	}

	if err := p.SetBufferOption(buf, "buftype", "nofile"); err != nil {
		log.Fatalf(err.Error())
		return err
	}

	win, err := p.OpenWindow(buf, false, &opts)
	if err != nil {
		log.Fatalf(err.Error())
		return err
	}
	p.wins[&win] = true

	if err := p.SetWindowOption(win, "winhl", "Normal:TelescopeBorder"); err != nil {
		log.Fatalf(err.Error())
		return err
	}

	if err := p.SetWindowOption(win, "winblend", 0); err != nil {
		log.Fatalf(err.Error())
		return err
	}

	if err := p.SetWindowOption(win, "foldlevel", 100); err != nil {
		log.Fatalf(err.Error())
		return err
	}

	return nil
}

func (p *Command) createInput() {
	log.Printf("Creating Input")
	buf, err := p.CreateBuffer(false, true)
	if err != nil {
		log.Fatalf(err.Error())
	}
	p.Buffer = &buf

	uis, err := p.UIs()
	if err != nil {
		log.Fatalf(err.Error())
	}

	opts := nvim.WindowConfig{
		Relative:  "editor",
		Width:     WIDTH - 7,
		Height:    1,
		Row:       (float64(uis[0].Height) / 2) - (float64(HEIGHT) / 2) - 1,
		Col:       (float64(uis[0].Width) / 2) - (float64(WIDTH) / 2) + 4,
		Style:     "minimal",
		ZIndex:    3,
		Focusable: false,
	}

	if err := p.Command("startinsert!"); err != nil {
		log.Fatalf(err.Error())
	}

	if err := p.SetBufferOption(buf, "bufhidden", "wipe"); err != nil {
		log.Fatalf(err.Error())
	}

	if err := p.SetBufferOption(buf, "buftype", "nofile"); err != nil {
		log.Fatalf(err.Error())
	}

	win, err := p.OpenWindow(buf, true, &opts)
	if err != nil {
		log.Fatalf(err.Error())
	}
	p.wins[&win] = true

	if err := p.SetWindowOption(win, "winhl", "Normal:TelescopeNormal"); err != nil {
		log.Fatalf(err.Error())
	}

	if err := p.SetWindowOption(win, "winblend", 0); err != nil {
		log.Fatalf(err.Error())
	}

	if err := p.SetWindowOption(win, "foldlevel", 100); err != nil {
		log.Fatalf(err.Error())
	}

	p.Command("autocmd QuitPre <buffer> ++nested ++once :silent call SpotifyCloseWin()")
	p.Command("autocmd BufLeave <buffer> ++nested ++once :silent call SpotifyCloseWin()")
}

func (p *Command) getCurrentlyPlayingTrack() error {
	res := p.call("http://localhost:3000/currently-playing")
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf(err.Error())
	}

	if res.StatusCode != 200 {
		log.Fatalf(res.Status, string(body))
		return err
	}

	var currentlyPlaying spotify.CurrentlyPlaying
	if err = json.Unmarshal(body, &currentlyPlaying); err != nil {
		log.Fatalf(err.Error())
	}

	if currentlyPlaying.Playing {
		log.Printf("Creating CurrentlyPlaying")
		buf, err := p.CreateBuffer(false, true)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		var artists string
		for i, artist := range currentlyPlaying.Item.Artists {
			if i == 0 {
				artists += artist.Name
			} else if i == len(currentlyPlaying.Item.Artists)-1 {
				artists += " and " + artist.Name
			} else {
				artists += ", " + artist.Name
			}
		}
		playingName := currentlyPlaying.Item.Name + " by " + artists
		log.Println(playingName)

		top_border := []byte("╭" + strings.Repeat("─", WIDTH-2) + "╮")
		empty_line := []byte("│ 墳" + strings.Repeat(" ", (WIDTH-7-len(playingName))/2) + playingName + strings.Repeat(" ", (WIDTH-2-len(playingName))/2) + "│")
		bot_border := []byte("╰" + strings.Repeat("─", WIDTH-2) + "╯")

		replacement := [][]byte{top_border, empty_line, bot_border}
		uis, err := p.UIs()
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}

		opts := nvim.WindowConfig{
			Relative:  "editor",
			Anchor:    "NW",
			Width:     WIDTH,
			Height:    HEIGHT,
			Row:       (float64(uis[0].Height) / 2) - (float64(HEIGHT) / 2) + 1,
			Col:       (float64(uis[0].Width) / 2) - (float64(WIDTH) / 2),
			Style:     "minimal",
			ZIndex:    1,
			Focusable: false,
		}

		if err := p.SetBufferLines(buf, 0, -1, true, replacement); err != nil {
			log.Fatalf(err.Error())
			return err
		}

		if err := p.SetBufferOption(buf, "modifiable", false); err != nil {
			log.Fatalf(err.Error())
			return err
		}

		if err := p.SetBufferOption(buf, "bufhidden", "wipe"); err != nil {
			log.Fatalf(err.Error())
			return err
		}

		if err := p.SetBufferOption(buf, "buftype", "nofile"); err != nil {
			log.Fatalf(err.Error())
			return err
		}

		win, err := p.OpenWindow(buf, false, &opts)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		p.wins[&win] = true

		if err := p.SetWindowOption(win, "winhl", "Normal:TelescopeBorder"); err != nil {
			log.Fatalf(err.Error())
			// return err
		}

		if err := p.SetWindowOption(win, "winblend", 0); err != nil {
			log.Fatalf(err.Error())
			return err
		}

		if err := p.SetWindowOption(win, "foldlevel", 100); err != nil {
			log.Fatalf(err.Error())
			return err
		}

		return nil
	}

	return nil
}

func (p *Command) setKeyMaps() {
	log.Printf("Setting Keymaps")
	keys := [][3]string{
		{"n", "<Esc>", ":call SpotifyCloseWin()<CR>"},
		{"n", "q", ":call SpotifyCloseWin()<CR>"},
		{"i", "<CR>", "<esc>:call SpotifySearch('track')<CR>"},
		{"i", "<C-T>", "<esc>:call SpotifySearch('track')<CR>"},
		{"n", "<C-T>", "<esc>:call SpotifySearch('track')<CR>"},
		{"i", "<C-R>", "<esc>:call SpotifySearch('artist')<CR>"},
		{"n", "<C-R>", "<esc>:call SpotifySearch('artist')<CR>"},
		{"i", "<C-L>", "<esc>:call SpotifySearch('album')<CR>"},
		{"n", "<C-L>", "<esc>:call SpotifySearch('album')<CR>"},
		{"i", "<C-P>", "<esc>:call SpotifySearch('playlist')<CR>"},
		{"n", "<C-P>", "<esc>:call SpotifySearch('playlist')<CR>"},
		{"", "<C-P>", ":call SpotifyPlay()<CR>"},
	}

	opts := map[string]bool{"noremap": true, "silent": true, "nowait": true}
	for _, k := range keys {
		p.SetBufferKeyMap(*p.Buffer, k[0], k[1], k[2], opts)
	}
}

func (p *Command) configPlugin() {
	log.Printf("Configuring Plugin")
	p.createPlaceholder()
	p.getCurrentlyPlayingTrack()
	p.createInput()
	p.setKeyMaps()
}

func (p *Command) closeWins() error {
	if err := p.Command("stopinsert!"); err != nil {
		log.Fatalf(err.Error())
	}

	for win := range p.wins {
		p.closeOpenWin(win)
	}
	return nil
}

func (p *Command) search(args []string) {
	log.Printf("starting search...")
	searchType := args[0]
	b, err := p.CurrentLine()
	if err != nil {
		log.Fatalf("Input cannot be empty")
	}
	input := string(b)
	log.Printf("search input: %s", input)
	log.Printf("search type: %s", searchType)
	p.SetVar("spotify_type", searchType)
	p.SetVar("spotify_title", input)

	res := p.call(fmt.Sprintf("http://localhost:3000/search/%s/%s", searchType, input))
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf(err.Error())
	}

	if res.StatusCode != 200 {
		log.Fatalf(string(body))
		return
	}

	var searchResult spotify.SearchResult
	if err = json.Unmarshal(body, &searchResult); err != nil {
		log.Fatalf(err.Error())
	}
	p.SetVar("spotify_search", searchResult)
	p.Command("lua require'nvim-spotify'.init()")
}

func (p *Command) play(args []string) {
	p.call(fmt.Sprintf("http://localhost:3000/play/%s", args[0]))
}

func (p *Command) closeOpenWin(w *nvim.Window) {
	if p.wins[w] {
		p.CloseWindow(*w, true)
	}
}

func Register(p *plugin.Plugin) error {
	log.Printf("Registering Plugin")
	c := NewCommand(p.Nvim)

	p.HandleCommand(&plugin.CommandOptions{Name: "Spotify"}, c.configPlugin)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifyCloseWin"}, c.closeWins)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifySearch"}, c.search)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifyPlay"}, c.play)

	return nil
}

func main() {
	l, _ := os.Create("/tmp/nvim-spotify-plugin.log")
	log.SetOutput(l)
	defer l.Close()

	plugin.Main(Register)
}
