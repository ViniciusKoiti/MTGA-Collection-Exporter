# Go Backend Decision Matrix

## Recommended Baseline

| Concern | Initial decision | Why | Revisit when |
| --- | --- | --- | --- |
| Deployment | Modular monolith; separate API and worker binaries | Simple operations and transactions | Ownership or scaling diverges |
| Repository | Separate `central` Go module | Independent desktop/server releases | Shared schemas need generated package |
| HTTP | Standard `net/http` plus OpenAPI contracts | Small dependency and explicit behavior | Routing complexity is measured |
| PostgreSQL | `pgxpool` plus `sqlc` | Native features, typed owned SQL | Never for ORM preference alone |
| Jobs | PostgreSQL table, leases and outbox | Durable without new infrastructure | Fan-out or throughput misses gates |
| Catalog | Object storage/CDN with signed manifests | Scalable downloads and integrity | Multi-region publication is required |
| Cache | CDN plus bounded process cache | Avoid Redis failure mode | Shared cache hit rate justifies it |
| Telemetry | Opt-in pseudonymous batches | No account or collection upload | Product requires authenticated sync |
| Observability | `slog` and OpenTelemetry contracts | Structured and vendor-neutral | Choose backend by operations cost |

## Goroutine Rules

Every goroutine must have:

1. A parent `context.Context` and cancellation path.
2. One named owner responsible for waiting for it.
3. A configured concurrency or downstream capacity limit.
4. A result or error path; no fire-and-forget work.
5. Defined behavior during graceful shutdown.

Use goroutines for independent bounded I/O, catalog pipeline stages, and concurrency
across graph runs. Keep one graph run sequential by default. Do not create a goroutine
per card, deck, telemetry event, SQL statement, or HTTP response continuation.

Channels transfer ownership between pipeline stages. They must be bounded, and every
send must be cancellable. Prefer a direct call for synchronous work and a mutex for
small shared state when a channel would only hide ownership.

## Backpressure Order

```text
edge rate limit
  -> HTTP admission limit
  -> pgxpool acquisition limit
  -> durable PostgreSQL job
  -> bounded worker group
  -> provider / object-store limit
```

When capacity is exhausted, reject or persist work with retry guidance. Never solve
saturation by increasing goroutines beyond the database and provider budgets.

## Development Sequence

1. Define schemas, limits, idempotency and errors.
2. Build PostgreSQL roles, migrations and repositories.
3. Implement synchronous HTTP contracts with timeouts.
4. Add durable jobs and sequential workers.
5. Add bounded concurrency and deterministic reducers.
6. Add load, race, leak, cancellation and soak tests.
7. Tune from profiles; introduce new infrastructure only from measured evidence.

## Decisions Still Needed

- Cloud and object-storage provider.
- Approved source for the first meta catalog.
- Operations OIDC provider.
- Raw telemetry retention justified by product need and privacy review.
- Windows code-signing provider and protected key custody.

References: [Go pipelines](https://go.dev/blog/pipelines),
[context](https://pkg.go.dev/context),
[errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup),
[pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool), and
[sqlc](https://docs.sqlc.dev/en/latest/).
