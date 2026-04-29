# RadishMind 产品范围与目标

更新时间：2026-04-29

## 项目目标

`RadishMind` 的目标不是单纯训练一个“会看图的模型”，也不是替代上层业务内核，而是构建一个能服务于 `RadishFlow` 与 `Radish` 的外部智能层仓库。

更准确地说，`RadishMind` 的长期目标应定义为“受控 Copilot / Agent 系统 + 可替换模型能力”，而不是“单一模型项目”。模型负责理解、生成和推理；Agent 层负责上下文打包、任务路由、工具调用、规则校验、权限边界、候选动作输出和评测闭环。

这个仓库至少应覆盖：

- 统一 Copilot 协议与上下文打包
- 面向任务的受控 Agent 编排、工具选择与响应校验
- 基于结构化状态、文档、附件和可选图像的多模态理解
- 结构化输出：问题、解释、证据、候选动作、风险等级与确认要求
- 检索、工具调用、规则校验与模型编排
- 数据集、评测、蒸馏和 student 模型实验
- `RadishMind-Core` 自研主模型路线，以及独立的图片生成 adapter / backend 编排能力

## 基于真实上下文的统一定位

在审查 `D:\Code\RadishFlow` 与 `D:\Code\Radish` 后，第一阶段应把 `RadishMind` 定位为：

- 面向两个项目的外部智能层，而不是业务真相源
- 默认读状态、读文档、读附件并给出结构化建议，而不是直接执行高风险写入
- 以“协议、上下文、评测、规则”优先，而不是先锁定某个最终模型或部署形态
- 同时支持 `Copilot`、`VLM` 和工具调用，但不要求每个任务都走图像路线

## Model 与 Agent 的长期定位

`model` 与 `agent` 在本项目中的边界应长期保持清晰：

- `model` 是可替换的推理能力，负责生成、理解、排序、归纳和多模态推理
- `agent` 是受控执行层，负责选择上下文、调用模型和工具、执行 schema / 规则校验、处理风险确认和沉淀评测记录
- `adapter` 负责把 `RadishFlow` / `Radish` 的业务状态转换为统一 `CopilotRequest`，并把结构化响应映射回各自项目
- `rule validation` 负责阻止模型输出越权、高风险直写或不符合业务契约的候选动作

`RadishMind-Core` 是本项目的自研主模型口径，但这里的“自研”默认指“基座适配型自研”，而不是从零预训练基础大模型。它应基于开源 `base model`，叠加 `RadishMind` 的指令数据、任务协议、`RadishFlow` / `Radish` 场景样本、风险标记和评测偏好，形成项目专属的理解、推理和结构化输出能力。

当前推荐目标为：

- 首个稳定目标：`3B` 或 `4B` 级 `RadishMind-Core`
- 长期本地部署上限：`7B` 级 `RadishMind-Core`
- 默认不把 `14B` / `32B` 作为自研主模型目标，除非后续部署硬件、训练数据和评测基线都显著升级
- 图片输入理解可以进入 `RadishMind-Core` 或其视觉适配路线；图片像素生成不进入主模型职责，默认由独立 `RadishMind-Image Adapter` 和生图 backend 承接

当前阶段应把这些能力放在同一个仓库中统一演进，因为协议、上下文、任务 prompt、评测数据和候选记录仍需要同步收口。模型权重、大体积训练产物和生产级推理镜像不应直接进入本仓库，只保留训练配置、适配代码、评测基线和模型选择策略。

后续当 agent runtime 与模型训练 / 部署开始出现明显不同的发布节奏、资产体量、团队职责或安全边界时，再评估拆分为独立仓库或独立发布单元。

## 第一阶段优先回答的问题

第一阶段优先回答以下问题：

- 能否把 `RadishFlow` 已存在的 `FlowsheetDocument`、选择集、诊断摘要、求解状态和控制面状态打包成稳定的 Copilot 上下文
- 能否在不侵入 `RadishFlow` 求解热路径的前提下，输出稳定、结构化、可校验的解释和候选编辑提案
- 能否把 `Radish` 已存在的固定文档、在线文档、论坛内容、Console 权限知识和附件协议接成可复用的问答与辅助能力
- 能否建立统一协议、任务评测和 student 路线，而不把两个项目强行压成同一种业务抽象

## 面向 `RadishFlow` 的首批高价值任务

### 首批任务

- `FlowsheetDocument + SelectionState + DiagnosticSummary` 驱动的问题解释与诊断问答
- 基于 `selected_units` / `selected_streams` 的候选编辑提案生成
- 基于单元放置、canonical ports 和邻近拓扑的 ghost port / ghost connection / ghost stream name 补全建议
- `SolveSessionState + SolveSnapshot` 驱动的运行状态说明、差异总结和下一步建议
- 控制面 / entitlement / lease / package sync 错误的人工可读解释
- 画布截图理解作为补充输入，在 `rf-canvas` 和真实 UI 边界更稳定后再增强

### 最适合喂给 Copilot 或 VLM 的上下文

- `FlowsheetDocument`
- `document_revision`
- `SelectionState`
- `DiagnosticSummary` / `DiagnosticSnapshot`
- `SolveSessionState` / `SolveSnapshot`
- `EntitlementSnapshot`、`PropertyPackageManifestList`、离线租约刷新结果与同步错误摘要
- `selected_unit`、`unconnected_ports`、`nearby_nodes`、`cursor_context` 与本地规则预先筛出的 `legal_candidate_completions`
- 可选 `canvas snapshot`

### 明确不能让 AI 侵入的部分

- `rf-thermo` / `rf-flash` / `rf-solver` 的数值求解热路径
- `.NET 10` CAPE-OPEN/COM 适配边界
- OIDC token、credential handle、auth cache 真值和包体完整性校验
- 未经确认直接改写 `FlowsheetDocument` 或执行高风险命令

## 面向 `Radish` 的首批高价值任务

### 首批任务

- 基于固定 `Docs/` 与在线文档的检索增强问答
- Console / 权限 / 运营流程说明与路径引导
- 论坛帖子、评论、在线文档的结构化摘要、标题、标签和分类建议
- 围绕 `attachment://{id}` 与 `/_assets/attachments/*` 的附件解释与引用辅助
- 截图和附件内容理解作为补充能力，而不是先上自动治理或自动运营

### 最适合喂给 Copilot 或 VLM 的上下文

- 固定文档 `Docs/`
- 在线文档 `WikiDocument`
- 论坛帖子、评论和 Markdown 正文
- Console 架构、权限矩阵和运营文档
- `attachment://{id}` 引用与运行时附件地址
- 当前查看的应用、路由、角色和权限摘要

### 明确不能让 AI 侵入的部分

- `Radish.Auth` 的身份真相源和 token 生命周期
- `Radish.Api` / `Radish.Gateway` / `Console` 的权限最终判定
- 附件访问控制、临时访问令牌和公开资源守卫
- 未经确认直接替代人工做治理结论、封禁、授权或数据写入

## 共享能力与项目专属能力

### 共享能力

- 统一输入输出协议
- 请求追踪、审计和版本标记
- 检索、工具调用、规则校验和引用生成
- 多模型路由、Teacher / Student 对比和离线评测
- 风险分级、`requires_confirmation` 和候选动作骨架

### 项目专属能力

- `RadishFlow`
  - flowsheet 语义、修订号、选择集、诊断与候选编辑
  - 基于本地合法候选集的 ghost 补全排序与命名建议
  - 控制面状态解释和离线授权边界
- `Radish`
  - 文档语义、论坛内容、Console 权限知识和附件协议
  - 内容结构化建议和运营辅助

## RadishMind-Core 与模型分工

基于当前上下文，`RadishMind-Core` 是项目自研主模型目标，`minimind-v` 当前正式作为它的默认 `student/base` 主线候选：

- 默认 `student/base` 主线
- M4 阶段的实验底座
- 与教师模型对齐后的域内蒸馏承接路线
- `3B` / `4B` 首版自研主模型的优先适配对象；若评测证明不足，再评估 `7B` 档位

与之配套的首轮模型分工为：

- `Qwen2.5-VL`：默认 `teacher` / 多模态强基线，用于高质量对照评测、复杂图文任务 PoC 和蒸馏参考
- `SmolVLM`：默认轻量本地对照组，用于验证小模型下限、资源敏感部署和轻量回归基线
- `RadishMind-Image Adapter`：负责把 `RadishMind-Core` 的结构化图片生成意图转换为 backend request，并记录 prompt、尺寸、风格、seed、负面词、编辑约束、安全门禁和 trace metadata
- `Image Generation Backend`：负责真正生成图片像素；首轮不从零训练，优先参考或接入 `SD1.5`、`PixArt-δ 0.6B`，中期再评估 `Segmind-Vega` 或 `SD3.5 Medium 2.5B`

当前阶段不再把 `minimind-v` 仅写成“候选”；默认路线是先围绕它建立领域适配与训练承接，再由离线评测结果决定是否调整主线。图片生成能力应作为 `RadishMind` 的工具 / backend 能力交付，而不是要求 `RadishMind-Core` 同时承担 Copilot 推理、协议遵循和像素生成。

由于 `RadishFlow` 与 `Radish` 暂时都还没有进入真实模型 / Agent 接入阶段，当前不把上层真实接线作为 `RadishMind` 的阻塞项。`RadishMind` 这边应先完成以下自身资产：

- `RadishMind-Core` 首版基座评估：先比较 `3B` / `4B` 的协议遵循、中文任务理解、结构化响应、citation 对齐和本地部署成本，再决定是否进入 `7B`
- 训练 / 蒸馏样本格式：以 `CopilotRequest -> CopilotResponse` 为核心，保留 `project / task / artifacts / context / safety / proposed_actions / citations / requires_confirmation`
- 训练样本转换入口：当前已能从 committed eval 样本的 `input_request + golden_response` 生成首批 9 条 `CopilotTrainingSample` JSONL，也能从 audit pass candidate record 生成首批 9 条 `teacher_capture` 样本，覆盖 `suggest_flowsheet_edits`、`suggest_ghost_completion` 与 `answer_docs_question`
- teacher / student / lightweight baseline 对照矩阵：`Qwen2.5-VL` 给出强基线和蒸馏参考，`minimind-v` 承接主线适配，`SmolVLM` 验证低资源下限
- `RadishMind-Image Adapter` 第一版 schema 与最小评测 manifest：主模型只输出图片生成意图、约束和审查信息，Adapter 再生成 backend request，图片像素生成交给独立 backend，结果以 artifact metadata 回到 `RadishMind`；当前评测 manifest 只覆盖结构化意图、backend request 映射、artifact metadata、safety gate 与 provenance，不评价图片像素质量
- 未来接入清单：保留现有 gateway smoke、UI consumption summary 与 candidate handoff summary 作为 `RadishFlow` / `Radish` 准备好后的验收门禁

## 当前仍缺的决策

- `Radish` 的第一批内部落点到底先选文档、Console 还是论坛创作辅助
- 第一批任务的任务卡粒度和样本标注格式
- `Qwen2.5-VL` 在当前任务集中的首选尺寸与推理预算
- `SmolVLM` 进入默认回归矩阵的任务边界
- `RadishFlow` 何时从“状态优先”扩展到“状态 + 截图并重”
- 候选动作的 patch 结构在两个项目中分别如何落地
- `RadishMind-Image Adapter` 已具备 intent、backend request、artifact metadata 三段 schema 和最小图片生成评测 manifest；后续仍需真实 backend 包装

## 远期备忘方向

以下内容当前仅作为未来想法备忘，用于防止后续遗忘；它们不改变现阶段路线图、阶段优先级、任务卡范围或退出标准。

### `RadishFlow`

- 在现有 `suggest_ghost_completion` 与 `suggest_flowsheet_edits` 之外，未来可进一步探索“像代码补全一样”的流程链补全能力：用户放入一个模块后，系统继续预测下一步更可能补上的单元、连接和局部结构
- 未来可探索围绕塔、分离器、换热器等典型对象的参数优化建议，但仍应坚持“建议 / 候选动作优先”，不直接侵入数值求解热路径
- 未来可探索参数输入错误提示、参数组合不合理预警和纠错辅助，把“诊断解释”继续前推到“输入前校验 + 输入后纠偏”这一层

### `Radish`

- 在现有文档问答、内容摘要和论坛元数据建议之外，未来可探索“社区小助手”方向，例如审贴辅助、回复草稿、社区互动建议和运营协作
- 对论坛或社区场景，未来可探索自动回复建议、跟帖建议和基于上下文的运营话术草稿，但默认仍应先以“人工确认后发送”为边界
- 更远期才评估更强的账号代操作能力，例如代发帖、代回复、代执行部分社区管理动作；若进入该方向，必须额外补强审计、权限边界、风险分级与显式确认机制

## 第一阶段明确不做

- 直接替代 `RadishFlow` 求解器、控制面或 CAPE-OPEN 适配层
- 直接接管 `Radish` 的 Auth / Gateway / API / Console 业务逻辑
- 让模型直接成为上层项目的真相源
- 在没有评测基线前就围绕底座模型频繁换路线
- 把 `RadishMind-Core` 定义成必须从零训练的基础大模型
- 让 `RadishMind-Core` 直接承担图片像素生成，导致主模型训练、评测和部署目标失控
- 为了追求“统一”而忽略两个项目的语义差异
- 先把所有任务都压成截图理解或自治代理
- 把 `RadishMind` 过早拆成模型仓库和 agent 仓库，导致协议、评测和任务边界提前分叉

## 成功标准

如果 `RadishMind` 在第一阶段满足以下条件，则方向通常正确：

- `RadishFlow` 侧至少有一个基于真实状态模型的可用 Copilot 场景落地
- `Radish` 侧至少有一个基于真实文档/内容体系的可用问答或辅助场景落地
- 已冻结统一输入输出协议的第一版，并允许项目专属上下文扩展
- 已建立小规模但可重复的评测集和任务指标
- `Radish` docs QA 已具备外部 record 回灌、负例回放与首批跨样本真实 replay，后续可以继续扩大真实坏输出批次
- 模型输出能被规则层或业务层复核，并明确要求确认
- `minimind-v` 已作为默认 `student/base` 主线进入实际评测与适配，而不是继续停留在候选状态
- `agent` 层已经能稳定证明模型可替换：同一任务至少能在 mock / teacher / student 或多 provider 之间比较，并仍保持统一协议、规则校验和评测口径
