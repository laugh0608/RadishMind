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
        description="Check RadishFlow candidate_edit handoff semantics after UI confirmation.",
    )
    parser.add_argument(
        "--summary-output",
        default="",
        help="Optional path for a stable RadishFlow candidate edit handoff summary json.",
    )
    parser.add_argument(
        "--check-summary",
        default="",
        help="Optional expected stable RadishFlow candidate edit handoff summary json.",
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
        raise SystemExit(f"generated RadishFlow candidate edit handoff summary does not match expected summary: {label}")


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
        ("confirmed_candidate_edit", success_request),
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


def target_summary(action: dict[str, Any]) -> dict[str, str]:
    target = action.get("target") if isinstance(action.get("target"), dict) else {}
    return {
        "type": str(target.get("type") or ""),
        "id": str(target.get("id") or ""),
    }


def command_candidate_from_action(
    *,
    request_id: str,
    route: str,
    action_index: int,
    action: dict[str, Any],
) -> dict[str, Any] | None:
    if action.get("kind") != "candidate_edit":
        return None
    if action.get("requires_confirmation") is not True:
        return None
    patch = action.get("patch") if isinstance(action.get("patch"), dict) else {}
    citation_ids = action.get("citation_ids") if isinstance(action.get("citation_ids"), list) else []
    return {
        "handoff_kind": "radishflow_candidate_edit_proposal",
        "command_candidate_id": f"{request_id or 'request'}:candidate_edit:{action_index}",
        "source_request_id": request_id,
        "source_route": route,
        "action_index": action_index,
        "title": action.get("title") or "",
        "target": target_summary(action),
        "patch_keys": list(patch.keys()),
        "risk_level": action.get("risk_level") or "",
        "citation_ids": citation_ids,
        "requires_human_confirmation": True,
        "can_execute": False,
    }


def build_handoff_case(case_id: str, envelope: dict[str, Any]) -> dict[str, Any]:
    metadata = envelope.get("metadata") if isinstance(envelope.get("metadata"), dict) else {}
    response = envelope.get("response")
    response_object = response if isinstance(response, dict) else {}
    actions = response_object.get("proposed_actions") if isinstance(response_object.get("proposed_actions"), list) else []
    route = str(metadata.get("route") or "")
    request_id = str(envelope.get("request_id") or "")
    command_candidates = [
        command_candidate
        for index, action in enumerate(actions)
        if isinstance(action, dict)
        for command_candidate in [
            command_candidate_from_action(
                request_id=request_id,
                route=route,
                action_index=index,
                action=action,
            )
        ]
        if command_candidate is not None
    ]
    return {
        "case_id": case_id,
        "source_status": envelope.get("status") or "",
        "source_route": route,
        "source_request_id": request_id,
        "eligible_action_count": len(command_candidates),
        "command_candidates": command_candidates,
        "handoff_blocked": len(command_candidates) == 0,
        "block_reason": "" if command_candidates else "no_confirmed_candidate_edit",
        "can_execute_any": False,
    }


def assert_handoff_contract(summary: dict[str, Any]) -> None:
    cases = summary.get("cases") if isinstance(summary.get("cases"), list) else []
    assert_condition(len(cases) == 3, "candidate edit handoff summary must cover three cases")
    by_id = {case.get("case_id"): case for case in cases if isinstance(case, dict)}

    confirmed = by_id.get("confirmed_candidate_edit") or {}
    assert_condition(confirmed.get("eligible_action_count") == 1, "confirmed case should produce one handoff candidate")
    assert_condition(confirmed.get("handoff_blocked") is False, "confirmed case should not be blocked")
    assert_condition(confirmed.get("can_execute_any") is False, "confirmed case must not execute directly")
    command_candidates = (
        confirmed.get("command_candidates")
        if isinstance(confirmed.get("command_candidates"), list)
        else []
    )
    assert_condition(len(command_candidates) == 1, "confirmed case should expose one command candidate")
    candidate = command_candidates[0]
    assert_condition(candidate.get("requires_human_confirmation") is True, "handoff candidate must require confirmation")
    assert_condition(candidate.get("can_execute") is False, "handoff candidate must not execute")
    assert_condition(candidate.get("target") == {"type": "stream", "id": "stream-out-1"}, "handoff target mismatch")
    assert_condition(candidate.get("patch_keys") == ["connection_placeholder"], "handoff patch keys mismatch")
    assert_condition(candidate.get("risk_level") == "high", "handoff risk level mismatch")
    assert_condition(len(candidate.get("citation_ids") or []) == 3, "handoff must preserve citation ids")

    for blocked_case_id in ("unsupported_route", "schema_invalid"):
        blocked = by_id.get(blocked_case_id) or {}
        assert_condition(blocked.get("eligible_action_count") == 0, f"{blocked_case_id} should not produce handoff")
        assert_condition(blocked.get("handoff_blocked") is True, f"{blocked_case_id} should be blocked")
        assert_condition(blocked.get("block_reason") == "no_confirmed_candidate_edit", f"{blocked_case_id} block reason mismatch")
        assert_condition(blocked.get("can_execute_any") is False, f"{blocked_case_id} must not execute")

    for case in cases:
        assert_condition(case.get("can_execute_any") is False, "candidate edit handoff must remain non-executing")


def build_handoff_summary() -> dict[str, Any]:
    cases = [
        build_handoff_case(case_id, envelope)
        for case_id, envelope in build_gateway_envelopes()
    ]
    summary = {
        "schema_version": 1,
        "handoff_surface": "radishflow_candidate_edit_confirmation",
        "source": "copilot_gateway_envelope",
        "case_count": len(cases),
        "cases": cases,
    }
    assert_handoff_contract(summary)
    return summary


def main() -> int:
    args = parse_args()
    summary = build_handoff_summary()

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

    print("radishflow candidate edit handoff check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
