# Workflow Saved Draft Store Selector Implementation v1 任务卡

状态：`draft_store_selector_smoke_implemented`  
更新日期：2026-06-17

## 任务目标

在不创建 repository adapter、数据库 schema、SQL migration、OIDC/token validation 或 production API 的前提下，实现 saved workflow draft store selector 的受控运行时入口。

本批输出 formal config、`SelectWorkflowSavedDraftStore`、selector tests、HTTP fail-closed tests 和 selector smoke fixture/checker，让后续 durable store 工作能基于明确 store mode 继续推进。

## 输入事实源

- `docs/features/workflow/saved-workflow-draft-v1.md`
- `docs/features/workflow/saved-workflow-draft-store-selector-enablement-preconditions-v1.md`
- `docs/features/workflow/saved-workflow-draft-store-selector-smoke-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-store-selector-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-artifact-materialization-review-v1.md`

## 实现范围

- 新增配置文件键 `workflow_saved_draft_store`。
- 新增环境变量 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE`。
- 在 sanitized summary 暴露 `workflow_saved_draft_store_mode`。
- 实现 `SelectWorkflowSavedDraftStore` 和 `WorkflowSavedDraftStoreSelector`。
- 默认 `memory_dev` 继续使用现有 memory dev store。
- `repository_disabled` 和 `repository` 返回 `repository_store_disabled`。
- unknown mode 返回 `invalid_draft_store_mode`。
- save / read / list 必须保留 selector failure code，不重写成 `draft_store_unavailable`。
- 新增 `workflow-saved-draft-store-selector-smoke-v1` fixture / checker。

## 验收证据

- `go test ./internal/config ./internal/httpapi`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-store-selector-smoke-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-store-selector-smoke-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-store-selector-implementation-entry-review-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-review-v1.py`
- `./scripts/check-repo.sh --fast`
- 若 current focus / capability matrix 等真相源变更，补跑 `./scripts/check-repo.sh`。

## 停止线

- 不创建 repository interface、repository adapter、adapter smoke fixture 或 adapter smoke checker。
- 不创建 schema artifact 文件、migration root、DDL review、rollback evidence、migration smoke、SQL migration、schema version table 或 migration runner。
- 不连接数据库，不调用 OIDC，不校验 token，不实现 production API consumer。
- 不实现 publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。

## 后续建议

selector implementation 完成后，下一批 durable store 工作应优先选择 `Saved Workflow Draft Schema Artifact Materialization` 或 repository adapter implementation 中的一个方向，并保持 schema/auth/adapter/smoke 证据链可复验。
