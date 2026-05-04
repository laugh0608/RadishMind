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
from scripts.eval.radishflow_batch_artifact_summary import (  # noqa: E402
    build_radishflow_batch_artifact_summary_document,
    derive_output_root_from_record_dir,
)
from scripts.eval.report_real_batch_governance_status import build_report as build_real_batch_governance_status_report  # noqa: E402
from scripts.eval.report_suggest_edits_profile_coverage import (  # noqa: E402
    build_report as build_suggest_edits_profile_coverage,
)
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
RADISHFLOW_GHOST_REAL_DERIVED_NEGATIVES = dict(load_fixture_json("radishflow-ghost-real-derived-negatives.json") or {})
RADISHFLOW_SUGGEST_EDITS_REAL_DERIVED_NEGATIVES = dict(
    load_fixture_json("radishflow-suggest-edits-real-derived-negatives.json") or {}
)
RADISHFLOW_SUGGEST_EDITS_POC_BATCHES = list(load_fixture_json("radishflow-suggest-edits-poc-batches.json") or [])
REQUIRED_FILES = list(load_fixture_json("required-files.json") or [])
MAX_COMMITTED_PATH_LENGTH = 180
RADISH_CANDIDATE_RECORDS_MAX_PATH_LENGTH = 178
RADISH_CANDIDATE_RECORDS_ROOT = Path("datasets/eval/candidate-records/radish")
RADISHFLOW_CANDIDATE_RECORDS_MAX_PATH_LENGTH = 120
RADISHFLOW_CANDIDATE_RECORDS_ROOT = Path("datasets/eval/candidate-records/radishflow")
RADISHFLOW_ALLOWED_ROOT_ENTRIES = {"README.md", "batches", "dry-run-check"}

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
    for batch in [*RADISHFLOW_GHOST_REAL_BATCHES, *RADISHFLOW_SUGGEST_EDITS_POC_BATCHES]:
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


def check_core_candidate_probe(
    *,
    temp_prefix: str,
    output_dir_name: str,
    summary_name: str,
    candidate_manifest: str,
    expected_summary: str,
    eval_manifest: str | None = None,
) -> None:
    with tempfile.TemporaryDirectory(prefix=temp_prefix) as temp_dir:
        output_dir = Path(temp_dir) / output_dir_name
        summary_path = Path(temp_dir) / summary_name
        run_python_script(
            "run-radishmind-core-candidate.py",
            [
                "--manifest",
                candidate_manifest,
                "--output-dir",
                str(output_dir),
                "--summary-output",
                str(summary_path),
                "--check-summary",
                expected_summary,
            ],
        )
        if eval_manifest is None:
            return
        eval_path = Path(temp_dir) / "offline-eval-run.json"
        run_python_script(
            "run-radishmind-core-offline-eval.py",
            [
                "--manifest",
                eval_manifest,
                "--candidate-summary",
                str(summary_path),
                "--candidate-output-dir",
                str(output_dir),
                "--output",
                str(eval_path),
                "--check-output",
                str(eval_path),
            ],
        )


def check_required_files() -> None:
    for relative_path in REQUIRED_FILES:
        if not (REPO_ROOT / relative_path).is_file():
            raise SystemExit(f"missing required file: {relative_path}")


def iter_tracked_files() -> list[Path]:
    result = subprocess.run(
        ["git", "ls-files", "-z"],
        cwd=REPO_ROOT,
        capture_output=True,
        text=False,
        check=True,
    )
    tracked_paths: list[Path] = []
    for raw_path in result.stdout.split(b"\0"):
        if not raw_path:
            continue
        tracked_paths.append(Path(raw_path.decode("utf-8")))
    return tracked_paths


def check_path_budget() -> None:
    tracked_files = iter_tracked_files()

    too_long_paths = [
        path.as_posix()
        for path in tracked_files
        if len(path.as_posix()) > MAX_COMMITTED_PATH_LENGTH
    ]
    if too_long_paths:
        raise SystemExit(
            "tracked path exceeds repository path budget "
            f"({MAX_COMMITTED_PATH_LENGTH}): {too_long_paths[0]}"
        )

    radishflow_root = REPO_ROOT / RADISHFLOW_CANDIDATE_RECORDS_ROOT
    unexpected_root_entries = sorted(
        child.name
        for child in radishflow_root.iterdir()
        if child.name not in RADISHFLOW_ALLOWED_ROOT_ENTRIES
    )
    if unexpected_root_entries:
        raise SystemExit(
            "unexpected radishflow candidate-record root entry: "
            f"{unexpected_root_entries[0]}"
        )

    radishflow_files = [
        path
        for path in tracked_files
        if path.parts[: len(RADISHFLOW_CANDIDATE_RECORDS_ROOT.parts)] == RADISHFLOW_CANDIDATE_RECORDS_ROOT.parts
    ]
    radish_files = [
        path
        for path in tracked_files
        if path.parts[: len(RADISH_CANDIDATE_RECORDS_ROOT.parts)] == RADISH_CANDIDATE_RECORDS_ROOT.parts
    ]
    for path in radish_files:
        relative_path = path.as_posix()
        if len(relative_path) > RADISH_CANDIDATE_RECORDS_MAX_PATH_LENGTH:
            raise SystemExit(
                "radish candidate-record path exceeds transition budget "
                f"({RADISH_CANDIDATE_RECORDS_MAX_PATH_LENGTH}): {relative_path}"
            )

    for path in radishflow_files:
        relative_path = path.as_posix()
        if len(relative_path) > RADISHFLOW_CANDIDATE_RECORDS_MAX_PATH_LENGTH:
            raise SystemExit(
                "radishflow candidate-record path exceeds strict budget "
                f"({RADISHFLOW_CANDIDATE_RECORDS_MAX_PATH_LENGTH}): {relative_path}"
            )

        relative_to_radishflow_root = path.relative_to(RADISHFLOW_CANDIDATE_RECORDS_ROOT)
        if relative_to_radishflow_root.parts[0] == "batches":
            if len(relative_to_radishflow_root.parts) < 4:
                raise SystemExit(f"incomplete radishflow batch path: {relative_path}")
            month_key = relative_to_radishflow_root.parts[1]
            batch_key = relative_to_radishflow_root.parts[2]
            if len(month_key) != 7 or month_key[4] != "-":
                raise SystemExit(f"invalid radishflow batch month key: {relative_path}")
            if not batch_key.startswith("rfb-") or len(batch_key) != 16:
                raise SystemExit(f"invalid radishflow batch key: {relative_path}")

    for batch in [*RADISHFLOW_GHOST_REAL_BATCHES, *RADISHFLOW_SUGGEST_EDITS_POC_BATCHES]:
        record_dir = Path(batch["record_dir"])
        if record_dir.parts[: len(RADISHFLOW_CANDIDATE_RECORDS_ROOT.parts)] != RADISHFLOW_CANDIDATE_RECORDS_ROOT.parts:
            continue
        relative_record_dir = record_dir.relative_to(RADISHFLOW_CANDIDATE_RECORDS_ROOT)
        if len(relative_record_dir.parts) < 3 or relative_record_dir.parts[0] != "batches":
            raise SystemExit(f"invalid radishflow record_dir layout: {record_dir.as_posix()}")
        batch_key = relative_record_dir.parts[2]
        if not batch_key.startswith("rfb-") or len(batch_key) != 16:
            raise SystemExit(f"invalid radishflow record_dir layout: {record_dir.as_posix()}")
        if len(relative_record_dir.parts) == 3:
            continue
        trailing_segment = relative_record_dir.parts[3]
        if len(relative_record_dir.parts) != 4 or trailing_segment not in {"r", "records"}:
            raise SystemExit(f"invalid radishflow record_dir layout: {record_dir.as_posix()}")


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
        REPO_ROOT / "contracts/copilot-gateway-envelope.schema.json",
        REPO_ROOT / "contracts/copilot-training-sample.schema.json",
        REPO_ROOT / "contracts/image-generation-intent.schema.json",
        REPO_ROOT / "contracts/image-generation-backend-request.schema.json",
        REPO_ROOT / "contracts/image-generation-artifact.schema.json",
        REPO_ROOT / "contracts/radishmind-core-offline-eval-run.schema.json",
        REPO_ROOT / "contracts/radishflow-ghost-candidate-set.schema.json",
        REPO_ROOT / "contracts/radishflow-adapter-snapshot.schema.json",
        REPO_ROOT / "contracts/radishflow-export-snapshot.schema.json",
        REPO_ROOT / "datasets/eval/radishflow-batch-artifact-summary.schema.json",
        REPO_ROOT / "datasets/eval/radishflow-export-smoke-manifest.schema.json",
        REPO_ROOT / "datasets/eval/radishflow-export-smoke-summary.schema.json",
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
    run_python_script(
        "run-radishflow-export-smoke.py",
        [
            "--manifest",
            "scripts/checks/fixtures/radishflow-export-smoke-fixtures.json",
            "--check-summary",
            "scripts/checks/fixtures/radishflow-export-smoke-summary.json",
        ],
    )
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
    radishflow_artifact_summary_schema = json.loads(
        (REPO_ROOT / "datasets/eval/radishflow-batch-artifact-summary.schema.json").read_text(encoding="utf-8")
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

    suggest_edits_coverage = build_suggest_edits_profile_coverage()
    if int(suggest_edits_coverage.get("sample_count") or 0) != 33:
        raise SystemExit("unexpected suggest_flowsheet_edits sample count in profile coverage report")
    if int(suggest_edits_coverage.get("fully_covered_count") or 0) != 33:
        raise SystemExit("suggest_flowsheet_edits four-main-apiyi profile coverage is incomplete")
    teacher_candidates = list(suggest_edits_coverage.get("teacher_comparison_candidates") or [])
    if teacher_candidates:
        raise SystemExit(
            "suggest_flowsheet_edits teacher comparison candidates should be fully closed, "
            f"but found: {teacher_candidates}"
        )

    governance_report = build_real_batch_governance_status_report()
    governance_summary = governance_report.get("summary") or {}
    if int(governance_summary.get("artifact_summary_connected_chain_count") or 0) != 3:
        raise SystemExit("unexpected artifact summary connected chain count in governance status report")
    if int(governance_summary.get("replay_connected_chain_count") or 0) != 3:
        raise SystemExit("unexpected replay connected chain count in governance status report")
    if int(governance_summary.get("cross_sample_replay_connected_chain_count") or 0) != 3:
        raise SystemExit("unexpected cross-sample replay connected chain count in governance status report")
    if int(governance_summary.get("cross_sample_recommended_replay_connected_chain_count") or 0) != 3:
        raise SystemExit("unexpected cross-sample recommended replay connected chain count in governance status report")
    if int(governance_summary.get("real_derived_connected_chain_count") or 0) != 3:
        raise SystemExit("unexpected real-derived connected chain count in governance status report")
    if int(governance_summary.get("full_governance_connected_chain_count") or 0) != 3:
        raise SystemExit("unexpected full governance connected chain count in governance status report")
    if int(governance_summary.get("coverage_visible_chain_count") or 0) != 3:
        raise SystemExit("unexpected coverage visible chain count in governance status report")
    if int(governance_summary.get("replay_asset_gap_chain_count") or 0) != 0:
        raise SystemExit("unexpected replay asset gap chain count in governance status report")
    if int(governance_summary.get("recommended_replay_asset_gap_chain_count") or 0) != 0:
        raise SystemExit("unexpected recommended replay asset gap chain count in governance status report")
    if int(governance_summary.get("real_derived_asset_gap_chain_count") or 0) != 0:
        raise SystemExit("unexpected real-derived asset gap chain count in governance status report")

    governance_chains = list(governance_report.get("chains") or [])
    if len(governance_chains) != 3:
        raise SystemExit("unexpected governance chain count in governance status report")
    priority_queue = list(governance_report.get("priority_queue") or [])
    if [str(item.get("chain_id") or "").strip() for item in priority_queue] != [
        "radishflow-suggest-flowsheet-edits",
        "radishflow-suggest-ghost-completion",
        "radish-answer-docs-question",
    ]:
        raise SystemExit("governance status report priority queue drifted from current M3 baseline")
    if "服务/API 骨架" not in str(governance_report.get("next_mainline_focus") or ""):
        raise SystemExit("governance status report next_mainline_focus drifted from current M3 baseline")

    suggest_chain = next((chain for chain in governance_chains if chain.get("chain_id") == "radishflow-suggest-flowsheet-edits"), None)
    ghost_chain = next((chain for chain in governance_chains if chain.get("chain_id") == "radishflow-suggest-ghost-completion"), None)
    docs_chain = next((chain for chain in governance_chains if chain.get("chain_id") == "radish-answer-docs-question"), None)
    if suggest_chain is None or ghost_chain is None or docs_chain is None:
        raise SystemExit("governance status report is missing one or more expected chains")

    suggest_governance = suggest_chain.get("governance") or {}
    suggest_coverage_summary = suggest_chain.get("coverage_summary") or {}
    if int(suggest_coverage_summary.get("real_captured_sample_count") or 0) != 33:
        raise SystemExit("radishflow suggest edits: unexpected real captured sample count in governance status report")
    if int(suggest_coverage_summary.get("latest_batch_matched_sample_count") or 0) != 6:
        raise SystemExit("radishflow suggest edits: unexpected latest batch matched sample count in governance status report")
    suggest_priority = suggest_chain.get("priority") or {}
    if int(suggest_priority.get("rank") or 0) != 1:
        raise SystemExit("radishflow suggest edits: unexpected priority rank in governance status report")
    suggest_coverage = suggest_chain.get("coverage") or {}
    if str(suggest_coverage.get("next_high_value_capture_group") or "").strip():
        raise SystemExit("radishflow suggest edits: unexpected next high-value group in governance status report")
    for blocker_key in (
        "negative_replay_index_blocker",
        "recommended_negative_replay_summary_blocker",
    ):
        if str(suggest_governance.get(blocker_key) or "").strip():
            raise SystemExit(f"radishflow suggest edits governance chain should not report blocker '{blocker_key}'")
    for blocker_key in (
        "cross_sample_negative_replay_index_blocker",
        "cross_sample_recommended_negative_replay_summary_blocker",
    ):
        if str(suggest_governance.get(blocker_key) or "").strip():
            raise SystemExit(f"radishflow suggest edits: unexpected {blocker_key} in governance status report")
    if str(suggest_governance.get("real_derived_negative_index_blocker") or "").strip():
        raise SystemExit("radishflow suggest edits: unexpected real-derived blocker in governance status report")

    ghost_governance = ghost_chain.get("governance") or {}
    ghost_coverage_summary = ghost_chain.get("coverage_summary") or {}
    if int(ghost_coverage_summary.get("real_captured_sample_count") or 0) != 78:
        raise SystemExit("radishflow ghost completion: unexpected real captured sample count in governance status report")
    if int(ghost_coverage_summary.get("latest_batch_matched_sample_count") or 0) != 4:
        raise SystemExit("radishflow ghost completion: unexpected latest batch matched sample count in governance status report")
    ghost_priority = ghost_chain.get("priority") or {}
    if int(ghost_priority.get("rank") or 0) != 2:
        raise SystemExit("radishflow ghost completion: unexpected priority rank in governance status report")
    ghost_coverage = ghost_chain.get("coverage") or {}
    if str(ghost_coverage.get("next_real_capture_group") or "").strip():
        raise SystemExit("radishflow ghost completion: unexpected next real capture group in governance status report")
    for blocker_key in (
        "negative_replay_index_blocker",
        "recommended_negative_replay_summary_blocker",
    ):
        if str(ghost_governance.get(blocker_key) or "").strip():
            raise SystemExit(f"radishflow ghost completion governance chain should not report blocker '{blocker_key}'")
    for blocker_key in (
        "cross_sample_negative_replay_index_blocker",
        "cross_sample_recommended_negative_replay_summary_blocker",
    ):
        if str(ghost_governance.get(blocker_key) or "").strip():
            raise SystemExit(f"radishflow ghost completion governance chain should not report blocker '{blocker_key}'")
    if str(ghost_governance.get("real_derived_negative_index_blocker") or "").strip():
        raise SystemExit("radishflow ghost completion: unexpected real-derived blocker in governance status report")

    docs_governance = docs_chain.get("governance") or {}
    docs_coverage_summary = docs_chain.get("coverage_summary") or {}
    if int(docs_coverage_summary.get("real_captured_sample_count") or 0) < int(
        docs_coverage_summary.get("latest_batch_matched_sample_count") or 0
    ):
        raise SystemExit("radish docs governance coverage summary should not shrink below latest batch coverage")
    docs_priority = docs_chain.get("priority") or {}
    if int(docs_priority.get("rank") or 0) != 3:
        raise SystemExit("radish docs governance chain: unexpected priority rank in governance status report")
    for blocker_key in (
        "negative_replay_index_blocker",
        "cross_sample_negative_replay_index_blocker",
        "recommended_negative_replay_summary_blocker",
        "cross_sample_recommended_negative_replay_summary_blocker",
        "real_derived_negative_index_blocker",
    ):
        if str(docs_governance.get(blocker_key) or "").strip():
            raise SystemExit(f"radish docs governance chain should not report blocker '{blocker_key}'")

    for batch in [*RADISHFLOW_GHOST_REAL_BATCHES, *RADISHFLOW_SUGGEST_EDITS_POC_BATCHES]:
        artifact_summary_path_value = str(batch.get("artifact_summary") or "").strip()
        if not artifact_summary_path_value:
            continue
        artifact_summary_path = REPO_ROOT / artifact_summary_path_value
        artifact_summary_document = json.loads(artifact_summary_path.read_text(encoding="utf-8"))
        jsonschema.validate(artifact_summary_document, radishflow_artifact_summary_schema)
        regenerated_artifact_summary = build_radishflow_batch_artifact_summary_document(
            output_root=derive_output_root_from_record_dir(REPO_ROOT / batch["record_dir"]),
            manifest_path=REPO_ROOT / batch["manifest"],
            audit_report_path=REPO_ROOT / batch["audit_report"],
            capture_exit_code=int((((artifact_summary_document or {}).get("execution") or {}).get("capture_exit_code")) or 0),
            audit_requested=bool(
                ((((artifact_summary_document or {}).get("artifacts") or {}).get("audit_report")) or {}).get("requested")
            ),
            provider_override=str(artifact_summary_document.get("provider") or "").strip(),
            model_override=str(artifact_summary_document.get("model") or "").strip(),
        )
        assert_json_equal(
            artifact_summary_document,
            regenerated_artifact_summary,
            label=artifact_summary_path_value,
        )
        negative_eval_task = (
            "radishflow-suggest-edits-negative"
            if batch.get("task") == "radishflow-suggest-edits"
            else "radishflow-ghost-completion-negative"
        )
        artifacts = artifact_summary_document.get("artifacts") or {}
        negative_replay_index = (artifacts.get("negative_replay_index") or {}) if isinstance(artifacts, dict) else {}
        recommended_summary = (
            (artifacts.get("recommended_negative_replay_summary") or {}) if isinstance(artifacts, dict) else {}
        )
        if negative_replay_index.get("exists") is True:
            negative_replay_index_path = str(negative_replay_index.get("path") or "").strip()
            run_python_script(
                "build-negative-replay-index.py",
                [
                    "--audit-report",
                    batch["audit_report"],
                    "--negative-sample-dir",
                    "datasets/eval/radishflow-negative",
                    "--output",
                    negative_replay_index_path,
                    "--check",
                ],
            )
            negative_replay_index_document = json.loads((REPO_ROOT / negative_replay_index_path).read_text(encoding="utf-8"))
            jsonschema.validate(negative_replay_index_document, schema)
        if recommended_summary.get("exists") is True:
            recommended_summary_path = str(recommended_summary.get("path") or "").strip()
            recommended_summary_document = json.loads((REPO_ROOT / recommended_summary_path).read_text(encoding="utf-8"))
            jsonschema.validate(recommended_summary_document, recommended_summary_schema)
            recommended_top = str(batch.get("recommended_top") or "").strip()
            if not recommended_top:
                raise SystemExit(
                    f"{artifact_summary_path_value} has recommended summary but fixture is missing recommended_top"
                )
            run_python_script(
                "run-radish-docs-qa-negative-recommended.py",
                [
                    "--batch-artifact-summary",
                    artifact_summary_path_value,
                    "--top",
                    recommended_top,
                    "--replay-mode",
                    "same_sample",
                    "--fail-on-violation",
                    "--summary-output",
                    recommended_summary_path,
                    "--check",
                ],
            )

        cross_sample_negative_replay_index = (
            (artifacts.get("cross_sample_negative_replay_index") or {}) if isinstance(artifacts, dict) else {}
        )
        cross_sample_recommended_summary = (
            (artifacts.get("cross_sample_recommended_negative_replay_summary") or {})
            if isinstance(artifacts, dict)
            else {}
        )
        if cross_sample_negative_replay_index.get("exists") is True:
            cross_sample_negative_replay_index_path = str(cross_sample_negative_replay_index.get("path") or "").strip()
            run_python_script(
                "build-negative-replay-index.py",
                [
                    "--audit-report",
                    batch["audit_report"],
                    "--negative-sample-dir",
                    "datasets/eval/radishflow-negative",
                    "--output",
                    cross_sample_negative_replay_index_path,
                    "--check",
                ],
            )
            cross_sample_negative_replay_index_document = json.loads(
                (REPO_ROOT / cross_sample_negative_replay_index_path).read_text(encoding="utf-8")
            )
            jsonschema.validate(cross_sample_negative_replay_index_document, schema)
            run_python_script(
                "run-eval-regression.py",
                [
                    negative_eval_task,
                    "--negative-replay-index",
                    cross_sample_negative_replay_index_path,
                    "--replay-mode",
                    "cross_sample",
                    "--fail-on-violation",
                ],
            )
        if cross_sample_recommended_summary.get("exists") is True:
            cross_sample_recommended_summary_path = str(cross_sample_recommended_summary.get("path") or "").strip()
            cross_sample_recommended_summary_document = json.loads(
                (REPO_ROOT / cross_sample_recommended_summary_path).read_text(encoding="utf-8")
            )
            jsonschema.validate(cross_sample_recommended_summary_document, recommended_summary_schema)
            cross_sample_recommended_top = str(batch.get("cross_sample_recommended_top") or "").strip()
            if not cross_sample_recommended_top:
                raise SystemExit(
                    f"{artifact_summary_path_value} has cross-sample recommended summary but fixture is missing cross_sample_recommended_top"
                )
            run_python_script(
                "run-eval-regression.py",
                [
                    negative_eval_task,
                    "--batch-artifact-summary",
                    artifact_summary_path_value,
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
                    artifact_summary_path_value,
                    "--top",
                    cross_sample_recommended_top,
                    "--replay-mode",
                    "cross_sample",
                    "--fail-on-violation",
                    "--summary-output",
                    cross_sample_recommended_summary_path,
                    "--check",
                ],
            )

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
    run_python_script(
        "build-real-derived-negative-index.py",
        [
            "--manifest",
            RADISHFLOW_SUGGEST_EDITS_REAL_DERIVED_NEGATIVES["manifest"],
            "--negative-sample-dir",
            RADISHFLOW_SUGGEST_EDITS_REAL_DERIVED_NEGATIVES["negative_sample_dir"],
            "--output",
            RADISHFLOW_SUGGEST_EDITS_REAL_DERIVED_NEGATIVES["index"],
            "--check",
        ],
    )
    suggest_edits_real_derived_index_document = load_real_derived_negative_index(
        REPO_ROOT / RADISHFLOW_SUGGEST_EDITS_REAL_DERIVED_NEGATIVES["index"]
    )
    jsonschema.validate(suggest_edits_real_derived_index_document, real_derived_index_schema)
    run_python_script("check-radishflow-suggest-edits-real-derived-negative-index.py", [])
    run_python_script(
        "build-real-derived-negative-index.py",
        [
            "--manifest",
            RADISHFLOW_GHOST_REAL_DERIVED_NEGATIVES["manifest"],
            "--negative-sample-dir",
            RADISHFLOW_GHOST_REAL_DERIVED_NEGATIVES["negative_sample_dir"],
            "--output",
            RADISHFLOW_GHOST_REAL_DERIVED_NEGATIVES["index"],
            "--check",
        ],
    )
    ghost_real_derived_index_document = load_real_derived_negative_index(
        REPO_ROOT / RADISHFLOW_GHOST_REAL_DERIVED_NEGATIVES["index"]
    )
    jsonschema.validate(ghost_real_derived_index_document, real_derived_index_schema)
    run_python_script("check-radishflow-ghost-real-derived-negative-index.py", [])
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
    run_python_script(
        "check-gateway-service-smoke.py",
        [
            "--check-summary",
            "scripts/checks/fixtures/gateway-service-smoke-summary.json",
        ],
    )
    run_python_script(
        "run-radishflow-gateway-demo.py",
        [
            "--manifest",
            "scripts/checks/fixtures/radishflow-gateway-demo-fixtures.json",
            "--check-summary",
            "scripts/checks/fixtures/radishflow-gateway-demo-summary.json",
            "--check",
        ],
    )
    run_python_script(
        "check-radishflow-gateway-ui-consumption.py",
        [
            "--check-summary",
            "scripts/checks/fixtures/radishflow-gateway-ui-consumption-summary.json",
        ],
    )
    run_python_script(
        "check-radishflow-candidate-edit-handoff.py",
        [
            "--check-summary",
            "scripts/checks/fixtures/radishflow-candidate-edit-handoff-summary.json",
        ],
    )
    run_python_script("check-radishmind-core-baseline-matrix.py", [])
    run_python_script("check-radishmind-core-eval-thresholds.py", [])
    run_python_script("check-radishmind-core-offline-eval-run-contract.py", [])
    run_python_script("check-radishmind-core-candidate-json-cleanup.py", [])
    run_python_script("check-radishmind-core-candidate-parameter-updates.py", [])
    run_python_script("check-radishmind-core-candidate-prompt-policy.py", [])
    run_python_script("check-radishmind-core-candidate-hard-field-freeze.py", [])
    run_python_script("check-radishmind-core-candidate-hard-field-injection.py", [])
    run_python_script("check-radishmind-core-suggest-edits-response-builder.py", [])
    run_python_script("check-radishmind-core-task-scoped-response-builder.py", [])
    run_python_script("check-radishmind-core-candidate-citation-scaffold.py", [])
    run_python_script("check-radishmind-core-candidate-answer-scaffold.py", [])
    run_python_script("check-radishmind-core-candidate-prompt-budget.py", [])
    run_python_script("run-radishmind-core-offline-eval.py", [])
    with tempfile.TemporaryDirectory(prefix="check-repo-core-candidate-") as temp_dir:
        candidate_output_dir = Path(temp_dir) / "candidate-run"
        candidate_summary_path = Path(temp_dir) / "candidate-summary.json"
        run_python_script(
            "run-radishmind-core-candidate.py",
            [
                "--output-dir",
                str(candidate_output_dir),
                "--summary-output",
                str(candidate_summary_path),
                "--check-summary",
                "scripts/checks/fixtures/radishmind-core-candidate-dry-run-summary.json",
            ],
        )
        run_python_script(
            "run-radishmind-core-offline-eval.py",
            [
                "--manifest",
                "scripts/checks/fixtures/radishmind-core-offline-eval-candidate-run-manifest.json",
                "--candidate-summary",
                str(candidate_summary_path),
                "--candidate-output-dir",
                str(candidate_output_dir),
                "--check-output",
                "scripts/checks/fixtures/radishmind-core-offline-eval-candidate-dry-run.json",
            ],
        )
    check_core_candidate_probe(
        temp_prefix="check-repo-core-timeout-probe-",
        output_dir_name="timeout-probe-candidate-run",
        summary_name="timeout-probe-candidate-summary.json",
        candidate_manifest="scripts/checks/fixtures/radishmind-core-timeout-probe-candidate-manifest.json",
        expected_summary="scripts/checks/fixtures/radishmind-core-timeout-probe-candidate-summary.json",
    )
    check_core_candidate_probe(
        temp_prefix="check-repo-core-holdout-probe-",
        output_dir_name="holdout-probe-candidate-run",
        summary_name="holdout-probe-candidate-summary.json",
        candidate_manifest="scripts/checks/fixtures/radishmind-core-holdout-probe-candidate-manifest.json",
        expected_summary="scripts/checks/fixtures/radishmind-core-holdout-probe-candidate-summary.json",
        eval_manifest="scripts/checks/fixtures/radishmind-core-holdout-probe-candidate-eval-manifest.json",
    )
    check_core_candidate_probe(
        temp_prefix="check-repo-core-holdout-probe-v2-",
        output_dir_name="holdout-probe-v2-candidate-run",
        summary_name="holdout-probe-v2-candidate-summary.json",
        candidate_manifest="scripts/checks/fixtures/radishmind-core-holdout-probe-v2-candidate-manifest.json",
        expected_summary="scripts/checks/fixtures/radishmind-core-holdout-probe-v2-candidate-summary.json",
        eval_manifest="scripts/checks/fixtures/radishmind-core-holdout-probe-v2-candidate-eval-manifest.json",
    )
    check_core_candidate_probe(
        temp_prefix="check-repo-core-full-holdout-",
        output_dir_name="full-holdout-candidate-run",
        summary_name="full-holdout-candidate-summary.json",
        candidate_manifest="scripts/checks/fixtures/radishmind-core-full-holdout-candidate-manifest.json",
        expected_summary="scripts/checks/fixtures/radishmind-core-full-holdout-candidate-summary.json",
        eval_manifest="scripts/checks/fixtures/radishmind-core-full-holdout-candidate-eval-manifest.json",
    )
    run_python_script("check-copilot-training-sample-contract.py", [])
    run_python_script("check-copilot-training-dataset-governance.py", [])
    with tempfile.TemporaryDirectory(prefix="check-repo-training-samples-") as temp_dir:
        run_python_script(
            "build-copilot-training-samples.py",
            [
                "--manifest",
                "scripts/checks/fixtures/copilot-training-sample-conversion-manifest.json",
                "--output-jsonl",
                str(Path(temp_dir) / "copilot-training-samples.jsonl"),
                "--check-summary",
                "scripts/checks/fixtures/copilot-training-sample-conversion-summary.json",
            ],
        )
        run_python_script(
            "build-copilot-training-samples.py",
            [
                "--manifest",
                "scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-manifest.json",
                "--output-jsonl",
                str(Path(temp_dir) / "copilot-training-samples-candidate-records.jsonl"),
                "--check-summary",
                "scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-summary.json",
            ],
        )
    run_python_script("check-image-generation-intent-contract.py", [])
    run_python_script("check-image-generation-eval-manifest.py", [])

    check_path_budget()
    check_required_files()
    check_content_baseline()
    check_contract_schemas()
    check_generated_eval_metadata()

    print("repository baseline checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
