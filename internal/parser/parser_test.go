package parser

import (
	"archive/zip"
	"encoding/binary"
	"net/netip"
	"os"
	"path/filepath"
	"testing"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/txyyddss/ip2country/internal/config"
	"github.com/txyyddss/ip2country/internal/model"
	"github.com/v2fly/geoip/plugin/v2ray"
	"go4.org/netipx"
	"google.golang.org/protobuf/proto"
)

func TestMMDBLocationFieldsPrecedeRegistrationAndRetainEvidence(t *testing.T) {
	db, err := mmdbwriter.New(mmdbwriter.Options{DatabaseType: "Test-Country", IPVersion: 6, DisableIPv4Aliasing: true})
	if err != nil {
		t.Fatal(err)
	}
	values := []mmdbtype.Map{
		{"country": mmdbtype.Map{"iso_code": mmdbtype.String("HK")}, "registered_country": mmdbtype.Map{"iso_code": mmdbtype.String("CN")}},
		{"country_code": mmdbtype.String("HK"), "registered_country": mmdbtype.Map{"iso_code": mmdbtype.String("CN")}},
		{"country_iso_code": mmdbtype.String("HK"), "registered_country": mmdbtype.Map{"iso_code": mmdbtype.String("CN")}},
		{"registered_country": mmdbtype.Map{"iso_code": mmdbtype.String("CN")}},
	}
	prefixes := []string{"8.8.8.0/24", "8.8.9.0/24", "8.8.10.0/24", "8.8.11.0/24"}
	for i, value := range values {
		if err := db.Insert(netipx.PrefixIPNet(netip.MustParsePrefix(prefixes[i])), value); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "evidence.mmdb")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_, writeErr := db.WriteTo(f)
	closeErr := f.Close()
	if writeErr != nil {
		t.Fatal(writeErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	for _, class := range []string{"geolocation", "fallback"} {
		source := testSource("mmdb")
		source.EvidenceClass = class
		parsed, err := ParseFile(source, path, 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(parsed.Records) != 4 {
			t.Fatalf("wrong record count: %+v", parsed)
		}
		for _, r := range parsed.Records {
			registration := r.Prefix.String() == prefixes[3]
			wantCountry := "HK"
			wantClass := model.Geolocation
			if registration {
				wantCountry = "CN"
			}
			if registration || class == "fallback" {
				wantClass = model.Fallback
			}
			if r.Country != wantCountry || r.Evidence != wantClass {
				t.Fatalf("wrong MMDB interpretation: %+v", r)
			}
		}
	}
}

func testSource(format string) config.Source {
	return config.Source{ID: "fixture", Name: "fixture", VoteGroup: "fixture", Role: "country", Format: format, URLs: []string{"fixture"}}
}

func TestHexadecimalIPv4Ranges(t *testing.T) {
	s := testSource("tsv")
	s.Family = "ipv4"
	s.IPEncoding = "hex"
	r, err := ParseFile(s, writeTemp(t, "hex.tsv", []byte("1000000\t10000ff\tAU\nffffffff\tffffffff\tUS\n")), 0)
	if err != nil {
		t.Fatal(err)
	}
	if r.Rejected != 0 || len(r.Records) != 2 || r.Records[0].Prefix.String() != "1.0.0.0/24" || r.Records[1].Prefix.String() != "255.255.255.255/32" {
		t.Fatalf("unexpected hexadecimal ranges: %+v", r)
	}
}

func TestCountryCodePrecedesCountryName(t *testing.T) {
	for _, header := range []string{"country_code,country", "country,country_code"} {
		row := "US,unrecognized country name"
		if header == "country,country_code" {
			row = "unrecognized country name,US"
		}
		r, err := ParseFile(testSource("csv"), writeTemp(t, "ipapi.csv", []byte("start_ip,end_ip,"+header+"\n8.8.8.0,8.8.8.255,"+row+"\n")), 0)
		if err != nil || r.Rejected != 0 || r.NoCountry != 0 || len(r.Records) != 1 || r.Records[0].Country != "US" {
			t.Fatalf("unexpected ISO column choice: %+v, %v", r, err)
		}
	}
}

func TestMalformedRangeCannotBecomeHostVote(t *testing.T) {
	r, err := ParseFile(testSource("csv"), writeTemp(t, "bad.csv", []byte("start_ip,end_ip,country_code\n8.8.8.0,8.8.8.255,US\n8.8.9.0,broken,JP\n8.8.10.255,8.8.10.0,JP\n")), 0)
	if err != nil {
		t.Fatal(err)
	}
	if r.Rejected != 2 || len(r.Records) != 1 {
		t.Fatalf("malformed endpoint silently accepted: %+v", r)
	}
}

func TestNoCountryIsSeparateFromMalformedRows(t *testing.T) {
	r, err := ParseFile(testSource("csv"), writeTemp(t, "no-country.csv", []byte("network,country_code\n8.8.8.0/24,US\n8.8.9.0/24,EU\n8.8.10.0/24,\n8.8.11.0/24,XK\n8.8.12.0/24,UK\n")), 0)
	if err != nil {
		t.Fatal(err)
	}
	if r.Rejected != 0 || r.NoCountry != 2 || len(r.Records) != 3 || r.Records[1].Country != "XK" || r.Records[2].Country != "GB" {
		t.Fatalf("unexpected no-country handling: %+v", r)
	}
}

func TestExplicitDecimalIPv6Encoding(t *testing.T) {
	s := testSource("tsv")
	s.Family = "ipv6"
	s.IPEncoding = "decimal"
	value := "10000000000000000000000000000000"
	r, err := ParseFile(s, writeTemp(t, "decimal.tsv", []byte(value+"\t"+value+"\tUS\n")), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Records) != 1 || r.Records[0].Prefix.String() != "0:7e:37be:2022:c091:4b26:8000:0/128" {
		t.Fatalf("ambiguous decimal IPv6 misread: %+v", r)
	}
}

func TestMMDBMetadataDoesNotCastCountryVote(t *testing.T) {
	db, err := mmdbwriter.New(mmdbwriter.Options{DatabaseType: "Test-Country", IPVersion: 6})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		prefix string
		data   mmdbtype.Map
	}{
		{"8.8.8.0/24", mmdbtype.Map{"country": mmdbtype.Map{"iso_code": mmdbtype.String("US")}}},
		{"8.8.9.0/24", mmdbtype.Map{"rir": mmdbtype.String("arin")}},
	} {
		if err := db.Insert(netipx.PrefixIPNet(netip.MustParsePrefix(row.prefix)), row.data); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "metadata.mmdb")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.WriteTo(f)
	closeErr := f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	r, err := ParseFile(testSource("mmdb"), path, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Records) != 1 || r.Rejected != 0 || r.NoCountry != 1 {
		t.Fatalf("unexpected metadata handling: %+v", r)
	}
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
