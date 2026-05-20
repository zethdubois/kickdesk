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

func drainReader(fd int, r io.Reader, wait time.Duration) {
	if f, ok := r.(*os.File); ok {
		_ = f.SetReadDeadline(time.Now().Add(wait))
		buf := make([]byte, 16)
		for {
			n, err := r.Read(buf)
			if err != nil || n == 0 {
				break
			}
		}
		_ = f.SetReadDeadline(time.Time{})
		return
	}
	_ = fd
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

// waitReturnToMenu pauses after a full procedure before refreshing the menu.
func waitReturnToMenu(reader *bufio.Reader) error {
	return waitFor(reader, "Press any key to return to menu (q quit)... ", nil)
}
