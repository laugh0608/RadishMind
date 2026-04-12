# `2026-04-05-radish-docs-qa-real-batch-v1` 治理说明

更新时间：2026-04-12

本目录存放 `Radish docs QA` 的一批真实 `candidate_response_record`、对应 `manifest / audit / replay index / recommended replay summary`，当前同时承担两类用途：

- 作为真实 bad record 的真相源，保留当时 provider 的原始失败面
- 作为 `same_sample / cross_sample recommended replay summary` 的只读输入批次，驱动负例治理链回放

## 批次快照

- manifest: `2026-04-05-radish-docs-qa-real-batch-v1.manifest.json`
- audit: `2026-04-05-radish-docs-qa-real-batch-v1.audit.json`
- same-sample replay index: `2026-04-05-radish-docs-qa-real-batch-v1.negative-replay-index.json`
- cross-sample replay index: `2026-04-05-radish-docs-qa-real-batch-v1.cross-sample-replay-index.json`
- artifact summary: `2026-04-05-radish-docs-qa-real-batch-v1.artifacts.json`

当前批次共有 11 条样本：

- `passed_count=1`
- `failed_count=10`
- `violation_count=22`

唯一通过的样本是 `answer-docs-question-direct-answer-001.json`。其余 10 条失败并不是仓库基线“未修复”的随机噪声，而是当前治理链刻意保留并复用的真实 bad record 输入。

## 失败分组

当前 22 条违规可收口为 4 类主失败面。

### 1. `read_only_check` 缺失

这是当前最大的一组真实失败面，共覆盖 6 条样本、11 条违规：

- `answer-docs-question-attachment-mixed-001.json`
- `answer-docs-question-docs-attachments-faq-001.json`
- `answer-docs-question-docs-attachments-forum-conflict-001.json`
- `answer-docs-question-docs-faq-forum-conflict-001.json`
- `answer-docs-question-docs-faq-mixed-001.json`
- `answer-docs-question-navigation-001.json`

共同特征：

- 回答正文通常还能给出大体可读的说明
- 但缺少当前任务明确要求的 `read_only_check` 动作
- 部分样本同时缺失 `$.proposed_actions[0].kind="read_only_check"`

治理现状：

- same-sample replay 已全部覆盖，见 `negative-replay-index` 中的 `group-003`
- cross-sample replay 现已补成一组 6 条的复用簇，见 `cross-sample-replay-index` 中的 `group-003`
- 当前 real-derived repeated pattern 已围绕这组失败扩成 `missing_read_only_check_confirmation_drift`、`missing_read_only_check_issue_drift` 与 `missing_read_only_check_issue_confirmation_drift`

### 2. 风险分级偏低且未产出 issue

这是第二组“解释看起来接近正确，但风险与结构判定偏轻”的真实失败面，共覆盖 2 条样本、7 条违规：

- `answer-docs-question-evidence-gap-001.json`
- `answer-docs-question-role-example-boundary-001.json`

共同特征：

- `risk_level` 低于样本预期
- `issues=[]`，没有把证据不足或授权未证实明确结构化输出
- `role-example-boundary` 还额外缺失 `AUTHORIZATION_NOT_PROVEN`

治理现状：

- same-sample replay 已覆盖，见 `negative-replay-index` 中的 `group-001` 与 `group-002`
- `role-example-boundary` 当前仍通过 `wiki-faq-mixed -> role-boundary` 这条 cross-sample 负例保留在推荐链中，用于验证跨样本回灌时仍会暴露更强的授权/风险违规
- real-derived repeated pattern 已继续扩成 `evidence-gap` 的 `multi_issues_confirmation` / `unconfirmed_operation`，以及 `role-boundary` 的 `missing_read_only_check_confirmation` / `multi_issues_confirmation`

### 3. `answers` 缺失

这一组当前只落在 1 条样本上，共 2 条违规：

- `answer-docs-question-forum-supplement-001.json`

共同特征：

- summary 存在，但 `answers=[]`
- 同时缺少 `$.answers[0]`

治理现状：

- same-sample replay 已覆盖，见 `negative-replay-index` 中的 `group-004`
- 当前 cross-sample recommended replay 也会复用这条失败面，把它作为“缺 `answer` + 缺 `read_only_check`”的组合治理入口之一
- real-derived repeated pattern 已继续扩出 `missing_answer_issue_action`

### 4. citation / official-source 漂移

这一组当前只落在 1 条样本上，共 2 条违规：

- `answer-docs-question-wiki-faq-mixed-001.json`

共同特征：

- answer 引用了 `faq-1`
- 但 `citations` 中并未提供 `faq-1`
- 同时没有满足 `official_source_precedence` 样本对“首条 answer 至少引用一个 primary artifact”的要求

治理现状：

- same-sample replay 已覆盖，见 `negative-replay-index` 中的 `group-005`
- cross-sample recommended replay 当前也明确挑中了这条失败面，见 `cross-sample-replay-index` 中的 `group-001`
- real-derived repeated pattern 已继续扩成 `citation_drift_action_confirmation` 与 `citation_drift_issue_confirmation`

## 当前覆盖状态

从治理角度看，这批真实失败目前已经不是“散点样本”，而是进入了可追踪状态：

- same-sample replay: `failed_sample_count=10`，`linked_failed_sample_count=10`，当前无未回灌失败样本
- cross-sample replay: `failed_sample_count=10`，`linked_failed_sample_count=8`，当前还剩 2 条失败未进入 cross-sample 推荐链
- violation groups: same-sample 当前收口到 5 组，cross-sample 当前收口到 3 组

这意味着当前仓库已经能稳定回答三个问题：

- 哪些真实失败已沉淀为同样本 replay
- 哪些失败已经值得跨样本复用
- 哪些失败虽然已有 same-sample replay，但尚未进入 cross-sample 推荐链

当前 remaining only-same-sample source 已缩到两条：

- `answer-docs-question-evidence-gap-001.json`
- `answer-docs-question-role-example-boundary-001.json`

## 为什么 `check-repo` 末尾会打印 `FAIL`

`check-repo` 末尾的三条 summary checker 会重新调用这批真实 bad record，验证：

- batch orchestration 是否还能稳定产出 `artifacts.json`
- same-sample / cross-sample recommended summary 是否还能稳定生成
- 推荐组选择与元数据是否仍与既有 committed 口径一致

因此该阶段打印的 `FAIL/WARNING` 是这批真实失败输入本身的审计结果，不代表仓库基线失败。真正的通过条件是：

- `check-radish-docs-qa-real-batch-*.py` 最终退出码为 `0`
- summary 文件存在且 schema-valid
- `recommended_negative_replay_exit_code` / `cross_sample_recommended_negative_replay_exit_code` 为 `0`

## 下一步建议

当前建议优先继续做两件事：

1. 继续把剩余 2 条 only same-sample 的真实失败推进到 cross-sample 推荐链，当前优先只剩“低风险无 issue”这一类主问题。
2. 继续把这批真实失败派生成更稳定的 repeated pattern，但保持“真实 bad record 原失败面 + 一层额外漂移”的口径，不回退到纯 simulated negative 堆叠。
