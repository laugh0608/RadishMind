# Production Secret Backend Audit Store Storage Adapter Table Schema Artifact Materialization v1

更新时间：2026-07-04

## 文档目的

本文档承接 `Production Secret Backend Audit Store Storage Adapter Table Schema Artifact Materialization Entry Review v1`，固定后续 table schema artifact materialization 的静态任务卡边界。

对应切片：`production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1`。

结论：状态为 `audit_store_storage_adapter_table_schema_artifact_materialization_task_card_defined`。本批只创建 metadata-only table schema artifact materialization task card 的证据链，固定 future artifact path、schema version pin、logical field group coverage、metadata contract envelope compatibility、positive / negative fixture 要求、no secret material scan、artifact guard 和停止线；不创建 `contracts/production-secret-audit-storage-adapter.table-schema.json`，不创建 positive / negative fixture，不创建 checker output，不创建 SQL、DDL、物理表名、列类型、schema marker runtime、migration runner、storage adapter runtime、DB provider、audit store runtime、repository mode 或 public production API。

下一项依赖固定为 `storage_adapter_table_schema_artifact_materialization`。后续真正物化 artifact 时，仍只能创建 metadata-only table schema artifact 与离线校验资产，不得选择具体数据库 / vendor 或创建 runtime。

## 输入证据

- `audit_store_storage_adapter_table_schema_artifact_materialization_entry_review_defined` 已确认物化任务卡可独立创建。
- `audit_store_storage_adapter_append_only_table_schema_boundary_readiness_defined` 已固定 logical table schema、field group、record identity、sequence reference、idempotency reference、retention / redaction reference 和 schema marker handoff。
- `audit_store_storage_adapter_metadata_contract_artifact_materialized` 已物化 metadata-only storage adapter contract、positive / negative fixtures、writer compatibility smoke 和 no secret material scan。
- `audit_store_storage_adapter_backend_product_selection_review_defined` 只静态选择 product class `managed_database_append_only_table` / `reserved_managed_database_append_only_table_profile`，不选择具体数据库产品。
- `audit_store_runtime_blocker_matrix_defined` 与 `implementation_readiness_defined` 仍保持 table schema artifact、SQL、schema marker runtime、migration runner、storage adapter runtime、audit store runtime、repository mode 和 production API 未创建。

## Task Card Boundary

| gate | 本次状态 | 说明 |
| --- | --- | --- |
| task card status | `created_static_task_card` | 只创建后续 artifact materialization 的任务卡证据 |
| task card decision | `table_schema_artifact_materialization_task_card_defined_after_entry_review` | entry review 已消费 |
| future table schema artifact path | `contracts/production-secret-audit-storage-adapter.table-schema.json` | 后续 artifact path proposal |
| table schema artifact version | `audit-storage-adapter-table-schema-v1` | 后续 artifact 必须固定静态版本 |
| logical field groups | `from_append_only_table_schema_boundary` | 只继承 logical group，不创建物理列 |
| metadata contract compatibility | `required_for_next_artifact_batch` | 后续 artifact 必须对齐 metadata contract envelope |
| validation plan | `required_for_next_artifact_batch` | 后续批次必须覆盖正例、负例和 no secret material scan |
| table schema artifact | `not_created` | 本批不生成 artifact |
| SQL / DDL | `not_created` | 本批不创建 SQL、DDL 或 migration |
| schema marker runtime | `not_created` | 本批不创建 runtime |
| storage adapter runtime | `not_created` | runtime task card 仍 blocked |
| next dependency | `storage_adapter_table_schema_artifact_materialization` | 下一批才能物化 artifact |

## Future Artifact Requirements

后续真正创建 `contracts/production-secret-audit-storage-adapter.table-schema.json` 的实现批次必须满足：

- artifact path 固定为 `contracts/production-secret-audit-storage-adapter.table-schema.json`，若命名调整必须同步 contract 入口、task card、fixture、checker 和 docs 索引。
- schema version pin 固定为 `audit-storage-adapter-table-schema-v1`，不能从数据库 introspection、SQL、driver output、raw storage payload 或 migration output 派生。
- logical field groups 必须逐项来自 append-only table schema boundary readiness。
- logical field keys 必须对齐 metadata contract envelope，不得写入物理 table name、column name、column type、index、constraint、partition、sequence、trigger 或 function。
- positive fixture 必须覆盖 metadata-only append-only audit record schema shape。
- negative fixtures 必须覆盖 physical table detail、raw storage payload、DSN / endpoint、SQL / DDL、payload hash、secret-derived hash 和 additionalProperties。
- artifact checker 必须验证 artifact path、version pin、logical groups、metadata contract compatibility、positive fixture、negative fixture、no secret material scan 和 check-repo 注册顺序。
- writer compatibility smoke 仍只做静态 metadata compatibility，不创建 writer runtime、storage adapter runtime 或 audit store runtime。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_table_schema_materialization_task_card_missing` | `task_card` | 缺少本 materialization task card |
| `audit_store_storage_adapter_table_schema_materialization_scope_missing` | `implementation_boundary` | 未固定 artifact path、version、logical groups、fixtures、checker 或停止线 |
| `audit_store_storage_adapter_table_schema_materialization_field_group_missing` | `logical_schema` | logical field group coverage 未对齐 append-only boundary |
| `audit_store_storage_adapter_table_schema_materialization_contract_compatibility_missing` | `contract_compatibility` | 未要求 metadata contract envelope compatibility |
| `audit_store_storage_adapter_table_schema_materialization_validation_plan_missing` | `schema_validation` | 缺少正例、负例、checker 或 no secret material scan 要求 |
| `audit_store_storage_adapter_table_schema_artifact_created_in_task_card` | `artifact_guard` | 本批创建 table schema artifact |
| `audit_store_storage_adapter_table_schema_physical_detail_created_in_task_card` | `schema_runtime_boundary` | 本批创建 SQL、DDL、物理表名、列名、列类型、索引、约束或 migration |
| `audit_store_storage_adapter_table_schema_runtime_created_in_task_card` | `implementation_boundary` | 本批创建 schema marker runtime、migration runner、storage adapter runtime、DB provider、audit store runtime、repository mode 或 production API |
| `audit_store_storage_adapter_table_schema_materialization_fallback_forbidden` | `no_fallback` | 缺少 artifact 时 fallback 到 readiness、sample、memory store、fake resolver、SQL introspection 或 payload-derived schema |
| `audit_store_storage_adapter_table_schema_materialization_secret_material_detected` | `artifact_guard` | 文档、fixture 或 checker 出现 secret / DSN / endpoint / raw payload 等敏感材料 |

所有 failure 必须 fail closed，并只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_storage_adapter_table_schema_artifact_materialization_task_card_status`
- `table_schema_artifact_path_status`
- `table_schema_artifact_version_status`
- `logical_field_group_coverage_status`
- `metadata_contract_compatibility_status`
- `schema_validation_plan_status`
- `table_schema_artifact_status`
- `positive_fixture_status`
- `negative_fixture_status`
- `no_secret_material_scan_status`
- `storage_adapter_runtime_task_card_status`
- `storage_adapter_runtime_status`
- `schema_marker_runtime_status`
- `migration_runner_status`
- `audit_store_runtime_status`
- `production_resolver_runtime_status`
- `repository_mode_status`
- `production_api_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、secret value、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、database hostname、database name、table name、column name、column type、index name、constraint name、partition name、SQL、DDL、migration command、database error detail、cloud credential、credential payload、full secret ref value、full credential handle、raw request / response payload、raw audit payload、raw writer payload、raw storage payload、payload hash、secret-derived hash、provider error detail、scanner raw finding、scan output、schema marker raw output 或 migration output。

## No Fallback / No Side Effects

- 不允许 materialization task card fallback 到 append-only readiness、metadata contract artifact、backend product selection review、memory store、test-only fake resolver、static sample、mock provider、historical smoke、SQL introspection、driver output、database migration draft 或 payload-derived schema。
- 不允许缺少 table schema artifact、schema marker runtime、migration runner、storage adapter runtime、audit writer runtime、delivery runtime 或 idempotency runtime 时创建 audit store runtime success。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / entry review fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不连接数据库、不打开 driver、不运行 SQL、不创建 table schema artifact、不创建 schema marker runtime、不创建 migration runner、不创建 storage adapter runtime、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.py`

不得新增或启用 `contracts/production-secret-audit-storage-adapter.table-schema.json`、table schema positive / negative fixture、table schema validation checker、database product selection artifact、database vendor selection artifact、DB provider、database connection provider、DB driver、DSN parser、SQL migration、DDL、schema marker reader / writer、migration runner、schema version table、storage adapter runtime task card、storage adapter runtime、storage adapter client、audit store runtime task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、repository mode runtime 或 public production API。

## 后续推进

下一步可进入真正的 table schema artifact materialization 批次：创建 `contracts/production-secret-audit-storage-adapter.table-schema.json`、positive / negative fixtures、schema validation checker、metadata contract compatibility smoke 和 no secret material scan。该后续批次仍不得创建 SQL、DDL、schema marker runtime、migration runner、storage adapter runtime、DB provider、audit store runtime、repository mode 或 public production API。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
