package fetch

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nadimtuhin/findable/internal/audit"
)

const pageHTML = `<!doctype html><html lang="en"><head><title>Page</title></head><body><h1>Hi</h1></body></html>`

func sitemapSite(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	var srv *httptest.Server
	mux.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `<?xml version="1.0"?><sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
<sitemap><loc>%[1]s/posts.xml</loc></sitemap><sitemap><loc>%[1]s/pages.xml</loc></sitemap></sitemapindex>`, srv.URL)
	})
	mux.HandleFunc("/posts.xml", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
<url><loc> %[1]s/a </loc></url><url><loc>%[1]s/b</loc></url></urlset>`, srv.URL)
	})
	mux.HandleFunc("/pages.xml", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `<urlset><url><loc>%[1]s/c</loc></url><url><loc>%[1]s/a</loc></url></urlset>`, srv.URL)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) { _, _ = fmt.Fprint(w, pageHTML) })
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestSitemapURLsFollowsIndexAndDropsDuplicates(t *testing.T) {
	srv := sitemapSite(t)
	got, err := SitemapURLs(srv.Client(), srv.URL+"/sitemap.xml", 10)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{srv.URL + "/a", srv.URL + "/b", srv.URL + "/c"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("SitemapURLs = %v, want %v", got, want)
	}
}

func TestSitemapURLsRespectsLimit(t *testing.T) {
	srv := sitemapSite(t)
	got, err := SitemapURLs(srv.Client(), srv.URL+"/sitemap.xml", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("SitemapURLs limit 2 = %v", got)
	}
}

func TestSitemapURLsFailsOnMissingSitemap(t *testing.T) {
	srv := sitemapSite(t)
	srv.Config.Handler = http.NotFoundHandler()
	if _, err := SitemapURLs(srv.Client(), srv.URL+"/sitemap.xml", 5); err == nil {
		t.Error("want an error for a 404 sitemap")
	}
}

func TestScanAllAuditsEachURL(t *testing.T) {
	srv := sitemapSite(t)
	var seen []string
	pages, err := ScanAll(srv.Client(), []string{srv.URL + "/a", srv.URL + "/b"}, CrawlOptions{
		OnPage: func(done, total int, p audit.Page) { seen = append(seen, fmt.Sprintf("%d/%d %s", done, total, p.URL)) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 || pages[1].URL != srv.URL+"/b" || pages[0].Scan.SEO.Score == nil {
		t.Errorf("ScanAll pages = %+v", pages)
	}
	if len(seen) != 2 || seen[1] != "2/2 "+srv.URL+"/b" {
		t.Errorf("progress = %v", seen)
	}
}

func TestSitemapURLsFetchesEachSitemapOnce(t *testing.T) {
	hits := 0
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		_, _ = fmt.Fprintf(w, `<sitemapindex><sitemap><loc>%[1]s/sitemap.xml</loc></sitemap><sitemap><loc>%[1]s/sitemap.xml</loc></sitemap></sitemapindex>`, srv.URL)
	}))
	t.Cleanup(srv.Close)
	if _, err := SitemapURLs(srv.Client(), srv.URL+"/sitemap.xml", 5); err != nil {
		t.Fatal(err)
	}
	if hits != 1 {
		t.Errorf("self-referencing sitemap fetched %d times, want 1", hits)
	}
}

func TestScanAllNamesTheCountWhenEveryPageFails(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()
	_, err := ScanAll(srv.Client(), []string{srv.URL + "/a", srv.URL + "/b"}, CrawlOptions{})
	if err == nil || !strings.Contains(err.Error(), "all 2 pages failed") {
		t.Errorf("ScanAll error = %v", err)
	}
}
