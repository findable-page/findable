package audit

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestEveryEmittedCheckHasRegistryMeta walks every test site page (plus the
// site-level category) and asserts each emitted check ID has real registry
// metadata, so a new check cannot ship without severity, tier and fix text.
func TestEveryEmittedCheckHasRegistryMeta(t *testing.T) {
	files, err := filepath.Glob("../../testdata/site/*.html")
	if err != nil || len(files) == 0 {
		t.Fatalf("no site fixtures: %v", err)
	}
	emitted := map[string]bool{}
	for _, file := range files {
		raw, err := os.ReadFile(filepath.Clean(file))
		if err != nil {
			t.Fatal(err)
		}
		scan := Analyze("https://example.com/"+filepath.Base(file), string(raw), "cloud cost", make(http.Header), 200)
		for _, c := range scan.Reported() {
			for _, k := range c.Cat.Checks {
				emitted[k.ID] = true
			}
		}
	}
	robots := "User-agent: *\nDisallow: /private\n"
	llms := "# Site"
	siteCat := SiteCategory("/", &robots, &llms)
	for _, k := range siteCat.Checks {
		emitted[k.ID] = true
	}

	if len(emitted) < 100 {
		t.Fatalf("suspiciously few emitted check IDs: %d", len(emitted))
	}
	ids := make([]string, 0, len(emitted))
	for id := range emitted {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if _, ok := registry[id]; !ok && !strings.HasPrefix(id, "robots-") {
			t.Errorf("check %s is emitted but missing from the meta registry", id)
			continue
		}
		m := MetaFor(id)
		if strings.TrimSpace(m.Fix) == "" {
			t.Errorf("check %s has no fix text", id)
		}
		if m.Severity.Weight() < 0 || m.Severity.Weight() > 8 {
			t.Errorf("check %s has invalid severity %q", id, m.Severity)
		}
		switch m.Tier {
		case TierDocumented, TierReported, TierHeuristic:
		default:
			t.Errorf("check %s has invalid tier %q", id, m.Tier)
		}
	}
}

// TestWeightedScoreModelHandComputed checks the weighted formula against
// hand-computed values, independent of the implementation.
func TestWeightedScoreModelHandComputed(t *testing.T) {
	// weights: robots_meta critical+documented = 8 (fail)
	//          title_length low+heuristic     = 0.5 (pass)
	//          desc_exists medium+documented  = 2 (pass)
	//          aeo_tables info+heuristic      = 0 (pass)
	// weighted: 2.5 / 10.5 * 100 = 23.809... -> 24
	// flat:     3/4 * 100 = 75
	checks := []Check{
		{ID: "robots_meta", Pass: false},
		{ID: "title_length", Pass: true},
		{ID: "desc_exists", Pass: true},
		{ID: "aeo_tables", Pass: true},
	}
	restore := ScoreModel
	t.Cleanup(func() { ScoreModel = restore })

	ScoreModel = ScoreModelWeighted
	if got := buildCategory(checks).Score; got == nil || *got != 24 {
		t.Fatalf("weighted score = %v, want 24", got)
	}
	ScoreModel = ScoreModelFlat
	if got := buildCategory(checks).Score; got == nil || *got != 75 {
		t.Fatalf("flat score = %v, want 75", got)
	}
}

func TestMetaForUnknownCheckFallsBack(t *testing.T) {
	m := MetaFor("totally_new_check")
	if m.Severity != SeverityMedium || m.Tier != TierHeuristic || m.Fix == "" {
		t.Fatalf("fallback meta = %+v", m)
	}
}

func TestMetaForRobotsBotChecks(t *testing.T) {
	search := MetaFor("robots-oai-searchbot")
	if search.Severity != SeverityHigh || search.Tier != TierDocumented {
		t.Fatalf("search crawler meta = %+v", search)
	}
	training := MetaFor("robots-gptbot")
	if training.Severity != SeverityInfo {
		t.Fatalf("training crawler meta = %+v", training)
	}
	if got := MetaFor("robots-page").Severity; got != SeverityCritical {
		t.Fatalf("robots-page severity = %v", got)
	}
}

func TestRegistryIDsIsComplete(t *testing.T) {
	ids := RegistryIDs()
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate registry id %s", id)
		}
		seen[id] = true
	}
	if !seen["robots-gptbot"] || !seen["robots-oai-searchbot"] {
		t.Error("RegistryIDs must include dynamic robots-bot entries")
	}
}
