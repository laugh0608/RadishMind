# `RadishFlow` 任务卡附件：`suggest_ghost_completion` 候选上下文与样本覆盖

更新时间：2026-05-16

本文档承接主任务卡中拆出的候选集装配、recent actions、chain pattern 和样本覆盖细节。

当前推荐本地规则层先按独立候选集契约落一次中间对象，再装配到模型请求：

- 候选集契约参考 [radishflow-ghost-candidate-set.schema.json](../../contracts/radishflow-ghost-candidate-set.schema.json)
- 装配脚本参考 [build-radishflow-ghost-request.py](../../scripts/build-radishflow-ghost-request.py)
- 当前默认装配 profile 为 `model-minimal`；若需要排查候选排序或命名证据，可切到 `debug-full` 对照 profile，但不应把它作为常态模型输入
- 这样前端、本地规则层和模型层都能围绕同一份 `legal_candidate_completions` 结构协作，而不是各自私有拼接

当前推荐默认装配策略：

- 将 `ranking_signals`、`naming_signals`、`conflict_flags` 保留在本地候选集对象
- 只把模型判断确实需要的候选摘要透传到 `CopilotRequest`
- 若后续需要做调试或对照，可再显式切到更完整的装配 profile，而不是默认把本地证据全量喂给模型

当前建议本地规则层尽量把候选证据也结构化透传出来：

- `legal_candidate_completions[*].ranking_signals`
- `legal_candidate_completions[*].naming_signals`
- `legal_candidate_completions[*].conflict_flags`

这样模型排序时不必重新发明本地几何/拓扑判断，回归也能直接校验“为何允许 Tab、为何不能默认 Tab”。

对连续搭建链场景，推荐额外透传：

- `cursor_context.recent_actions[*].kind`
- `cursor_context.recent_actions[*].candidate_ref`
- `accept_ghost_completion -> accepted_at_revision`
- `reject_ghost_completion -> rejected_at_revision`
- `dismiss_ghost_completion -> dismissed_at_revision`
- `skip_ghost_completion -> skipped_at_revision`

这样模型可以明确知道“上一步刚接受了哪条 ghost”或“刚否掉了哪条 ghost”，而不必只靠模糊的画布静态快照猜测当前处于哪一步。

但 `recent_actions` 只用于表达链式上下文，不得覆盖本地规则层的合法候选边界：

- 若 `legal_candidate_completions=[]`，即使刚接受过上一条 ghost，也仍应允许返回空建议
- 不应因为“链已经开始”就跳过本地规则层，主观补一个并不存在的下一跳
- 若同一 `candidate_ref` 刚被 `reject` / `dismiss` / `skip`，则它当前至多只能保留为 `manual_only` 或直接退回空建议，不应立即再次被升级成默认 `Tab`
- 当前第一版交互语义先明确收口为：`reject` / `dismiss` / `skip` 都共享“同一 candidate 的即时 suppress-Tab”语义；若该候选仍然合法，可继续展示，但只能 `manual_only`
- 该 suppress 语义当前同样明确限定在“同一 `candidate_ref`”范围内：若最近被否掉的是另一条 candidate，而当前高置信候选已经切换到新的 `candidate_ref`，则不应被旧反馈误伤
- 当前最小恢复窗口也已先固定一条共用基线：同一 `candidate_ref` 的 `reject` / `dismiss` / `skip` suppress 都只压制下一帧；若当前 `document_revision` 与对应 recent-action 修订号之间已隔一帧，且该候选仍是本地规则层给出的高置信默认候选，则可恢复默认 `Tab`
- 当 `recent_actions` 中同时存在多条 ghost 反馈时，当前应以“当前 `candidate_ref` 的最近一条相关动作”作为 suppress / cooldown 判断基线：更早的同 candidate 动作不应覆盖更新动作，而其他 candidate 的更新动作也不应外溢影响当前候选
- 这条“最近一条相关动作优先”约束同样适用于恢复窗口：若同一 `candidate_ref` 的最新否定动作 cooldown 已过，则更早的同 candidate `reject` / `dismiss` / `skip` 不应继续把该候选压成 `manual_only`
- 当前样本基线已把这条约束补到更细的交错组合：若同一 candidate 先 `skip`、随后最新一帧又 `reject`，则应以最新 `reject` 为准继续保持 `manual_only`
- 当前样本基线也已补到对称的跨 candidate 恢复态：若 same-candidate `skip` cooldown 已过，而最新一帧 `reject` 针对的是其他 candidate，则该 other-candidate 反馈不得外溢误伤当前默认 `Tab`

当前仓库已将这条约束从 `eval` 样本推进到 `datasets/examples/` 基线：

- [radishflow-ghost-candidate-set-chain-feed-valve-flash-flash-outlets-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-flash-outlets-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-flash-outlets-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-stop-no-legal-outlet-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-stop-no-legal-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-stop-no-legal-outlet-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-name-conflict-no-tab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-reject-no-retab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-reject-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-reject-no-retab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-reject-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-reject-no-retab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-reject-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-dismiss-no-retab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-dismiss-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-skip-no-retab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-skip-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-skip-no-retab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-skip-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-skip-no-retab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-skip-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json)
- [radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-heater-outlet-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-heater-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-heater-outlet-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-reject-no-retab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-reject-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-reject-no-retab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-reject-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-reject-no-retab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-reject-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-dismiss-no-retab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-dismiss-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-skip-no-retab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-skip-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-skip-no-retab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-skip-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-skip-no-retab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-skip-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-name-conflict-no-tab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-stop-no-legal-outlet-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-stop-no-legal-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-stop-no-legal-outlet-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-cooler-outlet-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-cooler-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-cooler-outlet-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-reject-no-retab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-reject-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-reject-no-retab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-reject-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-reject-no-retab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-reject-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-dismiss-no-retab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-dismiss-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-skip-no-retab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-skip-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-skip-no-retab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-skip-no-retab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-skip-no-retab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-skip-no-retab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-stop-no-legal-outlet-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-stop-no-legal-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-stop-no-legal-outlet-001-debug-full.json)
- [radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json](../../datasets/examples/radishflow-ghost-candidate-set-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json)
- [radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001-debug-full.json](../../datasets/examples/radishflow-copilot-request-ghost-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001-debug-full.json)

该组示例用于固定 `Feed -> Valve -> FlashDrum` 连续搭建链里“阀后入口 ghost 刚被接受，下一步默认转向 `FlashDrum` 的标准 outlet 补全”的 pre-model handoff 与 request assembly 口径。
其中新增的 `stop-no-legal-outlet` 示例用于固定另一条同样重要的边界：即使连续搭建链已经发生，若本地规则层没有提供任何合法 outlet 候选，模型侧也应继续停在空建议边界。
而 `outlets-name-conflict-no-tab` 示例则固定第三条边界：即使候选已经存在，只要命名冲突或手动消歧标记仍在，本地规则层也不应把任一候选升级成默认 `Tab`。
而 `outlets-ranking-ambiguous-no-tab` 示例则进一步固定第四条边界：即使命名没有冲突，只要两个 outlet 候选的排序分差过小，也不应把任一候选升级成默认 `Tab`。
而 `outlet-reject-no-retab`、`outlet-dismiss-no-retab` 与 `outlet-skip-no-retab` 三组示例则进一步固定第五条边界：即使某个 outlet 候选本身仍然合法，只要它刚在当前链式步骤里被用户 `reject` / `dismiss` / `skip`，也都不应立即再次作为默认 `Tab` 强推；若仍展示该候选，也应退回 `manual_only`。
而 `feed-valve-flash-outlet-tab-after-reject-cooldown`、`feed-valve-flash-outlet-tab-after-dismiss-cooldown` 与 `feed-valve-flash-outlet-tab-after-skip-cooldown` 三组示例则把同一条一帧 cooldown 后可恢复 `Tab` 的时间窗口语义推进到第一模板 `Feed -> Valve -> FlashDrum`。
而 `feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss` 示例则进一步验证同一条恢复规则也适用于 `reject` 作为最新动作：当同一 candidate 同时存在更早的 `dismiss` 和更新一帧的 `reject` 时，恢复窗口仍只看最新的同 candidate `reject`；一旦最新 `reject` cooldown 已过，旧 `dismiss` 不得继续阻止默认 `Tab` 恢复。
而 `feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject` 示例则进一步固定更细一层恢复语义：当同一 candidate 同时存在更早的 `reject` 和更新一帧的 `dismiss` 时，恢复窗口仍只看最新的同 candidate 动作；一旦最新 `dismiss` cooldown 已过，旧 `reject` 不得继续阻止默认 `Tab` 恢复。
而 `feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss` 示例则进一步验证同一条恢复规则并不依赖 `dismiss` 作为最新动作：当同一 candidate 同时存在更早的 `dismiss` 和更新一帧的 `skip` 时，恢复窗口仍只看最新的同 candidate `skip`；一旦最新 `skip` cooldown 已过，旧 `dismiss` 不得继续阻止默认 `Tab` 恢复。
而 `feed-heater-flash-heater-outlet` 示例则验证这套链式 handoff 不只适用于 `Valve`，同样适用于 `Feed -> Heater -> FlashDrum` 这类第二模板。
而 `feed-valve-flash-alternate-candidate-tab-after-other-reject / dismiss / skip`、`feed-heater-flash-alternate-candidate-tab-after-other-reject / dismiss / skip` 与 `feed-cooler-flash-alternate-candidate-tab-after-other-reject / dismiss / skip` 三组对称示例则进一步固定更细一层边界：`reject / dismiss / skip` 的 suppress 信号都只压制同一 `candidate_ref`，若当前高置信候选已经切到另一条 candidate，则仍允许恢复默认 `Tab`。
而 `feed-heater-flash-outlet-tab-after-reject-cooldown`、`feed-heater-flash-outlet-tab-after-dismiss-cooldown` 与 `feed-heater-flash-outlet-tab-after-skip-cooldown` 三组示例则进一步固定 recent suppress 的最小时间窗口：同一 `candidate_ref` 在下一帧必须 suppress-Tab，但若与对应 recent-action 修订号之间已隔一帧，且它仍是高置信合法候选，则允许恢复默认 `Tab`。
而 `feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss` 示例则把同一条“恢复窗口只看最近一条相关动作”的约束推进到第二模板：若更新一帧的同 candidate `reject` cooldown 已过，则更早的同 candidate `dismiss` 不得继续阻止默认 `Tab` 恢复。
而 `feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject` 示例则把同一条“恢复窗口只看最近一条相关动作”的约束推进到第二模板：若更新一帧的同 candidate `dismiss` cooldown 已过，则更早的同 candidate `reject` 不得继续阻止默认 `Tab` 恢复。
而 `feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss` 示例则把同一条“恢复窗口只看最近一条相关动作”的约束继续推进到第二模板：若更新一帧的同 candidate `skip` cooldown 已过，则更早的同 candidate `dismiss` 不得继续阻止默认 `Tab` 恢复。
而 `feed-heater-flash-outlet-reject-no-retab` 示例则进一步验证第二模板对 `reject` 也沿用同一条 suppress-Tab 语义：同一候选刚被拒绝后，若仍展示，也只能退回 `manual_only`。
而 `feed-heater-flash-outlet-dismiss-no-retab` 与 `feed-heater-flash-outlet-skip-no-retab` 两组示例则进一步验证第二模板对 `dismiss` / `skip` 也沿用同一条 suppress-Tab 语义：同一候选刚被关闭或跳过后，若仍展示，也只能退回 `manual_only`。
而 `feed-heater-flash-outlet-name-conflict-no-tab` 示例则进一步验证第二模板也能稳定落到 `manual_only`，而不是只存在一条正向 `Tab` 路径。
而 `feed-heater-flash-stop-no-legal-outlet` 示例则进一步验证第二模板同样能在合法候选为空时稳定停住，不会因为 recent-actions 已存在就强行继续补下一跳。
而 `feed-heater-flash-outlet-ranking-ambiguous-no-tab` 示例则进一步验证第二模板的 `manual_only` 不只来自命名冲突，也可以来自两个合法 `FlashDrum inlet` 候选的分差过小。
而 `feed-cooler-flash-cooler-outlet`、`feed-cooler-flash-outlet-name-conflict-no-tab` 与 `feed-cooler-flash-stop-no-legal-outlet` 三组示例则进一步验证第三模板 `Feed -> Cooler -> FlashDrum` 同样具备 `Tab / manual_only / empty` 完整对照组。
而 `feed-cooler-flash-outlet-reject-no-retab` 示例则进一步验证第三模板对 `reject` 也沿用同一条 suppress-Tab 语义：同一候选刚被拒绝后，若仍展示，也只能退回 `manual_only`。
而 `feed-cooler-flash-outlet-dismiss-no-retab` 与 `feed-cooler-flash-outlet-skip-no-retab` 两组示例则进一步验证第三模板对 `dismiss` / `skip` 也沿用同一条 suppress-Tab 语义：同一候选刚被关闭或跳过后，若仍展示，也只能退回 `manual_only`。
而 `feed-cooler-flash-outlet-tab-after-reject-cooldown`、`feed-cooler-flash-outlet-tab-after-dismiss-cooldown` 与 `feed-cooler-flash-outlet-tab-after-skip-cooldown` 三组示例则把同一条一帧 cooldown 后可恢复 `Tab` 的时间窗口语义推进到第三模板 `Feed -> Cooler -> FlashDrum`。
而 `feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss` 示例则把同一条“恢复窗口只看最近一条相关动作”的约束继续推进到第三模板：若更新一帧的同 candidate `reject` cooldown 已过，则更早的同 candidate `dismiss` 不得继续阻止默认 `Tab` 恢复。
而 `feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject` 示例则把同一条“恢复窗口只看最近一条相关动作”的约束继续推进到第三模板：若更新一帧的同 candidate `dismiss` cooldown 已过，则更早的同 candidate `reject` 不得继续阻止默认 `Tab` 恢复。
而 `feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject` 示例则把同一条“恢复窗口只看最近一条相关动作”的约束继续推进到第三模板：若更新一帧的同 candidate `skip` cooldown 已过，则更早的同 candidate `reject` 不得继续阻止默认 `Tab` 恢复。
而 `feed-cooler-flash-outlet-ranking-ambiguous-no-tab` 示例则进一步验证第三模板的 `manual_only` 不只来自命名冲突，也可以来自两个合法 `FlashDrum inlet` 候选的分差过小。

当前 `datasets/eval/` 也已补对应回归样本：

- [suggest-ghost-completion-chain-feed-valve-flash-stop-no-legal-outlet-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-stop-no-legal-outlet-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlets-name-conflict-no-tab-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlets-ranking-ambiguous-no-tab-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-outlet-reject-no-retab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-reject-no-retab-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-dismiss-no-retab-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-outlet-skip-no-retab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-skip-no-retab-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-reject-cooldown-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-skip-cooldown-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-alternate-candidate-tab-after-other-reject-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [suggest-ghost-completion-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-valve-flash-alternate-candidate-tab-after-other-skip-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-heater-outlet-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-heater-outlet-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-alternate-candidate-tab-after-other-reject-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-alternate-candidate-tab-after-other-skip-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-reject-cooldown-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-skip-cooldown-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-tab-after-latest-skip-cooldown-with-older-dismiss-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-outlet-reject-no-retab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-reject-no-retab-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-dismiss-no-retab-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-outlet-skip-no-retab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-skip-no-retab-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-name-conflict-no-tab-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-stop-no-legal-outlet-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-stop-no-legal-outlet-001.json)
- [suggest-ghost-completion-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-heater-flash-outlet-ranking-ambiguous-no-tab-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-cooler-outlet-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-cooler-outlet-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-alternate-candidate-tab-after-other-reject-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-alternate-candidate-tab-after-other-dismiss-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-alternate-candidate-tab-after-other-skip-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-reject-cooldown-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-latest-reject-cooldown-with-older-dismiss-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-dismiss-cooldown-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-latest-dismiss-cooldown-with-older-reject-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-skip-cooldown-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-tab-after-latest-skip-cooldown-with-older-reject-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-outlet-reject-no-retab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-reject-no-retab-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-dismiss-no-retab-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-outlet-skip-no-retab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-skip-no-retab-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-name-conflict-no-tab-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-stop-no-legal-outlet-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-stop-no-legal-outlet-001.json)
- [suggest-ghost-completion-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json](../../datasets/eval/radishflow/suggest-ghost-completion-chain-feed-cooler-flash-outlet-ranking-ambiguous-no-tab-001.json)

它们用于把 `Feed -> Valve -> FlashDrum` 的“链式停住空建议”“链式继续但只能 manual-only”“同一候选刚被 reject / dismiss / skip 后不可立即 retab”，以及三条模板里的“一帧 cooldown 后可恢复 Tab”边界，从 pre-model handoff 再推进到 response-level regression。其中三条链式模板当前都已额外覆盖“排序分差不足导致 manual-only”的分叉态，并且也都已补齐“旧 candidate 的 reject / dismiss / skip 不应误伤新 candidate”的 cross-candidate 对称基线；同时三条模板都已补齐“同一 candidate 的 reject / dismiss / skip 都只 suppress 下一帧、隔一帧后可恢复 Tab”的 finer-grained baseline。最新一轮又补上了两组 stacked recent-actions 交错边界：一组固定“same-candidate dismiss cooldown 已过、但 latest other-candidate skip 不外溢时仍可恢复 `Tab`”，另一组固定“same-candidate 先 reject、后 skip 时以最新 skip 继续 no-retab”，避免 latest-action precedence 与恢复窗口只在单一动作组合上成立。

