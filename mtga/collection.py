"""
Transformação e persistência da coleção: enriquecer o resultado bruto do scan,
exportar (TXT/JSON/CSV) e reler a exportação salva.
"""

from __future__ import annotations

import csv
import json
from pathlib import Path
from typing import Optional

from .config import (
    OUTPUT_CSV_NAME,
    OUTPUT_JSON_NAME,
    OUTPUT_TXT_NAME,
    SCRIPT_DIR,
)


def enrich_collection(raw: dict, db: dict) -> list:
    """
    Converte {grp_id: qty} numa lista de dicts com metadados, agregando
    versões de mesma carta+set.
    """
    processed: dict = {}
    for cid, qty in raw.items():
        info = db.get(cid)
        if not info:
            continue
        key = (info["name"], info.get("set", ""))
        if key not in processed:
            processed[key] = {
                "count": 0,
                "name": info["name"],
                "set": info.get("set", ""),
                "collector_number": info.get("collector_number", ""),
                "rarity": info.get("rarity", ""),
                "colors": info.get("colors", []),
                "type_line": info.get("type_line", ""),
                "mana_cost": info.get("mana_cost", ""),
                "cmc": info.get("cmc"),
                "image": info.get("image", ""),
                "grp_id": cid,
            }
        processed[key]["count"] += qty
    return sorted(processed.values(), key=lambda x: (x["name"], x["set"]))


def export_all(final_list: list, output_folder: Path) -> dict:
    """Escreve TXT, JSON e CSV. Retorna os caminhos gerados."""
    output_folder = Path(output_folder)
    output_folder.mkdir(parents=True, exist_ok=True)
    txt = output_folder / OUTPUT_TXT_NAME
    js = output_folder / OUTPUT_JSON_NAME
    cv = output_folder / OUTPUT_CSV_NAME

    with txt.open("w", encoding="utf-8") as f:
        for i in final_list:
            set_str = f" ({i['set']})" if i["set"] else ""
            f.write(f"{i['count']} {i['name']}{set_str}\n")

    with js.open("w", encoding="utf-8") as f:
        json.dump(final_list, f, indent=2, ensure_ascii=False)

    with cv.open("w", encoding="utf-8", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(
            ["Count", "Name", "Edition", "Condition", "Language", "Foil", "Tag"]
        )
        for i in final_list:
            writer.writerow(
                [i["count"], i["name"], i["set"], "Near Mint", "English", "", ""]
            )

    return {"txt": txt, "json": js, "csv": cv}


def load_exported_collection(output_folder: Optional[Path] = None) -> list:
    """Lê o mtga_collection.json já exportado (usado pelo MCP sem scan ao vivo)."""
    folder = Path(output_folder) if output_folder else SCRIPT_DIR
    path = folder / OUTPUT_JSON_NAME
    if not path.exists():
        path = SCRIPT_DIR / OUTPUT_JSON_NAME
    if path.exists():
        try:
            return json.loads(path.read_text(encoding="utf-8"))
        except Exception:
            return []
    return []
