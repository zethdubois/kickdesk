package menu

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// readKey reads one key (TTY: raw, no Enter; otherwise: line with prompt on stderr).
// While waiting, the tty is in raw mode: Ctrl+C is not SIGINT (it is byte 3 = quit
// once the read receives it). A wedged Read therefore looks like a "frozen" menu.
func readKey(reader *bufio.Reader) (byte, error) {
	tty, err := inputTTY()
	if err != nil || !term.IsTerminal(int(tty.Fd())) {
		return readKeyLine(reader, "choice: ")
	}
	fd := int(tty.Fd())
	var key byte
	err = withRawTerminalOn(fd, tty, func(r io.Reader) error {
		var err error
		key, err = readKeyRaw(fd, r)
		return err
	})
	return key, err
}

func readKeyLine(reader *bufio.Reader, prompt string) (byte, error) {
	fmt.Fprint(os.Stderr, prompt)
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

func withRawTerminalOn(fd int, tty *os.File, fn func(io.Reader) error) error {
	if !term.IsTerminal(fd) {
		return fn(tty)
	}
	old, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer func() { _ = term.Restore(fd, old) }()
	return fn(tty)
}

// readKeyRaw reads one key from r; fd is used only to drain escape sequences.
func readKeyRaw(fd int, r io.Reader) (byte, error) {
	key, err := readKeyFromFD(fd, r)
	if err != nil {
		return 0, err
	}
	switch key {
	case 3, 4:
		return 0, errQuit
	case 27:
		drainReader(fd, r, 50*time.Millisecond)
		return 27, nil
	}
	return key, nil
}

// drainReader consumes the rest of an escape/CSI sequence after ESC.
// If SetReadDeadline is unavailable on this tty, it returns immediately —
// a blocking Read here used to wedge the menu forever (keys and Ctrl+C
// appear dead because MakeRaw disables ISIG; Ctrl+C is only byte 3).
func drainReader(fd int, r io.Reader, wait time.Duration) {
	f, ok := r.(*os.File)
	if !ok {
		_ = fd
		return
	}
	if err := f.SetReadDeadline(time.Now().Add(wait)); err != nil {
		return
	}
	defer func() { _ = f.SetReadDeadline(time.Time{}) }()
	buf := make([]byte, 64)
	for {
		n, err := f.Read(buf)
		if err != nil || n == 0 {
			return
		}
	}
}

func flushScreen() {
	_ = os.Stdout.Sync()
	_ = os.Stderr.Sync()
}

// waitFor pauses until accept returns true for a key (nil accept = any key except q).
func waitFor(reader *bufio.Reader, prompt string, accept func(byte) bool) error {
	fmt.Fprint(os.Stderr, prompt)
	flushScreen()
	tty, err := inputTTY()
	if err != nil || !term.IsTerminal(int(tty.Fd())) {
		_, err := reader.ReadString('\n')
		return err
	}
	fd := int(tty.Fd())
	return withRawTerminalOn(fd, tty, func(r io.Reader) error {
		for {
			k, err := readKeyRaw(fd, r)
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
	})
}

// waitAnyKeyOrQuit pauses so command output can be read; any key continues, q quits.
func waitAnyKeyOrQuit(reader *bufio.Reader) error {
	return waitFor(reader, "Press any key (q quit)... ", nil)
}

// waitReturnToMenu pauses after a partial action before refreshing the app submenu.
func waitReturnToMenu(reader *bufio.Reader) error {
	return waitFor(reader, "Press any key to return to menu (q quit)... ", nil)
}

// waitReturnToHub pauses after a full procedure before returning to the app hub.
func waitReturnToHub(reader *bufio.Reader) error {
	return waitFor(reader, "Press any key to return to main menu (q quit)... ", nil)
}
