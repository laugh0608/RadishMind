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
    ensure_schema,
    load_json_document,
    make_repo_relative,
    resolve_relative_to_repo,
    write_json_document,
)

INDEX_SCHEMA_PATH = REPO_ROOT / "datasets/eval/negative-replay-index.schema.json"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Build an index that links audited batch failures to negative replay samples.",
    )
    parser.add_argument("--audit-report", required=True, help="Structured audit report json path.")
    parser.add_argument(
        "--negative-sample-dir",
        default="datasets/eval/radish-negative",
        help="Negative replay sample directory to scan. Default: datasets/eval/radish-negative",
    )
    parser.add_argument(
        "--output",
        default="",
        help="Output index json path. Defaults to <audit-report>.negative-replay-index.json",
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


def default_output_path(audit_report_path: Path) -> Path:
    name = audit_report_path.name
    if name.endswith(".audit.json"):
        return audit_report_path.with_name(name[: -len(".audit.json")] + ".negative-replay-index.json")
    if name.endswith(".json"):
        return audit_report_path.with_name(name[: -len(".json")] + ".negative-replay-index.json")
    return audit_report_path.with_name(name + ".negative-replay-index.json")


def load_manifest(manifest_path: Path) -> dict[str, Any]:
    manifest = expect_object(load_json_document(manifest_path), make_repo_relative(manifest_path))
    ensure_schema(manifest, RECORD_BATCH_SCHEMA_PATH, make_repo_relative(manifest_path))
    return manifest


def build_sample_file_map(sample_dir: Path) -> dict[str, dict[str, str]]:
    sample_file_map: dict[str, dict[str, str]] = {}
    sample_paths = sorted(sample_dir.glob("*.json"))
    if not sample_paths:
        raise SystemExit(f"no sample json files found in: {make_repo_relative(sample_dir)}")
    for sample_path in sample_paths:
        sample = expect_object(load_json_document(sample_path), make_repo_relative(sample_path))
        sample_id = expect_non_empty_string(sample.get("sample_id"), f"{make_repo_relative(sample_path)} sample_id")
        sample_file_map[sample_path.name] = {
            "sample_id": sample_id,
            "path": make_repo_relative(sample_path),
        }
    return sample_file_map


def build_manifest_record_map(manifest: dict[str, Any], manifest_label: str) -> dict[str, dict[str, str]]:
    record_map: dict[str, dict[str, str]] = {}
    for entry in manifest.get("records") or []:
        if not isinstance(entry, dict):
            raise SystemExit(f"{manifest_label}: each record entry must be a json object")
        record_id = expect_non_empty_string(entry.get("record_id"), f"{manifest_label} record_id")
        sample_id = expect_non_empty_string(entry.get("sample_id"), f"{manifest_label} sample_id")
        path_value = expect_non_empty_string(entry.get("path"), f"{manifest_label} path")
        if record_id in record_map:
            raise SystemExit(f"{manifest_label}: duplicate record_id found: {record_id}")
        record_map[record_id] = {
            "sample_id": sample_id,
            "path": path_value,
        }
    if not record_map:
        raise SystemExit(f"{manifest_label}: manifest does not contain any records")
    return record_map


def build_status_entries(
    audit_report: dict[str, Any],
    *,
    sample_file_map: dict[str, dict[str, str]],
    manifest_record_map: dict[str, dict[str, str]],
) -> tuple[dict[str, dict[str, Any]], dict[str, dict[str, Any]]]:
    records_by_sample_id = {entry["sample_id"]: record_id for record_id, entry in manifest_record_map.items()}
    failed_entries: dict[str, dict[str, Any]] = {}
    non_failed_entries: dict[str, dict[str, Any]] = {}

    for sample_entry in audit_report.get("samples") or []:
        if not isinstance(sample_entry, dict):
            raise SystemExit("audit report samples entries must be json objects")
        sample_file = expect_non_empty_string(sample_entry.get("sample_file"), "audit report sample_file")
        status = expect_non_empty_string(sample_entry.get("status"), f"audit report status for {sample_file}")
        sample_info = sample_file_map.get(sample_file)
        if sample_info is None:
            raise SystemExit(f"audit report sample file not found in audit sample dir: {sample_file}")
        source_sample_id = sample_info["sample_id"]
        record_id = records_by_sample_id.get(source_sample_id)
        if record_id is None:
            raise SystemExit(f"audit report sample_id is missing from manifest records: {source_sample_id}")

        entry = {
            "source_sample_id": source_sample_id,
            "source_sample_file": sample_file,
            "record_id": record_id,
            "audit_violations": unique_strings(list(sample_entry.get("violations") or [])),
        }

        if status == "fail":
            if not entry["audit_violations"]:
                raise SystemExit(f"audit report failed sample is missing violations: {sample_file}")
            failed_entries[record_id] = entry
        else:
            non_failed_entries[record_id] = entry

    return failed_entries, non_failed_entries


def build_negative_replay_entry(
    sample: dict[str, Any],
    sample_path: Path,
    source_entry: dict[str, Any],
) -> dict[str, Any]:
    negative_sample_id = expect_non_empty_string(sample.get("sample_id"), f"{make_repo_relative(sample_path)} sample_id")
    expectations = sample.get("negative_replay_expectations") or {}
    if not isinstance(expectations, dict):
        raise SystemExit(f"{make_repo_relative(sample_path)} negative_replay_expectations must be a json object")
    expected_candidate_violations = unique_strings(list(expectations.get("expected_candidate_violations") or []))
    if not expected_candidate_violations:
        raise SystemExit(
            f"{make_repo_relative(sample_path)} expected_candidate_violations is required for negative replay indexing"
        )

    return {
        "negative_sample_id": negative_sample_id,
        "negative_sample_path": make_repo_relative(sample_path),
        "source_sample_id": source_entry["source_sample_id"],
        "source_sample_file": source_entry["source_sample_file"],
        "record_id": source_entry["record_id"],
        "replay_mode": "same_sample" if negative_sample_id == source_entry["source_sample_id"] else "cross_sample",
        "audit_violations": source_entry["audit_violations"],
        "expected_candidate_violations": expected_candidate_violations,
    }


def build_index_document(
    audit_report: dict[str, Any],
    *,
    audit_report_path: Path,
    audit_sample_dir: Path,
    negative_sample_dir: Path,
) -> dict[str, Any]:
    manifest_path = resolve_relative_to_repo(expect_non_empty_string(audit_report.get("manifest_path"), "manifest_path"))
    manifest = load_manifest(manifest_path)
    manifest_label = make_repo_relative(manifest_path)
    manifest_record_map = build_manifest_record_map(manifest, manifest_label)
    sample_file_map = build_sample_file_map(audit_sample_dir)
    failed_entries, non_failed_entries = build_status_entries(
        audit_report,
        sample_file_map=sample_file_map,
        manifest_record_map=manifest_record_map,
    )

    target_manifest_path = make_repo_relative(manifest_path)
    linked_failed_entries: list[dict[str, Any]] = []
    linked_non_failed_entries: list[dict[str, Any]] = []

    for negative_sample_path in sorted(negative_sample_dir.glob("*.json")):
        sample = expect_object(load_json_document(negative_sample_path), make_repo_relative(negative_sample_path))
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

        if record_id in failed_entries:
            linked_failed_entries.append(build_negative_replay_entry(sample, negative_sample_path, failed_entries[record_id]))
            continue
        if record_id in non_failed_entries:
            linked_non_failed_entries.append(
                build_negative_replay_entry(sample, negative_sample_path, non_failed_entries[record_id])
            )
            continue
        raise SystemExit(
            f"{make_repo_relative(negative_sample_path)} references record_id not present in audit report: {record_id}"
        )

    linked_failed_record_ids = {entry["record_id"] for entry in linked_failed_entries}
    unlinked_failed_samples = [
        failed_entries[record_id]
        for record_id in sorted(failed_entries)
        if record_id not in linked_failed_record_ids
    ]

    grouped_entries: dict[tuple[str, ...], list[dict[str, Any]]] = {}
    for entry in linked_failed_entries:
        group_key = tuple(entry["expected_candidate_violations"])
        grouped_entries.setdefault(group_key, []).append(entry)

    violation_groups: list[dict[str, Any]] = []
    for index, group_key in enumerate(sorted(grouped_entries), start=1):
        entries = sorted(
            grouped_entries[group_key],
            key=lambda item: (item["source_sample_file"], item["negative_sample_path"]),
        )
        audit_violation_fragments = unique_strings(
            [fragment for entry in entries for fragment in entry["audit_violations"]]
        )
        violation_groups.append(
            {
                "group_id": f"group-{index:03d}",
                "expected_candidate_violations": list(group_key),
                "audit_violation_fragments": audit_violation_fragments,
                "entry_count": len(entries),
                "entries": entries,
            }
        )

    document: dict[str, Any] = {
        "schema_version": 1,
        "eval_task": expect_non_empty_string(audit_report.get("task"), "audit report task"),
        "project": expect_non_empty_string(manifest.get("project"), "manifest project"),
        "task": expect_non_empty_string(manifest.get("task"), "manifest task"),
        "source_manifest_path": target_manifest_path,
        "audit_report_path": make_repo_relative(audit_report_path),
        "audit_sample_dir": make_repo_relative(audit_sample_dir),
        "negative_sample_dir": make_repo_relative(negative_sample_dir),
        "collection_batch": expect_non_empty_string(manifest.get("collection_batch"), "manifest collection_batch"),
        "source": expect_non_empty_string(manifest.get("source"), "manifest source"),
        "summary": {
            "audited_sample_count": len(list(audit_report.get("samples") or [])),
            "failed_sample_count": len(failed_entries),
            "linked_negative_sample_count": len(linked_failed_entries) + len(linked_non_failed_entries),
            "linked_failed_sample_count": len(linked_failed_entries),
            "linked_non_failed_sample_count": len(linked_non_failed_entries),
            "unlinked_failed_sample_count": len(unlinked_failed_samples),
            "violation_group_count": len(violation_groups),
        },
        "violation_groups": violation_groups,
        "linked_non_failed_samples": sorted(
            linked_non_failed_entries,
            key=lambda item: (item["source_sample_file"], item["negative_sample_path"]),
        ),
        "unlinked_failed_samples": unlinked_failed_samples,
    }

    capture_origin = str(manifest.get("capture_origin") or "").strip()
    if capture_origin:
        document["capture_origin"] = capture_origin

    ensure_schema(document, INDEX_SCHEMA_PATH, "negative replay index")
    return document


def main() -> int:
    args = parse_args()
    audit_report_path = resolve_relative_to_repo(args.audit_report)
    if not audit_report_path.is_file():
        raise SystemExit(f"audit report file not found: {args.audit_report}")
    audit_report = expect_object(load_json_document(audit_report_path), make_repo_relative(audit_report_path))

    audit_sample_dir = resolve_relative_to_repo(expect_non_empty_string(audit_report.get("sample_dir"), "sample_dir"))
    if not audit_sample_dir.is_dir():
        raise SystemExit(f"audit sample directory not found: {make_repo_relative(audit_sample_dir)}")

    negative_sample_dir = resolve_relative_to_repo(args.negative_sample_dir)
    if not negative_sample_dir.is_dir():
        raise SystemExit(f"negative sample directory not found: {args.negative_sample_dir}")

    output_path = resolve_relative_to_repo(args.output) if args.output else default_output_path(audit_report_path)
    document = build_index_document(
        audit_report,
        audit_report_path=audit_report_path,
        audit_sample_dir=audit_sample_dir,
        negative_sample_dir=negative_sample_dir,
    )

    if args.check:
        if not output_path.is_file():
            raise SystemExit(f"negative replay index file not found for --check: {make_repo_relative(output_path)}")
        existing = load_json_document(output_path)
        if existing != document:
            raise SystemExit(f"negative replay index is out of date: {make_repo_relative(output_path)}")
        print(f"negative replay index is up to date: {make_repo_relative(output_path)}", file=sys.stdout)
        return 0

    write_json_document(output_path, document)
    print(
        f"wrote negative replay index with {document['summary']['linked_failed_sample_count']} linked failed sample(s) to {make_repo_relative(output_path)}",
        file=sys.stdout,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
