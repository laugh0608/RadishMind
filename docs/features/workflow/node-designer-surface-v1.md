# Workflow Node Designer Surface v1 专题

更新时间：2026-06-24

状态：`workflow_node_designer_surface_v1_defined`

## 专题定位

`Workflow Node Designer Surface v1` 是 `Workflow / Agent Runtime` 下的节点画布设计专题，承接现有 `Workflow Draft Designer Surface`、`Workflow Draft Designer Editing Model v2`、`Workflow Draft Node Attribute Editing Model v1` 和 `Workflow Review Handoff Active Draft v1`。

本专题定义节点画布作为 workflow 草案设计、审查和 plan preview 的交互方向；首批前端画布接入已由 [Workflow Node Designer Surface Implementation v1 任务卡](../../task-cards/workflow-node-designer-surface-implementation-v1-plan.md) 承接并实现。它仍不打开完整拖拽 builder、workflow publish、run、executor、node executor、tool executor、agent loop、confirmation decision、writeback、replay、resume、repository mode、真实数据库、OIDC middleware、token validation、membership adapter 或 public production API。

交互参考可以借鉴 ComfyUI 的节点图工作流形态，但 RadishMind 的目标不是复制 image-generation-only 节点图，而是建立面向 AI 工具、模型网关、RAG、审查、候选动作和 artifact handoff 的可审查 workflow 草案设计面。

## 阶段位置

本专题位于 Workflow Builder 体验方向，顺序如下：

1. 已完成：保存 / 读取 / 校验 saved workflow draft。
2. 已完成：Draft Designer 本地结构编辑与节点属性编辑。
3. 已完成：active draft validation / plan / readiness 的 Review Handoff。
4. 当前定义：节点画布 Designer Surface 的信息架构、graph model、typed port、validation overlay 和停止线。
5. 已完成：[Workflow Node Designer Library Selection v1](node-designer-library-selection-v1.md)，选定 `@xyflow/react` 作为首批画布实现依赖。
6. 已实现：[Workflow Node Designer Surface Implementation v1 任务卡](../../task-cards/workflow-node-designer-surface-implementation-v1-plan.md)，完成 `@xyflow/react` 前端画布、active draft graph adapter、custom node / edge、inspector bridge 和连线校验反馈。
7. 已定义：[Workflow Node Designer Saved Draft Mapping v1](node-designer-saved-draft-mapping-v1.md)，确认 layout metadata、edge kind 与 validation overlay 的 saved draft / Review Handoff 映射边界。
8. 已实现：[Workflow Node Designer Saved Draft Mapping Implementation v1](../../task-cards/workflow-node-designer-saved-draft-mapping-implementation-v1-plan.md)，完成 UI-only layout、拖拽回写和 mapping summary。
9. 已实现：[Workflow Node Designer Review Handoff v1](node-designer-review-handoff-v1.md)，把 canvas layout、validation overlay、inspector state 和 saved draft mapping 汇总进现有 Review Handoff。
10. 已实现：[Workflow Node Designer Persisted Layout v1](node-designer-persisted-layout-v1.md)，通过 `additional_fields.designer_layout_v1` 保存受控节点坐标，恢复时兼容旧草案和非法 layout metadata。
11. 已定义：[Workflow Node Designer Edge Editing Save Preconditions v1](node-designer-edge-editing-save-preconditions-v1.md)，状态为 `workflow_node_designer_edge_editing_save_preconditions_v1_defined`，固定画布连线新增 / 删除进入 `draft.edges` 和 saved draft 保存链路前的字段、保存前置与 validation 消费。
12. 已实现：[Workflow Node Designer Controlled Edge Mutation Implementation v1 任务卡](../../task-cards/workflow-node-designer-controlled-edge-mutation-implementation-v1-plan.md)，状态为 `workflow_node_designer_controlled_edge_mutation_implementation_v1_implemented`，受控新增 / 删除 edge 只修改 active draft，并继续复用 saved draft edge mapping、validation inspector 和 Review Handoff。
12. 后续独立目标：publish、run、executor、confirmation、writeback 和 replay。

它不替代 durable store 上游前置，也不解锁 repository mode。若下一批选择继续 durable store，上游 auth、membership、schema marker、secret resolver、connection provider 和 production resolver blocker 仍按既有专题推进。

## 目标用户

- `Workspace Builder`：用画布方式组织 workflow 草案，能快速添加节点、连线、查看输入输出契约和 blocked capability。
- `Workflow Reviewer`：在同一张图上审查 contract、risk、readiness、requires confirmation 和 handoff 摘要。
- `Platform Maintainer`：维护节点类型、端口类型、edge 语义、保存映射、validation 规则和停止线，避免 UI 把草案误表达为可执行 runtime。

## 设计目标

- 用节点画布承载 workflow draft 的可视化结构，不再只靠列表式节点卡片表达图关系。
- 节点、端口和边必须保留 typed contract 语义，而不是只有视觉连线。
- validation、execution plan preview、runtime readiness inspector 和 Review Handoff 继续消费 active draft。
- 保存仍复用 Saved Workflow Draft v1 的 sanitized payload 和 dev-only consumer 语义；读取失败、保存失败和版本冲突继续 fail closed。
- 画布 UI 必须明确区分 `sample`、`local_edit`、`unsaved_local`、`saved_dev_record`、`version_conflict`、`save_failed` 和 `read_failed`。
- 顶部主动作应优先表达 `Validate`、`Preview Plan`、`Save Draft` 或 `Review Handoff`；不得把按钮写成会误导用户的 production `Run`。

## 节点模型

v1 的持久化节点类型默认沿用 Saved Workflow Draft v1 已允许的草案类型：

- `prompt`
- `llm`
- `http_tool`
- `rag_retrieval`
- `condition`
- `output`

画布可以在 UI 层提供以下分组或虚拟提示，但它们不得在没有 schema / task card 的情况下变成新的持久化 node type：

- `input / intent`：表达用户输入、上下文和触发条件，可映射为 input contract 或 prompt 前置摘要。
- `review_gate`：表达人工审查、requires confirmation 或 risk gate，可映射为 validation finding / readiness blocker。
- `artifact`：表达 artifact ref、preview 或 handoff 摘要，可映射为 output contract / output mapping。
- `audit / handoff`：表达 request / audit metadata 和 review handoff 摘要，可映射为 existing metadata，不写 runtime event。

后续若需要新增持久化 node type，必须先更新 saved draft schema / consumer mapping，并新增对应实现任务卡、fixture 和 checker。

## Port 与 Edge 语义

节点端口应以 contract 为核心，至少能表达：

- `port_id`
- `label`
- `direction`
- `value_type`
- `required`
- `source_summary`
- `risk_marker`

建议先从以下 `value_type` 组织 UI 与 validation：

- `text`
- `json`
- `artifact_ref`
- `provider_ref`
- `tool_ref`
- `rag_ref`
- `risk_finding`
- `review_record`

边应区分语义，而不是只保存视觉连线：

- `data_edge`：传递 text / json / artifact ref 等数据契约。
- `control_edge`：表达 condition 后的控制流摘要。
- `guard_edge`：表达 requires confirmation、blocked capability 或 policy gate。
- `audit_edge`：表达 request / audit / handoff metadata 的可审查关系。

如果当前 saved draft schema 尚不能表达上述 edge kind，初期实现只能把它们作为 UI derived state 或 validation finding 展示，不得静默扩展持久化格式。

## 页面结构

Node Designer Surface 建议保持四区布局：

- 左侧：节点库、模板、节点类型过滤和 blocked capability 分组。
- 中间：节点画布、port、edge、缩放、框选、节点位置和 validation overlay。
- 右侧：当前节点 inspector，编辑 label、provider / profile ref、tool ref、RAG ref、input / output contract fields 和 output mapping。
- 底部或侧边：validation findings、plan preview、runtime readiness 和 review handoff 摘要。

画布不是 marketing hero，也不是静态说明页。实现时第一屏应直接是可操作的 workflow designer。

## 数据边界

允许保存或恢复：

- 节点位置和布局 metadata，前提是映射为 UI-only 或明确的 draft metadata。
- 节点 label、summary、provider / profile ref、tool ref、RAG ref。
- input / output contract fields。
- output mapping summary。
- edge condition summary 和 validation finding。

明确排除：

- secret value、API key value、OAuth / OIDC token。
- provider response、工具执行结果、真实 artifact payload。
- confirmation decision、run input / output、writeback payload、replay / resume state。
- repository mode runtime metadata、database connection runtime state、schema marker runtime state。

## 实现拆分建议

后续实现不应直接从运行器开始，应按以下顺序选择：

1. [Workflow Node Designer Library Selection v1](node-designer-library-selection-v1.md)：已确定 `@xyflow/react`、状态模型、依赖引入方式、bundle / test 影响和 fallback 策略。
2. [Workflow Node Designer Surface Implementation v1](../../task-cards/workflow-node-designer-surface-implementation-v1-plan.md)：已接入首批前端画布，只覆盖节点展示、拖拽、连线、选中、inspector、validation overlay 和 active draft 派生。
3. [Workflow Node Designer Saved Draft Mapping v1](node-designer-saved-draft-mapping-v1.md)：已定义 layout metadata、edge kind、validation overlay 与 saved draft / Review Handoff 的映射边界；不保存 React Flow 原始对象。
4. [Workflow Node Designer Saved Draft Mapping Implementation v1](../../task-cards/workflow-node-designer-saved-draft-mapping-implementation-v1-plan.md)：已实现 UI-only session layout、节点拖拽回写和 mapping summary。
5. [Workflow Node Designer Review Handoff v1](node-designer-review-handoff-v1.md)：已把 node designer review handoff、validation overlay、inspector state 和 saved draft mapping 汇总到 existing Review Handoff，不创建 runtime review store。
6. [Workflow Node Designer Persisted Layout v1](node-designer-persisted-layout-v1.md)：已实现 `additional_fields.designer_layout_v1` 受控 schema、保存 / 恢复兼容、Go `AdditionalFields` 保留和 saved draft layout metadata checker。
7. [Workflow Node Designer Edge Editing Save Preconditions v1](node-designer-edge-editing-save-preconditions-v1.md)：已固定后续画布连线新增 / 删除只写入 `draft.edges` 的 `edgeId`、`fromNodeId`、`toNodeId` 和 `conditionSummary`，并要求 validation inspector、local edit 状态和 saved draft mapping 保持一致。
8. [Workflow Node Designer Controlled Edge Mutation Implementation v1 任务卡](../../task-cards/workflow-node-designer-controlled-edge-mutation-implementation-v1-plan.md)：已完成 `onConnect` 受控新增 edge、edge 删除入口、`local_edit` / `unsaved_local` 状态和专项 checker 更新，不保存 React Flow 原始 edge、handle id、port id、derived edge kind 或 runtime order。

## 验收方式

本专题定义与首批实现阶段：

- 文档入口已收录本专题。
- 当前焦点能说明它位于 Workflow Builder 体验方向。
- 实现任务卡已收录首批前端画布接入；不新增 schema、backend route、fixture、checker 或 runtime artifact。
- `./scripts/check-repo.sh --fast` 通过。

后续实现阶段：

- Web build 通过。
- 画布交互测试覆盖节点新增、移动、选中、连线、删除保护、validation overlay 和 read / save failure state。
- Saved draft consumer smoke 继续覆盖 no sample fallback、version conflict 和 failure code。
- 若引入新 schema、route、dependency 或高风险边界，补专项 task card、fixture、checker 和对应仓库基线。

## 停止线

- 除已完成的 `Workflow Node Designer Surface Implementation v1` 前端依赖与 UI 接入外，不继续扩大画布库、持久化格式或 runtime 能力。
- 不实现完整拖拽 builder、publish、run、executor、node executor、tool executor、agent loop、confirmation decision、decision store、writeback、replay、resume 或 materialized result reader。
- 不接 repository mode、真实数据库、database connection provider、secret resolver、production resolver runtime、schema marker runtime、migration runner、Radish OIDC middleware、token validation、membership adapter、API key lifecycle、quota、billing 或 public production API。
- 不把画布上的 `valid_for_review`、`Preview Plan`、readiness 或 handoff 摘要写成 publish ready、run ready 或 production ready。
