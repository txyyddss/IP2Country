package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/txyyddss/ip2country/internal/config"
)

const maxDownloadBytes = 768 << 20

type Candidate struct {
	Path  string
	URL   string
	Fresh bool
	Bytes int64
	Temp  string
}

func Candidates(ctx context.Context, s config.Source, cacheDir string, offline bool) ([]Candidate, []error) {
	_ = os.MkdirAll(cacheDir, 0o755)
	urls := append(append([]string{}, s.URLs...), s.FallbackURLs...)
	var out []Candidate
	var errs []error
	client := &http.Client{Timeout: 2 * time.Minute}
	if !offline {
		for _, url := range urls {
			if strings.TrimSpace(url) == "" {
				continue
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			req.Header.Set("User-Agent", "ip2country-geoip-generator/1.0")
			resp, err := client.Do(req)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", url, err))
				continue
			}
			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				_ = resp.Body.Close()
				errs = append(errs, fmt.Errorf("%s: HTTP %s", url, resp.Status))
				continue
			}
			tmp, err := os.CreateTemp(cacheDir, s.ID+"-download-*.tmp")
			if err != nil {
				_ = resp.Body.Close()
				errs = append(errs, err)
				continue
			}
			n, copyErr := io.Copy(tmp, io.LimitReader(resp.Body, maxDownloadBytes+1))
			closeErr := resp.Body.Close()
			if copyErr == nil {
				copyErr = tmp.Sync()
			}
			fileErr := tmp.Close()
			if copyErr == nil {
				copyErr = closeErr
			}
			if copyErr == nil {
				copyErr = fileErr
			}
			if copyErr == nil && n > maxDownloadBytes {
				copyErr = fmt.Errorf("download exceeds %d-byte limit", maxDownloadBytes)
			}
			if copyErr != nil || n == 0 {
				_ = os.Remove(tmp.Name())
				if copyErr == nil {
					copyErr = fmt.Errorf("empty download")
				}
				errs = append(errs, fmt.Errorf("%s: %w", url, copyErr))
				continue
			}
			out = append(out, Candidate{Path: tmp.Name(), Temp: tmp.Name(), URL: url, Fresh: true, Bytes: n})
		}
	}
	cachePath := filepath.Join(cacheDir, s.ID+".source")
	if st, err := os.Stat(cachePath); err == nil && st.Size() > 0 {
		out = append(out, Candidate{Path: cachePath, URL: "local cache", Bytes: st.Size()})
	}
	return out, errs
}

func Commit(c Candidate, s config.Source, cacheDir string) (string, error) {
	if !c.Fresh || c.Temp == "" {
		return c.Path, nil
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}
	dst := filepath.Join(cacheDir, s.ID+".source")
	if err := os.Rename(c.Temp, dst); err != nil {
		// Windows does not replace an existing destination atomically.
		old := dst + ".old"
		_ = os.Remove(old)
		if _, statErr := os.Stat(dst); statErr == nil {
			if err := os.Rename(dst, old); err != nil {
				return "", err
			}
		}
		if err := os.Rename(c.Temp, dst); err != nil {
			_ = os.Rename(old, dst)
			return "", err
		}
		_ = os.Remove(old)
	}
	return dst, nil
}

func RemoveCandidate(c Candidate) {
	if c.Temp != "" {
		_ = os.Remove(c.Temp)
	}
}
