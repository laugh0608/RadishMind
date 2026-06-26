# Workflow Node Designer Layout Review Findings v1 任务卡

更新时间：2026-06-24

## 任务标识

- 切片：`workflow-node-designer-layout-review-findings-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_node_designer_layout_review_findings_v1_implemented`

## 目标

在 controlled edge mutation 已落地后，对 `Workflow Node Designer` 的真实 Builder 使用路径做布局复验，并修正画布、inspector 和 edge 列表在中宽 / 窄屏下的可读性与删除入口密度。

本任务只调整前端布局、展示结构、文档和既有专项 checker；不扩 Go schema，不新增 backend route，不保存 React Flow 原始 edge / handle id / port id / derived edge kind / runtime order，不打开 repository mode、真实数据库、OIDC、production API、publish、run、executor、confirmation、writeback 或 replay。

## 输入事实源

- [Workflow Node Designer Surface v1 专题](../features/workflow/node-designer-surface-v1.md)
- [Workflow Node Designer Controlled Edge Mutation Implementation v1 任务卡](workflow-node-designer-controlled-edge-mutation-implementation-v1-plan.md)
- [Workflow Node Designer Edge Editing Save Preconditions v1 专题](../features/workflow/node-designer-edge-editing-save-preconditions-v1.md)
- [Workflow Node Designer Review Handoff v1 专题](../features/workflow/node-designer-review-handoff-v1.md)

## 实现范围

- 为 Node Designer shell 增加中宽响应式断点，避免画布与 inspector 在 tablet 宽度互相挤压。
- 让 toolbar、status、mapping summary、edge grid 和 inspector 在窄屏下保持稳定换行，不因长 node id / edge id 溢出。
- 把 inspector 的 connected edge 删除入口改为结构化条目，展示 `edgeKind` 与 `edgeId`，删除按钮保持受控 `onRemoveEdge(edge.edgeId)`。
- 把 Draft Designer 下方 edge 卡片的标题区改为可换行 grid，展示 `edgeId`，并让删除按钮在窄屏下完整占行。
- 复用既有 edge editing / controlled mutation checker，补充布局 class、responsive breakpoint 和文档引用检查。

## 验收口径

- `WorkflowNodeDesigner.onConnect`、`onAddEdge`、`onRemoveEdge` 的 active draft mutation 语义不变。
- validation inspector、Preview Plan、Runtime Readiness 和 Review Handoff 仍消费 active draft，不消费 React Flow raw edge。
- 中宽布局有明确 breakpoint，Node Designer shell 不再只依赖 `820px` 以下才折叠。
- Inspector connected edge 条目与 Draft edge 卡片能容纳长 `nodeId` / `edgeId`，删除入口不挤占主文本。
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
