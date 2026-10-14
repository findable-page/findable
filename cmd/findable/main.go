// findable scans a site locally and saves the report CSV.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/nadimtuhin/findable/internal/audit"
	"github.com/nadimtuhin/findable/internal/fetch"
	"github.com/nadimtuhin/findable/internal/mcpserver"
	"github.com/nadimtuhin/findable/internal/report"
	"github.com/nadimtuhin/findable/internal/server"
)

const (
	usageLine     = "findable: is your site findable by Google and by AI answer engines?\nusage: findable [-o dir] [-format csv|json|sarif] [-fail-under N] [-q] [-k keyphrase] [-site] [-crawl] [-sitemap url] [-limit N] [-serve :port] [-mcp] <url or file>"
	clientTimeout = 60 * time.Second
	reportDirMode = 0o755
	reportMode    = 0o644
	exitUsage     = 2
)

type options struct {
	outDir    string
	quiet     bool
	keyphrase string
	crawl     bool
	sitemap   string
	site      bool
	limit     int
	serve     string
	mcp       bool
	format    report.Format
	failUnder int
}

func parseFlags() options {
	var o options
	flag.StringVar(&o.outDir, "o", ".", "directory for the CSV")
	flag.BoolVar(&o.quiet, "q", false, "do not list failing checks")
	flag.StringVar(&o.keyphrase, "k", "", "target keyphrase for audit")
	flag.BoolVar(&o.crawl, "crawl", false, "crawl internal links up to limit pages")
	flag.BoolVar(&o.site, "site", false, "also check robots.txt (search and AI crawlers) and /llms.txt")
	flag.StringVar(&o.sitemap, "sitemap", "", "audit the pages listed in this sitemap.xml, up to limit pages")
	flag.IntVar(&o.limit, "limit", 20, "max pages to crawl when -crawl or -sitemap is enabled")
	flag.StringVar(&o.serve, "serve", "", "run HTTP API server on [host]:port (e.g. :8080)")
	flag.BoolVar(&o.mcp, "mcp", false, "run Model Context Protocol (MCP) server over stdio")
	flag.IntVar(&o.failUnder, "fail-under", 0, "exit 1 when the overall score of any page is below N")
	format := flag.String("format", string(report.FormatCSV), "report format: csv, json or sarif")
	flag.Usage = func() { fmt.Fprintln(os.Stderr, usageLine) }
	flag.Parse()
	f, err := report.ParseFormat(*format)
	if err != nil {
		fmt.Fprintln(os.Stderr, "findable:", err)
		flag.Usage()
		os.Exit(exitUsage)
	}
	o.format = f
	return o
}

func main() {
	opts := parseFlags()
	client := &http.Client{Timeout: clientTimeout}

	switch {
	case opts.mcp:
		mcpserver.Run(os.Stdin, os.Stdout, client)
	case opts.serve != "":
		serveHTTP(opts.serve)
	default:
		if opts.sitemap != "" {
			if flag.NArg() != 0 {
				flag.Usage()
				os.Exit(exitUsage)
			}
			runSitemap(client, opts)
			return
		}
		if flag.NArg() != 1 {
			flag.Usage()
			os.Exit(exitUsage)
		}
		runAudit(client, flag.Arg(0), opts)
	}
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
		printFailures("", s)
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
