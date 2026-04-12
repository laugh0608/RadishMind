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


REPO_ROOT = Path(__file__).resolve().parent.parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.inference import build_artifact_citation_fields  # noqa: E402

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


def test_ordered_citation_sequences(
    items: list[Any],
    expectations: Any,
    *,
    response_label: str,
    sample_name: str,
    violations: list[str],
    field_name: str,
    index_key: str,
    item_label: str,
) -> None:
    for ordered_citations in get_array(expectations):
        item_index = ordered_citations.get(index_key)
        if not isinstance(item_index, int):
            add_violation(
                violations,
                f"{sample_name}: {response_label} evaluation.{field_name}.{index_key} must be an integer",
            )
            continue
        if item_index < 0 or item_index >= len(items):
            add_violation(
                violations,
                f"{sample_name}: {response_label} is missing {item_label}[{item_index}] required by evaluation.{field_name}",
            )
            continue

        expected_values = [str(value) for value in get_array(ordered_citations.get("values"))]
        actual_values = [str(value) for value in get_array((items[item_index] or {}).get("citation_ids"))]
        if len(actual_values) < len(expected_values):
            add_violation(
                violations,
                f"{sample_name}: {response_label} {item_label}[{item_index}].citation_ids must contain at least "
                f"{len(expected_values)} items for {field_name}",
            )
            continue
        if actual_values[: len(expected_values)] != expected_values:
            add_violation(
                violations,
                f"{sample_name}: {response_label} {item_label}[{item_index}].citation_ids must remain ordered as {expected_values}",
            )


def test_artifact_citation_expectations(
    sample: dict[str, Any],
    citations: list[Any],
    *,
    response_label: str,
    sample_name: str,
    violations: list[str],
) -> None:
    evaluation = sample.get("evaluation") or {}
    expectations = get_array(evaluation.get("artifact_citation_expectations"))
    if not expectations:
        return

    request = sample.get("input_request") or {}
    request_context = request.get("context") or {}
    resource = request_context.get("resource") or {}
    resource_title = str(resource.get("title") or "").strip()
    request_artifacts = {
        str((artifact or {}).get("name") or "").strip(): artifact
        for artifact in get_array(request.get("artifacts"))
        if str((artifact or {}).get("name") or "").strip()
    }
    citations_by_id = {
        str((citation or {}).get("id") or "").strip(): citation
        for citation in citations
        if isinstance(citation, dict) and str((citation or {}).get("id") or "").strip()
    }

    for expectation in expectations:
        citation_id = str((expectation or {}).get("citation_id") or "").strip()
        artifact_name = str((expectation or {}).get("artifact_name") or "").strip()
        if not citation_id or not artifact_name:
            continue

        citation = citations_by_id.get(citation_id)
        if citation is None:
            add_violation(
                violations,
                f"{sample_name}: {response_label} is missing citation '{citation_id}' required by evaluation.artifact_citation_expectations",
            )
            continue

        artifact = request_artifacts.get(artifact_name)
        if artifact is None:
            add_violation(
                violations,
                f"{sample_name}: evaluation.artifact_citation_expectations references unknown input_request artifact '{artifact_name}'",
            )
            continue

        expected_fields = build_artifact_citation_fields(artifact, resource_title)
        if str(citation.get("kind") or "").strip() != "artifact":
            add_violation(
                violations,
                f"{sample_name}: {response_label} citation '{citation_id}' must remain kind='artifact' for artifact-backed expectations",
            )

        expected_label = expected_fields.get("label")
        actual_label = str(citation.get("label") or "").strip()
        if expected_label and actual_label != expected_label:
            add_violation(
                violations,
                f"{sample_name}: {response_label} citation '{citation_id}' label must remain '{expected_label}'",
            )

        expected_locator = expected_fields.get("locator")
        actual_locator = str(citation.get("locator") or "").strip()
        if expected_locator and actual_locator != expected_locator:
            add_violation(
                violations,
                f"{sample_name}: {response_label} citation '{citation_id}' locator must remain '{expected_locator}'",
            )

        expected_excerpt = expected_fields.get("excerpt")
        actual_excerpt = str(citation.get("excerpt") or "").strip()
        if expected_excerpt:
            if actual_excerpt != expected_excerpt:
                add_violation(
                    violations,
                    f"{sample_name}: {response_label} citation '{citation_id}' excerpt must remain canonicalized from input_request artifact '{artifact_name}'",
                )
        elif actual_excerpt:
            add_violation(
                violations,
                f"{sample_name}: {response_label} citation '{citation_id}' must not synthesize excerpt when artifact '{artifact_name}' has no canonical excerpt source",
            )

        for substring in [
            str(item)
            for item in get_array((expectation or {}).get("excerpt_must_not_contain"))
            if str(item).strip()
        ]:
            if substring in actual_excerpt:
                add_violation(
                    violations,
                    f"{sample_name}: {response_label} citation '{citation_id}' excerpt must not contain '{substring}'",
                )


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

