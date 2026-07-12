# Production Secret Backend Audit Store Storage Adapter Database Connection Lifecycle Readiness v1

更新时间：2026-07-05

## 文档目的

本文档在 `Production Secret Backend Audit Store Storage Adapter Database Driver Selection Review v1` 之后，固定 future storage adapter database connection lifecycle 的静态准入边界。

对应切片：`production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1`。

结论：状态为 `audit_store_storage_adapter_database_connection_lifecycle_readiness_defined`，readiness decision 为 `database_connection_lifecycle_readiness_defined_without_connection_runtime`，runtime task card decision 为 `storage_adapter_runtime_task_card_still_blocked_after_database_connection_lifecycle_readiness`。本批只定义 secret-ref-only DSN handoff、TLS / role / environment binding、pool policy、timeout budget、retry / transaction / partial write recovery、duplicate / replay fail-closed、sanitized diagnostics、schema marker / migration handoff、offline verification、negative leakage scan、rollback / rollout 边界；不创建 connection runtime、connection provider、DB provider、connection factory、pool runtime、health check runtime、SQL、DDL、物理表名、列名、列类型、schema marker runtime、migration runner、storage adapter runtime、audit store runtime、production resolver runtime、repository mode 或 public production API。

本批之后 storage adapter runtime task card 仍 blocked。后续如需继续推进，应先做 `storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_readiness` 之类的独立复评；不得把本 readiness 写成 connection ready、DB provider ready、SQL ready、storage adapter runtime ready 或 production ready。

## 输入证据

- `audit_store_storage_adapter_database_driver_selection_review_defined` 已选择 reference-only driver candidate `github.com/jackc/pgx/v5`，但未导入 driver、未 pin version、未定义 DSN parser 或 connection provider。
- `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined` 已固定 secret-ref-only DSN policy、TLS mode policy 和 least privilege role policy；这些仍是静态策略，不是 runtime。
- `audit_store_storage_adapter_table_schema_artifact_materialized` 与 `audit_store_storage_adapter_metadata_contract_artifact_materialized` 已提供 metadata-only logical schema / contract artifact；它们不包含 SQL、DDL、物理表名、列名或列类型。
- `audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined` 与 `audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined` 已固定 metadata-only offline verification / scan 边界。
- `audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined` 已固定 partial write recovery、duplicate / replay recovery 和 rollback / recovery 证据边界。
- `audit_store_runtime_blocker_matrix_defined` 与 `implementation_readiness_defined` 仍确认 storage adapter runtime、audit store runtime、DB provider、SQL、schema marker、repository mode 和 public production API 未创建。

## Readiness Boundary

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| database connection lifecycle readiness | `audit_store_storage_adapter_database_connection_lifecycle_readiness_defined` | connection lifecycle 静态准入已定义 |
| readiness decision | `database_connection_lifecycle_readiness_defined_without_connection_runtime` | 只定义准入，不创建 runtime |
| runtime task card decision | `storage_adapter_runtime_task_card_still_blocked_after_database_connection_lifecycle_readiness` | storage adapter runtime task card 仍 blocked |
| selected driver candidate | `github.com/jackc/pgx/v5` | 继承 driver selection review 的 reference-only candidate |
| DSN handoff | `secret_ref_only_dsn_handoff_defined` | future runtime 只能通过 secret ref / credential handle 接收 DSN material；本批不解析、不构造、不输出 |
| TLS / role / environment binding | `static_tls_role_environment_binding_defined` | 只定义 local / test / prod 环境绑定、least privilege role 和 migration role 分离要求 |
| pool policy | `static_pool_policy_defined_without_pool_runtime` | 只定义 future pool key、隔离维度和关闭责任，不创建 pool |
| timeout budget | `static_timeout_budget_defined_without_runtime_timer` | 只定义 acquire / transaction / write / recovery / diagnostic 阶段预算，不创建 timer |
| retry / transaction / partial write recovery | `static_recovery_policy_defined_without_runtime` | 只定义 fail-closed retry、transaction 和 recovery 前置 |
| duplicate / replay | `duplicate_replay_fail_closed_policy_defined` | duplicate 或 replay 不得成功写入，也不得回退到 memory / fake store |
| sanitized diagnostics | `sanitized_diagnostics_allowlist_defined` | 只允许 metadata-only diagnostics，不输出 secret、DSN、endpoint、host、database detail 或 provider detail |
| schema marker / migration handoff | `schema_marker_migration_handoff_defined` | 只定义 handoff 给 future schema marker / migration runner，不创建 runtime |
| offline verification | `metadata_only_offline_verification_defined` | 只定义离线验证输入、negative cases 和 fixture 边界 |
| negative leakage scan | `metadata_only_negative_leakage_scan_boundary_defined` | 只定义扫描目标和 forbidden material，不创建 scanner 或 scan output |
| rollout / rollback | `metadata_only_rollout_rollback_boundary_defined` | 只定义 rollout / rollback 停止线，不创建 executor |

## Lifecycle Policy

- secret-ref-only DSN handoff：future connection lifecycle 只能消费 secret ref / credential handle 的引用状态；任何 raw secret、raw DSN、endpoint、host、database name、credential payload、provider detail 或 connection string 都必须保持不可见、不可记录、不可输出。
- TLS / role / environment binding：future runtime 必须把 TLS mode、certificate verification、least privilege append-only writer role、migration role、environment binding 和 provider profile binding 作为显式准入输入；不能跨环境复用 pool、role 或 credential reference。
- pool policy：future pool key 至少区分 environment、role class、tenant / workspace scope 和 policy version；write role 与 migration role 不共享 pool；pool close responsibility 归 connection provider runtime，不下放给 storage adapter 或 audit writer。
- timeout budget：future lifecycle 必须区分 credential resolution、connection acquire、transaction begin、append-only write、commit、recovery 和 diagnostics 阶段；任一预算缺失必须 fail closed，不得 fallback 到无限等待或同步阻塞路径。
- retry / transaction / partial write recovery：future retry 必须只覆盖明确可重试的 transient boundary，并且必须绑定 transaction outcome、idempotency key reference 和 partial write recovery policy；未知提交结果必须进入 fail-closed diagnostics 和 recovery handoff。
- duplicate / replay fail-closed：duplicate key、replay attempt、idempotency mismatch 或 recovery ambiguity 不得成功写入，不得变成 overwrite / update / delete / truncate，也不得回退到 memory store、test-only fake resolver、sample 或 fixture。
- sanitized diagnostics：diagnostics 只能包含 status、failure code、lifecycle phase、request id、audit ref、policy version、environment class、role class 和 sanitized dependency state；不得包含 raw payload、secret material、DSN、endpoint、host、database name、table name、column name、column type、credential payload、provider raw URL、provider error detail 或 database error detail。
- schema marker / migration handoff：本批只定义 future schema marker / migration runner 的前置依赖，不创建 schema marker reader / writer、migration runner、SQL、DDL、物理 schema 或 runtime marker。
- offline verification：本批只允许 committed fixture 和 checker 校验静态字段、依赖链、forbidden artifact、no secret material 和注册顺序；不连接数据库、不访问 provider、不下载依赖、不执行 SQL。
- rollout / rollback：future rollout 必须保留人工复核、negative leakage scan、offline verification、rollback boundary 和 no side effects 证据；本批不创建 rollout executor、rollback executor、recovery executor 或 compensating event writer。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_database_connection_lifecycle_readiness_dependency_missing` | `dependency_chain` | 缺少 driver review、database policy、schema artifact、smoke / scan boundary、rollback recovery、blocker matrix 或 implementation readiness |
| `audit_store_storage_adapter_database_connection_lifecycle_readiness_secret_ref_handoff_missing` | `dsn_handoff` | secret-ref-only DSN handoff、credential handle reference 或 no raw material rule 缺失 |
| `audit_store_storage_adapter_database_connection_lifecycle_readiness_tls_role_environment_missing` | `tls_role_environment_binding` | TLS、role、environment 或 provider profile binding 缺失 |
| `audit_store_storage_adapter_database_connection_lifecycle_readiness_pool_policy_missing` | `pool_policy` | pool key、pool sharing ban 或 close responsibility 缺失 |
| `audit_store_storage_adapter_database_connection_lifecycle_readiness_timeout_budget_missing` | `timeout_budget` | acquire / transaction / write / recovery / diagnostics timeout budget 缺失 |
| `audit_store_storage_adapter_database_connection_lifecycle_readiness_recovery_policy_missing` | `retry_transaction_recovery` | retry、transaction outcome、partial write recovery 或 idempotency reference 缺失 |
| `audit_store_storage_adapter_database_connection_lifecycle_readiness_duplicate_replay_policy_missing` | `duplicate_replay` | duplicate / replay fail-closed policy 缺失 |
| `audit_store_storage_adapter_database_connection_lifecycle_readiness_diagnostics_missing` | `sanitized_diagnostics` | diagnostics allowlist / forbidden material list 缺失 |
| `audit_store_storage_adapter_database_connection_lifecycle_readiness_schema_marker_handoff_missing` | `schema_marker_migration_handoff` | schema marker / migration handoff 缺失 |
| `audit_store_storage_adapter_database_connection_lifecycle_readiness_runtime_forbidden` | `runtime_gate` | 创建 connection runtime、pool runtime、health check runtime、SQL、DDL、schema marker runtime、migration runner、storage adapter runtime 或 audit store runtime |
| `audit_store_storage_adapter_database_connection_lifecycle_readiness_secret_material_detected` | `artifact_guard` | 文档或 fixture 出现 raw secret、DSN、endpoint、host、database detail、credential payload 或 provider detail |
| `audit_store_storage_adapter_database_connection_lifecycle_readiness_scope_overreach` | `implementation_boundary` | 打开 repository mode、production API、production resolver runtime、DB provider 或真实数据库连接 |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_storage_adapter_database_connection_lifecycle_readiness_status`
- `readiness_decision`
- `runtime_task_decision`
- `selected_driver_candidate`
- `dsn_handoff_status`
- `tls_role_environment_binding_status`
- `pool_policy_status`
- `timeout_budget_status`
- `retry_transaction_recovery_status`
- `duplicate_replay_status`
- `schema_marker_migration_handoff_status`
- `offline_verification_status`
- `negative_leakage_scan_status`
- `rollout_rollback_status`
- `connection_provider_status`
- `connection_runtime_status`
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

禁止输出 raw secret、secret value、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、connection string、endpoint、host、database name、table name、column name、column type、database error detail、cloud credential、credential payload、full secret ref value、full credential handle、raw request / response payload、raw audit payload、raw storage payload、payload hash、secret-derived hash 或 provider error detail。

## No Fallback / No Side Effects

- 不允许把 selected driver candidate、secret-ref-only DSN policy、test-only fake resolver、memory store、offline smoke strategy、negative leakage scan boundary、metadata contract artifact 或 table schema artifact 替代 connection lifecycle runtime。
- 不允许把本 readiness 写成 DSN ready、connection ready、pool ready、health check ready、SQL ready、schema marker ready、migration ready、storage adapter runtime task card ready、audit store ready、production resolver ready、repository mode ready、public API ready 或 production ready。
- 本批 checker 只读取 committed 文档、fixture、聚合 readiness、Go module file 和 Go source import 文本；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不下载依赖、不创建 DB provider、不打开 driver、不解析 DSN、不运行 SQL、不创建 schema marker、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.py`

不得新增或启用 database connection lifecycle runtime implementation task card、database provider implementation task card、storage adapter runtime implementation task card、storage adapter runtime、storage adapter client、durable audit backend runtime、database connection provider、DB driver import、driver version pin、DSN parser、connection factory、pool runtime、health check runtime、SQL migration、DDL、physical table schema、schema marker reader / writer、migration runner、audit store runtime implementation task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批之后 storage adapter runtime task card 仍 blocked。若继续 audit store durable backend，应先复评 `storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_readiness` 是否具备定义条件；不得跳过 connection lifecycle readiness 直接创建 DB provider、connection provider、storage adapter runtime、SQL、schema marker、audit store runtime task card 或 production resolver runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
