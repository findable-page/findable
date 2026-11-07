package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/findable-page/findable/internal/audit"
	"github.com/findable-page/findable/internal/fetch"
	"github.com/findable-page/findable/internal/mcpserver"
	"github.com/findable-page/findable/internal/report"
	"github.com/findable-page/findable/internal/server"
)

const (
	clientTimeout = 60 * time.Second
	reportDirMode = 0o755
	reportMode    = 0o644
)

func newClient(allowPrivate bool) *http.Client {
	if allowPrivate {
		return &http.Client{Timeout: clientTimeout}
	}
	client := fetch.NewGuardedClient()
	client.Timeout = clientTimeout
	return client
}

func runMCP(client *http.Client) {
	mcpserver.Run(os.Stdin, os.Stdout, client)
}

func serveHTTP(addr string) {
	fmt.Printf("findable API server listening on %s\n", addr)
	if err := server.NewServer(addr).ListenAndServe(); err != nil {
		fail("server error: %v", err)
	}
}

func runAudit(client *http.Client, target string, opts options) {
	hostName := reportHostName(target)
	if err := os.MkdirAll(opts.outDir, reportDirMode); err != nil { //nolint:gosec // reports are shareable output, not secrets
		fail("create %s: %v", opts.outDir, err)
	}
	if opts.site && !fetch.IsHTTPURL(target) {
		fail("-site requires an http(s) URL, not a local file")
	}
	if opts.crawl {
		if !fetch.IsHTTPURL(target) {
			fail("-crawl requires an http(s) URL, not a local file")
		}
		runCrawl(client, target, hostName, opts)
		return
	}

	s, err := fetch.Scan(client, target, opts.keyphrase)
	if err != nil {
		fail("%v", err)
	}
	if opts.site {
		s.Site = fetch.SiteChecks(client, target)
	}
	path := reportPath(opts.outDir, "findable-report", hostName, opts.format)
	writeReport(path, report.PageReport(opts.format, hostName, audit.Page{URL: target, Scan: s}))

	fmt.Printf("overall %d  SEO %s  AEO %s  GEO %s\n", audit.Overall(s), audit.ScoreText(s.SEO), audit.ScoreText(s.AEO), audit.ScoreText(s.GEO))
	if !opts.quiet {
		printFixList([]audit.Page{{URL: target, Scan: s}}, opts.cfg, opts.top, opts.minSeverity)
	}
	fmt.Println("saved", path)
	exitIfBelow([]audit.Page{{URL: target, Scan: s}}, opts.failUnder)
}

func reportHostName(target string) string {
	host, ok := report.HostName(target)
	if !ok {
		fail("need a valid URL or local HTML file path, got %q", target)
	}
	return host
}

func runCrawl(client *http.Client, start, hostName string, opts options) {
	pages, err := fetch.Crawl(client, start, crawlOptions(opts))
	if err != nil {
		fail("%v", err)
	}
	saveSiteReport(pages, hostName, opts)
}

func runSitemap(client *http.Client, opts options) {
	if !fetch.IsHTTPURL(opts.sitemap) {
		fail("-sitemap requires an http(s) URL")
	}
	hostName := reportHostName(opts.sitemap)
	if err := os.MkdirAll(opts.outDir, reportDirMode); err != nil { //nolint:gosec // reports are shareable output, not secrets
		fail("create %s: %v", opts.outDir, err)
	}
	urls, err := fetch.SitemapURLs(client, opts.sitemap, opts.limit)
	if err != nil {
		fail("%v", err)
	}
	if len(urls) == 0 {
		fail("sitemap %s lists no pages", opts.sitemap)
	}
	pages, err := fetch.ScanAll(client, urls, crawlOptions(opts))
	if err != nil {
		fail("%v", err)
	}
	saveSiteReport(pages, hostName, opts)
}

func crawlOptions(opts options) fetch.CrawlOptions {
	return fetch.CrawlOptions{
		Keyphrase: opts.keyphrase,
		Limit:     opts.limit,
		Site:      opts.site,
		OnPage: func(done, limit int, p audit.Page) {
			printCrawlProgress(done, limit, p, opts.quiet)
		},
	}
}

func saveSiteReport(pages []audit.Page, hostName string, opts options) {
	path := reportPath(opts.outDir, "findable-site-report", hostName, opts.format)
	writeReport(path, report.SiteReport(opts.format, hostName, pages))
	if !opts.quiet {
		printFixList(pages, opts.cfg, opts.top, opts.minSeverity)
	}
	fmt.Printf("saved %s (%d pages)\n", path, len(pages))
	exitIfBelow(pages, opts.failUnder)
}

func exitIfBelow(pages []audit.Page, minimum int) {
	below := audit.PagesBelow(pages, minimum)
	for _, p := range below {
		fmt.Fprintf(os.Stderr, "findable: %s overall %d is below -fail-under %d\n", p.URL, audit.Overall(p.Scan), minimum)
	}
	if len(below) > 0 {
		os.Exit(1)
	}
}

func printCrawlProgress(done, limit int, p audit.Page, quiet bool) {
	s := p.Scan
	fmt.Printf("[%d/%d] %s  overall %d  SEO %s  AEO %s  GEO %s  failed %d\n",
		done, limit, p.URL, audit.Overall(s), audit.ScoreText(s.SEO), audit.ScoreText(s.AEO), audit.ScoreText(s.GEO), audit.FailedCount(s))
	if !quiet {
		printFailures("  ", s)
	}
}

func printFailures(indent string, s audit.Scan) {
	for _, c := range s.Reported() {
		for _, k := range c.Cat.Checks {
			if audit.Status(k) == audit.StatusFail {
				fmt.Printf("%sFAIL %s %s: %s\n", indent, c.Name, k.ID, k.Label)
			}
		}
	}
}

func reportPath(outDir, prefix, hostName string, f report.Format) string {
	return filepath.Join(outDir, fmt.Sprintf("%s-%s-%s%s", prefix, hostName, time.Now().Format(report.DateLayout), f.Ext()))
}

func writeReport(path, content string) {
	if err := os.WriteFile(path, []byte(content), reportMode); err != nil { //nolint:gosec // reports are shareable output, not secrets
		fail("write %s: %v", path, err)
	}
}

func fail(f string, a ...any) {
	fmt.Fprintf(os.Stderr, "findable: "+f+"\n", a...)
	os.Exit(1)
}

const exitRuntime = 3

// failRuntime reports a runtime/network error with exit code 3 so CI can
// tell "the site is bad" (1) from "the run broke" (3).
func failRuntime(f string, a ...any) {
	fmt.Fprintf(os.Stderr, "findable: "+f+"\n", a...)
	os.Exit(exitRuntime)
}
