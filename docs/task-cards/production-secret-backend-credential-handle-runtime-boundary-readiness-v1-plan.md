# `Production Secret Backend Credential Handle Runtime Boundary Readiness` v1 计划

更新时间：2026-06-20

## 任务目标

本批固定 production secret backend 真实 resolver runtime implementation task card 之前的 credential handle runtime boundary。状态固定为 `credential_handle_runtime_boundary_readiness_defined`。

本任务卡只定义 future opaque credential handle 的 reference 定义、允许 metadata、禁止 payload / secret material、secret ref / provider profile / environment binding 前置、operator approval / audit / rotation 依赖、handle lifecycle、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不创建 credential handle runtime，不创建 credential payload，不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不连接数据库，不启用 repository mode，不创建 no secret leakage smoke runtime，不新增 public production API。

## 输入证据

- `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 已固定 `real_resolver_runtime_implementation_entry_review_defined`，entry decision 为 blocked before runtime task card。
- `production-secret-backend-real-resolver-runtime-preconditions-v1` 已固定 `real_resolver_runtime_preconditions_defined`。
- `production-secret-backend-resolver-backend-profile-selection-readiness-v1` 已固定 `resolver_backend_profile_selection_readiness_defined`。
- `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1` 已固定 `real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`。
- `production-secret-backend-fake-resolver-runtime-implementation-v1` 只服务 test-only fake resolver runtime。
- `production-ops-secret-backend-implementation-readiness` 仍保持 `production_secret_backend=not_satisfied`、`resolver_runtime_status=not_created`。

## 本批交付

- 新增平台专题：`docs/platform/production-secret-backend-credential-handle-runtime-boundary-readiness-v1.md`。
- 新增 fixture：`scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-boundary-readiness-v1.json`。
- 新增 checker：`scripts/check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py`。
- 更新 production secret backend implementation readiness 聚合 fixture / checker。
- 更新真实 resolver runtime implementation entry review，标记 credential handle boundary 已定义但 runtime 未创建。
- 更新 no leakage strategy alignment、`check-repo.py`、`scripts/README.md`、平台 / feature / task card 入口、当前焦点和周志。

## Boundary Contract

future credential handle boundary 必须固定：

- `opaque_reference_definition`
- `allowed_metadata`
- `forbidden_material`
- `required_bindings`
- `lifecycle_states`
- `failure_mapping`
- `diagnostic_allowlist`
- `side_effect_counters`
- `artifact_guard`

允许的 metadata 字段只能是 reference-only 与审计字段：

- `credential_handle_id`
- `credential_handle_kind`
- `credential_handle_status`
- `credential_handle_lifecycle_state`
- `environment`
- `provider`
- `provider_profile`
- `secret_ref_key`
- `secret_ref_version_ref`
- `backend_profile_ref`
- `operator_approval_ref`
- `audit_ref`
- `policy_version`
- `rotation_policy_version`
- `created_at`
- `expires_at`
- `failure_code`
- `sanitized_diagnostic`

禁止 material：

- credential payload
- secret value
- password、token、API key、cloud credential
- provider raw URL、resolver backend URL、DSN、database hostname
- full secret ref value
- full credential handle payload
- authorization header、cookie、完整用户 claim

## Lifecycle States

本批只定义 future lifecycle allowlist，不创建 runtime 状态机：

- `reference_planned`
- `metadata_bound`
- `issuance_blocked_pending_operator_approval`
- `future_issued_metadata_only`
- `rotation_pending_rebind`
- `revoked`
- `expired`
- `resolution_failed_closed`

## Failure Mapping

本批必须固定以下 fail-closed code：

- `credential_handle_boundary_missing`
- `credential_handle_opaque_reference_missing`
- `credential_handle_metadata_allowlist_missing`
- `credential_handle_payload_detected`
- `credential_handle_secret_material_detected`
- `credential_handle_secret_ref_binding_missing`
- `credential_handle_provider_profile_binding_missing`
- `credential_handle_environment_binding_mismatch`
- `credential_handle_operator_approval_missing`
- `credential_handle_audit_dependency_missing`
- `credential_handle_rotation_dependency_missing`
- `credential_handle_lifecycle_state_invalid`
- `credential_handle_failure_mapping_missing`
- `credential_handle_diagnostic_exposure_detected`
- `credential_handle_runtime_created_forbidden`
- `credential_handle_side_effect_forbidden`
- `credential_handle_fallback_forbidden`
- `credential_handle_scope_overreach`

所有 failure diagnostics 必须脱敏，不输出 secret value、DSN、provider raw URL、cloud credential、credential payload、完整 secret ref、完整 credential handle、authorization header 或 cookie。

## No Fallback / No Side Effects

- credential handle boundary 缺失时真实 resolver runtime implementation task card 必须 fail closed。
- 不允许 fallback 到 credential payload、secret value、fake resolver runtime、mock provider、local-smoke profile、developer env plaintext、fixture credential、committed value、sample 或 repository memory store。
- 不允许调用 fake resolver、production resolver、cloud secret service、provider、DB driver、SQL、schema marker、audit store 或 public production API。
- 不允许创建 credential payload、credential handle runtime、production resolver runtime 或 no leakage smoke runtime。
- side effect counters 必须全部为 `0`。

## Artifact Guard

本批不得新增：

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
./scripts/run-python.sh scripts/check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
