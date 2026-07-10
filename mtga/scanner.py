"""
Leitura da coleção a partir da memória do processo MTGA.exe (via pymem).
Somente Windows, com o jogo aberto.
"""

from __future__ import annotations

import struct

from .config import ProgressCallback, notify


class MtgaNotRunning(Exception):
    """MTG Arena não está em execução."""


class ScanError(Exception):
    """Falha ao localizar/ler a coleção na memória."""


def connect_mtga():
    """Anexa ao processo MTGA.exe. Requer pymem (somente Windows)."""
    try:
        import pymem  # import tardio: só é preciso na hora do scan
    except ImportError as e:  # pragma: no cover
        raise ScanError("pymem não está instalado. Rode: pip install pymem") from e
    try:
        return pymem.Pymem("MTGA.exe")
    except Exception as e:
        raise MtgaNotRunning(
            "MTG Arena não está aberto. Abra o jogo e a aba 'Decks' primeiro."
        ) from e


def _find_blocks(pm, addr: int) -> list:
    """Lê a memória ao redor de um endereço e tenta achar o array de cartas."""
    try:
        data = pm.read_bytes(max(0, addr - 1024 * 1024), 4 * 1024 * 1024)
        ints = struct.unpack(f"<{len(data) // 4}I", data)
        blocks = []
        for off in (0, 1):
            curr: dict = {}
            misses = 0
            for i in range(off, len(ints) - 1, 2):
                k, v = ints[i], ints[i + 1]
                if 1000 <= k < 500000 and 1 <= v <= 400:
                    curr[k] = v
                    misses = 0
                else:
                    misses += 1
                if misses > 50:
                    if len(curr) > 50:
                        blocks.append(curr)
                    curr = {}
                    misses = 0
            if len(curr) > 50:
                blocks.append(curr)
        return blocks
    except Exception:
        return []


def scan_collection(pm, anchors: list, progress: ProgressCallback = None) -> dict:
    """
    Usa as âncoras (cartas conhecidas do usuário) para localizar o array da
    coleção na memória e retorna {grp_id: quantidade}.
    """
    if not anchors:
        raise ScanError("Nenhuma âncora de calibração definida.")

    matches = []
    total = len(anchors)
    for i, anchor in enumerate(anchors):
        aid, aqty, aname = anchor[0], anchor[1], anchor[2]
        display = (aname[:15] + "..") if len(aname) > 15 else aname
        notify(progress, i, total, f"Procurando {display}")
        res = pm.pattern_scan_all(
            struct.pack("<II", int(aid), int(aqty)), return_multiple=True
        )
        if res:
            matches.extend(res)
            notify(progress, i + 1, total, "Âncora encontrada!")
            if aqty > 1:
                break
        notify(progress, i + 1, total, "Concluído")

    if not matches:
        raise ScanError(
            "Não foi possível localizar a coleção a partir das âncoras. "
            "Confira as quantidades/cartas e tente cartas raras únicas."
        )

    candidates = []
    for m in matches:
        candidates.extend(_find_blocks(pm, m))
    if not candidates:
        raise ScanError("Nenhum bloco de dados válido encontrado na memória.")

    return max(candidates, key=len)
