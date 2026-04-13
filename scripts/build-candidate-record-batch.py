#!/usr/bin/env python3
from __future__ import annotations

import argparse
import sys
from pathlib import Path

try:
    import jsonschema
except ModuleNotFoundError:
    print(
        "python package 'jsonschema' is required for build-candidate-record-batch.py. "
        "Install it in the active environment before generating manifests.",
        file=sys.stderr,
    )
    raise SystemExit(2)

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import (  # noqa: E402
    build_candidate_record_batch_manifest,
    make_repo_relative,
    resolve_relative_to_repo,
    write_json_document,
)

IGNORED_JSON_SUFFIXES = (
    ".manifest.json",
    ".audit.json",
    ".artifacts.json",
    ".negative-replay-index.json",
    ".real-derived-index.json",
    ".summary.json",
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Build a candidate record batch manifest from a directory of candidate_response_record files.",
    )
    parser.add_argument("--record-dir", required=True, help="Directory containing candidate response record json files.")
    parser.add_argument("--output", required=True, help="Output manifest json path.")
    parser.add_argument("--description", default="", help="Optional manifest description.")
    parser.add_argument("--project", default="", help="Optional project override; otherwise inferred from records.")
    parser.add_argument("--task", default="", help="Optional task override; otherwise inferred from records.")
    parser.add_argument("--source", default="", help="Optional source override; otherwise inferred from records.")
    parser.add_argument(
        "--collection-batch",
        default="",
        help="Optional collection batch override; otherwise inferred from capture_metadata.collection_batch.",
    )
    parser.add_argument(
        "--capture-origin",
        default="",
        help="Optional capture origin override; otherwise inferred from capture_metadata.capture_origin when consistent.",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    record_dir = resolve_relative_to_repo(args.record_dir)
    output_path = resolve_relative_to_repo(args.output)

    if not record_dir.is_dir():
        raise SystemExit(f"record directory does not exist: {args.record_dir}")

    record_paths = sorted(
        path
        for path in record_dir.glob("*.json")
        if not any(path.name.endswith(suffix) for suffix in IGNORED_JSON_SUFFIXES)
    )
    if len(record_paths) == 0:
        raise SystemExit(f"no candidate record files found in: {args.record_dir}")

    manifest = build_candidate_record_batch_manifest(
        record_paths,
        description=args.description,
        project_override=args.project,
        task_override=args.task,
        source_override=args.source,
        collection_batch_override=args.collection_batch,
        capture_origin_override=args.capture_origin,
    )
    write_json_document(output_path, manifest)

    print(
        f"wrote manifest with {len(manifest['records'])} records to {make_repo_relative(output_path)}",
        file=sys.stdout,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
