# Production Secret Backend Audit Store Storage Adapter Metadata Contract Artifact Materialization v1 任务卡

更新时间：2026-07-02

## 任务目标

创建 `storage_adapter_metadata_contract_artifact_materialization_task_card` 对应实现批次，把前序 entry review 允许的 metadata-only contract artifact 物化为可复验文件。

状态固定为 `audit_store_storage_adapter_metadata_contract_artifact_materialized`。

## 输入证据

- `audit_store_storage_adapter_metadata_contract_artifact_materialization_entry_review_defined`
- `audit_store_storage_adapter_runtime_implementation_entry_refresh_defined`
- `audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined`
- `audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined`
- `audit_store_storage_adapter_offline_validation_evidence_readiness_defined`
- `audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined`
- `audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined`
- `audit_store_storage_adapter_metadata_contract_artifact_readiness_defined`
- `audit_store_runtime_blocker_matrix_defined`
- `implementation_readiness_defined`

## 交付内容

| 交付 | 路径 |
| --- | --- |
| contract artifact | `contracts/production-secret-audit-storage-adapter.metadata-contract.json` |
| contract 专题 | `docs/contracts/production-secret-audit-storage-adapter-metadata-contract.md` |
| platform 专题 | `docs/platform/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.md` |
| materialization fixture | `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.json` |
| positive fixture | `scripts/checks/fixtures/production-secret-audit-storage-adapter-metadata-contract-positive-v1.json` |
| missing required negative fixture | `scripts/checks/fixtures/production-secret-audit-storage-adapter-metadata-contract-missing-required-negative-v1.json` |
| forbidden field negative fixture | `scripts/checks/fixtures/production-secret-audit-storage-adapter-metadata-contract-forbidden-field-negative-v1.json` |
| additional properties negative fixture | `scripts/checks/fixtures/production-secret-audit-storage-adapter-metadata-contract-additional-properties-negative-v1.json` |
| writer compatibility fixture | `scripts/checks/fixtures/production-secret-audit-storage-adapter-metadata-contract-writer-compatibility-v1.json` |
| checker | `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.py` |

## 验收条件

- contract artifact 固定 `audit-storage-adapter-metadata-contract-v1`。
- input envelope、result envelope、record identity、failure taxonomy 与前置 readiness 一致。
- positive fixture 通过 checker。
- missing required、forbidden field 和 additionalProperties 负例被拒绝。
- writer compatibility smoke 能证明 writer output projection 可被 storage adapter input envelope 消费。
- no secret material scan 覆盖 contract artifact 与全部新增 contract fixtures。
- runtime blocker matrix 和 implementation readiness 已更新到 `storage_adapter_backend_product_selection_review`。

## 停止线

- 不选择具体 backend product。
- 不创建 storage adapter runtime task card、storage adapter runtime、storage adapter client、DB provider、database connection、SQL migration、schema marker、audit store runtime task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、repository mode 或 public production API。
- 不保存、读取、派生或扫描真实 secret、credential payload、provider raw URL、DSN、数据库细节、raw payload、payload hash、scanner output 或 recovery output。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
