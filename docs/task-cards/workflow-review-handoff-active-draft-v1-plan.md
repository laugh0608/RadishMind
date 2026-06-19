# Workflow Review Handoff Active Draft v1 任务卡

更新时间：2026-06-16

## 任务标识

- 切片：`workflow-review-handoff-active-draft-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_review_handoff_active_draft_v1_implemented`

## 目标

在 Draft Designer 已支持 saved draft list / restore、本地图结构编辑和节点属性编辑之后，把当前 active draft 的 validation inspector、execution plan preview 和 runtime readiness inspector 汇总到 Workflow Review Handoff，形成可交接审查记录。

本任务卡只承接前端 view model / panel 汇总、专项 checker 和文档同步，不新增 backend route、repository、数据库、OIDC、public production API、publish、run、executor、confirmation decision、writeback 或 replay。

## 本轮实现

- `WorkflowReviewHandoffSource` 显式接收 active draft validation inspector、execution plan preview 和 runtime readiness inspector。
- `WorkflowReviewHandoffViewModel` 新增 active draft review record，记录 validation / plan / readiness 三段 section。
- Review Handoff 面板新增 active draft review record 区域，展示 source surface、status、primary ref、blocker count、request、audit、summary 和 reviewer question。
- evidence checklist 增加 validation inspector、execution plan preview 和 runtime readiness inspector 三项，避免 handoff 只依赖 workspace review stage 摘要。
- 新增 `workflow-review-handoff-active-draft-v1` fixture / checker，并接入 fast baseline。

## 输入事实源

- [Workflow Review Handoff Active Draft v1 专题](../features/workflow/review-handoff-active-draft-v1.md)
- [Workflow Draft Node Attribute Editing Model v1 专题](../features/workflow/draft-node-attribute-editing-model-v1.md)
- [Workflow Draft Designer Editing Model v2 专题](../features/workflow/draft-designer-editing-model-v2.md)
- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)

## 验收口径

- handoff record 必须消费当前 active draft 的 validation inspector、execution plan preview 和 runtime readiness inspector，不能重新从旧 selected draft 或 sample 手拼。
- Review Handoff 必须明确 advisory-only，不提供保存、导出、发送、publish、run、confirmation decision、writeback 或 replay 控件。
- evidence checklist 必须包含 validation / plan / readiness 三段来源的 request / audit metadata。
- Web build、专项 checker 和 `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不实现 review persistence、handoff persistence、handoff export、handoff send、workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、business writeback、run replay、run resume 或 materialized result reader。
- 不实现 durable persistence、repository adapter、真实数据库、schema migration、store selector、Radish OIDC middleware、token validation、API key lifecycle、quota enforcement、billing、cost ledger 或 public production API。
- 不把 active draft review record、validation summary、execution plan preview、runtime readiness inspector 或 `valid_for_review` 解释为 runtime binding、durable persistence、publish、run 或 production readiness。
