# Saved Workflow Draft Conflict Review v1

更新时间：2026-07-10

状态：`workflow_saved_draft_conflict_review_v1_implemented`

## 文档目的

本文档定义并记录 saved workflow draft 在版本冲突、恢复和审查交接中的用户工作流。它承接现有 dev-only saved draft route、consumer `version_conflict` 状态、Draft Designer active draft 和 Review Handoff，不创建新的 runtime 或生产后端。

## 当前状态

- Draft Designer 已在 `version_conflict` 后展示 saved draft conflict review，保留本地 active draft，并展示 saved version、updated metadata、validation state 和 blocked capability count。
- 前端 consumer 已新增 `conflict_local_continued` 状态；用户显式选择继续编辑本地草案后，后续保存会使用当前 saved version 作为 expected version。
- 2026-07-10 已修正 saved version 生命周期：已保存草案进入 `unsaved_local` 或 `validation_ready` 后继续保留 persisted base version；未处理的 `version_conflict` 会阻止 validate / save / read，只有显式 Continue 或 Restore 后才能继续 dev route 动作。
- Draft Designer 已提供显式恢复 saved version 的入口；恢复动作复用既有 dev-only read route 和 saved draft list summary，不由保存失败自动触发。
- 保存返回 `version_conflict` 后，Draft Designer 会刷新当前 application 的 sanitized saved draft list，为显式恢复 saved version 准备当前 metadata；该刷新不读取 secret、不恢复草案、不覆盖本地 active draft。
- Review Handoff 已消费同一份 conflict review summary，并以 advisory-only 形式展示冲突状态、saved version metadata、validation 状态、blocked capability 和 auto overwrite / auto merge 停止线。
- 2026-07-01 已整理冲突审查卡片与 Review Handoff 可读性：前端 conflict review summary 现在显式派生 `savedMetadataLoaded`、`savedMetadataState`、`restoreActionState`、`restoreUnavailableReason`、本地草案保留说明和 reviewer 下一步；恢复入口在 metadata 刷新中、列表为空、列表失败或缺少匹配 summary 时保持禁用，并说明本地草案仍未被覆盖或合并。
- `workflow-saved-draft-consumer-smoke-v1`、Web lifecycle behavior test 与 `workflow-review-handoff-active-draft-v1` 已同步覆盖该实现；本批不新增 checker、backend route、repository mode、数据库、runtime 或 public production API。
- 2026-06-30 dev-live 浏览器复核已覆盖正常保存、外部版本推进、UI 冲突保存、冲突后列表刷新、继续本地草案、显式恢复 saved version 和 Review Handoff 摘要展示；复核期间只出现 `favicon.ico` 404，不影响 workflow 功能。

## 目标用户

- `Workspace Builder`：保存草案时遇到版本冲突，需要知道当前本地草案是否仍可继续编辑。
- `Workflow Reviewer`：审查恢复后的 active draft 时，需要看到冲突来源和后续处理建议。
- `Platform Maintainer`：确认冲突处理不绕过 owner / workspace、version conflict、no sample fallback 和 repository mode 停止线。

## 当前输入

- `POST /v1/user-workspace/workflow-drafts` 已能返回 `version_conflict`，前端 consumer 会保留本地草案并展示 saved draft version metadata。
- `GET /v1/user-workspace/workflow-drafts/{draft_id}` 已能恢复 saved draft 到 Draft Designer。
- Review Handoff 已能消费 active draft 的 validation、plan 和 readiness。
- `memory_dev` 是当前唯一可成功读写的 store mode；`repository` 和 `repository_disabled` 仍 fail closed。

## 用户流程

1. 用户在 Draft Designer 编辑 active draft，并通过 dev-only saved draft consumer 保存。
2. 如果后端返回 `version_conflict`，页面保留本地 active draft，不自动读取 saved version，也不覆盖本地节点、边、layout 或属性编辑。
3. 页面刷新当前 application 的 saved draft list，只用 sanitized summary 准备恢复入口；metadata 未就绪时恢复入口应保持不可用。
4. 用户可以继续本地草案；此时状态进入 `conflict_local_continued`，后续保存使用当前 saved version 作为 expected version。
5. 用户也可以显式恢复 saved version；恢复动作复用 read route，并把恢复结果带回 Draft Designer。
6. Reviewer 在 Review Handoff 中查看 conflict review summary，确认冲突来源、saved version metadata、validation 状态、blocked capability 和停止线。

## 界面读法

- 冲突状态首先说明“本地草案仍保留”，再说明 saved version 的版本号、更新时间、更新人、校验状态和 blocked capability 数量。
- 恢复入口必须显示 metadata 状态，以及当前是 `restore_available` 还是 `restore_requires_saved_list`；metadata 未加载、列表为空、列表失败或缺少匹配 summary 时，应说明恢复入口等待 sanitized saved draft list 重新可用。
- “继续本地草案”代表用户选择先保留当前编辑上下文，不代表系统已经合并远端版本。
- “恢复 saved version”代表用户主动把 saved draft 读回 Draft Designer；该动作必须可区分于保存失败后的自动 fallback。
- Review Handoff 的 conflict summary 是审查证据，不是发布、执行、确认或业务写回前置条件。

## 已实现能力

1. Draft Designer 在 `version_conflict` 后展示冲突审查状态，说明本地草案、远端 saved version 和下一步选择。
2. 用户可以明确选择继续编辑本地草案，或从 saved draft 重新恢复；系统不得自动覆盖本地 active draft。
3. 冲突发生后会刷新当前 application 的 saved draft list，让恢复按钮基于已保存版本摘要可用；刷新过程仍只消费 sanitized summary。
4. Review Handoff 增加 conflict review summary，把冲突状态、saved version metadata、validation 状态和 blocked capability 一起交给 reviewer。
5. Workspace saved draft list 保持 sanitized summary，只提供恢复入口，不暴露 secret、token、完整 claim 或 runtime material。
6. Draft Designer 与 Review Handoff 共用同一份恢复状态、metadata 状态、本地草案保留说明和 reviewer 下一步文案；这些字段只属于前端派生摘要，不改变 dev-only route、schema 或 runtime 边界。

## 后续开发

- 若继续用户工作流路径，应基于实际审查反馈继续做小范围阅读路径整理；不得重复把普通 UI 体验整理升级为新的生产后端能力。
- 若要扩大自动化验证，优先复用现有 workflow consumer smoke、Review Handoff checker、web build 和仓库基线；只有新增协议字段、route 行为或高风险边界时再新增专项 task card / fixture / checker。
- 若转回 durable store 上游，应先独立推进 `storage_adapter_metadata_contract_artifact_materialization_entry_review`，不得把本功能实现解释为 repository mode、数据库、生产 API 或 runtime ready。

## 数据边界

允许使用：

- `draft_id`、`draft_version`、`updated_at`、`updated_by_actor_ref`、`workspace_id`、`application_id`。
- `version_conflict` failure code、当前 saved draft metadata、active draft validation summary、readiness summary 和 audit metadata reference。
- 本地 active draft 的草案名称、说明、节点、边和受控 layout metadata。

禁止使用：

- secret value、API key value、OAuth / OIDC token、完整用户 claim。
- force overwrite、auto merge、runtime execution、publish、confirmation decision、writeback、replay 或 materialized result。
- repository mode、DB provider、SQL migration、schema marker runtime 或 public production API。

## 验收方式

- `version_conflict` 后本地 active draft 不丢失，用户能看到冲突来源和 saved version metadata。
- 保存冲突后只刷新当前 application 的 saved draft list，并在 metadata 未就绪时保持恢复入口禁用。
- 冲突审查卡片和 Review Handoff 都能展示 metadata state、restore action state、本地草案保留说明和 reviewer 下一步。
- 恢复 saved draft 必须是显式动作，不能由保存失败自动触发。
- Review Handoff 能显示 conflict review summary，并保持 advisory-only。
- 失败状态不得回退到 sample、fixture 或 memory dev 的其它草案。
- 现有 workflow consumer smoke、Review Handoff 检查和 web build 可复验本批实现；只有新增协议字段、route 行为或高风险边界时再新增专项 task card / fixture / checker。
- 已保存草案经过本地编辑或 validate 后再次保存，必须继续使用原 persisted base version；validate 的非持久化 `current_draft_version=0` 不得覆盖本地 saved baseline。
- `version_conflict` 未显式处理前，Validate / Save / Read 和草案编辑控件保持禁用，不能通过重复 Save 绕过冲突审查。
- dev-live 浏览器复核记录应能证明本地 active draft 不被自动覆盖、冲突后列表刷新只准备恢复入口、恢复 saved version 必须显式触发，且 Review Handoff 仍只提供 advisory-only 审查语义。

## 停止线

- 不新增 backend route，不改 public production API，不启用 repository mode。
- 不创建 durable store、DB provider、SQL、schema marker runtime、migration runner、auth middleware、token validation 或 membership adapter。
- 不创建 workflow executor、publish / run、confirmation、writeback、replay 或 materialized result reader。
- 不把冲突审查状态写成自动合并、自动覆盖或执行解锁条件。
