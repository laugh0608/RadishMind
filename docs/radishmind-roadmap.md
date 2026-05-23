# RadishMind 阶段路线图

更新时间：2026-05-23

## 路线图原则

路线图只记录阶段目标、当前进度、下一步和停止线。批次细节、历史失败、完整实验输出和长命令不放在本入口文档中，应进入周志、实验 manifest、run record 或任务卡。

当前长期目标保持不变：`RadishMind` 是受控 Copilot / Agent runtime platform + 可替换模型能力，不是单一万能模型。

若要理解“为什么路线这样排”，先读 [战略定义](radishmind-strategy.md)。

## 当前路线切换

从 2026-05-10 开始，仓库主线正式从“围绕 `M3/M4` 收口继续做局部维护”切换为“基于已收口证据继续建设平台本体”。

当前已经冻结的历史证据：

- `M3`：gateway、service smoke、UI consumption 与 candidate handoff 已收口为服务/API 门禁。
- `M4`：broader 15 样本人工复核为 15/15 `reviewed_pass`，`3B/4B` guided capacity review 已收口为正式审计记录。

这些资产继续保留，但不再等同于当前唯一主线。

## 五条主线

### 1. `Runtime Service`

目标：把现有 CLI runtime、进程内 gateway、route 识别和 smoke gate 收口为明确的 provider registry、协议兼容层、本地运行、配置、启动和部署基础。

状态：`scripts/run-copilot-inference.py`、`services/gateway/copilot_gateway.py`、`services/runtime/inference_provider.py`、`services/runtime/provider_registry.py`、`services/platform/`、`RadishFlow` gateway demo 与 service smoke matrix 已具备基础骨架；当前 southbound 已通过统一 registry 收口 `mock`、`openai-compatible`、`HuggingFace`、`Ollama` 主入口与 `openai-compatible chat`、`gemini-native`、`anthropic-messages` 分流，`local_transformers` 则主要存在于 candidate/runtime 实验链路中。平台表层语言分工已固定为 `UI=React + Vite + TypeScript`、`Platform Service Layer=Go`、`Model Side=Python`。当前 `Go` 层已落最小服务启动、`/healthz`、`/v1/models`、`/v1/chat/completions`、`/v1/responses` 和 `/v1/messages` bridge，并补了第一版 SSE 流式兼容骨架、bridge-backed provider/profile inventory、`GET /v1/models/{id}` 精确 lookup、request-side provider/profile 选择、流式增量转发、`HuggingFace` / `Ollama` coverage。平台级 `ops smoke` 已固定 `go test ./...`、provider registry 与受控 profile inventory 门禁；本地启动 runbook、runbook drift check、脱敏配置摘要 / config check、JSON 配置文件层级、稳定本地启动 wrapper、最小 deployment smoke、结构化 diagnostics/failure boundary、provider/profile discoverability 对齐、request-level observability 与 error taxonomy 已补齐。`P1 Runtime Foundation` 已达到 short close；第一版 northbound 仍是窄切片，但继续横向扩同层配置、别名和兜底的收益已经下降，主要实现重心切到 `P2 Session & Tooling Foundation`。

下一步：不再继续把 `P1` 做成无限硬化阶段；`P2 Session & Tooling Foundation` 已有 metadata / blocked 产品外壳，主线切到 `P3 Local Product Shell / Ops Surface`。

### 2. `Conversation & Session`

目标：让多轮对话、历史压缩、恢复和审计成为平台能力，而不是各任务自己拼上下文。

状态：已补首版 `session-record.schema.json`、`session-recovery-checkpoint.schema.json`、`session-recovery-checkpoint-manifest.schema.json`、`session-recovery-checkpoint-read.schema.json`、fixture 和快速门禁，并让 `Go` northbound 兼容层在显式 `radishmind` 会话扩展存在时写入 `context.northbound.session`；`state_policy` 已固定会话状态与 tool result cache 的 v1 落点只允许 northbound metadata / session recovery checkpoint，不启用 durable memory；recovery checkpoint v1 只保存 request/session/tool audit/tool metadata 引用，read result 只暴露 metadata refs 和 tool audit 治理摘要，不保存或返回真实工具结果，也不自动 replay。平台层已新增 metadata-only route smoke，并通过 denied query fixture 拒绝 materialized result、result ref、executor ref、durable memory 与 replay 类查询参数；readiness summary、implementation preconditions、negative regression skeleton、governance-only `session-tooling-negative-regression-suite.json`、`session-tooling-negative-regression-suite-readiness.json`、deny-by-default implementation gates、`session-tooling-negative-coverage-rollup.json`、`session-tooling-route-negative-coverage-matrix.json`、`session-tooling-route-smoke-readiness-rollup.json`、`session-tooling-short-close-readiness-delta.json`、`session-tooling-upper-layer-confirmation-flow-readiness.json`、`session-tooling-short-close-entry-checklist.json`、`session-tooling-readiness-consistency-rollup.json`、`session-tooling-executor-storage-confirmation-enablement-plan.json`、`session-tooling-stop-line-manifest.json`、confirmation flow design、independent audit records design、result materialization policy design、executor boundary design、storage backend design、`session-tooling-foundation-status-summary.json` 和 `session-tooling-close-candidate-readiness-rollup.json` 已把当前状态收口为 `close candidate / governance-only`，并把到 `P2 short close` 仍缺的硬前置条件与 future route / gate smoke 要求标为 `not_satisfied`；upper-layer confirmation readiness 已把 handoff contract、decision binding、negative gate consumers 和 confirmed action boundary 收口为接线前证据清单；entry checklist 已把 stop-line manifest、short close delta、route smoke readiness 和 suite readiness 聚合为进入条件预检；route negative coverage matrix 当前只声明 2 个 suite case 被 checkpoint read metadata-only route 覆盖，另 7 个仍需要 future route requirement；stop-line manifest 已明确真实 executor、durable store、confirmation 接线、materialized result reader、长期记忆和 replay 仍 blocked；当前仍没有 durable session store、durable checkpoint store、durable audit store、durable result store、长期记忆、真实 checkpoint storage backend 或跨轮恢复执行器，也不声明 P2 short close。

下一步：保持 P2 close-candidate readiness 口径可检查，但不再主动扩新的 readiness、rollup、manifest 或 task card；优先把现有 session metadata 作为 P3 本地产品面的一部分复用。在 short close 前置条件满足前，不进入真实实现设计。

### 3. `Tooling Framework`

目标：把检索、局部规则、候选生成和 builder 经验收口为正式工具契约、registry、policy 和 audit。

状态：当前已有 task-local 的 deterministic tooling 与 builder 资产；最小 `tool.schema.json`、`tool-registry.schema.json`、`tool-audit-record.schema.json`、registry fixture、policy/audit fixture 和快速门禁已开始落地，用于固定工具注册、调用轨、timeout/retry/policy、session binding、metadata-only result cache 和 audit 的结构边界。tool audit summary 已进入 checkpoint read route smoke，用于固定 execution disabled、not executed、metadata-only cache 和 no result ref；session/tooling promotion gate 分层、负向消费 summary、route smoke coverage summary、readiness summary、implementation preconditions、negative regression skeleton、governance-only negative regression suite、`session-tooling-negative-regression-suite-readiness.json`、deny-by-default implementation gates、`session-tooling-negative-coverage-rollup.json`、`session-tooling-route-negative-coverage-matrix.json`、`session-tooling-route-smoke-readiness-rollup.json`、`session-tooling-short-close-readiness-delta.json`、`session-tooling-upper-layer-confirmation-flow-readiness.json`、`session-tooling-short-close-entry-checklist.json`、`session-tooling-readiness-consistency-rollup.json`、`session-tooling-executor-storage-confirmation-enablement-plan.json`、`session-tooling-stop-line-manifest.json`、confirmation flow design、independent audit records design、result materialization policy design、executor boundary design、storage backend design、close candidate status summary 与 close-candidate readiness rollup 已进入快速门禁。当前仍没有真实工具执行器、durable audit store、durable result store、长期记忆或新的 provider/model 实验，negative skeleton、governance suite、suite readiness、deny-by-default gate contract、negative coverage rollup、route negative coverage matrix、route smoke readiness rollup、short close delta、upper-layer confirmation readiness、entry checklist、readiness consistency rollup、enablement plan、stop-line manifest、audit design、materialization policy design、executor boundary design 和 storage backend design 也不等同于完整 `negative_regression_suite`。

下一步：继续守住 tooling contract、audit、result materialization、executor boundary 与 storage backend 的设计停止线；相关 metadata / blocked shell 只作为 UI 设计与只读消费输入复用。在上层确认流接线和完整负向回归满足前，不启动真实执行。

### 4. `Evaluation & Governance`

目标：让 runtime、session、tooling、deployment 和 model adaptation 都有统一门禁，而不是只校验模型输出。

状态：schema、offline eval、candidate record、review record、`check-repo`、service smoke 和 runtime provider dispatch smoke 已具备基础；平台级 smoke 已继续扩展到 runtime config/deployment/diagnostics/request observability 与 P2 session/tooling/checkpoint governance。当前 `session-tooling-foundation-status-summary.json`、`scripts/checks/fixtures/session-tooling-upper-layer-confirmation-flow-readiness.json` 与 `scripts/checks/fixtures/session-tooling-short-close-entry-checklist.json` 只声明 `close candidate / governance-only`、upper-layer confirmation readiness 和 entry checklist 可检查，不声明实现完成。

下一步：维持 advisory-only、confirmation、route、citation、handoff 不执行和 metadata-only 这些不变量；完整负向回归和真实实现 gate 必须先于 executor/storage/confirmation 实现落地。

### 5. `Model Adaptation`

目标：在平台契约稳定后，再定义首版基座、蒸馏和训练升级路径。

状态：raw、repair、injection、guided、task-scoped builder、offline eval 和 training sample conversion 已有资产，但当前还不具备“直接扩大训练规模”的时机。

下一步：先以平台契约为前提锁定 v1 训练目标，再决定新的实验或蒸馏路线；没有新能力假设前，不继续重跑同一批 `M4` 实验。

## 辅助支线

### `Image Path`

状态：intent、backend request、artifact schema 与最小评测 manifest 已具备；真实 backend 仍未接入。

下一步：继续收口 image adapter handshake、safety gate 和 artifact 返回链路，不下载模型、不生成图片。

### 上层项目接入

状态：`RadishFlow` 门禁已冻结，`Radish` docs QA 资产已具备，`RadishCatalyst` 仍只做文档预留；三个上层项目当前都不具备真实接入能力。

下一步：先推进平台本体；待上层具备真实挂载点、确认流和命令承接接口后，再只选一个切片真实接入。

### `UI Design Topic / Pencil Draft`

状态：`close candidate`。`docs/designs/radishmind-console-ops-surface-v0.pen` 已覆盖 7 个主要页面并通过 Pencil layout 检查；`apps/radishmind-console/` 第二批 React 已重排为浅色侧栏、主工作区和 readiness / stop-line 辅助栏结构，并完成本地 mock platform ready 态桌面 / 窄屏临时截图复核。它仍不等同于 production console 或 production packaging。

触发条件：已经满足并完成首轮 close candidate。后续只在真实使用暴露新可读性缺口时继续做 UI polish；外部参考素材见 [UI 设计参考](radishmind-ui-design-reference.md)。

停止线：不把当前本地 console 壳写成 production console，不提前实现复杂交互、生产导航、确认流或业务写回 UI。后续任何 UI 能力扩张仍必须先回到设计稿或任务卡说明范围。

## 阶段顺序

### `P0`：项目重定义与能力盘点

目标：把“项目到底是什么、有哪些主线、哪些能力缺口最关键”写成正式文档和能力矩阵。

状态：已完成平台重定义、能力矩阵和主线切换；后续只维护文档口径，不再作为主要实现阶段。

### `P1`：Runtime Foundation

目标：收口最小本地 service bootstrap、provider registry、northbound/southbound 协议兼容、配置、调用和 smoke 路径。

状态：short close。provider registry、Go service bootstrap、northbound bridge、provider/profile discoverability、config layering、wrapper、deployment smoke、diagnostics、request observability、error taxonomy 与三种 northbound 协议的 selection metadata smoke 已进入平台单元测试和快速门禁。

### `P2`：Session & Tooling Foundation

目标：补齐 conversation/session contract、tool contract、registry、policy 和审计轨。

状态：`close candidate / governance-only`，并已具备可消费的 metadata / blocked 外壳。session contract、history policy、state policy、recovery checkpoint、tool schema、tool registry、tool policy、audit record、northbound session metadata、`GET /v1/session/metadata`、`GET /v1/tools/metadata` 与 `POST /v1/tools/actions` 已能支撑上层或 UI 展示 session/tool metadata 和 blocked action。既有 readiness、rollup、matrix、enablement plan、stop-line manifest 与 entry checklist 继续作为停止线证据保留；当前仍不声明 P2 short close，也不具备真实 executor、durable storage、上层 confirmation flow 接线、materialized result reader、durable audit store、durable result store 或完整 `negative_regression_suite`。

停止线证据仍以 governance-only fixture 保留：`session-tooling-foundation-status-summary.json`、`session-tooling-negative-regression-suite-readiness.json`、`session-tooling-close-candidate-readiness-rollup.json`、`session-tooling-route-negative-coverage-matrix.json`、`session-tooling-short-close-entry-checklist.json` 等继续固定 `P2 short close` 前的 `not_satisfied` 条件和 `negative_regression_suite` 边界；这些文件不再作为默认新增工作方向。

### `P3`：Local Product Shell / Ops Surface

目标：让本地长驻服务、启动说明、只读 console、观测、故障边界和 ops surface 具备正式口径。

状态：`local usable / read-only close`。本地治理第一版已具备 wrapper、配置文件层级、deployment smoke、启动前 diagnostics、runbook drift check、`GET /v1/platform/overview` 只读产品 overview、`GET /v1/platform/local-smoke` 本地 readiness 摘要、overview / local-smoke consumer smoke 和 `apps/radishmind-console/` 本地 console 壳；console 当前已补一键 dev 启动/验证入口、refresh 状态、Dev Diagnostics、`Local Readiness` 面板、Provider/Profile Details 只读详情、Stop-line Details 只读详情、overview / local-smoke failure surface、连接失败诊断、更可读的 overview 展示、`scripts/check-radishmind-console-behavior.py` 行为门禁、`scripts/check-radishmind-console-visual-smoke-record.py` 视觉 smoke 记录门禁和 console production packaging 边界门禁。`scripts/checks/fixtures/p3-local-product-shell-short-close-checklist.json` 已把本地只读产品面标为可用，同时把 production hardening 固定为 `not_ready`：production secret backend、process supervisor、deployment environment isolation 和 console production packaging 仍为 `not_satisfied`。后续不再默认继续补同类只读 console 小切片，真实使用暴露缺口时再补。

### `P3 衔接专题`：UI Design Topic / Pencil Draft

目标：在基础平台和本地只读产品壳足够稳定后，先用 `pencil` 完成 UI 信息架构和界面设计稿，再进入正式 UI 实现。

状态：`close candidate`。当前已有 [UI 设计参考](radishmind-ui-design-reference.md)、[UI 设计规范](radishmind-ui-design-spec.md)、`.pen` 设计稿和第二批 React ops surface 结构重排；后续不再默认扩当前 console 小功能。

进入条件：已满足。P3 的 overview、local-smoke、Dev Diagnostics、只读失败态、Provider/Profile Details、Stop-line Details 和 P3 checklist 已足以说明真实界面要承载哪些状态；同时 production packaging、supervisor、secret backend、confirmation flow 等边界仍清楚标记为未完成。

退出条件：已达到 close candidate。核心页面、状态层级、只读/可执行边界、错误诊断、窄屏布局和 React 第二批实现切片已收口；后续只在真实使用暴露问题时做定向修正。

### `P4`：Model Adaptation & Training

目标：在平台边界稳定后，定义首版基座、蒸馏和训练升级计划。

状态：正在进入前置计划定义，但不提前放量。v1 模型能力目标、teacher/student 边界、样本分层、晋级门槛、预检 runbook、治理复核记录和 `Qwen2.5-1.5B-Instruct` full-holdout-9 预检结果已有首版记录。raw student 在 docs QA 与 ghost completion 上通过，但在 `suggest_flowsheet_edits` 上 3/3 blocked；repaired comparison 可通过机器门禁，但只作为后处理证据。不启动大规模训练、不下载模型权重、不把 builder / guided / repaired 结果写成 raw 晋级。

### `P5`：Real Upstream Integration

目标：在上层项目具备真实挂载点后，选择首个切片完成真正接入。

状态：当前等待上层条件成熟。

## 下一步

1. 推进 P4 模型适配前置计划：基于 `Qwen2.5-1.5B-Instruct` raw blocked / repaired pass 结论，决定是否在同一 full-holdout-9 边界上比较更大 raw 模型；不训练放量、不下载权重。
2. 将 `P3 Local Product Shell / Ops Surface` 与 UI 第二批维持在 `local usable / read-only close candidate`；不再默认补同类只读 console 小切片，除非真实使用暴露新缺口。
3. UI 后续扩张必须先回到设计稿或任务卡，不直接增加 confirmation、writeback、replay 或 production packaging。
4. 只为新增 API、执行边界、生产声明、数据格式、外部 provider 风险或高风险能力新增专项门禁；普通 UI 展示改动优先复用现有 console behavior / visual smoke / fast baseline。
5. 继续维持上层项目接入前置条件总表，不提前细化不存在的真实接线。

## 停止线

- 不把 repaired、injected、guided 或 builder 轨通过写成 raw 模型能力通过。
- 不把机器指标通过写成人工可接受度通过。
- 不在没有非重复能力假设时继续扩同一批 `M4` 实验。
- 不在上层项目没有真实挂载点时继续细化假想接线设计。
- 不把 `P1` 继续扩成无止境的 provider/config/diagnostics 细化阶段。
- 不把 P3 继续扩成无止境的本地只读 console 小切片阶段。
- 不把当前本地 console 壳扩成 production console、大面积复杂交互或真实确认 / 写回 / replay UI。
- 不让模型直接写上层业务真相源。
- 不用晦涩抽象、空泛 helper 或多层 fallback 掩盖代码职责不清。
