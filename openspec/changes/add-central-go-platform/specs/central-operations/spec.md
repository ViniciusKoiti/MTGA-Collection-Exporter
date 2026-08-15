## ADDED Requirements

### Requirement: Background work is durable and leased
The worker SHALL claim PostgreSQL jobs transactionally with idempotency, attempts,
visibility time, bounded leases, heartbeat, and recoverable terminal outcomes.

#### Scenario: Worker dies during a job
- **WHEN** its lease expires without a terminal commit
- **THEN** another worker may resume according to retry policy without duplicating effects

### Requirement: Concurrency is owned and bounded
Every goroutine SHALL have a parent context, owner, result or error path, configured
limit, and shutdown behavior; channels SHALL have bounded capacity.

#### Scenario: Downstream consumer stops
- **WHEN** a pipeline stage fails or returns early
- **THEN** cancellation unblocks upstream senders and all owned goroutines exit

### Requirement: Database capacity controls admission
API and worker processes SHALL enforce pool, acquisition, query, job-worker, and replica
budgets derived from a reserved PostgreSQL connection capacity.

#### Scenario: Pool budget is exhausted
- **WHEN** work cannot acquire a connection within its deadline
- **THEN** it fails or retries through bounded policy without starting more goroutines

### Requirement: Health and readiness are distinct
Liveness SHALL report process health, while readiness SHALL verify compatible schema,
essential dependencies, pool pressure, queue age, and required publication capability.

#### Scenario: Queue age exceeds its safety threshold
- **WHEN** the oldest essential job is too old
- **THEN** readiness degrades with a stable internal reason and no sensitive detail

### Requirement: Shutdown is graceful and bounded
On termination the platform SHALL fail readiness, stop claims, cancel producers, drain
bounded workers, persist checkpoints, shut down HTTP, and close pools within deadlines.

#### Scenario: Shutdown deadline expires
- **WHEN** one dependency does not drain in time
- **THEN** incomplete durable work remains recoverable and the process exits non-successfully

### Requirement: Observability is structured and private
The platform SHALL emit correlated structured logs, traces, and low-cardinality metrics
without payloads, credentials, decks, collection data, or identifiers as metric labels.

#### Scenario: Operation fails
- **WHEN** a request, node, job, query, or publication fails
- **THEN** evidence includes bounded codes, duration, component, and correlation without private payload

### Requirement: Scale gates precede capacity claims
The platform SHALL pass documented load, burst, soak, fault, restore, and privacy gates
with at least 30 percent PostgreSQL headroom before a target stage is declared supported.

#### Scenario: Burst test passes but soak fails
- **WHEN** queue, goroutine, memory, pool, or error growth is unbounded during soak
- **THEN** the target stage remains unsupported despite short load-test success
