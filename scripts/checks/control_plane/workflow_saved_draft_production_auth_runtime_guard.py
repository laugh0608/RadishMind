from __future__ import annotations

import json
from pathlib import Path


PRODUCTION_AUTH_RUNTIME_FIXTURE = (
    "scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json"
)

PRODUCTION_AUTH_RUNTIME_ALLOWED_FILES = {
    "docs/features/workflow/saved-workflow-draft-production-auth-runtime-v1.md",
    "docs/task-cards/workflow-saved-draft-production-auth-runtime-v1-plan.md",
    "scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json",
    "scripts/checks/control_plane/check-workflow-saved-draft-production-auth-runtime-v1.py",
    "scripts/checks/control_plane/workflow_saved_draft_production_auth_runtime_guard.py",
    "services/platform/internal/httpapi/workflow_saved_draft_production_auth_runtime.go",
    "services/platform/internal/httpapi/workflow_saved_draft_production_auth_runtime_test.go",
}

PRODUCTION_AUTH_RUNTIME_ALLOWED_LITERALS = {
    "workflow-saved-draft-production-auth-runtime-v1",
    "workflow-saved-draft-production-auth-runtime-v1.json",
    "check-workflow-saved-draft-production-auth-runtime-v1.py",
    "workflow_saved_draft_production_auth_runtime",
    "workflow_saved_draft_production_auth_runtime_test",
    "Saved Workflow Draft Production Auth Runtime v1",
    "draft_production_auth_runtime_bridge_implemented",
    "SavedWorkflowDraftVerifiedAuthContext",
    "SavedWorkflowDraftWorkspaceBinding",
    "SavedWorkflowDraftProductionAuthRuntimeResult",
    "BuildSavedWorkflowDraftRepositoryActorContextForOperation",
    "BuildSavedWorkflowDraftRepositoryActorContextFromProductionAuth",
    "TestSavedWorkflowDraftProductionAuthRuntimeBuildsRepositoryActorContext",
    "TestSavedWorkflowDraftProductionAuthRuntimeOperationScopes",
    "TestSavedWorkflowDraftProductionAuthRuntimeFailureMapping",
    "draft_application_scope_denied",
    "draft_owner_scope_denied",
    "draft_audit_context_missing",
}


def production_auth_runtime_active(repo_root: Path) -> bool:
    fixture_path = repo_root / PRODUCTION_AUTH_RUNTIME_FIXTURE
    if not fixture_path.exists():
        return False
    try:
        fixture = json.loads(fixture_path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return False
    slice_info = fixture.get("slice") if isinstance(fixture, dict) else {}
    return bool(
        isinstance(slice_info, dict)
        and slice_info.get("status") == "draft_production_auth_runtime_bridge_implemented"
    )


def production_auth_runtime_file_allowed(repo_root: Path, relative_path: str) -> bool:
    return production_auth_runtime_active(repo_root) and relative_path in PRODUCTION_AUTH_RUNTIME_ALLOWED_FILES


def production_auth_runtime_literal_allowed(repo_root: Path, literal: str) -> bool:
    return production_auth_runtime_active(repo_root) and literal in PRODUCTION_AUTH_RUNTIME_ALLOWED_LITERALS
