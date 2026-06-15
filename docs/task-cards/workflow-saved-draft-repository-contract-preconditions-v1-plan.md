# Workflow Saved Draft Repository Contract Preconditions v1 任务卡

更新时间：2026-06-15

## 任务标识

- 切片：`workflow-saved-draft-repository-contract-preconditions-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`draft_repository_contract_preconditions_defined`

## 目标

在 durable store 迁移前置设计之后，固定 `SavedWorkflowDraftRepository` 未来 contract 的 actor context、operation matrix、request / result contract、failure policy、sanitized projection 和 artifact guard。

本任务卡只定义 repository contract preconditions，不创建 repository interface、adapter、store selector、SQL、migration、Radish OIDC、production API、publish、run、executor、confirmation、writeback 或 replay。

## 输入事实源

- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)
- [Saved Workflow Draft Durable Store Preconditions v1 专题](../features/workflow/saved-workflow-draft-durable-store-preconditions-v1.md)
- [Saved Workflow Draft Repository Contract Preconditions v1 专题](../features/workflow/saved-workflow-draft-repository-contract-preconditions-v1.md)
- [User Workspace 设计与开发文档](../features/user-workspace.md)

## 本轮交付

- 新增 repository contract preconditions 细专题，固定 precondition-only 状态。
- 新增 `workflow-saved-draft-repository-contract-preconditions-v1` fixture / checker。
- checker 接入 fast baseline，校验 dependency、actor context、operation matrix、request/result field、failure/projection policy、后续准入和 implementation artifact guard。
- 同步更新 workflow 入口、Saved Draft 主专题、durable store 前置专题、User Workspace、当前焦点、任务卡索引、脚本说明和周志。

## Actor Context Contract

未来 repository contract 必须通过结构化 actor context 接收身份和 scope：

- `request_id`
- `tenant_ref`
- `workspace_id`
- `application_id`
- `actor_subject_ref`
- `owner_subject_ref`
- `scope_grants`
- `audit_ref`

repository contract 不读取 HTTP request、query string、raw authorization header、cookie、全局 current tenant 或 provider runtime。

## Operation Matrix

本批只定义 future contract，不创建 interface。未来 contract 至少覆盖：

- `SaveWorkflowDraftRecord`
- `ReadWorkflowDraftRecord`
- `ListWorkflowDraftRecords`

`SaveWorkflowDraftRecord` 必须使用 `expected_draft_version`，并在冲突时返回 `draft_version_conflict`。`ListWorkflowDraftRecords` 只返回 summary / metadata，用于后续 User Workspace saved draft list，不打开完整 durable persistence UI。

## Request / Result Contract

保存 contract 必须覆盖 `repository_actor_context`、`draft_scope`、`expected_draft_version`、`sanitized_draft_payload`、`validation_summary`、`blocked_capability_summary` 和 `request_audit_metadata`。

读取 contract 必须覆盖 `repository_actor_context`、`draft_scope`、`draft_id` 和 `projection`。

列表 contract 必须覆盖 `repository_actor_context`、`workspace_id`、`application_id`、`owner_subject_ref`、`cursor`、`limit` 和 `sort`。

result 必须覆盖 `draft_id`、`draft_version`、`schema_version`、`draft_status`、`owner_subject_ref`、`created_by_actor_ref`、`updated_by_actor_ref`、`validation_summary`、`blocked_capability_summary`、`failure_code`、`request_id` 和 `audit_ref`。

## Failure / Projection Policy

contract 必须保持 fail closed：

- scope denied、not found、schema unsupported、payload invalid、version conflict、store unavailable、store contract mismatch、repository disabled 和 invalid store mode 都不得 fallback 到 sample、fixture 或 memory dev store。
- result projection 不返回 secret、token、authorization header、cookie、raw prompt dump、raw tool payload、provider response body、confirmation decision、business writeback payload、run result 或 replay / resume state。

## 后续准入

本任务之后可以继续推进 schema / migration 设计、auth context contract 或 store selector enablement；不得直接创建 repository adapter。进入 repository interface / adapter 前，必须先有 schema / migration、store selector、Radish OIDC / auth context、repository contract smoke、adapter smoke、no sample fallback、version conflict、scope denied、store unavailable 和 no side effects 测试。

## 验收口径

- `workflow-saved-draft-repository-contract-preconditions-v1` checker 通过。
- `./scripts/check-repo.sh --fast` 通过。
- 文档和 fixture 均保持 precondition-only，不声明 repository adapter ready、database ready 或 production ready。

## 停止线

- 不实现 durable persistence、repository interface、repository adapter、schema migration、store selector、真实数据库、Radish OIDC、token validation、public production API、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把本任务卡、fixture 或 checker 解释为 durable store、repository adapter、saved draft list、publish、run 或 production readiness。
