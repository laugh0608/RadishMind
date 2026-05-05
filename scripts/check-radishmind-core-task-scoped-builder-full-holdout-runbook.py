#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
RUNBOOK_PATH = REPO_ROOT / "training/experiments/radishmind-core-task-scoped-builder-full-holdout-runbook-v0.json"

EXPECTED_SAMPLES = {
    "radishflow-suggest-flowsheet-edits-action-citation-ordering-diagnostic-artifact-snapshot-001",
    "radishflow-suggest-flowsheet-edits-compressor-parameter-update-ordering-001",
    "radishflow-suggest-flowsheet-edits-valve-local-fix-vs-global-balance-001",
    "radishflow-suggest-ghost-completion-flash-inlet-001",
    "radishflow-suggest-ghost-completion-heater-stream-name-001",
    "radishflow-suggest-ghost-completion-mixer-standard-outlet-001",
    "radish-answer-docs-question-attachment-mixed-001",
    "radish-answer-docs-question-docs-attachments-faq-001",
    "radish-answer-docs-question-navigation-001",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse {path.relative_to(REPO_ROOT)}: {exc}") from exc


def require_existing_file(path_value: str, *, field_name: str) -> None:
    require(path_value.strip() == path_value and path_value.strip(), f"{field_name} must be a non-empty clean path")
    require((REPO_ROOT / path_value).is_file(), f"{field_name} points to a missing file: {path_value}")


def require_tmp_path(path_value: str, *, field_name: str) -> None:
    require(path_value.startswith("tmp/"), f"{field_name} must stay under tmp/: {path_value}")


def require_args(args: list[Any], *, field_name: str, required_flags: tuple[str, ...]) -> None:
    require(all(isinstance(item, str) and item for item in args), f"{field_name} command_args must be non-empty strings")
    for flag in required_flags:
        require(flag in args, f"{field_name} command_args missing {flag}")
    require("--repair-hard-fields" not in args, f"{field_name} must not use --repair-hard-fields")
    require("--inject-hard-fields" not in args, f"{field_name} must not use --inject-hard-fields")
    require("--build-suggest-edits-response" not in args, f"{field_name} must not use --build-suggest-edits-response")


def arg_value(args: list[str], flag: str) -> str:
    require(flag in args, f"missing flag: {flag}")
    index = args.index(flag)
    require(index + 1 < len(args), f"{flag} is missing a value")
    return args[index + 1]


def main() -> int:
    document = load_json(RUNBOOK_PATH)
    require(isinstance(document, dict), "runbook must be a JSON object")
    require(document.get("schema_version") == 1, "runbook schema_version must be 1")
    require(
        document.get("kind") == "radishmind_core_task_scoped_builder_full_holdout_runbook",
        "runbook kind mismatch",
    )
    require(
        document.get("runbook_id") == "radishmind-core-task-scoped-builder-full-holdout-runbook-v0",
        "runbook_id mismatch",
    )
    require(document.get("status") == "planned", "runbook status must remain planned")
    require(document.get("does_not_run_models") is True, "committed runbook must not run models")
    require(document.get("does_not_generate_jsonl") is True, "runbook must not generate JSONL")
    require(document.get("does_not_mark_review_pass") is True, "runbook must not mark review pass")
    require(document.get("candidate_track") == "--build-task-scoped-response", "candidate_track mismatch")
    require(document.get("sample_set_id") == "full-holdout-9", "sample_set_id mismatch")
    require_existing_file(str(document.get("review_plan") or ""), field_name="review_plan")

    model = document.get("model")
    require(isinstance(model, dict), "model must be an object")
    require(model.get("provider") == "local_transformers", "provider must be local_transformers")
    require(str(model.get("model_dir") or "").startswith("/"), "model_dir must be an explicit local path")
    require(model.get("sample_timeout_seconds") == 300, "sample_timeout_seconds must be 300")

    input_manifests = document.get("input_manifests")
    require(isinstance(input_manifests, dict), "input_manifests must be an object")
    require_existing_file(str(input_manifests.get("candidate_manifest") or ""), field_name="candidate_manifest")
    require_existing_file(str(input_manifests.get("candidate_eval_manifest") or ""), field_name="candidate_eval_manifest")
    require_existing_file(str(input_manifests.get("review_plan") or ""), field_name="input review_plan")

    artifact_paths = document.get("artifact_paths")
    require(isinstance(artifact_paths, dict), "artifact_paths must be an object")
    for key in ("candidate_output_dir", "candidate_summary", "offline_eval_run", "natural_language_audit"):
        require_tmp_path(str(artifact_paths.get(key) or ""), field_name=f"artifact_paths.{key}")

    batch = document.get("official_batch_run")
    require(isinstance(batch, dict), "official_batch_run must be an object")
    batch_args = batch.get("command_args")
    require(isinstance(batch_args, list), "official_batch_run.command_args must be a list")
    require_args(
        batch_args,
        field_name="official_batch_run",
        required_flags=(
            "--manifest",
            "--provider",
            "--model-dir",
            "--output-dir",
            "--summary-output",
            "--validate-task",
            "--build-task-scoped-response",
            "--sample-timeout-seconds",
        ),
    )
    require(arg_value(batch_args, "--provider") == "local_transformers", "official batch provider mismatch")
    require(arg_value(batch_args, "--manifest") == input_manifests["candidate_manifest"], "official batch manifest mismatch")
    require_tmp_path(arg_value(batch_args, "--output-dir"), field_name="official_batch_run.output_dir")
    require_tmp_path(arg_value(batch_args, "--summary-output"), field_name="official_batch_run.summary_output")
    require(arg_value(batch_args, "--sample-timeout-seconds") == "300", "official batch timeout mismatch")

    offline_eval = document.get("offline_eval")
    require(isinstance(offline_eval, dict), "offline_eval must be an object")
    offline_args = offline_eval.get("command_args")
    require(isinstance(offline_args, list), "offline_eval.command_args must be a list")
    require_args(
        offline_args,
        field_name="offline_eval",
        required_flags=("--manifest", "--candidate-summary", "--candidate-output-dir", "--output", "--check-output"),
    )
    require(arg_value(offline_args, "--manifest") == input_manifests["candidate_eval_manifest"], "offline eval manifest mismatch")
    require_tmp_path(arg_value(offline_args, "--output"), field_name="offline_eval.output")

    audit = document.get("natural_language_audit")
    require(isinstance(audit, dict), "natural_language_audit must be an object")
    audit_args = audit.get("command_args")
    require(isinstance(audit_args, list), "natural_language_audit.command_args must be a list")
    require_args(
        audit_args,
        field_name="natural_language_audit",
        required_flags=("--manifest", "--candidate-summary", "--candidate-output-dir", "--output"),
    )
    require(arg_value(audit_args, "--manifest") == input_manifests["candidate_eval_manifest"], "audit manifest mismatch")
    require_tmp_path(arg_value(audit_args, "--output"), field_name="natural_language_audit.output")

    locator_runs = document.get("single_sample_locator_runs")
    require(isinstance(locator_runs, list) and len(locator_runs) == 9, "single_sample_locator_runs must include 9 samples")
    seen_samples: set[str] = set()
    for entry in locator_runs:
        require(isinstance(entry, dict), "single-sample locator entry must be an object")
        sample_id = str(entry.get("sample_id") or "")
        seen_samples.add(sample_id)
        args = entry.get("command_args")
        require(isinstance(args, list), f"{sample_id} command_args must be a list")
        require_args(
            args,
            field_name=f"single_sample_locator_runs.{sample_id}",
            required_flags=(
                "--manifest",
                "--provider",
                "--model-dir",
                "--sample-id",
                "--output-dir",
                "--summary-output",
                "--validate-task",
                "--build-task-scoped-response",
                "--sample-timeout-seconds",
            ),
        )
        require(arg_value(args, "--sample-id") == sample_id, f"{sample_id} --sample-id mismatch")
        require(arg_value(args, "--manifest") == input_manifests["candidate_manifest"], f"{sample_id} manifest mismatch")
        require(arg_value(args, "--provider") == "local_transformers", f"{sample_id} provider mismatch")
        require_tmp_path(arg_value(args, "--output-dir"), field_name=f"{sample_id}.output_dir")
        require_tmp_path(arg_value(args, "--summary-output"), field_name=f"{sample_id}.summary_output")
        require(arg_value(args, "--sample-timeout-seconds") == "300", f"{sample_id} timeout mismatch")
    require(seen_samples == EXPECTED_SAMPLES, "single_sample_locator_runs sample set mismatch")

    expectations = document.get("acceptance_expectations")
    require(isinstance(expectations, dict), "acceptance_expectations must be an object")
    candidate_summary = expectations.get("candidate_summary")
    require(isinstance(candidate_summary, dict), "candidate_summary expectations must be an object")
    require(candidate_summary.get("sample_count") == 9, "candidate summary expected sample_count mismatch")
    require(candidate_summary.get("schema_valid_rate") == 1.0, "candidate summary schema_valid_rate expectation mismatch")
    require(candidate_summary.get("task_valid_rate") == 1.0, "candidate summary task_valid_rate expectation mismatch")
    require(candidate_summary.get("timeout_count") == 0, "candidate summary timeout_count expectation mismatch")
    require(candidate_summary.get("builder_output_count") == 9, "candidate summary builder_output_count mismatch")

    artifact_policy = document.get("artifact_policy")
    require(isinstance(artifact_policy, dict), "artifact_policy must be an object")
    require(artifact_policy.get("generated_outputs_default_location") == "tmp/", "generated outputs must default to tmp/")
    disallowed = artifact_policy.get("committed_disallowed")
    require(isinstance(disallowed, list), "committed_disallowed must be a list")
    for blocked in ("candidate responses", "provider raw dump", "large jsonl", "model weights", "checkpoint"):
        require(blocked in disallowed, f"artifact policy must disallow {blocked}")

    print("radishmind core task-scoped builder full-holdout runbook check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
