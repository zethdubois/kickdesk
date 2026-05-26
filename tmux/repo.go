package tmux

import (
	"os"
	"path/filepath"
)

// RepoRoot returns the canonical kickdesk checkout directory.
// It always resolves to ~/projects/kickdesk.
func RepoRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "projects", "kickdesk"), nil
}
