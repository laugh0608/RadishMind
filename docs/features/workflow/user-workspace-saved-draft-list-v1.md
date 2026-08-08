# User Workspace Saved Draft List v1 专题

更新时间：2026-08-08

## 专题定位

`User Workspace Saved Draft List v1` 承接本地草案创建、受控编辑和 dev-only saved draft consumer，让用户在 Workspace Home 中看到当前 application 下已保存 dev draft 的 sanitized summary，并可打开到 Draft Designer 继续审查和编辑。

本专题不是 production workflow builder 或 public production API。列表只消费显式 dev-only HTTP route 和已配置的开发测试态 store，打开动作仍通过既有 read route 获取单个 saved draft record，不表示 workflow 可发布、可运行或可执行。

状态：`user_workspace_saved_draft_list_implemented`

## 当前实现

- 后端新增 `GET /v1/user-workspace/workflow-drafts` dev-only list route，继续要求 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP=1`、dev auth、workspace / application headers 和 `workflow_drafts:read` scope。
- `savedWorkflowDraftService.ListDrafts` 只按 workspace + application scope 返回 `SavedWorkflowDraftSummary`，HTTP envelope 字段为 `draft_summaries`；不返回完整 draft payload、secret、token、tool result、confirmation decision、run result 或 writeback payload。
- Saved Draft store 使用 `ListDraftSummariesByScope`，只枚举当前 scope 下已保存草案；empty、scope denied 和 store unavailable 都 fail closed，不回退 sample。
- Web consumer 提供 `listWorkflowDraftDevRecords`、`openWorkflowDraftDevRecord` 和 `WorkflowSavedDraftListState`，区分 `sample`、`loading`、`ready`、`empty`、`list_failed` 和 `open_failed`。
- Workspace Home 的 Saved Draft Library 提供活动 / 归档、name prefix / validation / provenance 筛选、加载更多、两步 archive 和显式 unarchive；活动草案可打开，归档草案只读审查。
- 活动与归档查询分别保存 lifecycle、严格 cursor、筛选、pending 和迟到响应 generation；列表不会把不同查询窗口拼接成全量统计。
- Draft Designer 在保存返回 `version_conflict` 后复用同一 list consumer 刷新当前 application 的 sanitized summary，并据此派生打开 saved draft 是否可用；该刷新不读取完整草案、不覆盖本地 active draft。
- S3 Workbench 保持本 Library 为唯一完整挂载点；Designer 只显示紧凑活动草案引用和 `Open library` 交接。活动 / 归档标签继续使用 action token 表达当前视图，并具有 `tablist`、`tab` 与 `aria-selected` 语义，lifecycle 状态色不承担选中。

## S3 产品化衔接

Saved Draft Library 在 S3 中继续作为 `B` 级页面复用现有列表和生命周期模式，不新增后端 owner。桌面负责提供活动 / 归档范围、筛选、当前草案摘要和精确打开入口，Designer 画布仍是进入后的唯一主任务面；窄屏按 Library → Designer → Inspector 的任务前后关系切换单列，不把列表、画布和 inspector 同时压缩。

活动草案使用“打开草案”进入可编辑 Designer；归档草案使用“只读审查”。解除归档后只刷新活动 / 归档列表，不自动打开、恢复或重放旧浏览器状态，用户必须从活动列表重新精确打开最新内容版本与 lifecycle 版本。

## 数据边界

列表 summary 允许展示：

- `draft_id`、`workspace_id`、`application_id`、`source_definition_id`
- `draft_version`、`schema_version`、`draft_status`
- 草案名称、说明、`updated_at`、`updated_by_actor_ref`
- `node_count`、`edge_count`、`blocked_capability_count`
- `validation_state`、`valid_for_review`
- `sample_or_unsaved_draft_status`

列表 summary 不返回完整节点内容、边条件详情、secret value、API key value、OAuth / OIDC token、完整用户 claim、工具执行结果、confirmation decision、run input / output、materialized result、writeback payload、replay / resume state。

打开动作只允许通过单个 `draft_id` 调用既有 read route。前端把 read route 返回的 saved record 转成本地 Draft Designer 草案；活动态后续保存仍复用 `POST /v1/user-workspace/workflow-drafts` 的双版本 conflict 和 no sample fallback 语义，归档态保持只读。

冲突审查只把 saved draft list 当作 metadata 来源。列表 summary 可用于显示 saved version、更新时间、validation state 和 blocked capability count，也可让打开按钮进入 `open_available`；但只要列表处于 `loading`、`empty`、`list_failed`、`sample` 或缺少匹配 summary，打开入口就必须保持禁用并说明本地草案仍被保留。

## 失败语义

- `empty`：当前 application 没有 saved dev draft summary，不展示 sample 作为替代。
- `draft_scope_denied`：workspace / application header、query 或 dev auth scope 不匹配，列表为空且显示失败。
- `draft_store_unavailable`：dev store 不可用，列表为空且显示失败，不回退 sample。
- `open_failed`：summary 存在但 read route 失败，保留当前本地草案，不把失败 summary 当成 saved draft。
- `version_conflict`：打开后的再次保存仍走既有双版本 conflict 状态，保留本地草案并展示 current version metadata。
- `open_requires_saved_list`：冲突后还没有可用匹配 summary，打开 saved draft 不可触发；用户仍可继续本地草案，或等待 / 手动刷新列表后再打开。

## 验收方式

- Go tests 覆盖 list summary、empty、scope、store unavailable、no sample fallback 和 no side effects。
- Web build 覆盖 consumer list / open 类型和 Workspace Home 渲染。
- `user-workspace-saved-draft-list-v1` checker 固定 route contract、consumer 状态、App open flow、Workspace Home UI 和文档引用。
- `./scripts/check-repo.sh --fast` 通过。
- 2026-08-08 应用内浏览器已复验活动列表打开、派生草案切换、两步归档、归档只读、解除归档、活动列表重新打开和 `390x844` 单列布局；页面无横向溢出，控制台零 warning / error。

## 停止线

- 本 S3 产品化批次不新增或替换 memory、SQLite、PostgreSQL 开发测试态 owner，不启用 production repository、production auth、public production API、API key lifecycle、quota enforcement、billing 或 cost ledger。
- 本专题不实现完整 builder、节点新增 / 删除 / 拖拽编排、publish、run、executor、agent loop、confirmation decision、decision store、writeback、replay、resume 或 materialized result reader；节点新增 / 删除 / 重排已由 `Workflow Draft Designer Editing Model v2` 独立承接。
- 不把 saved draft list、open、`valid_for_review`、validation summary 或 risk summary 解释为 publish ready、run ready 或 production ready。
- 不把冲突后列表刷新解释为自动打开、自动覆盖、自动合并或 production repository mode；打开 saved draft 必须由用户显式操作。
