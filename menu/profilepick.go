package menu

import (
	"fmt"
	"os"

	"github.com/Kick-Asset-Management/kickdesk/config"
)

// ResolveProfileChoice maps picker index 0–9 to a profile name (0 = setup placeholder, returns "").
func ResolveProfileChoice(n int) (string, error) {
	if n == 0 {
		fmt.Fprintln(os.Stderr, "Config editor — coming soon.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, config.SetupConfigHint())
		return "", nil
	}
	profiles, err := config.ListProfiles()
	if err != nil {
		return "", err
	}
	if n < 1 || n > len(profiles) {
		return "", fmt.Errorf("profile choice %d invalid (have %d profile(s), use 1–%d)", n, len(profiles), len(profiles))
	}
	return profiles[n-1], nil
}
