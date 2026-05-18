package menu

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// readKey reads a single key from the terminal (no Enter required).
// Ctrl+C and Ctrl+D quit (restores terminal). Falls back to line input when not a TTY.
func readKey(reader *bufio.Reader) (byte, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		line, err := reader.ReadString('\n')
		if err != nil {
			return 0, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			return '\r', nil
		}
		if strings.EqualFold(line, "q") {
			return 0, errQuit
		}
		return line[0], nil
	}

	old, err := term.MakeRaw(fd)
	if err != nil {
		return 0, err
	}
	defer func() { _ = term.Restore(fd, old) }()

	buf := make([]byte, 1)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, nil
	}
	switch buf[0] {
	case 3, 4: // Ctrl+C, Ctrl+D
		return 0, errQuit
	case 27: // Esc
		return 27, nil
	}
	return buf[0], nil
}

// waitFor pauses until accept returns true for a key (nil accept = any key except q).
func waitFor(reader *bufio.Reader, prompt string, accept func(byte) bool) error {
	fmt.Print(prompt)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		for {
			k, err := readKey(reader)
			if errors.Is(err, errQuit) {
				return errQuit
			}
			if err != nil {
				return err
			}
			if k == 'q' || k == 'Q' {
				return errQuit
			}
			if accept == nil || accept(k) {
				return nil
			}
		}
	}
	_, err := reader.ReadString('\n')
	return err
}

// waitAnyKeyOrQuit pauses so command output can be read; any key continues, q quits.
func waitAnyKeyOrQuit(reader *bufio.Reader) error {
	return waitFor(reader, "\nPress any key to continue (q quit)...", nil)
}

// waitReturnToMenu pauses after a full procedure before refreshing the menu.
func waitReturnToMenu(reader *bufio.Reader) error {
	return waitFor(reader, "\nPress any key to return to menu (q quit)...", nil)
}
