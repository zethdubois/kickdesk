package tmux

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepoRootFixedPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	root, err := RepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "projects", "kickdesk")
	if root != want {
		t.Fatalf("got %q", root)
	}
}
