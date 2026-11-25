package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/findable-page/findable/internal/audit"
)

// CheckDelta is one check whose status changed between baseline and current.
type CheckDelta struct {
	URL    string
	ID     string
	Label  string
	Before string
	After  string
}

// ScoreDelta carries one page's score movement between the two reports.
type ScoreDelta struct {
	URL                      string
	BaseOverall, CurrOverall int
	BaseSEO, CurrSEO         float64
	BaseAEO, CurrAEO         float64
	BaseGEO, CurrGEO         float64
}

// Diff is the comparison of a baseline report against a current report.
type Diff struct {
	Scores       []ScoreDelta
	Regressions  []CheckDelta // passed in baseline, fails now
	Improvements []CheckDelta // failed in baseline, passes now
	NotRescanned []string     // URLs only present in the baseline
}

// LoadReportPages decodes a findable JSON report: a site crawl/sitemap array
// of pages, a {url, scan} page object, or a bare page scan object.
func LoadReportPages(data []byte) ([]audit.Page, error) {
	if trimmed := bytes.TrimLeft(data, " \t\r\n"); len(trimmed) > 0 && trimmed[0] == '[' {
		var pages []audit.Page
		if err := json.Unmarshal(data, &pages); err != nil {
			return nil, fmt.Errorf("decode report array: %w", err)
		}
		return pages, nil
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, fmt.Errorf("decode report object: %w", err)
	}
	if _, hasScan := top["scan"]; hasScan {
		var p audit.Page
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("decode page report: %w", err)
		}
		return []audit.Page{p}, nil
	}
	var s audit.Scan
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("decode scan report: %w", err)
	}
	return []audit.Page{{Scan: s}}, nil
}

// DiffReports compares two loaded reports page by page, matched by URL.
func DiffReports(baseline, current []audit.Page) Diff {
	byURL := make(map[string]audit.Page, len(current))
	for _, p := range current {
		byURL[p.URL] = p
	}
	var d Diff
	seen := make(map[string]bool, len(baseline))
	for _, b := range baseline {
		if seen[b.URL] {
			continue
		}
		seen[b.URL] = true
		c, ok := byURL[b.URL]
		if !ok {
			d.NotRescanned = append(d.NotRescanned, b.URL)
			continue
		}
		d.Scores = append(d.Scores, scoreDelta(b.URL, b.Scan, c.Scan))
		d.collectCheckDeltas(b.URL, b.Scan, c.Scan)
	}
	sortDeltas(d)
	return d
}

func scoreDelta(url string, base, curr audit.Scan) ScoreDelta {
	return ScoreDelta{
		URL:         url,
		BaseOverall: audit.Overall(base), CurrOverall: audit.Overall(curr),
		BaseSEO: audit.ScoreValue(base.SEO), CurrSEO: audit.ScoreValue(curr.SEO),
		BaseAEO: audit.ScoreValue(base.AEO), CurrAEO: audit.ScoreValue(curr.AEO),
		BaseGEO: audit.ScoreValue(base.GEO), CurrGEO: audit.ScoreValue(curr.GEO),
	}
}

func (d *Diff) collectCheckDeltas(url string, base, curr audit.Scan) {
	baseChecks := checkStatuses(base)
	for _, c := range curr.Reported() {
		for _, k := range c.Cat.Checks {
			before, ok := baseChecks[k.ID]
			if !ok {
				continue
			}
			after := audit.Status(k)
			delta := CheckDelta{URL: url, ID: k.ID, Label: k.Label, Before: before, After: after}
			switch {
			case before == audit.StatusPass && after == audit.StatusFail:
				d.Regressions = append(d.Regressions, delta)
			case before == audit.StatusFail && after == audit.StatusPass:
				d.Improvements = append(d.Improvements, delta)
			}
		}
	}
}

func checkStatuses(s audit.Scan) map[string]string {
	out := make(map[string]string)
	for _, c := range s.Reported() {
		for _, k := range c.Cat.Checks {
			out[k.ID] = audit.Status(k)
		}
	}
	return out
}

func sortDeltas(d Diff) {
	for _, g := range [][]CheckDelta{d.Regressions, d.Improvements} {
		sort.Slice(g, func(i, j int) bool {
			if g[i].URL != g[j].URL {
				return g[i].URL < g[j].URL
			}
			return g[i].ID < g[j].ID
		})
	}
	sort.Slice(d.Scores, func(i, j int) bool { return d.Scores[i].URL < d.Scores[j].URL })
	sort.Strings(d.NotRescanned)
}

// Tripped reports whether any check regressed or any page's overall score
// dropped by more than maxDrop points, for -fail-on-regression gating.
func (d Diff) Tripped(maxDrop int) bool {
	if len(d.Regressions) > 0 {
		return true
	}
	for _, s := range d.Scores {
		if s.BaseOverall-s.CurrOverall > maxDrop {
			return true
		}
	}
	return false
}

func (d Diff) summary() string {
	return fmt.Sprintf("pages compared %d  regressions %d  improvements %d  not rescanned %d\n",
		len(d.Scores), len(d.Regressions), len(d.Improvements), len(d.NotRescanned))
}

func pageLabel(url string) string {
	if url == "" {
		return "(page report)"
	}
	return url
}

func (d Diff) Terminal() string {
	var b strings.Builder
	b.WriteString(d.summary())
	for _, s := range d.Scores {
		fmt.Fprintf(&b, "  %s  overall %d -> %d (%+d)  SEO %.0f -> %.0f  AEO %.0f -> %.0f  GEO %.0f -> %.0f\n",
			pageLabel(s.URL), s.BaseOverall, s.CurrOverall, s.CurrOverall-s.BaseOverall,
			s.BaseSEO, s.CurrSEO, s.BaseAEO, s.CurrAEO, s.BaseGEO, s.CurrGEO)
	}
	for _, g := range []struct {
		title  string
		deltas []CheckDelta
	}{{"regressions (passed in baseline, failing now)", d.Regressions}, {"improvements (failed in baseline, passing now)", d.Improvements}} {
		if len(g.deltas) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n%s (%d):\n", g.title, len(g.deltas))
		for _, k := range g.deltas {
			fmt.Fprintf(&b, "  %s  %s (%s -> %s) - %s\n", pageLabel(k.URL), k.ID, k.Before, k.After, k.Label)
		}
	}
	if len(d.NotRescanned) > 0 {
		fmt.Fprintf(&b, "\nnot rescanned (%d): %s\n", len(d.NotRescanned), strings.Join(d.NotRescanned, ", "))
	}
	return b.String()
}

func (d Diff) Markdown() string {
	var b strings.Builder
	b.WriteString(d.summary())
	b.WriteString("\n| URL | Overall | SEO | AEO | GEO |\n|---|---|---|---|---|\n")
	for _, s := range d.Scores {
		fmt.Fprintf(&b, "| %s | %d -> %d | %.0f -> %.0f | %.0f -> %.0f | %.0f -> %.0f |\n",
			pageLabel(s.URL), s.BaseOverall, s.CurrOverall, s.BaseSEO, s.CurrSEO, s.BaseAEO, s.CurrAEO, s.BaseGEO, s.CurrGEO)
	}
	b.WriteString("\n| URL | Check | Baseline | Current |\n|---|---|---|---|\n")
	for _, k := range d.Regressions {
		fmt.Fprintf(&b, "| %s | `%s` %s | %s | %s |\n", pageLabel(k.URL), k.ID, k.Label, k.Before, k.After)
	}
	for _, k := range d.Improvements {
		fmt.Fprintf(&b, "| %s | `%s` %s | %s | %s |\n", pageLabel(k.URL), k.ID, k.Label, k.Before, k.After)
	}
	return b.String()
}
