package run

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Kick-Asset-Management/kickdesk/config"
	"github.com/Kick-Asset-Management/kickdesk/tmux"
)

const (
	envChild = "KICKDESK_TMUX_CHILD"
	envTop   = "KICKDESK_TMUX_TOP"
)

// TmuxChildMode reports whether blocking commands should target tmux server panes.
func TmuxChildMode() (topN int, ok bool) {
	if os.Getenv(envChild) != "1" {
		return 0, false
	}
	n, err := strconv.Atoi(os.Getenv(envTop))
	if err != nil || n < 1 {
		return 0, false
	}
	return n, true
}

// LaunchInTmuxPane runs shell in the named server pane for appName.
// Returns true if the command was sent, false if app has no top-row pane.
func LaunchInTmuxPane(cfg *config.Config, appName, dir, shell string) (bool, error) {
	topN, ok := TmuxChildMode()
	if !ok {
		return false, nil
	}
	idx := -1
	for i, name := range cfg.OrderedAppNames() {
		if name == appName {
			idx = i
			break
		}
	}
	if idx < 0 || idx >= topN {
		return false, nil
	}

	target, err := tmux.PaneTargetForApp(appName)
	if err != nil {
		return false, nil
	}
	script := fmt.Sprintf("cd %q && %s", dir, shell)

	// Stop any running process in the pane, then start the command.
	_ = tmuxRun("send-keys", "-t", target, "C-c", "")
	if err := tmuxRun("send-keys", "-t", target, script, "C-m"); err != nil {
		return false, err
	}
	return true, nil
}

func tmuxRun(args ...string) error {
	cmd := exec.Command("tmux", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}
