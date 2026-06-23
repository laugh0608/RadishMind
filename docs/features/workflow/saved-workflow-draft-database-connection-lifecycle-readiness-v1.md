# Saved Workflow Draft Database Connection Lifecycle Readiness v1

更新时间：2026-06-23

## 专题定位

`Saved Workflow Draft Database Connection Lifecycle Readiness v1` 承接 [Saved Workflow Draft Database Connection Smoke Strategy v1](saved-workflow-draft-database-connection-smoke-strategy-v1.md)，只定义 future saved draft database connection provider 之前的连接生命周期策略。

结论：状态为 `draft_database_connection_lifecycle_readiness_defined`。本批只固定 future timeout budget、pool policy、health check boundary、close responsibility、request id / audit ref propagation、sanitized diagnostics runtime 前置、failure mapping、no fallback、no side effects 和 artifact guard；不创建 connection lifecycle runtime、connection factory、database connection provider、DB driver import、DSN parser runtime、pool runtime、health check runtime、connection smoke runtime、SQL、schema marker、migration runner、repository mode runtime、OIDC middleware、token validator、membership adapter、production API、executor、confirmation、writeback 或 replay。

## 输入证据

- `workflow-saved-draft-database-connection-smoke-strategy-v1` 已定义 explicit test database boundary、safe placeholder credential handoff、smoke input / output record shape、role denial cases、no leakage scan 和 manual-only execution boundary，但不创建 connection smoke runtime。
- `workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1` 已定义 driver / DSN / TLS policy、DSN redaction、environment binding 和 forbidden diagnostics scan 的 future 边界。
- `workflow-saved-draft-database-role-policy-readiness-v1` 已定义 runtime DML role、migration DDL / marker role、least privilege review 和 cross-environment denial smoke 前置。
- `workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1` 已确认 connection provider implementation task card 当前仍为 `blocked_before_implementation_task_card`。
- `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 已确认 secret resolver implementation 仍 blocked。
- `workflow-saved-draft-repository-mode-runtime-boundary-review-v1` 已确认 repository mode runtime task card 仍 blocked。
- `radish-oidc-token-membership-implementation-entry-review-v1` 已确认 OIDC middleware、token validator 和 membership adapter 仍 blocked。

## Readiness Decision

| lifecycle area | 本次结论 | 阻塞或停止线 |
| --- | --- | --- |
| timeout budget | `defined_static_policy_no_runtime_timer` | 只定义 future connect / acquire / query / health check timeout budget；本批不创建 timer、context wrapper 或 driver call |
| pool policy | `defined_static_policy_no_pool_runtime` | 只定义 future max open / idle / lifetime / per-role pool separation；本批不创建 pool runtime 或 connection factory |
| health check boundary | `defined_no_health_check_runtime` | 只定义 startup、preflight、request path 和 manual smoke 的 health check 禁止 / 允许边界；本批不执行 health check |
| close responsibility | `defined_static_ownership_no_connection` | 只定义 future owner、shutdown、request cancellation 和 leak handling；本批不打开连接也不关闭连接 |
| request / audit propagation | `defined_metadata_only_no_runtime_binding` | future lifecycle diagnostics 必须携带 request id、audit ref、policy version 和 purpose；本批不创建 runtime binding |
| sanitized diagnostics runtime prerequisite | `defined_no_diagnostics_runtime` | 只定义 failure code、policy ids、redacted target / role fingerprint 和 lifecycle phase allowlist；本批不创建 diagnostics runtime |
| connection provider dependency | `blocked_before_implementation_task_card` | provider task card 仍需等待 lifecycle、secret resolver、schema marker、repository mode runtime 和 auth upstream evidence 可复验 |

## Lifecycle Boundary

future connection provider 必须遵守以下生命周期边界：

- timeout budget 必须按 lifecycle phase 分离：`resolve_credential`、`build_connection_config`、`open_or_acquire_connection`、`schema_marker_read`、`repository_query`、`health_check` 和 `close_or_release` 不能共享一个不可审计的总超时。
- pool policy 必须按 environment、role class、tenant / workspace scope 和 policy version 分离；runtime DML role 和 migration DDL / marker role 不得复用同一个 pool。
- health check 不得在 service startup、HTTP request path 或 store selector 中隐式连接数据库；future health check 只能由独立 runtime task card、manual smoke 或显式 readiness command 创建。
- close responsibility 必须由 connection provider runtime 持有，repository adapter 只能消费受控 query executor / handle reference，不能持有 raw driver connection、raw pool 或 close function。
- request id、audit ref、caller purpose、policy version 和 lifecycle phase 必须贯穿 future diagnostics；缺失时必须 fail closed。
- sanitized diagnostics 只能输出 lifecycle phase、failure code、policy id、environment id、redacted target fingerprint、redacted role fingerprint、request id 和 audit ref，不得输出 raw DSN、database host、username、password、secret value、credential payload、driver error detail、SQL detail、schema marker payload 或草案主体。

## Failure Mapping

| 条件 | failure code |
| --- | --- |
| repository mode reserved / runtime 未打开 | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| timeout policy 缺失、phase 未定义或 timeout budget 不可信 | `draft_store_unavailable` |
| pool policy 缺失、role pool 混用或 environment mismatch | `draft_store_unavailable` |
| health check boundary 缺失或隐式 startup / request health check 被要求 | `draft_store_unavailable` |
| close owner 缺失、request cancellation 不可观测或 leak handling 不可信 | `draft_store_unavailable` |
| request id / audit ref / purpose / policy version 缺失 | `draft_audit_context_missing` |
| lifecycle diagnostics allowlist 缺失或可能泄露 raw DB / secret material | `draft_store_unavailable` |
| marker 缺失或 migration 未应用 | `draft_schema_migration_not_applied` |
| marker version mismatch | `draft_store_schema_version_mismatch` |
| marker 不可读、runner 状态不可观测或 lock 状态不可信 | `draft_store_migration_unavailable` |
| stored record contract mismatch | `draft_store_contract_mismatch` |
| repository actor context 不可信 | `draft_auth_context_contract_mismatch` |

失败时不得返回 raw DSN、database host、database error detail、secret value、credential payload、完整 secret ref、完整 credential handle、完整 user claim、authorization header、cookie、SQL detail、schema marker payload、pool state dump、driver connection object 或草案主体。

## No Fallback / No Side Effects

- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 因本 readiness 被启用。
- 不允许 service startup、HTTP route、store selector 或 saved draft request path 自动解析 secret、打开 driver、创建 pool、连接数据库、执行 health check、读取 marker 或写 repository。
- 不允许 test-only fake resolver runtime、fake query executor smoke、static schema artifact、rollback evidence、driver / DSN / TLS policy、role policy、connection smoke strategy 或 container smoke 替代 lifecycle runtime 后续实现前置。
- 本批 checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不读取 env secret、不调用 fake resolver、不调用 production resolver、不调用云 secret 服务、不导入 driver、不创建 DSN parser、不创建 pool policy runtime、不创建 connection factory、不连接数据库、不执行 health check、不运行 connection smoke、不运行 SQL、不读写 schema marker、不创建 runner、不调用 OIDC、不校验 token、不创建 membership adapter、不创建 production API、不执行 workflow。
- side effect counters 必须保持 `secret_resolver_call_count=0`、`fake_resolver_call_count=0`、`production_resolver_call_count=0`、`secret_value_read_count=0`、`credential_handle_created_count=0`、`cloud_secret_call_count=0`、`network_call_count=0`、`database_connection_count=0`、`driver_import_count=0`、`driver_open_count=0`、`dsn_parse_count=0`、`dsn_build_count=0`、`tls_config_create_count=0`、`pool_create_count=0`、`pool_acquire_count=0`、`connection_health_check_count=0`、`connection_close_count=0`、`sanitized_diagnostics_runtime_count=0`、`connection_smoke_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`repository_mode_enablement_count=0`、`repository_read_count=0`、`repository_write_count=0`、`oidc_call_count=0`、`token_validation_count=0`、`membership_lookup_count=0`、`production_api_call_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0` 和 `replay_call_count=0`。

## 后续方向

本批后，connection provider implementation task card 仍不创建。后续若继续 durable store 上游，应从以下方向选择一个独立推进：

1. `Radish OIDC upstream evidence refresh` 已固定为 `radish_oidc_token_membership_upstream_evidence_refresh_defined`；后续 auth runtime 仍需独立 entry review。
2. `connection provider implementation entry refresh v2` 已固定为 `draft_database_connection_provider_implementation_entry_refresh_v2_defined`；connection lifecycle 只满足静态 readiness，不打开 provider task card。
3. `schema marker runtime dependency refresh`：只复评 marker reader / writer、manual runner 和 connection lifecycle 的依赖，不创建 marker runtime 或 SQL。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-lifecycle-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-smoke-strategy-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-database-connection-provider-implementation-v1-plan.md`。
- 不创建 connection lifecycle runtime、connection factory、database connection provider、DB driver import、DSN parser runtime、pool runtime、health check runtime、connection smoke runtime、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command、dry-run output 或 smoke output。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_database_connection_lifecycle_readiness_defined` 解释为 connection lifecycle runtime ready、connection provider ready、database ready、secret resolver ready、repository mode ready、durable persistence ready、production API ready 或 production ready。
