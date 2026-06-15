# Saved Workflow Draft Repository Contract Preconditions v1 专题

更新时间：2026-06-15

## 专题定位

`Saved Workflow Draft Repository Contract Preconditions v1` 承接 [Saved Workflow Draft Durable Store Preconditions v1](saved-workflow-draft-durable-store-preconditions-v1.md)，用于在 durable repository adapter 实现前固定 repository contract 的输入、输出、scope predicate、owner / workspace 字段、failure code、sanitized projection 和验证策略。

本专题只定义 repository contract preconditions，不创建 `SavedWorkflowDraftRepository` interface、adapter、store selector、SQL、migration、Radish OIDC、production API、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_repository_contract_preconditions_defined`

## 当前输入事实

- `Saved Workflow Draft v1` 已有 memory dev store、save / read / validate、版本冲突、no sample fallback 和 no side effects tests。
- `Saved Workflow Draft Durable Store Preconditions v1` 已明确 draft scope、owner / workspace、version conflict、no sample fallback 和 store 切换停止线。
- [Saved Workflow Draft Schema / Migration Preconditions v1](saved-workflow-draft-schema-migration-preconditions-v1.md) 已固定 `draft_schema_migration_preconditions_defined`，覆盖 future durable store logical schema、index strategy、migration gate、failure mapping、no sample fallback 和 artifact guard。
- 当前仍没有 durable repository adapter、database schema、store selector、Radish OIDC 或 production API consumer。

## Repository Actor Context

未来 repository contract 必须接收结构化 actor context，而不是读取 HTTP request、query string、header 或全局当前用户：

- `request_id`
- `tenant_ref`
- `workspace_id`
- `application_id`
- `actor_subject_ref`
- `owner_subject_ref`
- `scope_grants`
- `audit_ref`

`tenant_ref + workspace_id + application_id` 必须来自已验证 auth / workspace context。`draft_id` 只能在该 scope 下解释；repository 层不得接受 query 中的 tenant / workspace override。

## Repository Operations

未来 `SavedWorkflowDraftRepository` contract 至少需要覆盖三类 operation：

| operation | scope | 作用 |
| --- | --- | --- |
| `SaveWorkflowDraftRecord` | `workflow_drafts:write` | 原子保存或更新草案记录，使用 `expected_draft_version` 做乐观并发控制 |
| `ReadWorkflowDraftRecord` | `workflow_drafts:read` | 按 scope + `draft_id` 读取 sanitized saved draft record |
| `ListWorkflowDraftRecords` | `workflow_drafts:read` | 按 workspace / application 列出 saved draft summary，服务后续 saved draft list |

contract 必须保持 repository 层只处理草案记录读写，不执行节点、不调用 provider/tool、不发布 workflow、不创建 run record、不提交 confirmation decision、不写业务真相源。

## Request / Result Contract

保存请求必须包含：

- `repository_actor_context`
- `draft_scope`
- `expected_draft_version`
- `sanitized_draft_payload`
- `validation_summary`
- `blocked_capability_summary`
- `request_audit_metadata`

读取请求必须包含：

- `repository_actor_context`
- `draft_scope`
- `draft_id`
- `projection`

列表请求必须包含：

- `repository_actor_context`
- `workspace_id`
- `application_id`
- `owner_subject_ref`
- `cursor`
- `limit`
- `sort`

返回结果必须保留：

- `draft_id`
- `draft_version`
- `schema_version`
- `draft_status`
- `owner_subject_ref`
- `created_by_actor_ref`
- `updated_by_actor_ref`
- `validation_summary`
- `blocked_capability_summary`
- `failure_code`
- `request_id`
- `audit_ref`

## Failure Policy

repository contract 必须沿用 saved draft 的 fail-closed 语义：

- `draft_scope_denied`
- `draft_not_found`
- `draft_schema_version_unsupported`
- `draft_payload_invalid`
- `draft_version_conflict`
- `draft_store_unavailable`
- `draft_store_contract_mismatch`
- `repository_store_disabled`
- `invalid_draft_store_mode`

失败不得回退 sample / fixture，也不得 fallback 到 memory dev store。`draft_version_conflict` 必须保留当前版本 metadata，并拒绝覆盖。

## Projection Policy

repository result 只能返回 sanitized saved draft record、summary 和可审查 metadata。必须拒绝或剔除：

- secret value、API key value、token、authorization header、cookie。
- raw prompt dump、raw tool payload、provider response body。
- execution plan persistence、runtime readiness persistence、confirmation decision、business writeback payload、run result、replay / resume state。

## 后续准入

进入 repository interface / adapter 实现前，必须另行完成：

1. schema / migration preconditions 已固定；进入实际 schema artifact 或 SQL migration 前，仍需 schema artifact manifest、DDL review evidence、rollback evidence 和 migration smoke。
2. store selector enablement 设计，且 repository / database mode 不允许 fallback 到 memory dev store。
3. Radish OIDC / auth context contract，明确 tenant、workspace membership、subject、owner 和 scope 映射。
4. repository contract smoke、adapter smoke、no sample fallback、version conflict、scope denied、store unavailable 和 no side effects 测试。

## 验收方式

- 新增 task card 固定本专题为 precondition-only。
- 新增 fixture / checker 固定 actor context、operation matrix、request / result contract、failure policy、projection policy 和 implementation artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建 `SavedWorkflowDraftRepository` interface、repository adapter、store selector、SQL、migration、Radish OIDC middleware、token validation 或 production API consumer。
- 不实现 durable persistence、saved draft list、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 repository contract preconditions 解释为 repository adapter ready、database ready 或 production ready。
