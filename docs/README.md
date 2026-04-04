# RadishMind 文档总览

更新时间：2026-04-04

## 文档目的

本目录用于沉淀 `RadishMind` 的产品定位、系统架构、阶段路线和跨项目集成边界，作为后续设计与实现的真相源入口。

当前版本已经基于对 `D:\Code\RadishFlow` 与 `D:\Code\Radish` 的只读审查完成了一轮收口；当前已冻结 `Python` 主实现栈，并将首轮模型路线收口为 `minimind-v` 主线、`Qwen2.5-VL` 强基线和 `SmolVLM` 轻量对照组。

## 当前优先文档

1. [产品范围与目标](radishmind-product-scope.md)
2. [系统架构草案](radishmind-architecture.md)
3. [阶段路线图](radishmind-roadmap.md)
4. [跨项目集成契约草案](radishmind-integration-contracts.md)
5. [ADR 0001: 分支与 PR 治理](adr/0001-branch-and-pr-governance.md)
6. [开发日志说明](devlogs/README.md)
7. [首批任务卡](task-cards/README.md)
8. [统一契约文件说明](../contracts/README.md)
9. [数据集与评测目录说明](../datasets/README.md)

## 当前规划原则

- `RadishMind` 是独立仓库，不与业务仓库强耦合
- 优先建设统一协议、上下文打包、评测和规则框架
- 优先支持 `RadishFlow`，但第一批能力以结构化状态与诊断解释为主，不把截图路线写成唯一入口
- `Radish` 的第一批任务以文档问答、Console/运营辅助和内容结构化建议为主
- 先做“可解释、可确认、可回退”的建议系统，再追求强自治
- 模型能力与工具能力并重，不把问题都压给单一模型
- 当前模型路线默认采用 `minimind-v` 主线 + `Qwen2.5-VL` 强基线 + `SmolVLM` 轻量对照组

## 当前仍缺的关键决策

- 在 `JSON Schema` 之外，是否同步生成 TypeScript 类型或其他契约产物
- 第一批评测集的任务粒度、标注格式与通过阈值
- `Qwen2.5-VL` 在当前任务上的首选尺寸、推理路由与成本上限
- `SmolVLM` 作为轻量对照组的准入任务和退场条件
- `RadishFlow` 截图/VLM 路线进入主线的触发条件
- `Radish` docs QA 在 manifest 导入流程落地后，真实 captured negative 批次应先补哪些优先违规类型

## 下一步优先推进

1. 继续为 `RadishFlow` 首批 3 个任务扩展输入快照样例与评测样本，优先补控制面冲突态和更多对抗样本。
2. 继续把 `Radish` 的 `answer_docs_question` 作为唯一最小入口推进；当前已接上真实 `candidate_response_record`、统一负例回放、跨样本真实 replay 和最小 manifest 导入流程，下一步主线转向扩大 captured negative 批次，仅按需补少量极端冲突样本。
3. 在 `contracts/` 基础上补充 schema 校验示例与后续类型生成策略。
4. 在 `datasets/eval/` 与最小回归 runner 基础上，再进入模型对照与 PoC。
