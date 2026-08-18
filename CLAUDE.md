# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

MTGA Collection Exporter: reads the user's MTG Arena collection directly from the game's process memory (`pymem`, Windows-only, game must be open) and exports it as JSON/CSV/TXT. Two user-facing frontends share the same core: a CustomTkinter desktop app (`run_gui.py`) and an MCP server (`run_mcp.py`) that lets an AI assistant query the exported collection.

## Commands

```bash
pip install -r requirements-dev.txt   # includes runtime deps + test/lint tools

python run_gui.py                     # desktop app
python run_mcp.py                     # MCP server over stdio

pytest                                            # all tests
pytest tests/unit/test_pipeline.py                # one file
pytest tests/unit/test_pipeline.py::test_name     # one test
pytest --cov=mtga --cov-report=term-missing       # coverage (fails under 70%; mtga/gui/* omitted)
pytest tests/evals -q                             # deterministic quality/SOLID/architecture gates

ruff check mtga tests                 # lint (line length 88, rules E/F/I/B)
xenon --max-absolute B --max-modules A --max-average A mtga   # complexity gate
bandit -q -r mtga -c pyproject.toml -ll
pip-audit -r requirements.txt
```

Packaging: `.\scripts\package_windows.ps1 -Clean` builds `dist/MTGA-Exporter-windows.zip` (PyInstaller). CI (`.github/workflows/windows-package.yml`) runs ruff + the deterministic evals + pytest, builds the zip, and attaches it to a GitHub Release on `v*` tags.

On Windows use the `.venv-win` virtualenv (`.venv` is a POSIX-layout venv and won't run here).

## Architecture

Core pipeline lives in `mtga/`; GUI and MCP are thin consumers that must not duplicate its logic. `run_full_scan` in `mtga/pipeline.py` is the single orchestration point:

```
config.py  →  database.py  →  scanner.py  →  collection.py
(Config,      (card lookup    (memory scan)   (enrich + export
 paths)        {grp_id: meta})                 JSON/CSV/TXT)
```

- **Scanning is anchor-based** (`scanner.py`): the user configures "anchors" — rare cards they own with exact quantity, stored as `[grp_id, qty, name]` in `config.json`. The scanner pattern-scans MTGA.exe memory for the packed `(grp_id, qty)` pair, then heuristically extracts the surrounding card array as `{grp_id: count}`. All of this is behind a late `import pymem` so the rest of the package works on any OS.
- **Card database** (`database.py`): source is `local` (MTGA's own `.mtga` SQLite files in the Raw folder), `scryfall` (bulk API, richer metadata: colors/type/mana — needed for good MCP deck building), or `auto` (local, falling back to Scryfall). Result is cached in `arena_id_lookup.json`; delete that file to force a refresh.
- **Validation** (`validation.py`) is called by the GUI after a scan — not by `run_full_scan` — comparing against the previous export and writing `mtga_scan_validation.json` with status ok/warning/error.
- **MCP server** (`mcp_server.py`, FastMCP): tools `get_collection`, `collection_stats`, `search_cards`, `check_deck` read the previously exported `mtga_collection.json` (no game needed); only `rescan` runs the live pipeline.
- **i18n** (`i18n.py`) covers GUI texts only (pt default, en). Core, MCP responses, and export file formats must stay language-independent.
- Progress reporting uses the `ProgressCallback` convention from `config.py` (`notify(progress, i, total, msg)`), threaded through all long-running core functions so the GUI can show status.

Generated local files are user data, not source: `config.json`, `arena_id_lookup.json`, `mtga_collection.{json,csv,txt}`, `mtga_scan_validation.json`.

## Testing conventions

Tests never touch the real game, network, or user files: scanner tests use a `FakePymem`, Scryfall tests mock `requests.get`, export tests use `tmp_path`. `tests/unit` is fast/offline; `tests/performance` holds simple memory-budget tests; `tests/evals` holds deterministic AST-based gates for code quality (complexity, function/module size, no `print`), SOLID proxies (class/method/parameter limits), and architecture (layer contract, third-party allowlist per module, lazy-only `pymem`, GUI/MCP isolation). When adding a legitimate new dependency or module, update the contract in `tests/evals/test_architecture.py` in the same commit. Quality targets and next steps are tracked in `docs/quality.md`.

## Style notes

**Language: all code and comments must be written in English** — identifiers, function/type names, docstrings, code comments, commit messages, and test names alike. This applies to every module (Python `mtga/`, Go `central/` and `companion/`). The only Portuguese allowed is in *user-facing strings* routed through the GUI i18n layer (`i18n.py`, pt default + en). Existing Portuguese identifiers/comments should be migrated to English opportunistically when a file is touched, not in big-bang renames.

Prefer `pathlib.Path`, small modules with clear ownership, and keep shared logic in `mtga/` rather than in the GUI or MCP layers (see AGENTS.md for the fuller guidelines).
