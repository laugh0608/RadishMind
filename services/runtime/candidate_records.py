from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent.parent
DUMP_SCHEMA_PATH = REPO_ROOT / "datasets/eval/candidate-response-dump.schema.json"
RECORD_SCHEMA_PATH = REPO_ROOT / "datasets/eval/candidate-response-record.schema.json"
RECORD_BATCH_SCHEMA_PATH = REPO_ROOT / "datasets/eval/candidate-record-batch.schema.json"
RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
SCHEMA_CACHE: dict[Path, Any] = {}


def resolve_relative_to_repo(path_value: str) -> Path:
    path = Path(path_value)
    if not path.is_absolute():
        path = (REPO_ROOT / path).resolve()
    return path


def make_repo_relative(path: Path) -> str:
    resolved = path.resolve()
    if resolved.is_relative_to(REPO_ROOT):
        return str(resolved.relative_to(REPO_ROOT)).replace("\\", "/")
    return str(resolved)


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json '{path}': {exc}") from exc


def load_schema(path: Path) -> Any:
    if path not in SCHEMA_CACHE:
        SCHEMA_CACHE[path] = load_json_document(path)
    return SCHEMA_CACHE[path]


def ensure_schema(document: Any, schema_path: Path, label: str) -> None:
    try:
        jsonschema.validate(document, load_schema(schema_path))
    except jsonschema.ValidationError as exc:
        raise SystemExit(f"{label}: schema validation failed: {exc.message}") from exc


def resolve_record_id(dump: dict[str, Any], record_id_override: str = "") -> str:
    record_id = (
        record_id_override.strip()
        or str(dump.get("record_id") or "").strip()
        or str(dump.get("dump_id") or "").strip()
    )
    if not record_id:
        raise SystemExit("record_id could not be resolved from --record-id, dump.record_id, or dump.dump_id")
    return record_id


def candidate_response_record_from_dump(
    dump: dict[str, Any],
    *,
    record_id_override: str = "",
    label: str = "candidate response dump",
) -> dict[str, Any]:
    ensure_schema(dump, DUMP_SCHEMA_PATH, label)
    ensure_schema(dump.get("response"), RESPONSE_SCHEMA_PATH, f"{label} response")

    record: dict[str, Any] = {
        "schema_version": 1,
        "record_id": resolve_record_id(dump, record_id_override),
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

    ensure_schema(record, RECORD_SCHEMA_PATH, f"{label} imported record")
    return record


def collect_unique(values: list[str], label: str, *, allow_empty: bool = False) -> str:
    filtered = {value for value in values if value}
    if not filtered:
        if allow_empty:
            return ""
        raise SystemExit(f"unable to infer {label}: all records are missing this value")
    if len(filtered) > 1:
        raise SystemExit(f"unable to infer {label}: found inconsistent values {sorted(filtered)}")
    return next(iter(filtered))


def build_candidate_record_batch_manifest(
    record_paths: list[Path],
    *,
    description: str = "",
    project_override: str = "",
    task_override: str = "",
    source_override: str = "",
    collection_batch_override: str = "",
    capture_origin_override: str = "",
) -> dict[str, Any]:
    if not record_paths:
        raise SystemExit("no candidate record files provided")

    manifests: list[dict[str, str]] = []
    projects: list[str] = []
    tasks: list[str] = []
    sources: list[str] = []
    collection_batches: list[str] = []
    capture_origins: list[str] = []
    seen_record_ids: set[str] = set()

    for path in sorted(record_paths):
        record = load_json_document(path)
        ensure_schema(record, RECORD_SCHEMA_PATH, make_repo_relative(path))

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

    project = project_override.strip() or collect_unique(projects, "project")
    task = task_override.strip() or collect_unique(tasks, "task")
    source = source_override.strip() or collect_unique(sources, "source")
    collection_batch = collection_batch_override.strip() or collect_unique(collection_batches, "collection_batch")
    capture_origin = capture_origin_override.strip() or collect_unique(capture_origins, "capture_origin", allow_empty=True)

    if project_override.strip() and any(value != project for value in projects):
        raise SystemExit("provided --project does not match all record.project values")
    if task_override.strip() and any(value != task for value in tasks):
        raise SystemExit("provided --task does not match all record.task values")
    if source_override.strip() and any(value != source for value in sources):
        raise SystemExit("provided --source does not match all record.source values")
    if collection_batch_override.strip() and any(value != collection_batch for value in collection_batches if value):
        raise SystemExit("provided --collection-batch does not match existing record capture_metadata.collection_batch values")
    if capture_origin_override.strip() and any(value != capture_origin for value in capture_origins if value):
        raise SystemExit("provided --capture-origin does not match existing record capture_metadata.capture_origin values")

    manifest: dict[str, Any] = {
        "schema_version": 1,
        "collection_batch": collection_batch,
        "project": project,
        "task": task,
        "source": source,
        "records": manifests,
    }
    if capture_origin:
        manifest["capture_origin"] = capture_origin
    if description.strip():
        manifest["description"] = description.strip()

    ensure_schema(manifest, RECORD_BATCH_SCHEMA_PATH, "candidate record batch manifest")
    return manifest


def write_json_document(path: Path, document: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
