package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

const WIDTH = 70
const HEIGHT = 3

type Command struct {
	*nvim.Nvim
	*nvim.Buffer
	wins   map[*nvim.Window]bool
	input  string
	anchor *nvim.Window
	nsID   int
}

func safeString(str string) string {
	if len(str) > WIDTH-10 {
		return str[0:WIDTH-10] + "..."
	}
	return str
}

func execCommand(name string, args ...string) (string, bool) {
	cmd := exec.Command(name, args...)
	stoud, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSuffix(string(stoud), "\n"), true
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
	log.Println("getting devices")
	res, ok := execCommand("spt", "list", "-d")
	devices := strings.Split(res, "\n")

	devicesNames := [][]string{}
	if ok {
		for _, dev := range devices {
			cur := strings.SplitN(dev, " ", 2)
			devicesNames = append(devicesNames, []string{cur[1]})
		}
	}
	log.Println(devicesNames)
	p.SetVar("spotify_devices", devicesNames)

	p.Command("lua require'nvim-spotify'.devices()")

	return nil
}

func (p *Command) getCurrentlyPlayingTrack() error {
	log.Println("cur playing")
	curPlaying, ok := execCommand("spt", "playback", "-s", "-f", "%t by %a")

	if ok {
		log.Printf("Creating CurrentlyPlaying")
		buf, err := p.CreateBuffer(false, true)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		playingName := safeString(curPlaying)

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

func (p *Command) configPlugin() {
	log.Printf("Configuring Plugin")

	p.Command(`hi SpotifyBorder guifg=#1db954`)
	p.Command(`hi SpotifyText guifg=#1ed760`)
	p.Command(`hi SpotifySelection guifg=#191414 guibg=#1ed760`)

	p.createAnchor()
}

func (p *Command) start() {
	p.configPlugin()

	p.createPlaceholder()
	p.getCurrentlyPlayingTrack()
	p.createInput()

	keys := [][3]string{
		{"n", "<Esc>", ":call SpotifyCloseWin()<CR>"},
		{"n", "q", ":call SpotifyCloseWin()<CR>"},
		{"n", "<C-T>", ":call SpotifySearch('tracks')<CR>"},
		{"n", "<C-R>", ":call SpotifySearch('artists')<CR>"},
		{"n", "<C-L>", ":call SpotifySearch('albums')<CR>"},
		{"n", "<C-Y>", ":call SpotifySearch('playlists')<CR>"},
		{"i", "<CR>", "<C-O>:call SpotifySearch('tracks')<CR>"},
		{"i", "<C-T>", "<C-O>:call SpotifySearch('tracks')<CR>"},
		{"i", "<C-R>", "<C-O>:call SpotifySearch('artists')<CR>"},
		{"i", "<C-L>", "<C-O>:call SpotifySearch('albums')<CR>"},
		{"i", "<C-Y>", "<C-O>:call SpotifySearch('playlists')<CR>"},
	}

	p.setKeyMaps(keys)
}

func (p *Command) showDevices() error {
	return p.getDevices()
}

func (p *Command) closeWins() {
	p.DeleteBuffer(*p.Buffer, map[string]bool{"force": true})
	for win := range p.wins {
		p.closeOpenWin(win)
	}
}

func (p *Command) search(args []string) {
	log.Printf("starting search...")
	searchType := args[0]
	b, err := p.CurrentLine()
	if err != nil {
		log.Fatalf("Input cannot be empty")
	}
	input := string(b)

	formatList := func(list string) [][]string {
		var spotifySearch [][]string
		line := strings.Split(list, "\n")
		for _, l := range line {
			spotifySearch = append(spotifySearch, strings.Split(l, "||"))
		}
		return spotifySearch
	}
	var searchResult string

	switch searchType {
	case "tracks":
		searchResult, _ = execCommand("spt", "search", "--tracks", input, "--format", "%t||%a||%u", "--limit", "20")
	case "playlists":
		searchResult, _ = execCommand("spt", "search", "--playlists", input, "--format", "%p|| ||%u", "--limit", "20")
	case "artists":
		searchResult, _ = execCommand("spt", "search", "--artists", input, "--format", "%a|| ||%u", "--limit", "20")
	case "albums":
		searchResult, _ = execCommand("spt", "search", "--albums", input, "--format", "%b||%a||%u", "--limit", "20")
	case "shows":
		searchResult, _ = execCommand("spt", "search", "--shows", input, "--format", "%h||%a||%u", "--limit", "20")
	}
	spotifySearch := formatList(searchResult)

	p.SetVar("spotify_type", searchType)
	p.SetVar("spotify_title", input)

	p.SetVar("spotify_search", spotifySearch)
	p.Command("lua require'nvim-spotify'.init()")
}

func (p *Command) play(args []string) {
	var selected []string
	p.Var("spotify_device", &selected)
	if selected[0] != "" {
		execCommand("spt", "play", "-u", args[0], "-d", selected[0])
	} else {
		execCommand("spt", "play", "-u", args[0])
	}
}

func (p *Command) playback(args []string) {
	switch args[0] {
	case "next":
		execCommand("spt", "playback", "--next")
	case "pause":
		execCommand("spt", "playback", "--toggle")
	}
}

func (p *Command) save() {
	execCommand("spt", "playback", "--like")
}

func (p *Command) closeOpenWin(w *nvim.Window) {
	if p.wins[w] {
		p.CloseWindow(*w, true)
	}
}

func Register(p *plugin.Plugin) error {
	log.Printf("Registering Plugin")
	c := NewCommand(p.Nvim)

	p.HandleCommand(&plugin.CommandOptions{Name: "Spotify"}, c.start)
	p.HandleCommand(&plugin.CommandOptions{Name: "SpotifyDevices"}, c.showDevices)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifyCloseWin"}, c.closeWins)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifySearch"}, c.search)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifyPlay"}, c.play)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifyPlayback"}, c.playback)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifySave"}, c.save)

	return nil
}

func init() {
	l, _ := os.Create("/tmp/nvim-spotify-plugin.log")
	log.SetOutput(l)
	defer l.Close()
}

func main() {
	plugin.Main(Register)
}
