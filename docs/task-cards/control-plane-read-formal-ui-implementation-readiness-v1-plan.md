# `Control Plane Read Formal UI Implementation Readiness` v1 计划

更新时间：2026-05-28

## 任务目标

本任务卡用于把 `control-plane-read-formal-ui-boundary-v1` 之后、正式 React 页面实现之前的工程落点写清楚：未来正式只读产品 UI 应落在哪个 app 边界内，页面按什么顺序实现，如何复用 `contracts/typescript/control-plane-read-api.ts`，以及实现前必须保留哪些测试策略和停止线。

当前任务只定义 `control-plane-read-formal-ui-implementation-readiness-v1`。它不创建 React 页面，不创建 `apps/radishmind-web/`，不修改 `apps/radishmind-console/`，不请求真实后端，也不把 fake-store-backed read route 写成生产管理端或正式用户端。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- [Control Plane Read Formal UI Boundary v1 计划](control-plane-read-formal-ui-boundary-v1-plan.md)
- `contracts/typescript/control-plane-read-api.ts`
- `scripts/run-control-plane-read-consumer-smoke.py`
- `services/platform/README.md`
- `scripts/checks/fixtures/product-surface-v1-boundary.json`
- `scripts/checks/fixtures/control-plane-read-consumer-contract-v1.json`
- `scripts/checks/fixtures/control-plane-read-formal-ui-boundary-v1.json`

## 工程落点

未来正式只读产品 UI 的首选落点固定为 `apps/radishmind-web/`，类型为 `React + Vite + TypeScript` product UI。首批 read-side 页面建议收口在未来 `apps/radishmind-web/src/features/control-plane-read/`，app shell 和路由组合放在未来 `apps/radishmind-web/src/app/`。

`apps/radishmind-console/` 继续保留为 `P3 Local Product Shell / Ops Surface`，只作为本地 ops surface。它可以贡献设计语言、语义 token 和只读状态处理经验，但不能被改写为正式 user workspace、production admin console、production hosting package 或 control plane write surface。

`services/platform/` 在该 readiness 中仍只是 fake-store-backed read route host。本文档不新增后端 route，不接数据库 query，不接 OIDC middleware，不接 API key lifecycle、quota enforcement、workflow executor、confirmation、writeback、replay 或部署行为。

## App 边界

- `apps/radishmind-web/`：未来正式产品 UI 预留 app，不在本切片创建；首批只允许承载 `Admin Control Plane` 与 `User Workspace` 的 read-only route catalog、collection view、loading / empty / denied / stale / partial failure / forbidden projection 状态。
- `apps/radishmind-console/`：当前本地 ops console，只展示 overview、local-smoke、provider/profile、session/tooling、stop-line 和诊断信息；不得变成正式用户端或生产管理端。
- `contracts/typescript/control-plane-read-api.ts`：正式 UI 与离线 smoke 的唯一 TypeScript consumer contract 来源。
- `scripts/run-control-plane-read-consumer-smoke.py --check`：正式 UI 实现前的离线消费基线，不请求真实服务。

## 页面实现顺序

1. `shared-read-shell`
   - 先建立未来产品 app shell、route catalog binding、共享状态组件和 forbidden output guard。
   - 必须复用 `CONTROL_PLANE_READ_ROUTES`、`CONTROL_PLANE_READ_ROUTE_DEFINITIONS`、`ControlPlaneReadCollectionViewModel`、`listControlPlaneReadRouteCatalog`、`toControlPlaneReadCollectionViewModel` 和 `controlPlaneReadResponseHasForbiddenOutput`。
2. `admin-tenant-overview`
   - 消费 `tenant-summary-route`。
   - 优先验证单资源 envelope、denied state 和 audit ref 展示。
3. `workspace-applications`
   - 消费 `application-summary-list-route`。
   - 先验证 cursor list、filter 和 pagination 复用。
4. `workspace-api-keys`
   - 消费 `api-key-summary-list-route`。
   - 必须在 forbidden output guard 已存在后实现；只显示 summary、状态、scope、owner、时间和 audit ref。
5. `workspace-usage-quota`
   - 消费 `quota-summary-route`。
   - 只展示 usage/quota summary，不实现 quota enforcement、billing 或 cost ledger。
6. `workspace-workflow-definitions`
   - 消费 `workflow-definition-summary-list-route`。
   - 不创建、编辑、执行或确认 workflow。
7. `workspace-run-history`
   - 消费 `run-record-summary-list-route`。
   - 不启动、取消、恢复、replay run，也不读取 materialized result。
8. `admin-audit-log`
   - 消费 `audit-summary-list-route`。
   - 在各资源页保留 `audit_ref` 与 failure state 后，再补 audit 过滤和只读列表。

## Consumer Contract 复用

未来页面必须从 `contracts/typescript/control-plane-read-api.ts` 复用以下来源：

- `CONTROL_PLANE_READ_ROUTES`
- `CONTROL_PLANE_READ_ROUTE_DEFINITIONS`
- `CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS`
- `ControlPlaneReadCollectionViewModel`
- `ControlPlaneReadRouteCatalogViewModel`
- `ControlPlaneReadRouteId`
- `ControlPlaneReadResponseByRoute`
- `isControlPlaneReadEnvelope`
- `listControlPlaneReadRouteCatalog`
- `toControlPlaneReadCollectionViewModel`
- `controlPlaneReadResponseHasForbiddenOutput`

页面不得复制 route path 字符串，不得绕过 view model 直接解释 response fixture，不得在页面层自写另一套 envelope parser。`denied` 页面必须渲染空 items 并保留 `failureCode` 与 `auditRef`；`stale` 页面只能保留上一份只读数据和诊断时间；`forbidden_projection` 必须先由 forbidden output guard 阻断敏感字段。

## 测试策略

当前 readiness 切片必须通过：

- `python scripts/checks/control_plane/check-control-plane-read-formal-ui-implementation-readiness-v1.py`
- `python scripts/run-control-plane-read-consumer-smoke.py --check`
- `pwsh ./scripts/check-repo.ps1 -Fast`

未来真正创建 React 页面后，还必须补：

- `apps/radishmind-web/` 下的 `npm run build`
- 每个页面的 loading、ready、empty、denied、stale、partial failure 和 forbidden projection 状态测试
- 桌面 / 窄屏视觉 smoke、文本溢出检查、禁止 action 缺失检查和敏感字段缺失检查

## 验收口径

- `scripts/checks/fixtures/control-plane-read-formal-ui-implementation-readiness-v1.json` 固定工程落点、app 边界、页面实现顺序、consumer contract 复用方式、测试策略和停止线。
- `scripts/checks/control_plane/check-control-plane-read-formal-ui-implementation-readiness-v1.py` 校验依赖 gate、app 边界、页面顺序、route 覆盖、TypeScript contract 复用、测试策略、文档引用和 fast baseline 接入。
- checker 接入 `scripts/check-repo.py --fast`。
- 文档同步说明该切片只是正式 UI 实现 readiness，不代表 React UI、production admin console、real API consumer、database、OIDC 或 production ready。

## 非目标

- 不创建 `apps/radishmind-web/`。
- 不实现正式 React 页面、正式 user workspace UI 或 production admin console。
- 不修改 `apps/radishmind-console/`，不把本地 ops console 升级成 production admin console。
- 不启动服务、不调用真实 route、不暴露 test-only fake auth context。
- 不实现数据库 schema、migration、query、durable store 或 repository。
- 不实现 `Radish` OIDC middleware、token validation、login / logout route 或 production auth policy。
- 不实现 API key lifecycle、quota enforcement、rate limiting、billing ledger 或 cost calculation。
- 不实现 workflow builder、workflow executor、tool executor、confirmation flow、writeback、replay 或 materialized result reader。

## 停止线

- 不把 `control-plane-read-formal-ui-implementation-readiness-v1` 写成 React UI ready、formal user workspace UI ready、production admin console ready、real API consumer ready、database ready、OIDC ready 或 production ready。
- 不让 `apps/radishmind-console/` 承载正式用户端、生产管理端或 control plane write surface。
- 不让任何未来页面暗示可以 create、edit、delete、issue、reveal、rotate、revoke、execute、confirm、writeback、materialize result 或 replay。
- 不把 offline fixture、fake store 或 test-only fake auth context 写成正式 UI 数据源。
