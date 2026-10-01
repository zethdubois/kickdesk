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

// LaunchInTmuxPane runs shell in the tmux pane for this app command.
// Primary "up" uses the app's dashboard pane. Other blocking commands (wiki,
// discovery, …) get a dedicated pane so they do not Ctrl-C the publish server.
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
	script := fmt.Sprintf("cd %q && %s", dir, shell)

	// Stop whatever is in this pane only — never the sibling service panes.
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
