package utils

import (
	"os/exec"
	"strings"

	"github.com/zmb3/spotify/v2"
)

func SafeString(str string, size int) string {
	if len(str) > size {
		return str[0:size] + "..."
	}
	return str
}

func ExecCommand(name string, args ...string) (string, bool) {
	cmd := exec.Command(name, args...)
	stoud, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSuffix(string(stoud), "\n"), true
}

func FormatArtistsName(artists []spotify.SimpleArtist) string {
	var artistsName string
	for i, artist := range artists {
		switch i {
		case 0:
			artistsName = artist.Name
		case len(artists) - 1:
			artistsName += ", " + artist.Name
		default:
			artistsName += "and " + artist.Name
		}
	}
	return artistsName
}
