# User Workspace Saved Draft List v1 任务卡

更新时间：2026-06-16

## 任务标识

- 切片：`user-workspace-saved-draft-list-v1`
- 轨道：`User Workspace / Workflow`
- 状态：`user_workspace_saved_draft_list_implemented`

## 目标

在 `User Workspace Draft Creation v1` 和 dev-only saved draft consumer 之后，为 Workspace Home 增加已保存 dev draft 的列表与恢复入口。用户可以看到当前 application 下的 sanitized saved draft summary，并恢复到 Draft Designer 继续审查和编辑。

本任务卡只承接 dev-only list route、consumer list / restore 和前端恢复入口，不声明 durable persistence、public production API、workflow publish、run、executor、confirmation decision、writeback 或 replay。

## 本轮实现

- `SavedWorkflowDraftSummary` 和 `SavedWorkflowDraftListResult` 固定列表 projection。
- memory dev store 新增 `ListDraftsByScope`，只按 workspace + application scope 枚举 saved draft。
- 后端新增 `GET /v1/user-workspace/workflow-drafts`，返回 `draft_summaries`、`failure_code`、request / audit metadata。
- Web consumer 新增 `WorkflowSavedDraftListState`、`listWorkflowDraftDevRecords` 和 `restoreWorkflowDraftDevRecord`。
- Workspace Home 新增 saved draft list 区块，展示 `sample` / `loading` / `ready` / `empty` / `list_failed` / `restore_failed`，并提供 refresh 和 restore。
- restore 通过既有 read route 获取单个 saved record，再投影为本地 Draft Designer 草案；再次保存仍使用已有 version conflict 语义。
- 新增 `user-workspace-saved-draft-list-v1` fixture / checker，并接入 fast baseline。

## 输入事实源

- [User Workspace Saved Draft List v1 专题](../features/workflow/user-workspace-saved-draft-list-v1.md)
- [User Workspace Draft Creation v1 专题](../features/workflow/user-workspace-draft-creation-v1.md)
- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)
- [Dev-only Saved Draft Consumer 专题](../features/workflow/dev-only-saved-draft-consumer.md)

## 验收口径

- list route 必须显式 dev-only，默认关闭，且要求 dev auth、workspace / application header 和 `workflow_drafts:read` scope。
- list response 只返回 sanitized summary，不返回完整 draft body、secret、token、tool result、confirmation decision、run result 或 writeback payload。
- empty、scope denied 和 store unavailable 必须 fail closed，不回退 sample。
- Workspace Home 必须展示 saved draft list 状态、failure code、summary 数量、refresh 和 restore。
- restore 必须通过 read route 获取单个 saved record，并进入 Draft Designer 的本地草案状态；不得绕过 version conflict 或 no sample fallback 语义。
- Go tests、web build、`user-workspace-saved-draft-list-v1` checker 和 `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不实现 durable persistence、repository adapter、真实数据库、schema migration、store selector、Radish OIDC、token validation、API key lifecycle、quota enforcement、billing、cost ledger 或 public production API。
- 本任务卡不实现完整 builder、拖拽编排、节点新增 / 删除、workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、business writeback、run replay、run resume 或 materialized result reader；节点新增 / 删除 / 重排已由 `Workflow Draft Designer Editing Model v2` 独立承接。
- 不把 saved draft list、restore、dev-only memory store、validation summary 或 `valid_for_review` 解释为 durable persistence、publish、run 或 production readiness。
