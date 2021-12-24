package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/KadoBOT/spotify/v2"
	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

const WIDTH = 70
const HEIGHT = 3

type Command struct {
	*nvim.Nvim
	*nvim.Buffer
	wins       map[*nvim.Window]bool
	input      string
	anchor     *nvim.Window
	devices    []spotify.PlayerDevice
	devicesBuf *nvim.Buffer
	selected   int
	nsID       int
}

func safeString(str string) string {
	if len(str) > WIDTH-10 {
		return str[0:WIDTH-10] + "..."
	}
	return str
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

	top_border := []byte("╭" + strings.Repeat("─", (WIDTH-17)/2) + " Spotify Search " + strings.Repeat("─", (WIDTH-18)/2) + "╮")
	empty_line := []byte("│ › " + strings.Repeat(" ", WIDTH-5) + "│")
	bot_border := []byte("╰" + strings.Repeat("─", WIDTH-2) + "╯")

	replacement := [][]byte{top_border, empty_line, bot_border}

	opts := nvim.WindowConfig{
		Relative:  "win",
		Win:       *p.anchor,
		Width:     WIDTH,
		Height:    HEIGHT,
		BufPos:    [2]int{0, 0},
		Row:       0.5,
		Col:       -2,
		Style:     "minimal",
		ZIndex:    1,
		Focusable: false,
	}

	p.SetBufferLines(buf, 0, -1, true, replacement)
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

	return nil
}

func (p *Command) createAnchor() {
	log.Printf("Creating Anchor")
	buf, err := p.CreateBuffer(false, true)
	if err != nil {
		log.Fatalf(err.Error())
	}

	uis, err := p.UIs()
	if err != nil {
		log.Fatalf(err.Error())
	}

	opts := nvim.WindowConfig{
		Relative:  "editor",
		Anchor:    "NW",
		Width:     1,
		Height:    1,
		Row:       (float64(uis[0].Height) / 2) - (float64(HEIGHT) / 2),
		Col:       (float64(uis[0].Width) / 2) - (float64(WIDTH) / 2) + 1.5,
		Style:     "minimal",
		ZIndex:    1,
		Focusable: false,
	}

	p.SetBufferOption(buf, "bufhidden", "wipe")
	p.SetBufferOption(buf, "buftype", "nofile")

	win, err := p.OpenWindow(buf, false, &opts)
	if err != nil {
		log.Fatalf(err.Error())
	}

	p.anchor = &win
	p.wins[&win] = true
}

func (p *Command) createInput() {
	log.Printf("Creating Input")
	buf, err := p.CreateBuffer(false, true)
	if err != nil {
		log.Fatalf(err.Error())
	}
	p.Buffer = &buf

	opts := nvim.WindowConfig{
		Relative:  "win",
		Win:       *p.anchor,
		Width:     WIDTH - 7,
		Height:    1,
		BufPos:    [2]int{0, 0},
		Row:       1,
		Col:       2,
		Style:     "minimal",
		ZIndex:    999,
		Focusable: true,
	}

	p.Command("startinsert!")
	p.SetBufferOption(buf, "bufhidden", "wipe")
	p.SetBufferOption(buf, "buftype", "nofile")

	win, err := p.OpenWindow(buf, true, &opts)
	if err != nil {
		log.Fatalf(err.Error())
	}
	p.wins[&win] = true

	p.SetWindowOption(win, "winhl", "Normal:SpotifyText")
	p.SetWindowOption(win, "winblend", 0)
	p.SetWindowOption(win, "foldlevel", 100)

	p.Command("autocmd QuitPre <buffer> ++nested ++once :silent call SpotifyCloseWin()")
	p.Command("autocmd BufLeave <buffer> ++nested ++once :silent call SpotifyCloseWin()")
}

func (p *Command) getDevices() error {
	res := p.call("https://europe-west3-nvim-spotify.cloudfunctions.net/devices-d76dbb1")
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf(err.Error())
	}

	if res.StatusCode != 200 {
		log.Fatalf(res.Status, string(body))
		return err
	}

	if err = json.Unmarshal(body, &p.devices); err != nil {
		log.Fatalf(err.Error())
	}

	if len(p.devices) != 0 {
		log.Printf("Listing Devices")
		buf, err := p.CreateBuffer(false, true)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		p.devicesBuf = &buf

		top_border := []byte("╭" + strings.Repeat("─", (WIDTH-22)/2) + " Connect to a Device " + strings.Repeat("─", (WIDTH-23)/2) + "╮")
		replacement := [][]byte{top_border}
		for i, device := range p.devices {
			emptyAmount := (WIDTH - 5 - len(device.Name))

			if i == p.selected {
				line := []byte("│ ▶ " + safeString(device.Name) + strings.Repeat(" ", emptyAmount) + "│")
				replacement = append(replacement, line)
			} else {
				line := []byte("│   " + safeString(device.Name) + strings.Repeat(" ", emptyAmount) + "│")
				replacement = append(replacement, line)
			}
		}

		bot_border := []byte("╰" + strings.Repeat("─", WIDTH-2) + "╯")
		replacement = append(replacement, bot_border)

		opts := nvim.WindowConfig{
			Relative:  "win",
			Win:       *p.anchor,
			Width:     WIDTH,
			Height:    len(replacement),
			BufPos:    [2]int{0, 0},
			Row:       3,
			Col:       -2,
			Style:     "minimal",
			ZIndex:    50,
			Focusable: true,
		}

		p.SetBufferLines(buf, 0, -1, false, replacement)
		p.SetBufferOption(buf, "bufhidden", "wipe")
		p.SetBufferOption(buf, "buftype", "nofile")

		win, err := p.OpenWindow(buf, false, &opts)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		p.wins[&win] = true

		p.SetWindowOption(win, "winhl", "Normal:SpotifyBorder")
		p.setDevicesHighlight(0)
		p.SetBufferOption(buf, "modifiable", false)

		return nil
	}

	return nil
}

func (p *Command) setDevicesHighlight(selected int) {
	log.Println("Setting devices highlight")
	p.SetBufferOption(*p.devicesBuf, "modifiable", true)
	p.Nvim.ClearBufferNamespace(*p.devicesBuf, p.nsID, 0, -1)
	p.Nvim.SetBufferText(*p.devicesBuf, p.selected+1, 4, p.selected+1, 7, [][]byte{[]byte(" ")})
	p.selected = selected
	p.Nvim.AddBufferHighlight(*p.devicesBuf, p.nsID, "SpotifySelection", p.selected+1, 3, WIDTH+3)
	p.Nvim.SetBufferText(*p.devicesBuf, p.selected+1, 4, p.selected+1, 5, [][]byte{[]byte("▶")})
	p.SetBufferOption(*p.devicesBuf, "modifiable", false)
}

func (p *Command) getCurrentlyPlayingTrack() error {
	res := p.call("https://europe-west3-nvim-spotify.cloudfunctions.net/cur_play-9dc7855")
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
		playingName := safeString(currentlyPlaying.Item.Name + " by " + artists)

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

		return nil
	}

	return nil
}

func (p *Command) setKeyMaps() {
	log.Printf("Setting Keymaps")

	keys := [][3]string{
		{"n", "<Esc>", ":call SpotifyCloseWin()<CR>"},
		{"n", "q", ":call SpotifyCloseWin()<CR>"},
		{"n", "<C-N>", ":call SpotifyDevices('next')<CR>"},
		{"n", "<Tab>", ":call SpotifyDevices('next')<CR>"},
		{"n", "<S-Tab>", ":call SpotifyDevices('prev')<CR>"},
		{"n", "<C-P>", ":call SpotifyDevices('prev')<CR>"},
		{"n", "<C-T>", ":call SpotifySearch('track')<CR>"},
		{"n", "<C-R>", ":call SpotifySearch('artist')<CR>"},
		{"n", "<C-L>", ":call SpotifySearch('album')<CR>"},
		{"n", "<C-Y>", ":call SpotifySearch('playlist')<CR>"},
		{"i", "<CR>", "<C-O>:call SpotifySearch('track')<CR>"},
		{"i", "<C-N>", "<C-O>:call SpotifyDevices('next')<CR>"},
		{"i", "<Tab>", "<C-O>:call SpotifyDevices('next')<CR>"},
		{"i", "<C-P>", "<C-O>:call SpotifyDevices('prev')<CR>"},
		{"i", "<S-Tab>", "<C-O>:call SpotifyDevices('prev')<CR>"},
		{"i", "<C-T>", "<C-O>:call SpotifySearch('track')<CR>"},
		{"i", "<C-R>", "<C-O>:call SpotifySearch('artist')<CR>"},
		{"i", "<C-L>", "<C-O>:call SpotifySearch('album')<CR>"},
		{"i", "<C-Y>", "<C-O>:call SpotifySearch('playlist')<CR>"},
	}

	opts := map[string]bool{"noremap": true, "silent": true, "nowait": true}
	for _, k := range keys {
		p.SetBufferKeyMap(*p.Buffer, k[0], k[1], k[2], opts)
	}
}

func (p *Command) configPlugin() {
	log.Printf("Configuring Plugin")

	p.Command(`hi SpotifyBorder guifg=#1db954`)
	p.Command(`hi SpotifyText guifg=#1ed760`)
	p.Command(`hi SpotifySelection guifg=#191414 guibg=#1ed760`)

	p.createAnchor()
	p.createPlaceholder()
	p.getCurrentlyPlayingTrack()
	p.getDevices()
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

	p.SetVar("spotify_type", searchType)
	p.SetVar("spotify_title", input)

	res := p.call(fmt.Sprintf("https://europe-west3-nvim-spotify.cloudfunctions.net/search-8821b6b?type=%s&query=%s", url.QueryEscape(searchType), url.QueryEscape(input)))
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
	p.call(fmt.Sprintf("https://europe-west3-nvim-spotify.cloudfunctions.net/play-7613342?uri=%s&id=%s", url.QueryEscape(args[0]), url.QueryEscape(p.devices[p.selected].ID.String())))
}

func (p *Command) playback(args []string) {
	switch args[0] {
	case "next":
		p.call("https://europe-west3-nvim-spotify.cloudfunctions.net/skip-09d0606")
	case "pause":
		p.call("https://europe-west3-nvim-spotify.cloudfunctions.net/pause-98016ef")
	}
}

func (p *Command) save() {
	p.call("https://europe-west3-nvim-spotify.cloudfunctions.net/save-9067f22")
}

func (p *Command) deviceSwitch(args []string) {
	selected := p.selected
	if args[0] == "next" {
		selected = (p.selected + 1) % len(p.devices)
	}

	if args[0] == "prev" {
		selected = (p.selected - 1) % len(p.devices)
	}

	p.setDevicesHighlight(int(math.Abs(float64(selected))))
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
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifyDevices"}, c.deviceSwitch)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifyPlayback"}, c.playback)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifySave"}, c.save)

	return nil
}

func main() {
	l, _ := os.Create("/tmp/nvim-spotify-plugin.log")
	log.SetOutput(l)
	defer l.Close()

	plugin.Main(Register)
}
