# `Control Plane Read Implementation Preconditions` v1 计划

更新时间：2026-05-28

## 任务目标

本任务卡用于把已完成的 read model、read-only route contract、response fixture 和 negative contract，收口为未来真正实现 read route 前必须满足的实现前置条件。

当前任务不实现 Go route handler、不创建 TypeScript consumer、不创建数据库 schema / query、不接 `Radish` OIDC、不发放真实 API key、不执行 quota / rate limit、不实现 workflow executor，也不新增 confirmation / writeback / replay。它只固定 handler ownership、fake store strategy、auth middleware dependency、response fixture conformance、negative route smoke readiness 和停止线。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- [Control Plane Read Model v1 计划](control-plane-read-model-v1-plan.md)
- [Control Plane Read Route Contract v1 计划](control-plane-read-route-contract-v1-plan.md)
- [Control Plane Read Response Fixtures v1 计划](control-plane-read-response-fixtures-v1-plan.md)
- [Control Plane Read Negative Contract v1 计划](control-plane-read-negative-contract-v1-plan.md)
- `scripts/checks/fixtures/control-plane-read-model-v1.json`
- `scripts/checks/fixtures/control-plane-read-route-contract-v1.json`
- `scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json`
- `scripts/checks/fixtures/control-plane-read-negative-contract-v1.json`

## 前置条件范围

1. `Go handler ownership`
   - 首个 read route handler 仍归 Go platform / control-plane read side。
   - handler 必须只消费 route contract、auth context、fake store 和 response builder，不直接访问数据库、不绕过 canonical envelope。

2. `Fake store strategy`
   - 首轮实现若要进入代码，只能使用 fixture-backed fake store。
   - fake store 只服务 route smoke 和 response conformance，不声明 durable store、database query 或 production data source ready。

3. `Auth middleware dependency`
   - read handler 必须依赖未来 Radish OIDC / auth middleware 提供的 identity context、tenant binding、subject binding、scope grants 和 audit context。
   - 在 middleware 未实现前，route smoke 只能用显式 fake auth context，不得写匿名 fallback 或 ops console elevation。

4. `Response fixture conformance`
   - 每个 route 的成功 / 失败输出必须匹配 `control-plane-read-response-fixtures-v1` 的 envelope 和脱敏策略。
   - `failure_code` 必须来自 route contract，不得新增未声明错误码。

5. `Negative route smoke readiness`
   - 未来 route smoke 必须覆盖 missing identity、tenant binding missing、scope denied、invalid filter / route-specific failure、forbidden method、forbidden query 和 no side effect。
   - route smoke 不得调用 executor、写数据库、触发 confirmation、业务写回或 replay。

## 验收口径

- `scripts/checks/fixtures/control-plane-read-implementation-preconditions-v1.json` 固定 handler ownership、fake store、auth dependency、response conformance、negative smoke readiness 和停止线。
- `scripts/checks/control_plane/check-control-plane-read-implementation-preconditions-v1.py` 校验依赖切片已满足，并校验七类 read-only route 都具备实现前置条件。
- checker 接入 `scripts/check-repo.py --fast`。
- 入口文档、read-side 契约、任务卡入口、脚本说明和周志同步说明该切片只固定实现前置条件，不实现 API、数据库、executor、confirmation、writeback 或 replay。

## 非目标

- 不实现 Go route handler 或 HTTP route 注册。
- 不创建 TypeScript consumer contract。
- 不创建数据库 schema、migration、query、durable store 或真实 repository。
- 不接 `Radish` OIDC middleware。
- 不实现正式用户端或 production admin console。
- 不实现 API key lifecycle、quota enforcement、rate limiting、billing ledger 或 cost calculation。
- 不实现 workflow builder、workflow executor、tool executor、confirmation flow、writeback、replay 或 materialized result reader。

## 停止线

- 不把 implementation preconditions 写成 route implementation ready。
- 不把 fake store 写成 durable store 或 database adapter。
- 不把 fake auth context 写成 Radish OIDC client ready。
- 不为了 route smoke 引入 executor、confirmation decision、business writeback、replay 或 database write。
- 不返回 secret、credential、raw request、raw tool payload、完整 prompt dump 或业务写回 payload。
