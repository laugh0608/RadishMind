# Saved Workflow Draft Database Connection Provider Implementation Entry Refresh v2

更新时间：2026-06-23

## 专题定位

`Saved Workflow Draft Database Connection Provider Implementation Entry Refresh v2` 承接 [Saved Workflow Draft Database Connection Lifecycle Readiness v1](saved-workflow-draft-database-connection-lifecycle-readiness-v1.md)，用于在 driver / DSN / TLS、role policy、connection smoke strategy 和 connection lifecycle readiness 已补齐后，重新评审 future saved draft database connection provider 是否可以进入实现任务卡。

结论：状态为 `draft_database_connection_provider_implementation_entry_refresh_v2_defined`。entry decision 仍为 `blocked_before_implementation_task_card`。本批只把 driver / DSN / TLS、role policy、connection smoke strategy 和 connection lifecycle 从“缺口”更新为“静态 readiness 已满足但无 runtime”，并继续固定 secret resolver、production resolver、credential handle、operator approval、audit store、backend health、no leakage smoke runtime、schema marker runtime 和 repository mode runtime 仍未满足；后续 `radish_oidc_token_membership_upstream_evidence_refresh_defined` 只补 OIDC 上游静态证据形状，仍不打开 auth runtime。不创建 connection provider implementation task card、database connection provider、secret resolver、DB driver、DSN parser、connection factory、pool runtime、health check runtime、role policy runtime、connection smoke runner、query executor、schema marker、SQL migration、migration runner、repository mode runtime、OIDC middleware、membership adapter 或 production API。

## 输入证据

- `workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1` 已确认 provider task card 在 driver / role / smoke / lifecycle 缺口存在时保持 blocked。
- `workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1` 已固定 future driver selection、DSN construction / redaction、TLS policy、environment binding 和 forbidden diagnostics scan，但不选择 driver、不创建 parser 或 TLS runtime。
- `workflow-saved-draft-database-role-policy-readiness-v1` 已固定 runtime DML role、migration DDL / marker role、least privilege review、role metadata shape 和 cross-environment denial smoke 前置，但不创建 role policy runtime 或数据库 role。
- `workflow-saved-draft-database-connection-smoke-strategy-v1` 已固定 explicit test database、placeholder credential handoff、smoke record shape、role denial cases、no leakage scan 和 manual-only boundary，但不创建 smoke runner 或 smoke output。
- `workflow-saved-draft-database-connection-lifecycle-readiness-v1` 已固定 timeout budget、pool policy、health check boundary、close responsibility、request / audit propagation 和 sanitized diagnostics runtime 前置，但不创建 lifecycle runtime。
- `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1`、production secret backend real resolver / credential handle / no leakage runtime entry review 仍保持 blocked。
- `workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1`、`workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1` 和 `workflow-saved-draft-repository-mode-runtime-boundary-review-v1` 仍确认 marker runtime、manual runner 和 repository mode runtime 未打开。
- `radish-oidc-token-membership-implementation-entry-review-v1` 仍确认 OIDC middleware、token validator 和 membership adapter 未打开；`radish-oidc-token-membership-upstream-evidence-refresh-v1` 已补静态 upstream evidence contract，但不创建 runtime。

## Entry Refresh Decision

| candidate | 本次结论 | 剩余阻塞或停止线 |
| --- | --- | --- |
| driver / DSN / TLS policy | `static_readiness_satisfied_no_runtime` | 已有静态策略；仍不选择 driver、不改 `go.mod`、不 import driver、不创建 DSN parser 或 TLS runtime |
| runtime / migration role policy | `static_readiness_satisfied_no_runtime` | 已有静态 role policy；仍不创建 runtime role、migration role、grant、policy runtime 或 denial smoke |
| connection smoke strategy | `static_strategy_satisfied_no_runtime` | 已有 smoke contract；仍不创建 smoke runner、不连接数据库、不生成 smoke output |
| connection lifecycle | `static_readiness_satisfied_no_runtime` | 已有 timeout / pool / health / close / diagnostics 前置；仍不创建 lifecycle runtime、pool runtime 或 health check runtime |
| secret resolver handoff | `blocked` | production real resolver、credential handle runtime、operator approval、audit store、backend health 和 no leakage smoke runtime 仍未满足 |
| schema marker runtime dependency | `blocked` | schema marker reader / writer、schema version table、manual runner 和 marker smoke 均未实现 |
| repository mode runtime dependency | `blocked` | repository mode 仍 fail closed，不能通过 provider refresh 打开成功路径 |
| auth upstream evidence | `static_contract_defined_no_runtime` | reviewed issuer、JWKS、client registration、auth middleware ownership、membership data source ownership 和 negative auth smoke matrix 已有静态契约；token validation schema、auth middleware、membership adapter 和 runtime smoke 仍未创建 |
| repository query executor handoff | `blocked` | provider、schema marker、auth / membership 和 repository mode runtime 均未满足 |
| connection provider implementation task card | `blocked_before_implementation_task_card` | 只完成静态策略补齐，不能创建 provider 实现任务卡 |

## Provider Contract Delta

v2 相比 v1 的变化只在评审口径：

- `driver_dsn_tls_policy`、`role_policy`、`connection_smoke` 和 `connection_lifecycle` 从 `blocked` 更新为静态 readiness / strategy satisfied，但这些状态都不等价于 runtime ready。
- `secret_resolver_handoff`、`schema_marker_runtime_dependency`、`repository_mode_runtime_dependency`、`auth_upstream_evidence` 和 `repository_query_executor_handoff` 继续 blocked。
- connection provider implementation task card 继续不得创建；任何后续实现任务卡都必须同时消费 secret resolver runtime、auth upstream evidence、schema marker runtime、repository mode runtime 和 sanitized diagnostics runtime 的可复验证据。

future connection provider 仍必须保持 metadata-only contract：输入只能消费 environment、secret ref 或 opaque credential handle、required role、request id、audit ref、caller purpose 和 policy version；输出只能是 readiness / handle reference 或 sanitized failure metadata，不得返回 raw DSN、secret value、database host、database error detail、credential payload、driver object、pool state、SQL detail、完整 user claim 或草案主体。

## Failure Mapping

entry refresh v2 继续固定 fail-closed 语义：

| 条件 | failure code |
| --- | --- |
| repository mode reserved / runtime 未打开 | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| secret ref 缺失、resolver disabled、real resolver 不可用或 credential handle 缺失 | `draft_store_unavailable` |
| DB driver / DSN / TLS policy 缺失或 runtime 未创建 | `draft_store_unavailable` |
| runtime / migration role policy runtime 缺失或角色混用 | `draft_store_unavailable` |
| connection smoke strategy 缺失、smoke record 不可信或 no leakage scan runtime 未创建 | `draft_store_unavailable` |
| connection lifecycle runtime 缺失、pool / health / close / diagnostics 不可信 | `draft_store_unavailable` |
| request id / audit ref / purpose / policy version 缺失 | `draft_audit_context_missing` |
| marker 缺失或 migration 未应用 | `draft_schema_migration_not_applied` |
| marker version mismatch | `draft_store_schema_version_mismatch` |
| marker 不可读、runner 状态不可观测或 lock 状态不可信 | `draft_store_migration_unavailable` |
| stored record contract mismatch | `draft_store_contract_mismatch` |
| repository actor context 不可信 | `draft_auth_context_contract_mismatch` |

失败时不得返回草案主体，不得回退到 `memory_dev`、sample、fixture、dev HTTP route、test auth、fake query executor、test-only fake resolver runtime、developer env plaintext 或 raw DB diagnostics。

## No Fallback / No Side Effects

- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 因本 refresh 被启用。
- 不允许 service startup、HTTP route、store selector 或 saved draft request path 自动解析 secret、选择 driver、构造 DSN、创建 pool、执行 health check、连接数据库、读取 marker 或写 repository。
- 不允许 driver / DSN / TLS policy、role policy、connection smoke strategy、connection lifecycle readiness、test-only fake resolver runtime、fake query executor smoke、static schema artifact、rollback evidence 或 container smoke 替代 provider implementation prerequisites。
- 本批 checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不读取 env secret、不调用 fake resolver、不调用 production resolver、不调用云 secret 服务、不连接数据库、不导入 driver、不运行 SQL、不读写 schema marker、不创建 runner、不调用 OIDC、不校验 token、不创建 membership adapter、不创建 production API、不执行 workflow。
- side effect counters 必须保持 `secret_resolver_call_count=0`、`fake_resolver_call_count=0`、`production_resolver_call_count=0`、`secret_value_read_count=0`、`credential_handle_created_count=0`、`cloud_secret_call_count=0`、`network_call_count=0`、`database_connection_count=0`、`driver_import_count=0`、`driver_open_count=0`、`dsn_parse_count=0`、`dsn_build_count=0`、`tls_config_create_count=0`、`role_policy_evaluation_count=0`、`pool_create_count=0`、`connection_health_check_count=0`、`connection_smoke_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`repository_mode_enablement_count=0`、`oidc_call_count=0`、`token_validation_count=0`、`membership_lookup_count=0`、`production_api_call_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0` 和 `replay_call_count=0`。

## 后续方向

本次 refresh 后，connection provider implementation task card 仍不创建。后续若继续 durable store 上游，应从以下方向选择一个独立推进：

1. `schema marker runtime dependency refresh`：复评 marker reader / writer、manual runner、connection lifecycle 和 repository mode dependency，但不创建 marker runtime 或 SQL。
2. `secret resolver runtime dependency refresh`：复评 production resolver、credential handle、approval、audit、backend health 和 no leakage runtime 是否足以打开 resolver task card。
3. `token validation schema / auth middleware runtime entry review`：若继续 auth 路线，应先复验 upstream evidence contract，再评审 token schema、middleware、membership adapter 和 runtime smoke 是否可以拆任务卡。

如果上述依赖仍 blocked，connection provider implementation task card、manual migration runner implementation task card、schema marker contract implementation task card 和 repository mode runtime implementation task card 继续保持不创建。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-lifecycle-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-smoke-strategy-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-database-connection-provider-implementation-v1-plan.md`。
- 不创建 database connection provider、secret resolver、DB driver、DSN parser、connection factory、pool runtime、health check runtime、role policy runtime、connection smoke runner、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_database_connection_provider_implementation_entry_refresh_v2_defined` 解释为 connection provider ready、database ready、secret resolver ready、schema marker ready、repository mode ready、durable persistence ready、production API ready 或 production ready。
