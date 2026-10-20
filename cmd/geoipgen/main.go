package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/txyyddss/ip2country/internal/config"
	"github.com/txyyddss/ip2country/internal/consensus"
	"github.com/txyyddss/ip2country/internal/fetch"
	"github.com/txyyddss/ip2country/internal/model"
	"github.com/txyyddss/ip2country/internal/output"
	"github.com/txyyddss/ip2country/internal/parser"
	"github.com/txyyddss/ip2country/internal/report"
)

func main() {
	configPath := flag.String("config", "config/sources.json", "source configuration file")
	workDir := flag.String("work-dir", ".cache/geoip", "download cache and temporary work directory")
	outDir := flag.String("output-dir", "dist", "generated output directory")
	readmePath := flag.String("readme", "README.md", "README file updated with run statistics")
	singBoxBinary := flag.String("sing-box-binary", "", "path to pinned sing-box compiler (or SING_BOX_BIN)")
	offline := flag.Bool("offline", false, "use only last-known-good cached source files")
	flag.Parse()
	if err := build(*configPath, *workDir, *outDir, *readmePath, *singBoxBinary, *offline); err != nil {
		fmt.Fprintln(os.Stderr, "geoipgen:", err)
		os.Exit(1)
	}
}

func build(configPath, workDir, outDir, readmePath, singBoxBinary string, offline bool) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	filters := consensus.BuiltinSpecialUse()
	statuses := make([]model.SourceStatus, len(cfg.Sources))
	groupOrder := make(map[string]int)
	groupIDs := map[string]uint16{"": 0}
	groupNames := []string{""}
	for order, source := range cfg.Sources {
		if source.Role != "country" || source.Enabled == nil || !*source.Enabled {
			continue
		}
		if _, ok := groupIDs[source.VoteGroup]; !ok {
			groupIDs[source.VoteGroup] = uint16(len(groupNames))
			groupNames = append(groupNames, source.VoteGroup)
			groupOrder[source.VoteGroup] = order
		}
	}
	ctx := context.Background()

	// Load filter feeds first so their exclusions apply to every per-source CSV.
	for order, source := range cfg.Sources {
		if source.Role != "filter" {
			continue
		}
		status, parsed := loadSource(ctx, source, order, filepath.Join(workDir, "cache"), offline)
		statuses[order] = status
		filters = append(filters, parsed.Filters...)
		logSourceStatus(order+1, len(cfg.Sources), status)
	}

	sourceCSV, err := output.NewSourceCSVWriter(filepath.Join(outDir, "source-csv"), filters)
	if err != nil {
		return err
	}
	defer sourceCSV.Close()
	spoolPath := filepath.Join(workDir, "records.spool.gz")
	spool, err := consensus.NewRecordSpool(spoolPath, groupIDs)
	if err != nil {
		return err
	}
	spoolClosed := false
	defer func() {
		if !spoolClosed {
			_ = spool.Close()
		}
	}()
	var inputRecords int64
	for order, source := range cfg.Sources {
		if source.Role != "country" {
			continue
		}
		status, parsed := loadSource(ctx, source, order, filepath.Join(workDir, "cache"), offline)
		statuses[order] = status
		logSourceStatus(order+1, len(cfg.Sources), status)
		if len(parsed.Records) == 0 {
			continue
		}
		if err := sourceCSV.WriteRecords(parsed.Records); err != nil {
			return fmt.Errorf("write per-source CSVs: %w", err)
		}
		for _, record := range parsed.Records {
			if err := spool.Write(record); err != nil {
				return err
			}
			inputRecords++
		}
		parsed = parser.Result{}
		runtime.GC()
	}
	if err := sourceCSV.Close(); err != nil {
		return err
	}
	if err := spool.Close(); err != nil {
		return err
	}
	spoolClosed = true
	if inputRecords == 0 {
		summary := report.BuildSummary(statuses, nil, 0)
		_ = report.WriteJSON(filepath.Join(outDir, "run-summary.json"), summary)
		_ = report.WriteJSON(filepath.Join(outDir, "manifest.json"), summary)
		_ = report.UpdateREADME(readmePath, summary)
		return fmt.Errorf("no country source produced usable IP data; see run-summary.json")
	}
	blocks, err := consensus.ResolveSpool(spoolPath, groupNames, groupOrder, filters)
	_ = os.Remove(spoolPath)
	if err != nil {
		return err
	}
	if len(blocks) == 0 {
		return fmt.Errorf("all consensus results were excluded by special-use filtering")
	}
	if err := output.Generate(outDir, blocks, singBoxBinary); err != nil {
		return err
	}
	if err := output.ZipDirectory(filepath.Join(outDir, "countries"), filepath.Join(outDir, "country-csv.zip")); err != nil {
		return err
	}
	if err := output.ZipDirectory(filepath.Join(outDir, "source-csv"), filepath.Join(outDir, "source-csv.zip")); err != nil {
		return err
	}
	if err := output.ZipDirectory(filepath.Join(outDir, "sing-box"), filepath.Join(outDir, "sing-box-rulesets.zip")); err != nil {
		return err
	}

	summary := report.BuildSummary(statuses, blocks, inputRecords)
	if err := report.WriteJSON(filepath.Join(outDir, "run-summary.json"), summary); err != nil {
		return err
	}
	if err := report.WriteJSON(filepath.Join(outDir, "manifest.json"), summary); err != nil {
		return err
	}
	if err := report.UpdateREADME(readmePath, summary); err != nil {
		return err
	}
	if err := writeChecksums(outDir); err != nil {
		return err
	}
	fmt.Printf("Generated %d country blocks from %d source rows across %d sources\n", len(blocks), inputRecords, len(statuses))
	return nil
}

func loadSource(ctx context.Context, source config.Source, order int, cacheDir string, offline bool) (model.SourceStatus, parser.Result) {
	status := model.SourceStatus{ID: source.ID, Name: source.Name, Status: "skipped"}
	if source.Enabled != nil && !*source.Enabled {
		status.Status = "disabled"
		return status, parser.Result{}
	}
	candidates, downloadErrs := fetch.Candidates(ctx, source, cacheDir, offline)
	var parseErrors []string
	var selected *fetch.Candidate
	var parsed parser.Result
	for i := range candidates {
		candidate := &candidates[i]
		result, err := parser.ParseFile(source, candidate.Path, order)
		if err != nil {
			parseErrors = append(parseErrors, err.Error())
			fetch.RemoveCandidate(*candidate)
			continue
		}
		parsed = result
		selected = candidate
		break
	}
	if selected == nil {
		status.Status = "skipped-no-usable-copy"
		var messages []string
		for _, err := range downloadErrs {
			messages = append(messages, err.Error())
		}
		messages = append(messages, parseErrors...)
		if len(messages) > 0 {
			status.Error = compact(strings.Join(messages, "; "), 500)
		}
		for _, candidate := range candidates {
			fetch.RemoveCandidate(candidate)
		}
		return status, parser.Result{}
	}
	status.URL = selected.URL
	status.Bytes = selected.Bytes
	status.ParsedRows = parsed.Parsed
	status.RejectedRows = parsed.Rejected
	status.NoCountryRows = parsed.NoCountry
	if selected.Fresh {
		if _, err := fetch.Commit(*selected, source, cacheDir); err != nil {
			status.Status = "fresh-not-cached"
			parseErrors = append(parseErrors, "cache replacement: "+err.Error())
		} else if len(source.URLs) > 0 && selected.URL == source.URLs[0] {
			status.Status = "fresh"
		} else {
			status.Status = "fallback"
		}
	} else {
		status.Status = "cached"
	}
	for _, record := range parsed.Records {
		if record.Prefix.Addr().Is4() {
			status.IPv4Records++
		} else {
			status.IPv6Records++
		}
	}
	var messages []string
	for _, err := range downloadErrs {
		messages = append(messages, err.Error())
	}
	messages = append(messages, parseErrors...)
	if len(messages) > 0 {
		status.Error = compact(strings.Join(messages, "; "), 500)
	}
	for _, candidate := range candidates {
		if selected.Temp == "" || candidate.Temp != selected.Temp {
			fetch.RemoveCandidate(candidate)
		}
	}
	return status, parsed
}

func writeChecksums(root string) error {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Base(path) == "SHA256SUMS" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(paths)
	var b strings.Builder
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		h := sha256.New()
		_, copyErr := io.Copy(h, f)
		closeErr := f.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		sum := hex.EncodeToString(h.Sum(nil))
		rel, _ := filepath.Rel(root, path)
		fmt.Fprintf(&b, "%s  %s\n", sum, filepath.ToSlash(rel))
	}
	return os.WriteFile(filepath.Join(root, "SHA256SUMS"), []byte(b.String()), 0o644)
}

func compact(s string, limit int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > limit {
		return s[:limit-3] + "..."
	}
	return s
}

func logSourceStatus(current, total int, status model.SourceStatus) {
	fmt.Printf("[%02d/%02d] %-28s %-22s rows=%d v4=%d v6=%d bytes=%d\n", current, total, status.ID, status.Status, status.ParsedRows, status.IPv4Records, status.IPv6Records, status.Bytes)
}
