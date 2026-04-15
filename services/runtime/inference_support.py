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
    ("radishflow", "suggest_flowsheet_edits"): REPO_ROOT / "prompts/tasks/radishflow-suggest-flowsheet-edits-system.md",
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


def resolve_openai_compatible_profile_chain(provider_profile: str | None) -> list[str]:
    explicit_profile = str(provider_profile or "").strip()
    if explicit_profile:
        return [explicit_profile]

    configured_profiles = [
        str(item).strip()
        for item in getenv_stripped("RADISHMIND_MODEL_PROFILE_FALLBACKS").split(",")
        if str(item).strip()
    ]
    default_profile = resolve_openai_compatible_profile(None)

    ordered_profiles: list[str] = []
    for profile in [*configured_profiles, default_profile]:
        if profile and profile not in ordered_profiles:
            ordered_profiles.append(profile)
    return ordered_profiles or [default_profile]


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
    if str(copilot_request.get("project") or "") == "radishflow" and str(copilot_request.get("task") or "") == "suggest_flowsheet_edits":
        selected_unit_ids = [
            str(unit_id).strip()
            for unit_id in (context.get("selected_unit_ids") or [])
            if str(unit_id).strip()
        ]
        selected_stream_ids = [
            str(stream_id).strip()
            for stream_id in (context.get("selected_stream_ids") or [])
            if str(stream_id).strip()
        ]
        diagnostic_codes = [
            str((diagnostic or {}).get("code") or "").strip()
            for diagnostic in (context.get("diagnostics") or [])
            if str((diagnostic or {}).get("code") or "").strip()
        ]
        input_record["document_revision"] = context.get("document_revision")
        if selected_unit_ids:
            input_record["selected_unit_ids"] = selected_unit_ids
        if selected_stream_ids:
            input_record["selected_stream_ids"] = selected_stream_ids
        if diagnostic_codes:
            input_record["diagnostic_codes"] = diagnostic_codes
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
    if project == "radishflow" and task == "suggest_flowsheet_edits":
        return build_suggest_edits_context_citations(copilot_request)
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


def summarize_flowdoc_object(object_type: str, object_document: dict[str, Any]) -> str:
    object_id = str(object_document.get("id") or "").strip()
    object_name = str(object_document.get("name") or "").strip()
    if object_type == "stream":
        parts = [
            object_id or "stream",
            object_name,
            f"source_unit_id={object_document.get('source_unit_id')}" if object_document.get("source_unit_id") is not None else "",
            f"target_unit_id={object_document.get('target_unit_id')}" if object_document.get("target_unit_id") is not None else "",
        ]
    else:
        parts = [
            object_id or "unit",
            object_name,
            f"kind={object_document.get('kind')}" if object_document.get("kind") is not None else "",
        ]
    summary = " ".join(part for part in parts if part).strip()
    if not summary:
        summary = normalize_text(object_document)
    return summary[:180]


def build_suggest_edits_context_citations(copilot_request: dict[str, Any]) -> list[dict[str, Any]]:
    context = copilot_request.get("context") or {}
    artifacts = copilot_request.get("artifacts") or []
    citations: list[dict[str, Any]] = []

    for index, diagnostic in enumerate(context.get("diagnostics") or [], start=1):
        if not isinstance(diagnostic, dict):
            continue
        code = str(diagnostic.get("code") or "").strip()
        message = normalize_text(diagnostic.get("message") or diagnostic.get("text") or diagnostic)
        citations.append(
            {
                "id": f"diag-{index}",
                "kind": "context",
                "label": f"诊断 {code or index}",
                "locator": f"context:diagnostics[{index - 1}]",
                "excerpt": message[:180] if message else f"diagnostic {index}",
            }
        )

    primary_flowdoc = next(
        (
            artifact
            for artifact in artifacts
            if str((artifact or {}).get("name") or "").strip() == "flowsheet_document"
            and isinstance((artifact or {}).get("content"), dict)
        ),
        None,
    )
    if isinstance(primary_flowdoc, dict):
        flowdoc_content = primary_flowdoc.get("content") or {}
        units = list(flowdoc_content.get("units") or [])
        streams = list(flowdoc_content.get("streams") or [])
        stream_id_to_entry = {
            str((stream or {}).get("id") or "").strip(): (stream_index, stream)
            for stream_index, stream in enumerate(streams)
            if isinstance(stream, dict) and str((stream or {}).get("id") or "").strip()
        }
        unit_id_to_entry = {
            str((unit or {}).get("id") or "").strip(): (unit_index, unit)
            for unit_index, unit in enumerate(units)
            if isinstance(unit, dict) and str((unit or {}).get("id") or "").strip()
        }
        ordered_targets: list[tuple[str, str]] = []
        for diagnostic in context.get("diagnostics") or []:
            if not isinstance(diagnostic, dict):
                continue
            target_type = str(diagnostic.get("target_type") or "").strip().lower()
            target_id = str(diagnostic.get("target_id") or "").strip()
            if target_type in {"stream", "unit"} and target_id:
                candidate = (target_type, target_id)
                if candidate not in ordered_targets:
                    ordered_targets.append(candidate)
        for stream_id in context.get("selected_stream_ids") or []:
            normalized_stream_id = str(stream_id).strip()
            if normalized_stream_id and ("stream", normalized_stream_id) not in ordered_targets:
                ordered_targets.append(("stream", normalized_stream_id))
        for unit_id in context.get("selected_unit_ids") or []:
            normalized_unit_id = str(unit_id).strip()
            if normalized_unit_id and ("unit", normalized_unit_id) not in ordered_targets:
                ordered_targets.append(("unit", normalized_unit_id))

        ordered_target_set = set(ordered_targets)
        for stream_index, stream_document in enumerate(streams):
            if not isinstance(stream_document, dict):
                continue
            target_id = str(stream_document.get("id") or "").strip()
            if not target_id or ("stream", target_id) not in ordered_target_set:
                continue
            citations.append(
                {
                    "id": f"flowdoc-stream-{stream_index + 1}",
                    "kind": "artifact",
                    "label": f"FlowsheetDocument / {target_id}",
                    "locator": f"artifact:flowsheet_document.streams[{stream_index}]",
                    "excerpt": summarize_flowdoc_object("stream", stream_document),
                }
            )
        for unit_index, unit_document in enumerate(units):
            if not isinstance(unit_document, dict):
                continue
            target_id = str(unit_document.get("id") or "").strip()
            if not target_id or ("unit", target_id) not in ordered_target_set:
                continue
            citations.append(
                {
                    "id": f"flowdoc-unit-{unit_index + 1}",
                    "kind": "artifact",
                    "label": f"FlowsheetDocument / {target_id}",
                    "locator": f"artifact:flowsheet_document.units[{unit_index}]",
                    "excerpt": summarize_flowdoc_object("unit", unit_document),
                }
            )

    latest_snapshot = context.get("latest_snapshot")
    if isinstance(latest_snapshot, dict) and latest_snapshot:
        excerpt_parts = [
            f"solver_status={latest_snapshot.get('solver_status')}" if latest_snapshot.get("solver_status") is not None else "",
            f"residual_norm={latest_snapshot.get('residual_norm')}" if latest_snapshot.get("residual_norm") is not None else "",
        ]
        citations.append(
            {
                "id": "snapshot-1",
                "kind": "context",
                "label": "求解快照 latest_snapshot",
                "locator": "context:latest_snapshot",
                "excerpt": " ".join(part for part in excerpt_parts if part).strip() or normalize_text(latest_snapshot)[:180],
            }
        )

    diagnostic_summary = context.get("diagnostic_summary")
    if isinstance(diagnostic_summary, dict) and diagnostic_summary and not any(
        str(citation.get("id") or "") == "snapshot-1" for citation in citations
    ):
        citations.append(
            {
                "id": "diag-summary-1",
                "kind": "context",
                "label": "诊断摘要 diagnostic_summary",
                "locator": "context:diagnostic_summary",
                "excerpt": normalize_text(diagnostic_summary)[:180],
            }
        )

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


def build_citation_maps(copilot_request: dict[str, Any]) -> tuple[list[dict[str, Any]], dict[str, str], dict[int, str]]:
    citations = build_citations(copilot_request)
    artifact_name_to_citation_id: dict[str, str] = {}
    artifact_index_to_citation_id: dict[int, str] = {}
    for citation, artifact in zip(citations, copilot_request.get("artifacts") or []):
        artifact_name = str((artifact or {}).get("name") or "").strip()
        citation_id = str(citation.get("id") or "")
        if artifact_name:
            artifact_name_to_citation_id[artifact_name] = citation_id
        artifact_index_to_citation_id[len(artifact_index_to_citation_id)] = citation_id
    return citations, artifact_name_to_citation_id, artifact_index_to_citation_id


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


def infer_stream_spec_placeholders(message: str) -> list[str]:
    lowered = message.lower()
    placeholders: list[str] = []
    mapping = [
        ("flow rate", "flow_rate_kg_per_h"),
        ("mass flow", "flow_rate_kg_per_h"),
        ("temperature", "temperature_c"),
        ("pressure", "pressure_kpa"),
        ("composition", "composition"),
        ("vapor fraction", "vapor_fraction"),
    ]
    for marker, placeholder in mapping:
        if marker in lowered and placeholder not in placeholders:
            placeholders.append(placeholder)
    return placeholders or ["flow_rate_kg_per_h"]


def infer_parameter_placeholders(message: str) -> list[str]:
    lowered = message.lower()
    placeholders: list[str] = []
    mapping = [
        ("operating pressure", "operating_pressure_kpa"),
        ("pressure target", "operating_pressure_kpa"),
        ("outlet temperature", "outlet_temperature_c"),
        ("temperature target", "outlet_temperature_c"),
        ("efficiency", "efficiency"),
        ("pressure ratio", "pressure_ratio"),
        ("duty", "duty_kw"),
        ("reflux", "reflux_ratio"),
    ]
    for marker, placeholder in mapping:
        if marker in lowered and placeholder not in placeholders:
            placeholders.append(placeholder)
    return placeholders or ["review_required_parameter"]


def build_suggest_edits_mock_action(
    diagnostic: dict[str, Any],
    *,
    artifact_citation_by_target: dict[tuple[str, str], str],
    snapshot_citation_id: str,
) -> dict[str, Any]:
    code = str(diagnostic.get("code") or "").strip()
    message = normalize_text(diagnostic.get("message") or diagnostic.get("text") or "")
    target_type = str(diagnostic.get("target_type") or "").strip().lower()
    target_id = str(diagnostic.get("target_id") or "").strip()
    if target_type not in {"stream", "unit"}:
        target_type = "stream" if "stream" in code.lower() else "unit"
    citation_ids = [citation_id for citation_id in [f"diag-{diagnostic.get('_index', 1)}", artifact_citation_by_target.get((target_type, target_id)), snapshot_citation_id] if citation_id]

    if code == "STREAM_DISCONNECTED":
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
    if code == "STREAM_SPEC_MISSING":
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


def make_mock_suggest_edits_response(copilot_request: dict[str, Any]) -> dict[str, Any]:
    context = copilot_request.get("context") or {}
    diagnostics = [diagnostic for diagnostic in (context.get("diagnostics") or []) if isinstance(diagnostic, dict)]
    citations = build_suggest_edits_context_citations(copilot_request)
    artifact_citation_by_target: dict[tuple[str, str], str] = {}
    snapshot_citation_id = ""
    for citation in citations:
        citation_id = str(citation.get("id") or "").strip()
        locator = str(citation.get("locator") or "").strip()
        if citation_id == "snapshot-1":
            snapshot_citation_id = citation_id
            continue
        if citation_id.startswith("flowdoc-stream-"):
            target_id = str(citation.get("label") or "").split("/")[-1].strip()
            if target_id:
                artifact_citation_by_target[("stream", target_id)] = citation_id
        if citation_id.startswith("flowdoc-unit-"):
            target_id = str(citation.get("label") or "").split("/")[-1].strip()
            if target_id:
                artifact_citation_by_target[("unit", target_id)] = citation_id
        if locator.startswith("artifact:flowsheet_document.streams[") and citation_id.startswith("flowdoc-stream-"):
            target_id = str(citation.get("label") or "").split("/")[-1].strip()
            if target_id:
                artifact_citation_by_target[("stream", target_id)] = citation_id
        if locator.startswith("artifact:flowsheet_document.units[") and citation_id.startswith("flowdoc-unit-"):
            target_id = str(citation.get("label") or "").split("/")[-1].strip()
            if target_id:
                artifact_citation_by_target[("unit", target_id)] = citation_id

    issues: list[dict[str, Any]] = []
    proposed_actions: list[dict[str, Any]] = []
    for index, diagnostic in enumerate(diagnostics, start=1):
        code = str(diagnostic.get("code") or "").strip()
        message = normalize_text(diagnostic.get("message") or diagnostic.get("text") or "")
        target_type = str(diagnostic.get("target_type") or "").strip().lower()
        target_id = str(diagnostic.get("target_id") or "").strip()
        citation_ids = [
            citation_id
            for citation_id in [
                f"diag-{index}",
                artifact_citation_by_target.get((target_type, target_id)),
                snapshot_citation_id,
            ]
            if citation_id
        ]
        issues.append(
            {
                "code": code,
                "message": message or f"{target_id or '当前对象'} 存在需要人工复核的编辑问题。",
                "severity": str(diagnostic.get("severity") or "warning").strip().lower() or "warning",
                "citation_ids": citation_ids or [f"diag-{index}"],
            }
        )
        action_diagnostic = dict(diagnostic)
        action_diagnostic["_index"] = index
        proposed_actions.append(
            build_suggest_edits_mock_action(
                action_diagnostic,
                artifact_citation_by_target=artifact_citation_by_target,
                snapshot_citation_id=snapshot_citation_id,
            )
        )

    if not proposed_actions:
        return {
            "schema_version": 1,
            "status": "failed",
            "project": "radishflow",
            "task": "suggest_flowsheet_edits",
            "summary": "当前请求缺少可直接驱动候选编辑的结构化诊断，因此无法生成稳妥的 candidate_edit。",
            "answers": [
                {
                    "kind": "edit_rationale",
                    "text": "缺少结构化 diagnostics 时，不应凭主观猜测输出 patch。",
                    "citation_ids": [],
                }
            ],
            "issues": [
                {
                    "code": "DIAGNOSTICS_REQUIRED",
                    "message": "当前请求没有足够的 diagnostics 证据来支撑 candidate_edit。",
                    "severity": "warning",
                    "citation_ids": [],
                }
            ],
            "proposed_actions": [],
            "citations": citations,
            "confidence": 0.2,
            "risk_level": "medium",
            "requires_confirmation": False,
        }

    highest_risk = "low"
    if any(str(action.get("risk_level") or "") == "high" for action in proposed_actions):
        highest_risk = "high"
    elif any(str(action.get("risk_level") or "") == "medium" for action in proposed_actions):
        highest_risk = "medium"
    answer_citation_ids = list(
        dict.fromkeys(
            citation_id
            for action in proposed_actions
            for citation_id in (action.get("citation_ids") or [])
            if str(citation_id).strip()
        )
    )
    first_target = str(((proposed_actions[0] or {}).get("target") or {}).get("id") or "").strip()
    summary = (
        f"当前更合适的是围绕 {first_target or '当前对象'} 生成可审查的候选编辑提案。"
        if len(proposed_actions) == 1
        else "当前更合适的是按诊断优先级输出多条局部 candidate_edit，并保持证据与动作顺序稳定。"
    )
    return {
        "schema_version": 1,
        "status": "partial",
        "project": "radishflow",
        "task": "suggest_flowsheet_edits",
        "summary": summary,
        "answers": [
            {
                "kind": "edit_rationale",
                "text": "这些提案都只保留为候选 patch：优先直接响应明确诊断，再补目标对象证据，最后才引用 supporting snapshot。",
                "citation_ids": answer_citation_ids,
            }
        ],
        "issues": issues,
        "proposed_actions": proposed_actions,
        "citations": citations,
        "confidence": 0.81 if highest_risk == "high" else 0.84,
        "risk_level": highest_risk,
        "requires_confirmation": True,
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


def build_suggest_edits_messages(copilot_request: dict[str, Any]) -> list[dict[str, str]]:
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
    if project == "radishflow" and task == "suggest_flowsheet_edits":
        return build_suggest_edits_messages(copilot_request)
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
