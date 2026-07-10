import json

from mtga import config
from mtga.config import Config, notify


def test_notify_ignores_callback_errors():
    def failing_callback(i, total, msg):
        raise RuntimeError("boom")

    notify(failing_callback, 1, 2, "testing")


def test_config_load_returns_default_for_invalid_file(tmp_path, monkeypatch):
    config_file = tmp_path / "config.json"
    config_file.write_text("{invalid", encoding="utf-8")
    monkeypatch.setattr(config, "CONFIG_FILE", config_file)

    loaded = Config.load()

    assert loaded == Config()


def test_config_save_and_load_round_trip(tmp_path, monkeypatch):
    config_file = tmp_path / "config.json"
    monkeypatch.setattr(config, "CONFIG_FILE", config_file)
    saved = Config(
        mtga_path="C:/MTGA/Raw",
        output_folder=str(tmp_path),
        database_source="scryfall",
        anchors=[[1, 4, "Opt"]],
    )

    saved.save()

    assert json.loads(config_file.read_text(encoding="utf-8")) == {
        "mtga_path": "C:/MTGA/Raw",
        "output_folder": str(tmp_path),
        "database_source": "scryfall",
        "language": "pt",
        "anchors": [[1, 4, "Opt"]],
    }
    assert Config.load() == saved


def test_config_load_normalizes_language(tmp_path, monkeypatch):
    config_file = tmp_path / "config.json"
    config_file.write_text(
        json.dumps({"language": "English"}),
        encoding="utf-8",
    )
    monkeypatch.setattr(config, "CONFIG_FILE", config_file)

    assert Config.load().language == "en"


def test_resolve_paths_use_existing_configured_folders(tmp_path):
    cfg = Config(mtga_path=str(tmp_path), output_folder=str(tmp_path / "out"))

    assert cfg.resolve_mtga_path() == tmp_path
    assert cfg.resolve_output_folder() == tmp_path / "out"
    assert (tmp_path / "out").exists()
