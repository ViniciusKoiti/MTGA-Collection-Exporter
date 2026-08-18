package observability

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// Span is a W3C trace-context span — OpenTelemetry-compatible IDs and
// header format without carrying the SDK.
type Span struct {
	TraceID string // 32 hex chars
	SpanID  string // 16 hex chars
}

func randomHex(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// NewSpan starts a fresh trace.
func NewSpan() (Span, error) {
	traceID, err := randomHex(16)
	if err != nil {
		return Span{}, err
	}
	spanID, err := randomHex(8)
	if err != nil {
		return Span{}, err
	}
	return Span{TraceID: traceID, SpanID: spanID}, nil
}

// Child keeps the trace and starts a new span under it.
func (s Span) Child() (Span, error) {
	spanID, err := randomHex(8)
	if err != nil {
		return Span{}, err
	}
	return Span{TraceID: s.TraceID, SpanID: spanID}, nil
}

// Traceparent renders the W3C header value (version 00, sampled).
func (s Span) Traceparent() string {
	return fmt.Sprintf("00-%s-%s-01", s.TraceID, s.SpanID)
}

// ParseTraceparent accepts a W3C traceparent header or refuses it.
func ParseTraceparent(header string) (Span, error) {
	parts := strings.Split(header, "-")
	if len(parts) != 4 || parts[0] != "00" ||
		len(parts[1]) != 32 || len(parts[2]) != 16 {
		return Span{}, fmt.Errorf("observability: malformed traceparent %q", header)
	}
	for _, hexPart := range parts[1:3] {
		if _, err := hex.DecodeString(hexPart); err != nil {
			return Span{}, fmt.Errorf("observability: malformed traceparent %q", header)
		}
	}
	return Span{TraceID: parts[1], SpanID: parts[2]}, nil
}
