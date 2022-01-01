package main

import (
	"log"
	"os"

	"github.com/kadobot/nvim-spotify/command"
	"github.com/neovim/go-client/nvim/plugin"
)

// Register the nvim handlers
func Register(p *plugin.Plugin) error {
	log.Printf("Registering Plugin")
	c := command.NewCommand(p.Nvim)

	p.HandleCommand(&plugin.CommandOptions{Name: "Spotify"}, c.Start)
	p.HandleCommand(&plugin.CommandOptions{Name: "SpotifyDevices"}, c.ShowDevices)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifyCloseWin"}, c.CloseWins)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifySearch"}, c.Search)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifyPlay"}, c.Play)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifyPlayback"}, c.Playback)
	p.HandleFunction(&plugin.FunctionOptions{Name: "SpotifySave"}, c.Save)

	return nil
}

func main() {
	file, _ := ioutil.TempFile("", "nvim-spotify-plugin.*.log")
	log.SetOutput(file)
	defer l.Close()
	plugin.Main(Register)
}
