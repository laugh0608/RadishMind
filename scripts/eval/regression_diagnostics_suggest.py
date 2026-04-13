from __future__ import annotations

from .regression_shared import *  # noqa: F401,F403


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

    test_ordered_citation_sequences(
        answers,
        evaluation.get("ordered_answer_citation_sequences"),
        response_label=response_label,
        sample_name=sample_name,
        violations=violations,
        field_name="ordered_answer_citation_sequences",
        index_key="answer_index",
        item_label="answers",
    )
    test_ordered_citation_sequences(
        issues,
        evaluation.get("ordered_issue_citation_sequences"),
        response_label=response_label,
        sample_name=sample_name,
        violations=violations,
        field_name="ordered_issue_citation_sequences",
        index_key="issue_index",
        item_label="issues",
    )
    test_ordered_citation_sequences(
        actions,
        evaluation.get("ordered_action_citation_sequences"),
        response_label=response_label,
        sample_name=sample_name,
        violations=violations,
        field_name="ordered_action_citation_sequences",
        index_key="action_index",
        item_label="proposed_action",
    )
    test_artifact_citation_expectations(
        sample,
        citations,
        response_label=response_label,
        sample_name=sample_name,
        violations=violations,
    )
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

