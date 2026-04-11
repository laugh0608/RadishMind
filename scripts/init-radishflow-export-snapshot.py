#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import ensure_schema, write_json_document  # noqa: E402


EXPORT_SNAPSHOT_SCHEMA_PATH = REPO_ROOT / "contracts/radishflow-export-snapshot.schema.json"
SUPPORTED_TASKS = (
    "explain_diagnostics",
    "suggest_flowsheet_edits",
    "explain_control_plane_state",
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Generate a schema-valid minimal RadishFlow export snapshot template "
            "for upstream exporter integration."
        ),
    )
    parser.add_argument("--task", required=True, choices=SUPPORTED_TASKS, help="Target radishflow task.")
    parser.add_argument("--output", default="", help="Optional output json path. Prints to stdout when omitted.")
    return parser.parse_args()


def build_minimal_export_snapshot(task: str) -> dict[str, Any]:
    if task == "explain_diagnostics":
        return {
            "schema_version": 1,
            "request_id": "rf-explain-diagnostics-template-001",
            "task": task,
            "locale": "zh-CN",
            "document_state": {
                "document_revision": 0,
                "flowsheet_document_uri": "artifact://radishflow/flowsheet/current.json",
            },
            "selection_state": {
                "selected_unit_ids": ["unit-placeholder"],
            },
            "diagnostics_export": {
                "diagnostic_summary": {
                    "error_count": 1,
                    "warning_count": 0,
                },
                "diagnostics": [
                    {
                        "code": "DIAGNOSTIC_CODE_PLACEHOLDER",
                        "message": "Replace with an upstream diagnostic message.",
                        "severity": "warning",
                        "target_type": "unit",
                        "target_id": "unit-placeholder",
                    }
                ],
            },
            "solve_snapshot": {
                "solver_status": "unknown",
                "last_updated": "1970-01-01T00:00:00Z",
            },
        }
    if task == "suggest_flowsheet_edits":
        return {
            "schema_version": 1,
            "request_id": "rf-suggest-flowsheet-edits-template-001",
            "task": task,
            "locale": "zh-CN",
            "document_state": {
                "document_revision": 0,
                "flowsheet_document_uri": "artifact://radishflow/flowsheet/current.json",
            },
            "selection_state": {
                "selected_stream_ids": ["stream-placeholder"],
            },
            "diagnostics_export": {
                "diagnostic_summary": {
                    "error_count": 1,
                    "warning_count": 0,
                },
                "diagnostics": [
                    {
                        "code": "EDIT_DIAGNOSTIC_CODE_PLACEHOLDER",
                        "message": "Replace with an upstream edit-related diagnostic.",
                        "severity": "error",
                        "target_type": "stream",
                        "target_id": "stream-placeholder",
                    }
                ],
            },
            "solve_snapshot": {
                "solver_status": "unknown",
                "last_updated": "1970-01-01T00:00:00Z",
            },
        }
    if task == "explain_control_plane_state":
        return {
            "schema_version": 1,
            "request_id": "rf-explain-control-plane-template-001",
            "task": task,
            "locale": "zh-CN",
            "document_state": {
                "document_revision": 0,
            },
            "control_plane_snapshot": {
                "entitlement_status": "unknown",
                "lease_status": "unknown",
                "sync_status": "unknown",
                "manifest_status": "unknown",
            },
            "solve_session_state": {
                "status": "unknown",
            },
        }
    raise ValueError(f"unsupported task: {task}")


def main() -> int:
    args = parse_args()
    template = build_minimal_export_snapshot(args.task)
    ensure_schema(template, EXPORT_SNAPSHOT_SCHEMA_PATH, f"generated template for {args.task}")

    if args.output.strip():
        output_path = Path(args.output)
        if not output_path.is_absolute():
            output_path = REPO_ROOT / output_path
        write_json_document(output_path, template)
        try:
            output_label = str(output_path.relative_to(REPO_ROOT))
        except ValueError:
            output_label = str(output_path)
        print(f"wrote minimal radishflow export snapshot template to {output_label}")
        return 0

    print(json.dumps(template, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
