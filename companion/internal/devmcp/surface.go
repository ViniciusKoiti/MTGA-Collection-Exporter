package devmcp

import (
	"context"
	"encoding/json"
)

// ToolDeps are the sources behind the five inspection tools — the
// ONLY tools the dev MCP exposes (task 7.3). A nil source answers a
// clean "not attached" error instead of pretending.
type ToolDeps struct {
	ListGraphs      func(context.Context) ([]string, error)
	ExecuteScenario func(context.Context, string) (string, error)
	RunTimeline     func(context.Context, string) ([]string, error)
	ValidateEvent   func(context.Context, json.RawMessage) error
	DiagnoseFixture func(context.Context, string) (string, error)
}

// toolNames is the closed tool surface, in stable order.
func toolNames() []string {
	return []string{"graphs_list", "scenario_execute", "run_timeline",
		"events_validate", "fixtures_diagnose"}
}
