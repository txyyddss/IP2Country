package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOverrideAndEvidenceValidation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		overrides string
		evidence  string
		wantError bool
	}{
		{"valid", `[{"cidr":"8.8.8.0/24","country":"hk"},{"cidr":"2001:4860::/32","country":"US"}]`, "fallback", false},
		{"default", `[]`, "", false},
		{"nested", `[{"cidr":"8.8.8.0/24","country":"HK"},{"cidr":"8.8.8.128/25","country":"JP"}]`, "geolocation", false},
		{"duplicate", `[{"cidr":"8.8.8.0/24","country":"HK"},{"cidr":"8.8.8.0/24","country":"hk"}]`, "", false},
		{"conflicting", `[{"cidr":"8.8.8.0/24","country":"HK"},{"cidr":"8.8.8.0/24","country":"JP"}]`, "", true},
		{"host-bits", `[{"cidr":"8.8.8.1/24","country":"HK"}]`, "", true},
		{"malformed", `[{"cidr":"8.217.174.44.0/24","country":"HK"}]`, "", true},
		{"mapped", `[{"cidr":"::ffff:8.8.8.0/120","country":"HK"}]`, "", true},
		{"country", `[{"cidr":"8.8.8.0/24","country":"EU"}]`, "", true},
		{"unknown-country", `[{"cidr":"8.8.8.0/24","country":"--"}]`, "", true},
		{"unknown-evidence", `[]`, "manual", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			evidence, _ := json.Marshal(tc.evidence)
			data := `{"version":1,"settings":{"ipv4_prefix":24,"ipv6_prefix":64},"sources":[{"id":"test","name":"Test","urls":["fixture"],"evidence_class":` + string(evidence) + `}],"overrides":` + tc.overrides + `}`
			if err := os.WriteFile(path, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(path)
			if (err != nil) != tc.wantError {
				t.Fatalf("config error=%v wanted error=%v", err, tc.wantError)
			}
			if err == nil && cfg.Sources[0].EvidenceClass == "" {
				t.Fatal("missing default evidence")
			}
			if err == nil && tc.name == "duplicate" && len(cfg.Overrides) != 1 {
				t.Fatal("duplicate override retained")
			}
		})
	}
}

func TestInventoryEvidenceAndOverrides(t *testing.T) {
	cfg, err := Load("../../config/sources.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Overrides) != 4 {
		t.Fatalf("expected four overrides: %+v", cfg.Overrides)
	}
	for _, o := range cfg.Overrides {
		if o.Country != "HK" {
			t.Fatalf("unexpected override %+v", o)
		}
	}
	groups := map[string]string{"v2ray-geoip": "dbip", "geoip-enhanced": "geolite", "meta-rules-dat": "geolite", "china-operator-46": "geolite", "geo-whois-v4": "sapics-country", "user-country-v4": "sapics-country", "server-country-sapics-v4": "sapics-country", "ip-geolocation-db": "sapics-country"}
	for _, s := range cfg.Sources {
		if group, ok := groups[s.ID]; ok {
			if s.VoteGroup != group || s.EvidenceClass != "fallback" {
				t.Fatalf("incorrect provenance: %+v", s)
			}
		}
	}
}

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
