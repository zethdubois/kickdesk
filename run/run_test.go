package run

import "testing"

func TestIsBlocking(t *testing.T) {
	cases := []struct {
		key, shell string
		want       bool
	}{
		{"up", "pnpm serve-publish", true},
		{"wiki", "pnpm kam:wiki", true},
		{"discovery", "pnpm kam:discovery", true},
		{"progress", "pnpm kam:progress", true},
		{"build", "pnpm build", false},
		{"down", "true", false},
		{"republish", "pnpm build && pnpm publish:artifacts", false},
		{"other", "node scripts/serve-kam-wiki.js", true}, // serve- heuristic
		{"other", "pnpm kam:wiki", true},
		{"other", "pnpm kam:progress", true},
	}
	for _, tc := range cases {
		got := IsBlocking(tc.key, tc.shell)
		if got != tc.want {
			t.Errorf("IsBlocking(%q, %q) = %v, want %v", tc.key, tc.shell, got, tc.want)
		}
	}
}

func TestUsesAppDashboardPane(t *testing.T) {
	if !UsesAppDashboardPane("up") {
		t.Fatal("up should use dashboard pane")
	}
	if !UsesAppDashboardPane("") {
		t.Fatal("empty key should use dashboard pane")
	}
	for _, key := range []string{"wiki", "discovery", "progress", "standalone"} {
		if UsesAppDashboardPane(key) {
			t.Fatalf("%s must not use dashboard pane", key)
		}
	}
}
