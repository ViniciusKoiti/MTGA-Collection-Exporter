package devmcp

import "fmt"

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
		result, rpcErr := s.guardedCall(req.Params)
		out.Result, out.Error = result, rpcErr
	default:
		out.Error = &rpcError{Code: -32601,
			Message: fmt.Sprintf("method %q not found", req.Method)}
	}
	return out
}
