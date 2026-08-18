# Specification scenario map

OpenSpec `add-graph-workflow-harness`, task 8.7: every `#### Scenario:`
in the change's four specs mapped to its automated test. Package paths
are relative to `companion/internal`. Entries marked *partial* say
exactly what is still open.

## development-harness

| Scenario | Automated test |
|---|---|
| Deterministic scenario runs | `testkit` `TestMesmoCenarioMesmaSeedEvidenciaIdentica` |
| Alternate transition logic is introduced | `testkit` `TestCenarioAprovadoConcluiComTransicoesOrdenadas` (ordered-transition evidence diffs) |
| Scenario schema is unsupported | `testkit` `TestParseScenarioRejeitaFormasInvalidas` |
| Fixture contains private source data | `testkit` `TestBundleNaoContemCamposProibidos`; fixtures are synthetic by construction (`TestFixturesAreDeterministicAndSized`) |
| Scenario is repeated | `testkit` `TestMesmoCenarioMesmaSeedEvidenciaIdentica` |
| Default CI suite runs | `.github/workflows/central-go.yml` (gofmt, vet, staticcheck, build, race tests, govulncheck on every push) |
| Crash is injected after checkpoint | `testkit` `TestCrashNoCheckpointPreservaUltimoEstadoConfirmado` |
| Undeclared fault target is requested | *partial*: fault decorators wrap only declared nodes (`testkit/faults.go`); there is no dynamic fault-targeting API to misuse, but no explicit refusal test either |
| Expected event is missing | `testkit` `TestPerfilSQLiteComEntryPointDeAtividade` (`AssertEventos`) |
| Deterministic profile requests a network host | `testkit` `TestDeterministicAndIntegrationAllowNoNetwork`, `TestGuardedTransportRefusesBeforeDialing` |
| Staging profile targets production | `adapters/pgstore` `TestProductionIdentitiesAreRejectedBeforeDialing`; `devmcp` `TestClassificacaoNegaProducaoEAmbiguidade` |
| Production package is inspected | *partial*: `arch` `TestPackagingManifestsExcludeTheDevMCP`; full release-artifact inspection is task 8.6 (needs a built release) |

## development-mcp

| Scenario | Automated test |
|---|---|
| Production database is configured | `devmcp` `TestClassificacaoNegaProducaoEAmbiguidade` |
| Environment cannot be proven safe | `devmcp` `TestAutorizacaoLiberaApenasDevEStaging` |
| Valid run inspection is requested | `devmcp` `TestToolsDispatchThroughTheirSources` |
| Client requests arbitrary SQL | `devmcp` `TestAbuseMalformedArgumentsAndForbiddenSurfaces`, `TestDevMCPImportsNoNetworkShellOrSQL` |
| Read role invokes scenario execution | `devmcp` `TestScenarioExecutionIsACapabilityRestrictedToIsolatedNamespaces` |
| Scenario namespace is not isolated | same capability test (`prod/export` refused) |
| Tool response exceeds its limit | `devmcp` `TestLimitsBoundRequestResponsePaginationAndDuration` |
| Client disconnects during a scenario | `devmcp` `TestAbuseDisconnectsNeitherPanicNorHang` |

## graph-workflow-runtime

| Scenario | Automated test |
|---|---|
| Registered activity starts | `activity` `TestLauncherExecutaAtividadeDeGrafo` |
| Invalid graph is registered | `workflow` `TestRejeitaDefinicoesInvalidas` |
| Repeated call limit is reached | `workflow` `TestToolGuardLimitaTotalERepeticao` |
| Approval waits beyond the active deadline | `workflow` `TestAprovacaoExpiradaEncerraRun` |
| Planner requests an unknown transition | `workflow` `TestOutcomeSemTransicaoFalhaEstavel` |
| Effect requires approval | `workflow` `TestAprovacaoPausaERetomaComSeguranca` |
| Process stops after a committed node | `workflow` `TestRecoverRetomaDoUltimoNoConfirmado`; `adapters/sqlitestore` `TestCrashERecuperacaoFimAFim` |
| Two workers acquire one run | `adapters/sqlitestore` `TestDoisWorkersDisputamRecuperacaoSemDuplicar`, `TestLeaseDisputaEExpiracao` |
| Effect acknowledgement is lost | `workflow` `TestCrashEntreEfeitoEAckReentregaSemDuplicar` |
| Run is cancelled | `workflow` `TestCancelamentoDoContextoEncerraRun` |
| Pinned version is unavailable | `workflow` `TestRecoverVersaoIncompativelFalhaEstavel` |
| Node completes successfully | `workflow` `TestRunLinearSucedeComEvidencia` |
| Node fails before commit | `workflow` `TestCrashAntesDoEfeitoMantemPendencia`; `adapters/sqlitestore` `TestCheckpointComEfeitosEhAtomico` |

## workflow-observability

| Scenario | Automated test |
|---|---|
| Forbidden field reaches the event sink | `workflow` `TestSinkRejeitaCamposProibidos` |
| Diagnostic bundle is created | `testkit` `TestBundleDiagnosticoBateComGoldenDePrivacidade` |
| Run timeline is inspected | `adapters/sqlitestore` `TestJournalDeEventosComTimelinePaginada` |
| Unknown run is requested | run-store contract suites (`TestRunStoreSQLiteCumpreContrato`, `TestRunStoreInMemoryCumpreContrato`, `TestRunStorePostgresContractSuite`) |
| Historical run is replayed | `workflow` `TestReplayDryRunNaoTocaProducaoNemDespachaEfeitos` |
| Replay input is unavailable | `workflow` `TestReplayExigeRunTerminal` (non-terminal input refused) |
| Retention job reaches an active run | `adapters/sqlitestore` `TestRetencaoPorIdadePreservaCheckpointsAtivos` |
| Evidence storage reaches its size cap | `adapters/sqlitestore` `TestRetencaoPorTamanhoRemoveApenasJournalTerminal` |
| Metrics are exported | `workflow` `TestMetricasAgregamPorGrafoEOutcomeSemRotulosPrivados` |

No scenario needs a Windows-only manual verification: everything above
runs in the deterministic or CI (Testcontainers) tiers.
