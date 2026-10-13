// seoscore scans a site locally or via seoscore.tools and saves the report CSV.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	usageLine     = "usage: seoscore [-o dir] [-q] [-k keyphrase] [-crawl] [-limit N] [-remote] [-serve :port] [-mcp] <url or file>"
	clientTimeout = 60 * time.Second
	reportDirMode = 0o755
	reportMode    = 0o644
	dateLayout    = "2006-01-02"
	exitUsage     = 2
)

type options struct {
	outDir    string
	quiet     bool
	keyphrase string
	crawl     bool
	limit     int
	delay     time.Duration
	remote    bool
	serve     string
	mcp       bool
}

func parseFlags() options {
	var o options
	flag.StringVar(&o.outDir, "o", ".", "directory for the CSV")
	flag.BoolVar(&o.quiet, "q", false, "do not list failing checks")
	flag.StringVar(&o.keyphrase, "k", "", "target keyphrase for audit")
	flag.BoolVar(&o.crawl, "crawl", false, "crawl internal links up to limit pages")
	flag.IntVar(&o.limit, "limit", 20, "max pages to crawl when -crawl is enabled")
	flag.DurationVar(&o.delay, "delay", 3*time.Second, "delay between crawl requests")
	flag.BoolVar(&o.remote, "remote", false, "use remote seoscore.tools API instead of local engine")
	flag.StringVar(&o.serve, "serve", "", "run HTTP API server on [host]:port (e.g. :8080)")
	flag.BoolVar(&o.mcp, "mcp", false, "run Model Context Protocol (MCP) server over stdio")
	flag.Usage = func() { fmt.Fprintln(os.Stderr, usageLine) }
	flag.Parse()
	return o
}

func main() {
	opts := parseFlags()
	client := &http.Client{Timeout: clientTimeout}

	switch {
	case opts.mcp:
		runMCP(os.Stdin, os.Stdout, client)
	case opts.serve != "":
		serveHTTP(opts.serve, client, opts.remote)
	default:
		if flag.NArg() != 1 {
			flag.Usage()
			os.Exit(exitUsage)
		}
		runAudit(client, flag.Arg(0), opts)
	}
}

func serveHTTP(addr string, client *http.Client, remote bool) {
	fmt.Printf("seoscore API server listening on %s (remote=%v)\n", addr, remote)
	if err := newServer(addr, client, remote).ListenAndServe(); err != nil {
		fail("server error: %v", err)
	}
}

func runAudit(client *http.Client, target string, opts options) {
	hostName := reportHostName(target)
	if err := os.MkdirAll(opts.outDir, reportDirMode); err != nil { //nolint:gosec // reports are shareable output, not secrets
		fail("create %s: %v", opts.outDir, err)
	}
	if opts.crawl {
		if !isHTTPURL(target) {
			fail("-crawl requires an http(s) URL, not a local file")
		}
		runCrawl(client, target, hostName, opts)
		return
	}

	s, err := scanTarget(client, target, opts.keyphrase, opts.remote)
	if err != nil {
		fail("%v", err)
	}
	path := reportPath(opts.outDir, "seoscore-report", hostName)
	writeReport(path, csv(hostName, s))

	fmt.Printf("overall %d  SEO %s  AEO %s  GEO %s\n", overall(s), score(s.SEO), score(s.AEO), score(s.GEO))
	if !opts.quiet {
		printFailures("", s)
	}
	fmt.Println("saved", path)
}

func reportHostName(target string) string {
	if !isHTTPURL(target) {
		return filepath.Base(target)
	}
	u, err := url.Parse(target)
	if err != nil || u.Host == "" {
		fail("need a valid URL or local HTML file path, got %q", target)
	}
	return u.Host
}

func runCrawl(client *http.Client, start, hostName string, opts options) {
	pages, err := crawl(client, start, crawlOptions{
		keyphrase: opts.keyphrase,
		limit:     opts.limit,
		delay:     opts.delay,
		remote:    opts.remote,
		onPage: func(done, limit int, p crawledPage) {
			printCrawlProgress(done, limit, p, opts.quiet)
		},
	})
	if err != nil {
		fail("%v", err)
	}
	path := reportPath(opts.outDir, "seoscore-site-report", hostName)
	writeReport(path, crawlCSV(hostName, pages))
	fmt.Printf("saved %s (%d pages)\n", path, len(pages))
}

func printCrawlProgress(done, limit int, p crawledPage, quiet bool) {
	s := p.scan
	fmt.Printf("[%d/%d] %s  overall %d  SEO %s  AEO %s  GEO %s  failed %d\n",
		done, limit, p.url, overall(s), score(s.SEO), score(s.AEO), score(s.GEO), failedCount(s))
	if !quiet {
		printFailures("  ", s)
	}
}

func printFailures(indent string, s scan) {
	for _, c := range s.scored() {
		for _, k := range c.cat.Checks {
			if status(k) == statusFail {
				fmt.Printf("%sFAIL %s %s: %s\n", indent, c.name, k.ID, k.Label)
			}
		}
	}
}

func reportPath(outDir, prefix, hostName string) string {
	return filepath.Join(outDir, fmt.Sprintf("%s-%s-%s.csv", prefix, hostName, time.Now().Format(dateLayout)))
}

func writeReport(path, content string) {
	if err := os.WriteFile(path, []byte(content), reportMode); err != nil { //nolint:gosec // reports are shareable output, not secrets
		fail("write %s: %v", path, err)
	}
}

func fail(f string, a ...any) {
	fmt.Fprintf(os.Stderr, "seoscore: "+f+"\n", a...)
	os.Exit(1)
}
