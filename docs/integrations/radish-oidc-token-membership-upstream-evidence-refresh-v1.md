# Radish OIDC Token / Membership Upstream Evidence Refresh v1

更新时间：2026-06-23

## 专题定位

`Radish OIDC Token / Membership Upstream Evidence Refresh v1` 承接 [Radish OIDC Token / Membership Implementation Entry Review v1](radish-oidc-token-membership-implementation-entry-review-v1.md)，用于把 future token validation / membership adapter 之前必须补齐的上游证据拆成可检查契约。

本专题只定义 reviewed issuer evidence、JWKS pin / refresh policy、client registration evidence、auth middleware ownership、membership data source ownership、membership cache policy 和 negative auth smoke matrix 的静态证据形状；不创建 OIDC client、token validation schema、token validator、auth middleware、membership adapter、negative auth smoke runtime、login / logout route、session cookie、repository mode runtime、数据库连接、production API consumer、workflow executor、confirmation、writeback 或 replay。

状态：`radish_oidc_token_membership_upstream_evidence_refresh_defined`

## 输入事实

- `radish_oidc_token_membership_readiness_defined` 已固定 token validation contract、membership contract、consumer matrix、failure mapping、no fallback 和 no side effects。
- `radish_oidc_token_membership_implementation_entry_review_defined` 已确认 token validation schema、auth middleware、membership adapter、negative auth smoke 和 production API consumer binding 仍为 `blocked_before_task_card`。
- 在线 `Radish` 仓库目前公开说明 `Radish.Auth` 是基于 OpenIddict 的 OIDC 认证服务器，但这只是上游观察，不是 RadishMind 可直接消费的 pinned issuer / JWKS / client registration 证据。
- control plane read 与 saved draft production auth runtime bridge 仍只消费已验证上下文，不校验 token、不查询 membership。

## Evidence Refresh Decision

结论：`blocked_before_runtime_task_card`。

本次 refresh 将下列阻塞项从“未定义要补什么”更新为“静态 evidence contract 已定义，但 runtime 仍未打开”：

| evidence | 当前结论 | runtime |
| --- | --- | --- |
| reviewed issuer evidence | `static_evidence_contract_defined_no_network` | 不 fetch discovery document，不调用 issuer |
| JWKS pin / refresh policy | `static_policy_defined_no_jwks_fetch` | 不下载 JWKS，不创建 cache |
| client registration evidence | `static_contract_defined_no_client_runtime` | 不创建 OIDC client，不保存 client secret |
| auth middleware ownership | `static_owner_defined_no_middleware` | 不创建 middleware、route binding 或 public fake auth |
| membership data source ownership | `static_owner_defined_no_adapter` | 不创建 membership adapter，不查询 membership |
| membership cache policy | `static_policy_defined_no_cache_runtime` | 不创建 membership cache 或 expiry runtime |
| negative auth smoke matrix | `static_matrix_defined_no_runtime_smoke` | 不创建 smoke fixture、runner 或 runtime checker |

## 证据字段

future reviewed issuer evidence 必须至少包含：

- `issuer_url`
- `discovery_document_url`
- `jwks_uri`
- `supported_signing_algorithms`
- `supported_scopes`
- `environment`
- `fetched_at`
- `expires_at`
- `operator_review_ref`
- `sanitized_evidence_ref`
- `policy_version`

future JWKS policy 必须至少包含：

- `jwks_uri_ref`
- `pin_set_ref`
- `refresh_interval_policy`
- `cache_expiry_policy`
- `rotation_failure_policy`
- `key_id_allowlist_ref`
- `operator_review_ref`

future client registration evidence 必须至少包含：

- `client_id_ref`
- `allowed_audiences`
- `allowed_scopes`
- `redirect_policy_ref`
- `environment`
- `operator_review_ref`
- `secret_ref_status`

future membership source ownership 必须至少包含：

- `source_owner_ref`
- `tenant_binding_owner_ref`
- `workspace_membership_owner_ref`
- `application_scope_owner_ref`
- `owner_scope_owner_ref`
- `cache_policy_ref`
- `audit_ref`

这些字段只能提交脱敏引用、策略版本和 review reference，不得提交 raw token、authorization header、cookie、client secret、raw JWKS dump、完整 claim dump、membership raw record 或数据库细节。

## Negative Auth Smoke Matrix

future runtime task card 至少要覆盖下列负向用例：

| case | failure boundary |
| --- | --- |
| missing authorization header | request_auth_boundary |
| malformed bearer header | request_auth_boundary |
| unknown issuer | issuer_boundary |
| JWKS unavailable | jwks_boundary |
| invalid signature | token_signature_boundary |
| expired token | token_time_boundary |
| missing required claim | token_claim_boundary |
| tenant binding denied | tenant_binding_boundary |
| workspace membership denied | workspace_membership_boundary |
| application scope denied | application_scope_boundary |
| owner scope denied | owner_scope_boundary |
| scope grant missing | scope_grant_boundary |
| membership source unavailable | membership_source_boundary |

所有失败输出必须使用 sanitized diagnostics，不得返回 raw token、完整 claim、membership record 或 provider detail。

## 后续要求

本次 refresh 后，auth upstream evidence 不再是“证据形状未知”的阻塞项，但以下 runtime 依赖仍未满足：

- token validation schema 尚未创建。
- auth middleware 尚未创建。
- membership adapter 尚未创建。
- negative auth smoke runtime 尚未创建。
- repository mode 仍 fail closed。
- production API consumer gate 仍未打开。

后续如果继续 auth 上游，应先创建独立的 token validation schema / auth middleware / membership adapter implementation entry review，而不是直接创建 runtime task card 或 production API consumer。

## 验收方式

- 新增 `radish-oidc-token-membership-upstream-evidence-refresh-v1` fixture / checker。
- checker 消费 readiness、implementation entry review、Radish OIDC client preconditions、control plane read production auth readiness、saved draft production auth readiness 和 saved draft production auth runtime bridge fixture。
- checker 固定 evidence contract matrix、ownership matrix、negative auth smoke matrix、failure mapping、diagnostic policy、no fallback、no side effects、future artifact guard 和 `check-repo.py` 注册顺序。
- checker 接入 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建 OIDC middleware、OIDC client、token validator、membership adapter、login / logout route、session cookie、negative auth smoke runtime 或 production auth runtime smoke。
- 不创建 token validation schema、runtime smoke fixture、runtime checker 或 production route binding。
- 不 fetch issuer discovery，不下载 JWKS，不校验 token，不查询 membership，不创建 membership cache。
- 不启用 repository mode，不创建真实 query executor、数据库连接、schema marker runtime、migration runner、SQL、secret resolver、audit store runtime 或 backend runtime。
- 不实现 workflow executor、publish、run、confirmation decision、writeback、replay、resume 或 materialized result reader。
