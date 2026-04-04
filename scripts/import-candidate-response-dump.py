#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any

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
DUMP_SCHEMA_PATH = REPO_ROOT / "datasets/eval/candidate-response-dump.schema.json"
RECORD_SCHEMA_PATH = REPO_ROOT / "datasets/eval/candidate-response-record.schema.json"
RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"


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


def resolve_relative_to_repo(path_value: str) -> Path:
    path = Path(path_value)
    if not path.is_absolute():
        path = (REPO_ROOT / path).resolve()
    return path


def format_path_label(path: Path) -> str:
    resolved = path.resolve()
    if resolved.is_relative_to(REPO_ROOT):
        return str(resolved.relative_to(REPO_ROOT))
    return str(resolved)


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json '{path}': {exc}") from exc


def load_schema(path: Path) -> Any:
    return load_json(path)


def validate(document: Any, schema: Any, label: str) -> None:
    try:
        jsonschema.validate(document, schema)
    except jsonschema.ValidationError as exc:
        raise SystemExit(f"{label}: schema validation failed: {exc.message}") from exc


def main() -> int:
    args = parse_args()
    input_path = resolve_relative_to_repo(args.input)
    output_path = resolve_relative_to_repo(args.output)

    if not input_path.is_file():
        raise SystemExit(f"input file not found: {args.input}")

    dump_schema = load_schema(DUMP_SCHEMA_PATH)
    record_schema = load_schema(RECORD_SCHEMA_PATH)
    response_schema = load_schema(RESPONSE_SCHEMA_PATH)

    dump = load_json(input_path)
    input_label = format_path_label(input_path)
    output_label = format_path_label(output_path)

    validate(dump, dump_schema, input_label)
    validate(dump.get("response"), response_schema, f"{input_label} response")

    record_id = args.record_id.strip() or str(dump.get("record_id") or "").strip() or str(dump.get("dump_id") or "").strip()
    if not record_id:
        raise SystemExit("record_id could not be resolved from --record-id, dump.record_id, or dump.dump_id")

    record: dict[str, Any] = {
        "schema_version": 1,
        "record_id": record_id,
        "project": dump["project"],
        "task": dump["task"],
        "sample_id": dump["sample_id"],
        "request_id": dump["request_id"],
        "captured_at": dump["captured_at"],
        "source": dump["source"],
        "model": dump["model"],
        "input_record": dump["input_record"],
        "response": dump["response"],
    }
    capture_metadata = dump.get("capture_metadata")
    if capture_metadata is not None:
        record["capture_metadata"] = capture_metadata

    validate(record, record_schema, output_label)

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(record, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"wrote candidate response record to {output_label}", file=sys.stdout)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
