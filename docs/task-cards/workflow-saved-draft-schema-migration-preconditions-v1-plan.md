# Workflow Saved Draft Schema / Migration Preconditions v1 任务卡

更新时间：2026-06-15

## 任务标识

- 切片：`workflow-saved-draft-schema-migration-preconditions-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`draft_schema_migration_preconditions_defined`

## 目标

在 repository contract preconditions 之后，固定 future saved workflow draft durable store 的 logical schema、index strategy、migration gate、failure mapping、no sample fallback 和 implementation artifact guard。

本任务卡只定义 schema / migration preconditions，不创建数据库 schema、SQL、migration runner、repository interface、repository adapter、store selector、Radish OIDC、production API、publish、run、executor、confirmation、writeback 或 replay。

## 输入事实源

- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)
- [Saved Workflow Draft Durable Store Preconditions v1 专题](../features/workflow/saved-workflow-draft-durable-store-preconditions-v1.md)
- [Saved Workflow Draft Repository Contract Preconditions v1 专题](../features/workflow/saved-workflow-draft-repository-contract-preconditions-v1.md)
- [Saved Workflow Draft Schema / Migration Preconditions v1 专题](../features/workflow/saved-workflow-draft-schema-migration-preconditions-v1.md)
- [User Workspace 设计与开发文档](../features/user-workspace.md)

## 本轮交付

- 新增 schema / migration preconditions 细专题，固定 precondition-only 状态。
- 新增 `workflow-saved-draft-schema-migration-preconditions-v1` fixture / checker。
- checker 接入 fast baseline，校验 dependency、logical schema、index strategy、migration gate、failure mapping、no sample fallback、文档引用和 forbidden implementation artifact。
- 同步更新 workflow 入口、Saved Draft 主专题、durable store / repository preconditions、User Workspace、当前焦点、任务卡索引、脚本说明和周志。

## Logical Schema Contract

future durable store 的逻辑记录为 `saved_workflow_draft_record`，必须覆盖：

- scope identity：`tenant_ref`、`workspace_id`、`application_id`、`draft_id`。
- ownership / actor：`owner_subject_ref`、`created_by_actor_ref`、`updated_by_actor_ref`。
- source / version：`source_definition_id`、`base_definition_version`、`draft_version`、`schema_version`、`store_schema_version`。
- payload / audit：`sanitized_draft_payload`、`validation_summary`、`blocked_capability_summary`、`request_id`、`audit_ref`、`created_at`、`updated_at`。

`schema_version` 只代表 draft payload schema；`store_schema_version` 代表 durable store / migration artifact version，二者不得混用。

## Index / Migration Gate

future schema artifact 至少需要覆盖：

- `saved_drafts_scope_lookup`
- `saved_drafts_owner_list`
- `saved_drafts_version_conflict`
- `saved_drafts_status_list`
- `saved_drafts_schema_version`

进入真实 migration 实现前，必须有 schema artifact manifest、DDL review evidence、rollback evidence、backup requirement、migration lock、schema version table、manual apply gate 和 migration smoke。服务启动不允许自动运行 migration；app runtime role 不具备 migration 权限。

## Failure / Fallback Policy

schema / migration failure 必须 fail closed，并覆盖：

- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`
- `draft_store_contract_mismatch`
- `repository_store_disabled`
- `invalid_draft_store_mode`

这些失败不得 fallback 到 sample、fixture 或 memory dev store。`draft_version_conflict` 继续保持并发冲突语义。

## 验收口径

- `workflow-saved-draft-schema-migration-preconditions-v1` checker 通过。
- `./scripts/check-repo.sh --fast` 通过。
- 文档和 fixture 均保持 precondition-only，不声明 database ready、migration ready、repository adapter ready 或 production ready。

## 停止线

- 不实现 durable persistence、database schema、SQL migration、migration runner、repository interface、repository adapter、store selector、真实数据库、Radish OIDC、token validation、public production API、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把本任务卡、fixture 或 checker 解释为 durable store、database、schema migration、repository adapter、saved draft list、publish、run 或 production readiness。
