from __future__ import annotations

import json
from pathlib import Path


ADAPTER_SMOKE_EXECUTION_FIXTURE = "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-v1.json"

ADAPTER_SMOKE_EXECUTION_ALLOWED_FILES = {
    "docs/features/workflow/saved-workflow-draft-adapter-smoke-execution-v1.md",
    "docs/task-cards/workflow-saved-draft-adapter-smoke-v1-plan.md",
    "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-v1.json",
    "scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-v1.py",
    "scripts/checks/control_plane/workflow_saved_draft_adapter_smoke_execution_guard.py",
    "services/platform/internal/httpapi/workflow_saved_draft_repository_adapter_smoke_test.go",
}

ADAPTER_SMOKE_EXECUTION_ALLOWED_LITERALS = {
    "workflow-saved-draft-adapter-smoke-v1",
    "workflow-saved-draft-adapter-smoke-v1.json",
    "check-workflow-saved-draft-adapter-smoke-v1.py",
    "workflow_saved_draft_adapter_smoke",
    "workflow_saved_draft_repository_adapter_smoke_test",
    "Saved Workflow Draft Adapter Smoke Execution v1",
    "draft_adapter_smoke_executed",
    "TestSavedWorkflowDraftRepositoryAdapterSmokeExecutionConsumesStaticCases",
    "TestSavedWorkflowDraftRepositoryAdapterSmokeExecutionFailureMapping",
    "TestSavedWorkflowDraftRepositoryAdapterSmokeFixtureMatchesGoTest",
}


def adapter_smoke_execution_active(repo_root: Path) -> bool:
    fixture_path = repo_root / ADAPTER_SMOKE_EXECUTION_FIXTURE
    if not fixture_path.exists():
        return False
    try:
        fixture = json.loads(fixture_path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return False
    slice_info = fixture.get("slice") if isinstance(fixture, dict) else {}
    return bool(isinstance(slice_info, dict) and slice_info.get("status") == "draft_adapter_smoke_executed")


def adapter_smoke_execution_file_allowed(repo_root: Path, relative_path: str) -> bool:
    return adapter_smoke_execution_active(repo_root) and relative_path in ADAPTER_SMOKE_EXECUTION_ALLOWED_FILES


def adapter_smoke_execution_literal_allowed(repo_root: Path, literal: str) -> bool:
    return adapter_smoke_execution_active(repo_root) and literal in ADAPTER_SMOKE_EXECUTION_ALLOWED_LITERALS
