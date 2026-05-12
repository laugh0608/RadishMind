# RadishMind 当前推进焦点

更新时间：2026-05-11

## 文档目的

本文档用于回答“根据项目规划和开发进度，今天要做什么以推进开发”。它是新会话的短入口，不承载历史细节、完整实验日志或长契约说明。

默认只读本文件仍应能判断下一步方向；只有准备实施具体改动时，才跳转到下方按需文档。

## 当前阶段

当前已经从“M3/M4 收口后的被动等待”切换到“平台重定义后的 `P1` 启动期”：

- `M3` 的 gateway、service smoke、UI consumption 与 candidate handoff 继续作为冻结门禁保留。
- `M4` 的 broader 15/15 `reviewed_pass` 与 `3B/4B` guided capacity review 继续作为路线证据保留。
- 当前不继续扩同一批 `M4` 实验，不提前设计不存在的上层真实接线。

如果今天开始写代码，默认先做 `Runtime Service` 主线里的“provider / protocol compatibility”，而不是继续重跑模型或补想象中的接入细节。

## 当前优先做什么

1. `Runtime Service`：在现有 `scripts/run-copilot-inference.py`、`services/gateway/copilot_gateway.py`、`services/runtime/inference_provider.py`、`services/platform/` 和 `RadishFlow` service smoke gate 之上，继续把已落地的最小 `provider registry` 与 `Go` service bootstrap 骨架扩到协议兼容层、本地启动、配置、调用和部署入口。当前第一版 `Go -> Python` bridge 已接通 `/v1/chat/completions`、`/v1/responses`、`/v1/messages` 与 `/v1/models`，并补了第一版 SSE 流式兼容骨架、把 `/v1/models` 从 provider 目录推进到 bridge-backed provider/profile inventory，再补上 `GET /v1/models/{id}` 的精确 lookup；当前已经把 `/v1/chat/completions` 的 request-side provider/profile 选择显式化，并把流式路径推进到 bridge 增量转发，其中 `HuggingFace` / `Ollama` 已有第一版 provider coverage，平台级 `ops smoke`、本地启动 runbook drift check、脱敏配置摘要 / config check、JSON 配置文件层级、稳定本地启动 wrapper、最小 deployment smoke 与结构化 diagnostics/failure boundary 已纳入快速门禁；下一步继续补更广 provider/profile discoverability 和更正式的部署观测策略。
2. `Conversation & Session`：补齐会话标识、历史压缩、恢复和审计边界，不再只停留在 `conversation_id` 透传。
3. `Tooling Framework`：把当前 task-local 的检索、候选生成和 builder 经验收口成正式工具契约、registry、timeout/retry/policy。
4. `Evaluation & Governance`：把已有 schema、offline eval、service smoke 和 runtime provider dispatch smoke 扩展到 runtime、session、tooling 和 deployment 门禁。
5. `Model Adaptation`：在前四项稳定后再定义首版基座、蒸馏和训练计划；当前不启动训练放量。

## 为什么是这些任务

- 上层项目目前没有真实挂载点、确认流和命令承接接口，继续细化接线设计收益很低。
- 仓库里已经有 runtime、gateway、adapter、eval 和 governance 资产，可以先把平台骨架做完整。
- `provider registry` 最小骨架已经把 CLI runtime、进程内 gateway 和 southbound provider 选择收口到同一条真相源，接下来更适合继续补 northbound 协议兼容和本地 service bootstrap。
- 平台表层语言边界已固定为：`UI` 用 `React + Vite + TypeScript`，平台服务层用 `Go`，模型侧继续保留 `Python`，所有层只共享 `contracts/` 里的 canonical protocol。
- 平台服务层当前已经有最小 `Go HTTP` 壳、`/healthz`、`/v1/models`、`/v1/chat/completions`、`/v1/responses` 与 `/v1/messages` bridge，后续只补协议兼容面，不再回头把模型逻辑写回 `Go`。
- 如果平台既不能稳定接外部模型，也不能对外暴露常见协议接口，就还不能算真正可用的模型平台。
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
- 查最近执行细节：读本周周志 `docs/devlogs/2026-W19.md`

## 验证基线

文档或治理改动完成后，优先执行：

```bash
./scripts/check-repo.sh --fast
```

Windows / PowerShell 环境使用：

```powershell
pwsh ./scripts/check-repo.ps1 -Fast
```
