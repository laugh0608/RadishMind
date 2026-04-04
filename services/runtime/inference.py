from __future__ import annotations

import json
import os
import re
from datetime import datetime, timezone
from pathlib import Path
from typing import Any
from urllib import error, request

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent.parent
REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-request.schema.json"
RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
PROMPT_PATH = REPO_ROOT / "prompts/tasks/radish-answer-docs-question-system.md"
SCHEMA_CACHE: dict[Path, Any] = {}
SENTENCE_BREAK_RE = re.compile(r"(?<=[。！？.!?])\s+")
ENV_FILE_PATH = REPO_ROOT / ".env"


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


def artifact_excerpt(artifact: dict[str, Any], limit: int = 180) -> str:
    text = artifact_content_text(artifact)
    if len(text) <= limit:
        return text
    return text[: limit - 1].rstrip() + "…"


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
    citations: list[dict[str, Any]] = []
    counters: dict[str, int] = {}
    context = copilot_request.get("context") or {}
    resource = context.get("resource") or {}
    for artifact in copilot_request.get("artifacts") or []:
        prefix = citation_prefix_for_artifact(artifact)
        counters[prefix] = counters.get(prefix, 0) + 1
        citation_id = f"{prefix}-{counters[prefix]}"
        label = (
            str((artifact.get("metadata") or {}).get("page_slug") or "").strip()
            or str(resource.get("title") or "").strip()
            or str(artifact.get("name") or "").strip()
            or citation_id
        )
        citation = {
            "id": citation_id,
            "kind": "artifact",
            "label": label,
            "locator": f"artifact:{artifact.get('name')}",
        }
        excerpt = artifact_excerpt(artifact)
        if excerpt:
            citation["excerpt"] = excerpt
        citations.append(citation)
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


def load_system_prompt() -> str:
    return PROMPT_PATH.read_text(encoding="utf-8").strip()


def build_docs_qa_messages(copilot_request: dict[str, Any]) -> list[dict[str, str]]:
    request_payload = json.dumps(copilot_request, ensure_ascii=False, indent=2)
    return [
        {"role": "system", "content": load_system_prompt()},
        {
            "role": "user",
            "content": (
                "请基于以下 CopilotRequest 生成一个 JSON 对象形式的 CopilotResponse。"
                "不要输出 markdown 代码块，也不要输出 JSON 之外的解释。\n\n"
                f"{request_payload}"
            ),
        },
    ]


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
) -> list[dict[str, Any]]:
    normalized_actions: list[dict[str, Any]] = []
    for action in actions or []:
        if not isinstance(action, dict):
            continue
        title = normalize_text(action.get("title") or action.get("name"))
        rationale = normalize_text(action.get("rationale") or action.get("reason") or action.get("description"))
        if not title or not rationale:
            continue
        kind = str(action.get("kind") or "read_only_check").strip()
        if kind not in {"candidate_edit", "candidate_operation", "read_only_check"}:
            kind = "read_only_check"
        risk_level = str(action.get("risk_level") or "low").strip().lower()
        if risk_level not in {"low", "medium", "high"}:
            risk_level = "low"
        normalized_action = {
            "kind": kind,
            "title": title,
            "rationale": rationale,
            "risk_level": risk_level,
            "requires_confirmation": bool(action.get("requires_confirmation", False)),
            "citation_ids": normalize_citation_ids(
                action.get("citation_ids") or action.get("citations"),
                artifact_name_to_citation_id,
            ),
        }
        if isinstance(action.get("target"), dict):
            normalized_action["target"] = action["target"]
        if isinstance(action.get("patch"), dict):
            normalized_action["patch"] = action["patch"]
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
            fallback = next((item for item in fallback_citations if str(item.get("locator") or "") == f"artifact:{artifact_name}"), None)
            citation_id = str((fallback or {}).get("id") or f"doc-{index}")
        label = (
            normalize_text(citation.get("label"))
            or str(((artifact.get("metadata") or {}).get("page_slug") or "")).strip()
            or artifact_name
            or citation_id
        )
        normalized_citation = {
            "id": citation_id,
            "kind": str(citation.get("kind") or "artifact").strip() or "artifact",
            "label": label,
        }
        locator = normalize_text(citation.get("locator")) or (f"artifact:{artifact_name}" if artifact_name else "")
        if locator:
            normalized_citation["locator"] = locator
        excerpt = normalize_text(citation.get("excerpt") or citation.get("text"))
        if excerpt:
            normalized_citation["excerpt"] = excerpt
        normalized_citations.append(normalized_citation)
    return normalized_citations or fallback_citations


def coerce_response_document(document: dict[str, Any], copilot_request: dict[str, Any], raw_text: str) -> dict[str, Any]:
    coerced = dict(document)
    fallback_citations, artifact_name_to_citation_id = build_citation_maps(copilot_request)
    coerced["status"] = map_status(coerced.get("status"))
    coerced["answers"] = normalize_answers(coerced.get("answers"), artifact_name_to_citation_id)
    coerced["issues"] = normalize_issues(coerced.get("issues"), artifact_name_to_citation_id)
    coerced["proposed_actions"] = normalize_actions(
        coerced.get("proposed_actions"),
        artifact_name_to_citation_id,
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
    if not normalize_text(coerced.get("summary")) and coerced.get("answers"):
        first_answer = (coerced.get("answers") or [{}])[0]
        if isinstance(first_answer, dict):
            derived_summary = normalize_text(first_answer.get("text") or first_answer.get("summary"))
            if derived_summary:
                coerced["summary"] = derived_summary
    coerced.setdefault("summary", "模型已返回结构化响应，但摘要字段为空。")
    coerced.setdefault("answers", [])
    coerced.setdefault("issues", [])
    coerced.setdefault("proposed_actions", [])
    coerced.setdefault("citations", [])
    coerced.setdefault("confidence", 0.2)
    coerced.setdefault("risk_level", "medium")
    coerced.setdefault("requires_confirmation", False)
    coerced.setdefault("status", "partial")
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
        if isinstance(document, dict):
            return coerce_response_document(document, copilot_request, raw_text=content)
    return make_failed_response(
        copilot_request,
        "模型未返回可解析的 JSON，当前只保留原始文本供调试。",
        raw_text=content,
        code="MODEL_OUTPUT_NOT_JSON",
    )


def resolve_chat_endpoint(base_url: str) -> str:
    trimmed = base_url.rstrip("/")
    if trimmed.endswith("/chat/completions"):
        return trimmed
    if trimmed.endswith("/v1"):
        return trimmed + "/chat/completions"
    return trimmed + "/v1/chat/completions"


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


def call_openai_compatible(
    copilot_request: dict[str, Any],
    *,
    model: str,
    base_url: str,
    api_key: str,
    temperature: float,
) -> dict[str, Any]:
    messages = build_docs_qa_messages(copilot_request)
    endpoint = resolve_chat_endpoint(base_url)
    payload = {
        "model": model,
        "temperature": temperature,
        "messages": messages,
    }
    raw_request = {
        "endpoint": endpoint,
        "payload": payload,
    }
    data = json.dumps(payload).encode("utf-8")
    http_request = request.Request(
        endpoint,
        data=data,
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {api_key}",
        },
        method="POST",
    )
    try:
        with request.urlopen(http_request, timeout=120) as response_obj:
            raw_body = response_obj.read().decode("utf-8")
    except error.HTTPError as exc:
        error_body = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"provider request failed with HTTP {exc.code}: {error_body}") from exc
    except error.URLError as exc:
        raise RuntimeError(f"provider request failed: {exc}") from exc
    raw_response = json.loads(raw_body)
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


def run_inference(
    copilot_request: dict[str, Any],
    *,
    provider: str,
    model: str | None = None,
    base_url: str | None = None,
    api_key: str | None = None,
    temperature: float = 0.0,
) -> dict[str, Any]:
    load_env_file(ENV_FILE_PATH)
    validate_request_document(copilot_request)
    if str(copilot_request.get("project")) != "radish" or str(copilot_request.get("task")) != "answer_docs_question":
        raise ValueError("minimal runtime currently only supports radish / answer_docs_question")

    if provider == "mock":
        response_document = make_mock_docs_qa_response(copilot_request)
        validate_response_document(response_document)
        return {
            "provider": "mock",
            "model": model or "radishmind-mock-answer-docs-question-v1",
            "messages": build_docs_qa_messages(copilot_request),
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
        resolved_model = model or os.getenv("RADISHMIND_MODEL_NAME", "").strip()
        resolved_base_url = base_url or os.getenv("RADISHMIND_MODEL_BASE_URL", "").strip()
        resolved_api_key = api_key or os.getenv("RADISHMIND_MODEL_API_KEY", "").strip()
        if not resolved_model:
            raise ValueError("provider=openai-compatible requires --model or RADISHMIND_MODEL_NAME")
        if not resolved_base_url:
            raise ValueError("provider=openai-compatible requires --base-url or RADISHMIND_MODEL_BASE_URL")
        if not resolved_api_key:
            raise ValueError("provider=openai-compatible requires --api-key or RADISHMIND_MODEL_API_KEY")
        return call_openai_compatible(
            copilot_request,
            model=resolved_model,
            base_url=resolved_base_url,
            api_key=resolved_api_key,
            temperature=temperature,
        )

    raise ValueError(f"unsupported provider: {provider}")


def build_candidate_response_dump(
    *,
    copilot_request: dict[str, Any],
    sample_id: str,
    inference_result: dict[str, Any],
    dump_id: str | None = None,
) -> dict[str, Any]:
    request_id = str(copilot_request.get("request_id") or "").strip() or f"{copilot_request['task']}-request"
    return {
        "schema_version": 1,
        "dump_id": dump_id or f"dump-{request_id}",
        "project": copilot_request["project"],
        "task": copilot_request["task"],
        "sample_id": sample_id,
        "request_id": request_id,
        "captured_at": utc_now_iso(),
        "source": "simulated_candidate_response" if inference_result["provider"] == "mock" else "captured_candidate_response",
        "capture_metadata": {
            "capture_origin": "manual_fixture" if inference_result["provider"] == "mock" else "adapter_debug_dump",
            "tags": [
                "radish_docs_qa",
                inference_result["provider"],
            ],
        },
        "model": inference_result["model"],
        "input_record": derive_input_record(copilot_request),
        "input_request": copilot_request,
        "raw_request": inference_result.get("raw_request"),
        "raw_response": inference_result.get("raw_response"),
        "response": inference_result["response"],
    }
