## 1. Safety And Contracts

- [ ] 1.1 Define pinned handoff preview, requested effect, launch descriptor, outcome, and stable error contracts.
- [ ] 1.2 Extend the policy registry with direct-click, agent-approval, copy, optional-launch, and prohibited-control rules.
- [ ] 1.3 Add architecture tests proving mouse, keyboard, window control, memory write, game-file, internal API, account, and crafting ports do not exist.
- [ ] 1.4 Add Arena text golden fixtures for owned, partially owned, sideboard, localized, reprint, malformed, and stale decks.

## 2. Handoff Graph

- [ ] 2.1 Register `arena-deck-handoff` with load, validate, allocate, render, preview, approve, copy, optional-launch, and terminal nodes.
- [ ] 2.2 Pin deck revision, collection, catalogs, ruleset, printing allocation, options, effects, and rendered hash before approval.
- [ ] 2.3 Implement direct visible-click approval and exact expiring agent approval without redundant confirmation.
- [ ] 2.4 Implement idempotency, cancellation, crash recovery, and argument-change invalidation for copy and launch effects.
- [ ] 2.5 Emit redacted evidence and truthful prepared, copied, already-running, client-started, and launch-failed outcomes.

## 3. Windows Effects

- [ ] 3.1 Implement the clipboard adapter with exact approved content, bounded size, cancellation, hash-only evidence, and tests.
- [ ] 3.2 Define allowlisted installed-client descriptors without accepting shell commands or arbitrary arguments.
- [ ] 3.3 Implement supported MTGA process presence and launch adapters without focus, navigation, or synthetic input.
- [ ] 3.4 Add copy-only fallback and independent launch retry when the client is missing, running, denied, or fails.

## 4. Frontend Experience

- [ ] 4.1 Add a primary handoff command to the deck workspace with preflight legality, missing-copy, wildcard, and effect status.
- [ ] 4.2 Add settings for copy-only versus copy-and-launch plus supported client selection and validation.
- [ ] 4.3 Map graph progress, approval, cancellation, copied, running, launch failure, and retry states to typed frontend views.
- [ ] 4.4 Add keyboard, screen-reader, responsive, and duplicate-click tests for the complete handoff flow.

## 5. Release Verification

- [ ] 5.1 Run deterministic graph tests for direct, agent, changed-preview, retry, crash, cancellation, and prohibited requests.
- [ ] 5.2 Run Windows integration tests for installed, running, missing, invalid, and failed launch descriptors.
- [ ] 5.3 Inspect the package and runtime registry to prove no automation or development-only primitive is shipped.
- [ ] 5.4 Document that the official in-game import remains a user action and validate this change in strict mode.
