## Context

The deck workspace already requires Arena-format preview and explicit clipboard
approval. The remaining user journey is to prepare that list and optionally start
MTG Arena without controlling its UI. The officially documented import flow still
requires the player to use the in-game Decks and Import controls.

## Goals / Non-Goals

**Goals:**

- Make one direct application click prepare the exact deck and optionally start MTGA.
- Preserve legality, ownership, missing-card, approval, and audit evidence.
- Make every reported state truthful and every local effect recoverable.

**Non-Goals:**

- Importing, opening, editing, crafting, or selecting the deck inside MTGA.
- Synthesizing input, focusing windows, writing game memory or files, or calling
  undocumented client endpoints.
- Claiming import success from process presence or clipboard completion.

## Decisions

### 1. Add a dedicated bounded graph

`arena-deck-handoff` runs `load -> validate -> allocate -> render -> preview ->
approve -> copy -> optional-launch -> complete`. It pins deck revision, collection,
catalog, ruleset, rendered text hash, and requested effects. Retries reuse effect
idempotency keys and never recompute after approval.

### 2. Treat direct click and agent proposal differently

A direct click on the fully visible handoff command is explicit approval for the
preview currently shown. An assistant or graph proposal creates an unexecuted token
bound to the same pinned inputs. Any deck, printing, option, or text change invalidates
approval. No redundant confirmation dialog is added to a direct user command.

### 3. Launch through a narrow Windows port

`ClientLauncher` accepts no shell string. It uses an allowlisted installed-client
descriptor selected through settings and invokes the platform process API with fixed
arguments. If MTGA is running, the graph does not focus or manipulate it. Launch
failure does not undo a successful clipboard copy and returns a recoverable result.

### 4. Report handoff facts only

Terminal evidence can say `prepared`, `copied`, `already_running`, `client_started`,
or `launch_failed`. It cannot say `imported` or `deck_opened`. The UI presents missing
copies and wildcard requirements before the command and retains the deck workspace.

### 5. Enforce the boundary structurally

The launcher and clipboard are the only effect ports. Architecture and negative tests
deny mouse, keyboard, window automation, memory write, game-file mutation, account,
network interception, internal API, and crafting adapters from the graph registry.

## Risks / Trade-offs

- **One click cannot complete official import** -> Name the result accurately and leave
  the final in-game action to the player.
- **Launcher paths differ by distribution** -> Detect only supported descriptors and
  provide explicit configuration plus copy-only fallback.
- **Clipboard contents are user data** -> Write only on approval, avoid logging text,
  and store only a bounded hash in evidence.
- **Terms or client behavior can change** -> Keep launch optional and disable it without
  affecting deck export.

## Migration Plan

1. Reuse and harden Arena text golden fixtures from `deck-workspace`.
2. Implement the copy-only graph and direct/agent approval tests.
3. Add allowlisted launcher detection behind a disabled-by-default setting.
4. Add Windows clean-machine tests for installed, running, missing, and failed clients.
5. Roll back by disabling launch while preserving preview and clipboard export.

## Open Questions

- Which installed distributions can be launched through stable supported descriptors?
- Should copy-only or copy-and-launch be the default user preference?
