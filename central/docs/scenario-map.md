# Central scenario map

OpenSpec `add-central-go-platform`, task 8.6: every `#### Scenario:` in
the change's five specs mapped to its automated test or named
operations drill. Paths are relative to `central/internal`. The change
validates in strict mode (`openspec validate add-central-go-platform
--strict`) — CI evidence is the green `go-modules` run on every push.

## catalog-publication

| Scenario | Automation / drill |
|---|---|
| Provider is disabled during retry | `app/providers` `TestGuardStopsDisabledProvidersMidRun` |
| Stored object hash differs | `adapters/objectstore` `TestFailedVerificationCleansUpAndErrors` |
| Activation transaction fails | `postgres` `TestActivationIsAtomicWithPreviousFallback` (container, CI) |
| Signature is unknown | `adapters/signing` verifier tests; tamper rejection in `app/publication` `TestSignedFixtureCatalogRoundTrip` |
| Database stage slows down | `app/publication` `TestBackpressureBoundsInFlightObservations` |
| Two workers start the same publication | `app/publication` `TestConcurrentRunsStayIndependent`; job-level ownership via `worker` reclaim fencing |

## central-api-contracts

| Scenario | Automation / drill |
|---|---|
| Client sends an unknown field | `httpapi` `TestDecodeStrictRefusesUnknownFields`, `FuzzDecodeStrict` |
| Decompressed payload exceeds its cap | `httpapi` `TestGzipBombIsRefused` |
| Client disconnects | `httpapi` `TestClientCancellationReachesTheHandler` |
| Completed request is retried | `httpapi` `TestTelemetryScopesBatchesToThePrincipal` (idempotent replay); `postgres` `TestTelemetryIngestionIsAtomicAndIdempotent` |
| Installation credential calls operations | `httpapi` `TestOpsPlaneIsSeparatelyAuthenticated` |
| Client already has the current manifest | `httpapi` `TestDocHandlerServesWithETagAndRevalidates` (304) |
| Development route is requested | `httpapi` `TestOpenAPIContractMatchesTheRouteInventory` (closed route inventory — nothing outside the contract mounts) |

## central-data-governance

| Scenario | Automation / drill |
|---|---|
| Telemetry role reads catalog governance | `postgres` roles grant/denial matrix (`roles_suite_test.go`, container) |
| API starts with unsupported schema | `postgres` migrator preflight ("ahead of this binary") + `SchemaReady` readiness refusal |
| Installation reads another installation | `postgres` RLS suites (`rls_test.go`, `rls_concurrency_test.go`, container) |
| Repository is scanned | CI `go-modules`: staticcheck + govulncheck on every push; secret scan is part of task 8.4 (open) |
| Deletion reaches a backup | operations drill: restore must reapply tombstones (`docs/data-inventory.md`, restore ordering); drill execution is task 7.6 (open) |
| Restore drill runs | operations drill, task 7.6 (open — needs an isolated environment) |

## central-operations

| Scenario | Automation / drill |
|---|---|
| Worker dies during a job | `worker` `TestWorkerDeathLeaseExpiryAndDuplicateEffectFencing` (container, CI) |
| Downstream consumer stops | `platform/concurrency` `TestPollBlockedSenderStaysBounded` |
| Pool budget is exhausted | `worker` `TestPollerRefusesWorkersBeyondTheDatabaseBudget`; `httpapi` readiness `pool` probe |
| Queue age exceeds its safety threshold | `httpapi` readiness `queue` probe (`TestReadinessNamesFailingCapabilitiesOnly`) |
| Shutdown deadline expires | `app/lifecycle` `TestFailingStageNeverStopsPoolClosure` (stages continue, pools always close) |
| Operation fails | `docs/runbooks.md` (six incident runbooks, detection→evidence) |
| Burst test passes but soak fails | open: 24h soak needs a long-lived environment (tasks 6.6/8.3) |

## consented-telemetry

| Scenario | Automation / drill |
|---|---|
| Client has not consented | `domain/consent` `TestAdmitDemandsCurrentConsentAndDeclaredAttrs` (missing/stale versions refused) |
| Batch contains a forbidden field | `domain/consent` `TestScreenRefusesDisallowedContentShapes`, `TestAdmitScreensDeclaredStringAttrs`; `httpapi` `TestTelemetryValidationIsAtomic` |
| Batch is delivered twice | `postgres` `TestTelemetryIngestionIsAtomicAndIdempotent` (container, CI) |
| Deletion completes | `postgres` `TestDeletionRequiresTheDeletionSecret` (container, CI) |
| Retention job runs | `postgres` `TestAggregationIsAtomicAndSuppressesDuplicates` (expired events purged, never mined) |
| Admission closes | `httpapi` `TestTelemetryAdmissionControlRefusesFloods` |

Open items are exactly the environment-bound ones: the restore drill
(7.6), secret/container scanning (8.4) and the 24-hour soak (6.6/8.3).
Everything else runs on every push.
