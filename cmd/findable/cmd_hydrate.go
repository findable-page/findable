package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

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
	if err := fs.Parse(args); err != nil {
		os.Exit(exitUsage)
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, hydrateUsage)
		os.Exit(exitUsage)
	}
	if (*renderedPath == "") == (*renderCmd == "") {
		fmt.Fprintln(os.Stderr, "findable: pass exactly one of -rendered or -render-cmd")
		os.Exit(exitUsage)
	}
	target := fs.Arg(0)
	client := newClient(*allowPrivate)
	raw, err := fetch.ReadHTML(client, target)
	if err != nil {
		failRuntime("%v", err)
	}
	rendered, err := renderedHTML(client, *renderedPath, *renderCmd, target)
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

const renderTimeout = 30 * time.Second

func renderedHTML(client *http.Client, path, command, target string) (string, error) {
	if path != "" {
		return fetch.ReadHTML(client, path)
	}
	ctx, cancel := context.WithTimeout(context.Background(), renderTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, target) //nolint:gosec // the operator named the renderer
	return readCommandOutput(cmd)
}

func readCommandOutput(cmd *exec.Cmd) (string, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("render command: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("render command: %w", err)
	}
	out, readErr := io.ReadAll(io.LimitReader(stdout, fetch.MaxPageBytes+1))
	if len(out) > fetch.MaxPageBytes && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	waitErr := cmd.Wait()
	if len(out) > fetch.MaxPageBytes {
		return "", fmt.Errorf("render command: %w", fetch.ErrPageTooLarge)
	}
	if readErr != nil {
		return "", fmt.Errorf("render command: %w", readErr)
	}
	if waitErr != nil {
		return "", fmt.Errorf("render command: %w", waitErr)
	}
	return string(out), nil
}

func printHydration(report audit.HydrationReport) {
	fmt.Printf("raw words %d  rendered words %d  text ratio %.2f\n", report.RawWords, report.RenderedWords, report.TextRatio)
	for _, check := range report.Checks {
		fmt.Printf("%-28s %s  %s\n", check.ID, audit.Status(check), check.Label)
	}
}
