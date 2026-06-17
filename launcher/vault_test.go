package main

import (
	"os/exec"
	"reflect"
	"testing"
)

// The vault is the most security-sensitive code, so test the crypto core
// directly: a real age round-trip through the on-disk file, wrong-passphrase
// rejection, and that secret delivery survives a real shell.

func TestVaultRoundTrip(t *testing.T) {
	cfg := &Config{Root: t.TempDir()}
	secrets := map[string]string{
		"GITHUB_MCP_TOKEN": "ghp_abc123",
		"WITH_EQUALS":      "a=b=c",
		"WITH_SPACES":      "value with spaces and 'quotes'",
	}
	if err := encryptVault(cfg, "correct horse battery", secrets); err != nil {
		t.Fatal(err)
	}

	if _, err := decryptVault(cfg, "wrong passphrase"); err == nil {
		t.Fatal("decrypt with the wrong passphrase should fail")
	}

	buf, err := decryptVault(cfg, "correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	got := parseEnv(buf.Bytes())
	buf.Destroy()
	if !reflect.DeepEqual(got, secrets) {
		t.Fatalf("round-trip mismatch:\n got %#v\nwant %#v", got, secrets)
	}
}

// renderEnv's output is sourced by the gateway, so prove the shell-quoting
// survives a real `sh` for nasty values (quotes, $, backslash, spaces).
func TestRenderEnvSurvivesShell(t *testing.T) {
	want := `a'b $HOME "c" \d e`
	env := renderEnv(map[string]string{"SECRET": want})
	out, err := exec.Command("sh", "-c", string(env)+`printf '%s' "$SECRET"`).Output()
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != want {
		t.Fatalf("shell round-trip: got %q want %q", out, want)
	}
}
