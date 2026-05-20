package tmux

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const paneEnvPrefix = "KICKDESK_TMUX_PANE_"

// PaneEnvKey is the environment variable for a pane's tmux target (e.g. %3).
func PaneEnvKey(name string) string {
	return paneEnvPrefix + strings.ReplaceAll(name, "-", "_")
}

// PaneTargetForApp resolves the tmux -t target for an app (env, then title lookup).
func PaneTargetForApp(app string) (string, error) {
	if id := os.Getenv(PaneEnvKey(app)); id != "" {
		return id, nil
	}
	return FindPaneByTitle(paneName(app))
}

// FindPaneByTitle finds a pane in the current window by its title (-T name).
func FindPaneByTitle(title string) (string, error) {
	win, err := currentWindow()
	if err != nil {
		return "", err
	}
	out, err := runOut("list-panes", "-t", win, "-F", "#{pane_id} #{pane_title}")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		got := strings.Join(parts[1:], " ")
		if got == title {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("pane %q not found", title)
}

// CurrentPaneTitle returns this pane's title.
func CurrentPaneTitle() (string, error) {
	out, err := runOut("display-message", "-p", "#{pane_title}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// TryAdoptDashboard enables child mode when running kickdesk manually in the menu pane.
func TryAdoptDashboard() bool {
	if InChild() || os.Getenv("TMUX") == "" {
		return false
	}
	title, err := CurrentPaneTitle()
	if err != nil || title != "kickdesk-menu" {
		return false
	}
	servers, err := serverPanesInCurrentWindow()
	if err != nil || len(servers) == 0 {
		return false
	}
	_ = os.Setenv(envChild, "1")
	_ = os.Setenv(envTop, strconv.Itoa(len(servers)))
	for app, paneID := range servers {
		_ = os.Setenv(PaneEnvKey(app), paneID)
	}
	if menuID, err := FindPaneByTitle("kickdesk-menu"); err == nil {
		_ = os.Setenv(PaneEnvKey("menu"), menuID)
	}
	return true
}

func serverPanesInCurrentWindow() (map[string]string, error) {
	win, err := currentWindow()
	if err != nil {
		return nil, err
	}
	out, err := runOut("list-panes", "-t", win, "-F", "#{pane_id} #{pane_title}")
	if err != nil {
		return nil, err
	}
	servers := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		title := strings.Join(parts[1:], " ")
		const prefix = "kickdesk-"
		if !strings.HasPrefix(title, prefix) || title == "kickdesk-menu" {
			continue
		}
		app := strings.TrimPrefix(title, prefix)
		servers[app] = parts[0]
	}
	return servers, nil
}

func buildChildMenuCmd(topN int, bin, profileName string, layout layoutResult) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("%s=1", envChild))
	parts = append(parts, fmt.Sprintf("%s=%d", envTop, topN))
	if profileName != "" {
		parts = append(parts, fmt.Sprintf("%s=%s", envProfile, profileName))
	}
	for app, id := range layout.apps {
		parts = append(parts, fmt.Sprintf("%s=%s", PaneEnvKey(app), id))
	}
	parts = append(parts, fmt.Sprintf("%s=%s", PaneEnvKey("menu"), layout.menuID))
	parts = append(parts, shellQuote(bin))
	if profileName != "" {
		parts = append(parts, fmt.Sprintf("%s=%s", envProfile, profileName))
	}
	parts = append(parts, "menu")
	return strings.Join(parts, " ")
}

type layoutResult struct {
	menuID string
	apps   map[string]string // app name -> pane id (%N)
}
