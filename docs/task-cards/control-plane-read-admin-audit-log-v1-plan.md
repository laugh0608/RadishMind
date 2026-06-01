# Control Plane Read Admin Audit Log v1 计划

## 目的

本任务卡用于记录 `control-plane-read-admin-audit-log-v1`：在 `apps/radishmind-web/` 的 read-only shared shell 内实现管理端审计日志只读页面切片 `admin-audit-log`。

## 前置依赖

- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- [Control Plane Read Consumer Contract v1 计划](control-plane-read-consumer-contract-v1-plan.md)
- [Control Plane Read Formal UI Boundary v1 计划](control-plane-read-formal-ui-boundary-v1-plan.md)
- [Control Plane Read Formal UI Implementation Readiness v1 计划](control-plane-read-formal-ui-implementation-readiness-v1-plan.md)
- [Control Plane Read Shared Shell v1 计划](control-plane-read-shared-shell-v1-plan.md)
- [Control Plane Read Workspace Run History v1 计划](control-plane-read-workspace-run-history-v1-plan.md)
- `contracts/typescript/control-plane-read-api.ts`

## 实施范围

1. 在 `apps/radishmind-web/src/features/control-plane-read/` 新增 `adminAuditLog.ts`。
   - 只消费 `audit-summary-list-route` 与 `AuditSummary`。
   - 输出离线 read-only view model，不请求 live backend。
2. 在 shared shell 内新增 `admin-audit-log` 页面区域。
   - 展示 audit ref、actor、event kind、resource、decision、failure code、trace id 和 recorded timestamp。
   - 展示 route metadata、request / audit ref、cursor 和状态预览。
3. 复用现有状态组件风格、forbidden output guard 和 route catalog binding。
4. 新增 `control-plane-read-admin-audit-log-v1.json` 与 `check-control-plane-read-admin-audit-log-v1.py`。
5. 接入 `scripts/check-repo.py` fast baseline，并同步入口文档和本周周志。

## 停止线

- 不把 `admin-audit-log` 写成 durable audit store、raw audit payload export 或 audit record mutation ready。
- 不请求 live backend，不接数据库、OIDC、repository、durable audit store、workflow executor、confirmation、writeback 或 replay。
- 不提供 edit / delete / raw payload export / reveal secret 控件。
- 不修改 `apps/radishmind-console/` 的本地 ops surface 定位。

## 验证

- `npm run build`（工作目录：`apps/radishmind-web/`）
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-admin-audit-log-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-workspace-run-history-v1.py`
- `./scripts/run-python.sh scripts/run-control-plane-read-consumer-smoke.py --check`
- `./scripts/check-repo-fast.sh`

## 后续顺序

`admin-audit-log` 是当前 read-side UI 页面集合的最后一个优先页面。完成后应进入 read-side UI 聚合收口，把普通展示页从逐页专项 checker 转为 surface matrix / 聚合 checker；若改走真实 auth/db，则先补 repository / auth middleware 任务卡与 checker，不与页面切片并行。
