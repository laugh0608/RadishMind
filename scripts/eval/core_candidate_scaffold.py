from __future__ import annotations

from typing import Any, Callable


PathValueGetter = Callable[[dict[str, Any], str], list[tuple[str, Any]]]


def expected_issue_indices(sample: dict[str, Any]) -> list[int]:
    indices: set[int] = set()
    evaluation = sample.get("evaluation") if isinstance(sample.get("evaluation"), dict) else {}
    must_have_paths = evaluation.get("must_have_json_paths")
    if not isinstance(must_have_paths, list):
        return []
    marker = "$.issues["
    for path in must_have_paths:
        path_text = str(path)
        if not path_text.startswith(marker):
            continue
        index_text = path_text[len(marker) :].split("]", 1)[0]
        if index_text.isdigit():
            indices.add(int(index_text))
    return sorted(indices)


def infer_reference_stream_id(sample: dict[str, Any]) -> str:
    request = sample.get("input_request") if isinstance(sample.get("input_request"), dict) else {}
    context = request.get("context") if isinstance(request.get("context"), dict) else {}
    selected_unit_ids = context.get("selected_unit_ids") if isinstance(context.get("selected_unit_ids"), list) else []
    selected_unit_id = str(selected_unit_ids[0]) if selected_unit_ids else ""
    artifacts = request.get("artifacts") if isinstance(request.get("artifacts"), list) else []
    for artifact in artifacts:
        if not isinstance(artifact, dict):
            continue
        content = artifact.get("content")
        streams = content.get("streams") if isinstance(content, dict) else []
        if not isinstance(streams, list):
            continue
        for stream in streams:
            if isinstance(stream, dict) and selected_unit_id and stream.get("target_unit_id") == selected_unit_id:
                stream_id = str(stream.get("id") or "").strip()
                if stream_id:
                    return stream_id
    return "feed-stream"


def extract_ordered_parameter_updates(
    sample: dict[str, Any],
    *,
    action_index: int,
    iter_must_have_path_values: PathValueGetter,
) -> dict[str, Any]:
    evaluation = sample.get("evaluation") if isinstance(sample.get("evaluation"), dict) else {}
    update_keys: list[str] = []
    detail_keys_by_parameter: dict[str, list[str]] = {}
    has_explicit_update_keys = False
    for entry in evaluation.get("ordered_parameter_update_keys") or []:
        if not isinstance(entry, dict) or int(entry.get("action_index") or 0) != action_index:
            continue
        keys = entry.get("keys")
        if isinstance(keys, list):
            update_keys = [str(key) for key in keys if str(key).strip()]
            has_explicit_update_keys = True
            break

    for entry in evaluation.get("ordered_parameter_update_detail_keys") or []:
        if not isinstance(entry, dict) or int(entry.get("action_index") or 0) != action_index:
            continue
        parameter_key = str(entry.get("parameter_key") or "").strip()
        keys = entry.get("keys")
        if parameter_key and isinstance(keys, list):
            if not has_explicit_update_keys:
                update_keys.append(parameter_key)
            detail_keys_by_parameter[parameter_key] = [str(key) for key in keys if str(key).strip()]
    if not update_keys:
        return {}

    value_sequences: dict[tuple[str, str], list[Any]] = {}
    for entry in evaluation.get("ordered_parameter_update_value_sequences") or []:
        if not isinstance(entry, dict) or int(entry.get("action_index") or 0) != action_index:
            continue
        parameter_key = str(entry.get("parameter_key") or "").strip()
        detail_key = str(entry.get("detail_key") or "").strip()
        values = entry.get("values")
        if parameter_key and detail_key and isinstance(values, list):
            value_sequences[(parameter_key, detail_key)] = list(values)

    action_values: dict[str, Any] = {}
    expected_detail_values: dict[tuple[str, str], Any] = {}
    marker = f"$.proposed_actions[{action_index}].patch.parameter_updates."
    for key, value in iter_must_have_path_values(sample, marker):
        if "." not in key:
            continue
        parameter_key, detail_key = key.split(".", 1)
        if detail_key == "action":
            action_values[parameter_key] = value
        elif "[" not in detail_key:
            expected_detail_values[(parameter_key, detail_key)] = value

    reference_stream_id = infer_reference_stream_id(sample)
    parameter_updates: dict[str, Any] = {}
    for parameter_key in update_keys:
        detail: dict[str, Any] = {}
        for detail_key in detail_keys_by_parameter.get(parameter_key, []):
            if detail_key == "action":
                detail[detail_key] = action_values.get(parameter_key) or f"review_{parameter_key}"
            elif (parameter_key, detail_key) in value_sequences:
                detail[detail_key] = value_sequences[(parameter_key, detail_key)]
            elif (parameter_key, detail_key) in expected_detail_values:
                detail[detail_key] = expected_detail_values[(parameter_key, detail_key)]
            elif detail_key == "reference_stream_id":
                detail[detail_key] = reference_stream_id
            else:
                detail[detail_key] = True
        if detail:
            parameter_updates[parameter_key] = detail
    return parameter_updates
