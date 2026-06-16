# Workflow Draft Node Attribute Editing Model v1 任务卡

更新时间：2026-06-16

## 任务标识

- 切片：`workflow-draft-node-attribute-editing-model-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_draft_node_attribute_editing_model_v1_implemented`

## 目标

在 Draft Designer 已支持受控字段编辑、用户工作区草案创建、saved draft list / restore 和本地图结构编辑之后，补齐节点属性编辑能力，让用户能编辑节点 provider / profile、tool ref、RAG ref、input / output contract fields 和 output mapping，并让这些属性进入 dev-only saved draft 保存、读取、校验和恢复链路。

本任务卡只承接 Draft Designer 节点属性编辑与 dev-only saved draft schema / consumer 映射，不声明 durable persistence、public production API、publish、run、executor、confirmation decision、writeback 或 replay。

## 本轮实现

- `WorkflowDraftDesignerNode` 增加 provider / tool / RAG refs、contract fields 和 output mapping 字段。
- Draft Designer 节点卡片增加 `workflow-draft-node-attribute-grid` 属性编辑区。
- `App.tsx` 增加节点属性 patch handlers、contract fields parse helper、新增节点默认属性和 clone 数组深拷贝。
- saved draft consumer 保存节点级 summary、contract fields、output mapping 和 provider / tool / RAG refs；恢复时反向填回 Draft Designer node。
- validation inspector 和 execution plan preview 消费节点显式属性，避免继续只依赖 summary 文本。
- Platform dev-only saved draft schema 增加节点属性字段；Go save / read / validate tests 必须覆盖节点属性保留。
- 新增 `workflow-draft-node-attribute-editing-model-v1` fixture / checker，并接入 fast baseline。

## 输入事实源

- [Workflow Draft Node Attribute Editing Model v1 专题](../features/workflow/draft-node-attribute-editing-model-v1.md)
- [Workflow Draft Designer Editing Model v2 专题](../features/workflow/draft-designer-editing-model-v2.md)
- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)
- [Workflow Draft Designer Surface 专题](../features/workflow/draft-designer-surface.md)

## 验收口径

- 节点属性必须进入 dev-only saved draft payload，不能只停留在 UI-only 状态。
- saved draft restore 必须恢复节点属性，读取失败仍按既有 fail-closed 语义处理。
- validation inspector 必须优先使用显式 input / output contract fields。
- execution plan preview 必须优先使用节点 provider / tool / policy ref。
- Go save / read / validate tests 必须覆盖节点属性保留。
- Web build、Go tests、专项 checker 和 `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不实现画布拖拽编排、分支 DSL 编辑器、完整执行参数 schema、workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、business writeback、run replay、run resume 或 materialized result reader。
- 不实现 durable persistence、repository adapter、真实数据库、schema migration、store selector、Radish OIDC middleware、token validation、API key lifecycle、quota enforcement、billing、cost ledger 或 public production API。
- 不把节点属性、saved dev draft、validation summary、execution plan preview 或 `valid_for_review` 解释为 runtime binding、durable persistence、publish、run 或 production readiness。
