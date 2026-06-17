from __future__ import annotations

import json
from pathlib import Path


SCHEMA_MATERIALIZATION_FIXTURE = (
    "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json"
)

SCHEMA_MATERIALIZATION_ALLOWED_FILES = {
    "docs/features/workflow/saved-workflow-draft-schema-artifact-materialization-v1.md",
    "docs/task-cards/workflow-saved-draft-schema-artifact-materialization-v1-plan.md",
    "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json",
    "scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-v1.py",
    "scripts/checks/control_plane/workflow_saved_draft_schema_materialization_guard.py",
    "services/platform/migrations/workflow_saved_drafts",
    "services/platform/migrations/workflow_saved_drafts/manifest.json",
    "services/platform/migrations/workflow_saved_drafts/ddl-review.md",
    "services/platform/migrations/workflow_saved_drafts/rollback-evidence.json",
    "services/platform/migrations/workflow_saved_drafts/migration-smoke.json",
}

SCHEMA_MATERIALIZATION_ALLOWED_LITERALS = {
    "workflow-saved-draft-schema-artifact-materialization-v1.json",
    "check-workflow-saved-draft-schema-artifact-materialization-v1.py",
    "workflow_saved_draft_schema_artifact_v1",
    "workflow_saved_draft_schema_manifest_v1",
    "saved_workflow_drafts_store_v1",
    "services/platform/migrations/workflow_saved_drafts",
    "services/platform/migrations/workflow_saved_drafts/manifest.json",
    "services/platform/migrations/workflow_saved_drafts/ddl-review.md",
    "services/platform/migrations/workflow_saved_drafts/rollback-evidence.json",
    "services/platform/migrations/workflow_saved_drafts/migration-smoke.json",
    "draft_schema_artifact_materialized_static",
}


def schema_materialization_active(repo_root: Path) -> bool:
    fixture_path = repo_root / SCHEMA_MATERIALIZATION_FIXTURE
    if not fixture_path.exists():
        return False
    try:
        fixture = json.loads(fixture_path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return False
    slice_info = fixture.get("slice") if isinstance(fixture, dict) else {}
    return bool(
        isinstance(slice_info, dict)
        and slice_info.get("status") == "draft_schema_artifact_materialized_static"
    )


def schema_materialization_file_allowed(repo_root: Path, relative_path: str) -> bool:
    return schema_materialization_active(repo_root) and relative_path in SCHEMA_MATERIALIZATION_ALLOWED_FILES


def schema_materialization_literal_allowed(repo_root: Path, literal: str) -> bool:
    return schema_materialization_active(repo_root) and literal in SCHEMA_MATERIALIZATION_ALLOWED_LITERALS
