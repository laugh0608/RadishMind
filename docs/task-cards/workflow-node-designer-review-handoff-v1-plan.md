# Workflow Node Designer Review Handoff v1 任务卡

更新时间：2026-06-25

## 任务标识

- 切片：`workflow-node-designer-review-handoff-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_node_designer_review_handoff_v1_implemented`

## 目标

在 `Workflow Node Designer Saved Draft Mapping v1` 的 UI-only layout mapping 已实现后，把节点画布相关审查信息汇总到现有 Review Handoff。

本任务卡只承接前端 view model / panel 汇总、专项 checker 和文档同步。不新增 saved draft persisted schema、backend route、Go document type、repository、数据库、OIDC、public production API、publish、run、executor、confirmation decision、writeback 或 replay。

## 本轮实现

- `WorkflowReviewHandoffSource` 新增 `activeWorkflowDraft`，确保 Node Designer handoff 消费当前 active draft。
- `WorkflowReviewHandoffViewModel` 新增 `nodeDesignerReviewRecord`，记录 canvas layout、validation overlay、inspector state 和 saved draft mapping。
- Review Handoff 面板新增 `Node Designer Review Handoff` 区域，展示四段 section 的 status、primary ref、item count、request / audit metadata、summary 和 reviewer question。
- evidence checklist 新增 `node_designer_review_handoff`，key findings 新增 `node_designer_review`。
- boundary locks 明确 layout、derived edge kind、validation overlay 和 inspector state 不创建 persisted runtime state。
- 新增 `workflow-node-designer-review-handoff-v1` fixture / checker，并接入 fast baseline。
- 后续 `Workflow Node Designer Graph Review Handoff Refinement v1` 复用本 checker，将 validation overlay detail review 派生为 `graphReviewFindings`，按 node / edge / graph-level finding 展示审查目标。

## 输入事实源

- [Workflow Node Designer Review Handoff v1 专题](../features/workflow/node-designer-review-handoff-v1.md)
- [Workflow Node Designer Surface v1 专题](../features/workflow/node-designer-surface-v1.md)
- [Workflow Node Designer Saved Draft Mapping v1 专题](../features/workflow/node-designer-saved-draft-mapping-v1.md)
- [Workflow Review Handoff Active Draft v1 专题](../features/workflow/review-handoff-active-draft-v1.md)

## 验收口径

- Node Designer handoff record 必须消费当前 active draft，不从旧 selected draft 或 sample 手拼。
- Graph review findings 必须消费当前 active draft 和 validation inspector，不保存 validation focus、selection、viewport 或 React Flow raw object。
- Review Handoff 必须明确 advisory-only，不提供保存、导出、发送、publish、run、confirmation decision、writeback 或 replay 控件。
- evidence checklist 必须包含 node designer review handoff 的 request / audit metadata。
- Web build、专项 checker 和 `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不实现 review persistence、handoff persistence、handoff export、handoff send、workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、business writeback、run replay、run resume 或 materialized result reader。
- 不实现 saved draft persisted layout schema，不写 `additional_fields.designer_layout_v1`。
- 不实现 durable persistence、repository adapter、真实数据库、schema migration、store selector、Radish OIDC middleware、token validation、membership adapter、API key lifecycle、quota enforcement、billing、cost ledger 或 public production API。
- 不把 node designer review record、layout、validation overlay、inspector state、derived edge kind 或 `valid_for_review` 解释为 runtime binding、durable persistence、publish、run 或 production readiness。
