#!/usr/bin/env python3
from __future__ import annotations

import argparse
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import (  # noqa: E402
    RECORD_BATCH_SCHEMA_PATH,
    RECORD_SCHEMA_PATH,
    ensure_schema,
    load_json_document,
    make_repo_relative,
    resolve_relative_to_repo,
    write_json_document,
)

INDEX_SCHEMA_PATH = REPO_ROOT / "datasets/eval/real-derived-negative-index.schema.json"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Build an index for negative samples derived from real candidate response records.",
    )
    parser.add_argument("--manifest", required=True, help="Derived candidate record batch manifest json path.")
    parser.add_argument(
        "--negative-sample-dir",
        default="datasets/eval/radish-negative",
        help="Negative sample directory to scan. Default: datasets/eval/radish-negative",
    )
    parser.add_argument(
        "--output",
        default="",
        help="Output index json path. Defaults to <manifest>.real-derived-index.json",
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="Do not write files; fail if the generated document differs from --output.",
    )
    return parser.parse_args()


def expect_object(document: Any, label: str) -> dict[str, Any]:
    if not isinstance(document, dict):
        raise SystemExit(f"{label} must be a json object")
    return document


def expect_non_empty_string(value: Any, label: str) -> str:
    normalized = str(value or "").strip()
    if not normalized:
        raise SystemExit(f"{label} is required")
    return normalized


def unique_strings(values: list[str]) -> list[str]:
    seen: set[str] = set()
    result: list[str] = []
    for value in values:
        normalized = str(value or "").strip()
        if normalized and normalized not in seen:
            seen.add(normalized)
            result.append(normalized)
    return result


def default_output_path(manifest_path: Path) -> Path:
    name = manifest_path.name
    if name.endswith(".manifest.json"):
        return manifest_path.with_name(name[: -len(".manifest.json")] + ".real-derived-index.json")
    if name.endswith(".json"):
        return manifest_path.with_name(name[: -len(".json")] + ".real-derived-index.json")
    return manifest_path.with_name(name + ".real-derived-index.json")


def load_manifest(manifest_path: Path) -> dict[str, Any]:
    manifest = expect_object(load_json_document(manifest_path), make_repo_relative(manifest_path))
    ensure_schema(manifest, RECORD_BATCH_SCHEMA_PATH, make_repo_relative(manifest_path))
    return manifest


def load_record(record_path: Path) -> dict[str, Any]:
    record = expect_object(load_json_document(record_path), make_repo_relative(record_path))
    ensure_schema(record, RECORD_SCHEMA_PATH, make_repo_relative(record_path))
    return record


def build_manifest_record_map(manifest: dict[str, Any], manifest_label: str) -> dict[str, dict[str, str]]:
    record_map: dict[str, dict[str, str]] = {}
    for entry in manifest.get("records") or []:
        entry_object = expect_object(entry, f"{manifest_label} record entry")
        record_id = expect_non_empty_string(entry_object.get("record_id"), f"{manifest_label} record_id")
        sample_id = expect_non_empty_string(entry_object.get("sample_id"), f"{manifest_label} sample_id")
        path_value = expect_non_empty_string(entry_object.get("path"), f"{manifest_label} path")
        if record_id in record_map:
            raise SystemExit(f"{manifest_label}: duplicate record_id found: {record_id}")
        record_map[record_id] = {
            "sample_id": sample_id,
            "path": path_value,
        }
    if not record_map:
        raise SystemExit(f"{manifest_label}: manifest does not contain any records")
    return record_map


def extract_expected_candidate_violations(sample: dict[str, Any], sample_path: Path) -> list[str]:
    expectations = sample.get("negative_replay_expectations") or {}
    expectations_object = expect_object(
        expectations,
        f"{make_repo_relative(sample_path)} negative_replay_expectations",
    )
    violations = unique_strings(list(expectations_object.get("expected_candidate_violations") or []))
    if not violations:
        raise SystemExit(
            f"{make_repo_relative(sample_path)} expected_candidate_violations is required for real-derived indexing"
        )
    return violations


def build_negative_sample_map(target_manifest_path: str, negative_sample_dir: Path) -> dict[str, list[dict[str, Any]]]:
    sample_map: dict[str, list[dict[str, Any]]] = {}
    sample_paths = sorted(negative_sample_dir.glob("*.json"))
    if not sample_paths:
        raise SystemExit(f"no negative sample json files found in: {make_repo_relative(negative_sample_dir)}")

    for sample_path in sample_paths:
        sample = expect_object(load_json_document(sample_path), make_repo_relative(sample_path))
        record_ref = sample.get("candidate_response_record")
        if not isinstance(record_ref, dict):
            continue
        manifest_path_value = str(record_ref.get("manifest_path") or "").strip()
        record_id = str(record_ref.get("record_id") or "").strip()
        if not manifest_path_value or not record_id:
            continue
        normalized_manifest_path = make_repo_relative(resolve_relative_to_repo(manifest_path_value))
        if normalized_manifest_path != target_manifest_path:
            continue

        sample_map.setdefault(record_id, []).append(
            {
                "negative_sample_id": expect_non_empty_string(
                    sample.get("sample_id"), f"{make_repo_relative(sample_path)} sample_id"
                ),
                "negative_sample_path": make_repo_relative(sample_path),
                "expected_candidate_violations": extract_expected_candidate_violations(sample, sample_path),
            }
        )

    return sample_map


def build_source_record_info(source_manifest_path: Path, source_record_id: str) -> dict[str, str]:
    source_manifest = load_manifest(source_manifest_path)
    source_manifest_label = make_repo_relative(source_manifest_path)
    source_record_map = build_manifest_record_map(source_manifest, source_manifest_label)
    source_entry = source_record_map.get(source_record_id)
    if source_entry is None:
        raise SystemExit(f"{source_manifest_label}: missing source record_id '{source_record_id}'")

    source_record_path = resolve_relative_to_repo(source_entry["path"])
    source_record = load_record(source_record_path)

    result = {
        "source_manifest_path": source_manifest_label,
        "source_collection_batch": expect_non_empty_string(
            source_manifest.get("collection_batch"),
            f"{source_manifest_label} collection_batch",
        ),
        "source_record_id": source_record_id,
        "source_sample_id": expect_non_empty_string(source_entry.get("sample_id"), f"{source_manifest_label} sample_id"),
        "source_record_path": make_repo_relative(source_record_path),
        "source_source": expect_non_empty_string(source_manifest.get("source"), f"{source_manifest_label} source"),
    }
    source_capture_origin = str((source_manifest.get("capture_origin") or "")).strip()
    if source_capture_origin:
        result["source_capture_origin"] = source_capture_origin

    if result["source_source"] != "captured_candidate_response":
        raise SystemExit(
            f"{source_manifest_label}: source manifest for real-derived negatives must be captured_candidate_response"
        )

    if expect_non_empty_string(source_record.get("record_id"), f"{result['source_record_path']} record_id") != source_record_id:
        raise SystemExit(f"{result['source_record_path']}: record_id does not match source manifest entry")

    return result


def build_index_document(
    manifest: dict[str, Any],
    *,
    manifest_path: Path,
    negative_sample_dir: Path,
) -> dict[str, Any]:
    manifest_label = make_repo_relative(manifest_path)
    manifest_record_map = build_manifest_record_map(manifest, manifest_label)
    target_manifest_path = make_repo_relative(manifest_path)
    negative_sample_map = build_negative_sample_map(target_manifest_path, negative_sample_dir)

    linked_entries: list[dict[str, Any]] = []
    unlinked_derived_records: list[dict[str, Any]] = []
    derived_record_count = 0
    source_manifest_keys: set[str] = set()
    source_record_keys: set[tuple[str, str]] = set()

    for record_id, manifest_entry in sorted(manifest_record_map.items()):
        record_path = resolve_relative_to_repo(manifest_entry["path"])
        record = load_record(record_path)

        capture_metadata = record.get("capture_metadata") or {}
        capture_metadata_object = expect_object(capture_metadata, f"{make_repo_relative(record_path)} capture_metadata")
        tags = {str(tag).strip() for tag in capture_metadata_object.get("tags") or [] if str(tag).strip()}
        source_record_ref = capture_metadata_object.get("source_candidate_response_record")

        if source_record_ref is None and "real_record_derived" not in tags:
            continue

        if "real_record_derived" not in tags:
            raise SystemExit(
                f"{make_repo_relative(record_path)}: source_candidate_response_record requires capture_metadata.tags to contain 'real_record_derived'"
            )
        source_record_ref_object = expect_object(
            source_record_ref,
            f"{make_repo_relative(record_path)} capture_metadata.source_candidate_response_record",
        )
        source_manifest_path = resolve_relative_to_repo(
            expect_non_empty_string(
                source_record_ref_object.get("manifest_path"),
                f"{make_repo_relative(record_path)} source manifest_path",
            )
        )
        source_record_id = expect_non_empty_string(
            source_record_ref_object.get("record_id"),
            f"{make_repo_relative(record_path)} source record_id",
        )
        source_info = build_source_record_info(source_manifest_path, source_record_id)
        source_manifest_keys.add(source_info["source_manifest_path"])
        source_record_keys.add((source_info["source_manifest_path"], source_info["source_record_id"]))
        derived_record_count += 1

        base_entry = {
            "derived_record_id": expect_non_empty_string(record.get("record_id"), f"{make_repo_relative(record_path)} record_id"),
            "derived_sample_id": expect_non_empty_string(record.get("sample_id"), f"{make_repo_relative(record_path)} sample_id"),
            "derived_record_path": make_repo_relative(record_path),
            **source_info,
        }

        sample_entries = sorted(
            negative_sample_map.get(record_id, []),
            key=lambda item: (item["negative_sample_id"], item["negative_sample_path"]),
        )
        if not sample_entries:
            unlinked_derived_records.append(base_entry)
            continue

        for sample_entry in sample_entries:
            linked_entries.append(
                {
                    **base_entry,
                    "negative_sample_id": sample_entry["negative_sample_id"],
                    "negative_sample_path": sample_entry["negative_sample_path"],
                    "expected_candidate_violations": sample_entry["expected_candidate_violations"],
                }
            )

    source_record_groups: list[dict[str, Any]] = []
    grouped_by_source: dict[tuple[str, str], list[dict[str, Any]]] = {}
    for entry in linked_entries:
        group_key = (entry["source_manifest_path"], entry["source_record_id"])
        grouped_by_source.setdefault(group_key, []).append(entry)
    for index, group_key in enumerate(sorted(grouped_by_source), start=1):
        entries = sorted(
            grouped_by_source[group_key],
            key=lambda item: (item["negative_sample_id"], item["negative_sample_path"]),
        )
        first_entry = entries[0]
        group = {
            "group_id": f"source-{index:03d}",
            "source_manifest_path": first_entry["source_manifest_path"],
            "source_collection_batch": first_entry["source_collection_batch"],
            "source_record_id": first_entry["source_record_id"],
            "source_sample_id": first_entry["source_sample_id"],
            "source_record_path": first_entry["source_record_path"],
            "source_source": first_entry["source_source"],
            "entry_count": len(entries),
            "entries": entries,
        }
        if first_entry.get("source_capture_origin"):
            group["source_capture_origin"] = first_entry["source_capture_origin"]
        source_record_groups.append(group)

    violation_groups: list[dict[str, Any]] = []
    grouped_by_violation: dict[tuple[str, ...], list[dict[str, Any]]] = {}
    for entry in linked_entries:
        group_key = tuple(entry["expected_candidate_violations"])
        grouped_by_violation.setdefault(group_key, []).append(entry)
    for index, group_key in enumerate(sorted(grouped_by_violation), start=1):
        entries = sorted(
            grouped_by_violation[group_key],
            key=lambda item: (item["source_sample_id"], item["negative_sample_path"]),
        )
        violation_groups.append(
            {
                "group_id": f"group-{index:03d}",
                "expected_candidate_violations": list(group_key),
                "entry_count": len(entries),
                "entries": entries,
            }
        )

    document: dict[str, Any] = {
        "schema_version": 1,
        "project": expect_non_empty_string(manifest.get("project"), f"{manifest_label} project"),
        "task": expect_non_empty_string(manifest.get("task"), f"{manifest_label} task"),
        "record_manifest_path": manifest_label,
        "negative_sample_dir": make_repo_relative(negative_sample_dir),
        "collection_batch": expect_non_empty_string(manifest.get("collection_batch"), f"{manifest_label} collection_batch"),
        "source": expect_non_empty_string(manifest.get("source"), f"{manifest_label} source"),
        "summary": {
            "derived_record_count": derived_record_count,
            "linked_negative_sample_count": len(linked_entries),
            "source_manifest_count": len(source_manifest_keys),
            "source_record_count": len(source_record_keys),
            "source_record_group_count": len(source_record_groups),
            "violation_group_count": len(violation_groups),
            "unlinked_derived_record_count": len(unlinked_derived_records),
        },
        "source_record_groups": source_record_groups,
        "violation_groups": violation_groups,
        "unlinked_derived_records": sorted(
            unlinked_derived_records,
            key=lambda item: (item["source_manifest_path"], item["source_record_id"], item["derived_record_path"]),
        ),
    }
    capture_origin = str(manifest.get("capture_origin") or "").strip()
    if capture_origin:
        document["capture_origin"] = capture_origin

    ensure_schema(document, INDEX_SCHEMA_PATH, "real-derived negative index")
    return document


def main() -> int:
    args = parse_args()
    manifest_path = resolve_relative_to_repo(args.manifest)
    if not manifest_path.is_file():
        raise SystemExit(f"manifest file not found: {args.manifest}")
    manifest = load_manifest(manifest_path)

    negative_sample_dir = resolve_relative_to_repo(args.negative_sample_dir)
    if not negative_sample_dir.is_dir():
        raise SystemExit(f"negative sample directory not found: {args.negative_sample_dir}")

    output_path = resolve_relative_to_repo(args.output) if args.output else default_output_path(manifest_path)
    document = build_index_document(
        manifest,
        manifest_path=manifest_path,
        negative_sample_dir=negative_sample_dir,
    )

    if args.check:
        if not output_path.is_file():
            raise SystemExit(f"real-derived negative index file not found for --check: {make_repo_relative(output_path)}")
        existing = load_json_document(output_path)
        if existing != document:
            raise SystemExit(f"real-derived negative index is out of date: {make_repo_relative(output_path)}")
        print(f"real-derived negative index is up to date: {make_repo_relative(output_path)}", file=sys.stdout)
        return 0

    write_json_document(output_path, document)
    print(
        f"wrote real-derived negative index with {document['summary']['linked_negative_sample_count']} linked sample(s) to {make_repo_relative(output_path)}",
        file=sys.stdout,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
