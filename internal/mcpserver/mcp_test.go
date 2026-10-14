package mcpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

var httpDefaultClient = http.Client{Timeout: 30 * time.Second}

func TestMCPInitialize(t *testing.T) {
	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}` + "\n"
	in := bytes.NewBufferString(initReq)
	out := &bytes.Buffer{}

	Run(in, out, &httpDefaultClient)

	var res map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse MCP response: %v\nOutput: %s", err, out.String())
	}

	if res["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %v", res["jsonrpc"])
	}
	result, ok := res["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing result object")
	}
	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	if !ok || serverInfo["name"] != "seoscore-mcp" {
		t.Errorf("unexpected serverInfo: %v", serverInfo)
	}
}

func TestMCPToolsList(t *testing.T) {
	req := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}` + "\n"
	in := bytes.NewBufferString(req)
	out := &bytes.Buffer{}

	Run(in, out, &httpDefaultClient)

	var res map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse MCP tools/list response: %v", err)
	}
	result := res["result"].(map[string]interface{})
	tools := result["tools"].([]interface{})
	if len(tools) < 3 {
		t.Errorf("expected at least 3 tools, got %d", len(tools))
	}

	var foundAuditPage, foundAuditHTML bool
	for _, toolObj := range tools {
		tItem := toolObj.(map[string]interface{})
		name := tItem["name"].(string)
		if name == "audit_page" {
			foundAuditPage = true
		}
		if name == "audit_html_content" {
			foundAuditHTML = true
		}
	}
	if !foundAuditPage || !foundAuditHTML {
		t.Errorf("expected audit_page and audit_html_content in tools list")
	}
}

func TestMCPToolCallAuditHTMLContent(t *testing.T) {
	htmlSample := `<!DOCTYPE html><html><head><title>Test Page with 30 Character Title</title><meta name="description" content="This is a test description that meets the standard length requirement of one hundred and twenty characters for search engine evaluation."></head><body><h1>Main Title</h1><p>Content</p></body></html>`
	callReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "audit_html_content",
			"arguments": map[string]interface{}{
				"html": htmlSample,
				"url":  "https://test.example.com",
			},
		},
	}
	raw, _ := json.Marshal(callReq)
	in := bytes.NewBuffer(append(raw, '\n'))
	out := &bytes.Buffer{}

	Run(in, out, &httpDefaultClient)

	var res map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse tool call response: %v\nOutput: %s", err, out.String())
	}
	result := res["result"].(map[string]interface{})
	content := result["content"].([]interface{})
	if len(content) == 0 {
		t.Fatalf("expected content in result")
	}
	first := content[0].(map[string]interface{})
	text := first["text"].(string)
	if !strings.Contains(text, "overall") && !strings.Contains(text, "seo") {
		t.Errorf("expected audit result text, got: %s", text)
	}
}
