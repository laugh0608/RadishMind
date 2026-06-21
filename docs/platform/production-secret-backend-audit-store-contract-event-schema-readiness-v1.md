# Production Secret Backend Audit Store Contract / Event Schema Readiness v1

更新时间：2026-06-21

## 文档目的

本文档固定 production secret backend audit store runtime task card 之前的 audit store contract / event schema readiness。

对应切片：`production-secret-backend-audit-store-contract-event-schema-readiness-v1`。

结论：状态为 `audit_store_contract_event_schema_readiness_defined`。本批只定义 future audit store 的 metadata-only event schema、event version、event kind allowlist、required / optional fields、reference-only writer input、idempotency key reference、delivery result envelope、retention / redaction binding、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。它消费 audit store handoff readiness 与 audit store runtime implementation entry review，但不创建 audit store runtime implementation task card，不创建 audit store runtime，不创建 audit writer，不写 audit event，不创建 runtime event schema，不连接数据库，不读取真实 secret，不调用云 secret 服务，不创建 credential payload，不执行 approval runtime，不执行 backend health check，不执行 no leakage smoke runtime，不创建 production resolver runtime，不启用 repository mode，也不新增 public production API。

## 输入证据

- `production-secret-backend-audit-store-handoff-readiness-v1` 已固定 reference-only handoff envelope、event kind allowlist、retention / redaction / delivery 依赖和 no side effects。
- `production-secret-backend-audit-store-runtime-implementation-entry-review-v1` 已固定 audit store runtime implementation task card 仍 blocked。
- `production-secret-backend-operator-approval-runtime-implementation-entry-review-v1` 已固定 approval runtime task card 仍 blocked。
- `production-secret-backend-credential-handle-runtime-implementation-entry-review-v1` 已固定 credential handle runtime task card 仍 blocked。
- `production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1` 已固定 backend health runtime task card 仍 blocked。
- `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1` 已固定 no leakage smoke runtime task card 仍 blocked。
- `production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1` 已固定 production resolver runtime implementation task card 仍 blocked。
- `production-ops-secret-backend-implementation-readiness` 仍保持 `production_secret_backend=not_satisfied`、`audit_store_runtime_status=not_created`、`audit_writer_status=not_created` 和 `audit_event_delivery_status=not_executed`。

## Contract Boundary

future audit store contract 只能处理 reference-only metadata。event schema 是后续 runtime task card 的输入约束，不是 runtime schema artifact，也不表示 store / writer 已创建。

| boundary | 状态 | 说明 |
| --- | --- | --- |
| contract readiness | `defined_without_store_runtime` | 只定义 future contract 与 event schema |
| runtime task card | `still_blocked_before_task_card` | 本批不创建 audit store runtime implementation task card |
| audit event schema | `static_contract_defined` | 只固定静态 schema 形状和字段规则 |
| runtime event schema | `not_created` | 不创建 runtime schema 或 writer type |
| audit store runtime | `not_created` | 不创建 store runtime |
| audit writer | `not_created` | 不创建 writer |
| audit event delivery | `not_executed` | 不写 event，不执行 delivery |
| idempotency | `reference_required_before_runtime` | 只允许 opaque idempotency key reference |
| retention / redaction | `policy_binding_required_before_runtime` | 只引用策略，不执行 runtime |
| production resolver runtime | `not_created` | 不创建 resolver runtime |
| repository / DB / API | `blocked` | 不连接数据库，不启用 repository mode，不创建 public production API |

## Event Schema

future audit event schema 必须是 metadata-only。所有字段都必须属于 allowlist；任何 runtime 如果需要新增字段，必须先更新本契约和 checker。

必填字段：

- `event_id`
- `event_version`
- `event_kind`
- `occurred_at`
- `environment`
- `provider`
- `provider_profile`
- `backend_profile_ref`
- `secret_ref_key_status`
- `secret_ref_version_ref`
- `operator_approval_ref`
- `credential_handle_boundary_ref`
- `request_id`
- `audit_ref`
- `policy_version`
- `retention_policy_ref`
- `redaction_profile_ref`
- `idempotency_key_ref`
- `delivery_mode`

可选字段：

- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `rotation_policy_version`
- `runbook_version`
- `approval_ticket_ref`
- `approval_window_ref`
- `backend_health_ref`
- `no_leakage_smoke_ref`

event kind allowlist：

- `secret_resolution_requested`
- `secret_resolution_denied`
- `secret_resolution_failed_closed`
- `credential_handle_boundary_checked`
- `operator_approval_evidence_checked`
- `backend_profile_selected`
- `backend_health_gate_checked`
- `no_leakage_gate_checked`
- `rotation_policy_checked`
- `audit_handoff_failed_closed`

禁止字段：

- `secret_value`
- `raw_secret`
- `password`
- `token`
- `api_key`
- `provider_raw_url`
- `resolver_backend_url`
- `dsn`
- `database_hostname`
- `cloud_credential`
- `credential_payload`
- `full_secret_ref_value`
- `full_credential_handle`
- `full_operator_identity_claim`
- `full_user_claim`
- `authorization_header`
- `cookie`
- `raw_request_payload`
- `raw_response_payload`
- `raw_audit_payload`

## Writer Contract

future writer input 只能接收 event schema allowlist 字段。writer 必须在执行前完成以下绑定：

| contract item | 要求 |
| --- | --- |
| store ownership | 必须由后续 runtime task card 明确，当前为 blocked |
| writer ownership | 必须与 store ownership 分离，当前不创建 writer |
| event version | 必须固定为 schema version reference，不能从 payload 派生 |
| idempotency | 必须使用 opaque `idempotency_key_ref`，不能使用 raw payload hash |
| delivery mode | 只能使用 fail-closed delivery mode，不能 fallback 到 memory store |
| retention | 必须绑定 `retention_policy_ref` |
| redaction | 必须绑定 `redaction_profile_ref` |
| operator approval | 只能引用 approval evidence / runtime readiness，不能执行 approval runtime |
| credential handle | 只能引用 boundary / runtime readiness，不能创建 handle |
| backend health | 只能引用 health readiness，不能执行 health check |
| no leakage | 只能引用 strategy / entry review，不能执行 smoke |

future delivery result 只能返回：

- `delivery_status`
- `event_id`
- `audit_ref`
- `idempotency_key_ref_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_contract_event_schema_missing` | `event_schema` | 缺少 metadata-only event schema |
| `audit_store_contract_event_kind_invalid` | `event_kind` | event kind 不在 allowlist |
| `audit_store_contract_required_field_missing` | `event_schema` | 缺少必填 metadata 字段 |
| `audit_store_contract_forbidden_field_detected` | `artifact_guard` | 文档、fixture 或 future output 出现禁止字段 |
| `audit_store_contract_secret_material_detected` | `artifact_guard` | 出现 secret value、credential payload、DSN 或 provider raw URL |
| `audit_store_contract_writer_boundary_missing` | `writer_contract` | 缺少 writer input / output 边界 |
| `audit_store_contract_idempotency_missing` | `delivery` | 缺少 idempotency key reference |
| `audit_store_contract_retention_redaction_missing` | `retention_redaction` | 缺少 retention 或 redaction policy reference |
| `audit_store_contract_delivery_semantics_missing` | `delivery` | 缺少 fail-closed delivery semantics |
| `audit_store_contract_runtime_task_card_blocked` | `implementation_gate` | audit store runtime task card 仍 blocked |
| `audit_store_contract_runtime_created_forbidden` | `artifact_guard` | 本批创建 audit store runtime |
| `audit_store_contract_writer_created_forbidden` | `artifact_guard` | 本批创建 audit writer |
| `audit_store_contract_event_written_forbidden` | `no_side_effects` | 本批写 audit event 或执行 delivery |
| `audit_store_contract_repository_mode_forbidden` | `repository_mode` | 本批启用 repository mode |
| `audit_store_contract_scope_overreach` | `implementation_boundary` | 本批合入 resolver runtime、DB provider、approval executor、backend health runtime 或 public API |

所有失败必须 fail closed，不返回 raw secret、secret value、password、token、API key、provider raw URL、resolver backend URL、DSN、cloud credential、database hostname、database error detail、credential payload、full secret ref value、full credential handle、完整 operator claim、完整用户 claim、authorization header、cookie、raw request payload、raw response payload 或 raw audit payload。

## Sanitized Diagnostics

允许输出：

- `audit_store_contract_status`
- `event_schema_status`
- `event_version_status`
- `event_kind_status`
- `writer_contract_status`
- `audit_store_runtime_status`
- `audit_writer_status`
- `audit_event_delivery_status`
- `idempotency_key_ref_status`
- `retention_policy_status`
- `redaction_profile_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、secret value、password、token、API key、provider raw URL、resolver backend URL、DSN、cloud credential、database hostname、database error detail、credential payload、full secret ref value、full credential handle、完整 operator claim、完整用户 claim、authorization header、cookie、raw request payload、raw response payload 或 raw audit payload。

## No Fallback

- 不允许 audit store contract 缺失时 fallback 到 handoff envelope、operator runbook 文本、developer env plaintext、fixture credential、committed value、sample、mock provider、local-smoke profile、fake resolver runtime、repository memory store、audit memory store 或历史 smoke evidence。
- 不允许 production audit contract fallback 到 test audit contract，也不允许 test audit contract fallback 到 production audit contract。
- 不允许缺少 event schema、writer contract、idempotency、retention、redaction 或 delivery semantics 时创建 audit success。
- 不把本 readiness 写成 audit store runtime ready、audit writer ready、audit event written、credential resolved、production resolver runtime ready、production secret backend ready 或 public production API ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、已有 readiness fixture 和 `check-repo.py` 注册顺序，不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 production resolver、不执行 approval runtime、不执行 backend health check、不执行 smoke runtime、不创建 credential payload、不创建 credential handle、不创建 audit store、不创建 audit writer、不写 audit event、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不调用 production API。

side effect counters 必须保持：

- `real_secret_read_count=0`
- `environment_secret_read_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `production_resolver_call_count=0`
- `audit_store_runtime_created_count=0`
- `audit_writer_created_count=0`
- `audit_event_schema_created_count=0`
- `audit_event_write_count=0`
- `audit_delivery_execution_count=0`
- `operator_approval_runtime_execution_count=0`
- `credential_payload_created_count=0`
- `credential_handle_created_count=0`
- `backend_health_check_count=0`
- `no_secret_leakage_smoke_runtime_execution_count=0`
- `network_call_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `database_connection_count=0`
- `driver_open_count=0`
- `sql_execution_count=0`
- `schema_marker_read_count=0`
- `schema_marker_write_count=0`
- `repository_mode_enablement_count=0`
- `production_api_call_count=0`

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-contract-event-schema-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-contract-event-schema-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-contract-event-schema-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-contract-event-schema-readiness-v1.py`

不得新增或启用以下 artifact：

- audit store runtime implementation task card
- audit store runtime
- audit writer
- audit event writer / runner
- runtime event schema
- durable audit backend
- production resolver runtime
- production resolver backend client
- cloud secret SDK
- secret value fixture
- production credential file
- credential payload runtime
- credential handle runtime
- approval runtime / approval executor
- backend health runtime
- backend health client
- no secret leakage smoke runtime
- database connection provider
- DB driver / DSN parser
- connection factory
- SQL migration
- schema marker reader / writer
- migration runner
- workflow saved draft repository mode runtime
- public production API

## 后续推进

本批之后不应直接创建 audit store runtime implementation task card。后续应先用本 contract / event schema readiness 重新评审 audit store runtime implementation 是否仍 blocked；只有当 store ownership、writer boundary、event schema runtime、idempotency、retention、redaction、operator approval runtime、credential handle runtime、backend health runtime、no leakage smoke runtime 和 production preflight 均满足时，才能创建对应 runtime implementation task card。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-contract-event-schema-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
