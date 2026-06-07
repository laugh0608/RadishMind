# `Workflow Definition Detail Read` v1 计划

更新时间：2026-06-07

## 任务目标

本任务卡用于落实 `workflow-definition-detail-read-v1`，在 `apps/radishmind-web/` 的 `workspace-workflow-definitions` 区域下方增加 workflow definition detail 的离线只读 surface。当前状态固定为 `workflow_definition_detail_read_defined`。

本切片只展示 nodes、edges、input/output summary、risk level、requires confirmation capability、blocked action preview、route / request / audit metadata；不新增 Go route、真实 API、数据库、Radish OIDC、workflow builder、workflow executor、tool executor、confirmation decision、business writeback、run replay、run resume 或 production API consumer。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [`Workflow Function Surface Boundary` v1 计划](workflow-function-surface-boundary-v1-plan.md)
- [`Control Plane Read Workspace Workflow Definitions` v1 计划](control-plane-read-workspace-workflow-definitions-v1-plan.md)
- `scripts/checks/fixtures/workflow-function-surface-boundary-v1.json`
- `scripts/checks/fixtures/control-plane-read-workspace-workflow-definitions-v1.json`

## 实现边界

1. 新增 `apps/radishmind-web/src/features/control-plane-read/workflowDefinitionDetail.ts`。
   - 复用 `WorkflowDefinitionSummary`、`CONTROL_PLANE_READ_ROUTE_DEFINITIONS` 和 forbidden output guard。
   - 默认展示 `wf_radishflow_copilot_latest` 的离线 detail view model。
2. 更新 `apps/radishmind-web/src/app/App.tsx`。
   - 在 `workspace-workflow-definitions` section 下方渲染 detail panel。
   - 显示 node list、edge list、input/output summary 和 blocked action preview。
3. 更新 `apps/radishmind-web/src/styles.css`。
   - 只补 `workflow-definition-detail` 命名空间样式。
   - 不添加 action button、form、mutation control 或 execution control。

## 程序化证据

- `scripts/checks/fixtures/workflow-definition-detail-read-v1.json`
- `scripts/checks/control_plane/check-workflow-definition-detail-read-v1.py`
- `scripts/check-repo.py` fast baseline

## 停止线

- 不新增 runtime API 或 live detail backend request。
- 不实现 workflow builder、definition mutation、workflow executor、node executor 或 tool executor。
- 不实现 confirmation decision、business writeback、run replay、run resume、durable run store 或 materialized result reader。
- 不实现真实数据库、Radish OIDC、production API consumer 或 production readiness。
