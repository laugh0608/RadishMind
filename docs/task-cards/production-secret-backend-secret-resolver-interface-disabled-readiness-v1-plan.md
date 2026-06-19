# `Production Secret Backend Secret Resolver Interface Disabled Readiness` v1 计划

更新时间：2026-06-19

## 任务目标

本任务卡用于把 `production-secret-backend-implementation-readiness` 中的 `secret-resolver-interface-disabled` 切片推进为可检查证据。

本批只固定 future resolver interface 的 reference-only 输入、disabled result、sanitized diagnostics、failure mapping、no fallback、no side effects 和 artifact guard；不实现 resolver runtime，不创建 fake resolver，不读取 secret value，不调用云 secret 服务，不访问 provider，不接数据库，不启用 workflow saved draft repository mode。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [平台专题入口](../platform/README.md)
- [Production Secret Backend Config / Secret Ref Readiness v1](../platform/production-secret-backend-config-secret-ref-readiness-v1.md)
- [Production Secret Backend Provider Profile Secret Binding Readiness v1](../platform/production-secret-backend-provider-profile-secret-binding-readiness-v1.md)
- [Production Secret Backend Secret Resolver Interface Disabled Readiness v1](../platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md)
- [Production Secret Backend Implementation v1 计划](production-secret-backend-implementation-v1-plan.md)
- `contracts/production-secret-reference.schema.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`
- `scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json`
- `scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py`
- `scripts/check-production-ops-secret-backend-implementation-readiness.py`
- `scripts/check-production-secret-reference-contract.py`
- `services/platform/README.md`

## 当前事实

- `secret-ref-schema-and-fixtures` 已落地，并保持 reference-only。
- `config-secret-ref-readiness` 已落地，并只满足配置注入点前置条件。
- `provider-profile-secret-binding` 已落地，并只满足 provider/profile binding 前置条件。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- `secret-resolver-interface-disabled` 在上一版 implementation readiness 中仍为 `planned_not_started`。
- workflow saved draft secret resolver implementation entry review 已确认 secret resolver 当前不打开。

## 本批输出

1. 新增平台专题，固定 disabled resolver interface 的语义和停止线。
2. 新增 fixture，声明：
   - resolver input 只能消费 reference-only 字段。
   - disabled result 只能返回 `resolver_state=disabled`、`resolver_available=false`、`credential_handle_present=false` 等脱敏状态。
   - missing ref、backend disabled、resolver unavailable、resolution denied、environment mismatch 必须 fail closed。
   - `production_secret_backend`、resolver runtime、fake resolver、cloud secret service、DB provider 和 repository mode 仍未启用。
3. 新增 checker，校验：
   - implementation readiness 已把 `secret-resolver-interface-disabled` 标记为 satisfied，并引用本批证据。
   - failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard 完整。
   - `check-repo.py` 注册顺序位于 provider profile secret binding readiness checker 之后。
4. 更新相关入口文档、周志和脚本说明。

## Failure Mapping

- `secret_resolver_secret_ref_missing`
- `secret_resolver_backend_disabled`
- `secret_resolver_unavailable`
- `secret_resolution_denied`
- `secret_resolver_environment_mismatch`
- `secret_resolver_invocation_disabled`

这些错误必须归入 `configuration` failure boundary，并直接暴露为 fail-closed 诊断；不得回退到 developer env credential、mock provider、local-smoke profile、fixture credential、committed secret value、跨环境 `secret_ref`、fake resolver 或 fake query executor。

## 验收口径

- `scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json` 存在并被 checker 消费。
- `production-ops-secret-backend-implementation-readiness.json` 中 `secret-resolver-interface-disabled` 状态为 `satisfied`。
- `sanitized-audit-fields` 和 `failure-taxonomy` 前置条件有本批 evidence。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- disabled result 不返回 credential handle。
- `production-secret-reference-basic` 仍为 reference-only，`resolver_enabled=false`、`cloud_calls_allowed=false`。
- 不新增 resolver runtime、fake resolver、cloud SDK、DB provider、DB driver、connection factory、SQL、schema marker、migration runner 或 repository mode runtime。

## 验证

```bash
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
- 不把 `RADISHMIND_SECRET_SOURCE`、`.env.example`、developer env override、reference fixture、profile binding、disabled resolver interface 或 readiness checker 写成 production secret backend。
- 不接 database connection provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode、production API、executor、confirmation、writeback 或 replay。
