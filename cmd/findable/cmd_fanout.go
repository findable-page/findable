package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/findable-page/findable/internal/audit"
	"github.com/findable-page/findable/internal/fetch"
	"github.com/findable-page/findable/internal/report"
)

const fanoutUsage = "usage: findable fanout -query TOPIC [-format text|json] [-allow-private] <url or file>"

func fanoutCmd(args []string) {
	fs := flag.NewFlagSet("fanout", flag.ContinueOnError)
	query := fs.String("query", "", "topic to expand into sub-intents")
	format := fs.String("format", formatText, "output format: text or json")
	allowPrivate := fs.Bool("allow-private", false, "allow fetching localhost and private-network targets")
	fs.Usage = func() { fmt.Fprintln(os.Stderr, fanoutUsage) }
	if err := fs.Parse(args); err != nil {
		os.Exit(exitUsage)
	}
	if fs.NArg() != 1 || strings.TrimSpace(*query) == "" {
		fmt.Fprintln(os.Stderr, fanoutUsage)
		os.Exit(exitUsage)
	}
	target := fs.Arg(0)
	html, err := loadPageHTML(newClient(*allowPrivate), target)
	if err != nil {
		failRuntime("%v", err)
	}
	result := audit.QueryFanout(target, html, *query)
	switch strings.ToLower(*format) {
	case formatJSON:
		fmt.Print(report.EncodeJSON(result))
	case formatText:
		printFanout(result)
	default:
		fmt.Fprintf(os.Stderr, "findable: unknown fanout format %q (want text or json)\n", *format)
		os.Exit(exitUsage)
	}
}

func loadPageHTML(client *http.Client, target string) (string, error) {
	if !fetch.IsHTTPURL(target) {
		body, err := os.ReadFile(strings.TrimPrefix(target, "file://")) //nolint:gosec // reading the user-named file is the feature
		if err != nil {
			return "", fmt.Errorf("read file: %w", err)
		}
		return string(body), nil
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "findable")
	resp, err := client.Do(req) //nolint:gosec // fetching the user-named URL is the feature; see the SSRF policy in SECURITY.md
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", target, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetch %s: HTTP %d", target, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, fetch.MaxPageBytes+1))
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if len(body) > fetch.MaxPageBytes {
		return "", fmt.Errorf("%s: %w", target, fetch.ErrPageTooLarge)
	}
	return string(body), nil
}

func printFanout(result audit.FanoutReport) {
	fmt.Printf("topic %q  covered %d/%d\n", result.Topic, result.Covered, result.Total)
	fmt.Printf("%-12s %-8s %s\n", "INTENT", "COVERED", "SUB-QUERY")
	for _, intent := range result.Intents {
		mark := "no"
		if intent.Covered {
			mark = "yes"
		}
		line := intent.Query
		if intent.Where != "" {
			line += " (" + intent.Where + ")"
		}
		fmt.Printf("%-12s %-8s %s\n", intent.ID, mark, line)
	}
}
