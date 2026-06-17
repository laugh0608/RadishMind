# Saved Workflow Draft Production Auth Readiness v1 专题

更新时间：2026-06-17

## 专题定位

`Saved Workflow Draft Production Auth Readiness v1` 承接 [Saved Workflow Draft Auth Context Preconditions v1](saved-workflow-draft-auth-context-preconditions-v1.md)、[Saved Workflow Draft Store Selector Implementation v1](saved-workflow-draft-store-selector-implementation-v1.md)、[Saved Workflow Draft Schema Artifact Materialization v1](saved-workflow-draft-schema-artifact-materialization-v1.md)、[Saved Workflow Draft Repository Adapter Implementation Plan v1](saved-workflow-draft-repository-adapter-implementation-plan-v1.md) 和 [Saved Workflow Draft Adapter Smoke Readiness v1](saved-workflow-draft-adapter-smoke-readiness-v1.md)，用于在创建 production auth middleware、token validation、membership adapter、repository adapter、production API 或 adapter smoke fixture 前，固定 saved draft durable store 所需的 production auth readiness 证据链。

本专题只定义 Radish OIDC issuer discovery evidence、token validation contract preconditions、claim mapping、tenant / workspace / application binding、scope projection、failure mapping、no fake fallback、no side effects 和 downstream readiness review。不创建 OIDC middleware、token validation、workspace membership adapter、repository interface、repository adapter、adapter smoke fixture、production API、SQL migration、migration runner、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_production_auth_readiness_defined`

## 当前输入事实

- `radish-oidc-client-preconditions` 已固定 Radish issuer、client、claim mapping、tenant binding、session boundary、logout / revocation、audit events 和 failure taxonomy，但没有实现 OIDC client、issuer network call 或 token validation。
- `Saved Workflow Draft Auth Context Preconditions v1` 已固定 `SavedWorkflowDraftRepositoryActorContext`、workspace membership policy、scope grant matrix、failure policy 和 audit / sanitization policy。
- `Saved Workflow Draft Store Selector Implementation v1` 已实现 formal store config、`SelectWorkflowSavedDraftStore`、selector tests 和 selector smoke checker；`repository_disabled`、`repository` 与 unknown mode 仍 fail closed。
- `Saved Workflow Draft Schema Artifact Materialization v1` 已物化 manifest、DDL review、rollback evidence 和 migration smoke 静态证据，但没有 SQL migration、schema version table 或 migration runner。
- `Saved Workflow Draft Repository Adapter Implementation Plan v1` 和 `Saved Workflow Draft Adapter Smoke Readiness v1` 已固定 future adapter / smoke 的 operation matrix、failure mapping、no fallback 和 no side effects，仍未打开 repository adapter 或 adapter smoke execution。

## Issuer Discovery Evidence Contract

future production auth 必须先提供经过人工 review 的 Radish issuer discovery evidence。该 evidence 至少包含：

- `issuer_url`
- `discovery_document_url`
- `jwks_uri`
- `supported_signing_algorithms`
- `supported_scopes`
- `fetched_at`
- `expires_at`
- `environment`
- `operator_review_ref`
- `sanitized_evidence_ref`

当前不允许联网拉取 issuer，不提交 raw discovery document 或 raw JWKS，不创建 OIDC client，也不把 issuer evidence 写成 runtime ready。

## Token Validation Preconditions

future token validation 必须先固定：

- issuer metadata pinned evidence。
- JWKS refresh、cache expiry 和 key rotation failure mapping。
- signing algorithm allowlist。
- issuer、audience 和 client id 检查。
- `exp`、`nbf`、`iat` 与 clock skew 规则。
- `sub`、`tenant_id` 和 permission-bearing claim 的必填检查。
- sanitized failure envelope，不能泄漏 token、authorization header、JWKS 或 raw claim dump。

本批不创建 `contracts/radish-oidc-token-validation.schema.json`、`workflow_saved_draft_token_validation.go` 或任何 token validation runtime。

## Claim Mapping

future saved draft actor context 的 claim mapping：

| claim | context field | required | failure |
| --- | --- | --- | --- |
| `sub` | `actor_subject_ref` | yes | `required_claim_missing` |
| `tenant_id` | `tenant_ref` | yes | `draft_tenant_binding_missing` |
| `roles` | `role_refs` | yes | `permission_claim_denied` |
| `permissions` | `scope_grants` | yes | `draft_scope_grant_missing` |
| `email` | `actor_email_ref` | no | none |
| `preferred_username` | `actor_display_ref` | no | none |

响应和 audit 只允许保存 sanitized ref；不得返回 raw token、raw claim dump、authorization header、cookie 或 client secret。

## Tenant / Workspace Binding

production auth readiness 必须同时约束 tenant、workspace、application 和 owner：

- `tenant_ref` 来自可信 auth context claim，不接受 query string、path 或 draft payload override。
- `workspace_id` 与 `application_id` 来自已验证 workspace context，并在 future membership adapter 中确认归属同一 tenant。
- `owner_subject_ref` 由 future owner policy resolver 提供；团队 owner、共享和转让仍是后续独立专题。
- tenant、workspace、application、owner 或 scope 缺失时必须 fail closed，不回退 local admin、test auth、memory dev store、sample 或 fixture。

## Scope Projection Matrix

| operation | required scope | current conclusion |
| --- | --- | --- |
| `SaveWorkflowDraftRecord` | `workflow_drafts:write` | readiness evidence 已定义；adapter smoke execution 与 repository adapter 仍 blocked |
| `ReadWorkflowDraftRecord` | `workflow_drafts:read` | readiness evidence 已定义；adapter smoke execution 与 repository adapter 仍 blocked |
| `ListWorkflowDraftRecords` | `workflow_drafts:read` | readiness evidence 已定义；adapter smoke execution 与 repository adapter 仍 blocked |

每个 operation 都必须消费 `tenant_ref`、`workspace_id`、`application_id`、`actor_subject_ref`、`owner_subject_ref` 和 `scope_grants`。缺失 identity、tenant binding、workspace membership 或 scope grant 时，不返回草案主体，不写 repository，不回退 test auth。

## Failure Mapping

本 readiness 固定两类 fail-closed code：

- saved draft auth context failure：`draft_identity_context_missing`、`draft_tenant_binding_missing`、`draft_workspace_membership_denied`、`draft_application_scope_denied`、`draft_owner_scope_denied`、`draft_scope_grant_missing`、`draft_auth_context_contract_mismatch`、`draft_audit_context_missing`、`draft_scope_denied`。
- Radish OIDC / token failure：`issuer_unavailable`、`issuer_metadata_invalid`、`jwks_unavailable`、`token_signature_invalid`、`token_expired`、`token_not_yet_valid`、`token_audience_invalid`、`token_issuer_invalid`、`unsupported_token_algorithm`、`required_claim_missing`、`permission_claim_denied`、`malformed_authorization_header`、`logout_failed`。

所有 failure 都必须使用 sanitized envelope，不能泄漏 token、issuer raw response、JWKS、database detail 或内部 adapter detail。

## Downstream Readiness Review

- Adapter smoke：production auth readiness evidence 已可作为后续 adapter smoke 输入；但 adapter smoke execution 仍等待真实 adapter smoke fixture / checker、repository adapter、auth middleware、token validation 和 membership adapter，当前不得执行。
- Repository adapter：production auth readiness evidence 已满足后续 implementation entry review 的 auth 准入输入；后续 [Saved Workflow Draft Repository Adapter Implementation Entry Review v1](saved-workflow-draft-repository-adapter-implementation-entry-review-v1.md) 已固定 `draft_repository_adapter_implementation_entry_review_defined`。repository interface、repository adapter、database query、token validation、membership adapter 和 production API 仍未打开。

结论是 `draft_production_auth_readiness_defined`，不是 `production_auth_ready`、`OIDC ready`、`repository adapter ready` 或 `adapter smoke ready`。

## 验收方式

- 新增 `workflow-saved-draft-production-auth-readiness-v1` fixture / checker，固定 issuer、token、claim、tenant / workspace binding、scope projection、failure mapping、no fake fallback、no side effects 和 downstream readiness review。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 `workflow-saved-draft-schema-artifact-materialization-v1` 之后。
- 本批至少运行专项 checker、auth context checker、adapter smoke readiness checker、repository adapter implementation plan checker、schema artifact materialization checker、`./scripts/check-repo.sh --fast`。
- 因同步 current focus，补跑 `./scripts/check-repo.sh`。

## 停止线

- 不创建 OIDC middleware、token validation、membership adapter、token validation schema、production auth smoke fixture / checker 或 production API consumer。
- 不创建 repository interface、repository adapter、adapter smoke fixture、adapter smoke checker、SQL migration、schema version table、migration runner、数据库连接或 database query。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 `draft_production_auth_readiness_defined` 解释为 Radish OIDC ready、token validation ready、production auth ready、repository adapter ready、adapter smoke ready、production API ready 或 production ready。
