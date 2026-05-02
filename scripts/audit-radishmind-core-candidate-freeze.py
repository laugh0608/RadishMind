#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--summary", required=True, help="Path to a radishmind core candidate run summary JSON.")
    parser.add_argument(
        "--output-dir",
        help="Candidate run output directory. Defaults to the summary parent directory.",
    )
    parser.add_argument("--audit-output", help="Optional path to write the freeze audit JSON summary.")
    return parser.parse_args()


def load_json(path: Path) -> Any:
    return json.loads(path.read_text(encoding="utf-8"))


def write_json(path: Path, document: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def resolve_path(path_value: str) -> Path:
    path = Path(path_value)
    return path if path.is_absolute() else REPO_ROOT / path


def path_segments(path: str) -> list[str | int]:
    if not path.startswith("$."):
        raise ValueError(f"unsupported JSON path: {path}")
    segments: list[str | int] = []
    current = path[2:]
    for part in current.split("."):
        while "[" in part:
            key, rest = part.split("[", 1)
            if key:
                segments.append(key)
            index_text, trailing = rest.split("]", 1)
            if not index_text.isdigit():
                raise ValueError(f"unsupported JSON array index in path: {path}")
            segments.append(int(index_text))
            part = trailing
        if part:
            segments.append(part)
    return segments


def get_path_value(document: Any, path: str) -> tuple[bool, Any]:
    current = document
    for segment in path_segments(path):
        if isinstance(segment, int):
            if not isinstance(current, list) or segment >= len(current):
                return False, None
            current = current[segment]
            continue
        if not isinstance(current, dict) or segment not in current:
            return False, None
        current = current[segment]
    return True, current


def audit_output(*, output_dir: Path, output_entry: dict[str, Any]) -> dict[str, Any]:
    prompt_path = output_dir / str(output_entry["prompt_file"])
    response_path = output_dir / str(output_entry["candidate_response_file"])
    prompt = load_json(prompt_path)
    response = load_json(response_path)
    freeze = prompt.get("hard_field_freeze") if isinstance(prompt, dict) else None
    fields = freeze.get("fields") if isinstance(freeze, dict) else None
    if not isinstance(fields, list):
        return {
            "sample_id": output_entry.get("sample_id"),
            "status": "failed",
            "violations": [
                {
                    "path": "$.hard_field_freeze.fields",
                    "reason": "missing_prompt_freeze_fields",
                }
            ],
        }

    violations: list[dict[str, Any]] = []
    for field in fields:
        if not isinstance(field, dict):
            violations.append({"path": "$.hard_field_freeze.fields[]", "reason": "invalid_freeze_field"})
            continue
        freeze_path = field.get("path")
        expected = field.get("value")
        if not isinstance(freeze_path, str):
            violations.append({"path": "$.hard_field_freeze.fields[].path", "reason": "invalid_freeze_path"})
            continue
        exists, actual = get_path_value(response, freeze_path)
        if not exists:
            violations.append(
                {
                    "path": freeze_path,
                    "reason": "missing_response_path",
                    "expected": expected,
                }
            )
            continue
        if actual != expected:
            violations.append(
                {
                    "path": freeze_path,
                    "reason": "value_mismatch",
                    "expected": expected,
                    "actual": actual,
                }
            )

    return {
        "sample_id": output_entry.get("sample_id"),
        "project": output_entry.get("project"),
        "task": output_entry.get("task"),
        "status": "passed" if not violations else "failed",
        "freeze_field_count": len(fields),
        "violation_count": len(violations),
        "violations": violations,
    }


def build_audit(args: argparse.Namespace) -> dict[str, Any]:
    summary_path = resolve_path(str(args.summary))
    output_dir = resolve_path(str(args.output_dir)) if args.output_dir else summary_path.parent
    summary = load_json(summary_path)
    outputs = summary.get("outputs") if isinstance(summary, dict) else None
    if not isinstance(outputs, list):
        raise SystemExit("candidate summary must contain outputs[]")

    audited_outputs: list[dict[str, Any]] = []
    skipped_outputs: list[dict[str, Any]] = []
    for output_entry in outputs:
        if not isinstance(output_entry, dict):
            continue
        if output_entry.get("copilot_response_schema_validated") is not True:
            skipped_outputs.append(
                {
                    "sample_id": output_entry.get("sample_id"),
                    "reason": "candidate_response_not_schema_valid",
                }
            )
            continue
        audited_outputs.append(audit_output(output_dir=output_dir, output_entry=output_entry))

    failed_outputs = [entry for entry in audited_outputs if entry.get("status") != "passed"]
    violation_counts_by_path: dict[str, int] = {}
    for entry in failed_outputs:
        for violation in entry.get("violations") or []:
            path = str(violation.get("path") or "$")
            violation_counts_by_path[path] = violation_counts_by_path.get(path, 0) + 1

    return {
        "schema_version": 1,
        "kind": "radishmind_core_candidate_freeze_audit",
        "source_summary": str(summary_path.relative_to(REPO_ROOT)) if summary_path.is_relative_to(REPO_ROOT) else str(summary_path),
        "output_dir": str(output_dir.relative_to(REPO_ROOT)) if output_dir.is_relative_to(REPO_ROOT) else str(output_dir),
        "audited_count": len(audited_outputs),
        "skipped_count": len(skipped_outputs),
        "passed_count": len(audited_outputs) - len(failed_outputs),
        "failed_count": len(failed_outputs),
        "violation_counts_by_path": dict(sorted(violation_counts_by_path.items())),
        "outputs": audited_outputs,
        "skipped_outputs": skipped_outputs,
    }


def main() -> int:
    args = parse_args()
    audit = build_audit(args)
    if args.audit_output:
        write_json(resolve_path(str(args.audit_output)), audit)
    print(
        "freeze audit: "
        f"audited={audit['audited_count']} "
        f"passed={audit['passed_count']} "
        f"failed={audit['failed_count']} "
        f"skipped={audit['skipped_count']}"
    )
    if audit["failed_count"]:
        for sample in audit["outputs"]:
            if sample.get("status") == "failed":
                print(f"FAIL {sample['sample_id']}: {sample['violation_count']} freeze violation(s)")
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
