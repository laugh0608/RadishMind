# RadishMind 文档入口

更新时间：2026-06-04

## 阅读原则

`docs/` 是 RadishMind 的正式文档源。关键入口文档只保留项目定位、当前阶段、最近进度、下一步和停止线；历史推演、批次细节、长实验观察和一次性讨论应沉淀到 `docs/devlogs/`、任务卡、manifest 或 run record 中。

文档默认按“短入口 + 专题页 + 证据附件”维护：入口文档保持可快速阅读，契约细节拆到稳定专题，长命令输出、批次流水和实验观察进入周志、任务卡附件、manifest、summary 或 run record。仓库级检查会对 Markdown 篇幅执行预算：入口文档超过 `250` 行失败，普通 Markdown 超过 `500` 行提示、超过 `800` 行失败，周志和任务卡超过 `350` 行提示、超过 `600` 行失败；临时超限必须在文件头写明 `markdown-size-allow:` 及拆分计划。

新会话优先按以下顺序读取：

1. 本文件
2. [项目总览与使用指南](radishmind-project-guide.md)
3. [当前推进焦点](radishmind-current-focus.md)
4. [产品范围与目标](radishmind-product-scope.md)
5. [战略定义](radishmind-strategy.md)
6. [能力矩阵](radishmind-capability-matrix.md)
7. [阶段路线图](radishmind-roadmap.md)
8. 与当次任务直接相关的架构、契约、任务卡、评测或周志

## 当前状态

- `RadishMind` 已正式从“模型实验 / 接入准备仓库”的狭义口径，收口为 `Radish` 体系下的 AI 工具、工作流、模型网关和 Copilot 集成平台。
- 当前仓库主线不再只是等待其他项目真实接入；接下来按四个产品面和五条工程主线推进。四个产品面是 `User Workspace`、`Admin Control Plane`、`Model Gateway / API Distribution`、`Workflow / Agent Runtime`；五条工程主线是 `Runtime Service`、`Conversation & Session`、`Tooling Framework`、`Evaluation & Governance`、`Model Adaptation`。
- 当前项目的更强正式定义已经固定在 [战略定义](radishmind-strategy.md)：`RadishMind` 是 `AI Tools / Workflow / Model Gateway / Copilot Integration Platform`，核心价值是把 AI 应用构建、工作流运行、模型 API 分发、多模型接入和 Copilot 集成收口成可控产品能力。
- 部署方式、数据库和登录 / 授权默认参考 `Radish`（https://github.com/laugh0608/Radish）；未来 RadishMind 应作为 OIDC client 接入 `Radish`，不自建第二套身份与权限真相源，也不因参考 Radish 默认引入 `.NET` / ASP.NET Core。
- 平台后续必须同时具备两类兼容能力：南向接入自研模型与外部模型，北向对外提供常见 AI 协议接口；当前仓库已落地最小 `provider registry` 骨架，并由同一 southbound 入口收口 `openai-compatible / HuggingFace / Ollama / gemini-native / anthropic-messages` 调用基础，同时已落地 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 的第一版 bridge-backed 兼容面，其中 `/v1/models` 已暴露 provider-qualified profile inventory 和 `/v1/models/{id}` lookup，并与请求选择、diagnostics 共享 selectable profile metadata。
- 平台表层实现分工已固定为：`UI=React + Vite + TypeScript`、`Platform / Control Plane / Gateway=Go`、`Model / Eval Side=Python`，并且所有层都只消费 `contracts/` 里的 canonical protocol。
- 仓库 Python 开发环境默认使用仓库根 `.venv`；首次拉取后，macOS / Linux / WSL 先执行 `./scripts/bootstrap-dev.sh`，Windows / PowerShell 先执行 `pwsh ./scripts/bootstrap-dev.ps1`，再运行 `check-repo` 系列入口。
- `P1 Runtime Foundation` 已达到 short close：`services/platform/` 下的最小 `Go` 平台服务层 bootstrap 已落地，当前已固定 `HTTP` 服务启动、`/healthz`、`/v1/models`、`/v1/chat/completions`、`/v1/responses` 和 `/v1/messages` 的第一版 bridge，并补齐本地 wrapper、配置文件层级、deployment smoke、结构化 diagnostics、provider/profile discoverability、request-level observability 与 error taxonomy。
- `P2 Session & Tooling Foundation` 已进入 close candidate / governance-only，并已补上最小可消费产品骨架：`GET /v1/session/metadata`、`GET /v1/tools/metadata` 与 `POST /v1/tools/actions` 能返回 session/tool metadata 和明确 blocked action response。当前仍不实现真实工具执行器、长期记忆、durable session/checkpoint/audit/result store、materialized result reader、业务写回或 replay executor，也不声明完整 `negative_regression_suite` 已完成。
- `session-tooling-negative-regression-suite-readiness.json`、`session-tooling-route-negative-coverage-matrix.json`、`session-tooling-route-smoke-readiness-rollup.json`、`session-tooling-short-close-readiness-delta.json`、`session-tooling-stop-line-manifest.json`、`session-tooling-upper-layer-confirmation-flow-readiness.json`、`session-tooling-short-close-entry-checklist.json` 等 P2 fixture 继续作为 governance-only 停止线证据保留，固定 `P2 short close` 前的 `not_satisfied` 条件；它们不再是默认新增工作方向。
- `P3 Local Product Shell / Ops Surface` 的本地只读产品壳已达到 `local usable / read-only close`：`GET /v1/platform/overview` 作为首个只读产品 overview，汇总服务状态、model/profile inventory、session/tooling metadata、blocked action route 和停止线；`GET /v1/platform/local-smoke` 作为本地开发 readiness 摘要，汇总 healthz、overview contract、model inventory、session/tooling metadata、blocked action no-side-effects、本地 console CORS 和停止线。当前已补 overview / local-smoke consumer smoke、`apps/radishmind-console/` 本地 console 壳、Dev Diagnostics、`Local Readiness` 面板、Provider/Profile Details、Stop-line Details、overview / local-smoke failure surface、console behavior / visual smoke record / dev entry / production boundary gate 与 P3 checklist，供本地控制台或上层 UI 只读消费；production secret backend、process supervisor、部署环境隔离和 console production packaging 仍作为后续 hardening 缺口保留。
- `Production Ops Hardening v1` 的静态边界已收口：production config / secret boundary、process supervisor / startup、deployment environment isolation、console production packaging、Docker local/test/prod compose、镜像命名治理、deployment readiness 静态 smoke、container smoke runbook 和 record template 都已可检查。2026-05-26 已完成一次本地 `docker_local` container smoke 运行记录；后续只有在明确测试或生产前复核窗口时才补测试环境 smoke 或 production preflight 证据。
- `Provider Runtime & Health v1` 已完成 `provider-capability-matrix-v1`、`provider-health-smoke-v1`、`provider-selection-policy-v1`、`provider-retry-fallback-policy-v1` 和 `provider-runtime-docs-refresh` 五个可检查切片，均已进入 fast baseline。当前结论是 provider capability、离线 health smoke、request-side selection、retry/fallback 审计策略和文档收口已可复验；不把 provider health 写成 production readiness，也不把 optional live health、retry/fallback execution 或 production secret backend 写成已完成。`UI Design Topic / React 第二批` 已进入 close candidate，P4 真实模型产出、3B/4B 长跑、训练 JSONL、蒸馏和权重相关工作转入后置专题。
- `Control Plane / User Workspace / Workflow v1` 已新增任务卡，作为下一条平台主线的边界文档：`product-surface-v1-boundary` 已用 fixture / checker 固定正式用户端、管理端、模型网关和 workflow runtime 的服务边界、数据边界和停止线；`control-plane-data-boundary` 已固定 control plane ownership；`radish-oidc-client-preconditions` 已固定 OIDC 前置条件；`gateway-api-key-quota-readiness` 已固定 API key、quota、rate limit、cost ledger 和 trace 前置条件；`workflow-definition-run-record-boundary` 已固定 workflow definition、run record、状态流转、失败分类、审计证据和停止线；`control-plane-read-model-v1` 已固定下一步只读 read model；`control-plane-read-route-contract-v1` 已固定七类 read-only route contract；`control-plane-read-response-fixtures-v1` 已固定 response fixture、统一 envelope 和 `failure_code`；`control-plane-read-negative-contract-v1` 已固定 negative contract、forbidden method / query / fallback 和 fail-closed 拒绝样例；`control-plane-read-implementation-preconditions-v1` 已固定 handler ownership、fake store strategy、auth middleware dependency、response conformance 和 negative route smoke readiness；`control-plane-read-fake-store-handler-plan-v1` 已固定 fake-store-backed read handler plan；`control-plane-read-fake-store-handler-implementation-v1` 已完成七条 fake-store-backed read handler implementation；`control-plane-read-consumer-contract-v1` 已固定 TypeScript consumer contract，正式 UI 边界、正式 UI readiness、shared shell、七个只读页面、formal UI readiness close、dev-only live consumer 和 auth/store transition preconditions 已收口；read store 迁移前置链路已推进到 repository contract types implementation、静态 contract smoke runner implementation、repository interface readiness、adapter implementation readiness refresh 和 store selector enablement preconditions。当前仍不实现完整 read-side API、OIDC、token validation、数据库、repository interface、repository adapter、repository migration、repository implementation、store selector、API key lifecycle、quota enforcement、rate limit、billing、cost ledger、workflow builder、workflow executor、run replay、run resume、materialized result reader、confirmation、writeback 或 replay，不接真实 OIDC，不发放真实 API key，不创建数据库 schema 或 migration。
- `control-plane-read-auth-db-preconditions-v1` 仍只固定真实 auth/db 前置条件，不代表真实 auth middleware、数据库 query、repository、OIDC 或 production API consumer 已实现。
- `control-plane-read-formal-ui-boundary-v1` 只固定正式 UI 边界，`control-plane-read-formal-ui-implementation-readiness-v1` 只固定正式 UI 实现 readiness；二者不直接实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay。
- `control-plane-read-shared-shell-v1` 已在 `apps/radishmind-web/` 固定只读 shared shell；它只消费离线 read-side view model，不请求真实后端。
- read-side 页面集合包括 `control-plane-read-admin-tenant-overview-v1` / `admin-tenant-overview`、`control-plane-read-admin-audit-log-v1` / `admin-audit-log`、`control-plane-read-workspace-applications-v1` / `workspace-applications`、`control-plane-read-workspace-api-keys-v1` / `workspace-api-keys`、`control-plane-read-workspace-usage-quota-v1` / `workspace-usage-quota`、`control-plane-read-workspace-workflow-definitions-v1` / `workspace-workflow-definitions` 和 `control-plane-read-workspace-run-history-v1` / `workspace-run-history`，全部保持只读。
- 当前 read-side UI 的门禁策略已调整：`control-plane-read-formal-ui-readiness-close-v1` 已完成页面集合聚合收口，使用 surface matrix / checker 管理普通只读展示页；`control-plane-read-dev-live-consumer-v1` 只允许显式 dev-only opt-in 消费 fake-store-backed handler 和测试身份上下文；`control-plane-read-auth-store-transition-preconditions-v1` 只固定迁移前置条件，不直接接真实数据库、Radish OIDC、API key lifecycle、quota enforcement、billing、workflow executor、confirmation、writeback 或 replay。
- 上述 read-side UI 相关切片仍不直接实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay；新增 readiness 只固定正式 UI 实现前的工程边界。
- 既有 `M3` service/API smoke matrix 与 `M4` broader review、`3B/4B` capacity review 继续保留为冻结证据和门禁；它们不再是当前唯一主线，也不再默认继续深挖同一批样本。
- `RadishFlow` 仍是第一优先应用面，但当前只冻结 gateway、UI consumption 和 candidate handoff 门禁；上层尚未具备真实接入能力前，不继续细化假想接线。
- `Radish` 当前保留 docs QA、文档检索增强和结构化问答资产；真实上层接入仍等待。
- `RadishCatalyst` 仍只做文档级预留，不扩真实 schema、adapter、gateway smoke 或模型接线。
- 图片生成能力继续由 `RadishMind-Image Adapter` 与独立 backend 承接；主模型只负责结构化 intent、约束、审查和 artifact metadata。

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
- [Control Plane Read-Side 契约](contracts/control-plane-read-side.md)
- [UI 设计规范](radishmind-ui-design-spec.md)
- [UI 设计参考](radishmind-ui-design-reference.md)
- [Production Ops Hardening v1 任务卡](task-cards/production-ops-hardening-v1-plan.md)
- [Production Ops Docker Deployment v1 任务卡](task-cards/production-ops-docker-deployment-v1-plan.md)
- [Provider Runtime & Health v1 任务卡](task-cards/provider-runtime-health-v1-plan.md)
- [Control Plane / User Workspace / Workflow v1 任务卡](task-cards/control-plane-user-workspace-workflow-v1-plan.md)
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
