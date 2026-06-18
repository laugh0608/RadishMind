# Saved Workflow Draft Repository Adapter Implementation Entry Review v1 专题

更新时间：2026-06-17

## 专题定位

`Saved Workflow Draft Repository Adapter Implementation Entry Review v1` 承接 [Saved Workflow Draft Repository Adapter Implementation Plan v1](saved-workflow-draft-repository-adapter-implementation-plan-v1.md)、[Saved Workflow Draft Store Selector Implementation v1](saved-workflow-draft-store-selector-implementation-v1.md)、[Saved Workflow Draft Schema Artifact Materialization v1](saved-workflow-draft-schema-artifact-materialization-v1.md)、[Saved Workflow Draft Production Auth Readiness v1](saved-workflow-draft-production-auth-readiness-v1.md) 和 [Saved Workflow Draft Adapter Smoke Readiness v1](saved-workflow-draft-adapter-smoke-readiness-v1.md)，用于评审 saved draft durable store 是否已具备后续 repository adapter implementation task 的准入证据。

本专题只完成 entry review。结论是：store selector、schema artifact、production auth readiness、adapter implementation plan 和 adapter smoke readiness 已满足后续 repository adapter implementation task 的评审输入；但本批不创建 repository interface、repository adapter、adapter unit tests、adapter smoke fixture、database query、SQL migration、migration runner、OIDC middleware、token validation、membership adapter、production API、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_repository_adapter_implementation_entry_review_defined`

## 当前输入事实

- `Saved Workflow Draft Repository Contract Smoke Runner Implementation v1` 已实现 static contract smoke runner，可作为 future adapter contract case 来源。
- `Saved Workflow Draft Repository Adapter Implementation Plan v1` 已固定 future adapter file layout、save / read / list operation matrix、failure mapping、no fallback 和 no side effects。
- `Saved Workflow Draft Store Selector Implementation v1` 已实现 formal config、`SelectWorkflowSavedDraftStore`、selector tests 和 selector smoke checker；`repository` / `repository_disabled` 仍 fail closed。
- `Saved Workflow Draft Schema Artifact Materialization v1` 已物化 manifest、DDL review、rollback evidence 和 migration smoke 静态证据；仍没有 SQL migration、schema version table 或 migration runner。
- `Saved Workflow Draft Production Auth Readiness v1` 已固定 Radish OIDC issuer evidence、claim mapping、tenant / workspace binding、scope projection 和 auth failure mapping；仍没有 OIDC middleware、token validation 或 membership adapter。
- `Saved Workflow Draft Adapter Smoke Readiness v1` 已固定 future adapter smoke dependency gate；后续 `Saved Workflow Draft Repository Adapter Implementation v1` 和 `Saved Workflow Draft Adapter Smoke Execution v1` 已分别独立完成。

## Entry Review Decision

| candidate | 本次结论 | 后续条件 |
| --- | --- | --- |
| `SavedWorkflowDraftRepository` interface | `ready_for_next_task` | 需先创建独立 implementation task card，明确 interface 文件、operation contract、auth context 输入和 no fallback tests |
| repository adapter | `ready_for_next_task` | 需在独立实现批次中声明 adapter 边界、database query policy、schema preflight、version compare-and-set 和 sanitized projection |
| adapter unit tests | `ready_for_next_task` | 需随 adapter 实现覆盖 save / read / list、version conflict、scope denied、not found、schema mismatch 和 no fallback |
| adapter smoke fixture / checker | `completed_later` | 后续已由 `workflow-saved-draft-adapter-smoke-v1` 独立打开并验证 |
| repository store mode enablement | `blocked` | 需等 adapter smoke execution、production auth runtime 和 production API 边界完成后，再评审 repository mode 是否可启用 |

本批不创建 `workflow-saved-draft-repository-adapter-implementation-v1` task card，也不创建 repository runtime artifact。后续 implementation task card、repository adapter implementation 和 adapter smoke execution 已分别独立完成；production API、token validation、membership adapter 和 repository mode enablement 仍保持停止线。

## Gate Matrix

| gate | 当前状态 | 说明 |
| --- | --- | --- |
| repository contract smoke runner | `satisfied` | static runner 和 Go tests 已可作为 future adapter contract case 来源 |
| repository adapter implementation plan | `satisfied` | file layout、operation matrix、failure mapping 和 no fallback 已固定 |
| store selector implementation | `satisfied` | formal config 和 selector 已存在，但 repository mode 仍 fail closed |
| schema artifact materialization | `satisfied_static` | manifest、DDL review、rollback evidence 和 migration smoke 已物化为静态证据 |
| production auth readiness | `satisfied_for_entry_review` | issuer / claim / scope / failure evidence 已固定；runtime token validation 和 membership adapter 未实现 |
| adapter smoke readiness | `satisfied_for_entry_review` | smoke dependency gate 已固定；后续 adapter smoke execution 已独立打开 |
| repository adapter implementation task | `not_opened_in_this_slice` | 本批只做 entry review，不创建实现任务卡或 runtime artifact |
| production auth runtime | `not_satisfied` | 不创建 OIDC middleware、token validation 或 membership adapter |
| production API | `not_satisfied` | 不创建 public production API consumer、production auth policy 或 CORS policy |
| database execution | `not_satisfied` | 不连接数据库，不创建 database query，不运行 migration |

## Operation Entry Matrix

| operation | scope | entry review conclusion |
| --- | --- | --- |
| `SaveWorkflowDraftRecord` | `workflow_drafts:write` | 可进入后续 implementation task 设计；本批不创建 adapter、不写 repository、不执行 adapter smoke |
| `ReadWorkflowDraftRecord` | `workflow_drafts:read` | 可进入后续 implementation task 设计；本批不创建 adapter、不回退 sample、不启用 repository mode |
| `ListWorkflowDraftRecords` | `workflow_drafts:read` | 可进入后续 implementation task 设计；本批不创建 adapter、不把 empty list 写成 sample fallback |

后续 adapter 必须消费 `tenant_ref`、`workspace_id`、`application_id`、`actor_subject_ref`、`owner_subject_ref` 和 `scope_grants`。缺失 identity、tenant binding、workspace membership、owner scope 或 operation scope 时必须 fail closed。

## Failure Mapping

entry review 继续固定以下 failure code：

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
- `draft_identity_context_missing`
- `draft_tenant_binding_missing`
- `draft_workspace_membership_denied`
- `draft_scope_grant_missing`

其中 `draft_version_conflict` 不能被改写成 generic adapter failure；`draft_not_found` 不能回退 sample；`draft_scope_denied` 不能返回草案主体；schema 或 auth failure 不得泄露数据库、OIDC、token、claim dump 或内部 adapter detail。

## No Fallback / No Side Effects

本批必须保持：

- 不回退 memory dev store、sample、fixture、dev HTTP route、test auth 或 local admin。
- 不写 repository、数据库、schema version table 或 migration record。
- 不创建 repository interface、repository adapter、adapter smoke fixture、adapter smoke checker、SQL migration、migration runner、OIDC middleware、token validation、membership adapter 或 production API。
- 不调用 workflow executor、tool executor、confirmation decision、business writeback 或 replay。

## 后续准入

后续若继续 durable store，应先创建 `Saved Workflow Draft Repository Adapter Implementation v1` implementation task card，明确 repository interface、adapter file layout、adapter tests、database query boundary、schema preflight、auth context 输入、failure mapping、no fallback 和验证链路。

即使进入 repository adapter implementation，production API、Radish OIDC token validation、workspace membership adapter、repository mode enablement、SQL migration runner、publish、run、executor、confirmation、writeback 和 replay 仍必须作为后续独立专题推进；adapter smoke execution 已在后续独立批次完成，但不替代 repository mode 或 production auth runtime。

2026-06-18 后续推进先创建 `docs/task-cards/workflow-saved-draft-repository-adapter-implementation-v1-plan.md`，随后完成 repository adapter implementation，状态为 `draft_repository_adapter_implemented`。该后续实现不改变本 entry review 批次“不创建 repository runtime artifact”的历史结论。

## 验收方式

- 新增 `workflow-saved-draft-repository-adapter-implementation-entry-review-v1` fixture / checker，固定 entry decision、gate matrix、candidate matrix、operation matrix、failure mapping、no fallback、no side effects 和 forbidden artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 `workflow-saved-draft-production-auth-readiness-v1` 之后、`product-surface-readiness-implementation-trigger-recheck-v1` 之前。
- 本批至少运行专项 checker、production auth readiness checker、schema artifact materialization checker、store selector implementation checker、adapter smoke readiness checker、repository adapter implementation plan checker 和 `./scripts/check-repo.sh --fast`。
- 因同步 current focus，补跑 `./scripts/check-repo.sh`。

## 停止线

- 不创建 repository interface、repository adapter、adapter unit tests、adapter smoke fixture、adapter smoke checker、database query、SQL migration、schema version table、migration runner、数据库连接、OIDC middleware、token validation、membership adapter 或 production API consumer。
- 不启用 `repository` store mode；selector 仍只允许 `memory_dev` 成功，reserved / unknown mode fail closed。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 `draft_repository_adapter_implementation_entry_review_defined` 解释为 repository adapter ready、database ready、repository mode ready、adapter smoke ready、OIDC ready、production API ready 或 production ready。
