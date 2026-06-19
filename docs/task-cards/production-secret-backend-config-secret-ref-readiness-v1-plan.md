# `Production Secret Backend Config / Secret Ref Readiness` v1 计划

更新时间：2026-06-19

## 任务目标

本任务卡用于把 `production-secret-backend-implementation-readiness` 中的 `config-secret-ref-readiness` 切片推进为可检查证据。

本批只固定配置注入点、secret reference manifest 消费边界、脱敏诊断字段、失败映射、no fallback、no side effects 和 artifact guard；不实现 resolver runtime，不创建 fake resolver，不读取 secret value，不调用云 secret 服务，不接数据库，不启用 workflow saved draft repository mode。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [平台专题入口](../platform/README.md)
- [Production Secret Backend Config / Secret Ref Readiness v1](../platform/production-secret-backend-config-secret-ref-readiness-v1.md)
- [Production Secret Backend Implementation v1 计划](production-secret-backend-implementation-v1-plan.md)
- `docs/contracts/production-secret-reference.md`
- `contracts/production-secret-reference.schema.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/check-production-ops-secret-backend-implementation-readiness.py`
- `scripts/check-production-secret-reference-contract.py`
- `services/platform/README.md`

## 当前事实

- `secret-ref-schema-and-fixtures` 已落地，并保持 reference-only。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- `config-secret-ref-readiness` 在上一版 implementation readiness 中仍为 `planned_not_started`。
- workflow saved draft secret resolver implementation entry review 已确认 secret resolver 当前不打开。

## 本批输出

1. 新增平台专题，固定 config / secret ref readiness 的语义和停止线。
2. 新增 fixture，声明：
   - 配置注入点只处理 `secret_ref` 存在性和脱敏状态。
   - `secret_backend_configured`、`secret_ref_present`、`missing_secret_refs`、`field_sources` 只能作为脱敏诊断字段。
   - `production_secret_backend`、resolver、fake resolver、cloud secret service、DB provider 和 repository mode 仍未启用。
3. 新增 checker，校验：
   - reference-only manifest 未被提升为 resolver ready。
   - implementation readiness 已把 `config-secret-ref-readiness` 标记为 satisfied，并引用本批证据。
   - 失败映射、no fallback、no side effects 和 artifact guard 完整。
   - `check-repo.py` 注册顺序位于 production secret reference contract 之后。
4. 更新相关入口文档、周志和脚本说明。

## Failure Mapping

- `secret_reference_manifest_missing`
- `secret_reference_manifest_invalid`
- `secret_ref_missing`
- `secret_backend_disabled`
- `resolver_invocation_forbidden`

这些错误必须归入 `configuration` failure boundary，并直接暴露为 fail-closed 诊断；不得回退到 developer env credential、mock provider、local-smoke profile、fixture credential 或 sample。

## 验收口径

- `scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json` 存在并被 checker 消费。
- `production-ops-secret-backend-implementation-readiness.json` 中 `config-secret-ref-readiness` 状态为 `satisfied`。
- `config-injection-point` 前置条件有本批 evidence。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- `production-secret-reference-basic` 仍为 reference-only，`resolver_enabled=false`、`cloud_calls_allowed=false`。
- 不新增 resolver runtime、fake resolver、cloud SDK、DB provider、DB driver、connection factory、SQL、schema marker、migration runner 或 repository mode runtime。

## 验证

```bash
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
- 不把 `RADISHMIND_SECRET_SOURCE`、`.env.example`、developer env override、reference fixture 或 readiness checker 写成 production secret backend。
- 不接 database connection provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode、production API、executor、confirmation、writeback 或 replay。

