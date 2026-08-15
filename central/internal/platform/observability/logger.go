// Package observability owns structured logging, W3C trace context
// and low-cardinality metrics (OpenSpec add-central-go-platform, task
// 7.1). Privacy is enforced at the handler and registry level, never
// by caller discipline.
package observability

import (
	"io"
	"log/slog"
	"strings"
)

// redactedKeys never reach the log output with their values.
var redactedKeys = []string{"token", "secret", "password", "dsn",
	"authorization", "credential"}

// NewLogger returns a JSON slog logger whose handler redacts any
// attribute with a sensitive key, at every nesting level.
func NewLogger(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			lower := strings.ToLower(attr.Key)
			for _, banned := range redactedKeys {
				if strings.Contains(lower, banned) {
					return slog.String(attr.Key, "[redacted]")
				}
			}
			return attr
		},
	}))
}
