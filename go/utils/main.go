package utils

import (
	"os/exec"
	"strings"
)

// SafeString sanitizes the given string so it doesn't break the window
func SafeString(str string, size int) string {
	if len(str) > size {
		return str[0:size] + "..."
	}
	return str
}

// ExecCommand executes the shell command
func ExecCommand(name string, args ...string) (string, bool) {
	cmd := exec.Command(name, args...)
	stoud, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSuffix(string(stoud), "\n"), true
}
