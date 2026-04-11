#!/usr/bin/env python3
from __future__ import annotations

import argparse
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from adapters.radishflow import build_adapter_snapshot_from_export_snapshot  # noqa: E402
from services.runtime.candidate_records import (  # noqa: E402
    ensure_schema,
    load_json_document,
    make_repo_relative,
    resolve_relative_to_repo,
    write_json_document,
)


EXPORT_SNAPSHOT_SCHEMA_PATH = REPO_ROOT / "contracts/radishflow-export-snapshot.schema.json"
ADAPTER_SNAPSHOT_SCHEMA_PATH = REPO_ROOT / "contracts/radishflow-adapter-snapshot.schema.json"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Build a minimal radishflow adapter snapshot from a more source-like export snapshot.",
    )
    parser.add_argument("--input", required=True, help="Path to a radishflow export snapshot json file.")
    parser.add_argument("--output", default="", help="Optional output adapter snapshot json path.")
    parser.add_argument(
        "--check-snapshot",
        default="",
        help="Optional adapter snapshot path. Fails unless generated output exactly matches it.",
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="Do not rewrite output; fail if the generated adapter snapshot does not match the existing output file.",
    )
    return parser.parse_args()


def resolve_expected_snapshot(document: Any, label: str) -> dict[str, Any]:
    if isinstance(document, dict):
        return document
    raise SystemExit(f"{label} must be a json object")


def main() -> int:
    args = parse_args()
    input_path = resolve_relative_to_repo(args.input)
    if not input_path.is_file():
        raise SystemExit(f"input file not found: {args.input}")

    export_snapshot = load_json_document(input_path)
    ensure_schema(export_snapshot, EXPORT_SNAPSHOT_SCHEMA_PATH, make_repo_relative(input_path))

    adapter_snapshot = build_adapter_snapshot_from_export_snapshot(export_snapshot)
    ensure_schema(adapter_snapshot, ADAPTER_SNAPSHOT_SCHEMA_PATH, "generated radishflow adapter snapshot")

    if args.check_snapshot.strip():
        expected_path = resolve_relative_to_repo(args.check_snapshot)
        if not expected_path.is_file():
            raise SystemExit(f"expected snapshot not found for --check-snapshot: {args.check_snapshot}")
        expected_snapshot = resolve_expected_snapshot(load_json_document(expected_path), make_repo_relative(expected_path))
        if expected_snapshot != adapter_snapshot:
            raise SystemExit(f"generated adapter snapshot does not match expected snapshot: {args.check_snapshot}")
        print(f"radishflow export snapshot assembly is up to date: {make_repo_relative(expected_path)}")

    if not args.output.strip():
        if args.check:
            raise SystemExit("--check requires --output")
        if args.check_snapshot.strip():
            return 0
        raise SystemExit("--output is required unless --check-snapshot is used")

    output_path = resolve_relative_to_repo(args.output)
    if args.check:
        if not output_path.is_file():
            raise SystemExit(f"expected output file not found for --check: {args.output}")
        existing = load_json_document(output_path)
        if existing != adapter_snapshot:
            raise SystemExit(f"generated adapter snapshot does not match expected output: {args.output}")
        print(f"radishflow export snapshot assembly is up to date: {make_repo_relative(output_path)}")
        return 0

    write_json_document(output_path, adapter_snapshot)
    print(f"wrote radishflow adapter snapshot to {make_repo_relative(output_path)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
