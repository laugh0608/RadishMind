#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-intent.schema.json"
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-intent-basic.json"
ALLOWED_RISK_LEVELS = {"low", "medium", "high"}


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def assert_condition(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def can_submit_to_backend(intent: dict[str, Any]) -> bool:
    safety = intent.get("safety") if isinstance(intent.get("safety"), dict) else {}
    return safety.get("requires_confirmation") is not True


def check_image_generation_intent(intent: dict[str, Any]) -> None:
    output = intent.get("output") if isinstance(intent.get("output"), dict) else {}
    prompt = intent.get("prompt") if isinstance(intent.get("prompt"), dict) else {}
    safety = intent.get("safety") if isinstance(intent.get("safety"), dict) else {}

    assert_condition(intent.get("schema_version") == 1, "image generation intent schema_version must be 1")
    assert_condition(intent.get("kind") == "image_generation", "image generation intent kind mismatch")
    assert_condition(int(output.get("width") or 0) > 0, "image generation output.width must be positive")
    assert_condition(int(output.get("height") or 0) > 0, "image generation output.height must be positive")
    assert_condition(int(output.get("count") or 0) >= 1, "image generation output.count must be at least 1")
    assert_condition(str(prompt.get("positive") or "").strip() != "", "image generation prompt.positive must be non-empty")
    assert_condition(safety.get("risk_level") in ALLOWED_RISK_LEVELS, "image generation safety.risk_level mismatch")

    if safety.get("requires_confirmation") is True:
        assert_condition(
            not can_submit_to_backend(intent),
            "image generation intent requiring confirmation must not be backend-submittable",
        )


def main() -> int:
    schema = load_json_document(SCHEMA_PATH)
    fixture = load_json_document(FIXTURE_PATH)

    jsonschema.Draft202012Validator.check_schema(schema)
    jsonschema.validate(fixture, schema)
    if not isinstance(fixture, dict):
        raise SystemExit("image generation intent fixture must be an object")
    check_image_generation_intent(fixture)

    confirmation_fixture = dict(fixture)
    confirmation_fixture["safety"] = {
        **fixture["safety"],
        "requires_confirmation": True,
        "risk_level": "medium",
        "review_notes": ["requires manual review before backend submission"],
    }
    jsonschema.validate(confirmation_fixture, schema)
    check_image_generation_intent(confirmation_fixture)

    print("image generation intent contract smoke passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
