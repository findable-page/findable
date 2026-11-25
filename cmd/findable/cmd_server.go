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
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		os.Exit(exitUsage)
	}
	runMCP(newClient(allowPrivate))
}

// serveCmd runs the HTTP API server (`findable serve <addr>`).
func serveCmd(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: findable serve <host:port>")
		os.Exit(exitUsage)
	}
	serveHTTP(args[0])
}
