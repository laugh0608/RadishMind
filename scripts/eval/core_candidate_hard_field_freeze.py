from __future__ import annotations

import copy
from typing import Any

from scripts.eval.core_candidate_paths import get_evaluation, parse_json_path_expected_value


TOP_LEVEL_HARD_FIELDS = (
    "schema_version",
    "project",
    "task",
    "status",
    "risk_level",
    "requires_confirmation",
)


def _expected_shape(sample: dict[str, Any]) -> dict[str, Any]:
    expected_shape = sample.get("expected_response_shape")
    return expected_shape if isinstance(expected_shape, dict) else {}


def _must_have_paths(sample: dict[str, Any]) -> list[str]:
    paths = get_evaluation(sample).get("must_have_json_paths")
    return [str(path) for path in paths] if isinstance(paths, list) else []


def _must_not_have_paths(sample: dict[str, Any]) -> set[str]:
    paths = get_evaluation(sample).get("must_not_have_json_paths")
    return {str(path) for path in paths} if isinstance(paths, list) else set()


def _explicit_path_values(sample: dict[str, Any]) -> dict[str, Any]:
    values: dict[str, Any] = {}
    for path in _must_have_paths(sample):
        if "=" not in path:
            continue
        raw_path, raw_value = path.split("=", 1)
        if raw_path.startswith("$."):
            values[raw_path] = parse_json_path_expected_value(raw_value)
    return values


def _has_explicit_path_or_child(explicit_paths: set[str], path: str) -> bool:
    return any(explicit_path == path or explicit_path.startswith(path + ".") or explicit_path.startswith(path + "[") for explicit_path in explicit_paths)


def _add_freeze_field(fields: list[dict[str, Any]], seen_paths: set[str], path: str, value: Any) -> None:
    if path in seen_paths:
        return
    seen_paths.add(path)
    fields.append({"path": path, "value": copy.deepcopy(value)})


def build_hard_field_freeze(sample: dict[str, Any], scaffold: dict[str, Any]) -> dict[str, Any]:
    expected_shape = _expected_shape(sample)
    must_not_have_paths = _must_not_have_paths(sample)
    explicit_values = _explicit_path_values(sample)
    explicit_paths = set(explicit_values)
    fields: list[dict[str, Any]] = []
    seen_paths: set[str] = set()

    for field in TOP_LEVEL_HARD_FIELDS:
        if field in scaffold:
            _add_freeze_field(fields, seen_paths, f"$.{field}", scaffold[field])

    citations = scaffold.get("citations")
    if (
        isinstance(citations, list)
        and citations
        and (
            isinstance(get_evaluation(sample).get("ordered_citation_ids"), list)
            or _has_explicit_path_or_child(explicit_paths, "$.citations")
        )
    ):
        _add_freeze_field(fields, seen_paths, "$.citations", citations)

    answers = scaffold.get("answers")
    if isinstance(answers, list) and answers:
        first_answer = answers[0]
        if isinstance(first_answer, dict):
            for field in ("kind", "citation_ids"):
                path = f"$.answers[0].{field}"
                if field in first_answer and _has_explicit_path_or_child(explicit_paths, path):
                    _add_freeze_field(fields, seen_paths, f"$.answers[0].{field}", first_answer[field])

    issues = scaffold.get("issues")
    ordered_issue_codes = get_evaluation(sample).get("ordered_issue_codes")
    if isinstance(issues, list) and issues:
        for index, issue in enumerate(issues):
            if not isinstance(issue, dict):
                continue
            for field in ("code", "severity", "citation_ids"):
                path = f"$.issues[{index}].{field}"
                freeze_issue_code = field == "code" and isinstance(ordered_issue_codes, list)
                if field in issue and (freeze_issue_code or _has_explicit_path_or_child(explicit_paths, path)):
                    _add_freeze_field(fields, seen_paths, f"$.issues[{index}].{field}", issue[field])

    actions = scaffold.get("proposed_actions")
    if expected_shape.get("allow_proposed_actions") is False or "$.proposed_actions[0]" in must_not_have_paths:
        _add_freeze_field(fields, seen_paths, "$.proposed_actions", [])
    elif isinstance(actions, list) and actions:
        required_action_kinds = expected_shape.get("required_action_kinds")
        for index, action in enumerate(actions):
            if not isinstance(action, dict):
                continue
            required_kind_value = (
                str(required_action_kinds[index])
                if isinstance(required_action_kinds, list) and index < len(required_action_kinds)
                else ""
            )
            for field in (
                "kind",
                "target",
                "patch",
                "preview",
                "apply",
                "risk_level",
                "requires_confirmation",
                "citation_ids",
            ):
                path = f"$.proposed_actions[{index}].{field}"
                freeze_required_action_kind = field == "kind" and bool(required_kind_value)
                freeze_read_only_boundary = (
                    action.get("kind") == "read_only_check"
                    and field == "requires_confirmation"
                    and required_kind_value == "read_only_check"
                )
                freeze_action_patch = field in {"target", "patch", "preview", "apply"} and _has_explicit_path_or_child(
                    explicit_paths,
                    path,
                )
                if field in action and (
                    freeze_required_action_kind
                    or freeze_read_only_boundary
                    or freeze_action_patch
                    or _has_explicit_path_or_child(explicit_paths, path)
                ):
                    _add_freeze_field(fields, seen_paths, f"$.proposed_actions[{index}].{field}", action[field])

    return {
        "schema_version": 1,
        "kind": "scaffold_derived_hard_field_freeze",
        "scope": (
            "Prompt-time contract only: candidate models must copy these path values exactly. "
            "This is not post-decode repair and does not change raw/repaired metric separation."
        ),
        "fields": fields,
    }
