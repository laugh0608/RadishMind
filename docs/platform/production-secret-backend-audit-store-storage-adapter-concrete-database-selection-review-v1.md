# Production Secret Backend Audit Store Storage Adapter Concrete Database Selection Review v1

更新时间：2026-07-04

## 文档目的

本文档在 `Production Secret Backend Audit Store Storage Adapter Concrete Database Selection Readiness v1` 之后，固定 future storage adapter 的具体数据库能力族选择评审。

对应切片：`production-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1`。

结论：状态为 `audit_store_storage_adapter_concrete_database_selection_review_defined`，selection decision 为 `concrete_database_engine_selected_postgresql_compatible_runtime_blocked`。本批只把 future storage adapter 的数据库能力族选择为 `postgresql_compatible_append_only_relational_database`，用于后续 provider / driver / DSN / migration / runtime 任务的输入边界；不选择云厂商、managed product、driver、DSN parser、connection provider、SQL、DDL、物理表名、列名、列类型、schema marker runtime、migration runner、storage adapter runtime、audit store runtime、repository mode 或 public production API。

下一项依赖固定为 `storage_adapter_database_provider_selection_readiness`。该依赖必须继续保持 provider / driver / DSN / connection lifecycle 的独立 readiness 边界，不能因为本批选择数据库能力族而直接创建 runtime task card。

## 输入证据

- `audit_store_storage_adapter_concrete_database_selection_readiness_defined` 已定义 candidate input evidence、metadata-only candidate fields、candidate evaluation dimensions、negative gates 和 artifact guard。
- `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined` 已固定 secret-ref-only DSN policy、TLS mode policy 和 least privilege role policy，但没有选择 driver、provider 或 connection factory。
- `audit_store_storage_adapter_table_schema_artifact_materialized` 已物化 metadata-only logical table schema artifact；该 artifact 仍不包含 SQL、DDL、物理表名、列名或列类型。
- `audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined` 与 `audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined` 已定义 metadata-only smoke / scan 边界，不创建 runner、scanner 或输出。
- `audit_store_runtime_blocker_matrix_defined` 与 `implementation_readiness_defined` 仍确认 audit store runtime、storage adapter runtime、writer / delivery / idempotency runtime、repository mode 和 public production API 未创建。

## Selection Review

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| database selection review | `audit_store_storage_adapter_concrete_database_selection_review_defined` | 固定选择评审结果 |
| selected database engine | `postgresql_compatible_append_only_relational_database` | 选择 PostgreSQL-compatible relational capability family，不绑定 vendor 或 hosted product |
| concrete database selection status | `selected_database_engine_without_vendor_or_provider` | 只选择数据库能力族 |
| runtime task card decision | `storage_adapter_runtime_task_card_still_blocked_after_concrete_database_selection_review` | runtime task card 仍 blocked |
| next dependency | `storage_adapter_database_provider_selection_readiness` | 后续独立定义 provider / product / driver 选择前置条件 |
| database product / vendor | `engine_selected_without_managed_product / not_selected` | 不选择 managed database product 或 vendor |
| DB provider / driver / DSN | `not_created / not_selected / not_defined` | 不创建 provider、driver、DSN parser 或 connection factory |
| SQL / schema marker / migration | `not_created` | 不创建 SQL、DDL、schema marker runtime 或 migration runner |
| storage adapter runtime | `not_created` | 不创建 runtime、client、writer 或 connection |
| audit store runtime | `not_created` | 不创建 audit store runtime task card |
| repository / API | `disabled / not_created` | 不启用 repository mode，不创建 public production API |

## Candidate Evaluation

| candidate | review result | 主要依据 |
| --- | --- | --- |
| `postgresql_compatible_append_only_relational_database` | `selected_engine_class` | 能承接 append-only insert semantics、logical table schema、metadata contract、ordering reference、idempotency reference、retention / redaction reference、TLS / role policy 与后续 migration handoff |
| `generic_key_value_log_store` | `rejected_for_schema_and_query_handoff_gap` | 无法稳定表达 logical field groups、schema marker handoff、idempotency reference 与 operator review 所需的结构化验证链 |
| `object_storage_append_only_log` | `rejected_for_transaction_and_duplicate_handoff_gap` | 难以在当前阶段固定 ordering、transaction boundary、duplicate replay 和 partial write recovery 的 fail-closed 语义 |

本批选择的是数据库能力族，不等于选择 PostgreSQL provider、managed product、driver package、DSN 格式、SQL dialect runtime、schema migration 方案或物理 schema 名称。

## 后续 Provider Readiness 要求

`storage_adapter_database_provider_selection_readiness` 必须重新审查：

- provider / hosted product 候选来源与 deployment model。
- driver selection policy 与 driver import path 的停止线。
- DSN secret ref、TLS mode、role policy 与 connection lifecycle。
- migration runner、schema marker、SQL / DDL 和 physical schema 的独立 artifact guard。
- offline adapter smoke 与 negative leakage runtime scan 对真实 provider 的触发边界。
- rollback / recovery、retention / redaction 与 append-only semantics 的复验方式。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_concrete_database_selection_review_dependency_missing` | `dependency_chain` | 缺少 concrete database selection readiness、database provider policy、table schema artifact、runtime blocker matrix 或 implementation readiness |
| `audit_store_storage_adapter_concrete_database_selection_review_candidate_evidence_missing` | `selection_review` | 候选评估缺少 input evidence、evaluation dimensions 或 rejected candidate reason |
| `audit_store_storage_adapter_concrete_database_selection_review_vendor_or_product_forbidden` | `database_product_boundary` | 本批选择云厂商、managed product、endpoint、resource id 或 account scoped resource |
| `audit_store_storage_adapter_concrete_database_selection_review_driver_or_dsn_forbidden` | `connection_boundary` | 本批选择 driver、DSN parser、connection provider 或 connection lifecycle |
| `audit_store_storage_adapter_concrete_database_selection_review_runtime_forbidden` | `runtime_gate` | 本批创建 SQL、DDL、schema marker、migration runner、storage adapter runtime 或 audit store runtime |
| `audit_store_storage_adapter_concrete_database_selection_review_secret_material_detected` | `artifact_guard` | 文档或 fixture 出现 secret、DSN、endpoint、database detail、raw payload 或 provider detail |
| `audit_store_storage_adapter_concrete_database_selection_review_scope_overreach` | `implementation_boundary` | 本批打开 repository mode、production API、production resolver runtime 或 storage adapter runtime task card |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_storage_adapter_concrete_database_selection_review_status`
- `selection_decision`
- `runtime_task_decision`
- `next_dependency`
- `selected_database_engine`
- `selected_database_engine_status`
- `concrete_database_selection_status`
- `database_product_status`
- `database_vendor_status`
- `database_connection_provider_status`
- `database_driver_status`
- `database_dsn_status`
- `schema_marker_runtime_status`
- `migration_runner_status`
- `storage_adapter_runtime_task_card_status`
- `storage_adapter_runtime_status`
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

禁止输出 raw secret、secret value、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、database hostname、database name、table name、column name、column type、database error detail、cloud credential、credential payload、full secret ref value、full credential handle、raw request / response payload、raw audit payload、raw storage payload、payload hash、secret-derived hash 或 provider error detail。

## No Fallback / No Side Effects

- 不允许把 static product class、reserved profile、metadata contract artifact、logical table schema artifact、offline smoke strategy、negative leakage runtime scan boundary、test-only fake resolver、memory store、mock database provider、historical smoke 或 previous checker success 替代 selection review。
- 不允许把本 review 写成 provider selected、driver selected、DSN ready、SQL ready、storage adapter runtime task card ready、DB provider ready、audit store ready、production resolver ready、repository mode ready、public API ready 或 production ready。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / review fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不创建 storage adapter runtime、不连接数据库、不打开 driver、不运行 SQL、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1.py`

不得新增或启用 database provider implementation task card、storage adapter runtime implementation task card、storage adapter runtime、storage adapter client、durable audit backend runtime、database connection provider、DB driver、DSN parser、SQL migration、DDL、physical table schema、schema marker reader / writer、migration runner、audit store runtime implementation task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批之后 storage adapter runtime task card 仍 blocked。若继续 audit store durable backend，应先做 `storage_adapter_database_provider_selection_readiness`，并把 provider / hosted product / driver / DSN / SQL / migration / runtime 的停止线重新写成独立证据；不得跳过 provider readiness 直接创建 DB provider、storage adapter runtime、SQL、schema marker、audit store runtime task card 或 production resolver runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
