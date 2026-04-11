# `RadishFlow` 真实候选记录说明

更新时间：2026-04-11

当前目录用于存放 `RadishFlow` 任务的正式 `candidate_response_record`、批次 `manifest` 与 `audit` 治理产物。

当前已正式入仓的 `suggest_ghost_completion` 真实 provider 批次有六批：

- `datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v2/`
- `datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v3/`
- `datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v4/`
- `datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v5/`
- `datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v6/`
- `datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v7/`

这六批当前都只收口同一组 3 条记录，原因不是“样本越少越好”，而是这 3 条已经同时满足：

- 对应默认 PoC 三条主路径：`Tab` / `manual_only` / `empty`
- 真实 raw dump 经过当前 runtime 归一化后，可回放通过 `radishflow-ghost-completion` 回归
- 已能生成正式 `manifest` 与 `audit`，因此适合作为第一批正式治理资产入仓
- 其中第二批 `v3` 还额外证明：即使真实 provider 返回“几乎完整但多闭合一个 `}`”的 malformed JSON，只要该坏法已被当前 runtime 以窄修复收口，原始 dump 仍可重新归一化后导回正式批次
- 第三批 `v4` 则额外确认：当前未观察到新的可重新归一化结构坏法；新增现象主要是批处理执行时 `manual_only` 样本一度出现 provider 卡顿，但拆成单样本后仍可成功 capture 并沿既有导入链入仓
- 第四批 `v5` 则继续验证：在批次入口改为逐样本单进程并补硬超时后，单样本卡顿不再把整批拖死；同时 `manual_only` 多动作输出里新出现的作用域提前闭合坏法，也已被当前 runtime 重新归一化后导回正式批次
- 第五批 `v6` 则进一步验证：当默认 `openrouter` 在同一时间窗内出现 provider-wide `HTTP 429` 时，可切到 `deepseek` fallback profile 继续完成同一组 `Tab / manual_only / empty` 三样本真实 capture，且当前未观察到新的结构坏法，最终仍收口到 `audit=3/3 pass`
- 第六批 `v7` 则继续固定两条新的 provider 侧观察项：其一是当前 `.env.example` 里的 openrouter 默认 free 模型已被上游废弃，直接返回 `404 deprecated`；其二是 `deepseek` 在 `empty` 样本上曾把 `summary` / `answer.text` 写成内嵌 JSON 字符串。当前这两条都已分别通过配置口径更新与 runtime 窄修复收口，最终仍导回正式批次并收口到 `audit=3/3 pass`

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
- 正式批次目录应尽量只保留上述治理资产，不混入执行态 `dumps/`、`responses/` 或中间 `records/` 子目录
- 只有通过当前回归 / audit 的记录，才应进入首批正式正向批次
- 若后续真实 capture 暴露出新的失败面，应优先新增下一批真实 batch，而不是回头篡改已入仓批次
- 若后续继续遇到类似 `v4` 的单样本 provider 卡顿，应优先视为 capture orchestration 稳定性观察项，先复跑或拆样本确认，再决定是否需要继续加强脚本级超时/重试治理
