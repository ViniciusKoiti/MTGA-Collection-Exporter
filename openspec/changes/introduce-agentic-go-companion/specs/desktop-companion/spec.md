## ADDED Requirements

### Requirement: Guided first-run setup
The desktop companion SHALL guide a first-time user through source detection, Detailed Logs configuration, privacy choices, and source verification before presenting sync as ready.

#### Scenario: First launch with logs disabled
- **WHEN** the application detects MTG Arena but cannot verify Detailed Logs
- **THEN** it shows the exact in-game setting to enable, a verification action, and an explicit import alternative

#### Scenario: Setup is complete
- **WHEN** a source is verified and privacy choices are saved
- **THEN** the application opens the operational home view and clearly reports the current sync state

### Requirement: Operational home view
The desktop companion SHALL present MTGA presence, active source, last successful sync, freshness, collection delta, and the next available action in one view.

#### Scenario: Collection is fresh
- **WHEN** a valid recent snapshot exists
- **THEN** Home shows the snapshot time, source, unique cards, total copies, change since the prior snapshot, and actions for Collection and Decks

#### Scenario: Sync fails
- **WHEN** synchronization fails with a recoverable error
- **THEN** Home keeps the last valid collection visible and places recovery beside the failed status

### Requirement: Collection exploration
The desktop companion SHALL let users search, sort, filter, inspect, and compare collection data without requiring the assistant.

#### Scenario: User filters owned cards
- **WHEN** the user combines text, color, rarity, set, type, and minimum-copy filters
- **THEN** the collection view updates the result count and keeps the filter state visible and removable

#### Scenario: User inspects an unresolved record
- **WHEN** the user opens an unresolved collection item
- **THEN** the application shows the source identifier, quantity, snapshot, and catalog recovery status

### Requirement: Accessible status and interaction
The desktop companion SHALL provide keyboard-operable controls, visible focus, readable labels, and non-color status cues for primary workflows.

#### Scenario: Keyboard-only sync workflow
- **WHEN** a user navigates setup and starts a sync using only the keyboard
- **THEN** focus order follows the task order and every action and state remains perceivable without color or mana symbols alone

### Requirement: Actionable error messages
The desktop companion SHALL map domain and adapter error codes to concise user-facing recovery actions and retain diagnostic detail separately.

#### Scenario: Card catalog is unavailable
- **WHEN** a sync observation is valid but the catalog cannot update
- **THEN** the application identifies whether cached metadata was used and offers retry without discarding the observation

