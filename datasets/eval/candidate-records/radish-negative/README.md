# `Radish` docs QA 负例候选记录说明

更新时间：2026-04-04

当前目录继续存放 `answer_docs_question` 的负例候选响应记录。

当前记录来源分两类：

- `simulated_candidate_response`：为了快速覆盖某一类稳定违规口径而构造的回放记录
- `captured_candidate_response`：真实候选输出快照，可直接用于同样本回灌，或跨样本回灌

当前推荐的最小真实负例回灌方式：

1. 先把真实快照存成 `candidate-response-record.schema.json` 允许的记录文件
2. 在 `capture_metadata` 中补 `capture_origin`、`collection_batch`、`tags`
3. 将同批次记录收口到一个 manifest，优先按 `collection_batch` 维护
4. 若该快照直接对应某条负例样本，就按 `manifest_path + record_id` 或原始 `path` 引用
5. 若当前还没有足够多的真实坏输出，也可以先把已有真实快照跨样本回放到另一条样本上，验证 `candidate_record_alignment` 和响应级规则会共同拦截

当前仓库里首批跨样本真实回放负例位于：

- `datasets/eval/radish-negative/answer-docs-question-negative-cross-sample-real-record-role-boundary-001.json`
- `datasets/eval/radish-negative/answer-docs-question-negative-cross-sample-real-record-docs-attachments-forum-001.json`
- `datasets/eval/radish-negative/answer-docs-question-negative-cross-sample-real-record-wiki-faq-001.json`

这些样本继续复用现有负例 runner，不新增第二套规则。

当前首个真实 captured batch manifest 位于：

- `datasets/eval/candidate-records/radish/2026-04-03-radish-docs-qa-real-captures-v1.manifest.json`

当前还新增了一份 simulated negative batch manifest：

- `datasets/eval/candidate-records/radish-negative/2026-04-04-radish-docs-qa-simulated-negatives-v1.manifest.json`

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
