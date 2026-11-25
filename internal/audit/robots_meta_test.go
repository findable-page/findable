package audit

import (
	"net/http"
	"testing"
)

// Google combines every robots meta tag on a page, so a noindex or nosnippet
// in any one of them applies even when a later tag says something else.
func TestRobotsMetaChecksSeeEveryRobotsTag(t *testing.T) {
	head := `<title>t</title>` +
		`<meta name="robots" content="noindex, nosnippet">` +
		`<meta name="robots" content="max-image-preview:large">`
	page := "<!DOCTYPE html><html lang='en'><head>" + head + "</head><body><p>x</p></body></html>"
	scan := Analyze("https://example.com/page", page, "", make(http.Header), 200)
	all := append(append(append([]Check{}, scan.SEO.Checks...), scan.AEO.Checks...), scan.GEO.Checks...)
	for _, id := range []string{"robots_meta", "aeo_ai_robots", "geo_snippet_control"} {
		c := findCheck(all, id)
		if c == nil {
			t.Fatalf("%s: check missing", id)
		}
		if c.Pass {
			t.Errorf("%s passed although an earlier robots meta tag says noindex, nosnippet", id)
		}
	}
}

// A local file has no web address, so there is nothing to compare the
// canonical with; failing canonical_self there blames a correct canonical.
func TestCanonicalSelfIsSkippedForLocalFiles(t *testing.T) {
	page := "<!DOCTYPE html><html lang='en'><head><title>t</title>" +
		"<link rel='canonical' href='https://example.com/page'></head><body></body></html>"
	scan := Analyze("file://page.html", page, "", make(http.Header), 200)
	c := findCheck(scan.SEO.Checks, "canonical_self")
	if c == nil {
		t.Fatal("canonical_self missing")
	}
	if c.Applicable == nil || *c.Applicable {
		t.Errorf("canonical_self for a local file = pass:%v label:%q, want skipped", c.Pass, c.Label)
	}
	if abs := findCheck(scan.SEO.Checks, "canonical_absolute"); abs == nil || !abs.Pass {
		t.Errorf("canonical_absolute should still be evaluated for a local file, got %+v", abs)
	}
}

func TestCanonicalSelfKeepsSchemeAndIgnoresDefaultPort(t *testing.T) {
	same := `<!DOCTYPE html><html><head><title>t</title><link rel="canonical" href="https://Example.com:443/docs"></head><body></body></html>`
	scan := Analyze("https://example.com/docs", same, "", make(http.Header), http.StatusOK)
	if c := findCheck(scan.SEO.Checks, "canonical_self"); c == nil || !c.Pass {
		t.Fatalf("default https port and host case should match, got %+v", c)
	}
	other := `<!DOCTYPE html><html><head><title>t</title><link rel="canonical" href="http://example.com/docs"></head><body></body></html>`
	scan = Analyze("https://example.com/docs", other, "", make(http.Header), http.StatusOK)
	if c := findCheck(scan.SEO.Checks, "canonical_self"); c == nil || c.Pass {
		t.Fatalf("http canonical on an https page should fail, got %+v", c)
	}
}
