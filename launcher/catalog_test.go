package main

import "testing"

// The synthetic catalog in assemble_test/config_test proves the logic; this
// guards the *actual* integrations/ we ship -- a malformed manifest or a bad
// setting type would otherwise only surface at runtime. Run from launcher/, so
// the repo root is "..".
func TestRealCatalogParses(t *testing.T) {
	cfg, err := LoadConfig("..")
	if err != nil {
		t.Fatalf("LoadConfig on the real catalog: %v", err)
	}
	for _, name := range []string{"github", "websearch", "webview", "backups"} {
		if cfg.find(name) == nil {
			t.Errorf("catalog is missing integration %q", name)
		}
	}
	for _, it := range cfg.Ints {
		if it.Name == "" || it.Summary == "" {
			t.Errorf("integration %q is missing name/summary", it.Name)
		}
		for k, s := range it.Settings {
			switch s.Type {
			case "string", "number", "bool":
			default:
				t.Errorf("%s.%s: bad setting type %q (want string/number/bool)", it.Name, k, s.Type)
			}
		}
		if it.Sidecar != nil && it.Sidecar.Compose == "" {
			t.Errorf("%s: [sidecar] without a compose file", it.Name)
		}
	}
}
