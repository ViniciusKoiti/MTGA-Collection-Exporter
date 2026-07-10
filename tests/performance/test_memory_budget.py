import gc
import tracemalloc

from mtga.collection import enrich_collection


def test_enrich_collection_memory_growth_stays_bounded():
    raw = {card_id: (card_id % 4) + 1 for card_id in range(1, 501)}
    db = {
        card_id: {
            "name": f"Card {card_id}",
            "set": "TST",
            "collector_number": str(card_id),
            "rarity": "common",
            "colors": [],
            "type_line": "Creature",
            "mana_cost": "",
            "cmc": 1,
        }
        for card_id in raw
    }

    tracemalloc.start()
    try:
        before = tracemalloc.take_snapshot()
        for _ in range(25):
            result = enrich_collection(raw, db)
            assert len(result) == len(raw)
        gc.collect()
        after = tracemalloc.take_snapshot()
    finally:
        tracemalloc.stop()

    growth = sum(stat.size_diff for stat in after.compare_to(before, "filename"))

    assert growth < 1_000_000
