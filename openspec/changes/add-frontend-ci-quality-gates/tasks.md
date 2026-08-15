## 1. CI Foundation

- [ ] 1.1 Inventory current Python checks, packaging, secrets, permissions, artifacts, and required external repository settings.
- [ ] 1.2 Add workflow YAML/schema lint and a policy test for permissions, triggers, immutable action SHAs, shells, secrets, and prohibited `pull_request_target` use.
- [ ] 1.3 Create PR/push `ci.yml` with read-only defaults, per-ref concurrency, explicit required check names, and existing Python gates.
- [ ] 1.4 Add OpenSpec strict validation for every active change and scenario-to-test traceability reports.
- [ ] 1.5 Configure branch-protection required checks, review policy, force-push denial, ownership, and emergency procedure outside the repository.

## 2. Frontend Toolchain And Static Gates

- [ ] 2.1 Pin Node, package manager, TypeScript, Wails frontend tools, locale, timezone, and lockfile metadata.
- [ ] 2.2 Add frozen dependency install, formatting check, ESLint, TypeScript no-emit, dependency audit, and production build scripts.
- [ ] 2.3 Add Vitest, Testing Library, deterministic timers, synthetic fixtures, and initial meaningful coverage thresholds.
- [ ] 2.4 Add cache keys derived only from pinned toolchain and lockfiles and prove caches contain no secrets or release outputs.
- [ ] 2.5 Split frontend static, unit, and build checks into independently diagnosable required jobs.

## 3. Generated Contracts

- [ ] 3.1 Generate Wails bindings, central OpenAPI client, graph event unions, command types, error codes, and JSON schemas from authoritative sources.
- [ ] 3.2 Normalize and commit generated outputs and add a clean-tree regeneration gate with pinned generators.
- [ ] 3.3 Define `CompanionClient` plus production Wails and deterministic fixture adapters using the same serializers.
- [ ] 3.4 Add current, previous-compatible, unknown-version, malformed, oversized, duplicate, and out-of-order contract fixtures.
- [ ] 3.5 Add architecture tests that reject frontend-defined graph transition strings or direct storage, process, shell, and network access.

## 4. Frontend Behavior And Accessibility

- [ ] 4.1 Add component tests for empty, loading, stale, partial, offline, error, success, approval, cancelled, and recovered states.
- [ ] 4.2 Add harness streams for collection sync, meta recommendation, Arena handoff, telemetry consent, and graph recovery journeys.
- [ ] 4.3 Add Playwright projects for setup, collection, deck recommendation, handoff, approvals, cancellation, errors, offline mode, and reconnect.
- [ ] 4.4 Add automated accessibility scans and keyboard-only flows with visible focus, semantic names, announcements, and non-color status assertions.
- [ ] 4.5 Add fixed minimum-window, desktop, and wide layout screenshots with pinned fonts, motion, fixtures, locale, and reviewed baseline process.
- [ ] 4.6 Add tests for long localized text, zoom, duplicate events, duplicate clicks, slow streams, cancellation races, and terminal cleanup.
- [ ] 4.7 Scan screenshots, traces, videos, logs, coverage, and reports for credentials, paths, logs, collections, prompts, and private fixtures before upload.

## 5. Windows Wails Validation

- [ ] 5.1 Create `windows-wails.yml` with pinned Go, Node, package manager, Wails, production flags, and read-only permissions.
- [ ] 5.2 Build the real Wails binary on pull requests and protected pushes with frozen frontend and Go dependencies.
- [ ] 5.3 Add Windows smoke tests for generated bindings, embedded assets, SQLite migration, offline startup, WebView strategy, and primary graph states.
- [ ] 5.4 Inspect binaries and archives for dev MCP, test credentials, private fixtures, debug endpoints, source maps, absolute paths, and development configuration.
- [ ] 5.5 Upload short-lived unsigned PR artifacts only after inspection and label them non-production.

## 6. Security Workflows

- [ ] 6.1 Create `security.yml` for CodeQL across Python, Go, and TypeScript plus dependency review, secret, vulnerability, and workflow scans.
- [ ] 6.2 Pin every action to a reviewed full commit SHA and configure Dependabot to propose reviewed action and ecosystem updates.
- [ ] 6.3 Prove fork pull requests receive no write token, OIDC, signing, deployment, provider, database, or production secrets.
- [ ] 6.4 Add artifact and dependency license inventory, forbidden-license policy, and time-bounded vulnerability exception process.
- [ ] 6.5 Add scheduled security validation that reports without silently bypassing pull-request required checks.

## 7. Release Integrity

- [ ] 7.1 Create protected `release.yml` that resolves an approved commit and verifies all required current checks before promotion.
- [ ] 7.2 Rebuild from pinned sources, create version/commit metadata, license inventory, SBOM, SHA-256 checksums, and immutable artifact manifest.
- [ ] 7.3 Provision Windows signing in a protected Environment with approval, timestamp, rotation, audit, and no fork access.
- [ ] 7.4 Sign and verify the final Windows package and refuse production publication when signing or verification fails.
- [ ] 7.5 Run isolated clean-machine startup, WebView, offline, migration, primary workflow, uninstall, and rollback-package smoke tests.
- [ ] 7.6 Publish immutable release artifacts with retained prior rollback package and bounded CI evidence.
- [ ] 7.7 Keep the current Python release workflow until Wails release parity passes, then retire it in a separately reviewed change.

## 8. Final Gates

- [ ] 8.1 Measure PR feedback time and optimize only after preserving all named required checks and trust boundaries.
- [ ] 8.2 Run deliberate failing scenarios for contract drift, accessibility, layout overlap, Windows build, secret exposure, action pin, and unsigned release.
- [ ] 8.3 Map every spec scenario to an automated check or documented protected-environment operation.
- [ ] 8.4 Validate this OpenSpec in strict mode and store the initial green evidence set.
