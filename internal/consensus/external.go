package consensus

import (
	"bufio"
	"compress/gzip"
	"container/heap"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"net/netip"
	"os"
	"path/filepath"
	"sort"

	"github.com/txyyddss/ip2country/internal/iprange"
)

const eventSize = 33
const eventChunkLimit = 250_000

const (
	eventBroad  = 0
	eventNarrow = 1
)

type sortEvent struct {
	ipv6    bool
	cell    coordinate
	kind    uint8
	group   uint16
	country uint16
	offset  coordinate
	order   uint64
	delta   int8
}

type eventSorter struct {
	dir      string
	chunk    []sortEvent
	files    []string
	sequence int
}

func newEventSorter(dir string) *eventSorter {
	return &eventSorter{dir: dir, chunk: make([]sortEvent, 0, eventChunkLimit)}
}

func (s *eventSorter) add(e sortEvent) error {
	s.chunk = append(s.chunk, e)
	if len(s.chunk) >= eventChunkLimit {
		return s.flush()
	}
	return nil
}

func (s *eventSorter) flush() error {
	if len(s.chunk) == 0 {
		return nil
	}
	sort.Slice(s.chunk, func(i, j int) bool { return lessEvent(s.chunk[i], s.chunk[j]) })
	path := filepath.Join(s.dir, fmt.Sprintf("events-%05d.bin.gz", s.sequence))
	s.sequence++
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	gz, err := gzip.NewWriterLevel(f, gzip.BestSpeed)
	if err != nil {
		_ = f.Close()
		return err
	}
	b := bufio.NewWriterSize(gz, 1<<20)
	var raw [eventSize]byte
	for _, e := range s.chunk {
		encodeEvent(raw[:], e)
		if _, err := b.Write(raw[:]); err != nil {
			_ = f.Close()
			return err
		}
	}
	if err := b.Flush(); err != nil {
		_ = f.Close()
		return err
	}
	if err := gz.Close(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	s.files = append(s.files, path)
	s.chunk = s.chunk[:0]
	return nil
}

func lessEvent(a, b sortEvent) bool {
	if a.ipv6 != b.ipv6 {
		return !a.ipv6
	}
	if a.cell != b.cell {
		return coordLess(a.cell, b.cell)
	}
	if a.kind != b.kind {
		return a.kind < b.kind
	}
	if a.group != b.group {
		return a.group < b.group
	}
	if a.country != b.country {
		return a.country < b.country
	}
	if a.offset != b.offset {
		return coordLess(a.offset, b.offset)
	}
	if a.order != b.order {
		return a.order < b.order
	}
	return a.delta < b.delta
}

func encodeEvent(b []byte, e sortEvent) {
	b[0] = 4
	if e.ipv6 {
		b[0] = 6
	}
	b[1] = e.kind
	binary.LittleEndian.PutUint64(b[2:10], e.cell.value)
	if e.cell.afterMax {
		b[10] = 1
	} else {
		b[10] = 0
	}
	binary.LittleEndian.PutUint16(b[11:13], e.group)
	binary.LittleEndian.PutUint16(b[13:15], e.country)
	binary.LittleEndian.PutUint64(b[15:23], e.offset.value)
	if e.offset.afterMax {
		b[23] = 1
	} else {
		b[23] = 0
	}
	binary.LittleEndian.PutUint64(b[24:32], e.order)
	b[32] = byte(e.delta)
}

func decodeEvent(b []byte) (sortEvent, error) {
	var e sortEvent
	switch b[0] {
	case 4:
		e.ipv6 = false
	case 6:
		e.ipv6 = true
	default:
		return e, fmt.Errorf("invalid event address family %d", b[0])
	}
	e.kind = b[1]
	e.cell = coordinate{value: binary.LittleEndian.Uint64(b[2:10]), afterMax: b[10] != 0}
	e.group = binary.LittleEndian.Uint16(b[11:13])
	e.country = binary.LittleEndian.Uint16(b[13:15])
	e.offset = coordinate{value: binary.LittleEndian.Uint64(b[15:23]), afterMax: b[23] != 0}
	e.order = binary.LittleEndian.Uint64(b[24:32])
	e.delta = int8(b[32])
	return e, nil
}

type eventCursor struct {
	file    *os.File
	gz      *gzip.Reader
	reader  *bufio.Reader
	current sortEvent
	valid   bool
}

func openEventCursor(path string) (*eventCursor, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	gz, err := gzip.NewReader(f)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	c := &eventCursor{file: f, gz: gz, reader: bufio.NewReaderSize(gz, 32*1024)}
	if err := c.advance(); err != nil {
		_ = gz.Close()
		_ = f.Close()
		return nil, err
	}
	return c, nil
}
func (c *eventCursor) advance() error {
	var raw [eventSize]byte
	_, err := io.ReadFull(c.reader, raw[:])
	if err == io.EOF {
		c.valid = false
		return nil
	}
	if err != nil {
		return err
	}
	e, err := decodeEvent(raw[:])
	if err != nil {
		return err
	}
	c.current = e
	c.valid = true
	return nil
}
func (c *eventCursor) close() {
	if c == nil {
		return
	}
	_ = c.gz.Close()
	_ = c.file.Close()
}

type cursorHeap []*eventCursor

func (h cursorHeap) Len() int           { return len(h) }
func (h cursorHeap) Less(i, j int) bool { return lessEvent(h[i].current, h[j].current) }
func (h cursorHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *cursorHeap) Push(v any)        { *h = append(*h, v.(*eventCursor)) }
func (h *cursorHeap) Pop() any          { old := *h; n := len(old); v := old[n-1]; *h = old[:n-1]; return v }

type eventMerge struct {
	heap cursorHeap
	all  []*eventCursor
}

func openEventMerge(paths []string) (*eventMerge, error) {
	m := &eventMerge{}
	for _, path := range paths {
		c, err := openEventCursor(path)
		if err != nil {
			m.close()
			return nil, err
		}
		m.all = append(m.all, c)
		if c.valid {
			heap.Push(&m.heap, c)
		}
	}
	heap.Init(&m.heap)
	return m, nil
}
func (m *eventMerge) next() (sortEvent, bool, error) {
	if len(m.heap) == 0 {
		return sortEvent{}, false, nil
	}
	c := heap.Pop(&m.heap).(*eventCursor)
	event := c.current
	if err := c.advance(); err != nil {
		return sortEvent{}, false, err
	}
	if c.valid {
		heap.Push(&m.heap, c)
	}
	return event, true, nil
}
func (m *eventMerge) close() {
	for _, c := range m.all {
		c.close()
	}
	m.all = nil
	m.heap = nil
}

func ResolveSpool(path string, groups []string, groupOrder map[string]int, exclusions []netip.Prefix) ([]Block, error) {
	tempDir, err := os.MkdirTemp(filepath.Dir(path), "ip2country-events-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)
	sorter := newEventSorter(tempDir)
	if err := spoolToEvents(path, groups, sorter); err != nil {
		return nil, err
	}
	if err := sorter.flush(); err != nil {
		return nil, err
	}
	if len(sorter.files) == 0 {
		return nil, fmt.Errorf("no vote ranges after normalization")
	}
	merge, err := openEventMerge(sorter.files)
	if err != nil {
		return nil, err
	}
	defer merge.close()
	return sweepEvents(merge, groups, groupOrder, exclusions)
}

func spoolToEvents(path string, groups []string, sorter *eventSorter) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	r := bufio.NewReaderSize(gz, 1<<20)
	for {
		family, err := r.ReadByte()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		var bitsBuf [1]byte
		if _, err := io.ReadFull(r, bitsBuf[:]); err != nil {
			return err
		}
		ipv6 := family == 6
		addrLen := 4
		if ipv6 {
			addrLen = 16
		} else if family != 4 {
			return fmt.Errorf("invalid spooled address family %d", family)
		}
		var addrBuf [16]byte
		if _, err := io.ReadFull(r, addrBuf[:addrLen]); err != nil {
			return err
		}
		var tail [10]byte
		if _, err := io.ReadFull(r, tail[:]); err != nil {
			return err
		}
		group := binary.LittleEndian.Uint16(tail[2:4])
		if group == 0 || int(group) >= len(groups) {
			return fmt.Errorf("invalid spooled vote group %d", group)
		}
		country := binary.BigEndian.Uint16(tail[:2])
		sourceOrder := binary.LittleEndian.Uint16(tail[4:6])
		rowOrder := binary.LittleEndian.Uint32(tail[6:10])
		order := uint64(sourceOrder)<<32 | uint64(rowOrder)
		var addr netip.Addr
		if ipv6 {
			addr = netip.AddrFrom16(addrBuf)
		} else {
			var a [4]byte
			copy(a[:], addrBuf[:4])
			addr = netip.AddrFrom4(a)
		}
		bits := int(bitsBuf[0])
		if bits > addr.BitLen() {
			return fmt.Errorf("invalid spooled prefix length %d", bits)
		}
		p := netip.PrefixFrom(addr, bits).Masked()
		unit := 24
		if ipv6 {
			unit = 64
		}
		startCell := cellIndex(p.Addr(), ipv6)
		if p.Bits() <= unit {
			endCell := cellIndex(iprange.PrefixLast(p), ipv6)
			end := cellEndCoordinate(ipv6, endCell)
			if err := sorter.add(sortEvent{ipv6: ipv6, cell: coordinate{value: startCell}, kind: eventBroad, group: group, country: country, order: order, delta: 1}); err != nil {
				return err
			}
			if err := sorter.add(sortEvent{ipv6: ipv6, cell: end, kind: eventBroad, group: group, country: country, order: order, delta: -1}); err != nil {
				return err
			}
		} else {
			start, end := cellOffsets(p, ipv6)
			endExclusive := coordinate{value: end + 1}
			if end == ^uint64(0) {
				endExclusive = coordinate{afterMax: true}
			}
			cell := coordinate{value: startCell}
			if err := sorter.add(sortEvent{ipv6: ipv6, cell: cell, kind: eventNarrow, group: group, country: country, offset: coordinate{value: start}, order: order, delta: 1}); err != nil {
				return err
			}
			if err := sorter.add(sortEvent{ipv6: ipv6, cell: cell, kind: eventNarrow, group: group, country: country, offset: endExclusive, order: order, delta: -1}); err != nil {
				return err
			}
		}
	}
}

type broadState struct {
	count  int
	orders map[uint64]int
	first  uint64
}
type partialState struct {
	coverage *big.Int
	first    uint64
}
type countryChoice struct {
	country  uint16
	coverage *big.Int
	first    uint64
}

type eventStream struct {
	merge   *eventMerge
	pending *sortEvent
}

func (s *eventStream) next() (sortEvent, bool, error) {
	if s.pending != nil {
		e := *s.pending
		s.pending = nil
		return e, true, nil
	}
	return s.merge.next()
}
func (s *eventStream) push(e sortEvent) error {
	if s.pending != nil {
		return fmt.Errorf("event stream pending slot is occupied")
	}
	s.pending = &e
	return nil
}

func sweepEvents(merge *eventMerge, groups []string, groupOrder map[string]int, exclusions []netip.Prefix) ([]Block, error) {
	active := make(map[uint32]*broadState)
	choices := make(map[uint16]countryChoice, 64)
	votes := make(map[uint16]int, 64)
	firstGroup := make(map[uint16]int, 64)
	full4 := big.NewInt(256)
	full6 := new(big.Int).Lsh(big.NewInt(1), 64)
	ranks := make([]int, len(groups))
	for i, g := range groups {
		if rank, ok := groupOrder[g]; ok {
			ranks[i] = rank
		} else {
			ranks[i] = int(^uint(0) >> 1)
		}
	}
	filter := iprange.NewPrefixFilter(exclusions)
	var result []Block
	var current *Block
	flush := func() {
		if current == nil {
			return
		}
		result = append(result, runPrefixes(*current, filter)...)
		current = nil
	}
	addSegment := func(ipv6 bool, start uint64, end coordinate, block Block, count *big.Int) {
		if block.Country == "" || count.Sign() == 0 {
			flush()
			return
		}
		endValue := uint64(0)
		after := end.afterMax
		if !after {
			if end.value == 0 {
				return
			}
			endValue = end.value - 1
		} else {
			endValue = maxCell(ipv6)
		}
		if current != nil && current.Country == block.Country && current.IPv6 == ipv6 && !current.EndAfterMax && current.EndCell+1 == start {
			weighted := current.Confidence*float64Big(current.CellCount) + block.Confidence*float64Big(count)
			current.CellCount.Add(current.CellCount, count)
			current.Confidence = weighted / float64Big(current.CellCount)
			current.EndCell = endValue
			current.EndAfterMax = after
			current.WinnerVotes.Add(current.WinnerVotes, block.WinnerVotes)
			current.TotalVotes.Add(current.TotalVotes, block.TotalVotes)
			return
		}
		flush()
		copy := block
		copy.IPv6 = ipv6
		copy.StartCell = start
		copy.EndCell = endValue
		copy.EndAfterMax = after
		copy.CellCount = new(big.Int).Set(count)
		current = &copy
	}
	stream := &eventStream{merge: merge}
	partials := make(map[uint32]partialState)
	for {
		first, ok, err := stream.next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		family, cell := first.ipv6, first.cell
		if cell.afterMax {
			e := first
			for e.ipv6 == family && e.cell == cell && e.kind == eventBroad {
				updateBroad(active, e)
				e, ok, err = stream.next()
				if err != nil {
					return nil, err
				}
				if !ok {
					break
				}
			}
			if ok {
				if err := stream.push(e); err != nil {
					return nil, err
				}
			}
			continue
		}
		clear(partials)
		partialUsed := false
		e := first
		for e.ipv6 == family && e.cell == cell && e.kind == eventBroad {
			updateBroad(active, e)
			e, ok, err = stream.next()
			if err != nil {
				return nil, err
			}
			if !ok {
				break
			}
		}
		for ok && e.ipv6 == family && e.cell == cell && e.kind == eventNarrow {
			partialUsed = true
			pair := uint32(e.group)<<16 | uint32(e.country)
			coverage, firstOrder, err := consumeNarrowGroup(e, stream)
			if err != nil {
				return nil, err
			}
			if coverage.Sign() > 0 {
				partials[pair] = partialState{coverage: coverage, first: firstOrder}
			}
			e, ok, err = stream.next()
			if err != nil {
				return nil, err
			}
		}
		if ok {
			if err := stream.push(e); err != nil {
				return nil, err
			}
		}
		var pointVotes map[uint32]partialState
		if partialUsed {
			pointVotes = partials
		}
		block, found := evaluateNumeric(family, active, pointVotes, choices, votes, firstGroup, ranks, full4, full6)
		end := cellEndCoordinate(family, cell.value)
		one := big.NewInt(1)
		if found {
			addSegment(family, cell.value, end, block, one)
		} else {
			flush()
		}
		following, hasNext, err := stream.next()
		if err != nil {
			return nil, err
		}
		if !hasNext {
			break
		}
		if following.ipv6 != family {
			if err := stream.push(following); err != nil {
				return nil, err
			}
			continue
		}
		gapStart := cellEndCoordinate(family, cell.value)
		if cmpCoord(gapStart, following.cell) < 0 {
			gapBlock, found := evaluateNumeric(family, active, nil, choices, votes, firstGroup, ranks, full4, full6)
			count := cellSpanFamily(gapStart, following.cell, family)
			if found {
				gapBlock.WinnerVotes.Mul(gapBlock.WinnerVotes, count)
				gapBlock.TotalVotes.Mul(gapBlock.TotalVotes, count)
				addSegment(family, gapStart.value, following.cell, gapBlock, count)
			} else {
				flush()
			}
		}
		if err := stream.push(following); err != nil {
			return nil, err
		}
	}
	flush()
	return result, nil
}

func updateBroad(active map[uint32]*broadState, e sortEvent) {
	pair := uint32(e.group)<<16 | uint32(e.country)
	state := active[pair]
	if state == nil {
		state = &broadState{orders: make(map[uint64]int), first: ^uint64(0)}
		active[pair] = state
	}
	state.count += int(e.delta)
	state.orders[e.order] += int(e.delta)
	if state.orders[e.order] <= 0 {
		delete(state.orders, e.order)
	}
	if e.delta > 0 && e.order < state.first {
		state.first = e.order
	}
	if state.count <= 0 {
		delete(active, pair)
		return
	}
	if state.first == e.order && e.delta < 0 {
		state.first = ^uint64(0)
		for order := range state.orders {
			if order < state.first {
				state.first = order
			}
		}
	}
}

func consumeNarrowGroup(first sortEvent, stream *eventStream) (*big.Int, uint64, error) {
	coverage := new(big.Int)
	minOrder := ^uint64(0)
	active := 0
	previous := coordinate{value: 0}
	e := first
	for {
		if cmpCoord(previous, e.offset) < 0 && active > 0 {
			coverage.Add(coverage, offsetSpan(previous, e.offset))
		}
		currentOffset := e.offset
		previous = currentOffset
		for {
			active += int(e.delta)
			if e.delta > 0 && e.order < minOrder {
				minOrder = e.order
			}
			next, ok, err := stream.next()
			if err != nil {
				return nil, 0, err
			}
			if !ok {
				return coverage, minOrder, nil
			}
			if next.ipv6 != first.ipv6 || next.cell != first.cell || next.kind != eventNarrow || next.group != first.group || next.country != first.country {
				if err := stream.push(next); err != nil {
					return nil, 0, err
				}
				return coverage, minOrder, nil
			}
			if next.offset != currentOffset {
				e = next
				break
			}
			e = next
		}
	}
}

func offsetSpan(a, b coordinate) *big.Int { return cellSpanFamily(a, b, true) }

func evaluateNumeric(ipv6 bool, active map[uint32]*broadState, partials map[uint32]partialState, choices map[uint16]countryChoice, votes map[uint16]int, firstGroup map[uint16]int, ranks []int, full4, full6 *big.Int) (Block, bool) {
	clear(choices)
	clear(votes)
	clear(firstGroup)
	full := full4
	if ipv6 {
		full = full6
	}
	for pair, state := range active {
		group := uint16(pair >> 16)
		country := uint16(pair)
		if state.count <= 0 {
			continue
		}
		addChoice(choices, group, country, full, state.first)
	}
	for pair, part := range partials {
		group := uint16(pair >> 16)
		country := uint16(pair)
		addChoice(choices, group, country, part.coverage, part.first)
	}
	total := 0
	for group, choice := range choices {
		total++
		votes[choice.country]++
		rank := int(^uint(0) >> 1)
		if int(group) < len(ranks) {
			rank = ranks[group]
		}
		if prev, ok := firstGroup[choice.country]; !ok || rank < prev {
			firstGroup[choice.country] = rank
		}
	}
	if total == 0 {
		return Block{}, false
	}
	winner := uint16(0)
	maxVotes := -1
	winnerRank := int(^uint(0) >> 1)
	for country, count := range votes {
		rank := firstGroup[country]
		if count > maxVotes || (count == maxVotes && rank < winnerRank) {
			winner, maxVotes, winnerRank = country, count, rank
		}
	}
	code := []byte{byte(winner >> 8), byte(winner)}
	return Block{Country: string(code), Confidence: 100 * float64(maxVotes) / float64(total), WinnerVotes: big.NewInt(int64(maxVotes)), TotalVotes: big.NewInt(int64(total))}, true
}

func addChoice(choices map[uint16]countryChoice, group, country uint16, coverage *big.Int, order uint64) {
	old, ok := choices[group]
	if !ok || coverage.Cmp(old.coverage) > 0 || (coverage.Cmp(old.coverage) == 0 && order < old.first) {
		choices[group] = countryChoice{country: country, coverage: coverage, first: order}
	}
}

func runPrefixes(run Block, filter *iprange.PrefixFilter) []Block {
	end := coordinate{value: run.EndCell + 1}
	if run.EndAfterMax {
		end = coordinate{afterMax: true}
	}
	start, last := cellsToRange(run.IPv6, run.StartCell, end)
	var out []Block
	for _, p := range iprange.RangeToPrefixes(start, last) {
		for _, part := range filter.SubtractPrefix(p) {
			copy := run
			copy.Prefix = part
			out = append(out, copy)
		}
	}
	return out
}
