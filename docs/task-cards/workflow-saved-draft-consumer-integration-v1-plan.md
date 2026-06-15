# Workflow Saved Draft Consumer Integration v1 任务卡

更新时间：2026-06-15

## 任务标识

- 切片：`workflow-saved-draft-consumer-integration-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`dev_only_consumer_integration_implemented`

## 目标

把已实现的 `SavedWorkflowDraft` platform Go domain service 接入开发态 HTTP route 和 web consumer，让 `apps/radishmind-web/` 能在显式 dev 配置下保存、读取和校验 workflow 草案，并明确区分 sample / unsaved / saved / failed 状态。

本任务卡只承接 dev-only consumer integration，不声明 public production API、durable persistence、workflow publish、run、executor、confirmation decision、writeback 或 replay。

## 本轮实现

- 新增 `services/platform/internal/httpapi/workflow_saved_draft_http.go`，注册 dev-only save / read / validate route，并把 HTTP DTO 映射到 `SavedWorkflowDraft` domain service。
- 新增 `services/platform/internal/httpapi/workflow_saved_draft_http_test.go`，覆盖 dev route 默认关闭、dev auth、write enablement、save / read / validate、version conflict、scope mismatch 和 no sample fallback。
- 更新 platform config，新增 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP` 和 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE`，默认均关闭。
- 新增 `apps/radishmind-web/src/features/control-plane-read/savedWorkflowDraftConsumer.ts`，默认 sample-only；显式 `VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE=dev-saved-draft-http` 时调用 dev route。
- 更新 Draft Designer 页面，展示 saved draft consumer 状态、failure code、audit ref、version，并提供 dev-only validate / save / read 操作入口。

## 输入事实源

- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)
- [Dev-only Saved Draft Consumer 实现专题](../features/workflow/dev-only-saved-draft-consumer.md)
- [Workflow Draft Designer Surface 专题](../features/workflow/draft-designer-surface.md)
- [Workflow Saved Draft v1 Implementation 任务卡](workflow-saved-draft-v1-implementation-plan.md)

## 验收口径

- 后端 route 默认关闭；未设置 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP=1` 时返回配置边界错误。
- save 必须同时满足 dev auth、workspace / application header、`workflow_drafts:write` scope 和 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE=1`。
- read 必须按 `workspace_id` + `application_id` + `draft_id` scope 查询；scope mismatch、not found 和 store failure fail closed。
- validate 不写入 store，也不 materialize saved draft。
- UI 必须区分 sample、unsaved local、saved dev record 和 failure 状态。
- 通过 Go 单元测试、web build 和 `./scripts/check-repo.sh --fast`。

## 停止线

- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、business writeback、run replay、run resume 或 materialized result reader。
- 不创建 public production API，不接真实数据库、repository adapter、schema migration、store selector、Radish OIDC、token validation、API key lifecycle、quota enforcement、billing 或 cost ledger。
- 不把 dev-only route、memory dev store 或 saved draft record 解释为 durable persistence ready。
