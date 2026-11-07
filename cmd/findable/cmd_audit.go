package main

import (
	"flag"
	"fmt"
	"os"
)

// auditCmd implements `findable audit`: the same engine and flags as the
// legacy path, reached through a subcommand.
func auditCmd(args []string) {
	var o options
	fs := flag.NewFlagSet("audit", flag.ContinueOnError)
	format, minSeverity := registerAuditFlags(fs, &o)
	fs.Usage = func() { fmt.Fprintln(os.Stderr, usageLine) }
	if err := fs.Parse(args); err != nil {
		os.Exit(exitUsage)
	}
	finishOptions(&o, *format, minSeverity)
	runAuditTarget(newClient(o.allowPrivate), fs.Args(), o)
}
