# RadishMind 最小评测样本说明

更新时间：2026-04-05

当前目录用于存放第一阶段的最小离线评测样本。

第一阶段先做两件事：

1. 冻结样本结构
2. 让样本能与任务卡、契约文件一一对应

当前样本至少应描述：

- 输入请求是什么
- 召回输入应满足哪些边界与元数据约束
- 期望输出应满足哪些结构约束
- 一份可作为对照基线的 `golden_response`
- 可选的 `candidate_response`，用于接入真实候选输出或模拟模型输出
- 或可选的 `candidate_response_record` 引用，用于从外部记录文件回灌真实候选输出
- 风险等级应如何判定
- 哪些字段必须出现
- 哪些越界字段或行为不得出现

当前先使用以下 schema 约束样本格式：

- `radishflow-task-sample.schema.json`
- `radish-task-sample.schema.json`
- `candidate-response-dump.schema.json`
- `candidate-response-record.schema.json`
- `candidate-record-batch.schema.json`
- `recommended-negative-replay-summary.schema.json`

当前 `Radish` 文档问答样本已新增 `retrieval_expectations`，用于把任务卡中的召回输入边界落成可执行检查。

当前已补最小回归脚本：

- `scripts/run-radishflow-control-plane-regression.ps1`
- `scripts/run-radishflow-control-plane-regression.sh`
- `scripts/check-radishflow-control-plane-eval.ps1`
- `scripts/check-radishflow-control-plane-eval.sh`
- `scripts/run-radishflow-diagnostics-regression.ps1`
- `scripts/run-radishflow-diagnostics-regression.sh`
- `scripts/check-radishflow-diagnostics-eval.ps1`
- `scripts/check-radishflow-diagnostics-eval.sh`
- `scripts/run-radishflow-ghost-completion-regression.ps1`
- `scripts/run-radishflow-ghost-completion-regression.sh`
- `scripts/check-radishflow-ghost-completion-eval.ps1`
- `scripts/check-radishflow-ghost-completion-eval.sh`
- `scripts/run-radishflow-suggest-edits-regression.ps1`
- `scripts/run-radishflow-suggest-edits-regression.sh`
- `scripts/check-radishflow-suggest-edits-eval.ps1`
- `scripts/check-radishflow-suggest-edits-eval.sh`
- `scripts/run-radish-docs-qa-regression.ps1`
- `scripts/run-radish-docs-qa-regression.sh`
- `scripts/run-radish-docs-qa-negative-regression.ps1`
- `scripts/run-radish-docs-qa-negative-regression.sh`
- `scripts/run-radish-docs-qa-negative-recommended.ps1`
- `scripts/run-radish-docs-qa-negative-recommended.sh`
- `scripts/run-radish-docs-qa-real-batch.ps1`
- `scripts/run-radish-docs-qa-real-batch.sh`
- `scripts/check-radish-docs-qa-eval.ps1`
- `scripts/check-radish-docs-qa-eval.sh`
- `scripts/import-candidate-response-dump.py`
- `scripts/build-candidate-record-batch.py`

关系说明：

- `run-radishflow-control-plane-regression.*` 负责执行 `RadishFlow explain_control_plane_state` 样本回归
- `check-radishflow-control-plane-eval.*` 负责把控制面状态说明回归接入仓库基线
- `run-radishflow-diagnostics-regression.*` 负责执行 `RadishFlow explain_diagnostics` 样本回归
- `check-radishflow-diagnostics-eval.*` 负责把该回归接入仓库基线
- `run-radishflow-ghost-completion-regression.*` 负责执行 `RadishFlow suggest_ghost_completion` 样本回归
- `check-radishflow-ghost-completion-eval.*` 负责把 ghost 补全回归接入仓库基线
- `run-radishflow-suggest-edits-regression.*` 负责执行 `RadishFlow suggest_flowsheet_edits` 样本回归
- `check-radishflow-suggest-edits-eval.*` 负责把候选编辑回归接入仓库基线
- `run-radish-docs-qa-regression.*` 是真正执行样本回归的 runner
- `run-radish-docs-qa-negative-regression.*` 是 `Radish` docs QA 的负例回放 runner，用来验证候选回答会被现有规则稳定拦下
- `run-radish-docs-qa-negative-recommended.*` 是 `Radish` docs QA 的推荐负例批量回放入口，用来直接执行 `artifacts.json` 中默认推荐的前 N 个失败组，并产出结构化回放摘要
- `run-radish-docs-qa-real-batch.*` 是 `Radish` docs QA 的真实/模拟 batch 编排入口，用来串起批跑、审计、replay 治理与可选的推荐回放摘要生成
- `check-radish-docs-qa-eval.*` 是仓库基线入口，对 runner 做包装
- `check-repo.*` 继续通过上述入口脚本把各任务回归纳入仓库级校验链路
- 当前 `ps1` / `sh` runner 都通过 `scripts/run-eval-regression.py` 共享同一份 Python 回归核心
- `import-candidate-response-dump.py` 用于把未来 adapter/mock/模型接口产出的 raw dump 裁剪成正式 `candidate_response_record`
- `build-candidate-record-batch.py` 用于从一批 `candidate_response_record` 文件生成 manifest，减少 captured batch 扩样时的手工清单维护
- `run-copilot-inference.py` 当前除单条推理外，也已支持按样本目录批量落 `response / dump / record / manifest`
- `audit-candidate-record-batch.py` 用于把一批真实 `candidate_response_record` 临时注入现有样本，再复用当前回归规则做批量审计
- `run-radish-docs-qa-real-batch.py` 当前在生成 `artifacts.json` 后，还可按推荐失败组顺序继续落一份 `recommended-negative-replay-summary.json`，把批量回放结果沉淀为可审计产物
- 因此执行这些回归脚本时，当前环境需要具备可用的 Python 启动器与 `jsonschema`

`RadishFlow` 的回归 runner 当前已覆盖 `explain_control_plane_state`、`explain_diagnostics`、`suggest_flowsheet_edits` 与 `suggest_ghost_completion` 四个任务，并支持样本内可选 `candidate_response` 校验，用于为后续真实模型输出接入预留稳定输入口。

其中 `explain_diagnostics` 当前已覆盖：

- 单对象规格缺失解释
- 单元未收敛与候选根因区分
- 全局诊断解释
- 多对象关联解释
- 链式诊断传播解释
- 证据不足下的相态/根因降级说明

同时该任务的回归当前会额外约束：

- `hypothesis_labeling` 样本必须包含 `ROOT_CAUSE_UNCONFIRMED`
- `cause_explanation` 必须显式使用不确定性表述
- `ROOT_CAUSE_UNCONFIRMED` 必须保持 `warning` 且消息中明确“未确认/证据不足/候选”口径

其中 `explain_control_plane_state` 当前已覆盖：

- entitlement 过期阻塞
- package sync 轻度异常
- 控制面冲突态
- 上游 403 授权边界对抗样本
- manifest / lease 版本错位组合态
- public / private package source 权限范围差异

同时该任务的回归当前会额外约束：

- `hypothesis_labeling` 样本必须通过 `cause_hypothesis` 或 `conflict_explanation` 显式标注不确定性
- `read_only_check` 必须保持 `low` 风险且不要求确认
- `candidate_operation` 必须要求确认，且不能伪装成自动修复

其中 `suggest_flowsheet_edits` 当前已覆盖：

- 流股缺失规格占位
- 高风险拓扑重连占位
- 选中流股与单元并存时的多动作局部提案
- 单元参数组合修正

同时该任务的回归当前会额外约束：

- `candidate_edit.target` 必须落在当前选择集或诊断目标内
- `patch` 必须保持可审查的局部结构
- `patch` 不得退化成命令式执行字段或整图重写字段

其中 `suggest_ghost_completion` 当前已覆盖：

- `FlashDrum inlet` 的标准入口补全
- `FlashDrum vapor_outlet` / `liquid_outlet` 的标准 ghost connection
- `FlashDrum` 基于 `nearby_nodes` 的多候选排序
- `Heater` 的 `ghost_stream_name` 命名补全
- `Mixer` 标准出口补全
- `Valve` 存在歧义时的“可见 ghost 但不默认 Tab 接受”
- 上下文不足时返回空建议

同时该任务的回归当前会额外约束：

- 候选必须来自 `context.legal_candidate_completions`
- 候选可显式携带 `ranking_signals`、`naming_signals` 与 `conflict_flags`，用于校验排序与命名证据是否充分
- `ghost_completion` 必须保持 pending 语义，不能升级成正式 patch
- 默认 `Tab` 接受键只能绑定到第一条 ghost 建议
- 多候选接近时可以返回建议，但不得强行把第一条伪装成默认 `Tab`
- 如果候选带有 `conflict_flags`，或本地没有把它标为 `is_tab_default=true` 且 `is_high_confidence=true`，回归会拒绝把它渲染成默认 `Tab` 建议
- 响应与 action 的 `requires_confirmation` 必须保持为 `false`

`Radish` 的 docs QA runner 当前已支持两种候选回答输入方式：

- 样本内可选 `candidate_response`
- 外部 `candidate_response_record.path`

当前还支持把一批外部记录先收口进 `candidate-record-batch` manifest，再由样本通过 `candidate_response_record.manifest_path + record_id` 解析具体快照。

对外部记录，当前最小回归会额外校验：

- `sample_id`、`request_id`、`project`、`task` 与样本请求对齐
- `input_record.current_app`、`route`、`resource_slug`、`search_scope`、`artifact_names` 与样本最小输入对齐
- `response` 仍必须通过统一 `CopilotResponse` schema 与任务级校验
- 若样本在 `candidate_response_record` 下声明 `expected_source`、`required_capture_origin`、`required_collection_batch` 或 `required_tags`，外部记录还需满足这些最小来源元数据约束

若通过 manifest 导入，当前回归还会补充校验：

- manifest 的 `project` / `task` 必须与样本一致
- `record_id` 必须能在 manifest 中唯一命中
- 记录文件中的 `record_id` / `sample_id` 必须与 manifest entry 对齐
- 记录文件中的 `capture_metadata.collection_batch` / `capture_origin` 必须与 manifest 顶层批次元数据一致

当前 `candidate_response_record` 允许为真实候选输出补最小来源元数据：

- `capture_metadata.capture_origin`：例如 `interactive_session`、`adapter_debug_dump`
- `capture_metadata.collection_batch`：用于把一批真实快照按同一采样批次收口
- `capture_metadata.tags`：用于标记 `real_capture`、任务名或专项回灌标签
- `capture_metadata.notes`：补充该快照的最小背景说明

当上游暂时还没有正式模型接入，但已经有 mock 推理器、adapter 调试输出或未来的模型 API 返回时，当前建议先落一份 raw dump，再导入为正式 record。

当前推荐的 raw dump 最小字段：

- `dump_id`
- `project` / `task` / `sample_id` / `request_id`
- `captured_at`
- `source`
- `model`
- `input_record`
- 已归一化到 `CopilotResponse` 结构的 `response`

可选保留的调试上下文：

- `input_request`
- `raw_request`
- `raw_response`

其中 raw dump 的职责是“保留调试现场”，而 `candidate_response_record` 的职责是“进入评测与回放链路”。两者不要混成一个文件格式。

当前推荐的真实负例回灌最小流程：

1. 若上游先产出的是 raw dump，先按 `candidate-response-dump.schema.json` 落盘，保留 `input_request` / `raw_response` 等调试上下文
2. 使用 `scripts/import-candidate-response-dump.py` 将 raw dump 裁剪为 `candidate-response-record.schema.json` 允许的正式 record
3. 将同一批真实快照补到 `candidate-record-batch.schema.json` 的 manifest 中，按 `collection_batch` 收口
4. 若该快照本身就是坏输出，可直接让负例样本通过 `manifest_path + record_id` 或直接 `path` 引用它
5. 若当前只有正向真实快照，也可以把它跨样本回放到另一条样本上，复用同一套 `candidate_record_alignment + response` 校验，验证“真实输出放错样本”会被稳定拒绝
6. 负例样本仍只通过 `negative_replay_expectations.expected_candidate_violations` 声明期望命中的 violation 片段，不再分叉第二套校验逻辑

最小导入命令示例：

```bash
python3 ./scripts/import-candidate-response-dump.py \
  --input /tmp/radish-docs-qa-bad-answer.json \
  --output datasets/eval/candidate-records/radish-negative/imported-bad-answer-001.json
```

如果需要直接按 `datasets/eval/radish/*.json` 批量跑最小推理，并把输出直接接到现有回灌链路，当前可直接使用：

```bash
python3 ./scripts/run-copilot-inference.py \
  --sample-dir datasets/eval/radish \
  --provider openai-compatible \
  --output-root /tmp/radish-docs-qa-batch \
  --collection-batch 2026-04-04-radish-docs-qa-real-batch-v1 \
  --manifest-description "Radish docs QA 批量真实候选输出快照。"
```

批量模式当前会在 `--output-root` 下生成：

- `responses/*.response.json`：归一化后的 `CopilotResponse`
- `dumps/*.dump.json`：保留 `input_request`、`raw_request`、`raw_response` 的 raw dump
- `records/*.record.json`：可直接被回归与 replay 使用的正式 `candidate_response_record`
- `<collection_batch>.manifest.json`：按同批次自动收口的 manifest

补充说明：

- `--collection-batch` 在批量模式下为必填，保证生成的 `record` 可直接纳入现有 manifest 回灌流程
- `--capture-origin` 可选覆盖默认来源；默认仍保持 `mock -> manual_fixture`、真实 provider -> `adapter_debug_dump`
- `--capture-tag` 可重复追加批次标签，但不会写入任何 API key
- `--max-attempts` 与 `--retry-base-delay-seconds` 可为真实 provider 补最小重试；当前默认会对 `429`、`5xx`、远端断连和部分网络抖动做指数退避重试
- `--resume` 可在批量模式下跳过已存在的 `response / dump / record`，适合真实 batch 中途被打断后断点续跑
- `--continue-on-error` 可在批量模式下记录失败样本并继续后续样本，最后统一输出失败清单
- 若只想单条推理，也可继续沿用原有 `--sample` / `--request` 入口；当前还额外支持 `--record-output`

当真实 batch 已导入仓库、但尚未准备好直接替换现有正向样本绑定时，当前建议先走批量审计入口，而不是立刻改写样本中的 `candidate_response_record` 引用：

```bash
python3 ./scripts/audit-candidate-record-batch.py \
  radish-docs-qa \
  --manifest datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.manifest.json \
  --report-output datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.audit.json
```

这一步会：

- 按 `sample_id` 将 manifest 中的真实 record 临时注入对应样本
- 自动补上 `expected_source`、`required_capture_origin`、`required_collection_batch`
- 继续复用 `scripts/run-eval-regression.py` 的既有规则，不分叉第二套校验逻辑
- 让“真实输出已入仓库”和“哪些真实输出已满足当前基线”先分离开
- 可选输出结构化 `audit.json`，把 pass/fail、violation 计数和逐样本违规明细沉淀为可追踪资产

如果希望在一次审计后顺手继续生成 replay index 和同样本真实 replay 负例，当前也可以直接使用：

```bash
python3 ./scripts/audit-candidate-record-batch.py \
  radish-docs-qa \
  --manifest datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.manifest.json \
  --report-output datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.audit.json \
  --replay-index-output datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.negative-replay-index.json \
  --build-negative-replay
```

这条链路当前会顺序完成：

- 批量审计并写出 `audit.json`
- 基于 audit 报告生成 `negative-replay-index.json`
- 基于 replay index 重建同样本真实 replay 负例

这样下一批真实 provider batch 进仓后，不需要再手工串行调用三段脚本。

如果希望把“真实 provider 批跑 -> manifest -> audit -> replay index -> same-sample negatives”收口成一次命令，当前也可以直接使用：

```bash
python3 ./scripts/run-radish-docs-qa-real-batch.py \
  --provider openai-compatible \
  --output-root /tmp/radish-docs-qa-real-batch \
  --collection-batch 2026-04-04-radish-docs-qa-real-batch-v1 \
  --manifest-description "Radish docs QA 批量真实候选输出快照。" \
  --build-negative-replay
```

这条更高层的入口当前会：

- 调用 `run-copilot-inference.py` 产出 `response / dump / record / manifest`
- 调用 `audit-candidate-record-batch.py` 产出 `audit.json`
- 顺手继续生成 `negative-replay-index.json`
- 若 audit 中存在失败样本，再继续重建同样本真实 replay 负例
- 额外写出 `<collection-batch>.artifacts.json`，把本批次的 `manifest / audit / replay index / same-sample replay` 产物路径与计数沉淀为结构化摘要

如果只是想验证编排链路，不想把 replay 负例写回仓库，可把 `--provider mock` 与 `--negative-output-dir /tmp/...` 组合使用。

`artifacts.json` 当前主要用于把高层编排产物批量交给后续治理脚本或人工复核，典型会包含：

- `manifest` 是否存在、路径和记录数
- `audit_report` / `negative_replay_index` 的目标路径与是否已写出
- `same_sample_negative_replay` 的目标目录与本批预计可复用的 same-sample 负例数量
- 本批次的 `passed / failed / violation / violation_group` 汇总计数
- `recommended_negative_replays`：按 same-sample 失败组汇总的默认复跑顺序，以及可直接执行的 `bash` / `python3` 命令

当前仓库内已新增第二批真实 batch：

- `datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.manifest.json`
- `datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1/`
- `datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.audit.json`
- `datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.negative-replay-index.json`

这批记录当前先作为“真实候选输出审计资产”保留，不直接替换现有正向样本绑定；待批量审计结果稳定后，再择优切换样本引用。

当一批真实审计已经把失败输出沉淀为 `datasets/eval/radish-negative/*.json` replay 样本后，当前推荐继续生成一份“负例索引”，不要再复制第二份真实 record manifest：

```bash
python3 ./scripts/build-negative-replay-index.py \
  --audit-report datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.audit.json \
  --output datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.negative-replay-index.json
```

这份索引的职责是：

- 继续以真实 batch manifest 作为 record 真相源，不复制同一批真实 record
- 扫描 `datasets/eval/radish-negative/` 下引用该 manifest 的 replay 负例
- 把“哪些 audit fail 已经沉淀为 replay 负例、按哪些 violation 片段分组、还剩哪些 fail 尚未回灌”收口成一份可追踪资产
- 为后续扩样时按 `group_id`、`expected_candidate_violations` 或 `record_id` 批量选样提供稳定索引

若想在仓库校验里确认索引仍与 audit 和负例样本一致，可执行：

```bash
python3 ./scripts/build-negative-replay-index.py \
  --audit-report datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.audit.json \
  --output datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.negative-replay-index.json \
  --check
```

当需要按这份索引批量复跑某一类真实 replay 负例时，当前可直接让负例 runner 按 `group_id`、`record_id` 或 `replay_mode` 选样，例如：

```bash
bash ./scripts/run-radish-docs-qa-negative-regression.sh \
  --negative-replay-index datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.negative-replay-index.json \
  --group-id group-004 \
  --fail-on-violation
```

这一步的目的不是新增另一套 runner，而是让现有 `radish-docs-qa-negative` 继续复用同一条校验链路，只把“选哪些负例样本”从手工文件列表切换为索引驱动。

如果上一层已经拿到了 `run-radish-docs-qa-real-batch.py` 产出的 `artifacts.json`，当前也可以不再手工传 `negative-replay-index` 路径，而是直接让负例 runner 从摘要里解析：

```bash
bash ./scripts/run-radish-docs-qa-negative-regression.sh \
  --batch-artifact-summary datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.artifacts.json \
  --group-id group-001 \
  --fail-on-violation
```

这条入口当前会：

- 先校验 `artifacts.json` 满足 `batch-orchestration-summary.schema.json`
- 从 `artifacts.negative_replay_index.path` 解析本批次的 replay index
- 继续复用现有 `group_id` / `record_id` / `replay_mode` 选样逻辑

这样“真实 batch 编排产物摘要”就不只是归档文件，而是可以直接驱动后续负例回放。

如果希望让高层编排在结束后直接给出“下一步建议怎么复跑失败组”，当前也可以直接查看 `artifacts.json` 里的 `recommended_negative_replays`。该区块会按 same-sample 失败组给出：

- `recommended_group_ids`：默认推荐优先复跑的失败组顺序，当前按组内样本数降序排列
- `groups[*].expected_candidate_violations`：该组对应的主要违规片段
- `groups[*].bash_command` / `groups[*].python_command`：可直接复制执行的复跑命令

这样真实 batch 跑完后，不需要再人工翻 `negative-replay-index.json` 组装命令。

如果希望直接按默认推荐顺序批量回放前 `N` 个失败组，当前也可以继续复用同一个负例 runner：

```bash
bash ./scripts/run-radish-docs-qa-negative-regression.sh \
  --batch-artifact-summary datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.artifacts.json \
  --recommended-groups-top 2 \
  --fail-on-violation
```

这条入口当前会：

- 从 `recommended_negative_replays.recommended_group_ids` 取前 `N` 个推荐失败组
- 若未显式指定 `--replay-mode`，默认采用摘要中的 `default_replay_mode`
- 继续复用同一条 `group_id` 选样和负例回放链路

这样真实 batch 跑完后，可以直接先回放“最值得优先治理”的前几组失败样本。

如果不希望记住底层 `run-eval-regression.py` 参数组合，当前也可以直接使用更高层的包装入口：

```bash
bash ./scripts/run-radish-docs-qa-negative-recommended.sh \
  --batch-artifact-summary datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.artifacts.json \
  --top 2 \
  --fail-on-violation
```

这条入口当前会：

- 直接转调 `run-eval-regression.py radish-docs-qa-negative`
- 自动使用 `--recommended-groups-top <top>`
- 保留 `--replay-mode` 和 `--sample-dir` 覆盖口

这样“真实 batch 跑完 -> 先回放推荐前 N 组失败样本”已经变成一条更短的日常操作命令。

如果需要把这份索引里的“同样本真实 replay”直接重建成 `datasets/eval/radish-negative/*.json`，当前还可使用：

```bash
python3 ./scripts/build-radish-docs-negative-replay.py \
  --index datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.negative-replay-index.json
```

脚本当前会：

- 读取 replay index、真实 batch manifest 和对应 record
- 从原正向样本复制 `input_request`、`retrieval_expectations`、`expected_response_shape`、`golden_response`
- 自动补 `candidate_response_record`、`negative_replay_expectations` 和标准化 `evaluation.notes`
- 为“同样本真实坏输出”推导稳定的负例文件名，例如 `partial-missing-read-only-check`、`low-risk-no-issue`

如需只检查这些已入仓负例是否仍可由脚本重建，可执行：

```bash
python3 ./scripts/build-radish-docs-negative-replay.py \
  --index datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.negative-replay-index.json \
  --check
```

若需要从一批记录重生成 manifest，当前可直接使用：

```bash
python3 ./scripts/build-candidate-record-batch.py \
  --record-dir datasets/eval/candidate-records/radish-negative \
  --output datasets/eval/candidate-records/radish-negative/2026-04-04-radish-docs-qa-simulated-negatives-v1.manifest.json \
  --description "Radish docs QA 现有 simulated negative 候选响应批次清单，用于把负例回放样本统一切到 manifest 导入入口。"
```

脚本会：

- 跳过同目录下已有的 `.manifest.json`
- 校验每条 record 都满足 `candidate-response-record.schema.json`
- 检查 `project`、`task`、`source` 以及 `capture_metadata.collection_batch` 是否能稳定收口
- 产出按仓库相对路径引用的 manifest 记录清单

对于显式启用 `official_source_precedence` 的 `Radish` 多来源问答样本，当前回归还会检查：

- 主回答必须至少引用一次 `primary` artifact
- 如果回答引用了 `faq` / `forum` 这类非正式来源，也必须同时引用至少一个正式来源

当前 `Radish` docs QA 还额外支持负例回放样本：

- 负例样本继续复用 `radish-task-sample.schema.json`
- 可选使用 `negative_replay_expectations.expected_candidate_violations` 声明预期触发的 violation 片段
- 负例 runner 仍会先校验 `golden_response` 与召回输入边界，再回放候选回答，要求其命中指定违规口径
- 当前首批负例聚焦 `official_source_precedence`，验证 runner 不只是“放行正确回答”，也能稳定拦截“只引用 forum/faq”或“未引用 primary artifact”的错误回答
- 当前负例还已覆盖“首答合规但后续回答只靠非正式来源”与“`citation_ids` 指向缺失 citation”两类真实回放错误
- 当前负例还已覆盖“普通问答里无端报 issue / 塞 action”与“read_only_check 动作被错误升级为中风险或要求确认”两类输出面越界
- 当前负例还已覆盖“证据不足样本被错误写成 `status=ok` / `risk_level=low`”以及“`issue.severity` 被错误升级或摘要直接写出越界确定性结论”两类分级口径错误
- 当前负例还已覆盖“可降级回答被过度写成 `failed/high/需确认`”以及“多 `issues` 叠加后错误打开 `requires_confirmation`”两类组合失稳
- 当前负例还已覆盖“`failed` 态里仍夹带 action / 确认要求”以及“`failed` 说明与确定性强答并存”的混合漂移
- 当前负例还已覆盖“`failed` 态只靠 `forum` / `faq` 这类非正式来源”以及“`failed` 态里 primary/citation 集合继续漂移”的来源与引用失稳
- 当前负例还已覆盖 `docs + attachments + forum` 与 `docs + faq + forum` 三路冲突下的 failed-state 漂移，验证失败态里社区经验、FAQ 或附件状态都不能反客为主
- 当前负例还已覆盖 `docs + attachments + faq + forum` 极端冲突与 failed-state 多 action 漂移，验证失败态里 FAQ/论坛不能联手覆盖正式口径，也不能一边失败一边继续堆动作
- 当前负例还已覆盖“多答案 + 多 action + 多来源冲突”三者同时出现的复合失稳，开始逼近真实模型完全漂移时的坏输出形态
- 当前负例已开始直接复用第二批真实 provider batch 的失败 record，并已把第二批 `8 fail` 全部沉淀为同样本 replay：一组是缺 `read_only_check`，另一组是 `partial` 场景下 `issues/risk_level` 不达标的真实坏输出

当前负例侧也已补最小 manifest 导入流程：

- `datasets/eval/candidate-records/radish-negative/2026-04-04-radish-docs-qa-simulated-negatives-v1.manifest.json` 已将现有 simulated negative 记录收口成一批
- `datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.negative-replay-index.json` 已将第二批真实 batch 的 `8 fail` replay 负例按 violation 分组收口
- `scripts/build-radish-docs-negative-replay.py` 已可根据这份 index 重建第二批真实 batch 的同样本 replay 负例
- 对应负例样本已可通过 `candidate_response_record.manifest_path + record_id` 引用，不再逐条手写单独路径
- 这一步的目的只是先稳定负例批量导入形态，不代表这批样本已经变成真实 captured negative
- 后续一旦拿到真实坏输出，应继续沿用同一入口，把真实 record 逐步补进新的 captured batch manifest
- 第二批真实 batch 当前已把 `8 fail` 全部沉淀为 `radish-negative` replay 样本，并可通过上述索引直接看出哪些失败类型已覆盖、哪些批次仍有漏灌

当前 `Radish` docs QA 已开始把部分代表性样本切到外部回灌记录：

- 角色示例不等于最终授权
- `docs + attachments + forum` 冲突样本
- `wiki + faq` 混合样本

当前 `Radish answer_docs_question` 已覆盖的最小样本类型包括：

- 直接回答
- 证据不足降级
- 低风险导航回答
- 角色示例不等于最终授权的授权边界样本
- `docs + attachments` 混合召回
- `docs + forum` 混合召回，且要求正式文档优先于社区经验
- `docs + faq` 混合召回，且 FAQ 只作为正式文档的补充说明
- `wiki + faq` 混合召回，且恢复/操作流程仍以 wiki 主结论为准
- `docs + faq + forum` 三路混合召回，且需显式拒绝用社区经验覆盖正式发布规则
- `docs + attachments + faq` 三路混合召回，且 FAQ 只能补充可读性说明，不能替代正式附件引用规则
- `docs + attachments + forum` 三路冲突样本，且需显式拒绝用社区经验覆盖正式附件引用与外链暴露规则

后续再补更多真实候选输出记录、最小对照指标与更完整的回灌流程。
