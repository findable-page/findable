package audit

import (
	"fmt"
	"testing"
)

func TestScoreTextRoundsAndTreatsAMissingScoreAsZero(t *testing.T) {
	if got := ScoreText(Category{}); got != "0" {
		t.Fatalf("ScoreText(nil) = %q, want 0", got)
	}
	half := 10.5
	if got := ScoreText(Category{Score: &half}); got != "11" {
		t.Fatalf("ScoreText(10.5) = %q, want 11", got)
	}
	below := 10.4
	if got := ScoreText(Category{Score: &below}); got != "10" {
		t.Fatalf("ScoreText(10.4) = %q, want 10", got)
	}
}

func TestFailedCountIgnoresPassesNotApplicableAndSiteChecks(t *testing.T) {
	notApplicable := false
	scan := Scan{
		SEO: Category{Checks: []Check{
			{ID: "failing", Pass: false},
			{ID: "passing", Pass: true},
			{ID: "skipped", Pass: false, Applicable: &notApplicable},
		}},
		AEO:  Category{Checks: []Check{{ID: "aeo-fail", Pass: false}}},
		GEO:  Category{Checks: []Check{{ID: "geo-pass", Pass: true}}},
		Site: &Category{Checks: []Check{{ID: "site-fail", Pass: false}}},
	}
	if got := FailedCount(scan); got != 2 {
		t.Fatalf("FailedCount = %d, want 2", got)
	}
}

func TestAddEvCapsEvidenceAndOmitsItWhenTheCheckPasses(t *testing.T) {
	ev := make([]Evidence, maxEvidence+5)
	for i := range ev {
		ev[i] = evidence("p", fmt.Sprintf("item-%d", i))
	}
	var checks checkList
	checks.addEv("capped", "too much evidence", false, ev...)
	got := checks[0].Evidence
	if len(got) != maxEvidence {
		t.Fatalf("evidence len = %d, want %d", len(got), maxEvidence)
	}
	if got[0].Value != "item-0" || got[maxEvidence-1].Value != fmt.Sprintf("item-%d", maxEvidence-1) {
		t.Fatalf("evidence was not the first %d items: %+v", maxEvidence, got)
	}
	checks.addEv("passing", "no evidence on a pass", true, ev...)
	if checks[1].Evidence != nil {
		t.Fatalf("passing check stored evidence: %+v", checks[1].Evidence)
	}
}
