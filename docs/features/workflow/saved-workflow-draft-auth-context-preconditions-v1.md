# Saved Workflow Draft Auth Context Preconditions v1 专题

更新时间：2026-06-15

## 专题定位

`Saved Workflow Draft Auth Context Preconditions v1` 承接 [Saved Workflow Draft Schema / Migration Preconditions v1](saved-workflow-draft-schema-migration-preconditions-v1.md)，用于在 store selector、repository adapter、真实数据库或 production API 前固定 saved workflow draft 的 auth context contract。[Saved Workflow Draft Store Selector Enablement Preconditions v1](saved-workflow-draft-store-selector-enablement-preconditions-v1.md) 已在本专题之后固定 `draft_store_selector_enablement_preconditions_defined`。

本专题只定义 auth context preconditions，不创建 Radish OIDC middleware、token validation、session cookie、workspace membership adapter、repository interface、repository adapter、store selector、真实数据库、production API、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_auth_context_preconditions_defined`

## 当前输入事实

- `Saved Workflow Draft v1` 已有 memory dev store、dev-only route + web consumer、save / read / validate、版本冲突、no sample fallback 和 no side effects tests。
- `Saved Workflow Draft Durable Store Preconditions v1` 已固定 draft scope、owner / workspace 归属、version conflict 和 dev store 切换停止线。
- `Saved Workflow Draft Repository Contract Preconditions v1` 已固定 `SavedWorkflowDraftRepositoryActorContext` 需要的 actor context 字段。
- `Saved Workflow Draft Schema / Migration Preconditions v1` 已固定 future durable store 的 logical schema、index strategy、migration gate、failure mapping 和 artifact guard。
- `radish-oidc-client-preconditions` 已固定 Radish OIDC issuer、claim mapping、tenant binding、session boundary 和 failure taxonomy 的治理边界，但当前仍没有真实 OIDC client、token validation 或 production auth middleware。

## Auth Context Contract

future saved draft repository contract 必须接收结构化 `SavedWorkflowDraftRepositoryActorContext`，不得在 repository 层读取 HTTP request、query string、authorization header、cookie 或全局 current user。

required context fields：

- `request_id`
- `tenant_ref`
- `workspace_id`
- `application_id`
- `actor_subject_ref`
- `owner_subject_ref`
- `scope_grants`
- `audit_ref`

字段来源必须明确：

- `tenant_ref`：来自未来 Radish OIDC tenant binding，不接受 query string 或 draft payload override。
- `workspace_id`：来自已验证 workspace context，不由 draft payload 决定访问权限。
- `application_id`：来自已验证 workspace application context，不由 draft payload 决定访问权限。
- `actor_subject_ref`：来自未来 subject claim mapping；当前 dev-only 路径只能使用测试身份。
- `owner_subject_ref`：来自显式 owner policy；默认可以是 actor subject，团队 owner / 转让 / 共享需要后续独立专题。
- `scope_grants`：来自 future permission / role / scope projection。
- `audit_ref`：来自 request audit context，只保存引用和 sanitized metadata。

## Workspace Membership Policy

进入 repository adapter 或 saved draft list 前，必须固定 workspace membership policy：

- tenant binding 缺失、workspace membership 缺失、application scope 缺失或 owner policy 缺失时，必须 fail closed。
- `workspace_id` 与 `application_id` 必须归属于同一 `tenant_ref`。
- `owner_subject_ref` 不是跨 workspace 授权；团队可见性、共享和转让不在本专题打开。
- `draft_id` 只能在 `tenant_ref + workspace_id + application_id` 内解释。
- dev-only route 可以继续使用测试身份，但不得把 fake auth header 写成 production auth 输入。

## Scope Grant Matrix

future auth context 至少要覆盖三类 repository operation：

| operation | required scope | owner policy |
| --- | --- | --- |
| `SaveWorkflowDraftRecord` | `workflow_drafts:write` | `owner_or_workspace_write_grant` |
| `ReadWorkflowDraftRecord` | `workflow_drafts:read` | `owner_or_workspace_read_grant` |
| `ListWorkflowDraftRecords` | `workflow_drafts:read` | `owner_or_workspace_list_grant` |

缺少 scope grant、workspace membership 或 owner policy 时，必须返回 fail-closed failure code，不得 fallback 到 sample、fixture 或 memory dev store。

## Failure Policy

auth context contract 必须新增或保留以下 failure code：

- `draft_identity_context_missing`
- `draft_tenant_binding_missing`
- `draft_workspace_membership_denied`
- `draft_application_scope_denied`
- `draft_owner_scope_denied`
- `draft_scope_grant_missing`
- `draft_auth_context_contract_mismatch`
- `draft_audit_context_missing`
- `draft_scope_denied`

这些 failure 不得返回草案主体，不得触发保存、发布、运行、confirmation decision、writeback 或 replay。`draft_scope_denied` 仍是面向 saved draft consumer 的统一拒绝语义；细分 auth failure code 可用于日志、审计和调试 metadata。

## Audit / Sanitization Policy

audit context 只能保存 request / actor / tenant / workspace / application / owner / failure 的 sanitized reference：

- allowed：`request_id`、`audit_ref`、`tenant_ref`、`workspace_id`、`application_id`、`actor_subject_ref`、`owner_subject_ref`、`scope_grant_ids`、`failure_code`。
- forbidden：`id_token`、`access_token`、`refresh_token`、`authorization_header`、`cookie_value`、`client_secret`、raw claim dump、raw permission payload、raw OIDC/JWKS response。

## 后续准入

本专题之后仍不能直接创建 repository adapter。后续若继续 durable store 方向，应按顺序选择一个独立批次：

1. schema artifact manifest / DDL review evidence，继续补足实际 schema artifact 前置证据。
2. selector smoke readiness / repository contract smoke，覆盖 auth context、scope denied、owner denied、version conflict、store unavailable、schema migration failure、no sample fallback 和 no side effects。
3. repository interface / adapter implementation plan，必须消费 schema、auth、selector 和 smoke gate。

## 验收方式

- 新增 task card 固定本专题为 precondition-only。
- 新增 fixture / checker 固定 auth context mapping、workspace membership、scope grant matrix、owner policy、failure policy、audit / sanitization policy 和 implementation artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`。
- 本批至少运行专项 checker 和 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建 Radish OIDC middleware、token validation、session cookie、workspace membership adapter、repository interface、repository adapter、store selector、真实数据库、production API consumer 或 saved draft list。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 auth context preconditions 解释为 Radish OIDC ready、token validation ready、production auth ready、repository adapter ready 或 production ready。
