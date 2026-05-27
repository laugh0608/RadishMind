# `Control Plane Read Response Fixtures` v1 计划

更新时间：2026-05-27

## 任务目标

本任务卡用于把 `control-plane-read-route-contract-v1` 的七类 read-only route contract，收口为 `control-plane-read-response-fixtures-v1` 响应样例和失败样例。

当前任务不实现 control plane API route、不创建 Go handler、不创建 TypeScript consumer、不创建数据库 schema / query、不接 `Radish` OIDC、不发放真实 API key、不执行 quota / rate limit、不实现 workflow executor，也不新增 confirmation / writeback / replay。它只固定 read-only response envelope、成功响应字段、失败响应字段、分页字段、脱敏输出和停止线。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [Control Plane Read Model v1 计划](control-plane-read-model-v1-plan.md)
- [Control Plane Read Route Contract v1 计划](control-plane-read-route-contract-v1-plan.md)
- `scripts/checks/fixtures/control-plane-read-model-v1.json`
- `scripts/checks/fixtures/control-plane-read-route-contract-v1.json`

## fixture 范围

v1 为七类 route contract 固定响应样例：

1. tenant summary response
2. application summary list response
3. API key summary list response
4. quota summary response
5. workflow definition summary list response
6. run record summary list response
7. audit summary list response

每个响应样例必须包含 `request_id`、`tenant_ref`、`items`、`next_cursor`、`failure_code` 和 `audit_ref`。列表响应必须保留 cursor 语义；单资源响应也使用统一 envelope，`next_cursor` 为 `null`。

## 失败样例

每个 route 至少固定一个 fail-closed response，覆盖缺少身份、tenant binding、scope denied 或 invalid filter 等稳定 `failure_code`。失败响应不得 fallback 到 anonymous、mock provider、ops console elevation 或 cross-tenant read。

## 非目标

- 不实现 control plane API route 或 Go handler。
- 不创建 TypeScript consumer contract。
- 不创建数据库 schema、migration、query 或 durable store。
- 不接 `Radish` OIDC。
- 不实现正式用户端或 production admin console。
- 不实现 API key lifecycle、quota enforcement、rate limiting、billing ledger 或 cost calculation。
- 不实现 workflow builder、workflow executor、tool executor、confirmation flow、writeback、replay 或 materialized result reader。

## 停止线

- 不把 response fixture 写成已实现 API。
- 不返回 secret、credential、raw request、raw tool payload、完整 prompt dump 或业务写回 payload。
- 不把响应样例解释为数据库、handler、consumer、executor、confirmation、writeback、replay 或 production ready。
