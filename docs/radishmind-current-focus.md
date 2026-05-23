# RadishMind 当前推进焦点

更新时间：2026-05-23

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

如果今天继续推进，默认进入 `UI Design Topic / Pencil Draft` 或 P4 前置的模型适配目标定义，而不是继续重跑模型、扩同层配置兜底、细化 P2 readiness / rollup / manifest / task card、补 P3 console 小展示项，或补想象中的上层接入细节。

`UI Design Topic / Pencil Draft` 当前可以启动：先基于 [UI 设计规范](radishmind-ui-design-spec.md) 和 [UI 设计参考](radishmind-ui-design-reference.md) 用 `pencil` 绘制 `.pen` 设计稿，定稿本地 console / ops surface 的信息架构、状态层级、只读与可执行边界、错误诊断和窄屏布局，再拆分正式 UI 实现任务。设计稿定稿前，不把当前 console 壳扩成正式产品 UI。

## 当前优先做什么

1. `UI Design Topic / Pencil Draft`：P3 本地只读产品壳已经能说明正式界面要承载的状态，下一步默认启动设计专题。先按 [UI 设计规范](radishmind-ui-design-spec.md) 和 [UI 设计参考](radishmind-ui-design-reference.md) 用 `pencil` 绘制 `.pen` 设计稿，覆盖本地 console / ops surface 的信息架构、状态层级、只读与可执行边界、错误诊断、Provider/Profile inventory、Stop-line Details、Local Readiness 和窄屏布局。设计定稿前不做正式 React UI 重构，不新增确认流、业务写回或 replay UI。
2. `P3 Local Product Shell / Ops Surface`：状态调整为 `local usable / read-only close`。`GET /v1/platform/overview`、`GET /v1/platform/local-smoke`、overview / local-smoke consumer smoke、`apps/radishmind-console/`、Dev Diagnostics、`Local Readiness`、overview / local-smoke failure surface、Stop-line Details、Provider/Profile Details、console dev entry、behavior / visual smoke 和 production boundary gate 已可复验；不再默认继续补同类只读 console 小切片。`scripts/checks/fixtures/p3-local-product-shell-short-close-checklist.json` 与 `scripts/check-p3-local-product-shell-short-close-checklist.py` 继续固定 production hardening 未完成：production secret backend、process supervisor、部署环境隔离和 console production packaging 仍为 `not_satisfied`。
3. `Conversation & Session` 与 `Tooling Framework`：保持 `P2 close candidate / governance-only`，既有 `scripts/checks/fixtures/session-tooling-readiness-summary.json`、`session-tooling-foundation-status-summary.json`、`session-tooling-negative-regression-suite-readiness.json`、`session-tooling-close-candidate-readiness-rollup.json`、`session-tooling-negative-coverage-rollup.json`、`session-tooling-route-negative-coverage-matrix.json`、`session-tooling-route-smoke-readiness-rollup.json`、`session-tooling-short-close-readiness-delta.json`、`session-tooling-readiness-consistency-rollup.json`、`session-tooling-executor-storage-confirmation-enablement-plan.json`、`session-tooling-stop-line-manifest.json`、`session-tooling-upper-layer-confirmation-flow-readiness.json` 与 `session-tooling-short-close-entry-checklist.json` 继续作为 `P2 short close` 停止线证据保留，相关 `not_satisfied`、`negative_regression_suite`、不实现真实工具执行器、不启用 automatic replay 的边界不变，但不再默认新增 readiness、rollup、manifest 或 task card。当前 `GET /v1/session/metadata`、`GET /v1/tools/metadata` 与 `POST /v1/tools/actions` 已足够支撑最小 metadata / blocked shell；下一步只允许作为 P3 overview 或 UI 消费面的一部分被复用。
4. `Model Adaptation`：可以开始定义 P4 前置计划，包括 v1 模型能力目标、teacher/student 边界、样本分层、晋级门槛和训练 runbook；仍不启动训练放量，不下载模型权重，不把 builder / guided / repaired 结果写成 raw 晋级。
5. `Evaluation & Governance`：后续门禁从“每个小 UI 展示项新增 fixture”放宽为“聚合门禁优先”。只有新增 API、执行边界、生产声明、数据格式、外部 provider 风险或高风险能力时才新增专项门禁；普通 UI 展示改动优先复用 console behavior / visual smoke / fast baseline。

## 为什么是这些任务

- P3 本地只读产品壳已经能被开发者实际读取、排障和复验，继续补同类小面板会降低边际收益。
- 上层项目目前没有真实挂载点、确认流和命令承接接口，继续细化接线设计收益很低。
- 仓库里已经有 runtime、gateway、adapter、eval 和 governance 资产，可以先把平台骨架做完整。
- `provider registry`、northbound bridge、本地 wrapper、config layering、diagnostics、deployment smoke、request observability 和 error taxonomy 已经给出 P1 short close 基础；继续在同一层增加更多别名、兜底和配置分支会开始降低边际收益。
- 平台表层语言边界已固定为：`UI` 用 `React + Vite + TypeScript`，平台服务层用 `Go`，模型侧继续保留 `Python`，所有层只共享 `contracts/` 里的 canonical protocol。
- 平台服务层当前已经有最小 `Go HTTP` 壳、`/healthz`、`/v1/platform/overview`、`/v1/models`、`/v1/chat/completions`、`/v1/responses` 与 `/v1/messages` bridge，并能解释一次请求命中了哪个 route、provider/profile、model、耗时和失败边界；overview console consumer smoke 与最小本地 console 壳已能消费只读产品面；后续不再回头把模型逻辑写回 `Go`。
- 如果在 service、session、tooling 边界还没稳定前继续深挖模型实验，容易再次陷入局部优化。

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
- 不在 Pencil UI 设计稿定稿前，把当前本地 console 壳扩成正式产品 UI 或大面积实现复杂交互。
- 不默认下载大于当前决策所需范围的模型、数据集或权重。
- 不把真实模型输出、训练 JSONL 或大体积实验产物提交入仓。
- 不直接修改 `RadishFlow`、`Radish` 或 `RadishCatalyst` 外部工作区。

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
