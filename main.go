package main

import (
	"fmt"
	"os"

	"github.com/Kick-Asset-Management/kickdesk/config"
	"github.com/Kick-Asset-Management/kickdesk/menu"
	"github.com/Kick-Asset-Management/kickdesk/run"
	"github.com/Kick-Asset-Management/kickdesk/status"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		runMenu()
		return
	}

	switch os.Args[1] {
	case "menu":
		runMenu()
	case "status":
		runStatus()
	case "run":
		runCommand()
	case "config":
		if len(os.Args) >= 3 && os.Args[2] == "validate" {
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

func runMenu() {
	cfg, _ := loadConfig()
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

func runCommand() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "usage: kickdesk run <app> <key>\n")
		os.Exit(1)
	}
	cfg, _ := loadConfig()
	appName := os.Args[2]
	key := os.Args[3]
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
	fmt.Fprintf(os.Stderr, "usage: kickdesk [command]\n")
	fmt.Fprintf(os.Stderr, "commands:\n")
	fmt.Fprintf(os.Stderr, "  (none)           interactive menu\n")
	fmt.Fprintf(os.Stderr, "  menu             interactive menu\n")
	fmt.Fprintf(os.Stderr, "  status           port and git status table\n")
	fmt.Fprintf(os.Stderr, "  run <app> <key>  run one configured command\n")
	fmt.Fprintf(os.Stderr, "  config validate  check config paths and ports\n")
}
