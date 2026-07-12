# Production Secret Backend Audit Store Storage Adapter Concrete Database Selection Readiness v1

更新时间：2026-07-04

## 文档目的

本文档在 `Production Secret Backend Audit Store Storage Adapter Runtime Implementation Entry Refresh After Negative Leakage Runtime Scan Boundary v1` 之后，定义 future storage adapter 具体数据库选择前的 readiness 边界。

对应切片：`production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1`。

结论：状态为 `audit_store_storage_adapter_concrete_database_selection_readiness_defined`，readiness decision 为 `concrete_database_selection_readiness_defined_without_database_selection`，runtime task card decision 为 `storage_adapter_runtime_task_card_still_blocked_after_concrete_database_selection_readiness`。本批只定义选择前必须提交的输入证据、metadata-only 候选字段、候选评估维度、拒绝条件、sanitized diagnostics、no fallback / no side effects 和 artifact guard；不选择具体数据库产品 / vendor，不创建 DB provider、driver、DSN parser、connection provider、SQL、DDL、物理表名、列名类型、schema marker runtime、migration runner、storage adapter runtime、audit store runtime、repository mode 或 public production API。

下一项依赖固定为 `storage_adapter_concrete_database_selection_review`。该 review 仍必须作为后续独立证据推进，且只有在本 readiness 的输入证据和停止线全部满足后，才可评审是否选择具体数据库。

## 输入证据

- `audit_store_storage_adapter_runtime_implementation_entry_refresh_after_negative_leakage_runtime_scan_boundary_defined` 已确认 negative leakage runtime scan boundary 后 storage adapter runtime task card 仍 blocked，上一项 next dependency 为 `storage_adapter_concrete_database_selection_readiness`。
- `audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined` 已定义 provider boundary、static driver policy、secret-ref-only DSN policy、TLS mode policy 和 least privilege role policy，但没有选择 driver 或创建 provider。
- `audit_store_storage_adapter_table_schema_artifact_materialized` 已物化 metadata-only logical table schema artifact；该 artifact 不包含 SQL、DDL、物理表名、列名或列类型。
- `audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined` 与 `audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined` 只定义 metadata-only smoke / scan 边界，不创建 runner、scanner 或输出。
- `audit_store_runtime_blocker_matrix_defined` 与 `implementation_readiness_defined` 仍确认 audit store runtime、storage adapter runtime、writer / delivery / idempotency runtime、repository mode 和 public production API 未创建。

## Readiness Boundary

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| concrete database selection readiness | `audit_store_storage_adapter_concrete_database_selection_readiness_defined` | 只定义选择前 readiness |
| readiness decision | `concrete_database_selection_readiness_defined_without_database_selection` | 不选择数据库产品或 vendor |
| runtime task card decision | `storage_adapter_runtime_task_card_still_blocked_after_concrete_database_selection_readiness` | runtime task card 仍 blocked |
| next dependency | `storage_adapter_concrete_database_selection_review` | 后续独立评审是否进入具体选择 |
| selected backend product class | `managed_database_append_only_table` | 只沿用 static product class，不等于具体数据库 |
| candidate source | `metadata_only_candidate_source_defined` | 只允许提交 metadata-only 候选引用 |
| candidate evaluation matrix | `metadata_only_evaluation_matrix_defined` | 只定义评估维度，不填入具体 vendor 结果 |
| database product / vendor | `not_selected` | 不绑定任何具体数据库或供应商 |
| DB provider / driver / DSN | `not_created / not_selected / not_defined` | 不创建 provider、driver、DSN parser 或 connection factory |
| SQL / schema marker / migration | `not_created` | 不创建 SQL、DDL、schema marker runtime 或 migration runner |
| storage adapter runtime | `not_created` | 不创建 runtime、client、writer 或 connection |
| audit store runtime | `not_created` | 不创建 audit store runtime task card |
| repository / API | `disabled / not_created` | 不启用 repository mode，不创建 public production API |

## Metadata-only Candidate Fields

后续 selection review 如需提交候选，只能使用 metadata-only 字段：

- `candidate_key`
- `candidate_profile_ref`
- `candidate_product_class`
- `deployment_model_ref`
- `append_only_capability_ref`
- `transaction_boundary_ref`
- `ordering_guarantee_ref`
- `idempotency_policy_ref`
- `retention_redaction_policy_ref`
- `tls_policy_ref`
- `role_policy_ref`
- `secret_ref_dsn_policy_ref`
- `schema_marker_handoff_ref`
- `migration_review_ref`
- `offline_smoke_strategy_ref`
- `negative_leakage_scan_boundary_ref`
- `rollback_recovery_ref`
- `operator_review_ref`
- `decision_rationale_ref`

这些字段只能承载短 key 或 reference，不得承载 endpoint、resource id、driver import path、DSN、数据库名、物理表名、列名、列类型、raw error、secret material、payload 或 provider detail。

## Candidate Evaluation Dimensions

后续候选评估必须至少覆盖：

1. Append-only insert semantics：能否表达只追加、不覆盖、不删除、不截断。
2. Metadata contract compatibility：能否承载已物化 metadata contract 的 input / result envelope、record identity 和 failure taxonomy。
3. Logical schema compatibility：能否承载 logical field group、record identity、sequence reference、idempotency reference、retention / redaction reference 与 schema marker handoff。
4. Idempotency and duplicate handling：能否支持 fail-closed duplicate / replay 语义。
5. Ordering and durability：能否提供可审计 ordering reference 与 durable write evidence。
6. Retention / redaction compatibility：能否保持 append-only 与 retention / redaction policy 不冲突。
7. Secret-ref-only connection policy：能否只通过 secret ref 传递 connection credential，不暴露 DSN 或 raw secret。
8. TLS and least privilege role policy：能否承接 TLS mode 与最小权限 role policy。
9. Migration and schema marker handoff：能否在后续独立任务中交给 schema marker / migration runner 评审。
10. Offline validation and negative leakage coverage：能否消费 offline smoke strategy 与 negative leakage runtime scan boundary。
11. Rollback / recovery semantics：能否表达 append-only recovery、partial write recovery 和 duplicate / replay recovery。
12. Sanitized diagnostics：能否只输出 allowlisted metadata，不泄露 database detail、provider detail 或 payload。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_concrete_database_selection_readiness_dependency_missing` | `dependency_chain` | 缺少 after-negative-leakage entry refresh、database provider policy、table schema artifact、runtime blocker matrix 或 implementation readiness |
| `audit_store_storage_adapter_concrete_database_selection_readiness_missing_evidence` | `database_selection_readiness` | selection review 缺少输入证据、candidate fields 或 evaluation dimensions |
| `audit_store_storage_adapter_concrete_database_selection_readiness_database_selection_forbidden` | `database_product_boundary` | 本批选择具体数据库产品、vendor、endpoint、resource id、driver 或 physical schema |
| `audit_store_storage_adapter_concrete_database_selection_readiness_runtime_forbidden` | `runtime_gate` | 本批创建 DB provider、connection provider、SQL、DDL、schema marker、migration runner、storage adapter runtime 或 audit store runtime |
| `audit_store_storage_adapter_concrete_database_selection_readiness_secret_material_detected` | `artifact_guard` | 文档或 fixture 出现 secret、DSN、endpoint、database detail、raw payload 或 provider detail |
| `audit_store_storage_adapter_concrete_database_selection_readiness_scope_overreach` | `implementation_boundary` | 本批打开 repository mode、production API、production resolver runtime 或 storage adapter runtime task card |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_storage_adapter_concrete_database_selection_readiness_status`
- `readiness_decision`
- `runtime_task_decision`
- `next_dependency`
- `selected_backend_product_class`
- `selected_backend_product_profile`
- `candidate_source_status`
- `candidate_evaluation_matrix_status`
- `database_selection_review_status`
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

- 不允许把 static product class、reserved profile、metadata contract artifact、logical table schema artifact、offline smoke strategy、negative leakage runtime scan boundary、test-only fake resolver、memory store、mock database provider、historical smoke 或 previous checker success 替代 concrete database selection readiness。
- 不允许把本 readiness 写成 concrete database selected、storage adapter runtime task card ready、DB provider ready、audit store ready、production resolver ready、repository mode ready、public API ready 或 production ready。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / review fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不选择数据库、不创建 storage adapter runtime、不连接数据库、不打开 driver、不运行 SQL、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.py`

不得新增或启用 concrete database selection review artifact、database product selection artifact、backend vendor selection artifact、database provider implementation task card、storage adapter runtime implementation task card、storage adapter runtime、storage adapter client、durable audit backend runtime、database connection provider、DB driver、DSN parser、SQL migration、DDL、schema marker reader / writer、migration runner、audit store runtime implementation task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、production resolver runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批之后 storage adapter runtime task card 仍 blocked。若继续 audit store durable backend，应先做 `storage_adapter_concrete_database_selection_review`，且该 review 必须消费本 readiness 的输入证据、candidate fields、evaluation dimensions、negative gates 和 artifact guard；不得跳过该 review 直接创建 DB provider、storage adapter runtime、SQL、schema marker、audit store runtime task card 或 production resolver runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-negative-leakage-runtime-scan-boundary-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
