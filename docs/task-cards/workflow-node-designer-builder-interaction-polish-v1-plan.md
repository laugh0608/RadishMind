# Workflow Node Designer Builder Interaction Polish v1 任务卡

更新时间：2026-06-24

## 任务标识

- 切片：`workflow-node-designer-builder-interaction-polish-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_node_designer_builder_interaction_polish_v1_implemented`

## 目标

在 controlled edge mutation 和 layout review findings 已落地后，补齐 Node Designer 的 Builder 操作反馈：让用户能明确看到当前选中节点、编辑锁定状态、连接 / 删除 / 拖拽结果，并能不用回到列表就快速切换节点。

本任务只调整前端交互、展示结构、文档和既有专项 checker；不扩 Go schema，不新增 backend route，不保存 React Flow 原始 edge / handle id / port id / derived edge kind / runtime order，不打开 repository mode、真实数据库、OIDC、production API、publish、run、executor、confirmation、writeback 或 replay。

## 输入事实源

- [Workflow Node Designer Surface v1 专题](../features/workflow/node-designer-surface-v1.md)
- [Workflow Node Designer Layout Review Findings v1 任务卡](workflow-node-designer-layout-review-findings-v1-plan.md)
- [Workflow Node Designer Controlled Edge Mutation Implementation v1 任务卡](workflow-node-designer-controlled-edge-mutation-implementation-v1-plan.md)
- [Workflow Node Designer Review Handoff v1 专题](../features/workflow/node-designer-review-handoff-v1.md)

## 实现范围

- 为 Node Designer 增加 Builder interaction bar，集中展示编辑状态、当前选中节点和最近交互反馈。
- 增加节点快速选择入口，允许在 canvas 外直接切换 inspector 目标，并保持 selection state 为 UI-only。
- 把连接、拒绝连接、删除 edge、拖拽位置保存和选择节点的反馈统一为 tone + message，不改变 active draft 数据结构。
- 在 editing disabled 时继续允许查看和选择节点，但明确提示草案操作被锁定。
- 继续复用 active draft mutation、saved draft edge mapping、validation inspector、Preview Plan 和 Review Handoff。

## 验收口径

- 快速选择节点只更新 `selectedNodeId` 和 UI feedback，不写 saved draft payload。
- 连接成功 / 失败、删除 edge 成功 / 失败、拖拽位置保存和选择节点都有明确反馈 tone。
- `onConnect` 仍先校验再调用 `onAddEdge`，删除仍通过 `onRemoveEdge(edge.edgeId)`。
- validation inspector、Preview Plan、Runtime Readiness 和 Review Handoff 仍消费 active draft。
- 保存 payload 字段仍只通过既有 saved draft consumer 输出 endpoint 与 condition summary。

## 验证要求

- `npm run build`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-node-designer-edge-editing-save-preconditions-v1.py`
- `git diff --check`
- `./scripts/check-repo.sh --fast`

## 停止线

- 不扩 saved draft schema、Go document type 或 backend route。
- 不保存 React Flow 原始 edge 对象、handle id、port id、viewport、selection、connection preview、visual edge style、derived edge kind 或 runtime order。
- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、writeback、replay、repository mode、真实数据库、OIDC middleware、token validation、membership adapter 或 public production API。
