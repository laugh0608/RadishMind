# User Workspace Draft Creation v1 任务卡

更新时间：2026-06-15

## 任务标识

- 切片：`user-workspace-draft-creation-v1`
- 轨道：`User Workspace / Workflow`
- 状态：`user_workspace_draft_creation_implemented`

## 目标

在 `Workflow Draft Editing Entry v1` 之后，为 Workspace Home 和 workflow definitions 增加创建本地草案入口，让用户能从工作区真实入口进入 Draft Designer，并继续使用已有 validate / save / read 链路。

本任务卡只承接前端本地草案创建入口，不声明 production workflow builder、durable persistence、public production API、publish、run、executor、confirmation decision、writeback 或 replay。

## 本轮实现

- `App.tsx` 新增 `workspaceCreatedDrafts` 状态，维护本地创建草案。
- `workflowWorkspaceContext` 增加 `localWorkflowDrafts` 输入，传入 Draft Designer view model。
- `workflowDraftDesigner` 合并 `localDrafts`，同时保持基础离线模板草案可渲染。
- Workspace Home application card 新增 `Create draft` 操作和 local draft 数量。
- Workflow Definition 行新增 `Create draft` 操作，并保留原有 selection row 行为。
- 创建后生成稳定短 ID、选中草案、进入 `local_edit` / `unsaved_local` 状态。
- 保存成功后把本地草案恢复为 `inspect_only`。
- 新增 `user-workspace-draft-creation-v1` fixture / checker，并接入 fast baseline。

## 输入事实源

- [User Workspace Draft Creation v1 专题](../features/workflow/user-workspace-draft-creation-v1.md)
- [Workflow Draft Editing Entry v1 专题](../features/workflow/draft-editing-entry-v1.md)
- [Workflow Draft Designer Surface 专题](../features/workflow/draft-designer-surface.md)
- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)

## 验收口径

- 创建入口必须来自 Workspace Home 或 workflow definitions，不绕过当前 selection context。
- 创建草案必须进入 `workspaceCreatedDrafts`，并通过 `localWorkflowDrafts` 纳入 workflow context。
- 创建后必须选中本地草案并进入 `unsaved_local`，validate / save / read 继续消费 active draft。
- 保存成功后本地草案不得继续保持 `local_edit`。
- `user-workspace-draft-creation-v1` checker、web build 和 `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 本任务卡不实现完整 builder、拖拽编排、节点新增 / 删除、workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、business writeback、run replay、run resume 或 materialized result reader；节点新增 / 删除 / 重排已由 `Workflow Draft Designer Editing Model v2` 独立承接。
- 不创建 public production API，不接真实数据库、repository adapter、schema migration、store selector、Radish OIDC、token validation、API key lifecycle、quota enforcement、billing 或 cost ledger。
- 不把本地创建草案、dev-only saved record、validation summary 或 `valid_for_review` 解释为 durable persistence、publish、run 或 production readiness。
