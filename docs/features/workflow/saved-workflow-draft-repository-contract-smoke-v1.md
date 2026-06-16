# Saved Workflow Draft Repository Contract Smoke v1 专题

更新时间：2026-06-16

## 专题定位

`Saved Workflow Draft Repository Contract Smoke v1` 承接 [Saved Workflow Draft Repository Contract Preconditions v1](saved-workflow-draft-repository-contract-preconditions-v1.md)、[Saved Workflow Draft Auth Context Preconditions v1](saved-workflow-draft-auth-context-preconditions-v1.md)、[Saved Workflow Draft Schema Artifact Evidence v1](saved-workflow-draft-schema-artifact-evidence-v1.md) 和 [Saved Workflow Draft Store Selector Smoke Readiness v1](saved-workflow-draft-store-selector-smoke-readiness-v1.md)，用于在实现 `SavedWorkflowDraftRepository` interface、repository adapter、store selector、数据库或 production API 前，固定 future repository contract smoke 的输入输出、operation matrix、failure mapping、no fallback、no side effects 和 implementation artifact guard。

本专题只定义 repository contract smoke 的静态准入记录，不创建 smoke runner、repository interface、repository adapter、store selector、SQL migration、Radish OIDC middleware、token validation、production API、saved draft list 新实现、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_repository_contract_smoke_defined`

## 当前输入事实

- `Saved Workflow Draft Repository Contract Preconditions v1` 已固定 future `SaveWorkflowDraftRecord`、`ReadWorkflowDraftRecord` 和 `ListWorkflowDraftRecords` 的 actor context、request / result、failure 和 sanitized projection。
- `Saved Workflow Draft Auth Context Preconditions v1` 已固定 future repository actor context 的身份来源、workspace membership、owner policy、scope grants、failure policy 和 audit / sanitization 边界。
- `Saved Workflow Draft Schema Artifact Evidence v1` 已固定 future schema artifact manifest、DDL review、rollback evidence、migration smoke、schema failure mapping 和 artifact guard。
- `Saved Workflow Draft Store Selector Smoke Readiness v1` 已固定 future selector smoke 的 `memory_dev` / `repository_disabled` / `repository` / unknown mode、operation matrix、schema artifact failure、no fallback 和 no side effects。
- `Saved Workflow Draft Repository Contract Smoke Runner Readiness v1` 已固定 `draft_repository_contract_smoke_runner_readiness_defined`，覆盖 future `SavedWorkflowDraftRepositoryContractSmokeRunner` 的 runner I/O、operation runner matrix、failure mapping、no fallback、no side effects 和 artifact guard。
- 当前仍没有 repository contract smoke runner、repository interface、repository adapter、store selector implementation、SQL migration、真实数据库、Radish OIDC middleware、token validation 或 production API consumer。

## Smoke Boundary

future smoke harness 名称固定为 `SavedWorkflowDraftRepositoryContractSmoke`。它只能在后续独立批次中消费 future `SavedWorkflowDraftRepository` adapter，不得直接读取 dev memory store、sample fixture、HTTP query string、raw auth header、cookie 或全局当前用户。

本批只记录以下事实：

- 当前 store source 仍是 `platform memory dev store`。
- future store source 仍是 `future SavedWorkflowDraftRepository adapter`。
- repository contract smoke runner、repository interface、repository adapter、selector、SQL、OIDC 和 production API 均未创建。
- future smoke 失败时不得 fallback 到 memory dev store、sample 或 fixture。

## Smoke I/O Contract

future smoke 输入必须包含结构化 `repository_actor_context`，字段保持：

- `request_id`
- `tenant_ref`
- `workspace_id`
- `application_id`
- `actor_subject_ref`
- `owner_subject_ref`
- `scope_grants`
- `audit_ref`

`draft_scope` 至少包含：

- `tenant_ref`
- `workspace_id`
- `application_id`
- `draft_id`

`SaveWorkflowDraftRecord` smoke 请求必须覆盖 `expected_draft_version`、sanitized draft payload、validation summary、blocked capability summary 和 request audit metadata。`ReadWorkflowDraftRecord` smoke 请求必须覆盖 draft scope、`draft_id` 和 projection。`ListWorkflowDraftRecords` smoke 请求必须覆盖 workspace、application、owner、cursor、limit 和 sort。

future smoke 输出只能是 sanitized saved draft record、saved draft summary、validation summary、blocked capability summary、current version metadata、request id、audit ref 或 failure code，不得返回 secret、token、raw provider payload、execution plan persistence、runtime readiness persistence、confirmation decision、business writeback payload、run result、replay 或 resume state。

## Operation Smoke Matrix

future smoke 必须覆盖以下 operation：

| operation | scope | success projection | 必须覆盖的失败 |
| --- | --- | --- | --- |
| `SaveWorkflowDraftRecord` | `workflow_drafts:write` | sanitized saved draft record + current version metadata | `draft_scope_denied`、`draft_schema_version_unsupported`、`draft_payload_invalid`、`draft_version_conflict`、`draft_store_unavailable`、`draft_store_contract_mismatch`、schema artifact failure |
| `ReadWorkflowDraftRecord` | `workflow_drafts:read` | sanitized saved draft record | `draft_scope_denied`、`draft_not_found`、`draft_schema_version_unsupported`、`draft_store_unavailable`、`draft_store_contract_mismatch`、schema artifact failure |
| `ListWorkflowDraftRecords` | `workflow_drafts:read` | sanitized saved draft summary list | `draft_scope_denied`、`repository_store_disabled`、`invalid_draft_store_mode`、`draft_store_unavailable`、`draft_store_contract_mismatch`、schema artifact failure |

每个 operation 都必须证明 denied / failed path 不调用 provider/tool、不创建 run record、不提交 confirmation decision、不写业务真相源、不触发 replay，也不把 version conflict 重写成 selector 或 schema failure。

## Failure Mapping

repository contract smoke 必须保留以下 failure code：

- `draft_scope_denied`
- `draft_not_found`
- `draft_schema_version_unsupported`
- `draft_payload_invalid`
- `draft_version_conflict`
- `draft_store_unavailable`
- `draft_store_contract_mismatch`
- `repository_store_disabled`
- `invalid_draft_store_mode`
- `draft_auth_context_contract_mismatch`
- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`

失败路径不得 fallback 到 memory dev store、sample 或 fixture。`draft_version_conflict` 必须保留当前版本 metadata；`draft_scope_denied` 和 `draft_not_found` 不得返回草案主体。

## 后续准入

本专题完成后仍不能直接创建 repository adapter。`Saved Workflow Draft Repository Contract Smoke Runner Readiness v1` 已固定 `draft_repository_contract_smoke_runner_readiness_defined`；后续若继续 durable store 方向，应选择一个独立批次：

1. repository contract smoke runner implementation，消费本专题定义的 harness、operation matrix、failure mapping 和 no side effects 约束，但仍不连接真实数据库。
2. repository adapter implementation plan，必须消费 schema artifact evidence、auth context、store selector smoke readiness 和 repository contract smoke；进入前仍需明确 adapter smoke、migration artifact、auth middleware 和 production API 的停止线。
3. selector implementation entry review，另行决定是否创建 formal config、selector 函数、selector tests 和 selector smoke fixture；进入该批前仍不得连接真实数据库。

## 验收方式

- 新增 task card 固定本专题为 static smoke definition。
- 新增 fixture / checker 固定 repository contract smoke boundary、I/O contract、operation smoke matrix、failure mapping、no fallback、no side effects 和 implementation artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 store selector smoke readiness 之后。
- 本批至少运行专项 checker、相邻 selector smoke readiness checker 和 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建 repository contract smoke runner、`SavedWorkflowDraftRepository` interface、repository adapter、store selector、SQL migration、真实数据库、Radish OIDC middleware、token validation、production API consumer 或新的 saved draft list 实现。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 repository contract smoke 定义解释为 smoke runner ready、repository adapter ready、database ready、selector ready、OIDC ready 或 production ready。
