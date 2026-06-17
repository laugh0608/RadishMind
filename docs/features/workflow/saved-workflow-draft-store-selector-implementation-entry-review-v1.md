# Saved Workflow Draft Store Selector Implementation Entry Review v1 专题

更新时间：2026-06-17

## 专题定位

`Saved Workflow Draft Store Selector Implementation Entry Review v1` 承接 [Saved Workflow Draft Store Selector Enablement Preconditions v1](saved-workflow-draft-store-selector-enablement-preconditions-v1.md)、[Saved Workflow Draft Store Selector Smoke Readiness v1](saved-workflow-draft-store-selector-smoke-readiness-v1.md)、[Saved Workflow Draft Repository Adapter Implementation Plan v1](saved-workflow-draft-repository-adapter-implementation-plan-v1.md)、[Saved Workflow Draft Schema Artifact Manifest v1](saved-workflow-draft-schema-artifact-manifest-v1.md) 和 [Saved Workflow Draft Adapter Smoke Readiness v1](saved-workflow-draft-adapter-smoke-readiness-v1.md)，用于评审是否打开 saved draft store selector 的正式实现入口。

本专题只完成 entry review。当前结论是不打开 formal config entry、`SelectWorkflowSavedDraftStore`、selector unit tests 或 selector smoke fixture；后续如需实现 selector，必须另开独立 implementation task card。

后续 `Saved Workflow Draft Store Selector Implementation v1` 已在独立任务卡中完成，状态为 `draft_store_selector_smoke_implemented`。本专题仍保留 entry review 当批未创建 selector artifact 的历史事实。

状态：`draft_store_selector_implementation_entry_review_defined`

## 当前输入事实

- `Saved Workflow Draft Store Selector Enablement Preconditions v1` 已固定 future store mode、selector gate、failure mapping、no fallback 和 dev flag boundary。
- `Saved Workflow Draft Store Selector Smoke Readiness v1` 已固定 future selector smoke mode matrix、operation matrix、schema artifact failure、no fallback 和 no side effects，但未创建 selector smoke fixture 或 checker。
- `Saved Workflow Draft Repository Adapter Implementation Plan v1` 已固定 future selector 文件落点和 adapter 依赖关系，但未创建 repository interface、repository adapter 或 selector。
- `Saved Workflow Draft Schema Artifact Manifest v1` 已固定 future manifest contract，但未创建 schema artifact 文件、migration root、DDL review 或 SQL migration。
- `Saved Workflow Draft Adapter Smoke Readiness v1` 已把 selector implementation gate 保持为 `not_satisfied`，要求 formal config、`SelectWorkflowSavedDraftStore`、selector tests 和 selector smoke fixture 后续单独准入。

## Entry Review Decision

| candidate | 本次结论 | 后续条件 |
| --- | --- | --- |
| formal store config entry | `blocked` | 需先明确 operator docs、allowed values、unset behavior、diagnostics redaction 和 dev flag 独立性 |
| `SelectWorkflowSavedDraftStore` | `blocked` | 需先有正式 config entry 任务卡，并保持 `repository` / `repository_disabled` fail closed |
| selector unit tests | `blocked` | 需等 selector function 进入实现任务后覆盖 default / reserved / unknown mode |
| selector smoke fixture | `blocked` | 需等 selector implementation 与 unit tests 完成，再用独立 smoke fixture 覆盖 operation matrix |

本批不创建 `workflow-saved-draft-store-selector-implementation-v1` task card，也不创建 selector runtime artifact。若下一批要进入实现，应先重跑本 review 对应 checker，确认只有 selector implementation 这一条 track 被打开。

## Gate Matrix

| gate | 当前状态 | 说明 |
| --- | --- | --- |
| selector enablement preconditions consumed | `satisfied` | 已消费 store mode、selector gate 和 dev flag boundary |
| selector smoke readiness consumed | `satisfied` | 已消费 mode / operation matrix 和 failure mapping |
| adapter smoke readiness consumed | `satisfied` | 已确认 selector implementation gate 仍未满足 |
| selector entry review defined | `satisfied` | 本专题定义 entry review 结论和后续准入 |
| formal config candidate | `blocked` | 不创建 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE` 正式 config entry |
| selector function candidate | `blocked` | 不创建 `SelectWorkflowSavedDraftStore` 或 selector type |
| selector tests candidate | `blocked` | 不创建 Go selector unit tests |
| selector smoke fixture candidate | `blocked` | 不创建 selector smoke fixture 或 checker |
| repository adapter gate | `not_satisfied` | repository interface、adapter 和 database query 仍未创建 |
| schema artifact file gate | `not_satisfied` | schema artifact manifest file、DDL review、rollback 和 SQL migration 仍未创建 |
| production auth gate | `not_satisfied` | Radish OIDC、token validation、membership adapter 和 scope projection 仍未实现 |
| production API gate | `not_satisfied` | public production API consumer、production auth policy 和 CORS policy 仍未打开 |
| no selector implementation artifacts leaked | `required_now` | 当前必须确认 config、selector、tests、smoke fixture、repository、SQL 和 OIDC artifact 未提前出现 |

## Failure Mapping

entry review 必须继续保留以下 failure code：

- `repository_store_disabled`
- `invalid_draft_store_mode`
- `draft_store_unavailable`
- `draft_store_contract_mismatch`
- `draft_auth_context_contract_mismatch`
- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`

其中 `draft_version_conflict`、`draft_scope_denied`、`draft_not_found` 和 `draft_store_unavailable` 不得被 selector failure 覆盖；reserved / unknown mode 不得 fallback 到 memory dev store、sample、fixture 或 dev HTTP route。

## No Side Effects

本批只能读取 committed fixture、文档和源码树。必须保持：

- 不创建 formal config entry、selector、selector tests、selector smoke fixture 或 selector smoke checker。
- 不创建 repository interface、repository adapter、schema artifact 文件、SQL migration、migration runner 或数据库连接。
- 不接 Radish OIDC、token validation、production API consumer、API key、quota 或 billing。
- 不调用 workflow executor、tool executor、confirmation decision、business writeback 或 replay。

## 后续准入

本专题完成后，已继续补齐 [Saved Workflow Draft Schema Artifact Materialization Review v1](saved-workflow-draft-schema-artifact-materialization-review-v1.md)，状态为 `draft_schema_artifact_materialization_review_defined`。后续可选方向：

1. `Saved Workflow Draft Store Selector Implementation v1`：只有在本 review 重新确认单一 selector track 后，才可创建正式 config entry、selector function、unit tests 和 selector smoke fixture。
2. `Saved Workflow Draft Schema Artifact Materialization v1`：只有在 materialization review 重新确认单一 schema artifact track 后，才可创建 migration root、manifest、DDL review、rollback evidence 和 migration smoke artifact。
3. `Saved Workflow Draft Adapter Smoke Execution v1`：只能在 selector implementation、schema artifact materialization、production auth 和 repository adapter implementation 均满足后进入。

## 验收方式

- 新增 `workflow-saved-draft-store-selector-implementation-entry-review-v1` fixture / checker，固定 entry decision、candidate matrix、gate matrix、failure mapping、no fallback、no side effects 和 forbidden artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 `workflow-saved-draft-adapter-smoke-readiness-v1` 之后、`product-surface-readiness-implementation-trigger-recheck-v1` 之前。
- 本批至少运行专项 checker、adapter smoke readiness checker、selector smoke readiness checker 和 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建 formal config entry、`SelectWorkflowSavedDraftStore`、selector type、selector unit tests、selector smoke fixture、selector smoke checker、repository interface、repository adapter、adapter tests、migration root、manifest 文件、DDL review、rollback evidence、migration smoke artifact、SQL migration、schema version table、migration runner、数据库连接、Radish OIDC middleware、token validation 或 production API consumer。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 `draft_store_selector_implementation_entry_review_defined` 解释为 store selector ready、selector smoke ready、repository mode ready、repository adapter ready、database ready、OIDC ready、production API ready 或 production ready。
