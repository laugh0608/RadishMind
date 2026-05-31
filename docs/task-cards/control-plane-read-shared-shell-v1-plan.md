# `Control Plane Read Shared Shell` v1 计划

更新时间：2026-05-31

## 任务目标

本任务卡用于启动正式产品 UI 的第一个实现切片：`control-plane-read-shared-shell-v1`。

当前切片创建 `apps/radishmind-web/` 的 `React + Vite + TypeScript` `shared-read-shell` 只读壳，只消费 `contracts/typescript/control-plane-read-api.ts`，渲染 route catalog、共享状态组件和 forbidden output guard。它不请求真实后端，不接数据库、OIDC、API key / quota、workflow executor、confirmation、writeback 或 replay，也不修改 `apps/radishmind-console/` 的本地 ops surface 定位。

## 输入事实源

- [Control Plane Read Formal UI Implementation Readiness v1 计划](control-plane-read-formal-ui-implementation-readiness-v1-plan.md)
- [Control Plane Read Formal UI Boundary v1 计划](control-plane-read-formal-ui-boundary-v1-plan.md)
- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- `contracts/typescript/control-plane-read-api.ts`
- `apps/radishmind-console/`

## 实现范围

- 新增 `apps/radishmind-web/`，作为未来正式产品 UI 的独立 app。
- 新增 `src/features/control-plane-read/readShell.ts`，从 TypeScript consumer contract 复用 route catalog、route definitions、collection view model 和 forbidden output guard。
- 新增 `src/app/App.tsx` 与 `src/styles.css`，只展示 route catalog、loading / ready / empty / denied / stale / partial failure / forbidden projection 状态和 forbidden output key 列表。
- 新增 `control-plane-read-shared-shell-v1.json` 与 `check-control-plane-read-shared-shell-v1.py`，将该切片接入 fast baseline。

## 非目标

- 不实现正式 user workspace 的完整页面。
- 不实现 production admin console。
- 不请求 live backend，不暴露 test-only fake auth context。
- 不接 `Radish` OIDC，不实现 auth middleware、login / logout 或 token validation。
- 不创建数据库 schema、migration、query 或 repository。
- 不实现 API key lifecycle、quota enforcement、rate limiting、billing ledger 或 cost calculation。
- 不实现 workflow builder、executor、tool executor、confirmation flow、writeback、replay 或 materialized result reader。
- 不把 `apps/radishmind-console/` 改成正式产品端。

## 验收口径

- `apps/radishmind-web/` 存在，并保持 `private=true`。
- 页面代码不得复制 route path 字符串，必须复用 `CONTROL_PLANE_READ_ROUTES` 与 `listControlPlaneReadRouteCatalog`。
- 页面必须通过 `toControlPlaneReadCollectionViewModel` 形成共享只读 view model。
- 页面必须通过 `controlPlaneReadResponseHasForbiddenOutput` 阻断 forbidden projection。
- 页面源码不得包含 `fetch(`、`XMLHttpRequest`、`axios` 或 `VITE_RADISHMIND_PLATFORM_BASE_URL`，避免把 shared shell 写成真实 API consumer。
- `scripts/checks/control_plane/check-control-plane-read-shared-shell-v1.py` 与 `scripts/run-control-plane-read-consumer-smoke.py --check` 通过。
- 仓库快速检查通过。

## 停止线

- 不把该切片写成完整 formal user workspace、production admin console、real API consumer、database ready、OIDC ready、workflow executor ready 或 production ready。
- 后续页面切片仍按 readiness 文档顺序推进：`admin-tenant-overview`、`workspace-applications`、`workspace-api-keys`、`workspace-usage-quota`、`workspace-workflow-definitions`、`workspace-run-history`、`admin-audit-log`。
