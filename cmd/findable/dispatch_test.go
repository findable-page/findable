package main

import (
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
		out, err := exec.Command("go", "build", "-o", binaryPath, ".").CombinedOutput()
		if err != nil {
			t.Fatalf("go build: %v\n%s", err, out)
		}
	})
	return binaryPath
}

func runFindable(t *testing.T, workDir string, args ...string) (stdout string, exitCode int) {
	t.Helper()
	cmd := exec.Command(findableBinary(t), args...)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if ee, ok := err.(*exec.ExitError); ok {
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
	if err := os.WriteFile(page, []byte("<html><head><title>Dispatch test</title></head><body><h1>x</h1></body></html>"), 0o644); err != nil {
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
		want, _ := os.ReadFile(f)
		got := reportFiles(t, outDir)[i]
		have, _ := os.ReadFile(got)
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
	var out []string
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
	_, code := runFindable(t, t.TempDir(), "frobnicate", "-q", "x.html")
	if code != exitUsage {
		t.Fatalf("exit = %d, want %d", code, exitUsage)
	}
}

func TestAuditSubcommandRequiresTarget(t *testing.T) {
	_, code := runFindable(t, t.TempDir(), "audit")
	if code != exitUsage {
		t.Fatalf("exit = %d, want %d", code, exitUsage)
	}
}
