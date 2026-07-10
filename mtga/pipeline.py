"""
Orquestração do fluxo completo, reutilizado pela GUI e pelo MCP.
"""

from __future__ import annotations

from .collection import enrich_collection, export_all
from .config import Config, ProgressCallback
from .database import load_card_database
from .scanner import ScanError, connect_mtga, scan_collection


def run_full_scan(config: Config, progress: ProgressCallback = None) -> list:
    """
    Executa o pipeline inteiro: banco -> conectar -> scan -> enriquecer -> exportar.
    Retorna a lista final enriquecida. Lança MtgaNotRunning/ScanError em falhas.
    """
    db = load_card_database(config, progress)
    if not db:
        raise ScanError("Falha ao inicializar o banco de cartas.")
    pm = connect_mtga()
    raw = scan_collection(pm, config.anchors, progress)
    final_list = enrich_collection(raw, db)
    export_all(final_list, config.resolve_output_folder())
    return final_list
