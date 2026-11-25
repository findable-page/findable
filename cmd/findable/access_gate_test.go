package main

import (
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
	if strings.Contains(out, "No robots meta tag blocking crawling") || !strings.Contains(out, "noindex") {
		t.Fatalf("access reason should name the blocking directive, got:\n%s", out)
	}
}
