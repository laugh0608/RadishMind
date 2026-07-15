#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-model-v1.json"
DEPENDENCY_FIXTURES = {
    "product-surface-v1-boundary": REPO_ROOT / "scripts/checks/fixtures/product-surface-v1-boundary.json",
    "control-plane-data-boundary": REPO_ROOT / "scripts/checks/fixtures/control-plane-data-boundary.json",
    "radish-oidc-client-preconditions": REPO_ROOT / "scripts/checks/fixtures/radish-oidc-client-preconditions.json",
    "gateway-api-key-quota-readiness": REPO_ROOT / "scripts/checks/fixtures/gateway-api-key-quota-readiness.json",
    "workflow-definition-run-record-boundary": REPO_ROOT / "scripts/checks/fixtures/workflow-definition-run-record-boundary.json",
}
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_READ_MODEL_IDS = {
    "tenant-summary",
    "application-summary",
    "api-key-summary",
    "quota-summary",
    "workflow-definition-summary",
    "run-record-summary",
    "audit-summary",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "control_plane_api_ready",
    "database_schema_ready",
    "database_query_ready",
    "user_workspace_ui_ready",
    "production_admin_console_ready",
    "radish_oidc_client_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "workflow_executor_ready",
    "durable_run_store_ready",
    "production_ready",
}
EXPECTED_FORBIDDEN_SCOPE = {
    "control plane API route",
    "database schema",
    "database migration",
    "database query implementation",
    "Radish OIDC integration",
    "tenant or user CRUD",
    "API key generation",
    "quota enforcement",
    "rate limiting implementation",
    "billing ledger implementation",
    "workflow builder",
    "workflow executor",
    "confirmation flow wiring",
    "business writeback",
    "replay execution",
    "production admin console",
    "user workspace UI",
}
EXPECTED_FORBIDDEN_OUTPUTS = {
    "raw_secret_value",
    "api_key_value",
    "api_key_hash",
    "authorization_header",
    "bearer_token",
    "cookie_value",
    "raw_request_body_dump",
    "raw_tool_payload",
    "business_writeback_payload",
    "full_prompt_dump_with_secret",
}
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "control-plane-read-model-v1",
        "read model",
        "不创建数据库 schema",
    ],
    "docs/radishmind-product-scope.md": [
        "control-plane-read-model-v1",
        "tenant summary",
        "workflow definition summary",
    ],
    "docs/radishmind-architecture.md": [
        "control-plane-read-model-v1",
        "read model",
        "tenant-scoped",
    ],
    "docs/radishmind-roadmap.md": [
        "control-plane-read-model-v1",
        "control-plane-read-model-v1.json",
        "check-control-plane-read-model-v1.py",
    ],
    "docs/radishmind-capability-matrix.md": [
        "control-plane-read-model-v1",
        "control-plane-read-model-v1.json",
        "check-control-plane-read-model-v1.py",
    ],
    "docs/task-cards/README.md": [
        "control-plane-read-model-v1",
        "Control Plane Read Model",
    ],
    "docs/task-cards/control-plane-read-model-v1-plan.md": [
        "control-plane-read-model-v1",
        "tenant summary",
        "workflow definition summary",
    ],
    "scripts/README.md": [
        "check-control-plane-read-model-v1.py",
        "control-plane-read-model-v1.json",
        "tenant summary",
    ],
    "docs/devlogs/2026-W22.md": [
        "control-plane-read-model-v1",
        "control-plane-read-model-v1.json",
        "check-control-plane-read-model-v1.py",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    require(set(DEPENDENCY_FIXTURES).issubset(declared), "fixture must declare all dependency slices")
    for expected_id, path in DEPENDENCY_FIXTURES.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(
            slice_info.get("id") == expected_id
            and slice_info.get("status") == "governance_boundary_satisfied",
            f"dependency {expected_id} must remain satisfied",
        )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "control_plane_read_model_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-model-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "control plane read model must only satisfy a governance boundary",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_read_models(fixture: dict[str, Any]) -> None:
    read_models = {
        str(item.get("id") or ""): item
        for item in fixture.get("read_models") or []
        if isinstance(item, dict)
    }
    require(set(read_models) == EXPECTED_READ_MODEL_IDS, f"unexpected read models: {sorted(read_models)}")

    for model_id, model in read_models.items():
        require(model.get("status") == "defined_not_implemented", f"{model_id} must remain defined_not_implemented")
        require(str(model.get("owner_surface") or "").strip(), f"{model_id}.owner_surface is required")
        require(str(model.get("boundary") or "").strip(), f"{model_id}.boundary is required")
        visible_to = model.get("visible_to")
        require(isinstance(visible_to, list) and visible_to, f"{model_id}.visible_to must be non-empty")
        required_fields = model.get("required_fields")
        forbidden_fields = model.get("forbidden_fields")
        require(isinstance(required_fields, list) and required_fields, f"{model_id}.required_fields must be non-empty")
        require(isinstance(forbidden_fields, list) and forbidden_fields, f"{model_id}.forbidden_fields must be non-empty")

    require(
        {"tenant_ref", "quota_summary_ref", "audit_summary_ref"}.issubset(
            set(read_models["tenant-summary"].get("required_fields") or [])
        ),
        "tenant summary must carry quota and audit refs",
    )
    require(
        {"application_ref", "application_kind", "latest_workflow_definition_ref", "last_run_status"}.issubset(
            set(read_models["application-summary"].get("required_fields") or [])
        ),
        "application summary must connect application, workflow and run state",
    )
    require(
        "api_key_value" in set(read_models["api-key-summary"].get("forbidden_fields") or []),
        "API key summary must forbid raw key values",
    )
    require(
        {"request_limit", "token_limit", "cost_limit", "usage_snapshot"}.issubset(
            set(read_models["quota-summary"].get("required_fields") or [])
        ),
        "quota summary must define limit and usage fields",
    )
    require(
        {"workflow_definition_id", "node_count", "risk_level"}.issubset(
            set(read_models["workflow-definition-summary"].get("required_fields") or [])
        ),
        "workflow definition summary must define graph metadata",
    )
    require(
        {"run_id", "status", "failure_code", "cost_summary", "trace_id"}.issubset(
            set(read_models["run-record-summary"].get("required_fields") or [])
        ),
        "run record summary must carry state, failure, cost and trace",
    )
    require(
        {"audit_ref", "actor_subject_ref", "event_kind", "decision"}.issubset(
            set(read_models["audit-summary"].get("required_fields") or [])
        ),
        "audit summary must define actor, event and decision",
    )


def assert_policy_and_sanitization(fixture: dict[str, Any]) -> None:
    access_policy = fixture.get("access_policy") or {}
    require(access_policy.get("default_scope") == "tenant_scoped", "read models must default to tenant scoped")
    require(
        {"identity_context", "tenant_binding", "subject_binding", "scope_grant"}.issubset(
            set(access_policy.get("fail_closed_when_missing") or [])
        ),
        "access policy must fail closed on missing identity, tenant, subject or scope",
    )
    require(
        {"tenant:read", "applications:read", "api_keys:read", "usage:read", "runs:read", "audit:read"}.issubset(
            set(access_policy.get("required_scopes") or [])
        ),
        "access policy must define read scopes",
    )
    require(
        {"cross_tenant_read", "anonymous_read", "read_model_as_write_api"}.issubset(
            set(access_policy.get("forbidden_access") or [])
        ),
        "access policy must forbid cross tenant, anonymous and write API use",
    )

    sanitization_policy = fixture.get("sanitization_policy") or {}
    allowed_outputs = set(sanitization_policy.get("allowed_outputs") or [])
    require("redacted_secret_ref" in allowed_outputs, "sanitization policy must allow only redacted secret refs")
    require("audit_ref" in allowed_outputs, "sanitization policy must allow audit refs")
    missing_forbidden_outputs = sorted(EXPECTED_FORBIDDEN_OUTPUTS - set(sanitization_policy.get("forbidden_outputs") or []))
    require(not missing_forbidden_outputs, f"missing forbidden outputs: {missing_forbidden_outputs}")

    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")


def assert_evidence_and_docs(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-model-v1.py", [])' in check_repo,
        "check-repo.py must run control plane read model check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_read_models(fixture)
    assert_policy_and_sanitization(fixture)
    assert_evidence_and_docs(fixture)
    print("control plane read model v1 checks passed.")


if __name__ == "__main__":
    main()
