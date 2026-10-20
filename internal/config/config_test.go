package config

import "testing"

func TestDefaultSourceInventoryLoadsEnabled(t *testing.T) {
	cfg, err := Load("../../config/sources.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Sources) < 30 {
		t.Fatalf("expected the full source inventory, got %d", len(cfg.Sources))
	}
	for _, source := range cfg.Sources {
		if source.Enabled == nil || !*source.Enabled {
			t.Errorf("default source %q is not enabled", source.ID)
		}
	}
}
