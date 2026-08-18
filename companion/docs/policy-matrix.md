# Integration policy matrix

Task 1.5 of OpenSpec `introduce-agentic-go-companion`. Every operation the
companion can perform falls into exactly one tier. The matrix is not
aspirational: each row cites the code and test that enforce it. Anything
not listed is **denied by default** (`policy.Classifier` returns
`prohibited` for unknown tools).

## Tier 1 — Allowed read operations (no approval)

Read-only over local state; never touch MTGA, the network or the disk
beyond the local database. Enforced by `internal/policy` (read list) and
served through `internal/application/readtools` under the registry
budgets (correlation, strict args, pagination).

| Operation | Backing code |
| --- | --- |
| `sync-status` | `readtools.syncStatus` over `SnapshotStore.Latest` |
| `collection-summary` | `readtools.summary` |
| `search-cards` | `readtools.search` (resolved entries only) |
| `card-lookup` | `readtools.lookup` via `Catalog` |
| `check-deck` | `readtools.checkDeck` (`CompareOwnership` + `Standard.Validate`) |
| `query-collection`, `query-stats` | pure-query exemptions in `activity.Inventario` |

## Tier 2 — Approval-required local effects

Local side effects. Assistant requests are proposal-only: an immutable
preview plus an expiring single-use token bound to
`sha256(operation + target + exact payload)`; execution redeems the token
(`agentgate`, `proposals.Authorize`, `exportsvc`). Direct user commands
run immediately — user intent is the approval. Every path is audited with
hashes only (`sqlitestore` audit refuses non-redacted records).

| Operation | Effect | Backing code |
| --- | --- | --- |
| `sync-collection` | commit a new immutable snapshot | `collection-sync` graph via `activity.Launcher` |
| `save-deck` | persist a deck revision | `decksvc.SaveRevision` + `proposals.ProposeSaveDeck` |
| `export-file` | write compatibility export to disk | `compatibility` writer + `proposals.ProposeFileExport` |
| `export-clipboard` / `copy-export` | copy export to clipboard | `exportsvc` + `proposals.ProposeClipboardExport` |
| `export-approved` | approved export graph | `approvedexport` graph (outbox idempotente) |

## Tier 3 — Prohibited MTGA control operations (never executable)

Absent by construction, not by configuration: `toolreg.Register` refuses
these capability names, invocation is denied, and the default-deny policy
classifies them prohibited — three independent layers, proven by the
negative suite in `toolreg/forbidden_test.go` (task 6.6).

| Capability | Examples proven denied |
| --- | --- |
| Process memory access | `memory-scan`, `read-process-memory` |
| Synthetic input / gameplay | `input-synthesis`, `send-input`, `gameplay-automation`, `play-card` |
| Purchases | `purchase-pack`, `buy-gems` |
| Account / credentials | `account-manage`, `credential-read`, `password-export` |
| Shell / filesystem | `shell`, `exec-command`, `filesystem-browse`, `file-write-anywhere` |

The legacy Python scanner (process memory) survives only as the isolated
explicit compatibility command (task 3.9) and is excluded from agent
tools by this matrix.
