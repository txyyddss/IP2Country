package report

import (
	"net/netip"
	"strings"
	"testing"

	"github.com/txyyddss/ip2country/internal/consensus"
	"github.com/txyyddss/ip2country/internal/model"
)

func TestCoverageSumsFractionsBeforeRoundingREADME(t *testing.T) {
	var blocks []consensus.Block
	for _, prefix := range []string{"8.8.8.0/25", "8.8.9.0/26", "2001:4860::/65", "2001:4860:0:1::/66"} {
		blocks = append(blocks, consensus.Block{Prefix: netip.MustParsePrefix(prefix), Country: "US", Confidence: 75})
	}
	summary := BuildSummary([]model.SourceStatus{{Name: "Fixture", NoCountryRows: 3}}, blocks, 4)
	c := summary.Countries[0]
	if c.IPv4Cells != "0.75" || c.IPv6Cells != "0.75" {
		t.Fatalf("fractional totals lost: %+v", c)
	}
	markdown := render(summary)
	if !strings.Contains(markdown, "| US | 2 | 1 | 2 | 1 | 75.00% |") || !strings.Contains(markdown, "No country") {
		t.Fatalf("unexpected README units: %s", markdown)
	}
}

func TestCoverageRetainsTinyIPv6AndLargeIntegerPrecision(t *testing.T) {
	blocks := []consensus.Block{{Prefix: netip.MustParsePrefix("::/1"), Country: "US", Confidence: 100}, {Prefix: netip.MustParsePrefix("8000::1/128"), Country: "US", Confidence: 100}}
	summary := BuildSummary(nil, blocks, 2)
	c := summary.Countries[0]
	if c.IPv6Cells != "9223372036854775808.0000000000000000000542101086242752217003726400434970855712890625" {
		t.Fatalf("lost exact IPv6 coverage: %s", c.IPv6Cells)
	}
	if !strings.Contains(render(summary), "| US | 0 | 0 | 2 | 9223372036854775808 |") {
		t.Fatal("README units rounded through floating point")
	}
}

func TestManualCoverageAndConfidenceExcludeForcedAssignments(t *testing.T) {
	overrides := []model.Override{{CIDR: "8.8.8.0/24", Country: "HK", Reason: "confirmed | range"}, {CIDR: "8.8.8.128/25", Country: "JP"}, {CIDR: "10.0.0.0/8", Country: "HK"}}
	records := []model.Record{{Prefix: netip.MustParsePrefix("8.8.9.0/24"), Country: "HK", VoteGroup: "first"}, {Prefix: netip.MustParsePrefix("8.8.9.0/24"), Country: "JP", VoteGroup: "second", Order: 1}}
	blocks, err := consensus.Resolve(records, map[string]int{"first": 0, "second": 1}, consensus.BuiltinSpecialUse(), overrides...)
	if err != nil {
		t.Fatal(err)
	}
	summary := BuildSummary(nil, blocks, 2, overrides...)
	if summary.Overrides[0].Applied.IPv4Addresses != "128" || summary.Overrides[1].Applied.IPv4Addresses != "128" || summary.Overrides[2].Applied.IPv4Addresses != "0" {
		t.Fatalf("wrong effective override coverage: %+v", summary.Overrides)
	}
	if summary.EvidenceCoverage["manual"].IPv4Addresses != "256" || summary.EvidenceCoverage["geolocation"].IPv4Addresses != "256" {
		t.Fatalf("wrong method coverage: %+v", summary.EvidenceCoverage)
	}
	for _, country := range summary.Countries {
		if country.Country == "HK" && country.AverageConf != "50.00%" {
			t.Fatalf("manual altered confidence: %+v", country)
		}
		if country.Country == "JP" && country.AverageConf != "" {
			t.Fatalf("manual-only country has confidence: %+v", country)
		}
	}
	if !strings.Contains(render(summary), `confirmed \| range`) {
		t.Fatal("override reason not escaped")
	}
}

func TestNestedSameCountryOverridesKeepSeparateAppliedCoverage(t *testing.T) {
	overrides := []model.Override{{CIDR: "8.8.8.0/24", Country: "HK"}, {CIDR: "8.8.8.128/25", Country: "HK"}}
	blocks, err := consensus.Resolve([]model.Record{{Prefix: netip.MustParsePrefix("8.8.8.0/24"), Country: "US", VoteGroup: "fixture"}}, nil, nil, overrides...)
	if err != nil {
		t.Fatal(err)
	}
	summary := BuildSummary(nil, blocks, 1, overrides...)
	if len(blocks) != 2 || summary.Overrides[0].Applied.IPv4Addresses != "128" || summary.Overrides[1].Applied.IPv4Addresses != "128" {
		t.Fatalf("same-country overrides merged: %+v", summary.Overrides)
	}
}
