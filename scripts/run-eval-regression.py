#!/usr/bin/env python3
from __future__ import annotations

import json
import re
import sys
from pathlib import Path
from typing import Any

try:
    import jsonschema
except ModuleNotFoundError:
    print(
        "python package 'jsonschema' is required for eval runners. "
        "Install it in the active environment before running repository checks.",
        file=sys.stderr,
    )
    raise SystemExit(2)


REPO_ROOT = Path(__file__).resolve().parent.parent
ALLOWED_PRIMARY_KINDS = {"markdown", "text"}
UNOFFICIAL_SOURCE_TYPES = {"faq", "forum"}
RISK_RANKS = {"low": 1, "medium": 2, "high": 3}
NEGATIVE_REPLAY_MODES = {"same_sample", "cross_sample"}
RECENT_GHOST_ACTION_REVISION_KEYS = {
    "accept_ghost_completion": "accepted_at_revision",
    "reject_ghost_completion": "rejected_at_revision",
    "dismiss_ghost_completion": "dismissed_at_revision",
    "skip_ghost_completion": "skipped_at_revision",
}
SUPPRESSIVE_RECENT_GHOST_ACTION_KINDS = {
    "reject_ghost_completion",
    "dismiss_ghost_completion",
    "skip_ghost_completion",
}
NEGATIVE_REPLAY_INDEX_SCHEMA_PATH = REPO_ROOT / "datasets/eval/negative-replay-index.schema.json"
BATCH_ARTIFACT_SUMMARY_SCHEMA_PATH = REPO_ROOT / "datasets/eval/batch-orchestration-summary.schema.json"
DISALLOWED_SUGGEST_PATCH_KEYS = {
    "command",
    "commands",
    "shell",
    "script",
    "execute",
    "execute_now",
    "apply",
    "apply_immediately",
    "full_document",
    "full_document_json",
    "flowsheet_document",
    "document",
}
UNCERTAINTY_MARKERS = (
    "候选",
    "可能",
    "证据不足",
    "不足以确认",
    "还不能",
    "暂不能",
    "仅能",
    "推断",
    "未确认",
)
SCHEMA_CACHE: dict[Path, Any] = {}


TASK_CONFIG = {
    "radish-docs-qa": {
        "sample_dir": REPO_ROOT / "datasets/eval/radish",
        "sample_schema": REPO_ROOT / "datasets/eval/radish-task-sample.schema.json",
        "candidate_record_schema": REPO_ROOT / "datasets/eval/candidate-response-record.schema.json",
        "candidate_record_batch_schema": REPO_ROOT / "datasets/eval/candidate-record-batch.schema.json",
        "request_schema": REPO_ROOT / "contracts/copilot-request.schema.json",
        "response_schema": REPO_ROOT / "contracts/copilot-response.schema.json",
        "task_card": REPO_ROOT / "docs/task-cards/radish-answer-docs-question.md",
        "sample_task": "answer_docs_question",
        "success_message": "radish docs qa regression passed.",
        "no_sample_message": "no answer_docs_question sample files found for Radish docs QA regression",
        "warning_prefix": "radish docs qa regression found",
    },
    "radish-docs-qa-negative": {
        "sample_dir": REPO_ROOT / "datasets/eval/radish-negative",
        "sample_schema": REPO_ROOT / "datasets/eval/radish-task-sample.schema.json",
        "candidate_record_schema": REPO_ROOT / "datasets/eval/candidate-response-record.schema.json",
        "candidate_record_batch_schema": REPO_ROOT / "datasets/eval/candidate-record-batch.schema.json",
        "request_schema": REPO_ROOT / "contracts/copilot-request.schema.json",
        "response_schema": REPO_ROOT / "contracts/copilot-response.schema.json",
        "task_card": REPO_ROOT / "docs/task-cards/radish-answer-docs-question.md",
        "sample_task": "answer_docs_question",
        "success_message": "radish docs qa negative regression passed.",
        "no_sample_message": "no answer_docs_question sample files found for Radish docs QA negative regression",
        "warning_prefix": "radish docs qa negative regression found",
    },
    "radishflow-diagnostics": {
        "sample_dir": REPO_ROOT / "datasets/eval/radishflow",
        "sample_schema": REPO_ROOT / "datasets/eval/radishflow-task-sample.schema.json",
        "candidate_record_schema": REPO_ROOT / "datasets/eval/candidate-response-record.schema.json",
        "request_schema": REPO_ROOT / "contracts/copilot-request.schema.json",
        "response_schema": REPO_ROOT / "contracts/copilot-response.schema.json",
        "task_card": REPO_ROOT / "docs/task-cards/radishflow-explain-diagnostics.md",
        "sample_task": "explain_diagnostics",
        "success_message": "radishflow diagnostics regression passed.",
        "no_sample_message": "no explain_diagnostics sample files found for RadishFlow diagnostics regression",
        "warning_prefix": "radishflow diagnostics regression found",
    },
    "radishflow-suggest-edits": {
        "sample_dir": REPO_ROOT / "datasets/eval/radishflow",
        "sample_schema": REPO_ROOT / "datasets/eval/radishflow-task-sample.schema.json",
        "candidate_record_schema": REPO_ROOT / "datasets/eval/candidate-response-record.schema.json",
        "request_schema": REPO_ROOT / "contracts/copilot-request.schema.json",
        "response_schema": REPO_ROOT / "contracts/copilot-response.schema.json",
        "task_card": REPO_ROOT / "docs/task-cards/radishflow-suggest-flowsheet-edits.md",
        "sample_task": "suggest_flowsheet_edits",
        "success_message": "radishflow suggest edits regression passed.",
        "no_sample_message": "no suggest_flowsheet_edits sample files found for RadishFlow suggest edits regression",
        "warning_prefix": "radishflow suggest edits regression found",
    },
    "radishflow-ghost-completion": {
        "sample_dir": REPO_ROOT / "datasets/eval/radishflow",
        "sample_schema": REPO_ROOT / "datasets/eval/radishflow-task-sample.schema.json",
        "candidate_record_schema": REPO_ROOT / "datasets/eval/candidate-response-record.schema.json",
        "request_schema": REPO_ROOT / "contracts/copilot-request.schema.json",
        "response_schema": REPO_ROOT / "contracts/copilot-response.schema.json",
        "task_card": REPO_ROOT / "docs/task-cards/radishflow-suggest-ghost-completion.md",
        "sample_task": "suggest_ghost_completion",
        "success_message": "radishflow ghost completion regression passed.",
        "no_sample_message": "no suggest_ghost_completion sample files found for RadishFlow ghost completion regression",
        "warning_prefix": "radishflow ghost completion regression found",
    },
    "radishflow-control-plane": {
        "sample_dir": REPO_ROOT / "datasets/eval/radishflow",
        "sample_schema": REPO_ROOT / "datasets/eval/radishflow-task-sample.schema.json",
        "candidate_record_schema": REPO_ROOT / "datasets/eval/candidate-response-record.schema.json",
        "request_schema": REPO_ROOT / "contracts/copilot-request.schema.json",
        "response_schema": REPO_ROOT / "contracts/copilot-response.schema.json",
        "task_card": REPO_ROOT / "docs/task-cards/radishflow-explain-control-plane-state.md",
        "sample_task": "explain_control_plane_state",
        "success_message": "radishflow control plane regression passed.",
        "no_sample_message": "no explain_control_plane_state sample files found for RadishFlow control plane regression",
        "warning_prefix": "radishflow control plane regression found",
    },
}


def add_violation(violations: list[str], message: str) -> None:
    violations.append(message)


def get_array(value: Any) -> list[Any]:
    if value is None:
        return []
    if isinstance(value, list):
        return value
    return [value]


def get_immediately_suppressed_candidate_refs(request_context: dict[str, Any]) -> set[str]:
    document_revision = request_context.get("document_revision")
    if not isinstance(document_revision, int):
        return set()

    suppressed_candidate_refs: set[str] = set()
    recent_actions = get_array((request_context.get("cursor_context") or {}).get("recent_actions"))
    for recent_action in recent_actions:
        action_kind = str((recent_action or {}).get("kind") or "").strip()
        if action_kind not in SUPPRESSIVE_RECENT_GHOST_ACTION_KINDS:
            continue

        candidate_ref = str((recent_action or {}).get("candidate_ref") or "").strip()
        if not candidate_ref:
            continue

        revision_key = RECENT_GHOST_ACTION_REVISION_KEYS.get(action_kind)
        action_revision = (recent_action or {}).get(revision_key or "")
        if not isinstance(action_revision, int):
            continue

        if document_revision - action_revision == 1:
            suppressed_candidate_refs.add(candidate_ref)

    return suppressed_candidate_refs


def get_value(obj: Any, name: str) -> tuple[bool, Any]:
    if obj is None:
        return False, None
    if isinstance(obj, dict):
        return (name in obj, obj.get(name))
    return False, None


def convert_expectation_literal(literal: str) -> Any:
    trimmed = literal.strip()
    lowered = trimmed.lower()
    if lowered == "true":
        return True
    if lowered == "false":
        return False
    if lowered == "null":
        return None
    if re.fullmatch(r"-?\d+", trimmed):
        return int(trimmed)
    if re.fullmatch(r"-?\d+\.\d+", trimmed):
        return float(trimmed)
    if len(trimmed) >= 2 and trimmed.startswith('"') and trimmed.endswith('"'):
        return trimmed[1:-1]
    return trimmed


def contains_uncertainty_marker(text: Any) -> bool:
    content = str(text or "")
    return any(marker in content for marker in UNCERTAINTY_MARKERS)


def parse_artifact_name_from_locator(locator: Any) -> str:
    locator_text = str(locator or "")
    if not locator_text.startswith("artifact:"):
        return ""
    artifact_ref = locator_text[len("artifact:") :]
    return artifact_ref.split(".", 1)[0].strip()


def resolve_json_path(document: Any, path: str) -> tuple[bool, Any]:
    if not path or not path.startswith("$"):
        raise ValueError(f"unsupported json path: {path}")

    cursor = document
    remaining = path[1:]

    while remaining:
        match = re.match(r"^\.([A-Za-z0-9_]+)(.*)$", remaining)
        if match:
            exists, value = get_value(cursor, match.group(1))
            if not exists:
                return False, None
            cursor = value
            remaining = match.group(2)
            continue

        match = re.match(r"^\[(\d+)\](.*)$", remaining)
        if match:
            index = int(match.group(1))
            items = get_array(cursor)
            if index >= len(items):
                return False, None
            cursor = items[index]
            remaining = match.group(2)
            continue

        raise ValueError(f"unsupported json path: {path}")

    return True, cursor


def test_path_expectations(
    document: Any,
    expressions: list[str],
    should_exist: bool,
    sample_name: str,
    violations: list[str],
) -> None:
    for expression in get_array(expressions):
        text = str(expression).strip()
        if not text:
            continue

        parts = text.split("=", 1)
        path = parts[0]
        has_expected_value = len(parts) == 2
        expected_value = convert_expectation_literal(parts[1]) if has_expected_value else None

        exists, value = resolve_json_path(document, path)
        if should_exist:
            if not exists:
                add_violation(violations, f"{sample_name}: missing required json path '{text}'")
                continue
            if has_expected_value and value != expected_value:
                add_violation(
                    violations,
                    f"{sample_name}: json path '{path}' expected '{expected_value}' but found '{value}'",
                )
            continue

        if not has_expected_value:
            if exists:
                add_violation(violations, f"{sample_name}: json path '{path}' should not exist")
            continue

        if exists and value == expected_value:
            add_violation(violations, f"{sample_name}: json path '{path}' should not equal '{expected_value}'")


def load_schema(schema_path: Path) -> Any:
    if schema_path not in SCHEMA_CACHE:
        SCHEMA_CACHE[schema_path] = json.loads(schema_path.read_text(encoding="utf-8"))
    return SCHEMA_CACHE[schema_path]


def test_document_against_schema(
    document: Any,
    schema_path: Path,
    document_name: str,
    violations: list[str],
) -> None:
    schema = load_schema(schema_path)
    try:
        jsonschema.validate(document, schema)
    except jsonschema.ValidationError as exc:
        add_violation(
            violations,
            f"{document_name}: schema validation failed against '{schema_path.name}': {exc.message}",
        )


def resolve_repo_relative_path(path_value: str) -> Path:
    path = Path(path_value)
    if not path.is_absolute():
        path = (REPO_ROOT / path).resolve()
    return path


def load_json_document(
    document_path: Path,
    document_label: str,
    violations: list[str],
) -> Any:
    try:
        return json.loads(document_path.read_text(encoding="utf-8"))
    except Exception as exc:
        add_violation(violations, f"{document_label}: failed to parse json: {exc}")
        return None


def load_negative_replay_index(index_path: Path) -> dict[str, Any]:
    try:
        document = json.loads(index_path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse negative replay index '{index_path}': {exc}") from exc
    if not isinstance(document, dict):
        raise SystemExit(f"negative replay index must be a json object: {index_path}")
    try:
        jsonschema.validate(document, load_schema(NEGATIVE_REPLAY_INDEX_SCHEMA_PATH))
    except jsonschema.ValidationError as exc:
        raise SystemExit(
            f"negative replay index '{index_path}': schema validation failed against "
            f"'{NEGATIVE_REPLAY_INDEX_SCHEMA_PATH.name}': {exc.message}"
        ) from exc
    return document


def load_batch_artifact_summary(summary_path: Path) -> dict[str, Any]:
    try:
        document = json.loads(summary_path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse batch artifact summary '{summary_path}': {exc}") from exc
    if not isinstance(document, dict):
        raise SystemExit(f"batch artifact summary must be a json object: {summary_path}")
    try:
        jsonschema.validate(document, load_schema(BATCH_ARTIFACT_SUMMARY_SCHEMA_PATH))
    except jsonschema.ValidationError as exc:
        raise SystemExit(
            f"batch artifact summary '{summary_path}': schema validation failed against "
            f"'{BATCH_ARTIFACT_SUMMARY_SCHEMA_PATH.name}': {exc.message}"
        ) from exc
    return document


def resolve_negative_replay_index_from_batch_artifact_summary(summary_path: Path, replay_mode: str = "") -> Path:
    summary_document = load_batch_artifact_summary(summary_path)
    eval_task = str(summary_document.get("eval_task") or "").strip()
    if eval_task != "radish-docs-qa":
        raise SystemExit(
            f"batch artifact summary '{summary_path}': unsupported eval_task '{eval_task}', expected 'radish-docs-qa'"
        )

    artifacts = summary_document.get("artifacts")
    if not isinstance(artifacts, dict):
        raise SystemExit(f"batch artifact summary '{summary_path}': artifacts must be a json object")
    normalized_replay_mode = replay_mode.strip() or "same_sample"
    artifact_key = "cross_sample_negative_replay_index" if normalized_replay_mode == "cross_sample" else "negative_replay_index"
    negative_replay_index = artifacts.get(artifact_key)
    if not isinstance(negative_replay_index, dict):
        raise SystemExit(f"batch artifact summary '{summary_path}': artifacts.{artifact_key} must be a json object")

    index_path_value = str(negative_replay_index.get("path") or "").strip()
    if not index_path_value:
        raise SystemExit(f"batch artifact summary '{summary_path}': artifacts.{artifact_key}.path is required")
    if negative_replay_index.get("requested") is not True:
        raise SystemExit(
            f"batch artifact summary '{summary_path}': artifacts.{artifact_key}.requested must be true"
        )
    if negative_replay_index.get("exists") is not True:
        raise SystemExit(f"batch artifact summary '{summary_path}': artifacts.{artifact_key}.exists must be true")

    index_path = resolve_repo_relative_path(index_path_value)
    if not index_path.is_file():
        raise SystemExit(
            f"batch artifact summary '{summary_path}': referenced negative replay index file not found: {index_path_value}"
        )
    return index_path


def resolve_recommended_negative_replay_groups_from_batch_artifact_summary(
    summary_path: Path,
    top_n: int,
    replay_mode: str = "",
) -> tuple[list[str], str]:
    if top_n <= 0:
        raise SystemExit("--recommended-groups-top must be greater than 0")

    summary_document = load_batch_artifact_summary(summary_path)
    recommended = summary_document.get("recommended_negative_replays")
    if not isinstance(recommended, dict):
        raise SystemExit(
            f"batch artifact summary '{summary_path}': recommended_negative_replays must be a json object"
        )

    default_replay_mode = str(recommended.get("default_replay_mode") or "").strip()
    if not default_replay_mode:
        raise SystemExit(
            f"batch artifact summary '{summary_path}': recommended_negative_replays.default_replay_mode is required"
        )
    if default_replay_mode not in NEGATIVE_REPLAY_MODES:
        raise SystemExit(
            f"batch artifact summary '{summary_path}': unsupported recommended default_replay_mode "
            f"'{default_replay_mode}'"
        )

    selected_replay_mode = replay_mode.strip() or default_replay_mode
    if selected_replay_mode not in NEGATIVE_REPLAY_MODES:
        raise SystemExit(
            f"batch artifact summary '{summary_path}': unsupported recommended replay_mode '{selected_replay_mode}'"
        )
    group_id_field = "cross_sample_recommended_group_ids" if selected_replay_mode == "cross_sample" else "recommended_group_ids"
    recommended_group_ids = [
        str(group_id).strip()
        for group_id in get_array(recommended.get(group_id_field))
        if str(group_id).strip()
    ]
    if not recommended_group_ids:
        raise SystemExit(
            f"batch artifact summary '{summary_path}': recommended_negative_replays.{group_id_field} is empty"
        )

    selected_group_ids = recommended_group_ids[:top_n]
    if not selected_group_ids:
        raise SystemExit(
            f"batch artifact summary '{summary_path}': no recommended group_ids matched --recommended-groups-top {top_n}"
        )
    return selected_group_ids, selected_replay_mode


def resolve_negative_replay_sample_paths(
    index_path: Path,
    *,
    group_ids: list[str],
    record_ids: list[str],
    replay_mode: str,
) -> list[Path]:
    index_document = load_negative_replay_index(index_path)
    group_entries: dict[str, list[dict[str, Any]]] = {}
    all_entries: list[dict[str, Any]] = []

    for group in get_array(index_document.get("violation_groups")):
        if not isinstance(group, dict):
            raise SystemExit(f"negative replay index '{index_path}': violation_groups entries must be json objects")
        group_id = str(group.get("group_id") or "").strip()
        if not group_id:
            raise SystemExit(f"negative replay index '{index_path}': violation_groups entries must include group_id")
        entries: list[dict[str, Any]] = []
        for entry in get_array(group.get("entries")):
            if not isinstance(entry, dict):
                raise SystemExit(
                    f"negative replay index '{index_path}': violation group '{group_id}' entries must be json objects"
                )
            entries.append(entry)
            all_entries.append(entry)
        group_entries[group_id] = entries

    for entry in get_array(index_document.get("linked_non_failed_samples")):
        if not isinstance(entry, dict):
            raise SystemExit(f"negative replay index '{index_path}': linked_non_failed_samples entries must be json objects")
        all_entries.append(entry)

    selected_entries: list[dict[str, Any]] = []
    normalized_group_ids = [group_id for group_id in group_ids if group_id]
    if normalized_group_ids:
        missing_group_ids = [group_id for group_id in normalized_group_ids if group_id not in group_entries]
        if missing_group_ids:
            raise SystemExit(
                f"negative replay index '{index_path}': unknown group_id value(s): {', '.join(sorted(missing_group_ids))}"
            )
        for group_id in normalized_group_ids:
            selected_entries.extend(group_entries[group_id])
    else:
        selected_entries.extend(all_entries)

    normalized_record_ids = {record_id for record_id in record_ids if record_id}
    if normalized_record_ids:
        selected_entries = [
            entry for entry in selected_entries if str(entry.get("record_id") or "").strip() in normalized_record_ids
        ]
        if not selected_entries:
            raise SystemExit(
                f"negative replay index '{index_path}': no replay samples matched the requested record_id filters"
            )

    normalized_replay_mode = replay_mode.strip()
    if normalized_replay_mode:
        if normalized_replay_mode not in NEGATIVE_REPLAY_MODES:
            raise SystemExit(
                f"unsupported --replay-mode '{normalized_replay_mode}'; expected one of: "
                f"{', '.join(sorted(NEGATIVE_REPLAY_MODES))}"
            )
        selected_entries = [
            entry for entry in selected_entries if str(entry.get("replay_mode") or "").strip() == normalized_replay_mode
        ]
        if not selected_entries:
            raise SystemExit(
                f"negative replay index '{index_path}': no replay samples matched the requested replay_mode filter"
            )

    selected_paths: list[Path] = []
    seen_paths: set[Path] = set()
    for entry in selected_entries:
        sample_path_value = str(entry.get("negative_sample_path") or "").strip()
        if not sample_path_value:
            raise SystemExit(f"negative replay index '{index_path}': selected replay entry is missing negative_sample_path")
        sample_path = resolve_repo_relative_path(sample_path_value)
        if not sample_path.is_file():
            raise SystemExit(
                f"negative replay index '{index_path}': selected replay sample file not found: {sample_path_value}"
            )
        if sample_path in seen_paths:
            continue
        seen_paths.add(sample_path)
        selected_paths.append(sample_path)

    if not selected_paths:
        raise SystemExit(f"negative replay index '{index_path}': no replay samples matched the requested filters")

    return sorted(selected_paths)


def load_candidate_response_from_record(
    sample: dict[str, Any],
    config: dict[str, Any],
    sample_name: str,
) -> tuple[Any, list[str]]:
    record_violations: list[str] = []
    record_ref = sample.get("candidate_response_record")
    if record_ref is None:
        return sample.get("candidate_response"), record_violations

    record_path_value = str((record_ref or {}).get("path") or "").strip()
    manifest_path_value = str((record_ref or {}).get("manifest_path") or "").strip()
    manifest_record_id = str((record_ref or {}).get("record_id") or "").strip()

    if record_path_value and manifest_path_value:
        add_violation(
            record_violations,
            f"{sample_name}: candidate_response_record.path and candidate_response_record.manifest_path cannot be used together",
        )
        return None, record_violations

    manifest: dict[str, Any] | None = None
    manifest_entry: dict[str, Any] | None = None
    resolved_record_label = record_path_value

    if record_path_value:
        record_path = resolve_repo_relative_path(record_path_value)
        if not record_path.is_file():
            add_violation(record_violations, f"{sample_name}: candidate_response_record file not found: {record_path_value}")
            return None, record_violations
    else:
        if not manifest_path_value:
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record.path or candidate_response_record.manifest_path is required",
            )
            return None, record_violations
        if not manifest_record_id:
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record.record_id is required when candidate_response_record.manifest_path is used",
            )
            return None, record_violations

        manifest_path = resolve_repo_relative_path(manifest_path_value)
        if not manifest_path.is_file():
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record manifest file not found: {manifest_path_value}",
            )
            return None, record_violations

        manifest_document = load_json_document(
            manifest_path,
            f"{sample_name} candidate_response_record manifest '{manifest_path_value}'",
            record_violations,
        )
        if manifest_document is None or not isinstance(manifest_document, dict):
            return None, record_violations
        manifest = manifest_document

        candidate_record_batch_schema = config.get("candidate_record_batch_schema")
        if candidate_record_batch_schema is not None:
            test_document_against_schema(
                manifest,
                candidate_record_batch_schema,
                f"{sample_name} candidate_response_record manifest",
                record_violations,
            )

        if str(manifest.get("project") or "") != str(sample.get("project") or ""):
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record manifest project must match sample.project",
            )
        if str(manifest.get("task") or "") != str(sample.get("task") or ""):
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record manifest task must match sample.task",
            )

        matching_entries = [
            entry
            for entry in get_array(manifest.get("records"))
            if isinstance(entry, dict) and str(entry.get("record_id") or "") == manifest_record_id
        ]
        if len(matching_entries) == 0:
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record manifest does not contain record_id '{manifest_record_id}'",
            )
            return None, record_violations
        if len(matching_entries) > 1:
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record manifest has duplicate record_id '{manifest_record_id}'",
            )
            return None, record_violations
        manifest_entry = matching_entries[0]

        entry_path_value = str((manifest_entry or {}).get("path") or "").strip()
        if not entry_path_value:
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record manifest entry '{manifest_record_id}' is missing path",
            )
            return None, record_violations

        record_path = resolve_repo_relative_path(entry_path_value)
        resolved_record_label = f"{manifest_path_value}#{manifest_record_id}"
        if not record_path.is_file():
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record file referenced by manifest was not found: {entry_path_value}",
            )
            return None, record_violations

    record_document = load_json_document(
        record_path,
        f"{sample_name} candidate_response_record '{resolved_record_label}'",
        record_violations,
    )
    if record_document is None or not isinstance(record_document, dict):
        return None, record_violations
    record = record_document

    candidate_record_schema = config.get("candidate_record_schema")
    if candidate_record_schema is not None:
        test_document_against_schema(
            record,
            candidate_record_schema,
            f"{sample_name} candidate_response_record",
            record_violations,
        )

    if manifest is not None and manifest_entry is not None:
        if str(record.get("record_id") or "") != str(manifest_entry.get("record_id") or ""):
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record.record_id must match manifest record_id",
            )
        if str(record.get("sample_id") or "") != str(manifest_entry.get("sample_id") or ""):
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record.sample_id must match manifest sample_id",
            )
        if str(record.get("project") or "") != str(manifest.get("project") or ""):
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record.project must match manifest project",
            )
        if str(record.get("task") or "") != str(manifest.get("task") or ""):
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record.task must match manifest task",
            )

        manifest_source = str(manifest.get("source") or "").strip()
        if manifest_source and str(record.get("source") or "") != manifest_source:
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record.source must match manifest source",
            )

        capture_metadata = record.get("capture_metadata") or {}
        manifest_collection_batch = str(manifest.get("collection_batch") or "").strip()
        if manifest_collection_batch and str(capture_metadata.get("collection_batch") or "") != manifest_collection_batch:
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record.capture_metadata.collection_batch must match manifest collection_batch",
            )

        manifest_capture_origin = str(manifest.get("capture_origin") or "").strip()
        if manifest_capture_origin and str(capture_metadata.get("capture_origin") or "") != manifest_capture_origin:
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record.capture_metadata.capture_origin must match manifest capture_origin",
            )

    expected_sample_id = str(sample.get("sample_id") or "")
    if str(record.get("sample_id") or "") != expected_sample_id:
        add_violation(record_violations, f"{sample_name}: candidate_response_record.sample_id must match sample_id")
    if str(record.get("project") or "") != str(sample.get("project") or ""):
        add_violation(record_violations, f"{sample_name}: candidate_response_record.project must match sample.project")
    if str(record.get("task") or "") != str(sample.get("task") or ""):
        add_violation(record_violations, f"{sample_name}: candidate_response_record.task must match sample.task")

    sample_request = sample.get("input_request") or {}
    if str(record.get("request_id") or "") != str(sample_request.get("request_id") or ""):
        add_violation(
            record_violations,
            f"{sample_name}: candidate_response_record.request_id must match input_request.request_id",
        )

    expected_source = str((record_ref or {}).get("expected_source") or "").strip()
    if expected_source and str(record.get("source") or "") != expected_source:
        add_violation(
            record_violations,
            f"{sample_name}: candidate_response_record.source must match candidate_response_record.expected_source",
        )

    capture_metadata = record.get("capture_metadata") or {}
    required_capture_origin = str((record_ref or {}).get("required_capture_origin") or "").strip()
    if required_capture_origin and str(capture_metadata.get("capture_origin") or "") != required_capture_origin:
        add_violation(
            record_violations,
            f"{sample_name}: candidate_response_record.capture_metadata.capture_origin must match candidate_response_record.required_capture_origin",
        )

    required_collection_batch = str((record_ref or {}).get("required_collection_batch") or "").strip()
    if required_collection_batch and str(capture_metadata.get("collection_batch") or "") != required_collection_batch:
        add_violation(
            record_violations,
            f"{sample_name}: candidate_response_record.capture_metadata.collection_batch must match candidate_response_record.required_collection_batch",
        )

    required_tags = {
        str(tag).strip()
        for tag in get_array((record_ref or {}).get("required_tags"))
        if str(tag).strip()
    }
    record_tags = {
        str(tag).strip()
        for tag in get_array(capture_metadata.get("tags"))
        if str(tag).strip()
    }
    for required_tag in sorted(required_tags):
        if required_tag not in record_tags:
            add_violation(
                record_violations,
                f"{sample_name}: candidate_response_record.capture_metadata.tags is missing required tag '{required_tag}'",
            )

    input_record = record.get("input_record") or {}
    context = sample_request.get("context") or {}
    resource = context.get("resource") or {}
    artifact_names = sorted(
        str((artifact or {}).get("name") or "").strip()
        for artifact in get_array(sample_request.get("artifacts"))
        if str((artifact or {}).get("name") or "").strip()
    )
    record_artifact_names = sorted(
        str(name).strip()
        for name in get_array(input_record.get("artifact_names"))
        if str(name).strip()
    )
    if record_artifact_names != artifact_names:
        add_violation(
            record_violations,
            f"{sample_name}: candidate_response_record.input_record.artifact_names must match input_request artifacts",
        )
    if str(input_record.get("route") or "") != str(context.get("route") or ""):
        add_violation(
            record_violations,
            f"{sample_name}: candidate_response_record.input_record.route must match input_request.context.route",
        )
    if str(input_record.get("resource_slug") or "") != str(resource.get("slug") or ""):
        add_violation(
            record_violations,
            f"{sample_name}: candidate_response_record.input_record.resource_slug must match input_request.context.resource.slug",
        )
    if str(input_record.get("current_app") or "") != str(context.get("current_app") or ""):
        add_violation(
            record_violations,
            f"{sample_name}: candidate_response_record.input_record.current_app must match input_request.context.current_app",
        )

    record_search_scope = sorted(str(scope).strip() for scope in get_array(input_record.get("search_scope")) if str(scope).strip())
    sample_search_scope = sorted(str(scope).strip() for scope in get_array(context.get("search_scope")) if str(scope).strip())
    if record_search_scope != sample_search_scope:
        add_violation(
            record_violations,
            f"{sample_name}: candidate_response_record.input_record.search_scope must match input_request.context.search_scope",
        )

    response = record.get("response")
    if response is None:
        add_violation(record_violations, f"{sample_name}: candidate_response_record.response is required")
        return None, record_violations

    return response, record_violations


def validate_radish_docs_retrieval(sample: dict[str, Any], sample_name: str, violations: list[str]) -> None:
    request = sample["input_request"]
    context = request.get("context") or {}
    retrieval = sample["retrieval_expectations"]
    artifacts = get_array(request.get("artifacts"))
    primary_artifacts = [artifact for artifact in artifacts if artifact.get("role") == "primary"]
    supporting_artifacts = [artifact for artifact in artifacts if artifact.get("role") == "supporting"]
    reference_artifacts = [artifact for artifact in artifacts if artifact.get("role") == "reference"]
    search_scopes = [str(scope) for scope in get_array(context.get("search_scope"))]

    if retrieval.get("require_primary_artifact") and len(primary_artifacts) < 1:
        add_violation(violations, f"{sample_name}: retrieval_expectations require a primary artifact")
    if len(primary_artifacts) > int(retrieval.get("max_primary_artifacts", 0)):
        add_violation(violations, f"{sample_name}: primary artifact count exceeds retrieval_expectations.max_primary_artifacts")
    if len(supporting_artifacts) > int(retrieval.get("max_supporting_artifacts", 0)):
        add_violation(violations, f"{sample_name}: supporting artifact count exceeds retrieval_expectations.max_supporting_artifacts")
    if len(reference_artifacts) > int(retrieval.get("max_reference_artifacts", 0)):
        add_violation(violations, f"{sample_name}: reference artifact count exceeds retrieval_expectations.max_reference_artifacts")

    allowed_search_scopes = {str(scope) for scope in get_array(retrieval.get("allowed_search_scopes"))}
    required_search_scopes = {str(scope) for scope in get_array(retrieval.get("required_search_scopes"))}
    disallowed_search_scopes = {str(scope) for scope in get_array(retrieval.get("disallowed_search_scopes"))}
    for scope in search_scopes:
        if scope not in allowed_search_scopes:
            add_violation(violations, f"{sample_name}: search_scope '{scope}' is not allowed by retrieval_expectations")
    for required_scope in required_search_scopes:
        if required_scope not in search_scopes:
            add_violation(violations, f"{sample_name}: search_scope is missing required scope '{required_scope}'")
    for disallowed_scope in disallowed_search_scopes:
        if disallowed_scope in search_scopes:
            add_violation(violations, f"{sample_name}: search_scope contains disallowed scope '{disallowed_scope}'")

    resource = context.get("resource") or {}
    resource_slug = str(resource.get("slug") or "")

    for artifact in artifacts:
        role = str(artifact.get("role") or "")
        metadata = artifact.get("metadata") or {}
        source_type = str(metadata.get("source_type") or "")
        page_slug = str(metadata.get("page_slug") or "")
        fragment_id = str(metadata.get("fragment_id") or "")
        retrieval_rank = metadata.get("retrieval_rank")
        is_official = metadata.get("is_official")

        if role == "primary" and str(artifact.get("kind") or "") not in ALLOWED_PRIMARY_KINDS:
            add_violation(violations, f"{sample_name}: primary artifact kind must be markdown or text")

        if retrieval.get("require_artifact_source_metadata"):
            if not source_type:
                add_violation(violations, f"{sample_name}: artifact '{artifact.get('name')}' is missing metadata.source_type")
            if not page_slug:
                add_violation(violations, f"{sample_name}: artifact '{artifact.get('name')}' is missing metadata.page_slug")
            if not fragment_id:
                add_violation(violations, f"{sample_name}: artifact '{artifact.get('name')}' is missing metadata.fragment_id")
            if retrieval_rank is None or int(retrieval_rank) < 1:
                add_violation(violations, f"{sample_name}: artifact '{artifact.get('name')}' must carry metadata.retrieval_rank >= 1")
            if is_official is None:
                add_violation(violations, f"{sample_name}: artifact '{artifact.get('name')}' is missing metadata.is_official")

        role_key = {
            "primary": "allowed_primary_source_types",
            "supporting": "allowed_supporting_source_types",
            "reference": "allowed_reference_source_types",
        }.get(role)
        allowed_source_types = {str(item) for item in get_array(retrieval.get(role_key))} if role_key else set()
        if allowed_source_types and source_type not in allowed_source_types:
            add_violation(
                violations,
                f"{sample_name}: artifact '{artifact.get('name')}' source_type '{source_type}' is not allowed for role '{role}'",
            )

        if retrieval.get("require_primary_resource_match") and role == "primary" and resource_slug and page_slug != resource_slug:
            add_violation(violations, f"{sample_name}: primary artifact '{artifact.get('name')}' must match context.resource.slug")

        if source_type in UNOFFICIAL_SOURCE_TYPES:
            if not retrieval.get("allow_unofficial_sources"):
                add_violation(violations, f"{sample_name}: unofficial source_type '{source_type}' is not allowed by retrieval_expectations")
            if role == "primary":
                add_violation(violations, f"{sample_name}: unofficial source_type '{source_type}' cannot be used as the primary artifact")
            if is_official is not False:
                add_violation(violations, f"{sample_name}: unofficial source_type '{source_type}' must set metadata.is_official to false")
            continue

        if retrieval.get("require_artifact_source_metadata") and is_official is not None and is_official is not True:
            add_violation(violations, f"{sample_name}: official source artifact '{artifact.get('name')}' must set metadata.is_official to true")


def validate_radish_docs_response(
    sample: dict[str, Any],
    response: dict[str, Any],
    response_label: str,
    sample_name: str,
    violations: list[str],
) -> None:
    shape = sample["expected_response_shape"]
    evaluation = sample["evaluation"]
    answers = get_array(response.get("answers"))
    issues = get_array(response.get("issues"))
    actions = get_array(response.get("proposed_actions"))
    citations = get_array(response.get("citations"))

    if response.get("project") != "radish":
        add_violation(violations, f"{sample_name}: {response_label}.project must be 'radish'")
    if response.get("task") != "answer_docs_question":
        add_violation(violations, f"{sample_name}: {response_label}.task must be 'answer_docs_question'")
    if str(response.get("status")) != str(shape.get("status")):
        add_violation(violations, f"{sample_name}: {response_label}.status does not match expected_response_shape.status")
    if str(response.get("risk_level")) != str(evaluation.get("expected_risk_level")):
        add_violation(violations, f"{sample_name}: {response_label}.risk_level does not match evaluation.expected_risk_level")
    if shape.get("requires_summary") and not str(response.get("summary") or "").strip():
        add_violation(violations, f"{sample_name}: {response_label}.summary is required")
    if shape.get("requires_answers") and len(answers) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 answer")
    if shape.get("requires_issues") and len(issues) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 issue")
    if not shape.get("requires_issues") and len(issues) > 0:
        add_violation(violations, f"{sample_name}: {response_label} should not contain issues")
    if shape.get("requires_citations") and len(citations) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 citation")
    if not shape.get("allow_proposed_actions") and len(actions) > 0:
        add_violation(violations, f"{sample_name}: {response_label} should not contain proposed_actions")

    actual_kinds = {str(action.get("kind") or "") for action in actions}
    for required_kind in [str(item) for item in get_array(shape.get("required_action_kinds"))]:
        if required_kind not in actual_kinds:
            add_violation(violations, f"{sample_name}: {response_label} is missing required action kind '{required_kind}'")

    for action in actions:
        if str(action.get("risk_level") or "") != "low":
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed action '{action.get('title')}' must remain low risk for answer_docs_question",
            )
        if action.get("requires_confirmation") is not False:
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed action '{action.get('title')}' must not require confirmation",
            )

    citation_ids = {str(citation.get("id") or "") for citation in citations}
    request_artifacts = {
        str((artifact or {}).get("name") or "").strip(): artifact
        for artifact in get_array((sample.get("input_request") or {}).get("artifacts"))
        if str((artifact or {}).get("name") or "").strip()
    }
    citation_artifact_meta: dict[str, tuple[bool | None, str]] = {}
    for citation in citations:
        citation_id = str(citation.get("id") or "")
        artifact_name = parse_artifact_name_from_locator(citation.get("locator"))
        if not citation_id or not artifact_name:
            continue
        artifact = request_artifacts.get(artifact_name) or {}
        metadata = (artifact or {}).get("metadata") or {}
        is_official = metadata.get("is_official")
        role = str((artifact or {}).get("role") or "")
        citation_artifact_meta[citation_id] = (is_official if isinstance(is_official, bool) else None, role)

    referenced_ids: set[str] = set()
    for answer in answers:
        referenced_ids.update(str(item) for item in get_array(answer.get("citation_ids")))
    for issue in issues:
        referenced_ids.update(str(item) for item in get_array(issue.get("citation_ids")))
    for action in actions:
        referenced_ids.update(str(item) for item in get_array(action.get("citation_ids")))
    for citation_id in sorted(citation_id for citation_id in referenced_ids if citation_id):
        if citation_id not in citation_ids:
            add_violation(violations, f"{sample_name}: referenced citation id '{citation_id}' is missing from {response_label}.citations")

    if "official_source_precedence" in {str(item) for item in get_array(evaluation.get("scoring_focus"))}:
        if len(answers) == 0:
            add_violation(violations, f"{sample_name}: {response_label} must contain answers for official_source_precedence checks")
        else:
            first_answer_citation_ids = [str(item) for item in get_array((answers[0] or {}).get("citation_ids")) if str(item).strip()]
            if not any(citation_artifact_meta.get(citation_id, (None, ""))[1] == "primary" for citation_id in first_answer_citation_ids):
                add_violation(
                    violations,
                    f"{sample_name}: {response_label} first answer must cite at least one primary artifact for official_source_precedence samples",
                )

        for answer in answers:
            answer_citation_ids = [str(item) for item in get_array(answer.get("citation_ids")) if str(item).strip()]
            unofficial_cited = any(citation_artifact_meta.get(citation_id, (None, ""))[0] is False for citation_id in answer_citation_ids)
            official_cited = any(citation_artifact_meta.get(citation_id, (None, ""))[0] is True for citation_id in answer_citation_ids)
            if unofficial_cited and not official_cited:
                add_violation(
                    violations,
                    f"{sample_name}: {response_label} answer '{answer.get('kind')}' cannot rely only on unofficial citations in official_source_precedence samples",
                )

    test_path_expectations(response, get_array(evaluation.get("must_have_json_paths")), True, f"{sample_name}:{response_label}", violations)
    test_path_expectations(response, get_array(evaluation.get("must_not_have_json_paths")), False, f"{sample_name}:{response_label}", violations)


def validate_radish_docs_negative_replay(
    sample: dict[str, Any],
    config: dict[str, Any],
    sample_name: str,
    violations: list[str],
) -> None:
    expectations = sample.get("negative_replay_expectations") or {}
    expected_candidate_violations = [
        str(item).strip()
        for item in get_array(expectations.get("expected_candidate_violations"))
        if str(item).strip()
    ]
    if len(expected_candidate_violations) == 0:
        add_violation(
            violations,
            f"{sample_name}: negative_replay_expectations.expected_candidate_violations is required for negative replay samples",
        )
        return

    candidate_response, record_violations = load_candidate_response_from_record(sample, config, sample_name)
    if candidate_response is None:
        add_violation(violations, f"{sample_name}: negative replay requires candidate_response or candidate_response_record")
        return

    candidate_violations: list[str] = []
    candidate_violations.extend(record_violations)
    test_document_against_schema(
        candidate_response,
        config["response_schema"],
        f"{sample_name} candidate_response",
        candidate_violations,
    )
    validate_radish_docs_response(sample, candidate_response, "candidate_response", sample_name, candidate_violations)

    if len(candidate_violations) == 0:
        add_violation(
            violations,
            f"{sample_name}: candidate_response unexpectedly passed all checks in negative replay mode",
        )
        return

    for expected_fragment in expected_candidate_violations:
        if not any(expected_fragment in message for message in candidate_violations):
            add_violation(
                violations,
                f"{sample_name}: negative replay did not trigger expected violation fragment '{expected_fragment}'",
            )


def validate_diagnostics_request(sample: dict[str, Any], sample_name: str, violations: list[str]) -> None:
    request = sample["input_request"]
    context = request.get("context") or {}
    artifacts = get_array(request.get("artifacts"))
    primary_artifacts = [artifact for artifact in artifacts if artifact.get("role") == "primary"]
    diagnostics = get_array(context.get("diagnostics"))
    selected_unit_ids = get_array(context.get("selected_unit_ids"))
    selected_stream_ids = get_array(context.get("selected_stream_ids"))

    if request.get("project") != "radishflow":
        add_violation(violations, f"{sample_name}: input_request.project must be 'radishflow'")
    if request.get("task") != "explain_diagnostics":
        add_violation(violations, f"{sample_name}: input_request.task must be 'explain_diagnostics'")
    if len(primary_artifacts) != 1:
        add_violation(violations, f"{sample_name}: input_request must contain exactly one primary artifact")
    else:
        primary = primary_artifacts[0]
        if str(primary.get("name") or "") != "flowsheet_document":
            add_violation(violations, f"{sample_name}: primary artifact name must be 'flowsheet_document'")
        if str(primary.get("kind") or "") != "json":
            add_violation(violations, f"{sample_name}: primary artifact kind must be 'json'")
        if str(primary.get("mime_type") or "") != "application/json":
            add_violation(violations, f"{sample_name}: primary artifact mime_type must be 'application/json'")

    if context.get("document_revision") is None:
        add_violation(violations, f"{sample_name}: context.document_revision is required")
    if context.get("diagnostic_summary") is None and len(diagnostics) == 0:
        add_violation(violations, f"{sample_name}: context must include diagnostic_summary or diagnostics")
    if len(diagnostics) == 0:
        add_violation(violations, f"{sample_name}: explain_diagnostics samples must include at least one diagnostics entry")
    for diagnostic in diagnostics:
        if not str(diagnostic.get("message") or "").strip():
            add_violation(violations, f"{sample_name}: each diagnostic must include message")
        if not str(diagnostic.get("severity") or "").strip():
            add_violation(violations, f"{sample_name}: each diagnostic must include severity")
    if len(selected_unit_ids) == 0 and len(selected_stream_ids) == 0 and str(context.get("diagnostic_scope") or "") != "global":
        add_violation(
            violations,
            f"{sample_name}: request must include selected_unit_ids, selected_stream_ids, or diagnostic_scope='global'",
        )


def validate_diagnostics_response(
    sample: dict[str, Any],
    response: dict[str, Any],
    response_label: str,
    sample_name: str,
    violations: list[str],
) -> None:
    shape = sample["expected_response_shape"]
    evaluation = sample["evaluation"]
    answers = get_array(response.get("answers"))
    issues = get_array(response.get("issues"))
    actions = get_array(response.get("proposed_actions"))
    citations = get_array(response.get("citations"))

    if response.get("project") != "radishflow":
        add_violation(violations, f"{sample_name}: {response_label}.project must be 'radishflow'")
    if response.get("task") != "explain_diagnostics":
        add_violation(violations, f"{sample_name}: {response_label}.task must be 'explain_diagnostics'")
    if str(response.get("status")) != str(shape.get("status")):
        add_violation(violations, f"{sample_name}: {response_label}.status does not match expected_response_shape.status")
    if str(response.get("risk_level")) != str(evaluation.get("expected_risk_level")):
        add_violation(violations, f"{sample_name}: {response_label}.risk_level does not match evaluation.expected_risk_level")
    if shape.get("requires_summary") and not str(response.get("summary") or "").strip():
        add_violation(violations, f"{sample_name}: {response_label}.summary is required")
    if shape.get("requires_answers") and len(answers) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 answer")
    if shape.get("requires_issues") and len(issues) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 issue")
    if shape.get("requires_citations") and len(citations) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 citation")
    if not shape.get("allow_proposed_actions") and len(actions) > 0:
        add_violation(violations, f"{sample_name}: {response_label} should not contain proposed_actions")

    actual_answer_kinds = {str(answer.get("kind") or "") for answer in answers}
    for required_answer_kind in [str(item) for item in get_array(shape.get("required_answer_kinds"))]:
        if required_answer_kind not in actual_answer_kinds:
            add_violation(violations, f"{sample_name}: {response_label} is missing required answer kind '{required_answer_kind}'")

    if "hypothesis_labeling" in {str(item) for item in get_array(evaluation.get("scoring_focus"))}:
        if "ROOT_CAUSE_UNCONFIRMED" not in {str(issue.get("code") or "") for issue in issues}:
            add_violation(
                violations,
                f"{sample_name}: {response_label} must include ROOT_CAUSE_UNCONFIRMED when hypothesis_labeling is required",
            )

        cause_answers = [answer for answer in answers if str(answer.get("kind") or "") == "cause_explanation"]
        if len(cause_answers) == 0:
            add_violation(
                violations,
                f"{sample_name}: {response_label} must contain a cause_explanation answer for hypothesis_labeling samples",
            )
        elif not any(contains_uncertainty_marker(answer.get("text")) for answer in cause_answers):
            add_violation(
                violations,
                f"{sample_name}: {response_label} cause_explanation must explicitly mark uncertainty for hypothesis_labeling samples",
            )

        root_cause_issues = [issue for issue in issues if str(issue.get("code") or "") == "ROOT_CAUSE_UNCONFIRMED"]
        if any(str(issue.get("severity") or "") != "warning" for issue in root_cause_issues):
            add_violation(
                violations,
                f"{sample_name}: {response_label} ROOT_CAUSE_UNCONFIRMED must remain warning severity",
            )
        if root_cause_issues and not any(contains_uncertainty_marker(issue.get("message")) for issue in root_cause_issues):
            add_violation(
                violations,
                f"{sample_name}: {response_label} ROOT_CAUSE_UNCONFIRMED message must explicitly state uncertainty",
            )

    actual_action_kinds = {str(action.get("kind") or "") for action in actions}
    for required_kind in [str(item) for item in get_array(shape.get("required_action_kinds"))]:
        if required_kind not in actual_action_kinds:
            add_violation(violations, f"{sample_name}: {response_label} is missing required action kind '{required_kind}'")

    if len(actions) > 0 and response.get("requires_confirmation") is not True:
        add_violation(violations, f"{sample_name}: {response_label} with proposed_actions must set requires_confirmation=true")

    citation_ids = {str(citation.get("id") or "") for citation in citations}
    referenced_ids: set[str] = set()
    for answer in answers:
        referenced_ids.update(str(item) for item in get_array(answer.get("citation_ids")))
    for issue in issues:
        referenced_ids.update(str(item) for item in get_array(issue.get("citation_ids")))
    for action in actions:
        referenced_ids.update(str(item) for item in get_array(action.get("citation_ids")))
    for citation_id in sorted(citation_id for citation_id in referenced_ids if citation_id):
        if citation_id not in citation_ids:
            add_violation(violations, f"{sample_name}: referenced citation id '{citation_id}' is missing from {response_label}.citations")

    test_path_expectations(response, get_array(evaluation.get("must_have_json_paths")), True, f"{sample_name}:{response_label}", violations)
    test_path_expectations(response, get_array(evaluation.get("must_not_have_json_paths")), False, f"{sample_name}:{response_label}", violations)


def validate_suggest_request(sample: dict[str, Any], sample_name: str, violations: list[str]) -> None:
    request = sample["input_request"]
    context = request.get("context") or {}
    artifacts = get_array(request.get("artifacts"))
    primary_artifacts = [artifact for artifact in artifacts if artifact.get("role") == "primary"]
    diagnostics = get_array(context.get("diagnostics"))
    selected_unit_ids = get_array(context.get("selected_unit_ids"))
    selected_stream_ids = get_array(context.get("selected_stream_ids"))

    if request.get("project") != "radishflow":
        add_violation(violations, f"{sample_name}: input_request.project must be 'radishflow'")
    if request.get("task") != "suggest_flowsheet_edits":
        add_violation(violations, f"{sample_name}: input_request.task must be 'suggest_flowsheet_edits'")
    if len(primary_artifacts) != 1:
        add_violation(violations, f"{sample_name}: input_request must contain exactly one primary artifact")
    else:
        primary = primary_artifacts[0]
        if str(primary.get("name") or "") != "flowsheet_document":
            add_violation(violations, f"{sample_name}: primary artifact name must be 'flowsheet_document'")
        if str(primary.get("kind") or "") != "json":
            add_violation(violations, f"{sample_name}: primary artifact kind must be 'json'")
        if str(primary.get("mime_type") or "") != "application/json":
            add_violation(violations, f"{sample_name}: primary artifact mime_type must be 'application/json'")

    if context.get("document_revision") is None:
        add_violation(violations, f"{sample_name}: context.document_revision is required")
    if len(diagnostics) == 0:
        add_violation(violations, f"{sample_name}: suggest_flowsheet_edits samples must include diagnostics")
    if len(selected_unit_ids) == 0 and len(selected_stream_ids) == 0:
        add_violation(violations, f"{sample_name}: request must include selected_unit_ids or selected_stream_ids")


def validate_suggest_response(
    sample: dict[str, Any],
    response: dict[str, Any],
    response_label: str,
    sample_name: str,
    violations: list[str],
) -> None:
    shape = sample["expected_response_shape"]
    evaluation = sample["evaluation"]
    request_context = sample["input_request"].get("context") or {}
    answers = get_array(response.get("answers"))
    issues = get_array(response.get("issues"))
    actions = get_array(response.get("proposed_actions"))
    citations = get_array(response.get("citations"))
    selected_target_ids = {
        str(target_id)
        for target_id in (
            get_array(request_context.get("selected_unit_ids"))
            + get_array(request_context.get("selected_stream_ids"))
        )
        if str(target_id).strip()
    }
    diagnostic_target_ids = {
        str((diagnostic or {}).get("target_id") or "").strip()
        for diagnostic in get_array(request_context.get("diagnostics"))
        if str((diagnostic or {}).get("target_id") or "").strip()
    }
    allowed_target_ids = selected_target_ids | diagnostic_target_ids

    if response.get("project") != "radishflow":
        add_violation(violations, f"{sample_name}: {response_label}.project must be 'radishflow'")
    if response.get("task") != "suggest_flowsheet_edits":
        add_violation(violations, f"{sample_name}: {response_label}.task must be 'suggest_flowsheet_edits'")
    if str(response.get("status")) != str(shape.get("status")):
        add_violation(violations, f"{sample_name}: {response_label}.status does not match expected_response_shape.status")
    if str(response.get("risk_level")) != str(evaluation.get("expected_risk_level")):
        add_violation(violations, f"{sample_name}: {response_label}.risk_level does not match evaluation.expected_risk_level")
    if shape.get("requires_summary") and not str(response.get("summary") or "").strip():
        add_violation(violations, f"{sample_name}: {response_label}.summary is required")
    if shape.get("requires_answers") and len(answers) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 answer")
    if shape.get("requires_issues") and len(issues) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 issue")
    if shape.get("requires_citations") and len(citations) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 citation")
    if len(actions) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 proposed_action")

    actual_action_kinds = {str(action.get("kind") or "") for action in actions}
    for required_kind in [str(item) for item in get_array(shape.get("required_action_kinds"))]:
        if required_kind not in actual_action_kinds:
            add_violation(violations, f"{sample_name}: {response_label} is missing required action kind '{required_kind}'")

    highest_action_risk = 0
    candidate_edit_count = 0
    for action in actions:
        highest_action_risk = max(highest_action_risk, RISK_RANKS.get(str(action.get("risk_level") or ""), 0))
        if str(action.get("kind") or "") != "candidate_edit":
            add_violation(violations, f"{sample_name}: {response_label} actions must remain candidate_edit for this task")
            continue

        candidate_edit_count += 1
        target = action.get("target")
        if target is None:
            add_violation(violations, f"{sample_name}: {response_label} candidate_edit must include target")
        else:
            target_type = str((target or {}).get("type") or "").strip()
            target_id = str((target or {}).get("id") or "").strip()
            if not target_type or not target_id:
                add_violation(violations, f"{sample_name}: {response_label} candidate_edit target must include non-empty type and id")
            elif allowed_target_ids and target_id not in allowed_target_ids:
                add_violation(
                    violations,
                    f"{sample_name}: {response_label} candidate_edit target '{target_id}' must stay within selected or diagnosed objects",
                )

        patch = action.get("patch")
        if patch is None:
            add_violation(violations, f"{sample_name}: {response_label} candidate_edit must include patch")
        elif not isinstance(patch, dict) or len(patch) < 1:
            add_violation(violations, f"{sample_name}: {response_label} candidate_edit patch must not be empty")
        else:
            for patch_key in sorted(str(key) for key in patch.keys()):
                if patch_key in DISALLOWED_SUGGEST_PATCH_KEYS:
                    add_violation(
                        violations,
                        f"{sample_name}: {response_label} candidate_edit patch must remain reviewable and cannot use '{patch_key}'",
                    )

        if action.get("requires_confirmation") is not True:
            add_violation(violations, f"{sample_name}: {response_label} candidate_edit must set requires_confirmation=true")

    if candidate_edit_count < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 candidate_edit")
    if response.get("requires_confirmation") is not True:
        add_violation(violations, f"{sample_name}: {response_label}.requires_confirmation must be true when proposed_actions exist")
    if highest_action_risk > 0 and RISK_RANKS.get(str(response.get("risk_level") or ""), 0) != highest_action_risk:
        add_violation(violations, f"{sample_name}: {response_label}.risk_level must equal the highest proposed_action risk")

    ordered_issue_codes = [str(code) for code in get_array(evaluation.get("ordered_issue_codes"))]
    if ordered_issue_codes:
        actual_issue_codes = [str((issue or {}).get("code") or "") for issue in issues]
        if len(actual_issue_codes) < len(ordered_issue_codes):
            add_violation(
                violations,
                f"{sample_name}: {response_label} must contain at least {len(ordered_issue_codes)} issues for evaluation.ordered_issue_codes",
            )
        elif actual_issue_codes[: len(ordered_issue_codes)] != ordered_issue_codes:
            add_violation(
                violations,
                f"{sample_name}: {response_label} issues must remain ordered as {ordered_issue_codes}",
            )

    ordered_citation_ids = [str(citation_id) for citation_id in get_array(evaluation.get("ordered_citation_ids"))]
    if ordered_citation_ids:
        actual_citation_ids = [str((citation or {}).get("id") or "") for citation in citations]
        if len(actual_citation_ids) < len(ordered_citation_ids):
            add_violation(
                violations,
                f"{sample_name}: {response_label} must contain at least {len(ordered_citation_ids)} citations for evaluation.ordered_citation_ids",
            )
        elif actual_citation_ids[: len(ordered_citation_ids)] != ordered_citation_ids:
            add_violation(
                violations,
                f"{sample_name}: {response_label} citations must remain ordered as {ordered_citation_ids}",
            )

    ordered_issue_citation_sequences = get_array(evaluation.get("ordered_issue_citation_sequences"))
    for ordered_issue_citations in ordered_issue_citation_sequences:
        issue_index = ordered_issue_citations.get("issue_index")
        if not isinstance(issue_index, int):
            add_violation(
                violations,
                f"{sample_name}: {response_label} evaluation.ordered_issue_citation_sequences.issue_index must be an integer",
            )
            continue
        if issue_index >= len(issues):
            add_violation(
                violations,
                f"{sample_name}: {response_label} is missing issue[{issue_index}] required by evaluation.ordered_issue_citation_sequences",
            )
            continue

        expected_values = [str(value) for value in get_array(ordered_issue_citations.get("values"))]
        actual_values = [str(value) for value in get_array(issues[issue_index].get("citation_ids"))]
        if len(actual_values) < len(expected_values):
            add_violation(
                violations,
                f"{sample_name}: {response_label} issues[{issue_index}].citation_ids must contain at least {len(expected_values)} items for ordered_issue_citation_sequences",
            )
            continue
        if actual_values[: len(expected_values)] != expected_values:
            add_violation(
                violations,
                f"{sample_name}: {response_label} issues[{issue_index}].citation_ids must remain ordered as {expected_values}",
            )

    ordered_action_targets = get_array(evaluation.get("ordered_action_targets"))
    for index, ordered_target in enumerate(ordered_action_targets):
        if index >= len(actions):
            add_violation(
                violations,
                f"{sample_name}: {response_label} is missing proposed_action at index {index} required by evaluation.ordered_action_targets",
            )
            continue
        actual_target = actions[index].get("target") or {}
        expected_type = str((ordered_target or {}).get("type") or "").strip()
        expected_id = str((ordered_target or {}).get("id") or "").strip()
        actual_type = str((actual_target or {}).get("type") or "").strip()
        actual_id = str((actual_target or {}).get("id") or "").strip()
        if actual_type != expected_type or actual_id != expected_id:
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{index}] target must remain ordered as {expected_type}:{expected_id}",
            )

    ordered_action_citation_sequences = get_array(evaluation.get("ordered_action_citation_sequences"))
    for ordered_action_citations in ordered_action_citation_sequences:
        action_index = ordered_action_citations.get("action_index")
        if not isinstance(action_index, int):
            add_violation(
                violations,
                f"{sample_name}: {response_label} evaluation.ordered_action_citation_sequences.action_index must be an integer",
            )
            continue
        if action_index >= len(actions):
            add_violation(
                violations,
                f"{sample_name}: {response_label} is missing proposed_action[{action_index}] required by evaluation.ordered_action_citation_sequences",
            )
            continue

        expected_values = [str(value) for value in get_array(ordered_action_citations.get("values"))]
        actual_values = [str(value) for value in get_array(actions[action_index].get("citation_ids"))]
        if len(actual_values) < len(expected_values):
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].citation_ids must contain at least {len(expected_values)} items for ordered_action_citation_sequences",
            )
            continue
        if actual_values[: len(expected_values)] != expected_values:
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].citation_ids must remain ordered as {expected_values}",
            )

    ordered_patch_keys = get_array(evaluation.get("ordered_patch_keys"))
    for ordered_patch in ordered_patch_keys:
        action_index = ordered_patch.get("action_index")
        if not isinstance(action_index, int):
            add_violation(
                violations,
                f"{sample_name}: {response_label} evaluation.ordered_patch_keys.action_index must be an integer",
            )
            continue
        if action_index >= len(actions):
            add_violation(
                violations,
                f"{sample_name}: {response_label} is missing proposed_action[{action_index}] required by evaluation.ordered_patch_keys",
            )
            continue

        patch = actions[action_index].get("patch")
        if not isinstance(patch, dict):
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch must be present for ordered_patch_keys",
            )
            continue

        actual_keys = [str(key) for key in patch.keys()]
        expected_keys = [str(key) for key in get_array(ordered_patch.get("keys"))]
        if actual_keys[: len(expected_keys)] != expected_keys:
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch keys must remain ordered as {expected_keys}",
            )

    ordered_parameter_update_keys = get_array(evaluation.get("ordered_parameter_update_keys"))
    for ordered_parameter_update in ordered_parameter_update_keys:
        action_index = ordered_parameter_update.get("action_index")
        if not isinstance(action_index, int):
            add_violation(
                violations,
                f"{sample_name}: {response_label} evaluation.ordered_parameter_update_keys.action_index must be an integer",
            )
            continue
        if action_index >= len(actions):
            add_violation(
                violations,
                f"{sample_name}: {response_label} is missing proposed_action[{action_index}] required by evaluation.ordered_parameter_update_keys",
            )
            continue

        patch = actions[action_index].get("patch") or {}
        parameter_updates = patch.get("parameter_updates")
        if not isinstance(parameter_updates, dict):
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.parameter_updates must be present for ordered_parameter_update_keys",
            )
            continue

        actual_keys = [str(key) for key in parameter_updates.keys()]
        expected_keys = [str(key) for key in get_array(ordered_parameter_update.get("keys"))]
        if actual_keys[: len(expected_keys)] != expected_keys:
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.parameter_updates keys must remain ordered as {expected_keys}",
            )

    ordered_parameter_update_detail_keys = get_array(evaluation.get("ordered_parameter_update_detail_keys"))
    for ordered_parameter_update_detail in ordered_parameter_update_detail_keys:
        action_index = ordered_parameter_update_detail.get("action_index")
        if not isinstance(action_index, int):
            add_violation(
                violations,
                f"{sample_name}: {response_label} evaluation.ordered_parameter_update_detail_keys.action_index must be an integer",
            )
            continue
        if action_index >= len(actions):
            add_violation(
                violations,
                f"{sample_name}: {response_label} is missing proposed_action[{action_index}] required by evaluation.ordered_parameter_update_detail_keys",
            )
            continue

        parameter_key = str(ordered_parameter_update_detail.get("parameter_key") or "").strip()
        if not parameter_key:
            add_violation(
                violations,
                f"{sample_name}: {response_label} evaluation.ordered_parameter_update_detail_keys.parameter_key must be a non-empty string",
            )
            continue

        patch = actions[action_index].get("patch") or {}
        parameter_updates = patch.get("parameter_updates")
        if not isinstance(parameter_updates, dict):
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.parameter_updates must be present for ordered_parameter_update_detail_keys",
            )
            continue

        parameter_detail = parameter_updates.get(parameter_key)
        if not isinstance(parameter_detail, dict):
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.parameter_updates.{parameter_key} must be present for ordered_parameter_update_detail_keys",
            )
            continue

        actual_keys = [str(key) for key in parameter_detail.keys()]
        expected_keys = [str(key) for key in get_array(ordered_parameter_update_detail.get("keys"))]
        if actual_keys[: len(expected_keys)] != expected_keys:
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.parameter_updates.{parameter_key} keys must remain ordered as {expected_keys}",
            )

    ordered_parameter_update_value_sequences = get_array(evaluation.get("ordered_parameter_update_value_sequences"))
    for ordered_parameter_update_value in ordered_parameter_update_value_sequences:
        action_index = ordered_parameter_update_value.get("action_index")
        if not isinstance(action_index, int):
            add_violation(
                violations,
                f"{sample_name}: {response_label} evaluation.ordered_parameter_update_value_sequences.action_index must be an integer",
            )
            continue
        if action_index >= len(actions):
            add_violation(
                violations,
                f"{sample_name}: {response_label} is missing proposed_action[{action_index}] required by evaluation.ordered_parameter_update_value_sequences",
            )
            continue

        parameter_key = str(ordered_parameter_update_value.get("parameter_key") or "").strip()
        if not parameter_key:
            add_violation(
                violations,
                f"{sample_name}: {response_label} evaluation.ordered_parameter_update_value_sequences.parameter_key must be a non-empty string",
            )
            continue

        detail_key = str(ordered_parameter_update_value.get("detail_key") or "").strip()
        if not detail_key:
            add_violation(
                violations,
                f"{sample_name}: {response_label} evaluation.ordered_parameter_update_value_sequences.detail_key must be a non-empty string",
            )
            continue

        patch = actions[action_index].get("patch") or {}
        parameter_updates = patch.get("parameter_updates")
        if not isinstance(parameter_updates, dict):
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.parameter_updates must be present for ordered_parameter_update_value_sequences",
            )
            continue

        parameter_detail = parameter_updates.get(parameter_key)
        if not isinstance(parameter_detail, dict):
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.parameter_updates.{parameter_key} must be present for ordered_parameter_update_value_sequences",
            )
            continue

        actual_values = get_array(parameter_detail.get(detail_key))
        expected_values = get_array(ordered_parameter_update_value.get("values"))
        if len(actual_values) < len(expected_values):
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.parameter_updates.{parameter_key}.{detail_key} must contain at least {len(expected_values)} items for ordered_parameter_update_value_sequences",
            )
            continue
        if actual_values[: len(expected_values)] != expected_values:
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.parameter_updates.{parameter_key}.{detail_key} must remain ordered as {expected_values}",
            )

    ordered_spec_placeholder_sequences = get_array(evaluation.get("ordered_spec_placeholder_sequences"))
    for ordered_spec_placeholders in ordered_spec_placeholder_sequences:
        action_index = ordered_spec_placeholders.get("action_index")
        if not isinstance(action_index, int):
            add_violation(
                violations,
                f"{sample_name}: {response_label} evaluation.ordered_spec_placeholder_sequences.action_index must be an integer",
            )
            continue
        if action_index >= len(actions):
            add_violation(
                violations,
                f"{sample_name}: {response_label} is missing proposed_action[{action_index}] required by evaluation.ordered_spec_placeholder_sequences",
            )
            continue

        patch = actions[action_index].get("patch") or {}
        spec_placeholders = get_array(patch.get("spec_placeholders"))
        expected_values = [str(value) for value in get_array(ordered_spec_placeholders.get("values"))]
        if len(spec_placeholders) < len(expected_values):
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.spec_placeholders must contain at least {len(expected_values)} items for ordered_spec_placeholder_sequences",
            )
            continue
        actual_values = [str(value) for value in spec_placeholders[: len(expected_values)]]
        if actual_values != expected_values:
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.spec_placeholders must remain ordered as {expected_values}",
            )

    ordered_parameter_placeholder_sequences = get_array(evaluation.get("ordered_parameter_placeholder_sequences"))
    for ordered_parameter_placeholders in ordered_parameter_placeholder_sequences:
        action_index = ordered_parameter_placeholders.get("action_index")
        if not isinstance(action_index, int):
            add_violation(
                violations,
                f"{sample_name}: {response_label} evaluation.ordered_parameter_placeholder_sequences.action_index must be an integer",
            )
            continue
        if action_index >= len(actions):
            add_violation(
                violations,
                f"{sample_name}: {response_label} is missing proposed_action[{action_index}] required by evaluation.ordered_parameter_placeholder_sequences",
            )
            continue

        patch = actions[action_index].get("patch") or {}
        parameter_placeholders = get_array(patch.get("parameter_placeholders"))
        expected_values = [str(value) for value in get_array(ordered_parameter_placeholders.get("values"))]
        if len(parameter_placeholders) < len(expected_values):
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.parameter_placeholders must contain at least {len(expected_values)} items for ordered_parameter_placeholder_sequences",
            )
            continue
        actual_values = [str(value) for value in parameter_placeholders[: len(expected_values)]]
        if actual_values != expected_values:
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.parameter_placeholders must remain ordered as {expected_values}",
            )

    ordered_connection_placeholder_keys = get_array(evaluation.get("ordered_connection_placeholder_keys"))
    for ordered_connection_placeholder in ordered_connection_placeholder_keys:
        action_index = ordered_connection_placeholder.get("action_index")
        if not isinstance(action_index, int):
            add_violation(
                violations,
                f"{sample_name}: {response_label} evaluation.ordered_connection_placeholder_keys.action_index must be an integer",
            )
            continue
        if action_index >= len(actions):
            add_violation(
                violations,
                f"{sample_name}: {response_label} is missing proposed_action[{action_index}] required by evaluation.ordered_connection_placeholder_keys",
            )
            continue

        patch = actions[action_index].get("patch") or {}
        connection_placeholder = patch.get("connection_placeholder")
        if not isinstance(connection_placeholder, dict):
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.connection_placeholder must be present for ordered_connection_placeholder_keys",
            )
            continue

        actual_keys = [str(key) for key in connection_placeholder.keys()]
        expected_keys = [str(key) for key in get_array(ordered_connection_placeholder.get("keys"))]
        if actual_keys[: len(expected_keys)] != expected_keys:
            add_violation(
                violations,
                f"{sample_name}: {response_label} proposed_action[{action_index}].patch.connection_placeholder keys must remain ordered as {expected_keys}",
            )

    citation_ids = {str(citation.get("id") or "") for citation in citations}
    referenced_ids: set[str] = set()
    for answer in answers:
        referenced_ids.update(str(item) for item in get_array(answer.get("citation_ids")))
    for issue in issues:
        referenced_ids.update(str(item) for item in get_array(issue.get("citation_ids")))
    for action in actions:
        referenced_ids.update(str(item) for item in get_array(action.get("citation_ids")))
    for citation_id in sorted(citation_id for citation_id in referenced_ids if citation_id):
        if citation_id not in citation_ids:
            add_violation(violations, f"{sample_name}: referenced citation id '{citation_id}' is missing from {response_label}.citations")

    test_path_expectations(response, get_array(evaluation.get("must_have_json_paths")), True, f"{sample_name}:{response_label}", violations)
    test_path_expectations(response, get_array(evaluation.get("must_not_have_json_paths")), False, f"{sample_name}:{response_label}", violations)


def validate_ghost_completion_request(sample: dict[str, Any], sample_name: str, violations: list[str]) -> None:
    request = sample["input_request"]
    context = request.get("context") or {}
    artifacts = get_array(request.get("artifacts"))
    primary_artifacts = [artifact for artifact in artifacts if artifact.get("role") == "primary"]
    selected_unit_ids = [str(unit_id).strip() for unit_id in get_array(context.get("selected_unit_ids")) if str(unit_id).strip()]
    selected_unit = context.get("selected_unit") or {}
    legal_candidates = get_array(context.get("legal_candidate_completions"))
    unconnected_ports = [str(port).strip() for port in get_array(context.get("unconnected_ports")) if str(port).strip()]
    missing_canonical_ports = [
        str(port).strip() for port in get_array(context.get("missing_canonical_ports")) if str(port).strip()
    ]
    cursor_context = context.get("cursor_context") or {}
    recent_actions = get_array(cursor_context.get("recent_actions"))

    if request.get("project") != "radishflow":
        add_violation(violations, f"{sample_name}: input_request.project must be 'radishflow'")
    if request.get("task") != "suggest_ghost_completion":
        add_violation(violations, f"{sample_name}: input_request.task must be 'suggest_ghost_completion'")
    if len(primary_artifacts) != 1:
        add_violation(violations, f"{sample_name}: input_request must contain exactly one primary artifact")
    else:
        primary = primary_artifacts[0]
        if str(primary.get("name") or "") != "flowsheet_document":
            add_violation(violations, f"{sample_name}: primary artifact name must be 'flowsheet_document'")
        if str(primary.get("kind") or "") != "json":
            add_violation(violations, f"{sample_name}: primary artifact kind must be 'json'")
        if str(primary.get("mime_type") or "") != "application/json":
            add_violation(violations, f"{sample_name}: primary artifact mime_type must be 'application/json'")

    if context.get("document_revision") is None:
        add_violation(violations, f"{sample_name}: context.document_revision is required")
    if len(selected_unit_ids) != 1:
        add_violation(violations, f"{sample_name}: suggest_ghost_completion samples must include exactly one selected unit")
    if selected_unit:
        if str(selected_unit.get("id") or "").strip() and str(selected_unit.get("id") or "").strip() not in selected_unit_ids:
            add_violation(violations, f"{sample_name}: context.selected_unit.id must match context.selected_unit_ids[0]")
        if not str(selected_unit.get("kind") or "").strip():
            add_violation(violations, f"{sample_name}: context.selected_unit.kind is required when selected_unit is present")
    if cursor_context and not isinstance(cursor_context, dict):
        add_violation(violations, f"{sample_name}: context.cursor_context must be an object when present")
    for recent_action in recent_actions:
        action_kind = str((recent_action or {}).get("kind") or "").strip()
        candidate_ref = str((recent_action or {}).get("candidate_ref") or "").strip()
        revision_key = RECENT_GHOST_ACTION_REVISION_KEYS.get(action_kind)
        if revision_key is None:
            add_violation(
                violations,
                f"{sample_name}: context.cursor_context.recent_actions kind must be one of {sorted(RECENT_GHOST_ACTION_REVISION_KEYS)}",
            )
        if not candidate_ref:
            add_violation(
                violations,
                f"{sample_name}: each context.cursor_context.recent_action must include candidate_ref",
            )
        if revision_key is not None:
            action_revision = (recent_action or {}).get(revision_key)
            if not isinstance(action_revision, int):
                add_violation(
                    violations,
                    f"{sample_name}: recent_action kind='{action_kind}' must include integer {revision_key}",
                )
            elif context.get("document_revision") is not None and action_revision >= int(context.get("document_revision")):
                add_violation(
                    violations,
                    f"{sample_name}: recent_action.{revision_key} must be earlier than context.document_revision",
                )

    if "legal_candidate_completions" not in context:
        add_violation(violations, f"{sample_name}: context.legal_candidate_completions is required")
    for candidate in legal_candidates:
        candidate_ref = str((candidate or {}).get("candidate_ref") or "").strip()
        ghost_kind = str((candidate or {}).get("ghost_kind") or "").strip()
        target_port_key = str((candidate or {}).get("target_port_key") or "").strip()
        target_unit_id = str((candidate or {}).get("target_unit_id") or "").strip()
        conflict_flags = [str(flag).strip() for flag in get_array((candidate or {}).get("conflict_flags")) if str(flag).strip()]
        ranking_signals = candidate.get("ranking_signals")
        naming_signals = candidate.get("naming_signals")
        is_high_confidence = candidate.get("is_high_confidence")
        is_tab_default = candidate.get("is_tab_default")
        if not candidate_ref:
            add_violation(violations, f"{sample_name}: each legal_candidate_completion must include candidate_ref")
        if not ghost_kind:
            add_violation(violations, f"{sample_name}: each legal_candidate_completion must include ghost_kind")
        if not target_port_key:
            add_violation(violations, f"{sample_name}: each legal_candidate_completion must include target_port_key")
        if target_unit_id and selected_unit_ids and target_unit_id not in selected_unit_ids:
            add_violation(
                violations,
                f"{sample_name}: legal_candidate_completion.target_unit_id must stay within context.selected_unit_ids",
            )
        if is_tab_default is True and is_high_confidence is not True:
            add_violation(
                violations,
                f"{sample_name}: legal_candidate_completion '{candidate_ref}' cannot set is_tab_default=true without is_high_confidence=true",
            )
        if is_tab_default is True and len(conflict_flags) > 0:
            add_violation(
                violations,
                f"{sample_name}: legal_candidate_completion '{candidate_ref}' cannot set is_tab_default=true when conflict_flags are present",
            )
        if ranking_signals is not None and not isinstance(ranking_signals, dict):
            add_violation(violations, f"{sample_name}: legal_candidate_completion '{candidate_ref}' ranking_signals must be an object")
        if naming_signals is not None and not isinstance(naming_signals, dict):
            add_violation(violations, f"{sample_name}: legal_candidate_completion '{candidate_ref}' naming_signals must be an object")

    if len(unconnected_ports) == 0 and len(missing_canonical_ports) == 0:
        add_violation(
            violations,
            f"{sample_name}: request must include context.unconnected_ports or context.missing_canonical_ports",
        )


def validate_ghost_completion_response(
    sample: dict[str, Any],
    response: dict[str, Any],
    response_label: str,
    sample_name: str,
    violations: list[str],
) -> None:
    shape = sample["expected_response_shape"]
    evaluation = sample["evaluation"]
    request_context = sample["input_request"].get("context") or {}
    answers = get_array(response.get("answers"))
    issues = get_array(response.get("issues"))
    actions = get_array(response.get("proposed_actions"))
    citations = get_array(response.get("citations"))
    selected_unit_ids = {
        str(unit_id).strip() for unit_id in get_array(request_context.get("selected_unit_ids")) if str(unit_id).strip()
    }
    legal_candidates = {
        str((candidate or {}).get("candidate_ref") or "").strip(): candidate
        for candidate in get_array(request_context.get("legal_candidate_completions"))
        if str((candidate or {}).get("candidate_ref") or "").strip()
    }
    recent_actions = get_array((request_context.get("cursor_context") or {}).get("recent_actions"))

    if response.get("project") != "radishflow":
        add_violation(violations, f"{sample_name}: {response_label}.project must be 'radishflow'")
    if response.get("task") != "suggest_ghost_completion":
        add_violation(violations, f"{sample_name}: {response_label}.task must be 'suggest_ghost_completion'")
    if str(response.get("status")) != str(shape.get("status")):
        add_violation(violations, f"{sample_name}: {response_label}.status does not match expected_response_shape.status")
    if str(response.get("risk_level")) != str(evaluation.get("expected_risk_level")):
        add_violation(violations, f"{sample_name}: {response_label}.risk_level does not match evaluation.expected_risk_level")
    if shape.get("requires_summary") and not str(response.get("summary") or "").strip():
        add_violation(violations, f"{sample_name}: {response_label}.summary is required")
    if shape.get("requires_answers") and len(answers) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 answer")
    if shape.get("requires_issues") and len(issues) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 issue")
    if shape.get("requires_citations") and len(citations) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 citation")
    if not shape.get("allow_proposed_actions") and len(actions) > 0:
        add_violation(violations, f"{sample_name}: {response_label} should not contain proposed_actions")
    if len(actions) > 3:
        add_violation(violations, f"{sample_name}: {response_label} must not contain more than 3 ghost_completion actions")

    actual_action_kinds = {str(action.get("kind") or "") for action in actions}
    for required_kind in [str(item) for item in get_array(shape.get("required_action_kinds"))]:
        if required_kind not in actual_action_kinds:
            add_violation(violations, f"{sample_name}: {response_label} is missing required action kind '{required_kind}'")

    suppressed_candidate_refs = get_immediately_suppressed_candidate_refs(request_context)
    highest_action_risk = 0
    for index, action in enumerate(actions):
        highest_action_risk = max(highest_action_risk, RISK_RANKS.get(str(action.get("risk_level") or ""), 0))
        if str(action.get("kind") or "") != "ghost_completion":
            add_violation(violations, f"{sample_name}: {response_label} actions must remain ghost_completion for this task")
            continue

        target = action.get("target")
        if not isinstance(target, dict):
            add_violation(violations, f"{sample_name}: {response_label} ghost_completion must include target")
        else:
            target_type = str(target.get("type") or "").strip()
            target_unit_id = str(target.get("unit_id") or target.get("id") or "").strip()
            if not target_type:
                add_violation(violations, f"{sample_name}: {response_label} ghost_completion target.type is required")
            if selected_unit_ids and target_unit_id and target_unit_id not in selected_unit_ids:
                add_violation(
                    violations,
                    f"{sample_name}: {response_label} ghost_completion target must stay within context.selected_unit_ids",
                )
            if target_type == "unit_port" and not str(target.get("port_key") or "").strip():
                add_violation(violations, f"{sample_name}: {response_label} unit_port target must include port_key")

        patch = action.get("patch")
        candidate_ref = ""
        if not isinstance(patch, dict) or len(patch) < 1:
            add_violation(violations, f"{sample_name}: {response_label} ghost_completion patch must not be empty")
        else:
            if str(patch.get("ghost_kind") or "").strip() == "":
                add_violation(violations, f"{sample_name}: {response_label} ghost_completion patch.ghost_kind is required")
            candidate_ref = str(patch.get("candidate_ref") or "").strip()
            if not candidate_ref:
                add_violation(violations, f"{sample_name}: {response_label} ghost_completion patch.candidate_ref is required")
            elif candidate_ref not in legal_candidates:
                add_violation(
                    violations,
                    f"{sample_name}: {response_label} ghost_completion patch.candidate_ref must come from context.legal_candidate_completions",
                )
            else:
                selected_candidate = legal_candidates[candidate_ref] or {}
                candidate_target_port = str(selected_candidate.get("target_port_key") or "").strip()
                patch_target_port = str(patch.get("target_port_key") or "").strip()
                if patch_target_port and candidate_target_port and patch_target_port != candidate_target_port:
                    add_violation(
                        violations,
                        f"{sample_name}: {response_label} patch.target_port_key must match the selected legal candidate",
                    )
                if str(patch.get("ghost_stream_name") or "").strip():
                    naming_signals = selected_candidate.get("naming_signals")
                    if str((selected_candidate or {}).get("ghost_kind") or "").strip() == "ghost_stream_name" and not isinstance(
                        naming_signals, dict
                    ):
                        add_violation(
                            violations,
                            f"{sample_name}: {response_label} ghost_stream_name candidate '{candidate_ref}' should include naming_signals",
                        )

        preview = action.get("preview")
        if not isinstance(preview, dict):
            add_violation(violations, f"{sample_name}: {response_label} ghost_completion preview is required")
        else:
            if not str(preview.get("ghost_color") or "").strip():
                add_violation(violations, f"{sample_name}: {response_label} ghost_completion preview.ghost_color is required")
            accept_key = str(preview.get("accept_key") or "").strip()
            if not accept_key:
                add_violation(violations, f"{sample_name}: {response_label} ghost_completion preview.accept_key is required")
            if preview.get("render_priority") is None:
                add_violation(violations, f"{sample_name}: {response_label} ghost_completion preview.render_priority is required")
            if accept_key == "Tab" and index > 0:
                add_violation(
                    violations,
                    f"{sample_name}: {response_label} only the first ghost_completion may claim the default Tab accept key",
                )
            if accept_key == "Tab" and candidate_ref:
                selected_candidate = legal_candidates.get(candidate_ref) or {}
                if selected_candidate.get("is_tab_default") is not True:
                    add_violation(
                        violations,
                        f"{sample_name}: {response_label} Tab accept key must map to a legal candidate marked is_tab_default=true",
                    )
                if selected_candidate.get("is_high_confidence") is not True:
                    add_violation(
                        violations,
                        f"{sample_name}: {response_label} Tab accept key must map to a high-confidence legal candidate",
                    )
                conflict_flags = [
                    str(flag).strip() for flag in get_array(selected_candidate.get("conflict_flags")) if str(flag).strip()
                ]
                if len(conflict_flags) > 0:
                    add_violation(
                        violations,
                        f"{sample_name}: {response_label} Tab accept key must not map to a legal candidate with conflict_flags",
                    )
                if candidate_ref in suppressed_candidate_refs:
                    add_violation(
                        violations,
                        f"{sample_name}: {response_label} Tab accept key must not immediately reuse a candidate recently rejected, dismissed, or skipped in cursor_context.recent_actions",
                    )

        apply_payload = action.get("apply")
        if not isinstance(apply_payload, dict):
            add_violation(violations, f"{sample_name}: {response_label} ghost_completion apply is required")
        else:
            if str(apply_payload.get("command_kind") or "").strip() != "accept_ghost_completion":
                add_violation(
                    violations,
                    f"{sample_name}: {response_label} ghost_completion apply.command_kind must be 'accept_ghost_completion'",
                )
            payload = apply_payload.get("payload")
            if not isinstance(payload, dict):
                add_violation(violations, f"{sample_name}: {response_label} ghost_completion apply.payload must be an object")
            elif candidate_ref and str(payload.get("candidate_ref") or "").strip() != candidate_ref:
                add_violation(
                    violations,
                    f"{sample_name}: {response_label} ghost_completion apply.payload.candidate_ref must match patch.candidate_ref",
                )

        if str(action.get("risk_level") or "") == "high":
            add_violation(violations, f"{sample_name}: {response_label} ghost_completion must not escalate to high risk")
        if action.get("requires_confirmation") is not False:
            add_violation(violations, f"{sample_name}: {response_label} ghost_completion must set requires_confirmation=false")

    if response.get("requires_confirmation") is not False:
        add_violation(violations, f"{sample_name}: {response_label}.requires_confirmation must remain false for pending ghost suggestions")
    legal_candidate_list = get_array(request_context.get("legal_candidate_completions"))
    has_accept_chain_action = any(
        str((recent_action or {}).get("kind") or "").strip() == "accept_ghost_completion"
        for recent_action in recent_actions
    )
    if has_accept_chain_action and legal_candidate_list and len(actions) == 0:
        add_violation(
            violations,
            f"{sample_name}: {response_label} should not drop to zero proposed_actions when recent_actions indicate an accepted ghost chain step",
        )
    if highest_action_risk > 0 and RISK_RANKS.get(str(response.get("risk_level") or ""), 0) != highest_action_risk:
        add_violation(violations, f"{sample_name}: {response_label}.risk_level must equal the highest proposed_action risk")

    citation_ids = {str(citation.get("id") or "") for citation in citations}
    referenced_ids: set[str] = set()
    for answer in answers:
        referenced_ids.update(str(item) for item in get_array(answer.get("citation_ids")))
    for issue in issues:
        referenced_ids.update(str(item) for item in get_array(issue.get("citation_ids")))
    for action in actions:
        referenced_ids.update(str(item) for item in get_array(action.get("citation_ids")))
    for citation_id in sorted(citation_id for citation_id in referenced_ids if citation_id):
        if citation_id not in citation_ids:
            add_violation(violations, f"{sample_name}: referenced citation id '{citation_id}' is missing from {response_label}.citations")

    test_path_expectations(response, get_array(evaluation.get("must_have_json_paths")), True, f"{sample_name}:{response_label}", violations)
    test_path_expectations(response, get_array(evaluation.get("must_not_have_json_paths")), False, f"{sample_name}:{response_label}", violations)


def validate_control_plane_request(sample: dict[str, Any], sample_name: str, violations: list[str]) -> None:
    request = sample["input_request"]
    context = request.get("context") or {}
    control_plane_state = context.get("control_plane_state") or {}
    artifacts = get_array(request.get("artifacts"))

    if request.get("project") != "radishflow":
        add_violation(violations, f"{sample_name}: input_request.project must be 'radishflow'")
    if request.get("task") != "explain_control_plane_state":
        add_violation(violations, f"{sample_name}: input_request.task must be 'explain_control_plane_state'")
    if not isinstance(control_plane_state, dict) or len(control_plane_state) == 0:
        add_violation(violations, f"{sample_name}: context.control_plane_state is required")
        return

    if context.get("document_revision") is None and context.get("latest_snapshot") is None:
        add_violation(violations, f"{sample_name}: context.document_revision or context.latest_snapshot is required")

    has_any_state = any(
        str(control_plane_state.get(key) or "").strip()
        for key in ("entitlement_status", "lease_status", "sync_status", "manifest_status", "last_error")
    )
    if not has_any_state:
        add_violation(violations, f"{sample_name}: control_plane_state must include at least one non-empty status field")

    for artifact in artifacts:
        if not str(artifact.get("name") or "").strip():
            add_violation(violations, f"{sample_name}: each artifact must include name")
        if not str(artifact.get("kind") or "").strip():
            add_violation(violations, f"{sample_name}: each artifact must include kind")


def validate_control_plane_response(
    sample: dict[str, Any],
    response: dict[str, Any],
    response_label: str,
    sample_name: str,
    violations: list[str],
) -> None:
    shape = sample["expected_response_shape"]
    evaluation = sample["evaluation"]
    answers = get_array(response.get("answers"))
    issues = get_array(response.get("issues"))
    actions = get_array(response.get("proposed_actions"))
    citations = get_array(response.get("citations"))

    if response.get("project") != "radishflow":
        add_violation(violations, f"{sample_name}: {response_label}.project must be 'radishflow'")
    if response.get("task") != "explain_control_plane_state":
        add_violation(violations, f"{sample_name}: {response_label}.task must be 'explain_control_plane_state'")
    if str(response.get("status")) != str(shape.get("status")):
        add_violation(violations, f"{sample_name}: {response_label}.status does not match expected_response_shape.status")
    if str(response.get("risk_level")) != str(evaluation.get("expected_risk_level")):
        add_violation(violations, f"{sample_name}: {response_label}.risk_level does not match evaluation.expected_risk_level")
    if shape.get("requires_summary") and not str(response.get("summary") or "").strip():
        add_violation(violations, f"{sample_name}: {response_label}.summary is required")
    if shape.get("requires_answers") and len(answers) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 answer")
    if shape.get("requires_issues") and len(issues) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 issue")
    if shape.get("requires_citations") and len(citations) < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 citation")
    if not shape.get("allow_proposed_actions") and len(actions) > 0:
        add_violation(violations, f"{sample_name}: {response_label} should not contain proposed_actions")

    actual_answer_kinds = {str(answer.get("kind") or "") for answer in answers}
    for required_answer_kind in [str(item) for item in get_array(shape.get("required_answer_kinds"))]:
        if required_answer_kind not in actual_answer_kinds:
            add_violation(violations, f"{sample_name}: {response_label} is missing required answer kind '{required_answer_kind}'")

    if "hypothesis_labeling" in {str(item) for item in get_array(evaluation.get("scoring_focus"))}:
        hypothesis_answers = [
            answer
            for answer in answers
            if str(answer.get("kind") or "") in {"cause_hypothesis", "conflict_explanation"}
        ]
        if len(hypothesis_answers) == 0:
            add_violation(
                violations,
                f"{sample_name}: {response_label} must contain cause_hypothesis or conflict_explanation for hypothesis_labeling samples",
            )
        elif not any(contains_uncertainty_marker(answer.get("text")) for answer in hypothesis_answers):
            add_violation(
                violations,
                f"{sample_name}: {response_label} hypothesis answer must explicitly mark uncertainty for control-plane hypothesis_labeling samples",
            )

    actual_action_kinds = {str(action.get("kind") or "") for action in actions}
    for required_kind in [str(item) for item in get_array(shape.get("required_action_kinds"))]:
        if required_kind not in actual_action_kinds:
            add_violation(violations, f"{sample_name}: {response_label} is missing required action kind '{required_kind}'")

    highest_action_risk = 0
    any_confirmation = False
    for action in actions:
        kind = str(action.get("kind") or "")
        highest_action_risk = max(highest_action_risk, RISK_RANKS.get(str(action.get("risk_level") or ""), 0))
        any_confirmation = any_confirmation or (action.get("requires_confirmation") is True)

        if kind not in {"read_only_check", "candidate_operation"}:
            add_violation(violations, f"{sample_name}: {response_label} actions must be read_only_check or candidate_operation")
            continue

        target = action.get("target")
        if target is not None:
            if not str((target or {}).get("type") or "").strip() or not str((target or {}).get("id") or "").strip():
                add_violation(violations, f"{sample_name}: {response_label} action target must include non-empty type and id")

        if kind == "read_only_check":
            if str(action.get("risk_level") or "") != "low":
                add_violation(violations, f"{sample_name}: {response_label} read_only_check must remain low risk")
            if action.get("requires_confirmation") is not False:
                add_violation(violations, f"{sample_name}: {response_label} read_only_check must not require confirmation")

        if kind == "candidate_operation" and action.get("requires_confirmation") is not True:
            add_violation(violations, f"{sample_name}: {response_label} candidate_operation must set requires_confirmation=true")

    if highest_action_risk > 0 and RISK_RANKS.get(str(response.get("risk_level") or ""), 0) != highest_action_risk:
        add_violation(violations, f"{sample_name}: {response_label}.risk_level must equal the highest proposed_action risk")
    if any_confirmation and response.get("requires_confirmation") is not True:
        add_violation(violations, f"{sample_name}: {response_label}.requires_confirmation must be true when any action requires confirmation")
    if not any_confirmation and len(actions) > 0 and response.get("requires_confirmation") is not False:
        add_violation(violations, f"{sample_name}: {response_label}.requires_confirmation must stay false for read-only control-plane actions")

    citation_ids = {str(citation.get("id") or "") for citation in citations}
    referenced_ids: set[str] = set()
    for answer in answers:
        referenced_ids.update(str(item) for item in get_array(answer.get("citation_ids")))
    for issue in issues:
        referenced_ids.update(str(item) for item in get_array(issue.get("citation_ids")))
    for action in actions:
        referenced_ids.update(str(item) for item in get_array(action.get("citation_ids")))
    for citation_id in sorted(citation_id for citation_id in referenced_ids if citation_id):
        if citation_id not in citation_ids:
            add_violation(violations, f"{sample_name}: referenced citation id '{citation_id}' is missing from {response_label}.citations")

    test_path_expectations(response, get_array(evaluation.get("must_have_json_paths")), True, f"{sample_name}:{response_label}", violations)
    test_path_expectations(response, get_array(evaluation.get("must_not_have_json_paths")), False, f"{sample_name}:{response_label}", violations)


def collect_sample_files(sample_dir: Path, sample_paths: list[Path]) -> list[Path]:
    if sample_paths:
        files: list[Path] = []
        for path in sample_paths:
            if not path.is_file():
                raise SystemExit(f"missing sample file: {path}")
            files.append(path)
        return files

    if not sample_dir.is_dir():
        raise SystemExit(f"missing sample directory: {sample_dir}")
    files = sorted(sample_dir.glob("*.json"))
    if not files:
        raise SystemExit("no sample files found for regression")
    return files


def run_radish_docs_qa(config: dict[str, Any], sample_dir: Path, sample_paths: list[Path], fail_on_violation: bool) -> int:
    sample_files = collect_sample_files(sample_dir, sample_paths)
    all_violations: list[str] = []
    matched_sample_count = 0

    for sample_file in sample_files:
        sample_name = sample_file.name
        violations: list[str] = []
        try:
            sample = json.loads(sample_file.read_text(encoding="utf-8"))
        except Exception as exc:
            add_violation(violations, f"{sample_name}: failed to parse json: {exc}")
            sample = None

        if sample is not None and str(sample.get("task") or "") == config["sample_task"]:
            matched_sample_count += 1
            test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", violations)
            test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", violations)
            test_document_against_schema(sample.get("golden_response"), config["response_schema"], f"{sample_name} golden_response", violations)
            validate_radish_docs_retrieval(sample, sample_name, violations)
            validate_radish_docs_response(sample, sample["golden_response"], "golden_response", sample_name, violations)

            candidate_response, record_violations = load_candidate_response_from_record(sample, config, sample_name)
            violations.extend(record_violations)
            if candidate_response is not None:
                test_document_against_schema(candidate_response, config["response_schema"], f"{sample_name} candidate_response", violations)
                validate_radish_docs_response(sample, candidate_response, "candidate_response", sample_name, violations)
        elif sample is None:
            pass
        else:
            continue

        if violations:
            print(f"FAIL {sample_name}")
            for violation in violations:
                print(f"  - {violation}")
                all_violations.append(violation)
            continue

        print(f"PASS {sample_name}")

    if matched_sample_count == 0:
        raise SystemExit(config["no_sample_message"])
    if all_violations:
        if fail_on_violation:
            return 1
        print(f"WARNING: {config['warning_prefix']} {len(all_violations)} violation(s).", file=sys.stderr)
        return 0

    print(config["success_message"])
    return 0


def run_radish_docs_qa_negative(
    config: dict[str, Any],
    sample_dir: Path,
    sample_paths: list[Path],
    fail_on_violation: bool,
) -> int:
    sample_files = collect_sample_files(sample_dir, sample_paths)
    all_violations: list[str] = []
    matched_sample_count = 0

    for sample_file in sample_files:
        sample_name = sample_file.name
        violations: list[str] = []
        try:
            sample = json.loads(sample_file.read_text(encoding="utf-8"))
        except Exception as exc:
            add_violation(violations, f"{sample_name}: failed to parse json: {exc}")
            sample = None

        if sample is not None and str(sample.get("task") or "") == config["sample_task"]:
            matched_sample_count += 1
            test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", violations)
            test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", violations)
            test_document_against_schema(sample.get("golden_response"), config["response_schema"], f"{sample_name} golden_response", violations)
            validate_radish_docs_retrieval(sample, sample_name, violations)
            validate_radish_docs_response(sample, sample["golden_response"], "golden_response", sample_name, violations)
            validate_radish_docs_negative_replay(sample, config, sample_name, violations)
        elif sample is None:
            pass
        else:
            continue

        if violations:
            print(f"FAIL {sample_name}")
            for violation in violations:
                print(f"  - {violation}")
                all_violations.append(violation)
            continue

        print(f"PASS {sample_name}")

    if matched_sample_count == 0:
        raise SystemExit(config["no_sample_message"])
    if all_violations:
        if fail_on_violation:
            return 1
        print(f"WARNING: {config['warning_prefix']} {len(all_violations)} violation(s).", file=sys.stderr)
        return 0

    print(config["success_message"])
    return 0


def run_radishflow_diagnostics(config: dict[str, Any], sample_dir: Path, sample_paths: list[Path], fail_on_violation: bool) -> int:
    sample_files = collect_sample_files(sample_dir, sample_paths)
    all_violations: list[str] = []
    matched_sample_count = 0

    for sample_file in sample_files:
        sample_name = sample_file.name
        violations: list[str] = []
        try:
            sample = json.loads(sample_file.read_text(encoding="utf-8"))
        except Exception as exc:
            add_violation(violations, f"{sample_name}: failed to parse json: {exc}")
            sample = None

        if sample is not None and str(sample.get("task") or "") == config["sample_task"]:
            matched_sample_count += 1
            test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", violations)
            test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", violations)
            test_document_against_schema(sample.get("golden_response"), config["response_schema"], f"{sample_name} golden_response", violations)
            validate_diagnostics_request(sample, sample_name, violations)
            validate_diagnostics_response(sample, sample["golden_response"], "golden_response", sample_name, violations)

            candidate_response, record_violations = load_candidate_response_from_record(sample, config, sample_name)
            violations.extend(record_violations)
            if candidate_response is not None:
                test_document_against_schema(candidate_response, config["response_schema"], f"{sample_name} candidate_response", violations)
                validate_diagnostics_response(sample, candidate_response, "candidate_response", sample_name, violations)
        elif sample is None:
            pass
        else:
            continue

        if violations:
            print(f"FAIL {sample_name}")
            for violation in violations:
                print(f"  - {violation}")
                all_violations.append(violation)
            continue

        print(f"PASS {sample_name}")

    if matched_sample_count == 0:
        raise SystemExit(config["no_sample_message"])
    if all_violations:
        if fail_on_violation:
            return 1
        print(f"WARNING: {config['warning_prefix']} {len(all_violations)} violation(s).", file=sys.stderr)
        return 0

    print(config["success_message"])
    return 0


def run_radishflow_suggest_edits(config: dict[str, Any], sample_dir: Path, sample_paths: list[Path], fail_on_violation: bool) -> int:
    sample_files = collect_sample_files(sample_dir, sample_paths)
    all_violations: list[str] = []
    matched_sample_count = 0

    for sample_file in sample_files:
        sample_name = sample_file.name
        violations: list[str] = []
        try:
            sample = json.loads(sample_file.read_text(encoding="utf-8"))
        except Exception as exc:
            add_violation(violations, f"{sample_name}: failed to parse json: {exc}")
            sample = None

        if sample is not None and str(sample.get("task") or "") == config["sample_task"]:
            matched_sample_count += 1
            test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", violations)
            test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", violations)
            test_document_against_schema(sample.get("golden_response"), config["response_schema"], f"{sample_name} golden_response", violations)
            validate_suggest_request(sample, sample_name, violations)
            validate_suggest_response(sample, sample["golden_response"], "golden_response", sample_name, violations)

            candidate_response, record_violations = load_candidate_response_from_record(sample, config, sample_name)
            violations.extend(record_violations)
            if candidate_response is not None:
                test_document_against_schema(candidate_response, config["response_schema"], f"{sample_name} candidate_response", violations)
                validate_suggest_response(sample, candidate_response, "candidate_response", sample_name, violations)
        elif sample is None:
            pass
        else:
            continue

        if violations:
            print(f"FAIL {sample_name}")
            for violation in violations:
                print(f"  - {violation}")
                all_violations.append(violation)
            continue

        print(f"PASS {sample_name}")

    if matched_sample_count == 0:
        raise SystemExit(config["no_sample_message"])
    if all_violations:
        if fail_on_violation:
            return 1
        print(f"WARNING: {config['warning_prefix']} {len(all_violations)} violation(s).", file=sys.stderr)
        return 0

    print(config["success_message"])
    return 0


def run_radishflow_ghost_completion(
    config: dict[str, Any],
    sample_dir: Path,
    sample_paths: list[Path],
    fail_on_violation: bool,
) -> int:
    sample_files = collect_sample_files(sample_dir, sample_paths)
    all_violations: list[str] = []
    matched_sample_count = 0

    for sample_file in sample_files:
        sample_name = sample_file.name
        violations: list[str] = []
        try:
            sample = json.loads(sample_file.read_text(encoding="utf-8"))
        except Exception as exc:
            add_violation(violations, f"{sample_name}: failed to parse json: {exc}")
            sample = None

        if sample is not None and str(sample.get("task") or "") == config["sample_task"]:
            matched_sample_count += 1
            test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", violations)
            test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", violations)
            test_document_against_schema(sample.get("golden_response"), config["response_schema"], f"{sample_name} golden_response", violations)
            validate_ghost_completion_request(sample, sample_name, violations)
            validate_ghost_completion_response(sample, sample["golden_response"], "golden_response", sample_name, violations)

            candidate_response, record_violations = load_candidate_response_from_record(sample, config, sample_name)
            violations.extend(record_violations)
            if candidate_response is not None:
                test_document_against_schema(candidate_response, config["response_schema"], f"{sample_name} candidate_response", violations)
                validate_ghost_completion_response(sample, candidate_response, "candidate_response", sample_name, violations)
        elif sample is None:
            pass
        else:
            continue

        if violations:
            print(f"FAIL {sample_name}")
            for violation in violations:
                print(f"  - {violation}")
                all_violations.append(violation)
            continue

        print(f"PASS {sample_name}")

    if matched_sample_count == 0:
        raise SystemExit(config["no_sample_message"])
    if all_violations:
        if fail_on_violation:
            return 1
        print(f"WARNING: {config['warning_prefix']} {len(all_violations)} violation(s).", file=sys.stderr)
        return 0

    print(config["success_message"])
    return 0


def run_radishflow_control_plane(config: dict[str, Any], sample_dir: Path, sample_paths: list[Path], fail_on_violation: bool) -> int:
    sample_files = collect_sample_files(sample_dir, sample_paths)
    all_violations: list[str] = []
    matched_sample_count = 0

    for sample_file in sample_files:
        sample_name = sample_file.name
        violations: list[str] = []
        try:
            sample = json.loads(sample_file.read_text(encoding="utf-8"))
        except Exception as exc:
            add_violation(violations, f"{sample_name}: failed to parse json: {exc}")
            sample = None

        if sample is not None and str(sample.get("task") or "") == config["sample_task"]:
            matched_sample_count += 1
            test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", violations)
            test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", violations)
            test_document_against_schema(sample.get("golden_response"), config["response_schema"], f"{sample_name} golden_response", violations)
            validate_control_plane_request(sample, sample_name, violations)
            validate_control_plane_response(sample, sample["golden_response"], "golden_response", sample_name, violations)

            candidate_response, record_violations = load_candidate_response_from_record(sample, config, sample_name)
            violations.extend(record_violations)
            if candidate_response is not None:
                test_document_against_schema(candidate_response, config["response_schema"], f"{sample_name} candidate_response", violations)
                validate_control_plane_response(sample, candidate_response, "candidate_response", sample_name, violations)
        elif sample is None:
            pass
        else:
            continue

        if violations:
            print(f"FAIL {sample_name}")
            for violation in violations:
                print(f"  - {violation}")
                all_violations.append(violation)
            continue

        print(f"PASS {sample_name}")

    if matched_sample_count == 0:
        raise SystemExit(config["no_sample_message"])
    if all_violations:
        if fail_on_violation:
            return 1
        print(f"WARNING: {config['warning_prefix']} {len(all_violations)} violation(s).", file=sys.stderr)
        return 0

    print(config["success_message"])
    return 0


def parse_args(argv: list[str]) -> tuple[str, Path | None, list[Path], bool, dict[str, Any]]:
    if len(argv) < 2 or argv[1] not in TASK_CONFIG:
        task_names = ", ".join(sorted(TASK_CONFIG))
        raise SystemExit(f"usage: {Path(argv[0]).name} <task> [options]\navailable tasks: {task_names}")

    task_name = argv[1]
    sample_dir: Path | None = None
    sample_paths: list[Path] = []
    fail_on_violation = False
    negative_replay_index: Path | None = None
    batch_artifact_summary: Path | None = None
    group_ids: list[str] = []
    record_ids: list[str] = []
    replay_mode = ""
    recommended_groups_top: int | None = None
    index = 2

    while index < len(argv):
        arg = argv[index]
        if arg in {"-FailOnViolation", "--fail-on-violation"}:
            fail_on_violation = True
            index += 1
            continue
        if arg in {"-SampleDir", "--sample-dir"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            sample_dir = (REPO_ROOT / argv[index + 1]).resolve() if not Path(argv[index + 1]).is_absolute() else Path(argv[index + 1])
            index += 2
            continue
        if arg in {"-SamplePaths", "--sample-paths"}:
            index += 1
            if index >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            while index < len(argv) and not argv[index].startswith("-"):
                path = Path(argv[index])
                sample_paths.append((REPO_ROOT / path).resolve() if not path.is_absolute() else path)
                index += 1
            continue
        if arg in {"-NegativeReplayIndex", "--negative-replay-index"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            path = Path(argv[index + 1])
            negative_replay_index = (REPO_ROOT / path).resolve() if not path.is_absolute() else path
            index += 2
            continue
        if arg in {"-BatchArtifactSummary", "--batch-artifact-summary"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            path = Path(argv[index + 1])
            batch_artifact_summary = (REPO_ROOT / path).resolve() if not path.is_absolute() else path
            index += 2
            continue
        if arg in {"-GroupId", "--group-id"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            group_ids.append(argv[index + 1].strip())
            index += 2
            continue
        if arg in {"-RecordId", "--record-id"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            record_ids.append(argv[index + 1].strip())
            index += 2
            continue
        if arg in {"-ReplayMode", "--replay-mode"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            replay_mode = argv[index + 1].strip()
            index += 2
            continue
        if arg in {"-RecommendedGroupsTop", "--recommended-groups-top"}:
            if index + 1 >= len(argv):
                raise SystemExit(f"missing value for {arg}")
            try:
                recommended_groups_top = int(argv[index + 1])
            except ValueError as exc:
                raise SystemExit(f"invalid integer value for {arg}: {argv[index + 1]}") from exc
            index += 2
            continue
        raise SystemExit(f"unsupported argument: {arg}")

    selection = {
        "negative_replay_index": negative_replay_index,
        "batch_artifact_summary": batch_artifact_summary,
        "group_ids": [group_id for group_id in group_ids if group_id],
        "record_ids": [record_id for record_id in record_ids if record_id],
        "replay_mode": replay_mode,
        "recommended_groups_top": recommended_groups_top,
    }
    return task_name, sample_dir, sample_paths, fail_on_violation, selection


def ensure_required_paths(config: dict[str, Any]) -> None:
    for required_path in (
        config["sample_schema"],
        config.get("candidate_record_schema"),
        config["request_schema"],
        config["response_schema"],
        config["task_card"],
    ):
        if required_path is None:
            continue
        if not Path(required_path).is_file():
            raise SystemExit(f"missing required file: {required_path}")


def main(argv: list[str]) -> int:
    task_name, sample_dir, sample_paths, fail_on_violation, selection = parse_args(argv)
    config = TASK_CONFIG[task_name]
    ensure_required_paths(config)
    resolved_sample_dir = sample_dir or config["sample_dir"]
    negative_replay_index = selection["negative_replay_index"]
    batch_artifact_summary = selection["batch_artifact_summary"]
    recommended_groups_top = selection["recommended_groups_top"]
    if recommended_groups_top is not None:
        if task_name != "radish-docs-qa-negative":
            raise SystemExit("--recommended-groups-top is only supported for radish-docs-qa-negative")
        if batch_artifact_summary is None:
            raise SystemExit("--recommended-groups-top requires --batch-artifact-summary")
        if sample_paths:
            raise SystemExit("--recommended-groups-top cannot be used together with --sample-paths")
        if negative_replay_index is not None:
            raise SystemExit("--recommended-groups-top cannot be used together with --negative-replay-index")
        if selection["group_ids"]:
            raise SystemExit("--recommended-groups-top cannot be used together with --group-id")
        recommended_group_ids, default_replay_mode = resolve_recommended_negative_replay_groups_from_batch_artifact_summary(
            batch_artifact_summary,
            recommended_groups_top,
            selection["replay_mode"],
        )
        selection["group_ids"] = recommended_group_ids
        if not selection["replay_mode"]:
            selection["replay_mode"] = default_replay_mode
    if batch_artifact_summary is not None:
        if task_name != "radish-docs-qa-negative":
            raise SystemExit("--batch-artifact-summary is only supported for radish-docs-qa-negative")
        if sample_paths:
            raise SystemExit("--batch-artifact-summary cannot be used together with --sample-paths")
        if negative_replay_index is not None:
            raise SystemExit("--batch-artifact-summary cannot be used together with --negative-replay-index")
        negative_replay_index = resolve_negative_replay_index_from_batch_artifact_summary(
            batch_artifact_summary,
            selection["replay_mode"],
        )
    if negative_replay_index is not None:
        if task_name != "radish-docs-qa-negative":
            raise SystemExit("--negative-replay-index is only supported for radish-docs-qa-negative")
        if sample_paths:
            raise SystemExit("--negative-replay-index cannot be used together with --sample-paths")
        sample_paths = resolve_negative_replay_sample_paths(
            negative_replay_index,
            group_ids=selection["group_ids"],
            record_ids=selection["record_ids"],
            replay_mode=selection["replay_mode"],
        )

    if task_name == "radish-docs-qa":
        return run_radish_docs_qa(config, resolved_sample_dir, sample_paths, fail_on_violation)
    if task_name == "radish-docs-qa-negative":
        return run_radish_docs_qa_negative(config, resolved_sample_dir, sample_paths, fail_on_violation)
    if task_name == "radishflow-diagnostics":
        return run_radishflow_diagnostics(config, resolved_sample_dir, sample_paths, fail_on_violation)
    if task_name == "radishflow-ghost-completion":
        return run_radishflow_ghost_completion(config, resolved_sample_dir, sample_paths, fail_on_violation)
    if task_name == "radishflow-control-plane":
        return run_radishflow_control_plane(config, resolved_sample_dir, sample_paths, fail_on_violation)
    return run_radishflow_suggest_edits(config, resolved_sample_dir, sample_paths, fail_on_violation)


if __name__ == "__main__":
    sys.exit(main(sys.argv))
