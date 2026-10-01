from __future__ import annotations

import json
from typing import Any


PROMPT_APPLICATION_PROTOCOL = "prompt-application-invocation-v1"
PROMPT_APPLICATION_ROUTE = "/v1/prompt-applications/invocations"


def is_prompt_application_request(document: dict[str, Any]) -> bool:
    context = document.get("context") or {}
    if not isinstance(context, dict):
        return False
    northbound = context.get("northbound") or {}
    if not isinstance(northbound, dict):
        return context.get("route") == PROMPT_APPLICATION_ROUTE
    return (
        context.get("route") == PROMPT_APPLICATION_ROUTE
        or northbound.get("protocol") == PROMPT_APPLICATION_PROTOCOL
        or northbound.get("request_kind") == PROMPT_APPLICATION_PROTOCOL
    )


def _unique_message_fields(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("duplicate Prompt application message field")
        result[key] = value
    return result


def build_prompt_application_messages(document: dict[str, Any]) -> list[dict[str, str]]:
    """Decode the Go-rendered packet; never render variables or infer authority here."""
    context = document.get("context") or {}
    northbound = context.get("northbound") or {}
    hints = document.get("tool_hints") or {}
    if not isinstance(northbound, dict) or not isinstance(hints, dict):
        raise ValueError("invalid Prompt application request boundary")
    if (
        document.get("project") != "radish"
        or document.get("task") != "answer_docs_question"
        or context.get("current_app") != "radishmind-platform"
        or context.get("route") != PROMPT_APPLICATION_ROUTE
        or northbound.get("protocol") != PROMPT_APPLICATION_PROTOCOL
        or northbound.get("request_kind") != PROMPT_APPLICATION_PROTOCOL
        or any(northbound.get(key) is not False for key in (
            "allow_tool_calls", "allow_retrieval", "writes_business_truth",
        ))
        or any(hints.get(key) is not False for key in (
            "allow_tool_calls", "allow_retrieval", "allow_image_reasoning",
        ))
    ):
        raise ValueError("invalid Prompt application request boundary")
    artifacts = document.get("artifacts")
    if not isinstance(artifacts, list) or len(artifacts) != 1:
        raise ValueError("invalid Prompt application message packet")
    artifact = artifacts[0]
    if not isinstance(artifact, dict) or any(artifact.get(key) != expected for key, expected in (
        ("name", "northbound_prompt"), ("kind", "text"),
        ("role", "primary"), ("mime_type", "text/plain"),
    )):
        raise ValueError("invalid Prompt application message artifact")
    content = artifact.get("content")
    if not isinstance(content, str) or len(content.encode("utf-8")) > 128 * 1024:
        raise ValueError("invalid Prompt application message budget")
    try:
        messages = json.loads(content, object_pairs_hook=_unique_message_fields)
    except (ValueError, RecursionError) as exc:
        raise ValueError("invalid Prompt application message JSON") from exc
    if not isinstance(messages, list) or not 1 <= len(messages) <= 16:
        raise ValueError("invalid Prompt application message count")
    for message in messages:
        if (
            not isinstance(message, dict)
            or set(message) != {"role", "content"}
            or message["role"] not in ("system", "developer", "user")
            or not isinstance(message["content"], str)
        ):
            raise ValueError("invalid Prompt application message shape")
        message["content"].encode("utf-8")
    return messages


def build_prompt_application_response(content: str, document: dict[str, Any]) -> dict[str, Any]:
    """Carry opaque output in the existing envelope; Go owns output validation."""
    build_prompt_application_messages(document)
    if not content.strip() or len(content.encode("utf-8")) > 64 * 1024:
        raise ValueError("invalid Prompt application output budget")
    return {
        "schema_version": 1,
        "status": "ok",
        "project": "radish",
        "task": "answer_docs_question",
        "summary": content,
        "answers": [],
        "issues": [],
        "proposed_actions": [],
        "citations": [],
        "confidence": 0,
        "risk_level": "medium",
        "requires_confirmation": False,
    }
