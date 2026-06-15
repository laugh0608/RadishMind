# Workflow Saved Draft Auth Context Preconditions v1 任务卡

更新时间：2026-06-15

## 任务标识

- 切片：`workflow-saved-draft-auth-context-preconditions-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`draft_auth_context_preconditions_defined`

## 目标

在 schema / migration preconditions 之后，固定 future saved workflow draft repository actor context 的来源、workspace membership、owner policy、scope grant matrix、failure policy、audit / sanitization policy 和 implementation artifact guard。

本任务卡只定义 auth context preconditions，不创建 Radish OIDC middleware、token validation、session cookie、workspace membership adapter、repository interface、repository adapter、store selector、真实数据库、production API、publish、run、executor、confirmation、writeback 或 replay。

## 输入事实源

- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)
- [Saved Workflow Draft Durable Store Preconditions v1 专题](../features/workflow/saved-workflow-draft-durable-store-preconditions-v1.md)
- [Saved Workflow Draft Repository Contract Preconditions v1 专题](../features/workflow/saved-workflow-draft-repository-contract-preconditions-v1.md)
- [Saved Workflow Draft Schema / Migration Preconditions v1 专题](../features/workflow/saved-workflow-draft-schema-migration-preconditions-v1.md)
- [Saved Workflow Draft Auth Context Preconditions v1 专题](../features/workflow/saved-workflow-draft-auth-context-preconditions-v1.md)
- [User Workspace 设计与开发文档](../features/user-workspace.md)

## 本轮交付

- 新增 auth context preconditions 细专题，固定 precondition-only 状态。
- 新增 `workflow-saved-draft-auth-context-preconditions-v1` fixture / checker。
- checker 接入 fast baseline，校验 dependency、actor context mapping、workspace membership policy、scope grant matrix、owner policy、failure policy、audit / sanitization policy、文档引用和 forbidden implementation artifact。
- 同步更新 workflow 入口、Saved Draft 主专题、durable store / repository / schema preconditions、User Workspace、当前焦点、任务卡索引、脚本说明和周志。

## Auth Context Contract

future `SavedWorkflowDraftRepositoryActorContext` 必须覆盖：

- `request_id`
- `tenant_ref`
- `workspace_id`
- `application_id`
- `actor_subject_ref`
- `owner_subject_ref`
- `scope_grants`
- `audit_ref`

这些字段必须来自 future Radish OIDC / workspace context / owner policy / audit context 的已验证映射，不由 query string、draft payload、raw header、cookie、global current user 或 sample fixture 推导。

## Scope / Owner Policy

future repository operation 至少覆盖：

- `SaveWorkflowDraftRecord`：`workflow_drafts:write`，要求 `owner_or_workspace_write_grant`。
- `ReadWorkflowDraftRecord`：`workflow_drafts:read`，要求 `owner_or_workspace_read_grant`。
- `ListWorkflowDraftRecords`：`workflow_drafts:read`，要求 `owner_or_workspace_list_grant`。

团队 owner、共享、转让和跨 workspace 可见性不在本任务打开；后续需要独立专题。

## Failure / Sanitization Policy

auth context failure 必须 fail closed，并覆盖：

- `draft_identity_context_missing`
- `draft_tenant_binding_missing`
- `draft_workspace_membership_denied`
- `draft_application_scope_denied`
- `draft_owner_scope_denied`
- `draft_scope_grant_missing`
- `draft_auth_context_contract_mismatch`
- `draft_audit_context_missing`
- `draft_scope_denied`

audit 和 response 只能返回 sanitized reference，不返回 `id_token`、`access_token`、`refresh_token`、authorization header、cookie、client secret、raw claim dump 或 raw permission payload。

## 验收口径

- `workflow-saved-draft-auth-context-preconditions-v1` checker 通过。
- `./scripts/check-repo.sh --fast` 通过。
- 文档和 fixture 均保持 precondition-only，不声明 Radish OIDC ready、token validation ready、production auth ready、repository adapter ready 或 production ready。

## 停止线

- 不实现 durable persistence、Radish OIDC middleware、token validation、session cookie、workspace membership adapter、repository interface、repository adapter、store selector、真实数据库、public production API、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把本任务卡、fixture 或 checker 解释为 auth middleware、repository adapter、saved draft list、publish、run 或 production readiness。
