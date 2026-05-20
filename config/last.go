package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	registryFilename = "config.json"
	lastConfigFile   = "last.cnfg"
)

// LastConfigPath returns ~/.config/kickdesk/last.cnfg.
func LastConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, lastConfigFile), nil
}

// ReadLastConfig returns the profile name stored in last.cnfg (empty if missing).
func ReadLastConfig() (string, error) {
	path, err := LastConfigPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// WriteLastConfig records the profile name for the next plain `kickdesk` run.
func WriteLastConfig(profileName string) error {
	path, err := LastConfigPath()
	if err != nil {
		return err
	}
	name := strings.TrimSpace(profileName)
	if name == "" {
		return fmt.Errorf("cannot save empty profile to last.cnfg")
	}
	return os.WriteFile(path, []byte(name+"\n"), 0o644)
}
