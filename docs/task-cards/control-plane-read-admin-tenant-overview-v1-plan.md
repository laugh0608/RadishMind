# Control Plane Read Admin Tenant Overview v1 计划

更新时间：2026-05-31

## 目的

本任务卡用于记录 `control-plane-read-admin-tenant-overview-v1`：在 `apps/radishmind-web/` 的 read-only shared shell 内实现第一个正式页面切片 `admin-tenant-overview`。

当前切片只消费 `tenant-summary-route` 的 TypeScript consumer contract 和离线 view model，展示租户摘要、route 绑定、审计引用、只读状态和 forbidden output guard。它不请求真实后端，不接数据库、OIDC、API key / quota、workflow executor、confirmation、writeback 或 replay，也不修改 `apps/radishmind-console/` 的本地 ops surface 定位。

## 依赖

- [Control Plane Read Shared Shell v1 计划](control-plane-read-shared-shell-v1-plan.md)
- [Control Plane Read Formal UI Implementation Readiness v1 计划](control-plane-read-formal-ui-implementation-readiness-v1-plan.md)
- [Control Plane Read Formal UI Boundary v1 计划](control-plane-read-formal-ui-boundary-v1-plan.md)
- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- `contracts/typescript/control-plane-read-api.ts`

## 实现范围

1. 新增 `src/features/control-plane-read/adminTenantOverview.ts`。
   - 复用 `CONTROL_PLANE_READ_ROUTE_DEFINITIONS`、`CONTROL_PLANE_READ_ROUTES`、`TenantSummary`、`ControlPlaneReadResponseByRoute` 与 `toControlPlaneReadCollectionViewModel`。
   - 只构造离线 `tenant-summary-route` view model，不复制 route path 字面量。
   - 继续使用 `controlPlaneReadResponseHasForbiddenOutput` 拦截敏感输出字段。
2. 更新 `src/app/App.tsx`。
   - 在 shared shell 内新增 `admin-tenant-overview` 页面区域。
   - 展示 tenant summary、route metadata、request / audit ref、事实卡和页面状态预览。
   - 不提供写入、执行、确认、重放、发放 key 或 live backend 请求按钮。
3. 更新 `src/styles.css`。
   - 为 tenant overview、事实卡和状态预览补充响应式布局。
4. 新增 `control-plane-read-admin-tenant-overview-v1.json` 与 `check-control-plane-read-admin-tenant-overview-v1.py`。
   - 固定页面切片、契约复用、状态覆盖、禁用 live backend / write controls 和文档同步。

## 停止线

- 不把 `apps/radishmind-web/` 写成完整 formal user workspace。
- 不把 `admin-tenant-overview` 写成 production admin console ready。
- 不在页面内请求 live backend。
- 不添加数据库 schema、migration、query 或 repository。
- 不添加 `Radish` OIDC middleware、token validation、login 或 logout。
- 不添加 API key lifecycle、quota enforcement、rate limiting、billing 或 cost ledger。
- 不添加 workflow builder、executor、confirmation、writeback、replay 或 materialized result reader。
- 不修改 `apps/radishmind-console/` 的本地 ops surface 边界。

## 验证

- `npm run build`（工作目录：`apps/radishmind-web/`）
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-admin-tenant-overview-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-shared-shell-v1.py`
- `./scripts/run-python.sh scripts/run-control-plane-read-consumer-smoke.py --check`
- `./scripts/check-repo-fast.sh`

## 后续顺序

若继续 formal UI read-side，应在 `workspace-applications`、`workspace-api-keys`、`workspace-usage-quota`、`workspace-workflow-definitions`、`workspace-run-history` 和 `admin-audit-log` 中按单页切片推进；若改走真实 auth/db，则先补 repository / auth middleware 任务卡与 checker，不与页面切片并行。
