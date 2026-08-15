## Why

Users need a low-friction path from a recommended or edited deck to MTG Arena, but
automatic client control would violate the product safety boundary. A supported
handoff can reduce the workflow to one explicit application click while leaving the
official in-game import action under user control.

## What Changes

- Add an `arena-deck-handoff` graph that validates the deck, renders the exact Arena
  clipboard format, reports missing cards, and creates a stable preview.
- Treat a direct user click as approval for the exact clipboard and optional launch
  effects; agent-initiated handoff still requires an approval token.
- Copy the previewed deck atomically and optionally start the configured MTGA launcher
  without focusing, navigating, clicking, typing, importing, editing, or crafting.
- Report `prepared` and `client_started` accurately; never claim that import succeeded
  without an official observable confirmation surface.
- Deny synthetic input, process-memory writes, game-file changes, internal endpoints,
  and automatic `Craft All` actions.

## Capabilities

### New Capabilities

- `arena-deck-handoff`: Safely prepares an Arena deck import, copies it on explicit
  approval, optionally launches the client, and enforces the no-control boundary.

### Modified Capabilities

None. The active `deck-workspace` change remains the source of deck editing behavior.

## Impact

- Adds a Windows launcher port, clipboard adapter, policy rules, and graph nodes.
- Extends the Decks workspace with one primary handoff command and recoverable status.
- Reuses Arena formatting, legality, ownership, approval, and audit application services.
- Adds negative tests proving no mouse, keyboard, memory, file, account, or crafting
  automation is reachable from direct UI, assistant, graph, or development MCP paths.
