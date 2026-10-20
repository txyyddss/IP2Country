package consensus

import (
	"math/rand/v2"
	"net/netip"
	"path/filepath"
	"reflect"
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

func TestExternalConsensusSplitsCompetingCountries(t *testing.T) {
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
	if len(got) != 2 || got[0].Prefix.String() != "198.51.100.0/25" || got[0].Country != "US" || got[0].Confidence != 50 || got[1].Prefix.String() != "198.51.100.128/25" || got[1].Country != "JP" || got[1].Confidence != 100 {
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

func TestOneVotePerGroupWithinEachFragment(t *testing.T) {
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
	if len(got) != 2 || got[0].Country != "US" || got[0].Confidence != 50 || got[1].Country != "JP" || got[1].Confidence != 100 {
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
	if got[0].WinnerVotes.String() != "512" || got[0].TotalVotes.String() != "768" {
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

func resolveBoth(t *testing.T, records []model.Record, ranks map[string]int, exclusions []netip.Prefix) []Block {
	t.Helper()
	got, err := Resolve(records, ranks, exclusions)
	if err != nil {
		t.Fatal(err)
	}
	groups, ids := []string{""}, map[string]uint16{}
	for _, r := range records {
		if _, ok := ids[r.VoteGroup]; !ok {
			ids[r.VoteGroup] = uint16(len(groups))
			groups = append(groups, r.VoteGroup)
		}
	}
	path := filepath.Join(t.TempDir(), "records.gz")
	spool, err := NewRecordSpool(path, ids)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range records {
		if err := spool.Write(r); err != nil {
			t.Fatal(err)
		}
	}
	if err := spool.Close(); err != nil {
		t.Fatal(err)
	}
	disk, err := ResolveSpool(path, groups, ranks, exclusions)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, disk) {
		t.Fatalf("memory/disk mismatch:\n%+v\n%+v", got, disk)
	}
	return got
}

func TestExactVotingBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name                string
		records             []model.Record
		prefixes, countries []string
	}{
		{"ipv4 minority", []model.Record{rec("8.8.8.0/25", "US", "a", 0), rec("8.8.8.128/26", "JP", "a", 0)}, []string{"8.8.8.0/25", "8.8.8.128/26"}, []string{"US", "JP"}},
		{"ipv6 minority", []model.Record{rec("2001:4860::/65", "US", "a", 0), rec("2001:4860:0:0:8000::/66", "JP", "a", 0)}, []string{"2001:4860::/65", "2001:4860:0:0:8000::/66"}, []string{"US", "JP"}},
		{"specific override", []model.Record{rec("8.8.8.0/24", "US", "a", 0), rec("8.8.8.128/25", "JP", "a", 1)}, []string{"8.8.8.0/25", "8.8.8.128/25"}, []string{"US", "JP"}},
		{"ipv4 last host", []model.Record{rec("255.255.255.255/32", "US", "a", 0)}, []string{"255.255.255.255/32"}, []string{"US"}},
		{"ipv6 last host", []model.Record{rec("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128", "US", "a", 0)}, []string{"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128"}, []string{"US"}},
		{"sparse single host", []model.Record{rec("8.8.8.7/32", "US", "a", 0)}, []string{"8.8.8.7/32"}, []string{"US"}},
		{"ipv6 zero route", []model.Record{rec("::/0", "US", "a", 0)}, []string{"::/0"}, []string{"US"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveBoth(t, tc.records, map[string]int{"a": 0}, nil)
			if len(got) != len(tc.prefixes) {
				t.Fatalf("unexpected blocks: %+v", got)
			}
			for i, b := range got {
				if b.Prefix.String() != tc.prefixes[i] || b.Country != tc.countries[i] || b.Confidence != 100 {
					t.Fatalf("unexpected block: %+v", b)
				}
			}
		})
	}
}

func TestVoteStatisticsAreSpecificToEachOutputPrefix(t *testing.T) {
	got := resolveBoth(t, []model.Record{rec("8.8.8.0/24", "US", "a", 0), rec("8.8.9.0/25", "US", "a", 0), rec("8.8.8.0/24", "JP", "b", 1)}, map[string]int{"a": 0, "b": 1}, nil)
	if len(got) != 2 || got[0].Confidence != 50 || got[0].WinnerVotes.String() != "256" || got[0].TotalVotes.String() != "512" || got[1].Confidence != 100 || got[1].WinnerVotes.String() != "128" || got[1].TotalVotes.String() != "128" {
		t.Fatalf("unexpected statistics: %+v", got)
	}
	filtered := resolveBoth(t, []model.Record{rec("8.8.8.0/23", "US", "a", 0), rec("8.8.9.0/24", "JP", "b", 1)}, map[string]int{"a": 0, "b": 1}, []netip.Prefix{netip.MustParsePrefix("8.8.9.0/24")})
	if len(filtered) != 1 || filtered[0].Confidence != 100 || filtered[0].TotalVotes.String() != "256" {
		t.Fatalf("excluded evidence leaked into confidence: %+v", filtered)
	}
}

// Compare compressed voting with independent per-address enumeration, including
// overlaps, duplicates, provider mirrors, uncovered holes, and both IP families.
func TestVotingMatchesPerAddressOracle(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 73))
	for _, ipv6 := range []bool{false, true} {
		for trial := 0; trial < 12; trial++ {
			var records []model.Record
			for row := 0; row < 80; row++ {
				bits := 24 + rng.IntN(9)
				address := netip.AddrFrom4([4]byte{8, 8, 8, byte(rng.IntN(256))})
				if ipv6 {
					bits += 96
					raw := netip.MustParseAddr("2001:4860::").As16()
					raw[15] = byte(rng.IntN(256))
					address = netip.AddrFrom16(raw)
				}
				group := []string{"a", "b", "c"}[rng.IntN(3)]
				r := rec(netip.PrefixFrom(address, bits).Masked().String(), []string{"US", "JP", "DE"}[rng.IntN(3)], group, rng.IntN(4))
				r.RowOrder = row
				records = append(records, r)
			}
			ranks := map[string]int{"a": 0, "b": 1, "c": 2}
			got := resolveBoth(t, records, ranks, nil)
			for host := 0; host < 256; host++ {
				address := netip.AddrFrom4([4]byte{8, 8, 8, byte(host)})
				if ipv6 {
					raw := netip.MustParseAddr("2001:4860::").As16()
					raw[15] = byte(host)
					address = netip.AddrFrom16(raw)
				}
				chosen := map[string]model.Record{}
				for _, r := range records {
					if r.Prefix.Contains(address) {
						old, ok := chosen[r.VoteGroup]
						if !ok || r.Prefix.Bits() > old.Prefix.Bits() || r.Prefix.Bits() == old.Prefix.Bits() && (r.Order < old.Order || r.Order == old.Order && r.RowOrder < old.RowOrder) {
							chosen[r.VoteGroup] = r
						}
					}
				}
				votes, first := map[string]int{}, map[string]int{}
				for g, r := range chosen {
					votes[r.Country]++
					old, ok := first[r.Country]
					if !ok || ranks[g] < old {
						first[r.Country] = ranks[g]
					}
				}
				winner := ""
				count, bestRank := 0, 999
				for country, n := range votes {
					if n > count || n == count && (first[country] < bestRank || first[country] == bestRank && country < winner) {
						winner, count, bestRank = country, n, first[country]
					}
				}
				actual := ""
				matches := 0
				for _, b := range got {
					if b.Prefix.Contains(address) {
						actual = b.Country
						matches++
					}
				}
				if actual != winner || matches > 1 {
					t.Fatalf("trial=%d address=%s got=%s want=%s matches=%d", trial, address, actual, winner, matches)
				}
			}
		}
	}
}

func TestExternalMergeAcrossChunkBoundaries(t *testing.T) {
	records := []model.Record{rec("8.8.8.0/24", "US", "a", 0), rec("8.8.8.128/25", "JP", "a", 0), rec("8.8.8.192/26", "DE", "b", 1), rec("2001:4860::/32", "US", "a", 0), rec("2001:4860::1/128", "JP", "a", 0)}
	ranks := map[string]int{"a": 0, "b": 1}
	want, err := Resolve(records, ranks, nil)
	if err != nil {
		t.Fatal(err)
	}
	sorter := newEventSorter(t.TempDir())
	// Put starts and ends in different sorted files, with several equal keys.
	for _, r := range records {
		group := uint16(1)
		if r.VoteGroup == "b" {
			group = 2
		}
		start, end := recordEvents(r, group)
		if err := sorter.add(end); err != nil {
			t.Fatal(err)
		}
		if err := sorter.flush(); err != nil {
			t.Fatal(err)
		}
		if err := sorter.add(start); err != nil {
			t.Fatal(err)
		}
		if err := sorter.flush(); err != nil {
			t.Fatal(err)
		}
	}
	merged, err := openEventMerge(sorter.files)
	if err != nil {
		t.Fatal(err)
	}
	defer merged.close()
	got, err := sweepEvents(merged, []string{"", "a", "b"}, ranks, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("chunk merge changed voting:\n%+v\n%+v", got, want)
	}
}

func TestDuplicateProviderRowsAndUnrankedTiesAreDeterministic(t *testing.T) {
	records := []model.Record{rec("8.8.8.0/24", "US", "mirror", 0), rec("8.8.8.0/24", "US", "mirror", 0), rec("8.8.8.0/24", "JP", "other", 1)}
	for i := 0; i < 10; i++ {
		got := resolveBoth(t, records, nil, nil)
		if len(got) != 1 || got[0].Country != "JP" || got[0].Confidence != 50 || got[0].TotalVotes.String() != "512" {
			t.Fatalf("duplicate vote or unstable tie: %+v", got)
		}
	}
}
