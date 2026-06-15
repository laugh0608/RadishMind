# Workflow Draft Editing Entry v1 专题

更新时间：2026-06-15

## 专题定位

`Workflow Draft Editing Entry v1` 是 `Workflow Draft Designer Surface` 从只读审查面走向可编辑草案面的第一批实现。它让用户在 Draft Designer 中直接编辑草案名称、说明、节点名称和边条件摘要，并把这些本地改动接入已有 dev-only saved draft validate / save / read consumer。

本专题不是完整拖拽 builder、workflow publish 或 executor 专题。它只固定“可控本地编辑状态 + dev-only 保存路径”的用户入口。

状态：`implemented`

## 当前实现

- Draft Designer 增加本地可编辑草案副本：切换 template 时重新克隆当前草案，避免直接修改离线 sample source。
- 页面提供受控输入：draft name、draft summary、node label 和 edge condition summary。
- 任一编辑会进入 `unsaved_local` 状态，并把草案标记为 `local_edit`。
- Reset 会丢弃本地改动并回到当前选中 template 的原始草案。
- Validate / Save / Read 全部使用 `activeWorkflowDraft`，不再只读取原始 `selectedWorkflowDraft`。
- validate / save / read 使用当前本地草案，这是本专题区别于只读审查面的核心验收点。
- Save 成功后清除 dirty 状态，并把本地副本恢复为 inspect-only saved view。
- 样式新增 `workflow-draft-edit-grid`、`workflow-draft-edit-field`、`workflow-draft-node-label-input` 和 `workflow-draft-edge-condition-input`，保证编辑入口在桌面与窄屏下保持稳定布局。

## 状态语义

页面需要区分：

- `inspect_only`：当前草案没有未保存本地编辑，可作为审查视图。
- `local_edit`：用户已修改本地副本，尚未确认写入 dev-only store。
- `unsaved_local`：saved draft consumer 状态显示存在未保存草案，可执行 validate 或 save。
- `saved_dev_record`：保存到 dev-only memory store 后的记录状态，不代表 durable persistence。
- `version_conflict`：保存发生版本冲突时保留当前本地草案，展示 current version metadata。

`local_edit` 只描述页面交互状态，不是生产持久化状态。

## 数据与保存边界

- 本批只允许编辑 draft label、summary、node label、edge condition summary。
- 保存 payload 继续由 `savedWorkflowDraftConsumer.ts` 的 sanitized mapping 生成。
- 本批不新增 HTTP route、schema、数据库表、repository adapter 或 production API。
- 本批不编辑 readiness、risk、blocked capability、route metadata、workflow definition、run record、validation result、execution plan 或 runtime readiness。
- 读取失败和保存失败继续 fail closed，不得回退成 sample 成功态。

## 验收方式

- Web build 通过。
- `workflow-draft-editing-entry-v1` checker 进入 fast baseline，固定编辑控件、active draft 保存路径、`local_edit` 类型、样式和文档停止线。
- `workflow-saved-draft-consumer-smoke-v1` 继续通过，确保 route contract、consumer smoke 和 version conflict guard 没有回退。
- `./scripts/check-repo.sh --fast` 通过。

## 下一批建议

完成本专题后，`User Workspace Draft Creation v1` 已补从 Workspace Home / workflow definitions 创建本地草案并进入 Draft Designer 的入口。Workflow 方向的后续批次应从以下方向择一推进：

- dev store 到未来 durable store 的迁移前置设计：先定义 store selector / repository adapter / schema migration 准入条件。
- Draft Designer 更完整的编辑模型：节点新增 / 删除 / 重排必须先补独立专题和任务卡，不能顺手并入当前入口。
- User Workspace saved draft list：读取已保存 dev draft 的列表视图，但必须先明确 no sample fallback、scope 和未来 durable store 边界。

## 停止线

- 不实现拖拽 builder、节点新增 / 删除、publish、run、executor、agent loop、confirmation decision、decision store、writeback、replay、resume 或 materialized result reader。
- 不接真实数据库、repository adapter、schema migration、store selector、Radish OIDC、token validation、API key lifecycle、quota enforcement、billing 或 public production API。
- 不把本地编辑、dev-only saved record、validation summary、risk summary 或 readiness summary 解释为 durable persistence ready、publish ready、run ready 或 production ready。
