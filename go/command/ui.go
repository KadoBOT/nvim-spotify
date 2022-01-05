package command

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/kadobot/nvim-spotify/utils"
	"github.com/neovim/go-client/nvim"
)

func (p *Command) createPlaceholder() {
	log.Printf("Creating Placeholder")
	buf, err := p.CreateBuffer(false, true)
	if err != nil {
		log.Println(err.Error())
		return
	}
	p.placeholder = &buf

	text := " Spotify Search: Tracks "
	repeatW := (WIDTH - len(text) - 1) / 2
	border := strings.Repeat("─", repeatW)

	topBorder := []byte("╭" + border + text + border + "╮")
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
		ZIndex:    50,
		Focusable: false,
	}

	p.SetBufferLines(buf, 0, -1, true, replacement)
	p.SetBufferOption(buf, "modifiable", false)
	p.SetBufferOption(buf, "bufhidden", "wipe")
	p.SetBufferOption(buf, "buftype", "nofile")

	win, _ := p.OpenWindow(buf, false, &opts)
	p.wins[&win] = true

	p.SetWindowOption(win, "winhl", "Normal:SpotifyBorder")
	p.SetWindowOption(win, "winblend", 0)
	p.SetWindowOption(win, "foldlevel", 100)
}

func (p *Command) createAnchor() {
	log.Printf("Creating Anchor")
	buf, _ := p.CreateBuffer(false, true)
	uis, _ := p.UIs()

	opts := nvim.WindowConfig{
		Relative:  "editor",
		Anchor:    "NW",
		Width:     1,
		Height:    1,
		Row:       (float64(uis[0].Height) / 2) - (float64(HEIGHT) / 2),
		Col:       (float64(uis[0].Width) / 2) - (float64(WIDTH) / 2) + 1.5,
		Style:     "minimal",
		ZIndex:    50,
		Focusable: false,
	}

	p.SetBufferOption(buf, "bufhidden", "wipe")
	p.SetBufferOption(buf, "buftype", "nofile")

	win, _ := p.OpenWindow(buf, false, &opts)

	p.anchor = &win
	p.wins[&win] = true
}

func (p *Command) changeInputTitle(searchType string) {
	p.SetBufferOption(*p.placeholder, "modifiable", true)
	t := strings.ToUpper(searchType[:1]) + searchType[1:]
	p.SetVar("spotify_type", t)

	text := fmt.Sprintf(" Spotify Search: %s ", t)
	repeatW := (WIDTH - len(text) - 1) / 2
	border := strings.Repeat("─", repeatW)
	border2 := strings.Repeat("─", repeatW-len(text)%2)

	topBorder := []byte("╭" + border + text + border2 + "╮")

	p.SetBufferLines(*p.placeholder, 0, 1, true, [][]byte{topBorder})
	p.SetBufferOption(*p.placeholder, "modifiable", false)
}

func (p *Command) createInput() {
	log.Printf("Creating Input")
	buf, _ := p.CreateBuffer(false, true)
	p.Buffer = &buf

	opts := nvim.WindowConfig{
		Relative:  "win",
		Win:       *p.anchor,
		Width:     WIDTH - 7,
		Height:    1,
		BufPos:    [2]int{0, 0},
		Row:       1,
		Col:       3,
		Style:     "minimal",
		ZIndex:    51,
		Focusable: true,
	}

	p.Command("startinsert!")
	p.SetBufferOption(buf, "bufhidden", "wipe")
	p.SetBufferOption(buf, "buftype", "nofile")

	win, _ := p.OpenWindow(buf, true, &opts)

	p.wins[&win] = true

	p.SetWindowOption(win, "winhl", "Normal:SpotifyText")
	p.SetWindowOption(win, "winblend", 0)
	p.SetWindowOption(win, "foldlevel", 100)

	p.Command("autocmd QuitPre <buffer> ++nested ++once :silent call SpotifyCloseWin()")
	p.Command("autocmd BufLeave <buffer> ++nested ++once :silent call SpotifyCloseWin()")
}

func (p *Command) showCurrentlyPlaying(curPlaying string) {
	log.Printf("Creating CurrentlyPlaying")
	buf, _ := p.CreateBuffer(false, true)
	playingName := utils.SafeString(curPlaying, WIDTH-10)

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
		ZIndex:    50,
		Focusable: false,
	}

	p.SetBufferLines(buf, 0, -1, true, replacement)

	p.SetBufferText(buf, 1, 7, 1, utf8.RuneCountInString(playingName)+7, [][]byte{[]byte(playingName)})
	p.SetBufferOption(buf, "modifiable", false)
	p.SetBufferOption(buf, "bufhidden", "wipe")
	p.SetBufferOption(buf, "buftype", "nofile")

	win, _ := p.OpenWindow(buf, false, &opts)
	p.wins[&win] = true

	p.SetWindowOption(win, "winhl", "Normal:SpotifyBorder")
	p.SetWindowOption(win, "winblend", 0)
	p.SetWindowOption(win, "foldlevel", 100)
}
