package command

import (
	"log"

	"github.com/kadobot/nvim-spotify/client"
	"github.com/neovim/go-client/nvim"
	"github.com/zmb3/spotify/v2"
)

type Command struct {
	*nvim.Nvim
	*nvim.Buffer
	wins   map[*nvim.Window]bool
	input  string
	anchor *nvim.Window
	nsID   int
	client *spotify.Client
}

const WIDTH = 70
const HEIGHT = 3

func NewCommand(v *nvim.Nvim) *Command {
	client := client.NewClient()
	return &Command{Nvim: v, wins: make(map[*nvim.Window]bool), client: client}
}

func (p *Command) Start() {
	p.ConfigPlugin()

	p.createPlaceholder()
	p.GetCurrentlyPlayingTrack()
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

func (p *Command) ConfigPlugin() {
	log.Printf("Configuring Plugin")

	p.Command(`hi SpotifyBorder guifg=#1db954`)
	p.Command(`hi SpotifyText guifg=#1ed760`)
	p.Command(`hi SpotifySelection guifg=#191414 guibg=#1ed760`)

	p.createAnchor()
}
