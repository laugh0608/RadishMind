# RadishMind 项目总览与使用指南

更新时间：2026-05-10

## 这份文档讲什么

这是一份面向新读者的项目说明书，回答四个问题：

- `RadishMind` 现在被定义成什么
- 仓库怎么分工
- 今天可以怎样实际运行它
- 如果你要继续推进开发，先看哪几条主线

它不替代 `docs/radishmind-current-focus.md`、`docs/devlogs/` 或任务卡，也不记录阶段推进流水。

## 项目定位

`RadishMind` 现在的正式定位是：

协议驱动、可审计、可本地部署、可工具化的 Copilot / Agent runtime platform。

它不是上层业务真相源，不替代 `RadishFlow`、`Radish` 或 `RadishCatalyst` 的业务决策权，也不是“只放模型实验”的仓库。

当前主要职责是：

- 接收结构化上下文
- 运行最小推理链路
- 兼容多种上游模型接入方式与多种下游协议接口
- 组织局部工具、规则和响应收口
- 输出解释、诊断、结构化建议和候选动作
- 维护统一协议、评测门禁、审计记录和训练治理

## 当前五条主线

1. `Runtime Service`：本地启动、gateway、route、provider/profile、协议兼容、响应封装、部署基础。
2. `Conversation & Session`：会话标识、历史压缩、恢复和审计边界。
3. `Tooling Framework`：检索、附件解析、候选生成、builder、tool policy 和 audit。
4. `Evaluation & Governance`：schema、smoke、offline eval、review、promotion gate。
5. `Model Adaptation`：基座选型、prompt/runtime 协同、蒸馏、训练样本治理和模型晋级。

如果你今天想推进开发，默认先看 `Runtime Service`，再看 `Conversation & Session` 与 `Tooling Framework`。

## 目录速览

- `docs/`：正式文档源
- `contracts/`：JSON Schema 真相源
- `scripts/`：检查、运行、转换、评测和最小运维入口
- `datasets/`：eval 样本、示例对象和 candidate record
- `training/`：训练治理、实验说明和复核记录
- `services/`：runtime 与 gateway 实现
- `adapters/`：上游项目适配层
- `tmp/`：本地临时产物，不提交

## 今天能怎么运行

### 1. 直接跑最小推理链路

当前最小运行入口是 `scripts/run-copilot-inference.py`。它不是长驻服务，而是单次 CLI runtime。

```bash
python3 scripts/run-copilot-inference.py \
  --sample datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json \
  --provider mock \
  --response-output tmp/rf-suggest-edit.response.json
```

如果后续要接真实 provider，再显式传 `--provider openai-compatible`、`--provider-profile`、`--model`、`--base-url`、`--api-key`。当前这条入口已经能按 profile 分流到 `openai-compatible chat`、`gemini-native` 和 `anthropic-messages` 三类上游协议，但还没有把 `HuggingFace`、`Ollama`、`/v1/responses` 或 `/v1/models` 收口为正式服务能力。

### 2. 跑进程内 gateway demo

如果你要看 `RadishFlow export -> adapter -> request -> gateway` 整条链路，当前正式入口是：

```bash
python3 scripts/run-radishflow-gateway-demo.py \
  --check-summary scripts/checks/fixtures/radishflow-gateway-demo-summary.json
```

这条链路当前使用 mock provider，只验证 route、schema、advisory-only 和 `requires_confirmation` 等不变量。

### 3. 跑 service/API smoke matrix

当前 `RadishFlow` 的正式服务门禁入口是：

```bash
python3 scripts/check-radishflow-service-smoke-matrix.py \
  --check-summary scripts/checks/fixtures/radishflow-service-smoke-matrix-summary.json
```

这条矩阵会一起核对：

- CLI runtime
- gateway API
- gateway demo
- UI consumption
- candidate handoff

它现在是仓库里最接近“服务切片验收”的正式门禁。

### 4. 跑本地候选模型输出

如果你要继续看 `RadishMind-Core` 本地候选输出，入口仍是：

```bash
python3 scripts/run-radishmind-core-candidate.py \
  --provider local_transformers \
  --model-dir /path/to/model \
  --output-dir tmp/radishmind-core-candidate-local \
  --allow-invalid-output \
  --validate-task
```

这仍属于模型评测 / 训练前治理链路，不等同于平台正式服务。

### 5. 当前协议兼容边界

当前平台必须区分两类兼容：

- 南向模型接入：平台去调用哪些模型和哪些 provider
- 北向协议输出：外部客户端如何调用 `RadishMind`

当前真实状态是：

- 南向已有一部分：`openai-compatible` 主入口、`gemini-native`、`anthropic-messages`，以及评测链路中的 `local_transformers`
- 北向还没有完成：还没有正式 HTTP server，也没有 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 对外兼容接口

## 今天还不能算完成的能力

当前仓库还没有这些正式能力：

- 官方长驻服务进程
- 正式 HTTP API 包装
- `HuggingFace` 与 `Ollama` 的正式服务级接入
- `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 对外兼容接口
- session store / history policy / recovery runbook
- 通用 tool registry 和 tool calling contract
- 官方 deployment runbook 或可发布部署包

所以如果你问“现在怎么部署”，准确答案是：当前只有本地 CLI runtime、进程内 gateway 和 smoke/demo 链路，还没有正式部署面。

## 读文档顺序

如果你刚接触这个仓库，建议按这个顺序读：

1. [文档入口](README.md)
2. 项目总览与使用指南
3. [当前推进焦点](radishmind-current-focus.md)
4. [产品范围](radishmind-product-scope.md)
5. [战略定义](radishmind-strategy.md)
6. [能力矩阵](radishmind-capability-matrix.md)
7. [系统架构](radishmind-architecture.md)
8. [跨项目集成契约](radishmind-integration-contracts.md)
9. [脚本目录说明](../scripts/README.md)
10. [数据集目录说明](../datasets/README.md)
11. [训练目录说明](../training/README.md)

## 默认不要做

- 不把 `RadishMind` 做成上层业务真相源
- 不默认把 builder/guided 结果当成 raw 晋级证据
- 不在上层项目还没具备真实挂载点时继续细化假想接线
- 不把长驻服务、session、tooling、外部模型兼容或协议兼容能力写成“已经具备”
- 不默认下载大模型、数据集或权重
