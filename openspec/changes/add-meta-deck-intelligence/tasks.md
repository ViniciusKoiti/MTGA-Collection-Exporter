## 1. Contracts And Provider Governance

- [ ] 1.1 Define meta observation, provider approval, normalized deck, archetype, snapshot, manifest, and quality diagnostic schemas.
- [ ] 1.2 Define format, BO1/BO3, source window, sample context, provenance, usage-rights, and freshness enums with compatibility tests.
- [ ] 1.3 Create an approved-provider registry with rate, attribution, disablement, and fixture-license requirements.
- [ ] 1.4 Add sanitized provider fixtures for valid, partial, conflicting, malformed, oversized, and schema-changed observations.

## 2. Meta Catalog Pipeline

- [ ] 2.1 Implement one approved provider adapter behind the central provider port with timeouts, bounded retries, and rate policy.
- [ ] 2.2 Implement validation, quarantine, normalization, raw hash, and stable external identity handling.
- [ ] 2.3 Implement deterministic archetype and variant grouping while retaining provider-level evidence.
- [ ] 2.4 Publish immutable compressed snapshots and manifests with schema, hash, size, source set, rulesets, and freshness.
- [ ] 2.5 Add catalog worker graph scenarios for success, provider disablement, conflict, partial data, retry, and rollback.

## 3. Card Identity Resolution

- [ ] 3.1 Add exact printing, Arena, Oracle, face-aware playable, and preferred export identity contracts.
- [ ] 3.2 Implement versioned mapping from printing to playable identity with explicit ambiguity and unresolved diagnostics.
- [ ] 3.3 Implement ruleset-aware legal-printing allocation with owned preference and stable tie-breaking.
- [ ] 3.4 Implement explicit basic-land, rebalanced-variant, multiface, sideboard, and exceptional-equivalence policies.
- [ ] 3.5 Add golden tests for mixed reprints, illegal reprints, ambiguous names, partial ownership, faces, and basic lands.

## 4. Buildability And Ranking

- [ ] 4.1 Define recommendation evidence, rarity vector, confidence, class, threshold, budget, and ranking-policy versions.
- [ ] 4.2 Implement per-deck allocation, legality, unresolved, missing-copy, sideboard, and wildcard-relevant calculations.
- [ ] 4.3 Implement `ready`, `nearly-buildable`, `within-budget`, and `not-buildable` classification with user settings.
- [ ] 4.4 Implement lexicographic fact-vector ordering and stable deck-ID tie-breaking without opaque model scores.
- [ ] 4.5 Add tests proving required wildcards are reported without inferring an unavailable wildcard balance.
- [ ] 4.6 Add stale, partial, unsupported-ruleset, low-confidence, and contradictory-model evidence tests.

## 5. Local Matching And Caching

- [ ] 5.1 Add hash-verified meta snapshot download, atomic activation, last-valid fallback, and bounded local retention.
- [ ] 5.2 Build immutable collection, identity, deck, and ruleset indexes keyed by snapshot versions.
- [ ] 5.3 Implement bounded cancellable parallel matching with deterministic partition merge and no goroutine leaks.
- [ ] 5.4 Cache results by collection, meta catalog, card catalog, ruleset, and ranking-policy versions.
- [ ] 5.5 Meet the 10,000-deck p95 and memory gates on the documented Windows reference machine.

## 6. Product And Graph Integration

- [ ] 6.1 Complete the `meta-deck-recommendation` graph with catalog, allocation, ranking, evidence, cancellation, and stale paths.
- [ ] 6.2 Add Decks views for ready, nearly buildable, within-budget, and all results with fact-vector filters.
- [ ] 6.3 Add deck detail evidence for source, observed period, format, mode, freshness, allocation, gaps, and confidence.
- [ ] 6.4 Add optional bounded explanation using deterministic evidence and conflict labeling.
- [ ] 6.5 Add frontend contract, accessibility, responsive, and Playwright scenarios for recommendation workflows.

## 7. Release Verification

- [ ] 7.1 Run provider, identity, allocation, ranking, deterministic parallelism, offline, and performance suites in CI.
- [ ] 7.2 Map every spec scenario to an automated test and validate this change in strict mode.
- [ ] 7.3 Document enabled providers, attribution, limitations, freshness, wildcard semantics, and disablement procedure.
- [ ] 7.4 Gate activation behind one validated Standard snapshot and retain rollback to user-supplied deck analysis.
