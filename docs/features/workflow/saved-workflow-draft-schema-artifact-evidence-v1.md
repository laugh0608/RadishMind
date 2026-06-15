# Saved Workflow Draft Schema Artifact Evidence v1 专题

更新时间：2026-06-15

## 专题定位

`Saved Workflow Draft Schema Artifact Evidence v1` 承接 [Saved Workflow Draft Schema / Migration Preconditions v1](saved-workflow-draft-schema-migration-preconditions-v1.md)、[Saved Workflow Draft Auth Context Preconditions v1](saved-workflow-draft-auth-context-preconditions-v1.md) 和 [Saved Workflow Draft Store Selector Enablement Preconditions v1](saved-workflow-draft-store-selector-enablement-preconditions-v1.md)，用于在创建 schema artifact manifest、DDL review artifact、SQL migration、migration runner、repository adapter 或 store selector 前，固定 future schema artifact 证据链。

本专题只定义 schema artifact manifest / DDL review evidence contract，不创建 migration root、manifest 文件、DDL review 文件、SQL migration、schema version table、migration runner、真实数据库、repository interface、repository adapter、store selector、Radish OIDC middleware、production API、saved draft list、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_schema_artifact_evidence_defined`

## 当前输入事实

- `Saved Workflow Draft v1` 已有 memory dev store、dev-only route + web consumer、save / read / validate、版本冲突、no sample fallback 和 no side effects tests。
- `Saved Workflow Draft Repository Contract Preconditions v1` 已固定 future repository operation matrix。
- `Saved Workflow Draft Schema / Migration Preconditions v1` 已固定 `saved_workflow_draft_record` logical schema、index strategy、migration gate、failure mapping 和 artifact guard。
- `Saved Workflow Draft Auth Context Preconditions v1` 已固定 future repository actor context、workspace membership、owner policy、scope grants 和 audit / sanitization policy。
- `Saved Workflow Draft Store Selector Enablement Preconditions v1` 已固定 `memory_dev` / `repository_disabled` / `repository` store mode、selector gate、failure mapping、no fallback、no side effects 和 dev flag boundary。
- 当前仍没有 schema artifact manifest、DDL review artifact、SQL migration、migration runner、store selector、repository adapter、database schema、Radish OIDC middleware、token validation 或 production API consumer。

## Schema Artifact Evidence Contract

future `workflow_saved_draft_schema_artifact_v1` 进入 materialized artifact 前，必须先固定以下 evidence：

- manifest：`services/platform/migrations/workflow_saved_drafts/manifest.json`，记录 artifact id、store schema version、logical entity、field mapping、index mapping、migration id、up / down migration path、DDL review ref、rollback ref、migration smoke ref、failure mapping 和 no side effect counters。
- DDL review：`services/platform/migrations/workflow_saved_drafts/ddl-review.md`，记录人工审查结论、manual apply command、backup requirement、migration lock、schema version table、rollback / forward-only 策略和 destructive change gate。
- rollback evidence：记录 rollback command 或 forward-only exception、备份要求、失败迁移恢复和 migration lock release。
- migration smoke：覆盖 schema version table、`tenant_ref + workspace_id + application_id` predicate、owner list projection、version conflict predicate、migration failure mapping、no sample fallback 和 no side effects。

本批不创建上述文件，只固定它们未来必须具备的字段、引用关系和停止线。

## Logical Entity Mapping

future schema artifact 必须消费 `Saved Workflow Draft Schema / Migration Preconditions v1` 的 logical schema：

- logical entity：`saved_workflow_draft_record`。
- store schema version：`saved_workflow_drafts_store_v1`。
- scope predicate：`tenant_ref + workspace_id + application_id + draft_id`。
- owner list predicate：`tenant_ref + workspace_id + application_id + owner_subject_ref`。
- version conflict predicate：`tenant_ref + workspace_id + application_id + draft_id + draft_version`。
- schema compatibility predicate：`tenant_ref + store_schema_version + schema_version`。

`schema_version` 仍表示 draft payload schema version；`store_schema_version` 仍表示 durable store artifact version。两者必须同时进入 manifest 和 migration smoke，不能互相替代。

## DDL Review Gate

DDL review evidence 必须保持 fail closed：

1. 人工审查完成前不得创建 SQL migration。
2. service startup 不允许自动运行 migration。
3. app runtime role 不具备 migration 权限。
4. destructive change 必须有人工审查、备份要求和 rollback / forward-only exception。
5. schema version table、migration lock 和 manual apply gate 必须先定义。
6. DDL review 缺失时，不得 fallback 到 sample、fixture 或 memory dev store。

## Failure Mapping

schema artifact evidence 相关 failure 必须继续使用既有 fail-closed code：

- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`
- `draft_store_contract_mismatch`
- `repository_store_disabled`
- `invalid_draft_store_mode`

`draft_version_conflict`、`draft_scope_denied` 和 `draft_store_unavailable` 仍保留原始语义，不得被 migration failure 覆盖。

## 后续准入

本专题之后仍不能直接创建 repository adapter。后续若继续 durable store 方向，应选择一个独立批次：

1. selector smoke readiness / repository contract smoke，消费本专题的 schema failure 和 no fallback 证据。
2. repository adapter implementation plan，必须消费 schema preconditions、auth context、store selector gate 和 schema artifact evidence。
3. schema artifact materialization review，另行决定是否创建 manifest / DDL review / migration smoke artifact；进入该批前仍不得连接真实数据库。

## 验收方式

- 新增 task card 固定本专题为 evidence-only。
- 新增 fixture / checker 固定 schema artifact boundary、manifest contract、logical entity mapping、index mapping、DDL review gate、migration smoke evidence、failure mapping、no fallback、no side effects 和 implementation artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`。
- 本批至少运行专项 checker 和 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建 migration root、manifest 文件、DDL review 文件、SQL migration、schema version table、migration runner、数据库连接、repository interface、repository adapter、store selector、Radish OIDC middleware、token validation、production API consumer 或 saved draft list。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 schema artifact evidence 解释为 schema artifact ready、DDL ready、migration ready、database ready、repository adapter ready、store selector ready 或 production ready。
