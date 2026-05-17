# RadishMind 战略定义

更新时间：2026-05-10

## 这份文档回答什么

本文档用于把 `RadishMind` 的项目定义再收紧一层，回答这些问题：

- 这个项目为什么成立
- 它到底是产品、平台，还是模型仓库
- 它的核心价值是什么
- 它最应该优先做什么
- 哪些事情即使“能做”，也不该成为主线

这份文档不是路线图，不替代产品范围、能力矩阵或架构说明；它负责冻结“项目是什么”和“为什么这样规划”。

## 一句话定义

`RadishMind` 是一个把多模型、多协议、多任务收口成可审计、可控、可本地部署的 AI runtime 的中间层平台。

更准确地说，它是 `Radish` 体系下的 `AI Middleware / AI Runtime`，而不是：

- 单一模型项目
- 通用聊天产品
- 只做接线的适配仓库
- 替代业务内核的自治系统

## 核心价值

`RadishMind` 成立的根本原因，不是“我们也要做 AI”，而是要解决四类长期问题：

1. 模型来源会变化  
   自研模型、`HuggingFace`、`Ollama`、OpenAI-compatible、Gemini、Anthropic 都可能进入系统，不能把产品绑死在单一模型或单一厂商接口上。

2. 协议形态会变化  
   外部客户端会希望使用 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 或其它兼容协议；内部上层项目也需要稳定 canonical contract。

3. 任务风险不一致  
   有些任务只是回答问题，有些任务会给出候选动作，有些任务可能走到 tool handoff；不能用“一个聊天接口”覆盖全部风险等级。

4. 生产约束比模型能力更难收口  
   真正困难的不是模型能不能说出一句像样的话，而是：
   - 是否可审计
   - 是否可确认
   - 是否可复跑
   - 是否能本地部署
   - 是否不写坏上层真相源

所以 `RadishMind` 的核心价值是：

把“模型能力”转化成“可控产品能力”。

## 三层定位

为了避免项目继续发散，正式定位建议固定成三层。

### 1. `Model Access Layer`

职责：统一接入不同模型和不同 provider。

它要解决的问题是：

- 自研 `RadishMind-Core` 怎么接
- 本地模型怎么接
- `HuggingFace`、`Ollama`、OpenAI-compatible、Gemini、Anthropic 怎么接
- 每个 provider 支持哪些能力
- 如何做 auth、timeout、retry、fallback、cost/latency 分层

这一层的目标不是“追求最强模型”，而是“把模型调用变成稳定能力”。

### 2. `Protocol / API Layer`

职责：统一对外暴露不同协议接口，并把它们归一到内部 canonical contract。

它要解决的问题是：

- 外部客户端如何调用 `RadishMind`
- `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 如何兼容
- 内部 native Copilot API 如何保持为真相源
- 不同 northbound 协议如何共享同一条 runtime / gateway truth

这一层的目标不是复制一堆接口，而是保证：

- 兼容层只做翻译
- 业务真相不分叉
- 不为每个协议复制一套核心逻辑

### 3. `Task & Governance Layer`

职责：统一管理会话、工具、任务执行边界、评测、审计和安全控制。

它要解决的问题是：

- 会话与历史如何压缩、恢复和审计
- 工具调用如何声明、限权和记录
- 候选动作如何保留确认边界
- 模型输出如何经过验证、builder、review 和 promotion gate

这一层的目标不是“让 agent 更像人”，而是“让系统更像产品”。

## Northbound / Southbound 战略

这是 `RadishMind` 和普通模型仓库最大的差异点，必须单独固定。

### `Northbound`

指外部客户端如何调用 `RadishMind`。

当前目标包括：

- native Copilot API
- `/v1/chat/completions`
- `/v1/responses`
- `/v1/messages`
- `/v1/models`

原则：

- 所有 northbound 协议都必须归一到 canonical `CopilotRequest / CopilotResponse / CopilotGatewayEnvelope`
- northbound compatibility layer 不是第二套业务真相源
- 不允许每个 northbound 接口各自私接 provider

### `Southbound`

指 `RadishMind` 如何调用模型和工具。

当前目标包括：

- `RadishMind-Core`
- `local_transformers`
- `HuggingFace`
- `Ollama`
- OpenAI-compatible
- Gemini native
- Anthropic messages

原则：

- 所有 southbound provider 都应进入统一 provider registry
- 每个 provider 的 capability、限制、鉴权、timeout/retry 策略都必须可声明
- task / route 不应直接依赖某一个 provider 的私有细节

## Provider Capability Matrix

`RadishMind` 后续需要一份独立的 provider capability matrix 文档或 fixture。最小维度应固定为：

- `provider_id`
- `transport`
- `local_or_remote`
- `chat`
- `responses`
- `messages`
- `models_list`
- `streaming`
- `json_schema_output`
- `tool_calling`
- `image_input`
- `image_output`
- `auth_mode`
- `timeout_policy`
- `retry_policy`
- `cost_profile`
- `latency_profile`
- `deployment_mode`

如果没有这张矩阵，后面的 routing、fallback、deployment 和训练路线都会继续凭印象推进。

## Service Mode Tiers

`RadishMind` 不应把所有请求都默认做成 agent。正式分级建议固定为三档：

### 1. `Stateless API`

- 单轮请求响应
- 不保留长期状态
- 适合兼容 `/v1/chat/completions` 这类请求

### 2. `Session Runtime`

- 保留 `conversation_id`
- 管理历史压缩、恢复和会话审计
- 适合多轮解释、问答和半结构化辅助

### 3. `Agent Runtime`

- 可带工具调用、计划和 handoff
- 必须显式治理 tool policy、risk 和 audit
- 默认不是所有任务的起点

这三档的意义是：先做清楚每一级能力，而不是一开始就把整个系统做成无边界 agent。

## Action Safety Ladder

当前只有 `requires_confirmation` 还不够，后续建议固定为动作梯度，而不是二元开关。

建议层级：

- `answer_only`
- `proposal_only`
- `handoff_ready`
- `tool_callable`
- `write_blocked`
- `write_allowed_by_policy`

解释：

- `answer_only`：只能解释，不给动作
- `proposal_only`：可以给建议，但不能执行
- `handoff_ready`：可以交给上层命令层或人工确认流
- `tool_callable`：允许进入受控工具层
- `write_blocked`：即使模型想写，也必须阻断
- `write_allowed_by_policy`：只有在规则层明确允许时才可进入更高风险动作

这比单独的高低风险更适合平台化治理。

## Tooling 战略

工具层不能长期停留在 task-local builder 和脚本散落实现。

后续最少要固定四件东西：

- `tool schema`
- `tool registry`
- `tool policy`
- `tool audit record`

原则：

- 能规则化的逻辑优先进入工具层
- 工具层输出证据、候选或受控结果，不直接成为业务真相源
- 工具调用必须可重放、可限权、可审计

## 训练战略

训练是重要方向，但不是当前项目存在的前提。

正式判断应固定为：

- `RadishMind` 不是为了训练自己的模型而存在
- 它是为了把模型能力收口成可控 runtime 而存在
- 训练、自研基座、蒸馏和晋级，是平台稳定后的增强项

这意味着：

- 短期优先级仍应低于 provider/protocol/runtime/session/tooling/governance
- 自研模型路线不能替代多模型接入路线
- builder / guided / repair 轨不能写成 raw 晋级证据

## 反向非目标

为了防止项目再次偏航，以下内容应长期明确为非目标：

- 不做通用聊天产品
- 不做通用 SaaS 聊天前端
- 不做基础大模型预训练项目
- 不做无边界自治 agent
- 不做直接写业务真相源的执行器
- 不做只绑定单一模型或单一协议的封闭系统
- 不做只服务单一实验链路的临时仓库

## 当前最该做什么

基于上面的定义，当前最高优先级不是继续扩旧实验，也不是继续补假想接线，而是把平台骨架补到可执行：

1. 建立正式 `provider registry`
2. 建立正式 `protocol compatibility layer`
3. 先落地 northbound `/v1/chat/completions` 与 `/v1/models`
4. 先落地 southbound `HuggingFace` 与 `Ollama`
5. 把 capability matrix、compatibility smoke 和 governance gate 一起补齐

## 判断标准

如果后续一个改动同时满足这些条件，方向通常是对的：

- 更接近 `AI Middleware / AI Runtime`，而不是聊天壳或模型脚手架
- 同时增强 northbound 与 southbound 的可替换性
- 不把协议兼容层变成第二套真相源
- 不把训练和模型实验重新抬回第一优先级
- 可审计、可确认、可复跑、可部署的边界更清晰
