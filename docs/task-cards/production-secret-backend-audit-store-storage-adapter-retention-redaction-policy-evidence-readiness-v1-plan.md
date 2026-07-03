# Production Secret Backend Audit Store Storage Adapter Retention / Redaction Policy Evidence Readiness v1 计划

更新时间：2026-06-30

## 目标

在 `production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1` 之后，固定 storage adapter runtime task card 前必须具备的 retention / redaction policy evidence。

本任务卡状态固定为 `audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined`，readiness decision 为 `retention_redaction_policy_evidence_defined_without_runtime`。

## 输入

- `audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined`
- `audit_store_storage_adapter_metadata_contract_artifact_readiness_defined`
- `audit_store_storage_adapter_backend_product_evidence_readiness_defined`
- `audit_store_storage_adapter_runtime_implementation_entry_review_defined`
- `audit_store_runtime_blocker_matrix_defined`
- `implementation_readiness_defined`

## 交付物

- 平台专题：`docs/platform/production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.md`
- Fixture：`scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.json`
- Checker：`scripts/check-production-ops-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.py`
- 聚合对齐：`check-repo.py`、runtime blocker matrix、implementation readiness、相关入口文档与周志

## 验收点

- retention window 只能以 `metadata_only_retention_window_reference_defined` 形式存在，不定义物理 TTL、cleanup job、partition purge、bucket lifecycle 或 provider rule。
- redaction 只能以 `metadata_only_redaction_policy_reference_defined` 形式存在，不读取、输出、写入或派生 raw payload、payload hash、credential payload 或 secret material。
- retention / redaction 必须与 `append_only_insert_only` 语义兼容，不允许 delete / overwrite / truncate / compact / inline redact / mutate identity 成为 adapter 成功路径。
- failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard 必须可由专项 checker 复验。
- checker 必须确认本批不创建 storage adapter runtime task card、storage adapter runtime、retention executor、redaction executor、DB provider、SQL、audit store runtime、writer / delivery / idempotency runtime、production resolver runtime、repository mode 或 public production API。

## 停止线

- 不创建 `contracts/production-secret-audit-storage-adapter.metadata-contract.json`。
- 不选择具体 backend product、database、object store、queue、log sink 或 vendor service。
- 不创建 storage adapter runtime implementation task card、storage adapter runtime、retention executor、redaction executor、client、driver、DB provider、DSN parser、SQL migration、schema marker reader / writer。
- 不创建 audit store runtime implementation task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、duplicate detector、retry executor 或 replay executor。
- 不读取 secret，不调用云 secret 服务，不连接数据库，不执行 SQL，不启用 repository mode，不调用 public production API。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
