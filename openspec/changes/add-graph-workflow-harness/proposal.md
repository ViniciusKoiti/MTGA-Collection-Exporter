## Why

The target companion has several multi-step activities, but no executable model
yet guarantees that production flows, tests, and development diagnostics follow
the same transitions. A bounded graph runtime and programmatic harness are needed
before agentic, synchronization, recommendation, and telemetry flows can be made
observable, replayable, and safe.

## What Changes

- Represent every orchestration activity as a versioned graph with typed state,
  nodes, transitions, terminal outcomes, budgets, and policy checkpoints.
- Keep pure domain operations such as card normalization, ownership arithmetic,
  legality checks, and ranking as ordinary deterministic functions called by graph
  nodes rather than creating a graph for every function.
- Add a production graph runtime used by local workflows and background jobs.
- Add a deterministic harness that executes the production runtime with fixtures,
  a controlled clock and planner, fault injection, approvals, and event assertions.
- Persist graph runs and redacted transition events so interrupted activities can
  be inspected, cancelled, retried, or resumed according to their graph policy.
- Add development-only MCP tools for running scenarios and inspecting local,
  development, or staging evidence; production targets are rejected.
- Add CI gates for graph contracts, budgets, replay, recovery, privacy, and a
  100-line limit for manually maintained Go source files with explicit exclusions.

## Capabilities

### New Capabilities

- `graph-workflow-runtime`: Defines versioned graph contracts, execution limits,
  policy checkpoints, effects, retries, cancellation, and recovery.
- `development-harness`: Defines deterministic scenarios that execute the same
  runtime and application services used by the product.
- `workflow-observability`: Defines run and transition evidence, redaction,
  correlation, retention, inspection, and replay semantics.
- `development-mcp`: Defines development-only tools for scenario execution and
  evidence inspection with strict environment isolation.

### Modified Capabilities

None. No capabilities have been archived into `openspec/specs` yet.

## Impact

- Adds graph and testkit packages to the planned Go module plus a development
  harness command; it does not change the current Python runtime initially.
- Adds local SQLite migrations for graph definitions, runs, steps, approvals, and
  outbox records; development and staging tests may also use PostgreSQL fixtures.
- Requires stable ports around time, IDs, planners, tools, effects, event storage,
  central APIs, and external providers so tests can replace boundaries safely.
- The Wails frontend observes graph progress through typed events but does not own
  workflow state or transition decisions.
- The development MCP is excluded from product packaging and is never a runtime
  dependency of the frontend, local backend, or central backend.
