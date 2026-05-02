from __future__ import annotations

import json
from typing import Any, NamedTuple


class ExtractedJsonObject(NamedTuple):
    document: dict[str, Any]
    json_text: str
    start_index: int
    end_index: int
    cleanup_applied: bool


def extract_json_object(text: str) -> ExtractedJsonObject:
    start = text.find("{")
    if start < 0:
        raise ValueError("local_transformers output did not contain a JSON object")
    depth = 0
    in_string = False
    escaped = False
    for index in range(start, len(text)):
        char = text[index]
        if in_string:
            if escaped:
                escaped = False
            elif char == "\\":
                escaped = True
            elif char == '"':
                in_string = False
            continue
        if char == '"':
            in_string = True
        elif char == "{":
            depth += 1
        elif char == "}":
            depth -= 1
            if depth == 0:
                json_text = text[start : index + 1]
                document, cleanup_applied = parse_json_object_with_cleanup(json_text)
                if not isinstance(document, dict):
                    raise ValueError("local_transformers output JSON must be an object")
                return ExtractedJsonObject(
                    document=document,
                    json_text=json_text,
                    start_index=start,
                    end_index=index,
                    cleanup_applied=cleanup_applied,
                )
    raise ValueError("local_transformers output JSON object was not closed")


def parse_json_object_with_cleanup(json_text: str) -> tuple[Any, bool]:
    try:
        return json.loads(json_text), False
    except json.JSONDecodeError as original_exc:
        cleaned = remove_trailing_json_commas(json_text)
        if cleaned == json_text:
            raise original_exc
        try:
            return json.loads(cleaned), True
        except json.JSONDecodeError:
            raise original_exc


def remove_trailing_json_commas(json_text: str) -> str:
    output: list[str] = []
    in_string = False
    escaped = False
    index = 0
    while index < len(json_text):
        char = json_text[index]
        if in_string:
            output.append(char)
            if escaped:
                escaped = False
            elif char == "\\":
                escaped = True
            elif char == '"':
                in_string = False
            index += 1
            continue
        if char == '"':
            in_string = True
            output.append(char)
            index += 1
            continue
        if char == ",":
            lookahead = index + 1
            while lookahead < len(json_text) and json_text[lookahead] in " \t\r\n":
                lookahead += 1
            if lookahead < len(json_text) and json_text[lookahead] in "}]":
                index += 1
                continue
        output.append(char)
        index += 1
    return "".join(output)
