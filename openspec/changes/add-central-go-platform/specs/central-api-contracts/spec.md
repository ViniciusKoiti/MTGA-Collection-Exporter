## ADDED Requirements

### Requirement: Public APIs are versioned and schema strict
The platform SHALL publish versioned HTTPS and OpenAPI contracts with bounded JSON
schemas, stable error codes, request IDs, compatibility policy, and unknown-field rejection.

#### Scenario: Client sends an unknown field
- **WHEN** a request includes a field absent from the selected API schema
- **THEN** the API rejects it before invoking an application service

### Requirement: HTTP resources are bounded
The API SHALL enforce header, compressed and decompressed body, batch, response,
concurrency, request-time, query-time, and idle-time limits.

#### Scenario: Decompressed payload exceeds its cap
- **WHEN** a compressed request expands beyond the endpoint limit
- **THEN** processing stops with a stable too-large error and no batch data commits

### Requirement: Cancellation reaches dependencies
Every handler SHALL propagate request cancellation and deadline through application,
PostgreSQL, object-store, and provider calls and SHALL NOT detach post-response work.

#### Scenario: Client disconnects
- **WHEN** the request context is cancelled during a dependency call
- **THEN** downstream work is cancelled or left as a durable job and no orphan goroutine remains

### Requirement: Mutating requests are idempotent
Telemetry, enrollment, deletion, and operations mutations SHALL require scoped
idempotency keys with atomic result storage and bounded retention.

#### Scenario: Completed request is retried
- **WHEN** the same principal and idempotency key submit an equivalent mutation
- **THEN** the API returns the original semantic result without applying it twice

### Requirement: Authentication boundaries are separate
The platform SHALL separate public catalog, pseudonymous installation, and privileged
operations credentials, audiences, roles, rate policies, and audit records.

#### Scenario: Installation credential calls operations
- **WHEN** an installation token is presented to an operations endpoint
- **THEN** authorization fails before any privileged query or job is created

### Requirement: Catalog retrieval supports efficient caching
Manifest and artifact contracts SHALL expose immutable identity, ETag, content hash,
signature, compatibility, and cache policy.

#### Scenario: Client already has the current manifest
- **WHEN** a conditional request presents the current ETag
- **THEN** the API returns not-modified without loading the artifact body

### Requirement: Production exposes no development control plane
The production API MUST NOT expose MCP, arbitrary SQL, shell, fixture mutation,
scenario execution, graph mutation, or generic database inspection endpoints.

#### Scenario: Development route is requested
- **WHEN** a caller requests a development-only path in production
- **THEN** routing returns not found and no development handler is linked
