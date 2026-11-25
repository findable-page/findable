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

func TestCanonicalKeyTreatsSchemeCaseDefaultPortAndSlashAsTheSamePage(t *testing.T) {
	raw := "HTTPS://Example.com.:443/docs/"
	rendered := "https://example.com/docs#section"
	if canonicalKey(raw) != canonicalKey(rendered) {
		t.Fatalf("canonical keys %q and %q", canonicalKey(raw), canonicalKey(rendered))
	}
	if canonicalKey("https://[2001:db8::1]:443/a/") != canonicalKey("https://[2001:db8::1]/a") {
		t.Fatalf("ipv6 keys %q and %q", canonicalKey("https://[2001:db8::1]:443/a/"), canonicalKey("https://[2001:db8::1]/a"))
	}
	page := func(href string) string {
		return `<!DOCTYPE html><html><head><title>Same</title><link rel="canonical" href="` + href + `"></head><body><h1>Same</h1><p>Same words.</p></body></html>`
	}
	report := CompareHydration("https://example.com/docs", page(raw), page(rendered))
	if check := findCheck(report.Checks, "seo_ssr_canonical_match"); check == nil || !check.Pass {
		t.Fatalf("cosmetic canonical = %+v", check)
	}
	otherPort := CompareHydration("https://example.com/docs", page("https://example.com:8443/docs"), page("https://example.com/docs"))
	if check := findCheck(otherPort.Checks, "seo_ssr_canonical_match"); check == nil || check.Pass {
		t.Fatalf("non-default port should fail, got %+v", check)
	}
}

func TestCompareHydrationIgnoresSignalsTheRendererDidNotAdd(t *testing.T) {
	raw := `<!DOCTYPE html><html><head><title>Server title</title>
	<link rel="canonical" href="https://example.com/docs">
	</head><body><h1>Server</h1><h4>Only a deep heading in raw</h4><p>one two three four</p></body></html>`
	rendered := `<!DOCTYPE html><html><head></head><body><h4>Client only deep heading</h4><p>one two three four</p></body></html>`
	report := CompareHydration("https://example.com/docs", raw, rendered)
	for _, id := range []string{"seo_ssr_title_match", "seo_ssr_canonical_match", "seo_ssr_headings_present", "geo_ssr_content_parity"} {
		check := findCheck(report.Checks, id)
		if check == nil || !check.Pass {
			t.Errorf("%s = %+v, want pass", id, check)
		}
	}
}

func TestCompareHydrationContentParityBoundary(t *testing.T) {
	words := func(n int) string {
		return strings.TrimSpace(strings.Repeat("word ", n))
	}
	page := func(n int) string {
		return `<!DOCTYPE html><html><body><p>` + words(n) + `</p></body></html>`
	}
	half := CompareHydration("https://example.com", page(5), page(10))
	if half.TextRatio != 0.5 {
		t.Fatalf("text ratio = %v, want 0.5", half.TextRatio)
	}
	if check := findCheck(half.Checks, "geo_ssr_content_parity"); check == nil || !check.Pass {
		t.Fatalf("half the rendered text should pass, got %+v", check)
	}
	under := CompareHydration("https://example.com", page(4), page(10))
	if check := findCheck(under.Checks, "geo_ssr_content_parity"); check == nil || check.Pass {
		t.Fatalf("under half should fail, got %+v", check)
	}
}
