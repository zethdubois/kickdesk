package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Manifest is the per-app operational config published to ~/.config/<id>/.
type Manifest struct {
	ID          string            `json:"id"`
	Path        string            `json:"path"`
	PortFamily  int               `json:"port_family"`
	PrimaryPort int               `json:"primary_port"`
	Ports       PortsList         `json:"ports"`
	URL         string            `json:"url"`
	Workflows   Workflows         `json:"workflows"`
	Commands    map[string]string `json:"commands"`
	Status      ManifestStatus    `json:"status"`
}

// ManifestStatus holds optional status file paths.
type ManifestStatus struct {
	Migrate string `json:"migrate"`
}

// AppRef is the thin registry entry (optional label + optional manifest override for dev).
type AppRef struct {
	Label    string `json:"label"`
	Manifest string `json:"manifest"`
}

// LoadManifest reads and validates a manifest file.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest %s: %w", path, err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest %s: %w", path, err)
	}
	if m.ID == "" {
		return nil, fmt.Errorf("manifest %s: missing id", path)
	}
	return &m, nil
}

// ToApp merges manifest fields into a runtime App; path comes from the manifest.
func (m *Manifest) ToApp(label string) App {
	return App{
		Path:        m.Path,
		Label:       label,
		PortFamily:  m.PortFamily,
		PrimaryPort: m.PrimaryPort,
		Ports:       m.Ports,
		URL:         m.URL,
		Workflows:   m.Workflows,
		Commands:    m.Commands,
		Status: AppStatusOpts{
			Migrate: m.Status.Migrate,
		},
	}
}

// DisplayName returns label if set, otherwise appName.
func (a *App) DisplayName(appName string) string {
	if a.Label != "" {
		return a.Label
	}
	return appName
}
