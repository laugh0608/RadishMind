# Control Plane Read Workspace Usage Quota v1 计划

更新时间：2026-05-31

## 目的

本任务卡用于记录 `control-plane-read-workspace-usage-quota-v1`：在 `apps/radishmind-web/` 的 read-only shared shell 内实现用户工作区使用量与 quota 只读页面切片 `workspace-usage-quota`。

当前切片只消费 `quota-summary-route` 的 TypeScript consumer contract 和离线 view model，展示 quota id、period、request / token / cost limit、usage snapshot、over quota failure code、route metadata、request / audit ref、只读状态和 forbidden output guard。它不请求真实后端，不接数据库、OIDC、真实 quota enforcement、rate limit、billing、cost ledger、workflow executor、confirmation、writeback 或 replay。

## 依赖

- [Control Plane Read Shared Shell v1 计划](control-plane-read-shared-shell-v1-plan.md)
- [Control Plane Read Admin Tenant Overview v1 计划](control-plane-read-admin-tenant-overview-v1-plan.md)
- [Control Plane Read Workspace Applications v1 计划](control-plane-read-workspace-applications-v1-plan.md)
- [Control Plane Read Workspace API Keys v1 计划](control-plane-read-workspace-api-keys-v1-plan.md)
- [Control Plane Read Formal UI Implementation Readiness v1 计划](control-plane-read-formal-ui-implementation-readiness-v1-plan.md)
- [Control Plane Read Formal UI Boundary v1 计划](control-plane-read-formal-ui-boundary-v1-plan.md)
- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- `contracts/typescript/control-plane-read-api.ts`

## 实现范围

1. 新增 `src/features/control-plane-read/workspaceUsageQuota.ts`。
   - 复用 `CONTROL_PLANE_READ_ROUTE_DEFINITIONS`、`CONTROL_PLANE_READ_ROUTES`、`QuotaSummary`、`ControlPlaneReadResponseByRoute` 与 `toControlPlaneReadCollectionViewModel`。
   - 只构造离线 `quota-summary-route` view model，不复制 route path 字面量。
   - 显示 quota id、period、request / token / cost limit、usage snapshot 与 over quota failure code。
   - 继续使用 `controlPlaneReadResponseHasForbiddenOutput` 拦截 forbidden output 字段。
2. 更新 `src/app/App.tsx`。
   - 在 shared shell 内新增 `workspace-usage-quota` 页面区域。
   - 展示 route metadata、request / audit ref、quota 限额、usage snapshot、状态预览和 forbidden output guard。
   - 不提供 change quota / enforce quota / price / billing ledger / cost ledger / confirm / replay / live backend 请求控件。
3. 更新 `src/styles.css`。
   - 为 quota route、limit、snapshot、failure code 和状态预览补充响应式布局。
4. 新增 `control-plane-read-workspace-usage-quota-v1.json` 与 `check-control-plane-read-workspace-usage-quota-v1.py`。
   - 固定页面切片、契约复用、只读状态覆盖、禁用 live backend / enforcement / billing controls 和文档同步。

## 停止线

- 不把 `apps/radishmind-web/` 写成完整 formal user workspace。
- 不把 `workspace-usage-quota` 写成 quota enforcement、rate limit、billing 或 cost ledger ready。
- 不在页面内请求 live backend。
- 不添加数据库 schema、migration、query 或 repository。
- 不添加 `Radish` OIDC middleware、token validation、login 或 logout。
- 不添加 API key lifecycle、quota enforcement、rate limiting、billing、cost ledger 或 workflow executor。
- 不添加 confirmation、writeback、replay 或 materialized result reader。
- 不修改 `apps/radishmind-console/` 的本地 ops surface 边界。

## 验证

- `npm run build`（工作目录：`apps/radishmind-web/`）
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-workspace-usage-quota-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-workspace-api-keys-v1.py`
- `./scripts/run-python.sh scripts/run-control-plane-read-consumer-smoke.py --check`
- `./scripts/check-repo-fast.sh`

## 后续顺序

若继续 formal UI read-side，应按 readiness 顺序推进 `workspace-workflow-definitions`、`workspace-run-history` 和 `admin-audit-log`；若改走真实 auth/db，则先补 repository / auth middleware 任务卡与 checker，不与页面切片并行。
