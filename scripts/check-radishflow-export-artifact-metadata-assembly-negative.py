#!/usr/bin/env python3
from __future__ import annotations

import json
import subprocess
import sys
import tempfile
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


VALIDATOR_PATH = REPO_ROOT / "scripts" / "validate-radishflow-export-snapshot.py"
EXPORT_SCHEMA_PATH = REPO_ROOT / "contracts" / "radishflow-export-snapshot.schema.json"
REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts" / "copilot-request.schema.json"
MIXED_SUMMARY_EXPORT_PATH = REPO_ROOT / "adapters/radishflow/exports/explain-control-plane-mixed-summary-variants-001.export.json"
REDACTED_EXPORT_PATH = REPO_ROOT / "adapters/radishflow/exports/explain-control-plane-redacted-support-summary-001.export.json"


def load_json_object(path: Path, label: str) -> dict[str, Any]:
    document = load_json_document(path)
    if not isinstance(document, dict):
        raise SystemExit(f"{label}: expected a json object")
    return document


def write_json_document(path: Path, document: dict[str, Any]) -> None:
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def run_export_validator(path: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [sys.executable, str(VALIDATOR_PATH), "--input", str(path)],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
    )


def combined_output(result: subprocess.CompletedProcess[str]) -> str:
    return f"{result.stdout}{result.stderr}"


def require_contains(output: str, needle: str, label: str) -> None:
    if needle not in output:
        raise SystemExit(f"{label}: expected output to contain {needle!r}, got {output!r}")


def expect_request_schema_failure(document: dict[str, Any], label: str, needle: str) -> None:
    try:
        ensure_schema(document, REQUEST_SCHEMA_PATH, label)
    except SystemExit as exc:
        message = str(exc)
        if needle not in message:
            raise SystemExit(f"{label}: expected request schema failure containing {needle!r}, got {message!r}") from exc
        return
    raise SystemExit(f"{label}: expected request schema validation to fail")


def find_named_artifact(artifacts: list[dict[str, Any]], artifact_name: str, label: str) -> dict[str, Any]:
    for artifact in artifacts:
        if str(artifact.get("name") or "").strip() == artifact_name:
            return artifact
    raise SystemExit(f"{label}: artifact {artifact_name!r} not found")


def require_metadata_field_match(
    export_snapshot: dict[str, Any],
    stage_document: dict[str, Any],
    *,
    export_artifacts_key: str,
    stage_artifacts_key: str,
    artifact_name: str,
    field_name: str,
    label: str,
) -> None:
    export_artifacts = export_snapshot.get(export_artifacts_key) or []
    stage_artifacts = stage_document.get(stage_artifacts_key) or []
    if not isinstance(export_artifacts, list) or not isinstance(stage_artifacts, list):
        raise SystemExit(f"{label}: expected artifact lists at both stages")

    export_artifact = find_named_artifact(export_artifacts, artifact_name, f"{label} export stage")
    stage_artifact = find_named_artifact(stage_artifacts, artifact_name, f"{label} stage document")

    export_value = (export_artifact.get("metadata") or {}).get(field_name)
    stage_value = (stage_artifact.get("metadata") or {}).get(field_name)
    if stage_value != export_value:
        raise SystemExit(
            f"{label}: {artifact_name}.metadata.{field_name} drifted across assembly; "
            f"expected {export_value!r}, got {stage_value!r}"
        )


def expect_metadata_field_drift_detected(
    export_snapshot: dict[str, Any],
    stage_document: dict[str, Any],
    *,
    export_artifacts_key: str,
    stage_artifacts_key: str,
    artifact_name: str,
    field_name: str,
    expected_fragment: str,
    label: str,
) -> None:
    try:
        require_metadata_field_match(
            export_snapshot,
            stage_document,
            export_artifacts_key=export_artifacts_key,
            stage_artifacts_key=stage_artifacts_key,
            artifact_name=artifact_name,
            field_name=field_name,
            label=label,
        )
    except SystemExit as exc:
        message = str(exc)
        if expected_fragment not in message:
            raise SystemExit(f"{label}: expected drift failure containing {expected_fragment!r}, got {message!r}") from exc
        return
    raise SystemExit(f"{label}: expected metadata drift to be detected")


def build_base_export_snapshot() -> dict[str, Any]:
    return {
        "schema_version": 1,
        "request_id": "rf-control-plane-negative-artifact-metadata-assembly-001",
        "task": "explain_control_plane_state",
        "locale": "zh-CN",
        "document_state": {
            "document_revision": 1,
        },
        "control_plane_snapshot": {
            "entitlement_status": "active",
            "lease_status": "stale",
            "sync_status": "warning",
            "manifest_status": "current",
        },
        "support_artifacts": [
            {
                "kind": "attachment_ref",
                "role": "supporting",
                "name": "lease_refresh_capture",
                "mime_type": "application/json",
                "uri": "capture://radishflow/control-plane/lease-refresh-negative-001",
                "metadata": {
                    "summary": "lease refresh endpoint returned a bounded 403 summary",
                    "redactions": ["authorization", "cookie"],
                    "source_scope": "control_plane",
                    "summary_variant": "summary_plus_redaction",
                },
            }
        ],
    }


def main() -> int:
    mixed_summary_export = load_json_object(MIXED_SUMMARY_EXPORT_PATH, "mixed summary export fixture")
    ensure_schema(mixed_summary_export, EXPORT_SCHEMA_PATH, "mixed summary export fixture")
    redacted_export = load_json_object(REDACTED_EXPORT_PATH, "redacted export fixture")
    ensure_schema(redacted_export, EXPORT_SCHEMA_PATH, "redacted export fixture")

    with tempfile.TemporaryDirectory(prefix="radishflow-export-artifact-metadata-assembly-negative-") as temp_dir:
        temp_root = Path(temp_dir)

        raw_metadata_snapshot = build_base_export_snapshot()
        raw_metadata_snapshot["support_artifacts"][0]["metadata"]["headers"] = {"authorization": "redacted"}
        raw_metadata_path = temp_root / "raw-metadata.export.json"
        write_json_document(raw_metadata_path, raw_metadata_snapshot)
        raw_metadata_result = run_export_validator(raw_metadata_path)
        if raw_metadata_result.returncode == 0:
            raise SystemExit("export raw metadata regression: expected export validator to fail before assembly")
        require_contains(
            combined_output(raw_metadata_result),
            "'headers' should not be valid",
            "export raw metadata regression",
        )

        redactions_type_drift_snapshot = build_base_export_snapshot()
        redactions_type_drift_snapshot["support_artifacts"][0]["metadata"]["redactions"] = "authorization"
        redactions_type_drift_path = temp_root / "redactions-type-drift.export.json"
        write_json_document(redactions_type_drift_path, redactions_type_drift_snapshot)
        redactions_type_drift_result = run_export_validator(redactions_type_drift_path)
        if redactions_type_drift_result.returncode == 0:
            raise SystemExit("export redactions type drift regression: expected export validator to fail before assembly")
        require_contains(
            combined_output(redactions_type_drift_result),
            "is not of type 'array'",
            "export redactions type drift regression",
        )

    request_raw_payload = build_request_from_export_snapshot(mixed_summary_export)
    request_raw_payload_artifact = find_named_artifact(
        request_raw_payload["artifacts"],
        "private_registry_capture",
        "request raw payload regression",
    )
    request_raw_payload_artifact.setdefault("metadata", {})["raw_payload"] = {"status": 429}
    expect_request_schema_failure(
        request_raw_payload,
        "request raw payload regression",
        "'raw_payload' should not be valid",
    )

    request_redactions_type_drift = build_request_from_export_snapshot(redacted_export)
    request_redactions_artifact = find_named_artifact(
        request_redactions_type_drift["artifacts"],
        "lease_refresh_capture",
        "request redactions type drift regression",
    )
    request_redactions_artifact.setdefault("metadata", {})["redactions"] = "authorization"
    expect_request_schema_failure(
        request_redactions_type_drift,
        "request redactions type drift regression",
        "is not of type 'array'",
    )

    adapter_source_scope_drift = build_adapter_snapshot_from_export_snapshot(mixed_summary_export)
    adapter_source_scope_artifact = find_named_artifact(
        adapter_source_scope_drift["supporting_artifacts"],
        "private_registry_capture",
        "adapter source_scope drift regression",
    )
    adapter_source_scope_artifact.setdefault("metadata", {})["source_scope"] = "workspace"
    expect_metadata_field_drift_detected(
        mixed_summary_export,
        adapter_source_scope_drift,
        export_artifacts_key="support_artifacts",
        stage_artifacts_key="supporting_artifacts",
        artifact_name="private_registry_capture",
        field_name="source_scope",
        expected_fragment="expected 'control_plane', got 'workspace'",
        label="adapter source_scope drift regression",
    )

    request_redactions_drift = build_request_from_export_snapshot(redacted_export)
    request_redactions_drift_artifact = find_named_artifact(
        request_redactions_drift["artifacts"],
        "lease_refresh_capture",
        "request redactions drift regression",
    )
    request_redactions_drift_artifact.setdefault("metadata", {})["redactions"] = ["authorization"]
    expect_metadata_field_drift_detected(
        redacted_export,
        {
            "artifacts": [
                artifact
                for artifact in request_redactions_drift["artifacts"]
                if str(artifact.get("role") or "").strip() == "supporting"
            ]
        },
        export_artifacts_key="support_artifacts",
        stage_artifacts_key="artifacts",
        artifact_name="lease_refresh_capture",
        field_name="redactions",
        expected_fragment="expected ['authorization', 'cookie', 'token'], got ['authorization']",
        label="request redactions drift regression",
    )

    request_summary_variant_drift = build_request_from_export_snapshot(mixed_summary_export)
    request_summary_variant_artifact = find_named_artifact(
        request_summary_variant_drift["artifacts"],
        "lease_cache_capture",
        "request summary_variant drift regression",
    )
    request_summary_variant_artifact.setdefault("metadata", {})["summary_variant"] = "sanitized_only"
    expect_metadata_field_drift_detected(
        mixed_summary_export,
        {
            "artifacts": [
                artifact
                for artifact in request_summary_variant_drift["artifacts"]
                if str(artifact.get("role") or "").strip() == "supporting"
            ]
        },
        export_artifacts_key="support_artifacts",
        stage_artifacts_key="artifacts",
        artifact_name="lease_cache_capture",
        field_name="summary_variant",
        expected_fragment="expected 'summary_plus_redaction', got 'sanitized_only'",
        label="request summary_variant drift regression",
    )

    print("radishflow export artifact metadata negative assembly regression passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
