# RadishMind 当前推进焦点

更新时间：2026-05-26

## 文档目的

本文档用于回答“根据项目规划和开发进度，今天要做什么以推进开发”。它是新会话的短入口，不承载历史细节、完整实验日志或长契约说明。

默认只读本文件仍应能判断下一步方向；只有准备实施具体改动时，才跳转到下方按需文档。

## 当前阶段

当前已经从“M3/M4 收口后的被动等待”切换到“平台重定义后的平台本体建设期”；`P1 Runtime Foundation` 已达到 short close，不再默认继续横向细磨 provider/config/diagnostics/observability 同一层：

- `M3` 的 gateway、service smoke、UI consumption 与 candidate handoff 继续作为冻结门禁保留。
- `M4` 的 broader 15/15 `reviewed_pass` 与 `3B/4B` guided capacity review 继续作为路线证据保留。
- 当前不继续扩同一批 `M4` 实验，不提前设计不存在的上层真实接线。
- 平台请求级观测和错误分类短收口已进入 `Go` northbound 层与平台单元测试，`P2 Session & Tooling Foundation` 也已从治理链路推进到最小 metadata / blocked 产品外壳。

`P3 Local Product Shell / Ops Surface` 的本地只读产品壳已经达到 `local usable / read-only close`：overview、local-smoke、console dev entry、Dev Diagnostics、Local Readiness、failure surface、Stop-line Details、Provider/Profile Details 和对应轻量门禁已形成可复验闭环。后续不再默认继续补同类只读 console 小切片，除非真实使用暴露新的可观测缺口。

`Production Ops Hardening v1` 的静态边界已经达到 close：production config / secret boundary、production-secret-backend-contract、startup / supervisor boundary、environment isolation、console production package smoke、docker local/test/prod compose、镜像命名治理、deployment readiness 静态 smoke、container smoke runbook 和 record template 都已可检查。`production-secret-backend-contract` 已落地为 `scripts/checks/fixtures/production-ops-secret-backend-contract.json` 与 `scripts/check-production-ops-secret-backend-contract.py`，只定义未来 external secret backend adapter contract、secret reference、脱敏输出和禁止项；它不实现真实云 secret 服务、不写入真实 secret、不声明 production ready。2026-05-26 已在明确 Docker 运行窗口下完成一次 `docker_local` container smoke，运行记录写入 `tmp/production-ops/container-smoke/20260526T2006-local-docker-smoke.json`；该记录只证明本地 mock provider 容器 smoke 跑通，不声明测试环境 smoke、生产前复核或 production ready。P3 的 production secret backend、process supervisor、deployment environment isolation 和 console production packaging 仍保持未完成，但它们不再阻塞下一条平台主线。

P3 checklist 继续由 `p3-local-product-shell-short-close-checklist.json` 与 `check-p3-local-product-shell-short-close-checklist.py` 固定 production secret backend、process supervisor、部署环境隔离和 console production packaging 仍为 `not_satisfied`。

`UI Design Topic / Pencil Draft` 已完成正式 React 第二批前的主要设计覆盖：`docs/designs/radishmind-console-ops-surface-v0.pen` 已包含 ready、failure/stale、narrow、loading/empty、settings/permissions、blocked action detail 与 token mapping notes。2026-05-23 Pencil `snapshot_layout` 已无 layout problems，`apps/radishmind-console/` 第二批 ops surface 结构重排已进入 `close candidate`，本地 mock platform ready 态和桌面 / 窄屏临时截图已复核。第二批只调整信息架构和视觉层级，不新增 API、不接 executor、不做 confirmation、writeback 或 replay。

## 当前优先做什么

1. `Provider Runtime & Health v1`：新增 [任务卡](task-cards/provider-runtime-health-v1-plan.md)，作为下一条平台主线。目标是把现有 provider registry、provider/profile inventory、request-side selection、diagnostics 和 error taxonomy 推进到可检查的 provider capability matrix、provider health smoke、provider selection policy 与 provider retry/fallback policy。`provider-capability-matrix-v1` 已落地为 `scripts/checks/fixtures/provider-capability-matrix-v1.json` 与 `scripts/check-provider-capability-matrix.py`；`provider-health-smoke-v1` 已落地为 `scripts/checks/fixtures/provider-health-smoke-v1.json` 与 `scripts/check-provider-health-smoke.py`；`provider-selection-policy-v1` 已落地为 `scripts/checks/fixtures/provider-selection-policy-v1.json`、`scripts/check-provider-selection-policy.py` 和 Go selection 单元测试；`provider-retry-fallback-policy-v1` 已落地为 `scripts/checks/fixtures/provider-retry-fallback-policy-v1.json`、`scripts/check-provider-retry-fallback-policy.py`、Go audit metadata 和失败路径单元测试，固定 `retry_policy=caller-managed`、`fallback_policy=disabled` 与 no implicit retry/fallback 口径；`provider-runtime-docs-refresh` 已落地为 `scripts/checks/fixtures/provider-runtime-docs-refresh.json` 与 `scripts/check-provider-runtime-docs-refresh.py`。五者均已接入 fast baseline，Provider Runtime & Health v1 进入 close candidate；下一步不继续默认新增 provider 同层小切片，只有明确任务窗口时才单独推进 optional live health、retry/fallback execution、production secret backend 或 container smoke。默认不联网、不要求真实 credential、不下载模型，也不把 provider health 写成 production readiness。
2. `Production Ops Hardening v1`：状态调整为 `static boundary close + docker_local smoke recorded`。已完成 `config-secret-boundary`、`production-secret-backend-contract`、`startup-supervisor-boundary`、`environment-isolation`、`console-production-package-smoke`、P3 checklist alignment、`docker-deployment-mode-definition`、`docker-local-compose`、`docker-test-prod-compose`、`docker-image-build-publish`、`deployment-readiness-smoke`、`container-smoke-runbook` 和 `container-smoke-record-template`；证据包括 `production-ops-config-secret-boundary.json`、`production-ops-secret-backend-contract.json`、`production-ops-startup-supervisor-boundary.json`、`production-ops-environment-isolation-boundary.json`、`production-ops-console-package-smoke.json`、`production-ops-docker-deployment-mode.json`、`production-ops-docker-local-compose.json`、`production-ops-docker-test-prod-compose.json`、`production-ops-docker-image-build-publish.json`、`production-ops-deployment-readiness-smoke.json`、`production-ops-container-smoke-runbook.json` 和 `production-ops-container-smoke-record-template.json`。部署态资产为 `deploy/docker-compose.yaml` 与 `deploy/.env.example`，镜像 tag 后缀固定为 `v*-dev`、`v*-test`、`v*-release`，deployment readiness 仅覆盖 `docker compose config` 静态展开；`container_smoke_ready` 已有一次 `docker_local` 可审计运行记录，测试环境 smoke、生产前复核、真实镜像发布、production secret backend、正式 auth / CORS policy 和 process supervisor 仍为 `not_satisfied`。
3. `Model Adaptation` P4 前置计划：P4 v1 runbook、治理复核记录和本地 `Qwen2.5-1.5B-Instruct` full-holdout-9 预检结果已落地。当前结论是 raw student 总体 `schema_valid_rate=0.6667`、`radishflow/suggest_flowsheet_edits` 3/3 因 `citations` scaffold 引用泄漏 blocked，ghost/docs 两类任务 6/6 通过；`--repair-hard-fields` comparison 可达到 9/9 schema/task valid，但只能作为后处理证据，不是 raw 晋级或训练准入。后续 `Qwen2.5-3B-Instruct` raw CPU 单样本 probe 在已知 edits blocker 上 300 秒 timeout 且未产出 token。真实模型产出、3B/4B 长跑、训练 JSONL、蒸馏和权重相关工作转入后置专题。
4. `P3 Local Product Shell / Ops Surface`、`Conversation & Session`、`Tooling Framework` 与 `UI Design Topic / React 第二批` 均保持 close / governance-only 状态，不再默认新增同类只读 UI、P2 readiness / rollup / manifest 或 UI polish。P2 停止线证据继续保留为 `session-tooling-readiness-summary.json`、`session-tooling-foundation-status-summary.json`、`session-tooling-negative-regression-suite-readiness.json`、`session-tooling-close-candidate-readiness-rollup.json`、`session-tooling-negative-coverage-rollup.json`、`session-tooling-route-negative-coverage-matrix.json`、`session-tooling-route-smoke-readiness-rollup.json`、`session-tooling-short-close-readiness-delta.json`、`session-tooling-readiness-consistency-rollup.json`、`session-tooling-executor-storage-confirmation-enablement-plan.json`、`session-tooling-stop-line-manifest.json`、`session-tooling-upper-layer-confirmation-flow-readiness.json` 与 `session-tooling-short-close-entry-checklist.json`；`P2 short close` 边界不变，相关 `negative_regression_suite` 边界不变，这些 fixture 不代表 executor、durable store、confirmation、materialized result reader、长期记忆或 replay 已完成。
5. `Evaluation & Governance`：后续门禁从“每个小 UI 展示项新增 fixture”放宽为“聚合门禁优先”。只有新增 API、执行边界、生产声明、数据格式、外部 provider 风险或高风险能力时才新增专项门禁；普通 UI 展示改动优先复用 console behavior / visual smoke / fast baseline。

## 明天事项（2026-05-27）

1. 先按协作约定检查 `git status`，读取本文档、[能力矩阵](radishmind-capability-matrix.md)、[路线图](radishmind-roadmap.md) 和本周周志，确认 2026-05-26 本地 container smoke 记录仍是最新运行事实。
2. 不继续默认新增 provider 同层小切片，也不重复跑同一 docker_local smoke，除非有新的失败假设或镜像 / compose 变更。
3. 可选独立重开方向只保留四类：optional live health 手动 runbook、retry/fallback execution、真实 production secret backend 实现、test smoke / production preflight 运行记录。任何一类都必须先写清楚输入、输出、停止线和验证方式；production secret backend 方案当前只完成 contract governance，不实现真实云 secret 服务。
4. 继续保持 P2 / P3 / UI / P4 的停止线：不接真实 executor、confirmation、writeback、replay，不重开真实模型长跑，不把 provider health、local-smoke 或 Docker 静态边界写成 production ready。

## 为什么是这些任务

- P3 本地只读产品壳已经能被开发者实际读取、排障和复验，继续补同类小面板会降低边际收益。
- 上层项目目前没有真实挂载点、确认流和命令承接接口，继续细化接线设计收益很低。
- 仓库里已经有 runtime、gateway、adapter、eval 和 governance 资产，可以先把平台骨架做完整。
- `provider registry`、northbound bridge、本地 wrapper、config layering、diagnostics、deployment smoke、request observability 和 error taxonomy 已经给出 P1 short close 基础；继续在同一层增加更多别名、兜底和配置分支会开始降低边际收益。
- 平台表层语言边界已固定为：`UI` 用 `React + Vite + TypeScript`，平台服务层用 `Go`，模型侧继续保留 `Python`，所有层只共享 `contracts/` 里的 canonical protocol。
- 平台服务层当前已经有最小 `Go HTTP` 壳、`/healthz`、`/v1/platform/overview`、`/v1/models`、`/v1/chat/completions`、`/v1/responses` 与 `/v1/messages` bridge，并能解释一次请求命中了哪个 route、provider/profile、model、耗时和失败边界；overview console consumer smoke 与最小本地 console 壳已能消费只读产品面；后续不再回头把模型逻辑写回 `Go`。
- 真实模型产出已经暴露出当前本机 CPU 路径的成本边界，Production Ops 静态边界也已足够可检查；Provider Runtime & Health v1 也已进入 close candidate，后续应转向明确运行窗口或下一条平台主线，而不是继续扩同层小切片。

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
- 查最近执行细节：读本周周志 `docs/devlogs/2026-W21.md`

## 验证基线

文档或治理改动完成后，优先执行：

```bash
./scripts/check-repo.sh --fast
```

Windows / PowerShell 环境使用：

```powershell
pwsh ./scripts/check-repo.ps1 -Fast
```
