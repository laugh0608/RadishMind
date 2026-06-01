# `Control Plane Read Route Contract` v1 计划

更新时间：2026-05-27

## 任务目标

本任务卡用于把 `control-plane-read-model-v1` 的七类只读摘要模型，收口为未来只读 API route 的契约前置条件。

当前任务不实现 control plane API route、不创建 Go handler、不创建 TypeScript consumer、不创建数据库 schema / query、不接 `Radish` OIDC、不发放真实 API key、不执行 quota / rate limit、不实现 workflow executor，也不新增 confirmation / writeback / replay。它只固定未来 read-only route 的路径分组、scope、tenant 约束、过滤 / 分页、错误分类、脱敏输出和停止线。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [Control Plane / User Workspace / Workflow v1 计划](control-plane-user-workspace-workflow-v1-plan.md)
- [Control Plane Read Model v1 计划](control-plane-read-model-v1-plan.md)
- `scripts/checks/fixtures/control-plane-read-model-v1.json`

## route contract 范围

v1 只固定七类未来只读 route contract：

1. `GET /v1/control-plane/tenants/{tenant_ref}/summary`
2. `GET /v1/user-workspace/applications`
3. `GET /v1/user-workspace/api-keys`
4. `GET /v1/user-workspace/usage/quota-summary`
5. `GET /v1/user-workspace/workflow-definitions`
6. `GET /v1/user-workspace/runs`
7. `GET /v1/control-plane/audit`

这些 route contract 只允许返回 `control-plane-read-model-v1` 中定义的只读摘要模型。所有 route 默认 tenant-scoped，缺少未来身份上下文、tenant binding、subject binding 或 scope grant 时必须 fail closed。

## 契约边界

- 所有 route 只读，禁止 `POST`、`PUT`、`PATCH`、`DELETE` 或写回副作用。
- 所有列表 route 必须定义 `limit`、`cursor`、`sort` 和允许的 filter key。
- 所有错误响应必须返回稳定 `failure_code`，不得 fallback 到 mock、anonymous 或 cross-tenant read。
- 所有响应必须使用脱敏字段，不返回 secret、credential、raw request、raw tool payload 或业务写回 payload。
- route contract 是未来 API / consumer 的输入约束，不代表 handler、store、query、UI 或 production readiness。

## 验收口径

- `scripts/checks/fixtures/control-plane-read-route-contract-v1.json` 固定 route contract、scope、分页 / 过滤、失败分类、脱敏策略和停止线。
- `scripts/checks/control_plane/check-control-plane-read-route-contract-v1.py` 校验依赖切片已满足，并校验本文档、入口文档、脚本说明和周志同步。
- `check-control-plane-read-route-contract-v1.py` 接入 `scripts/check-repo.py --fast`。
- `pwsh ./scripts/check-repo.ps1 -Fast` 通过。

## 非目标

- 不实现 control plane API route 或 Go handler。
- 不创建 TypeScript consumer contract。
- 不创建数据库 schema、migration、query 或 durable store。
- 不接 `Radish` OIDC。
- 不实现正式用户端或 production admin console。
- 不实现 API key lifecycle、quota enforcement、rate limiting、billing ledger 或 cost calculation。
- 不实现 workflow builder、workflow executor、tool executor、confirmation flow、writeback、replay 或 materialized result reader。

## 停止线

- 不把 route contract 写成已实现 API。
- 不让本地 ops console 变成正式用户端或生产管理端。
- 不让 route contract 绕过 tenant / subject / scope 边界。
- 不返回 secret、credential、raw request、raw tool payload 或业务写回 payload。
- 不把只读 route contract 解释为数据库、executor、confirmation、writeback、replay 或 production ready。
