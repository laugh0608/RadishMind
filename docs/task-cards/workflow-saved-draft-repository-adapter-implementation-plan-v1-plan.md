# Workflow Saved Draft Repository Adapter Implementation Plan v1 任务卡

更新时间：2026-06-17

## 任务定位

任务 ID：`workflow-saved-draft-repository-adapter-implementation-plan-v1`

状态：`draft_repository_adapter_implementation_plan_defined`

本任务承接 `Saved Workflow Draft Repository Contract Smoke Runner Implementation v1`，用于固定 future saved workflow draft repository adapter 的实现计划。它只定义 adapter 进入实现前必须消费的 contract smoke runner、auth context、schema artifact evidence、selector smoke readiness、failure mapping、no fallback 和 no side effects 证据，不创建 repository interface 或 adapter。

## 输入

- `docs/features/workflow/saved-workflow-draft-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-contract-smoke-runner-implementation-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-contract-smoke-v1.md`
- `docs/features/workflow/saved-workflow-draft-auth-context-preconditions-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-artifact-evidence-v1.md`
- `docs/features/workflow/saved-workflow-draft-store-selector-smoke-readiness-v1.md`
- `scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.json`

## 输出

- 新增 `docs/features/workflow/saved-workflow-draft-repository-adapter-implementation-plan-v1.md`。
- 新增 `scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-plan-v1.json`。
- 新增 `scripts/checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-plan-v1.py`。
- 将 checker 接入 `scripts/check-repo.py` fast baseline。
- 同步 `docs/features/README.md`、`docs/features/workflow/README.md`、`docs/features/workflow/saved-workflow-draft-v1.md`、`docs/radishmind-current-focus.md`、`scripts/README.md`、`docs/task-cards/README.md` 和周志。

## 实现范围

本任务固定三类 future adapter operation：

- `SaveWorkflowDraftRecord`
- `ReadWorkflowDraftRecord`
- `ListWorkflowDraftRecords`

每个 operation 必须消费 static runner、repository contract smoke、auth context、schema artifact evidence 和 selector smoke readiness。计划必须明确 required future adapter checks、failure mapping、no memory dev fallback、no sample fallback 和 no side effects。

## 不在范围

- 不创建 repository interface、repository adapter、store selector、formal config entry、SQL migration、真实数据库、Radish OIDC middleware、token validation 或 production API consumer。
- 不实现 durable persistence、saved draft list 新实现、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 implementation plan 写成 adapter smoke ready、database ready、selector ready、OIDC ready 或 production ready。

## 验收方式

本任务完成后至少运行：

```bash
cd services/platform
go test ./internal/httpapi
cd ../..
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-plan-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.py
./scripts/check-repo.sh --fast
```

若后续批次创建 repository interface、adapter、selector、SQL、OIDC 或 production API，应新增独立专题、任务卡和验证记录，不复用本任务状态提升实现结论。

## 停止线

- `draft_repository_adapter_implementation_plan_defined` 只表示 future adapter implementation plan 已固定。
- 不声明 `repository_interface_implemented`、`repository_adapter_ready`、`durable_store_implemented`、`database_schema_ready`、`store_selector_implemented`、`radish_oidc_ready`、`production_api_consumer_ready` 或 `production_ready`。
