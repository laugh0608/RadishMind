# Workflow Saved Draft Store Selector Implementation Entry Review v1 任务卡

## 任务标识

- 切片：`workflow-saved-draft-store-selector-implementation-entry-review-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`draft_store_selector_implementation_entry_review_defined`

## 目标

在创建正式 saved workflow draft store selector 前，评审是否打开 formal config entry、`SelectWorkflowSavedDraftStore`、selector unit tests 和 selector smoke fixture。本任务卡只固定 entry review 结论、候选项、阻塞条件、failure mapping、no fallback、no side effects 和 forbidden artifact guard。

当前结论：本批不打开 selector implementation entry，不创建后续 implementation task card，不写 runtime selector 代码。

## 范围

- 新增 store selector implementation entry review fixture 和 checker。
- 消费 `workflow-saved-draft-store-selector-enablement-preconditions-v1`、`workflow-saved-draft-store-selector-smoke-readiness-v1`、`workflow-saved-draft-repository-adapter-implementation-plan-v1`、`workflow-saved-draft-schema-artifact-manifest-v1` 和 `workflow-saved-draft-adapter-smoke-readiness-v1`。
- 固定四个候选项：formal config entry、`SelectWorkflowSavedDraftStore`、selector unit tests、selector smoke fixture。
- 固定本批结论：四个候选项均 `blocked`，后续只能在单独 implementation task card 中打开。
- 将 checker 接入仓库 fast baseline，并同步 workflow 专题入口、Saved Workflow Draft v1、current focus、脚本说明、任务卡入口和周志。

## 不在本次范围

- 不创建 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE` 正式 config entry。
- 不创建 `workflow_saved_draft_store_selector.go`、`workflow_saved_draft_store_selector_test.go`、`workflow-saved-draft-store-selector-smoke-v1.json` 或 selector smoke checker。
- 不实现 repository interface、repository adapter、database query、database connection、SQL、migration、migration runner 或 schema artifact 文件。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle、quota enforcement 或 billing。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何执行能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-store-selector-implementation-entry-review-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-store-selector-smoke-readiness-v1.py`
- `./scripts/check-repo.sh --fast`
