package command

import (
	"log"
	"strings"
	"unicode/utf8"

	"github.com/kadobot/nvim-spotify/utils"
	"github.com/neovim/go-client/nvim"
)

func (p *Command) createPlaceholder() error {
	log.Printf("Creating Placeholder")
	buf, err := p.CreateBuffer(false, true)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	topBorder := []byte("╭" + strings.Repeat("─", (WIDTH-17)/2) + " Spotify Search " + strings.Repeat("─", (WIDTH-18)/2) + "╮")
	emptyLine := []byte("│ › " + strings.Repeat(" ", WIDTH-5) + "│")
	botBorder := []byte("╰" + strings.Repeat("─", WIDTH-2) + "╯")

	replacement := [][]byte{topBorder, emptyLine, botBorder}

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
		log.Println(err.Error())
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
		log.Println(err.Error())
	}

	uis, err := p.UIs()
	if err != nil {
		log.Println(err.Error())
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
		log.Println(err.Error())
	}

	p.anchor = &win
	p.wins[&win] = true
}

func (p *Command) createInput() {
	log.Printf("Creating Input")
	buf, err := p.CreateBuffer(false, true)
	if err != nil {
		log.Println(err.Error())
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
		log.Println(err.Error())
	}
	p.wins[&win] = true

	p.SetWindowOption(win, "winhl", "Normal:SpotifyText")
	p.SetWindowOption(win, "winblend", 0)
	p.SetWindowOption(win, "foldlevel", 100)

	p.Command("autocmd QuitPre <buffer> ++nested ++once :silent call SpotifyCloseWin()")
	p.Command("autocmd BufLeave <buffer> ++nested ++once :silent call SpotifyCloseWin()")
}

func (p *Command) showCurrentlyPlaying(curPlaying string) {
	log.Printf("Creating CurrentlyPlaying")
	buf, err := p.CreateBuffer(false, true)
	if err != nil {
		log.Println(err.Error())
	}
	playingName := utils.SafeString(curPlaying, WIDTH-10)

	log.Println(playingName)

	topBorder := []byte("╭" + strings.Repeat("─", (WIDTH-19)/2) + " Currently Playing " + strings.Repeat("─", (WIDTH-21)/2) + "╮")
	emptyLine := []byte("│ 墳" + strings.Repeat(" ", WIDTH-5) + "│")
	botBorder := []byte("╰" + strings.Repeat("─", WIDTH-2) + "╯")

	replacement := [][]byte{topBorder, emptyLine, botBorder}

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
		log.Println(err.Error())
	}

	p.SetBufferText(buf, 1, 7, 1, utf8.RuneCountInString(playingName)+7, [][]byte{[]byte(playingName)})
	p.SetBufferOption(buf, "modifiable", false)
	p.SetBufferOption(buf, "bufhidden", "wipe")
	p.SetBufferOption(buf, "buftype", "nofile")

	win, err := p.OpenWindow(buf, false, &opts)
	if err != nil {
		log.Println(err.Error())
	}
	p.wins[&win] = true

	p.SetWindowOption(win, "winhl", "Normal:SpotifyBorder")
	p.SetWindowOption(win, "winblend", 0)
	p.SetWindowOption(win, "foldlevel", 100)
}
