# Workflow Node Designer Edge Editing Save Preconditions v1 专题

更新时间：2026-06-24

状态：`workflow_node_designer_edge_editing_save_preconditions_v1_defined`

## 专题定位

`Workflow Node Designer Edge Editing Save Preconditions v1` 承接 [Workflow Node Designer Persisted Layout v1](node-designer-persisted-layout-v1.md)，定义画布连线从 preview feedback 进入 saved draft 保存链路前必须满足的边界。

本专题定义 edge mutation 的保存前置、允许字段、验证证据和实现拆分；后续已由 `Workflow Node Designer Controlled Edge Mutation Implementation v1` 完成画布新增 / 删除边的 active draft mutation。该链路不扩展 Go schema，不保存 React Flow 原始 edge 对象，不新增 backend route、repository mode、真实数据库、OIDC middleware、token validation、membership adapter、public production API、publish、run、executor、confirmation decision、writeback、replay 或 materialized result reader。

## 已知事实

- 本专题定义时，`WorkflowNodeDesigner.onConnect` 只执行 typed connection validation，并显示 `Preview only` 反馈；它不会修改 `draft.edges`，也不会触发 saved draft 保存映射。
- `Workflow Node Designer Controlled Edge Mutation Implementation v1` 已把合法 `onConnect` 升级为受控 `onAddEdge`，并补受控 `onRemoveEdge` 入口；mutation 仍只作用于 active draft。
- `validateWorkflowNodeDesignerConnection` 当前只检查 source / target 存在、不同节点、没有重复 from-to pair。
- 列表式 Draft Designer 已允许编辑 `WorkflowDraftDesignerEdge.conditionSummary`，并通过 `handleWorkflowDraftEdgeConditionChange` 写回 active draft。
- `savedWorkflowDraftConsumer` 当前保存 edge 的稳定字段为 `edge_id`、`from_node_id`、`to_node_id` 和 `condition_summary`。
- saved draft restore 当前会把 edge 恢复成 `WorkflowDraftDesignerEdge`，但 `edgeKind` 使用本地 `context` 默认值；画布 edge color / kind 仍由 node lane、risk、policy 和 audit 关系派生。
- `workflowDraftValidationInspector` 已有 `output_audit_path` 与 `orphan_node_scan`，能在 edge mutation 后暴露 output / audit 缺失和孤立节点问题。

## 允许保存的内容

后续实现如需让画布连线进入保存链路，只允许写入现有 `draft.edges` 结构：

- `edgeId`
- `fromNodeId`
- `toNodeId`
- `conditionSummary`

保存 payload 仍映射为：

```json
{
  "edge_id": "edge_context_to_model_context",
  "from_node_id": "node_context",
  "to_node_id": "node_model",
  "condition_summary": "Sanitized context flows to model reasoning."
}
```

这些字段已经属于 Saved Workflow Draft v1 主体，不需要新增 `additional_fields` schema。若未来要持久化 edge kind、handle id、port id 或 graph policy metadata，必须另开 schema 任务。

## 必须满足的保存前置

画布 edge mutation 进入实现任务前，必须同时满足：

1. mutation 只作用于当前 active draft 的已知节点。
2. 禁止 self-loop。
3. 禁止重复的 `fromNodeId -> toNodeId` pair。
4. 新增边必须生成稳定 `edgeId`，不得直接保存 React Flow generated id。
5. 删除边后必须重新运行 validation inspector，不能让孤立节点、缺 output path 或缺 audit path 被静默保存为 ready。
6. `conditionSummary` 为空时必须生成可审查默认摘要，不能保存空白条件。
7. derived edge kind 只能继续作为 UI / review state，不进入 persisted saved draft schema。
8. mutation 后必须标记 `local_edit` / `unsaved_local`，并继续复用 existing saved draft version conflict 与 no sample fallback 语义。

## 已完成实现拆分

`Workflow Node Designer Controlled Edge Mutation Implementation v1` 已按以下范围完成：

- 在 `WorkflowNodeDesigner` 中把合法 `onConnect` 从 preview feedback 升级为调用 `onAddEdge`。
- 在 `App.tsx` 中新增 active draft edge add / remove helper，复用现有 `workflowDraftEdgeId`、`workflowDraftEdgeKindForConnection` 和 `workflowDraftEdgeConditionSummary`。
- 在 Node Designer inspector connected edge 条目中提供删除边入口，并保留删除保护提示。
- 保存 / restore 继续走现有 `savedWorkflowDraftConsumer` edge mapping。
- Review Handoff 和 validation inspector 继续消费 active draft，不新增 handoff persistence。

如果实现过程中需要保存 edge kind、handle id、port id 或视觉样式，应停止当前任务，先更新本专题与 schema 任务卡。

2026-06-24 已完成 [Workflow Node Designer Controlled Edge Mutation Implementation v1 任务卡](../../task-cards/workflow-node-designer-controlled-edge-mutation-implementation-v1-plan.md)，用于承接 `onConnect` 受控新增 edge、受控删除 edge、active draft `local_edit` / `unsaved_local` 语义和专项 checker 更新。该实现任务未扩 schema、backend route、repository mode 或执行链路。

## 验收方式

- 本专题和任务卡被文档入口收录。
- 专项 checker 固定当前代码事实、保存前置、文档引用和 fast baseline 顺序。
- 本批不改运行时代码时，运行专项 checker、`git diff --check` 和 `./scripts/check-repo.sh --fast`。
- controlled edge mutation 实现已补 Web build、专项 checker 更新和 fast baseline；后续若继续扩 edge persisted schema，再补前端 / Go schema / saved draft consumer smoke。

## 停止线

- 不保存 React Flow 原始 edge 对象、handle id、viewport、selection、drag state、connection preview、visual edge style 或 derived edge kind。
- 不把画布 edge 顺序解释为 runtime execution order。
- 不把 validation overlay、Preview Plan、runtime readiness 或 Review Handoff 写成 publish ready、run ready 或 production ready。
- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、writeback、replay、resume 或 materialized result reader。
- 不接 repository mode、真实数据库、database connection provider、secret resolver、production resolver runtime、schema marker runtime、migration runner、Radish OIDC middleware、token validation、membership adapter、API key lifecycle、quota、billing 或 public production API。
