# Repository Guidelines

## Project Structure & Module Organization

This repository is a Python MTG Arena collection exporter. The reusable core lives in `mtga/`: `scanner.py` reads MTGA memory, `database.py` builds card metadata, `collection.py` enriches and exports results, and `pipeline.py` orchestrates the full scan. User-facing entry points are `run_gui.py` for the desktop app and `run_mcp.py` for the MCP server. GUI code is isolated under `mtga/gui/`. Generated local files such as `config.json`, `arena_id_lookup.json`, and `mtga_collection.{json,csv,txt}` should not be treated as source.

## Build, Test, and Development Commands

Install dependencies:

```bash
pip install -r requirements.txt
```

Run the desktop app:

```bash
python run_gui.py
```

Run the MCP server over stdio:

```bash
python run_mcp.py
```

Build a Windows GUI executable:

```bash
pip install pyinstaller
pyinstaller --onefile --windowed --name "MTGA Exporter" run_gui.py
```

The live scan depends on Windows, `pymem`, and MTG Arena running on the Decks screen. Use administrator privileges if memory access fails.

## Coding Style & Naming Conventions

Use standard Python style with 4-space indentation, type hints where practical, and small modules with clear ownership. Keep public functions in `snake_case`; classes and dataclasses use `PascalCase`. Follow the existing pattern of short module docstrings and concise comments for non-obvious behavior. Prefer `pathlib.Path` for filesystem paths and keep shared logic in `mtga/` rather than duplicating it in GUI or MCP layers.

## Testing Guidelines

No automated test suite is currently present. When adding tests, place them under `tests/` and use `pytest` naming such as `test_pipeline.py` and `test_run_full_scan_exports_collection`. Mock MTGA memory access, filesystem writes, and network calls to Scryfall. At minimum, manually verify `python run_gui.py` starts and `python run_mcp.py` imports/runs before submitting changes.

## Commit & Pull Request Guidelines

This checkout has no `.git` history, so no project-specific commit convention can be inferred. Use concise imperative commit subjects, for example `Add collection export validation`. Pull requests should include a short summary, manual test results, linked issues when applicable, and screenshots for GUI changes. Call out Windows-only behavior, generated files, or changes that affect MCP tool responses.

## Security & Configuration Tips

Do not commit personal `config.json`, exported collections, cache files, virtual environments, `node_modules/`, or `__pycache__/`. Treat exported collections as user data. Keep MCP output deterministic and avoid exposing local paths unless the tool explicitly needs them.
