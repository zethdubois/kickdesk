package config

import (
	"strings"
	"testing"
)

func newRepublishApp(stop, republish []string, commands map[string]string) *Config {
	if commands == nil {
		commands = map[string]string{
			"up":          "true",
			"down":        "true",
			"build":       "true",
			"stop-server": "true",
			"republish":   "true",
		}
	}
	return &Config{
		Apps: map[string]App{
			"hot": {
				Path: ".",
				Ports: PortsList{
					{Role: "app", Port: 7099},
				},
				PrimaryPort: 7099,
				Workflows: Workflows{
					Start:     []string{"build", "up"},
					Stop:      stop,
					Republish: republish,
				},
				Commands: commands,
			},
		},
	}
}

func TestRepublishKeys_optional(t *testing.T) {
	cfg := newRepublishApp([]string{"stop-server"}, nil, nil)
	keys, err := cfg.RepublishKeys("hot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if keys != nil {
		t.Fatalf("expected nil republish keys, got %v", keys)
	}
}

func TestRepublishKeys_returnsCopy(t *testing.T) {
	cfg := newRepublishApp([]string{"stop-server"}, []string{"republish"}, nil)
	keys, err := cfg.RepublishKeys("hot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 1 || keys[0] != "republish" {
		t.Fatalf("unexpected keys: %v", keys)
	}
	keys[0] = "tampered"
	again, _ := cfg.RepublishKeys("hot")
	if again[0] != "republish" {
		t.Fatalf("RepublishKeys returned shared slice; got %v after mutation", again)
	}
}

func TestRepublishKeys_unknownApp(t *testing.T) {
	cfg := newRepublishApp([]string{"stop-server"}, []string{"republish"}, nil)
	if _, err := cfg.RepublishKeys("nope"); err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestValidate_republish_unknownCommand(t *testing.T) {
	cmds := map[string]string{
		"up":          "true",
		"down":        "true",
		"build":       "true",
		"stop-server": "true",
	}
	cfg := newRepublishApp([]string{"stop-server"}, []string{"republish"}, cmds)
	errs := cfg.Validate()
	found := false
	for _, e := range errs {
		if e.App == "hot" && strings.Contains(e.Message, `workflows.republish: unknown command "republish"`) {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected workflows.republish unknown command error, got %v", errs)
	}
}

func TestValidate_republish_emptyKey(t *testing.T) {
	cfg := newRepublishApp([]string{"stop-server"}, []string{""}, nil)
	errs := cfg.Validate()
	found := false
	for _, e := range errs {
		if e.App == "hot" && strings.Contains(e.Message, "workflows.republish: empty command key") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected empty key validation error, got %v", errs)
	}
}

func TestValidate_republish_omittedIsOK(t *testing.T) {
	cfg := newRepublishApp([]string{"stop-server"}, nil, nil)
	for _, e := range cfg.Validate() {
		if strings.Contains(e.Message, "workflows.republish") {
			t.Fatalf("did not expect republish validation error when omitted: %v", e)
		}
	}
}

func TestValidate_republish_allKeysKnownPasses(t *testing.T) {
	cfg := newRepublishApp([]string{"stop-server"}, []string{"republish"}, nil)
	for _, e := range cfg.Validate() {
		if strings.Contains(e.Message, "workflows.republish") {
			t.Fatalf("unexpected republish error: %v", e)
		}
	}
}
