# RadishMind 阶段路线图

更新时间：2026-05-10

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

状态：`scripts/run-copilot-inference.py`、`services/gateway/copilot_gateway.py`、`services/runtime/inference_provider.py`、`services/runtime/provider_registry.py`、`RadishFlow` gateway demo 与 service smoke matrix 已具备基础骨架；当前 southbound 已通过统一 registry 收口 `mock`、`openai-compatible` 主入口与 `openai-compatible chat`、`gemini-native`、`anthropic-messages` 分流，`local_transformers` 则主要存在于 candidate/runtime 实验链路中。但目前还没有正式的 `HuggingFace / Ollama` 服务接入、northbound `/v1/chat/completions` / `/v1/responses` / `/v1/messages` / `/v1/models` 兼容面、正式长驻服务、配置分层和部署 runbook。

下一步：在现有 provider registry 骨架上先完成协议翻译真相源和最小本地 service bootstrap；在此基础上优先补 `/v1/chat/completions` 与 `/v1/models`，再补 `/v1/responses`、`/v1/messages`、`HuggingFace` 和 `Ollama`。

### 2. `Conversation & Session`

目标：让多轮对话、历史压缩、恢复和审计成为平台能力，而不是各任务自己拼上下文。

状态：当前只有 `conversation_id` 透传和局部 snapshot 语义，还没有正式 session schema、history policy、recovery record 或跨轮 smoke。

下一步：补 session contract、fixture 和最小 smoke，再决定状态落点和缓存策略。

### 3. `Tooling Framework`

目标：把检索、局部规则、候选生成和 builder 经验收口为正式工具契约、registry、policy 和 audit。

状态：当前已有 task-local 的 deterministic tooling 与 builder 资产，但还没有通用工具注册、调用轨、timeout/retry/policy 和 tool audit。

下一步：先定义最小 tool contract 和 registry 原型，避免继续把工具能力散落在 adapter 与脚本里。

### 4. `Evaluation & Governance`

目标：让 runtime、session、tooling、deployment 和 model adaptation 都有统一门禁，而不是只校验模型输出。

状态：schema、offline eval、candidate record、review record、`check-repo` 和 service smoke 已具备基础，但平台级 smoke 仍主要集中在 `RadishFlow` 任务面。

下一步：把 smoke gate 扩展到 runtime、session、tooling 和部署边界，并维持 advisory-only、confirmation、route、citation 和 handoff 不执行这些不变量。

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

状态：当前正在完成。

### `P1`：Runtime Foundation

目标：收口最小本地 service bootstrap、provider registry、northbound/southbound 协议兼容、配置、调用和 smoke 路径。

状态：这是当前默认第一实现项；provider registry 最小骨架已落地，下一步继续补协议兼容层和本地 service bootstrap。

### `P2`：Session & Tooling Foundation

目标：补齐 conversation/session contract、tool contract、registry、policy 和审计轨。

状态：紧随 `P1`，不提前空转架构。

### `P3`：Local Deployment & Ops Governance

目标：让本地长驻服务、启动脚本、观测、故障边界和 deployment smoke 具备正式口径。

状态：当前尚未开始。

### `P4`：Model Adaptation & Training

目标：在平台边界稳定后，定义首版基座、蒸馏和训练升级计划。

状态：当前不提前放量。

### `P5`：Real Upstream Integration

目标：在上层项目具备真实挂载点后，选择首个切片完成真正接入。

状态：当前等待上层条件成熟。

## 下一步

1. 先推进 `Runtime Service`，把 provider registry、外部模型接入和对外协议兼容收口成正式平台能力。
2. 再补 `Conversation & Session` 与 `Tooling Framework` 的最小契约，不再只靠 task-local 透传和脚本散落逻辑。
3. 把 `Evaluation & Governance` 从“任务输出门禁”扩展为“平台能力门禁”。
4. 只有在前述平台边界稳定后，才定义新的训练 / 蒸馏主线。
5. 继续维持上层项目接入前置条件总表，不提前细化不存在的真实接线。

## 停止线

- 不把 repaired、injected、guided 或 builder 轨通过写成 raw 模型能力通过。
- 不把机器指标通过写成人工可接受度通过。
- 不在没有非重复能力假设时继续扩同一批 `M4` 实验。
- 不在上层项目没有真实挂载点时继续细化假想接线设计。
- 不让模型直接写上层业务真相源。
- 不用晦涩抽象、空泛 helper 或多层 fallback 掩盖代码职责不清。
