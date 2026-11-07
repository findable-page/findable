package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/findable-page/findable/internal/audit"
	"github.com/findable-page/findable/internal/config"
	"github.com/findable-page/findable/internal/report"
)

const exitUsage = 2

type options struct {
	outDir       string
	quiet        bool
	keyphrase    string
	crawl        bool
	sitemap      string
	site         bool
	limit        int
	serve        string
	mcp          bool
	allowPrivate bool
	format       report.Format
	failUnder    int
	scoreModel   string
	top          int
	minSeverity  audit.Severity
	configPath   string
	cfg          config.Config
}

// registerAuditFlags registers the shared audit flag set on fs, so the legacy
// parser (flag.CommandLine) and `findable audit` stay in sync.
func registerAuditFlags(fs *flag.FlagSet, o *options) (*string, string) {
	var minSeverity string
	fs.StringVar(&o.outDir, "o", ".", "directory for the report")
	fs.BoolVar(&o.quiet, "q", false, "do not list failing checks")
	fs.StringVar(&o.keyphrase, "k", "", "target keyphrase for audit")
	fs.BoolVar(&o.crawl, "crawl", false, "crawl internal links up to limit pages")
	fs.BoolVar(&o.site, "site", false, "also check robots.txt (search and AI crawlers) and /llms.txt")
	fs.StringVar(&o.sitemap, "sitemap", "", "audit the pages listed in this sitemap.xml, up to limit pages")
	fs.IntVar(&o.limit, "limit", 20, "max pages to crawl when -crawl or -sitemap is enabled")
	fs.StringVar(&o.serve, "serve", "", "run HTTP API server on [host]:port (e.g. 127.0.0.1:8080)")
	fs.BoolVar(&o.mcp, "mcp", false, "run Model Context Protocol (MCP) server over stdio")
	fs.BoolVar(&o.allowPrivate, "allow-private", false, "allow scanning localhost and private-network targets")
	fs.IntVar(&o.failUnder, "fail-under", 0, "exit 1 when the overall score of any page is below N")
	fs.StringVar(&o.scoreModel, "score-model", audit.ScoreModelWeighted, "scoring model: weighted (severity x evidence tier) or flat (legacy)")
	fs.IntVar(&o.top, "top", 10, "how many findings to show in the fix list")
	fs.StringVar(&minSeverity, "min-severity", string(audit.SeverityLow), "lowest finding severity to show: low, medium, high or critical")
	fs.StringVar(&o.configPath, "config", "", fmt.Sprintf("config file path (default: %s when present)", config.DefaultPath))
	return fs.String("format", string(report.FormatCSV), "report format: csv, json or sarif"), minSeverity
}

func finishOptions(o *options, format, minSeverity string) {
	f, err := report.ParseFormat(format)
	if err != nil {
		fmt.Fprintln(os.Stderr, "findable:", err)
		os.Exit(exitUsage)
	}
	o.format = f
	sev, err := report.ParseSeverity(minSeverity)
	if err != nil {
		fmt.Fprintln(os.Stderr, "findable:", err)
		os.Exit(exitUsage)
	}
	o.minSeverity = sev
	switch o.scoreModel {
	case audit.ScoreModelWeighted, audit.ScoreModelFlat:
		audit.ScoreModel = o.scoreModel
	default:
		fmt.Fprintf(os.Stderr, "findable: unknown score model %q (want weighted or flat)\n", o.scoreModel)
		os.Exit(exitUsage)
	}
	if o.configPath != "" {
		cfg, err := config.Load(o.configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "findable:", err)
			os.Exit(exitUsage)
		}
		o.cfg = cfg
	} else {
		cfg, err := config.LoadDefault()
		if err != nil {
			fmt.Fprintln(os.Stderr, "findable:", err)
			os.Exit(exitUsage)
		}
		o.cfg = cfg
	}
}

func legacyMain(args []string) {
	var o options
	fs := flag.CommandLine
	format, minSeverity := registerAuditFlags(fs, &o)
	fs.Usage = func() { fmt.Fprintln(os.Stderr, usageLine) }
	_ = fs.Parse(args)
	finishOptions(&o, *format, minSeverity)
	client := newClient(o.allowPrivate)

	switch {
	case o.mcp:
		runMCP(client)
	case o.serve != "":
		serveHTTP(o.serve)
	default:
		runAuditTarget(client, fs.Args(), o)
	}
}

// runAuditTarget handles the default (non -mcp/-serve) legacy path: sitemap
// mode with no positional args, or exactly one target.
func runAuditTarget(client *http.Client, args []string, o options) {
	if o.sitemap != "" {
		if len(args) != 0 {
			fmt.Fprintln(os.Stderr, usageLine)
			os.Exit(exitUsage)
		}
		runSitemap(client, o)
		return
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, usageLine)
		os.Exit(exitUsage)
	}
	runAudit(client, args[0], o)
}
