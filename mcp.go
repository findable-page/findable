package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var httpDefaultClient = http.Client{Timeout: 30 * time.Second}

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type jsonRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func runMCP(in io.Reader, out io.Writer, client *http.Client) {
	scanner := bufio.NewScanner(in)
	// Allow larger payloads up to 10MB
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	encoder := json.NewEncoder(out)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}

		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			resp := jsonRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error: map[string]interface{}{
					"code":    -32700,
					"message": "Parse error",
				},
			}
			encoder.Encode(resp)
			continue
		}

		handleMCPRequest(req, encoder, client)
	}
}

func handleMCPRequest(req jsonRPCRequest, enc *json.Encoder, client *http.Client) {
	switch req.Method {
	case "initialize":
		res := map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools":     map[string]interface{}{},
				"prompts":   map[string]interface{}{},
				"resources": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "seoscore-mcp",
				"version": "1.0.0",
			},
		}
		enc.Encode(jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: res})

	case "notifications/initialized":
		// No response required for notifications

	case "tools/list":
		tools := []map[string]interface{}{
			{
				"name":        "audit_page",
				"description": "Audits a live URL or local HTML file for SEO, AEO, and GEO signals, returning 0-100 scores and detailed diagnostics.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"url": map[string]interface{}{
							"type":        "string",
							"description": "The URL (http:// or https://) or local file path to audit.",
						},
						"keyphrase": map[string]interface{}{
							"type":        "string",
							"description": "Optional focus keyword to evaluate in title, description, and headings.",
						},
						"remote": map[string]interface{}{
							"type":        "boolean",
							"description": "If true, queries remote seoscore.tools instead of the local offline engine.",
						},
					},
					"required": []string{"url"},
				},
			},
			{
				"name":        "audit_html_content",
				"description": "Audits raw HTML string directly in-memory for SEO, AEO, and GEO signals without needing a network request.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"html": map[string]interface{}{
							"type":        "string",
							"description": "Raw HTML markup to evaluate.",
						},
						"url": map[string]interface{}{
							"type":        "string",
							"description": "Optional virtual URL context (defaults to https://example.com).",
						},
					},
					"required": []string{"html"},
				},
			},
			{
				"name":        "crawl_site",
				"description": "Crawls a website up to N internal pages and returns per-page SEO/AEO/GEO scores and aggregate statistics.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"url": map[string]interface{}{
							"type":        "string",
							"description": "The root URL to begin crawling internal links.",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Maximum number of pages to crawl (default 10, max 50).",
						},
						"keyphrase": map[string]interface{}{
							"type":        "string",
							"description": "Optional focus keyword to audit across pages.",
						},
					},
					"required": []string{"url"},
				},
			},
			{
				"name":        "generate_csv_report",
				"description": "Generates a full UTF-8 BOM CSV audit report string for a target URL or file.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"url": map[string]interface{}{
							"type":        "string",
							"description": "The target URL or local HTML file path.",
						},
					},
					"required": []string{"url"},
				},
			},
		}
		enc.Encode(jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{"tools": tools}})

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			enc.Encode(jsonRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: map[string]interface{}{
					"code":    -32602,
					"message": "Invalid params",
				},
			})
			return
		}

		resText, isError := executeTool(params.Name, params.Arguments, client)
		result := map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": resText,
				},
			},
			"isError": isError,
		}
		enc.Encode(jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: result})

	case "prompts/list":
		prompts := []map[string]interface{}{
			{
				"name":        "seo_audit_prompt",
				"description": "Analyze an audit report and generate prioritized fixes for Traditional SEO, AEO, and GEO.",
				"arguments": []map[string]interface{}{
					{
						"name":        "url",
						"description": "URL to audit and provide recommendations for",
						"required":    true,
					},
				},
			},
		}
		enc.Encode(jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{"prompts": prompts}})

	case "prompts/get":
		var params struct {
			Name      string            `json:"name"`
			Arguments map[string]string `json:"arguments"`
		}
		json.Unmarshal(req.Params, &params)
		targetURL := params.Arguments["url"]
		if targetURL == "" {
			targetURL = "https://example.com"
		}
		promptMsg := fmt.Sprintf("Please use the 'audit_page' tool on %s and generate an executive summary with:\n1. Overall, SEO, AEO, and GEO scores.\n2. Top 3 critical fixes for Traditional SEO.\n3. Top 3 optimizations for Answer Engines (ChatGPT, Perplexity).\n4. Top 3 optimizations for Generative Search (Google AI Overviews).", targetURL)
		res := map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"role": "user",
					"content": map[string]interface{}{
						"type": "text",
						"text": promptMsg,
					},
				},
			},
		}
		enc.Encode(jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: res})

	default:
		enc.Encode(jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: map[string]interface{}{
				"code":    -32601,
				"message": "Method not found",
			},
		})
	}
}

func executeTool(name string, args map[string]interface{}, client *http.Client) (string, bool) {
	switch name {
	case "audit_page":
		targetURL, _ := args["url"].(string)
		if strings.TrimSpace(targetURL) == "" {
			return "Error: missing or empty 'url' argument", true
		}
		keyphrase, _ := args["keyphrase"].(string)
		remote, _ := args["remote"].(bool)

		res, err := scanTarget(client, targetURL, keyphrase, remote)
		if err != nil {
			return fmt.Sprintf("Error auditing %s: %v", targetURL, err), true
		}
		summary := map[string]interface{}{
			"url":          targetURL,
			"overallScore": overall(res),
			"seoScore":     val(res.SEO),
			"aeoScore":     val(res.AEO),
			"geoScore":     val(res.GEO),
			"failedChecks": failedCount(res),
			"scan":         res,
		}
		b, _ := json.MarshalIndent(summary, "", "  ")
		return string(b), false

	case "audit_html_content":
		rawHTML, _ := args["html"].(string)
		if strings.TrimSpace(rawHTML) == "" {
			return "Error: missing or empty 'html' argument", true
		}
		targetURL, _ := args["url"].(string)
		if targetURL == "" {
			targetURL = "https://example.com"
		}

		res := AuditPage(targetURL, rawHTML, make(http.Header), 200)
		summary := map[string]interface{}{
			"url":          targetURL,
			"overallScore": overall(res),
			"seoScore":     val(res.SEO),
			"aeoScore":     val(res.AEO),
			"geoScore":     val(res.GEO),
			"failedChecks": failedCount(res),
			"scan":         res,
		}
		b, _ := json.MarshalIndent(summary, "", "  ")
		return string(b), false

	case "crawl_site":
		targetURL, _ := args["url"].(string)
		if strings.TrimSpace(targetURL) == "" {
			return "Error: missing or empty 'url' argument", true
		}
		limit := 10
		if l, ok := args["limit"].(float64); ok && l > 0 {
			limit = int(l)
			if limit > 50 {
				limit = 50
			}
		}
		keyphrase, _ := args["keyphrase"].(string)

		res0, err := scanTarget(client, targetURL, keyphrase, false)
		if err != nil {
			return fmt.Sprintf("Error crawling start URL %s: %v", targetURL, err), true
		}

		pages := []crawledPage{{url: targetURL, scan: res0}}
		visited := map[string]bool{targetURL: true}
		queue := append([]string{}, res0.InternalLinks...)

		for len(queue) > 0 && len(pages) < limit {
			curr := queue[0]
			queue = queue[1:]
			if visited[curr] {
				continue
			}
			visited[curr] = true

			s, err := scanTarget(client, curr, keyphrase, false)
			if err != nil {
				continue
			}
			pages = append(pages, crawledPage{url: curr, scan: s})
			for _, l := range s.InternalLinks {
				if !visited[l] {
					queue = append(queue, l)
				}
			}
		}

		type pageSummary struct {
			URL     string `json:"url"`
			Overall int    `json:"overall"`
			SEO     string `json:"seo"`
			AEO     string `json:"aeo"`
			GEO     string `json:"geo"`
			Failed  int    `json:"failed"`
		}
		var list []pageSummary
		for _, p := range pages {
			list = append(list, pageSummary{
				URL:     p.url,
				Overall: overall(p.scan),
				SEO:     score(p.scan.SEO),
				AEO:     score(p.scan.AEO),
				GEO:     score(p.scan.GEO),
				Failed:  failedCount(p.scan),
			})
		}
		output := map[string]interface{}{
			"rootUrl":    targetURL,
			"totalPages": len(pages),
			"pages":      list,
		}
		b, _ := json.MarshalIndent(output, "", "  ")
		return string(b), false

	case "generate_csv_report":
		targetURL, _ := args["url"].(string)
		if strings.TrimSpace(targetURL) == "" {
			return "Error: missing or empty 'url' argument", true
		}
		keyphrase, _ := args["keyphrase"].(string)
		res, err := scanTarget(client, targetURL, keyphrase, false)
		if err != nil {
			return fmt.Sprintf("Error generating CSV for %s: %v", targetURL, err), true
		}
		return csv(targetURL, res), false

	default:
		return fmt.Sprintf("Unknown tool name: %s", name), true
	}
}
