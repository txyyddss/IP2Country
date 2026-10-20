package consensus

import (
	"net/netip"
	"path/filepath"
	"testing"

	"github.com/txyyddss/ip2country/internal/model"
)

func rec(cidr, country, group string, order int) model.Record {
	return model.Record{Prefix: netip.MustParsePrefix(cidr), Country: country, VoteGroup: group, SourceID: group, Order: order, RowOrder: order}
}

func TestCompressedRecordSpoolRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "records.spool.gz")
	spool, err := NewRecordSpool(path, map[string]uint16{"": 0, "source": 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range []model.Record{rec("8.8.8.0/24", "US", "source", 0), rec("2001:4860::/64", "US", "source", 0)} {
		if err := spool.Write(record); err != nil {
			t.Fatal(err)
		}
	}
	if err := spool.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveSpool(path, []string{"", "source"}, map[string]int{"source": 0}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d blocks: %+v", len(got), got)
	}
}

func TestExternalConsensusSortMatchesPartialCoverageAndMerging(t *testing.T) {
	records := []model.Record{
		rec("198.51.100.0/23", "US", "first", 0),
		rec("198.51.101.0/24", "JP", "second", 1),
	}
	path := filepath.Join(t.TempDir(), "consensus.spool.gz")
	spool, err := NewRecordSpool(path, map[string]uint16{"": 0, "first": 1, "second": 2})
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		if err := spool.Write(record); err != nil {
			t.Fatal(err)
		}
	}
	if err := spool.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveSpool(path, []string{"", "first", "second"}, map[string]int{"first": 0, "second": 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Prefix.String() != "198.51.100.0/23" || got[0].Country != "US" || got[0].Confidence != 75 {
		t.Fatalf("unexpected external result: %+v", got)
	}
}

func TestExternalConsensusResolvesPerSourceCoverage(t *testing.T) {
	records := []model.Record{
		rec("198.51.100.0/25", "US", "first", 0),
		rec("198.51.100.128/26", "JP", "first", 0),
		rec("198.51.100.0/24", "JP", "second", 1),
	}
	records[0].RowOrder = 1
	records[1].RowOrder = 2
	path := filepath.Join(t.TempDir(), "partial.spool.gz")
	spool, err := NewRecordSpool(path, map[string]uint16{"": 0, "first": 1, "second": 2})
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		if err := spool.Write(record); err != nil {
			t.Fatal(err)
		}
	}
	if err := spool.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveSpool(path, []string{"", "first", "second"}, map[string]int{"first": 0, "second": 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Country != "US" || got[0].Confidence != 50 {
		t.Fatalf("unexpected coverage result: %+v", got)
	}
}

func TestExternalConsensusHandlesIPv4MaximumAndIPv6ZeroRoute(t *testing.T) {
	for _, cidr := range []string{"255.255.255.0/24", "::/0"} {
		t.Run(cidr, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "route.spool.gz")
			spool, err := NewRecordSpool(path, map[string]uint16{"": 0, "source": 1})
			if err != nil {
				t.Fatal(err)
			}
			if err := spool.Write(rec(cidr, "US", "source", 0)); err != nil {
				t.Fatal(err)
			}
			if err := spool.Close(); err != nil {
				t.Fatal(err)
			}
			got, err := ResolveSpool(path, []string{"", "source"}, map[string]int{"source": 0}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 1 || got[0].Prefix.String() != cidr {
				t.Fatalf("unexpected result: %+v", got)
			}
		})
	}
}

func TestTieUsesConfiguredSourceOrder(t *testing.T) {
	got, err := Resolve([]model.Record{rec("198.51.100.0/24", "US", "first", 0), rec("198.51.100.0/24", "JP", "second", 1)}, map[string]int{"first": 0, "second": 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Country != "US" || got[0].Confidence != 50 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestOneVotePerSourceAndLargestCoverage(t *testing.T) {
	records := []model.Record{
		rec("198.51.100.0/25", "US", "first", 0),
		rec("198.51.100.128/26", "JP", "first", 0),
		rec("198.51.100.0/24", "JP", "second", 1),
		rec("198.51.100.0/24", "US", "second", 1), // duplicate source coverage cannot add a second vote
	}
	records[3].RowOrder = 2
	got, err := Resolve(records, map[string]int{"first": 0, "second": 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Country != "US" || got[0].Confidence != 50 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestAdjacentSameCountryMergesAndAveragesConfidence(t *testing.T) {
	records := []model.Record{
		rec("198.51.100.0/23", "US", "first", 0),
		rec("198.51.101.0/24", "JP", "second", 1),
	}
	got, err := Resolve(records, map[string]int{"first": 0, "second": 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Prefix.String() != "198.51.100.0/23" || got[0].Country != "US" || got[0].Confidence != 75 {
		t.Fatalf("unexpected merged result: %+v", got)
	}
	if got[0].WinnerVotes.String() != "2" || got[0].TotalVotes.String() != "3" {
		t.Fatalf("unexpected vote totals: %s/%s", got[0].WinnerVotes, got[0].TotalVotes)
	}
}

func TestHugeIPv6RangeRemainsCompressed(t *testing.T) {
	got, err := Resolve([]model.Record{rec("2001:4860::/32", "US", "first", 0)}, map[string]int{"first": 0}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Prefix.String() != "2001:4860::/32" {
		t.Fatalf("unexpected IPv6 result: %+v", got)
	}
}

func TestSpecialUseExclusionsAreSubtractedPrecisely(t *testing.T) {
	got, err := Resolve([]model.Record{rec("192.0.2.0/23", "US", "first", 0)}, map[string]int{"first": 0}, []netip.Prefix{netip.MustParsePrefix("192.0.2.0/24")})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Prefix.String() != "192.0.3.0/24" {
		t.Fatalf("unexpected filtered result: %+v", got)
	}
}
