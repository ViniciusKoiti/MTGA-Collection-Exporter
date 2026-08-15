## ADDED Requirements

### Requirement: Handoff uses a pinned preview
The system SHALL validate, allocate, and render an Arena deck preview pinned to deck,
collection, catalog, ruleset, printing choices, options, and rendered text hash.

#### Scenario: Preview is valid
- **WHEN** a supported deck revision completes preflight
- **THEN** the workspace shows legality, missing copies, wildcard requirements, exact effects, and a handoff command

#### Scenario: Pinned input changes
- **WHEN** any pinned input changes after preview
- **THEN** prior approval is invalid and no clipboard or launch effect executes

### Requirement: User and agent approval are explicit
A direct user handoff click SHALL approve the visible pinned effects, while an
assistant-initiated handoff SHALL require an unexpired exact approval token.

#### Scenario: User clicks the visible command
- **WHEN** the current preview and requested effects are visible and unchanged
- **THEN** the graph may execute them once without a redundant confirmation dialog

#### Scenario: Agent proposes handoff
- **WHEN** the assistant requests clipboard or launch effects
- **THEN** the system returns a preview without executing either effect

### Requirement: Clipboard copy is exact and idempotent
The handoff SHALL copy only the approved Arena-format text and SHALL record a hash,
not the deck text, in workflow evidence.

#### Scenario: Copy is retried after acknowledgement loss
- **WHEN** recovery retries the approved copy with the same idempotency key
- **THEN** the approved content remains the clipboard result and only one logical effect is recorded

### Requirement: Client launch is narrow and optional
The system SHALL launch only an allowlisted configured MTGA client descriptor through
the platform API with fixed arguments and SHALL retain copy-only behavior on failure.

#### Scenario: Client is already running
- **WHEN** launch was requested and supported process presence is detected
- **THEN** the system reports `already_running` without focusing or controlling a window

#### Scenario: Launch fails after copy
- **WHEN** clipboard copy succeeds but the configured client cannot start
- **THEN** the result reports copied plus launch failure and offers a launch retry without recopying

### Requirement: Handoff never controls MTGA
The product MUST NOT synthesize input, focus or navigate MTGA, write its memory or
files, call undocumented endpoints, import a deck, edit a deck, or initiate crafting.

#### Scenario: Caller requests automatic import
- **WHEN** any UI, assistant, graph, or development tool requests an in-client action
- **THEN** policy denies it before any operating-system control primitive is invoked

### Requirement: Completion status is truthful
The graph SHALL report only preparation, clipboard, process-presence, and launch facts
that it directly observed and SHALL NOT infer import or deck-open success.

#### Scenario: Client starts successfully
- **WHEN** the configured client process starts after an approved request
- **THEN** the result is `client_started`, not `imported` or `deck_opened`
