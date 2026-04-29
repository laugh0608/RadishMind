# RadishMind 训练目录

更新时间：2026-04-29

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

## 后续可扩展结构

后续进入真实微调或蒸馏前，可按稳定职责补充浅层目录：

- `training/datasets/`: committed 小型训练集合索引、manifest、summary 和抽样复核记录
- `training/experiments/`: 小规模实验说明、参数摘要、评测结果和退场判断
- `training/adapters/`: LoRA / adapter 配置摘要和可复跑命令说明

这些目录不用于提交权重、缓存、下载模型或大规模生成数据。
