package httpapi

import (
	"context"
	"errors"
)

// TelemetryEventInput is one event handed to the store after the
// policy accepted the whole batch.
type TelemetryEventInput struct {
	Name  string
	Attrs map[string]string
}

// ErrDuplicateBatch marks an idempotent replay: already accepted,
// nothing new persisted.
var ErrDuplicateBatch = errors.New("httpapi: duplicate telemetry batch")

// TelemetryStore persists one validated batch scoped to the
// authenticated installation.
type TelemetryStore interface {
	IngestBatch(ctx context.Context, installationID, batchID string,
		sequence int64, events []TelemetryEventInput) error
}

type telemetryEvent struct {
	Name  string            `json:"name"`
	Attrs map[string]string `json:"attrs"`
}

type telemetryRequest struct {
	BatchID  string           `json:"batch_id"`
	Sequence int64            `json:"sequence"`
	Events   []telemetryEvent `json:"events"`
}

// EventPolicy is the closed vocabulary of accepted events. Validation
// is atomic: one bad event refuses the whole batch, so a batch is
// either fully accepted or fully rejected.
type EventPolicy struct {
	Names       map[string]bool
	MaxEvents   int
	MaxAttrs    int
	MaxValueLen int
}

func (p EventPolicy) validate(events []telemetryEvent) error {
	if len(events) == 0 || len(events) > p.MaxEvents {
		return errors.New("event count out of bounds")
	}
	for _, event := range events {
		if !p.Names[event.Name] {
			return errors.New("event name outside the allowlist")
		}
		if len(event.Attrs) > p.MaxAttrs {
			return errors.New("too many attributes")
		}
		for key, value := range event.Attrs {
			if len(key) > p.MaxValueLen || len(value) > p.MaxValueLen {
				return errors.New("oversized attribute")
			}
		}
	}
	return nil
}
