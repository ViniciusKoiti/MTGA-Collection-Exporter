// Package flags implements the independent feature switches of the
// staged rollout (OpenSpec add-central-go-platform, task 8.7):
// catalog publication enables first, telemetry enrollment separately,
// and each rolls back with ONE flip — no deploy, no restart.
package flags

import "sync/atomic"

// Flag is one independently flippable feature switch.
type Flag struct {
	name  string
	state atomic.Bool
}

// New creates a named flag; features ship dark unless enabled.
func New(name string, initial bool) *Flag {
	f := &Flag{name: name}
	f.state.Store(initial)
	return f
}

// Name identifies the flag in diagnostics.
func (f *Flag) Name() string { return f.name }

// Enabled answers the current state.
func (f *Flag) Enabled() bool { return f.state.Load() }

// Enable turns the feature on.
func (f *Flag) Enable() { f.state.Store(true) }

// Disable is the rollback: one flip, effective immediately.
func (f *Flag) Disable() { f.state.Store(false) }

// Set holds the two independently rolled-out features.
type Set struct {
	CatalogPublication  *Flag
	TelemetryEnrollment *Flag
}

// FromEnv reads the flags from the environment; anything that is not
// exactly "true" means dark — features never enable by accident.
func FromEnv(getenv func(string) string) Set {
	return Set{
		CatalogPublication: New("catalog_publication",
			getenv("CENTRAL_FLAG_CATALOG_PUBLICATION") == "true"),
		TelemetryEnrollment: New("telemetry_enrollment",
			getenv("CENTRAL_FLAG_TELEMETRY_ENROLLMENT") == "true"),
	}
}
