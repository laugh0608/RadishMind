#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from collections import Counter
from pathlib import Path
from typing import Any

import jsonschema

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from scripts.checks.copilot_training_sample import (  # noqa: E402
    TRAIN_FIELDS,
    build_training_sample_from_eval,
    check_training_sample,
    iter_proposed_actions,
)

TRAINING_SAMPLE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-training-sample.schema.json"
COPILOT_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-request.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def write_json_document(path: Path, document: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def normalize_repo_relative_path(path_value: str) -> Path:
    path = Path(path_value)
    if path.is_absolute():
        try:
            return path.relative_to(REPO_ROOT)
        except ValueError as exc:
            raise SystemExit(f"path is outside repository: {path_value}") from exc
    return path


def resolve_output_path(path_value: str) -> Path:
    path = Path(path_value)
    if path.is_absolute():
        return path
    return REPO_ROOT / path


def iter_manifest_sample_entries(manifest: dict[str, Any]) -> list[dict[str, Any]]:
    if manifest.get("schema_version") != 1:
        raise SystemExit("training sample conversion manifest schema_version must be 1")
    if manifest.get("kind") != "copilot_training_sample_conversion_manifest":
        raise SystemExit("training sample conversion manifest kind mismatch")
    selection = manifest.get("selection")
    if not isinstance(selection, list) or not selection:
        raise SystemExit("training sample conversion manifest must include non-empty selection")

    entries: list[dict[str, Any]] = []
    for group in selection:
        if not isinstance(group, dict):
            raise SystemExit("training sample conversion selection group must be an object")
        project = str(group.get("project") or "").strip()
        task = str(group.get("task") or "").strip()
        samples = group.get("samples")
        if not project or not task:
            raise SystemExit("training sample conversion selection group is missing project/task")
        if not isinstance(samples, list) or not samples:
            raise SystemExit(f"{project}/{task} selection group must include samples")
        for sample in samples:
            if not isinstance(sample, dict):
                raise SystemExit(f"{project}/{task} sample entry must be an object")
            path_value = str(sample.get("path") or "").strip()
            if not path_value:
                raise SystemExit(f"{project}/{task} sample entry is missing path")
            entries.append(
                {
                    "project": project,
                    "task": task,
                    "path": normalize_repo_relative_path(path_value),
                }
            )
    return entries


def validate_and_build_samples(
    manifest: dict[str, Any],
    *,
    training_sample_schema: dict[str, Any],
    copilot_request_schema: dict[str, Any],
    copilot_response_schema: dict[str, Any],
) -> tuple[list[dict[str, Any]], list[Path]]:
    created_for = str(manifest.get("created_for") or "teacher-student-distillation").strip()
    samples: list[dict[str, Any]] = []
    source_paths: list[Path] = []
    seen_generated_ids: set[str] = set()

    for entry in iter_manifest_sample_entries(manifest):
        relative_path = entry["path"]
        eval_sample_path = REPO_ROOT / relative_path
        eval_sample = load_json_document(eval_sample_path)
        if not isinstance(eval_sample, dict):
            raise SystemExit(f"{relative_path.as_posix()} must be a JSON object")
        if eval_sample.get("project") != entry["project"] or eval_sample.get("task") != entry["task"]:
            raise SystemExit(f"{relative_path.as_posix()} project/task does not match manifest")

        sample = build_training_sample_from_eval(
            eval_sample,
            source_eval_sample=relative_path,
            source_eval_sample_label=relative_path.as_posix(),
            created_for=created_for,
        )
        sample_id = str(sample.get("sample_id") or "").strip()
        if sample_id in seen_generated_ids:
            raise SystemExit(f"duplicate generated training sample id: {sample_id}")
        seen_generated_ids.add(sample_id)

        jsonschema.validate(sample, training_sample_schema)
        jsonschema.validate(sample["input_request"], copilot_request_schema)
        jsonschema.validate(sample["target_response"], copilot_response_schema)
        try:
            check_training_sample(sample)
        except ValueError as exc:
            raise SystemExit(f"{relative_path.as_posix()}: {exc}") from exc
        samples.append(sample)
        source_paths.append(relative_path)

    return samples, source_paths


def build_summary(
    samples: list[dict[str, Any]],
    source_paths: list[Path],
    *,
    manifest: dict[str, Any],
) -> dict[str, Any]:
    task_counts = Counter(f"{sample['project']}/{sample['task']}" for sample in samples)
    quality_gate_counts: Counter[str] = Counter()
    for sample in samples:
        gates = sample.get("quality_gates") if isinstance(sample.get("quality_gates"), dict) else {}
        for key in ("schema_validated", "risk_reviewed", "citation_checked"):
            if gates.get(key) is True:
                quality_gate_counts[key] += 1

    proposed_actions = [
        action
        for sample in samples
        for action in iter_proposed_actions(sample["target_response"])
    ]
    return {
        "schema_version": 1,
        "kind": "copilot_training_sample_conversion_summary",
        "source": "golden_response",
        "created_for": str(manifest.get("created_for") or "teacher-student-distillation").strip(),
        "does_not_run_models": True,
        "sample_count": len(samples),
        "task_counts": dict(sorted(task_counts.items())),
        "source_eval_samples": [path.as_posix() for path in source_paths],
        "generated_sample_ids": [str(sample["sample_id"]) for sample in samples],
        "train_fields": list(TRAIN_FIELDS),
        "quality_gate_counts": dict(sorted(quality_gate_counts.items())),
        "requires_confirmation_count": sum(
            1 for sample in samples if sample["target_response"].get("requires_confirmation") is True
        ),
        "action_sample_count": sum(
            1 for sample in samples if iter_proposed_actions(sample["target_response"])
        ),
        "medium_high_action_count": sum(
            1 for action in proposed_actions if action.get("risk_level") in {"medium", "high"}
        ),
    }


def write_jsonl(path: Path, samples: list[dict[str, Any]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    lines = [json.dumps(sample, ensure_ascii=False, separators=(",", ":")) for sample in samples]
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", required=True)
    parser.add_argument("--output-jsonl")
    parser.add_argument("--summary-output")
    parser.add_argument("--check-summary")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    manifest_path = REPO_ROOT / normalize_repo_relative_path(args.manifest)
    manifest = load_json_document(manifest_path)
    if not isinstance(manifest, dict):
        raise SystemExit("training sample conversion manifest must be a JSON object")

    training_sample_schema = load_json_document(TRAINING_SAMPLE_SCHEMA_PATH)
    copilot_request_schema = load_json_document(COPILOT_REQUEST_SCHEMA_PATH)
    copilot_response_schema = load_json_document(COPILOT_RESPONSE_SCHEMA_PATH)
    jsonschema.Draft202012Validator.check_schema(training_sample_schema)
    samples, source_paths = validate_and_build_samples(
        manifest,
        training_sample_schema=training_sample_schema,
        copilot_request_schema=copilot_request_schema,
        copilot_response_schema=copilot_response_schema,
    )
    summary = build_summary(samples, source_paths, manifest=manifest)

    if args.output_jsonl:
        write_jsonl(resolve_output_path(args.output_jsonl), samples)
    if args.summary_output:
        write_json_document(resolve_output_path(args.summary_output), summary)
    if args.check_summary:
        expected_summary = load_json_document(REPO_ROOT / normalize_repo_relative_path(args.check_summary))
        if expected_summary != summary:
            raise SystemExit(f"{args.check_summary} does not match regenerated training sample conversion summary")

    print(f"copilot training sample conversion passed: {len(samples)} samples.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
