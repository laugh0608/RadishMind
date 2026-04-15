# RadishMind 阶段路线图

更新时间：2026-04-15

## 路线图目标

本文档用于把 `RadishMind` 拆成可执行的阶段，不在一开始把任务混成“先造一个万能模型”。

本路线已经基于 `D:\Code\RadishFlow` 与 `D:\Code\Radish` 的真实上下文做过一轮调整：

- `RadishFlow` 第一批能力优先走结构化状态与诊断解释，而不是截图先行
- `Radish` 第一批能力优先走文档问答、Console/运营辅助与内容结构化建议
- `minimind-v` 当前作为默认 `student/base` 主线，围绕评测闭环进入适配与训练阶段
- `Qwen2.5-VL` 当前作为默认 `teacher` / 强基线，优先承担多模态对照评测与蒸馏参考
- `SmolVLM` 当前作为轻量本地对照组，优先承担小模型回归与部署下限比较

## M0：真实上下文审查与规划冻结

### 目标

建立最小文档真相源，并把项目定位、边界和任务矩阵从“推定口径”收口到“基于真实仓库的口径”。

### 任务

- 检查 `RadishMind` 仓库现状
- 审查 `RadishFlow` 与 `Radish` 的正式文档、关键目录与 AI 相关代码
- 修正文档中与真实仓库不一致的假设
- 冻结第一版项目定位、非目标和阶段定义

### 退出标准

- 新成员能在 15 分钟内理解项目定位
- 后续讨论能以文档为入口，而不是只靠聊天历史

## M1：统一协议与上下文打包

### 目标

建立可以接上层项目的最小 Copilot 协议和 context packer 骨架，并将评测、验证与模型工具链统一收口到 `Python`。

### 任务

- 定义 `CopilotRequest` / `CopilotResponse`
- 明确 artifact、引用、风险分级和 `requires_confirmation`
- 建立 `adapter-radishflow` 和 `adapter-radish` 的上下文打包骨架
- 明确哪些字段必须脱敏、禁止透传或只能摘要透传
- 将 `RadishFlow export -> adapter -> request` 主线推进到可回归状态：至少冻结 export snapshot、bootstrap 模板、preflight smoke validator，以及几条 committed exporter edge fixtures

### 退出标准

- 能完成一个端到端假请求
- 协议字段有第一版稳定口径
- `RadishFlow` 和 `Radish` 都能提供各自最小上下文包
- `RadishFlow` 的 exporter 在真实联合选择、`multi-unit + multi-stream + 单 primary focus`、更复杂 selection chronology + 单 actionable target、selection 顺序保持、纯 `uri + metadata.summary` 与 mixed support summary 脱敏摘要、以及多动作排序样本上有 committed fixture 和仓库级回归

## M2：`RadishFlow` 首个 PoC

### 目标

优先在 `RadishFlow` 上证明 Copilot 的业务价值，但第一批 PoC 以真实状态模型为中心。

### 推荐首个场景

- `flowsheet document + selection + diagnostics -> structured explanation`
- `flowsheet document + selection + diagnostics -> candidate edit suggestions`
- `entitlement / manifest / lease sync summary -> operator guidance`
- `selected unit placement + legal candidate set -> ghost topology completion`

### 任务

- 约定 `RadishFlow` 上下文快照格式
- 基于 `FlowsheetDocument`、选择集、诊断摘要和求解状态建立首批样本
- 为编辑器辅助场景冻结 `suggest_ghost_completion` 的输入输出口径，并优先围绕 `FlashDrum` / `Mixer` 建立最小样本
- 在 `suggest_ghost_completion` 上把 pre-model handoff、request assembly 与 response-level regression 推进到链式基线，并已先在三条链式模板上收口 `Tab / manual_only / empty / reject-no-retab / dismiss-no-retab / skip-no-retab`、same-candidate 一帧 cooldown 恢复 `Tab`、恢复窗口只看最近一条同 candidate 动作，以及 other-candidate 不共享 suppress 信号这四类交互边界
- 将 `suggest_ghost_completion` 从“只有 golden/eval 样本”推进到“已具备正式导入链”的最小 PoC：补齐任务 prompt、最小 runtime、批次 capture 入口，以及 `candidate_response_record -> manifest -> audit` 的最小回灌链；当前八批真实 batch 已正式收口到 `datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v2/`、`datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v3/`、`datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v4/`、`datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v5/`、`datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v6/`、`datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v7/`、`datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v8/` 与 `datasets/eval/candidate-records/radishflow/2026-04-11-radishflow-ghost-poc-real-v9/`，固定 3 个代表样本覆盖 `Tab / manual_only / empty`，其中 `v3` 验证了 malformed JSON 的重新归一化修复，`v4` 暴露出批处理卡顿观察项，`v5` 则进一步收口了逐样本硬超时编排与 `manual_only` 多动作坏法修复，`v6` 验证了在 `openrouter` 限流时切到 `deepseek` fallback profile 仍可继续完成 3/3 pass 的真实 capture，`v7` 则进一步收口了 openrouter 默认模型废弃与 `summary` / `answer.text` JSON 字符串漂移两条新观察项，`v8` 继续确认了 openrouter 候选模型的 `404`/`429` 可用性阻塞可通过 fallback profile 绕过，而 `v9` 则进一步确认即便同 provider 的备选模型已可调用，也仍可能在 `manual_only` 主路径上暴露不可正式导入的 schema-invalid 质量漂移；当前正式批次因此继续由 `deepseek` fallback profile 完成收口
- 将 `suggest_flowsheet_edits` 从“只有 fixture/eval 回归”推进到“已形成真实 teacher 批次治理入口”：当前除 `datasets/eval/candidate-records/radishflow/2026-04-12-radishflow-suggest-edits-poc-mock-v1/` 这批最小 mock PoC 外，还已正式导入 `2026-04-13-radishflow-suggest-edits-poc-real-v3/`、`2026-04-13-radishflow-suggest-edits-poc-real-v4/`、`2026-04-14-radishflow-suggest-edits-poc-real-v5/`、`2026-04-14-radishflow-suggest-edits-poc-real-v6/`、`2026-04-14-radishflow-suggest-edits-poc-real-v7-apiyi-de/`、`2026-04-14-radishflow-suggest-edits-poc-real-v8-apiyi-de/`、`2026-04-14-radishflow-suggest-edits-poc-real-v9-apiyi-cc/`、`2026-04-14-radishflow-suggest-edits-poc-real-v10-apiyi-ch/`、`2026-04-14-radishflow-suggest-edits-poc-real-v11-apiyi-de-cross-object/`、`2026-04-14-radishflow-suggest-edits-poc-real-v12-apiyi-de-parameter-ordering/`、`2026-04-14-radishflow-suggest-edits-poc-real-v13-apiyi-cc-cross-object/`、`2026-04-14-radishflow-suggest-edits-poc-real-v14-apiyi-cc-parameter-ordering/`、`2026-04-14-radishflow-suggest-edits-poc-real-v15-apiyi-ch-cross-object/`、`2026-04-14-radishflow-suggest-edits-poc-real-v16-apiyi-ch-parameter-ordering/`、`2026-04-15-radishflow-suggest-edits-poc-real-v17-apiyi-de-cross-object-citation/`、`2026-04-15-radishflow-suggest-edits-poc-real-v18-apiyi-cc-cross-object-citation/`、`2026-04-15-radishflow-suggest-edits-poc-real-v19-apiyi-ch-cross-object-citation/`、`2026-04-15-radishflow-suggest-edits-poc-real-v20-apiyi-cx-cross-object-citation/`、`2026-04-15-radishflow-suggest-edits-poc-real-v21-apiyi-cx-mixed-risk-cross-object/`、`2026-04-15-radishflow-suggest-edits-poc-real-v22-apiyi-cc-mixed-risk-cross-object/`、`2026-04-15-radishflow-suggest-edits-poc-real-v23-apiyi-ch-mixed-risk-cross-object/`、`2026-04-15-radishflow-suggest-edits-poc-real-v24-apiyi-de-mixed-risk-cross-object/`、`2026-04-15-radishflow-suggest-edits-poc-real-v25-apiyi-cx-triad-mixed-risk-cross-object/`、`2026-04-15-radishflow-suggest-edits-poc-real-v26-apiyi-cc-triad-mixed-risk-cross-object/` 与 `2026-04-15-radishflow-suggest-edits-poc-real-v28-apiyi-de-triad-mixed-risk-cross-object/` 二十五批真实 capture，已把高风险重连、局部规格/参数占位、selection 顺序与 chronology、多动作链式优先级、顶层/issue/action citation 顺序、`apiyi_cx / apiyi_de / apiyi_cc / apiyi_ch` 在同一组复杂样本上的扩样本 profile 对照、更偏跨对象 primary-focus 的 selection 样本、更复杂的跨对象 citation 交错样本，以及 compressor/heater/pump 局部 patch 的 parameter / patch ordering 与 triad patch-shape 样本都接进正式 `candidate_response_record -> manifest -> audit` 治理链；其中 `v5 / v6` 两批均已收口到 `audit=4/4 pass`，`v7-apiyi-de` 收口到 `audit=1/1 pass`，`v8-apiyi-de`、`v9-apiyi-cc` 与 `v10-apiyi-ch` 则都在 4 条 mixed-risk / citation / reconnect 顺序样本上收口到 `audit=4/4 pass`，`v11-apiyi-de-cross-object`、`v13-apiyi-cc-cross-object` 与 `v15-apiyi-ch-cross-object` 已分别在 2 条跨对象 primary-focus 样本上收口到 `audit=2/2 pass`，`v12-apiyi-de-parameter-ordering` 与 `v14-apiyi-cc-parameter-ordering` 也已分别在 4 条 parameter / patch ordering 样本上收口到 `audit=4/4 pass`，`v17-apiyi-de-cross-object-citation`、`v18-apiyi-cc-cross-object-citation`、`v19-apiyi-ch-cross-object-citation` 与 `v20-apiyi-cx-cross-object-citation` 则都已分别在 2 条更复杂的跨对象 citation 交错样本上收口到 `audit=2/2 pass`，`v21-apiyi-cx-mixed-risk-cross-object`、`v22-apiyi-cc-mixed-risk-cross-object`、`v23-apiyi-ch-mixed-risk-cross-object` 与 `v24-apiyi-de-mixed-risk-cross-object` 也都已在 2 条 mixed-risk + cross-object citation 组合样本上通过当前 runtime `recanonicalize-response` 收口到 `audit=2/2 pass`，而 `v25-apiyi-cx-triad-mixed-risk-cross-object`、`v26-apiyi-cc-triad-mixed-risk-cross-object` 与 `v28-apiyi-de-triad-mixed-risk-cross-object` 也都已在 2 条 triad patch-shape mixed-risk 样本上收口到 `audit=2/2 pass`；其中 `apiyi_de` 仅出现过单次请求级 read-timeout retry，`apiyi_ch` 在同组 `v27` 上则已在首条样本重试后仍触发 `180s` 读超时，因此当前更应把它归类为 provider 可用性阻塞，而非正式质量阻塞
- 在 `suggest_flowsheet_edits` 这条真实 capture 主线上，继续把 task-level canonicalization 从“只兜底最小结构合法”推进到“可稳定吸收真实 teacher 的窄范围任务漂移”：当前已补齐 `flowdoc-*` 编号稳定化、`flow_rate` / `flow_rate_kg_h` / `outlet_temperature_target_c` / `outlet_temperature_target` / `target_outlet_temperature_c` 等近义占位归一、路径式 placeholder 回收到稳定字段名、`mass_flow_kg_per_h` / `mass_flow_kg_h -> flow_rate_kg_per_h` 等别名收口、`STREAM_DISCONNECTED` 单诊断样本上 `issues[*].citation_ids` 对齐 `diag -> artifact -> snapshot` 的正式口径，以及 reconnect patch 中 `retain_existing_source_binding=true` 的稳定保留、warning issue 无论 provider 原始顺序如何都优先收口到 `diag -> artifact -> snapshot` 的排序规则、`suggest_flowsheet_edits` 任务下 malformed JSON 中多余闭合 brace 的窄范围修复和无效 `flowdoc-*` 引用回退到当前样本 canonical target artifact 的收口；在 `v12 / v14` 暴露的新坏法上，还继续补上了 compressor / pump / valve 局部参数诊断进入 auto-synthesize 集合、空 `proposed_actions` 回收到稳定 `parameter_placeholders` / `parameter_updates`、“纯局部占位 patch”样本不再因为 `latest_snapshot` 存在而被过度压成 `status=partial`、dict/path 形态的 placeholder 回收到稳定字段名、首项 provider placeholder 与合成 placeholder 序列稳定合并，以及局部参数 patch 的 `risk_level` 收口回正式 `medium` 的规则；后续仍应优先按真实 dump 暴露的新坏法窄修复，而不是先去堆更宽泛的 prompt fallback
- 将 Python 工具链基线继续收口到仓库正式入口：当前已补最小 `pyproject.toml` / `uv` 配置声明，使后续脚本、校验与依赖治理不再只依赖隐式环境
- 在上述正式导入 PoC 上继续接入后续真实教师批次，而不是继续无上限扩张 fallback 样本
- 设计结构化输出与 UI 侧回显方式
- 保留 `canvas screenshot` 作为补充输入，而不是强依赖入口

### 退出标准

- 至少一个场景可稳定演示
- 输出格式可被 UI 侧稳定消费
- 没有侵入求解热路径或控制面真相源

## M3：数据集与评测体系

### 目标

建立可迭代的样本与评测闭环。

### 任务

- 建立 `synthetic`、`annotated`、`eval` 三类数据目录
- `RadishFlow` 建立状态优先样本和后续截图补充样本
- `Radish` 建立文档、论坛和 Console 知识样本
- 定义任务指标与离线回归脚本
- 先把 `Radish` 的 `answer_docs_question` 落成“召回输入 + golden_response”最小回归 runner，再补外部 `candidate_response_record`、统一负例回放与真实候选输出回灌
- 对 `RadishFlow suggest_flowsheet_edits`，继续把 response-level regression 从“边界样本存在”推进到“顺序约束显式可检验”：至少覆盖 `issues`、顶层 `citations`、`issues[*].citation_ids`、`candidate_edit` 动作顺序、`candidate_edit.citation_ids` 与 `patch` 内部多层键/数组顺序
- 为 `Radish docs QA` 的真实/模拟 batch 建立 same-sample / cross-sample replay index、recommended replay summary 与 batch artifact summary 的统一治理链
- 对“基于真实 bad record 派生、但仍以本地 fixture 入仓”的负例，补 `source_candidate_response_record` 结构化反链，并生成独立 real-derived index，至少支持 `source_record_groups`、`violation_groups` 与 `pattern_groups` 三个审计维度
- 优先把真实风格 bad output 扩样成 repeated pattern，而不是继续泛化堆叠 simulated negative；当前重点围绕 `read_only_check` 缺失、citation/source drift 与 `answers` 缺失三类主失败面推进
- 当前 `Radish docs QA` 这条治理子线已把 `2026-04-05` real batch 中仍为 singleton 的 source 全部推进到同源双派生，real-derived index 已收口到 `34` 条 linked negatives；下一步重点转向跨 source 复合 drift 扩样，以及 `pattern` / `violation` 是否需要进一步结构化

### 推荐指标

- 结构合法率
- 字段完整率
- 引用命中率
- 建议可执行率
- 风险分级正确率
- 稳定顺序一致率

### 退出标准

- 不同模型和提示词可以被稳定比较
- 结果不再只靠主观观感判断
- 对 `suggest_flowsheet_edits` 这类结构化建议任务，能够明确看见“是否答对”和“是否保持稳定排序”两个维度，而不把顺序漂移混进主观体验
- 对真实 batch 的失败样本，能够看清“哪些已沉淀为 replay、哪些已扩成 repeated real-derived pattern、哪些仍缺治理闭环”

## M4：`minimind-v` student/base 主线推进

### 目标

在 `Qwen2.5-VL` 强基线验证过任务后，推进 `minimind-v` 主线的训练、蒸馏和领域适配。

### 任务

- 评估 `minimind-v` 作为默认 `student/base` 主线的改造成本与接入优先级
- 明确哪些任务值得进入 `minimind-v` 主线，哪些继续保留在 `Qwen2.5-VL` 或工具侧
- 建立领域样本格式转换脚本
- 做第一轮小规模微调或蒸馏实验
- 引入 `SmolVLM` 作为轻量对照组，验证低资源场景下的回归下限

### 退出标准

- `minimind-v` 在目标任务上达到可对比的基础表现
- 可以明确知道下一步应继续训练、补数据、优化工具路由，还是调整对照模型规模

## M5：`Radish` 首批任务接入

### 目标

在共享协议和评测基线上，把统一服务扩展到 `Radish` 的首批高价值任务。

### 推荐首批任务

- `answer_docs_question`
- `explain_console_capability`
- `suggest_forum_metadata`
- `summarize_doc_or_thread`
- `interpret_attachment`

### 任务

- 接入 `Docs/`、在线文档、论坛 Markdown 和 Console 文档作为知识来源
- 比较 `Radish` 与 `RadishFlow` 的共享协议和差异字段
- 设计内容辅助与文档问答的结构化输出
- 明确不进入自动治理、自动授权和自动写回

### 退出标准

- `Radish` 至少有一个内部可用场景落地
- 支持第二个项目后不破坏 `RadishFlow` 第一条接入链

## M6：双项目工程化收口

### 目标

在两个项目都已有可用场景后，再收口多项目路由、审计、缓存和部署策略。

### 任务

- 建立多项目路由与版本策略
- 收口缓存、审计、超时、降级和失败策略
- 对齐多模型回退、检索开关和工具调用策略
- 明确本地、局域网和远端三类部署模式的职责

### 退出标准

- 两个项目都能复用统一网关与协议
- 工程化策略清晰，不因支持第二个项目而出现边界漂移

## 当前推荐顺序

建议严格按以下顺序推进：

1. `M0`
2. `M1`
3. `M2`
4. `M3`
5. `M4`
6. `M5`
7. `M6`

不要在 `M1` 之前就直接开大规模训练，也不要在没有评测基线前频繁切换底座模型。

## 当前优先推进

在正式进入实现期前，当前建议按以下顺序继续推进：

1. 为 `RadishFlow` 首批 3 个任务继续扩展真实样本与 `golden_response` / `candidate_response` 口径，优先补控制面冲突态和对抗样本
2. 将 `RadishFlow suggest_flowsheet_edits` 从“首批真实 provider capture 已形成”继续推进到“更复杂真实样本面已稳定收口”；当前 `2026-04-12-radishflow-suggest-edits-poc-mock-v1`、`2026-04-13-radishflow-suggest-edits-poc-real-v3`、`2026-04-13-radishflow-suggest-edits-poc-real-v4`、`2026-04-14-radishflow-suggest-edits-poc-real-v5`、`2026-04-14-radishflow-suggest-edits-poc-real-v6`、`2026-04-14-radishflow-suggest-edits-poc-real-v7-apiyi-de`、`2026-04-14-radishflow-suggest-edits-poc-real-v8-apiyi-de`、`2026-04-14-radishflow-suggest-edits-poc-real-v9-apiyi-cc`、`2026-04-14-radishflow-suggest-edits-poc-real-v10-apiyi-ch`、`2026-04-14-radishflow-suggest-edits-poc-real-v11-apiyi-de-cross-object`、`2026-04-14-radishflow-suggest-edits-poc-real-v12-apiyi-de-parameter-ordering`、`2026-04-14-radishflow-suggest-edits-poc-real-v13-apiyi-cc-cross-object`、`2026-04-14-radishflow-suggest-edits-poc-real-v14-apiyi-cc-parameter-ordering`、`2026-04-14-radishflow-suggest-edits-poc-real-v15-apiyi-ch-cross-object`、`2026-04-14-radishflow-suggest-edits-poc-real-v16-apiyi-ch-parameter-ordering`、`2026-04-15-radishflow-suggest-edits-poc-real-v17-apiyi-de-cross-object-citation`、`2026-04-15-radishflow-suggest-edits-poc-real-v18-apiyi-cc-cross-object-citation`、`2026-04-15-radishflow-suggest-edits-poc-real-v19-apiyi-ch-cross-object-citation`、`2026-04-15-radishflow-suggest-edits-poc-real-v20-apiyi-cx-cross-object-citation`、`2026-04-15-radishflow-suggest-edits-poc-real-v21-apiyi-cx-mixed-risk-cross-object`、`2026-04-15-radishflow-suggest-edits-poc-real-v22-apiyi-cc-mixed-risk-cross-object`、`2026-04-15-radishflow-suggest-edits-poc-real-v23-apiyi-ch-mixed-risk-cross-object`、`2026-04-15-radishflow-suggest-edits-poc-real-v24-apiyi-de-mixed-risk-cross-object`、`2026-04-15-radishflow-suggest-edits-poc-real-v25-apiyi-cx-triad-mixed-risk-cross-object`、`2026-04-15-radishflow-suggest-edits-poc-real-v26-apiyi-cc-triad-mixed-risk-cross-object` 与 `2026-04-15-radishflow-suggest-edits-poc-real-v28-apiyi-de-triad-mixed-risk-cross-object` 已把 3 条代表样本、4 条中风险局部编辑样本、4 条多动作/selection 顺序/citation 交叉引用样本、4 条 issue/citation/action 顺序样本、`apiyi_cx / apiyi_de / apiyi_cc / apiyi_ch` 在相同复杂样本集上的扩样本 profile 对照、2 条跨对象 primary-focus 样本、2 条更复杂的跨对象 citation 交错样本、2 条 mixed-risk + cross-object citation 组合样本，以及 2 条 triad patch-shape mixed-risk 样本都接进 `candidate_response_record -> manifest -> audit`；同时这轮也已确认新样本暴露的问题可以继续沿“修正 eval 口径 / 窄范围 runtime canonicalization / 用既有 dump 重导正式 batch”的既有治理路径收口，而不是回头反复重打 provider。当前更合理的下一步，是把 `v27 / apiyi_ch` 正式收口为 triad 样本上的 provider 可用性阻塞，再决定是否进一步拉长超时，还是转入下一组更复杂 mixed-risk patch 组合样本设计。`apiyi_ch` 与 `apiyi_de` 的零星请求级 timeout 现象则继续按可用性观察项跟踪
3. 将 `RadishFlow / suggest_ghost_completion` 从“链式基线已闭环”继续推进到“真实 capture 已正式入仓的 editor assist PoC”；当前仓库已补齐任务 prompt、最小 runtime、轻量批次 capture 入口，以及 dump 重新归一化后的正式导入链，`v2 / v3 / v4 / v5 / v6 / v7 / v8 / v9` 八批 `Tab / manual_only / empty` 三样本真实 batch 都已完成入仓与 `audit=3/3 pass`，其中 `v3` 已把真实 provider 的稳定 malformed JSON 失败面收口进 runtime，`v4` 暴露出批处理中的单样本 provider 卡顿观察项，`v5` 则继续把这条执行层问题收口为逐样本单进程 + 硬超时治理，并新增吸收了 `manual_only` 多动作 JSON 提前关掉 `proposed_actions` / `answers` 作用域的坏法，`v6` 验证了当前 provider 链路已可通过 `openrouter / deepseek` fallback 继续完成真实 capture，`v7` 则继续把 openrouter 默认模型废弃与 `summary` / `answer.text` JSON 字符串漂移收口进配置口径和 runtime，`v8` 进一步确认当前新增阻塞主要仍集中在 openrouter 模型可用性与短窗口限流，而 `v9` 则继续确认即便某些 openrouter 备选模型已可调用，也仍可能在 `manual_only` 主路径上暴露不可正式导入的 schema-invalid 质量漂移，因此正式批次仍需按既定口径切回 `deepseek` fallback 收口。下一步应继续跑下一批真实 teacher provider capture，并同时观察这些结构失败面、模型质量漂移与供应商可用性阻塞是否还会复现，再根据新增真实 batch 暴露的失败面决定是否补 recent-actions 样本或扩展导入治理
4. 继续沿 `RadishFlow export -> adapter -> request` 主线补更真实的 exporter 边界，但从“补单个 fixture”转到“优先补批量 smoke 或正式契约”；当前已具备 selection 契约、priority 契约与 batch smoke 入口，后续除非上游 exporter 继续暴露新边界，否则不再优先深挖这一层
5. 维护 `Radish` 文档问答已覆盖 `docs/wiki/attachments/forum/faq` 的混合召回基线，仅按需补少量极端冲突样本
6. 将 `Radish` 文档问答从“真实候选响应已接入”继续推进到 captured negative 批次扩充、real-derived repeated pattern 治理与最小导入流程；当前已完成 `2026-04-05` batch singleton source 收口，下一主线转向跨 source 复合 drift 扩样与结构化治理评估
7. 在 `contracts/` 基础上补 schema 校验示例与后续类型生成策略
8. 再进入模型对照、PoC 与训练路线验证

## 当前仍缺的关键决策

- 契约文件的具体实现形式
- `Qwen2.5-VL` 的首选尺寸与推理预算
- `SmolVLM` 的回归任务边界与保留条件
- `RadishFlow` 截图路线的进入时点
- 评测样本的标注和维护流程
- `Radish` docs QA 的真实 captured negative 批次采用目录清单、导入脚本还是两者并行维护
- `Radish` docs QA 的 real-derived negative pattern 是否继续只靠 `capture_metadata.tags` 约定，还是需要升级成更正式的结构化字段
- `Radish` docs QA 的 `violation_groups` 是否要长期保持“按完整违规文案分组”，还是后续补一层更轻量的 violation pattern 归一化
