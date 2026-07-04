# Workflow Node Designer Review Handoff v1 专题

更新时间：2026-07-04

状态：`workflow_node_designer_review_handoff_v1_implemented`

## 专题定位

`Workflow Node Designer Review Handoff v1` 承接 `Workflow Node Designer Surface v1`、`Workflow Node Designer Saved Draft Mapping v1` 和 `Workflow Review Handoff Active Draft v1`，把节点画布的 validation overlay、inspector state、UI-only layout 和 saved draft mapping 汇总到现有 Review Handoff。当前 [Workflow Node Designer Persisted Layout v1](node-designer-persisted-layout-v1.md) 已完成，Review Handoff 会展示 layout 来源可能是 active draft session 或 saved draft metadata，但它自身仍不保存、导出或发送 handoff record。

本专题不创建独立 review store、handoff store、backend route、database table、OIDC middleware、repository mode、public production API、publish、run、executor、confirmation decision、writeback、replay 或 materialized result reader。saved draft layout schema 由 `Workflow Node Designer Persisted Layout v1` 单独承接。

## 当前实现

- `WorkflowReviewHandoffSource` 显式接收 `activeWorkflowDraft`，避免从旧 selected draft 或 sample 重新手拼画布审查状态。
- `WorkflowReviewHandoffViewModel` 新增 `nodeDesignerReviewRecord`，记录 canvas layout、validation overlay、inspector state 和 saved draft mapping 四段 section。
- Review Handoff 面板新增 `Node Designer Review Handoff` 区域，展示每段 source surface、status、primary ref、item count、request / audit metadata、summary 和 reviewer question。
- `nodeDesignerReviewRecord` 只消费 active draft session 的 `designerLayout`、节点 / 边、validation inspector、execution plan preview 和 runtime readiness inspector；`designerLayout.persistence` 可能是 `ui_only` 或 `saved_draft_metadata`，但 Handoff 不把 layout、derived edge kind、overlay 或 inspector state 写入自己的持久化记录。
- `Workflow Node Designer Persisted Layout v1` 已把节点坐标保存为 `additional_fields.designer_layout_v1`，Review Handoff 只展示该 saved draft metadata 来源。
- `Workflow Node Designer Controlled Edge Mutation Implementation v1` 已让 Review Handoff 消费新增 / 删除 edge 后的 active draft，不消费 React Flow raw edge 或 handle id。
- `Workflow Node Designer Validation Overlay Navigation v1` 已把 finding focus 映射到画布节点 / 连线 / inspector；Review Handoff 只汇总 validation overlay evidence，不保存当前 focus 或 selection。
- `Workflow Node Designer Graph Review Handoff Refinement v1` 已把 validation overlay detail review 收束为 `graphReviewFindings`，按 node-targeted、edge-targeted 和 graph-level finding 展示 source check、severity、target refs、summary 和 reviewer question，并在面板中提供 graph review summary 与分组阅读路径。
- 2026-07-04 已补 `handoffPath` 与 `handoffPathRefs`，让每条 graph review finding 显示 validation overlay、node inspector / edge review / runtime readiness 与 evidence refs 的 handoff path；该路径只属于只读审查视图，不保存、不导出、不写入 saved draft。
- 新增 `workflow-node-designer-review-handoff-v1` fixture / checker，并接入 fast baseline。

## 数据边界

允许汇总：

- active draft id、node count、edge count、positioned node count、layout persistence source 和 default layout node count。
- validation overlay 中的 structural / contract / blocked capability evidence refs。
- graph review handoff refinement 中的 node / edge / graph-level finding、source check id、severity、target refs、target summary、handoff path、handoff path refs、evidence refs 和 reviewer question。
- inspector 中的 label、summary、provider / tool / RAG refs、contract fields、output mapping、risk marker 和 requires confirmation marker。
- saved draft mapping 的 node attributes、contract fields、edge endpoints、condition summary、saved draft metadata layout 与 derived edge kind 边界说明。
- request id、audit ref、source surface、reviewer question 和 advisory-only status。

明确排除：

- React Flow 原始 `nodes` / `edges` JSON。
- 由 Review Handoff 自行写入或修改 `additional_fields.designer_layout_v1`。
- viewport、selection、drag state、connection preview、visual edge style 或 derived edge kind 的持久化。
- secret value、API key value、OAuth / OIDC token、provider response、tool result、artifact payload。
- review / handoff persistence、handoff export、handoff send、execution plan persistence、runtime readiness persistence。
- confirmation decision、run input / output、writeback payload、replay / resume state。

## Graph Review Finding 阅读说明

Review Handoff 中的 `graphReviewFindings` 是给人工审查者使用的只读索引，用于把 validation overlay 中的结构检查、契约检查和 blocked capability 检查映射到更容易定位的节点、连线或全图审查项。它不是执行顺序、保存记录、发布条件或 runtime binding。

每条 finding 的字段含义如下：

- `sourceCheckId` 指向 validation inspector 中的原始检查项。
- `targetKind` 只表达审查定位层级：`node`、`edge` 或 `graph`。
- `targetRefs` 是当前 active draft 中可读的节点、连线或全图审查引用，不是持久化主键。
- `targetSummary` 给出受影响范围摘要，帮助 reviewer 先判断要看节点、连线还是全图能力。
- `handoffPath` 给出建议阅读路径：节点类问题从 validation overlay 进入 node inspector 和 saved draft mapping；连线类问题从 validation overlay 进入 connected edge review 和 draft edge summary；全图类问题从 validation overlay 进入 runtime readiness 和 decision blockers。
- `handoffPathRefs` 把上述阅读路径展开为 section ref 与 target ref，方便面板展示可追溯 token。
- `evidenceRefs` 保留检查项、缺失字段、目标引用或 audit ref 的脱敏证据引用，不包含 payload、secret、provider response 或执行结果。
- `reviewerQuestion` 是人工审查提示，不是自动修复、保存、运行或确认命令。

建议阅读顺序：

1. 先看 graph review summary，确认当前是 node-targeted、edge-targeted 还是 graph-level finding。
2. 再看 `targetRefs` 和 `targetSummary`，定位需要审查的节点、连线或全图能力。
3. 按 `handoffPath` 进入对应只读区块，对照 validation overlay、inspector、saved draft mapping、runtime readiness 或 decision blockers。
4. 最后用 `evidenceRefs` 核对来源，确认 finding 是否足以支撑人工判断。

这条阅读路径不会被保存到 saved draft，不会被导出为 handoff artifact，也不会转换成 executor、adapter、repository 或 runtime 的输入。

## 审查流程

1. active draft 先由 Draft Designer / Node Designer 形成当前本地草案状态。
2. validation inspector、execution plan preview 和 runtime readiness inspector 继续从同一 active draft 派生。
3. Review Handoff 先展示已有 active draft validation / plan / readiness record。
4. Node Designer Review Handoff 紧随其后展示 canvas layout、validation overlay、inspector state 和 saved draft mapping。
5. Graph review finding 明细按 active draft 和 validation inspector 重新派生，帮助 reviewer 区分需要查看的节点、连线或全图 blocked capability。
6. 如果 reviewer 在画布中聚焦某个 validation finding，该 focus 只帮助定位节点 / 连线；handoff record 仍按 active draft 和 inspector evidence 重新派生。
7. Reviewer 只能把这些内容作为审查上下文，不能据此发布、运行、提交确认或写回上层业务真相源。

## 验收方式

- Web build 通过，覆盖 `nodeDesignerReviewRecord` 类型、派生逻辑和面板渲染。
- `workflow-node-designer-review-handoff-v1` checker 固定 source、record sections、`graphReviewFindings`、handoff path / evidence refs、graph review summary / grouping UI、文档引用、fast baseline 顺序和停止线。
- `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不保存、不导出、不发送 node designer review record。
- 不由 Review Handoff 新增或修改 saved draft persisted schema；`additional_fields.designer_layout_v1` 只由 Persisted Layout 保存链路承接。
- 不把 graph review finding、layout、overlay、inspector state、derived edge kind 或 `valid_for_review` 解释为运行顺序、runtime binding、publish ready、run ready 或 production ready。
- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、writeback、replay、resume 或 materialized result reader。
- 不接 durable persistence、repository adapter、真实数据库、schema migration、store selector、Radish OIDC middleware、token validation、membership adapter、API key lifecycle、quota、billing 或 public production API。
