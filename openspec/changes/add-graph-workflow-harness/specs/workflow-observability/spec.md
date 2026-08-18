## ADDED Requirements

### Requirement: Every transition emits a structured event
Each attempted and committed graph transition SHALL produce a versioned event with
run, step, graph, correlation, causation, time, duration, attempt, and outcome fields.

#### Scenario: Node completes successfully
- **WHEN** a node outcome and next state commit
- **THEN** the journal contains a correlated completion event with the pinned graph version

#### Scenario: Node fails before commit
- **WHEN** a node returns a typed error or reaches its deadline
- **THEN** the journal records the attempt and stable error code without claiming a committed transition

### Requirement: Evidence uses an allowlisted redacted schema
Events and diagnostics SHALL exclude raw logs, collection contents, credentials,
local paths, model prompts, tool payloads, and unbounded error strings.

#### Scenario: Forbidden field reaches the event sink
- **WHEN** a producer attempts to attach a forbidden or unknown attribute
- **THEN** the sink rejects or removes it and emits a bounded redaction diagnostic

#### Scenario: Diagnostic bundle is created
- **WHEN** a developer exports run evidence
- **THEN** it contains schemas, hashes, outcome codes, and timing but no private payload by default

### Requirement: Run timelines are queryable
Authorized local and development callers SHALL retrieve a bounded, ordered timeline
and current checkpoint by run ID without reading arbitrary database tables.

#### Scenario: Run timeline is inspected
- **WHEN** an authorized caller requests an existing run with pagination limits
- **THEN** events are returned in committed order with redacted attributes

#### Scenario: Unknown run is requested
- **WHEN** the run ID does not exist in the selected environment
- **THEN** the query returns a stable not-found result without environment details

### Requirement: Replay is effect-free by default
Replay SHALL use the pinned graph version plus fixture references or normalized input
hashes and SHALL never repeat external effects unless a new approved scenario requests them.

#### Scenario: Historical run is replayed
- **WHEN** a developer requests default replay of a completed run
- **THEN** the engine executes in dry-run mode and compares transitions without dispatching effects

#### Scenario: Replay input is unavailable
- **WHEN** required sanitized fixture data cannot be resolved
- **THEN** replay reports insufficient evidence instead of using production data

### Requirement: Evidence retention is bounded
Run evidence SHALL have configurable age and size limits, preserve active runs, and
delete expired terminal evidence without deleting authoritative collection snapshots.

#### Scenario: Retention job reaches an active run
- **WHEN** an active or approval-waiting run exceeds the normal terminal retention age
- **THEN** its recovery checkpoint remains until it terminates or an explicit safety limit applies

#### Scenario: Evidence storage reaches its size cap
- **WHEN** the configured cap is exceeded
- **THEN** oldest eligible terminal evidence is removed and a metric records the cleanup

### Requirement: Metrics are low-cardinality
Workflow metrics SHALL aggregate by graph kind, version, node kind, outcome, and
bounded error code, and SHALL NOT use run, user, card, or deck identifiers as labels.

#### Scenario: Metrics are exported
- **WHEN** a run advances or terminates
- **THEN** counters and durations update without adding high-cardinality or private labels
