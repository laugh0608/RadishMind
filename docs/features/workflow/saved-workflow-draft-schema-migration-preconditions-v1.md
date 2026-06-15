# Saved Workflow Draft Schema / Migration Preconditions v1 专题

更新时间：2026-06-15

## 专题定位

`Saved Workflow Draft Schema / Migration Preconditions v1` 承接 [Saved Workflow Draft Repository Contract Preconditions v1](saved-workflow-draft-repository-contract-preconditions-v1.md)，用于在创建任何数据库 schema、SQL migration、migration runner、repository adapter 或 store selector 前，固定 future durable store 的逻辑 schema、索引策略、migration gate、failure mapping、no sample fallback 和 artifact guard。

本专题只定义 schema / migration preconditions，不创建真实数据库 schema、migration 文件、DDL、repository interface、repository adapter、store selector、Radish OIDC、production API、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_schema_migration_preconditions_defined`

## 当前输入事实

- `Saved Workflow Draft v1` 已有 memory dev store、save / read / validate、版本冲突、失败语义、sanitized response、no sample fallback 和 no side effects tests。
- `Saved Workflow Draft Durable Store Preconditions v1` 已固定 draft scope、owner / workspace 归属、version conflict、no sample fallback 和 dev store 切换停止线。
- `Saved Workflow Draft Repository Contract Preconditions v1` 已固定 future repository actor context、`SaveWorkflowDraftRecord` / `ReadWorkflowDraftRecord` / `ListWorkflowDraftRecords` operation matrix、request / result、failure 和 projection 边界。
- [Saved Workflow Draft Auth Context Preconditions v1](saved-workflow-draft-auth-context-preconditions-v1.md) 已固定 `draft_auth_context_preconditions_defined`，覆盖 future repository actor context 的身份来源、workspace membership、owner policy、scope grants、failure policy 和 audit / sanitization policy。
- [Saved Workflow Draft Store Selector Enablement Preconditions v1](saved-workflow-draft-store-selector-enablement-preconditions-v1.md) 已固定 `draft_store_selector_enablement_preconditions_defined`，覆盖 future store mode、selector gate、failure mapping、no fallback、no side effects 和 artifact guard。
- 当前仍没有 durable repository、database schema、SQL migration、migration runner、store selector、Radish OIDC 或 production API consumer。

## Logical Schema Scope

future durable store 的逻辑记录是 `saved_workflow_draft_record`。进入实际 schema artifact 前，必须保留以下字段分层：

- scope identity：`tenant_ref`、`workspace_id`、`application_id`、`draft_id`。
- ownership / actor：`owner_subject_ref`、`created_by_actor_ref`、`updated_by_actor_ref`。
- source / version：`source_definition_id`、`base_definition_version`、`draft_version`、`schema_version`、`store_schema_version`。
- status：`draft_status`。
- payload：`sanitized_draft_payload`、`validation_summary`、`blocked_capability_summary`。
- audit：`request_id`、`audit_ref`、`created_at`、`updated_at`。

`draft_id` 只能在 `tenant_ref + workspace_id + application_id` 内解释。`schema_version` 是 draft payload schema version；`store_schema_version` 是 durable store / migration artifact version，二者不能互相替代。

逻辑 schema 不保存 secret value、API key value、OAuth / OIDC token、authorization header、cookie、raw prompt dump、raw tool payload、provider response body、execution plan persistence、runtime readiness persistence、confirmation decision、business writeback payload、run result、replay state 或 resume state。

## Index Strategy

future schema artifact 至少需要固定以下逻辑索引策略：

| index id | fields | 用途 |
| --- | --- | --- |
| `saved_drafts_scope_lookup` | `tenant_ref, workspace_id, application_id, draft_id` | save / read 的唯一 scope lookup |
| `saved_drafts_owner_list` | `tenant_ref, workspace_id, application_id, owner_subject_ref, updated_at, draft_id` | User Workspace saved draft list |
| `saved_drafts_version_conflict` | `tenant_ref, workspace_id, application_id, draft_id, draft_version` | optimistic concurrency predicate |
| `saved_drafts_status_list` | `tenant_ref, workspace_id, application_id, owner_subject_ref, draft_status, updated_at, draft_id` | saved draft list 的状态过滤 |
| `saved_drafts_schema_version` | `tenant_ref, store_schema_version, schema_version` | migration smoke 和 schema compatibility 审查 |

所有 read / write predicate 都必须包含 `tenant_ref`、`workspace_id` 和 `application_id`，不得允许 query string 或 payload 覆盖 auth / workspace context。

## Migration Gate

进入真实 schema / migration 实现前，必须另行满足：

1. schema artifact manifest 已固定，包括 artifact id、store schema version、logical entity、DDL review evidence、rollback evidence 和 migration smoke evidence。
2. migration root 与命名规则已固定，例如 `services/platform/migrations/workflow_saved_drafts/` 和 `YYYYMMDDHHMMSS_workflow_saved_drafts_<short_slug>.sql`。
3. DDL review、rollback plan、backup requirement、migration lock、schema version table 和 manual apply gate 均已定义。
4. service startup 不允许自动运行 migration；app runtime role 不具备 migration 权限。
5. migration smoke 必须覆盖 schema version、tenant / workspace / application predicate、owner list projection、version conflict predicate、failure mapping、no sample fallback 和 no side effects。

本专题只保留规划，不创建 `services/platform/migrations/workflow_saved_drafts/` 目录或 SQL 文件。

## Failure Mapping

schema / migration gate 必须保持 fail closed，并新增 future store 层失败映射：

- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`
- `draft_store_contract_mismatch`
- `repository_store_disabled`
- `invalid_draft_store_mode`

这些失败不得 fallback 到 sample、fixture 或 memory dev store。`draft_version_conflict` 仍是业务并发冲突，不得被 migration failure 覆盖。

## 后续准入

本专题之后仍不能直接创建 repository adapter。后续若继续 durable store 方向，应按顺序选择一个独立批次：

1. schema artifact manifest / DDL review evidence。
2. schema artifact manifest / DDL review evidence，继续补足实际 schema artifact 前置证据。
3. selector smoke readiness / repository contract smoke，覆盖 no sample fallback、scope denied、owner denied、version conflict、store unavailable、schema migration failure 和 no side effects。

## 验收方式

- 新增 task card 固定本专题为 precondition-only。
- 新增 fixture / checker 固定 logical schema、index strategy、migration gate、failure mapping、no sample fallback 和 implementation artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`。
- 本批至少运行专项 checker 和 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建真实数据库 schema、SQL migration、migration runner、repository interface、repository adapter、store selector、Radish OIDC middleware、token validation 或 production API consumer。
- 不实现 durable persistence、saved draft list、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 schema / migration preconditions 解释为 database ready、migration ready、repository adapter ready、store selector ready 或 production ready。
