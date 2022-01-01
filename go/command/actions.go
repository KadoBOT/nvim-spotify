package command

import (
	"log"
	"strings"

	"github.com/kadobot/nvim-spotify/utils"
)

// GetCurrentlyPlayingTrack shows the currently playing track when the plugin is open
func (p *Command) GetCurrentlyPlayingTrack() {
	log.Println("cur playing")
	curPlaying, ok := utils.ExecCommand("spt", "playback", "-s", "-f", "%t by %a")

	if ok {
		p.showCurrentlyPlaying(curPlaying)
	}
}

func (p *Command) setKeyMaps(keys [][3]string) {
	log.Printf("Setting Keymaps")

	opts := map[string]bool{"noremap": true, "silent": true, "nowait": true}
	for _, k := range keys {
		p.SetBufferKeyMap(*p.Buffer, k[0], k[1], k[2], opts)
	}
}

// Search calls the Spotify API responsible for searching
func (p *Command) Search(args []string) {
	log.Printf("starting search...")
	searchType := args[0]
	b, err := p.CurrentLine()
	if err != nil {
		log.Fatalf("Input cannot be empty")
	}
	input := string(b)

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

	var spotifySearch [][]string
	line := strings.Split(searchResult, "\n")
	for _, l := range line {
		spotifySearch = append(spotifySearch, strings.Split(l, "||"))
	}

	p.SetVar("spotify_type", searchType)
	p.SetVar("spotify_title", input)

	p.SetVar("spotify_search", spotifySearch)
	p.Command("lua require'nvim-spotify'.init()")
}

// Play plays the given URI
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

// Playback skips the track or pause/resume a track
func (p *Command) Playback(args []string) {
	switch args[0] {
	case "next":
		utils.ExecCommand("spt", "playback", "--next")
	case "pause":
		utils.ExecCommand("spt", "playback", "--toggle")
	case "prev":
		utils.ExecCommand("spt", "playback", "--previous")
	}
}

// Save adds the currently playing track to "My Library"
func (p *Command) Save() {
	utils.ExecCommand("spt", "playback", "--like")
}

// ShowDevices displays the list of devices
func (p *Command) ShowDevices(_ []string) {
	log.Println("getting devices")
	res, ok := utils.ExecCommand("spt", "list", "-d")

	if ok {
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
	}
}

// CloseWins closes all open windows
func (p *Command) CloseWins() {
	p.DeleteBuffer(*p.Buffer, map[string]bool{"force": true})
	for win := range p.wins {
		if p.wins[win] {
			p.CloseWindow(*win, true)
		}
	}
}
