#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-storage-backend-design.json"
PRECONDITIONS_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-implementation-preconditions.json"
INDEPENDENT_AUDIT_DESIGN_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-independent-audit-records-design.json"
RESULT_MATERIALIZATION_POLICY_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-result-materialization-policy-design.json"
EXECUTOR_BOUNDARY_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-executor-boundary-design.json"
NEGATIVE_SKELETON_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-negative-regression-skeleton.json"
DENIED_QUERY_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/session-recovery-checkpoint-read-denied-queries.json"
TASK_CARD_PATH = REPO_ROOT / "docs/task-cards/session-tooling-storage-backend-design.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_STORES = {"session_store", "checkpoint_store", "audit_store", "result_store"}
REQUIRED_POLICY_PREREQS = {
    "retention_policy",
    "redaction_policy",
    "secret_handling_policy",
    "encryption_at_rest_policy",
    "access_control_policy",
    "deletion_policy",
}
REQUIRED_FORBIDDEN_WRITES = {
    "durable_session_store_write",
    "durable_checkpoint_store_write",
    "durable_audit_store_write",
    "durable_result_store_write",
    "long_term_memory_write",
    "business_truth_write",
    "replay_state_write",
}
REQUIRED_FAILURE_CASES = {
    "materialized-result-read-denied": "CHECKPOINT_MATERIALIZED_RESULTS_DISABLED",
    "durable-memory-write-denied": "CHECKPOINT_DURABLE_MEMORY_DISABLED",
    "business-truth-write-denied": "BUSINESS_TRUTH_WRITE_DISABLED",
}
REQUIRED_BOUNDARY_AREAS = {"session", "checkpoint", "audit", "result", "executor"}
REQUIRED_NOT_ENABLED = {
    "automatic_replay",
    "business_truth_write",
    "confirmed_action_execution",
    "durable_audit_store",
    "durable_checkpoint_store",
    "durable_result_store",
    "durable_session_store",
    "durable_tool_store",
    "long_term_memory",
    "materialized_result_reader",
    "real_tool_executor",
    "result_ref_reader",
    "replay_state_store",
}
REQUIRED_UNSATISFIED = {"upper_layer_confirmation_flow", "negative_regression_suite"}


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def relative_path(path: Path) -> str:
    return path.relative_to(REPO_ROOT).as_posix()


def require_object(value: Any, message: str) -> dict[str, Any]:
    require(isinstance(value, dict), message)
    return value


def require_string_list(value: Any, message: str) -> list[str]:
    require(isinstance(value, list) and value, message)
    items: list[str] = []
    for item in value:
        normalized = str(item or "").strip()
        require(bool(normalized), message)
        items.append(normalized)
    return items


def check_fixture(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "storage backend design schema_version must be 1")
    require(document.get("kind") == "session_tooling_storage_backend_design", "storage backend design kind mismatch")
    require(document.get("stage") == "P2 Session & Tooling Foundation", "storage backend stage mismatch")
    require(document.get("status") == "design_only_not_connected", "storage backend must stay design-only")
    require(document.get("source_preconditions") == relative_path(PRECONDITIONS_PATH), "storage backend must reference preconditions")
    require(
        document.get("source_independent_audit_records_design") == relative_path(INDEPENDENT_AUDIT_DESIGN_PATH),
        "storage backend must reference independent audit design",
    )
    require(
        document.get("source_result_materialization_policy_design") == relative_path(RESULT_MATERIALIZATION_POLICY_PATH),
        "storage backend must reference result materialization policy",
    )
    require(
        document.get("source_executor_boundary_design") == relative_path(EXECUTOR_BOUNDARY_PATH),
        "storage backend must reference executor boundary",
    )
    require(
        document.get("source_negative_regression_skeleton") == relative_path(NEGATIVE_SKELETON_PATH),
        "storage backend must reference negative skeleton",
    )
    require(
        document.get("source_denied_query_fixture") == relative_path(DENIED_QUERY_FIXTURE),
        "storage backend must reference denied query fixture",
    )
    require(document.get("task_card") == relative_path(TASK_CARD_PATH), "storage backend must reference task card")

    scope = require_object(document.get("scope"), "storage backend must include scope")
    require(scope.get("area") == "storage_backend_design", "storage backend scope area mismatch")
    for key in (
        "durable_session_store_enabled",
        "durable_checkpoint_store_enabled",
        "durable_audit_store_enabled",
        "durable_result_store_enabled",
        "long_term_memory_enabled",
        "materialized_result_reader_enabled",
        "writes_business_truth",
        "replay_enabled",
    ):
        require(scope.get(key) is False, f"{key} must remain false")

    responsibilities = document.get("storage_responsibilities")
    require(isinstance(responsibilities, list) and responsibilities, "storage responsibilities missing")
    stores_by_id: dict[str, dict[str, Any]] = {}
    for item in responsibilities:
        item_obj = require_object(item, "storage responsibility must be object")
        store = str(item_obj.get("store") or "").strip()
        require(store in REQUIRED_STORES, f"unexpected store responsibility: {store}")
        require(store not in stores_by_id, f"duplicate store responsibility: {store}")
        stores_by_id[store] = item_obj
        require(item_obj.get("current_status") == "future_only_disabled", f"{store} must stay future-only disabled")
        require(str(item_obj.get("responsibility") or "").strip(), f"{store} must include responsibility")
        forbidden = set(require_string_list(item_obj.get("forbidden_current_data"), f"{store} forbidden data missing"))
        require(
            any(marker in forbidden for marker in ("raw_tool_output", "credentials", "business_truth_state", "materialized_results", "result_ref")),
            f"{store} must forbid sensitive, materialized, or business-truth data",
        )
    missing_stores = sorted(REQUIRED_STORES - set(stores_by_id))
    require(not missing_stores, f"storage backend missing stores: {missing_stores}")
    require(
        "result_ref" in set(require_string_list(stores_by_id["result_store"].get("forbidden_current_data"), "result store forbidden data missing")),
        "result store must forbid result_ref in current scope",
    )

    policy = require_object(document.get("retention_redaction_secret_policy"), "retention/redaction/secret policy missing")
    require(policy.get("current_status") == "not_defined", "retention/redaction/secret policy must remain not_defined")
    prereqs = set(require_string_list(policy.get("required_before_enablement"), "storage policy prereqs missing"))
    missing_prereqs = sorted(REQUIRED_POLICY_PREREQS - prereqs)
    require(not missing_prereqs, f"storage policy missing prerequisites: {missing_prereqs}")
    forbidden_scope = set(require_string_list(policy.get("current_forbidden_scope"), "storage policy forbidden scope missing"))
    for marker in ("store_credentials", "store_raw_tool_output", "store_business_truth_state"):
        require(marker in forbidden_scope, f"storage policy must forbid {marker}")

    writes = require_object(document.get("write_boundaries"), "write boundaries missing")
    allowed_writes = set(require_string_list(writes.get("allowed_current_writes"), "allowed current writes missing"))
    require("committed_fixture_files" in allowed_writes, "storage design must allow committed fixture files only as design artifacts")
    forbidden_writes = set(require_string_list(writes.get("forbidden_current_writes"), "forbidden current writes missing"))
    missing_writes = sorted(REQUIRED_FORBIDDEN_WRITES - forbidden_writes)
    require(not missing_writes, f"storage backend missing forbidden writes: {missing_writes}")
    business_truth_sources = set(require_string_list(writes.get("business_truth_sources"), "business truth sources missing"))
    require({"RadishFlow", "Radish", "RadishCatalyst"}.issubset(business_truth_sources), "business truth sources mismatch")

    failures = document.get("failure_boundaries")
    require(isinstance(failures, list) and failures, "failure boundaries missing")
    failures_by_id: dict[str, dict[str, Any]] = {}
    for failure in failures:
        failure_obj = require_object(failure, "failure boundary must be object")
        case_id = str(failure_obj.get("case_id") or "").strip()
        require(case_id in REQUIRED_FAILURE_CASES, f"unexpected storage failure case: {case_id}")
        require(case_id not in failures_by_id, f"duplicate storage failure case: {case_id}")
        failures_by_id[case_id] = failure_obj
        require(failure_obj.get("expected_error_code") == REQUIRED_FAILURE_CASES[case_id], f"{case_id} error code mismatch")
        forbidden_outputs = set(require_string_list(failure_obj.get("forbidden_outputs"), f"{case_id} forbidden outputs missing"))
        require(
            any("result" in item or "memory" in item or "business_truth" in item for item in forbidden_outputs),
            f"{case_id} must forbid result, memory, or business truth outputs",
        )
    missing_cases = sorted(set(REQUIRED_FAILURE_CASES) - set(failures_by_id))
    require(not missing_cases, f"storage backend missing failure cases: {missing_cases}")

    boundary_alignment = require_object(document.get("boundary_alignment"), "boundary alignment missing")
    missing_areas = sorted(REQUIRED_BOUNDARY_AREAS - set(boundary_alignment))
    require(not missing_areas, f"storage backend missing boundary areas: {missing_areas}")
    for area in REQUIRED_BOUNDARY_AREAS:
        statements = require_string_list(boundary_alignment.get(area), f"{area} boundary statements missing")
        joined = "\n".join(statements)
        if area == "session":
            require("no durable session store exists" in joined and "long_term_memory" in joined, "session boundary mismatch")
        if area == "checkpoint":
            require("metadata-only" in joined and "materialized_result_reader" in joined, "checkpoint boundary mismatch")
        if area == "audit":
            require("future_only_disabled" in joined and "raw tool output" in joined, "audit boundary mismatch")
        if area == "result":
            require("durable_result_store_enabled=false" in joined or "future_only_disabled" in joined, "result boundary mismatch")
        if area == "executor":
            require("does not create durable stores" in joined and "real_tool_executor is disabled" in joined, "executor boundary mismatch")

    not_enabled = set(require_string_list(document.get("not_enabled_capabilities"), "not enabled capabilities missing"))
    missing_not_enabled = sorted(REQUIRED_NOT_ENABLED - not_enabled)
    require(not missing_not_enabled, f"storage backend missing blocked capabilities: {missing_not_enabled}")

    precondition_status = require_object(document.get("precondition_status_after_design"), "post-design status missing")
    require(
        precondition_status.get("storage_backend_design") == "design_boundary_defined_not_implemented",
        "storage backend must be design-only, not implemented",
    )
    for key in ("executor_boundary", "result_materialization_policy", "independent_audit_records"):
        require(
            precondition_status.get(key) == "design_boundary_defined_not_implemented",
            f"{key} must remain design-only",
        )
    for key in REQUIRED_UNSATISFIED:
        require(precondition_status.get(key) == "not_satisfied", f"{key} must remain not_satisfied")


def check_preconditions_alignment(preconditions: dict[str, Any]) -> None:
    storage_area = next(
        (
            area
            for area in preconditions.get("areas") or []
            if isinstance(area, dict) and area.get("area") == "storage"
        ),
        None,
    )
    storage = require_object(storage_area, "preconditions must include storage area")
    require(storage.get("status") == "not_ready", "storage precondition must remain not_ready")
    blockers = set(require_string_list(storage.get("required_before_enablement"), "storage blockers missing"))
    for blocker in (
        "storage_backend_design",
        "result_materialization_policy",
        "independent_audit_records",
        "negative_regression_suite",
    ):
        require(blocker in blockers, f"storage preconditions missing blocker: {blocker}")


def check_result_materialization_alignment(result_policy: dict[str, Any]) -> None:
    scope = require_object(result_policy.get("scope"), "result materialization policy must include scope")
    require(scope.get("durable_result_store_enabled") is False, "result policy must keep durable result store disabled")
    require(scope.get("materialized_result_reader_enabled") is False, "result policy must keep materialized reader disabled")


def check_independent_audit_alignment(independent_audit_design: dict[str, Any]) -> None:
    record_shape = require_object(independent_audit_design.get("audit_record_shape"), "independent audit must include record shape")
    storage_policy = require_object(record_shape.get("storage_policy"), "independent audit must include storage policy")
    require(storage_policy.get("durable_store_status") == "not_implemented", "audit store must remain not implemented")


def check_executor_alignment(executor_boundary: dict[str, Any]) -> None:
    boundary_alignment = require_object(executor_boundary.get("boundary_alignment"), "executor boundary must include alignment")
    storage_statements = "\n".join(require_string_list(boundary_alignment.get("storage"), "executor storage alignment missing"))
    require("does not create durable" in storage_statements, "executor boundary must not create durable stores")


def check_negative_skeleton_alignment(document: dict[str, Any], negative_skeleton: dict[str, Any]) -> None:
    expected_errors = {
        str(failure.get("expected_error_code") or "").strip()
        for failure in document.get("failure_boundaries") or []
        if isinstance(failure, dict)
    }
    skeleton_errors: set[str] = set()
    for group in negative_skeleton.get("groups") or []:
        if not isinstance(group, dict) or group.get("precondition_area") != "storage":
            continue
        for case in group.get("cases") or []:
            if isinstance(case, dict):
                skeleton_errors.add(str(case.get("expected_error_code") or "").strip())
    missing = sorted(expected_errors - skeleton_errors)
    require(not missing, f"storage backend missing negative skeleton error codes: {missing}")


def check_denied_query_alignment(document: dict[str, Any], denied_queries: dict[str, Any]) -> None:
    mirrored_categories = {
        str(failure.get("mirrors_denied_query_category") or "").strip()
        for failure in document.get("failure_boundaries") or []
        if isinstance(failure, dict) and str(failure.get("mirrors_denied_query_category") or "").strip()
    }
    denied_categories = {
        str(case.get("category") or "").strip()
        for case in denied_queries.get("cases") or []
        if isinstance(case, dict)
    }
    missing = sorted(mirrored_categories - denied_categories)
    require(not missing, f"storage backend mirrors unknown denied query categories: {missing}")


def check_docs_and_consumers() -> None:
    task_card = TASK_CARD_PATH.read_text(encoding="utf-8")
    readme = TASK_CARDS_README.read_text(encoding="utf-8")
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    fixture_name = FIXTURE_PATH.name
    for marker in (
        fixture_name,
        "design_only_not_connected",
        "Store 职责分离",
        "Retention / Redaction / Secret",
        "写入边界",
        "失败边界",
        "与 Session 的边界",
        "与 Checkpoint 的边界",
        "与 Audit 的边界",
        "与 Result 的边界",
        "不实现 durable session store",
        "不启用 automatic replay",
    ):
        require(marker in task_card, f"storage backend task card missing marker: {marker}")
    require(
        "session-tooling-storage-backend-design.md" in readme,
        "task-cards README must reference storage backend task card",
    )
    require(
        "check-session-tooling-storage-backend-design.py" in check_repo,
        "fast baseline must run storage backend design check",
    )


def main() -> int:
    document = require_object(load_json_document(FIXTURE_PATH), "storage backend fixture must be object")
    preconditions = require_object(load_json_document(PRECONDITIONS_PATH), "preconditions fixture must be object")
    independent_audit_design = require_object(load_json_document(INDEPENDENT_AUDIT_DESIGN_PATH), "independent audit fixture must be object")
    result_policy = require_object(load_json_document(RESULT_MATERIALIZATION_POLICY_PATH), "result materialization fixture must be object")
    executor_boundary = require_object(load_json_document(EXECUTOR_BOUNDARY_PATH), "executor boundary fixture must be object")
    negative_skeleton = require_object(load_json_document(NEGATIVE_SKELETON_PATH), "negative skeleton fixture must be object")
    denied_queries = require_object(load_json_document(DENIED_QUERY_FIXTURE), "denied query fixture must be object")
    check_fixture(document)
    check_preconditions_alignment(preconditions)
    check_result_materialization_alignment(result_policy)
    check_independent_audit_alignment(independent_audit_design)
    check_executor_alignment(executor_boundary)
    check_negative_skeleton_alignment(document, negative_skeleton)
    check_denied_query_alignment(document, denied_queries)
    check_docs_and_consumers()
    print("session/tooling storage backend design checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
