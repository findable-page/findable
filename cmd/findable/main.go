// findable scans a site locally and saves the report CSV.
package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	formatText = "text"
	formatJSON = "json"
)

// subcommands dispatch on os.Args[1]; anything else falls through to the
// legacy flag parser, which exits 2 on an unexpected positional argument.
// Entries are added as each Phase 1 slice lands.
var subcommands = map[string]func([]string){
	"audit":    auditCmd,
	"bots":     botsCmd,
	"diff":     diffCmd,
	"fanout":   fanoutCmd,
	"hydrate":  hydrateCmd,
	"mcp":      mcpCmd,
	"serve":    serveCmd,
	"simulate": simulateCmd,
}

func main() {
	if len(os.Args) > 1 {
		if run, ok := subcommands[os.Args[1]]; ok {
			run(os.Args[2:])
			return
		}
	}
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") && !strings.ContainsAny(os.Args[1], "/.:") && (len(os.Args) == 2 || strings.HasPrefix(os.Args[2], "-")) {
		fmt.Fprintf(os.Stderr, "findable: unknown command %q\nRun \"findable -h\" to list commands.\n", os.Args[1])
		os.Exit(exitUsage)
	}
	legacyMain(os.Args[1:])
}
