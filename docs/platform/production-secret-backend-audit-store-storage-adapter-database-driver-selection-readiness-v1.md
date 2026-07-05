# Production Secret Backend Audit Store Storage Adapter Database Driver Selection Readiness v1

更新时间：2026-07-05

## 文档目的

本文档在 `Production Secret Backend Audit Store Storage Adapter Database Provider Selection Review v1` 之后，固定 future storage adapter 的 database driver 选择前准入边界。

对应切片：`production-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1`。

结论：状态为 `audit_store_storage_adapter_database_driver_selection_readiness_defined`，readiness decision 为 `database_driver_selection_readiness_defined_without_driver_selection`。本批只定义 driver candidate source、import boundary、driver capability evidence、DSN secret-ref compatibility、TLS mode compatibility、least privilege role compatibility、connection lifecycle boundary、migration / schema marker boundary、offline adapter smoke boundary、negative leakage runtime scan boundary 和 rollout / rollback boundary；不选择 driver package、DSN parser、connection provider、SQL、DDL、物理表名、列名、列类型、schema marker runtime、migration runner、storage adapter runtime、audit store runtime、repository mode 或 public production API。

下一项依赖固定为 `storage_adapter_database_driver_selection_review`。该依赖必须在本批 readiness 之后单独评审 driver 选择，不能因为已选择 `managed_postgresql_compatible_service` provider candidate class 而直接创建 driver、connection provider、DB provider、SQL 或 runtime task card。

## 输入证据

- `audit_store_storage_adapter_database_provider_selection_review_defined` 已把 provider candidate class 选择为 `managed_postgresql_compatible_service`，但明确不选择 vendor、managed product、具体 provider、driver 或 DSN。
- `audit_store_storage_adapter_database_provider_selection_readiness_defined` 已定义 provider / hosted product 候选输入证据、候选类别、评估维度、停止线和 artifact guard。
- `audit_store_storage_adapter_concrete_database_selection_review_defined` 已把数据库能力族选择为 `postgresql_compatible_append_only_relational_database`，但仍不选择 driver、DSN、connection provider 或 SQL runtime。
- `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined` 已固定 secret-ref-only DSN policy、TLS mode policy 和 least privilege role policy；这些是 driver readiness 的约束输入，不是 driver 选择结论。
- `audit_store_storage_adapter_table_schema_artifact_materialized` 与 `audit_store_storage_adapter_metadata_contract_artifact_materialized` 已提供 metadata-only logical schema / contract artifact；它们不包含 SQL、DDL、物理表名、列名或列类型。
- `audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined` 与 `audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined` 已固定 metadata-only smoke / scan 边界，不创建 runner、scanner 或输出。
- `audit_store_runtime_blocker_matrix_defined` 与 `implementation_readiness_defined` 仍确认 storage adapter runtime、audit store runtime、DB provider、SQL、schema marker、repository mode 和 public production API 未创建。

## Readiness Boundary

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| database driver selection readiness | `audit_store_storage_adapter_database_driver_selection_readiness_defined` | driver 选择前准入边界已定义 |
| readiness decision | `database_driver_selection_readiness_defined_without_driver_selection` | 只定义 driver selection review 所需证据，不选择 driver |
| selected database engine | `postgresql_compatible_append_only_relational_database` | 继承 concrete database selection review 的能力族选择 |
| selected provider candidate class | `managed_postgresql_compatible_service` | 继承 provider selection review 的候选类选择，不绑定 vendor、product 或 endpoint |
| driver candidate source | `metadata_only_driver_candidate_source_defined` | 后续 review 必须从显式候选来源取证，但本批不列出具体 package |
| driver import boundary | `metadata_only_driver_import_boundary_defined` | 后续 review 必须评估 import、license、support 和 runtime ownership |
| driver capability evidence | `metadata_only_driver_capability_evidence_defined` | 后续 review 必须评估 PostgreSQL compatibility、append-only transaction、duplicate / replay fail-closed 和 sanitized error |
| DSN secret-ref compatibility | `metadata_only_dsn_secret_ref_compatibility_defined` | 只允许 secret ref / credential handle 作为 future runtime 输入，不提交 DSN 或 raw credential |
| TLS mode compatibility | `metadata_only_tls_mode_compatibility_defined` | 后续 review 必须评估 TLS mode、certificate verification 和 local/test/prod 差异 |
| least privilege role compatibility | `metadata_only_role_policy_compatibility_defined` | 后续 review 必须证明 role policy 与 append-only writer、schema marker 和 migration handoff 一致 |
| connection lifecycle boundary | `metadata_only_connection_lifecycle_boundary_defined` | 只定义 pooling、timeout、retry、partial write recovery 和 sanitized failure 要求 |
| migration / schema marker boundary | `logical_schema_marker_handoff_boundary_defined` | 仍不创建 schema marker runtime、migration runner、SQL 或 DDL |
| offline adapter smoke boundary | `metadata_only_offline_adapter_smoke_boundary_defined` | 后续 driver review 必须继续复用 offline smoke strategy，不创建 runner |
| negative leakage runtime scan boundary | `metadata_only_negative_leakage_runtime_scan_boundary_defined` | 后续 driver review 必须继续证明无 raw secret / DSN / provider detail 泄漏 |
| runtime task card decision | `storage_adapter_runtime_task_card_still_blocked_after_database_driver_selection_readiness` | runtime task card 仍 blocked |
| next dependency | `storage_adapter_database_driver_selection_review` | 后续独立评审 driver 选择 |

## Driver Review 前置要求

`storage_adapter_database_driver_selection_review` 必须重新审查：

- driver candidate source 是否来自显式清单，且清单只保留 metadata、license、maintenance、support、runtime ownership 和风险摘要。
- driver import boundary 是否不会引入 runtime side effect、隐式环境读取、隐式 DSN 解析或真实 provider 连接。
- driver capability evidence 是否覆盖 PostgreSQL-compatible behavior、append-only insert semantics、transaction boundary、duplicate / replay fail-closed、sanitized error 和 failure taxonomy。
- DSN secret ref、credential handle、TLS mode、certificate verification、least privilege role 与 connection lifecycle 是否一致。
- pooling、timeout、retry、partial write recovery、rollback / recovery、operator review 和 rollout / rollback 是否有可复验证据。
- migration runner、schema marker、physical schema handoff、SQL / DDL artifact guard 是否仍保持独立边界。
- offline adapter smoke、negative leakage runtime scan、local/test/prod environment parity 是否不依赖真实 secret、真实 provider 或真实数据库连接。

本批不写入具体 driver name，不比较 driver package，不新增 Go import，不创建 config key，不创建 connection factory，也不引入 provider SDK。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_database_driver_selection_readiness_dependency_missing` | `dependency_chain` | 缺少 provider selection review、provider selection readiness、concrete database selection review、provider policy、schema artifact、smoke / scan boundary、blocker matrix 或 implementation readiness |
| `audit_store_storage_adapter_database_driver_selection_readiness_input_evidence_missing` | `readiness_boundary` | 缺少 driver candidate source、import boundary、capability evidence、DSN / TLS / role compatibility、connection lifecycle 或 migration handoff 要求 |
| `audit_store_storage_adapter_database_driver_selection_readiness_candidate_source_overreach` | `candidate_source` | 本批写入具体 driver package、driver version、repository URL、vendor selection 或 runtime import |
| `audit_store_storage_adapter_database_driver_selection_readiness_driver_selection_forbidden` | `driver_selection_boundary` | 本批选择 driver、driver package、driver version、DSN parser、connection provider 或 connection lifecycle runtime |
| `audit_store_storage_adapter_database_driver_selection_readiness_runtime_forbidden` | `runtime_gate` | 本批创建 SQL、DDL、schema marker、migration runner、storage adapter runtime、DB provider 或 audit store runtime |
| `audit_store_storage_adapter_database_driver_selection_readiness_secret_material_detected` | `artifact_guard` | 文档或 fixture 出现 raw secret、DSN、endpoint、database detail、credential payload、provider detail 或 connection string |
| `audit_store_storage_adapter_database_driver_selection_readiness_scope_overreach` | `implementation_boundary` | 本批打开 repository mode、production API、production resolver、runtime task card 或真实数据库连接 |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_storage_adapter_database_driver_selection_readiness_status`
- `readiness_decision`
- `runtime_task_decision`
- `next_dependency`
- `selected_database_engine`
- `selected_provider_candidate_class`
- `driver_candidate_source_status`
- `driver_import_boundary_status`
- `driver_capability_evidence_status`
- `driver_selection_status`
- `database_driver_status`
- `database_dsn_status`
- `connection_provider_status`
- `tls_mode_status`
- `least_privilege_role_status`
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

禁止输出 raw secret、secret value、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、connection string、database hostname、database name、table name、column name、column type、database error detail、cloud credential、credential payload、full secret ref value、full credential handle、raw request / response payload、raw audit payload、raw storage payload、payload hash、secret-derived hash 或 provider error detail。

## No Fallback / No Side Effects

- 不允许把 PostgreSQL-compatible 能力族选择、provider candidate class selection、static product class、reserved profile、metadata contract artifact、logical table schema artifact、offline smoke strategy、negative leakage runtime scan boundary、test-only fake resolver、memory store、mock database provider、historical smoke 或 previous checker success 替代 driver selection readiness。
- 不允许把本 readiness 写成 driver selected、DSN ready、SQL ready、connection ready、storage adapter runtime task card ready、DB provider ready、audit store ready、production resolver ready、repository mode ready、public API ready 或 production ready。
- 不创建 storage adapter runtime、storage adapter runtime task card、storage adapter client、audit store runtime、DB provider、database connection provider、driver import、DSN parser、SQL、DDL、schema marker runtime 或 migration runner。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / review fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不选择托管产品、不创建 DB provider、不打开 driver、不解析 DSN、不运行 SQL、不创建 schema marker、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1.py`

不得新增或启用 database provider implementation task card、storage adapter runtime implementation task card、storage adapter runtime、storage adapter client、durable audit backend runtime、database connection provider、DB driver、driver import、DSN parser、SQL migration、DDL、physical table schema、schema marker reader / writer、migration runner、audit store runtime implementation task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批之后 storage adapter runtime task card 仍 blocked。若继续 audit store durable backend，应先做 `storage_adapter_database_driver_selection_review`，基于本批 readiness 证据单独评审 driver 选择；不得跳过 driver review 直接创建 DB provider、storage adapter runtime、SQL、schema marker、audit store runtime task card 或 production resolver runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
