import struct

import pytest

from mtga import scanner


class FakePymem:
    def __init__(self, data=b"", matches=None):
        self.data = data
        self.matches = matches or []
        self.patterns = []

    def read_bytes(self, address, size):
        return self.data

    def pattern_scan_all(self, pattern, return_multiple=True):
        self.patterns.append((pattern, return_multiple))
        return self.matches


def test_find_blocks_extracts_valid_card_quantity_pairs():
    pairs = []
    for index in range(60):
        pairs.extend([1000 + index, 1])
    data = struct.pack(f"<{len(pairs)}I", *pairs)

    blocks = scanner._find_blocks(FakePymem(data), addr=1024)

    assert len(blocks) == 1
    assert blocks[0][1000] == 1
    assert blocks[0][1059] == 1


def test_scan_collection_requires_anchors():
    with pytest.raises(scanner.ScanError, match="Nenhuma"):
        scanner.scan_collection(FakePymem(), [])


def test_scan_collection_uses_largest_candidate(monkeypatch):
    fake_pm = FakePymem(matches=[2048])
    candidates = [{1000: 1}, {1000: 1, 1001: 2, 1002: 3}]
    monkeypatch.setattr(scanner, "_find_blocks", lambda pm, addr: candidates)

    result = scanner.scan_collection(fake_pm, [[1000, 1, "Opt"]])

    assert result == candidates[1]
    assert fake_pm.patterns == [(struct.pack("<II", 1000, 1), True)]
