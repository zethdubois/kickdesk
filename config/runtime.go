package config

import "fmt"

// Runtime is the resolved cockpit: registry + profile + merged apps.
type Runtime struct {
	Registry      *Config
	Config        *Config
	Profile       *Profile
	ProfileName   string
	ProfilePath   string
	ConfigPath    string
	ManifestPaths map[string]string
}

// LoadRuntime loads registry, applies profile, and returns a Runtime.
// profileName empty uses KICKDESK_PROFILE, then default_profile (if useDefault), then no profile filter.
func LoadRuntime(profileName string) (*Runtime, error) {
	return LoadRuntimeOptions(profileName, true)
}

// LoadRuntimeOptions loads registry and optional profile by name.
// useDefault is legacy; when false and profileName empty, all registry apps are shown.
func LoadRuntimeOptions(profileName string, useDefault bool) (*Runtime, error) {
	_ = useDefault

	cfgPath, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	cfg, manifestPaths, err := LoadRegistry(cfgPath)
	if err != nil {
		return nil, err
	}

	rt := &Runtime{
		Registry:      cfg,
		Config:        cfg,
		ConfigPath:    cfgPath,
		ManifestPaths: manifestPaths,
		ProfileName:   profileName,
	}

	if profileName != "" {
		prof, path, err := LoadProfile(profileName)
		if err != nil {
			return nil, err
		}
		rt.Profile = prof
		rt.ProfilePath = path
		rt.Config = ApplyProfile(cfg, prof)
	}

	return rt, nil
}

// WithEphemeralTmux applies deprecated -t N as a one-off tmux overlay (does not filter apps).
func (r *Runtime) WithEphemeralTmux(topN int) *Runtime {
	if topN < 1 {
		return r
	}
	cp := *r
	cp.Profile = EphemeralProfile(topN)
	cp.ProfileName = ""
	cp.ProfilePath = ""
	return &cp
}

// TmuxEnabled reports whether dashboard bootstrap should run.
func (r *Runtime) TmuxEnabled() bool {
	if r.Profile == nil {
		return false
	}
	return r.Profile.Tmux.Enabled || r.Profile.Tmux.Top > 0
}

// TmuxTop returns server pane count for tmux layout.
func (r *Runtime) TmuxTop() int {
	if r.Profile == nil {
		return 0
	}
	top := r.Profile.Tmux.Top
	if top < 1 && r.Profile.Tmux.Enabled {
		top = len(r.Config.OrderedAppNames())
	}
	return top
}

// TmuxSessionName returns the tmux session for bootstrap.
func (r *Runtime) TmuxSessionName() string {
	if r.Profile != nil {
		return r.Profile.TmuxSession()
	}
	return "kickdesk"
}

// Validate checks the resolved runtime (profile + apps + manifests).
func (r *Runtime) Validate() []ValidationError {
	var errs []ValidationError
	if r.Profile != nil {
		errs = append(errs, r.ValidateProfile()...)
	}
	errs = append(errs, r.Registry.Validate()...)
	return errs
}

// ValidateProfile checks profile app_order and tmux settings against the registry.
func (r *Runtime) ValidateProfile() []ValidationError {
	var errs []ValidationError
	p := r.Profile
	top := p.Tmux.Top
	if p.Tmux.Enabled && top < 1 {
		top = len(p.AppOrder)
	}
	if top > 0 && len(p.AppOrder) > 0 && top > len(p.AppOrder) {
		errs = append(errs, ValidationError{
			Message: fmt.Sprintf("profile %q: tmux.top %d exceeds app_order length %d", r.ProfileName, top, len(p.AppOrder)),
		})
	}
	for _, name := range p.AppOrder {
		if _, ok := r.Registry.Apps[name]; !ok {
			errs = append(errs, ValidationError{
				Message: fmt.Sprintf("profile %q: app_order references unknown app %q", r.ProfileName, name),
			})
		}
	}
	return errs
}

// ValidateRegistry checks all apps in the registry (before profile filter).
func ValidateRegistry(cfgPath string, onlyApp string) ([]ValidationError, map[string]string, error) {
	cfg, manifestPaths, err := LoadRegistry(cfgPath)
	if err != nil {
		return nil, nil, err
	}
	if onlyApp != "" {
		app, ok := cfg.Apps[onlyApp]
		if !ok {
			return []ValidationError{{App: onlyApp, Message: "unknown app"}}, manifestPaths, nil
		}
		cfg = &Config{Apps: map[string]App{onlyApp: app}}
		manifestPaths = map[string]string{onlyApp: manifestPaths[onlyApp]}
	}
	return cfg.Validate(), manifestPaths, nil
}
