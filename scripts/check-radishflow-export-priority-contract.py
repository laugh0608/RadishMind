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
        "label": "single-actionable-target",
        "export_path": REPO_ROOT / "adapters/radishflow/exports/suggest-flowsheet-edits-multi-selection-single-actionable-target-001.export.json",
        "expected_adapter_path": REPO_ROOT / "adapters/radishflow/examples/suggest-flowsheet-edits-multi-selection-single-actionable-target-001.snapshot.json",
        "expected_request_path": REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-multi-selection-single-actionable-target-001.json",
        "expected_actionable_target": {"type": "stream", "id": "feed-6"},
        "disallowed_target_ids": ("heater-2",),
    },
    {
        "label": "same-risk-input-first-ordering",
        "export_path": REPO_ROOT / "adapters/radishflow/exports/suggest-flowsheet-edits-same-risk-input-first-ordering-001.export.json",
        "expected_adapter_path": REPO_ROOT / "adapters/radishflow/examples/suggest-flowsheet-edits-same-risk-input-first-ordering-001.snapshot.json",
        "expected_request_path": REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-same-risk-input-first-ordering-001.json",
    },
    {
        "label": "three-step-priority-chain",
        "export_path": REPO_ROOT / "adapters/radishflow/exports/suggest-flowsheet-edits-three-step-priority-chain-001.export.json",
        "expected_adapter_path": REPO_ROOT / "adapters/radishflow/examples/suggest-flowsheet-edits-three-step-priority-chain-001.snapshot.json",
        "expected_request_path": REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-three-step-priority-chain-001.json",
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


def request_context_projection(request: dict[str, Any], label: str) -> dict[str, Any]:
    context = request.get("context")
    if not isinstance(context, dict):
        raise SystemExit(f"{label}: expected request context to be an object")
    return context


def diagnostics_projection(diagnostics: Any, label: str) -> list[dict[str, Any]]:
    if not isinstance(diagnostics, list):
        raise SystemExit(f"{label}: expected diagnostics to be a list")
    projected: list[dict[str, Any]] = []
    for index, diagnostic in enumerate(diagnostics):
        if not isinstance(diagnostic, dict):
            raise SystemExit(f"{label}: diagnostics[{index}] must be an object")
        projected.append(diagnostic)
    return projected


def diagnostic_codes(diagnostics: list[dict[str, Any]]) -> list[str]:
    return [str(diagnostic.get("code") or "").strip() for diagnostic in diagnostics]


def diagnostic_targets(diagnostics: list[dict[str, Any]]) -> list[dict[str, str]]:
    return [
        {
            "type": str(diagnostic.get("target_type") or "").strip(),
            "id": str(diagnostic.get("target_id") or "").strip(),
        }
        for diagnostic in diagnostics
    ]


def verify_diagnostics_passthrough(
    export_snapshot: dict[str, Any],
    adapter_snapshot: dict[str, Any],
    request_context: dict[str, Any],
    expected_adapter_snapshot: dict[str, Any],
    expected_request: dict[str, Any],
    *,
    label: str,
) -> tuple[list[dict[str, Any]], list[dict[str, Any]], list[dict[str, Any]]]:
    export_diagnostics = diagnostics_projection(
        (export_snapshot.get("diagnostics_export") or {}).get("diagnostics"),
        f"{label} export diagnostics",
    )
    adapter_diagnostics = diagnostics_projection(adapter_snapshot.get("diagnostics"), f"{label} adapter diagnostics")
    request_diagnostics = diagnostics_projection(request_context.get("diagnostics"), f"{label} request diagnostics")
    expected_adapter_diagnostics = diagnostics_projection(
        expected_adapter_snapshot.get("diagnostics"),
        f"{label} checked-in adapter diagnostics",
    )
    expected_request_diagnostics = diagnostics_projection(
        (expected_request.get("context") or {}).get("diagnostics"),
        f"{label} checked-in request diagnostics",
    )

    require_equal(adapter_diagnostics, export_diagnostics, f"{label} export -> adapter diagnostics passthrough")
    require_equal(request_diagnostics, export_diagnostics, f"{label} export -> request diagnostics passthrough")
    require_equal(
        adapter_diagnostics,
        expected_adapter_diagnostics,
        f"{label} checked-in adapter diagnostics alignment",
    )
    require_equal(
        request_diagnostics,
        expected_request_diagnostics,
        f"{label} checked-in request diagnostics alignment",
    )
    return export_diagnostics, adapter_diagnostics, request_diagnostics


def verify_selection_passthrough(
    export_snapshot: dict[str, Any],
    adapter_snapshot: dict[str, Any],
    request_context: dict[str, Any],
    *,
    label: str,
) -> None:
    export_selection = export_snapshot.get("selection_state")
    if not isinstance(export_selection, dict):
        raise SystemExit(f"{label}: expected selection_state to be an object")
    for key in ("selected_unit_ids", "selected_stream_ids"):
        export_value = export_selection.get(key)
        require_equal(adapter_snapshot.get(key), export_value, f"{label} export -> adapter {key}")
        require_equal(request_context.get(key), export_value, f"{label} export -> request {key}")


def verify_order_contract(
    diagnostics: list[dict[str, Any]],
    expected_request_sample: dict[str, Any],
    *,
    label: str,
) -> None:
    evaluation = expected_request_sample.get("evaluation")
    if not isinstance(evaluation, dict):
        raise SystemExit(f"{label}: expected eval sample to contain evaluation")

    expected_issue_codes = evaluation.get("ordered_issue_codes")
    if isinstance(expected_issue_codes, list):
        require_equal(
            diagnostic_codes(diagnostics),
            [str(code) for code in expected_issue_codes],
            f"{label} diagnostics -> ordered_issue_codes",
        )

    expected_action_targets = evaluation.get("ordered_action_targets")
    if isinstance(expected_action_targets, list):
        require_equal(
            diagnostic_targets(diagnostics),
            [
                {
                    "type": str((target or {}).get("type") or "").strip(),
                    "id": str((target or {}).get("id") or "").strip(),
                }
                for target in expected_action_targets
            ],
            f"{label} diagnostics -> ordered_action_targets",
        )


def verify_single_actionable_target(
    diagnostics: list[dict[str, Any]],
    expected_request_sample: dict[str, Any],
    *,
    expected_actionable_target: dict[str, str],
    disallowed_target_ids: tuple[str, ...],
    label: str,
) -> None:
    evaluation = expected_request_sample.get("evaluation")
    if not isinstance(evaluation, dict):
        raise SystemExit(f"{label}: expected eval sample to contain evaluation")

    expected_action_targets = evaluation.get("ordered_action_targets")
    if expected_action_targets is not None:
        raise SystemExit(f"{label}: single actionable target fixture should not expose ordered_action_targets")

    diagnostics_targets = diagnostic_targets(diagnostics)
    if expected_actionable_target not in diagnostics_targets:
        raise SystemExit(
            f"{label}: expected actionable target {expected_actionable_target!r} to remain present in diagnostics order seed"
        )

    must_not_have_json_paths = evaluation.get("must_not_have_json_paths")
    if not isinstance(must_not_have_json_paths, list):
        raise SystemExit(f"{label}: expected must_not_have_json_paths in eval sample")
    for target_id in disallowed_target_ids:
        expected_fragment = f'$.proposed_actions[0].target.id="{target_id}"'
        if expected_fragment not in must_not_have_json_paths:
            raise SystemExit(
                f"{label}: expected eval sample to keep disallowed first-action target fragment {expected_fragment!r}"
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
        request_context = request_context_projection(request, case["label"])

        verify_selection_passthrough(
            export_snapshot,
            adapter_snapshot,
            request_context,
            label=case["label"],
        )
        export_diagnostics, _, _ = verify_diagnostics_passthrough(
            export_snapshot,
            adapter_snapshot,
            request_context,
            expected_adapter_snapshot,
            expected_request,
            label=case["label"],
        )
        verify_order_contract(
            export_diagnostics,
            expected_request_sample,
            label=case["label"],
        )
        if "expected_actionable_target" in case:
            verify_single_actionable_target(
                export_diagnostics,
                expected_request_sample,
                expected_actionable_target=case["expected_actionable_target"],
                disallowed_target_ids=tuple(case.get("disallowed_target_ids") or ()),
                label=case["label"],
            )

    print("radishflow export priority contract regression passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
