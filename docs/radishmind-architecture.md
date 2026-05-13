# RadishMind 系统架构

更新时间：2026-05-13

## 架构目标

`RadishMind` 的正式架构目标不再只是“把单次模型推理接上去”，而是建设一个可本地运行、可审计、可工具化的 Copilot / Agent runtime platform。

这套平台当前从两个视角理解：

- 平台视角：看五条主线怎么协同
- 请求视角：看单次 `CopilotRequest -> CopilotResponse` 是怎么流动的

## 平台视角

### 1. `Runtime Service`

- 负责启动、配置、provider/profile 选择、route 识别、gateway 封装、协议兼容和部署边界。
- 当前实现核心在 `scripts/run-copilot-inference.py`、`services/gateway/copilot_gateway.py`、`scripts/run-platform-bridge.py` 与 `services/platform/`。
- 当前 southbound 已开始由统一 `provider registry` 收口：现有 `mock`、`openai-compatible`、`HuggingFace`、`Ollama` 主入口与 `openai-compatible chat`、`gemini-native`、`anthropic-messages` 分流都归到同一条 provider truth；`local_transformers` 目前主要停留在 candidate/runtime 评测链路。
- 当前 northbound 对外形态已经开始由 `Go` 承载最小正式 `HTTP` 服务壳；`Python` 继续保留 CLI runtime 和 canonical gateway 语义，`Go` 只做协议兼容与进程调度，避免把平台服务层锁死在 `Python`。
- `UI` 层默认 `React + Vite + TypeScript`，通过北向协议消费平台能力，不直接承载模型实现逻辑。

### 2. `Conversation & Session`

- 负责 `conversation_id`、会话历史、恢复、压缩和审计边界。
- 当前已有首版 session record、history policy、state policy、recovery record、recovery checkpoint record/manifest/read result、northbound session metadata 和 metadata-only route smoke；仍没有 durable session store、长期记忆、真实 checkpoint storage backend 或跨轮恢复执行器。

### 3. `Tooling Framework`

- 负责检索、附件解析、项目语义转换、本地候选生成、response builder 和工具策略。
- 当前已有首版 tool contract、registry、policy/audit record、session binding、metadata-only result cache 和 checkpoint read `tool_audit_summary`；仍没有真实工具执行器、materialized tool result cache、durable tool store 或上层确认流接线。

### 4. `Model Runtime`

- 负责模型推理、候选文本生成、结构化约束、guided/runtime 协同和图片理解。
- `RadishMind-Core` 属于这一层，但不是整个平台本身。

### 5. `Evaluation & Governance`

- 负责 schema、smoke、offline eval、candidate record、review、promotion gate 与仓库级检查。
- 这一层保证平台不是“能跑一次”，而是“能长期复跑并解释为什么通过或不通过”。

## 请求视角

当前还有一层必须继续补齐、但已经开始正式落地的协议翻译边界：

- 北向：`/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 等兼容接口如何映射到 canonical `CopilotRequest`
- 南向：`RadishMind-Core`、`HuggingFace`、`Ollama`、OpenAI-compatible、Gemini、Anthropic 等 provider 如何被统一调度

平台内部真相源仍应保持 `CopilotRequest / CopilotResponse / CopilotGatewayEnvelope`，兼容接口只做翻译层，不另起第二套真相源。

当前单次请求的主流程仍保持六层：

1. Client Adapters & Context Packers
2. Copilot Gateway / Task Router
3. Retrieval & Tool Layer
4. Model Runtime Layer
5. Rule Validation & Response Builder
6. Data / Evaluation / Training Pipeline

这个拆分用于隔离项目语义、工具编排、模型推理、安全校验和评测闭环，让 `RadishFlow`、`Radish` 与后续 `RadishCatalyst` 能通过统一协议接入，而不是各自私接模型。

## 请求流程

```text
外部客户端协议 / 上层项目状态 / 文档 / 附件 / 图像
        ↓
Protocol Compatibility Layer 归一到 canonical request
        ↓
Adapter 打包为 CopilotRequest
        ↓
Gateway 识别 project / task / schema_version
        ↓
Retrieval / Tools 获取证据并压缩上下文
        ↓
Model Runtime 生成解释、问题和候选意图
        ↓
Rule Validation / Response Builder 收口结构和风险
        ↓
CopilotResponse / GatewayEnvelope
        ↓
Protocol Compatibility Layer 翻译回 northbound response
```

## 分层职责

### 1. Client Adapters & Context Packers

- 将上层状态转换为统一 `CopilotRequest`。
- 裁剪、脱敏或摘要敏感字段。
- 将 `CopilotResponse` 映射回 UI、日志或候选提案。
- 前端 UI 默认使用 `React + Vite + TypeScript`，通过协议对接后端，不直接依赖模型实现语言。
- 当前 `RadishFlow` 优先维护 `export -> adapter -> request` 链路；`Radish` 优先维护文档和内容上下文；`RadishCatalyst` 暂不落真实 adapter。

### 2. Copilot Gateway / Task Router

- 统一校验请求、识别任务、选择 provider/profile，并返回 `CopilotGatewayEnvelope`。
- 当前 `SUPPORTED_ROUTES` 仍然有限，说明平台还在先做骨架而不是全量铺开任务面。
- 当前 `Go` 平台服务层已经通过 Python bridge 接到 `/v1/chat/completions`、`/v1/responses`、`/v1/messages` 与 `/v1/models` 的第一版兼容层；这条 bridge 目前仍是窄切片，先把非流式文本消息固定映射到 `radish/answer_docs_question`，并通过 SSE 做出第一版流式兼容骨架、把 `/v1/models` 从 provider 目录推进到 bridge-backed provider/profile inventory，再补上 `GET /v1/models/{id}` 的精确 lookup；当前已经把 `/v1/chat/completions` 的 request-side provider/profile 选择显式化，并把流式路径推进到 bridge 增量转发，同时补了 `HuggingFace` / `Ollama` 的第一版 southbound provider coverage；`/v1/models`、northbound request selection 与 diagnostics 已共享 provider-qualified profile inventory、credential state、deployment mode、streaming、route 和 protocol metadata，三种 northbound 协议也共享 `request_id`、latency、error code 与 failure boundary 观测口径，但这些路径仍必须复用同一条 gateway truth，而不是绕过 gateway 直接拼 provider 请求。
- 服务/API smoke 当前锁定 advisory-only、schema validation、route metadata、error envelope 和 handoff 不执行这些不变量。

### 3. Retrieval & Tool Layer

- 承载文档检索、附件/Markdown/JSON 解析、项目语义转换和本地合法候选生成。
- 能规则化的逻辑优先留在工具层，例如 ghost completion 的合法候选空间和 recent-actions suppress 信号。
- 工具层只生成证据或候选动作，不直接写业务真相源。

### 4. Model Runtime Layer

- Teacher models 用于强基线、标注参考、蒸馏和复杂任务对照。
- Student models 用于本地化、小成本部署和回归实验。
- `RadishMind-Core` 负责理解、推理、结构化建议、候选排序、风险标记和可选图片输入理解。
- Image Generation Runtime 独立负责图片像素生成；主模型只输出结构化 image intent 和约束。

### 5. Rule Validation & Response Builder

- 校验响应结构、目标对象、风险等级、确认要求、citation 和 action shape。
- 对可规则化字段保持确定性，例如 `status / risk_level / requires_confirmation / proposed_actions / patch / issue code`。
- 当前 `task-scoped response builder` 仍是 `M4` 决策实验和 tooling 分工证据，不是 raw 模型晋级或 production contract。

### 6. Data / Evaluation / Training Pipeline

- 管理 eval sample、candidate record、audit、replay、offline eval、training sample manifest 和 review records。
- 真实模型输出、生成 JSONL 和大体积实验产物默认留在 `tmp/`。
- 训练、蒸馏和模型晋级必须同时看 raw 输出、后处理轨、离线评测、自然语言 audit、人工 review 和 holdout 泄漏边界。

## 当前架构映射

- `Frontend UI`：`React + Vite + TypeScript`
- `Runtime Service`：`scripts/run-copilot-inference.py`、`services/gateway/copilot_gateway.py`、`scripts/run-platform-bridge.py`
- `Platform Service Layer`：`services/platform/`，使用 `Go` 承载 `HTTP API`、`gateway`、鉴权、流式转发、长驻进程、观测和部署壳；当前已落第一版 bridge-backed northbound
- `Southbound Provider Layer`：`services/runtime/provider_registry.py`、`services/runtime/inference_provider.py`
- `Conversation & Session`：`contracts/session-record.schema.json`、`contracts/session-recovery-checkpoint*.schema.json`、northbound session metadata 和平台 checkpoint metadata-only route smoke
- `Tooling Framework`：`contracts/tool*.schema.json`、tool registry / audit fixture、`scripts/check-tooling-framework-contract.py`、`scripts/check-session-recovery-checkpoint-contract.py` 和各类 deterministic builder/check
- `Evaluation & Governance`：`scripts/check-repo.py`、`scripts/check-radishflow-service-smoke-matrix.py`、offline eval、review records
- `Model Runtime`：`services/runtime/`、`scripts/run-radishmind-core-candidate.py`

## 当前缺口

- 当前只有 first-pass `Go` platform service 和 bridge-backed `HTTP API`，还不是 production deployment
- northbound `/v1/chat/completions`、`/v1/responses`、`/v1/messages` 与 `/v1/models` 已具备第一版兼容接口；当前已补第一版 SSE 流式兼容骨架、bridge-backed provider/profile inventory、request-side provider/profile selection、流式增量转发、`/v1/models` 列表 + 精确 lookup、结构化 diagnostics、discoverability 对齐、请求级观测和错误分类，但生产部署边界还未正式落地
- `HuggingFace` 与 `Ollama` 已进入 provider/profile inventory 和 diagnostics 门禁，但正式 secret backend、环境隔离和外部 provider 健康探测仍未补齐
- 已有 session/tooling 首版契约和 metadata-only 门禁，但没有 durable session store、长期记忆、真实 checkpoint storage backend 或跨轮恢复执行器
- 已有 tool registry、tool audit 和 metadata-only result cache 契约，但没有真实工具执行器、materialized result reader、durable tool store 或上层确认流接线
- 尚未具备 production secret backend、process supervisor、正式部署环境隔离和可发布部署包

这些缺口说明：`P1 Runtime Foundation` 已达到 short close，当前不应继续在 provider/config/diagnostics/observability 同层无限细化；下一步应把主要实现重心切到 session 与 tooling。

## 当前进度

- `contracts/` 已具备 Copilot request / response / gateway envelope / training sample / image generation intent / backend request / artifact schema。
- `RadishFlow` 的 gateway demo、service smoke matrix、UI consumption 和 candidate edit handoff 已作为未来接入门禁保留；在上层项目尚未具备真实接入能力前，当前只收口前置条件与阻塞项，不继续细化新的接线设计或模拟接入 summary。
- `suggest_flowsheet_edits` 与 `suggest_ghost_completion` 的真实 candidate record、audit、replay 和治理链已阶段性收口；新增真实 capture 需要先说明非重复 drift 假设。
- `RadishMind-Core` 本地小模型观测显示 raw 仍 blocked；broader review 的 15/15 `reviewed_pass` 与 `3B/4B` guided capacity review 当前只保留为路线证据，在没有新假设前不再默认继续扩 `M4` 实验面。
- `RadishMind-Image Adapter` 已具备 intent、backend request、artifact metadata 与最小评测 manifest；暂不调用真实生图 backend。

## 工程约束

- 层之间通过 schema、明确数据类型和稳定函数边界连接，避免隐式全局状态或字符串拼装协议。
- 代码优先使用对应语言的标准库和惯用结构；本仓库 Python 代码应保持直接、可测试、易读。
- 方法名和模块名必须说明真实职责；不要用空泛 helper、manager、processor 掩盖边界不清。
- 不为简单调用链增加多层抽象；只有当职责稳定、复用真实存在或复杂度明显下降时才抽 helper、builder 或 adapter。
- 修复结构漂移时优先修正 schema、builder 或任务边界，不用无限 fallback 包裹模型输出。
