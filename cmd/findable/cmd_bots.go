package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/findable-page/findable/internal/fetch"
	"github.com/findable-page/findable/internal/report"
)

// botsCmd implements `findable bots`: live per-crawler access truth for one
// URL. One request per bot, sequential, polite delay between requests.
func botsCmd(args []string) {
	fs := flag.NewFlagSet("bots", flag.ContinueOnError)
	bots := fs.String("bots", "", "comma-separated bot names to probe (default: every known bot)")
	path := fs.String("path", "/", "path to request on the host")
	format := fs.String("format", formatText, "output format: text or json")
	allowPrivate := fs.Bool("allow-private", false, "allow probing localhost and private-network targets")
	delayMs := fs.Int("delay-ms", 250, "pause between bot requests, in milliseconds")
	commandUsage(fs, botsUsage)
	parseCommand(fs, args)
	if fs.NArg() != 1 {
		failUsage("bots needs one http(s) URL", botsUsage)
	}
	target := fs.Arg(0)
	if !fetch.IsHTTPURL(target) {
		failUsage("bots needs an http(s) URL, not a local file", botsUsage)
	}
	*format = strings.ToLower(*format)
	if *format != formatText && *format != formatJSON {
		fmt.Fprintf(os.Stderr, "findable: unknown bots format %q (want text or json)\n%s\n", *format, botsUsage)
		os.Exit(exitUsage)
	}
	var names []string
	if *bots != "" {
		names = strings.Split(*bots, ",")
	}
	verdicts, err := fetch.ProbeBots(newClient(*allowPrivate), target, fetch.ProbeOptions{
		Path:  *path,
		Bots:  names,
		Delay: time.Duration(*delayMs) * time.Millisecond,
	})
	if err != nil {
		failRuntime("%v", err)
	}
	if *format == formatJSON {
		fmt.Println(report.EncodeJSON(verdicts))
	} else {
		printBotVerdicts(verdicts)
	}
	for _, v := range verdicts {
		if v.Verdict == fetch.VerdictError {
			failRuntime("probe %s: %s", v.Bot, v.Error)
		}
	}
}

func printBotVerdicts(verdicts []fetch.BotVerdict) {
	fmt.Printf("%-18s %-11s %-18s %s\n", "BOT", "PURPOSE", "VERDICT", "EVIDENCE")
	for _, v := range verdicts {
		fmt.Printf("%-18s %-11s %-18s %s\n", v.Bot, v.Purpose, v.Verdict, botEvidence(v))
	}
	fmt.Println("\nrobots.txt verdicts come from the live robots.txt; server verdicts come from one real request per bot from YOUR IP.")
	fmt.Println("A WAF that verifies bot IPs may block a spoofed user agent the real bot would pass, and vice versa.")
}

func botEvidence(v fetch.BotVerdict) string {
	parts := []string{fmt.Sprintf("HTTP %d", v.Status)}
	if v.RobotsAllowed {
		parts = append(parts, "robots: allow")
	} else {
		parts = append(parts, "robots: disallow")
	}
	if v.Parity > 0 {
		parts = append(parts, fmt.Sprintf("%d%% content parity", v.Parity))
	}
	if v.Challenge != "" {
		parts = append(parts, v.Challenge)
	}
	if v.XNoIndex {
		parts = append(parts, "X-Robots-Tag: noindex")
	}
	if v.MetaNoIndex {
		parts = append(parts, "meta robots: noindex")
	}
	if v.Error != "" {
		parts = append(parts, v.Error)
	}
	return strings.Join(parts, ", ")
}
