## ADDED Requirements

### Requirement: Frontend contracts are generated
The project SHALL generate Wails bindings, central API clients, graph events, error
codes, and schemas from authoritative Go or OpenAPI definitions.

#### Scenario: Backend contract changes
- **WHEN** CI regenerates contracts after an authoritative source change
- **THEN** it fails until the reviewed generated outputs and frontend usage are consistent

### Requirement: Generated outputs are reproducible
Contract generation SHALL use pinned tools and normalized output and SHALL leave no
repository diff when authoritative inputs are unchanged.

#### Scenario: Generation is repeated
- **WHEN** the same commit generates contracts twice in clean environments
- **THEN** tracked generated files are byte-equivalent

### Requirement: Supported event versions remain compatible
The frontend SHALL parse current and declared previous graph event schema versions and
SHALL reject unknown incompatible versions with a recoverable update state.

#### Scenario: Previous event fixture is loaded
- **WHEN** a supported prior event stream reconstructs a run
- **THEN** the UI reaches the same semantic state with declared compatibility mapping

### Requirement: Frontend cannot invent workflow transitions
The presentation layer MUST NOT define arbitrary graph transitions and SHALL send only
registered typed commands such as start, approve, cancel, retry, and resume.

#### Scenario: Component requests an unknown transition
- **WHEN** frontend code attempts to call a transition outside generated commands
- **THEN** type or architecture validation fails before merge

### Requirement: Test and production clients share serialization
The fake `CompanionClient` SHALL consume the same generated request, response, and event
serializers as the Wails production adapter.

#### Scenario: Fixture violates production schema
- **WHEN** a test fixture has an unknown field or invalid discriminant
- **THEN** fixture validation fails before the component test starts
