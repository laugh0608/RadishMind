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
SCHEMA_CACHE: dict[Path, Any] = {}


TASK_CONFIG = {
    "radish-docs-qa": {
        "sample_dir": REPO_ROOT / "datasets/eval/radish",
        "sample_schema": REPO_ROOT / "datasets/eval/radish-task-sample.schema.json",
        "request_schema": REPO_ROOT / "contracts/copilot-request.schema.json",
        "response_schema": REPO_ROOT / "contracts/copilot-response.schema.json",
        "task_card": REPO_ROOT / "docs/task-cards/radish-answer-docs-question.md",
        "sample_task": "answer_docs_question",
        "success_message": "radish docs qa regression passed.",
        "no_sample_message": "no answer_docs_question sample files found for Radish docs QA regression",
        "warning_prefix": "radish docs qa regression found",
    },
    "radishflow-diagnostics": {
        "sample_dir": REPO_ROOT / "datasets/eval/radishflow",
        "sample_schema": REPO_ROOT / "datasets/eval/radishflow-task-sample.schema.json",
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
        "request_schema": REPO_ROOT / "contracts/copilot-request.schema.json",
        "response_schema": REPO_ROOT / "contracts/copilot-response.schema.json",
        "task_card": REPO_ROOT / "docs/task-cards/radishflow-suggest-flowsheet-edits.md",
        "sample_task": "suggest_flowsheet_edits",
        "success_message": "radishflow suggest edits regression passed.",
        "no_sample_message": "no suggest_flowsheet_edits sample files found for RadishFlow suggest edits regression",
        "warning_prefix": "radishflow suggest edits regression found",
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


def validate_radish_docs_response(sample: dict[str, Any], sample_name: str, violations: list[str]) -> None:
    response = sample["golden_response"]
    shape = sample["expected_response_shape"]
    evaluation = sample["evaluation"]
    answers = get_array(response.get("answers"))
    issues = get_array(response.get("issues"))
    actions = get_array(response.get("proposed_actions"))
    citations = get_array(response.get("citations"))

    if response.get("project") != "radish":
        add_violation(violations, f"{sample_name}: golden_response.project must be 'radish'")
    if response.get("task") != "answer_docs_question":
        add_violation(violations, f"{sample_name}: golden_response.task must be 'answer_docs_question'")
    if str(response.get("status")) != str(shape.get("status")):
        add_violation(violations, f"{sample_name}: golden_response.status does not match expected_response_shape.status")
    if str(response.get("risk_level")) != str(evaluation.get("expected_risk_level")):
        add_violation(violations, f"{sample_name}: golden_response.risk_level does not match evaluation.expected_risk_level")
    if shape.get("requires_summary") and not str(response.get("summary") or "").strip():
        add_violation(violations, f"{sample_name}: golden_response.summary is required")
    if shape.get("requires_answers") and len(answers) < 1:
        add_violation(violations, f"{sample_name}: golden_response must contain at least 1 answer")
    if shape.get("requires_issues") and len(issues) < 1:
        add_violation(violations, f"{sample_name}: golden_response must contain at least 1 issue")
    if not shape.get("requires_issues") and len(issues) > 0:
        add_violation(violations, f"{sample_name}: golden_response should not contain issues")
    if shape.get("requires_citations") and len(citations) < 1:
        add_violation(violations, f"{sample_name}: golden_response must contain at least 1 citation")
    if not shape.get("allow_proposed_actions") and len(actions) > 0:
        add_violation(violations, f"{sample_name}: golden_response should not contain proposed_actions")

    actual_kinds = {str(action.get("kind") or "") for action in actions}
    for required_kind in [str(item) for item in get_array(shape.get("required_action_kinds"))]:
        if required_kind not in actual_kinds:
            add_violation(violations, f"{sample_name}: golden_response is missing required action kind '{required_kind}'")

    for action in actions:
        if str(action.get("risk_level") or "") != "low":
            add_violation(
                violations,
                f"{sample_name}: proposed action '{action.get('title')}' must remain low risk for answer_docs_question",
            )
        if action.get("requires_confirmation") is not False:
            add_violation(
                violations,
                f"{sample_name}: proposed action '{action.get('title')}' must not require confirmation",
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
            add_violation(violations, f"{sample_name}: referenced citation id '{citation_id}' is missing from golden_response.citations")

    test_path_expectations(response, get_array(evaluation.get("must_have_json_paths")), True, sample_name, violations)
    test_path_expectations(response, get_array(evaluation.get("must_not_have_json_paths")), False, sample_name, violations)


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
    answers = get_array(response.get("answers"))
    issues = get_array(response.get("issues"))
    actions = get_array(response.get("proposed_actions"))
    citations = get_array(response.get("citations"))

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
            if not str((target or {}).get("type") or "").strip() or not str((target or {}).get("id") or "").strip():
                add_violation(violations, f"{sample_name}: {response_label} candidate_edit target must include non-empty type and id")

        patch = action.get("patch")
        if patch is None:
            add_violation(violations, f"{sample_name}: {response_label} candidate_edit must include patch")
        elif not isinstance(patch, dict) or len(patch) < 1:
            add_violation(violations, f"{sample_name}: {response_label} candidate_edit patch must not be empty")

        if action.get("requires_confirmation") is not True:
            add_violation(violations, f"{sample_name}: {response_label} candidate_edit must set requires_confirmation=true")

    if candidate_edit_count < 1:
        add_violation(violations, f"{sample_name}: {response_label} must contain at least 1 candidate_edit")
    if response.get("requires_confirmation") is not True:
        add_violation(violations, f"{sample_name}: {response_label}.requires_confirmation must be true when proposed_actions exist")
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
            validate_radish_docs_response(sample, sample_name, violations)
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

            candidate_response = sample.get("candidate_response")
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

            candidate_response = sample.get("candidate_response")
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


def parse_args(argv: list[str]) -> tuple[str, Path | None, list[Path], bool]:
    if len(argv) < 2 or argv[1] not in TASK_CONFIG:
        task_names = ", ".join(sorted(TASK_CONFIG))
        raise SystemExit(f"usage: {Path(argv[0]).name} <task> [options]\navailable tasks: {task_names}")

    task_name = argv[1]
    sample_dir: Path | None = None
    sample_paths: list[Path] = []
    fail_on_violation = False
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
        raise SystemExit(f"unsupported argument: {arg}")

    return task_name, sample_dir, sample_paths, fail_on_violation


def ensure_required_paths(config: dict[str, Any]) -> None:
    for required_path in (
        config["sample_schema"],
        config["request_schema"],
        config["response_schema"],
        config["task_card"],
    ):
        if not Path(required_path).is_file():
            raise SystemExit(f"missing required file: {required_path}")


def main(argv: list[str]) -> int:
    task_name, sample_dir, sample_paths, fail_on_violation = parse_args(argv)
    config = TASK_CONFIG[task_name]
    ensure_required_paths(config)
    resolved_sample_dir = sample_dir or config["sample_dir"]

    if task_name == "radish-docs-qa":
        return run_radish_docs_qa(config, resolved_sample_dir, sample_paths, fail_on_violation)
    if task_name == "radishflow-diagnostics":
        return run_radishflow_diagnostics(config, resolved_sample_dir, sample_paths, fail_on_violation)
    return run_radishflow_suggest_edits(config, resolved_sample_dir, sample_paths, fail_on_violation)


if __name__ == "__main__":
    sys.exit(main(sys.argv))
