from mtga.i18n import language_label, normalize_language, translate


def test_normalize_language_accepts_code_and_label():
    assert normalize_language("en") == "en"
    assert normalize_language("English") == "en"
    assert normalize_language("unknown") == "pt"


def test_language_label_returns_display_label():
    assert language_label("en") == "English"
    assert language_label("pt") == "Português"


def test_translate_formats_language_specific_text():
    assert translate("en", "scan_done", status="OK", unique=2, total=6) == (
        "OK: 2 unique cards, 6 total."
    )
    assert translate("pt", "scan_done", status="OK", unique=2, total=6) == (
        "OK: 2 cartas únicas, 6 no total."
    )
