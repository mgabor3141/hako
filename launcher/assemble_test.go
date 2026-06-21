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

// ~/.agents/skills is shared with skills the user or agent installs by hand.
// assemble must reconcile only its own catalog links and never wipe those.
func TestAssemblePreservesHandInstalledSkills(t *testing.T) {
	root := t.TempDir()
	writeAt(t, filepath.Join(root, "gateway/config.base.json"),
		`{"mcpProxy":{"addr":":8096"},"mcpServers":{}}`)
	writeAt(t, filepath.Join(root, "integrations/foo/integration.toml"),
		"name = \"foo\"\nsummary = \"foo\"\n[skill]\n")
	writeAt(t, filepath.Join(root, "integrations/foo/skill/SKILL.md"), "# foo")
	writeAt(t, filepath.Join(root, "hako.toml"), "[integrations.foo]\nenabled = true\n")

	// a skill the user/agent installed by hand into the shared dir
	mine := filepath.Join(root, "agent/.agents/skills/mine/SKILL.md")
	writeAt(t, mine, "# mine")

	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	// run twice: reconcile must be idempotent and never wipe the hand-installed skill
	for i := 1; i <= 2; i++ {
		if err := assemble(cfg); err != nil {
			t.Fatalf("assemble #%d: %v", i, err)
		}
		if _, err := os.Stat(mine); err != nil {
			t.Fatalf("hand-installed skill wiped on assemble #%d: %v", i, err)
		}
		tgt, err := os.Readlink(filepath.Join(root, "agent/.agents/skills/foo"))
		if err != nil || tgt != catalogMount+"/foo/skill" {
			t.Fatalf("foo link after assemble #%d: tgt=%q err=%v", i, tgt, err)
		}
	}
}

// A hand-installed skill that happens to share a catalog integration's name must
// not be clobbered by the catalog symlink.
func TestAssembleSkipsNameCollision(t *testing.T) {
	root := t.TempDir()
	writeAt(t, filepath.Join(root, "gateway/config.base.json"),
		`{"mcpProxy":{"addr":":8096"},"mcpServers":{}}`)
	writeAt(t, filepath.Join(root, "integrations/foo/integration.toml"),
		"name = \"foo\"\nsummary = \"foo\"\n[skill]\n")
	writeAt(t, filepath.Join(root, "integrations/foo/skill/SKILL.md"), "# foo")
	writeAt(t, filepath.Join(root, "hako.toml"), "[integrations.foo]\nenabled = true\n")

	// user installed their own "foo" (a real dir) before enabling the catalog one
	foo := filepath.Join(root, "agent/.agents/skills/foo")
	writeAt(t, filepath.Join(foo, "SKILL.md"), "# my own foo")

	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := assemble(cfg); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Lstat(foo)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Error("assemble clobbered a hand-installed skill that shares a catalog name")
	}
	if b, _ := os.ReadFile(filepath.Join(foo, "SKILL.md")); string(b) != "# my own foo" {
		t.Errorf("hand-installed foo content changed: %q", b)
	}
}
