package command

import (
	"log"

	"github.com/neovim/go-client/nvim"
)

// Command stores nvim and buffer
type Command struct {
	*nvim.Nvim
	*nvim.Buffer
	wins   map[*nvim.Window]bool
	anchor *nvim.Window
}

// WIDTH of the buffers
const WIDTH = 70

// HEIGHT of the buffers
const HEIGHT = 3

// NewCommand creates the wrapper for handling the plugin
func NewCommand(v *nvim.Nvim) *Command {
	return &Command{Nvim: v, wins: make(map[*nvim.Window]bool)}
}

// Start calls all methods necessary to run the plugin
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

// ConfigPlugin configures the plugin
func (p *Command) ConfigPlugin() {
	log.Printf("Configuring Plugin")

	p.Command(`hi SpotifyBorder guifg=#1db954`)
	p.Command(`hi SpotifyText guifg=#1ed760`)
	p.Command(`hi SpotifySelection guifg=#191414 guibg=#1ed760`)

	p.createAnchor()
}
