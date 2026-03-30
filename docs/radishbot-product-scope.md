# RadishBot 产品范围与目标

更新时间：2026-03-30

## 项目目标

`RadishBot` 的目标不是单纯训练一个“会看图的模型”，也不是替代上层业务内核，而是构建一个能服务于 `RadishFlow` 与 `Radish` 的外部智能层仓库。

这个仓库至少应覆盖：

- 统一 Copilot 协议与上下文打包
- 基于结构化状态、文档、附件和可选图像的多模态理解
- 结构化输出：问题、解释、证据、候选动作、风险等级与确认要求
- 检索、工具调用、规则校验与模型编排
- 数据集、评测、蒸馏和 student 模型实验

## 基于真实上下文的统一定位

在审查 `D:\Code\RadishFlow` 与 `D:\Code\Radish` 后，第一阶段应把 `RadishBot` 定位为：

- 面向两个项目的外部智能层，而不是业务真相源
- 默认读状态、读文档、读附件并给出结构化建议，而不是直接执行高风险写入
- 以“协议、上下文、评测、规则”优先，而不是先锁定某个最终模型或部署形态
- 同时支持 `Copilot`、`VLM` 和工具调用，但不要求每个任务都走图像路线

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
  - 控制面状态解释和离线授权边界
- `Radish`
  - 文档语义、论坛内容、Console 权限知识和附件协议
  - 内容结构化建议和运营辅助

## minimind-v 的位置

基于当前上下文，`minimind-v` 更适合作为：

- student/base 候选
- M4 阶段的实验底座
- 与教师模型对比的域内蒸馏路线

当前不应把它先写死为第一阶段主模型路线。第一阶段先冻结任务、协议和评测，再决定它是否继续作为长期主线的一部分。

## 当前仍缺的决策

- `Radish` 的第一批内部落点到底先选文档、Console 还是论坛创作辅助
- 第一批任务的任务卡粒度和样本标注格式
- 教师模型与 student/base 的对照方式
- `RadishFlow` 何时从“状态优先”扩展到“状态 + 截图并重”
- 候选动作的 patch 结构在两个项目中分别如何落地

## 第一阶段明确不做

- 直接替代 `RadishFlow` 求解器、控制面或 CAPE-OPEN 适配层
- 直接接管 `Radish` 的 Auth / Gateway / API / Console 业务逻辑
- 让模型直接成为上层项目的真相源
- 在没有评测基线前就围绕底座模型频繁换路线
- 为了追求“统一”而忽略两个项目的语义差异
- 先把所有任务都压成截图理解或自治代理

## 成功标准

如果 `RadishBot` 在第一阶段满足以下条件，则方向通常正确：

- `RadishFlow` 侧至少有一个基于真实状态模型的可用 Copilot 场景落地
- `Radish` 侧至少有一个基于真实文档/内容体系的可用问答或辅助场景落地
- 已冻结统一输入输出协议的第一版，并允许项目专属上下文扩展
- 已建立小规模但可重复的评测集和任务指标
- 模型输出能被规则层或业务层复核，并明确要求确认
- `minimind-v` 已作为 student/base 候选进入对比，而不是在缺评测时被过早定为主线
