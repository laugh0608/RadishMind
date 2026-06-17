# Saved Workflow Draft Schema Artifact Materialization Review v1 专题

更新时间：2026-06-17

## 专题定位

`Saved Workflow Draft Schema Artifact Materialization Review v1` 承接 [Saved Workflow Draft Schema Artifact Evidence v1](saved-workflow-draft-schema-artifact-evidence-v1.md)、[Saved Workflow Draft Schema Artifact Manifest v1](saved-workflow-draft-schema-artifact-manifest-v1.md)、[Saved Workflow Draft Adapter Smoke Readiness v1](saved-workflow-draft-adapter-smoke-readiness-v1.md)、[Saved Workflow Draft Repository Adapter Implementation Plan v1](saved-workflow-draft-repository-adapter-implementation-plan-v1.md) 和 [Saved Workflow Draft Store Selector Implementation Entry Review v1](saved-workflow-draft-store-selector-implementation-entry-review-v1.md)，用于评审是否打开 saved draft schema artifact 的物化入口。

本专题只完成 materialization review。当前结论是不打开 migration root、manifest 文件、DDL review artifact、rollback evidence artifact 或 migration smoke artifact；后续如需物化 schema artifact，必须另开独立 implementation task card。

Review conclusion: schema artifact materialization entry not opened.

状态：`draft_schema_artifact_materialization_review_defined`

## 当前输入事实

- `Saved Workflow Draft Schema Artifact Evidence v1` 已固定 future schema artifact manifest、DDL review、rollback evidence、migration smoke、failure mapping 和 artifact guard。
- `Saved Workflow Draft Schema Artifact Manifest v1` 已固定 future manifest shape、section matrix、operation predicate coverage、failure mapping、no fallback 和 no side effects，但未创建任何 schema artifact 文件。
- `Saved Workflow Draft Adapter Smoke Readiness v1` 已把 schema artifact materialization gate 保持为 `not_satisfied`，要求 migration root、manifest file、DDL review evidence、rollback evidence 和 migration smoke evidence 后续单独准入。
- `Saved Workflow Draft Repository Adapter Implementation Plan v1` 已固定 future adapter 对 schema artifact 的依赖，但未创建 repository interface、repository adapter 或数据库 query。
- `Saved Workflow Draft Store Selector Implementation Entry Review v1` 已确认 selector implementation entry 当前不打开，schema artifact 物化不能借 selector 评审绕过。

后续 `Saved Workflow Draft Store Selector Implementation v1` 已在独立任务卡中完成，状态为 `draft_store_selector_smoke_implemented`。本专题仍保留 materialization review 当批未创建 schema artifact materialization artifact 的历史事实。

## Materialization Review Decision

| candidate | 本次结论 | 后续条件 |
| --- | --- | --- |
| migration root | `blocked` | 需先创建独立 schema artifact materialization task card，并明确目录职责、手动 apply gate 和 no startup auto-migration |
| manifest file | `blocked` | 需先满足 manifest shape contract、DDL review、rollback evidence 和 migration smoke artifact 的同批一致性 |
| DDL review artifact | `blocked` | 需先明确人工 review 记录、破坏性变更策略、backup requirement 和 migration lock 口径 |
| rollback evidence artifact | `blocked` | 需先明确 rollback command、forward-only exception 或失败恢复路径 |
| migration smoke artifact | `blocked` | 需先有物化 manifest、DDL review 和 rollback evidence，并覆盖 schema version、scope predicate、owner list、version conflict 和 no side effects |

本批不创建 `workflow-saved-draft-schema-artifact-materialization-v1` task card，也不创建 schema artifact runtime artifact。若下一批要进入物化，应先重跑本 review 对应 checker，确认只有 schema artifact materialization 这一条 track 被打开。

## Gate Matrix

| gate | 当前状态 | 说明 |
| --- | --- | --- |
| schema artifact evidence consumed | `satisfied` | 已消费 schema artifact evidence、DDL review、rollback 和 migration smoke 前置证据 |
| schema artifact manifest consumed | `satisfied` | 已消费 manifest shape、section matrix 和 operation predicate coverage |
| repository adapter plan consumed | `satisfied` | 已确认 future adapter 仍依赖 schema artifact gate |
| selector entry review consumed | `satisfied` | 已确认 selector implementation entry 未打开 |
| adapter smoke readiness consumed | `satisfied` | 已确认 schema artifact materialization gate 仍未满足 |
| materialization review defined | `satisfied` | 本专题定义 materialization review 结论和后续准入 |
| migration root candidate | `blocked` | 不创建 `services/platform/migrations/workflow_saved_drafts` |
| manifest file candidate | `blocked` | 不创建 `manifest.json` |
| DDL review candidate | `blocked` | 不创建 `ddl-review.md` |
| rollback evidence candidate | `blocked` | 不创建 `rollback-evidence.json` |
| migration smoke candidate | `blocked` | 不创建 `migration-smoke.json` |
| store selector gate | `satisfied` | 已由 `Saved Workflow Draft Store Selector Implementation v1` 创建 formal config、selector、selector tests 和 selector smoke fixture |
| repository adapter gate | `not_satisfied` | repository interface、adapter 和 database query 仍未创建 |
| production auth gate | `not_satisfied` | Radish OIDC、token validation、membership adapter 和 scope projection 仍未实现 |
| database connection gate | `not_satisfied` | 不连接真实数据库，不运行 migration，不创建 schema version table |
| production API gate | `not_satisfied` | public production API consumer、production auth policy 和 CORS policy 仍未打开 |
| no schema artifact materialization artifacts leaked | `required_now` | 当前必须确认 migration root、schema artifact、SQL、repository、selector、OIDC artifact 未提前出现 |

## Failure Mapping

materialization review 必须继续保留以下 fail-closed failure code：

- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`
- `draft_store_contract_mismatch`
- `repository_store_disabled`
- `invalid_draft_store_mode`

其中 `draft_version_conflict`、`draft_scope_denied`、`draft_not_found` 和 `draft_store_unavailable` 不得被 migration failure 覆盖；缺失 manifest、DDL review、rollback evidence 或 migration smoke 不得 fallback 到 memory dev store、sample、fixture 或 dev HTTP route。

## No Side Effects

本批只能读取 committed fixture、文档和源码树。必须保持：

- 不创建 migration root、manifest 文件、DDL review、rollback evidence、migration smoke、SQL migration、schema version table 或 migration runner。
- 不创建 repository interface、repository adapter、store selector、selector smoke fixture 或 adapter smoke fixture。
- 不连接数据库、不运行 migration、不调用 Radish OIDC 或 token validation。
- 不调用 workflow executor、tool executor、confirmation decision、business writeback 或 replay。

## 后续准入

后续可选方向：

1. `Saved Workflow Draft Schema Artifact Materialization v1`：只有在本 review 重新确认单一 materialization track 后，才可创建 migration root、manifest、DDL review、rollback evidence 和 migration smoke artifact。
2. `Saved Workflow Draft Store Selector Implementation v1`：独立打开 formal config、selector、selector tests 和 selector smoke fixture，不能和 schema artifact materialization 混在同一批。
3. `Saved Workflow Draft Adapter Smoke Execution v1`：只能在 selector implementation、schema artifact materialization、production auth 和 repository adapter implementation 均满足后进入。

## 验收方式

- 新增 `workflow-saved-draft-schema-artifact-materialization-review-v1` fixture / checker，固定 materialization decision、candidate matrix、gate matrix、failure mapping、no fallback、no side effects 和 forbidden artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 `workflow-saved-draft-store-selector-implementation-entry-review-v1` 之后、`product-surface-readiness-implementation-trigger-recheck-v1` 之前。
- 本批至少运行专项 checker、schema artifact manifest checker、schema artifact evidence checker、adapter smoke readiness checker、selector implementation entry review checker 和 `./scripts/check-repo.sh --fast`；因同步 current focus 和能力矩阵，补跑全量 `./scripts/check-repo.sh`。

## 停止线

- 不创建 migration root、manifest 文件、DDL review artifact、rollback evidence artifact、migration smoke artifact、SQL migration、schema version table、migration runner、数据库连接、repository interface、repository adapter、store selector、selector smoke fixture、adapter smoke fixture、Radish OIDC middleware、token validation 或 production API consumer。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 `draft_schema_artifact_materialization_review_defined` 解释为 schema artifact file ready、DDL ready、migration ready、database ready、repository adapter ready、repository mode ready、adapter smoke ready、OIDC ready 或 production ready。
