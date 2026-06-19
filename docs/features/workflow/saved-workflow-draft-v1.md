# Saved Workflow Draft v1 功能专题

更新时间：2026-06-19

## 专题定位

`Saved Workflow Draft v1` 是 `Workflow / Agent Runtime` 中第一个从只读审查面走向可保存用户草案的功能专题。它让用户工作区中的 workflow 草案可以被保存、读取、列出和校验，并让 reviewer 看到同一份可审查记录。

本专题不是 workflow publish / run / executor 专题。成功保存只表示“草案已保存且可审查”，不表示 workflow 已发布、可执行或已接入上层项目。

## 当前状态

- Platform Go domain service 已实现，文件为 `services/platform/internal/httpapi/workflow_saved_draft.go`。
- 已覆盖 `SavedWorkflowDraft` v1 类型、memory dev store、`SaveDraft` / `ReadDraft` / `ValidateDraft` / `ListDrafts`、版本冲突、失败语义、sanitized response、no sample fallback 和 no side effects tests。
- 当前已新增 dev-only HTTP route 和 web consumer 状态区分：`POST /v1/user-workspace/workflow-drafts`、`GET /v1/user-workspace/workflow-drafts/{draft_id}`、`GET /v1/user-workspace/workflow-drafts` 和 `POST /v1/user-workspace/workflow-drafts/validate` 默认关闭，只在显式 dev 配置下工作。
- 当前已补 route contract 和 consumer smoke：Go route test 固定 envelope、header、CORS、not found / store unavailable no sample fallback；前端 consumer 固定 `version_conflict` 状态，version conflict 时保留本地草案并展示当前 saved draft version metadata。
- 当前已接入 [Workflow Draft Editing Entry v1](draft-editing-entry-v1.md)：Draft Designer 可编辑草案名称、说明、节点名称和边条件摘要，validate / save / read 使用当前本地草案。
- 当前已接入 [User Workspace Draft Creation v1](user-workspace-draft-creation-v1.md)：用户可从 Workspace Home 或 workflow definitions 创建本地草案，再复用 dev-only saved draft consumer 保存。
- 当前已接入 [User Workspace Saved Draft List v1](user-workspace-saved-draft-list-v1.md)：Workspace Home 可读取当前 application 的 saved dev draft sanitized summary，并通过 read route 恢复到 Draft Designer。
- 当前已接入 [Workflow Draft Designer Editing Model v2](draft-designer-editing-model-v2.md)：Draft Designer 可本地新增节点、移动节点、删除非受保护节点并重建边，validate / save / read 继续消费当前 active draft。
- 当前已接入 [Workflow Draft Node Attribute Editing Model v1](draft-node-attribute-editing-model-v1.md)：Draft Designer 可编辑节点级 provider / profile、tool ref、RAG ref、summary、contract fields 和 output mapping，dev-only saved draft payload / response 会保存并恢复节点级 summary、contract fields、output mapping 和 refs。
- 当前已接入 [Workflow Review Handoff Active Draft v1](review-handoff-active-draft-v1.md)：Review Handoff 会把恢复后的 active draft validation inspector、execution plan preview 和 runtime readiness inspector 汇总为 advisory-only 审查交接记录。
- 当前已新增 [Saved Workflow Draft Durable Store Preconditions v1](saved-workflow-draft-durable-store-preconditions-v1.md)：固定 durable store 迁移前的 draft scope、`owner_subject_ref` / workspace 归属、version conflict、no sample fallback、dev store 与未来 repository adapter 的切换停止线；它只定义前置设计，不实现 durable persistence。
- 当前已新增 [Saved Workflow Draft Repository Contract Preconditions v1](saved-workflow-draft-repository-contract-preconditions-v1.md)：固定 future `SavedWorkflowDraftRepository` 的 `SaveWorkflowDraftRecord`、`ReadWorkflowDraftRecord` 和 `ListWorkflowDraftRecords` contract preconditions；它只定义 actor context、request / result、failure 和 projection 边界，不创建 repository interface。
- 当前已新增 [Saved Workflow Draft Schema / Migration Preconditions v1](saved-workflow-draft-schema-migration-preconditions-v1.md)：固定 future durable store 的 logical schema、index strategy、migration gate、failure mapping、no sample fallback 和 artifact guard；状态为 `draft_schema_migration_preconditions_defined`，不创建真实数据库 schema 或 SQL migration。
- 当前已新增 [Saved Workflow Draft Auth Context Preconditions v1](saved-workflow-draft-auth-context-preconditions-v1.md)：固定 future repository actor context 的身份来源、workspace membership、owner policy、scope grants、failure policy 和 audit / sanitization 边界；状态为 `draft_auth_context_preconditions_defined`，不创建 Radish OIDC middleware 或 token validation。
- 当前已新增 [Saved Workflow Draft Store Selector Enablement Preconditions v1](saved-workflow-draft-store-selector-enablement-preconditions-v1.md)：固定 future `memory_dev` / `repository_disabled` / `repository` store mode、selector gate、failure mapping、no fallback、no side effects 和 artifact guard；状态为 `draft_store_selector_enablement_preconditions_defined`，不创建正式 config entry 或 selector 代码。
- 当前已新增 [Saved Workflow Draft Schema Artifact Evidence v1](saved-workflow-draft-schema-artifact-evidence-v1.md)：固定 future schema artifact manifest、DDL review、rollback evidence、migration smoke、failure mapping 和 artifact guard；状态为 `draft_schema_artifact_evidence_defined`，不创建 migration root、manifest、DDL review artifact 或 SQL migration。
- 当前已新增 [Saved Workflow Draft Store Selector Smoke Readiness v1](saved-workflow-draft-store-selector-smoke-readiness-v1.md)：固定 future selector smoke 的 `memory_dev` / `repository_disabled` / `repository` / unknown mode、save / read / list operation matrix、schema artifact failure、no fallback 和 no side effects；状态为 `draft_store_selector_smoke_readiness_defined`，不创建正式 config entry、selector 代码或 selector smoke fixture。
- 当前已新增 [Saved Workflow Draft Repository Contract Smoke v1](saved-workflow-draft-repository-contract-smoke-v1.md)：固定 future repository contract smoke 的 `SavedWorkflowDraftRepositoryContractSmoke` harness、save / read / list I/O、failure mapping、no sample / memory dev fallback 和 no side effects；状态为 `draft_repository_contract_smoke_defined`，不创建 smoke runner、repository interface、repository adapter、store selector 或数据库 artifact。
- 当前已新增 [Saved Workflow Draft Repository Contract Smoke Runner Readiness v1](saved-workflow-draft-repository-contract-smoke-runner-readiness-v1.md)：固定 future `SavedWorkflowDraftRepositoryContractSmokeRunner` 的 runner I/O、save / read / list operation runner matrix、failure mapping、no fallback 和 no side effects；状态为 `draft_repository_contract_smoke_runner_readiness_defined`，不创建 runner Go 文件、runner test、repository interface、repository adapter、store selector 或数据库 artifact。
- 当前已新增 [Saved Workflow Draft Repository Contract Smoke Runner Implementation v1](saved-workflow-draft-repository-contract-smoke-runner-implementation-v1.md)：实现 static `SavedWorkflowDraftRepositoryContractSmokeRunner` 和 Go tests；状态为 `draft_repository_contract_smoke_runner_implemented`，仍不创建 repository interface、repository adapter、store selector、SQL migration、真实数据库、Radish OIDC middleware、token validation 或 production API artifact。
- 当前已新增 [Saved Workflow Draft Repository Adapter Implementation Plan v1](saved-workflow-draft-repository-adapter-implementation-plan-v1.md)：固定 future repository adapter implementation plan、operation adapter matrix、failure mapping、no memory dev / sample fallback 和 no side effects；状态为 `draft_repository_adapter_implementation_plan_defined`，仍不创建 repository interface、repository adapter、store selector、SQL migration、真实数据库、Radish OIDC middleware、token validation 或 production API artifact。
- 当前已新增 [Saved Workflow Draft Schema Artifact Manifest v1](saved-workflow-draft-schema-artifact-manifest-v1.md)：固定 future schema artifact manifest shape、section matrix、operation predicate coverage、failure mapping、no memory dev / sample fallback 和 no side effects；状态为 `draft_schema_artifact_manifest_defined`，仍不创建 migration root、manifest 文件、DDL review artifact、SQL migration、真实数据库、repository adapter、store selector、Radish OIDC middleware、token validation 或 production API artifact。
- 当前已新增 [Saved Workflow Draft Adapter Smoke Readiness v1](saved-workflow-draft-adapter-smoke-readiness-v1.md)：固定 future adapter smoke readiness、依赖 gate、operation adapter smoke matrix、failure mapping、no memory dev / sample / fixture / dev route / test auth fallback 和 no side effects；状态为 `draft_adapter_smoke_readiness_defined`，仍不创建 adapter smoke fixture、adapter smoke checker、repository interface、repository adapter、selector、schema artifact 文件、SQL migration、真实数据库、Radish OIDC middleware、token validation 或 production API artifact。
- 当前已新增 [Saved Workflow Draft Store Selector Implementation Entry Review v1](saved-workflow-draft-store-selector-implementation-entry-review-v1.md)：评审 formal config entry、`SelectWorkflowSavedDraftStore`、selector unit tests 和 selector smoke fixture 是否打开；状态为 `draft_store_selector_implementation_entry_review_defined`，当前结论是不打开 selector implementation entry，不创建 implementation task card、formal config、selector、selector tests 或 selector smoke fixture。
- 当前已新增 [Saved Workflow Draft Schema Artifact Materialization Review v1](saved-workflow-draft-schema-artifact-materialization-review-v1.md)：评审 migration root、manifest file、DDL review artifact、rollback evidence artifact 和 migration smoke artifact 是否打开；状态为 `draft_schema_artifact_materialization_review_defined`，当前结论是不打开 schema artifact materialization entry，不创建 implementation task card、migration root、manifest、DDL review、rollback evidence、migration smoke、SQL 或 migration runner。
- 当前已新增 [Saved Workflow Draft Store Selector Implementation v1](saved-workflow-draft-store-selector-implementation-v1.md)：实现 formal config、`SelectWorkflowSavedDraftStore`、selector tests、HTTP fail-closed tests 和 selector smoke checker；状态为 `draft_store_selector_smoke_implemented`。当前 selector 只允许 `memory_dev` 成功，`repository_disabled` / `repository` / unknown mode 均 fail closed。
- 当前已新增 [Saved Workflow Draft Schema Artifact Materialization v1](saved-workflow-draft-schema-artifact-materialization-v1.md)：物化 `manifest.json`、`ddl-review.md`、`rollback-evidence.json` 和 `migration-smoke.json` 静态证据；状态为 `draft_schema_artifact_materialized_static`。本批不创建 SQL migration、schema version table、migration runner、数据库连接、repository adapter、adapter smoke fixture、Radish OIDC middleware、token validation 或 production API。
- 当前已新增 [Saved Workflow Draft Production Auth Readiness v1](saved-workflow-draft-production-auth-readiness-v1.md)：固定 Radish OIDC issuer discovery evidence、token validation contract preconditions、claim mapping、tenant / workspace / application binding、scope projection、failure mapping、no fake fallback 和 no side effects；状态为 `draft_production_auth_readiness_defined`。本批不创建 OIDC middleware、token validation、membership adapter、repository adapter、production API 或 adapter smoke fixture。
- 当前已新增 [Saved Workflow Draft Repository Adapter Implementation Entry Review v1](saved-workflow-draft-repository-adapter-implementation-entry-review-v1.md)：评审 repository adapter implementation task 准入、operation matrix、failure mapping、no fallback 和 no side effects；状态为 `draft_repository_adapter_implementation_entry_review_defined`。本批不创建 repository interface、repository adapter、adapter smoke fixture、database query、OIDC middleware、token validation、membership adapter 或 production API。
- 当前已完成 [Workflow Saved Draft Repository Adapter Implementation v1 任务卡](../../task-cards/workflow-saved-draft-repository-adapter-implementation-v1-plan.md)：实现 `SavedWorkflowDraftRepository` interface、注入式 query executor adapter、schema preflight、auth context failure mapping 和 adapter unit tests；状态为 `draft_repository_adapter_implemented`。本批仍不创建 adapter smoke fixture、database connection、SQL migration、migration runner、OIDC middleware、token validation、membership adapter、production API 或执行链路。
- 当前已完成 [Saved Workflow Draft Adapter Smoke Execution v1](saved-workflow-draft-adapter-smoke-execution-v1.md)：使用 static contract smoke cases、`SavedWorkflowDraftRepositoryAdapter` 和 injected fake query executor 执行 save / read / list adapter smoke；状态为 `draft_adapter_smoke_executed`。本批不启用 `repository` store mode，不连接数据库，不运行 SQL，不接 OIDC runtime，不创建 production API 或执行链路。
- 当前已完成 [Saved Workflow Draft Production Auth Runtime v1](saved-workflow-draft-production-auth-runtime-v1.md)：实现 `SavedWorkflowDraftVerifiedAuthContext` + `SavedWorkflowDraftWorkspaceBinding` 到 `SavedWorkflowDraftRepositoryActorContext` 的 runtime bridge；状态为 `draft_production_auth_runtime_bridge_implemented`。本批不创建 OIDC middleware、token validation、membership adapter、repository mode enablement、production API 或数据库连接。
- 当前已完成 [Saved Workflow Draft Repository Mode Enablement v1](saved-workflow-draft-repository-mode-enablement-v1.md)：固定 repository mode runtime boundary、config gate、schema preflight、adapter smoke / production auth runtime dependency、failure mapping、rollback、no fallback 和 no side effects；状态为 `draft_repository_mode_enablement_review_defined`。本批结论是不启用 repository mode，`repository` / `repository_disabled` 仍返回 `repository_store_disabled`。
- 当前已完成 [Saved Workflow Draft Schema Migration Runner Readiness v1](saved-workflow-draft-schema-migration-runner-readiness-v1.md)：定义 future schema migration runner 的 manual boundary、config gate、schema preflight、failure mapping、rollback、no fallback 和 no side effects；状态为 `draft_schema_migration_runner_readiness_defined`。本批不创建 SQL migration、schema version table、migration runner、runner command、database query executor 或数据库连接。
- 当前已完成 [Saved Workflow Draft Schema Migration Runner Implementation Entry Review v1](saved-workflow-draft-schema-migration-runner-implementation-entry-review-v1.md)：评审 runner implementation 是否打开，状态为 `draft_schema_migration_runner_implementation_entry_review_defined`。本批固定 executable migration artifact、schema version marker contract、manual runner command、dry-run plan、migration apply smoke、rollback observability 和 repository mode runtime enablement 均为 blocked，不创建 runner implementation task card、SQL、schema version table、runner、database query executor 或数据库连接。
- 当前已完成 [Saved Workflow Draft Database Connection / Schema Marker Preconditions v1](saved-workflow-draft-database-connection-schema-marker-preconditions-v1.md)：固定 future database connection provider、secret ref、query role、environment isolation 和 schema marker read / write contract 的前置条件；状态为 `draft_database_connection_schema_marker_preconditions_defined`。本批不创建数据库连接、secret resolver、SQL migration、schema version table、schema marker reader / writer、database query executor、migration runner、runner command 或 repository mode 成功路径。
- 当前已完成 [Saved Workflow Draft Database Connection Provider Implementation Entry Review v1](saved-workflow-draft-database-connection-provider-implementation-entry-review-v1.md)：评审 database connection provider implementation 是否打开，状态为 `draft_database_connection_provider_implementation_entry_review_defined`。本批固定 secret resolver、driver policy、connection factory、runtime query role policy、connection smoke 和 repository query executor binding 均为 blocked，不创建 provider implementation task card、DB driver、connection provider、secret resolver、database query executor 或 repository mode 成功路径。
- 当前已完成 [Saved Workflow Draft Database Secret Resolver Readiness v1](saved-workflow-draft-database-secret-resolver-readiness-v1.md)：固定 future saved draft database secret resolver 的 secret ref key、resolver input / result shape、disabled runtime、sanitized diagnostics、environment binding、failure taxonomy 和 offline fake resolver strategy；状态为 `draft_database_secret_resolver_readiness_defined`。本批不创建 secret resolver implementation、fake resolver、connection provider、DB driver、connection factory、connection smoke、SQL、schema marker、migration runner 或 repository mode 成功路径。
- 当前已完成 [Saved Workflow Draft Database Secret Resolver Implementation Entry Review v1](saved-workflow-draft-database-secret-resolver-implementation-entry-review-v1.md)：评审 secret resolver implementation 是否打开，状态为 `draft_database_secret_resolver_implementation_entry_review_defined`。后续 `Production Secret Backend Config / Secret Ref Readiness v1` 已固定 `config_secret_ref_readiness_defined`，`Production Secret Backend Provider Profile Secret Binding Readiness v1` 已固定 `provider_profile_secret_binding_readiness_defined`，`Production Secret Backend Secret Resolver Interface Disabled Readiness v1` 已固定 `secret_resolver_interface_disabled_readiness_defined`，`Production Secret Backend Operator Runbook / Negative Gates Readiness v1` 已固定 `operator_runbook_negative_gates_readiness_defined`，`Production Secret Backend Rotation / Audit Policy Readiness v1` 已固定 `rotation_audit_policy_readiness_defined`，`Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review v1` 已固定 `test_fixture_strategy_fake_resolver_entry_review_defined`，`Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy v1` 已固定 `fake_resolver_contract_no_secret_leakage_smoke_strategy_defined`，`Production Secret Backend Fake Resolver Implementation Task Card Entry Readiness Review v1` 已固定 `fake_resolver_implementation_task_card_entry_readiness_review_defined`，`Production Secret Backend Fake Resolver Implementation v1` 已固定 `fake_resolver_implementation_task_card_defined`，但这些证据只满足 config 注入点、reference-only manifest 消费边界、provider/profile binding、future resolver interface disabled result、operator runbook、negative gates、脱敏验证、smoke record reference、rotation trigger、audit event fields、secret ref version reference、test fixture strategy / fake resolver entry review、fake resolver static contract / no secret leakage smoke strategy、fake resolver implementation task card entry readiness 和 fake resolver implementation 静态任务卡前置；production secret backend resolver 仍为 `not_started`，fake resolver runtime、no secret leakage smoke runtime、sanitized diagnostics runtime、connection factory handoff、rotation runtime、audit store 和 repository mode integration 均为 blocked，不创建 resolver runtime、fake resolver runtime 或任何 DB artifact。
- 当前仍没有 durable persistence、真实 SQL migration、schema version table、schema marker reader / writer、secret resolver implementation、database connection provider、migration runner、Radish OIDC middleware、token validation、membership adapter、repository mode enablement runtime 或 production API。
- 当前功能推进状态为 `draft_database_secret_resolver_implementation_entry_review_defined`；早期 [Workflow Saved Draft v1 Implementation](../../task-cards/workflow-saved-draft-v1-implementation-plan.md) 任务卡仍记录 domain service 实现批次。platform-level `test-fixture-strategy` 与 fake resolver implementation entry 已由 `production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1` 推进为 `test_fixture_strategy_fake_resolver_entry_review_defined` 可检查证据，fake resolver static contract 与 no secret leakage smoke strategy 已由 `production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1` 固定为 `fake_resolver_contract_no_secret_leakage_smoke_strategy_defined`，fake resolver implementation task card entry readiness 已由 `production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1` 固定为 `fake_resolver_implementation_task_card_entry_readiness_review_defined`，fake resolver implementation 静态任务卡已由 `production-secret-backend-fake-resolver-implementation-v1` 固定为 `fake_resolver_implementation_task_card_defined`。后续 durable store 若继续推进，可以单独评审 fake resolver runtime implementation，或在真实 resolver runtime preconditions、saved draft connection provider implementation、schema marker contract implementation、SQL migration runner implementation、OIDC middleware / token validation、membership adapter 或 production API 中选择单一前置方向重新开题，而不是直接启用 repository mode。

## 目标用户

- `Workspace Builder`：保存未完成的 workflow 设计，后续继续编辑。
- `Workflow Reviewer`：读取同一份草案，审查结构、风险和 blocked capability。
- `Platform Maintainer`：维护 schema、校验规则、失败语义和停止线，避免后续 executor 或 confirmation 被隐式打开。

## 数据边界

Saved draft 是用户工作区中的可编辑设计记录，不是 published workflow definition，也不是 run record。

允许保存：

- `draft_id`、`workspace_id`、`application_id`、`source_definition_id`、`base_definition_version`、`draft_version`、`schema_version`、`draft_status`。
- 未来 durable store 前必须补齐 `tenant_ref`、`owner_subject_ref`、`created_by_actor_ref`、`updated_by_actor_ref`、workspace membership 和 tenant predicate；当前 dev-only memory store 只作为开发态 dev store，不代表 owner / workspace production auth 已就绪。
- 草案名称、说明、节点、边、输入契约、输出契约、provider / profile ref、tool ref、RAG ref、condition 摘要和 output mapping。
- 节点级 summary、contract fields、output mapping 和 provider / tool / RAG refs；这些字段只描述草案设计，不解锁 runtime binding。
- validation summary、blocked capability summary、request / audit metadata。

明确排除：

- secret value、API key value、OAuth / OIDC token、完整用户 claim。
- 真实工具执行结果、materialized result、confirmation decision。
- execution plan persistence、runtime readiness persistence、scenario / review / handoff persistence。
- run input / output、business writeback payload、replay / resume state。

## 平台实现说明

当前平台实现分成四层，后续维护时应按职责修改，不把任一层解释为生产持久化已经启用：

| 文件 | 职责 | 边界 |
| --- | --- | --- |
| `workflow_saved_draft.go` | `SavedWorkflowDraft` v1 domain service、memory dev store、dev-only save / read / validate / list route 消费的失败语义 | 只承载开发态草案记录，不连接数据库，不执行 workflow |
| `workflow_saved_draft_repository.go` | `SavedWorkflowDraftRepository` interface、save / read / list request / result、stored record 和 `SavedWorkflowDraftRepositoryQueryExecutor` 注入边界 | 只定义 adapter 与未来 query executor 的窄契约，不实现 SQL、migration runner 或数据库连接 |
| `workflow_saved_draft_repository_adapter.go` | `NewSavedWorkflowDraftRepositoryAdapter`、schema preflight、actor context 检查、stored record contract 检查和 sanitized projection | 只通过注入式 query executor 调用未来 store；没有 query executor 时返回 `draft_store_unavailable`，不得回退 memory dev、sample 或 fixture |
| `workflow_saved_draft_production_auth_runtime.go` | 把已验证的 `SavedWorkflowDraftVerifiedAuthContext` 与 `SavedWorkflowDraftWorkspaceBinding` 投影为 repository actor context | 只接受已验证输入，不做 OIDC middleware、token validation 或 membership lookup |

`SavedWorkflowDraftRepositorySchemaPreflight` 当前只检查 `saved_workflow_drafts_store_v1` 和 migration state。`not_applied` 必须返回 `draft_schema_migration_not_applied`，`unavailable` 必须返回 `draft_store_migration_unavailable`，store schema mismatch 必须返回 `draft_store_schema_version_mismatch`。这些 failure code 说明 schema gate 未满足，不代表 runner 已存在。

production auth runtime bridge 的唯一允许 auth source 是 `radish_oidc_verified_context`。这是上游已验证上下文的来源标记，不是本仓库已经实现 Radish OIDC token validation。workspace / application / owner scope 也必须由输入 binding 明确标记为 verified，缺失时应 fail closed。

## 功能流程

保存流程：

1. UI 或 consumer 提交 sanitized draft payload，必须包含 workspace、application、schema version、draft version 和 graph 主体。
2. 服务端执行 normalize、field allowlist、forbidden field reject 和大小预算检查。
3. 校验 schema、graph、contract、capability 和 risk。
4. 成功时原子保存并递增 `draft_version`，返回 sanitized draft、validation summary、blocked capability summary 和 request / audit metadata。
5. 失败时不得产生部分写入，不得回退 sample / fixture，不得创建 run record 或 confirmation decision。

读取流程：

1. 按 workspace + application + draft scope 查询。
2. scope mismatch、not found 或 store unavailable 必须 fail closed。
3. 读取结果只返回 sanitized draft、版本信息、validation summary、blocked capability summary 和 request / audit metadata。
4. UI 可以继续展示 sample，但必须明确标记 `sample / unsaved`，不能混成 saved draft record。

校验流程：

- `valid_for_review` 只表示草案可审查，不表示 publish ready 或 run ready。
- `invalid_draft` 可用于可恢复的不完整业务草案。
- `blocked_capability` 可用于保留 code、sandbox、agent_loop、executor、confirmation decision、writeback、replay 等禁止能力的审查 finding。
- `schema_unsupported` 用于旧 schema 或不支持 schema 的可审查失败状态。

## 失败语义

必须保持以下 failure code 和行为一致：

| failure code | 行为 |
| --- | --- |
| `draft_scope_denied` | 拒绝读写，不返回草案主体 |
| `draft_not_found` | 返回 not found，不回退 sample |
| `draft_schema_version_unsupported` | 拒绝保存；读取时只返回可审查 metadata |
| `draft_payload_invalid` | 拒绝保存，无部分写入 |
| `draft_graph_invalid` | 结构损坏则拒绝；可恢复问题可保存为 invalid finding |
| `draft_contract_invalid` | 可保存为 invalid，不允许 publish / run |
| `draft_blocked_capability` | 可保存为 blocked finding 或拒绝危险字段 |
| `draft_version_conflict` | 拒绝覆盖，返回当前版本 metadata |
| `draft_payload_too_large` | 拒绝保存 |
| `draft_store_unavailable` | fail closed，不回退 fixture |
| `draft_store_contract_mismatch` | 拒绝读取 / 列出不符合 stored record contract 的记录，不回退 dev store |
| `draft_write_disabled` | 拒绝保存，允许继续查看 sample |
| `repository_store_disabled` | reserved repository store 未启用，fail closed，不回退 memory dev |
| `invalid_draft_store_mode` | store mode 配置非法，fail closed，不回退 memory dev |
| `draft_auth_context_contract_mismatch` | repository actor context 或 operation contract 不可信，拒绝读写 |
| `draft_schema_migration_not_applied` | schema preflight 发现 migration 未应用，拒绝进入 repository adapter 成功路径 |
| `draft_store_schema_version_mismatch` | stored record 或 preflight schema version 与 `saved_workflow_drafts_store_v1` 不一致 |
| `draft_store_migration_unavailable` | schema migration 状态不可观测，拒绝进入 repository adapter 成功路径 |
| `draft_identity_context_missing` | production / repository actor subject 缺失 |
| `draft_tenant_binding_missing` | tenant 为空或 auth tenant 与 workspace binding 不一致 |
| `draft_workspace_membership_denied` | workspace membership 未验证 |
| `draft_application_scope_denied` | application scope 未验证 |
| `draft_owner_scope_denied` | owner scope 未验证 |
| `draft_scope_grant_missing` | 缺少 `workflow_drafts:read` 或 `workflow_drafts:write` scope |
| `draft_audit_context_missing` | audit ref 缺失，拒绝进入 production auth runtime bridge 成功路径 |

## 下一批开发

dev-only consumer integration 已按 [Dev-only Saved Draft Consumer](dev-only-saved-draft-consumer.md) 落地，并已补 route contract、consumer smoke 和 version conflict UI 状态；正式草案编辑入口、用户工作区创建、saved dev draft list / restore、本地图结构编辑、节点属性编辑和 active draft review record 均已落地；durable store 迁移前置设计、repository contract、schema / auth / selector evidence、static runner、schema artifact、production auth readiness、repository adapter implementation、adapter smoke execution、production auth runtime bridge、repository mode enablement 准入评审、schema migration runner readiness、runner implementation entry review、database connection / schema marker preconditions、connection provider entry review、secret resolver readiness、secret resolver implementation entry review、production secret backend config / secret ref readiness、Production Secret Backend Provider Profile Secret Binding Readiness v1、Production Secret Backend Secret Resolver Interface Disabled Readiness v1、Production Secret Backend Operator Runbook / Negative Gates Readiness v1、Production Secret Backend Rotation / Audit Policy Readiness v1、Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review v1、Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy v1、Production Secret Backend Fake Resolver Implementation Task Card Entry Readiness Review v1 和 Production Secret Backend Fake Resolver Implementation v1 均已完成。后续若继续 durable store 的 secret backend 上游，可以单独评审 fake resolver runtime implementation，或在真实 resolver runtime preconditions、saved draft connection provider implementation、schema marker contract implementation、真实 SQL migration runner implementation、OIDC middleware / token validation、membership adapter 或 production API consumer 中选择单一前置方向独立评审；任何 durable persistence、public production API、OIDC token validation、membership adapter、repository mode enablement runtime、fake resolver runtime、no secret leakage smoke runtime 或 executor 仍必须作为独立专题和 task card 推进。

## 验收方式

- Go 单元测试覆盖 save / read / validate / list 成功和失败路径。
- Consumer 能区分 sample / unsaved draft 与 saved draft record。
- Workspace Home 能读取 saved dev draft sanitized summary，并通过 read route 恢复到 Draft Designer；empty / failure state 不回退 sample。
- Draft Designer 本地结构编辑后的 active draft 能继续 validate / save / read，validation inspector、execution plan preview 和 runtime readiness inspector 能消费当前 active draft。
- Draft Designer 节点属性编辑后的 active draft 能保存和恢复节点级 summary、contract fields、output mapping 和 refs。
- Review Handoff 能把 active draft validation inspector、execution plan preview 和 runtime readiness inspector 汇总为 active draft review record，且不保存、不导出、不发送 handoff。
- Repository contract smoke 定义能覆盖 save / read / list I/O、failure mapping、no sample / memory dev fallback 和 no side effects，且不创建 smoke runner、repository interface、repository adapter 或数据库 artifact。
- Repository contract smoke runner readiness 能覆盖 future runner I/O、save / read / list operation runner matrix、failure mapping、no fallback 和 no side effects，且不创建 runner Go 文件、repository interface、repository adapter 或数据库 artifact。
- Repository contract smoke runner implementation 能覆盖 static runner、Go tests、failure mapping、no fallback 和 no side effects，且不创建 repository interface、repository adapter、selector、SQL、OIDC 或 production API artifact。
- Repository adapter implementation plan 能覆盖 future adapter file layout、save / read / list operation adapter matrix、failure mapping、no memory dev / sample fallback 和 no side effects，且不创建 repository interface、repository adapter、selector、SQL、OIDC 或 production API artifact。
- Schema artifact manifest 能覆盖 future manifest shape、section matrix、operation predicate coverage、failure mapping、no memory dev / sample fallback 和 no side effects，且不创建 migration root、manifest 文件、DDL review、rollback evidence、migration smoke、SQL、repository adapter、selector、OIDC 或 production API artifact。
- Adapter smoke readiness 能覆盖 future adapter smoke dependency gate、save / read / list operation adapter smoke matrix、failure mapping、no memory dev / sample / fixture / dev route / test auth fallback 和 no side effects，且不创建 adapter smoke fixture、adapter smoke checker、repository interface、repository adapter、selector、schema artifact、SQL、OIDC 或 production API artifact。
- Store selector implementation entry review 能覆盖 formal config、`SelectWorkflowSavedDraftStore`、selector tests 和 selector smoke fixture 四个候选项，固定当前不打开 implementation entry，且不创建 formal config、selector、selector tests、selector smoke fixture、repository adapter、schema artifact、SQL、OIDC 或 production API artifact。
- Schema artifact materialization review 能覆盖 migration root、manifest file、DDL review artifact、rollback evidence artifact 和 migration smoke artifact 五个候选项，固定当前不打开 implementation entry，且不创建 materialization task card、migration root、schema artifact、SQL、migration runner、repository adapter、selector、OIDC 或 production API artifact。
- Store selector implementation 能覆盖 formal config、`SelectWorkflowSavedDraftStore`、selector tests、HTTP fail-closed tests 和 selector smoke checker，且 `repository_disabled` / `repository` / unknown mode 不回退 memory dev、sample 或 fixture。
- Schema artifact materialization 能覆盖 manifest、DDL review、rollback evidence、migration smoke、operation predicate coverage、failure mapping、no fallback 和 no side effects，且不创建 SQL migration、migration runner、数据库连接、repository adapter、OIDC 或 production API artifact。
- Production auth readiness 能覆盖 issuer discovery evidence contract、token validation contract preconditions、claim mapping、tenant / workspace / application binding、scope projection、failure mapping、no fake fallback、no side effects 和 downstream readiness review，且不创建 OIDC middleware、token validation、membership adapter、repository adapter、adapter smoke fixture 或 production API artifact。
- Repository adapter implementation entry review 能覆盖 repository interface / adapter / adapter unit tests 的后续实现准入、adapter smoke / repository mode / production API 的继续阻塞、operation matrix、failure mapping、no fallback 和 no side effects，且不创建 repository interface、repository adapter、database query、adapter smoke fixture、OIDC runtime 或 production API artifact。
- Repository adapter implementation 能覆盖 repository interface、adapter boundary、adapter unit tests、schema preflight、auth context 输入、failure mapping、no fallback 和验证链路，且不启用 repository mode、不运行 adapter smoke、不接 OIDC runtime、数据库连接或 production API。
- Adapter smoke execution 能覆盖 static contract smoke cases、adapter save / read / list、fake query executor 调用边界、version conflict、not found、store unavailable、store contract mismatch、auth context mismatch、schema preflight failure、no fallback 和 no side effects，且不启用 repository mode、不连接数据库、不接 OIDC runtime 或 production API。
- Production auth runtime bridge 能覆盖 verified auth context、workspace / application / owner binding、operation scope projection、failure mapping、no fake fallback 和 no side effects，且不创建 OIDC middleware、token validation、membership adapter、repository mode、数据库连接或 production API。
- Repository mode enablement 准入评审能覆盖 runtime boundary、config gate、schema preflight、adapter smoke / production auth runtime dependency、failure mapping、rollback、no fallback 和 no side effects，且结论仍是不启用 repository mode、不连接数据库、不写 SQL / migration runner、不接 OIDC middleware、token validation、membership adapter 或 production API。
- Schema migration runner readiness 能覆盖 future runner 的 manual boundary、config gate、schema preflight、applied marker 缺口、failure mapping、rollback、no fallback 和 no side effects，且不创建 SQL migration、schema version table、runner、database query executor、数据库连接或 repository mode runtime。
- Schema migration runner implementation entry review 能覆盖 executable migration artifact、schema version marker contract、manual runner command、dry-run plan、migration apply smoke、rollback observability 和 repository mode runtime enablement 的 blocked 结论，且不创建 runner implementation task card、SQL migration、schema version table、runner command、database query executor、数据库连接或 repository mode runtime。
- Database connection / schema marker preconditions 能覆盖 future database connection provider、secret ref、query role、environment isolation、schema marker read / write contract、failure mapping、no fallback 和 no side effects，且不创建数据库连接、secret resolver、SQL migration、schema version table、schema marker reader / writer、database query executor、migration runner 或 repository mode runtime。
- Database connection provider implementation entry review 能覆盖 secret resolver、driver / DSN policy、connection factory、runtime / migration role policy、connection smoke、failure mapping、no fallback 和 no side effects，且不创建 connection provider implementation task card、DB driver、secret resolver、connection factory、database query executor、SQL、schema marker reader / writer、migration runner 或 repository mode runtime。
- Database secret resolver implementation entry review 能覆盖 production secret backend `not_started`、reference-only secret manifest、fake resolver blocked、sanitized diagnostics、failure mapping、no fallback、no side effects 和 artifact guard，且不创建 resolver implementation task card、resolver runtime、fake resolver、DB provider / driver、SQL、schema marker、migration runner 或 repository mode runtime。
- Production Secret Backend Operator Runbook / Negative Gates Readiness v1 能覆盖 operator runbook、test / production secret source、operator approval evidence、sanitized verification、smoke record reference、negative gates、failure mapping、no fallback、no side effects 和 artifact guard，且不创建 resolver runtime、fake resolver、cloud secret SDK、DB provider / driver、connection factory、SQL、schema marker、migration runner、repository mode runtime 或 public production API。
- Production Secret Backend Rotation / Audit Policy Readiness v1 能覆盖 rotation trigger、audit event fields、secret ref version reference、rollback policy、sanitized verification、failure mapping、no fallback、no side effects 和 artifact guard，且不创建 rotation runtime、production secret audit store、audit writer、resolver runtime、fake resolver、cloud secret SDK、DB provider / driver、connection factory、SQL、schema marker、migration runner、repository mode runtime 或 public production API。
- Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review v1 能覆盖 `test-fixture-strategy` 与 fake resolver implementation entry 的 blocked 结论、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard，且不创建 fake resolver implementation task card、resolver runtime、fake resolver runtime、no secret leakage smoke runtime、cloud secret SDK、DB provider / driver、connection factory、SQL、schema marker、migration runner、repository mode runtime 或 public production API。
- Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy v1 能覆盖 fake resolver static contract、输入 / 输出 allowlist、secret-bearing forbidden fields、no secret leakage smoke strategy、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard，且不创建 fake resolver implementation task card、resolver runtime、fake resolver runtime、no secret leakage smoke runtime、cloud secret SDK、DB provider / driver、connection factory、SQL、schema marker、migration runner、repository mode runtime 或 public production API。
- Production Secret Backend Fake Resolver Implementation Task Card Entry Readiness Review v1 能覆盖 `production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1`、`fake_resolver_implementation_task_card_entry_readiness_review_defined`、task card entry decision、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard，且只允许下一步创建任务卡，不实现 resolver runtime、fake resolver runtime、no secret leakage smoke runtime、cloud secret SDK、DB provider / driver、connection factory、SQL、schema marker、migration runner、repository mode runtime 或 public production API。
- Consumer 能区分 `version_conflict`，并在冲突时保留本地草案、展示当前版本 metadata，不把失败回退成 sample。
- route contract 和 consumer smoke checker 进入 fast baseline。
- Web build 和 workflow 相关聚合检查通过。
- `./scripts/check-repo.sh --fast` 通过。
- 若新增 committed schema、route contract、fixture 或 checker，先更新 task card，并按风险补全专项验证。

## 停止线

- 不实现 publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不接真实数据库、database connection provider、secret resolver、fake resolver、真实 schema migration、schema version table、schema marker reader / writer、migration runner、Radish OIDC middleware、token validation、membership adapter、API key lifecycle、quota enforcement、billing 或 public production API；当前 store selector 只提供 fail-closed mode selection，repository adapter 已实现并通过 adapter smoke execution，production auth runtime bridge 已实现，repository mode enablement、schema migration runner readiness、schema migration runner implementation entry review、database connection / schema marker preconditions、connection provider implementation entry review、secret resolver readiness、secret resolver implementation entry review、production secret backend operator runbook / negative gates readiness、production secret backend rotation / audit policy readiness、production secret backend test fixture strategy / fake resolver entry review、production secret backend fake resolver contract / no secret leakage smoke strategy 和 production secret backend fake resolver implementation task card entry readiness review 已完成准入、策略或前置证据评审但不启用 repository mode、不创建 DB runtime artifact、不创建 fake resolver runtime、不创建 no secret leakage smoke runtime、不创建 audit store。
- 不把 `valid_for_review`、validation summary、risk summary 或 readiness summary 当作执行解锁条件。
