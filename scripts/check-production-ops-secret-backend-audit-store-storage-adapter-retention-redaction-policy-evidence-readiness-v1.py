#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.json"
        ),
        "audit_store_storage_adapter_metadata_contract_artifact_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_backend_product_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.json"
        ),
        "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-audit-store-runtime-blocker-matrix-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json",
        "audit_store_runtime_blocker_matrix_defined",
    ),
    "production-secret-backend-implementation-readiness": (
        "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        "implementation_readiness_defined",
    ),
}

EXPECTED_BOUNDARY = {
    "status": "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
    "readiness_decision": "retention_redaction_policy_evidence_defined_without_runtime",
    "append_only_semantics_evidence_readiness_status": (
        "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined"
    ),
    "append_only_semantics_status": "append_only_semantics_evidence_defined_without_runtime",
    "append_only_operation_status": "append_only_insert_only",
    "forbidden_mutation_policy_status": "update_delete_overwrite_truncate_reject_policy_defined",
    "metadata_contract_artifact_readiness_status": (
        "audit_store_storage_adapter_metadata_contract_artifact_readiness_defined"
    ),
    "metadata_contract_artifact_status": "readiness_defined_without_materialized_artifact",
    "backend_product_evidence_readiness_status": (
        "audit_store_storage_adapter_backend_product_evidence_readiness_defined"
    ),
    "backend_product_selection_status": "not_selected",
    "selected_backend_family": "append_only_metadata_audit_log",
    "selected_reserved_candidate": "reserved_append_only_audit_log",
    "retention_redaction_policy_evidence_status": "retention_redaction_policy_evidence_defined",
    "retention_redaction_status": "retention_redaction_policy_evidence_defined_without_runtime",
    "retention_window_reference_status": "metadata_only_retention_window_reference_defined",
    "redaction_policy_reference_status": "metadata_only_redaction_policy_reference_defined",
    "append_only_compatibility_status": "append_only_immutability_compatible_policy_defined",
    "forbidden_erasure_policy_status": "delete_overwrite_inline_redaction_forbidden",
    "offline_validation_status": "not_created",
    "negative_leakage_scan_status": "not_created",
    "rollback_recovery_status": "required_before_runtime_task_card",
    "next_dependency": "storage_adapter_offline_validation_evidence_readiness",
    "storage_adapter_runtime_task_card_status": "not_created",
    "storage_adapter_runtime_status": "not_created",
    "storage_adapter_client_status": "not_created",
    "retention_executor_status": "not_created",
    "redaction_executor_status": "not_created",
    "database_connection_provider_status": "blocked",
    "database_driver_status": "not_selected",
    "sql_migration_status": "not_created",
    "schema_marker_status": "not_created",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "audit_writer_runtime_status": "not_created",
    "idempotency_runtime_status": "not_created",
    "delivery_runtime_status": "not_created",
    "production_resolver_runtime_status": "not_created",
    "production_secret_backend_status": "not_satisfied",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "backend_product_selected_in_this_slice",
    "contract_artifact_materialized_in_this_slice",
    "retention_runtime_created_in_this_slice",
    "redaction_runtime_created_in_this_slice",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "storage_adapter_client_created_in_this_slice",
    "database_connection_provider_enabled",
    "sql_migration_created_in_this_slice",
    "schema_marker_created_in_this_slice",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_runtime_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "duplicate_detector_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_RETENTION_FIELDS = {
    "retention_policy_ref",
    "retention_window_ref",
    "retention_anchor_ref",
    "retention_decision_ref",
    "append_only_contract_ref",
    "storage_record_identity_ref",
    "policy_version",
    "audit_ref",
}
EXPECTED_REDACTION_FIELDS = {
    "redaction_policy_ref",
    "redaction_decision_ref",
    "redaction_reason_ref",
    "redaction_actor_ref",
    "append_only_contract_ref",
    "storage_record_identity_ref",
    "policy_version",
    "audit_ref",
}
EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_retention_redaction_dependency_missing",
    "audit_store_storage_adapter_retention_window_reference_missing",
    "audit_store_storage_adapter_redaction_reference_missing",
    "audit_store_storage_adapter_retention_redaction_mutation_forbidden",
    "audit_store_storage_adapter_retention_redaction_payload_material_detected",
    "audit_store_storage_adapter_retention_redaction_runtime_scope_overreach",
    "audit_store_storage_adapter_retention_redaction_fallback_detected",
    "audit_store_storage_adapter_retention_redaction_next_dependency_missing",
}
EXPECTED_ALLOWED_ARTIFACTS = {
    (
        "docs/platform/"
        "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.md"
    ),
    (
        "docs/task-cards/"
        "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1-plan.md"
    ),
    (
        "scripts/checks/fixtures/"
        "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.json"
    ),
    "scripts/check-production-ops-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.py",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    path = REPO_ROOT / relative_path
    require(path.exists(), f"required file missing: {relative_path}")
    return path.read_text(encoding="utf-8")


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must contain a JSON object")
    return document


def source_status(document: dict[str, Any]) -> str:
    slice_info = document.get("slice") or {}
    return str(slice_info.get("status") or document.get("status") or "")


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {str(row.get(id_field) or ""): row for row in fixture.get(key) or [] if isinstance(row, dict)}
    require(rows, f"{key} must not be empty")
    return rows


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id")
        == "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status")
        == "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
        "unexpected status",
    )
    require(
        slice_info.get("readiness_decision") == "retention_redaction_policy_evidence_defined_without_runtime",
        "unexpected readiness decision",
    )
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "retention_runtime_implemented",
        "redaction_runtime_implemented",
        "storage_adapter_runtime_task_card_created",
        "storage_adapter_runtime_created",
        "audit_store_runtime_task_card_created",
        "audit_store_runtime_created",
        "production_resolver_runtime_created",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in claims, f"does_not_claim missing {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    require(set(dependencies) == set(EXPECTED_DEPENDENCIES), "dependency ids drifted")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("evidence") == relative_path, f"{dependency_id} evidence path drifted")
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        source = load_json(REPO_ROOT / relative_path)
        require(source_status(source) == expected_status, f"{dependency_id} source status drifted")


def assert_readiness_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("readiness_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY.items():
        require(boundary.get(field) == expected, f"readiness_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"readiness_boundary.{field} must stay false")


def assert_retention_redaction_policy(fixture: dict[str, Any]) -> None:
    retention = fixture.get("retention_policy_evidence") or {}
    require(
        retention.get("status") == "metadata_only_retention_window_reference_defined",
        "retention policy status drifted",
    )
    require(
        set(retention.get("required_reference_fields") or []) == EXPECTED_RETENTION_FIELDS,
        "retention reference fields drifted",
    )
    for mechanism in {
        "physical_ttl_rule",
        "cleanup_job",
        "partition_purge",
        "bucket_lifecycle_rule",
        "topic_retention_rule",
        "table_delete_job",
        "log_compaction",
        "provider_retention_rule",
    }:
        require(mechanism in set(retention.get("forbidden_runtime_mechanisms") or []), f"missing {mechanism}")
    for action in {
        "delete_existing_record",
        "overwrite_existing_record",
        "truncate_log",
        "compact_log",
        "rewrite_sequence",
        "mutate_record_identity",
    }:
        require(action in set(retention.get("forbidden_success_actions") or []), f"missing {action}")

    redaction = fixture.get("redaction_policy_evidence") or {}
    require(
        redaction.get("status") == "metadata_only_redaction_policy_reference_defined",
        "redaction policy status drifted",
    )
    require(
        set(redaction.get("required_reference_fields") or []) == EXPECTED_REDACTION_FIELDS,
        "redaction reference fields drifted",
    )
    for mode in {
        "inline_raw_payload_redaction",
        "overwrite_record_payload",
        "delete_record_payload",
        "mutate_record_identity",
        "write_payload_hash",
        "write_secret_derived_hash",
        "reveal_payload_before_redaction",
    }:
        require(mode in set(redaction.get("forbidden_redaction_modes") or []), f"missing forbidden mode {mode}")
    for artifact in {"redaction_marker_runtime", "redaction_executor", "payload_reader", "payload_hash_writer"}:
        require(artifact in set(redaction.get("does_not_create") or []), f"creates {artifact}")

    compatibility = fixture.get("append_only_compatibility_contract") or {}
    require(
        compatibility.get("status") == "append_only_immutability_compatible_policy_defined",
        "append-only compatibility status drifted",
    )
    for field in {"append_only_sequence_ref", "storage_record_identity_ref", "writer_result_ref", "policy_version"}:
        require(field in set(compatibility.get("must_preserve") or []), f"missing preserve field {field}")
    for claim in {
        "retention_deletes_record",
        "retention_overwrites_record",
        "retention_compacts_log",
        "redaction_overwrites_payload",
        "redaction_deletes_payload",
        "redaction_mutates_identity",
        "policy_rewrites_sequence",
    }:
        require(claim in set(compatibility.get("forbidden_claims") or []), f"missing forbidden claim {claim}")
    require(
        compatibility.get("future_runtime_requires_independent_entry") is True,
        "future runtime must require independent entry",
    )


def assert_diagnostics_failures_and_policies(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("diagnostic_envelope") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    sample = diagnostics.get("sample") or {}
    require(set(sample) <= allowed, "diagnostic sample contains non-allowlisted fields")
    require(not (allowed & forbidden), "diagnostic allowlist intersects forbidden fields")
    require(sample.get("storage_adapter_runtime_status") == "not_created", "diagnostic sample created runtime")
    require(
        sample.get("next_dependency") == "storage_adapter_offline_validation_evidence_readiness",
        "diagnostic sample next dependency drifted",
    )

    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} boundary missing")
        require(item.get("sanitized_diagnostic"), f"{code} diagnostic missing")

    no_fallback = fixture.get("no_fallback_policy") or {}
    require(no_fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    for source in {
        "append_only_semantics_evidence",
        "metadata_contract_artifact_readiness",
        "backend_product_evidence_readiness",
        "storage_adapter_runtime_entry_review",
        "static_delivery_idempotency_readiness",
        "historical_smoke",
    }:
        require(source in set(no_fallback.get("forbidden_sources") or []), f"missing forbidden fallback {source}")

    counters = fixture.get("side_effect_counters") or {}
    require(counters, "side effect counters missing")
    for counter, value in counters.items():
        require(value == 0, f"{counter} must stay 0")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    require(set(guard.get("allowed_added_artifacts") or []) == EXPECTED_ALLOWED_ARTIFACTS, "allowed artifacts drifted")
    for path in EXPECTED_ALLOWED_ARTIFACTS:
        require((REPO_ROOT / path).exists(), f"allowed artifact missing: {path}")
    forbidden = set(guard.get("forbidden_artifact_kinds") or [])
    for artifact in {
        "storage_adapter_runtime_implementation_task_card",
        "storage_adapter_runtime",
        "retention_executor",
        "redaction_executor",
        "database_connection_provider",
        "sql_migration",
        "audit_store_runtime_implementation_task_card",
        "audit_store_runtime",
        "audit_writer_runtime",
        "delivery_runtime",
        "idempotency_runtime",
        "duplicate_detector",
        "production_resolver_runtime",
        "repository_mode_runtime",
        "public_production_api",
    }:
        require(artifact in forbidden, f"forbidden artifact missing: {artifact}")
    for path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(path)).exists(), f"forbidden artifact exists: {path}")


def assert_blocker_matrix_alignment() -> None:
    matrix = load_json(BLOCKER_MATRIX_PATH)
    boundary = matrix.get("matrix_boundary") or {}
    require(
        boundary.get("storage_adapter_retention_redaction_policy_evidence_readiness_status")
        == "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
        "matrix boundary missing retention redaction readiness status",
    )
    require(
        boundary.get("storage_adapter_retention_redaction_status")
        == "retention_redaction_policy_evidence_defined_without_runtime",
        "matrix boundary retention/redaction status drifted",
    )
    require(
        boundary.get("storage_adapter_offline_validation_status") == "not_created",
        "matrix boundary offline validation status drifted",
    )
    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    require(
        durable.get("status") == "retention_redaction_policy_evidence_readiness_defined_task_card_blocked",
        "durable backend blocker status drifted",
    )
    require(
        durable.get("source")
        == "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1",
        "durable backend blocker source drifted",
    )
    require(durable.get("blocks_audit_store_runtime_task_card") is True, "durable backend must block audit runtime")
    require(durable.get("blocks_production_resolver_task_card") is True, "durable backend must block resolver runtime")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    alignment = fixture.get("implementation_readiness_alignment") or {}
    for field, expected in alignment.items():
        if field == "status":
            continue
        require(target.get(field) == expected, f"implementation readiness {field} drifted")

    planned = {str(row.get("id")): row for row in readiness.get("planned_slices") or [] if isinstance(row, dict)}
    item = planned.get("audit-store-storage-adapter-retention-redaction-policy-evidence-readiness") or {}
    require(
        item.get("status") == "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
        "implementation readiness missing retention redaction planned slice",
    )
    require(EXPECTED_ALLOWED_ARTIFACTS <= set(item.get("evidence") or []), "planned slice evidence drifted")


def assert_docs_and_registration() -> None:
    docs = {
        (
            "docs/platform/"
            "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.md"
        ): [
            "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
            "retention_redaction_policy_evidence_defined_without_runtime",
            "storage_adapter_offline_validation_evidence_readiness",
        ],
        (
            "docs/task-cards/"
            "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1-plan.md"
        ): [
            "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
            "metadata_only_retention_window_reference_defined",
            "停止线",
        ],
        "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": [
            "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
            "retention_redaction_policy_evidence_readiness_defined_task_card_blocked",
        ],
        "docs/platform/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Retention / Redaction Policy Evidence Readiness v1",
            "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
        ],
        "docs/features/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Retention / Redaction Policy Evidence Readiness v1",
            "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
        ],
        "docs/features/workflow/README.md": [
            "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
        ],
        "docs/features/workflow/saved-workflow-draft-v1.md": [
            "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
            "storage_adapter_offline_validation_evidence_readiness",
        ],
        "docs/radishmind-current-focus.md": [
            "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
            "storage_adapter_offline_validation_evidence_readiness",
        ],
        "docs/task-cards/README.md": [
            "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1",
            "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
        ],
        "scripts/README.md": [
            "check-production-ops-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.py",
            "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.json",
        ],
        "docs/devlogs/2026-W27.md": [
            "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
        ],
    }
    for path, literals in docs.items():
        text = read(path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous = "check-production-ops-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.py"
    current = "check-production-ops-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.py"
    matrix = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    for script in {previous, current, matrix}:
        require(script in check_repo, f"check-repo.py missing {script}")
    require(check_repo.index(previous) < check_repo.index(current) < check_repo.index(matrix), "check order drifted")


def assert_no_secret_literals() -> None:
    text = "\n".join(
        read(path)
        for path in EXPECTED_ALLOWED_ARTIFACTS
        if path.endswith(".md") or path.endswith(".json")
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "-----BEGIN", "authorization:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"retention redaction readiness contains forbidden literal: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "secret-looking sk token found")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "dsn-like credential found")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_readiness_boundary(fixture)
    assert_retention_redaction_policy(fixture)
    assert_diagnostics_failures_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_blocker_matrix_alignment()
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print("production ops secret backend audit store storage adapter retention redaction policy evidence readiness checks passed.")


if __name__ == "__main__":
    main()
