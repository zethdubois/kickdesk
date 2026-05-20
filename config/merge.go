package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// registryFile is the on-disk config.json before apps are resolved.
type registryFile struct {
	DefaultProfile string                     `json:"default_profile"`
	AppOrder       []string                   `json:"app_order"`
	Apps           map[string]json.RawMessage `json:"apps"`
}

// LoadRegistry reads config.json and resolves manifests into Apps.
func LoadRegistry(path string) (*Config, map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var raw registryFile
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if len(raw.Apps) == 0 {
		return nil, nil, fmt.Errorf("config %s: no apps defined", path)
	}

	manifestPaths := make(map[string]string)
	apps := make(map[string]App, len(raw.Apps))

	for name, entry := range raw.Apps {
		app, mpath, err := parseAppEntry(name, entry)
		if err != nil {
			return nil, nil, fmt.Errorf("app %q: %w", name, err)
		}
		apps[name] = app
		if mpath != "" {
			manifestPaths[name] = mpath
		}
	}

	cfg := &Config{
		DefaultProfile: raw.DefaultProfile,
		AppOrder:       raw.AppOrder,
		Apps:           apps,
	}
	return cfg, manifestPaths, nil
}

func parseAppEntry(name string, raw json.RawMessage) (App, string, error) {
	var app App
	if err := json.Unmarshal(raw, &app); err != nil {
		return App{}, "", err
	}

	// Legacy inline: commands or ports defined in central config.
	if len(app.Commands) > 0 || len(app.Ports) > 0 ||
		len(app.Workflows.Start) > 0 || len(app.Workflows.Stop) > 0 {
		if app.Path == "" {
			return App{}, "", fmt.Errorf("missing path")
		}
		return app, "", nil
	}

	ref := AppRef{
		Label:    app.Label,
		Manifest: app.Manifest,
	}
	return resolveFromManifest(name, ref)
}

func resolveFromManifest(registryName string, ref AppRef) (App, string, error) {
	manifestPath, err := discoverManifest(registryName, ref)
	if err != nil {
		return App{}, "", err
	}

	man, err := LoadManifest(manifestPath)
	if err != nil {
		return App{}, "", err
	}
	if man.ID != registryName {
		return App{}, "", fmt.Errorf("manifest id %q does not match registry key %q", man.ID, registryName)
	}
	if man.Path == "" {
		return App{}, "", fmt.Errorf("manifest %s: missing path — re-run publish", manifestPath)
	}

	app := man.ToApp(ref.Label)
	return overlaySparse(app, ref), manifestPath, nil
}

func overlaySparse(base App, ref AppRef) App {
	if ref.Label != "" {
		base.Label = ref.Label
	}
	return base
}

func discoverManifest(registryName string, ref AppRef) (string, error) {
	if ref.Manifest != "" {
		p, err := ExpandPath(ref.Manifest)
		if err != nil {
			return "", err
		}
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("manifest %s: %w", p, err)
		}
		return p, nil
	}

	if env := os.Getenv("KICKDESK_MANIFEST"); env != "" {
		p, err := ExpandPath(env)
		if err != nil {
			return "", err
		}
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve home directory: %w", err)
	}
	xdg := filepath.Join(home, ".config", registryName, "manifest.json")
	if _, err := os.Stat(xdg); err == nil {
		return xdg, nil
	}
	return "", fmt.Errorf(
		"no published manifest at %s — run the app's publish command (see docs/MANIFEST.md)",
		xdg,
	)
}
