package main

import (
	"fmt"

	"github.com/findable-page/findable/internal/audit"
	"github.com/findable-page/findable/internal/config"
	"github.com/findable-page/findable/internal/report"
)

// printFixList prints the prioritized fix list, dropping checks ignored via
// findable.json.
func printFixList(pages []audit.Page, cfg config.Config, top int, minSeverity audit.Severity) {
	findings := report.Findings(pages)
	if len(cfg.Ignore) > 0 {
		findings = report.FilterFindings(findings, cfg.Ignores)
	}
	if out := report.FormatFixList(findings, top, minSeverity); out != "" {
		fmt.Println()
		fmt.Print(out)
	}
}
