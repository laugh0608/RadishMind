# `Workflow Application Detail Read` v1 计划

更新时间：2026-06-07

## 任务目标

本任务卡用于落实 `workflow-application-detail-read-v1`，在 `apps/radishmind-web/` 的 `workspace-applications` 区域下方增加 Application detail 的离线只读 product surface。当前状态固定为 `workflow_application_detail_read_defined`。

本切片只从现有 `WorkspaceApplicationRow` / applications summary 和离线 fixture metadata 派生 application detail view model，展示 application identity、tenant / owner ref、application type、status、provider profile、risk summary、latest run ref、route / request / audit metadata，以及 blocked capability preview；不新增 Go route、真实 API、数据库、Radish OIDC、executor、confirmation decision、execution unlock、business writeback、run replay、run resume 或 production API consumer。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [阶段路线图](../radishmind-roadmap.md)
- [`Workflow / Agent Runtime Function Surface` v1 计划](workflow-agent-runtime-function-surface-v1-plan.md)
- [`Workflow Function Surface Boundary` v1 计划](workflow-function-surface-boundary-v1-plan.md)
- [`Control Plane Read Workspace Applications` v1 计划](control-plane-read-workspace-applications-v1-plan.md)

## 实现范围

1. 新增 `apps/radishmind-web/src/features/control-plane-read/workflowApplicationDetail.ts`。
   - 复用 `WorkspaceApplicationRow`、`CONTROL_PLANE_READ_ROUTE_DEFINITIONS` 和 forbidden output guard。
   - 默认展示 `app_flow_copilot` 的离线 detail view model。
2. 更新 `apps/radishmind-web/src/app/App.tsx`。
   - 在 `workspace-applications` section 下方渲染 application detail panel。
   - 显示 identity、tenant / owner、type、status、provider profile、risk summary、latest workflow / run、route / request / audit metadata 和 blocked capability preview。
3. 更新 `apps/radishmind-web/src/styles.css`。
   - 只补 `workflow-application-detail` 命名空间样式。
   - 不添加 action button、form、mutation control、execution control 或 confirmation submit control。

## 程序化证据

- `scripts/checks/fixtures/workflow-application-detail-read-v1.json`
- `scripts/checks/control_plane/check-workflow-application-detail-read-v1.py`
- `scripts/check-repo.py` fast baseline

## 验收口径

- `workflow-application-detail-read-v1` 状态固定为 `workflow_application_detail_read_defined`。
- detail surface 只从 applications summary 与离线 metadata 派生，不请求 live detail backend。
- `canRequestLiveBackend`、`canMutate`、`canCreateApplication`、`canEditApplication`、`canDeleteApplication`、`canPublishApplication`、`canStartRun`、`canExecuteWorkflow`、`canSubmitConfirmationDecision`、`canWriteBusinessTruth` 和 `canReplayRun` 必须保持 `false`。

## 停止线

- 不新增 Go route、runtime API 或 live detail backend request。
- 不实现 application create / edit / delete / publish 写能力。
- 不实现 workflow executor、node executor、tool executor、confirmation decision、execution unlock 或 agent loop。
- 不实现 business writeback、run replay、run resume、durable run store、durable result store 或 materialized result reader。
- 不实现真实数据库、Radish OIDC、token validation、production API consumer 或 production readiness。
