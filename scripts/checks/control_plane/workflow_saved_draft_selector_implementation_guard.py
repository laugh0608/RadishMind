from __future__ import annotations

import json
from pathlib import Path


SELECTOR_IMPLEMENTATION_FIXTURE = "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-v1.json"

SELECTOR_IMPLEMENTATION_ALLOWED_FILES = {
    "docs/features/workflow/saved-workflow-draft-store-selector-implementation-v1.md",
    "docs/task-cards/workflow-saved-draft-store-selector-implementation-v1-plan.md",
    "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-v1.json",
    "scripts/checks/control_plane/check-workflow-saved-draft-store-selector-smoke-v1.py",
    "services/platform/internal/httpapi/workflow_saved_draft_store_selector.go",
    "services/platform/internal/httpapi/workflow_saved_draft_store_selector_test.go",
}

SELECTOR_IMPLEMENTATION_ALLOWED_LITERALS = {
    "RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE",
    "SelectWorkflowSavedDraftStore",
    "WorkflowSavedDraftStoreSelector",
    "workflow_saved_draft_store",
    "workflow_saved_draft_store_mode",
    "workflow_saved_draft_store_selector",
    "workflow_saved_draft_store_selector.go",
    "workflow_saved_draft_store_selector_test.go",
    "workflow-saved-draft-store-selector-smoke-v1.json",
    "check-workflow-saved-draft-store-selector-smoke-v1.py",
}


def selector_implementation_active(repo_root: Path) -> bool:
    fixture_path = repo_root / SELECTOR_IMPLEMENTATION_FIXTURE
    if not fixture_path.exists():
        return False
    try:
        fixture = json.loads(fixture_path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return False
    slice_info = fixture.get("slice") if isinstance(fixture, dict) else {}
    return bool(
        isinstance(slice_info, dict)
        and slice_info.get("status") == "draft_store_selector_smoke_implemented"
    )


def selector_implementation_file_allowed(repo_root: Path, relative_path: str) -> bool:
    return selector_implementation_active(repo_root) and relative_path in SELECTOR_IMPLEMENTATION_ALLOWED_FILES


def selector_implementation_literal_allowed(repo_root: Path, literal: str) -> bool:
    return selector_implementation_active(repo_root) and literal in SELECTOR_IMPLEMENTATION_ALLOWED_LITERALS
