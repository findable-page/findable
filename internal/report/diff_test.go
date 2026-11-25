package report

import (
	"strings"
	"testing"

	"github.com/findable-page/findable/internal/audit"
)

func scanFor(scores [3]float64, checks map[string]bool) audit.Scan {
	mk := func(score float64) audit.Category {
		return audit.Category{Score: &score}
	}
	seo := mk(scores[0])
	for id, pass := range checks {
		seo.Checks = append(seo.Checks, audit.Check{ID: id, Label: id + " label", Pass: pass})
	}
	return audit.Scan{SEO: seo, AEO: mk(scores[1]), GEO: mk(scores[2])}
}

func TestLoadReportPagesAcceptsAllJSONShapes(t *testing.T) {
	page := audit.Page{URL: rootPage, Scan: scanFor([3]float64{90, 80, 70}, map[string]bool{"seo_title_present": true})}

	array := []byte(EncodeJSON([]audit.Page{page}))
	got, err := LoadReportPages(array)
	if err != nil || len(got) != 1 || got[0].URL != rootPage {
		t.Fatalf("array decode = %v, %v", got, err)
	}

	got, err = LoadReportPages([]byte(EncodeJSON(page)))
	if err != nil || len(got) != 1 || got[0].URL != rootPage {
		t.Fatalf("page-object decode = %v, %v", got, err)
	}

	bare := []byte(EncodeJSON(page.Scan))
	got, err = LoadReportPages(bare)
	if err != nil || len(got) != 1 || got[0].URL != "" || len(got[0].Scan.SEO.Checks) != 1 {
		t.Fatalf("bare-scan decode = %v, %v", got, err)
	}

	if _, err = LoadReportPages([]byte("not json")); err == nil {
		t.Fatal("garbage decode must fail")
	}
}

const (
	chkKeep    = "chk_keep"
	chkRegress = "chk_regress"
	chkFixed   = "chk_fixed"
)

func TestDiffReportsFlagsCheckAndScoreMovement(t *testing.T) {
	baseline := []audit.Page{
		{URL: rootPage, Scan: scanFor([3]float64{90, 80, 70}, map[string]bool{chkKeep: true, chkRegress: true, chkFixed: false})},
		{URL: subPage, Scan: scanFor([3]float64{90, 80, 70}, nil)},
	}
	current := []audit.Page{
		{URL: rootPage, Scan: scanFor([3]float64{80, 80, 70}, map[string]bool{chkKeep: true, chkRegress: false, chkFixed: true})},
	}

	d := DiffReports(baseline, current)
	if len(d.Regressions) != 1 || d.Regressions[0].ID != chkRegress || d.Regressions[0].Before != audit.StatusPass || d.Regressions[0].After != audit.StatusFail {
		t.Fatalf("regressions = %+v", d.Regressions)
	}
	if len(d.Improvements) != 1 || d.Improvements[0].ID != chkFixed {
		t.Fatalf("improvements = %+v", d.Improvements)
	}
	if len(d.NotRescanned) != 1 || d.NotRescanned[0] != subPage {
		t.Fatalf("not rescanned = %v", d.NotRescanned)
	}
	if len(d.Scores) != 1 || d.Scores[0].URL != rootPage || d.Scores[0].BaseOverall != 83 || d.Scores[0].CurrOverall != 78 {
		t.Fatalf("scores = %+v", d.Scores)
	}
}

func TestDiffTrippedHonorsMaxDrop(t *testing.T) {
	drop := func(base, curr float64) Diff {
		return DiffReports(
			[]audit.Page{{URL: rootPage, Scan: scanFor([3]float64{base, 80, 70}, nil)}},
			[]audit.Page{{URL: rootPage, Scan: scanFor([3]float64{curr, 80, 70}, nil)}},
		)
	}
	noChange := DiffReports(
		[]audit.Page{{URL: rootPage, Scan: scanFor([3]float64{90, 80, 70}, nil)}},
		[]audit.Page{{URL: rootPage, Scan: scanFor([3]float64{90, 80, 70}, nil)}},
	)
	if noChange.Tripped(0) {
		t.Fatal("identical reports must not trip")
	}
	if !drop(90, 80).Tripped(3) {
		t.Fatal("10-point drop must trip with max-drop 3")
	}
	if drop(90, 80).Tripped(20) {
		t.Fatal("10-point drop must not trip with max-drop 20")
	}
}

func TestDiffOutputIsStableAndRendered(t *testing.T) {
	baseline := []audit.Page{{URL: rootPage, Scan: scanFor([3]float64{90, 80, 70}, map[string]bool{chkRegress: true, chkFixed: false})}}
	current := []audit.Page{{URL: rootPage, Scan: scanFor([3]float64{80, 80, 70}, map[string]bool{chkRegress: false, chkFixed: true})}}
	one, two := DiffReports(baseline, current), DiffReports(baseline, current)
	if one.Terminal() != two.Terminal() || one.Markdown() != two.Markdown() {
		t.Fatal("diff output must be deterministic")
	}
	if s := one.Terminal(); !strings.Contains(s, "regressions 1") || !strings.Contains(s, chkRegress) || !strings.Contains(s, "->") {
		t.Fatalf("terminal output missing pieces:\n%s", s)
	}
	if s := one.Markdown(); !strings.Contains(s, "| URL |") || !strings.Contains(s, "`chk_fixed`") {
		t.Fatalf("markdown output missing pieces:\n%s", s)
	}
}
