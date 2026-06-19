# Workflow Saved Draft Adapter Smoke v1 任务卡

更新时间：2026-06-18

## 任务标识

- 任务 ID：`workflow-saved-draft-adapter-smoke-v1`
- 状态：`draft_adapter_smoke_executed`

## 任务定位

本任务卡承接 `Saved Workflow Draft Adapter Smoke Execution v1`，用于固定并记录 repository adapter smoke execution 的实现边界、验证链路和停止线。

当前只打开 adapter 层 smoke：通过 static repository contract smoke cases、`SavedWorkflowDraftRepositoryAdapter` 和 injected fake query executor 验证 save / read / list 以及关键失败语义。本任务不启用 `repository` store mode，不连接数据库，不运行 SQL，不创建 migration runner，不接 OIDC，不创建 production API，不进入 workflow publish / run / executor。

## 输入

- `docs/features/workflow/saved-workflow-draft-v1.md`
- `docs/features/workflow/saved-workflow-draft-adapter-smoke-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-adapter-smoke-execution-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-contract-smoke-runner-implementation-v1.md`
- `docs/task-cards/workflow-saved-draft-repository-adapter-implementation-v1-plan.md`
- `services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner.go`
- `services/platform/internal/httpapi/workflow_saved_draft_repository.go`
- `services/platform/internal/httpapi/workflow_saved_draft_repository_adapter.go`
- `services/platform/migrations/workflow_saved_drafts/manifest.json`

## 已实现范围

- 新增 `workflow_saved_draft_repository_adapter_smoke_test.go`，直接执行 adapter smoke，不通过 HTTP route。
- 新增 `workflow-saved-draft-adapter-smoke-v1.json` fixture，固定本批 status、依赖、operation smoke matrix、failure mapping、no fallback 和 no side effects。
- 新增 `check-workflow-saved-draft-adapter-smoke-v1.py`，校验 fixture、Go test、依赖状态、禁止项和 `check-repo.py` 注册顺序。
- 新增 `workflow_saved_draft_adapter_smoke_execution_guard.py`，让旧 readiness / entry review checker 只放行本批声明的 smoke artifact。
- 更新 `check-repo.py`，让 adapter smoke checker 位于 repository adapter implementation checker 之后。

## Operation Matrix

| operation | scope | smoke 要求 |
| --- | --- | --- |
| `SaveWorkflowDraftRecord` | `workflow_drafts:write` | 消费 static contract case，通过 adapter 保存 sanitized draft，并返回 current version metadata |
| `ReadWorkflowDraftRecord` | `workflow_drafts:read` | 消费 static contract case，通过 adapter 读取 saved draft，不回退 sample |
| `ListWorkflowDraftRecords` | `workflow_drafts:read` | 消费 static contract case，通过 adapter 返回 sanitized summary list，不回退 fixture |

## Failure Mapping

本批 Go smoke test 至少执行：

- `draft_not_found`
- `draft_version_conflict`
- `draft_store_unavailable`
- `draft_store_contract_mismatch`
- `draft_auth_context_contract_mismatch`
- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`
- `draft_scope_grant_missing`

完整 contract failure 集继续由 repository contract smoke runner、adapter implementation fixture 和本批 fixture 联合固定。

## 验收方式

代码实现完成时至少运行：

```bash
cd services/platform
go test ./internal/httpapi
cd ../..
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-readiness-v1.py
./scripts/check-repo.sh --fast
```

若同步 current focus、roadmap 或其它阶段真相源，补跑全量：

```bash
./scripts/check-repo.sh
```

## 停止线

- 不启用 `repository` store mode；selector 的 `repository_disabled` / `repository` / unknown mode 仍必须 fail closed。
- 不创建 SQL migration、schema version table、migration runner、真实数据库连接或外部数据库 fixture。
- 不创建 Radish OIDC middleware、token validation、session cookie、workspace membership adapter 或 production auth runtime。
- 不创建 public production API consumer、production CORS policy、API key lifecycle、quota enforcement、billing 或 cost ledger。
- 不实现 durable publish、run、workflow executor、node executor、tool executor、agent loop、confirmation decision、business writeback、replay、resume 或 materialized result reader。
- 不把 `draft_adapter_smoke_executed` 解释为 durable persistence ready、database ready、repository mode ready、OIDC ready、production API ready 或 production ready。
