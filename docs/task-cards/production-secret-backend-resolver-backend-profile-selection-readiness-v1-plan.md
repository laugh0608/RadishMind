# `Production Secret Backend Resolver Backend Profile Selection Readiness` v1 计划

更新时间：2026-06-20

## 任务目标

本批固定 production secret backend 真实 resolver runtime 之前的 resolver backend profile selection 前置证据。状态固定为 `resolver_backend_profile_selection_readiness_defined`。

本任务卡只定义 backend profile selection 的静态准入和停止线；不实现 production resolver runtime，不创建 production resolver runtime implementation task card，不读取真实 secret，不调用云 secret 服务，不创建 credential payload 或 credential handle runtime，不连接数据库，不启用 repository mode，不新增 public production API。

## 输入证据

- `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 已固定 `real_resolver_runtime_implementation_entry_review_defined`；本批完成后，entry review 中的 `resolver-backend-profile-selection-readiness` 已转为静态前置证据，runtime task card 仍被其它 blocker 阻塞。
- `production-secret-backend-real-resolver-runtime-preconditions-v1` 已固定 `real_resolver_runtime_preconditions_defined`。
- `production-secret-backend-provider-profile-secret-binding-readiness-v1` 已固定 reference-only provider/profile binding。
- `production-secret-backend-secret-resolver-interface-disabled-readiness-v1` 仍确认 production resolver interface 默认 disabled。
- `production-secret-backend-operator-runbook-negative-gates-readiness-v1` 与 `production-secret-backend-rotation-audit-policy-readiness-v1` 只提供策略前置，不提供 runtime audit store、rotation runtime 或 operator approval execution。
- `production-ops-secret-backend-implementation-readiness` 仍保持 `production_secret_backend=not_satisfied`、`resolver_runtime_status=not_created`。

## 本批交付

- 新增平台专题：`docs/platform/production-secret-backend-resolver-backend-profile-selection-readiness-v1.md`。
- 新增 fixture：`scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json`。
- 新增 checker：`scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py`。
- 更新 production secret backend implementation readiness 聚合 fixture / checker。
- 更新 `check-repo.py`、`scripts/README.md`、平台 / feature / task card 入口、当前焦点和周志。

## Selection Contract

backend profile selection 必须定义：

- `backend_profile_id`
- `backend_kind`
- `environment`
- `provider_profile_id`
- `policy_version`
- `secret_ref_namespace`
- `allowed_secret_ref_kinds`
- `operator_approval_required`
- `audit_policy_ref`
- `rotation_policy_ref`
- `health_boundary_status`

允许的 backend kind 只能是 reserved reference：

- `external_secret_manager_reserved`
- `operator_managed_secret_store_reserved`

禁止 backend kind：

- `committed_secret_file`
- `env_plaintext`
- `fake_resolver`
- `mock_provider`
- `local_smoke_profile`
- `repository_memory_store`
- `database_dsn_resolver`

## Failure Mapping

本批必须固定以下 fail-closed code：

- `resolver_backend_profile_selection_missing`
- `resolver_backend_profile_kind_unsupported`
- `resolver_backend_profile_id_missing`
- `resolver_backend_profile_environment_mismatch`
- `resolver_backend_profile_provider_mismatch`
- `resolver_backend_profile_secret_ref_namespace_missing`
- `resolver_backend_profile_policy_version_missing`
- `resolver_backend_profile_health_boundary_missing`
- `resolver_backend_profile_operator_approval_missing`
- `resolver_backend_profile_audit_policy_missing`
- `resolver_backend_profile_rotation_policy_missing`
- `resolver_backend_profile_secret_value_detected`
- `resolver_backend_profile_cloud_sdk_forbidden`
- `resolver_backend_profile_runtime_created_forbidden`
- `resolver_backend_profile_repository_mode_forbidden`
- `resolver_backend_profile_scope_overreach`

所有 failure diagnostics 必须脱敏，不输出 secret value、DSN、provider raw URL、cloud credential、credential payload、完整 secret ref、完整 credential handle、authorization header 或 cookie。

## No Fallback / No Side Effects

- backend profile 缺失、环境不匹配、provider profile 不匹配、policy version 缺失或 health boundary 缺失时必须 fail closed。
- 不允许 production profile fallback 到 test profile、fake resolver、mock provider、local-smoke profile、developer env plaintext、fixture credential、committed value、sample 或 repository memory store。
- 不允许调用 fake resolver、production resolver、cloud secret service、provider health check、DB driver、SQL、schema marker、audit store 或 public production API。
- side effect counters 必须全部为 `0`。

## Artifact Guard

本批不得新增：

- `services/platform/internal/secretbackend/resolver.go`
- `services/platform/internal/secretbackend/production_resolver.go`
- `services/platform/internal/secretbackend/cloud_secret_client.go`
- `services/platform/internal/secretbackend/backend_profile_runtime.go`
- `services/platform/internal/secretbackend/credential_payload.go`
- `services/platform/internal/secretbackend/credential_handle.go`
- `services/platform/internal/secretbackend/no_secret_leakage_smoke.go`
- `services/platform/internal/secretbackend/database_connection_provider.go`
- `services/platform/internal/secretbackend/audit_store.go`
- `docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-v1-plan.md`

## 验收方式

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

## 后续对齐

后续已由 `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1` 固定 no leakage smoke runtime strategy 静态证据，并由 `production-secret-backend-credential-handle-runtime-boundary-readiness-v1` 固定 credential handle boundary 静态证据。下一步若继续 production secret backend，必须在 operator approval runtime evidence、audit store handoff 或 backend health boundary 等剩余 blocker 中选择单一方向；不得把 backend profile selection readiness、no leakage strategy 或 credential handle boundary readiness 写成 backend runtime、smoke runtime、handle runtime 或 production resolver runtime ready。
