from __future__ import annotations

import hashlib
import json
import re
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent.parent
DSET_CANDIDATE_RECORDS_ROOT = REPO_ROOT / "datasets/eval/candidate-records"
DUMP_SCHEMA_PATH = REPO_ROOT / "datasets/eval/candidate-response-dump.schema.json"
RECORD_SCHEMA_PATH = REPO_ROOT / "datasets/eval/candidate-response-record.schema.json"
RECORD_BATCH_SCHEMA_PATH = REPO_ROOT / "datasets/eval/candidate-record-batch.schema.json"
RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
SCHEMA_CACHE: dict[Path, Any] = {}
SHORT_LAYOUT_PROJECTS = {"radishflow"}
SHORT_LAYOUT_BATCHES_DIR = "batches"
SHORT_LAYOUT_RECORDS_DIR = "r"
SHORT_LAYOUT_RESPONSES_DIR = "o"
SHORT_LAYOUT_DUMPS_DIR = "d"
LEGACY_LAYOUT_RECORDS_DIR = "records"
LEGACY_LAYOUT_RESPONSES_DIR = "responses"
LEGACY_LAYOUT_DUMPS_DIR = "dumps"

TASK_SAMPLE_KEY_PREFIXES = {
    "suggest_flowsheet_edits": "sfe",
    "suggest_ghost_completion": "sgc",
}


def stable_short_key(value: str, *, prefix: str, length: int = 12) -> str:
    normalized = str(value or "").strip()
    if not normalized:
        raise SystemExit(f"cannot derive short key for empty value: prefix={prefix}")
    digest = hashlib.sha1(normalized.encode("utf-8")).hexdigest()[:length]
    return f"{prefix}-{digest}"


def uses_short_candidate_record_layout(project: str) -> bool:
    return str(project or "").strip() in SHORT_LAYOUT_PROJECTS


def uses_short_candidate_record_output_root(output_root: Path) -> bool:
    resolved = output_root.resolve()
    return resolved.parent.parent.name == SHORT_LAYOUT_BATCHES_DIR


def derive_legacy_candidate_sample_stem(sample_id: str) -> str:
    normalized = str(sample_id or "").strip()
    if normalized.startswith("radish-"):
        return normalized[len("radish-") :]
    return normalized


def derive_collection_batch_month(collection_batch: str) -> str:
    match = re.match(r"(?P<month>\d{4}-\d{2})-\d{2}(?:-|$)", str(collection_batch or "").strip())
    if match is not None:
        return match.group("month")
    return "unknown"


def derive_candidate_batch_key(*, project: str, collection_batch: str) -> str:
    prefix = "rfb" if str(project or "").strip() == "radishflow" else "bat"
    return stable_short_key(collection_batch, prefix=prefix)


def derive_candidate_sample_key(*, task: str, sample_id: str) -> str:
    prefix = TASK_SAMPLE_KEY_PREFIXES.get(str(task or "").strip(), "rec")
    return stable_short_key(sample_id, prefix=prefix)


def derive_candidate_batch_output_root(
    *,
    project: str,
    collection_batch: str,
    batch_key: str = "",
) -> Path:
    normalized_project = str(project or "").strip()
    if not normalized_project:
        raise SystemExit("project is required to derive candidate record output root")
    if uses_short_candidate_record_layout(normalized_project):
        normalized_batch_key = batch_key.strip() or derive_candidate_batch_key(
            project=normalized_project,
            collection_batch=collection_batch,
        )
        return (
            DSET_CANDIDATE_RECORDS_ROOT
            / normalized_project
            / SHORT_LAYOUT_BATCHES_DIR
            / derive_collection_batch_month(collection_batch)
            / normalized_batch_key
        )
    return DSET_CANDIDATE_RECORDS_ROOT / normalized_project / collection_batch


def derive_candidate_batch_manifest_path(output_root: Path) -> Path:
    return output_root / "manifest.json"


def derive_candidate_batch_audit_path(output_root: Path) -> Path:
    return output_root / "audit.json"


def derive_candidate_batch_artifact_summary_path(output_root: Path) -> Path:
    return output_root / "artifacts.json"


def derive_candidate_batch_records_dir(output_root: Path) -> Path:
    return output_root / (
        SHORT_LAYOUT_RECORDS_DIR
        if uses_short_candidate_record_output_root(output_root)
        else LEGACY_LAYOUT_RECORDS_DIR
    )


def derive_candidate_batch_responses_dir(output_root: Path) -> Path:
    return output_root / (
        SHORT_LAYOUT_RESPONSES_DIR
        if uses_short_candidate_record_output_root(output_root)
        else LEGACY_LAYOUT_RESPONSES_DIR
    )


def derive_candidate_batch_dumps_dir(output_root: Path) -> Path:
    return output_root / (
        SHORT_LAYOUT_DUMPS_DIR
        if uses_short_candidate_record_output_root(output_root)
        else LEGACY_LAYOUT_DUMPS_DIR
    )


def derive_candidate_batch_record_path(
    *,
    output_root: Path,
    task: str,
    sample_id: str,
) -> Path:
    file_stem = (
        derive_candidate_sample_key(task=task, sample_id=sample_id)
        if uses_short_candidate_record_output_root(output_root)
        else derive_legacy_candidate_sample_stem(sample_id)
    )
    return derive_candidate_batch_records_dir(output_root) / f"{file_stem}.record.json"


def derive_candidate_batch_response_path(
    *,
    output_root: Path,
    task: str,
    sample_id: str,
) -> Path:
    file_stem = (
        derive_candidate_sample_key(task=task, sample_id=sample_id)
        if uses_short_candidate_record_output_root(output_root)
        else derive_legacy_candidate_sample_stem(sample_id)
    )
    return derive_candidate_batch_responses_dir(output_root) / f"{file_stem}.response.json"


def derive_candidate_batch_dump_path(
    *,
    output_root: Path,
    task: str,
    sample_id: str,
) -> Path:
    file_stem = (
        derive_candidate_sample_key(task=task, sample_id=sample_id)
        if uses_short_candidate_record_output_root(output_root)
        else derive_legacy_candidate_sample_stem(sample_id)
    )
    return derive_candidate_batch_dumps_dir(output_root) / f"{file_stem}.dump.json"


def derive_output_root_from_record_path(record_path: Path) -> Path:
    resolved = record_path.resolve()
    if resolved.parent.name in {"records", SHORT_LAYOUT_RECORDS_DIR}:
        return resolved.parent.parent
    return resolved.parent


def infer_batch_key_from_output_root(output_root: Path) -> str:
    resolved = output_root.resolve()
    if resolved.parent.parent.name == SHORT_LAYOUT_BATCHES_DIR:
        return resolved.name
    return ""


def make_manifest_entry_record_relpath(record_path: Path, output_root: Path) -> str:
    return str(record_path.resolve().relative_to(output_root.resolve())).replace("\\", "/")


def extract_sample_key_from_record_path(record_path: Path) -> str:
    name = record_path.name
    if name.endswith(".record.json"):
        return name[: -len(".record.json")]
    if name.endswith(".json"):
        return name[: -len(".json")]
    return record_path.stem


def resolve_manifest_output_root(manifest_path: Path, manifest: dict[str, Any]) -> Path:
    output_root_value = str(manifest.get("output_root") or "").strip()
    if output_root_value:
        return resolve_relative_to_repo(output_root_value)
    return manifest_path.resolve().parent


def resolve_manifest_record_path(manifest_path: Path, manifest: dict[str, Any], entry: dict[str, Any]) -> Path:
    record_relpath = str(entry.get("record_relpath") or "").strip()
    if record_relpath:
        return resolve_manifest_output_root(manifest_path, manifest) / record_relpath
    path_value = str(entry.get("path") or "").strip()
    if path_value:
        return resolve_relative_to_repo(path_value)
    raise SystemExit(f"{make_repo_relative(manifest_path)} record entry is missing both path and record_relpath")


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
    output_roots: set[Path] = set()
    ordered_records: list[tuple[str, str, Path, dict[str, Any]]] = []

    for path in record_paths:
        record = load_json_document(path)
        ensure_schema(record, RECORD_SCHEMA_PATH, make_repo_relative(path))

        record_id = str(record.get("record_id") or "").strip()
        sample_id = str(record.get("sample_id") or "").strip()
        if record_id in seen_record_ids:
            raise SystemExit(f"duplicate record_id detected while building manifest: {record_id}")
        seen_record_ids.add(record_id)
        projects.append(str(record.get("project") or "").strip())
        tasks.append(str(record.get("task") or "").strip())
        sources.append(str(record.get("source") or "").strip())
        output_roots.add(derive_output_root_from_record_path(path))

        capture_metadata = record.get("capture_metadata") or {}
        collection_batches.append(str(capture_metadata.get("collection_batch") or "").strip())
        capture_origins.append(str(capture_metadata.get("capture_origin") or "").strip())
        ordered_records.append((sample_id, record_id, path, record))

    project = project_override.strip() or collect_unique(projects, "project")
    task = task_override.strip() or collect_unique(tasks, "task")
    source = source_override.strip() or collect_unique(sources, "source")
    collection_batch = collection_batch_override.strip() or collect_unique(collection_batches, "collection_batch")
    capture_origin = capture_origin_override.strip() or collect_unique(capture_origins, "capture_origin", allow_empty=True)
    if len(output_roots) != 1:
        raise SystemExit(
            "unable to infer output_root for candidate record batch: "
            + ", ".join(sorted(make_repo_relative(path) for path in output_roots))
        )
    output_root = next(iter(output_roots))
    batch_key = ""
    if uses_short_candidate_record_layout(project):
        batch_key = infer_batch_key_from_output_root(output_root)
        if not batch_key:
            batch_key = derive_candidate_batch_key(project=project, collection_batch=collection_batch)

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
        "output_root": make_repo_relative(output_root),
        "records": manifests,
    }
    if batch_key:
        manifest["batch_key"] = batch_key
    if capture_origin:
        manifest["capture_origin"] = capture_origin
    if description.strip():
        manifest["description"] = description.strip()

    ordered_records.sort(key=lambda item: (item[0], item[1], make_repo_relative(item[2])))

    for sample_id, record_id, path, _record in ordered_records:
        entry: dict[str, str] = {
            "record_id": record_id,
            "sample_id": sample_id,
            "path": make_repo_relative(path),
        }
        if uses_short_candidate_record_layout(project):
            entry["sample_key"] = extract_sample_key_from_record_path(path)
            entry["record_relpath"] = make_manifest_entry_record_relpath(path, output_root)
        manifests.append(entry)

    ensure_schema(manifest, RECORD_BATCH_SCHEMA_PATH, "candidate record batch manifest")
    return manifest


def write_json_document(path: Path, document: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
