package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestEndToEndBuildWithSingBoxCompiler(t *testing.T) {
	binary := os.Getenv("SING_BOX_BIN")
	if binary == "" {
		t.Skip("set SING_BOX_BIN to run the integration test")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("network,country_code\n8.8.8.0/24,US\n2001:4860::/64,US\n"))
	}))
	defer server.Close()

	root := t.TempDir()
	configPath := filepath.Join(root, "sources.json")
	configJSON := fmt.Sprintf(`{"version":1,"settings":{"ipv4_prefix":24,"ipv6_prefix":64,"special_use_source":"iana-global-blocks"},"sources":[{"id":"fixture","name":"Fixture","vote_group":"fixture","format":"csv","urls":[%q]}]}`, server.URL)
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
}

func contains(s, part string) bool {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
