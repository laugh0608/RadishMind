#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-promotion-gates.json"
SESSION_DOC_PATH = REPO_ROOT / "docs/contracts/session.md"
TOOLING_DOC_PATH = REPO_ROOT / "docs/contracts/tooling.md"
CONTRACTS_README_PATH = REPO_ROOT / "contracts/README.md"

REQUIRED_GATE_IDS = {
    "contract_gate",
    "checkpoint_read_gate",
    "platform_route_smoke",
    "future_implementation_gate",
}
REQUIRED_FUTURE_BLOCKERS = {
    "upper_layer_confirmation_flow",
    "executor_boundary",
    "storage_backend_design",
    "result_materialization_policy",
    "independent_audit_records",
    "negative_regression_suite",
}
REQUIRED_BLOCKED_CLAIMS = {
    "durable session store",
    "durable tool store",
    "long-term memory",
    "real checkpoint storage backend",
    "real tool executor",
    "materialized result reader",
    "automatic replay",
    "durable memory enabled",
    "result materialization enabled",
}


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def require_non_empty_string_list(value: Any, message: str) -> list[str]:
    require(isinstance(value, list) and len(value) > 0, message)
    items: list[str] = []
    for item in value:
        normalized = str(item or "").strip()
        require(bool(normalized), message)
        items.append(normalized)
    return items


def check_fixture(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "promotion gate fixture schema_version must be 1")
    require(document.get("kind") == "session_tooling_promotion_gates", "promotion gate fixture kind mismatch")
    require(document.get("stage") == "P2 Session & Tooling Foundation", "promotion gate stage mismatch")
    require(
        document.get("status") == "contract_and_metadata_smoke_only",
        "promotion gate fixture must not claim implementation status",
    )

    gates = document.get("gates")
    require(isinstance(gates, list) and len(gates) > 0, "promotion gate fixture must include gates")
    gates_by_id: dict[str, dict[str, Any]] = {}
    blocked_claims: set[str] = set()
    for gate in gates:
        require(isinstance(gate, dict), "promotion gate must be an object")
        gate_id = str(gate.get("gate_id") or "").strip()
        require(bool(gate_id), "promotion gate must name gate_id")
        require(gate_id not in gates_by_id, f"duplicate promotion gate_id: {gate_id}")
        gates_by_id[gate_id] = gate

        require_non_empty_string_list(gate.get("scope"), f"{gate_id} must include scope")
        require_non_empty_string_list(gate.get("current_evidence"), f"{gate_id} must include current_evidence")
        require_non_empty_string_list(gate.get("allowed_claims"), f"{gate_id} must include allowed_claims")
        blocked_claims.update(require_non_empty_string_list(gate.get("blocked_claims"), f"{gate_id} must include blocked_claims"))

    missing_gate_ids = sorted(REQUIRED_GATE_IDS - set(gates_by_id))
    require(not missing_gate_ids, f"promotion gate fixture missing gates: {missing_gate_ids}")

    missing_blocked_claims = sorted(REQUIRED_BLOCKED_CLAIMS - blocked_claims)
    require(not missing_blocked_claims, f"promotion gate fixture missing blocked claims: {missing_blocked_claims}")

    future_gate = gates_by_id["future_implementation_gate"]
    future_blockers = set(
        require_non_empty_string_list(
            future_gate.get("required_before_enablement"),
            "future implementation gate must include required_before_enablement",
        )
    )
    missing_future_blockers = sorted(REQUIRED_FUTURE_BLOCKERS - future_blockers)
    require(not missing_future_blockers, f"future gate missing blockers: {missing_future_blockers}")


def check_docs_reference_promotion_gates() -> None:
    session_doc = SESSION_DOC_PATH.read_text(encoding="utf-8")
    tooling_doc = TOOLING_DOC_PATH.read_text(encoding="utf-8")
    contracts_readme = CONTRACTS_README_PATH.read_text(encoding="utf-8")

    for label, content in (
        ("docs/contracts/session.md", session_doc),
        ("docs/contracts/tooling.md", tooling_doc),
    ):
        require("Promotion 门禁分层" in content, f"{label} must include promotion gate section")
        require("Future implementation gate" in content, f"{label} must name future implementation gate")
        require("durable" in content, f"{label} must mention durable boundary")
        require("replay" in content, f"{label} must mention replay boundary")

    require(
        "session-tooling-promotion-gates.json" in contracts_readme,
        "contracts/README.md must reference session-tooling promotion gate fixture",
    )


def main() -> int:
    document = load_json_document(FIXTURE_PATH)
    if not isinstance(document, dict):
        raise SystemExit("promotion gate fixture must be an object")
    check_fixture(document)
    check_docs_reference_promotion_gates()
    print("session/tooling promotion gate checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
