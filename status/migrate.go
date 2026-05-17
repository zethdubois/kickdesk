package status

import (
	"context"
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
	shell, ok := app.Commands[migrateStatusKey]
	if !ok {
		return "n/a"
	}
	port := dbPort(app)
	if port <= 0 || !IsPortOpen(port) {
		return "unavailable"
	}
	return runMigrateStatus(dir, shell)
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
