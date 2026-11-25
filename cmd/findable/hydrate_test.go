package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHydrateCommandReportsClientOnlyTitle(t *testing.T) {
	workDir := t.TempDir()
	raw := filepath.Join(workDir, "raw.html")
	rendered := filepath.Join(workDir, "rendered.html")
	if err := os.WriteFile(raw, []byte(`<!DOCTYPE html><html><head><title>Shell</title></head><body><div id="app"></div></body></html>`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rendered, []byte(`<!DOCTYPE html><html><head><title>Real Title</title></head><body><h1>Real Title</h1><p>Rendered answer text.</p></body></html>`), 0o600); err != nil {
		t.Fatal(err)
	}
	out, code := runFindable(t, workDir, "hydrate", "-rendered", rendered, "-format", "text", raw)
	if code != 0 {
		t.Fatalf("exit = %d\n%s", code, out)
	}
	if !strings.Contains(out, "seo_ssr_title_match") || !strings.Contains(out, "FAIL") {
		t.Fatalf("stdout missing failing title check:\n%s", out)
	}

	jsonOut, code := runFindable(t, workDir, "hydrate", "-rendered", rendered, "-format", "json", raw)
	if code != 0 || !strings.Contains(jsonOut, "seo_ssr_title_match") {
		t.Fatalf("json exit = %d\n%s", code, jsonOut)
	}
	if !strings.Contains(jsonOut, `"id":"seo_ssr_title_match"`) && !strings.Contains(jsonOut, `"id": "seo_ssr_title_match"`) {
		t.Fatalf("json missing title check:\n%s", jsonOut)
	}

	script := filepath.Join(workDir, "render.sh")
	body := "#!/bin/sh\nprintf '%s' '<html><head><title>From renderer</title></head></html>'\n"
	// The renderer is executed as a program, so it has to be executable.
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil { //nolint:gosec // test-only renderer, not a secret
		t.Fatal(err)
	}
	cmdOut, code := runFindable(t, workDir, "hydrate", "-render-cmd", script, raw)
	if code != 0 || !strings.Contains(cmdOut, "seo_ssr_title_match") || !strings.Contains(cmdOut, "FAIL") {
		t.Fatalf("render-cmd exit = %d\n%s", code, cmdOut)
	}
}
