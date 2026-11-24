package mcpserver

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func runLines(t *testing.T, lines ...string) []map[string]any {
	t.Helper()
	out := &bytes.Buffer{}
	Run(strings.NewReader(strings.Join(lines, "\n")+"\n"), out, &httpDefaultClient)
	var replies []map[string]any
	dec := json.NewDecoder(out)
	for dec.More() {
		var r map[string]any
		if err := dec.Decode(&r); err != nil {
			t.Fatalf("bad reply: %v\n%s", err, out.String())
		}
		replies = append(replies, r)
	}
	return replies
}

// JSON-RPC 2.0: a notification (no id) must never get a response, or the
// client sees a stray reply with id null on its stdio channel.
func TestMCPNeverRepliesToNotifications(t *testing.T) {
	replies := runLines(t,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":3,"reason":"timeout"}}`,
		`{"jsonrpc":"2.0","method":"notifications/roots/list_changed"}`,
	)
	if len(replies) != 0 {
		t.Fatalf("got %d replies to notifications, want 0: %v", len(replies), replies)
	}
}

// MCP requires every server to answer ping with an empty result.
func TestMCPAnswersPing(t *testing.T) {
	replies := runLines(t, `{"jsonrpc":"2.0","id":7,"method":"ping"}`)
	if len(replies) != 1 {
		t.Fatalf("got %d replies, want 1", len(replies))
	}
	if replies[0]["error"] != nil {
		t.Fatalf("ping returned error %v", replies[0]["error"])
	}
	if res, ok := replies[0]["result"].(map[string]any); !ok || len(res) != 0 || replies[0]["id"] != float64(7) {
		t.Fatalf("ping reply = %v, want id 7 and an empty result", replies[0])
	}
}

// Advertising a capability promises its methods; resources/list is not served.
func TestMCPDoesNotAdvertiseUnservedResources(t *testing.T) {
	replies := runLines(t, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	caps := replies[0]["result"].(map[string]any)["capabilities"].(map[string]any)
	if _, ok := caps["resources"]; ok {
		t.Error("initialize advertises resources, but resources/list returns Method not found")
	}
}

// The handler reads a keyphrase, so the schema must declare it.
func TestMCPCSVReportSchemaDeclaresKeyphrase(t *testing.T) {
	for _, tl := range toolDefinitions() {
		if tl["name"] != "generate_csv_report" {
			continue
		}
		props := tl["inputSchema"].(map[string]any)["properties"].(schemaProps)
		if _, ok := props[argKeyphrase]; !ok {
			t.Error("generate_csv_report reads 'keyphrase' but its inputSchema does not declare it")
		}
	}
}
