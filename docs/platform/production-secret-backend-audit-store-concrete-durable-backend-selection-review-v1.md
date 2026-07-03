# Production Secret Backend Audit Store Concrete Durable Backend Selection Review v1

更新时间：2026-06-29

## 文档目的

本文档在 `Production Secret Backend Audit Store Durable Backend Selection Readiness v1`、writer / idempotency / delivery runtime implementation entry review 与 blocker matrix 之后，固定 future audit store durable backend 的具体 backend family 选择评审。

对应切片：`production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1`。

结论：状态为 `audit_store_concrete_durable_backend_selection_review_defined`。本批只把 durable backend family 静态选择为 `append_only_metadata_audit_log`，对应保留候选 `reserved_append_only_audit_log`；该选择只代表后续存储适配器和 runtime task card 的目标形态，不选择 vendor / database product，不创建 storage adapter runtime、DB provider，不创建 audit store runtime task card，不创建 audit store runtime、writer runtime、delivery runtime、idempotency runtime、production resolver runtime、repository mode 或 public production API。

selection decision：`durable_backend_family_selected_static_append_only_audit_log_runtime_blocked`。

## 输入证据

- `audit_store_durable_backend_boundary_readiness_defined` 已固定 durable backend owner 与 metadata-only storage adapter responsibility。
- `audit_store_durable_backend_selection_readiness_defined` 已定义 reserved-only candidate shape、selection matrix、依赖顺序、failure mapping 与停止线。
- `audit_store_runtime_event_schema_artifact_implemented` 已固定 metadata-only audit event schema artifact、positive / negative fixtures 与 writer compatibility smoke。
- `audit_store_runtime_blocker_matrix_defined` 已确认 audit store runtime task card 仍被 durable backend、writer、idempotency、delivery、approval、credential handle、backend health、no leakage smoke 与 production resolver runtime 阻塞。
- `audit_store_writer_runtime_implementation_entry_review_defined`、`audit_store_idempotency_runtime_implementation_entry_review_defined` 与 `audit_store_delivery_runtime_implementation_entry_review_defined` 已确认 writer / idempotency / delivery runtime task card 均仍 blocked。
- `implementation_readiness_defined` 当前仍保持 `production_secret_backend_status=not_satisfied`、`audit_store_runtime_task_card_status=not_created`、`audit_store_runtime_status=not_created`、`audit_writer_status=not_created` 和 `audit_event_delivery_status=not_executed`。

## Selection Review

| gate | 结论 | 说明 |
| --- | --- | --- |
| selected backend family | `append_only_metadata_audit_log` | 后续 durable audit store 目标形态为 append-only metadata audit log |
| selected reserved candidate | `reserved_append_only_audit_log` | 从 reserved-only candidate 中选择 append-only 方向 |
| selection scope | `static_family_only` | 不绑定 PostgreSQL、MySQL、SQLite、object store、queue、log sink 或 vendor service |
| storage adapter runtime | `not_created` | 后续仍需独立 task card 和 runtime 实现评审 |
| database provider | `blocked` | 不创建 DB provider、driver、DSN parser、migration 或 schema marker |
| audit store runtime task card | `not_created` | backend family 选择不解锁 runtime task card |
| writer runtime | `not_created` | writer runtime task card 仍 blocked |
| idempotency runtime | `not_created` | idempotency runtime task card 仍 blocked |
| delivery runtime | `not_created` | delivery runtime task card 仍 blocked |
| production resolver runtime | `not_created` | 不创建 production resolver runtime task card 或 runtime |
| repository / API | `blocked` | repository mode runtime 与 public production API 不打开 |

## Why Append-only Metadata Audit Log

`append_only_metadata_audit_log` 是当前最适合 RadishMind 阶段边界的 durable backend family，原因如下：

- audit store 的核心职责是记录不可变的 metadata-only 事件引用、delivery result reference 与 policy binding，不承载 secret material 或业务 payload。
- append-only 形态天然匹配 audit event 的 ordered record、idempotency reference、writer result reference 和后续 retention / redaction policy。
- 该 family 可以先固定 storage adapter contract，再由后续 task card 选择具体数据库、日志服务或 operator-managed store；不会在本批提前引入 vendor SDK、driver、SQL 或外部 provider 依赖。
- 与 `runtime_event_schema_artifact_implemented` 的 metadata-only schema 对齐，能够避免把 event schema artifact 误写成 runtime persistence ready。

本批不选择 `reserved_managed_audit_log`，因为当前尚未选择云 secret service 或 vendor managed audit service；也不选择 `reserved_operator_managed_audit_store`，因为 operator-managed store 的运行手册、备份、恢复、retention 和 access review 仍未形成独立 runtime 证据。

## Future Backend Contract Requirements

后续 storage adapter / runtime task card 必须继续满足：

- 只接收 `audit_ref`、`event_ref`、`event_kind`、`event_version`、`request_id`、`workspace_ref`、`environment`、`provider_profile_ref`、`backend_profile_ref`、`secret_ref_key`、`credential_handle_ref`、`approval_evidence_ref`、`idempotency_key_ref`、`delivery_result_ref`、`retention_policy_ref`、`redaction_profile_ref`、`policy_version`、`failure_code`、`failure_boundary` 和 `sanitized_diagnostic`。
- event write 必须 append-only，更新语义只能通过后续 event 或 delivery result reference 表达。
- storage adapter 必须 fail closed；缺少 durable backend、writer runtime、idempotency runtime、delivery runtime、approval runtime、credential handle runtime、backend health runtime 或 no leakage smoke runtime 时不得写 audit success。
- runtime task card 必须独立说明 backend product、driver / client、connection lifecycle、migration / schema marker、offline smoke、negative leakage scan 和 rollback / retention evidence。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_concrete_durable_backend_selection_dependency_missing` | `dependency_chain` | 必需 readiness / entry review / blocker matrix 证据缺失 |
| `audit_store_concrete_durable_backend_selection_family_missing` | `selection_gate` | 未固定 selected backend family |
| `audit_store_concrete_durable_backend_selection_invalid_family` | `selection_gate` | 选择了非 allowlist backend family |
| `audit_store_concrete_durable_backend_selection_runtime_forbidden` | `runtime_gate` | 本批创建 storage adapter runtime、audit store runtime 或 writer runtime |
| `audit_store_concrete_durable_backend_selection_database_provider_forbidden` | `database_boundary` | 本批创建 DB provider、driver、DSN parser、SQL、migration 或 schema marker |
| `audit_store_concrete_durable_backend_selection_secret_material_detected` | `artifact_guard` | 文档、fixture 或 diagnostics 出现 secret-bearing material |
| `audit_store_concrete_durable_backend_selection_scope_overreach` | `implementation_boundary` | 本批合入 production resolver runtime、repository mode、public API 或 provider/cloud client |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_concrete_durable_backend_selection_review_status`
- `selection_decision`
- `selected_backend_family`
- `selected_reserved_candidate`
- `selection_scope`
- `gate`
- `gate_status`
- `storage_adapter_runtime_status`
- `database_connection_provider_status`
- `audit_store_runtime_task_card_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、secret value、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、database hostname、database error detail、cloud credential、credential payload、完整 secret ref value、完整 credential handle、raw request payload、raw response payload、raw audit payload、secret-derived hash 或 provider error detail。

## No Fallback / No Side Effects

- 不允许把 memory store、fake resolver、test profile、local smoke profile、developer env、static fixture、sample、mock provider、repository memory store、static handoff envelope、static schema artifact 或 static boundary 文档提升为 durable audit backend。
- 不允许把 backend family selection 写成 storage adapter ready、database ready、audit store ready、production resolver ready、repository mode ready、production API ready 或 production ready。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / review fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不创建 storage adapter runtime、不创建 audit store、不创建 audit writer、不写 audit event、不执行 delivery、不执行 idempotency decision、不连接数据库、不打开 driver、不运行 SQL、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-concrete-durable-backend-selection-review-v1.py`

不得新增或启用 storage adapter runtime、durable audit backend runtime、audit store runtime implementation task card、audit store runtime、audit writer runtime、audit event writer / runner、audit delivery runtime、audit idempotency runtime、duplicate detector runtime、retry executor、replay executor、production resolver runtime、cloud secret SDK / client、secret value fixture、production credential file、credential payload runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、database connection provider、DB driver、DSN parser、SQL migration、schema marker reader / writer、migration runner、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批完成后，durable backend family 已有静态选择，但 runtime 仍 blocked。后续若继续 audit store，应重新评审 `audit_store_runtime_blocker_matrix` 或选择 writer / idempotency / delivery / approval / credential handle / backend health / no leakage smoke 中一个仍 blocked 的 runtime 依赖推进，不能直接创建 audit store runtime 或 production resolver runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-concrete-durable-backend-selection-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
