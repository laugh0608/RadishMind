# Production Secret Backend Audit Store Storage Adapter Database Provider Selection Readiness v1

更新时间：2026-07-04

## 文档目的

本文档在 `Production Secret Backend Audit Store Storage Adapter Concrete Database Selection Review v1` 之后，定义 future storage adapter 的 database provider 选择前准入边界。

对应切片：`production-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1`。

结论：状态为 `audit_store_storage_adapter_database_provider_selection_readiness_defined`，readiness decision 为 `database_provider_selection_readiness_defined_without_provider_selection`。本批只定义 provider / hosted product 候选输入证据、候选类别、评估维度、停止线、artifact guard 和 fail-closed diagnostics；不选择云厂商、托管产品、provider、account scoped resource、endpoint、driver、DSN parser、connection provider、SQL、DDL、物理表名、列名、列类型、schema marker runtime、migration runner、storage adapter runtime、audit store runtime、repository mode 或 public production API。

下一项依赖固定为 `storage_adapter_database_provider_selection_review`。该依赖必须基于本批输入证据单独评审 provider / hosted product 候选，不能因为已选择 `postgresql_compatible_append_only_relational_database` 能力族而直接创建 DB provider 或 runtime task card。

## 输入证据

- `audit_store_storage_adapter_concrete_database_selection_review_defined` 已把数据库能力族选择为 `postgresql_compatible_append_only_relational_database`，但明确不选择 vendor、managed product、provider、driver 或 DSN。
- `audit_store_storage_adapter_concrete_database_selection_readiness_defined` 已定义 candidate input evidence、metadata-only candidate fields、evaluation dimensions、negative gates 和 artifact guard。
- `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined` 已固定 secret-ref-only DSN policy、TLS mode policy 和 least privilege role policy；这些仍只是选择约束，不是 driver 或 connection runtime。
- `audit_store_storage_adapter_table_schema_artifact_materialized` 与 `audit_store_storage_adapter_metadata_contract_artifact_materialized` 已提供 metadata-only logical schema / contract artifact；它们不包含 SQL、DDL、物理表名、列名或列类型。
- `audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined` 与 `audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined` 已固定 metadata-only smoke / scan 边界，不创建 runner、scanner 或输出。
- `audit_store_runtime_blocker_matrix_defined` 与 `implementation_readiness_defined` 仍确认 storage adapter runtime、audit store runtime、DB provider、SQL、schema marker、repository mode 和 public production API 未创建。

## Readiness Boundary

| gate | 本批结论 | 说明 |
| --- | --- | --- |
| database provider selection readiness | `audit_store_storage_adapter_database_provider_selection_readiness_defined` | provider 选择前准入证据已定义 |
| readiness decision | `database_provider_selection_readiness_defined_without_provider_selection` | 只定义选择前边界，不选择 provider |
| runtime task card decision | `storage_adapter_runtime_task_card_still_blocked_after_database_provider_selection_readiness` | runtime task card 仍 blocked |
| next dependency | `storage_adapter_database_provider_selection_review` | 下一项单独评审 provider / hosted product |
| selected database engine | `postgresql_compatible_append_only_relational_database` | 继承上一批能力族选择 |
| provider candidate source | `metadata_only_provider_candidate_source_defined` | 只允许 metadata-only 来源说明 |
| provider input evidence | `metadata_only_provider_input_evidence_defined` | 只记录候选证据引用和评估所需字段 |
| provider evaluation dimensions | `metadata_only_provider_evaluation_dimensions_defined` | 固定评估维度，不做产品选择 |
| provider selection status | `readiness_defined_without_provider_selection` | 不选择 provider、vendor 或 hosted product |
| managed database product | `not_selected` | 不选择具体托管产品 |
| database vendor | `not_selected` | 不绑定云厂商或服务商 |
| provider account resource / endpoint | `not_defined` | 不提交 resource id、endpoint、hostname 或区域 |
| driver / DSN / connection provider | `not_selected / not_defined / not_created` | 不选择 driver，不定义 DSN，不创建连接 |
| SQL / migration / schema marker | `not_created` | 不创建 SQL、DDL、schema marker runtime 或 migration runner |
| storage adapter runtime | `not_created` | 不创建 runtime、client、writer 或 connection |
| audit store runtime | `not_created` | 不创建 audit store runtime task card |
| repository / API | `disabled / not_created` | 不启用 repository mode，不创建 public production API |

## Provider 候选输入证据

每个后续候选 provider / hosted product 只能提交 metadata-only 证据：

- `candidate_key`：稳定短键，不写厂商账号、项目号、资源 id、endpoint 或 region detail。
- `deployment_model`：托管服务、自管集群或本地替身类别；本批只定义类别，不选择实例。
- `compatibility_claim_ref`：PostgreSQL-compatible 能力证明引用，不提交产品页面长摘录或运行时探测输出。
- `append_only_capability_ref`：append-only insert、transaction boundary、duplicate / replay fail-closed 的证据引用。
- `secret_ref_policy_ref`：DSN / credential 只能通过 secret ref 或 credential handle 进入 future runtime 的证据引用。
- `tls_role_policy_ref`：TLS mode、certificate / verification policy、least privilege role 拆分证据引用。
- `migration_schema_marker_ref`：schema marker、migration boundary 和 physical schema handoff 的证据引用。
- `backup_restore_recovery_ref`：backup / restore、partial write recovery、rollback / recovery 与 retention / redaction 的兼容证据引用。
- `offline_validation_ref`：offline adapter smoke、negative leakage runtime scan 和 sanitized diagnostics 的证据引用。
- `operator_review_ref`：人工评审、成本 / 运维责任和 rollout / rollback 的记录引用。

## Candidate Evaluation

| candidate class | readiness result | 后续评审要求 |
| --- | --- | --- |
| `managed_postgresql_compatible_service` | `candidate_class_allowed_for_later_review_not_selected` | 必须提供托管边界、backup / restore、TLS / role policy、auditability、region / tenancy 隔离和成本证据引用 |
| `self_managed_postgresql_compatible_cluster` | `candidate_class_allowed_for_later_review_not_selected` | 必须提供运维责任、升级窗口、backup / restore、TLS / role policy、migration handoff 和故障恢复证据引用 |
| `embedded_or_file_database` | `rejected_for_production_audit_store_boundary` | 不满足 production audit store 的 durability、operator review、connection lifecycle 和 multi-tenant isolation 边界 |

本批不选择以上任何候选类别的具体 provider，也不把候选类别写成推荐厂商、默认托管产品或实现入口。

## 评估维度

后续 `storage_adapter_database_provider_selection_review` 必须至少覆盖：

- PostgreSQL-compatible 行为与版本 / 方言差异。
- append-only insert、transaction boundary、duplicate / replay fail-closed 与 partial write recovery。
- backup / restore、rollback / recovery、retention / redaction 和审计留痕。
- secret-ref-only DSN、credential handle、TLS mode、certificate verification 和 least privilege role。
- connection lifecycle、pooling、timeout、retry、sanitized error 和 no raw payload diagnostics。
- migration runner、schema marker、physical schema handoff 和 SQL / DDL artifact guard。
- offline adapter smoke、negative leakage runtime scan、operator review 和 rollout / rollback。
- 成本、可观测性、运维责任、region / tenancy 隔离和生产 / 本地环境差异。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_database_provider_selection_readiness_dependency_missing` | `dependency_chain` | 缺少 concrete database selection review、provider policy、schema artifact、smoke / scan boundary、blocker matrix 或 implementation readiness |
| `audit_store_storage_adapter_database_provider_selection_readiness_input_evidence_missing` | `provider_candidate_input` | 候选输入缺少 metadata-only evidence refs、评估字段或 operator review ref |
| `audit_store_storage_adapter_database_provider_selection_readiness_candidate_source_overreach` | `candidate_source` | 候选来源包含账号资源、endpoint、region detail、vendor raw URL 或运行时探测输出 |
| `audit_store_storage_adapter_database_provider_selection_readiness_provider_or_product_forbidden` | `provider_selection_boundary` | 本批选择云厂商、managed product、provider、hosted product、resource id 或 endpoint |
| `audit_store_storage_adapter_database_provider_selection_readiness_driver_or_dsn_forbidden` | `connection_boundary` | 本批选择 driver、DSN parser、connection provider、connection lifecycle 或 credential material |
| `audit_store_storage_adapter_database_provider_selection_readiness_runtime_forbidden` | `runtime_gate` | 本批创建 SQL、DDL、schema marker、migration runner、storage adapter runtime 或 audit store runtime |
| `audit_store_storage_adapter_database_provider_selection_readiness_secret_material_detected` | `artifact_guard` | 文档或 fixture 出现 raw secret、DSN、endpoint、database detail、credential payload 或 provider detail |
| `audit_store_storage_adapter_database_provider_selection_readiness_scope_overreach` | `implementation_boundary` | 本批打开 repository mode、production API、production resolver 或 runtime task card |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_storage_adapter_database_provider_selection_readiness_status`
- `readiness_decision`
- `runtime_task_decision`
- `next_dependency`
- `selected_database_engine`
- `provider_candidate_source_status`
- `provider_input_evidence_status`
- `provider_evaluation_dimension_status`
- `database_provider_selection_status`
- `provider_selection_review_status`
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

- 不允许把 PostgreSQL-compatible 能力族选择、static product class、reserved profile、metadata contract artifact、logical table schema artifact、offline smoke strategy、negative leakage runtime scan boundary、test-only fake resolver、memory store、mock database provider、historical smoke 或 previous checker success 替代 provider selection readiness。
- 不允许把本 readiness 写成 provider selected、hosted product selected、driver selected、DSN ready、SQL ready、connection ready、storage adapter runtime task card ready、DB provider ready、audit store ready、production resolver ready、repository mode ready、public API ready 或 production ready。
- 不创建 storage adapter runtime、storage adapter runtime task card、storage adapter client、audit store runtime、DB provider、database connection provider、SQL、DDL、schema marker runtime 或 migration runner。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / review fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不选择托管产品、不创建 DB provider、不打开 driver、不解析 DSN、不运行 SQL、不创建 schema marker、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1.py`

不得新增或启用 database provider implementation task card、storage adapter runtime implementation task card、storage adapter runtime、storage adapter client、durable audit backend runtime、database connection provider、DB driver、DSN parser、SQL migration、DDL、physical table schema、schema marker reader / writer、migration runner、audit store runtime implementation task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批之后 storage adapter runtime task card 仍 blocked。若继续 audit store durable backend，应先做 `storage_adapter_database_provider_selection_review`，基于本批证据单独评审 provider / hosted product 候选；不得跳过 provider selection review 直接创建 DB provider、storage adapter runtime、SQL、schema marker、audit store runtime task card 或 production resolver runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
