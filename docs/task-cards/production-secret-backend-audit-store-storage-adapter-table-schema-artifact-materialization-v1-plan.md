# Production Secret Backend Audit Store Storage Adapter Table Schema Artifact Materialization v1 计划

更新时间：2026-07-04

## 任务目标

本任务卡固定 `production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1` 的静态实施计划。目标是把后续 metadata-only table schema artifact 的 path、schema version pin、logical field groups、metadata contract compatibility、positive / negative fixtures、schema checker、no secret material scan、no fallback、no side effects 和 artifact guard 固定为可复验要求。

结论状态为 `audit_store_storage_adapter_table_schema_artifact_materialization_task_card_defined`。

本批不创建 `contracts/production-secret-audit-storage-adapter.table-schema.json`，不创建 positive / negative fixture，不创建 SQL、DDL、物理表名、列类型、schema marker runtime、migration runner、storage adapter runtime、DB provider、audit store runtime、repository mode 或 public production API。

## 输入

- [Production Secret Backend Audit Store Storage Adapter Table Schema Artifact Materialization Entry Review v1](../platform/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-entry-review-v1.md)
- [Production Secret Backend Audit Store Storage Adapter Append-only Table Schema Boundary Readiness v1](../platform/production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1.md)
- [Production Secret Backend Audit Store Storage Adapter Metadata Contract Artifact Materialization v1](../platform/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.md)
- [Production Secret Backend Audit Store Storage Adapter Backend Product Selection Review v1](../platform/production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.md)
- [Production Secret Backend Audit Store Runtime Blocker Matrix v1](../platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md)
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-entry-review-v1.json`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1.json`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.json`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.json`
- `scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`

## 交付物

1. 新增平台专题：
   - `docs/platform/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.md`
2. 新增 task card：
   - `docs/task-cards/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1-plan.md`
3. 新增 fixture：
   - `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.json`
4. 新增 checker：
   - `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.py`
5. 更新聚合与入口：
   - `docs/radishmind-current-focus.md`
   - `docs/features/README.md`
   - `docs/features/workflow/README.md`
   - `docs/features/workflow/saved-workflow-draft-v1.md`
   - `docs/platform/README.md`
   - `docs/platform/production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md`
   - `docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md`
   - `docs/radishmind-architecture.md`
   - `docs/radishmind-integration-contracts.md`
   - `docs/radishmind-product-scope.md`
   - `docs/task-cards/README.md`
   - `docs/task-cards/production-secret-backend-audit-store-runtime-blocker-matrix-v1-plan.md`
   - `docs/task-cards/production-secret-backend-implementation-v1-plan.md`
   - `scripts/README.md`
   - `scripts/check-repo.py`
6. 更新关联静态证据：
   - runtime blocker matrix fixture 与 checker
   - implementation readiness fixture 与 checker
   - 本周周志

## Future Artifact Requirements

| gate | 后续 artifact 实现要求 |
| --- | --- |
| artifact path | `contracts/production-secret-audit-storage-adapter.table-schema.json` |
| schema version pin | `audit-storage-adapter-table-schema-v1` |
| logical field groups | 必须来自 append-only table schema boundary readiness |
| metadata contract compatibility | 必须对齐 storage adapter metadata contract envelope |
| logical field keys | 只允许 metadata-only logical key，不允许物理 column detail |
| positive fixture | 必须覆盖 append-only audit record schema shape |
| negative fixtures | 必须覆盖 physical table detail、raw storage payload、DSN / endpoint、SQL / DDL、payload hash、secret-derived hash 和 additionalProperties |
| schema checker | 必须验证 artifact、正例、负例、contract compatibility、no secret material scan 和 check-repo 注册顺序 |
| writer compatibility | 只做静态 metadata compatibility，不创建 writer runtime |

## 停止线

- 不创建 table schema artifact、positive fixture、negative fixture、schema checker 或 no secret material scan output。
- 不选择具体 database product、vendor service、database driver 或 provider resource。
- 不创建 SQL、DDL、物理表名、列名、列类型、索引、约束、partition、sequence、trigger、function、schema version table 或 migration runner。
- 不创建 storage adapter runtime task card、storage adapter runtime、storage adapter client、durable audit backend runtime、database connection provider 或 SQL migration。
- 不创建 audit store runtime task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、repository mode runtime 或 public production API。
- 不读取真实 secret、developer env secret、credential payload、DSN、provider raw URL、raw request / response / audit / writer / storage payload、payload hash、secret-derived hash、schema marker output 或 migration output。
- 不把本 task card 写成 table schema artifact materialized、SQL ready、database ready、storage adapter ready、audit store ready、production resolver ready 或 production ready。

## 验证

本批至少执行：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
