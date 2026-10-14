package audit

import (
	"fmt"
	"math"
)

type Check struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Pass       bool   `json:"pass"`
	Applicable *bool  `json:"applicable"`
}

type Category struct {
	Score  *float64 `json:"score"`
	Passed int      `json:"passed"`
	Total  int      `json:"total"`
	Checks []Check  `json:"checks"`
}

type Scan struct {
	Overall       *int      `json:"overall,omitempty"`
	SEO           Category  `json:"seo"`
	AEO           Category  `json:"aeo"`
	GEO           Category  `json:"geo"`
	CWV           Category  `json:"cwv"`
	InternalLinks []string  `json:"internalLinks"`
	Site          *Category `json:"site,omitempty"`
}

type Page struct {
	URL  string `json:"url"`
	Scan Scan   `json:"scan"`
}

type NamedCategory struct {
	Name string
	Cat  Category
}

func (s Scan) Scored() []NamedCategory {
	return []NamedCategory{{"SEO", s.SEO}, {"AEO", s.AEO}, {"GEO", s.GEO}}
}

func (s Scan) Reported() []NamedCategory {
	if s.Site == nil {
		return s.Scored()
	}
	return append(s.Scored(), NamedCategory{"Site", *s.Site})
}

const (
	StatusPass          = "PASS"
	StatusFail          = "FAIL"
	StatusNotApplicable = "N/A"
)

func Status(c Check) string {
	switch {
	case c.Applicable != nil && !*c.Applicable:
		return StatusNotApplicable
	case c.Pass:
		return StatusPass
	}
	return StatusFail
}

func ScoreValue(c Category) float64 {
	if c.Score == nil {
		return 0
	}
	return *c.Score
}

func ScoreText(c Category) string { return fmt.Sprintf("%.0f", ScoreValue(c)) }

func Overall(s Scan) int {
	return int(math.Round(ScoreValue(s.SEO)*.5 + ScoreValue(s.AEO)*.25 + ScoreValue(s.GEO)*.25))
}

func FailedCount(s Scan) int {
	cnt := 0
	for _, c := range s.Scored() {
		for _, k := range c.Cat.Checks {
			if Status(k) == StatusFail {
				cnt++
			}
		}
	}
	return cnt
}

func PagesBelow(pages []Page, minimum int) []Page {
	var below []Page
	for _, p := range pages {
		if Overall(p.Scan) < minimum {
			below = append(below, p)
		}
	}
	return below
}
