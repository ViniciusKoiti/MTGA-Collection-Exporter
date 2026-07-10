from mtga import pipeline
from mtga.config import Config


def test_run_full_scan_orchestrates_core_steps(monkeypatch, tmp_path):
    calls = []
    config = Config(output_folder=str(tmp_path), anchors=[[1, 2, "Anchor"]])

    monkeypatch.setattr(
        pipeline,
        "load_card_database",
        lambda cfg, progress=None: calls.append("db") or {1: {"name": "Opt"}},
    )
    monkeypatch.setattr(
        pipeline,
        "connect_mtga",
        lambda: calls.append("connect") or object(),
    )
    monkeypatch.setattr(
        pipeline,
        "scan_collection",
        lambda pm, anchors, progress=None: calls.append(("scan", anchors)) or {1: 4},
    )
    monkeypatch.setattr(
        pipeline,
        "enrich_collection",
        lambda raw, db: calls.append(("enrich", raw, db)) or [{"name": "Opt"}],
    )
    monkeypatch.setattr(
        pipeline,
        "export_all",
        lambda final_list, output_folder: calls.append(("export", output_folder)),
    )

    result = pipeline.run_full_scan(config)

    assert result == [{"name": "Opt"}]
    assert calls == [
        "db",
        "connect",
        ("scan", [[1, 2, "Anchor"]]),
        ("enrich", {1: 4}, {1: {"name": "Opt"}}),
        ("export", tmp_path),
    ]


def test_run_full_scan_fails_when_database_is_empty(monkeypatch):
    monkeypatch.setattr(pipeline, "load_card_database", lambda cfg, progress=None: {})

    try:
        pipeline.run_full_scan(Config())
    except pipeline.ScanError as exc:
        assert "banco de cartas" in str(exc)
    else:
        raise AssertionError("ScanError was not raised")
