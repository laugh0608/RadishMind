# User Workspace Saved Draft List v1 专题

更新时间：2026-06-16

## 专题定位

`User Workspace Saved Draft List v1` 承接本地草案创建、受控编辑和 dev-only saved draft consumer，让用户在 Workspace Home 中看到当前 application 下已保存 dev draft 的 sanitized summary，并可恢复到 Draft Designer 继续审查和编辑。

本专题不是 durable persistence、production workflow builder 或 public production API。列表只消费显式 dev-only HTTP route 和 memory dev store，恢复动作仍通过既有 read route 获取单个 saved draft record，不表示 workflow 可发布、可运行或可执行。

状态：`user_workspace_saved_draft_list_implemented`

## 当前实现

- 后端新增 `GET /v1/user-workspace/workflow-drafts` dev-only list route，继续要求 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP=1`、dev auth、workspace / application headers 和 `workflow_drafts:read` scope。
- `savedWorkflowDraftService.ListDrafts` 只按 workspace + application scope 返回 `SavedWorkflowDraftSummary`，HTTP envelope 字段为 `draft_summaries`；不返回完整 draft payload、secret、token、tool result、confirmation decision、run result 或 writeback payload。
- memory dev store 新增 `ListDraftsByScope`，只枚举当前 scope 下已保存草案；empty、scope denied 和 store unavailable 都 fail closed，不回退 sample。
- Web consumer 新增 `listWorkflowDraftDevRecords`、`restoreWorkflowDraftDevRecord` 和 `WorkflowSavedDraftListState`，区分 `sample`、`loading`、`ready`、`empty`、`list_failed` 和 `restore_failed`。
- Workspace Home 新增 saved draft list 区块，展示 summary、empty / failure state、refresh 和 restore 操作；restore 后把 saved record 投影为本地 Draft Designer 草案并选中。

## 数据边界

列表 summary 允许展示：

- `draft_id`、`workspace_id`、`application_id`、`source_definition_id`
- `draft_version`、`schema_version`、`draft_status`
- 草案名称、说明、`updated_at`、`updated_by_actor_ref`
- `node_count`、`edge_count`、`blocked_capability_count`
- `validation_state`、`valid_for_review`
- `sample_or_unsaved_draft_status`

列表 summary 不返回完整节点内容、边条件详情、secret value、API key value、OAuth / OIDC token、完整用户 claim、工具执行结果、confirmation decision、run input / output、materialized result、writeback payload、replay / resume state。

恢复动作只允许通过单个 `draft_id` 调用既有 read route。前端把 read route 返回的 saved record 转成本地 Draft Designer 草案；后续保存仍复用 `POST /v1/user-workspace/workflow-drafts` 的 version conflict 和 no sample fallback 语义。

## 失败语义

- `empty`：当前 application 没有 saved dev draft summary，不展示 sample 作为替代。
- `draft_scope_denied`：workspace / application header、query 或 dev auth scope 不匹配，列表为空且显示失败。
- `draft_store_unavailable`：dev store 不可用，列表为空且显示失败，不回退 sample。
- `restore_failed`：summary 存在但 read route 失败，保留当前本地草案，不把失败 summary 当成 saved draft。
- `version_conflict`：恢复后的再次保存仍走既有 conflict 状态，保留本地草案并展示 current version metadata。

## 验收方式

- Go tests 覆盖 list summary、empty、scope、store unavailable、no sample fallback 和 no side effects。
- Web build 覆盖 consumer list / restore 类型和 Workspace Home 渲染。
- `user-workspace-saved-draft-list-v1` checker 固定 route contract、consumer 状态、App restore flow、Workspace Home UI 和文档引用。
- `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不实现 durable persistence、repository adapter、真实数据库、schema migration、store selector、Radish OIDC middleware、token validation、public production API、API key lifecycle、quota enforcement、billing 或 cost ledger。
- 本专题不实现完整 builder、节点新增 / 删除 / 拖拽编排、publish、run、executor、agent loop、confirmation decision、decision store、writeback、replay、resume 或 materialized result reader；节点新增 / 删除 / 重排已由 `Workflow Draft Designer Editing Model v2` 独立承接。
- 不把 saved draft list、restore、`valid_for_review`、validation summary 或 risk summary 解释为 durable persistence ready、publish ready、run ready 或 production ready。
