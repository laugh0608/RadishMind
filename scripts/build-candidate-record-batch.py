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
        "python package 'jsonschema' is required for build-candidate-record-batch.py. "
        "Install it in the active environment before generating manifests.",
        file=sys.stderr,
    )
    raise SystemExit(2)


REPO_ROOT = Path(__file__).resolve().parent.parent
RECORD_SCHEMA_PATH = REPO_ROOT / "datasets/eval/candidate-response-record.schema.json"


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


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json '{path}': {exc}") from exc


def load_schema(path: Path) -> Any:
    return load_json_document(path)


def ensure_schema(document: Any, schema: Any, label: str) -> None:
    try:
        jsonschema.validate(document, schema)
    except jsonschema.ValidationError as exc:
        raise SystemExit(f"{label}: schema validation failed: {exc.message}") from exc


def collect_unique(values: list[str], label: str, *, allow_empty: bool = False) -> str:
    filtered = {value for value in values if value}
    if not filtered:
        if allow_empty:
            return ""
        raise SystemExit(f"unable to infer {label}: all records are missing this value")
    if len(filtered) > 1:
        raise SystemExit(f"unable to infer {label}: found inconsistent values {sorted(filtered)}")
    return next(iter(filtered))


def resolve_relative_to_repo(path_value: str) -> Path:
    path = Path(path_value)
    if not path.is_absolute():
        path = (REPO_ROOT / path).resolve()
    return path


def make_repo_relative(path: Path) -> str:
    return str(path.resolve().relative_to(REPO_ROOT)).replace("\\", "/")


def main() -> int:
    args = parse_args()
    record_dir = resolve_relative_to_repo(args.record_dir)
    output_path = resolve_relative_to_repo(args.output)

    if not record_dir.is_dir():
        raise SystemExit(f"record directory does not exist: {args.record_dir}")

    record_schema = load_schema(RECORD_SCHEMA_PATH)
    record_paths = sorted(
        path for path in record_dir.glob("*.json") if not path.name.endswith(".manifest.json")
    )
    if len(record_paths) == 0:
        raise SystemExit(f"no candidate record files found in: {args.record_dir}")

    manifests: list[dict[str, str]] = []
    projects: list[str] = []
    tasks: list[str] = []
    sources: list[str] = []
    collection_batches: list[str] = []
    capture_origins: list[str] = []
    seen_record_ids: set[str] = set()

    for path in record_paths:
        record = load_json_document(path)
        ensure_schema(record, record_schema, str(path.relative_to(REPO_ROOT)))

        record_id = str(record.get("record_id") or "").strip()
        sample_id = str(record.get("sample_id") or "").strip()
        if record_id in seen_record_ids:
            raise SystemExit(f"duplicate record_id detected while building manifest: {record_id}")
        seen_record_ids.add(record_id)

        manifests.append(
            {
                "record_id": record_id,
                "sample_id": sample_id,
                "path": make_repo_relative(path),
            }
        )
        projects.append(str(record.get("project") or "").strip())
        tasks.append(str(record.get("task") or "").strip())
        sources.append(str(record.get("source") or "").strip())

        capture_metadata = record.get("capture_metadata") or {}
        collection_batches.append(str(capture_metadata.get("collection_batch") or "").strip())
        capture_origins.append(str(capture_metadata.get("capture_origin") or "").strip())

    project = args.project.strip() or collect_unique(projects, "project")
    task = args.task.strip() or collect_unique(tasks, "task")
    source = args.source.strip() or collect_unique(sources, "source")
    collection_batch = args.collection_batch.strip() or collect_unique(collection_batches, "collection_batch")
    capture_origin = args.capture_origin.strip() or collect_unique(capture_origins, "capture_origin", allow_empty=True)

    if args.project.strip() and any(value != project for value in projects):
        raise SystemExit("provided --project does not match all record.project values")
    if args.task.strip() and any(value != task for value in tasks):
        raise SystemExit("provided --task does not match all record.task values")
    if args.source.strip() and any(value != source for value in sources):
        raise SystemExit("provided --source does not match all record.source values")
    if args.collection_batch.strip() and any(value != collection_batch for value in collection_batches if value):
        raise SystemExit("provided --collection-batch does not match existing record capture_metadata.collection_batch values")
    if args.capture_origin.strip() and any(value != capture_origin for value in capture_origins if value):
        raise SystemExit("provided --capture-origin does not match existing record capture_metadata.capture_origin values")

    manifest: dict[str, Any] = {
        "schema_version": 1,
        "collection_batch": collection_batch,
        "project": project,
        "task": task,
        "source": source,
    }
    if capture_origin:
        manifest["capture_origin"] = capture_origin
    if args.description.strip():
        manifest["description"] = args.description.strip()
    manifest["records"] = manifests

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    print(
        f"wrote manifest with {len(manifests)} records to {make_repo_relative(output_path)}",
        file=sys.stdout,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
