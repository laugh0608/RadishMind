# `Production Secret Backend Operator Runbook / Negative Gates Readiness` v1 计划

更新时间：2026-06-19

## 任务目标

本任务卡用于把 `production-secret-backend-implementation-readiness` 中的 `operator-runbook-and-negative-gates` 切片推进为可检查证据。

本批只固定 operator runbook、test / production secret source 要求、脱敏验证、smoke record、negative gates、failure mapping、no fallback、no side effects 和 artifact guard；不实现 resolver runtime，不创建 fake resolver，不读取 secret value，不调用云 secret 服务，不访问 provider，不接数据库，不启用 workflow saved draft repository mode。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [平台专题入口](../platform/README.md)
- [Production Secret Backend Config / Secret Ref Readiness v1](../platform/production-secret-backend-config-secret-ref-readiness-v1.md)
- [Production Secret Backend Provider Profile Secret Binding Readiness v1](../platform/production-secret-backend-provider-profile-secret-binding-readiness-v1.md)
- [Production Secret Backend Secret Resolver Interface Disabled Readiness v1](../platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md)
- [Production Secret Backend Operator Runbook / Negative Gates Readiness v1](../platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md)
- [Production Secret Backend Implementation v1 计划](production-secret-backend-implementation-v1-plan.md)
- `contracts/production-secret-reference.schema.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`
- `scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json`
- `scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json`
- `scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py`
- `scripts/check-production-ops-secret-backend-implementation-readiness.py`
- `services/platform/README.md`

## 当前事实

- `secret-ref-schema-and-fixtures` 已落地，并保持 reference-only。
- `config-secret-ref-readiness` 已落地，并只满足配置注入点前置条件。
- `provider-profile-secret-binding` 已落地，并只满足 provider/profile binding 前置条件。
- `secret-resolver-interface-disabled` 已落地，并只固定 disabled result 与 fail-closed failure mapping。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- `operator-runbook-and-negative-gates` 已由本批推进为 `satisfied`。

## 本批输出

1. 新增平台专题，固定 operator runbook 与 negative gates 的语义和停止线。
2. 新增 fixture，声明：
   - operator evidence 只能保存 reference-only 和脱敏字段。
   - test / production secret source 必须明确，但不能提交真实 secret。
   - operator approval、sanitized verification、smoke record 和 negative gate review 是后续启用前置。
   - `production_secret_backend`、resolver runtime、fake resolver、cloud secret service、DB provider 和 repository mode 仍未启用。
3. 新增 checker，校验：
   - implementation readiness 已把 `operator-runbook-and-negative-gates` 标记为 satisfied，并引用本批证据。
   - `operator-runbook` 前置条件有本批 evidence。
   - negative gates、sanitized diagnostics、no fallback、no side effects 和 artifact guard 完整。
   - `check-repo.py` 注册顺序位于 disabled resolver interface readiness checker 之后。
4. 更新相关入口文档、周志和脚本说明。

## Negative Gates

- `operator_runbook_missing`
- `operator_approval_missing`
- `operator_environment_mismatch`
- `operator_secret_source_missing`
- `operator_sanitized_verification_missing`
- `operator_smoke_record_missing`
- `operator_resolver_invocation_disabled`
- `operator_secret_value_exposure_detected`
- `operator_fallback_forbidden`
- `operator_production_ready_claim_forbidden`

这些错误必须归入 `operator_gate` failure boundary，并直接暴露为 fail-closed 诊断；不得回退到 developer env credential、mock provider、local-smoke profile、fixture credential、committed secret value、跨环境 `secret_ref`、fake resolver 或 fake query executor。

## 验收口径

- `scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json` 存在并被 checker 消费。
- `production-ops-secret-backend-implementation-readiness.json` 中 `operator-runbook-and-negative-gates` 状态为 `satisfied`。
- `operator-runbook` 前置条件状态为 `satisfied`，且引用本批 evidence。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- operator evidence 不返回 credential handle，不保存 secret value。
- `production-secret-reference-basic` 仍为 reference-only，`resolver_enabled=false`、`cloud_calls_allowed=false`。
- 不新增 resolver runtime、fake resolver、operator runbook executor、negative gate runtime、cloud SDK、DB provider、DB driver、connection factory、SQL、schema marker、migration runner、public production API 或 repository mode runtime。

## 验证

```bash
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

## 停止线

- 不实现真实 secret backend。
- 不实现 secret resolver runtime。
- 不创建 fake resolver。
- 不调用云 secret API。
- 不读取、写入或提交真实 secret。
- 不把 `RADISHMIND_SECRET_SOURCE`、`.env.example`、developer env override、reference fixture、profile binding、disabled resolver interface、operator runbook、negative gate evidence 或 readiness checker 写成 production secret backend。
- 不接 database connection provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode、production API、executor、confirmation、writeback 或 replay。
