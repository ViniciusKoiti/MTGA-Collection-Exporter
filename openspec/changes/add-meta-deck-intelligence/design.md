## Context

The companion currently plans card metadata, ownership checks, legality, substitutions,
and saved decks. It lacks an approved source model for meta decks and a canonical way
to satisfy one deck requirement from several owned legal printings. Recommendations
must work from immutable central catalog snapshots, remain useful offline, and expose
facts independently from optional model explanations.

## Goals / Non-Goals

**Goals:**

- Publish trustworthy, versioned meta observations with provenance and freshness.
- Resolve exact printings into a legal playable identity without losing evidence.
- Rank decks deterministically by buildability and user-selected wildcard budgets.
- Keep matching fast enough for interactive filtering over a large catalog.

**Non-Goals:**

- Scraping providers without explicit approval or documented usage rights.
- Predicting win rate beyond the evidence supplied by an approved provider.
- Inferring wildcard balances from memory, account APIs, or synthetic input.
- Building a full Magic rules engine or allowing an LLM to decide factual ranking.

## Decisions

### 1. Treat provider data as observations, not truth

Each adapter emits `MetaDeckObservation` with provider, external ID, format, queue,
BO1/BO3 mode, event or ladder context, observed window, sample metadata, source URL,
license status, fetched time, and raw content hash. The worker validates and normalizes
observations before publishing a catalog; unsupported or unlicensed data is quarantined.

Catalog snapshots are immutable and addressed by schema version plus content hash.
Corrections create a new snapshot. Clients retain the last valid compatible snapshot
and expose stale or partial status instead of silently replacing it.

### 2. Preserve printing identity and add a playable identity

Inventory remains keyed by exact Arena printing. A catalog maps that printing to a
stable playable identity, normally Oracle identity plus face semantics, and to ruleset
legality. Deck requirements use playable identity but may retain a preferred printing.
Owned copies satisfy a requirement only when their printing is legal for the selected
ruleset. Allocation order is owned preference, preferred printing, then stable ID.

Unresolved identity never counts as owned. Basic-land supply is a versioned ruleset
policy rather than an implicit infinite quantity. Export chooses an allocated owned
printing first and a catalog default only for missing copies.

### 3. Rank from a fact vector, not one opaque score

For each deck compute legality, required and allocated copies, missing copies by card
and rarity, unresolved requirements, sideboard gaps, and catalog confidence. Classify
as `ready`, `nearly-buildable`, `within-budget`, or `not-buildable`. Nearly-buildable
and rarity budgets are explicit user settings with versioned defaults.

Ordering is lexicographic: valid/fresh evidence, class, unresolved count, missing
mythic, rare, uncommon, common, total missing, then stable deck ID. The UI may apply
user-selected rarity weights but must display the underlying vector. A wildcard plan
means required wildcards, not proof of the user's available balance.

### 4. Match locally with bounded parallelism

The companion downloads a compact normalized snapshot and builds immutable collection
and deck indexes. Matching partitions decks across a bounded worker group, carries a
cancellable context, and performs a deterministic stable merge. It never starts one
unbounded goroutine per deck. Cached results key on collection, catalog, ruleset, and
ranking-policy versions.

### 5. Keep explanation downstream from facts

Deterministic services return `RecommendationEvidence`. The UI can render it directly.
An optional model receives only selected deck facts and may explain trade-offs; its
text cannot alter class, counts, legality, order, or confidence.

## Risks / Trade-offs

- **Provider licensing or schema changes** -> Require an approval registry, adapters,
  quarantines, fixture contracts, and rapid provider disablement.
- **Oracle identity is not sufficient for every card face** -> Include face semantics
  and ruleset-tested exceptional mappings rather than name-only matching.
- **Rarity weighting can look objective** -> Show the fact vector and versioned policy.
- **Large catalogs can increase local memory** -> Publish compact projections, index
  once per snapshot, benchmark, and paginate presentation results.
- **Concurrent matching can become nondeterministic** -> Workers return immutable
  partitions and one stable reducer owns ordering.

## Migration Plan

1. Approve one fixture-backed provider and define observation and snapshot schemas.
2. Add playable identity mappings and golden reprint-allocation tests.
3. Implement buildability for user-supplied decks before enabling meta ranking.
4. Publish one Standard BO1 snapshot and validate it against known deck fixtures.
5. Add BO3, additional formats, bounded parallel matching, and cached indexes.
6. Roll back by pinning clients to the last valid snapshot and disabling the provider.

## Open Questions

- Which provider is legally and technically suitable for the first catalog?
- What default thresholds define `nearly-buildable` for each rarity?
- Which formats follow Standard BO1 and BO3?
