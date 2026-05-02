#!/usr/bin/env python3
from __future__ import annotations

import argparse
import copy
import json
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.gateway import GatewayOptions, handle_copilot_request, validate_gateway_envelope  # noqa: E402
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


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Check the in-process Copilot Gateway call contract with stable smoke cases.",
    )
    parser.add_argument(
        "--summary-output",
        default="",
        help="Optional path for a stable gateway smoke summary json.",
    )
    parser.add_argument(
        "--check-summary",
        default="",
        help="Optional expected stable gateway smoke summary json. Fails if regenerated summary differs.",
    )
    return parser.parse_args()


def resolve_repo_path(path_value: str) -> Path:
    path = Path(path_value)
    if not path.is_absolute():
        path = REPO_ROOT / path
    return path


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to load json document: {path.relative_to(REPO_ROOT)}: {exc}") from exc


def write_json_document(path: Path, document: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def assert_json_equal(expected: Any, actual: Any, *, label: str) -> None:
    if expected != actual:
        raise SystemExit(f"generated gateway service smoke summary does not match expected summary: {label}")


def check_successful_gateway_call() -> dict[str, Any]:
    request = load_sample_request()
    envelope = handle_copilot_request(request, options=GatewayOptions(provider="mock"))
    validate_gateway_envelope(envelope)
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
    return envelope


def check_unsupported_route() -> dict[str, Any]:
    request = copy.deepcopy(load_sample_request())
    request["task"] = "explain_diagnostics"
    envelope = handle_copilot_request(request, options=GatewayOptions(provider="mock"))
    validate_gateway_envelope(envelope)
    response = envelope.get("response")
    error = envelope.get("error") or {}
    assert_condition(envelope.get("status") == "failed", "unsupported route should return failed envelope")
    assert_condition(error.get("code") == "UNSUPPORTED_TASK", "unsupported route error code mismatch")
    assert_condition(isinstance(response, dict), "unsupported route should still return failed CopilotResponse")
    validate_response_document(response)
    return envelope


def check_invalid_request() -> dict[str, Any]:
    envelope = handle_copilot_request({}, options=GatewayOptions(provider="mock"))
    validate_gateway_envelope(envelope)
    error = envelope.get("error") or {}
    assert_condition(envelope.get("status") == "failed", "invalid request should return failed envelope")
    assert_condition(error.get("code") == "REQUEST_SCHEMA_INVALID", "invalid request error code mismatch")
    assert_condition(envelope.get("response") is None, "invalid schema request should not fabricate CopilotResponse")
    return envelope


def summarize_gateway_envelope(case_id: str, envelope: dict[str, Any]) -> dict[str, Any]:
    metadata = envelope.get("metadata") or {}
    provider = metadata.get("provider") or {}
    response = envelope.get("response")
    error = envelope.get("error")
    response_object = response if isinstance(response, dict) else {}
    issues = response_object.get("issues") if isinstance(response_object.get("issues"), list) else []
    proposed_actions = (
        response_object.get("proposed_actions")
        if isinstance(response_object.get("proposed_actions"), list)
        else []
    )
    return {
        "case_id": case_id,
        "status": envelope.get("status"),
        "request_id": envelope.get("request_id"),
        "project": envelope.get("project"),
        "task": envelope.get("task"),
        "route": metadata.get("route"),
        "request_validated": metadata.get("request_validated"),
        "response_validated": metadata.get("response_validated"),
        "advisory_only": metadata.get("advisory_only"),
        "provider": {
            "name": provider.get("name"),
            "profile": provider.get("profile"),
            "model": provider.get("model"),
            "request_timeout_seconds": provider.get("request_timeout_seconds"),
        },
        "response_present": isinstance(response, dict),
        "response_status": response_object.get("status") if isinstance(response, dict) else None,
        "response_requires_confirmation": (
            response_object.get("requires_confirmation") if isinstance(response, dict) else None
        ),
        "response_issue_codes": [
            issue.get("code")
            for issue in issues
            if isinstance(issue, dict) and str(issue.get("code") or "").strip()
        ],
        "response_action_count": len(proposed_actions),
        "error_code": error.get("code") if isinstance(error, dict) else None,
    }


def build_gateway_smoke_summary() -> dict[str, Any]:
    cases = [
        summarize_gateway_envelope("successful_radishflow_suggest_edits", check_successful_gateway_call()),
        summarize_gateway_envelope("unsupported_valid_route", check_unsupported_route()),
        summarize_gateway_envelope("schema_invalid_request", check_invalid_request()),
    ]
    return {
        "schema_version": 1,
        "call_mode": "in_process_python",
        "case_count": len(cases),
        "cases": cases,
    }


def main() -> int:
    args = parse_args()
    summary = build_gateway_smoke_summary()

    if args.summary_output.strip():
        write_json_document(resolve_repo_path(args.summary_output), summary)
    if args.check_summary.strip():
        expected_summary_path = resolve_repo_path(args.check_summary)
        if not expected_summary_path.is_file():
            raise SystemExit(f"expected summary file not found: {args.check_summary}")
        assert_json_equal(
            load_json_document(expected_summary_path),
            summary,
            label=expected_summary_path.relative_to(REPO_ROOT).as_posix(),
        )

    print("gateway service smoke passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
