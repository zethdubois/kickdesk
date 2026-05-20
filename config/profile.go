package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Profile selects which apps appear in the menu and optional tmux layout.
type Profile struct {
	AppOrder []string          `json:"app_order"`
	Labels   map[string]string `json:"labels"`
	Tmux     TmuxSettings      `json:"tmux"`
}

// TmuxSettings configures the dashboard layout (-c profile with tmux.enabled).
type TmuxSettings struct {
	Enabled bool   `json:"enabled"`
	Top     int    `json:"top"`
	Session string `json:"session"`
}

// ConfigDir returns ~/.config/kickdesk.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "kickdesk"), nil
}

// ProfilesDir returns ~/.config/kickdesk/profiles.
func ProfilesDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "profiles"), nil
}

// ListProfiles returns sorted profile names: *.json in ~/.config/kickdesk/ and profiles/
// (excludes config.json, the app registry).
func ListProfiles() ([]string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var names []string

	addDir := func(d string) error {
		entries, err := os.ReadDir(d)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasSuffix(name, ".json") || name == registryFilename {
				continue
			}
			base := strings.TrimSuffix(name, ".json")
			if base == "" || seen[base] {
				continue
			}
			seen[base] = true
			names = append(names, base)
		}
		return nil
	}

	if err := addDir(dir); err != nil {
		return nil, err
	}
	profDir, err := ProfilesDir()
	if err == nil {
		if err := addDir(profDir); err != nil {
			return nil, err
		}
	}
	sort.Strings(names)
	return names, nil
}

// ProfilePath resolves a profile name to a JSON file.
// Tries profiles/<name>.json then <name>.json under the config dir.
func ProfilePath(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty profile name")
	}
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	candidates := []string{
		filepath.Join(dir, "profiles", name+".json"),
		filepath.Join(dir, name+".json"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("profile %q not found (looked in %s/profiles/ and %s/)", name, dir, dir)
}

// LoadProfile reads a profile file by name.
func LoadProfile(name string) (*Profile, string, error) {
	path, err := ProfilePath(name)
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, path, fmt.Errorf("read profile %s: %w", path, err)
	}
	var prof Profile
	if err := json.Unmarshal(data, &prof); err != nil {
		return nil, path, fmt.Errorf("parse profile %s: %w", path, err)
	}
	return &prof, path, nil
}

// EphemeralProfile builds an in-memory profile for deprecated -t N (tmux only; menu keeps full registry order).
func EphemeralProfile(topN int) *Profile {
	return &Profile{
		Tmux: TmuxSettings{
			Enabled: true,
			Top:     topN,
			Session: "kickdesk",
		},
	}
}

// ApplyProfile returns a copy of cfg limited to the profile's app_order and labels.
func ApplyProfile(cfg *Config, prof *Profile) *Config {
	if prof == nil {
		return cfg
	}
	out := &Config{
		DefaultProfile: cfg.DefaultProfile,
		Apps:           make(map[string]App, len(prof.AppOrder)),
		AppErrors:      make(map[string]error),
		DisplayLabels:  make(map[string]string),
	}
	if len(prof.Labels) > 0 {
		for k, v := range prof.Labels {
			out.DisplayLabels[k] = v
		}
	}
	order := prof.AppOrder
	if len(order) == 0 {
		order = cfg.OrderedAppNames()
	}
	for _, name := range order {
		if err := cfg.AppErrors[name]; err != nil {
			out.AppErrors[name] = err
			continue
		}
		app, ok := cfg.Apps[name]
		if !ok {
			continue
		}
		if label := prof.Labels[name]; label != "" {
			app.Label = label
		}
		out.Apps[name] = app
	}
	out.AppOrder = order
	return out
}

// TmuxSession returns the tmux session name for this profile.
func (p *Profile) TmuxSession() string {
	if p != nil && p.Tmux.Session != "" {
		return p.Tmux.Session
	}
	return "kickdesk"
}

// SetupConfigHint returns first-time setup steps for the profile picker.
func SetupConfigHint() string {
	dir, err := ConfigDir()
	if err != nil {
		dir = "~/.config/kickdesk"
	}
	return fmt.Sprintf(`Kickdesk setup

  Registry:     %s/config.json     (app ids + labels — see examples/config.json)
  Profiles:     %s/*.json          (layout — see examples/profiles/)
  App manifests: ~/.config/<id>/manifest.json (includes path; see manifest.sample.json)

  cp examples/config.json %s/config.json
  cp examples/profiles/two-up.json %s/two-up.json
  kickdesk -c              # pick profile (hotkeys); saves last.cnfg

Each app must publish a manifest — see docs/MANIFEST.md
`, dir, dir, dir, dir)
}
