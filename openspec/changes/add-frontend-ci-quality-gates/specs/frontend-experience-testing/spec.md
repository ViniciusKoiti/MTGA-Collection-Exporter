## ADDED Requirements

### Requirement: Primary journeys have layered tests
The frontend SHALL test pure logic, components, browser journeys, and one real Windows
Wails boundary at their cheapest faithful layers.

#### Scenario: Pull request changes a primary journey
- **WHEN** required CI runs
- **THEN** static, unit/component, Playwright, accessibility, and applicable Windows checks report independently

### Requirement: Every graph state is represented safely
The UI SHALL handle queued, running, approval-wait, completed, failed, cancelled, stale,
offline, duplicate, and recovered event streams without inventing state.

#### Scenario: Duplicate completion arrives
- **WHEN** the same committed terminal event is delivered twice
- **THEN** the UI remains completed and triggers no duplicate command or effect

#### Scenario: Connection resumes
- **WHEN** a live stream reconnects after missed events
- **THEN** the UI rebuilds from checkpoint and ordered timeline before enabling commands

### Requirement: Accessibility is a merge gate
Primary workflows SHALL be keyboard operable, expose visible focus and semantic names,
use non-color status cues, and have no configured serious accessibility violations.

#### Scenario: Keyboard-only approval flow runs
- **WHEN** Playwright completes preview, approval, cancellation, and recovery without a pointer
- **THEN** focus order, labels, state announcements, and commands remain usable

### Requirement: Constrained layouts remain coherent
The frontend SHALL pass defined minimum-window, desktop, and wide viewport checks with
no overlapping controls, clipped required text, unstable fixed-format elements, or hidden actions.

#### Scenario: Long localized status is rendered
- **WHEN** the longest supported fixture appears at the minimum window size
- **THEN** text wraps or constrains without occluding adjacent content

### Requirement: Offline and failure behavior is testable
The frontend SHALL preserve local collection and deck workflows when central services
fail and SHALL display stable recovery actions for Wails, catalog, graph, and storage errors.

#### Scenario: Central API is unavailable
- **WHEN** an offline fixture starts the application with a valid local snapshot
- **THEN** local browsing and deck work remain usable and remote freshness is identified

### Requirement: Test evidence contains no private data
Screenshots, traces, videos, logs, and reports SHALL use sanitized fixtures and SHALL
be scanned before artifact upload.

#### Scenario: Artifact contains an absolute user path
- **WHEN** evidence scanning detects a forbidden path or payload
- **THEN** upload is blocked and the job fails with a bounded diagnostic
