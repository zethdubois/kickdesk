package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const requiredCommands = "up, down, build"

var defaultAppOrder = []string{"publicweb", "kickagent", "merch-api"}

// Config is the top-level kickdesk registry.
type Config struct {
	DefaultProfile string         `json:"default_profile"`
	AppOrder       []string       `json:"app_order"`
	Apps           map[string]App `json:"apps"`
	// AppErrors holds per-app manifest load failures (app still listed in app_order).
	AppErrors map[string]error `json:"-"`
	// DisplayLabels are profile display names keyed by app id (including unloadable apps).
	DisplayLabels map[string]string `json:"-"`
}

// App describes one managed repository (inline legacy or merged from manifest).
type App struct {
	Path        string            `json:"path"`
	Label       string            `json:"label"`
	Manifest    string            `json:"manifest"`
	PortFamily  int               `json:"port_family"`
	PrimaryPort int               `json:"primary_port"`
	Ports       PortsList         `json:"ports"`
	URL         string            `json:"url"`
	Workflows   Workflows         `json:"workflows"`
	Commands    map[string]string `json:"commands"`
	Status      AppStatusOpts     `json:"status"`
}

// AppStatusOpts holds optional status sources (manifest status.*).
type AppStatusOpts struct {
	Migrate string `json:"migrate"`
}

// Workflows lists ordered command keys for startup and shutdown.
type Workflows struct {
	Start []string `json:"start"`
	Stop  []string `json:"stop"`
}

// CommandEntry is one named shell command for an app.
type CommandEntry struct {
	App   string
	Key   string
	Shell string
}

// OrderedAppNames returns apps in app_order (or default publicweb, kickagent, merch-api).
func (c *Config) OrderedAppNames() []string {
	if len(c.AppOrder) > 0 {
		return append([]string(nil), c.AppOrder...)
	}
	names := c.AppNames()
	if len(names) == len(defaultAppOrder) {
		// Prefer default order when all default apps present.
		var ordered []string
		for _, name := range defaultAppOrder {
			if _, ok := c.Apps[name]; ok {
				ordered = append(ordered, name)
			}
		}
		for _, name := range names {
			found := false
			for _, o := range ordered {
				if o == name {
					found = true
					break
				}
			}
			if !found {
				ordered = append(ordered, name)
			}
		}
		if len(ordered) > 0 {
			return ordered
		}
	}
	return names
}

// MainPort returns the port used to detect running vs stopped.
func (a *App) MainPort() int {
	if a.PrimaryPort > 0 {
		return a.PrimaryPort
	}
	for _, b := range a.Ports {
		if b.Role == "app" || b.Role == "api" {
			return b.Port
		}
	}
	if len(a.Ports) > 0 {
		return a.Ports[0].Port
	}
	return 0
}

// WorkflowKeys returns start or stop command keys for an app.
func (c *Config) WorkflowKeys(appName string, running bool) ([]string, error) {
	app, ok := c.Apps[appName]
	if !ok {
		return nil, fmt.Errorf("unknown app %q", appName)
	}
	if running {
		if len(app.Workflows.Stop) == 0 {
			return nil, fmt.Errorf("app %q: no workflows.stop defined", appName)
		}
		return append([]string(nil), app.Workflows.Stop...), nil
	}
	if len(app.Workflows.Start) == 0 {
		return nil, fmt.Errorf("app %q: no workflows.start defined", appName)
	}
	return append([]string(nil), app.Workflows.Start...), nil
}

// CommandsList returns sorted command entries for one app.
func (c *Config) CommandsList(appName string) ([]CommandEntry, error) {
	app, ok := c.Apps[appName]
	if !ok {
		return nil, fmt.Errorf("unknown app %q", appName)
	}
	if app.Commands == nil {
		return nil, fmt.Errorf("app %q: no commands", appName)
	}
	keys := make([]string, 0, len(app.Commands))
	for k := range app.Commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]CommandEntry, 0, len(keys))
	for _, k := range keys {
		out = append(out, CommandEntry{App: appName, Key: k, Shell: app.Commands[k]})
	}
	return out, nil
}

// DefaultPath returns ~/.config/kickdesk/config.json (or KICKDESK_CONFIG).
func DefaultPath() (string, error) {
	if p := os.Getenv("KICKDESK_CONFIG"); p != "" {
		return ExpandPath(p)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "kickdesk", "config.json"), nil
}

// ExpandPath expands a leading ~ to the user home directory.
func ExpandPath(p string) (string, error) {
	if p == "~" {
		return os.UserHomeDir()
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}

// Load reads config.json and resolves app manifests.
func Load(path string) (*Config, error) {
	cfg, _, err := LoadRegistry(path)
	return cfg, err
}

// AppNames returns sorted app keys.
func (c *Config) AppNames() []string {
	names := make([]string, 0, len(c.Apps))
	for name := range c.Apps {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ResolvedPath returns the expanded filesystem path for an app.
func (a *App) ResolvedPath() (string, error) {
	return ExpandPath(a.Path)
}

// ValidationError describes one config problem.
type ValidationError struct {
	App     string
	Message string
}

func (e ValidationError) Error() string {
	if e.App != "" {
		return fmt.Sprintf("%s: %s", e.App, e.Message)
	}
	return e.Message
}

func (c *Config) validateWorkflowKeys(appName string, keys []string, label string) []ValidationError {
	var errs []ValidationError
	for _, key := range keys {
		if key == "" {
			errs = append(errs, ValidationError{App: appName, Message: fmt.Sprintf("workflows.%s: empty command key", label)})
			continue
		}
		app := c.Apps[appName]
		if app.Commands == nil || app.Commands[key] == "" {
			errs = append(errs, ValidationError{App: appName, Message: fmt.Sprintf("workflows.%s: unknown command %q", label, key)})
		}
	}
	return errs
}

// Validate checks paths, required commands, workflows, and port uniqueness.
func (c *Config) Validate() []ValidationError {
	var errs []ValidationError
	portOwners := make(map[int][]string)

	for _, name := range c.AppOrder {
		if _, ok := c.Apps[name]; !ok {
			errs = append(errs, ValidationError{Message: fmt.Sprintf("app_order: unknown app %q", name)})
		}
	}

	for _, name := range c.AppNames() {
		app := c.Apps[name]
		if app.Path == "" {
			errs = append(errs, ValidationError{App: name, Message: "missing path"})
			continue
		}
		resolved, err := app.ResolvedPath()
		if err != nil {
			errs = append(errs, ValidationError{App: name, Message: err.Error()})
			continue
		}
		if st, err := os.Stat(resolved); err != nil || !st.IsDir() {
			errs = append(errs, ValidationError{App: name, Message: fmt.Sprintf("path does not exist: %s", resolved)})
		}
		if len(app.Ports) == 0 {
			errs = append(errs, ValidationError{App: name, Message: "ports must not be empty"})
		}
		pp := app.MainPort()
		if pp > 0 {
			found := false
			for _, b := range app.Ports {
				if b.Port == pp {
					found = true
					break
				}
			}
			if !found {
				errs = append(errs, ValidationError{App: name, Message: fmt.Sprintf("primary_port %d not in ports list", pp)})
			}
		}
		for _, b := range app.Ports {
			portOwners[b.Port] = append(portOwners[b.Port], name)
			if why, crowded := CrowdedPorts[b.Port]; crowded {
				errs = append(errs, ValidationError{
					App:     name,
					Message: fmt.Sprintf("port %d is a crowded default (%s); prefer a KAM family port — see docs/DEV_PORTS.md", b.Port, why),
				})
			}
		}
		if app.Commands == nil {
			errs = append(errs, ValidationError{App: name, Message: "missing commands"})
			continue
		}
		for _, cmd := range []string{"up", "down", "build"} {
			if app.Commands[cmd] == "" {
				errs = append(errs, ValidationError{App: name, Message: fmt.Sprintf("missing required command %q", cmd)})
			}
		}
		if len(app.Workflows.Start) == 0 {
			errs = append(errs, ValidationError{App: name, Message: "workflows.start must not be empty"})
		} else {
			errs = append(errs, c.validateWorkflowKeys(name, app.Workflows.Start, "start")...)
		}
		if len(app.Workflows.Stop) == 0 {
			errs = append(errs, ValidationError{App: name, Message: "workflows.stop must not be empty"})
		} else {
			errs = append(errs, c.validateWorkflowKeys(name, app.Workflows.Stop, "stop")...)
		}
	}

	for port, owners := range portOwners {
		if len(owners) > 1 {
			sort.Strings(owners)
			errs = append(errs, ValidationError{
				Message: fmt.Sprintf("port %d claimed by multiple apps: %s", port, strings.Join(owners, ", ")),
			})
		}
	}

	sort.Slice(errs, func(i, j int) bool {
		if errs[i].App != errs[j].App {
			return errs[i].App < errs[j].App
		}
		return errs[i].Message < errs[j].Message
	})
	return errs
}
