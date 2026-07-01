# `Production Secret Backend Audit Store Storage Adapter Negative Leakage Scan Evidence Readiness` v1 计划

更新时间：2026-07-01

## 任务目标

在 `production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1` 之后，固定 storage adapter runtime task card 前必须具备的 negative leakage scan evidence。

本任务卡状态固定为 `audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined`，readiness decision 为 `negative_leakage_scan_evidence_defined_without_runtime`。

## 范围

本批只定义：

- metadata-only negative leakage scan manifest reference（`metadata_only_negative_leakage_scan_manifest_reference_defined`）
- metadata-only scan target reference
- raw payload / secret / credential / provider / backend forbidden material coverage
- diagnostic allowlist
- failure mapping
- sanitized diagnostics
- no fallback / no side effects
- artifact guard

## 不做

- 不创建 negative leakage scanner、CLI、runtime smoke 或 committed scan output
- 不创建 storage adapter runtime task card 或 storage adapter runtime
- 不创建 offline validation runner、retention executor、redaction executor、DB provider、DB driver、DSN parser 或 SQL
- 不创建 audit store runtime、audit writer、delivery runtime、idempotency runtime 或 duplicate detector
- 不创建 production resolver runtime、repository mode runtime 或 public production API
- 不读取真实 secret、credential payload、raw event payload、raw writer payload、raw storage payload、payload hash、provider detail 或 backend detail
- 不选择具体数据库、对象存储、队列、日志 sink、vendor service 或云 secret backend

## 交付物

- 平台专题：`docs/platform/production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.md`
- Fixture：`scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.json`
- Checker：`scripts/check-production-ops-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.py`
- 聚合入口：`scripts/check-repo.py`
- 同步文档：runtime blocker matrix、implementation readiness、current focus、features/workflow 入口、platform README、scripts README 和周志

## 验收

1. `audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined` 只表示 negative leakage scan evidence 已定义，不代表 scanner 或 runtime ready。
2. scan manifest、scan target、forbidden material matrix 和 diagnostic allowlist 都是 metadata-only reference。
3. forbidden material coverage 必须覆盖 raw payload、secret material、credential material、provider detail、backend detail、payload hash 和 committed scan output。
4. durable backend blocker 只能推进到 `negative_leakage_scan_evidence_readiness_defined_task_card_blocked`。
5. 下一项依赖固定为 `storage_adapter_rollback_recovery_evidence_readiness`。
6. checker、runtime blocker matrix、implementation readiness 和 `check-repo.py` 注册顺序保持一致。

## 停止线

本批完成后仍不得创建 scanner、scan output、storage adapter runtime、offline validation runner、DB provider、SQL、audit store runtime、repository mode、production resolver runtime 或 public production API。
