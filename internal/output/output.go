package output

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/biter777/countries"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/txyyddss/ip2country/internal/consensus"
	"github.com/txyyddss/ip2country/internal/iprange"
	"github.com/txyyddss/ip2country/internal/model"
	"github.com/v2fly/geoip/plugin/v2ray"
	"go4.org/netipx"
	"google.golang.org/protobuf/proto"
)

type RuleSet struct {
	Version int      `json:"version"`
	Rules   []IPRule `json:"rules"`
}

type IPRule struct {
	IPCIDR []string `json:"ip_cidr"`
}

func Generate(root string, blocks []consensus.Block, singBoxBinary string) error {
	if len(blocks) == 0 {
		return fmt.Errorf("no consensus blocks to publish")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	for _, name := range []string{"countries", "sing-box"} {
		if err := os.RemoveAll(filepath.Join(root, name)); err != nil {
			return err
		}
	}
	for _, name := range []string{"geoip.dat", "country.mmdb"} {
		if err := os.Remove(filepath.Join(root, name)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	byCountry := make(map[string][]consensus.Block)
	for _, b := range blocks {
		if !b.Prefix.IsValid() {
			continue
		}
		byCountry[b.Country] = append(byCountry[b.Country], b)
	}
	if len(byCountry) == 0 {
		return fmt.Errorf("no valid consensus networks")
	}
	if err := writeGeoIPDat(filepath.Join(root, "geoip.dat"), byCountry); err != nil {
		return err
	}
	if err := writeMMDB(filepath.Join(root, "country.mmdb"), blocks); err != nil {
		return err
	}
	if err := writeRuleSets(filepath.Join(root, "sing-box"), byCountry, singBoxBinary); err != nil {
		return err
	}
	if err := writeCountryCSVs(filepath.Join(root, "countries"), byCountry); err != nil {
		return err
	}
	return nil
}

type sourceWriterState struct {
	file   *os.File
	writer *csv.Writer
	used   uint64
}

type SourceCSVWriter struct {
	root        string
	filter      *iprange.PrefixFilter
	writers     map[string]*sourceWriterState
	initialized map[string]bool
	tick        uint64
}

func NewSourceCSVWriter(root string, exclusions []netip.Prefix) (*SourceCSVWriter, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if filepath.Base(root) != "source-csv" {
		return nil, fmt.Errorf("refusing to reset unexpected source CSV directory %q", root)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &SourceCSVWriter{root: root, filter: iprange.NewPrefixFilter(exclusions), writers: make(map[string]*sourceWriterState), initialized: make(map[string]bool)}, nil
}

func (s *SourceCSVWriter) closeWriter(state *sourceWriterState) error {
	state.writer.Flush()
	err := state.writer.Error()
	closeErr := state.file.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func (s *SourceCSVWriter) getWriter(path string) (*sourceWriterState, error) {
	s.tick++
	if state := s.writers[path]; state != nil {
		state.used = s.tick
		return state, nil
	}
	const maxOpenWriters = 48
	if len(s.writers) >= maxOpenWriters {
		var oldestPath string
		oldest := ^uint64(0)
		for path, state := range s.writers {
			if state.used < oldest {
				oldestPath, oldest = path, state.used
			}
		}
		if err := s.closeWriter(s.writers[oldestPath]); err != nil {
			return nil, err
		}
		delete(s.writers, oldestPath)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if !s.initialized[path] {
		flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}
	file, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		return nil, err
	}
	state := &sourceWriterState{file: file, writer: csv.NewWriter(file), used: s.tick}
	s.writers[path] = state
	if !s.initialized[path] {
		if err := state.writer.Write([]string{"network", "country", "vote_group"}); err != nil {
			return nil, err
		}
		s.initialized[path] = true
	}
	return state, nil
}

func (s *SourceCSVWriter) WriteRecords(records []model.Record) error {
	for _, r := range records {
		for _, prefix := range s.filter.SubtractPrefix(r.Prefix) {
			path := filepath.Join(s.root, safe(r.SourceID), strings.ToLower(r.Country)+".csv")
			state, err := s.getWriter(path)
			if err != nil {
				return err
			}
			if err := state.writer.Write([]string{prefix.String(), r.Country, r.VoteGroup}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SourceCSVWriter) Close() error {
	var first error
	for _, state := range s.writers {
		if err := s.closeWriter(state); err != nil && first == nil {
			first = err
		}
	}
	s.writers = nil
	var stale []string
	err := filepath.WalkDir(s.root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".csv") && !s.initialized[path] {
			stale = append(stale, path)
		}
		return nil
	})
	if err != nil && first == nil {
		first = err
	}
	for _, path := range stale {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) && first == nil {
			first = err
		}
	}
	return first
}

func WriteSourceCSVs(root string, records []model.Record, exclusions []netip.Prefix) error {
	w, err := NewSourceCSVWriter(root, exclusions)
	if err != nil {
		return err
	}
	if err := w.WriteRecords(records); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func writeCountryCSVs(root string, byCountry map[string][]consensus.Block) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	codes := sortedKeys(byCountry)
	for _, cc := range codes {
		path := filepath.Join(root, strings.ToLower(cc)+".csv")
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		w := csv.NewWriter(f)
		if err = w.Write([]string{"network", "country", "confidence_pct", "winner_votes", "total_votes"}); err != nil {
			f.Close()
			return err
		}
		items := byCountry[cc]
		sort.Slice(items, func(i, j int) bool { return items[i].Prefix.String() < items[j].Prefix.String() })
		for _, b := range items {
			if err := w.Write([]string{b.Prefix.String(), cc, fmt.Sprintf("%.4f", b.Confidence), b.WinnerVotes.String(), b.TotalVotes.String()}); err != nil {
				f.Close()
				return err
			}
		}
		w.Flush()
		err = w.Error()
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func writeGeoIPDat(path string, byCountry map[string][]consensus.Block) error {
	codes := sortedKeys(byCountry)
	list := &v2ray.GeoIPList{Entry: make([]*v2ray.GeoIP, 0, len(codes))}
	for _, cc := range codes {
		entry := &v2ray.GeoIP{CountryCode: strings.ToLower(cc)}
		items := append([]consensus.Block(nil), byCountry[cc]...)
		sort.Slice(items, func(i, j int) bool { return items[i].Prefix.String() < items[j].Prefix.String() })
		for _, b := range items {
			a := b.Prefix.Addr()
			entry.Cidr = append(entry.Cidr, &v2ray.CIDR{Ip: a.AsSlice(), Prefix: uint32(b.Prefix.Bits())})
		}
		list.Entry = append(list.Entry, entry)
	}
	b, err := proto.Marshal(list)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func writeMMDB(path string, blocks []consensus.Block) error {
	tree, err := mmdbwriter.New(mmdbwriter.Options{DatabaseType: "GeoIP-Country", Description: map[string]string{"en": "Country-level IP geolocation generated from configured public sources"}, Languages: []string{"en"}, IPVersion: 6, RecordSize: 24, DisableIPv4Aliasing: true})
	if err != nil {
		return err
	}
	for _, b := range blocks {
		name := countries.ByName(b.Country).String()
		value := mmdbtype.Map{
			mmdbtype.String("country"): mmdbtype.Map{
				mmdbtype.String("iso_code"): mmdbtype.String(b.Country),
				mmdbtype.String("names"):    mmdbtype.Map{mmdbtype.String("en"): mmdbtype.String(name)},
			},
			mmdbtype.String("registered_country"): mmdbtype.Map{mmdbtype.String("iso_code"): mmdbtype.String(b.Country)},
		}
		if err := tree.Insert(netipx.PrefixIPNet(b.Prefix), value); err != nil {
			return fmt.Errorf("insert %s %s into MMDB: %w", b.Country, b.Prefix, err)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = tree.WriteTo(f)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func writeRuleSets(root string, byCountry map[string][]consensus.Block, binary string) error {
	if strings.TrimSpace(binary) == "" {
		binary = os.Getenv("SING_BOX_BIN")
	}
	if strings.TrimSpace(binary) == "" {
		binary = "sing-box"
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	for _, cc := range sortedKeys(byCountry) {
		cidrs := make([]string, 0, len(byCountry[cc]))
		for _, b := range byCountry[cc] {
			cidrs = append(cidrs, b.Prefix.String())
		}
		sort.Strings(cidrs)
		payload, err := json.MarshalIndent(RuleSet{Version: 1, Rules: []IPRule{{IPCIDR: cidrs}}}, "", "  ")
		if err != nil {
			return err
		}
		base := filepath.Join(root, "geoip-"+strings.ToLower(cc))
		jsonPath := base + ".json"
		srsPath := base + ".srs"
		if err := os.WriteFile(jsonPath, append(payload, '\n'), 0o644); err != nil {
			return err
		}
		cmd := exec.Command(binary, "rule-set", "compile", "--output", srsPath, jsonPath)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("compile sing-box rule set %s: %w: %s", cc, err, strings.TrimSpace(stderr.String()))
		}
	}
	return nil
}

func ZipDirectory(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(out)
	err = filepath.WalkDir(source, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, e := filepath.Rel(source, path)
		if e != nil {
			return e
		}
		info, e := d.Info()
		if e != nil {
			return e
		}
		h, e := zip.FileInfoHeader(info)
		if e != nil {
			return e
		}
		h.Name = filepath.ToSlash(rel)
		h.Method = zip.Deflate
		w, e := zw.CreateHeader(h)
		if e != nil {
			return e
		}
		f, e := os.Open(path)
		if e != nil {
			return e
		}
		_, copyErr := io.Copy(w, f)
		closeErr := f.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	closeZip := zw.Close()
	closeFile := out.Close()
	if err != nil {
		return err
	}
	if closeZip != nil {
		return closeZip
	}
	return closeFile
}

func sortedKeys(m map[string][]consensus.Block) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func safe(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, s)
}
