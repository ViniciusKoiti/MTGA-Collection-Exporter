package devmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Limits bound every dimension of a dev MCP session (task 7.5). The
// audit trail records tool names, sizes and durations ONLY — never
// arguments — so no prompt can ever be recorded.
type Limits struct {
	MaxRequestBytes  int
	MaxResponseBytes int
	MaxCallDuration  time.Duration
	MaxConcurrent    int
	MaxPageItems     int
	MaxAuditEntries  int
}

func (l Limits) withDefaults() Limits {
	pick := func(value, fallback int) int {
		if value > 0 {
			return value
		}
		return fallback
	}
	l.MaxRequestBytes = pick(l.MaxRequestBytes, 64*1024)
	l.MaxResponseBytes = pick(l.MaxResponseBytes, 256*1024)
	l.MaxConcurrent = pick(l.MaxConcurrent, 1)
	l.MaxPageItems = pick(l.MaxPageItems, 100)
	l.MaxAuditEntries = pick(l.MaxAuditEntries, 200)
	if l.MaxCallDuration <= 0 {
		l.MaxCallDuration = 5 * time.Second
	}
	return l
}

// page truncates a listing to the pagination budget, naming what was
// dropped instead of hiding it.
func (l Limits) page(lines []string) string {
	if len(lines) > l.MaxPageItems {
		dropped := len(lines) - l.MaxPageItems
		lines = append(lines[:l.MaxPageItems], fmt.Sprintf("(+%d more)", dropped))
	}
	return strings.Join(lines, "\n")
}

// guardedCall enforces concurrency, duration, response and audit
// limits around one tool call.
func (s *Server) guardedCall(raw json.RawMessage) (any, *rpcError) {
	limits := s.Lim.withDefaults()
	s.mu.Lock()
	if s.inFlight >= limits.MaxConcurrent {
		s.mu.Unlock()
		return nil, &rpcError{Code: -32002, Message: "concurrency limit reached"}
	}
	s.inFlight++
	s.mu.Unlock()
	defer func() { s.mu.Lock(); s.inFlight--; s.mu.Unlock() }()

	ctx, cancel := context.WithTimeout(context.Background(), limits.MaxCallDuration)
	defer cancel()
	start := time.Now()
	result, rpcErr := s.call(ctx, raw)
	if rpcErr == nil && ctx.Err() != nil {
		result, rpcErr = nil, &rpcError{Code: -32001,
			Message: "call exceeded the duration limit"}
	}
	responseBytes := 0
	if encoded, err := json.Marshal(result); err == nil {
		responseBytes = len(encoded)
	}
	if rpcErr == nil && responseBytes > limits.MaxResponseBytes {
		result, rpcErr = nil, &rpcError{Code: -32001,
			Message: "response exceeds the size limit"}
	}
	var name struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal(raw, &name)
	s.record(limits, AuditEntry{Tool: name.Name, RequestBytes: len(raw),
		ResponseBytes: responseBytes, Duration: time.Since(start),
		Refused: rpcErr != nil})
	return result, rpcErr
}
