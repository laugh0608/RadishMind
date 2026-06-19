# Workflow Saved Draft Store Selector Smoke Readiness v1 任务卡

更新时间：2026-06-15

## 任务标识

- 切片：`workflow-saved-draft-store-selector-smoke-readiness-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`draft_store_selector_smoke_readiness_defined`

## 目标

在 store selector enablement preconditions 和 schema artifact evidence 之后，固定 future saved workflow draft store selector smoke 的 mode matrix、operation matrix、schema artifact failure matrix、failure mapping、no fallback、no side effects 和 implementation artifact guard。

本任务卡只定义 selector smoke readiness，不创建正式 config entry、selector 函数、selector 类型、selector smoke fixture、selector smoke checker、Go selector test、repository interface、repository adapter、SQL migration、真实数据库、Radish OIDC middleware、production API、saved draft list、publish、run、executor、confirmation、writeback 或 replay。

## 输入事实源

- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)
- [Saved Workflow Draft Repository Contract Preconditions v1 专题](../features/workflow/saved-workflow-draft-repository-contract-preconditions-v1.md)
- [Saved Workflow Draft Store Selector Enablement Preconditions v1 专题](../features/workflow/saved-workflow-draft-store-selector-enablement-preconditions-v1.md)
- [Saved Workflow Draft Schema Artifact Evidence v1 专题](../features/workflow/saved-workflow-draft-schema-artifact-evidence-v1.md)
- [Saved Workflow Draft Store Selector Smoke Readiness v1 专题](../features/workflow/saved-workflow-draft-store-selector-smoke-readiness-v1.md)

## 本轮交付

- 新增 selector smoke readiness 细专题，固定 readiness-only 状态。
- 新增 `workflow-saved-draft-store-selector-smoke-readiness-v1` fixture / checker。
- checker 接入 fast baseline，校验 dependency、selector smoke boundary、mode matrix、operation matrix、schema artifact failure matrix、failure mapping、no fallback、no side effects、文档引用和 forbidden implementation artifact。
- 同步更新 workflow 入口、Saved Draft 主专题、store selector preconditions、schema artifact evidence、User Workspace、当前焦点、任务卡索引、脚本说明和周志。

## Smoke Readiness Contract

future selector smoke 必须覆盖：

- unset / `memory_dev`：选择 platform memory dev store，仅用于 dev-only saved draft route。
- `repository_disabled`：返回 `repository_store_disabled`，不 fallback。
- `repository`：adapter / schema / auth / smoke gate 未满足前返回 `repository_store_disabled`，不 fallback。
- unknown mode：返回 `invalid_draft_store_mode`，不 fallback。
- `SaveWorkflowDraftRecord`、`ReadWorkflowDraftRecord` 和 `ListWorkflowDraftRecords` 三类 operation。
- schema artifact failure：`draft_schema_migration_not_applied`、`draft_store_schema_version_mismatch` 和 `draft_store_migration_unavailable`。

## 验收口径

- `workflow-saved-draft-store-selector-smoke-readiness-v1` checker 通过。
- `./scripts/check-repo.sh --fast` 通过。
- 文档和 fixture 均保持 readiness-only，不声明 selector smoke ready、store selector ready、repository adapter ready 或 production ready。

## 停止线

- 不创建正式 config entry、selector 函数、selector 类型、selector smoke fixture、selector smoke checker、Go selector test、repository interface、repository adapter、SQL migration、真实数据库、Radish OIDC middleware、token validation、public production API、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把本任务卡、fixture 或 checker 解释为 durable persistence、saved draft list、store selector implementation、repository adapter、saved draft list、publish、run 或 production readiness。
