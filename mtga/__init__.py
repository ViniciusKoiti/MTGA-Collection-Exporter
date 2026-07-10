"""
MTGA Collection Exporter — pacote principal.

Reexporta a API pública para que consumidores possam fazer, por exemplo:
    from mtga import Config, run_full_scan
"""

from .collection import enrich_collection, export_all, load_exported_collection
from .config import Config
from .database import load_card_database, name_to_id_map
from .pipeline import run_full_scan
from .scanner import MtgaNotRunning, ScanError, connect_mtga, scan_collection
from .validation import export_validation_report, validate_collection

__all__ = [
    "Config",
    "load_card_database",
    "name_to_id_map",
    "connect_mtga",
    "scan_collection",
    "enrich_collection",
    "export_all",
    "load_exported_collection",
    "run_full_scan",
    "MtgaNotRunning",
    "ScanError",
    "validate_collection",
    "export_validation_report",
]
