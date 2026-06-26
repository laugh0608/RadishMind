# Workflow Node Designer Edge Editing Save Preconditions v1 任务卡

更新时间：2026-06-24

## 任务标识

- 切片：`workflow-node-designer-edge-editing-save-preconditions-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_node_designer_edge_editing_save_preconditions_v1_defined`

## 目标

在 `Workflow Node Designer Persisted Layout v1` 已完成后，固定画布连线新增 / 删除进入 saved draft 保存链路之前的准入条件。

本任务卡只承接设计前置、现有代码事实、后续实现拆分、专项 fixture / checker 和文档同步。不修改运行时代码，不实现画布新增边 / 删除边，不扩 Go schema，不新增 backend route、repository mode、数据库、OIDC、public production API、publish、run、executor、confirmation decision、writeback 或 replay。

## 输入事实源

- [Workflow Node Designer Edge Editing Save Preconditions v1 专题](../features/workflow/node-designer-edge-editing-save-preconditions-v1.md)
- [Workflow Node Designer Persisted Layout v1 专题](../features/workflow/node-designer-persisted-layout-v1.md)
- [Workflow Node Designer Surface v1 专题](../features/workflow/node-designer-surface-v1.md)
- [Workflow Node Designer Saved Draft Mapping v1 专题](../features/workflow/node-designer-saved-draft-mapping-v1.md)

## 本轮输出

- 定义 edge mutation 保存前置：已知节点、非 self-loop、无重复 from-to pair、稳定 edge id、非空 condition summary、validation inspector 消费和 local edit 状态。
- 明确允许保存字段只限 `edgeId`、`fromNodeId`、`toNodeId` 和 `conditionSummary`。
- 明确不保存 React Flow 原始 edge、handle id、port id、visual style、connection preview、derived edge kind 或 runtime order。
- 明确后续实现应把 `WorkflowNodeDesigner.onConnect` 从 preview feedback 升级为受控 `onAddEdge`，并通过 active draft `draft.edges` 进入现有 saved draft consumer。
- 新增 `workflow-node-designer-edge-editing-save-preconditions-v1` fixture / checker，并接入 fast baseline。

## 后续实现准入

实现任务卡只有在以下证据仍成立时才能打开：

- `savedWorkflowDraftConsumer` 保存 edge endpoints 和 `condition_summary` 的字段没有漂移。
- `WorkflowNodeDesigner.onConnect` 当前仍是 preview-only，后续改动必须显式新增 mutation callback。
- `workflowDraftValidationInspector` 仍能在 edge mutation 后暴露 output / audit path 和 orphan node 问题。
- Review Handoff 继续消费 active draft，不创建独立 handoff persistence。
- Web build 和 fast baseline 可作为实现后的必要验证。

## 验收口径

- 专题、任务卡、入口文档和周志已同步。
- 专项 checker 能固定当前代码事实和后续实现停止线。
- `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不实现画布 edge mutation runtime。
- 不新增 saved draft edge schema、Go document type 或 backend route。
- 不保存 React Flow 原始 edge 对象、handle id、port id、viewport、selection、connection preview、visual edge style 或 derived edge kind。
- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、writeback、replay、repository mode、真实数据库、OIDC middleware、token validation、membership adapter 或 public production API。
