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
