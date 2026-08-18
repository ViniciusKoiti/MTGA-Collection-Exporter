## ADDED Requirements

### Requirement: Enrollment requires explicit consent
The platform SHALL accept telemetry only after the desktop records explicit purpose-
and-version-specific consent and obtains a pseudonymous installation credential.

#### Scenario: Client has not consented
- **WHEN** a client attempts enrollment or event upload without the required consent state
- **THEN** no telemetry event is accepted and catalog access remains available

### Requirement: Telemetry uses an allowlisted minimal schema
The ingest service SHALL accept only registered bounded event types and attributes and
MUST reject raw logs, collections, decks, paths, prompts, credentials, and MTGA identities.

#### Scenario: Batch contains a forbidden field
- **WHEN** any event includes a forbidden or unknown attribute
- **THEN** the batch is rejected atomically with the schema violation identified

### Requirement: Batches are idempotent and ordered enough
Every batch SHALL carry installation scope, schema version, batch ID, sequence range,
created time, and event IDs so retries do not change aggregates.

#### Scenario: Batch is delivered twice
- **WHEN** an accepted batch is retried with equivalent identity and content
- **THEN** the API returns the prior result and aggregates receive no duplicate contribution

### Requirement: Revocation and deletion are enforceable
The platform SHALL block new events after revocation and SHALL provide authenticated
deletion that erases or irreversibly anonymizes traceable installation data.

#### Scenario: Deletion completes
- **WHEN** a valid deletion secret requests deletion
- **THEN** traceable data is removed or anonymized and the installation credential cannot upload again

### Requirement: Raw accepted events have short retention
Accepted events SHALL expire after the documented aggregation and abuse-review window,
while aggregates SHALL exclude identifiers that permit reconstruction of a user history.

#### Scenario: Retention job runs
- **WHEN** accepted events exceed their approved age
- **THEN** they are deleted without changing finalized aggregate totals

### Requirement: Backpressure is explicit
Telemetry ingest SHALL reject or defer batches with bounded retry guidance when database,
queue, rate, or concurrency budgets are exhausted.

#### Scenario: Admission closes
- **WHEN** the service cannot durably accept a batch within its deadline
- **THEN** it returns a retryable outcome and does not acknowledge acceptance
