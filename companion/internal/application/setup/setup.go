// Package setup drives the first-run experience (OpenSpec
// introduce-agentic-go-companion, task 4.2): detect the collection
// source, flag Detailed Logs guidance when they cannot serve, capture
// EXPLICIT privacy choices, verify the chosen source and fall back to
// the explicit JSON import — never silently.
package setup

import (
	"context"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// PrivacyChoices are explicit decisions; nothing defaults to sharing.
type PrivacyChoices struct {
	TelemetryOptIn   bool `json:"telemetry_opt_in"`
	AssistantEnabled bool `json:"assistant_enabled"`
}

// Probe decides which source the installed client can support.
type Probe interface {
	Decide(ctx context.Context) (collection.SourceKind, error)
}

// Verifier proves the chosen source is actually operable.
type Verifier interface {
	Verify(ctx context.Context, source collection.SourceKind) error
}

// Result is everything the first-run screen renders.
type Result struct {
	Source        collection.SourceKind `json:"source"`
	NeedsGuidance bool                  `json:"needs_guidance"`
	Verified      bool                  `json:"verified"`
	FellBack      bool                  `json:"fell_back"`
	Privacy       PrivacyChoices        `json:"privacy"`
}

// Flow wires the ports; the desktop injects the real adapters.
type Flow struct {
	Probe  Probe
	Verify Verifier
}

// Run executes first-run. A probe failure or an unverifiable source
// falls back to the explicit import with FellBack set, and guidance
// stays on whenever Detailed Logs are not the working source.
func (f Flow) Run(ctx context.Context, privacy PrivacyChoices) Result {
	source, err := f.Probe.Decide(ctx)
	if err != nil || source == "" {
		source = collection.SourceJSONImport
	}
	result := Result{Source: source, Privacy: privacy,
		NeedsGuidance: source != collection.SourceDetailedLogs}
	if err := f.Verify.Verify(ctx, source); err != nil {
		if source == collection.SourceJSONImport {
			return result
		}
		result.FellBack = true
		result.Source = collection.SourceJSONImport
		result.NeedsGuidance = true
		result.Verified = f.Verify.Verify(ctx,
			collection.SourceJSONImport) == nil
		return result
	}
	result.Verified = true
	return result
}
