# Workflow Node Designer Persisted Layout v1 专题

更新时间：2026-06-24

状态：`workflow_node_designer_persisted_layout_v1_implemented`

## 专题定位

`Workflow Node Designer Persisted Layout v1` 承接 [Workflow Node Designer Saved Draft Mapping v1](node-designer-saved-draft-mapping-v1.md) 与 [Workflow Node Designer Review Handoff v1](node-designer-review-handoff-v1.md)，把节点画布手工位置从 active draft session 推进为 saved draft 的受控 UI metadata。

本专题只允许保存节点位置 metadata，不保存 React Flow 原始 `nodes` / `edges` JSON，不保存 viewport、selection、drag state、connection preview、visual edge style 或 derived edge kind。它不创建新 backend route、repository mode、真实数据库、OIDC middleware、token validation、membership adapter、public production API、publish、run、executor、confirmation decision、writeback、replay 或 materialized result reader。

## 当前实现

- `WorkflowDraftDesignerLayout.persistence` 已扩展为 `ui_only | saved_draft_metadata`。
- `savedWorkflowDraftConsumer` 在 Save / Validate Draft payload 中写入 `additional_fields.designer_layout_v1`，只包含 `layout_version`、`source`、`persistence` 和节点坐标数组。
- saved draft restore 会读取 `additional_fields.designer_layout_v1`，只接受版本、source、persistence 和 `pinned: false` 的节点位置；未知节点会被忽略，缺失节点继续使用 lane / index 默认布局。
- Go `SavedWorkflowDraft` domain record 已保留 `AdditionalFields`，HTTP document 会随 save / read 响应返回该 metadata。
- Go 保存链路会规范化 designer layout：只保留当前 draft 已知节点、坐标四舍五入并限制在 `-10000..10000`，`pinned` 固定为 `false`。
- Go 校验会递归扫描 `additional_fields` 中的敏感字段，嵌套出现 secret、token、run output、confirmation decision、writeback payload 或 executor result 时直接返回 `draft_payload_invalid`。
- Node Designer mapping summary 与 Review Handoff 已同步区分 `active draft session` 与 `saved draft metadata` 来源，但 Review Handoff 本身仍不持久化、导出或发送记录。

## Schema

`additional_fields.designer_layout_v1` 的 v1 形状固定为：

```json
{
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
```

字段口径：

- `layout_version` 固定为 `designer_layout_v1`。
- `source` 固定为 `workflow_node_designer`。
- `persistence` 固定为 `saved_draft_metadata`。
- `nodes[].node_id` 必须引用当前 saved draft 的节点。
- `nodes[].x` 与 `nodes[].y` 是画布坐标，保存 / 恢复时限制在 `-10000..10000`。
- `nodes[].pinned` v1 固定为 `false`，不打开 pinning 功能。

## Restore 兼容策略

- 缺少 `additional_fields` 或缺少 `designer_layout_v1` 时，继续按 lane / node order 派生默认布局。
- `designer_layout_v1` 版本、source 或 persistence 不匹配时忽略 layout metadata，不拒绝草案主体。
- layout 中出现未知 `node_id` 时忽略该节点坐标。
- layout 缺少现有节点时，只对缺失节点使用默认位置。
- 坐标非数字或非法时忽略对应节点坐标；合法坐标会在保存和恢复两侧进行限幅。
- 旧 saved draft 没有 layout metadata 时仍可读取、校验、保存和审查。

## 后续衔接

`Workflow Node Designer Edge Editing Save Preconditions v1` 与 `Workflow Node Designer Controlled Edge Mutation Implementation v1` 已承接下一段 Builder 体验：画布连线新增 / 删除只能写入 active draft `draft.edges` 的 `edgeId`、`fromNodeId`、`toNodeId` 和 `conditionSummary`，不复用 `additional_fields.designer_layout_v1` 保存 edge kind、handle id、port id、runtime order 或 React Flow 原始 edge。

## 验收方式

- Web build 覆盖前端类型、payload 序列化、restore guard、Node Designer 摘要和 Review Handoff 文案。
- Go `services/platform` 测试覆盖 `AdditionalFields` 保存 / 读取 / HTTP 返回、layout 规范化、嵌套 forbidden field 拦截和无 runtime side effects。
- `workflow-node-designer-persisted-layout-v1` checker 固定前端、后端、文档引用、fixture shape 和 fast baseline 顺序。
- `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不保存 React Flow 原始 `nodes` / `edges` JSON。
- 不保存 viewport、selection、drag state、connection preview、visual edge style 或 derived edge kind。
- 不把 layout 坐标解释为运行顺序、执行计划、runtime binding、publish ready、run ready 或 production ready。
- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、writeback、replay、resume 或 materialized result reader。
- 不接 repository mode、真实数据库、database connection provider、secret resolver、production resolver runtime、schema marker runtime、migration runner、Radish OIDC middleware、token validation、membership adapter、API key lifecycle、quota、billing 或 public production API。
