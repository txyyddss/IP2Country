package output

import (
	"archive/zip"
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/oschwald/maxminddb-golang/v2"
	"github.com/txyyddss/ip2country/internal/config"
	"github.com/txyyddss/ip2country/internal/iprange"
	"github.com/txyyddss/ip2country/internal/model"
	"github.com/v2fly/geoip/plugin/v2ray"
	"go4.org/netipx"
	"google.golang.org/protobuf/proto"
)

// Run after a complete generation with IP2COUNTRY_DIST and SING_BOX_BIN set.
// Compare complete address sets so harmless format-specific CIDR merges are allowed.
func TestGeneratedOverridesAndIntegrity(t *testing.T) {
	root, binary := os.Getenv("IP2COUNTRY_DIST"), os.Getenv("SING_BOX_BIN")
	if root == "" || binary == "" {
		t.Skip("set IP2COUNTRY_DIST and SING_BOX_BIN to verify generated artifacts")
	}
	cfg, err := config.Load("../../config/sources.json")
	if err != nil {
		t.Fatal(err)
	}
	checksums := readAndVerifyChecksums(t, root)
	for archive, directory := range map[string]string{"country-csv.zip": "countries", "source-csv.zip": "source-csv", "sing-box-rulesets.zip": "sing-box"} {
		verifyArchive(t, root, archive, directory, checksums)
	}
	files, err := filepath.Glob(filepath.Join(root, "countries", "*.csv"))
	if err != nil {
		t.Fatal(err)
	}
	expected := make(map[string]*netipx.IPSet)
	var all []netip.Prefix
	manual := new(netipx.IPSetBuilder)
	for _, path := range files {
		cc := strings.ToUpper(strings.TrimSuffix(filepath.Base(path), ".csv"))
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		r := csv.NewReader(f)
		header, err := r.Read()
		if err != nil {
			t.Fatal(err)
		}
		if strings.Join(header, ",") != "network,country,confidence_pct,winner_votes,total_votes,assignment_method" {
			t.Fatalf("unexpected CSV header: %v", header)
		}
		builder := new(netipx.IPSetBuilder)
		for {
			row, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			p, err := netip.ParsePrefix(row[0])
			if err != nil {
				t.Fatal(err)
			}
			if row[1] != cc {
				t.Fatalf("wrong country in %s", path)
			}
			if row[5] == "manual" {
				if row[2] != "" || row[3] != "0" || row[4] != "0" {
					t.Fatalf("invented manual confidence: %v", row)
				}
				manual.AddPrefix(p)
			} else if row[5] != "geolocation" && row[5] != "fallback" {
				t.Fatalf("unknown method: %v", row)
			}
			builder.AddPrefix(p)
			all = append(all, p)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		expected[cc], err = builder.IPSet()
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(expected) == 0 {
		t.Fatal("no countries")
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Addr() == all[j].Addr() {
			return all[i].Bits() < all[j].Bits()
		}
		return all[i].Addr().Less(all[j].Addr())
	})
	for i := 1; i < len(all); i++ {
		if all[i-1].Addr().Is4() == all[i].Addr().Is4() && iprange.PrefixLast(all[i-1]).Compare(all[i].Addr()) >= 0 {
			t.Fatalf("overlapping country CIDRs: %s and %s", all[i-1], all[i])
		}
	}
	forced, err := manual.IPSet()
	if err != nil {
		t.Fatal(err)
	}
	for _, override := range cfg.Overrides {
		p := netip.MustParsePrefix(override.CIDR)
		if expected[override.Country] == nil || !expected[override.Country].ContainsPrefix(p) || !forced.ContainsPrefix(p) {
			t.Fatalf("override not covered: %+v", override)
		}
	}
	var summary model.RunSummary
	readJSON(t, filepath.Join(root, "run-summary.json"), &summary)
	if int64(len(all)) != summary.OutputBlocks {
		t.Fatalf("summary block mismatch: %d vs %d", len(all), summary.OutputBlocks)
	}
	if len(summary.Overrides) != len(cfg.Overrides) {
		t.Fatal("missing override report")
	}
	for _, o := range summary.Overrides {
		if o.Applied.IPv4Addresses != "256" {
			t.Fatalf("incorrect applied /24 coverage: %+v", o)
		}
	}

	data, err := os.ReadFile(filepath.Join(root, "geoip.dat"))
	if err != nil {
		t.Fatal(err)
	}
	var geo v2ray.GeoIPList
	if err := proto.Unmarshal(data, &geo); err != nil {
		t.Fatal(err)
	}
	dat := make(map[string]*netipx.IPSet)
	for _, entry := range geo.Entry {
		builder := new(netipx.IPSetBuilder)
		for _, cidr := range entry.Cidr {
			addr, ok := netip.AddrFromSlice(cidr.Ip)
			if !ok {
				t.Fatal("invalid DAT IP")
			}
			builder.AddPrefix(netip.PrefixFrom(addr, int(cidr.Prefix)))
		}
		cc := strings.ToUpper(entry.CountryCode)
		dat[cc], err = builder.IPSet()
		if err != nil {
			t.Fatal(err)
		}
	}
	compareCountries(t, "DAT", expected, dat)
	reader, err := maxminddb.Open(filepath.Join(root, "country.mmdb"))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if err := reader.Verify(); err != nil {
		t.Fatal(err)
	}
	builders := make(map[string]*netipx.IPSetBuilder)
	for network := range reader.Networks() {
		if err := network.Err(); err != nil {
			t.Fatal(err)
		}
		var record struct {
			Country struct {
				Code string `maxminddb:"iso_code"`
			} `maxminddb:"country"`
		}
		if err := network.Decode(&record); err != nil {
			t.Fatal(err)
		}
		if builders[record.Country.Code] == nil {
			builders[record.Country.Code] = new(netipx.IPSetBuilder)
		}
		builders[record.Country.Code].AddPrefix(network.Prefix())
	}
	mmdb := make(map[string]*netipx.IPSet)
	for cc, builder := range builders {
		mmdb[cc], err = builder.IPSet()
		if err != nil {
			t.Fatal(err)
		}
	}
	compareCountries(t, "MMDB", expected, mmdb)
	for _, override := range cfg.Overrides {
		p := netip.MustParsePrefix(override.CIDR)
		for ip, count := p.Addr(), 0; ip.IsValid() && p.Contains(ip) && count < 1024; ip, count = ip.Next(), count+1 {
			var record struct {
				Country struct {
					Code string `maxminddb:"iso_code"`
				} `maxminddb:"country"`
			}
			if err := reader.Lookup(ip).Decode(&record); err != nil {
				t.Fatal(err)
			}
			if record.Country.Code != override.Country {
				t.Fatalf("MMDB override %s: %s", ip, record.Country.Code)
			}
		}
	}

	decompiled := t.TempDir()
	for cc, want := range expected {
		name := "geoip-" + strings.ToLower(cc)
		if got := ruleSetIPs(t, filepath.Join(root, "sing-box", name+".json")); !want.Equal(got) {
			t.Fatalf("JSON coverage differs for %s", cc)
		}
		path := filepath.Join(decompiled, name+".json")
		if out, err := exec.Command(binary, "rule-set", "decompile", filepath.Join(root, "sing-box", name+".srs"), "-o", path).CombinedOutput(); err != nil {
			t.Fatalf("decompile %s: %v: %s", cc, err, out)
		}
		if got := ruleSetIPs(t, path); !want.Equal(got) {
			t.Fatalf("SRS coverage differs for %s", cc)
		}
	}
	t.Logf("Verified %d non-overlapping CIDRs, %d countries in CSV/DAT/MMDB/JSON/SRS, %d overrides, all checksums and archive entries", len(all), len(expected), len(cfg.Overrides))
}

func readJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, value); err != nil {
		t.Fatal(err)
	}
}

func ruleSetIPs(t *testing.T, path string) *netipx.IPSet {
	t.Helper()
	var rules RuleSet
	readJSON(t, path, &rules)
	builder := new(netipx.IPSetBuilder)
	for _, rule := range rules.Rules {
		for _, raw := range rule.IPCIDR {
			p, err := netip.ParsePrefix(raw)
			if err != nil {
				t.Fatal(err)
			}
			builder.AddPrefix(p)
		}
	}
	set, err := builder.IPSet()
	if err != nil {
		t.Fatal(err)
	}
	return set
}

func compareCountries(t *testing.T, format string, expected, actual map[string]*netipx.IPSet) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Fatalf("%s country count: %d want %d", format, len(actual), len(expected))
	}
	for cc, want := range expected {
		if actual[cc] == nil || !actual[cc].Equal(want) {
			t.Fatalf("%s coverage differs for %s", format, cc)
		}
	}
}

func hashStream(t *testing.T, r io.Reader) string {
	t.Helper()
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func readAndVerifyChecksums(t *testing.T, root string) map[string]string {
	t.Helper()
	f, err := os.Open(filepath.Join(root, "SHA256SUMS"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	checksums := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "  ", 2)
		if len(parts) != 2 || len(parts[0]) != 64 {
			t.Fatalf("bad checksum line %q", scanner.Text())
		}
		name := parts[1]
		if !filepath.IsLocal(filepath.FromSlash(name)) {
			t.Fatalf("unsafe checksum path %s", name)
		}
		if _, ok := checksums[name]; ok {
			t.Fatalf("duplicate checksum %s", name)
		}
		file, err := os.Open(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		got := hashStream(t, file)
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
		if got != parts[0] {
			t.Fatalf("checksum mismatch for %s", name)
		}
		checksums[name] = got
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(root, path)
			name := filepath.ToSlash(rel)
			if name != "SHA256SUMS" && checksums[name] == "" {
				t.Errorf("missing checksum for %s", name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return checksums
}

func verifyArchive(t *testing.T, root, archive, directory string, checksums map[string]string) {
	t.Helper()
	r, err := zip.OpenReader(filepath.Join(root, archive))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	seen := make(map[string]bool)
	for _, file := range r.File {
		if file.FileInfo().IsDir() {
			continue
		}
		name := directory + "/" + file.Name
		if seen[name] || checksums[name] == "" {
			t.Fatalf("unexpected/duplicate entry %s in %s", name, archive)
		}
		stream, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		got := hashStream(t, stream)
		if err := stream.Close(); err != nil {
			t.Fatal(err)
		}
		if got != checksums[name] {
			t.Fatalf("archive content differs for %s", name)
		}
		seen[name] = true
	}
	for name := range checksums {
		if strings.HasPrefix(name, directory+"/") && !seen[name] {
			t.Fatalf("missing archive entry %s", name)
		}
	}
}
