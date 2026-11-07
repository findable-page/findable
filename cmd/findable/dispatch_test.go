package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var (
	builtOnce  sync.Once
	binaryPath string
)

// findableBinary builds the CLI once per test run; exit-code and output-shape
// tests exercise the real binary instead of reimplementing main.
func findableBinary(t *testing.T) string {
	t.Helper()
	builtOnce.Do(func() {
		dir, err := os.MkdirTemp("", "findable-bin-*")
		if err != nil {
			t.Fatalf("temp dir: %v", err)
		}
		binaryPath = filepath.Join(dir, "findable")
		out, err := exec.CommandContext(context.Background(), "go", "build", "-o", binaryPath, ".").CombinedOutput() //nolint:gosec // binaryPath is generated in a temporary directory
		if err != nil {
			t.Fatalf("go build: %v\n%s", err, out)
		}
	})
	return binaryPath
}

func runFindable(t *testing.T, workDir string, args ...string) (stdout string, exitCode int) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), findableBinary(t), args...) //nolint:gosec // test invokes its own built binary with test-controlled arguments
	cmd.Dir = workDir
	out, err := cmd.Output()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return string(out), ee.ExitCode()
	}
	if err != nil {
		t.Fatalf("run findable %v: %v", args, err)
	}
	return string(out), 0
}

func TestAuditSubcommandMatchesLegacyFlags(t *testing.T) {
	workDir := t.TempDir()
	page := filepath.Join(workDir, "page.html")
	if err := os.WriteFile(page, []byte("<html><head><title>Dispatch test</title></head><body><h1>x</h1></body></html>"), 0o600); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(workDir, "reports")

	legacyOut, code := runFindable(t, workDir, "-q", "-format", "json", "-o", outDir, page)
	if code != 0 {
		t.Fatalf("legacy exit = %d, out: %s", code, legacyOut)
	}
	legacyReport := reportFiles(t, outDir)
	for _, f := range legacyReport {
		removeAll(t, f)
	}

	auditOut, code := runFindable(t, workDir, "audit", "-q", "-format", "json", "-o", outDir, page)
	if code != 0 {
		t.Fatalf("audit exit = %d, out: %s", code, auditOut)
	}
	if legacyOut != auditOut {
		t.Fatalf("legacy and audit stdout differ:\nlegacy: %q\naudit:  %q", legacyOut, auditOut)
	}
	if got := reportFiles(t, outDir); len(got) != len(legacyReport) {
		t.Fatalf("report count = %d, want %d", len(got), len(legacyReport))
	}
	for i, f := range legacyReport {
		want, _ := os.ReadFile(f) //nolint:gosec // f comes from os.ReadDir of test-owned output directory
		got := reportFiles(t, outDir)[i]
		have, _ := os.ReadFile(got) //nolint:gosec // got comes from os.ReadDir of test-owned output directory
		if string(want) != string(have) {
			t.Fatalf("report %s differs from %s", got, f)
		}
	}
}

func reportFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, filepath.Join(dir, e.Name()))
	}
	return out
}

func removeAll(t *testing.T, path string) {
	t.Helper()
	if err := os.RemoveAll(path); err != nil {
		t.Fatal(err)
	}
}

func TestUnknownSubcommandExitsUsage(t *testing.T) {
	const missingPage = "x.html"
	for _, args := range [][]string{{"frobnicate"}, {"frobnicate", "-q", missingPage}, {"audit", "-mcp", missingPage}, {"audit", "-serve", missingPage}} {
		_, code := runFindable(t, t.TempDir(), args...)
		if code != exitUsage {
			t.Fatalf("%v exit = %d, want %d", args, code, exitUsage)
		}
	}
}

func TestAuditSubcommandRequiresTarget(t *testing.T) {
	_, code := runFindable(t, t.TempDir(), "audit")
	if code != exitUsage {
		t.Fatalf("exit = %d, want %d", code, exitUsage)
	}
}
