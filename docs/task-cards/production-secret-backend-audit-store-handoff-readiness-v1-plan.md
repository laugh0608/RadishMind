# `Production Secret Backend Audit Store Handoff Readiness` v1 计划

更新时间：2026-06-20

## 任务目标

本批固定 production secret backend 真实 resolver runtime implementation task card 之前的 audit store handoff boundary。状态固定为 `audit_store_handoff_readiness_defined`。

本任务卡只定义 future audit store handoff 的 reference-only handoff envelope、允许 metadata、禁止 payload / secret material、event kind allowlist、secret ref / provider profile / environment / backend profile binding、credential handle boundary、operator approval evidence、rotation / retention / redaction policy、delivery semantics、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不创建 audit store，不创建 audit writer，不写 audit event，不连接数据库，不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不创建 credential payload 或 credential handle runtime，不启用 repository mode，不执行 no secret leakage smoke runtime，不新增 public production API。

## 输入证据

- `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 已固定 `real_resolver_runtime_implementation_entry_review_defined`，entry decision 为 blocked before runtime task card。
- `production-secret-backend-rotation-audit-policy-readiness-v1` 已固定 `rotation_audit_policy_readiness_defined`。
- `production-secret-backend-operator-approval-runtime-evidence-readiness-v1` 已固定 `operator_approval_runtime_evidence_readiness_defined`。
- `production-secret-backend-credential-handle-runtime-boundary-readiness-v1` 已固定 `credential_handle_runtime_boundary_readiness_defined`。
- `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1` 已固定 `real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`。
- `production-secret-backend-resolver-backend-profile-selection-readiness-v1` 已固定 `resolver_backend_profile_selection_readiness_defined`。
- `production-ops-secret-backend-implementation-readiness` 仍保持 `production_secret_backend=not_satisfied`、`resolver_runtime_status=not_created`。

## 本批交付

- 新增平台专题：`docs/platform/production-secret-backend-audit-store-handoff-readiness-v1.md`。
- 新增 fixture：`scripts/checks/fixtures/production-secret-backend-audit-store-handoff-readiness-v1.json`。
- 新增 checker：`scripts/check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py`。
- 更新 production secret backend implementation readiness 聚合 fixture / checker。
- 更新真实 resolver runtime implementation entry review，标记 audit store handoff 已定义但 audit store / writer 未创建、event 未写入。
- 更新 operator approval evidence / credential handle boundary / no leakage strategy alignment、`check-repo.py`、`scripts/README.md`、平台 / feature / task card 入口、当前焦点和周志。

## Boundary Contract

future audit store handoff boundary 必须固定：

- `audit_handoff_contract`
- `allowed_metadata`
- `forbidden_material`
- `required_bindings`
- `event_kind_allowlist`
- `handoff_lifecycle`
- `failure_mapping`
- `diagnostic_allowlist`
- `side_effect_counters`
- `artifact_guard`

允许的 metadata 字段只能是 reference-only 与审计字段：

- `audit_handoff_id`
- `audit_event_id`
- `event_kind`
- `environment`
- `provider`
- `provider_profile`
- `secret_ref_key_status`
- `secret_ref_version_ref`
- `backend_profile_ref`
- `credential_handle_boundary_ref`
- `operator_approval_evidence_ref`
- `approval_ticket_ref`
- `approval_window_ref`
- `request_id`
- `audit_ref`
- `runbook_version`
- `policy_version`
- `rotation_policy_version`
- `delivery_mode`
- `idempotency_key_ref`
- `retention_policy_ref`
- `redaction_profile_ref`
- `failure_code`
- `sanitized_diagnostic`
- `timestamp`

禁止 material：

- credential payload
- secret value
- password、token、API key、cloud credential
- provider raw URL、resolver backend URL、DSN、database hostname
- full secret ref value
- full credential handle
- full operator identity claim
- authorization header、cookie、完整用户 claim
- raw audit payload、raw request payload、raw response payload

## Event Kind Allowlist

本批只定义 future event kind，不创建 audit writer：

- `secret_resolution_requested`
- `secret_resolution_denied`
- `secret_resolution_failed_closed`
- `credential_handle_boundary_checked`
- `operator_approval_evidence_checked`
- `backend_profile_selected`
- `no_leakage_gate_checked`
- `rotation_policy_checked`
- `audit_handoff_failed_closed`

## Handoff Lifecycle

本批只定义 future lifecycle allowlist，不创建 runtime 状态机：

- `handoff_planned`
- `metadata_bound`
- `event_reference_prepared`
- `delivery_pending`
- `delivery_blocked_store_not_created`
- `delivered_metadata_only`
- `delivery_failed_closed`
- `redaction_required`
- `retention_policy_required`

## Failure Mapping

本批必须固定以下 fail-closed code：

- `audit_handoff_boundary_missing`
- `audit_handoff_envelope_missing`
- `audit_handoff_metadata_allowlist_missing`
- `audit_handoff_payload_detected`
- `audit_handoff_secret_material_detected`
- `audit_handoff_event_kind_invalid`
- `audit_handoff_secret_ref_binding_missing`
- `audit_handoff_provider_profile_binding_missing`
- `audit_handoff_environment_binding_mismatch`
- `audit_handoff_backend_profile_binding_missing`
- `audit_handoff_credential_boundary_missing`
- `audit_handoff_operator_approval_reference_missing`
- `audit_handoff_rotation_policy_missing`
- `audit_handoff_redaction_policy_missing`
- `audit_handoff_retention_policy_missing`
- `audit_handoff_delivery_semantics_missing`
- `audit_handoff_idempotency_missing`
- `audit_handoff_store_created_forbidden`
- `audit_handoff_writer_created_forbidden`
- `audit_handoff_write_executed_forbidden`
- `audit_handoff_side_effect_forbidden`
- `audit_handoff_fallback_forbidden`
- `audit_handoff_scope_overreach`

所有 failure diagnostics 必须脱敏，不输出 secret value、DSN、provider raw URL、cloud credential、credential payload、完整 secret ref、完整 credential handle、完整 operator claim、authorization header、cookie、raw audit payload、raw request payload 或 raw response payload。

## No Fallback / No Side Effects

- audit handoff boundary 缺失时真实 resolver runtime implementation task card 必须 fail closed。
- 不允许 fallback 到 runbook 文本、developer env plaintext、fixture credential、committed value、sample、mock provider、local-smoke profile、fake resolver runtime、repository memory store、console local record 或 previously approved test evidence。
- 不允许调用 fake resolver、production resolver、cloud secret service、provider、DB driver、SQL、schema marker、audit store、audit writer、identity provider 或 public production API。
- 不允许创建 audit store、audit writer、audit repository、audit delivery runtime、credential payload、credential handle runtime、production resolver runtime 或 no leakage smoke runtime。
- side effect counters 必须全部为 `0`。

## Artifact Guard

本批不得新增：

- `services/platform/internal/secretbackend/audit_store.go`
- `services/platform/internal/secretbackend/audit_writer.go`
- `services/platform/internal/secretbackend/audit_repository.go`
- `services/platform/internal/secretbackend/audit_delivery.go`
- `services/platform/internal/secretbackend/audit_retention.go`
- `services/platform/internal/secretbackend/real_resolver.go`
- `services/platform/internal/secretbackend/production_resolver.go`
- `services/platform/internal/secretbackend/cloud_secret_client.go`
- `services/platform/internal/secretbackend/credential_payload.go`
- `services/platform/internal/secretbackend/credential_handle.go`
- `services/platform/internal/secretbackend/no_secret_leakage_smoke.go`
- `services/platform/internal/secretbackend/database_connection_provider.go`
- `docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-v1-plan.md`

## 验收方式

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
