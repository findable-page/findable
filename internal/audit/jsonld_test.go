package audit

import (
	"net/http"
	"strings"
	"testing"
)

func analyzeLD(t *testing.T, ld, keyphrase string) Scan {
	t.Helper()
	page := `<html><head><title>T</title><script type="application/ld+json">` + ld + `</script></head><body><p>x</p></body></html>`
	return Analyze("https://example.com", page, keyphrase, make(http.Header), 200)
}

func TestJSONLDEvidenceOnlyInvalidBlock(t *testing.T) {
	page := `<html><head><script>ignored</script><script type="application/ld+json">{"@type":"WebPage"}</script><script type="application/ld+json">{bad</script></head></html>`
	scan := Analyze("https://example.com", page, "", make(http.Header), 200)
	check := findCheck(scan.SEO.Checks, "jsonld_valid")
	if check == nil || check.Pass || len(check.Evidence) != 1 || !strings.HasPrefix(check.Evidence[0].Value, "block 2:") {
		t.Fatalf("invalid JSON-LD evidence: %+v", check)
	}
}

func TestJSONLDTypeArrayAndTopLevelArray(t *testing.T) {
	cases := map[string]string{
		"type array":      `{"@type":["FAQPage","WebPage"]}`,
		"top-level array": `[{"@type":"WebPage"},{"@type":"FAQPage"}]`,
		"nested object":   `{"@type":"WebPage","mainEntity":{"@type":"FAQPage"}}`,
	}
	for name, ld := range cases {
		t.Run(name, func(t *testing.T) {
			faq := findCheck(analyzeLD(t, ld, "").AEO.Checks, "aeo_faq")
			if !faq.Pass {
				t.Errorf("expected FAQPage to be detected in %s", ld)
			}
		})
	}
}

func TestJSONLDPersonOnlyCountsAsAuthorWhereAllowed(t *testing.T) {
	cases := map[string]bool{
		`{"@type":"Person","name":"A"}`:                                                   true,
		`{"@graph":[{"@type":"Person","name":"A"}]}`:                                      true,
		`{"@type":"Article","author":{"@type":["Person"],"name":""}}`:                     true,
		`{"@type":"Organization","founder":{"@type":"Person","name":"F"}}`:                false,
		`{"@type":"Product","review":{"@type":"Review","reviewedBy":{"@type":"Person"}}}`: false,
		`{"@type":"Organization","employee":[{"@type":"Person"}]}`:                        false,
	}
	for ld, want := range cases {
		got := findCheck(analyzeLD(t, ld, "").GEO.Checks, "geo_author_schema").Pass
		if got != want {
			t.Errorf("geo_author_schema = %v, want %v for %s", got, want, ld)
		}
	}
}

func TestJSONLDNestedSameAsNotCounted(t *testing.T) {
	ld := `{"@type":"Organization","sameAs":["https://a.example"],"founder":{"@type":"Thing","sameAs":["https://b.example","https://c.example"]}}`
	kg := findCheck(analyzeLD(t, ld, "").GEO.Checks, "geo_knowledge_graph")
	if !strings.HasPrefix(kg.Label, "1 ") {
		t.Errorf("expected only top-level sameAs counted, got %q", kg.Label)
	}
}

func TestKeyphraseLabelKeepsQuotesVerbatim(t *testing.T) {
	title := findCheck(analyzeLD(t, `{}`, `say "hi"`).SEO.Checks, "keyphrase_in_title")
	if want := `Focus keyphrase "say "hi"" in title tag`; title.Label != want {
		t.Errorf("label = %q, want %q", title.Label, want)
	}
}

func TestKeyphraseIsTrimmedAndBlankIsSkipped(t *testing.T) {
	padded := findCheck(analyzeLD(t, `{}`, "  seo  ").SEO.Checks, "keyphrase_in_title")
	if padded == nil || padded.Label != `Focus keyphrase "seo" in title tag` {
		t.Errorf("padded keyphrase should be trimmed, got %+v", padded)
	}
	if blank := analyzeLD(t, `{}`, "   "); findCheck(blank.SEO.Checks, "keyphrase_in_title") != nil {
		t.Errorf("whitespace-only keyphrase should skip keyphrase checks")
	}
}

func TestJSONLDSpeakableDetection(t *testing.T) {
	cases := map[string]struct {
		ld   string
		want bool
	}{
		"top-level":    {`{"@type":"WebPage","speakable":{"@type":"SpeakableSpecification","cssSelector":["h1"]}}`, true},
		"inside graph": {`{"@graph":[{"@type":"WebPage","speakable":{"cssSelector":["h1"]}}]}`, true},
		"absent":       {`{"@type":"WebPage"}`, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := findCheck(analyzeLD(t, tc.ld, "").AEO.Checks, "aeo_speakable")
			if got == nil || got.Pass != tc.want {
				t.Errorf("aeo_speakable = %v, want pass=%v for %s", got, tc.want, tc.ld)
			}
		})
	}
}

func TestGeoSchemaKnowsAbout(t *testing.T) {
	cases := map[string]struct {
		ld   string
		want bool // pass or fail; skip cases handled separately
	}{
		"org with three topics":      {`{"@type":"Organization","knowsAbout":["SEO","AEO","GEO"]}`, true},
		"website with three topics":  {`{"@type":"WebSite","knowsAbout":["SEO","AEO","GEO"]}`, true},
		"topic objects with names":   {`{"@type":"Organization","knowsAbout":[{"name":"SEO"},{"name":"AEO"},"GEO"]}`, true},
		"website with one topic":     {`{"@type":"WebSite","knowsAbout":["SEO"]}`, false},
		"localbusiness with three":   {`{"@type":"LocalBusiness","knowsAbout":["SEO","AEO","GEO"]}`, true},
		"org with two topics":        {`{"@type":"Organization","knowsAbout":["SEO","AEO"]}`, false},
		"single string topic":        {`{"@type":"Organization","knowsAbout":"SEO"}`, false},
		"empty and duplicate topics": {`{"@type":"Organization","knowsAbout":["SEO","seo","AEO"]}`, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			check := findCheck(analyzeLD(t, tc.ld, "").GEO.Checks, "geo_schema_knows_about")
			if check == nil {
				t.Fatal("check missing")
			}
			if check.Pass != tc.want {
				t.Errorf("pass = %v, want %v for %s", check.Pass, tc.want, tc.ld)
			}
		})
	}
	combined := `{"@graph":[{"@type":"Organization","knowsAbout":["SEO"]},{"@type":"WebSite","knowsAbout":["AEO","GEO"]}]}`
	check := findCheck(analyzeLD(t, combined, "").GEO.Checks, "geo_schema_knows_about")
	if check == nil || !check.Pass {
		t.Errorf("combined org+website topics should pass, got %+v", check)
	}
	noSchema := findCheck(analyzeLD(t, `{"@type":"WebPage"}`, "").GEO.Checks, "geo_schema_knows_about")
	if noSchema == nil || noSchema.Applicable == nil || *noSchema.Applicable {
		t.Errorf("page without Organization or WebSite schema should skip, got %+v", noSchema)
	}
}

func TestGeoAuthorCredentialDepth(t *testing.T) {
	cases := map[string]struct {
		ld   string
		want bool
	}{
		"jobTitle":               {`{"@type":"Article","author":{"@type":"Person","name":"A","jobTitle":"Editor"}}`, true},
		"alumniOf":               {`{"@type":"Article","author":{"@type":"Person","name":"A","alumniOf":"MIT"}}`, true},
		"hasCredential":          {`{"@type":"Article","author":{"@type":"Person","name":"A","hasCredential":"PhD"}}`, true},
		"orcid in sameAs":        {`{"@type":"Article","author":{"@type":"Person","name":"A","sameAs":["https://orcid.org/0000-0002-1825-0097"]}}`, true},
		"linkedin in sameAs":     {`{"@type":"Article","author":{"name":"A","sameAs":["https://www.linkedin.com/in/a"]}}`, true},
		"untyped author map":     {`{"@type":"Article","author":{"name":"A","jobTitle":"Reporter"}}`, true},
		"multiple authors all":   {`{"@type":"Article","author":[{"@type":"Person","jobTitle":"E"},{"@type":"Person","hasCredential":"MD"}]}`, true},
		"multiple authors mixed": {`{"@type":"Article","author":[{"@type":"Person","jobTitle":"E"},{"@type":"Person","name":"B"}]}`, false},
		"name only":              {`{"@type":"Article","author":{"@type":"Person","name":"A"}}`, false},
		"social sameAs only":     {`{"@type":"Article","author":{"@type":"Person","name":"A","sameAs":["https://instagram.com/a"]}}`, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			check := findCheck(analyzeLD(t, tc.ld, "").GEO.Checks, "geo_author_credential_depth")
			if check == nil {
				t.Fatal("check missing")
			}
			if check.Pass != tc.want {
				t.Errorf("pass = %v, want %v for %s", check.Pass, tc.want, tc.ld)
			}
		})
	}
	graph := `{"@graph":[{"@type":"Article","author":{"@id":"#a"}},{"@id":"#a","@type":"Person","name":"A","jobTitle":"Editor"}]}`
	check := findCheck(analyzeLD(t, graph, "").GEO.Checks, "geo_author_credential_depth")
	if check == nil || !check.Pass {
		t.Errorf("author resolved through @id reference should pass, got %+v", check)
	}
	noAuthor := findCheck(analyzeLD(t, `{"@type":"WebPage"}`, "").GEO.Checks, "geo_author_credential_depth")
	if noAuthor == nil || noAuthor.Applicable == nil || *noAuthor.Applicable {
		t.Errorf("page without author schema should skip, got %+v", noAuthor)
	}
	orgAuthor := findCheck(analyzeLD(t, `{"@type":"Article","author":{"@type":"Organization","name":"Desk"}}`, "").GEO.Checks, "geo_author_credential_depth")
	if orgAuthor == nil || orgAuthor.Applicable == nil || *orgAuthor.Applicable {
		t.Errorf("organization author should skip the Person credential check, got %+v", orgAuthor)
	}
}
