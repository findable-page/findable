package main

import (
	"encoding/json"
	"strings"
	"testing"
)

const recorded = `{
  "seo": {"score": 92, "passed": 2, "total": 3, "checks": [
    {"id": "title", "label": "Title is \"short\"", "pass": true},
    {"id": "hreflang", "label": "No hreflang", "pass": false},
    {"id": "amp", "label": "AMP", "pass": false, "applicable": false}]},
  "aeo": {"score": 90, "passed": 1, "total": 1, "checks": [{"id": "faq", "label": "FAQ schema", "pass": true}]},
  "geo": {"score": 84, "passed": 0, "total": 1, "checks": [{"id": "quotes", "label": "No expert quotes", "pass": false}]},
  "cwv": {"checks": []}
}`

func decode(t *testing.T) scan {
	t.Helper()
	var s scan
	if err := json.Unmarshal([]byte(recorded), &s); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestOverallWeightsSEOHalfAndRounds(t *testing.T) {
	// 0.5*92 + 0.25*90 + 0.25*84 = 89.5, which rounds up to 90
	if got := overall(decode(t)); got != 90 {
		t.Fatalf("overall = %d, want 90", got)
	}
}

func TestMissingScoreCountsAsZero(t *testing.T) {
	if got := overall(scan{}); got != 0 {
		t.Fatalf("overall of empty scan = %d, want 0", got)
	}
}

func TestStatusDistinguishesPassFailAndNotApplicable(t *testing.T) {
	checks := decode(t).SEO.Checks
	for i, want := range []string{"PASS", "FAIL", "N/A"} {
		if got := status(checks[i]); got != want {
			t.Errorf("%s: status = %s, want %s", checks[i].ID, got, want)
		}
	}
}

func TestCSVHasBOMCategoriesAndQuotedLabels(t *testing.T) {
	lines := strings.Split(csv("example.com", decode(t)), "\n")
	if !strings.HasPrefix(lines[0], "\xef\xbb\xbfSEO/AEO/GEO Report,example.com,") {
		t.Errorf("header row = %q", lines[0])
	}
	for _, want := range []string{
		"Overall,90,,",
		"SEO,92,2,3",
		"AEO,90,1,1",
		"GEO,84,0,1",
		`title,SEO,PASS,"Title is ""short"""`,
		`hreflang,SEO,FAIL,"No hreflang"`,
		`amp,SEO,N/A,"AMP"`,
		`quotes,GEO,FAIL,"No expert quotes"`,
	} {
		if !contains(lines, want) {
			t.Errorf("csv is missing row %q", want)
		}
	}
	if contains(lines, "CWV,0,0,0") {
		t.Error("CWV row must be left out when there is no CWV score")
	}
}

func contains(lines []string, want string) bool {
	for _, l := range lines {
		if l == want {
			return true
		}
	}
	return false
}
