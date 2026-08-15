// Package ports declares the contracts the application core demands
// from adapters (providers, PostgreSQL, object storage, signing). Real
// adapters live in internal/adapters; tests use fakes.
package ports

import (
	"context"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
)

// ProviderFetcher fetches observations from an approved provider,
// honoring the adapter's rate/timeout budget and ctx cancellation.
type ProviderFetcher interface {
	Fetch(ctx context.Context, job catalog.ProviderJob) ([]catalog.Observation, error)
}

// Normalizer converts one raw observation into a validated card.
type Normalizer interface {
	Normalize(ctx context.Context, obs catalog.Observation) (catalog.Card, error)
}

// CardValidator decides whether a normalized card is sound enough to
// publish; refusal routes the card to quarantine, not to failure.
type CardValidator interface {
	Check(ctx context.Context, card catalog.Card) error
}

// QuarantineSink records a card refused by validation together with
// the reason, without failing the publication.
type QuarantineSink interface {
	Quarantine(ctx context.Context, card catalog.Card, reason string) error
}

// BatchWriter persists cards in bounded batches within the pgxpool budget.
type BatchWriter interface {
	WriteBatch(ctx context.Context, cards []catalog.Card) error
}

// ObjectWriter stores the canonical snapshot as an immutable object and
// returns the reference with the hash verified after upload.
type ObjectWriter interface {
	WriteSnapshot(ctx context.Context, snap catalog.Snapshot) (catalog.ObjectRef, error)
}

// Signer signs the manifest of the published object (Ed25519, key ID).
type Signer interface {
	Sign(ctx context.Context, ref catalog.ObjectRef) (catalog.Manifest, error)
}

// Activator makes the signed manifest the current snapshot in one
// transaction; on failure the previous manifest stays current.
type Activator interface {
	Activate(ctx context.Context, man catalog.Manifest) error
}
