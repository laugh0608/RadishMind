# training/datasets 说明

更新时间：2026-04-29

本目录预留给训练 / 蒸馏数据集的轻量治理文件。

当前首个治理草案为：

- `copilot-training-dataset-governance-v0.json`

该 manifest 只描述训练集合治理口径，不包含生成后的训练 JSONL。它固定以下内容：

- 训练集合当前覆盖 `radishflow/suggest_flowsheet_edits`、`radishflow/suggest_ghost_completion` 与 `radish/answer_docs_question`
- 训练样本结构继续以 `contracts/copilot-training-sample.schema.json` 为真相源
- 转换入口继续使用 `scripts/build-copilot-training-samples.py`
- 当前 seed set 来自 committed eval `golden_response` 与 audit pass `teacher_capture`，各 9 条
- 大规模 JSONL、权重、checkpoint、adapter 二进制与 provider dump 不入仓

当前只允许提交：

- 小型训练集合 manifest
- 转换 summary
- 抽样复核记录
- 样本来源索引
- 与离线评测接线有关的说明文件

默认不提交：

- 大规模 JSONL
- 模型权重
- tokenizer、checkpoint、adapter 二进制产物
- provider 原始输出临时 dump
- 本地探测缓存

首批 JSONL 应继续由 `scripts/build-copilot-training-samples.py` 从 committed eval 样本或 audit pass candidate record 生成，并默认输出到 `tmp/`。

## Candidate Record 入选条件

更大 candidate record 池进入训练 / 蒸馏集合前，必须同时满足：

- 被训练转换 manifest 显式列入
- batch manifest 与 audit report 均存在，且 audit report 指向同一个 batch manifest
- 选中样本在 audit report 中为 `pass`
- candidate record 通过 `datasets/eval/candidate-response-record.schema.json`
- 内嵌 `response` 通过 `contracts/copilot-response.schema.json`
- record 的 `project / task / sample_id` 与 committed eval 样本一致
- `risk_level`、`requires_confirmation`、citation 和 advisory-only 边界没有被弱化
- teacher provider、model 与 record id 可追踪

以下记录不得进入训练集合：

- audit failed、audit 缺失或未被 manifest 显式列入的记录
- schema-invalid、citation/source drift 未复核或风险边界不稳定的记录
- 高风险动作缺少 `requires_confirmation=true` 的记录
- provider 原始 dump、本地临时输出或未 committed batch 中的记录
- 人工复核状态为 `rejected`、`deprecated` 或仍需修正的记录

## 抽样与人工复核

当前 seed set 规模小，默认 `100%` 复核。

后续扩大 candidate record 池时，采用分层抽样复核：

- 分层维度至少包含 `project`、`task`、`source`、`provider`、`model`、`risk_level`、`requires_confirmation`、`coverage_tags` 与 `batch_id`
- 默认每个分层至少抽 `20%`，向上取整
- 每个分层至少复核 `5` 条；不足 `5` 条时全量复核
- `new_project_or_task`、新 provider/model、高风险动作、`requires_confirmation=true`、新 action/patch 结构、新检索来源族、schema/contract 版本变化和历史失败模式必须全量复核
- 至少保留 `10%` 离线评测 holdout，且每个任务至少 `3` 条 holdout；同一训练 JSONL split 不得重复包含这些 holdout 样本

人工复核状态先按以下枚举记录：

- `not_required_seed_fixture`
- `pending_review`
- `sampled_for_review`
- `reviewed_pass`
- `reviewed_changes_required`
- `rejected`
- `deprecated`

## 质量门禁与退场条件

训练 JSONL 生成前，必须确认：

- 转换 manifest 已复核
- 来源 eval 样本和 candidate record 已 committed
- summary 可由转换入口重生成且一致
- `CopilotTrainingSample`、嵌套 `CopilotRequest` 与嵌套 `CopilotResponse` 均通过 schema 校验
- `risk_reviewed`、`citation_checked` 与人工复核要求已解决

训练运行前，还必须确认：

- offline eval holdout 已声明，且没有训练 / 评测泄漏
- 没有 committed 大规模 JSONL、权重、checkpoint、adapter 二进制或 provider dump
- 训练命令只记录可复跑口径，不在当前阶段下载模型或启动训练

样本满足以下任一条件时应退场：

- 来源 batch audit 从 `pass` 变为 failed
- schema 或契约变化导致样本不再合法
- 人工复核标记为 `rejected` 或 `deprecated`
- 样本弱化 advisory-only 或 `requires_confirmation` 边界
- citation、retrieval source 或 record provenance 无法追踪
- 与既有样本重复且没有明确覆盖理由
- 被发现同时进入训练 split 与离线评测 holdout
- 后续治理报表确认该模式已过期或对当前任务契约有害
