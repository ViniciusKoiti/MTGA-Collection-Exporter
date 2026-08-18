// Package diaglog provides telemetry-free structured local logging and
// the redacted diagnostic-bundle exporter (OpenSpec
// introduce-agentic-go-companion, task 7.1). Logs are written only to
// the local writer handed in by the composition root — there is no
// network path here by construction.
package diaglog

import (
	"io"
	"log/slog"
	"strings"
)

// forbiddenFragments redact any attribute value that looks like a path
// or a credential before it reaches the local log file.
var forbiddenFragments = []string{
	"\\", "/", "c:", "bearer ", "password", "token=", "secret", "apikey", "api_key",
}

// scrub replaces sensitive-looking values with a fixed marker.
func scrub(value string) string {
	lower := strings.ToLower(value)
	for _, fragment := range forbiddenFragments {
		if strings.Contains(lower, fragment) {
			return "[redacted]"
		}
	}
	return value
}

// NewLocalLogger builds the structured logger used for sync and tool
// calls: every string attribute is scrubbed by the handler itself, so a
// careless call site cannot leak a path or credential into the log.
func NewLocalLogger(w io.Writer) *slog.Logger {
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Value.Kind() == slog.KindString {
				attr.Value = slog.StringValue(scrub(attr.Value.String()))
			}
			return attr
		},
	})
	return slog.New(handler)
}
