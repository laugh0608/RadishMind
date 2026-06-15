# Dev-only Saved Draft Consumer 实现专题

更新时间：2026-06-15

## 专题定位

本专题承接 `Saved Workflow Draft v1` 的下一批 consumer integration。目标是在开发态把已实现的 platform Go domain service 暴露给 web consumer，让 `apps/radishmind-web/` 能保存、读取或校验草案，并明确区分 sample / unsaved draft 与 saved draft record。

这是 dev-only / non-production 实现专题，不是 public production API、durable database persistence、workflow publish 或 executor。

状态：`implemented`

## 当前实现

已实现 `dev-only HTTP route + web consumer`：

- 后端 route：`POST /v1/user-workspace/workflow-drafts`、`GET /v1/user-workspace/workflow-drafts/{draft_id}`、`POST /v1/user-workspace/workflow-drafts/validate`。
- 后端显式开关：`RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP=1` 才允许访问 route；`RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE=1` 才允许保存。
- dev auth 继续要求 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1`，并通过 `X-RadishMind-Dev-Workflow-Workspace`、`X-RadishMind-Dev-Workflow-Application`、subject、tenant 和 scope headers 固定开发态 scope。
- 前端 consumer：`apps/radishmind-web/src/features/control-plane-read/savedWorkflowDraftConsumer.ts`，默认 sample-only；只有 `VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE=dev-saved-draft-http` 才调用 dev route。
- 页面状态：`sample`、`unsaved_local`、`saving`、`validating`、`reading`、`saved_dev_record`、`validation_ready`、`version_conflict`、`save_failed`、`read_failed`、`validation_failed`。
- 路径稳定化：已补 route contract、consumer smoke 和 `version_conflict` UI 状态；冲突时展示当前 saved draft version metadata，保留用户当前本地草案，不回退 sample。

## 推荐实现路径

默认路径为 `dev-only HTTP route + web consumer`。

候选 route 形态：

- `POST /v1/user-workspace/workflow-drafts`：保存草案。
- `GET /v1/user-workspace/workflow-drafts/{draft_id}`：读取草案。
- `POST /v1/user-workspace/workflow-drafts/validate`：只校验草案，不保存。

如果实现时发现 route 风险或范围过大，可以先落 `web 本地提交 / saved record 区分`，但仍必须让 UI 状态和 consumer contract 明确区分 sample、unsaved local draft、saved dev draft 和 failed saved draft。

## 准入条件

- `SavedWorkflowDraft` domain service 和测试保持通过。
- route 必须由显式 dev-only 配置打开，默认环境不得启用写入。
- dev auth 必须显式提供 workspace、application、actor、tenant 或等价 scope 上下文。
- save 必须要求 write enablement；只读或 sample 模式返回 `draft_write_disabled`。
- read 必须按 workspace + application + draft scope 查询，不允许跨 scope 返回草案。
- route response 必须保留 `failure_code`、request / audit metadata、validation summary 和 blocked capability summary。
- route 不得被写成 public production API，也不得接真实数据库、Radish OIDC、repository adapter、schema migration 或 store selector。

## Web Consumer 边界

consumer 至少需要表达四类状态：

| 状态 | 含义 |
| --- | --- |
| `sample` | 离线样例，只能审查，不能当作保存结果 |
| `unsaved_local` | 用户当前本地草案，尚未写入 dev store |
| `saved_dev_record` | 已通过 dev-only consumer 保存并可读取的草案 |
| `version_conflict` | 保存遇到当前版本冲突，展示 current version metadata，保留本地草案 |
| `save_failed` | 保存失败，展示 failure code 和可恢复建议 |

consumer 不得把 `saved_dev_record` 展示为 publish ready、run ready 或 production ready。`valid_for_review` 只能用于审查入口。

## 必测场景

- 成功保存后返回 `saved_dev_record`，并展示递增后的 `draft_version`。
- 基于旧 `draft_version` 保存时返回 `draft_version_conflict`，不得覆盖当前版本；UI 映射为 `version_conflict`，并保留本地草案。
- `draft_write_disabled` 时 UI 仍可展示 sample / unsaved 草案，但不得声明已保存。
- `draft_scope_denied`、`draft_not_found`、`draft_store_unavailable` 均 fail closed，不回退 sample。
- blocked capability 草案可以作为审查 finding 展示，但不能解锁 executor、confirmation、writeback 或 replay。

## 验收方式

- Go route tests 覆盖 dev auth、write enablement、scope、version conflict、no sample fallback 和 store failure。
- TypeScript consumer 或 web smoke 覆盖 sample / unsaved / saved / failed 状态区分。
- route contract 和 consumer smoke checker 覆盖 envelope keys、dev headers、env flags、failure code、`version_conflict`、no sample fallback 和 App 状态展示。
- `npm run build` 通过。
- `go test ./services/platform/...` 或等价 Go 单元测试通过。
- `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不实现 workflow executor、publish、run、confirmation decision、decision store、writeback、replay、resume 或 materialized result reader。
- 不创建 public production API，不接真实数据库、repository adapter、schema migration、store selector、Radish OIDC、token validation、API key lifecycle、quota enforcement、billing 或 cost ledger。
- 不把 dev-only route、memory dev store 或 saved draft 解释为 durable persistence ready。
