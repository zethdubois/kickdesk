package run

import "testing"

func TestIsBlocking(t *testing.T) {
	cases := []struct {
		key, shell string
		want       bool
	}{
		{"up", "pnpm serve-publish", true},
		{"wiki", "pnpm kam:wiki", true},
		{"standalone", "pnpm standalone", true},
		{"build", "pnpm build", false},
		{"down", "true", false},
		{"republish", "pnpm build && pnpm publish:artifacts", false},
		{"other", "node scripts/serve-kam-wiki.js", true}, // serve- heuristic
		{"other", "pnpm kam:wiki", true},
	}
	for _, tc := range cases {
		got := IsBlocking(tc.key, tc.shell)
		if got != tc.want {
			t.Errorf("IsBlocking(%q, %q) = %v, want %v", tc.key, tc.shell, got, tc.want)
		}
	}
}
