from __future__ import annotations

from .regression_shared import *  # noqa: F401,F403


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

