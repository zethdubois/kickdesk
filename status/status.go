package status

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kick-Asset-Management/kickdesk/config"
)

// PortLine is one host port with role and listen state.
type PortLine struct {
	Role  string
	Port  int
	State string
}

func (p PortLine) String() string {
	label := fmt.Sprintf("%d", p.Port)
	if p.Role != "" {
		label = fmt.Sprintf("%s:%d", p.Role, p.Port)
	}
	return fmt.Sprintf("%s > %s", label, p.State)
}

// AppStatus is the status snapshot for one app.
type AppStatus struct {
	Name       string // registry id (menu hotkeys)
	Label      string // display name in table when set
	PortFamily string
	Ports      []PortLine
	Migrate    string
	Branch     string
	Dirty      string
	URL        string
	Running    bool
	MainPort   int
}

// Row is a legacy aggregate for callers that still expect a single port string.
type Row struct {
	Name       string
	PortFamily string
	PortState  string
	Branch     string
	Dirty      string
	URL        string
	Running    bool
	MainPort   int
}

// Collect builds status for every app in app_order.
func Collect(cfg *config.Config) ([]AppStatus, error) {
	var out []AppStatus
	for _, name := range cfg.OrderedAppNames() {
		app, ok := cfg.Apps[name]
		if !ok {
			continue
		}
		st, err := collectApp(name, app)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, nil
}

// CollectRows converts AppStatus to legacy Row slice (for compatibility).
func CollectRows(apps []AppStatus) []Row {
	rows := make([]Row, len(apps))
	for i, a := range apps {
		parts := make([]string, len(a.Ports))
		for j, p := range a.Ports {
			parts[j] = p.String()
		}
		rows[i] = Row{
			Name:       a.Name,
			PortFamily: a.PortFamily,
			PortState:  strings.Join(parts, " "),
			Branch:     a.Branch,
			Dirty:      a.Dirty,
			URL:        a.URL,
			Running:    a.Running,
			MainPort:   a.MainPort,
		}
	}
	return rows
}

// AppRunning reports whether the app's main port is listening.
func AppRunning(cfg *config.Config, appName string) bool {
	app, ok := cfg.Apps[appName]
	if !ok {
		return false
	}
	return IsPortOpen(app.MainPort())
}

// IsPortOpen returns true if something is listening on 127.0.0.1:port.
func IsPortOpen(port int) bool {
	if port <= 0 {
		return false
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 150*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func portLines(bindings config.PortsList) []PortLine {
	out := make([]PortLine, 0, len(bindings))
	for _, b := range bindings {
		state := "down"
		if IsPortOpen(b.Port) {
			state = "up"
		}
		out = append(out, PortLine{Role: b.Role, Port: b.Port, State: state})
	}
	return out
}

func collectApp(name string, app config.App) (AppStatus, error) {
	path, err := app.ResolvedPath()
	if err != nil {
		return AppStatus{}, err
	}
	mainPort := app.MainPort()
	branch, dirty := gitInfo(path)
	return AppStatus{
		Name:       name,
		Label:      app.DisplayName(name),
		PortFamily: config.FamilyLabel(app.PortFamily),
		Ports:      portLines(app.Ports),
		Migrate:    collectMigrate(app, path),
		Branch:     branch,
		Dirty:      dirty,
		URL:        app.URL,
		Running:    IsPortOpen(mainPort),
		MainPort:   mainPort,
	}, nil
}

func gitInfo(dir string) (branch, dirty string) {
	branch = "—"
	dirty = "—"
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		return branch, dirty
	}
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err == nil {
		branch = strings.TrimSpace(string(out))
	}
	porcelain, err := exec.Command("git", "-C", dir, "status", "--porcelain").Output()
	if err == nil {
		if len(strings.TrimSpace(string(porcelain))) > 0 {
			dirty = "dirty"
		} else {
			dirty = "clean"
		}
	}
	return branch, dirty
}

func formatGit(branch, dirty string) string {
	if branch == "—" {
		return "—"
	}
	if dirty != "—" {
		return fmt.Sprintf("%s (%s)", branch, dirty)
	}
	return branch
}

// Print writes the bordered status table (kickdesk status).
func Print(apps []AppStatus) {
	renderTable(apps, false, "")
}

// PrintHub prints the status table with hotkey column [1], [2], ...
// focusApp highlights that app's row(s) when non-empty (e.g. active submenu).
func PrintHub(apps []AppStatus, focusApp string) {
	renderTable(apps, true, focusApp)
}

// PrintAppSummary prints one app's ports (for workflow screen header).
func PrintAppSummary(cfg *config.Config, appName string) error {
	app, ok := cfg.Apps[appName]
	if !ok {
		return fmt.Errorf("unknown app %q", appName)
	}
	st, err := collectApp(appName, app)
	if err != nil {
		return err
	}
	running := st.Running
	state := "STOPPED"
	if running {
		state = "RUNNING"
	}
	stStyle := stdoutStyle()
	stateCol := stStyle.red + state + stStyle.reset
	if running {
		stateCol = stStyle.bold + stStyle.green + state + stStyle.reset
	}
	fmt.Printf("%s — %s\n", appName, stateCol)
	renderAppPortTable(st.Ports)
	if st.URL != "" {
		fmt.Printf("url: %s\n", st.URL)
	}
	fmt.Println()
	return nil
}
