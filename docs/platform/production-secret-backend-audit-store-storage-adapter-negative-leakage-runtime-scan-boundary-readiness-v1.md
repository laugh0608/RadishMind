# Production Secret Backend Audit Store Storage Adapter Negative Leakage Runtime Scan Boundary Readiness v1

状态：`audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined`  
readiness decision：`negative_leakage_runtime_scan_boundary_defined_without_runtime`

## 目标

本批在 `audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined` 之后，定义 future negative leakage runtime scan 的 metadata-only 边界。它只固定后续 scan runner 可以消费哪些静态引用、必须覆盖哪些 forbidden material classes、如何输出 sanitized diagnostics、以及哪些 runtime touch 和 artifact 必须继续拒绝。

本批不创建 negative leakage scanner、scan runner、scan output、storage adapter runtime task card、storage adapter runtime、audit store runtime task card、audit store runtime、database provider、database connection、driver、DSN parser、SQL、DDL、schema marker runtime、migration runner、repository mode 或 public production API。

## 输入

- `production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1`
- `production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1`
- `production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1`
- `production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1`
- `production-secret-backend-audit-store-runtime-blocker-matrix-v1`
- `production-secret-backend-implementation-readiness`

## Boundary

本批新增 boundary fixture：`scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-negative-leakage-runtime-scan-boundary-readiness-v1.json`。

future scan boundary 只允许 metadata-only reference：

- `runtime_scan_manifest_ref`
- `offline_smoke_strategy_ref`
- `negative_leakage_evidence_ref`
- `metadata_contract_ref`
- `table_schema_artifact_ref`
- `scan_target_allowlist_ref`
- `forbidden_material_matrix_ref`
- `diagnostic_allowlist_ref`
- `failure_taxonomy_ref`
- `policy_version`
- `audit_ref`

正向 case 固定为 `scripts/checks/fixtures/production-secret-audit-storage-adapter-negative-leakage-runtime-scan-boundary-positive-v1.json`，只表达 future scan manifest reference、target allowlist reference、forbidden material matrix reference 和 sanitized diagnostic envelope reference。

负向 case 固定为：

- `scripts/checks/fixtures/production-secret-audit-storage-adapter-negative-leakage-runtime-scan-boundary-runtime-touch-negative-v1.json`
- `scripts/checks/fixtures/production-secret-audit-storage-adapter-negative-leakage-runtime-scan-boundary-secret-material-negative-v1.json`
- `scripts/checks/fixtures/production-secret-audit-storage-adapter-negative-leakage-runtime-scan-boundary-scan-output-negative-v1.json`

这些 fixture 只用于静态检查。负向 case 说明 runtime / backend touch、secret material、raw payload、provider / backend detail 或提前 committed scan output 都必须失败关闭，不能产生真实 scan、真实 backend 调用或 scan output。

## Guard

checker：`scripts/check-production-ops-secret-backend-audit-store-storage-adapter-negative-leakage-runtime-scan-boundary-readiness-v1.py`

checker 覆盖：

- boundary fixture、positive fixture、negative fixture 的静态一致性；
- no secret material scan；
- runtime touch、real backend touch、scan runner / output 提前创建的失败关闭；
- target allowlist、forbidden material coverage、diagnostic allowlist 和 failure taxonomy；
- artifact guard；
- runtime blocker matrix 与 implementation readiness 对齐。

新增 current blocker：

- `storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined_runtime_blocked`

当前下一依赖：

- `storage_adapter_runtime_implementation_entry_refresh_after_negative_leakage_runtime_scan_boundary`

runtime task card 结论：

- `storage_adapter_runtime_task_card_still_blocked_after_negative_leakage_runtime_scan_boundary`

## 停止线

本批不得创建或承诺：

- negative leakage scanner、scan runner、scan CLI 或 committed scan output；
- storage adapter runtime task card 或 storage adapter runtime；
- audit store runtime task card、audit store runtime、audit writer runtime、delivery runtime 或 idempotency runtime；
- concrete database vendor、driver、DSN parser、connection provider 或 connection；
- SQL、DDL、physical table name、physical column name、physical column type、index、constraint；
- schema marker runtime、migration runner、repository mode、public production API；
- production resolver runtime 或 resolver backend call。

## 验证

定向验证入口：

```bash
python3 scripts/check-production-ops-secret-backend-audit-store-storage-adapter-negative-leakage-runtime-scan-boundary-readiness-v1.py
python3 scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
python3 scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/check-repo.sh --fast
```

如果继续 durable store 上游，下一批应独立推进 `storage_adapter_runtime_implementation_entry_refresh_after_negative_leakage_runtime_scan_boundary`，重新评审 storage adapter runtime task card 是否仍 blocked；不得直接创建 storage adapter runtime、DB provider、audit store runtime、repository mode 或 production API。
