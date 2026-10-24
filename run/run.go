package run

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Kick-Asset-Management/kickdesk/config"
	"github.com/Kick-Asset-Management/kickdesk/tmux"
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
	if IsBlocking(key, shell) {
		// Primary "up" → dashboard pane. Side services → GUI terminal when a
		// display exists; otherwise a new tmux window (tab) so headless
		// thin_host sessions still run the server without splitting the layout.
		if UsesAppDashboardPane(key) {
			sent, tmuxErr := LaunchInTmuxPane(cfg, appName, key, dir, shell)
			if tmuxErr != nil {
				return tmuxErr
			}
			if sent {
				fmt.Printf("(started in tmux pane: %s)\n", appName)
				return nil
			}
		}
		opened, launchErr := LaunchInTerminal(dir, shell)
		if launchErr != nil {
			return launchErr
		}
		if opened {
			fmt.Println("(started in new terminal)")
			return nil
		}
		sent, winErr := LaunchInTmuxWindow(appName, key, dir, shell)
		if winErr != nil {
			return winErr
		}
		if sent {
			fmt.Printf("(started in tmux window: %s)\n", tmux.PaneTitle(appName, key))
			return nil
		}
		fmt.Fprintln(os.Stderr, "warning: no display and no tmux window; running inline (Ctrl+C returns to menu)")
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

// UsesAppDashboardPane reports whether a blocking command should run in the
// app's primary tmux pane (kickdesk-<app>). All other blocking commands open
// in a new OS terminal instead of splitting that pane.
func UsesAppDashboardPane(key string) bool {
	return key == "" || key == "up"
}

// BlockingHint is a short UI suffix for a blocking command (leading spaces included).
func BlockingHint(key string) string {
	if UsesAppDashboardPane(key) {
		if _, ok := TmuxChildMode(); ok {
			return "  (tmux pane)"
		}
	}
	if HasGUIDisplay() {
		return "  (new terminal)"
	}
	if _, ok := TmuxChildMode(); ok {
		return "  (tmux window)"
	}
	return "  (new terminal)"
}

// IsBlocking reports whether a command is likely long-running.
func IsBlocking(key, shell string) bool {
	switch key {
	case "up", "gateway", "wiki", "standalone", "discovery", "progress":
		return true
	}
	lower := strings.ToLower(shell)
	for _, sub := range []string{
		" dev", "dev ", "run ", " serve", "serve-",
		"uvicorn", "vite",
		"kam:wiki", "kam:discovery", "kam:progress",
	} {
		if strings.Contains(lower, sub) {
			return true
		}
	}
	return false
}
