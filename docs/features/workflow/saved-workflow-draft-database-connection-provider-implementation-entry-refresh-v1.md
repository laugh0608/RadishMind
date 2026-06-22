# Saved Workflow Draft Database Connection Provider Implementation Entry Refresh v1

更新时间：2026-06-22

## 专题定位

`Saved Workflow Draft Database Connection Provider Implementation Entry Refresh v1` 承接 [Saved Workflow Draft Manual Migration Runner Implementation Entry Refresh v1](saved-workflow-draft-manual-migration-runner-implementation-entry-refresh-v1.md)，用于复评 future saved draft database connection provider 是否可以进入实现任务卡。

结论：状态为 `draft_database_connection_provider_implementation_entry_refresh_defined`。entry decision 仍为 `blocked_before_implementation_task_card`。本批只刷新 secret resolver dependency、test-only fake resolver 可用性、real resolver runtime blocker、driver / DSN / TLS policy、runtime / migration role policy、connection lifecycle、connection smoke 和 repository query executor handoff 的准入结论；不创建 connection provider implementation task card、database connection provider、secret resolver、DB driver、DSN parser、connection factory、role policy runtime、connection smoke、query executor、schema marker、SQL migration、migration runner、repository mode runtime、OIDC middleware、membership adapter 或 production API。

## 输入证据

- `workflow-saved-draft-database-connection-provider-implementation-entry-review-v1` 已确认 connection provider implementation 当前不打开。
- `workflow-saved-draft-database-secret-resolver-readiness-v1` 与 `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 已定义 resolver input / result、disabled runtime 和 failure taxonomy，但 secret resolver implementation task card 仍不创建。
- `production-secret-backend-fake-resolver-runtime-implementation-v1` 已实现 test-only、默认 disabled 的 fake resolver runtime；该 runtime 只用于离线测试，不能替代 production resolver、credential handle、no leakage smoke、DB provider 或 repository mode。
- `production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1` 已确认 production real resolver runtime implementation 仍 blocked before task card。
- `production-secret-backend-credential-handle-runtime-implementation-entry-review-v1` 与 `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1` 已确认 credential handle runtime 和 no leakage smoke runtime 仍 blocked before task card。
- `workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1` 已确认 manual migration runner task card 仍 blocked，且缺少 connection provider、database role 和 marker writer。
- `workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1` 与 `workflow-saved-draft-repository-mode-runtime-boundary-review-v1` 已确认 marker runtime 和 repository mode runtime 仍 blocked。

## Entry Refresh Decision

| candidate | 本次结论 | 阻塞原因 |
| --- | --- | --- |
| secret resolver handoff | `blocked` | test-only fake resolver 已实现但只能离线使用；production real resolver、credential handle runtime、operator approval、audit store、backend health 和 no leakage smoke runtime 仍未满足 |
| driver / DSN / TLS policy | `blocked` | 尚未选择 driver、DSN contract、TLS policy、environment binding 和 forbidden diagnostic scan |
| connection lifecycle | `blocked` | timeout、pool、health check、close responsibility、request id / audit ref propagation 和 sanitized diagnostics runtime 仍未实现 |
| runtime / migration role policy | `blocked` | runtime DML role、migration DDL / marker role、least privilege review 和 cross-environment denial smoke 仍缺失 |
| offline connection smoke | `blocked` | 没有显式 test database、safe placeholder credential handoff、driver policy、role policy、no secret leakage smoke 和 zero-side-effect fast baseline boundary |
| repository query executor handoff | `blocked` | connection provider、schema marker reader、OIDC token validation、membership adapter 和 repository mode runtime 均未打开 |
| connection provider implementation task card | `blocked_before_implementation_task_card` | 上述运行时依赖未满足，不能创建实现任务卡 |

后续若要打开 connection provider implementation task card，必须先独立满足 real resolver runtime 或被明确限制为离线 smoke 的 fake resolver handoff、driver / DSN / TLS policy、role separation、connection lifecycle、safe diagnostics、connection smoke 记录、schema marker read dependency 和 repository mode runtime dependency。

## Connection Provider Contract Refresh

future connection provider 必须保持以下边界：

- 只能由显式实现任务卡创建；service startup、HTTP route、store selector 或 saved draft request path 不得隐式创建连接。
- 输入只能消费 environment、secret ref 或 opaque credential handle、required role、request id、audit ref、caller purpose 和 policy version。
- 成功输出只能是 metadata-only connection readiness / handle reference，不得返回 raw DSN、secret value、database host、database error detail、credential payload 或 driver-specific object 到 response / fixture / audit。
- runtime role 与 migration role 必须分离；runtime role 不能执行 DDL 或 marker write，migration role 不能被 repository adapter 消费。
- connection smoke 必须在显式 test 环境或人工运行记录中执行；fast baseline 不联网、不连接真实数据库、不读取真实 secret、不打开 driver。
- test-only fake resolver 可以作为 future offline smoke 的输入替身，但不能解锁 production connection provider、real resolver、repository mode 或 production API。

## Failure Mapping

entry refresh 继续固定 fail-closed 语义：

| 条件 | failure code |
| --- | --- |
| repository mode reserved / runtime 未打开 | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| secret ref 缺失、resolver disabled、real resolver 不可用或 credential handle 缺失 | `draft_store_unavailable` |
| DB driver / DSN / TLS policy 未配置 | `draft_store_unavailable` |
| database connection unavailable、role 不足或 environment mismatch | `draft_store_unavailable` |
| marker 缺失或 migration 未应用 | `draft_schema_migration_not_applied` |
| marker version mismatch | `draft_store_schema_version_mismatch` |
| marker 不可读、runner 状态不可观测或 lock 状态不可信 | `draft_store_migration_unavailable` |
| stored record contract mismatch | `draft_store_contract_mismatch` |
| repository actor context 不可信 | `draft_auth_context_contract_mismatch` |

失败时不得返回草案主体，不得回退到 `memory_dev`、sample、fixture、dev HTTP route、test auth、fake query executor、test-only fake resolver runtime 或 developer env plaintext。

## No Fallback / No Side Effects

- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 因本 refresh 被启用。
- 不允许 service startup、HTTP route、store selector 或 saved draft request path 自动解析 secret、打开 driver、连接数据库、读取 marker 或写 repository。
- 不允许 test-only fake resolver runtime、fake query executor smoke、static schema artifact、rollback evidence 或 container smoke 替代真实 connection provider implementation prerequisites。
- 本批 checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不读取 env secret、不调用 fake resolver、不调用 production resolver、不调用云 secret 服务、不连接数据库、不导入 driver、不运行 SQL、不读写 schema marker、不创建 runner、不调用 OIDC、不校验 token、不创建 membership adapter、不创建 production API、不执行 workflow。
- side effect counters 必须保持 `secret_resolver_call_count=0`、`fake_resolver_call_count=0`、`production_resolver_call_count=0`、`secret_value_read_count=0`、`credential_handle_created_count=0`、`cloud_secret_call_count=0`、`network_call_count=0`、`database_connection_count=0`、`driver_open_count=0`、`connection_smoke_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`repository_mode_enablement_count=0`、`repository_read_count=0`、`repository_write_count=0`、`oidc_call_count=0`、`token_validation_count=0`、`membership_lookup_count=0`、`production_api_call_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0` 和 `replay_call_count=0`。

## 后续方向

本次 refresh 后，下一步可以在以下方向中选择一个独立推进：

1. `Radish OIDC token / membership upstream evidence refresh`：等待 reviewed issuer、JWKS、client registration 和 membership data source ownership 后再复评。
2. `database driver / DSN / TLS policy readiness` 与 `database role policy readiness` 已分别固定为 `draft_database_driver_dsn_tls_policy_readiness_defined` 和 `draft_database_role_policy_readiness_defined`。
3. `connection smoke strategy`：只定义 explicit test database、offline placeholder handoff、role denial cases、smoke output shape 和 no secret leakage scan，不连接数据库。
4. `production API consumer readiness`：只有 repository mode、auth、database 和 smoke 都可复验后才重新评审。

如果上述依赖仍 blocked，connection provider implementation task card、manual migration runner implementation task card、schema marker contract implementation task card 和 repository mode runtime implementation task card 继续保持不创建。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-database-connection-provider-implementation-v1-plan.md`。
- 不创建 database connection provider、secret resolver、DB driver、DSN parser、connection factory、role policy runtime、connection smoke、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_database_connection_provider_implementation_entry_refresh_defined` 解释为 connection provider ready、database ready、secret resolver ready、repository mode ready、durable persistence ready、production API ready 或 production ready。
