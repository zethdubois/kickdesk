package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Kick-Asset-Management/kickdesk/config"
	"github.com/Kick-Asset-Management/kickdesk/menu"
	"github.com/Kick-Asset-Management/kickdesk/run"
	"github.com/Kick-Asset-Management/kickdesk/status"
	"github.com/Kick-Asset-Management/kickdesk/tmux"
)

const version = "0.1.0-dev"

type cliOpts struct {
	profile       string
	profilePicker bool // -c: interactive profile menu
	profileChoice int  // -c N: pick by number without menu (-1 = unset)
	tmuxTop       int  // deprecated; maps to ephemeral tmux overlay
	args          []string
}

func main() {
	opts := parseGlobalFlags(os.Args[1:])
	if len(opts.args) == 0 {
		runMenu(opts)
		return
	}

	switch opts.args[0] {
	case "menu":
		runMenu(opts)
	case "status":
		runStatus(opts)
	case "run":
		runCommand(opts, opts.args[1:])
	case "config":
		if len(opts.args) >= 2 && opts.args[1] == "validate" {
			runConfigValidate(opts, opts.args[2:])
			return
		}
		printUsage()
		os.Exit(1)
	default:
		printUsage()
		os.Exit(1)
	}
}

func parseGlobalFlags(args []string) cliOpts {
	opts := cliOpts{profileChoice: -1, args: args}
	for len(opts.args) > 0 && strings.HasPrefix(opts.args[0], "-") {
		switch opts.args[0] {
		case "-c":
			opts.args = opts.args[1:]
			if len(opts.args) > 0 {
				if n, err := strconv.Atoi(opts.args[0]); err == nil && n >= 0 && n <= 9 {
					opts.profileChoice = n
					opts.args = opts.args[1:]
					break
				}
			}
			opts.profilePicker = true
		case "-t":
			if len(opts.args) < 2 {
				fmt.Fprintln(os.Stderr, "kickdesk: -t requires a number (deprecated; use kickdesk -c <profile>)")
				os.Exit(1)
			}
			n, err := strconv.Atoi(opts.args[1])
			if err != nil || n < 1 || n > 9 {
				fmt.Fprintln(os.Stderr, "kickdesk: -t must be a number from 1 to 9")
				os.Exit(1)
			}
			fmt.Fprintln(os.Stderr, "kickdesk: warning: -t is deprecated; use a profile with tmux.top (kickdesk -c two-up)")
			opts.tmuxTop = n
			opts.args = opts.args[2:]
		default:
			fmt.Fprintf(os.Stderr, "kickdesk: unknown flag %q\n", opts.args[0])
			os.Exit(1)
		}
	}
	return opts
}

// loadRuntimeForCommand uses last.cnfg for subcommands (status, run) when no -c picker ran.
func loadRuntimeForCommand(opts cliOpts) *config.Runtime {
	if opts.profile == "" {
		last, err := config.ReadLastConfig()
		if err != nil || last == "" {
			fmt.Fprintln(os.Stderr, "kickdesk: no last.cnfg — run kickdesk -c to choose a profile first")
			os.Exit(1)
		}
		opts.profile = last
	}
	return loadRuntime(opts, false)
}

func loadRuntime(opts cliOpts, useDefaultProfile bool) *config.Runtime {
	rt, err := config.LoadRuntimeOptions(opts.profile, useDefaultProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
		dir, _ := config.ConfigDir()
		fmt.Fprintf(os.Stderr, "copy examples/config.json to %s/config.json\n", dir)
		os.Exit(1)
	}
	if opts.tmuxTop > 0 {
		rt = rt.WithEphemeralTmux(opts.tmuxTop)
	}
	return rt
}

func runMenu(opts cliOpts) {
	if opts.profilePicker {
		pick, err := menu.RunProfilePicker()
		if err != nil {
			fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
			os.Exit(1)
		}
		if pick.Quit {
			os.Exit(0)
		}
		opts.profile = pick.ProfileName
		if err := config.WriteLastConfig(opts.profile); err != nil {
			fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Using profile %q (saved to last.cnfg)\n\n", opts.profile)
	} else if opts.profileChoice >= 0 {
		name, err := menu.ResolveProfileChoice(opts.profileChoice)
		if err != nil {
			fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
			os.Exit(1)
		}
		if name == "" {
			os.Exit(0)
		}
		opts.profile = name
		if err := config.WriteLastConfig(opts.profile); err != nil {
			fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Using profile %q (saved to last.cnfg)\n\n", opts.profile)
	} else if opts.profile == "" {
		if p := os.Getenv("KICKDESK_PROFILE"); p != "" {
			opts.profile = p
		} else {
			last, err := config.ReadLastConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
				os.Exit(1)
			}
			if last == "" {
				fmt.Fprintln(os.Stderr, "kickdesk: no last.cnfg — run kickdesk -c to choose a profile")
				os.Exit(1)
			}
			opts.profile = last
		}
	}

	rt := loadRuntime(opts, false)
	tmux.TryAdoptDashboard()
	// Only create a new tmux session from outside tmux; never split/relayout the
	// current window when the operator is already in a dashboard.
	if rt.TmuxEnabled() && !tmux.InChild() && !tmux.Active() {
		top := rt.TmuxTop()
		names := rt.Config.OrderedAppNames()
		appPaths := tmuxAppPaths(rt, names, top)
		if err := tmux.Bootstrap(tmux.BootstrapOpts{
			AppNames:    names,
			TopN:        top,
			Session:     rt.TmuxSessionName(),
			ProfileName: rt.ProfileName,
			AppPaths:    appPaths,
		}); err != nil {
			if err.Error() == "cancelled" {
				os.Exit(0)
			}
			fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
			os.Exit(1)
		}
		return
	}
	printManifestHint(rt)
	if err := menu.Run(rt.Config, version); err != nil {
		fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
		os.Exit(1)
	}
}

func tmuxAppPaths(rt *config.Runtime, names []string, topN int) map[string]string {
	paths := make(map[string]string)
	if topN > len(names) {
		topN = len(names)
	}
	for _, name := range names[:topN] {
		if rt.Config.AppErrors[name] != nil {
			continue
		}
		app, ok := rt.Config.Apps[name]
		if !ok {
			continue
		}
		dir, err := app.ResolvedPath()
		if err != nil {
			continue
		}
		paths[name] = dir
	}
	return paths
}

func printManifestHint(rt *config.Runtime) {
	for _, name := range rt.Config.OrderedAppNames() {
		if err := rt.Config.AppErrors[name]; err != nil {
			fmt.Fprintf(os.Stderr, "app %s: %v\n", name, err)
			continue
		}
		if p := rt.ManifestPaths[name]; p != "" {
			fmt.Fprintf(os.Stderr, "manifest %s ← %s\n", name, p)
		}
	}
	if len(rt.ManifestPaths) > 0 || len(rt.Config.AppErrors) > 0 {
		fmt.Fprintln(os.Stderr)
	}
}

func runStatus(opts cliOpts) {
	rt := loadRuntimeForCommand(opts)
	rows, err := status.Collect(rt.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
		os.Exit(1)
	}
	status.Print(rows)
}

func runCommand(opts cliOpts, args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: kickdesk run <app> <key>\n")
		os.Exit(1)
	}
	rt := loadRuntimeForCommand(opts)
	appName := args[0]
	key := args[1]
	if err := run.Execute(rt.Config, appName, key); err != nil {
		fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
		os.Exit(1)
	}
}

func runConfigValidate(opts cliOpts, args []string) {
	onlyApp := ""
	for i := 0; i < len(args); i++ {
		if args[i] == "--app" && i+1 < len(args) {
			onlyApp = args[i+1]
			break
		}
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
		os.Exit(1)
	}

	if onlyApp != "" {
		errs, manifestPaths, err := config.ValidateRegistry(cfgPath, onlyApp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
			os.Exit(1)
		}
		printValidateResult(cfgPath, errs, manifestPaths, onlyApp)
		return
	}

	cfg, manifestPaths, err := config.LoadRegistry(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
		os.Exit(1)
	}

	var errs []config.ValidationError
	errs = append(errs, cfg.Validate()...)

	profileName := opts.profile
	if profileName == "" {
		profileName = cfg.DefaultProfile
	}
	var profilePath string
	if profileName != "" {
		prof, path, perr := config.LoadProfile(profileName)
		if perr != nil {
			errs = append(errs, config.ValidationError{
				Message: fmt.Sprintf("profile %q: %v", profileName, perr),
			})
		} else {
			profilePath = path
			rt := &config.Runtime{
				Registry:      cfg,
				Config:        cfg,
				Profile:       prof,
				ProfileName:   profileName,
				ProfilePath:   path,
				ConfigPath:    cfgPath,
				ManifestPaths: manifestPaths,
			}
			rt.Config = config.ApplyProfile(cfg, prof)
			errs = append(errs, rt.ValidateProfile()...)
		}
	}

	printValidateResult(cfgPath, errs, manifestPaths, "")
	if profileName != "" && profilePath != "" && len(errs) == 0 {
		fmt.Printf("profile: %s (%s)\n", profileName, profilePath)
	}
}

func printValidateResult(cfgPath string, errs []config.ValidationError, manifestPaths map[string]string, onlyApp string) {
	if len(errs) == 0 {
		n := len(manifestPaths)
		if onlyApp != "" {
			if p := manifestPaths[onlyApp]; p != "" {
				fmt.Printf("ok: %s app %s (manifest %s)\n", cfgPath, onlyApp, p)
			} else {
				fmt.Printf("ok: %s app %s (inline config)\n", cfgPath, onlyApp)
			}
			return
		}
		fmt.Printf("ok: %s", cfgPath)
		if n > 0 {
			fmt.Printf(" (%d manifests)", n)
		}
		fmt.Println()
		for name, p := range manifestPaths {
			if p != "" {
				fmt.Printf("  %s: %s\n", name, p)
			}
		}
		return
	}
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "error: %v\n", e)
	}
	os.Exit(1)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: kickdesk [-c] [command]\n")
	fmt.Fprintf(os.Stderr, "options:\n")
	fmt.Fprintf(os.Stderr, "  (none)         menu using ~/.config/kickdesk/last.cnfg\n")
	fmt.Fprintf(os.Stderr, "  -c             choose profile (saves last.cnfg); 0 editor, 1–N *.json\n")
	fmt.Fprintf(os.Stderr, "  -t N           deprecated tmux overlay\n")
	fmt.Fprintf(os.Stderr, "commands:\n")
	fmt.Fprintf(os.Stderr, "  (none)           interactive menu\n")
	fmt.Fprintf(os.Stderr, "  menu             interactive menu\n")
	fmt.Fprintf(os.Stderr, "  status           port and git status table\n")
	fmt.Fprintf(os.Stderr, "  run <app> <key>  run one configured command\n")
	fmt.Fprintf(os.Stderr, "  config validate  check config paths and ports\n")
	fmt.Fprintf(os.Stderr, "  config validate --app <name>  validate one app\n")
}
