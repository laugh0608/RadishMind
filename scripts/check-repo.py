#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Any

import jsonschema

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.real_derived_negative_index import load_real_derived_negative_index  # noqa: E402
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"


def load_fixture_json(name: str) -> Any:
    path = FIXTURE_DIR / name
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse fixture json '{path.relative_to(REPO_ROOT)}': {exc}") from exc


GHOST_REQUEST_ASSEMBLY_FIXTURES: list[dict[str, Any]] = []
for fixture_name in (
    "ghost-request-assembly-base.json",
    "ghost-request-assembly-chain-valve.json",
    "ghost-request-assembly-chain-heater.json",
    "ghost-request-assembly-chain-cooler.json",
):
    GHOST_REQUEST_ASSEMBLY_FIXTURES.extend(load_fixture_json(fixture_name))

RADISHFLOW_ASSEMBLY_FIXTURES = load_fixture_json("radishflow-assembly-fixtures.json")
RADISHFLOW_ADAPTER_REQUEST_ASSEMBLY_FIXTURES = list(
    RADISHFLOW_ASSEMBLY_FIXTURES.get("adapter_request_assembly") or []
)
RADISHFLOW_EXPORT_SNAPSHOT_ASSEMBLY_FIXTURES = list(
    RADISHFLOW_ASSEMBLY_FIXTURES.get("export_snapshot_assembly") or []
)
RADISH_DOCS_QA_REAL_BATCHES = list(load_fixture_json("radish-docs-real-batches.json") or [])
RADISH_DOCS_QA_REAL_DERIVED_NEGATIVES = dict(load_fixture_json("radish-docs-real-derived-negatives.json") or {})
RADISHFLOW_GHOST_REAL_BATCHES = list(load_fixture_json("radishflow-ghost-real-batches.json") or [])
REQUIRED_FILES = list(load_fixture_json("required-files.json") or [])

def run_python_script(script_name: str, args: list[str]) -> None:
    result = subprocess.run([sys.executable, str(REPO_ROOT / "scripts" / script_name), *args], cwd=REPO_ROOT)
    if result.returncode != 0:
        raise SystemExit(result.returncode)


def load_json_file(relative_path: str) -> object:
    return json.loads((REPO_ROOT / relative_path).read_text(encoding="utf-8"))


def run_python_script_capture(script_name: str, args: list[str]) -> None:
    result = subprocess.run(
        [sys.executable, str(REPO_ROOT / "scripts" / script_name), *args],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        if result.stdout:
            print(result.stdout, end="")
        if result.stderr:
            print(result.stderr, end="", file=sys.stderr)
        raise SystemExit(result.returncode)


def assert_json_equal(expected: object, actual: object, *, label: str) -> None:
    if expected != actual:
        raise SystemExit(f"{label} does not match regenerated output")


def check_committed_candidate_record_batches() -> None:
    for batch in RADISHFLOW_GHOST_REAL_BATCHES:
        manifest_document = load_json_file(batch["manifest"])
        description = str((manifest_document or {}).get("description") or "").strip() if isinstance(manifest_document, dict) else ""
        with tempfile.TemporaryDirectory(prefix="check-repo-ghost-batch-") as temp_dir:
            temp_dir_path = Path(temp_dir)
            temp_manifest_path = temp_dir_path / "manifest.json"
            run_python_script_capture(
                "build-candidate-record-batch.py",
                [
                    "--record-dir",
                    batch["record_dir"],
                    "--output",
                    str(temp_manifest_path),
                    *(
                        ["--description", description]
                        if description
                        else []
                    ),
                ],
            )
            regenerated_manifest = json.loads(temp_manifest_path.read_text(encoding="utf-8"))
            assert_json_equal(
                manifest_document,
                regenerated_manifest,
                label=batch["manifest"],
            )

            temp_audit_path = temp_dir_path / "audit.json"
            run_python_script_capture(
                "audit-candidate-record-batch.py",
                [
                    batch["task"],
                    "--manifest",
                    batch["manifest"],
                    "--report-output",
                    str(temp_audit_path),
                ],
            )
            committed_audit = load_json_file(batch["audit_report"])
            regenerated_audit = json.loads(temp_audit_path.read_text(encoding="utf-8"))
            assert_json_equal(
                committed_audit,
                regenerated_audit,
                label=batch["audit_report"],
            )


def check_required_files() -> None:
    for relative_path in REQUIRED_FILES:
        if not (REPO_ROOT / relative_path).is_file():
            raise SystemExit(f"missing required file: {relative_path}")


def check_content_baseline() -> None:
    agents_content = (REPO_ROOT / "AGENTS.md").read_text(encoding="utf-8")
    if "当前常态开发分支为 `dev`" not in agents_content:
        raise SystemExit("AGENTS.md does not mention dev as the default development branch")

    pr_template_content = (REPO_ROOT / ".github/PULL_REQUEST_TEMPLATE.md").read_text(encoding="utf-8")
    if "默认目标分支为 `dev`" not in pr_template_content:
        raise SystemExit("PULL_REQUEST_TEMPLATE.md does not mention dev as the default target branch")

    ruleset = json.loads((REPO_ROOT / ".github/rulesets/master-protection.json").read_text(encoding="utf-8"))
    include_refs = (((ruleset.get("conditions") or {}).get("ref_name") or {}).get("include") or [])
    if "refs/heads/master" not in include_refs:
        raise SystemExit("master-protection.json does not target refs/heads/master")

    required_check_rule = next((rule for rule in ruleset.get("rules", []) if rule.get("type") == "required_status_checks"), None)
    if required_check_rule is None:
        raise SystemExit("master-protection.json is missing required_status_checks")

    contexts = [
        item.get("context")
        for item in ((required_check_rule.get("parameters") or {}).get("required_status_checks") or [])
    ]
    if "Repo Hygiene" not in contexts:
        raise SystemExit("master-protection.json is missing Repo Hygiene required check")
    if "Planning Baseline" not in contexts:
        raise SystemExit("master-protection.json is missing Planning Baseline required check")

    pr_workflow = (REPO_ROOT / ".github/workflows/pr-check.yml").read_text(encoding="utf-8")
    for pattern in ("name: PR Checks", "- master", "name: Repo Hygiene", "name: Planning Baseline"):
        if pattern not in pr_workflow:
            raise SystemExit(f".github/workflows/pr-check.yml is missing expected content: {pattern}")

    release_workflow = (REPO_ROOT / ".github/workflows/release-check.yml").read_text(encoding="utf-8")
    for pattern in (
        "name: Release Checks",
        "v*-dev",
        "v*-test",
        "v*-release",
        "name: Release Repo Hygiene",
        "name: Release Planning Baseline",
    ):
        if pattern not in release_workflow:
            raise SystemExit(f".github/workflows/release-check.yml is missing expected content: {pattern}")


def check_contract_schemas() -> None:
    contract_schema_paths = [
        REPO_ROOT / "contracts/copilot-request.schema.json",
        REPO_ROOT / "contracts/copilot-response.schema.json",
        REPO_ROOT / "contracts/radishflow-ghost-candidate-set.schema.json",
        REPO_ROOT / "contracts/radishflow-adapter-snapshot.schema.json",
        REPO_ROOT / "contracts/radishflow-export-snapshot.schema.json",
    ]
    for schema_path in contract_schema_paths:
        document = json.loads(schema_path.read_text(encoding="utf-8"))
        jsonschema.Draft202012Validator.check_schema(document)

    ghost_candidate_schema = json.loads(
        (REPO_ROOT / "contracts/radishflow-ghost-candidate-set.schema.json").read_text(encoding="utf-8")
    )
    adapter_snapshot_schema = json.loads(
        (REPO_ROOT / "contracts/radishflow-adapter-snapshot.schema.json").read_text(encoding="utf-8")
    )
    export_snapshot_schema = json.loads(
        (REPO_ROOT / "contracts/radishflow-export-snapshot.schema.json").read_text(encoding="utf-8")
    )
    copilot_request_schema = json.loads((REPO_ROOT / "contracts/copilot-request.schema.json").read_text(encoding="utf-8"))
    for fixture in GHOST_REQUEST_ASSEMBLY_FIXTURES:
        ghost_candidate_example = json.loads((REPO_ROOT / fixture["candidate_set"]).read_text(encoding="utf-8"))
        jsonschema.validate(ghost_candidate_example, ghost_candidate_schema)

        for request_fixture in fixture["requests"]:
            ghost_request_example = json.loads((REPO_ROOT / request_fixture["output"]).read_text(encoding="utf-8"))
            jsonschema.validate(ghost_request_example, copilot_request_schema)

            run_python_script(
                "build-radishflow-ghost-request.py",
                [
                    "--input",
                    fixture["candidate_set"],
                    "--output",
                    request_fixture["output"],
                    "--artifact-uri",
                    "artifact://flowsheet/current",
                    "--request-id",
                    request_fixture["request_id"],
                    "--assembly-profile",
                    request_fixture["assembly_profile"],
                    "--check",
                ],
            )

    for fixture in RADISHFLOW_ADAPTER_REQUEST_ASSEMBLY_FIXTURES:
        adapter_snapshot_example = json.loads((REPO_ROOT / fixture["snapshot"]).read_text(encoding="utf-8"))
        jsonschema.validate(adapter_snapshot_example, adapter_snapshot_schema)

        eval_sample = json.loads((REPO_ROOT / fixture["sample"]).read_text(encoding="utf-8"))
        if not isinstance(eval_sample, dict) or not isinstance(eval_sample.get("input_request"), dict):
            raise SystemExit(f"{fixture['sample']} is missing input_request")
        jsonschema.validate(eval_sample["input_request"], copilot_request_schema)

        run_python_script(
            "build-radishflow-request.py",
            [
                "--input",
                fixture["snapshot"],
                "--check-sample",
                fixture["sample"],
            ],
        )

    for fixture in RADISHFLOW_EXPORT_SNAPSHOT_ASSEMBLY_FIXTURES:
        export_snapshot_example = json.loads((REPO_ROOT / fixture["export_snapshot"]).read_text(encoding="utf-8"))
        jsonschema.validate(export_snapshot_example, export_snapshot_schema)

        adapter_snapshot_example = json.loads((REPO_ROOT / fixture["adapter_snapshot"]).read_text(encoding="utf-8"))
        jsonschema.validate(adapter_snapshot_example, adapter_snapshot_schema)

        eval_sample = json.loads((REPO_ROOT / fixture["sample"]).read_text(encoding="utf-8"))
        if not isinstance(eval_sample, dict) or not isinstance(eval_sample.get("input_request"), dict):
            raise SystemExit(f"{fixture['sample']} is missing input_request")
        jsonschema.validate(eval_sample["input_request"], copilot_request_schema)

        run_python_script(
            "validate-radishflow-export-snapshot.py",
            [
                "--input",
                fixture["export_snapshot"],
            ],
        )

        run_python_script(
            "build-radishflow-adapter-snapshot.py",
            [
                "--input",
                fixture["export_snapshot"],
                "--check-snapshot",
                fixture["adapter_snapshot"],
            ],
        )

        run_python_script(
            "build-radishflow-export-request.py",
            [
                "--input",
                fixture["export_snapshot"],
                "--check-sample",
                fixture["sample"],
            ],
        )

    with tempfile.TemporaryDirectory(prefix="check-repo-export-template-") as temp_dir:
        temp_dir_path = Path(temp_dir)
        for task in ("explain_diagnostics", "suggest_flowsheet_edits", "explain_control_plane_state"):
            output_path = temp_dir_path / f"{task}.export.json"
            run_python_script(
                "init-radishflow-export-snapshot.py",
                [
                    "--task",
                    task,
                    "--output",
                    str(output_path),
                ],
            )
            generated_template = json.loads(output_path.read_text(encoding="utf-8"))
            jsonschema.validate(generated_template, export_snapshot_schema)
            run_python_script(
                "validate-radishflow-export-snapshot.py",
                [
                    "--input",
                    str(output_path),
                ],
            )

    run_python_script("check-radishflow-artifact-summary-consumption.py", [])
    run_python_script("check-radishflow-runtime-artifact-metadata-consumption.py", [])
    run_python_script("check-radishflow-runtime-artifact-metadata-response-coercion.py", [])
    run_python_script("check-artifact-metadata-contract.py", [])
    run_python_script("check-copilot-request-artifact-metadata.py", [])
    run_python_script("check-radishflow-export-selection-contract.py", [])
    run_python_script("check-radishflow-export-priority-contract.py", [])
    run_python_script("check-radishflow-export-artifact-metadata-assembly.py", [])
    run_python_script("check-radishflow-export-artifact-metadata-assembly-negative.py", [])
    run_python_script("check-radishflow-export-validator-support-artifacts.py", [])


def check_generated_eval_metadata() -> None:
    schema = json.loads(
        (REPO_ROOT / "datasets/eval/negative-replay-index.schema.json").read_text(encoding="utf-8")
    )
    real_derived_index_schema = json.loads(
        (REPO_ROOT / "datasets/eval/real-derived-negative-index.schema.json").read_text(encoding="utf-8")
    )
    artifact_summary_schema = json.loads(
        (REPO_ROOT / "datasets/eval/batch-orchestration-summary.schema.json").read_text(encoding="utf-8")
    )
    recommended_summary_schema = json.loads(
        (REPO_ROOT / "datasets/eval/recommended-negative-replay-summary.schema.json").read_text(encoding="utf-8")
    )
    for batch in RADISH_DOCS_QA_REAL_BATCHES:
        run_python_script(
            "build-negative-replay-index.py",
            [
                "--audit-report",
                batch["audit_report"],
                "--negative-sample-dir",
                batch["negative_output_dir"],
                "--output",
                batch["replay_index"],
                "--check",
            ],
        )

        document = json.loads((REPO_ROOT / batch["replay_index"]).read_text(encoding="utf-8"))
        jsonschema.validate(document, schema)

        artifact_summary_document = json.loads((REPO_ROOT / batch["artifact_summary"]).read_text(encoding="utf-8"))
        jsonschema.validate(artifact_summary_document, artifact_summary_schema)

        recommended_summary_document = json.loads((REPO_ROOT / batch["recommended_summary"]).read_text(encoding="utf-8"))
        jsonschema.validate(recommended_summary_document, recommended_summary_schema)

        cross_sample_recommended_summary = batch.get("cross_sample_recommended_summary")
        if cross_sample_recommended_summary:
            cross_sample_recommended_summary_document = json.loads(
                (REPO_ROOT / cross_sample_recommended_summary).read_text(encoding="utf-8")
            )
            jsonschema.validate(cross_sample_recommended_summary_document, recommended_summary_schema)

        run_python_script(
            "build-radish-docs-negative-replay.py",
            [
                "--index",
                batch["replay_index"],
                "--output-dir",
                batch["negative_output_dir"],
                "--check",
            ],
        )

        cross_sample_replay_index = batch.get("cross_sample_replay_index")
        cross_sample_negative_output_dir = batch.get("cross_sample_negative_output_dir")
        if cross_sample_replay_index and cross_sample_negative_output_dir:
            run_python_script(
                "build-negative-replay-index.py",
                [
                    "--audit-report",
                    batch["audit_report"],
                    "--negative-sample-dir",
                    cross_sample_negative_output_dir,
                    "--output",
                    cross_sample_replay_index,
                    "--check",
                ],
            )

            cross_sample_document = json.loads((REPO_ROOT / cross_sample_replay_index).read_text(encoding="utf-8"))
            jsonschema.validate(cross_sample_document, schema)

            run_python_script(
                "build-radish-docs-negative-replay.py",
                [
                    "--index",
                    cross_sample_replay_index,
                    "--replay-mode",
                    "cross_sample",
                    "--check",
                ],
            )

            run_python_script(
                "run-eval-regression.py",
                [
                    "radish-docs-qa-negative",
                    "--negative-replay-index",
                    cross_sample_replay_index,
                    "--replay-mode",
                    "cross_sample",
                    "--fail-on-violation",
                ],
            )

            if cross_sample_recommended_summary:
                run_python_script(
                    "run-eval-regression.py",
                    [
                        "radish-docs-qa-negative",
                        "--batch-artifact-summary",
                        batch["artifact_summary"],
                        "--recommended-groups-top",
                        "1",
                        "--replay-mode",
                        "cross_sample",
                        "--fail-on-violation",
                    ],
                )

                run_python_script(
                    "run-radish-docs-qa-negative-recommended.py",
                    [
                        "--batch-artifact-summary",
                        batch["artifact_summary"],
                        "--top",
                        batch["cross_sample_recommended_top"],
                        "--replay-mode",
                        "cross_sample",
                        "--fail-on-violation",
                        "--summary-output",
                        cross_sample_recommended_summary,
                        "--check",
                    ],
                )

        run_python_script(
            "run-eval-regression.py",
            [
                "radish-docs-qa-negative",
                "--batch-artifact-summary",
                batch["artifact_summary"],
                "--recommended-groups-top",
                "1",
                "--fail-on-violation",
            ],
        )

        run_python_script(
            "run-radish-docs-qa-negative-recommended.py",
            [
                "--batch-artifact-summary",
                batch["artifact_summary"],
                "--top",
                batch["recommended_top"],
                "--fail-on-violation",
                "--summary-output",
                batch["recommended_summary"],
                "--check",
            ],
        )

    run_python_script("check-radish-docs-qa-real-batch-dual-recommended-summary.py", [])
    run_python_script("check-radish-docs-qa-real-batch-cross-sample-only-summary.py", [])
    run_python_script("check-radish-docs-qa-real-batch-same-sample-only-summary.py", [])

    run_python_script(
        "build-real-derived-negative-index.py",
        [
            "--manifest",
            RADISH_DOCS_QA_REAL_DERIVED_NEGATIVES["manifest"],
            "--negative-sample-dir",
            RADISH_DOCS_QA_REAL_DERIVED_NEGATIVES["negative_sample_dir"],
            "--output",
            RADISH_DOCS_QA_REAL_DERIVED_NEGATIVES["index"],
            "--check",
        ],
    )
    real_derived_index_document = load_real_derived_negative_index(
        REPO_ROOT / RADISH_DOCS_QA_REAL_DERIVED_NEGATIVES["index"]
    )
    jsonschema.validate(real_derived_index_document, real_derived_index_schema)
    run_python_script("check-radish-docs-qa-real-derived-negative-index.py", [])
    check_committed_candidate_record_batches()


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--skip-text-files", action="store_true")
    return parser.parse_args()


def main() -> int:
    args = parse_args()

    if not args.skip_text_files:
        run_python_script("check-text-files.py", [])

    run_python_script("run-eval-regression.py", ["radish-docs-qa", "--fail-on-violation"])
    run_python_script("run-eval-regression.py", ["radishflow-control-plane", "--fail-on-violation"])
    run_python_script("run-eval-regression.py", ["radishflow-diagnostics", "--fail-on-violation"])
    run_python_script("run-eval-regression.py", ["radishflow-ghost-completion", "--fail-on-violation"])
    run_python_script("run-eval-regression.py", ["radishflow-suggest-edits", "--fail-on-violation"])

    check_required_files()
    check_content_baseline()
    check_contract_schemas()
    check_generated_eval_metadata()

    print("repository baseline checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
