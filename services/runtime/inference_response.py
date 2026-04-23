from __future__ import annotations

import ast
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
    build_ghost_completion_action,
    extract_embedded_summary_text,
    infer_parameter_placeholders,
    infer_stream_spec_placeholders,
    load_schema,
    make_failed_response,
    normalize_text,
    pick_primary_ghost_candidate,
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
        kind = str(source_action.get("kind") or source_action.get("action") or "read_only_check").strip()
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


def normalize_spec_placeholders(values: Any) -> list[str]:
    alias_map = {
        "temperature": "temperature_c",
        "temperature_k": "temperature_c",
        "temperature_deg_c": "temperature_c",
        "temp": "temperature_c",
        "pressure": "pressure_kpa",
        "pressure_bar": "pressure_kpa",
        "pressure_pa": "pressure_kpa",
        "flow_rate": "flow_rate_kg_per_h",
        "flow_rate_kgh": "flow_rate_kg_per_h",
        "flow_rate_kgph": "flow_rate_kg_per_h",
        "flow_rate_kg_h": "flow_rate_kg_per_h",
        "mass_flow": "flow_rate_kg_per_h",
        "mass_flow_rate": "flow_rate_kg_per_h",
        "mass_flow_rate_kg_h": "flow_rate_kg_per_h",
        "mass_flow_rate_kg_per_h": "flow_rate_kg_per_h",
        "mass_flow_kgph": "flow_rate_kg_per_h",
        "mass_flow_kg_h": "flow_rate_kg_per_h",
        "mass_flow_kg_per_h": "flow_rate_kg_per_h",
    }
    canonical_order = {
        "temperature_c": 0,
        "pressure_kpa": 1,
        "flow_rate_kg_per_h": 2,
        "composition": 3,
        "vapor_fraction": 4,
    }
    if values is None:
        return []
    if isinstance(values, dict):
        add_specs = values.get("add_specs")
        if isinstance(add_specs, dict):
            values = list(add_specs.keys())
        else:
            values = list(values.keys())
    elif not isinstance(values, list):
        values = [values]

    normalized: list[str] = []
    for value in values:
        if isinstance(value, dict):
            placeholder = str(value.get("path") or value.get("name") or "").strip()
        else:
            placeholder = str(value).strip()
        path_match = re.search(r"specs\.([A-Za-z0-9_]+)", placeholder)
        if path_match:
            placeholder = path_match.group(1)
        if not placeholder:
            continue
        placeholder = alias_map.get(placeholder, placeholder)
        if placeholder not in normalized:
            normalized.append(placeholder)
    return sorted(
        normalized,
        key=lambda placeholder: (canonical_order.get(placeholder, len(canonical_order)), normalized.index(placeholder)),
    )


def normalize_parameter_placeholders(values: list[Any]) -> list[str]:
    alias_map = {
        "outlet_temperature_target": "outlet_temperature_c",
        "outlet_temperature_target_c": "outlet_temperature_c",
        "target_outlet_temperature_c": "outlet_temperature_c",
        "outlet_temp_target_c": "outlet_temperature_c",
        "temperature_target_c": "outlet_temperature_c",
    }
    normalized: list[str] = []
    for value in values:
        placeholder = ""
        if isinstance(value, dict):
            raw_value = str(value.get("path") or value.get("name") or value.get("key") or "").strip()
            placeholder = raw_value
            path_match = re.search(r"(?:config|parameters|parameter_updates)\.([A-Za-z0-9_]+)", raw_value)
            if path_match:
                placeholder = path_match.group(1)
            elif "." in raw_value:
                trailing_key = raw_value.split(".")[-1].strip()
                if re.fullmatch(r"[A-Za-z0-9_]+", trailing_key):
                    placeholder = trailing_key
        else:
            raw_value = str(value).strip()
            if not raw_value:
                continue
            placeholder = raw_value
            if raw_value[:1] in {"{", "["}:
                parsed_value: Any = None
                try:
                    parsed_value = json.loads(raw_value)
                except json.JSONDecodeError:
                    try:
                        parsed_value = ast.literal_eval(raw_value)
                    except (SyntaxError, ValueError):
                        parsed_value = None
                if isinstance(parsed_value, dict):
                    placeholder = str(
                        parsed_value.get("path")
                        or parsed_value.get("name")
                        or parsed_value.get("key")
                        or ""
                    ).strip()
                elif isinstance(parsed_value, list) and parsed_value:
                    first_item = parsed_value[0]
                    if isinstance(first_item, str):
                        placeholder = first_item.strip()
                    elif isinstance(first_item, dict):
                        placeholder = str(
                            first_item.get("path")
                            or first_item.get("name")
                            or first_item.get("key")
                            or ""
                        ).strip()
            path_match = re.search(r"(?:config|parameters|parameter_updates)\.([A-Za-z0-9_]+)", raw_value)
            if path_match:
                placeholder = path_match.group(1)
            elif "." in raw_value:
                trailing_key = raw_value.split(".")[-1].strip()
                if re.fullmatch(r"[A-Za-z0-9_]+", trailing_key):
                    placeholder = trailing_key
        if not placeholder:
            continue
        placeholder = alias_map.get(placeholder, placeholder)
        if placeholder not in normalized:
            normalized.append(placeholder)
    return normalized


def build_suggest_edits_action_target_order(
    diagnostics: list[dict[str, Any]],
    existing_actions_by_target: dict[tuple[str, str], dict[str, Any]],
    auto_synthesize_codes: set[str],
) -> list[tuple[str, str]]:
    ordered_targets: list[tuple[str, str]] = []
    diagnostic_targets: set[tuple[str, str]] = set()
    for diagnostic in diagnostics:
        diagnostic_code = str(diagnostic.get("code") or "").strip()
        target_type = str(diagnostic.get("target_type") or "").strip().lower()
        target_id = str(diagnostic.get("target_id") or "").strip()
        if target_type not in {"stream", "unit"} or not target_id:
            continue
        target_key = (target_type, target_id)
        diagnostic_targets.add(target_key)
        if target_key in ordered_targets:
            continue
        severity = str(diagnostic.get("severity") or "").strip().lower()
        has_existing_action = target_key in existing_actions_by_target
        if diagnostic_code in auto_synthesize_codes or has_existing_action:
            ordered_targets.append(target_key)
    for target_key in existing_actions_by_target:
        if target_key not in ordered_targets and target_key not in diagnostic_targets:
            ordered_targets.append(target_key)
    return ordered_targets


def should_keep_suggest_edits_existing_action_target(
    *,
    target_type: str,
    target_id: str,
    diagnostics: list[dict[str, Any]],
    copilot_request: dict[str, Any],
) -> bool:
    if target_type not in {"stream", "unit"} or not target_id:
        return False

    diagnostic_targets = {
        (
            str((diagnostic or {}).get("target_type") or "").strip().lower(),
            str((diagnostic or {}).get("target_id") or "").strip(),
        )
        for diagnostic in diagnostics
        if isinstance(diagnostic, dict)
        and str((diagnostic or {}).get("target_type") or "").strip().lower() in {"stream", "unit"}
        and str((diagnostic or {}).get("target_id") or "").strip()
    }
    if (target_type, target_id) in diagnostic_targets:
        target_diagnostics = [
            diagnostic
            for diagnostic in diagnostics
            if str((diagnostic or {}).get("target_type") or "").strip().lower() == target_type
            and str((diagnostic or {}).get("target_id") or "").strip() == target_id
        ]
        dependent_warning_codes = {
            "SEPARATOR_STATE_DEPENDENT",
            "HEATER_OUTLET_EFFECT_UNCONFIRMED",
            "COOLER_OUTLET_EFFECT_UNCONFIRMED",
        }
        if target_type == "unit" and target_diagnostics:
            target_codes = {
                str((diagnostic or {}).get("code") or "").strip()
                for diagnostic in target_diagnostics
            }
            only_warning_diagnostics = all(
                str((diagnostic or {}).get("severity") or "").strip().lower() == "warning"
                for diagnostic in target_diagnostics
            )
            has_error_stream_diagnostic = any(
                str((diagnostic or {}).get("target_type") or "").strip().lower() == "stream"
                and str((diagnostic or {}).get("severity") or "").strip().lower() == "error"
                for diagnostic in diagnostics
            )
            if (
                only_warning_diagnostics
                and target_codes
                and target_codes.issubset(dependent_warning_codes)
                and has_error_stream_diagnostic
            ):
                return False
        return True

    selected_unit_ids, selected_stream_ids, _, _ = build_selected_connection_maps(copilot_request)
    if len(selected_unit_ids) != 1 or selected_stream_ids:
        return True

    selected_unit_id = next(iter(selected_unit_ids))
    unit_only_diagnostics = [
        diagnostic
        for diagnostic in diagnostics
        if str((diagnostic or {}).get("target_type") or "").strip().lower() == "unit"
        and str((diagnostic or {}).get("target_id") or "").strip() == selected_unit_id
    ]
    if len(unit_only_diagnostics) != len(diagnostics):
        return True

    return target_type == "unit" and target_id == selected_unit_id


def build_suggest_edits_group_citation_ids(
    *,
    copilot_request: dict[str, Any],
    diagnostics: list[dict[str, Any]],
    diagnostic_indexes: list[int],
    target_type: str,
    target_id: str,
    target_citation_lookup: dict[tuple[str, str], str],
    snapshot_citation_id: str,
) -> list[str]:
    citation_ids = [f"diag-{index}" for index in diagnostic_indexes]
    primary_diagnostic = diagnostics[0] if diagnostics else {}
    citation_ids.extend(
        build_suggest_edits_context_support_citation_ids(
            copilot_request=copilot_request,
            diagnostic_code=str((primary_diagnostic or {}).get("code") or "").strip(),
            severity=str((primary_diagnostic or {}).get("severity") or "").strip().lower(),
            target_type=target_type,
            target_id=target_id,
            item_kind="action",
            target_citation_lookup=target_citation_lookup,
            snapshot_citation_id=snapshot_citation_id,
        )
    )
    if any(
        str((diagnostic or {}).get("code") or "").strip() == "COMPRESSOR_ROOT_CAUSE_UNCONFIRMED"
        for diagnostic in diagnostics
    ) and snapshot_citation_id:
        append_unique_citation_id(citation_ids, snapshot_citation_id)
    return list(dict.fromkeys(citation_ids))


def normalize_parameter_updates(value: Any) -> dict[str, Any]:
    if not isinstance(value, dict):
        return {}
    normalized_updates: dict[str, Any] = {}
    for key, detail in value.items():
        normalized_key = str(key).strip()
        if not normalized_key or not isinstance(detail, dict):
            continue
        normalized_updates[normalized_key] = detail
    return normalized_updates


def load_flowsheet_document(copilot_request: dict[str, Any]) -> dict[str, Any]:
    for artifact in copilot_request.get("artifacts") or []:
        if not isinstance(artifact, dict):
            continue
        if str(artifact.get("name") or "").strip() != "flowsheet_document":
            continue
        content = artifact.get("content")
        if isinstance(content, dict):
            return content
    return {}


def build_flowsheet_lookup(copilot_request: dict[str, Any]) -> tuple[dict[str, dict[str, Any]], dict[str, dict[str, Any]], dict[str, str]]:
    flowsheet_document = load_flowsheet_document(copilot_request)
    unit_by_id: dict[str, dict[str, Any]] = {}
    stream_by_id: dict[str, dict[str, Any]] = {}
    primary_inlet_stream_by_unit_id: dict[str, str] = {}
    for unit in flowsheet_document.get("units") or []:
        if not isinstance(unit, dict):
            continue
        unit_id = str(unit.get("id") or "").strip()
        if unit_id:
            unit_by_id[unit_id] = unit
    for stream in flowsheet_document.get("streams") or []:
        if not isinstance(stream, dict):
            continue
        stream_id = str(stream.get("id") or "").strip()
        target_unit_id = str(stream.get("target_unit_id") or "").strip()
        if stream_id:
            stream_by_id[stream_id] = stream
        if target_unit_id and target_unit_id not in primary_inlet_stream_by_unit_id:
            primary_inlet_stream_by_unit_id[target_unit_id] = stream_id
    return unit_by_id, stream_by_id, primary_inlet_stream_by_unit_id


def append_unique_citation_id(citation_ids: list[str], citation_id: str) -> None:
    normalized = str(citation_id or "").strip()
    if normalized and normalized not in citation_ids:
        citation_ids.append(normalized)


def build_selected_connection_maps(copilot_request: dict[str, Any]) -> tuple[set[str], list[str], dict[str, list[str]], dict[str, list[str]]]:
    context = copilot_request.get("context") or {}
    selected_unit_ids = {
        str(unit_id).strip()
        for unit_id in (context.get("selected_unit_ids") or [])
        if str(unit_id).strip()
    }
    selected_stream_ids = [
        str(stream_id).strip()
        for stream_id in (context.get("selected_stream_ids") or [])
        if str(stream_id).strip()
    ]
    _, stream_by_id, _ = build_flowsheet_lookup(copilot_request)
    input_stream_ids_by_unit_id: dict[str, list[str]] = {}
    output_stream_ids_by_unit_id: dict[str, list[str]] = {}
    for stream_id in selected_stream_ids:
        stream_document = stream_by_id.get(stream_id) or {}
        source_unit_id = str(stream_document.get("source_unit_id") or "").strip()
        target_unit_id = str(stream_document.get("target_unit_id") or "").strip()
        if target_unit_id:
            input_stream_ids_by_unit_id.setdefault(target_unit_id, []).append(stream_id)
        if source_unit_id:
            output_stream_ids_by_unit_id.setdefault(source_unit_id, []).append(stream_id)
    return selected_unit_ids, selected_stream_ids, input_stream_ids_by_unit_id, output_stream_ids_by_unit_id


def build_suggest_edits_context_support_citation_ids(
    *,
    copilot_request: dict[str, Any],
    diagnostic_code: str,
    severity: str,
    target_type: str,
    target_id: str,
    item_kind: str,
    target_citation_lookup: dict[tuple[str, str], str],
    snapshot_citation_id: str,
) -> list[str]:
    selected_unit_ids, selected_stream_ids, input_stream_ids_by_unit_id, output_stream_ids_by_unit_id = build_selected_connection_maps(
        copilot_request
    )
    unit_by_id, stream_by_id, _ = build_flowsheet_lookup(copilot_request)
    complex_cross_object_context = len(selected_unit_ids) > 1
    diagnostics = [
        diagnostic
        for diagnostic in ((copilot_request.get("context") or {}).get("diagnostics") or [])
        if isinstance(diagnostic, dict)
    ]

    def source_unit_has_primary_stream_supported_action(source_unit_id: str) -> bool:
        normalized_source_unit_id = str(source_unit_id).strip()
        if not normalized_source_unit_id:
            return False
        primary_source_unit_diagnostic = next(
            (
                diagnostic
                for diagnostic in diagnostics
                if str((diagnostic or {}).get("target_type") or "").strip().lower() == "unit"
                and str((diagnostic or {}).get("target_id") or "").strip() == normalized_source_unit_id
            ),
            None,
        )
        if not isinstance(primary_source_unit_diagnostic, dict):
            return False
        primary_code = str((primary_source_unit_diagnostic or {}).get("code") or "").strip()
        if primary_code in {
            "COMPRESSOR_OUTLET_PRESSURE_TARGET_INVALID",
            "COMPRESSOR_OUTLET_PRESSURE_TARGET_TOO_CLOSE",
        }:
            return bool(input_stream_ids_by_unit_id.get(normalized_source_unit_id, []))
        if primary_code == "PUMP_OUTLET_PRESSURE_TARGET_INVALID":
            return any(
                any(
                    str((diagnostic or {}).get("target_type") or "").strip().lower() == "stream"
                    and str((diagnostic or {}).get("target_id") or "").strip() == stream_id
                    for diagnostic in diagnostics
                )
                for stream_id in input_stream_ids_by_unit_id.get(normalized_source_unit_id, [])
            )
        return False

    citation_ids: list[str] = []
    target_citation_id = target_citation_lookup.get((target_type, target_id))
    skip_target_citation = (
        (
            diagnostic_code == "COMPRESSOR_ROOT_CAUSE_UNCONFIRMED"
            or (
                target_type == "stream"
                and diagnostic_code in {"HEATER_OUTLET_EFFECT_UNCONFIRMED", "COOLER_OUTLET_EFFECT_UNCONFIRMED"}
                and not complex_cross_object_context
                and target_id not in selected_stream_ids
            )
        )
        and item_kind == "issue"
    )
    if target_citation_id and not skip_target_citation:
        append_unique_citation_id(citation_ids, target_citation_id)

    if target_type == "stream":
        stream_document = stream_by_id.get(target_id) or {}
        source_unit_id = str(stream_document.get("source_unit_id") or "").strip()
        target_unit_id = str(stream_document.get("target_unit_id") or "").strip()
        primary_unit_id = ""
        if diagnostic_code == "STREAM_SPEC_MISSING":
            primary_unit_id = target_unit_id or source_unit_id
        else:
            primary_unit_id = source_unit_id or target_unit_id
        direct_unit_ids = [primary_unit_id] if primary_unit_id else []
        if complex_cross_object_context:
            for unit_id in direct_unit_ids:
                if unit_id and unit_id in selected_unit_ids:
                    append_unique_citation_id(citation_ids, target_citation_lookup.get(("unit", unit_id), ""))

        if complex_cross_object_context and (
            diagnostic_code == "DOWNSTREAM_STATE_DEPENDENT"
            or (diagnostic_code == "STREAM_DISCONNECTED" and item_kind == "action")
        ):
            for unit_id in direct_unit_ids:
                if not unit_id:
                    continue
                connected_stream_ids = [
                    *output_stream_ids_by_unit_id.get(unit_id, []),
                    *input_stream_ids_by_unit_id.get(unit_id, []),
                ]
                for stream_id in connected_stream_ids:
                    if stream_id == target_id:
                        continue
                    peer_stream_document = stream_by_id.get(stream_id) or {}
                    if diagnostic_code == "STREAM_DISCONNECTED" and item_kind == "action":
                        append_unique_citation_id(citation_ids, target_citation_lookup.get(("stream", stream_id), ""))
                    peer_source_unit_id = str(peer_stream_document.get("source_unit_id") or "").strip()
                    peer_target_unit_id = str(peer_stream_document.get("target_unit_id") or "").strip()
                    opposite_unit_id = ""
                    if peer_source_unit_id and peer_source_unit_id != unit_id:
                        opposite_unit_id = peer_source_unit_id
                    elif peer_target_unit_id and peer_target_unit_id != unit_id:
                        opposite_unit_id = peer_target_unit_id
                    if opposite_unit_id and opposite_unit_id in selected_unit_ids:
                        append_unique_citation_id(
                            citation_ids,
                            target_citation_lookup.get(("unit", opposite_unit_id), ""),
                        )

    elif target_type == "unit":
        output_stream_ids = list(output_stream_ids_by_unit_id.get(target_id, []))
        input_stream_ids = list(input_stream_ids_by_unit_id.get(target_id, []))
        if diagnostic_code == "UNIT_PARAMETER_INCOMPLETE":
            for stream_id in output_stream_ids:
                append_unique_citation_id(citation_ids, target_citation_lookup.get(("stream", stream_id), ""))
            if item_kind == "action":
                for stream_id in input_stream_ids:
                    input_stream_document = stream_by_id.get(stream_id) or {}
                    source_unit_id = str(input_stream_document.get("source_unit_id") or "").strip()
                    if (
                        source_unit_id
                        and source_unit_id in selected_unit_ids
                        and not source_unit_has_primary_stream_supported_action(source_unit_id)
                    ):
                        append_unique_citation_id(citation_ids, target_citation_lookup.get(("unit", source_unit_id), ""))
            if snapshot_citation_id:
                append_unique_citation_id(citation_ids, snapshot_citation_id)
        elif diagnostic_code == "DOWNSTREAM_SEPARATOR_STATE_DEPENDENT":
            supporting_stream_ids = input_stream_ids or output_stream_ids
            for stream_id in supporting_stream_ids:
                append_unique_citation_id(citation_ids, target_citation_lookup.get(("stream", stream_id), ""))
            if severity != "error" and snapshot_citation_id:
                append_unique_citation_id(citation_ids, snapshot_citation_id)
        elif diagnostic_code.startswith("DOWNSTREAM_") or diagnostic_code.endswith("_STATE_DEPENDENT"):
            supporting_stream_ids = output_stream_ids or input_stream_ids
            for stream_id in supporting_stream_ids:
                append_unique_citation_id(citation_ids, target_citation_lookup.get(("stream", stream_id), ""))
            if severity != "error" and snapshot_citation_id:
                append_unique_citation_id(citation_ids, snapshot_citation_id)
        elif diagnostic_code in {
            "COMPRESSOR_OUTLET_PRESSURE_TARGET_INVALID",
            "COMPRESSOR_OUTLET_PRESSURE_TARGET_TOO_CLOSE",
        }:
            for stream_id in input_stream_ids:
                append_unique_citation_id(citation_ids, target_citation_lookup.get(("stream", stream_id), ""))
        elif diagnostic_code == "PUMP_OUTLET_PRESSURE_TARGET_INVALID":
            for stream_id in input_stream_ids:
                has_direct_stream_diagnostic = any(
                    str((diagnostic or {}).get("target_type") or "").strip().lower() == "stream"
                    and str((diagnostic or {}).get("target_id") or "").strip() == stream_id
                    for diagnostic in diagnostics
                )
                if not has_direct_stream_diagnostic:
                    continue
                append_unique_citation_id(citation_ids, target_citation_lookup.get(("stream", stream_id), ""))
        elif severity != "error" and snapshot_citation_id:
            append_unique_citation_id(citation_ids, snapshot_citation_id)

    if target_type == "stream" and severity != "error" and snapshot_citation_id:
        append_unique_citation_id(citation_ids, snapshot_citation_id)
    if diagnostic_code == "STREAM_DISCONNECTED" and snapshot_citation_id:
        append_unique_citation_id(citation_ids, snapshot_citation_id)
    if target_type == "unit" and target_id not in unit_by_id and snapshot_citation_id:
        append_unique_citation_id(citation_ids, snapshot_citation_id)
    return citation_ids


def synthesize_parameter_updates_from_diagnostics(
    *,
    target_type: str,
    target_id: str,
    diagnostics: list[dict[str, Any]],
    copilot_request: dict[str, Any],
) -> dict[str, Any]:
    if target_type != "unit" or not target_id:
        return {}

    unit_by_id, _, primary_inlet_stream_by_unit_id = build_flowsheet_lookup(copilot_request)
    unit_document = unit_by_id.get(target_id) or {}
    unit_kind = str(unit_document.get("kind") or "").strip().lower()
    config = unit_document.get("config") if isinstance(unit_document.get("config"), dict) else {}
    diagnostic_codes = {str((diagnostic or {}).get("code") or "").strip() for diagnostic in diagnostics}
    synthesized_patch: dict[str, Any] = {}
    parameter_updates: dict[str, Any] = {}
    parameter_placeholders: list[str] = []
    primary_inlet_stream_id = primary_inlet_stream_by_unit_id.get(target_id, "")
    efficiency_percent = config.get("efficiency_percent")

    def is_efficiency_out_of_review_range(value: Any) -> bool:
        if not isinstance(value, (int, float)):
            return False
        numeric = float(value)
        return numeric < 60 or numeric > 85

    def suggested_efficiency_review_range() -> list[int]:
        if "PUMP_OUTLET_PRESSURE_TARGET_INVALID" in diagnostic_codes:
            return [60, 85]
        if "COMPRESSOR_OUTLET_PRESSURE_TARGET_INVALID" in diagnostic_codes:
            return [60, 85]
        if unit_kind == "pump":
            return [65, 82]
        if unit_kind == "compressor":
            return [65, 85]
        return [60, 85]

    for diagnostic in diagnostics:
        diagnostic_code = str((diagnostic or {}).get("code") or "").strip()
        message = normalize_text((diagnostic or {}).get("message") or (diagnostic or {}).get("text") or "")
        if diagnostic_code == "COMPRESSOR_OUTLET_PRESSURE_TARGET_INVALID":
            pressure_update: dict[str, Any] = {
                "action": "review_and_raise_above_inlet",
            }
            if primary_inlet_stream_id:
                pressure_update["minimum_reference_stream_id"] = primary_inlet_stream_id
            parameter_updates["outlet_pressure_target_kpa"] = pressure_update
            continue
        if diagnostic_code == "COMPRESSOR_OUTLET_PRESSURE_TARGET_MISSING":
            parameter_placeholders.append("outlet_pressure_target_kpa")
            continue
        if diagnostic_code == "COMPRESSOR_MINIMUM_FLOW_BYPASS_MISSING":
            parameter_placeholders.append("minimum_flow_bypass_percent")
            continue
        if diagnostic_code == "COMPRESSOR_EFFICIENCY_BASELINE_MISSING":
            parameter_placeholders.append("efficiency_percent")
            continue
        if diagnostic_code == "UNIT_PARAMETER_INCOMPLETE":
            inferred_placeholders = infer_parameter_placeholders(message)
            if unit_kind == "pump" and "outlet_pressure_target_kpa" in inferred_placeholders:
                parameter_placeholders.append("outlet_pressure_target_kpa")
                continue
            parameter_placeholders.extend(inferred_placeholders)

    if "PUMP_OUTLET_PRESSURE_TARGET_INVALID" in diagnostic_codes:
        pressure_update: dict[str, Any] = {
            "action": "review_and_raise_above_inlet",
        }
        if primary_inlet_stream_id:
            pressure_update["minimum_reference_stream_id"] = primary_inlet_stream_id
        parameter_updates["outlet_pressure_target_kpa"] = pressure_update

        if (
            "UNIT_PARAMETER_OUT_OF_RANGE" in diagnostic_codes
            and is_efficiency_out_of_review_range(efficiency_percent)
        ):
            parameter_updates["efficiency_percent"] = {
                "action": "clamp_to_review_range",
                "suggested_range": suggested_efficiency_review_range(),
            }
        synthesized_patch["preserve_topology"] = True

    if "COMPRESSOR_OUTLET_PRESSURE_TARGET_TOO_CLOSE" in diagnostic_codes:
        minimum_delta_kpa = 95
        if "COMPRESSOR_ROOT_CAUSE_UNCONFIRMED" in diagnostic_codes:
            minimum_delta_kpa = 80
        elif "UNIT_PARAMETER_OUT_OF_RANGE" in diagnostic_codes:
            minimum_delta_kpa = 90
        pressure_update = {
            "action": "review_and_raise_margin",
            "minimum_delta_kpa": minimum_delta_kpa,
        }
        if primary_inlet_stream_id:
            pressure_update["reference_stream_id"] = primary_inlet_stream_id
        parameter_updates["outlet_pressure_target_kpa"] = pressure_update
        synthesized_patch["preserve_topology"] = True
        synthesized_patch["review_scope"] = "local_unit_parameters_only"

    if "COMPRESSOR_MINIMUM_FLOW_BYPASS_TOO_LOW" in diagnostic_codes:
        parameter_updates["minimum_flow_bypass_percent"] = {
            "action": "raise_to_anti_surge_review_floor",
            "suggested_minimum": 8,
        }
        synthesized_patch["preserve_topology"] = True

    if "UNIT_PARAMETER_OUT_OF_RANGE" in diagnostic_codes:
        if (
            "efficiency_percent" not in parameter_updates
            and is_efficiency_out_of_review_range(efficiency_percent)
        ):
            parameter_updates["efficiency_percent"] = {
                "action": "clamp_to_review_range",
                "suggested_range": suggested_efficiency_review_range(),
            }
        if "efficiency_percent" in parameter_updates:
            synthesized_patch["preserve_topology"] = True

    if "VALVE_PRESSURE_DROP_INVALID" in diagnostic_codes:
        parameter_updates["pressure_drop_kpa"] = {
            "action": "review_and_make_positive",
            "minimum_value": 10,
        }
        if isinstance(config.get("opening_percent"), (int, float)):
            parameter_updates["opening_percent"] = {
                "action": "review_with_throttling_margin",
                "suggested_maximum": 85,
            }
        synthesized_patch["patch_scope"] = "local_unit_parameters"

    if parameter_updates:
        synthesized_patch["parameter_updates"] = parameter_updates
    if parameter_placeholders:
        synthesized_patch["parameter_placeholders"] = list(dict.fromkeys(parameter_placeholders))
    if parameter_placeholders:
        synthesized_patch["patch_scope"] = "unit_parameters"
    return synthesized_patch


def merge_parameter_patch(
    *,
    existing_patch: dict[str, Any],
    synthesized_patch: dict[str, Any],
) -> dict[str, Any]:
    merged_patch: dict[str, Any] = {}
    synthesized_parameter_updates = normalize_parameter_updates(synthesized_patch.get("parameter_updates"))
    existing_parameter_updates = normalize_parameter_updates(existing_patch.get("parameter_updates"))
    synthesized_parameter_placeholders = normalize_parameter_placeholders(
        list(synthesized_patch.get("parameter_placeholders") or [])
    )
    parameter_updates = dict(synthesized_parameter_updates)
    if not parameter_updates and not synthesized_parameter_placeholders:
        parameter_updates = existing_parameter_updates
    if parameter_updates:
        merged_patch["parameter_updates"] = parameter_updates

    parameter_placeholders: list[str] = []
    existing_parameter_placeholders = normalize_parameter_placeholders(
        list(existing_patch.get("parameter_placeholders") or [])
    )
    if synthesized_parameter_placeholders:
        parameter_placeholders = list(synthesized_parameter_placeholders)
        for placeholder in existing_parameter_placeholders:
            if (
                placeholder == "operating_pressure_kpa"
                and "outlet_pressure_target_kpa" in synthesized_parameter_placeholders
                and placeholder not in synthesized_parameter_placeholders
            ):
                continue
            if placeholder not in parameter_placeholders:
                parameter_placeholders.append(placeholder)
    elif parameter_updates:
        parameter_placeholders = []
    else:
        parameter_placeholders = list(existing_parameter_placeholders)
    if parameter_placeholders:
        merged_patch["parameter_placeholders"] = parameter_placeholders

    patch_scope = str(
        synthesized_patch.get("patch_scope")
        or existing_patch.get("patch_scope")
        or ""
    ).strip()
    if patch_scope:
        merged_patch["patch_scope"] = patch_scope

    preserve_existing_connections: bool | None = None
    if "preserve_existing_connections" in existing_patch or "preserve_existing_connections" in synthesized_patch:
        preserve_existing_connections = bool(
            existing_patch.get(
                "preserve_existing_connections",
                synthesized_patch.get("preserve_existing_connections", True),
            )
        )
    elif patch_scope == "unit_parameters" and (parameter_placeholders or parameter_updates):
        preserve_existing_connections = True
    if preserve_existing_connections is not None:
        merged_patch["preserve_existing_connections"] = preserve_existing_connections

    if (
        ("preserve_topology" in existing_patch or "preserve_topology" in synthesized_patch)
        and not (patch_scope == "unit_parameters" and parameter_placeholders)
    ):
        merged_patch["preserve_topology"] = bool(
            existing_patch.get("preserve_topology", synthesized_patch.get("preserve_topology", True))
        )

    review_scope = normalize_text(existing_patch.get("review_scope") or synthesized_patch.get("review_scope"))
    if review_scope:
        merged_patch["review_scope"] = review_scope

    for key in ("retain_existing_specs",):
        if key in existing_patch:
            merged_patch[key] = bool(existing_patch.get(key))

    return merged_patch


def build_suggest_edits_response_citation_ids(
    *,
    copilot_request: dict[str, Any],
    diagnostic_index: int,
    diagnostic_code: str,
    severity: str,
    target_type: str,
    target_id: str,
    target_citation_lookup: dict[tuple[str, str], str],
    snapshot_citation_id: str,
    total_diagnostics: int,
    item_kind: str,
) -> list[str]:
    citation_ids: list[str] = [f"diag-{diagnostic_index}"]
    citation_ids.extend(
        build_suggest_edits_context_support_citation_ids(
            copilot_request=copilot_request,
            diagnostic_code=diagnostic_code,
            severity=severity,
            target_type=target_type,
            target_id=target_id,
            item_kind=item_kind,
            target_citation_lookup=target_citation_lookup,
            snapshot_citation_id=snapshot_citation_id,
        )
    )
    return list(dict.fromkeys(citation_ids))


def synthesize_suggest_edits_action(
    diagnostic: dict[str, Any],
    *,
    copilot_request: dict[str, Any],
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
        copilot_request=copilot_request,
        diagnostic_index=diagnostic_index,
        diagnostic_code=diagnostic_code,
        severity=str(diagnostic.get("severity") or "").strip().lower(),
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
    known_citation_ids = {
        str(citation.get("id") or "").strip()
        for citation in fallback_citations
        if str(citation.get("id") or "").strip()
    }
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
        if (
            target_type in {"stream", "unit"}
            and target_id
            and should_keep_suggest_edits_existing_action_target(
                target_type=target_type,
                target_id=target_id,
                diagnostics=diagnostics,
                copilot_request=copilot_request,
            )
            and (target_type, target_id) not in existing_actions_by_target
        ):
            existing_actions_by_target[(target_type, target_id)] = dict(action)

    normalized_actions: list[dict[str, Any]] = []
    actionable_diagnostic_codes = {
        "STREAM_DISCONNECTED",
        "STREAM_SPEC_MISSING",
        "UNIT_PARAMETER_INCOMPLETE",
        "UNIT_PARAMETER_OUT_OF_RANGE",
        "COMPRESSOR_OUTLET_PRESSURE_TARGET_INVALID",
        "COMPRESSOR_OUTLET_PRESSURE_TARGET_MISSING",
        "COMPRESSOR_MINIMUM_FLOW_BYPASS_MISSING",
        "COMPRESSOR_EFFICIENCY_BASELINE_MISSING",
        "COMPRESSOR_OUTLET_PRESSURE_TARGET_TOO_CLOSE",
        "COMPRESSOR_MINIMUM_FLOW_BYPASS_TOO_LOW",
        "PUMP_OUTLET_PRESSURE_TARGET_INVALID",
        "VALVE_PRESSURE_DROP_INVALID",
    }
    ordered_action_targets = build_suggest_edits_action_target_order(
        diagnostics,
        existing_actions_by_target,
        actionable_diagnostic_codes,
    )
    for target_type, target_id in ordered_action_targets:
        target_diagnostics_with_indexes = [
            (diagnostic_index, diagnostic)
            for diagnostic_index, diagnostic in enumerate(diagnostics, start=1)
            if str((diagnostic or {}).get("target_type") or "").strip().lower() == target_type
            and str((diagnostic or {}).get("target_id") or "").strip() == target_id
        ]
        contextual_diagnostics_with_indexes: list[tuple[int, dict[str, Any]]] = []
        if target_type == "stream" and target_diagnostics_with_indexes:
            primary_stream_diagnostic = target_diagnostics_with_indexes[0][1]
            if str((primary_stream_diagnostic or {}).get("code") or "").strip() == "STREAM_DISCONNECTED":
                _, stream_by_id, _ = build_flowsheet_lookup(copilot_request)
                stream_document = stream_by_id.get(target_id) or {}
                connected_unit_ids = {
                    str(stream_document.get("source_unit_id") or "").strip(),
                    str(stream_document.get("target_unit_id") or "").strip(),
                }
                contextual_diagnostics_with_indexes = [
                    (diagnostic_index, diagnostic)
                    for diagnostic_index, diagnostic in enumerate(diagnostics, start=1)
                    if str((diagnostic or {}).get("target_type") or "").strip().lower() == "unit"
                    and str((diagnostic or {}).get("target_id") or "").strip() in connected_unit_ids
                    and (
                        str((diagnostic or {}).get("code") or "").strip().startswith("DOWNSTREAM_")
                        or str((diagnostic or {}).get("code") or "").strip().endswith("_STATE_DEPENDENT")
                    )
                ]
        seen_action_diagnostic_indexes = {
            diagnostic_index for diagnostic_index, _ in target_diagnostics_with_indexes
        }
        action_context_diagnostics_with_indexes = [
            *target_diagnostics_with_indexes,
            *[
                (diagnostic_index, diagnostic)
                for diagnostic_index, diagnostic in contextual_diagnostics_with_indexes
                if diagnostic_index not in seen_action_diagnostic_indexes
            ],
            *[
                (diagnostic_index, diagnostic)
                for diagnostic_index, diagnostic in enumerate(diagnostics, start=1)
                if str((diagnostic or {}).get("target_type") or "").strip().lower() not in {"stream", "unit"}
                and diagnostic_index not in seen_action_diagnostic_indexes
            ],
        ]
        primary_diagnostic_index = target_diagnostics_with_indexes[0][0] if target_diagnostics_with_indexes else 1
        primary_diagnostic = target_diagnostics_with_indexes[0][1] if target_diagnostics_with_indexes else {
            "target_type": target_type,
            "target_id": target_id,
        }
        synthesized_action = synthesize_suggest_edits_action(
            primary_diagnostic,
            copilot_request=copilot_request,
            diagnostic_index=primary_diagnostic_index,
            total_diagnostics=len(diagnostics),
            target_citation_lookup=target_citation_lookup,
            snapshot_citation_id=snapshot_citation_id,
        )
        existing_action = existing_actions_by_target.get((target_type, target_id), {})
        merged_action = dict(existing_action)
        merged_action["kind"] = "candidate_edit"
        merged_action["target"] = {
            "type": target_type,
            "id": target_id,
        }
        merged_action["title"] = normalize_text(merged_action.get("title")) or synthesized_action["title"]
        merged_action["rationale"] = normalize_text(merged_action.get("rationale")) or synthesized_action["rationale"]
        merged_patch = merged_action.get("patch")
        if not isinstance(merged_patch, dict):
            merged_patch = {}
        synthesized_patch = dict(synthesized_action.get("patch") or {})
        primary_diagnostic_code = str(primary_diagnostic.get("code") or "").strip()
        if primary_diagnostic_code == "STREAM_DISCONNECTED":
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
                "retain_existing_source_binding": bool(
                    connection_placeholder.get("retain_existing_source_binding", True)
                ),
            }
        elif primary_diagnostic_code == "STREAM_SPEC_MISSING":
            spec_placeholders = normalize_spec_placeholders(merged_patch.get("spec_placeholders"))
            merged_patch = {
                "spec_placeholders": spec_placeholders
                or normalize_spec_placeholders(synthesized_patch.get("spec_placeholders")),
                "retain_existing_specs": bool(merged_patch.get("retain_existing_specs", True)),
            }
        else:
            structured_parameter_patch = synthesize_parameter_updates_from_diagnostics(
                target_type=target_type,
                target_id=target_id,
                diagnostics=[diagnostic for _, diagnostic in target_diagnostics_with_indexes],
                copilot_request=copilot_request,
            )
            merged_patch = merge_parameter_patch(
                existing_patch=merged_patch,
                synthesized_patch=structured_parameter_patch or synthesized_patch,
            )
        merged_action["patch"] = merged_patch
        merged_action["risk_level"] = max_risk_level(
            [str(existing_action.get("risk_level") or ""), str(synthesized_action.get("risk_level") or "")]
        )
        if primary_diagnostic_code == "STREAM_DISCONNECTED":
            merged_action["risk_level"] = "high"
        elif any(key in merged_patch for key in ("parameter_updates", "parameter_placeholders", "spec_placeholders")):
            merged_action["risk_level"] = "medium"
        merged_action["requires_confirmation"] = True
        merged_action["citation_ids"] = build_suggest_edits_group_citation_ids(
            copilot_request=copilot_request,
            diagnostics=[diagnostic for _, diagnostic in action_context_diagnostics_with_indexes],
            diagnostic_indexes=[diagnostic_index for diagnostic_index, _ in action_context_diagnostics_with_indexes],
            target_type=target_type,
            target_id=target_id,
            target_citation_lookup=target_citation_lookup,
            snapshot_citation_id=snapshot_citation_id,
        )
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
        severity = str(diagnostic.get("severity") or existing_issue.get("severity") or "warning").strip().lower() or "warning"
        issue_message = normalize_text(existing_issue.get("message")) or normalize_text(
            diagnostic.get("message") or diagnostic.get("text") or ""
        )
        canonical_issue_citation_ids = build_suggest_edits_response_citation_ids(
            copilot_request=copilot_request,
            diagnostic_index=diagnostic_index,
            diagnostic_code=diagnostic_code,
            severity=severity,
            target_type=target_type,
            target_id=target_id,
            target_citation_lookup=target_citation_lookup,
            snapshot_citation_id=snapshot_citation_id,
            total_diagnostics=len(diagnostics),
            item_kind="issue",
        )
        raw_issue_citation_ids = [str(citation_id).strip() for citation_id in (existing_issue.get("citation_ids") or []) if str(citation_id).strip()]
        normalized_existing_issue_citation_ids = [
            citation_id for citation_id in raw_issue_citation_ids if citation_id in known_citation_ids
        ]
        extras = [
            citation_id
            for citation_id in normalized_existing_issue_citation_ids
            if citation_id not in canonical_issue_citation_ids
        ]
        issue_citation_ids = canonical_issue_citation_ids + extras
        if severity != "error":
            ordered_warning_citation_ids = list(canonical_issue_citation_ids)
            extras = [
                citation_id
                for citation_id in issue_citation_ids
                if citation_id not in ordered_warning_citation_ids
            ]
            issue_citation_ids = ordered_warning_citation_ids + extras
        synthesized_issues.append(
            {
                "code": diagnostic_code,
                "message": issue_message or f"{target_id or '当前对象'} 存在需要人工复核的编辑问题。",
                "severity": severity,
                "citation_ids": list(dict.fromkeys(issue_citation_ids)),
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
            normalized_answer_citation_ids = list(
                dict.fromkeys(
                    citation_id
                    for citation_id in (answer.get("citation_ids") or [])
                    if str(citation_id).strip()
                    and (
                        citation_id in {
                            referenced_citation_id
                            for item in [*normalized_actions, *synthesized_issues]
                            if isinstance(item, dict)
                            for referenced_citation_id in (item.get("citation_ids") or [])
                        }
                    )
                )
            )
            answer["citation_ids"] = normalized_answer_citation_ids or list(
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

    coerced["issues"] = synthesized_issues
    coerced["proposed_actions"] = normalized_actions
    coerced["requires_confirmation"] = bool(normalized_actions)
    has_structured_parameter_updates = any(
        isinstance((action.get("patch") or {}).get("parameter_updates"), dict)
        and bool((action.get("patch") or {}).get("parameter_updates"))
        for action in normalized_actions
        if isinstance(action, dict)
    )
    if not normalized_actions:
        status = "failed"
    elif any(str(action.get("risk_level") or "").strip().lower() == "high" for action in normalized_actions):
        status = "partial"
    elif has_structured_parameter_updates:
        status = "partial"
    elif snapshot_citation_id and len(normalized_actions) > 1:
        status = "partial"
    elif any(str((diagnostic or {}).get("code") or "").strip() not in actionable_diagnostic_codes for diagnostic in diagnostics):
        status = "partial"
    else:
        status = "ok"
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
    if status == "partial" and snapshot_citation_id and snapshot_citation_id not in referenced_citation_ids:
        referenced_citation_ids.append(snapshot_citation_id)
    citation_by_id = {
        str(citation.get("id") or "").strip(): citation
        for citation in fallback_citations
        if str(citation.get("id") or "").strip()
    }
    ordered_citation_ids: list[str] = []
    for diagnostic_index, _diagnostic in enumerate(diagnostics, start=1):
        append_unique_citation_id(ordered_citation_ids, f"diag-{diagnostic_index}")
    action_target_citation_ids: list[str] = []
    for action in normalized_actions:
        target = action.get("target") or {}
        target_type = str((target or {}).get("type") or "").strip().lower()
        target_id = str((target or {}).get("id") or "").strip()
        target_citation_id = target_citation_lookup.get((target_type, target_id), "")
        if target_citation_id:
            action_target_citation_ids.append(target_citation_id)
            append_unique_citation_id(ordered_citation_ids, target_citation_id)
    action_target_citation_id_set = set(action_target_citation_ids)
    ordered_unit_action_supporting_citation_ids: list[str] = []
    ordered_stream_action_supporting_citation_ids: list[str] = []
    ordered_residual_supporting_citation_ids: list[str] = []

    def is_stream_artifact_citation_id(citation_id: str) -> bool:
        return str(citation_id).strip().startswith("flowdoc-stream-")

    def append_supporting_citation_id(citation_ids: list[str], citation_id: str) -> None:
        normalized_citation_id = str(citation_id).strip()
        if not normalized_citation_id:
            return
        if (
            normalized_citation_id.startswith("diag-")
            or normalized_citation_id == snapshot_citation_id
            or normalized_citation_id in action_target_citation_id_set
        ):
            return
        append_unique_citation_id(citation_ids, normalized_citation_id)

    for action in normalized_actions:
        target = action.get("target") or {}
        target_type = str((target or {}).get("type") or "").strip().lower()
        if target_type != "unit":
            continue
        for citation_id in (action.get("citation_ids") or []):
            append_supporting_citation_id(ordered_unit_action_supporting_citation_ids, str(citation_id))
    for action in normalized_actions:
        target = action.get("target") or {}
        target_type = str((target or {}).get("type") or "").strip().lower()
        if target_type != "stream":
            continue
        for citation_id in (action.get("citation_ids") or []):
            append_supporting_citation_id(ordered_stream_action_supporting_citation_ids, str(citation_id))
    for issue in synthesized_issues:
        for citation_id in (issue.get("citation_ids") or []):
            append_supporting_citation_id(ordered_residual_supporting_citation_ids, str(citation_id))
    for answer in coerced.get("answers") or []:
        if not isinstance(answer, dict):
            continue
        for citation_id in (answer.get("citation_ids") or []):
            append_supporting_citation_id(ordered_residual_supporting_citation_ids, str(citation_id))
    supporting_citation_groups = [
        ordered_unit_action_supporting_citation_ids,
        ordered_stream_action_supporting_citation_ids,
    ]
    if (
        ordered_unit_action_supporting_citation_ids
        and ordered_stream_action_supporting_citation_ids
        and all(is_stream_artifact_citation_id(citation_id) for citation_id in ordered_unit_action_supporting_citation_ids)
        and all(is_stream_artifact_citation_id(citation_id) for citation_id in ordered_stream_action_supporting_citation_ids)
    ):
        supporting_citation_groups = [
            ordered_stream_action_supporting_citation_ids,
            ordered_unit_action_supporting_citation_ids,
        ]
    for citation_group in supporting_citation_groups:
        for citation_id in citation_group:
            append_unique_citation_id(ordered_citation_ids, citation_id)
    for citation_id in ordered_residual_supporting_citation_ids:
        append_unique_citation_id(ordered_citation_ids, citation_id)
    if snapshot_citation_id and snapshot_citation_id in referenced_citation_ids:
        append_unique_citation_id(ordered_citation_ids, snapshot_citation_id)
    for citation_id in referenced_citation_ids:
        append_unique_citation_id(ordered_citation_ids, citation_id)
    coerced["citations"] = [
        citation_by_id[citation_id]
        for citation_id in ordered_citation_ids
        if citation_id in referenced_citation_ids and citation_id in citation_by_id
    ]
    coerced["status"] = status
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
    candidate_subset: list[dict[str, Any]]
    if has_default_tab_candidate and primary_candidate:
        candidate_subset = [primary_candidate]
        for candidate in legal_candidates:
            if str(candidate.get("candidate_ref") or "").strip() == str(primary_candidate.get("candidate_ref") or "").strip():
                continue
            candidate_subset.append(candidate)
            break
    else:
        candidate_subset = legal_candidates[:2]
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
