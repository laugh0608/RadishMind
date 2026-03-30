# RadishMind 文档总览

更新时间：2026-03-30

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
- 统一协议是否先落为 JSON Schema、TypeScript 类型还是其他契约形式
- 第一批评测集的任务粒度、标注格式与通过阈值
- 教师模型对照组、student/base 路线与推理路由策略
- `RadishFlow` 截图/VLM 路线进入主线的触发条件

## 下一步优先推进

1. 把 `RadishFlow` 的首批 3 个任务细化成正式任务卡与评测样本：
   - `explain_diagnostics`
   - `suggest_flowsheet_edits`
   - `explain_control_plane_state`
2. 把统一 `Copilot` 协议落成可执行契约文件，而不只停留在文档描述。
3. 为 `Radish` 只选 1 个最小入口先落地，默认优先“固定文档 + 在线文档问答”。
4. 在引入 student 路线前，先建立最小离线评测集和回归脚本。
