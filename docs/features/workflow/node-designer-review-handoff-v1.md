# Workflow Node Designer Review Handoff v1 专题

更新时间：2026-06-24

状态：`workflow_node_designer_review_handoff_v1_implemented`

## 专题定位

`Workflow Node Designer Review Handoff v1` 承接 `Workflow Node Designer Surface v1`、`Workflow Node Designer Saved Draft Mapping v1` 和 `Workflow Review Handoff Active Draft v1`，把节点画布的 validation overlay、inspector state、UI-only layout 和 saved draft mapping 汇总到现有 Review Handoff。

本专题不创建独立 review store、handoff store、backend route、saved draft persisted schema、Go document type、database table、OIDC middleware、repository mode、public production API、publish、run、executor、confirmation decision、writeback、replay 或 materialized result reader。

## 当前实现

- `WorkflowReviewHandoffSource` 显式接收 `activeWorkflowDraft`，避免从旧 selected draft 或 sample 重新手拼画布审查状态。
- `WorkflowReviewHandoffViewModel` 新增 `nodeDesignerReviewRecord`，记录 canvas layout、validation overlay、inspector state 和 saved draft mapping 四段 section。
- Review Handoff 面板新增 `Node Designer Review Handoff` 区域，展示每段 source surface、status、primary ref、item count、request / audit metadata、summary 和 reviewer question。
- `nodeDesignerReviewRecord` 只消费 active draft session 的 UI-only `designerLayout`、节点 / 边、validation inspector、execution plan preview 和 runtime readiness inspector；不把 layout、derived edge kind、overlay 或 inspector state 写入 persisted schema。
- 新增 `workflow-node-designer-review-handoff-v1` fixture / checker，并接入 fast baseline。

## 数据边界

允许汇总：

- active draft id、node count、edge count、UI-only positioned node count 和 default layout node count。
- validation overlay 中的 structural / contract / blocked capability evidence refs。
- inspector 中的 label、summary、provider / tool / RAG refs、contract fields、output mapping、risk marker 和 requires confirmation marker。
- saved draft mapping 的 node attributes、contract fields、edge endpoints、condition summary、UI-only layout 与 derived edge kind 边界说明。
- request id、audit ref、source surface、reviewer question 和 advisory-only status。

明确排除：

- React Flow 原始 `nodes` / `edges` JSON。
- persisted `additional_fields.designer_layout_v1` 或等价跨会话 layout schema。
- viewport、selection、drag state、connection preview、visual edge style 或 derived edge kind 的持久化。
- secret value、API key value、OAuth / OIDC token、provider response、tool result、artifact payload。
- review / handoff persistence、handoff export、handoff send、execution plan persistence、runtime readiness persistence。
- confirmation decision、run input / output、writeback payload、replay / resume state。

## 审查流程

1. active draft 先由 Draft Designer / Node Designer 形成当前本地草案状态。
2. validation inspector、execution plan preview 和 runtime readiness inspector 继续从同一 active draft 派生。
3. Review Handoff 先展示已有 active draft validation / plan / readiness record。
4. Node Designer Review Handoff 紧随其后展示 canvas layout、validation overlay、inspector state 和 saved draft mapping。
5. Reviewer 只能把这些内容作为审查上下文，不能据此发布、运行、提交确认或写回上层业务真相源。

## 验收方式

- Web build 通过，覆盖 `nodeDesignerReviewRecord` 类型、派生逻辑和面板渲染。
- `workflow-node-designer-review-handoff-v1` checker 固定 source、record sections、UI 文案、文档引用、fast baseline 顺序和停止线。
- `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不保存、不导出、不发送 node designer review record。
- 不新增 saved draft persisted schema，不写 `additional_fields.designer_layout_v1`。
- 不把 layout、overlay、inspector state、derived edge kind 或 `valid_for_review` 解释为运行顺序、runtime binding、publish ready、run ready 或 production ready。
- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、writeback、replay、resume 或 materialized result reader。
- 不接 durable persistence、repository adapter、真实数据库、schema migration、store selector、Radish OIDC middleware、token validation、membership adapter、API key lifecycle、quota、billing 或 public production API。
