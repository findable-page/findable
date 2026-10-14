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
