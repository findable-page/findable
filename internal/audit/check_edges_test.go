package audit

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestSameSiteDefaultPortIsNotAnExternalCitation(t *testing.T) {
	rhet := strings.Repeat("This passage argues a position with purely rhetorical language and no facts at all whatsoever. ", 8)
	for _, href := range []string{
		"https://example.com:443/about",
		"https://EXAMPLE.com/about",
		"https://example.com./about",
		"javascript:void(0)",
		"mailto:ada@example.com",
		"#section",
	} {
		doc := "<!DOCTYPE html><html><body><h2>Opinion</h2><p>" + rhet +
			"<a href='" + href + "'>more</a></p></body></html>"
		res := Analyze("https://example.com/page", doc, "", make(http.Header), 200)
		c := findCheck(res.GEO.Checks, "geo_assertion_evidence_pairing")
		if c.Pass {
			t.Errorf("href %s counted as an external citation, check %+v", href, c)
		}
	}

	external := "<!DOCTYPE html><html><body><h2>Opinion</h2><p>" + rhet +
		"<a href='https://research.example/paper'>source</a></p></body></html>"
	res := Analyze("https://example.com/page", external, "", make(http.Header), 200)
	if c := findCheck(res.GEO.Checks, "geo_assertion_evidence_pairing"); !c.Pass {
		t.Fatalf("a real external link should satisfy the section, got %+v", c)
	}
}

func TestIsExternalHrefMatchesClassifyLink(t *testing.T) {
	base, err := url.Parse("https://example.com/page")
	if err != nil {
		t.Fatal(err)
	}
	sig := &PageSignals{URL: base}
	cases := []struct {
		href string
		want bool
	}{
		{"https://example.com:443/about", false},
		{"https://EXAMPLE.com./blog", false},
		{"https://example.com:8443/admin", true},
		{"https://other.org/", true},
		{"/relative", false},
		{"javascript:void(0)", false},
		{"mailto:ada@example.com", false},
		{"tel:+15551212", false},
		{"#top", false},
		{"%zz", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isExternalHref(sig, tc.href); got != tc.want {
			t.Errorf("isExternalHref(%q) = %v, want %v", tc.href, got, tc.want)
		}
	}
	if isExternalHref(&PageSignals{}, "https://other.org/") {
		t.Fatal("nil page URL should not count as external")
	}
}

func TestAuthorCredentialAcceptsPortAndWWW(t *testing.T) {
	page := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":"https://schema.org","@type":"Person","name":"Ada","sameAs":"https://www.orcid.org:443/0000-0001"}
	</script></head></html>`
	res := Analyze("https://example.com", page, "", make(http.Header), 200)
	c := findCheck(res.GEO.Checks, "geo_author_credential_depth")
	if !c.Pass || (c.Applicable != nil && !*c.Applicable) {
		t.Fatalf("orcid host with a port and www should count, got %+v", c)
	}
}

func TestSchemaIDStringRefs(t *testing.T) {
	plain := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":"https://schema.org","@type":"Article","author":"Ada Lovelace"}
	</script></head></html>`
	res := Analyze("https://example.com", plain, "", make(http.Header), 200)
	c := findCheck(res.GEO.Checks, "geo_schema_id_graph_coherence")
	if c.Applicable == nil || *c.Applicable {
		t.Fatalf("a plain-text author is not an @id reference, got %+v", c)
	}

	blank := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":"https://schema.org","@type":"Article","author":"  "}
	</script></head></html>`
	res = Analyze("https://example.com", blank, "", make(http.Header), 200)
	c = findCheck(res.GEO.Checks, "geo_schema_id_graph_coherence")
	if c.Applicable == nil || *c.Applicable {
		t.Fatalf("a blank author string is not an @id reference, got %+v", c)
	}

	fragment := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":"https://schema.org","@type":"Article","publisher":"#missing"}
	</script></head></html>`
	res = Analyze("https://example.com", fragment, "", make(http.Header), 200)
	c = findCheck(res.GEO.Checks, "geo_schema_id_graph_coherence")
	if c.Pass || len(c.Evidence) != 1 || c.Evidence[0].Value != "#missing" {
		t.Fatalf("unresolved fragment @id should fail, got %+v", c)
	}
}

func TestBrandNameEdgesAndEmptyTitleSide(t *testing.T) {
	suffix := `<!DOCTYPE html><html><head>
	<meta property="og:site_name" content="Widgets">
	<script type="application/ld+json">{"@type":"Organization","name":"Acme Widgets"}</script>
	</head></html>`
	res := Analyze("https://example.com", suffix, "", make(http.Header), 200)
	if c := findCheck(res.GEO.Checks, "geo_brand_entity_consistency"); !c.Pass {
		t.Fatalf("a brand that is a whole suffix should agree, got %+v", c)
	}

	stuck := `<!DOCTYPE html><html><head>
	<meta property="og:site_name" content="Acme">
	<script type="application/ld+json">{"@type":"Organization","name":"AcmeX"}</script>
	</head></html>`
	res = Analyze("https://example.com", stuck, "", make(http.Header), 200)
	if c := findCheck(res.GEO.Checks, "geo_brand_entity_consistency"); c.Pass {
		t.Fatal("Acme should not match AcmeX")
	}

	emptySide := `<!DOCTYPE html><html><head>
	<title>Pricing | </title>
	<meta property="og:site_name" content="Acme">
	</head></html>`
	res = Analyze("https://example.com", emptySide, "", make(http.Header), 200)
	c := findCheck(res.GEO.Checks, "geo_brand_entity_consistency")
	if c.Applicable == nil || *c.Applicable {
		t.Fatalf("a title with an empty side is not a second brand, got %+v", c)
	}
}

func TestHowToStepShapes(t *testing.T) {
	valid := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":"https://schema.org","@type":"HowTo","name":"Bake","step":["Mix","Bake"]}
	</script></head></html>`
	res := Analyze("https://example.com", valid, "", make(http.Header), 200)
	if c := findCheck(res.AEO.Checks, "aeo_howto_valid"); !c.Pass {
		t.Fatalf("string steps should pass, got %+v", c)
	}

	emptyStep := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":"https://schema.org","@type":"HowTo","name":"Bake","step":["Mix",""]}
	</script></head></html>`
	res = Analyze("https://example.com", emptyStep, "", make(http.Header), 200)
	if c := findCheck(res.AEO.Checks, "aeo_howto_valid"); c.Pass {
		t.Fatal("an empty step string should fail")
	}

	nameless := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":"https://schema.org","@type":"HowTo","name":"Bake","step":[{"@type":"HowToStep"},{"text":"Bake"}]}
	</script></head></html>`
	res = Analyze("https://example.com", nameless, "", make(http.Header), 200)
	if c := findCheck(res.AEO.Checks, "aeo_howto_valid"); c.Pass {
		t.Fatal("a step object without text, name or itemListElement should fail")
	}
}

func TestJSONLDContextShapesAndOfferPriceSpecification(t *testing.T) {
	arrayCtx := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":["https://schema.org",{"@language":"en"}],"@type":"Organization","name":"Acme"}
	</script></head></html>`
	res := Analyze("https://example.com", arrayCtx, "", make(http.Header), 200)
	if c := findCheck(res.SEO.Checks, "jsonld_context"); !c.Pass {
		t.Fatalf("array @context should pass, got %+v", c)
	}

	vocab := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":{"@vocab":"https://schema.org/"},"@type":"Organization","name":"Acme"}
	</script></head></html>`
	res = Analyze("https://example.com", vocab, "", make(http.Header), 200)
	if c := findCheck(res.SEO.Checks, "jsonld_context"); !c.Pass {
		t.Fatalf("@vocab context should pass, got %+v", c)
	}

	spec := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":"https://schema.org","@type":"Product","name":"Widget","offers":{"priceSpecification":{"price":"10","priceCurrency":"USD"}}}
	</script></head></html>`
	res = Analyze("https://example.com", spec, "", make(http.Header), 200)
	if c := findCheck(res.SEO.Checks, "product_price"); !c.Pass {
		t.Fatalf("priceSpecification should satisfy product_price, got %+v", c)
	}

	bare := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":"https://schema.org","@type":"Product","name":"Widget","offers":"free"}
	</script></head></html>`
	res = Analyze("https://example.com", bare, "", make(http.Header), 200)
	if c := findCheck(res.SEO.Checks, "product_price"); c.Pass {
		t.Fatal("a string offer is not a price")
	}

	noCurrency := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":"https://schema.org","@type":"Product","name":"Widget","offers":{"price":"10"}}
	</script></head></html>`
	res = Analyze("https://example.com", noCurrency, "", make(http.Header), 200)
	if c := findCheck(res.SEO.Checks, "product_price"); c.Pass {
		t.Fatal("a price without a currency should fail")
	}
}

func TestOGURLRejectsAnUnparseableCanonical(t *testing.T) {
	page := `<!DOCTYPE html><html><head>
	<meta property="og:url" content="https://example.com/%zz">
	<link rel="canonical" href="https://example.com/ok">
	</head></html>`
	res := Analyze("https://example.com/ok", page, "", make(http.Header), 200)
	c := findCheck(res.SEO.Checks, "og_url_matches_canonical")
	if c.Applicable != nil && !*c.Applicable {
		t.Fatal("both URLs are present, so the check applies")
	}
	if c.Pass {
		t.Fatal("an unparseable og:url should not match the canonical")
	}
}

func TestEmptyIDAttributesAreNotDuplicates(t *testing.T) {
	page := `<!DOCTYPE html><html><body><div id=""></div><div id=""></div><p id="once">x</p></body></html>`
	res := Analyze("https://example.com", page, "", make(http.Header), 200)
	if c := findCheck(res.SEO.Checks, "duplicate_ids"); !c.Pass {
		t.Fatalf("blank ids should not count as duplicates, got %+v", c)
	}
}

func TestVisibleTextDropsNonContent(t *testing.T) {
	page := `<!DOCTYPE html><html><head><title>Hello</title><style>.x{color:red}</style><script>secret()</script></head><body><svg><title>icon</title></svg><p>Visible words</p><noscript>enable js</noscript></body></html>`
	got := VisibleText(page)
	for _, hidden := range []string{"secret", "color", "icon", "enable"} {
		if strings.Contains(got, hidden) {
			t.Fatalf("VisibleText = %q, still contains %q", got, hidden)
		}
	}
	if !strings.Contains(got, "Hello") || !strings.Contains(got, "Visible words") {
		t.Fatalf("VisibleText = %q, want the title and paragraph", got)
	}
}

func TestBotByName(t *testing.T) {
	bot, ok := BotByName("gptbot")
	if !ok || !strings.EqualFold(bot.Name, "gptbot") {
		t.Fatalf("BotByName(gptbot) = %+v, %v", bot, ok)
	}
	if _, ok := BotByName("NotABot"); ok {
		t.Fatal("unknown bot name should miss")
	}
}

func TestQueryFanoutBlankTopicUsesAPlaceholder(t *testing.T) {
	report := QueryFanout("page.html", "<html><body></body></html>", "   ")
	if report.Topic != "" {
		t.Fatalf("stored topic = %q, want blank", report.Topic)
	}
	if len(report.Intents) == 0 || report.Intents[0].Query != "what is this topic" {
		t.Fatalf("definition query = %+v, want the placeholder topic", report.Intents)
	}
}

func TestHydrationRatioWhenRenderedTextIsEmpty(t *testing.T) {
	raw := `<!DOCTYPE html><html><body><p>Some words here.</p></body></html>`
	rendered := `<!DOCTYPE html><html><body></body></html>`
	report := CompareHydration("https://example.com", raw, rendered)
	if report.RenderedWords != 0 || report.TextRatio != 1 {
		t.Fatalf("words %d ratio %v, want 0 and 1", report.RenderedWords, report.TextRatio)
	}
	if c := findCheck(report.Checks, "geo_ssr_content_parity"); c == nil || !c.Pass {
		t.Fatalf("empty rendered text should pass parity, got %+v", c)
	}
}

func TestHydrationIgnoresHeadingsBelowH3(t *testing.T) {
	raw := `<!DOCTYPE html><html><body><h1>Same</h1></body></html>`
	rendered := `<!DOCTYPE html><html><body><h1>Same</h1><h4>Only after render</h4></body></html>`
	report := CompareHydration("https://example.com", raw, rendered)
	if c := findCheck(report.Checks, "seo_ssr_headings_present"); c == nil || !c.Pass {
		t.Fatalf("h4 is outside the h1-h3 comparison, got %+v", c)
	}
}
