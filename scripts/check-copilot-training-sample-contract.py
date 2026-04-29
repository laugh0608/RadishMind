#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from scripts.checks.copilot_training_sample import check_training_sample  # noqa: E402

TRAINING_SAMPLE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-training-sample.schema.json"
COPILOT_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-request.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/copilot-training-sample-basic.json"


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def main() -> int:
    training_sample_schema = load_json_document(TRAINING_SAMPLE_SCHEMA_PATH)
    copilot_request_schema = load_json_document(COPILOT_REQUEST_SCHEMA_PATH)
    copilot_response_schema = load_json_document(COPILOT_RESPONSE_SCHEMA_PATH)
    fixture = load_json_document(FIXTURE_PATH)

    jsonschema.Draft202012Validator.check_schema(training_sample_schema)
    jsonschema.validate(fixture, training_sample_schema)
    if not isinstance(fixture, dict):
        raise SystemExit("copilot training sample fixture must be an object")
    jsonschema.validate(fixture["input_request"], copilot_request_schema)
    jsonschema.validate(fixture["target_response"], copilot_response_schema)
    try:
        check_training_sample(fixture)
    except ValueError as exc:
        raise SystemExit(str(exc)) from exc

    print("copilot training sample contract smoke passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
