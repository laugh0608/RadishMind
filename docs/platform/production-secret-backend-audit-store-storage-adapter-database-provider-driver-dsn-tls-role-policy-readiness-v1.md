# Production Secret Backend Audit Store Storage Adapter Database Provider / Driver / DSN / TLS / Role Policy Readiness v1

更新时间：2026-07-02

## 文档目的

本文档在 `Production Secret Backend Audit Store Storage Adapter Runtime Implementation Entry Refresh After Product Selection v1` 之后，定义 future storage adapter runtime task card 之前必须具备的 database provider、driver、DSN secret-ref、TLS mode 和 role / privilege policy 准入证据。

对应切片：`production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1`。

结论：状态为 `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined`，readiness decision 为 `database_provider_driver_dsn_tls_role_policy_defined_without_runtime`。本批只定义静态准入边界：provider boundary、driver selection policy、DSN secret-ref policy、TLS policy、role / privilege policy、sanitized diagnostics、no fallback 和 no side effects。仍不选择 PostgreSQL、MySQL、SQLite、cloud database、vendor service、resource id、endpoint 或具体 driver；也不创建 DB provider、DSN parser、connection factory、SQL migration、schema marker、storage adapter runtime、audit store runtime、repository mode 或 public production API。

下一项依赖固定为 `storage_adapter_append_only_table_schema_boundary_readiness`。该步骤应继续定义 append-only table schema、record identity、sequence / idempotency / retention reference 与 migration marker 的静态边界，不得跳过 schema boundary 直接创建 storage adapter runtime。

## 输入证据

- `audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined` 已确认 static product class 为 `managed_database_append_only_table`，并确认 runtime task card 仍 blocked。
- `audit_store_storage_adapter_metadata_contract_artifact_materialized` 已提供 metadata-only contract artifact、positive / negative fixtures、writer compatibility smoke 和 no secret material scan。
- `audit_store_storage_adapter_backend_product_selection_review_defined` 只选择 static product class 和 reserved profile，不选择具体数据库或 vendor。
- `audit_store_runtime_blocker_matrix_defined` 与 `implementation_readiness_defined` 仍保持 storage adapter runtime、audit store runtime、DB provider、SQL、schema marker、repository mode 和 production API 未创建。

## Readiness Boundary

| gate | 本批结论 | 说明 |
| --- | --- | --- |
| database provider / driver / DSN / TLS / role policy readiness | `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined` | 准入证据已定义 |
| readiness decision | `database_provider_driver_dsn_tls_role_policy_defined_without_runtime` | 只定义静态边界，不创建 runtime |
| next dependency | `storage_adapter_append_only_table_schema_boundary_readiness` | 下一项进入 append-only table schema boundary |
| database product / vendor | `not_selected` | 不绑定具体数据库或云服务 |
| provider boundary | `metadata_only_provider_boundary_defined` | 只定义 future provider 输入、输出、失败和诊断边界 |
| driver allowlist | `static_driver_policy_defined_without_driver_selection` | 只定义选择规则，不选择 driver |
| DSN policy | `secret_ref_only_dsn_policy_defined` | DSN 只能通过 secret-ref 引用，不提交明文 |
| TLS policy | `tls_mode_policy_defined` | 固定 future TLS mode 证据要求，不创建连接 |
| role / privilege policy | `least_privilege_role_policy_defined` | 固定 append-only writer / reader / migration role 边界 |
| connection provider runtime | `not_created` | 不创建 provider、connection factory 或 driver open path |
| storage adapter runtime | `not_created` | 不创建 runtime、client、writer 或 SQL path |
| schema / migration | `required_before_runtime_task_card` | append-only table schema 与 migration marker 仍需独立证据 |
| repository / API | `disabled / not_created` | 不启用 repository mode，不创建 public production API |

## Policy Scope

本批准入证据必须覆盖：

- provider boundary：future provider 只能消费 metadata-only config reference、secret-ref handle、policy version 和 audit reference；不得消费 raw DSN、password、token、provider raw URL 或 database hostname。
- driver selection policy：future driver 必须由独立任务选择，并满足 append-only transaction、parameterized statement、TLS mode、least-privilege role、sanitized error 和 no raw payload diagnostics 要求。
- DSN secret-ref policy：DSN 只能以 secret ref / credential handle 进入 future runtime；committed 文档、fixture、diagnostics 和 checker 不得出现明文 DSN 或 endpoint。
- TLS policy：future runtime 必须显式声明 TLS mode、certificate / verification policy reference、failure semantics 和 diagnostics allowlist；本批不创建 TLS config object。
- role / privilege policy：future database role 必须拆分 append-only writer、read-only verifier、migration / marker role；不得允许 update、delete、truncate、inline redaction 或 schema mutation 成功路径。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_dependency_missing` | `dependency_chain` | 缺少 after-product-selection refresh、product selection、metadata contract artifact、blocker matrix 或 implementation readiness |
| `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_database_selection_forbidden` | `database_product_boundary` | 本批选择具体数据库、vendor、resource id、endpoint 或 driver |
| `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_provider_runtime_forbidden` | `provider_runtime_boundary` | 本批创建 provider、driver open path、DSN parser、connection factory 或 SQL path |
| `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_secret_material_detected` | `artifact_guard` | 文档或 fixture 出现 secret、DSN、endpoint、credential、raw URL 或 database detail |
| `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_schema_scope_overreach` | `schema_boundary` | 本批创建 table schema、migration marker、SQL migration 或 schema runtime |
| `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_runtime_scope_overreach` | `implementation_boundary` | 本批打开 storage adapter runtime、audit store runtime、repository mode、production resolver runtime 或 production API |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_status`
- `readiness_decision`
- `next_dependency`
- `selected_backend_product_class`
- `selected_backend_product_profile`
- `database_product_status`
- `database_vendor_status`
- `provider_boundary_status`
- `driver_selection_policy_status`
- `dsn_secret_ref_policy_status`
- `tls_policy_status`
- `role_policy_status`
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

禁止输出 raw secret、secret value、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、database hostname、database name、table name、database error detail、cloud credential、credential payload、full secret ref value、full credential handle、raw request / response payload、raw audit payload、raw storage payload、payload hash、secret-derived hash 或 provider error detail。

## No Fallback / No Side Effects

- 不允许把 static product class selection、reserved backend product profile、metadata contract artifact、writer compatibility fixture、test-only fake resolver、memory store、mock database provider、historical smoke 或 previous checker success 替代 provider / driver / DSN / TLS / role policy readiness。
- 不允许把本 readiness 写成具体 database ready、driver ready、DB provider ready、connection ready、storage adapter runtime ready、audit store ready、production resolver ready、repository mode ready、public API ready 或 production ready。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / review fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不选择数据库、不创建 DB provider、不打开 driver、不解析 DSN、不运行 SQL、不创建 schema marker、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.py`

不得新增或启用 database product selection artifact、backend vendor selection artifact、database provider implementation task card、storage adapter runtime implementation task card、database connection provider、DB driver、DSN parser、connection factory、SQL migration、schema marker reader / writer、storage adapter runtime、storage adapter client、audit store runtime implementation task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、repository mode runtime 或 public production API。

## 后续推进

本批之后 storage adapter runtime task card 仍 blocked。若继续 audit store durable backend，应推进 `storage_adapter_append_only_table_schema_boundary_readiness`，定义 append-only table schema、record identity、sequence / idempotency / retention reference、schema marker boundary、migration review input 和 sanitized diagnostics；不得跳过 schema boundary 直接创建 storage adapter runtime、DB provider、SQL migration、schema marker runtime、audit store runtime task card 或 production resolver runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
