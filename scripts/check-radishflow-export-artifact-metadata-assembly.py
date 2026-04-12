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
        "label": "mixed-summary-variants",
        "export_path": REPO_ROOT / "adapters/radishflow/exports/explain-control-plane-mixed-summary-variants-001.export.json",
        "adapter_path": REPO_ROOT / "adapters/radishflow/examples/explain-control-plane-mixed-summary-variants-001.snapshot.json",
        "eval_sample_path": REPO_ROOT / "datasets/eval/radishflow/explain-control-plane-mixed-summary-variants-001.json",
        "focused_metadata_fields": {
            "private_registry_capture": ("source_scope", "summary_variant"),
            "lease_cache_capture": ("source_scope", "summary_variant"),
        },
    },
    {
        "label": "redacted-support-summary",
        "export_path": REPO_ROOT / "adapters/radishflow/exports/explain-control-plane-redacted-support-summary-001.export.json",
        "adapter_path": REPO_ROOT / "adapters/radishflow/examples/explain-control-plane-redacted-support-summary-001.snapshot.json",
        "eval_sample_path": REPO_ROOT / "datasets/eval/radishflow/explain-control-plane-redacted-support-summary-001.json",
        "focused_metadata_fields": {
            "lease_refresh_capture": ("redactions",),
        },
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


def find_named_artifact(artifacts: list[dict[str, Any]], artifact_name: str, label: str) -> dict[str, Any]:
    for artifact in artifacts:
        if str(artifact.get("name") or "").strip() == artifact_name:
            return artifact
    raise SystemExit(f"{label}: artifact {artifact_name!r} not found")


def metadata_by_artifact_name(artifacts: list[dict[str, Any]], label: str) -> dict[str, dict[str, Any]]:
    result: dict[str, dict[str, Any]] = {}
    for artifact in artifacts:
        name = str(artifact.get("name") or "").strip()
        if not name:
            raise SystemExit(f"{label}: encountered artifact without a name")
        metadata = artifact.get("metadata")
        if metadata is None:
            continue
        if not isinstance(metadata, dict):
            raise SystemExit(f"{label}: artifact {name!r} metadata must be an object when present")
        result[name] = metadata
    return result


def main() -> int:
    for case in CASES:
        export_snapshot = load_json_object(case["export_path"], f"{case['label']} export snapshot")
        ensure_schema(export_snapshot, EXPORT_SCHEMA_PATH, f"{case['label']} export snapshot")

        expected_adapter_snapshot = load_json_object(case["adapter_path"], f"{case['label']} adapter snapshot")
        ensure_schema(expected_adapter_snapshot, ADAPTER_SCHEMA_PATH, f"{case['label']} checked-in adapter snapshot")

        eval_sample = load_json_object(case["eval_sample_path"], f"{case['label']} eval sample")
        expected_request = eval_sample.get("input_request")
        if not isinstance(expected_request, dict):
            raise SystemExit(f"{case['label']} eval sample: expected input_request to be an object")
        ensure_schema(expected_request, REQUEST_SCHEMA_PATH, f"{case['label']} checked-in input_request")

        generated_adapter_snapshot = build_adapter_snapshot_from_export_snapshot(export_snapshot)
        ensure_schema(generated_adapter_snapshot, ADAPTER_SCHEMA_PATH, f"{case['label']} generated adapter snapshot")
        require_equal(
            generated_adapter_snapshot,
            expected_adapter_snapshot,
            f"{case['label']} export -> adapter snapshot assembly",
        )

        generated_request = build_request_from_export_snapshot(export_snapshot)
        ensure_schema(generated_request, REQUEST_SCHEMA_PATH, f"{case['label']} generated request")
        require_equal(
            generated_request,
            expected_request,
            f"{case['label']} export -> request assembly",
        )

        export_artifacts = export_snapshot.get("support_artifacts") or []
        adapter_artifacts = generated_adapter_snapshot.get("supporting_artifacts") or []
        request_artifacts = [
            artifact
            for artifact in generated_request.get("artifacts") or []
            if str(artifact.get("role") or "").strip() == "supporting"
        ]
        if not isinstance(export_artifacts, list) or not isinstance(adapter_artifacts, list) or not isinstance(request_artifacts, list):
            raise SystemExit(f"{case['label']}: supporting artifacts must remain lists across assembly stages")

        export_metadata = metadata_by_artifact_name(export_artifacts, f"{case['label']} export artifacts")
        adapter_metadata = metadata_by_artifact_name(adapter_artifacts, f"{case['label']} adapter artifacts")
        request_metadata = metadata_by_artifact_name(request_artifacts, f"{case['label']} request artifacts")
        require_equal(
            sorted(export_metadata),
            sorted(adapter_metadata),
            f"{case['label']} export -> adapter metadata artifact set",
        )
        require_equal(
            sorted(export_metadata),
            sorted(request_metadata),
            f"{case['label']} export -> request metadata artifact set",
        )

        for artifact_name, export_metadata_entry in export_metadata.items():
            require_equal(
                adapter_metadata.get(artifact_name),
                export_metadata_entry,
                f"{case['label']} export -> adapter metadata passthrough for {artifact_name}",
            )
            require_equal(
                request_metadata.get(artifact_name),
                export_metadata_entry,
                f"{case['label']} export -> request metadata passthrough for {artifact_name}",
            )

        for artifact_name, field_names in case["focused_metadata_fields"].items():
            export_artifact = find_named_artifact(export_artifacts, artifact_name, f"{case['label']} export artifacts")
            adapter_artifact = find_named_artifact(adapter_artifacts, artifact_name, f"{case['label']} adapter artifacts")
            request_artifact = find_named_artifact(request_artifacts, artifact_name, f"{case['label']} request artifacts")
            for field_name in field_names:
                export_value = (export_artifact.get("metadata") or {}).get(field_name)
                adapter_value = (adapter_artifact.get("metadata") or {}).get(field_name)
                request_value = (request_artifact.get("metadata") or {}).get(field_name)
                require_equal(
                    adapter_value,
                    export_value,
                    f"{case['label']} export -> adapter field {artifact_name}.{field_name}",
                )
                require_equal(
                    request_value,
                    export_value,
                    f"{case['label']} export -> request field {artifact_name}.{field_name}",
                )

    print("radishflow export artifact metadata assembly regression passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
