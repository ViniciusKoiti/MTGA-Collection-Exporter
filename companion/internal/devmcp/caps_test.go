package devmcp

import "testing"

func toolText(t *testing.T, resp map[string]any) string {
	t.Helper()
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected a result: %v", resp)
	}
	content := result["content"].([]any)[0].(map[string]any)
	return content["text"].(string)
}

// TestScenarioExecutionIsACapabilityRestrictedToIsolatedNamespaces:
// read-only is the default, and even the grant only reaches isolated
// namespaces (task 7.4).
func TestScenarioExecutionIsACapabilityRestrictedToIsolatedNamespaces(t *testing.T) {
	readOnly := wiredServer()
	readOnly.Caps = Capabilities{}
	resp := callTool(t, readOnly,
		`{"jsonrpc":"2.0","id":20,"method":"tools/call","params":{"name":"scenario_execute","arguments":{"scenario":"dev/export"}}}`)
	if resp["error"] == nil {
		t.Fatal("without the capability the server must stay read-only")
	}
	if toolText(t, callTool(t, readOnly,
		`{"jsonrpc":"2.0","id":21,"method":"tools/call","params":{"name":"graphs_list"}}`)) == "" {
		t.Fatal("read-only inspection must keep working")
	}
	granted := wiredServer()
	resp = callTool(t, granted,
		`{"jsonrpc":"2.0","id":22,"method":"tools/call","params":{"name":"scenario_execute","arguments":{"scenario":"prod/export"}}}`)
	if resp["error"] == nil {
		t.Fatal("scenarios outside every isolated namespace must be refused")
	}
	resp = callTool(t, granted,
		`{"jsonrpc":"2.0","id":23,"method":"tools/call","params":{"name":"scenario_execute","arguments":{"scenario":"dev/smoke"}}}`)
	if resp["error"] != nil {
		t.Fatalf("an isolated scenario with the grant must run: %v", resp["error"])
	}
}
