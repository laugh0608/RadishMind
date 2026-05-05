#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.inference import validate_response_document  # noqa: E402


SAMPLE_PATH = REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json"
GATEWAY_SERVICE_SUMMARY_PATH = REPO_ROOT / "scripts/checks/fixtures/gateway-service-smoke-summary.json"
GATEWAY_DEMO_SUMMARY_PATH = REPO_ROOT / "scripts/checks/fixtures/radishflow-gateway-demo-summary.json"
UI_CONSUMPTION_SUMMARY_PATH = REPO_ROOT / "scripts/checks/fixtures/radishflow-gateway-ui-consumption-summary.json"
CANDIDATE_HANDOFF_SUMMARY_PATH = REPO_ROOT / "scripts/checks/fixtures/radishflow-candidate-edit-handoff-summary.json"

ROUTE = "radishflow/suggest_flowsheet_edits"
PROJECT = "radishflow"
TASK = "suggest_flowsheet_edits"
PROVIDER = "mock"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Check the RadishFlow suggest_flowsheet_edits service/API smoke matrix.",
    )
    parser.add_argument(
        "--summary-output",
        default="",
        help="Optional path for a stable service smoke matrix summary json.",
    )
    parser.add_argument(
        "--check-summary",
        default="",
        help="Optional expected stable service smoke matrix summary json.",
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
        raise SystemExit(f"generated RadishFlow service smoke matrix summary does not match expected summary: {label}")


def load_sample_request() -> dict[str, Any]:
    sample = load_json_document(SAMPLE_PATH)
    request = sample.get("input_request") if isinstance(sample, dict) else None
    if not isinstance(request, dict):
        raise SystemExit(f"sample is missing input_request: {SAMPLE_PATH.relative_to(REPO_ROOT)}")
    assert_condition(request.get("project") == PROJECT, "matrix sample project mismatch")
    assert_condition(request.get("task") == TASK, "matrix sample task mismatch")
    return request


def repo_relative(path: Path) -> str:
    return path.relative_to(REPO_ROOT).as_posix()


def find_case(summary: dict[str, Any], case_id: str) -> dict[str, Any]:
    cases = summary.get("cases")
    if not isinstance(cases, list):
        raise SystemExit(f"summary is missing cases array while searching {case_id}")
    for case in cases:
        if isinstance(case, dict) and case.get("case_id") == case_id:
            return case
    raise SystemExit(f"summary is missing expected case: {case_id}")


def run_cli_runtime(sample_request: dict[str, Any]) -> dict[str, Any]:
    command = [
        sys.executable,
        str(REPO_ROOT / "scripts/run-copilot-inference.py"),
        "--sample",
        repo_relative(SAMPLE_PATH),
        "--provider",
        PROVIDER,
    ]
    completed = subprocess.run(
        command,
        cwd=REPO_ROOT,
        check=False,
        capture_output=True,
        text=True,
    )
    if completed.returncode != 0:
        raise SystemExit(
            "run-copilot-inference.py failed for service smoke matrix: "
            f"exit={completed.returncode}\n{completed.stderr.strip()}"
        )
    response = json.loads(completed.stdout)
    validate_response_document(response)
    actions = response.get("proposed_actions") if isinstance(response.get("proposed_actions"), list) else []
    candidate_edits = [
        action
        for action in actions
        if isinstance(action, dict) and action.get("kind") == "candidate_edit"
    ]
    assert_condition(response.get("requires_confirmation") is True, "CLI response must require confirmation")
    assert_condition(len(candidate_edits) == 1, "CLI response must expose one candidate_edit")
    target = candidate_edits[0].get("target") if isinstance(candidate_edits[0].get("target"), dict) else {}
    patch = candidate_edits[0].get("patch") if isinstance(candidate_edits[0].get("patch"), dict) else {}
    return {
        "entrypoint": "scripts/run-copilot-inference.py",
        "call_mode": "cli_runtime",
        "sample": repo_relative(SAMPLE_PATH),
        "request_id": sample_request.get("request_id") or "",
        "route": ROUTE,
        "provider": PROVIDER,
        "response_status": response.get("status") or "",
        "requires_confirmation": response.get("requires_confirmation") is True,
        "candidate_edit_count": len(candidate_edits),
        "target": {
            "type": target.get("type") or "",
            "id": target.get("id") or "",
        },
        "patch_keys": list(patch.keys()),
    }


def summarize_gateway_service(sample_request: dict[str, Any]) -> dict[str, Any]:
    summary = load_json_document(GATEWAY_SERVICE_SUMMARY_PATH)
    success = find_case(summary, "successful_radishflow_suggest_edits")
    unsupported = find_case(summary, "unsupported_valid_route")
    invalid = find_case(summary, "schema_invalid_request")
    assert_condition(success.get("request_id") == sample_request.get("request_id"), "gateway service request_id mismatch")
    assert_condition(success.get("route") == ROUTE, "gateway service route mismatch")
    assert_condition(success.get("provider", {}).get("name") == PROVIDER, "gateway service provider mismatch")
    assert_condition(success.get("advisory_only") is True, "gateway service must stay advisory-only")
    assert_condition(success.get("request_validated") is True, "gateway service request validation missing")
    assert_condition(success.get("response_validated") is True, "gateway service response validation missing")
    assert_condition(success.get("response_requires_confirmation") is True, "gateway service confirmation flag mismatch")
    assert_condition(unsupported.get("error_code") == "UNSUPPORTED_TASK", "gateway unsupported error mismatch")
    assert_condition(invalid.get("error_code") == "REQUEST_SCHEMA_INVALID", "gateway invalid error mismatch")
    return {
        "entrypoint": "scripts/check-gateway-service-smoke.py",
        "call_mode": summary.get("call_mode") or "",
        "summary_fixture": repo_relative(GATEWAY_SERVICE_SUMMARY_PATH),
        "success_case_id": success.get("case_id") or "",
        "route": success.get("route") or "",
        "provider": success.get("provider", {}).get("name") or "",
        "advisory_only": success.get("advisory_only") is True,
        "request_validated": success.get("request_validated") is True,
        "response_validated": success.get("response_validated") is True,
        "requires_confirmation": success.get("response_requires_confirmation") is True,
        "error_cases": [
            {
                "case_id": unsupported.get("case_id") or "",
                "error_code": unsupported.get("error_code") or "",
            },
            {
                "case_id": invalid.get("case_id") or "",
                "error_code": invalid.get("error_code") or "",
            },
        ],
    }


def summarize_gateway_demo() -> dict[str, Any]:
    summary = load_json_document(GATEWAY_DEMO_SUMMARY_PATH)
    results = summary.get("results") if isinstance(summary.get("results"), list) else []
    assert_condition(len(results) >= 3, "gateway demo matrix must cover at least three fixtures")
    request_ids: list[str] = []
    for result in results:
        assert_condition(isinstance(result, dict), "gateway demo result must be object")
        assert_condition(result.get("route") == ROUTE, "gateway demo route mismatch")
        assert_condition(result.get("provider") == PROVIDER, "gateway demo provider mismatch")
        assert_condition(result.get("advisory_only") is True, "gateway demo must stay advisory-only")
        assert_condition(result.get("request_validated") is True, "gateway demo request validation missing")
        assert_condition(result.get("response_validated") is True, "gateway demo response validation missing")
        assert_condition(result.get("requires_confirmation") is True, "gateway demo confirmation flag mismatch")
        request_ids.append(str(result.get("request_id") or ""))
    return {
        "entrypoint": "scripts/run-radishflow-gateway-demo.py",
        "call_mode": "export_adapter_request_gateway",
        "summary_fixture": repo_relative(GATEWAY_DEMO_SUMMARY_PATH),
        "manifest_fixture": "scripts/checks/fixtures/radishflow-gateway-demo-fixtures.json",
        "fixture_count": len(results),
        "route": ROUTE,
        "provider": PROVIDER,
        "advisory_only": True,
        "request_validated": True,
        "response_validated": True,
        "requires_confirmation": True,
        "request_ids": request_ids,
    }


def summarize_ui_consumption(sample_request: dict[str, Any]) -> dict[str, Any]:
    summary = load_json_document(UI_CONSUMPTION_SUMMARY_PATH)
    proposal = find_case(summary, "proposal_ready")
    unsupported = find_case(summary, "unsupported_route")
    invalid = find_case(summary, "schema_invalid")
    assert_condition(proposal.get("request_id") == sample_request.get("request_id"), "UI request_id mismatch")
    assert_condition(proposal.get("route") == ROUTE, "UI route mismatch")
    assert_condition(proposal.get("display_state") == "proposal_ready", "UI proposal state mismatch")
    assert_condition(proposal.get("confirmation_required") is True, "UI confirmation flag mismatch")
    assert_condition(proposal.get("can_write_flowsheet_document") is False, "UI must not write FlowsheetDocument")
    assert_condition(unsupported.get("can_write_flowsheet_document") is False, "unsupported UI case must not write")
    assert_condition(invalid.get("can_write_flowsheet_document") is False, "invalid UI case must not write")
    cards = proposal.get("candidate_cards") if isinstance(proposal.get("candidate_cards"), list) else []
    assert_condition(len(cards) == 1, "UI proposal must expose one candidate card")
    card = cards[0]
    return {
        "entrypoint": "scripts/check-radishflow-gateway-ui-consumption.py",
        "call_mode": "gateway_envelope_to_ui_consumption",
        "summary_fixture": repo_relative(UI_CONSUMPTION_SUMMARY_PATH),
        "ui_surface": summary.get("ui_surface") or "",
        "proposal_case_id": proposal.get("case_id") or "",
        "display_state": proposal.get("display_state") or "",
        "route": proposal.get("route") or "",
        "provider": proposal.get("audit_metadata", {}).get("provider") or "",
        "confirmation_required": proposal.get("confirmation_required") is True,
        "candidate_card_count": len(cards),
        "target": card.get("target") if isinstance(card, dict) else {},
        "patch_keys": card.get("patch_keys") if isinstance(card.get("patch_keys"), list) else [],
        "can_write_flowsheet_document": proposal.get("can_write_flowsheet_document") is True,
        "error_cases": [
            {
                "case_id": unsupported.get("case_id") or "",
                "display_state": unsupported.get("display_state") or "",
                "can_write_flowsheet_document": unsupported.get("can_write_flowsheet_document") is True,
            },
            {
                "case_id": invalid.get("case_id") or "",
                "display_state": invalid.get("display_state") or "",
                "can_write_flowsheet_document": invalid.get("can_write_flowsheet_document") is True,
            },
        ],
    }


def summarize_candidate_handoff(sample_request: dict[str, Any]) -> dict[str, Any]:
    summary = load_json_document(CANDIDATE_HANDOFF_SUMMARY_PATH)
    confirmed = find_case(summary, "confirmed_candidate_edit")
    unsupported = find_case(summary, "unsupported_route")
    invalid = find_case(summary, "schema_invalid")
    assert_condition(confirmed.get("source_request_id") == sample_request.get("request_id"), "handoff request_id mismatch")
    assert_condition(confirmed.get("source_route") == ROUTE, "handoff route mismatch")
    assert_condition(confirmed.get("eligible_action_count") == 1, "handoff must expose one eligible candidate")
    assert_condition(confirmed.get("handoff_blocked") is False, "confirmed handoff must not be blocked")
    assert_condition(confirmed.get("can_execute_any") is False, "handoff matrix must remain non-executing")
    command_candidates = (
        confirmed.get("command_candidates")
        if isinstance(confirmed.get("command_candidates"), list)
        else []
    )
    assert_condition(len(command_candidates) == 1, "confirmed handoff must contain one command candidate")
    candidate = command_candidates[0]
    assert_condition(candidate.get("requires_human_confirmation") is True, "handoff must require human confirmation")
    assert_condition(candidate.get("can_execute") is False, "handoff candidate must not execute")
    for blocked in (unsupported, invalid):
        assert_condition(blocked.get("handoff_blocked") is True, "failed handoff case must be blocked")
        assert_condition(blocked.get("can_execute_any") is False, "failed handoff case must not execute")
    return {
        "entrypoint": "scripts/check-radishflow-candidate-edit-handoff.py",
        "call_mode": "ui_confirmed_candidate_to_command_handoff",
        "summary_fixture": repo_relative(CANDIDATE_HANDOFF_SUMMARY_PATH),
        "handoff_surface": summary.get("handoff_surface") or "",
        "confirmed_case_id": confirmed.get("case_id") or "",
        "source_route": confirmed.get("source_route") or "",
        "eligible_action_count": confirmed.get("eligible_action_count") or 0,
        "handoff_blocked": confirmed.get("handoff_blocked") is True,
        "can_execute_any": confirmed.get("can_execute_any") is True,
        "command_candidate": {
            "handoff_kind": candidate.get("handoff_kind") or "",
            "target": candidate.get("target") if isinstance(candidate, dict) else {},
            "patch_keys": candidate.get("patch_keys") if isinstance(candidate.get("patch_keys"), list) else [],
            "risk_level": candidate.get("risk_level") or "",
            "requires_human_confirmation": candidate.get("requires_human_confirmation") is True,
            "can_execute": candidate.get("can_execute") is True,
        },
        "blocked_case_count": 2,
    }


def assert_matrix_contract(matrix: dict[str, Any]) -> None:
    surfaces = matrix.get("surfaces") if isinstance(matrix.get("surfaces"), dict) else {}
    cli = surfaces.get("cli_runtime") if isinstance(surfaces.get("cli_runtime"), dict) else {}
    gateway = surfaces.get("gateway_api") if isinstance(surfaces.get("gateway_api"), dict) else {}
    demo = surfaces.get("gateway_demo") if isinstance(surfaces.get("gateway_demo"), dict) else {}
    ui = surfaces.get("ui_consumption") if isinstance(surfaces.get("ui_consumption"), dict) else {}
    handoff = surfaces.get("candidate_handoff") if isinstance(surfaces.get("candidate_handoff"), dict) else {}

    for label, surface in (
        ("cli_runtime", cli),
        ("gateway_api", gateway),
        ("gateway_demo", demo),
        ("ui_consumption", ui),
    ):
        assert_condition(surface.get("route") == ROUTE, f"{label} route mismatch")
    assert_condition(handoff.get("source_route") == ROUTE, "candidate_handoff route mismatch")

    for label, surface in (
        ("cli_runtime", cli),
        ("gateway_api", gateway),
        ("gateway_demo", demo),
    ):
        assert_condition(surface.get("provider") == PROVIDER, f"{label} provider mismatch")

    assert_condition(cli.get("requires_confirmation") is True, "CLI confirmation mismatch")
    assert_condition(gateway.get("requires_confirmation") is True, "gateway confirmation mismatch")
    assert_condition(demo.get("requires_confirmation") is True, "demo confirmation mismatch")
    assert_condition(ui.get("confirmation_required") is True, "UI confirmation mismatch")
    assert_condition(ui.get("can_write_flowsheet_document") is False, "UI must remain non-writing")
    assert_condition(handoff.get("can_execute_any") is False, "handoff must remain non-executing")

    assert_condition(cli.get("target") == ui.get("target"), "CLI/UI target mismatch")
    assert_condition(cli.get("target") == handoff.get("command_candidate", {}).get("target"), "CLI/handoff target mismatch")
    assert_condition(cli.get("patch_keys") == ui.get("patch_keys"), "CLI/UI patch keys mismatch")
    assert_condition(
        cli.get("patch_keys") == handoff.get("command_candidate", {}).get("patch_keys"),
        "CLI/handoff patch keys mismatch",
    )


def build_service_smoke_matrix() -> dict[str, Any]:
    sample_request = load_sample_request()
    matrix = {
        "schema_version": 1,
        "matrix_id": "radishflow-suggest-flowsheet-edits-service-smoke",
        "route": ROUTE,
        "project": PROJECT,
        "task": TASK,
        "provider": PROVIDER,
        "sample": repo_relative(SAMPLE_PATH),
        "request_id": sample_request.get("request_id") or "",
        "purpose": "service_api_gate",
        "surfaces": {
            "cli_runtime": run_cli_runtime(sample_request),
            "gateway_api": summarize_gateway_service(sample_request),
            "gateway_demo": summarize_gateway_demo(),
            "ui_consumption": summarize_ui_consumption(sample_request),
            "candidate_handoff": summarize_candidate_handoff(sample_request),
        },
        "contract": {
            "route_consistent": True,
            "provider_consistent": True,
            "advisory_only": True,
            "requires_confirmation_preserved": True,
            "ui_does_not_write_flowsheet_document": True,
            "handoff_is_non_executing": True,
        },
    }
    assert_matrix_contract(matrix)
    return matrix


def main() -> int:
    args = parse_args()
    matrix = build_service_smoke_matrix()

    if args.summary_output.strip():
        write_json_document(resolve_repo_path(args.summary_output), matrix)
    if args.check_summary.strip():
        expected_summary_path = resolve_repo_path(args.check_summary)
        if not expected_summary_path.is_file():
            raise SystemExit(f"expected summary file not found: {args.check_summary}")
        assert_json_equal(
            load_json_document(expected_summary_path),
            matrix,
            label=expected_summary_path.relative_to(REPO_ROOT).as_posix(),
        )

    print("radishflow service smoke matrix passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
