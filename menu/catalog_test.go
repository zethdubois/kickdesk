package menu

import (
	"strings"
	"testing"

	"github.com/Kick-Asset-Management/kickdesk/config"
)

func TestAssignCatalogKeys_mnemonicWiki(t *testing.T) {
	// Mirrors a busy kickagent-like command map (more than 9 entries).
	entries := []config.CommandEntry{
		{Key: "build"},
		{Key: "discovery"},
		{Key: "down"},
		{Key: "publish-artifacts"},
		{Key: "republish"},
		{Key: "stop-discovery"},
		{Key: "stop-server"},
		{Key: "stop-wiki"},
		{Key: "test"},
		{Key: "up"},
		{Key: "wiki"},
	}
	keys := assignCatalogKeys(entries)
	if len(keys) != len(entries) {
		t.Fatalf("len = %d, want %d", len(keys), len(entries))
	}
	// Every entry gets a key.
	for i, k := range keys {
		if k == 0 {
			t.Fatalf("entry %d (%s) has no hotkey", i, entries[i].Key)
		}
	}
	// wiki → w for discoverability in catalog.
	wikiIdx := -1
	for i, e := range entries {
		if e.Key == "wiki" {
			wikiIdx = i
			break
		}
	}
	if wikiIdx < 0 || keys[wikiIdx] != 'w' {
		t.Fatalf("wiki hotkey = %q, want 'w'", string(keys[wikiIdx]))
	}
	// Unique bindings.
	seen := map[byte]string{}
	for i, k := range keys {
		if prev, ok := seen[k]; ok {
			t.Fatalf("duplicate key %c for %s and %s", k, prev, entries[i].Key)
		}
		seen[k] = entries[i].Key
	}
	// b/q/r stay free for navigation.
	for _, forbidden := range []byte{'b', 'q', 'r'} {
		if name, ok := seen[forbidden]; ok {
			t.Fatalf("nav key %c claimed by %s", forbidden, name)
		}
	}
}

func TestAssignCatalogKeys_avoidsNavLetters(t *testing.T) {
	// First-letter collision with reserved b and q.
	entries := []config.CommandEntry{
		{Key: "build"},
		{Key: "quick"},
		{Key: "run-x"},
	}
	keys := assignCatalogKeys(entries)
	for i, k := range keys {
		if catalogNavKeys[k] {
			t.Fatalf("%s got reserved key %c", entries[i].Key, k)
		}
	}
	// build cannot take 'b'; should get a pool key (digit).
	if keys[0] < '1' || keys[0] > '9' {
		// still ok if it got another free letter; just not b.
		if keys[0] == 'b' {
			t.Fatal("build claimed reserved b")
		}
	}
}

func TestAssignCatalogKeys_uniqueStrings(t *testing.T) {
	entries := []config.CommandEntry{
		{Key: "alpha"}, {Key: "beta"}, {Key: "apple"},
	}
	keys := assignCatalogKeys(entries)
	// alpha gets a; apple cannot also get a.
	if keys[0] == keys[2] {
		t.Fatalf("alpha and apple both bound to %c", keys[0])
	}
	joined := string(keys)
	if strings.Count(joined, "a") > 1 {
		t.Fatalf("keys %q has duplicate a", joined)
	}
}
