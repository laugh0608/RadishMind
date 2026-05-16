# RadishMind 最小评测样本说明

更新时间：2026-05-16

当前目录用于存放第一阶段的最小离线评测样本。入口文档只保留样本目录职责、核心 schema、常用验证入口和专题链接；任务覆盖、真实 batch、replay 与长命令细节拆到专题页。

## 目录职责

第一阶段先做两件事：

1. 冻结样本结构
2. 让样本能与任务卡、契约文件一一对应

当前样本至少应描述：

- 输入请求是什么
- 召回输入应满足哪些边界与元数据约束
- 期望输出应满足哪些结构约束
- 一份可作为对照基线的 `golden_response`
- 可选的 `candidate_response`，用于接入真实候选输出或模拟模型输出
- 可选的 `candidate_response_record` 引用，用于从外部记录文件回灌真实候选输出
- 风险等级应如何判定
- 哪些字段必须出现
- 哪些越界字段或行为不得出现

## 核心 schema

当前先使用以下 schema 约束样本格式：

- `radishflow-task-sample.schema.json`
- `radish-task-sample.schema.json`
- `candidate-response-dump.schema.json`
- `candidate-response-record.schema.json`
- `candidate-record-batch.schema.json`
- `recommended-negative-replay-summary.schema.json`

当前 `Radish` 文档问答样本已新增 `retrieval_expectations`，用于把任务卡中的召回输入边界落成可执行检查。

## 常用入口

日常仓库验证优先走根目录快速检查：

```powershell
pwsh ./scripts/check-repo.ps1 -Fast
```

Linux / WSL 环境使用：

```bash
./scripts/check-repo.sh --fast
```

如果只想查看三条真实 batch 治理链的状态，可执行：

```bash
python3 ./scripts/eval/report_real_batch_governance_status.py
```

该报告会统一给出 `suggest_flowsheet_edits`、`suggest_ghost_completion` 与 `Radish docs QA` 的 formal real batch、latest audit、coverage、replay / real-derived 接线状态。

## 专题页

- [评测 runner 与任务覆盖说明](eval-runner-and-task-coverage.md)
- [Radish docs QA 真实 batch 与 replay 治理说明](radish-docs-qa-replay-governance.md)

## 当前停止线

- 不把真实模型输出、训练 JSONL 或大体积实验产物直接提交入仓。
- 新增真实候选输出应先进入 candidate record / manifest / audit 链路，再决定是否回灌样本。
- 长批次流水、完整命令输出和历史失败细节继续沉淀到专题页、manifest、summary 或 run record，不回填到本入口文档。

后续再补更多真实候选输出记录、最小对照指标与更完整的回灌流程。
