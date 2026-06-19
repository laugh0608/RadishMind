# Workflow Saved Draft Schema Artifact Materialization v1 任务卡

状态：`draft_schema_artifact_materialized_static`

## 任务目标

把 saved workflow draft durable store 的 schema artifact 证据链物化为 committed 静态文件，为后续 repository adapter implementation 提供可复验依赖。

本任务不创建 SQL migration、migration runner、repository interface、repository adapter、数据库连接、OIDC/token validation、production API、publish、run、executor、confirmation、writeback 或 replay。

## 输入

- `Saved Workflow Draft Schema Artifact Evidence v1`
- `Saved Workflow Draft Schema Artifact Manifest v1`
- `Saved Workflow Draft Schema Artifact Materialization Review v1`
- `Saved Workflow Draft Store Selector Implementation v1`
- `Saved Workflow Draft Adapter Smoke Readiness v1`
- `Saved Workflow Draft Repository Adapter Implementation Plan v1`

## 本批输出

- `services/platform/migrations/workflow_saved_drafts/manifest.json`
- `services/platform/migrations/workflow_saved_drafts/ddl-review.md`
- `services/platform/migrations/workflow_saved_drafts/rollback-evidence.json`
- `services/platform/migrations/workflow_saved_drafts/migration-smoke.json`
- `scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json`
- `scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-v1.py`
- `docs/features/workflow/saved-workflow-draft-schema-artifact-materialization-v1.md`

## 验收要求

- `manifest.json` 必须固定 `workflow_saved_draft_schema_artifact_v1`、`workflow_saved_draft_schema_manifest_v1`、`saved_workflow_drafts_store_v1`、logical field mapping、index mapping、operation predicate coverage、failure mapping 和 no side effect counters。
- `ddl-review.md` 必须说明 manual apply gate、no startup auto-migration、runtime role 不具备 migration 权限、destructive change gate、no fallback 和 no side effects。
- `rollback-evidence.json` 必须说明当前无 SQL migration、无数据库 mutation、无 migration runner，并列出 future rollback requirement。
- `migration-smoke.json` 必须覆盖 store schema version、scope predicate、owner list、version conflict、migration failure 和 no side effects。
- 新 checker 必须进入 `./scripts/check-repo.sh --fast`。
- 相邻 readiness / review checker 必须只允许本批静态 schema artifact，不放开 SQL、adapter、OIDC 或 production API。

## 停止线

- 不创建 `.sql` 文件、schema version table、migration runner、数据库连接或 database query。
- 不创建 repository interface、repository adapter、adapter smoke fixture 或 adapter smoke checker。
- 不启用 `repository` store mode。
- 不调用 Radish OIDC、不做 token validation、不创建 production API consumer。
- 不实现 publish、run、executor、confirmation、writeback、replay、resume 或 materialized result reader。
