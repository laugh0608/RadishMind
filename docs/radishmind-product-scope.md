# RadishMind 产品范围与目标

更新时间：2026-05-06

## 当前定位

`RadishMind` 是 `Radish` 体系下的外部智能层仓库，不是上层业务真相源，也不是单一大模型仓库。它的目标是提供受控 Copilot / Agent 能力、统一协议、评测闭环和可替换模型路线。

当前能力边界：

- 读状态、读文档、读附件和可选图像，输出解释、诊断、结构化建议和候选动作。
- 高风险动作必须保留 `requires_confirmation`，由人工确认或上层规则层复核后再执行。
- 模型负责理解、推理、排序、归纳和生成建议；agent / adapter / rule validation 负责上下文打包、工具调用、结构校验、权限边界和评测记录。
- `RadishMind-Core` 是基座适配型自研主模型路线，不是从零预训练基础大模型。
- 图片像素生成不并入主模型职责，默认由 `RadishMind-Image Adapter` 与独立 backend 承接。

## 当前阶段

近期推进集中在 `M3/M4` 收口：

- 维持 `RadishFlow` gateway、service smoke、UI consumption 与 candidate handoff 作为未来上层接入门禁。
- 继续推进 `RadishMind-Core` 结构化输出决策实验，明确 raw、repair、hard-field injection、task-scoped response builder、natural-language audit 与 human review 的职责边界。
- 训练 / 蒸馏样本只提交 manifest、summary、复核策略和实验说明；生成的 JSONL 和真实模型产物默认留在 `tmp/`。
- 不继续把同类真实 capture 或 prompt/scaffold 加长当作默认主线。

## 项目优先级

### `RadishFlow`

首批能力优先围绕结构化流程状态：

- `FlowsheetDocument + SelectionState + DiagnosticSummary` 的解释与诊断问答。
- 基于选中对象的候选编辑提案。
- 基于本地合法候选集的 ghost port / ghost connection / ghost stream name 补全建议。
- 求解状态、控制面、entitlement、package sync 与离线授权摘要解释。

不可侵入：

- 数值求解热路径。
- CAPE-OPEN / COM 适配边界。
- token、credential、auth cache 和包体完整性真相。
- 未经确认直接改写 `FlowsheetDocument`。

### `Radish`

首批能力优先围绕知识与内容：

- 固定文档、在线文档、论坛内容和 Console 权限知识的检索增强问答。
- 文档、帖子、评论和附件的摘要、标题、标签、分类与引用辅助。
- 当前页面、路由、角色和权限摘要解释。

不可侵入：

- 身份、token 生命周期和权限最终判定。
- 附件访问控制和临时访问令牌。
- 未经确认直接执行治理、封禁、授权或数据写入。

### `RadishCatalyst`

当前只做文档级预留：

- 可预留玩家知识问答、进度解释、生产链规划和开发侧静态数据一致性检查。
- 不扩真实 schema、adapter、gateway smoke 或模型接线，直到上层明确首个真实任务面。

不可侵入：

- Godot 运行时、存档、任务、战斗、掉落、配方、公开等级和联机权威。
- 玩家侧不能泄露 `internal` 或默认隐藏的 `spoiler` 内容。

## 当前非目标

- 不把 `RadishMind` 做成替代业务内核的自治系统。
- 不让模型替代 `RadishFlow` 求解、`Radish` 权限判定或 `RadishCatalyst` 游戏权威。
- 不在没有评测和人工复核边界前扩大训练规模。
- 不默认下载大模型、数据集或权重。
- 不把 `14B/32B` 写成当前自研主模型默认目标；首版仍优先 `3B/4B`，长期本地部署上限 `7B`。

## 实现原则

- 协议优先采用结构化 JSON。
- 能规则化或工具化的逻辑，不强行压给模型。
- 代码遵循对应语言的惯用实践，命名清楚、职责明确、边界稳定。
- 禁止语义不明的方法、空转 helper、过度泛化的 manager/factory 和晦涩抽象封装。
- 抽象必须服务于真实职责边界、复用或复杂度收敛；不能为了隐藏简单逻辑而增加理解成本。
