from __future__ import annotations

import copy
from typing import Any

from scripts.eval.core_candidate_paths import get_evaluation


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


def _has_path_prefix(paths: list[str], prefix: str) -> bool:
    return any(path == prefix or path.startswith(prefix + ".") or path.startswith(prefix + "[") for path in paths)


def _add_freeze_field(fields: list[dict[str, Any]], seen_paths: set[str], path: str, value: Any) -> None:
    if path in seen_paths:
        return
    seen_paths.add(path)
    fields.append({"path": path, "value": copy.deepcopy(value)})


def build_hard_field_freeze(sample: dict[str, Any], scaffold: dict[str, Any]) -> dict[str, Any]:
    expected_shape = _expected_shape(sample)
    must_have_paths = _must_have_paths(sample)
    must_not_have_paths = _must_not_have_paths(sample)
    fields: list[dict[str, Any]] = []
    seen_paths: set[str] = set()

    for field in TOP_LEVEL_HARD_FIELDS:
        if field in scaffold:
            _add_freeze_field(fields, seen_paths, f"$.{field}", scaffold[field])

    citations = scaffold.get("citations")
    if (
        isinstance(citations, list)
        and citations
        and (expected_shape.get("requires_citations") is True or _has_path_prefix(must_have_paths, "$.citations"))
    ):
        _add_freeze_field(fields, seen_paths, "$.citations", citations)

    answers = scaffold.get("answers")
    if (
        isinstance(answers, list)
        and answers
        and (expected_shape.get("requires_answers") is True or _has_path_prefix(must_have_paths, "$.answers"))
    ):
        first_answer = answers[0]
        if isinstance(first_answer, dict):
            for field in ("kind", "citation_ids"):
                if field in first_answer:
                    _add_freeze_field(fields, seen_paths, f"$.answers[0].{field}", first_answer[field])

    issues = scaffold.get("issues")
    ordered_issue_codes = get_evaluation(sample).get("ordered_issue_codes")
    if (
        isinstance(issues, list)
        and issues
        and (
            expected_shape.get("requires_issues") is True
            or isinstance(ordered_issue_codes, list)
            or _has_path_prefix(must_have_paths, "$.issues")
        )
    ):
        for index, issue in enumerate(issues):
            if not isinstance(issue, dict):
                continue
            for field in ("code", "severity", "citation_ids"):
                if field in issue:
                    _add_freeze_field(fields, seen_paths, f"$.issues[{index}].{field}", issue[field])

    actions = scaffold.get("proposed_actions")
    if expected_shape.get("allow_proposed_actions") is False or "$.proposed_actions[0]" in must_not_have_paths:
        _add_freeze_field(fields, seen_paths, "$.proposed_actions", [])
    elif isinstance(actions, list) and actions:
        for index, action in enumerate(actions):
            if not isinstance(action, dict):
                continue
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
                if field in action:
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
