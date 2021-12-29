package command

import (
	"log"
	"strings"
	"unicode/utf8"

	"github.com/kadobot/nvim-spotify/utils"
	"github.com/neovim/go-client/nvim"
)

func (p *Command) GetCurrentlyPlayingTrack() error {
	log.Println("cur playing")
	curPlaying, ok := utils.ExecCommand("spt", "playback", "-s", "-f", "%t by %a")

	if ok {
		log.Printf("Creating CurrentlyPlaying")
		buf, err := p.CreateBuffer(false, true)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		playingName := utils.SafeString(curPlaying, WIDTH-10)

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

func (p *Command) Search(args []string) {
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
		searchResult, _ = utils.ExecCommand("spt", "search", "--tracks", input, "--format", "%t||%a||%u", "--limit", "20")
	case "playlists":
		searchResult, _ = utils.ExecCommand("spt", "search", "--playlists", input, "--format", "%p|| ||%u", "--limit", "20")
	case "artists":
		searchResult, _ = utils.ExecCommand("spt", "search", "--artists", input, "--format", "%a|| ||%u", "--limit", "20")
	case "albums":
		searchResult, _ = utils.ExecCommand("spt", "search", "--albums", input, "--format", "%b||%a||%u", "--limit", "20")
	case "shows":
		searchResult, _ = utils.ExecCommand("spt", "search", "--shows", input, "--format", "%h||%a||%u", "--limit", "20")
	}
	spotifySearch := formatList(searchResult)

	p.SetVar("spotify_type", searchType)
	p.SetVar("spotify_title", input)

	p.SetVar("spotify_search", spotifySearch)
	p.Command("lua require'nvim-spotify'.init()")
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
