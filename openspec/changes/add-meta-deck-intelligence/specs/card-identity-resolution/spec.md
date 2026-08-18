## ADDED Requirements

### Requirement: Exact and playable identities are preserved
Every owned card SHALL retain its exact printing identity and, when resolved, a
separate playable identity used for deck requirement allocation.

#### Scenario: Two legal reprints are owned
- **WHEN** both printings resolve to the same playable identity
- **THEN** their legal quantities can jointly satisfy one deck requirement without losing printing evidence

### Requirement: Legality constrains equivalence
The allocator SHALL count an equivalent printing only when it is legal under the
selected versioned ruleset and permitted in the requested deck zone.

#### Scenario: Owned reprint is not legal
- **WHEN** an owned equivalent printing is illegal for the selected format
- **THEN** it does not satisfy the requirement and appears in allocation diagnostics

### Requirement: Allocation is deterministic
The allocator SHALL use a stable, documented order for owned preference, requested
printing, legal alternatives, and stable identity tie-breaking.

#### Scenario: Allocation is repeated
- **WHEN** the same deck, collection, catalog, and ruleset versions are evaluated twice
- **THEN** the exact allocated printings and gaps are identical

### Requirement: Unresolved identities are conservative
Unresolved or ambiguous source records MUST NOT be counted as satisfying a deck
requirement and SHALL remain visible as diagnostics.

#### Scenario: Name matches but identity is ambiguous
- **WHEN** multiple playable identities share a normalized display name
- **THEN** resolution remains unresolved instead of selecting by name alone

### Requirement: Special card structures are versioned
The catalog or ruleset SHALL define face semantics, rebalanced variants, basic-land
supply, and exceptional equivalence rules explicitly with fixture coverage.

#### Scenario: Basic land policy applies
- **WHEN** a ruleset treats an eligible basic land as freely available
- **THEN** allocation cites that policy rather than inventing owned copies

### Requirement: Export identity follows allocation
Arena export SHALL prefer an allocated owned legal printing and SHALL distinguish
missing copies that require a catalog-default printing.

#### Scenario: Requirement is partly owned
- **WHEN** two of four copies are allocated from owned printings
- **THEN** export evidence identifies the owned choices and the two missing copies separately
