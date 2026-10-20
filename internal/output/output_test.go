package output

import (
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oschwald/maxminddb-golang/v2"
	"github.com/txyyddss/ip2country/internal/config"
	"github.com/txyyddss/ip2country/internal/consensus"
	"github.com/txyyddss/ip2country/internal/parser"
	"github.com/v2fly/geoip/plugin/v2ray"
	"google.golang.org/protobuf/proto"
)

func TestMMDBOutputCanBeReadBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "country.mmdb")
	block := consensus.Block{Prefix: netip.MustParsePrefix("8.8.8.0/24"), Country: "US", Confidence: 90}
	if err := writeMMDB(path, []consensus.Block{block}); err != nil {
		t.Fatal(err)
	}
	r, err := parser.ParseFile(config.Source{ID: "roundtrip", Name: "roundtrip", VoteGroup: "roundtrip", Format: "mmdb", URLs: []string{"fixture"}}, path, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Records) != 1 || r.Records[0].Prefix.String() != "8.8.8.0/24" || r.Records[0].Country != "US" {
		t.Fatalf("unexpected MMDB readback: %+v", r)
	}
}

func TestV2RayDatOutputCanBeReadBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "geoip.dat")
	block := consensus.Block{Prefix: netip.MustParsePrefix("8.8.8.0/24"), Country: "US"}
	if err := writeGeoIPDat(path, map[string][]consensus.Block{"US": {block}}); err != nil {
		t.Fatal(err)
	}
	r, err := parser.ParseFile(config.Source{ID: "roundtrip", Name: "roundtrip", VoteGroup: "roundtrip", Format: "v2ray-dat", URLs: []string{"fixture"}}, path, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Records) != 1 || r.Records[0].Country != "US" || r.Records[0].Prefix.String() != "8.8.8.0/24" {
		t.Fatalf("unexpected protobuf readback: %+v", r)
	}
}

func TestPublishedFormatsCanBeRead(t *testing.T) {
	dist := os.Getenv("IP2COUNTRY_DIST")
	if dist == "" {
		t.Skip("set IP2COUNTRY_DIST to validate a generated release directory")
	}
	data, err := os.ReadFile(filepath.Join(dist, "geoip.dat"))
	if err != nil {
		t.Fatal(err)
	}
	var geo v2ray.GeoIPList
	if err := proto.Unmarshal(data, &geo); err != nil {
		t.Fatalf("decode geoip.dat: %v", err)
	}
	if len(geo.Entry) == 0 {
		t.Fatal("geoip.dat has no country entries")
	}
	reader, err := maxminddb.Open(filepath.Join(dist, "country.mmdb"))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	for _, ip := range []string{"8.8.8.8", "2606:4700:4700::1111"} {
		addr := netip.MustParseAddr(ip)
		var rec map[string]any
		if err := reader.Lookup(addr).Decode(&rec); err != nil {
			t.Fatalf("lookup %s: %v", ip, err)
		}
		country, ok := rec["country"].(map[string]any)
		if !ok || fmt.Sprint(country["iso_code"]) == "" {
			t.Errorf("MMDB has no country for %s: %v", ip, rec)
		}
	}
	entries, err := os.ReadDir(filepath.Join(dist, "sing-box"))
	if err != nil {
		t.Fatal(err)
	}
	jsonCount, srsCount := 0, 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".json") {
			jsonCount++
		}
		if strings.HasSuffix(entry.Name(), ".srs") {
			srsCount++
			info, err := entry.Info()
			if err != nil {
				t.Fatal(err)
			}
			if info.Size() == 0 {
				t.Errorf("empty rule set %s", entry.Name())
			}
		}
	}
	if jsonCount == 0 || jsonCount != srsCount {
		t.Fatalf("sing-box sources/binaries do not pair: json=%d srs=%d", jsonCount, srsCount)
	}
}
