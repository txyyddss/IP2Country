package report

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/txyyddss/ip2country/internal/consensus"
	"github.com/txyyddss/ip2country/internal/model"
)

const startMarker = "<!-- GEOIP-STATISTICS:START -->"
const endMarker = "<!-- GEOIP-STATISTICS:END -->"

func BuildSummary(sources []model.SourceStatus, blocks []consensus.Block, inputRecords int64) model.RunSummary {
	summary := model.RunSummary{GeneratedAt: timeNowUTC(), Sources: sources, InputRecords: inputRecords, OutputBlocks: int64(len(blocks))}
	stats := make(map[string]*model.CountryStats)
	weight := make(map[string]*big.Rat)
	confidence := make(map[string]*big.Rat)
	for _, b := range blocks {
		if !b.Prefix.IsValid() {
			continue
		}
		s := stats[b.Country]
		if s == nil {
			s = &model.CountryStats{Country: b.Country}
			stats[b.Country] = s
			weight[b.Country] = new(big.Rat)
			confidence[b.Country] = new(big.Rat)
		}
		unitPrefix := 24
		if b.Prefix.Addr().Is6() {
			unitPrefix = 64
			s.IPv6Blocks++
		} else {
			s.IPv4Blocks++
		}
		cells := cellWeight(b.Prefix, unitPrefix)
		if unitPrefix == 24 {
			s.IPv4Cells = addDecimal(s.IPv4Cells, cells)
		} else {
			s.IPv6Cells = addDecimal(s.IPv6Cells, cells)
		}
		addresses := addressWeight(b.Prefix)
		conf := new(big.Rat).SetFloat64(b.Confidence / 100)
		if conf != nil {
			confidence[b.Country].Add(confidence[b.Country], new(big.Rat).Mul(addresses, conf))
			weight[b.Country].Add(weight[b.Country], addresses)
		}
	}
	for cc, s := range stats {
		if weight[cc].Sign() > 0 {
			avg := new(big.Rat).Quo(confidence[cc], weight[cc])
			f, _ := avg.Float64()
			s.AverageConf = fmt.Sprintf("%.2f%%", f*100)
		}
		summary.Countries = append(summary.Countries, *s)
	}
	sort.Slice(summary.Countries, func(i, j int) bool { return summary.Countries[i].Country < summary.Countries[j].Country })
	return summary
}

func WriteJSON(path string, summary model.RunSummary) error {
	b, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func UpdateREADME(path string, summary model.RunSummary) error {
	body := render(summary)
	old, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(old)
	start := strings.Index(content, startMarker)
	end := strings.Index(content, endMarker)
	section := startMarker + "\n" + body + "\n" + endMarker
	if start >= 0 && end >= start {
		content = content[:start] + section + content[end+len(endMarker):]
	} else {
		content = strings.TrimRight(content, "\r\n") + "\n\n" + section + "\n"
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func render(summary model.RunSummary) string {
	var b strings.Builder
	b.WriteString("## Generated data status\n\n")
	fmt.Fprintf(&b, "Last run: `%s`  \nInput rows: **%d** · Published blocks: **%d**\n\n", summary.GeneratedAt, summary.InputRecords, summary.OutputBlocks)
	b.WriteString("### Source availability\n\n| Source | Status | IPv4 rows | IPv6 rows | Rejected | Bytes | Details |\n|---|---:|---:|---:|---:|---:|---|\n")
	for _, s := range summary.Sources {
		details := s.URL
		if details == "" {
			details = "—"
		}
		if s.Error != "" {
			details = s.Error
		}
		details = strings.ReplaceAll(details, "|", "\\|")
		fmt.Fprintf(&b, "| %s | %s | %d | %d | %d | %d | %s |\n", s.Name, s.Status, s.IPv4Records, s.IPv6Records, s.RejectedRows, s.Bytes, details)
	}
	b.WriteString("\n### Final country coverage\n\n| Country | IPv4 blocks | IPv4 /24 units | IPv6 blocks | IPv6 /64 units | Average confidence |\n|---|---:|---:|---:|---:|---:|\n")
	for _, c := range summary.Countries {
		fmt.Fprintf(&b, "| %s | %d | %s | %d | %s | %s |\n", c.Country, c.IPv4Blocks, c.IPv4Cells, c.IPv6Blocks, c.IPv6Cells, c.AverageConf)
	}
	if len(summary.Sources) == 0 {
		b.WriteString("\nNo source results were recorded.\n")
	}
	return b.String()
}

func cellWeight(p netip.Prefix, unit int) *big.Rat {
	exponent := unit - p.Bits()
	if exponent >= 0 {
		return new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(exponent)))
	}
	return new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), uint(-exponent)))
}

func addressWeight(p netip.Prefix) *big.Rat {
	return new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(p.Addr().BitLen()-p.Bits())))
}

func addDecimal(current string, value *big.Rat) string {
	base := new(big.Rat)
	if current != "" {
		if _, ok := base.SetString(current); !ok {
			base.SetInt64(0)
		}
	}
	base.Add(base, value)
	return base.FloatString(4)
}

func timeNowUTC() string { return time.Now().UTC().Format(time.RFC3339) }
