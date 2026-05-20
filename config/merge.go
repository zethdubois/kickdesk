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
	if app.Path == "" {
		return App{}, "", fmt.Errorf("missing path")
	}

	// Legacy inline: commands or ports defined in central config.
	if len(app.Commands) > 0 || len(app.Ports) > 0 ||
		len(app.Workflows.Start) > 0 || len(app.Workflows.Stop) > 0 {
		return app, "", nil
	}

	ref := AppRef{
		Path:     app.Path,
		Label:    app.Label,
		Manifest: app.Manifest,
	}
	return resolveFromManifest(name, ref)
}

func resolveFromManifest(registryName string, ref AppRef) (App, string, error) {
	resolvedPath, err := ExpandPath(ref.Path)
	if err != nil {
		return App{}, "", err
	}

	manifestPath, err := discoverManifest(registryName, ref, resolvedPath)
	if err != nil {
		return App{}, "", err
	}
	if manifestPath == "" {
		return App{}, "", fmt.Errorf("no manifest found; add commands inline or publish a manifest")
	}

	man, err := LoadManifest(manifestPath)
	if err != nil {
		return App{}, "", err
	}
	if man.ID != registryName {
		return App{}, "", fmt.Errorf("manifest id %q does not match registry key %q", man.ID, registryName)
	}

	app := man.ToApp(ref.Path, ref.Label)
	return overlaySparse(app, ref), manifestPath, nil
}

func overlaySparse(base App, ref AppRef) App {
	// Registry only overrides path/label/manifest; manifest is authoritative for ops.
	if ref.Label != "" {
		base.Label = ref.Label
	}
	if ref.Path != "" {
		base.Path = ref.Path
	}
	return base
}

func discoverManifest(registryName string, ref AppRef, repoPath string) (string, error) {
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
	if err == nil {
		xdg := filepath.Join(home, ".config", registryName, "manifest.json")
		if _, err := os.Stat(xdg); err == nil {
			return xdg, nil
		}
	}

	candidates := []string{
		filepath.Join(repoPath, ".kickdesk.json"),
		filepath.Join(repoPath, ".kickdesk", "manifest.json"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", nil
}
