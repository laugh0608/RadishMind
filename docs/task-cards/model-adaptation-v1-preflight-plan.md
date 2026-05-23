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
- `training/datasets/radishmind-core-model-adaptation-v1-governance-review-v0.json`
- `training/experiments/radishmind-core-model-adaptation-v1-preflight-result-v0.json`

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

## 样本治理复核

P4 v1 首轮治理复核已落到 `training/datasets/radishmind-core-model-adaptation-v1-governance-review-v0.json`。

当前结论：

- `golden_response`、`teacher_capture` 与 `holdout` 均覆盖三条 v1 任务，每条任务各有 3 条 seed / teacher / holdout 样本。
- holdout split 仍声明不与当前训练 seed manifest 重叠，并显式排除两份训练转换 manifest。
- `copilot-training-review-record-v0.json` 的 planned review batch 仍保持 `pending_review`；本次复核不伪造 reviewer、timestamp、逐样本维度结果或 `reviewed_pass`。
- 在选择本地 `model-dir` 且确认人工可接受本机 CPU/GPU 负载后，可以先请求一次 `raw_student_probe`；`repair_boundary_probe` 必须排在 raw 之后，只能作为后处理对照。

## 本地预检结果

2026-05-23 已完成 `Qwen2.5-1.5B-Instruct` full-holdout-9 raw / repaired comparison，摘要落到 `training/experiments/radishmind-core-model-adaptation-v1-preflight-result-v0.json`。

结论：

- raw student 完成 9 条样本，`timeout_count=0`、`hit_max_new_tokens_count=0`、`json_extracted_count=9`。
- raw student 总体 `schema_valid_rate=0.6667`，`radishflow/suggest_flowsheet_edits` 3/3 因 `citations` scaffold 引用泄漏 blocked；ghost completion 与 docs QA 6/6 schema/task valid。
- repaired comparison 使用 `--repair-hard-fields` 后 9/9 schema/task valid，修复集中在 3 条 edits 的 `$.citations`，并额外修复部分 `$.status`、`$.answers`、`$.issues`。
- repaired comparison 只能证明后处理有价值，不代表 raw 模型晋级、训练准入或 production model ready。
- 后续 `Qwen2.5-3B-Instruct` raw CPU 单样本 probe 选取同一已知 edits blocker，300 秒内未产出 token 并 timeout；该结果不可评价 3B 输出质量，但足以阻止在同一 CPU 路径上直接启动 3B full-holdout-9。

## 验收口径

本任务卡首轮验收只要求：

- P4 v1 目标、teacher/student 边界、样本分层和晋级门槛已写清。
- P4 v1 预检 runbook 已明确 manifest、provider/model-dir、output-dir、summary-output、timeout、raw/repaired 边界和 `tmp/` 产物政策。
- P4 v1 样本治理复核记录已明确 seed / teacher / holdout 覆盖、holdout 泄漏边界、人工复核未完成状态和 raw-first 执行顺序。
- P4 v1 预检结果摘要已明确 raw blocked、repaired comparison pass 以及不晋级边界。
- 停止线明确禁止训练放量、权重下载、真实大产物入仓和 builder/guided/repaired 冒充 raw 晋级。
- `scripts/check-radishmind-core-model-adaptation-v1-governance-review.py` 已接入 fast baseline，固定该复核记录不能漂成训练启动许可。
- `scripts/check-radishmind-core-model-adaptation-v1-preflight-result.py` 已接入 fast baseline，固定 raw blocked 与 repaired-only 口径。
- `docs/radishmind-current-focus.md` 与 `docs/radishmind-roadmap.md` 指向 P4 前置计划。
- `pwsh ./scripts/check-repo.ps1 -Fast` 通过。

## 下一步

1. 不在同一 CPU 路径上直接运行 `Qwen2.5-3B-Instruct` full-holdout-9。
2. 若 GPU 可用，或明确接受更长 timeout 成本，则先重跑同一已知 edits blocker 的 3B raw 单样本 probe；仍不启用 `--repair-hard-fields`、guided、injected 或 builder。
3. 若 3B raw 单样本仍 timeout 或仍出现 scaffold 泄漏，则停止继续容量探测，转向 raw 输出 scaffold 泄漏治理策略。
4. 只有新的 raw 结果和人工复核同时支持时，才讨论训练样本或蒸馏路线；当前不生成训练 JSONL。

## 停止线

- 不启动训练放量。
- 不下载模型权重。
- 不提交训练 JSONL、真实模型 raw output 或大体积实验产物。
- 不把 repaired、guided、injected 或 builder 轨写成 raw 晋级。
- 不把 P4 前置计划写成 production model ready。
