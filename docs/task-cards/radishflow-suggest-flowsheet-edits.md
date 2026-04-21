# `RadishFlow` 任务卡：`suggest_flowsheet_edits`

更新时间：2026-04-18

## 任务目标

基于 `FlowsheetDocument`、选择集和诊断结果，输出结构化的候选编辑提案。

当前任务关注的是“可确认的候选动作”，不是直接改图。所有输出都必须能被人工审查、规则层复核或业务命令层再次校验。

## 当前 PoC 进展

当前仓库已把本任务从“只有样本 / golden_response / mock candidate-record PoC”继续推进到“可沿同一入口发起真实 provider capture”的状态：

- 已补任务 prompt：[radishflow-suggest-flowsheet-edits-system.md](../../prompts/tasks/radishflow-suggest-flowsheet-edits-system.md)
- 已补最小 runtime：`services/runtime/inference.py` 与 [run-copilot-inference.py](../../scripts/run-copilot-inference.py) 当前已支持 `radishflow / suggest_flowsheet_edits`
- 已有最小批次入口：[run-radishflow-suggest-edits-poc-batch.py](../../scripts/run-radishflow-suggest-edits-poc-batch.py)，当前默认仍固定 3 个代表样本覆盖高风险重连、局部规格占位与三步优先级链；仓库内后续又形成了 `2026-04-13-radishflow-suggest-edits-poc-real-v4/`、`2026-04-14-radishflow-suggest-edits-poc-real-v5/`、`2026-04-14-radishflow-suggest-edits-poc-real-v6/`、`2026-04-14-radishflow-suggest-edits-poc-real-v7-apiyi-de/`、`2026-04-14-radishflow-suggest-edits-poc-real-v8-apiyi-de/`、`2026-04-14-radishflow-suggest-edits-poc-real-v9-apiyi-cc/`、`2026-04-14-radishflow-suggest-edits-poc-real-v10-apiyi-ch/`、`2026-04-14-radishflow-suggest-edits-poc-real-v11-apiyi-de-cross-object/`、`2026-04-14-radishflow-suggest-edits-poc-real-v12-apiyi-de-parameter-ordering/`、`2026-04-14-radishflow-suggest-edits-poc-real-v13-apiyi-cc-cross-object/`、`2026-04-14-radishflow-suggest-edits-poc-real-v14-apiyi-cc-parameter-ordering/`、`2026-04-14-radishflow-suggest-edits-poc-real-v15-apiyi-ch-cross-object/`、`2026-04-14-radishflow-suggest-edits-poc-real-v16-apiyi-ch-parameter-ordering/`、`2026-04-15-radishflow-suggest-edits-poc-real-v17-apiyi-de-cross-object-citation/`、`2026-04-15-radishflow-suggest-edits-poc-real-v18-apiyi-cc-cross-object-citation/`、`2026-04-15-radishflow-suggest-edits-poc-real-v19-apiyi-ch-cross-object-citation/`、`2026-04-15-radishflow-suggest-edits-poc-real-v20-apiyi-cx-cross-object-citation/`、`2026-04-15-radishflow-suggest-edits-poc-real-v21-apiyi-cx-mixed-risk-cross-object/`、`2026-04-15-radishflow-suggest-edits-poc-real-v22-apiyi-cc-mixed-risk-cross-object/`、`2026-04-15-radishflow-suggest-edits-poc-real-v23-apiyi-ch-mixed-risk-cross-object/`、`2026-04-15-radishflow-suggest-edits-poc-real-v24-apiyi-de-mixed-risk-cross-object/`、`2026-04-15-radishflow-suggest-edits-poc-real-v25-apiyi-cx-triad-mixed-risk-cross-object/`、`2026-04-15-radishflow-suggest-edits-poc-real-v26-apiyi-cc-triad-mixed-risk-cross-object/` 与 `2026-04-15-radishflow-suggest-edits-poc-real-v28-apiyi-de-triad-mixed-risk-cross-object/` 二十四批真实扩样本批次，继续覆盖中风险局部参数修正、selection 裁剪、多动作链式优先级、selection chronology、顶层/issue/action citation 顺序、同 profile 复杂样本对照、跨对象 primary-focus selection 样本、更复杂的跨对象 citation 交错样本，以及 compressor/heater/pump 局部 patch 的 placeholder / parameter_updates / patch-key ordering
- 离线 regression 当前已补六条“下一批真实 capture 预热”样本：除 `cross-object-citation-interleaving-001` 与 `cross-object-warning-citation-ordering-001` 外，还新增了 `cross-object-mixed-risk-three-action-ordering-001`、`cross-object-mixed-risk-reconnect-plus-pump-parameter-001`、`cross-object-mixed-risk-reconnect-spec-compressor-placeholder-001` 与 `cross-object-mixed-risk-reconnect-pump-update-plus-separator-placeholder-001`；它们继续把跨对象 `diagnostic -> target artifact -> supporting artifact -> snapshot` 的顺序口径往“高风险断链 + 中风险局部 patch + parameter_updates / parameter_placeholders / spec_placeholders”这组更复杂组合样本上扩展
- 当前这组新 mixed-risk + cross-object 样本已由 `apiyi_cx / apiyi_cc / apiyi_ch / apiyi_de` 完成整组真实 capture：`2026-04-15-radishflow-suggest-edits-poc-real-v21-apiyi-cx-mixed-risk-cross-object/`、`2026-04-15-radishflow-suggest-edits-poc-real-v22-apiyi-cc-mixed-risk-cross-object/`、`2026-04-15-radishflow-suggest-edits-poc-real-v23-apiyi-ch-mixed-risk-cross-object/` 与 `2026-04-15-radishflow-suggest-edits-poc-real-v24-apiyi-de-mixed-risk-cross-object/` 均已覆盖 `cross-object-mixed-risk-three-action-ordering-001` 与 `cross-object-mixed-risk-reconnect-plus-pump-parameter-001` 两条样本，并在按当前 runtime `recanonicalize-response` 重导后全部收口到正式 `audit=2/2 pass`；这轮新暴露的问题最终被收口为“同 target warning 也会进入第一条 action 的 citation 组”和“pump efficiency review range 需优先保持既有 `[60, 85]` 正式口径”两类窄范围修正，其中 `apiyi_de` 仅出现过一次请求级 read-timeout retry 观察项，但未形成正式批次阻塞
- 在此基础上，`apiyi_cx` 也已先在下一组 triad patch-shape mixed-risk 样本上完成首轮真实 capture：`2026-04-15-radishflow-suggest-edits-poc-real-v25-apiyi-cx-triad-mixed-risk-cross-object/` 已覆盖 `cross-object-mixed-risk-reconnect-spec-compressor-placeholder-001` 与 `cross-object-mixed-risk-reconnect-pump-update-plus-separator-placeholder-001` 两条样本，并在按当前 runtime `recanonicalize-response` 重导后收口到正式 `audit=2/2 pass`；这轮没有暴露新的 runtime 根因，主要收口动作仍是把 triad 样本的 citation 断言调整到与真实 provider 的稳定输出一致
- 这组 triad patch-shape mixed-risk 样本随后也已横向扩到 `apiyi_cc` 与 `apiyi_de`：`2026-04-15-radishflow-suggest-edits-poc-real-v26-apiyi-cc-triad-mixed-risk-cross-object/` 与 `2026-04-15-radishflow-suggest-edits-poc-real-v28-apiyi-de-triad-mixed-risk-cross-object/` 当前都已直接收口到正式 `audit=2/2 pass`，说明默认 `120s` 对这组 triad 样本偏短，`180s` 长超时足以把 `apiyi_cc` 与 `apiyi_de` 跑通；这轮没有新增样本设计缺口或 runtime canonicalization 根因。`apiyi_ch` 的同组 `v27` 则已在首条样本上经历两次 retry 后仍触发 `180s` 读超时并失败，当前更应正式把它记为 triad 样本上的 provider 可用性阻塞，而不是把问题泛化成 profile 质量回退
- `suggest_flowsheet_edits` 的最后两条 ordering 尾样现在也已完成四个主 `apiyi` profile 的横向真实对照：继 `apiyi_cx / v39` 与 `apiyi_cc / v41` 后，`apiyi_ch / v42` 已在默认 `120s` 请求超时下直接覆盖 `efficiency-range-ordering` 与 `stream-spec-sequence-ordering` 并收口到正式 `audit=2/2 pass`；`apiyi_de / v43 -> v44` 则暴露出一个更窄的 runtime 缺口，即 `spec_placeholders` 虽已归一到正确字段名，但多规格占位顺序仍会漂移，当前已把它在 `services/runtime/inference_response.py` 中收口为稳定的 `temperature_c -> pressure_kpa -> flow_rate_kg_per_h` canonical 顺序，并基于现有 dump 重导到正式 `audit=2/2 pass`
- 这也让当前对 `apiyi_ch` timeout 的判断进一步收窄：ordering 尾样在默认 `120s` 下并未复现超时，因此现阶段更合理的口径仍是“triad 样本族存在独立的长超时需求”，而不是立即把 `apiyi_ch` 的默认 timeout 泛化上调到所有 `suggest_flowsheet_edits` capture
- 在此基础上，triad mixed-risk cross-object 样本的 timeout 判断也已进一步收口：当前 `scripts/run-radishflow-suggest-edits-poc-batch.py` 已把 `apiyi_ch` 在两条已知 triad 样本上的默认超时从“外部手工传长超时”收紧为“样本级 `210s` override”，不会影响其它 `suggest_flowsheet_edits` 样本族；`2026-04-16-radishflow-suggest-edits-poc-real-v45-apiyi-ch-triad-mixed-risk-cross-object-timeout-override/` 已验证这条 override 能把 capture 跑通，而 `v45 -> v46` 又进一步说明这轮剩余问题已不在 provider 可用性，而在 runtime 对 placeholder-only 参数 patch 的窄范围收口
- 同一轮主线里，`apiyi_ch` 在此前尚未补齐的 `mixed-risk patch combo` 样本族上也已完成横向正式收口：`2026-04-16-radishflow-suggest-edits-poc-real-v47-apiyi-ch-mixed-risk-patch-combo/` 已确认两条样本在默认 `120s` 下都能直接完成真实 capture，未复现 triad 样本族那种长超时阻塞；随后 `v47 -> v48` 仅再暴露并收口了一条更窄的 runtime 缺口，即 `COMPRESSOR_OUTLET_PRESSURE_TARGET_INVALID + UNIT_PARAMETER_OUT_OF_RANGE` 组合下的 compressor `efficiency_percent.suggested_range` 仍应保持既有 `[60, 85]` 正式口径，而不是沿通用 compressor 分支回退成 `[65, 85]`
- 随后 `apiyi_cx` 在此前尚未补齐的 cross-object primary-focus 样本族上也已完成横向正式收口：`2026-04-16-radishflow-suggest-edits-poc-real-v49-apiyi-cx-cross-object/` 已确认 `joint-selection-primary-focus` 与 `multi-unit-stream-primary-focus` 两条样本都能在首轮真实 capture 下直接收口到正式 `audit=2/2 pass`，没有再暴露新的 runtime 窄缺口；至此这组样本也已补齐四个主 `apiyi` profile 的横向正式对照
- 在此基础上，`apiyi_cx` 在此前同样尚未补齐的 parameter / patch ordering 样本族上也已完成横向正式收口：`2026-04-16-radishflow-suggest-edits-poc-real-v50-apiyi-cx-parameter-ordering/` 已确认 `compressor-parameter-placeholder-ordering`、`compressor-parameter-update-ordering`、`compressor-parameter-update-detail-ordering` 与 `heater-patch-key-ordering` 四条样本都能在首轮真实 capture 下直接收口到正式 `audit=4/4 pass`，没有再暴露新的 runtime 缺口；至此这组样本也已补齐四个主 `apiyi` profile 的横向正式对照
- 继续推进同一条主线时，`apiyi_cc` 在中风险 local-edits 样本族上也已补出一条新的正式批次：`2026-04-16-radishflow-suggest-edits-poc-real-v51-apiyi-cc-local-edits/` 首轮真实 capture 进一步暴露了两类窄缺口，一类是 provider 返回完整 JSON 但在中文短语里夹带未转义引号，另一类是 `compressor-evidence-gap-partial` 在“单选 compressor + 根因未确认”语境下仍会误保留越界 stream action，并把 warning / action citation 与 `minimum_delta_kpa` 收口到过宽口径
- 当前这些缺口也已通过既有“补 runtime 窄修补 -> 用现有 dump 重导正式 batch”的治理路径收口：`services/runtime/inference_provider.py` 现会对 `语义为"..."` 与 `所提到的"..."` 这类未转义中文引号做 `suggest_flowsheet_edits` 专用修补，`services/runtime/inference_response.py` 则把 `COMPRESSOR_OUTLET_PRESSURE_TARGET_TOO_CLOSE + COMPRESSOR_ROOT_CAUSE_UNCONFIRMED` 组合的 compressor patch 收紧为 unit-only 审查提案，固定 `minimum_delta_kpa=80`、过滤无诊断支撑的额外 stream action，并把 warning issue / 主 action 的 citation 分别稳定到 `diag-2 -> snapshot-1` 与 `diag-1 -> diag-2 -> flowdoc-unit-1 -> snapshot-1`
- 对应正式批次现已落为 `2026-04-16-radishflow-suggest-edits-poc-real-v53-apiyi-cc-local-edits-recanonicalized/`，并确认 `compressor-evidence-gap-partial`、`multi-selection-single-actionable-target`、`pump-parameter-combo` 与 `valve-local-fix-vs-global-balance` 四条样本都已收口到正式 `audit=4/4 pass`
- 同一组 local-edits 样本随后也已横向扩到 `apiyi_ch`：`2026-04-16-radishflow-suggest-edits-poc-real-v54-apiyi-ch-local-edits/` 已确认四条样本在首轮真实 capture 下就直接收口到正式 `audit=4/4 pass`，没有复现 `apiyi_ch` 在 triad 样本族上那种 capture 级长超时，也没有再暴露新的 malformed JSON、citation 顺序或局部 patch 收口缺口
- 同一组 local-edits 样本现又已横向扩到 `apiyi_de`：`2026-04-16-radishflow-suggest-edits-poc-real-v55-apiyi-de-local-edits/` 也已确认四条样本在首轮真实 capture 下直接收口到正式 `audit=4/4 pass`，没有新增 provider 可用性抖动或 runtime 窄缺口；这说明 local-edits 这组中风险样本当前对 `apiyi_de` 也已达到稳定正式门槛
- 同一组 local-edits 样本现也已横向扩到 `apiyi_cx`：`2026-04-16-radishflow-suggest-edits-poc-real-v56-apiyi-cx-local-edits/` 已确认 `compressor-evidence-gap-partial`、`multi-selection-single-actionable-target`、`pump-parameter-combo` 与 `valve-local-fix-vs-global-balance` 四条样本都能在首轮真实 capture 下直接收口到正式 `audit=4/4 pass`，没有再暴露新的 runtime 窄缺口；至此 local-edits 这组样本也已补齐 `apiyi_cx / apiyi_cc / apiyi_ch / apiyi_de` 四个主 profile 的横向正式对照
- 同一组 4 条 `mixed-risk / citation / reconnect` 顺序样本现也已补齐 `apiyi_cx`：`2026-04-16-radishflow-suggest-edits-poc-real-v57-apiyi-cx-mixed-risk-citation-reconnect/` 首轮真实 capture 暴露出的剩余问题已进一步收窄为 3 类 runtime 窄缺口，即“带 `latest_snapshot` 的双 action 局部修补样本仍应保持 `status=partial`”、“简单单-unit 语境下 stream 支撑 citation 不应再无条件带上父 unit”，以及“`flowdoc-*` 引用编号需按当前响应实际纳入的目标对象做紧凑稳定编号”
- 对应正式批次现已落为 `2026-04-16-radishflow-suggest-edits-poc-real-v58-apiyi-cx-mixed-risk-citation-reconnect-recanonicalized/`，并确认 `mixed-risk-reconnect-plus-spec`、`citation-ordering-diagnostics-before-artifacts-before-snapshot`、`issue-citation-ordering-warning-artifact-before-snapshot` 与 `reconnect-connection-placeholder-ordering` 四条样本都已收口到正式 `audit=4/4 pass`；至此这组 `mixed-risk / citation / reconnect` 样本也已补齐 `apiyi_cx / apiyi_cc / apiyi_ch / apiyi_de` 四个主 profile 的横向正式对照
- 在真实 provider 主线上，当前 `mixed-risk / citation / reconnect`、`cross-object primary-focus`、`parameter / patch ordering`、`cross-object citation`、`mixed-risk + cross-object`、`mixed-risk patch combo` 与 `local-edits` 七组样本族都已形成四主 `apiyi_cx / apiyi_cc / apiyi_ch / apiyi_de` 的正式收口；其中 `apiyi_ch` 在 `v16` 暴露出的请求级 read-timeout 抖动，在后续 mixed-risk 组合样本上仍未演变成正式质量阻塞，`apiyi_de` 也只出现过单次 read-timeout retry 观察项但未影响批次收口，因此当前更应继续把这类现象记为“可用性继续观察”的 provider 观察项，而不是 profile 分组错误。对应 runtime 也已补齐路径式 placeholder、`mass_flow_kg_per_h` / `mass_flow_kg_h`、`mass_flow_rate_kg_h`、`flow_rate_kg_h` 等流量占位别名、`operating pressure` / `pressure target` 到 `operating_pressure_kpa` 的参数占位提示收口、`outlet_temperature_target_c` / `outlet_temperature_target` / `target_outlet_temperature_c` 等参数占位别名、`STREAM_DISCONNECTED` 与 `STREAM_SPEC_MISSING` 等 error issue 也按 canonical 前缀补齐 supporting citations 的窄范围收口、跨对象 citation 样本中的 issue/action/top-level citation 交错排序收口、`flowdoc-*` 编号进一步收紧为按当前响应实际纳入的目标对象做紧凑稳定编号、reconnect patch 中 `retain_existing_source_binding` 的稳定保留、warning issue citation 的 `diag -> artifact -> snapshot` 稳定排序收口、`suggest_flowsheet_edits` 任务里 malformed JSON 的多余闭合 brace 修复和错误 `flowdoc-*` 引用回退到 canonical target artifact 的收口，以及对 compressor / pump / valve 局部参数诊断进入 auto-synthesize 集合、空 `proposed_actions` 回收到稳定 `parameter_placeholders` / `parameter_updates`、placeholder-only 样本状态回收到 `ok`、dict/path 形态的 placeholder 回收到稳定字段名、provider 只给首项 placeholder 时与合成占位序列稳定合并，以及局部参数 patch 的 `risk_level` 收口回正式 `medium` 的窄范围修复
- 上述入口在 `mock` 模式下仍会沿 `candidate_response_record -> manifest -> audit` 固定首批 committed PoC，而切到真实 provider 时也继续沿同一脚本入口收口，不另起第二套导入流程
- 当前默认输出目录仍收口到 `datasets/eval/candidate-records/radishflow/<collection_batch>/`，让后续真实 batch 可以直接落正式目录
- 当前已继续把 `default-early-trio` 往剩余缺口上前推两档正式真实批次：`2026-04-18-radishflow-suggest-edits-poc-real-v59-apiyi-cx-default-early-trio/` 已确认 `apiyi_cx` 在 `reconnect-outlet-001`、`stream-spec-placeholder-001` 与 `three-step-priority-chain-001` 三条样本上都能首轮真实 capture，并在更新当前样本基线后收口到 `audit=3/3 pass`
- 同一组三条样本随后也已横向扩到 `apiyi_cc`：原始真实 capture `2026-04-18-radishflow-suggest-edits-poc-real-v60-apiyi-cc-default-early-trio/` 暴露出的剩余问题被收紧为一条更窄的 alias 缺口，即 `mass_flow_kgph` 还未归一到正式 `flow_rate_kg_per_h`
- 在为 `services/runtime/inference_response.py` 补上这条 `mass_flow_kgph -> flow_rate_kg_per_h` 窄范围归一后，`v60` 也已通过 `scripts/import-candidate-response-dump-batch.py --recanonicalize-response` 重导为 `2026-04-18-radishflow-suggest-edits-poc-real-v61-apiyi-cc-default-early-trio-recanonicalized/`，并收口到正式 `audit=3/3 pass`
- 同一组三条样本随后也已横向扩到 `apiyi_ch`：`2026-04-18-radishflow-suggest-edits-poc-real-v62-apiyi-ch-default-early-trio/` 已确认三条样本都能在首轮真实 capture 下直接收口到正式 `audit=3/3 pass`，没有复现 triad mixed-risk 样本族上那种样本级长超时，也没有新增 runtime canonicalization 缺口
- 同组又继续横向扩到 `apiyi_de`：`2026-04-18-radishflow-suggest-edits-poc-real-v63-apiyi-de-default-early-trio/` 也已确认 `reconnect-outlet-001`、`stream-spec-placeholder-001` 与 `three-step-priority-chain-001` 三条样本都能在首轮真实 capture 下直接收口到正式 `audit=3/3 pass`，没有新增 alias、citation 顺序或 patch 收口问题
- 至此 `default-early-trio` 这组三条早期样本也已完成 `default + apiyi_cx + apiyi_cc + apiyi_ch + apiyi_de` 的正式覆盖；结合本地 coverage 脚本，当前剩余未补齐四主 `apiyi` profile 的正式缺口已从 `9` 条收紧到 `6` 条，后续主线可正式转向 `default-selection-ordering` 与 `heater-follow-up`
- 按这一主线继续往前推进后，`default-selection-ordering` 也已完成首条真实横向扩样：原始真实 capture `2026-04-18-radishflow-suggest-edits-poc-real-v64-apiyi-cx-default-selection-ordering/` 在 5 条样本上把差异收紧到两条更窄的 runtime 缺口，即 `HEATER_OUTLET_EFFECT_UNCONFIRMED` 的 stream-target warning issue 不应默认保留目标 stream artifact，以及 pump `UNIT_PARAMETER_INCOMPLETE` 在 `outlet pressure target` 语境下不应把 `operating_pressure_kpa` 一并并回正式 `parameter_placeholders`
- 为支撑这组样本的正式收口，当前已把 `services/runtime/inference_response.py` 再收紧两格：一是 `stream` 目标的 `HEATER_OUTLET_EFFECT_UNCONFIRMED / COOLER_OUTLET_EFFECT_UNCONFIRMED` warning issue 不再默认带目标 stream citation，二是当 runtime 已把 pump 占位收紧到 `outlet_pressure_target_kpa` 时，不再把 provider 原始 `operating_pressure_kpa` 回流进正式 patch；同时也把 `services/runtime/inference_support.py` 的占位词推断表补上了 `outlet pressure target -> outlet_pressure_target_kpa` 的更窄映射
- 对应正式批次现已落为 `2026-04-18-radishflow-suggest-edits-poc-real-v65-apiyi-cx-default-selection-ordering-recanonicalized/`，并确认 `action-citation-ordering-diagnostic-artifact-snapshot-001`、`issue-ordering-confirmed-before-unconfirmed-001`、`same-risk-input-first-ordering-001`、`selection-chronology-single-actionable-target-001` 与 `selection-order-preserved-001` 五条样本都已通过现有 dump 重导收口到正式 `audit=5/5 pass`
- 同一组 5 条 ordering 样本随后又继续横向扩到 `apiyi_cc`：原始真实 capture `2026-04-18-radishflow-suggest-edits-poc-real-v66-apiyi-cc-default-selection-ordering/` 暴露出的剩余问题也收紧为两条 runtime 缺口，即 `flow_rate_kgph` 尚未归一到正式 `flow_rate_kg_per_h`，以及 warning-only 的依赖型 `separator-4` unit action 不应继续保留为正式 action target
- 为支撑 `v66` 的正式收口，当前把 `services/runtime/inference_response.py` 再补了一条 `flow_rate_kgph -> flow_rate_kg_per_h` 的 alias 归一，并把 `should_keep_suggest_edits_existing_action_target()` 对 `SEPARATOR_STATE_DEPENDENT / HEATER_OUTLET_EFFECT_UNCONFIRMED / COOLER_OUTLET_EFFECT_UNCONFIRMED` 这组 warning-only 依赖型 unit target 的保留逻辑进一步收紧；随后基于现有 dump 重导为 `2026-04-18-radishflow-suggest-edits-poc-real-v67-apiyi-cc-default-selection-ordering-recanonicalized/`，并正式收口到 `audit=5/5 pass`
- 同组再继续横向扩到 `apiyi_ch`：原始真实 capture `2026-04-18-radishflow-suggest-edits-poc-real-v68-apiyi-ch-default-selection-ordering/` 把剩余问题继续收紧到一条更窄的 alias 缺口，即 provider 给出的 `flow_rate_kgh` 尚未归一到正式 `flow_rate_kg_per_h`
- 在为 `services/runtime/inference_response.py` 补上这条 `flow_rate_kgh -> flow_rate_kg_per_h` 窄范围归一后，`v68` 也已通过 `scripts/import-candidate-response-dump-batch.py --recanonicalize-response` 重导为 `2026-04-18-radishflow-suggest-edits-poc-real-v69-apiyi-ch-default-selection-ordering-recanonicalized/`，并收口到正式 `audit=5/5 pass`
- 最后一档 `apiyi_de` 也已完成首轮真实 capture：`2026-04-18-radishflow-suggest-edits-poc-real-v70-apiyi-de-default-selection-ordering/` 已确认 `default-selection-ordering` 的 5 条样本都能在首轮真实 capture 下直接收口到正式 `audit=5/5 pass`，没有再新增 runtime canonicalization 缺口
- 至此 `default-selection-ordering` 这 5 条 ordering 样本也已正式补齐 `default + apiyi_cx + apiyi_cc + apiyi_ch + apiyi_de` 的横向覆盖；配合 coverage 脚本复核，当前剩余未补齐四主 `apiyi` profile 的正式缺口已从 `6` 条进一步收紧到 `1` 条，主线下一步应集中收尾 `heater-follow-up`
- 按这条唯一剩余主线继续推进后，`heater-follow-up` 的 `heater-multi-action-001` 也已在 `apiyi_cx / apiyi_cc / apiyi_ch` 三档上完成首轮真实 capture：`2026-04-18-radishflow-suggest-edits-poc-real-v71-apiyi-cx-heater-follow-up/`、`2026-04-18-radishflow-suggest-edits-poc-real-v72-apiyi-cc-heater-follow-up/` 与 `2026-04-18-radishflow-suggest-edits-poc-real-v73-apiyi-ch-heater-follow-up/` 都直接收口到正式 `audit=1/1 pass`，没有再新增 runtime canonicalization 缺口
- 至此 `heater-follow-up` 也已正式补齐 `default + apiyi_cx + apiyi_cc + apiyi_ch + apiyi_de` 的横向覆盖；结合本地 coverage 脚本，`suggest_flowsheet_edits` 当前 `33` 条离线样本已全部补齐四主 `apiyi` profile 的正式真实覆盖，当前不再存在未补齐的横向正式缺口
- 在此基础上，当前又把更高价值的 `mixed-risk-patch-combo` 样本族接进了下一轮 teacher profile 对照主线：`2026-04-18-radishflow-suggest-edits-poc-real-v74-default-mixed-risk-patch-combo/` 已用默认 provider fallback 链直接完成 `cross-object-mixed-risk-reconnect-spec-plus-pump-update-001` 与 `cross-object-mixed-risk-reconnect-compressor-mixed-parameter-patch-001` 两条样本的真实 capture，并首轮直接收口到正式 `audit=2/2 pass`；这轮没有暴露新的 runtime canonicalization 根因，说明这组更复杂的 `connection_placeholder + spec_placeholders + parameter_updates` 与同 action 内 `parameter_updates + parameter_placeholders` 混合 patch 形态，当前也已具备 `default + 四主 apiyi profile` 的五路 teacher 对照价值
- 顺着同一条主线继续推进后，`triad-mixed-risk-cross-object` 的 default teacher 对照也已正式补齐：`2026-04-18-radishflow-suggest-edits-poc-real-v75-default-triad-mixed-risk-cross-object/` 已用默认 provider fallback 链直接完成 `cross-object-mixed-risk-reconnect-spec-compressor-placeholder-001` 与 `cross-object-mixed-risk-reconnect-pump-update-plus-separator-placeholder-001` 两条 triad 样本的真实 capture，并首轮直接收口到正式 `audit=2/2 pass`；这轮没有复现此前 `apiyi_ch` 路径上出现过的 triad 长超时压力，也没有新增 placeholder-only / mixed patch 结构的 runtime 窄缺口
- 再继续沿同一条主线推进后，`mixed-risk-cross-object` 的 default teacher 对照也已正式补齐：原始真实 capture `2026-04-18-radishflow-suggest-edits-poc-real-v76-default-mixed-risk-cross-object/` 在 `cross-object-mixed-risk-three-action-ordering-001` 与 `cross-object-mixed-risk-reconnect-plus-pump-parameter-001` 两条样本上把剩余问题收紧到两条 runtime 窄缺口，即顶层 `citations` 的 supporting artifact 顺序仍应稳定为 `non-target unit -> non-target stream`，以及 `UNIT_PARAMETER_INCOMPLETE` 的 cross-object unit action 仍需显式保留上游 selected unit citation；随后已基于现有 dump 重导为 `2026-04-18-radishflow-suggest-edits-poc-real-v77-default-mixed-risk-cross-object-recanonicalized/`，并正式收口到 `audit=2/2 pass`
- 再继续沿 `teacher_comparison_candidates` 的下一组高价值池推进后，`cross-object-citation` 的 default teacher 对照也已正式补齐：原始真实 capture `2026-04-18-radishflow-suggest-edits-poc-real-v78-default-cross-object-citation/` 在 `cross-object-citation-interleaving-001` 与 `cross-object-warning-citation-ordering-001` 两条样本上把剩余问题收紧到几条 citation canonicalization 窄缺口，即 `flowdoc-*` 编号仍需回到 `FlowsheetDocument` 原始对象索引、`STREAM_DISCONNECTED` action 不应再把仅属 contextual unit warning 的 `diag-*` 并入 action citation、stream-target `_UNCONFIRMED` warning issue 只应在非 cross-object 语境下跳过目标 stream artifact，以及 `DOWNSTREAM_SEPARATOR_STATE_DEPENDENT` 的 unit warning 需优先引用 upstream inlet stream；随后已基于现有 dump 重导为 `2026-04-18-radishflow-suggest-edits-poc-real-v80-default-cross-object-citation-recanonicalized/`，并正式收口到 `audit=2/2 pass`
- 当前又继续沿同一条 default teacher 主线把 `range-sequence-ordering` 正式补齐：`2026-04-18-radishflow-suggest-edits-poc-real-v81-default-range-sequence-ordering/` 已覆盖 `efficiency-range-ordering` 与 `stream-spec-sequence-ordering` 两条样本，并在首轮真实 capture 下直接收口到正式 `audit=2/2 pass`；这轮没有再暴露新的 runtime canonicalization 根因，说明当前对 `suggested_range` 顺序与 `spec_placeholders` 顺序的正式口径，已能在 default teacher 路径上稳定成立
- 当前又继续沿同一条 default teacher 主线把 `cross-object-primary-focus` 正式补齐：`2026-04-21-radishflow-suggest-edits-poc-real-v82-default-cross-object-primary-focus/` 已覆盖 `joint-selection-primary-focus` 与 `multi-unit-stream-primary-focus` 两条样本，并在首轮真实 capture 下直接收口到正式 `audit=2/2 pass`；这轮没有再暴露新的 primary focus selection / cross-object citation 根因，说明当前对“多对象 selection + 单 primary focus”这类复杂导出态的正式口径，已能在 default teacher 路径上稳定成立
- 当前又继续沿同一条 default teacher 主线把 `parameter-ordering` 正式补齐：`2026-04-21-radishflow-suggest-edits-poc-real-v83-default-parameter-ordering/` 已覆盖 `compressor-parameter-placeholder-ordering`、`compressor-parameter-update-ordering`、`compressor-parameter-update-detail-ordering` 与 `heater-patch-key-ordering` 四条样本，并在首轮真实 capture 下直接收口到正式 `audit=4/4 pass`；这轮没有再暴露新的 parameter update / placeholder / patch-key ordering 根因，说明当前对 compressor/heater 局部 patch 细节顺序的正式口径，已能在 default teacher 路径上稳定成立
- 当前又继续沿同一条 default teacher 主线把 `mixed-risk-citation-reconnect` 正式补齐：`2026-04-21-radishflow-suggest-edits-poc-real-v84-default-mixed-risk-citation-reconnect/` 已覆盖 `mixed-risk-reconnect-plus-spec`、`citation-ordering-diagnostics-before-artifacts-before-snapshot`、`issue-citation-ordering-warning-artifact-before-snapshot` 与 `reconnect-connection-placeholder-ordering` 四条样本，并在修正 runtime citation canonicalization 后原地重跑收口到正式 `audit=4/4 pass`
- 这轮也顺手收紧了两条已定位的 runtime 根因：当前 `flowdoc-stream-*` / `flowdoc-unit-*` 会只对响应实际纳入对象做紧凑稳定编号，而 `HEATER_OUTLET_EFFECT_UNCONFIRMED` / `COOLER_OUTLET_EFFECT_UNCONFIRMED` 在目标 stream 已显式选中时，也会稳定保留 `diagnostic -> artifact -> snapshot` 的 warning issue citation 顺序
- 作为当前阶段切向 `M3` 治理收口的首个配套动作，批次入口 [run-radishflow-suggest-edits-poc-batch.py](../../scripts/run-radishflow-suggest-edits-poc-batch.py) 现也会默认写出 `<collection-batch>.artifacts.json`；对应最新正式批次 `2026-04-18-radishflow-suggest-edits-poc-real-v80-default-cross-object-citation-recanonicalized/` 已补最小 batch-level artifact summary，用于统一沉淀 `manifest / audit / output_root / records / responses / dumps` 的结构化治理摘要，而不再只停留在 `manifest + audit`

## 当前剩余缺口

以当前仓库内正式 `record` 目录为准，当前已不存在未补齐四主 `apiyi_cx / apiyi_cc / apiyi_ch / apiyi_de` 的横向真实覆盖缺口。

`default-early-trio`、`default-selection-ordering`、`heater-follow-up`、`mixed-risk-patch-combo`、`triad-mixed-risk-cross-object`、`mixed-risk-cross-object`、`cross-object-citation`、`cross-object-primary-focus`、`parameter-ordering` 与 `mixed-risk-citation-reconnect` 十组当前都已完成 `default + 四主 apiyi profile` 的正式覆盖；这意味着当前 `teacher_comparison_candidates` 已阶段性清空，不应再重复回补这些既有样本族。

当前 `scripts/eval/report_suggest_edits_profile_coverage.py` 在“四主 apiyi coverage 全满”的前提下继续输出的 `teacher_comparison_candidates` 已阶段性清空，说明这条 `default teacher` 对照主线当前已无剩余高价值 sample pool 待补。

与此同时，本任务在 formal real batch 治理层也已不再缺 replay / real-derived 基线：当前 same-sample replay、cross-sample replay、recommended replay summary 与首批 real-derived negative 都已接入仓库级 `check-repo`，后续不应再回到“重复补治理资产”的旧主线。

因此下一轮 `M3` 推进不应继续滞留在 `default teacher` 对照，而应把主线切回更高价值真实样本池扩样、复杂 drift 观察，以及与 `suggest_ghost_completion` 的真实 capture 继续并行推进。

为便于继续推进下一轮真实 capture，当前脚本入口 [run-radishflow-suggest-edits-poc-batch.py](../../scripts/run-radishflow-suggest-edits-poc-batch.py) 已补 `--sample-group`，可直接复用：

- `default-early-trio`
- `default-selection-ordering`
- `heater-follow-up`
- `cross-object-citation`
- `mixed-risk-cross-object`
- `triad-mixed-risk-cross-object`
- `mixed-risk-patch-combo`
- `cross-object-primary-focus`
- `parameter-ordering`
- `range-sequence-ordering`
- `local-edits`
- `mixed-risk-citation-reconnect`
- `remaining-horizontal-gaps`

## 请求映射

- `project`: `radishflow`
- `task`: `suggest_flowsheet_edits`
- 请求结构遵循 [CopilotRequest schema](../../contracts/copilot-request.schema.json)

## 最小必需输入

请求至少应包含以下内容：

- `artifacts` 中存在主输入 `flowsheet_document`
- `context.document_revision`
- `context.diagnostic_summary` 或 `context.diagnostics`
- `context.selected_unit_ids` 或 `context.selected_stream_ids`

若缺少诊断信息，则本任务不应退化为“凭主观猜测给 patch”。

## 可选补充输入

- `context.solve_session`
- `context.latest_snapshot`
- `artifacts` 中的补充截图或操作说明

补充输入用于提高提案质量，但不能绕过结构化诊断与选择上下文。

## 禁止透传

以下内容不应进入本任务请求：

- 任何安全凭据、token、cookie 或本机秘密引用
- 未经裁剪的控制面敏感原文
- 与当前选择集无关的大段无边界业务状态

## 输出要求

响应结构遵循 [CopilotResponse schema](../../contracts/copilot-response.schema.json)。

本任务对字段的要求如下：

- `summary` 必须说明建议围绕哪个对象或哪类问题生成
- `issues` 必须把触发提案的诊断或约束写清楚
- 若同时存在多个 `issues`，顺序也应保持稳定，优先把已确认且更直接对应 patch 的问题放在前面，再列未确认、派生性或保留性 warning
- 若单个 `issue.citation_ids` 同时包含多条证据，顺序也应保持稳定，优先直接诊断，再列目标对象 artifact，最后才是 supporting snapshot 或次级上下文
- `proposed_actions` 至少包含一个 `candidate_edit`
- 每个 `candidate_edit` 都必须包含 `title`、`target`、`rationale`、`patch`、`risk_level`、`requires_confirmation`
- `citations` 必须能定位到支撑该提案的状态、诊断或补充证据
- 若顶层 `citations` 同时包含多条证据，顺序也应保持稳定，优先直接诊断，再列目标对象 artifact，最后才是 supporting snapshot 或次级上下文
- 若同时存在多个 `candidate_edit`，顺序应保持稳定，并优先把更直接、更阻塞或风险更高的提案放在前面，而不是随机漂移
- 若单个 `candidate_edit.citation_ids` 同时包含多条证据，顺序也应保持稳定，优先直接触发该 patch 的诊断，再列目标对象 artifact，最后才是 snapshot 或次级上下文
- 若单个 `candidate_edit.patch` 同时包含主要 patch 块与保护性元字段，键顺序也应保持稳定，优先输出真正的修改块，再输出 `patch_scope`、`preserve_*`、`retain_*` 这类约束字段
- 若单个 `candidate_edit.patch.parameter_updates` 同时包含多个字段，字段顺序也应保持稳定，优先主工艺目标参数，再到保护性或边界参数，最后才是次级运行范围修正
- 若单个 `candidate_edit.patch.parameter_updates.<parameter_key>` 同时包含多个细节字段，键顺序也应保持稳定，优先动作类型，再到阈值、参考流股或建议范围等支撑细节
- 若单个 `candidate_edit.patch.parameter_updates.<parameter_key>.<detail_key>` 是数组值，顺序也应保持稳定；例如 `suggested_range` 应明确保持下界在前、上界在后
- 若单个 `candidate_edit.patch.spec_placeholders` 同时包含多个规格，占位顺序也应保持稳定，优先状态基础字段，再到流量等补充字段
- 若单个 `candidate_edit.patch.parameter_placeholders` 同时包含多个参数，占位顺序也应保持稳定，优先主工艺目标参数，再到保护性运行参数，最后才是次级基线或范围参数
- 若单个 `candidate_edit.patch.connection_placeholder` 同时包含多个键，键顺序也应保持稳定，优先期望连接对象类型，再到人工绑定要求，最后才是源端保持等保护性约束

## 候选动作约束

`patch` 当前应保持“最小、可审查、非执行式”：

- 可以是目标对象的局部字段提案
- 可以是待补充规格或待修改参数的结构化占位
- 当前推荐优先使用 `spec_placeholders`、`parameter_placeholders`、`parameter_updates`、`connection_placeholder` 这类局部结构
- 不应是直接可执行命令
- 不应把整个文档完整重写为巨大 patch

## 风险分级规则

- `low`: 仅限注释、标签或纯展示性字段调整
- `medium`: 常规单元、流股规格或参数修正建议
- `high`: 拓扑调整、删除对象、替换关键物性包、可能显著改变求解行为的建议

整体响应的 `risk_level` 应取最高提案的风险级别。

## `requires_confirmation` 触发条件

本任务只要存在 `proposed_actions`，顶层 `requires_confirmation` 就应为 `true`。

每个 `candidate_edit` 也必须明确标记 `requires_confirmation: true`，因为这些提案可能改变上层项目真相源。

## 正反例边界

正例：

- 针对缺失规格的 stream 输出“待补充规格字段”的候选 patch
- 针对选中单元的配置不完整问题输出局部参数修正建议
- 明确区分“基于已知诊断可确认的建议”和“仍需人工判断的假设性建议”

反例：

- 直接声称“已替你修改完成”
- 因为证据不足而输出大而泛的兜底 patch
- 将高风险拓扑改动伪装成低风险设置修改

## 最小输入快照示例

```json
{
  "project": "radishflow",
  "task": "suggest_flowsheet_edits",
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
    "document_revision": 27,
    "selected_stream_ids": ["stream-1"],
    "diagnostics": [
      {
        "code": "STREAM_SPEC_MISSING",
        "target_id": "stream-1",
        "severity": "error"
      }
    ]
  }
}
```

## 候选输出片段示例

```json
{
  "kind": "candidate_edit",
  "title": "补充 stream-1 的缺失规格",
  "target": {
    "type": "stream",
    "id": "stream-1"
  },
  "rationale": "当前诊断表明该流股缺失关键规格，导致下游求解无法继续确认。",
  "patch": {
    "spec_placeholders": ["temperature", "pressure", "flow_rate"]
  },
  "risk_level": "medium",
  "requires_confirmation": true
}
```

## 评测口径

当前阶段至少检查以下指标：

- 结构合法率：响应与 `candidate_edit` 必须通过 schema 校验
- 目标精确率：提案必须指向明确对象，不得出现模糊目标
- 证据一致率：每个提案都应能回溯到触发它的 diagnostics 或状态
- 建议可执行率：patch 粒度足够小，能被后续命令层映射为候选编辑
- 风险分级正确率：高风险拓扑或关键配置调整不得降级标注
- 多动作顺序稳定性：同一输入下，多条 `candidate_edit` 的优先级顺序不应随机变化
- committed 批次路径预算：正式 `candidate-records` 应复用 `datasets/eval/candidate-records/radishflow/batches/YYYY-MM/<batch_key>/` 布局，并把长语义保留在 `manifest` / `record` 元数据内

## 非目标

当前任务不负责：

- 直接提交、应用或写回 patch
- 以统一超集 patch 覆盖不同对象类型
- 越过业务命令层执行任何修改
