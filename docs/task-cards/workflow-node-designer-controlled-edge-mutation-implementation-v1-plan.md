# Workflow Node Designer Controlled Edge Mutation Implementation v1 任务卡

更新时间：2026-06-24

## 任务标识

- 切片：`workflow-node-designer-controlled-edge-mutation-implementation-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_node_designer_controlled_edge_mutation_implementation_v1_implemented`

## 目标

在 `Workflow Node Designer Edge Editing Save Preconditions v1` 已固定保存前置后，把 `WorkflowNodeDesigner.onConnect` 从 preview-only feedback 升级为受控新增 edge，并补受控删除 edge 入口。

本任务只修改前端 active draft mutation、UI 入口、专项 checker 和文档状态；不扩 Go schema，不新增 backend route，不保存 React Flow 原始 edge / handle id / port id / derived edge kind / runtime order，不打开 repository mode、真实数据库、OIDC、production API、publish、run、executor、confirmation、writeback 或 replay。

## 输入事实源

- [Workflow Node Designer Edge Editing Save Preconditions v1 专题](../features/workflow/node-designer-edge-editing-save-preconditions-v1.md)
- [Workflow Node Designer Persisted Layout v1 专题](../features/workflow/node-designer-persisted-layout-v1.md)
- [Workflow Node Designer Saved Draft Mapping v1 专题](../features/workflow/node-designer-saved-draft-mapping-v1.md)
- [Workflow Review Handoff Active Draft v1 专题](../features/workflow/review-handoff-active-draft-v1.md)

## 实现范围

- `WorkflowNodeDesigner.onConnect` 只把合法 source / target draft node id 传给受控 `onAddEdge` callback。
- `App.tsx` 在 active draft 上新增 edge，并复用 `workflowDraftEdgeId`、`workflowDraftEdgeKindForConnection` 和 `workflowDraftEdgeConditionSummary` 生成稳定 edge。
- 删除 edge 通过受控 `onRemoveEdge` callback 按 active draft 的 `edgeId` 删除，不消费 React Flow raw edge object。
- mutation 后统一标记 `local_edit` / `unsaved_local`，让 saved draft consumer、validation inspector、Preview Plan 和 Review Handoff 继续消费 active draft。
- 保存 / restore 继续复用 `savedWorkflowDraftConsumer` 现有 edge mapping。

## 验收口径

- 新增连线写入 `draft.edges` 的字段只限 `edgeId`、`fromNodeId`、`toNodeId`、`edgeKind` 和 `conditionSummary`；保存 payload 仍只输出 `edge_id`、`from_node_id`、`to_node_id`、`condition_summary`。
- 重复 from-to pair、self-loop、未知节点必须被拒绝。
- 新增 edge 的 `conditionSummary` 不为空。
- 删除 edge 后不补隐式 fallback edge，validation inspector 继续暴露 orphan / audit path 问题。
- 专项 checker 更新为实现事实，并继续禁止 React Flow raw edge、handle id、port id、derived edge kind、runtime order 和 executor / production 能力声明。

## 验证要求

- `npm run build`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-node-designer-edge-editing-save-preconditions-v1.py`
- `git diff --check`
- `./scripts/check-repo.sh --fast`

## 停止线

- 不扩 saved draft schema、Go document type 或 backend route。
- 不保存 React Flow 原始 edge 对象、handle id、port id、viewport、selection、connection preview、visual edge style、derived edge kind 或 runtime order。
- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、writeback、replay、repository mode、真实数据库、OIDC middleware、token validation、membership adapter 或 public production API。
