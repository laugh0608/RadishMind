# Saved Workflow Draft Conflict Review v1

更新时间：2026-06-30

状态：`workflow_saved_draft_conflict_review_v1_implemented`

## 文档目的

本文档定义并记录 saved workflow draft 在版本冲突、恢复和审查交接中的用户工作流。它承接现有 dev-only saved draft route、consumer `version_conflict` 状态、Draft Designer active draft 和 Review Handoff，不创建新的 runtime 或生产后端。

## 当前状态

- Draft Designer 已在 `version_conflict` 后展示 saved draft conflict review，保留本地 active draft，并展示 saved version、updated metadata、validation state 和 blocked capability count。
- 前端 consumer 已新增 `conflict_local_continued` 状态；用户显式选择继续编辑本地草案后，后续保存会使用当前 saved version 作为 expected version。
- Draft Designer 已提供显式恢复 saved version 的入口；恢复动作复用既有 dev-only read route 和 saved draft list summary，不由保存失败自动触发。
- 保存返回 `version_conflict` 后，Draft Designer 会刷新当前 application 的 sanitized saved draft list，为显式恢复 saved version 准备当前 metadata；该刷新不读取 secret、不恢复草案、不覆盖本地 active draft。
- Review Handoff 已消费同一份 conflict review summary，并以 advisory-only 形式展示冲突状态、saved version metadata、validation 状态、blocked capability 和 auto overwrite / auto merge 停止线。
- `workflow-saved-draft-consumer-smoke-v1` 与 `workflow-review-handoff-active-draft-v1` 已同步覆盖该实现；本批不新增 backend route、repository mode、数据库、runtime 或 public production API。

## 目标用户

- `Workspace Builder`：保存草案时遇到版本冲突，需要知道当前本地草案是否仍可继续编辑。
- `Workflow Reviewer`：审查恢复后的 active draft 时，需要看到冲突来源和后续处理建议。
- `Platform Maintainer`：确认冲突处理不绕过 owner / workspace、version conflict、no sample fallback 和 repository mode 停止线。

## 当前输入

- `POST /v1/user-workspace/workflow-drafts` 已能返回 `version_conflict`，前端 consumer 会保留本地草案并展示 saved draft version metadata。
- `GET /v1/user-workspace/workflow-drafts/{draft_id}` 已能恢复 saved draft 到 Draft Designer。
- Review Handoff 已能消费 active draft 的 validation、plan 和 readiness。
- `memory_dev` 是当前唯一可成功读写的 store mode；`repository` 和 `repository_disabled` 仍 fail closed。

## 已实现能力

1. Draft Designer 在 `version_conflict` 后展示冲突审查状态，说明本地草案、远端 saved version 和下一步选择。
2. 用户可以明确选择继续编辑本地草案，或从 saved draft 重新恢复；系统不得自动覆盖本地 active draft。
3. 冲突发生后会刷新当前 application 的 saved draft list，让恢复按钮基于已保存版本摘要可用；刷新过程仍只消费 sanitized summary。
4. Review Handoff 增加 conflict review summary，把冲突状态、saved version metadata、validation 状态和 blocked capability 一起交给 reviewer。
5. Workspace saved draft list 保持 sanitized summary，只提供恢复入口，不暴露 secret、token、完整 claim 或 runtime material。

## 后续开发

- 若继续用户工作流路径，应在用户显式启动 dev-only HTTP 服务后复核冲突审查、列表刷新、继续本地编辑、恢复 saved version 和 Review Handoff 展示体验。
- 若要扩大自动化验证，优先复用现有 workflow consumer smoke、Review Handoff checker、web build 和仓库基线；只有新增协议字段、route 行为或高风险边界时再新增专项 task card / fixture / checker。
- 若转回 durable store 上游，应独立推进 `storage_adapter_negative_leakage_scan_evidence_readiness`，不得把本功能实现解释为 repository mode、数据库、生产 API 或 runtime ready。

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
- 恢复 saved draft 必须是显式动作，不能由保存失败自动触发。
- Review Handoff 能显示 conflict review summary，并保持 advisory-only。
- 失败状态不得回退到 sample、fixture 或 memory dev 的其它草案。
- 现有 workflow consumer smoke、Review Handoff 检查和 web build 可复验本批实现；只有新增协议字段、route 行为或高风险边界时再新增专项 task card / fixture / checker。

## 停止线

- 不新增 backend route，不改 public production API，不启用 repository mode。
- 不创建 durable store、DB provider、SQL、schema marker runtime、migration runner、auth middleware、token validation 或 membership adapter。
- 不创建 workflow executor、publish / run、confirmation、writeback、replay 或 materialized result reader。
- 不把冲突审查状态写成自动合并、自动覆盖或执行解锁条件。
