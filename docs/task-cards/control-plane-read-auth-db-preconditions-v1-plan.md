# `Control Plane Read Auth/DB Preconditions` v1 计划

更新时间：2026-05-28

## 任务目标

本任务卡用于在七条 fake-store-backed read route 之后，固定未来把 read-side 从 test-only fake auth context 与 in-memory fixture fake store 迁移到真实鉴权上下文和数据库 read store 之前必须满足的前置条件。

当前任务不实现 `Radish` OIDC middleware、不做 token validation、不创建数据库 schema / migration / query、不实现 repository、不替换 fake store、不改 TypeScript consumer、不实现 API key lifecycle、quota enforcement、workflow executor、confirmation、writeback、replay 或正式 UI。它只固定 `control-plane-read-auth-db-preconditions-v1` 的 auth context contract、read store repository contract、route transition requirements、failure taxonomy、smoke transition plan 和停止线。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- [Control Plane Read Fake-Store Handler Implementation v1 计划](control-plane-read-fake-store-handler-implementation-v1-plan.md)
- [Control Plane Read Implementation Preconditions v1 计划](control-plane-read-implementation-preconditions-v1-plan.md)
- `scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json`
- `scripts/checks/fixtures/control-plane-read-implementation-preconditions-v1.json`
- `scripts/checks/fixtures/radish-oidc-client-preconditions.json`
- `scripts/checks/fixtures/control-plane-data-boundary.json`
- `scripts/checks/fixtures/control-plane-read-route-contract-v1.json`
- `scripts/checks/fixtures/control-plane-read-negative-contract-v1.json`

## 前置条件范围

1. `Auth context contract`
   - 未来 read route 只接受 `future Radish OIDC / auth middleware` 注入的 identity context、tenant binding、subject binding、scope grants、audit context、issuer ref、session ref 和 request id。
   - claim mapping 必须覆盖 issuer、subject、tenant ref、scopes、roles 和 permissions。
   - 缺少 identity、tenant、subject、scope 或 audit context 时必须 fail closed。
   - 不允许 public fake auth header、query scope override、anonymous read、cross-tenant read、ops console elevation 或 mock provider fallback。

2. `Database read store preconditions`
   - 真实 read store 必须先定义窄 repository interface，再讨论 SQL、migration 或 durable adapter。
   - 每个 read 必须使用 auth context 中的 tenant ref 作为强制 predicate，不信任 query string 的 tenant override。
   - repository 只能返回 sanitized summary projection，不返回 secret、token、raw prompt、raw request、raw tool payload 或业务写回 payload。
   - 当前只定义 `future control plane read store repository`，不实现数据库 read path。

3. `Route transition requirements`
   - 七条 route 必须保留既有 `method`、`path`、`read_model`、`required_scope` 和 fake-store implementation 当前状态。
   - 本切片只允许声明未来 auth source 与 future store source，不允许在本切片打开 OIDC validation、database query、migration 或 write。

4. `Failure taxonomy`
   - 必须继承 route contract 中的 `identity_context_missing`、`tenant_binding_missing`、`scope_denied`、`tenant_not_found`、`quota_policy_missing` 与 `invalid_filter`。
   - 未来 read store 还必须预留 `read_store_unavailable`、`read_store_contract_mismatch`、`database_read_disabled` 和 `auth_context_contract_mismatch`。

5. `Smoke transition plan`
   - 当前 smoke 仍来自 test-only fake auth context 与 in-memory fixture fake store。
   - 后续替换 fake store 前，必须补 middleware-backed auth context smoke、repository contract fake store smoke 和 read store repository contract smoke。
   - 既有成功、missing identity、tenant binding missing、scope denied、cross-tenant query denied、invalid filter、forbidden projection、forbidden method、forbidden query 和 no-side-effects 覆盖必须保留。

## 验收口径

- `scripts/checks/fixtures/control-plane-read-auth-db-preconditions-v1.json` 固定 auth context、database read store、route transition、failure taxonomy、smoke transition 和停止线。
- `scripts/checks/control_plane/check-control-plane-read-auth-db-preconditions-v1.py` 校验该切片依赖已完成的 fake-store handler implementation、read implementation preconditions、Radish OIDC preconditions、control plane data boundary、route contract 和 negative contract。
- checker 接入 `scripts/check-repo.py --fast`。
- 入口文档、read-side 契约、任务卡入口、脚本说明、platform README 和周志同步说明该切片是 auth/db preconditions，不实现真实 auth middleware、数据库 query 或 repository。

## 非目标

- 不实现 `Radish` OIDC middleware、token validation、login / logout route、session cookie 或 production auth policy。
- 不创建数据库 schema、migration、query、durable store、真实 repository 或 production data source。
- 不替换现有 in-memory fixture fake store。
- 不写 public fake auth header 或 query-based scope override。
- 不创建 TypeScript consumer contract、正式 user workspace UI 或 production admin console。
- 不实现 API key lifecycle、quota enforcement、rate limiting、billing ledger 或 cost calculation。
- 不实现 workflow builder、workflow executor、tool executor、confirmation flow、writeback、replay 或 materialized result reader。

## 停止线

- 不把 `control-plane-read-auth-db-preconditions-v1` 写成 `Radish` OIDC ready、auth middleware ready、database ready、repository ready 或完整 read-side API ready。
- 不为了迁移路线提前写 SQL、migration、ORM、durable adapter 或 production data source。
- 不把 test-only fake auth context 暴露成 public auth 输入。
- 不把 repository contract 写成 API key lifecycle、quota enforcement、workflow executor、confirmation、business writeback 或 replay。
- 不返回 secret、credential、raw request、raw tool payload、完整 prompt dump 或业务写回 payload。
