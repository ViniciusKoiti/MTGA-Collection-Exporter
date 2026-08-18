package consent

import (
	"errors"
	"regexp"
	"strings"
)

// Content screening (OpenSpec add-central-go-platform, task 5.3):
// even a declared string attribute must not smuggle raw logs, deck
// lists, filesystem paths, prompts, credentials or MTGA identities.
// Screening is shape-based and runs after the type checks, and one
// bad value refuses the whole batch upstream — atomically.
var (
	errMultiline  = errors.New("multi-line content (raw log or deck list)")
	errPath       = errors.New("filesystem path")
	errCredential = errors.New("credential-shaped value")
	errIdentifier = errors.New("identifier-shaped value")
	errFreeText   = errors.New("free-text prose (prompt-shaped)")

	pathShape   = regexp.MustCompile(`(?i)^[a-z]:[\\/]|[\\/][\w.-]+[\\/]|\.(log|json|mtga|txt)$`)
	hexIdentity = regexp.MustCompile(`[0-9a-fA-F]{16,}`)
	numIdentity = regexp.MustCompile(`\d{12,}`)
)

// Screen refuses values whose shape betrays disallowed content.
func Screen(value string) error {
	if strings.ContainsAny(value, "\n\r") {
		return errMultiline
	}
	if strings.Contains(value, `\`) || pathShape.MatchString(value) {
		return errPath
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "bearer ") || strings.Contains(value, "eyJ") ||
		strings.Contains(lower, "password=") || strings.Contains(lower, "token=") {
		return errCredential
	}
	if hexIdentity.MatchString(value) {
		return errIdentifier
	}
	if numIdentity.MatchString(value) {
		return errIdentifier
	}
	if strings.Count(value, " ") > 6 {
		return errFreeText
	}
	return nil
}
