#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import math
import subprocess
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import (  # noqa: E402
    build_candidate_record_batch_manifest,
    write_json_document,
)

DEFAULT_SAMPLE_PATHS = [
    "datasets/eval/radishflow/suggest-ghost-completion-flash-vapor-outlet-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-valve-ambiguous-no-tab-001.json",
    "datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-stop-no-legal-outlet-001.json",
]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Run the minimal RadishFlow ghost completion batch capture pipeline: "
            "curated PoC samples -> inference capture -> manifest -> audit."
        ),
    )
    parser.add_argument("--provider", choices=["mock", "openai-compatible"], default="mock")
    parser.add_argument("--model", default="", help="Provider model name override.")
    parser.add_argument("--base-url", default="", help="Provider base URL override.")
    parser.add_argument("--api-key", default="", help="Provider API key override.")
    parser.add_argument("--temperature", type=float, default=0.0)
    parser.add_argument(
        "--request-timeout-seconds",
        type=float,
        default=120.0,
        help="Per-request timeout for provider calls. Default: 120",
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
    parser.add_argument("--resume", action="store_true", help="Skip already captured samples when outputs exist.")
    parser.add_argument(
        "--continue-on-error",
        action="store_true",
        help="Continue later samples after retry exhaustion and still attempt audit if a manifest is produced.",
    )
    parser.add_argument("--skip-audit", action="store_true", help="Only run capture; skip audit.")
    parser.add_argument("--fail-on-audit-violation", action="store_true", help="Fail when audit finds violations.")
    parser.add_argument(
        "--sample-path",
        action="append",
        default=[],
        help="Eval sample path to include. Can be repeated. Defaults to the curated ghost PoC trio.",
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
        stdout_text = exc.stdout or ""
        stderr_text = exc.stderr or ""
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


def load_sample_id(sample_path: Path) -> str:
    try:
        sample = json.loads(sample_path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse sample json '{make_repo_relative(sample_path)}': {exc}") from exc
    sample_id = str((sample or {}).get("sample_id") or "").strip()
    if not sample_id:
        raise SystemExit(f"sample is missing sample_id: {make_repo_relative(sample_path)}")
    return sample_id


def build_record_id(provider: str, sample_id: str) -> str:
    prefix = "simulated" if provider == "mock" else "captured"
    return f"{prefix}-{sample_id}"


def derive_capture_timeout_seconds(args: argparse.Namespace) -> float | None:
    if args.provider != "openai-compatible":
        return None
    request_timeout_seconds = max(1.0, float(args.request_timeout_seconds))
    max_attempts = max(1, int(args.max_attempts))
    retry_delay_budget = 0.0
    if max_attempts > 1:
        base_delay = max(0.0, float(args.retry_base_delay_seconds))
        retry_delay_budget = base_delay * sum(2 ** retry_index for retry_index in range(max_attempts - 1))
    return math.ceil((request_timeout_seconds * max_attempts) + retry_delay_budget + 30.0)


def select_sample_paths(args: argparse.Namespace) -> list[Path]:
    raw_paths = args.sample_path or DEFAULT_SAMPLE_PATHS
    selected: list[Path] = []
    for raw_path in raw_paths:
        path = resolve_repo_relative(raw_path)
        if not path.is_file():
            raise SystemExit(f"sample file not found: {raw_path}")
        selected.append(path)
    return selected


def derive_manifest_path(args: argparse.Namespace, output_root: Path) -> Path:
    if args.manifest_output.strip():
        return resolve_repo_relative(args.manifest_output)
    return output_root / f"{args.collection_batch}.manifest.json"


def derive_report_path(args: argparse.Namespace, output_root: Path) -> Path:
    if args.report_output.strip():
        return resolve_repo_relative(args.report_output)
    return output_root / f"{args.collection_batch}.audit.json"


def derive_output_root(args: argparse.Namespace) -> Path:
    if args.output_root.strip():
        return resolve_repo_relative(args.output_root)
    return REPO_ROOT / "datasets/eval/candidate-records/radishflow" / args.collection_batch


def main() -> int:
    args = parse_args()
    output_root = derive_output_root(args)
    manifest_path = derive_manifest_path(args, output_root)
    report_path = derive_report_path(args, output_root)
    sample_paths = select_sample_paths(args)
    print(
        f"[ghost-poc] provider={args.provider} sample_count={len(sample_paths)} "
        f"output_root={make_repo_relative(output_root)} collection_batch={args.collection_batch}"
    )
    sample_timeout_seconds = derive_capture_timeout_seconds(args)
    if sample_timeout_seconds is not None:
        print(f"[ghost-poc] per-sample hard timeout={sample_timeout_seconds:.0f}s")

    with tempfile.TemporaryDirectory(prefix="radishflow-ghost-poc-") as temp_dir:
        temp_path = Path(temp_dir)
        for sample_path in sample_paths:
            (temp_path / sample_path.name).write_text(sample_path.read_text(encoding="utf-8"), encoding="utf-8")
        responses_dir = output_root / "responses"
        dumps_dir = output_root / "dumps"
        records_dir = output_root / "records"
        responses_dir.mkdir(parents=True, exist_ok=True)
        dumps_dir.mkdir(parents=True, exist_ok=True)
        records_dir.mkdir(parents=True, exist_ok=True)

        record_paths: list[Path] = []
        failed_samples: list[dict[str, str]] = []
        captured_count = 0
        skipped_existing_count = 0

        for index, sample_path in enumerate(sorted(temp_path.glob("*.json")), start=1):
            sample_id = load_sample_id(sample_path)
            response_path = responses_dir / f"{sample_path.stem}.response.json"
            dump_path = dumps_dir / f"{sample_path.stem}.dump.json"
            record_path = records_dir / f"{sample_path.stem}.record.json"
            if args.resume and response_path.is_file() and dump_path.is_file() and record_path.is_file():
                print(f"[skip {index}/{len(sample_paths)}] {sample_id}: existing outputs found")
                skipped_existing_count += 1
                record_paths.append(record_path)
                continue

            print(f"[start {index}/{len(sample_paths)}] {sample_id}")
            inference_command = [
                sys.executable,
                str(REPO_ROOT / "scripts/run-copilot-inference.py"),
                "--sample",
                str(sample_path),
                "--provider",
                args.provider,
                "--temperature",
                str(args.temperature),
                "--request-timeout-seconds",
                str(args.request_timeout_seconds),
                "--response-output",
                str(response_path),
                "--dump-output",
                str(dump_path),
                "--record-output",
                str(record_path),
                "--record-id",
                build_record_id(args.provider, sample_id),
                "--collection-batch",
                args.collection_batch,
                "--max-attempts",
                str(args.max_attempts),
                "--retry-base-delay-seconds",
                str(args.retry_base_delay_seconds),
            ]
            if args.model.strip():
                inference_command.extend(["--model", args.model.strip()])
            if args.base_url.strip():
                inference_command.extend(["--base-url", args.base_url.strip()])
            if args.api_key.strip():
                inference_command.extend(["--api-key", args.api_key.strip()])
            if args.capture_origin.strip():
                inference_command.extend(["--capture-origin", args.capture_origin.strip()])
            if args.capture_notes.strip():
                inference_command.extend(["--capture-notes", args.capture_notes.strip()])
            for tag in args.capture_tag:
                if str(tag).strip():
                    inference_command.extend(["--capture-tag", str(tag).strip()])

            inference_result = run_command(inference_command, timeout_seconds=sample_timeout_seconds)
            if inference_result.returncode != 0:
                if not args.continue_on_error:
                    return inference_result.returncode
                failed_samples.append(
                    {
                        "sample_id": sample_id,
                        "sample_path": make_repo_relative(sample_path),
                        "error": f"inference exited with code {inference_result.returncode}",
                    }
                )
                continue

            captured_count += 1
            record_paths.append(record_path)
            print(f"[captured {index}/{len(sample_paths)}] {sample_id}: record={make_repo_relative(record_path)}")

        manifest_written = False
        if record_paths:
            manifest = build_candidate_record_batch_manifest(
                record_paths,
                description=(
                    args.manifest_description.strip()
                    or "RadishFlow suggest_ghost_completion 首批 Tab/manual_only/empty teacher PoC capture 批次。"
                ),
                collection_batch_override=args.collection_batch,
                capture_origin_override=args.capture_origin,
            )
            write_json_document(manifest_path, manifest)
            manifest_written = True

        audit_result: subprocess.CompletedProcess[str] | None = None
        if not args.skip_audit and manifest_path.is_file():
            print(f"[ghost-poc] auditing manifest={make_repo_relative(manifest_path)}")
            audit_command = [
                sys.executable,
                str(REPO_ROOT / "scripts/audit-candidate-record-batch.py"),
                "radishflow-ghost-completion",
                "--manifest",
                str(manifest_path),
                "--sample-dir",
                str(temp_path),
                "--report-output",
                str(report_path),
            ]
            if args.fail_on_audit_violation:
                audit_command.append("--fail-on-violation")
            audit_result = run_command(audit_command)
            if audit_result.returncode != 0:
                return audit_result.returncode

        summary = {
            "schema_version": 1,
            "pipeline": "radishflow-ghost-completion-poc-batch",
            "provider": args.provider,
            "collection_batch": args.collection_batch,
            "output_root": make_repo_relative(output_root),
            "manifest_path": make_repo_relative(manifest_path) if manifest_written else "",
            "audit_report_path": make_repo_relative(report_path) if report_path.is_file() else "",
            "sample_paths": [make_repo_relative(path) for path in sample_paths],
            "captured_count": captured_count,
            "skipped_existing_count": skipped_existing_count,
            "failed_count": len(failed_samples),
            "failures": failed_samples,
            "execution": {
                "audit_exit_code": None if audit_result is None else audit_result.returncode,
            },
        }
        print(json.dumps(summary, ensure_ascii=False, indent=2))
        if failed_samples:
            return 1
        return 0


if __name__ == "__main__":
    raise SystemExit(main())
