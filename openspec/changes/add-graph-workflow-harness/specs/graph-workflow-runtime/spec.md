## ADDED Requirements

### Requirement: Orchestration activities use registered graphs
Every application orchestration activity SHALL execute through a registered,
versioned graph when it is multi-step, effectful, retryable, approval-gated,
scheduled, or resumable. Pure domain calculations SHALL remain callable
deterministic services.

#### Scenario: Registered activity starts
- **WHEN** a caller starts a supported activity with schema-valid input
- **THEN** the runtime creates a run pinned to the registered graph kind and version

#### Scenario: Invalid graph is registered
- **WHEN** a graph has an unknown node, unreachable terminal state, duplicate identity, or transition without an outcome
- **THEN** registration fails before the graph can accept a run

### Requirement: Graph execution is bounded
The runtime SHALL enforce step, tool-call, repeated-call, input-size, and active-time
limits even when a planner or node requests further work.

#### Scenario: Repeated call limit is reached
- **WHEN** a run requests the same tool with equivalent arguments more than its limit
- **THEN** the run terminates with a stable budget-exhausted outcome and no further tool executes

#### Scenario: Approval waits beyond the active deadline
- **WHEN** a run is persisted in an approval-wait state
- **THEN** waiting time does not consume active execution time and the approval itself expires independently

### Requirement: Transitions and tools are policy controlled
Only compiled transitions and allowlisted typed tools SHALL be executable. A model,
fixture, catalog, log entry, or tool result SHALL NOT modify the graph definition.

#### Scenario: Planner requests an unknown transition
- **WHEN** planner output names a transition absent from the pinned definition
- **THEN** the runtime rejects the output, records a redacted policy event, and executes no effect

#### Scenario: Effect requires approval
- **WHEN** a node proposes an approval-gated effect
- **THEN** the run persists an exact preview and pauses until a matching live approval is presented

### Requirement: Runs checkpoint transactionally
The runtime SHALL atomically persist every committed node outcome, next node,
run version, and outbox record before continuing execution.

#### Scenario: Process stops after a committed node
- **WHEN** the process restarts after committing a transition but before starting the next node
- **THEN** recovery resumes from the committed next node without repeating the completed node

#### Scenario: Two workers acquire one run
- **WHEN** two executors attempt to advance the same runnable version
- **THEN** at most one executor commits and the other receives a concurrency conflict

### Requirement: Effects are idempotent and cancellable
Every effect SHALL use a stable idempotency key, and cancellation or deadline signals
SHALL propagate to active nodes and prevent new nodes from starting.

#### Scenario: Effect acknowledgement is lost
- **WHEN** an effect succeeds but its acknowledgement is interrupted
- **THEN** retrying with the same idempotency key does not apply the effect twice

#### Scenario: Run is cancelled
- **WHEN** cancellation is accepted for a non-terminal run
- **THEN** no new node starts and the run reaches a persisted cancelled outcome

### Requirement: Graph version compatibility is explicit
A run SHALL remain pinned to its original graph version, and the runtime SHALL refuse
automatic recovery when that version or its state migration is unavailable.

#### Scenario: Pinned version is unavailable
- **WHEN** recovery loads a run whose graph version is not registered and has no approved migration
- **THEN** the run is marked incompatible without executing a node or effect
