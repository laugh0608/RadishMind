# Control Plane Read Workspace Run History v1 计划

## 目的

本任务卡用于记录 `control-plane-read-workspace-run-history-v1`：在 `apps/radishmind-web/` 的 read-only shared shell 内实现用户工作区运行记录只读页面切片 `workspace-run-history`。

## 前置依赖

- [Control Plane Read-Side 契约](../contracts/control-plane-read-side.md)
- [Control Plane Read Consumer Contract v1 计划](control-plane-read-consumer-contract-v1-plan.md)
- [Control Plane Read Formal UI Boundary v1 计划](control-plane-read-formal-ui-boundary-v1-plan.md)
- [Control Plane Read Formal UI Implementation Readiness v1 计划](control-plane-read-formal-ui-implementation-readiness-v1-plan.md)
- [Control Plane Read Shared Shell v1 计划](control-plane-read-shared-shell-v1-plan.md)
- [Control Plane Read Workspace Workflow Definitions v1 计划](control-plane-read-workspace-workflow-definitions-v1-plan.md)
- `contracts/typescript/control-plane-read-api.ts`

## 实施范围

1. 在 `apps/radishmind-web/src/features/control-plane-read/` 新增 `workspaceRunHistory.ts`。
   - 只消费 `run-record-summary-list-route` 与 `RunRecordSummary`。
   - 输出离线 read-only view model，不请求 live backend。
2. 在 shared shell 内新增 `workspace-run-history` 页面区域。
   - 展示 run id、workflow definition ref、application ref、status、failure code、cost summary、trace id、started / completed timestamp。
   - 展示 route metadata、request / audit ref、cursor 和状态预览。
3. 复用现有状态组件风格、forbidden output guard 和 route catalog binding。
4. 新增 `control-plane-read-workspace-run-history-v1.json` 与 `check-control-plane-read-workspace-run-history-v1.py`。
5. 接入 `scripts/check-repo.py` fast baseline，并同步入口文档和本周周志。

## 停止线

- 不把 `workspace-run-history` 写成 workflow executor、run replay、run resume、materialized result reader 或 business writeback ready。
- 不请求 live backend，不接数据库、OIDC、repository、workflow executor、confirmation、writeback 或 replay。
- 不提供 start / cancel / resume / replay / materialize result / write business truth 控件。
- 不修改 `apps/radishmind-console/` 的本地 ops surface 定位。

## 验证

- `npm run build`（工作目录：`apps/radishmind-web/`）
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-workspace-run-history-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-workspace-workflow-definitions-v1.py`
- `./scripts/run-python.sh scripts/run-control-plane-read-consumer-smoke.py --check`
- `./scripts/check-repo-fast.sh`

## 后续顺序

若继续 formal UI read-side，应按 readiness 顺序推进 `admin-audit-log`。若改走真实 auth/db，则先补 repository / auth middleware 任务卡与 checker，不与页面切片并行。
