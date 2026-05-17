# RadishMind 产品范围与目标

更新时间：2026-05-17

## 核心定义

`RadishMind` 是 `Radish` 体系下的协议驱动、可审计、可本地部署、可工具化的 Copilot / Agent runtime platform。

更强的项目定义、三层定位、northbound/southbound 战略、service mode 分级与 action safety ladder 见 [战略定义](radishmind-strategy.md)。

它不是上层业务真相源，不是单一大模型仓库，也不只是等待真实接入的中转站。它的职责是把上层项目上下文、局部规则、工具能力、模型推理和审计治理收口为同一条可复跑的运行链路。

当前核心边界：

- 读状态、读文档、读附件和可选图像，输出解释、诊断、结构化建议和候选动作。
- 高风险动作必须保留 `requires_confirmation`，由人工确认或上层规则层复核后再执行。
- 对内保持统一 canonical protocol，对外兼容常见模型调用协议和常见 AI 服务协议，而不是被单一厂商接口绑死。
- 模型负责理解、推理、归纳、排序和建议生成；runtime、adapter、tooling、rule validation 和 audit 负责上下文打包、工具调用、结构校验、权限边界和可追溯性。
- `RadishMind-Core` 是基座适配型自研主模型路线，不是从零预训练基础大模型。
- 图片像素生成不并入主模型职责，默认由 `RadishMind-Image Adapter` 与独立 backend 承接。

## 项目范围

### 1. `Runtime Service`

- 提供最小可运行的推理入口、gateway、route 识别、provider/profile 选择、响应封装和本地产品 discovery 面。
- 当前已有 CLI runtime、进程内 Python gateway、最小 `Go` HTTP bridge、本地启动 runbook、`GET /v1/platform/overview` 只读产品 overview、session/tooling metadata shell 和 blocked action shell；后续要补更完整的本地 console、生产部署边界和外部 provider health check，服务层与 `gateway` 不锁死在单一语言上，可按职责采用 `Go`。
- 这一层必须同时覆盖两条方向：
  - 北向协议兼容：对外提供 native Copilot API，以及 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models`、`/v1/platform/overview`、session/tooling metadata shell 这类常见兼容接口和只读产品发现接口。
  - 南向模型接入：对内接入 `RadishMind-Core`、`local_transformers / HuggingFace`、`Ollama`、OpenAI-compatible、Gemini native、Anthropic messages 等 provider / transport。

### 2. `Conversation & Session`

- 管理 `conversation_id`、会话上下文、历史压缩、恢复和审计边界。
- 当前已有 session contract、history/state policy、checkpoint metadata-only read route 和 `GET /v1/session/metadata`；这些只支撑 metadata / overview 展示，不代表 durable session store、长期记忆或 replay executor 已启用。

### 3. `Tooling Framework`

- 承载检索、附件解析、项目语义转换、本地候选生成、response builder 与工具策略。
- 当前已有通用 tool contract、registry、audit、`GET /v1/tools/metadata` 和 blocked `POST /v1/tools/actions`；它们只支撑 contract-only 展示和 blocked action response，不代表真实工具执行器、confirmation flow 或业务写回已启用。

### 4. `Evaluation & Governance`

- 负责 schema、smoke、offline eval、candidate record、review、promotion gate 和仓库级检查。
- 目标不是只看机器指标，而是让 advisory-only、confirmation、route、citation 和 review 边界能长期复跑。

### 5. `Model Adaptation`

- 负责基座选型、prompt/runtime 协同、蒸馏、训练样本治理、模型晋级与回归。
- 当前重点仍是“配合平台契约收口模型路线”，不是先扩大训练规模。
- 自研模型只是平台的一类 provider，不应和 `HuggingFace`、`Ollama`、OpenAI-compatible 或其它外部模型接入能力互相替代。

### 6. `Image Path`

- 主模型只输出结构化 image intent、约束、审查和 artifact metadata。
- 真正的图片生成由独立 image adapter 和 backend 承接。

### 7. 上层项目接入面

- `RadishFlow`、`Radish`、`RadishCatalyst` 是应用面，不是项目本体的全部意义。
- 这些接入面复用同一套 runtime、contract、tooling、evaluation 和 governance，而不是各自私接模型。

## 当前阶段判断

- 历史上的 `M3` service/API smoke 与 `M4` broader review、`3B/4B` capacity review 已经收口为冻结证据。
- 当前正式主线切换为“平台重定义 + 平台基础能力建设”，不再把“继续深挖同一批实验”或“提前设计不存在的真实接线”当作默认推进方式。
- 当前实现焦点已进入 `P3 Local Product Shell / Ops Surface`：先用 `/v1/platform/overview` 和 overview consumer smoke 固定本地 console 可展示能力，再补真正的只读 console 壳。
- 训练 / 蒸馏样本继续只提交 manifest、summary、复核策略和实验说明；生成的 JSONL 和真实模型产物默认留在 `tmp/`。

## 当前优先支持的应用面

### `RadishFlow`

首批能力继续围绕结构化流程状态：

- `FlowsheetDocument + SelectionState + DiagnosticSummary` 的解释与诊断问答。
- 基于选中对象的候选编辑提案。
- 基于本地合法候选集的 ghost port、ghost connection、ghost stream name 补全建议。
- 求解状态、控制面、entitlement、package sync 与离线授权摘要解释。

当前不可侵入：

- 数值求解热路径。
- CAPE-OPEN / COM 适配边界。
- token、credential、auth cache 和包体完整性真相。
- 未经确认直接改写 `FlowsheetDocument`。

### `Radish`

首批能力继续围绕知识与内容：

- 固定文档、在线文档、论坛内容和 Console 权限知识的检索增强问答。
- 文档、帖子、评论和附件的摘要、标题、标签、分类与引用辅助。
- 当前页面、路由、角色和权限摘要解释。

当前不可侵入：

- 身份、token 生命周期和权限最终判定。
- 附件访问控制和临时访问令牌。
- 未经确认直接执行治理、封禁、授权或数据写入。

### `RadishCatalyst`

当前只做文档级预留：

- 可预留玩家知识问答、进度解释、生产链规划和开发侧静态数据一致性检查。
- 不扩真实 schema、adapter、gateway smoke 或模型接线，直到上层明确首个真实任务面。

当前不可侵入：

- Godot 运行时、存档、任务、战斗、掉落、配方、公开等级和联机权威。
- 玩家侧不能泄露 `internal` 或默认隐藏的 `spoiler` 内容。

## 当前非目标

- 不把 `RadishMind` 做成替代业务内核的自治系统。
- 不让模型替代 `RadishFlow` 求解、`Radish` 权限判定或 `RadishCatalyst` 游戏权威。
- 不把通用 unrestricted tool calling 当成当前默认能力。
- 不把平台锁死在单一模型、单一 provider、单一上游协议或单一对外接口上。
- 不在 runtime、session、tooling 契约还没稳定前扩大训练规模。
- 不默认下载大模型、数据集或权重。
- 不把 `14B/32B` 写成当前自研主模型默认目标；首版仍优先本地可承受的小中型路线，长期本地部署上限暂定 `7B`。

## 实现原则

- 协议优先采用结构化 JSON。
- 能规则化或工具化的逻辑，不强行压给模型。
- 先把 runtime、session、tooling、governance 边界做清楚，再扩大训练和接入面。
- 代码遵循对应语言的惯用实践，命名清楚、职责明确、边界稳定。
- 禁止语义不明的方法、空转 helper、过度泛化的 manager/factory 和晦涩抽象封装。
- 抽象必须服务于真实职责边界、复用或复杂度收敛；不能为了隐藏简单逻辑而增加理解成本。
