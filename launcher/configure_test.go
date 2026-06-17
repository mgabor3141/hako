package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func fakeCfg(dir string) *Config {
	return &Config{Root: dir, Ints: []*Integration{
		{Manifest: Manifest{Name: "github", Summary: "gh"}, Enabled: false, Values: map[string]string{}},
		{Manifest: Manifest{Name: "websearch", Summary: "ws", Settings: map[string]Setting{
			"sidecar": {Type: "bool", Default: true},
			"url":     {Type: "string", Default: ""},
		}}, Enabled: true, Values: map[string]string{"sidecar": "true", "url": ""}},
	}}
}

func TestConfigureModelAndSave(t *testing.T) {
	dir := t.TempDir()
	cfg := fakeCfg(dir)
	var m tea.Model = newCfgModel(cfg)
	send := func(k string) { m, _ = m.Update(key(k)) }

	send("space") // enable github (cursor at row 0)
	if !cfg.Ints[0].Enabled {
		t.Fatal("github should be enabled after space")
	}
	send("j")     // -> websearch
	send("enter") // expand its settings: rows become [github, websearch, sidecar, url]
	send("j")     // -> sidecar
	send("space") // sidecar true -> false
	if cfg.Ints[1].Values["sidecar"] != "false" {
		t.Fatalf("sidecar should be false, got %q", cfg.Ints[1].Values["sidecar"])
	}
	send("j")          // -> url
	send("enter")      // edit url
	send("x.com")      // type
	send("enter")      // confirm
	if got := cfg.Ints[1].Values["url"]; got != "x.com" {
		t.Fatalf("url = %q, want x.com", got)
	}
	send("w") // save

	out, err := os.ReadFile(filepath.Join(dir, "hako.toml"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{
		"[integrations.github]\nenabled = true",
		"[integrations.websearch]\nenabled = true",
		"sidecar = false",
		`url = "x.com"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("hako.toml missing %q\n---\n%s", want, s)
		}
	}
}

func TestWriteHakoOmitsDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg := fakeCfg(dir) // websearch settings all at default
	if err := writeHako(cfg); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(filepath.Join(dir, "hako.toml"))
	if strings.Contains(string(out), "sidecar =") || strings.Contains(string(out), "url =") {
		t.Errorf("default settings should be omitted, got:\n%s", out)
	}
}
