# Saved Workflow Draft Schema Artifact Materialization v1 专题

更新时间：2026-06-17

## 专题定位

`Saved Workflow Draft Schema Artifact Materialization v1` 承接 [Saved Workflow Draft Schema Artifact Materialization Review v1](saved-workflow-draft-schema-artifact-materialization-review-v1.md)，用于把 saved draft durable store 所需的 schema artifact 证据从 future contract 物化为 committed 静态文件。

本专题只物化 schema artifact 证据链，不创建 SQL migration、schema migration runner、数据库连接、repository interface、repository adapter、Radish OIDC middleware、token validation、production API、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_schema_artifact_materialized_static`

## 物化范围

本批新增以下 committed artifact：

- `services/platform/migrations/workflow_saved_drafts/manifest.json`
- `services/platform/migrations/workflow_saved_drafts/ddl-review.md`
- `services/platform/migrations/workflow_saved_drafts/rollback-evidence.json`
- `services/platform/migrations/workflow_saved_drafts/migration-smoke.json`
- `scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json`
- `scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-v1.py`

这些文件固定 schema artifact id、store schema version、logical entity、field mapping、index mapping、manual apply gate、DDL review、rollback evidence、migration smoke、failure mapping 和 no side effect counters。它们不包含可执行 SQL，不创建 SQL migration，不创建 repository adapter，不连接真实数据库，也不启用 `repository` store mode。

## Artifact Contract

`manifest.json` 固定：

- schema artifact id：`workflow_saved_draft_schema_artifact_v1`
- manifest version：`workflow_saved_draft_schema_manifest_v1`
- store schema version：`saved_workflow_drafts_store_v1`
- logical entity：`saved_workflow_draft_record`
- schema version table 名称：`workflow_saved_draft_schema_versions`
- operation predicate coverage：`SaveWorkflowDraftRecord`、`ReadWorkflowDraftRecord`、`ListWorkflowDraftRecords`

`ddl-review.md` 固定人工 review 结论、manual apply gate、no startup auto-migration、runtime role 不具备 migration 权限、destructive change gate、failure mapping 和 no fallback。

`rollback-evidence.json` 固定当前没有 SQL migration、没有数据库 mutation、没有 migration runner，因此 rollback command 在本批不适用；后续真实 SQL migration 必须独立提供 rollback command 或 forward-only exception。

`migration-smoke.json` 固定静态 smoke 覆盖：store schema version、tenant / workspace / application predicate、owner list projection、version conflict predicate、migration failure mapping 和 no side effect counters。

## Gate Matrix

| gate | 当前状态 | 说明 |
| --- | --- | --- |
| schema artifact evidence consumed | `satisfied` | 已消费 evidence contract |
| schema artifact manifest consumed | `satisfied` | 已消费 manifest shape contract |
| materialization review consumed | `satisfied` | 已按 review 独立打开本批 implementation |
| selector implementation consumed | `satisfied` | 已有 fail-closed selector implementation |
| schema artifact materialized | `satisfied` | 本批新增 manifest、DDL review、rollback evidence 和 migration smoke |
| repository adapter gate | `not_satisfied` | 不创建 interface、adapter 或 database query |
| production auth gate | `not_satisfied` | 不创建 OIDC 或 token validation |
| database execution gate | `not_satisfied` | 不创建 SQL，不运行 migration，不创建 schema version table |
| adapter smoke execution gate | `not_satisfied_in_this_slice` | 本批不创建 adapter smoke fixture 或 checker；后续 `workflow-saved-draft-adapter-smoke-v1` 已独立完成 |

## 失败语义

本批继续固定以下 fail-closed code：

- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`
- `draft_store_contract_mismatch`
- `repository_store_disabled`
- `invalid_draft_store_mode`

`draft_version_conflict`、`draft_scope_denied`、`draft_not_found` 和 `draft_store_unavailable` 保持原语义，不得被 migration failure 覆盖。

## 验收方式

- 新增 `workflow-saved-draft-schema-artifact-materialization-v1` fixture / checker，检查 artifact 文件存在、字段 / index / operation / failure mapping 一致、no fallback 和 no side effects。
- 更新相邻 checker，使旧的 evidence / manifest / readiness / review 只允许本批明确列出的静态 schema artifact。
- 运行专项 checker、相邻 workflow checker、`./scripts/check-repo.sh --fast`。
- 因同步 current focus、能力矩阵和周志，补跑 `./scripts/check-repo.sh`。

## 停止线

- 不创建 SQL migration、schema version table、migration runner、数据库连接、database query、repository interface、repository adapter、adapter smoke fixture、adapter smoke checker、Radish OIDC middleware、token validation 或 production API consumer。
- 不启用 `repository` store mode；selector 仍只允许 `memory_dev` 成功，reserved / unknown mode fail closed。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 `draft_schema_artifact_materialized_static` 解释为 database ready、migration applied、repository adapter ready、repository mode ready、adapter smoke ready、OIDC ready 或 production ready。
