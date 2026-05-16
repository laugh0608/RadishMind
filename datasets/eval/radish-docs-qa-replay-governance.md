# Radish docs QA 真实 batch 与 replay 治理说明

更新时间：2026-05-16

本文档承接 `datasets/eval/README.md` 中拆出的 `Radish answer_docs_question` 外部记录、真实 batch、负例 replay 和 real-derived 治理细节。

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

- `o/*.response.json`：归一化后的 `CopilotResponse`
- `d/*.dump.json`：保留 `input_request`、`raw_request`、`raw_response` 的 raw dump
- `r/*.record.json`：可直接被回归与 replay 使用的正式 `candidate_response_record`
- `manifest.json`：按同批次自动收口的 manifest
- `audit.json`：批次审计结果
- `artifacts.json`：最小批次资产摘要

对 `RadishFlow` 而言，当前正式推荐的 `--output-root` 不再直接使用长 `collection_batch` 目录，而是收口到：

```text
datasets/eval/candidate-records/radishflow/batches/YYYY-MM/<batch_key>/
```

其中长 `collection_batch` 语义继续保留在 `manifest.json` 元数据中，不再直接写进物理目录和文件名。

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
- 若传入 `--cross-sample-negative-sample-dir`，再额外生成 `cross-sample-replay-index.json`
- 若 audit 中存在失败样本，再继续重建同样本真实 replay 负例
- 若传入 `--build-recommended-negative-replay-summary`，默认会先产 same-sample 推荐摘要；当同一批次已存在 cross-sample 推荐组时，再自动追加一份 cross-sample 推荐摘要
- 额外写出 `<collection-batch>.artifacts.json`，把本批次的 `manifest / audit / replay index / cross-sample replay index / same-sample replay / recommended replay summaries` 产物路径与计数沉淀为结构化摘要

若希望控制推荐摘要的默认输出，当前可继续使用：

- `--recommended-groups-top` / `--recommended-summary-output`：控制 same-sample 推荐摘要
- `--cross-sample-recommended-groups-top` / `--cross-sample-recommended-summary-output`：控制 cross-sample 推荐摘要
- `--recommended-replay-mode`：若显式指定，则退回为只生成单一模式的推荐摘要

当前仓库基线已分别覆盖三条参数分支：

- 默认不显式指定 `--recommended-replay-mode` 时，若批次里存在 cross-sample 推荐组，则会自动同时产出 same-sample 与 cross-sample 两份摘要
- 显式指定 `--recommended-replay-mode cross_sample` 时，只会产出 cross-sample 摘要，并保持 same-sample 摘要位点处于未请求状态
- 显式指定 `--recommended-replay-mode same_sample` 时，只会产出 same-sample 摘要，并保持 cross-sample 摘要位点处于未请求状态

如果只是想验证编排链路，不想把 replay 负例写回仓库，可把 `--provider mock` 与 `--negative-output-dir /tmp/...` 组合使用。

`artifacts.json` 当前主要用于把高层编排产物批量交给后续治理脚本或人工复核，典型会包含：

- `manifest` 是否存在、路径和记录数
- `audit_report` / `negative_replay_index` / 可选 `cross_sample_negative_replay_index` 的目标路径与是否已写出
- `same_sample_negative_replay` 的目标目录与本批预计可复用的 same-sample 负例数量
- `recommended_negative_replay_summary` 与可选 `cross_sample_recommended_negative_replay_summary` 的目标路径、是否已写出与退出码
- 本批次的 `passed / failed / violation / violation_group` 汇总计数，以及 same-sample / cross-sample 两路推荐回放与 cross-sample linked/unlinked 计数
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

如果希望让高层编排在结束后直接给出“下一步建议怎么复跑失败组”，当前也可以直接查看 `artifacts.json` 里的 `recommended_negative_replays`。该区块默认仍以 same-sample 为主，但当前也可额外携带 cross-sample 推荐组：

- `recommended_group_ids`：默认推荐优先复跑的失败组顺序，当前按组内样本数降序排列
- `cross_sample_recommended_group_ids`：若当前 batch 已提供 cross-sample replay index，可额外给出 cross-sample 推荐组顺序
- `groups[*].expected_candidate_violations`：该组对应的主要违规片段
- `groups[*].bash_command` / `groups[*].python_command`：same-sample 推荐组的可直接执行复跑命令
- `cross_sample_groups[*].bash_command` / `cross_sample_groups[*].python_command`：cross-sample 推荐组的可直接执行复跑命令

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
- 若显式传入 `--replay-mode cross_sample`，则会改为消费 `recommended_negative_replays.cross_sample_recommended_group_ids`
- 若同时传入 `--summary-output`，当前可把 same-sample 或 cross-sample 的推荐回放结果都沉淀成 committed 的结构化摘要

这样“真实 batch 跑完 -> 先回放推荐前 N 组失败样本”已经变成一条更短的日常操作命令。

当前仓库也已把 `2026-04-05-radish-docs-qa-real-batch-v1` 的两种推荐回放摘要都正式入仓并接入 `check-repo`：

- `2026-04-05-radish-docs-qa-real-batch-v1.recommended-negative-replay-top5-same_sample.summary.json`
- `2026-04-05-radish-docs-qa-real-batch-v1.recommended-negative-replay-top2-cross_sample.summary.json`

这样 same-sample 与 cross-sample 的推荐治理结果都不再只是临时 `/tmp` 产物，而是仓库基线的一部分。

如果需要把这份索引里的真实 replay 直接重建成 `datasets/eval/radish-negative/*.json`，当前还可使用：

```bash
python3 ./scripts/build-radish-docs-negative-replay.py \
  --index datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.negative-replay-index.json \
  --replay-mode same_sample
```

脚本当前会：

- 读取 replay index、真实 batch manifest 和对应 record
- 对 `same_sample`：从原正向样本复制 `input_request`、`retrieval_expectations`、`expected_response_shape`、`golden_response`
- 对 `cross_sample`：复用现有 cross-sample 负例文件作为模板，重建其中的 `candidate_response_record`、`negative_replay_expectations` 与标准化说明文本
- 自动补 `candidate_response_record`、`negative_replay_expectations` 和标准化 `evaluation.notes`
- 为“同样本真实坏输出”推导稳定的负例文件名，例如 `partial-missing-read-only-check`、`low-risk-no-issue`

如需只检查这些已入仓负例是否仍可由脚本重建，可执行：

```bash
python3 ./scripts/build-radish-docs-negative-replay.py \
  --index datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.negative-replay-index.json \
  --replay-mode same_sample \
  --check
```

如果当前索引里已经有 cross-sample 条目，也可以按模式单独重建：

```bash
python3 ./scripts/build-radish-docs-negative-replay.py \
  --index /tmp/radish-docs-qa-with-cross-sample.negative-replay-index.json \
  --replay-mode cross_sample
```

补充说明：

- `--replay-mode` 当前支持 `same_sample`、`cross_sample`、`all`
- `cross_sample` 重建当前要求索引里的 `negative_sample_path` 已存在，因为它会把现有 cross-sample fixture 当作模板再标准化重建
- 若同时传 `--group-id`，当前只会命中索引 `violation_groups` 中已有分组的 replay 条目；未分组的 cross-sample 条目仍可通过 `--record-id` 或 `--replay-mode cross_sample` 选择

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
- 当前负例还补了十三条基于 `2026-04-05` real record 派生的本地候选输出：`evidence-gap` 上叠加 `candidate_operation + requires_confirmation=true`，`role-example-boundary` 上叠加多 `issues + requires_confirmation=true`，`wiki-faq-mixed` 上叠加 citation drift 之外的 `issues / read_only_check / requires_confirmation`，`forum-supplement` 与 `direct-answer` 两条短答型 source 上分别派生出 `answers` 缺失再叠加 `issues / read_only_check` 的复合变体，六条围绕 `read_only_check` 缺失继续叠加 `issue / confirmation / candidate_operation` 的复合变体，以及两条围绕 forum citation/source drift 再叠加 `issue + confirmation` 型 `read_only_check` 的复用样本
- 这批 `real-derived` candidate record 现已补上结构化 `source_candidate_response_record` 来源字段，并额外生成 `datasets/eval/candidate-records/radish-negative/2026-04-04-radish-docs-qa-simulated-negatives-v1.real-derived-index.json`，可直接审计“派生自哪条 real batch record、覆盖了哪组额外违规、当前是否全部挂到负例样本”

当前负例侧也已补最小 manifest 导入流程：

- `datasets/eval/candidate-records/radish-negative/2026-04-04-radish-docs-qa-simulated-negatives-v1.manifest.json` 已将现有 simulated negative 与 real-record-derived negative 记录收口成一批
- `datasets/eval/candidate-records/radish-negative/2026-04-04-radish-docs-qa-simulated-negatives-v1.real-derived-index.json` 已把 manifest 中带 `real_record_derived` 标签的候选记录，和其对应负例样本、源 real batch record 做了结构化关联
- 当前这份 real-derived index 已覆盖十三条派生记录、八类更接近真实 provider 措辞的复合失稳：`issues + proposed_actions + requires_confirmation`、citation/source drift 叠加越界动作、`answers` 缺失继续叠加 issue/action，以及 `read_only_check` 缺失继续叠加 `issue / confirmation / candidate_operation`
- 当前这份 real-derived index 还会从 `capture_metadata.tags` 中解析 `derived_pattern:*`，额外输出 `pattern_groups`，用来区分“同源 real record 下到底派生了哪一类复合漂移”；当前已经出现 5 组 pattern 被不同 source record 复用、各自 `entry_count=2` 的场景，其中新增一组是 `missing_answer_issue_action_drift` 开始跨 `forum-supplement` 与 `direct-answer` 两条真实 source 复用
- `violation_groups` 仍按完整 `expected_candidate_violations` 文案分组；因此当 confirmation 型负例的 action title 不同时，即使复用同一 `derived_pattern`，也会继续拆成不同 violation group
- `datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.negative-replay-index.json` 已将第二批真实 batch 的 `8 fail` replay 负例按 violation 分组收口
- `scripts/build-radish-docs-negative-replay.py` 已可根据这份 index 按 `replay_mode` 重建 same-sample / cross-sample replay 负例
- 对应负例样本已可通过 `candidate_response_record.manifest_path + record_id` 引用，不再逐条手写单独路径
- 这一步的目的只是先稳定负例批量导入形态，不代表这批样本已经变成真实 captured negative
- 后续一旦拿到真实坏输出，应继续沿用同一入口，把真实 record 逐步补进新的 captured batch manifest
- 第二批真实 batch 当前已把 `8 fail` 全部沉淀为 `radish-negative` replay 样本，并可通过上述索引直接看出哪些失败类型已覆盖、哪些批次仍有漏灌
- 当前还额外提交了一份 `2026-04-05-radish-docs-qa-real-batch-v1.cross-sample-replay-index.json`，专门收口根目录 `radish-negative` 中已入仓的 cross-sample replay，并已接入 `check-repo`

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
