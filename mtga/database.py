"""
Carregamento do banco de cartas: arquivos SQLite locais do MTGA ou API Scryfall,
com cache em arena_id_lookup.json.
"""

from __future__ import annotations

import json
import sqlite3
from contextlib import closing
from pathlib import Path
from typing import Optional

import requests

from .config import LOOKUP_FILE, Config, ProgressCallback, notify

_RARITY_MAP = {0: "", 1: "basic", 2: "common", 3: "uncommon", 4: "rare", 5: "mythic"}
_COLOR_MAP = {"1": "W", "2": "U", "3": "B", "4": "R", "5": "G"}
_CARD_METADATA_COLUMNS = ("ExpansionCode", "CollectorNumber", "Rarity", "Colors")


def _map_rarity(value) -> str:
    if isinstance(value, int):
        return _RARITY_MAP.get(value, "")
    if isinstance(value, str):
        return value.lower()
    return ""


def _parse_colors(value) -> list:
    if not value:
        return []
    return [
        _COLOR_MAP[item.strip()]
        for item in str(value).split(",")
        if item.strip() in _COLOR_MAP
    ]


def _read_localizations(cursor, tables: set) -> dict:
    loc_map: dict = {}
    localization_queries = []
    if "Localizations_enUS" in tables:
        localization_queries.append("SELECT LocId, Loc FROM Localizations_enUS")
    if "Localizations" in tables:
        localization_queries.extend(
            [
                "SELECT Id, Text FROM Localizations "
                "WHERE Format LIKE '%en-US%' OR Format IS NULL",
                "SELECT Id, Text FROM Localizations",
                "SELECT LocId, Loc FROM Localizations",
            ]
        )

    rows = []
    for query in localization_queries:
        try:
            cursor.execute(query)
            rows = cursor.fetchall()
            if rows:
                break
        except sqlite3.Error:
            continue

    for lid, text in rows:
        if text:
            loc_map[lid] = text
    return loc_map


def _card_select_columns(card_columns: list) -> list:
    existing = set(card_columns)
    return [
        column if column in existing else "NULL"
        for column in _CARD_METADATA_COLUMNS
    ]


def _read_cards(cursor, tables: set) -> dict:
    if "Cards" not in tables or not (
        "Localizations" in tables or "Localizations_enUS" in tables
    ):
        return {}

    loc_map = _read_localizations(cursor, tables)
    card_columns = [row[1] for row in cursor.execute("PRAGMA table_info(Cards)")]
    select_columns = _card_select_columns(card_columns)
    query = (
        "SELECT GrpId, TitleId, "
        + ", ".join(select_columns)
        + " FROM Cards"
    )

    lookup: dict = {}
    cursor.execute(query)
    for grp_id, title_id, set_code, cn, rarity, colors in cursor.fetchall():
        if title_id not in loc_map:
            continue
        lookup[grp_id] = {
            "name": loc_map[title_id],
            "set": (set_code or "").upper(),
            "collector_number": str(cn) if cn else "",
            "rarity": _map_rarity(rarity),
            "colors": _parse_colors(colors),
            "type_line": "",
            "mana_cost": "",
            "cmc": None,
        }
    return lookup


def _load_mtga_file_lookup(path: Path) -> dict:
    try:
        with closing(sqlite3.connect(f"file:{path}?mode=ro", uri=True)) as conn:
            cursor = conn.cursor()
            tables = {
                row[0]
                for row in cursor.execute(
                    "SELECT name FROM sqlite_master WHERE type='table'"
                )
            }
            return _read_cards(cursor, tables)
    except sqlite3.Error:
        return {}


def load_local_mtga_database(
    raw_path: Optional[Path], progress: ProgressCallback = None
) -> dict:
    """Varre os arquivos SQLite locais do MTGA em busca das definições de carta."""
    if not raw_path or not raw_path.exists():
        return {}

    lookup: dict = {}
    try:
        all_files = sorted(
            raw_path.glob("*.mtga"), key=lambda f: f.stat().st_size, reverse=True
        )
    except Exception:
        return {}

    total = len(all_files)
    notify(progress, 0, total, "Lendo banco local...")
    for i, path in enumerate(all_files):
        notify(progress, i + 1, total, f"Verificando {path.name[:16]}...")
        if path.stat().st_size < 500 * 1024:
            continue
        lookup.update(_load_mtga_file_lookup(path))
        if len(lookup) > 1000:
            notify(progress, total, total, f"{len(lookup)} cartas locais.")
            return lookup

    notify(progress, total, total, f"{len(lookup)} cartas locais.")
    return lookup


def fetch_scryfall_database(progress: ProgressCallback = None) -> dict:
    """Baixa os dados de carta da API da Scryfall (metadados ricos)."""
    notify(progress, 0, 1, "Baixando dados da Scryfall...")
    try:
        meta = requests.get(
            "https://api.scryfall.com/bulk-data/default-cards", timeout=30
        ).json()
        cards = requests.get(meta["download_uri"], timeout=180).json()
        lookup: dict = {}
        total = len(cards)
        for i, c in enumerate(cards):
            if i % 2000 == 0:
                notify(progress, i, total, "Processando cartas...")
            aid = c.get("arena_id")
            if aid:
                lookup[aid] = {
                    "name": c.get("name", "Unknown"),
                    "set": c.get("set", "").upper(),
                    "collector_number": c.get("collector_number", ""),
                    "rarity": c.get("rarity", ""),
                    "colors": c.get("colors", c.get("color_identity", [])),
                    "type_line": c.get("type_line", ""),
                    "mana_cost": c.get("mana_cost", ""),
                    "cmc": c.get("cmc"),
                    "image": (c.get("image_uris") or {}).get("normal", ""),
                }
        notify(progress, total, total, f"{len(lookup)} cartas da Scryfall.")
        return lookup
    except Exception as e:
        notify(progress, 1, 1, f"Falha na Scryfall: {e}")
        return {}


def _read_lookup_cache(progress: ProgressCallback = None) -> dict | None:
    if not LOOKUP_FILE.exists():
        return None
    try:
        notify(progress, 0, 1, "Carregando cache...")
        data = json.loads(LOOKUP_FILE.read_text(encoding="utf-8"))
        return {int(k): v for k, v in data.items() if isinstance(v, dict)}
    except Exception:
        return None


def _write_lookup_cache(lookup: dict) -> None:
    if not lookup:
        return
    try:
        LOOKUP_FILE.write_text(
            json.dumps({str(k): v for k, v in lookup.items()}, ensure_ascii=False),
            encoding="utf-8",
        )
    except Exception:
        pass


def _load_lookup_from_source(config: Config, progress: ProgressCallback = None) -> dict:
    source = config.database_source
    lookup: dict = {}

    if source in ("auto", "local"):
        lookup = load_local_mtga_database(config.resolve_mtga_path(), progress)

    if not lookup and source in ("auto", "scryfall"):
        lookup = fetch_scryfall_database(progress)

    return lookup


def load_card_database(
    config: Config, progress: ProgressCallback = None, use_cache: bool = True
) -> dict:
    """Orquestra o carregamento do banco: Cache -> fonte configurada."""
    cache = _read_lookup_cache(progress) if use_cache else None
    if cache is not None:
        return cache

    lookup = _load_lookup_from_source(config, progress)
    _write_lookup_cache(lookup)
    return lookup


def name_to_id_map(db: dict) -> dict:
    """Mapa {nome_minusculo: grp_id} para busca de âncoras."""
    return {v["name"].lower(): k for k, v in db.items()}
