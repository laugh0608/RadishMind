# `RadishFlow` 任务卡：`suggest_ghost_completion`

更新时间：2026-04-21

## 任务目标

基于 `RadishFlow` 当前画布中的单元放置、canonical ports、本地规则筛出的合法候选和邻近拓扑上下文，输出 1 到 3 条可供前端渲染为灰态 ghost 的补全建议。

当前任务关注的是“编辑器内 pending 建议”，不是正式 patch，也不是直接改图。所有输出都必须先经过本地规则系统过滤，并在用户接受后继续走本地命令与校验流程。

## 请求映射

- `project`: `radishflow`
- `task`: `suggest_ghost_completion`
- 请求结构遵循 [CopilotRequest schema](../../contracts/copilot-request.schema.json)

## 任务定位

本任务和 [`suggest_flowsheet_edits`](radishflow-suggest-flowsheet-edits.md) 不同：

- `suggest_flowsheet_edits` 面向正式候选编辑提案，默认会影响上层真相源，因此必须 `requires_confirmation: true`
- `suggest_ghost_completion` 面向前端灰态占位补全，接受前只是 pending ghost，不应被视为正式写回

因此本任务优先解决：

- 放置标准单元后，下一个最可能补全的 canonical port
- 哪个邻近节点或物流线最适合作为 ghost connection 候选
- 哪个 ghost stream name 最符合当前局部命名习惯

## 当前 PoC 进展

当前仓库已将本任务从“只有样本和契约”推进到“已完成多批真实 capture 正式导入”的最小工程骨架：

- 已补任务 prompt：[radishflow-suggest-ghost-completion-system.md](../../prompts/tasks/radishflow-suggest-ghost-completion-system.md)
- 已补最小 runtime：`services/runtime/inference.py` 与 [run-copilot-inference.py](../../scripts/run-copilot-inference.py) 现已支持 `radishflow / suggest_ghost_completion`
- 已补轻量批次入口：[run-radishflow-ghost-real-batch.py](../../scripts/run-radishflow-ghost-real-batch.py)，默认固定 3 个代表样本，覆盖 `Tab / manual_only / empty`
- 上述批次入口当前若未显式传 `--output-root`，会默认落到 `datasets/eval/candidate-records/radishflow/batches/YYYY-MM/<batch_key>/`，减少后续真实 batch 入仓时的人工路径拼接，并把长语义留在 `manifest` 元数据
- 上述批次入口当前也已从“单进程多请求 capture”收口为“逐样本单进程 capture + openai-compatible provider 单样本硬超时”，避免单条真实 provider 请求失控时把整批卡死
- `datasets/eval/radishflow-task-sample.schema.json` 当前已支持外部 `candidate_response_record`，因此真实或模拟 capture 可回灌到同一条 `manifest -> audit` 校验链
- 已补批次导入入口：[import-candidate-response-dump-batch.py](../../scripts/import-candidate-response-dump-batch.py)，可将一批 raw dump 重新归一化后正式导入仓库
- 单条 dump 导入入口 [import-candidate-response-dump.py](../../scripts/import-candidate-response-dump.py) 当前也已支持 `--recanonicalize-response`，用于处理 canonicalization 修复前采集的旧 dump
- 已补两条窄范围 malformed JSON 修复：针对 `radishflow / suggest_ghost_completion` 的稳定 provider 坏法，当前会在首次 `json.loads` 失败后尝试修复“多闭合一个 `}`”以及 `manual_only` 多动作输出中提前关掉 `proposed_actions` / `answers` 作用域的坏法，再继续走既有 canonicalizer，避免把接近完整的真实输出直接固化成 `MODEL_OUTPUT_NOT_JSON`

当前已正式入仓的真实 batch 统一位于 `datasets/eval/candidate-records/radishflow/batches/YYYY-MM/<batch_key>/` 短路径目录；批次内固定使用 `manifest.json`、`audit.json`、`artifacts.json`、`r/`、`o/` 与 `d/` 这些短结构名，而不是继续把长 `collection_batch` 和 sample slug 直接编码到物理路径。

前九批当前都只收口同一组 3 条 record：

- `suggest-ghost-completion-flash-vapor-outlet-001.record.json`
- `suggest-ghost-completion-valve-ambiguous-no-tab-001.record.json`
- `suggest-ghost-completion-chain-feed-valve-flash-stop-no-legal-outlet-001.record.json`

它们之所以应入仓，是因为当前同时满足：

- 覆盖默认 PoC 三条主路径：`Tab` / `manual_only` / `empty`
- 对应真实 raw dump 经当前 runtime 归一化后，可回放通过 `radishflow-ghost-completion` 回归
- 已生成正式 `manifest` 与 `audit`，不再依赖 `/tmp` 下的临时批次产物
- 第二批 `v3` 还额外暴露出一个真实 provider 失败面：两条非空样本返回的是“几乎完整但多闭合一个 `}` 的 JSON”；当前已以窄修复收口后重新导入，并恢复到 `audit=3/3 pass`
- 第三批 `v4` 未再暴露新的可重新归一化结构坏法；默认批次执行时 `manual_only` 样本一度出现 provider 卡顿，但拆成单样本复跑后可正常 capture，并最终仍沿同一条正式导入链收口到 `audit=3/3 pass`
- 第四批 `v5` 则继续把上面的执行层问题推进到可治理状态：一方面批次入口已改为逐样本单进程并加硬超时，另一方面 `manual_only` 样本新暴露出的“多动作 JSON 提前关掉 `proposed_actions` / `answers` 作用域”坏法也已被窄修复收口，并恢复到 `audit=3/3 pass`
- 紧接着发起的第五批尝试 `v6` 起初一度被 `openrouter` 的 provider-wide `HTTP 429` 阻塞，未能形成新 dump；但在补上 `openrouter / deepseek` 双 profile 切换后，当前已使用 `deepseek` fallback profile 重跑并正式入仓，确认这轮并未新增新的任务级 malformed JSON / 导入归一化失败面，而是供应商可用性阻塞已可绕过
- 第六批 `v7` 则继续暴露并收口了两条新的真实失败面：其一是当前 openrouter 默认 free 模型已被上游废弃，三条样本会直接 `404 deprecated`；其二是 `deepseek` 在 `empty` 样本上会把 `summary` 与 `empty_suggestion_reason.text` 写成内嵌 JSON 字符串。前者当前已通过 provider profile 切换与 `.env.example` 口径更新收口，后者则已通过只作用于 `radishflow / suggest_ghost_completion` 的摘要文本窄修复收口，并恢复到 `audit=3/3 pass`
- 第七批 `v8` 则继续证明：在坚持“先试 openrouter”且先轮换 openrouter 候选模型后，本轮观察到的新增阻塞仍主要是 provider/model 可用性层面的 `HTTP 404` 与 `HTTP 429`，而不是新的任务级输出坏法；当前已按既定策略切到 `deepseek` fallback profile 完成正式 capture，并确认三条样本均可回归通过、且 `v7` 暴露的摘要 JSON 字符串漂移未再次出现
- 第八批 `v9` 则继续把“先试 openrouter，再按情况切 provider”的边界推进到更细一层：本轮除了继续遇到默认 free 模型废弃、付费额度门槛与短窗口限流外，还观察到 `openrouter` 的 `nemotron-nano` 虽可完成 `Tab`/`empty` 样本，却在 `manual_only` 样本上退化成 project 名错误、空动作和错误 citation 结构的 schema-invalid JSON。当前这条坏法不属于适合通过 runtime 窄修复后正式导入的范围，因此只作为模型质量漂移观察项保留；正式入仓仍改由 `deepseek` fallback profile 完成，并确认当前未新增新的可治理导入/归一化失败面
- 作为仓库主线转入 `M3` 治理收口后的首个配套动作，批次入口 [run-radishflow-ghost-real-batch.py](../../scripts/run-radishflow-ghost-real-batch.py) 现也会默认写出 `<collection-batch>.artifacts.json`；对应最新正式批次 `2026-04-21-radishflow-ghost-poc-real-v10-chain-recovery-boundaries/` 已补最小 batch-level artifact summary，用于统一沉淀 `manifest / audit / output_root / records / responses / dumps` 的结构化治理摘要，而不再只有 `manifest + audit`
- 当前又为最新 clean formal batch `rfb-22d033e634d2` 补齐首批 3 条 committed same-sample negative 样本，并生成 `negative-replay-index.json` 与 `recommended-negative-replay.summary.json`；这批负例分别覆盖默认 `Tab` 被错误降级、ambiguous 候选被错误升级为 `Tab`、空 `legal_candidate_completions` 下主观补出 ghost action 三类稳定边界
- 在此基础上，本链又补齐首批 3 条 committed cross-sample negative 样本，并生成 `cross-sample-replay-index.json` 与 `cross-recommended.summary.json`；当前先把 cross-sample replay 收口为两类稳定错配组：一类是把 empty record 回放到需要默认 `ghost_completion` 的样本，另一类是把 ambiguous manual-only record 回放到必须保持空建议的样本
- 第十批 `v10` 则正式把真实 capture 从固定 trio 往外扩到首批高价值链式样本：当前 `rfb-0e8e36644e3a` 已覆盖 same-candidate cooldown 恢复、mixed-history 空建议边界与 other-candidate 不外溢恢复三类真实链式场景，并首轮直接收口到 `audit=3/3 pass`
- 第十一批 `v11` 继续沿这条主线扩下一组非重复边界：当前 `rfb-3a35f67997de` 已覆盖 alternate candidate 切换不误伤新的高置信默认候选、latest same-candidate reject 后保持 `manual_only` 而不误回 `Tab`、以及 cooler 模板上的 mixed-history 空建议边界，并首轮直接收口到 `audit=3/3 pass`
- 第十二批 `v12` 继续按模板分散扩下一组非重复边界：当前 `rfb-c230e09fde5a` 已覆盖 valve 模板上的 mixed-history 空建议、heater 模板上的 alternate candidate 切换不误伤，以及 cooler 模板上的 latest dismiss 后继续保持 `manual_only`，并首轮直接收口到 `audit=3/3 pass`
- 第十三批 `v13` 继续把 recent-action 变体推进到剩余模板：当前 `rfb-88afee19f042` 已覆盖 heater 与 valve 模板上的 latest dismiss 后继续保持 `manual_only`，以及 cooler 模板上的 alternate candidate 在 other dismiss 后继续作为默认候选，并首轮直接收口到 `audit=3/3 pass`
- 第十四批 `v14` 则把 cooldown 恢复语义正式推进到三模板：当前 `rfb-4bc5dfa547a5` 已覆盖 valve / heater / cooler 三模板上 latest dismiss cooldown 已过且 older reject 不再阻止 `Tab` 恢复的恢复形态，并首轮直接收口到 `audit=3/3 pass`
- 第十五批 `v15` 继续把其余恢复语义正式推进到剩余模板：当前 `rfb-6ed2ed2be612` 已覆盖 valve / heater 模板上的 latest reject cooldown 已过且 older dismiss 不再阻止 `Tab` 恢复，以及 cooler 模板上的 latest skip cooldown 已过恢复 `Tab`，并首轮直接收口到 `audit=3/3 pass`
- 截至 `2026-04-21`，本链正式真实样本池已从固定 trio 扩到 6 批高价值链式样本；按当前治理报表口径，`suggest_ghost_completion` 已达到 `real_captured=21/78`，并已完整接通 `manifest / audit / artifact summary / same-sample replay / cross-sample replay / real-derived negative` 正式治理链

当前这条 PoC 仍是轻量版：

- 目标仍是先证明本任务可以稳定做真实候选输出 capture 与正式导入，而不是一次性复制 `Radish docs QA` 的完整 batch 治理编排
- 下一步不再是补 simulated 样本，而是继续走正式入口跑下一批真实 teacher capture；只有当新增真实 batch 暴露出新的失败面且当前 runtime 修复不足以治理时，再回头补 recent-actions 或导入治理边界样本
- 当前在继续扩批时，还应同时观察两件事：其一是真实 provider 在批处理场景下是否仍会出现单样本卡顿，其二是 `manual_only` 多动作输出是否还会继续暴露新的结构坏法；若前者继续复现，应优先继续收口批次编排或重试/超时治理，而不是误判成新的 response malformed 模式
- 若后续某一轮再次出现三条默认样本同时 `HTTP 429` 且零 capture 的情况，当前应先归类为 provider 侧容量/限流阻塞；这类现象最多驱动 capture 脚本健壮性修复或 provider 切换，不应误记成新的 `suggest_ghost_completion` 输出坏法
- 若后续继续使用多 provider fallback 采集真实 batch，当前还应额外观察 provider 间的输出风格漂移，例如把本应是纯文本的 `summary` / `answer.text` 写成 JSON 字符串；这类现象若稳定复现，应优先在 runtime 做任务级窄修复，而不是把坏输出原样固化进正式批次
- 当前 formal real batch 治理层已不再缺最小 `artifact summary` 口径，也已接通首批 same-sample / cross-sample negative replay、两路 recommended replay summary，以及首批 real-derived negative pattern
- 这批 real-derived 当前先收口为 3 条 committed simulated negative，分别覆盖默认高置信 `Tab` 被错误降级、ambiguous no-tab 候选被错误升级成 `Tab`、以及空 `legal_candidate_completions` 下凭 `recent_actions` 主观补出 ghost action 三类稳定漂移
- 最近连续六批正式真实 capture 均首轮 `audit pass`，未再暴露新的 runtime 根因；因此当前兜底层可阶段性收口，不必继续围绕旧坏法深挖
- 因此本链下一轮 `M3` 推进不应回到 teacher capture 或重复补 replay，而应继续沿固定 trio 之外的高价值链式样本扩真实 capture，优先转向尚未真实化的高价值恢复边界，而不是继续重复已完成的 cooldown 对称组

## 最小必需输入

请求至少应包含以下内容：

- `artifacts` 中存在主输入 `flowsheet_document`
- `context.document_revision`
- `context.selected_unit_ids`，且当前任务应只聚焦一个被放置或当前激活的单元
- `context.unconnected_ports` 或 `context.missing_canonical_ports`
- `context.legal_candidate_completions`

若没有本地规则层提供的合法候选集，本任务应优先返回空建议，而不是凭主观猜测拼接拓扑。

## 推荐补充输入

- `context.selected_unit`
- `context.nearby_nodes`
- `context.cursor_context`
- `context.naming_hints`
- `context.topology_pattern_hints`

推荐先由本地规则层把“合法候选空间”压缩出来，再由模型做排序、命名和空结果判断。

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

## 禁止透传

以下内容不应进入本任务请求：

- 任何安全凭据、token、cookie 或本机秘密引用
- 绕过本地规则系统才能成立的候选连接
- 未经裁剪的整页无关状态和与当前单元无关的大范围拓扑噪声

## 第一阶段优先支持

当前阶段优先覆盖标准化程度高的单元：

- `FlashDrum`
- `Mixer`
- `Valve`
- `Heater`
- `Cooler`
- `Feed`

其中第一优先是 `FlashDrum`：

1. `inlet`
2. `vapor_outlet`
3. `liquid_outlet`

## 输出要求

响应结构遵循 [CopilotResponse schema](../../contracts/copilot-response.schema.json)。

本任务对字段的要求如下：

- `summary` 必须说明当前是否存在高置信 ghost 建议，以及建议围绕哪个单元生成
- `proposed_actions` 可以为空；若存在建议，数量应限制在 1 到 3 条
- 每条建议都必须使用 `kind: "ghost_completion"`
- 每条 `ghost_completion` 都必须包含 `title`、`target`、`rationale`、`patch`、`preview`、`apply`、`risk_level`、`requires_confirmation`
- `citations` 必须能回溯到当前单元、缺失端口、本地候选集或邻近拓扑证据

## `ghost_completion` 语义

`ghost_completion` 不是直接 patch，而是可供前端渲染的待接受占位建议。

建议使用以下结构表达：

- `target`
  - 指向 unit、port 或局部 candidate slot
- `patch`
  - 只描述 ghost 语义，不描述正式文档写回
  - 推荐包含 `ghost_kind`、`target_port_key`、`candidate_ref`、`ghost_stream_name`
- `preview`
  - 前端渲染提示
  - 推荐包含 `ghost_color`、`accept_key`、`render_priority`
- `apply`
  - 用户接受 ghost 后交给本地命令层执行的命令描述
  - 推荐包含 `command_kind` 与 `payload`

## 风险与确认规则

本任务默认只输出低副作用 ghost 建议。

- `low`
  - 标准 canonical port 补全
  - 标准 ghost connection
  - 低歧义命名补全
- `medium`
  - 多候选接近、但仍能给出一条领先建议的局部补全
- `high`
  - 当前阶段原则上不应输出；若候选具有明显拓扑副作用或规则冲突风险，应直接返回空建议

本任务默认 `requires_confirmation: false`，因为：

- 建议本身只是 pending ghost
- 用户接受后仍要经过本地命令系统与本地校验
- 这一步不是绕过规则的直接写回

若某条建议只有在越过本地规则前提下才能成立，则不应输出。

## `Tab` 接受规则

本任务不应默认把“第一条建议”直接等同于“可按 Tab 接受”。

建议至少满足以下条件之一，前端才将第一条视为默认 `Tab` 候选：

- 本地规则层已将其标记为唯一高置信合法候选
- 或模型排序后显著领先第二候选，且无本地 conflict 标记

当前推荐将“可默认 Tab”尽量显式编码在候选里：

- `is_tab_default = true`
- `is_high_confidence = true`
- `conflict_flags = []`

若不存在高置信候选，应返回空建议或仅返回可见 ghost，但不应强推默认接受。

## 正反例边界

正例：

- `FlashDrum` 刚放置后，针对缺失的 `vapor_outlet` 返回 ghost connection 建议
- `Mixer` 已有两个入口候选时，只返回本地规则认可的那个标准出口补全
- 在合法候选集中为新 ghost stream 生成符合当前命名习惯的补全名

反例：

- 直接声称“已为你创建并落盘”
- 忽略本地 canonical ports，自行发明一个不存在的 port
- 明明上下文不足，还输出看起来聪明但不可验证的连线猜测
- 把高副作用拓扑调整伪装成低风险 ghost 建议

## 最小输入快照示例

```json
{
  "project": "radishflow",
  "task": "suggest_ghost_completion",
  "artifacts": [
    {
      "kind": "json",
      "role": "primary",
      "name": "flowsheet_document",
      "mime_type": "application/json",
      "uri": "artifact://flowsheet/current"
    }
  ],
  "context": {
    "document_revision": 31,
    "selected_unit_ids": ["U-12"],
    "selected_unit": {
      "id": "U-12",
      "kind": "FlashDrum"
    },
    "unconnected_ports": [
      "inlet",
      "vapor_outlet",
      "liquid_outlet"
    ],
    "legal_candidate_completions": [
      {
        "candidate_ref": "cand-vapor-stub",
        "ghost_kind": "ghost_connection",
        "target_port_key": "vapor_outlet"
      },
      {
        "candidate_ref": "cand-liquid-stub",
        "ghost_kind": "ghost_connection",
        "target_port_key": "liquid_outlet"
      }
    ]
  }
}
```

## 候选输出片段示例

```json
{
  "kind": "ghost_completion",
  "title": "补全 FlashDrum 的 vapor outlet ghost 连线",
  "target": {
    "type": "unit_port",
    "unit_id": "U-12",
    "port_key": "vapor_outlet"
  },
  "rationale": "FlashDrum 的 canonical ports 中 vapor outlet 尚未连接，且本地规则已给出合法 ghost connection 候选。",
  "patch": {
    "ghost_kind": "ghost_connection",
    "candidate_ref": "cand-vapor-stub",
    "target_port_key": "vapor_outlet",
    "ghost_stream_name": "V-12"
  },
  "preview": {
    "ghost_color": "gray",
    "accept_key": "Tab",
    "render_priority": 1
  },
  "apply": {
    "command_kind": "accept_ghost_completion",
    "payload": {
      "candidate_ref": "cand-vapor-stub"
    }
  },
  "risk_level": "low",
  "requires_confirmation": false
}
```

## 评测口径

当前阶段至少检查以下指标：

- 结构合法率：`ghost_completion` 必须通过 schema 校验
- 本地合法率：建议必须来自本地规则认可的候选空间
- `Tab` 命中率：第一条默认建议被接受后应尽量命中用户真实下一步
- 空建议正确率：信息不足时，应稳定返回空建议而不是硬猜
- 建议后校验通过率：用户接受 ghost 后，本地命令与校验的通过率应足够高

## 非目标

当前任务不负责：

- 直接修改 `FlowsheetDocument`
- 绕过本地 canonical ports、连接校验或命令系统
- 把复杂多步搭建一次性扩成自治代理
- 替代 `suggest_flowsheet_edits` 去生成正式 patch
