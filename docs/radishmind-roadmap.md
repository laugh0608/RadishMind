# RadishMind 阶段路线图

更新时间：2026-04-29

## 路线图目标

本文档用于把 `RadishMind` 拆成可执行的阶段，不在一开始把任务混成“先造一个万能模型”。

本路线已经基于 `D:\Code\RadishFlow` 与 `D:\Code\Radish` 的真实上下文做过一轮调整：

- `RadishFlow` 第一批能力优先走结构化状态与诊断解释，而不是截图先行
- `Radish` 第一批能力优先走文档问答、Console/运营辅助与内容结构化建议
- 长期目标从“做一个模型”收口为“受控 Copilot / Agent 系统 + 可替换模型能力”，模型训练、推理 provider、工具调用和规则校验都应服务于任务级 agent 编排
- `minimind-v` 当前作为默认 `student/base` 主线，围绕评测闭环进入适配与训练阶段
- `Qwen2.5-VL` 当前作为默认 `teacher` / 强基线，优先承担多模态对照评测与蒸馏参考
- `SmolVLM` 当前作为轻量本地对照组，优先承担小模型回归与部署下限比较
- `RadishMind-Core` 当前定义为基座适配型自研主模型：首版目标 `3B` / `4B`，长期本地部署上限 `7B`
- 图片生成能力不并入 `RadishMind-Core` 主模型参数目标，默认通过独立 `RadishMind-Image Adapter` 与生图 backend 提供

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
- 将 `suggest_ghost_completion` 从“只有 golden/eval 样本”推进到“已具备正式导入链”的最小 PoC：补齐任务 prompt、最小 runtime、批次 capture 入口，以及 `candidate_response_record -> manifest -> audit` 的最小回灌链；当前真实 batch 已统一收口到 `datasets/eval/candidate-records/radishflow/batches/YYYY-MM/<batch_key>/` 短路径布局，固定 3 个代表样本覆盖 `Tab / manual_only / empty`，其中 `v3` 验证了 malformed JSON 的重新归一化修复，`v4` 暴露出批处理卡顿观察项，`v5` 则进一步收口了逐样本硬超时编排与 `manual_only` 多动作坏法修复，`v6` 验证了在 `openrouter` 限流时切到 `deepseek` fallback profile 仍可继续完成 3/3 pass 的真实 capture，`v7` 则进一步收口了 openrouter 默认模型废弃与 `summary` / `answer.text` JSON 字符串漂移两条新观察项，`v8` 继续确认了 openrouter 候选模型的 `404`/`429` 可用性阻塞可通过 fallback profile 绕过，而 `v9` 则进一步确认即便同 provider 的备选模型已可调用，也仍可能在 `manual_only` 主路径上暴露不可正式导入的 schema-invalid 质量漂移；当前正式批次因此继续由 `deepseek` fallback profile 完成收口。此后 `v10` 到 `v25` 已把真实 capture 从固定 trio 扩到十六批高价值链式样本，正式覆盖 mixed-history 空建议、alternate candidate 切换、latest same-candidate manual_only 保持、latest-action precedence、other-candidate recovery、基础 no-retab/cooldown、foundation/conflict basics、general basics、latest dismiss / reject / skip cooldown 恢复、模板 residual backfill、residual conflict / recovery、residual cooldown tail 与 other-candidate cooldown tail 等语义；当前 replay / recommended replay / real-derived / audit 治理链均已接通，runtime 兜底层可阶段性收口；最新治理报表显示 `suggest_ghost_completion` 已达到 `real_captured=78/78`
- 将 `suggest_flowsheet_edits` 从“只有 fixture/eval 回归”推进到“已形成真实 teacher 批次治理入口”：当前除 `datasets/eval/candidate-records/radishflow/batches/YYYY-MM/<batch_key>/` 下的一批最小 mock PoC 外，还已把正式真实批次推进到 `v93`，并让现有 `33/33` 条离线样本至少各有一条真实覆盖；其中 `mixed-risk / citation / reconnect`、ordering 尾样、`mixed-risk patch combo`、cross-object primary-focus、parameter / patch ordering 与 local-edits 六组样本族都已完成 `apiyi_cx / apiyi_cc / apiyi_ch / apiyi_de` 四主 profile 的横向正式收口，`default` teacher 对照也已阶段性清空。今天又继续把高价值真实样本池的 `high-value-real-expansion-risk-localization` 与 `high-value-real-expansion-action-filtering` 两组正式跑通，其中 `v93` 的 action filtering 六条复合样本在基于现有 dump 修正 runtime canonicalization 后收口到 `audit=6/6 pass`。当前这条主线的重点已不再是回补旧 teacher pool 或复跑 `remaining-horizontal-gaps`，而是把既有样本资产上提为服务/API 接入门禁；真实 capture 只在新增非重复高价值假设时触发
- `RadishFlow` 的真实批次目录当前已进入路径治理阶段：后续在 `M2/M3` 推进中，新增 committed 批次必须继续复用 `batches/YYYY-MM/<batch_key>/` 短路径布局，而不是重新把长 `collection_batch` 和 sample 语义编码回物理路径
- 在 `suggest_flowsheet_edits` 这条真实 capture 主线上，task-level canonicalization 也已从“只兜底最小结构合法”推进到“可稳定吸收真实 teacher 的窄范围任务漂移”：当前 `flowdoc-*` 编号已正式收口为沿 `FlowsheetDocument` 原始对象索引稳定编号，`flow_rate` / `flow_rate_kg_h` / `mass_flow_kg_per_h` / `mass_flow_kg_h` / `mass_flow_rate_kg_h`、`outlet_temperature_target_c` / `outlet_temperature_target` / `target_outlet_temperature_c` 等近义占位会统一归一，路径式 placeholder 与多规格 `spec_placeholders` 顺序会回收到稳定字段名和固定顺序，`STREAM_DISCONNECTED` 等 error issue 的 citation、reconnect patch 中的 `retain_existing_source_binding=true`、warning citation 顺序、未转义中文引号与多余闭合 brace 的 malformed JSON 修补、placeholder-only patch 收口，以及 `efficiency_percent` 评审区间都已进入窄范围 runtime 治理；其中最新一轮还把 `STREAM_DISCONNECTED` action 对 connected-unit contextual warning 的并入边界收紧为“只在同一响应存在其他 actionable edit 时保留”，而 `apiyi_ch` 在 triad 路径上的超时也已收紧为样本级 `210s` override，而不是 profile 全局默认 timeout 上调
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
- 截至 `2026-04-27`，三条主治理链均已接通 `artifact summary / same-sample replay / cross-sample replay / recommended replay / real-derived negative` 资产；`suggest_ghost_completion` 的 `high-value-residual-other-candidate-cooldown-tail / v25` 已正式跑通并把真实覆盖推进到 `78/78`，`suggest_flowsheet_edits` 的 `high-value-real-expansion-action-filtering / v93` 也已正式跑通，当前仓库级优先级队列重新回到两条 RadishFlow 链都应避免复跑旧入口，只在新增非重复高价值样本时继续扩真实 capture
- 从 `v93` 之后开始，真实样本扩展降级为“有明确非重复高价值假设才触发”的支线；当前 `M3` 主线不再以新增 batch 数量衡量推进，而应把现有评测资产上提为服务/API 接入、回归门禁和集成演示的基础
- 当前已完成 `M3` 后半段服务/API 最小切片的第一轮落点：`services/gateway/copilot_gateway.py` 已把 schema-valid `CopilotRequest -> runtime -> CopilotResponse` 包成 advisory-only gateway envelope，`contracts/copilot-gateway-envelope.schema.json` 已冻结 `status / response / error / metadata` 最小结构，`scripts/run-radishflow-gateway-demo.py` 则通过 3 条 `RadishFlow suggest_flowsheet_edits` export fixture 固定 `export -> adapter/request -> handle_copilot_request -> gateway envelope` 的可复跑 demo，并用 summary fixture 锁定 `route / provider / advisory_only / request_validated / response_validated / requires_confirmation` 等上层调用依赖字段；随后又补 `scripts/check-radishflow-gateway-ui-consumption.py`，把 proposal-ready、unsupported route 与 schema-invalid 三类 envelope 映射成稳定 UI consumption summary，固定调用侧不得直接写回 `FlowsheetDocument`；当前还补了 `scripts/check-radishflow-candidate-edit-handoff.py`，把人工确认后的 `candidate_edit` 固定为非执行式 command candidate handoff，确保 failed / invalid envelope 不会进入命令层

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
- 三条已接线任务的 `artifact summary / replay / recommended replay / real-derived` 治理状态能由同一只读报告稳定盘点，且报告的 `next_mainline_focus` 不再把已跑通入口误导成必须继续复跑
- 至少一条 `RadishFlow` 场景能把 `adapter -> runtime -> response builder -> audit fixture` 串成服务/API 级最小切片，并复用现有 eval 样本作为验收门禁；当前这条门槛已由 `suggest_flowsheet_edits` 的 gateway demo manifest、summary fixture、UI consumption summary 与 candidate edit handoff summary 初步满足。在 `RadishFlow` / `Radish` 尚未准备真实模型 / Agent 接入前，这些 smoke 先冻结为未来验收门禁，后续重点转向 `M4` 的 Core 基座评估、训练样本格式和 Image Adapter schema，而不是继续证明单脚本可跑
- 继续扩样必须先写清楚新 drift 假设、覆盖缺口和退出条件；没有新假设时，不再把跑真实 provider batch 当作默认推进方式

## M4：`RadishMind-Core` student/base 主线推进

### 目标

在 `Qwen2.5-VL` 强基线验证过任务后，推进 `RadishMind-Core` 的训练、蒸馏和领域适配。当前 `RadishMind-Core` 默认不是从零预训练基础模型，而是基于开源 `student/base` 的项目专属适配模型。

模型规模目标按以下顺序收口：

- PoC / 首个稳定版：`3B` 或 `4B`
- 增强版与长期本地部署上限：`7B`
- 默认不把 `14B` / `32B` 作为自研主模型目标

### 任务

- 评估 `minimind-v` 作为 `RadishMind-Core` 默认 `student/base` 主线的改造成本与接入优先级
- 明确哪些任务值得进入 `RadishMind-Core` 主线，哪些继续保留在 `Qwen2.5-VL`、规则工具或外部 backend 侧
- 固定 `RadishMind-Core` 首版基座评估矩阵；当前已新增 `docs/radishmind-core-baseline-evaluation.md`、`scripts/checks/fixtures/radishmind-core-baseline-matrix.json` 与 `scripts/check-radishmind-core-baseline-matrix.py`
- 建立领域样本格式转换脚本；当前已先固定 `contracts/copilot-training-sample.schema.json`、`scripts/checks/fixtures/copilot-training-sample-basic.json` 与 `scripts/check-copilot-training-sample-contract.py`，并新增 `scripts/build-copilot-training-samples.py`、`scripts/checks/fixtures/copilot-training-sample-conversion-manifest.json` 与 `scripts/checks/fixtures/copilot-training-sample-conversion-summary.json`，把 `suggest_flowsheet_edits`、`suggest_ghost_completion` 与 `answer_docs_question` 各 3 条 committed eval golden_response 样本转换为可回归 `CopilotTrainingSample`；当前又补 `scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-manifest.json` 与 `scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-summary.json`，只允许 audit pass 的真实 candidate record 进入首批 `teacher_capture` 训练样本；`training/` 当前固定为训练治理目录，JSONL 默认输出到 `tmp/`，不提交大规模生成产物；更大训练集合的首个治理 manifest 已落到 `training/datasets/copilot-training-dataset-governance-v0.json`，固定 candidate record 入选、分层抽样复核、离线评测 holdout、质量门禁与退场条件；当前还补 `training/datasets/copilot-training-review-record-v0.json` 与 `training/datasets/copilot-training-holdout-split-v0.json`，把 planned 人工复核记录模板和三条主任务各 3 条的非重叠 holdout split 接入仓库级检查
- 做第一轮小规模微调或蒸馏实验
- 引入 `SmolVLM` 作为轻量对照组，验证低资源场景下的回归下限
- 定义 `RadishMind-Image Adapter` 的第一版结构化输出：图片生成意图、prompt、尺寸、风格、seed、负面词、编辑约束、风险确认和 artifact 元数据
- 对生图 backend 先采用现成小参数开源模型或其推理服务，不从零训练；首轮参考 `SD1.5` / `PixArt-δ 0.6B`，中期再评估 `Segmind-Vega` 或 `SD3.5 Medium 2.5B`

### 首版基座评估表

当前 `RadishFlow` 与 `Radish` 都尚未进入真实模型 / Agent 接入阶段，因此 `M4` 不应等待上层接线，也不应继续扩展本仓库内的模拟 UI / 命令层 summary。当前更合理的推进方式是先把 `RadishMind-Core` 的基座评估、领域数据格式和图片生成 adapter schema 做实，等上层项目具备接入条件后再使用既有 gateway smoke 作为验收门禁。

| 对象 | 当前定位 | 进入首轮工作的条件 | 暂不做的事 |
| --- | --- | --- | --- |
| `minimind-v` | 默认 `student/base` 主线候选 | 能承接 `CopilotRequest -> CopilotResponse` 结构化输出、中文任务说明、风险标记和 citation 对齐 | 不把它直接承诺为最终生产模型；先用离线评测决定是否保留 |
| `3B` `RadishMind-Core` | 首选本地友好档 | 在 32GB 级开发 / 部署机上能稳定跑通核心文本任务和结构化协议遵循 | 不强行承担复杂图片理解和图片像素生成 |
| `4B` `RadishMind-Core` | 首版增强候选 | 当 `3B` 在协议遵循、长上下文或多步骤推理上明显不足时进入对照 | 不在没有评测差距前默认扩大规模 |
| `7B` `RadishMind-Core` | 长期本地部署上限 | 只有当 `3B/4B` 无法满足关键任务，且部署机预算、量化策略和延迟目标可接受时进入 | 不作为首轮训练和默认部署基线 |
| `Qwen2.5-VL` | teacher / 强基线 | 用于复杂图文任务、标注参考、蒸馏数据和质量上界评估 | 不作为 RadishMind-Core 的本地部署目标 |
| `SmolVLM` | 轻量对照组 | 用于低资源回归、图片输入下限和部署成本对照 | 不替代主线 student/base |
| 生图 backend | 独立图片像素生成能力 | 先通过 `RadishMind-Image Adapter` 接收结构化意图并产出 artifact metadata | 不并入主模型训练目标，不从零训练 |

首轮建议先完成三件事：固定 `3B/4B` 的离线评测矩阵、生成最小 `CopilotRequest/CopilotResponse` 训练样本格式、为 `RadishMind-Image Adapter` 建立一组可回归的图片生成意图样本。当前 `RadishMind-Core` 首版基座评估矩阵已落成 `docs/radishmind-core-baseline-evaluation.md`、`scripts/checks/fixtures/radishmind-core-baseline-matrix.json` 与 `scripts/check-radishmind-core-baseline-matrix.py`，离线评测阻塞阈值也已落成 `scripts/checks/fixtures/radishmind-core-eval-thresholds.json` 与 `scripts/check-radishmind-core-eval-thresholds.py`，离线评测样本选择与结果记录格式已落成 `contracts/radishmind-core-offline-eval-run.schema.json`、`scripts/checks/fixtures/radishmind-core-offline-eval-run-basic.json` 与 `scripts/check-radishmind-core-offline-eval-run-contract.py`；训练 / 蒸馏样本格式已落成 `contracts/copilot-training-sample.schema.json`、`scripts/checks/fixtures/copilot-training-sample-basic.json` 与 `scripts/check-copilot-training-sample-contract.py`，并已补 `scripts/build-copilot-training-samples.py` 将三条主任务各 3 条 committed eval golden_response 样本和各 3 条 audit pass candidate record 转换成 JSONL 训练样本且接入 `check-repo`；当前还补 `training/README.md`、`training/datasets/README.md`、`training/datasets/copilot-training-dataset-governance-v0.json`、`training/datasets/copilot-training-review-record-v0.json` 与 `training/datasets/copilot-training-holdout-split-v0.json`，明确 JSONL 默认是 `tmp/` 下的本地生成产物，入仓优先采用 manifest、summary、复核记录和实验说明，并把更大集合的入选条件、抽样复核比例、holdout、质量门禁、退场条件、planned 复核模板和非重叠 holdout split 接入仓库级检查；`RadishMind-Image Adapter` 也已把第一版图片生成意图落成 `contracts/image-generation-intent.schema.json`、`scripts/checks/fixtures/image-generation-intent-basic.json` 与 `scripts/check-image-generation-intent-contract.py`。后续只有这些基础资产可复跑后，才进入微调、蒸馏或量化实验。

### 退出标准

- `RadishMind-Core` 在目标任务上达到可对比的基础表现，并能通过统一协议、评测和规则校验稳定替换 teacher / mock provider
- 可以明确知道下一步应继续训练、补数据、优化工具路由，还是调整对照模型规模
- 图片生成能力具备最小可演示链路：`RadishMind-Core` 输出结构化生成意图，`RadishMind-Image Adapter` 调度 backend，结果以 artifact 形式返回；主模型不直接生成图片像素

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

## M7：Agent / Model 边界拆分评估

### 目标

当 `RadishMind` 已经同时具备稳定 agent runtime、真实多项目接入和可比较的 student/model 路线后，评估是否需要把 agent / copilot runtime 与模型训练、模型服务、权重资产拆成独立仓库或独立发布单元。

这个节点不是为了提前拆仓库，而是为了避免两类风险：

- agent runtime、协议、adapter、工具和评测仍在快速共振时过早拆分，导致边界漂移
- 模型权重、训练产物、推理镜像和服务部署变重后仍强行塞在同一仓库，导致协作、CI、权限和发布节奏失控

### 触发条件

满足以下多数条件时，应正式进入拆分评估：

- `RadishFlow` 与 `Radish` 至少各有一个生产前稳定任务接入同一 agent runtime
- `CopilotRequest` / `CopilotResponse`、artifact、risk、confirmation、citation 等核心协议连续多个迭代保持兼容
- 至少有一条任务完成 teacher / student / mock 或多 provider 的可重复对照，证明模型已是可替换组件，而不是写死在任务逻辑里
- `RadishMind-Core` / `minimind-v` 或其它 student 路线进入持续训练、蒸馏、量化、评测发布节奏，且产生不适合直接入仓的大体积权重或训练产物
- agent runtime 的发布节奏开始与模型训练 / 推理服务发布节奏明显不同
- 模型服务需要独立 GPU / 推理基础设施、安全凭据、镜像发布或部署权限
- 仓库 CI 因模型资产、训练脚本、推理镜像或大数据集显著拖慢，影响协议、adapter 和评测的日常协作

### 拆分候选形态

优先考虑“先拆发布单元，再拆仓库”。如果同仓库仍能保持清晰边界，可以先只拆目录、包和 CI job。

若确实需要拆仓库，建议目标形态为：

- `RadishMind`：保留 agent runtime、contracts、adapters、prompts、eval、candidate records、工具编排和规则校验
- `RadishMind-Models`：承接训练配置、模型适配、蒸馏脚本、量化脚本、模型评测报告和权重发布元数据
- 外部模型 / 对象存储：承接大体积权重、训练产物、推理镜像和不可直接 committed 的数据资产

### 退出标准

- 能明确判断继续单仓库、拆包不拆仓库、或拆成多个仓库的收益与成本
- 如果拆分，协议真相源、评测入口、模型版本引用和发布节奏都有明确归属
- 拆分后不破坏 `RadishFlow` / `Radish` 的统一协议和任务级回归

## 当前推荐顺序

建议严格按以下顺序推进：

1. `M0`
2. `M1`
3. `M2`
4. `M3`
5. `M4`
6. `M5`
7. `M6`
8. `M7`

不要在 `M1` 之前就直接开大规模训练，也不要在没有评测基线前频繁切换底座模型。
不要把 `RadishMind-Core` 误收口成从零预训练的大模型项目；当前合理路线是 `3B` / `4B` 起步、`7B` 作为本地上限，并通过独立 backend 提供图片生成能力。
不要在 `M6` 之前急于把 agent / model 拆成多仓库；在协议、adapter、评测和至少两个项目的任务接入稳定前，保持单仓库更利于收口。

## 当前优先推进

### 阶段判断

当前仓库已经可以开始推进下一个大节点，不宜继续把 `RadishFlow / suggest_flowsheet_edits` 当作唯一主战场深挖。

按现有正式口径判断：

- `suggest_flowsheet_edits` 已完成 `33/33` 离线样本的四主 `apiyi_cx / apiyi_cc / apiyi_ch / apiyi_de` 横向真实覆盖
- `default` teacher 对照也已从旧 early pool 推进到更高价值 sample pool，并正式补齐 `mixed-risk-patch-combo`、`triad-mixed-risk-cross-object`、`mixed-risk-cross-object`、`cross-object-citation` 与 `range-sequence-ordering`
- `suggest_ghost_completion` 已具备真实 batch 入仓、dump 重导和正式 audit 链，并已从固定 trio 扩到十六批高价值链式真实样本，当前覆盖为 `78/78`
- `suggest_flowsheet_edits` 已接上 `artifact summary`、same-sample / cross-sample replay、recommended replay summary 与 real-derived negative，治理链已不再缺基础连通性
- `suggest_ghost_completion` 已进一步接上 same-sample / cross-sample replay、两路 recommended replay summary 与首批 real-derived negative，当前已不再阻塞于 `missing_real_derived_negative_samples`，且最近的 residual `v21` 到 `v25` 正式真实 capture 未再暴露新的 runtime 根因
- `Radish docs QA` 已把 same-sample / cross-sample replay、recommended replay summary 与 real-derived negative index 接进仓库级治理链

因此当前更合理的阶段判断是：

- `M2` 对 `RadishFlow` 的核心 PoC 能力已达到“可真实 capture、可正式回归、可沿 dump 重导治理”的退出门槛
- 仓库主线应从“单任务持续扩样”切到 `M3：数据集与评测体系`
- `suggest_flowsheet_edits` 后续仍继续推进，但应从“回补 teacher comparison”转成“非重复高价值真实样本池扩样”支线，而不是继续占据整个项目节奏

### 接下来两周主线

接下来两周建议把仓库主线从“继续跑样本”切到“`M3` 退出收口 + `M4` 前置准备”。由于 `RadishFlow` 与 `Radish` 暂时都还没有进入真实模型 / Agent 接入阶段，当前不应继续把“真实上层调用位置”作为阻塞项，也不应继续扩展本仓库内的模拟 UI / 命令层 summary。

1. 继续保持 `M3` 的退出门槛：三条真实治理链保持可机读、可复跑、无基础资产缺口；新 batch 只在有明确非重复 drift 假设时触发
2. 把已落地的 gateway demo manifest / summary fixture 作为 `RadishFlow suggest_flowsheet_edits` 服务/API 改动的默认门禁，后续任何 gateway metadata、错误 envelope 或确认语义改动都必须同步更新该 summary 或解释兼容性
3. 冻结当前进程内 gateway、UI consumption summary 与 candidate edit handoff summary，把它们视为未来上层接入验收门禁；在上层项目未准备好前，不继续新增同类模拟 summary
4. 转入 `M4` 前置准备：基座评估矩阵、离线评测阈值、离线评测运行记录、训练样本格式、eval sample 转换入口、audit pass candidate record 转换入口、更大训练集合治理 manifest、planned 复核记录模板和非重叠 holdout split 已先落成仓库级契约，下一步补 teacher / student / lightweight baseline 的实际评测结果、成本观测和真实复核结果
5. 继续沿 `contracts/image-generation-intent.schema.json` 与最小图片生成意图 fixture 推进 `RadishMind-Image Adapter`，下一步再补 backend 抽象和图片生成评测样本，使图片生成能力先以 adapter / backend 形式站住，而不是并入主模型参数目标
6. 只有当服务/API、模型评测或后续真实接入暴露现有样本无法覆盖的新失败面时，才回到真实 capture 扩样；扩样完成后必须同步更新治理报表和周志

### 大节点切换条件

当前已经满足“开始推进下一个大节点”的判断条件：

- 至少两条 `RadishFlow` 核心任务已完成真实 provider capture 与正式 audit 链接线
- `suggest_flowsheet_edits` 不再只有单边 profile 或少量早期样本，而已形成可复用的 teacher 批次治理方式
- 仓库现阶段的主要风险，已从“PoC 能否站住”转成“评测、negative、replay 和 profile coverage 能否统一治理”

因此从当前提交点开始，项目主线应正式按 `M3` 来组织，而不是继续把阶段判断停留在 `M2` 的单任务推进视角。

在正式进入实现期前，当前建议按以下顺序继续推进：

1. 把 `M3` 退出标准固定到文档和治理报表：当前阶段的主线目标是“评测资产可作为实现门禁”，不是继续增加真实 batch 数
2. 继续收紧 `services/gateway` 的最小职责：接收 `CopilotRequest`、路由到现有 runtime、统一错误响应、写入可审计 metadata，并保持 provider profile 可配置；当前最小 Python 骨架已落地，下一步不再停留在职责定义，而应补调用形态和错误消费约定
3. 继续以 `RadishFlow suggest_flowsheet_edits` 做服务/API 最小切片，因为它已有最完整的结构化建议、风险确认、citation、replay、real-derived 基线和 gateway demo summary 门禁
4. 把现有 `scripts/run-copilot-inference.py`、`scripts/check-gateway-service-smoke.py`、`scripts/run-radishflow-gateway-demo.py`、`scripts/check-radishflow-gateway-ui-consumption.py` 与 `scripts/check-radishflow-candidate-edit-handoff.py` 能验证的路径收束成服务级 smoke 矩阵：同一类请求应能走 CLI、gateway 函数入口、UI consumption summary、candidate handoff summary 与未来 API 入口，并落到兼容的 `CopilotResponse` / `CopilotGatewayEnvelope`
5. 明确服务层不负责直接写回业务真相源；`candidate_edit` 只能作为 UI 可审查提案，所有实际修改仍交由上层项目命令层和人工确认；当前 `CopilotGatewayEnvelope` 调用口径已在集成契约中补齐，应继续作为后续 HTTP/API 包装的语义真相源
6. 在 `RadishFlow` / `Radish` 未准备接入模型或 Agent 前，停止继续深挖真实上层调用位置；保留现有 smoke 作为未来验收门禁即可
7. 继续维护 `Radish docs QA` 与 `suggest_ghost_completion` 的治理链，但不主动扩样；只有服务/API 集成、模型评测或新产品场景暴露评测缺口时再补真实 capture
8. 从当前点开始并行进入 `M4 RadishMind-Core / minimind-v` 的模型对照与训练路线验证，先做基座评估、样本格式和 image adapter schema，不直接启动重训练

## 当前仍缺的关键决策

- `Qwen2.5-VL` 的具体首选尺寸与推理预算
- `SmolVLM` 的具体回归任务边界与保留条件
- `RadishMind-Core` 首版基座评估矩阵、阻塞阈值与离线评测记录格式已固定 `3B` 优先、`4B` 对照、`7B` 延后；具体观测结果和成本实测仍需后续补齐
- `RadishMind-Image Adapter` 的第一版 schema 已落到仓库级门禁；backend 抽象和最小评测样本如何继续接入同一契约链
- `RadishFlow` 截图路线的进入时点
- 评测样本的标注和维护流程
- `M7` 到来时，agent runtime 与 model / training / inference service 是继续同仓库分包，还是拆成独立仓库与独立发布单元
- `Radish` docs QA 的真实 captured negative 批次采用目录清单、导入脚本还是两者并行维护
- `Radish` docs QA 的 real-derived negative pattern 是否继续只靠 `capture_metadata.tags` 约定，还是需要升级成更正式的结构化字段
- `Radish` docs QA 的 `violation_groups` 是否要长期保持“按完整违规文案分组”，还是后续补一层更轻量的 violation pattern 归一化

## 后续可探索方向

以下内容当前只作为远期备忘，不纳入当前 `M3 / M4 / M5` 主线，也不改变当前阶段顺序、退出标准和资源分配。

### `RadishFlow`

- 在 `suggest_ghost_completion` 已覆盖的 ghost 拓扑补全基础上，后续可探索更完整的流程链补全，使“放入一个模块后继续预测下一步流程结构”成为独立方向
- 在 `suggest_flowsheet_edits` 已覆盖的局部参数修正基础上，后续可探索面向典型设备的参数优化建议，例如塔、分离器、换热器等对象的经验性优化候选
- 在现有诊断与候选 patch 之外，后续可探索“参数输入错误提示 / 参数组合不合理预警 / 输入后纠偏辅助”这类更前置的编辑器智能校验能力

### `Radish`

- 在 `answer_docs_question`、`summarize_doc_or_thread` 与 `suggest_forum_metadata` 之外，后续可探索社区小助手方向，包括审贴辅助、回复草稿、互动建议和运营协作
- 在保持人工确认的前提下，后续可探索自动回复建议、社区互动建议和有限的运营流程辅助，但不把它们提前写成当前阶段承诺
- 对“代操作账号”这类更强自治方向，仅保留为更远期探索项；若未来进入评估，必须先补强权限边界、审计记录、风险分级与显式确认链路
