import json

from mtga.validation import (
    STATUS_ERROR,
    STATUS_OK,
    STATUS_WARNING,
    export_validation_report,
    validate_collection,
)


def _card(name, count=1):
    return {"name": name, "count": count}


def test_validate_collection_reports_ok_for_stable_collection():
    collection = [_card("Opt", 4), _card("Lightning Bolt", 2)]
    previous = [_card("Opt", 4), _card("Lightning Bolt", 2)]

    report = validate_collection(
        collection,
        previous_collection=previous,
        anchors=[[1, 4, "Opt"], [2, 2, "Bolt"], [3, 1, "Ugin"]],
        database_size=5000,
    )

    assert report.status == STATUS_OK
    assert report.unique_cards == 2
    assert report.total_copies == 6
    assert report.unique_delta == 0
    assert report.total_delta == 0
    assert report.issues == []


def test_validate_collection_flags_empty_collection_as_error():
    report = validate_collection([])

    assert report.status == STATUS_ERROR
    assert [issue.code for issue in report.issues] == ["empty_collection"]


def test_validate_collection_warns_about_weak_inputs():
    report = validate_collection([_card("Opt", 1)], anchors=[[1, 1, "Opt"]])

    assert report.status == STATUS_WARNING
    assert [issue.code for issue in report.issues] == ["few_anchors"]


def test_validate_collection_warns_about_large_drop_from_previous_scan():
    previous = [_card(f"Card {index}", 4) for index in range(8)]
    current = [_card("Card 1", 4), _card("Card 2", 4)]

    report = validate_collection(
        current,
        previous_collection=previous,
        anchors=[[1, 4, "A"], [2, 4, "B"], [3, 4, "C"]],
    )

    assert report.status == STATUS_WARNING
    assert {issue.code for issue in report.issues} == {
        "large_unique_drop",
        "large_total_drop",
    }
    assert report.unique_delta == -6
    assert report.total_delta == -24


def test_export_validation_report_writes_json(tmp_path):
    report = validate_collection(
        [_card("Opt", 4)],
        anchors=[[1, 4, "A"], [2, 1, "B"], [3, 1, "C"]],
        database_size=2000,
    )

    path = export_validation_report(report, tmp_path)

    data = json.loads(path.read_text(encoding="utf-8"))
    assert path.name == "mtga_scan_validation.json"
    assert data["status"] == STATUS_OK
    assert data["summary"] == "OK: 1 cartas unicas, 4 copias"
