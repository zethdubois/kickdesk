package tmux

import (
	"strings"
	"testing"
)

func TestPaneTitle(t *testing.T) {
	if got := PaneTitle("kickagent", "up"); got != "kickdesk-kickagent" {
		t.Fatalf("up title = %q", got)
	}
	if got := PaneTitle("kickagent", ""); got != "kickdesk-kickagent" {
		t.Fatalf("empty title = %q", got)
	}
	if got := PaneTitle("kickagent", "wiki"); got != "kickdesk-kickagent-wiki" {
		t.Fatalf("wiki title = %q", got)
	}
}

func TestBuildChildMenuCmd(t *testing.T) {
	cmd := buildChildMenuCmd(2, "/usr/bin/kickdesk", "kam2", layoutResult{
		menuID: "%1",
		apps:   map[string]string{"kickstaff": "%0", "kickagent": "%2"},
	}, "/home/me/projects/kickdesk")
	if !strings.HasPrefix(cmd, "cd '/home/me/projects/kickdesk' && ") {
		t.Fatalf("expected cd prefix; got %q", cmd)
	}
	// Profile must be env prefix, not a positional argument after the binary.
	if strings.Contains(cmd, "'/usr/bin/kickdesk' KICKDESK_PROFILE=") {
		t.Fatal("KICKDESK_PROFILE must not appear after binary path")
	}
	if !strings.HasSuffix(cmd, "'/usr/bin/kickdesk' menu") {
		t.Fatalf("expected binary then menu; got %q", cmd)
	}
}
