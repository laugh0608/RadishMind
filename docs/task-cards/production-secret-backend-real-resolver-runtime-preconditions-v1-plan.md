# Production Secret Backend Real Resolver Runtime Preconditions v1 计划

更新时间：2026-06-20

## 任务目标

本任务卡固定真实 resolver runtime implementation 之前必须满足的前置条件、失败语义、诊断字段、负向停止线和验证方式。它承接 test-only fake resolver runtime 完成后的下一步，但不实现 production resolver。

状态：`real_resolver_runtime_preconditions_defined`

## 输入事实

- [Production Secret Backend Fake Resolver Runtime Implementation v1](../platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md)
- [Production Secret Backend Secret Resolver Interface Disabled Readiness v1](../platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md)
- [Production Secret Backend Provider Profile Secret Binding Readiness v1](../platform/production-secret-backend-provider-profile-secret-binding-readiness-v1.md)
- [Production Secret Backend Operator Runbook / Negative Gates Readiness v1](../platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md)
- [Production Secret Backend Rotation / Audit Policy Readiness v1](../platform/production-secret-backend-rotation-audit-policy-readiness-v1.md)
- `contracts/production-secret-reference.schema.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`

## 本批范围

本批新增：

- 平台专题：`docs/platform/production-secret-backend-real-resolver-runtime-preconditions-v1.md`
- 任务卡：`docs/task-cards/production-secret-backend-real-resolver-runtime-preconditions-v1-plan.md`
- fixture：`scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-preconditions-v1.json`
- checker：`scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py`
- `check-repo.py` 注册和入口文档同步

本批只定义：

- 真实 resolver runtime 启用条件
- secret ref / provider profile / environment binding 前置
- operator approval、audit policy、rotation policy 依赖
- future resolver input / result allowed fields
- failure mapping
- sanitized diagnostics
- no fallback
- no side effects
- artifact guard
- 验证方式和后续实现拆分

## 不做范围

- 不读取真实 secret。
- 不调用云 secret 服务。
- 不选择云厂商 SDK。
- 不连接数据库。
- 不创建 credential payload。
- 不创建 credential handle runtime。
- 不实现 production resolver runtime。
- 不创建 no secret leakage smoke runtime。
- 不启用 workflow saved draft repository mode。
- 不新增 public production API。
- 不把 test-only fake resolver runtime 写成 production resolver。

## 启用前置

真实 resolver runtime 后续实现任务必须先满足：

1. 独立 implementation entry review 明确只打开 resolver runtime。
2. reference-only `secret_ref` manifest / config layer 已可验证。
3. provider/profile binding 已固定 credential requirement、secret ref status 和 environment binding。
4. operator approval evidence、runbook version、policy version 和 change window reference 已存在。
5. audit / rotation policy 已固定 secret ref version reference、rollback / disable policy 和 sanitized verification。
6. resolver backend profile 已在独立任务中选择，并明确 allowed environment、backend kind 和 policy version。
7. no secret leakage smoke runtime 的实现计划已单独固定。
8. artifact guard 证明未提前出现 production resolver、cloud client、DB provider、repository mode、audit store 或 public API artifact。

## Failure Mapping

必须覆盖以下 failure code：

- `real_resolver_runtime_preconditions_missing`
- `real_resolver_runtime_task_not_approved`
- `real_resolver_backend_profile_missing`
- `real_resolver_secret_ref_missing`
- `real_resolver_provider_profile_binding_missing`
- `real_resolver_environment_binding_missing`
- `real_resolver_operator_approval_missing`
- `real_resolver_audit_policy_missing`
- `real_resolver_rotation_policy_missing`
- `real_resolver_no_leakage_gate_missing`
- `real_resolver_backend_unavailable`
- `real_resolver_resolution_denied`
- `real_resolver_secret_value_exposure_detected`
- `real_resolver_fallback_forbidden`
- `real_resolver_side_effect_forbidden`
- `real_resolver_artifact_guard_violation`

所有失败必须 fail closed，并输出 sanitized diagnostic。

## Sanitized Diagnostics

允许输出：

- `real_resolver_runtime_preconditions_status`
- `resolver_state`
- `resolver_backend_profile_status`
- `secret_ref_status`
- `secret_ref_version_status`
- `provider_profile_binding_status`
- `environment_binding_status`
- `operator_approval_status`
- `audit_policy_status`
- `rotation_policy_status`
- `no_secret_leakage_gate_status`
- `credential_handle_present`
- `failure_code`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、full secret ref value、full credential handle、完整用户 claim、authorization header 或 cookie。

## Artifact Guard

必须确认不存在：

- `services/platform/internal/secretbackend/resolver.go`
- `services/platform/internal/secretbackend/production_resolver.go`
- `services/platform/internal/secretbackend/cloud_secret_client.go`
- `services/platform/internal/secretbackend/credential_payload.go`
- `services/platform/internal/secretbackend/credential_handle.go`
- `services/platform/internal/secretbackend/no_secret_leakage_smoke.go`
- `services/platform/internal/secretbackend/database_connection_provider.go`
- `services/platform/internal/secretbackend/schema_marker.go`
- `services/platform/internal/secretbackend/audit_store.go`

源码扫描不得出现 `ResolveProductionSecret`、`NewProductionSecretResolver`、`CloudSecretClient`、`database/sql`、`RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 或 secret-looking literal。

## 验收

- 新平台专题、task card、fixture 和 checker 均存在。
- checker 校验 dependencies、enablement gates、failure mapping、sanitized diagnostics、no fallback、no side effects、artifact guard、implementation readiness alignment 和 check-repo 注册顺序。
- `production_secret_backend` 保持 `not_satisfied`。
- `resolver_runtime_status` 保持 `not_created`。
- `fake_resolver_runtime_status` 只能解释为 `implemented_test_only_disabled_by_default`。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
