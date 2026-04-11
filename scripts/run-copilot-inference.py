#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
import time
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.inference import (  # noqa: E402
    build_candidate_response_dump,
    run_inference,
    validate_request_document,
    validate_response_document,
)
from services.runtime.candidate_records import (  # noqa: E402
    build_candidate_record_batch_manifest,
    candidate_response_record_from_dump,
    resolve_relative_to_repo,
    write_json_document,
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Run the minimal RadishMind inference flow for supported Copilot tasks.",
    )
    input_group = parser.add_mutually_exclusive_group(required=True)
    input_group.add_argument("--sample", help="Path to an eval sample file; inference uses sample.input_request.")
    input_group.add_argument("--request", help="Path to a CopilotRequest json file.")
    input_group.add_argument("--sample-dir", help="Directory containing eval sample json files for batch inference.")
    parser.add_argument("--provider", choices=["mock", "openai-compatible"], default="mock")
    parser.add_argument(
        "--provider-profile",
        default="",
        help="Optional openai-compatible provider profile override, for example openrouter or deepseek.",
    )
    parser.add_argument("--model", default="", help="Provider model name. Required for openai-compatible when env is absent.")
    parser.add_argument("--base-url", default="", help="Provider base URL or /v1 endpoint for openai-compatible.")
    parser.add_argument("--api-key", default="", help="Provider API key for openai-compatible.")
    parser.add_argument("--temperature", type=float, default=0.0)
    parser.add_argument(
        "--request-timeout-seconds",
        type=float,
        default=120.0,
        help="Per-request timeout for provider calls. Default: 120",
    )
    parser.add_argument("--sample-id", default="", help="Optional sample_id override when using --request.")
    parser.add_argument("--sample-pattern", default="*.json", help="Glob used with --sample-dir. Default: *.json")
    parser.add_argument("--response-output", default="", help="Optional path to write the normalized response json.")
    parser.add_argument("--dump-output", default="", help="Optional path to write a raw candidate response dump.")
    parser.add_argument("--record-output", default="", help="Optional path to write a candidate response record json.")
    parser.add_argument("--record-id", default="", help="Optional record_id override for single-run record output.")
    parser.add_argument("--output-root", default="", help="Batch output root. Writes responses/dumps/records beneath it.")
    parser.add_argument(
        "--collection-batch",
        default="",
        help="Collection batch name written into capture metadata. Required for --sample-dir.",
    )
    parser.add_argument(
        "--capture-origin",
        default="",
        help="Optional capture origin override. Defaults to manual_fixture or adapter_debug_dump by provider.",
    )
    parser.add_argument(
        "--capture-tag",
        action="append",
        default=[],
        help="Additional capture tag. Can be specified multiple times.",
    )
    parser.add_argument("--capture-notes", default="", help="Optional capture notes written into dump/record metadata.")
    parser.add_argument(
        "--max-attempts",
        type=int,
        default=3,
        help="Maximum attempts for retryable provider failures. Default: 3",
    )
    parser.add_argument(
        "--retry-base-delay-seconds",
        type=float,
        default=5.0,
        help="Base delay in seconds before retrying retryable provider failures. Default: 5",
    )
    parser.add_argument(
        "--manifest-output",
        default="",
        help="Optional batch manifest output path. Defaults to <output-root>/<collection-batch>.manifest.json in batch mode.",
    )
    parser.add_argument(
        "--manifest-description",
        default="",
        help="Optional manifest description for batch mode.",
    )
    parser.add_argument(
        "--resume",
        action="store_true",
        help="In batch mode, skip samples whose response/dump/record outputs already exist.",
    )
    parser.add_argument(
        "--continue-on-error",
        action="store_true",
        help="In batch mode, continue processing later samples after retry exhaustion and report failures at the end.",
    )
    return parser.parse_args()


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json '{path}': {exc}") from exc


def load_sample_request(sample_path: Path) -> tuple[str, dict[str, Any]]:
    sample = load_json(sample_path)
    copilot_request = sample.get("input_request")
    sample_id = str(sample.get("sample_id") or "").strip()
    if not sample_id:
        raise SystemExit(f"sample is missing sample_id: {sample_path}")
    if not isinstance(copilot_request, dict):
        raise SystemExit(f"sample is missing input_request: {sample_path}")
    return sample_id, copilot_request


def normalize_capture_tags(values: list[str]) -> list[str]:
    normalized: list[str] = []
    for value in values:
        tag = str(value).strip()
        if tag and tag not in normalized:
            normalized.append(tag)
    return normalized


def build_batch_record_id(provider: str, sample_id: str) -> str:
    prefix = "simulated" if provider == "mock" else "captured"
    return f"{prefix}-{sample_id}"


def is_retryable_inference_error(exc: Exception) -> bool:
    if isinstance(exc, RuntimeError):
        message = str(exc)
        if "provider request failed" not in message:
            return False
        if any(marker in message for marker in ("HTTP 408", "HTTP 409", "HTTP 425", "HTTP 429", "HTTP 500", "HTTP 502", "HTTP 503", "HTTP 504")):
            return True
        if any(
            marker in message
            for marker in (
                "timed out",
                "Temporary failure",
                "temporarily rate-limited",
                "Remote end closed connection without response",
                "Connection reset by peer",
                "Connection aborted",
                "Name or service not known",
                "nodename nor servname provided",
            )
        ):
            return True
    return False


def run_inference_with_retry(
    copilot_request: dict[str, Any],
    *,
    args: argparse.Namespace,
    sample_label: str,
) -> dict[str, Any]:
    max_attempts = max(1, int(args.max_attempts))
    base_delay_seconds = max(0.0, float(args.retry_base_delay_seconds))
    last_exception: Exception | None = None

    for attempt in range(1, max_attempts + 1):
        try:
            return run_inference(
                copilot_request,
                provider=args.provider,
                provider_profile=args.provider_profile.strip() or None,
                model=args.model.strip() or None,
                base_url=args.base_url.strip() or None,
                api_key=args.api_key.strip() or None,
                temperature=args.temperature,
                request_timeout_seconds=args.request_timeout_seconds,
            )
        except Exception as exc:
            last_exception = exc
            if attempt >= max_attempts or not is_retryable_inference_error(exc):
                raise

            delay_seconds = base_delay_seconds * (2 ** (attempt - 1))
            print(
                f"[retry {attempt}/{max_attempts}] {sample_label}: {exc}; sleeping {delay_seconds:.1f}s before retry",
                file=sys.stderr,
            )
            time.sleep(delay_seconds)

    if last_exception is not None:
        raise last_exception
    raise RuntimeError(f"{sample_label}: inference retry loop ended unexpectedly")


def run_single(args: argparse.Namespace) -> int:
    if args.sample:
        sample_path = resolve_relative_to_repo(args.sample)
        sample_id, copilot_request = load_sample_request(sample_path)
    else:
        request_path = resolve_relative_to_repo(args.request)
        copilot_request = load_json(request_path)
        sample_id = args.sample_id.strip() or f"adhoc-{copilot_request.get('task', 'request')}"

    validate_request_document(copilot_request)
    result = run_inference_with_retry(
        copilot_request,
        args=args,
        sample_label=sample_id,
    )
    validate_response_document(result["response"])

    capture_tags = normalize_capture_tags(args.capture_tag)
    dump_document = None
    if args.dump_output or args.record_output:
        dump_document = build_candidate_response_dump(
            copilot_request=copilot_request,
            sample_id=sample_id,
            inference_result=result,
            record_id=args.record_id.strip() or None,
            capture_origin=args.capture_origin.strip() or None,
            collection_batch=args.collection_batch.strip() or None,
            tags=capture_tags,
            notes=args.capture_notes.strip() or None,
        )

    if args.response_output:
        write_json_document(resolve_relative_to_repo(args.response_output), result["response"])
    if args.dump_output and dump_document is not None:
        write_json_document(resolve_relative_to_repo(args.dump_output), dump_document)
    if args.record_output and dump_document is not None:
        record_document = candidate_response_record_from_dump(
            dump_document,
            record_id_override=args.record_id,
            label=sample_id,
        )
        write_json_document(resolve_relative_to_repo(args.record_output), record_document)

    print(json.dumps(result["response"], ensure_ascii=False, indent=2))
    return 0


def run_batch(args: argparse.Namespace) -> int:
    if not args.output_root.strip():
        raise SystemExit("--output-root is required when using --sample-dir")
    sample_dir = resolve_relative_to_repo(args.sample_dir)
    output_root = resolve_relative_to_repo(args.output_root)
    collection_batch = args.collection_batch.strip()
    if not sample_dir.is_dir():
        raise SystemExit(f"sample directory not found: {args.sample_dir}")
    if not collection_batch:
        raise SystemExit("--collection-batch is required when using --sample-dir")

    sample_paths = sorted(
        path for path in sample_dir.glob(args.sample_pattern) if path.is_file() and path.suffix == ".json"
    )
    if not sample_paths:
        raise SystemExit(f"no sample json files matched in {args.sample_dir} with pattern {args.sample_pattern}")

    responses_dir = output_root / "responses"
    dumps_dir = output_root / "dumps"
    records_dir = output_root / "records"
    responses_dir.mkdir(parents=True, exist_ok=True)
    dumps_dir.mkdir(parents=True, exist_ok=True)
    records_dir.mkdir(parents=True, exist_ok=True)
    manifest_path = (
        resolve_relative_to_repo(args.manifest_output)
        if args.manifest_output
        else output_root / f"{collection_batch}.manifest.json"
    )

    capture_tags = normalize_capture_tags(args.capture_tag)
    record_paths: list[Path] = []
    batch_results: list[dict[str, str]] = []
    failed_samples: list[dict[str, str]] = []
    print(
        f"[batch-start] provider={args.provider} sample_count={len(sample_paths)} "
        f"output_root={output_root} collection_batch={collection_batch}"
    )

    for index, sample_path in enumerate(sample_paths, start=1):
        sample_id, copilot_request = load_sample_request(sample_path)
        validate_request_document(copilot_request)

        response_path = responses_dir / f"{sample_path.stem}.response.json"
        dump_path = dumps_dir / f"{sample_path.stem}.dump.json"
        record_path = records_dir / f"{sample_path.stem}.record.json"
        if args.resume and response_path.is_file() and dump_path.is_file() and record_path.is_file():
            print(f"[skip {index}/{len(sample_paths)}] {sample_id}: existing outputs found")
            record_paths.append(record_path)
            batch_results.append(
                {
                    "sample_id": sample_id,
                    "response_path": str(response_path.relative_to(output_root)),
                    "dump_path": str(dump_path.relative_to(output_root)),
                    "record_path": str(record_path.relative_to(output_root)),
                    "status": "skipped_existing",
                }
            )
            continue

        print(f"[start {index}/{len(sample_paths)}] {sample_id}")
        try:
            result = run_inference_with_retry(
                copilot_request,
                args=args,
                sample_label=sample_id,
            )
            validate_response_document(result["response"])
        except Exception as exc:
            if not args.continue_on_error:
                raise
            failed_samples.append(
                {
                    "sample_id": sample_id,
                    "error": str(exc),
                }
            )
            print(f"[failed] {sample_id}: {exc}", file=sys.stderr)
            continue

        record_id = build_batch_record_id(args.provider, sample_id)
        dump_document = build_candidate_response_dump(
            copilot_request=copilot_request,
            sample_id=sample_id,
            inference_result=result,
            dump_id=f"dump-{sample_id}",
            record_id=record_id,
            capture_origin=args.capture_origin.strip() or None,
            collection_batch=collection_batch,
            tags=capture_tags,
            notes=args.capture_notes.strip() or None,
        )
        record_document = candidate_response_record_from_dump(
            dump_document,
            label=sample_id,
        )

        write_json_document(response_path, result["response"])
        write_json_document(dump_path, dump_document)
        write_json_document(record_path, record_document)

        record_paths.append(record_path)
        batch_results.append(
            {
                "sample_id": sample_id,
                "response_path": str(response_path.relative_to(output_root)),
                "dump_path": str(dump_path.relative_to(output_root)),
                "record_path": str(record_path.relative_to(output_root)),
                "status": "captured",
            }
        )
        print(
            f"[captured {index}/{len(sample_paths)}] {sample_id}: "
            f"record={record_path.relative_to(output_root)}"
        )

    manifest_written = False
    if record_paths:
        manifest = build_candidate_record_batch_manifest(
            record_paths,
            description=args.manifest_description,
            collection_batch_override=collection_batch,
            capture_origin_override=args.capture_origin,
        )
        write_json_document(manifest_path, manifest)
        manifest_written = True

    print(
        json.dumps(
            {
                "mode": "batch",
                "provider": args.provider,
                "sample_count": len(sample_paths),
                "captured_count": sum(1 for item in batch_results if item.get("status") == "captured"),
                "skipped_existing_count": sum(1 for item in batch_results if item.get("status") == "skipped_existing"),
                "failed_count": len(failed_samples),
                "output_root": str(output_root),
                "manifest_path": str(manifest_path) if manifest_written else "",
                "results": batch_results,
                "failures": failed_samples,
            },
            ensure_ascii=False,
            indent=2,
        )
    )
    if failed_samples:
        return 1
    return 0


def main() -> int:
    args = parse_args()
    if args.sample_dir:
        if args.response_output or args.dump_output or args.record_output or args.sample_id or args.record_id:
            raise SystemExit(
                "--response-output/--dump-output/--record-output/--sample-id/--record-id are single-run options and cannot be used with --sample-dir"
            )
        return run_batch(args)
    if args.resume or args.continue_on_error:
        raise SystemExit("--resume and --continue-on-error can only be used with --sample-dir")
    return run_single(args)


if __name__ == "__main__":
    raise SystemExit(main())
