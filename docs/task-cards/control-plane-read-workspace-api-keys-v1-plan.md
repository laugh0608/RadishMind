# Control Plane Read Workspace API Keys v1 计划

更新时间：2026-05-31

## 目的

本任务卡用于记录 `control-plane-read-workspace-api-keys-v1`：在 `apps/radishmind-web/` 的 read-only shared shell 内实现用户工作区 API key 列表页切片 `workspace-api-keys`。

当前切片只消费 `api-key-summary-list-route` 的 TypeScript consumer contract 和离线 view model，展示 API key summary 列表、route 绑定、scope、状态、时间字段、审计引用、只读状态和 forbidden output guard。它不请求真实后端，不接数据库、OIDC、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay，也不展示 key value、key hash、raw secret 或 authorization material。

## 依赖

- [Control Plane Read Shared Shell v1 计划](control-plane-read-shared-shell-v1-plan.md)
- [Control Plane Read Admin Tenant Overview v1 计划](control-plane-read-admin-tenant-overview-v1-plan.md)
- [Control Plane Read Workspace Applications v1 计划](control-plane-read-workspace-applications-v1-plan.md)
- [Control Plane Read Formal UI Implementation Readiness v1 计划](control-plane-read-formal-ui-implementation-readiness-v1-plan.md)
- [Control Plane Read Formal UI Boundary v1 计划](control-plane-read-formal-ui-boundary-v1-plan.md)
- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- `contracts/typescript/control-plane-read-api.ts`

## 实现范围

1. 新增 `src/features/control-plane-read/workspaceApiKeys.ts`。
   - 复用 `CONTROL_PLANE_READ_ROUTE_DEFINITIONS`、`CONTROL_PLANE_READ_ROUTES`、`APIKeySummary`、`ControlPlaneReadResponseByRoute` 与 `toControlPlaneReadCollectionViewModel`。
   - 只构造离线 `api-key-summary-list-route` view model，不复制 route path 字面量。
   - 显示 API key id、owner、scopes、state、created / expires / last used 时间字段。
   - 继续使用 `controlPlaneReadResponseHasForbiddenOutput` 拦截敏感输出字段。
2. 更新 `src/app/App.tsx`。
   - 在 shared shell 内新增 `workspace-api-keys` 页面区域。
   - 展示 route metadata、request / audit ref、API key 列表、指标卡和页面状态预览。
   - 不提供 issue / rotate / revoke / edit / confirm / replay / live backend 请求按钮。
3. 更新 `src/styles.css`。
   - 为 API key route、metric、list row、scope chips 和状态预览补充响应式布局。
4. 新增 `control-plane-read-workspace-api-keys-v1.json` 与 `check-control-plane-read-workspace-api-keys-v1.py`。
   - 固定页面切片、契约复用、敏感字段不展示、状态覆盖、禁用 live backend / write controls 和文档同步。

## 停止线

- 不把 `apps/radishmind-web/` 写成完整 formal user workspace。
- 不把 `workspace-api-keys` 写成 production user workspace ready。
- 不在页面内请求 live backend。
- 不展示或构造 API key value、API key hash、raw secret、authorization header、bearer token 或 cookie value。
- 不添加数据库 schema、migration、query 或 repository。
- 不添加 `Radish` OIDC middleware、token validation、login 或 logout。
- 不添加 API key issuance、rotation、revoke、quota enforcement、rate limiting、billing 或 cost ledger。
- 不添加 workflow builder、executor、confirmation、writeback、replay 或 materialized result reader。
- 不修改 `apps/radishmind-console/` 的本地 ops surface 边界。

## 验证

- `npm run build`（工作目录：`apps/radishmind-web/`）
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-workspace-api-keys-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-workspace-applications-v1.py`
- `./scripts/run-python.sh scripts/run-control-plane-read-consumer-smoke.py --check`
- `./scripts/check-repo-fast.sh`

## 后续顺序

若继续 formal UI read-side，应按 readiness 顺序推进 `workspace-usage-quota`、`workspace-workflow-definitions`、`workspace-run-history` 和 `admin-audit-log`；若改走真实 auth/db，则先补 repository / auth middleware 任务卡与 checker，不与页面切片并行。
