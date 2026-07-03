# Workflow Review Handoff Active Draft v1 专题

更新时间：2026-07-01

## 专题定位

`Workflow Review Handoff Active Draft v1` 承接 Draft Designer 已有的 saved draft list / restore、本地图结构编辑和节点属性编辑，把当前 active draft 的 validation inspector、execution plan preview 和 runtime readiness inspector 汇总成可交接审查记录。

本专题不创建 review / handoff 持久化结果，不导出或发送交接包，也不实现 publish、run、executor、confirmation decision、writeback、replay、durable persistence、repository adapter、数据库、OIDC 或 public production API。

状态：`workflow_review_handoff_active_draft_v1_implemented`

## 当前实现

- `workflowWorkspaceContext` 继续作为唯一派生入口，active draft 先进入 validation inspector，再进入 execution plan preview 和 runtime readiness inspector。
- `WorkflowReviewHandoffViewModel` 新增 active draft review record，显式记录 validation / plan / readiness 三段审查来源、状态、blocker count、request / audit metadata 和 reviewer question。
- Review Handoff 面板新增 `Active Draft Review Record` 区域，供人工审查者先确认当前 draft 的 validation、plan 和 readiness 证据链。
- Review Handoff 现在消费 Draft Designer 的 `WorkflowSavedDraftConflictReviewSummary`，在 `version_conflict` 后显示 saved metadata state、restore action state、本地草案保留说明、reviewer 下一步和 auto overwrite / auto merge 停止线。
- handoff record 只由当前浏览器内 view model 派生，不保存、不导出、不发送、不请求 live backend。
- 新增 `workflow-review-handoff-active-draft-v1` fixture / checker，固定 handoff 对 active draft 三段 inspector 的消费链、UI 渲染和停止线。

## 数据边界

允许汇总：

- active draft id、validation status、execution plan preview status、runtime readiness status。
- validation structural / contract / blocked capability summary。
- plan stage order、provider requirement、confirmation / audit gate 和 blocked reason summary。
- readiness prerequisite、blocker 和 implementation gate summary。
- request id、audit ref、source page id 和本地 review question。
- saved draft conflict failure code、saved version metadata state、restore action state、本地草案保留说明和 reviewer next step。

明确排除：

- secret value、API key value、OAuth / OIDC token、authorization header、cookie。
- raw prompt dump、raw tool payload、provider response body。
- review / handoff persistence、execution plan persistence、runtime readiness persistence。
- confirmation decision、run input / output、tool execution result、business writeback payload、replay / resume state。

## 功能流程

1. 用户从 Workspace Home 或 saved dev draft list 恢复草案后，Draft Designer 把当前草案作为 active draft。
2. validation inspector 从 active draft 派生结构、契约和 blocked capability 检查。
3. execution plan preview 消费同一 active draft 和 validation inspector，生成离线 stage / provider / gate / blocked reason 预览。
4. runtime readiness inspector 消费 execution plan preview，生成 executor、provider、confirmation、store、audit、writeback、replay、auth / store 和 publish 前置状态。
5. Review Handoff 汇总上述三段 view model，形成 active draft review record，并与 recipients、key findings、evidence checklist、decision blockers 和 boundary locks 一起展示。
6. 如果 Draft Designer 最近一次保存进入 `version_conflict`，Review Handoff 额外展示同一份 conflict review summary，供 reviewer 判断当前应继续本地草案、等待 metadata 刷新，还是显式恢复 saved version。

## 验收方式

- Web build 通过，覆盖 Review Handoff 新增 active draft review record 类型和面板渲染。
- `workflow-review-handoff-active-draft-v1` checker 固定 handoff source、record sections、conflict summary、UI 文案、文档引用和 fast baseline 顺序。
- `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不保存、不导出、不发送 handoff record，不创建 review repository、handoff repository、database table 或 public production API。
- 不持久化 validation result、execution plan preview 或 runtime readiness result。
- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、writeback、replay、resume 或 materialized result reader。
- 不接 durable persistence、repository adapter、真实数据库、schema migration、store selector、Radish OIDC middleware、token validation、API key lifecycle、quota enforcement、billing 或 public production API。
- 不把 active draft review record、validation summary、execution plan preview、runtime readiness inspector 或 `valid_for_review` 解释为 runtime binding ready、publish ready、run ready 或 production ready。
- 不把 conflict review summary 解释为 handoff 持久化、执行解锁、自动覆盖、自动合并或恢复 saved version 的隐式确认。
