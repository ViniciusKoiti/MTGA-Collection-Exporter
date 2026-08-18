package devmcp

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestLimitsBoundRequestResponsePaginationAndDuration: every session
// dimension has a hard bound (task 7.5).
func TestLimitsBoundRequestResponsePaginationAndDuration(t *testing.T) {
	s := wiredServer()
	s.Lim = Limits{MaxRequestBytes: 128, MaxResponseBytes: 256,
		MaxCallDuration: 30 * time.Millisecond, MaxPageItems: 3}
	s.Deps.ListGraphs = func(context.Context) ([]string, error) {
		names := make([]string, 50)
		for i := range names {
			names[i] = fmt.Sprintf("graph-%02d", i)
		}
		return names, nil
	}
	s.Deps.RunTimeline = func(ctx context.Context, _ string) ([]string, error) {
		select { // a slow source that respects its deadline
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
			return []string{"never"}, nil
		}
	}
	oversized := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"graphs_list","arguments":{"scenario":"` +
		strings.Repeat("x", 200) + `"}}}`
	if resp := callTool(t, s, oversized); resp["error"] == nil {
		t.Fatal("oversized requests must be refused")
	}
	paged := toolText(t, callTool(t, s,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"graphs_list"}}`))
	if !strings.Contains(paged, "(+47 more)") {
		t.Fatalf("pagination must name what was dropped: %q", paged)
	}
	if resp := callTool(t, s,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"run_timeline","arguments":{"run_id":"r"}}}`); resp["error"] == nil {
		t.Fatal("a call past the duration limit must be refused")
	}
	s.Lim.MaxPageItems = 40 // response bytes now trip before pagination
	if resp := callTool(t, s,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"graphs_list"}}`); resp["error"] == nil {
		t.Fatal("oversized responses must be refused")
	}
}

// TestAuditRecordsToolNamesButNeverArguments: the trail is bounded
// and prompt-free.
func TestAuditRecordsToolNamesButNeverArguments(t *testing.T) {
	s := wiredServer()
	s.Lim = Limits{MaxAuditEntries: 2}
	secret := "dev/super-secret-prompt-about-decks"
	for i := range 3 {
		callTool(t, s, fmt.Sprintf(
			`{"jsonrpc":"2.0","id":%d,"method":"tools/call","params":{"name":"scenario_execute","arguments":{"scenario":"%s"}}}`,
			30+i, secret))
	}
	audit := s.Audit()
	if len(audit) != 2 {
		t.Fatalf("the audit trail must stay bounded: %d", len(audit))
	}
	for _, entry := range audit {
		if entry.Tool != "scenario_execute" || entry.Duration < 0 ||
			entry.RequestBytes == 0 {
			t.Fatalf("audit must carry tool, sizes and duration: %+v", entry)
		}
		if strings.Contains(fmt.Sprintf("%+v", entry), "super-secret") {
			t.Fatalf("audit must never record arguments: %+v", entry)
		}
	}
}
