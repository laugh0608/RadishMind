# Workflow / Agent Runtime 设计与开发文档

更新时间：2026-07-16

## 功能定位

`Workflow / Agent Runtime` 承载 AI 应用执行链路，包括 Prompt、LLM、HTTP tool、RAG retrieval、condition、output、后续受控 code / sandbox 和 agent loop。

它面向 `RadishMind` 用户工作区中的 workflow 设计、审查和后续受控执行：先让用户能保存、读取、列出和校验 workflow 草案，再进入 publish、run、confirmation、writeback 或 replay 等更高风险能力。

## 当前状态

2026-07-17 当前结论：Saved Draft durable dev/test repository、R4 Gateway、executor v0、Run History、Failure Review、Run Comparison、Evaluation Cases、[Baseline / Case Versioning](workflow/workflow-evaluation-baseline-case-versioning-v1.md)、[Evaluation Suite / Release Review](workflow/workflow-evaluation-suite-release-review-v1.md)、Gateway Request History 与 Gateway Playground 已完成。[Workflow 受控 HTTP Tool 与人工确认执行（开发 / 测试态）v1](workflow/controlled-http-tool-human-confirmation-dev-test-v1.md) 的三个批次也已完成；独立 `/executions`、Web、run v2 与双数据库浏览器重启链状态为 `workflow_controlled_http_tool_human_confirmation_dev_test_v1_completed`。下一步先设计 RAG retrieval 的开发 / 测试态纵向切片；业务写回、publish、replay / resume、agent loop 与 production enablement 继续关闭。以下 production durable-store readiness 只保留为历史兼容锚点。

2026-06-29 Production Secret Backend audit store runtime blocker matrix 及其后续 storage adapter readiness / review 链只作为历史静态锚点保留，不再定义当前顺位，也不影响已经完成的 Workflow memory / SQLite / PostgreSQL 开发测试态存储。production secret、production audit store、production repository mode 与公开生产 API 仍未启用。

- `apps/radishmind-web/` 已有 workflow application detail、definition detail、run detail、blocked action preview、confirmation placeholder、draft designer、validation inspector、execution plan preview、runtime readiness inspector、surface overview、context selection、scenario inspector、Review Workspace、Workspace Home 和 Review Handoff。
- 当前页面同时包含离线 / 只读审查层和显式开发测试态执行层；默认离线入口不发请求，执行器、持久运行历史与 Gateway 调用只在受控配置下启用，不能据此声明生产能力。
- `workflowWorkspaceContext` 已作为本地选择和派生关系的共享入口。
- `Saved Workflow Draft v1` 已完成 platform Go domain service 实现：`SavedWorkflowDraft` v1 类型、内存 dev store boundary、save / read / validate / list 契约、版本冲突、失败语义、sanitized response、no sample fallback 和 no side effects tests 已落地。
- `Saved Workflow Draft v1` 已完成 dev-only HTTP route、Web consumer、memory / SQLite / PostgreSQL 开发测试态 repository、save / read / validate / list / restore、版本冲突与重启恢复；`Workflow Executor v0` 已完成 Prompt / LLM / condition / output 受控执行和 scoped run record。受控 HTTP Tool 已沿独立版本化 action plan / confirmation / run v2 路径完成三个批次；不得在 v0 上原地放开。
- 当前已实现 `User Workspace Saved Draft List v1`：dev-only list route 只返回 sanitized summary，Workspace Home 可显示已保存 dev draft、empty / failure state，并通过既有 read route 恢复到 Draft Designer。
- 当前已实现 `Workflow Draft Designer Editing Model v2`：Draft Designer 可本地新增节点、移动节点、删除非受保护节点、重建边，并让 validation inspector、execution plan preview 和 runtime readiness inspector 消费当前 active draft。
- 当前已实现 `Workflow Draft Node Attribute Editing Model v1`：Draft Designer 提供节点属性编辑，并把节点级 summary、contract fields、output mapping 和 provider / tool / RAG refs 保存 / 恢复到 dev-only saved draft。
- 当前已实现 `Workflow Review Handoff Active Draft v1`：Review Handoff 新增 active draft review record，汇总 active draft validation inspector、execution plan preview 和 runtime readiness inspector 的状态、blocker count、request / audit metadata 和 reviewer question。
- 当前已定义 `Workflow Node Designer Surface v1`，完成 `Workflow Node Designer Library Selection v1`，实现 `Workflow Node Designer Surface Implementation v1` 任务卡，完成 `Workflow Node Designer Saved Draft Mapping v1` 的 UI-only implementation，并实现 `Workflow Node Designer Review Handoff v1`、`Workflow Node Designer Persisted Layout v1`、`Workflow Node Designer Edge Editing Save Preconditions v1`、`Workflow Node Designer Controlled Edge Mutation Implementation v1`、`Workflow Node Designer Layout Review Findings v1`、`Workflow Node Designer Builder Interaction Polish v1`、`Workflow Node Designer Validation Overlay Navigation v1` 与 `Workflow Node Designer Graph Review Handoff Refinement v1`：节点画布已接入首批前端 `@xyflow/react` 画布，active draft session 可记录拖拽后的 layout，Save Draft 会把受控节点坐标写入 `additional_fields.designer_layout_v1`，restore 会兼容旧草案和非法 layout metadata；edge editing 已固定只允许通过 active draft 写入 `edgeId`、`fromNodeId`、`toNodeId`、transient `edgeKind` 和 `conditionSummary`，saved draft payload 仍只保存 endpoint 与 condition summary，并要求 validation inspector 继续消费 mutated active draft；Builder 布局已补中宽响应式、inspector connected edge 删除条目和 Draft edge 卡片长 id 可读性，已补交互状态条、节点快速选择和连接 / 拖拽 / 删除反馈，并把 validation finding 映射为节点 / 连线 / inspector 的 UI-only 导航和高亮；Review Handoff 的 `nodeDesignerReviewRecord` 已汇总 canvas layout、validation overlay、inspector state、saved draft mapping 和 `graphReviewFindings`，按 node / edge / graph-level finding 展示审查目标，并补充 `handoffPath`、`handoffPathRefs` 与 evidence refs 阅读路径；该 Builder 链本身不创建 runtime artifact，也不改变 executor v0 的能力边界。
- 当前已定义并实现 Saved Workflow Draft durable store 迁移的主要前置证据链：durable store preconditions、repository contract、schema / migration preconditions、auth context、store selector、schema artifact、repository smoke runner、repository adapter、adapter smoke execution、production auth runtime bridge、Radish OIDC token / membership readiness / implementation entry review / upstream evidence refresh、token validation schema / auth middleware runtime entry review、token validation schema task card readiness / implementation task card / artifact implementation、auth middleware / membership adapter task card entry readiness、negative auth smoke runtime readiness、repository mode enablement、schema migration runner readiness / entry review、database connection / schema marker preconditions、connection provider entry review / entry refresh v2、driver / DSN / TLS、role policy、connection smoke strategy、connection lifecycle、database secret resolver readiness / implementation entry review / runtime dependency refresh、repository mode runtime boundary review、schema marker / migration runner readiness refresh、schema marker contract implementation entry review、manual migration runner entry refresh 和 schema marker runtime dependency refresh，以及 platform-level Production Secret Backend 从 config / secret ref readiness 到 production resolver blocker consolidation、credential handle refresh、operator approval refresh、cloud service selection、backend health refresh、no leakage refresh、audit store durable backend / writer / runtime schema / delivery / idempotency readiness、audit store runtime implementation entry refresh v5、production resolver runtime implementation entry refresh v2、runtime event schema artifact entry review、runtime event schema artifact implementation task card、runtime event schema artifact、audit store runtime blocker matrix、audit store durable backend selection readiness、audit store concrete durable backend selection review、audit store writer runtime implementation entry review、audit store idempotency runtime implementation entry review 和 audit store delivery runtime implementation entry review 的静态证据均已固定。当前状态为 `audit_store_runtime_implementation_entry_refresh_v5_defined`；该状态只表示 metadata-only schema artifact、fixtures、checker、writer compatibility smoke、schema artifact 后续 blocker matrix、durable backend family selection、writer runtime implementation entry review、idempotency runtime implementation entry review、delivery runtime implementation entry review 和 audit store runtime task card still blocked 结论已离线验证，不能解锁 repository store mode、真实 DB artifact、storage adapter runtime、production resolver runtime、backend health runtime、no leakage smoke runtime、credential handle runtime、approval runtime、audit store runtime、writer runtime、delivery runtime、idempotency runtime、repository mode runtime task card 或 production API。
- 保留 resolver runtime 停止线字面锚点：不实现 resolver runtime。
- 保留 Saved Workflow Draft durable store 文档锚点：`Saved Workflow Draft Schema Artifact Manifest v1`、`Saved Workflow Draft Token Validation Schema / Auth Middleware Runtime Entry Review v1`、`Saved Workflow Draft Token Validation Schema Task Card Readiness v1`、`Saved Workflow Draft Token Validation Schema Implementation v1`、`Saved Workflow Draft Token Validation Schema Artifact Implementation v1`、`Saved Workflow Draft Auth Middleware / Membership Adapter Task Card Entry Readiness v1`、`Saved Workflow Draft Negative Auth Smoke Runtime Readiness v1`、`Saved Workflow Draft Repository Mode Enablement Review v1`、`Saved Workflow Draft Database Connection Provider Implementation Entry Refresh v1`、`Saved Workflow Draft Database Connection Provider Implementation Entry Refresh v2`、`Saved Workflow Draft Database Connection Lifecycle Readiness v1`、`Saved Workflow Draft Schema Marker Contract Implementation Entry Review v1`、`Saved Workflow Draft Manual Migration Runner Implementation Entry Refresh v1`、`Saved Workflow Draft Schema Marker Runtime Dependency Refresh v1` 和 `Saved Workflow Draft Database Secret Resolver Runtime Dependency Refresh v1`。
- 保留 checker 锚点：`Saved Workflow Draft Store Selector Implementation Entry Review v1`、`Saved Workflow Draft Schema Migration Runner Readiness v1`、`Saved Workflow Draft Schema Migration Runner Implementation Entry Review v1`、`Saved Workflow Draft Schema Marker / Migration Runner Readiness Refresh v1`、`Saved Workflow Draft Schema Artifact Materialization Review v1`、`Saved Workflow Draft Adapter Smoke Readiness v1`、`Saved Workflow Draft Repository Adapter Implementation Entry Review v1`、`Saved Workflow Draft Production Auth Readiness v1`、`Radish OIDC Token / Membership Readiness v1`、`Radish OIDC Token / Membership Implementation Entry Review v1`、`Radish OIDC Token / Membership Upstream Evidence Refresh v1`、`Saved Workflow Draft Database Connection / Schema Marker Preconditions v1`、`Saved Workflow Draft Database Connection Provider Implementation Entry Review v1`、`Saved Workflow Draft Database Driver / DSN / TLS Policy Readiness v1`、`Saved Workflow Draft Database Role Policy Readiness v1`、`Saved Workflow Draft Database Connection Smoke Strategy v1`、`Saved Workflow Draft Database Secret Resolver Readiness v1`、`Saved Workflow Draft Database Secret Resolver Implementation Entry Review v1`、`Saved Workflow Draft Database Secret Resolver Runtime Dependency Refresh v1`、`Saved Workflow Draft Repository Mode Runtime Boundary Review v1`、`Production Secret Backend Config / Secret Ref Readiness v1`、`Production Secret Backend Provider Profile Secret Binding Readiness v1`、`Production Secret Backend Secret Resolver Interface Disabled Readiness v1`、`Production Secret Backend Operator Runbook / Negative Gates Readiness v1`、`Production Secret Backend Rotation / Audit Policy Readiness v1`、`Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review v1`、`Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy v1`、`Production Secret Backend Fake Resolver Implementation Task Card Entry Readiness Review v1`、`Production Secret Backend Fake Resolver Implementation v1`、`Production Secret Backend Fake Resolver Runtime Implementation Entry Review v1`、`Production Secret Backend Fake Resolver Runtime Implementation v1`、`Production Secret Backend Real Resolver Runtime Preconditions v1`、`Production Secret Backend Real Resolver Runtime Implementation Entry Review v1`、`Production Secret Backend Real Resolver Runtime Implementation Entry Refresh v1`、`Production Secret Backend Resolver Backend Profile Selection Readiness v1`、`Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Strategy v1`、`Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Implementation Entry Review v1`、`Production Secret Backend Credential Handle Runtime Boundary Readiness v1`、`Production Secret Backend Credential Handle Runtime Implementation Entry Review v1`、`Production Secret Backend Credential Handle Runtime Implementation Entry Refresh v1`、`Production Secret Backend Operator Approval Runtime Evidence Readiness v1`、`Production Secret Backend Operator Approval Runtime Implementation Entry Review v1`、`Production Secret Backend Operator Approval Runtime Implementation Entry Refresh v1`、`Production Secret Backend Cloud Secret Service Selection Readiness v1`、`Production Secret Backend Resolver Backend Health Runtime Implementation Entry Refresh v1`、`Production Secret Backend Audit Store Handoff Readiness v1`、`Production Secret Backend Audit Store Contract / Event Schema Readiness v1`、`Production Secret Backend Audit Store Ownership Boundary Readiness v1`、`Production Secret Backend Audit Store Delivery / Idempotency Runtime Boundary Readiness v1`、`Production Secret Backend Audit Store Runtime Implementation Entry Refresh v3`、`Production Secret Backend Audit Store Durable Backend Boundary Readiness v1`、`Production Secret Backend Audit Store Writer Runtime Boundary Readiness v1`、`Production Secret Backend Audit Store Runtime Event Schema Materialization Readiness v1`、`Production Secret Backend Audit Store Delivery Runtime Readiness v1`、`Production Secret Backend Audit Store Idempotency Runtime Readiness v1`、`Production Secret Backend Audit Store Runtime Implementation Entry Refresh v4`、`Production Secret Backend Audit Store Runtime Event Schema Artifact Implementation Entry Review v1`、`Production Secret Backend Audit Store Runtime Event Schema Artifact Implementation v1`、`Production Secret Backend Audit Store Runtime Event Schema Artifact v1`、`Production Secret Backend Audit Store Runtime Blocker Matrix v1`、`Production Secret Backend Audit Store Durable Backend Selection Readiness v1`、`Production Secret Backend Audit Store Concrete Durable Backend Selection Review v1`、`Production Secret Backend Audit Store Runtime Implementation Entry Refresh v5`、`Production Secret Backend Audit Store Writer Runtime Implementation Entry Review v1`、`Production Secret Backend Audit Store Idempotency Runtime Implementation Entry Review v1`、`Production Secret Backend Audit Store Delivery Runtime Implementation Entry Review v1`、`Production Secret Backend Resolver Backend Health Boundary Readiness v1`、`Production Secret Backend Resolver Backend Health Runtime Implementation Entry Review v1`、`Production Secret Backend Production Resolver Runtime Blocker Consolidation v1`、`Production Secret Backend Production Resolver Runtime Implementation Entry Refresh v2`。
- 相关状态锚点包括 `draft_repository_contract_smoke_runner_implemented`、`draft_repository_adapter_implementation_plan_defined`、`draft_schema_artifact_manifest_defined`、`draft_adapter_smoke_readiness_defined`、`draft_store_selector_implementation_entry_review_defined`、`draft_schema_artifact_materialization_review_defined`、`draft_store_selector_smoke_implemented`、`draft_schema_artifact_materialized_static`、`draft_production_auth_readiness_defined`、`draft_repository_adapter_implementation_entry_review_defined`、`draft_repository_adapter_implemented`、`draft_adapter_smoke_executed`、`draft_production_auth_runtime_bridge_implemented`、`radish_oidc_token_membership_readiness_defined`、`radish_oidc_token_membership_implementation_entry_review_defined`、`radish_oidc_token_membership_upstream_evidence_refresh_defined`、`draft_token_validation_auth_middleware_runtime_entry_review_defined`、`draft_token_validation_schema_task_card_readiness_defined`、`draft_token_validation_schema_implementation_task_card_defined`、`draft_token_validation_schema_artifact_implemented`、`draft_auth_middleware_membership_adapter_task_card_entry_readiness_defined`、`draft_negative_auth_smoke_runtime_readiness_defined`、`production_resolver_runtime_blocker_consolidation_defined`、`production_resolver_runtime_implementation_entry_refresh_v2_defined`、`credential_handle_runtime_implementation_entry_refresh_defined`、`operator_approval_runtime_implementation_entry_refresh_defined`、`cloud_secret_service_selection_readiness_defined`、`resolver_backend_health_runtime_implementation_entry_refresh_defined`、`real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_refresh_defined`、`audit_store_durable_backend_boundary_readiness_defined`、`audit_store_writer_runtime_boundary_readiness_defined`、`audit_store_runtime_event_schema_materialization_readiness_defined`、`audit_store_delivery_runtime_readiness_defined`、`audit_store_idempotency_runtime_readiness_defined`、`audit_store_runtime_implementation_entry_refresh_v4_defined`、`audit_store_runtime_event_schema_artifact_implementation_entry_review_defined`、`audit_store_runtime_event_schema_artifact_implementation_task_card_defined`、`audit_store_runtime_event_schema_artifact_implemented`、`audit_store_runtime_blocker_matrix_defined`、`audit_store_durable_backend_selection_readiness_defined`、`audit_store_concrete_durable_backend_selection_review_defined`、`audit_store_runtime_implementation_entry_refresh_v5_defined`、`audit_store_writer_runtime_implementation_entry_review_defined`、`audit_store_idempotency_runtime_implementation_entry_review_defined`、`audit_store_delivery_runtime_implementation_entry_review_defined`、`draft_repository_mode_enablement_review_defined`、`draft_schema_migration_runner_readiness_defined`、`draft_schema_migration_runner_implementation_entry_review_defined`、`draft_database_connection_schema_marker_preconditions_defined`、`draft_database_connection_provider_implementation_entry_review_defined`、`draft_database_connection_provider_implementation_entry_refresh_defined`、`draft_database_connection_provider_implementation_entry_refresh_v2_defined`、`draft_database_driver_dsn_tls_policy_readiness_defined`、`draft_database_role_policy_readiness_defined`、`draft_database_connection_smoke_strategy_defined`、`draft_database_connection_lifecycle_readiness_defined`、`draft_database_secret_resolver_readiness_defined`、`draft_database_secret_resolver_implementation_entry_review_defined`、`draft_database_secret_resolver_runtime_dependency_refresh_defined`、`draft_repository_mode_runtime_boundary_review_defined`、`draft_schema_marker_migration_runner_readiness_refresh_defined`、`draft_schema_marker_contract_implementation_entry_review_defined`、`draft_manual_migration_runner_implementation_entry_refresh_defined`、`draft_schema_marker_runtime_dependency_refresh_defined`、`provider_profile_secret_binding_readiness_defined`、`secret_resolver_interface_disabled_readiness_defined`、`operator_runbook_negative_gates_readiness_defined`、`rotation_audit_policy_readiness_defined`、`test_fixture_strategy_fake_resolver_entry_review_defined`、`fake_resolver_contract_no_secret_leakage_smoke_strategy_defined`、`fake_resolver_implementation_task_card_entry_readiness_review_defined`、`fake_resolver_implementation_task_card_defined`、`fake_resolver_runtime_implementation_entry_review_defined`、`fake_resolver_runtime_test_only_implemented`、`real_resolver_runtime_preconditions_defined`、`real_resolver_runtime_implementation_entry_review_defined`、`real_resolver_runtime_implementation_entry_refresh_defined`、`resolver_backend_profile_selection_readiness_defined`、`real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`、`real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined`、`credential_handle_runtime_boundary_readiness_defined`、`credential_handle_runtime_implementation_entry_review_defined`、`operator_approval_runtime_evidence_readiness_defined`、`operator_approval_runtime_implementation_entry_review_defined`、`audit_store_handoff_readiness_defined`、`audit_store_contract_event_schema_readiness_defined`、`audit_store_ownership_boundary_readiness_defined`、`audit_store_delivery_idempotency_runtime_boundary_readiness_defined`、`audit_store_runtime_implementation_entry_refresh_v3_defined`、`resolver_backend_health_boundary_readiness_defined` 和 `resolver_backend_health_runtime_implementation_entry_review_defined`。
- `Workflow Saved Draft v1 Implementation` 与 consumer integration 已完成，后续 [Workflow Executor v0](workflow/workflow-executor-v0.md) 也已按独立功能设计和任务卡完成；confirmation、writeback 与 replay 仍不得随 executor v0 隐式打开。
- 细专题已拆到 [Workflow 细专题入口](workflow/README.md)：`Saved Workflow Draft v1`、Draft Designer、User Workspace saved draft、Review Handoff 和 durable store 迁移前置专题分别承接功能、页面、用户流和后续实现准入。

## 设计边界

- draft、validation、plan、readiness、review 和 execution 是不同阶段，不能在 UI 中混成一个“已可执行”状态。
- 功能文档不再把旧停止线作为全局禁止清单复述；每个功能应说明本轮允许打开的能力、依赖的验证证据，以及需要留到后续独立目标的能力。
- 高风险 tool/action 默认要求 confirmation。
- executor v0 已具备 run record、audit ref、failure taxonomy 和 no side effects 证据；HTTP Tool 与 confirmation 的 workflow-specific durable audit、原子 claim、run v2 和高风险执行策略已经由独立专题固定，首版明确不创建 materialized result store。writeback 或 replay 仍需未来独立功能设计。
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
- 当前实现已注册 dev-only HTTP route，并接入 web consumer 状态区分；store selector 已提供 fail-closed mode selection，`memory_dev` 使用内存 dev store，`repository_disabled` / `repository` 返回 `repository_store_disabled`，unknown mode 返回 `invalid_draft_store_mode`；repository adapter、adapter smoke execution、production auth runtime bridge、repository mode enablement、schema migration runner readiness、runner implementation entry review、database connection / schema marker preconditions、connection provider entry review / entry refresh v2、secret resolver readiness、secret resolver implementation entry review、repository mode runtime boundary review、schema marker / migration runner readiness refresh、schema marker contract implementation entry review、manual migration runner implementation entry refresh、schema marker runtime dependency refresh、secret resolver runtime dependency refresh，以及 production secret backend 从 operator runbook / rotation policy 到 audit store runtime implementation entry refresh v4、operator approval refresh、credential handle refresh、real resolver runtime entry refresh、no leakage smoke refresh、backend health refresh 和 production resolver runtime blocker consolidation 的静态证据已落地；仍不创建 public production API，不解析真实 secret，不接真实数据库、OIDC token validation、membership adapter、repository mode runtime、production secret audit store、backend health runtime、production resolver runtime、no secret leakage smoke runtime、credential handle runtime、approval runtime、schema marker runtime、manual runner 或可执行 schema migration。

### Durable Store 接入说明

当前开发 / 测试态 durable store 已形成三种同构实现，production persistence 仍关闭：

- `memory_dev` 是快速进程内模式；聚合 `sqlite_dev` 使用共享 runtime 与独立 Workflow migration；显式 `postgres_dev_test` 使用独立 migration、运行角色、schema marker、连接池和 repository query executor。
- 三种模式复用同一 `SavedWorkflowDraftRepository` 领域契约、完整 tenant / workspace / application / owner scope、预期版本 CAS、sanitized projection、稳定排序和 no-fallback 失败语义。
- SQLite 与 PostgreSQL 已覆盖迁移、重启恢复、并发单写者、作用域隔离、marker mismatch、连接失败和敏感材料禁入；store selector 或 schema 不满足时不得回退 memory / sample。
- production `repository`、production database resource、production secret resolver、真实 Radish membership 与公开生产 API 仍 fail closed。历史 Production Secret Backend / Storage Adapter 链不能反向否定已完成的开发测试态存储，也不能被解释为生产存储已就绪。

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

Run History、Failure Review、Run Comparison、Evaluation Cases、Baseline / Case Versioning、Evaluation Suite / Release Review、Model Gateway Request History、Gateway Playground、API 密钥开发测试态认证与[受控 HTTP Tool 与人工确认专题](workflow/controlled-http-tool-human-confirmation-dev-test-v1.md)均已完成。下一产品动作固定为设计先行：

1. 优先建立 RAG retrieval 的开发 / 测试态短功能设计，明确知识源所有权、租户 / workspace / application 隔离、索引与更新边界、检索预算、引用、失败语义和数据泄漏停止线。
2. 设计评审通过后才拆实现批次；不得把现有 `rag_ref`、离线占位或未来 adapter 解释为真实 retrieval runtime。
3. Builder、普通只读审查和历史 Production Secret Backend / Storage Adapter readiness 链不占用下一产品顺位；writeback、replay / resume 与 agent loop 不随 RAG 设计自动打开。

## 验收方式

- 本次功能设计文档更新：确认受控 HTTP Tool 专题已固定兼容策略、状态机、allow / deny、SSRF / credential、claim / crash、run v2、失败码、三种 store、Web 验收和停止线；运行快速与全量仓库检查。
- Saved Workflow Draft v1 domain service 实现批次：schema 或等价类型定义、保存 / 读取 / 校验单元测试、negative tests、version conflict tests、storage boundary checks、no sample fallback tests、no side effects tests 和 `./scripts/check-repo.sh --fast`。
- Saved Workflow Draft v1 consumer integration 批次：按选择补 dev-only route / web consumer / consumer contract / fixture / checker，至少运行 Go 单元测试、web build、workflow 聚合 checker 和 `./scripts/check-repo.sh --fast`。
- 若实现批次改变阶段口径、治理入口、schema 真相源或高风险执行边界，补跑全量 `./scripts/check-repo.sh`。
- confirmation 已由受控 HTTP Tool 专题完成设计；实现涉及新 API、schema 和高风险执行边界，必须由单一任务卡承接并补 scope、audit、SSRF、CAS / crash、side-effect 与全量仓库验证。writeback 与 replay 仍需未来独立设计。

## 停止线

- [Workflow Executor v0](workflow/workflow-executor-v0.md) 继续只允许开发 / 测试态 `prompt`、`llm`、`condition`、`output` 执行；HTTP Tool 与 confirmation 只能进入新的版本化 action plan / run v2 路径。RAG、code、sandbox、agent loop、publish、business writeback、run replay、run resume 和 materialized result reader 继续关闭。
- 不把 Saved Workflow Draft v1 已完成的 memory / SQLite / PostgreSQL 开发测试态存储扩张成 production repository、production database resource、production secret、Radish membership 或公开生产 API；API key lifecycle、quota、billing 和 cost ledger 也不随 Workflow 工具批次打开。
- 不把 saved draft 的 `valid_for_review`、validation summary、risk summary 或 readiness summary 解释为 production ready、publish ready 或 executable ready。
- 不继续新增同层只读 evidence 面板来替代 Saved Workflow Draft v1；普通只读整理只复用现有 workflow gate、web build 和 fast baseline。
- 不把 `RadishFlow` / `Radish` 上层挂载点缺失写成本功能阻塞；同时也不声明真实接入 ready。
