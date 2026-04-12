#!/usr/bin/env python3
from __future__ import annotations

import json
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
VALIDATOR_PATH = REPO_ROOT / "scripts" / "validate-radishflow-export-snapshot.py"


def build_base_snapshot() -> dict[str, Any]:
    return {
        "schema_version": 1,
        "request_id": "rf-control-plane-validator-support-artifacts-001",
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
    }


def write_snapshot(path: Path, document: dict[str, Any]) -> None:
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def run_validator(path: Path, *extra_args: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [sys.executable, str(VALIDATOR_PATH), "--input", str(path), *extra_args],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
    )


def require_contains(output: str, needle: str, label: str) -> None:
    if needle not in output:
        raise SystemExit(f"{label}: expected output to contain {needle!r}, got: {output!r}")


def main() -> int:
    with tempfile.TemporaryDirectory(prefix="radishflow-export-validator-support-artifacts-") as temp_dir:
        temp_root = Path(temp_dir)

        sanitized_summary_snapshot = build_base_snapshot()
        sanitized_summary_snapshot["support_artifacts"] = [
            {
                "kind": "attachment_ref",
                "role": "supporting",
                "name": "private_registry_capture",
                "mime_type": "application/json",
                "uri": "capture://radishflow/control-plane/private-registry-429-validator",
                "metadata": {
                    "sanitized_summary": "private source returned 429; bodies and auth headers were stripped",
                },
            }
        ]
        sanitized_summary_path = temp_root / "sanitized-summary.export.json"
        write_snapshot(sanitized_summary_path, sanitized_summary_snapshot)
        sanitized_summary_result = run_validator(sanitized_summary_path)
        if sanitized_summary_result.returncode != 0:
            raise SystemExit(
                "validator should accept sanitized_summary-only support artifact metadata; "
                f"stdout={sanitized_summary_result.stdout!r} stderr={sanitized_summary_result.stderr!r}"
            )
        require_contains(
            sanitized_summary_result.stdout,
            "validation passed",
            "sanitized_summary-only support artifact regression",
        )

        warning_snapshot = build_base_snapshot()
        warning_snapshot["support_artifacts"] = [
            {
                "kind": "attachment_ref",
                "role": "supporting",
                "name": "summary_missing_capture",
                "mime_type": "application/json",
                "uri": "capture://radishflow/control-plane/summary-missing-validator",
                "metadata": {
                    "source_scope": "control_plane",
                },
            }
        ]
        warning_path = temp_root / "warning-summary-missing.export.json"
        write_snapshot(warning_path, warning_snapshot)
        warning_result = run_validator(warning_path)
        if warning_result.returncode != 0:
            raise SystemExit(
                "validator should warn but not fail for uri-only support artifact without minimal summary; "
                f"stdout={warning_result.stdout!r} stderr={warning_result.stderr!r}"
            )
        require_contains(warning_result.stdout, "passed with warnings", "warning support artifact regression")
        require_contains(
            warning_result.stdout,
            "prefer uri + minimal redacted summary",
            "warning support artifact regression",
        )

        warning_strict_result = run_validator(warning_path, "--strict")
        if warning_strict_result.returncode == 0:
            raise SystemExit("validator --strict should fail when uri-only support artifact is missing minimal summary")
        require_contains(
            warning_strict_result.stdout,
            "passed with warnings",
            "strict warning support artifact regression",
        )

        raw_payload_snapshot = build_base_snapshot()
        raw_payload_snapshot["support_artifacts"] = [
            {
                "kind": "json",
                "role": "supporting",
                "name": "raw_payload_capture",
                "mime_type": "application/json",
                "content": {
                    "raw_payload": {
                        "status": 429,
                    }
                },
                "metadata": {
                    "summary": "should fail because raw payload leaked into export snapshot",
                },
            }
        ]
        raw_payload_path = temp_root / "raw-payload.export.json"
        write_snapshot(raw_payload_path, raw_payload_snapshot)
        raw_payload_result = run_validator(raw_payload_path)
        if raw_payload_result.returncode == 0:
            raise SystemExit("validator should fail when support_artifacts.content includes raw_payload-like keys")
        require_contains(
            raw_payload_result.stdout,
            "validation failed",
            "raw payload support artifact regression",
        )
        require_contains(
            raw_payload_result.stdout,
            "looks like a raw payload field",
            "raw payload support artifact regression",
        )

    print("radishflow export validator support_artifacts regression passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
