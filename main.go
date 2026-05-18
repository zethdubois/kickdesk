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
	tmuxTop int
	args    []string
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
		runStatus()
	case "run":
		runCommand(opts.args[1:])
	case "config":
		if len(opts.args) >= 2 && opts.args[1] == "validate" {
			runConfigValidate()
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
	opts := cliOpts{args: args}
	for len(opts.args) > 0 && strings.HasPrefix(opts.args[0], "-") {
		switch opts.args[0] {
		case "-t":
			if len(opts.args) < 2 {
				fmt.Fprintln(os.Stderr, "kickdesk: -t requires a number (e.g. kickdesk -t 2)")
				os.Exit(1)
			}
			n, err := strconv.Atoi(opts.args[1])
			if err != nil || n < 1 || n > 9 {
				fmt.Fprintln(os.Stderr, "kickdesk: -t must be a number from 1 to 9")
				os.Exit(1)
			}
			opts.tmuxTop = n
			opts.args = opts.args[2:]
		default:
			fmt.Fprintf(os.Stderr, "kickdesk: unknown flag %q\n", opts.args[0])
			os.Exit(1)
		}
	}
	return opts
}

func loadConfig() (*config.Config, string) {
	path, err := config.DefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
		os.Exit(1)
	}
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
		fmt.Fprintf(os.Stderr, "copy examples/config.json to %s\n", path)
		os.Exit(1)
	}
	return cfg, path
}

func runMenu(opts cliOpts) {
	cfg, _ := loadConfig()
	if opts.tmuxTop > 0 && !tmux.InChild() {
		if err := tmux.Bootstrap(cfg.OrderedAppNames(), opts.tmuxTop); err != nil {
			if err.Error() == "cancelled" {
				os.Exit(0)
			}
			fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if err := menu.Run(cfg, version); err != nil {
		fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
		os.Exit(1)
	}
}

func runStatus() {
	cfg, _ := loadConfig()
	rows, err := status.Collect(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
		os.Exit(1)
	}
	status.Print(rows)
}

func runCommand(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: kickdesk run <app> <key>\n")
		os.Exit(1)
	}
	cfg, _ := loadConfig()
	appName := args[0]
	key := args[1]
	if err := run.Execute(cfg, appName, key); err != nil {
		fmt.Fprintf(os.Stderr, "kickdesk: %v\n", err)
		os.Exit(1)
	}
}

func runConfigValidate() {
	cfg, path := loadConfig()
	errs := cfg.Validate()
	if len(errs) == 0 {
		fmt.Printf("ok: %s (%d apps)\n", path, len(cfg.Apps))
		return
	}
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "error: %v\n", e)
	}
	os.Exit(1)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: kickdesk [-t N] [command]\n")
	fmt.Fprintf(os.Stderr, "options:\n")
	fmt.Fprintf(os.Stderr, "  -t N             tmux dashboard: N server panes on top, menu on bottom\n")
	fmt.Fprintf(os.Stderr, "commands:\n")
	fmt.Fprintf(os.Stderr, "  (none)           interactive menu\n")
	fmt.Fprintf(os.Stderr, "  menu             interactive menu\n")
	fmt.Fprintf(os.Stderr, "  status           port and git status table\n")
	fmt.Fprintf(os.Stderr, "  run <app> <key>  run one configured command\n")
	fmt.Fprintf(os.Stderr, "  config validate  check config paths and ports\n")
}
