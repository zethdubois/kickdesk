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
	if root, err := tmux.RepoRoot(); err == nil && root != "" {
		_ = os.Chdir(root)
		defer func() { _ = tmux.QueueMenuPaneCd(root) }()
	}
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
	var proc Procedure

	if state.selected != "" {
		if err := cfg.AppErrors[state.selected]; err != nil {
			fmt.Printf("\n%s — not available\n", state.selected)
			fmt.Println(strings.Repeat("─", 56))
			fmt.Printf("  %v\n", err)
			fmt.Println(strings.Repeat("─", 56))
			fmt.Println("  Fix: publish manifest to ~/.config/<app-id>/manifest.json (see docs/MANIFEST.md)")
			fmt.Println(strings.Repeat("─", 56))
		} else {
			running := status.AppRunning(cfg, state.selected)
			migrateSt := migrateStatusForApp(apps, state.selected)
			proc, err = EffectiveProcedure(cfg, state.selected, running, migrateSt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				if werr := waitAnyKeyOrQuit(reader); errors.Is(werr, errQuit) {
					return errQuit
				}
				state.selected = ""
				state.completed = 0
				return nil
			}
			if state.completed > len(proc.Keys) {
				state.completed = len(proc.Keys)
			}
			printWorkflow(cfg, state.selected, proc, state.completed)
		}
	}

	printFooter(state.selected != "", len(appNames), proc.HasRepublish(), tmuxTopHint(cfg))

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
	return handleAppKey(cfg, reader, state, key, proc)
}

func printWorkflow(cfg *config.Config, appName string, proc Procedure, completed int) {
	app := cfg.Apps[appName]
	fmt.Printf("\n%s — %s procedure\n", appName, proc.Title)
	fmt.Println(strings.Repeat("─", 56))
	for i, key := range proc.Keys {
		if proc.HasRepublish() {
			if i == 0 {
				fmt.Println("  Shutdown")
			}
			if i == proc.RepublishFrom {
				fmt.Println("  Republish")
			}
		}
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

func printFooter(appSelected bool, appCount int, hasRepublish bool, tmuxHint string) {
	if tmuxHint != "" {
		fmt.Println(tmuxHint)
	}
	if appSelected {
		republishHint := ""
		if hasRepublish {
			republishHint = " · p republish all"
		}
		fmt.Printf("Space next (✓) · Enter all · 1-N run (✓ if next in order)%s · b back · c catalog (key run) · r refresh · q quit\n", republishHint)
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

func handleAppKey(cfg *config.Config, reader *bufio.Reader, state *menuState, key byte, proc Procedure) error {
	appName := state.selected
	keys := proc.Keys

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
	case 'p', 'P':
		if !proc.HasRepublish() {
			fmt.Println("Republish is not configured for this app.")
			if werr := waitAnyKeyOrQuit(reader); errors.Is(werr, errQuit) {
				return errQuit
			}
			return nil
		}
		republish := proc.RepublishKeys()
		if err := run.ExecuteSequence(cfg, appName, republish); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		// Republish never advances the Shutdown ✓ counter: the app stays
		// running and the operator may republish repeatedly.
		if werr := waitReturnToMenu(reader); errors.Is(werr, errQuit) {
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
				} else if idx == state.completed && !proc.IsRepublishIndex(idx) {
					// Republish steps are non-destructive and may be re-run
					// at any time; do not advance the shutdown ✓ cursor.
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
		hint := "Unknown key. Space, Enter, 1-N, c, b, r, or q."
		if proc.HasRepublish() {
			hint = "Unknown key. Space, Enter, 1-N, p, c, b, r, or q."
		}
		fmt.Println(hint)
		if werr := waitAnyKeyOrQuit(reader); errors.Is(werr, errQuit) {
			return errQuit
		}
	}
	return nil
}

func afterProcedureRun(reader *bufio.Reader, state *menuState) error {
	state.selected = ""
	state.completed = 0
	if werr := waitReturnToHub(reader); errors.Is(werr, errQuit) {
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

// catalogNavKeys are reserved in the catalog screen (not used as run hotkeys).
var catalogNavKeys = map[byte]bool{
	'b': true, 'B': true,
	'q': true, 'Q': true,
	'r': true, 'R': true,
	27: true, // Esc back
}

// catalogRunKeys is the ordered pool of single-key bindings for command rows.
// Letters skip b/q/r so navigation still works with 10+ commands.
func catalogRunKeys() []byte {
	keys := []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9'}
	for c := byte('a'); c <= 'z'; c++ {
		if catalogNavKeys[c] {
			continue
		}
		keys = append(keys, c)
	}
	return keys
}

// assignCatalogKeys maps each command entry to a single hotkey.
// Prefers the first letter of the command name when free (wiki → w); otherwise
// takes the next free pool key so sidecars remain reachable past 9 entries.
func assignCatalogKeys(entries []config.CommandEntry) []byte {
	pool := catalogRunKeys()
	used := make(map[byte]bool)
	out := make([]byte, len(entries))
	// Pass 1: prefer mnemonic first letter.
	for i, e := range entries {
		if e.Key == "" {
			continue
		}
		pref := e.Key[0]
		if pref >= 'A' && pref <= 'Z' {
			pref += 'a' - 'A'
		}
		if pref >= 'a' && pref <= 'z' && !catalogNavKeys[pref] && !used[pref] {
			out[i] = pref
			used[pref] = true
		}
	}
	// Pass 2: fill remaining from pool order.
	pi := 0
	for i := range entries {
		if out[i] != 0 {
			continue
		}
		for pi < len(pool) && used[pool[pi]] {
			pi++
		}
		if pi >= len(pool) {
			break // no more single-key slots
		}
		out[i] = pool[pi]
		used[pool[pi]] = true
		pi++
	}
	return out
}

// showCatalog returns true if the user chose to quit kickdesk.
// While open, each listed command has a single-key hotkey (e.g. w = wiki).
func showCatalog(cfg *config.Config, appName string, reader *bufio.Reader) bool {
	for {
		clearScreen()
		fmt.Printf("Commands — %s\n\n", appName)
		entries, err := cfg.CommandsList(appName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return errors.Is(waitAnyKeyOrQuit(reader), errQuit)
		}
		hotkeys := assignCatalogKeys(entries)
		byKey := make(map[byte]string, len(entries))
		for i, e := range entries {
			label := " "
			if i < len(hotkeys) && hotkeys[i] != 0 {
				label = string(hotkeys[i])
				byKey[hotkeys[i]] = e.Key
			}
			hint := ""
			if run.IsBlocking(e.Key, e.Shell) {
				if _, ok := run.TmuxChildMode(); ok {
					hint = "  (tmux pane)"
				} else {
					hint = "  (new terminal)"
				}
			}
			fmt.Printf("  %s  %-14s  %s%s\n", label, e.Key, truncate(e.Shell, maxCmdDisplay), hint)
		}
		fmt.Println(strings.Repeat("─", 56))
		if len(byKey) > 0 {
			fmt.Println("key run · b back · r refresh · q quit")
		} else {
			fmt.Println("b back · r refresh · q quit")
		}
		fmt.Printf("kickdesk run %s <key>\n", appName)

		key, err := readKey(reader)
		if err != nil {
			return errors.Is(err, errQuit)
		}
		fmt.Println()
		switch {
		case key == 'q' || key == 'Q':
			return true
		case key == 27 || key == 'b' || key == 'B':
			return false
		case key == 'r' || key == 'R':
			continue
		default:
			// Normalize A–Z to a–z for mnemonic matches.
			lookup := key
			if lookup >= 'A' && lookup <= 'Z' {
				lookup += 'a' - 'A'
			}
			if cmdKey, ok := byKey[lookup]; ok {
				if err := runStep(cfg, appName, cmdKey); err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
				}
				if werr := waitReturnToMenu(reader); errors.Is(werr, errQuit) {
					return true
				}
				continue
			}
			hint := "Unknown key. b, r, or q."
			if len(byKey) > 0 {
				hint = "Unknown key. Use a listed key, b, r, or q."
			}
			fmt.Println(hint)
			if werr := waitAnyKeyOrQuit(reader); errors.Is(werr, errQuit) {
				return true
			}
		}
	}
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
