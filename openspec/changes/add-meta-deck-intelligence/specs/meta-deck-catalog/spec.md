## ADDED Requirements

### Requirement: Only approved providers are ingested
The catalog worker SHALL ingest a provider only when its adapter, usage rights,
formats, rate policy, and provenance fields are approved in configuration.

#### Scenario: Unapproved provider is configured
- **WHEN** a worker receives a provider without an active approval record
- **THEN** it refuses ingestion and publishes no observation from that provider

### Requirement: Observations preserve provenance
Every meta deck observation SHALL include provider identity, external deck identity,
format, play mode, observed period, context, fetch time, source reference, and content hash.

#### Scenario: Required provenance is missing
- **WHEN** an adapter returns a deck without a required provenance field
- **THEN** the observation is quarantined with a stable validation code

### Requirement: Catalog snapshots are immutable
The platform SHALL publish normalized meta decks in immutable, schema-versioned,
content-addressed snapshots with creation time, source set, and quality diagnostics.

#### Scenario: Provider corrects a deck
- **WHEN** normalized content changes after a snapshot was published
- **THEN** the platform publishes a new snapshot and leaves the old artifact unchanged

### Requirement: Deck observations are validated
The normalizer SHALL validate zones, quantities, format, mode, card identities,
duplicate entries, sample metadata, and bounded deck size before publication.

#### Scenario: Deck contains an unresolved card
- **WHEN** a card cannot be resolved against the pinned card catalog
- **THEN** the deck remains quarantined or explicitly partial and is not ranked as complete

### Requirement: Freshness and compatibility are explicit
Every manifest SHALL declare catalog time, source window, schema compatibility,
supported rulesets, freshness policy, object hash, and compressed size.

#### Scenario: Client is offline after catalog expiry
- **WHEN** the active snapshot exceeds freshness policy and the platform is unavailable
- **THEN** the client retains it, marks results stale, and does not claim current meta status

### Requirement: Conflicting observations remain explainable
The catalog SHALL retain provider-level evidence when equivalent archetypes or deck
identities are merged and SHALL use deterministic merge rules.

#### Scenario: Two providers disagree on a deck list
- **WHEN** observations map to one archetype but have different card quantities
- **THEN** both variants remain addressable and the catalog records the merge evidence
