import csv
import json

from mtga.collection import enrich_collection, export_all, load_exported_collection


def test_enrich_collection_aggregates_same_name_and_set():
    raw = {10: 2, 11: 1, 99: 4}
    db = {
        10: {
            "name": "Lightning Bolt",
            "set": "STA",
            "collector_number": "42",
            "rarity": "rare",
            "colors": ["R"],
            "type_line": "Instant",
            "mana_cost": "{R}",
            "cmc": 1,
        },
        11: {
            "name": "Lightning Bolt",
            "set": "STA",
            "collector_number": "43",
            "rarity": "rare",
            "colors": ["R"],
            "type_line": "Instant",
            "mana_cost": "{R}",
            "cmc": 1,
        },
    }

    enriched = enrich_collection(raw, db)

    assert enriched == [
        {
            "count": 3,
            "name": "Lightning Bolt",
            "set": "STA",
            "collector_number": "42",
            "rarity": "rare",
            "colors": ["R"],
            "type_line": "Instant",
            "mana_cost": "{R}",
            "cmc": 1,
            "image": "",
            "grp_id": 10,
        }
    ]


def test_export_all_writes_consistent_formats(tmp_path):
    cards = [
        {
            "count": 4,
            "name": "Opt",
            "set": "M21",
            "collector_number": "59",
            "rarity": "common",
            "colors": ["U"],
            "type_line": "Instant",
            "mana_cost": "{U}",
            "cmc": 1,
            "image": "",
            "grp_id": 123,
        }
    ]

    paths = export_all(cards, tmp_path)

    assert paths["txt"].read_text(encoding="utf-8") == "4 Opt (M21)\n"
    assert json.loads(paths["json"].read_text(encoding="utf-8")) == cards

    with paths["csv"].open(encoding="utf-8", newline="") as csv_file:
        rows = list(csv.reader(csv_file))
    assert rows == [
        ["Count", "Name", "Edition", "Condition", "Language", "Foil", "Tag"],
        ["4", "Opt", "M21", "Near Mint", "English", "", ""],
    ]


def test_load_exported_collection_returns_empty_list_for_invalid_json(tmp_path):
    (tmp_path / "mtga_collection.json").write_text("{invalid", encoding="utf-8")

    assert load_exported_collection(tmp_path) == []
