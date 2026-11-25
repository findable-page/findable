package audit

import (
	"net/http"
	"testing"
)

func TestSchemaIDGraphResolvesReferences(t *testing.T) {
	linked := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":"https://schema.org","@graph":[
	  {"@type":"Article","author":{"@id":"https://example.com/#author"}},
	  {"@type":"Person","@id":"https://example.com/#author","name":"Ada"}
	]}</script></head></html>`
	res := Analyze("https://example.com", linked, "", make(http.Header), 200)
	if c := findCheck(res.GEO.Checks, "geo_schema_id_graph_coherence"); !c.Pass || (c.Applicable != nil && !*c.Applicable) {
		t.Fatalf("resolved @id should pass, got %+v", c)
	}

	dangling := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@context":"https://schema.org","@graph":[
	  {"@type":"Article","publisher":{"@id":"https://example.com/#missing"}}
	]}</script></head></html>`
	res = Analyze("https://example.com", dangling, "", make(http.Header), 200)
	c := findCheck(res.GEO.Checks, "geo_schema_id_graph_coherence")
	if c.Pass || len(c.Evidence) != 1 || c.Evidence[0].Value != "https://example.com/#missing" {
		t.Fatalf("dangling @id should fail with evidence, got %+v", c)
	}

	inline := `<!DOCTYPE html><html><head><script type="application/ld+json">
	{"@type":"Article","author":{"@type":"Person","name":"Ada"}}</script></head></html>`
	res = Analyze("https://example.com", inline, "", make(http.Header), 200)
	c = findCheck(res.GEO.Checks, "geo_schema_id_graph_coherence")
	if c.Applicable == nil || *c.Applicable {
		t.Fatal("inline author without @id should be not applicable")
	}
}

func TestBrandEntityConsistency(t *testing.T) {
	agree := `<!DOCTYPE html><html><head>
	<title>Pricing | Acme Widgets</title>
	<meta property="og:site_name" content="Acme">
	<script type="application/ld+json">{"@type":"Organization","name":"Acme Widgets"}</script>
	</head></html>`
	res := Analyze("https://example.com", agree, "", make(http.Header), 200)
	if c := findCheck(res.GEO.Checks, "geo_brand_entity_consistency"); !c.Pass {
		t.Fatalf("contained brand names should pass, got %+v", c)
	}

	disagree := `<!DOCTYPE html><html><head>
	<meta property="og:site_name" content="Acme">
	<script type="application/ld+json">{"@type":"Organization","name":"Globex"}</script>
	</head></html>`
	res = Analyze("https://example.com", disagree, "", make(http.Header), 200)
	c := findCheck(res.GEO.Checks, "geo_brand_entity_consistency")
	if c.Pass || len(c.Evidence) == 0 {
		t.Fatalf("divergent brand names should fail, got %+v", c)
	}

	embedded := `<!DOCTYPE html><html><head>
	<meta property="og:site_name" content="AI">
	<script type="application/ld+json">{"@type":"Organization","name":"Email Weekly"}</script>
	</head></html>`
	res = Analyze("https://example.com", embedded, "", make(http.Header), 200)
	c = findCheck(res.GEO.Checks, "geo_brand_entity_consistency")
	if c.Pass {
		t.Fatal("a short brand inside another word should not count as the same name")
	}

	headline := `<!DOCTYPE html><html><head><title>CleanPower announces a new plant</title>
	<script type="application/ld+json">{"@type":"Organization","name":"Atlantic Energy Tribune"}</script>
	</head></html>`
	res = Analyze("https://example.com", headline, "", make(http.Header), 200)
	c = findCheck(res.GEO.Checks, "geo_brand_entity_consistency")
	if c.Applicable == nil || *c.Applicable {
		t.Fatal("a headline title without a site suffix is not a second brand signal")
	}
}

func TestSchemaIDGraphResolvesBareAndRelativeURIs(t *testing.T) {
	page := func(jsonld string) string {
		return `<!DOCTYPE html><html><head><script type="application/ld+json">` + jsonld + `</script></head></html>`
	}
	resolved := Analyze("https://example.com", page(`{"@graph":[
		{"@type":"Article","author":"/authors/ada"},
		{"@type":"Person","@id":" /authors/ada ","name":"Ada"}
	]}`), "", make(http.Header), 200)
	if c := findCheck(resolved.GEO.Checks, "geo_schema_id_graph_coherence"); c == nil || !c.Pass || (c.Applicable != nil && !*c.Applicable) {
		t.Fatalf("relative author URI should resolve, got %+v", c)
	}

	dangling := Analyze("https://example.com", page(`{"@type":"Article","author":"/authors/ada"}`), "", make(http.Header), 200)
	c := findCheck(dangling.GEO.Checks, "geo_schema_id_graph_coherence")
	if c == nil || c.Pass || c.Applicable != nil && !*c.Applicable || len(c.Evidence) != 1 || c.Evidence[0].Value != "/authors/ada" {
		t.Fatalf("dangling relative author URI = %+v", c)
	}

	name := Analyze("https://example.com", page(`{"@type":"Article","author":"Ada Lovelace"}`), "", make(http.Header), 200)
	c = findCheck(name.GEO.Checks, "geo_schema_id_graph_coherence")
	if c == nil || c.Applicable == nil || *c.Applicable {
		t.Fatalf("a plain author name is not an @id reference, got %+v", c)
	}

	dup := Analyze("https://example.com", page(`{"@graph":[
		{"@type":"Article","author":{"@id":" https://example.com/#missing "}},
		{"@type":"NewsArticle","publisher":{"@id":"https://example.com/#missing"}}
	]}`), "", make(http.Header), 200)
	c = findCheck(dup.GEO.Checks, "geo_schema_id_graph_coherence")
	if c == nil || c.Pass || len(c.Evidence) != 1 || c.Evidence[0].Value != "https://example.com/#missing" {
		t.Fatalf("duplicate dangling @id = %+v", c)
	}
}

func TestAuthoritativeSameAsAcceptsRegistryHostsOnly(t *testing.T) {
	page := func(sameAs string) string {
		return `<!DOCTYPE html><html><head><script type="application/ld+json">{"@type":"Organization","name":"Acme","sameAs":` + sameAs + `}</script></head></html>`
	}
	pass := []string{
		`["https://en.wikipedia.org/wiki/Acme"]`,
		`["https://www.crunchbase.com/organization/acme"]`,
		`["//www.wikidata.org/wiki/Q1"]`,
		`["https://en.wikipedia.org:443/wiki/Acme"]`,
	}
	for _, sameAs := range pass {
		res := Analyze("https://example.com", page(sameAs), "", make(http.Header), 200)
		if c := findCheck(res.GEO.Checks, "geo_entity_authoritative_sameas"); c == nil || !c.Pass {
			t.Errorf("%s should pass, got %+v", sameAs, c)
		}
	}
	fail := []string{
		`["https://notwikipedia.org/wiki/Acme"]`,
		`["https://wikipedia.org.example/wiki/Acme"]`,
		`["https://crunchbase.com.evil.test/organization/acme"]`,
	}
	for _, sameAs := range fail {
		res := Analyze("https://example.com", page(sameAs), "", make(http.Header), 200)
		if c := findCheck(res.GEO.Checks, "geo_entity_authoritative_sameas"); c == nil || c.Pass {
			t.Errorf("%s should fail, got %+v", sameAs, c)
		}
	}
}

func TestBrandEntityConsistencyTitleSeparators(t *testing.T) {
	page := func(title string) string {
		return `<!DOCTYPE html><html><head><title>` + title + `</title>
		<meta property="og:site_name" content="Acme"></head></html>`
	}
	for _, title := range []string{"Pricing — Acme", "Pricing – Acme Widgets", "Pricing - Acme"} {
		res := Analyze("https://example.com", page(title), "", make(http.Header), 200)
		if c := findCheck(res.GEO.Checks, "geo_brand_entity_consistency"); c == nil || !c.Pass {
			t.Errorf("%q should agree, got %+v", title, c)
		}
	}
	res := Analyze("https://example.com", page("Pricing | Globex"), "", make(http.Header), 200)
	if c := findCheck(res.GEO.Checks, "geo_brand_entity_consistency"); c == nil || c.Pass {
		t.Fatalf("divergent title suffix = %+v", c)
	}
	only := Analyze("https://example.com", `<!DOCTYPE html><html><head><meta property="og:site_name" content="Acme"></head></html>`, "", make(http.Header), 200)
	if c := findCheck(only.GEO.Checks, "geo_brand_entity_consistency"); c == nil || c.Applicable == nil || *c.Applicable {
		t.Fatalf("one brand signal = %+v", c)
	}
}
