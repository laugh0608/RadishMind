# RadishMind 文档总览

更新时间：2026-04-28

## 文档目的

本目录用于沉淀 `RadishMind` 的产品定位、系统架构、阶段路线和跨项目集成边界，作为后续设计与实现的真相源入口。

当前版本已经基于对 `D:\Code\RadishFlow` 与 `D:\Code\Radish` 的只读审查完成了一轮收口；当前已冻结 `Python` 主实现栈，并将首轮模型路线收口为 `RadishMind-Core` 基座适配型自研主模型、`minimind-v` 默认 `student/base` 主线、`Qwen2.5-VL` 强基线和 `SmolVLM` 轻量对照组。
当前 `RadishFlow export -> adapter -> request` 主线也已从“只有扁平 adapter fixture”推进到“存在上游导出边界、bootstrap 模板、preflight smoke validator、batch smoke 与 committed exporter edge fixtures”的状态。
当前 `RadishFlow suggest_flowsheet_edits` 与 `suggest_ghost_completion` 都已具备真实批次、artifact summary、replay、recommended replay 与 real-derived negative 治理链；`suggest_flowsheet_edits v93` 与 `suggest_ghost_completion v25` 收口后，下一步主线转为服务/API 最小实现切片，而不是继续默认跑样本。
当前模型规模口径为：`RadishMind-Core` 首版优先 `3B` / `4B`，长期本地部署上限 `7B`；图片生成能力不并入主模型参数目标，默认通过 `RadishMind-Image Adapter` 和独立生图 backend 提供。`RadishMind-Core` 首版基座评估矩阵、离线评测阈值与离线评测运行记录已落成可回归门禁，用于固定 `minimind-v`、`3B/4B/7B`、`Qwen2.5-VL` 与 `SmolVLM` 的首轮评估边界、阻塞指标、样本选择和结果记录格式。

## 当前优先文档

1. [产品范围与目标](radishmind-product-scope.md)
2. [系统架构草案](radishmind-architecture.md)
3. [阶段路线图](radishmind-roadmap.md)
4. [跨项目集成契约草案](radishmind-integration-contracts.md)
5. [RadishMind-Core 首版基座评估矩阵](radishmind-core-baseline-evaluation.md)
6. [ADR 0001: 分支与 PR 治理](adr/0001-branch-and-pr-governance.md)
7. [开发日志说明](devlogs/README.md)
8. [首批任务卡](task-cards/README.md)
9. [统一契约文件说明](../contracts/README.md)
10. [数据集与评测目录说明](../datasets/README.md)
11. [脚本目录说明](../scripts/README.md)

## 当前规划原则

- `RadishMind` 是独立仓库，不与业务仓库强耦合
- 优先建设统一协议、上下文打包、评测和规则框架
- 优先支持 `RadishFlow`，但第一批能力以结构化状态与诊断解释为主，不把截图路线写成唯一入口
- 对 `RadishFlow` 的真实接线，优先先冻结 `export -> adapter -> request` 这条结构化链路，再考虑更重的服务编排或模型接线
- `Radish` 的第一批任务以文档问答、Console/运营辅助和内容结构化建议为主
- 先做“可解释、可确认、可回退”的建议系统，再追求强自治
- 模型能力与工具能力并重，不把问题都压给单一模型
- 当前模型路线默认采用 `RadishMind-Core` 基座适配型自研主模型 + `minimind-v` 主线 + `Qwen2.5-VL` 强基线 + `SmolVLM` 轻量对照组
- `RadishMind-Core` 默认负责理解、推理、结构化建议和图片生成意图，不直接生成图片像素；图片生成由独立 adapter / backend 承接

## 当前仍缺的关键决策

- 在 `JSON Schema` 之外，是否同步生成 TypeScript 类型或其他契约产物
- 第一批评测集的任务粒度和标注格式
- `Qwen2.5-VL` 在当前任务上的具体首选尺寸、推理路由与成本上限
- `SmolVLM` 作为轻量对照组的具体准入任务和退场条件
- `RadishMind-Core` 首版基座评估矩阵、阻塞阈值和离线评测记录格式已固定；具体观测结果和成本实测仍需后续补齐
- `RadishMind-Image Adapter` 的第一版 schema、backend 选择和最小评测样本
- `RadishFlow` 截图/VLM 路线进入主线的触发条件
- `RadishFlow export` 在更多真实 exporter 接线后，是否需要继续把 `selection` 排序、focus 归一与 `support_artifacts` 摘要策略升级成更正式约束
- `Radish` docs QA 在真实 batch 已能落 `manifest / audit / replay index / artifacts / recommended replay summary` 后，下一批真实 captured negative 应先补哪些优先违规类型

## 下一步优先推进

1. 将当前 `M3` 主线从继续跑真实 batch 切到服务/API 最小实现切片：优先把 `RadishFlow suggest_flowsheet_edits` 的 `CopilotRequest -> runtime -> CopilotResponse` 路径上提为 Gateway 入口和回归门禁。
2. 把 `suggest_flowsheet_edits`、`suggest_ghost_completion` 与 `Radish docs QA` 的现有治理资产继续维护为服务改动的验收基础，而不是默认继续扩样。
3. 继续沿 `RadishFlow export -> adapter -> request` 主线维护真实 exporter 契约；除非上游暴露新边界，否则不再优先补单个边界样本。
4. 维护 `Radish` 的 `answer_docs_question` 治理链；只有真实 captured negative 或跨 source drift 出现新增高价值假设时，再扩对应样本。
5. 沿 `radishmind-core-baseline-evaluation.md` 已固定的离线评测运行记录契约，后续只在明确执行离线模型评测时补真实观测结果和成本实测；当前不下载模型、不启动训练。
