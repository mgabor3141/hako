package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeAt(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// LoadConfig is the heart of the manifest model: it must merge the catalog with
// the user's hako.toml -- enabled flags, setting defaults vs overrides, and the
// sidecar gate. Exercise all of those against a temp catalog.
func TestLoadConfigResolution(t *testing.T) {
	root := t.TempDir()
	writeAt(t, filepath.Join(root, "compose.yaml"), "services: {}\n")
	writeAt(t, filepath.Join(root, "integrations/web/integration.toml"), `
name = "web"
summary = "web search"
[settings]
sidecar = { type = "bool", default = true }
url = { type = "string", default = "https://default" }
[sidecar]
compose = "compose.yaml"
enabled_by = "sidecar"
`)
	// a sidecar with no gate -> always on when enabled
	writeAt(t, filepath.Join(root, "integrations/always/integration.toml"), `
name = "always"
summary = "ungated sidecar"
[sidecar]
compose = "compose.yaml"
`)
	writeAt(t, filepath.Join(root, "integrations/gh/integration.toml"), `
name = "gh"
summary = "github"
`)
	writeAt(t, filepath.Join(root, "hako.toml"), `
[integrations.web]
enabled = true
sidecar = false
# url left unset -> should resolve to the manifest default
[integrations.always]
enabled = true
[integrations.gh]
enabled = false
`)

	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	web, always, gh := cfg.find("web"), cfg.find("always"), cfg.find("gh")
	if web == nil || always == nil || gh == nil {
		t.Fatalf("missing integration(s): %v", cfg.Ints)
	}

	if !web.Enabled || !always.Enabled || gh.Enabled {
		t.Errorf("enabled flags wrong: web=%v always=%v gh=%v", web.Enabled, always.Enabled, gh.Enabled)
	}
	if got := web.Values["url"]; got != "https://default" {
		t.Errorf("unset setting should resolve to default, got %q", got)
	}
	if got := web.Values["sidecar"]; got != "false" {
		t.Errorf("overridden setting should win, got %q", got)
	}
	if web.sidecarOn() {
		t.Error("web sidecar should be gated off by sidecar=false")
	}
	if !always.sidecarOn() {
		t.Error("ungated sidecar should be on when enabled")
	}

	enabled := cfg.Enabled()
	if len(enabled) != 2 {
		t.Errorf("Enabled() = %d integrations, want 2", len(enabled))
	}
}
