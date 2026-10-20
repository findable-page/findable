package fetch

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/findable-page/findable/internal/audit"
)

func statusOf(c *audit.Category) map[string]string {
	got := map[string]string{}
	for _, k := range c.Checks {
		got[k.ID] = audit.Status(k)
	}
	return got
}

func TestSiteChecksReadsRobotsAndLLMSFromTheOrigin(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "User-agent: GPTBot\nDisallow: /blog\n")
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	got := statusOf(SiteChecks(srv.Client(), srv.URL+"/blog/post?x=1"))
	want := map[string]string{"robots-txt": audit.StatusPass, "robots-gptbot": audit.StatusFail, "robots-claudebot": audit.StatusPass, "llms-txt": audit.StatusFail}
	for id, status := range want {
		if got[id] != status {
			t.Errorf("%s = %q, want %q (all: %v)", id, got[id], status, got)
		}
	}
}

func TestSiteChecksTreatsUnreachableFilesAsMissing(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()
	got := statusOf(SiteChecks(srv.Client(), srv.URL+"/"))
	if got["robots-txt"] != audit.StatusFail || got["llms-txt"] != audit.StatusFail || got["robots-page"] != audit.StatusPass {
		t.Errorf("unreachable origin = %v", got)
	}
}

func TestCrawlAddsSiteChecksOnlyWhenAsked(t *testing.T) {
	srv := sitemapSite(t)
	plain, err := Crawl(srv.Client(), srv.URL+"/a", CrawlOptions{Limit: 1})
	if err != nil || plain[0].Scan.Site != nil {
		t.Fatalf("default crawl must not add site checks: %v %+v", err, plain)
	}
	withSite, err := Crawl(srv.Client(), srv.URL+"/a", CrawlOptions{Limit: 1, Site: true})
	if err != nil || withSite[0].Scan.Site == nil {
		t.Fatalf("crawl with Site must add site checks: %v", err)
	}
}
