## Context

The repository currently has one Windows workflow that runs manually or for tags,
grants write permission to the whole job, and validates only the Python product. The
future Wails/TypeScript frontend must consume typed graph events, render approval and
recovery states, build on Windows, and remain testable without MTGA or an LLM.

## Goals / Non-Goals

**Goals:**

- Make every pull request prove frontend types, behavior, accessibility, and contracts.
- Validate graph-driven UI states with deterministic harness event fixtures.
- Build and inspect the actual Wails Windows binary before merge and release.
- Minimize GitHub token, action supply-chain, secret, and artifact risk.
- Produce traceable release evidence, SBOM, checksums, signature, and rollback package.

**Non-Goals:**

- Automating the live MTGA client in CI.
- Giving fork pull requests signing, deployment, provider, or production credentials.
- Treating snapshot tests or coverage percentages as substitutes for behavior tests.
- Publishing a release merely because a tag exists when required checks failed.

## Decisions

### 1. Separate workflows by trust and purpose

Use `ci.yml` for PR/push validation, `windows-wails.yml` for Windows build and smoke,
`security.yml` for CodeQL and dependency/security schedules, and `release.yml` for
protected tag/manual publication. Validation defaults to `contents: read`; only the
release publish job receives `contents: write` through a protected Environment.

Use `pull_request`, not `pull_request_target`, for untrusted code. PR workflows receive
no secrets. Pin third-party and GitHub Actions to reviewed full commit SHAs and allowlist
publishers. Add workflow concurrency that cancels obsolete runs for the same ref.

### 2. Pin the frontend toolchain and install deterministically

Use TypeScript with a pinned Node toolchain and package manager recorded in project
metadata. Prefer `pnpm` with a committed lockfile and frozen install. Caches key only
from lockfile/toolchain and never contain secrets or build outputs accepted as releases.

Run formatting check, ESLint, TypeScript no-emit, Vitest, Testing Library components,
coverage thresholds, dependency audit, and production build as separate named steps.

### 3. Generate contracts and require a clean tree

Generate Wails bindings, API clients, graph event types, error codes, and JSON schemas
from Go/OpenAPI sources. CI regenerates them and fails on a diff. Compatibility tests
load current and previous supported fixture versions. TypeScript presentation models
may wrap generated types but cannot redefine backend state or transition strings.

The frontend calls only a `CompanionClient` interface. Production uses generated Wails
bindings; tests use a deterministic adapter that emits the same serialized contracts.

### 4. Test behavior at the cheapest faithful layer

Pure transformations use unit tests. Components use Testing Library with the fake
client. Playwright drives the built web frontend for setup, collection, recommendation,
handoff, approval, cancellation, errors, offline mode, and recovery. One Windows smoke
build verifies generated bindings, embedded assets, startup, and WebView integration.

Do not attempt to drive the native WebView for every browser assertion. This keeps PR
feedback fast while preserving one real Wails boundary test on Windows.

### 5. Make graph state coverage explicit

The harness publishes sanitized event-stream fixtures for queued, running, waiting for
approval, completed, failed, cancelled, stale, offline, and recovered runs. Frontend
tests assert commands, visible state, disabled actions, duplicate-event handling,
reconnection, and terminal cleanup. The UI never requests an arbitrary transition.

### 6. Gate accessibility and responsive integrity

Run automated accessibility checks plus keyboard flows for every primary journey.
Use fixed desktop and mobile-like viewport sizes for layout and screenshots even though
the shipping shell is desktop; this catches constrained-window failures. Pin fonts,
locale, timezone, animations, and fixture data to reduce visual nondeterminism. Mask only
declared dynamic fields and review every baseline update.

### 7. Keep CI evidence private and bounded

Screenshots, traces, videos, logs, coverage, and diagnostic bundles use synthetic data.
Artifacts have short retention and upload only on failure where useful. CI scans outputs
for absolute paths, credentials, raw logs, collections, private fixtures, and dev MCP.

### 8. Treat Windows release as a promotion

The Windows job runs `wails build` with pinned dependencies and production flags, then
inspects binary/package contents, WebView policy, version metadata, licenses, and absence
of dev tools. Produce an unsigned artifact for PR validation. A protected release job
rebuilds the approved commit, generates SBOM and checksums, signs with a timestamped code
signing identity, verifies the signature, and publishes immutable artifacts.

If signing is not yet provisioned, releases remain explicitly pre-production; CI must
not silently publish an unsigned artifact as a trusted production package.

### 9. Make branch protection part of the contract

Required checks are OpenSpec strict, Python, Go/race, graph deterministic, frontend
static, frontend unit, frontend e2e/accessibility, Windows Wails build, and security.
Path optimization may skip expensive steps internally, but every required workflow
must report a terminal check to avoid bypass through path filters.

## Risks / Trade-offs

- **Windows and Playwright increase CI time** -> Parallelize independent jobs, cache
  dependencies safely, and keep one faithful native smoke layer.
- **Visual tests can be flaky** -> Pin rendering inputs and require reviewed baselines.
- **Generated contracts create repository churn** -> Commit them for auditable desktop
  builds and fail only on semantic generation differences.
- **Action SHA pins need maintenance** -> Use Dependabot review rather than mutable tags.
- **Signing credentials are high value** -> Isolate in protected Environment, release
  only, with approval and no fork exposure.
- **Coverage gates can reward weak assertions** -> Combine modest thresholds with graph
  scenario traceability and mutation/failure-focused tests.

## Migration Plan

1. Add PR validation for existing Python, OpenSpec, workflow lint, and security first.
2. Add pinned frontend toolchain and static/unit gates with the Wails scaffold.
3. Add generated contracts, harness fixtures, Playwright, accessibility, and visuals.
4. Add Windows Wails build, offline startup, artifact inspection, and clean-machine smoke.
5. Split protected release, provision signing, add SBOM/checksums, and require checks.
6. Retire the existing tag workflow only after equivalent Python packaging remains green.

## Open Questions

- Which Windows code-signing provider and custody model will be used?
- What CI duration budget is acceptable for required pull-request checks?
- Which browser engine best approximates the selected WebView2 runtime in PR tests?

## External References

- GitHub secure use and immutable action pins: https://docs.github.com/en/actions/reference/security/secure-use
- Protected branches: https://docs.github.com/en/repositories/configuring-branches-and-merges/managing-protected-branches/about-protected-branches
- Wails build lifecycle: https://wails.io/docs/guides/manual-builds/
