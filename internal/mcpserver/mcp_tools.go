package mcpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/nadimtuhin/findable/internal/audit"
	"github.com/nadimtuhin/findable/internal/fetch"
	"github.com/nadimtuhin/findable/internal/report"
)

type schemaProps map[string]map[string]any

func prop(kind, description string) map[string]any {
	return map[string]any{"type": kind, "description": description}
}

func tool(name, description string, props schemaProps, required string) map[string]any {
	return map[string]any{
		"name":        name,
		"description": description,
		"inputSchema": map[string]any{
			"type":       "object",
			"properties": props,
			"required":   []string{required},
		},
	}
}

func toolDefinitions() []map[string]any {
	return []map[string]any{
		tool("audit_page",
			"Audits a live URL or local HTML file for SEO, AEO, and GEO signals, returning 0-100 scores and detailed diagnostics.",
			schemaProps{
				argURL:       prop(typeString, "The URL (http:// or https://) or local file path to audit."),
				argKeyphrase: prop(typeString, "Optional focus keyword to evaluate in title, description, and headings."),
				"remote":     prop("boolean", "If true, queries remote seoscore.tools instead of the local offline engine."),
				"site":       prop("boolean", "If true, also checks robots.txt (search and AI crawlers) and /llms.txt for a live URL."),
			}, argURL),
		tool("audit_html_content",
			"Audits raw HTML string directly in-memory for SEO, AEO, and GEO signals without needing a network request.",
			schemaProps{
				"html": prop(typeString, "Raw HTML markup to evaluate."),
				argURL: prop(typeString, "Optional virtual URL context (defaults to https://example.com)."),
			}, "html"),
		tool("crawl_site",
			"Crawls a website up to N internal pages and returns per-page SEO/AEO/GEO scores and aggregate statistics.",
			schemaProps{
				argURL:       prop(typeString, "The root URL to begin crawling internal links."),
				"limit":      prop("integer", "Maximum number of pages to crawl (default 10, max 50)."),
				argKeyphrase: prop(typeString, "Optional focus keyword to audit across pages."),
			}, argURL),
		tool("generate_csv_report",
			"Generates a full UTF-8 BOM CSV audit report string for a target URL or file.",
			schemaProps{
				argURL: prop(typeString, "The target URL or local HTML file path."),
			}, argURL),
	}
}

func promptDefinitions() []map[string]any {
	return []map[string]any{{
		"name":        "seo_audit_prompt",
		"description": "Analyze an audit report and generate prioritized fixes for Traditional SEO, AEO, and GEO.",
		"arguments": []map[string]any{{
			"name":        argURL,
			"description": "URL to audit and provide recommendations for",
			"required":    true,
		}},
	}}
}

type toolArgs map[string]any

func (a toolArgs) str(key string) string {
	v, _ := a[key].(string)
	return v
}

func (a toolArgs) requiredURL() (string, bool) {
	u := a.str(argURL)
	return u, strings.TrimSpace(u) != ""
}

func executeTool(name string, args map[string]any, client *http.Client) (text string, isError bool) {
	a := toolArgs(args)
	switch name {
	case "audit_page":
		return toolAuditPage(a, client)
	case "audit_html_content":
		return toolAuditHTML(a)
	case "crawl_site":
		return toolCrawlSite(a, client)
	case "generate_csv_report":
		return toolCSVReport(a, client)
	default:
		return "Unknown tool name: " + name, true
	}
}

func toolResult(v any) (text string, isError bool) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "Error encoding result: " + err.Error(), true
	}
	return string(b), false
}

func scanSummary(targetURL string, res audit.Scan) map[string]any {
	return map[string]any{
		"url":          targetURL,
		"overallScore": audit.Overall(res),
		"seoScore":     audit.ScoreValue(res.SEO),
		"aeoScore":     audit.ScoreValue(res.AEO),
		"geoScore":     audit.ScoreValue(res.GEO),
		"failedChecks": audit.FailedCount(res),
		"scan":         res,
	}
}

func toolAuditPage(a toolArgs, client *http.Client) (text string, isError bool) {
	targetURL, ok := a.requiredURL()
	if !ok {
		return errMissedURL, true
	}
	remote, _ := a["remote"].(bool)
	res, err := fetch.Scan(client, targetURL, a.str(argKeyphrase), remote)
	if err != nil {
		return fmt.Sprintf("Error auditing %s: %v", targetURL, err), true
	}
	if site, _ := a["site"].(bool); site && fetch.IsHTTPURL(targetURL) {
		res.Site = fetch.SiteChecks(client, targetURL)
	}
	return toolResult(scanSummary(targetURL, res))
}

func toolAuditHTML(a toolArgs) (text string, isError bool) {
	rawHTML := a.str("html")
	if strings.TrimSpace(rawHTML) == "" {
		return "Error: missing or empty 'html' argument", true
	}
	targetURL := a.str(argURL)
	if targetURL == "" {
		targetURL = defaultVirtualURL
	}
	res := audit.Analyze(targetURL, rawHTML, make(http.Header), http.StatusOK)
	return toolResult(scanSummary(targetURL, res))
}

func crawlLimit(a toolArgs) int {
	limit := defaultCrawlLimit
	if l, ok := a["limit"].(float64); ok && l > 0 {
		limit = min(int(l), maxCrawlLimit)
	}
	return limit
}

type pageSummary struct {
	URL     string `json:"url"`
	Overall int    `json:"overall"`
	SEO     string `json:"seo"`
	AEO     string `json:"aeo"`
	GEO     string `json:"geo"`
	Failed  int    `json:"failed"`
}

func toolCrawlSite(a toolArgs, client *http.Client) (text string, isError bool) {
	targetURL, ok := a.requiredURL()
	if !ok {
		return errMissedURL, true
	}
	keyphrase := a.str(argKeyphrase)
	first, err := fetch.Scan(client, targetURL, keyphrase, false)
	if err != nil {
		return fmt.Sprintf("Error crawling start URL %s: %v", targetURL, err), true
	}
	pages := followLinks(client, audit.Page{URL: targetURL, Scan: first}, keyphrase, crawlLimit(a))

	var list []pageSummary
	for _, p := range pages {
		list = append(list, pageSummary{
			URL:     p.URL,
			Overall: audit.Overall(p.Scan),
			SEO:     audit.ScoreText(p.Scan.SEO),
			AEO:     audit.ScoreText(p.Scan.AEO),
			GEO:     audit.ScoreText(p.Scan.GEO),
			Failed:  audit.FailedCount(p.Scan),
		})
	}
	return toolResult(map[string]any{
		"rootUrl":    targetURL,
		"totalPages": len(pages),
		"pages":      list,
	})
}

func followLinks(client *http.Client, first audit.Page, keyphrase string, limit int) []audit.Page {
	pages := []audit.Page{first}
	visited := map[string]bool{first.URL: true}
	queue := append([]string{}, first.Scan.InternalLinks...)

	for len(queue) > 0 && len(pages) < limit {
		curr := queue[0]
		queue = queue[1:]
		if visited[curr] {
			continue
		}
		visited[curr] = true

		s, err := fetch.Scan(client, curr, keyphrase, false)
		if err != nil {
			continue
		}
		pages = append(pages, audit.Page{URL: curr, Scan: s})
		for _, l := range s.InternalLinks {
			if !visited[l] {
				queue = append(queue, l)
			}
		}
	}
	return pages
}

func toolCSVReport(a toolArgs, client *http.Client) (text string, isError bool) {
	targetURL, ok := a.requiredURL()
	if !ok {
		return errMissedURL, true
	}
	res, err := fetch.Scan(client, targetURL, a.str(argKeyphrase), false)
	if err != nil {
		return fmt.Sprintf("Error generating CSV for %s: %v", targetURL, err), true
	}
	return report.CSV(targetURL, res), false
}
