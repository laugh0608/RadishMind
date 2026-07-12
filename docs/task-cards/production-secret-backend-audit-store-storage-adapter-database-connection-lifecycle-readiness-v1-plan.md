# Production Secret Backend Audit Store Storage Adapter Database Connection Lifecycle Readiness v1 Plan

更新时间：2026-07-05

## 任务边界

本任务推进 `production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1`，状态为 `audit_store_storage_adapter_database_connection_lifecycle_readiness_defined`。

目标是消费 database driver selection review、database provider / driver / DSN / TLS / role policy readiness、metadata-only table schema artifact、metadata contract artifact、offline adapter smoke strategy、negative leakage runtime scan boundary、rollback / recovery evidence、runtime blocker matrix 和 implementation readiness，固定 future database connection lifecycle 的静态准入。

本批只定义 secret-ref-only DSN handoff、TLS / role / environment binding、pool policy、timeout budget、retry / transaction / partial write recovery、duplicate / replay fail-closed、sanitized diagnostics、schema marker / migration handoff、offline verification、negative leakage scan、rollback / rollout 边界；不新增 Go import、不固定 dependency version、不改 `go.mod` / `go.sum`，不创建 DSN parser、connection provider、DB provider、connection factory、pool runtime、health check runtime、SQL、DDL、物理表名、列名、列类型、schema marker runtime、migration runner、storage adapter runtime、audit store runtime、repository mode 或 public production API。

## 输入

- `docs/platform/production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.json`
- `docs/platform/production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.json`
- `contracts/production-secret-audit-storage-adapter.metadata-contract.json`
- `contracts/production-secret-audit-storage-adapter.table-schema.json`
- `docs/platform/production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.md`
- `docs/platform/production-secret-backend-audit-store-storage-adapter-negative-leakage-runtime-scan-boundary-readiness-v1.md`
- `docs/platform/production-secret-backend-audit-store-storage-adapter-rollback-recovery-evidence-readiness-v1.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`

## 输出

- `docs/platform/production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.py`
- `check-repo.py` 快速验证链路注册
- runtime blocker matrix、implementation readiness、evidence rollup、current focus、features README、workflow saved draft 文档、platform README、task card README、scripts README 与本周 devlog 同步

## Readiness Decision

| 项 | 结论 |
| --- | --- |
| readiness status | `audit_store_storage_adapter_database_connection_lifecycle_readiness_defined` |
| readiness decision | `database_connection_lifecycle_readiness_defined_without_connection_runtime` |
| selected driver candidate | `github.com/jackc/pgx/v5` |
| selected database engine | `postgresql_compatible_append_only_relational_database` |
| selected provider candidate class | `managed_postgresql_compatible_service` |
| DSN handoff status | `secret_ref_only_dsn_handoff_defined` |
| TLS / role / environment binding status | `static_tls_role_environment_binding_defined` |
| pool policy status | `static_pool_policy_defined_without_pool_runtime` |
| timeout budget status | `static_timeout_budget_defined_without_runtime_timer` |
| retry / transaction / recovery status | `static_recovery_policy_defined_without_runtime` |
| duplicate / replay status | `duplicate_replay_fail_closed_policy_defined` |
| runtime task card decision | `storage_adapter_runtime_task_card_still_blocked_after_database_connection_lifecycle_readiness` |

## 停止线

- 不新增 Go import，不改 `go.mod` / `go.sum`，不下载依赖，不固定 dependency version。
- 不输出 raw secret、raw DSN、endpoint、host、database name、credential payload、full secret ref、full credential handle、provider raw URL 或 provider detail。
- 不选择云厂商、managed database product、hosted product、endpoint、region detail、resource id 或 account scoped resource。
- 不创建 DSN parser、connection provider、DB provider、connection lifecycle runtime、connection factory、pool runtime、health check runtime、database connection 或 credential material。
- 不创建 SQL、DDL、物理表名、列名、列类型、schema marker runtime 或 migration runner。
- 不创建 storage adapter runtime task card、storage adapter runtime、audit store runtime task card、audit store runtime 或 production resolver runtime task card。
- 不启用 repository mode，不创建 public production API。
- 不把 connection lifecycle readiness 写成 connection ready、pool ready、health check ready、DB provider ready、SQL ready、runtime ready、repository ready、production API ready 或 production ready。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
