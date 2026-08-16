# Stage 1 load evidence (CI scale)

OpenSpec `add-central-go-platform`, task 8.1. The measured half runs
per push in `internal/loadtest` against the real HTTP handlers and a
real PostgreSQL container; this note pins the assumptions and the
extrapolation.

## 10k-DAU assumptions

- **Manifest**: each client revalidates the current manifest ~4x/day
  (launch + periodic). 10k DAU → ~40k req/day ≈ **0.46 req/s steady**,
  ~5 req/s at a 10x burst. ETag/304 and the single-entry cache mean
  most hits never touch the database.
- **Telemetry**: opted-in clients send ~2 batches/day. At 100% opt-in
  (upper bound) → ~20k batches/day ≈ **0.23 req/s steady**, ~2.5 req/s
  burst. Each batch is one transaction (batch row + events).

## What CI measures

The suite compresses far past those rates: 2000 manifest reads over 16
workers and 400 unique telemetry batches over 8 workers, back to back.
Budgets asserted per push: manifest p95 ≤ 100ms, telemetry p95 ≤ 250ms,
zero errors, pool never above its 8-connection budget. The evidence
line (p50/p95/p99 per path, pool totals, goroutines, GC pause total as
the CPU proxy) is logged in the job output of every green run.

## Cost extrapolation

At the steady rates above, the platform fits comfortably in the
smallest common tiers: one 2-vCPU container for the API (the CI run
sustains hundreds of req/s in-process; 5 req/s of burst is noise) and
a 2-vCPU / 4GB managed PostgreSQL with the 60-connection capacity
already modeled in `Capacity` (worker + API replicas budgeted). Object
storage for catalogs is a few GB of immutable snapshots. Order of
magnitude: tens of dollars/month, dominated by the managed database.

## Honest limits

CI-scale is not staging-scale: the runner shares CPU, the network is
loopback, and disk I/O is the runner's. The full-scale Stage 1 run on
a deployed environment — and everything beyond (Stage 2 burst, soak) —
stays with tasks 8.2/8.3/8.5.
