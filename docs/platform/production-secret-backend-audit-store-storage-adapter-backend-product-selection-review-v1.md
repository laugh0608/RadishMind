# Production Secret Backend Audit Store Storage Adapter Backend Product Selection Review v1

更新时间：2026-07-02

## 文档目的

本文档承接 `Production Secret Backend Audit Store Storage Adapter Metadata Contract Artifact Materialization v1`，固定 storage adapter backend product selection 的静态评审结论。

对应切片：`production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1`。

结论：状态为 `audit_store_storage_adapter_backend_product_selection_review_defined`。本批只把 storage adapter 的 backend product class 静态选择为 `managed_database_append_only_table`，对应保留 profile 为 `reserved_managed_database_append_only_table_profile`；该选择只代表后续 runtime task card 的目标产品类别，不选择 PostgreSQL / MySQL / SQLite、vendor service、driver、DSN、SQL schema、连接参数或真实数据库，不创建 storage adapter runtime、DB provider、audit store runtime、repository mode 或 public production API。

selection decision：`storage_adapter_product_class_selected_managed_database_append_only_table_runtime_blocked`。

## 输入证据

- `audit_store_storage_adapter_backend_product_evidence_readiness_defined` 已固定 backend product evidence envelope、product class allowlist 与选择前证据要求。
- `audit_store_storage_adapter_metadata_contract_artifact_materialized` 已物化 metadata-only contract artifact、positive / negative fixtures、writer compatibility smoke 与 no secret material scan。
- `audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined`、`audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined`、`audit_store_storage_adapter_offline_validation_evidence_readiness_defined`、`audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined` 与 `audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined` 已收束 storage adapter runtime 之前的主要证据。
- `audit_store_concrete_durable_backend_selection_review_defined` 已把 durable backend family 固定为 `append_only_metadata_audit_log`，但不等于 storage adapter backend product selected。
- `implementation_readiness_defined` 和 `audit_store_runtime_blocker_matrix_defined` 仍保持 runtime、DB provider、audit store、production resolver、repository mode 与 public API 未创建。

## Selection Review

| gate | 结论 | 说明 |
| --- | --- | --- |
| selection review status | `audit_store_storage_adapter_backend_product_selection_review_defined` | storage adapter product class 选择已完成静态评审 |
| selection decision | `storage_adapter_product_class_selected_managed_database_append_only_table_runtime_blocked` | 只选择 product class，不打开 runtime |
| selected backend family | `append_only_metadata_audit_log` | 继承 durable backend family 选择 |
| selected backend product class | `managed_database_append_only_table` | 后续 adapter 以 append-only table 形态作为目标类别 |
| selected backend product profile | `reserved_managed_database_append_only_table_profile` | 仅作为 reserved profile ref，不包含 vendor、endpoint、DSN 或 resource id |
| selection scope | `static_product_class_only` | 不选择具体数据库、对象存储、队列、日志服务或 vendor product |
| storage adapter runtime | `not_created` | 后续仍需独立 runtime entry refresh / task card |
| DB provider / driver / SQL | `blocked / not_selected / not_created` | 不创建连接提供器、driver、DSN parser、migration 或 schema marker |
| audit store runtime | `not_created` | 不创建 audit store runtime task card、writer、delivery 或 idempotency runtime |
| production resolver / repository / API | `not_created / disabled / not_created` | 不打开 production resolver runtime、repository mode 或 public production API |

## Why Managed Database Append-only Table

`managed_database_append_only_table` 是当前 storage adapter 最合适的 product class，原因如下：

- 它与 `append_only_metadata_audit_log` family 对齐，能够承载 ordered metadata record、idempotency key reference、writer result reference、retention policy reference 与 redaction policy reference。
- table 形态便于后续用统一 contract 约束 insert-only、duplicate / replay fail-closed、offline validation、negative leakage scan 与 rollback / recovery evidence。
- 当前阶段尚未选择 vendor、数据库服务、driver、连接生命周期、migration runner 或 schema marker，因此只能选择 product class，不能写成 backend product ready。
- 相比对象存储或日志 sink，append-only table 更适合作为后续 storage adapter runtime task card 的审查对象，但仍必须先完成 runtime entry refresh 和 DB provider / schema / smoke 证据。

## 后续准入条件

进入 storage adapter runtime task card 前，至少还需要独立完成：

- `storage_adapter_runtime_implementation_entry_refresh_after_product_selection`：重新评审 selected product class 后 runtime task card 是否仍 blocked。
- DB provider / driver / DSN / TLS / role policy 选择证据，且不得提交真实 secret、endpoint 或连接串。
- append-only table schema boundary、migration / schema marker boundary、offline adapter smoke strategy 与 negative leakage scan runner boundary。
- writer、idempotency、delivery、operator approval、credential handle、backend health、no leakage smoke 和 production resolver runtime 仍按各自 entry review 解锁。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_backend_product_selection_dependency_missing` | `dependency_chain` | 必需 readiness / materialization / blocker matrix 证据缺失 |
| `audit_store_storage_adapter_backend_product_selection_class_missing` | `selection_gate` | 未固定 selected backend product class |
| `audit_store_storage_adapter_backend_product_selection_invalid_class` | `selection_gate` | 选择了非 allowlist product class |
| `audit_store_storage_adapter_backend_product_selection_vendor_forbidden` | `product_boundary` | 本批选择 vendor、DB product、endpoint、bucket、queue、topic 或 resource id |
| `audit_store_storage_adapter_backend_product_selection_runtime_forbidden` | `runtime_gate` | 本批创建 storage adapter runtime、DB provider、SQL 或 audit store runtime |
| `audit_store_storage_adapter_backend_product_selection_secret_material_detected` | `artifact_guard` | 文档、fixture 或 diagnostics 出现 secret-bearing material |
| `audit_store_storage_adapter_backend_product_selection_scope_overreach` | `implementation_boundary` | 本批打开 production resolver runtime、repository mode 或 public API |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## 停止线

- 不选择 PostgreSQL / MySQL / SQLite、云数据库、对象存储、队列、日志 sink、vendor service、region resource、bucket、topic、table name 或 endpoint。
- 不创建 storage adapter runtime task card、storage adapter runtime、storage adapter client、database connection provider、driver、DSN parser、SQL migration、schema marker、audit store runtime task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、repository mode 或 public production API。
- 不启用 repository mode，不调用 public production API。
- 不写入 raw secret、secret value、token、API key、authorization header、provider raw URL、DSN、database hostname、credential payload、raw request / response / audit / event / writer / storage payload、payload hash、scanner output 或 recovery output。
- 不把本批 static product class selection 写成 storage adapter ready、database ready、durable audit backend ready、audit store ready、production resolver ready 或 production ready。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
