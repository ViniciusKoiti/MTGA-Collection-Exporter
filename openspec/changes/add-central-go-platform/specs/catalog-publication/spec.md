## ADDED Requirements

### Requirement: Catalog imports use approved adapters
The worker SHALL fetch only actively approved providers with explicit timeout, rate,
size, attribution, usage-rights, and schema policies.

#### Scenario: Provider is disabled during retry
- **WHEN** approval is revoked before the next attempt
- **THEN** the worker stops retrying and cannot publish data from that run

### Requirement: Catalog artifacts are immutable and verified
Every published artifact SHALL be schema-versioned, compressed, content-addressed,
SHA-256 hashed, Ed25519 signed through its manifest, and immutable after publication.

#### Scenario: Stored object hash differs
- **WHEN** post-upload verification does not match the expected hash
- **THEN** publication fails and the previous current manifest remains unchanged

### Requirement: Publication is atomic
The worker SHALL stage object upload, verification, metadata transaction, and current
manifest activation so partial work never becomes the advertised current snapshot.

#### Scenario: Activation transaction fails
- **WHEN** storage succeeded but database activation fails
- **THEN** clients continue receiving the previous manifest and cleanup remains retryable

### Requirement: Clients can verify and recover
The manifest SHALL provide key ID, signature, hashes, sizes, schemas, rulesets,
freshness, and compatibility so clients can reject damage and retain the last valid snapshot.

#### Scenario: Signature is unknown
- **WHEN** a client cannot verify the manifest against its trusted key set
- **THEN** it rejects activation and retains its last valid compatible catalog

### Requirement: Import concurrency is bounded
Provider fetch, decode, normalize, persist, artifact, and publish stages SHALL use
configured worker and buffer limits with cancellation and downstream backpressure.

#### Scenario: Database stage slows down
- **WHEN** persist capacity is exhausted
- **THEN** upstream stages block or stop within bounds instead of growing goroutines or memory

### Requirement: Publication is single-owner per target
Concurrent runs for the same catalog kind and target window SHALL deduplicate or lease
ownership and SHALL NOT race current-manifest activation.

#### Scenario: Two workers start the same publication
- **WHEN** both claim an equivalent catalog target
- **THEN** at most one activates a manifest and the other exits or observes the result
