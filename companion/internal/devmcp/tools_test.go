package devmcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func wiredServer() *Server {
	return &Server{Caps: Capabilities{ScenarioExecute: true,
		IsolatedNamespaces: []string{"dev"}}, Deps: ToolDeps{
		ListGraphs: func(context.Context) ([]string, error) {
			return []string{"collection-sync", "approved-export"}, nil
		},
		ExecuteScenario: func(_ context.Context, name string) (string, error) {
			return "scenario " + name + ": succeeded", nil
		},
		RunTimeline: func(_ context.Context, runID string) ([]string, error) {
			if runID != "run-1" {
				return nil, errors.New("unknown run")
			}
			return []string{"valida ok", "exporta ok"}, nil
		},
		ValidateEvent: func(_ context.Context, raw json.RawMessage) error {
			if !json.Valid(raw) {
				return errors.New("not json")
			}
			return nil
		},
		DiagnoseFixture: func(_ context.Context, name string) (string, error) {
			return "fixture " + name + ": 1 unknown card", nil
		},
	}}
}

func callTool(t *testing.T, s *Server, body string) map[string]any {
	t.Helper()
	var out bytes.Buffer
	if err := s.Serve(strings.NewReader(body), &out); err != nil {
		t.Fatalf("serve: %v", err)
	}
	scanner := bufio.NewScanner(&out)
	if !scanner.Scan() {
		t.Fatal("no response written")
	}
	var resp map[string]any
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		t.Fatalf("response: %v", err)
	}
	return resp
}

func TestToolsDispatchThroughTheirSources(t *testing.T) {
	s := wiredServer()
	calls := map[string]string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"graphs_list"}}`:                                              "collection-sync\napproved-export",
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"scenario_execute","arguments":{"scenario":"dev/export"}}}`:   "scenario dev/export: succeeded",
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"run_timeline","arguments":{"run_id":"run-1"}}}`:              "valida ok\nexporta ok",
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"events_validate","arguments":{"event":{"kind":"step"}}}}`:    "valid",
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"fixtures_diagnose","arguments":{"fixture":"unknown-card"}}}`: "fixture unknown-card: 1 unknown card",
	}
	for body, want := range calls {
		if got := toolText(t, callTool(t, s, body)); got != want {
			t.Fatalf("dispatch broke:\nwant %q\ngot  %q", want, got)
		}
	}
}

func TestToolsRefuseTheUnknownAndTheUnattached(t *testing.T) {
	s := wiredServer()
	resp := callTool(t, s,
		`{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"sql_execute"}}`)
	if resp["error"] == nil {
		t.Fatal("tools outside the closed surface must be refused")
	}
	bare := &Server{}
	resp = callTool(t, bare,
		`{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"graphs_list"}}`)
	if resp["error"] == nil {
		t.Fatal("an unattached source must answer a clean error")
	}
	resp = callTool(t, s,
		`{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"scenario_execute","arguments":{}}}`)
	if resp["error"] == nil {
		t.Fatal("scenario_execute without a scenario must be refused")
	}
}
