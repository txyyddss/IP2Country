package consensus

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/txyyddss/ip2country/internal/model"
)

type RecordSpool struct {
	file       *os.File
	buffered   *bufio.Writer
	compressed *gzip.Writer
	groupIDs   map[string]uint16
}

func NewRecordSpool(path string, groupIDs map[string]uint16) (*RecordSpool, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	b := bufio.NewWriterSize(f, 1<<20)
	gz, err := gzip.NewWriterLevel(b, gzip.BestSpeed)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	return &RecordSpool{file: f, buffered: b, compressed: gz, groupIDs: groupIDs}, nil
}

func (s *RecordSpool) Write(r model.Record) error {
	group, ok := s.groupIDs[r.VoteGroup]
	if !ok || group == 0 {
		return fmt.Errorf("unknown vote group %q", r.VoteGroup)
	}
	if len(r.Country) != 2 || !r.Prefix.IsValid() || r.Evidence > model.Fallback {
		return fmt.Errorf("invalid country/network record %q %s", r.Country, r.Prefix)
	}
	p := r.Prefix.Masked()
	family := byte(6)
	var addr []byte
	if p.Addr().Is4() {
		family = 4
		a := p.Addr().As4()
		addr = a[:]
	} else {
		a := p.Addr().As16()
		addr = a[:]
	}
	if r.Order < 0 || r.Order > 65535 || r.RowOrder < 0 || uint64(r.RowOrder) > uint64(^uint32(0)) {
		return fmt.Errorf("source order or row order exceeds spool format")
	}
	var header [2 + 2 + 2 + 2 + 4 + 1]byte
	header[0] = family
	header[1] = byte(p.Bits())
	header[12] = byte(r.Evidence)
	copy(header[2:4], r.Country)
	binary.LittleEndian.PutUint16(header[4:6], group)
	binary.LittleEndian.PutUint16(header[6:8], uint16(r.Order))
	binary.LittleEndian.PutUint32(header[8:12], uint32(r.RowOrder))
	if _, err := s.compressed.Write(header[:2]); err != nil {
		return err
	}
	if _, err := s.compressed.Write(addr); err != nil {
		return err
	}
	_, err := s.compressed.Write(header[2:13])
	return err
}

func (s *RecordSpool) Close() error {
	if s == nil || s.file == nil {
		return nil
	}
	if err := s.compressed.Close(); err != nil {
		_ = s.file.Close()
		return err
	}
	if err := s.buffered.Flush(); err != nil {
		_ = s.file.Close()
		return err
	}
	if err := s.file.Sync(); err != nil {
		_ = s.file.Close()
		return err
	}
	return s.file.Close()
}
