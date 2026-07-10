from mtga import mcp_server
from mtga.config import Config

COLLECTION = [
    {
        "count": 4,
        "name": "Opt",
        "set": "M21",
        "rarity": "common",
        "colors": ["U"],
        "type_line": "Instant",
    },
    {
        "count": 2,
        "name": "Lightning Bolt",
        "set": "STA",
        "rarity": "rare",
        "colors": ["R"],
        "type_line": "Instant",
    },
    {
        "count": 1,
        "name": "Ugin",
        "set": "M21",
        "rarity": "mythic",
        "colors": [],
        "type_line": "Planeswalker",
    },
]


def test_collection_stats_counts_cards(monkeypatch):
    monkeypatch.setattr(mcp_server, "_load", lambda: COLLECTION)

    stats = mcp_server.collection_stats()

    assert stats["unique_cards"] == 3
    assert stats["total_copies"] == 7
    assert stats["by_rarity"] == {"common": 1, "rare": 1, "mythic": 1}
    assert stats["by_color"] == {"U": 1, "R": 1, "Colorless": 1}
    assert stats["top_sets"] == {"M21": 2, "STA": 1}


def test_search_cards_filters_by_color_rarity_type_and_count(monkeypatch):
    monkeypatch.setattr(mcp_server, "_load", lambda: COLLECTION)

    result = mcp_server.search_cards(
        query="bolt",
        colors="R",
        rarity="rare",
        type_contains="instant",
        min_count=2,
    )

    assert result == [COLLECTION[1]]


def test_check_deck_reports_missing_cards(monkeypatch):
    monkeypatch.setattr(mcp_server, "_load", lambda: COLLECTION)

    result = mcp_server.check_deck(
        [{"name": "Opt", "count": 4}, {"name": "Lightning Bolt", "count": 4}]
    )

    assert result == {
        "buildable": False,
        "owned": [{"name": "Opt", "want": 4, "have": 4}],
        "missing": [
            {"name": "Lightning Bolt", "want": 4, "have": 2, "need": 2}
        ],
        "missing_count": 2,
    }


def test_rescan_returns_scan_summary(monkeypatch):
    monkeypatch.setattr(mcp_server.Config, "load", lambda: Config())
    monkeypatch.setattr(
        mcp_server,
        "run_full_scan",
        lambda cfg, progress=None: [{"count": 4}, {"count": 1}],
    )

    assert mcp_server.rescan() == {
        "ok": True,
        "unique_cards": 2,
        "total_copies": 5,
    }


def test_rescan_returns_expected_scan_errors(monkeypatch):
    monkeypatch.setattr(mcp_server.Config, "load", lambda: Config())

    def failing_scan(cfg, progress=None):
        raise mcp_server.ScanError("failed")

    monkeypatch.setattr(mcp_server, "run_full_scan", failing_scan)

    assert mcp_server.rescan() == {"ok": False, "error": "failed"}
