package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

// Setting is a typed knob an integration declares in its manifest.
type Setting struct {
	Type        string `toml:"type"` // string | number | bool
	Default     any    `toml:"default"`
	Description string `toml:"description"`
}

// Manifest is an integration's integration.toml. Every section is optional.
type Manifest struct {
	Name    string `toml:"name"`
	Summary string `toml:"summary"`
	// AuthGuide is multi-line setup help shown before `hako auth <name>` prompts
	// for a secret (e.g. where to make a token and which scopes to grant).
	AuthGuide string `toml:"auth_guide"`
	Skill     *struct {
		Dir      string   `toml:"dir"`
		Commands []string `toml:"commands"`
	} `toml:"skill"`
	Gateway *struct {
		Config string `toml:"config"`
	} `toml:"gateway"`
	Settings map[string]Setting `toml:"settings"`
	Sidecar  *struct {
		Compose   string `toml:"compose"`
		EnabledBy string `toml:"enabled_by"`
	} `toml:"sidecar"`
	Secrets []struct {
		Env    string `toml:"env"`
		Prompt string `toml:"prompt"`
	} `toml:"secrets"`

	dir string // catalog dir, e.g. integrations/github
}

// Integration is a catalog manifest plus the user's choices from hako.toml.
type Integration struct {
	Manifest
	Enabled bool
	Values  map[string]string // resolved settings (user override, else default)
}

// Networks lists extra external docker networks the stack joins, so an
// integration pointed at a private endpoint (sidecar = false, url = ...) can
// resolve it by service DNS. Agent nets attach to the hako container (the usual
// case: skill integrations call the endpoint directly); gateway nets attach to
// the gateway container (a private MCP backend it fronts). The networks must
// already exist (external) -- typically another compose project's <name>_default.
type Networks struct {
	Agent   []string `toml:"agent"`
	Gateway []string `toml:"gateway"`
}

type Config struct {
	Root     string
	Ints     []*Integration // every catalog integration, Enabled reflecting hako.toml
	Networks Networks       // extra external networks from [networks] in hako.toml
}

// hakoFile is the structure of hako.toml / hako.example.toml.
type hakoFile struct {
	Integrations map[string]map[string]any `toml:"integrations"`
	Networks     Networks                  `toml:"networks"`
}

func LoadConfig(root string) (*Config, error) {
	// user choices: hako.toml, falling back to the shipped example
	choicesPath := filepath.Join(root, "hako.toml")
	if _, err := os.Stat(choicesPath); err != nil {
		choicesPath = filepath.Join(root, "hako.example.toml")
	}
	var hf hakoFile
	if _, err := os.Stat(choicesPath); err == nil {
		if _, err := toml.DecodeFile(choicesPath, &hf); err != nil {
			return nil, fmt.Errorf("reading %s: %w", filepath.Base(choicesPath), err)
		}
	}

	// catalog: integrations/*/integration.toml
	entries, _ := os.ReadDir(filepath.Join(root, "integrations"))
	cfg := &Config{Root: root}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, "integrations", e.Name())
		mfPath := filepath.Join(dir, "integration.toml")
		if _, err := os.Stat(mfPath); err != nil {
			continue
		}
		var m Manifest
		if _, err := toml.DecodeFile(mfPath, &m); err != nil {
			return nil, fmt.Errorf("reading %s: %w", mfPath, err)
		}
		m.dir = dir
		if m.Name == "" {
			m.Name = e.Name()
		}
		choice := hf.Integrations[m.Name]
		it := &Integration{Manifest: m, Values: map[string]string{}}
		if v, ok := choice["enabled"].(bool); ok {
			it.Enabled = v
		}
		// resolve declared settings: user override, else manifest default
		for key, s := range m.Settings {
			if uv, ok := choice[key]; ok {
				it.Values[key] = toStr(uv)
			} else if s.Default != nil {
				it.Values[key] = toStr(s.Default)
			}
		}
		cfg.Ints = append(cfg.Ints, it)
	}
	sort.Slice(cfg.Ints, func(i, j int) bool { return cfg.Ints[i].Name < cfg.Ints[j].Name })
	cfg.Networks = hf.Networks
	return cfg, nil
}

// hasExtraNetworks reports whether any external network is configured.
func (c *Config) hasExtraNetworks() bool {
	return len(c.Networks.Agent) > 0 || len(c.Networks.Gateway) > 0
}

func (c *Config) Enabled() []*Integration {
	var out []*Integration
	for _, it := range c.Ints {
		if it.Enabled {
			out = append(out, it)
		}
	}
	return out
}

func (c *Config) find(name string) *Integration {
	for _, it := range c.Ints {
		if it.Name == name {
			return it
		}
	}
	return nil
}

func (c *Config) HasVault() bool {
	m, _ := filepath.Glob(filepath.Join(c.Root, "vault", "*.age"))
	return len(m) > 0
}

// sidecarOn reports whether an enabled integration's sidecar should run:
// true when it has a [sidecar] and either no enabled_by gate or the gated
// bool setting resolves true.
func (it *Integration) sidecarOn() bool {
	if it.Sidecar == nil || it.Sidecar.Compose == "" {
		return false
	}
	if it.Sidecar.EnabledBy == "" {
		return true
	}
	return it.Values[it.Sidecar.EnabledBy] != "false"
}

func toStr(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case int64:
		return fmt.Sprintf("%d", x)
	case float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", x), "0"), ".")
	default:
		return fmt.Sprint(x)
	}
}
