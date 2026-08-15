## ADDED Requirements

### Requirement: Shared typed tool registry
The system SHALL expose agent capabilities through a typed tool registry backed by the same application services used by the desktop UI.

#### Scenario: Assistant searches the collection
- **WHEN** the in-app assistant calls a collection search tool with schema-valid arguments
- **THEN** the registry invokes the collection query service and returns a bounded structured result with snapshot provenance

#### Scenario: Tool arguments are invalid
- **WHEN** a client supplies arguments outside the published schema or size limits
- **THEN** the registry rejects the call without invoking the application service

### Requirement: Read-only default policy
The agent policy engine SHALL classify unrecognized tools as denied and SHALL allow only registered read operations without approval.

#### Scenario: Registered read tool is called
- **WHEN** the assistant requests collection statistics, card lookup, snapshot status, deck validation, or ownership gaps
- **THEN** the tool runs without side effects and without a confirmation prompt

#### Scenario: Unknown tool is called
- **WHEN** a model or other registered caller requests a tool absent from the policy registry
- **THEN** the system denies the request and records the denial

### Requirement: Exact approval for side effects
The system SHALL require explicit user confirmation before an agent executes a local side effect and SHALL bind approval to the exact operation and arguments.

#### Scenario: Assistant proposes clipboard export
- **WHEN** the assistant proposes copying an Arena deck list
- **THEN** the system displays a preview and returns an unexecuted approval request

#### Scenario: User confirms an unchanged proposal
- **WHEN** the user confirms an unexpired approval request whose operation and arguments are unchanged
- **THEN** the system executes the operation once and records its outcome

#### Scenario: Arguments change after preview
- **WHEN** a caller changes any approved argument before execution
- **THEN** the system invalidates the approval and requires a new preview

### Requirement: Prohibited MTGA operations
The agent tool registry MUST NOT expose process-memory access, synthetic input, gameplay actions, purchases, account changes, or credential access.

#### Scenario: Model asks to play a match action
- **WHEN** a model requests a click, keypress, card play, target selection, queue action, purchase, or other MTGA control
- **THEN** the policy engine refuses the request and no operating-system control primitive is invoked

### Requirement: Minimal model context
The system SHALL keep remote model use opt-in and SHALL send only the minimum structured data needed for the approved task.

#### Scenario: Assistant explains deck ownership gaps
- **WHEN** the user asks a configured remote model to explain a deterministic ownership result
- **THEN** the context includes the relevant deck entries and gap result but excludes raw logs, credentials, unrelated collection entries, and local paths

### Requirement: Auditable tool execution
The system SHALL record agent tool requests, policy decisions, approvals, durations, and outcomes with sensitive values redacted.

#### Scenario: Tool call completes
- **WHEN** any agent tool succeeds, fails, or is denied
- **THEN** the audit log records correlation ID, tool name, risk class, decision, duration, result status, and redacted argument summary

### Requirement: Deterministic facts
The system SHALL calculate card counts, ownership, legality, and validation in deterministic application services rather than accepting those facts from model output.

#### Scenario: Model statement conflicts with a tool result
- **WHEN** a model claims a deck is buildable but the ownership service reports missing cards
- **THEN** the application presents the deterministic result as authoritative and labels the model statement as unsupported
