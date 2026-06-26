# Production Secret Backend Audit Store Delivery / Idempotency Runtime Boundary Readiness v1

更新时间：2026-06-21

## 文档目的

本文档固定 production secret backend audit store runtime task card 之前的 audit store delivery / idempotency runtime boundary。

对应切片：`production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1`。

结论：状态为 `audit_store_delivery_idempotency_runtime_boundary_readiness_defined`。本批只定义 future audit store delivery owner、idempotency key owner、duplicate handling、retry / failure semantics、delivery result envelope、metadata-only diagnostics、failure mapping、no fallback、no side effects、artifact guard 和 implementation readiness alignment；不创建 audit store runtime implementation task card，不创建 audit store runtime，不创建 audit writer，不写 audit event，不创建 runtime event schema，不执行 delivery，不创建 idempotency runtime，不连接数据库，不读取真实 secret，不调用云 secret 服务，不创建 credential payload，不执行 approval runtime，不执行 backend health check，不执行 no leakage smoke runtime，不创建 production resolver runtime，不启用 repository mode，也不新增 public production API。

## 输入证据

- `production-secret-backend-audit-store-handoff-readiness-v1` 已固定 reference-only handoff envelope、delivery mode、idempotency key reference 和 fail-closed delivery 依赖。
- `production-secret-backend-audit-store-contract-event-schema-readiness-v1` 已固定 metadata-only event schema、writer input、idempotency key reference、delivery result envelope 和 retention / redaction binding。
- `production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2` 已确认 delivery runtime、runtime event schema、store ownership、writer ownership 与 runtime 依赖仍 blocked。
- `production-secret-backend-audit-store-ownership-boundary-readiness-v1` 已把 delivery / idempotency owner 标记为下一前置边界。
- `production-secret-backend-operator-approval-runtime-implementation-entry-review-v1`、`production-secret-backend-credential-handle-runtime-implementation-entry-review-v1`、`production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1` 和 `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1` 均确认各自 runtime task card 仍 blocked。
- `production-ops-secret-backend-implementation-readiness` 仍保持 `production_secret_backend=not_satisfied`、`audit_store_runtime_status=not_created`、`audit_writer_status=not_created` 和 `audit_event_delivery_status=not_executed`。

## Delivery / Idempotency Boundary

future delivery / idempotency boundary 只能是 reference-only 的运行前置记录。它不是 delivery runtime，不是 duplicate detector，不是 retry executor，不是 audit writer，也不是可用于重放 secret resolution 的 payload。

| boundary item | 本批结论 | 要求 |
| --- | --- | --- |
| delivery owner reference | `defined_static_only` | 只能声明稳定短键，例如 `platform_secret_backend_audit_delivery_boundary` |
| idempotency key owner reference | `defined_static_only` | 只能声明 opaque key owner，不允许 raw payload hash |
| duplicate handling | `static_fail_closed_policy_defined` | 只定义 duplicate / conflict 的 fail-closed 语义，不执行检测 |
| retry / failure semantics | `static_fail_closed_policy_defined` | 只定义 bounded retry policy reference，不执行 retry |
| delivery result envelope | `metadata_only_static_envelope_defined` | 只允许 delivery metadata，不表示 event 已写入 |
| metadata-only diagnostics | `fixed_sanitized_only` | 只允许脱敏状态、failure code 和 reference id |
| audit event delivery | `not_executed` | 不写 event，不执行 delivery，不持久化结果 |
| delivery runtime | `not_created` | 不创建 runtime、runner、queue、scheduler 或 duplicate detector |

future delivery / idempotency record 只允许包含以下 metadata：

- `delivery_boundary_id`
- `delivery_owner_ref`
- `idempotency_key_owner_ref`
- `idempotency_key_ref`
- `idempotency_scope_ref`
- `duplicate_handling_policy_ref`
- `retry_policy_ref`
- `delivery_result_ref`
- `delivery_mode`
- `event_id`
- `event_version`
- `event_kind`
- `request_id`
- `audit_ref`
- `environment`
- `provider`
- `provider_profile`
- `backend_profile_ref`
- `policy_version`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`

禁止 delivery / idempotency record、fixture 或 future diagnostics 输出：

- credential payload
- secret value
- password、token、API key、cloud credential
- provider raw URL、resolver backend URL、DSN、database hostname
- full secret ref value
- full credential handle
- full operator identity claim
- full user claim
- authorization header、cookie
- raw request payload、raw response payload、raw audit payload
- raw payload hash、secret-derived hash 或可反推 payload 的幂等 key 原文

## Duplicate Handling

本批只定义 duplicate handling policy，不创建 duplicate detector：

| duplicate case | future result |
| --- | --- |
| same `idempotency_key_ref` and same metadata reference | `duplicate_same_event_ref_fail_closed`，除非 future runtime 证明已有 metadata-only delivery result 可安全复用 |
| same `idempotency_key_ref` but conflicting event metadata | `duplicate_conflicting_event_ref_fail_closed` |
| missing `idempotency_key_ref` | `idempotency_key_missing_fail_closed` |
| raw payload hash used as key | `idempotency_key_raw_hash_forbidden` |
| cross-environment duplicate | `cross_environment_duplicate_forbidden` |

任何 duplicate handling 都不能读取 raw request / response payload、secret material、credential payload、provider endpoint、database row 或 audit store payload。

## Retry / Failure Semantics

future retry / failure semantics 必须保持 fail-closed：

- retry policy 只能通过 `retry_policy_ref` 引用。
- 不允许无限重试、隐式后台重试或跨环境重试。
- delivery runtime 缺失时不能把 `delivery_status` 写成 success。
- idempotency key 缺失、duplicate conflict、retention / redaction policy 缺失、operator approval runtime 缺失、credential handle runtime 缺失、backend health runtime 缺失或 no leakage runtime 缺失时都必须 fail closed。
- 所有 failure 只允许返回 `failure_code`、`failure_boundary`、`sanitized_diagnostic`、`request_id`、`audit_ref` 和 policy reference。

## Delivery Result Envelope

future delivery result envelope 只能包含：

- `delivery_status`
- `delivery_result_ref`
- `event_id`
- `audit_ref`
- `idempotency_key_ref_status`
- `duplicate_status`
- `retry_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`

该 envelope 只描述 future delivery 结果形状，不表示 delivery result 已持久化，不表示 audit event 已写入，也不表示 audit store runtime 已创建。

## Gate Matrix

| gate | 状态 | 说明 |
| --- | --- | --- |
| delivery / idempotency boundary | `defined_static_only` | 只定义前置边界，不创建 runtime |
| delivery owner reference | `defined_static_only` | 不表示 delivery runtime 已创建 |
| idempotency key owner reference | `defined_static_only` | key 只能 opaque reference |
| duplicate handling | `static_fail_closed_policy_defined` | 不执行 duplicate detection |
| retry / failure semantics | `static_fail_closed_policy_defined` | 不执行 retry |
| delivery result envelope | `metadata_only_static_envelope_defined` | 不持久化 result |
| metadata diagnostics | `fixed_sanitized_only` | 禁止 payload / secret material |
| audit store runtime task card | `still_blocked_before_task_card` | 本批不创建 task card |
| delivery runtime | `not_created` | 本批不创建 runtime |
| idempotency runtime | `not_created` | 本批不创建 runtime |
| duplicate detector runtime | `not_created` | 本批不创建 detector |
| retry runtime | `not_created` | 本批不创建 retry executor |
| audit event delivery | `not_executed` | 本批不写 event |
| production resolver runtime | `not_created` | 本批不创建 resolver runtime |
| cloud secret service | `forbidden` | 不调用云 secret 服务 |
| database / repository / API | `blocked` | 不连接数据库、不启用 repository mode、不创建 public production API |

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_delivery_idempotency_boundary_missing` | `delivery_idempotency_boundary` | 缺少 delivery / idempotency boundary |
| `audit_store_delivery_idempotency_delivery_owner_missing` | `delivery_owner` | 缺少 delivery owner reference |
| `audit_store_delivery_idempotency_key_owner_missing` | `idempotency_owner` | 缺少 idempotency key owner reference |
| `audit_store_delivery_idempotency_key_ref_missing` | `idempotency_key` | 缺少 idempotency key reference |
| `audit_store_delivery_idempotency_raw_hash_detected` | `artifact_guard` | 使用 raw payload hash 或 secret-derived hash |
| `audit_store_delivery_idempotency_duplicate_policy_missing` | `duplicate_handling` | 缺少 duplicate handling policy |
| `audit_store_delivery_idempotency_duplicate_conflict_not_fail_closed` | `duplicate_handling` | duplicate conflict 未 fail closed |
| `audit_store_delivery_idempotency_retry_semantics_missing` | `retry_failure` | 缺少 retry / failure semantics |
| `audit_store_delivery_idempotency_delivery_result_envelope_missing` | `delivery_result` | 缺少 metadata-only delivery result envelope |
| `audit_store_delivery_idempotency_metadata_diagnostics_missing` | `diagnostics` | 缺少 metadata-only diagnostics allowlist |
| `audit_store_delivery_idempotency_secret_material_detected` | `artifact_guard` | 文档、fixture 或 checker 出现 secret-bearing material |
| `audit_store_delivery_idempotency_runtime_task_card_created` | `artifact_guard` | 本批创建 audit store runtime implementation task card |
| `audit_store_delivery_idempotency_runtime_created_forbidden` | `artifact_guard` | 本批创建 delivery / idempotency runtime |
| `audit_store_delivery_idempotency_writer_created_forbidden` | `artifact_guard` | 本批创建 audit writer |
| `audit_store_delivery_idempotency_event_written_forbidden` | `no_side_effects` | 本批写 audit event |
| `audit_store_delivery_idempotency_delivery_executed_forbidden` | `no_side_effects` | 本批执行 delivery |
| `audit_store_delivery_idempotency_database_connection_forbidden` | `no_side_effects` | 本批连接数据库或打开 driver |
| `audit_store_delivery_idempotency_fallback_forbidden` | `no_fallback` | 缺少依赖时 fallback 到 fake / env / mock / memory store |
| `audit_store_delivery_idempotency_scope_overreach` | `implementation_boundary` | 本批合入 runtime、DB provider、approval executor、backend health runtime 或 public API |

所有失败必须 fail closed，不返回 raw secret、secret value、password、token、API key、provider raw URL、resolver backend URL、DSN、cloud credential、database hostname、database error detail、credential payload、full secret ref value、full credential handle、完整 operator claim、完整用户 claim、authorization header、cookie、raw request payload、raw response payload、raw audit payload、raw payload hash 或 secret-derived hash。

## Sanitized Diagnostics

允许输出：

- `audit_store_delivery_idempotency_boundary_status`
- `delivery_owner_ref_status`
- `idempotency_key_owner_ref_status`
- `idempotency_key_ref_status`
- `duplicate_handling_status`
- `retry_failure_semantics_status`
- `delivery_result_envelope_status`
- `metadata_only_diagnostics_status`
- `audit_store_runtime_task_card_status`
- `audit_store_runtime_status`
- `audit_writer_status`
- `audit_event_delivery_status`
- `delivery_runtime_status`
- `idempotency_runtime_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、secret value、password、token、API key、provider raw URL、resolver backend URL、DSN、cloud credential、database hostname、database error detail、credential payload、full secret ref value、full credential handle、完整 operator claim、完整用户 claim、authorization header、cookie、raw request payload、raw response payload、raw audit payload、raw payload hash 或 secret-derived hash。

## No Fallback

- delivery / idempotency boundary 缺失时 audit store runtime implementation task card 必须继续 blocked。
- 不允许 fallback 到 static handoff envelope、static contract、operator runbook 文本、developer env plaintext、fixture credential、committed value、sample、mock provider、local-smoke profile、fake resolver runtime、repository memory store、audit memory store、previously delivered test evidence 或跨环境 idempotency record。
- 不允许缺少 delivery owner、idempotency key owner、duplicate handling、retry semantics、delivery result envelope、retention / redaction policy 或 runtime dependency 时创建 audit success。
- 不把本 readiness 写成 delivery runtime ready、audit store runtime ready、audit writer ready、audit event written、credential resolved、production resolver runtime ready、production secret backend ready 或 public production API ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、已有 readiness fixture 和 `check-repo.py` 注册顺序，不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 production resolver、不执行 backend health check、不执行 approval runtime、不执行 smoke runtime、不创建 credential payload、不创建 credential handle、不创建 audit store、不创建 audit writer、不写 audit event、不执行 delivery、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不调用 production API。

side effect counters 必须保持：

- `real_secret_read_count=0`
- `environment_secret_read_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `production_resolver_call_count=0`
- `audit_store_runtime_created_count=0`
- `audit_writer_created_count=0`
- `runtime_event_schema_created_count=0`
- `audit_event_write_count=0`
- `audit_delivery_execution_count=0`
- `idempotency_runtime_created_count=0`
- `duplicate_detector_runtime_created_count=0`
- `retry_runtime_created_count=0`
- `delivery_result_persistence_count=0`
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

- `docs/platform/production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.py`

不得新增或启用以下 artifact：

- audit store runtime implementation task card
- audit store runtime
- audit writer
- audit event writer / runner
- runtime event schema artifact
- audit delivery runtime
- audit idempotency runtime
- duplicate detector runtime
- retry executor
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

本批之后仍不应直接创建 audit store runtime implementation task card。后续应先重新评审 audit store runtime implementation entry，确认 store / writer ownership、runtime event schema materialization、delivery / idempotency runtime boundary、operator approval runtime、credential handle runtime、backend health runtime、no leakage smoke runtime 和 production preflight 是否全部满足；若任一 runtime 依赖仍 blocked，audit store runtime implementation task card 继续 blocked。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-ownership-boundary-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
