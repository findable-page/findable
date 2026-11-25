package audit

import (
	"strings"
	"testing"
)

func TestCompareHydrationFlagsClientOnlySignals(t *testing.T) {
	raw := `<!DOCTYPE html><html><head><title>Shell</title></head><body><div id="app"></div></body></html>`
	rendered := `<!DOCTYPE html><html><head><title>Real Title</title>
	<link rel="canonical" href="https://example.com/real">
	<script type="application/ld+json">{"@type":"Article","headline":"Real"}</script>
	</head><body><h1>Real Title</h1><p>` + strings.Repeat("word ", 40) + `</p></body></html>`

	report := CompareHydration("https://example.com", raw, rendered)
	if report.TextRatio >= 0.5 {
		t.Fatalf("text ratio = %v, want below 0.5", report.TextRatio)
	}
	for _, id := range []string{"seo_ssr_title_match", "seo_ssr_canonical_match", "seo_ssr_headings_present", "aeo_ssr_schema_present", "geo_ssr_content_parity"} {
		check := findCheck(report.Checks, id)
		if check == nil || check.Pass {
			t.Errorf("%s should fail, got %+v", id, check)
		}
	}
}

func TestCompareHydrationPassesWhenRawMatchesRendered(t *testing.T) {
	page := `<!DOCTYPE html><html><head><title>Same</title>
	<link rel="canonical" href="https://example.com/same">
	<script type="application/ld+json">{"@type":"Article"}</script>
	</head><body><h1>Same</h1><h2>Details</h2><p>The same paragraph.</p></body></html>`
	report := CompareHydration("https://example.com/same", page, page)
	for _, check := range report.Checks {
		if !check.Pass {
			t.Errorf("%s failed on identical HTML: %s", check.ID, check.Label)
		}
	}
}

func TestCompareHydrationEquatesCosmeticCanonicalAndTitleDifferences(t *testing.T) {
	raw := `<!DOCTYPE html><html><head><title>Real   Title</title>
	<link rel="canonical" href="https://Example.com:443/docs">
	</head><body><h1>Same</h1><p>Same words here.</p></body></html>`
	rendered := `<!DOCTYPE html><html><head><title>Real Title</title>
	<link rel="canonical" href="https://example.com/docs">
	</head><body><h1>Same</h1><p>Same words here.</p></body></html>`
	report := CompareHydration("https://example.com/docs", raw, rendered)
	for _, id := range []string{"seo_ssr_title_match", "seo_ssr_canonical_match"} {
		check := findCheck(report.Checks, id)
		if check == nil || !check.Pass {
			t.Errorf("%s = %+v, want pass", id, check)
		}
	}
	scheme := `<!DOCTYPE html><html><head><link rel="canonical" href="http://example.com/docs"></head><body><p>x</p></body></html>`
	https := `<!DOCTYPE html><html><head><link rel="canonical" href="https://example.com/docs"></head><body><p>x</p></body></html>`
	mismatch := CompareHydration("https://example.com/docs", scheme, https)
	if check := findCheck(mismatch.Checks, "seo_ssr_canonical_match"); check == nil || check.Pass {
		t.Fatalf("scheme change should fail, got %+v", check)
	}
}
