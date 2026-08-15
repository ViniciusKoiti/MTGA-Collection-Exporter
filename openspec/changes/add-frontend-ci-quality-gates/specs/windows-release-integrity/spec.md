## ADDED Requirements

### Requirement: Windows build uses pinned production inputs
The Windows workflow SHALL build Wails from a pinned Go, Node, package-manager,
lockfile, Wails, WebView policy, and production configuration.

#### Scenario: Lockfile is inconsistent
- **WHEN** frozen dependency installation detects manifest drift
- **THEN** the build fails before compiling or packaging the application

### Requirement: Product artifacts exclude development content
The package MUST NOT contain development MCP, test credentials, private fixtures,
source maps with private paths, debug endpoints, or development configuration.

#### Scenario: Artifact inspection finds dev MCP
- **WHEN** a release candidate contains a development entry point or manifest reference
- **THEN** promotion fails and the artifact is not published

### Requirement: Release artifacts have provenance
Every candidate SHALL include version and commit metadata, dependency/license inventory,
SBOM, SHA-256 checksums, build evidence, and immutable artifact identity.

#### Scenario: Artifact checksum is missing
- **WHEN** release promotion inspects the candidate
- **THEN** promotion fails before signing or publication

### Requirement: Production releases are signed and verified
A production Windows release SHALL be signed with a protected timestamped code-signing
identity and SHALL pass signature verification after final packaging.

#### Scenario: Signing is unavailable
- **WHEN** no approved signing identity can sign the candidate
- **THEN** it may remain a labeled development artifact but cannot be a production release

### Requirement: Clean-machine behavior is proven
The release candidate SHALL pass isolated Windows startup, WebView dependency, offline,
local migration, rollback export, and primary workflow smoke tests.

#### Scenario: WebView runtime is unavailable
- **WHEN** the clean-machine scenario follows the selected WebView strategy
- **THEN** startup follows the documented recovery behavior without silent failure

### Requirement: Release is a protected promotion
Only a commit with current required checks and protected-environment approval SHALL be
eligible for immutable release publication, and the prior rollback package SHALL remain available.

#### Scenario: Tag points to an unvalidated commit
- **WHEN** release workflow resolves a commit without required successful evidence
- **THEN** publication is refused even though the tag exists
