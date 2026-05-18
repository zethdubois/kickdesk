package tmux

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmStartTmux asks whether to start a tmux dashboard when not inside tmux.
func ConfirmStartTmux() bool {
	fmt.Fprint(os.Stderr, "Dashboard mode (-t) uses tmux: server panes on top, kickdesk menu on bottom.\n")
	fmt.Fprint(os.Stderr, "Start tmux session \"kickdesk\" now? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}
