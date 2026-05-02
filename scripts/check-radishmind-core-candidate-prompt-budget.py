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
DEFAULT_MANIFEST = "scripts/checks/fixtures/radishmind-core-holdout-probe-v2-candidate-manifest.json"
DEFAULT_EXPECTED = "scripts/checks/fixtures/radishmind-core-holdout-probe-v2-prompt-budget-summary.json"
LARGEST_EXPECTED_SAMPLE_ID = "radishflow-suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-plus-pump-parameter-001"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", default=DEFAULT_MANIFEST)
    parser.add_argument("--expected-summary", default=DEFAULT_EXPECTED)
    parser.add_argument("--output")
    return parser.parse_args()


def repo_path(path_value: str) -> Path:
    path = Path(path_value)
    return path if path.is_absolute() else REPO_ROOT / path


def repo_rel(path: Path) -> str:
    try:
        return path.relative_to(REPO_ROOT).as_posix()
    except ValueError:
        return path.as_posix()


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{repo_rel(path)}': {exc}") from exc


def write_json(path: Path, document: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def run_candidate_manifest(manifest: str, output_dir: Path, summary_path: Path) -> None:
    result = subprocess.run(
        [
            sys.executable,
            str(REPO_ROOT / "scripts/run-radishmind-core-candidate.py"),
            "--manifest",
            manifest,
            "--output-dir",
            str(output_dir),
            "--summary-output",
            str(summary_path),
        ],
        cwd=REPO_ROOT,
    )
    if result.returncode != 0:
        raise SystemExit(result.returncode)


def load_prompt_budgets(output_dir: Path) -> list[dict[str, Any]]:
    prompt_dir = output_dir / "prompts"
    require(prompt_dir.is_dir(), f"prompt output directory is missing: {repo_rel(prompt_dir)}")
    rows: list[dict[str, Any]] = []
    for prompt_path in sorted(prompt_dir.glob("*.prompt.json")):
        prompt = load_json(prompt_path)
        require(isinstance(prompt, dict), f"{repo_rel(prompt_path)} must be a JSON object")
        budget = prompt.get("prompt_budget")
        require(isinstance(budget, dict), f"{repo_rel(prompt_path)} is missing prompt_budget")
        required_budget_keys = {
            "total_message_chars",
            "request_json_chars",
            "output_contract_chars",
            "scaffold_json_chars",
            "hard_field_freeze_json_chars",
            "hard_field_freeze_field_count",
            "scaffold_counts",
            "scaffold_payload_chars",
        }
        missing_keys = sorted(required_budget_keys - set(budget))
        require(not missing_keys, f"{repo_rel(prompt_path)} prompt_budget is missing keys: {missing_keys}")
        rows.append(
            {
                "sample_id": prompt.get("sample_id"),
                "project": prompt.get("project"),
                "task": prompt.get("task"),
                "total_message_chars": budget["total_message_chars"],
                "request_json_chars": budget["request_json_chars"],
                "output_contract_chars": budget["output_contract_chars"],
                "scaffold_json_chars": budget["scaffold_json_chars"],
                "hard_field_freeze_json_chars": budget["hard_field_freeze_json_chars"],
                "hard_field_freeze_field_count": budget["hard_field_freeze_field_count"],
                "scaffold_counts": budget["scaffold_counts"],
                "scaffold_payload_chars": budget["scaffold_payload_chars"],
            }
        )
    require(rows, "candidate prompt budget check found no prompts")
    return rows


def max_row(rows: list[dict[str, Any]], key: str) -> dict[str, Any]:
    return max(rows, key=lambda row: int(row[key]))


def build_summary(manifest: str, rows: list[dict[str, Any]]) -> dict[str, Any]:
    largest = max_row(rows, "total_message_chars")
    freeze_largest = max_row(rows, "hard_field_freeze_field_count")
    require(largest["sample_id"] == LARGEST_EXPECTED_SAMPLE_ID, "v2 largest prompt sample changed; review prompt budget")
    require(int(largest["total_message_chars"]) <= 18000, "v2 largest prompt exceeds current static char budget")
    require(int(largest["hard_field_freeze_field_count"]) <= 20, "v2 largest prompt exceeds freeze field budget")
    require(int(largest["request_json_chars"]) <= 4000, "v2 largest prompt request payload exceeds budget")
    require(int(largest["output_contract_chars"]) <= 12500, "v2 largest prompt output contract exceeds budget")

    return {
        "schema_version": 1,
        "kind": "radishmind_core_candidate_prompt_budget_summary",
        "phase": "M4-preparation",
        "source_candidate_manifest": manifest,
        "sample_count": len(rows),
        "budget_policy": {
            "measurement": "static_prompt_character_budget",
            "max_total_message_chars": 18000,
            "max_request_json_chars": 4000,
            "max_output_contract_chars": 12500,
            "max_hard_field_freeze_field_count": 20,
            "largest_prompt_sample_must_remain_reviewed": LARGEST_EXPECTED_SAMPLE_ID,
            "does_not_run_models": True,
        },
        "max_observed": {
            "total_message_chars": largest["total_message_chars"],
            "request_json_chars": max_row(rows, "request_json_chars")["request_json_chars"],
            "output_contract_chars": max_row(rows, "output_contract_chars")["output_contract_chars"],
            "scaffold_json_chars": max_row(rows, "scaffold_json_chars")["scaffold_json_chars"],
            "hard_field_freeze_json_chars": max_row(rows, "hard_field_freeze_json_chars")["hard_field_freeze_json_chars"],
            "hard_field_freeze_field_count": freeze_largest["hard_field_freeze_field_count"],
        },
        "largest_prompt_sample": {
            key: largest[key]
            for key in (
                "sample_id",
                "project",
                "task",
                "total_message_chars",
                "request_json_chars",
                "output_contract_chars",
                "scaffold_json_chars",
                "hard_field_freeze_json_chars",
                "hard_field_freeze_field_count",
                "scaffold_counts",
            )
        },
        "samples": sorted(rows, key=lambda row: str(row["sample_id"])),
    }


def main() -> int:
    args = parse_args()
    with tempfile.TemporaryDirectory(prefix="check-core-prompt-budget-") as temp_dir:
        output_dir = Path(temp_dir) / "candidate-run"
        summary_path = Path(temp_dir) / "summary.json"
        run_candidate_manifest(args.manifest, output_dir, summary_path)
        rows = load_prompt_budgets(output_dir)
        summary = build_summary(args.manifest, rows)

    if args.output:
        write_json(repo_path(args.output), summary)
    expected_path = repo_path(args.expected_summary)
    if expected_path.is_file():
        expected = load_json(expected_path)
        if expected != summary:
            raise SystemExit(f"prompt budget summary does not match {repo_rel(expected_path)}")
    print("radishmind core candidate prompt budget check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
