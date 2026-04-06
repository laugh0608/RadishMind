from __future__ import annotations

import subprocess
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import (  # noqa: E402
    ensure_schema,
    load_json_document,
    make_repo_relative,
)

SUMMARY_SCHEMA_PATH = REPO_ROOT / "datasets/eval/batch-orchestration-summary.schema.json"
RECOMMENDED_SUMMARY_SCHEMA_PATH = REPO_ROOT / "datasets/eval/recommended-negative-replay-summary.schema.json"
REAL_BATCH_OUTPUT_ROOT = REPO_ROOT / "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1"
REAL_BATCH_COLLECTION = "2026-04-05-radish-docs-qa-real-batch-v1"
REAL_BATCH_CROSS_SAMPLE_NEGATIVE_DIR = REPO_ROOT / "datasets/eval/radish-negative"


def expect_object(document: Any, label: str) -> dict[str, Any]:
    if not isinstance(document, dict):
        raise SystemExit(f"{label} must be a json object")
    return document


def run_command(command: list[str]) -> None:
    result = subprocess.run(
        command,
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
    )
    if result.stdout:
        print(result.stdout, end="")
    if result.stderr:
        print(result.stderr, end="", file=sys.stderr)
    if result.returncode != 0:
        raise SystemExit(result.returncode)


def require_equal(actual: Any, expected: Any, label: str) -> None:
    if actual != expected:
        raise SystemExit(f"{label}: expected {expected!r}, got {actual!r}")


def require_true(value: Any, label: str) -> None:
    if value is not True:
        raise SystemExit(f"{label} must be true")


def require_false(value: Any, label: str) -> None:
    if value is not False:
        raise SystemExit(f"{label} must be false")


def require_non_empty_list(value: Any, label: str) -> list[Any]:
    if not isinstance(value, list) or not value:
        raise SystemExit(f"{label} must be a non-empty json array")
    return value


def require_missing_path(path: Path, label: str) -> None:
    if path.exists():
        raise SystemExit(f"{label} should not exist: {path}")


def load_summary(path: Path, schema_path: Path) -> dict[str, Any]:
    document = expect_object(load_json_document(path), make_repo_relative(path))
    ensure_schema(document, schema_path, make_repo_relative(path))
    return document
