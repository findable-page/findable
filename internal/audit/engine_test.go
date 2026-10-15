package audit

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
)

func findCheck(checks []Check, id string) *Check {
	for i := range checks {
		if checks[i].ID == id {
			return &checks[i]
		}
	}
	return nil
}

func TestSignalsEmptyAndMalformedHTML(t *testing.T) {
	// Empty HTML
	sig := extractSignals("https://example.com", "", make(http.Header), 200)
	if sig.Title != "" || sig.H1Count != 0 || len(sig.Words) != 0 {
		t.Errorf("empty html should yield empty signals, got: %+v", sig)
	}

	// Malformed unclosed tags
	malformed := `<html><head><title>Unclosed Tag Test</head><body><div><p><span>Hello World<!-- unclosed comment`
	sig2 := extractSignals("https://example.com", malformed, make(http.Header), 200)
	if sig2.Title != "Unclosed Tag Test" {
		t.Errorf("expected title to be parsed despite malformed markup, got: %q", sig2.Title)
	}
	if len(sig2.Words) == 0 {
		t.Errorf("expected words to be extracted from unclosed tags")
	}

	// Unicode and emoji title
	unicodeHTML := `<!DOCTYPE html><html><head><title>🚀 Performance & Growth — 2026</title></head></html>`
	sig3 := extractSignals("https://example.com", unicodeHTML, make(http.Header), 200)
	if !strings.Contains(sig3.Title, "🚀") {
		t.Errorf("expected emoji to be preserved in title, got: %q", sig3.Title)
	}
}

func TestSEOTitleLengthBoundaries(t *testing.T) {
	tests := []struct {
		titleLen int
		wantPass bool
	}{
		{0, false},
		{29, false}, // 1 below lower bound (30)
		{30, true},  // lower bound
		{45, true},  // middle
		{60, true},  // upper bound
		{61, false}, // 1 above upper bound (60)
		{120, false},
	}

	for _, tc := range tests {
		titleStr := strings.Repeat("a", tc.titleLen)
		raw := fmt.Sprintf("<!DOCTYPE html><html><head><title>%s</title></head></html>", titleStr)
		res := Analyze("https://example.com", raw, "", make(http.Header), 200)
		c := findCheck(res.SEO.Checks, "title_length")
		if c == nil {
			t.Fatalf("check title_length not found")
		}
		if c.Pass != tc.wantPass {
			t.Errorf("title length %d: got pass=%v, want %v", tc.titleLen, c.Pass, tc.wantPass)
		}
	}
}

func TestSEOMetaDescBoundaries(t *testing.T) {
	tests := []struct {
		descLen  int
		wantPass bool
	}{
		{0, false},
		{119, false}, // 1 below 120
		{120, true},  // lower bound
		{140, true},  // optimal
		{160, true},  // upper bound
		{161, false}, // 1 above 160
		{200, false},
	}

	for _, tc := range tests {
		descStr := strings.Repeat("b", tc.descLen)
		raw := fmt.Sprintf(`<!DOCTYPE html><html><head><meta name="description" content="%s"></head></html>`, descStr)
		res := Analyze("https://example.com", raw, "", make(http.Header), 200)
		c := findCheck(res.SEO.Checks, "desc_length")
		if c == nil {
			t.Fatalf("check desc_length not found")
		}
		if c.Pass != tc.wantPass {
			t.Errorf("desc length %d: got pass=%v, want %v", tc.descLen, c.Pass, tc.wantPass)
		}
	}
}

func TestSEOHeadingHierarchyRules(t *testing.T) {
	// Normal: H1 -> H2 -> H3
	validHTML := `<!DOCTYPE html><html><body><h1>Title</h1><h2>Section</h2><h3>Sub</h3></body></html>`
	res1 := Analyze("https://example.com", validHTML, "", make(http.Header), 200)
	c1 := findCheck(res1.SEO.Checks, "heading_hierarchy")
	h1Check1 := findCheck(res1.SEO.Checks, "h1_exists")
	if !c1.Pass || !h1Check1.Pass {
		t.Errorf("expected valid hierarchy and single H1 to pass")
	}

	// Skipped hierarchy: H1 -> H3 directly (skipping H2)
	skippedHTML := `<!DOCTYPE html><html><body><h1>Title</h1><h3>Skipped to Sub</h3></body></html>`
	res2 := Analyze("https://example.com", skippedHTML, "", make(http.Header), 200)
	c2 := findCheck(res2.SEO.Checks, "heading_hierarchy")
	if c2.Pass {
		t.Errorf("expected skipped hierarchy (H1 -> H3) to fail")
	}

	// Multiple H1s
	multiH1HTML := `<!DOCTYPE html><html><body><h1>Title 1</h1><h1>Title 2</h1></body></html>`
	res3 := Analyze("https://example.com", multiH1HTML, "", make(http.Header), 200)
	h1Check3 := findCheck(res3.SEO.Checks, "h1_exists")
	if h1Check3.Pass {
		t.Errorf("expected multiple H1s to fail h1_exists")
	}
}

func TestSEOLinkSafetyAndOpener(t *testing.T) {
	// Safe blank link
	safeHTML := `<!DOCTYPE html><html><body><a href="https://external.com" target="_blank" rel="noopener noreferrer">External</a></body></html>`
	res1 := Analyze("https://example.com", safeHTML, "", make(http.Header), 200)
	c1 := findCheck(res1.SEO.Checks, "link_noopener")
	if !c1.Pass {
		t.Errorf("expected safe link with rel='noopener' to pass link_noopener")
	}

	// Unsafe blank link
	unsafeHTML := `<!DOCTYPE html><html><body><a href="https://external.com" target="_blank">External</a></body></html>`
	res2 := Analyze("https://example.com", unsafeHTML, "", make(http.Header), 200)
	c2 := findCheck(res2.SEO.Checks, "link_noopener")
	if c2.Pass {
		t.Errorf("expected blank link missing rel='noopener' to fail link_noopener")
	}

	// Empty hrefs
	emptyHrefHTML := `<!DOCTYPE html><html><body><a href="#">Link</a><a href="">Empty</a></body></html>`
	res3 := Analyze("https://example.com", emptyHrefHTML, "", make(http.Header), 200)
	c3 := findCheck(res3.SEO.Checks, "empty_hrefs")
	if c3.Pass {
		t.Errorf("expected empty hrefs to fail empty_hrefs check")
	}
}

func TestSEOImageAltAndDimensions(t *testing.T) {
	// Image with alt, width, height
	goodImgHTML := `<!DOCTYPE html><html><body><img src="pic.jpg" alt="Description" width="300" height="200"></body></html>`
	res1 := Analyze("https://example.com", goodImgHTML, "", make(http.Header), 200)
	altC1 := findCheck(res1.SEO.Checks, "img_alt")
	dimC1 := findCheck(res1.SEO.Checks, "img_dimensions")
	if !altC1.Pass || !dimC1.Pass {
		t.Errorf("expected good image to pass alt and dimensions checks")
	}

	// Image missing alt
	noAltHTML := `<!DOCTYPE html><html><body><img src="pic.jpg" width="300" height="200"></body></html>`
	res2 := Analyze("https://example.com", noAltHTML, "", make(http.Header), 200)
	altC2 := findCheck(res2.SEO.Checks, "img_alt")
	if altC2.Pass {
		t.Errorf("expected image missing alt to fail img_alt check")
	}

	// Image missing dimensions
	noDimHTML := `<!DOCTYPE html><html><body><img src="pic.jpg" alt="Valid Alt"></body></html>`
	res3 := Analyze("https://example.com", noDimHTML, "", make(http.Header), 200)
	dimC3 := findCheck(res3.SEO.Checks, "img_dimensions")
	if dimC3.Pass {
		t.Errorf("expected image missing width/height to fail img_dimensions check")
	}
}

func TestAEOFAQAndAnswerBoxRules(t *testing.T) {
	// With FAQPage schema
	faqHTML := `<!DOCTYPE html><html><head>
	<script type="application/ld+json">
	{"@context":"https://schema.org","@type":"FAQPage","mainEntity":[]}
	</script>
	</head><body>
	<h2>What is Cloud Cost Tuning?</h2>
	<p>` + strings.Repeat("word ", 45) + `</p>
	</body></html>`

	res := Analyze("https://example.com", faqHTML, "", make(http.Header), 200)
	faqCheck := findCheck(res.AEO.Checks, "aeo_faq")
	qHeadingCheck := findCheck(res.AEO.Checks, "aeo_question_headings")
	answerBoxCheck := findCheck(res.AEO.Checks, "aeo_direct_answers")

	if !faqCheck.Pass {
		t.Errorf("expected FAQPage schema to pass aeo_faq")
	}
	if !qHeadingCheck.Pass {
		t.Errorf("expected 'What is...' heading to pass aeo_question_headings")
	}
	if !answerBoxCheck.Pass {
		t.Errorf("expected 45-word paragraph under heading to pass aeo_direct_answers")
	}
}

func TestAEOStructuralElements(t *testing.T) {
	structHTML := `<!DOCTYPE html><html><body>
	<dl><dt>Term</dt><dd>Definition of term</dd></dl>
	<ol><li>Step 1</li><li>Step 2</li></ol>
	<table><tr><th>Header</th></tr><tr><td>Row</td></tr></table>
	<pre><code>curl -X GET https://example.com</code></pre>
	<details><summary>Toggle</summary><p>Hidden</p></details>
	</body></html>`

	res := Analyze("https://example.com", structHTML, "", make(http.Header), 200)

	for _, checkID := range []string{
		"aeo_definition_lists",
		"aeo_numbered_steps",
		"aeo_tables",
		"aeo_code_blocks",
		"aeo_collapsible",
	} {
		c := findCheck(res.AEO.Checks, checkID)
		if c == nil || !c.Pass {
			t.Errorf("expected %s to pass, got %v", checkID, c)
		}
	}
}

func TestAEOConversationalTone(t *testing.T) {
	convHTML := `<!DOCTYPE html><html><body>
	<p>When you start your project, you should ensure that we understand your goals.</p>
	</body></html>`

	res := Analyze("https://example.com", convHTML, "", make(http.Header), 200)
	c := findCheck(res.AEO.Checks, "aeo_conversational")
	if !c.Pass {
		t.Errorf("expected pronoun-rich text to pass aeo_conversational")
	}

	formalHTML := `<!DOCTYPE html><html><body>
	<p>The system execution occurs upon initialization of the operational state.</p>
	</body></html>`
	resFormal := Analyze("https://example.com", formalHTML, "", make(http.Header), 200)
	cF := findCheck(resFormal.AEO.Checks, "aeo_conversational")
	if cF.Pass {
		t.Errorf("expected formal text without pronouns to fail aeo_conversational")
	}
}

func TestGEOStatisticsDensityAndQuotes(t *testing.T) {
	geoHTML := `<!DOCTYPE html><html><body>
	<p>Our benchmark showed 78.4% improvement, latency reduced by 250ms, and saved $500 monthly across 300 users.</p>
	<blockquote>"Generative search requires empirical proof." — Expert Scientist</blockquote>
	<a href="https://arxiv.org/abs/1234.5678">Primary Research Study</a>
	</body></html>`

	res := Analyze("https://example.com", geoHTML, "", make(http.Header), 200)

	statC := findCheck(res.GEO.Checks, "geo_statistics_density")
	if !statC.Pass {
		t.Errorf("expected multiple statistics (%%, ms, $, users) to pass geo_statistics_density")
	}

	quoteC := findCheck(res.GEO.Checks, "geo_expert_quotes")
	if !quoteC.Pass {
		t.Errorf("expected blockquote to pass geo_expert_quotes")
	}

	citC := findCheck(res.GEO.Checks, "geo_citations")
	if !citC.Pass {
		t.Errorf("expected external link to pass geo_citations")
	}
}

func TestGEOKnowledgeGraphAndTrustSignals(t *testing.T) {
	kgHTML := `<!DOCTYPE html><html><head>
	<script type="application/ld+json">
	{
		"@context": "https://schema.org",
		"@type": "Organization",
		"name": "Global Tech",
		"sameAs": ["https://www.wikidata.org/wiki/Q12345", "https://linkedin.com/company/tech"]
	}
	</script>
	</head><body>
	<nav>
		<a href="/about.html">About Us</a>
		<a href="/contact.html">Contact Us</a>
	</nav>
	<p>By Alex Smith | <time datetime="2026-10-10">October 10, 2026</time></p>
	</body></html>`

	res := Analyze("https://example.com", kgHTML, "", make(http.Header), 200)

	kgCheck := findCheck(res.GEO.Checks, "geo_knowledge_graph")
	if !kgCheck.Pass {
		t.Errorf("expected sameAs links to pass geo_knowledge_graph")
	}

	trustCheck := findCheck(res.GEO.Checks, "geo_trust")
	if !trustCheck.Pass {
		t.Errorf("expected about and contact links to pass geo_trust")
	}

	timeCheck := findCheck(res.GEO.Checks, "geo_time_element")
	if !timeCheck.Pass {
		t.Errorf("expected <time> tag to pass geo_time_element")
	}
}

func TestLocalAuditEngineOnThinPage(t *testing.T) {
	htmlBytes, err := os.ReadFile("../../testdata/site/01-thin-poor.html")
	if err != nil {
		t.Fatal(err)
	}

	scanResult := Analyze("http://localhost/01-thin-poor.html", string(htmlBytes), "", make(http.Header), 200)
	ov := Overall(scanResult)

	if ov > 45 {
		t.Errorf("expected thin page to score low, got overall=%d", ov)
	}
	if scanResult.SEO.Score == nil || *scanResult.SEO.Score > 50 {
		t.Errorf("expected low SEO score, got %v", *scanResult.SEO.Score)
	}
}

func TestLocalAuditEngineOnGEOResearchPage(t *testing.T) {
	htmlBytes, err := os.ReadFile("../../testdata/site/04-geo-research.html")
	if err != nil {
		t.Fatal(err)
	}

	headers := make(http.Header)
	headers.Set("Strict-Transport-Security", "max-age=31536000")
	headers.Set("X-Content-Type-Options", "nosniff")
	headers.Set("X-Frame-Options", "DENY")
	headers.Set("Content-Security-Policy", "default-src 'self'")

	scanResult := Analyze("https://example.com/04-geo-research.html", string(htmlBytes), "", headers, 200)
	ov := Overall(scanResult)

	if ov < 65 {
		t.Errorf("expected rich research page to score >=65, got overall=%d", ov)
	}
	if scanResult.GEO.Score == nil || *scanResult.GEO.Score < 60 {
		t.Errorf("expected high GEO score, got %v", scanResult.GEO.Score)
	}
}

func TestLocalAuditEngineAll10Archetypes(t *testing.T) {
	pages := []struct {
		name     string
		minScore int
		maxScore int
	}{
		{"01-thin-poor.html", 10, 45},
		{"02-standard-blog.html", 50, 95},
		{"03-faq-aeo.html", 50, 95},
		{"04-geo-research.html", 65, 100},
		{"05-ecommerce-product.html", 55, 95},
		{"06-tech-docs.html", 50, 95},
		{"07-local-business.html", 50, 95},
		{"08-news-press.html", 50, 95},
		{"09-voice-conversational.html", 50, 95},
		{"10-landing-rich.html", 55, 98},
	}

	for _, p := range pages {
		b, err := os.ReadFile("../../testdata/site/" + p.name)
		if err != nil {
			continue
		}
		res := Analyze("https://example.com/"+p.name, string(b), "", make(http.Header), 200)
		score := Overall(res)
		if score < p.minScore || score > p.maxScore {
			t.Errorf("%s: score %d outside expected range [%d, %d]", p.name, score, p.minScore, p.maxScore)
		}
	}
}
