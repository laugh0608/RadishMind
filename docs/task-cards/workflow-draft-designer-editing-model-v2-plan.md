# Workflow Draft Designer Editing Model v2 任务卡

更新时间：2026-06-16

## 任务标识

- 切片：`workflow-draft-designer-editing-model-v2`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_draft_designer_editing_model_v2_implemented`

## 目标

在 Draft Designer 已支持字段级编辑、用户工作区创建和 saved draft list / restore 之后，补齐本地草案图的结构编辑能力，让用户能在 Draft Designer 中添加节点、移动节点、删除非受保护节点，并让校验、计划预览和 readiness 继续基于当前 active draft。

本任务卡只承接本地 graph edit model、UI 操作和 dev-only saved draft consumer 的继续复用，不声明完整 builder、durable persistence、public production API、publish、run、executor、confirmation decision、writeback 或 replay。

## 本轮实现

- Draft Designer 增加 `workflow-draft-structure-controls` 和 `workflow-draft-add-node-grid`，提供五类节点添加入口。
- `App.tsx` 新增 `buildLocalWorkflowDraftNode`、`workflowDraftWithStructureEdits`、`insertWorkflowDraftNode`、`moveWorkflowDraftNode`、`canRemoveWorkflowDraftNode` 和 `rebuildWorkflowDraftEdges` 等 graph helper。
- 节点卡片增加 `Up`、`Down`、`Remove` 操作；删除保护至少保留 context、model、policy / preview 组合和两个 output 节点。
- validation inspector、execution plan preview 和 runtime readiness inspector 使用当前 `activeWorkflowDraft` 派生结果。
- saved draft restore 将 `http_tool` 映射为 `preview` lane。
- 新增 `workflow-draft-designer-editing-model-v2` fixture / checker，并接入 fast baseline。

## 输入事实源

- [Workflow Draft Designer Editing Model v2 专题](../features/workflow/draft-designer-editing-model-v2.md)
- [Workflow Draft Editing Entry v1 专题](../features/workflow/draft-editing-entry-v1.md)
- [User Workspace Draft Creation v1 专题](../features/workflow/user-workspace-draft-creation-v1.md)
- [User Workspace Saved Draft List v1 专题](../features/workflow/user-workspace-saved-draft-list-v1.md)
- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)

## 验收口径

- 新增、移动和删除操作必须统一标记 `local_edit` / `unsaved_local`。
- 结构编辑必须重建 `edges`，并显式保留 audit edge，避免保存 payload 中出现孤立图或缺失审查路径。
- 删除操作必须保护必要 lane，不允许把草案编辑成不可审查的空图。
- validate / save / read、validation inspector、execution plan preview 和 runtime readiness inspector 必须消费当前 active draft。
- Web build、专项 checker 和 `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不实现画布拖拽编排、分支 DSL 编辑器、完整节点参数表单、workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、business writeback、run replay、run resume 或 materialized result reader。
- 不实现 durable persistence、repository adapter、真实数据库、schema migration、store selector、Radish OIDC middleware、token validation、API key lifecycle、quota enforcement、billing、cost ledger 或 public production API。
- 不把本地结构编辑、saved dev draft、validation summary、execution plan preview 或 `valid_for_review` 解释为 durable persistence、publish、run 或 production readiness。
