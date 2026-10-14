package fetch

import (
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestScanResolvesLinksAgainstTheFinalURLAfterRedirects(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/a", http.RedirectHandler("/dir/page", http.StatusMovedPermanently))
	mux.HandleFunc("/dir/page", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body><a href="next">next</a></body></html>`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	s, err := Scan(srv.Client(), srv.URL+"/a", "")
	if err != nil {
		t.Fatal(err)
	}
	if want := srv.URL + "/dir/next"; !slices.Contains(s.InternalLinks, want) {
		t.Errorf("InternalLinks = %v, want %s", s.InternalLinks, want)
	}
}

type hostRedirect struct{}

func (hostRedirect) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := &http.Response{Request: req, Header: make(http.Header), StatusCode: http.StatusOK, Body: http.NoBody}
	if req.URL.Host == "example.com" {
		resp.StatusCode = http.StatusMovedPermanently
		resp.Header.Set("Location", "http://www.example.com"+req.URL.Path)
		return resp, nil
	}
	page := `<html><body><a href="sms:+1">s</a><a href="/b">b</a></body></html>`
	resp.Body = io.NopCloser(strings.NewReader(page))
	return resp, nil
}

func TestCrawlFollowsInternalLinksWhenTheStartRedirectsToAnotherHost(t *testing.T) {
	pages, err := Crawl(&http.Client{Transport: hostRedirect{}}, "http://example.com/", CrawlOptions{Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 {
		t.Errorf("crawled %d pages, want 2: %v", len(pages), pages)
	}
}

func TestScanAddsKeyphraseChecksOnlyWhenKeyphraseGiven(t *testing.T) {
	page := filepath.Join(t.TempDir(), "page.html")
	html := `<html><head><title>Best Espresso Guide</title><meta name="description" content="All about espresso"></head>` +
		`<body><h1>Brewing</h1><p>Espresso needs fine grounds.</p></body></html>`
	if err := os.WriteFile(page, []byte(html), 0o600); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"keyphrase_in_title":   true,
		"keyphrase_in_desc":    true,
		"keyphrase_in_h1":      false,
		"keyphrase_in_content": true,
	}

	with, err := Scan(http.DefaultClient, page, "ESPRESSO")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, c := range with.SEO.Checks {
		if strings.HasPrefix(c.ID, "keyphrase_") {
			got[c.ID] = c.Pass
		}
	}
	if !maps.Equal(got, want) {
		t.Errorf("keyphrase checks = %v, want %v", got, want)
	}

	without, err := Scan(http.DefaultClient, page, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range without.SEO.Checks {
		if strings.HasPrefix(c.ID, "keyphrase_") {
			t.Errorf("unexpected %s without keyphrase", c.ID)
		}
	}
}
