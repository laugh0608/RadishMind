# RadishMind-Core 首版基座评估矩阵

更新时间：2026-05-07

## 文档目的

本文档用于固定 `M4` 前置阶段的模型基座评估口径。

当前目标不是下载模型、启动训练或承诺最终生产模型，而是先明确 `RadishMind-Core`、`minimind-v`、`Qwen2.5-VL` 与 `SmolVLM` 在首轮评估中的角色、准入门槛、退出条件和非目标。

## 当前判断

`RadishMind-Core` 是基座适配型自研主模型，不是从零预训练基础大模型。首版应优先比较 `3B` 与 `4B` 的协议遵循、中文任务理解、结构化输出、citation 对齐、风险确认和本地部署成本；只有当 `3B/4B` 在关键任务上明显不足，且部署资源与量化策略可接受时，才评估 `7B`。

图片输入理解可以进入 `RadishMind-Core` 或视觉适配路线；图片像素生成不进入主模型参数目标，继续由 `RadishMind-Image Adapter` 和独立生图 backend 承接。

当前 `task-scoped builder` 扩样前的复核口径已经单独落到 `training/datasets/radishmind-core-task-scoped-builder-review-plan-v0.json`。该文件只固定 planned review 维度、planned batch 和 acceptance policy，不声明任何真实 `reviewed_pass`，也不允许把 builder / repaired 轨误写成 raw 晋级证据。

2026-05-06 已完成 citation-tightened full-holdout-9 `Qwen2.5-1.5B-Instruct --build-task-scoped-response` 复跑：candidate summary 重新达到 `schema_valid_rate=1.0`、`task_valid_rate=1.0`、`builder_output_count=9`、`timeout_count=0`，offline eval 三组任务 blocking metrics 全部通过，自然语言 audit 为 `pass`、`violation_count=0`、`warning_count=3`、`fallback_natural_field_rate=0.142857`。正式 review records 已更新为 9/9 `reviewed_pass`；docs QA 三条短 action title warning、task-specific fallback text、risk/advisory boundary 与 holdout leakage 均已复核通过，`compressor-parameter-update` 的 broad citation blocker 也已关闭。当前这只恢复 builder/tooling 轨机器与人工复核通过状态，不代表 raw 晋级、训练准入或 production contract 接受证据。

同一轮人工复核已确认 `compressor-parameter-update` 不再使用 broad `artifact:flowsheet_document` citation，而是收口到 indexed diagnostics、unit config 与 latest snapshot evidence；deterministic scaffold 也已覆盖这条边界。这个修正只说明 fixture/scaffold 与人工复核口径已收紧，不追认旧本地产物为 raw 晋级证据。broader task-scoped builder review 的 15 样本 entry、两段 runbook 和 pending records 骨架已经固定；2026-05-07 已按 runbook 完成两段本地 broader review 并审计 `tmp/` 产物，但 records 仍保持 `pending_review`，不能提前写 `reviewed_pass`。

2026-05-06 已将 broader task-scoped builder review 的可执行样本面固定为 15 条：`full-holdout-9` 的 9 条 reviewed_pass 样本，加上 `holdout6-v2-non-overlap` 的 6 条非重叠回归样本，三类任务各 5 条。该入口落在 `training/experiments/radishmind-core-task-scoped-builder-broader-review-entry-v0.json`，并由 `scripts/check-radishmind-core-task-scoped-builder-broader-review-entry.py` 接入仓库级验证；它只收口下一轮 broader review 的执行面，不代表 raw 晋级、训练准入或 production contract 接受证据。

同日补充 broader review runbook：`training/experiments/radishmind-core-task-scoped-builder-broader-review-runbook-v0.json` 将 15 样本 surface 拆成 full-holdout-9 与 holdout6-v2-non-overlap 两段本地执行清单，并明确两段各自的 candidate summary、offline eval、natural-language audit 和后续 human review record 期望。`scripts/check-radishmind-core-task-scoped-builder-broader-review-runbook.py` 负责固定该 runbook 只描述计划，不生成结果，也不提前把 broader review 伪装成 reviewed_pass。

同日再补 pending review records 骨架：`training/datasets/radishmind-core-task-scoped-builder-broader-review-records-v0.json` 预置 broader 15 样本逐条复核入口和 `tmp/` artifact 路径，但状态保持 `pending_review`。`scripts/check-radishmind-core-task-scoped-builder-broader-review-records.py` 会阻止该记录集在真实两段本地执行、offline eval、natural-language audit 和人工复核完成前被标成 `reviewed_pass`。

2026-05-07 已按 runbook 完成两段 broader review 本地执行：full-holdout-9 与 holdout6-v2-non-overlap 的 candidate summary 均达到 `schema_valid_rate=1.0`、`task_valid_rate=1.0`、`timeout_count=0`，offline eval 与 natural-language audit 均为 pass；full-holdout-9 仅保留 3 条 docs QA short-title warning，v2 仅保留 1 条 warning，均无 violation。当前 broader review records 已把这两段 machine gate、offline eval 与 natural-language audit 的真实结论写实到批次级 summary，但由于样本级人工复核仍未完成，记录集继续保持 `pending_review`；这批产物只能作为 tooling-route evidence，不应写成 `reviewed_pass`、raw 晋级或训练准入证据。

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
- task-scoped builder 的后续扩样也遵循同一原则：机器通过、自然语言 audit 通过和 planned review 口径固定，都只能证明 tooling 路线可继续观察，不能直接替代 raw 晋级或训练准入判断
- 2026-05-06 citation-tightened full-holdout-9 复跑已完成：同一批 9 条样本重新达到 `9/9` schema/task valid、offline eval 全绿与自然语言 audit `violation_count=0`，并且 review records 现在已推进为 9/9 `reviewed_pass`。`compressor-parameter-update` 的 numeric detail 断言与 citation locator 断言都已通过；该结果只恢复 builder/tooling 轨机器与人工复核通过状态，仍需把它和 raw 晋级、训练准入严格分开。
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

2026-05-01 继续新增 `scripts/checks/fixtures/radishmind-core-holdout-probe-eval-manifest.json`、`scripts/checks/fixtures/radishmind-core-holdout-probe-candidate-manifest.json` 与 `scripts/checks/fixtures/radishmind-core-holdout-probe-candidate-eval-manifest.json`，从 `training/datasets/copilot-training-holdout-split-v0.json` 中各选 1 条训练外 holdout 样本，作为扩样前的轻量观测入口：

- `radishflow-suggest-flowsheet-edits-compressor-parameter-update-ordering-001`
- `radishflow-suggest-ghost-completion-mixer-standard-outlet-001`
- `radish-answer-docs-question-navigation-001`

该 holdout probe 的真实本地观测同样使用 `--sample-timeout-seconds 300`、`--allow-invalid-output`、`--validate-task`，并保持 raw / repaired 双轨同 timeout：

- raw holdout probe：`schema_valid_rate=0.6666666666666666`、`task_valid_rate=0.0`、`timeout_count=0`、`hit_max_new_tokens_count=0`、总生成耗时约 `174.198s`；`mixer-standard-outlet` 输出为 JSON parse error，`compressor-parameter-update` 与 `navigation` 虽 schema-valid 但任务级失败
- repaired holdout probe：`schema_valid_rate=0.6666666666666666`、`task_valid_rate=0.5`、`timeout_count=0`、`hit_max_new_tokens_count=0`、总生成耗时约 `173.570s`；`navigation` 经修复后通过任务级校验，但 `mixer-standard-outlet` 仍 schema-invalid，`compressor-parameter-update` 仍未通过参数 patch ordering 与多 issue 任务级要求
- 该结果说明上一轮 9 fixture repaired 全通过不能外推为 holdout 晋级结论；`--repair-hard-fields` 对 citation / hard-field / action scaffold 有帮助，但仍不足以治理 JSON 解析失败和更细的任务级结构要求

随后针对该 3 条 holdout 的两个集中失败面补了窄范围治理：`local_transformers` JSON 抽取会对对象 / 数组闭合前的尾逗号做一次结构安全清理，并在 generation metrics 中记录 `json_cleanup_applied`；`candidate_edit` scaffold 会从 `evaluation.ordered_parameter_update_*` 与 issue path 中生成有序 `parameter_updates` 与多 issue 结构。复跑同一 3 条 holdout 后：

- raw fix1：`schema_valid_rate=1.0`、`task_valid_rate=0.3333333333333333`，`mixer-standard-outlet` 通过 schema 与任务级校验，且记录 `json_cleanup_applied=true`；但 compressor 与 docs navigation 仍为任务级失败，raw 继续保持 `blocked`
- repaired fix1：`schema_valid_rate=1.0`、`task_valid_rate=1.0`，3 条样本均通过任务级校验；修复路径为 `$.answers`、`$.answers[0]`、`$.citations`、`$.issues`、`$.proposed_actions` 与 `$.status`
- 该结果只证明当前 3 条 holdout 的后处理与结构化抽取失败面已收敛，不改变 raw / repaired 分离口径；后续扩到完整 planned holdout 前仍不得把 repaired 结果视为 raw 能力晋级或训练样本准入依据

随后新增完整 planned holdout 的 dry-run 入口：`scripts/checks/fixtures/radishmind-core-full-holdout-eval-manifest.json`、`scripts/checks/fixtures/radishmind-core-full-holdout-candidate-manifest.json`、`scripts/checks/fixtures/radishmind-core-full-holdout-candidate-eval-manifest.json` 与 committed summary。该 manifest 镜像 `training/datasets/copilot-training-holdout-split-v0.json` 的 9 条样本，并接入 `check-repo` 的 `golden_fixture` 门禁。

使用本地 `Qwen2.5-1.5B-Instruct` 对完整 planned holdout 执行真实 `300s` raw / repaired 双轨观测后：

- raw full holdout：`schema_valid_rate=1.0`、`task_valid_rate=0.5555555555555556`，`timeout_count=0`、`hit_max_new_tokens_count=0`、总生成耗时约 `628.422s`；3 条 ghost completion 全部 task-valid，docs QA 为 `2/3`，suggest edits 为 `0/3`
- repaired full holdout：`schema_valid_rate=1.0`、`task_valid_rate=0.8888888888888888`，`timeout_count=0`、`hit_max_new_tokens_count=0`、总生成耗时约 `609.850s`；docs QA 与 ghost completion 均全通过，但 `suggest_flowsheet_edits` 仍有 `valve-local-fix-vs-global-balance` 的 `parameter_updates` 内层 detail key ordering 失败，offline eval 仍为 `blocked`
- 该结果确认 3 条 holdout fix1 repaired 全通过不能外推为完整 holdout 通过；下一步应窄化处理 `minimum_value`、`suggested_maximum` 等 parameter update detail-key ordering，同时继续把 repaired 视为后处理观测，不作为 raw 模型晋级或训练样本准入依据

针对上述剩余失败，随后补齐 `candidate_edit` scaffold 的 detail-key-only 参数更新路径：当样本声明 `evaluation.ordered_parameter_update_detail_keys` 但未声明 `ordered_parameter_update_keys` 时，scaffold 会按 detail-key metadata 推导 `parameter_updates` 外层参数顺序，并保留 `action -> minimum_value`、`action -> suggested_maximum` 等内层键顺序。复跑同一 9 条 full holdout 的 repaired 轨后：

- repaired full holdout fix1：`schema_valid_rate=1.0`、`task_valid_rate=1.0`，`timeout_count=0`、`hit_max_new_tokens_count=0`、总生成耗时约 `749.582s`；3 个任务组均通过 offline eval 指标
- 该结果只说明当前 scaffold / repair 能覆盖完整 planned holdout 的已知结构失败面；raw full holdout 仍为 `task_valid_rate=0.5555555555555556` 且保持 `blocked`，因此不得把 repaired fix1 视为 raw 模型晋级或训练样本准入依据

人工复核 `valve-local-fix-vs-global-balance` 后发现，旧 repaired fix1 虽通过机器指标，但 `parameter_updates.pressure_drop_kpa.minimum_value` 与 `parameter_updates.opening_percent.suggested_maximum` 被补成 `true`，不是可执行的数值阈值；同时 citation 只落到泛化的 `artifact:flowsheet_document`，没有精确指向 `diagnostics[0]`、`diagnostics[1]`、`flowsheet_document.units[0]` 与 `latest_snapshot`。因此旧 repaired fix1 应视为“机器指标通过、人工复核阻塞”，后续必须按新收紧的数值和 citation 断言重跑，不能沿用旧 `9/9` 作为通过结论。

随后补齐 indexed citation scaffold：当样本通过 `must_have_json_paths` 声明 `$.citations[index].id` / `locator` 等字段时，candidate scaffold 会优先按这些 index 断言和 golden citation 元数据生成精确 citation，`--repair-hard-fields` 也会在存在 indexed citation 断言时替换为该稳定顺序。按同一 9 条 full holdout、同一 `300s` timeout、`--allow-invalid-output`、`--validate-task` 和 `--repair-hard-fields` 复跑后：

- repaired full holdout fix3：`schema_valid_rate=1.0`、`task_valid_rate=1.0`，`timeout_count=0`、`hit_max_new_tokens_count=0`、总生成耗时约 `649.201s`；`valve-local-fix-vs-global-balance` 的 `minimum_value=10`、`suggested_maximum=85` 与 `diagnostics[0]` / `diagnostics[1]` / `flowsheet_document.units[0]` / `latest_snapshot` citation locator 均通过新门禁
- 该结果只说明当前 indexed scaffold / repair 能覆盖完整 planned holdout 的已知后处理失败面；raw full holdout 仍为 `task_valid_rate=0.5555555555555556` 且保持 `blocked`，不得把 repaired fix3 视为 raw 模型晋级或训练样本准入依据

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

.venv/bin/python scripts/run-radishmind-core-candidate.py \
  --manifest scripts/checks/fixtures/radishmind-core-holdout-probe-candidate-manifest.json \
  --provider local_transformers \
  --model-dir /home/luobo/Code/Models/Qwen2.5-1.5B-Instruct \
  --output-dir tmp/radishmind-core-holdout-probe-qwen15b-raw-timeout300 \
  --summary-output tmp/radishmind-core-holdout-probe-qwen15b-raw-timeout300/summary.json \
  --allow-invalid-output \
  --validate-task \
  --sample-timeout-seconds 300

.venv/bin/python scripts/run-radishmind-core-candidate.py \
  --manifest scripts/checks/fixtures/radishmind-core-holdout-probe-candidate-manifest.json \
  --provider local_transformers \
  --model-dir /home/luobo/Code/Models/Qwen2.5-1.5B-Instruct \
  --output-dir tmp/radishmind-core-holdout-probe-qwen15b-repaired-timeout300 \
  --summary-output tmp/radishmind-core-holdout-probe-qwen15b-repaired-timeout300/summary.json \
  --allow-invalid-output \
  --validate-task \
  --repair-hard-fields \
  --sample-timeout-seconds 300

.venv/bin/python scripts/run-radishmind-core-candidate.py \
  --manifest scripts/checks/fixtures/radishmind-core-full-holdout-candidate-manifest.json \
  --provider local_transformers \
  --model-dir /home/luobo/Code/Models/Qwen2.5-1.5B-Instruct \
  --output-dir tmp/radishmind-core-full-holdout-qwen15b-raw-timeout300 \
  --summary-output tmp/radishmind-core-full-holdout-qwen15b-raw-timeout300/summary.json \
  --allow-invalid-output \
  --validate-task \
  --sample-timeout-seconds 300

.venv/bin/python scripts/run-radishmind-core-candidate.py \
  --manifest scripts/checks/fixtures/radishmind-core-full-holdout-candidate-manifest.json \
  --provider local_transformers \
  --model-dir /home/luobo/Code/Models/Qwen2.5-1.5B-Instruct \
  --output-dir tmp/radishmind-core-full-holdout-qwen15b-repaired-timeout300 \
  --summary-output tmp/radishmind-core-full-holdout-qwen15b-repaired-timeout300/summary.json \
  --allow-invalid-output \
  --validate-task \
  --repair-hard-fields \
  --sample-timeout-seconds 300
```

上述命令的输出目录仍在 `tmp/` 下，只作为本地 artifact；提交时只记录 summary 摘要、指标和复跑命令，不提交候选响应本体、provider dump 或权重。

完成 full holdout fix3 后，下一步不再继续围绕同一 9 条做 repaired 收敛；新增 `radishmind-core-holdout-probe-v2-*` committed manifest，从 committed eval 中选择 6 条不与当前 9 条 planned full holdout 重叠的样本，用于观察 raw 在跨对象参数 patch、ghost ambiguity、docs source conflict 与 evidence gap 上的缺口。真实本地观测仍应使用 `300s`、raw / repaired 双轨、`--allow-invalid-output`、`--validate-task`，并只提交 summary / 实验记录 / 文档。

```bash
.venv/bin/python scripts/run-radishmind-core-candidate.py \
  --manifest scripts/checks/fixtures/radishmind-core-holdout-probe-v2-candidate-manifest.json \
  --provider local_transformers \
  --model-dir /home/luobo/Code/Models/Qwen2.5-1.5B-Instruct \
  --output-dir tmp/radishmind-core-holdout-probe-v2-qwen15b-raw-timeout300 \
  --summary-output tmp/radishmind-core-holdout-probe-v2-qwen15b-raw-timeout300/summary.json \
  --allow-invalid-output \
  --validate-task \
  --sample-timeout-seconds 300

.venv/bin/python scripts/run-radishmind-core-candidate.py \
  --manifest scripts/checks/fixtures/radishmind-core-holdout-probe-v2-candidate-manifest.json \
  --provider local_transformers \
  --model-dir /home/luobo/Code/Models/Qwen2.5-1.5B-Instruct \
  --output-dir tmp/radishmind-core-holdout-probe-v2-qwen15b-repaired-timeout300 \
  --summary-output tmp/radishmind-core-holdout-probe-v2-qwen15b-repaired-timeout300/summary.json \
  --allow-invalid-output \
  --validate-task \
  --repair-hard-fields \
  --sample-timeout-seconds 300
```

本地 `Qwen2.5-1.5B-Instruct` v2 probe 已完成。raw 结果为 `schema_valid_rate=1.0`、`task_valid_rate=0.3333333333333333`、`task_validation_attempted=6`、`timeout_count=0`、`hit_max_new_tokens_count=0`、总生成耗时约 `760.997s`；raw 通过的两条样本均为 `suggest_ghost_completion`，`suggest_flowsheet_edits` 与 `answer_docs_question` 仍 blocked。repaired 结果为 `schema_valid_rate=1.0`、`task_valid_rate=0.8333333333333334`、`task_validation_attempted=6`、`timeout_count=0`、`hit_max_new_tokens_count=0`、总生成耗时约 `692.28s`，但 repaired 仍 blocked：`radishflow-suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-plus-pump-parameter-001` 的第二个 action target、citation order、`parameter_updates` payload 与 patch ordering 未通过。

该结果说明 v2 非重叠样本没有复现 full holdout fix3 的 repaired 全通过：raw 继续不能晋级，repaired 也不能作为当前 v2 fixture pass 使用。下一步应优先复核跨对象 `suggest_flowsheet_edits` 参数 patch 家族，区分 scaffold 覆盖不足与模型 action planning 缺口；继续保持 raw / repaired 双轨和 `tmp/` artifact 禁入仓口径。

2026-05-02 已完成该 v2 阻塞样本的根因复核：失败主因是 candidate scaffold / repair 对多个 `candidate_edit` action 没有按 `action_index` 读取 target、citation、patch、`connection_placeholder` 与 `parameter_updates`，导致第二条 pump 参数 action 被修成第一条 stream reconnect action；同时还发现 ordered `connection_placeholder` 键未与已有 patch 合并、`minimum_reference_stream_id` 被填成布尔占位。当前已在 wrapper 与共享 scaffold helper 中收口这些缺口，并补仓库级检查锁住该 cross-object 样本的第二 action target、patch 顺序、`parameter_updates` payload 与 citation ids。基于旧 `tmp/` candidate response 重新执行 repair 和任务 validator 后已无 violation；这说明该 repaired blocker 属于 scaffold 覆盖不足，不应解读为该样本必须依赖更强模型 action planning。但 raw 仍保持 blocked，不能作为模型晋级或训练准入证据。

随后继续完成 v2 raw 失败复核。6 条非重叠样本中只有两条 `suggest_ghost_completion` raw 通过；两条 `suggest_flowsheet_edits` 和两条 `answer_docs_question` raw 失败需要分开处理：cross-object reconnect + pump 参数样本受旧 scaffold 影响，不能直接作为模型 action planning 失败证据；单 action efficiency range 样本主要是 `status=partial` 与必需 `answers` 没有保留；docs source-conflict 样本语义上遵循了 official docs，但漏掉 sample-required `read_only_check`，且当前通用 docs QA prompt 的“通常不生成 proposed_actions”与样本要求存在张力；evidence-gap 样本识别了证据不足，但把 `partial/medium` 降为 `ok/low`。因此下一步优先级应是先消除 prompt 中对 sample-level required action 的口径冲突，并评估 hard-field freeze / constrained decoding，而不是直接把 v2 raw 失败当作训练数据准入或扩大模型尺寸的证据。

同日已先收口上述 docs QA prompt 口径冲突：`required_action_kinds` 现在被视为必须输出的样本级规则，而不是“如果输出动作则优先使用”的弱偏好；当 `Radish docs QA` 样本声明 required action 时，通用“通常不生成 proposed_actions”文案会替换为必须按样本规则输出 action 的口径。新增 `scripts/check-radishmind-core-candidate-prompt-policy.py` 锁定 v2 docs source-conflict 样本必须提示 `read_only_check`，同时 evidence-gap no-action 样本仍保留显式 no-action 规则。该修正只证明 prompt policy 更一致，不代表 raw 模型指标已改善；如需刷新指标，应在同一 v2 manifest 上重新跑 `local_transformers` raw / repaired 双轨。

继续推进 hard-field freeze 的轻量实现：`run-radishmind-core-candidate.py` 现在会在 prompt document 中写入 `hard_field_freeze`，并在输出契约文本里列出由 scaffold 派生的不可改写 JSON path/value。该 freeze 合约只覆盖顶层 `status / risk_level / requires_confirmation`、显式 no-action 样本的空 `proposed_actions` 边界，以及样本通过 JSON path / required action 明确声明的稳定 action kind / target / patch / confirmation 字段；不冻结仅由 scaffold 占位生成的 answer kind、citation_ids、issue severity 等自然回答细节。新增 `scripts/eval/core_candidate_hard_field_freeze.py`、`scripts/check-radishmind-core-candidate-hard-field-freeze.py` 和 `scripts/audit-radishmind-core-candidate-freeze.py`，分别负责生成 prompt-time freeze、锁住 v2 evidence-gap / docs source-conflict / efficiency-range 三类 raw 漂移点，以及在本地模型复跑后审计 response 是否遵守 freeze。该策略不是 `--repair-hard-fields` 后处理；后续仍需用 raw / repaired 双轨重跑才能判断真实模型是否遵守 freeze。

先用用户本地执行的单样本 `radish-answer-docs-question-evidence-gap-001` 验证了该审计链路：`Qwen2.5-1.5B-Instruct` raw 输出耗时约 `78.957s`，schema-valid 但 task-invalid；freeze audit 显示 7 个冻结字段中 `$.status` 从 `partial` 漂移为 `ok`、`$.risk_level` 从 `medium` 漂移为 `low`。这说明 prompt-time freeze 已经可审计，但单靠当前 prompt 仍不足以让 1.5B raw 稳定遵守 evidence-gap 的风险边界；后续若继续推进 raw 能力，应优先验证 constrained decoding / 结构化字段冻结，而不是把该样本误判为检索失败。

随后用户用同一单样本执行 repaired 轨，`--repair-hard-fields` 后输出 schema-valid 且 task-valid，freeze audit 通过；summary 中 `repaired_paths` 只有 `$.status` 与 `$.risk_level`。这说明该样本 raw 的 answer、issue、citation 与 no-action 边界已经可用，阻塞点集中在两个硬字段；repaired 轨的收益边界也因此更清晰，仍只能作为显式后处理工程证据，不能替代 raw 模型服从能力。

继续用用户本地执行的单样本 `radish-answer-docs-question-docs-faq-forum-conflict-001` 验证 prompt policy 修正后效果：raw 输出耗时约 `73.326s`，schema-valid 但 task-invalid。与旧 v2 raw 结论不同，本次模型已经输出样本要求的 `read_only_check`，且 freeze audit 通过；剩余 violation 只剩缺少 `$.citations[1]` 与 `$.citations[2]`。因此该样本不应再归为 required-action preservation failure，而应更新为 multi-citation / source-context preservation failure：模型保留了官方 docs citation，但没有保留 FAQ 与 forum 两条上下文引用。

同一样本的 repaired 轨随后达到 schema/task 通过，`repaired_paths` 为 `$.citations` 与 `$.answers[0]`，freeze audit 通过。但人工复核发现 pre-fix repaired 输出只是把 `citations[1] / citations[2]` 补成重复的 docs citation，并没有恢复 golden 中的 FAQ / forum source-context。根因是 citation scaffold 在只看到 `$.citations[index]` 存在性要求时，会复制 primary artifact 生成 `artifact-2 / artifact-3`。当前已修正为优先按 `golden_response.citations` 的 index 恢复 citation，并新增 `check-radishmind-core-candidate-citation-scaffold.py` 锁住 docs / FAQ / forum 的 id 与 locator 顺序。修复后用户重跑同一单样本 repaired 轨，输出仍 schema/task 通过，freeze audit 通过，citations 已恢复为 `doc-1 / faq-1 / forum-1`，`repaired_paths` 收窄为 `$.answers[0]`。

随后继续用用户本地执行的单样本 `radishflow-suggest-flowsheet-edits-efficiency-range-ordering-001` 验证 v2 中的单 action 参数范围样本。raw 输出耗时约 `70.147s`，schema-valid 但 task-invalid；模型保留了 `candidate_edit` target、`parameter_updates.efficiency_percent.suggested_range=[65,82]`、`preserve_topology=true`、`risk_level=medium` 与确认边界，但把 frozen `$.status` 从 `partial` 漂移为 `ok`，并漏掉必需 answer。因此该样本应继续归为 hard-field/status 与 answer-shape preservation gap，而不是参数 patch planning 失败。

同一样本 repaired 轨随后 schema/task 通过，freeze audit 通过，`repaired_paths` 为 `$.status` 与 `$.answers`，且未改动 `candidate_edit` target 或 parameter update patch。但人工复核发现 pre-fix repaired answer 使用了通用占位句。当前已将 `suggest_flowsheet_edits` 的 answer scaffold 改成任务相关 `edit_rationale`，并新增 `check-radishmind-core-candidate-answer-scaffold.py` 锁住缺失 answer 修复时不得恢复通用占位文本。该修正提升 repaired 轨的人工可审计性，但不改变 raw 模型仍会漂移 `status` / 漏 answer 的结论。

随后按修复后的 action-index scaffold 与 hard-field freeze prompt 重新执行 v2 中最复杂的 `radishflow-suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-plus-pump-parameter-001` 单样本 raw / repaired 观测。两轨都使用本地 `Qwen2.5-1.5B-Instruct`、同一 `300s` timeout、`--allow-invalid-output` 与 `--validate-task`，repaired 轨额外启用 `--repair-hard-fields`。prompt 已包含 `5010` 个 input token 与 `17` 个冻结字段，覆盖顶层 `partial/high/requires_confirmation`、完整 citations、四个 issue code、两条 `candidate_edit` 的 target 与 patch。

- raw 单样本：`schema_valid=false`、`task_valid=false`、`generation_seconds=300.028`、`output_tokens=0`、`json_extracted=false`、`timeout_count=1`，invalid response 仅记录 `local_transformers sample timed out after 300s`
- repaired 单样本：`schema_valid=false`、`task_valid=false`、`generation_seconds=300.021`、`output_tokens=0`、`json_extracted=false`、`timeout_count=1`，`repaired_output_count=0`，因为没有生成 JSON，`--repair-hard-fields` 没有可修复对象
- freeze audit 两轨均为 `audited_count=0`、`skipped_count=1`，原因是 candidate response schema-invalid；因此本轮不能判断模型是否遵守冻结字段

该结果应记录为复杂 cross-object suggest edits 样本的成本 / timeout 边界，而不是 action planning 质量证据。它确认修复后的 scaffold 已进入 prompt，但当前 WSL CPU + `Qwen2.5-1.5B-Instruct` 在 `5010` input token、`300s` 档位下无法产出可审计 JSON。后续若要刷新该样本质量指标，应明确选择更长 timeout probe、缩短 prompt/scaffold，或换硬件 / backend；任何方案都必须继续保留 raw / repaired 双轨与 `tmp/` artifact 禁入仓口径。

随后为该成本边界补了静态 prompt budget 观测闭环。`run-radishmind-core-candidate.py` 会在每个 prompt document 中写入 `prompt_budget`，按字符拆出 system/user message、request JSON、task guidance、output contract、prompt scaffold、完整 response scaffold、hard-field freeze、freeze 字段数和 scaffold payload 体量；`scripts/check-radishmind-core-candidate-prompt-budget.py` 使用 v2 candidate manifest 的 `golden_fixture` dry-run 生成 prompt，不加载模型、不访问 provider，并把预算 summary 固定在 `scripts/checks/fixtures/radishmind-core-holdout-probe-v2-prompt-budget-summary.json`。

当前已把 prompt 内的完整 scaffold 切为 `compact_response_scaffold`：完整 response scaffold 仍由 wrapper/repair 逻辑在本地生成，但发送给模型的 scaffold 会把已由 `hard_field_freeze` 覆盖的大块对象替换成 `copy_from_hard_field_freeze` 引用。v2 prompt budget 显示，最大 prompt 仍是 `radishflow-suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-plus-pump-parameter-001`：

- `total_message_chars=14922`
- `request_json_chars=3632`
- `output_contract_chars=9610`
- `prompt_scaffold_json_chars=3227`
- `response_scaffold_json_chars=5950`
- `compact_scaffold_reduction_chars=2723`
- `compact_scaffold_copy_marker_count=5`
- `hard_field_freeze_json_chars=4351`
- `hard_field_freeze_field_count=17`
- `scaffold_counts` 为 `answers=1`、`issues=4`、`actions=2`、`citations=9`

这说明本轮 `300s` CPU timeout 的更直接工程解释是长 output contract + scaffold + freeze + 多 citation 组合造成的执行成本边界；在进一步验证 constrained decoding 或换 backend 之前，不应继续盲目要求用户复跑同一样本。新增预算门禁当前约束 `max_total_message_chars=15000`、`max_request_json_chars=4000`、`max_output_contract_chars=10000`、`max_prompt_scaffold_json_chars=3500`、`max_hard_field_freeze_field_count=20`，并要求最大样本的 compact scaffold 至少节省 `2000` 字符，用于防止后续 scaffold / freeze 为了追求 repaired 通过率继续无审计扩张。

## 当前执行停止线

截至本轮复盘，`Qwen2.5-0.5B-Instruct` 与 `Qwen2.5-1.5B-Instruct` 的本地观测已经足够支撑阶段性判断：prompt scaffold 能提高 schema 合法率，`--repair-hard-fields` 能验证后处理链路价值，但 raw 仍不能稳定保持硬字段、citation 和复杂 action shape；复杂 cross-object 样本还会先撞到 prompt/scaffold 成本边界。

因此后续不再把以下事项作为默认主线：

- 继续围绕同一 9 条 fixture 或同一 v2 cross-object 样本反复加长自然语言 prompt
- 用 repaired 全通过替代 raw 模型能力、训练样本准入或模型晋级判断
- 在没有新实验假设时要求用户延长 `local_transformers` timeout 或重跑同一样本
- 继续扩大 0.5B / 1.5B 细节补丁来证明本地小模型路线可行

后续 M4 实验必须先写清楚它要回答的路线决策问题，并满足至少一类条件：

- 验证 constrained decoding、JSON schema guided decoding 或硬字段外部注入是否能降低 raw 硬字段漂移
- 对比 `minimind-v`、`3B` 或 `4B` 候选，判断当前失败是否来自模型容量 / 架构，而不是 prompt 表达
- 补真实 reviewer 复核，把 repaired paths、citation 语义和训练 seed 质量从机器通过推进到人工可接受
- 判断某类复杂任务是否应继续交给规则 / 工具层、response builder 或后处理，而不是压给 `RadishMind-Core` raw 生成
- 形成可写入 `radishmind-core-offline-eval-run` 的真实观察记录、成本指标和晋级 / 暂缓结论

除非满足上述条件，新的本地模型复跑只作为人工调试，不应进入项目主线或周志的主要完成项。

## 今日主线

基于当前阶段判断，今天的主线已经从“继续补观测”收口为“先把决策型实验落成可复跑入口”。当前优先实验记录固定为 `training/experiments/radishmind-core-structured-output-decision-experiment-v0.json`，它明确回答以下问题：

- 当前 raw blocked，主因是否主要来自结构化输出约束方式不足，而不是单纯模型容量不足

该实验当前固定三类输入资产：

- 9 条 fixture baseline
- 9 条 planned full holdout
- 6 条 v2 非重叠 holdout probe

当前优先比较的不是更多 prompt 文本，而是以下变体：

- raw baseline
- prompt-time `hard_field_freeze`
- constrained / guided decoding
- 顶层硬字段外部注入
- `suggest_flowsheet_edits` response builder
- 可组合的 task-scoped response builder

只有当这些更强约束后，raw 仍主要卡在复杂 reasoning 或 action planning，下一步才应进入 `minimind-v`、`3B` 或 `4B` 对照。反过来，如果更强约束已经明显改善 `status / risk_level / requires_confirmation`、citation 顺序或 action boundary，则下一步优先推进 response builder / tool 分工，而不是先扩模型尺寸。

2026-05-04 首轮 v2 非重叠 holdout raw / repaired 双轨已经完成。两轨都使用本地 `Qwen2.5-1.5B-Instruct`、同一 6 条样本、同一 `300s` timeout、`--allow-invalid-output` 与 `--validate-task`；repaired 轨额外启用 `--repair-hard-fields`。随后 offline eval runner 只读取 `tmp/` candidate outputs 生成本地 run record，不重跑模型、不下载权重。

- raw：`schema_valid_rate=0.8333333333333334`、`task_valid_rate=0.4`、`timeout_count=0`，offline eval 决策仍为 `blocked`
- repaired：`schema_valid_rate=1.0`、`task_valid_rate=1.0`、`timeout_count=0`，但决策仍为 `no_promotion_planned`，因为它是显式后处理工程轨

该结果强化当前路线判断：同样模型、同样样本和同样 timeout 下，raw blocked 到 repaired `6/6` task-valid 的差异更支持先验证 constrained / guided decoding、硬字段外部注入和 response builder / tooling 分工，而不是直接把失败归因到模型容量并进入 `3B/4B` 对照。

当前已先落成硬字段外部注入的最小 candidate wrapper 变体：`--inject-hard-fields`。它只写回 prompt document 中 `hard_field_freeze.fields` 明确声明的 JSON path/value，并在 summary 中记录 `postprocess_policy.inject_hard_fields`、`injected_output_count` 与 `injected_path_counts`；它不会像 `--repair-hard-fields` 那样重建完整 response scaffold，因此可以作为 raw 与 repaired 之间的第三轨对照。下一步应优先让用户在同一 v2 非重叠 holdout 上运行该第三轨，再判断“只注入稳定硬字段”是否足以改善 raw 阻塞指标。

用户随后完成 `--inject-hard-fields` 第三轨真实本地运行和 offline eval。旧 wrapper 下第三轨仍为 `blocked`：candidate summary 为 `schema_valid_rate=0.8333333333333334`、`task_valid_rate=1.0`、`task_validation_attempted=5`、`timeout_count=0`，offline eval 中 `suggest_ghost_completion` 与 `answer_docs_question` 两组已全量通过，但 `suggest_flowsheet_edits` 仍因 `efficiency-range` 样本 schema-invalid 被阻塞。根因不是新增 action planning 失败，而是 `--inject-hard-fields` 只注入了 `$.issues[0].code`，未补齐 CopilotResponse schema 要求的 issue `message / severity`，导致注入后的对象不满足 schema。

当前已将第三轨 wrapper 修正为 schema-minimum completion：只对被 hard-field injection 触碰到的 `issues[*]` 与 `proposed_actions[*]` 补齐 schema 必需字段，仍不重建完整 `answers / issues / proposed_actions / citations` scaffold，也不会覆盖模型明确输出的合法布尔值。该修正已通过 `scripts/check-radishmind-core-candidate-hard-field-injection.py`、v2 dry-run candidate wrapper 与 offline eval 最小验证。下一步需要在同一 v2 非重叠 holdout 上重新运行 `--inject-hard-fields`，刷新第三轨真实指标；若修正后 `suggest_flowsheet_edits` 也通过，则路线信号将更明确地偏向 response builder / tooling track。

随后复核当前 `tmp/radishmind-core-holdout-probe-v2-qwen15b-injected-timeout300/` 与 `tmp/radishmind-core-holdout-probe-v2-qwen15b-injected-timeout300-run.json` 时确认，这组产物写入时间早于 `3e2b9bf fix(core): complete injected schema minimums`，仍是旧 wrapper 产物，不应视为修正后第三轨最终观测。该 stale 产物仍显示 `schema_valid_rate=0.8333333333333334`、`task_valid_rate=1.0`、offline eval `blocked`，阻塞点仍是 `efficiency-range` 样本缺少 `issues[0].message`。

同时据此补强 wrapper 边界：schema-minimum completion 现在按 explicit `hard_field_freeze` 覆盖的 issue/action path 判定，而不只依赖本轮实际发生值变化的 injected path。这样即使模型已经输出了正确的 frozen `issue.code`，但漏掉 schema 必需的 `message / severity`，`--inject-hard-fields` 也会补齐被 freeze 触达对象的最小合法字段。该补强不改变第三轨边界：仍不重建完整 answer、citation 或 action scaffold，也不把第三轨当作 raw 能力晋级证据。下一次真实本地复跑应覆盖这项补强后的 wrapper。

补强后用户重新执行同一 v2 非重叠 holdout 第三轨与 offline eval。两条命令均正常完成，但第三轨 overall 仍为 `blocked`：candidate summary 为 `schema_valid_rate=0.8333333333333334`、`task_valid_rate=0.8`、`task_validation_attempted=5`、`timeout_count=1`、`hit_max_new_tokens_count=0`、总生成耗时约 `873.147s`、平均 `145.525s`。`suggest_ghost_completion` 与 `answer_docs_question` 两组已通过所有 blocking offline-eval 指标；剩余阻塞集中在 `suggest_flowsheet_edits`。

本轮阻塞面已经不同于旧 wrapper schema-minimum bug：`cross-object-mixed-risk-reconnect-plus-pump-parameter` 样本在 `300s` 前未抽取到 JSON，injection 没有可约束对象；`efficiency-range` 样本经 injection 后已补齐 `issues[0].message / severity` 并 schema-valid，但仍因缺少必需 answer 与 citation 而 task-invalid。因此 `--inject-hard-fields` 被确认“有用但不足”：它能治理顶层硬字段、部分 action patch 和 issue schema minimum，不能单独恢复非冻结 answer/citation，也不能解决复杂 cross-object prompt 的本地 CPU timeout 成本边界。下一步应优先推进 constrained/guided decoding 或 response builder / tool 分工，而不是继续把 hard-field injection 单独扩成完整 scaffold repair。

随后新增 `--build-suggest-edits-response` 作为 response builder / tool 分工窄实验。该变体只作用于 `radishflow/suggest_flowsheet_edits`，由 builder 组装稳定的 `answers / issues / proposed_actions / citations / risk_level / requires_confirmation` 结构，模型输出只保留可展示的自然语言字段和 `confidence`；`suggest_ghost_completion` 与 `answer_docs_question` 仍走原输出。当前 deterministic smoke 未运行本地模型：`scripts/check-radishmind-core-suggest-edits-response-builder.py` 已覆盖 v2 中 `cross-object-mixed-risk-reconnect-plus-pump-parameter` 与 `efficiency-range` 两条阻塞样本；v2 `golden_fixture` smoke 的 candidate summary 为 `schema_valid_rate=1.0`、`task_valid_rate=1.0`、`builder_output_count=2`，offline eval 为 `no_promotion_planned`。该结果只证明 wrapper/builder 路径可用，下一步仍需要用户本地重跑同一 v2 非重叠 holdout 的 `--build-suggest-edits-response` 轨，观察真实 `timeout_count`、`builder_output_count` 与 offline eval 是否从 injected 第三轨的 `blocked` 转为通过。

用户随后完成真实本地 `--build-suggest-edits-response` v2 轨与 offline eval。该轨 candidate summary 为 `schema_valid_rate=1.0`、`task_valid_rate=0.6666666666666666`、`task_validation_attempted=6`、`builder_output_count=2`、`timeout_count=0`、总生成耗时约 `844.708s`、平均 `140.785s`。`suggest_flowsheet_edits` 两条样本均 schema-valid/task-valid，offline eval 中该任务的 schema、citation、risk confirmation、high-risk confirmation 与 advisory boundary 指标全部通过；旧 injected 轨中复杂 cross-object 样本的 `300s` timeout 未复现，实际约 `272.763s` 抽取到 JSON。整体 `promotion_status` 仍为 `blocked`，但阻塞面已经转移到 builder 未覆盖的两个 raw 任务：`suggest_ghost_completion` 的 ambiguous valve 样本缺少 `patch.ghost_kind` / `patch.candidate_ref`，`answer_docs_question` 的 evidence-gap 样本漂移了 expected `status` 与 `risk_level`。因此结论不是“builder 轨整体通过”，而是“`suggest_flowsheet_edits` 已验证适合模型意图 + builder 结构化 response 分工；下一步应做可组合的 task-scoped 结构化输出策略，补齐 ghost action patch 与 docs evidence-gap status/risk 边界”，且该结果仍不得视为 raw 模型晋级证据。

随后已落成 `--build-task-scoped-response` 组合实验轨。该变体对当前已有 task validator 的三类 eval task 统一使用 scaffold-derived response builder 组装 CopilotResponse 结构，模型输出只保留 `summary`、首条 answer text、issue message、action title/rationale 与 `confidence`。新增 `scripts/check-radishmind-core-task-scoped-response-builder.py` 锁定两个剩余 blocker 家族：ambiguous ghost completion 必须恢复有序 `patch.candidate_ref`、`patch.ghost_kind` 和 `manual_only` preview；docs evidence-gap 必须恢复 `status=partial`、`risk_level=medium`、空 `proposed_actions`、`INSUFFICIENT_EVIDENCE` issue 与 `doc-1` citation。v2 `golden_fixture` smoke 不运行本地模型，candidate summary 为 `schema_valid_rate=1.0`、`task_valid_rate=1.0`、`builder_output_count=6`，offline eval 为 `no_promotion_planned`。该 smoke 只证明组合 builder 链路可复跑；下一步需要在同一 v2 非重叠 holdout 上执行真实本地 `Qwen2.5-1.5B-Instruct --build-task-scoped-response` 轨，观察 overall 是否从 `blocked` 转为通过，以及自然语言字段是否仍可人工接受。

用户随后完成真实本地 `--build-task-scoped-response` v2 轨与 offline eval。该轨 candidate summary 为 `schema_valid_rate=1.0`、`task_valid_rate=1.0`、`task_validation_attempted=6`、`builder_output_count=6`、`timeout_count=0`、`hit_max_new_tokens_count=0`、`json_extracted_count=6`、总生成耗时约 `779.401s`、平均 `129.9s`；三组任务均为 `2/2` task-valid。offline eval 中 `suggest_flowsheet_edits`、`suggest_ghost_completion` 与 `answer_docs_question` 的所有 blocking metrics 均通过，整体 `promotion_status=no_promotion_planned`。这说明 v2 结构化阻塞已经可以由 task-scoped response builder / tooling 分工消除，下一步不应直接切 `3B/4B` 或把该结果当作 raw 晋级证据。抽查也发现机器指标未覆盖的自然语言风险：`docs-faq-forum-conflict` 的 answer text 仍是通用占位句，`ghost flash vapor` 把 `legal_candidate_completions` 误写成法律/法规语义。因此该轨是强路线信号，但 production contract 前仍需要补自然语言 merge/fallback 的人工复核和 guardrail。

当前已补首轮自然语言 merge/fallback guardrail：task-scoped builder 在合并模型自然语言字段前会拒绝通用占位句，以及 `suggest_ghost_completion` 中把 `legal_candidate_completions` 误译为法律/法规语义的文本；同时为 `answer_docs_question` 和 `suggest_ghost_completion` 生成任务感知 fallback，避免拒绝后回落到占位文本。`scripts/check-radishmind-core-task-scoped-response-builder.py` 已覆盖 docs source-conflict 占位 answer 与 ghost vapor 法律/法规误译两个真实风险，并保持 ambiguous ghost patch 与 docs evidence-gap status/risk 结构边界。v2 `golden_fixture` guardrail smoke 仍为 `schema_valid_rate=1.0`、`task_valid_rate=1.0`、`builder_output_count=6`，offline eval 仍为 `no_promotion_planned`。下一步需要用户在同一 v2 非重叠 holdout 上重跑真实本地 `--build-task-scoped-response` 轨，确认 guardrail 后的真实模型输出仍通过机器指标，并抽查 fallback 是否替代了坏自然语言。

用户随后重跑同一 v2 非重叠 holdout 的真实本地 `--build-task-scoped-response` 轨与 offline eval。机器指标仍通过：candidate summary 为 `schema_valid_rate=1.0`、`task_valid_rate=1.0`、`task_validation_attempted=6`、`builder_output_count=6`、`timeout_count=0`、`hit_max_new_tokens_count=0`、`json_extracted_count=6`、总生成耗时约 `779.401s`、平均 `129.9s`；offline eval 三组任务所有 blocking metrics 均为 `1.0`，整体仍是 `promotion_status=no_promotion_planned`。但人工复核目标未达成：`docs-faq-forum-conflict` 的 answer text 仍残留 `给出可展示给用户的回答。`，`ghost flash vapor` 仍残留 `法律候选完成体`、`合法 ghost completion` 与 `法规要求` 等误译。根因是首轮 guardrail 词表和回归样例过窄，且 ghost scaffold fallback 自身使用了“合法 ghost completion”措辞。当前已补强为真实失败面回归：拒绝 `法律候选完成体`、`符合法规要求`、`生成合法 ghost completion 候选` 等变体，并将 ghost scaffold fallback 改为不含法律/合法语义的候选补全描述；deterministic guardrail smoke 和 offline eval 仍通过。下一步仍需用户再跑修正后的真实本地轨，不能把本轮自然语言复核视为通过。

用户继续完成修正后同一 v2 非重叠 holdout 的真实本地复跑。该轨机器指标继续通过：`schema_valid_rate=1.0`、`task_valid_rate=1.0`、`task_validation_attempted=6`、`builder_output_count=6`、`timeout_count=0`、`hit_max_new_tokens_count=0`、`json_extracted_count=6`、总生成耗时约 `894.117s`、平均 `149.019s`；offline eval 三组任务所有 blocking metrics 仍为 `1.0`，整体 `promotion_status=no_promotion_planned`。目标 response 抽查也通过：`docs-faq-forum-conflict` 的 answer 已替换为 docs / FAQ / forum source precedence 的任务感知文本，不再出现通用占位句；`ghost flash vapor` 使用本地候选集与预览描述，不再出现 `法律`、`法规`、`合法` 或 `合规` 误译。至此，task-scoped builder 在 v2 非重叠 holdout 上同时具备机器通过与两类目标自然语言 guardrail 验证；该结论仍是 tooling 分工路线证据，不是 raw 模型晋级证据。下一步应在扩大样本面前先固定自然语言 review / audit 口径，覆盖语义正确性、引用解释质量和 fallback 使用比例。

当前已把该自然语言 review 口径落为仓库级 deterministic audit gate：新增 `scripts/audit-radishmind-core-task-scoped-natural-language.py` 与 `scripts/checks/fixtures/radishmind-core-task-scoped-natural-language-audit-summary.json`，并由 `scripts/check-repo.py` 重新生成 v2 `golden_fixture --build-task-scoped-response` 临时 candidate 输出后执行审计。该审计不运行模型、不下载权重，只接受 task-scoped builder summary；它会阻塞通用占位自然语言、ghost completion 的法律/法规/合法/合规误译，以及 docs source-conflict answer 中丢失 docs / FAQ / forum 来源语境的问题，同时记录 merged 与 fallback 自然语言字段比例。当前 fixture 审计结果为 `sample_count=6`、`violation_count=0`、`warning_count=0`、`natural_field_count=32`、`merged_natural_field_count=30`、`fallback_natural_field_count=2`、`fallback_natural_field_rate=0.0625`。这使下一步扩大 task-scoped builder 样本面前已有可复跑的自然语言质量门禁，但仍不能替代人工 reviewer 对引用解释质量、事实充分性和业务语义的判断。

当前又补 `scripts/check-radishmind-core-structured-output-run-set.py` 与 `scripts/checks/fixtures/radishmind-core-structured-output-run-set-summary.json`，把上述 v2 多轨真实观测从长实验记录抽成仓库级 run-set 门禁。该门禁只读取 `training/experiments/radishmind-core-structured-output-decision-experiment-v0.json`，不访问 `tmp/` 产物、不运行本地模型、不下载权重；它固定 raw / repaired / hard-field injection / suggest-edits builder / task-scoped builder / fixed guardrail 六条轨道的关键指标、`promotion_status`、`timeout_count`、`builder_output_count`、人工复核状态和 route signal。由此可以防止后续把 repaired 或 builder 通过误写成 raw 晋级，也防止遗漏“下一步应扩大 task-scoped builder 样本面与自然语言 review，而不是直接切 `3B/4B`”这个阶段结论。

当前已进一步补 `training/experiments/radishmind-core-task-scoped-builder-full-holdout-runbook-v0.json` 与 `scripts/check-radishmind-core-task-scoped-builder-full-holdout-runbook.py`，把 full-holdout-9 的 task-scoped builder 本地执行清单、`tmp/` 产物路径、offline eval、自然语言 audit 和 9 条单样本定位命令固定成仓库级 planned runbook。该 runbook 不运行模型、不声明结果；真实执行仍由开发者在本机终端完成，AI 后续只读取 `tmp/` 下的 summary、offline eval run、audit 和必要 candidate response 进行审计。

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

`training/datasets/radishmind-core-task-scoped-builder-review-plan-v0.json` 则单独固定 task-scoped builder 的扩样前复核维度、planned batch 和阻断规则；它只是一份 planned review 计划，不等于真实 reviewer 结论。

`training/experiments/radishmind-core-task-scoped-builder-full-holdout-runbook-v0.json` 则固定 full-holdout-9 的本地执行准备，确保下一轮真实 task-scoped builder 观测有统一命令、统一 `tmp/` 输出路径和统一审计顺序。

这些 smoke 只检查评估口径和边界是否稳定，不下载模型、不启动训练、不访问外部 provider。

## 暂不做

- 不下载 `minimind-v`、`Qwen2.5-VL`、`SmolVLM` 或任何权重
- 不启动微调、蒸馏、量化或长驻推理服务
- 不把 `14B/32B` 写成默认自研主模型目标
- 不把图片像素生成并入 `RadishMind-Core`
- 不为尚未接入的上层 UI / 命令层继续新增模拟 summary
