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

// Capabilities gate what the server may DO beyond reading (task 7.4).
// Read-only inspection is the default: without the explicit grant,
// scenario_execute is refused even when its source is attached, and
// with it every scenario must live inside an isolated namespace.
type Capabilities struct {
	ScenarioExecute    bool
	IsolatedNamespaces []string // allowed "<namespace>/" prefixes
}

// allowsScenario answers whether the scenario name sits inside one of
// the isolated namespaces.
func (c Capabilities) allowsScenario(name string) bool {
	for _, namespace := range c.IsolatedNamespaces {
		if len(name) > len(namespace)+1 &&
			name[:len(namespace)+1] == namespace+"/" {
			return true
		}
	}
	return false
}
