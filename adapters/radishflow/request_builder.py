from __future__ import annotations

from typing import Any


DEFAULT_TOOL_HINTS = {
    "allow_retrieval": False,
    "allow_tool_calls": False,
    "allow_image_reasoning": False,
}

DEFAULT_SAFETY = {
    "mode": "advisory",
    "requires_confirmation_for_actions": True,
}

SUPPORTED_TASKS = {
    "explain_diagnostics",
    "suggest_flowsheet_edits",
    "explain_control_plane_state",
}


def _copy_json_value(value: Any) -> Any:
    if isinstance(value, dict):
        return {key: _copy_json_value(item) for key, item in value.items()}
    if isinstance(value, list):
        return [_copy_json_value(item) for item in value]
    return value


def _copy_non_empty_dict_fields(snapshot: dict[str, Any], keys: tuple[str, ...]) -> dict[str, Any]:
    context: dict[str, Any] = {}
    for key in keys:
        value = snapshot.get(key)
        if value is None:
            continue
        if isinstance(value, list) and not value:
            continue
        if isinstance(value, dict) and not value:
            continue
        context[key] = _copy_json_value(value)
    return context


def build_request_from_snapshot(snapshot: dict[str, Any]) -> dict[str, Any]:
    task = str(snapshot.get("task") or "").strip()
    if task not in SUPPORTED_TASKS:
        raise ValueError(f"unsupported radishflow adapter task: {task}")

    request: dict[str, Any] = {
        "schema_version": 1,
        "request_id": str(snapshot["request_id"]).strip(),
        "project": "radishflow",
        "task": task,
        "locale": str(snapshot["locale"]).strip(),
        "artifacts": [],
        "context": {},
        "tool_hints": _copy_json_value(snapshot.get("tool_hints") or DEFAULT_TOOL_HINTS),
        "safety": _copy_json_value(snapshot.get("safety") or DEFAULT_SAFETY),
    }

    conversation_id = str(snapshot.get("conversation_id") or "").strip()
    if conversation_id:
        request["conversation_id"] = conversation_id

    flowsheet_document = snapshot.get("flowsheet_document")
    flowsheet_document_uri = str(snapshot.get("flowsheet_document_uri") or "").strip()
    if flowsheet_document is not None or flowsheet_document_uri:
        primary_artifact: dict[str, Any] = {
            "kind": "json",
            "role": "primary",
            "name": "flowsheet_document",
            "mime_type": "application/json",
        }
        if flowsheet_document_uri:
            primary_artifact["uri"] = flowsheet_document_uri
        else:
            primary_artifact["content"] = _copy_json_value(flowsheet_document)
        request["artifacts"].append(primary_artifact)

    for artifact in snapshot.get("supporting_artifacts") or []:
        request["artifacts"].append(_copy_json_value(artifact))

    request["context"] = _copy_non_empty_dict_fields(
        snapshot,
        (
            "document_revision",
            "selected_unit_ids",
            "selected_unit",
            "selected_stream_ids",
            "diagnostic_summary",
            "diagnostics",
            "solve_session",
            "latest_snapshot",
            "control_plane_state",
        ),
    )
    return request
