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
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-entry-review-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

SLICE_ID = (
    "production-secret-backend-audit-store-storage-adapter-"
    "metadata-contract-artifact-materialization-entry-review-v1"
)
SLICE_STATUS = "audit_store_storage_adapter_metadata_contract_artifact_materialization_entry_review_defined"
ENTRY_DECISION = "metadata_contract_artifact_materialization_task_card_ready_after_entry_review"
NEXT_DEPENDENCY = "storage_adapter_metadata_contract_artifact_materialization_task_card"
RESERVED_CONTRACT_ARTIFACT = "contracts/production-secret-audit-storage-adapter.metadata-contract.json"
FOLLOWUP_MATERIALIZATION_FIXTURE = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.json"
)
FOLLOWUP_MATERIALIZATION_STATUS = "audit_store_storage_adapter_metadata_contract_artifact_materialized"
FOLLOWUP_NEXT_DEPENDENCY = "storage_adapter_backend_product_selection_review"
FOLLOWUP_SELECTION_FIXTURE = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.json"
)
FOLLOWUP_SELECTION_STATUS = "audit_store_storage_adapter_backend_product_selection_review_defined"
FOLLOWUP_SELECTION_NEXT_DEPENDENCY = "storage_adapter_runtime_implementation_entry_refresh_after_product_selection"
FOLLOWUP_AFTER_SELECTION_FIXTURE = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1.json"
)
FOLLOWUP_AFTER_SELECTION_STATUS = (
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined"
)
FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY = "storage_adapter_database_connection_lifecycle_readiness"
FOLLOWUP_ALIGNMENT = {
    "audit_storage_adapter_contract_materialization_task_card_status": "created",
    "audit_storage_adapter_contract_artifact_materialization_status": FOLLOWUP_MATERIALIZATION_STATUS,
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_NEXT_DEPENDENCY,
}
FOLLOWUP_SELECTION_ALIGNMENT = {
    "audit_store_storage_adapter_backend_product_selection_review_status": FOLLOWUP_SELECTION_STATUS,
    "audit_storage_adapter_backend_product_selection_status": "selected_static_product_class_without_backend_provider",
    "audit_storage_adapter_selected_backend_product_class": "managed_database_append_only_table",
    "audit_storage_adapter_selected_backend_product_profile": "reserved_managed_database_append_only_table_profile",
    "audit_storage_adapter_database_product_status": "engine_selected_without_managed_product",
    "audit_storage_adapter_database_connection_provider_status": "not_created",
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_SELECTION_NEXT_DEPENDENCY,
}
FOLLOWUP_AFTER_SELECTION_ALIGNMENT = {
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_status": (
        FOLLOWUP_AFTER_SELECTION_STATUS
    ),
    "audit_storage_adapter_runtime_task_card_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_database_driver_selection_review"
    ),
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY,
    "audit_storage_adapter_database_provider_driver_dsn_tls_role_policy_status": (
        "defined_without_runtime"
    ),
    "audit_storage_adapter_append_only_table_schema_boundary_status": "defined_without_sql_or_runtime",
    "audit_storage_adapter_migration_schema_marker_boundary_status": "logical_schema_marker_handoff_boundary_defined",
    "audit_storage_adapter_offline_adapter_smoke_strategy_status": "required_before_runtime_task_card",
    "audit_storage_adapter_negative_leakage_runtime_scan_boundary_status": "defined_without_runtime",
}
FOLLOWUP_ALLOWED_ARTIFACTS = {
    RESERVED_CONTRACT_ARTIFACT,
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1-plan.md",
}

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-v1.json"
        ),
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-rollback-recovery-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-rollback-recovery-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_offline_validation_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
    ),
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
    "status": SLICE_STATUS,
    "entry_decision": ENTRY_DECISION,
    "previous_runtime_entry_refresh_status": "audit_store_storage_adapter_runtime_implementation_entry_refresh_defined",
    "previous_runtime_task_decision": "storage_adapter_runtime_task_card_still_blocked_after_evidence_readiness",
    "evidence_chain_status": "static_evidence_chain_ready_for_contract_materialization_entry_review",
    "rollback_recovery_evidence_readiness_status": (
        "audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined"
    ),
    "negative_leakage_scan_evidence_readiness_status": (
        "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined"
    ),
    "offline_validation_evidence_readiness_status": (
        "audit_store_storage_adapter_offline_validation_evidence_readiness_defined"
    ),
    "retention_redaction_policy_evidence_readiness_status": (
        "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined"
    ),
    "append_only_semantics_evidence_readiness_status": (
        "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined"
    ),
    "metadata_contract_artifact_readiness_status": (
        "audit_store_storage_adapter_metadata_contract_artifact_readiness_defined"
    ),
    "backend_product_evidence_readiness_status": "audit_store_storage_adapter_backend_product_evidence_readiness_defined",
    "reserved_contract_artifact_path": RESERVED_CONTRACT_ARTIFACT,
    "materialization_task_card_decision": ENTRY_DECISION,
    "materialization_task_card_status": "not_created",
    "contract_artifact_materialization_status": "not_created",
    "backend_product_selection_status": "not_selected",
    "next_dependency": NEXT_DEPENDENCY,
    "storage_adapter_runtime_task_card_status": "not_created",
    "storage_adapter_runtime_status": "not_created",
    "storage_adapter_client_status": "not_created",
    "database_connection_provider_status": "blocked",
    "database_driver_status": "not_selected",
    "database_connection_status": "not_created",
    "sql_migration_status": "not_created",
    "schema_marker_status": "not_created",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "production_resolver_runtime_status": "not_created",
    "production_secret_backend_status": "not_satisfied",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "contract_artifact_materialization_task_card_created_in_this_slice",
    "contract_artifact_materialized_in_this_slice",
    "backend_product_selected_in_this_slice",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "storage_adapter_client_created_in_this_slice",
    "database_connection_provider_enabled",
    "database_connection_created_in_this_slice",
    "sql_migration_created_in_this_slice",
    "schema_marker_created_in_this_slice",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_runtime_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_TASK_CARD_REQUIREMENTS = {
    "metadata-only contract schema version pin",
    "input result envelope validation",
    "record identity validation",
    "failure taxonomy validation",
    "positive contract fixture",
    "forbidden field negative fixture",
    "writer compatibility smoke",
    "no secret material scan",
    "artifact guard for no runtime no backend selection",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_metadata_contract_materialization_entry_dependency_missing",
    "audit_store_storage_adapter_metadata_contract_materialization_task_card_not_created",
    "audit_store_storage_adapter_metadata_contract_materialization_artifact_forbidden",
    "audit_store_storage_adapter_metadata_contract_materialization_backend_product_forbidden",
    "audit_store_storage_adapter_metadata_contract_materialization_runtime_forbidden",
    "audit_store_storage_adapter_metadata_contract_materialization_secret_material_detected",
    "audit_store_storage_adapter_metadata_contract_materialization_fallback_detected",
}

EXPECTED_ALLOWED_ARTIFACTS = {
    (
        "docs/platform/production-secret-backend-audit-store-storage-adapter-"
        "metadata-contract-artifact-materialization-entry-review-v1.md"
    ),
    (
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-"
        "metadata-contract-artifact-materialization-entry-review-v1-plan.md"
    ),
    (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-"
        "metadata-contract-artifact-materialization-entry-review-v1.json"
    ),
    (
        "scripts/check-production-ops-secret-backend-audit-store-storage-adapter-"
        "metadata-contract-artifact-materialization-entry-review-v1.py"
    ),
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


def followup_materialization_exists() -> bool:
    if not FOLLOWUP_MATERIALIZATION_FIXTURE.exists():
        return False
    followup = load_json(FOLLOWUP_MATERIALIZATION_FIXTURE)
    return source_status(followup) == FOLLOWUP_MATERIALIZATION_STATUS


def followup_selection_exists() -> bool:
    if not FOLLOWUP_SELECTION_FIXTURE.exists():
        return False
    followup = load_json(FOLLOWUP_SELECTION_FIXTURE)
    return source_status(followup) == FOLLOWUP_SELECTION_STATUS


def followup_after_selection_exists() -> bool:
    if not FOLLOWUP_AFTER_SELECTION_FIXTURE.exists():
        return False
    followup = load_json(FOLLOWUP_AFTER_SELECTION_FIXTURE)
    return source_status(followup) == FOLLOWUP_AFTER_SELECTION_STATUS


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_storage_adapter_metadata_contract_artifact_materialization_entry_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == SLICE_STATUS, "unexpected status")
    require(slice_info.get("entry_decision") == ENTRY_DECISION, "unexpected entry decision")
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "contract_artifact_materialization_task_card_created",
        "contract_artifact_materialized",
        "backend_product_selected",
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


def assert_entry_review_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_review_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY.items():
        require(boundary.get(field) == expected, f"entry_review_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"entry_review_boundary.{field} must stay false")


def assert_future_requirements_and_blockers(fixture: dict[str, Any]) -> None:
    require(
        set(fixture.get("future_materialization_task_card_requirements") or []) == EXPECTED_TASK_CARD_REQUIREMENTS,
        "future materialization task card requirements drifted",
    )
    blockers = rows_by_id(fixture, "remaining_blockers", "id")
    require(
        set(blockers)
        == {
            "contract_artifact_materialization_task_card",
            "contract_artifact_materialization",
            "backend_product_selection",
            "peer_runtime_dependencies",
        },
        "remaining blocker ids drifted",
    )
    if followup_materialization_exists():
        require(
            blockers["contract_artifact_materialization_task_card"].get("status") in {"not_created", "created"},
            "task card status drifted after follow-up",
        )
        require(
            blockers["contract_artifact_materialization"].get("status")
            in {"not_created", FOLLOWUP_MATERIALIZATION_STATUS},
            "contract materialization status drifted after follow-up",
        )
    else:
        require(
            blockers["contract_artifact_materialization_task_card"].get("status") == "not_created",
            "task card created",
        )
        require(blockers["contract_artifact_materialization"].get("status") == "not_created", "contract materialized")
    require(blockers["backend_product_selection"].get("status") == "not_selected", "backend product selected")
    require(
        blockers["contract_artifact_materialization_task_card"].get("next_dependency") == NEXT_DEPENDENCY,
        "next dependency drifted",
    )


def assert_diagnostics_failures_and_policies(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("diagnostic_envelope") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    sample = diagnostics.get("sample") or {}
    require(set(sample) <= allowed, "diagnostic sample contains non-allowlisted fields")
    require(not (allowed & forbidden), "diagnostic allowlist intersects forbidden fields")
    require(sample.get("materialization_task_card_decision") == ENTRY_DECISION, "diagnostic decision drifted")
    require(sample.get("materialization_task_card_status") == "not_created", "diagnostic created task card")
    require(sample.get("contract_artifact_materialization_status") == "not_created", "diagnostic created contract")
    require(sample.get("backend_product_selection_status") == "not_selected", "diagnostic selected backend")
    require(sample.get("storage_adapter_runtime_status") == "not_created", "diagnostic created runtime")

    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} boundary missing")
        require(item.get("sanitized_diagnostic"), f"{code} diagnostic missing")

    no_fallback = fixture.get("no_fallback_policy") or {}
    require(no_fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    for source in {
        "backend_product_evidence_readiness",
        "metadata_contract_artifact_readiness",
        "append_only_semantics_evidence",
        "retention_redaction_policy_evidence",
        "offline_validation_evidence",
        "negative_leakage_scan_evidence",
        "rollback_recovery_evidence",
        "runtime_entry_refresh",
        "historical_smoke",
        "previous_checker_success",
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
        "contract_artifact_materialization_task_card",
        "metadata_contract_artifact",
        "backend_product_selection_artifact",
        "storage_adapter_runtime_implementation_task_card",
        "storage_adapter_runtime",
        "storage_adapter_client",
        "database_connection_provider",
        "database_connection",
        "db_driver",
        "dsn_parser",
        "sql_migration",
        "schema_marker",
        "audit_store_runtime_implementation_task_card",
        "audit_store_runtime",
        "production_resolver_runtime",
        "repository_mode_runtime",
        "public_production_api",
    }:
        require(artifact in forbidden, f"forbidden artifact missing: {artifact}")
    for path in guard.get("files_must_not_exist") or []:
        if followup_materialization_exists() and str(path) in FOLLOWUP_ALLOWED_ARTIFACTS:
            continue
        require(not (REPO_ROOT / str(path)).exists(), f"forbidden artifact exists: {path}")


def assert_blocker_matrix_alignment() -> None:
    matrix = load_json(BLOCKER_MATRIX_PATH)
    boundary = matrix.get("matrix_boundary") or {}
    require(
        boundary.get("storage_adapter_metadata_contract_artifact_materialization_entry_review_status")
        == SLICE_STATUS,
        "matrix boundary missing materialization entry review status",
    )
    require(
        boundary.get("storage_adapter_contract_artifact_materialization_task_card_decision") == ENTRY_DECISION,
        "matrix boundary materialization task card decision drifted",
    )
    if followup_materialization_exists():
        require(
            boundary.get("storage_adapter_contract_artifact_materialization_task_card_status") == "created",
            "matrix boundary materialization task card status drifted after follow-up",
        )
        require(
            boundary.get("storage_adapter_contract_artifact_materialization_status") == FOLLOWUP_MATERIALIZATION_STATUS,
            "matrix boundary materialization status drifted after follow-up",
        )
        if followup_selection_exists():
            expected_next_dependency = (
                FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY
                if followup_after_selection_exists()
                else FOLLOWUP_SELECTION_NEXT_DEPENDENCY
            )
            require(
                boundary.get("storage_adapter_current_next_dependency") == expected_next_dependency,
                "matrix next dependency drifted after selection follow-up",
            )
        else:
            require(
                boundary.get("storage_adapter_current_next_dependency") == FOLLOWUP_NEXT_DEPENDENCY,
                "matrix next dependency drifted after follow-up",
            )
    else:
        require(
            boundary.get("storage_adapter_contract_artifact_materialization_task_card_status") == "not_created",
            "matrix boundary created materialization task card",
        )
        require(
            boundary.get("storage_adapter_contract_artifact_materialization_status") == "not_created",
            "matrix boundary materialized contract",
        )
        require(boundary.get("storage_adapter_current_next_dependency") == NEXT_DEPENDENCY, "matrix next dependency drifted")
    require(boundary.get("storage_adapter_runtime_status") == "not_created", "matrix created runtime")
    if followup_selection_exists():
        require(
            boundary.get("storage_adapter_backend_product_selection_review_status") == FOLLOWUP_SELECTION_STATUS,
            "matrix selection review status drifted after follow-up",
        )
        require(
            boundary.get("storage_adapter_backend_product_selection_status")
            == "selected_static_product_class_without_backend_provider",
            "matrix selected product class status drifted after follow-up",
        )
        require(
            boundary.get("storage_adapter_selected_backend_product_class") == "managed_database_append_only_table",
            "matrix selected product class drifted after follow-up",
        )
        require(
            boundary.get("storage_adapter_selected_backend_product_profile")
            == "reserved_managed_database_append_only_table_profile",
            "matrix selected product profile drifted after follow-up",
        )
        if followup_after_selection_exists():
            require(
                boundary.get("storage_adapter_runtime_implementation_entry_refresh_after_product_selection_status")
                == FOLLOWUP_AFTER_SELECTION_STATUS,
                "matrix after-selection refresh status drifted",
            )
    else:
        require(boundary.get("storage_adapter_backend_product_selection_status") == "not_selected", "matrix selected product")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    followup_exists = followup_materialization_exists()
    for field, expected in (fixture.get("implementation_readiness_alignment") or {}).items():
        if followup_exists and field in FOLLOWUP_ALIGNMENT:
            expected = FOLLOWUP_ALIGNMENT[field]
        if followup_selection_exists() and field in FOLLOWUP_SELECTION_ALIGNMENT:
            expected = FOLLOWUP_SELECTION_ALIGNMENT[field]
        if followup_after_selection_exists() and field in FOLLOWUP_AFTER_SELECTION_ALIGNMENT:
            expected = FOLLOWUP_AFTER_SELECTION_ALIGNMENT[field]
        require(target.get(field) == expected, f"implementation readiness {field} drifted")

    planned = {str(row.get("id")): row for row in readiness.get("planned_slices") or [] if isinstance(row, dict)}
    item = planned.get("audit-store-storage-adapter-metadata-contract-artifact-materialization-entry-review") or {}
    require(item.get("status") == SLICE_STATUS, "implementation readiness missing materialization entry review slice")
    require(EXPECTED_ALLOWED_ARTIFACTS <= set(item.get("evidence") or []), "planned slice evidence drifted")


def assert_docs_and_registration() -> None:
    docs = {
        (
            "docs/platform/production-secret-backend-audit-store-storage-adapter-"
            "metadata-contract-artifact-materialization-entry-review-v1.md"
        ): [
            SLICE_STATUS,
            ENTRY_DECISION,
            NEXT_DEPENDENCY,
            "not_created",
            "not_selected",
        ],
        (
            "docs/task-cards/production-secret-backend-audit-store-storage-adapter-"
            "metadata-contract-artifact-materialization-entry-review-v1-plan.md"
        ): [
            SLICE_STATUS,
            ENTRY_DECISION,
            NEXT_DEPENDENCY,
            "停止线",
        ],
        "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": [
            SLICE_STATUS,
            ENTRY_DECISION,
            NEXT_DEPENDENCY,
        ],
        "docs/platform/production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md": [
            SLICE_STATUS,
            ENTRY_DECISION,
            NEXT_DEPENDENCY,
        ],
        "docs/platform/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Metadata Contract Artifact Materialization Entry Review v1",
            SLICE_STATUS,
        ],
        "docs/features/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Metadata Contract Artifact Materialization Entry Review v1",
            SLICE_STATUS,
        ],
        "docs/features/workflow/README.md": [SLICE_STATUS],
        "docs/features/workflow/saved-workflow-draft-v1.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/radishmind-current-focus.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/task-cards/README.md": [SLICE_ID, SLICE_STATUS],
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [SLICE_ID, SLICE_STATUS],
        "scripts/README.md": [
            "check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-entry-review-v1.py",
            "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-entry-review-v1.json",
        ],
        "docs/devlogs/2026-W27.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    }
    for path, literals in docs.items():
        text = read(path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous = "check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-v1.py"
    current = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-"
        "metadata-contract-artifact-materialization-entry-review-v1.py"
    )
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
    require(not found, f"materialization entry review contains forbidden literal: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "secret-looking sk token found")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "dsn-like credential found")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_entry_review_boundary(fixture)
    assert_future_requirements_and_blockers(fixture)
    assert_diagnostics_failures_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_blocker_matrix_alignment()
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print(
        "production ops secret backend audit store storage adapter metadata contract "
        "artifact materialization entry review checks passed."
    )


if __name__ == "__main__":
    main()
