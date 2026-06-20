# Workflow / Agent Runtime 设计与开发文档

更新时间：2026-06-20

## 功能定位

`Workflow / Agent Runtime` 承载 AI 应用执行链路，包括 Prompt、LLM、HTTP tool、RAG retrieval、condition、output、后续受控 code / sandbox 和 agent loop。

它面向 `RadishMind` 用户工作区中的 workflow 设计、审查和后续受控执行：先让用户能保存、读取、列出和校验 workflow 草案，再进入 publish、run、confirmation、writeback 或 replay 等更高风险能力。

## 当前状态

- `apps/radishmind-web/` 已有 workflow application detail、definition detail、run detail、blocked action preview、confirmation placeholder、draft designer、validation inspector、execution plan preview、runtime readiness inspector、surface overview、context selection、scenario inspector、Review Workspace、Workspace Home 和 Review Handoff。
- 当前能力全部是 offline-only / read-only / advisory-only / blocked capability 组织层。
- `workflowWorkspaceContext` 已作为本地选择和派生关系的共享入口。
- `Saved Workflow Draft v1` 已完成 platform Go domain service 实现：`SavedWorkflowDraft` v1 类型、内存 dev store boundary、save / read / validate / list 契约、版本冲突、失败语义、sanitized response、no sample fallback 和 no side effects tests 已落地。
- 当前已实现 `Saved Workflow Draft v1` dev-only HTTP route + web consumer 状态区分，包括 save / read / validate / list 和 restore mapping；仍未实现 builder UI durable persistence、validation result persistence、execution plan persistence、runtime readiness persistence、workflow executor、confirmation decision store、writeback、replay 或 resume。
- 当前已实现 `User Workspace Saved Draft List v1`：dev-only list route 只返回 sanitized summary，Workspace Home 可显示已保存 dev draft、empty / failure state，并通过既有 read route 恢复到 Draft Designer。
- 当前已实现 `Workflow Draft Designer Editing Model v2`：Draft Designer 可本地新增节点、移动节点、删除非受保护节点、重建边，并让 validation inspector、execution plan preview 和 runtime readiness inspector 消费当前 active draft。
- 当前已实现 `Workflow Draft Node Attribute Editing Model v1`：Draft Designer 可编辑节点属性，并把节点级 summary、contract fields、output mapping 和 provider / tool / RAG refs 保存 / 恢复到 dev-only saved draft。
- 当前已实现 `Workflow Review Handoff Active Draft v1`：Review Handoff 新增 active draft review record，汇总 active draft validation inspector、execution plan preview 和 runtime readiness inspector 的状态、blocker count、request / audit metadata 和 reviewer question。
- 当前已定义并实现 Saved Workflow Draft durable store 迁移的主要前置证据链：`Saved Workflow Draft Durable Store Preconditions v1`、`Saved Workflow Draft Repository Contract Preconditions v1`、`Saved Workflow Draft Schema / Migration Preconditions v1`、`Saved Workflow Draft Auth Context Preconditions v1`、`Saved Workflow Draft Store Selector Enablement Preconditions v1`、`Saved Workflow Draft Schema Artifact Evidence v1`、`Saved Workflow Draft Store Selector Smoke Readiness v1`、`Saved Workflow Draft Repository Contract Smoke v1`、`Saved Workflow Draft Repository Contract Smoke Runner Implementation v1`、`Saved Workflow Draft Repository Adapter Implementation Plan v1`、`Saved Workflow Draft Schema Artifact Manifest v1`、`Saved Workflow Draft Adapter Smoke Readiness v1`、`Saved Workflow Draft Store Selector Implementation Entry Review v1`、`Saved Workflow Draft Schema Artifact Materialization Review v1`、`Saved Workflow Draft Store Selector Implementation v1`、`Saved Workflow Draft Schema Artifact Materialization v1`、`Saved Workflow Draft Production Auth Readiness v1`、`Saved Workflow Draft Repository Adapter Implementation Entry Review v1`、`Saved Workflow Draft Repository Adapter Implementation v1`、`Saved Workflow Draft Adapter Smoke Execution v1`、`Saved Workflow Draft Production Auth Runtime v1`、`Saved Workflow Draft Repository Mode Enablement v1`、`Saved Workflow Draft Schema Migration Runner Readiness v1`、`Saved Workflow Draft Schema Migration Runner Implementation Entry Review v1`、`Saved Workflow Draft Database Connection / Schema Marker Preconditions v1`、`Saved Workflow Draft Database Connection Provider Implementation Entry Review v1`、`Saved Workflow Draft Database Secret Resolver Readiness v1`、`Saved Workflow Draft Database Secret Resolver Implementation Entry Review v1`、platform-level `Production Secret Backend Config / Secret Ref Readiness v1`、`Production Secret Backend Provider Profile Secret Binding Readiness v1`、`Production Secret Backend Secret Resolver Interface Disabled Readiness v1`、`Production Secret Backend Operator Runbook / Negative Gates Readiness v1`、`Production Secret Backend Rotation / Audit Policy Readiness v1`、`Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review v1`、`Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy v1`、`Production Secret Backend Fake Resolver Implementation Task Card Entry Readiness Review v1`、`Production Secret Backend Fake Resolver Implementation v1`、`Production Secret Backend Fake Resolver Runtime Implementation Entry Review v1`、`Production Secret Backend Fake Resolver Runtime Implementation v1`、`Production Secret Backend Real Resolver Runtime Preconditions v1`、`Production Secret Backend Real Resolver Runtime Implementation Entry Review v1`、`Production Secret Backend Resolver Backend Profile Selection Readiness v1`、`Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Strategy v1`、`Production Secret Backend Credential Handle Runtime Boundary Readiness v1`、`Production Secret Backend Operator Approval Runtime Evidence Readiness v1` 和 `Production Secret Backend Audit Store Handoff Readiness v1` 均已固定。当前状态为 `draft_database_secret_resolver_implementation_entry_review_defined`；platform-level `fake_resolver_runtime_test_only_implemented` 只确认 test-only fake resolver runtime 已实现，`real_resolver_runtime_implementation_entry_review_defined` 只确认 production resolver runtime task card blocked，`resolver_backend_profile_selection_readiness_defined`、`real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`、`credential_handle_runtime_boundary_readiness_defined`、`operator_approval_runtime_evidence_readiness_defined` 与 `audit_store_handoff_readiness_defined` 只确认 backend profile selection、no leakage strategy、credential handle boundary、operator approval evidence 和 audit handoff 静态前置已定义，结论仍是不启用 repository store mode、不创建真实 DB artifact、不创建 backend runtime、不创建 production resolver runtime、不创建 no secret leakage smoke runtime、不创建 credential handle runtime、不执行 approval runtime、不创建 audit store / writer / event。这些证据仍不解析真实 secret、不连接数据库，不创建 SQL migration / migration runner，不实现 resolver runtime 或 production resolver runtime、OIDC middleware、token validation、membership adapter、production API 或执行链路。
- 相关状态锚点包括 `draft_repository_contract_smoke_runner_implemented`、`draft_repository_adapter_implementation_plan_defined`、`draft_schema_artifact_manifest_defined`、`draft_adapter_smoke_readiness_defined`、`draft_store_selector_implementation_entry_review_defined`、`draft_schema_artifact_materialization_review_defined`、`draft_store_selector_smoke_implemented`、`draft_schema_artifact_materialized_static`、`draft_production_auth_readiness_defined`、`draft_repository_adapter_implementation_entry_review_defined`、`draft_repository_adapter_implemented`、`draft_adapter_smoke_executed`、`draft_production_auth_runtime_bridge_implemented`、`draft_repository_mode_enablement_review_defined`、`draft_schema_migration_runner_readiness_defined`、`draft_schema_migration_runner_implementation_entry_review_defined`、`draft_database_connection_schema_marker_preconditions_defined`、`draft_database_connection_provider_implementation_entry_review_defined`、`draft_database_secret_resolver_readiness_defined`、`draft_database_secret_resolver_implementation_entry_review_defined`、`provider_profile_secret_binding_readiness_defined`、`secret_resolver_interface_disabled_readiness_defined`、`operator_runbook_negative_gates_readiness_defined`、`rotation_audit_policy_readiness_defined`、`test_fixture_strategy_fake_resolver_entry_review_defined`、`fake_resolver_contract_no_secret_leakage_smoke_strategy_defined`、`fake_resolver_implementation_task_card_entry_readiness_review_defined`、`fake_resolver_implementation_task_card_defined`、`fake_resolver_runtime_implementation_entry_review_defined`、`fake_resolver_runtime_test_only_implemented`、`real_resolver_runtime_preconditions_defined`、`real_resolver_runtime_implementation_entry_review_defined`、`resolver_backend_profile_selection_readiness_defined`、`real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`、`credential_handle_runtime_boundary_readiness_defined`、`operator_approval_runtime_evidence_readiness_defined` 和 `audit_store_handoff_readiness_defined`。
- `Workflow Saved Draft v1 Implementation` 任务卡已进入 `saved_workflow_draft_domain_service_implemented` 状态，consumer integration 由 `workflow-saved-draft-consumer-integration-v1` 承接；后续不应跳到 executor / confirmation / writeback / replay。
- 细专题已拆到 [Workflow 细专题入口](workflow/README.md)：`Saved Workflow Draft v1`、Draft Designer、User Workspace saved draft、Review Handoff 和 durable store 迁移前置专题分别承接功能、页面、用户流和后续实现准入。

## 设计边界

- draft、validation、plan、readiness、review 和 execution 是不同阶段，不能在 UI 中混成一个“已可执行”状态。
- 功能文档不再把旧停止线作为全局禁止清单复述；每个功能应说明本轮允许打开的能力、依赖的验证证据，以及需要留到后续独立目标的能力。
- 高风险 tool/action 默认要求 confirmation。
- executor 之前必须先有 run record、audit、failure taxonomy、materialized result 和 no side effects 策略。
- 上层 `RadishFlow` / `Radish` 未提供挂载点时，不阻塞本仓库设计离线草案、审查和 readiness，但不能声明真实接入 ready。

## Saved Workflow Draft v1

### 目标用户

- `Workspace Builder`：在用户工作区中编辑 workflow 草案，需要保存未完成设计、稍后继续修改，并看到结构和契约校验结果。
- `Workflow Reviewer`：审查草案是否满足发布或后续执行前置条件，需要读取同一份草案、确认风险、blocked capability 和缺失前置条件。
- `Platform Maintainer`：维护 workflow 草案 schema、校验规则、失败语义和停止线，确保后续 executor / confirmation / writeback 不会被隐式打开。

Saved Workflow Draft v1 不面向 public production API consumer，也不面向自动执行器。它的成功状态只能表示“草案已保存且可审查”，不能表示“workflow 已发布、可执行或已接入上层项目”。

### 可打开范围

Saved Workflow Draft v1 后续实现批次允许打开：

- draft save / read / validate / list 的受控入口。
- draft schema、类型定义或等价结构化格式。
- 草案版本、冲突检测、sanitized response 和 validation summary。
- 面向当前开发阶段的存储边界，例如 local file、dev store、fake store bridge 或后续经 task card 明确选择的正式存储实现。
- 与上述能力直接相关的 unit tests、negative tests、fixture、checker 和 web build 验证。

这些能力不再被旧的 `no persistence / no mutation` 口径笼统阻挡。是否进入真实数据库、OIDC、repository adapter 或 public production API，应由对应实现批次单独判断和记录；它们不是 Saved Workflow Draft v1 功能设计本身的默认前置条件。

### 当前实现

- Platform 内部已新增 Go domain service 和 memory dev store boundary，提供 `SaveDraft`、`ReadDraft`、`ValidateDraft` 和 `ListDrafts` 四个可测试入口。
- `SavedWorkflowDraft` v1 覆盖 identity、scope、version、schema、status、editable graph、contracts、provider / tool / RAG refs、requested blocked capabilities、validation summary、blocked capability summary 和 request / audit metadata。
- Go 单元测试已覆盖成功、invalid、blocked、scope denied、not found、schema unsupported、version conflict、payload too large、store unavailable、write disabled、no sample fallback、no partial write 和 no executor / confirmation / writeback / replay side effects。
- 当前实现已注册 dev-only HTTP route，并接入 web consumer 状态区分；store selector 已提供 fail-closed mode selection，`memory_dev` 使用内存 dev store，`repository_disabled` / `repository` 返回 `repository_store_disabled`，unknown mode 返回 `invalid_draft_store_mode`；repository adapter、adapter smoke execution、production auth runtime bridge、repository mode enablement、schema migration runner readiness、runner implementation entry review、database connection / schema marker preconditions、connection provider entry review、secret resolver readiness、secret resolver implementation entry review、production secret backend operator runbook / negative gates readiness、production secret backend rotation / audit policy readiness、fake resolver implementation task card entry readiness、fake resolver implementation task card、fake resolver runtime implementation entry review、test-only fake resolver runtime、真实 resolver runtime preconditions、真实 resolver runtime implementation entry review、resolver backend profile selection readiness 和真实 resolver no leakage smoke runtime strategy 已落地；仍不创建 public production API，不解析真实 secret，不接真实数据库、OIDC token validation、membership adapter、repository mode runtime enablement、production secret audit store、production resolver runtime、no secret leakage smoke runtime 或可执行 schema migration。

### Durable Store 接入说明

当前 durable store 相关代码只打开受控 adapter 边界，不打开 production persistence：

- `workflow_saved_draft_repository.go` 定义 `SavedWorkflowDraftRepository`、save / read / list request / result、stored record 和 `SavedWorkflowDraftRepositoryQueryExecutor`。`QueryExecutor` 是未来数据库或 store adapter 的注入点，不是当前 SQL 实现。
- `workflow_saved_draft_repository_adapter.go` 负责 actor context、scope grant、payload scope、schema preflight、stored record contract 和 sanitized summary 投影。schema preflight 只能证明 adapter 会 fail closed，不能证明 migration runner 已存在。
- `workflow_saved_draft_production_auth_runtime.go` 只把已验证的 `SavedWorkflowDraftVerifiedAuthContext` 与 `SavedWorkflowDraftWorkspaceBinding` 投影到 repository actor context。它不解析 token、不查询 membership、不接 OIDC middleware。
- `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 仍必须返回 `repository_store_disabled`。任何未来 runtime enablement 都需要先补真实 query executor、可执行 migration / applied marker、database connection provider、production auth middleware 和 membership adapter 的独立说明与验证；当前 `Saved Workflow Draft Database Connection Provider Implementation Entry Review v1`、`Saved Workflow Draft Database Secret Resolver Readiness v1` 和 `Saved Workflow Draft Database Secret Resolver Implementation Entry Review v1` 也确认 provider / resolver implementation 暂不打开。

### 数据边界

Saved draft 是用户工作区内的可编辑设计记录，不是 workflow definition 的发布版本，也不是 run record。v1 数据边界如下：

- 身份与作用域：`draft_id`、`workspace_id`、`application_id`、可选 `source_definition_id`、可选 `base_definition_version`、`draft_version`、`schema_version`、`draft_status`、`created_at`、`updated_at`、`created_by_actor_ref`、`updated_by_actor_ref`。
- 用户可编辑内容：草案名称、说明、节点、边、输入契约、输出契约、provider / profile 需求、tool reference、RAG reference、condition 表达式摘要和 output mapping。
- 允许的节点类型：`prompt`、`llm`、`http_tool`、`rag_retrieval`、`condition`、`output`。`code`、`sandbox`、`agent_loop` 可作为 reserved / blocked capability 标记，不在 v1 中进入可执行能力。
- 派生审查信息：risk summary、requires confirmation summary、blocked capability summary、validation state 和 validation findings 可以随 save / read / validate 返回；list 只返回 sanitized summary / metadata。它们必须能从草案主体和当前校验策略重新计算，不能成为解锁执行的真相源。
- 明确排除：secret value、API key value、OAuth / OIDC token、完整用户 claim、真实工具执行结果、materialized result、confirmation decision、execution plan persistence、runtime readiness persistence、scenario / review / handoff persistence、run input / output、business writeback payload、replay / resume state。

实现批次若需要 committed schema 或持久化格式，必须先拆 task card，并按风险新增 fixture / checker；功能文档本身只定义设计边界。

### 保存流程

1. UI 从 draft designer 生成草案提交载荷，必须包含 workspace、application、schema version、draft version 和节点 / 边主体。
2. 服务端或本地实现入口先做 normalize：排序稳定化、默认值补齐、字段 allowlist、禁止字段剔除和长度 / 数量预算检查。
3. 校验器执行结构校验：节点 id 唯一、边端点存在、图结构可遍历、输入输出契约存在、节点类型在 allowlist 内、required provider/profile/tool/reference 只保存 ref 不保存 secret。
4. 校验器执行能力校验：发现 executor、confirmation decision、writeback、replay、public API、真实数据库、OIDC、repository adapter 或 production-only 字段时，必须标记 blocked 或拒绝保存。
5. 可恢复的业务不完整状态允许保存为 `invalid` 或 `blocked`，例如缺 provider/profile ref、缺 input contract 或存在 blocked capability；不可解析、越权、schema 不兼容或结构损坏的载荷必须拒绝保存。
6. 保存必须是原子写入：成功时创建或更新同一 `draft_id`，递增 `draft_version`，返回 sanitized draft、validation summary 和 audit/request metadata；失败时不得产生部分草案，也不得回退写入 sample / fixture。

### 读取流程

1. 读取必须按 workspace + application + draft scope 查询；未来接入真实身份前，只允许使用明确标记的 dev / fake auth context，不声明 production auth ready。
2. `draft_id` 不存在、scope 不匹配或 store 不可用时 fail closed；不能静默返回离线样例作为“已保存草案”。
3. 读取结果只返回 sanitized draft、版本信息、validation summary、blocked capability summary 和 request / audit metadata。
4. 读取时应重新运行或复核校验策略，确保旧草案在策略升级后能显示 `schema_version_unsupported`、`validation_policy_changed` 或新的 blocked finding。
5. UI 可以继续展示离线样例，但必须明确标记为 sample / unsaved，不能和 saved draft record 混同。

### 校验流程

Saved Workflow Draft v1 的校验目标是“可保存、可审查、可进入后续实现前置评估”，不是“可执行”。校验分层如下：

- Schema 校验：schema version、required fields、字段类型、长度预算、节点 / 边数组、引用字段和 forbidden fields。
- Graph 校验：节点 id 唯一、边端点存在、入口 / 出口可解释、循环策略明确、孤立节点和 unreachable output 给出 finding。
- Contract 校验：input contract、output contract、LLM prompt input、tool input/output、RAG retrieval input 和 condition output 的最小结构一致。
- Capability 校验：executor、confirmation decision、writeback、replay、真实数据库、OIDC、repository adapter、production API、secret resolver 和 public delivery 字段必须保持 blocked。
- Risk 校验：高风险 tool/action 必须产生 `requires_confirmation` finding；由于 v1 不实现 confirmation decision，它只能阻止发布 / 执行，不能解锁运行。

校验结果允许返回 `valid_for_review`、`invalid_draft`、`blocked_capability` 或 `schema_unsupported`。`valid_for_review` 只表示草案可以被审查或进入下一批实现前置评估，不表示 publish / run ready。

### 失败语义

Saved Workflow Draft v1 采用 fail-closed 语义。建议固定以下失败码，并在实现批次中保持 response、日志和测试一致：

| failure code | 语义 | 行为 |
| --- | --- | --- |
| `draft_scope_denied` | workspace / application / actor scope 不匹配 | 拒绝读写，不返回草案主体 |
| `draft_not_found` | `draft_id` 不存在或不属于当前 scope | 返回 not found，不回退 sample |
| `draft_schema_version_unsupported` | 草案 schema 版本不被当前实现支持 | 拒绝保存；读取时只返回可审查 metadata 和升级提示 |
| `draft_payload_invalid` | JSON / 类型 / required fields 无法解析 | 拒绝保存，无部分写入 |
| `draft_graph_invalid` | 节点、边、入口、出口或循环策略不合法 | 可保存为 `invalid` 的情况需有明确 finding；结构损坏则拒绝 |
| `draft_contract_invalid` | input / output / tool / RAG / condition 契约不一致 | 可保存为 `invalid`，不得进入 publish / run |
| `draft_blocked_capability` | 包含 executor、confirmation decision、writeback、replay 等 v1 禁止能力 | 可保存为 `blocked` 或拒绝，取决于字段是否可安全降级为标记 |
| `draft_version_conflict` | 客户端基于旧 `draft_version` 写入 | 拒绝覆盖，返回当前版本 metadata |
| `draft_payload_too_large` | 节点数、边数或文本字段超出预算 | 拒绝保存 |
| `draft_store_unavailable` | 保存或读取存储不可用 | fail closed，不回退 fixture，不声明成功 |
| `draft_store_contract_mismatch` | store 返回的 record 与 tenant / workspace / application / owner / draft contract 不一致 | 拒绝返回草案主体，不回退 dev store |
| `draft_write_disabled` | 当前环境只允许只读或 sample 模式 | 拒绝保存，允许继续查看 sample |
| `repository_store_disabled` | store selector 选择了 reserved repository mode | fail closed，不回退 memory dev、sample 或 fixture |
| `invalid_draft_store_mode` | store selector mode 不在 allowlist 内 | fail closed，不回退 memory dev、sample 或 fixture |
| `draft_auth_context_contract_mismatch` | repository actor context、operation 或 auth source 不可信 | 拒绝读写，不回退 dev auth |
| `draft_schema_migration_not_applied` | schema preflight 缺少 applied migration marker | 拒绝进入 repository adapter 成功路径 |
| `draft_store_schema_version_mismatch` | store schema version 与 `saved_workflow_drafts_store_v1` 不一致 | 拒绝进入 repository adapter 成功路径 |
| `draft_store_migration_unavailable` | migration 状态不可观测 | 拒绝进入 repository adapter 成功路径 |
| `draft_identity_context_missing` | actor subject 缺失 | 拒绝 repository 读写 |
| `draft_tenant_binding_missing` | tenant binding 缺失或不一致 | 拒绝 repository 读写 |
| `draft_workspace_membership_denied` | workspace membership 未验证 | 拒绝 repository 读写 |
| `draft_application_scope_denied` | application scope 未验证 | 拒绝 repository 读写 |
| `draft_owner_scope_denied` | owner scope 未验证 | 拒绝 repository 读写 |
| `draft_scope_grant_missing` | 缺少 `workflow_drafts:read` 或 `workflow_drafts:write` | 拒绝对应 operation |
| `draft_audit_context_missing` | audit ref 缺失 | 拒绝 production auth runtime bridge 成功路径 |

所有失败都必须保留 no side effects：不执行节点、不调用 provider、不调用 tool、不写上层业务真相源、不创建 run record、不提交 confirmation decision。

## 下一批开发方向

1. Saved Workflow Draft v1 consumer integration、受控本地编辑、用户工作区创建入口、User Workspace saved draft list / restore、Draft Designer 本地结构编辑、节点属性编辑、Workflow Review Handoff Active Draft v1、durable store 迁移证据链、selector implementation、静态 schema artifact materialization、production auth readiness、repository adapter implementation entry review、repository adapter implementation、adapter smoke execution、production auth runtime bridge、repository mode enablement 准入评审、schema migration runner readiness、schema migration runner implementation entry review、database connection / schema marker preconditions、database connection provider implementation entry review、database secret resolver readiness、database secret resolver implementation entry review、production secret backend config / secret ref readiness、Production Secret Backend Provider Profile Secret Binding Readiness v1、Production Secret Backend Secret Resolver Interface Disabled Readiness v1、Production Secret Backend Operator Runbook / Negative Gates Readiness v1、Production Secret Backend Rotation / Audit Policy Readiness v1、Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review v1、Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy v1、Production Secret Backend Fake Resolver Implementation Task Card Entry Readiness Review v1、Production Secret Backend Fake Resolver Implementation v1、Production Secret Backend Fake Resolver Runtime Implementation Entry Review v1、Production Secret Backend Fake Resolver Runtime Implementation v1、Production Secret Backend Real Resolver Runtime Preconditions v1、Production Secret Backend Real Resolver Runtime Implementation Entry Review v1、Production Secret Backend Resolver Backend Profile Selection Readiness v1、Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Strategy v1、Production Secret Backend Credential Handle Runtime Boundary Readiness v1、Production Secret Backend Operator Approval Runtime Evidence Readiness v1 和 Production Secret Backend Audit Store Handoff Readiness v1 已完成，新增 platform-level 状态为 `audit_store_handoff_readiness_defined`。下一步若继续 durable store 的 secret backend 上游，应在 backend health boundary、saved draft connection provider、schema marker contract implementation、真实 SQL migration runner implementation、OIDC middleware / token validation、membership adapter 或 production API 中选择单一前置方向重新开题；不能直接启用 repository mode、production API、OIDC token validation、membership adapter、production resolver runtime、no secret leakage smoke runtime、approval runtime、audit store runtime 或执行链路。
2. 如果继续新增 consumer contract、fixture 或 checker，应先更新任务卡或对应细专题，明确 route 不属于 public production API，也不接真实数据库、OIDC、repository adapter、schema migration 或 store selector。
3. 若只整理现有离线审查 UI、文案或阅读路径，不再新增逐项 task card / fixture / checker，复用现有 workflow 聚合 gate、web build 和 fast baseline。
4. 可执行 run、confirmation decision、writeback 和 replay 是 Saved Workflow Draft v1 之后的独立功能目标；进入前必须重新定义 run record、audit、failure taxonomy、materialized result、no side effects、人工确认和执行存储前置条件。

## 验收方式

- 本次功能设计文档更新：确认 Saved Workflow Draft v1 已明确目标用户、数据边界、保存 / 读取 / 校验流程、失败语义、验收方式和停止线；运行必要文档 / 仓库检查。
- Saved Workflow Draft v1 domain service 实现批次：schema 或等价类型定义、保存 / 读取 / 校验单元测试、negative tests、version conflict tests、storage boundary checks、no sample fallback tests、no side effects tests 和 `./scripts/check-repo.sh --fast`。
- Saved Workflow Draft v1 consumer integration 批次：按选择补 dev-only route / web consumer / consumer contract / fixture / checker，至少运行 Go 单元测试、web build、workflow 聚合 checker 和 `./scripts/check-repo.sh --fast`。
- 若实现批次改变阶段口径、治理入口、schema 真相源或高风险执行边界，补跑全量 `./scripts/check-repo.sh`。
- executor / confirmation / writeback / replay：必须作为后续独立目标，另开 task card，补 audit tests、人工确认路径、no side effects tests 和全量仓库验证。

## 停止线

- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、business writeback、run replay、run resume 或 materialized result reader。
- 不把 Saved Workflow Draft v1 的草案存储边界直接扩张成 repository adapter、真实 schema migration、Radish OIDC、token validation、public production API consumer、API key lifecycle、quota enforcement、billing 或 cost ledger；当前 selector 只提供 fail-closed mode selection，后续若选择 durable repository 或 schema materialization，应回到对应功能文档和实现批次单独推进。
- 不把 saved draft 的 `valid_for_review`、validation summary、risk summary 或 readiness summary 解释为 production ready、publish ready 或 executable ready。
- 不继续新增同层只读 evidence 面板来替代 Saved Workflow Draft v1；普通只读整理只复用现有 workflow gate、web build 和 fast baseline。
- 不把 `RadishFlow` / `Radish` 上层挂载点缺失写成本功能阻塞；同时也不声明真实接入 ready。
