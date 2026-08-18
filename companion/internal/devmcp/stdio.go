package devmcp

import (
	"bufio"
	"encoding/json"
	"io"
	"sync"
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

// Server is the dev MCP with its closed five-tool surface; without
// explicit Capabilities it is read-only, and Limits bound every call.
type Server struct {
	Deps ToolDeps
	Caps Capabilities
	Lim  Limits

	mu       sync.Mutex
	inFlight int
	audit    []AuditEntry
}

// ServeStdio runs a dependency-free server: same closed surface, tool
// sources not attached (each call answers a clean error).
func ServeStdio(r io.Reader, w io.Writer) error {
	return (&Server{}).Serve(r, w)
}

// Serve speaks line-delimited JSON-RPC 2.0 over the given streams
// (task 7.1); a malformed or oversized line answers an error and
// never kills the loop.
func (s *Server) Serve(r io.Reader, w io.Writer) error {
	limits := s.Lim.withDefaults()
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	encoder := json.NewEncoder(w)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if len(line) > limits.MaxRequestBytes {
			s.record(limits, AuditEntry{RequestBytes: len(line), Refused: true})
			_ = encoder.Encode(response{JSONRPC: "2.0",
				Error: &rpcError{Code: -32600, Message: "request too large"}})
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
