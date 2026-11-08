package audit

import (
	"math"
	"net/http"
	"strings"
)

func Analyze(targetURL, rawHTML, keyphrase string, headers http.Header, statusCode int) Scan {
	sig := extractSignals(targetURL, rawHTML, headers, statusCode)
	sig.Keyphrase = strings.TrimSpace(keyphrase)

	s := Scan{
		SEO:           buildCategory(evalSEO(&sig)),
		AEO:           buildCategory(evalAEO(&sig)),
		GEO:           buildCategory(evalGEO(&sig)),
		CWV:           Category{Checks: []Check{}},
		InternalLinks: sig.InternalLinks,
	}
	overall := Overall(s)
	s.Overall = &overall
	return s
}

// ScoreModel selects the scoring formula. Weighted (default) multiplies
// each check's severity weight by its evidence tier factor; "flat" keeps
// the legacy passed/total behavior for one minor release.
var ScoreModel = ScoreModelWeighted

const (
	ScoreModelWeighted = "weighted"
	ScoreModelFlat     = "flat"
)

func buildCategory(checks []Check) Category {
	attachMeta(checks)
	passed := 0
	total := 0
	var weightedPass, weightedTotal float64
	for _, c := range checks {
		if c.Applicable != nil && !*c.Applicable {
			continue
		}
		total++
		weight := c.Severity.Weight() * c.Tier.Factor()
		weightedTotal += weight
		if c.Pass {
			passed++
			weightedPass += weight
		}
	}
	var scoreVal float64
	switch ScoreModel {
	case ScoreModelFlat:
		if total > 0 {
			scoreVal = math.Round((float64(passed) / float64(total)) * 100)
		}
	default:
		if weightedTotal > 0 {
			scoreVal = math.Round((weightedPass / weightedTotal) * 100)
		}
	}
	return Category{
		Score:  &scoreVal,
		Passed: passed,
		Total:  total,
		Checks: checks,
	}
}

// attachMeta fills Severity, Tier and Fix from the registry by check ID.
func attachMeta(checks []Check) {
	for i := range checks {
		m := MetaFor(checks[i].ID)
		checks[i].Severity = m.Severity
		checks[i].Tier = m.Tier
		checks[i].Fix = m.Fix
	}
}

type checkList []Check

func (l *checkList) add(id, label string, pass bool) {
	*l = append(*l, Check{ID: id, Label: label, Pass: pass})
}

const maxEvidence = 20

// addEv records a check with evidence for the failure case, capped at 20
// items so reports stay readable on pathological pages.
func (l *checkList) addEv(id, label string, pass bool, ev ...Evidence) {
	l.add(id, label, pass)
	if pass {
		return
	}
	if len(ev) > maxEvidence {
		ev = ev[:maxEvidence]
	}
	(*l)[len(*l)-1].Evidence = ev
}

func evidence(sel, val string) Evidence {
	return Evidence{Selector: sel, Value: val}
}

func (l *checkList) skip(id, label string) {
	notApplicable := false
	*l = append(*l, Check{ID: id, Label: label, Pass: true, Applicable: &notApplicable})
}

func runChecks(s *PageSignals, groups ...func(*checkList, *PageSignals)) []Check {
	var checks checkList
	for _, group := range groups {
		group(&checks, s)
	}
	return checks
}

func lowerContainsAny(text string, needles ...string) bool {
	lower := strings.ToLower(text)
	for _, n := range needles {
		if strings.Contains(lower, n) {
			return true
		}
	}
	return false
}

func hasSchemaType(types []string, wanted ...string) bool {
	for _, t := range types {
		for _, w := range wanted {
			if strings.EqualFold(t, w) {
				return true
			}
		}
	}
	return false
}

func hasParagraphOfLength(paragraphs []string, minWords, maxWords int) bool {
	for _, p := range paragraphs {
		n := len(strings.Fields(p))
		if n >= minWords && n <= maxWords {
			return true
		}
	}
	return false
}
