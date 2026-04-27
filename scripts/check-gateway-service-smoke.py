#!/usr/bin/env python3
from __future__ import annotations

import copy
import json
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.gateway import GatewayOptions, handle_copilot_request  # noqa: E402
from services.runtime.inference import validate_response_document  # noqa: E402


SAMPLE_PATH = REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json"


def load_sample_request() -> dict[str, Any]:
    sample = json.loads(SAMPLE_PATH.read_text(encoding="utf-8"))
    request = sample.get("input_request")
    if not isinstance(request, dict):
        raise SystemExit(f"sample is missing input_request: {SAMPLE_PATH.relative_to(REPO_ROOT)}")
    return request


def assert_condition(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def check_successful_gateway_call() -> None:
    request = load_sample_request()
    envelope = handle_copilot_request(request, options=GatewayOptions(provider="mock"))
    response = envelope.get("response")
    assert_condition(envelope.get("schema_version") == 1, "gateway envelope schema_version mismatch")
    assert_condition(envelope.get("status") in {"ok", "partial"}, "gateway smoke expected ok or partial status")
    assert_condition(isinstance(response, dict), "gateway smoke response must be an object")
    validate_response_document(response)
    metadata = envelope.get("metadata") or {}
    provider = metadata.get("provider") or {}
    assert_condition(metadata.get("route") == "radishflow/suggest_flowsheet_edits", "gateway route metadata mismatch")
    assert_condition(metadata.get("advisory_only") is True, "gateway must remain advisory-only")
    assert_condition(metadata.get("request_validated") is True, "gateway did not mark request as validated")
    assert_condition(metadata.get("response_validated") is True, "gateway did not mark response as validated")
    assert_condition(provider.get("name") == "mock", "gateway provider metadata mismatch")
    assert_condition(response.get("requires_confirmation") is True, "suggest edits response must require confirmation")


def check_unsupported_route() -> None:
    request = copy.deepcopy(load_sample_request())
    request["task"] = "explain_diagnostics"
    envelope = handle_copilot_request(request, options=GatewayOptions(provider="mock"))
    response = envelope.get("response")
    error = envelope.get("error") or {}
    assert_condition(envelope.get("status") == "failed", "unsupported route should return failed envelope")
    assert_condition(error.get("code") == "UNSUPPORTED_TASK", "unsupported route error code mismatch")
    assert_condition(isinstance(response, dict), "unsupported route should still return failed CopilotResponse")
    validate_response_document(response)


def check_invalid_request() -> None:
    envelope = handle_copilot_request({}, options=GatewayOptions(provider="mock"))
    error = envelope.get("error") or {}
    assert_condition(envelope.get("status") == "failed", "invalid request should return failed envelope")
    assert_condition(error.get("code") == "REQUEST_SCHEMA_INVALID", "invalid request error code mismatch")
    assert_condition(envelope.get("response") is None, "invalid schema request should not fabricate CopilotResponse")


def main() -> int:
    check_successful_gateway_call()
    check_unsupported_route()
    check_invalid_request()
    print("gateway service smoke passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
