# Control Plane Read-Side 契约

更新时间：2026-05-28

## 契约目的

本专题说明 `Control Plane / User Workspace / Workflow v1` 的只读控制面契约层。它把用户工作区和管理端会消费的 summary、route、response、negative contract、implementation preconditions、fake-store-backed read handler plan、fake-store-backed handler implementation、auth/db preconditions 和 consumer contract 固定为可检查治理边界，避免在正式数据库、OIDC 或 UI 尚未准备好时，从本地 ops console 直接堆出产品功能。

当前 read-side 已实现七条 fake-store-backed Go read route：`tenant-summary-route`、`application-summary-list-route`、`api-key-summary-list-route`、`quota-summary-route`、`workflow-definition-summary-list-route`、`run-record-summary-list-route` 与 `audit-summary-list-route`。这些 route 只使用 in-memory fixture fake store 与 test-only fake auth context，不代表完整 read-side API、数据库 query、真实 OIDC 或正式 UI 已实现。

`control-plane-read-auth-db-preconditions-v1` 进一步固定未来迁移到 `future Radish OIDC / auth middleware` 与 `future control plane read store repository` 之前的准入条件。它只定义 auth context contract、read store repository contract、route transition requirements、failure taxonomy 和 smoke transition plan，不实现真实 auth middleware、数据库 query 或 repository。

`control-plane-read-consumer-contract-v1` 固定 `contracts/typescript/control-plane-read-api.ts`、`scripts/run-control-plane-read-consumer-smoke.py` 和上层 view model 停止线。它只定义 TypeScript consumer contract 与离线 fixture 消费方式，不实现正式 user workspace UI、production admin console、真实 auth/db 或 production API。

## 分层关系

当前 read-side 契约按九层固定：

1. `control-plane-read-model-v1`
   - 固定 tenant、application、API key、quota、workflow definition、run record 和 audit 的只读 summary 模型。
   - 固定访问策略、脱敏策略和停止线。
2. `control-plane-read-route-contract-v1`
   - 固定七类 tenant-scoped read-only route contract。
   - 固定 `GET` 方法、scope、分页 / 过滤、失败分类和 fail-closed 访问边界。
3. `control-plane-read-response-fixtures-v1`
   - 固定统一 response envelope：`request_id`、`tenant_ref`、`items`、`next_cursor`、`failure_code`、`audit_ref`。
   - 固定成功 / 失败样例、`failure_code` 来源和敏感字段脱敏。
4. `control-plane-read-negative-contract-v1`
   - 固定缺少身份、cross-tenant read、scope denied、invalid filter、forbidden method、forbidden query、forbidden fallback 和敏感字段投影拒绝。
   - 固定 fail-closed、无副作用、无 executor、无 database write、无 confirmation decision、无 writeback 和无 replay。
5. `control-plane-read-implementation-preconditions-v1`
   - 固定未来 read route 实现前必须具备的 `handler ownership`、`fake store` strategy、`auth middleware` dependency、response fixture conformance 和 negative route smoke readiness。
   - 固定首轮实现只能进入 fixture-backed fake store 与显式 fake auth context，不声明 Go handler、route smoke、数据库、OIDC、API key lifecycle、quota enforcement、executor、confirmation、writeback 或 replay ready。
6. `control-plane-read-fake-store-handler-plan-v1`
   - 固定未来 fake-store-backed read handler plan 的 Go package 落点、future file layout、test-only fake auth context、route smoke 顺序和 no-side-effect gate。
   - 这是计划证据，本身不创建 handler；已被下一层 implementation 消费。
7. `control-plane-read-fake-store-handler-implementation-v1`
   - 固定七条 read route 的 fake-store-backed handler implementation。
   - 只使用 `services/platform/internal/httpapi` 内的 in-memory fake store 与 test-only fake auth context，不接数据库、OIDC、executor、confirmation、writeback 或 replay。
8. `control-plane-read-auth-db-preconditions-v1`
   - 固定未来替换 test-only fake auth context 与 in-memory fixture fake store 前必须满足的 auth/db preconditions。
   - 固定 `future Radish OIDC / auth middleware`、`future control plane read store repository`、route transition、failure taxonomy、smoke transition 和停止线；不接真实 OIDC、不写数据库 query、不实现 repository。
9. `control-plane-read-consumer-contract-v1`
   - 固定上层消费契约：route catalog、request / response 类型、统一 envelope、failure view、cursor view、forbidden output 检测和只读 view model。
   - 由 `contracts/typescript/control-plane-read-api.ts` 与 `scripts/run-control-plane-read-consumer-smoke.py --check` 固定；不实现正式 UI、不请求真实后端、不接数据库或 OIDC。

## 程序化证据

read-side 契约当前由以下 fixture 和 checker 固定：

- `scripts/checks/fixtures/control-plane-read-model-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-model-v1.py`
- `scripts/checks/fixtures/control-plane-read-route-contract-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-route-contract-v1.py`
- `scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-response-fixtures-v1.py`
- `scripts/checks/fixtures/control-plane-read-negative-contract-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-negative-contract-v1.py`
- `scripts/checks/fixtures/control-plane-read-implementation-preconditions-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-implementation-preconditions-v1.py`
- `scripts/checks/fixtures/control-plane-read-fake-store-handler-plan-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-fake-store-handler-plan-v1.py`
- `scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-fake-store-handler-implementation-v1.py`
- `scripts/checks/fixtures/control-plane-read-auth-db-preconditions-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-auth-db-preconditions-v1.py`
- `scripts/checks/fixtures/control-plane-read-consumer-contract-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-consumer-contract-v1.py`
- `contracts/typescript/control-plane-read-api.ts`
- `scripts/run-control-plane-read-consumer-smoke.py`

这些 checker 已接入 `scripts/check-repo.py --fast`。它们的作用是防止契约、样例、负向边界、实现前置条件、fake-store-backed read handler plan、handler implementation、auth/db preconditions、consumer contract 和文档说明互相漂移，不负责启动服务或模拟真实数据库。

## 路由范围

当前固定的七类只读 route 是：

- `GET /v1/control-plane/tenants/{tenant_ref}/summary`
- `GET /v1/user-workspace/applications`
- `GET /v1/user-workspace/api-keys`
- `GET /v1/user-workspace/usage/quota-summary`
- `GET /v1/user-workspace/workflow-definitions`
- `GET /v1/user-workspace/runs`
- `GET /v1/control-plane/audit`

这些 route 已由 `control-plane-read-fake-store-handler-implementation-v1` 注册为 fake-store-backed Go route，并由 Go 单元测试覆盖成功、missing identity、tenant binding missing、cross-tenant query denied、scope denied、invalid filter、forbidden sensitive projection、forbidden method、forbidden query 和 no-side-effects。`control-plane-read-consumer-contract-v1` 已固定 TypeScript consumer contract；这些 route 仍不是数据库 read path、真实 OIDC auth path、正式 UI consumer ready 或 production ready。

## 输出边界

read-side 输出只允许暴露已脱敏 summary、计数、状态、成本摘要、trace id、audit ref 和 redacted secret reference。以下内容不得通过 read-side 输出或投影泄漏：

- raw secret value
- API key value 或 API key hash
- authorization header、bearer token、cookie value
- raw request body dump
- raw tool payload
- business writeback payload
- full prompt dump with secret

## 停止线

- 不把 read model 写成数据库 schema。
- 不把 route contract 写成 Go handler 已完成。
- 不把 response fixture 写成真实 API 返回。
- 不把 negative contract 写成 route smoke 已完成。
- 不把 implementation preconditions 写成 Go handler、fake store、auth middleware 或 route smoke 已完成。
- 不把 fake-store-backed read handler plan 写成完整 route implementation。
- 不把 `control-plane-read-fake-store-handler-implementation-v1` 写成完整 read-side API、真实数据库 read path、OIDC auth path 或 UI consumer ready。
- 不把 `control-plane-read-auth-db-preconditions-v1` 写成 auth middleware ready、database ready、repository ready 或完整 read-side API ready。
- 不把 `future control plane read store repository` 写成当前数据库 query、migration 或 durable adapter。
- 不把 `control-plane-read-consumer-contract-v1` 写成正式 user workspace UI、production admin console、真实 API consumer 或 production ready。
- 不把 `scripts/run-control-plane-read-consumer-smoke.py` 写成真实 API smoke；它只消费 committed fixture。
- 不通过 read-side 契约启用 OIDC、API key lifecycle、quota enforcement、billing ledger、workflow executor、confirmation、business writeback 或 replay。
- 不把本地 ops console 提升为正式 user workspace 或 production admin console。
