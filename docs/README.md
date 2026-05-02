# RadishMind 文档总览

更新时间：2026-05-02

## 文档目的

本目录用于沉淀 `RadishMind` 的产品定位、系统架构、阶段路线和跨项目集成边界，作为后续设计与实现的真相源入口。

当前版本已经基于对 `D:\Code\RadishFlow` 与 `D:\Code\Radish` 的只读审查完成了一轮收口；当前已冻结 `Python` 主实现栈，并将首轮模型路线收口为 `RadishMind-Core` 基座适配型自研主模型、`minimind-v` 默认 `student/base` 主线、`Qwen2.5-VL` 强基线和 `SmolVLM` 轻量对照组。
当前 `RadishFlow export -> adapter -> request` 主线也已从“只有扁平 adapter fixture”推进到“存在上游导出边界、bootstrap 模板、preflight smoke validator、batch smoke 与 committed exporter edge fixtures”的状态。
当前 `RadishFlow suggest_flowsheet_edits` 与 `suggest_ghost_completion` 都已具备真实批次、artifact summary、replay、recommended replay 与 real-derived negative 治理链；`suggest_flowsheet_edits v93` 与 `suggest_ghost_completion v25` 收口后，gateway、UI consumption 与 candidate handoff smoke 已冻结为未来上层接入验收门禁，不再继续扩展本仓库内模拟上层接线。
当前模型规模口径为：`RadishMind-Core` 首版优先 `3B` / `4B`，长期本地部署上限 `7B`；图片生成能力不并入主模型参数目标，默认通过 `RadishMind-Image Adapter` 和独立生图 backend 提供。`RadishMind-Core` 首版基座评估矩阵、离线评测阈值与离线评测运行记录已落成可回归门禁，用于固定 `minimind-v`、`3B/4B/7B`、`Qwen2.5-VL` 与 `SmolVLM` 的首轮评估边界、阻塞指标、样本选择和结果记录格式；当前离线评测 runner 已能同时读取 committed `golden_response` fixture 和 candidate wrapper 生成的 response 文件，固定 `candidate output -> offline eval` 的仓库级门禁。真实本地模型只通过 `--provider local_transformers` 加 `--model-dir` 或 `RADISHMIND_MODEL_DIR` 显式接入，脚本使用本地文件且不自动下载权重，并已补 per-sample timeout、JSON cleanup 指标、task validator、raw / `--repair-hard-fields` repaired 双轨 summary 与 pair summary。当前已经用 WSL CPU 环境下的本地 `Qwen2.5-0.5B-Instruct` 与 `Qwen2.5-1.5B-Instruct` 完成同批 9 条 M4 fixture、timeout probe、planned holdout、full holdout 和 v2 非重叠 holdout 观测；完整 planned holdout repaired fix3 可达到 `9/9` task-valid，但 raw 仍 blocked，v2 非重叠 holdout repaired 也仍 blocked，说明后处理结果不能外推为 raw 模型晋级、训练准入或更大样本面质量结论。当前还已新增从 committed eval 样本和 audit pass candidate record 生成 `CopilotTrainingSample` 的转换入口，先固定三条主任务各 3 条 golden_response 蒸馏样本与各 3 条 teacher_capture 样本，不运行模型、不下载权重；训练 JSONL 默认作为 `tmp/` 下的本地生成产物，`training/` 只提交 manifest、summary、复核策略和实验说明。更大训练集合的首个治理草案已落到 `training/datasets/copilot-training-dataset-governance-v0.json`，用于固定 candidate record 入选、分层抽样复核、离线评测 holdout、质量门禁与退场条件；当前还补了 `copilot-training-review-record-v0.json` 与 `copilot-training-holdout-split-v0.json`，把 planned 人工复核模板和三条主任务各 3 条的非重叠 holdout split 接入仓库级检查。`RadishMind-Image Adapter` 也已从 intent schema 扩展到 backend request 与 artifact metadata 两段契约，并补首个最小图片生成评测 manifest，只评估结构化意图、backend request 映射、artifact metadata、safety gate 与 provenance，不调用真实 backend、不生成图片。

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
11. [训练目录说明](../training/README.md)
12. [脚本目录说明](../scripts/README.md)

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
- 更大训练 / 蒸馏样本集已具备首个治理 manifest 草案、planned 复核记录模板和 holdout split；后续仍缺真实 reviewer 复核结果和离线评测结果
- `Qwen2.5-VL` 在当前任务上的具体首选尺寸、推理路由与成本上限
- `SmolVLM` 作为轻量对照组的具体准入任务和退场条件
- `RadishMind-Core` 首版基座评估矩阵、阻塞阈值、离线评测记录格式、fixture-run runner、candidate wrapper 与 candidate-output offline eval 门禁已固定；0.5B / 1.5B 本地 raw 对照、timeout 分层、planned/full/v2 holdout raw / repaired 观测已完成，后续仍需围绕跨对象 `suggest_flowsheet_edits` 参数 patch 家族和 repaired paths 做人工复核与更强结构化输出策略
- `RadishMind-Image Adapter` 已具备 backend request、artifact metadata 契约和首个最小图片生成评测 manifest；后续仍缺真实 backend 实现
- `RadishFlow` 截图/VLM 路线进入主线的触发条件
- `RadishFlow export` 在更多真实 exporter 接线后，是否需要继续把 `selection` 排序、focus 归一与 `support_artifacts` 摘要策略升级成更正式约束
- `Radish` docs QA 在真实 batch 已能落 `manifest / audit / replay index / artifacts / recommended replay summary` 后，下一批真实 captured negative 应先补哪些优先违规类型

## 下一步优先推进

1. 继续把现有 gateway、UI consumption 与 candidate handoff smoke 维护为未来上层接入验收门禁；在 `RadishFlow` / `Radish` 尚未准备真实模型或 Agent 接入前，不继续新增同类模拟 summary。
2. 沿 `run-radishmind-core-candidate.py` 与 `run-radishmind-core-offline-eval.py` 已接通的 candidate 输出评测入口，继续保持 raw / repaired 双轨、同 timeout、`--allow-invalid-output` 与 `--validate-task`；下一步优先处理 v2 cross-object 样本暴露的 prompt/scaffold 成本边界，先评估更短的 hard-field freeze 表达或 constrained decoding，再决定是否重跑本地 `local_transformers`，当前仍不启动训练。
3. 在已落成的训练集合治理 manifest、planned 复核记录和 holdout split 上，后续补真实 reviewer 复核结果与离线评测运行记录；JSONL 继续默认输出到 `tmp/`。
4. 继续沿 `image-generation-intent / backend-request / artifact` 三段契约和最小图片生成评测 manifest 推进 `RadishMind-Image Adapter` 的真实 backend 包装；当前不下载模型、不生成图片。
5. 继续维护 `suggest_flowsheet_edits`、`suggest_ghost_completion` 与 `Radish docs QA` 的现有治理资产；只有服务/API、模型评测或真实接入暴露新增非重复缺口时，再扩真实 capture。
