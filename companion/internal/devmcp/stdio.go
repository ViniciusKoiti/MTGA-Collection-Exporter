package devmcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// request is the JSON-RPC 2.0 envelope the dev MCP accepts.
type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
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

// Server is the dev MCP with its closed five-tool surface.
type Server struct {
	Deps ToolDeps
}

// ServeStdio runs a dependency-free server: same closed surface, tool
// sources not attached (each call answers a clean error).
func ServeStdio(r io.Reader, w io.Writer) error {
	return (&Server{}).Serve(r, w)
}

// Serve speaks line-delimited JSON-RPC 2.0 over the given streams
// (task 7.1); a malformed line answers an error, never kills the loop.
func (s *Server) Serve(r io.Reader, w io.Writer) error {
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
		_ = encoder.Encode(s.handle(req))
	}
	return scanner.Err()
}

func (s *Server) handle(req request) response {
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
		descriptors := make([]map[string]string, 0, len(toolNames()))
		for _, name := range toolNames() {
			descriptors = append(descriptors, map[string]string{"name": name})
		}
		out.Result = map[string]any{"tools": descriptors}
	case "tools/call":
		result, rpcErr := s.call(context.Background(), req.Params)
		out.Result, out.Error = result, rpcErr
	default:
		out.Error = &rpcError{Code: -32601,
			Message: fmt.Sprintf("method %q not found", req.Method)}
	}
	return out
}
