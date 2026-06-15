# Workflow Draft Editing Entry v1 任务卡

更新时间：2026-06-15

## 任务标识

- 切片：`workflow-draft-editing-entry-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_draft_editing_entry_implemented`

## 目标

在已完成 dev-only saved draft consumer 的基础上，为 Draft Designer 增加受控本地编辑入口，让用户能编辑草案名称、说明、节点名称和边条件摘要，并通过现有 validate / save / read 操作提交当前本地草案。

本任务卡只承接 Draft Designer 的本地编辑入口，不声明完整 builder、durable persistence、public production API、publish、run、executor、confirmation decision、writeback 或 replay。

## 本轮实现

- `apps/radishmind-web/src/app/App.tsx` 增加 `editableWorkflowDraft` 与 `workflowDraftEditDirty` 状态。
- template 切换时克隆当前草案，避免直接修改离线 sample source。
- 新增 draft name、draft summary、node label、edge condition summary 的受控输入。
- 任一编辑会进入 `local_edit` / `unsaved_local` 状态，并保留 validate / save / read 按钮的 pending 禁用行为。
- validate / save / read 使用 `activeWorkflowDraft`，保存成功后清除 dirty 状态。
- `apps/radishmind-web/src/features/control-plane-read/workflowDraftDesigner.ts` 将 `localOnlyInteraction` 扩展为 `inspect_only | local_edit`。
- `apps/radishmind-web/src/styles.css` 增加编辑区、节点名称输入和边条件输入样式，并纳入窄屏响应式布局。
- 新增 `workflow-draft-editing-entry-v1` fixture / checker，并接入 fast baseline。

## 输入事实源

- [Workflow Draft Editing Entry v1 专题](../features/workflow/draft-editing-entry-v1.md)
- [Workflow Draft Designer Surface 专题](../features/workflow/draft-designer-surface.md)
- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)
- [Workflow Saved Draft Consumer Integration v1 任务卡](workflow-saved-draft-consumer-integration-v1-plan.md)

## 验收口径

- 编辑入口必须是受控输入，不能用临时 DOM 读取或 UI-only 字段拼 payload。
- 本地编辑必须显式进入 `unsaved_local`，并保留 reset 能力。
- validate / save / read 必须消费当前本地草案。
- 保存成功后 dirty 状态必须清除；version conflict 时必须保留当前本地草案。
- 样式必须覆盖桌面与窄屏，不因按钮、输入框或长文本导致主要布局溢出。
- `workflow-draft-editing-entry-v1` checker、web build 和 `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不实现拖拽 builder、节点新增 / 删除、workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、business writeback、run replay、run resume 或 materialized result reader。
- 不创建 public production API，不接真实数据库、repository adapter、schema migration、store selector、Radish OIDC、token validation、API key lifecycle、quota enforcement、billing 或 cost ledger。
- 不把本地编辑、dev-only saved record、validation summary 或 `valid_for_review` 解释为 durable persistence、publish、run 或 production readiness。
