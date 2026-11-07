// findable scans a site locally and saves the report CSV.
package main

import (
	"os"
)

const usageLine = "findable: is your site findable by Google and by AI answer engines?\n" +
	"usage: findable [command] [flags] <url or file>\n" +
	"\n" +
	"commands:\n" +
	"  audit    (default) audit one page, a crawl or a sitemap\n" +
	"  crawl    crawl the site and run cross-page checks\n" +
	"  bots     probe how AI crawlers are actually treated by the server\n" +
	"  render   compare raw HTML with the JS-rendered DOM\n" +
	"  fix      generate fix files (robots.txt, llms.txt, JSON-LD, head snippets)\n" +
	"  compare  diff two JSON reports\n" +
	"  mcp      run MCP server over stdio\n" +
	"  serve    run HTTP API server\n" +
	"\n" +
	"legacy: findable [-o dir] [-format csv|json|sarif] [-fail-under N] [-q] [-k keyphrase] [-site] [-crawl] [-sitemap url] [-limit N] [-serve 127.0.0.1:8080] [-mcp] [-allow-private] <url or file>"

// subcommands dispatch on os.Args[1]; anything else falls through to the
// legacy flag parser, which exits 2 on an unexpected positional argument.
// Entries are added as each Phase 1 slice lands.
var subcommands = map[string]func([]string){
	"audit": auditCmd,
	"mcp":   mcpCmd,
	"serve": serveCmd,
}

func main() {
	if len(os.Args) > 1 {
		if run, ok := subcommands[os.Args[1]]; ok {
			run(os.Args[2:])
			return
		}
	}
	legacyMain(os.Args[1:])
}
