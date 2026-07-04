# Production Secret Backend Audit Store Storage Adapter Table Schema Artifact Materialization v1

更新时间：2026-07-04

## 文档目的

本文档承接 `Production Secret Backend Audit Store Storage Adapter Table Schema Artifact Materialization Entry Review v1` 和已定义的物化任务卡，固定 storage adapter table schema artifact 的 metadata-only 物化结果。

对应切片：`production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1`。

结论：状态为 `audit_store_storage_adapter_table_schema_artifact_materialized`。本批创建 `contracts/production-secret-audit-storage-adapter.table-schema.json`，schema version 固定为 `audit-storage-adapter-table-schema-v1`，并补齐 positive / negative fixtures、离线 checker、metadata contract compatibility smoke、no secret material scan 和 artifact guard。该 artifact 只描述 future storage adapter 的 logical append-only table schema，不创建 SQL、DDL、物理表名、列名、列类型、索引、约束、partition、sequence、trigger、function、schema marker runtime、migration runner、storage adapter runtime、DB provider、audit store runtime、repository mode 或 public production API。

下一项依赖固定为 `storage_adapter_offline_adapter_smoke_strategy_readiness`。runtime task card 仍保持 blocked；后续必须先独立定义 offline adapter smoke strategy 与 negative leakage runtime scan boundary，不能从 table schema artifact 直接跳到 runtime。

## 输入证据

- `audit_store_storage_adapter_table_schema_artifact_materialization_entry_review_defined` 已确认 artifact 物化任务可独立创建。
- `audit_store_storage_adapter_append_only_table_schema_boundary_readiness_defined` 已固定 logical table schema、field group、record identity、sequence reference、idempotency reference、retention / redaction reference 和 schema marker handoff。
- `audit_store_storage_adapter_metadata_contract_artifact_materialized` 已物化 metadata-only storage adapter contract、positive / negative fixtures、writer compatibility smoke 和 no secret material scan。
- `audit_store_storage_adapter_backend_product_selection_review_defined` 只静态选择 product class `managed_database_append_only_table` / `reserved_managed_database_append_only_table_profile`，不选择具体数据库产品。
- `audit_store_runtime_blocker_matrix_defined` 与 `implementation_readiness_defined` 已同步 table schema artifact 的静态证据，同时继续阻断 runtime。

## Artifact Boundary

| gate | 本次状态 | 说明 |
| --- | --- | --- |
| table schema artifact | `materialized_static_logical_table_schema` | 已创建 metadata-only JSON Schema artifact |
| artifact path | `contracts/production-secret-audit-storage-adapter.table-schema.json` | 稳定 committed contract path |
| table schema version | `audit-storage-adapter-table-schema-v1` | 静态版本 pin，不从数据库 introspection 派生 |
| metadata contract version | `audit-storage-adapter-metadata-contract-v1` | 与 metadata contract artifact 对齐 |
| logical field groups | `from_append_only_table_schema_boundary` | 只继承 logical group，不创建物理列 |
| positive fixture | `implemented_static_fixture` | 覆盖 metadata-only append-only audit record shape |
| negative fixtures | `implemented_static_fixtures` | 覆盖缺字段、物理细节、secret material 和 additionalProperties |
| metadata contract compatibility | `implemented_static_contract_compatibility` | 校验 artifact 字段覆盖 contract envelope |
| no secret material scan | `implemented_static_scan` | committed artifact / fixtures / checker 不含 secret、DSN 或 raw payload |
| SQL / DDL | `not_created` | 本批不创建 SQL、DDL 或 migration |
| schema marker runtime | `not_created` | 本批不创建 runtime |
| storage adapter runtime | `not_created` | runtime task card 仍 blocked |
| next dependency | `storage_adapter_offline_adapter_smoke_strategy_readiness` | 下一批进入 offline smoke strategy readiness |

## Schema Coverage

`contracts/production-secret-audit-storage-adapter.table-schema.json` 是 JSON Schema artifact，不是 SQL schema。它固定以下 logical field groups：

- `identity`：record / audit / idempotency / policy version reference。
- `ordering`：append-only sequence、storage adapter result 和 write status reference。
- `payload_reference`：audit event、metadata contract、writer result 与 append-only contract reference。
- `retention_redaction`：retention 与 redaction policy reference。
- `delivery_recovery`：delivery attempt、delivery commit 与 dedupe decision reference。
- `diagnostics`：request、backend product evidence、failure boundary 和 sanitized diagnostic metadata。

artifact 明确拒绝 secret value、credential payload、provider raw URL、DSN、database / table / column 物理细节、SQL、DDL、migration、raw request / response / audit / writer / storage payload、payload hash、secret-derived hash、schema marker output 和 migration output。

## Validation Fixtures

正例：

- `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-positive-v1.json`

负例：

- `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-missing-required-negative-v1.json`
- `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-physical-detail-negative-v1.json`
- `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-secret-material-negative-v1.json`
- `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-additional-properties-negative-v1.json`

这些 fixture 只用于离线 JSON Schema 校验，不连接数据库、不读取环境 secret、不创建 runtime output。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_table_schema_artifact_missing` | `artifact_guard` | 缺少 table schema artifact |
| `audit_store_storage_adapter_table_schema_version_mismatch` | `schema_contract` | schema version 未固定为 `audit-storage-adapter-table-schema-v1` |
| `audit_store_storage_adapter_table_schema_logical_group_drift` | `logical_schema` | logical field group 未对齐 append-only boundary |
| `audit_store_storage_adapter_table_schema_contract_compatibility_drift` | `contract_compatibility` | metadata contract envelope 字段覆盖不完整或漂移 |
| `audit_store_storage_adapter_table_schema_fixture_drift` | `schema_validation` | positive / negative fixture 未按预期校验 |
| `audit_store_storage_adapter_table_schema_secret_material_detected` | `artifact_guard` | artifact、fixture 或 checker 出现 secret / DSN / raw payload material |
| `audit_store_storage_adapter_table_schema_physical_detail_detected` | `schema_runtime_boundary` | artifact 声明 SQL、DDL、物理表名、列名、列类型或 migration |
| `audit_store_storage_adapter_table_schema_runtime_created` | `implementation_boundary` | 本批创建 schema marker runtime、migration runner、storage adapter runtime、DB provider、audit store runtime、repository mode 或 production API |

所有 failure 必须 fail closed，并只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `table_schema_artifact_path_status`
- `table_schema_artifact_version_status`
- `logical_field_group_coverage_status`
- `metadata_contract_compatibility_status`
- `schema_validation_status`
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

- 不允许 table schema artifact fallback 到 append-only readiness、metadata contract artifact、backend product selection review、memory store、test-only fake resolver、static sample、mock provider、historical smoke、SQL introspection、driver output、database migration draft 或 payload-derived schema。
- 不允许缺少 offline adapter smoke strategy、negative leakage runtime scan boundary、schema marker runtime、migration runner、storage adapter runtime、audit writer runtime、delivery runtime 或 idempotency runtime 时创建 audit store runtime success。
- 本批 checker 只读取 committed artifact、fixture、文档、已有 readiness / entry review fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不连接数据库、不打开 driver、不运行 SQL、不创建 schema marker runtime、不创建 migration runner、不创建 storage adapter runtime、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增或修改以下静态证据：

- `contracts/production-secret-audit-storage-adapter.table-schema.json`
- `docs/contracts/production-secret-audit-storage-adapter-table-schema.md`
- `docs/platform/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1-plan.md`
- `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-positive-v1.json`
- `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-missing-required-negative-v1.json`
- `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-physical-detail-negative-v1.json`
- `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-secret-material-negative-v1.json`
- `scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-additional-properties-negative-v1.json`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.py`

不得新增或启用 database product artifact、database vendor artifact、DB provider、database connection provider、DB driver、DSN parser、SQL migration、DDL、schema marker reader / writer、migration runner、schema version table、storage adapter runtime task card、storage adapter runtime、storage adapter client、audit store runtime task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、repository mode runtime 或 public production API。

## 后续推进

下一步应进入 `storage_adapter_offline_adapter_smoke_strategy_readiness`：定义 offline adapter smoke strategy 的 static manifest、fixture 消费关系、failure mapping、no side effects 与 artifact guard。该后续批次仍不得创建 SQL、DDL、schema marker runtime、migration runner、storage adapter runtime、DB provider、audit store runtime、repository mode 或 public production API。

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
