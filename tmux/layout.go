package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	envChild    = "KICKDESK_TMUX_CHILD"
	envTop      = "KICKDESK_TMUX_TOP"
	envProfile  = "KICKDESK_PROFILE"
	defaultSess = "kickdesk"
)

// Active reports whether this process is attached to a tmux client.
func Active() bool {
	return os.Getenv("TMUX") != ""
}

// InChild reports whether this process is the menu pane inside a dashboard layout.
func InChild() bool {
	return os.Getenv(envChild) == "1"
}

// TopN returns the number of server panes when InChild is true.
func TopN() int {
	n, _ := strconv.Atoi(os.Getenv(envTop))
	return n
}

// BootstrapOpts configures tmux dashboard bootstrap.
type BootstrapOpts struct {
	AppNames    []string
	TopN        int
	Session     string
	ProfileName string
	// AppPaths maps app id → expanded checkout path (manifest path) for server pane cwd.
	AppPaths map[string]string
}

// Bootstrap builds the N+1 pane layout and attaches (or re-layouts the current window).
func Bootstrap(opts BootstrapOpts) error {
	topN := opts.TopN
	appNames := opts.AppNames
	if topN < 1 || topN > 9 {
		return fmt.Errorf("tmux top row must be 1-9, got %d", topN)
	}
	if topN > len(appNames) {
		return fmt.Errorf("tmux.top %d requires at least %d apps in app_order", topN, len(appNames))
	}
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("tmux not found on PATH: install tmux or use a profile without tmux")
	}

	sess := opts.Session
	if sess == "" {
		sess = defaultSess
	}

	bin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve kickdesk binary: %w", err)
	}
	bin, _ = filepath.Abs(bin)

	apps := appNames[:topN]

	if Active() {
		return fmt.Errorf("already inside tmux — layout is unchanged; run kickdesk in the kickdesk-menu pane")
	}
	if !ConfirmStartTmux() {
		return fmt.Errorf("cancelled")
	}
	return bootstrapNewSession(opts, apps, topN, bin, sess)
}

func bootstrapNewSession(opts BootstrapOpts, apps []string, topN int, bin, sess string) error {
	profileName := opts.ProfileName
	_ = run("kill-session", "-t", sess)

	w, h := termSize()
	if err := run("new-session", "-d", "-s", sess, "-n", "main",
		"-x", strconv.Itoa(w), "-y", strconv.Itoa(h)); err != nil {
		return err
	}
	target := sess + ":0"
	layout, err := applyLayout(target, apps)
	if err != nil {
		return err
	}
	if err := cdServerPanes(layout, opts.AppPaths); err != nil {
		return err
	}
	menuRoot, err := RepoRoot()
	if err != nil {
		return err
	}
	if err := cdPane(layout.menuID, menuRoot); err != nil {
		return fmt.Errorf("cd menu pane: %w", err)
	}
	menuCmd := buildChildMenuCmd(topN, bin, profileName, layout, menuRoot)
	if err := run("send-keys", "-t", layout.menuID, menuCmd, "C-m"); err != nil {
		return err
	}
	cmd := exec.Command("tmux", "attach", "-t", sess)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// cdServerPanes sends cd to each app pane so git/file work starts in the manifest path.
func cdServerPanes(layout layoutResult, paths map[string]string) error {
	if len(paths) == 0 {
		return nil
	}
	for app, paneID := range layout.apps {
		dir, ok := paths[app]
		if !ok || dir == "" {
			continue
		}
		if err := cdPane(paneID, dir); err != nil {
			return fmt.Errorf("cd pane %s: %w", app, err)
		}
	}
	return nil
}

func cdPane(paneID, dir string) error {
	cmd := "cd " + shellQuote(dir)
	return run("send-keys", "-t", paneID, cmd, "C-m")
}

func applyLayout(target string, apps []string) (layoutResult, error) {
	var result layoutResult
	result.apps = make(map[string]string)
	if err := run("select-window", "-t", target); err != nil {
		return result, err
	}

	_, height := windowSize(target)
	splitLines := height / 2
	if splitLines < 5 {
		splitLines = 5
	}
	// Use -l (lines) not -p: detached or small windows often fail with "size missing" on -p.
	if err := run("split-window", "-v", "-l", strconv.Itoa(splitLines)); err != nil {
		if err := run("split-window", "-v"); err != nil {
			return result, fmt.Errorf("split window: %w", err)
		}
	}

	panes, err := listPanes(target)
	if err != nil {
		return result, err
	}
	if len(panes) < 2 {
		return result, fmt.Errorf("expected at least 2 panes after split, got %d", len(panes))
	}

	// Bottom pane has the largest pane_top (menu).
	sort.Slice(panes, func(i, j int) bool {
		if panes[i].top == panes[j].top {
			return panes[i].index < panes[j].index
		}
		return panes[i].top < panes[j].top
	})
	menu := panes[len(panes)-1]
	var top []paneInfo
	for _, p := range panes[:len(panes)-1] {
		top = append(top, p)
	}

	if err := run("select-pane", "-t", menu.id, "-T", "kickdesk-menu"); err != nil {
		return result, err
	}
	result.menuID = menu.id

	if len(apps) == 1 {
		if err := run("select-pane", "-t", top[0].id, "-T", paneName(apps[0])); err != nil {
			return result, err
		}
		result.apps[apps[0]] = top[0].id
		return result, nil
	}

	// Split top row horizontally until we have len(apps) panes.
	for len(top) < len(apps) {
		rightmost := top[len(top)-1]
		if err := run("select-pane", "-t", rightmost.id); err != nil {
			return result, err
		}
		width, _ := windowSize(target)
		colLines := width / len(apps)
		if colLines < 20 {
			colLines = 20
		}
		if err := run("split-window", "-h", "-l", strconv.Itoa(colLines), "-t", rightmost.id); err != nil {
			if err := run("split-window", "-h", "-t", rightmost.id); err != nil {
				return result, err
			}
		}
		panes, err = listPanes(target)
		if err != nil {
			return result, err
		}
		top, menu, err = splitTopMenu(panes)
		if err != nil {
			return result, err
		}
		_ = run("select-pane", "-t", menu.id, "-T", "kickdesk-menu")
	}

	top = top[:len(apps)]
	for i, app := range apps {
		if err := run("select-pane", "-t", top[i].id, "-T", paneName(app)); err != nil {
			return result, err
		}
		result.apps[app] = top[i].id
	}
	return result, run("select-pane", "-t", menu.id)
}

type paneInfo struct {
	index string
	id    string
	top   int
}

func listPanes(target string) ([]paneInfo, error) {
	out, err := runOut("list-panes", "-t", target, "-F", "#{pane_index} #{pane_id} #{pane_top}")
	if err != nil {
		return nil, err
	}
	var panes []paneInfo
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		t, _ := strconv.Atoi(parts[2])
		panes = append(panes, paneInfo{index: parts[0], id: parts[1], top: t})
	}
	return panes, nil
}

func splitTopMenu(panes []paneInfo) (top []paneInfo, menu paneInfo, err error) {
	if len(panes) < 2 {
		return nil, paneInfo{}, fmt.Errorf("not enough panes")
	}
	sort.Slice(panes, func(i, j int) bool {
		if panes[i].top == panes[j].top {
			return panes[i].index < panes[j].index
		}
		return panes[i].top < panes[j].top
	})
	menu = panes[len(panes)-1]
	for _, p := range panes[:len(panes)-1] {
		top = append(top, p)
	}
	return top, menu, nil
}

func paneName(app string) string {
	return "kickdesk-" + app
}

func currentWindow() (string, error) {
	out, err := runOut("display-message", "-p", "#{session_name}:#{window_index}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func run(args ...string) error {
	cmd := exec.Command("tmux", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func runOut(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("tmux %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// termSize returns cols x rows for a new detached session (avoids 0-size "size missing").
func termSize() (width, height int) {
	width, height = 220, 50
	if Active() {
		if w, h, ok := tmuxClientSize(); ok {
			return w, h
		}
	}
	if w, h, ok := sttySize(); ok {
		if w > 0 {
			width = w
		}
		if h > 0 {
			height = h
		}
	}
	return width, height
}

func tmuxClientSize() (w, h int, ok bool) {
	wOut, err := runOut("display-message", "-p", "#{client_width}")
	if err != nil {
		return 0, 0, false
	}
	hOut, err := runOut("display-message", "-p", "#{client_height}")
	if err != nil {
		return 0, 0, false
	}
	w, _ = strconv.Atoi(strings.TrimSpace(wOut))
	h, _ = strconv.Atoi(strings.TrimSpace(hOut))
	if w < 40 || h < 10 {
		return 0, 0, false
	}
	return w, h, true
}

func windowSize(target string) (width, height int) {
	width, height = 220, 50
	wOut, err := runOut("display-message", "-t", target, "-p", "#{window_width}")
	if err == nil {
		if w, e := strconv.Atoi(strings.TrimSpace(wOut)); e == nil && w > 0 {
			width = w
		}
	}
	hOut, err := runOut("display-message", "-t", target, "-p", "#{window_height}")
	if err == nil {
		if h, e := strconv.Atoi(strings.TrimSpace(hOut)); e == nil && h > 0 {
			height = h
		}
	}
	return width, height
}

func sttySize() (width, height int, ok bool) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, false
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) != 2 {
		return 0, 0, false
	}
	h, err1 := strconv.Atoi(fields[0])
	w, err2 := strconv.Atoi(fields[1])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return w, h, true
}
