package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	baselineJSON = `{"seo":{"score":90,"checks":[{"id":"seo_title_present","label":"Page has a title","pass":true}]},"aeo":{"score":80},"geo":{"score":70}}`
	currentJSON  = `{"seo":{"score":70,"checks":[{"id":"seo_title_present","label":"Page has a title","pass":false}]},"aeo":{"score":80},"geo":{"score":70}}`
)

func writeDiffFixture(t *testing.T, dir string) (baseline, current string) {
	t.Helper()
	baseline = filepath.Join(dir, "baseline.json")
	current = filepath.Join(dir, "current.json")
	if err := os.WriteFile(baseline, []byte(baselineJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(current, []byte(currentJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	return baseline, current
}

func TestDiffFlagsRegressionAndExitCode(t *testing.T) {
	dir := t.TempDir()
	baseline, current := writeDiffFixture(t, dir)

	out, code := runFindable(t, dir, "diff", baseline, current)
	if code != 0 || !strings.Contains(out, "regressions 1") || !strings.Contains(out, "seo_title_present") {
		t.Fatalf("diff exit = %d, out:\n%s", code, out)
	}

	if _, code = runFindable(t, dir, "diff", "-fail-on-regression", baseline, current); code != 1 {
		t.Fatalf("-fail-on-regression exit = %d, want 1", code)
	}
	if _, code = runFindable(t, dir, "diff", "-markdown", baseline, current); code != 0 {
		t.Fatalf("-markdown exit = %d, want 0", code)
	}
	out, code = runFindable(t, dir, "diff", "-markdown", "-fail-on-regression", "-max-drop", "20", baseline, current)
	if code != 1 || !strings.Contains(out, "| URL |") {
		t.Fatalf("markdown + regression exit = %d, out:\n%s", code, out)
	}
}

func TestDiffUsageAndRuntimeErrors(t *testing.T) {
	dir := t.TempDir()
	baseline, current := writeDiffFixture(t, dir)

	if _, code := runFindable(t, dir, "diff"); code != exitUsage {
		t.Fatalf("missing args exit = %d, want %d", code, exitUsage)
	}
	if _, code := runFindable(t, dir, "diff", baseline); code != exitUsage {
		t.Fatalf("one arg exit = %d, want %d", code, exitUsage)
	}
	if _, code := runFindable(t, dir, "diff", filepath.Join(dir, "missing.json"), current); code != exitRuntime {
		t.Fatalf("unreadable baseline exit = %d, want %d", code, exitRuntime)
	}
	if _, code := runFindable(t, dir, "diff", current, filepath.Join(dir, "also-missing.json")); code != exitRuntime {
		t.Fatalf("unreadable current exit = %d, want %d", code, exitRuntime)
	}
}
