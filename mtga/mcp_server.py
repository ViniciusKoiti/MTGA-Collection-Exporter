"""
Servidor MCP (Model Context Protocol): expõe a coleção para uma IA ajudar a
montar decks. Por padrão lê a exportação salva; `rescan` refaz a leitura ao vivo.

Requisitos: pip install "mcp[cli]"
"""

from __future__ import annotations

from collections import Counter

from mcp.server.fastmcp import FastMCP

from .collection import load_exported_collection
from .config import Config
from .pipeline import run_full_scan
from .scanner import MtgaNotRunning, ScanError

mcp = FastMCP("mtga-collection")


def _load() -> list:
    """Carrega a coleção exportada mais recente."""
    cfg = Config.load()
    return load_exported_collection(cfg.resolve_output_folder())


@mcp.tool()
def get_collection() -> list:
    """
    Retorna a coleção completa do usuário como uma lista de cartas.

    Cada carta tem: count (quantas cópias), name, set, collector_number,
    rarity, colors (lista tipo ["R","G"]), type_line, mana_cost, cmc.
    Use isto para ter uma visão geral do que o usuário possui.
    """
    return _load()


@mcp.tool()
def collection_stats() -> dict:
    """
    Resumo estatístico da coleção: total de cartas únicas e cópias,
    e a contagem por raridade, por cor e por set.
    """
    coll = _load()
    by_rarity = Counter()
    by_color = Counter()
    by_set = Counter()
    total_copies = 0
    for c in coll:
        cnt = c.get("count", 0)
        total_copies += cnt
        by_rarity[c.get("rarity", "") or "unknown"] += 1
        by_set[c.get("set", "") or "?"] += 1
        colors = c.get("colors") or []
        if not colors:
            by_color["Colorless"] += 1
        for col in colors:
            by_color[col] += 1
    return {
        "unique_cards": len(coll),
        "total_copies": total_copies,
        "by_rarity": dict(by_rarity),
        "by_color": dict(by_color),
        "top_sets": dict(by_set.most_common(15)),
    }


@mcp.tool()
def search_cards(
    query: str = "",
    colors: str = "",
    rarity: str = "",
    type_contains: str = "",
    min_count: int = 1,
    limit: int = 100,
) -> list:
    """
    Filtra a coleção do usuário.

    Args:
        query: texto buscado no nome/tipo/set (case-insensitive).
        colors: filtro de cor. Ex.: "R" (contém vermelho) ou "RG"
                (contém vermelho E verde). Vazio = qualquer cor.
        rarity: "mythic" | "rare" | "uncommon" | "common" | "basic".
        type_contains: substring do type_line (ex.: "Creature", "Instant").
        min_count: só cartas com pelo menos N cópias (útil p/ playsets: 4).
        limit: máximo de resultados.
    """
    coll = _load()
    q = query.lower().strip()
    want_colors = [c.strip().upper() for c in colors if c.strip()]
    rarity = rarity.lower().strip()
    tc = type_contains.lower().strip()

    out = []
    for c in coll:
        if not _matches_search(c, q, want_colors, rarity, tc, min_count):
            continue
        out.append(c)
        if len(out) >= limit:
            break
    return out


def _matches_search(
    card: dict,
    query: str,
    want_colors: list,
    rarity: str,
    type_contains: str,
    min_count: int,
) -> bool:
    return all(
        (
            card.get("count", 0) >= min_count,
            _matches_rarity(card, rarity),
            _matches_colors(card, want_colors),
            _matches_type(card, type_contains),
            _matches_query(card, query),
        )
    )


def _matches_rarity(card: dict, rarity: str) -> bool:
    return not rarity or card.get("rarity", "") == rarity


def _matches_colors(card: dict, want_colors: list) -> bool:
    card_colors = [color.upper() for color in (card.get("colors") or [])]
    return not want_colors or all(color in card_colors for color in want_colors)


def _matches_type(card: dict, type_contains: str) -> bool:
    return not type_contains or type_contains in card.get("type_line", "").lower()


def _matches_query(card: dict, query: str) -> bool:
    if not query:
        return True
    haystack = (
        f"{card['name']} {card.get('set', '')} {card.get('type_line', '')}".lower()
    )
    return query in haystack


@mcp.tool()
def check_deck(cards: list) -> dict:
    """
    Verifica se o usuário possui as cartas de uma decklist e o que falta.

    Args:
        cards: lista de itens {"name": <nome>, "count": <qtd desejada>}.
               `count` é opcional (padrão 1).
    """
    coll = _load()
    owned_map: dict = {}
    for c in coll:
        key = c["name"].lower()
        owned_map[key] = owned_map.get(key, 0) + c.get("count", 0)

    owned, missing = [], []
    for item in cards:
        name = str(item.get("name", "")).strip()
        want = int(item.get("count", 1) or 1)
        have = owned_map.get(name.lower(), 0)
        if have >= want:
            owned.append({"name": name, "want": want, "have": have})
        else:
            missing.append(
                {"name": name, "want": want, "have": have, "need": want - have}
            )
    return {
        "buildable": len(missing) == 0,
        "owned": owned,
        "missing": missing,
        "missing_count": sum(m["need"] for m in missing),
    }


@mcp.tool()
def rescan() -> dict:
    """
    Refaz a leitura AO VIVO da coleção a partir da memória do MTGA e reexporta.
    Requer: Windows, MTG Arena aberto na aba 'Decks', pymem e âncoras configuradas.
    """
    cfg = Config.load()
    try:
        final_list = run_full_scan(cfg, progress=None)
    except (MtgaNotRunning, ScanError) as e:
        return {"ok": False, "error": str(e)}
    except Exception as e:  # pragma: no cover
        return {"ok": False, "error": f"Erro inesperado: {e}"}
    return {
        "ok": True,
        "unique_cards": len(final_list),
        "total_copies": sum(c["count"] for c in final_list),
    }
