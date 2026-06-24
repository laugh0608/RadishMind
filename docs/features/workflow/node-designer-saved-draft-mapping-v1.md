# Workflow Node Designer Saved Draft Mapping v1 专题

更新时间：2026-06-24

状态：`workflow_node_designer_persisted_layout_v1_implemented`

## 专题定位

`Workflow Node Designer Saved Draft Mapping v1` 承接 [Workflow Node Designer Surface v1](node-designer-surface-v1.md) 与 [Workflow Node Designer Surface Implementation v1](../../task-cards/workflow-node-designer-surface-implementation-v1-plan.md)，固定节点画布状态进入 saved draft 前的映射边界。

本专题定义画布 layout metadata、typed edge kind、validation overlay 与 saved draft / Review Handoff 的消费关系。当前已通过 [Workflow Node Designer Persisted Layout v1](node-designer-persisted-layout-v1.md) 把节点坐标保存为受控 UI metadata；它仍不保存 React Flow 原始对象、不新增 backend route、不打开 repository mode、真实数据库、OIDC middleware、token validation、membership adapter、publish、run、executor、confirmation decision、writeback、replay 或 public production API。

## 当前事实

- `WorkflowNodeDesigner` 已把 active draft 派生为 React Flow `nodes` / `edges`，并提供节点拖拽、选中、typed handle、连线校验反馈、custom node / edge 和 inspector bridge。
- saved draft consumer 当前保存节点属性、contract fields、output mapping、provider / tool / RAG refs、edge endpoints 和 `condition_summary`。
- saved draft restore 当前会读取 `additional_fields.designer_layout_v1`，缺失或非法时继续由 lane / index 派生默认布局。
- saved draft persisted edge 当前不保存 React Flow edge kind；restore 后 local draft edge kind 使用 `context`，画布 edge kind 再由 source / target / policy / audit 关系派生。
- HTTP payload 已使用 `additional_fields.designer_layout_v1` 保存 layout metadata，并由 Go `SavedWorkflowDraft.AdditionalFields`、HTTP document conversion、layout 规范化和嵌套 forbidden field scan 保留边界。
- `Workflow Node Designer Saved Draft Mapping Implementation v1` 已实现 UI-only layout：拖拽节点会回写 active draft session，Save Draft 前展示 mapping summary。
- `Workflow Node Designer Review Handoff v1` 已实现 Review Handoff 消费：`nodeDesignerReviewRecord` 汇总 layout、validation overlay、inspector state 和 saved draft mapping，不创建独立 handoff persistence。
- `Workflow Node Designer Persisted Layout v1` 已实现：Save Draft 写入受控节点坐标，Restore 恢复 `saved_draft_metadata` layout，专项 checker 已接入 fast baseline。
- `Workflow Node Designer Controlled Edge Mutation Implementation v1` 已实现：合法画布连接会新增 active draft edge，inspector connected edge 条目可删除 active draft edge；两者都会保持 `local_edit` / `unsaved_local` 语义。
- `Workflow Node Designer Validation Overlay Navigation v1` 已实现：finding focus 和高亮只存在于画布 UI，不进入 saved draft payload、layout metadata 或 Review Handoff persistence。

## 映射结论

v1 保持 active draft 为业务真相源，React Flow state 只能作为 UI view model。

本阶段不把 React Flow `nodes` / `edges` JSON 直接保存为长期 schema，也不把 viewport、selection、drag state 或 connection preview 写入 saved draft 主体。用户手工布局只通过 `additional_fields.designer_layout_v1` 保存受控节点坐标，不保存 React Flow 原始对象。

`edge kind` v1 继续作为 derived UI state：

- `WorkflowDraftDesignerEdge.edgeKind` 可用于本地 review / validation / plan preview。
- saved draft persisted edge 继续以 `from_node_id`、`to_node_id` 和 `condition_summary` 为稳定字段。
- `data_edge`、`control_edge`、`guard_edge`、`audit_edge` 只能由 node type、lane、requires confirmation、risk 和 condition summary 派生。
- 如果后续确实需要持久化 edge kind，必须先定义 schema 字段、restore 兼容策略、fixture、checker 和 migration 行为。

## 受控 Edge Mutation 映射

画布新增 / 删除连线的业务真相源仍是 active draft 的 `draft.edges`。新增连线时，前端只在校验通过后生成或复用稳定 `edgeId`、`fromNodeId`、`toNodeId`、transient `edgeKind` 和 reviewable `conditionSummary`；保存时 `savedWorkflowDraftConsumer` 仍只把 edge 映射为 `edge_id`、`from_node_id`、`to_node_id` 和 `condition_summary`。

删除连线时，Node Designer inspector 调用 active draft edge removal；validation inspector、Preview Plan、Runtime Readiness 和 Review Handoff 在下一次 render 中消费 mutation 后的 active draft。删除操作不会写入 deletion log、runtime order、React Flow edge、handle id、port id 或额外 `additional_fields`。

validation overlay focus 不参与保存映射。用户点击 finding 只改变画布高亮和当前 inspector 选择，不改变 `draft.edges`、`designer_layout_v1` 或 saved draft record。

## 已完成实现

`Workflow Node Designer Saved Draft Mapping Implementation v1` 与 `Workflow Node Designer Persisted Layout v1` 已完成：

- 将画布节点位置保存为本地 UI-only state，并在同一 active draft session 内恢复。
- 让保存前的 graph adapter 输出稳定、可审查的 mapping summary，明确哪些字段会进入 saved draft，哪些字段只属于画布视图。
- 在 saved draft restore 后优先消费 `additional_fields.designer_layout_v1`，缺失、非法或未知节点时继续按 lane / node order 生成默认布局。
- 在 Node Designer mapping summary 中展示 layout metadata 来源，避免 reviewer 把视觉位置理解为运行顺序。
- 在 Review Handoff 中展示 node designer review handoff，避免 reviewer 把 validation overlay、inspector state 或 derived edge kind 理解为 persisted runtime state。
- 用专项 fixture / checker 固定 schema、前端映射、Go domain / HTTP 保留、嵌套 forbidden field 和 fast baseline 顺序。

当前跨会话 layout metadata 固定为：

```json
{
  "additional_fields": {
    "designer_layout_v1": {
      "layout_version": "designer_layout_v1",
      "source": "workflow_node_designer",
      "persistence": "saved_draft_metadata",
      "nodes": [
        {
          "node_id": "draft_node_id",
          "x": 0,
          "y": 0,
          "pinned": false
        }
      ]
    }
  }
}
```

该候选只能保存 UI layout，不得保存 secret、token、provider response、tool result、run input / output、confirmation decision、writeback payload、executor result 或 materialized result。

## Restore 兼容策略

- 缺少 `designer_layout_v1` 时，使用 lane / index 派生默认布局。
- layout 中出现未知 `node_id` 时忽略对应 layout entry，不拒绝草案主体。
- layout 缺少已存在节点时，对缺失节点使用默认位置。
- layout 坐标必须有数字范围预算；保存与恢复都限制在 `-10000..10000`，不拒绝草案主体。
- edge kind 缺失时继续由 active draft 派生；不得因为缺 edge kind 而拒绝读取旧草案。

## 验收方式

本专题定义与 UI-only 实现阶段已完成：

- 文档入口收录本专题。
- UI-only layout implementation 已落地；当时不新增 schema、route、fixture、checker 或 runtime artifact。
- `./scripts/check-repo.sh --fast` 通过。

persisted layout implementation 阶段已完成：

- 已新增 task card、保存 / 读取 / 兼容 restore fixture、forbidden field negative case 和 checker。
- 已运行 web build、Go tests、专项 checker 和 fast baseline。
- 若扩展 edge kind persisted schema，必须同步 Go document type、web consumer mapping、schema artifact、version conflict 用例和 no sample fallback 验证。

## 停止线

- 不把 React Flow 原始 `nodes` / `edges` JSON 作为 RadishMind 长期 persisted schema。
- 不把 layout、viewport、selection 或 visual edge style 解释为运行顺序、执行计划或 runtime binding。
- 不把 `edge kind` schema 扩展塞进普通 UI 整理；需要持久化时必须单独 task card。
- 不实现 publish、run、workflow executor、node executor、tool executor、agent loop、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不接 repository mode、真实数据库、database connection provider、secret resolver、production resolver runtime、schema marker runtime、migration runner、Radish OIDC middleware、token validation、membership adapter、API key lifecycle、quota、billing 或 public production API。
