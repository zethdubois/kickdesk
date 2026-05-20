package menu

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Kick-Asset-Management/kickdesk/config"
	"golang.org/x/term"
)

// ProfilePickerResult is the outcome of the profile selection screen.
type ProfilePickerResult struct {
	ProfileName string
	Quit        bool
}

// RunProfilePicker shows hotkey profile selection (0–9, no Enter).
func RunProfilePicker() (ProfilePickerResult, error) {
	profiles, err := config.ListProfiles()
	if err != nil {
		return ProfilePickerResult{}, err
	}

	tty, err := openControllingTTY()
	if err != nil {
		return ProfilePickerResult{}, err
	}
	defer func() { _ = tty.Close() }()

	fd := int(tty.Fd())
	if !term.IsTerminal(fd) {
		return runProfilePickerLine(profiles)
	}

	old, err := term.MakeRaw(fd)
	if err != nil {
		return runProfilePickerLine(profiles)
	}
	defer func() { _ = term.Restore(fd, old) }()

	for {
		drawProfilePickerTo(tty, profiles)
		key, err := readKeyRaw(fd, tty)
		if err != nil {
			if errors.Is(err, errQuit) || errors.Is(err, io.EOF) {
				return ProfilePickerResult{Quit: true}, nil
			}
			return ProfilePickerResult{}, err
		}

		switch {
		case key == 'q' || key == 'Q':
			return ProfilePickerResult{Quit: true}, nil
		case key == 'r' || key == 'R':
			profiles, err = config.ListProfiles()
			if err != nil {
				return ProfilePickerResult{}, err
			}
		case key == '0':
			showConfigEditorOn(tty, fd)
		case key >= '1' && key <= '9':
			n := int(key - '0')
			if n >= 1 && n <= len(profiles) {
				return ProfilePickerResult{ProfileName: profiles[n-1]}, nil
			}
			flashOn(tty, "No profile for that number")
		default:
			flashOn(tty, "Use 0–9, r, or q")
		}
	}
}

func openControllingTTY() (*os.File, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return os.Stdin, nil
	}
	return os.OpenFile("/dev/tty", os.O_RDWR, 0)
}

func drawProfilePickerTo(w *os.File, profiles []string) {
	fmt.Fprint(w, "\033[H\033[2J")
	fmt.Fprintln(w, "kickdesk — choose config")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  [0]  Config editor (coming soon)")
	for i, name := range profiles {
		fmt.Fprintf(w, "  [%d]  %s\n", i+1, name)
	}
	fmt.Fprintln(w)
	if len(profiles) == 0 {
		fmt.Fprintln(w, "  No profiles yet — add <name>.json next to config.json")
		fmt.Fprintln(w, "  (config.json is the registry, not selectable here)")
	}
	fmt.Fprintln(w, "  0–9 select · r refresh · q quit")
	_ = w.Sync()
}

func showConfigEditorOn(w *os.File, fd int) {
	fmt.Fprint(w, "\033[H\033[2J")
	fmt.Fprintln(w, "Config editor — coming soon.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, config.SetupConfigHint())
	fmt.Fprintln(w)
	fmt.Fprint(w, "Press any key...")
	_ = w.Sync()
	_, _ = readKeyRaw(fd, w)
}

func flashOn(w *os.File, msg string) {
	fmt.Fprintf(w, "\r%s                    \r", msg)
	_ = w.Sync()
}

func runProfilePickerLine(profiles []string) (ProfilePickerResult, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		drawProfilePickerTo(os.Stderr, profiles)
		fmt.Fprint(os.Stderr, "Choice [0-9 q]: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return ProfilePickerResult{}, err
		}
		line = strings.TrimSpace(line)
		if line == "" || strings.EqualFold(line, "q") {
			return ProfilePickerResult{Quit: true}, nil
		}
		if strings.EqualFold(line, "r") {
			profiles, err = config.ListProfiles()
			if err != nil {
				return ProfilePickerResult{}, err
			}
			continue
		}
		n, err := strconv.Atoi(line)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Enter 0–9, r, or q")
			continue
		}
		if n == 0 {
			fmt.Fprintln(os.Stderr, config.SetupConfigHint())
			continue
		}
		if n >= 1 && n <= len(profiles) {
			return ProfilePickerResult{ProfileName: profiles[n-1]}, nil
		}
		fmt.Fprintln(os.Stderr, "Invalid choice")
	}
}
