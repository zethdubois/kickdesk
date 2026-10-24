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

// LaunchInTmuxPane runs shell in the app's primary dashboard pane (kickdesk-<app>).
// Callers should only use this for primary "up". Side services use
// LaunchInTerminal or LaunchInTmuxWindow.
func LaunchInTmuxPane(cfg *config.Config, appName, cmdKey, dir, shell string) (bool, error) {
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

	target, err := tmux.PaneTargetForCommand(appName, cmdKey)
	if err != nil {
		return false, nil
	}
	return sendToTmuxTarget(target, dir, shell)
}

// LaunchInTmuxWindow runs shell in a dedicated tmux window (tab) named for the
// app command. Used for side services when no GUI display is available.
func LaunchInTmuxWindow(appName, cmdKey, dir, shell string) (bool, error) {
	if _, ok := TmuxChildMode(); !ok && !tmux.Active() {
		return false, nil
	}
	target, err := tmux.WindowTargetForCommand(appName, cmdKey)
	if err != nil {
		return false, nil
	}
	return sendToTmuxTarget(target, dir, shell)
}

func sendToTmuxTarget(target, dir, shell string) (bool, error) {
	script := fmt.Sprintf("cd %q && %s", dir, shell)
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
