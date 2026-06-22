# Radish OIDC Token / Membership Readiness v1

更新时间：2026-06-22

## 专题定位

`Radish OIDC Token / Membership Readiness v1` 是 RadishMind 接入 `Radish` 身份体系前的跨产品面前置专题。它消费既有 `radish-oidc-client-preconditions`、control plane read production auth readiness、saved workflow draft production auth readiness 和 saved draft production auth runtime bridge，把 future token validation 与 membership adapter 的输入、输出、失败语义和停止线收束为可检查证据。

本专题不实现 OIDC client、auth middleware、token validator、membership adapter、login / logout、session cookie、production API consumer、repository mode、数据库连接、workflow executor、confirmation、writeback 或 replay。

状态：`radish_oidc_token_membership_readiness_defined`

## 当前输入事实

- `radish-oidc-client-preconditions` 已固定 issuer、client、claim mapping、tenant binding、session boundary、logout / revocation、audit events 和 failure taxonomy。
- `control-plane-read-production-auth-readiness-v1` 已固定 read-side production auth readiness，但不创建 auth middleware、token validation 或 production API consumer。
- `workflow-saved-draft-production-auth-readiness-v1` 已固定 saved draft repository actor context 所需的 issuer / token / claim / tenant-workspace binding / scope projection readiness。
- `workflow-saved-draft-production-auth-runtime-v1` 已实现 verified auth context + workspace binding 到 repository actor context 的纯函数投影；它只消费已验证上下文，不校验 token、不查询 membership。

## Token Validation Contract

future token validation 必须先消费人工 review 后的 Radish issuer evidence，并满足以下 gate：

- issuer metadata pinned evidence。
- JWKS refresh、cache expiry 和 key rotation failure mapping。
- signing algorithm allowlist。
- issuer、audience 和 client id 检查。
- `exp`、`nbf`、`iat` 与 clock skew 规则。
- `sub`、`tenant_id`、`roles` 和 `permissions` 必填检查。
- sanitized failure envelope，不泄漏 raw token、authorization header、JWKS、raw claim dump、cookie 或 client secret。

本批不创建 `contracts/radish-oidc-token-validation.schema.json`，不调用 issuer，不读取 JWKS，不做 runtime token validation。

## Membership Contract

future membership adapter 的输入只能来自已验证 token context、workspace / application / owner binding request 和 audit context。输出只能是 metadata-only membership decision：

- `tenant_ref`
- `workspace_id`
- `application_id`
- `actor_subject_ref`
- `owner_subject_ref`
- `scope_grants`
- `membership_verified`
- `application_scope_verified`
- `owner_scope_verified`
- `audit_ref`

membership 缺失、tenant mismatch、workspace denied、application denied、owner denied 或 scope 缺失时必须 fail closed，不回退 dev fake auth、local admin、memory dev store、sample 或 fixture。

## Consumer Matrix

| consumer | future use | readiness conclusion |
| --- | --- | --- |
| Control Plane Read | read route auth middleware 与 read repository context | readiness defined；middleware / production API 仍未打开 |
| Saved Workflow Draft | repository actor context 与 repository mode runtime 前置 | readiness defined；membership adapter / repository mode 仍未打开 |
| Admin Control Plane | tenant / role / audit 管理端前置 | readiness defined；正式管理端 auth runtime 仍未打开 |
| Model Gateway / API Distribution | API key / quota / trace 后续 subject binding | readiness defined；API key lifecycle 和 quota enforcement 仍未打开 |

## Failure Mapping

本专题固定以下失败码必须进入 future token / membership readiness：

- OIDC / token：`issuer_unavailable`、`issuer_metadata_invalid`、`jwks_unavailable`、`token_signature_invalid`、`token_expired`、`token_not_yet_valid`、`token_audience_invalid`、`token_issuer_invalid`、`unsupported_token_algorithm`、`required_claim_missing`、`permission_claim_denied`、`malformed_authorization_header`
- binding / membership：`tenant_binding_missing`、`tenant_binding_denied`、`draft_identity_context_missing`、`draft_tenant_binding_missing`、`draft_workspace_membership_denied`、`draft_application_scope_denied`、`draft_owner_scope_denied`、`draft_scope_grant_missing`、`draft_auth_context_contract_mismatch`、`draft_audit_context_missing`

所有失败都必须使用 sanitized envelope，不返回 raw token、raw claim、authorization header、cookie、client secret、JWKS、database detail 或内部 adapter detail。

## 验收方式

- 新增 `radish-oidc-token-membership-readiness-v1` fixture / checker，固定 token validation contract、membership contract、consumer matrix、failure mapping、no fallback、no side effects 和 artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`。
- 本批至少运行专项 checker、upstream auth readiness checker、saved draft production auth runtime checker 和 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建 OIDC middleware、token validator、membership adapter、login / logout route、session cookie、token validation schema、production API consumer 或 auth runtime smoke。
- 不启用 repository mode，不创建真实 query executor、数据库连接、schema marker runtime、migration runner、SQL、secret resolver、audit store runtime 或 production API。
- 不实现 workflow executor、publish、run、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 `radish_oidc_token_membership_readiness_defined` 写成 Radish OIDC ready、token validation ready、membership adapter ready、repository mode ready、production API ready 或 production ready。
