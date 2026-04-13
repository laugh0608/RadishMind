from __future__ import annotations

import json
import re
from typing import Any

import jsonschema

from .inference_support import (
    GHOST_MALFORMED_JSON_REPAIR_PATTERNS,
    GHOST_MANUAL_MULTI_ACTION_REPAIR_PATTERNS,
    RESPONSE_TOP_LEVEL_KEYS,
    RESPONSE_SCHEMA_PATH,
    build_artifact_citation_fields,
    build_citation_maps,
    infer_parameter_placeholders,
    infer_stream_spec_placeholders,
    load_schema,
    make_failed_response,
    normalize_text,
    validate_response_document,
)


def map_status(value: Any) -> str:
    lowered = str(value or "").strip().lower()
    if lowered in {"ok", "partial", "failed"}:
        return lowered
    if lowered in {"success", "done", "complete", "completed"}:
        return "ok"
    if lowered in {"warning", "degraded", "limited"}:
        return "partial"
    if lowered in {"error", "failure"}:
        return "failed"
    return "partial"


def normalize_citation_ids(
    value: Any,
    artifact_name_to_citation_id: dict[str, str],
    artifact_index_to_citation_id: dict[int, str],
) -> list[str]:
    normalized_ids: list[str] = []
    for item in value or []:
        if isinstance(item, dict):
            artifact_index = item.get("artifact_index")
            if isinstance(artifact_index, int):
                fallback_id = artifact_index_to_citation_id.get(artifact_index, f"doc-{artifact_index + 1}")
                normalized_ids.append(fallback_id)
                continue
            artifact_name = str(item.get("artifact_id") or item.get("artifact_name") or "").strip()
            if artifact_name:
                normalized_ids.append(artifact_name_to_citation_id.get(artifact_name, artifact_name))
                continue
        text = str(item or "").strip()
        if not text:
            continue
        normalized_ids.append(artifact_name_to_citation_id.get(text, text))
    return list(dict.fromkeys(normalized_ids))


def normalize_answers(
    answers: Any,
    artifact_name_to_citation_id: dict[str, str],
    artifact_index_to_citation_id: dict[int, str],
) -> list[dict[str, Any]]:
    normalized_answers: list[dict[str, Any]] = []
    for answer in answers or []:
        if isinstance(answer, str):
            text = normalize_text(answer)
            if not text:
                continue
            normalized_answers.append({"kind": "direct_answer", "text": text, "citation_ids": []})
            continue
        if not isinstance(answer, dict):
            continue
        text = normalize_text(
            answer.get("text")
            or answer.get("content")
            or answer.get("answer")
            or answer.get("summary")
        )
        if not text:
            continue
        normalized_answers.append(
            {
                "kind": str(answer.get("kind") or "direct_answer"),
                "text": text,
                "citation_ids": normalize_citation_ids(
                    answer.get("citation_ids") or answer.get("citations"),
                    artifact_name_to_citation_id,
                    artifact_index_to_citation_id,
                ),
            }
        )
    return normalized_answers


def extract_answer_embedded_actions(answers: Any) -> list[Any]:
    embedded_actions: list[Any] = []
    for answer in answers or []:
        if not isinstance(answer, dict):
            continue
        for key in ("proposed_actions", "actions", "candidate_actions"):
            value = answer.get(key)
            if isinstance(value, list):
                embedded_actions.extend(value)
    return embedded_actions


def normalize_issues(
    issues: Any,
    artifact_name_to_citation_id: dict[str, str],
    artifact_index_to_citation_id: dict[int, str],
) -> list[dict[str, Any]]:
    normalized_issues: list[dict[str, Any]] = []
    for issue in issues or []:
        if not isinstance(issue, dict):
            continue
        message = normalize_text(issue.get("message") or issue.get("text") or issue.get("content"))
        if not message:
            continue
        severity = str(issue.get("severity") or "warning").strip().lower()
        if severity not in {"info", "warning", "error"}:
            severity = "warning"
        normalized_issues.append(
            {
                "code": str(issue.get("code") or "").strip(),
                "message": message,
                "severity": severity,
                "citation_ids": normalize_citation_ids(
                    issue.get("citation_ids") or issue.get("citations"),
                    artifact_name_to_citation_id,
                    artifact_index_to_citation_id,
                ),
            }
        )
    return normalized_issues


def normalize_actions(
    actions: Any,
    artifact_name_to_citation_id: dict[str, str],
    artifact_index_to_citation_id: dict[int, str],
    *,
    project: str,
    task: str,
) -> list[dict[str, Any]]:
    normalized_actions: list[dict[str, Any]] = []
    for action in actions or []:
        if not isinstance(action, dict):
            continue
        source_action = action
        if project == "radishflow" and task == "suggest_ghost_completion" and isinstance(action.get("ghost_completion"), dict):
            source_action = dict(action["ghost_completion"])
            source_action.setdefault("kind", "ghost_completion")
        title = normalize_text(action.get("title") or action.get("name"))
        rationale = normalize_text(action.get("rationale") or action.get("reason") or action.get("description"))
        if source_action is not action:
            title = normalize_text(source_action.get("title") or source_action.get("name"))
            rationale = normalize_text(
                source_action.get("rationale") or source_action.get("reason") or source_action.get("description")
            )
        if not title or not rationale:
            continue
        kind = str(source_action.get("kind") or "read_only_check").strip()
        allowed_kinds = {"candidate_edit", "candidate_operation", "read_only_check"}
        if project == "radishflow" and task == "suggest_ghost_completion":
            allowed_kinds = {"ghost_completion"}
        if kind not in allowed_kinds:
            kind = "ghost_completion" if project == "radishflow" and task == "suggest_ghost_completion" else "read_only_check"
        risk_level = str(source_action.get("risk_level") or "low").strip().lower()
        if risk_level not in {"low", "medium", "high"}:
            risk_level = "low"
        normalized_action = {
            "kind": kind,
            "title": title,
            "rationale": rationale,
            "risk_level": risk_level,
            "requires_confirmation": bool(source_action.get("requires_confirmation", False)),
            "citation_ids": normalize_citation_ids(
                source_action.get("citation_ids") or source_action.get("citations"),
                artifact_name_to_citation_id,
                artifact_index_to_citation_id,
            ),
        }
        if isinstance(source_action.get("target"), dict):
            normalized_action["target"] = source_action["target"]
        if isinstance(source_action.get("patch"), dict):
            normalized_action["patch"] = source_action["patch"]
        if project == "radishflow" and task == "suggest_ghost_completion":
            if isinstance(source_action.get("preview"), dict):
                preview = source_action["preview"]
                normalized_preview = {
                    key: preview[key]
                    for key in ("ghost_color", "accept_key", "render_priority")
                    if key in preview
                }
                if normalized_preview:
                    normalized_action["preview"] = normalized_preview
            if isinstance(source_action.get("apply"), dict):
                apply = source_action["apply"]
                payload = apply.get("payload")
                if not isinstance(payload, dict):
                    payload = {key: value for key, value in apply.items() if key != "command_kind"}
                normalized_action["apply"] = {
                    "command_kind": str(apply.get("command_kind") or "accept_ghost_completion").strip()
                    or "accept_ghost_completion",
                    "payload": payload,
                }
        normalized_actions.append(normalized_action)
    return normalized_actions


def normalize_citations_from_document(
    citations: Any,
    copilot_request: dict[str, Any],
    fallback_citations: list[dict[str, Any]],
) -> list[dict[str, Any]]:
    artifact_map = {
        str((artifact or {}).get("name") or "").strip(): artifact
        for artifact in copilot_request.get("artifacts") or []
        if str((artifact or {}).get("name") or "").strip()
    }
    context = copilot_request.get("context") or {}
    resource = context.get("resource") or {}
    normalized_citations: list[dict[str, Any]] = []
    for index, citation in enumerate(citations or [], start=1):
        if not isinstance(citation, dict):
            continue
        artifact_name = str(citation.get("artifact_id") or "").strip()
        if not artifact_name and isinstance(citation.get("artifact_index"), int):
            artifact_index = int(citation["artifact_index"])
            artifacts = copilot_request.get("artifacts") or []
            if 0 <= artifact_index < len(artifacts):
                artifact_name = str((artifacts[artifact_index] or {}).get("name") or "").strip()
        artifact = artifact_map.get(artifact_name) or {}
        citation_id = str(citation.get("id") or "").strip()
        if not citation_id:
            fallback = next(
                (
                    item
                    for item in fallback_citations
                    if str(item.get("locator") or "").startswith(f"artifact:{artifact_name}")
                ),
                None,
            )
            citation_id = str((fallback or {}).get("id") or f"doc-{index}")
        citation_fields = build_artifact_citation_fields(artifact, str(resource.get("title") or "").strip())
        has_artifact_backing = bool(artifact_name and artifact)
        if has_artifact_backing:
            label = citation_fields.get("label") or artifact_name or citation_id
        else:
            label = (
                normalize_text(citation.get("label"))
                or citation_fields.get("label")
                or artifact_name
                or citation_id
            )
        normalized_citation = {
            "id": citation_id,
            "kind": str(citation.get("kind") or "artifact").strip() or "artifact",
            "label": label,
        }
        if has_artifact_backing:
            locator = citation_fields.get("locator") or f"artifact:{artifact_name}"
        else:
            locator = normalize_text(citation.get("locator")) or citation_fields.get("locator") or (
                f"artifact:{artifact_name}" if artifact_name else ""
            )
        if locator:
            normalized_citation["locator"] = locator
        if has_artifact_backing:
            excerpt = citation_fields.get("excerpt") or ""
        else:
            excerpt = normalize_text(citation.get("excerpt") or citation.get("text")) or citation_fields.get("excerpt")
        if excerpt:
            normalized_citation["excerpt"] = excerpt
        normalized_citations.append(normalized_citation)
    return normalized_citations or fallback_citations


def normalize_existing_citation_ids(value: Any, known_ids: set[str], fallback_ids: list[str]) -> list[str]:
    normalized = [str(item).strip() for item in (value or []) if str(item).strip() and str(item).strip() in known_ids]
    if normalized:
        return list(dict.fromkeys(normalized))
    return fallback_ids


def risk_level_rank(value: str) -> int:
    order = {"low": 0, "medium": 1, "high": 2}
    return order.get(str(value).strip().lower(), 0)


def max_risk_level(values: list[str]) -> str:
    normalized = [str(value).strip().lower() for value in values if str(value).strip()]
    if not normalized:
        return "medium"
    return max(normalized, key=risk_level_rank)


def synthesize_suggest_edits_citation_ids(
    *,
    target_type: str,
    target_id: str,
    diagnostics: list[dict[str, Any]],
    fallback_citations: list[dict[str, Any]],
    existing_ids: list[str] | None = None,
) -> list[str]:
    known_ids = {
        str(citation.get("id") or "").strip()
        for citation in fallback_citations
        if str(citation.get("id") or "").strip()
    }
    if existing_ids:
        normalized_existing = [
            str(item).strip() for item in existing_ids if str(item).strip() and str(item).strip() in known_ids
        ]
        if normalized_existing:
            return list(dict.fromkeys(normalized_existing))

    citation_ids: list[str] = []
    normalized_target_type = str(target_type or "").strip().lower()
    normalized_target_id = str(target_id or "").strip()
    for index, diagnostic in enumerate(diagnostics, start=1):
        diagnostic_target_type = str((diagnostic or {}).get("target_type") or "").strip().lower()
        diagnostic_target_id = str((diagnostic or {}).get("target_id") or "").strip()
        if normalized_target_id and diagnostic_target_id != normalized_target_id:
            continue
        if normalized_target_type and diagnostic_target_type and diagnostic_target_type != normalized_target_type:
            continue
        citation_ids.append(f"diag-{index}")

    for citation in fallback_citations:
        citation_id = str(citation.get("id") or "").strip()
        if not citation_id or not citation_id.startswith("flowdoc-"):
            continue
        label = str(citation.get("label") or "").strip()
        if normalized_target_id and label.endswith(normalized_target_id):
            citation_ids.append(citation_id)
            break

    for citation in fallback_citations:
        citation_id = str(citation.get("id") or "").strip()
        if citation_id.startswith("snapshot-") or citation_id.startswith("diag-summary-"):
            citation_ids.append(citation_id)
            break
    return list(dict.fromkeys(citation_ids))


def build_suggest_edits_target_citation_lookup(citations: list[dict[str, Any]]) -> dict[tuple[str, str], str]:
    target_citation_lookup: dict[tuple[str, str], str] = {}
    for citation in citations:
        citation_id = str(citation.get("id") or "").strip()
        label = str(citation.get("label") or "").strip()
        if citation_id.startswith("flowdoc-stream-") and "/" in label:
            target_citation_lookup[("stream", label.split("/")[-1].strip())] = citation_id
        elif citation_id.startswith("flowdoc-unit-") and "/" in label:
            target_citation_lookup[("unit", label.split("/")[-1].strip())] = citation_id
    return target_citation_lookup


def normalize_spec_placeholders(values: list[Any]) -> list[str]:
    alias_map = {
        "flow_rate": "flow_rate_kg_per_h",
        "mass_flow": "flow_rate_kg_per_h",
        "mass_flow_rate": "flow_rate_kg_per_h",
    }
    normalized: list[str] = []
    for value in values:
        placeholder = str(value).strip()
        if not placeholder:
            continue
        placeholder = alias_map.get(placeholder, placeholder)
        if placeholder not in normalized:
            normalized.append(placeholder)
    return normalized


def normalize_parameter_placeholders(values: list[Any]) -> list[str]:
    alias_map = {
        "outlet_temperature_target_c": "outlet_temperature_c",
        "outlet_temp_target_c": "outlet_temperature_c",
        "temperature_target_c": "outlet_temperature_c",
    }
    normalized: list[str] = []
    for value in values:
        placeholder = str(value).strip()
        if not placeholder:
            continue
        placeholder = alias_map.get(placeholder, placeholder)
        if placeholder not in normalized:
            normalized.append(placeholder)
    return normalized


def build_suggest_edits_response_citation_ids(
    *,
    diagnostic_index: int,
    diagnostic_code: str,
    target_type: str,
    target_id: str,
    target_citation_lookup: dict[tuple[str, str], str],
    snapshot_citation_id: str,
    total_diagnostics: int,
    item_kind: str,
) -> list[str]:
    citation_ids: list[str] = [f"diag-{diagnostic_index}"]
    target_citation_id = target_citation_lookup.get((target_type, target_id))
    is_single_disconnected = total_diagnostics == 1 and diagnostic_code == "STREAM_DISCONNECTED"

    if item_kind == "issue" and is_single_disconnected:
        return citation_ids

    if target_citation_id and not is_single_disconnected:
        citation_ids.append(target_citation_id)
    if diagnostic_code == "STREAM_DISCONNECTED" and snapshot_citation_id:
        citation_ids.append(snapshot_citation_id)
    if target_citation_id and is_single_disconnected and item_kind == "answer":
        citation_ids.append(target_citation_id)
    return list(dict.fromkeys(citation_ids))


def synthesize_suggest_edits_action(
    diagnostic: dict[str, Any],
    *,
    diagnostic_index: int,
    total_diagnostics: int,
    target_citation_lookup: dict[tuple[str, str], str],
    snapshot_citation_id: str,
) -> dict[str, Any]:
    diagnostic_code = str(diagnostic.get("code") or "").strip()
    message = normalize_text(diagnostic.get("message") or diagnostic.get("text") or "")
    target_type = str(diagnostic.get("target_type") or "").strip().lower()
    target_id = str(diagnostic.get("target_id") or "").strip()
    if target_type not in {"stream", "unit"}:
        target_type = "stream" if "stream" in diagnostic_code.lower() else "unit"
    citation_ids = build_suggest_edits_response_citation_ids(
        diagnostic_index=diagnostic_index,
        diagnostic_code=diagnostic_code,
        target_type=target_type,
        target_id=target_id,
        target_citation_lookup=target_citation_lookup,
        snapshot_citation_id=snapshot_citation_id,
        total_diagnostics=total_diagnostics,
        item_kind="action",
    )

    if diagnostic_code == "STREAM_DISCONNECTED":
        return {
            "kind": "candidate_edit",
            "title": f"为 {target_id} 预留下游重连占位",
            "target": {
                "type": "stream",
                "id": target_id,
            },
            "rationale": "当前断链问题只能通过人工确认后续去向来处理，因此 patch 仅保留候选连接占位。",
            "patch": {
                "connection_placeholder": {
                    "expected_downstream_kind": "consumer_or_export_sink",
                    "requires_manual_binding": True,
                }
            },
            "risk_level": "high",
            "requires_confirmation": True,
            "citation_ids": citation_ids,
        }
    if diagnostic_code == "STREAM_SPEC_MISSING":
        return {
            "kind": "candidate_edit",
            "title": f"为 {target_id} 补充缺失规格占位",
            "target": {
                "type": "stream",
                "id": target_id,
            },
            "rationale": "当前诊断已明确指出该流股规格不完整，因此更合适的是补局部规格占位，而不是给整图 patch。",
            "patch": {
                "spec_placeholders": infer_stream_spec_placeholders(message),
                "retain_existing_specs": True,
            },
            "risk_level": "medium",
            "requires_confirmation": True,
            "citation_ids": citation_ids,
        }
    return {
        "kind": "candidate_edit",
        "title": f"为 {target_id} 补充局部参数占位",
        "target": {
            "type": target_type,
            "id": target_id,
        },
        "rationale": "当前诊断集中在局部参数或配置完整性，因此更合适的是输出可审查的局部参数占位提案。",
        "patch": {
            "parameter_placeholders": infer_parameter_placeholders(message),
            "patch_scope": "unit_parameters" if target_type == "unit" else "stream_parameters",
        },
        "risk_level": "medium",
        "requires_confirmation": True,
        "citation_ids": citation_ids,
    }


def canonicalize_suggest_edits_response(
    coerced: dict[str, Any],
    copilot_request: dict[str, Any],
    fallback_citations: list[dict[str, Any]],
) -> dict[str, Any]:
    context = copilot_request.get("context") or {}
    diagnostics = [diagnostic for diagnostic in (context.get("diagnostics") or []) if isinstance(diagnostic, dict)]
    target_citation_lookup = build_suggest_edits_target_citation_lookup(fallback_citations)
    snapshot_citation_id = next(
        (
            str(citation.get("id") or "").strip()
            for citation in fallback_citations
            if str(citation.get("id") or "").strip() == "snapshot-1"
        ),
        "",
    )
    existing_actions_by_target: dict[tuple[str, str], dict[str, Any]] = {}
    for action in list(coerced.get("proposed_actions") or []):
        if not isinstance(action, dict):
            continue
        target = action.get("target") or {}
        target_type = str(target.get("type") or "").strip().lower()
        target_id = str(target.get("id") or "").strip()
        if target_type in {"stream", "unit"} and target_id and (target_type, target_id) not in existing_actions_by_target:
            existing_actions_by_target[(target_type, target_id)] = dict(action)

    normalized_actions: list[dict[str, Any]] = []
    actionable_diagnostic_codes = {"STREAM_DISCONNECTED", "STREAM_SPEC_MISSING", "UNIT_PARAMETER_INCOMPLETE"}
    for diagnostic_index, diagnostic in enumerate(diagnostics, start=1):
        synthesized_action = synthesize_suggest_edits_action(
            diagnostic,
            diagnostic_index=diagnostic_index,
            total_diagnostics=len(diagnostics),
            target_citation_lookup=target_citation_lookup,
            snapshot_citation_id=snapshot_citation_id,
        )
        target = synthesized_action.get("target") or {}
        target_type = str(target.get("type") or "").strip().lower()
        target_id = str(target.get("id") or "").strip()
        existing_action = existing_actions_by_target.get((target_type, target_id), {})
        merged_action = dict(existing_action)
        merged_action["kind"] = "candidate_edit"
        merged_action["target"] = synthesized_action["target"]
        merged_action["title"] = normalize_text(merged_action.get("title")) or synthesized_action["title"]
        merged_action["rationale"] = normalize_text(merged_action.get("rationale")) or synthesized_action["rationale"]
        merged_patch = merged_action.get("patch")
        if not isinstance(merged_patch, dict):
            merged_patch = {}
        synthesized_patch = dict(synthesized_action.get("patch") or {})
        if str(diagnostic.get("code") or "").strip() == "STREAM_DISCONNECTED":
            connection_placeholder = merged_patch.get("connection_placeholder")
            if not isinstance(connection_placeholder, dict):
                connection_placeholder = {}
            merged_patch["connection_placeholder"] = {
                "expected_downstream_kind": str(
                    connection_placeholder.get("expected_downstream_kind")
                    or synthesized_patch["connection_placeholder"]["expected_downstream_kind"]
                ),
                "requires_manual_binding": bool(
                    connection_placeholder.get("requires_manual_binding", True)
                ),
            }
        elif str(diagnostic.get("code") or "").strip() == "STREAM_SPEC_MISSING":
            spec_placeholders = normalize_spec_placeholders(list(merged_patch.get("spec_placeholders") or []))
            merged_patch = {
                "spec_placeholders": spec_placeholders
                or normalize_spec_placeholders(list(synthesized_patch.get("spec_placeholders") or [])),
                "retain_existing_specs": bool(merged_patch.get("retain_existing_specs", True)),
            }
        else:
            parameter_placeholders = normalize_parameter_placeholders(list(merged_patch.get("parameter_placeholders") or []))
            merged_patch = {
                "parameter_placeholders": parameter_placeholders
                or normalize_parameter_placeholders(list(synthesized_patch.get("parameter_placeholders") or [])),
                "patch_scope": str(
                    merged_patch.get("patch_scope") or synthesized_patch.get("patch_scope") or ""
                ).strip()
                or str(synthesized_patch.get("patch_scope") or ""),
                "preserve_existing_connections": bool(merged_patch.get("preserve_existing_connections", True)),
            }
        merged_action["patch"] = merged_patch
        merged_action["risk_level"] = max_risk_level(
            [str(existing_action.get("risk_level") or ""), str(synthesized_action.get("risk_level") or "")]
        )
        if str(diagnostic.get("code") or "").strip() == "STREAM_DISCONNECTED":
            merged_action["risk_level"] = "high"
        merged_action["requires_confirmation"] = True
        merged_action["citation_ids"] = list(synthesized_action.get("citation_ids") or [])
        normalized_actions.append(merged_action)

    synthesized_issues: list[dict[str, Any]] = []
    existing_issues_by_code: dict[str, dict[str, Any]] = {}
    for issue in list(coerced.get("issues") or []):
        if not isinstance(issue, dict):
            continue
        code = str(issue.get("code") or "").strip()
        if code and code not in existing_issues_by_code:
            existing_issues_by_code[code] = dict(issue)
    for diagnostic_index, diagnostic in enumerate(diagnostics, start=1):
        diagnostic_code = str(diagnostic.get("code") or "").strip()
        target_type = str(diagnostic.get("target_type") or "").strip().lower()
        target_id = str(diagnostic.get("target_id") or "").strip()
        existing_issue = existing_issues_by_code.get(diagnostic_code, {})
        issue_message = normalize_text(existing_issue.get("message")) or normalize_text(
            diagnostic.get("message") or diagnostic.get("text") or ""
        )
        synthesized_issues.append(
            {
                "code": diagnostic_code,
                "message": issue_message or f"{target_id or '当前对象'} 存在需要人工复核的编辑问题。",
                "severity": str(diagnostic.get("severity") or existing_issue.get("severity") or "warning").strip().lower()
                or "warning",
                "citation_ids": build_suggest_edits_response_citation_ids(
                    diagnostic_index=diagnostic_index,
                    diagnostic_code=diagnostic_code,
                    target_type=target_type,
                    target_id=target_id,
                    target_citation_lookup=target_citation_lookup,
                    snapshot_citation_id=snapshot_citation_id,
                    total_diagnostics=len(diagnostics),
                    item_kind="issue",
                ),
            }
        )

    if not coerced.get("answers"):
        answer_citation_ids = list(
            dict.fromkeys(
                citation_id
                for action in normalized_actions
                for citation_id in (action.get("citation_ids") or [])
                if str(citation_id).strip()
            )
        )
        coerced["answers"] = [
            {
                "kind": "edit_rationale",
                "text": normalize_text(coerced.get("summary"))
                or "当前响应已收口为候选编辑提案，仍需人工确认后再决定是否应用。",
                "citation_ids": answer_citation_ids,
            }
        ]
    else:
        for answer in coerced.get("answers") or []:
            if not isinstance(answer, dict):
                continue
            if not (answer.get("citation_ids") or []):
                answer["citation_ids"] = list(
                    dict.fromkeys(
                        citation_id
                        for action in normalized_actions
                        for citation_id in (action.get("citation_ids") or [])
                        if str(citation_id).strip()
                    )
                )
    if len(diagnostics) == 1 and str(diagnostics[0].get("code") or "").strip() == "STREAM_DISCONNECTED":
        primary_action_citation_ids = list((normalized_actions[0] or {}).get("citation_ids") or []) if normalized_actions else []
        for answer in coerced.get("answers") or []:
            if not isinstance(answer, dict):
                continue
            answer["citation_ids"] = primary_action_citation_ids

    referenced_citation_ids = list(
        dict.fromkeys(
            citation_id
            for section in [coerced.get("answers") or [], synthesized_issues, normalized_actions]
            for item in section
            if isinstance(item, dict)
            for citation_id in (item.get("citation_ids") or [])
            if str(citation_id).strip()
        )
    )
    coerced["issues"] = synthesized_issues
    coerced["proposed_actions"] = normalized_actions
    coerced["citations"] = [
        citation
        for citation in fallback_citations
        if str(citation.get("id") or "").strip() in referenced_citation_ids
    ]
    coerced["requires_confirmation"] = bool(normalized_actions)
    if not normalized_actions:
        coerced["status"] = "failed"
    elif any(str(action.get("risk_level") or "").strip().lower() == "high" for action in normalized_actions):
        coerced["status"] = "partial"
    elif any(str((diagnostic or {}).get("code") or "").strip() not in actionable_diagnostic_codes for diagnostic in diagnostics):
        coerced["status"] = "partial"
    else:
        coerced["status"] = "ok"
    coerced["risk_level"] = max_risk_level(
        [str(action.get("risk_level") or "") for action in normalized_actions]
        or [str(coerced.get("risk_level") or "")]
    )
    if not normalize_text(coerced.get("summary")) and normalized_actions:
        first_action = normalized_actions[0]
        target_id = str(((first_action.get("target") or {}).get("id")) or "").strip()
        coerced["summary"] = f"当前围绕 {target_id or '当前对象'} 生成候选编辑提案。"
    return coerced


def canonicalize_ghost_response(
    coerced: dict[str, Any],
    copilot_request: dict[str, Any],
    fallback_citations: list[dict[str, Any]],
) -> dict[str, Any]:
    context = copilot_request.get("context") or {}
    legal_candidates = [
        candidate
        for candidate in (context.get("legal_candidate_completions") or [])
        if isinstance(candidate, dict) and str(candidate.get("candidate_ref") or "").strip()
    ]
    citations = fallback_citations
    citation_ids = [str(citation.get("id") or "") for citation in citations if str(citation.get("id") or "")]
    known_citation_ids = set(citation_ids)
    summary = normalize_text(coerced.get("summary"))
    answers = list(coerced.get("answers") or [])
    selected_unit = context.get("selected_unit") or {}
    selected_unit_id = str(selected_unit.get("id") or "").strip() or str((context.get("selected_unit_ids") or [""])[0] or "")

    if not legal_candidates:
        answer_text = summary or "当前上下文里没有本地规则确认的合法 ghost completion 候选，因此此时应保持空建议。"
        issue_message = "当前没有本地规则确认的合法 ghost completion 候选，不能安全地产生默认或手动 ghost 建议。"
        coerced["status"] = "partial"
        coerced["summary"] = summary or f"{selected_unit_id or '当前单元'} 当前没有可确认的合法 ghost 候选，因此不返回默认补全建议。"
        coerced["answers"] = [
            {
                "kind": "empty_suggestion_reason",
                "text": answer_text,
                "citation_ids": normalize_existing_citation_ids(
                    (answers[0] or {}).get("citation_ids") if answers else [],
                    known_citation_ids,
                    [citation_id for citation_id in citation_ids if citation_id in {"ctx-empty-candidates", "ctx-pattern", "ctx-recent-action"}],
                ),
            }
        ]
        coerced["issues"] = [
            {
                "code": "INSUFFICIENT_GHOST_CONTEXT",
                "message": issue_message,
                "severity": "info",
                "citation_ids": [citation_id for citation_id in citation_ids if citation_id in {"ctx-empty-candidates", "ctx-pattern"}],
            }
        ]
        coerced["proposed_actions"] = []
        coerced["citations"] = citations
        coerced["risk_level"] = "low"
        coerced["requires_confirmation"] = False
        coerced["confidence"] = coerced.get("confidence") if isinstance(coerced.get("confidence"), (int, float)) else 0.9
        return coerced

    primary_candidate = pick_primary_ghost_candidate(legal_candidates)
    has_default_tab_candidate = bool(
        primary_candidate
        and primary_candidate.get("is_tab_default") is True
        and primary_candidate.get("is_high_confidence") is True
        and not [str(flag).strip() for flag in (primary_candidate.get("conflict_flags") or []) if str(flag).strip()]
    )
    candidate_subset = [primary_candidate] if has_default_tab_candidate and primary_candidate else legal_candidates[:2]
    normalized_actions: list[dict[str, Any]] = []
    for index, candidate in enumerate(candidate_subset, start=1):
        candidate_citation_id = f"ctx-candidate-{index}"
        action = build_ghost_completion_action(
            candidate,
            action_index=index,
            accept_key="Tab" if has_default_tab_candidate and index == 1 else "manual_only",
            risk_level="low" if has_default_tab_candidate else "medium",
            citation_ids=[citation_id for citation_id in citation_ids if citation_id in {candidate_citation_id, "ctx-unit", "ctx-pattern", "ctx-naming-hints"}],
        )
        normalized_actions.append(action)

    existing_answer_text = normalize_text((answers[0] or {}).get("text") if answers else "")
    coerced["status"] = "ok"
    coerced["summary"] = summary or (
        f"{selected_unit_id or '当前单元'} 当前最适合作为默认 ghost 的候选已明确，可优先渲染首条 `Tab` 建议。"
        if has_default_tab_candidate
        else f"{selected_unit_id or '当前单元'} 当前存在多个合法 ghost 候选，但证据不足以默认选中其中任意一条。"
    )
    coerced["answers"] = [
        {
            "kind": "ghost_rationale",
            "text": existing_answer_text
            or (
                f"{selected_unit_id or '当前单元'} 的候选已明确，可优先接受首个默认 ghost。"
                if has_default_tab_candidate
                else "当前合法候选之间没有形成足够清晰的领先者，因此只应返回可见 ghost，而不应把任何一条候选直接升级成默认 Tab。"
            ),
            "citation_ids": normalize_existing_citation_ids(
                (answers[0] or {}).get("citation_ids") if answers else [],
                known_citation_ids,
                [citation_id for citation_id in citation_ids if citation_id in {"ctx-candidate-1", "ctx-candidate-2", "ctx-pattern", "ctx-unit", "ctx-naming-hints"}],
            ),
        }
    ]
    coerced["issues"] = []
    coerced["proposed_actions"] = normalized_actions
    coerced["citations"] = citations
    coerced["risk_level"] = "low" if has_default_tab_candidate else "medium"
    coerced["requires_confirmation"] = False
    coerced["confidence"] = coerced.get("confidence") if isinstance(coerced.get("confidence"), (int, float)) else (0.93 if has_default_tab_candidate else 0.78)
    return coerced


def coerce_response_document(document: dict[str, Any], copilot_request: dict[str, Any], raw_text: str) -> dict[str, Any]:
    coerced = dict(document)
    fallback_citations, artifact_name_to_citation_id, artifact_index_to_citation_id = build_citation_maps(copilot_request)
    project = str(copilot_request.get("project") or "").strip()
    task = str(copilot_request.get("task") or "").strip()
    embedded_actions = extract_answer_embedded_actions(coerced.get("answers"))
    if embedded_actions and not isinstance(coerced.get("proposed_actions"), list):
        coerced["proposed_actions"] = embedded_actions
    elif embedded_actions and not (coerced.get("proposed_actions") or []):
        coerced["proposed_actions"] = embedded_actions
    coerced["status"] = map_status(coerced.get("status"))
    coerced["answers"] = normalize_answers(
        coerced.get("answers"),
        artifact_name_to_citation_id,
        artifact_index_to_citation_id,
    )
    coerced["issues"] = normalize_issues(
        coerced.get("issues"),
        artifact_name_to_citation_id,
        artifact_index_to_citation_id,
    )
    coerced["proposed_actions"] = normalize_actions(
        coerced.get("proposed_actions"),
        artifact_name_to_citation_id,
        artifact_index_to_citation_id,
        project=project,
        task=task,
    )
    coerced["citations"] = normalize_citations_from_document(
        coerced.get("citations"),
        copilot_request,
        fallback_citations,
    )
    if "confidence" in coerced:
        try:
            coerced["confidence"] = max(0.0, min(1.0, float(coerced["confidence"])))
        except (TypeError, ValueError):
            coerced["confidence"] = 0.2
    if "risk_level" in coerced:
        risk_level = str(coerced.get("risk_level") or "").strip().lower()
        coerced["risk_level"] = risk_level if risk_level in {"low", "medium", "high"} else "medium"
    if "requires_confirmation" in coerced:
        coerced["requires_confirmation"] = bool(coerced["requires_confirmation"])
    coerced.setdefault("schema_version", 1)
    coerced.setdefault("project", copilot_request.get("project"))
    coerced.setdefault("task", copilot_request.get("task"))
    if project == "radishflow" and task == "suggest_ghost_completion":
        extracted_summary = extract_embedded_summary_text(coerced.get("summary"))
        if extracted_summary:
            coerced["summary"] = extracted_summary
        for answer in coerced.get("answers") or []:
            if not isinstance(answer, dict):
                continue
            extracted_answer_text = extract_embedded_summary_text(answer.get("text"))
            if extracted_answer_text:
                answer["text"] = extracted_answer_text
    if not normalize_text(coerced.get("summary")) and coerced.get("answers"):
        first_answer = (coerced.get("answers") or [{}])[0]
        if isinstance(first_answer, dict):
            derived_summary = normalize_text(first_answer.get("text") or first_answer.get("summary"))
            if derived_summary:
                coerced["summary"] = derived_summary
    if project == "radishflow" and task == "suggest_ghost_completion" and not coerced.get("answers"):
        fallback_answer_text = normalize_text(coerced.get("summary"))
        if fallback_answer_text:
            coerced["answers"] = [
                {
                    "kind": "ghost_rationale",
                    "text": fallback_answer_text,
                    "citation_ids": [],
                }
            ]
    if project == "radishflow" and task == "suggest_ghost_completion":
        coerced = canonicalize_ghost_response(coerced, copilot_request, fallback_citations)
    if project == "radishflow" and task == "suggest_flowsheet_edits":
        coerced = canonicalize_suggest_edits_response(coerced, copilot_request, fallback_citations)
    coerced.setdefault("summary", "模型已返回结构化响应，但摘要字段为空。")
    coerced.setdefault("answers", [])
    coerced.setdefault("issues", [])
    coerced.setdefault("proposed_actions", [])
    coerced.setdefault("citations", [])
    coerced.setdefault("confidence", 0.2)
    coerced.setdefault("risk_level", "medium")
    coerced.setdefault("requires_confirmation", False)
    coerced.setdefault("status", "partial")
    coerced = {key: value for key, value in coerced.items() if key in RESPONSE_TOP_LEVEL_KEYS}
    try:
        validate_response_document(coerced)
    except jsonschema.ValidationError:
        return make_failed_response(
            copilot_request,
            "模型返回了 JSON，但未通过 CopilotResponse schema 校验，当前已保留原始文本供调试。",
            raw_text=raw_text,
            code="MODEL_OUTPUT_SCHEMA_INVALID",
        )
    return coerced
