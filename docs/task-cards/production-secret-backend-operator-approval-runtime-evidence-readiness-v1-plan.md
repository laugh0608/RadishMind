# `Production Secret Backend Operator Approval Runtime Evidence Readiness` v1 计划

更新时间：2026-06-20

## 任务目标

本批固定 production secret backend 真实 resolver runtime implementation task card 之前的 operator approval runtime evidence boundary。状态固定为 `operator_approval_runtime_evidence_readiness_defined`。

本任务卡只定义 future operator approval runtime evidence 的 evidence shape、允许 metadata、禁止 payload / secret material、approval subject binding、secret ref / provider profile / environment / backend profile binding、credential handle boundary、no leakage strategy、operator identity、dual control、ticket / change window、audit / rotation dependency、lifecycle、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不创建 operator approval runtime，不执行 approval runtime，不写 audit store，不创建 credential handle runtime，不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不连接数据库，不启用 repository mode，不执行 no secret leakage smoke runtime，不新增 public production API。

## 输入证据

- `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 已固定 `real_resolver_runtime_implementation_entry_review_defined`，entry decision 为 blocked before runtime task card。
- `production-secret-backend-operator-runbook-negative-gates-readiness-v1` 已固定 `operator_runbook_negative_gates_readiness_defined`。
- `production-secret-backend-rotation-audit-policy-readiness-v1` 已固定 `rotation_audit_policy_readiness_defined`。
- `production-secret-backend-resolver-backend-profile-selection-readiness-v1` 已固定 `resolver_backend_profile_selection_readiness_defined`。
- `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1` 已固定 `real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`。
- `production-secret-backend-credential-handle-runtime-boundary-readiness-v1` 已固定 `credential_handle_runtime_boundary_readiness_defined`。
- `production-ops-secret-backend-implementation-readiness` 仍保持 `production_secret_backend=not_satisfied`、`resolver_runtime_status=not_created`。

## 本批交付

- 新增平台专题：`docs/platform/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.md`。
- 新增 fixture：`scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.json`。
- 新增 checker：`scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py`。
- 更新 production secret backend implementation readiness 聚合 fixture / checker。
- 更新真实 resolver runtime implementation entry review，标记 operator approval runtime evidence boundary 已定义但 runtime 未执行。
- 更新 no leakage strategy / credential handle boundary alignment、`check-repo.py`、`scripts/README.md`、平台 / feature / task card 入口、当前焦点和周志。

## Boundary Contract

future operator approval runtime evidence boundary 必须固定：

- `approval_evidence_contract`
- `allowed_metadata`
- `forbidden_material`
- `required_bindings`
- `lifecycle_states`
- `failure_mapping`
- `diagnostic_allowlist`
- `side_effect_counters`
- `artifact_guard`

允许的 metadata 字段只能是 reference-only 与审计字段：

- `approval_evidence_id`
- `approval_evidence_kind`
- `approval_status`
- `approval_lifecycle_state`
- `approval_subject`
- `environment`
- `provider`
- `provider_profile`
- `secret_ref_key`
- `secret_ref_version_ref`
- `backend_profile_ref`
- `credential_handle_boundary_ref`
- `requested_by_ref`
- `approved_by_ref`
- `approval_ticket_ref`
- `approval_window_ref`
- `runbook_version`
- `policy_version`
- `rotation_policy_version`
- `expires_at`
- `revoked_at`
- `request_id`
- `audit_ref`
- `failure_code`
- `sanitized_diagnostic`

禁止 material：

- credential payload
- secret value
- password、token、API key、cloud credential
- provider raw URL、resolver backend URL、DSN、database hostname
- full secret ref value
- full credential handle
- full operator identity claim
- authorization header、cookie、完整用户 claim

## Lifecycle States

本批只定义 future lifecycle allowlist，不创建 runtime 状态机：

- `evidence_planned`
- `request_metadata_bound`
- `approval_pending`
- `approved_metadata_only`
- `denied`
- `expired`
- `revoked`
- `revalidation_required`
- `execution_blocked`
- `resolution_failed_closed`

## Failure Mapping

本批必须固定以下 fail-closed code：

- `operator_approval_runtime_evidence_missing`
- `operator_approval_evidence_schema_missing`
- `operator_approval_metadata_allowlist_missing`
- `operator_approval_payload_detected`
- `operator_approval_secret_material_detected`
- `operator_approval_subject_binding_missing`
- `operator_approval_runbook_binding_missing`
- `operator_approval_secret_ref_binding_missing`
- `operator_approval_provider_profile_binding_missing`
- `operator_approval_environment_binding_mismatch`
- `operator_approval_backend_profile_binding_missing`
- `operator_approval_credential_boundary_missing`
- `operator_approval_no_leakage_strategy_missing`
- `operator_approval_audit_dependency_missing`
- `operator_approval_rotation_dependency_missing`
- `operator_approval_identity_reference_missing`
- `operator_approval_dual_control_missing`
- `operator_approval_ticket_missing`
- `operator_approval_change_window_missing`
- `operator_approval_lifecycle_state_invalid`
- `operator_approval_diagnostic_exposure_detected`
- `operator_approval_runtime_created_forbidden`
- `operator_approval_runtime_executed_forbidden`
- `operator_approval_side_effect_forbidden`
- `operator_approval_fallback_forbidden`
- `operator_approval_scope_overreach`

所有 failure diagnostics 必须脱敏，不输出 secret value、DSN、provider raw URL、cloud credential、credential payload、完整 secret ref、完整 credential handle、完整 operator claim、authorization header 或 cookie。

## No Fallback / No Side Effects

- operator approval runtime evidence boundary 缺失时真实 resolver runtime implementation task card 必须 fail closed。
- 不允许 fallback 到 runbook 文本、developer env plaintext、fixture credential、committed value、sample、mock provider、local-smoke profile、fake resolver runtime、repository memory store 或 previously approved test evidence。
- 不允许调用 fake resolver、production resolver、cloud secret service、provider、DB driver、SQL、schema marker、audit store、identity provider 或 public production API。
- 不允许创建 approval runtime、approval executor、credential payload、credential handle runtime、production resolver runtime 或 no leakage smoke runtime。
- side effect counters 必须全部为 `0`。

## Artifact Guard

本批不得新增：

- `services/platform/internal/secretbackend/operator_approval.go`
- `services/platform/internal/secretbackend/operator_approval_runtime.go`
- `services/platform/internal/secretbackend/approval_executor.go`
- `services/platform/internal/secretbackend/real_resolver.go`
- `services/platform/internal/secretbackend/production_resolver.go`
- `services/platform/internal/secretbackend/cloud_secret_client.go`
- `services/platform/internal/secretbackend/credential_payload.go`
- `services/platform/internal/secretbackend/credential_handle.go`
- `services/platform/internal/secretbackend/no_secret_leakage_smoke.go`
- `services/platform/internal/secretbackend/database_connection_provider.go`
- `services/platform/internal/secretbackend/audit_store.go`
- `docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-v1-plan.md`

## 验收方式

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
