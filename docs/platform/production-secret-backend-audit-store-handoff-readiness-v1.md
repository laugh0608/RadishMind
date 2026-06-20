# Production Secret Backend Audit Store Handoff Readiness v1

更新时间：2026-06-20

## 文档目的

本文档固定 production secret backend 真实 resolver runtime implementation task card 之前的 production secret audit store handoff boundary。

对应切片：`production-secret-backend-audit-store-handoff-readiness-v1`。

结论：状态为 `audit_store_handoff_readiness_defined`。本批只定义 future audit store handoff 的 reference-only handoff envelope、允许 metadata、禁止 payload / secret material、event kind allowlist、secret ref / provider profile / environment / backend profile binding、credential handle boundary、operator approval evidence、rotation / retention / redaction policy、delivery semantics、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不创建 audit store，不创建 audit writer，不写 audit event，不连接数据库，不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不创建 credential payload 或 credential handle runtime，不启用 repository mode，不执行 no secret leakage smoke runtime，也不新增 public production API。

## 输入证据

- `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 已固定真实 resolver runtime implementation task card 仍 blocked。
- `production-secret-backend-rotation-audit-policy-readiness-v1` 只固定 rotation / audit policy，不提供 audit store handoff。
- `production-secret-backend-operator-approval-runtime-evidence-readiness-v1` 已固定 operator approval runtime evidence boundary，但不执行 approval runtime。
- `production-secret-backend-credential-handle-runtime-boundary-readiness-v1` 已固定 credential handle boundary，但不创建 credential handle runtime。
- `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1` 已固定 no leakage smoke runtime strategy，但不创建 smoke runtime。
- `production-secret-backend-resolver-backend-profile-selection-readiness-v1` 已固定 backend profile selection，但不创建 backend runtime 或 health check。

## Audit Handoff Boundary

future audit store handoff 只能是 reference-only、metadata-only 的 handoff envelope。它不是 audit store runtime，不是 audit writer，不是 resolver runtime output，不是 credential payload，不是 secret value，也不是可被重放为 secret 解析请求的结构化 payload。

允许作为 future audit handoff metadata 暴露的字段只限：

| 字段 | 要求 |
| --- | --- |
| `audit_handoff_id` | opaque、不可从 secret ref / provider profile / operator claim 推导 |
| `audit_event_id` | 只允许 reference id，不表示事件已写入 store |
| `event_kind` | 必须来自固定 event kind allowlist |
| `environment` | 必须与 secret ref、provider profile、backend profile 绑定一致 |
| `provider` / `provider_profile` | 只允许稳定短键或 profile id |
| `secret_ref_key_status` | 只描述 reference key 是否存在，不输出 full secret ref |
| `secret_ref_version_ref` | 只允许版本引用，不允许 secret value 或 backend URL |
| `backend_profile_ref` | 只允许静态 profile reference |
| `credential_handle_boundary_ref` | 只允许 readiness reference，不是 handle runtime output |
| `operator_approval_evidence_ref` | 只允许 approval evidence reference，不表示 approval runtime executed |
| `approval_ticket_ref` / `approval_window_ref` | 只允许 ticket / window reference |
| `request_id` / `audit_ref` | 只允许引用，不写 audit store |
| `runbook_version` / `policy_version` / `rotation_policy_version` | 只允许策略版本 |
| `delivery_mode` | 只允许 future delivery 语义，不执行 delivery |
| `idempotency_key_ref` | 只允许幂等引用，不包含 payload hash 原文 |
| `retention_policy_ref` / `redaction_profile_ref` | 只允许保留和脱敏策略引用 |
| `failure_code` / `sanitized_diagnostic` | 只允许 fail-closed 脱敏诊断 |
| `timestamp` | 只允许事件时间 metadata |

禁止任何 future audit handoff 或 fixture 输出：

- credential payload
- secret value
- password、token、API key、cloud credential
- provider raw URL、resolver backend URL、DSN、database hostname
- full secret ref value
- full credential handle
- full operator identity claim
- authorization header、cookie、完整用户 claim
- raw audit payload、raw request payload、raw response payload
- 可解码、可反序列化或可拼接成 credential 的结构化 blob

## Binding Preconditions

future audit handoff 必须在任何 store write 或 writer execution 前满足以下绑定前置：

| binding | 要求 |
| --- | --- |
| runbook binding | 必须引用 operator runbook / negative gates readiness |
| rotation policy binding | 必须引用 rotation / audit policy readiness |
| secret ref binding | 只能记录 secret ref key status 与 version reference |
| provider profile binding | provider / provider profile 必须与 provider profile secret binding evidence 对齐 |
| environment binding | test / production / local 不允许互相 fallback |
| backend profile binding | 必须引用 resolver backend profile selection readiness |
| credential handle boundary binding | 只能引用 credential handle boundary readiness，不创建 handle runtime |
| operator approval evidence binding | 只能引用 approval evidence readiness，不执行 approval runtime |
| no leakage strategy binding | 必须引用 no leakage smoke runtime strategy，runtime 未创建时不能声明 production ready |
| retention / redaction dependency | 必须定义 retention policy reference 与 redaction profile reference |
| delivery semantics dependency | 必须定义 delivery mode、idempotency key reference 和 fail-closed 写入语义 |

## Event Kind Allowlist

本批只固定 future audit event kind 名称，不写 audit store：

- `secret_resolution_requested`
- `secret_resolution_denied`
- `secret_resolution_failed_closed`
- `credential_handle_boundary_checked`
- `operator_approval_evidence_checked`
- `backend_profile_selected`
- `no_leakage_gate_checked`
- `rotation_policy_checked`
- `audit_handoff_failed_closed`

任何 event kind 都不能携带 credential payload、secret material、完整 operator claim、full credential handle 或 raw request / response payload。

## Handoff Lifecycle

本批只固定 future lifecycle 名称，不创建 runtime 状态机：

| lifecycle state | 含义 |
| --- | --- |
| `handoff_planned` | 仅完成 audit handoff 边界规划 |
| `metadata_bound` | secret ref / provider profile / environment metadata 已绑定 |
| `event_reference_prepared` | 已形成 reference-only event envelope |
| `delivery_pending` | 等待 future audit writer，不表示已写 store |
| `delivery_blocked_store_not_created` | audit store 未创建，必须 fail closed |
| `delivered_metadata_only` | future runtime 只允许写 metadata，不写 payload |
| `delivery_failed_closed` | delivery 失败，只返回 sanitized failure |
| `redaction_required` | 写入前必须应用 redaction profile |
| `retention_policy_required` | 写入前必须绑定 retention policy |

## Gate Matrix

| gate | 状态 | 说明 |
| --- | --- | --- |
| audit handoff envelope | `defined_static_only` | 只定义 handoff shape，不创建 runtime |
| metadata allowlist | `fixed_sanitized_only` | 只允许 reference-only metadata 字段 |
| payload and secret material | `forbidden` | payload / secret material 一律禁止 |
| event kind allowlist | `defined_static_only` | 只定义 future event kind |
| runbook binding | `required_before_runtime` | 未绑定时 fail closed |
| rotation policy binding | `required_before_runtime` | 未绑定时 fail closed |
| secret ref binding | `required_before_runtime` | 未绑定时 fail closed |
| provider profile binding | `required_before_runtime` | 未绑定时 fail closed |
| environment binding | `required_no_cross_environment` | 禁止跨环境 fallback |
| backend profile binding | `required_before_runtime` | 未绑定时 fail closed |
| credential handle boundary binding | `required_before_runtime` | 只能引用 boundary readiness |
| operator approval evidence binding | `required_before_runtime` | 只能引用 evidence readiness |
| no leakage strategy binding | `required_before_runtime` | strategy 未定义时 fail closed |
| redaction and retention policy | `required_before_runtime` | 写入前必须绑定 |
| delivery semantics | `required_before_runtime` | 写入语义必须 fail closed |
| audit store | `not_created` | 本批不创建 store |
| audit writer | `not_created` | 本批不创建 writer |
| audit event write | `not_executed` | 本批不写 event |
| production resolver runtime | `not_created` | 本批不创建 runtime |
| cloud secret service | `forbidden` | 不调用云 secret 服务 |
| database / repository / API | `blocked` | 不连接 DB、不启用 repository mode、不创建 public production API |

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_handoff_boundary_missing` | `audit_handoff` | 缺少 audit handoff boundary |
| `audit_handoff_envelope_missing` | `audit_handoff` | 未定义 handoff envelope |
| `audit_handoff_metadata_allowlist_missing` | `metadata_contract` | 未定义允许 metadata |
| `audit_handoff_payload_detected` | `artifact_guard` | 文档、fixture 或 future output 出现 payload |
| `audit_handoff_secret_material_detected` | `artifact_guard` | 出现 secret value、token、DSN 或 cloud credential |
| `audit_handoff_event_kind_invalid` | `event_kind` | event kind 不在 allowlist |
| `audit_handoff_secret_ref_binding_missing` | `secret_ref` | 缺少 secret ref key status / version reference |
| `audit_handoff_provider_profile_binding_missing` | `provider_profile` | 缺少 provider profile binding |
| `audit_handoff_environment_binding_mismatch` | `environment_binding` | audit handoff 与 environment binding 不一致 |
| `audit_handoff_backend_profile_binding_missing` | `backend_profile` | 缺少 backend profile reference |
| `audit_handoff_credential_boundary_missing` | `credential_boundary` | 缺少 credential handle boundary reference |
| `audit_handoff_operator_approval_reference_missing` | `operator_approval` | 缺少 operator approval evidence reference |
| `audit_handoff_rotation_policy_missing` | `rotation_policy` | 缺少 rotation / audit policy reference |
| `audit_handoff_redaction_policy_missing` | `redaction` | 缺少 redaction profile reference |
| `audit_handoff_retention_policy_missing` | `retention` | 缺少 retention policy reference |
| `audit_handoff_delivery_semantics_missing` | `delivery` | 缺少 delivery mode 或 fail-closed write semantics |
| `audit_handoff_idempotency_missing` | `delivery` | 缺少 idempotency key reference |
| `audit_handoff_store_created_forbidden` | `artifact_guard` | 本批创建 audit store |
| `audit_handoff_writer_created_forbidden` | `artifact_guard` | 本批创建 audit writer |
| `audit_handoff_write_executed_forbidden` | `no_side_effects` | 本批写 audit event |
| `audit_handoff_side_effect_forbidden` | `no_side_effects` | 本批读取 secret、调用云服务、连接 DB 或调用 API |
| `audit_handoff_fallback_forbidden` | `no_fallback` | handoff 缺失时 fallback 到 runbook、fixture、memory store 或 fake resolver |
| `audit_handoff_scope_overreach` | `implementation_boundary` | 本批合入 production resolver runtime、DB provider、audit store runtime 或 public API |

所有失败必须 fail closed，不返回 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、完整 secret ref value、完整 credential handle、完整 operator claim、完整用户 claim、authorization header、cookie、raw request payload、raw response payload 或 raw audit payload。

## Sanitized Diagnostics

允许输出：

- `audit_handoff_status`
- `audit_store_status`
- `audit_writer_status`
- `audit_event_delivery_status`
- `event_kind_status`
- `environment`
- `provider`
- `provider_profile`
- `secret_ref_key_status`
- `secret_ref_version_ref_status`
- `backend_profile_ref_status`
- `credential_handle_boundary_status`
- `operator_approval_evidence_status`
- `redaction_profile_status`
- `retention_policy_status`
- `delivery_mode_status`
- `idempotency_key_ref_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、secret value、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、full secret ref value、full credential handle、完整 operator claim、完整用户 claim、authorization header、cookie、raw request payload、raw response payload 或 raw audit payload。

## No Fallback

- audit handoff boundary 缺失时真实 resolver runtime implementation task card 必须继续 blocked。
- 不允许 audit handoff fallback 到 operator runbook 文本、developer env plaintext、fixture credential、committed value、sample、mock provider、local-smoke profile、fake resolver runtime、repository memory store、console local record 或 previously approved test evidence。
- 不允许 production audit handoff fallback 到 test audit handoff，也不允许 test audit handoff fallback 到 production audit handoff。
- 不允许缺少 retention / redaction / idempotency / delivery semantics 时创建 audit success。
- 不把 audit handoff readiness 写成 audit store ready、audit writer ready、audit event written、credential resolved、production resolver runtime ready、production secret backend ready 或 public production API ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema、已有 readiness fixture 和 `check-repo.py` 注册顺序，不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 production resolver、不执行 approval runtime、不执行 smoke runtime、不创建 credential payload、不创建 credential handle、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不创建 audit store、不写 audit event、不调用 production API。

side effect counters 必须保持：

- `real_secret_read_count=0`
- `environment_secret_read_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `production_resolver_call_count=0`
- `operator_approval_runtime_execution_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `credential_payload_created_count=0`
- `credential_handle_created_count=0`
- `credential_handle_runtime_created_count=0`
- `audit_store_created_count=0`
- `audit_writer_created_count=0`
- `audit_event_write_count=0`
- `audit_delivery_execution_count=0`
- `database_connection_count=0`
- `driver_open_count=0`
- `sql_execution_count=0`
- `schema_marker_read_count=0`
- `schema_marker_write_count=0`
- `repository_mode_enablement_count=0`
- `production_api_call_count=0`

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-handoff-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-handoff-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-handoff-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py`

不得新增或启用以下 artifact：

- audit store runtime
- audit writer
- audit repository
- audit delivery runtime
- audit retention runtime
- production resolver runtime
- production resolver implementation task card
- production resolver backend client
- cloud secret SDK
- secret value fixture
- production credential file
- credential payload runtime
- credential handle runtime
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

本批只把真实 resolver runtime implementation entry review 中的 audit handoff blocker 固定为静态前置证据。真实 audit store、audit writer、store schema、DB provider、repository mode、production resolver runtime、cloud secret backend 和 production API 仍必须作为独立目标推进。

下一步若继续 production secret backend，应优先推进 `resolver-backend-health-boundary-readiness`，仍不得直接创建 production resolver runtime implementation task card。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
