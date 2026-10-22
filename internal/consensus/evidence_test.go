package consensus

import (
	"net/netip"
	"testing"

	"github.com/txyyddss/ip2country/internal/model"
)

func fallback(cidr, country, group string, order int) model.Record {
	r := rec(cidr, country, group, order)
	r.Evidence = model.Fallback
	return r
}

func lookupBlock(t *testing.T, blocks []Block, ip string) *Block {
	t.Helper()
	a := netip.MustParseAddr(ip)
	var found *Block
	for i := range blocks {
		if blocks[i].Prefix.Contains(a) {
			if found != nil {
				t.Fatalf("overlapping output for %s", ip)
			}
			found = &blocks[i]
		}
	}
	return found
}

func TestGeolocationBeatsFallbackMajorityAndRestoresFallback(t *testing.T) {
	records := []model.Record{
		fallback("8.8.8.0/24", "CN", "registry-a", 0),
		fallback("8.8.8.0/24", "CN", "registry-b", 1),
		fallback("8.8.8.0/24", "US", "registry-c", 2),
		rec("8.8.8.64/26", "HK", "location", 3),
	}
	blocks := resolveBoth(t, records, nil, nil)
	for _, ip := range []string{"8.8.8.0", "8.8.8.63", "8.8.8.128", "8.8.8.255"} {
		b := lookupBlock(t, blocks, ip)
		if b == nil || b.Country != "CN" || b.Method != model.Fallback || b.Confidence < 66.66 || b.Confidence > 66.67 {
			t.Fatalf("fallback %s: %+v", ip, b)
		}
	}
	b := lookupBlock(t, blocks, "8.8.8.64")
	if b == nil || b.Country != "HK" || b.Method != model.Geolocation || b.Confidence != 100 || b.WinnerVotes.String() != "64" || b.TotalVotes.String() != "64" {
		t.Fatalf("location evidence: %+v", b)
	}
}

func TestEvidencePriorityWithinProviderAndSameCountryTransitions(t *testing.T) {
	records := []model.Record{
		fallback("8.8.8.0/24", "HK", "shared", 0),
		rec("8.8.8.0/25", "HK", "shared", 1),
		fallback("8.8.8.0/26", "CN", "shared", 0),
		rec("8.8.8.0/25", "HK", "shared", 2),
		rec("8.8.8.0/24", "US", "independent", 3),
	}
	blocks := resolveBoth(t, records, map[string]int{"shared": 0, "independent": 1}, nil)
	for _, ip := range []string{"8.8.8.1", "8.8.8.70"} {
		b := lookupBlock(t, blocks, ip)
		if b == nil || b.Country != "HK" || b.Confidence != 50 || b.Method != model.Geolocation {
			t.Fatalf("group duplicated or weaker prefix won: %+v", b)
		}
	}
	if b := lookupBlock(t, blocks, "8.8.8.200"); b == nil || b.Country != "US" || b.Confidence != 100 {
		t.Fatalf("same-country class change was not applied: %+v", b)
	}
}

func TestManualOverridesNestedUncoveredAndExcluded(t *testing.T) {
	overrides := []model.Override{
		{CIDR: "8.8.8.0/24", Country: "HK"},
		{CIDR: "8.8.8.128/25", Country: "JP"},
		{CIDR: "9.9.9.0/24", Country: "HK"},
		{CIDR: "10.0.0.0/8", Country: "HK"},
		{CIDR: "2001:4860::/32", Country: "HK"},
		{CIDR: "2001:4860::1/128", Country: "JP"},
	}
	blocks := resolveBoth(t, []model.Record{rec("8.8.8.0/23", "US", "provider", 0)}, nil, BuiltinSpecialUse(), overrides...)
	for ip, country := range map[string]string{"8.8.8.0": "HK", "8.8.8.127": "HK", "8.8.8.128": "JP", "8.8.8.255": "JP", "9.9.9.255": "HK", "2001:4860::": "HK", "2001:4860::1": "JP"} {
		b := lookupBlock(t, blocks, ip)
		if b == nil || b.Country != country || b.Method != model.Manual || b.Confidence != 0 || b.WinnerVotes.Sign() != 0 || b.TotalVotes.Sign() != 0 {
			t.Fatalf("manual %s: %+v", ip, b)
		}
	}
	for _, ip := range []string{"8.8.9.0", "8.8.9.255"} {
		if b := lookupBlock(t, blocks, ip); b == nil || b.Country != "US" || b.Method != model.Geolocation {
			t.Fatalf("neighbor changed: %+v", b)
		}
	}
	for _, ip := range []string{"9.9.8.255", "9.9.10.0", "10.0.0.1"} {
		if b := lookupBlock(t, blocks, ip); b != nil {
			t.Fatalf("unexpected coverage %s: %+v", ip, b)
		}
	}
}

func TestMethodAndOverrideBoundariesRemainDistinct(t *testing.T) {
	blocks := resolveBoth(t, []model.Record{rec("8.8.8.0/25", "HK", "geo", 0), fallback("8.8.8.128/25", "HK", "fallback", 1)}, nil, nil,
		model.Override{CIDR: "8.8.8.192/26", Country: "HK"})
	if len(blocks) != 3 || blocks[0].Method != model.Geolocation || blocks[1].Method != model.Fallback || blocks[2].Method != model.Manual {
		t.Fatalf("methods merged: %+v", blocks)
	}
}
