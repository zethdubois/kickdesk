package run

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LaunchInTerminal starts shell in a new terminal window (fire-and-forget).
// Returns true if a terminal was launched, false if none found (caller should fall back).
func LaunchInTerminal(dir, shell string) (bool, error) {
	emulator, args := terminalCommand(dir, shell)
	if emulator == "" {
		return false, nil
	}
	cmd := exec.Command(emulator, args...)
	if err := cmd.Start(); err != nil {
		return false, err
	}
	return true, nil
}

func terminalCommand(dir, shell string) (string, []string) {
	script := fmt.Sprintf("cd %q && %s; exec bash", dir, shell)

	if custom := strings.TrimSpace(os.Getenv("KICKDESK_TERMINAL")); custom != "" {
		parts := strings.Fields(custom)
		if len(parts) == 0 {
			return "", nil
		}
		return parts[0], append(parts[1:], "bash", "-lc", script)
	}

	candidates := []string{"gnome-terminal", "kgx", "kitty", "alacritty", "konsole", "xterm"}
	for _, name := range candidates {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		return path, terminalArgs(path, script)
	}
	return "", nil
}

func terminalArgs(emulator, script string) []string {
	base := filepath.Base(emulator)
	switch base {
	case "gnome-terminal", "kgx":
		return []string{"--", "bash", "-lc", script}
	case "kitty":
		return []string{"bash", "-lc", script}
	default:
		return []string{"-e", "bash", "-lc", script}
	}
}
