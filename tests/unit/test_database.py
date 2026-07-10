import json
import sqlite3

from mtga import database
from mtga.config import Config
from mtga.database import (
    _parse_colors,
    _read_localizations,
    fetch_scryfall_database,
    load_card_database,
    load_local_mtga_database,
    name_to_id_map,
)


def test_load_card_database_reads_cache_with_integer_keys(tmp_path, monkeypatch):
    cache = tmp_path / "arena_id_lookup.json"
    cache.write_text(
        json.dumps({"123": {"name": "Opt"}, "bad": ["not", "a", "dict"]}),
        encoding="utf-8",
    )
    monkeypatch.setattr(database, "LOOKUP_FILE", cache)

    result = load_card_database(Config(), use_cache=True)

    assert result == {123: {"name": "Opt"}}


def test_load_card_database_prefers_local_before_scryfall(monkeypatch, tmp_path):
    cache = tmp_path / "missing_cache.json"
    monkeypatch.setattr(database, "LOOKUP_FILE", cache)
    monkeypatch.setattr(
        database,
        "load_local_mtga_database",
        lambda raw_path, progress=None: {1: {"name": "Local Card"}},
    )
    monkeypatch.setattr(
        database,
        "fetch_scryfall_database",
        lambda progress=None: {2: {"name": "Remote Card"}},
    )

    result = load_card_database(Config(database_source="auto"), use_cache=True)

    assert result == {1: {"name": "Local Card"}}
    assert json.loads(cache.read_text(encoding="utf-8")) == {
        "1": {"name": "Local Card"}
    }


def test_name_to_id_map_uses_lowercase_names():
    db = {10: {"name": "Goblin Guide"}, 20: {"name": "Opt"}}

    assert name_to_id_map(db) == {"goblin guide": 10, "opt": 20}


def test_parse_colors_maps_arena_color_ids():
    assert _parse_colors("1, 3, 5, 9") == ["W", "B", "G"]
    assert _parse_colors(None) == []


def test_read_localizations_accepts_en_us_table():
    conn = sqlite3.connect(":memory:")
    try:
        cursor = conn.cursor()
        cursor.execute("CREATE TABLE Localizations_enUS (LocId INTEGER, Loc TEXT)")
        cursor.execute("INSERT INTO Localizations_enUS VALUES (10, 'Opt')")

        result = _read_localizations(cursor, {"Localizations_enUS"})
    finally:
        conn.close()

    assert result == {10: "Opt"}


def test_load_local_mtga_database_returns_empty_for_missing_path(tmp_path):
    assert load_local_mtga_database(tmp_path / "missing") == {}


def test_fetch_scryfall_database_builds_arena_lookup(monkeypatch):
    class FakeResponse:
        def __init__(self, payload):
            self.payload = payload

        def json(self):
            return self.payload

    responses = [
        FakeResponse({"download_uri": "https://example.test/cards.json"}),
        FakeResponse(
            [
                {
                    "arena_id": 77,
                    "name": "Opt",
                    "set": "m21",
                    "collector_number": "59",
                    "rarity": "common",
                    "colors": ["U"],
                    "type_line": "Instant",
                    "mana_cost": "{U}",
                    "cmc": 1,
                    "image_uris": {"normal": "https://example.test/opt.jpg"},
                },
                {"name": "No Arena Id"},
            ]
        ),
    ]
    calls = []

    def fake_get(url, timeout):
        calls.append((url, timeout))
        return responses.pop(0)

    monkeypatch.setattr(database.requests, "get", fake_get)

    assert fetch_scryfall_database() == {
        77: {
            "name": "Opt",
            "set": "M21",
            "collector_number": "59",
            "rarity": "common",
            "colors": ["U"],
            "type_line": "Instant",
            "mana_cost": "{U}",
            "cmc": 1,
            "image": "https://example.test/opt.jpg",
        }
    }
    assert calls == [
        ("https://api.scryfall.com/bulk-data/default-cards", 30),
        ("https://example.test/cards.json", 180),
    ]
