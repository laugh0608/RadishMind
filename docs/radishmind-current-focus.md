# RadishMind 当前推进焦点

更新时间：2026-06-03

## 文档目的

本文档用于回答“根据项目规划和开发进度，今天要做什么以推进开发”。它是新会话的短入口，不承载历史细节、完整实验日志或长契约说明。

默认只读本文件仍应能判断下一步方向；只有准备实施具体改动时，才跳转到下方按需文档。

## 当前阶段

当前已经从“M3/M4 收口后的被动等待”切换到“平台重定义后的平台本体建设期”；`P1 Runtime Foundation` 已达到 short close，不再默认继续横向细磨 provider/config/diagnostics/observability 同一层：

- `M3` 的 gateway、service smoke、UI consumption 与 candidate handoff 继续作为冻结门禁保留。
- `M4` 的 broader 15/15 `reviewed_pass` 与 `3B/4B` guided capacity review 继续作为路线证据保留。
- 当前不继续扩同一批 `M4` 实验，不提前设计不存在的上层真实接线。
- 平台请求级观测和错误分类短收口已进入 `Go` northbound 层与平台单元测试，`P2 Session & Tooling Foundation` 也已从治理链路推进到最小 metadata / blocked 产品外壳。

当前产品定位进一步明确为 `Radish` 体系下的 AI 工具、工作流、模型网关和 Copilot 集成平台。长期产品面固定为 `User Workspace`、`Admin Control Plane`、`Model Gateway / API Distribution` 和 `Workflow / Agent Runtime`：用户端支持 AI 应用、工作流、API key、调用量和运行记录；管理端支持租户、权限、provider/profile、模型路由、quota、secret、审计和部署状态；模型网关负责模型 API 分发；workflow runtime 负责应用执行和受控 agent loop。部署方式、数据库和登录 / 授权优先参考 `Radish`，未来 RadishMind 作为 OIDC client 接入 `Radish`。参考 `Radish` 不代表默认引入 `.NET`；当前长期语言边界固定为 Go control plane / gateway、Python 模型与评测、TypeScript/Vite 前端。当前本地 console 仍只是 ops surface，不是正式用户端或生产管理端。

`P3 Local Product Shell / Ops Surface` 的本地只读产品壳已经达到 `local usable / read-only close`：overview、local-smoke、console dev entry、Dev Diagnostics、Local Readiness、failure surface、Stop-line Details、Provider/Profile Details 和对应轻量门禁已形成可复验闭环。后续不再默认继续补同类只读 console 小切片，除非真实使用暴露新的可观测缺口。

`Production Ops Hardening v1` 的静态边界已经达到 close：production config / secret boundary、production-secret-backend-contract、production-secret-backend-implementation-readiness、secret-ref-schema-and-fixtures、startup / supervisor boundary、environment isolation、console production package smoke、docker local/test/prod compose、镜像命名治理、deployment readiness 静态 smoke、container smoke runbook 和 record template 都已可检查。`production-secret-backend-contract` 已落地为 `scripts/checks/fixtures/production-ops-secret-backend-contract.json` 与 `scripts/check-production-ops-secret-backend-contract.py`，只定义未来 external secret backend adapter contract、secret reference、脱敏输出和禁止项；`production-secret-backend-implementation-readiness` 已落地为 `docs/task-cards/production-secret-backend-implementation-v1-plan.md`、`scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json` 与 `scripts/check-production-ops-secret-backend-implementation-readiness.py`，只固定真实实现前必须先满足的 schema、注入点、profile binding、脱敏审计、failure taxonomy、测试策略、operator runbook 和 rotation / audit policy；`secret-ref-schema-and-fixtures` 已落地为 `contracts/production-secret-reference.schema.json`、`scripts/checks/fixtures/production-secret-reference-basic.json` 与 `scripts/check-production-secret-reference-contract.py`，只保存 secret reference 和脱敏字段，不保存 secret value。上述切片都不实现真实云 secret 服务、不写入真实 secret、不声明 production ready。2026-05-26 已在明确 Docker 运行窗口下完成一次 `docker_local` container smoke，运行记录写入 `tmp/production-ops/container-smoke/20260526T2006-local-docker-smoke.json`；该记录只证明本地 mock provider 容器 smoke 跑通，不声明测试环境 smoke、生产前复核或 production ready。P3 的 production secret backend、process supervisor、deployment environment isolation 和 console production packaging 仍保持未完成，但它们不再阻塞下一条平台主线。

P3 checklist 继续由 `p3-local-product-shell-short-close-checklist.json` 与 `check-p3-local-product-shell-short-close-checklist.py` 固定 production secret backend、process supervisor、部署环境隔离和 console production packaging 仍为 `not_satisfied`。

`UI Design Topic / Pencil Draft` 已完成正式 React 第二批前的主要设计覆盖：`docs/designs/radishmind-console-ops-surface-v0.pen` 已包含 ready、failure/stale、narrow、loading/empty、settings/permissions、blocked action detail 与 token mapping notes。2026-05-23 Pencil `snapshot_layout` 已无 layout problems，`apps/radishmind-console/` 第二批 ops surface 结构重排已进入 `close candidate`，本地 mock platform ready 态和桌面 / 窄屏临时截图已复核。第二批只调整信息架构和视觉层级，不新增 API、不接 executor、不做 confirmation、writeback 或 replay。

## 当前优先做什么

1. `Control Plane / User Workspace / Workflow v1`：已新增 [任务卡](task-cards/control-plane-user-workspace-workflow-v1-plan.md)，作为下一条平台主线的边界文档。`product-surface-v1-boundary`、`control-plane-data-boundary`、`radish-oidc-client-preconditions`、`gateway-api-key-quota-readiness` 与 `workflow-definition-run-record-boundary` 已分别落地为 `product-surface-v1-boundary.json` / `check-product-surface-v1-boundary.py`、`control-plane-data-boundary.json` / `check-control-plane-data-boundary.py`、`radish-oidc-client-preconditions.json` / `check-radish-oidc-client-preconditions.py`、`gateway-api-key-quota-readiness.json` / `check-gateway-api-key-quota-readiness.py`、`workflow-definition-run-record-boundary.json` / `check-workflow-definition-run-record-boundary.py`；read-side 已落地 read model、route contract、response fixture、negative contract、implementation preconditions、fake-store handler plan / implementation、auth/db preconditions、consumer contract、`control-plane-read-formal-ui-boundary-v1`、`control-plane-read-formal-ui-implementation-readiness-v1` 与 `control-plane-read-shared-shell-v1`；`control-plane-read-admin-tenant-overview-v1`、`control-plane-read-admin-audit-log-v1`、`control-plane-read-workspace-applications-v1`、`control-plane-read-workspace-api-keys-v1`、`control-plane-read-workspace-usage-quota-v1`、`control-plane-read-workspace-workflow-definitions-v1` 和 `control-plane-read-workspace-run-history-v1` 已在 `apps/radishmind-web/` 内形成只读页面切片；`control-plane-read-formal-ui-readiness-close-v1` 已用 surface matrix 聚合校验七个页面的 route binding、状态预览、request / audit ref、forbidden output guard 与停止线；`control-plane-read-dev-live-consumer-v1` 已建立 dev-only live read consumer；repository/read store 迁移链路已推进到 `control-plane-read-repository-interface-readiness-v1`：Go contract type 与静态 contract/type runner 已落地，未来 `ControlPlaneReadRepository` interface method matrix 和 adapter gate 已固定为 `repository_interface_readiness_defined`。该阶段仍不直接实现数据库、OIDC、executor、confirmation、writeback 或 replay，也不直接实现 repository interface、SQL、migration、migration runner、repository adapter、真实数据库、Radish OIDC、token validation、production API consumer、repository implementation、store selector、API key lifecycle、quota enforcement、rate limit、billing、cost ledger、workflow builder、workflow executor、run replay、run resume 或 materialized result reader；也不把当前 ops console 改成完整产品或生产管理端。

repository/read store 已完成状态字面量继续保留为：`repository_contract_preconditions_defined`、`disabled_database_guard_defined`、`repository_contract_smoke_defined`、`repository_implementation_readiness_defined`、`store_selection_readiness_defined`、`schema_migration_readiness_defined`、`repository_contract_types_readiness_defined`、`repository_contract_types_implemented`、`repository_contract_smoke_runner_readiness_defined`、`repository_contract_smoke_runner_implemented` 与 `repository_interface_readiness_defined`。

read-side 程序化证据包括：`control-plane-read-model-v1`（`control-plane-read-model-v1.json` / `check-control-plane-read-model-v1.py`）、`control-plane-read-route-contract-v1`（`control-plane-read-route-contract-v1.json` / `check-control-plane-read-route-contract-v1.py`）、`control-plane-read-response-fixtures-v1`（`control-plane-read-response-fixtures-v1.json` / `check-control-plane-read-response-fixtures-v1.py`）、`control-plane-read-negative-contract-v1`（`control-plane-read-negative-contract-v1.json` / `check-control-plane-read-negative-contract-v1.py`）、`control-plane-read-implementation-preconditions-v1`（`control-plane-read-implementation-preconditions-v1.json` / `check-control-plane-read-implementation-preconditions-v1.py`，固定 handler ownership、fake store 和 negative route smoke readiness）、`control-plane-read-fake-store-handler-plan-v1`（`control-plane-read-fake-store-handler-plan-v1.json` / `check-control-plane-read-fake-store-handler-plan-v1.py`）、`control-plane-read-fake-store-handler-implementation-v1`（`control-plane-read-fake-store-handler-implementation-v1.json` / `check-control-plane-read-fake-store-handler-implementation-v1.py`，固定 fake-store-backed read handler implementation）、`control-plane-read-auth-db-preconditions-v1`（`control-plane-read-auth-db-preconditions-v1.json` / `check-control-plane-read-auth-db-preconditions-v1.py`）、`control-plane-read-consumer-contract-v1`（`control-plane-read-consumer-contract-v1.json` / `check-control-plane-read-consumer-contract-v1.py`）、`control-plane-read-formal-ui-boundary-v1.json` / `check-control-plane-read-formal-ui-boundary-v1.py`、`control-plane-read-formal-ui-implementation-readiness-v1.json` / `check-control-plane-read-formal-ui-implementation-readiness-v1.py`、`control-plane-read-shared-shell-v1.json` / `check-control-plane-read-shared-shell-v1.py`、`control-plane-read-admin-tenant-overview-v1.json` / `check-control-plane-read-admin-tenant-overview-v1.py`、`control-plane-read-workspace-applications-v1.json` / `check-control-plane-read-workspace-applications-v1.py`、`control-plane-read-workspace-api-keys-v1.json` / `check-control-plane-read-workspace-api-keys-v1.py`、`control-plane-read-workspace-usage-quota-v1.json` / `check-control-plane-read-workspace-usage-quota-v1.py`、`control-plane-read-workspace-workflow-definitions-v1.json` / `check-control-plane-read-workspace-workflow-definitions-v1.py`、`control-plane-read-workspace-run-history-v1.json` / `check-control-plane-read-workspace-run-history-v1.py`、`control-plane-read-admin-audit-log-v1.json` / `check-control-plane-read-admin-audit-log-v1.py`、`control-plane-read-formal-ui-readiness-close-v1.json` / `check-control-plane-read-formal-ui-readiness-close-v1.py`、`control-plane-read-dev-live-consumer-v1.json` / `check-control-plane-read-dev-live-consumer-v1.py`、`control-plane-read-auth-store-transition-preconditions-v1.json` / `check-control-plane-read-auth-store-transition-preconditions-v1.py`、`control-plane-read-repository-contract-preconditions-v1.json` / `check-control-plane-read-repository-contract-preconditions-v1.py`、`control-plane-read-disabled-database-guard-v1.json` / `check-control-plane-read-disabled-database-guard-v1.py`、`control-plane-read-repository-contract-smoke-v1.json` / `check-control-plane-read-repository-contract-smoke-v1.py`、`control-plane-read-repository-implementation-readiness-v1.json` / `check-control-plane-read-repository-implementation-readiness-v1.py`、`control-plane-read-store-selection-readiness-v1.json` / `check-control-plane-read-store-selection-readiness-v1.py`、`control-plane-read-schema-migration-readiness-v1.json` / `check-control-plane-read-schema-migration-readiness-v1.py`、`control-plane-read-repository-contract-types-readiness-v1.json` / `check-control-plane-read-repository-contract-types-readiness-v1.py`、`control-plane-read-repository-contract-types-implementation-v1.json` / `check-control-plane-read-repository-contract-types-implementation-v1.py`、`control-plane-read-repository-contract-smoke-runner-readiness-v1.json` / `check-control-plane-read-repository-contract-smoke-runner-readiness-v1.py`、`control-plane-read-repository-contract-smoke-runner-implementation-v1.json` / `check-control-plane-read-repository-contract-smoke-runner-implementation-v1.py` 与 `control-plane-read-repository-interface-readiness-v1.json` / `check-control-plane-read-repository-interface-readiness-v1.py`。

`control-plane-read-repository-implementation-readiness-v1` 已把未来 repository implementation readiness 固定为 `repository_implementation_readiness_defined`：只定义未来文件落点、实现准入 gate、七条 route readiness matrix、dual smoke plan、failure mapping、no fake fallback、no side effects 和停止线，不创建 Go repository 文件、不实现 repository interface / adapter、不写 SQL、不建 migration、不接真实数据库、Radish OIDC、token validation 或 production API consumer。程序化证据为 `control-plane-read-repository-implementation-readiness-v1.json` / `check-control-plane-read-repository-implementation-readiness-v1.py`。

`control-plane-read-store-selection-readiness-v1` 已把未来 store selection readiness 固定为 `store_selection_readiness_defined`：只定义默认 read source、保留 read source、失败映射、七条 route selection matrix、no fake fallback、no side effects 和停止线，不创建正式配置入口、不实现 store selector、不实现 repository interface / adapter、不写 SQL、不建 migration、不接真实数据库、Radish OIDC、token validation 或 production API consumer。程序化证据为 `control-plane-read-store-selection-readiness-v1.json` / `check-control-plane-read-store-selection-readiness-v1.py`。

`control-plane-read-schema-migration-readiness-v1` 已把未来 schema migration readiness 固定为 `schema_migration_readiness_defined`：只定义 schema ownership、migration layout、rollback plan、tenant index strategy、read-only role policy、migration smoke、failure mapping 和停止线，不创建 migration 目录、不写 SQL、不实现 migration runner、store selector、repository interface / adapter、真实数据库、Radish OIDC、token validation 或 production API consumer。程序化证据为 `control-plane-read-schema-migration-readiness-v1.json` / `check-control-plane-read-schema-migration-readiness-v1.py`。

`control-plane-read-repository-contract-types-readiness-v1` 已把未来 repository contract types readiness 固定为 `repository_contract_types_readiness_defined`：只定义 `ReadRepositoryContext`、七条 read route request / result type、failure code type、projection / filter / sort type 和 contract smoke type 输入，不创建 `control_plane_read_repository_contract.go`，不实现 repository interface / adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。程序化证据为 `control-plane-read-repository-contract-types-readiness-v1.json` / `check-control-plane-read-repository-contract-types-readiness-v1.py`。

`control-plane-read-repository-contract-types-implementation-v1` 已把 repository contract types implementation 固定为 `repository_contract_types_implemented`：已创建 `services/platform/internal/httpapi/control_plane_read_repository_contract.go` 与对应 Go 单元测试，落地 `ReadRepositoryContext`、七条 read route request / result type、failure code、projection / filter / sort type 和 route type matrix；不声明 `ControlPlaneReadRepository` interface，不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。程序化证据为 `control-plane-read-repository-contract-types-implementation-v1.json` / `check-control-plane-read-repository-contract-types-implementation-v1.py`。

`control-plane-read-repository-contract-smoke-runner-readiness-v1` 已把未来 repository contract smoke runner readiness 固定为 `repository_contract_smoke_runner_readiness_defined`：只定义 `ControlPlaneReadRepositoryContractSmokeRunner` 未来如何消费 `controlPlaneReadRepositoryRouteTypeContracts()`、既有 smoke fixture、failure mapping、no fake fallback 和 no side effects，不创建 runner 文件、不实现 repository interface / adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。程序化证据为 `control-plane-read-repository-contract-smoke-runner-readiness-v1.json` / `check-control-plane-read-repository-contract-smoke-runner-readiness-v1.py`。

`control-plane-read-repository-contract-smoke-runner-implementation-v1` 已把 repository contract smoke runner implementation 固定为 `repository_contract_smoke_runner_implemented`：只创建静态 Go runner 与测试，消费 `controlPlaneReadRepositoryRouteTypeContracts()` 并与既有 smoke fixture 对齐七条 read route、failure mapping、no fake fallback 和 no side effects，不声明 repository interface / adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。程序化证据为 `control-plane-read-repository-contract-smoke-runner-implementation-v1.json` / `check-control-plane-read-repository-contract-smoke-runner-implementation-v1.py`。

`control-plane-read-repository-interface-readiness-v1` 已把 repository interface readiness 固定为 `repository_interface_readiness_defined`：只定义未来 `ControlPlaneReadRepository` method matrix、adapter implementation gate、production auth gate、failure mapping 和 no side effects，消费已落地 Go contract type 与静态 runner 证据；不创建 interface 文件、不声明 repository interface / adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。程序化证据为 `control-plane-read-repository-interface-readiness-v1.json` / `check-control-plane-read-repository-interface-readiness-v1.py`。
2. `Provider Runtime & Health v1`：已有 [任务卡](task-cards/provider-runtime-health-v1-plan.md)，并已完成 `provider-capability-matrix-v1`、`provider-health-smoke-v1`、`provider-selection-policy-v1`、`provider-retry-fallback-policy-v1` 和 `provider-runtime-docs-refresh` 五个 gate。证据包括 `provider-capability-matrix-v1.json`、`check-provider-capability-matrix.py`、`provider-health-smoke-v1.json`、`check-provider-health-smoke.py`、`provider-selection-policy-v1.json`、`check-provider-selection-policy.py`、`provider-retry-fallback-policy-v1.json`、`check-provider-retry-fallback-policy.py`、`provider-runtime-docs-refresh.json` 和 `check-provider-runtime-docs-refresh.py`。五者均已接入 fast baseline，Provider Runtime & Health v1 进入 close candidate；下一步不继续默认新增 provider 同层小切片，只有明确任务窗口时才单独推进 optional live health、retry/fallback execution、production secret backend 或 container smoke。默认不联网、不要求真实 credential、不下载模型，也不把 provider health 写成 production readiness。
3. `Production Ops Hardening v1`：状态调整为 `static boundary close + docker_local smoke recorded`。已完成 `config-secret-boundary`、`production-secret-backend-contract`、`production-secret-backend-implementation-readiness`、`secret-ref-schema-and-fixtures`、`startup-supervisor-boundary`、`environment-isolation`、`console-production-package-smoke`、P3 checklist alignment、`docker-deployment-mode-definition`、`docker-local-compose`、`docker-test-prod-compose`、`docker-image-build-publish`、`deployment-readiness-smoke`、`container-smoke-runbook` 和 `container-smoke-record-template`；证据包括 `production-ops-config-secret-boundary.json`、`production-ops-secret-backend-contract.json`、`production-ops-secret-backend-implementation-readiness.json`、`production-secret-backend-implementation-v1-plan.md`、`production-secret-reference.schema.json`、`production-secret-reference-basic.json`、`production-ops-startup-supervisor-boundary.json`、`production-ops-environment-isolation-boundary.json`、`production-ops-console-package-smoke.json`、`production-ops-docker-deployment-mode.json`、`production-ops-docker-local-compose.json`、`production-ops-docker-test-prod-compose.json`、`production-ops-docker-image-build-publish.json`、`production-ops-deployment-readiness-smoke.json`、`production-ops-container-smoke-runbook.json` 和 `production-ops-container-smoke-record-template.json`。部署态资产为 `deploy/docker-compose.yaml` 与 `deploy/.env.example`，镜像 tag 后缀固定为 `v*-dev`、`v*-test`、`v*-release`，deployment readiness 仅覆盖 `docker compose config` 静态展开；`container_smoke_ready` 已有一次 `docker_local` 可审计运行记录，测试环境 smoke、生产前复核、真实镜像发布、production secret backend、正式 auth / CORS policy 和 process supervisor 仍为 `not_satisfied`。
4. `Model Adaptation` P4 前置计划：P4 v1 runbook、治理复核记录和本地 `Qwen2.5-1.5B-Instruct` full-holdout-9 预检结果已落地。当前结论是 raw student 总体 `schema_valid_rate=0.6667`、`radishflow/suggest_flowsheet_edits` 3/3 因 `citations` scaffold 引用泄漏 blocked，ghost/docs 两类任务 6/6 通过；`--repair-hard-fields` comparison 可达到 9/9 schema/task valid，但只能作为后处理证据，不是 raw 晋级或训练准入。后续 `Qwen2.5-3B-Instruct` raw CPU 单样本 probe 在已知 edits blocker 上 300 秒 timeout 且未产出 token。真实模型产出、3B/4B 长跑、训练 JSONL、蒸馏和权重相关工作转入后置专题。
5. `P3 Local Product Shell / Ops Surface`、`Conversation & Session`、`Tooling Framework` 与 `UI Design Topic / React 第二批` 均保持 close / governance-only 状态，不再默认新增同类只读 UI、P2 readiness / rollup / manifest 或 UI polish。P2 停止线证据继续保留为 `session-tooling-readiness-summary.json`、`session-tooling-foundation-status-summary.json`、`session-tooling-negative-regression-suite-readiness.json`、`session-tooling-close-candidate-readiness-rollup.json`、`session-tooling-negative-coverage-rollup.json`、`session-tooling-route-negative-coverage-matrix.json`、`session-tooling-route-smoke-readiness-rollup.json`、`session-tooling-short-close-readiness-delta.json`、`session-tooling-readiness-consistency-rollup.json`、`session-tooling-executor-storage-confirmation-enablement-plan.json`、`session-tooling-stop-line-manifest.json`、`session-tooling-upper-layer-confirmation-flow-readiness.json` 与 `session-tooling-short-close-entry-checklist.json`；`P2 short close` 边界不变，相关 `negative_regression_suite` 边界不变，这些 fixture 不代表 executor、durable store、confirmation、materialized result reader、长期记忆或 replay 已完成。
6. `Evaluation & Governance`：当前阶段门禁已从“每个小 UI 展示项新增 fixture / checker / task card”调整为“能力边界与聚合门禁优先”。`control-plane-read-formal-ui-readiness-close-v1` 已作为聚合 surface matrix 接入 fast baseline，后续普通 UI 展示和文案 / 布局改动优先复用 web build、consumer smoke、聚合 read-side checker 和 fast baseline；只有新增 API、执行边界、生产声明、数据格式、外部 provider 风险或高风险能力时才新增专项门禁。

## 明天事项（2026-06-04）

1. 先按协作约定检查 `git status`，读取本文档、[能力矩阵](radishmind-capability-matrix.md)、[路线图](radishmind-roadmap.md)、[Control Plane Read-Side 契约](contracts/control-plane-read-side.md) 和本周周志。
2. 先确认 2026-06-03 的 read-side repository/type/runner/interface 提交均在本地：`f448ca3`、`a5ca57c`、`e128f7e`、`99a47a0` 与 `5852be3`。
3. 不继续默认新增 provider 同层小切片，也不重复跑同一 docker_local smoke，除非有新的失败假设或镜像 / compose 变更。
4. `shared-read-shell`、七个只读页面、formal UI readiness close、dev-only live read consumer 与 auth/store transition preconditions 已完成；明天不要把这些 gate 误读成数据库、OIDC、完整正式 UI、production API consumer 或 production ready。
5. `control-plane-read-repository-contract-preconditions-v1` 已完成：只定义未来 read store repository interface、tenant predicate、sanitized projection、cursor/filter/sort allowlist、contract smoke 和 failure mapping，不写 SQL、不建 migration、不实现 repository、不接真实数据库。
6. `control-plane-read-disabled-database-guard-v1` 已完成：database / postgres / repository read mode 当前仍是 reserved disabled，误请求必须返回 `database_read_disabled`，不得静默回退到 fake store。
7. `control-plane-read-repository-contract-smoke-v1` 已完成：未来 repository contract smoke 的输入输出、七条 read route 覆盖、failure mapping、no fake fallback、no side effects 和文档停止线已固定；它不代表 smoke runner、SQL、migration、repository adapter、真实数据库、Radish OIDC、token validation 或 production API consumer 已实现。
8. `control-plane-read-repository-implementation-readiness-v1` 已完成：repository implementation readiness 已固定，但不代表 repository interface、adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer 已实现。
9. `control-plane-read-store-selection-readiness-v1` 已完成：store selection readiness 已固定，但不代表正式配置入口、store selector、repository interface、adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer 已实现。
10. `control-plane-read-schema-migration-readiness-v1` 已完成：schema migration readiness 已固定，但不代表 migration 目录、SQL、migration runner、database schema、真实数据库、repository adapter、Radish OIDC、token validation 或 production API consumer 已实现。
11. `control-plane-read-repository-contract-types-readiness-v1` 已完成：repository contract types readiness 已固定，并已被后续 implementation gate 消费。
12. `control-plane-read-repository-contract-types-implementation-v1` 已完成：Go contract type 文件已创建，但不代表 repository interface、adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer 已实现。
13. `control-plane-read-repository-contract-smoke-runner-readiness-v1` 已完成：未来 smoke runner 如何消费 Go type matrix 和既有 smoke contract 已固定，且已被后续 implementation gate 消费。
14. `control-plane-read-repository-contract-smoke-runner-implementation-v1` 已完成：静态 contract/type runner 已实现，但不代表 repository interface、adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer 已实现。
15. `control-plane-read-repository-interface-readiness-v1` 已完成：未来 interface method matrix 已固定，但不代表 repository interface、adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer 已实现。
16. 下一步如果继续 read-side，按整体计划优先推进 `control-plane-read-repository-adapter-implementation-readiness-refresh-v1`：目标是把未来 adapter 实现准入、文件落点、fixture/checker 依赖、static runner 证据和 interface method matrix 对齐清楚；仍不创建 adapter/interface 文件，不实现 SQL、migration、store selector、真实数据库、OIDC token validation、API key lifecycle 或 quota enforcement。
17. 如果明天判断 adapter readiness refresh 还缺 store selector 前置条件，可先推进 `control-plane-read-store-selector-enablement-preconditions-v1` 作为备选前置工作；它只能固定 selector enablement gates、配置禁用态、failure mapping 和 no fake fallback，不能同时接入真实数据库、repository adapter、OIDC、production consumer 或 workflow executor。
18. `control-plane-read-formal-ui-readiness-close-v1`、`control-plane-read-dev-live-consumer-v1`、`control-plane-read-auth-store-transition-preconditions-v1`、`control-plane-read-repository-contract-preconditions-v1`、`control-plane-read-disabled-database-guard-v1`、`control-plane-read-repository-contract-smoke-v1`、`control-plane-read-repository-implementation-readiness-v1`、`control-plane-read-store-selection-readiness-v1`、`control-plane-read-schema-migration-readiness-v1`、`control-plane-read-repository-contract-types-readiness-v1`、`control-plane-read-repository-contract-types-implementation-v1`、`control-plane-read-repository-contract-smoke-runner-readiness-v1`、`control-plane-read-repository-contract-smoke-runner-implementation-v1` 与 `control-plane-read-repository-interface-readiness-v1` 仍只围绕 fake-store-backed handler、测试身份上下文和迁移前置条件，不接真实数据库、Radish OIDC、token validation、repository migration、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
19. read-side UI 相关切片仍不直接实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay；产品机会池仍只作为候选，不写成 roadmap commitment。

## 为什么是这些任务

- P3 本地只读产品壳已经能被开发者实际读取、排障和复验，继续补同类小面板会降低边际收益。
- 上层项目目前没有真实挂载点、确认流和命令承接接口，继续细化接线设计收益很低。
- 仓库里已经有 runtime、gateway、adapter、eval 和 governance 资产，可以先把平台骨架做完整。
- `provider registry`、northbound bridge、本地 wrapper、config layering、diagnostics、deployment smoke、request observability 和 error taxonomy 已经给出 P1 short close 基础；继续在同一层增加更多别名、兜底和配置分支会开始降低边际收益。
- 平台表层语言边界已固定为：`UI` 用 `React + Vite + TypeScript`，平台服务层用 `Go`，模型侧继续保留 `Python`，所有层只共享 `contracts/` 里的 canonical protocol。
- 长期维护优先级高于短期技术栈对齐；不因参考 Radish 而引入 `.NET`，除非未来证明复用 Radish 后端包或共发布能降低系统总复杂度。
- 正式产品面已经扩大为用户端、管理端、模型网关和工作流运行时，但当前实现仍应先复用既有 runtime、provider、ops、secret 和治理资产，不把产品愿景写成已完成能力。
- 平台服务层当前已经有最小 `Go HTTP` 壳、`/healthz`、`/v1/platform/overview`、`/v1/models`、`/v1/chat/completions`、`/v1/responses` 与 `/v1/messages` bridge，并能解释一次请求命中了哪个 route、provider/profile、model、耗时和失败边界；overview console consumer smoke 与最小本地 console 壳已能消费只读产品面；后续不再回头把模型逻辑写回 `Go`。
- 真实模型产出已经暴露出当前本机 CPU 路径的成本边界，Production Ops 静态边界也已足够可检查；Provider Runtime & Health v1 也已进入 close candidate，后续应转向明确运行窗口或 control plane / user workspace / workflow 边界主线，而不是继续扩同层小切片。

## 当前已有可直接利用的基础

- `scripts/run-copilot-inference.py`
- `services/gateway/copilot_gateway.py`
- `scripts/check-radishflow-service-smoke-matrix.py`
- `services/runtime/inference_provider.py`
- `scripts/run-radishflow-gateway-demo.py`
- `scripts/run-radishmind-core-candidate.py`
- `scripts/build-copilot-training-samples.py`

## 默认不要做

- 不继续加长同一批 prompt/scaffold 当作默认推进。
- 不再默认补 P3 console 同类只读小切片；发现真实使用缺口时再补。
- `control-plane-read-formal-ui-readiness-close-v1` 之后不再默认为普通 read-only UI 展示页逐项新增 task card、fixture 和 checker；优先使用聚合 read-side UI surface matrix。
- 不在 tool registry v1 后立即接真实工具执行器、长期记忆或新的 provider/model 实验。
- 不在 recovery checkpoint v1 后立即实现跨轮自动 replay。
- 不把 checkpoint read route smoke 升级为 durable checkpoint store、materialized result reader 或跨轮 replay executor。
- 不扩 `RadishFlow` 同类真实 capture，除非先写清楚非重复 drift 假设。
- 不把 `RadishCatalyst` 从文档预留提前扩成真实 schema、adapter 或 gateway smoke。
- 不在 runtime、session、tooling 契约还没稳定前启动训练放量。
- 不在设计任务卡停止线外扩当前 console：第二批 React 只做 ops surface 结构重排，不新增 API、真实工具执行器、confirmation、业务写回、replay 或 production packaging。
- 不默认继续真实模型产出、3B/4B 长跑、训练 JSONL、蒸馏或权重相关工作；这些内容后续作为独立专题重开。
- 不默认下载大于当前决策所需范围的模型、数据集或权重。
- 不把真实模型输出、训练 JSONL 或大体积实验产物提交入仓。
- 不直接修改 `RadishFlow`、`Radish` 或 `RadishCatalyst` 外部工作区。
- 不再默认新增 Production Ops 同类静态治理切片；除非有明确 Docker 运行窗口，否则不要把 provider runtime 同层小切片当作默认下一步。
- 不把 dev-only live read path 或 auth/store transition preconditions 放宽为真实数据库、Radish OIDC、token validation、repository migration、repository implementation、production API consumer、API key lifecycle、quota enforcement、billing、workflow executor、confirmation、writeback 或 replay。

## 最小读取路径

回答“今天做什么”时，默认读取：

1. `AGENTS.md` 或 `CLAUDE.md`
2. `docs/README.md`
3. `docs/radishmind-current-focus.md`
4. `docs/radishmind-capability-matrix.md`
5. 必要时读取 `docs/radishmind-roadmap.md`

## 按需读取

- 动产品边界：读 `docs/radishmind-product-scope.md`
- 动架构或服务分层：读 `docs/radishmind-architecture.md`
- 动协议、schema 或 API：读 `docs/radishmind-integration-contracts.md`
- 动 `RadishMind-Core` 评测：读 `docs/radishmind-core-baseline-evaluation.md`
- 动代码风格、抽象或脚本组织：读 `docs/radishmind-code-standards.md`
- 查最近执行细节：读本周周志 `docs/devlogs/2026-W22.md`

## 验证基线

文档或治理改动完成后，macOS / Linux / WSL 环境优先执行：

```bash
./scripts/bootstrap-dev.sh
./scripts/check-repo.sh --fast
```

Windows / PowerShell 环境使用：

```powershell
pwsh ./scripts/bootstrap-dev.ps1
pwsh ./scripts/check-repo.ps1 -Fast
```
