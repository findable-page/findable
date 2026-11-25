package mcpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// A crawled page must not be able to make the MCP host read its own files:
// crawl_site follows only http(s) links on the start host.
func TestMCPCrawlSiteFollowsOnlyWebLinksOnTheStartHost(t *testing.T) {
	secret := filepath.Join(t.TempDir(), "secret.html")
	if err := os.WriteFile(secret, []byte("<html><head><title>secret</title></head></html>"), 0o600); err != nil {
		t.Fatal(err)
	}
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body><a href="file://` + secret + `">x</a><a href="/next">n</a></body></html>`))
	}))
	t.Cleanup(site.Close)

	text, isError := executeTool("crawl_site", map[string]any{"url": site.URL + "/", "limit": float64(10)}, site.Client())
	if isError {
		t.Fatalf("crawl_site error: %s", text)
	}
	var got struct {
		Pages []pageSummary `json:"pages"`
	}
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatal(err)
	}
	for _, p := range got.Pages {
		if p.URL != site.URL+"/" && p.URL != site.URL+"/next" {
			t.Errorf("crawl_site audited %q; want only http pages on %s", p.URL, site.URL)
		}
	}
}
