package menu

import (
	"strings"

	"github.com/Kick-Asset-Management/kickdesk/config"
	"github.com/Kick-Asset-Management/kickdesk/status"
)

// isPortOpen is swapped in tests.
var isPortOpen = status.IsPortOpen

// inspectorReservedKeys must not be used as Services hotkeys.
var inspectorReservedKeys = map[byte]bool{
	'q': true, 'Q': true,
	'r': true, 'R': true,
	'b': true, 'B': true,
	'c': true, 'C': true,
	'p': true, 'P': true, // republish
	27: true,
}

// SideService is one extra port (wiki, discovery, gateway, …) with a start or
// stop command the inspector can run without walking the main procedure.
type SideService struct {
	Role   string
	Port   int
	Up     bool
	Key    string // command key to run
	Hotkey byte
}

// ListSideServices returns extra-port actions for the app inspector.
//
// The main procedure is still keyed off the primary port (startup vs shutdown).
// Extra ports that are down get their start command (role name, e.g. "wiki").
// Extra ports that are up get stop-<role> when that stop is not already a
// numbered shutdown step (i.e. when the primary is down and the inspector
// is showing Startup).
func ListSideServices(cfg *config.Config, appName string, primaryRunning bool) []SideService {
	app, ok := cfg.Apps[appName]
	if !ok || app.Commands == nil {
		return nil
	}
	primary := app.MainPort()
	inStart := keySet(app.Workflows.Start)
	inStop := keySet(app.Workflows.Stop)

	var sides []SideService
	for _, b := range app.Ports {
		if b.Port <= 0 || b.Port == primary {
			continue
		}
		role := strings.TrimSpace(b.Role)
		if role == "" || role == "db" {
			continue
		}
		up := isPortOpen(b.Port)
		startKey := role
		stopKey := "stop-" + role
		var cmd string
		if up {
			if app.Commands[stopKey] == "" {
				continue
			}
			if primaryRunning && inStop[stopKey] {
				continue
			}
			cmd = stopKey
		} else {
			if app.Commands[startKey] == "" {
				continue
			}
			if !primaryRunning && inStart[startKey] {
				continue
			}
			cmd = startKey
		}
		sides = append(sides, SideService{
			Role: role,
			Port: b.Port,
			Up:   up,
			Key:  cmd,
		})
	}
	assignSideHotkeys(sides)
	return sides
}

func keySet(keys []string) map[string]bool {
	out := make(map[string]bool, len(keys))
	for _, k := range keys {
		out[k] = true
	}
	return out
}

func assignSideHotkeys(sides []SideService) {
	used := make(map[byte]bool)
	for k, v := range inspectorReservedKeys {
		if v {
			used[k] = true
		}
	}
	for i := range sides {
		pref := sideMnemonic(sides[i].Role)
		if pref != 0 && !used[pref] {
			sides[i].Hotkey = pref
			used[pref] = true
		}
	}
	for i := range sides {
		if sides[i].Hotkey != 0 {
			continue
		}
		for c := byte('a'); c <= 'z'; c++ {
			if used[c] {
				continue
			}
			sides[i].Hotkey = c
			used[c] = true
			break
		}
	}
}

func sideMnemonic(role string) byte {
	if role == "" {
		return 0
	}
	c := role[0]
	if c >= 'A' && c <= 'Z' {
		c += 'a' - 'A'
	}
	if c < 'a' || c > 'z' {
		return 0
	}
	if inspectorReservedKeys[c] {
		return 0
	}
	return c
}

func sideByHotkey(sides []SideService, key byte) (SideService, bool) {
	if key >= 'A' && key <= 'Z' {
		key += 'a' - 'A'
	}
	for _, s := range sides {
		if s.Hotkey != 0 && s.Hotkey == key {
			return s, true
		}
	}
	return SideService{}, false
}

func sideFooterHint(sides []SideService) string {
	if len(sides) == 0 {
		return ""
	}
	parts := make([]string, 0, len(sides))
	for _, s := range sides {
		if s.Hotkey == 0 {
			continue
		}
		parts = append(parts, string(s.Hotkey)+" "+s.Key)
	}
	return strings.Join(parts, " · ")
}
