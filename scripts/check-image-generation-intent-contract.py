#!/usr/bin/env python3
from __future__ import annotations

import sys
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from scripts.checks.image_generation import (  # noqa: E402
    check_artifact,
    check_backend_request,
    check_image_generation_intent,
    load_json_document,
    with_confirmation_required,
)

INTENT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-intent.schema.json"
BACKEND_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-backend-request.schema.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
INTENT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-intent-basic.json"
BACKEND_REQUEST_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-backend-request-basic.json"
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"


def check_confirmation_gate(intent_schema: dict[str, Any], backend_schema: dict[str, Any], intent: dict[str, Any]) -> None:
    confirmation_intent = with_confirmation_required(intent)
    jsonschema.validate(confirmation_intent, intent_schema)
    check_image_generation_intent(confirmation_intent)

    blocked_backend_request = load_json_document(BACKEND_REQUEST_FIXTURE_PATH)
    blocked_backend_request["safety"] = {
        **blocked_backend_request["safety"],
        "gate": "blocked_requires_confirmation",
        "requires_confirmation": True,
        "risk_level": "medium",
        "review_notes": ["blocked because the source intent requires confirmation"],
    }
    jsonschema.validate(blocked_backend_request, backend_schema)
    if blocked_backend_request["safety"]["gate"] != "blocked_requires_confirmation":
        raise SystemExit("blocked backend request must keep blocked gate")
    if blocked_backend_request["safety"]["requires_confirmation"] is not True:
        raise SystemExit("blocked backend request must preserve confirmation requirement")
    if blocked_backend_request["safety"]["risk_level"] != "medium":
        raise SystemExit("blocked backend request must preserve medium risk")
    if blocked_backend_request.get("intent_id") != confirmation_intent.get("intent_id"):
        raise SystemExit("blocked backend request intent_id mismatch")
    if blocked_backend_request.get("output") != confirmation_intent.get("output"):
        raise SystemExit("blocked backend request output must still mirror intent")
    if not any(str(note).strip() for note in blocked_backend_request["safety"].get("review_notes") or []):
        raise SystemExit("blocked backend request must include review notes")


def main() -> int:
    intent_schema = load_json_document(INTENT_SCHEMA_PATH)
    backend_schema = load_json_document(BACKEND_REQUEST_SCHEMA_PATH)
    artifact_schema = load_json_document(ARTIFACT_SCHEMA_PATH)
    intent_fixture = load_json_document(INTENT_FIXTURE_PATH)
    backend_fixture = load_json_document(BACKEND_REQUEST_FIXTURE_PATH)
    artifact_fixture = load_json_document(ARTIFACT_FIXTURE_PATH)

    for schema in (intent_schema, backend_schema, artifact_schema):
        jsonschema.Draft202012Validator.check_schema(schema)

    jsonschema.validate(intent_fixture, intent_schema)
    jsonschema.validate(backend_fixture, backend_schema)
    jsonschema.validate(artifact_fixture, artifact_schema)

    if not isinstance(intent_fixture, dict):
        raise SystemExit("image generation intent fixture must be an object")
    if not isinstance(backend_fixture, dict):
        raise SystemExit("image generation backend request fixture must be an object")
    if not isinstance(artifact_fixture, dict):
        raise SystemExit("image generation artifact fixture must be an object")

    check_image_generation_intent(intent_fixture)
    check_backend_request(intent_fixture, backend_fixture, expected_gate="approved_for_backend")
    check_artifact(intent_fixture, backend_fixture, artifact_fixture)
    check_confirmation_gate(intent_schema, backend_schema, intent_fixture)

    print("image generation contract smoke passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
