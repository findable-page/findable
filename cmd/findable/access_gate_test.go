package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditPrintsAccessBlockedBanner(t *testing.T) {
	workDir := t.TempDir()
	page := filepath.Join(workDir, "blocked.html")
	html := `<!DOCTYPE html><html><head><meta name="robots" content="noindex, nosnippet"><title>Blocked</title></head><body><h1>Blocked</h1></body></html>`
	if err := os.WriteFile(page, []byte(html), 0o600); err != nil {
		t.Fatal(err)
	}
	out, code := runFindable(t, workDir, "-q", page)
	if code != 0 {
		t.Fatalf("exit = %d, out: %s", code, out)
	}
	if !strings.Contains(out, "BLOCKED_FROM_INDEXING") || !strings.Contains(out, "cannot be cited") {
		t.Fatalf("stdout missing access banner:\n%s", out)
	}
	if !strings.Contains(out, "robots_meta") {
		t.Fatalf("stdout missing robots_meta reason:\n%s", out)
	}
}

func TestAuditSiteRobotsDisallowPrintsBlockedGate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nDisallow: /hidden\n"))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Open page with a long enough title here</title></head><body><h1>Open</h1></body></html>`))
	}))
	t.Cleanup(srv.Close)

	out, code := runFindable(t, t.TempDir(), "audit", "-q", "-site", "-allow-private", srv.URL+"/hidden")
	if code != 0 || !strings.Contains(out, "BLOCKED_FROM_INDEXING") || !strings.Contains(out, "robots-page") {
		t.Fatalf("exit = %d, stdout:\n%s", code, out)
	}

	open, code := runFindable(t, t.TempDir(), "audit", "-q", "-site", "-allow-private", srv.URL+"/")
	if code != 0 || strings.Contains(open, "BLOCKED_FROM_INDEXING") {
		t.Fatalf("allowed page exit = %d, stdout:\n%s", code, open)
	}
}
