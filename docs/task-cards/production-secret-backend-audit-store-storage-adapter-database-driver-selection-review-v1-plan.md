# Production Secret Backend Audit Store Storage Adapter Database Driver Selection Review v1 Plan

更新时间：2026-07-05

## 任务边界

本任务推进 `production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1`，状态为 `audit_store_storage_adapter_database_driver_selection_review_defined`。

目标是消费 `audit_store_storage_adapter_database_driver_selection_readiness_defined`、`audit_store_storage_adapter_database_provider_selection_review_defined`、`audit_store_storage_adapter_concrete_database_selection_review_defined` 和前序 storage adapter 静态证据，固定 future Go PostgreSQL driver candidate 的选择评审结论。

本批是 selection review：只选择 reference-only driver candidate `github.com/jackc/pgx/v5`，不新增 Go import、不固定 dependency version、不改 `go.mod`、不创建 DSN parser、connection provider、provider account resource、endpoint、SQL、DDL、物理表名、列名、列类型、schema marker runtime、migration runner、storage adapter runtime、audit store runtime、repository mode 或 public production API。

## 输入

- `docs/platform/production-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1.json`
- `docs/platform/production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1.json`
- `docs/platform/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1.json`
- `docs/platform/production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.md`
- `contracts/production-secret-audit-storage-adapter.metadata-contract.json`
- `contracts/production-secret-audit-storage-adapter.table-schema.json`
- `docs/platform/production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.md`
- `docs/platform/production-secret-backend-audit-store-storage-adapter-negative-leakage-runtime-scan-boundary-readiness-v1.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- 外部只读来源：`https://github.com/jackc/pgx`、`https://pkg.go.dev/github.com/jackc/pgx/v5`

## 输出

- `docs/platform/production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.py`
- `check-repo.py` 快速验证链路注册
- runtime blocker matrix、implementation readiness、evidence rollup、current focus、features README、platform README、task card README、scripts README 与本周 devlog 同步

## Selection Decision

| 项 | 结论 |
| --- | --- |
| review status | `audit_store_storage_adapter_database_driver_selection_review_defined` |
| selection decision | `database_driver_candidate_selected_pgx_v5_runtime_blocked` |
| selected driver candidate | `github.com/jackc/pgx/v5` |
| selected database engine | `postgresql_compatible_append_only_relational_database` |
| selected provider candidate class | `managed_postgresql_compatible_service` |
| driver selection status | `selected_driver_candidate_without_runtime_import` |
| driver package status | `selected_candidate_reference_only` |
| driver import status | `not_created` |
| runtime task card decision | `storage_adapter_runtime_task_card_still_blocked_after_database_driver_selection_review` |
| next dependency | `storage_adapter_database_connection_lifecycle_readiness` |

## 停止线

- 不新增 Go import，不改 `go.mod` / `go.sum`，不下载依赖，不固定 dependency version。
- 不选择云厂商、managed database product、hosted product、endpoint、region detail、resource id 或 account scoped resource。
- 不创建 DSN parser、connection provider、connection lifecycle runtime、connection factory、database connection 或 credential material。
- 不创建 SQL、DDL、物理表名、列名、列类型、schema marker runtime 或 migration runner。
- 不创建 storage adapter runtime task card、storage adapter runtime、audit store runtime task card、audit store runtime 或 production resolver runtime task card。
- 不启用 repository mode，不创建 public production API。
- 不把 driver candidate 选择写成 import ready、DSN ready、connection ready、runtime ready、repository ready、production API ready 或 production ready。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
