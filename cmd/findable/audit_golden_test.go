package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAuditJSONGoldenOnThinFixture runs `findable audit -format json` on the
// thin fixture page and compares the report bytes to a saved snapshot. The
// filename includes the date; the JSON body does not.
func TestAuditJSONGoldenOnThinFixture(t *testing.T) {
	fixture, err := filepath.Abs(filepath.Join("..", "..", "testdata", "site", "01-thin-poor.html"))
	if err != nil {
		t.Fatal(err)
	}
	got := auditJSON(t, fixture)
	if !strings.Contains(got, `"status":"OPEN"`) || !strings.Contains(got, `"accessBlocked":false`) {
		t.Fatal("audit JSON missing the open access gate")
	}
	golden := filepath.Join("..", "..", "testdata", "golden", "audit-01-thin-poor.json")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("audit JSON drifted at byte %d (%d bytes, golden %d). UPDATE_GOLDEN=1 refreshes the snapshot", firstDiff(got, string(want)), len(got), len(want))
	}
	if again := auditJSON(t, fixture); again != got {
		t.Fatalf("audit JSON changed between runs at byte %d", firstDiff(got, again))
	}
}

func auditJSON(t *testing.T, fixture string) string {
	t.Helper()
	outDir := t.TempDir()
	stdout, code := runFindable(t, t.TempDir(), "audit", "-format", "json", "-q", "-o", outDir, fixture)
	if code != 0 {
		t.Fatalf("exit = %d\n%s", code, stdout)
	}
	if !strings.Contains(stdout, "saved ") {
		t.Fatalf("stdout missing saved path:\n%s", stdout)
	}
	matches, err := filepath.Glob(filepath.Join(outDir, "findable-report-01-thin-poor.html-*.json"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("report files %v: %v", matches, err)
	}
	raw, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), outDir) {
		t.Fatal("audit JSON includes the output directory")
	}
	return string(raw)
}

func firstDiff(a, b string) int {
	n := min(len(a), len(b))
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}
