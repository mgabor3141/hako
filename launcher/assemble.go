package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// catalogMount is where compose mounts ./integrations inside the agent
// container; skill symlinks point here so they resolve in-container.
const catalogMount = "/opt/hako/integrations"

// gatewayRoute is the agent-facing base of the in-network gateway; it matches
// the gateway service name + addr in gateway/compose.gateway.yaml and
// gateway/config.base.json. Each gateway integration is served at
// <gatewayRoute>/<name>/.
const gatewayRoute = "http://gateway:8096"

// composeFiles is the ordered -f list for docker compose: base + gateway,
// the vault overlay when a vault exists, and each enabled integration's sidecar
// (when its gate is on).
func composeFiles(cfg *Config) []string {
	f := []string{"-f", "compose.yaml", "-f", "gateway/compose.gateway.yaml"}
	if cfg.HasVault() {
		f = append(f, "-f", "gateway/compose.vault.yaml")
	}
	for _, it := range cfg.Enabled() {
		if it.sidecarOn() {
			f = append(f, "-f", filepath.Join("integrations", it.Name, it.Sidecar.Compose))
		}
	}
	return f
}

// assemble regenerates the stack from the enabled set: skill symlinks, the
// merged gateway config, and the settings env file.
func assemble(cfg *Config) error {
	enabled := cfg.Enabled()

	// 1. pi's skill dir: only enabled skills, symlinked to the mounted catalog.
	skillDir := filepath.Join(cfg.Root, "agent", ".agents", "skills")
	if err := os.RemoveAll(skillDir); err != nil {
		return err
	}
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return err
	}
	for _, it := range enabled {
		if it.Skill == nil {
			continue
		}
		sub := it.Skill.Dir
		if sub == "" {
			sub = "skill"
		}
		target := fmt.Sprintf("%s/%s/%s", catalogMount, it.Name, sub)
		if err := os.Symlink(target, filepath.Join(skillDir, it.Name)); err != nil {
			return err
		}
	}

	// 2. gateway config: merge enabled backends into the base skeleton.
	var base struct {
		McpProxy   json.RawMessage            `json:"mcpProxy"`
		McpServers map[string]json.RawMessage `json:"mcpServers"`
	}
	raw, err := os.ReadFile(filepath.Join(cfg.Root, "gateway", "config.base.json"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &base); err != nil {
		return fmt.Errorf("gateway/config.base.json: %w", err)
	}
	if base.McpServers == nil {
		base.McpServers = map[string]json.RawMessage{}
	}
	for _, it := range enabled {
		if it.Gateway == nil || it.Gateway.Config == "" {
			continue
		}
		gj, err := os.ReadFile(filepath.Join(it.dir, it.Gateway.Config))
		if err != nil {
			return err
		}
		base.McpServers[it.Name] = json.RawMessage(gj)
	}
	out, err := json.MarshalIndent(base, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cfg.Root, "gateway", "config.json"), append(out, '\n'), 0o644); err != nil {
		return err
	}

	// 3. settings env for the agent container (+ compose interpolation).
	var env strings.Builder
	for _, it := range enabled {
		nm := envKey(it.Name)
		for k, v := range it.Values {
			fmt.Fprintf(&env, "HAKO_%s_%s=%s\n", nm, envKey(k), v)
		}
	}
	// gateway integrations: wire the agent's route to each (no auth -- the
	// private network is the boundary). The skill reads <NAME>_MCP_URL.
	for _, it := range enabled {
		if it.Gateway != nil {
			fmt.Fprintf(&env, "%s_MCP_URL=%s/%s/\n", strings.ToUpper(it.Name), gatewayRoute, it.Name)
		}
	}
	if err := os.WriteFile(filepath.Join(cfg.Root, ".hako.env"), []byte(env.String()), 0o644); err != nil {
		return err
	}

	names := make([]string, 0, len(enabled))
	for _, it := range enabled {
		names = append(names, it.Name)
	}
	fmt.Printf("hako: assembled (enabled: %s)\n", orNone(strings.Join(names, " ")))
	return nil
}

func envKey(s string) string {
	return strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace(s))
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}
