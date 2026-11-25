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
// parser and `findable audit` stay in sync.
func registerAuditFlags(fs *flag.FlagSet, o *options) (*string, string) {
	var minSeverity string
	fs.StringVar(&o.outDir, "o", ".", "directory to write the report into")
	fs.BoolVar(&o.quiet, "q", false, "hide the prioritized fix list")
	fs.StringVar(&o.keyphrase, "k", "", "focus keyphrase; adds checks for the title, description, H1 and body")
	fs.BoolVar(&o.crawl, "crawl", false, "follow internal links, up to -limit pages")
	fs.BoolVar(&o.site, "site", false, "also check robots.txt, AI crawlers and /llms.txt (unscored)")
	fs.StringVar(&o.sitemap, "sitemap", "", "audit the pages listed in this sitemap URL, up to -limit")
	fs.IntVar(&o.limit, "limit", 20, "maximum pages for -crawl or -sitemap")
	fs.BoolVar(&o.allowPrivate, "allow-private", false, "allow localhost and private-network targets (disables the SSRF guard)")
	fs.IntVar(&o.failUnder, "fail-under", 0, "exit 1 when any page's overall score is below N (0 disables the gate)")
	fs.StringVar(&o.scoreModel, "score-model", audit.ScoreModelWeighted, "weighted (severity x evidence tier) or flat (passed/total)")
	fs.IntVar(&o.top, "top", 10, "maximum findings to print in the fix list")
	fs.StringVar(&minSeverity, "min-severity", string(audit.SeverityLow), "lowest severity to print: low, medium, high or critical")
	fs.StringVar(&o.configPath, "config", "", fmt.Sprintf("JSON config file (default %s when that file exists)", config.DefaultPath))
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
	fs := flag.NewFlagSet("findable", flag.ContinueOnError)
	format, minSeverity := registerAuditFlags(fs, &o)
	fs.StringVar(&o.serve, "serve", "", "legacy: run the HTTP API on host:port (prefer \"findable serve\")")
	fs.BoolVar(&o.mcp, "mcp", false, "legacy: run the MCP server on stdio (prefer \"findable mcp\")")
	commandUsage(fs, usageLine)
	parseCommand(fs, args)
	finishOptions(&o, *format, minSeverity)
	client := newClient(o.allowPrivate)

	switch {
	case o.mcp:
		runMCP(client)
	case o.serve != "":
		serveHTTP(o.serve)
	default:
		runAuditTarget(client, fs.Args(), o, usageLine)
	}
}

// runAuditTarget handles the default (non -mcp/-serve) path: sitemap mode
// with no positional args, or exactly one target.
func runAuditTarget(client *http.Client, args []string, o options, usage string) {
	if o.sitemap != "" {
		if len(args) != 0 {
			failUsage("-sitemap does not take a positional argument; pass the sitemap URL as the flag value", usage)
		}
		runSitemap(client, o)
		return
	}
	if len(args) != 1 {
		problem := "need one URL or HTML file"
		if len(args) > 1 {
			problem = fmt.Sprintf("got %d arguments, want one URL or HTML file", len(args))
		}
		failUsage(problem, usage)
	}
	runAudit(client, args[0], o)
}
