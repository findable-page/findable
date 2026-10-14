package mcpserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	jsonRPCVersion     = "2.0"
	mcpProtocolVersion = "2024-11-05"
	maxMCPLineBytes    = 10 * 1024 * 1024
	initialLineBytes   = 64 * 1024

	codeParseError     = -32700
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602

	defaultVirtualURL = "https://example.com"
	defaultCrawlLimit = 10
	maxCrawlLimit     = 50

	argURL       = "url"
	argKeyphrase = "keyphrase"
	typeString   = "string"
	errMissedURL = "Error: missing or empty 'url' argument"
)

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type jsonRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

type mcpSession struct {
	enc    *json.Encoder
	client *http.Client
}

func Run(in io.Reader, out io.Writer, client *http.Client) {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, initialLineBytes), maxMCPLineBytes)
	session := mcpSession{enc: json.NewEncoder(out), client: client}

	for scanner.Scan() {
		line := scanner.Bytes()
		if strings.TrimSpace(string(line)) == "" {
			continue
		}
		var err error
		var req jsonRPCRequest
		if jsonErr := json.Unmarshal(line, &req); jsonErr != nil {
			err = session.replyError(nil, codeParseError, "Parse error")
		} else {
			err = session.handle(req)
		}
		if err != nil {
			return
		}
	}
}

func (m mcpSession) reply(id, result any) error {
	return m.enc.Encode(jsonRPCResponse{JSONRPC: jsonRPCVersion, ID: id, Result: result})
}

func (m mcpSession) replyError(id any, code int, message string) error {
	return m.enc.Encode(jsonRPCResponse{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Error:   map[string]any{"code": code, "message": message},
	})
}

func (m mcpSession) handle(req jsonRPCRequest) error {
	switch req.Method {
	case "initialize":
		return m.reply(req.ID, initializeResult())
	case "notifications/initialized":
		return nil
	case "tools/list":
		return m.reply(req.ID, map[string]any{"tools": toolDefinitions()})
	case "tools/call":
		return m.callTool(req)
	case "prompts/list":
		return m.reply(req.ID, map[string]any{"prompts": promptDefinitions()})
	case "prompts/get":
		return m.reply(req.ID, auditPrompt(req.Params))
	default:
		return m.replyError(req.ID, codeMethodNotFound, "Method not found")
	}
}

func initializeResult() map[string]any {
	return map[string]any{
		"protocolVersion": mcpProtocolVersion,
		"capabilities": map[string]any{
			"tools":     map[string]any{},
			"prompts":   map[string]any{},
			"resources": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "seoscore-mcp",
			"version": "1.0.0",
		},
	}
}

func (m mcpSession) callTool(req jsonRPCRequest) error {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return m.replyError(req.ID, codeInvalidParams, "Invalid params")
	}
	text, isError := executeTool(params.Name, params.Arguments, m.client)
	return m.reply(req.ID, map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
		"isError": isError,
	})
}

func auditPrompt(rawParams json.RawMessage) map[string]any {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}
	_ = json.Unmarshal(rawParams, &params)
	targetURL := params.Arguments[argURL]
	if targetURL == "" {
		targetURL = defaultVirtualURL
	}
	text := fmt.Sprintf("Please use the 'audit_page' tool on %s and generate an executive summary with:\n1. Overall, SEO, AEO, and GEO scores.\n2. Top 3 critical fixes for Traditional SEO.\n3. Top 3 optimizations for Answer Engines (ChatGPT, Perplexity).\n4. Top 3 optimizations for Generative Search (Google AI Overviews).", targetURL)
	return map[string]any{
		"messages": []map[string]any{{
			"role":    "user",
			"content": map[string]any{"type": "text", "text": text},
		}},
	}
}
