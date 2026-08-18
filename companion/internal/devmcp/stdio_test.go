package devmcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func serveLines(t *testing.T, lines ...string) []map[string]any {
	t.Helper()
	var out bytes.Buffer
	if err := ServeStdio(strings.NewReader(strings.Join(lines, "\n")),
		&out); err != nil {
		t.Fatalf("serve: %v", err)
	}
	var responses []map[string]any
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		var resp map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			t.Fatalf("response is not json: %v", err)
		}
		responses = append(responses, resp)
	}
	return responses
}

// TestStdioServerExposesExactlyTheFiveInspectionTools: initialize
// works and tools/list names the closed surface of task 7.3 — nothing
// more, nothing less.
func TestStdioServerExposesExactlyTheFiveInspectionTools(t *testing.T) {
	responses := serveLines(t,
		`{"jsonrpc":"2.0","id":1,"method":"initialize"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
	info := responses[0]["result"].(map[string]any)["serverInfo"].(map[string]any)
	if info["name"] != "mtga-dev-mcp" {
		t.Fatalf("handshake must identify the dev server: %v", info)
	}
	tools := responses[1]["result"].(map[string]any)["tools"].([]any)
	want := []string{"graphs_list", "scenario_execute", "run_timeline",
		"events_validate", "fixtures_diagnose"}
	if len(tools) != len(want) {
		t.Fatalf("the tool surface must be exactly five: %v", tools)
	}
	for i, tool := range tools {
		if tool.(map[string]any)["name"] != want[i] {
			t.Fatalf("tool %d must be %s: %v", i, want[i], tool)
		}
	}
}

// TestStdioServerSurvivesGarbageAndUnknownMethods: a malformed line
// answers an error and the loop keeps serving.
func TestStdioServerSurvivesGarbageAndUnknownMethods(t *testing.T) {
	responses := serveLines(t,
		`{not json`,
		`{"jsonrpc":"2.0","id":7,"method":"shell/execute"}`,
		`{"jsonrpc":"2.0","id":8,"method":"tools/list"}`)
	if len(responses) != 3 {
		t.Fatalf("every line must be answered: %d", len(responses))
	}
	parseErr := responses[0]["error"].(map[string]any)
	if parseErr["code"].(float64) != -32700 {
		t.Fatalf("garbage must answer a parse error: %v", parseErr)
	}
	unknown := responses[1]["error"].(map[string]any)
	if unknown["code"].(float64) != -32601 {
		t.Fatalf("unknown methods must answer method-not-found: %v", unknown)
	}
	if responses[2]["result"] == nil {
		t.Fatal("the loop must keep serving after errors")
	}
}
