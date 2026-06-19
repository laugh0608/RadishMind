# Saved Workflow Draft Schema Artifact DDL Review v1

Status: `static_schema_artifact_reviewed_no_sql`

## Review Scope

This artifact reviews the logical durable-store shape for `saved_workflow_draft_record` before any SQL migration or database connection is introduced.

Reviewed artifact:

- manifest: `services/platform/migrations/workflow_saved_drafts/manifest.json`
- schema artifact id: `workflow_saved_draft_schema_artifact_v1`
- store schema version: `saved_workflow_drafts_store_v1`
- schema version table: `workflow_saved_draft_schema_versions`

## Human Review Record

- review decision: `approved_for_static_artifact_only`
- SQL migration created: `false`
- database connection allowed: `false`
- migration runner created: `false`
- repository adapter created: `false`
- production API consumer created: `false`

## Manual Apply Gate

Any future executable migration must be applied manually by an operator-controlled command after a separate SQL migration review. Service startup must not run schema migration automatically, and the application runtime role must not receive migration permissions.

Required future apply inputs:

- reviewed SQL migration artifact
- database backup record
- migration lock acquisition record
- expected schema version before apply
- expected schema version after apply
- rollback command or forward-only exception

## Destructive Change Gate

The current artifact introduces no destructive change because it has no executable DDL. A future migration that drops columns, rewrites payload shape, changes uniqueness, or removes index coverage must carry a separate destructive-change review and backup requirement.

## Predicate Coverage

The reviewed manifest preserves these predicates:

- scope lookup: `tenant_ref + workspace_id + application_id + draft_id`
- owner list: `tenant_ref + workspace_id + application_id + owner_subject_ref`
- version conflict: `tenant_ref + workspace_id + application_id + draft_id + draft_version`
- schema compatibility: `tenant_ref + store_schema_version + schema_version`

## Failure Mapping

Missing or unapplied schema artifact evidence must fail closed with:

- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`
- `draft_store_contract_mismatch`

It must not rewrite:

- `draft_version_conflict`
- `draft_scope_denied`
- `draft_not_found`
- `draft_store_unavailable`

## No Fallback

Schema artifact absence must not fall back to:

- memory dev store
- sample draft
- fixture draft
- dev HTTP route
- test auth

## No Side Effects

The review requires these counters to remain zero:

- `repository_write_count=0`
- `database_write_count=0`
- `migration_apply_count=0`
- `executor_call_count=0`
- `confirmation_call_count=0`
- `business_writeback_count=0`
- `replay_call_count=0`
