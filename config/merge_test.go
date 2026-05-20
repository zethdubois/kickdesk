package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverManifest_missingPublished(t *testing.T) {
	ref := AppRef{Label: "Test"}
	_, err := discoverManifest("missing-app", ref)
	if err == nil {
		t.Fatal("expected error for missing published manifest")
	}
	want := "no published manifest at"
	if err != nil && !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %q, want substring %q", err.Error(), want)
	}
}

func TestDiscoverManifest_explicitOverride(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")
	writeManifest(t, manifestPath, `{"id":"fixture-web","path":"`+dir+`","ports":[{"role":"app","port":5000}],"commands":{"up":"true","down":"true","build":"true"},"workflows":{"start":["up"],"stop":["down"]}}`)

	ref := AppRef{Manifest: manifestPath}
	got, err := discoverManifest("fixture-web", ref)
	if err != nil {
		t.Fatal(err)
	}
	if got != manifestPath {
		t.Fatalf("got %q, want %q", got, manifestPath)
	}
}

func TestResolveFromManifest_ignoresRepoKickdeskJSON(t *testing.T) {
	dir := t.TempDir()
	repoKickdesk := filepath.Join(dir, ".kickdesk.json")
	if err := os.WriteFile(repoKickdesk, []byte(`{"id":"fixture-web","path":"`+dir+`","ports":[{"role":"app","port":5000}],"commands":{"up":"true","down":"true","build":"true"},"workflows":{"start":["up"],"stop":["down"]}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	ref := AppRef{}
	_, _, err := resolveFromManifest("fixture-web", ref)
	if err == nil {
		t.Fatal("expected error when only repo .kickdesk.json exists")
	}
	if !strings.Contains(err.Error(), "no published manifest at") {
		t.Fatalf("error = %q, want missing published manifest", err.Error())
	}
}

func TestResolveFromManifest_missingPathInManifest(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")
	writeManifest(t, manifestPath, `{"id":"fixture-web","ports":[{"role":"app","port":5000}],"commands":{"up":"true","down":"true","build":"true"},"workflows":{"start":["up"],"stop":["down"]}}`)

	ref := AppRef{Manifest: manifestPath}
	_, _, err := resolveFromManifest("fixture-web", ref)
	if err == nil {
		t.Fatal("expected error for missing path in manifest")
	}
	if !strings.Contains(err.Error(), "missing path") {
		t.Fatalf("error = %q, want missing path", err.Error())
	}
}

func TestResolveFromManifest_pathFromManifest(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")
	writeManifest(t, manifestPath, `{
		"id": "fixture-web",
		"path": "`+dir+`",
		"ports": [{"role": "app", "port": 5000}],
		"commands": {"up": "true", "down": "true", "build": "true"},
		"workflows": {"start": ["up"], "stop": ["down"]}
	}`)

	ref := AppRef{Manifest: manifestPath}
	app, mpath, err := resolveFromManifest("fixture-web", ref)
	if err != nil {
		t.Fatal(err)
	}
	if mpath != manifestPath {
		t.Fatalf("manifest path = %q", mpath)
	}
	resolved, err := app.ResolvedPath()
	if err != nil {
		t.Fatal(err)
	}
	if resolved != dir {
		t.Fatalf("app path = %q, want %q", resolved, dir)
	}
}

func writeManifest(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
