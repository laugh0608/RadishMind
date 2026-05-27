# `Control Plane Read Model` v1 计划

更新时间：2026-05-27

## 任务目标

本任务卡用于把 `Control Plane / User Workspace / Workflow v1` 已完成的五个治理边界，收口为下一步可实现前的只读 read model 前置条件。

当前任务不实现 control plane API、不创建数据库 schema 或 migration、不接 `Radish` OIDC、不发放真实 API key、不执行 quota / rate limit、不实现 workflow executor，也不新增 confirmation / writeback / replay。它只定义用户端和管理端未来可读取的最小摘要模型、访问边界、脱敏字段和停止线，避免后续直接从本地 ops console 堆正式产品功能。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [Control Plane / User Workspace / Workflow v1 计划](control-plane-user-workspace-workflow-v1-plan.md)
- [产品范围与目标](../radishmind-product-scope.md)
- [系统架构](../radishmind-architecture.md)
- [阶段路线图](../radishmind-roadmap.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- `scripts/checks/fixtures/product-surface-v1-boundary.json`
- `scripts/checks/fixtures/control-plane-data-boundary.json`
- `scripts/checks/fixtures/radish-oidc-client-preconditions.json`
- `scripts/checks/fixtures/gateway-api-key-quota-readiness.json`
- `scripts/checks/fixtures/workflow-definition-run-record-boundary.json`

## read model 范围

v1 只固定七类摘要模型：

1. `tenant summary`
2. `application summary`
3. `API key summary`
4. `quota summary`
5. `workflow definition summary`
6. `run record summary`
7. `audit summary`

这些模型只允许承载 ID、状态、归属、摘要、计数、失败分类、成本摘要、trace / audit 引用和脱敏状态；不得包含 raw secret、真实 API key、authorization header、cookie、未脱敏 tool payload、业务写回 payload 或完整 prompt dump。

## 访问边界

- 所有 read model 默认 tenant-scoped。
- 缺少未来身份上下文、tenant binding、subject binding 或 scope 时必须 fail closed。
- 用户端只能读取已授权 tenant / project / application / run 的摘要。
- 管理端只能读取配置与审计摘要，不绕过 runtime 执行业务写回。
- read model 是未来 API / UI 的输入约束，不是数据库 schema，也不是 durable store。

## 验收口径

- `scripts/checks/fixtures/control-plane-read-model-v1.json` 固定 read model、访问策略、脱敏策略和停止线。
- `scripts/check-control-plane-read-model-v1.py` 校验依赖切片已满足，并校验本文档、入口文档、脚本说明和周志同步。
- `check-control-plane-read-model-v1.py` 接入 `scripts/check-repo.py --fast`。
- `pwsh ./scripts/check-repo.ps1 -Fast` 通过。

## 非目标

- 不实现正式用户端。
- 不实现 production admin console。
- 不实现 control plane API route。
- 不创建数据库 schema、migration、query 或 durable store。
- 不接 `Radish` OIDC。
- 不实现 API key 生命周期、quota enforcement、rate limiting、billing ledger 或 cost calculation。
- 不实现 workflow builder、workflow executor、tool executor、confirmation flow、writeback、replay 或 materialized result reader。
- 不接真实 production secret backend，不写入真实 secret。

## 停止线

- 不把 read model fixture 写成真实存储模型。
- 不让本地 ops console 变成正式用户端或生产管理端。
- 不让 read model 绕过 tenant / subject / scope 边界。
- 不返回 secret、credential、raw request、raw tool payload 或业务写回 payload。
- 不把只读摘要模型解释为 API、数据库、executor、confirmation、writeback、replay 或 production ready。
