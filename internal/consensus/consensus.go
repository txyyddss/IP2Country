package consensus

import (
	"container/heap"
	"encoding/binary"
	"fmt"
	"math/big"
	"net/netip"
	"sort"

	"github.com/txyyddss/ip2country/internal/iprange"
	"github.com/txyyddss/ip2country/internal/model"
)

type Block struct {
	Prefix      netip.Prefix
	Country     string
	Confidence  float64
	WinnerVotes *big.Int
	TotalVotes  *big.Int
}

// position is an exact address boundary, including the exclusive end of /0.
type position struct {
	hi, lo   uint64
	afterMax bool
}

func addressPosition(a netip.Addr) position {
	if a.Is4() {
		b := a.As4()
		return position{lo: uint64(binary.BigEndian.Uint32(b[:]))}
	}
	b := a.As16()
	return position{hi: binary.BigEndian.Uint64(b[:8]), lo: binary.BigEndian.Uint64(b[8:])}
}

func (p position) addr(ipv6 bool) netip.Addr {
	if !ipv6 {
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], uint32(p.lo))
		return netip.AddrFrom4(b)
	}
	var b [16]byte
	binary.BigEndian.PutUint64(b[:8], p.hi)
	binary.BigEndian.PutUint64(b[8:], p.lo)
	return netip.AddrFrom16(b)
}

func endPosition(p netip.Prefix) position {
	next := iprange.PrefixLast(p).Next()
	if !next.IsValid() {
		return position{afterMax: true}
	}
	return addressPosition(next)
}

func lessPosition(a, b position) bool {
	if a.afterMax || b.afterMax {
		return !a.afterMax && b.afterMax
	}
	return a.hi < b.hi || a.hi == b.hi && a.lo < b.lo
}

func positionInt(p position, ipv6 bool) *big.Int {
	if p.afterMax {
		width := uint(32)
		if ipv6 {
			width = 128
		}
		return new(big.Int).Lsh(big.NewInt(1), width)
	}
	n := new(big.Int).SetUint64(p.hi)
	n.Lsh(n, 64)
	return n.Or(n, new(big.Int).SetUint64(p.lo))
}

func addressSpan(a, b position, ipv6 bool) *big.Int {
	return new(big.Int).Sub(positionInt(b, ipv6), positionInt(a, ipv6))
}

func recordEvents(r model.Record, group uint16) (sortEvent, sortEvent) {
	p := r.Prefix.Masked()
	e := sortEvent{ipv6: p.Addr().Is6(), at: addressPosition(p.Addr()), bits: uint8(p.Bits()), group: group,
		country: binary.BigEndian.Uint16([]byte(r.Country)), order: uint64(r.Order)<<32 | uint64(uint32(r.RowOrder)), delta: 1}
	end := e
	end.at, end.delta = endPosition(p), -1
	return e, end
}

func Resolve(records []model.Record, groupOrder map[string]int, exclusions []netip.Prefix) ([]Block, error) {
	groups := []string{""}
	ids := make(map[string]uint16)
	var events []sortEvent
	for _, r := range records {
		if !r.Prefix.IsValid() || len(r.Country) != 2 || r.VoteGroup == "" {
			continue
		}
		id, ok := ids[r.VoteGroup]
		if !ok {
			if len(groups) > 65535 {
				return nil, fmt.Errorf("too many vote groups")
			}
			id = uint16(len(groups))
			ids[r.VoteGroup] = id
			groups = append(groups, r.VoteGroup)
		}
		a, b := recordEvents(r, id)
		events = append(events, a, b)
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("no vote ranges after normalization")
	}
	sort.Slice(events, func(i, j int) bool { return lessEvent(events[i], events[j]) })
	return sweepEvents(&memoryEvents{events: events}, groups, groupOrder, exclusions)
}

type memoryEvents struct {
	events []sortEvent
	index  int
}

func (m *memoryEvents) next() (sortEvent, bool, error) {
	if m.index == len(m.events) {
		return sortEvent{}, false, nil
	}
	e := m.events[m.index]
	m.index++
	return e, true, nil
}

// Each group selects its most specific active assignment. Duplicate rows and
// mirrors never increase its vote; equal prefixes follow source and row order.
type assignmentKey struct {
	bits    uint8
	country uint16
	order   uint64
}
type assignment struct {
	key          assignmentKey
	count, index int
}
type assignmentHeap []*assignment

func (h assignmentHeap) Len() int { return len(h) }
func (h assignmentHeap) Less(i, j int) bool {
	a, b := h[i].key, h[j].key
	if a.bits != b.bits {
		return a.bits > b.bits
	}
	if a.order != b.order {
		return a.order < b.order
	}
	return a.country < b.country
}
func (h assignmentHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }
func (h *assignmentHeap) Push(v any)   { a := v.(*assignment); a.index = len(*h); *h = append(*h, a) }
func (h *assignmentHeap) Pop() any {
	old := *h
	a := old[len(old)-1]
	old[len(old)-1] = nil
	*h = old[:len(old)-1]
	return a
}

type groupVotes struct {
	assignments map[assignmentKey]*assignment
	heap        assignmentHeap
}

func (g *groupVotes) country() uint16 {
	if len(g.heap) == 0 {
		return 0
	}
	return g.heap[0].key.country
}

type votingState struct {
	groups     map[uint16]*groupVotes
	votes      map[uint16]int
	supporters map[uint16]map[uint16]bool
	ranks      []int
}

func newVotingState(groups []string, groupOrder map[string]int) *votingState {
	s := &votingState{groups: make(map[uint16]*groupVotes), votes: make(map[uint16]int), supporters: make(map[uint16]map[uint16]bool), ranks: make([]int, len(groups))}
	for i, g := range groups {
		s.ranks[i] = int(^uint(0) >> 1)
		if rank, ok := groupOrder[g]; ok {
			s.ranks[i] = rank
		}
	}
	return s
}
func (s *votingState) update(e sortEvent) error {
	g := s.groups[e.group]
	if g == nil {
		g = &groupVotes{assignments: make(map[assignmentKey]*assignment)}
		s.groups[e.group] = g
	}
	old := g.country()
	key := assignmentKey{bits: e.bits, country: e.country, order: e.order}
	a := g.assignments[key]
	if a == nil {
		if e.delta < 0 {
			return fmt.Errorf("unmatched end event for group %d", e.group)
		}
		a = &assignment{key: key}
		g.assignments[key] = a
		heap.Push(&g.heap, a)
	}
	a.count += int(e.delta)
	if a.count == 0 {
		heap.Remove(&g.heap, a.index)
		delete(g.assignments, key)
	}
	next := g.country()
	if old != next {
		if old != 0 {
			s.votes[old]--
			delete(s.supporters[old], e.group)
			if s.votes[old] == 0 {
				delete(s.votes, old)
				delete(s.supporters, old)
			}
		}
		if next != 0 {
			s.votes[next]++
			if s.supporters[next] == nil {
				s.supporters[next] = make(map[uint16]bool)
			}
			s.supporters[next][e.group] = true
		}
	}
	return nil
}
func (s *votingState) winner() (uint16, int, int) {
	winner := uint16(0)
	winning, total := 0, 0
	bestRank := int(^uint(0) >> 1)
	for country, count := range s.votes {
		total += count
		if count < winning {
			continue
		}
		rank := int(^uint(0) >> 1)
		for group := range s.supporters[country] {
			if s.ranks[group] < rank {
				rank = s.ranks[group]
			}
		}
		if count > winning || rank < bestRank || rank == bestRank && country < winner {
			winner, winning, bestRank = country, count, rank
		}
	}
	return winner, winning, total
}

type voteSpan struct {
	start, end     position
	winning, total int
}
type countryRun struct {
	ipv6    bool
	country uint16
	spans   []voteSpan
}

func (r *countryRun) append(span voteSpan) {
	if len(r.spans) > 0 {
		last := &r.spans[len(r.spans)-1]
		if last.end == span.start && last.winning == span.winning && last.total == span.total {
			last.end = span.end
			return
		}
	}
	r.spans = append(r.spans, span)
}

// Emit minimal CIDRs while integrating the actual vote evidence for each CIDR.
// This avoids copying a whole run's vote counts or confidence onto every part.
func (r *countryRun) blocks(filter *iprange.PrefixFilter) []Block {
	if len(r.spans) == 0 {
		return nil
	}
	start, end := r.spans[0].start, r.spans[len(r.spans)-1].end
	var last netip.Addr
	if end.afterMax {
		last = iprange.PrefixLast(netip.PrefixFrom(start.addr(r.ipv6), 0))
	} else {
		last = end.addr(r.ipv6).Prev()
	}
	var blocks []Block
	index := 0
	for _, p := range iprange.RangeToPrefixes(start.addr(r.ipv6), last) {
		for _, part := range filter.SubtractPrefix(p) {
			from, to := addressPosition(part.Addr()), endPosition(part)
			for index < len(r.spans) && !lessPosition(from, r.spans[index].end) {
				index++
			}
			winning, total := new(big.Int), new(big.Int)
			weighted := 0.0
			for i := index; i < len(r.spans) && lessPosition(r.spans[i].start, to); i++ {
				s := r.spans[i]
				a, b := from, to
				if lessPosition(a, s.start) {
					a = s.start
				}
				if lessPosition(s.end, b) {
					b = s.end
				}
				n := addressSpan(a, b, r.ipv6)
				winning.Add(winning, new(big.Int).Mul(n, big.NewInt(int64(s.winning))))
				total.Add(total, new(big.Int).Mul(n, big.NewInt(int64(s.total))))
				weighted += float64Big(n) * 100 * float64(s.winning) / float64(s.total)
			}
			code := string([]byte{byte(r.country >> 8), byte(r.country)})
			blocks = append(blocks, Block{Prefix: part, Country: code, Confidence: weighted / float64Big(addressSpan(from, to, r.ipv6)), WinnerVotes: winning, TotalVotes: total})
		}
	}
	return blocks
}
func float64Big(n *big.Int) float64 { f, _ := new(big.Float).SetInt(n).Float64(); return f }

type eventReader interface {
	next() (sortEvent, bool, error)
}

func sweepEvents(reader eventReader, groups []string, groupOrder map[string]int, exclusions []netip.Prefix) ([]Block, error) {
	state := newVotingState(groups, groupOrder)
	filter := iprange.NewPrefixFilter(exclusions)
	var result []Block
	var run countryRun
	flush := func() { result = append(result, run.blocks(filter)...); run.spans = run.spans[:0] }
	current, ok, err := reader.next()
	if err != nil {
		return nil, err
	}
	for ok {
		family, at := current.ipv6, current.at
		for {
			if err := state.update(current); err != nil {
				return nil, err
			}
			current, ok, err = reader.next()
			if err != nil {
				return nil, err
			}
			if !ok || current.ipv6 != family || current.at != at {
				break
			}
		}
		if !ok || current.ipv6 != family {
			flush()
			state = newVotingState(groups, groupOrder)
			continue
		}
		country, winning, total := state.winner()
		if total == 0 {
			flush()
			continue
		}
		if len(run.spans) > 0 && (run.country != country || run.ipv6 != family || run.spans[len(run.spans)-1].end != at) {
			flush()
		}
		run.country, run.ipv6 = country, family
		run.append(voteSpan{start: at, end: current.at, winning: winning, total: total})
	}
	return result, nil
}
