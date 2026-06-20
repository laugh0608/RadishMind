# `Production Secret Backend Resolver Backend Health Boundary Readiness` v1 计划

更新时间：2026-06-20

## 任务目标

本批固定 production secret backend 真实 resolver runtime implementation task card 之前的 resolver backend health boundary。状态固定为 `resolver_backend_health_boundary_readiness_defined`。

本任务卡只定义 future resolver backend health boundary 的 reference-only 输入、profile binding、environment binding、允许 health metadata、禁止 secret material / credential payload / provider raw URL / DSN、health lifecycle、failure mapping、sanitized diagnostics、operator / audit / rotation / no leakage / credential handle 依赖、no fallback、no side effects、artifact guard、entry review alignment、implementation readiness alignment 和后续 runtime implementation 拆分；不创建 backend health runtime，不执行 backend health check，不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不连接数据库，不创建 credential payload，不创建 credential handle runtime，不执行 approval runtime，不创建 audit store / writer / event，不启用 repository mode，不新增 public production API。

## 输入证据

- `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 已固定 `real_resolver_runtime_implementation_entry_review_defined`，runtime task card 不在当前切片创建。
- `production-secret-backend-resolver-backend-profile-selection-readiness-v1` 已固定 `resolver_backend_profile_selection_readiness_defined`，并要求 backend health reference。
- `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1` 已固定 `real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`。
- `production-secret-backend-credential-handle-runtime-boundary-readiness-v1` 已固定 `credential_handle_runtime_boundary_readiness_defined`。
- `production-secret-backend-operator-approval-runtime-evidence-readiness-v1` 已固定 `operator_approval_runtime_evidence_readiness_defined`。
- `production-secret-backend-audit-store-handoff-readiness-v1` 已固定 `audit_store_handoff_readiness_defined`。
- `production-ops-secret-backend-implementation-readiness` 仍保持 `production_secret_backend=not_satisfied`、`resolver_runtime_status=not_created`。

## 本批交付

- 新增平台专题：`docs/platform/production-secret-backend-resolver-backend-health-boundary-readiness-v1.md`。
- 新增 fixture：`scripts/checks/fixtures/production-secret-backend-resolver-backend-health-boundary-readiness-v1.json`。
- 新增 checker：`scripts/check-production-ops-secret-backend-resolver-backend-health-boundary-readiness-v1.py`。
- 更新 production secret backend implementation readiness 聚合 fixture / checker。
- 更新真实 resolver runtime implementation entry review，标记 backend health boundary 已定义但 backend health runtime 未创建、health check 未执行。
- 更新 no leakage / credential handle / operator approval / audit handoff alignment、`check-repo.py`、`scripts/README.md`、平台 / feature / task card 入口、当前焦点和周志。

## Boundary Contract

future backend health boundary 必须固定：

- `health_boundary_contract`
- `required_inputs`
- `allowed_metadata`
- `forbidden_material`
- `required_bindings`
- `health_lifecycle`
- `health_status_allowlist`
- `failure_mapping`
- `diagnostic_allowlist`
- `side_effect_counters`
- `artifact_guard`

允许的 input / metadata 字段只能是 reference-only 与脱敏健康状态：

- `backend_health_ref`
- `backend_profile_ref`
- `backend_profile_id`
- `backend_kind`
- `environment`
- `provider_profile_id`
- `secret_ref_key_status`
- `secret_ref_version_ref_status`
- `health_policy_version`
- `health_lifecycle_state`
- `last_known_health_ref`
- `sanitized_backend_diagnostic`
- `failure_code`
- `failure_boundary`
- `request_id`
- `audit_ref`
- `policy_version`

禁止 material：

- secret material、secret value、password、token、API key、cloud credential
- credential payload、full credential handle
- provider raw URL、resolver backend URL、backend endpoint URL、DSN、database hostname
- full secret ref value、full operator identity claim、authorization header、cookie、完整用户 claim
- raw health request、raw health response、raw provider error、raw backend error detail

## Health Lifecycle

本批只定义 future lifecycle allowlist，不创建 runtime 状态机：

- `boundary_planned`
- `metadata_bound`
- `health_reference_prepared`
- `runtime_not_created`
- `health_check_not_executed`
- `health_status_unknown`
- `health_status_available_metadata_only`
- `health_status_degraded_metadata_only`
- `health_status_unavailable_fail_closed`
- `diagnostics_sanitized`
- `runtime_implementation_deferred`

## Failure Mapping

本批必须固定以下 fail-closed code：

- `resolver_backend_health_boundary_missing`
- `resolver_backend_health_reference_missing`
- `resolver_backend_health_profile_binding_missing`
- `resolver_backend_health_environment_mismatch`
- `resolver_backend_health_provider_profile_mismatch`
- `resolver_backend_health_secret_ref_binding_missing`
- `resolver_backend_health_policy_missing`
- `resolver_backend_health_metadata_forbidden`
- `resolver_backend_health_secret_material_detected`
- `resolver_backend_health_raw_endpoint_detected`
- `resolver_backend_health_runtime_created_forbidden`
- `resolver_backend_health_check_executed_forbidden`
- `resolver_backend_health_fallback_forbidden`
- `resolver_backend_health_side_effect_forbidden`
- `resolver_backend_health_repository_mode_forbidden`
- `resolver_backend_health_scope_overreach`

所有 failure diagnostics 必须脱敏，不输出 secret value、DSN、provider raw URL、resolver backend URL、backend endpoint URL、cloud credential、credential payload、完整 secret ref、完整 credential handle、完整 operator claim、authorization header、cookie、raw health request、raw health response 或 raw backend error detail。

## No Fallback / No Side Effects

- backend health boundary 缺失时不能用 fake resolver、developer env plaintext、fixture credential、committed value、sample、mock provider、local-smoke profile、operator runbook 文本、previously approved test evidence 或 repository memory store 替代。
- 不允许 production health boundary fallback 到 test health boundary，也不允许 test health boundary fallback 到 production health boundary。
- 不允许调用 fake resolver、production resolver、cloud secret service、provider、backend health check、DB driver、SQL、schema marker、audit store、audit writer、identity provider 或 public production API。
- 不允许创建 backend health runtime、backend health client、production resolver runtime、credential payload、credential handle runtime、approval runtime、audit store、audit writer、no leakage smoke runtime、DB provider 或 repository mode runtime。
- side effect counters 必须全部为 `0`。

## Artifact Guard

本批不得新增：

- `services/platform/internal/secretbackend/backend_health.go`
- `services/platform/internal/secretbackend/backend_health_runtime.go`
- `services/platform/internal/secretbackend/backend_health_client.go`
- `services/platform/internal/secretbackend/health_checker.go`
- `services/platform/internal/secretbackend/real_resolver.go`
- `services/platform/internal/secretbackend/production_resolver.go`
- `services/platform/internal/secretbackend/cloud_secret_client.go`
- `services/platform/internal/secretbackend/credential_payload.go`
- `services/platform/internal/secretbackend/credential_handle.go`
- `services/platform/internal/secretbackend/operator_approval_runtime.go`
- `services/platform/internal/secretbackend/audit_store.go`
- `services/platform/internal/secretbackend/audit_writer.go`
- `services/platform/internal/secretbackend/no_secret_leakage_smoke.go`
- `services/platform/internal/secretbackend/database_connection_provider.go`
- `docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-v1-plan.md`

## 后续拆分

本批完成后，后续 runtime implementation 仍必须单独开题：

1. `resolver-backend-health-runtime-implementation-entry-review`
2. `resolver-backend-health-runtime-implementation`
3. `production-secret-backend-real-resolver-runtime-implementation-entry-refresh`
4. `production-secret-backend-real-resolver-runtime-implementation`

上述目标都不得绕过 no leakage、credential handle、operator approval、audit handoff、rotation policy、health boundary、no fallback 和 no side effects 证据链。

## 验收方式

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-resolver-backend-health-boundary-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
