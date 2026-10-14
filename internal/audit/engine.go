package audit

import (
	"math"
	"net/http"
	"strings"
)

func Analyze(targetURL, rawHTML, keyphrase string, headers http.Header, statusCode int) Scan {
	sig := extractSignals(targetURL, rawHTML, headers, statusCode)
	sig.Keyphrase = keyphrase

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

func buildCategory(checks []Check) Category {
	passed := 0
	total := 0
	for _, c := range checks {
		if c.Applicable != nil && !*c.Applicable {
			continue
		}
		total++
		if c.Pass {
			passed++
		}
	}
	var scoreVal float64
	if total > 0 {
		scoreVal = math.Round((float64(passed) / float64(total)) * 100)
	}
	return Category{
		Score:  &scoreVal,
		Passed: passed,
		Total:  total,
		Checks: checks,
	}
}

type checkList []Check

func (l *checkList) add(id, label string, pass bool) {
	*l = append(*l, Check{ID: id, Label: label, Pass: pass})
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
