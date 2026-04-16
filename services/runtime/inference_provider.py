from __future__ import annotations

import http.client
import json
import re
from typing import Any
from urllib import error, request

from .inference_response import coerce_response_document
from .inference_support import (
    ENV_FILE_PATH,
    GHOST_MALFORMED_JSON_REPAIR_PATTERNS,
    GHOST_MANUAL_MULTI_ACTION_REPAIR_PATTERNS,
    build_messages,
    derive_input_record,
    infer_profile_api_style,
    load_env_file,
    make_failed_response,
    make_mock_docs_qa_response,
    make_mock_suggest_edits_response,
    make_mock_ghost_completion_response,
    normalize_text,
    profile_env_key,
    resolve_openai_compatible_config,
    utc_now_iso,
    validate_request_document,
    validate_response_document,
)


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
            elif (
                str(copilot_request.get("project") or "").strip() == "radishflow"
                and str(copilot_request.get("task") or "").strip() == "suggest_flowsheet_edits"
            ):
                repaired_candidate = repair_malformed_suggest_edits_json(candidate)
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


def repair_malformed_suggest_edits_json(candidate: str) -> str:
    repaired = candidate
    repair_patterns = (
        (
            re.compile(r"(\}\}\})\},(\"risk_level\"\s*:)", flags=re.DOTALL),
            r"\1,\2",
        ),
        (
            re.compile(r'语义为"([^"]+)"', flags=re.DOTALL),
            r'语义为\\"\1\\"',
        ),
        (
            re.compile(r'所提到的"([^"]+)"', flags=re.DOTALL),
            r'所提到的\\"\1\\"',
        ),
    )
    for _ in range(2):
        previous = repaired
        for pattern, replacement in repair_patterns:
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


def resolve_anthropic_messages_endpoint(base_url: str) -> str:
    trimmed = base_url.rstrip("/")
    if trimmed.endswith("/messages"):
        return trimmed
    if trimmed.endswith("/v1"):
        return trimmed + "/messages"
    return trimmed + "/v1/messages"


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


def extract_anthropic_message_content(payload: dict[str, Any]) -> str:
    content = payload.get("content")
    if not isinstance(content, list):
        return ""
    text_parts: list[str] = []
    for item in content:
        if not isinstance(item, dict):
            continue
        if item.get("type") == "text" and isinstance(item.get("text"), str):
            text_parts.append(item["text"])
    return "\n".join(text_parts).strip()


def extract_provider_message_content(payload: dict[str, Any]) -> str:
    content = extract_openai_message_content(payload)
    if content.strip():
        return content
    content = extract_gemini_message_content(payload)
    if content.strip():
        return content
    return extract_anthropic_message_content(payload)


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


def build_anthropic_payload(messages: list[dict[str, str]], model: str, temperature: float) -> dict[str, Any]:
    system_texts: list[str] = []
    anthropic_messages: list[dict[str, Any]] = []
    for message in messages:
        role = str(message.get("role") or "").strip()
        content = str(message.get("content") or "").strip()
        if not content:
            continue
        if role == "system":
            system_texts.append(content)
            continue
        anthropic_role = "assistant" if role == "assistant" else "user"
        anthropic_messages.append(
            {
                "role": anthropic_role,
                "content": content,
            }
        )
    payload: dict[str, Any] = {
        "model": model,
        "max_tokens": 4096,
        "temperature": temperature,
        "messages": anthropic_messages,
    }
    if system_texts:
        payload["system"] = "\n\n".join(system_texts)
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


def call_anthropic_messages(
    copilot_request: dict[str, Any],
    *,
    model: str,
    base_url: str,
    api_key: str,
    temperature: float,
    request_timeout_seconds: float,
) -> dict[str, Any]:
    messages = build_messages(copilot_request)
    endpoint = resolve_anthropic_messages_endpoint(base_url)
    payload = build_anthropic_payload(messages, model, temperature)
    raw_request = {
        "endpoint": endpoint,
        "payload": payload,
        "transport": "anthropic-messages",
    }
    raw_response = post_json_request(
        endpoint=endpoint,
        payload=payload,
        headers={
            "Content-Type": "application/json",
            "x-api-key": api_key,
            "anthropic-version": "2023-06-01",
        },
        request_timeout_seconds=request_timeout_seconds,
    )
    content = extract_anthropic_message_content(raw_response)
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
        elif project == "radishflow" and task == "suggest_flowsheet_edits":
            response_document = make_mock_suggest_edits_response(copilot_request)
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
        if resolved_api_style not in {"openai-compatible", "gemini-native", "anthropic-messages"}:
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
        elif resolved_api_style == "anthropic-messages":
            result = call_anthropic_messages(
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
