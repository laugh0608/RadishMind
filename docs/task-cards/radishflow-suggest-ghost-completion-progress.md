# `RadishFlow` 任务卡附件：`suggest_ghost_completion` PoC 进展

更新时间：2026-05-16

本文档承接主任务卡中拆出的真实 capture、batch 治理与 PoC 进展细节。主任务卡只保留任务目标、输入输出契约、风险边界和评测口径。

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
- 第十六批 `v16` 则开始沿共享 sample-group 入口推进“尚未真实化”的高价值恢复边界：当前 `rfb-3106a3d75ca8` 已覆盖 cooler / heater / valve 三模板上的 latest-action precedence 变体，即 same-candidate 先 `reject` 后 `skip` 或先 `skip` 后 `reject` 时仍保持 `manual_only`，并首轮直接收口到 `audit=5/5 pass`
- 第十七批 `v17` 继续沿共享 sample-group 入口补 `remaining-other-candidate-recovery`：当前 `rfb-baabe9013c13` 已覆盖 cooler / heater / valve 三模板上 other-candidate 被 `reject / skip / dismiss` 后主候选仍可 `Tab` 恢复的 6 条样本，并首轮直接收口到 `audit=6/6 pass`
- 第十八批 `v18` 继续补 `remaining-basic-no-retab-and-cooldown`：当前 `rfb-e980c3c6f8e5` 已覆盖 cooler / heater / valve 三模板上的基础 `skip / reject / dismiss` no-retab 与 cooldown 后恢复 `Tab` 的 6 条样本，并首轮直接收口到 `audit=6/6 pass`
- 第十九批 `v19` 继续补 `remaining-foundation-and-conflict-basics`：当前 `rfb-93cdefa451d9` 已覆盖 cooler/heater 基础 outlet、heater 无合法出口停住、FlashDrum 标准双出口 split，以及双出口命名冲突 no-tab 的 6 条样本；其中首轮暴露的唯一失败已收口为 ghost canonicalization 过窄地丢弃了默认 `Tab` 主候选后的次级 `manual_only` 候选，当前已在 runtime 修正并基于现有 dump 重导收口到 `audit=6/6 pass`
- 第二十批 `v20` 继续补 `remaining-general-basics`：当前 `rfb-72e953f68142` 已覆盖 context-gap 空建议、FlashDrum inlet / liquid_outlet / nearby-node ranking、heater 命名提示，以及 mixer 标准 outlet 共 6 条通用基础样本，并首轮直接收口到 `audit=6/6 pass`
- 第二十一批 `v21` 则把 remaining groups 之后的首批 residual 样本池正式跑通：当前 `rfb-da969d6cb4e8` 已覆盖 `high-value-template-asymmetry-backfill` 这 6 条模板非对称高价值样本，包括 cooler 的 reject-no-retab 与 reject-cooldown+other-dismiss、heater 的 dismiss-no-retab 与 latest-skip cooldown precedence，以及 valve 的 ranking ambiguity 与基础 valve-outlet 场景，并首轮直接收口到 `audit=6/6 pass`
- 第二十二批 `v22` 继续沿 residual sample-group 主线补模板对称 suppress/cooldown 空边界：当前 `rfb-a066883057a7` 已覆盖 `high-value-suppression-cooldown-symmetry-backfill` 这 6 条样本，包括 cooler 的 dismiss-no-retab 与 dismiss-cooldown 恢复、heater 的 skip-no-retab 与 skip-cooldown 恢复，以及 valve 的 reject-no-retab 与 reject-cooldown 恢复，并首轮直接收口到 `audit=6/6 pass`
- 第二十三批 `v23` 继续把 residual 冲突与恢复边界补进正式真实池：当前 `rfb-b84f799e0ea3` 已覆盖 `high-value-residual-conflict-recovery-backfill` 这 6 条样本，包括 cooler / heater 的 name-conflict、no-legal-outlet、ranking-ambiguous 边界，以及 cooler reject cooldown 恢复与 valve skip no-retab 两条残余 suppress 边界，并首轮直接收口到 `audit=6/6 pass`
- 第二十四批 `v24` 继续补 residual cooldown 尾样：当前 `rfb-86d3e0d9f25f` 已覆盖 `high-value-residual-cooldown-tail-backfill` 这 6 条样本，包括 cooler latest reject cooldown、cooler skip cooldown + other reject、heater dismiss / reject cooldown、valve dismiss + other skip 与 valve skip cooldown，并首轮直接收口到 `audit=6/6 pass`
- 第二十五批 `v25` 则把最后 4 条未真实覆盖样本收口成 `high-value-residual-other-candidate-cooldown-tail`：当前 `rfb-56e4a3f16933` 已覆盖 heater dismiss+other skip、heater skip+other reject、valve reject+other dismiss 与 valve skip+other reject 四条 other-candidate cooldown tail 样本，并首轮直接收口到 `audit=4/4 pass`
- 截至 `2026-04-26`，本链正式真实样本池已从固定 trio 扩到 16 批高价值链式样本；按当前治理报表口径，`suggest_ghost_completion` 已达到 `real_captured=78/78`，并已完整接通 `manifest / audit / artifact summary / same-sample replay / cross-sample replay / real-derived negative` 正式治理链

当前这条 PoC 仍是轻量版：

- 目标仍是先证明本任务可以稳定做真实候选输出 capture 与正式导入，而不是一次性复制 `Radish docs QA` 的完整 batch 治理编排
- 下一步不再是补 simulated 样本，而是继续走正式入口跑下一批真实 teacher capture；只有当新增真实 batch 暴露出新的失败面且当前 runtime 修复不足以治理时，再回头补 recent-actions 或导入治理边界样本
- 当前在继续扩批时，还应同时观察两件事：其一是真实 provider 在批处理场景下是否仍会出现单样本卡顿，其二是 `manual_only` 多动作输出是否还会继续暴露新的结构坏法；若前者继续复现，应优先继续收口批次编排或重试/超时治理，而不是误判成新的 response malformed 模式
- 若后续某一轮再次出现三条默认样本同时 `HTTP 429` 且零 capture 的情况，当前应先归类为 provider 侧容量/限流阻塞；这类现象最多驱动 capture 脚本健壮性修复或 provider 切换，不应误记成新的 `suggest_ghost_completion` 输出坏法
- 若后续继续使用多 provider fallback 采集真实 batch，当前还应额外观察 provider 间的输出风格漂移，例如把本应是纯文本的 `summary` / `answer.text` 写成 JSON 字符串；这类现象若稳定复现，应优先在 runtime 做任务级窄修复，而不是把坏输出原样固化进正式批次
- 当前 formal real batch 治理层已不再缺最小 `artifact summary` 口径，也已接通首批 same-sample / cross-sample negative replay、两路 recommended replay summary，以及首批 real-derived negative pattern
- 这批 real-derived 当前先收口为 3 条 committed simulated negative，分别覆盖默认高置信 `Tab` 被错误降级、ambiguous no-tab 候选被错误升级成 `Tab`、以及空 `legal_candidate_completions` 下凭 `recent_actions` 主观补出 ghost action 三类稳定漂移
- 最近连续七批正式真实 capture 均首轮 `audit pass`，未再暴露新的 runtime 根因；因此当前兜底层可阶段性收口，不必继续围绕旧坏法深挖
- 因此本链下一轮 `M3` 推进不应回到 teacher capture 或重复补 replay，而应在 `78/78` 真实覆盖基线上继续观察是否需要新增非重复高价值样本；当前 `high-value-template-asymmetry-backfill`、`high-value-suppression-cooldown-symmetry-backfill`、`high-value-residual-conflict-recovery-backfill`、`high-value-residual-cooldown-tail-backfill` 与 `high-value-residual-other-candidate-cooldown-tail` 已正式跑通
- 为避免下一轮真实 capture 再回到人工临时挑样本，批次入口 [run-radishflow-ghost-real-batch.py](../../scripts/run-radishflow-ghost-real-batch.py) 当前已支持 `--sample-group` 并继续由共享真相源维护下一组正式入口；`v21` 到 `v25` 已证明这条“remaining groups 之后继续拆 residual sample-group”口径可以稳定落地
- 当前 `high-value-residual-other-candidate-cooldown-tail / v25` 已正式跑通；治理报表下一步不再指向旧 remaining 或 residual 入口，也不把继续扩样作为默认主线。后续只有服务/API 接入或编辑器集成暴露新的非重复交互假设时，才继续扩真实 capture

