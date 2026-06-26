# `Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Strategy` v1 计划

更新时间：2026-06-20

## 任务目标

本批固定 production secret backend 真实 resolver runtime implementation task card 之前的 no-secret-leakage smoke runtime strategy。状态固定为 `real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`。

本任务卡只定义 future smoke runtime gate 的离线策略、扫描面、输入 / 输出 allowlist、负向 probe、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不创建 no secret leakage smoke runtime，不执行 resolver smoke，不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不创建 credential payload 或 credential handle runtime，不连接数据库，不启用 repository mode，不新增 public production API。

## 输入证据

- `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 已固定 `real_resolver_runtime_implementation_entry_review_defined`，entry decision 为 blocked before runtime task card。
- `production-secret-backend-real-resolver-runtime-preconditions-v1` 已固定 `real_resolver_runtime_preconditions_defined`。
- `production-secret-backend-resolver-backend-profile-selection-readiness-v1` 已固定 `resolver_backend_profile_selection_readiness_defined`。
- `production-secret-backend-fake-resolver-runtime-implementation-v1` 已实现 test-only fake resolver runtime；它不能替代 production resolver no leakage gate。
- `production-ops-secret-backend-implementation-readiness` 仍保持 `production_secret_backend=not_satisfied`、`resolver_runtime_status=not_created`。

## 本批交付

- 新增平台专题：`docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.md`。
- 新增 fixture：`scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.json`。
- 新增 checker：`scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py`。
- 更新 production secret backend implementation readiness 聚合 fixture / checker。
- 更新真实 resolver runtime implementation entry review，标记 no leakage strategy 已定义但 runtime 未创建。
- 更新 `check-repo.py`、`scripts/README.md`、平台 / feature / task card 入口、当前焦点和周志。

## Strategy Contract

future no leakage smoke runtime gate 必须固定：

- `scan_surfaces`
- `input_allowlist`
- `success_output_allowlist`
- `failure_output_allowlist`
- `audit_metadata_allowlist`
- `diagnostic_allowlist`
- `probe_categories`
- `forbidden_output_classes`
- `side_effect_counters`
- `artifact_guard`

允许的 input 字段只能是 reference-only 与审计字段：

- `environment`
- `provider`
- `provider_profile`
- `credential_requirement`
- `secret_ref_key`
- `secret_ref_version_ref`
- `caller_purpose`
- `request_id`
- `audit_ref`
- `operator_approval_ref`
- `policy_version`
- `rotation_policy_version`
- `backend_profile_ref`

允许的 success output 只能是 sanitized metadata：

- `resolver_state`
- `credential_handle_present`
- `credential_handle_id`
- `credential_kind`
- `environment`
- `provider`
- `provider_profile`
- `secret_ref_version_ref`
- `request_id`
- `audit_ref`
- `policy_version`

允许的 failure output 只能是 fail-closed metadata：

- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

## Failure Mapping

本批必须固定以下 fail-closed code：

- `real_resolver_no_leakage_strategy_missing`
- `real_resolver_no_leakage_input_allowlist_missing`
- `real_resolver_no_leakage_output_allowlist_missing`
- `real_resolver_no_leakage_probe_matrix_missing`
- `real_resolver_no_leakage_scan_surface_missing`
- `real_resolver_no_leakage_secret_value_detected`
- `real_resolver_no_leakage_credential_payload_detected`
- `real_resolver_no_leakage_diagnostic_exposure_detected`
- `real_resolver_no_leakage_audit_exposure_detected`
- `real_resolver_no_leakage_runtime_created_forbidden`
- `real_resolver_no_leakage_cloud_call_forbidden`
- `real_resolver_no_leakage_fake_resolver_substitution_forbidden`
- `real_resolver_no_leakage_repository_mode_forbidden`
- `real_resolver_no_leakage_scope_overreach`

所有 failure diagnostics 必须脱敏，不输出 secret value、DSN、provider raw URL、cloud credential、credential payload、完整 secret ref、完整 credential handle、authorization header 或 cookie。

## No Fallback / No Side Effects

- no leakage strategy 缺失时真实 resolver runtime implementation task card 必须 fail closed。
- 不允许 production resolver no leakage gate fallback 到 fake resolver runtime、mock provider、local-smoke profile、developer env plaintext、fixture credential、committed value、sample 或 repository memory store。
- 不允许执行 smoke runtime、调用 fake resolver、production resolver、cloud secret service、provider、DB driver、SQL、schema marker、audit store 或 public production API。
- side effect counters 必须全部为 `0`。

## Artifact Guard

本批不得新增：

- `services/platform/internal/secretbackend/real_resolver.go`
- `services/platform/internal/secretbackend/production_resolver.go`
- `services/platform/internal/secretbackend/cloud_secret_client.go`
- `services/platform/internal/secretbackend/no_secret_leakage_smoke.go`
- `services/platform/internal/secretbackend/credential_payload.go`
- `services/platform/internal/secretbackend/credential_handle.go`
- `services/platform/internal/secretbackend/database_connection_provider.go`
- `services/platform/internal/secretbackend/audit_store.go`
- `docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-v1-plan.md`

## 验收方式

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
