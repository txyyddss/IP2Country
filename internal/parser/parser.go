package parser

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/biter777/countries"
	regionxdb "github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/oschwald/maxminddb-golang/v2"
	"github.com/txyyddss/ip2country/internal/config"
	"github.com/txyyddss/ip2country/internal/iprange"
	"github.com/txyyddss/ip2country/internal/model"
	"github.com/v2fly/geoip/plugin/v2ray"
	"google.golang.org/protobuf/proto"
)

type Result struct {
	Records  []model.Record
	Filters  []netip.Prefix
	Parsed   int64
	Rejected int64
}

type Collector struct {
	Source   config.Source
	Result   Result
	Order    int
	rowOrder int
}

func ParseFile(s config.Source, path string, order int) (Result, error) {
	c := &Collector{Source: s, Order: order}
	var err error
	switch s.Format {
	case "csv", "tsv", "csv-zip":
		err = c.parseDelimitedPath(path, s.Format == "tsv" || s.Delimiter == "\\t", strings.HasSuffix(s.Format, "zip"))
	case "txt":
		err = c.parseTextPath(path)
	case "json":
		err = c.parseJSONPath(path)
	case "mmdb":
		err = c.parseMMDB(path)
	case "v2ray-dat":
		err = c.parseV2RayDat(path)
	case "sxgeo-zip":
		err = c.parseSxGeoZip(path)
	case "ip2region-xdb":
		err = c.parseIP2Region(path)
	default:
		return Result{}, fmt.Errorf("unsupported source format %q", s.Format)
	}
	if err != nil {
		return c.Result, err
	}
	if s.Role == "country" && len(c.Result.Records) == 0 {
		return c.Result, fmt.Errorf("source produced no country/IP records")
	}
	return c.Result, nil
}

func (c *Collector) add(prefix netip.Prefix, country string) {
	country = normalizeCountry(country)
	if !prefix.IsValid() || country == "" {
		c.Result.Rejected++
		return
	}
	prefix = prefix.Masked()
	if c.Source.Family == "ipv4" && !prefix.Addr().Is4() || c.Source.Family == "ipv6" && !prefix.Addr().Is6() {
		c.Result.Rejected++
		return
	}
	c.Result.Records = append(c.Result.Records, model.Record{
		Prefix: prefix, Country: country, SourceID: c.Source.ID,
		VoteGroup: c.Source.VoteGroup, Order: c.Order, RowOrder: c.rowOrder,
	})
	c.Result.Parsed++
	c.rowOrder++
}

func normalizeCountry(raw string) string {
	s := strings.ToUpper(strings.TrimSpace(raw))
	if len(s) == 2 && s[0] >= 'A' && s[0] <= 'Z' && s[1] >= 'A' && s[1] <= 'Z' {
		cc := countries.ByName(s)
		if cc.Alpha2() == s {
			return s
		}
	}
	cc := countries.ByName(strings.TrimSpace(raw))
	code := cc.Alpha2()
	if len(code) == 2 && code != "--" && code != "ZZ" {
		return code
	}
	return ""
}

func openMaybeZip(path, member string, requireZip bool, fn func(io.Reader) error) error {
	if !requireZip {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		return fn(f)
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if member != "" && filepath.Base(f.Name) != filepath.Base(member) {
			continue
		}
		if member == "" && !strings.HasSuffix(strings.ToLower(f.Name), ".csv") {
			continue
		}
		r, err := f.Open()
		if err != nil {
			return err
		}
		err = fn(r)
		_ = r.Close()
		return err
	}
	return fmt.Errorf("archive %s has no matching member %q", path, member)
}

func (c *Collector) parseDelimitedPath(path string, tab, zipped bool) error {
	return openMaybeZip(path, c.Source.ArchiveMember, zipped, func(r io.Reader) error {
		cr := csv.NewReader(bufio.NewReaderSize(r, 256*1024))
		if tab {
			cr.Comma = '\t'
		} else if c.Source.Delimiter != "" && c.Source.Delimiter != "\\t" {
			cr.Comma = []rune(c.Source.Delimiter)[0]
		}
		cr.FieldsPerRecord = -1
		cr.ReuseRecord = true
		first, err := cr.Read()
		if err != nil {
			return err
		}
		idx := detectColumns(first, c.Source)
		header := idx.header
		if c.Source.Header != nil {
			header = *c.Source.Header
		}
		if !header {
			if err := c.parseCSVRow(first, idx); err != nil {
				c.Result.Rejected++
			}
		}
		for {
			row, err := cr.Read()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				c.Result.Rejected++
				continue
			}
			if rowBlank(row) {
				continue
			}
			if err := c.parseCSVRow(row, idx); err != nil {
				c.Result.Rejected++
			}
		}
		return nil
	})
}

type columns struct {
	header                    bool
	cidr, start, end, country int
}

func detectColumns(row []string, s config.Source) columns {
	x := columns{cidr: -1, start: -1, end: -1, country: -1}
	for i, raw := range row {
		k := normalizeHeader(raw)
		switch k {
		case "cidr", "network", "iprange", "ipblock", "prefix", "ipnetwork":
			x.cidr = i
		case "startip", "ipstart", "ipfrom", "start", "from", "ipfrominteger", "ipstartinteger":
			x.start = i
		case "endip", "ipend", "ipto", "end", "to", "iptointeger", "ipendinteger":
			x.end = i
		case "country", "countrycode", "countryiso", "countryisocode", "iso", "isocode", "cc", "iso3166":
			x.country = i
		}
	}
	if x.cidr >= 0 || x.start >= 0 || x.country >= 0 {
		x.header = true
	}
	if v, ok := columnIndex(s.CIDRColumn, row); ok {
		x.cidr = v
	}
	if v, ok := columnIndex(s.StartColumn, row); ok {
		x.start = v
	}
	if v, ok := columnIndex(s.EndColumn, row); ok {
		x.end = v
	}
	if v, ok := columnIndex(s.CountryColumn, row); ok {
		x.country = v
	}
	if x.cidr < 0 && x.start < 0 {
		x.cidr = 0
	}
	if x.start < 0 && x.cidr == 0 && len(row) >= 3 && parseIPToken(row[0], s.Family).IsValid() && parseIPToken(row[1], s.Family).IsValid() {
		x.cidr, x.start, x.end = -1, 0, 1
	}
	if x.country < 0 && len(row) >= 2 {
		x.country = len(row) - 1
	}
	return x
}

func normalizeHeader(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func columnIndex(name string, row []string) (int, bool) {
	if name == "" {
		return 0, false
	}
	if n, err := strconv.Atoi(name); err == nil && n >= 0 && n < len(row) {
		return n, true
	}
	want := normalizeHeader(name)
	for i, cell := range row {
		if normalizeHeader(cell) == want {
			return i, true
		}
	}
	return 0, false
}

func (c *Collector) parseCSVRow(row []string, idx columns) error {
	get := func(i int) string {
		if i >= 0 && i < len(row) {
			return strings.TrimSpace(row[i])
		}
		return ""
	}
	country := c.Source.CountryOverride
	if country == "" {
		country = get(idx.country)
	}
	var prefixes []netip.Prefix
	if idx.cidr >= 0 && get(idx.cidr) != "" {
		p, err := parsePrefixOrAddress(get(idx.cidr), c.Source.Family)
		if err == nil {
			prefixes = []netip.Prefix{p}
		}
	} else {
		start, end := parseIPToken(get(idx.start), c.Source.Family), parseIPToken(get(idx.end), c.Source.Family)
		if start.IsValid() && end.IsValid() && start.Is4() == end.Is4() {
			prefixes = iprange.RangeToPrefixes(start, end)
		}
	}
	if len(prefixes) == 0 {
		// Some feeds put the prefix and ISO country in columns that have no header.
		for _, value := range row {
			if p, err := parsePrefixOrAddress(strings.TrimSpace(value), c.Source.Family); err == nil {
				prefixes = []netip.Prefix{p}
				break
			}
		}
		if country == "" {
			for _, value := range row {
				if normalizeCountry(value) != "" {
					country = value
					break
				}
			}
		}
	}
	if country == "" || len(prefixes) == 0 {
		return fmt.Errorf("row has no recognized network or country")
	}
	for _, p := range prefixes {
		c.add(p, country)
	}
	return nil
}

func parseIPToken(s, family string) netip.Addr {
	s = strings.TrimSpace(strings.Trim(s, "\"'"))
	if a, err := netip.ParseAddr(s); err == nil {
		return a.Unmap()
	}
	if family == "ipv6" {
		value := strings.ToLower(strings.TrimPrefix(s, "0x"))
		base := 10
		if len(value) == 32 || strings.ContainsAny(value, "abcdef") {
			base = 16
		}
		if n, ok := new(big.Int).SetString(value, base); ok && n.Sign() >= 0 && n.BitLen() <= 128 {
			var raw [16]byte
			b := n.Bytes()
			copy(raw[16-len(b):], b)
			return netip.AddrFrom16(raw)
		}
	}
	if family == "ipv4" {
		if n, ok := new(big.Int).SetString(s, 10); ok && n.Sign() >= 0 && n.BitLen() <= 32 {
			var raw [4]byte
			b := n.Bytes()
			copy(raw[4-len(b):], b)
			return netip.AddrFrom4(raw)
		}
	}
	return netip.Addr{}
}

func parsePrefixOrAddress(s, family string) (netip.Prefix, error) {
	s = strings.TrimSpace(strings.Trim(s, "\"'"))
	if p, err := netip.ParsePrefix(s); err == nil {
		return p.Masked(), nil
	}
	if a := parseIPToken(s, family); a.IsValid() {
		return netip.PrefixFrom(a, a.BitLen()), nil
	}
	return netip.Prefix{}, fmt.Errorf("invalid IP network %q", s)
}

func rowBlank(row []string) bool {
	for _, v := range row {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}

func (c *Collector) parseTextPath(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var pattern *regexp.Regexp
	if c.Source.TextRegex != "" {
		pattern, err = regexp.Compile(c.Source.TextRegex)
		if err != nil {
			return err
		}
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		country := c.Source.CountryOverride
		var network string
		if pattern != nil {
			m := pattern.FindStringSubmatch(line)
			if m == nil {
				c.Result.Rejected++
				continue
			}
			for i, name := range pattern.SubexpNames() {
				if name == "cidr" {
					network = m[i]
				}
				if name == "country" && country == "" {
					country = m[i]
				}
			}
		} else {
			fields := strings.Fields(strings.SplitN(line, "#", 2)[0])
			if len(fields) == 0 {
				continue
			}
			network = fields[0]
			if country == "" {
				for _, f := range fields[1:] {
					if normalizeCountry(f) != "" {
						country = f
						break
					}
				}
			}
		}
		p, e := parsePrefixOrAddress(network, c.Source.Family)
		if e != nil || country == "" {
			c.Result.Rejected++
			continue
		}
		c.add(p, country)
	}
	return sc.Err()
}

func (c *Collector) parseJSONPath(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !json.Valid(b) {
		b = repairAdjacentArrayObjects(b)
		if !json.Valid(b) {
			return fmt.Errorf("invalid JSON source")
		}
	}
	trimmed := bytes.TrimSpace(b)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		decoder := json.NewDecoder(bytes.NewReader(b))
		decoder.UseNumber()
		if _, err := decoder.Token(); err != nil {
			return err
		}
		for decoder.More() {
			var item any
			if err := decoder.Decode(&item); err != nil {
				return err
			}
			if err := c.walkJSON(item, ""); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	}
	var value any
	if err := json.Unmarshal(b, &value); err != nil {
		return err
	}
	return c.walkJSON(value, "")
}

func repairAdjacentArrayObjects(input []byte) []byte {
	var out bytes.Buffer
	out.Grow(len(input) + len(input)/32)
	arrayDepth, objectDepth := 0, 0
	inString, escaped, closedObject := false, false, false
	for _, ch := range input {
		if inString {
			out.WriteByte(ch)
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			closedObject = false
			out.WriteByte(ch)
			continue
		}
		if ch == '{' && arrayDepth > 0 && objectDepth == 0 && closedObject {
			out.WriteByte(',')
			closedObject = false
		}
		out.WriteByte(ch)
		switch ch {
		case '[':
			arrayDepth++
			closedObject = false
		case ']':
			arrayDepth--
			closedObject = false
		case '{':
			objectDepth++
			closedObject = false
		case '}':
			objectDepth--
			if arrayDepth > 0 && objectDepth == 0 {
				closedObject = true
			}
		case ',':
			closedObject = false
		default:
			if ch != ' ' && ch != '\n' && ch != '\r' && ch != '\t' {
				closedObject = false
			}
		}
	}
	return out.Bytes()
}

func (c *Collector) walkJSON(v any, inherited string) error {
	switch x := v.(type) {
	case []any:
		for _, item := range x {
			if err := c.walkJSON(item, inherited); err != nil {
				return err
			}
		}
	case map[string]any:
		country := c.Source.CountryOverride
		if country == "" {
			country = inherited
		}
		var network, prefixLength string
		var start, end string
		for k, v := range x {
			key := normalizeHeader(k)
			value := fmt.Sprint(v)
			switch key {
			case "cidr", "network", "iprange", "ipblock", "prefix", "range", "ip", "address":
				if network == "" {
					network = value
				}
			case "prefixlength", "prefixlen", "mask":
				prefixLength = value
			case "startip", "ipstart", "ipfrom", "start", "from":
				start = value
			case "endip", "ipend", "ipto", "end", "to":
				end = value
			case "country", "countrycode", "countryisocode", "isocode", "iso3166", "cc":
				if country == "" {
					country = value
				}
			}
		}
		if network != "" && prefixLength != "" && !strings.Contains(network, "/") {
			network += "/" + prefixLength
		}
		added := false
		if network != "" && country != "" {
			if p, e := parsePrefixOrAddress(network, c.Source.Family); e == nil {
				c.add(p, country)
				added = true
			}
		}
		if start != "" && end != "" && country != "" {
			a, b := parseIPToken(start, c.Source.Family), parseIPToken(end, c.Source.Family)
			if a.IsValid() && b.IsValid() {
				for _, p := range iprange.RangeToPrefixes(a, b) {
					c.add(p, country)
					added = true
				}
			}
		}
		if !added && c.Source.Role == "filter" {
			c.parseFilterObject(x)
		}
		for k, child := range x {
			next := inherited
			if normalizeCountry(k) != "" {
				next = normalizeCountry(k)
			}
			if err := c.walkJSON(child, next); err != nil {
				return err
			}
		}
	case string:
		_ = x
	}
	return nil
}

func (c *Collector) parseFilterObject(x map[string]any) {
	globally, hasGlobal := x["isGloballyReachable"].(bool)
	if !hasGlobal {
		globally, hasGlobal = x["globally_reachable"].(bool)
	}
	special, hasSpecial := x["special_use"].(bool)
	if !hasSpecial {
		special, hasSpecial = x["is_special_use"].(bool)
	}
	if !((hasGlobal && !globally) || (hasSpecial && special)) {
		return
	}
	for k, v := range x {
		key := normalizeHeader(k)
		if key == "prefix" || key == "network" || key == "cidr" || key == "addressblock" {
			if p, e := parsePrefixOrAddress(fmt.Sprint(v), ""); e == nil {
				c.Result.Filters = append(c.Result.Filters, p)
			}
		}
	}
}

func (c *Collector) parseMMDB(path string) error {
	reader, err := maxminddb.Open(path)
	if err != nil {
		return err
	}
	defer reader.Close()
	for result := range reader.Networks() {
		if err := result.Err(); err != nil {
			c.Result.Rejected++
			continue
		}
		var rec map[string]any
		if err := result.Decode(&rec); err != nil {
			c.Result.Rejected++
			continue
		}
		country := c.Source.CountryOverride
		if country == "" {
			country = findCountryValue(rec)
		}
		if country == "" {
			c.Result.Rejected++
			continue
		}
		c.add(result.Prefix(), country)
	}
	if len(c.Result.Records) == 0 {
		return fmt.Errorf("MMDB has no country-coded networks")
	}
	return nil
}

func findCountryValue(m map[string]any) string {
	for _, path := range [][]string{{"country", "iso_code"}, {"registered_country", "iso_code"}, {"country_code"}, {"countryCode"}, {"iso_code"}, {"isoCode"}, {"country"}, {"country_iso_code"}} {
		var current any = m
		ok := true
		for _, key := range path {
			obj, yes := current.(map[string]any)
			if !yes {
				ok = false
				break
			}
			current = obj[key]
			if current == nil {
				ok = false
				break
			}
		}
		if ok {
			if s, yes := current.(string); yes && normalizeCountry(s) != "" {
				return s
			}
		}
	}
	return ""
}

func (c *Collector) parseV2RayDat(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var list v2ray.GeoIPList
	if err := proto.Unmarshal(b, &list); err != nil {
		return fmt.Errorf("decode V2Fly GeoIP protobuf: %w", err)
	}
	for _, entry := range list.GetEntry() {
		country := c.Source.CountryOverride
		if country == "" {
			country = entry.GetCountryCode()
		}
		for _, cidr := range entry.GetCidr() {
			addr, ok := netip.AddrFromSlice(cidr.GetIp())
			if !ok {
				c.Result.Rejected++
				continue
			}
			if addr.Is4In6() {
				addr = addr.Unmap()
			}
			if int(cidr.GetPrefix()) > addr.BitLen() {
				c.Result.Rejected++
				continue
			}
			c.add(netip.PrefixFrom(addr, int(cidr.GetPrefix())).Masked(), country)
		}
	}
	if len(c.Result.Records) == 0 {
		return fmt.Errorf("V2Fly dat has no recognized country prefixes")
	}
	return nil
}

func (c *Collector) parseSxGeoZip(path string) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".dat") {
			r, e := f.Open()
			if e != nil {
				return e
			}
			data, e := io.ReadAll(r)
			_ = r.Close()
			if e != nil {
				return e
			}
			return c.parseSxGeo(data)
		}
	}
	return fmt.Errorf("SYPEX archive contains no .dat file")
}

func (c *Collector) parseSxGeo(data []byte) (err error) {
	if len(data) < 40 || string(data[:3]) != "SxG" || data[3] < 21 {
		return fmt.Errorf("unsupported SYPEX header")
	}
	typeID := data[8]
	if typeID != 1 && typeID != 2 {
		return fmt.Errorf("SYPEX file is not a country database (type %d)", typeID)
	}
	blen := uint64(data[10])
	mlen := uint64(binary.BigEndian.Uint16(data[11:13]))
	blocks := uint64(binary.BigEndian.Uint16(data[13:15]))
	items := uint64(binary.BigEndian.Uint32(data[15:19]))
	idlen := uint64(data[19])
	pack := uint64(binary.BigEndian.Uint16(data[38:40]))
	if idlen == 0 || items == 0 || blocks == 0 {
		return fmt.Errorf("invalid SYPEX index header")
	}
	bstart := uint64(40) + pack
	mstart := bstart + blen*4
	dbstart := mstart + mlen*4
	dbend := dbstart + items*(3+idlen)
	if dbend > uint64(len(data)) || bstart+blen*4 > uint64(len(data)) {
		return fmt.Errorf("truncated SYPEX database")
	}
	bindex := make([]uint32, blen)
	for i := range bindex {
		bindex[i] = binary.BigEndian.Uint32(data[bstart+uint64(i*4):])
	}
	starts := make([]netip.Addr, 0, items)
	ids := make([]uint64, 0, items)
	for first := uint64(0); first < blen; first++ {
		lo := uint64(bindex[first])
		hi := items
		if first+1 < uint64(len(bindex)) {
			hi = uint64(bindex[first+1])
		}
		if lo > items {
			continue
		}
		if hi > items {
			hi = items
		}
		if hi < lo {
			continue
		}
		for idx := lo; idx < hi; idx++ {
			off := dbstart + idx*(3+idlen)
			low := uint32(data[off])<<16 | uint32(data[off+1])<<8 | uint32(data[off+2])
			ip := uint32(first)<<24 | low
			var raw [4]byte
			binary.BigEndian.PutUint32(raw[:], ip)
			starts = append(starts, netip.AddrFrom4(raw))
			id := uint64(0)
			for j := uint64(0); j < idlen; j++ {
				id |= uint64(data[off+3+j]) << (8 * j)
			}
			ids = append(ids, id)
		}
	}
	if len(starts) < 1 {
		return fmt.Errorf("SYPEX has no IP ranges")
	}
	for i, start := range starts {
		end := netip.AddrFrom4([4]byte{255, 255, 255, 255})
		if i+1 < len(starts) {
			end = decrementAddr(starts[i+1])
		}
		id := ids[i]
		if id == 0 || id >= uint64(len(sxGeoCountryCodes)) {
			continue
		}
		country := sxGeoCountryCodes[id]
		if country == "" {
			continue
		}
		for _, p := range iprange.RangeToPrefixes(start, end) {
			c.add(p, country)
		}
	}
	return nil
}

func decrementAddr(a netip.Addr) netip.Addr {
	if a.Is4() {
		b := a.As4()
		for i := 3; i >= 0; i-- {
			if b[i] > 0 {
				b[i]--
				break
			}
			b[i] = 255
		}
		return netip.AddrFrom4(b)
	}
	b := a.As16()
	for i := 15; i >= 0; i-- {
		if b[i] > 0 {
			b[i]--
			break
		}
		b[i] = 255
	}
	return netip.AddrFrom16(b)
}

func (c *Collector) parseIP2Region(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return err
	}
	header, err := regionxdb.LoadHeader(f)
	if err != nil {
		return fmt.Errorf("read IP2Region header: %w", err)
	}
	version, err := regionxdb.VersionFromHeader(header)
	if err != nil {
		return fmt.Errorf("read IP2Region version: %w", err)
	}
	start, end := uint64(header.StartIndexPtr), uint64(header.EndIndexPtr)
	indexSize := uint64(version.SegmentIndexSize)
	if start == 0 || end < start || end+indexSize > uint64(st.Size()) || (end-start)%indexSize != 0 {
		return fmt.Errorf("invalid IP2Region segment index bounds %d..%d (file %d)", start, end, st.Size())
	}
	addrLen := version.Bytes
	buf := make([]byte, indexSize)
	valid := 0
	for off := start; off <= end; off += indexSize {
		if _, err := f.ReadAt(buf, int64(off)); err != nil {
			return err
		}
		startBytes := append([]byte(nil), buf[:addrLen]...)
		endBytes := append([]byte(nil), buf[addrLen:addrLen*2]...)
		if version.Id == regionxdb.IPv4VersionNo {
			reverse(startBytes)
			reverse(endBytes)
		}
		startIP, ok := netip.AddrFromSlice(startBytes)
		if !ok {
			continue
		}
		endIP, ok := netip.AddrFromSlice(endBytes)
		if !ok || startIP.Compare(endIP) > 0 {
			continue
		}
		dataLen := int(binary.LittleEndian.Uint16(buf[addrLen*2 : addrLen*2+2]))
		dataPtr := binary.LittleEndian.Uint32(buf[addrLen*2+2:])
		if dataLen == 0 || uint64(dataPtr)+uint64(dataLen) > uint64(st.Size()) {
			continue
		}
		region := make([]byte, dataLen)
		if _, err := f.ReadAt(region, int64(dataPtr)); err != nil {
			continue
		}
		country := ""
		parts := strings.Split(strings.TrimSpace(string(region)), "|")
		for i := len(parts) - 1; i >= 0; i-- {
			if normalizeCountry(parts[i]) != "" {
				country = parts[i]
				break
			}
		}
		if country == "" {
			continue
		}
		for _, p := range iprange.RangeToPrefixes(startIP, endIP) {
			c.add(p, country)
		}
		valid++
	}
	if valid == 0 {
		return fmt.Errorf("no country-coded segments found in IP2Region xdb")
	}
	return nil
}

func reverse(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}

func (c *Collector) validateReader(path string) error { _, err := os.Stat(path); return err }

var _ = bytes.Compare
