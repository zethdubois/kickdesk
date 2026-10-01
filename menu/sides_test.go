package menu

import (
	"testing"

	"github.com/Kick-Asset-Management/kickdesk/config"
	"github.com/Kick-Asset-Management/kickdesk/status"
)

func agentApp(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		Apps: map[string]config.App{
			"kickagent": {
				PrimaryPort: 7099,
				Ports: config.PortsList{
					{Role: "publish", Port: 7099},
					{Role: "standalone", Port: 7098},
					{Role: "wiki", Port: 7097},
					{Role: "discovery", Port: 7096},
					{Role: "db", Port: 7043},
				},
				Workflows: config.Workflows{
					Start:     []string{"install", "build", "up"},
					Stop:      []string{"stop-server", "stop-wiki", "stop-discovery", "down"},
					Republish: []string{"republish"},
				},
				Commands: map[string]string{
					"up":             "serve",
					"down":           "true",
					"build":          "true",
					"install":        "true",
					"stop-server":    "true",
					"wiki":           "pnpm kam:wiki",
					"stop-wiki":      "true",
					"discovery":      "pnpm kam:discovery",
					"stop-discovery": "true",
					"republish":      "true",
				},
			},
		},
	}
}

func TestListSideServices_wikiStartWhilePublishUp(t *testing.T) {
	open := map[int]bool{7099: true}
	isPortOpen = func(port int) bool { return open[port] }
	t.Cleanup(func() { isPortOpen = status.IsPortOpen })

	sides := ListSideServices(agentApp(t), "kickagent", true)
	got := map[string]SideService{}
	for _, s := range sides {
		got[s.Key] = s
	}
	if _, ok := got["wiki"]; !ok {
		t.Fatalf("expected wiki start, got %+v", sides)
	}
	if got["wiki"].Hotkey != 'w' {
		t.Fatalf("wiki hotkey = %q, want w", got["wiki"].Hotkey)
	}
	if _, ok := got["discovery"]; !ok {
		t.Fatalf("expected discovery start, got %+v", sides)
	}
	if got["discovery"].Hotkey != 'd' {
		t.Fatalf("discovery hotkey = %q", got["discovery"].Hotkey)
	}
	if _, ok := got["stop-wiki"]; ok {
		t.Fatal("stop-wiki is already a shutdown step; should not duplicate in Services")
	}
	for _, s := range sides {
		if s.Role == "db" || s.Port == 7043 {
			t.Fatal("db ports must not appear as side services")
		}
		if s.Port == 7099 {
			t.Fatal("primary port must not appear as a side service")
		}
	}
}

func TestListSideServices_wikiStopWhenPrimaryDown(t *testing.T) {
	open := map[int]bool{7097: true}
	isPortOpen = func(port int) bool { return open[port] }
	t.Cleanup(func() { isPortOpen = status.IsPortOpen })

	sides := ListSideServices(agentApp(t), "kickagent", false)
	var wiki *SideService
	for i := range sides {
		if sides[i].Role == "wiki" {
			wiki = &sides[i]
			break
		}
	}
	if wiki == nil {
		t.Fatalf("expected wiki row, got %+v", sides)
	}
	if wiki.Key != "stop-wiki" {
		t.Fatalf("key = %q, want stop-wiki (primary is down so shutdown list is hidden)", wiki.Key)
	}
}

func TestListSideServices_skipsUnknownApp(t *testing.T) {
	if got := ListSideServices(agentApp(t), "missing", true); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}
