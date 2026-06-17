package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"filippo.io/age"
	"github.com/awnumar/memguard"
)

// The vault is a single age-passphrase-encrypted file holding every secret
// (one global passphrase -- ADR-0011). Decryption happens HERE, on the host,
// in locked memory (filippo.io/age + memguard); the gateway never sees the
// passphrase or the ciphertext -- it just receives the decrypted env over a
// pipe and sources it. No `script`/pty hack, no age in the gateway, no secret
// in any process's argv/env at rest.

func vaultPath(cfg *Config) string { return filepath.Join(cfg.Root, "vault", "secrets.age") }

// decryptVault reads the vault into a locked buffer (caller must Destroy it).
func decryptVault(cfg *Config, pass string) (*memguard.LockedBuffer, error) {
	ct, err := os.ReadFile(vaultPath(cfg))
	if err != nil {
		return nil, err
	}
	id, err := age.NewScryptIdentity(pass)
	if err != nil {
		return nil, err
	}
	r, err := age.Decrypt(bytes.NewReader(ct), id)
	if err != nil {
		return nil, err // wrong passphrase or corrupt vault
	}
	return memguard.NewBufferFromEntireReader(r)
}

// encryptVault writes the secrets (KEY=VALUE lines) to the vault. The plaintext
// buffer is zeroed before returning.
func encryptVault(cfg *Config, pass string, secrets map[string]string) error {
	var pt bytes.Buffer
	for k, v := range secrets {
		fmt.Fprintf(&pt, "%s=%s\n", k, v)
	}
	rcp, err := age.NewScryptRecipient(pass)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	w, err := age.Encrypt(&out, rcp)
	if err != nil {
		return err
	}
	if _, err := w.Write(pt.Bytes()); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	zero(pt.Bytes())
	if err := os.MkdirAll(filepath.Dir(vaultPath(cfg)), 0o700); err != nil {
		return err
	}
	return os.WriteFile(vaultPath(cfg), out.Bytes(), 0o600)
}

// parseEnv turns "KEY=VALUE" lines into a map.
func parseEnv(b []byte) map[string]string {
	m := map[string]string{}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			m[strings.TrimSpace(k)] = v
		}
	}
	return m
}

// seal collects an integration's declared secrets, merges them into the single
// vault under the global passphrase, and re-encrypts.
func seal(cfg *Config, name string) {
	it := cfg.find(name)
	if it == nil {
		fatal("unknown integration: " + name)
	}
	if len(it.Secrets) == 0 {
		fatal(name + " declares no secrets")
	}
	add := map[string]string{}
	for _, s := range it.Secrets {
		prompt := s.Prompt
		if prompt == "" {
			prompt = s.Env
		}
		v := readSecret(prompt + " (hidden): ")
		if v == "" {
			fatal("empty secret, aborting")
		}
		add[s.Env] = v
	}

	exists := isFile(vaultPath(cfg))
	var pass string
	if exists {
		pass = readSecret("vault passphrase (hidden): ")
	} else {
		p1 := readSecret("new vault passphrase (hidden): ")
		if p2 := readSecret("confirm passphrase (hidden): "); p1 != p2 {
			fatal("passphrases do not match")
		}
		pass = p1
	}

	secrets := map[string]string{}
	if exists {
		buf, err := decryptVault(cfg, pass)
		if err != nil {
			fatal("could not open the vault (wrong passphrase?)")
		}
		secrets = parseEnv(buf.Bytes())
		buf.Destroy()
	}
	for k, v := range add {
		secrets[k] = v
	}
	if err := encryptVault(cfg, pass, secrets); err != nil {
		fatal("seal failed: " + err.Error())
	}
	fmt.Printf("hako: sealed %d secret(s) for %q into vault/secrets.age\n", len(add), name)
}

// unlock decrypts the vault on the host and pipes the env into the gateway's
// tmpfs; boot.sh sources it and starts mcp-proxy.
func unlock(cfg *Config) {
	if !isFile(vaultPath(cfg)) {
		fatal("no vault at vault/secrets.age (run: hako seal <integration>)")
	}
	pass := readSecret("hako: vault passphrase: ")
	buf, err := decryptVault(cfg, pass)
	if err != nil {
		fatal("could not open the vault (wrong passphrase?)")
	}
	defer buf.Destroy()

	// build the sourceable env (export KEY='value'), shell-quoted.
	var env bytes.Buffer
	for k, v := range parseEnv(buf.Bytes()) {
		fmt.Fprintf(&env, "export %s='%s'\n", k, strings.ReplaceAll(v, "'", `'\''`))
	}
	defer zero(env.Bytes())

	for i := 0; i < 30; i++ {
		if exec.Command("docker", "exec", "hako-gateway", "true").Run() == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	c := exec.Command("docker", "exec", "-i", "hako-gateway", "sh", "-c",
		"umask 077; mkdir -p /run/hako; cat > /run/hako/env")
	c.Stdin = bytes.NewReader(env.Bytes())
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		fatal("unlock: could not deliver secrets to the gateway: " + err.Error())
	}
	fmt.Println("hako: unsealed.")
}

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
