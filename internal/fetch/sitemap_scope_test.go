package fetch

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// A remote sitemap must not be able to point the auditor at local files.
func TestSitemapURLsSkipsNonWebLocations(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `<urlset><url><loc>/etc/passwd</loc></url><url><loc>file:///etc/hosts</loc></url><url><loc>%s/ok</loc></url></urlset>`, srv.URL)
	}))
	t.Cleanup(srv.Close)
	urls, err := SitemapURLs(srv.Client(), srv.URL+"/sitemap.xml", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 1 || urls[0] != srv.URL+"/ok" {
		t.Fatalf("SitemapURLs = %v, want only %s/ok", urls, srv.URL)
	}
}
