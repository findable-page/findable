package fetch

import (
	"io"
	"net/http"
	"net/http/httptest"
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
