package main

import "flag"

// auditCmd implements `findable audit`: the same engine and flags as the
// legacy path, reached through a subcommand.
func auditCmd(args []string) {
	var o options
	fs := flag.NewFlagSet("audit", flag.ContinueOnError)
	format, minSeverity := registerAuditFlags(fs, &o)
	commandUsage(fs, auditUsage)
	parseCommand(fs, args)
	finishOptions(&o, *format, minSeverity)
	runAuditTarget(newClient(o.allowPrivate), fs.Args(), o, auditUsage)
}
