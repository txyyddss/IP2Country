package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultSourceInventoryLoads(t *testing.T) {
	cfg, err := Load("../../config/sources.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Sources) < 30 {
		t.Fatalf("expected the full source inventory, got %d", len(cfg.Sources))
	}
	for _, source := range cfg.Sources {
		if source.Enabled == nil {
			t.Errorf("source %q has no enabled state", source.ID)
		}
		if strings.Contains(strings.ToLower(source.Name), "pikonetwork") {
			t.Errorf("PikoNetwork remains configured")
		}
		if source.Format == "tsv" && source.IPEncoding != "hex" {
			t.Errorf("source %s must use hexadecimal endpoints", source.ID)
		}
	}
}

func TestEnabledDefaultsAndExplicitOptOut(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := `{"version":1,"settings":{"ipv4_prefix":24,"ipv6_prefix":64},"sources":[{"id":"default","name":"Default","urls":["fixture"]},{"id":"disabled","name":"Disabled","urls":["fixture"],"enabled":false}]}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !*cfg.Sources[0].Enabled || *cfg.Sources[1].Enabled {
		t.Fatalf("enabled defaults not preserved: %+v", cfg.Sources)
	}
}
