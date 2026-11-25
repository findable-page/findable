package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFanoutCommandPrintsCoverageMatrix(t *testing.T) {
	workDir := t.TempDir()
	page := filepath.Join(workDir, "page.html")
	html := `<!DOCTYPE html><html><body><h1>What is findable?</h1><p>Findable is a scanner.</p><h2>Pricing</h2><p>The price is zero.</p></body></html>`
	if err := os.WriteFile(page, []byte(html), 0o600); err != nil {
		t.Fatal(err)
	}
	out, code := runFindable(t, workDir, "fanout", "-query", "findable", page)
	if code != 0 {
		t.Fatalf("exit = %d\n%s", code, out)
	}
	if !strings.Contains(out, "covered ") || !strings.Contains(out, "definition") || !strings.Contains(out, "pricing") {
		t.Fatalf("stdout missing coverage matrix:\n%s", out)
	}
	jsonOut, code := runFindable(t, workDir, "fanout", "-query", "findable", "-format", "json", page)
	if code != 0 || !strings.Contains(jsonOut, `"covered":true`) {
		t.Fatalf("json exit = %d\n%s", code, jsonOut)
	}
}
