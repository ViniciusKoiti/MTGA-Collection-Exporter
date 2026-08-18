## ADDED Requirements

### Requirement: Buildability facts are complete
The matcher SHALL return required, allocated, missing, unresolved, sideboard, legality,
rarity, snapshot, ruleset, and catalog evidence for every evaluated deck.

#### Scenario: Deck is fully owned
- **WHEN** every required copy has a legal allocation and the deck is legal
- **THEN** the deck is classified `ready` with zero missing and unresolved copies

### Requirement: Recommendation classes are policy driven
The system SHALL classify `ready`, `nearly-buildable`, `within-budget`, and
`not-buildable` using explicit versioned thresholds and user-selected rarity budgets.

#### Scenario: Deck fits the selected budget
- **WHEN** its missing rarity vector is within every configured budget
- **THEN** it is classified `within-budget` and returns the exact required vector

#### Scenario: Wildcard balance is unknown
- **WHEN** no trusted wildcard balance is available
- **THEN** the system reports required wildcards without claiming the user can afford them

### Requirement: Ranking is deterministic and explainable
Ranking SHALL use the documented fact-vector order and a stable deck identity
tie-breaker, and SHALL expose all values used in ordering.

#### Scenario: Parallel matching completes in a different order
- **WHEN** workers return deck partitions in nondeterministic completion order
- **THEN** the final ranked result remains byte-semantically equivalent

### Requirement: Stale and partial evidence reduces confidence
The matcher SHALL disclose stale catalogs, partial decks, unresolved cards, unsupported
rulesets, and weak provider evidence and SHALL NOT rank them as fully current facts.

#### Scenario: Catalog is stale
- **WHEN** ranking uses an expired but compatible cached snapshot
- **THEN** every result is marked stale with snapshot time and recovery guidance

### Requirement: Interactive matching has a performance gate
The local matcher SHALL evaluate 10,000 normalized decks against a 30,000-printing
catalog within the documented reference-machine budget and bounded worker limit.

#### Scenario: Performance fixture runs
- **WHEN** the standard large fixture executes on the reference Windows machine
- **THEN** p95 is at most one second and peak process memory is at most 250 MB

### Requirement: Model explanation cannot change facts
Optional explanation SHALL receive a bounded evidence projection and SHALL NOT modify
classification, allocation, legality, missing counts, confidence, or ranking.

#### Scenario: Explanation contradicts evidence
- **WHEN** model text calls a deck ready but missing copies are nonzero
- **THEN** the UI keeps the deterministic class and labels the conflicting statement unsupported
