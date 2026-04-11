from __future__ import annotations

from typing import Any


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


def _copy_if_present(target: dict[str, Any], source: dict[str, Any], source_key: str, target_key: str | None = None) -> None:
    if source_key not in source:
        return
    value = source.get(source_key)
    if value is None:
        return
    target[target_key or source_key] = _copy_json_value(value)


def _extract_primary_selected_unit(selection_state: dict[str, Any]) -> dict[str, Any] | None:
    primary_selected_unit = selection_state.get("primary_selected_unit")
    if isinstance(primary_selected_unit, dict) and primary_selected_unit:
        return _copy_json_value(primary_selected_unit)
    return None


def build_adapter_snapshot_from_export_snapshot(export_snapshot: dict[str, Any]) -> dict[str, Any]:
    task = str(export_snapshot.get("task") or "").strip()
    if task not in SUPPORTED_TASKS:
        raise ValueError(f"unsupported radishflow export task: {task}")

    document_state = export_snapshot.get("document_state") or {}
    selection_state = export_snapshot.get("selection_state") or {}
    diagnostics_export = export_snapshot.get("diagnostics_export") or {}
    adapter_snapshot: dict[str, Any] = {
        "schema_version": 1,
        "request_id": str(export_snapshot["request_id"]).strip(),
        "task": task,
        "locale": str(export_snapshot["locale"]).strip(),
    }

    _copy_if_present(adapter_snapshot, export_snapshot, "conversation_id")
    _copy_if_present(adapter_snapshot, document_state, "document_revision")
    _copy_if_present(adapter_snapshot, document_state, "flowsheet_document")
    _copy_if_present(adapter_snapshot, document_state, "flowsheet_document_uri")
    _copy_if_present(adapter_snapshot, selection_state, "selected_unit_ids")
    _copy_if_present(adapter_snapshot, selection_state, "selected_stream_ids")
    _copy_if_present(adapter_snapshot, diagnostics_export, "diagnostic_summary")
    _copy_if_present(adapter_snapshot, diagnostics_export, "diagnostics")
    _copy_if_present(adapter_snapshot, export_snapshot, "solve_session_state", "solve_session")
    _copy_if_present(adapter_snapshot, export_snapshot, "solve_snapshot", "latest_snapshot")
    _copy_if_present(adapter_snapshot, export_snapshot, "control_plane_snapshot", "control_plane_state")
    _copy_if_present(adapter_snapshot, export_snapshot, "support_artifacts", "supporting_artifacts")
    _copy_if_present(adapter_snapshot, export_snapshot, "tool_hints")
    _copy_if_present(adapter_snapshot, export_snapshot, "safety")

    selected_unit = _extract_primary_selected_unit(selection_state)
    if selected_unit:
        adapter_snapshot["selected_unit"] = selected_unit
    return adapter_snapshot
