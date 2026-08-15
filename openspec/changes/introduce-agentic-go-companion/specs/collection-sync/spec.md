## ADDED Requirements

### Requirement: User-controlled collection sources
The system SHALL synchronize collections only from sources enabled by the user and SHALL prefer supported MTG Arena Detailed Logs or explicit file import over legacy process-memory scanning.

#### Scenario: Supported Detailed Logs are detected
- **WHEN** the user enables log synchronization and the system recognizes a supported collection payload
- **THEN** the system watches the log without requiring administrator privileges or process-memory access

#### Scenario: Existing export is imported
- **WHEN** the user selects a valid legacy `mtga_collection.json`
- **THEN** the system imports it as a collection observation and identifies the source as `legacy-json`

#### Scenario: Legacy scanner is requested
- **WHEN** the user explicitly selects the legacy scanner compatibility action
- **THEN** the system explains its compatibility status and runs it outside the agent tool surface

### Requirement: Immutable collection snapshots
The system SHALL normalize each accepted observation into an immutable, versioned collection snapshot with source provenance and timestamps.

#### Scenario: New observation is accepted
- **WHEN** a source produces a valid observation not previously committed
- **THEN** the system transactionally stores a new snapshot with source, observed time, imported time, schema version, quantities, and diagnostics

#### Scenario: Duplicate observation is received
- **WHEN** a source repeats an observation with the same source identity and content
- **THEN** the system treats it as idempotent and does not create a duplicate snapshot

### Requirement: Unknown card preservation
The system SHALL preserve source records that cannot be resolved by the card catalog and SHALL expose them as diagnostics.

#### Scenario: Arena identifier is unknown
- **WHEN** an observation contains a card identifier absent from the active catalog
- **THEN** the resulting snapshot retains the identifier and quantity as unresolved instead of silently discarding it

### Requirement: Sync health and recovery
The system SHALL report sync state, freshness, confidence, and stable recovery guidance without exposing raw implementation exceptions to users.

#### Scenario: Log payload becomes unsupported
- **WHEN** the watcher observes data but cannot recognize a supported collection payload
- **THEN** the system enters an error or stale state with a stable code and offers explicit import as a recovery path

#### Scenario: Last snapshot becomes stale
- **WHEN** no valid observation arrives within the configured freshness threshold
- **THEN** the system retains the last valid snapshot and marks it stale rather than replacing it with an empty collection

### Requirement: Compatibility collection export
The system SHALL project the latest valid snapshot into the current `mtga_collection.json` contract during the migration period.

#### Scenario: Compatibility export succeeds
- **WHEN** the user or configured export policy requests the legacy JSON projection
- **THEN** the system writes the supported fields atomically and existing read-only MCP consumers can parse the result

### Requirement: No MTGA control
The collection synchronization capability MUST NOT write MTGA memory, modify game files, synthesize input, intercept credentials, or automate gameplay.

#### Scenario: A caller requests a prohibited integration
- **WHEN** any UI, agent, or external client requests an MTGA control operation
- **THEN** the system rejects it with a policy error and records the rejected operation without executing it

