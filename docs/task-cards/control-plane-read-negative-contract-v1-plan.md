# `Control Plane Read Negative Contract` v1 计划

更新时间：2026-05-27

## 任务目标

本任务卡用于把 `control-plane-read-route-contract-v1` 与 `control-plane-read-response-fixtures-v1` 的只读路线，补齐为 `control-plane-read-negative-contract-v1` negative contract。

当前任务不实现 control plane API route、不创建 Go handler、不创建 TypeScript consumer、不创建数据库 schema / query、不接 `Radish` OIDC、不发放真实 API key、不执行 quota / rate limit、不实现 workflow executor，也不新增 confirmation / writeback / replay。它只固定 read-only route 的拒绝样例、fail-closed envelope、禁止方法、禁止查询参数、禁止敏感字段投影、禁止 fallback 和停止线。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [Control Plane Read Model v1 计划](control-plane-read-model-v1-plan.md)
- [Control Plane Read Route Contract v1 计划](control-plane-read-route-contract-v1-plan.md)
- [Control Plane Read Response Fixtures v1 计划](control-plane-read-response-fixtures-v1-plan.md)
- `scripts/checks/fixtures/control-plane-read-route-contract-v1.json`
- `scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json`

## negative contract 范围

v1 为七类 route contract 固定负向样例：

1. 缺少 identity context 时拒绝 tenant summary。
2. cross-tenant read attempt 拒绝 application summary。
3. 缺少 `api_keys:read` scope 时拒绝 API key summary。
4. 缺少 quota policy 时拒绝 quota summary。
5. 使用 executor-only filter 时拒绝 workflow definition summary。
6. 请求 raw tool payload filter 时拒绝 run record summary。
7. 缺少 `audit:read` scope 时拒绝 audit summary。

所有负向样例必须返回统一 envelope：`request_id`、`tenant_ref`、`items`、`next_cursor`、`failure_code` 和 `audit_ref`。拒绝响应必须 fail-closed，`items=[]`、`next_cursor=null`，且 `failure_code` 必须来自对应 route contract。

## 共享拒绝策略

- `POST`、`PUT`、`PATCH`、`DELETE` 在这些 read-only route 上全部保持 forbidden method。
- `execute`、`replay`、`confirmation_decision_ref`、`writeback_payload`、`raw_tool_payload`、`include_secret` 等查询参数不得被解释为执行入口。
- `raw_secret_value`、`api_key_value`、`api_key_hash`、`authorization_header`、`bearer_token`、`cookie_value`、`raw_request_body_dump`、`raw_tool_payload`、`business_writeback_payload`、`full_prompt_dump_with_secret` 不得出现在负向输出投影中。
- `anonymous_read`、`cross_tenant_read`、`mock_provider_fallback` 和 `ops_console_elevation` 不得作为 fallback。

## 验收口径

- `scripts/checks/fixtures/control-plane-read-negative-contract-v1.json` 固定 route negative case、共享 negative case、拒绝不变量和停止线。
- `scripts/check-control-plane-read-negative-contract-v1.py` 校验依赖切片已满足，并校验负向样例覆盖所有 read-only route。
- `check-control-plane-read-negative-contract-v1.py` 接入 `scripts/check-repo.py --fast`。
- 入口文档、任务卡入口、脚本说明和周志同步说明该切片只固定 negative contract，不实现 API、数据库、executor、confirmation、writeback 或 replay。

## 非目标

- 不实现 control plane API route 或 Go handler。
- 不创建 TypeScript consumer contract。
- 不创建数据库 schema、migration、query 或 durable store。
- 不接 `Radish` OIDC。
- 不实现正式用户端或 production admin console。
- 不实现 API key lifecycle、quota enforcement、rate limiting、billing ledger 或 cost calculation。
- 不实现 workflow builder、workflow executor、tool executor、confirmation flow、writeback、replay 或 materialized result reader。

## 停止线

- 不把 negative contract 写成已实现 API 或可运行 route。
- 不把 forbidden method、forbidden query、forbidden fallback 写成未来 handler 已完成。
- 不为了负向样例引入 executor、confirmation decision、business writeback、replay 或 database write。
- 不返回 secret、credential、raw request、raw tool payload、完整 prompt dump 或业务写回 payload。
