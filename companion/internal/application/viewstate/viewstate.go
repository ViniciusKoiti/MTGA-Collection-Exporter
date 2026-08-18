// Package viewstate defines the closed vocabulary of UI states every
// desktop view renders (OpenSpec introduce-agentic-go-companion, task
// 4.5): empty, loading, stale, partial, error and success. Errors are
// carried as stable apperr codes — the frontend never sees raw causes.
package viewstate

import (
	"errors"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
)

// Status is one of the six render states; the set is closed.
type Status string

// The closed status vocabulary.
const (
	StatusEmpty   Status = "empty"
	StatusLoading Status = "loading"
	StatusStale   Status = "stale"
	StatusPartial Status = "partial"
	StatusError   Status = "error"
	StatusSuccess Status = "success"
)

// State is what a view renders. Code is set only on error and partial
// states and is ALWAYS a stable application code; Detail is the
// presentable form — never a raw cause, path or driver message.
type State struct {
	Status Status      `json:"status"`
	Code   apperr.Code `json:"code,omitempty"`
	Detail string      `json:"detail,omitempty"`
}

// Empty is the state of a view with nothing to show yet.
func Empty() State { return State{Status: StatusEmpty} }

// Loading is the state while data is being fetched.
func Loading() State { return State{Status: StatusLoading} }

// Success is the state of fresh, complete data.
func Success() State { return State{Status: StatusSuccess} }

// Stale marks data served from an old snapshot, with a presentable
// note about why.
func Stale(detail string) State {
	return State{Status: StatusStale, Detail: detail}
}

// Partial marks incomplete data with the stable code explaining what
// was missing.
func Partial(code apperr.Code, detail string) State {
	return State{Status: StatusPartial, Code: code, Detail: detail}
}

// FromError maps any error to the error state via its stable code;
// untyped errors are internal by definition, and the raw cause never
// reaches Detail.
func FromError(err error) State {
	if err == nil {
		return Success()
	}
	var typed *apperr.Error
	if errors.As(err, &typed) {
		return State{Status: StatusError, Code: typed.Code,
			Detail: typed.Apresentavel()}
	}
	return State{Status: StatusError, Code: apperr.CodeInternal,
		Detail: string(apperr.CodeInternal)}
}

// Valid enforces the contract: a known status, codes exactly where
// they belong, and no code-less error states.
func (s State) Valid() error {
	switch s.Status {
	case StatusEmpty, StatusLoading, StatusSuccess:
		if s.Code != "" {
			return fmt.Errorf("viewstate: %s must not carry a code", s.Status)
		}
	case StatusStale:
	case StatusPartial, StatusError:
		if s.Code == "" {
			return fmt.Errorf("viewstate: %s requires a stable code", s.Status)
		}
	default:
		return fmt.Errorf("viewstate: unknown status %q", s.Status)
	}
	return nil
}
