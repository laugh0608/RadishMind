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


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Check RadishFlow UI consumption semantics for CopilotGatewayEnvelope.",
    )
    parser.add_argument(
        "--summary-output",
        default="",
        help="Optional path for a stable RadishFlow UI consumption summary json.",
    )
    parser.add_argument(
        "--check-summary",
        default="",
        help="Optional expected stable RadishFlow UI consumption summary json.",
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


def assert_condition(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def assert_json_equal(expected: Any, actual: Any, *, label: str) -> None:
    if expected != actual:
        raise SystemExit(f"generated RadishFlow UI consumption summary does not match expected summary: {label}")


def load_sample_request() -> dict[str, Any]:
    sample = load_json_document(SAMPLE_PATH)
    request = sample.get("input_request") if isinstance(sample, dict) else None
    if not isinstance(request, dict):
        raise SystemExit(f"sample is missing input_request: {SAMPLE_PATH.relative_to(REPO_ROOT)}")
    return request


def build_gateway_envelopes() -> list[tuple[str, dict[str, Any]]]:
    success_request = load_sample_request()
    unsupported_request = copy.deepcopy(success_request)
    unsupported_request["task"] = "explain_diagnostics"
    cases = [
        ("proposal_ready", success_request),
        ("unsupported_route", unsupported_request),
        ("schema_invalid", {}),
    ]
    envelopes: list[tuple[str, dict[str, Any]]] = []
    for case_id, request in cases:
        envelope = handle_copilot_request(request, options=GatewayOptions(provider="mock"))
        validate_gateway_envelope(envelope)
        response = envelope.get("response")
        if isinstance(response, dict):
            validate_response_document(response)
        envelopes.append((case_id, envelope))
    return envelopes


def action_target_summary(action: dict[str, Any]) -> dict[str, Any]:
    target = action.get("target") if isinstance(action.get("target"), dict) else {}
    return {
        "type": target.get("type") or "",
        "id": target.get("id") or "",
    }


def candidate_card_summary(action: dict[str, Any]) -> dict[str, Any]:
    patch = action.get("patch") if isinstance(action.get("patch"), dict) else {}
    citation_ids = action.get("citation_ids") if isinstance(action.get("citation_ids"), list) else []
    return {
        "kind": action.get("kind") or "",
        "title": action.get("title") or "",
        "target": action_target_summary(action),
        "risk_level": action.get("risk_level") or "",
        "requires_confirmation": action.get("requires_confirmation") is True,
        "patch_keys": list(patch.keys()),
        "citation_count": len(citation_ids),
    }


def issue_chip_summary(issue: dict[str, Any]) -> dict[str, Any]:
    return {
        "code": issue.get("code") or "",
        "severity": issue.get("severity") or "",
    }


def display_state_for_envelope(envelope: dict[str, Any]) -> str:
    response = envelope.get("response")
    if envelope.get("status") == "failed":
        return "gateway_failed"
    if not isinstance(response, dict):
        return "empty"
    actions = response.get("proposed_actions")
    if isinstance(actions, list) and actions:
        return "proposal_ready"
    issues = response.get("issues")
    if isinstance(issues, list) and issues:
        return "read_only_notice"
    return "empty"


def build_ui_consumption_case(case_id: str, envelope: dict[str, Any]) -> dict[str, Any]:
    metadata = envelope.get("metadata") if isinstance(envelope.get("metadata"), dict) else {}
    provider = metadata.get("provider") if isinstance(metadata.get("provider"), dict) else {}
    response = envelope.get("response")
    response_object = response if isinstance(response, dict) else {}
    error = envelope.get("error") if isinstance(envelope.get("error"), dict) else None
    issues = response_object.get("issues") if isinstance(response_object.get("issues"), list) else []
    actions = response_object.get("proposed_actions") if isinstance(response_object.get("proposed_actions"), list) else []
    citations = response_object.get("citations") if isinstance(response_object.get("citations"), list) else []

    candidate_cards = [
        candidate_card_summary(action)
        for action in actions
        if isinstance(action, dict) and action.get("kind") == "candidate_edit"
    ]
    confirmation_required = bool(response_object.get("requires_confirmation")) or any(
        card["requires_confirmation"] for card in candidate_cards
    )
    return {
        "case_id": case_id,
        "display_state": display_state_for_envelope(envelope),
        "request_id": envelope.get("request_id") or "",
        "route": metadata.get("route") or "",
        "status": envelope.get("status") or "",
        "response_status": response_object.get("status") if isinstance(response, dict) else None,
        "error_banner": {
            "code": error.get("code") or "",
            "message_present": bool(error.get("message")),
        }
        if error
        else None,
        "audit_metadata": {
            "provider": provider.get("name") or "",
            "request_validated": metadata.get("request_validated") is True,
            "response_validated": metadata.get("response_validated") is True,
            "advisory_only": metadata.get("advisory_only") is True,
        },
        "summary_present": isinstance(response_object.get("summary"), str) and bool(response_object.get("summary")),
        "issue_chips": [
            issue_chip_summary(issue)
            for issue in issues
            if isinstance(issue, dict)
        ],
        "candidate_cards": candidate_cards,
        "citation_count": len(citations),
        "confirmation_required": confirmation_required,
        "can_write_flowsheet_document": False,
    }


def assert_ui_consumption_contract(summary: dict[str, Any]) -> None:
    cases = summary.get("cases") if isinstance(summary.get("cases"), list) else []
    assert_condition(len(cases) == 3, "RadishFlow UI consumption summary must cover three cases")
    by_id = {case.get("case_id"): case for case in cases if isinstance(case, dict)}

    proposal = by_id.get("proposal_ready") or {}
    assert_condition(proposal.get("display_state") == "proposal_ready", "proposal case display state mismatch")
    assert_condition(proposal.get("confirmation_required") is True, "proposal case must require confirmation")
    assert_condition(proposal.get("can_write_flowsheet_document") is False, "proposal case must not write flowsheet")
    cards = proposal.get("candidate_cards") if isinstance(proposal.get("candidate_cards"), list) else []
    assert_condition(len(cards) == 1, "proposal case should expose one candidate card")
    assert_condition(cards[0].get("requires_confirmation") is True, "candidate card must require confirmation")

    unsupported = by_id.get("unsupported_route") or {}
    assert_condition(unsupported.get("display_state") == "gateway_failed", "unsupported case display state mismatch")
    assert_condition((unsupported.get("error_banner") or {}).get("code") == "UNSUPPORTED_TASK", "unsupported error code mismatch")
    assert_condition(unsupported.get("can_write_flowsheet_document") is False, "unsupported case must not write flowsheet")

    invalid = by_id.get("schema_invalid") or {}
    assert_condition(invalid.get("display_state") == "gateway_failed", "invalid case display state mismatch")
    assert_condition((invalid.get("error_banner") or {}).get("code") == "REQUEST_SCHEMA_INVALID", "invalid error code mismatch")
    assert_condition(invalid.get("summary_present") is False, "invalid case should not fabricate response summary")
    assert_condition(invalid.get("can_write_flowsheet_document") is False, "invalid case must not write flowsheet")

    for case in cases:
        assert_condition(case.get("can_write_flowsheet_document") is False, "UI consumption must remain advisory-only")
        audit_metadata = case.get("audit_metadata") if isinstance(case.get("audit_metadata"), dict) else {}
        assert_condition(audit_metadata.get("advisory_only") is True, "UI consumption must preserve advisory metadata")


def build_ui_consumption_summary() -> dict[str, Any]:
    cases = [
        build_ui_consumption_case(case_id, envelope)
        for case_id, envelope in build_gateway_envelopes()
    ]
    summary = {
        "schema_version": 1,
        "ui_surface": "radishflow_suggest_flowsheet_edits_panel",
        "source": "copilot_gateway_envelope",
        "case_count": len(cases),
        "cases": cases,
    }
    assert_ui_consumption_contract(summary)
    return summary


def main() -> int:
    args = parse_args()
    summary = build_ui_consumption_summary()

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

    print("radishflow gateway UI consumption check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
