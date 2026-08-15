## ADDED Requirements

### Requirement: Structured deck workspace
The system SHALL represent a deck with stable card identifiers, quantities, zones, selected format, and catalog version rather than relying only on unparsed text.

#### Scenario: Arena deck text is imported
- **WHEN** the user pastes a valid Arena-format deck list
- **THEN** the system parses it into main deck and sideboard entries while preserving diagnostics for ambiguous or unknown cards

### Requirement: Ownership validation
The system SHALL compare a deck against a selected collection snapshot and report owned, missing, and wildcard-relevant quantities per card.

#### Scenario: Deck exceeds owned copies
- **WHEN** a deck requires more copies of a card than the selected snapshot contains
- **THEN** the workspace reports required, owned, and missing quantities without modifying the deck

### Requirement: Versioned format validation
The system SHALL validate deck constraints against a named ruleset and catalog version and SHALL disclose when legality data is stale or unavailable.

#### Scenario: Deck violates a known format constraint
- **WHEN** a deck is checked against a supported format and violates deck size, copy, sideboard, or card-legality rules
- **THEN** the workspace returns stable violations tied to the relevant entries and ruleset version

#### Scenario: Legality data is stale
- **WHEN** the selected catalog exceeds its freshness policy
- **THEN** the workspace labels legality as stale and does not present it as current

### Requirement: Evidence-based substitutions
The system SHALL generate substitution candidates from deterministic constraints and SHALL distinguish candidate generation from model explanation.

#### Scenario: User requests an owned replacement
- **WHEN** a missing card has candidate replacements in the owned collection under the selected color, mana, type, and format constraints
- **THEN** the workspace returns ranked candidates with the facts used for ranking

### Requirement: Explicit Arena export
The system SHALL generate Arena-format deck text and SHALL require explicit user action or agent approval before writing a file or clipboard.

#### Scenario: User copies deck from the workspace
- **WHEN** the user directly activates the Copy for Arena command
- **THEN** the system writes the previewed deck text to the clipboard and confirms completion

#### Scenario: Assistant proposes deck export
- **WHEN** the assistant requests the same clipboard operation
- **THEN** the operation remains pending until the user approves the exact preview

### Requirement: Saved deck history
The system SHALL retain local deck revisions with source and validation context.

#### Scenario: User saves a revised deck
- **WHEN** a saved deck has changed entries or format settings
- **THEN** the system creates a new revision linked to the prior revision and records the collection snapshot used for validation

