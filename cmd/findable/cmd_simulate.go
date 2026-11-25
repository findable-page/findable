package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/findable-page/findable/internal/report"
	"github.com/findable-page/findable/internal/simulate"
)

const simulateUsage = "usage: findable simulate -query TEXT -brand NAME -domain HOST [-engine URL] [-format text|json]"

func simulateCmd(args []string) {
	fs := flag.NewFlagSet("simulate", flag.ContinueOnError)
	query := fs.String("query", "", "prompt to send to the engine")
	brand := fs.String("brand", "", "brand name to look for in the answer")
	domain := fs.String("domain", "", "site host that counts as a direct citation")
	engine := fs.String("engine", os.Getenv("FINDABLE_SIMULATE_ENGINE"), "answer engine URL (or FINDABLE_SIMULATE_ENGINE)")
	format := fs.String("format", formatText, "output format: text or json")
	fs.Usage = func() { fmt.Fprintln(os.Stderr, simulateUsage) }
	if err := fs.Parse(args); err != nil {
		os.Exit(exitUsage)
	}
	if strings.TrimSpace(*query) == "" || strings.TrimSpace(*brand) == "" || strings.TrimSpace(*domain) == "" || strings.TrimSpace(*engine) == "" {
		fmt.Fprintln(os.Stderr, simulateUsage)
		os.Exit(exitUsage)
	}
	answer, err := simulate.Query(context.Background(), nil, *engine, *query, os.Getenv("FINDABLE_SIMULATE_TOKEN"))
	if err != nil {
		failRuntime("%v", err)
	}
	result := simulate.ReadAnswer(answer, *query, *engine, *brand, *domain)
	switch strings.ToLower(*format) {
	case formatJSON:
		fmt.Print(report.EncodeJSON(result))
	case formatText:
		printSimulation(result)
	default:
		fmt.Fprintf(os.Stderr, "findable: unknown simulate format %q (want text or json)\n", *format)
		os.Exit(exitUsage)
	}
}

func printSimulation(result simulate.Report) {
	mention := "no"
	if result.BrandMentioned {
		mention = "yes (" + result.Sentiment + ")"
	}
	cited := "no"
	if result.Cited {
		cited = "yes"
	}
	fmt.Printf("query %q\nbrand %s  mentioned %s\ncited %s\n", result.Query, result.Brand, mention, cited)
	if len(result.CompetitorDomains) == 0 {
		fmt.Println("competitors none")
		return
	}
	fmt.Println("competitors " + strings.Join(result.CompetitorDomains, ", "))
}
