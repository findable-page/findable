package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/findable-page/findable/internal/audit"
	"github.com/findable-page/findable/internal/fetch"
	"github.com/findable-page/findable/internal/report"
)

const hydrateUsage = "usage: findable hydrate [-rendered file | -render-cmd program] [-format text|json] [-allow-private] <url or file>"

func hydrateCmd(args []string) {
	fs := flag.NewFlagSet("hydrate", flag.ContinueOnError)
	renderedPath := fs.String("rendered", "", "HTML file of the page after client rendering")
	renderCmd := fs.String("render-cmd", "", "program that prints rendered HTML; the target is its first argument")
	format := fs.String("format", "text", "output format: text or json")
	allowPrivate := fs.Bool("allow-private", false, "allow fetching localhost and private-network targets")
	fs.Usage = func() { fmt.Fprintln(os.Stderr, hydrateUsage) }
	if err := fs.Parse(args); err != nil || fs.NArg() != 1 {
		os.Exit(exitUsage)
	}
	if (*renderedPath == "") == (*renderCmd == "") {
		fmt.Fprintln(os.Stderr, "findable: pass exactly one of -rendered or -render-cmd")
		os.Exit(exitUsage)
	}
	target := fs.Arg(0)
	raw, err := fetch.ReadHTML(newClient(*allowPrivate), target)
	if err != nil {
		failRuntime("%v", err)
	}
	rendered, err := renderedHTML(*renderedPath, *renderCmd, target)
	if err != nil {
		failRuntime("%v", err)
	}
	result := audit.CompareHydration(target, raw, rendered)
	switch strings.ToLower(*format) {
	case "json":
		fmt.Print(report.EncodeJSON(result))
	case "text":
		printHydration(result)
	default:
		fmt.Fprintf(os.Stderr, "findable: unknown hydrate format %q (want text or json)\n", *format)
		os.Exit(exitUsage)
	}
}

func renderedHTML(path, command, target string) (string, error) {
	if path != "" {
		return fetch.ReadHTML(newClient(true), path)
	}
	cmd := exec.CommandContext(context.Background(), command, target) //nolint:gosec // the operator named the renderer
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("render command: %w", err)
	}
	return string(out), nil
}

func printHydration(report audit.HydrationReport) {
	fmt.Printf("raw words %d  rendered words %d  text ratio %.2f\n", report.RawWords, report.RenderedWords, report.TextRatio)
	for _, check := range report.Checks {
		fmt.Printf("%-28s %s  %s\n", check.ID, audit.Status(check), check.Label)
	}
}
