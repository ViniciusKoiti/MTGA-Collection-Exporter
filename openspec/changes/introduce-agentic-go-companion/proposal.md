## Why

The current application proves the collection-export workflow, but its primary acquisition path depends on fragile process-memory scanning and manual calibration anchors. A safe, local-first companion should synchronize collection data through user-controlled sources, expose trustworthy provenance and freshness, and let an AI assist through constrained tools without controlling MTG Arena gameplay.

## What Changes

- Add a Go companion foundation alongside the existing Python application; preserve the current JSON export contract during migration.
- Prefer MTG Arena Detailed Logs and explicit file imports as collection sources, with source status, freshness, confidence, and actionable recovery guidance.
- Introduce a task-oriented desktop flow for setup, collection sync, deck work, and assistant conversations.
- Add deterministic collection and deck services behind stable ports, independent from GUI, storage, MTGA, Scryfall, MCP, and model providers.
- Add permissioned agent tools for collection analysis and deck planning. Tools are read-only by default; export or clipboard actions require explicit user confirmation.
- Prohibit gameplay automation, memory writes, synthetic input, credential access, and autonomous actions in MTG Arena.
- Retain the Python memory scanner only as an isolated legacy compatibility path while the log importer is validated; do not extend it with control or write capabilities.

## Capabilities

### New Capabilities

- `collection-sync`: Acquire, normalize, validate, snapshot, and report collection data from safe local sources.
- `desktop-companion`: Guide users through setup, sync health, collection exploration, deck workflows, and recoverable errors.
- `agent-tools`: Provide policy-controlled, auditable tools for collection analysis and deck assistance.
- `deck-workspace`: Create, validate, compare, and export decks against the user's owned collection and selected format.

### Modified Capabilities

None. This repository has no archived OpenSpec capabilities yet.

## Impact

- Adds a Go module and desktop shell while keeping Python entry points operational during the strangler migration.
- Introduces application/domain contracts, source adapters, and local snapshot storage while preserving the current Python MCP only as a legacy JSON consumer during migration.
- Changes collection records to carry stable identifiers, source provenance, observed time, and validation state while continuing to emit the existing `mtga_collection.json` shape for compatibility.
- Requires Windows integration for process presence and log discovery, Scryfall bulk-data caching, and explicit privacy/permission settings for agent features.
- Adds contract, fixture, migration, policy, and end-to-end tests; no live MTGA process or network access is required in the default test suite.
- Excludes development MCP tooling from the product binaries; that tooling is specified by `add-graph-workflow-harness`.
