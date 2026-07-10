"""
Validacao de qualidade do scan da colecao.

As regras aqui sao deterministicas e offline: comparam a colecao final com
o snapshot anterior e sinalizam casos que merecem revisao pelo usuario.
"""

from __future__ import annotations

import json
from dataclasses import asdict, dataclass, field
from datetime import datetime, timezone
from pathlib import Path

from .config import VALIDATION_REPORT_NAME

STATUS_OK = "ok"
STATUS_WARNING = "warning"
STATUS_ERROR = "error"


@dataclass
class ValidationIssue:
    severity: str
    code: str
    message: str


@dataclass
class CollectionValidationReport:
    status: str
    generated_at: str
    unique_cards: int
    total_copies: int
    previous_unique_cards: int | None = None
    previous_total_copies: int | None = None
    unique_delta: int | None = None
    total_delta: int | None = None
    anchor_count: int = 0
    database_size: int | None = None
    issues: list[ValidationIssue] = field(default_factory=list)

    @property
    def status_label(self) -> str:
        labels = {
            STATUS_OK: "OK",
            STATUS_WARNING: "Atencao",
            STATUS_ERROR: "Erro",
        }
        return labels.get(self.status, self.status)

    @property
    def summary(self) -> str:
        return (
            f"{self.status_label}: {self.unique_cards} cartas unicas, "
            f"{self.total_copies} copias"
        )

    def to_dict(self) -> dict:
        data = asdict(self)
        data["status_label"] = self.status_label
        data["summary"] = self.summary
        return data


def validate_collection(
    collection: list,
    previous_collection: list | None = None,
    anchors: list | None = None,
    database_size: int | None = None,
) -> CollectionValidationReport:
    """Gera um relatorio simples de confiabilidade para o scan final."""
    unique_cards = len(collection)
    total_copies = _total_copies(collection)
    previous_unique = None
    previous_total = None
    unique_delta = None
    total_delta = None

    if previous_collection is not None:
        previous_unique = len(previous_collection)
        previous_total = _total_copies(previous_collection)
        unique_delta = unique_cards - previous_unique
        total_delta = total_copies - previous_total

    issues = _build_issues(
        unique_cards=unique_cards,
        total_copies=total_copies,
        previous_unique=previous_unique,
        previous_total=previous_total,
        anchor_count=len(anchors or []),
        database_size=database_size,
    )

    return CollectionValidationReport(
        status=_status_from_issues(issues),
        generated_at=datetime.now(timezone.utc).isoformat(),
        unique_cards=unique_cards,
        total_copies=total_copies,
        previous_unique_cards=previous_unique,
        previous_total_copies=previous_total,
        unique_delta=unique_delta,
        total_delta=total_delta,
        anchor_count=len(anchors or []),
        database_size=database_size,
        issues=issues,
    )


def export_validation_report(
    report: CollectionValidationReport, output_folder: Path
) -> Path:
    """Salva o relatorio de validacao no mesmo destino dos exports."""
    output_folder = Path(output_folder)
    output_folder.mkdir(parents=True, exist_ok=True)
    path = output_folder / VALIDATION_REPORT_NAME
    path.write_text(
        json.dumps(report.to_dict(), indent=2, ensure_ascii=False),
        encoding="utf-8",
    )
    return path


def _build_issues(
    unique_cards: int,
    total_copies: int,
    previous_unique: int | None,
    previous_total: int | None,
    anchor_count: int,
    database_size: int | None,
) -> list[ValidationIssue]:
    issues: list[ValidationIssue] = []
    if unique_cards == 0:
        issues.append(
            ValidationIssue(
                STATUS_ERROR,
                "empty_collection",
                "A colecao exportada ficou vazia.",
            )
        )
        return issues

    if anchor_count < 3:
        issues.append(
            ValidationIssue(
                STATUS_WARNING,
                "few_anchors",
                "Use 3 ou mais ancoras para aumentar a confiabilidade do scan.",
            )
        )
    if database_size is not None and database_size < 1000:
        issues.append(
            ValidationIssue(
                STATUS_WARNING,
                "small_database",
                "O banco de cartas carregado parece pequeno.",
            )
        )
    if total_copies < unique_cards:
        issues.append(
            ValidationIssue(
                STATUS_WARNING,
                "inconsistent_totals",
                "O total de copias ficou menor que a quantidade de cartas unicas.",
            )
        )

    issues.extend(
        _compare_with_previous(
            unique_cards=unique_cards,
            total_copies=total_copies,
            previous_unique=previous_unique,
            previous_total=previous_total,
        )
    )
    return issues


def _compare_with_previous(
    unique_cards: int,
    total_copies: int,
    previous_unique: int | None,
    previous_total: int | None,
) -> list[ValidationIssue]:
    if not previous_unique or not previous_total:
        return []

    issues: list[ValidationIssue] = []
    unique_drop = _drop_ratio(unique_cards, previous_unique)
    total_drop = _drop_ratio(total_copies, previous_total)

    if unique_drop >= 0.25:
        issues.append(
            ValidationIssue(
                STATUS_WARNING,
                "large_unique_drop",
                "A quantidade de cartas unicas caiu mais de 25% desde o ultimo scan.",
            )
        )
    if total_drop >= 0.25:
        issues.append(
            ValidationIssue(
                STATUS_WARNING,
                "large_total_drop",
                "O total de copias caiu mais de 25% desde o ultimo scan.",
            )
        )
    return issues


def _total_copies(collection: list) -> int:
    return sum(int(card.get("count", 0) or 0) for card in collection)


def _drop_ratio(current: int, previous: int) -> float:
    if previous <= 0 or current >= previous:
        return 0
    return (previous - current) / previous


def _status_from_issues(issues: list[ValidationIssue]) -> str:
    if any(issue.severity == STATUS_ERROR for issue in issues):
        return STATUS_ERROR
    if issues:
        return STATUS_WARNING
    return STATUS_OK
