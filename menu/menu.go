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
	"github.com/Kick-Asset-Management/kickdesk/tmux"
)

const maxCmdDisplay = 56

var errQuit = errors.New("quit")

type menuState struct {
	selected  string
	completed int
}

// Run is the interactive command panel loop.
func Run(cfg *config.Config, version string) error {
	tmux.TryAdoptDashboard()
	reader := bufio.NewReader(os.Stdin)
	state := menuState{}
	for {
		if err := showMenu(cfg, version, reader, &state); err != nil {
			if errors.Is(err, errQuit) {
				return nil
			}
			return err
		}
	}
}

func showMenu(cfg *config.Config, version string, reader *bufio.Reader, state *menuState) error {
	clearScreen()
	fmt.Printf("kickdesk — development command center (%s)\n\n", version)

	apps, err := status.Collect(cfg)
	if err != nil {
		return err
	}
	status.PrintHub(apps, state.selected)

	appNames := cfg.OrderedAppNames()
	var keys []string
	var procedure string
	var running bool

	if state.selected != "" {
		if err := cfg.AppErrors[state.selected]; err != nil {
			fmt.Printf("\n%s — not available\n", state.selected)
			fmt.Println(strings.Repeat("─", 56))
			fmt.Printf("  %v\n", err)
			fmt.Println(strings.Repeat("─", 56))
			fmt.Println("  Fix: publish manifest to ~/.config/<app-id>/manifest.json (see docs/MANIFEST.md)")
			fmt.Println(strings.Repeat("─", 56))
		} else {
			running = status.AppRunning(cfg, state.selected)
			if running {
				procedure = "Shutdown"
			} else {
				procedure = "Startup"
			}
			migrateSt := migrateStatusForApp(apps, state.selected)
			keys, err = EffectiveWorkflowKeys(cfg, state.selected, running, migrateSt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				if werr := waitAnyKeyOrQuit(reader); errors.Is(werr, errQuit) {
					return errQuit
				}
				state.selected = ""
				state.completed = 0
				return nil
			}
			if state.completed > len(keys) {
				state.completed = len(keys)
			}
			printWorkflow(cfg, state.selected, procedure, keys, state.completed)
		}
	}

	printFooter(state.selected != "", len(appNames), len(keys), tmuxTopHint(cfg))

	key, err := readKey(reader)
	if err != nil {
		if errors.Is(err, errQuit) {
			return errQuit
		}
		return err
	}
	fmt.Println()

	switch {
	case key == 'q' || key == 'Q':
		return errQuit
	case key == 'r' || key == 'R':
		return nil
	}

	if state.selected == "" {
		return handleHubKey(cfg, reader, state, appNames, key)
	}
	return handleAppKey(cfg, reader, state, key, keys)
}

func printWorkflow(cfg *config.Config, appName, procedure string, keys []string, completed int) {
	app := cfg.Apps[appName]
	fmt.Printf("\n%s — %s procedure\n", appName, procedure)
	fmt.Println(strings.Repeat("─", 56))
	for i, key := range keys {
		shell := app.Commands[key]
		mark := " "
		if i < completed {
			mark = "✓"
		}
		hint := ""
		if run.IsBlocking(key, shell) {
			if _, ok := run.TmuxChildMode(); ok {
				hint = "  (tmux pane)"
			} else {
				hint = "  (new terminal)"
			}
		}
		fmt.Printf("  %s %-2d  %-10s  %s%s\n", mark, i+1, key, truncate(shell, maxCmdDisplay), hint)
	}
	fmt.Println(strings.Repeat("─", 56))
}

func tmuxTopHint(cfg *config.Config) string {
	topN, ok := run.TmuxChildMode()
	if !ok {
		return ""
	}
	names := cfg.OrderedAppNames()
	if topN > len(names) {
		topN = len(names)
	}
	return fmt.Sprintf("tmux top: %s", strings.Join(names[:topN], ", "))
}

func printFooter(appSelected bool, appCount, stepCount int, tmuxHint string) {
	if tmuxHint != "" {
		fmt.Println(tmuxHint)
	}
	if appSelected {
		fmt.Println("Space next (✓) · Enter all · 1-N run (✓ if next in order) · b back · c catalog · r refresh · q quit")
	} else {
		fmt.Printf("1-%d select app · r refresh · q quit\n", appCount)
	}
}

func handleHubKey(cfg *config.Config, reader *bufio.Reader, state *menuState, appNames []string, key byte) error {
	if key >= '1' && key <= '9' {
		n := int(key - '0')
		if n >= 1 && n <= len(appNames) {
			state.selected = appNames[n-1]
			state.completed = 0
		}
		return nil
	}
	fmt.Printf("Unknown key. Use 1-%d to select an app, r, or q.\n", len(appNames))
	if werr := waitAnyKeyOrQuit(reader); errors.Is(werr, errQuit) {
		return errQuit
	}
	return nil
}

func handleAppKey(cfg *config.Config, reader *bufio.Reader, state *menuState, key byte, keys []string) error {
	appName := state.selected

	switch key {
	case 27, 'b', 'B':
		state.selected = ""
		state.completed = 0
		return nil
	case 'c', 'C':
		if showCatalog(cfg, appName, reader) {
			return errQuit
		}
		return nil
	case ' ':
		if state.completed < len(keys) {
			if err := runStep(cfg, appName, keys[state.completed]); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			} else {
				state.completed++
			}
			if state.completed >= len(keys) {
				return afterProcedureRun(reader, state)
			}
			if werr := waitAnyKeyOrQuit(reader); errors.Is(werr, errQuit) {
				return errQuit
			}
		}
	case '\r', '\n':
		remaining := keys[state.completed:]
		if len(remaining) == 0 {
			return nil
		}
		if err := run.ExecuteSequence(cfg, appName, remaining); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		return afterProcedureRun(reader, state)
	default:
		if key >= '1' && key <= '9' {
			idx := int(key - '1')
			if idx < len(keys) {
				if err := runStep(cfg, appName, keys[idx]); err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
				} else if idx == state.completed {
					state.completed++
				}
				if state.completed >= len(keys) {
					return afterProcedureRun(reader, state)
				}
				if werr := waitAnyKeyOrQuit(reader); errors.Is(werr, errQuit) {
					return errQuit
				}
				return nil
			}
		}
		fmt.Println("Unknown key. Space, Enter, 1-N, c, b, r, or q.")
		if werr := waitAnyKeyOrQuit(reader); errors.Is(werr, errQuit) {
			return errQuit
		}
	}
	return nil
}

func afterProcedureRun(reader *bufio.Reader, state *menuState) error {
	state.completed = 0
	if werr := waitReturnToMenu(reader); errors.Is(werr, errQuit) {
		return errQuit
	}
	return nil
}

func runStep(cfg *config.Config, appName, key string) error {
	app := cfg.Apps[appName]
	shell := app.Commands[key]
	fmt.Printf("\n── %s / %s ──\n", appName, key)
	fmt.Printf("$ %s\n\n", shell)
	if run.IsBlocking(key, shell) {
		if _, ok := run.TmuxChildMode(); ok {
			fmt.Println("(runs in tmux server pane above)")
		} else {
			fmt.Println("(opens in a new terminal when supported)")
		}
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
		return errors.Is(waitAnyKeyOrQuit(reader), errQuit)
	}
	for _, e := range entries {
		fmt.Printf("  %-12s  %s\n", e.Key, truncate(e.Shell, maxCmdDisplay))
	}
	fmt.Println("\n  kickdesk run", appName, "<key>")
	return errors.Is(waitAnyKeyOrQuit(reader), errQuit)
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
