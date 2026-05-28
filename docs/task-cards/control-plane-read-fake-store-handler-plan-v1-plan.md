# `Control Plane Read Fake-Store Handler Plan` v1 计划

更新时间：2026-05-28

## 任务目标

本任务卡用于把 `control-plane-read-implementation-preconditions-v1` 继续拆成下一步实现计划：如果开始写 read route，只允许先做 fake-store-backed handler 切片。

该计划已被 `control-plane-read-fake-store-handler-implementation-v1` 消费前两条 single-resource route。计划本身仍只固定 Go package 落点、fake store 输入、test-only fake auth context、response conformance、route smoke 顺序和停止线；后续实现不得越过数据库、OIDC、executor、confirmation、writeback 或 replay 边界。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- [Control Plane Read Implementation Preconditions v1 计划](control-plane-read-implementation-preconditions-v1-plan.md)
- `scripts/checks/fixtures/control-plane-read-implementation-preconditions-v1.json`
- `scripts/checks/fixtures/control-plane-read-route-contract-v1.json`
- `scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json`
- `scripts/checks/fixtures/control-plane-read-negative-contract-v1.json`
- `services/platform/internal/httpapi/`

## 计划范围

1. `Go package ownership`
   - 未来 read handler 只落在 `services/platform/internal/httpapi`。
   - 未来 route registration 只修改 `services/platform/internal/httpapi/server.go`。
   - 建议文件名为 `control_plane_read.go`、`control_plane_read_fake_store.go` 与 `control_plane_read_test.go`；这些文件已由 `control-plane-read-fake-store-handler-implementation-v1` 创建并只覆盖首批 route。

2. `Fixture-backed fake store`
   - fake store 只从已提交的 read model、route contract、response fixture 和 negative contract 取数。
   - fake store 不访问数据库、不写 durable store、不调用 provider、不读 secret、不生成业务写回 payload。

3. `Test-only fake auth context`
   - route smoke 只能通过 Go test request context 注入显式 fake auth context。
   - 不新增公开 fake auth header、query scope override、anonymous read 或 ops console elevation。
   - 生产 truth source 仍是未来 `Radish OIDC / auth middleware`。

4. `Route smoke 顺序`
   - 第一阶段只做 shared read shell、fake store、auth context 和 response builder 计划。
   - 第二阶段优先计划 `tenant-summary-route` 和 `quota-summary-route` 这类 single-resource summary。
   - 第三阶段再计划 cursor list summary route。
   - 第四阶段才计划 forbidden method / query、scope denied、missing identity、tenant binding missing 和 no side effect smoke。

## 验收口径

- `scripts/checks/fixtures/control-plane-read-fake-store-handler-plan-v1.json` 固定 future Go file layout、route plan、fake store、fake auth context、route smoke 和停止线。
- `scripts/checks/control_plane/check-control-plane-read-fake-store-handler-plan-v1.py` 校验该计划完全依赖前置 fixture，不漂移到数据库、OIDC、executor 或生产能力。
- checker 接入 `scripts/check-repo.py --fast`。
- 入口文档、read-side 契约、任务卡入口、脚本说明、platform README 和周志同步说明该计划已被 partial implementation 消费，但不代表完整 read-side API ready。

## 非目标

- 不把计划本身写成完整 Go handler ready。
- 不注册计划外的 `/v1/control-plane/*` 或 `/v1/user-workspace/*` read route。
- 不创建 TypeScript consumer contract。
- 不创建数据库 schema、migration、query、repository、durable store 或真实 data source。
- 不接 `Radish` OIDC middleware，不新增公开 fake auth 入口。
- 不实现 API key lifecycle、quota enforcement、rate limiting、billing ledger 或 cost calculation。
- 不实现 workflow builder、workflow executor、tool executor、confirmation flow、writeback、replay 或 materialized result reader。

## 停止线

- 不把 fake-store-backed read handler plan 写成完整 read-side implementation ready。
- 不把未来 fake store 写成 durable store、database adapter 或 production data source。
- 不把 test-only fake auth context 写成 Radish OIDC client ready。
- 不为了 route smoke 引入 executor、confirmation decision、business writeback、replay 或 database write。
- 不返回 secret、credential、raw request、raw tool payload、完整 prompt dump 或业务写回 payload。
