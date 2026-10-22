package main

import (
	"encoding/json"
	"fmt"
	"github.com/oschwald/maxminddb-golang/v2"
	"github.com/txyyddss/ip2country/internal/config"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestEndToEndBuildWithSingBoxCompiler(t *testing.T) {
	binary := os.Getenv("SING_BOX_BIN")
	if binary == "" {
		t.Skip("set SING_BOX_BIN to run the integration test")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("network,country_code\n8.8.8.0/25,US\n8.8.8.128/25,JP\n2001:4860::/65,US\n2001:4860:0:0:8000::/65,JP\n"))
	}))
	defer server.Close()

	root := t.TempDir()
	configPath := filepath.Join(root, "sources.json")
	defaults, err := config.Load("../../config/sources.json")
	if err != nil {
		t.Fatal(err)
	}
	overrides, err := json.Marshal(defaults.Overrides)
	if err != nil {
		t.Fatal(err)
	}
	configJSON := fmt.Sprintf(`{"version":1,"settings":{"ipv4_prefix":24,"ipv6_prefix":64},"sources":[{"id":"fixture","name":"Fixture","vote_group":"fixture","format":"csv","urls":[%q]}],"overrides":%s}`, server.URL, overrides)
	if err := os.WriteFile(configPath, []byte(configJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	workDir := filepath.Join(root, "work")
	outDir := filepath.Join(root, "dist")
	readme := filepath.Join(root, "README.md")
	if err := os.WriteFile(readme, []byte("# Fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := build(configPath, workDir, outDir, readme, binary, false); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"geoip.dat", "country.mmdb", "country-csv.zip", "source-csv.zip", "sing-box-rulesets.zip", "run-summary.json", "manifest.json", "SHA256SUMS"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Errorf("missing generated asset %s: %v", name, err)
		}
	}
	readmeBytes, err := os.ReadFile(readme)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(readmeBytes), "Fixture") || !contains(string(readmeBytes), "US") {
		t.Fatalf("README statistics were not refreshed: %s", readmeBytes)
	}
	db, err := maxminddb.Open(filepath.Join(outDir, "country.mmdb"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	lookups := map[string]string{"8.8.8.1": "US", "8.8.8.200": "JP", "2001:4860::1": "US", "2001:4860:0:0:8000::1": "JP"}
	for _, override := range defaults.Overrides {
		p := netip.MustParsePrefix(override.CIDR)
		for ip, count := p.Addr(), 0; ip.IsValid() && p.Contains(ip) && count < 1024; ip, count = ip.Next(), count+1 {
			lookups[ip.String()] = override.Country
		}
	}
	for ip, want := range lookups {
		var rec struct {
			Country struct {
				Code string `maxminddb:"iso_code"`
			} `maxminddb:"country"`
		}
		if err := db.Lookup(netip.MustParseAddr(ip)).Decode(&rec); err != nil {
			t.Fatal(err)
		}
		if rec.Country.Code != want {
			t.Fatalf("%s: got %s want %s", ip, rec.Country.Code, want)
		}
	}
}

func contains(s, part string) bool {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return true
		}
	}
	return false
}

func TestManualOverridesCannotPublishWithoutUpstreamData(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "config.json")
	data := `{"version":1,"settings":{"ipv4_prefix":24,"ipv6_prefix":64},"sources":[],"overrides":[{"cidr":"8.8.8.0/24","country":"HK"}]}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	err := build(path, filepath.Join(root, "work"), filepath.Join(root, "dist"), filepath.Join(root, "README.md"), "", true)
	if err == nil || !strings.Contains(err.Error(), "no country source produced") {
		t.Fatalf("unexpected build result: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "dist", "geoip.dat")); !os.IsNotExist(err) {
		t.Fatal("published override-only database")
	}
}

func TestInvalidOverrideFailsBeforeDownloadsOrOutputChanges(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	root := t.TempDir()
	path := filepath.Join(root, "config.json")
	data := fmt.Sprintf(`{"version":1,"settings":{"ipv4_prefix":24,"ipv6_prefix":64},"sources":[{"id":"fixture","name":"Fixture","format":"csv","urls":[%q]}],"overrides":[{"cidr":"8.217.174.44.0/24","country":"HK"}]}`, server.URL)
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	err := build(path, filepath.Join(root, "work"), filepath.Join(root, "dist"), filepath.Join(root, "README.md"), "", false)
	if err == nil || requests.Load() != 0 {
		t.Fatalf("invalid override did not fail early: %v requests=%d", err, requests.Load())
	}
	if _, err := os.Stat(filepath.Join(root, "dist")); !os.IsNotExist(err) {
		t.Fatal("invalid configuration changed outputs")
	}
}
