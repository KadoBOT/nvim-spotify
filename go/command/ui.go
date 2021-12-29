package command

import (
	"log"
	"strings"

	"github.com/neovim/go-client/nvim"
)

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

func (p *Command) ShowDevices(devices []string) error {
	devices, ok := p.GetDevices()

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

func (p *Command) CloseWins() {
	p.DeleteBuffer(*p.Buffer, map[string]bool{"force": true})
	for win := range p.wins {
		if p.wins[win] {
			p.CloseWindow(*win, true)
		}
	}
}
