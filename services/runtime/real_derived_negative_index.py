from __future__ import annotations

from pathlib import Path
from typing import Any

from .candidate_records import load_json_document, make_repo_relative, resolve_relative_to_repo, write_json_document


GROUP_PART_FILE_NAMES = {
    "source_record_groups": "source-record-groups.json",
    "violation_groups": "violation-groups.json",
    "pattern_groups": "pattern-groups.json",
    "unlinked_derived_records": "unlinked-derived-records.json",
}
RADISHFLOW_SHORT_GROUP_PART_FILE_NAMES = {
    "source_record_groups": "src.json",
    "violation_groups": "viol.json",
    "pattern_groups": "pat.json",
    "unlinked_derived_records": "unlinked.json",
}


def expect_object(document: Any, label: str) -> dict[str, Any]:
    if not isinstance(document, dict):
        raise SystemExit(f"{label} must be a json object")
    return document


def uses_radishflow_short_parts(index_path: Path) -> bool:
    normalized_parts = index_path.as_posix().split("/")
    target_parts = ["datasets", "eval", "candidate-records", "radishflow"]
    for start in range(0, max(0, len(normalized_parts) - len(target_parts) + 1)):
        if normalized_parts[start : start + len(target_parts)] == target_parts:
            return True
    return False


def resolve_parts_dir(index_path: Path) -> Path:
    if uses_radishflow_short_parts(index_path):
        return index_path.with_name("rdi.parts")
    suffix = ".json"
    if index_path.name.endswith(suffix):
        parts_name = index_path.name[: -len(suffix)] + ".parts"
    else:
        parts_name = index_path.name + ".parts"
    return index_path.with_name(parts_name)


def expand_real_derived_negative_index(document: Any, *, index_path: Path, label: str = "") -> dict[str, Any]:
    index_label = label or make_repo_relative(index_path)
    expanded = dict(expect_object(document, index_label))
    for key in GROUP_PART_FILE_NAMES:
        value = expanded.get(key)
        path_key = f"{key}_path"
        if isinstance(value, list):
            expanded.pop(path_key, None)
            continue
        part_path_value = str(expanded.get(path_key) or "").strip()
        if not part_path_value:
            raise SystemExit(f"{index_label}: missing '{key}' array or '{path_key}' reference")
        part_path = resolve_relative_to_repo(part_path_value)
        part_document = load_json_document(part_path)
        if not isinstance(part_document, list):
            raise SystemExit(f"{make_repo_relative(part_path)} must be a json array")
        expanded[key] = part_document
        expanded.pop(path_key, None)
    return expanded


def load_real_derived_negative_index(index_path: Path, *, label: str = "") -> dict[str, Any]:
    document = load_json_document(index_path)
    return expand_real_derived_negative_index(document, index_path=index_path, label=label)


def build_compact_real_derived_negative_index(document: dict[str, Any], *, output_path: Path) -> tuple[dict[str, Any], dict[str, Any]]:
    compact = {
        key: value
        for key, value in document.items()
        if key not in GROUP_PART_FILE_NAMES
    }
    part_documents: dict[str, Any] = {}
    parts_dir = resolve_parts_dir(output_path)
    part_file_names = (
        RADISHFLOW_SHORT_GROUP_PART_FILE_NAMES if uses_radishflow_short_parts(output_path) else GROUP_PART_FILE_NAMES
    )
    for key, filename in part_file_names.items():
        part_path = parts_dir / filename
        compact[f"{key}_path"] = make_repo_relative(part_path)
        part_documents[filename] = document.get(key) or []
    return compact, part_documents


def write_real_derived_negative_index(output_path: Path, document: dict[str, Any]) -> None:
    compact, part_documents = build_compact_real_derived_negative_index(document, output_path=output_path)
    parts_dir = resolve_parts_dir(output_path)
    for filename, part_document in part_documents.items():
        write_json_document(parts_dir / filename, part_document)
    write_json_document(output_path, compact)
