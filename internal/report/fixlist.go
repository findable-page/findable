package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/findable-page/findable/internal/audit"
)

// Finding aggregates every failing instance of one check across pages,
// ranked by severity weight x tier factor x pages affected.
type Finding struct {
	CheckID  string
	Category string
	Label    string
	Severity audit.Severity
	Tier     audit.Tier
	Fix      string
	Evidence []audit.Evidence
	Pages    []string
}

// Key is the ranking key: severity weight * tier factor * pages affected.
func (f Finding) Key() float64 {
	return f.Severity.Weight() * f.Tier.Factor() * float64(len(f.Pages))
}

// Findings collapses failing checks across pages into one ordered list,
// highest impact first; ties break by check ID for stable output.
func Findings(pages []audit.Page) []Finding {
	type group struct {
		f        Finding
		seenPage map[string]bool
	}
	groups := map[string]*group{}
	for _, p := range pages {
		for _, cat := range p.Scan.Reported() {
			for _, k := range cat.Cat.Checks {
				if audit.Status(k) != audit.StatusFail || k.Severity == audit.SeverityInfo {
					continue
				}
				g := groups[k.ID]
				if g == nil {
					g = &group{f: Finding{
						CheckID:  k.ID,
						Category: cat.Name,
						Label:    k.Label,
						Severity: k.Severity,
						Tier:     k.Tier,
						Fix:      k.Fix,
						Evidence: k.Evidence,
					}, seenPage: map[string]bool{}}
					groups[k.ID] = g
				}
				if !g.seenPage[p.URL] {
					g.seenPage[p.URL] = true
					g.f.Pages = append(g.f.Pages, p.URL)
				}
			}
		}
	}
	out := make([]Finding, 0, len(groups))
	for _, g := range groups {
		out = append(out, g.f)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Key() != out[j].Key() {
			return out[i].Key() > out[j].Key()
		}
		return out[i].CheckID < out[j].CheckID
	})
	return out
}

// FilterFindings drops pages a predicate wants ignored (per-check, per-URL);
// findings left with no pages disappear entirely.
func FilterFindings(fs []Finding, drop func(checkID, pageURL string) bool) []Finding {
	out := fs[:0]
	for _, f := range fs {
		pages := f.Pages[:0]
		for _, p := range f.Pages {
			if !drop(f.CheckID, p) {
				pages = append(pages, p)
			}
		}
		if len(pages) > 0 {
			f.Pages = pages
			out = append(out, f)
		}
	}
	return out
}

// FormatFixList renders the ranked list for the terminal: severity, tier,
// label, page count, fix line and first evidence item.
func FormatFixList(findings []Finding, top int, minSeverity audit.Severity) string {
	if top <= 0 {
		top = 10
	}
	var b strings.Builder
	shown := 0
	var rest map[audit.Severity]int
	for _, f := range findings {
		if f.Severity.Weight() < minSeverity.Weight() {
			continue
		}
		if shown == top {
			if rest == nil {
				rest = map[audit.Severity]int{}
			}
			rest[f.Severity]++
			continue
		}
		shown++
		pages := ""
		if len(f.Pages) > 1 {
			pages = fmt.Sprintf("    [%d pages]", len(f.Pages))
		}
		fmt.Fprintf(&b, "%2d %-8s  %-11s %s%s\n", shown, strings.ToUpper(string(f.Severity)), f.Tier, f.Label, pages)
		fmt.Fprintf(&b, "   fix: %s\n", f.Fix)
		if len(f.Evidence) > 0 {
			ev := f.Evidence[0]
			value := ev.Value
			if len(value) > 60 {
				value = value[:60] + "..."
			}
			if value != "" {
				fmt.Fprintf(&b, "   evidence: %s  (%s)\n", ev.Selector, value)
			} else {
				fmt.Fprintf(&b, "   evidence: %s\n", ev.Selector)
			}
		}
	}
	if rest != nil {
		var parts []string
		for _, sev := range []audit.Severity{audit.SeverityCritical, audit.SeverityHigh, audit.SeverityMedium, audit.SeverityLow} {
			if n := rest[sev]; n > 0 {
				parts = append(parts, fmt.Sprintf("%s %d", sev, n))
			}
		}
		fmt.Fprintf(&b, "%d more (%s)\n", len(findings)-shown, strings.Join(parts, ", "))
	}
	return b.String()
}
