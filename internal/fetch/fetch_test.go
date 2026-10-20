package fetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/txyyddss/ip2country/internal/config"
)

func TestFallbackDownloadAndCacheReuse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/primary" {
			http.Error(w, "unavailable", http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte("valid feed"))
	}))
	defer server.Close()

	source := config.Source{ID: "fixture", URLs: []string{server.URL + "/primary"}, FallbackURLs: []string{server.URL + "/fallback"}}
	cacheDir := t.TempDir()
	candidates, failures := Candidates(context.Background(), source, cacheDir, false)
	if len(failures) != 1 || len(candidates) != 1 || !candidates[0].Fresh {
		t.Fatalf("unexpected fallback result: candidates=%+v failures=%v", candidates, failures)
	}
	cachePath, err := Commit(candidates[0], source, cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(cachePath); err != nil || string(got) != "valid feed" {
		t.Fatalf("cache replacement failed: %q, %v", got, err)
	}

	failedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "down", http.StatusBadGateway) }))
	defer failedServer.Close()
	source.URLs = []string{failedServer.URL + "/primary"}
	source.FallbackURLs = []string{failedServer.URL + "/fallback"}
	candidates, failures = Candidates(context.Background(), source, cacheDir, false)
	if len(candidates) != 1 || candidates[0].Fresh || candidates[0].URL != "local cache" || len(failures) != 2 {
		t.Fatalf("cached fallback not retained: candidates=%+v failures=%v", candidates, failures)
	}
	if filepath.Base(candidates[0].Path) != "fixture.source" {
		t.Fatalf("unexpected cache path: %s", candidates[0].Path)
	}
}
