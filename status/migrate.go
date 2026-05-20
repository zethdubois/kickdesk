package status

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Kick-Asset-Management/kickdesk/config"
)

const migrateStatusKey = "migrate-status"

func dbPort(app config.App) int {
	for _, p := range app.Ports {
		if p.Role == "db" {
			return p.Port
		}
	}
	return 0
}

func collectMigrate(app config.App, dir string) string {
	port := dbPort(app)
	if port <= 0 || !IsPortOpen(port) {
		return "unavailable"
	}

	if app.Status.Migrate != "" {
		if st := readMigrateStatusFile(app.Status.Migrate); st != "" {
			return st
		}
	}

	shell, ok := app.Commands[migrateStatusKey]
	if !ok {
		return "n/a"
	}
	return runMigrateStatus(dir, shell)
}

func readMigrateStatusFile(path string) string {
	expanded, err := config.ExpandPath(path)
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(expanded)
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(string(data))
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	switch {
	case line == "ok", line == "unavailable":
		return line
	case strings.HasPrefix(line, "pending:"):
		return line
	default:
		return ""
	}
}

func runMigrateStatus(dir, shell string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-lc", shell)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "error"
	}
	line := strings.TrimSpace(string(out))
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	switch {
	case line == "ok", line == "unavailable":
		return line
	case strings.HasPrefix(line, "pending:"):
		return line
	default:
		return "error"
	}
}
