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

// agentContainer is the compose service/container name the human attaches to.
const agentContainer = "hako"

// shellArgs is the `docker exec` invocation that drops the human into the agent
// container's interactive zsh (pi runs bash; humans get zsh -- ADR-0005). Shared
// by the `shell` verb and the configure TUI so the two can't drift.
func shellArgs() []string { return []string{"exec", "-it", agentContainer, "zsh"} }

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
	composeEnv = settingsEnv(cfg) // resolved settings for sidecar interpolation
	files := composeFiles(cfg)

	switch cmd {
	case "up":
		mustAssemble(cfg)
		// --remove-orphans reconciles: a now-disabled integration's container is
		// torn down even though its overlay is no longer in the compose set.
		dc(files, append([]string{"up", "-d", "--remove-orphans"}, args...)...)
		if cfg.HasVault() {
			unlock(cfg)
		}
		fmt.Printf("hako: up -- gmux UI at %s  (run 'hako token' for the UI token)\n", gmuxURL)
	case "down":
		// --remove-orphans so down cleans up every container in the hako project,
		// including sidecars from integrations disabled since they were started.
		dc(files, append([]string{"down", "--remove-orphans"}, args...)...)
	case "restart":
		dc(files, "down", "--remove-orphans")
		mustAssemble(cfg)
		dc(files, "up", "-d", "--remove-orphans")
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
		run("docker", shellArgs()...)
	case "pi":
		run("docker", "exec", "-it", agentContainer, "gmux", "pi")
	case "token":
		dc(files, "exec", "hako", "gmuxd", "auth")
	case "open":
		openBrowser(gmuxURL)
	case "version", "--version":
		fmt.Printf("hako %s\n", version)
	case "auth":
		name := "github"
		if len(args) > 0 {
			name = args[0]
		}
		setupAuth(cfg, name)
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

// composeEnv holds resolved HAKO_<INT>_<SETTING> vars, passed to docker compose
// so sidecar overlays can interpolate them.
var composeEnv []string

func dc(files []string, args ...string) {
	c := exec.Command("docker", append(append([]string{"compose"}, files...), args...)...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	c.Env = append(os.Environ(), composeEnv...)
	if err := c.Run(); err != nil {
		os.Exit(exitCode(err))
	}
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
		s := readMasked(fd)
		fmt.Fprintln(os.Stderr)
		return s
	}
	line, _ := stdinBuf.ReadString('\n')
	return strings.TrimRight(line, "\r\n")
}

// readMasked reads a secret from the terminal in raw mode, echoing a '•' per
// character so the user sees their typing land without revealing it. Handles
// backspace and ctrl-c; falls back to a no-echo read if raw mode is unavailable.
func readMasked(fd int) string {
	old, err := term.MakeRaw(fd)
	if err != nil {
		b, _ := term.ReadPassword(fd)
		return string(b)
	}
	defer term.Restore(fd, old)
	var buf []byte
	b := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(b)
		if n == 0 || err != nil {
			break
		}
		switch c := b[0]; c {
		case '\r', '\n':
			return string(buf)
		case 3: // ctrl-c
			term.Restore(fd, old)
			fmt.Fprintln(os.Stderr)
			os.Exit(130)
		case 8, 127: // backspace / delete
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Fprint(os.Stderr, "\b \b")
			}
		default:
			if c >= 0x20 { // skip other control bytes
				buf = append(buf, c)
				fmt.Fprint(os.Stderr, "\u2022")
			}
		}
	}
	return string(buf)
}

func usage(cfg *Config) {
	fmt.Print(`hako -- launcher for the agent stack

  new here? run 'hako configure' for an interactive menu
  (toggle integrations, set options, set up auth, drop to a shell)

usage: hako <command> [args]

  up [--build]    assemble + start (detached); unlocks credentials if present
  down            stop and remove the stack
  restart         down, then up
  assemble        (re)generate the stack from hako.toml (skills + gateway config)
  configure       enable/disable integrations + set options + set up auth (TUI)
  auth [name]     set up an integration's credentials (default name: github)
  unlock          re-enter the credentials passphrase (after a gateway restart)
  ps | logs [svc] | shell | pi | token | open | version
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
