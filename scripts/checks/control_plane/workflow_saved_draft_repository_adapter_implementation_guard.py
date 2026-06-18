from __future__ import annotations

import json
from pathlib import Path


REPOSITORY_ADAPTER_IMPLEMENTATION_FIXTURE = (
    "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-v1.json"
)

REPOSITORY_ADAPTER_IMPLEMENTATION_ALLOWED_FILES = {
    "docs/task-cards/workflow-saved-draft-repository-adapter-implementation-v1-plan.md",
    "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-v1.json",
    "scripts/checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-v1.py",
    "scripts/checks/control_plane/workflow_saved_draft_repository_adapter_implementation_guard.py",
    "services/platform/internal/httpapi/workflow_saved_draft_repository.go",
    "services/platform/internal/httpapi/workflow_saved_draft_repository_adapter.go",
    "services/platform/internal/httpapi/workflow_saved_draft_repository_adapter_test.go",
}

REPOSITORY_ADAPTER_IMPLEMENTATION_ALLOWED_LITERALS = {
    "SavedWorkflowDraftRepository",
    "SavedWorkflowDraftRepositoryAdapter",
    "SavedWorkflowDraftRepositoryQueryExecutor",
    "SavedWorkflowDraftRepositorySchemaPreflight",
    "NewSavedWorkflowDraftRepositoryAdapter",
    "SaveWorkflowDraftRecord",
    "ReadWorkflowDraftRecord",
    "ListWorkflowDraftRecords",
    "type SavedWorkflowDraftRepository interface",
    "func NewSavedWorkflowDraftRepositoryAdapter",
    "workflow_saved_draft_repository.go",
    "workflow_saved_draft_repository_adapter.go",
    "workflow_saved_draft_repository_adapter_test.go",
    "workflow_saved_draft_repository_adapter",
    "workflow-saved-draft-repository-adapter-implementation-v1.json",
    "check-workflow-saved-draft-repository-adapter-implementation-v1.py",
    "workflow_saved_draft_repository_adapter_implementation_guard.py",
    "draft_repository_adapter_implemented",
    "saved_workflow_drafts_store_v1",
}


def repository_adapter_implementation_active(repo_root: Path) -> bool:
    fixture_path = repo_root / REPOSITORY_ADAPTER_IMPLEMENTATION_FIXTURE
    if not fixture_path.exists():
        return False
    try:
        fixture = json.loads(fixture_path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return False
    slice_info = fixture.get("slice") if isinstance(fixture, dict) else {}
    return bool(
        isinstance(slice_info, dict)
        and slice_info.get("status") == "draft_repository_adapter_implemented"
    )


def repository_adapter_implementation_file_allowed(repo_root: Path, relative_path: str) -> bool:
    return (
        repository_adapter_implementation_active(repo_root)
        and relative_path in REPOSITORY_ADAPTER_IMPLEMENTATION_ALLOWED_FILES
    )


def repository_adapter_implementation_literal_allowed(repo_root: Path, literal: str) -> bool:
    return (
        repository_adapter_implementation_active(repo_root)
        and literal in REPOSITORY_ADAPTER_IMPLEMENTATION_ALLOWED_LITERALS
    )
