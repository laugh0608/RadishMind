#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import subprocess
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent

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
    parser.add_argument("--output-root", required=True, help="Batch output root.")
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


def run_command(command: list[str]) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(
        command,
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
    )
    if result.stdout:
        print(result.stdout, end="")
    if result.stderr:
        print(result.stderr, end="", file=sys.stderr)
    return result


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


def main() -> int:
    args = parse_args()
    output_root = resolve_repo_relative(args.output_root)
    manifest_path = derive_manifest_path(args, output_root)
    report_path = derive_report_path(args, output_root)
    sample_paths = select_sample_paths(args)

    with tempfile.TemporaryDirectory(prefix="radishflow-ghost-poc-") as temp_dir:
        temp_path = Path(temp_dir)
        for sample_path in sample_paths:
            (temp_path / sample_path.name).write_text(sample_path.read_text(encoding="utf-8"), encoding="utf-8")

        inference_command = [
            sys.executable,
            str(REPO_ROOT / "scripts/run-copilot-inference.py"),
            "--sample-dir",
            str(temp_path),
            "--sample-pattern",
            "*.json",
            "--provider",
            args.provider,
            "--output-root",
            str(output_root),
            "--collection-batch",
            args.collection_batch,
            "--temperature",
            str(args.temperature),
            "--manifest-output",
            str(manifest_path),
            "--manifest-description",
            args.manifest_description.strip()
            or "RadishFlow suggest_ghost_completion 首批 Tab/manual_only/empty teacher PoC capture 批次。",
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
        if args.resume:
            inference_command.append("--resume")
        if args.continue_on_error:
            inference_command.append("--continue-on-error")

        inference_result = run_command(inference_command)
        if inference_result.returncode != 0 and not args.continue_on_error:
            return inference_result.returncode

        audit_result: subprocess.CompletedProcess[str] | None = None
        if not args.skip_audit and manifest_path.is_file():
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
            "manifest_path": make_repo_relative(manifest_path) if manifest_path.is_file() else "",
            "audit_report_path": make_repo_relative(report_path) if report_path.is_file() else "",
            "sample_paths": [make_repo_relative(path) for path in sample_paths],
            "execution": {
                "inference_exit_code": inference_result.returncode,
                "audit_exit_code": None if audit_result is None else audit_result.returncode,
            },
        }
        print(json.dumps(summary, ensure_ascii=False, indent=2))
        return 0


if __name__ == "__main__":
    raise SystemExit(main())
