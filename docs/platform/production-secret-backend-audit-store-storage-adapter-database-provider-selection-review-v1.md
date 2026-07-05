# Production Secret Backend Audit Store Storage Adapter Database Provider Selection Review v1

更新时间：2026-07-05

## 文档目的

本文档在 `Production Secret Backend Audit Store Storage Adapter Database Provider Selection Readiness v1` 之后，固定 future storage adapter 的 database provider 选择评审结论。

对应切片：`production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1`。

结论：状态为 `audit_store_storage_adapter_database_provider_selection_review_defined`，selection decision 为 `database_provider_candidate_class_selected_managed_postgresql_compatible_service_runtime_blocked`。本批只把 provider deployment model / candidate class 选择为 `managed_postgresql_compatible_service`，用于后续 driver、connection lifecycle、provider contract 和 runtime 前置评审；不选择云厂商、托管产品、provider、account scoped resource、endpoint、region、driver、DSN parser、connection provider、SQL、DDL、物理表名、列名、列类型、schema marker runtime、migration runner、storage adapter runtime、audit store runtime、repository mode 或 public production API。

下一项依赖固定为 `storage_adapter_database_driver_selection_readiness`。该依赖必须继续保持 driver / DSN / connection provider / physical schema / migration 的独立 readiness 边界，不能因为本批选择 managed PostgreSQL-compatible provider class 而直接创建 DB provider、driver、SQL 或 runtime task card。

## 输入证据

- `audit_store_storage_adapter_database_provider_selection_readiness_defined` 已定义 provider / hosted product 候选输入证据、候选类别、评估维度、停止线和 artifact guard。
- `audit_store_storage_adapter_concrete_database_selection_review_defined` 已把数据库能力族选择为 `postgresql_compatible_append_only_relational_database`，但明确不选择 vendor、managed product、provider、driver 或 DSN。
- `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined` 已固定 secret-ref-only DSN policy、TLS mode policy 和 least privilege role policy；这些仍只是选择约束，不是 driver 或 connection runtime。
- `audit_store_storage_adapter_table_schema_artifact_materialized` 与 `audit_store_storage_adapter_metadata_contract_artifact_materialized` 已提供 metadata-only logical schema / contract artifact；它们不包含 SQL、DDL、物理表名、列名或列类型。
- `audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined` 与 `audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined` 已固定 metadata-only smoke / scan 边界，不创建 runner、scanner 或输出。
- `audit_store_runtime_blocker_matrix_defined` 与 `implementation_readiness_defined` 仍确认 storage adapter runtime、audit store runtime、DB provider、SQL、schema marker、repository mode 和 public production API 未创建。

## Selection Review

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| database provider selection review | `audit_store_storage_adapter_database_provider_selection_review_defined` | provider candidate class 选择评审已完成 |
| selection decision | `database_provider_candidate_class_selected_managed_postgresql_compatible_service_runtime_blocked` | 只选择 managed PostgreSQL-compatible provider class |
| selected database engine | `postgresql_compatible_append_only_relational_database` | 继承上一批能力族选择 |
| selected provider candidate class | `managed_postgresql_compatible_service` | 选择托管 PostgreSQL-compatible 服务类别，不绑定 vendor 或 hosted product |
| provider selection status | `selected_provider_candidate_class_without_vendor_or_product` | 不选择具体 provider、vendor、product、resource 或 endpoint |
| runtime task card decision | `storage_adapter_runtime_task_card_still_blocked_after_database_provider_selection_review` | runtime task card 仍 blocked |
| next dependency | `storage_adapter_database_driver_selection_readiness` | 后续独立定义 driver / DSN / connection 选择前置 |
| managed database product | `not_selected` | 不选择具体托管产品 |
| database vendor / provider | `not_selected / not_selected` | 不绑定云厂商、服务商或 provider |
| provider account resource / endpoint | `not_defined` | 不提交 resource id、endpoint、hostname、account 或 region detail |
| driver / DSN / connection provider | `not_selected / not_defined / not_created` | 不选择 driver，不定义 DSN，不创建连接 |
| SQL / migration / schema marker | `not_created` | 不创建 SQL、DDL、schema marker runtime 或 migration runner |
| storage adapter runtime | `not_created` | 不创建 runtime、client、writer 或 connection |
| audit store runtime | `not_created` | 不创建 audit store runtime task card |
| repository / API | `disabled / not_created` | 不启用 repository mode，不创建 public production API |

## Candidate Evaluation

| candidate class | review result | 主要依据 |
| --- | --- | --- |
| `managed_postgresql_compatible_service` | `selected_candidate_class_without_vendor_or_product` | 与 PostgreSQL-compatible 能力族、append-only table product class、secret-ref-only credential policy、TLS / role policy、operator review、backup / restore 和 rollout / rollback 证据链最一致 |
| `self_managed_postgresql_compatible_cluster` | `deferred_for_operations_ownership_and_recovery_burden` | 需要更重的升级窗口、备份恢复、运维责任、故障演练和本地 / 生产隔离证据；当前阶段不作为首选类别 |
| `embedded_or_file_database` | `rejected_for_production_audit_store_boundary` | 不满足 production audit store 的 durability、operator review、connection lifecycle、multi-tenant isolation 和 recovery 边界 |

本批选择的是 provider candidate class，不等于选择具体 cloud provider、managed database product、hosted product、provider account、driver package、DSN 格式、SQL dialect runtime、schema migration 方案或物理 schema 名称。

## 后续 Driver Readiness 要求

`storage_adapter_database_driver_selection_readiness` 必须重新审查：

- PostgreSQL-compatible driver candidate source、import boundary 和 license / support risk。
- DSN secret ref、credential handle、TLS mode、certificate verification、least privilege role 与 connection lifecycle。
- pooling、timeout、retry、partial write recovery、duplicate / replay fail-closed 和 sanitized error。
- migration runner、schema marker、physical schema handoff、SQL / DDL artifact guard。
- offline adapter smoke、negative leakage runtime scan、operator review、rollout / rollback 和 local/test/prod environment parity。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_database_provider_selection_review_dependency_missing` | `dependency_chain` | 缺少 provider selection readiness、concrete database selection review、provider policy、schema artifact、smoke / scan boundary、blocker matrix 或 implementation readiness |
| `audit_store_storage_adapter_database_provider_selection_review_candidate_evidence_missing` | `selection_review` | 候选评估缺少 input evidence、evaluation dimensions 或 rejected candidate reason |
| `audit_store_storage_adapter_database_provider_selection_review_vendor_or_product_forbidden` | `provider_selection_boundary` | 本批选择云厂商、managed product、hosted product、endpoint、resource id、region detail 或 account scoped resource |
| `audit_store_storage_adapter_database_provider_selection_review_driver_or_dsn_forbidden` | `connection_boundary` | 本批选择 driver、DSN parser、connection provider 或 connection lifecycle |
| `audit_store_storage_adapter_database_provider_selection_review_runtime_forbidden` | `runtime_gate` | 本批创建 SQL、DDL、schema marker、migration runner、storage adapter runtime 或 audit store runtime |
| `audit_store_storage_adapter_database_provider_selection_review_secret_material_detected` | `artifact_guard` | 文档或 fixture 出现 raw secret、DSN、endpoint、database detail、credential payload 或 provider detail |
| `audit_store_storage_adapter_database_provider_selection_review_scope_overreach` | `implementation_boundary` | 本批打开 repository mode、production API、production resolver 或 runtime task card |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_storage_adapter_database_provider_selection_review_status`
- `selection_decision`
- `runtime_task_decision`
- `next_dependency`
- `selected_database_engine`
- `selected_provider_candidate_class`
- `provider_selection_status`
- `managed_database_product_status`
- `database_vendor_status`
- `database_provider_status`
- `provider_account_resource_status`
- `database_endpoint_status`
- `database_driver_status`
- `database_dsn_status`
- `connection_provider_status`
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

- 不允许把 PostgreSQL-compatible 能力族选择、static product class、reserved profile、metadata contract artifact、logical table schema artifact、offline smoke strategy、negative leakage runtime scan boundary、test-only fake resolver、memory store、mock database provider、historical smoke 或 previous checker success 替代 provider selection review。
- 不允许把本 review 写成 cloud provider selected、hosted product selected、driver selected、DSN ready、SQL ready、connection ready、storage adapter runtime task card ready、DB provider ready、audit store ready、production resolver ready、repository mode ready、public API ready 或 production ready。
- 不创建 storage adapter runtime、storage adapter runtime task card、storage adapter client、audit store runtime、DB provider、database connection provider、SQL、DDL、schema marker runtime 或 migration runner。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / review fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不选择托管产品、不创建 DB provider、不打开 driver、不解析 DSN、不运行 SQL、不创建 schema marker、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1.py`

不得新增或启用 database provider implementation task card、storage adapter runtime implementation task card、storage adapter runtime、storage adapter client、durable audit backend runtime、database connection provider、DB driver、DSN parser、SQL migration、DDL、physical table schema、schema marker reader / writer、migration runner、audit store runtime implementation task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批之后 storage adapter runtime task card 仍 blocked。若继续 audit store durable backend，应先做 `storage_adapter_database_driver_selection_readiness`，基于本批 provider class 选择结论单独定义 driver / DSN / connection provider / SQL / schema marker / migration 的选择前证据；不得跳过 driver readiness 直接创建 DB provider、storage adapter runtime、SQL、schema marker、audit store runtime task card 或 production resolver runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
