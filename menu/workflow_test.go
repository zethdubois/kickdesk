package menu

import (
	"reflect"
	"testing"

	"github.com/Kick-Asset-Management/kickdesk/config"
)

func cfgWithApp(name string, wf config.Workflows, commands map[string]string) *config.Config {
	if commands == nil {
		commands = map[string]string{
			"up":          "true",
			"down":        "true",
			"build":       "true",
			"stop-server": "true",
			"republish":   "true",
		}
	}
	return &config.Config{
		Apps: map[string]config.App{
			name: {
				Path:        ".",
				Workflows:   wf,
				Commands:    commands,
				PrimaryPort: 7099,
				Ports:       config.PortsList{{Role: "app", Port: 7099}},
			},
		},
	}
}

func TestEffectiveProcedure_startupSkipsMigrate(t *testing.T) {
	cfg := cfgWithApp("web", config.Workflows{
		Start: []string{"db-up", "migrate", "up"},
		Stop:  []string{"stop-server"},
	}, map[string]string{
		"up":          "true",
		"down":        "true",
		"build":       "true",
		"stop-server": "true",
		"db-up":       "true",
		"migrate":     "true",
	})
	proc, err := EffectiveProcedure(cfg, "web", false, "ok")
	if err != nil {
		t.Fatal(err)
	}
	if proc.Title != "Startup" {
		t.Fatalf("title = %q, want Startup", proc.Title)
	}
	want := []string{"db-up", "up"}
	if !reflect.DeepEqual(proc.Keys, want) {
		t.Fatalf("keys = %v, want %v", proc.Keys, want)
	}
	if proc.HasRepublish() {
		t.Fatal("startup should never expose republish section")
	}
	if proc.RepublishFrom != len(proc.Keys) {
		t.Fatalf("RepublishFrom = %d, want %d", proc.RepublishFrom, len(proc.Keys))
	}
}

func TestEffectiveProcedure_shutdownOnlyWhenNoRepublish(t *testing.T) {
	cfg := cfgWithApp("agent", config.Workflows{
		Start: []string{"build", "up"},
		Stop:  []string{"stop-server"},
	}, nil)
	proc, err := EffectiveProcedure(cfg, "agent", true, "n/a")
	if err != nil {
		t.Fatal(err)
	}
	if proc.Title != "Shutdown" {
		t.Fatalf("title = %q, want Shutdown", proc.Title)
	}
	if proc.HasRepublish() {
		t.Fatal("did not expect republish section")
	}
	if !reflect.DeepEqual(proc.Keys, []string{"stop-server"}) {
		t.Fatalf("keys = %v", proc.Keys)
	}
}

func TestEffectiveProcedure_shutdownPlusRepublish(t *testing.T) {
	cfg := cfgWithApp("agent", config.Workflows{
		Start:     []string{"build", "publish-artifacts", "up"},
		Stop:      []string{"stop-server", "down"},
		Republish: []string{"republish"},
	}, map[string]string{
		"up":                "true",
		"down":              "true",
		"build":             "true",
		"stop-server":       "true",
		"publish-artifacts": "true",
		"republish":         "true",
	})
	proc, err := EffectiveProcedure(cfg, "agent", true, "n/a")
	if err != nil {
		t.Fatal(err)
	}
	if proc.Title != "Shutdown · Republish" {
		t.Fatalf("title = %q, want 'Shutdown · Republish'", proc.Title)
	}
	want := []string{"stop-server", "down", "republish"}
	if !reflect.DeepEqual(proc.Keys, want) {
		t.Fatalf("keys = %v, want %v", proc.Keys, want)
	}
	if !proc.HasRepublish() {
		t.Fatal("expected HasRepublish == true")
	}
	if proc.RepublishFrom != 2 {
		t.Fatalf("RepublishFrom = %d, want 2", proc.RepublishFrom)
	}
	if !proc.IsRepublishIndex(2) || proc.IsRepublishIndex(1) {
		t.Fatalf("IsRepublishIndex misclassified: %v", proc)
	}
	if got := proc.RepublishKeys(); !reflect.DeepEqual(got, []string{"republish"}) {
		t.Fatalf("RepublishKeys = %v", got)
	}
}

func TestEffectiveProcedure_republishIgnoredWhenStopped(t *testing.T) {
	cfg := cfgWithApp("agent", config.Workflows{
		Start:     []string{"build", "up"},
		Stop:      []string{"stop-server"},
		Republish: []string{"republish"},
	}, nil)
	proc, err := EffectiveProcedure(cfg, "agent", false, "n/a")
	if err != nil {
		t.Fatal(err)
	}
	if proc.HasRepublish() {
		t.Fatal("republish section should not appear while app is stopped")
	}
	if proc.Title != "Startup" {
		t.Fatalf("title = %q", proc.Title)
	}
}

func TestEffectiveProcedure_unknownApp(t *testing.T) {
	cfg := cfgWithApp("agent", config.Workflows{
		Start: []string{"up"},
		Stop:  []string{"stop-server"},
	}, nil)
	if _, err := EffectiveProcedure(cfg, "missing", true, ""); err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestEffectiveProcedure_emptyRepublishBehavesAsNoRepublish(t *testing.T) {
	cfg := cfgWithApp("agent", config.Workflows{
		Start:     []string{"build", "up"},
		Stop:      []string{"stop-server"},
		Republish: []string{},
	}, nil)
	proc, err := EffectiveProcedure(cfg, "agent", true, "n/a")
	if err != nil {
		t.Fatal(err)
	}
	if proc.HasRepublish() {
		t.Fatal("empty republish slice must be treated as not configured")
	}
	if proc.Title != "Shutdown" {
		t.Fatalf("title = %q", proc.Title)
	}
}

func TestEffectiveWorkflowKeys_matchesProcedure(t *testing.T) {
	cfg := cfgWithApp("agent", config.Workflows{
		Start:     []string{"build", "up"},
		Stop:      []string{"stop-server"},
		Republish: []string{"republish"},
	}, nil)
	keys, err := EffectiveWorkflowKeys(cfg, "agent", true, "n/a")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(keys, []string{"stop-server", "republish"}) {
		t.Fatalf("keys = %v", keys)
	}
}
