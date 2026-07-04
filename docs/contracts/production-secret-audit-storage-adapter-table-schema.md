# Production Secret Audit Storage Adapter Table Schema 契约

更新时间：2026-07-04

## 定位

`contracts/production-secret-audit-storage-adapter.table-schema.json` 是 future production secret backend audit store storage adapter 的 metadata-only logical table schema artifact。它固定 schema version、logical field groups、metadata contract compatibility、positive / negative fixtures 和 no secret material scan；它不是 SQL、DDL、物理表、物理列或数据库迁移设计。

对应实现切片为 `production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1`，状态为 `audit_store_storage_adapter_table_schema_artifact_materialized`。

## 字段边界

- `table_schema_version` 固定为 `audit-storage-adapter-table-schema-v1`。
- `storage_adapter_contract_version` 固定为 `audit-storage-adapter-metadata-contract-v1`。
- logical field groups 来自 append-only table schema boundary readiness：`identity`、`ordering`、`payload_reference`、`retention_redaction`、`delivery_recovery` 与 `diagnostics`。
- schema properties 只覆盖 metadata contract artifact 中的 input envelope、result envelope 与 record identity 字段，再额外固定 `table_schema_version`。
- `backend_product_class` 只允许静态 product class `managed_database_append_only_table`；它不选择具体数据库、vendor、driver、endpoint、DSN、table name 或 column type。

## 正负例

| 目标 | fixture | 预期 |
| --- | --- | --- |
| positive metadata-only table schema | `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-positive-v1.json` | 通过 schema checker |
| missing required field | `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-missing-required-negative-v1.json` | 被拒绝 |
| physical detail | `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-physical-detail-negative-v1.json` | 被拒绝 |
| secret material | `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-secret-material-negative-v1.json` | 被拒绝 |
| additional properties | `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-additional-properties-negative-v1.json` | 被拒绝 |

## 禁止字段

schema artifact 和 checker 显式拒绝 secret value、raw secret、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、database hostname、database name、table name、column name、column type、index name、constraint name、partition policy、DDL、SQL、SQL migration、database sequence、trigger、database function、schema version table、migration command、cloud credential、credential payload、raw request / response / audit / event / writer / storage payload、payload hash、secret-derived hash、provider error detail、database error detail、scanner raw finding、scan output、schema marker raw output 和 migration output。

## 消费规则

- future storage adapter runtime 只能把该 artifact 当作 metadata-only logical schema input，不能把 fixture 当作真实 audit record、storage record 或 runtime output。
- `storage_record_ref`、`storage_record_identity_ref`、`audit_event_ref`、`append_only_sequence_ref`、`idempotency_key_ref`、`writer_result_ref` 和 `delivery_commit_ref` 都是 opaque reference，不是数据库主键、物理列、provider resource id 或 secret。
- 后续如果扩字段或 failure code，必须同步更新 table schema artifact、positive / negative fixtures、checker、platform 专题、runtime blocker matrix 和 implementation readiness。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.py
./scripts/check-repo.sh --fast
```

## 停止线

本契约 artifact 不创建 SQL、DDL、物理表名、列名、列类型、索引、约束、schema marker runtime、migration runner、storage adapter runtime、DB provider、audit store runtime、repository mode 或 public production API。
