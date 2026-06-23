# Saved Workflow Draft Token Validation Schema / Auth Middleware Runtime Entry Review v1

更新时间：2026-06-23

## 专题定位

`Saved Workflow Draft Token Validation Schema / Auth Middleware Runtime Entry Review v1` 承接 Radish OIDC upstream evidence refresh、Saved Workflow Draft production auth runtime bridge、repository mode / schema marker / database secret resolver 依赖 refresh，用于评审 future token validation schema、auth middleware、membership adapter、negative auth smoke 与 repository actor context handoff 是否可以进入实现任务卡。

结论：状态为 `draft_token_validation_auth_middleware_runtime_entry_review_defined`。entry decision 为 `token_validation_auth_middleware_runtime_still_blocked_before_implementation_task_card`。本批只固定 runtime entry review 的静态矩阵、失败语义和 artifact guard；不创建 token validation schema、OIDC middleware、token validator、membership adapter、negative auth smoke runtime、repository mode runtime、database runtime 或 production API。

## 输入证据

- `radish_oidc_token_membership_readiness_defined` 已固定 token validation contract、membership contract、consumer matrix、failure mapping、no fallback 和 no side effects。
- `radish_oidc_token_membership_implementation_entry_review_defined` 已确认 token validation schema、auth middleware、membership adapter、negative auth smoke 和 production API consumer binding 仍为 `blocked_before_task_card`。
- `radish_oidc_token_membership_upstream_evidence_refresh_defined` 已定义 reviewed issuer evidence、JWKS pin / refresh policy、client registration evidence、auth middleware ownership、membership source ownership、membership cache policy 和 negative auth smoke matrix 的静态证据形状。
- `draft_production_auth_runtime_bridge_implemented` 已实现 verified auth context + workspace binding 到 repository actor context 的 bridge，但它不解析 token、不 fetch JWKS、不查询 membership。
- `draft_database_secret_resolver_runtime_dependency_refresh_defined`、`draft_schema_marker_runtime_dependency_refresh_defined`、`draft_database_connection_provider_implementation_entry_refresh_v2_defined` 和 `draft_repository_mode_runtime_boundary_review_defined` 均确认 repository mode、DB runtime、schema marker runtime、secret resolver runtime 和 auth runtime 依赖仍未满足。

## Entry Review Decision

| review item | 本次结论 | 说明 |
| --- | --- | --- |
| token validation schema | `blocked_schema_contract_reviewed_no_schema_file` | 只固定 future schema 字段和禁止字段，不创建 `contracts/radish-oidc-token-validation.schema.json` |
| auth middleware ownership | `blocked_owner_reviewed_no_middleware` | 只确认 route owner、failure envelope owner、audit owner 与 dev fake auth 隔离边界 |
| membership adapter | `blocked_contract_reviewed_no_adapter` | 只确认 tenant / workspace / application / owner membership source 和 cache policy owner |
| negative auth smoke | `blocked_matrix_reviewed_no_runtime_smoke` | 只复用 upstream negative auth smoke matrix，不创建 fixture、runner 或 runtime checker |
| repository actor context handoff | `static_handoff_reviewed_existing_bridge_only` | 只消费 existing bridge，不创建新的 token-to-context runtime |
| failure mapping | `fail_closed_mapping_reviewed_no_runtime` | 继续映射到 Saved Workflow Draft 现有 failure code |
| no fallback | `no_fallback_reviewed_required` | 不允许回退 dev auth、memory dev、fixture、sample、fake query executor 或 test-only resolver |
| no side effects | `no_side_effects_reviewed_required` | 不启动服务、不调用 OIDC、不校验 token、不查询 membership、不连接数据库 |
| artifact guard | `artifact_guard_reviewed_required` | 禁止 token schema、middleware、validator、adapter、runtime smoke 和 production API artifact 提前出现 |

## Token Validation Schema Review

future token validation schema 必须只描述已验证 token 的脱敏输出，不保存或输出 raw token。最小字段包括：

- `issuer_ref`
- `subject_ref`
- `tenant_ref`
- `audience_refs`
- `scope_grants`
- `workspace_binding_refs`
- `application_scope_refs`
- `owner_subject_ref`
- `key_id_ref`
- `algorithm`
- `issued_at`
- `expires_at`
- `auth_time`
- `policy_version`
- `request_id`
- `audit_ref`

禁止字段包括 `raw_token`、`authorization_header`、`cookie`、`client_secret`、`refresh_token`、`authorization_code`、`jwks_raw_dump`、`raw_claim_dump`、`membership_raw_record` 和 `database_detail`。schema 创建前仍不得让 runtime route 接受 production bearer token。

## Auth Middleware Ownership

future auth middleware owner 必须明确：

- route owner 只负责把已验证上下文交给 consumer，不直接写 repository、database 或 production API。
- token validator owner 只负责 issuer / audience / signature / time / claim 校验，不查询 membership。
- failure envelope owner 只输出 sanitized diagnostics，不返回 provider detail、raw claim 或 token。
- dev fake auth isolation owner 负责保证 dev-only fake context 不进入 public production route。
- audit context owner 负责 `request_id`、`audit_ref` 和 policy version 透传。

本批不创建 middleware 文件、route binding、token validator、session cookie、login / logout route 或 public fake auth。

## Membership Adapter Review

future membership adapter 必须把 tenant binding、workspace membership、application scope、owner scope 和 scope grant projection 分开实现，并输出可被 existing production auth runtime bridge 消费的 `SavedWorkflowDraftWorkspaceBinding` 与 scope grants。membership cache policy 只能缓存脱敏引用和过期策略，不保存 raw membership record 或 full claim。

当前结论仍为 blocked：membership source ownership 是静态证据，不能替代 adapter、cache runtime、membership query 或 negative auth smoke runtime。

## Repository Actor Context Handoff

本批只允许消费既有 `BuildSavedWorkflowDraftRepositoryActorContextFromProductionAuth` bridge。future runtime 必须先完成 token validation schema、auth middleware、membership adapter 和 negative auth smoke，再把 verified auth context、workspace binding、operation scope、request id 和 audit ref 交给 repository actor context。

失败时必须在 repository adapter 前 fail closed，不能打开 query executor、connection provider、schema marker、secret resolver 或 repository mode runtime。

## Negative Auth Smoke Matrix

future runtime task card 至少覆盖：

- missing / malformed authorization header
- unknown issuer、JWKS unavailable、invalid signature、expired token、missing required claim
- tenant binding denied
- workspace membership denied
- application scope denied
- owner scope denied
- scope grant missing
- membership source unavailable

这些用例目前只进入静态 matrix 和 checker；不创建 runtime smoke fixture、runner、HTTP route smoke 或 production auth runtime smoke。

## Failure Mapping

| 条件 | failure code |
| --- | --- |
| token schema、validator、middleware、issuer / JWKS evidence 不满足 | `draft_auth_context_contract_mismatch` |
| subject 缺失 | `draft_identity_context_missing` |
| tenant binding 缺失或不一致 | `draft_tenant_binding_missing` |
| workspace membership denied | `draft_workspace_membership_denied` |
| application scope denied | `draft_application_scope_denied` |
| owner scope denied | `draft_owner_scope_denied` |
| scope grant missing | `draft_scope_grant_missing` |
| request id / audit ref 缺失 | `draft_audit_context_missing` |
| repository mode 仍未启用 | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| membership source、DB provider、schema marker、secret resolver 或 query executor 不可用 | `draft_store_unavailable` 或 `draft_store_migration_unavailable` |

所有失败输出必须使用 sanitized diagnostics，不得返回 raw token、authorization header、cookie、完整 claim、membership raw record、secret、DSN、provider URL、database detail 或草案主体。

## No Fallback / No Side Effects

- 不允许 `repository` store mode 因本 review 被启用。
- 不允许 auth 缺失时回退 `memory_dev`、sample、fixture、dev route、test auth、fake query executor、test-only fake resolver 或 developer env plaintext。
- checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不调用 OIDC、不 fetch issuer discovery、不下载 JWKS、不校验 token、不查询 membership、不创建 membership cache、不连接数据库、不运行 SQL、不读写 schema marker、不解析 secret、不创建 production API、不执行 workflow。
- side effect counters 必须保持 `issuer_network_call_count=0`、`jwks_fetch_count=0`、`token_validation_call_count=0`、`auth_middleware_invocation_count=0`、`membership_query_count=0`、`membership_cache_create_count=0`、`repository_mode_enablement_count=0`、`database_connection_count=0`、`sql_execution_count=0`、`production_api_call_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0` 和 `replay_call_count=0`。

## 后续方向

本次 entry review 后，token validation schema、auth middleware、membership adapter、negative auth smoke runtime、repository mode runtime、database runtime 和 production API 仍不创建。后续若继续 auth 上游，应优先拆分 token validation schema implementation task card readiness 或 auth middleware / membership adapter task card entry readiness，并保持 runtime smoke 独立评审。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-runtime-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `contracts/radish-oidc-token-validation.schema.json`。
- 不创建 OIDC middleware、token validator、membership adapter、membership cache、negative auth smoke runtime、login / logout route、session cookie、repository mode runtime、database runtime、production API、executor、confirmation、writeback 或 replay。
- 不 fetch issuer discovery，不下载 JWKS，不校验 token，不查询 membership，不连接数据库，不运行 SQL，不读写 schema marker，不解析 secret。
- 不把 `draft_token_validation_auth_middleware_runtime_entry_review_defined` 解释为 auth ready、repository mode ready、durable persistence ready、production API ready 或 production ready。
