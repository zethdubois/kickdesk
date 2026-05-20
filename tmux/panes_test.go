package tmux

import (
	"strings"
	"testing"
)

func TestBuildChildMenuCmd(t *testing.T) {
	cmd := buildChildMenuCmd(2, "/usr/bin/kickdesk", "kam2", layoutResult{
		menuID: "%1",
		apps:   map[string]string{"publicweb": "%0", "kickagent": "%2"},
	})
	// Profile must be env prefix, not a positional argument after the binary.
	if strings.Contains(cmd, "'/usr/bin/kickdesk' KICKDESK_PROFILE=") {
		t.Fatal("KICKDESK_PROFILE must not appear after binary path")
	}
	if !strings.HasSuffix(cmd, "'/usr/bin/kickdesk' menu") {
		t.Fatalf("expected binary then menu; got %q", cmd)
	}
}
