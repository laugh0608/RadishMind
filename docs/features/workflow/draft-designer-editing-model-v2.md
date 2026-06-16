# Workflow Draft Designer Editing Model v2 专题

更新时间：2026-06-16

## 专题定位

`Workflow Draft Designer Editing Model v2` 承接 `Workflow Draft Editing Entry v1`、`User Workspace Draft Creation v1` 和 `User Workspace Saved Draft List v1`，把 Draft Designer 从字段级受控编辑推进到本地 workflow graph 的结构化编辑。

本专题只覆盖 Draft Designer 内的本地草案结构编辑和 dev-only saved draft consumer 接线，不声明完整拖拽 builder、workflow publish、run、executor、durable persistence 或 public production API。

状态：`workflow_draft_designer_editing_model_v2_implemented`

## 当前实现

- Draft Designer 新增结构编辑区，按 `prompt`、`llm`、`condition`、`http_tool` 和 `output` 五类节点添加本地节点。
- 新增节点会按 lane 插入到输出节点之前或输出区末尾，并生成稳定 `node_id`、label、input / output summary、risk 和 confirmation marker。
- 每个节点卡片支持 `Up`、`Down` 和 `Remove` 操作；删除会保护 context、model、policy / preview 组合和至少两个 output 节点，避免破坏可审查图的基本结构。
- 结构编辑会统一调用 graph rebuild helper，按当前节点顺序重建边，并确保至少保留 audit edge 供 validation inspector 审查。
- 任一结构编辑都会进入 `local_edit` / `unsaved_local`，后续 validate / save / read 继续消费 `activeWorkflowDraft`。
- validation inspector、execution plan preview 和 runtime readiness inspector 改为消费当前 active draft，避免结构编辑后下游预览仍展示旧图。
- dev-only saved draft restore 将 `http_tool` 恢复为 `preview` lane，和 offline designer 的 `http_tool` 语义保持一致。

## 交互与状态边界

- Add node：只创建本地 draft node，不创建 workflow definition，不写入上层业务真相源。
- Move node：只调整当前本地草案节点顺序，并重建预览边。
- Remove node：只允许删除非受保护结构；被保护的节点按钮保持 disabled，用户仍可通过 validation inspector 看到当前图状态。
- Reset：仍回到当前选中 saved / sample / workspace-created draft，不保留未保存结构编辑。
- Save：仍通过 `POST /v1/user-workspace/workflow-drafts` dev-only route，继续使用 version conflict、scope 和 no sample fallback 语义。

## 数据边界

结构编辑只修改 `WorkflowDraftDesignerDraft` 的 `nodes`、`edges` 和 `localOnlyInteraction`：

- 不修改 workflow definition、run record、confirmation placeholder、blocked capability truth source 或 user workspace 真实业务数据。
- 不保存 secret value、API key value、OAuth / OIDC token、工具执行结果、confirmation decision、run input / output、writeback payload、replay / resume state。
- 保存 payload 仍由 `savedWorkflowDraftConsumer.ts` 的 sanitized mapping 生成，后端继续执行 dev-only scope、schema version、failure code 和 validation summary 约束。

## 验收方式

- Web build 通过，验证新增结构编辑控件、active draft 派生 inspector / plan / readiness 和 saved restore lane mapping 的类型正确性。
- `workflow-draft-designer-editing-model-v2` checker 固定结构编辑 helper、UI 操作、下游 active draft 消费、样式和文档停止线。
- `workflow-draft-editing-entry-v1`、`user-workspace-draft-creation-v1`、`user-workspace-saved-draft-list-v1` 相关检查继续通过。
- `./scripts/check-repo.sh --fast` 通过。

## 后续方向

完成本专题后，Draft Designer 已具备字段编辑、用户工作区创建、保存列表恢复和本地图结构编辑能力；`Workflow Draft Node Attribute Editing Model v1` 已进一步补齐节点属性编辑、保存映射、恢复映射和下游 validation / plan 消费。下一批若继续 Workflow Builder 方向，应优先从以下方向择一推进：

- 恢复后的 Review Handoff：把 active draft validation / plan / readiness 汇总为可交接审查记录。
- durable store 独立准入：只在 repository contract smoke 或 repository adapter implementation plan 中选择一个方向推进。

## 停止线

- 不实现画布拖拽编排、分支 DSL 编辑器、节点执行参数完整表单、workflow publish、run、executor、agent loop、confirmation decision、decision store、writeback、replay、resume 或 materialized result reader。
- 不接真实数据库、repository adapter、schema migration、store selector、Radish OIDC middleware、token validation、API key lifecycle、quota enforcement、billing、cost ledger 或 public production API。
- 不把本地结构编辑、dev-only saved record、validation summary、execution plan preview 或 runtime readiness inspector 解释为 durable persistence ready、publish ready、run ready 或 production ready。
