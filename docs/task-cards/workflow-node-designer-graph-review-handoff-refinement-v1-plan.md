# Workflow Node Designer Graph Review Handoff Refinement v1 任务卡

更新时间：2026-06-25

## 任务标识

- 切片：`workflow-node-designer-graph-review-handoff-refinement-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_node_designer_graph_review_handoff_refinement_v1_implemented`

## 目标

在 `Workflow Node Designer Validation Overlay Navigation v1` 已实现后，把 validation overlay detail review 收束进现有 Review Handoff：Reviewer 能在 handoff 中直接看到 node-targeted、edge-targeted 和 graph-level finding，而不需要从四段概览重新推断需要审查的节点、连线或全图阻塞项。

本任务只改前端 view model / panel、专题文档、任务卡和既有 checker fixture。不新增 saved draft schema、backend route、Go document type、repository、数据库、OIDC、public production API、publish、run、executor、confirmation decision、writeback、replay 或 handoff persistence。

## 本轮实现

- `WorkflowReviewHandoffNodeDesignerReviewRecord` 新增 `graphReviewFindings`、`nodeTargetedFindingCount`、`edgeTargetedFindingCount` 和 `graphLevelFindingCount`。
- `buildNodeDesignerGraphReviewFindings` 从 active draft、structural checks、contract checks 和 blocked capability checks 派生 graph review finding。
- Review Handoff 面板在 `Node Designer Review Handoff` 下展示 graph review finding 卡片，包含 source check、severity、target kind、target refs、summary 和 reviewer question。
- Review Handoff 面板已补 graph review summary，并按 node / edge / graph-level 分组展示 finding，降低 reviewer 从长卡片列表中反推目标类型的成本。
- key findings 新增 `node_designer_graph_review`，把 validation overlay detail review 作为 graph review handoff 的首要审查项。
- 复用 `workflow-node-designer-review-handoff-v1` fixture / checker 增加 graph review handoff refinement 检查，不新增独立专项 checker。

## 输入事实源

- [Workflow Node Designer Review Handoff v1 专题](../features/workflow/node-designer-review-handoff-v1.md)
- [Workflow Node Designer Validation Overlay Navigation v1 任务卡](workflow-node-designer-validation-overlay-navigation-v1-plan.md)
- [Workflow Node Designer Builder Interaction Polish v1 任务卡](workflow-node-designer-builder-interaction-polish-v1-plan.md)
- [Workflow Draft Validation Inspector Offline v1 任务卡](workflow-draft-validation-inspector-offline-v1-plan.md)

## 验收口径

- Graph review findings 必须消费当前 active draft 和现有 validation inspector，不另建 validation 真相源。
- Node-targeted finding 只指向已有 draft node，edge-targeted finding 只指向已有 active draft edge，graph-level finding 只保留 blocked capability / prerequisite 审查语义。
- Handoff 只能展示审查上下文，不保存 validation focus、selection、viewport、React Flow raw object、derived edge kind 或 runtime order。
- Web build、既有 Node Designer Review Handoff checker、`git diff --check` 和 `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不实现 review persistence、handoff persistence、handoff export、handoff send、workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、business writeback、run replay、run resume 或 materialized result reader。
- 不保存 validation focus、selection、viewport、React Flow 原始 node / edge、handle id、port id、derived edge kind 或 runtime order。
- 不新增 saved draft persisted schema、backend route、Go document type、repository mode、真实数据库、schema migration、store selector、Radish OIDC middleware、token validation、membership adapter、API key lifecycle、quota enforcement、billing、cost ledger 或 public production API。
- 不把 graph review finding、validation overlay detail review、layout、inspector state、derived edge kind 或 `valid_for_review` 解释为 runtime binding、publish ready、run ready 或 production ready。
