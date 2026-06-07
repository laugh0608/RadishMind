# `Workflow Run Detail Read` v1 计划

更新时间：2026-06-07

## 任务目标

本任务卡用于落实 `workflow-run-detail-read-v1`，把 `Workflow / Agent Runtime Function Surface v1` 中的 run detail 从字段边界推进到 `apps/radishmind-web/` 内的离线只读产品 surface。当前状态固定为 `workflow_run_detail_read_defined`。

本切片只展示 run detail、state timeline、input / output summary、cost / token snapshot、trace id、failure code、audit refs、blocked materialized result reader 和 blocked replay / resume preview；不新增 live detail backend、runtime API、workflow executor、confirmation decision、business writeback、run replay、run resume、durable run store、materialized result reader、真实数据库、Radish OIDC 或 production API consumer。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [阶段路线图](../radishmind-roadmap.md)
- [`Workflow / Agent Runtime Function Surface` v1 计划](workflow-agent-runtime-function-surface-v1-plan.md)
- [`Workflow Function Surface Boundary` v1 计划](workflow-function-surface-boundary-v1-plan.md)
- [`Workflow Definition Detail Read` v1 计划](workflow-definition-detail-read-v1-plan.md)
- [`Control Plane Read Workspace Run History` v1 计划](control-plane-read-workspace-run-history-v1-plan.md)

## 实现范围

1. 在 `apps/radishmind-web/src/features/control-plane-read/workflowRunDetail.ts` 中新增离线只读 detail view model。
2. 复用 `WorkspaceRunRecordRow`，默认展示 `run_radishflow_copilot_20260531_001`。
3. 在 `App.tsx` 的 `workspace-run-history` 区域下方新增 detail panel。
4. 展示 `stateTimeline`、`inputSummary`、`outputSummary`、`costSummary`、`tokenSummary`、`traceId`、`failureCode`、`auditRefs`、`blockedResultPreview` 和 `blockedReplayPreview`。
5. 样式复用现有 read-only card 风格，只补必要 class，不添加按钮、表单、执行控件或写操作入口。

## 程序化证据

- `scripts/checks/fixtures/workflow-run-detail-read-v1.json`
- `scripts/checks/control_plane/check-workflow-run-detail-read-v1.py`
- `scripts/check-repo.py` fast baseline

## 验收口径

- `workflow-run-detail-read-v1` 状态固定为 `workflow_run_detail_read_defined`。
- detail surface 只从现有 run summary / offline view model 派生，不请求 live detail backend。
- 页面只能解释 run state、trace、failure、cost / token 摘要和 blocked guard。
- `canRequestLiveBackend`、`canMutate`、`canStartRun`、`canCancelRun`、`canResumeRun`、`canReplayRun`、`canMaterializeResult` 和 `canWriteBusinessTruth` 必须保持 `false`。

## 停止线

- 不新增 Go route、runtime API 或 dev-live detail route。
- 不实现 workflow executor、node executor、tool executor 或 agent loop。
- 不实现 run replay、run resume、materialized result reader、durable run store 或 durable result store。
- 不实现 confirmation decision、business writeback、真实数据库、Radish OIDC、token validation 或 production API consumer。
