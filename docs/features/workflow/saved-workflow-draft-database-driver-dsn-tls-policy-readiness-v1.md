# Saved Workflow Draft Database Driver / DSN / TLS Policy Readiness v1

更新时间：2026-06-22

## 专题定位

`Saved Workflow Draft Database Driver / DSN / TLS Policy Readiness v1` 承接 [Saved Workflow Draft Database Connection Provider Implementation Entry Refresh v1](saved-workflow-draft-database-connection-provider-implementation-entry-refresh-v1.md)，只定义 future saved draft database connection provider 之前的 driver、DSN、TLS 和环境绑定策略。

结论：状态为 `draft_database_driver_dsn_tls_policy_readiness_defined`。本批只固定 future DB driver selection、DSN construction / redaction boundary、TLS policy、environment binding、forbidden diagnostics scan、role policy dependency 和 connection smoke 前置；不创建 DB provider、DB driver import、DSN parser runtime、connection factory、connection smoke、SQL、schema marker、migration runner、repository mode runtime、OIDC middleware、membership adapter、production API、executor、confirmation、writeback 或 replay。

## 输入证据

- `workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1` 已确认 connection provider implementation task card 当前仍为 `blocked_before_implementation_task_card`。
- `workflow-saved-draft-database-secret-resolver-readiness-v1` 与 `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 已固定 secret ref、resolver result、disabled runtime 和 failure taxonomy，但 resolver implementation 仍不创建。
- `workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1` 与 `workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1` 已确认 runner、marker reader / writer 和 schema version table 仍 blocked。
- `workflow-saved-draft-repository-mode-runtime-boundary-review-v1` 已确认 repository mode runtime task card 仍 blocked，不能通过本批绕过 database、auth、marker、audit 或 production API 前置。
- `radish-oidc-token-membership-implementation-entry-review-v1` 已确认 OIDC middleware、token validator 和 membership adapter 仍 blocked。
- `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1` 已确认 no leakage smoke runtime 仍 blocked；本批只能定义 diagnostics scan 的静态边界，不能执行 smoke runtime。

## Readiness Decision

| policy area | 本次结论 | 阻塞或停止线 |
| --- | --- | --- |
| future DB driver selection | `defined_for_future_selection` | 只允许后续实现任务卡选择 driver family；本批不选择具体 driver、不修改 `go.mod`、不 import driver |
| DSN construction boundary | `defined_no_parser_runtime` | future DSN 只能由 opaque credential handle、environment、role 和 TLS policy 组合生成；本批不创建 parser、builder 或 raw DSN |
| DSN redaction boundary | `defined_static_policy` | diagnostics、fixture、audit、response 和 smoke record 只能输出 redacted fingerprint / policy id，不输出 raw DSN、host、username、password 或 credential payload |
| TLS policy | `defined_no_runtime_tls` | production future default 必须要求 verified TLS；显式 test 环境才允许 test profile；本批不创建 TLS config runtime |
| environment binding | `defined_no_runtime_binding` | future environment、secret ref environment、credential handle environment、role policy environment 和 smoke environment 必须一致；本批不读取 env secret |
| forbidden diagnostics scan | `defined_static_scan_only` | 只定义未来 scanner 需要覆盖的 surfaces 和 forbidden terms；本批不创建 smoke runner 或 runtime scanner |
| role policy dependency | `blocked_before_role_policy_readiness` | runtime DML role 与 migration DDL / marker role 仍需独立 role policy readiness |
| connection smoke precondition | `blocked_before_connection_smoke_strategy` | explicit test database、placeholder credential handoff、no leakage runtime 和 zero-side-effect smoke 仍需独立 strategy |

## Policy Boundary

future connection provider 必须遵守以下策略：

- driver selection 只能由后续 implementation task card 明确选择，且必须包含 driver family、version pin、license / security review、failure mapping 和 forbidden fallback。
- DSN construction 只能在 connection provider runtime 内部发生，输入只能是 `environment`、`secret_ref` 或 opaque credential handle、`required_role`、`tls_policy_ref`、`policy_version`、`request_id` 和 `audit_ref`。
- DSN 不得进入 HTTP response、fixture、audit event、diagnostics、task card、devlog、smoke record 或 exception detail；允许的替代输出只有 redacted fingerprint、policy id、environment id、role id 和 sanitized failure code。
- TLS policy 必须按 environment fail closed：production / test-prod 必须 verified TLS，local test 必须显式声明 test profile，未知环境或 TLS policy 缺失返回 `draft_store_unavailable`。
- environment binding mismatch 必须 fail closed，不得回退 `memory_dev`、sample、fixture、developer env plaintext、test-only fake resolver runtime 或 fake query executor。
- forbidden diagnostics scan 后续必须覆盖 docs / fixtures / smoke output / audit event / logs / response envelope，但本批只定义扫描要求，不创建 scanner 或 smoke runner。

## Failure Mapping

| 条件 | failure code |
| --- | --- |
| repository mode reserved / runtime 未打开 | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| driver policy 缺失、driver 未选择、driver import 被禁止 | `draft_store_unavailable` |
| DSN construction policy 缺失、DSN redaction policy 缺失 | `draft_store_unavailable` |
| TLS policy 缺失、不被允许或 environment mismatch | `draft_store_unavailable` |
| role policy 未定义、runtime / migration role 混用或 cross-environment role mismatch | `draft_store_unavailable` |
| connection smoke 前置缺失或 no leakage smoke runtime 未创建 | `draft_store_unavailable` |
| marker 缺失或 migration 未应用 | `draft_schema_migration_not_applied` |
| marker version mismatch | `draft_store_schema_version_mismatch` |
| marker 不可读、runner 状态不可观测或 lock 状态不可信 | `draft_store_migration_unavailable` |
| stored record contract mismatch | `draft_store_contract_mismatch` |
| repository actor context 不可信 | `draft_auth_context_contract_mismatch` |

失败时不得返回 raw DSN、database host、database error detail、secret value、credential payload、完整 secret ref、完整 credential handle、完整 user claim、authorization header、cookie、SQL detail 或草案主体。

## No Fallback / No Side Effects

- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 因本 readiness 被启用。
- 不允许 service startup、HTTP route、store selector 或 saved draft request path 自动选择 driver、构造 DSN、读取 secret、打开 TLS config、连接数据库、读取 marker 或写 repository。
- 不允许 test-only fake resolver runtime、fake query executor smoke、static schema artifact、rollback evidence 或 container smoke 替代 driver / DSN / TLS policy readiness 后续实现前置。
- 本批 checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不读取 env secret、不调用 fake resolver、不调用 production resolver、不调用云 secret 服务、不导入 driver、不创建 DSN parser、不连接数据库、不运行 connection smoke、不运行 SQL、不读写 schema marker、不创建 runner、不调用 OIDC、不校验 token、不创建 membership adapter、不创建 production API、不执行 workflow。
- side effect counters 必须保持 `secret_resolver_call_count=0`、`fake_resolver_call_count=0`、`production_resolver_call_count=0`、`secret_value_read_count=0`、`credential_handle_created_count=0`、`cloud_secret_call_count=0`、`network_call_count=0`、`database_connection_count=0`、`driver_import_count=0`、`driver_open_count=0`、`dsn_parse_count=0`、`dsn_build_count=0`、`tls_config_create_count=0`、`connection_smoke_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`repository_mode_enablement_count=0`、`repository_read_count=0`、`repository_write_count=0`、`oidc_call_count=0`、`token_validation_count=0`、`membership_lookup_count=0`、`production_api_call_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0` 和 `replay_call_count=0`。

## 后续方向

本批后，connection provider implementation task card 仍不创建。后续若继续 durable store 上游，应从以下方向选择一个独立推进：

1. `database role policy readiness`：定义 runtime DML role、migration DDL / marker role、least privilege review 和 cross-environment denial smoke。
2. `connection smoke strategy`：定义 explicit test database、placeholder credential handoff、smoke output shape、no leakage scan 和 zero production side effect 记录。
3. `Radish OIDC upstream evidence refresh`：补 reviewed issuer、JWKS、client registration 和 membership data source ownership。
4. `connection provider implementation entry refresh v2`：只有 driver / DSN / TLS、role policy、connection smoke、secret resolver、schema marker 和 repository mode runtime 依赖均可复验后再复评。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-implementation-entry-review-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-database-connection-provider-implementation-v1-plan.md`。
- 不创建 database connection provider、DB driver import、DSN parser runtime、connection factory、connection smoke、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_database_driver_dsn_tls_policy_readiness_defined` 解释为 driver ready、DSN ready、TLS runtime ready、connection provider ready、database ready、secret resolver ready、repository mode ready、durable persistence ready、production API ready 或 production ready。
