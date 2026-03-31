# RadishMind 文档总览

更新时间：2026-03-31

## 文档目的

本目录用于沉淀 `RadishMind` 的产品定位、系统架构、阶段路线和跨项目集成边界，作为后续设计与实现的真相源入口。

当前版本已经基于对 `D:\Code\RadishFlow` 与 `D:\Code\Radish` 的只读审查完成了一轮收口，重点不是先定实现栈，而是先把任务边界、协议、数据和评测讲清楚。

## 当前优先文档

1. [产品范围与目标](radishmind-product-scope.md)
2. [系统架构草案](radishmind-architecture.md)
3. [阶段路线图](radishmind-roadmap.md)
4. [跨项目集成契约草案](radishmind-integration-contracts.md)
5. [ADR 0001: 分支与 PR 治理](adr/0001-branch-and-pr-governance.md)
6. [开发日志说明](devlogs/README.md)
7. [RadishFlow 首批任务卡](task-cards/README.md)
8. [统一契约文件说明](../contracts/README.md)
9. [数据集与评测目录说明](../datasets/README.md)

## 当前规划原则

- `RadishMind` 是独立仓库，不与业务仓库强耦合
- 优先建设统一协议、上下文打包、评测和规则框架
- 优先支持 `RadishFlow`，但第一批能力以结构化状态与诊断解释为主，不把截图路线写成唯一入口
- `Radish` 的第一批任务以文档问答、Console/运营辅助和内容结构化建议为主
- 先做“可解释、可确认、可回退”的建议系统，再追求强自治
- 模型能力与工具能力并重，不把问题都压给单一模型

## 当前仍缺的关键决策

- `Radish` 第一批到底先落在哪个真实入口：
  - 文档应用问答
  - Console 能力解释
  - 论坛创作辅助
- 在 `JSON Schema` 之外，是否同步生成 TypeScript 类型或其他契约产物
- 第一批评测集的任务粒度、标注格式与通过阈值
- 教师模型对照组、student/base 路线与推理路由策略
- `RadishFlow` 截图/VLM 路线进入主线的触发条件

## 下一步优先推进

1. 为 `RadishFlow` 首批 3 个任务补首批输入快照样例和评测样本。
2. 在 `contracts/` 基础上补充 schema 校验示例与后续类型生成策略。
3. 为 `Radish` 只选 1 个最小入口先落地，默认优先“固定文档 + 在线文档问答”。
4. 在 `datasets/eval/` 基础上补最小回归脚本，再进入模型对照与 PoC。
