# Control Plane Read Formal UI Readiness Close v1 计划

## 目的

本任务卡用于记录 `control-plane-read-formal-ui-readiness-close-v1`：在 `admin-audit-log` 完成后，对 `apps/radishmind-web/` 内当前 read-side formal UI 页面集合做聚合收口，固定七个只读页面与 read route、consumer view model、状态预览、request / audit ref 和 forbidden output guard 的一致性。

## 前置依赖

- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- [Control Plane Read Consumer Contract v1 计划](control-plane-read-consumer-contract-v1-plan.md)
- [Control Plane Read Formal UI Boundary v1 计划](control-plane-read-formal-ui-boundary-v1-plan.md)
- [Control Plane Read Formal UI Implementation Readiness v1 计划](control-plane-read-formal-ui-implementation-readiness-v1-plan.md)
- [Control Plane Read Shared Shell v1 计划](control-plane-read-shared-shell-v1-plan.md)
- [Control Plane Read Admin Audit Log v1 计划](control-plane-read-admin-audit-log-v1-plan.md)
- `contracts/typescript/control-plane-read-api.ts`

## 实施范围

1. 新增 `control-plane-read-formal-ui-readiness-close-v1.json`，以 surface matrix 固定七个页面：
   - `admin-tenant-overview`
   - `admin-audit-log`
   - `workspace-applications`
   - `workspace-api-keys`
   - `workspace-usage-quota`
   - `workspace-workflow-definitions`
   - `workspace-run-history`
2. 新增 `check-control-plane-read-formal-ui-readiness-close-v1.py`。
   - 校验页面到 `CONTROL_PLANE_READ_ROUTES` / `CONTROL_PLANE_READ_ROUTE_DEFINITIONS` 的绑定。
   - 校验 `ready`、`empty`、`denied`、`stale`、`partial_failure` 与 `forbidden_projection` 状态覆盖。
   - 校验每个页面仍通过 `toControlPlaneReadCollectionViewModel` 和 `controlPlaneReadResponseHasForbiddenOutput` 消费离线 contract。
   - 校验 `apps/radishmind-console/` 仍是本地 ops surface。
3. 接入 `scripts/check-repo.py` fast baseline，并同步入口文档、契约文档、能力矩阵、路线图、web README、scripts README 和本周周志。

## 停止线

- 不新增页面功能，不把普通 read-only 展示页继续逐项拆成专项 gate。
- 不请求 live backend，不接数据库、OIDC、repository、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
- 不把 `apps/radishmind-web/` 写成 production admin console 或完整 formal user workspace。
- 不把 `apps/radishmind-console/` 改成正式产品端或 production admin console。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-formal-ui-readiness-close-v1.py`
- `npm run build`（工作目录：`apps/radishmind-web/`）
- `./scripts/run-python.sh scripts/run-control-plane-read-consumer-smoke.py --check`
- `./scripts/check-repo-fast.sh`

## 后续顺序

read-side formal UI 页面集合进入 close 后，下一步应在 dev-only live read consumer 与真实 auth/store 前置任务之间单线选择。若放宽“不请求 live backend”，只能先连接现有 fake-store-backed handler 与测试身份上下文；真实数据库、Radish OIDC、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 和 replay 仍保持停止线。
