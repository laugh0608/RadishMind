#!/usr/bin/env python3
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import ensure_schema, load_json_document, make_repo_relative, resolve_relative_to_repo  # noqa: E402


EXPORT_SNAPSHOT_SCHEMA_PATH = REPO_ROOT / "contracts/radishflow-export-snapshot.schema.json"
SENSITIVE_KEY_TOKENS = (
    "token",
    "secret",
    "cookie",
    "credential",
    "authorization",
    "api_key",
    "access_key",
    "refresh_key",
    "refresh_token",
    "private_key",
    "license_key",
)
SENSITIVE_VALUE_PATTERNS = (
    re.compile(r"\bBearer\s+[A-Za-z0-9._\-+/=]{12,}"),
    re.compile(r"\bsk-[A-Za-z0-9]{16,}\b"),
    re.compile(r"\bAIza[0-9A-Za-z\-_]{20,}\b"),
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Validate a RadishFlow export snapshot before wiring it into the runtime path. "
            "Checks schema, task-specific semantics, and common sensitive-data mistakes."
        ),
    )
    parser.add_argument("--input", required=True, help="Path to a radishflow export snapshot json file.")
    parser.add_argument(
        "--strict",
        action="store_true",
        help="Fail on warnings as well as errors.",
    )
    return parser.parse_args()


def _is_non_empty_list(value: Any) -> bool:
    return isinstance(value, list) and len(value) > 0


def _scan_sensitive_content(value: Any, path: str, errors: list[str]) -> None:
    if isinstance(value, dict):
        for key, item in value.items():
            normalized_key = str(key).strip().lower().replace("-", "_")
            if any(token in normalized_key for token in SENSITIVE_KEY_TOKENS):
                errors.append(f"{path}.{key}: contains sensitive-looking key name '{key}'")
            _scan_sensitive_content(item, f"{path}.{key}", errors)
        return

    if isinstance(value, list):
        for index, item in enumerate(value):
            _scan_sensitive_content(item, f"{path}[{index}]", errors)
        return

    if isinstance(value, str):
        for pattern in SENSITIVE_VALUE_PATTERNS:
            if pattern.search(value):
                errors.append(f"{path}: contains sensitive-looking credential material")
                break


def _validate_selection_semantics(selection_state: dict[str, Any], errors: list[str], warnings: list[str]) -> None:
    selected_unit_ids = selection_state.get("selected_unit_ids")
    primary_selected_unit = selection_state.get("primary_selected_unit")

    if primary_selected_unit is None:
        return

    if not isinstance(primary_selected_unit, dict) or not primary_selected_unit:
        errors.append("selection_state.primary_selected_unit must be a non-empty object when provided")
        return

    if not _is_non_empty_list(selected_unit_ids):
        errors.append("selection_state.primary_selected_unit requires selection_state.selected_unit_ids")
        return

    if len(selected_unit_ids) != 1:
        errors.append("selection_state.primary_selected_unit requires exactly one selected_unit_id")
        return

    primary_id = str(primary_selected_unit.get("id") or "").strip()
    selected_id = str(selected_unit_ids[0]).strip()
    if primary_id and primary_id != selected_id:
        errors.append("selection_state.primary_selected_unit.id does not match selection_state.selected_unit_ids[0]")
    if not primary_id:
        warnings.append("selection_state.primary_selected_unit is missing id; adapter can only treat it as opaque context")


def _validate_task_specific_semantics(snapshot: dict[str, Any], errors: list[str], warnings: list[str]) -> None:
    task = str(snapshot.get("task") or "").strip()
    document_state = snapshot.get("document_state") or {}
    selection_state = snapshot.get("selection_state") or {}
    diagnostics_export = snapshot.get("diagnostics_export") or {}

    if task in {"explain_diagnostics", "suggest_flowsheet_edits"}:
        if "flowsheet_document" in document_state and "flowsheet_document_uri" in document_state:
            warnings.append("document_state simultaneously includes flowsheet_document and flowsheet_document_uri")

        if not (_is_non_empty_list(selection_state.get("selected_unit_ids")) or _is_non_empty_list(selection_state.get("selected_stream_ids"))):
            errors.append(f"{task} requires at least one non-empty selection list")

        if not ("diagnostic_summary" in diagnostics_export or _is_non_empty_list(diagnostics_export.get("diagnostics"))):
            errors.append(f"{task} requires diagnostics_export.diagnostic_summary or diagnostics_export.diagnostics")

        _validate_selection_semantics(selection_state, errors, warnings)

        if task == "suggest_flowsheet_edits":
            if _is_non_empty_list(selection_state.get("selected_unit_ids")) and _is_non_empty_list(selection_state.get("selected_stream_ids")):
                warnings.append(
                    "suggest_flowsheet_edits carries both selected_unit_ids and selected_stream_ids; "
                    "confirm the exporter is preserving a true multi-selection state rather than duplicating focus"
                )
        return

    if task == "explain_control_plane_state":
        if "selection_state" in snapshot:
            warnings.append("explain_control_plane_state usually does not need selection_state")


def main() -> int:
    args = parse_args()
    input_path = resolve_relative_to_repo(args.input)
    if not input_path.is_file():
        raise SystemExit(f"input file not found: {args.input}")

    snapshot = load_json_document(input_path)
    ensure_schema(snapshot, EXPORT_SNAPSHOT_SCHEMA_PATH, make_repo_relative(input_path))

    errors: list[str] = []
    warnings: list[str] = []
    _validate_task_specific_semantics(snapshot, errors, warnings)
    _scan_sensitive_content(snapshot, "$", errors)

    label = make_repo_relative(input_path)
    if errors:
        print(f"radishflow export snapshot validation failed: {label}")
        for issue in errors:
            print(f"ERROR: {issue}")
        for issue in warnings:
            print(f"WARNING: {issue}")
        raise SystemExit(1)

    if warnings:
        print(f"radishflow export snapshot validation passed with warnings: {label}")
        for issue in warnings:
            print(f"WARNING: {issue}")
        if args.strict:
            raise SystemExit(1)
        return 0

    print(f"radishflow export snapshot validation passed: {label}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
