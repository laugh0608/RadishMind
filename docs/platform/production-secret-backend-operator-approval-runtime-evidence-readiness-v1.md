# Production Secret Backend Operator Approval Runtime Evidence Readiness v1

更新时间：2026-06-20

## 文档目的

本文档固定 production secret backend 真实 resolver runtime implementation task card 之前的 operator approval runtime evidence boundary。

对应切片：`production-secret-backend-operator-approval-runtime-evidence-readiness-v1`。

结论：状态为 `operator_approval_runtime_evidence_readiness_defined`。本批只定义 future operator approval runtime evidence 的证据形状、允许 metadata、禁止 secret material、approval subject binding、operator identity / ticket / change window 前置、lifecycle 状态、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不执行 operator approval runtime，不创建 approval executor，不写 audit store，不创建 credential handle runtime，不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不连接数据库，不启用 repository mode，不执行 no secret leakage smoke runtime，也不新增 public production API。

## 输入证据

- `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 已固定真实 resolver runtime implementation task card 仍 blocked。
- `production-secret-backend-operator-runbook-negative-gates-readiness-v1` 只固定 runbook 与 negative gates，不提供 runtime evidence。
- `production-secret-backend-rotation-audit-policy-readiness-v1` 只固定 rotation / audit policy，不写 audit store。
- `production-secret-backend-resolver-backend-profile-selection-readiness-v1` 已固定 backend profile static selection，但不创建 backend runtime 或 health check。
- `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1` 已固定 no leakage smoke runtime strategy，但不创建 smoke runtime。
- `production-secret-backend-credential-handle-runtime-boundary-readiness-v1` 已固定 credential handle boundary，但不创建 credential handle runtime。

## Approval Evidence Boundary

future operator approval runtime evidence 只能是 reference-only、metadata-only 的审批证据。它不是 secret value，不是 credential payload，不是 credential handle runtime output，不是 provider token，不是 DSN，不是云 secret backend URL，也不能包含可从字符串中还原 secret material 的 path、query、claim、header 或结构化 payload。

允许作为 future approval evidence metadata 暴露的字段只限：

| 字段 | 要求 |
| --- | --- |
| `approval_evidence_id` | opaque、不可从 secret ref / provider profile / user claim 推导 |
| `approval_evidence_kind` | 只允许 `operator_approval_reference` 或 `approval_revalidation_reference` |
| `approval_status` | 只描述审批状态，不表示 credential value 已暴露 |
| `approval_lifecycle_state` | 必须来自固定 lifecycle allowlist |
| `approval_subject` | 只允许引用 `production_secret_resolver_enablement` 等受控 subject |
| `environment` | 必须与 secret ref、provider profile、backend profile 绑定一致 |
| `provider` / `provider_profile` | 只允许稳定短键或 profile id |
| `secret_ref_key` | 只允许 reference key，不允许 full secret ref value |
| `secret_ref_version_ref` | 只允许版本引用，不允许 secret value 或 backend URL |
| `backend_profile_ref` | 只允许静态 profile reference |
| `credential_handle_boundary_ref` | 只允许 readiness reference，不是 handle runtime output |
| `requested_by_ref` / `approved_by_ref` | 只允许 operator identity reference，不允许完整用户 claim |
| `approval_ticket_ref` / `approval_window_ref` | 只允许 ticket / window reference |
| `runbook_version` / `policy_version` / `rotation_policy_version` | 只允许策略版本 |
| `expires_at` / `revoked_at` | 只允许生命周期 metadata |
| `request_id` / `audit_ref` | 只允许引用，不写 audit store |
| `failure_code` / `sanitized_diagnostic` | 只允许 fail-closed 脱敏诊断 |

禁止任何 future approval evidence 或 fixture 输出：

- credential payload
- secret value
- password、token、API key、cloud credential
- provider raw URL、resolver backend URL、DSN、database hostname
- full secret ref value
- full credential handle
- full operator identity claim
- authorization header、cookie、完整用户 claim
- 可解码、可反序列化或可拼接成 credential 的结构化 blob

## Binding Preconditions

future operator approval runtime 必须在生成任何 approval evidence metadata 前满足以下绑定前置：

| binding | 要求 |
| --- | --- |
| runbook binding | 必须引用 operator runbook / negative gates readiness |
| secret ref binding | `secret_ref_key`、`secret_ref_version_ref` 与 reference-only secret schema 对齐 |
| provider profile binding | provider / provider profile 必须与 provider profile secret binding evidence 对齐 |
| environment binding | test / production / local 不允许互相 fallback |
| backend profile binding | 必须引用 resolver backend profile selection readiness |
| credential handle boundary binding | 只能引用 credential handle boundary readiness，不创建 handle runtime |
| no leakage strategy binding | 必须引用 no leakage smoke runtime strategy，runtime 未创建时不能声明 production ready |
| audit / rotation dependency | 必须引用 audit policy / future audit handoff，不在本批写 store |
| operator identity dependency | requester / approver 必须是 reference-only identity，禁止完整 claim |
| dual control dependency | production approval 不能自批，必须保留 separation-of-duty evidence |
| ticket / change window dependency | production approval 必须绑定 ticket 与 change window reference |

## Approval Lifecycle

本批只固定 future lifecycle 名称，不创建 runtime 状态机：

| lifecycle state | 含义 |
| --- | --- |
| `evidence_planned` | 仅完成 approval evidence 边界规划 |
| `request_metadata_bound` | secret ref / provider profile / environment metadata 已绑定 |
| `approval_pending` | 等待 operator approval evidence |
| `approved_metadata_only` | future runtime 只返回 approval metadata，不返回 payload |
| `denied` | operator 拒绝，后续使用必须 fail closed |
| `expired` | approval evidence 已过期，禁止 fallback |
| `revoked` | approval evidence 已撤销，后续使用必须 fail closed |
| `revalidation_required` | policy / rotation / backend profile 变化后需要重新审批 |
| `execution_blocked` | 证据边界存在但 runtime 未执行 |
| `resolution_failed_closed` | 解析失败，只返回 sanitized failure |

任何状态都不能暴露 credential payload、secret material、完整 operator claim 或 full credential handle。

## Gate Matrix

| gate | 状态 | 说明 |
| --- | --- | --- |
| approval evidence schema | `defined_static_only` | 只定义 evidence shape，不创建 runtime |
| metadata allowlist | `fixed_sanitized_only` | 只允许 reference-only metadata 字段 |
| payload and secret material | `forbidden` | payload / secret material 一律禁止 |
| runbook binding | `required_before_runtime` | 未绑定时 fail closed |
| secret ref binding | `required_before_runtime` | 未绑定时 fail closed |
| provider profile binding | `required_before_runtime` | 未绑定时 fail closed |
| environment binding | `required_no_cross_environment` | 禁止跨环境 fallback |
| backend profile binding | `required_before_runtime` | 未绑定时 fail closed |
| credential handle boundary binding | `required_before_runtime` | 只能引用 boundary readiness |
| no leakage strategy binding | `required_before_runtime` | strategy 未定义时 fail closed |
| audit / rotation dependency | `required_before_runtime` | 只引用 audit handoff，不写 store |
| operator identity reference | `required_reference_only` | 不输出 full claim |
| dual control | `required_before_runtime` | production approval 禁止自批 |
| ticket and change window | `required_before_runtime` | production approval 必须绑定引用 |
| approval runtime | `not_created` | 本批不创建 runtime |
| approval runtime execution | `not_executed` | 本批不执行 runtime |
| production resolver runtime | `not_created` | 本批不创建 runtime |
| cloud secret service | `forbidden` | 不调用云 secret 服务 |
| audit store / DB / repository / API | `blocked` | 不写 audit store、不连接 DB、不启用 repository mode、不创建 public production API |

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `operator_approval_runtime_evidence_missing` | `operator_gate` | 缺少 operator approval runtime evidence boundary |
| `operator_approval_evidence_schema_missing` | `approval_schema` | 未定义 approval evidence shape |
| `operator_approval_metadata_allowlist_missing` | `metadata_contract` | 未定义允许 metadata |
| `operator_approval_payload_detected` | `artifact_guard` | 文档、fixture 或 future output 出现 payload |
| `operator_approval_secret_material_detected` | `artifact_guard` | 出现 secret value、token、DSN 或 cloud credential |
| `operator_approval_subject_binding_missing` | `approval_subject` | 缺少受控 approval subject |
| `operator_approval_runbook_binding_missing` | `runbook` | 缺少 runbook / negative gate binding |
| `operator_approval_secret_ref_binding_missing` | `secret_ref` | 缺少 secret ref key / version reference |
| `operator_approval_provider_profile_binding_missing` | `provider_profile` | 缺少 provider profile binding |
| `operator_approval_environment_binding_mismatch` | `environment_binding` | approval evidence 与 environment binding 不一致 |
| `operator_approval_backend_profile_binding_missing` | `backend_profile` | 缺少 backend profile reference |
| `operator_approval_credential_boundary_missing` | `credential_boundary` | 缺少 credential handle boundary reference |
| `operator_approval_no_leakage_strategy_missing` | `no_secret_leakage` | 缺少 no leakage smoke runtime strategy reference |
| `operator_approval_audit_dependency_missing` | `audit_policy` | 缺少 audit dependency 或 audit handoff reference |
| `operator_approval_rotation_dependency_missing` | `rotation_policy` | 缺少 rotation policy / version dependency |
| `operator_approval_identity_reference_missing` | `operator_identity` | 缺少 requester / approver reference |
| `operator_approval_dual_control_missing` | `operator_identity` | production approval 缺少 separation-of-duty evidence |
| `operator_approval_ticket_missing` | `approval_ticket` | 缺少 ticket reference |
| `operator_approval_change_window_missing` | `approval_window` | 缺少 change window reference |
| `operator_approval_lifecycle_state_invalid` | `lifecycle` | lifecycle state 不在 allowlist |
| `operator_approval_diagnostic_exposure_detected` | `diagnostics` | diagnostics 暴露 forbidden fields |
| `operator_approval_runtime_created_forbidden` | `artifact_guard` | 本批创建 approval runtime 或 executor |
| `operator_approval_runtime_executed_forbidden` | `no_side_effects` | 本批执行 approval runtime |
| `operator_approval_side_effect_forbidden` | `no_side_effects` | 本批读取 secret、调用云服务、写 audit store 或连接 DB |
| `operator_approval_fallback_forbidden` | `no_fallback` | approval 缺失时 fallback 到 env、fake resolver、mock provider 或 fixture |
| `operator_approval_scope_overreach` | `implementation_boundary` | 本批合入 production resolver runtime、DB provider、audit store 或 public API |

所有失败必须 fail closed，不返回 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、完整 secret ref value、完整 credential handle、完整 operator claim、完整用户 claim、authorization header 或 cookie。

## Sanitized Diagnostics

允许输出：

- `operator_approval_boundary_status`
- `operator_approval_runtime_status`
- `operator_approval_execution_status`
- `approval_evidence_status`
- `approval_lifecycle_state`
- `approval_subject_status`
- `environment`
- `provider`
- `provider_profile`
- `secret_ref_key_status`
- `secret_ref_version_ref_status`
- `backend_profile_ref_status`
- `credential_handle_boundary_status`
- `no_secret_leakage_strategy_status`
- `operator_identity_status`
- `dual_control_status`
- `approval_ticket_status`
- `approval_window_status`
- `audit_policy_status`
- `rotation_policy_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、secret value、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、full secret ref value、full credential handle、完整 operator claim、完整用户 claim、authorization header 或 cookie。

## No Fallback

- operator approval runtime evidence boundary 缺失时真实 resolver runtime implementation task card 必须继续 blocked。
- 不允许 approval evidence fallback 到 operator runbook 文本、developer env plaintext、fixture credential、committed value、sample、mock provider、local-smoke profile、fake resolver runtime、repository memory store 或 previously approved test evidence。
- 不允许 production approval fallback 到 test approval，也不允许 test approval fallback 到 production approval。
- 不允许缺少 requester / approver / ticket / change window 时通过 `RADISHMIND_PLATFORM_API_KEY` 或其它环境变量 credential 生成 approval success。
- 不把 operator approval runtime evidence readiness 写成 approval runtime executed、credential resolved、production resolver runtime ready、production secret backend ready 或 public production API ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema、已有 readiness fixture 和 `check-repo.py` 注册顺序，不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 production resolver、不执行 approval runtime、不执行 smoke runtime、不创建 credential payload、不创建 credential handle、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不写 audit store、不调用 production API。

side effect counters 必须保持：

- `real_secret_read_count=0`
- `environment_secret_read_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `production_resolver_call_count=0`
- `operator_approval_runtime_created_count=0`
- `operator_approval_runtime_execution_count=0`
- `operator_identity_provider_call_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `credential_payload_created_count=0`
- `credential_handle_created_count=0`
- `credential_handle_runtime_created_count=0`
- `database_connection_count=0`
- `driver_open_count=0`
- `sql_execution_count=0`
- `schema_marker_read_count=0`
- `schema_marker_write_count=0`
- `repository_mode_enablement_count=0`
- `production_audit_store_write_count=0`
- `production_api_call_count=0`

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.md`
- `docs/task-cards/production-secret-backend-operator-approval-runtime-evidence-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py`

不得新增或启用以下 artifact：

- operator approval runtime
- operator approval executor
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
- production secret audit store
- audit writer
- public production API

## 后续推进

本批完成后，真实 resolver runtime implementation task card 仍不能在本切片创建。`operator-approval-runtime-evidence-readiness` 只固定 future operator approval runtime evidence 的边界和停止线，不执行 approval runtime。后续已由 `production-secret-backend-audit-store-handoff-readiness-v1` 固定 audit store handoff readiness，并由 `production-secret-backend-resolver-backend-health-boundary-readiness-v1` 固定 backend health boundary readiness；当前仍不创建 audit store、audit writer、audit event、backend health runtime 或 health check。下一步若继续 production secret backend，应单独评审 backend health runtime implementation entry 或 real resolver runtime implementation entry refresh。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
