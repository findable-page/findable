package report

import (
	"os"
	"strings"
	"path/filepath"
	"testing"

	"github.com/findable-page/findable/internal/audit"
)

func page(url string, checks ...audit.Check) audit.Page {
	scan := audit.Scan{}
	for _, k := range checks {
		scan.SEO.Checks = append(scan.SEO.Checks, k)
	}
	return audit.Page{URL: url, Scan: scan}
}

func check(id string, sev audit.Severity, tier audit.Tier) audit.Check {
	return audit.Check{ID: id, Label: id + " label", Pass: false, Severity: sev, Tier: tier, Fix: id + " fix sentence"}
}

func samplePages() []audit.Page {
	return []audit.Page{
		page("https://x.com/a",
			check("robots_meta", audit.SeverityCritical, audit.TierDocumented),
			check("title_length", audit.SeverityLow, audit.TierHeuristic),
		),
		page("https://x.com/b",
			check("robots_meta", audit.SeverityCritical, audit.TierDocumented),
			check("img_alt", audit.SeverityMedium, audit.TierDocumented),
		),
	}
}

// Keys: robots_meta 8*1*2=16, img_alt 2*1*1=2, title_length 1*0.5*1=0.5.
func TestFindingsRankingOrder(t *testing.T) {
	fs := Findings(samplePages())
	want := []string{"robots_meta", "img_alt", "title_length"}
	if len(fs) != len(want) {
		t.Fatalf("findings = %v", fs)
	}
	for i, id := range want {
		if fs[i].CheckID != id {
			t.Errorf("finding %d = %s, want %s", i, fs[i].CheckID, id)
		}
	}
	if got := len(fs[0].Pages); got != 2 {
		t.Errorf("robots_meta pages = %d, want 2", got)
	}
}

func TestFindingsTieBreaksByCheckID(t *testing.T) {
	pages := []audit.Page{
		page("https://x.com/a",
			check("zzz_same", audit.SeverityMedium, audit.TierDocumented),
			check("aaa_same", audit.SeverityMedium, audit.TierDocumented),
		),
	}
	fs := Findings(pages)
	if fs[0].CheckID != "aaa_same" || fs[1].CheckID != "zzz_same" {
		t.Fatalf("tie order = %s, %s", fs[0].CheckID, fs[1].CheckID)
	}
}

func TestFindingsSkipsPassAndNAAndInfo(t *testing.T) {
	notApplicable := false
	pages := []audit.Page{page("https://x.com/a",
		audit.Check{ID: "passing", Pass: true, Severity: audit.SeverityHigh, Tier: TierDoc()},
		audit.Check{ID: "na", Pass: false, Applicable: &notApplicable, Severity: audit.SeverityHigh, Tier: TierDoc()},
		check("info_one", audit.SeverityInfo, audit.TierHeuristic),
		check("real", audit.SeverityLow, audit.TierHeuristic),
	)}
	fs := Findings(pages)
	if len(fs) != 1 || fs[0].CheckID != "real" {
		t.Fatalf("findings = %+v", fs)
	}
}

func TierDoc() audit.Tier { return audit.TierDocumented }

func TestFormatFixListGolden(t *testing.T) {
	pages := samplePages()
	pages[1].Scan.SEO.Checks[1].Evidence = []audit.Evidence{{Selector: "img", Value: "hero.png"}}
	got := FormatFixList(Findings(pages), 2, audit.SeverityLow)
	want, err := os.ReadFile(filepath.Join("..", "..", "testdata", "golden", "fixlist.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("fix list:\n%s\nwant:\n%s", got, string(want))
	}
}

func TestFormatFixListTopAndSeverityFilters(t *testing.T) {
	findings := Findings(samplePages())
	got := FormatFixList(findings, 1, audit.SeverityLow)
	if len(got) == 0 || !strings.Contains(got, "robots_meta label") || strings.Contains(got, "img_alt") {
		t.Fatalf("top=1 output:\n%s", got)
	}
	if !strings.Contains(got, "2 more (medium 1, low 1)") {
		t.Fatalf("missing summary line:\n%s", got)
	}
	got = FormatFixList(findings, 10, audit.SeverityHigh)
	if strings.Contains(got, "title_length") || strings.Contains(got, "img_alt") {
		t.Fatalf("min-severity filter failed:\n%s", got)
	}
}

