## Context

The Python application currently has a useful separation between scanning, card metadata, collection export, validation, GUI, and MCP. The main flow is still a synchronous `database -> process memory -> enrich -> export` pipeline. The desktop experience starts with configuration and calibration anchors, then exposes the result as a searchable table. The MCP server reloads a JSON export for every tool call and provides collection search, statistics, deck ownership checks, and live rescan.

The main constraints are:

- Process-memory layouts and anchor heuristics are version-sensitive and difficult to diagnose.
- Direct process-memory access and automated control create product-policy and user-account risk. The new architecture must not write memory, synthesize input, or play the game.
- Detailed Logs are a safer, user-enabled integration surface, but their current collection payload and stability must be verified with sanitized fixtures before relying on them.
- The application is Windows-first, local-first, and handles private collection and log data.
- Existing users and MCP clients depend on `mtga_collection.json`; migration cannot require an immediate rewrite.
- Deck legality, ownership, ranking, and natural-language explanation are separate concerns. Deterministic services must own facts; a model may plan and explain but cannot invent tool results.

## Goals / Non-Goals

**Goals:**

- Deliver a first vertical slice in Go without stopping the current Python application.
- Replace configuration-first scanning with a guided, observable sync flow.
- Define stable domain and application contracts that do not depend on Wails, SQLite, Scryfall, MTGA logs, MCP, or a model provider.
- Make agent actions least-privilege, schema-validated, auditable, and explicitly confirmed when they cause a side effect.
- Preserve the current export format while introducing richer internal snapshots.
- Make all default tests deterministic and runnable without MTG Arena, administrator privileges, network access, or an LLM.

**Non-Goals:**

- Automating matches, drafting, purchases, navigation, clicks, or keyboard input in MTG Arena.
- Writing to MTGA memory, modifying game files, bypassing client protections, or reverse engineering network protocols.
- Building a complete Magic rules engine or a competitive deck-ranking model from scratch.
- Removing the Python application before Go reaches fixture, contract, UX, and packaging parity.
- Sending raw logs, local paths, credentials, or a complete collection to a remote model by default.

## Decisions

### 1. Use a strangler migration, not a rewrite cutover

Create the Go companion beside the Python application. The first Go source adapter reads the existing JSON export, which gives immediate domain, storage, legacy-consumer, and UI feedback without changing acquisition. A Detailed Logs adapter follows after fixture validation. Python remains the rollback path until parity criteria pass.

Alternatives considered:

- A full rewrite has a simpler final tree but combines acquisition, product, packaging, and agent risks into one release.
- Embedding Python in Go preserves code reuse but complicates packaging and hides the boundary that needs to become explicit.

### 2. Build a modular monolith with ports and adapters

Use one Go module for the desktop product. Development tooling is built separately by `add-graph-workflow-harness` and is never a product dependency. Keep dependencies pointing inward:

```text
cmd/companion/              Wails desktop entry point
internal/domain/            Card, collection snapshot, deck, policy types
internal/application/       Sync, query, deck, export, and agent use cases
internal/ports/             Source, catalog, store, clock, clipboard, model ports
internal/adapters/source/   JSON import, MTGA log watcher, legacy bridge
internal/adapters/catalog/  Scryfall bulk-data cache
internal/adapters/store/    SQLite and compatibility JSON export
internal/adapters/windows/  Process presence, paths, clipboard, notifications
internal/policy/            Tool risk classification and approval decisions
frontend/                   Desktop presentation only
```

Domain packages MUST NOT import adapter or presentation packages. Application services receive interfaces and return typed results with stable error codes.

### 3. Model acquisition as observations and snapshots

`CollectionSource` returns an observation with source ID, observed time, raw identity, diagnostics, and quantities. The sync service normalizes it through the card catalog and commits an immutable `CollectionSnapshot` transactionally.

Each snapshot includes:

- stable snapshot ID and schema version;
- source kind and source instance;
- observed and imported timestamps;
- card printing identifiers plus Arena and Oracle identifiers when known;
- quantity, normalization diagnostics, confidence, and freshness state.

Unknown cards remain visible as unresolved records instead of being silently discarded. Compatibility export projects the latest valid snapshot into the current JSON shape.

### 4. Prefer user-controlled sources

Source priority is:

1. Detailed Logs watcher after the user enables it and the parser recognizes a supported payload.
2. Explicit import of the current JSON export or another documented file format.
3. Isolated legacy Python scanner launched only through a clearly labeled compatibility action.

The companion may detect whether `MTGA.exe` is running and whether logs are changing, but detection MUST NOT grant process-memory access or control the client. The legacy scanner is not extended and is excluded from the new agent tool surface.

### 5. Persist internal state in SQLite and keep exports at the edge

Use `database/sql` behind a repository adapter with embedded migrations. Store snapshots, cards, saved decks, approvals, and agent audit events locally. The exact SQLite driver is selected only after a Windows single-file packaging spike; adapter tests must run against the selected driver.

JSON, CSV, Arena deck text, and the legacy collection JSON are export projections, not the source of truth. Writes use a temporary file plus atomic replace where the platform supports it.

### 6. Use Wails for the new desktop shell

Use the stable Wails release line with a TypeScript frontend. Wails keeps Go as the backend while providing a richer, testable presentation layer through the Windows WebView. Do not adopt a beta release in the first production slice.

The primary navigation is `Home`, `Collection`, `Decks`, `Assistant`, and `Settings`:

- First launch opens a short setup flow: detect logs, explain how to enable Detailed Logs, choose privacy settings, then verify a source.
- Home is operational: MTGA presence, last sync, source health, collection delta, and the next clear action.
- Collection supports search, compact filters, sortable columns, card details, unresolved items, and snapshot comparison.
- Decks combines deck list, ownership/legality results, substitutions, and Arena-format export in one workspace.
- Assistant shows its active scope, tool calls, evidence, approvals, and results; it is not the only way to use collection or deck features.

The UI uses keyboard-accessible semantic controls, visible focus, text labels for statuses, and error recovery next to the failed step. Color and mana symbols are supplemental, never the sole carrier of meaning.

### 7. Treat the agent as a client of application tools

The product agent layer consists of a typed tool registry, policy engine, approval service, audit log, and optional model client. The in-app assistant calls the same application use cases as the deterministic UI. A separate development MCP may call test and inspection ports but is not part of this product runtime.

Initial tool classes are:

- Read: collection summary, collection search, card lookup, snapshot status, deck validation, and ownership gaps.
- Proposed local side effect: save a deck, create an export, copy an Arena deck list, or request a sync.
- Forbidden: process-memory access, synthetic input, gameplay actions, purchases, account changes, or credential access.

Read tools run without confirmation. A proposed side effect returns a preview and approval token; execution requires an explicit, unexpired confirmation bound to the exact arguments. Tool inputs and outputs are schema-validated, size-limited, and recorded with sensitive fields redacted.

The model never receives unrestricted filesystem, shell, clipboard, database, or process handles. Card data and log content are untrusted data, not instructions. Deterministic services calculate counts, legality, ownership, and validation; the model may select tools and explain their results.

### 8. Keep model providers optional and privacy explicit

The deterministic desktop works without a model. The in-app assistant is disabled until the user selects a provider and reviews what data may leave the device. Provider credentials use the operating system credential store rather than config files. The existing Python MCP remains a time-bounded legacy consumer of compatibility exports, not the foundation of the Go product.

The request context builder sends the smallest useful projection, preferring aggregate statistics and explicit card subsets over full snapshots. Raw logs and local paths are never model context. Each response displays whether it used local deterministic logic, an external model, or both.

### 9. Make sync and tool state observable

Represent sync as a state machine: `not_configured`, `ready`, `watching`, `syncing`, `fresh`, `stale`, and `error`. Progress events carry a stable phase, percentage when knowable, and a recoverable error code. Repeated observations with the same identity are idempotent.

Structured local logs use correlation IDs for sync runs and agent tool calls. Diagnostic bundles exclude collection contents and paths unless the user explicitly opts in.

## Risks / Trade-offs

- **Detailed Logs may stop exposing a complete collection** -> Make log support fixture-driven and capability-detected; retain explicit import and the compatibility bridge.
- **A broad Go migration can stall before user value appears** -> Ship vertical slices: import and status first, then logs, deck workspace, and optional assistant.
- **Two implementations can drift** -> Define golden JSON fixtures and contract tests consumed by Python and Go; track parity before retiring Python.
- **LLM output can be incorrect or manipulated by data** -> Keep facts in deterministic tools, isolate untrusted content, validate schemas, show tool evidence, and require confirmation for effects.
- **SQLite or WebView dependencies can complicate packaging** -> Run a signed Windows packaging and clean-machine smoke-test spike before committing to the driver and frontend build chain.
- **Format legality and card data change over time** -> Version catalog snapshots, record update time, and report stale legality rather than presenting it as current.
- **Legacy memory scanning carries policy and stability risk** -> Label it unsupported compatibility, keep it out of the agent, collect no new capabilities around it, and remove it after a proven replacement exists.

## Migration Plan

1. Capture sanitized golden fixtures for the current JSON export and current Detailed Logs; confirm whether logs contain a complete, stable collection payload.
2. Add the Go module, domain types, application ports, JSON import adapter, in-memory repositories, and contract tests.
3. Add snapshot persistence and compatibility export. Verify byte-semantic parity for fields consumed by existing MCP clients.
4. Add the Wails shell with onboarding, sync health, and collection views using imported snapshots.
5. Add the log watcher behind a feature flag. Compare its snapshot against the Python export over multiple real collections before enabling it by default.
6. Add deck services and verify that the legacy Python MCP can still consume the compatibility export; do not add a product Go MCP.
7. Add the optional in-app assistant, approval flow, audit view, and provider privacy controls.
8. Package and smoke-test on a clean Windows machine. Keep the Python release downloadable until one stable release meets parity and recovery criteria.

Rollback is per adapter and entry point: disable log watching or the new desktop binary and continue using the Python exporter and existing JSON contract. Snapshot migrations are forward-only, with pre-migration backup and export-based recovery.

## Open Questions

- Does the current MTG Arena Detailed Logs payload contain the entire collection reliably, or only changes/session data?
- Which Arena formats should follow the first versioned Standard ruleset, and in what order?
- Should the first release include an embedded model client, or ship deterministic assistant tools without a provider first?
- Which model providers and data-residency constraints are acceptable to the intended users?
- Is the existing `image.png` a product reference for an overlay flow? It is too cropped to establish a visual target and is not used as design evidence.

## External Constraints

- Wizards documents Detailed Logs for community-built tools: https://mtgarena-support.wizards.com/hc/en-us/articles/360000726823-Creating-Log-Files-on-PC-Mac-Steam
- Wizards terms restrict unauthorized data mining, connections, reverse engineering, and bots: https://company.wizards.com/en/legal/terms
- OpenSpec change artifacts remain editable Markdown and archive into living specifications: https://github.com/Fission-AI/OpenSpec/blob/main/docs/overview.md
- Development-only MCP architecture is specified by `add-graph-workflow-harness`.
