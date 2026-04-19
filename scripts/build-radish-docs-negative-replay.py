#!/usr/bin/env python3
from __future__ import annotations

import argparse
import copy
import json
import re
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
    resolve_manifest_record_path,
    write_json_document,
)

NEGATIVE_REPLAY_INDEX_SCHEMA_PATH = REPO_ROOT / "datasets/eval/negative-replay-index.schema.json"
RADISH_SAMPLE_SCHEMA_PATH = REPO_ROOT / "datasets/eval/radish-task-sample.schema.json"
STABLE_CAPTURE_TAGS = {"real_capture"}
REPLAY_MODES = {"same_sample", "cross_sample", "all"}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Build Radish docs QA negative replay fixtures from a negative replay index.",
    )
    parser.add_argument("--index", required=True, help="Negative replay index json path.")
    parser.add_argument(
        "--output-dir",
        default="",
        help="Optional output directory override. Defaults to negative_sample_dir from the index.",
    )
    parser.add_argument("--group-id", action="append", default=[], help="Optional group_id filter. Can be repeated.")
    parser.add_argument("--record-id", action="append", default=[], help="Optional record_id filter. Can be repeated.")
    parser.add_argument(
        "--replay-mode",
        default="same_sample",
        help="Replay mode filter: same_sample, cross_sample, or all. Default: same_sample",
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="Do not write files; fail if generated replay fixtures differ from on-disk files.",
    )
    return parser.parse_args()


def expect_object(document: Any, label: str) -> dict[str, Any]:
    if not isinstance(document, dict):
        raise SystemExit(f"{label} must be a json object")
    return document


def expect_string(value: Any, label: str) -> str:
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


def load_index(index_path: Path) -> dict[str, Any]:
    index_document = expect_object(load_json_document(index_path), make_repo_relative(index_path))
    ensure_schema(index_document, NEGATIVE_REPLAY_INDEX_SCHEMA_PATH, make_repo_relative(index_path))
    return index_document


def load_manifest(manifest_path: Path) -> dict[str, Any]:
    manifest_document = expect_object(load_json_document(manifest_path), make_repo_relative(manifest_path))
    ensure_schema(manifest_document, RECORD_BATCH_SCHEMA_PATH, make_repo_relative(manifest_path))
    return manifest_document


def load_record(record_path: Path) -> dict[str, Any]:
    record_document = expect_object(load_json_document(record_path), make_repo_relative(record_path))
    ensure_schema(record_document, RECORD_SCHEMA_PATH, make_repo_relative(record_path))
    return record_document


def normalize_violation_fragment(violation: str) -> str:
    normalized = str(violation or "").strip()
    normalized = re.sub(r"^[^:]+:(?:candidate_response:)?\s*", "", normalized)
    if normalized.startswith("candidate_response."):
        normalized = normalized[len("candidate_response") :]
    elif normalized.startswith("candidate_response "):
        normalized = normalized[len("candidate_response ") :]
    return normalized.strip()


def derive_expected_candidate_violations(audit_violations: list[str]) -> list[str]:
    normalized = unique_strings([normalize_violation_fragment(violation) for violation in audit_violations])
    filtered: list[str] = []
    has_missing_action = any(fragment.startswith("is missing required action kind ") for fragment in normalized)
    has_missing_issue = "must contain at least 1 issue" in normalized
    has_risk_mismatch = ".risk_level does not match evaluation.expected_risk_level" in normalized

    for fragment in normalized:
        if has_missing_action and fragment.startswith("missing required json path '$.proposed_actions["):
            continue
        if has_missing_issue and fragment == "missing required json path '$.issues[0]'":
            continue
        if has_risk_mismatch and fragment.startswith("json path '$.risk_level' expected "):
            continue
        filtered.append(fragment)
    return filtered


def build_record_map(manifest: dict[str, Any], manifest_path: Path) -> dict[str, Path]:
    record_map: dict[str, Path] = {}
    for entry in manifest.get("records") or []:
        if not isinstance(entry, dict):
            raise SystemExit("manifest records entries must be json objects")
        record_id = expect_string(entry.get("record_id"), "manifest record_id")
        record_map[record_id] = resolve_manifest_record_path(manifest_path, manifest, entry)
    return record_map


def build_index_entries(index_document: dict[str, Any]) -> list[dict[str, Any]]:
    entries: list[dict[str, Any]] = []
    for group in index_document.get("violation_groups") or []:
        if not isinstance(group, dict):
            raise SystemExit("negative replay index violation_groups entries must be json objects")
        group_id = expect_string(group.get("group_id"), "negative replay index group_id")
        for entry in group.get("entries") or []:
            if not isinstance(entry, dict):
                raise SystemExit(f"negative replay index group '{group_id}' entries must be json objects")
            enriched = copy.deepcopy(entry)
            enriched["group_id"] = group_id
            entries.append(enriched)

    for entry in index_document.get("unlinked_failed_samples") or []:
        if not isinstance(entry, dict):
            raise SystemExit("negative replay index unlinked_failed_samples entries must be json objects")
        enriched = copy.deepcopy(entry)
        enriched["group_id"] = ""
        enriched["replay_mode"] = "same_sample"
        enriched["negative_sample_id"] = expect_string(entry.get("source_sample_id"), "unlinked source_sample_id")
        enriched["expected_candidate_violations"] = derive_expected_candidate_violations(
            [str(value) for value in entry.get("audit_violations") or []]
        )
        entries.append(enriched)

    for entry in index_document.get("linked_non_failed_samples") or []:
        if not isinstance(entry, dict):
            raise SystemExit("negative replay index linked_non_failed_samples entries must be json objects")
        enriched = copy.deepcopy(entry)
        enriched["group_id"] = ""
        entries.append(enriched)
    return entries


def filter_entries(
    entries: list[dict[str, Any]],
    group_ids: list[str],
    record_ids: list[str],
    replay_mode: str,
) -> list[dict[str, Any]]:
    normalized_replay_mode = str(replay_mode or "").strip() or "same_sample"
    if normalized_replay_mode not in REPLAY_MODES:
        raise SystemExit(
            f"unsupported replay_mode '{normalized_replay_mode}', expected one of: {', '.join(sorted(REPLAY_MODES))}"
        )

    if normalized_replay_mode == "all":
        filtered = list(entries)
    else:
        filtered = [entry for entry in entries if str(entry.get("replay_mode") or "").strip() == normalized_replay_mode]

    requested_group_ids = [value for value in group_ids if value]
    if requested_group_ids:
        known_group_ids = {str(entry.get("group_id") or "").strip() for entry in filtered}
        missing_group_ids = [group_id for group_id in requested_group_ids if group_id not in known_group_ids]
        if missing_group_ids:
            raise SystemExit(f"unknown group_id value(s): {', '.join(sorted(missing_group_ids))}")
        filtered = [entry for entry in filtered if str(entry.get("group_id") or "").strip() in requested_group_ids]

    requested_record_ids = {value for value in record_ids if value}
    if requested_record_ids:
        filtered = [entry for entry in filtered if str(entry.get("record_id") or "").strip() in requested_record_ids]
        if not filtered:
            raise SystemExit(f"no {normalized_replay_mode} replay entries matched the requested record_id filters")

    if not filtered:
        raise SystemExit(f"no {normalized_replay_mode} replay entries matched the requested filters")
    return filtered


def build_stable_required_tags(record: dict[str, Any]) -> list[str]:
    capture_metadata = record.get("capture_metadata") or {}
    tags = unique_strings([str(value) for value in capture_metadata.get("tags") or []])
    stable_tags = [tag for tag in tags if tag in STABLE_CAPTURE_TAGS or tag.startswith("radish_docs_qa")]
    return stable_tags


def build_violation_slug(record: dict[str, Any], expected_candidate_violations: list[str]) -> str:
    response = expect_object(record.get("response"), f"{record.get('record_id')} response")
    slug_parts: list[str] = []
    if ".status does not match expected_response_shape.status" in expected_candidate_violations:
        actual_status = str(response.get("status") or "").strip()
        if actual_status:
            slug_parts.append(actual_status)
    if ".risk_level does not match evaluation.expected_risk_level" in expected_candidate_violations:
        actual_risk = str(response.get("risk_level") or "").strip()
        if actual_risk:
            slug_parts.append(f"{actual_risk}-risk")
    if "must contain at least 1 issue" in expected_candidate_violations:
        issues = response.get("issues") or []
        if isinstance(issues, list) and len(issues) == 0:
            slug_parts.append("no-issue")
    if any(fragment.startswith("is missing required action kind ") for fragment in expected_candidate_violations):
        for fragment in expected_candidate_violations:
            match = re.search(r"'([^']+)'", fragment)
            if match:
                slug_parts.append(f"missing-{match.group(1).replace('_', '-')}")
                break

    slug_parts = unique_strings(slug_parts)
    if slug_parts:
        return "-".join(slug_parts)
    return "replay-violation"


def derive_output_path(
    source_sample_file: str,
    record: dict[str, Any],
    expected_candidate_violations: list[str],
    output_dir: Path,
) -> Path:
    source_stem = Path(source_sample_file).stem
    match = re.match(r"^(.*)-(\d{3})$", source_stem)
    if match:
        sample_body = match.group(1)
        sample_number = match.group(2)
    else:
        sample_body = source_stem
        sample_number = "001"

    if sample_body.startswith("answer-docs-question-"):
        sample_body = sample_body[len("answer-docs-question-") :]

    violation_slug = build_violation_slug(record, expected_candidate_violations)
    filename = f"answer-docs-question-negative-real-record-{sample_body}-{violation_slug}-{sample_number}.json"
    return output_dir / filename


def build_note(collection_batch: str, expected_candidate_violations: list[str], replay_mode: str) -> str:
    joined = "；".join(f"`{fragment}`" for fragment in expected_candidate_violations)
    if replay_mode == "cross_sample":
        return (
            f"使用批次 `{collection_batch}` 中另一条真实候选快照做跨样本 replay，"
            f"将其回放到当前样本并验证命中这些违规：{joined}。"
        )
    return (
        f"使用批次 `{collection_batch}` 中同样本的真实候选快照做负例 replay，"
        f"保持原样本的 sample_id/request_id，对齐通过后验证真实候选回答命中这些违规：{joined}。"
    )


def load_template_sample(
    *,
    entry: dict[str, Any],
    audit_sample_dir: Path,
) -> dict[str, Any]:
    replay_mode = expect_string(entry.get("replay_mode"), "entry.replay_mode")
    if replay_mode == "same_sample":
        source_sample_file = expect_string(entry.get("source_sample_file"), "entry.source_sample_file")
        source_sample_path = audit_sample_dir / source_sample_file
        if not source_sample_path.is_file():
            raise SystemExit(f"source sample file not found: {make_repo_relative(source_sample_path)}")
        return expect_object(load_json_document(source_sample_path), make_repo_relative(source_sample_path))

    negative_sample_path_value = expect_string(entry.get("negative_sample_path"), "entry.negative_sample_path")
    negative_sample_path = resolve_relative_to_repo(negative_sample_path_value)
    if not negative_sample_path.is_file():
        raise SystemExit(
            "cross-sample replay generation requires an existing negative fixture template: "
            f"{make_repo_relative(negative_sample_path)}"
        )
    return expect_object(load_json_document(negative_sample_path), make_repo_relative(negative_sample_path))


def build_negative_sample(
    source_sample: dict[str, Any],
    *,
    index_document: dict[str, Any],
    entry: dict[str, Any],
    record: dict[str, Any],
) -> dict[str, Any]:
    base_sample = copy.deepcopy(source_sample)
    base_sample.pop("candidate_response", None)
    base_sample.pop("candidate_response_record", None)
    base_sample.pop("negative_replay_expectations", None)

    candidate_response_record: dict[str, Any] = {
        "manifest_path": expect_string(index_document.get("source_manifest_path"), "source_manifest_path"),
        "record_id": expect_string(entry.get("record_id"), "entry.record_id"),
        "expected_source": expect_string(index_document.get("source"), "source"),
        "required_collection_batch": expect_string(index_document.get("collection_batch"), "collection_batch"),
    }

    capture_origin = str(index_document.get("capture_origin") or "").strip()
    if capture_origin:
        candidate_response_record["required_capture_origin"] = capture_origin

    required_tags = build_stable_required_tags(record)
    if required_tags:
        candidate_response_record["required_tags"] = required_tags

    evaluation = copy.deepcopy(expect_object(base_sample.get("evaluation"), "source sample evaluation"))
    scoring_focus = unique_strings([str(value) for value in evaluation.get("scoring_focus") or []])
    if "candidate_record_alignment" not in scoring_focus:
        scoring_focus.append("candidate_record_alignment")
    evaluation["scoring_focus"] = scoring_focus

    expected_candidate_violations = unique_strings(
        [str(value) for value in entry.get("expected_candidate_violations") or []]
    )
    if not expected_candidate_violations:
        expected_candidate_violations = derive_expected_candidate_violations(
            [str(value) for value in entry.get("audit_violations") or []]
        )
    if not expected_candidate_violations:
        raise SystemExit(
            f"unable to derive expected_candidate_violations for record_id '{entry.get('record_id')}'"
        )

    document = {
        "schema_version": base_sample["schema_version"],
        "sample_id": base_sample["sample_id"],
        "project": base_sample["project"],
        "task": base_sample["task"],
        "input_request": base_sample["input_request"],
        "retrieval_expectations": base_sample["retrieval_expectations"],
        "expected_response_shape": base_sample["expected_response_shape"],
        "evaluation": evaluation,
        "negative_replay_expectations": {
            "expected_candidate_violations": expected_candidate_violations,
        },
        "candidate_response_record": candidate_response_record,
        "golden_response": base_sample["golden_response"],
    }

    document["evaluation"]["notes"] = build_note(
        expect_string(index_document.get("collection_batch"), "collection_batch"),
        expected_candidate_violations,
        expect_string(entry.get("replay_mode"), "entry.replay_mode"),
    )

    ensure_schema(document, RADISH_SAMPLE_SCHEMA_PATH, str(document.get("sample_id") or "generated negative replay"))
    return document


def main() -> int:
    args = parse_args()
    index_path = resolve_relative_to_repo(args.index)
    if not index_path.is_file():
        raise SystemExit(f"index file not found: {args.index}")

    index_document = load_index(index_path)
    output_dir = (
        resolve_relative_to_repo(args.output_dir)
        if args.output_dir.strip()
        else resolve_relative_to_repo(expect_string(index_document.get("negative_sample_dir"), "negative_sample_dir"))
    )

    manifest_path = resolve_relative_to_repo(expect_string(index_document.get("source_manifest_path"), "source_manifest_path"))
    manifest = load_manifest(manifest_path)
    record_map = build_record_map(manifest, manifest_path)
    audit_sample_dir = resolve_relative_to_repo(expect_string(index_document.get("audit_sample_dir"), "audit_sample_dir"))
    entries = filter_entries(build_index_entries(index_document), args.group_id, args.record_id, args.replay_mode)

    generated_count = 0
    checked_count = 0

    for entry in entries:
        source_sample_file = expect_string(entry.get("source_sample_file"), "entry.source_sample_file")
        source_sample = load_template_sample(entry=entry, audit_sample_dir=audit_sample_dir)

        record_id = expect_string(entry.get("record_id"), "entry.record_id")
        record_path = record_map.get(record_id)
        if record_path is None:
            raise SystemExit(f"manifest does not contain record_id '{record_id}'")
        if not record_path.is_file():
            raise SystemExit(f"record file not found: {make_repo_relative(record_path)}")
        record = load_record(record_path)

        expected_candidate_violations = unique_strings(
            [str(value) for value in entry.get("expected_candidate_violations") or []]
        ) or derive_expected_candidate_violations([str(value) for value in entry.get("audit_violations") or []])
        negative_sample_path_value = str(entry.get("negative_sample_path") or "").strip()
        output_path = (
            resolve_relative_to_repo(negative_sample_path_value)
            if negative_sample_path_value
            else derive_output_path(source_sample_file, record, expected_candidate_violations, output_dir)
        )
        document = build_negative_sample(
            source_sample,
            index_document=index_document,
            entry={**entry, "expected_candidate_violations": expected_candidate_violations},
            record=record,
        )

        if args.check:
            if not output_path.is_file():
                raise SystemExit(f"generated negative replay fixture is missing: {make_repo_relative(output_path)}")
            existing = load_json_document(output_path)
            if existing != document:
                raise SystemExit(f"negative replay fixture is out of date: {make_repo_relative(output_path)}")
            checked_count += 1
            continue

        write_json_document(output_path, document)
        generated_count += 1

    if args.check:
        print(f"negative replay fixtures are up to date: {checked_count} checked", file=sys.stdout)
    else:
        print(f"wrote {generated_count} negative replay fixture(s) to {make_repo_relative(output_dir)}", file=sys.stdout)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
