## Why

The current GitHub workflow validates Python only on tags or manual runs, so the
planned Wails frontend could merge with broken types, inaccessible graph states, or
an unbuildable Windows package. Pull-request gates must validate the frontend contract,
experience, supply chain, and product artifact before release.

## What Changes

- Add pull-request CI for TypeScript formatting, lint, type checking, unit/component
  tests, deterministic graph fixtures, accessibility, responsive layout, and Playwright.
- Generate Go-to-TypeScript bindings and event schemas and fail when committed outputs
  differ, preventing silent frontend/backend contract drift.
- Build the real Wails desktop on Windows and smoke-test offline startup, WebView
  behavior, graph progress, approval, cancellation, failure, and recovery states.
- Split validation, security, Windows build, and release workflows with least-privilege
  tokens, concurrency cancellation, immutable action references, and no PR secrets.
- Add CodeQL, dependency review, secret checks, SBOM, checksums, artifact inspection,
  signing policy, and protected-branch required checks.

## Capabilities

### New Capabilities

- `frontend-contract-validation`: Verifies generated bindings, graph events, schemas,
  error codes, and compatibility between Go services and TypeScript clients.
- `frontend-experience-testing`: Verifies behavior, accessibility, responsiveness,
  visual stability, offline states, and primary Wails workflows.
- `github-delivery-gates`: Defines secure PR, security, Windows build, and release
  workflows plus required checks and evidence retention.
- `windows-release-integrity`: Defines reproducible packaging, artifact contents,
  checksums, SBOM, signing, clean-machine smoke tests, and rollback evidence.

### Modified Capabilities

None. Existing frontend and harness capabilities are active changes, not archived specs.

## Impact

- Replaces the single tag-oriented workflow with separate least-privilege workflows.
- Adds frontend scripts, Playwright projects, accessibility tooling, contract fixtures,
  Windows smoke tests, CodeQL configuration, Dependabot, and ownership rules.
- Requires branch protection configuration outside the repository and optional signing
  credentials protected by a GitHub Environment for release jobs only.
- Preserves the current Python gates while adding Go, graph, frontend, and package gates.
