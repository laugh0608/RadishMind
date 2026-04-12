from __future__ import annotations

import json
import http.client
import os
import re
from datetime import datetime, timezone
from pathlib import Path
from typing import Any
from urllib import error, request

import jsonschema

from .artifact_summary import extract_artifact_summary_metadata


REPO_ROOT = Path(__file__).resolve().parent.parent.parent
REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-request.schema.json"
RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
PROMPT_PATHS = {
    ("radish", "answer_docs_question"): REPO_ROOT / "prompts/tasks/radish-answer-docs-question-system.md",
    ("radishflow", "suggest_ghost_completion"): REPO_ROOT / "prompts/tasks/radishflow-suggest-ghost-completion-system.md",
}
SCHEMA_CACHE: dict[Path, Any] = {}
SENTENCE_BREAK_RE = re.compile(r"(?<=[。！？.!?])\s+")
ENV_FILE_PATH = REPO_ROOT / ".env"
# Narrow repairs for stable provider failures observed in real ghost batches:
# - near-complete JSON with one extra closing brace before ghost action tail fields
# - manual_only multi-action payloads that prematurely close proposed_actions / answer scopes
GHOST_MALFORMED_JSON_REPAIR_PATTERNS = (
    (
        re.compile(r"}}(?=,\"(?:preview|apply|risk_level|requires_confirmation|citation_ids)\")"),
        "}",
    ),
    (
        re.compile(r"}}(?=\],\"(?:citations|issues|confidence|risk_level|requires_confirmation|status|summary)\")"),
        "}",
    ),
)
GHOST_MANUAL_MULTI_ACTION_REPAIR_PATTERNS = (
    (
        re.compile(r"}}]\},(?=\{\"action_id\":\"act-[^\"]+\",\"action_kind\":\"ghost_completion\",\"ghost_completion\":)"),
        "}},",
    ),
    (
        re.compile(r"\"manual_only\":true}}}],(?=\"issues\")"),
        "\"manual_only\":true}}],",
    ),
)
RESPONSE_TOP_LEVEL_KEYS = {
    "schema_version",
    "status",
    "project",
    "task",
    "summary",
    "answers",
    "issues",
    "proposed_actions",
    "citations",
    "confidence",
    "risk_level",
    "requires_confirmation",
}


def load_schema(path: Path) -> Any:
    if path not in SCHEMA_CACHE:
        SCHEMA_CACHE[path] = json.loads(path.read_text(encoding="utf-8"))
    return SCHEMA_CACHE[path]


def load_env_file(path: Path) -> None:
    if not path.is_file():
        return
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        key = key.strip()
        if not key or key in os.environ:
            continue
        parsed = value.strip()
        if len(parsed) >= 2 and parsed[0] == parsed[-1] and parsed[0] in {"'", '"'}:
            parsed = parsed[1:-1]
        os.environ[key] = parsed


def getenv_stripped(name: str) -> str:
    return os.getenv(name, "").strip()


def normalize_provider_profile_name(profile: str | None) -> str:
    normalized = re.sub(r"[^A-Za-z0-9]+", "_", str(profile or "").strip()).strip("_").upper()
    return normalized


def profile_env_key(profile: str, suffix: str) -> str:
    normalized_profile = normalize_provider_profile_name(profile)
    return f"RADISHMIND_MODEL_PROFILE_{normalized_profile}_{suffix}"


def profile_env_value(profile: str, suffix: str) -> str:
    normalized_profile = normalize_provider_profile_name(profile)
    if not normalized_profile:
        return ""
    return getenv_stripped(profile_env_key(profile, suffix))


def resolve_openai_compatible_profile(provider_profile: str | None) -> str:
    return str(provider_profile or getenv_stripped("RADISHMIND_MODEL_PROFILE") or "anyrouter").strip()


def infer_profile_api_style(base_url: str) -> str:
    normalized_base_url = str(base_url or "").strip().lower()
    if "generativelanguage.googleapis.com" in normalized_base_url:
        return "gemini-native"
    return "openai-compatible"


def resolve_openai_compatible_config(
    *,
    provider_profile: str | None,
    model: str | None,
    base_url: str | None,
    api_key: str | None,
    request_timeout_seconds: float | None,
) -> dict[str, Any]:
    resolved_profile = resolve_openai_compatible_profile(provider_profile)
    normalized_profile = normalize_provider_profile_name(resolved_profile)
    resolved_model = model or profile_env_value(resolved_profile, "NAME") or getenv_stripped("RADISHMIND_MODEL_NAME")
    resolved_base_url = base_url or profile_env_value(resolved_profile, "BASE_URL") or getenv_stripped("RADISHMIND_MODEL_BASE_URL")
    resolved_api_key = api_key or profile_env_value(resolved_profile, "API_KEY") or getenv_stripped("RADISHMIND_MODEL_API_KEY")
    resolved_api_style = (
        profile_env_value(resolved_profile, "API_STYLE")
        or getenv_stripped("RADISHMIND_MODEL_API_STYLE")
        or infer_profile_api_style(resolved_base_url)
    )
    if request_timeout_seconds is None:
        timeout_env_value = (
            profile_env_value(resolved_profile, "REQUEST_TIMEOUT_SECONDS")
            or getenv_stripped("RADISHMIND_MODEL_REQUEST_TIMEOUT_SECONDS")
        )
        resolved_request_timeout_seconds = float(timeout_env_value) if timeout_env_value else 120.0
    else:
        resolved_request_timeout_seconds = float(request_timeout_seconds)
    return {
        "profile": resolved_profile,
        "normalized_profile": normalized_profile,
        "api_style": resolved_api_style,
        "model": resolved_model,
        "base_url": resolved_base_url,
        "api_key": resolved_api_key,
        "request_timeout_seconds": resolved_request_timeout_seconds,
    }


def validate_request_document(document: Any) -> None:
    jsonschema.validate(document, load_schema(REQUEST_SCHEMA_PATH))


def validate_response_document(document: Any) -> None:
    jsonschema.validate(document, load_schema(RESPONSE_SCHEMA_PATH))


def utc_now_iso() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def normalize_text(value: Any) -> str:
    if value is None:
        return ""
    if isinstance(value, str):
        return re.sub(r"\s+", " ", value).strip()
    return re.sub(r"\s+", " ", json.dumps(value, ensure_ascii=False)).strip()


def artifact_content_text(artifact: dict[str, Any]) -> str:
    if "content" not in artifact:
        return ""
    return normalize_text(artifact.get("content"))


def artifact_metadata_summary_text(artifact: dict[str, Any]) -> tuple[str, str, str]:
    return extract_artifact_summary_metadata(artifact.get("metadata") or {})


def artifact_citation_excerpt_source(artifact: dict[str, Any]) -> tuple[str, str, str]:
    content_text = artifact_content_text(artifact)
    if content_text:
        return content_text, "", ""
    return artifact_metadata_summary_text(artifact)


def artifact_excerpt(artifact: dict[str, Any], limit: int = 180) -> str:
    text, _, _ = artifact_citation_excerpt_source(artifact)
    if len(text) <= limit:
        return text
    return text[: limit - 1].rstrip() + "…"


def build_artifact_citation_fields(artifact: dict[str, Any], resource_title: str = "", limit: int = 180) -> dict[str, str]:
    artifact_name = str(artifact.get("name") or "").strip()
    metadata = artifact.get("metadata") or {}
    page_slug = str((metadata.get("page_slug") if isinstance(metadata, dict) else "") or "").strip()
    excerpt_source_text, summary_key, summary_label_suffix = artifact_citation_excerpt_source(artifact)

    label = page_slug or str(resource_title or "").strip() or artifact_name
    if not label and artifact_name and summary_label_suffix:
        label = f"{artifact_name} {summary_label_suffix}"
    elif label == artifact_name and summary_label_suffix:
        label = f"{artifact_name} {summary_label_suffix}"

    locator = f"artifact:{artifact_name}" if artifact_name else ""
    if locator and summary_key:
        locator = f"{locator}.metadata.{summary_key}"

    fields: dict[str, str] = {}
    if label:
        fields["label"] = label
    if locator:
        fields["locator"] = locator

    excerpt = excerpt_source_text
    if excerpt:
        fields["excerpt"] = excerpt if len(excerpt) <= limit else excerpt[: limit - 1].rstrip() + "…"
    return fields


def first_sentences(text: str, max_sentences: int = 2) -> str:
    normalized = normalize_text(text)
    if not normalized:
        return ""
    parts = [part.strip() for part in SENTENCE_BREAK_RE.split(normalized) if part.strip()]
    if not parts:
        return normalized
    return " ".join(parts[:max_sentences]).strip()


def derive_input_record(copilot_request: dict[str, Any]) -> dict[str, Any]:
    context = copilot_request.get("context") or {}
    resource = context.get("resource") or {}
    artifacts = copilot_request.get("artifacts") or []
    artifact_names = [
        str((artifact or {}).get("name") or "").strip()
        for artifact in artifacts
        if str((artifact or {}).get("name") or "").strip()
    ]
    input_record = {
        "current_app": str(context.get("current_app") or ""),
        "route": str(context.get("route") or ""),
        "resource_slug": str(resource.get("slug") or ""),
        "search_scope": [str(item).strip() for item in (context.get("search_scope") or []) if str(item).strip()],
        "artifact_names": artifact_names,
    }
    if str(copilot_request.get("project") or "") == "radishflow" and str(copilot_request.get("task") or "") == "suggest_ghost_completion":
        selected_unit_ids = [
            str(unit_id).strip()
            for unit_id in (context.get("selected_unit_ids") or [])
            if str(unit_id).strip()
        ]
        legal_candidate_refs = [
            str((candidate or {}).get("candidate_ref") or "").strip()
            for candidate in (context.get("legal_candidate_completions") or [])
            if str((candidate or {}).get("candidate_ref") or "").strip()
        ]
        input_record["document_revision"] = context.get("document_revision")
        if selected_unit_ids:
            input_record["selected_unit_ids"] = selected_unit_ids
        if legal_candidate_refs:
            input_record["legal_candidate_refs"] = legal_candidate_refs
    return {key: value for key, value in input_record.items() if value}


def citation_prefix_for_artifact(artifact: dict[str, Any]) -> str:
    metadata = artifact.get("metadata") or {}
    source_type = str(metadata.get("source_type") or "").strip()
    if source_type in {"docs", "wiki"}:
        return "doc"
    if source_type == "attachments":
        return "att"
    if source_type == "forum":
        return "forum"
    if source_type == "faq":
        return "faq"
    return "ref"


def build_citations(copilot_request: dict[str, Any]) -> list[dict[str, Any]]:
    project = str(copilot_request.get("project") or "").strip()
    task = str(copilot_request.get("task") or "").strip()
    if project == "radishflow" and task == "suggest_ghost_completion":
        return build_ghost_context_citations(copilot_request)

    citations: list[dict[str, Any]] = []
    counters: dict[str, int] = {}
    context = copilot_request.get("context") or {}
    resource = context.get("resource") or {}
    for artifact in copilot_request.get("artifacts") or []:
        prefix = citation_prefix_for_artifact(artifact)
        counters[prefix] = counters.get(prefix, 0) + 1
        citation_id = f"{prefix}-{counters[prefix]}"
        citation_fields = build_artifact_citation_fields(artifact, str(resource.get("title") or "").strip())
        citation = {
            "id": citation_id,
            "kind": "artifact",
            "label": citation_fields.get("label") or citation_id,
        }
        locator = citation_fields.get("locator")
        if locator:
            citation["locator"] = locator
        excerpt = citation_fields.get("excerpt") or artifact_excerpt(artifact)
        if excerpt:
            citation["excerpt"] = excerpt
        citations.append(citation)
    return citations


def build_ghost_context_citations(copilot_request: dict[str, Any]) -> list[dict[str, Any]]:
    context = copilot_request.get("context") or {}
    citations: list[dict[str, Any]] = []
    selected_unit = context.get("selected_unit") or {}
    if isinstance(selected_unit, dict) and str(selected_unit.get("id") or "").strip():
        citations.append(
            {
                "id": "ctx-unit",
                "kind": "context",
                "label": f"当前选中的 {selected_unit.get('id')}",
                "locator": "context:selected_unit",
                "excerpt": (
                    f"selected_unit={selected_unit.get('id')} "
                    f"kind={selected_unit.get('kind')} "
                    f"name={selected_unit.get('name')}"
                ).strip(),
            }
        )

    legal_candidates = list(context.get("legal_candidate_completions") or [])
    if not legal_candidates:
        citations.append(
            {
                "id": "ctx-empty-candidates",
                "kind": "context",
                "label": "空的 legal_candidate_completions",
                "locator": "context:legal_candidate_completions",
                "excerpt": "legal_candidate_completions=[]",
            }
        )
    else:
        for index, candidate in enumerate(legal_candidates[:3], start=1):
            candidate_ref = str((candidate or {}).get("candidate_ref") or "").strip()
            target_port_key = str((candidate or {}).get("target_port_key") or "").strip()
            citations.append(
                {
                    "id": f"ctx-candidate-{index}",
                    "kind": "context",
                    "label": f"候选 {candidate_ref or index}",
                    "locator": f"context:legal_candidate_completions[{index - 1}]",
                    "excerpt": (
                        f"candidate_ref={candidate_ref} "
                        f"target_port_key={target_port_key} "
                        f"is_tab_default={candidate.get('is_tab_default')} "
                        f"is_high_confidence={candidate.get('is_high_confidence')}"
                    ).strip(),
                }
            )

    naming_hints = context.get("naming_hints")
    if isinstance(naming_hints, dict) and naming_hints:
        citations.append(
            {
                "id": "ctx-naming-hints",
                "kind": "context",
                "label": "ghost 命名提示",
                "locator": "context:naming_hints",
                "excerpt": normalize_text(naming_hints)[:180],
            }
        )

    topology_pattern_hints = [
        str(item).strip()
        for item in (context.get("topology_pattern_hints") or [])
        if str(item).strip()
    ]
    if topology_pattern_hints:
        citations.append(
            {
                "id": "ctx-pattern",
                "kind": "context",
                "label": "拓扑模式提示",
                "locator": "context:topology_pattern_hints[0]",
                "excerpt": topology_pattern_hints[0],
            }
        )

    recent_actions = list(((context.get("cursor_context") or {}).get("recent_actions") or []))
    if recent_actions:
        latest_recent_action = recent_actions[-1] or {}
        citations.append(
            {
                "id": "ctx-recent-action",
                "kind": "context",
                "label": "最近 ghost 动作",
                "locator": f"context:cursor_context.recent_actions[{len(recent_actions) - 1}]",
                "excerpt": normalize_text(latest_recent_action)[:180],
            }
        )

    return citations


def build_citation_maps(copilot_request: dict[str, Any]) -> tuple[list[dict[str, Any]], dict[str, str]]:
    citations = build_citations(copilot_request)
    artifact_name_to_citation_id: dict[str, str] = {}
    for citation, artifact in zip(citations, copilot_request.get("artifacts") or []):
        artifact_name = str((artifact or {}).get("name") or "").strip()
        if artifact_name:
            artifact_name_to_citation_id[artifact_name] = str(citation.get("id") or "")
    return citations, artifact_name_to_citation_id


def classify_docs_qa_request(copilot_request: dict[str, Any], primary_text: str) -> tuple[str, str, list[dict[str, Any]]]:
    issues: list[dict[str, Any]] = []
    status = "ok"
    risk_level = "low"
    if any(marker in primary_text for marker in ("不能", "仍以", "证据不足", "不足以", "示例", "最终判定", "不应")):
        status = "partial"
        risk_level = "medium"
        code = "EVIDENCE_LIMITED"
        message = "当前文档给出了边界或限制，回答应保持克制，不把说明性口径写成确定性结论。"
        if any(marker in primary_text for marker in ("授权", "权限", "角色", "最终判定")):
            code = "AUTHORIZATION_NOT_PROVEN"
            message = "当前文档更像说明性边界，仍不足以替代最终权限判定。"
        issues.append(
            {
                "code": code,
                "message": message,
                "severity": "warning",
            }
        )
    viewer = (copilot_request.get("context") or {}).get("viewer") or {}
    if viewer and status == "ok":
        risk_level = "medium"
    return status, risk_level, issues


def make_mock_docs_qa_response(copilot_request: dict[str, Any]) -> dict[str, Any]:
    artifacts = copilot_request.get("artifacts") or []
    primary = next((artifact for artifact in artifacts if artifact.get("role") == "primary"), artifacts[0] if artifacts else {})
    primary_text = artifact_content_text(primary)
    summary = first_sentences(primary_text, max_sentences=1) or "当前没有足够文档内容可回答该问题。"
    answer_text = first_sentences(primary_text, max_sentences=2) or summary
    citations = build_citations(copilot_request)
    status, risk_level, issues = classify_docs_qa_request(copilot_request, primary_text)
    citation_ids = [citation["id"] for citation in citations[:1]]
    for issue in issues:
        issue["citation_ids"] = citation_ids
    return {
        "schema_version": 1,
        "status": status,
        "project": "radish",
        "task": "answer_docs_question",
        "summary": summary,
        "answers": [
            {
                "kind": "evidence_limited_answer" if status == "partial" else "direct_answer",
                "text": answer_text,
                "citation_ids": citation_ids,
            }
        ],
        "issues": issues,
        "proposed_actions": [],
        "citations": citations,
        "confidence": 0.72 if status == "partial" else 0.86,
        "risk_level": risk_level,
        "requires_confirmation": False,
    }


def load_system_prompt(copilot_request: dict[str, Any]) -> str:
    project = str(copilot_request.get("project") or "").strip()
    task = str(copilot_request.get("task") or "").strip()
    prompt_path = PROMPT_PATHS.get((project, task))
    if prompt_path is None:
        raise ValueError(f"unsupported prompt target: {project} / {task}")
    return prompt_path.read_text(encoding="utf-8").strip()


def build_docs_qa_messages(copilot_request: dict[str, Any]) -> list[dict[str, str]]:
    request_payload = json.dumps(copilot_request, ensure_ascii=False, indent=2)
    return [
        {"role": "system", "content": load_system_prompt(copilot_request)},
        {
            "role": "user",
            "content": (
                "请基于以下 CopilotRequest 生成一个 JSON 对象形式的 CopilotResponse。"
                "不要输出 markdown 代码块，也不要输出 JSON 之外的解释。\n\n"
                f"{request_payload}"
            ),
        },
    ]


def build_ghost_completion_messages(copilot_request: dict[str, Any]) -> list[dict[str, str]]:
    request_payload = json.dumps(copilot_request, ensure_ascii=False, indent=2)
    return [
        {"role": "system", "content": load_system_prompt(copilot_request)},
        {
            "role": "user",
            "content": (
                "请基于以下 CopilotRequest 生成一个 JSON 对象形式的 CopilotResponse。"
                "不要输出 markdown 代码块，也不要输出 JSON 之外的解释。\n\n"
                f"{request_payload}"
            ),
        },
    ]


def build_messages(copilot_request: dict[str, Any]) -> list[dict[str, str]]:
    project = str(copilot_request.get("project") or "").strip()
    task = str(copilot_request.get("task") or "").strip()
    if project == "radish" and task == "answer_docs_question":
        return build_docs_qa_messages(copilot_request)
    if project == "radishflow" and task == "suggest_ghost_completion":
        return build_ghost_completion_messages(copilot_request)
    raise ValueError(f"unsupported message target: {project} / {task}")


def pick_primary_ghost_candidate(legal_candidates: list[dict[str, Any]]) -> dict[str, Any] | None:
    for candidate in legal_candidates:
        conflict_flags = [str(flag).strip() for flag in (candidate.get("conflict_flags") or []) if str(flag).strip()]
        if candidate.get("is_tab_default") is True and candidate.get("is_high_confidence") is True and not conflict_flags:
            return candidate
    return legal_candidates[0] if legal_candidates else None


def build_ghost_completion_action(
    candidate: dict[str, Any],
    *,
    action_index: int,
    accept_key: str,
    risk_level: str,
    citation_ids: list[str],
) -> dict[str, Any]:
    candidate_ref = str(candidate.get("candidate_ref") or "").strip()
    target_unit_id = str(candidate.get("target_unit_id") or "").strip()
    target_port_key = str(candidate.get("target_port_key") or "").strip()
    suggested_stream_name = str(candidate.get("suggested_stream_name") or "").strip()
    target_label = target_port_key or candidate_ref or f"ghost-{action_index}"
    title_prefix = "补全" if accept_key == "Tab" else "保留"
    rationale_prefix = (
        "该候选当前被本地规则层标记为默认高置信 ghost，因此可作为首个建议。"
        if accept_key == "Tab"
        else "该候选当前仍然合法，但不满足默认 Tab 条件，因此仅保留为手动可选 ghost。"
    )
    return {
        "kind": "ghost_completion",
        "title": f"{title_prefix} {target_unit_id} 的 {target_label} ghost 候选",
        "target": {
            "type": "unit_port",
            "unit_id": target_unit_id,
            "port_key": target_port_key,
        },
        "rationale": rationale_prefix,
        "patch": {
            "ghost_kind": str(candidate.get("ghost_kind") or "ghost_connection"),
            "candidate_ref": candidate_ref,
            "target_port_key": target_port_key,
            **({"ghost_stream_name": suggested_stream_name} if suggested_stream_name else {}),
        },
        "preview": {
            "ghost_color": "gray",
            "accept_key": accept_key,
            "render_priority": action_index,
        },
        "apply": {
            "command_kind": "accept_ghost_completion",
            "payload": {
                "candidate_ref": candidate_ref,
            },
        },
        "risk_level": risk_level,
        "requires_confirmation": False,
        "citation_ids": citation_ids,
    }


def make_mock_ghost_completion_response(copilot_request: dict[str, Any]) -> dict[str, Any]:
    context = copilot_request.get("context") or {}
    selected_unit = context.get("selected_unit") or {}
    selected_unit_id = str(selected_unit.get("id") or "").strip() or str((context.get("selected_unit_ids") or [""])[0] or "")
    legal_candidates = [
        candidate
        for candidate in (context.get("legal_candidate_completions") or [])
        if isinstance(candidate, dict) and str(candidate.get("candidate_ref") or "").strip()
    ]
    citations = build_ghost_context_citations(copilot_request)
    citation_ids = [str(citation.get("id") or "") for citation in citations if str(citation.get("id") or "")]

    if not legal_candidates:
        return {
            "schema_version": 1,
            "status": "partial",
            "project": "radishflow",
            "task": "suggest_ghost_completion",
            "summary": f"{selected_unit_id or '当前单元'} 当前没有可确认的合法 ghost 候选，因此不返回默认补全建议。",
            "answers": [
                {
                    "kind": "empty_suggestion_reason",
                    "text": "当前上下文里没有本地规则确认的合法 ghost completion 候选，因此此时应保持空建议。",
                    "citation_ids": [citation_id for citation_id in citation_ids if citation_id in {"ctx-empty-candidates", "ctx-unit", "ctx-pattern", "ctx-recent-action"}],
                }
            ],
            "issues": [
                {
                    "code": "INSUFFICIENT_GHOST_CONTEXT",
                    "message": "当前没有本地规则确认的合法 ghost completion 候选，不能安全地产生默认或手动 ghost 建议。",
                    "severity": "info",
                    "citation_ids": [citation_id for citation_id in citation_ids if citation_id in {"ctx-empty-candidates", "ctx-pattern"}],
                }
            ],
            "proposed_actions": [],
            "citations": citations,
            "confidence": 0.9,
            "risk_level": "low",
            "requires_confirmation": False,
        }

    primary_candidate = pick_primary_ghost_candidate(legal_candidates)
    actions: list[dict[str, Any]] = []
    answers: list[dict[str, Any]] = []

    if primary_candidate and primary_candidate.get("is_tab_default") is True and primary_candidate.get("is_high_confidence") is True:
        actions.append(
            build_ghost_completion_action(
                primary_candidate,
                action_index=1,
                accept_key="Tab",
                risk_level="low",
                citation_ids=[citation_id for citation_id in citation_ids if citation_id in {"ctx-unit", "ctx-candidate-1", "ctx-naming-hints", "ctx-pattern"}] or citation_ids[:2],
            )
        )
        primary_port = str(primary_candidate.get("target_port_key") or "").strip() or "默认端口"
        answers.append(
            {
                "kind": "ghost_rationale",
                "text": f"{selected_unit_id or '当前单元'} 的 {primary_port} 当前是本地规则层给出的默认高置信候选，因此应作为首个 ghost 建议。",
                "citation_ids": [citation_id for citation_id in citation_ids if citation_id in {"ctx-candidate-1", "ctx-naming-hints", "ctx-unit"}] or citation_ids[:2],
            }
        )
        summary = f"{selected_unit_id or '当前单元'} 当前最适合作为默认 ghost 的候选已明确，可优先渲染首条 `Tab` 建议。"
        confidence = 0.93
        risk_level = "low"
    else:
        for index, candidate in enumerate(legal_candidates[:2], start=1):
            actions.append(
                build_ghost_completion_action(
                    candidate,
                    action_index=index,
                    accept_key="manual_only",
                    risk_level="medium",
                    citation_ids=[citation_id for citation_id in citation_ids if citation_id in {f"ctx-candidate-{index}", "ctx-pattern"}] or citation_ids[:2],
                )
            )
        answers.append(
            {
                "kind": "ghost_rationale",
                "text": "当前合法候选之间没有形成足够清晰的领先者，因此只应返回可见 ghost，而不应把任何一条候选直接升级成默认 Tab。",
                "citation_ids": [citation_id for citation_id in citation_ids if citation_id in {"ctx-candidate-1", "ctx-candidate-2", "ctx-pattern"}] or citation_ids[:2],
            }
        )
        summary = f"{selected_unit_id or '当前单元'} 当前存在多个合法 ghost 候选，但证据不足以默认选中其中任意一条。"
        confidence = 0.78
        risk_level = "medium"

    return {
        "schema_version": 1,
        "status": "ok",
        "project": "radishflow",
        "task": "suggest_ghost_completion",
        "summary": summary,
        "answers": answers,
        "issues": [],
        "proposed_actions": actions,
        "citations": citations,
        "confidence": confidence,
        "risk_level": risk_level,
        "requires_confirmation": False,
    }


def make_failed_response(copilot_request: dict[str, Any], summary: str, raw_text: str, code: str) -> dict[str, Any]:
    preview = normalize_text(raw_text)[:360]
    return {
        "schema_version": 1,
        "status": "failed",
        "project": copilot_request.get("project"),
        "task": copilot_request.get("task"),
        "summary": summary,
        "answers": [
            {
                "kind": "raw_text_fallback",
                "text": preview or "模型未返回可用文本。",
                "citation_ids": [],
            }
        ],
        "issues": [
            {
                "code": code,
                "message": summary,
                "severity": "warning",
                "citation_ids": [],
            }
        ],
        "proposed_actions": [],
        "citations": [],
        "confidence": 0.05,
        "risk_level": "medium",
        "requires_confirmation": False,
    }


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


def normalize_citation_ids(value: Any, artifact_name_to_citation_id: dict[str, str]) -> list[str]:
    normalized_ids: list[str] = []
    for item in value or []:
        if isinstance(item, dict):
            artifact_index = item.get("artifact_index")
            if isinstance(artifact_index, int):
                fallback_id = f"doc-{artifact_index + 1}"
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
                ),
            }
        )
    return normalized_issues


def normalize_actions(
    actions: Any,
    artifact_name_to_citation_id: dict[str, str],
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
        locator = normalize_text(citation.get("locator")) or citation_fields.get("locator") or (
            f"artifact:{artifact_name}" if artifact_name else ""
        )
        if locator:
            normalized_citation["locator"] = locator
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
    fallback_citations, artifact_name_to_citation_id = build_citation_maps(copilot_request)
    project = str(copilot_request.get("project") or "").strip()
    task = str(copilot_request.get("task") or "").strip()
    embedded_actions = extract_answer_embedded_actions(coerced.get("answers"))
    if embedded_actions and not isinstance(coerced.get("proposed_actions"), list):
        coerced["proposed_actions"] = embedded_actions
    elif embedded_actions and not (coerced.get("proposed_actions") or []):
        coerced["proposed_actions"] = embedded_actions
    coerced["status"] = map_status(coerced.get("status"))
    coerced["answers"] = normalize_answers(coerced.get("answers"), artifact_name_to_citation_id)
    coerced["issues"] = normalize_issues(coerced.get("issues"), artifact_name_to_citation_id)
    coerced["proposed_actions"] = normalize_actions(
        coerced.get("proposed_actions"),
        artifact_name_to_citation_id,
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


def normalize_openai_content(content: str, copilot_request: dict[str, Any]) -> dict[str, Any]:
    normalized = content.strip()
    if normalized.startswith("```"):
        fenced = re.findall(r"```(?:json)?\s*(.*?)```", normalized, flags=re.DOTALL)
        if fenced:
            normalized = fenced[0].strip()
    first_brace = normalized.find("{")
    last_brace = normalized.rfind("}")
    if first_brace != -1 and last_brace != -1 and last_brace > first_brace:
        candidate = normalized[first_brace : last_brace + 1]
        try:
            document = json.loads(candidate)
        except json.JSONDecodeError:
            document = None
            if (
                str(copilot_request.get("project") or "").strip() == "radishflow"
                and str(copilot_request.get("task") or "").strip() == "suggest_ghost_completion"
            ):
                repaired_candidate = repair_malformed_ghost_json(candidate)
                if repaired_candidate != candidate:
                    try:
                        document = json.loads(repaired_candidate)
                    except json.JSONDecodeError:
                        document = None
        if isinstance(document, dict):
            return coerce_response_document(document, copilot_request, raw_text=content)
    return make_failed_response(
        copilot_request,
        "模型未返回可解析的 JSON，当前只保留原始文本供调试。",
        raw_text=content,
        code="MODEL_OUTPUT_NOT_JSON",
    )


def repair_malformed_ghost_json(candidate: str) -> str:
    manual_multi_action_repaired = candidate
    for pattern, replacement in GHOST_MANUAL_MULTI_ACTION_REPAIR_PATTERNS:
        manual_multi_action_repaired = pattern.sub(replacement, manual_multi_action_repaired)
    if manual_multi_action_repaired != candidate:
        try:
            json.loads(manual_multi_action_repaired)
            return manual_multi_action_repaired
        except json.JSONDecodeError:
            pass

    repaired = candidate
    for _ in range(4):
        previous = repaired
        for pattern, replacement in GHOST_MALFORMED_JSON_REPAIR_PATTERNS:
            repaired = pattern.sub(replacement, repaired)
        if repaired == previous:
            break
    return repaired


def extract_embedded_summary_text(value: Any) -> str:
    normalized_value = normalize_text(value)
    if not normalized_value or not normalized_value.startswith("{") or not normalized_value.endswith("}"):
        return ""
    try:
        parsed = json.loads(normalized_value)
    except json.JSONDecodeError:
        return ""
    if not isinstance(parsed, dict):
        return ""
    return normalize_text(parsed.get("summary"))


def resolve_chat_endpoint(base_url: str) -> str:
    trimmed = base_url.rstrip("/")
    if trimmed.endswith("/chat/completions"):
        return trimmed
    if trimmed.endswith("/v1"):
        return trimmed + "/chat/completions"
    return trimmed + "/v1/chat/completions"


def resolve_gemini_generate_content_endpoint(base_url: str, model: str) -> str:
    trimmed = base_url.rstrip("/")
    if trimmed.endswith(":generateContent"):
        return trimmed
    if trimmed.endswith(f"/models/{model}"):
        return trimmed + ":generateContent"
    return f"{trimmed}/models/{model}:generateContent"


def extract_openai_message_content(payload: dict[str, Any]) -> str:
    choices = payload.get("choices") or []
    if not choices:
        return ""
    message = (choices[0] or {}).get("message") or {}
    content = message.get("content")
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        text_parts = []
        for item in content:
            if not isinstance(item, dict):
                continue
            if item.get("type") == "text" and isinstance(item.get("text"), str):
                text_parts.append(item["text"])
        return "\n".join(text_parts).strip()
    return ""


def extract_gemini_message_content(payload: dict[str, Any]) -> str:
    candidates = payload.get("candidates") or []
    if not candidates:
        return ""
    content = (candidates[0] or {}).get("content") or {}
    parts = content.get("parts") or []
    text_parts: list[str] = []
    for item in parts:
        if not isinstance(item, dict):
            continue
        if isinstance(item.get("text"), str):
            text_parts.append(item["text"])
    return "\n".join(text_parts).strip()


def extract_provider_message_content(payload: dict[str, Any]) -> str:
    content = extract_openai_message_content(payload)
    if content.strip():
        return content
    return extract_gemini_message_content(payload)


def post_json_request(
    *,
    endpoint: str,
    payload: dict[str, Any],
    headers: dict[str, str],
    request_timeout_seconds: float,
) -> dict[str, Any]:
    data = json.dumps(payload).encode("utf-8")
    http_request = request.Request(
        endpoint,
        data=data,
        headers=headers,
        method="POST",
    )
    try:
        with request.urlopen(http_request, timeout=request_timeout_seconds) as response_obj:
            raw_body = response_obj.read().decode("utf-8")
    except error.HTTPError as exc:
        error_body = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"provider request failed with HTTP {exc.code}: {error_body}") from exc
    except (error.URLError, TimeoutError, http.client.RemoteDisconnected, http.client.IncompleteRead) as exc:
        raise RuntimeError(f"provider request failed: {exc}") from exc
    return json.loads(raw_body)


def call_openai_compatible(
    copilot_request: dict[str, Any],
    *,
    model: str,
    base_url: str,
    api_key: str,
    temperature: float,
    request_timeout_seconds: float,
) -> dict[str, Any]:
    messages = build_messages(copilot_request)
    endpoint = resolve_chat_endpoint(base_url)
    payload = {
        "model": model,
        "temperature": temperature,
        "messages": messages,
    }
    raw_request = {
        "endpoint": endpoint,
        "payload": payload,
        "transport": "openai-chat-completions",
    }
    raw_response = post_json_request(
        endpoint=endpoint,
        payload=payload,
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {api_key}",
        },
        request_timeout_seconds=request_timeout_seconds,
    )
    content = extract_openai_message_content(raw_response)
    response_document = normalize_openai_content(content, copilot_request)
    validate_response_document(response_document)
    return {
        "provider": "openai-compatible",
        "model": model,
        "messages": messages,
        "raw_request": raw_request,
        "raw_response": raw_response,
        "response": response_document,
    }


def build_gemini_payload(messages: list[dict[str, str]], temperature: float) -> dict[str, Any]:
    system_texts: list[str] = []
    contents: list[dict[str, Any]] = []
    for message in messages:
        role = str(message.get("role") or "").strip()
        content = str(message.get("content") or "").strip()
        if not content:
            continue
        if role == "system":
            system_texts.append(content)
            continue
        gemini_role = "model" if role == "assistant" else "user"
        contents.append(
            {
                "role": gemini_role,
                "parts": [
                    {
                        "text": content,
                    }
                ],
            }
        )
    payload: dict[str, Any] = {
        "contents": contents,
        "generationConfig": {
            "temperature": temperature,
        },
    }
    if system_texts:
        payload["system_instruction"] = {
            "parts": [{"text": text} for text in system_texts],
        }
    return payload


def call_gemini_native(
    copilot_request: dict[str, Any],
    *,
    model: str,
    base_url: str,
    api_key: str,
    temperature: float,
    request_timeout_seconds: float,
) -> dict[str, Any]:
    messages = build_messages(copilot_request)
    endpoint = resolve_gemini_generate_content_endpoint(base_url, model)
    payload = build_gemini_payload(messages, temperature)
    raw_request = {
        "endpoint": endpoint,
        "payload": payload,
        "transport": "gemini-generate-content",
    }
    raw_response = post_json_request(
        endpoint=endpoint,
        payload=payload,
        headers={
            "Content-Type": "application/json",
            "x-goog-api-key": api_key,
        },
        request_timeout_seconds=request_timeout_seconds,
    )
    content = extract_gemini_message_content(raw_response)
    response_document = normalize_openai_content(content, copilot_request)
    validate_response_document(response_document)
    return {
        "provider": "openai-compatible",
        "model": model,
        "messages": messages,
        "raw_request": raw_request,
        "raw_response": raw_response,
        "response": response_document,
    }


def run_inference(
    copilot_request: dict[str, Any],
    *,
    provider: str,
    provider_profile: str | None = None,
    model: str | None = None,
    base_url: str | None = None,
    api_key: str | None = None,
    temperature: float = 0.0,
    request_timeout_seconds: float | None = None,
) -> dict[str, Any]:
    load_env_file(ENV_FILE_PATH)
    validate_request_document(copilot_request)
    project = str(copilot_request.get("project") or "").strip()
    task = str(copilot_request.get("task") or "").strip()

    if provider == "mock":
        if project == "radish" and task == "answer_docs_question":
            response_document = make_mock_docs_qa_response(copilot_request)
        elif project == "radishflow" and task == "suggest_ghost_completion":
            response_document = make_mock_ghost_completion_response(copilot_request)
        else:
            raise ValueError(f"mock provider does not support {project} / {task}")
        validate_response_document(response_document)
        return {
            "provider": "mock",
            "model": model or f"radishmind-mock-{project}-{task}-v1",
            "messages": build_messages(copilot_request),
            "raw_request": {
                "provider": "mock",
                "mode": "deterministic-rule-based",
            },
            "raw_response": {
                "provider": "mock",
                "response_preview": response_document["summary"],
            },
            "response": response_document,
        }

    if provider == "openai-compatible":
        resolved = resolve_openai_compatible_config(
            provider_profile=provider_profile,
            model=model,
            base_url=base_url,
            api_key=api_key,
            request_timeout_seconds=request_timeout_seconds,
        )
        resolved_profile = str(resolved["profile"])
        normalized_profile = str(resolved["normalized_profile"])
        resolved_api_style = str(resolved["api_style"] or "").strip() or infer_profile_api_style(str(resolved["base_url"]))
        resolved_model = str(resolved["model"])
        resolved_base_url = str(resolved["base_url"])
        resolved_api_key = str(resolved["api_key"])
        resolved_request_timeout_seconds = float(resolved["request_timeout_seconds"])
        profile_api_style_key = profile_env_key(resolved_profile, "API_STYLE")
        profile_name_key = profile_env_key(resolved_profile, "NAME")
        profile_base_url_key = profile_env_key(resolved_profile, "BASE_URL")
        profile_api_key_key = profile_env_key(resolved_profile, "API_KEY")
        if resolved_api_style not in {"openai-compatible", "gemini-native"}:
            raise ValueError(
                "provider=openai-compatible requires a supported api style via "
                f"{profile_api_style_key} or RADISHMIND_MODEL_API_STYLE"
            )
        if not resolved_model:
            raise ValueError(
                "provider=openai-compatible requires --model, "
                f"{profile_name_key}, or RADISHMIND_MODEL_NAME"
            )
        if not resolved_base_url:
            raise ValueError(
                "provider=openai-compatible requires --base-url, "
                f"{profile_base_url_key}, or RADISHMIND_MODEL_BASE_URL"
            )
        if not resolved_api_key:
            raise ValueError(
                "provider=openai-compatible requires --api-key, "
                f"{profile_api_key_key}, or RADISHMIND_MODEL_API_KEY"
            )
        if resolved_request_timeout_seconds <= 0:
            raise ValueError("provider=openai-compatible requires a positive request timeout")
        if resolved_api_style == "gemini-native":
            result = call_gemini_native(
                copilot_request,
                model=resolved_model,
                base_url=resolved_base_url,
                api_key=resolved_api_key,
                temperature=temperature,
                request_timeout_seconds=resolved_request_timeout_seconds,
            )
        else:
            result = call_openai_compatible(
                copilot_request,
                model=resolved_model,
                base_url=resolved_base_url,
                api_key=resolved_api_key,
                temperature=temperature,
                request_timeout_seconds=resolved_request_timeout_seconds,
            )
        raw_request = result.get("raw_request")
        if isinstance(raw_request, dict):
            raw_request["provider_profile"] = resolved_profile
            raw_request["provider_profile_env"] = normalized_profile
            raw_request["api_style"] = resolved_api_style
        return result

    raise ValueError(f"unsupported provider: {provider}")


def build_candidate_response_dump(
    *,
    copilot_request: dict[str, Any],
    sample_id: str,
    inference_result: dict[str, Any],
    dump_id: str | None = None,
    record_id: str | None = None,
    captured_at: str | None = None,
    capture_origin: str | None = None,
    collection_batch: str | None = None,
    tags: list[str] | None = None,
    notes: str | None = None,
) -> dict[str, Any]:
    request_id = str(copilot_request.get("request_id") or "").strip() or f"{copilot_request['task']}-request"
    source = "simulated_candidate_response" if inference_result["provider"] == "mock" else "captured_candidate_response"
    default_capture_origin = "manual_fixture" if inference_result["provider"] == "mock" else "adapter_debug_dump"
    project = str(copilot_request.get("project") or "").strip()
    task = str(copilot_request.get("task") or "").strip()
    merged_tags: list[str] = [f"{project}_{task}", inference_result["provider"]]
    if source == "captured_candidate_response":
        merged_tags.append("real_capture")
    for tag in tags or []:
        normalized_tag = str(tag).strip()
        if normalized_tag and normalized_tag not in merged_tags:
            merged_tags.append(normalized_tag)
    capture_metadata: dict[str, Any] = {
        "capture_origin": capture_origin or default_capture_origin,
        "tags": merged_tags,
    }
    if collection_batch:
        capture_metadata["collection_batch"] = collection_batch
    if notes:
        capture_metadata["notes"] = notes
    return {
        "schema_version": 1,
        "dump_id": dump_id or f"dump-{request_id}",
        **({"record_id": record_id} if record_id else {}),
        "project": copilot_request["project"],
        "task": copilot_request["task"],
        "sample_id": sample_id,
        "request_id": request_id,
        "captured_at": captured_at or utc_now_iso(),
        "source": source,
        "capture_metadata": capture_metadata,
        "model": inference_result["model"],
        "input_record": derive_input_record(copilot_request),
        "input_request": copilot_request,
        "raw_request": inference_result.get("raw_request"),
        "raw_response": inference_result.get("raw_response"),
        "response": inference_result["response"],
    }


def recanonicalize_response_dump(
    dump: dict[str, Any],
    *,
    label: str = "candidate response dump",
) -> dict[str, Any]:
    copilot_request = dump.get("input_request")
    if not isinstance(copilot_request, dict):
        raise ValueError(f"{label}: input_request is required to recanonicalize response")
    validate_request_document(copilot_request)

    raw_response = dump.get("raw_response")
    if not isinstance(raw_response, dict):
        raise ValueError(f"{label}: raw_response is required to recanonicalize response")

    content = extract_provider_message_content(raw_response)
    if not content.strip():
        raise ValueError(f"{label}: raw_response does not contain assistant content")

    normalized_response = normalize_openai_content(content, copilot_request)
    validate_response_document(normalized_response)

    updated_dump = dict(dump)
    updated_dump["response"] = normalized_response
    return updated_dump
