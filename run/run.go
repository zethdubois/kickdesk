package run

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Kick-Asset-Management/kickdesk/config"
)

// Execute runs commands.<key> for app in its repo directory.
func Execute(cfg *config.Config, appName, key string) error {
	app, ok := cfg.Apps[appName]
	if !ok {
		return fmt.Errorf("unknown app %q", appName)
	}
	shell, ok := app.Commands[key]
	if !ok {
		return fmt.Errorf("app %q: unknown command %q", appName, key)
	}
	dir, err := app.ResolvedPath()
	if err != nil {
		return err
	}
	return executeShell(dir, shell)
}

// ExecuteSequence runs multiple command keys in order; stops on first error.
func ExecuteSequence(cfg *config.Config, appName string, keys []string) error {
	for _, key := range keys {
		if err := Execute(cfg, appName, key); err != nil {
			return fmt.Errorf("%s/%s: %w", appName, key, err)
		}
	}
	return nil
}

func executeShell(dir, shell string) error {
	cmd := exec.Command("bash", "-lc", shell)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("exit code %d", exit.ExitCode())
		}
		return err
	}
	return nil
}

// IsBlocking reports whether a command is likely long-running.
func IsBlocking(key, shell string) bool {
	switch key {
	case "up", "gateway":
		return true
	}
	lower := strings.ToLower(shell)
	for _, sub := range []string{" dev", "dev ", "run ", " serve", "serve-", "uvicorn", "vite"} {
		if strings.Contains(lower, sub) {
			return true
		}
	}
	return false
}
