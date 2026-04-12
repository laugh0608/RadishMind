from __future__ import annotations

import json
import re
from typing import Any

from .inference_support import (
    GHOST_MALFORMED_JSON_REPAIR_PATTERNS,
    GHOST_MANUAL_MULTI_ACTION_REPAIR_PATTERNS,
    RESPONSE_TOP_LEVEL_KEYS,
    RESPONSE_SCHEMA_PATH,
    build_artifact_citation_fields,
    build_citation_maps,
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
