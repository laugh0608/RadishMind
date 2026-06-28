# Production Secret Backend Audit Store Durable Backend Selection Readiness v1

更新时间：2026-06-28

## 文档目的

本文档在 `Production Secret Backend Audit Store Runtime Blocker Matrix v1` 之后，固定 future audit store durable backend selection 的实现准入。

对应切片：`production-secret-backend-audit-store-durable-backend-selection-readiness-v1`。

结论：状态为 `audit_store_durable_backend_selection_readiness_defined`。本批只定义 future durable backend 选择前必须满足的 metadata-only candidate shape、selection matrix、依赖顺序、可解锁条件、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不选择 durable audit backend，不创建 audit store runtime task card，不创建 audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、operator approval runtime、credential handle runtime、backend health runtime、no leakage smoke runtime、production resolver runtime、DB provider、repository mode runtime 或 public production API。

selection decision：`durable_backend_selection_deferred_until_backend_evidence_and_runtime_task_card`。

## 输入证据

- `audit_store_durable_backend_boundary_readiness_defined` 已固定 durable backend owner 与 storage adapter responsibility，但明确不选择 durable backend。
- `audit_store_runtime_event_schema_artifact_implemented` 已提供 metadata-only audit event schema artifact、positive / negative fixtures 与 writer compatibility smoke。
- `audit_store_runtime_implementation_entry_refresh_v4_defined` 与 `audit_store_runtime_blocker_matrix_defined` 已确认 schema artifact 完成后，audit store runtime task card 仍被 durable backend、writer、delivery、idempotency、approval、credential handle、backend health、no leakage smoke 与 production resolver runtime 依赖阻塞。
- `audit_store_writer_runtime_boundary_readiness_defined`、`audit_store_delivery_runtime_readiness_defined` 与 `audit_store_idempotency_runtime_readiness_defined` 仍是静态边界，不代表 runtime 已创建。
- `credential_handle_runtime_implementation_entry_refresh_defined`、`operator_approval_runtime_implementation_entry_refresh_defined`、`resolver_backend_health_runtime_implementation_entry_refresh_defined` 与 `real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_refresh_defined` 均确认相关 runtime task card 仍 blocked。
- `production_resolver_runtime_implementation_entry_refresh_v2_defined` 仍保持 production resolver runtime task card blocked。
- `implementation_readiness_defined` 仍保持 `production_secret_backend_status=not_satisfied`、`durable_audit_backend_status=not_selected`、`audit_store_runtime_task_card_status=not_created`、`audit_store_runtime_status=not_created`、`audit_writer_status=not_created` 与 `audit_event_delivery_status=not_executed`。

## Selection Boundary

durable backend selection readiness 只允许定义选择准入，不允许执行选择或打开存储运行时：

| gate | 状态 | 说明 |
| --- | --- | --- |
| candidate backend kind | `reserved_only` | 只允许 future reserved kind，不绑定具体 DB、object store、queue、log sink 或 vendor service |
| durable backend selection | `deferred_without_backend_selection` | 本批不选择或启用 durable audit backend |
| storage adapter contract | `metadata_only_static_reference` | 只引用 storage adapter contract，不创建 adapter runtime |
| event schema artifact | `implemented_static_schema_artifact` | 只消费 committed schema artifact，不创建 runtime writer |
| retention / redaction policy | `required_reference_only` | 只允许 policy ref，不持久化 secret 或 payload |
| audit writer runtime | `not_created` | writer runtime 必须后续独立评审 |
| idempotency runtime | `not_created` | key store、duplicate detector 和 replay decision 不创建 |
| delivery runtime | `not_created` | delivery、retry 和 result persistence 不执行 |
| operator approval runtime | `not_created` | 不执行 approval、ticket 或 change window gate |
| credential handle runtime | `not_created` | 不生成 credential handle 或 credential payload |
| backend health runtime | `not_created` | 不执行 backend health check 或 provider call |
| no leakage smoke runtime | `not_created` | 不声明 runtime smoke 已执行 |
| audit store runtime task card | `not_created` | 本批不创建 audit store runtime task card |
| production resolver runtime | `not_created` | 不创建 resolver runtime task card 或 runtime |
| database / repository / API | `blocked` | DB provider、repository mode runtime 和 public production API 不打开 |

## Candidate Shape

未来 durable backend selection task card 必须只消费 metadata-only manifest：

- `candidate_backend_kind`
- `selection_policy_version`
- `environment`
- `storage_owner_ref`
- `storage_adapter_contract_ref`
- `event_schema_ref`
- `retention_policy_ref`
- `redaction_profile_ref`
- `writer_runtime_dependency`
- `delivery_runtime_dependency`
- `idempotency_runtime_dependency`
- `operator_approval_dependency`
- `credential_handle_dependency`
- `backend_health_dependency`
- `no_leakage_smoke_dependency`
- `offline_validation_manifest_ref`

允许的 candidate kind 仅保留为 reserved kind：`reserved_append_only_audit_log`、`reserved_managed_audit_log`、`reserved_operator_managed_audit_store`。这些名字只表达后续候选类别，不代表当前选择了数据库、对象存储、托管日志服务或 operator-managed store。

禁止 manifest、fixture、docs 或 diagnostics 中出现 raw secret、secret value、password、token、API key、provider raw URL、resolver backend URL、DSN、database hostname、database error detail、cloud credential、credential payload、完整 secret ref value、完整 credential handle、raw request payload、raw response payload、raw audit payload、secret-derived hash、authorization header、cookie 或 provider error detail。

## Dependency Order

durable backend selection readiness 将 blocker matrix 中的顺序位固化为后续开发节奏：

1. `runtime_event_schema_artifact_implemented`
2. `durable_backend_selection_readiness`
3. `audit_writer_runtime_entry_review`
4. `idempotency_runtime_entry_review`
5. `delivery_runtime_entry_review`
6. `operator_approval_runtime_entry_refresh`
7. `credential_handle_runtime_entry_refresh`
8. `backend_health_runtime_entry_refresh`
9. `no_leakage_smoke_runtime_entry_refresh`
10. `audit_store_runtime_entry_refresh_after_blocker_matrix`
11. `production_resolver_runtime_entry_refresh_after_audit_store_runtime`

本批只完成第 2 项的准入证据。第 3 项及以后仍必须分别有独立专题、fixture、checker 和 task card 评审。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_durable_backend_selection_dependency_missing` | `dependency_chain` | 必需 readiness / refresh / schema artifact 证据缺失 |
| `audit_store_durable_backend_selection_backend_selected_forbidden` | `selection_gate` | 本批选择或启用 durable audit backend |
| `audit_store_durable_backend_selection_adapter_runtime_forbidden` | `storage_adapter` | 本批创建 storage adapter runtime 或 provider binding |
| `audit_store_durable_backend_selection_database_provider_forbidden` | `artifact_guard` | 本批创建 DB provider、driver、DSN parser、migration 或 schema marker |
| `audit_store_durable_backend_selection_writer_runtime_blocked` | `audit_writer` | writer runtime 仍未独立解锁 |
| `audit_store_durable_backend_selection_delivery_runtime_blocked` | `delivery` | delivery runtime 仍未独立解锁 |
| `audit_store_durable_backend_selection_idempotency_runtime_blocked` | `idempotency` | idempotency runtime 仍未独立解锁 |
| `audit_store_durable_backend_selection_operator_approval_blocked` | `operator_approval` | operator approval runtime 仍 blocked |
| `audit_store_durable_backend_selection_credential_handle_blocked` | `credential_handle` | credential handle runtime 仍 blocked |
| `audit_store_durable_backend_selection_backend_health_blocked` | `backend_health` | backend health runtime 仍 blocked |
| `audit_store_durable_backend_selection_no_leakage_blocked` | `no_secret_leakage` | no leakage smoke runtime 仍 blocked |
| `audit_store_durable_backend_selection_secret_material_detected` | `artifact_guard` | 静态 artifact 出现 secret-bearing material |
| `audit_store_durable_backend_selection_repository_mode_forbidden` | `workflow_durable_store` | 本批启用 repository mode runtime |
| `audit_store_durable_backend_selection_scope_overreach` | `implementation_boundary` | 本批合入 runtime、DB、API、executor 或 production resolver 能力 |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_durable_backend_selection_status`
- `selection_decision`
- `candidate_backend_kind`
- `candidate_status`
- `gate`
- `gate_status`
- `environment`
- `storage_owner_ref`
- `storage_adapter_contract_status`
- `event_schema_ref`
- `retention_policy_ref`
- `redaction_profile_ref`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 forbidden material 中列出的任一原文、完整 payload、provider error detail 或 secret-derived material。

## No Fallback

- 不允许把 memory store、fake resolver、test profile、local smoke profile、developer env、static fixture、sample、mock provider、repository memory store、static handoff envelope、static schema artifact 或 static boundary 文档提升为 durable audit backend。
- 不允许缺少 durable backend 时 fallback 到 audit writer runtime、delivery runtime、idempotency runtime、operator approval runtime、credential handle runtime、backend health runtime 或 no leakage smoke runtime。
- 不允许在 writer、delivery、idempotency、operator approval、credential handle、backend health 或 no leakage smoke runtime 仍 blocked 时创建 audit store runtime task card。
- 不允许把 `audit_store_durable_backend_selection_readiness_defined` 写成 durable backend ready、audit store ready、production resolver ready、repository mode ready、database ready、production API ready 或 production ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema、已有 readiness / refresh fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 production resolver、不创建 storage adapter runtime、不创建 audit store、不创建 audit writer、不写 audit event、不执行 delivery、不执行 idempotency decision、不执行 duplicate detection、不执行 approval runtime、不创建 credential payload、不创建 credential handle、不执行 backend health check、不执行 smoke runtime、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不调用 production API。

side effect counters 必须保持：

- `real_secret_read_count=0`
- `environment_secret_read_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `production_resolver_call_count=0`
- `durable_audit_backend_selected_count=0`
- `storage_adapter_runtime_created_count=0`
- `audit_store_runtime_created_count=0`
- `audit_writer_runtime_created_count=0`
- `audit_event_write_count=0`
- `delivery_execution_count=0`
- `idempotency_decision_count=0`
- `duplicate_detection_count=0`
- `operator_approval_runtime_execution_count=0`
- `credential_handle_runtime_created_count=0`
- `credential_payload_created_count=0`
- `backend_health_runtime_created_count=0`
- `backend_health_check_count=0`
- `no_secret_leakage_smoke_runtime_created_count=0`
- `database_connection_count=0`
- `driver_open_count=0`
- `sql_execution_count=0`
- `schema_marker_read_count=0`
- `schema_marker_write_count=0`
- `repository_mode_enablement_count=0`
- `production_api_call_count=0`

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-durable-backend-selection-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-durable-backend-selection-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-durable-backend-selection-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-durable-backend-selection-readiness-v1.py`

不得新增或启用 durable audit backend、storage adapter runtime、audit store runtime implementation task card、audit store runtime、audit writer runtime、audit event writer / runner、audit delivery runtime、audit idempotency runtime、duplicate detector runtime、retry executor、replay executor、production resolver runtime、cloud secret SDK / client、secret value fixture、production credential file、credential payload runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、database connection provider、DB driver、DSN parser、SQL migration、schema marker reader / writer、migration runner、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批完成后，durable backend 仍保持 `not_selected`，audit store runtime task card 仍不能创建。后续应按 blocker matrix 顺序先进入 writer runtime entry review、idempotency runtime entry review 与 delivery runtime entry review 的独立准入；只有这些 runtime 依赖和 durable backend concrete selection 各自解除阻塞后，才允许重新评审 audit store runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-durable-backend-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不选择 durable audit backend，不创建 storage adapter runtime，不绑定 DB / object store / queue / log sink / vendor service。
- 不创建 audit store runtime task card、audit store runtime、audit writer runtime、delivery runtime、idempotency runtime、operator approval runtime、credential handle runtime、backend health runtime、no leakage smoke runtime、production resolver runtime task card 或 production resolver runtime。
- 不连接数据库，不打开 driver，不运行 SQL，不读写 schema marker，不启用 repository mode runtime，不创建 public production API。
- 不读取真实 secret，不读取环境 secret，不访问 provider，不执行 approval / health / smoke runtime，不写 audit event，不执行 delivery / duplicate detection / replay。
- 不把 `audit_store_durable_backend_selection_readiness_defined` 写成 durable persistence ready、audit store ready、production resolver ready、secret backend ready、repository mode ready、database ready、production API ready 或 production ready。
