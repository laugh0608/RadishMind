#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from adapters.radishflow import build_adapter_snapshot_from_export_snapshot, build_request_from_export_snapshot  # noqa: E402
from services.runtime.candidate_records import (  # noqa: E402
    ensure_schema,
    load_json_document,
    make_repo_relative,
    resolve_relative_to_repo,
    write_json_document,
)


MANIFEST_SCHEMA_PATH = REPO_ROOT / "datasets/eval/radishflow-export-smoke-manifest.schema.json"
SUMMARY_SCHEMA_PATH = REPO_ROOT / "datasets/eval/radishflow-export-smoke-summary.schema.json"
EXPORT_SCHEMA_PATH = REPO_ROOT / "contracts/radishflow-export-snapshot.schema.json"
ADAPTER_SCHEMA_PATH = REPO_ROOT / "contracts/radishflow-adapter-snapshot.schema.json"
REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-request.schema.json"
VALIDATOR_PATH = REPO_ROOT / "scripts" / "validate-radishflow-export-snapshot.py"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Run a batch RadishFlow exporter smoke pass: "
            "validate export snapshots, build adapter snapshots, and build request payloads from a manifest."
        ),
    )
    parser.add_argument("--manifest", required=True, help="Path to a radishflow export smoke manifest json file.")
    parser.add_argument(
        "--output-root",
        default="",
        help="Optional directory where generated adapter snapshots and requests are written for inspection.",
    )
    parser.add_argument(
        "--summary-output",
        default="",
        help="Optional structured smoke summary output path. Prints to stdout when omitted.",
    )
    parser.add_argument(
        "--check-summary",
        default="",
        help="Optional committed summary path. Fails unless the generated summary exactly matches it.",
    )
    parser.add_argument(
        "--strict-validator",
        action="store_true",
        help="Force every case to run export validation in --strict mode.",
    )
    return parser.parse_args()


def load_json_object(path: Path, label: str) -> dict[str, Any]:
    document = load_json_document(path)
    if not isinstance(document, dict):
        raise SystemExit(f"{label}: expected a json object")
    return document


def run_export_validator(path: Path, *, strict: bool) -> subprocess.CompletedProcess[str]:
    command = [sys.executable, str(VALIDATOR_PATH), "--input", str(path)]
    if strict:
        command.append("--strict")
    return subprocess.run(command, cwd=REPO_ROOT, capture_output=True, text=True)


def validator_status(result: subprocess.CompletedProcess[str]) -> str:
    if result.returncode != 0:
        return "failed"
    if "passed with warnings" in result.stdout:
        return "passed_with_warnings"
    return "passed"


def write_generated_documents(
    output_root: Path,
    label: str,
    adapter_snapshot: dict[str, Any],
    request: dict[str, Any],
) -> tuple[str, str]:
    safe_label = label.replace("/", "-").replace(" ", "-")
    adapter_path = output_root / f"{safe_label}.adapter.snapshot.json"
    request_path = output_root / f"{safe_label}.request.json"
    write_json_document(adapter_path, adapter_snapshot)
    write_json_document(request_path, request)
    return make_repo_relative(adapter_path), make_repo_relative(request_path)


def build_case_result(
    case: dict[str, Any],
    *,
    output_root: Path | None,
    record_generated_paths: bool,
    force_strict_validator: bool,
) -> dict[str, Any]:
    export_path = resolve_relative_to_repo(str(case["export_snapshot"]))
    label = str(case.get("label") or export_path.stem).strip()
    result: dict[str, Any] = {
        "label": label,
        "export_snapshot": make_repo_relative(export_path),
        "validator_status": "not_checked",
        "adapter_snapshot_status": "not_checked",
        "request_status": "not_checked",
        "passed": False,
    }

    expected_adapter_value = str(case.get("expected_adapter_snapshot") or "").strip()
    expected_request_value = str(case.get("expected_request_sample") or "").strip()
    if expected_adapter_value:
        result["expected_adapter_snapshot"] = expected_adapter_value
    if expected_request_value:
        result["expected_request_sample"] = expected_request_value

    strict = bool(case.get("strict_validator")) or force_strict_validator
    result["strict_validator"] = strict

    validator_result = run_export_validator(export_path, strict=strict)
    result["validator_status"] = validator_status(validator_result)
    if validator_result.returncode != 0:
        result["failure_reason"] = (validator_result.stdout + validator_result.stderr).strip()
        return result

    export_snapshot = load_json_object(export_path, f"{label} export snapshot")
    ensure_schema(export_snapshot, EXPORT_SCHEMA_PATH, f"{label} export snapshot")
    adapter_snapshot = build_adapter_snapshot_from_export_snapshot(export_snapshot)
    ensure_schema(adapter_snapshot, ADAPTER_SCHEMA_PATH, f"{label} generated adapter snapshot")
    request = build_request_from_export_snapshot(export_snapshot)
    ensure_schema(request, REQUEST_SCHEMA_PATH, f"{label} generated request")

    if output_root is not None and record_generated_paths:
        generated_adapter_path, generated_request_path = write_generated_documents(output_root, label, adapter_snapshot, request)
        result["generated_adapter_snapshot"] = generated_adapter_path
        result["generated_request"] = generated_request_path

    if expected_adapter_value:
        expected_adapter_snapshot = load_json_object(
            resolve_relative_to_repo(expected_adapter_value),
            f"{label} expected adapter snapshot",
        )
        ensure_schema(expected_adapter_snapshot, ADAPTER_SCHEMA_PATH, f"{label} expected adapter snapshot")
        if adapter_snapshot != expected_adapter_snapshot:
            result["adapter_snapshot_status"] = "failed"
            result["failure_reason"] = "generated adapter snapshot does not match expected adapter snapshot"
            return result
        result["adapter_snapshot_status"] = "passed"

    if expected_request_value:
        expected_request_sample = load_json_object(
            resolve_relative_to_repo(expected_request_value),
            f"{label} expected request sample",
        )
        expected_request = expected_request_sample.get("input_request")
        if not isinstance(expected_request, dict):
            raise SystemExit(f"{label} expected request sample: missing input_request object")
        ensure_schema(expected_request, REQUEST_SCHEMA_PATH, f"{label} expected request")
        if request != expected_request:
            result["request_status"] = "failed"
            result["failure_reason"] = "generated request does not match expected request sample"
            return result
        result["request_status"] = "passed"

    result["passed"] = True
    return result


def build_summary_document(manifest_path: Path, output_root: Path | None, results: list[dict[str, Any]]) -> dict[str, Any]:
    document: dict[str, Any] = {
        "schema_version": 1,
        "pipeline": "radishflow-export-smoke",
        "project": "radishflow",
        "manifest_path": make_repo_relative(manifest_path),
        "summary": {
            "case_count": len(results),
            "passed_case_count": sum(1 for item in results if item["passed"] is True),
            "warning_case_count": sum(1 for item in results if item["validator_status"] == "passed_with_warnings"),
            "failed_case_count": sum(1 for item in results if item["passed"] is not True),
            "checked_adapter_snapshot_count": sum(1 for item in results if item["adapter_snapshot_status"] != "not_checked"),
            "checked_request_sample_count": sum(1 for item in results if item["request_status"] != "not_checked"),
        },
        "results": results,
    }
    if output_root is not None:
        document["output_root"] = make_repo_relative(output_root)
    ensure_schema(document, SUMMARY_SCHEMA_PATH, "radishflow export smoke summary")
    return document


def write_or_check_summary(summary: dict[str, Any], *, summary_output: str, check_summary: str) -> None:
    if check_summary.strip():
        check_path = resolve_relative_to_repo(check_summary)
        existing_summary = load_json_document(check_path)
        if existing_summary != summary:
            raise SystemExit(f"radishflow export smoke summary is stale: {make_repo_relative(check_path)}")
        print(f"radishflow export smoke summary is up to date: {make_repo_relative(check_path)}")
        return

    if summary_output.strip():
        output_path = resolve_relative_to_repo(summary_output)
        write_json_document(output_path, summary)
        print(f"wrote radishflow export smoke summary to {make_repo_relative(output_path)}")
        return

    print(json.dumps(summary, ensure_ascii=False, indent=2))


def main() -> int:
    args = parse_args()
    manifest_path = resolve_relative_to_repo(args.manifest)
    manifest = load_json_object(manifest_path, "radishflow export smoke manifest")
    ensure_schema(manifest, MANIFEST_SCHEMA_PATH, make_repo_relative(manifest_path))

    cases = manifest.get("cases")
    if not isinstance(cases, list):
        raise SystemExit("radishflow export smoke manifest: cases must be a list")

    output_root: Path | None = None
    temp_output_dir: tempfile.TemporaryDirectory[str] | None = None
    record_generated_paths = False
    if args.output_root.strip():
        output_root = resolve_relative_to_repo(args.output_root)
        record_generated_paths = True
    elif not args.check_summary.strip():
        temp_output_dir = tempfile.TemporaryDirectory(prefix="radishflow-export-smoke-")
        output_root = Path(temp_output_dir.name)

    results = [
        build_case_result(
            case if isinstance(case, dict) else {},
            output_root=output_root,
            record_generated_paths=record_generated_paths,
            force_strict_validator=args.strict_validator,
        )
        for case in cases
    ]
    summary = build_summary_document(
        manifest_path,
        None if temp_output_dir is not None else output_root,
        results,
    )
    write_or_check_summary(summary, summary_output=args.summary_output, check_summary=args.check_summary)

    failed = [item["label"] for item in results if item["passed"] is not True]
    if failed:
        raise SystemExit("radishflow export smoke found failed case(s): " + ", ".join(failed))

    print("radishflow export smoke passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
