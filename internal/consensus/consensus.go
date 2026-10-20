package consensus

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"net/netip"
	"sort"
	"strings"

	"github.com/txyyddss/ip2country/internal/iprange"
	"github.com/txyyddss/ip2country/internal/model"
)

type Block struct {
	Prefix      netip.Prefix
	Country     string
	Confidence  float64
	WinnerVotes *big.Int
	TotalVotes  *big.Int
	CellCount   *big.Int
	IPv6        bool
	StartCell   uint64
	EndCell     uint64
	EndAfterMax bool
}

type coordinate struct {
	value    uint64
	afterMax bool
}
type cellKey struct {
	ipv6  bool
	value uint64
}
type event struct {
	key            coordinate
	ipv6           bool
	group, country string
	order, delta   int
}
type narrowRange struct {
	start, end     uint64
	group, country string
	order          int
}
type sourceCountry struct {
	count  int
	orders map[int]int
}

func Resolve(records []model.Record, groupOrder map[string]int, exclusions []netip.Prefix) ([]Block, error) {
	return resolveEach(func(emit func(model.Record) error) error {
		for _, record := range records {
			if err := emit(record); err != nil {
				return err
			}
		}
		return nil
	}, groupOrder, exclusions)
}

func resolveEach(each func(func(model.Record) error) error, groupOrder map[string]int, exclusions []netip.Prefix) ([]Block, error) {
	var events []event
	points := make(map[cellKey][]narrowRange)
	readErr := each(func(r model.Record) error {
		p := r.Prefix.Masked()
		if !p.IsValid() || r.Country == "" || r.VoteGroup == "" {
			return nil
		}
		rowOrder := int((uint64(r.Order) << 32) | uint64(uint32(r.RowOrder)))
		ipv6 := p.Addr().Is6()
		unit := 24
		if ipv6 {
			unit = 64
		}
		startCell := cellIndex(p.Addr(), ipv6)
		if p.Bits() <= unit {
			endCell := cellIndex(iprange.PrefixLast(p), ipv6)
			start := coordinate{value: startCell}
			end := cellEndCoordinate(ipv6, endCell)
			events = append(events, event{key: start, ipv6: ipv6, group: r.VoteGroup, country: r.Country, order: rowOrder, delta: 1}, event{key: end, ipv6: ipv6, group: r.VoteGroup, country: r.Country, order: rowOrder, delta: -1})
		} else {
			start, end := cellOffsets(p, ipv6)
			ck := cellKey{ipv6: ipv6, value: startCell}
			points[ck] = append(points[ck], narrowRange{start: start, end: end, group: r.VoteGroup, country: r.Country, order: rowOrder})
		}
		return nil
	})
	if readErr != nil {
		return nil, readErr
	}
	for key := range points {
		end := cellEndCoordinate(key.ipv6, key.value)
		events = append(events, event{key: coordinate{value: key.value}, ipv6: key.ipv6}, event{key: end, ipv6: key.ipv6})
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("no vote ranges after normalization")
	}
	sort.Slice(events, func(i, j int) bool {
		a, b := events[i], events[j]
		if a.ipv6 != b.ipv6 {
			return !a.ipv6
		}
		return coordLess(a.key, b.key)
	})

	active := map[bool]map[string]map[string]*sourceCountry{false: {}, true: {}}
	var runs []Block
	for i := 0; i < len(events); {
		j := i + 1
		for j < len(events) && events[j].ipv6 == events[i].ipv6 && coordEqual(events[j].key, events[i].key) {
			j++
		}
		family := events[i].ipv6
		if active[family] == nil {
			active[family] = make(map[string]map[string]*sourceCountry)
		}
		for k := i; k < j; k++ {
			e := events[k]
			if e.group == "" {
				continue
			}
			byCountry := active[family][e.group]
			if byCountry == nil {
				byCountry = make(map[string]*sourceCountry)
				active[family][e.group] = byCountry
			}
			sc := byCountry[e.country]
			if sc == nil {
				sc = &sourceCountry{orders: make(map[int]int)}
				byCountry[e.country] = sc
			}
			sc.count += e.delta
			sc.orders[e.order] += e.delta
			if sc.orders[e.order] <= 0 {
				delete(sc.orders, e.order)
			}
			if sc.count <= 0 {
				delete(byCountry, e.country)
			}
			if len(byCountry) == 0 {
				delete(active[family], e.group)
			}
		}
		if j < len(events) && events[j].ipv6 == family {
			from, to := events[i].key, events[j].key
			if cmpCoord(from, to) < 0 {
				if from.afterMax {
					i = j
					continue
				}
				endValue := to.value - 1
				if to.afterMax {
					endValue = maxCell(family)
				}
				if from.value == endValue && len(points[cellKey{ipv6: family, value: from.value}]) > 0 {
					block, ok := evaluateCell(family, from.value, active[family], points[cellKey{ipv6: family, value: from.value}], groupOrder)
					if ok {
						block.CellCount = big.NewInt(1)
						block.IPv6 = family
						block.StartCell = from.value
						block.EndCell = endValue
						block.EndAfterMax = to.afterMax
						runs = append(runs, block)
					}
				} else {
					// Point cells have start/end marker events, so a wider segment has no partial assignments.
					block, ok := evaluateCell(family, from.value, active[family], nil, groupOrder)
					if ok {
						block.CellCount = cellSpanFamily(from, to, family)
						block.IPv6 = family
						block.StartCell = from.value
						block.EndCell = endValue
						block.EndAfterMax = to.afterMax
						block.WinnerVotes.Mul(block.WinnerVotes, block.CellCount)
						block.TotalVotes.Mul(block.TotalVotes, block.CellCount)
						runs = append(runs, block)
					}
				}
			}
		}
		i = j
	}
	merged := mergeRuns(runs)
	return toPrefixes(merged, exclusions), nil
}

func evaluateCell(ipv6 bool, cell uint64, active map[string]map[string]*sourceCountry, narrow []narrowRange, groupOrder map[string]int) (Block, bool) {
	coverage := make(map[string]map[string]*big.Int)
	firstOrder := make(map[string]map[string]int)
	full := new(big.Int).Lsh(big.NewInt(1), 8)
	if ipv6 {
		full.Lsh(big.NewInt(1), 64)
	}
	for group, countries := range active {
		for country, sc := range countries {
			ensureCoverage(coverage, group)
			coverage[group][country] = new(big.Int).Set(full)
			ensureOrder(firstOrder, group)
			firstOrder[group][country] = minOrder(sc.orders)
		}
	}
	if len(narrow) > 0 {
		byCountry := make(map[string]map[string][]narrowRange)
		for _, r := range narrow {
			if byCountry[r.group] == nil {
				byCountry[r.group] = make(map[string][]narrowRange)
			}
			byCountry[r.group][r.country] = append(byCountry[r.group][r.country], r)
		}
		for group, countries := range byCountry {
			for country, ranges := range countries {
				n := unionLength(ranges)
				ensureCoverage(coverage, group)
				if old := coverage[group][country]; old == nil || n.Cmp(old) > 0 {
					coverage[group][country] = n
				}
				ensureOrder(firstOrder, group)
				order := ranges[0].order
				for _, r := range ranges {
					if r.order < order {
						order = r.order
					}
				}
				if old, ok := firstOrder[group][country]; !ok || order < old {
					firstOrder[group][country] = order
				}
			}
		}
	}
	votes := make(map[string]int)
	firstGroup := make(map[string]int)
	total := 0
	for group, countries := range coverage {
		chosen := ""
		var max *big.Int
		order := int(^uint(0) >> 1)
		for country, n := range countries {
			rowOrder := firstOrder[group][country]
			if max == nil || n.Cmp(max) > 0 || (n.Cmp(max) == 0 && rowOrder < order) {
				chosen, max, order = country, n, rowOrder
			}
		}
		if chosen != "" {
			total++
			votes[chosen]++
			rank, ok := groupOrder[group]
			if !ok {
				rank = int(^uint(0) >> 1)
			}
			if old, ok := firstGroup[chosen]; !ok || rank < old {
				firstGroup[chosen] = rank
			}
		}
	}
	if total == 0 {
		return Block{}, false
	}
	winner, maxVotes := "", -1
	winnerRank := int(^uint(0) >> 1)
	for country, count := range votes {
		rank := firstGroup[country]
		if count > maxVotes || (count == maxVotes && rank < winnerRank) {
			winner, maxVotes, winnerRank = country, count, rank
		}
	}
	return Block{Country: winner, Confidence: 100 * float64(maxVotes) / float64(total), WinnerVotes: big.NewInt(int64(maxVotes)), TotalVotes: big.NewInt(int64(total))}, true
}

func ensureCoverage(m map[string]map[string]*big.Int, group string) {
	if m[group] == nil {
		m[group] = make(map[string]*big.Int)
	}
}
func ensureOrder(m map[string]map[string]int, group string) {
	if m[group] == nil {
		m[group] = make(map[string]int)
	}
}
func minOrder(m map[int]int) int {
	v := int(^uint(0) >> 1)
	for k := range m {
		if k < v {
			v = k
		}
	}
	return v
}

func unionLength(ranges []narrowRange) *big.Int {
	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].start != ranges[j].start {
			return ranges[i].start < ranges[j].start
		}
		return ranges[i].end < ranges[j].end
	})
	total := new(big.Int)
	start, end := ranges[0].start, ranges[0].end
	for _, r := range ranges[1:] {
		if r.start <= end || (end != ^uint64(0) && r.start == end+1) {
			if r.end > end {
				end = r.end
			}
			continue
		}
		total.Add(total, rangeLength(start, end))
		start, end = r.start, r.end
	}
	total.Add(total, rangeLength(start, end))
	return total
}

func rangeLength(start, end uint64) *big.Int {
	n := new(big.Int).SetUint64(end)
	n.Sub(n, new(big.Int).SetUint64(start))
	return n.Add(n, big.NewInt(1))
}

func cellOffsets(p netip.Prefix, ipv6 bool) (uint64, uint64) {
	if !ipv6 {
		a := p.Addr().As4()
		start := uint64(a[3])
		host := uint(32 - p.Bits())
		size := uint64(1) << host
		return start, start + size - 1
	}
	a := p.Addr().As16()
	start := binary.BigEndian.Uint64(a[8:])
	host := uint(128 - p.Bits())
	size := uint64(1) << host
	return start, start + size - 1
}

func cellIndex(a netip.Addr, ipv6 bool) uint64 {
	if !ipv6 {
		b := a.As4()
		return uint64(binary.BigEndian.Uint32(b[:]) >> 8)
	}
	b := a.As16()
	return binary.BigEndian.Uint64(b[:8])
}
func coordLess(a, b coordinate) bool {
	if a.afterMax {
		return false
	}
	if b.afterMax {
		return true
	}
	return a.value < b.value
}
func coordEqual(a, b coordinate) bool { return a == b }
func cmpCoord(a, b coordinate) int {
	if coordEqual(a, b) {
		return 0
	}
	if coordLess(a, b) {
		return -1
	}
	return 1
}
func maxCell(ipv6 bool) uint64 {
	if ipv6 {
		return ^uint64(0)
	}
	return (1 << 24) - 1
}
func cellEndCoordinate(ipv6 bool, last uint64) coordinate {
	if last == maxCell(ipv6) {
		return coordinate{afterMax: true}
	}
	return coordinate{value: last + 1}
}
func cellSpanFamily(a, b coordinate, ipv6 bool) *big.Int {
	if b.afterMax {
		n := new(big.Int).SetUint64(maxCell(ipv6) - a.value)
		return n.Add(n, big.NewInt(1))
	}
	return new(big.Int).SetUint64(b.value - a.value)
}

func mergeRuns(in []Block) []Block {
	if len(in) == 0 {
		return nil
	}
	var out []Block
	for _, b := range in {
		if b.CellCount == nil || b.CellCount.Sign() == 0 {
			continue
		}
		if len(out) > 0 && out[len(out)-1].Country == b.Country && adjacentRuns(out[len(out)-1], b) {
			prev := &out[len(out)-1]
			n := new(big.Int).Add(prev.CellCount, b.CellCount)
			weighted := prev.Confidence*float64Big(prev.CellCount) + b.Confidence*float64Big(b.CellCount)
			prev.Confidence = weighted / float64Big(n)
			prev.CellCount = n
			prev.EndCell = b.EndCell
			prev.EndAfterMax = b.EndAfterMax
			prev.WinnerVotes.Add(prev.WinnerVotes, b.WinnerVotes)
			prev.TotalVotes.Add(prev.TotalVotes, b.TotalVotes)
		} else {
			out = append(out, b)
		}
	}
	return out
}

// Float conversion saturates only beyond 2^53; the average remains stable for a single run.
func float64Big(n *big.Int) float64 { f, _ := new(big.Float).SetInt(n).Float64(); return f }
func adjacentRuns(a, b Block) bool {
	if a.IPv6 != b.IPv6 || a.EndAfterMax {
		return false
	}
	return a.EndCell+1 == b.StartCell
}

func toPrefixes(runs []Block, exclusions []netip.Prefix) []Block {
	filter := iprange.NewPrefixFilter(exclusions)
	out := make([]Block, 0, len(runs))
	for _, b := range runs {
		end := coordinate{value: b.EndCell + 1}
		if b.EndAfterMax {
			end = coordinate{afterMax: true}
		}
		start, last := cellsToRange(b.IPv6, b.StartCell, end)
		base := iprange.RangeToPrefixes(start, last)
		var parts []netip.Prefix
		for _, prefix := range base {
			parts = append(parts, filter.SubtractPrefix(prefix)...)
		}
		for _, p := range parts {
			copy := b
			copy.Prefix = p
			out = append(out, copy)
		}
	}
	return out
}

func cellsToRange(ipv6 bool, start uint64, end coordinate) (netip.Addr, netip.Addr) {
	var a, z netip.Addr
	if !ipv6 {
		lo := uint32(start) << 8
		hi := uint32(lo)
		if end.afterMax {
			hi = ^uint32(0)
		} else {
			hi = (uint32(end.value-1) << 8) | 255
		}
		var lb, ub [4]byte
		binary.BigEndian.PutUint32(lb[:], lo)
		binary.BigEndian.PutUint32(ub[:], hi)
		a, z = netip.AddrFrom4(lb), netip.AddrFrom4(ub)
		return a, z
	}
	var lb, ub [16]byte
	binary.BigEndian.PutUint64(lb[:8], start)
	if end.afterMax {
		for i := 8; i < 16; i++ {
			ub[i] = 255
		}
		for i := 0; i < 8; i++ {
			ub[i] = 255
		}
	} else {
		binary.BigEndian.PutUint64(ub[:8], end.value-1)
		for i := 8; i < 16; i++ {
			ub[i] = 255
		}
	}
	a, z = netip.AddrFrom16(lb), netip.AddrFrom16(ub)
	return a, z
}

var _ = strings.Compare
