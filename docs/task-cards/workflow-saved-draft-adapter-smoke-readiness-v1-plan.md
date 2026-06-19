# Workflow Saved Draft Adapter Smoke Readiness v1 任务卡

## 任务标识

- 切片：`workflow-saved-draft-adapter-smoke-readiness-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`draft_adapter_smoke_readiness_defined`

## 目标

在创建 durable repository adapter、repository interface、adapter smoke fixture、selector implementation、schema artifact 文件、SQL migration、Radish OIDC middleware 或 production API 前，固定 future saved workflow draft adapter smoke 的可检查治理边界。本任务卡只定义 future adapter smoke 如何消费 static runner、repository adapter implementation plan、schema artifact manifest、selector smoke readiness 和 auth context preconditions。

本任务卡不表示 adapter smoke ready、repository adapter ready、durable store ready、database ready、schema artifact file ready、store selector ready、production auth ready、production API ready 或 production ready。

## 范围

- 新增 adapter smoke readiness fixture 和 checker。
- 消费 `workflow-saved-draft-repository-adapter-implementation-plan-v1`、`workflow-saved-draft-schema-artifact-manifest-v1`、`workflow-saved-draft-store-selector-smoke-readiness-v1`、`workflow-saved-draft-auth-context-preconditions-v1` 和 `workflow-saved-draft-repository-contract-smoke-runner-implementation-v1`。
- 固定 future adapter smoke 必须覆盖 `SaveWorkflowDraftRecord`、`ReadWorkflowDraftRecord` 和 `ListWorkflowDraftRecords`，并与 adapter plan、static runner、schema artifact manifest、selector smoke readiness 和 auth context preconditions 的 operation matrix 对齐。
- 固定 adapter smoke 前仍未满足的 gate：selector implementation、schema artifact materialization、production auth、repository adapter implementation、adapter smoke execution 和 production API consumer。
- 固定 failure mapping、no fallback、no side effects 和 no implementation leak 检查。
- 将 checker 接入仓库 fast baseline，并同步 workflow 专题入口、Saved Workflow Draft v1、current focus、脚本说明、任务卡入口和周志。

## 不在本次范围

- 不创建 `workflow-saved-draft-adapter-smoke-v1.json` 或 adapter smoke checker。
- 不创建 `workflow_saved_draft_repository.go`、`workflow_saved_draft_repository_adapter.go`、adapter test 或 adapter contract smoke test。
- 不实现 repository interface、repository adapter、store selector、database query、database connection、SQL、migration 或 migration runner。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle、quota enforcement 或 billing。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何执行能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-manifest-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-plan-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-store-selector-smoke-readiness-v1.py`
- `./scripts/check-repo.sh --fast`
