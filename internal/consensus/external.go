package consensus

import (
	"bufio"
	"compress/gzip"
	"container/heap"
	"encoding/binary"
	"fmt"
	"io"
	"net/netip"
	"os"
	"path/filepath"
	"sort"

	"github.com/txyyddss/ip2country/internal/model"
)

const eventSize = 32
const eventChunkLimit = 250_000

// Events are sorted at exact address boundaries, including sub-/24 and sub-/64 prefixes.
type sortEvent struct {
	ipv6           bool
	at             position
	bits           uint8
	group, country uint16
	order          uint64
	delta          int8
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
	if a.at != b.at {
		return lessPosition(a.at, b.at)
	}
	if a.group != b.group {
		return a.group < b.group
	}
	if a.bits != b.bits {
		return a.bits < b.bits
	}
	if a.country != b.country {
		return a.country < b.country
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
	b[1] = e.bits
	binary.LittleEndian.PutUint64(b[2:10], e.at.hi)
	binary.LittleEndian.PutUint64(b[10:18], e.at.lo)
	b[18] = 0
	if e.at.afterMax {
		b[18] = 1
	}
	binary.LittleEndian.PutUint16(b[19:21], e.group)
	binary.LittleEndian.PutUint16(b[21:23], e.country)
	binary.LittleEndian.PutUint64(b[23:31], e.order)
	b[31] = byte(e.delta)
}

func decodeEvent(b []byte) (sortEvent, error) {
	e := sortEvent{ipv6: b[0] == 6, bits: b[1], at: position{hi: binary.LittleEndian.Uint64(b[2:10]), lo: binary.LittleEndian.Uint64(b[10:18]), afterMax: b[18] != 0}, group: binary.LittleEndian.Uint16(b[19:21]), country: binary.LittleEndian.Uint16(b[21:23]), order: binary.LittleEndian.Uint64(b[23:31]), delta: int8(b[31])}
	if b[0] != 4 && b[0] != 6 {
		return e, fmt.Errorf("invalid event address family %d", b[0])
	}
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
		bits, err := r.ReadByte()
		if err != nil {
			return err
		}
		addrLen := 4
		if family == 6 {
			addrLen = 16
		} else if family != 4 {
			return fmt.Errorf("invalid spooled address family %d", family)
		}
		var raw [16]byte
		if _, err := io.ReadFull(r, raw[:addrLen]); err != nil {
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
		addr, _ := netip.AddrFromSlice(raw[:addrLen])
		if int(bits) > addr.BitLen() {
			return fmt.Errorf("invalid spooled prefix length %d", bits)
		}
		record := model.Record{Prefix: netip.PrefixFrom(addr, int(bits)), Country: string(tail[:2]), Order: int(binary.LittleEndian.Uint16(tail[4:6])), RowOrder: int(binary.LittleEndian.Uint32(tail[6:10]))}
		start, end := recordEvents(record, group)
		if err := sorter.add(start); err != nil {
			return err
		}
		if err := sorter.add(end); err != nil {
			return err
		}
	}
}
