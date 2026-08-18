## ADDED Requirements

### Requirement: Development MCP is outside the product runtime
The development MCP SHALL run as a separate stdio process and SHALL be excluded from
desktop and central production packaging, startup, and dependency paths.

#### Scenario: Production package is inspected
- **WHEN** packaging verification examines binaries, manifests, and runtime dependencies
- **THEN** no development MCP entry point, configuration, or credential is present

### Requirement: Environment identity is verified before access
The MCP SHALL require an explicit local, development, or staging identity and SHALL
reject production DSNs, hosts, certificates, database markers, and ambiguous targets.

#### Scenario: Production database is configured
- **WHEN** any resolved connection attribute identifies production
- **THEN** startup fails before DNS resolution, authentication, or a database connection

#### Scenario: Environment cannot be proven safe
- **WHEN** environment markers are missing or contradictory
- **THEN** the MCP defaults to denial and exposes no tools

### Requirement: Tools are narrow and schema validated
The MCP SHALL expose only graph listing, validated scenario execution, run timeline,
event validation, and fixture diagnostics; it SHALL expose no arbitrary SQL, shell,
filesystem, URL fetch, graph mutation, or generic effect tool.

#### Scenario: Valid run inspection is requested
- **WHEN** a client calls the timeline tool with a valid run ID and page limit
- **THEN** the tool returns bounded redacted evidence through the observability service

#### Scenario: Client requests arbitrary SQL
- **WHEN** a prompt or tool argument asks to execute SQL outside a registered operation
- **THEN** the request is rejected and no database statement is created

### Requirement: Read-only access is the default
Inspection tools SHALL use a restricted read role. Scenario execution SHALL require
an explicit capability and SHALL write only to isolated fixture or staging namespaces.

#### Scenario: Read role invokes scenario execution
- **WHEN** a client without the scenario-execute capability calls `run_scenario`
- **THEN** authorization fails before scenario fixtures or stores are initialized

#### Scenario: Scenario namespace is not isolated
- **WHEN** the target database cannot guarantee the configured test namespace
- **THEN** execution is refused without modifying data

### Requirement: MCP operations are auditable and bounded
Every MCP request SHALL have size, time, concurrency, and result limits and SHALL
record client, tool, environment, outcome, and correlation data without prompt text.

#### Scenario: Tool response exceeds its limit
- **WHEN** an inspection result is larger than the configured page or byte limit
- **THEN** the response is truncated through pagination and the limit outcome is audited

#### Scenario: Client disconnects during a scenario
- **WHEN** the stdio client disconnects while a scenario is active
- **THEN** policy either cancels the run or leaves a recoverable checkpoint and records the decision
