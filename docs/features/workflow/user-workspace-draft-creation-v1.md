# User Workspace Draft Creation v1 专题

更新时间：2026-06-15

## 专题定位

`User Workspace Draft Creation v1` 承接 `Workflow Draft Editing Entry v1`，让用户能从 Workspace Home 或 workflow definition 列表创建一个本地草案，并直接进入 Draft Designer 编辑、校验和保存路径。

本专题不是 production workflow builder、durable persistence、publish 或 executor 专题。它只固定“用户工作区入口 -> 本地草案 -> Draft Designer -> dev-only saved draft consumer”的开发态产品路径。

状态：`implemented`

## 当前实现

- Workspace Home 的 application card 提供 `Create draft` 操作，并显示对应 workflow definition 下的 local draft 数量。
- Workflow Definitions 列表行提供 `Create draft` 操作，并保持原有选择行行为。
- 创建动作从对应 workflow definition 的基础 draft 派生本地草案，生成稳定 `draft_<workflow_definition_id>_workspace_<NN>` 短 ID。
- 本地草案写入 `workspaceCreatedDrafts` 前端状态，并通过 `localWorkflowDrafts` 进入 workflow workspace context。
- 创建后选中该草案，进入 `local_edit` / `unsaved_local` 状态，Draft Designer 继续使用现有 validate / save / read 路径。
- 保存成功后同步把本地草案恢复为 `inspect_only`，避免后续重新选中时误认为仍有未保存创建状态。

## 交互边界

- 创建草案只发生在前端本地状态，不调用新 API。
- 草案保存仍只走已有 dev-only saved draft consumer；没有启用 dev route 时仍是 local-only。
- 创建入口不新增节点、删除节点、拖拽排序或发布操作。
- 创建入口不写 workflow definition、run record、validation result、execution plan、runtime readiness、review handoff 或 scenario。

## 数据边界

本批本地草案只继承基础 draft 的可审查字段：

- `draftId`、`templateRef`、`applicationRef`、`workflowDefinitionId`、`providerProfileRef`
- draft label、summary、nodes、edges、readiness、risks、blockedCapabilities
- route request / audit metadata

本批不保存 secret、token、API key value、工具执行结果、confirmation decision、business writeback payload、run result、replay state 或 durable store record。

## 验收方式

- `user-workspace-draft-creation-v1` checker 进入 fast baseline，固定 Workspace Home / Workflow Definitions 创建入口、context local draft source、Draft Designer local draft merge、样式和停止线。
- `workflow-draft-editing-entry-v1` checker 继续通过，确保创建后的草案仍走 active draft validate / save / read。
- Web build 通过。
- `./scripts/check-repo.sh --fast` 通过。

## 下一批建议

完成本专题后，Workflow 方向下一批应在以下方向择一推进：

- durable store 迁移后续准入：在 [Saved Workflow Draft Durable Store Preconditions v1](saved-workflow-draft-durable-store-preconditions-v1.md) 已固定的停止线之后，选择 repository contract、schema migration 或 auth contract 中一个方向独立推进。
- Draft Designer 更完整编辑模型：节点新增 / 删除 / 重排需要独立专题、任务卡和更强验证。
- User Workspace saved draft list 已由 [User Workspace Saved Draft List v1](user-workspace-saved-draft-list-v1.md) 落地；后续可继续评估恢复后的审查 handoff 或更完整编辑模型。

## 停止线

- 不实现完整 builder、拖拽编排、节点新增 / 删除、publish、run、executor、agent loop、confirmation decision、decision store、writeback、replay、resume 或 materialized result reader。
- 不创建 public production API，不接真实数据库、repository adapter、schema migration、store selector、Radish OIDC、token validation、API key lifecycle、quota enforcement、billing 或 cost ledger。
- 不把本地创建草案、dev-only saved record、validation summary、risk summary 或 readiness summary 解释为 durable persistence ready、publish ready、run ready 或 production ready。
