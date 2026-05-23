# `P4 Model Adaptation` v1 前置计划

更新时间：2026-05-23

## 任务目标

本任务卡用于把 `P4 Model Adaptation & Training` 从“可以开始规划”收口为可执行前置计划。

当前任务只定义 v1 模型能力目标、teacher/student 边界、样本分层、晋级门槛和训练 runbook 停止线；不启动训练放量，不下载模型权重，不生成训练 JSONL，不提交真实模型输出。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [阶段路线图](../radishmind-roadmap.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [产品范围](../radishmind-product-scope.md)
- [系统架构](../radishmind-architecture.md)
- [RadishMind-Core 基线评估](../radishmind-core-baseline-evaluation.md)
- [训练样本契约](../contracts/training-samples.md)
- `training/experiments/radishmind-core-model-adaptation-v1-preflight-runbook-v0.json`

## v1 能力目标

v1 模型适配优先覆盖已经有协议、评测和治理证据的任务面：

1. `radishflow/suggest_flowsheet_edits`
   - 目标：理解 flowsheet 诊断、选中对象、候选编辑和高风险确认边界。
   - 不要求：模型直接写 `FlowsheetDocument` 或绕过 deterministic validator。
2. `radishflow/suggest_ghost_completion`
   - 目标：解释合法候选排序、不可默认接受的歧义候选和 recent action 抑制边界。
   - 不要求：模型替代合法候选生成器。
3. `radish/answer_docs_question`
   - 目标：基于证据回答文档问题，保留 citation、证据不足和 read-only action 边界。
   - 不要求：模型替代权限判定、附件访问控制或治理动作。

首版不把 `RadishCatalyst`、真实 image backend、production tool execution、writeback 或 replay 纳入训练目标。

## Teacher / Student 边界

`teacher` 用于质量上界、标注参考、对照评测和蒸馏参考：

- 可以使用外部强模型或已审计通过的真实 candidate record。
- teacher 输出必须保留来源、provider/profile、record id、audit 状态和人工复核状态。
- teacher capture 不等于生产执行许可。

`student` 用于本地化、小成本部署和协议遵循能力验证：

- 优先评估本地可承受的小中型模型路线。
- `14B/32B` 不作为当前默认目标；长期本地部署上限仍按既有口径暂定 `7B`。
- student 晋级必须看 raw 输出，不能用 repaired、guided 或 builder 轨替代。

`RadishMind-Core` 是基座适配型自研主模型路线，不是从零预训练基础大模型。

## 样本分层

样本进入 P4 v1 计划前必须分层管理：

- `golden_response`：用于协议、validator 和离线评测稳定性检查。
- `teacher_capture`：来自 audit pass 且可追溯的真实候选响应，可作为蒸馏候选。
- `builder/tooling evidence`：用于证明 deterministic builder 或工具分工，不得写成 raw 模型能力。
- `holdout`：至少保留训练外评测集合，禁止进入同一训练 split。
- `negative / boundary`：覆盖 requires_confirmation、advisory-only、citation gap、歧义候选和 stop-line。

样本退场条件：

- audit failed 或人工复核 rejected。
- schema/task validator 失败且未被明确归类为实验观察。
- weakening `requires_confirmation`、business truth write、citation 失真或来源不可追踪。
- 与 holdout 泄漏、重复样本或过期协议绑定。

## 晋级门槛

任何 v1 student 晋级至少需要同时满足：

- raw 输出 schema valid 和 task valid 达到任务卡约定阈值。
- high-risk action 保留 `requires_confirmation=true`。
- citation、evidence gap 和 advisory-only 边界通过 deterministic audit。
- 人工复核接受关键样本，不以机器指标单独决定晋级。
- 同一结论在非重叠 holdout 上复验。
- 成本、延迟、上下文长度和失败类别有记录。

不得晋级的情况：

- 只靠 `--repair-hard-fields`、`--inject-hard-fields`、guided decoding 或 task-scoped builder 通过。
- summary 泄漏、硬字段漂移、citation 泛化或风险降级未清楚治理。
- 只在同一批 M4 样本上重复通过，没有非重复能力假设。

## 训练 Runbook 草案

P4 v1 预检 runbook 已落到 `training/experiments/radishmind-core-model-adaptation-v1-preflight-runbook-v0.json`，当前只作为人工后续本机执行清单，不代表训练启动许可。

它已固定：

- manifest 输入。
- provider / model-dir 参数。
- sample id 或 batch 范围。
- output-dir 与 summary-output。
- sample-timeout-seconds。
- 是否启用 `--repair-hard-fields`、`--inject-hard-fields`、`--build-task-scoped-response` 或 `--validate-task`。
- 产物审计、人工复核和文档回填步骤。

当前第一版只规划 `raw_student` 与 `repaired_student_comparison` 两条 full-holdout 预检轨；`repaired` 仅用于对照，不作为 raw 晋级证据。在数据治理复核完成前，不要求人工执行本地模型长跑、不下载权重、不生成训练 JSONL。

## 验收口径

本任务卡首轮验收只要求：

- P4 v1 目标、teacher/student 边界、样本分层和晋级门槛已写清。
- P4 v1 预检 runbook 已明确 manifest、provider/model-dir、output-dir、summary-output、timeout、raw/repaired 边界和 `tmp/` 产物政策。
- 停止线明确禁止训练放量、权重下载、真实大产物入仓和 builder/guided/repaired 冒充 raw 晋级。
- `docs/radishmind-current-focus.md` 与 `docs/radishmind-roadmap.md` 指向 P4 前置计划。
- `pwsh ./scripts/check-repo.ps1 -Fast` 通过。

## 下一步

1. 复核现有训练样本治理 manifest、review record 模板与 holdout split 是否足够支撑 v1 目标。
2. 补一份小型治理复核记录后，再判断是否要求人工执行 runbook 中的 raw student 单批脚本。
3. 人工执行完成后，读取 `tmp/` 下 summary、candidate response、prompt、audit 或 offline eval 产物，再更新结论。

## 停止线

- 不启动训练放量。
- 不下载模型权重。
- 不提交训练 JSONL、真实模型 raw output 或大体积实验产物。
- 不把 repaired、guided、injected 或 builder 轨写成 raw 晋级。
- 不把 P4 前置计划写成 production model ready。
