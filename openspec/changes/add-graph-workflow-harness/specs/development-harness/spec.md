## ADDED Requirements

### Requirement: Harness executes the production runtime
The development harness SHALL construct and invoke the same graph engine, registry,
nodes, policies, migrations, and application entry points used by the product.

#### Scenario: Deterministic scenario runs
- **WHEN** a scenario starts a registered activity
- **THEN** the harness invokes its production entry point and captures the resulting persisted run

#### Scenario: Alternate transition logic is introduced
- **WHEN** harness code attempts to define transitions outside the production registry
- **THEN** architecture tests fail and the scenario cannot run with the alternate definition

### Requirement: Scenarios are typed and versioned
A scenario SHALL declare its schema version, target graph, input fixture, dependency
profile, scripted decisions, injected failures, approvals, and expected assertions.

#### Scenario: Scenario schema is unsupported
- **WHEN** a scenario uses an unknown schema version or field
- **THEN** validation fails before fixtures, databases, or graph nodes are started

#### Scenario: Fixture contains private source data
- **WHEN** fixture validation detects credentials, absolute user paths, raw identifiers, or unredacted logs
- **THEN** the scenario is rejected with the offending fixture and rule identified

### Requirement: Deterministic boundaries are controllable
The deterministic profile SHALL replace time, IDs, planners, external APIs, process
detection, model providers, and effect adapters through the production ports.

#### Scenario: Scenario is repeated
- **WHEN** the same deterministic scenario and seed run twice
- **THEN** normalized states, events, outcomes, and effect intents are identical

#### Scenario: Default CI suite runs
- **WHEN** the deterministic suite runs in CI
- **THEN** it requires no MTGA process, administrator access, network, or LLM

### Requirement: Faults occur at declared boundaries
The harness SHALL inject latency, timeout, malformed response, disconnection, crash,
and dependency errors through port decorators or explicit runtime checkpoints.

#### Scenario: Crash is injected after checkpoint
- **WHEN** a scenario stops the runner after a node commit and restarts it
- **THEN** the scenario can assert recovery position and absence of duplicated effects

#### Scenario: Undeclared fault target is requested
- **WHEN** a scenario targets an internal function without an injectable boundary
- **THEN** validation rejects the fault instead of modifying production code at runtime

### Requirement: Assertions cover behavior and evidence
The harness SHALL assert terminal outcome, ordered transitions, attempts, effects,
budgets, database state, event redaction, and absence of forbidden behavior.

#### Scenario: Expected event is missing
- **WHEN** execution finishes without a required event or transition
- **THEN** the report identifies the first divergence with expected and actual evidence

### Requirement: Profiles have explicit trust levels
Integration, end-to-end, staging, and provider-smoke profiles SHALL be opt-in and
SHALL declare their allowed hosts, credentials, databases, and side effects.

#### Scenario: Deterministic profile requests a network host
- **WHEN** any deterministic scenario attempts outbound network access
- **THEN** access is denied and the scenario fails

#### Scenario: Staging profile targets production
- **WHEN** a staging scenario resolves a production host or database identity
- **THEN** the harness refuses startup before opening a connection
