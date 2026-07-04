# Production Secret Backend Audit Store Storage Adapter Offline Adapter Smoke Strategy Readiness v1 Plan

状态：`audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined`  
readiness decision：`offline_adapter_smoke_strategy_defined_without_runtime`

## 背景

当前 durable store 上游已完成 metadata-only table schema artifact materialization：`audit_store_storage_adapter_table_schema_artifact_materialized`。该锚点只把 `contracts/production-secret-audit-storage-adapter.table-schema.json` 固定为 logical table schema artifact，并不意味着 storage adapter runtime task card 可以创建。

本批消费 table schema artifact、metadata contract artifact、static product class selection、database provider / driver / DSN / TLS / role policy readiness、runtime blocker matrix 和 implementation readiness，补齐 offline adapter smoke strategy readiness。

## 交付

- 新增 `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.json`。
- 新增 positive fixture：`scripts/checks/fixtures/production-secret-audit-storage-adapter-offline-smoke-positive-v1.json`。
- 新增 negative fixtures：
  - `scripts/checks/fixtures/production-secret-audit-storage-adapter-offline-smoke-missing-manifest-negative-v1.json`
  - `scripts/checks/fixtures/production-secret-audit-storage-adapter-offline-smoke-runtime-touch-negative-v1.json`
  - `scripts/checks/fixtures/production-secret-audit-storage-adapter-offline-smoke-secret-material-negative-v1.json`
- 新增 checker：`scripts/check-production-ops-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.py`。
- 同步 runtime blocker matrix、implementation readiness、evidence rollup、current focus、features README、platform 文档和周志。

## 验收

- strategy fixture 固定 `offline_adapter_smoke_strategy_defined_without_runtime`。
- positive / negative fixture 均为 metadata-only static case，不执行真实 backend。
- checker 覆盖 no secret material scan、runtime touch guard、sanitized diagnostic allowlist、artifact guard、matrix alignment 和 implementation readiness alignment。
- runtime blocker matrix 推进为 `storage_adapter_offline_adapter_smoke_strategy_readiness_defined_runtime_blocked`。
- 当前下一依赖推进为 `storage_adapter_negative_leakage_runtime_scan_boundary_readiness`。
- `audit_storage_adapter_runtime_task_card_status`、`audit_storage_adapter_runtime_status`、`audit_store_runtime_status` 均保持 `not_created`。

## 停止线

本批不创建 offline adapter smoke runner、committed smoke output、negative leakage runtime scan、scan output、storage adapter runtime task card、storage adapter runtime、audit store runtime task card、audit store runtime、database provider、database connection、driver、DSN parser、SQL、DDL、physical table schema、schema marker runtime、migration runner、repository mode 或 public production API。

## 验证

```bash
python3 scripts/check-production-ops-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.py
python3 scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
python3 scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/check-repo.sh --fast
```
