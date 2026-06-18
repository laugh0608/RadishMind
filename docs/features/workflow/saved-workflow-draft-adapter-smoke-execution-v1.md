# Saved Workflow Draft Adapter Smoke Execution v1 专题

更新时间：2026-06-18

## 专题定位

`Saved Workflow Draft Adapter Smoke Execution v1` 承接 [Saved Workflow Draft Adapter Smoke Readiness v1](saved-workflow-draft-adapter-smoke-readiness-v1.md) 和 [Saved Workflow Draft Repository Adapter Implementation v1 任务卡](../../task-cards/workflow-saved-draft-repository-adapter-implementation-v1-plan.md)，用于验证已实现的 `SavedWorkflowDraftRepositoryAdapter` 能消费 static repository contract smoke cases，并按 save / read / list operation matrix 返回既定成功投影和失败语义。

本专题只执行 adapter 层 smoke。它使用注入式 fake query executor，不通过 dev HTTP route，不启用 `repository` store mode，不连接数据库，不运行 SQL，不调用 Radish OIDC，不创建 production API，不发布或执行 workflow。

状态：`draft_adapter_smoke_executed`

## 当前输入事实

- `Saved Workflow Draft Repository Contract Smoke Runner Implementation v1` 已提供 static `SavedWorkflowDraftRepositoryContractSmokeRunner` 和 save / read / list cases。
- `Saved Workflow Draft Repository Adapter Implementation v1` 已实现 `SavedWorkflowDraftRepository` interface、注入式 query executor adapter、schema preflight、auth context failure mapping 和 adapter unit tests。
- `Saved Workflow Draft Store Selector Implementation v1` 已实现 fail-closed selector，但 `repository` mode 仍返回 `repository_store_disabled`。
- `Saved Workflow Draft Schema Artifact Materialization v1` 已物化静态 schema artifact；它不是 SQL migration 或 migration runner。
- `Saved Workflow Draft Production Auth Readiness v1` 只固定 production auth readiness evidence，尚未实现 OIDC middleware、token validation 或 membership adapter。

## 实现内容

本批新增：

- `services/platform/internal/httpapi/workflow_saved_draft_repository_adapter_smoke_test.go`
- `scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-v1.json`
- `scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-v1.py`
- `scripts/checks/control_plane/workflow_saved_draft_adapter_smoke_execution_guard.py`

Go smoke test 做三类验证：

- 先运行 `RunSavedWorkflowDraftRepositoryContractSmoke`，确认 static contract smoke cases 自身通过。
- 再通过 `NewSavedWorkflowDraftRepositoryAdapter` 和 injected fake query executor 执行 `SaveWorkflowDraftRecord`、`ReadWorkflowDraftRecord` 和 `ListWorkflowDraftRecords`。
- 覆盖 version conflict、not found、store unavailable、store contract mismatch、auth context mismatch、schema migration not applied、store schema version mismatch、store migration unavailable 和 missing scope grant 的 fail-closed 语义。

## 验证边界

| 项 | 结论 |
| --- | --- |
| adapter smoke fixture / checker | 已创建并接入仓库检查 |
| adapter smoke Go test | 已创建，直接调用 adapter，不通过 HTTP route |
| static contract smoke cases | 已消费 |
| fake query executor | 仅用于本地测试状态，不代表 durable store |
| repository store mode | 未启用 |
| 数据库 / SQL / migration runner | 未创建 |
| Radish OIDC / token validation / membership adapter | 未创建 |
| production API / executor / confirmation / writeback / replay | 未创建 |

## Failure Mapping

adapter smoke 执行验证以下关键失败语义：

- `draft_not_found`
- `draft_version_conflict`
- `draft_store_unavailable`
- `draft_store_contract_mismatch`
- `draft_auth_context_contract_mismatch`
- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`
- `draft_scope_grant_missing`

完整 failure code 集仍来自 repository contract 和 adapter implementation fixture，包括 `draft_scope_denied`、`draft_schema_version_unsupported`、`draft_payload_invalid`、`repository_store_disabled`、`invalid_draft_store_mode`、`draft_identity_context_missing`、`draft_tenant_binding_missing` 和 `draft_workspace_membership_denied`。

## 验收方式

- `go test ./internal/httpapi`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-readiness-v1.py`
- `./scripts/check-repo.sh --fast`

若同步当前焦点、路线图或其它真相源，补跑全量 `./scripts/check-repo.sh`。

## 后续准入

本专题完成后，durable store 方向仍不能直接进入生产能力声明。下一步只能在以下方向中选一个独立推进：

1. `Saved Workflow Draft Repository Mode Enablement v1`：把 selector 的 `repository` mode 从 fail-closed 转为受控启用前，必须先定义 repository mode runtime boundary、配置门禁、schema preflight、adapter smoke 依赖和回滚策略。
2. `Saved Workflow Draft Production Auth Runtime v1`：实现 Radish OIDC middleware、token validation 和 workspace membership adapter 前，必须独立定义 runtime evidence、failure envelope 和 sanitized audit。

## 停止线

- 不启用 `repository` store mode。
- 不连接真实数据库，不运行 SQL，不创建 SQL migration、schema version table 或 migration runner。
- 不创建 OIDC middleware、token validation、workspace membership adapter 或 production auth runtime。
- 不创建 public production API consumer、production CORS policy、API key lifecycle、quota enforcement、billing 或 cost ledger。
- 不实现 publish、run、workflow executor、tool executor、agent loop、confirmation decision、business writeback、replay、resume 或 materialized result reader。
- 不把 `draft_adapter_smoke_executed` 解释为 durable persistence ready、database ready、repository mode ready、OIDC ready、production API ready 或 production ready。
