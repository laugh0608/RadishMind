# `Control Plane Read Fake-Store Handler Implementation` v1 计划

更新时间：2026-05-28

## 任务目标

本任务卡用于记录 `control-plane-read-fake-store-handler-implementation-v1` 的首个真实 Go 切片：只把 `tenant-summary-route` 与 `quota-summary-route` 从计划推进到 fake-store-backed handler。

该切片不实现完整 read-side API，不接数据库、不接 `Radish` OIDC、不创建 TypeScript consumer、不实现 API key lifecycle / quota enforcement、不实现 workflow executor，也不新增 confirmation / writeback / replay。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- [Control Plane Read Fake-Store Handler Plan v1 计划](control-plane-read-fake-store-handler-plan-v1-plan.md)
- `scripts/checks/fixtures/control-plane-read-fake-store-handler-plan-v1.json`
- `scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json`
- `services/platform/internal/httpapi/`

## 实现范围

1. `tenant-summary-route`
   - 注册 `GET /v1/control-plane/tenants/{tenant_ref}/summary`。
   - handler 为 `handleControlPlaneTenantSummary`。
   - 只读取 fixture-backed fake store 中的 `tenant_demo` summary。
   - 通过 test-only fake auth context 验证 `tenant:read` scope 和 tenant binding。

2. `quota-summary-route`
   - 注册 `GET /v1/user-workspace/usage/quota-summary`。
   - handler 为 `handleUserWorkspaceQuotaSummary`。
   - 只读取 fixture-backed fake store 中的 `quota_demo_current` summary。
   - 通过 test-only fake auth context 验证 `usage:read` scope。

3. `负向 smoke`
   - 覆盖 missing identity、tenant binding mismatch、scope denied、forbidden method、forbidden query 和 no-side-effects。
   - 拒绝 `execute`、`replay`、`confirmation_decision_ref`、`writeback_payload`、`raw_tool_payload` 与 `include_secret` 查询参数。

## 验收口径

- `services/platform/internal/httpapi/control_plane_read.go` 承载 handler、auth check、envelope 和 forbidden query gate。
- `services/platform/internal/httpapi/control_plane_read_fake_store.go` 承载 fixture-backed fake store 和 no-side-effect counters。
- `services/platform/internal/httpapi/control_plane_read_test.go` 覆盖两条成功 route 与主要拒绝路径。
- `scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json` 固定实现范围、剩余 route、fake store、auth mode、smoke coverage 和停止线。
- `scripts/checks/control_plane/check-control-plane-read-fake-store-handler-implementation-v1.py` 校验 Go 文件、route registration、error taxonomy、测试覆盖和文档同步。

## 非目标

- 不注册 `application-summary-list-route`、`api-key-summary-list-route`、`workflow-definition-summary-list-route`、`run-record-summary-list-route` 或 `audit-summary-list-route`。
- 不创建数据库 schema、migration、query、repository、durable store 或真实 data source。
- 不接 `Radish` OIDC middleware，不新增公开 fake auth header。
- 不实现 API key 生成、哈希、验证、quota enforcement、rate limiting、billing ledger 或 cost calculation。
- 不实现 workflow builder、workflow executor、tool executor、confirmation flow、writeback、replay 或 materialized result reader。

## 停止线

- 不把 partial fake-store-backed handler implementation 写成完整 control plane API ready。
- 不把 fake store 写成 durable store、database adapter 或 production data source。
- 不把 test-only fake auth context 写成 Radish OIDC client ready。
- 不为了 route smoke 引入 executor、confirmation decision、business writeback、replay 或 database write。
- 不返回 secret、credential、raw request、raw tool payload、完整 prompt dump 或业务写回 payload。
