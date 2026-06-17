package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var nameRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// unlock pipes the (masked) passphrase to the gateway's in-container unlock.
// Phase A keeps the existing decrypt path; Phase A2 moves it in-process
// (filippo.io/age + locked memory) and to a single multi-secret vault.
func unlock(cfg *Config) {
	for i := 0; i < 30; i++ {
		if exec.Command("docker", "exec", "hako-gateway", "true").Run() == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	pass := readSecret("hako: vault passphrase: ")
	c := exec.Command("docker", "exec", "-i", "hako-gateway", "/seal/hako-unlock")
	c.Stdin = strings.NewReader(pass)
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	if err := c.Run(); err != nil {
		os.Exit(exitCode(err))
	}
}

// seal age-encrypts a secret into vault/<name>.age using the gateway image
// (which ships age + script). Secret on stdin, passphrase via env to the
// ephemeral container.
func seal(cfg *Config, name string) {
	if !nameRe.MatchString(name) {
		fatal("bad secret name " + name)
	}
	if exec.Command("docker", "image", "inspect", "hako-gateway").Run() != nil {
		fmt.Fprintln(os.Stderr, "hako: building the gateway image (first run)...")
		dc(composeFiles(cfg, false), "build", "gateway")
	}
	secret := readSecret(fmt.Sprintf("secret value for %q (hidden): ", name))
	if secret == "" {
		fatal("empty secret, aborting")
	}
	if p1, p2 := readSecret("vault passphrase (hidden): "), readSecret("confirm passphrase (hidden): "); p1 != p2 {
		fatal("passphrases do not match")
	} else {
		if err := os.MkdirAll(filepath.Join(cfg.Root, "vault"), 0o700); err != nil {
			fatal(err.Error())
		}
		script := fmt.Sprintf(
			`umask 077; cat > /tmp/s; printf '%%s\n%%s\n' "$PASS" "$PASS" | `+
				`script -qec "age -p -o /out/%s.age /tmp/s" /dev/null >/dev/null 2>&1; `+
				`rm -f /tmp/s; [ -s /out/%s.age ]`, name, name)
		c := exec.Command("docker", "run", "--rm", "-i",
			"-e", "PASS="+p1,
			"-v", filepath.Join(cfg.Root, "vault")+":/out",
			"--entrypoint", "sh", "hako-gateway", "-c", script)
		c.Stdin = strings.NewReader(secret)
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			fatal("seal FAILED")
		}
	}
	fmt.Printf("hako: sealed vault/%s.age -- start with 'hako up' (or 'hako unlock')\n", name)
}
