# Saved Workflow Draft Adapter Smoke Readiness v1 专题

更新时间：2026-06-17

## 专题定位

`Saved Workflow Draft Adapter Smoke Readiness v1` 承接 [Saved Workflow Draft Repository Adapter Implementation Plan v1](saved-workflow-draft-repository-adapter-implementation-plan-v1.md)、[Saved Workflow Draft Schema Artifact Manifest v1](saved-workflow-draft-schema-artifact-manifest-v1.md)、[Saved Workflow Draft Store Selector Smoke Readiness v1](saved-workflow-draft-store-selector-smoke-readiness-v1.md)、[Saved Workflow Draft Auth Context Preconditions v1](saved-workflow-draft-auth-context-preconditions-v1.md) 和 [Saved Workflow Draft Repository Contract Smoke Runner Implementation v1](saved-workflow-draft-repository-contract-smoke-runner-implementation-v1.md)，用于在创建 durable repository adapter、adapter smoke fixture、schema artifact 文件、SQL migration、Radish OIDC middleware 或 production API 前，固定 future adapter smoke 的依赖、operation matrix、failure mapping、no fallback、no side effects 和 artifact guard。

本专题只定义 adapter smoke readiness，不创建 adapter smoke fixture、adapter smoke checker、repository interface、repository adapter、selector、migration root、manifest 文件、DDL review、SQL migration、数据库连接、OIDC token validation、production API consumer、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_adapter_smoke_readiness_defined`

## 当前输入事实

- `Saved Workflow Draft Repository Adapter Implementation Plan v1` 已固定 future `SavedWorkflowDraftRepository` adapter file layout、save / read / list operation adapter matrix、failure mapping、no fallback 和 no side effects。
- `Saved Workflow Draft Schema Artifact Manifest v1` 已固定 future manifest shape、section matrix、operation predicate coverage、failure mapping、no fallback 和 no side effects；后续 `Saved Workflow Draft Schema Artifact Materialization v1` 已物化静态 migration root、manifest、DDL review、rollback evidence 和 migration smoke。
- `Saved Workflow Draft Store Selector Smoke Readiness v1` 已固定 future selector smoke mode matrix、operation matrix、schema artifact failure、no fallback 和 no side effects；后续 `Saved Workflow Draft Store Selector Implementation v1` 已创建正式 config entry、selector 和 selector smoke fixture，状态为 `draft_store_selector_smoke_implemented`。
- `Saved Workflow Draft Auth Context Preconditions v1` 已固定 future repository actor context、workspace membership、owner policy、scope grants、failure policy 和 audit / sanitization 边界，但未创建 Radish OIDC middleware 或 token validation。
- `Saved Workflow Draft Repository Contract Smoke Runner Implementation v1` 已实现 static runner 和 Go tests，可作为 future adapter smoke 的 case 来源。
- 后续 [Saved Workflow Draft Adapter Smoke Execution v1](saved-workflow-draft-adapter-smoke-execution-v1.md) 已作为独立批次消费本 readiness 证据，状态为 `draft_adapter_smoke_executed`；该执行仍使用 injected fake query executor，不启用 repository store mode，不连接数据库，不接 OIDC runtime。

## Adapter Smoke Contract

future adapter smoke 路径预留为：

- `scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-v1.json`
- `scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-v1.py`

future adapter smoke 必须消费：

- static `SavedWorkflowDraftRepositoryContractSmokeRunner` cases。
- repository adapter implementation plan。
- schema artifact manifest contract 和后续 materialized evidence。
- selector smoke readiness 和后续 selector smoke fixture。
- auth context preconditions 和后续 production auth evidence。

当前只允许声明 `draft_adapter_smoke_readiness_defined`。后续 `Saved Workflow Draft Schema Artifact Materialization v1` 已物化静态 schema artifact，状态为 `draft_schema_artifact_materialized_static`。进入 adapter smoke execution 前，production auth、repository adapter implementation 和 adapter smoke fixture 仍必须分别通过后续独立准入；selector implementation 和 schema artifact materialization 已完成，但不代表 repository mode 可用。

## Adapter Smoke Gate Matrix

| gate | 当前状态 | 说明 |
| --- | --- | --- |
| repository adapter implementation plan consumed | `satisfied` | 已消费 adapter plan、future file layout 和 operation adapter matrix |
| schema artifact manifest consumed | `satisfied` | 已消费 manifest shape 和 operation predicate coverage；后续静态 manifest 文件已由 materialization v1 物化 |
| selector smoke readiness consumed | `satisfied` | 已消费 selector smoke mode / operation matrix，但未创建 selector smoke fixture |
| auth context preconditions consumed | `satisfied` | 已消费 actor context、membership、scope grant 和 audit 边界 |
| static runner consumed | `satisfied` | 已有 static runner 和 Go tests |
| adapter smoke contract defined | `satisfied` | 本专题定义 future adapter smoke 的依赖、operation、failure 和停止线 |
| selector implementation gate | `satisfied` | 已由 `Saved Workflow Draft Store Selector Implementation v1` 打开 formal config、selector function、selector tests 和 selector smoke fixture |
| schema artifact materialization gate | `satisfied` | 已由 `Saved Workflow Draft Schema Artifact Materialization v1` 物化 migration root、manifest、DDL review、rollback evidence 和 migration smoke 静态证据 |
| production auth gate | `not_satisfied` | Radish OIDC、token validation、membership adapter 和 scope projection 仍未实现 |
| repository adapter implementation gate | `not_satisfied` | repository interface、adapter、adapter tests 和 database query 仍未创建 |
| adapter smoke execution gate | `not_satisfied` | adapter smoke fixture、checker、adapter contract smoke test 和真实执行仍未创建 |
| production API consumer gate | `not_satisfied` | public production API consumer、auth policy 和 CORS policy 仍未打开 |
| no adapter smoke artifacts leaked | `required_now` | 当前必须确认没有 adapter smoke、adapter、selector、SQL、OIDC artifact 泄漏 |

## Operation Adapter Smoke Matrix

| operation | scope | future smoke 依赖 | 当前允许 |
| --- | --- | --- | --- |
| `SaveWorkflowDraftRecord` | `workflow_drafts:write` | static runner case、adapter plan、schema artifact manifest、selector smoke、auth context | 不创建 adapter smoke fixture，不创建 adapter，不写 repository / DB |
| `ReadWorkflowDraftRecord` | `workflow_drafts:read` | static runner case、adapter plan、schema artifact manifest、selector smoke、auth context | 不创建 adapter smoke fixture，不创建 adapter，不回退 sample |
| `ListWorkflowDraftRecords` | `workflow_drafts:read` | static runner case、adapter plan、schema artifact manifest、selector smoke、auth context | 不创建 adapter smoke fixture，不创建 adapter，不把 empty list 写成 sample fallback |

本 readiness 批次自身必须保持 `adapter_smoke_fixture_created_now=false`、`adapter_smoke_checker_created_now=false`、`repository_adapter_allowed_now=false`、`selector_implementation_allowed_now=false`、`schema_artifact_files_allowed_now=false`、`oidc_validation_allowed_now=false`，并且不得 fallback 到 memory dev store、sample、fixture、dev HTTP route 或 test auth。后续 selector implementation 和 schema artifact materialization 已分别由独立批次满足，不改变 adapter smoke execution、repository adapter、production auth 和 repository mode 仍未打开的事实。

## Failure Mapping

adapter smoke readiness 必须继续保留 fail-closed failure code：

- `draft_scope_denied`
- `draft_not_found`
- `draft_schema_version_unsupported`
- `draft_payload_invalid`
- `draft_version_conflict`
- `draft_store_unavailable`
- `draft_store_contract_mismatch`
- `repository_store_disabled`
- `invalid_draft_store_mode`
- `draft_auth_context_contract_mismatch`
- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`

其中：

- `draft_version_conflict` 不能被改写成 adapter smoke、selector 或 migration failure。
- `draft_scope_denied` 不能返回草案主体。
- `draft_not_found` 不能回退 sample。
- `draft_store_unavailable` 不能因为 memory dev store 可用而变成成功。
- adapter contract mismatch、schema / migration failure、selector failure 和 auth context mismatch 均不得泄露数据库、OIDC 或内部实现细节。

## No Side Effects

本批只能读取 committed fixture、文档和源码树。必须保持：

- 不创建 adapter smoke fixture 或 checker。
- 不写 repository、数据库或 schema artifact。
- 不创建 repository interface、repository adapter、selector、SQL 或 OIDC artifact。
- 不调用 workflow executor、tool executor、confirmation decision、business writeback 或 replay。
- 不启动服务、不连接数据库、不验证 token。

## 后续准入

本专题完成后，已继续补齐 [Saved Workflow Draft Store Selector Implementation Entry Review v1](saved-workflow-draft-store-selector-implementation-entry-review-v1.md)、[Saved Workflow Draft Schema Artifact Materialization Review v1](saved-workflow-draft-schema-artifact-materialization-review-v1.md)、[Saved Workflow Draft Store Selector Implementation v1](saved-workflow-draft-store-selector-implementation-v1.md)、[Saved Workflow Draft Schema Artifact Materialization v1](saved-workflow-draft-schema-artifact-materialization-v1.md)、[Saved Workflow Draft Production Auth Readiness v1](saved-workflow-draft-production-auth-readiness-v1.md)、[Saved Workflow Draft Repository Adapter Implementation Entry Review v1](saved-workflow-draft-repository-adapter-implementation-entry-review-v1.md)、`Saved Workflow Draft Repository Adapter Implementation v1` 和 [Saved Workflow Draft Adapter Smoke Execution v1](saved-workflow-draft-adapter-smoke-execution-v1.md)，状态分别为 `draft_store_selector_implementation_entry_review_defined`、`draft_schema_artifact_materialization_review_defined`、`draft_store_selector_smoke_implemented`、`draft_schema_artifact_materialized_static`、`draft_production_auth_readiness_defined`、`draft_repository_adapter_implementation_entry_review_defined`、`draft_repository_adapter_implemented` 和 `draft_adapter_smoke_executed`。durable store 方向仍必须继续按独立专题推进。后续可选方向：

1. `Saved Workflow Draft Repository Mode Enablement v1`：只能在 adapter smoke execution 后独立定义 runtime boundary、config gate、schema preflight、rollback 和 repository mode failure policy。
2. `Saved Workflow Draft Production Auth Runtime v1`：仍需独立实现 OIDC middleware、token validation 和 membership adapter，不由 adapter smoke readiness、repository adapter implementation 或 adapter smoke execution 代替。

## 验收方式

- 新增 `workflow-saved-draft-adapter-smoke-readiness-v1` fixture / checker，固定依赖证据、gate matrix、operation adapter smoke matrix、failure mapping、no fallback、no side effects 和 implementation artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 `workflow-saved-draft-schema-artifact-manifest-v1` 之后。
- 本批至少运行专项 checker、schema artifact manifest checker、repository adapter implementation plan checker、selector smoke readiness checker 和 fast repo check。

## 停止线

- 不创建 adapter smoke fixture、adapter smoke checker、repository interface、repository adapter、adapter tests、selector、selector smoke fixture、migration root、manifest 文件、DDL review artifact、rollback evidence artifact、migration smoke artifact、SQL migration、schema version table、migration runner、数据库连接、Radish OIDC middleware、token validation 或 production API consumer。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 `draft_adapter_smoke_readiness_defined` 解释为 adapter smoke ready、repository adapter ready、database ready、SQL migration ready、repository mode ready、OIDC ready、production API ready 或 production ready；schema artifact 静态证据来自后续 materialization v1，并不由本 readiness 批次单独声明。
