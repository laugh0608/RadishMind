# RadishMind 产品范围与目标

更新时间：2026-06-21

## 核心定义

`RadishMind` 是 `Radish` 体系下的 AI 工具、工作流、模型网关和 Copilot 集成平台。

更强的项目定义、三层定位、northbound/southbound 战略、service mode 分级与 action safety ladder 见 [战略定义](radishmind-strategy.md)。

它不是上层业务真相源，不是单一大模型仓库，也不只是等待真实接入的中转站。它的职责是把用户端 AI 应用、管理端治理、模型 API 分发、工作流运行、上层项目上下文、局部规则、工具能力、模型推理和审计治理收口为同一条可复跑的运行链路。

当前核心边界：

- 读状态、读文档、读附件和可选图像，输出解释、诊断、结构化建议和候选动作。
- 高风险动作必须保留 `requires_confirmation`，由人工确认或上层规则层复核后再执行。
- 对内保持统一 canonical protocol，对外兼容常见模型调用协议和常见 AI 服务协议，而不是被单一厂商接口绑死。
- 模型负责理解、推理、归纳、排序和建议生成；runtime、adapter、tooling、rule validation 和 audit 负责上下文打包、工具调用、结构校验、权限边界和可追溯性。
- `RadishMind-Core` 是基座适配型自研主模型路线，不是从零预训练基础大模型。
- 图片像素生成不并入主模型职责，默认由 `RadishMind-Image Adapter` 与独立 backend 承接。
- 部署方式、数据库选型、登录 / 授权边界优先参考 `Radish`；未来 RadishMind 作为 OIDC client 接入 `Radish`，不自建第二套身份真相源。参考 `Radish` 不代表默认引入 `.NET` / ASP.NET Core；RadishMind 后端默认继续使用 `Go` 承载 control plane / gateway / API 服务，`Python` 只保留在模型、评测和 AI 生态强相关链路，`TypeScript/Vite` 承载前端。
- `RadishFlow` 和 `Radish` 是优先接入对象与产品参考，但不是 RadishMind 平台本体开发的阻塞条件。上层暂时没有稳定 UI、command 或 API 挂载点时，本仓库应继续推进可离线验证、可复用到后续真实接入的用户端、workflow runtime、control plane 和模型网关功能；不把等待上层接线写成产品停滞理由。

## 产品形态

长期产品形态按四个一级面组织：

2026-06-14 起，产品形态的长期设计默认沉淀到 [功能设计文档](features/README.md)。`docs/task-cards/` 不再作为产品功能的默认主文档，只用于具体实现批次、前置条件或高风险边界；普通只读展示、文案、布局和 evidence 组织优先复用现有聚合门禁、web build、consumer smoke 和仓库基线。新增 API、执行边界、生产声明、数据格式、外部 provider 风险、schema 变化或高风险能力时，才新增专项 fixture / checker。

2026-05-27 已新增 [Control Plane / User Workspace / Workflow v1 计划](task-cards/control-plane-user-workspace-workflow-v1-plan.md)，用于固定四个产品面的 v1 服务边界、数据边界和停止线；`product-surface-v1-boundary` 已进一步把 `User Workspace`、`Admin Control Plane`、`Model Gateway / API Distribution` 和 `Workflow / Agent Runtime` 的资源、读模型和写边界写入可检查 fixture；`control-plane-data-boundary` 已固定 tenant、user、role、permission、provider profile、model route、quota、price、audit、secret ref 与 deployment status 的 ownership；`radish-oidc-client-preconditions` 已固定 issuer、client、claim mapping、tenant binding、logout、audit 和 failure taxonomy；`gateway-api-key-quota-readiness` 已固定 API key、quota、rate limit、cost ledger 和 trace 前置条件；`workflow-definition-run-record-boundary` 已固定 workflow definition、run record、状态流转、失败分类、审计证据和停止线；`control-plane-read-model-v1` 已固定 tenant summary、application summary、API key summary、quota summary、workflow definition summary、run record summary 和 audit summary 的只读 read model；`control-plane-read-route-contract-v1` 已固定 `GET /v1/user-workspace/runs` 等七类 tenant-scoped read-only route contract；`control-plane-read-response-fixtures-v1` 已固定 response fixture、统一 envelope、`failure_code` 和脱敏输出；`control-plane-read-negative-contract-v1` 已固定负向契约、forbidden method / query / fallback、敏感字段投影拒绝和 fail-closed 输出；`control-plane-read-fake-store-handler-implementation-v1` 已实现七条 fake-store-backed read route，覆盖 tenant summary、applications、api-keys、quota summary、workflow definitions、runs 与 audit；`control-plane-read-auth-db-preconditions-v1` 已固定真实 auth/db 前置条件、future auth middleware 和 future read store repository；`control-plane-read-consumer-contract-v1` 已固定 TypeScript consumer contract 和离线消费 smoke；`control-plane-read-formal-ui-boundary-v1` 已固定正式 UI 边界、页面到 read route 的分配、只读状态和敏感字段停止线。它不代表正式用户端、生产管理端、workflow executor、API key / quota、数据库 read path 或 Radish OIDC 已实现。

2026-05-28 已新增 `control-plane-read-formal-ui-implementation-readiness-v1`，固定未来 `apps/radishmind-web/` 预留落点、`apps/radishmind-console/` app 边界、页面实现顺序、consumer contract 复用、测试策略和停止线。

2026-05-31 已创建 `apps/radishmind-web/`，作为正式产品 UI 的 read-only product shell 首个实现落点。当前 shell 默认只消费 `contracts/typescript/control-plane-read-api.ts` 中的离线 view model，已包含 route catalog、共享状态组件、forbidden output guard、只读 `admin-tenant-overview`、`admin-audit-log`、`workspace-applications`、`workspace-api-keys`、`workspace-usage-quota`、`workspace-workflow-definitions` 与 `workspace-run-history` 页面切片。2026-06-01 已补显式 opt-in 的 dev-only live read consumer：只有设置 `VITE_RADISHMIND_READ_SOURCE=dev-live-http`，且后端设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 时，页面才通过 HTTP 消费 fake-store-backed read handlers 和测试身份上下文。2026-06-09 已把 RadishFlow Copilot 与 Radish Docs Assistant 两组只读产品样例收敛到 response fixture、Go fake store、consumer smoke 和前端离线默认数据的一致性校验，并把 User Workspace Home、Workflow Review Workspace、Workflow Review Handoff 和 workflow context selection 组织成普通离线审查流。2026-06-10 已新增普通离线 Model Gateway Overview、Route Evidence、Usage/Audit Evidence 与 Evidence Review / Readiness，复用 shared read shell、API key、quota、run history、audit、provider runtime、gateway readiness 和前三个网关 view model 证据，展示 northbound API surfaces、provider/profile inventory、route binding、selection cases、key scope、quota / cost snapshot、trace / failure、audit decision、readiness rollup、evidence checklist、route / usage / audit risks 与 locked distribution capabilities；同日新增 Admin Operations Review / Readiness，复用 tenant overview、audit log、Model Gateway Evidence Review 和 Production Ops 静态证据，展示管理端 readiness、evidence checklist、operational risks 和 boundary locks。2026-06-13 新增 Admin Provider/Profile & Deployment Evidence Review / Readiness，继续复用 Model Gateway route / review、Admin Operations、tenant overview 和 audit log，展示 provider/profile readiness、model route readiness、secret / deployment evidence、operator risks 和 locked capabilities。2026-06-16 workflow 用户端已补 saved dev draft list / restore、Draft Designer 本地结构编辑、节点属性编辑和 active draft review handoff；2026-06-17 已补 saved draft repository adapter implementation plan、schema artifact manifest contract、adapter smoke readiness、selector implementation、静态 schema artifact materialization、production auth readiness 和 repository adapter implementation entry review；2026-06-18 已补 repository adapter implementation、adapter smoke execution、production auth runtime bridge、`workflow-saved-draft-repository-mode-enablement-v1` 和 `workflow-saved-draft-schema-migration-runner-readiness-v1`，分别固定 `draft_repository_mode_enablement_review_defined` 与 `draft_schema_migration_runner_readiness_defined`；2026-06-19 已补 `workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1`、`workflow-saved-draft-database-connection-schema-marker-preconditions-v1`、`workflow-saved-draft-database-connection-provider-implementation-entry-review-v1`、`workflow-saved-draft-database-secret-resolver-readiness-v1`、`workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1`、`production-secret-backend-config-secret-ref-readiness-v1`、`production-secret-backend-provider-profile-secret-binding-readiness-v1`、`production-secret-backend-secret-resolver-interface-disabled-readiness-v1`、`production-secret-backend-operator-runbook-negative-gates-readiness-v1`、`production-secret-backend-rotation-audit-policy-readiness-v1`、`production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1`、`production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1`、`production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1` 和 `production-secret-backend-fake-resolver-implementation-v1`，固定 `draft_schema_migration_runner_implementation_entry_review_defined`、`draft_database_connection_schema_marker_preconditions_defined`、`draft_database_connection_provider_implementation_entry_review_defined`、`draft_database_secret_resolver_readiness_defined`、`draft_database_secret_resolver_implementation_entry_review_defined`、`config_secret_ref_readiness_defined`、`provider_profile_secret_binding_readiness_defined`、`secret_resolver_interface_disabled_readiness_defined`、`operator_runbook_negative_gates_readiness_defined`、`rotation_audit_policy_readiness_defined`、`test_fixture_strategy_fake_resolver_entry_review_defined`、`fake_resolver_contract_no_secret_leakage_smoke_strategy_defined`、`fake_resolver_implementation_task_card_entry_readiness_review_defined` 与 `fake_resolver_implementation_task_card_defined`；2026-06-20 已补 `production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1`、`production-secret-backend-fake-resolver-runtime-implementation-v1`、`production-secret-backend-real-resolver-runtime-preconditions-v1`、`production-secret-backend-real-resolver-runtime-implementation-entry-review-v1`、`production-secret-backend-resolver-backend-profile-selection-readiness-v1` 和 `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1`，固定 `fake_resolver_runtime_implementation_entry_review_defined`、`fake_resolver_runtime_test_only_implemented`、`real_resolver_runtime_preconditions_defined`、`real_resolver_runtime_implementation_entry_review_defined`、`resolver_backend_profile_selection_readiness_defined` 与 `real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`。这些证据确认 runner、connection provider、production resolver runtime、no secret leakage smoke runtime、operator gate runtime、rotation runtime、audit store 和 production secret backend resolver 当前不打开，disabled resolver interface 只返回 fail-closed 脱敏状态，operator runbook / negative gates 只固定人工启用、脱敏验证和 smoke record reference 前置，rotation / audit policy 只固定策略、审计字段和 rollback 前置，test-only fake resolver runtime 只服务离线测试，真实 resolver runtime implementation entry review 结论为 blocked before runtime task card，resolver backend profile selection 与 no leakage smoke runtime strategy 只形成静态前置证据。这些仍属于 dev-only / local-only / advisory-only 工作流设计与 durable store / Production Ops 前置证据，不接生产后端、不接数据库、OIDC token validation、membership adapter、repository mode runtime、migration runner、production API、API key lifecycle、quota enforcement、rate limit、billing、production secret resolver implementation、production resolver runtime、retry/fallback execution、deployment preflight、workflow executor、confirmation、writeback 或 replay；`apps/radishmind-console/` 仍只是本地 ops surface。

2026-06-20 日终补充：同日后续已补 `production-secret-backend-credential-handle-runtime-boundary-readiness-v1`、`production-secret-backend-operator-approval-runtime-evidence-readiness-v1`、`production-secret-backend-audit-store-handoff-readiness-v1`、`production-secret-backend-resolver-backend-health-boundary-readiness-v1` 和 `production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1`，固定 `credential_handle_runtime_boundary_readiness_defined`、`operator_approval_runtime_evidence_readiness_defined`、`audit_store_handoff_readiness_defined`、`resolver_backend_health_boundary_readiness_defined` 和 `resolver_backend_health_runtime_implementation_entry_review_defined`。这些仍只属于 Production Ops 前置证据：不创建 credential handle runtime、operator approval runtime、audit store / writer / event、backend health runtime task card、backend health runtime、backend health check、production resolver runtime、cloud secret service、DB provider、repository mode 或 public production API。

2026-06-21 已补 `workflow-saved-draft-repository-mode-runtime-boundary-review-v1`，固定 `draft_repository_mode_runtime_boundary_review_defined`。该评审消费 repository adapter、adapter smoke、production auth runtime bridge、repository mode enablement、runner / connection / resolver entry review、audit store runtime entry refresh v3 和 production secret backend implementation readiness 证据，结论仍是 repository mode runtime task card blocked；不启用 repository store mode，不创建真实 query executor、schema marker runtime、OIDC / membership、production API、audit store runtime、executor、confirmation、writeback 或 replay。

2026-06-21 继续补 `workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1`，固定 `draft_schema_marker_migration_runner_readiness_refresh_defined`。该 refresh 消费 runner readiness / entry review、database connection / schema marker preconditions、connection provider / secret resolver entry review、repository mode runtime boundary review、schema artifact、repository adapter、adapter smoke 和 production auth runtime bridge 证据，只收束 applied marker、manual runner、dry-run、idempotency / lock、duplicate handling 和 rollback observability；不创建 schema marker implementation task card、migration runner implementation task card、SQL、schema version table、marker runtime、runner、DB provider、repository mode runtime、production API、executor、confirmation、writeback 或 replay。

2026-06-22 已补 `workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1`，固定 `draft_schema_marker_contract_implementation_entry_review_defined`。该 review 消费 schema marker / migration runner readiness refresh、runner implementation entry review、database connection / schema marker preconditions、connection provider / secret resolver entry review 和 repository mode runtime boundary review，确认 marker reader / writer contract implementation task card 当前仍为 blocked；不创建 schema marker implementation task card、marker runtime、schema version table、SQL、runner、DB provider、repository mode runtime、production API、executor、confirmation、writeback 或 replay。

2026-06-22 继续补 `workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1`，固定 `draft_manual_migration_runner_implementation_entry_refresh_defined`。该 refresh 消费 schema marker contract entry review、schema marker / migration runner readiness refresh、runner implementation entry review、静态 schema artifact、connection provider / secret resolver entry review 和 repository mode runtime boundary review，确认 manual migration runner implementation task card 当前仍为 blocked；不创建 runner runtime、runner command、dry-run output、SQL、schema marker runtime、DB provider、repository mode runtime、production API、executor、confirmation、writeback 或 replay。

当前产品 UI 的门禁策略已经从普通展示页逐项专项证明，调整为能力边界与聚合门禁优先。`control-plane-read-formal-ui-readiness-close-v1` 已用 surface matrix 聚合固定七个页面的 route binding、状态预览、request / audit ref 和 forbidden output guard；`control-plane-read-auth-store-transition-preconditions-v1` 已固定从 dev fake auth / fixture-backed fake store 迁移到未来 auth middleware / read store repository 前必须满足的 gates。2026-06-14 起，普通展示、文案、布局和 evidence 组织默认复用聚合门禁，不再逐项新增 task card / fixture / checker；新增 API、执行边界、生产声明、数据格式、外部 provider 风险或高风险能力时才新增专项 gate。上述内容都不能解释为真实数据库、Radish OIDC、production API consumer、API key / quota、repository adapter 或 workflow executor ready。

read store 的产品范围现在已经从“继续固定未来迁移契约”推进到“先做一层可替换的后端接口，再实现真实持久化”。`Control Plane Durable Read Foundation v1` 已定义 `ControlPlaneReadRepository` interface，并让现有七条 fake-store-backed read handlers 通过 interface 消费数据；`control-plane-read-repository-contract-smoke-v1`、`control-plane-read-repository-implementation-readiness-v1`、`control-plane-read-store-selection-readiness-v1` 和 `control-plane-read-schema-migration-readiness-v1` 继续说明未来七条 read route 如何从 fake store 迁移到 repository/database。它不把 read-side 页面升级成 production API consumer，也不实现 SQL、migration、repository adapter、真实数据库、Radish OIDC、token validation、API key lifecycle、quota enforcement 或 workflow executor。

1. `User Workspace`

- 面向终端用户和项目成员。
- 支持创建 AI 应用、Prompt 应用、Workflow、Agent / Copilot 应用、RAG 或知识问答应用。
- 用户可以管理自己的应用、API key、调用量、运行记录和成本摘要。
- 当前 `apps/radishmind-web/` 只提供 read-side 页面切片：applications、API keys、usage quota、workflow definitions 和 run history 默认都是离线只读展示；dev-only live read path 也只能读取 fake-store-backed handler，不提供创建、编辑、执行、replay 或写回控件。
- `Saved Workflow Draft v1` 已具备 platform Go domain service、内存 dev store、dev-only HTTP route、web consumer、save / read / validate / list 契约、版本冲突、失败语义、sanitized response、no sample fallback 测试、Draft Designer 受控本地编辑、本地结构编辑、节点属性编辑、User Workspace 本地草案创建入口、saved dev draft list / restore 和 active draft review handoff；repository contract smoke、runner readiness、static runner implementation、repository adapter implementation plan、schema artifact manifest contract、adapter smoke readiness、store selector implementation、schema artifact materialization、production auth readiness、repository adapter implementation entry review、repository adapter implementation、adapter smoke execution、production auth runtime bridge、repository mode enablement 准入评审、schema migration runner readiness、schema migration runner implementation entry review、database connection / schema marker preconditions、connection provider implementation entry review、database secret resolver readiness、database secret resolver implementation entry review、production secret backend config / secret ref readiness、provider profile secret binding、disabled resolver interface、operator runbook / negative gates、production-secret-backend-rotation-audit-policy-readiness-v1、production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1、production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1、production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1、production-secret-backend-fake-resolver-implementation-v1、production-secret-backend-fake-resolver-runtime-implementation-v1、production-secret-backend-real-resolver-runtime-preconditions-v1、production-secret-backend-real-resolver-runtime-implementation-entry-review-v1、production-secret-backend-resolver-backend-profile-selection-readiness-v1 和 production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1 已作为 durable store / Production Ops 前置证据落地，状态包含 `fake_resolver_runtime_test_only_implemented`、`real_resolver_runtime_preconditions_defined`、`real_resolver_runtime_implementation_entry_review_defined`、`resolver_backend_profile_selection_readiness_defined` 和 `real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`。它还没有 durable persistence、repository mode runtime、真实数据库、SQL migration runner、production secret resolver implementation、backend runtime、production resolver runtime、no secret leakage smoke runtime、rotation runtime、production secret audit store、OIDC token validation、membership adapter、production API、publish、run 或 executor。
- 后续同日新增的 credential handle boundary、operator approval evidence、audit store handoff、backend health boundary 和 backend health runtime entry review 只补齐真实 resolver runtime 之前的静态前置证据；它们不创建 handle runtime、approval runtime、audit store runtime、backend health runtime、health check、production resolver runtime、repository mode runtime、production API、publish、run 或 executor。
- 工作流方向参考 `Dify` 的应用构建与 workflow 编排，但首版只实现 Radish 体系当前需要的可治理切片，不追求一次性复刻全量能力。

2. `Admin Control Plane`

- 面向平台管理员和运维。
- 管理租户、用户、角色、权限、模型供应商、provider profile、模型路由、API key、额度、价格、审计、secret backend 和部署状态。
- 认证、授权、数据库、部署和运维习惯优先对齐 `Radish`；未来通过 OIDC 接入 `Radish` Auth。
- Control Plane 可以拆成独立 Go 服务，但不因为职责扩张而引入新后端语言或塞进 gateway 单体。
- 当前 `apps/radishmind-web/` 只提供只读 `admin-tenant-overview`、`admin-audit-log`、普通离线 Admin Operations Review / Readiness 和 Admin Provider/Profile & Deployment Evidence Review / Readiness；它不是 production admin console，也不提供 tenant mutation、audit mutation、provider/profile mutation、model route change、raw payload export、durable audit store、production backend、secret resolver、deployment preflight 或生产管理操作。

3. `Model Gateway / API Distribution`

- 面向 API 调用者和上层服务。
- 提供 OpenAI-compatible / Responses / Messages / Models 等 northbound API，统一分发到多 provider、多 profile 和多模型。
- 支持后续的 API key 分发、配额、限流、成本统计、trace、fallback / load balancing 和 provider health。
- 当前 `apps/radishmind-web/` 已提供普通离线 Model Gateway Overview、Route Evidence、Usage/Audit Evidence 与 Evidence Review / Readiness，只解释 provider/profile、route、key / quota、trace、audit、risk 和 locked capability 证据，不执行真实分发能力。
- 模型 API 分发方向参考 `sub2api` 与 `axonhub`，但必须保留本仓库的 provider registry、审计和生产停止线。

4. `Workflow / Agent Runtime`

- 面向 AI 应用执行。
- 承载 Prompt、LLM、HTTP tool、RAG retrieval、condition、output、后续受控 code / sandbox 与 agent loop。
- 每次运行都应有 trace、输入输出摘要、成本、错误分类和风险边界。
- `workflow-definition-run-record-boundary` 只把 workflow definition、run record、node execution、tool audit、result materialization、confirmation decision、状态流转、失败分类、审计证据和停止线固定为治理证据，不代表 executor、confirmation、writeback 或 replay 已实现。
- `Workflow / Agent Runtime Function Surface v1` 已把现阶段可推进功能面限定为 application detail、workflow definition detail、run detail、tool action preview、confirmation placeholder、Draft Designer、offline validation inspector、execution plan preview、runtime readiness inspector、surface overview、context selection、scenario inspector、Workflow Review Workspace、User Workspace Home 和 Workflow Review Handoff 的只读 / blocked / local-only surface，优先走 fixture 或 fake-store dev path。它们只展示 selected context、draft、validation、plan、readiness、scenario、blocked capability、stop line、route / request / audit metadata、review rollup 和 human review handoff；Draft Designer 可做受控本地编辑、本地结构编辑、节点属性编辑，并在显式 dev-only 配置下保存 / 读取 / 校验 / 列出 memory dev draft，User Workspace Home 可恢复 saved dev draft，Review Handoff 可显示 active draft review record，但不提供 durable draft persistence、validation / execution plan / readiness / scenario / review / handoff persistence、publish、executor、confirmation decision、writeback 或 replay。
- 上层挂载点未成熟时，workflow 产品面继续先做离线草案设计、结构检查、execution plan preview、readiness 展示、场景解释、review workspace 和 blocked capability 说明；这些产品能力应复用未来真实接入所需的 canonical contract 和停止线，而不是等待 `RadishFlow` 或 `Radish` 提供承接入口后才开始。
- 高风险 tool/action 默认 `requires_confirmation`，不得直接写上层业务真相源。

## 项目范围

### 1. `Runtime Service / Model Gateway`

- 提供最小可运行的推理入口、gateway、route 识别、provider/profile 选择、响应封装和本地产品 discovery 面。
- 当前已有 CLI runtime、进程内 Python gateway、最小 `Go` HTTP bridge、本地启动 runbook、`GET /v1/platform/overview` 只读产品 overview、`GET /v1/platform/local-smoke` 本地 readiness 摘要、session/tooling metadata shell、blocked action shell、`apps/radishmind-console/` 本地 console 壳、Dev Diagnostics、`Local Readiness` 面板、Provider/Profile Details、Stop-line Details、overview / local-smoke failure surface、console behavior / visual smoke record / dev entry / production boundary gate、P3 checklist、Docker local compose、测试 / 生产共用部署态 compose、镜像命名治理、deployment readiness 静态 smoke、container smoke runbook、运行记录模板、一次 `docker_local` container smoke 运行记录、provider capability matrix、provider health smoke、provider selection policy、provider runtime docs refresh、production-secret-backend-config-secret-ref-readiness-v1、`config_secret_ref_readiness_defined`、production-secret-backend-provider-profile-secret-binding-readiness-v1、`provider_profile_secret_binding_readiness_defined`、production-secret-backend-secret-resolver-interface-disabled-readiness-v1、`secret_resolver_interface_disabled_readiness_defined`、production-secret-backend-operator-runbook-negative-gates-readiness-v1、`operator_runbook_negative_gates_readiness_defined`、production-secret-backend-rotation-audit-policy-readiness-v1、`rotation_audit_policy_readiness_defined`、production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1、`test_fixture_strategy_fake_resolver_entry_review_defined`、production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1、`fake_resolver_contract_no_secret_leakage_smoke_strategy_defined`、production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1、`fake_resolver_implementation_task_card_entry_readiness_review_defined`、production-secret-backend-fake-resolver-implementation-v1、`fake_resolver_implementation_task_card_defined`、production-secret-backend-fake-resolver-runtime-implementation-v1、`fake_resolver_runtime_test_only_implemented`、production-secret-backend-real-resolver-runtime-preconditions-v1、`real_resolver_runtime_preconditions_defined`、production-secret-backend-real-resolver-runtime-implementation-entry-review-v1、`real_resolver_runtime_implementation_entry_review_defined`、production-secret-backend-resolver-backend-profile-selection-readiness-v1、`resolver_backend_profile_selection_readiness_defined`、production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1 和 `real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`；本地只读产品壳已达到 `local usable / read-only close`，Docker 静态部署边界已可检查，本地 mock 容器 smoke 已跑通，Provider Runtime & Health v1 进入 close candidate。真实镜像发布 workflow、production secret backend、secret resolver implementation、production resolver runtime、no secret leakage smoke runtime、optional live health、真实 retry/fallback、测试环境 smoke、生产前复核和生产部署边界只在明确运行窗口或独立任务卡下推进，服务层与 `gateway` 不锁死在单一语言上，可按职责采用 `Go`。
- Production Ops 后续已补 `credential_handle_runtime_boundary_readiness_defined`、`operator_approval_runtime_evidence_readiness_defined`、`audit_store_handoff_readiness_defined`、`resolver_backend_health_boundary_readiness_defined` 和 `resolver_backend_health_runtime_implementation_entry_review_defined`；这些不改变 runtime 缺口，真实 production secret backend、production resolver runtime task card、production resolver runtime、no leakage smoke runtime、credential handle runtime、approval runtime、audit store runtime、backend health runtime 和 health check 仍未创建。
- 这一层必须同时覆盖两条方向：
  - 北向协议兼容：对外提供 native Copilot API，以及 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models`、`/v1/platform/overview`、`/v1/platform/local-smoke`、session/tooling metadata shell 这类常见兼容接口和只读产品发现接口。
  - 南向模型接入：对内接入 `RadishMind-Core`、`local_transformers / HuggingFace`、`Ollama`、OpenAI-compatible、Gemini native、Anthropic messages 等 provider / transport。

### 2. `Conversation & Session`

- 管理 `conversation_id`、会话上下文、历史压缩、恢复和审计边界。
- 当前已有 session contract、history/state policy、checkpoint metadata-only read route 和 `GET /v1/session/metadata`；这些只支撑 metadata / overview 展示，不代表 durable session store、长期记忆或 replay executor 已启用。

### 3. `Tooling Framework`

- 承载检索、附件解析、项目语义转换、本地候选生成、response builder 与工具策略。
- 当前已有通用 tool contract、registry、audit、`GET /v1/tools/metadata` 和 blocked `POST /v1/tools/actions`；它们只支撑 contract-only 展示和 blocked action response，不代表真实工具执行器、confirmation flow 或业务写回已启用。

### 4. `Evaluation & Governance`

- 负责 schema、smoke、offline eval、candidate record、review、promotion gate 和仓库级检查。
- 目标不是只看机器指标，而是让 advisory-only、confirmation、route、citation 和 review 边界能长期复跑。

### 5. `Model Adaptation`

- 负责基座选型、prompt/runtime 协同、蒸馏、训练样本治理、模型晋级与回归。
- 当前已形成 P4 v1 前置证据：1.5B raw 在 docs / ghost 上可跑但 edits blocked，repaired comparison 只作后处理证据，3B CPU 单样本 timeout。真实模型产出、3B/4B 长跑、训练 JSONL、蒸馏和权重相关工作转入后置专题。
- 自研模型只是平台的一类 provider，不应和 `HuggingFace`、`Ollama`、OpenAI-compatible 或其它外部模型接入能力互相替代。

### 6. `Image Path`

- 主模型只输出结构化 image intent、约束、审查和 artifact metadata。
- 真正的图片生成由独立 image adapter 和 backend 承接。
- 当前 `services/runtime/image_artifact_runtime_mapper.py`、`services/runtime/image_artifact_response_consumer.py` 与 `services/runtime/inference_response.py#coerce_response_document` 已形成 metadata-only response builder 链路：从 `copilot_request.artifacts[*].metadata.image_generation_artifact` 读取 metadata，投影为 artifact citation，并合并进现有 `CopilotResponse.citations`。它们不读取 artifact 二进制、不查 artifact store、不解析 public URL、不调用真实生图 backend、不上传 artifact、不修改 `CopilotResponse` schema。
- `blocked / failed / pending_review` artifact、invalid metadata、hash / mime / dimensions mismatch、public URL claim、signed URL policy missing、binary payload、provider raw dump、store / reader 缺失、safety review not passed 和 provenance missing 都必须 fail closed，不能进入成功 response。

### 7. 用户端、管理端和上层项目接入面

- 用户端用于 AI 工作流、应用、模型 API key、调用量和运行记录。
- 管理端用于 provider/profile、模型路由、租户、权限、quota、secret、审计和部署状态。
- 当前 `apps/radishmind-console/` 只是本地 ops surface 和只读产品壳，不等同于正式用户端或生产管理端。
- 当前 `apps/radishmind-web/` 是正式产品 UI 的 read-side shell，默认离线，可显式 opt-in 到 dev-only live read；它不等同于完整用户端、production admin console 或真实 API consumer。

- `RadishFlow`、`Radish`、`RadishCatalyst` 是应用面，不是项目本体的全部意义。
- 这些接入面复用同一套 runtime、contract、tooling、evaluation 和 governance，而不是各自私接模型。
- 这些接入面若暂时无法提供真实挂载点，不应拖慢 RadishMind 的平台功能建设；本仓库应先把成熟的离线产品面、协议、风险边界和验证基线做好，等上层条件满足后再选择一个切片真实接入。

## 当前阶段判断

- 历史上的 `M3` service/API smoke 与 `M4` broader review、`3B/4B` capacity review 已经收口为冻结证据。
- 当前正式主线切换为“AI 工具 / 工作流 / 模型网关 / Copilot 集成平台重定义 + 平台基础能力建设”，不再把“继续深挖同一批实验”或“提前设计不存在的真实接线”当作默认推进方式。
- 当前 `P3 Local Product Shell / Ops Surface` 的本地只读产品壳已收口为 `local usable / read-only close`：已用 `/v1/platform/overview`、`/v1/platform/local-smoke`、overview / local-smoke consumer smoke、最小本地 console 壳、Dev Diagnostics、`Local Readiness` 面板、Provider/Profile Details、Stop-line Details、overview / local-smoke failure surface、console behavior / visual smoke record / dev entry / production boundary gate 和 P3 checklist 固定本地 console 可展示能力与未满足的生产前置条件。`Production Ops Hardening v1` 已进一步固定 Docker local/test/prod 部署形态、compose 边界、镜像命名、静态 smoke、runbook 和运行记录模板，并完成一次 `docker_local` container smoke；`Provider Runtime & Health v1` 已固定 capability / health / selection / docs 四个可检查切片并进入 close candidate。2026-06-14 阶段评估后，默认停止继续扩同层只读 UI / gate-only 切片；Image Path 的 metadata-only response builder 接线和 Control Plane Read 的 repository interface + fake store interface 化均已完成，后续产品范围按功能设计文档选择单一实现方向，不在无运行窗口时继续补 console 小切片、provider 同层小切片、Production Ops 静态治理、真实模型长跑或假想上层接线。
- 训练 / 蒸馏样本继续只提交 manifest、summary、复核策略和实验说明；生成的 JSONL 和真实模型产物默认留在 `tmp/`。

## 当前优先支持的应用面

### `RadishFlow`

首批能力继续围绕结构化流程状态：

- `FlowsheetDocument + SelectionState + DiagnosticSummary` 的解释与诊断问答。
- 基于选中对象的候选编辑提案。
- 基于本地合法候选集的 ghost port、ghost connection、ghost stream name 补全建议。
- 求解状态、控制面、entitlement、package sync 与离线授权摘要解释。

当前不可侵入：

- 数值求解热路径。
- CAPE-OPEN / COM 适配边界。
- token、credential、auth cache 和包体完整性真相。
- 未经确认直接改写 `FlowsheetDocument`。

### `Radish`

首批能力继续围绕知识与内容：

- 固定文档、在线文档、论坛内容和 Console 权限知识的检索增强问答。
- 文档、帖子、评论和附件的摘要、标题、标签、分类与引用辅助。
- 当前页面、路由、角色和权限摘要解释。

当前不可侵入：

- 身份、token 生命周期和权限最终判定。
- 附件访问控制和临时访问令牌。
- 未经确认直接执行治理、封禁、授权或数据写入。

### `RadishCatalyst`

当前只做文档级预留：

- 可预留玩家知识问答、进度解释、生产链规划和开发侧静态数据一致性检查。
- 不扩真实 schema、adapter、gateway smoke 或模型接线，直到上层明确首个真实任务面。

当前不可侵入：

- Godot 运行时、存档、任务、战斗、掉落、配方、公开等级和联机权威。
- 玩家侧不能泄露 `internal` 或默认隐藏的 `spoiler` 内容。

## 当前非目标

- 不把 `RadishMind` 做成替代业务内核的自治系统。
- 不把当前仓库写成已经具备完整 Dify / sub2api / axonhub 等同能力的产品。
- 不让模型替代 `RadishFlow` 求解、`Radish` 权限判定或 `RadishCatalyst` 游戏权威。
- 不把通用 unrestricted tool calling 当成当前默认能力。
- 不把平台锁死在单一模型、单一 provider、单一上游协议或单一对外接口上。
- 不自建与 `Radish` 冲突的用户身份、权限和部署真相源；未来 RadishMind 应作为 OIDC client 接入 `Radish`。
- 不把“参考 Radish”解释成复制 Radish 后端语言栈；除非未来必须直接复用 Radish 后端包或共发布，否则不新增 `.NET` 作为默认后端栈。
- 不默认继续真实模型产出、3B/4B 长跑、训练 JSONL、蒸馏或权重相关工作；这些内容后续作为独立专题重开。
- 不默认下载大模型、数据集或权重。
- 不把 `14B/32B` 写成当前自研主模型默认目标；首版仍优先本地可承受的小中型路线，长期本地部署上限暂定 `7B`。

## 实现原则

- 协议优先采用结构化 JSON。
- 能规则化或工具化的逻辑，不强行压给模型。
- 先把 runtime、session、tooling、governance 边界做清楚，再扩大训练和接入面。
- 代码遵循对应语言的惯用实践，命名清楚、职责明确、边界稳定。
- 禁止语义不明的方法、空转 helper、过度泛化的 manager/factory 和晦涩抽象封装。
- 抽象必须服务于真实职责边界、复用或复杂度收敛；不能为了隐藏简单逻辑而增加理解成本。
