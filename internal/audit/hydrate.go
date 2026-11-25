package audit

import "strings"

// HydrationReport compares the raw server HTML with a rendered DOM. The
// checks are not part of the scored audit: they only run when a rendered
// copy is available.
type HydrationReport struct {
	RawWords      int     `json:"rawWords"`
	RenderedWords int     `json:"renderedWords"`
	TextRatio     float64 `json:"textRatio"`
	Checks        []Check `json:"checks"`
}

// CompareHydration diffs the signals AI crawlers see in the raw response
// against the signals present after client rendering.
func CompareHydration(pageURL, rawHTML, renderedHTML string) HydrationReport {
	raw := extractSignals(pageURL, rawHTML, nil, 0)
	rendered := extractSignals(pageURL, renderedHTML, nil, 0)
	report := HydrationReport{
		RawWords:      len(raw.Words),
		RenderedWords: len(rendered.Words),
		TextRatio:     contentRatio(len(raw.Words), len(rendered.Words)),
	}
	checks := checkList{}
	compareTextSignal(&checks, "seo_ssr_title_match", "Server HTML title matches the rendered title", "title", raw.Title, rendered.Title)
	compareTextSignal(&checks, "seo_ssr_canonical_match", "Server HTML canonical matches the rendered canonical", "link[rel=canonical]", raw.Canonical, rendered.Canonical)
	compareHeadings(&checks, raw.Headings, rendered.Headings)
	compareSchema(&checks, raw.SchemaTypes, rendered.SchemaTypes)
	compareContentParity(&checks, report)
	attachMeta(checks)
	report.Checks = checks
	return report
}

func contentRatio(raw, rendered int) float64 {
	if rendered == 0 {
		return 1
	}
	return float64(raw) / float64(rendered)
}

func compareTextSignal(checks *checkList, id, label, selector, raw, rendered string) {
	raw, rendered = strings.TrimSpace(raw), strings.TrimSpace(rendered)
	if rendered == "" {
		checks.add(id, label, true)
		return
	}
	checks.addEv(id, label, raw == rendered, evidence(selector, rendered))
}

func compareHeadings(checks *checkList, raw, rendered []Heading) {
	have := map[string]bool{}
	for _, heading := range raw {
		if heading.Level >= 1 && heading.Level <= 3 {
			have[normSignal(heading.Text)] = true
		}
	}
	var missing []Evidence
	for _, heading := range rendered {
		if heading.Level < 1 || heading.Level > 3 {
			continue
		}
		text := normSignal(heading.Text)
		if text != "" && !have[text] {
			missing = append(missing, evidence(heading.Tag, heading.Text))
		}
	}
	checks.addEv("seo_ssr_headings_present", "Rendered h1-h3 headings are present in the server HTML", len(missing) == 0, missing...)
}

func compareSchema(checks *checkList, raw, rendered []string) {
	have := map[string]bool{}
	for _, kind := range raw {
		have[strings.ToLower(kind)] = true
	}
	var missing []Evidence
	seen := map[string]bool{}
	for _, kind := range rendered {
		key := strings.ToLower(kind)
		if key == "" || have[key] || seen[key] {
			continue
		}
		seen[key] = true
		missing = append(missing, evidence("script[type=application/ld+json]", kind))
	}
	checks.addEv("aeo_ssr_schema_present", "JSON-LD types in the rendered DOM are present in the server HTML", len(missing) == 0, missing...)
}

func compareContentParity(checks *checkList, report HydrationReport) {
	pass := report.RenderedWords == 0 || report.TextRatio >= 0.5
	checks.addEv("geo_ssr_content_parity", "Server-rendered text is at least half of the rendered text", pass, evidence("body", "raw text is under half the rendered text"))
}

func normSignal(text string) string {
	return strings.Join(strings.Fields(strings.ToLower(text)), " ")
}
