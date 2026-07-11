# RadishMind 文档入口

更新时间：2026-07-11

## 阅读原则

`docs/` 是 RadishMind 的正式文档源。关键入口文档只保留项目定位、当前阶段、最近进度、下一步和停止线；历史推演、批次细节、长实验观察和一次性讨论应沉淀到 `docs/devlogs/`、任务卡、manifest 或 run record 中。

文档默认按“短入口 + 专题页 + 证据附件”维护：入口文档保持可快速阅读，契约细节拆到稳定专题，长命令输出、批次流水和实验观察进入周志、任务卡附件、manifest、summary 或 run record。仓库级检查会对 Markdown 篇幅执行预算：入口文档超过 `250` 行失败，普通 Markdown 超过 `500` 行提示、超过 `800` 行失败，周志和任务卡超过 `350` 行提示、超过 `600` 行失败；临时超限必须在文件头写明 `markdown-size-allow:` 及拆分计划。

文档正文默认中文；没有稳定中文对应的专业名词、代码、命令、路径、配置键、协议字段、状态锚点、fixture / checker ID、外部产品名和必要引用保留原文。历史英文工程短语按入口文档、专题文档、任务卡和周志顺序逐批收口，不做会破坏机器检查 literal 的机械翻译；具体边界见 [文档语言治理 v1](document-language-governance-v1.md)。

新会话优先按以下顺序读取：

1. 本文件
2. [项目总览与使用指南](radishmind-project-guide.md)
3. [当前推进焦点](radishmind-current-focus.md)
4. [产品范围与目标](radishmind-product-scope.md)
5. [战略定义](radishmind-strategy.md)
6. [能力矩阵](radishmind-capability-matrix.md)
7. [阶段路线图](radishmind-roadmap.md)
8. [功能设计文档入口](features/README.md)
9. 与当次任务直接相关的细专题、平台专题、集成专题、架构、契约、任务卡、评测或周志

## 当前状态

- 当前成熟度是“内部开发者预览”；整改与当前执行顺序以 [工程健康与产品化整改专题 v1](platform/engineering-health-productization-remediation-v1.md) 和 [当前推进焦点](radishmind-current-focus.md) 为准。
- Workflow Draft Review Loop、Saved Draft PostgreSQL dev/test repository、R4 Gateway stdio worker pool 与受控 Workflow Executor v0 已于 2026-07-11 完成；当前第一产品主线转向运行历史 / durable dev-test run store 与执行可观测性设计，第一工程主线是 R5 Web 拆包和性能预算。旧 Production Secret Backend / Storage Adapter next dependency 只保留为历史 checker 锚点，不再作为当前开发任务。
- `RadishMind` 已正式从“模型实验 / 接入准备仓库”的狭义口径，收口为 `Radish` 体系下的 AI 工具、工作流、模型网关和 Copilot 集成平台。
- 当前仓库主线不再只是等待其他项目真实接入；长期按四个一级产品面和五条工程主线组织。四个产品面是 `User Workspace`、`Admin Control Plane`、`Model Gateway / API Distribution`、`Workflow / Agent Runtime`；`Image Generation / Artifact Return` 作为横切适配能力保留。五条工程主线是 `Runtime Service`、`Conversation & Session`、`Tooling Framework`、`Evaluation & Governance`、`Model Adaptation`。
- 当前项目的更强正式定义已经固定在 [战略定义](radishmind-strategy.md)：`RadishMind` 是 `AI Tools / Workflow / Model Gateway / Copilot Integration Platform`，核心价值是把 AI 应用构建、工作流运行、模型 API 分发、多模型接入和 Copilot 集成收口成可控产品能力。
- 部署方式、数据库和登录 / 授权默认参考 `Radish`（https://github.com/laugh0608/Radish）；未来 RadishMind 应作为 OIDC client 接入 `Radish`，不自建第二套身份与权限真相源，也不因参考 Radish 默认引入 `.NET` / ASP.NET Core。
- 平台后续必须同时具备两类兼容能力：南向接入自研模型与外部模型，北向对外提供常见 AI 协议接口；当前仓库已落地最小 `provider registry` 骨架，并由同一 southbound 入口收口 `openai-compatible / HuggingFace / Ollama / gemini-native / anthropic-messages` 调用基础，同时已落地 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 的第一版 bridge-backed 兼容面，其中 `/v1/models` 已暴露 provider-qualified profile inventory 和 `/v1/models/{id}` lookup，并与请求选择、diagnostics 共享 selectable profile metadata。
- 平台表层实现分工已固定为：`UI=React + Vite + TypeScript`、`Platform / Control Plane / Gateway=Go`、`Model / Eval Side=Python`，并且所有层都只消费 `contracts/` 里的 canonical protocol。
- 仓库 Python 开发环境默认使用仓库根 `.venv`；首次拉取后，macOS / Linux / WSL 先执行 `./scripts/bootstrap-dev.sh`，Windows / PowerShell 先执行 `pwsh ./scripts/bootstrap-dev.ps1`，再运行 `check-repo` 系列入口。
- `P1 Runtime Foundation` 已达到 short close：`services/platform/` 下的最小 `Go` 平台服务层 bootstrap 已落地，当前已固定 `HTTP` 服务启动、`/healthz`、`/v1/models`、`/v1/chat/completions`、`/v1/responses` 和 `/v1/messages` 的第一版 bridge，并补齐本地 wrapper、配置文件层级、deployment smoke、结构化 diagnostics、provider/profile discoverability、request-level observability 与 error taxonomy。
- `P2 Session & Tooling Foundation` 已进入 close candidate / governance-only，并已补上最小可消费产品骨架：`GET /v1/session/metadata`、`GET /v1/tools/metadata` 与 `POST /v1/tools/actions` 能返回 session/tool metadata 和明确 blocked action response。当前仍不实现真实工具执行器、长期记忆、durable session/checkpoint/audit/result store、materialized result reader、业务写回或 replay executor，也不声明完整 `negative_regression_suite` 已完成。
- `session-tooling-negative-regression-suite-readiness.json`、`session-tooling-route-negative-coverage-matrix.json`、`session-tooling-route-smoke-readiness-rollup.json`、`session-tooling-short-close-readiness-delta.json`、`session-tooling-stop-line-manifest.json`、`session-tooling-upper-layer-confirmation-flow-readiness.json`、`session-tooling-short-close-entry-checklist.json` 等 P2 fixture 继续作为 governance-only 停止线证据保留，固定 `P2 short close` 前的 `not_satisfied` 条件；它们不再是默认新增工作方向。
- `P3 Local Product Shell / Ops Surface` 的本地只读产品壳已达到 `local usable / read-only close`：`GET /v1/platform/overview` 作为首个只读产品 overview，汇总服务状态、model/profile inventory、session/tooling metadata、blocked action route 和停止线；`GET /v1/platform/local-smoke` 作为本地开发 readiness 摘要，汇总 healthz、overview contract、model inventory、session/tooling metadata、blocked action no-side-effects、本地 console CORS 和停止线。当前已补 overview / local-smoke consumer smoke、`apps/radishmind-console/` 本地 console 壳、Dev Diagnostics、`Local Readiness` 面板、Provider/Profile Details、Stop-line Details、overview / local-smoke failure surface、console behavior / visual smoke record / dev entry / production boundary gate 与 P3 checklist，供本地控制台或上层 UI 只读消费；production secret backend、process supervisor、部署环境隔离和 console production packaging 仍作为后续 hardening 缺口保留。
- `Production Ops Hardening v1` 的静态边界已收口：production config / secret boundary、process supervisor / startup、deployment environment isolation、console production packaging、Docker local/test/prod compose、镜像命名治理、deployment readiness 静态 smoke、container smoke runbook 和 record template 都已可检查。Production secret backend 仍是 reference-only / static readiness 链路；audit store storage adapter 已推进到 provider account / resource / endpoint review（`audit_store_storage_adapter_provider_account_resource_endpoint_review_defined`），下一项为 `storage_adapter_runtime_implementation_entry_refresh_after_provider_account_resource_endpoint_review`，历史 managed database product selection readiness（`audit_store_storage_adapter_managed_database_product_selection_readiness_defined`）、managed database product selection review（`audit_store_storage_adapter_managed_database_product_selection_review_defined`）、`storage_adapter_runtime_implementation_entry_refresh_after_managed_database_product_selection_review`、concrete managed database provider selection readiness（`audit_store_storage_adapter_concrete_managed_database_provider_selection_readiness_defined` / `storage_adapter_concrete_managed_database_provider_selection_review`）、concrete managed database provider selection review（`audit_store_storage_adapter_concrete_managed_database_provider_selection_review_defined`）、review 后 runtime entry refresh（`audit_store_storage_adapter_runtime_implementation_entry_refresh_after_concrete_managed_database_provider_selection_review_defined`）、`storage_adapter_provider_account_resource_endpoint_readiness` 和 `storage_adapter_provider_account_resource_endpoint_review` 已被本批消费。这不是 production resolver runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、audit store runtime、storage adapter runtime、DB provider、connection provider、SQL / DDL、repository mode 或 production API ready；读法见 [storage adapter 静态准入说明](platform/production-secret-backend-audit-store-storage-adapter-static-readiness-guide-v1.md)。
- 2026-07-08 覆盖说明：audit store storage adapter 已继续完成 [Production Secret Backend Audit Store Storage Adapter Provider Account / Resource / Endpoint Review v1](platform/production-secret-backend-audit-store-storage-adapter-provider-account-resource-endpoint-review-v1.md)，最新状态为 `audit_store_storage_adapter_provider_account_resource_endpoint_review_defined`，review decision 为 `provider_account_resource_endpoint_review_defined_runtime_blocked`，当前下一项为 `storage_adapter_runtime_implementation_entry_refresh_after_provider_account_resource_endpoint_review`。该状态只审查 metadata-only readiness 是否可作为后续 runtime entry refresh 输入，不选择真实 vendor、cloud product、provider account、provider resource、endpoint、region、DSN、driver version、DB provider、connection provider、SQL / DDL、storage adapter runtime、audit store runtime、repository mode 或 production API。
- `Provider Runtime & Health v1` 已完成 `provider-capability-matrix-v1`、`provider-health-smoke-v1`、`provider-selection-policy-v1`、`provider-retry-fallback-policy-v1` 和 `provider-runtime-docs-refresh` 五个可检查切片，均已进入 fast baseline。当前结论是 provider capability、离线 health smoke、request-side selection、retry/fallback 审计策略和文档收口已可复验；不把 provider health 写成 production readiness，也不把 optional live health、retry/fallback execution 或 production secret backend 写成已完成。`UI Design Topic / React 第二批` 已进入 close candidate，P4 真实模型产出、3B/4B 长跑、训练 JSONL、蒸馏和权重相关工作转入后置专题。
- `Control Plane / User Workspace / Workflow v1` 已用 `product-surface-v1-boundary`、`control-plane-data-boundary`、`radish-oidc-client-preconditions`、`gateway-api-key-quota-readiness`、`workflow-definition-run-record-boundary`、read-side contract、任务卡、fixture 和 checker 固定四个产品面的服务边界、数据边界、read model、read-only route、workflow definition / run record、正式 UI 边界、dev-only live consumer、Workflow review surface、Model Gateway evidence 和 Admin readiness。`Saved Workflow Draft v1` 已落地 platform Go domain service、dev-only HTTP route + web consumer、formal store selector、静态 schema artifact、repository adapter、adapter smoke execution 和 production auth runtime bridge；随后固定 repository mode enablement、schema migration runner readiness、runner implementation entry review、database connection / schema marker preconditions、connection provider entry review / refresh v2、database secret resolver readiness / implementation entry review / runtime dependency refresh、database driver / DSN / TLS policy、role policy、connection smoke strategy、connection lifecycle readiness、schema marker runtime dependency refresh、Radish OIDC upstream evidence refresh 和 token validation auth middleware runtime entry review，并由 production secret backend 链路补齐 test-only fake resolver runtime 和真实 resolver 前置说明。`Workflow Node Designer Surface v1` 已作为 Builder 体验专题归位并接入 `@xyflow/react` active-draft 画布、受控 layout metadata、受控 edge mutation、validation overlay navigation 和 Review Handoff evidence，但不实现 builder runtime。当前仍没有 repository mode runtime、真实数据库，仍不接真实 OIDC，不发放真实 API key，不创建数据库 schema 或 migration，也不创建 SQL migration runner、schema marker reader / writer、production resolver runtime、backend health runtime、no leakage smoke runtime、credential handle runtime、approval runtime、audit store runtime、Radish OIDC token validation、membership adapter、production API、workflow executor、confirmation、writeback 或 replay。
- 2026-07-11 覆盖说明：显式 `postgres_dev_test`、真实 migration / schema marker 和受控 Workflow Executor v0 已成立；上段“没有真实数据库 / workflow executor”只描述 production repository 与完整运行时停止线。当前仍不开放 production repository、OIDC、production secret、tool、confirmation、writeback、replay 或公开生产 API。
- 2026-06-14 起，开发节奏改为功能设计文档先行；2026-06-15 起，专题进一步分层：产品面大方向写入 `docs/features/*.md`，具体功能和复杂页面写入对应子目录，平台横切能力写入 `docs/platform/`，外部接入写入 `docs/integrations/`。task card 只服务具体实现批次、前置条件或高风险边界，专项 checker 只在协议、schema、执行边界、生产声明、外部 provider 风险或高风险能力变化时新增。
- read-side 证据锚点继续保留为：`control-plane-read-model-v1` / read model、`control-plane-read-route-contract-v1` / read-only route、`control-plane-read-response-fixtures-v1` / response fixture、`control-plane-read-negative-contract-v1` / negative contract、fake store、`control-plane-read-implementation-preconditions-v1`、`control-plane-read-fake-store-handler-plan-v1` / fake-store-backed read handler plan、`control-plane-read-fake-store-handler-implementation-v1` / fake-store-backed read handler implementation、`control-plane-durable-read-foundation-v1` / repository interface + fake-store bridge、`control-plane-read-auth-db-preconditions-v1` / 真实 auth/db 前置条件、`control-plane-read-consumer-contract-v1` / TypeScript consumer contract、`control-plane-read-formal-ui-boundary-v1` / 正式 UI 边界、`control-plane-read-formal-ui-implementation-readiness-v1` / 正式 UI 实现 readiness。
- 这些 read-side 锚点的说明边界保持不变：早期模型、route、fixture、negative contract 与 fake-store 计划不创建数据库 schema 或 migration；`control-plane-read-implementation-preconditions-v1`、`control-plane-read-fake-store-handler-implementation-v1`、`control-plane-read-auth-db-preconditions-v1` 和 `control-plane-read-consumer-contract-v1` 不实现完整 read-side API；`control-plane-read-fake-store-handler-plan-v1` 不实现 Go handler。
- `control-plane-read-auth-db-preconditions-v1` 仍只固定真实 auth/db 前置条件，不代表真实 auth middleware、数据库 query、repository、OIDC 或 production API consumer 已实现。
- `control-plane-read-formal-ui-boundary-v1` 只固定正式 UI 边界，`control-plane-read-formal-ui-implementation-readiness-v1` 只固定正式 UI 实现 readiness；二者不直接实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay。
- `control-plane-read-shared-shell-v1` 已在 `apps/radishmind-web/` 固定只读 shared shell；它只消费离线 read-side view model，不请求真实后端。
- read-side 页面集合包括 `control-plane-read-admin-tenant-overview-v1` / `admin-tenant-overview`、`control-plane-read-admin-audit-log-v1` / `admin-audit-log`、普通离线 Admin Operations Review / Readiness、Admin Provider/Profile & Deployment Evidence Review / Readiness、`control-plane-read-workspace-applications-v1` / `workspace-applications`、`control-plane-read-workspace-api-keys-v1` / `workspace-api-keys`、`control-plane-read-workspace-usage-quota-v1` / `workspace-usage-quota`、`control-plane-read-workspace-workflow-definitions-v1` / `workspace-workflow-definitions`、`control-plane-read-workspace-run-history-v1` / `workspace-run-history`、User Workspace Home、Model Gateway 四个 evidence 面板和 workflow review 只读面板，全部保持只读。
- 当前 read-side UI 的门禁策略已调整：`control-plane-read-formal-ui-readiness-close-v1` 已完成页面集合聚合收口，使用 surface matrix / checker 管理普通只读展示页；`control-plane-read-dev-live-consumer-v1` 只允许显式 dev-only opt-in 消费 fake-store-backed handler 和测试身份上下文；`control-plane-read-auth-store-transition-preconditions-v1` 只固定迁移前置条件，不直接接真实数据库、Radish OIDC、API key lifecycle、quota enforcement、billing、workflow executor、confirmation、writeback 或 replay。
- Workflow Function Surface 已完成 `workflow-function-surface-readiness-close-v1`，状态为 `workflow_function_surface_readiness_closed`；该状态只说明离线 function surface readiness、页面集合和检查链路已收口，不代表 workflow executor、confirmation、writeback 或 replay 可用。
- 2026-06-14 阶段策略调整：普通展示、文案、布局和 evidence 组织不再默认逐项新增 task card / fixture / checker；Image Path 已在 `coerce_response_document` 完成 metadata-only response builder 接线，Control Plane Read 已完成 `ControlPlaneReadRepository` interface 边界和 fake-store-backed handlers interface 化。二者都不进入数据库、OIDC、token validation、repository adapter、SQL / migration、store selector、production API consumer、API key lifecycle、quota enforcement、billing、workflow executor、confirmation、writeback 或 replay。
- 上述 read-side UI 相关切片仍不直接实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay；新增 readiness 只固定正式 UI 实现前的工程边界。
- 既有 `M3` service/API smoke matrix 与 `M4` broader review、`3B/4B` capacity review 继续保留为冻结证据和门禁；它们不再是当前唯一主线，也不再默认继续深挖同一批样本。
- `RadishFlow` 仍是第一优先应用面，但当前只冻结 gateway、UI consumption 和 candidate handoff 门禁；上层尚未具备真实接入能力前，不继续细化假想接线。
- `Radish` 当前保留 docs QA、文档检索增强和结构化问答资产；真实上层接入仍等待。
- `RadishCatalyst` 仍只做文档级预留，不扩真实 schema、adapter、gateway smoke 或模型接线。
- 图片生成能力继续由 `RadishMind-Image Adapter` 与独立 backend 承接；主模型只负责结构化 intent、约束、审查和 artifact metadata。当前 Image Path 已补齐 artifact return、安全 runbook、backend adapter readiness、artifact runtime mapping readiness、implementation entry review、`image-artifact-store-binary-reader-boundary-readiness-v1`、`image-artifact-runtime-mapper-implementation-plan-v1` runtime mapper implementation plan、`image-artifact-runtime-mapper-implementation-entry-v1` runtime mapper implementation entry review、`image-artifact-runtime-mapper-implementation-v1` runtime mapper implementation task card、`image-artifact-runtime-mapper-runtime-implementation-v1` metadata-only runtime mapper、`image-artifact-runtime-mapper-response-consumer-integration-review-v1` response consumer 入口评审证据、`image-artifact-response-consumer-implementation-readiness-v1` response consumer implementation readiness、`image-artifact-response-consumer-implementation-v1` response consumer implementation task card、`image-artifact-response-consumer-runtime-implementation-v1` metadata-only response consumer runtime、`image-artifact-response-builder-integration-entry-review-v1` response builder integration entry review、`image-artifact-response-builder-integration-v1` response builder integration task card、`image-artifact-response-builder-runtime-integration-entry-review-v1` runtime integration entry review 与 `image-artifact-response-builder-runtime-integration-implementation-v1` metadata-only runtime integration；状态为 `image_artifact_response_builder_runtime_integration_implemented`。仍不调用真实生图 backend、不改 `CopilotResponse` schema、不创建 artifact store 或 public URL。

## 文档约束

- 入口文档必须简短，优先描述“项目是什么、现在能推进什么、哪些边界不能越”。
- 不在入口文档重复长批次流水、历史失败细节或完整命令输出；这些内容放入周志、实验记录、summary 或 task card。
- 回答“今天要做什么”这类问题时，默认读 [当前推进焦点](radishmind-current-focus.md) 与 [能力矩阵](radishmind-capability-matrix.md)，不要默认展开长契约或长评测文档。
- 新增或更新文档时，优先更新既有正式文档；只有当内容有长期复用价值且无法自然归入现有文档时，才新增文档。
- 文档中提到外部项目时，默认使用项目名和在线仓库 URL，不写个人机器路径。

## 语言与实现约束

- 代码应趋近对应语言的惯用、清晰和可维护实践；本仓库按职责分层：模型训练、评测和脚本优先 `Python`，前端 UI 默认 `React + Vite + TypeScript`，服务 / `gateway` / `API` 可按职责采用 `Go`。
- 命名必须表达真实职责和领域含义，避免 `process_data`、`handle_item`、`manager`、`helper` 这类无法说明边界的泛名。
- 抽象只在能稳定表达职责边界、消除真实重复或收敛复杂度时引入；禁止为了“看起来通用”增加晦涩封装、空转方法或多层转发。
- 能用语言标准库、结构化 schema、明确数据类型和小而直接的函数解决的问题，不应写成难追踪的动态包装或隐式 fallback 链。
- 修复问题优先定位根因；不要用连续叠加兜底逻辑掩盖协议、数据或职责边界错误。

## 关键文档

- [当前推进焦点](radishmind-current-focus.md)
- [项目总览与使用指南](radishmind-project-guide.md)
- [产品范围与目标](radishmind-product-scope.md)
- [产品机会池](radishmind-product-ideas.md)
- [战略定义](radishmind-strategy.md)
- [能力矩阵](radishmind-capability-matrix.md)
- [系统架构](radishmind-architecture.md)
- [阶段路线图](radishmind-roadmap.md)
- [文档语言治理 v1](document-language-governance-v1.md)
- [功能设计文档入口](features/README.md)
- [Workflow 细专题入口](features/workflow/README.md)
- [Saved Workflow Draft v1 功能专题](features/workflow/saved-workflow-draft-v1.md)
- [User Workspace Saved Draft List v1 专题](features/workflow/user-workspace-saved-draft-list-v1.md)
- [Workflow Draft Designer Editing Model v2 专题](features/workflow/draft-designer-editing-model-v2.md)
- [Workflow Draft Node Attribute Editing Model v1 专题](features/workflow/draft-node-attribute-editing-model-v1.md)
- [Production Secret Audit Store Storage Adapter 静态准入说明 v1](platform/production-secret-backend-audit-store-storage-adapter-static-readiness-guide-v1.md)
- [Workflow Review Handoff Active Draft v1 专题](features/workflow/review-handoff-active-draft-v1.md)
- [Workflow Node Designer Surface v1 专题](features/workflow/node-designer-surface-v1.md)
- [Saved Workflow Draft Repository Contract Smoke Runner Implementation v1 专题](features/workflow/saved-workflow-draft-repository-contract-smoke-runner-implementation-v1.md)
- [平台专题入口](platform/README.md)
- [Production Secret Backend Fake Resolver Implementation v1](platform/production-secret-backend-fake-resolver-implementation-v1.md)
- [扩展与集成专题入口](integrations/README.md)
- [Radish OIDC Token / Membership Readiness v1](integrations/radish-oidc-token-membership-readiness-v1.md)
- [Radish OIDC Token / Membership Upstream Evidence Refresh v1](integrations/radish-oidc-token-membership-upstream-evidence-refresh-v1.md)
- [Control Plane Read-Side 契约](contracts/control-plane-read-side.md)
- [UI 设计规范](radishmind-ui-design-spec.md)
- [UI 设计参考](radishmind-ui-design-reference.md)
- [Production Ops Hardening v1 任务卡](task-cards/production-ops-hardening-v1-plan.md)
- [Production Ops Docker Deployment v1 任务卡](task-cards/production-ops-docker-deployment-v1-plan.md)
- [Provider Runtime & Health v1 任务卡](task-cards/provider-runtime-health-v1-plan.md)
- [Control Plane / User Workspace / Workflow v1 任务卡](task-cards/control-plane-user-workspace-workflow-v1-plan.md)
- [Workflow / Agent Runtime Function Surface v1 任务卡](task-cards/workflow-agent-runtime-function-surface-v1-plan.md)
- [部署目录说明](../deploy/README.md)
- [代码规范](radishmind-code-standards.md)
- [跨项目集成契约](radishmind-integration-contracts.md)
- [契约专题目录](contracts/README.md)
- [ADR 0002: 仓库集成边界](adr/0002-repository-integration-boundary.md)
- [RadishMind-Core 首版基座评估](radishmind-core-baseline-evaluation.md)
- [ADR 0001: 分支与 PR 治理](adr/0001-branch-and-pr-governance.md)
- [开发日志说明](devlogs/README.md)
- [任务卡入口](task-cards/README.md)
- [统一契约文件说明](../contracts/README.md)
- [数据集与评测目录说明](../datasets/README.md)
- [训练目录说明](../training/README.md)
- [脚本目录说明](../scripts/README.md)
