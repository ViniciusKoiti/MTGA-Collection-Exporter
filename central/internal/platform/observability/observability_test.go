package observability

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// TestLoggerRedactsSensitiveKeysAtTheHandler: privacy holds even when
// a careless caller logs a credential.
func TestLoggerRedactsSensitiveKeysAtTheHandler(t *testing.T) {
	var out bytes.Buffer
	logger := NewLogger(&out, slog.LevelInfo)
	logger.Info("enrolled", "installation", "inst-1",
		"api_token", "raw-secret-value", "db_dsn", "postgres://u:p@h/db")
	line := out.String()
	if strings.Contains(line, "raw-secret-value") ||
		strings.Contains(line, "postgres://") {
		t.Fatalf("sensitive values leaked: %s", line)
	}
	if !strings.Contains(line, "[redacted]") ||
		!strings.Contains(line, "inst-1") {
		t.Fatalf("redaction must hit only sensitive keys: %s", line)
	}
}

func TestTraceparentRoundTripAndLineage(t *testing.T) {
	root, err := NewSpan()
	if err != nil {
		t.Fatalf("new span: %v", err)
	}
	child, err := root.Child()
	if err != nil {
		t.Fatalf("child: %v", err)
	}
	if child.TraceID != root.TraceID || child.SpanID == root.SpanID {
		t.Fatalf("child must keep the trace, not the span: %+v", child)
	}
	parsed, err := ParseTraceparent(child.Traceparent())
	if err != nil || parsed != child {
		t.Fatalf("roundtrip broke: %+v %v", parsed, err)
	}
	for _, bad := range []string{"", "00-short-span-01",
		"zz-00000000000000000000000000000000-0000000000000000-01"} {
		if _, err := ParseTraceparent(bad); err == nil {
			t.Fatalf("malformed traceparent must be refused: %q", bad)
		}
	}
}

func TestMetricsRefuseCardinalityAndIdentifiers(t *testing.T) {
	m := &Metrics{AllowedLabels: map[string]bool{"route": true,
		"status": true}, MaxSeries: 2}
	if err := m.Inc("http_requests", map[string]string{"route": "/v1/manifest",
		"status": "200"}); err != nil {
		t.Fatalf("allowed series: %v", err)
	}
	if err := m.Inc("http_requests", map[string]string{
		"installation": "inst-1"}); err == nil {
		t.Fatal("labels outside the allowlist must be refused")
	}
	if err := m.Inc("http_requests", map[string]string{
		"route": "0123456789abcdef0123"}); err == nil {
		t.Fatal("identifier-shaped values must be refused")
	}
	if err := m.Inc("http_requests", map[string]string{"route": "/v1/x",
		"status": "500"}); err != nil {
		t.Fatalf("second series within budget: %v", err)
	}
	if err := m.Inc("http_requests", map[string]string{"route": "/v1/y",
		"status": "500"}); err == nil {
		t.Fatal("the series budget must be hard")
	}
	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("refusals must not grow the registry: %v", snap)
	}
}
