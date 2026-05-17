package menu

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Kick-Asset-Management/kickdesk/config"
	"github.com/Kick-Asset-Management/kickdesk/run"
	"github.com/Kick-Asset-Management/kickdesk/status"
)

const maxCmdDisplay = 56

var errQuit = errors.New("quit")

// Run is the interactive command panel loop.
func Run(cfg *config.Config, version string) error {
	reader := bufio.NewReader(os.Stdin)
	for {
		if err := showHub(cfg, version, reader); err != nil {
			if errors.Is(err, errQuit) {
				return nil
			}
			return err
		}
	}
}

func showHub(cfg *config.Config, version string, reader *bufio.Reader) error {
	clearScreen()
	fmt.Printf("kickdesk — development command center (%s)\n\n", version)

	apps, err := status.Collect(cfg)
	if err != nil {
		return err
	}
	status.PrintHub(apps)

	appNames := cfg.OrderedAppNames()
	fmt.Printf("\n1-%d  app workflow    r  refresh    q  quit\n", len(appNames))
	fmt.Print("\n> ")

	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	line = strings.TrimSpace(line)
	switch strings.ToLower(line) {
	case "q", "quit", "exit":
		return errQuit
	case "r", "refresh":
		return nil
	}

	var n int
	if _, err := fmt.Sscanf(line, "%d", &n); err != nil || n < 1 || n > len(appNames) {
		fmt.Println("Invalid selection.")
		if err := waitKey(reader); errors.Is(err, errQuit) {
			return errQuit
		}
		return nil
	}

	return showWorkflow(cfg, appNames[n-1], reader)
}

func showWorkflow(cfg *config.Config, appName string, reader *bufio.Reader) error {
	running := status.AppRunning(cfg, appName)
	keys, err := cfg.WorkflowKeys(appName, running)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if werr := waitKey(reader); errors.Is(werr, errQuit) {
			return errQuit
		}
		return nil
	}

	app := cfg.Apps[appName]
	completed := 0
	procedure := "Startup"
	if running {
		procedure = "Shutdown"
	}

	for {
		clearScreen()
		fmt.Printf("kickdesk — %s\n\n", appName)
		_ = status.PrintAppSummary(cfg, appName)

		fmt.Printf("%s procedure\n", procedure)
		fmt.Println(strings.Repeat("─", 56))
		for i, key := range keys {
			shell := app.Commands[key]
			mark := " "
			if i < completed {
				mark = "✓"
			}
			hint := ""
			if run.IsBlocking(key, shell) {
				hint = "  (blocks — Ctrl+C stops command, q quits kickdesk)"
			}
			fmt.Printf("  %s %-2d  %-10s  %s%s\n", mark, i+1, key, truncate(shell, maxCmdDisplay), hint)
		}

		fmt.Println(strings.Repeat("─", 56))
		fmt.Println("  Space   next step")
		fmt.Println("  Enter   all remaining steps")
		fmt.Println("  c       command catalog")
		fmt.Println("  b       back")
		fmt.Println("  q       quit kickdesk")
		fmt.Print("\n> ")

		key, err := readKey(reader)
		if err != nil {
			if errors.Is(err, errQuit) {
				return errQuit
			}
			return err
		}
		fmt.Println()

		switch key {
		case 'q', 'Q':
			return errQuit
		case 27, 'b', 'B': // Esc or b
			return nil
		case 'c', 'C':
			if showCatalog(cfg, appName, reader) {
				return errQuit
			}
			continue
		case ' ':
			if completed < len(keys) {
				if err := runStep(cfg, appName, keys[completed]); err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					if werr := waitKey(reader); errors.Is(werr, errQuit) {
						return errQuit
					}
				} else {
					completed++
				}
			}
			if completed >= len(keys) {
				newRunning := status.AppRunning(cfg, appName)
				if newRunning != running {
					if werr := waitKey(reader); errors.Is(werr, errQuit) {
						return errQuit
					}
					return nil
				}
				completed = 0
				running = newRunning
				keys, _ = cfg.WorkflowKeys(appName, running)
			}
		case '\r', '\n':
			remaining := keys[completed:]
			if len(remaining) == 0 {
				continue
			}
			if err := run.ExecuteSequence(cfg, appName, remaining); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				if werr := waitKey(reader); errors.Is(werr, errQuit) {
					return errQuit
				}
			}
			newRunning := status.AppRunning(cfg, appName)
			if newRunning != running {
				if werr := waitKey(reader); errors.Is(werr, errQuit) {
					return errQuit
				}
				return nil
			}
			running = newRunning
			keys, err = cfg.WorkflowKeys(appName, running)
			if err != nil {
				return nil
			}
			completed = 0
		default:
			fmt.Println("Unknown key. Use Space, Enter, c, b, or q.")
			if werr := waitKey(reader); errors.Is(werr, errQuit) {
				return errQuit
			}
		}
	}
}

func runStep(cfg *config.Config, appName, key string) error {
	app := cfg.Apps[appName]
	shell := app.Commands[key]
	fmt.Printf("\n── %s / %s ──\n", appName, key)
	fmt.Printf("$ %s\n\n", shell)
	if run.IsBlocking(key, shell) {
		fmt.Println("(running — Ctrl+C stops this command; you return to the menu)")
	}
	return run.Execute(cfg, appName, key)
}

// showCatalog returns true if the user chose to quit kickdesk.
func showCatalog(cfg *config.Config, appName string, reader *bufio.Reader) bool {
	clearScreen()
	fmt.Printf("Commands — %s\n\n", appName)
	entries, err := cfg.CommandsList(appName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return errors.Is(waitKey(reader), errQuit)
	}
	for _, e := range entries {
		fmt.Printf("  %-12s  %s\n", e.Key, truncate(e.Shell, maxCmdDisplay))
	}
	fmt.Println("\n  kickdesk run", appName, "<key>")
	return errors.Is(waitKey(reader), errQuit)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
