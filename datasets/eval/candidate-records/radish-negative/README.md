# `Radish` docs QA 负例候选记录说明

更新时间：2026-04-04

当前目录继续存放 `answer_docs_question` 的负例候选响应记录。

当前记录来源分两类：

- `simulated_candidate_response`：为了快速覆盖某一类稳定违规口径而构造的回放记录
- `captured_candidate_response`：真实候选输出快照，可直接用于同样本回灌，或跨样本回灌

当前推荐的最小真实负例回灌方式：

1. 若上游先产出 raw dump，先保留该 dump，不要手工删掉 `raw_request` / `raw_response`
2. 用 `scripts/import-candidate-response-dump.py` 转成 `candidate-response-record.schema.json` 允许的记录文件
3. 在 `capture_metadata` 中补 `capture_origin`、`collection_batch`、`tags`
4. 将同批次记录收口到一个 manifest，优先按 `collection_batch` 维护
5. 若该快照直接对应某条负例样本，就按 `manifest_path + record_id` 或原始 `path` 引用
6. 若当前还没有足够多的真实坏输出，也可以先把已有真实快照跨样本回放到另一条样本上，验证 `candidate_record_alignment` 和响应级规则会共同拦截

当前仓库里首批跨样本真实回放负例位于：

- `datasets/eval/radish-negative/answer-docs-question-negative-cross-sample-real-record-role-boundary-001.json`
- `datasets/eval/radish-negative/answer-docs-question-negative-cross-sample-real-record-docs-attachments-forum-001.json`
- `datasets/eval/radish-negative/answer-docs-question-negative-cross-sample-real-record-wiki-faq-001.json`

这些样本继续复用现有负例 runner，不新增第二套规则。

当前首个真实 captured batch manifest 位于：

- `datasets/eval/candidate-records/radish/2026-04-03-radish-docs-qa-real-captures-v1.manifest.json`

当前还新增了一份 simulated negative batch manifest：

- `datasets/eval/candidate-records/radish-negative/2026-04-04-radish-docs-qa-simulated-negatives-v1.manifest.json`

针对真实 provider batch，当前不再额外复制第二份“negative record manifest”来重登记同一批真实 record。
现在推荐改为保留真实 batch manifest 作为真相源，再为对应 audit 生成一份 replay 索引，例如：

- `datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.negative-replay-index.json`

这份索引会把：

- 哪些 `audit fail` 已经沉淀为 `datasets/eval/radish-negative/*.json`
- 各条 replay 负例对应的 `record_id`
- 同类 `expected_candidate_violations` 的分组情况
- 尚未回灌的失败样本

统一收口，避免后续扩样时继续手工点文件、同时又维护两份真实 record 清单。

补充说明：

- 这份 manifest 只是把现有负例 fixture 统一切到批量导入入口，便于后续维护
- 它不代表仓库里已经有同规模的真实 captured negative
- 下一步仍应优先把真实坏输出按新的批次入口补进来，再逐步替换或扩充 simulated negative

如需在补充或替换记录文件后重生成这份 manifest，当前建议使用：

```bash
python3 ./scripts/build-candidate-record-batch.py \
  --record-dir datasets/eval/candidate-records/radish-negative \
  --output datasets/eval/candidate-records/radish-negative/2026-04-04-radish-docs-qa-simulated-negatives-v1.manifest.json \
  --description "Radish docs QA 现有 simulated negative 候选响应批次清单，用于把负例回放样本统一切到 manifest 导入入口。"
```

如果未来先拿到的是 raw dump，当前建议先执行：

```bash
python3 ./scripts/import-candidate-response-dump.py \
  --input /tmp/radish-docs-qa-bad-answer.json \
  --output datasets/eval/candidate-records/radish-negative/imported-bad-answer-001.json
```
