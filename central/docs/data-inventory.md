# Central data inventory

OpenSpec `add-central-go-platform`, task 7.4. One entry per table in
`internal/postgres/migrations`. Classifications: **public** (served to
anyone), **internal** (operational, no personal data), **pseudonymous**
(keyed by installation ID, no direct identity), **secret-hash** (only
SHA-256 of credentials — raw values never stored).

Global rules:

- Backups are encrypted; the restore drill (task 7.6) must replay
  migrations, reapply deletion tombstones and pass smoke checks before
  a restore is declared valid.
- A legal hold suspends purges for the named installation IDs only;
  everything else keeps its retention clock.
- Deletion requests revoke the installation immediately and purge its
  pseudonymous rows; completion evidence is non-identifying.

| Table | Purpose | Classification | Retention | Deletion path | Backup | Owner |
|---|---|---|---|---|---|---|
| `sources` | Approved catalog providers and rights notes | internal | while approved + 1y | manual, ops review | daily | platform |
| `snapshots` | Published catalog snapshots per source | public | last 10 per source | prune job beyond window | daily | platform |
| `artifacts` | Immutable objects, signatures, current flag | public | as `snapshots` | prune with parent snapshot | daily | platform |
| `installations` | Enrollment identity: credential + deletion hashes | secret-hash | until deletion request | `RequestDeletion` → revoke + purge after 30d | daily | privacy |
| `consent_receipts` | Proof of purpose-versioned consent | pseudonymous | 5y after revocation (legal proof) | kept as tombstoned proof, ID pseudonymized | daily | privacy |
| `telemetry_batches` | Idempotency ledger of accepted batches | pseudonymous | 90d | purge with installation deletion | daily | privacy |
| `accepted_events` | Short-lived events awaiting aggregation | pseudonymous | `expires_at` (≤ 7d) | row TTL purge job; deletion purges early | none (ephemeral) | privacy |
| `aggregates` | Deterministic non-identifying rollups | internal | 2y | none needed (aggregate-only) | daily | product |
| `deletion_requests` | Deletion ledger + completion proof | pseudonymous | 5y (evidence) | proof is non-identifying by design | daily | privacy |
| `jobs` | Durable job queue with leases | internal | terminal jobs 30d | prune job | none (rebuildable) | platform |
| `outbox` | Transactional outbox for side effects | internal | dispatched 7d | prune job | none (rebuildable) | platform |
| `audit` | Append-only operational audit | internal | 2y | never (append-only), legal hold aware | daily | security |
| `schema_migrations` | Applied migration versions | internal | forever | never | daily | platform |

Restore ordering: `schema_migrations` first (via the migrator, never by
row copy), then parents before children (`sources` → `snapshots` →
`artifacts`; `installations` → consent/telemetry/deletions). `jobs` and
`outbox` are intentionally NOT restored — workers rebuild them, and
replaying stale jobs after a restore is a duplicate-effect risk.
