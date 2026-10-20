package parser

import (
	"archive/zip"
	"encoding/binary"
	"net/netip"
	"os"
	"path/filepath"
	"testing"

	"github.com/txyyddss/ip2country/internal/config"
	"github.com/v2fly/geoip/plugin/v2ray"
	"google.golang.org/protobuf/proto"
)

func testSource(format string) config.Source {
	return config.Source{ID: "fixture", Name: "fixture", VoteGroup: "fixture", Role: "country", Format: format, URLs: []string{"fixture"}}
}

func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseCSVAndTSV(t *testing.T) {
	for _, tc := range []struct{ format, content string }{{"csv", "network,country_code\n198.51.100.0/24,US\n"}, {"tsv", "start_ip\tend_ip\tcountry_code\n198.51.100.0\t198.51.100.255\tUS\n"}} {
		t.Run(tc.format, func(t *testing.T) {
			r, err := ParseFile(testSource(tc.format), writeTemp(t, "fixture.txt", []byte(tc.content)), 0)
			if err != nil {
				t.Fatal(err)
			}
			if len(r.Records) != 1 || r.Records[0].Prefix.String() != "198.51.100.0/24" || r.Records[0].Country != "US" {
				t.Fatalf("unexpected parse: %+v", r)
			}
		})
	}
}

func TestParseIPv6HexRangeTSV(t *testing.T) {
	s := testSource("tsv")
	s.Family = "ipv6"
	content := "start_ip\tend_ip\tcountry_code\n20000000000000000000000000000000\t2000ffffffffffffffffffffffffffff\tCH\n"
	r, err := ParseFile(s, writeTemp(t, "ipv6.tsv", []byte(content)), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Records) != 1 || r.Records[0].Prefix.String() != "2000::/16" || r.Records[0].Country != "CH" {
		t.Fatalf("unexpected IPv6 TSV parse: %+v", r)
	}
}

func TestParseCSVArchive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("ranges.csv")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte("network,country\n198.51.100.0/24,US\n"))
	if err = zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	s := testSource("csv-zip")
	r, err := ParseFile(s, path, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Records) != 1 {
		t.Fatalf("got %d records", len(r.Records))
	}
}

func TestParseCustomTXT(t *testing.T) {
	s := testSource("txt")
	s.TextRegex = `^(?P<cidr>\S+)\s+(?P<country>[A-Z]{2})$`
	r, err := ParseFile(s, writeTemp(t, "ranges.txt", []byte("198.51.100.0/24 US\n")), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Records) != 1 || r.Records[0].Country != "US" {
		t.Fatalf("unexpected parse: %+v", r)
	}
}

func TestParseCountryJSON(t *testing.T) {
	r, err := ParseFile(testSource("json"), writeTemp(t, "ranges.json", []byte(`{"countries":{"US":[{"network":"198.51.100.0/24"}],"JP":[{"cidr":"203.0.113.0/24"}]}}`)), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Records) != 2 {
		t.Fatalf("unexpected records: %+v", r)
	}
}

func TestParseRepairableCountryRangesJSON(t *testing.T) {
	raw := `[{"country":"au","range":"1.0.0.0/24"}{"country":"us","range":"8.8.8.0/24"}]`
	r, err := ParseFile(testSource("json"), writeTemp(t, "ranges.json", []byte(raw)), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Records) != 2 {
		t.Fatalf("unexpected records: %+v", r)
	}
}

func TestParseIANAFilterDoesNotExcludeGlobalBlocks(t *testing.T) {
	s := testSource("json")
	s.Role = "filter"
	r, err := ParseFile(s, writeTemp(t, "iana.json", []byte(`{"blocks":[{"prefix":"192.0.2.0/24","isGloballyReachable":false},{"prefix":"8.8.8.0/24","isGloballyReachable":true}]}`)), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Filters) != 1 || r.Filters[0] != netip.MustParsePrefix("192.0.2.0/24") {
		t.Fatalf("unexpected IANA exclusions: %v", r.Filters)
	}
}

func TestParseIANAAllocationCountryJSON(t *testing.T) {
	s := testSource("json")
	r, err := ParseFile(s, writeTemp(t, "iana-country.json", []byte(`[{"Country":"AD","IP":"2.152.216.0","PrefixLength":"21","Version":"ipv4"}]`)), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Records) != 1 || r.Records[0].Country != "AD" || r.Records[0].Prefix.String() != "2.152.216.0/21" {
		t.Fatalf("unexpected IANA country record: %+v", r)
	}
}

func TestParseV2RayDat(t *testing.T) {
	message := &v2ray.GeoIPList{Entry: []*v2ray.GeoIP{{CountryCode: "us", Cidr: []*v2ray.CIDR{{Ip: []byte{198, 51, 100, 0}, Prefix: 24}}}}}
	raw, err := proto.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	r, err := ParseFile(testSource("v2ray-dat"), writeTemp(t, "geoip.dat", raw), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Records) != 1 || r.Records[0].Prefix.String() != "198.51.100.0/24" {
		t.Fatalf("unexpected parse: %+v", r)
	}
}

func TestParseSxGeoCountryDat(t *testing.T) {
	data := make([]byte, 40+9*4+4)
	copy(data[:3], []byte("SxG"))
	data[3] = 22
	data[8] = 1
	data[9] = 0
	data[10] = 9
	binary.BigEndian.PutUint16(data[13:15], 1)
	binary.BigEndian.PutUint32(data[15:19], 1)
	data[19] = 1
	dbStart := 40 + 9*4
	data[dbStart+2] = 0 // start address lower 24 bits: 8.0.0.0
	data[dbStart+3] = 3 // SxGeo country catalog ID 3 -> AD
	zpath := filepath.Join(t.TempDir(), "sypex.zip")
	f, err := os.Create(zpath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	member, err := zw.Create("SxGeo.dat")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := member.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	r, err := ParseFile(testSource("sxgeo-zip"), zpath, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Records) == 0 {
		t.Fatal("SYPEX country records were not decoded")
	}
	for _, record := range r.Records {
		if record.Country != "AD" {
			t.Fatalf("unexpected SYPEX country: %+v", record)
		}
	}
}

func TestParseIP2RegionXDB(t *testing.T) {
	const header = 256
	const vector = 65536 * 8
	const indexOffset = header + vector
	region := []byte("Country|Province|City|ISP|US")
	data := make([]byte, indexOffset+14+len(region))
	binary.LittleEndian.PutUint16(data[0:], 3)
	binary.LittleEndian.PutUint16(data[2:], 1)
	binary.LittleEndian.PutUint32(data[8:], indexOffset)
	binary.LittleEndian.PutUint32(data[12:], indexOffset)
	binary.LittleEndian.PutUint16(data[16:], 4)
	binary.LittleEndian.PutUint16(data[18:], 4)
	for i := 0; i < 65536; i++ {
		binary.LittleEndian.PutUint32(data[header+i*8:], indexOffset)
		binary.LittleEndian.PutUint32(data[header+i*8+4:], indexOffset)
	}
	// IP2Region stores IPv4 segment endpoint bytes in little-endian order.
	binary.LittleEndian.PutUint32(data[indexOffset:], 0)
	binary.LittleEndian.PutUint32(data[indexOffset+4:], ^uint32(0))
	binary.LittleEndian.PutUint16(data[indexOffset+8:], uint16(len(region)))
	binary.LittleEndian.PutUint32(data[indexOffset+10:], uint32(indexOffset+14))
	copy(data[indexOffset+14:], region)
	r, err := ParseFile(testSource("ip2region-xdb"), writeTemp(t, "fixture.xdb", data), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Records) != 1 || r.Records[0].Prefix.String() != "0.0.0.0/0" || r.Records[0].Country != "US" {
		t.Fatalf("unexpected parse: %+v", r)
	}
	if _, err := netip.ParsePrefix(r.Records[0].Prefix.String()); err != nil {
		t.Fatal(err)
	}
}
