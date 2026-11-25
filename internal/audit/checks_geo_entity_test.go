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
