# RadishMind 训练目录

更新时间：2026-05-08

## 目录目标

`training/` 用于承载 `RadishMind-Core` 的训练、蒸馏和离线评测准备材料。

当前阶段只提交轻量、可审计、可复跑的训练治理文件，不提交模型权重、大规模训练产物或默认生成的 JSONL 数据。

## 当前职责

- 记录训练 / 蒸馏样本集合的选择口径、复核策略和生成命令
- 放置小型 manifest、summary、schema 补充说明和实验记录
- 串联 `datasets/eval/`、`datasets/eval/candidate-records/` 与 `CopilotTrainingSample` 转换入口
- 为后续微调、蒸馏、量化和离线评测提供稳定入口说明

## 与其他目录的边界

- `datasets/` 继续承载 committed eval 样本、candidate record、示例对象和后续可复用的原始样本资产
- `contracts/copilot-training-sample.schema.json` 是训练样本结构真相源
- `scripts/build-copilot-training-samples.py` 是当前唯一正式训练样本 JSONL 转换入口
- `training/datasets/copilot-training-dataset-governance-v0.json` 是首个训练集合治理 manifest 草案，用于固定 candidate record 入选、抽样复核、质量门禁、holdout 和退场条件
- `training/datasets/copilot-training-review-record-v0.json` 是首个 planned 人工复核记录模板，用于后续记录 reviewer、逐维度结果和泄漏判断
- `training/datasets/copilot-training-holdout-split-v0.json` 是首个 planned offline eval holdout split，当前每条主任务各保留 3 条且不与现有训练 seed manifest 重叠
- `training/datasets/radishmind-core-task-scoped-builder-review-plan-v0.json` 固定 task-scoped builder 扩样前的 planned review 维度、批次和准入 / 阻断规则；它只做计划，不记任何真实 `reviewed_pass`
- `training/experiments/radishmind-core-qwen15b-offline-eval-v0.json` 是首个本地 `Qwen2.5-1.5B-Instruct` raw / repaired 双轨离线评测观察摘要；它只记录指标、修复路径和 `tmp/` artifact 位置，不提交候选输出本体
- `training/experiments/radishmind-core-structured-output-decision-experiment-v0.json` 是当前 `M4` 主线的决策型实验骨架，用于固定“结构化输出约束是否足以改变路线判断”这一问题的样本面、对照变体、退出条件和本地复跑命令
- `training/experiments/radishmind-core-constrained-guided-decoding-runbook-v0.json` 固定 broader builder 15/15 `reviewed_pass` 之后的下一轮优先路线：先在同一 `holdout6-v2-non-overlap`、同一 `300s` 边界上验证 constrained/guided decoding，再决定是否扩样或切 `3B/4B`
- `training/experiments/radishmind-core-task-scoped-builder-full-holdout-runbook-v0.json` 固定 full-holdout-9 task-scoped builder 的本地执行清单、`tmp/` 产物路径、offline eval 和自然语言 audit 顺序；它只描述计划，不声明结果
- `training/experiments/radishmind-core-task-scoped-builder-broader-review-entry-v0.json` 固定 broader task-scoped builder review 的 15 样本 surface、现有证据来源和验证入口；它不运行模型、不生成 JSONL、不声明新的 `reviewed_pass`
- `training/experiments/radishmind-core-task-scoped-builder-broader-review-runbook-v0.json` 固定 broader task-scoped builder review 的两段本地执行清单、`tmp/` 产物路径、offline eval、自然语言 audit 和后续 review record 口径；它只描述计划，不声明结果
- `training/datasets/radishmind-core-task-scoped-builder-broader-review-records-v0.json` 是 broader 15 样本 review records；当前已写实两段本地 machine gate、offline eval、natural-language audit 与 15 条人工复核结论，整批状态现为 15/15 `reviewed_pass`；该结果只作为 builder/tooling 路线证据，不代表 raw 晋级、训练准入或 production contract 接受
- `scripts/checks/fixtures/radishmind-core-holdout-probe-candidate-manifest.json` 是当前轻量 holdout 观测入口，从 planned holdout split 中各取 1 条主任务样本；真实本地运行继续使用 raw / repaired 双轨、同一 `300s` timeout、`--allow-invalid-output` 和 `--validate-task`
- `scripts/checks/fixtures/radishmind-core-full-holdout-candidate-manifest.json` 与 `scripts/checks/fixtures/radishmind-core-holdout-probe-v2-candidate-manifest.json` 分别固定完整 planned holdout 和 6 条非重叠 holdout probe；当前观测结论是 full holdout repaired fix3 与 2026-05-04 v2 repaired 轨都可作为后处理链路证据，但 raw 仍 blocked，因此训练准入不能只看 repaired pass
- `scripts/check-radishmind-core-task-scoped-builder-review-plan.py` 已接入 `check-repo`，用于固定该 review plan 只保持 planned 状态，不伪造 reviewer、timestamp 或 `reviewed_pass`
- `scripts/check-radishmind-core-task-scoped-builder-full-holdout-runbook.py` 已接入 `check-repo`，用于固定 full-holdout-9 runbook 的必需参数、样本覆盖、`tmp/` 产物边界和非 raw 晋级口径
- `scripts/check-radishmind-core-task-scoped-builder-broader-review-records.py` 已接入 `check-repo`，用于固定 broader review records 的 15 样本覆盖、`tmp/` 产物路径、批次级 `reviewed_pass` 结论、accepted warning / fallback 样本和非 raw / 非训练准入口径
- `scripts/check-radishmind-core-constrained-guided-decoding-runbook.py` 已接入 `check-repo`，用于固定 guided/constrained 下一轮 runbook 已完成 wrapper/provider 契约接线、保持 `holdout6-v2-non-overlap` + `300s` 的同边界对照，并且在本机 `transformers` runtime 真正具备 guided-decoding hook 前不要求用户执行本地模型命令
- `scripts/check-radishmind-core-guided-decoding-contract.py` 已接入 `check-repo`，用于固定 `--guided-decoding json_schema` CLI、互斥边界、summary policy 和 runtime-support failure boundary；该检查不运行本地模型
- `tmp/` 用于本地生成的临时 JSONL、探测输出和一次性中间产物，默认不提交
- 后续若需要提交小型 JSONL fixture，必须先写清楚样本数、用途、来源、复核状态和退场条件

## JSONL 输出约定

当前训练样本 JSONL 默认视为本地生成产物：

```bash
python3 scripts/build-copilot-training-samples.py \
  --manifest scripts/checks/fixtures/copilot-training-sample-conversion-manifest.json \
  --output-jsonl tmp/copilot-training-samples-golden.jsonl
```

```bash
python3 scripts/build-copilot-training-samples.py \
  --manifest scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-manifest.json \
  --output-jsonl tmp/copilot-training-samples-teacher-capture.jsonl
```

默认只提交转换 manifest 与 summary，不提交上述 `tmp/*.jsonl` 输出。

提交训练相关文件前，应确认：

- 训练样本仍通过 `contracts/copilot-training-sample.schema.json`
- 嵌套 `input_request` 仍通过 `contracts/copilot-request.schema.json`
- 嵌套 `target_response` 仍通过 `contracts/copilot-response.schema.json`
- `teacher_capture` 样本只来自 audit pass 的 committed candidate record
- 高风险动作和需要确认的输出没有弱化 `requires_confirmation`
- 更大 candidate record 池已按 `training/datasets/copilot-training-dataset-governance-v0.json` 完成分层抽样复核，并保留离线评测 holdout

## 训练集合治理

当前首个治理 manifest 是 `training/datasets/copilot-training-dataset-governance-v0.json`，状态为 `draft`。它不生成训练样本，也不启动训练，只固定 M4 前置阶段的集合治理规则：

- 当前 seed set 来自两类转换 manifest：committed eval `golden_response` 与 audit pass `teacher_capture`
- candidate record 必须显式列入转换 manifest，且通过 batch manifest、audit report、record schema、response schema、project/task/sample_id 一致性与风险/citation 检查
- 当前小型 seed set 默认全量复核；后续更大池按 `project / task / source / provider / model / risk_level / requires_confirmation / coverage_tags / batch_id` 分层抽样
- 默认每个分层至少复核 `20%` 且不少于 `5` 条；新任务、新 provider/model、高风险动作、确认边界样本、新 action/patch 结构、新来源族、schema 版本变化和历史失败模式必须全量复核
- 至少保留 `10%` 离线评测 holdout，且每个任务至少 `3` 条，避免训练 / 评测泄漏
- 样本一旦出现 audit 失效、schema 失效、人工复核拒绝、确认边界弱化、来源不可追踪、无理由重复或 holdout 泄漏，应从训练集合退场

`scripts/check-copilot-training-dataset-governance.py` 已接入 `check-repo`，用于检查该治理 manifest 的关键字段、来源 summary、抽样比例、artifact 禁入仓规则和离线评测接线。

当前还补了两份 planned 资产：

- `training/datasets/copilot-training-review-record-v0.json`：记录 seed set 全量复核、teacher_capture 全量复核和 holdout 泄漏检查的模板；在没有真实 reviewer、timestamp 与逐维度结果前，保持 `pending_review`
- `training/datasets/copilot-training-holdout-split-v0.json`：每个主任务保留 3 条 planned offline eval holdout，并显式排除当前训练 seed manifest 已列入样本

这两份资产同样不生成 JSONL、不下载模型、不启动训练。

## 后续可扩展结构

后续进入真实微调或蒸馏前，可按稳定职责补充浅层目录：

- `training/datasets/`: committed 小型训练集合索引、manifest、summary 和抽样复核记录
- `training/experiments/`: 小规模实验说明、参数摘要、评测结果和退场判断；实验记录应优先提交轻量 JSON/Markdown 摘要，不提交本地模型输出、provider dump 或权重
- `training/adapters/`: LoRA / adapter 配置摘要和可复跑命令说明

这些目录不用于提交权重、缓存、下载模型或大规模生成数据。

## 当前实验主线

当前训练/评测侧的主线不是继续扩 `RadishFlow` capture，也不是继续围绕同一批 0.5B / 1.5B 样本堆 prompt 文本，而是先回答一个路线决策问题：

- 当前 raw blocked，主因是否主要来自结构化输出约束方式不足，而不是单纯模型容量不足

为此，当前优先实验入口是 `training/experiments/radishmind-core-structured-output-decision-experiment-v0.json`。它要求：

- 优先复用现有 `candidate output -> offline eval` 入口
- 继续保留 raw / repaired 双轨、同 timeout、`tmp/` artifact 禁入仓
- 已比较 raw baseline、prompt-time hard-field freeze、`--inject-hard-fields` 硬字段外部注入、`--build-suggest-edits-response` 单任务 builder 与 `--build-task-scoped-response` 组合 builder 轨
- 2026-05-08 的阶段结论是：broader 15 样本 review 的 machine gate、offline eval、natural-language audit 和人工复核都已完成，records 现为 15/15 `reviewed_pass`；这能作为 builder/tooling 路线的更强人工复核证据，但仍不能当作 raw 晋级或训练准入证据
- 当前下一步不是继续重跑同一批 broader review，而是先确认本机 `local_transformers` runtime 是否支持 guided-decoding hook，再把 constrained/guided decoding 作为下一轮正式主线；只有在这条轨仍不能明显改善 raw 后，才决定是否扩大样本面或做 `minimind-v` / `3B/4B` 对照
- 当前 task-scoped builder 扩样前复核口径已单独落到 `training/datasets/radishmind-core-task-scoped-builder-review-plan-v0.json`，后续扩大样本面前必须先满足该 planned review 维度和阻断规则
- full-holdout-9 的执行准备已落到 `training/experiments/radishmind-core-task-scoped-builder-full-holdout-runbook-v0.json`，broader review 的 15 样本入口、两段执行清单和 review records 分别落到 `training/experiments/radishmind-core-task-scoped-builder-broader-review-entry-v0.json`、`training/experiments/radishmind-core-task-scoped-builder-broader-review-runbook-v0.json` 和 `training/datasets/radishmind-core-task-scoped-builder-broader-review-records-v0.json`；当前这批 broader review records 已完成从 `reviewed_changes_required` 到 15/15 `reviewed_pass` 的推进，下一步应基于这批通过结果稳定 builder/tooling 路线，而不是重新回到同一批 blocked 样本修复循环
