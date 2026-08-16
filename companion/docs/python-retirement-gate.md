# Python GUI retirement gate

OpenSpec `introduce-agentic-go-companion`, task 7.7: the criteria that
must ALL hold before the CustomTkinter GUI stops shipping, each mapped
to its verification. The gate is evaluated at release time; nothing
retires implicitly.

## Criteria

| # | Criterion | Verification | Status |
|---|---|---|---|
| 1 | Legacy JSON compatibility frozen | `adapters/legacyjson` frozen-contract suite + `compat`/`compatibility` exporters (JSON/CSV/TXT projections), green per push | automated, green |
| 2 | Python parity proven | `internal/parity` golden comparisons against the Python exporter (`python_stats`, `python_check_deck`) | automated, green |
| 3 | Multi-snapshot parity trial | task 7.5: accepted Detailed Logs vs Python exporter vs Go normalization across several snapshots | **pending** (7.5) |
| 4 | Desktop shell feature-complete for the export path | first-run setup, import fallback, Collection view, export preview — sections 4-5 suites per push | automated, green |
| 5 | Clean-machine desktop run | task 7.4: fresh Windows VM, offline startup, WebView2 bootstrap, no dev MCP packaged | **pending** (7.4) |
| 6 | Rollback package retained | the Python zip (`MTGA-Exporter-windows.zip`) keeps building and attaching to releases for **two release cycles after retirement**; the packaging workflow and its release-artifact inspection stay in place during that window | policy, wired in CI |
| 7 | No regression window open | no open P1 issue against the Go companion export path at gate time | manual check at release |

## Procedure

1. At the release where retirement is proposed, verify criteria 1-5
   are green in that release's CI runs and record the run IDs.
2. Cut the release with BOTH packages (Go desktop + Python zip); mark
   the Python package "maintenance mode" in the release notes.
3. For the following two release cycles, keep building the Python zip
   (criterion 6). Any P1 export regression during the window reverts
   the default to the Python GUI — one release-notes line, no code
   rollback needed since both ship.
4. After the window closes with no P1 regressions, stop building the
   Python zip and archive `run_gui.py` behind a tag.

The legacy JSON format itself is NOT retired by this gate — it stays
frozen as the compatibility contract regardless of which GUI ships.
