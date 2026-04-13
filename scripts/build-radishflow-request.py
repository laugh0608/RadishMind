#!/usr/bin/env python3
from __future__ import annotations

import argparse
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from adapters.radishflow import build_request_from_snapshot  # noqa: E402
from services.runtime.candidate_records import (  # noqa: E402
    ensure_schema,
    load_json_document,
    make_repo_relative,
    resolve_relative_to_repo,
    write_json_document,
)


ADAPTER_SNAPSHOT_SCHEMA_PATH = REPO_ROOT / "contracts/radishflow-adapter-snapshot.schema.json"
COPILOT_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-request.schema.json"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Build a CopilotRequest for RadishFlow explain/control-plane/edit tasks "
            "from a minimal adapter snapshot."
        ),
    )
    parser.add_argument("--input", required=True, help="Path to a radishflow adapter snapshot json file.")
    parser.add_argument("--output", default="", help="Optional output CopilotRequest json path.")
    parser.add_argument(
        "--check-sample",
        default="",
        help="Optional eval sample path. Fails unless generated request exactly matches sample.input_request.",
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="Do not rewrite output; fail if the generated request does not match the existing output file.",
    )
    return parser.parse_args()


def resolve_expected_request(document: Any, label: str) -> dict[str, Any]:
    if isinstance(document, dict) and isinstance(document.get("input_request"), dict):
        return document["input_request"]
    if isinstance(document, dict):
        return document
    raise SystemExit(f"{label} must be a json object or an eval sample containing input_request")


def main() -> int:
    args = parse_args()
    input_path = resolve_relative_to_repo(args.input)
    if not input_path.is_file():
        raise SystemExit(f"input file not found: {args.input}")

    snapshot = load_json_document(input_path)
    ensure_schema(snapshot, ADAPTER_SNAPSHOT_SCHEMA_PATH, make_repo_relative(input_path))

    request = build_request_from_snapshot(snapshot)
    ensure_schema(request, COPILOT_REQUEST_SCHEMA_PATH, "generated radishflow CopilotRequest")

    if args.check_sample.strip():
        sample_path = resolve_relative_to_repo(args.check_sample)
        if not sample_path.is_file():
            raise SystemExit(f"expected sample not found for --check-sample: {args.check_sample}")
        expected_request = resolve_expected_request(load_json_document(sample_path), make_repo_relative(sample_path))
        if expected_request != request:
            raise SystemExit(f"generated request does not match expected sample input_request: {args.check_sample}")
        print(f"radishflow adapter request assembly is up to date: {make_repo_relative(sample_path)}")

    if not args.output.strip():
        if args.check:
            raise SystemExit("--check requires --output")
        if args.check_sample.strip():
            return 0
        raise SystemExit("--output is required unless --check-sample is used")

    output_path = resolve_relative_to_repo(args.output)
    if args.check:
        if not output_path.is_file():
            raise SystemExit(f"expected output file not found for --check: {args.output}")
        existing = load_json_document(output_path)
        if existing != request:
            raise SystemExit(f"generated request does not match expected output: {args.output}")
        print(f"radishflow adapter request assembly is up to date: {make_repo_relative(output_path)}")
        return 0

    write_json_document(output_path, request)
    print(f"wrote radishflow CopilotRequest to {make_repo_relative(output_path)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
