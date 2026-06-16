# Workflow Draft Node Attribute Editing Model v1 专题

更新时间：2026-06-16

## 专题定位

`Workflow Draft Node Attribute Editing Model v1` 承接 `Workflow Draft Designer Editing Model v2`，把 Draft Designer 从本地图结构编辑推进到节点属性编辑。

本专题覆盖 provider / profile、tool ref、RAG ref、input / output contract fields 和 output mapping 的受控编辑、dev-only saved draft 保存映射、恢复映射和本地 validation / plan preview 消费。它不实现 workflow publish、run、executor、durable persistence、repository adapter、真实数据库、Radish OIDC 或 public production API。

状态：`workflow_draft_node_attribute_editing_model_v1_implemented`

## 当前实现

- `WorkflowDraftDesignerNode` 新增 `providerRef`、`toolRef`、`ragRef`、`inputContractFields`、`outputContractFields` 和 `outputMappingSummary`。
- Draft Designer 节点卡片新增节点属性编辑区，可编辑 provider / profile、tool ref、RAG ref、输入输出摘要、输入输出 contract fields 和 output mapping。
- 本地新增节点会按节点类型生成稳定的默认 provider / tool / contract fields / output mapping，避免新增节点进入空属性状态。
- `cloneWorkflowDraftForEditing` 深拷贝 contract fields 数组，避免本地编辑状态和原 draft 共享数组引用。
- saved draft consumer 在保存 payload 中写入节点级 summary、contract fields、output mapping、provider ref、tool ref 和 RAG ref，并聚合生成 global input / output contract required fields。
- saved draft restore 会把节点属性恢复到 Draft Designer 节点模型，读取失败仍按既有 fail-closed 语义处理。
- validation inspector 优先消费显式 contract fields，再结合 summary / output mapping 做补充识别。
- execution plan preview 优先消费节点自己的 provider / tool / policy ref，再落到 draft 默认 provider profile。
- `Workflow Review Handoff Active Draft v1` 已消费上述 active draft validation / plan / readiness 派生结果，生成 advisory-only active draft review record。
- platform dev-only saved draft schema 扩展节点级 `input_summary`、`output_summary`、`input_contract_fields`、`output_contract_fields` 和 `output_mapping_summary`，Go save / read / validate tests 覆盖节点属性保留。

## 数据边界

节点属性仍属于 saved draft 设计记录，不是 runtime binding，也不是生产执行配置。

允许保存：

- 节点输入 / 输出摘要。
- 节点级 input / output contract fields。
- provider / profile ref、tool ref、RAG ref。
- output mapping summary。

明确排除：

- secret value、API key value、OAuth / OIDC token。
- 工具执行结果、真实 provider response、confirmation decision。
- runtime executable plan、run input / output、writeback payload、replay / resume state。

## 验收方式

- Web build 通过，覆盖节点属性编辑表单、状态更新和下游 view model 类型。
- `go test ./...` 通过，覆盖 platform dev-only saved draft 节点属性保存、读取和校验。
- `workflow-draft-node-attribute-editing-model-v1` checker 固定字段模型、UI、consumer 映射、platform schema、文档引用和 fast baseline 顺序。
- `./scripts/check-repo.sh --fast` 通过。

## 后续方向

完成本专题后，Draft Designer 已具备字段编辑、用户工作区创建、saved draft list / restore、本地图结构编辑和节点属性编辑；`Workflow Review Handoff Active Draft v1` 已补恢复后的审查交接记录。下一批若继续 Workflow Builder 方向，应优先从以下方向择一推进：

- durable store 独立准入：只在 repository contract smoke 或 repository adapter implementation plan 中选择一个方向推进。
- 节点属性后续增强：仅在新增明确 schema、参数边界或执行前置评审时推进，不直接扩成 executor 配置。

## 停止线

- 不实现 publish、run、executor、node executor、tool executor、agent loop、confirmation decision、decision store、writeback、replay、resume 或 materialized result reader。
- 不接 durable persistence、repository adapter、真实数据库、schema migration、store selector、Radish OIDC middleware、token validation、API key lifecycle、quota enforcement、billing 或 public production API。
- 不把 provider / tool / RAG ref、contract fields、output mapping、validation summary、execution plan preview 或 runtime readiness inspector 解释为 runtime binding ready、publish ready、run ready 或 production ready。
