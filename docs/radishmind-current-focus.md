# RadishMind 当前推进焦点

更新时间：2026-05-16

## 文档目的

本文档用于回答“根据项目规划和开发进度，今天要做什么以推进开发”。它是新会话的短入口，不承载历史细节、完整实验日志或长契约说明。

默认只读本文件仍应能判断下一步方向；只有准备实施具体改动时，才跳转到下方按需文档。

## 当前阶段

当前已经从“M3/M4 收口后的被动等待”切换到“平台重定义后的平台本体建设期”；`P1 Runtime Foundation` 已达到 short close，不再默认继续横向细磨 provider/config/diagnostics/observability 同一层：

- `M3` 的 gateway、service smoke、UI consumption 与 candidate handoff 继续作为冻结门禁保留。
- `M4` 的 broader 15/15 `reviewed_pass` 与 `3B/4B` guided capacity review 继续作为路线证据保留。
- 当前不继续扩同一批 `M4` 实验，不提前设计不存在的上层真实接线。
- 平台请求级观测和错误分类短收口已进入 `Go` northbound 层与平台单元测试，主要实现重心可以切到 `P2 Session & Tooling Foundation`。

如果今天继续写代码，默认进入 `Conversation & Session` 与 `Tooling Framework`，而不是继续重跑模型、扩同层配置兜底或补想象中的接入细节。

## 当前优先做什么

1. `Conversation & Session`：首版 session contract、history policy、state policy、recovery record、recovery checkpoint record/manifest/read result、fixture、northbound session metadata 和平台 metadata-only route smoke 已开始落地；checkpoint read route 已通过负向 fixture 固定 materialized result、result ref、executor ref、durable memory 与 replay 类查询参数拒绝口径。`scripts/checks/fixtures/session-tooling-readiness-summary.json` 已把当前已完成门禁和 storage / replay 停止线收口为可检查 summary；`docs/task-cards/session-tooling-implementation-preconditions.md` 已把 executor / storage / confirmation 的实现前置条件拆开。`scripts/checks/fixtures/session-tooling-negative-regression-suite.json` 已把 9 个 skeleton case 收口为 governance-only suite 消费，固定治理消费者、audit non-write 边界、side-effect absence 断言，并对齐 `scripts/checks/fixtures/session-tooling-deny-by-default-implementation-gates.json` 中三类默认拒绝 implementation gate contract；`scripts/checks/fixtures/session-tooling-negative-coverage-rollup.json` 已把 route smoke、负向 fixture consumer、governance suite 和 deny-by-default gate contract 的覆盖关系收口为 governance-only 可检查 rollup；`scripts/checks/fixtures/session-tooling-route-negative-coverage-matrix.json` 已把 9 个 negative suite case 与当前 checkpoint read metadata-only route、future route requirement 的覆盖矩阵收口为可检查事实，其中 2 个 case 当前 route-covered、7 个仍是 governance-only 且 future requirement 为 `not_satisfied`；`scripts/checks/fixtures/session-tooling-route-smoke-readiness-rollup.json` 已把当前 checkpoint read metadata-only route smoke 与未来 executor / storage / confirmation route smoke 缺口收口为 `not_satisfied` 报告；`scripts/checks/fixtures/session-tooling-readiness-consistency-rollup.json` 已把 close-candidate readiness rollup、route smoke readiness rollup、short close readiness delta、negative coverage rollup 与 `negative_regression_suite` readiness 的 status、source reference、`not_satisfied` hard prerequisite 和覆盖计数收口为 drift 检查；`scripts/checks/fixtures/session-tooling-executor-storage-confirmation-enablement-plan.json` 已把 executor / storage / confirmation 进入未来 gated plan 前的证据拆成可检查治理条件，但三类能力仍是 `blocked_not_gated_plan`；`scripts/checks/fixtures/session-tooling-foundation-status-summary.json`、`scripts/checks/fixtures/session-tooling-close-candidate-readiness-rollup.json`、`scripts/checks/fixtures/session-tooling-negative-regression-suite-readiness.json` 与 `scripts/checks/fixtures/session-tooling-short-close-readiness-delta.json` 已把当前状态收口为 P2 close candidate，并把到 `P2 short close` 仍缺的硬前置条件保持为 `not_satisfied`，但只代表 governance-only、deny-by-default gate contract、负向覆盖 rollup、route smoke gap、route negative coverage matrix、完整 `negative_regression_suite` 验收口径、short close delta、readiness drift 和 enablement plan 可检查，不代表 P2 short close 或真实实现完成。下一步仍不引入长期记忆。
2. `Tooling Framework`：最小 tool schema、registry fixture、policy/audit record、session binding、metadata-only result cache 和快速门禁已开始落地；tool audit summary 已进入 checkpoint read metadata smoke，session/tooling promotion gate 分层、负向消费 summary、route smoke coverage summary、readiness summary、implementation preconditions check、negative regression skeleton、governance-only negative regression suite、negative regression suite readiness、deny-by-default implementation gates、negative coverage rollup、route negative coverage matrix、route smoke readiness rollup、readiness consistency rollup、executor/storage/confirmation enablement plan、close candidate status summary、close-candidate readiness rollup、short close readiness delta、confirmation flow design、independent audit records design、result materialization policy design、executor boundary design 和 storage backend design 已进入快速门禁。下一步仍只推进契约和治理边界，不实现真实工具执行器。
3. `Evaluation & Governance`：把已有 schema、offline eval、service smoke、runtime provider dispatch smoke 和 platform config/deployment/diagnostics/request-observability checks 扩展到 session 与 tooling 门禁。
4. `Model Adaptation`：在前三项稳定后再定义首版基座、蒸馏和训练计划；当前不启动训练放量。

## 为什么是这些任务

- 上层项目目前没有真实挂载点、确认流和命令承接接口，继续细化接线设计收益很低。
- 仓库里已经有 runtime、gateway、adapter、eval 和 governance 资产，可以先把平台骨架做完整。
- `provider registry`、northbound bridge、本地 wrapper、config layering、diagnostics、deployment smoke、request observability 和 error taxonomy 已经给出 P1 short close 基础；继续在同一层增加更多别名、兜底和配置分支会开始降低边际收益。
- 平台表层语言边界已固定为：`UI` 用 `React + Vite + TypeScript`，平台服务层用 `Go`，模型侧继续保留 `Python`，所有层只共享 `contracts/` 里的 canonical protocol。
- 平台服务层当前已经有最小 `Go HTTP` 壳、`/healthz`、`/v1/models`、`/v1/chat/completions`、`/v1/responses` 与 `/v1/messages` bridge，并能解释一次请求命中了哪个 route、provider/profile、model、耗时和失败边界；后续不再回头把模型逻辑写回 `Go`。
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
- 不在 tool registry v1 后立即接真实工具执行器、长期记忆或新的 provider/model 实验。
- 不在 recovery checkpoint v1 后立即实现跨轮自动 replay。
- 不把 checkpoint read route smoke 升级为 durable checkpoint store、materialized result reader 或跨轮 replay executor。
- 不扩 `RadishFlow` 同类真实 capture，除非先写清楚非重复 drift 假设。
- 不把 `RadishCatalyst` 从文档预留提前扩成真实 schema、adapter 或 gateway smoke。
- 不在 runtime、session、tooling 契约还没稳定前启动训练放量。
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
- 查最近执行细节：读本周周志 `docs/devlogs/2026-W20.md`

## 验证基线

文档或治理改动完成后，优先执行：

```bash
./scripts/check-repo.sh --fast
```

Windows / PowerShell 环境使用：

```powershell
pwsh ./scripts/check-repo.ps1 -Fast
```
