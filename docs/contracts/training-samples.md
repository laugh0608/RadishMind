# RadishMind 训练 / 蒸馏样本契约

更新时间：2026-05-12

## `CopilotTrainingSample` 训练 / 蒸馏样本契约

`RadishMind-Core` 的首版训练 / 蒸馏样本以 `CopilotRequest -> CopilotResponse` 为核心，不引入第二套任务协议。样本 wrapper 只负责记录训练模式、teacher/source、训练字段、质量门禁和来源 metadata。

当前已落成仓库级可回归契约：

- Schema：`contracts/copilot-training-sample.schema.json`
- 最小 fixture：`scripts/checks/fixtures/copilot-training-sample-basic.json`
- Smoke：`scripts/check-copilot-training-sample-contract.py`
- Eval 转换入口：`scripts/build-copilot-training-samples.py`
- 首批转换清单：`scripts/checks/fixtures/copilot-training-sample-conversion-manifest.json`
- 首批转换 summary：`scripts/checks/fixtures/copilot-training-sample-conversion-summary.json`
- 首批 candidate record 转换清单：`scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-manifest.json`
- 首批 candidate record 转换 summary：`scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-summary.json`
- 训练集合治理 manifest 草案：`training/datasets/copilot-training-dataset-governance-v0.json`
- 训练集合人工复核记录模板：`training/datasets/copilot-training-review-record-v0.json`
- 训练集合 offline eval holdout split：`training/datasets/copilot-training-holdout-split-v0.json`
- 训练集合治理 smoke：`scripts/check-copilot-training-dataset-governance.py`
- 离线评测 runner：`scripts/run-radishmind-core-offline-eval.py`
- 离线评测 fixture-run manifest：`scripts/checks/fixtures/radishmind-core-offline-eval-fixture-run-manifest.json`
- 离线评测 fixture-run 输出：`scripts/checks/fixtures/radishmind-core-offline-eval-golden-run.json`
- Candidate 输出离线评测 manifest：`scripts/checks/fixtures/radishmind-core-offline-eval-candidate-run-manifest.json`
- Candidate 输出离线评测 dry-run：`scripts/checks/fixtures/radishmind-core-offline-eval-candidate-dry-run.json`
- Candidate wrapper：`scripts/run-radishmind-core-candidate.py`
- Candidate dry-run manifest：`scripts/checks/fixtures/radishmind-core-candidate-dry-run-manifest.json`
- Candidate dry-run summary：`scripts/checks/fixtures/radishmind-core-candidate-dry-run-summary.json`
- Timeout probe manifest / summary：`scripts/checks/fixtures/radishmind-core-timeout-probe-*`
- Holdout probe manifest / summary：`scripts/checks/fixtures/radishmind-core-holdout-probe-*`
- Full holdout manifest / summary：`scripts/checks/fixtures/radishmind-core-full-holdout-*`
- Non-overlapping holdout probe v2 manifest / summary：`scripts/checks/fixtures/radishmind-core-holdout-probe-v2-*`

当前 eval 转换入口先只使用 committed eval 样本中的 `input_request` 与 `golden_response`，不调用 provider、不下载模型、不读取 candidate record。首批转换固定三条主任务各 3 条样本：

- `radishflow/suggest_flowsheet_edits`
- `radishflow/suggest_ghost_completion`
- `radish/answer_docs_question`

该入口可输出 JSONL 训练样本，并用 summary fixture 固定样本数、任务分布、来源 eval 样本、生成 sample id、训练字段、质量门禁计数和确认边界统计。

当前训练输出布局固定为：

- `training/` 只提交训练集合治理说明、manifest、summary、抽样复核记录和实验说明
- 默认 JSONL 输出到 `tmp/`，作为本地生成产物，不直接入仓
- `datasets/` 继续承载 eval 样本、candidate record 和示例对象，不作为生成后训练 JSONL 的默认落点
- 若后续确需提交小型 JSONL fixture，必须先在 `training/` 中说明样本数、用途、来源、复核状态和退场条件
- 大规模 JSONL、权重、checkpoint、adapter 二进制和 provider 临时 dump 不进入本仓库 committed 资产

当前更大训练集合的首个治理 manifest 已固定为 `training/datasets/copilot-training-dataset-governance-v0.json`，状态为 `draft`。它不替代转换 manifest，也不包含生成后的 JSONL；职责是约束哪些样本未来能进入更大的训练 / 蒸馏集合：

- 当前 seed set 继续来自 committed eval `golden_response` 与 audit pass `teacher_capture` 两类转换 manifest
- candidate record 必须被转换 manifest 显式列入，并通过 batch manifest、audit report、record schema、response schema、`project/task/sample_id` 一致性、风险确认和 citation 检查
- 当前小型 seed set 默认全量复核；后续更大 candidate record 池按 `project / task / source / provider / model / risk_level / requires_confirmation / coverage_tags / batch_id` 分层抽样
- 默认每个分层至少复核 `20%` 且不少于 `5` 条；新项目/任务、新 provider/model、高风险动作、`requires_confirmation=true`、新 action/patch 结构、新 retrieval source、schema/contract 版本变化和历史失败模式必须全量复核
- 至少保留 `10%` 离线评测 holdout，且每个任务至少 `3` 条；holdout 不得进入同一训练 JSONL split
- 样本若出现 audit 失效、schema 失效、人工复核拒绝、确认边界弱化、来源不可追踪、无理由重复、holdout 泄漏或模式过期，应从训练集合退场

当前 planned 人工复核与 holdout 接线已补两份轻量资产：

- `training/datasets/copilot-training-review-record-v0.json` 只记录复核模板与三组 planned review batch：`golden_response` seed set、`teacher_capture` seed set 和 offline eval holdout 泄漏检查。在没有真实 reviewer、timestamp、逐维度结果和泄漏判断前，不得标记为 `reviewed_pass`
- `training/datasets/copilot-training-holdout-split-v0.json` 固定三条主任务各 3 条 planned holdout，且显式排除当前两份训练转换 manifest 中已经列入的样本，避免训练 / 评测泄漏。该 split 不生成 JSONL、不运行模型

当前 candidate record 转换入口只允许显式列入转换 manifest 的记录进入训练样本，且必须同时满足：

- 所属 batch manifest 与 audit report 都存在且互相指向一致
- 选中的 `sample_file` 在 audit report 中为 `pass`
- candidate record 通过 `datasets/eval/candidate-response-record.schema.json`
- candidate record 内嵌 `response` 通过 `contracts/copilot-response.schema.json`
- record 的 `project / task / sample_id` 必须与 committed eval 样本一致

首批 candidate record 转换同样固定三条主任务各 3 条样本，`distillation.source=teacher_capture`，`teacher.model` 与 `teacher.record_id` 来自真实 record。该入口仍不重新调用 provider、不下载模型、不启动训练；它只把已审计通过的真实候选响应转换成可复验的训练 / 蒸馏样本。

当前 `scripts/run-radishmind-core-offline-eval.py` 是 M4 的首个可执行离线评测 runner。它支持 `response_source=golden_response` 的 fixture-run，也支持 `response_source=candidate_response_file`，从 `run-radishmind-core-candidate.py` 生成的 summary 与 response 目录读取候选输出，复用现有任务级 validator 计算阻塞指标，并生成符合 `contracts/radishmind-core-offline-eval-run.schema.json` 的 completed run record。离线评测阶段不重新调用模型、不访问 provider、不下载权重；schema-invalid raw 输出会进入指标失败统计，而不是被伪装成通过或阻断整批记录生成。该 runner 仍不代表真实模型晋级，职责是证明真实 `student/base` raw / repaired 输出可以接入同一条评测管线。

当前 `scripts/run-radishmind-core-candidate.py` 是真实模型 candidate 输出前的本地 wrapper。它复用同一份离线评测 fixture-run manifest，先把每条 `input_request` 包装成 prompt document，再把 provider 产出的候选 JSON 校验为 `CopilotResponse`，并把 prompt 与 candidate response 写到运行时指定的输出目录。仓库级检查只使用 `golden_fixture` provider，输出 summary 固定为 `scripts/checks/fixtures/radishmind-core-candidate-dry-run-summary.json`；这一步只验证 prompt 包装、响应文件布局、schema 校验、`project/task` 一致性和高风险 `requires_confirmation` 边界，不代表真实模型能力。调试真实小模型时可加 `--sample-id` 限定单条样本，并用 `--sample-timeout-seconds` 固定单样本生成边界；`--allow-invalid-output` 会将 schema-invalid、JSON parse error 或 timeout 原始输出写入 `tmp/.../invalid-responses/`，这类输出只能作为观测证据，不得进入离线评测通过记录或训练集合。本地 `local_transformers` summary 会记录 per-sample 输入 token、输出 token、JSON 抽取、JSON cleanup、是否触达 `max_new_tokens`、timeout 和推理耗时，便于比较不同模型容量的本地成本。

若需要验证“解码后结构化修复”能否治理小模型硬字段漂移，可显式加 `--repair-hard-fields`。该模式会在 schema/task 校验前，用 prompt scaffold 派生的硬字段修复 `status / risk_level / requires_confirmation / action kind / action shape / issue / citation / answer kind` 等协议边界，并在 summary 的 `postprocess_policy` 与每条输出的 `postprocess.repaired_paths` 中记录修复范围。该模式是实验性后处理，不得替代 raw 模型能力观测；同一模型应同时保留未开启该开关的 raw summary 作为真实能力基线。当前 `Qwen2.5-1.5B-Instruct` 的 full holdout repaired fix3 可在既有 9 条 planned holdout 上达到 `9/9` task-valid，但 v2 非重叠 holdout repaired 仍被复杂跨对象 `suggest_flowsheet_edits` 参数 patch 阻塞，因此任何训练准入、模型晋级或样本面结论都必须继续以 raw / repaired 双轨、同 timeout、人工复核和非重叠 holdout 结果共同判断。

若需要验证“硬字段外部注入”是否能作为比完整 repair 更窄的结构化输出约束，可显式加 `--inject-hard-fields`。该模式只会把 prompt document 中 `hard_field_freeze.fields` 已明确声明的 JSON path/value 写回 candidate response，不重建完整 answers / issues / proposed_actions / citations scaffold，也不会补未冻结的自然语言或结构字段。它是 raw 与 `--repair-hard-fields` 之间的独立实验变体，用于判断顶层硬字段、no-action 边界、必需 action target/patch 等稳定字段是否应由 response builder / tooling 承接；不得与 `--repair-hard-fields` 同时启用，也不得作为 raw 模型能力晋级证据。

若需要验证“模型只输出任务意图/解释，工具层组装结构化 CopilotResponse”这一路线，可显式加 `--build-task-scoped-response`。该模式对当前已有 task validator 的 `radishflow/suggest_flowsheet_edits`、`radishflow/suggest_ghost_completion` 与 `radish/answer_docs_question` 使用任务级 response builder 组装稳定结构，模型输出只保留 summary、answer/issue/action 自然语言字段与 confidence。它必须与 `--repair-hard-fields`、`--inject-hard-fields` 和 `--build-suggest-edits-response` 互斥。2026-05-04 的 v2 非重叠 holdout 真实本地轨已证明该模式能消除当前三类任务的结构化阻塞，并已补通用占位句、ghost 法律/法规/合法误译、docs source-conflict 来源语境缺失三类自然语言 audit gate；但它仍是 M4 决策实验和 tooling 分工证据，不是 raw 模型晋级、训练准入或 production contract。

2026-05-06 的 full-holdout-9 正式 human review records 进一步固定该模式的当前边界：tightened 重跑的机器门禁、deterministic natural-language audit 和 human review 通过，9 条样本全部 `reviewed_pass`，docs QA 三条 short action title warning、task-specific fallback text、risk/advisory boundary 与 holdout leakage 已复核接受；`radishflow-suggest-flowsheet-edits-compressor-parameter-update-ordering-001` 也已从 broad `artifact:flowsheet_document` 收口到 indexed diagnostics、unit config 与 latest snapshot citation locator。当前 source eval fixture 与 deterministic scaffold 已收紧到 indexed diagnostics、unit config 与 latest snapshot citation locator；这只证明契约门禁已收紧，不追认旧本地产物为 raw 晋级证据。2026-05-08 读取并审计两段新的 broader `tmp/` 产物后，broader 15 样本人工复核现已推进到 15/15 `reviewed_pass`：full-holdout-9 与 holdout6-v2-non-overlap 的 machine gate / natural-language audit 继续通过，docs QA short-title warning 与更广范围的 task-grounded deterministic fallback 也已被接受。但这批 broader 通过结果仍只代表 builder/tooling 路线证据，不能把 builder 结果当作 raw 晋级或训练准入证据。

真实本地模型接入必须显式运行：

```bash
python3 scripts/run-radishmind-core-candidate.py \
  --provider local_transformers \
  --model-dir /path/to/local/model \
  --output-dir tmp/radishmind-core-candidate-local \
  --allow-invalid-output \
  --validate-task
```

或先设置 `RADISHMIND_MODEL_DIR`。该 provider 只使用本地模型目录和 `local_files_only`，不会自动下载权重、安装依赖、启动训练或提交生成的 candidate response 文件。若未提供模型目录，脚本应明确失败，而不是回退到联网下载。当前 WSL CPU 环境下的 `Qwen2.5-0.5B-Instruct` 与 `Qwen2.5-1.5B-Instruct` 已完成同批 9 条 M4 fixture raw 对照；1.5B raw 仍会改写硬字段，但显式 `--repair-hard-fields` 后处理实验已能在同批 1.5B 输出上达到 `9/9` schema-valid 与 `9/9` task-valid。该结果证明结构化修复策略有价值，但仍需后续与 raw 能力、更多样本和更严格的离线评测记录分开治理。

第一版结构如下：

```json
{
  "schema_version": 1,
  "kind": "copilot_training_sample",
  "sample_id": "radishflow-suggest-flowsheet-edits-training-basic-001",
  "training_mode": "distillation",
  "project": "radishflow",
  "task": "suggest_flowsheet_edits",
  "input_request": {},
  "target_response": {},
  "distillation": {
    "source": "golden_response",
    "teacher": {
      "provider": "fixture",
      "model": "golden-response",
      "record_id": "radishflow-suggest-flowsheet-edits-training-basic-001"
    },
    "train_fields": [
      "summary",
      "answers",
      "issues",
      "proposed_actions",
      "citations",
      "risk_level",
      "requires_confirmation"
    ]
  },
  "quality_gates": {
    "schema_validated": true,
    "risk_reviewed": true,
    "citation_checked": true,
    "human_review_required": false
  },
  "metadata": {
    "source_eval_sample": "datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json",
    "created_for": "teacher-student-distillation",
    "notes": []
  }
}
```

字段边界：

- `input_request` 必须继续通过 `contracts/copilot-request.schema.json` 校验
- `target_response` 必须继续通过 `contracts/copilot-response.schema.json` 校验
- `project / task` 必须在 wrapper、`input_request` 与 `target_response` 三处一致
- 训练样本必须保持 advisory safety：`input_request.safety.mode=advisory`
- 当目标响应包含需要确认的输出、`high` 风险动作，或非 `manual_only` 的中风险执行边界动作时，`input_request.safety.requires_confirmation_for_actions` 必须保持 `true`
- `quality_gates` 表示该样本进入训练 / 蒸馏集合前已经通过 schema、风险与 citation 检查，不表示真实业务动作可自动执行
- `teacher_capture` 样本必须来自 audit pass 的 committed candidate record；audit failed、warning-only 未复核、schema-invalid 或 citation/risk 边界不稳定的 record 不得进入训练集合
- `high` 风险 `proposed_actions` 仍必须保留 `requires_confirmation=true`，不得因进入训练样本而弱化人工确认边界
- `suggest_ghost_completion` 的 `manual_only` 中风险候选属于可见候选排序边界，不等同于直接写回动作；只要响应本身不要求确认，且没有默认 `Tab` 或高风险动作，训练样本应保留原始 eval request 的确认口径
