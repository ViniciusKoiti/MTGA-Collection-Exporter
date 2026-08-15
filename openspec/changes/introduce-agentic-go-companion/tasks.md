## 1. Discovery And Safety Gates

- [ ] 1.1 Create sanitized golden fixtures for the current Python JSON export, including duplicate printings, unknown cards, empty data, and malformed data.
- [ ] 1.2 Capture sanitized Detailed Logs fixtures from the current MTG Arena client and document whether a complete collection payload can be detected reliably.
- [ ] 1.3 Add a fixture-based decision test that enables the log source only for recognized payload versions and falls back to explicit import otherwise.
- [ ] 1.4 Run a Windows packaging spike for stable Wails and candidate `database/sql` SQLite drivers; record binary, WebView, migration, and clean-machine results.
- [ ] 1.5 Document the integration policy matrix for allowed read operations, approval-required local effects, and prohibited MTGA control operations.

## 2. Go Domain And Application Foundation

- [x] 2.1 Initialize the Go module and create `cmd`, domain, application, ports, adapters, policy, and frontend package boundaries from the design.
- [x] 2.2 Implement typed card identity, collection observation, immutable snapshot, diagnostic, freshness, deck, ruleset, approval, and audit models.
- [x] 2.3 Define source, catalog, snapshot store, deck store, clock, clipboard, exporter, approval, audit, and optional model interfaces with compile-time adapter assertions.
- [x] 2.4 Implement stable application error codes and presentation-safe error mapping without leaking raw paths or exceptions.
- [x] 2.5 Add unit tests that enforce domain invariants and import-boundary tests that prevent domain/application packages from depending on adapters or UI.

## 3. Collection Sync And Storage

- [ ] 3.1 Implement the legacy JSON source adapter and golden contract tests against Python exports.
- [ ] 3.2 Implement normalization that preserves unresolved source records and attaches catalog provenance and diagnostics.
- [ ] 3.3 Implement idempotent snapshot orchestration and the `not_configured`, `ready`, `watching`, `syncing`, `fresh`, `stale`, and `error` state transitions.
- [ ] 3.4 Implement the selected SQLite adapter with embedded migrations, transactional snapshot writes, and repository integration tests.
- [ ] 3.5 Implement atomic compatibility JSON, CSV, and text projections and verify the existing MCP consumer can parse the generated JSON.
- [ ] 3.6 Implement Scryfall bulk-data caching with explicit version/freshness metadata, bounded retries, timeouts, and offline fallback tests.
- [ ] 3.7 Implement the Detailed Logs parser and incremental watcher behind a disabled-by-default feature flag using only accepted fixtures.
- [ ] 3.8 Implement Windows log-path and MTGA-process presence detection without opening the process for memory access.
- [ ] 3.9 Implement the isolated legacy Python scanner bridge as an explicit compatibility command, excluding it from agent tools.

## 4. Desktop Companion Flow

- [ ] 4.1 Scaffold the stable Wails TypeScript desktop shell with Home, Collection, Decks, Assistant, and Settings navigation.
- [ ] 4.2 Implement first-run setup for source detection, Detailed Logs guidance, privacy choices, source verification, and explicit import fallback.
- [ ] 4.3 Implement Home with MTGA presence, source health, last sync, freshness, snapshot totals, delta, progress, and in-context recovery actions.
- [ ] 4.4 Implement Collection search, sorting, compact compound filters, details, unresolved records, and snapshot comparison.
- [ ] 4.5 Implement consistent empty, loading, stale, partial, error, and success states using stable application error codes.
- [ ] 4.6 Add keyboard navigation, visible focus, text status labels, scalable layout checks, and automated accessibility assertions for primary flows.
- [ ] 4.7 Add component and end-to-end tests for setup, successful import, stale collection, failed sync recovery, and unresolved-card inspection.

## 5. Deck Workspace

- [ ] 5.1 Implement Arena deck-text parsing and formatting with fixtures for main deck, sideboard, localized names, unknown cards, and malformed lines.
- [ ] 5.2 Implement ownership comparison against a selected snapshot with required, owned, missing, and wildcard-relevant quantities.
- [ ] 5.3 Implement a versioned Standard ruleset first, including deck size, copy limits, sideboard constraints, card legality, and stale-catalog reporting.
- [ ] 5.4 Implement deterministic owned-card substitution candidates with documented color, mana, type, format, and ranking evidence.
- [ ] 5.5 Implement local saved-deck revisions linked to the collection snapshot and ruleset version used for validation.
- [ ] 5.6 Implement the Decks workspace with structured editing, ownership and legality results, substitutions, revision history, and Arena export preview.
- [ ] 5.7 Implement direct user copy/export commands and approval-token execution for assistant-requested copy/export operations.

## 6. Agent Tools And Legacy Compatibility

- [ ] 6.1 Implement the typed tool registry with JSON schemas, argument limits, bounded result pagination, cancellation, and correlation IDs.
- [ ] 6.2 Implement default-deny policy classification, exact argument-bound approval tokens, expiration, single-use execution, and denial tests.
- [ ] 6.3 Implement redacted audit storage and an Assistant activity view for requested, approved, executed, failed, and denied tools.
- [ ] 6.4 Expose read tools for sync status, collection summary/search, card lookup, deck validation, and ownership gaps to the in-app assistant and development harness without shipping a product MCP.
- [ ] 6.5 Expose preview-only proposals for save, sync, file export, and clipboard export, requiring approval before execution.
- [ ] 6.6 Add explicit negative tests proving that memory access, synthetic input, gameplay, purchase, account, credential, shell, and unrestricted filesystem tools are absent or denied.
- [ ] 6.7 Implement a provider-neutral context builder that excludes raw logs, local paths, credentials, and unrelated collection data, with snapshot tests for redaction.
- [ ] 6.8 Add the disabled-by-default in-app Assistant shell and a fake model adapter; defer production provider adapters until provider and data-residency decisions are recorded.
- [ ] 6.9 Add compatibility tests that compare legacy Python MCP results with Go application read services over the same collection and deck fixtures.

## 7. Migration, Packaging, And Release Verification

- [ ] 7.1 Add telemetry-free structured local logs for sync and tool calls plus a redacted diagnostic-bundle exporter with explicit opt-in fields.
- [ ] 7.2 Add migration backup, forward migration, and export-based recovery tests for the local database.
- [ ] 7.3 Run fixture, unit, contract, race, frontend, accessibility, and end-to-end suites in CI without MTGA, network, administrator privileges, or an LLM.
- [ ] 7.4 Build and smoke-test the Go desktop on a clean Windows environment, including offline startup and WebView dependency handling, and prove no development MCP is packaged.
- [ ] 7.5 Execute a multi-snapshot parity trial between accepted Detailed Logs, the Python exporter, and Go normalization before enabling log sync by default.
- [ ] 7.6 Update user documentation with source trust, privacy, permissions, recovery, compatibility, and clear statements that the companion does not automate gameplay.
- [ ] 7.7 Define and verify the release gate for retiring the Python GUI while retaining a time-bounded rollback package and legacy JSON compatibility.
