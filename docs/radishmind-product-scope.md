# RadishMind 产品范围与目标

更新时间：2026-06-07

## 核心定义

`RadishMind` 是 `Radish` 体系下的 AI 工具、工作流、模型网关和 Copilot 集成平台。

更强的项目定义、三层定位、northbound/southbound 战略、service mode 分级与 action safety ladder 见 [战略定义](radishmind-strategy.md)。

它不是上层业务真相源，不是单一大模型仓库，也不只是等待真实接入的中转站。它的职责是把用户端 AI 应用、管理端治理、模型 API 分发、工作流运行、上层项目上下文、局部规则、工具能力、模型推理和审计治理收口为同一条可复跑的运行链路。

当前核心边界：

- 读状态、读文档、读附件和可选图像，输出解释、诊断、结构化建议和候选动作。
- 高风险动作必须保留 `requires_confirmation`，由人工确认或上层规则层复核后再执行。
- 对内保持统一 canonical protocol，对外兼容常见模型调用协议和常见 AI 服务协议，而不是被单一厂商接口绑死。
- 模型负责理解、推理、归纳、排序和建议生成；runtime、adapter、tooling、rule validation 和 audit 负责上下文打包、工具调用、结构校验、权限边界和可追溯性。
- `RadishMind-Core` 是基座适配型自研主模型路线，不是从零预训练基础大模型。
- 图片像素生成不并入主模型职责，默认由 `RadishMind-Image Adapter` 与独立 backend 承接。
- 部署方式、数据库选型、登录 / 授权边界优先参考 `Radish`；未来 RadishMind 作为 OIDC client 接入 `Radish`，不自建第二套身份真相源。参考 `Radish` 不代表默认引入 `.NET` / ASP.NET Core；RadishMind 后端默认继续使用 `Go` 承载 control plane / gateway / API 服务，`Python` 只保留在模型、评测和 AI 生态强相关链路，`TypeScript/Vite` 承载前端。
- `RadishFlow` 和 `Radish` 是优先接入对象与产品参考，但不是 RadishMind 平台本体开发的阻塞条件。上层暂时没有稳定 UI、command 或 API 挂载点时，本仓库应继续推进可离线验证、可复用到后续真实接入的用户端、workflow runtime、control plane 和模型网关功能；不把等待上层接线写成产品停滞理由。

## 产品形态

长期产品形态按四个一级面组织：

2026-05-27 已新增 [Control Plane / User Workspace / Workflow v1 计划](task-cards/control-plane-user-workspace-workflow-v1-plan.md)，用于固定四个产品面的 v1 服务边界、数据边界和停止线；`product-surface-v1-boundary` 已进一步把 `User Workspace`、`Admin Control Plane`、`Model Gateway / API Distribution` 和 `Workflow / Agent Runtime` 的资源、读模型和写边界写入可检查 fixture；`control-plane-data-boundary` 已固定 tenant、user、role、permission、provider profile、model route、quota、price、audit、secret ref 与 deployment status 的 ownership；`radish-oidc-client-preconditions` 已固定 issuer、client、claim mapping、tenant binding、logout、audit 和 failure taxonomy；`gateway-api-key-quota-readiness` 已固定 API key、quota、rate limit、cost ledger 和 trace 前置条件；`workflow-definition-run-record-boundary` 已固定 workflow definition、run record、状态流转、失败分类、审计证据和停止线；`control-plane-read-model-v1` 已固定 tenant summary、application summary、API key summary、quota summary、workflow definition summary、run record summary 和 audit summary 的只读 read model；`control-plane-read-route-contract-v1` 已固定 `GET /v1/user-workspace/runs` 等七类 tenant-scoped read-only route contract；`control-plane-read-response-fixtures-v1` 已固定 response fixture、统一 envelope、`failure_code` 和脱敏输出；`control-plane-read-negative-contract-v1` 已固定负向契约、forbidden method / query / fallback、敏感字段投影拒绝和 fail-closed 输出；`control-plane-read-fake-store-handler-implementation-v1` 已实现七条 fake-store-backed read route，覆盖 tenant summary、applications、api-keys、quota summary、workflow definitions、runs 与 audit；`control-plane-read-auth-db-preconditions-v1` 已固定真实 auth/db 前置条件、future auth middleware 和 future read store repository；`control-plane-read-consumer-contract-v1` 已固定 TypeScript consumer contract 和离线消费 smoke；`control-plane-read-formal-ui-boundary-v1` 已固定正式 UI 边界、页面到 read route 的分配、只读状态和敏感字段停止线。它不代表正式用户端、生产管理端、workflow executor、API key / quota、数据库 read path 或 Radish OIDC 已实现。

2026-05-28 已新增 `control-plane-read-formal-ui-implementation-readiness-v1`，固定未来 `apps/radishmind-web/` 预留落点、`apps/radishmind-console/` app 边界、页面实现顺序、consumer contract 复用、测试策略和停止线。

2026-05-31 已创建 `apps/radishmind-web/`，作为正式产品 UI 的 read-only product shell 首个实现落点。当前 shell 默认只消费 `contracts/typescript/control-plane-read-api.ts` 中的离线 view model，已包含 route catalog、共享状态组件、forbidden output guard、只读 `admin-tenant-overview`、`admin-audit-log`、`workspace-applications`、`workspace-api-keys`、`workspace-usage-quota`、`workspace-workflow-definitions` 与 `workspace-run-history` 页面切片。2026-06-01 已补显式 opt-in 的 dev-only live read consumer：只有设置 `VITE_RADISHMIND_READ_SOURCE=dev-live-http`，且后端设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 时，页面才通过 HTTP 消费 fake-store-backed read handlers 和测试身份上下文。该路径不接生产后端、不接数据库、OIDC、repository、API key lifecycle、quota enforcement、billing、workflow executor、confirmation、writeback 或 replay；`apps/radishmind-console/` 仍只是本地 ops surface。

当前产品 UI 的门禁策略已经从普通展示页逐项专项证明，调整为能力边界与聚合门禁优先。`control-plane-read-formal-ui-readiness-close-v1` 已用 surface matrix 聚合固定七个页面的 route binding、状态预览、request / audit ref 和 forbidden output guard；`control-plane-read-auth-store-transition-preconditions-v1` 已固定从 dev fake auth / fixture-backed fake store 迁移到未来 auth middleware / read store repository 前必须满足的 gates。上述内容都不能解释为真实数据库、Radish OIDC、production API consumer、API key / quota、repository implementation 或 workflow executor ready。

read store 的产品范围现在已经明确为“先固定未来迁移契约，再实现真实持久化”。`control-plane-read-repository-contract-smoke-v1`、`control-plane-read-repository-implementation-readiness-v1`、`control-plane-read-store-selection-readiness-v1` 和 `control-plane-read-schema-migration-readiness-v1` 只说明未来七条 read route 如何从 fake store 迁移到 repository/database：输入输出、tenant context、失败映射、no fake fallback、no side effects、schema ownership 和 migration smoke 都必须先可检查。它们不把 read-side 页面升级成 production API consumer，也不实现 SQL、migration、repository adapter、真实数据库、Radish OIDC、token validation、API key lifecycle、quota enforcement 或 workflow executor。

1. `User Workspace`

- 面向终端用户和项目成员。
- 支持创建 AI 应用、Prompt 应用、Workflow、Agent / Copilot 应用、RAG 或知识问答应用。
- 用户可以管理自己的应用、API key、调用量、运行记录和成本摘要。
- 当前 `apps/radishmind-web/` 只提供 read-side 页面切片：applications、API keys、usage quota、workflow definitions 和 run history 默认都是离线只读展示；dev-only live read path 也只能读取 fake-store-backed handler，不提供创建、编辑、执行、replay 或写回控件。
- 工作流方向参考 `Dify` 的应用构建与 workflow 编排，但首版只实现 Radish 体系当前需要的可治理切片，不追求一次性复刻全量能力。

2. `Admin Control Plane`

- 面向平台管理员和运维。
- 管理租户、用户、角色、权限、模型供应商、provider profile、模型路由、API key、额度、价格、审计、secret backend 和部署状态。
- 认证、授权、数据库、部署和运维习惯优先对齐 `Radish`；未来通过 OIDC 接入 `Radish` Auth。
- Control Plane 可以拆成独立 Go 服务，但不因为职责扩张而引入新后端语言或塞进 gateway 单体。
- 当前 `apps/radishmind-web/` 只提供只读 `admin-tenant-overview` 和 `admin-audit-log` 页面切片；它不是 production admin console，也不提供 audit mutation、raw payload export、durable audit store 或生产管理操作。

3. `Model Gateway / API Distribution`

- 面向 API 调用者和上层服务。
- 提供 OpenAI-compatible / Responses / Messages / Models 等 northbound API，统一分发到多 provider、多 profile 和多模型。
- 支持后续的 API key 分发、配额、限流、成本统计、trace、fallback / load balancing 和 provider health。
- 模型 API 分发方向参考 `sub2api` 与 `axonhub`，但必须保留本仓库的 provider registry、审计和生产停止线。

4. `Workflow / Agent Runtime`

- 面向 AI 应用执行。
- 承载 Prompt、LLM、HTTP tool、RAG retrieval、condition、output、后续受控 code / sandbox 与 agent loop。
- 每次运行都应有 trace、输入输出摘要、成本、错误分类和风险边界。
- `workflow-definition-run-record-boundary` 只把 workflow definition、run record、node execution、tool audit、result materialization、confirmation decision、状态流转、失败分类、审计证据和停止线固定为治理证据，不代表 executor、confirmation、writeback 或 replay 已实现。
- `Workflow / Agent Runtime Function Surface v1` 已把现阶段可推进功能面限定为 application detail、workflow definition detail、run detail、tool action preview、confirmation placeholder、offline draft designer 和 offline validation inspector 的只读 / blocked / local-only surface，优先走 fixture 或 fake-store dev path。`workflow-draft-validation-inspector-offline-v1` 已固定为 `workflow_draft_validation_inspector_offline_defined`，只展示 selected draft 的 validation summary、structural checks、contract checks、blocked capability checks 和 route / request / audit metadata；不提供 draft persistence、validation result persistence、publish、executor、confirmation decision、writeback 或 replay。
- 上层挂载点未成熟时，workflow 产品面继续先做离线草案设计、结构检查、execution plan preview、readiness 展示和 blocked capability 说明；这些产品能力应复用未来真实接入所需的 canonical contract 和停止线，而不是等待 `RadishFlow` 或 `Radish` 提供承接入口后才开始。
- 高风险 tool/action 默认 `requires_confirmation`，不得直接写上层业务真相源。

## 项目范围

### 1. `Runtime Service / Model Gateway`

- 提供最小可运行的推理入口、gateway、route 识别、provider/profile 选择、响应封装和本地产品 discovery 面。
- 当前已有 CLI runtime、进程内 Python gateway、最小 `Go` HTTP bridge、本地启动 runbook、`GET /v1/platform/overview` 只读产品 overview、`GET /v1/platform/local-smoke` 本地 readiness 摘要、session/tooling metadata shell、blocked action shell、`apps/radishmind-console/` 本地 console 壳、Dev Diagnostics、`Local Readiness` 面板、Provider/Profile Details、Stop-line Details、overview / local-smoke failure surface、console behavior / visual smoke record / dev entry / production boundary gate、P3 checklist、Docker local compose、测试 / 生产共用部署态 compose、镜像命名治理、deployment readiness 静态 smoke、container smoke runbook、运行记录模板、一次 `docker_local` container smoke 运行记录、provider capability matrix、provider health smoke、provider selection policy 和 provider runtime docs refresh；本地只读产品壳已达到 `local usable / read-only close`，Docker 静态部署边界已可检查，本地 mock 容器 smoke 已跑通，Provider Runtime & Health v1 进入 close candidate。真实镜像发布 workflow、production secret backend、optional live health、真实 retry/fallback、测试环境 smoke、生产前复核和生产部署边界只在明确运行窗口或独立任务卡下推进，服务层与 `gateway` 不锁死在单一语言上，可按职责采用 `Go`。
- 这一层必须同时覆盖两条方向：
  - 北向协议兼容：对外提供 native Copilot API，以及 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models`、`/v1/platform/overview`、`/v1/platform/local-smoke`、session/tooling metadata shell 这类常见兼容接口和只读产品发现接口。
  - 南向模型接入：对内接入 `RadishMind-Core`、`local_transformers / HuggingFace`、`Ollama`、OpenAI-compatible、Gemini native、Anthropic messages 等 provider / transport。

### 2. `Conversation & Session`

- 管理 `conversation_id`、会话上下文、历史压缩、恢复和审计边界。
- 当前已有 session contract、history/state policy、checkpoint metadata-only read route 和 `GET /v1/session/metadata`；这些只支撑 metadata / overview 展示，不代表 durable session store、长期记忆或 replay executor 已启用。

### 3. `Tooling Framework`

- 承载检索、附件解析、项目语义转换、本地候选生成、response builder 与工具策略。
- 当前已有通用 tool contract、registry、audit、`GET /v1/tools/metadata` 和 blocked `POST /v1/tools/actions`；它们只支撑 contract-only 展示和 blocked action response，不代表真实工具执行器、confirmation flow 或业务写回已启用。

### 4. `Evaluation & Governance`

- 负责 schema、smoke、offline eval、candidate record、review、promotion gate 和仓库级检查。
- 目标不是只看机器指标，而是让 advisory-only、confirmation、route、citation 和 review 边界能长期复跑。

### 5. `Model Adaptation`

- 负责基座选型、prompt/runtime 协同、蒸馏、训练样本治理、模型晋级与回归。
- 当前已形成 P4 v1 前置证据：1.5B raw 在 docs / ghost 上可跑但 edits blocked，repaired comparison 只作后处理证据，3B CPU 单样本 timeout。真实模型产出、3B/4B 长跑、训练 JSONL、蒸馏和权重相关工作转入后置专题。
- 自研模型只是平台的一类 provider，不应和 `HuggingFace`、`Ollama`、OpenAI-compatible 或其它外部模型接入能力互相替代。

### 6. `Image Path`

- 主模型只输出结构化 image intent、约束、审查和 artifact metadata。
- 真正的图片生成由独立 image adapter 和 backend 承接。

### 7. 用户端、管理端和上层项目接入面

- 用户端用于 AI 工作流、应用、模型 API key、调用量和运行记录。
- 管理端用于 provider/profile、模型路由、租户、权限、quota、secret、审计和部署状态。
- 当前 `apps/radishmind-console/` 只是本地 ops surface 和只读产品壳，不等同于正式用户端或生产管理端。
- 当前 `apps/radishmind-web/` 是正式产品 UI 的 read-side shell，默认离线，可显式 opt-in 到 dev-only live read；它不等同于完整用户端、production admin console 或真实 API consumer。

- `RadishFlow`、`Radish`、`RadishCatalyst` 是应用面，不是项目本体的全部意义。
- 这些接入面复用同一套 runtime、contract、tooling、evaluation 和 governance，而不是各自私接模型。
- 这些接入面若暂时无法提供真实挂载点，不应拖慢 RadishMind 的平台功能建设；本仓库应先把成熟的离线产品面、协议、风险边界和验证基线做好，等上层条件满足后再选择一个切片真实接入。

## 当前阶段判断

- 历史上的 `M3` service/API smoke 与 `M4` broader review、`3B/4B` capacity review 已经收口为冻结证据。
- 当前正式主线切换为“AI 工具 / 工作流 / 模型网关 / Copilot 集成平台重定义 + 平台基础能力建设”，不再把“继续深挖同一批实验”或“提前设计不存在的真实接线”当作默认推进方式。
- 当前 `P3 Local Product Shell / Ops Surface` 的本地只读产品壳已收口为 `local usable / read-only close`：已用 `/v1/platform/overview`、`/v1/platform/local-smoke`、overview / local-smoke consumer smoke、最小本地 console 壳、Dev Diagnostics、`Local Readiness` 面板、Provider/Profile Details、Stop-line Details、overview / local-smoke failure surface、console behavior / visual smoke record / dev entry / production boundary gate 和 P3 checklist 固定本地 console 可展示能力与未满足的生产前置条件。`Production Ops Hardening v1` 已进一步固定 Docker local/test/prod 部署形态、compose 边界、镜像命名、静态 smoke、runbook 和运行记录模板，并完成一次 `docker_local` container smoke；`Provider Runtime & Health v1` 已固定 capability / health / selection / docs 四个可检查切片并进入 close candidate。后续只有在明确测试或生产前复核窗口后才执行测试环境 smoke 或 production preflight；无运行窗口时，默认转向 `Workflow / Agent Runtime Function Surface v1` 以及离线 workflow 产品功能，而不是继续补同类只读 console 小切片、provider 同层小切片、Production Ops 静态治理、重开真实模型长跑或提前设计不存在的上层接线。
- 训练 / 蒸馏样本继续只提交 manifest、summary、复核策略和实验说明；生成的 JSONL 和真实模型产物默认留在 `tmp/`。

## 当前优先支持的应用面

### `RadishFlow`

首批能力继续围绕结构化流程状态：

- `FlowsheetDocument + SelectionState + DiagnosticSummary` 的解释与诊断问答。
- 基于选中对象的候选编辑提案。
- 基于本地合法候选集的 ghost port、ghost connection、ghost stream name 补全建议。
- 求解状态、控制面、entitlement、package sync 与离线授权摘要解释。

当前不可侵入：

- 数值求解热路径。
- CAPE-OPEN / COM 适配边界。
- token、credential、auth cache 和包体完整性真相。
- 未经确认直接改写 `FlowsheetDocument`。

### `Radish`

首批能力继续围绕知识与内容：

- 固定文档、在线文档、论坛内容和 Console 权限知识的检索增强问答。
- 文档、帖子、评论和附件的摘要、标题、标签、分类与引用辅助。
- 当前页面、路由、角色和权限摘要解释。

当前不可侵入：

- 身份、token 生命周期和权限最终判定。
- 附件访问控制和临时访问令牌。
- 未经确认直接执行治理、封禁、授权或数据写入。

### `RadishCatalyst`

当前只做文档级预留：

- 可预留玩家知识问答、进度解释、生产链规划和开发侧静态数据一致性检查。
- 不扩真实 schema、adapter、gateway smoke 或模型接线，直到上层明确首个真实任务面。

当前不可侵入：

- Godot 运行时、存档、任务、战斗、掉落、配方、公开等级和联机权威。
- 玩家侧不能泄露 `internal` 或默认隐藏的 `spoiler` 内容。

## 当前非目标

- 不把 `RadishMind` 做成替代业务内核的自治系统。
- 不把当前仓库写成已经具备完整 Dify / sub2api / axonhub 等同能力的产品。
- 不让模型替代 `RadishFlow` 求解、`Radish` 权限判定或 `RadishCatalyst` 游戏权威。
- 不把通用 unrestricted tool calling 当成当前默认能力。
- 不把平台锁死在单一模型、单一 provider、单一上游协议或单一对外接口上。
- 不自建与 `Radish` 冲突的用户身份、权限和部署真相源；未来 RadishMind 应作为 OIDC client 接入 `Radish`。
- 不把“参考 Radish”解释成复制 Radish 后端语言栈；除非未来必须直接复用 Radish 后端包或共发布，否则不新增 `.NET` 作为默认后端栈。
- 不默认继续真实模型产出、3B/4B 长跑、训练 JSONL、蒸馏或权重相关工作；这些内容后续作为独立专题重开。
- 不默认下载大模型、数据集或权重。
- 不把 `14B/32B` 写成当前自研主模型默认目标；首版仍优先本地可承受的小中型路线，长期本地部署上限暂定 `7B`。

## 实现原则

- 协议优先采用结构化 JSON。
- 能规则化或工具化的逻辑，不强行压给模型。
- 先把 runtime、session、tooling、governance 边界做清楚，再扩大训练和接入面。
- 代码遵循对应语言的惯用实践，命名清楚、职责明确、边界稳定。
- 禁止语义不明的方法、空转 helper、过度泛化的 manager/factory 和晦涩抽象封装。
- 抽象必须服务于真实职责边界、复用或复杂度收敛；不能为了隐藏简单逻辑而增加理解成本。
