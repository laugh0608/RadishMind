# Production Secret Backend Audit Store Storage Adapter Append-only Table Schema Boundary Readiness v1

更新时间：2026-07-02

## 文档目的

本文档在 `Production Secret Backend Audit Store Storage Adapter Database Provider / Driver / DSN / TLS / Role Policy Readiness v1` 之后，定义 future storage adapter runtime task card 之前必须具备的 append-only table schema 静态边界。

对应切片：`production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1`。

结论：状态为 `audit_store_storage_adapter_append_only_table_schema_boundary_readiness_defined`，readiness decision 为 `append_only_table_schema_boundary_defined_without_sql_or_runtime`。本批只定义 logical table schema、field group、record identity、sequence reference、idempotency reference、retention / redaction reference、marker handoff、sanitized diagnostics、no fallback 和 no side effects。仍不选择具体数据库、vendor、resource id、endpoint、table name、driver 或 column type；也不创建 DB provider、DSN parser、connection factory、SQL migration、schema marker runtime、migration runner、storage adapter runtime、audit store runtime、repository mode 或 public production API。

下一项依赖固定为 `storage_adapter_table_schema_artifact_materialization_entry_review`。该步骤应评审是否可以物化静态 schema artifact，不得跳过 artifact review 直接创建 SQL、migration runner、schema marker runtime 或 storage adapter runtime。

## 输入证据

- `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined` 已确认 provider boundary、driver policy、DSN secret-ref policy、TLS policy 和 least privilege role policy 只达到 metadata-only 准入。
- `audit_store_storage_adapter_metadata_contract_artifact_materialized` 已提供 metadata-only contract artifact、positive / negative fixtures、writer compatibility smoke 和 no secret material scan。
- `audit_store_storage_adapter_backend_product_selection_review_defined` 只选择 static product class `managed_database_append_only_table` 与 reserved profile，不选择具体数据库或 vendor。
- `audit_store_runtime_blocker_matrix_defined` 与 `implementation_readiness_defined` 仍保持 storage adapter runtime、audit store runtime、DB provider、SQL、schema marker、repository mode 和 production API 未创建。

## Readiness Boundary

| gate | 本批结论 | 说明 |
| --- | --- | --- |
| append-only table schema boundary readiness | `audit_store_storage_adapter_append_only_table_schema_boundary_readiness_defined` | 准入证据已定义 |
| readiness decision | `append_only_table_schema_boundary_defined_without_sql_or_runtime` | 只定义 logical schema 边界，不创建 SQL 或 runtime |
| next dependency | `storage_adapter_table_schema_artifact_materialization_entry_review` | 下一项评审静态 schema artifact materialization |
| logical table schema | `logical_append_only_table_schema_boundary_defined` | 只定义 field group 和语义，不定义 table name / column type |
| record identity boundary | `logical_record_identity_boundary_defined` | 固定 record key、audit ref、event ref、idempotency ref 的静态职责 |
| sequence boundary | `logical_sequence_reference_boundary_defined` | 固定 monotonic / append-only sequence reference，不创建 sequence generator |
| idempotency boundary | `logical_idempotency_reference_boundary_defined` | 固定 duplicate / replay reference，不创建 key store |
| retention / redaction boundary | `logical_retention_redaction_reference_boundary_defined` | 固定 retention / redaction reference，不允许 inline erasure |
| marker handoff boundary | `logical_schema_marker_handoff_boundary_defined` | 固定 future migration marker handoff，不创建 marker runtime |
| SQL / migration | `not_created` | 不创建 DDL、migration runner、schema marker reader / writer |
| storage adapter runtime | `not_created` | 不创建 runtime、client、writer 或 DB path |
| repository / API | `disabled / not_created` | 不启用 repository mode，不创建 public production API |

## Logical Schema Scope

本批只允许定义以下 logical field groups：

- identity：`record_ref`、`audit_ref`、`event_ref`、`idempotency_ref`、`policy_version_ref`。
- ordering：`sequence_ref`、`created_at_ref`、`producer_ref`。
- payload reference：`payload_ref`、`payload_schema_ref`、`payload_redaction_ref`。
- retention / redaction：`retention_policy_ref`、`redaction_policy_ref`、`retention_window_ref`。
- delivery / recovery：`delivery_ref`、`recovery_ref`、`duplicate_replay_ref`。
- diagnostics：`failure_code`、`failure_boundary`、`sanitized_diagnostic`、`request_id`。

这些 field group 只能作为静态语义边界进入文档、fixture 和 checker；不得声明真实 table name、database name、column type、index DDL、constraint DDL、partition policy、database sequence、trigger、function、schema version table 或 migration command。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_append_only_table_schema_boundary_dependency_missing` | `dependency_chain` | 缺少 database provider policy readiness、metadata contract artifact、blocker matrix 或 implementation readiness |
| `audit_store_storage_adapter_append_only_table_schema_boundary_sql_forbidden` | `schema_runtime_boundary` | 本批创建 SQL、DDL、migration、index、constraint、sequence generator、trigger 或 function |
| `audit_store_storage_adapter_append_only_table_schema_boundary_database_detail_forbidden` | `database_product_boundary` | 本批选择具体数据库、vendor、resource id、endpoint、table name、database name 或 column type |
| `audit_store_storage_adapter_append_only_table_schema_boundary_runtime_forbidden` | `implementation_boundary` | 本批创建 DB provider、connection provider、schema marker runtime、migration runner、storage adapter runtime 或 audit store runtime |
| `audit_store_storage_adapter_append_only_table_schema_boundary_secret_material_detected` | `artifact_guard` | 文档或 fixture 出现 secret、DSN、endpoint、credential、raw URL、raw payload 或 database detail |
| `audit_store_storage_adapter_append_only_table_schema_boundary_artifact_review_bypassed` | `artifact_guard` | 本批跳过后续 static schema artifact materialization entry review |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_storage_adapter_append_only_table_schema_boundary_readiness_status`
- `readiness_decision`
- `next_dependency`
- `selected_backend_product_class`
- `selected_backend_product_profile`
- `logical_table_schema_status`
- `record_identity_boundary_status`
- `sequence_reference_boundary_status`
- `idempotency_reference_boundary_status`
- `retention_redaction_reference_boundary_status`
- `schema_marker_handoff_boundary_status`
- `storage_adapter_runtime_task_card_status`
- `storage_adapter_runtime_status`
- `audit_store_runtime_status`
- `repository_mode_status`
- `production_api_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、secret value、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、database hostname、database name、table name、column name、column type、index name、constraint name、database error detail、cloud credential、credential payload、full secret ref value、full credential handle、raw request / response payload、raw audit payload、raw storage payload、payload hash、secret-derived hash 或 provider error detail。

## No Fallback / No Side Effects

- 不允许把 metadata contract artifact、writer compatibility fixture、test-only fake resolver、memory store、mock database provider、historical smoke 或 previous checker success 替代 append-only table schema boundary readiness。
- 不允许把本 readiness 写成具体 database ready、table schema artifact ready、SQL ready、migration ready、schema marker ready、DB provider ready、connection ready、storage adapter runtime ready、audit store ready、production resolver ready、repository mode ready、public API ready 或 production ready。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / review fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不选择数据库、不创建 DB provider、不打开 driver、不解析 DSN、不运行 SQL、不创建 schema marker、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1.py`

不得新增或启用 database product selection artifact、backend vendor selection artifact、database provider implementation task card、storage adapter runtime implementation task card、database connection provider、DB driver、DSN parser、connection factory、SQL migration、schema marker reader / writer、migration runner、schema version table、storage adapter runtime、storage adapter client、audit store runtime implementation task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、repository mode runtime 或 public production API。

## 后续推进

本批之后 storage adapter runtime task card 仍 blocked。若继续 audit store durable backend，应推进 `storage_adapter_table_schema_artifact_materialization_entry_review`，评审是否可以物化 metadata-only table schema artifact、positive / negative fixtures 和 no secret material scan；不得跳过 artifact review 直接创建 SQL migration、schema marker runtime、storage adapter runtime、audit store runtime task card 或 production resolver runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
