package audit

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

const edgeOrigin = "https://example.com/"

func edgeScan(raw string) Scan {
	return Analyze(edgeOrigin+"page", raw, "", make(http.Header), 200)
}

func edgeDoc(head, body string) string {
	return "<!DOCTYPE html><html><head>" + head + "</head><body>" + body + "</body></html>"
}

func checkApplicable(c *Check) bool {
	return c != nil && (c.Applicable == nil || *c.Applicable)
}

func TestEdgeInputsStayBounded(t *testing.T) {
	huge := "<p>" + strings.Repeat("café naïve — ", 4000) + "。</p>"
	cases := []struct {
		name, url, html string
	}{
		{name: "empty", html: ""},
		{name: "empty document", html: "<html></html>"},
		{name: "broken url", url: "http://[", html: "<html><body><p>Hi</p></body></html>"},
		{name: "garbage markup", html: "<<<<>>>>"},
		{name: "null byte", html: "<p>\x00café</p>"},
		{name: "unclosed tags", html: "<html><head><title>Café</title><body><article><h1>What is <b>naïve</b>?<p>Unclosed"},
		{name: "broken json-ld", html: `<script type="application/ld+json">{not json</script><p>Hi</p>`},
		{name: "huge unicode", html: huge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			url := tc.url
			if url == "" {
				url = edgeOrigin
			}
			started := time.Now()
			scan := Analyze(url, tc.html, "café", make(http.Header), 200)
			if elapsed := time.Since(started); elapsed > time.Second {
				t.Fatalf("took %s", elapsed)
			}
			if len(scan.AEO.Checks) == 0 || len(scan.GEO.Checks) == 0 {
				t.Fatal("expected AEO and GEO checks")
			}
		})
	}
}

func TestQuestionHeadingEdges(t *testing.T) {
	cases := []struct {
		name, heading string
		want          bool
	}{
		{name: "what", heading: "What is GEO?", want: true},
		{name: "ascii apostrophe", heading: "What's a canonical tag", want: true},
		{name: "unicode apostrophe", heading: "What’s a canonical tag", want: true},
		{name: "how to hyphen", heading: "How-to bake bread", want: true},
		{name: "why alone", heading: "Why", want: true},
		{name: "fullwidth question mark", heading: "Cost of indexing？", want: true},
		{name: "arabic question mark", heading: "هل هذا مفيد؟", want: true},
		{name: "question suffix on however", heading: "However?", want: true},
		{name: "however", heading: "However the team ships worldwide", want: false},
		{name: "whatever", heading: "Whatever you budget for it", want: false},
		{name: "whyte", heading: "Whyte and company history", want: false},
		{name: "wholesale", heading: "Wholesale pricing guide", want: false},
		{name: "empty heading", heading: "   ", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scan := edgeScan(edgeDoc("", "<h2>"+tc.heading+"</h2>"))
			c := findCheck(scan.AEO.Checks, "aeo_question_headings")
			if c == nil || c.Pass != tc.want {
				t.Fatalf("aeo_question_headings for %q = %+v, want pass %v", tc.heading, c, tc.want)
			}
		})
	}
}

func TestArticleHeadingIgnoresNonElements(t *testing.T) {
	cases := []struct {
		name, body string
		article    bool
	}{
		{name: "comment", body: `<!-- <article><h1>ghost</h1></article> --><article><p>No heading here.</p></article>`, article: false},
		{name: "script", body: `<script>var x="<article><h1>ghost</h1></article>";</script><article><p>No heading here.</p></article>`, article: false},
		{name: "noscript", body: `<noscript><article><h1>Fallback</h1></article></noscript><article><p>App shell.</p></article>`, article: false},
		{name: "attribute", body: `<div data-note="<article><h1>x</h1></article>"></div><article><p>No heading.</p></article>`, article: false},
		{name: "real article", body: `<article><h1>Title</h1><p>Post</p></article>`, article: true},
		{name: "heading outside", body: `<h1>Title</h1><article><p>Heading is outside.</p></article>`, article: false},
		{name: "two articles", body: `<article><h1>A</h1></article><article><p>B</p></article>`, article: false},
		{name: "empty article", body: `<article></article>`, article: false},
		{name: "unicode heading", body: `<article><h1>Café — pourquoi ?</h1><p>Updated: 4 March 2025</p></article>`, article: true},
		{name: "unclosed", body: `<article><h1>What is naïve?`, article: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := findCheck(edgeScan(edgeDoc("", tc.body)).GEO.Checks, "geo_date_published")
			if c == nil || checkApplicable(c) != tc.article {
				t.Fatalf("geo_date_published applicable=%v, want %v (%+v)", checkApplicable(c), tc.article, c)
			}
		})
	}
}

func TestPublishDatesRejectNonISO(t *testing.T) {
	cases := []struct {
		name, head                             string
		publishedPass, modifiedPass, orderPass bool
		publishedApp, modifiedApp, orderApp    bool
	}{
		{
			name:          "iso dates in order",
			head:          `<meta property="article:published_time" content="2025-01-02T00:00:00Z"><meta property="article:modified_time" content="2025-03-04">`,
			publishedPass: true, publishedApp: true,
			modifiedPass: true, modifiedApp: true,
			orderPass: true, orderApp: true,
		},
		{
			name:          "same calendar day",
			head:          `<meta property="article:published_time" content="2025-01-02"><meta property="article:modified_time" content="2025-01-02T23:00:00Z">`,
			publishedPass: true, publishedApp: true,
			modifiedPass: true, modifiedApp: true,
			orderPass: true, orderApp: true,
		},
		{
			name:          "modified earlier",
			head:          `<meta property="article:published_time" content="2025-03-04"><meta property="article:modified_time" content="2025-01-02">`,
			publishedPass: true, publishedApp: true,
			modifiedPass: true, modifiedApp: true,
			orderPass: false, orderApp: true,
		},
		{
			name:          "prose modified date",
			head:          `<meta property="article:published_time" content="2024-12-31"><meta property="article:modified_time" content="December 1, 2019">`,
			publishedPass: true, publishedApp: true,
			modifiedPass: false, modifiedApp: true,
			orderPass: true, orderApp: false,
		},
		{
			name:          "slash publish date",
			head:          `<meta property="article:published_time" content="2025/01/02"><meta property="article:modified_time" content="2025-03-04">`,
			publishedPass: false, publishedApp: true,
			modifiedPass: true, modifiedApp: true,
			orderPass: true, orderApp: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scan := edgeScan(edgeDoc(tc.head, "<p>x</p>"))
			assertCheck(t, scan.GEO.Checks, "geo_date_published", tc.publishedPass, tc.publishedApp)
			assertCheck(t, scan.GEO.Checks, "geo_date_modified", tc.modifiedPass, tc.modifiedApp)
			assertCheck(t, scan.GEO.Checks, "geo_modified_after_published", tc.orderPass, tc.orderApp)
		})
	}
}

func TestVisibleDateAllowsUnicodeSpace(t *testing.T) {
	body := `<article><h1>T</h1><p>Updated:&nbsp;March 4, 2025 and café notes.</p></article>`
	c := findCheck(edgeScan(edgeDoc("", body)).GEO.Checks, "geo_visible_date")
	if c == nil || !c.Pass || !checkApplicable(c) {
		t.Fatalf("geo_visible_date = %+v, want a pass", c)
	}
}

func TestTitleMatchUnicodeAndPunctuation(t *testing.T) {
	cases := []struct {
		name, title, h1 string
		want            bool
	}{
		{name: "punctuation", title: "Artisan, Bakery Dhaka!", h1: "Artisan Bakery in Dhaka", want: true},
		{name: "unicode", title: "Café, Bakery Dhaka!", h1: "Café Bakery in Dhaka", want: true},
		{name: "mismatch", title: "Products, Catalog!", h1: "Artisan Bakery in Dhaka", want: false},
		{name: "short tokens", title: "SEO, AI!", h1: "SEO AI tools", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			page := edgeDoc("<title>"+tc.title+"</title>", "<h1>"+tc.h1+"</h1>")
			c := findCheck(edgeScan(page).GEO.Checks, "geo_intent_clarity")
			if c == nil || !checkApplicable(c) || c.Pass != tc.want {
				t.Fatalf("geo_intent_clarity = %+v, want pass %v", c, tc.want)
			}
		})
	}
}

func TestSentenceLengthEdges(t *testing.T) {
	cases := []struct {
		name, body, avg string
	}{
		{name: "trailing period", body: "<p>Alpha beta.</p>", avg: "2.0"},
		{name: "decimal", body: "<p>Score improved by 78.4 percent.</p>", avg: "5.0"},
		{name: "exclamation", body: "<p>One two three! Four five</p>", avg: "2.5"},
		{name: "ideographic stop", body: "<p>" + strings.Repeat("word ", 19) + "word。" + strings.Repeat(" word", 10) + "</p>", avg: "15.0"},
		{name: "empty", body: "", avg: "0.0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := findCheck(edgeScan(edgeDoc("", tc.body)).AEO.Checks, "aeo_reading_level")
			want := "Average sentence length: " + tc.avg + " words"
			if c == nil || !strings.Contains(c.Label, want) {
				t.Fatalf("label %q, want it to contain %q", labelOf(c), want)
			}
		})
	}
}

func TestUnicodeSpacesCountAsSeparators(t *testing.T) {
	cases := []struct {
		name, body, id string
	}{
		{name: "definition", body: "<p>GEO&nbsp;is&nbsp;a&nbsp;discipline for answer engines.</p>", id: "geo_definitions"},
		{name: "comparison", body: "<p>Pricing&nbsp;vs&nbsp;quality matters here.</p>", id: "geo_comparison"},
		{name: "statistic", body: "<p>Saved&nbsp;$&nbsp;500 and cut latency by 78&nbsp;% today.</p>", id: "geo_statistics_density"},
		{name: "pronouns", body: "<p>When&nbsp;you start&nbsp;your project&nbsp;you should know that&nbsp;we help.</p>", id: "aeo_conversational"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scan := edgeScan(edgeDoc("", tc.body))
			checks := scan.GEO.Checks
			if tc.id == "aeo_conversational" {
				checks = scan.AEO.Checks
			}
			c := findCheck(checks, tc.id)
			if c == nil || !c.Pass {
				t.Fatalf("%s = %+v, want a pass", tc.id, c)
			}
		})
	}
}

func TestPronounsNeedARealSeparator(t *testing.T) {
	// A comma is not a space, so "you," still does not count as the pronoun.
	body := "<p>When you, start your, project you, should know that we, help you.</p>"
	c := findCheck(edgeScan(edgeDoc("", body)).AEO.Checks, "aeo_conversational")
	if c == nil || c.Pass {
		t.Fatalf("aeo_conversational = %+v, want a fail", c)
	}
}

func TestUnicodeHostCountsAsASource(t *testing.T) {
	body := `<article><h1>T</h1>` +
		`<a href="https://münchen.de/a">a</a>` +
		`<a href="https://example.org/b">b</a>` +
		`<a href="https://nih.gov/c">c</a></article>`
	scan := edgeScan(edgeDoc("", body))
	div := findCheck(scan.GEO.Checks, "geo_source_diversity")
	if div == nil || !checkApplicable(div) || !div.Pass {
		t.Fatalf("geo_source_diversity = %+v, want a pass", div)
	}
	auth := findCheck(scan.GEO.Checks, "geo_authoritative_links")
	if auth == nil || !auth.Pass {
		t.Fatalf("geo_authoritative_links = %+v, want a pass", auth)
	}
}

func assertCheck(t *testing.T, checks []Check, id string, pass, applicable bool) {
	t.Helper()
	c := findCheck(checks, id)
	if c == nil || c.Pass != pass || checkApplicable(c) != applicable {
		t.Fatalf("%s = %+v, want pass=%v applicable=%v", id, c, pass, applicable)
	}
}

func labelOf(c *Check) string {
	if c == nil {
		return "<nil>"
	}
	return c.Label
}
