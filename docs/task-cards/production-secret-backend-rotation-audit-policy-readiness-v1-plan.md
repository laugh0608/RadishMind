# `Production Secret Backend Rotation / Audit Policy Readiness` v1 计划

更新时间：2026-06-19

## 任务目标

本任务卡用于把 `production-secret-backend-implementation-readiness` 中的 `rotation-and-audit-policy` 切片推进为可检查证据。

本批只固定 rotation trigger、approval / change window、secret ref versioning、rollback / disable policy、sanitized verification、audit event fields、failure mapping、no fallback、no side effects 和 artifact guard；不实现 resolver runtime，不创建 fake resolver，不执行 rotation，不写 audit store，不读取 secret value，不调用云 secret 服务，不访问 provider，不接数据库，不启用 workflow saved draft repository mode。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [平台专题入口](../platform/README.md)
- [Production Secret Backend Config / Secret Ref Readiness v1](../platform/production-secret-backend-config-secret-ref-readiness-v1.md)
- [Production Secret Backend Provider Profile Secret Binding Readiness v1](../platform/production-secret-backend-provider-profile-secret-binding-readiness-v1.md)
- [Production Secret Backend Secret Resolver Interface Disabled Readiness v1](../platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md)
- [Production Secret Backend Operator Runbook / Negative Gates Readiness v1](../platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md)
- [Production Secret Backend Rotation / Audit Policy Readiness v1](../platform/production-secret-backend-rotation-audit-policy-readiness-v1.md)
- [Production Secret Backend Implementation v1 计划](production-secret-backend-implementation-v1-plan.md)
- `contracts/production-secret-reference.schema.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`
- `scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json`
- `scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json`
- `scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json`
- `scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py`
- `scripts/check-production-ops-secret-backend-implementation-readiness.py`
- `services/platform/README.md`

## 当前事实

- `secret-ref-schema-and-fixtures` 已落地，并保持 reference-only。
- `config-secret-ref-readiness` 已落地，并只满足配置注入点前置条件。
- `provider-profile-secret-binding` 已落地，并只满足 provider/profile binding 前置条件。
- `secret-resolver-interface-disabled` 已落地，并只固定 disabled result 与 fail-closed failure mapping。
- `operator-runbook-and-negative-gates` 已落地，并固定人工启用、negative gates 和 production ready 停止线。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- `rotation-and-audit-policy` 已由本批推进为 `satisfied`。

## 本批输出

1. 新增平台专题，固定 rotation / audit policy 的语义和停止线。
2. 新增 fixture，声明：
   - rotation policy 只能保存 reference-only、policy version、secret ref version reference 和脱敏验证字段。
   - audit policy 只能保存脱敏字段、引用字段和 failure code。
   - approval、change window、rollback / disable、post-rotation smoke record 和 audit event field policy 是后续 production ready 前置。
   - `production_secret_backend`、resolver runtime、fake resolver、cloud secret service、audit store、DB provider 和 repository mode 仍未启用。
3. 新增 checker，校验：
   - implementation readiness 已把 `rotation-and-audit-policy` 标记为 satisfied，并引用本批证据。
   - `rotation-and-audit-policy` 前置条件有本批 evidence。
   - failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard 完整。
   - `check-repo.py` 注册顺序位于 operator runbook / negative gates readiness checker 之后。
4. 更新相关入口文档、周志和脚本说明。

## Failure Mapping

- `rotation_policy_missing`
- `rotation_trigger_missing`
- `rotation_approval_missing`
- `rotation_window_missing`
- `rotation_secret_ref_version_missing`
- `rotation_sanitized_verification_missing`
- `rotation_smoke_record_missing`
- `rotation_audit_event_missing`
- `rotation_audit_secret_exposure_detected`
- `rotation_fallback_forbidden`
- `rotation_production_ready_claim_forbidden`

这些错误必须归入 `rotation_audit_policy` failure boundary，并直接暴露为 fail-closed 诊断；不得回退到 developer env credential、mock provider、local-smoke profile、fixture credential、committed secret value、跨环境 `secret_ref`、fake resolver 或 fake query executor。

## 验收口径

- `scripts/checks/fixtures/production-secret-backend-rotation-audit-policy-readiness-v1.json` 存在并被 checker 消费。
- `production-ops-secret-backend-implementation-readiness.json` 中 `rotation-and-audit-policy` 状态为 `satisfied`。
- `rotation-and-audit-policy` 前置条件状态为 `satisfied`，且引用本批 evidence。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- audit event fields 不返回 credential handle，不保存 secret value。
- `production-secret-reference-basic` 仍为 reference-only，`resolver_enabled=false`、`cloud_calls_allowed=false`。
- 不新增 rotation runtime、production secret audit store、audit writer、resolver runtime、fake resolver、cloud SDK、DB provider、DB driver、connection factory、SQL、schema marker、migration runner、public production API 或 repository mode runtime。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-ops-config-secret-boundary.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

## 后续对齐

后续已由 `production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1`、`production-secret-backend-fake-resolver-runtime-implementation-v1`、`production-secret-backend-real-resolver-runtime-preconditions-v1`、`production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 和 `production-secret-backend-resolver-backend-profile-selection-readiness-v1` 固定 test-only fake resolver runtime、真实 resolver runtime 前置、blocked-before-task-card 结论与 backend profile selection 静态前置。下一步若继续 production secret backend，必须在 no leakage smoke runtime strategy、credential handle runtime boundary、operator approval runtime evidence、audit store handoff、backend health boundary 或其它单一 blocker 中开题。

## 停止线

- 不实现真实 secret backend。
- 不实现 secret resolver runtime。
- 不创建 fake resolver。
- 不执行 rotation。
- 不写 production secret audit store。
- 不调用云 secret API。
- 不读取、写入或提交真实 secret。
- 不把 `RADISHMIND_SECRET_SOURCE`、`.env.example`、developer env override、reference fixture、profile binding、disabled resolver interface、operator runbook、negative gate evidence、rotation policy、audit policy 或 readiness checker 写成 production secret backend。
- 不接 database connection provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode、production API、executor、confirmation、writeback 或 replay。
