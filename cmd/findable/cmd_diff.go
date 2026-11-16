package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/findable-page/findable/internal/audit"
	"github.com/findable-page/findable/internal/report"
)

const diffUsage = "usage: findable diff [-fail-on-regression] [-max-drop N] [-markdown] <baseline.json> <current.json>"

// diffCmd implements `findable diff`: compare a current JSON report against a
// stored baseline and flag score regressions and checks that stopped passing.
func diffCmd(args []string) {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	failOnRegression := fs.Bool("fail-on-regression", false, "exit 1 when a check regresses or the overall score drops by more than -max-drop")
	maxDrop := fs.Int("max-drop", 0, "allowed overall score drop per page before -fail-on-regression trips")
	markdown := fs.Bool("markdown", false, "print a markdown table instead of the terminal table")
	fs.Usage = func() { fmt.Fprintln(os.Stderr, diffUsage) }
	if err := fs.Parse(args); err != nil {
		os.Exit(exitUsage)
	}
	if fs.NArg() != 2 {
		fmt.Fprintln(os.Stderr, diffUsage)
		os.Exit(exitUsage)
	}
	baseline, err := readReportFile(fs.Arg(0))
	if err != nil {
		failRuntime("%v", err)
	}
	current, err := readReportFile(fs.Arg(1))
	if err != nil {
		failRuntime("%v", err)
	}
	d := report.DiffReports(baseline, current)
	if *markdown {
		fmt.Print(d.Markdown())
	} else {
		fmt.Print(d.Terminal())
	}
	if *failOnRegression && d.Tripped(*maxDrop) {
		fmt.Fprintln(os.Stderr, "findable: score regression against baseline")
		os.Exit(1)
	}
}

func readReportFile(path string) ([]audit.Page, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path comes from the operator's CLI argument
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	pages, err := report.LoadReportPages(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return pages, nil
}
