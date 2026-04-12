from __future__ import annotations

from .regression_shared import *  # noqa: F401,F403


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

