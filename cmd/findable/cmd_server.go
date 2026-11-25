package main

import (
	"flag"
	"fmt"
	"os"
)

// mcpCmd runs the MCP server over stdio (`findable mcp [-allow-private]`).
func mcpCmd(args []string) {
	allowPrivate := false
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	fs.BoolVar(&allowPrivate, "allow-private", false, "allow scanning localhost and private-network targets")
	commandUsage(fs, mcpUsage)
	parseCommand(fs, args)
	if fs.NArg() != 0 {
		failUsage("mcp does not take a URL or other arguments", mcpUsage)
	}
	runMCP(newClient(allowPrivate))
}

// serveCmd runs the HTTP API server (`findable serve <addr>`).
func serveCmd(args []string) {
	if len(args) == 1 && helpArg(args[0]) {
		fmt.Fprintln(os.Stderr, serveUsage)
		fmt.Fprintln(os.Stderr, "127.0.0.1:8080 serves localhost. :8080 listens on every interface.")
		os.Exit(0)
	}
	if len(args) != 1 {
		failUsage("serve needs one address, for example 127.0.0.1:8080", serveUsage)
	}
	serveHTTP(args[0])
}
