# Production Secret Backend Audit Store Storage Adapter Database Driver Selection Review v1

更新时间：2026-07-05

## 文档目的

本文档在 `Production Secret Backend Audit Store Storage Adapter Database Driver Selection Readiness v1` 之后，固定 future storage adapter 的 database driver 选择评审结论。

对应切片：`production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1`。

结论：状态为 `audit_store_storage_adapter_database_driver_selection_review_defined`，selection decision 为 `database_driver_candidate_selected_pgx_v5_runtime_blocked`。本批只把 future Go PostgreSQL driver candidate 选择为 `github.com/jackc/pgx/v5`，并把该候选的来源、import 边界、能力证据、DSN / TLS / role 兼容、connection lifecycle、offline smoke / negative leakage scan 和 rollout / rollback 边界固定为静态评审事实；不新增 Go import、不固定 dependency version、不创建 DSN parser、connection provider、DB provider、SQL、DDL、物理表名、列名、列类型、schema marker runtime、migration runner、storage adapter runtime、audit store runtime、repository mode 或 public production API。

下一项依赖固定为 `storage_adapter_database_connection_lifecycle_readiness`。该依赖必须在 driver candidate 选择之后单独定义 secret-ref-only DSN handoff、pooling、timeout、retry、partial write recovery、sanitized failure、schema marker handoff 和 offline verification 的 connection lifecycle 边界，不能因为本批选择 `github.com/jackc/pgx/v5` 而直接创建 connection provider、DB provider、SQL 或 runtime task card。

## 输入证据

- `audit_store_storage_adapter_database_driver_selection_readiness_defined` 已固定 driver candidate source、import boundary、capability evidence、DSN / TLS / role compatibility、connection lifecycle、migration / schema marker、offline smoke、negative leakage runtime scan 和 rollout / rollback 选择前证据。
- `audit_store_storage_adapter_database_provider_selection_review_defined` 已把 provider candidate class 选择为 `managed_postgresql_compatible_service`，但不选择 vendor、managed product、具体 provider、driver 或 DSN。
- `audit_store_storage_adapter_concrete_database_selection_review_defined` 已把数据库能力族选择为 `postgresql_compatible_append_only_relational_database`。
- `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined` 已固定 secret-ref-only DSN policy、TLS mode policy 和 least privilege role policy；这些是 driver review 的约束输入，不是 connection runtime。
- `audit_store_storage_adapter_table_schema_artifact_materialized` 与 `audit_store_storage_adapter_metadata_contract_artifact_materialized` 已提供 metadata-only logical schema / contract artifact；它们不包含 SQL、DDL、物理表名、列名或列类型。
- `audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined` 与 `audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined` 已固定 metadata-only smoke / scan 边界，不创建 runner、scanner 或输出。
- `audit_store_runtime_blocker_matrix_defined` 与 `implementation_readiness_defined` 仍确认 storage adapter runtime、audit store runtime、DB provider、SQL、schema marker、repository mode 和 public production API 未创建。

## 外部候选证据

本批只记录可复查的公开来源，不下载依赖、不读取网络内容到运行时、不引入 module import。

| source | 结论用途 |
| --- | --- |
| `https://github.com/jackc/pgx` | 官方仓库把 pgx 定位为 PostgreSQL driver and toolkit，仓库声明 MIT license，并说明 pgx 是 pure Go driver / toolkit，包含 PostgreSQL-specific features、`database/sql` adapter、TLS control、connection pool 和 stable `v5` policy |
| `https://pkg.go.dev/github.com/jackc/pgx/v5` | Go package 页面显示 module path `github.com/jackc/pgx/v5`、MIT license、stable tagged module、公开文档和 package metadata，可作为 future dependency review 的候选来源 |

本批不把外部页面内容复制为长期事实源；后续若真正引入依赖，必须在 dependency import / license review 任务中重新确认版本、license、release notes、安全公告和 transitive dependency 风险。

## Selection Review

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| database driver selection review | `audit_store_storage_adapter_database_driver_selection_review_defined` | driver 选择评审已完成 |
| selection decision | `database_driver_candidate_selected_pgx_v5_runtime_blocked` | 只选择 future driver candidate，不导入依赖 |
| selected driver candidate | `github.com/jackc/pgx/v5` | 作为 reference-only candidate；不固定具体 dependency version |
| selected database engine | `postgresql_compatible_append_only_relational_database` | 继承 concrete database selection review 的能力族选择 |
| selected provider candidate class | `managed_postgresql_compatible_service` | 继承 provider selection review 的候选类选择 |
| driver selection status | `selected_driver_candidate_without_runtime_import` | 已选择 driver candidate，但 import、version pin 和 runtime ownership 仍未打开 |
| driver package status | `selected_candidate_reference_only` | 不新增 `go.mod` / import / dependency lock |
| driver import status | `not_created` | 不新增 Go import，不调用 driver |
| DSN / connection provider | `not_defined / not_created` | 不创建 DSN parser、connection provider、connection lifecycle runtime |
| SQL / migration / schema marker | `not_created` | 不创建 SQL、DDL、schema marker runtime 或 migration runner |
| runtime task card decision | `storage_adapter_runtime_task_card_still_blocked_after_database_driver_selection_review` | storage adapter runtime task card 仍 blocked |
| next dependency | `storage_adapter_database_connection_lifecycle_readiness` | 后续独立定义 connection lifecycle 准入 |

## Candidate Evaluation

| candidate | review result | 依据 |
| --- | --- | --- |
| `github.com/jackc/pgx/v5` | `selected_driver_candidate_without_runtime_import` | 与 Go 服务层、PostgreSQL-compatible 能力族、managed PostgreSQL-compatible provider class、append-only transaction boundary、TLS control、connection pool、`database/sql` adapter 兼容、MIT license 和 stable `v5` policy 最一致 |
| `database/sql_only_abstraction_without_selected_driver` | `deferred_until_connection_lifecycle_readiness` | 只能作为调用边界候选，不能替代 driver candidate；后续可评审是否通过 pgx stdlib adapter 暴露 `database/sql` |
| `mock_or_memory_database_driver` | `rejected_for_production_audit_store_boundary` | 只能用于测试替身或 offline fixture，不满足 production audit store durability、connection lifecycle、TLS / role policy 和 operator review 边界 |

本批选择的是 driver candidate，不等于选择 provider、database account、endpoint、DSN format runtime、connection provider、SQL dialect runtime、schema migration 方案或物理 schema 名称。

## 后续 Connection Lifecycle Readiness 要求

`storage_adapter_database_connection_lifecycle_readiness` 必须重新审查：

- `github.com/jackc/pgx/v5` candidate 如何通过 secret ref / credential handle 接收 DSN 输入，不暴露 raw DSN、host、database name 或 credential payload。
- TLS mode、certificate verification、least privilege role、migration role、append-only writer role 和 environment binding 如何在 local / test / prod 中区分。
- Pooling、timeout、retry、transaction boundary、partial write recovery、duplicate / replay fail-closed 和 sanitized error mapping 如何落到 future connection lifecycle。
- Offline adapter smoke、negative leakage runtime scan、rollback / recovery、operator review 和 rollout / rollback 如何继续保持 metadata-only evidence。
- SQL / DDL、physical schema、schema marker runtime、migration runner、connection provider、DB provider 和 storage adapter runtime task card 是否仍保持独立边界。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_database_driver_selection_review_dependency_missing` | `dependency_chain` | 缺少 driver readiness、provider review、concrete database review、provider policy、schema artifact、smoke / scan boundary、blocker matrix 或 implementation readiness |
| `audit_store_storage_adapter_database_driver_selection_review_candidate_evidence_missing` | `selection_review` | 缺少候选来源、license、maintenance、support、feature、TLS、pooling、stable version 或 public package metadata |
| `audit_store_storage_adapter_database_driver_selection_review_import_forbidden` | `driver_import_boundary` | 本批新增 Go import、改 `go.mod`、固定 dependency version 或执行 driver code |
| `audit_store_storage_adapter_database_driver_selection_review_connection_forbidden` | `connection_boundary` | 本批创建 DSN parser、connection provider、connection factory、pooling runtime、database connection 或 provider account resource |
| `audit_store_storage_adapter_database_driver_selection_review_runtime_forbidden` | `runtime_gate` | 本批创建 SQL、DDL、schema marker、migration runner、storage adapter runtime、DB provider 或 audit store runtime |
| `audit_store_storage_adapter_database_driver_selection_review_secret_material_detected` | `artifact_guard` | 文档或 fixture 出现 raw secret、DSN、endpoint、database detail、credential payload、provider detail 或 connection string |
| `audit_store_storage_adapter_database_driver_selection_review_scope_overreach` | `implementation_boundary` | 本批打开 repository mode、production API、production resolver、runtime task card 或真实数据库连接 |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_storage_adapter_database_driver_selection_review_status`
- `selection_decision`
- `runtime_task_decision`
- `next_dependency`
- `selected_database_engine`
- `selected_provider_candidate_class`
- `selected_driver_candidate`
- `driver_selection_status`
- `database_driver_status`
- `database_driver_package_status`
- `database_driver_import_status`
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

- 不允许把 PostgreSQL-compatible 能力族选择、managed provider candidate class、static product class、reserved profile、metadata contract artifact、logical table schema artifact、offline smoke strategy、negative leakage runtime scan boundary、test-only fake resolver、memory store、mock database provider、historical smoke 或 previous checker success 替代 driver selection review。
- 不允许把本 review 写成 driver imported、dependency version pinned、DSN ready、SQL ready、connection ready、storage adapter runtime task card ready、DB provider ready、audit store ready、production resolver ready、repository mode ready、public API ready 或 production ready。
- 不创建 storage adapter runtime、storage adapter runtime task card、storage adapter client、audit store runtime、DB provider、database connection provider、driver import、DSN parser、SQL、DDL、schema marker runtime 或 migration runner。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / review fixture、外部来源 URL 字面量和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不下载依赖、不创建 DB provider、不打开 driver、不解析 DSN、不运行 SQL、不创建 schema marker、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.py`

不得新增或启用 database connection lifecycle implementation task card、database provider implementation task card、storage adapter runtime implementation task card、storage adapter runtime、storage adapter client、durable audit backend runtime、database connection provider、DB driver import、DSN parser、SQL migration、DDL、physical table schema、schema marker reader / writer、migration runner、audit store runtime implementation task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批之后 storage adapter runtime task card 仍 blocked。若继续 audit store durable backend，应先做 `storage_adapter_database_connection_lifecycle_readiness`，基于本批 driver candidate 选择结论单独定义 secret-ref-only DSN handoff、connection lifecycle、pooling、timeout、retry、partial write recovery、sanitized failure、migration / schema marker handoff 和 verification 边界；不得跳过 connection lifecycle readiness 直接创建 DB provider、storage adapter runtime、SQL、schema marker、audit store runtime task card 或 production resolver runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
