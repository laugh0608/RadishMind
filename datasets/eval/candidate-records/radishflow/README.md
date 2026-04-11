# `RadishFlow` 真实候选记录说明

更新时间：2026-04-11

当前目录用于存放 `RadishFlow` 任务的正式 `candidate_response_record`、批次 `manifest` 与 `audit` 治理产物。

当前首批正式入仓的是 `suggest_ghost_completion` 的真实 provider 批次：

- `datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v2/`

这批次当前只收口 3 条记录，原因不是“样本越少越好”，而是这 3 条已经同时满足：

- 对应默认 PoC 三条主路径：`Tab` / `manual_only` / `empty`
- 真实 raw dump 经过当前 runtime 归一化后，可回放通过 `radishflow-ghost-completion` 回归
- 已能生成正式 `manifest` 与 `audit`，因此适合作为第一批正式治理资产入仓

当前不建议把 `/tmp` 下的旧 `record` / `manifest` / `audit` 直接复制进仓库。  
若这批 dump 采集于 runtime canonicalization 修复之前，应先按当前 runtime 重新归一化 `dump.response`，再导入正式批次。

当前推荐导入方式：

```bash
python3 ./scripts/import-candidate-response-dump-batch.py \
  radishflow-ghost-completion \
  --dump-dir /tmp/radishflow-ghost-real-batch-v2/dumps \
  --output-dir datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v2 \
  --manifest-description "RadishFlow suggest_ghost_completion 首批真实 provider 候选响应快照，覆盖 Tab/manual_only/empty 三条正式导入主路径。" \
  --recanonicalize-response
```

若只需要处理单个 dump，当前也可先执行：

```bash
python3 ./scripts/import-candidate-response-dump.py \
  --input /tmp/example.dump.json \
  --output datasets/eval/candidate-records/radishflow/example.record.json \
  --recanonicalize-response
```

当前治理口径：

- 原始 raw dump 继续保留在临时或外部采集目录，不强制入仓
- 正式入仓资产以 `candidate_response_record`、`manifest`、`audit` 为主
- 只有通过当前回归 / audit 的记录，才应进入首批正式正向批次
- 若后续真实 capture 暴露出新的失败面，应优先新增下一批真实 batch，而不是回头篡改已入仓批次
