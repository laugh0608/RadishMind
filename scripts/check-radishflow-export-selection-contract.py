#!/usr/bin/env python3
from __future__ import annotations

import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from adapters.radishflow import (  # noqa: E402
    build_adapter_snapshot_from_export_snapshot,
    build_request_from_export_snapshot,
)
from services.runtime.candidate_records import ensure_schema, load_json_document  # noqa: E402


EXPORT_SCHEMA_PATH = REPO_ROOT / "contracts" / "radishflow-export-snapshot.schema.json"
ADAPTER_SCHEMA_PATH = REPO_ROOT / "contracts" / "radishflow-adapter-snapshot.schema.json"
REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts" / "copilot-request.schema.json"

CASES = (
    {
        "label": "stream-only-selection-does-not-infer-primary-unit",
        "export_path": REPO_ROOT / "adapters/radishflow/exports/suggest-flowsheet-edits-reconnect-outlet-001.export.json",
        "expected_adapter_path": REPO_ROOT / "adapters/radishflow/examples/suggest-flowsheet-edits-reconnect-outlet-001.snapshot.json",
        "expected_request_path": REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json",
        "expect_selected_unit": False,
    },
    {
        "label": "joint-selection-without-primary-unit-does-not-infer-selected-unit",
        "export_path": REPO_ROOT / "adapters/radishflow/exports/suggest-flowsheet-edits-multi-selection-single-actionable-target-001.export.json",
        "expected_adapter_path": REPO_ROOT / "adapters/radishflow/examples/suggest-flowsheet-edits-multi-selection-single-actionable-target-001.snapshot.json",
        "expected_request_path": REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-multi-selection-single-actionable-target-001.json",
        "expect_selected_unit": False,
    },
    {
        "label": "joint-selection-primary-focus",
        "export_path": REPO_ROOT / "adapters/radishflow/exports/suggest-flowsheet-edits-joint-selection-primary-focus-001.export.json",
        "expected_adapter_path": REPO_ROOT / "adapters/radishflow/examples/suggest-flowsheet-edits-joint-selection-primary-focus-001.snapshot.json",
        "expected_request_path": REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-joint-selection-primary-focus-001.json",
        "expect_selected_unit": True,
    },
    {
        "label": "multi-unit-stream-primary-focus",
        "export_path": REPO_ROOT / "adapters/radishflow/exports/suggest-flowsheet-edits-multi-unit-stream-primary-focus-001.export.json",
        "expected_adapter_path": REPO_ROOT / "adapters/radishflow/examples/suggest-flowsheet-edits-multi-unit-stream-primary-focus-001.snapshot.json",
        "expected_request_path": REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-multi-unit-stream-primary-focus-001.json",
        "expect_selected_unit": True,
    },
    {
        "label": "selection-order-preserved",
        "export_path": REPO_ROOT / "adapters/radishflow/exports/suggest-flowsheet-edits-selection-order-preserved-001.export.json",
        "expected_adapter_path": REPO_ROOT / "adapters/radishflow/examples/suggest-flowsheet-edits-selection-order-preserved-001.snapshot.json",
        "expected_request_path": REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-selection-order-preserved-001.json",
        "expect_selected_unit": True,
    },
    {
        "label": "selection-chronology-separate-from-primary-focus",
        "export_path": REPO_ROOT / "adapters/radishflow/exports/suggest-flowsheet-edits-selection-chronology-single-actionable-target-001.export.json",
        "expected_adapter_path": REPO_ROOT / "adapters/radishflow/examples/suggest-flowsheet-edits-selection-chronology-single-actionable-target-001.snapshot.json",
        "expected_request_path": REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-selection-chronology-single-actionable-target-001.json",
        "expect_selected_unit": True,
        "expect_primary_not_first": True,
    },
)


def require_equal(actual: object, expected: object, label: str) -> None:
    if actual != expected:
        raise SystemExit(f"{label}: expected {expected!r}, got {actual!r}")


def load_json_object(path: Path, label: str) -> dict[str, Any]:
    document = load_json_document(path)
    if not isinstance(document, dict):
        raise SystemExit(f"{label}: expected a json object")
    return document


def request_context_projection(request: dict[str, Any]) -> dict[str, Any]:
    context = request.get("context")
    if not isinstance(context, dict):
        raise SystemExit("generated request: expected context to be an object")
    return context


def verify_selection_lists(
    export_selection: dict[str, Any],
    adapter_snapshot: dict[str, Any],
    request_context: dict[str, Any],
    *,
    label: str,
) -> None:
    for export_key, adapter_key, request_key in (
        ("selected_unit_ids", "selected_unit_ids", "selected_unit_ids"),
        ("selected_stream_ids", "selected_stream_ids", "selected_stream_ids"),
    ):
        export_value = export_selection.get(export_key)
        adapter_value = adapter_snapshot.get(adapter_key)
        request_value = request_context.get(request_key)
        require_equal(adapter_value, export_value, f"{label} export -> adapter {adapter_key}")
        require_equal(request_value, export_value, f"{label} export -> request {request_key}")


def verify_primary_focus(
    export_selection: dict[str, Any],
    adapter_snapshot: dict[str, Any],
    request_context: dict[str, Any],
    *,
    expect_selected_unit: bool,
    expect_primary_not_first: bool,
    label: str,
) -> None:
    primary_selected_unit = export_selection.get("primary_selected_unit")
    adapter_selected_unit = adapter_snapshot.get("selected_unit")
    request_selected_unit = request_context.get("selected_unit")

    if expect_selected_unit:
        if not isinstance(primary_selected_unit, dict) or not primary_selected_unit:
            raise SystemExit(f"{label}: expected export snapshot to include primary_selected_unit")
        require_equal(
            adapter_selected_unit,
            primary_selected_unit,
            f"{label} export -> adapter selected_unit passthrough",
        )
        require_equal(
            request_selected_unit,
            primary_selected_unit,
            f"{label} export -> request selected_unit passthrough",
        )

        selected_unit_ids = export_selection.get("selected_unit_ids")
        if not isinstance(selected_unit_ids, list) or not selected_unit_ids:
            raise SystemExit(f"{label}: expected non-empty selected_unit_ids for primary focus regression")
        primary_id = str(primary_selected_unit.get("id") or "").strip()
        if primary_id not in {str(unit_id).strip() for unit_id in selected_unit_ids}:
            raise SystemExit(f"{label}: primary_selected_unit.id must remain inside selected_unit_ids")
        if expect_primary_not_first and primary_id == str(selected_unit_ids[0]).strip():
            raise SystemExit(f"{label}: expected primary_selected_unit.id to differ from selection chronology head")
        return

    if primary_selected_unit is not None:
        raise SystemExit(f"{label}: did not expect export snapshot to include primary_selected_unit")
    if "selected_unit" in adapter_snapshot:
        raise SystemExit(f"{label}: adapter snapshot should not infer selected_unit without primary_selected_unit")
    if "selected_unit" in request_context:
        raise SystemExit(f"{label}: request context should not infer selected_unit without primary_selected_unit")


def verify_checked_in_outputs(
    adapter_snapshot: dict[str, Any],
    request_context: dict[str, Any],
    expected_adapter_snapshot: dict[str, Any],
    expected_request: dict[str, Any],
    *,
    label: str,
) -> None:
    expected_request_context = expected_request.get("context")
    if not isinstance(expected_request_context, dict):
        raise SystemExit(f"{label}: expected checked-in request context to be an object")

    for key in ("selected_unit_ids", "selected_stream_ids", "selected_unit"):
        require_equal(
            adapter_snapshot.get(key),
            expected_adapter_snapshot.get(key),
            f"{label} checked-in adapter {key}",
        )
        require_equal(
            request_context.get(key),
            expected_request_context.get(key),
            f"{label} checked-in request context {key}",
        )


def main() -> int:
    for case in CASES:
        export_snapshot = load_json_object(case["export_path"], f"{case['label']} export snapshot")
        ensure_schema(export_snapshot, EXPORT_SCHEMA_PATH, f"{case['label']} export snapshot")

        expected_adapter_snapshot = load_json_object(case["expected_adapter_path"], f"{case['label']} adapter snapshot")
        ensure_schema(expected_adapter_snapshot, ADAPTER_SCHEMA_PATH, f"{case['label']} checked-in adapter snapshot")

        expected_request_sample = load_json_object(case["expected_request_path"], f"{case['label']} eval sample")
        expected_request = expected_request_sample.get("input_request")
        if not isinstance(expected_request, dict):
            raise SystemExit(f"{case['label']} eval sample: expected input_request to be an object")
        ensure_schema(expected_request, REQUEST_SCHEMA_PATH, f"{case['label']} checked-in request")

        adapter_snapshot = build_adapter_snapshot_from_export_snapshot(export_snapshot)
        ensure_schema(adapter_snapshot, ADAPTER_SCHEMA_PATH, f"{case['label']} generated adapter snapshot")

        request = build_request_from_export_snapshot(export_snapshot)
        ensure_schema(request, REQUEST_SCHEMA_PATH, f"{case['label']} generated request")

        export_selection = export_snapshot.get("selection_state")
        if not isinstance(export_selection, dict):
            raise SystemExit(f"{case['label']}: expected export snapshot selection_state to be an object")
        request_context = request_context_projection(request)

        verify_selection_lists(
            export_selection,
            adapter_snapshot,
            request_context,
            label=case["label"],
        )
        verify_primary_focus(
            export_selection,
            adapter_snapshot,
            request_context,
            expect_selected_unit=bool(case["expect_selected_unit"]),
            expect_primary_not_first=bool(case.get("expect_primary_not_first")),
            label=case["label"],
        )
        verify_checked_in_outputs(
            adapter_snapshot,
            request_context,
            expected_adapter_snapshot,
            expected_request,
            label=case["label"],
        )

    print("radishflow export selection contract regression passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
