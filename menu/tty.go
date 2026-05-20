package menu

import (
	"io"
	"os"

	"golang.org/x/term"
)

// inputTTY returns the file to read interactive input from.
// Prefer stdin when it is a terminal; else /dev/tty; else stdin (pipe).
func inputTTY() (*os.File, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return os.Stdin, nil
	}
	if tty, err := os.Open("/dev/tty"); err == nil {
		if term.IsTerminal(int(tty.Fd())) {
			return tty, nil
		}
		_ = tty.Close()
	}
	return os.Stdin, nil
}

// readKeyFromFD reads one byte from the given terminal fd (must match MakeRaw target).
func readKeyFromFD(fd int, f io.Reader) (byte, error) {
	buf := make([]byte, 1)
	n, err := f.Read(buf)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, io.EOF
	}
	return buf[0], nil
}
