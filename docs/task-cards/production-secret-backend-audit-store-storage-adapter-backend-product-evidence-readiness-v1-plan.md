# `Production Secret Backend Audit Store Storage Adapter Backend Product Evidence Readiness` v1 计划

更新时间：2026-06-30

## 任务目标

本任务卡固定 audit store storage adapter runtime task card 前的 backend product evidence readiness。当前只定义 evidence envelope、candidate class、选择前证据、失败语义、脱敏诊断、artifact guard 和验证链路；不选择具体 backend product，不创建 storage adapter runtime、DB provider、SQL、audit store runtime、repository mode 或 public production API。

对应平台专题：[Production Secret Backend Audit Store Storage Adapter Backend Product Evidence Readiness v1](../platform/production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.md)。

## 输入事实

- `audit_store_storage_adapter_runtime_implementation_entry_review_defined`
- `audit_store_concrete_durable_backend_selection_review_defined`
- `audit_store_runtime_blocker_matrix_defined`
- `audit_store_runtime_event_schema_artifact_implemented`
- `implementation_readiness_defined`

## 输出范围

1. 新增 backend product evidence readiness 平台文档。
2. 新增 fixture 固定 readiness decision、evidence envelope、候选类别、failure mapping、diagnostics allowlist、no fallback、no side effects 和 artifact guard。
3. 新增 checker 校验 fixture、上游依赖、blocker matrix、implementation readiness、文档引用和 `check-repo.py` 注册顺序。
4. 同步 blocker matrix，把 durable backend blocker 推进为 `backend_product_evidence_readiness_defined_task_card_blocked`。
5. 同步 implementation readiness，把 backend product evidence 从 `not_selected` 推进为 `readiness_defined_without_product_selection`，同时保持 backend product selection 为 `not_selected`。

## 停止线

- 不选择具体 DB、object store、queue、log sink、vendor service 或 managed service。
- 不创建 storage adapter runtime implementation task card。
- 不创建 storage adapter runtime、client、driver、writer、connection factory、DSN parser、SQL、migration、schema marker 或 DB provider。
- 不创建 audit store runtime task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、repository mode runtime 或 public production API。
- 不写 secret、DSN、provider endpoint、hostname、bucket name、queue name、topic name、region resource id、token、API key、authorization header、cookie 或 raw provider dump。

## 验收

- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.json` 存在并固定 `audit_store_storage_adapter_backend_product_evidence_readiness_defined`。
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.py` 校验通过。
- `Production Secret Backend Audit Store Runtime Blocker Matrix v1`、implementation readiness、current focus、features、Saved Workflow Draft 专题、platform README、task card 索引、scripts README 和周志同步。
- `git diff --check` 通过。
- `./scripts/check-repo.sh --fast` 通过。

## 后续

后续若继续 durable backend 上游，应推进 `storage_adapter_metadata_contract_artifact_readiness`，固定 metadata-only adapter contract artifact；仍不创建 runtime、DB provider、repository mode 或 production API。
