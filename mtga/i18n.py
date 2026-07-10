"""
Textos da interface desktop.

Esta camada e intencionalmente pequena: traduz os textos visiveis da GUI sem
alterar os contratos do core, MCP ou arquivos exportados.
"""

from __future__ import annotations

DEFAULT_LANGUAGE = "pt"
LANGUAGE_LABELS = {
    "pt": "Português",
    "en": "English",
}

_LABEL_TO_CODE = {label: code for code, label in LANGUAGE_LABELS.items()}

_TEXTS = {
    "pt": {
        "app_ready": "Pronto.",
        "tab_config": "⚙  Configuração",
        "tab_collection": "▦  Coleção",
        "paths_title": "Caminhos",
        "mtga_raw_folder": "Pasta 'Raw' do MTGA:",
        "auto_detect_placeholder": "vazio = detectar automaticamente",
        "browse": "📁 Procurar",
        "detect": "Detectar",
        "output_folder": "Pasta de saída:",
        "output_placeholder": "vazio = pasta do programa",
        "database_source": "Fonte do banco de cartas:",
        "language": "Idioma:",
        "language_restart_note": "A troca de idioma e aplicada imediatamente.",
        "anchors_title": "Âncoras de calibração",
        "anchors_help": (
            "Cartas raras/míticas que você possui, usadas para localizar a "
            "coleção na memória. Informe o nome e a quantidade exata."
        ),
        "anchor_name_placeholder": (
            "Nome da carta (ex.: Sheoldred, the Apocalypse)"
        ),
        "anchor_qty_placeholder": "Qtd",
        "add_anchor": "＋ Adicionar",
        "save_config": "💾 Salvar configuração",
        "no_anchors": (
            "Nenhuma âncora ainda. Adicione ao menos 1 (idealmente 3-5)."
        ),
        "scan_collection": "▶  Escanear Coleção",
        "scanning": "Escaneando...",
        "db_loaded": "Banco carregado: {count} cartas. Pronto para escanear.",
        "need_anchor": "Adicione ao menos 1 âncora antes de escanear.",
        "scan_done": "{status}: {unique} cartas únicas, {total} no total.",
        "unexpected_error": "Erro inesperado: {error}",
        "card_name_required": "Informe o nome da carta.",
        "invalid_quantity": "Quantidade inválida.",
        "db_loading": "Banco ainda carregando... tente em instantes.",
        "card_not_found": "Carta '{name}' não encontrada.",
        "anchor_added": "Âncora adicionada: {name} x{qty}",
        "choose_card": "Escolha a carta",
        "did_you_mean": "Você quis dizer?",
        "tutorial_title": "Passo a passo",
        "tutorial_heading": "Como escanear sua coleção",
        "close": "Fechar",
        "config_saved": "Configuração salva em config.json.",
        "select_mtga_raw": "Selecione a pasta 'Raw' do MTGA",
        "mtga_found": "Instalação encontrada: {path}",
        "mtga_not_found": "Instalação do MTGA não encontrada automaticamente.",
        "select_output": "Selecione a pasta de saída",
        "search_placeholder": "🔍  Buscar por nome, tipo ou set...",
        "all": "Todas",
        "open_folder": "📂 Abrir pasta",
        "count_header": "Qtd",
        "name_header": "Nome",
        "set_header": "Set",
        "collector_header": "Nº",
        "rarity_header": "Raridade",
        "colors_header": "Cores",
        "type_header": "Tipo",
        "no_collection": "Nenhuma coleção carregada ainda.",
        "showing_cards": "Exibindo {shown} cartas ({total} cópias)",
        "filter_suffix": " · filtro: {filter}",
        "folder_status": "Pasta: {folder}",
        "status_ok": "OK",
        "status_warning": "Atencao",
        "status_error": "Erro",
        "validation_summary": "{status}: {unique} cartas unicas, {total} copias",
    },
    "en": {
        "app_ready": "Ready.",
        "tab_config": "⚙  Settings",
        "tab_collection": "▦  Collection",
        "paths_title": "Paths",
        "mtga_raw_folder": "MTGA 'Raw' folder:",
        "auto_detect_placeholder": "empty = auto-detect",
        "browse": "📁 Browse",
        "detect": "Detect",
        "output_folder": "Output folder:",
        "output_placeholder": "empty = application folder",
        "database_source": "Card database source:",
        "language": "Language:",
        "language_restart_note": "Language changes are applied immediately.",
        "anchors_title": "Calibration anchors",
        "anchors_help": (
            "Rare/mythic cards you own, used to locate the collection in "
            "memory. Enter the card name and exact quantity."
        ),
        "anchor_name_placeholder": (
            "Card name (e.g. Sheoldred, the Apocalypse)"
        ),
        "anchor_qty_placeholder": "Qty",
        "add_anchor": "＋ Add",
        "save_config": "💾 Save settings",
        "no_anchors": "No anchors yet. Add at least 1 (ideally 3-5).",
        "scan_collection": "▶  Scan Collection",
        "scanning": "Scanning...",
        "db_loaded": "Database loaded: {count} cards. Ready to scan.",
        "need_anchor": "Add at least 1 anchor before scanning.",
        "scan_done": "{status}: {unique} unique cards, {total} total.",
        "unexpected_error": "Unexpected error: {error}",
        "card_name_required": "Enter the card name.",
        "invalid_quantity": "Invalid quantity.",
        "db_loading": "Database still loading... try again in a moment.",
        "card_not_found": "Card '{name}' not found.",
        "anchor_added": "Anchor added: {name} x{qty}",
        "choose_card": "Choose the card",
        "did_you_mean": "Did you mean?",
        "tutorial_title": "Step by step",
        "tutorial_heading": "How to scan your collection",
        "close": "Close",
        "config_saved": "Settings saved to config.json.",
        "select_mtga_raw": "Select the MTGA 'Raw' folder",
        "mtga_found": "Installation found: {path}",
        "mtga_not_found": "MTGA installation was not found automatically.",
        "select_output": "Select the output folder",
        "search_placeholder": "🔍  Search by name, type, or set...",
        "all": "All",
        "open_folder": "📂 Open folder",
        "count_header": "Qty",
        "name_header": "Name",
        "set_header": "Set",
        "collector_header": "No.",
        "rarity_header": "Rarity",
        "colors_header": "Colors",
        "type_header": "Type",
        "no_collection": "No collection loaded yet.",
        "showing_cards": "Showing {shown} cards ({total} copies)",
        "filter_suffix": " · filter: {filter}",
        "folder_status": "Folder: {folder}",
        "status_ok": "OK",
        "status_warning": "Warning",
        "status_error": "Error",
        "validation_summary": "{status}: {unique} unique cards, {total} copies",
    },
}


def normalize_language(value: str) -> str:
    """Converte codigo/rotulo em codigo suportado."""
    if value in LANGUAGE_LABELS:
        return value
    return _LABEL_TO_CODE.get(value, DEFAULT_LANGUAGE)


def language_label(language: str) -> str:
    return LANGUAGE_LABELS.get(normalize_language(language), LANGUAGE_LABELS["pt"])


def translate(language: str, key: str, **kwargs) -> str:
    texts = _TEXTS.get(normalize_language(language), _TEXTS[DEFAULT_LANGUAGE])
    template = texts.get(key, _TEXTS[DEFAULT_LANGUAGE].get(key, key))
    return template.format(**kwargs)
