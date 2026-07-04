# Production Secret Backend Audit Store Storage Adapter Offline Adapter Smoke Strategy Readiness v1

状态：`audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined`  
readiness decision：`offline_adapter_smoke_strategy_defined_without_runtime`

## 目标

本批在 `audit_store_storage_adapter_table_schema_artifact_materialized` 之后，定义 future storage adapter offline smoke 的 metadata-only 策略边界。它只说明后续 smoke runner 应读取哪些静态引用、覆盖哪些正负路径、如何失败关闭、以及哪些字段和 runtime touch 必须拒绝。

本批不创建 offline adapter smoke runner、committed smoke output、storage adapter runtime task card、storage adapter runtime、audit store runtime task card、audit store runtime、database provider、database connection、driver、DSN parser、SQL、DDL、physical table schema、schema marker runtime、migration runner、repository mode 或 public production API。

## 输入

- `contracts/production-secret-audit-storage-adapter.metadata-contract.json`
- `contracts/production-secret-audit-storage-adapter.table-schema.json`
- `production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1`
- `production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1`
- `production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1`
- `production-secret-backend-audit-store-runtime-blocker-matrix-v1`
- `production-secret-backend-implementation-readiness`

## Strategy Boundary

本批新增 strategy fixture：`scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.json`。

strategy 只允许 metadata-only reference：

- `smoke_manifest_ref`
- `metadata_contract_ref`
- `table_schema_artifact_ref`
- `backend_product_evidence_ref`
- `positive_case_ref`
- `negative_case_ref`
- `failure_taxonomy_ref`
- `policy_version`
- `audit_ref`

正向 case 固定为 `scripts/checks/fixtures/production-secret-audit-storage-adapter-offline-smoke-positive-v1.json`，只表达 logical write candidate、logical storage record reference、metadata contract reference 和 table schema artifact reference。

负向 case 固定为：

- `scripts/checks/fixtures/production-secret-audit-storage-adapter-offline-smoke-missing-manifest-negative-v1.json`
- `scripts/checks/fixtures/production-secret-audit-storage-adapter-offline-smoke-runtime-touch-negative-v1.json`
- `scripts/checks/fixtures/production-secret-audit-storage-adapter-offline-smoke-secret-material-negative-v1.json`

这些 fixture 只用于静态检查。负向 case 说明缺少 manifest、触碰 runtime / real backend、携带 forbidden material 时必须失败关闭，不产生真实 backend 调用或输出。

## Guard

checker：`scripts/check-production-ops-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.py`

checker 覆盖：

- strategy fixture、positive fixture、negative fixture 的静态一致性；
- no secret material scan；
- `runtime_touch`、`real_backend_touch`、forbidden runtime mechanism 的失败关闭；
- sanitized diagnostic allowlist；
- artifact guard；
- runtime blocker matrix 与 implementation readiness 对齐。

新增 current blocker：

- `storage_adapter_offline_adapter_smoke_strategy_readiness_defined_runtime_blocked`

当前下一依赖：

- `storage_adapter_negative_leakage_runtime_scan_boundary_readiness`

runtime task card 结论：

- `storage_adapter_runtime_task_card_still_blocked_after_offline_adapter_smoke_strategy_readiness`

## 停止线

本批不得创建或承诺：

- offline adapter smoke runner 或 committed smoke output；
- negative leakage runtime scan、scan runner 或 scan output；
- storage adapter runtime task card 或 storage adapter runtime；
- audit store runtime task card、audit store runtime、audit writer runtime、delivery runtime 或 idempotency runtime；
- concrete database vendor、driver、DSN parser、connection provider 或 connection；
- SQL、DDL、physical table name、physical column name、physical column type、index、constraint；
- schema marker runtime、migration runner、repository mode、public production API；
- production resolver runtime 或 resolver backend call。

## 验证

定向验证入口：

```bash
python3 scripts/check-production-ops-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.py
python3 scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
python3 scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/check-repo.sh --fast
```

如果继续 durable store 上游，下一批应独立推进 `storage_adapter_negative_leakage_runtime_scan_boundary_readiness`，先定义 runtime scan boundary、target allowlist、forbidden material coverage、diagnostic policy 和 artifact guard，再考虑 storage adapter runtime task card。

## 后续推进状态

`storage_adapter_negative_leakage_runtime_scan_boundary_readiness` 已由后续批次消费，当前 runtime task card 结论已推进为 `storage_adapter_runtime_task_card_still_blocked_after_negative_leakage_runtime_scan_boundary`，当前下一依赖为 `storage_adapter_runtime_implementation_entry_refresh_after_negative_leakage_runtime_scan_boundary`。
