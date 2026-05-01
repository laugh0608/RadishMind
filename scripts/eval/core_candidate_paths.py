from __future__ import annotations

from typing import Any


def get_evaluation(sample: dict[str, Any]) -> dict[str, Any]:
    evaluation = sample.get("evaluation")
    return evaluation if isinstance(evaluation, dict) else {}


def parse_json_path_expected_value(raw_value: str) -> Any:
    stripped = raw_value.strip()
    if stripped == "true":
        return True
    if stripped == "false":
        return False
    if stripped.startswith('"') and stripped.endswith('"'):
        return stripped[1:-1]
    return stripped


def iter_must_have_path_values(sample: dict[str, Any], prefix: str) -> list[tuple[str, Any]]:
    values: list[tuple[str, Any]] = []
    must_have_paths = get_evaluation(sample).get("must_have_json_paths")
    if not isinstance(must_have_paths, list):
        return values
    for path in must_have_paths:
        path_text = str(path)
        if not path_text.startswith(prefix) or "=" not in path_text:
            continue
        key = path_text[len(prefix) :].split("=", 1)[0]
        raw_value = path_text.split("=", 1)[1]
        if key:
            values.append((key, parse_json_path_expected_value(raw_value)))
    return values


def expected_action_kinds_by_index(sample: dict[str, Any]) -> dict[int, str]:
    action_kinds: dict[int, str] = {}
    must_have_paths = get_evaluation(sample).get("must_have_json_paths")
    if not isinstance(must_have_paths, list):
        return action_kinds
    marker = "$.proposed_actions["
    for path in must_have_paths:
        path_text = str(path)
        if not path_text.startswith(marker) or "].kind=" not in path_text:
            continue
        index_text = path_text[len(marker) :].split("]", 1)[0]
        if not index_text.isdigit():
            continue
        raw_value = path_text.split("=", 1)[1]
        value = parse_json_path_expected_value(raw_value)
        if isinstance(value, str) and value:
            action_kinds[int(index_text)] = value
    return action_kinds


def must_not_have_path(sample: dict[str, Any], path: str) -> bool:
    paths = get_evaluation(sample).get("must_not_have_json_paths")
    return isinstance(paths, list) and path in {str(item) for item in paths}


def set_nested_patch_value(patch: dict[str, Any], key_path: str, value: Any) -> None:
    segments = key_path.split(".")
    current: Any = patch
    for index, segment in enumerate(segments):
        is_last = index == len(segments) - 1
        if "[" in segment and segment.endswith("]"):
            key = segment.split("[", 1)[0]
            index_text = segment.split("[", 1)[1].rstrip("]")
            if not index_text.isdigit():
                return
            item_index = int(index_text)
            target_list = current.setdefault(key, []) if isinstance(current, dict) else []
            if not isinstance(target_list, list):
                return
            while len(target_list) <= item_index:
                target_list.append(None)
            if is_last:
                target_list[item_index] = value
            else:
                if not isinstance(target_list[item_index], dict):
                    target_list[item_index] = {}
                current = target_list[item_index]
            continue
        if is_last:
            current[segment] = value
            continue
        current = current.setdefault(segment, {})
        if not isinstance(current, dict):
            return
