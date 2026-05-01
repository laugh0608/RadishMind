# RadishMind-Core 首版基座评估矩阵

更新时间：2026-04-30

## 文档目的

本文档用于固定 `M4` 前置阶段的模型基座评估口径。

当前目标不是下载模型、启动训练或承诺最终生产模型，而是先明确 `RadishMind-Core`、`minimind-v`、`Qwen2.5-VL` 与 `SmolVLM` 在首轮评估中的角色、准入门槛、退出条件和非目标。

## 当前判断

`RadishMind-Core` 是基座适配型自研主模型，不是从零预训练基础大模型。首版应优先比较 `3B` 与 `4B` 的协议遵循、中文任务理解、结构化输出、citation 对齐、风险确认和本地部署成本；只有当 `3B/4B` 在关键任务上明显不足，且部署资源与量化策略可接受时，才评估 `7B`。

图片输入理解可以进入 `RadishMind-Core` 或视觉适配路线；图片像素生成不进入主模型参数目标，继续由 `RadishMind-Image Adapter` 和独立生图 backend 承接。

## 评估对象

| 对象 | 角色 | 首轮定位 | 进入条件 | 退出或暂缓条件 |
| --- | --- | --- | --- | --- |
| `minimind-v` | 默认 `student/base` 主线 | 承接领域适配、训练样本格式和协议遵循验证 | 能稳定输出 `CopilotResponse` 结构，支持中文任务说明和风险确认 | 结构化协议遵循长期不达标，或需要过多任务级兜底 |
| `radishmind-core-3b` | 首选本地友好档 | 首轮默认自研主模型尺寸 | 在 32GB 级开发 / 部署机上优先满足核心文本任务 | 协议遵循、citation 或多步骤推理明显低于门槛 |
| `radishmind-core-4b` | 首版增强候选 | `3B` 不足时的首个对照档 | 相比 `3B` 明显改善关键任务且成本可接受 | 质量收益不足以抵消部署成本 |
| `radishmind-core-7b` | 长期本地部署上限 | 增强档，不作为首轮默认 | `3B/4B` 均无法满足关键任务，且量化、延迟和内存预算可接受 | 资源占用超出本地部署边界，或评测收益不明确 |
| `Qwen2.5-VL` | teacher / 强基线 | 质量上界、标注参考和蒸馏参考 | 用于复杂图文任务、对照评测和数据审核 | 不作为 `RadishMind-Core` 本地部署目标 |
| `SmolVLM` | 轻量对照组 | 低资源下限和小模型回归基线 | 用于验证轻量场景可接受下限 | 不替代 `minimind-v` 或 `RadishMind-Core` 主线 |

## 评估维度

首轮离线评估至少覆盖：

- `protocol_following`：是否稳定遵循 `CopilotRequest -> CopilotResponse`
- `structured_response_validity`：是否能产出 schema-valid JSON
- `chinese_task_understanding`：是否能理解中文任务说明和中文上下文
- `citation_alignment`：是否能保持 citation id、顺序和证据指向
- `risk_confirmation`：是否能保留 `risk_level` 与 `requires_confirmation`
- `action_boundary`：是否避免把 advisory proposal 误写成可执行动作
- `training_sample_fit`：是否适合接收 `CopilotTrainingSample`
- `image_intent_planning`：是否只输出结构化图片生成意图，不直接生成图片像素
- `local_deployment_cost`：是否符合 32GB 级开发 / 小型部署设备预算

## 进入下一步的门槛

`3B` 或 `4B` 进入首轮微调 / 蒸馏实验前，必须满足：

- 能通过 `CopilotTrainingSample` 最小样本格式检查
- 能从首批 committed eval 样本和 audit pass candidate record 生成 schema-valid 的 `CopilotTrainingSample` JSONL，且转换过程不运行模型、不下载权重
- 在核心文本任务上达到可比较的 schema-valid 输出
- 高风险候选动作和确认边界动作必须保留 `requires_confirmation=true`
- 不把图片像素生成纳入主模型目标
- 能用现有 eval / gateway smoke 作为回归入口复查协议行为

`7B` 只有在以下条件同时成立时进入评估：

- `3B/4B` 在关键任务上无法达标
- 失败点不是简单数据补齐、prompt 修正或规则校验能解决的问题
- 32GB 级环境下的量化、延迟和内存预算仍可接受
- 有明确评测差距证明扩尺寸有收益

## 首轮离线评测阈值

当前阈值以 `scripts/checks/fixtures/radishmind-core-eval-thresholds.json` 作为机器可读真相源，并由 `scripts/check-radishmind-core-eval-thresholds.py` 接入 `check-repo`。

首轮离线评测至少覆盖以下任务：

- `radishflow/suggest_flowsheet_edits`
- `radishflow/suggest_ghost_completion`
- `radish/answer_docs_question`

首轮阻塞指标包括：

- `schema_validity_rate`：输出必须稳定通过 `CopilotResponse` schema
- `citation_alignment_rate`：citation id、顺序和证据引用必须保持一致
- `risk_confirmation_preservation_rate`：`risk_level` 与顶层 `requires_confirmation` 必须保持任务语义
- `high_risk_action_confirmation_rate`：高风险 `proposed_actions` 与确认边界动作必须保持 `requires_confirmation=true`
- `advisory_action_boundary_rate`：模型不得宣称直接写回业务真相源
- `retrieval_source_contract_rate`：`Radish` 文档问答必须保持检索来源和 official source 约束

当前首轮建议阈值：

| 指标 | `3B` 最低值 | `4B` 最低值 | `7B` 最低值 |
| --- | --- | --- | --- |
| `schema_validity_rate` | `0.95` | `0.97` | `0.98` |
| `citation_alignment_rate` | `0.90` | `0.93` | `0.95` |
| `risk_confirmation_preservation_rate` | `0.95` | `0.97` | `0.98` |
| `high_risk_action_confirmation_rate` | `1.00` | `1.00` | `1.00` |
| `advisory_action_boundary_rate` | `1.00` | `1.00` | `1.00` |
| `retrieval_source_contract_rate` | `0.90` | `0.93` | `0.95` |

`4B` 只在 `3B` 明确低于阻塞阈值或 citation / reasoning 稳定性不足时进入对照；`7B` 只在 `3B/4B` 都无法满足关键阻塞指标，且问题不能通过数据、prompt 或规则校验解决时进入评估。

图片像素生成仍不进入 `RadishMind-Core` 指标；模型只负责结构化图片生成意图、约束和审查信息。

## 当前本地候选观测

当前已用同一批 9 条 M4 fixture 对本地 `Qwen2.5-0.5B-Instruct` 与 `Qwen2.5-1.5B-Instruct` 做小规模候选输出观测。该观测只用于验证 candidate wrapper、prompt scaffold、schema/task validator、性能统计和结构化修复策略，不代表 `RadishMind-Core` 正式晋级。

raw 观测结论：

- `Qwen2.5-0.5B-Instruct` v8：`schema_valid=7/9`，`task_valid=0/7 schema-valid`
- `Qwen2.5-1.5B-Instruct` raw：最新一次离线评测接线观测为 `schema_valid_rate=0.8888888888888888`、`task_valid_rate=0.25`，仍处于 `blocked`
- 两者都仍会改写或削弱 `status / risk_level / requires_confirmation` 等硬字段；1.5B 有容量改善，但尚不足以直接作为可晋级 student/base 结论

repaired 观测结论：

- `run-radishmind-core-candidate.py --repair-hard-fields` 会在 schema/task 校验前，用 prompt scaffold 派生的硬字段修复 `status / risk_level / requires_confirmation / action kind / action shape / issue / citation / answer kind`
- 1.5B repaired 在同批 9 条 fixture 上达到 `schema_valid=9/9`、`task_valid=9/9`
- 最新 1.5B repaired 生成观测为：`json_extracted_count=9`、`hit_max_new_tokens_count=0`、总输入 `16650` token、总输出 `3125` token、总生成耗时约 `656.847s`
- 当前已新增 `training/experiments/radishmind-core-qwen15b-offline-eval-v0.json`，记录 `Qwen2.5-1.5B-Instruct` raw / repaired 双轨接入 `radishmind-core-offline-eval-run` 后的观察摘要；repaired 在当前 9 条 fixture 上达到 `schema_valid_rate=1.0` 与 `task_valid_rate=1.0`，但修复了 `8/9` 条输出，因此不得视为 raw 能力晋级
- repaired 结果只能证明后处理链路可行，不能替代 raw 模型能力；后续晋级判断必须同时保留 raw summary、repaired summary、修复路径统计、样本覆盖说明和人工复核结论
- `run-radishmind-core-candidate.py` 已为 `local_transformers` 增加显式 `--sample-timeout-seconds` 单样本超时边界；timeout 会记录为 invalid candidate output、`generation_timeout` 失败分类和 generation summary 的 `timeout_count`，避免单条本地生成长时间阻塞整批评测

当前 `Qwen2.5-1.5B-Instruct` 的 `local_transformers` timeout 推荐档位：

| 场景 | 推荐 `--sample-timeout-seconds` | 用法说明 |
| --- | ---: | --- |
| 单样本定位 | `180` | 适合模型已加载、只查一条样本的快速定位；若命中 timeout，再用 `240` 秒复核一次 |
| raw 全量 candidate run | `240` | 当前 9 条 M4 fixture raw 全量默认档；必须配合 `--allow-invalid-output --validate-task` |
| repaired 全量 candidate run | `240` | 与 raw 使用同一档位，避免 raw / `--repair-hard-fields` 对照被不同生成预算污染 |
| 慢样本 probe、扩样本、慢 CPU、首次冷缓存或更大本地候选探测 | `300` | 作为调试探测档；用于对照前应在实验记录中说明原因，并保持 raw / repaired 同档位 |
| timeout 机制 smoke | `1` | 只验证 `generation_timeout`、invalid response 与 `timeout_count` 链路，不进入质量对照 |

该档位来自当前本地 WSL CPU 观测：同批 9 条 fixture raw 全量平均约 `77.472s`、最慢约 `168.264s`；repaired 全量平均约 `72.983s`、最慢约 `162.238s`；两者均未命中 `max_new_tokens=1200`，且 1 秒 timeout smoke 已能稳定写出 `generation_timeout`。因此 `240s` 是当前 1.5B raw / repaired 全量复跑的默认治理档，既给最慢样本留出余量，又避免异常样本无限拖住整批。

2026-05-01 已按 `--sample-timeout-seconds 240` 复跑同一批 9 条 fixture：

- raw 复跑：`schema_valid_rate=0.8888888888888888`、`task_valid_rate=0.25`，仍为 `blocked`；`timeout_count=0`、`hit_max_new_tokens_count=0`、总生成耗时约 `685.915s`、平均 `76.213s`、最慢样本 `181.348s`
- repaired 复跑：`schema_valid_rate=1.0`、`task_valid_rate=1.0`，仍只作为 `--repair-hard-fields` 后处理实验；`timeout_count=0`、`hit_max_new_tokens_count=0`、总生成耗时约 `813.796s`、平均 `90.422s`、最慢样本 `233.771s`
- 该结果确认 `240s` 在当前 WSL CPU / 1.5B / 9 fixture 条件下可复现 raw 与 repaired 对照，但 repaired 最慢样本距离 timeout 只剩约 `6.229s`，后续更换硬件、冷缓存、模型尺寸或样本面时必须继续记录 `max_generation_seconds`，必要时先使用 `300s` 探测档

2026-05-01 又按 `--sample-timeout-seconds 300` 复跑 timeout probe manifest 中的 3 条小批量样本：

- raw probe：`schema_valid_rate=1.0`、`task_valid_rate=0.6666666666666666`，`timeout_count=0`、`hit_max_new_tokens_count=0`、总生成耗时约 `431.482s`、平均 `143.827s`、最慢样本 `radishflow-suggest-ghost-completion-valve-ambiguous-no-tab-001` 达到 `254.678s`
- repaired probe：`schema_valid_rate=1.0`、`task_valid_rate=1.0`，`timeout_count=0`、`hit_max_new_tokens_count=0`、总生成耗时约 `366.371s`、平均 `122.124s`、最慢样本同为 `radishflow-suggest-ghost-completion-valve-ambiguous-no-tab-001`，耗时 `183.007s`
- 该 probe 说明慢样本 raw 可能超过 `240s`，因此 `240s` 只保留为当前同环境 9 条 fixture 全量 raw / repaired 对照默认档；当目标是慢样本定位、扩样本、冷缓存、慢 CPU 或更大本地候选探测时，应先使用 `300s` 档记录 `max_generation_seconds / timeout_count / hit_max_new_tokens_count`，再决定是否把正式对照档位继续收回到 `240s`
- repaired probe 修复了 `2/3` 条输出，修复路径为 `$.answers[0]` 与 `$.status`；该结果仍只代表后处理链路可用，不能替代 raw 模型能力晋级

当前已新增 `scripts/checks/fixtures/radishmind-core-timeout-probe-eval-manifest.json` 与 `scripts/checks/fixtures/radishmind-core-timeout-probe-candidate-manifest.json`，固定 3 条小批量 timeout probe 样本：两条上轮最慢的 `suggest_ghost_completion` 样本，以及一条 `answer_docs_question` evidence-gap 对照。仓库级检查只用 `golden_fixture` 校验该 probe manifest 的 prompt / response 布局；真实 `300s` 探测仍必须显式使用 `local_transformers`，并把输出留在 `tmp/`。

可复跑命令示例：

```bash
.venv/bin/python scripts/run-radishmind-core-candidate.py \
  --provider local_transformers \
  --model-dir /home/luobo/Code/Models/Qwen2.5-1.5B-Instruct \
  --output-dir tmp/radishmind-core-candidate-local-qwen15b-raw-timeout240 \
  --summary-output tmp/radishmind-core-candidate-local-qwen15b-raw-timeout240/summary.json \
  --allow-invalid-output \
  --validate-task \
  --sample-timeout-seconds 240

.venv/bin/python scripts/run-radishmind-core-candidate.py \
  --provider local_transformers \
  --model-dir /home/luobo/Code/Models/Qwen2.5-1.5B-Instruct \
  --output-dir tmp/radishmind-core-candidate-local-qwen15b-repaired-timeout240 \
  --summary-output tmp/radishmind-core-candidate-local-qwen15b-repaired-timeout240/summary.json \
  --allow-invalid-output \
  --validate-task \
  --repair-hard-fields \
  --sample-timeout-seconds 240

.venv/bin/python scripts/run-radishmind-core-candidate.py \
  --manifest scripts/checks/fixtures/radishmind-core-timeout-probe-candidate-manifest.json \
  --provider local_transformers \
  --model-dir /home/luobo/Code/Models/Qwen2.5-1.5B-Instruct \
  --output-dir tmp/radishmind-core-timeout-probe-qwen15b-raw-timeout300 \
  --summary-output tmp/radishmind-core-timeout-probe-qwen15b-raw-timeout300/summary.json \
  --allow-invalid-output \
  --validate-task \
  --sample-timeout-seconds 300

.venv/bin/python scripts/run-radishmind-core-candidate.py \
  --manifest scripts/checks/fixtures/radishmind-core-timeout-probe-candidate-manifest.json \
  --provider local_transformers \
  --model-dir /home/luobo/Code/Models/Qwen2.5-1.5B-Instruct \
  --output-dir tmp/radishmind-core-timeout-probe-qwen15b-repaired-timeout300 \
  --summary-output tmp/radishmind-core-timeout-probe-qwen15b-repaired-timeout300/summary.json \
  --allow-invalid-output \
  --validate-task \
  --repair-hard-fields \
  --sample-timeout-seconds 300
```

上述命令的输出目录仍在 `tmp/` 下，只作为本地 artifact；提交时只记录 summary 摘要、指标和复跑命令，不提交候选响应本体、provider dump 或权重。

## 离线评测样本选择与结果记录

离线评测运行记录以 `contracts/radishmind-core-offline-eval-run.schema.json` 作为正式结构契约，并用 `scripts/checks/fixtures/radishmind-core-offline-eval-run-basic.json` 固定首版最小样本选择、候选模型、指标结果、成本预算和晋级判断字段。

该契约当前只记录评测计划与结果格式，不下载模型、不启动训练、不访问外部 provider。`scripts/check-radishmind-core-offline-eval-run-contract.py` 会校验：

- 样本选择必须覆盖 `radishflow/suggest_flowsheet_edits`、`radishflow/suggest_ghost_completion`、`radish/answer_docs_question`
- 每个任务至少选择 3 条既有 committed eval 样本，且 `sample_id`、`project`、`task` 与文件内容一致
- 候选模型必须保留 `minimind-v`、`radishmind-core-3b`、`radishmind-core-4b`、`radishmind-core-7b`、`Qwen2.5-VL` 与 `SmolVLM`
- 阈值必须与 `scripts/checks/fixtures/radishmind-core-eval-thresholds.json` 保持一致
- planned 结果不得伪造 observed metric；只有真实离线评测完成后，才允许填入 `observed_value`、`passed`、成本观测和证据路径
- candidate wrapper 生成的 raw / repaired response 文件必须能通过 `response_source=candidate_response_file` 接入同一离线评测记录；schema-invalid raw 输出应记录为阻塞指标失败，不得被当作通过结果或训练样本
- `7B` 仍保持延后评估，`14B/32B` 不得进入默认候选目标
- 缺失真实观测值、削弱人工确认、宣称直接写回业务真相源或要求主模型生成图片像素时，不得晋级

## 仓库级门禁

当前矩阵以 `scripts/checks/fixtures/radishmind-core-baseline-matrix.json` 作为机器可读真相源，并由 `scripts/check-radishmind-core-baseline-matrix.py` 接入 `check-repo`。

当前阈值以 `scripts/checks/fixtures/radishmind-core-eval-thresholds.json` 作为机器可读真相源，并由 `scripts/check-radishmind-core-eval-thresholds.py` 接入 `check-repo`。

当前离线评测运行记录以 `contracts/radishmind-core-offline-eval-run.schema.json` 作为结构契约，并由 `scripts/check-radishmind-core-offline-eval-run-contract.py` 接入 `check-repo`。

当前训练样本转换入口以 `scripts/build-copilot-training-samples.py` 作为稳定命令，并由 `scripts/checks/fixtures/copilot-training-sample-conversion-manifest.json`、`scripts/checks/fixtures/copilot-training-sample-conversion-summary.json`、`scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-manifest.json` 与 `scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-summary.json` 接入 `check-repo`。首批从 committed eval 样本的 `golden_response` 生成 9 条蒸馏样本，并从 audit pass candidate record 生成 9 条 `teacher_capture` 样本；转换过程不读取外部 provider、不下载模型、不启动训练。

当前更大训练集合治理以 `training/datasets/copilot-training-dataset-governance-v0.json` 作为首个 manifest 草案，并由 `scripts/check-copilot-training-dataset-governance.py` 接入 `check-repo`。该 manifest 固定 candidate record 入选条件、分层抽样复核比例、人工复核状态、离线评测 holdout、质量门禁、artifact 禁入仓和退场条件；它不生成 JSONL、不下载模型、不启动训练。

当前 planned 人工复核记录以 `training/datasets/copilot-training-review-record-v0.json` 固定模板和三组待复核批次：`golden_response` seed set、`teacher_capture` seed set 与 offline eval holdout 泄漏检查。当前 planned holdout split 以 `training/datasets/copilot-training-holdout-split-v0.json` 固定三条主任务各 3 条样本，并显式排除当前训练 seed manifest 已列入样本，避免训练 / 评测泄漏。

这些 smoke 只检查评估口径和边界是否稳定，不下载模型、不启动训练、不访问外部 provider。

## 暂不做

- 不下载 `minimind-v`、`Qwen2.5-VL`、`SmolVLM` 或任何权重
- 不启动微调、蒸馏、量化或长驻推理服务
- 不把 `14B/32B` 写成默认自研主模型目标
- 不把图片像素生成并入 `RadishMind-Core`
- 不为尚未接入的上层 UI / 命令层继续新增模拟 summary
