# RadishMind 阶段路线图

更新时间：2026-05-14

## 路线图原则

路线图只记录阶段目标、当前进度、下一步和停止线。批次细节、历史失败、完整实验输出和长命令不放在本入口文档中，应进入周志、实验 manifest、run record 或任务卡。

当前长期目标保持不变：`RadishMind` 是受控 Copilot / Agent runtime platform + 可替换模型能力，不是单一万能模型。

若要理解“为什么路线这样排”，先读 [战略定义](radishmind-strategy.md)。

## 当前路线切换

从 2026-05-10 开始，仓库主线正式从“围绕 `M3/M4` 收口继续做局部维护”切换为“基于已收口证据继续建设平台本体”。

当前已经冻结的历史证据：

- `M3`：gateway、service smoke、UI consumption 与 candidate handoff 已收口为服务/API 门禁。
- `M4`：broader 15 样本人工复核为 15/15 `reviewed_pass`，`3B/4B` guided capacity review 已收口为正式审计记录。

这些资产继续保留，但不再等同于当前唯一主线。

## 五条主线

### 1. `Runtime Service`

目标：把现有 CLI runtime、进程内 gateway、route 识别和 smoke gate 收口为明确的 provider registry、协议兼容层、本地运行、配置、启动和部署基础。

状态：`scripts/run-copilot-inference.py`、`services/gateway/copilot_gateway.py`、`services/runtime/inference_provider.py`、`services/runtime/provider_registry.py`、`services/platform/`、`RadishFlow` gateway demo 与 service smoke matrix 已具备基础骨架；当前 southbound 已通过统一 registry 收口 `mock`、`openai-compatible`、`HuggingFace`、`Ollama` 主入口与 `openai-compatible chat`、`gemini-native`、`anthropic-messages` 分流，`local_transformers` 则主要存在于 candidate/runtime 实验链路中。平台表层语言分工已固定为 `UI=React + Vite + TypeScript`、`Platform Service Layer=Go`、`Model Side=Python`。当前 `Go` 层已落最小服务启动、`/healthz`、`/v1/models`、`/v1/chat/completions`、`/v1/responses` 和 `/v1/messages` bridge，并补了第一版 SSE 流式兼容骨架、bridge-backed provider/profile inventory、`GET /v1/models/{id}` 精确 lookup、request-side provider/profile 选择、流式增量转发、`HuggingFace` / `Ollama` coverage。平台级 `ops smoke` 已固定 `go test ./...`、provider registry 与受控 profile inventory 门禁；本地启动 runbook、runbook drift check、脱敏配置摘要 / config check、JSON 配置文件层级、稳定本地启动 wrapper、最小 deployment smoke、结构化 diagnostics/failure boundary、provider/profile discoverability 对齐、request-level observability 与 error taxonomy 已补齐。`P1 Runtime Foundation` 已达到 short close；第一版 northbound 仍是窄切片，但继续横向扩同层配置、别名和兜底的收益已经下降，主要实现重心切到 `P2 Session & Tooling Foundation`。

下一步：不再继续把 `P1` 做成无限硬化阶段；进入 `P2 Session & Tooling Foundation`，先补 session contract、history policy、recovery record、tool schema、tool registry、tool policy 和 audit record。

### 2. `Conversation & Session`

目标：让多轮对话、历史压缩、恢复和审计成为平台能力，而不是各任务自己拼上下文。

状态：已补首版 `session-record.schema.json`、`session-recovery-checkpoint.schema.json`、`session-recovery-checkpoint-manifest.schema.json`、`session-recovery-checkpoint-read.schema.json`、fixture 和快速门禁，并让 `Go` northbound 兼容层在显式 `radishmind` 会话扩展存在时写入 `context.northbound.session`；`state_policy` 已固定会话状态与 tool result cache 的 v1 落点只允许 northbound metadata / session recovery checkpoint，不启用 durable memory；recovery checkpoint v1 只保存 request/session/tool audit/tool metadata 引用，read result 只暴露 metadata refs 和 tool audit 治理摘要，不保存或返回真实工具结果，也不自动 replay。平台层已新增 metadata-only route smoke，并通过 denied query fixture 拒绝 materialized result、result ref、executor ref、durable memory 与 replay 类查询参数；readiness summary、implementation preconditions、negative regression skeleton、`session-tooling-negative-regression-suite-readiness.json`、confirmation flow design、independent audit records design、result materialization policy design、executor boundary design、storage backend design、`session-tooling-foundation-status-summary.json` 和 `session-tooling-close-candidate-readiness-rollup.json` 已把当前状态收口为 `close candidate / governance-only`，但当前仍没有 durable session store、durable checkpoint store、durable audit store、durable result store、长期记忆、真实 checkpoint storage backend 或跨轮恢复执行器，也不声明 P2 short close。

下一步：维护 P2 close-candidate readiness rollup，继续把 confirmation、independent audit、result materialization、executor boundary、storage backend 与完整负向回归的满足状态保持为可检查口径；在 short close 前置条件满足前，不进入真实实现设计。

### 3. `Tooling Framework`

目标：把检索、局部规则、候选生成和 builder 经验收口为正式工具契约、registry、policy 和 audit。

状态：当前已有 task-local 的 deterministic tooling 与 builder 资产；最小 `tool.schema.json`、`tool-registry.schema.json`、`tool-audit-record.schema.json`、registry fixture、policy/audit fixture 和快速门禁已开始落地，用于固定工具注册、调用轨、timeout/retry/policy、session binding、metadata-only result cache 和 audit 的结构边界。tool audit summary 已进入 checkpoint read route smoke，用于固定 execution disabled、not executed、metadata-only cache 和 no result ref；session/tooling promotion gate 分层、负向消费 summary、route smoke coverage summary、readiness summary、implementation preconditions、negative regression skeleton、`session-tooling-negative-regression-suite-readiness.json`、confirmation flow design、independent audit records design、result materialization policy design、executor boundary design、storage backend design、close candidate status summary 与 close-candidate readiness rollup 已进入快速门禁。当前仍没有真实工具执行器、durable audit store、durable result store、长期记忆或新的 provider/model 实验，negative skeleton、suite readiness、audit design、materialization policy design、executor boundary design 和 storage backend design 也不等同于完整 `negative_regression_suite`。

下一步：继续通过 close-candidate rollup 守住 tooling contract、audit、result materialization、executor boundary 与 storage backend 的设计门禁；在上层确认流接线和完整负向回归满足前，不启动真实执行。

### 4. `Evaluation & Governance`

目标：让 runtime、session、tooling、deployment 和 model adaptation 都有统一门禁，而不是只校验模型输出。

状态：schema、offline eval、candidate record、review record、`check-repo`、service smoke 和 runtime provider dispatch smoke 已具备基础；平台级 smoke 已继续扩展到 runtime config/deployment/diagnostics/request observability 与 P2 session/tooling/checkpoint governance。当前 `session-tooling-foundation-status-summary.json` 只声明 `close candidate / governance-only`，不声明实现完成。

下一步：维持 advisory-only、confirmation、route、citation、handoff 不执行和 metadata-only 这些不变量；完整负向回归和真实实现 gate 必须先于 executor/storage/confirmation 实现落地。

### 5. `Model Adaptation`

目标：在平台契约稳定后，再定义首版基座、蒸馏和训练升级路径。

状态：raw、repair、injection、guided、task-scoped builder、offline eval 和 training sample conversion 已有资产，但当前还不具备“直接扩大训练规模”的时机。

下一步：先以平台契约为前提锁定 v1 训练目标，再决定新的实验或蒸馏路线；没有新能力假设前，不继续重跑同一批 `M4` 实验。

## 辅助支线

### `Image Path`

状态：intent、backend request、artifact schema 与最小评测 manifest 已具备；真实 backend 仍未接入。

下一步：继续收口 image adapter handshake、safety gate 和 artifact 返回链路，不下载模型、不生成图片。

### 上层项目接入

状态：`RadishFlow` 门禁已冻结，`Radish` docs QA 资产已具备，`RadishCatalyst` 仍只做文档预留；三个上层项目当前都不具备真实接入能力。

下一步：先推进平台本体；待上层具备真实挂载点、确认流和命令承接接口后，再只选一个切片真实接入。

## 阶段顺序

### `P0`：项目重定义与能力盘点

目标：把“项目到底是什么、有哪些主线、哪些能力缺口最关键”写成正式文档和能力矩阵。

状态：已完成平台重定义、能力矩阵和主线切换；后续只维护文档口径，不再作为主要实现阶段。

### `P1`：Runtime Foundation

目标：收口最小本地 service bootstrap、provider registry、northbound/southbound 协议兼容、配置、调用和 smoke 路径。

状态：short close。provider registry、Go service bootstrap、northbound bridge、provider/profile discoverability、config layering、wrapper、deployment smoke、diagnostics、request observability、error taxonomy 与三种 northbound 协议的 selection metadata smoke 已进入平台单元测试和快速门禁。

### `P2`：Session & Tooling Foundation

目标：补齐 conversation/session contract、tool contract、registry、policy 和审计轨。

状态：`close candidate / governance-only`。session contract、history policy、state policy、recovery record、recovery checkpoint record/manifest/read result、denied query fixture、promotion gate fixture、negative consumption summary、route smoke coverage summary、readiness summary、implementation preconditions、negative regression skeleton、`session-tooling-negative-regression-suite-readiness.json`、confirmation flow design、independent audit records design、result materialization policy design、executor boundary design、storage backend design、foundation status summary、close-candidate readiness rollup 与 northbound session metadata 已有首版门禁；tool schema、tool registry、tool policy、session binding、metadata-only result cache、audit record 和 promotion gate 分层已有最小契约与快速门禁。当前仍不声明 P2 short close，也不具备真实 executor、durable storage、上层 confirmation flow 接线、materialized result reader、durable audit store、durable result store 或完整 `negative_regression_suite`。

### `P3`：Local Deployment & Ops Governance

目标：让本地长驻服务、启动脚本、观测、故障边界和 deployment smoke 具备正式口径。

状态：本地治理第一版已具备 wrapper、配置文件层级、deployment smoke、启动前 diagnostics 和 runbook drift check；尚未进入 production secret backend、进程守护、正式部署环境隔离或可发布部署包。

### `P4`：Model Adaptation & Training

目标：在平台边界稳定后，定义首版基座、蒸馏和训练升级计划。

状态：当前不提前放量。

### `P5`：Real Upstream Integration

目标：在上层项目具备真实挂载点后，选择首个切片完成真正接入。

状态：当前等待上层条件成熟。

## 下一步

1. 继续维护 `P2 Session & Tooling Foundation` close-candidate readiness rollup，确保 `scripts/checks/fixtures/session-tooling-close-candidate-readiness-rollup.json` 能明确 design gate 已完成项、仍不满足项和进入 short close 前的硬前置条件。
2. 继续把 `Evaluation & Governance` 从“任务输出门禁”扩展为“平台能力门禁”，重点覆盖 session、tooling、negative regression suite readiness、implementation preconditions 和 close-candidate rollup。
3. 只有在前述平台边界稳定后，才定义新的训练 / 蒸馏主线。
4. 继续维持上层项目接入前置条件总表，不提前细化不存在的真实接线。

## 停止线

- 不把 repaired、injected、guided 或 builder 轨通过写成 raw 模型能力通过。
- 不把机器指标通过写成人工可接受度通过。
- 不在没有非重复能力假设时继续扩同一批 `M4` 实验。
- 不在上层项目没有真实挂载点时继续细化假想接线设计。
- 不把 `P1` 继续扩成无止境的 provider/config/diagnostics 细化阶段。
- 不让模型直接写上层业务真相源。
- 不用晦涩抽象、空泛 helper 或多层 fallback 掩盖代码职责不清。
