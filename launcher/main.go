// hako -- host-side launcher for the agent stack (the Go engine; ADR-0012).
//
// Assembles the stack from the enabled integrations in hako.toml (skills,
// gateway config, settings env, sidecars) and wraps docker compose + the vault.
// Phase A: parity with the shell ./hako, reading manifests properly. Phase B
// adds the `configure` TUI; A2 moves the vault in-process (age + locked memory).
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/awnumar/memguard"
	"golang.org/x/term"
)

const gmuxURL = "http://localhost:8791/"

// version is set by the release build (-X main.version); "dev" for source builds.
var version = "dev"

func main() {
	memguard.CatchInterrupt()
	defer memguard.Purge()

	root, err := findRoot()
	if err != nil {
		fatal(err.Error())
	}
	if err := os.Chdir(root); err != nil {
		fatal(err.Error())
	}

	args := os.Args[1:]
	cmd := ""
	if len(args) > 0 {
		cmd, args = args[0], args[1:]
	}

	cfg, err := LoadConfig(root)
	if err != nil {
		fatal(err.Error())
	}
	files := composeFiles(cfg)

	switch cmd {
	case "up":
		mustAssemble(cfg)
		dc(files, append([]string{"up", "-d"}, args...)...)
		if cfg.HasVault() {
			unlock(cfg)
		}
		fmt.Printf("hako: up -- gmux UI at %s  (run 'hako auth' for the token)\n", gmuxURL)
	case "down":
		dc(files, append([]string{"down"}, args...)...)
	case "restart":
		dc(files, "down")
		mustAssemble(cfg)
		dc(files, "up", "-d")
		if cfg.HasVault() {
			unlock(cfg)
		}
	case "assemble":
		mustAssemble(cfg)
	case "ps":
		dc(files, append([]string{"ps"}, args...)...)
	case "logs":
		dc(files, append([]string{"logs", "-f"}, args...)...)
	case "shell":
		run("docker", "exec", "-it", "hako", "zsh")
	case "pi":
		run("docker", "exec", "-it", "hako", "gmux", "pi")
	case "auth":
		dc(files, "exec", "hako", "gmuxd", "auth")
	case "open":
		openBrowser(gmuxURL)
	case "version", "--version":
		fmt.Printf("hako %s\n", version)
	case "seal":
		name := "github"
		if len(args) > 0 {
			name = args[0]
		}
		seal(cfg, name)
	case "unlock":
		if !cfg.HasVault() {
			fatal("no vault (vault/*.age) found")
		}
		unlock(cfg)
	case "configure":
		runConfigure(cfg)
	case "", "-h", "--help", "help":
		usage(cfg)
	default:
		dc(files, append([]string{cmd}, args...)...)
	}
}

func mustAssemble(cfg *Config) {
	if err := assemble(cfg); err != nil {
		fatal("assemble: " + err.Error())
	}
}

// findRoot walks up from the cwd to the hako checkout (the dir with
// compose.yaml + integrations/), falling back to the executable's dir.
func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if isFile(filepath.Join(dir, "compose.yaml")) && isDir(filepath.Join(dir, "integrations")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if exe, err := os.Executable(); err == nil {
		d := filepath.Dir(exe)
		if isFile(filepath.Join(d, "compose.yaml")) {
			return d, nil
		}
	}
	return "", fmt.Errorf("not inside a hako checkout (no compose.yaml found)")
}

func dc(files []string, args ...string) {
	run("docker", append(append([]string{"compose"}, files...), args...)...)
}

func run(name string, args ...string) {
	c := exec.Command(name, args...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := c.Run(); err != nil {
		os.Exit(exitCode(err))
	}
}

func openBrowser(url string) {
	for _, try := range browserCmds(url) {
		if exec.Command(try[0], try[1:]...).Start() == nil {
			return
		}
	}
	fmt.Println(url)
}

func browserCmds(url string) [][]string {
	if runtime.GOOS == "darwin" {
		return [][]string{{"open", url}}
	}
	return [][]string{{"xdg-open", url}}
}

// stdinBuf is shared so successive readSecret calls on a pipe don't each
// buffer-and-drop the remaining lines.
var stdinBuf = bufio.NewReader(os.Stdin)

// readSecret reads a line with echo off on a tty; on a pipe (tests) it reads
// a plain line so the command stays scriptable.
func readSecret(prompt string) string {
	fmt.Fprint(os.Stderr, prompt)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, _ := term.ReadPassword(fd)
		fmt.Fprintln(os.Stderr)
		return string(b)
	}
	line, _ := stdinBuf.ReadString('\n')
	return strings.TrimRight(line, "\r\n")
}

func usage(cfg *Config) {
	fmt.Print(`hako -- launcher for the agent stack

usage: hako <command> [args]

  up [--build]    assemble + start (detached); unseals the vault if present
  down            stop and remove the stack
  restart         down, then up
  assemble        (re)generate the stack from hako.toml (skills + gateway config)
  configure       enable/disable integrations + set settings + seal secrets (TUI)
  seal [name]     encrypt a secret into the vault (default name: github)
  unlock          re-enter the vault passphrase (after a gateway restart)
  ps | logs [svc] | shell | pi | auth | open | version
  <other>         passed straight through to docker compose

integrations (integrations/, toggled in hako.toml):
`)
	for _, it := range cfg.Ints {
		mark := " "
		if it.Enabled {
			mark = "x"
		}
		fmt.Printf("  [%s] %-12s %s\n", mark, it.Name, it.Summary)
	}
}

func isFile(p string) bool { fi, err := os.Stat(p); return err == nil && !fi.IsDir() }
func isDir(p string) bool  { fi, err := os.Stat(p); return err == nil && fi.IsDir() }

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "hako: "+msg)
	os.Exit(1)
}

func exitCode(err error) int {
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	return 1
}
