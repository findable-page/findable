package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

const usageLine = `findable: is your site findable by Google and by AI answer engines?
usage: findable [command] [flags] <url or file>

commands:
  audit      audit one page, a crawl, or a sitemap (default)
  bots       probe how the server treats AI crawlers
  diff       compare a JSON report with a baseline
  fanout     report which query fan-out sub-intents a page covers
  hydrate    compare raw HTML with a rendered DOM
  mcp        run the MCP server on standard input
  serve      run the local HTTP API and web UI
  simulate   ask an answer engine and report citations

findable <command> -h lists that command's flags.
Exit codes: 0 ok, 1 a score gate failed, 2 invalid usage, 3 a runtime or network error.`

const (
	auditUsage    = "usage: findable audit [flags] <url or file>\n       findable audit -sitemap <sitemap url> [flags]"
	botsUsage     = "usage: findable bots [flags] <url>"
	diffUsage     = "usage: findable diff [flags] <baseline.json> <current.json>"
	fanoutUsage   = "usage: findable fanout -query TOPIC [flags] <url or file>"
	hydrateUsage  = "usage: findable hydrate (-rendered <file> | -render-cmd <program>) [flags] <url or file>"
	mcpUsage      = "usage: findable mcp [-allow-private]"
	serveUsage    = "usage: findable serve <host:port>"
	simulateUsage = "usage: findable simulate -query TEXT -brand NAME -domain HOST [-engine URL] [flags]"
)

func commandUsage(fs *flag.FlagSet, summary string) {
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, summary)
		fmt.Fprintln(os.Stderr)
		fs.PrintDefaults()
	}
}

// parseCommand exits 0 on -h and 2 on any other parse error. The flag
// package has already printed the reason and the command's usage.
func parseCommand(fs *flag.FlagSet, args []string) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		os.Exit(exitUsage)
	}
}

func failUsage(problem, usage string) {
	fmt.Fprintln(os.Stderr, "findable:", problem)
	fmt.Fprintln(os.Stderr, usage)
	os.Exit(exitUsage)
}

func helpArg(s string) bool {
	return s == "-h" || s == "-help" || s == "--help"
}
