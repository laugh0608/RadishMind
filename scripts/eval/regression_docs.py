from __future__ import annotations

from .regression_shared import *  # noqa: F401,F403


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

