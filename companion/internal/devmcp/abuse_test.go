package devmcp

import (
	"bytes"
	"context"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
	"testing/iotest"
)

// Abuse battery (add-graph-workflow-harness, task 7.6). Oversized
// results and request bombs are already covered in limits_test.go.

func TestAbuseMalformedArgumentsAndForbiddenSurfaces(t *testing.T) {
	s := wiredServer()
	for name, body := range map[string]string{
		"params not an object": `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":[1,2]}`,
		"sql tool":             `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"sql_execute","arguments":{"scenario":"SELECT * FROM runs"}}}`,
		"shell tool":           `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"shell","arguments":{}}}`,
		"shell method":         `{"jsonrpc":"2.0","id":4,"method":"shell/execute"}`,
		"resource write":       `{"jsonrpc":"2.0","id":5,"method":"resources/write"}`,
	} {
		if resp := callTool(t, s, body); resp["error"] == nil {
			t.Fatalf("%s must be refused: %v", name, resp)
		}
	}
}

// TestAbuseInjectionStaysOpaqueData: SQL- and SSRF-shaped arguments
// reach the source as plain data — nothing interprets them.
func TestAbuseInjectionStaysOpaqueData(t *testing.T) {
	var received string
	s := wiredServer()
	s.Deps.ExecuteScenario = func(_ context.Context, name string) (string, error) {
		received = name
		return "ok", nil
	}
	payload := `dev/x'; DROP TABLE runs; -- http://169.254.169.254/meta`
	body := `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"scenario_execute","arguments":{"scenario":"` +
		payload + `"}}}`
	resp := callTool(t, s, body)
	if resp["error"] != nil {
		t.Fatalf("opaque data must not be rejected for its content: %v", resp)
	}
	if !strings.Contains(received, "DROP TABLE") ||
		!strings.Contains(received, "169.254.169.254") {
		t.Fatalf("the source must receive the argument verbatim: %q", received)
	}
}

// TestDevMCPImportsNoNetworkShellOrSQL is the structural SSRF/shell/
// SQL proof: the production package cannot fetch, exec or query.
func TestDevMCPImportsNoNetworkShellOrSQL(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	forbidden := []string{`"net/http"`, `"net"`, `"os/exec"`, `"database/sql"`}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), name, nil,
			parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imported := range file.Imports {
			for _, banned := range forbidden {
				if imported.Path.Value == banned {
					t.Errorf("%s imports %s: the dev MCP must not fetch, exec or query",
						name, banned)
				}
			}
		}
	}
}

func TestAbuseDisconnectsNeitherPanicNorHang(t *testing.T) {
	var out bytes.Buffer
	if err := (&Server{}).Serve(strings.NewReader(
		`{"jsonrpc":"2.0","id":1,"method":"tools/li`), &out); err != nil {
		t.Fatalf("an abrupt EOF must end cleanly: %v", err)
	}
	if err := (&Server{}).Serve(iotest.ErrReader(
		errors.New("connection reset")), &out); err == nil {
		t.Fatal("a transport error must surface to the caller")
	}
}
