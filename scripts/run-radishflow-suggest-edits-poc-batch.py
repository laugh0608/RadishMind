#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import math
import subprocess
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import (  # noqa: E402
    build_candidate_record_batch_manifest,
    candidate_response_record_from_dump,
    derive_candidate_batch_artifact_summary_path,
    derive_candidate_batch_audit_path,
    derive_candidate_batch_dump_path,
    derive_candidate_batch_manifest_path,
    derive_candidate_batch_output_root,
    derive_candidate_batch_record_path,
    derive_candidate_batch_records_dir,
    derive_candidate_batch_response_path,
    derive_candidate_batch_responses_dir,
    derive_candidate_batch_dumps_dir,
    write_json_document,
)
from scripts.eval.radishflow_batch_artifact_summary import (  # noqa: E402
    build_radishflow_batch_artifact_summary_document,
    derive_artifact_summary_path,
)
from services.runtime.inference import (  # noqa: E402
    build_candidate_response_dump,
    validate_request_document,
    validate_response_document,
)

DEFAULT_SAMPLE_PATHS = [
    "datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json",
    "datasets/eval/radishflow/suggest-flowsheet-edits-stream-spec-placeholder-001.json",
    "datasets/eval/radishflow/suggest-flowsheet-edits-three-step-priority-chain-001.json",
]
SAMPLE_GROUP_PATHS = {
    "default-early-trio": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-stream-spec-placeholder-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-three-step-priority-chain-001.json",
    ],
    "default-selection-ordering": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-action-citation-ordering-diagnostic-artifact-snapshot-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-issue-ordering-confirmed-before-unconfirmed-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-same-risk-input-first-ordering-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-selection-chronology-single-actionable-target-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-selection-order-preserved-001.json",
    ],
    "heater-follow-up": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-heater-multi-action-001.json",
    ],
    "cross-object-citation": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-citation-interleaving-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-warning-citation-ordering-001.json",
    ],
    "mixed-risk-cross-object": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-mixed-risk-three-action-ordering-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-plus-pump-parameter-001.json",
    ],
    "triad-mixed-risk-cross-object": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-spec-compressor-placeholder-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-pump-update-plus-separator-placeholder-001.json",
    ],
    "mixed-risk-patch-combo": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-spec-plus-pump-update-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-compressor-mixed-parameter-patch-001.json",
    ],
    "cross-object-primary-focus": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-joint-selection-primary-focus-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-multi-unit-stream-primary-focus-001.json",
    ],
    "parameter-ordering": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-compressor-parameter-placeholder-ordering-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-compressor-parameter-update-ordering-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-compressor-parameter-update-detail-ordering-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-heater-patch-key-ordering-001.json",
    ],
    "range-sequence-ordering": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-efficiency-range-ordering-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-stream-spec-sequence-ordering-001.json",
    ],
    "local-edits": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-compressor-evidence-gap-partial-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-multi-selection-single-actionable-target-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-pump-parameter-combo-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-valve-local-fix-vs-global-balance-001.json",
    ],
    "mixed-risk-citation-reconnect": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-mixed-risk-reconnect-plus-spec-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-citation-ordering-diagnostics-before-artifacts-before-snapshot-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-issue-citation-ordering-warning-artifact-before-snapshot-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-connection-placeholder-ordering-001.json",
    ],
    "high-value-real-expansion-core": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-spec-compressor-placeholder-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-compressor-mixed-parameter-patch-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-joint-selection-primary-focus-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-compressor-parameter-update-detail-ordering-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-compressor-evidence-gap-partial-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-mixed-risk-reconnect-plus-spec-001.json",
    ],
    "high-value-real-expansion-secondary": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-pump-update-plus-separator-placeholder-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-spec-plus-pump-update-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-multi-unit-stream-primary-focus-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-heater-patch-key-ordering-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-multi-selection-single-actionable-target-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-connection-placeholder-ordering-001.json",
    ],
    "remaining-horizontal-gaps": [
        "datasets/eval/radishflow/suggest-flowsheet-edits-action-citation-ordering-diagnostic-artifact-snapshot-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-heater-multi-action-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-issue-ordering-confirmed-before-unconfirmed-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-same-risk-input-first-ordering-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-selection-chronology-single-actionable-target-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-selection-order-preserved-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-stream-spec-placeholder-001.json",
        "datasets/eval/radishflow/suggest-flowsheet-edits-three-step-priority-chain-001.json",
    ],
}
DEFAULT_REQUEST_TIMEOUT_SECONDS = 120.0
# Keep the apiyi_ch timeout bump scoped to the known triad samples that have
# repeatedly shown capture-level timeout pressure under the default 120s.
KNOWN_REQUEST_TIMEOUT_OVERRIDES: dict[str, dict[str, float]] = {
    "apiyi_ch": {
        "radishflow-suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-spec-compressor-placeholder-001": 210.0,
        "radishflow-suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-pump-update-plus-separator-placeholder-001": 210.0,
    }
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Run the minimal RadishFlow suggest_flowsheet_edits batch pipeline: "
            "curated samples -> capture -> manifest -> audit."
        ),
    )
    parser.add_argument("--provider", choices=["mock", "openai-compatible"], default="mock")
    parser.add_argument(
        "--provider-profile",
        default="",
        help="Optional real-provider profile override, for example anyrouter, sub_jlypx, qaq, or google_gemini.",
    )
    parser.add_argument("--model", default="", help="Provider model name override.")
    parser.add_argument("--base-url", default="", help="Provider base URL override.")
    parser.add_argument("--api-key", default="", help="Provider API key override.")
    parser.add_argument("--temperature", type=float, default=0.0)
    parser.add_argument(
        "--request-timeout-seconds",
        type=float,
        default=None,
        help=(
            "Per-request timeout for provider calls. Defaults to env/global 120 unless "
            "a known sample-specific override applies."
        ),
    )
    parser.add_argument(
        "--output-root",
        default="",
        help=(
            "Batch output root. Defaults to "
            "datasets/eval/candidate-records/radishflow/<collection-batch>."
        ),
    )
    parser.add_argument("--collection-batch", required=True, help="Batch collection name.")
    parser.add_argument("--capture-origin", default="", help="Optional capture origin override.")
    parser.add_argument("--capture-tag", action="append", default=[], help="Additional capture tag. Can be repeated.")
    parser.add_argument("--capture-notes", default="", help="Optional capture notes.")
    parser.add_argument("--max-attempts", type=int, default=3, help="Maximum attempts for retryable provider failures.")
    parser.add_argument(
        "--retry-base-delay-seconds",
        type=float,
        default=5.0,
        help="Base delay in seconds before retrying retryable provider failures.",
    )
    parser.add_argument("--manifest-output", default="", help="Optional manifest output path override.")
    parser.add_argument("--manifest-description", default="", help="Optional manifest description.")
    parser.add_argument("--report-output", default="", help="Optional audit report output path override.")
    parser.add_argument(
        "--artifact-summary-output",
        default="",
        help="Optional structured batch artifact summary output path override.",
    )
    parser.add_argument("--resume", action="store_true", help="Skip already captured samples when outputs exist.")
    parser.add_argument(
        "--continue-on-error",
        action="store_true",
        help="Continue later samples after retry exhaustion and still attempt audit if a manifest is produced.",
    )
    parser.add_argument("--skip-audit", action="store_true", help="Only run capture; skip audit.")
    parser.add_argument("--fail-on-audit-violation", action="store_true", help="Fail when audit finds violations.")
    parser.add_argument(
        "--sample-group",
        action="append",
        default=[],
        help=(
            "Named sample group to include. Can be repeated. Supported groups: "
            + ", ".join(sorted(SAMPLE_GROUP_PATHS))
            + "."
        ),
    )
    parser.add_argument(
        "--sample-path",
        action="append",
        default=[],
        help="Eval sample path to include. Can be repeated. Defaults to the curated suggest_edits PoC trio.",
    )
    return parser.parse_args()


def resolve_repo_relative(path_value: str) -> Path:
    path = Path(path_value)
    if not path.is_absolute():
        path = (REPO_ROOT / path).resolve()
    return path


def make_repo_relative(path: Path) -> str:
    resolved = path.resolve()
    if resolved.is_relative_to(REPO_ROOT):
        return str(resolved.relative_to(REPO_ROOT)).replace("\\", "/")
    return str(resolved)


def coerce_subprocess_text(payload: str | bytes | None) -> str:
    if payload is None:
        return ""
    if isinstance(payload, bytes):
        return payload.decode("utf-8", errors="replace")
    return payload


def run_command(command: list[str], *, timeout_seconds: float | None = None) -> subprocess.CompletedProcess[str]:
    try:
        result = subprocess.run(
            command,
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            timeout=timeout_seconds,
        )
    except subprocess.TimeoutExpired as exc:
        stdout_text = coerce_subprocess_text(exc.stdout)
        stderr_text = coerce_subprocess_text(exc.stderr)
        timeout_message = (
            f"command timed out after {timeout_seconds:.1f}s: "
            + " ".join(command)
            + "\n"
        )
        if stdout_text:
            print(stdout_text, end="")
        print(timeout_message, end="", file=sys.stderr)
        if stderr_text:
            print(stderr_text, end="", file=sys.stderr)
        return subprocess.CompletedProcess(
            command,
            124,
            stdout_text,
            timeout_message + stderr_text,
        )
    if result.stdout:
        print(result.stdout, end="")
    if result.stderr:
        print(result.stderr, end="", file=sys.stderr)
    return result


def load_json(path: Path, label: str) -> dict[str, Any]:
    try:
        document = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse {label} '{make_repo_relative(path)}': {exc}") from exc
    if not isinstance(document, dict):
        raise SystemExit(f"{label} must be a json object: {make_repo_relative(path)}")
    return document


def load_sample_bundle(sample_path: Path) -> tuple[str, dict[str, Any], dict[str, Any]]:
    sample = load_json(sample_path, "eval sample")
    sample_id = str(sample.get("sample_id") or "").strip()
    request = sample.get("input_request")
    golden_response = sample.get("golden_response")
    if not sample_id:
        raise SystemExit(f"sample is missing sample_id: {make_repo_relative(sample_path)}")
    if not isinstance(request, dict):
        raise SystemExit(f"sample is missing input_request: {make_repo_relative(sample_path)}")
    if not isinstance(golden_response, dict):
        raise SystemExit(f"sample is missing golden_response: {make_repo_relative(sample_path)}")
    validate_request_document(request)
    validate_response_document(golden_response)
    return sample_id, request, golden_response


def build_record_id(provider: str, sample_id: str) -> str:
    prefix = "simulated" if provider == "mock" else "captured"
    return f"{prefix}-{sample_id}"


def derive_capture_timeout_seconds(
    *,
    provider: str,
    request_timeout_seconds: float | None,
    max_attempts: int,
    retry_base_delay_seconds: float,
) -> float | None:
    if provider != "openai-compatible" or request_timeout_seconds is None:
        return None
    request_timeout_seconds = max(1.0, float(request_timeout_seconds))
    max_attempts = max(1, int(max_attempts))
    retry_delay_budget = 0.0
    if max_attempts > 1:
        base_delay = max(0.0, float(retry_base_delay_seconds))
        retry_delay_budget = base_delay * sum(2 ** retry_index for retry_index in range(max_attempts - 1))
    return math.ceil((request_timeout_seconds * max_attempts) + retry_delay_budget + 30.0)


def resolve_effective_request_timeout_seconds(args: argparse.Namespace, sample_id: str) -> float | None:
    if args.request_timeout_seconds is not None:
        return float(args.request_timeout_seconds)
    if args.provider != "openai-compatible":
        return None
    profile_name = str(args.provider_profile or "").strip()
    if not profile_name:
        return None
    profile_overrides = KNOWN_REQUEST_TIMEOUT_OVERRIDES.get(profile_name, {})
    if sample_id in profile_overrides:
        return float(profile_overrides[sample_id])
    return None


def select_sample_paths(args: argparse.Namespace) -> list[Path]:
    raw_paths: list[str] = []
    for group_name in args.sample_group:
        normalized_group_name = str(group_name).strip()
        if not normalized_group_name:
            continue
        group_paths = SAMPLE_GROUP_PATHS.get(normalized_group_name)
        if group_paths is None:
            supported_groups = ", ".join(sorted(SAMPLE_GROUP_PATHS))
            raise SystemExit(f"unsupported sample group '{normalized_group_name}'; expected one of: {supported_groups}")
        raw_paths.extend(group_paths)
    raw_paths.extend(args.sample_path)
    if not raw_paths:
        raw_paths = list(DEFAULT_SAMPLE_PATHS)
    selected: list[Path] = []
    seen_paths: set[Path] = set()
    for raw_path in raw_paths:
        path = resolve_repo_relative(raw_path)
        if not path.is_file():
            raise SystemExit(f"sample file not found: {raw_path}")
        if path in seen_paths:
            continue
        seen_paths.add(path)
        selected.append(path)
    return selected


def derive_manifest_path(args: argparse.Namespace, output_root: Path) -> Path:
    if args.manifest_output.strip():
        return resolve_repo_relative(args.manifest_output)
    return derive_candidate_batch_manifest_path(output_root)


def derive_report_path(args: argparse.Namespace, output_root: Path) -> Path:
    if args.report_output.strip():
        return resolve_repo_relative(args.report_output)
    return derive_candidate_batch_audit_path(output_root)


def derive_artifact_summary_output_path(args: argparse.Namespace, output_root: Path) -> Path:
    if args.artifact_summary_output.strip():
        return resolve_repo_relative(args.artifact_summary_output)
    return derive_candidate_batch_artifact_summary_path(output_root)


def derive_output_root(args: argparse.Namespace) -> Path:
    if args.output_root.strip():
        return resolve_repo_relative(args.output_root)
    return derive_candidate_batch_output_root(
        project="radishflow",
        collection_batch=args.collection_batch,
    )


def capture_mock_sample(
    *,
    sample_path: Path,
    collection_batch: str,
    capture_origin: str,
    capture_tags: list[str],
    capture_notes: str,
    response_path: Path,
    dump_path: Path,
    record_path: Path,
) -> Path:
    sample_id, request, golden_response = load_sample_bundle(sample_path)
    inference_result = {
        "provider": "mock",
        "model": "radishmind-mock-radishflow-suggest_flowsheet_edits-golden-v1",
        "raw_request": {
            "provider": "mock",
            "mode": "golden_response_fixture",
            "sample_path": make_repo_relative(sample_path),
        },
        "raw_response": {
            "provider": "mock",
            "response_preview": str(golden_response.get("summary") or "").strip(),
        },
        "response": golden_response,
    }
    dump_document = build_candidate_response_dump(
        copilot_request=request,
        sample_id=sample_id,
        inference_result=inference_result,
        record_id=build_record_id("mock", sample_id),
        capture_origin=capture_origin or None,
        collection_batch=collection_batch,
        tags=capture_tags,
        notes=capture_notes or None,
    )
    record_document = candidate_response_record_from_dump(
        dump_document,
        record_id_override=build_record_id("mock", sample_id),
        label=sample_id,
    )
    write_json_document(response_path, golden_response)
    write_json_document(dump_path, dump_document)
    write_json_document(record_path, record_document)
    return record_path


def capture_real_sample(
    *,
    args: argparse.Namespace,
    sample_path: Path,
    collection_batch: str,
    request_timeout_seconds: float | None,
    response_path: Path,
    dump_path: Path,
    record_path: Path,
    timeout_seconds: float | None,
) -> subprocess.CompletedProcess[str]:
    sample_id, _, _ = load_sample_bundle(sample_path)
    command = [
        sys.executable,
        str(REPO_ROOT / "scripts/run-copilot-inference.py"),
        "--sample",
        str(sample_path),
        "--provider",
        args.provider,
        "--temperature",
        str(args.temperature),
        "--response-output",
        str(response_path),
        "--dump-output",
        str(dump_path),
        "--record-output",
        str(record_path),
        "--record-id",
        build_record_id(args.provider, sample_id),
        "--collection-batch",
        collection_batch,
        "--max-attempts",
        str(args.max_attempts),
        "--retry-base-delay-seconds",
        str(args.retry_base_delay_seconds),
    ]
    if request_timeout_seconds is not None:
        command.extend(["--request-timeout-seconds", str(request_timeout_seconds)])
    if args.provider_profile.strip():
        command.extend(["--provider-profile", args.provider_profile.strip()])
    if args.model.strip():
        command.extend(["--model", args.model.strip()])
    if args.base_url.strip():
        command.extend(["--base-url", args.base_url.strip()])
    if args.api_key.strip():
        command.extend(["--api-key", args.api_key.strip()])
    if args.capture_origin.strip():
        command.extend(["--capture-origin", args.capture_origin.strip()])
    if args.capture_notes.strip():
        command.extend(["--capture-notes", args.capture_notes.strip()])
    for tag in args.capture_tag:
        normalized_tag = str(tag).strip()
        if normalized_tag:
            command.extend(["--capture-tag", normalized_tag])
    return run_command(command, timeout_seconds=timeout_seconds)


def build_result_document(
    *,
    provider: str,
    sample_paths: list[Path],
    output_root: Path,
    manifest_path: Path,
    manifest_written: bool,
    failed_samples: list[dict[str, str]],
) -> dict[str, Any]:
    result_items: list[dict[str, str]] = []
    failed_by_sample_id = {item["sample_id"]: item for item in failed_samples}
    for sample_path in sample_paths:
        sample_id, _, _ = load_sample_bundle(sample_path)
        response_path = derive_candidate_batch_response_path(
            output_root=output_root,
            task="suggest_flowsheet_edits",
            sample_id=sample_id,
        )
        dump_path = derive_candidate_batch_dump_path(
            output_root=output_root,
            task="suggest_flowsheet_edits",
            sample_id=sample_id,
        )
        record_path = derive_candidate_batch_record_path(
            output_root=output_root,
            task="suggest_flowsheet_edits",
            sample_id=sample_id,
        )
        status = "failed" if sample_id in failed_by_sample_id else "captured"
        if status == "captured" and not (response_path.is_file() and dump_path.is_file() and record_path.is_file()):
            continue
        result_items.append(
            {
                "sample_id": sample_id,
                "response_path": str(response_path.relative_to(output_root)),
                "dump_path": str(dump_path.relative_to(output_root)),
                "record_path": str(record_path.relative_to(output_root)),
                "status": status,
            }
        )
    return {
        "schema_version": 1,
        "pipeline": "radishflow-suggest-edits-poc-batch",
        "mode": "batch",
        "provider": provider,
        "sample_count": len(sample_paths),
        "captured_count": sum(1 for item in result_items if item["status"] == "captured"),
        "failed_count": len(failed_samples),
        "output_root": str(output_root),
        "manifest_path": str(manifest_path) if manifest_written else "",
        "results": result_items,
        "failures": failed_samples,
    }


def main() -> int:
    args = parse_args()
    output_root = derive_output_root(args)
    manifest_path = derive_manifest_path(args, output_root)
    report_path = derive_report_path(args, output_root)
    artifact_summary_path = derive_artifact_summary_output_path(args, output_root)
    sample_paths = select_sample_paths(args)
    responses_dir = derive_candidate_batch_responses_dir(output_root)
    dumps_dir = derive_candidate_batch_dumps_dir(output_root)
    records_dir = derive_candidate_batch_records_dir(output_root)
    responses_dir.mkdir(parents=True, exist_ok=True)
    dumps_dir.mkdir(parents=True, exist_ok=True)
    records_dir.mkdir(parents=True, exist_ok=True)
    print(
        f"[suggest-edits-poc] provider={args.provider} sample_count={len(sample_paths)} "
        f"output_root={make_repo_relative(output_root)} collection_batch={args.collection_batch}"
    )

    record_paths: list[Path] = []
    failed_samples: list[dict[str, str]] = []
    capture_tags = [
        tag
        for tag in [
            "teacher-poc",
            "radishflow-suggest-edits",
            *[str(tag).strip() for tag in args.capture_tag if str(tag).strip()],
        ]
        if tag
    ]

    for index, sample_path in enumerate(sample_paths, start=1):
        sample_id, _, _ = load_sample_bundle(sample_path)
        response_path = derive_candidate_batch_response_path(
            output_root=output_root,
            task="suggest_flowsheet_edits",
            sample_id=sample_id,
        )
        dump_path = derive_candidate_batch_dump_path(
            output_root=output_root,
            task="suggest_flowsheet_edits",
            sample_id=sample_id,
        )
        record_path = derive_candidate_batch_record_path(
            output_root=output_root,
            task="suggest_flowsheet_edits",
            sample_id=sample_id,
        )
        if args.resume and response_path.is_file() and dump_path.is_file() and record_path.is_file():
            print(f"[skip {index}/{len(sample_paths)}] {sample_id}: existing outputs found")
            record_paths.append(record_path)
            continue

        print(f"[start {index}/{len(sample_paths)}] {sample_id}")
        effective_request_timeout_seconds = resolve_effective_request_timeout_seconds(args, sample_id)
        sample_timeout_seconds = derive_capture_timeout_seconds(
            provider=args.provider,
            request_timeout_seconds=effective_request_timeout_seconds,
            max_attempts=args.max_attempts,
            retry_base_delay_seconds=args.retry_base_delay_seconds,
        )
        if effective_request_timeout_seconds is not None:
            print(
                f"[timeout {index}/{len(sample_paths)}] {sample_id}: "
                f"request_timeout={effective_request_timeout_seconds:.0f}s"
                + (
                    f" hard_timeout={sample_timeout_seconds:.0f}s"
                    if sample_timeout_seconds is not None
                    else ""
                )
            )
        if args.provider == "mock":
            capture_mock_sample(
                sample_path=sample_path,
                collection_batch=args.collection_batch,
                capture_origin=args.capture_origin.strip(),
                capture_tags=capture_tags,
                capture_notes=args.capture_notes.strip(),
                response_path=response_path,
                dump_path=dump_path,
                record_path=record_path,
            )
        else:
            result = capture_real_sample(
                args=args,
                sample_path=sample_path,
                collection_batch=args.collection_batch,
                request_timeout_seconds=effective_request_timeout_seconds,
                response_path=response_path,
                dump_path=dump_path,
                record_path=record_path,
                timeout_seconds=sample_timeout_seconds,
            )
            if result.returncode != 0:
                if not args.continue_on_error:
                    return result.returncode
                failed_samples.append(
                    {
                        "sample_id": sample_id,
                        "sample_path": make_repo_relative(sample_path),
                        "error": f"inference exited with code {result.returncode}",
                    }
                )
                continue

        record_paths.append(record_path)
        print(f"[captured {index}/{len(sample_paths)}] {sample_id}: record={make_repo_relative(record_path)}")

    manifest_written = False
    if record_paths:
        manifest = build_candidate_record_batch_manifest(
            record_paths,
            description=(
                args.manifest_description.strip()
                or (
                    "RadishFlow suggest_flowsheet_edits 首批真实 provider capture 批次，覆盖高风险重连、局部规格占位与三步优先级链。"
                    if args.provider != "mock"
                    else "RadishFlow suggest_flowsheet_edits 首批 mock PoC 批次，覆盖高风险重连、局部规格占位与三步优先级链。"
                )
            ),
            collection_batch_override=args.collection_batch,
            capture_origin_override=args.capture_origin.strip(),
        )
        write_json_document(manifest_path, manifest)
        manifest_written = True

    audit_result: subprocess.CompletedProcess[str] | None = None
    if not args.skip_audit and manifest_path.is_file():
        audit_command = [
            sys.executable,
            str(REPO_ROOT / "scripts/audit-candidate-record-batch.py"),
            "radishflow-suggest-edits",
            "--manifest",
            str(manifest_path),
            "--report-output",
            str(report_path),
        ]
        if args.fail_on_audit_violation:
            audit_command.append("--fail-on-violation")
        audit_result = run_command(audit_command)

    result_document = build_result_document(
        provider=args.provider,
        sample_paths=sample_paths,
        output_root=output_root,
        manifest_path=manifest_path,
        manifest_written=manifest_written,
        failed_samples=failed_samples,
    )
    if manifest_written:
        artifact_summary = build_radishflow_batch_artifact_summary_document(
            output_root=output_root,
            manifest_path=manifest_path,
            audit_report_path=report_path,
            capture_exit_code=1 if failed_samples else 0,
            audit_requested=not args.skip_audit,
            provider_override=args.provider,
            model_override=args.model,
        )
        write_json_document(artifact_summary_path, artifact_summary)
        result_document["artifact_summary_path"] = make_repo_relative(artifact_summary_path)
    print(json.dumps(result_document, ensure_ascii=False, indent=2))

    if failed_samples:
        return 1
    if audit_result is not None and audit_result.returncode != 0:
        return audit_result.returncode
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
