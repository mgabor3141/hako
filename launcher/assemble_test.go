package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// assemble is the heart of the integration model: enabled integrations get a
// skill symlink, their gateway backend merged in, and their settings exported;
// disabled ones leave no trace. Drive it against a temp catalog and check all
// three outputs.
func TestAssemble(t *testing.T) {
	root := t.TempDir()
	writeAt(t, filepath.Join(root, "gateway/config.base.json"),
		`{"mcpProxy":{"addr":":8096"},"mcpServers":{}}`)
	writeAt(t, filepath.Join(root, "integrations/foo/integration.toml"), `
name = "foo"
summary = "foo"
[skill]
[gateway]
config = "gateway.json"
[settings]
mode = { type = "string", default = "fast" }
`)
	writeAt(t, filepath.Join(root, "integrations/foo/gateway.json"),
		`{"url":"https://foo","type":"streamable-http"}`)
	writeAt(t, filepath.Join(root, "integrations/foo/skill/SKILL.md"), "# foo")
	writeAt(t, filepath.Join(root, "integrations/bar/integration.toml"),
		"name = \"bar\"\nsummary = \"bar\"\n[skill]\n")
	writeAt(t, filepath.Join(root, "integrations/bar/skill/SKILL.md"), "# bar")
	writeAt(t, filepath.Join(root, "hako.toml"), `
[integrations.foo]
enabled = true
[integrations.bar]
enabled = false
`)

	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := assemble(cfg); err != nil {
		t.Fatal(err)
	}

	// enabled skill -> symlink into the in-container catalog mount
	tgt, err := os.Readlink(filepath.Join(root, "agent/.agents/skills/foo"))
	if err != nil {
		t.Fatalf("foo skill symlink: %v", err)
	}
	if tgt != catalogMount+"/foo/skill" {
		t.Errorf("symlink target = %q, want %q", tgt, catalogMount+"/foo/skill")
	}
	// disabled integration leaves no trace
	if _, err := os.Lstat(filepath.Join(root, "agent/.agents/skills/bar")); !os.IsNotExist(err) {
		t.Error("disabled bar should not be linked")
	}
	// gateway config gains foo's backend
	gw, _ := os.ReadFile(filepath.Join(root, "gateway/config.json"))
	if !strings.Contains(string(gw), `"foo"`) || !strings.Contains(string(gw), "https://foo") {
		t.Errorf("gateway config missing foo backend:\n%s", gw)
	}
	// resolved setting is exported for the agent container
	env, _ := os.ReadFile(filepath.Join(root, ".hako.env"))
	if !strings.Contains(string(env), "HAKO_FOO_MODE=fast") {
		t.Errorf(".hako.env missing setting:\n%s", env)
	}
}
