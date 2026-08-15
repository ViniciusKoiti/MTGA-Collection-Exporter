package assistctx

import "strings"

// forbiddenFragments are scrubbed from every field that reaches the
// context: path separators and credential-shaped tokens must not survive
// even if a caller passes them by mistake — the builder is the last line
// of defense before data leaves the device.
var forbiddenFragments = []string{
	"\\", "/", "c:", "bearer ", "password", "token=", "secret", "apikey", "api_key",
}

// scrub replaces any forbidden fragment with a fixed marker, keeping the
// output deterministic for the snapshot test.
func scrub(value string) string {
	lower := strings.ToLower(value)
	for _, fragment := range forbiddenFragments {
		if strings.Contains(lower, fragment) {
			return "[redacted]"
		}
	}
	return value
}
