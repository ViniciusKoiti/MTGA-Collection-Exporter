"""
Constantes, caminhos base e a configuração persistida (config.json).
"""

from __future__ import annotations

import json
import sys
from dataclasses import asdict, dataclass, field
from pathlib import Path
from typing import Callable, Optional

# ---------------------------------------------------------------------------
# Caminhos base
# ---------------------------------------------------------------------------

if getattr(sys, "frozen", False):
    SCRIPT_DIR = Path(sys.executable).parent
else:
    # raiz do projeto (pasta que contém o pacote "mtga")
    SCRIPT_DIR = Path(__file__).resolve().parent.parent

CONFIG_FILE = SCRIPT_DIR / "config.json"
LOOKUP_FILE = SCRIPT_DIR / "arena_id_lookup.json"

OUTPUT_JSON_NAME = "mtga_collection.json"
OUTPUT_TXT_NAME = "mtga_collection.txt"
OUTPUT_CSV_NAME = "mtga_collection.csv"
VALIDATION_REPORT_NAME = "mtga_scan_validation.json"

# Locais padrão da instalação do MTGA (pasta "Raw" com os arquivos .mtga)
DEFAULT_MTGA_PATHS = [
    r"C:\Program Files (x86)\Steam\steamapps\common\MTGA\MTGA_Data\Downloads\Raw",
    r"C:\Program Files\Wizards of the Coast\MTGA\MTGA_Data\Downloads\Raw",
    r"C:\Program Files (x86)\Wizards of the Coast\MTGA\MTGA_Data\Downloads\Raw",
    r"D:\MTGA\MTGA_Data\Downloads"
]

# Um callback de progresso recebe (iteração_atual, total, mensagem_texto).
ProgressCallback = Optional[Callable[[int, int, str], None]]


def notify(progress: ProgressCallback, i: int, total: int, msg: str) -> None:
    """Chama o callback de progresso com segurança (ignora exceções)."""
    if progress:
        try:
            progress(i, total, msg)
        except Exception:
            pass


# ---------------------------------------------------------------------------
# Configuração
# ---------------------------------------------------------------------------

@dataclass
class Config:
    """Parâmetros configuráveis pelo usuário, persistidos em config.json."""

    mtga_path: str = ""            # pasta "Raw"; vazio = detectar automaticamente
    output_folder: str = ""        # vazio = usar SCRIPT_DIR
    database_source: str = "auto"  # "auto" | "local" | "scryfall"
    # Cada âncora é [grp_id, quantidade, nome_exibicao]
    anchors: list = field(default_factory=list)

    @classmethod
    def load(cls) -> "Config":
        if CONFIG_FILE.exists():
            try:
                data = json.loads(CONFIG_FILE.read_text(encoding="utf-8"))
                return cls(
                    mtga_path=data.get("mtga_path", ""),
                    output_folder=data.get("output_folder", ""),
                    database_source=data.get("database_source", "auto"),
                    anchors=data.get("anchors", []),
                )
            except Exception:
                pass
        return cls()

    def save(self) -> None:
        try:
            CONFIG_FILE.write_text(
                json.dumps(asdict(self), indent=2, ensure_ascii=False),
                encoding="utf-8",
            )
        except Exception:
            pass

    def resolve_mtga_path(self) -> Optional[Path]:
        """Retorna a pasta Raw configurada, ou tenta detectar automaticamente."""
        if self.mtga_path:
            p = Path(self.mtga_path)
            if p.exists():
                return p
        for candidate in DEFAULT_MTGA_PATHS:
            p = Path(candidate)
            if p.exists():
                return p
        return None

    def resolve_output_folder(self) -> Path:
        if self.output_folder:
            p = Path(self.output_folder)
            p.mkdir(parents=True, exist_ok=True)
            return p
        return SCRIPT_DIR
