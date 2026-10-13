package main

import (
	"fmt"
	"math"
)

type check struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Pass       bool   `json:"pass"`
	Applicable *bool  `json:"applicable"`
}

type category struct {
	Score  *float64 `json:"score"`
	Passed int      `json:"passed"`
	Total  int      `json:"total"`
	Checks []check  `json:"checks"`
}

type scan struct {
	SEO           category `json:"seo"`
	AEO           category `json:"aeo"`
	GEO           category `json:"geo"`
	CWV           category `json:"cwv"`
	InternalLinks []string `json:"internalLinks"`
}

type named struct {
	name string
	cat  category
}

func (s scan) scored() []named {
	return []named{{"SEO", s.SEO}, {"AEO", s.AEO}, {"GEO", s.GEO}}
}

const (
	statusPass          = "PASS"
	statusFail          = "FAIL"
	statusNotApplicable = "N/A"
)

func status(c check) string {
	switch {
	case c.Applicable != nil && !*c.Applicable:
		return statusNotApplicable
	case c.Pass:
		return statusPass
	}
	return statusFail
}

func val(c category) float64 {
	if c.Score == nil {
		return 0
	}
	return *c.Score
}

func score(c category) string { return fmt.Sprintf("%.0f", val(c)) }

func overall(s scan) int {
	return int(math.Round(val(s.SEO)*.5 + val(s.AEO)*.25 + val(s.GEO)*.25))
}

func failedCount(s scan) int {
	cnt := 0
	for _, c := range s.scored() {
		for _, k := range c.cat.Checks {
			if status(k) == statusFail {
				cnt++
			}
		}
	}
	return cnt
}
