# RadishMind 系统架构草案

更新时间：2026-04-20

## 架构目标

`RadishMind` 的目标架构建议冻结为六层：

1. Client Adapters & Context Packers
2. Copilot Gateway / Task Router
3. Retrieval & Tool Layer
4. Model Runtime Layer
5. Rule Validation & Response Builder
6. Data / Evaluation / Training Pipeline

这个拆分的核心目的，是让不同上层项目通过统一协议接入，同时把“项目语义”“工具编排”“模型推理”“安全校验”和“评测闭环”解耦。

## 总体工作流

当前建议按以下顺序理解一次完整请求：

```text
上层项目状态 / 文档 / 附件 / 图像
        ↓
Adapter 打包为统一 CopilotRequest
        ↓
Copilot Gateway 识别任务与项目
        ↓
Retrieval / Tools 获取补充证据并压缩上下文
        ↓
LLM / VLM 生成解释、问题与候选动作
        ↓
Rule Validation 校验风险、目标和结构
        ↓
Response Builder 输出可消费 JSON
        ↓
Adapter 映射回各自 UI / 日志 / 候选提案
```

## 分层说明

### 1. Client Adapters & Context Packers

面向上层项目的集成适配层。

当前建议至少预留：

- `adapter-radishflow`
- `adapter-radish`

职责：

- 将业务上下文转换为统一 `CopilotRequest`
- 对敏感字段做裁剪、脱敏和禁止透传
- 处理认证、会话、超时、重试、缓存和本地回填
- 将结构化响应映射回各自项目的 UI、日志或候选编辑提案

项目重点：

- `RadishFlow`
  - 当前仓库已先落最小 `adapter-radishflow` 骨架：`adapters/radishflow/request_builder.py` 与 `scripts/build-radishflow-request.py` 可把上游快照稳定装配为 `CopilotRequest`，并已对齐六条既有 eval sample，覆盖最小样本、控制面冲突态与多选裁剪态
  - 当前还已补 export -> adapter 的中间转换层：`adapters/radishflow/export_snapshot.py` 与 `scripts/build-radishflow-adapter-snapshot.py` 可先把更贴近真实导出对象的嵌套快照收口成 adapter snapshot，避免 adapter 直接绑定到手工拼装的扁平 fixture
  - 当前还已补 export -> request 的直达入口：`scripts/build-radishflow-export-request.py` 可让更贴近真实导出对象的输入直接落到 runtime 请求，避免端到端链路只靠“两段脚本各自正确”来间接保证
  - `RadishFlowExportSnapshot` 当前已在集成契约中补上字段映射约定：上游负责直接导出 `document_state / selection_state / diagnostics_export / solve_session_state / solve_snapshot / control_plane_snapshot`，而 adapter 不再擅自反推 `selected_unit` 或提前替上游做根因归一
  - 当前还已补 exporter bootstrap 入口：`scripts/init-radishflow-export-snapshot.py` 可按任务生成最小 schema-valid 模板，作为真实接线前的起步骨架
  - 当前还已补 exporter preflight 入口：`scripts/validate-radishflow-export-snapshot.py` 可在真实接线前先做 schema、任务级语义与敏感透传 smoke 校验，再进入 export -> adapter -> request 正式链路
  - 当前还已补 exporter batch smoke 入口：`scripts/run-radishflow-export-smoke.py` 可对一组 committed export snapshot 批量执行 `validate -> adapter -> request`，并用结构化 summary 固定“真实 exporter 接线最小入口”是否仍保持稳定
  - 当前 exporter fixture 已继续补到更贴近真实接线的边界：联合选择态下 `primary_selected_unit` 与完整 selection 并存、`multi-unit + multi-stream + 单 primary focus` 组合仍稳定映射成 `selected_unit + 完整 selection` 且不重排 selection 顺序、更复杂 selection chronology 下主焦点与 chronology 解耦但仍只保留单 actionable target、同风险多动作并列时仍稳定优先补输入完整性、`support_artifacts` 既覆盖纯 `uri + metadata.summary` 的最小脱敏摘要，也覆盖 `attachment_ref + json summary + text note` 与 `sanitized_summary + summary/redaction_summary + 最小 json rollup + text note` 两档 mixed summary 组合，以及多对象 selection 下三动作优先级顺序不漂移
  - 对 `RadishFlow suggest_flowsheet_edits`，当前除 response-level regression 外，也已补任务 prompt、最小 runtime 与正式 `candidate_response_record -> manifest -> audit` 批次入口：`scripts/run-radishflow-suggest-edits-poc-batch.py` 现已把真实主线推进到 `v81`、累计五十八批正式真实批次，且现有 `33/33` 条离线样本都至少已有一条正式真实覆盖；其中 `mixed-risk / citation / reconnect`、ordering 尾样、`mixed-risk patch combo`、cross-object primary-focus、parameter / patch ordering 与 local-edits 六组样本族都已完成 `apiyi_cx / apiyi_cc / apiyi_ch / apiyi_de` 四主 profile 的横向正式收口，`default` teacher 对照也已继续补到 `range-sequence-ordering`。当前 `apiyi_ch` 在 triad mixed-risk cross-object 上的正式口径也已收紧为“仅该两条样本使用样本级 `210s` timeout override”，而不是抬高 profile 全局默认超时；这条主线的下一步则已转向 `cross-object-primary-focus` 这类仍停留在 `default-only` 的高价值样本池，而不再继续围绕已收口样本族反复兜底
  - 优先走状态优先打包：`FlowsheetDocument`、`document_revision`、`SelectionState`、`DiagnosticSummary`、`SolveSessionState`、`SolveSnapshot`
  - 对编辑器辅助场景，额外打包 `selected_unit`、`unconnected_ports`、`nearby_nodes`、`cursor_context` 与本地规则筛出的 `legal_candidate_completions`
  - `cursor_context.recent_actions` 当前不仅承接“最近 accept 了哪条 ghost”，也承接“最近 reject / dismiss / skip 了哪条 ghost”；其第一版正式语义已收口为三条链式模板共享的同一套规则：只压制同一 `candidate_ref` 的下一帧默认 `Tab`，隔一帧且候选仍是高置信合法默认项时允许恢复，而不同 `candidate_ref` 不共享 suppress 信号
  - 控制面相关只打包 entitlement / manifest / lease / sync 的摘要，不透传 token 或 credential
- `Radish`
  - 优先走知识优先打包：固定文档、在线文档、论坛/文档 Markdown、Console 权限知识、附件引用
  - 角色和权限只传最小必要摘要，不透传原始 token、cookie 或安全凭据

不负责：

- 训练模型
- 持有业务真相源
- 直接执行未经确认的破坏性修改

### 2. Copilot Gateway / Task Router

作为 `RadishMind` 的统一入口服务。

职责：

- 接收多模态请求
- 识别 `project`、`task`、`schema_version`
- 路由任务到对应的 retrieval / tool / model 组合
- 统一输出结构化响应
- 提供鉴权、审计、请求追踪与版本标记

建议接口风格：

- HTTP JSON 为主
- 图片和大文件通过对象引用或附件引用传递
- 输出保持统一骨架，但允许项目专属上下文字段

### 3. Retrieval & Tool Layer

这里承载“模型之外的确定性能力”。

职责：

- 文档检索
- 项目语义转换
- 附件 / Markdown / JSON 解析
- 候选动作构建
- 调用外部工具或项目专属只读解析器

当前建议至少拆出两类工具：

- 证据获取工具
  - 读取固定文档、在线文档、论坛正文、附件摘要、结构化状态
- 候选动作工具
  - 在不直接写入业务真相源的前提下，生成可确认的候选编辑或操作提案
  - 为编辑器辅助场景先由本地规则生成合法 ghost completion 候选集，再交给模型做排序、命名和空结果判断
- 对 `RadishFlow suggest_ghost_completion`，本地规则层还应显式输出 recent-actions 反馈衍生出的 suppress-Tab 信号，并把“same-candidate 即时 suppress / 一帧 cooldown 恢复 / latest-action precedence / other-candidate 不共享 suppress”作为可审计的结构化口径，而不是只留在提示词或前端隐式逻辑里

原则：

- 能规则化的逻辑，不强行让模型背
- 能提前压缩的上下文，不把全部原始状态直接丢给模型
- 任何高风险动作都只能输出候选动作，不直接执行
- 对 `RadishFlow` ghost completion 这类编辑器辅助任务，应先由本地规则系统裁剪到“合法候选空间”，再让模型排序，而不是直接从零猜拓扑
- 对同一 ghost candidate 的最近接受/拒绝/关闭/跳过反馈，优先由适配层与本地规则层编码成结构化 recent-actions 和 conflict/suppress 信号，再决定是否仍允许 `Tab`
- 对 `RadishFlow suggest_ghost_completion` 当前这条 editor assist 主线，recent-actions 相关判定不应只在单模板成立，而应在 `Feed -> Valve -> FlashDrum`、`Feed -> Heater -> FlashDrum` 与 `Feed -> Cooler -> FlashDrum` 三条链式模板上保持对称

### 4. Model Runtime Layer

这里承载真正的 LLM / VLM 推理。

建议分为两类：

- Teacher Models
  - 用于强能力推理、数据蒸馏、标注参考和 PoC 对照
- Student Models
  - 用于本地化、小成本部署和项目内推理实验

当前判断：

- `RadishFlow` 第一阶段不应把全部任务都压成截图推理；结构化状态和诊断解释优先
- `Radish` 第一阶段以文档、内容和 Console 知识问答为主，VLM 只在附件或截图理解场景补充
- `minimind-v` 当前作为默认 `student/base` 主线，承接领域适配、训练实验与后续部署路线
- `Qwen2.5-VL` 当前作为默认 `teacher` / 多模态强基线，优先承担复杂图文任务 PoC、标注参考与蒸馏输入
- `SmolVLM` 当前作为轻量本地对照组，优先承担低资源回归与部署下限比较

### 5. Rule Validation & Response Builder

这里承载模型输出后的硬约束和结构收口。

职责：

- 校验响应结构、字段完整性和版本兼容
- 检查目标对象、风险等级和确认要求
- 过滤不允许的动作或越权建议
- 生成统一引用、证据和 `requires_confirmation`

当前原则：

- 模型输出默认是建议，不直接成为最终状态
- 高风险动作必须带 `requires_confirmation`
- 若模型证据不足，应允许退化为检索或模板式回答

### 6. Data / Evaluation / Training Pipeline

职责：

- 构建训练样本
- 管理合成数据与人工校正数据
- 执行离线评测
- 跟踪模型版本、任务表现和回归

第一阶段优先建立：

- `RadishFlow`
  - 基于 `FlowsheetDocument`、选择集、诊断摘要、求解快照和控制面错误摘要的合成样本
  - 后续再补基于真实画布的截图样本
- `Radish`
  - 基于固定文档、在线文档、论坛 Markdown、Console 文档和附件引用的样本
- 共享指标
  - 结构合法率
  - 引用命中率
  - 建议可执行率
  - 风险分级正确率
  - 任务级回归稳定性

当前评测管线还应统一支持以下输入与回放形态：

- `golden_response` 对照
- 样本内 `candidate_response`
- 外部 `candidate_response_record`
- 与正例共用同一套规则的负例回放
- 通过 `capture_metadata` 标记来源批次的真实 captured replay
- 对真实/模拟 batch 同时沉淀 same-sample / cross-sample replay index、recommended replay summary 与 artifact summary，避免 batch 审计、推荐回放和 committed fixture 漂成三套口径
- 对 `RadishFlow` 当前这两条已接线任务，也应先承认治理成熟度分层：`suggest_flowsheet_edits` 当前已接上最小 `artifact summary`、same-sample replay 与 recommended replay summary，但仍缺 cross-sample replay 与 real-derived；`suggest_ghost_completion` 则已接上最小 `artifact summary`、same-sample / cross-sample replay、两路 recommended replay summary 与首批 real-derived negative。后续统一治理时不应再把两条链与 `Radish docs QA` 误判成同一成熟度，也不应继续把它们笼统归类成“只到 artifact summary”
- 对基于真实 bad record 派生的本地负例，生成独立 real-derived negative index，并按 `source_record_groups`、`violation_groups`、`pattern_groups` 做结构化审计
- 对 repeated real-derived pattern，优先在索引层保留“source 维度”和“pattern 维度”两个视角，而不是一开始就把所有违规文本做重度归一化
- 当前 `Radish docs QA` 的 same-sample / cross-sample replay 与 real-derived negative index 已统一纳入 `check-repo`，且 `2026-04-05` real batch 已无 singleton source；这条治理链的后续重点应转向跨 source 复合 drift、真实 captured 扩充，以及 `pattern` / `violation` 结构化升级时机评估
- 对 `RadishFlow suggest_ghost_completion`，评测当前不应只停在“同一 candidate 刚被 reject / dismiss / skip 后不立即 retab”，还应继续覆盖 same-candidate 一帧 cooldown 恢复 `Tab`、latest-action precedence 下的 reject / dismiss / skip 恢复态、other-candidate 不共享 suppress，以及多动作 recent-actions 交错下的恢复窗口
- 对 `RadishFlow suggest_ghost_completion`，除 response-level regression 外，当前还应允许把外部 `candidate_response_record` 回灌到同一条 audit / regression 链；仓库内已先落一条 3 样本的轻量 `capture -> manifest -> audit` PoC，且真实 batch 已统一迁到 `datasets/eval/candidate-records/radishflow/batches/YYYY-MM/<batch_key>/` 短路径布局，用于把 editor assist 任务从“只有 fixture”推进到“可真实捕获并治理候选输出”；其中当前新增观察项已分成五层：批处理编排下的 provider 卡顿稳定性、`manual_only` 多动作输出的结构坏法、供应商级限流或路由不可用时通过备用 profile 继续保持真实 capture 连续性、同 provider 备选模型虽可调用但仍可能出现不可正式导入的任务质量漂移，以及不同 provider 输出风格漂移下的字段文本归一化
- 对 `RadishFlow suggest_flowsheet_edits`，评测管线还应把响应稳定性当作一等能力校验，而不只检查字段存在：至少需要显式覆盖 `issues`、顶层 `citations`、`issues[*].citation_ids`、`candidate_edit` 动作顺序、`candidate_edit.citation_ids` 以及 `patch` 内部多层键/数组的稳定顺序
- 对 `RadishFlow suggest_flowsheet_edits`，除顺序回归外，当前也已允许把外部 `candidate_response_record` 回灌到同一条 audit / regression 链；仓库内除 `2026-04-12-radishflow-suggest-edits-poc-mock-v1` 这批 3 样本 mock PoC 外，还已把真实主线推进到 `v81`、累计五十八批正式真实 teacher 批次，并全部接入 `check-repo`。其中现有 `33/33` 条离线样本都至少已有一条真实批次覆盖，且 `mixed-risk / citation / reconnect`、ordering 尾样、`mixed-risk patch combo`、cross-object primary-focus、parameter / patch ordering 与 local-edits 六组样本族都已完成四主 profile 的横向正式收口；下一步则转向仍停留在 `default-only` 的高价值样本池，并以 `cross-object-primary-focus` 作为当前第一优先级，而不是继续在已闭环样本族上做重复扩样
- 对 `RadishFlow suggest_flowsheet_edits`，task-level canonicalization 当前也已正式承接几类只在真实 teacher 输出中暴露出来的任务漂移：`flowdoc-*` 引用已从“按完整 `flowsheet_document` 原始对象顺序稳定编号”进一步收紧为“按当前响应实际纳入的目标对象做紧凑稳定编号”，`flow_rate` / `flow_rate_kg_h` / `mass_flow_kg_per_h` / `mass_flow_kg_h` / `mass_flow_rate_kg_h` 等近义占位会统一归一，`outlet_temperature_target_c` / `outlet_temperature_target` / `target_outlet_temperature_c` 与 `operating pressure` / `pressure target` 这类 message 线索也会回收到稳定参数名，路径式 placeholder 会回收到稳定字段名，多规格 `spec_placeholders` 会稳定收口到 `temperature_c -> pressure_kpa -> flow_rate_kg_per_h`，`STREAM_DISCONNECTED` 与 `STREAM_SPEC_MISSING` 等 error issue 的 `issues[*].citation_ids` 会与 `diag -> artifact -> snapshot` 或 `diag -> artifact -> supporting artifact -> snapshot` 正式口径对齐，reconnect patch 里的 `connection_placeholder` 会稳定补回 `retain_existing_source_binding=true`，跨对象真实样本中的 issue/action/top-level citation 交错顺序也会统一收口到 `diagnostic -> target artifact -> supporting artifact -> snapshot`。同时，若真实输出出现多余闭合 brace、未转义中文引号、错误 `flowdoc-*` 引用编号、warning citation 漂移、同 target warning 遗漏到首条 action citation 组、placeholder-only patch 误保留 `parameter_updates`、或 `efficiency_percent` review range 偏离既有 `[60, 85]` 正式口径，也会优先在 provider/runtime 层做窄范围修复并回收到当前样本的 canonical 输出；`apiyi_ch` 在 triad 路径上的超时治理也已收紧为样本级 `210s` override，而不是 profile 全局默认超时提升

## 当前文件与脚本组织约定

- committed `Python` 与 `JSON` 文件默认不超过 `1000` 行；超过时优先按职责拆分，而不是继续堆长
- 大型 committed JSON 索引优先采用“主索引 + `.parts/` 分片文件”方式收口，避免单个长数组文件失控增长
- `scripts/` 根目录优先只保留稳定入口与平台包装；较长实现、内部 helper 与静态 fixture 应放进浅层分类子目录
- 当前脚本目录已先收口为 `scripts/checks/` 与 `scripts/eval/` 两个浅层分组：
  - `scripts/checks/` 承接仓库检查实现和静态 fixture
  - `scripts/eval/` 承接评测 runner 的共享实现与任务级校验
- 后续若还需按项目拆分脚本，优先继续新增同层级目录，而不是把嵌套层级继续拉深
- committed 相对路径默认应控制在 `180` 个字符以内；接近预算时，优先缩短目录语义或回收元数据，而不是继续拉长物理路径
- `datasets/eval/candidate-records/radishflow/` 当前执行更严格的专项预算：committed 文件路径默认不超过 `120` 个字符，且根目录只保留 `README.md`、`batches/` 与 `dry-run-check/`
- `RadishFlow` candidate-records 的正式物理布局固定为 `batches/YYYY-MM/<batch_key>/`，批次内只保留 `manifest.json`、`audit.json`、`artifacts.json`、`r/`、`o/` 与 `d/` 这些短结构名
- `datasets/eval/candidate-records/radish/` 当前仍处于旧布局过渡期，committed 文件路径默认不超过 `178` 个字符；后续新增或迁移 `Radish` 批次也应采用同类短键布局，而不是继续扩张旧长路径
- `RadishFlow` 与 `Radish` 评测资产的长语义都应优先沉淀到 `manifest`、`record`、fixture 与索引文档中，而不是重复编码进目录层级、文件名前缀或 sample slug

## 推荐仓库结构

建议先采用如下结构：

```text
RadishMind/
├─ README.md
├─ docs/
├─ contracts/
├─ services/
│  ├─ gateway/
│  ├─ orchestrator/
│  ├─ tools/
│  └─ evaluation/
├─ adapters/
│  ├─ radishflow/
│  └─ radish/
├─ datasets/
│  ├─ synthetic/
│  ├─ annotated/
│  └─ eval/
├─ training/
│  ├─ configs/
│  ├─ scripts/
│  └─ notebooks/
├─ prompts/
│  ├─ system/
│  ├─ tasks/
│  └─ eval/
└─ experiments/
```

## 关键设计口径

- 统一先走结构化协议，不让不同项目各自发散
- 协议采用“通用骨架 + 项目专属上下文块”，而不是强行做一个业务超集
- 当前主实现栈收口为 `Python`，优先统一评测、数据处理、模型适配和自动化校验工具链
- `RadishMind` 的长期架构应按“受控 agent / copilot runtime + 可替换 model runtime”理解：agent 层负责任务路由、上下文选择、工具调用、规则校验、风险确认和响应组装，model 层只负责可替换的推理能力
- 当前不把 agent runtime 与模型训练 / 模型服务拆成不同仓库；只有当双项目接入、协议稳定、student 训练和推理部署都进入独立节奏后，才按路线图 `M7` 评估是否拆包或拆仓库
- `RadishFlow` 优先走状态优先上下文，截图是补充，不是第一阶段唯一中心
- `Radish` 优先走知识优先上下文，重点是 Docs / Wiki / Forum / Console 语义
- 模型输出默认是建议，不直接成为最终状态
- 训练数据优先从自有项目生成，不依赖大量外部脏数据
- 评测必须从第一阶段就开始建立
- 对 `suggest_flowsheet_edits` 这类 advisory patch 任务，评测不仅要检查“能不能答”，还要检查同一输入下的结构化顺序是否稳定，避免候选动作、证据引用和 patch 细节在回归中随机漂移
- 部署形态允许本地、局域网和远端三种模式并存，但当前不先锁定最终部署形态
- 对 `Radish docs QA`，真实 captured negative、same-sample replay、cross-sample replay 与 real-derived negative 应保持同一条治理链，而不是分别演化成互不对齐的 fixture 体系
- 对 `Radish docs QA` 的 real-derived 扩样，当前继续优先复用既有 `derived_pattern` 与 `expected_candidate_violations`，避免为单次样本扩张过早引入新的 `pattern_group` / `violation_group`

## 当前最小实现补充

当前仓库已补第一条最小实现链路，用于把“只有评测资产”的状态推进到“能实际吐出结构化响应并最小落盘 capture”的状态：

- `services/runtime/inference.py`
  - 提供 `radish / answer_docs_question` 与 `radishflow / suggest_ghost_completion` 的最小 runtime
  - 内置 `mock` provider，用于工程闭环验证
  - 预留 `openai-compatible` provider，用于真实模型接入
- `scripts/run-copilot-inference.py`
  - 允许从 `datasets/eval` 样本或独立 `CopilotRequest` 运行最小推理
  - 可直接写出 normalized `CopilotResponse` 与 raw dump
- `scripts/run-radishflow-ghost-real-batch.py`
  - 为 `RadishFlow suggest_ghost_completion` 提供 3 样本轻量 PoC 批次入口
  - 串起 `capture -> manifest -> audit` 的最小闭环，默认覆盖 `Tab / manual_only / empty`
  - 若未显式提供 `--output-root`，默认直接写入 `datasets/eval/candidate-records/radishflow/batches/YYYY-MM/<batch_key>/`
- `scripts/import-candidate-response-dump-batch.py`
  - 将一批 raw dump 正式导入为仓库内 `candidate_response_record`
  - 支持对 canonicalization 修复前采集的旧 dump 按当前 runtime 重新归一化后再生成正式 `manifest` 与 `audit`
- `prompts/tasks/radish-answer-docs-question-system.md`
  - 冻结当前单任务系统提示的最小口径
- `prompts/tasks/radishflow-suggest-ghost-completion-system.md`
  - 冻结 ghost completion 的最小任务提示口径，并限制模型只能从 `legal_candidate_completions` 中选候选

这一步仍不是完整服务化实现，但已经把“prompt 组装 -> provider 调用 -> 响应归一化 -> raw dump 落盘 / 最小批次审计”这条链路落地，可作为后续 gateway、adapter 和真实模型 provider 的起点。
