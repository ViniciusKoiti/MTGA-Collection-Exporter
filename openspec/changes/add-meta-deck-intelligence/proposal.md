## Why

The companion can validate a supplied deck, but it cannot yet obtain trustworthy
meta decks or determine which current archetypes a user can build efficiently.
Versioned meta data, canonical card identity, and deterministic ranking are needed
before recommendations can be accurate, explainable, and testable.

## What Changes

- Ingest deck observations only from approved providers with provenance, format,
  BO1/BO3 mode, observed period, sample context, and licensing status.
- Normalize equivalent legal printings without losing the exact printing needed
  for collection evidence or Arena-format export.
- Publish immutable meta catalog snapshots with freshness and quality diagnostics.
- Compare each deck against a selected collection snapshot and allocate owned legal
  printings deterministically before counting missing cards.
- Rank ready, nearly buildable, and wildcard-cost profiles with explicit facts and
  user-selected rarity budgets; do not infer the user's wildcard balance.
- Keep model-generated explanations optional and subordinate to deterministic facts.

## Capabilities

### New Capabilities

- `meta-deck-catalog`: Acquires, validates, normalizes, versions, and exposes approved
  meta deck observations with provenance and freshness.
- `card-identity-resolution`: Resolves deck requirements across legal equivalent
  printings while preserving printing-level collection and export identities.
- `deck-buildability-ranking`: Computes and ranks buildability, missing copies,
  wildcard-relevant cost, confidence, and deterministic evidence.

### Modified Capabilities

None. Related companion capabilities are still active changes, not archived specs.

## Impact

- Adds meta provider ports and catalog snapshot contracts to the central platform.
- Adds local cached meta snapshots and matching indexes to the companion.
- Extends deck recommendation graphs with catalog selection, allocation, ranking,
  stale-data handling, and evidence nodes.
- Requires approved-source governance and fixture licensing before a provider is enabled.
- Adds performance fixtures for thousands of decks and tens of thousands of printings.
