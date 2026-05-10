# 跨项目任务卡：接入前置条件总表

更新时间：2026-05-10

## 文档目的

本表用于回答一件事：在 `RadishFlow`、`Radish`、`RadishCatalyst` 当前都还没有形成真实接入能力的前提下，`RadishMind` 应该停在哪里，哪些条件满足后才值得重新打开 `M3` 接入准备。

它不是接线设计文档，也不是 UI/命令层方案草图；职责是冻结“当前不能继续前推的原因”和“未来何时才该继续”。

## 判定原则

只有同时具备以下条件，某个上层项目才算进入“可开始真实接入准备”状态：

- 上层明确给出真实的调用挂载点，而不是只存在抽象任务名或文档预留。
- 上层明确给出用户确认流，尤其是 `requires_confirmation` 的落点。
- 上层明确给出命令层 / handoff 承接接口，或等价的只读展示边界。
- 上层明确给出审计日志或排障落点。

只具备 schema、fixture、smoke、summary 或评测资产，不等于已经具备真实接入能力。

## 当前总表

| 项目 | 当前成熟度 | 已有资产 | 主要阻塞项 | 当前结论 |
| --- | --- | --- | --- | --- |
| `RadishFlow` | `M3` 门禁已冻结，但未进入真实接入 | gateway、service smoke matrix、UI consumption summary、candidate handoff summary | 缺真实 UI 挂载点、确认流、命令层承接接口、审计落点 | 维持现有门禁，不继续细化接线设计 |
| `Radish` | 任务/评测资产已存在，但未进入服务接入 | docs QA、文档检索增强、结构化问答评测、训练转换资产 | 缺真实调用挂载点、消费视图、确认/审计边界，且当前主线不在这里 | 维持任务与评测资产，不前推接入 |
| `RadishCatalyst` | 仅文档级预留 | 项目边界、预留任务面、禁止事项口径 | 连首个真实任务、schema、adapter、eval、gateway smoke 都没有 | 不进入接入准备 |

## `RadishFlow`

当前是三个项目里最接近接入的一条线，但也只到“门禁足够稳定”这一步。

已冻结资产：

- `radishflow/suggest_flowsheet_edits` 的 gateway envelope
- service/API smoke matrix
- advisory-only 的 UI consumption summary
- non-executing 的 candidate handoff summary

阻塞项：

- `RadishFlow` 里没有明确的 proposal view 真实挂载点
- 没有真实的 `candidate_edit` 用户确认流
- 没有承接 handoff proposal 的命令层接口
- 没有明确的审计记录与排障入口

当前动作：

- 继续维护现有门禁
- 不新增模拟 UI / 命令层 summary
- 等上层给出真实接点后再重开接入准备

## `Radish`

当前更像“任务资产和评测资产已具备”，而不是“服务/API 接入准备已开始”。

已有资产：

- `answer_docs_question` 任务卡与评测样本
- 文档检索增强和结构化问答相关治理资产
- 训练样本转换与离线评测相关入口

阻塞项：

- 没有明确的真实调用挂载点
- 没有明确的消费视图或确认边界
- 没有需要承接高风险 handoff 的真实命令层语义
- 当前仓库主线仍优先 `RadishFlow`

当前动作：

- 保持 docs QA / retrieval / structured QA 资产稳定
- 不把 `Radish` 提前拉进 `M3` 服务接入准备

## `RadishCatalyst`

当前仍然只适合停在文档预留。

已有资产：

- 项目边界定义
- 可预留的任务方向
- 明确的禁止事项和游戏权威边界

阻塞项：

- 没有首个真实任务面
- 没有 schema、adapter、eval sample、gateway smoke
- 没有可以讨论 UI、确认流或命令层的位置

当前动作：

- 继续只保留文档级口径
- 不把它拉入任何真实接入准备、schema 扩展或 smoke 扩展

## `RadishMind` 当前应做什么

- 维护现有 `M3` 门禁和正式文档真相源
- 把“尚未具备真实接入能力”的停止线写清楚
- 在出现新的上层接点前，避免继续细化不存在的接线方案

## `RadishMind` 当前不要做什么

- 不为任一上层项目继续扩新的模拟 UI / 命令层 summary
- 不在没有真实挂载点时设计字段映射、确认流文案或 handoff 接口
- 不因为 `M4` 证据已经收口，就误判为上层已经可以接

## 重新打开接入准备的触发条件

只有满足以下任一条件，才重新进入更具体的 `M3` 接入准备：

- 某个上层项目明确给出真实 UI / 页面 /交互挂载点
- 某个上层项目明确给出确认流与命令层承接接口
- 现有门禁在真实试点里暴露出新的非重复 drift

在这些条件出现前，当前最稳妥的结论就是：门禁保留，设计停止，等待上层成熟度变化。
