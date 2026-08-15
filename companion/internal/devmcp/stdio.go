package devmcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// request is the JSON-RPC 2.0 envelope the dev MCP accepts.
type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ServeStdio speaks line-delimited JSON-RPC 2.0 over the given
// streams (OpenSpec add-graph-workflow-harness, task 7.1). The tool
// surface stays empty until task 7.3 mounts the read-only inspection
// tools; a malformed line answers an error and never kills the loop.
func ServeStdio(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	encoder := json.NewEncoder(w)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			_ = encoder.Encode(response{JSONRPC: "2.0",
				Error: &rpcError{Code: -32700, Message: "parse error"}})
			continue
		}
		_ = encoder.Encode(handle(req))
	}
	return scanner.Err()
}

func handle(req request) response {
	out := response{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		out.Result = map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name": "mtga-dev-mcp", "version": "0.1.0"},
			"capabilities": map[string]any{"tools": map[string]any{}},
		}
	case "tools/list":
		out.Result = map[string]any{"tools": []any{}}
	default:
		out.Error = &rpcError{Code: -32601,
			Message: fmt.Sprintf("method %q not found", req.Method)}
	}
	return out
}
