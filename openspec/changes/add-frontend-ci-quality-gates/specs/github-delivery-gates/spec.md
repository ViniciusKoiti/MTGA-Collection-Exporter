## ADDED Requirements

### Requirement: Pull requests execute required validation
The repository SHALL run named OpenSpec, Python, Go, graph, frontend, Windows, and
security checks for pull requests and protected-branch pushes.

#### Scenario: Documentation-only pull request runs
- **WHEN** path optimization skips an internal expensive step
- **THEN** every required workflow still reports an explicit successful or failed check

### Requirement: Workflow permissions are least privilege
GitHub workflows SHALL default to read-only contents and SHALL grant write, identity,
security-event, or artifact permissions only to the exact job that requires them.

#### Scenario: Pull request from a fork runs
- **WHEN** untrusted code executes in CI
- **THEN** it receives no repository write token, signing secret, deployment secret, or production credential

### Requirement: Action dependencies are immutable
Every external action SHALL be allowlisted and pinned to a reviewed full commit SHA,
with automated update proposals requiring normal review.

#### Scenario: Workflow uses a mutable tag
- **WHEN** workflow policy scans an action reference that is not a permitted immutable SHA
- **THEN** required validation fails before the workflow change can merge

### Requirement: Obsolete runs are cancelled safely
PR and branch workflows SHALL use per-ref concurrency to cancel obsolete validation
without cancelling protected releases or leaving a required check absent.

#### Scenario: New commit is pushed to a pull request
- **WHEN** prior validation is still running
- **THEN** the old run is cancelled and the new commit receives all required checks

### Requirement: Security analysis is continuous
The repository SHALL run CodeQL, dependency review, secret detection, vulnerability
checks, and workflow lint on pull requests or an appropriate protected schedule.

#### Scenario: Known vulnerable production dependency is introduced
- **WHEN** the configured severity policy detects it
- **THEN** merge is blocked or an approved time-bounded exception is required

### Requirement: Protected branches require evidence
The default branch SHALL require current successful checks, reviewed changes, and no
force push or direct bypass outside documented emergency procedure.

#### Scenario: Required check is missing
- **WHEN** a commit has not produced a required status
- **THEN** branch protection prevents merge or release promotion
