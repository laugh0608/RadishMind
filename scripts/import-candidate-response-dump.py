#!/usr/bin/env python3
from __future__ import annotations

import argparse
import sys
from pathlib import Path

try:
    import jsonschema
except ModuleNotFoundError:
    print(
        "python package 'jsonschema' is required for import-candidate-response-dump.py. "
        "Install it in the active environment before importing dumps.",
        file=sys.stderr,
    )
    raise SystemExit(2)

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import (  # noqa: E402
    candidate_response_record_from_dump,
    load_json_document,
    make_repo_relative,
    resolve_relative_to_repo,
    write_json_document,
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Import a raw candidate response dump into candidate_response_record format.",
    )
    parser.add_argument("--input", required=True, help="Input raw candidate dump json path.")
    parser.add_argument("--output", required=True, help="Output candidate response record json path.")
    parser.add_argument(
        "--record-id",
        default="",
        help="Optional record_id override. Defaults to dump.record_id, then dump.dump_id.",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    input_path = resolve_relative_to_repo(args.input)
    output_path = resolve_relative_to_repo(args.output)

    if not input_path.is_file():
        raise SystemExit(f"input file not found: {args.input}")

    dump = load_json_document(input_path)
    input_label = make_repo_relative(input_path)
    output_label = make_repo_relative(output_path)
    record = candidate_response_record_from_dump(
        dump,
        record_id_override=args.record_id,
        label=input_label,
    )
    write_json_document(output_path, record)
    print(f"wrote candidate response record to {output_label}", file=sys.stdout)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
