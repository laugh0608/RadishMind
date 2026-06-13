# RadishMind 系统架构

更新时间：2026-06-10

## 架构目标

`RadishMind` 的正式架构目标不再只是“把单次模型推理接上去”，而是建设 `Radish` 体系下可本地运行、可审计、可工具化的 AI 工具、工作流、模型网关和 Copilot 集成平台。

这套平台当前从三个视角理解：

- 产品视角：看用户端、管理端、模型网关和 workflow runtime 怎么分层
- 平台视角：看五条主线怎么协同
- 请求视角：看单次 `CopilotRequest -> CopilotResponse` 是怎么流动的

2026-05-27 已新增 [Control Plane / User Workspace / Workflow v1 计划](task-cards/control-plane-user-workspace-workflow-v1-plan.md)，只固定四个产品面的服务边界、数据边界和停止线；`product-surface-v1-boundary` 已用 `product-surface-v1-boundary.json` 与 `check-product-surface-v1-boundary.py` 固定资源、读模型、写边界和 shared stop-line；`control-plane-data-boundary` 已固定 control plane 对象 ownership，并继续要求 Radish remains identity truth；`radish-oidc-client-preconditions` 已固定未来 OIDC client 的 issuer、client、claim mapping、tenant binding、logout、audit 和 failure taxonomy；`gateway-api-key-quota-readiness` 已固定 API key、quota、rate limit、cost ledger 和 trace 前置条件；`workflow-definition-run-record-boundary` 已固定 workflow definition、run record、状态流转、失败分类、审计证据和停止线；`control-plane-read-model-v1` 已固定 tenant-scoped 只读 read model、访问策略和脱敏策略；`control-plane-read-route-contract-v1` 已固定 read-only route contract、分页 / 过滤和 fail-closed 失败边界；`control-plane-read-response-fixtures-v1` 已固定 response fixture、`failure_code` 和脱敏输出；`control-plane-read-negative-contract-v1` 已固定 negative contract、forbidden method / query / fallback、敏感字段投影拒绝和 fail-closed envelope；`control-plane-read-implementation-preconditions-v1` 已固定未来 read route 的 handler ownership、fixture-backed fake store、auth middleware dependency、response conformance 和 negative route smoke readiness；`control-plane-read-fake-store-handler-plan-v1` 已固定 future handler 仍落在 `services/platform/internal/httpapi`；`control-plane-read-fake-store-handler-implementation-v1` 已在 `services/platform/internal/httpapi` 落地七条 fake store backed read route；`control-plane-read-auth-db-preconditions-v1` 已固定 future Radish OIDC / auth middleware 与 future control plane read store repository 的准入条件；`control-plane-read-consumer-contract-v1` 已固定 `contracts/typescript/control-plane-read-api.ts` 和上层 TypeScript consumer contract；`control-plane-read-formal-ui-boundary-v1` 已固定正式 UI 边界、页面到 read route 的分配、只读状态和敏感字段停止线；`control-plane-read-formal-ui-readiness-close-v1` 已用 surface matrix 聚合收口七个只读页面；`control-plane-read-dev-live-consumer-v1` 已提供显式 opt-in 的 dev-only HTTP consumer；`control-plane-read-auth-store-transition-preconditions-v1` 已固定未来 auth/store 迁移前置条件。这一层不实现 OIDC、token validation、数据库、repository migration、repository implementation、API key / quota、workflow executor、confirmation、writeback 或 replay。

兼容旧有 read-side 停止线表述：这一层不实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-formal-ui-implementation-readiness-v1` 已进一步固定正式 UI 实现前的工程边界：product UI 落点为 `apps/radishmind-web/`，当前 `apps/radishmind-console/` 仍是本地 ops surface；页面实现顺序、consumer contract 复用和测试策略已可检查。当前已创建 `apps/radishmind-web/` 的 read-only shared shell，并实现只读 `admin-tenant-overview`、`admin-audit-log`、`workspace-applications`、`workspace-api-keys`、`workspace-usage-quota`、`workspace-workflow-definitions` 与 `workspace-run-history` 页面切片。这些页面默认只消费离线 TypeScript view model；显式 `dev_live_http` 仅用于本地开发消费 fake-store-backed Go read route 和测试身份上下文，不接生产后端、数据库、OIDC、executor、confirmation、writeback 或 replay。

当前 read-side UI 的架构门禁调整为能力边界优先：普通只读展示页由聚合 surface matrix / checker 收口，不再默认逐页新增专项门禁；dev-only live read consumer 只证明 HTTP consumer shape；auth/store transition preconditions 只定义未来迁移 gates。真实 Radish OIDC / auth middleware、repository interface / adapter、read store repository、数据库 query / migration、API key / quota、workflow executor、confirmation、writeback 和 replay 仍独立排期。

Control Plane read store 的迁移架构固定为分层推进，而不是从 fake store 直接跳到数据库实现：先固定 repository contract preconditions，再固定 disabled database guard，随后定义 repository contract smoke、repository implementation readiness、store selection readiness、schema migration readiness、repository contract types、静态 contract smoke runner、repository interface readiness、adapter implementation readiness refresh、store selector enablement preconditions、schema migration implementation preconditions、repository adapter implementation plan、schema artifact manifest readiness 和 store selector smoke readiness。当前 Go 层已经有 `control_plane_read_repository_contract.go` 承载 `ReadRepositoryContext`、七条 route request / result type、failure code、projection / filter / sort 和 route type matrix；`control_plane_read_repository_contract_smoke_runner.go` 只消费这份 type matrix 与既有 smoke fixture，验证 route、failure mapping、no fake fallback 和 no side effects 的静态一致性。这个顺序要求未来 `ControlPlaneReadRepository`、adapter、store selector、schema artifact、migration runner 和数据库 role 必须先满足 tenant predicate、sanitized projection、schema artifact manifest gate、selector gate、failure mapping、no fake fallback 和 no side effects，再允许进入真实实现；当前仍没有 repository interface 文件、repository adapter、selector 文件、SQL、manifest、migration 文件、真实数据库、Radish OIDC、token validation 或 production API consumer。

`Workflow / Agent Runtime Function Surface v1` 已在 `apps/radishmind-web/` 落地 application detail、workflow definition detail、run detail、blocked action preview、confirmation placeholder、offline draft designer、offline validation inspector、execution plan preview、runtime readiness inspector、Workflow Surface Overview、context selection、Workflow Scenario Inspector、Workflow Review Workspace、User Workspace Home 和 Workflow Review Handoff。它们都是 fixture-derived、read-only、blocked-capability-first 的产品面：detail / draft / validation / plan / readiness 负责解释对象和准入证据，overview / scenario / review workspace / home / handoff 负责把当前 application、definition、run、draft、scenario、blocked capability 和 stop line 组织成可审查的用户工作区视图；这一层不新增 runtime API、builder mutation、draft / validation / execution plan / readiness / scenario / review / home / handoff persistence、executor、confirmation decision、writeback、replay、数据库、Radish OIDC 或 production API consumer。

2026-06-10 已在 `apps/radishmind-web/` 补齐普通离线 Model Gateway evidence 组织层和 Admin Operations Review / Readiness。Model Gateway Overview、Route Evidence、Usage/Audit Evidence 与 Evidence Review / Readiness 复用 read shell、API key、quota、run history、audit、provider runtime 和 gateway readiness 证据；Admin Operations Review / Readiness 复用 tenant overview、audit log、Model Gateway Evidence Review 和 Production Ops 静态证据。它们都只属于 read-only product surface，不新增 Go route、production gateway、secret resolver、API key lifecycle、quota enforcement、cost write、database、Radish auth、repository adapter、deployment preflight、executor、writeback 或 replay。

## 产品视角

### 1. `User Workspace`

- 面向用户创建和运行 AI 应用、Prompt 应用、Workflow、Agent / Copilot 应用、RAG / 知识问答应用。
- 消费模型网关、工作流 runtime、session metadata、运行记录、成本摘要和 API key 状态。
- 当前 `apps/radishmind-web/` 已实现 User Workspace 的 read-side 页面切片：applications、API keys、usage quota、workflow definitions 和 run history。它们只读、离线、无执行控件，不能替代完整 user workspace。
- 已有本地 console 不能替代用户 workspace。

### 2. `Admin Control Plane`

- 面向管理员管理租户、用户、角色、权限、provider/profile、模型路由、API key、quota、price、secret backend、审计和部署状态。
- 登录 / 授权、数据库、部署方式优先对齐 `Radish`；未来通过 OIDC 接入 `Radish` Auth，不在 RadishMind 内部自建身份真相源。
- Control Plane 默认使用 Go，可独立于 gateway 拆服务；当前不新增 `.NET` / ASP.NET Core 作为默认后端栈。
- 当前 `apps/radishmind-web/` 已实现只读 `admin-tenant-overview`、`admin-audit-log` 和普通离线 Admin Operations Review / Readiness。它们都不代表 production admin console。
- 当前本地 console 只是 ops surface，不等同于 production admin console。

### 3. `Model Gateway / API Distribution`

- 面向 API 调用者暴露 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models` 和后续 API key / quota / trace / cost 能力。
- 对内统一调度 provider registry、provider profile、selection policy、health signal、secret reference 和后续 retry/fallback execution。
- 当前已有 northbound bridge、provider/profile inventory、selection metadata、provider runtime gates，以及普通离线 Model Gateway Overview、Route Evidence、Usage/Audit Evidence 和 Evidence Review / Readiness；完整 API 分发、租户 quota、billing、load balancing 和 production secret backend 仍未实现。

### 4. `Workflow / Agent Runtime`

- 承载 Prompt、LLM、HTTP tool、RAG retrieval、condition、output 和后续受控 code / sandbox / agent loop。
- 每次运行必须保留 trace、输入输出摘要、错误分类、成本和风险边界。
- 当前已有 session/tooling metadata、blocked action shell，以及 `apps/radishmind-web/` 中的离线 application detail、definition detail、run detail、confirmation placeholder、draft designer、validation inspector、execution plan preview、runtime readiness inspector、surface overview、context selection、scenario inspector 和 review workspace；`workflow-definition-run-record-boundary` 已把 workflow definition、run record、node execution、tool audit、result materialization、confirmation decision、状态流转和审计证据固定为治理边界；真实 workflow builder、executor、durable store、confirmation flow、writeback 和 replay 仍未实现。

## 平台视角

### 1. `Runtime Service / Model Gateway`

- 负责启动、配置、provider/profile 选择、route 识别、gateway 封装、协议兼容和部署边界。
- 当前实现核心在 `scripts/run-copilot-inference.py`、`services/gateway/copilot_gateway.py`、`scripts/run-platform-bridge.py` 与 `services/platform/`。
- 当前 southbound 已开始由统一 `provider registry` 收口：现有 `mock`、`openai-compatible`、`HuggingFace`、`Ollama` 主入口与 `openai-compatible chat`、`gemini-native`、`anthropic-messages` 分流都归到同一条 provider truth；`local_transformers` 目前主要停留在 candidate/runtime 评测链路。
- 当前 northbound 对外形态已经开始由 `Go` 承载最小正式 `HTTP` 服务壳；`Python` 继续保留 CLI runtime 和 canonical gateway 语义，`Go` 只做协议兼容与进程调度，避免把平台服务层锁死在 `Python`。本地 console origin 的 CORS / preflight 只服务 `P3` 本地消费面，不代表 production 鉴权或公开部署策略。
- `UI` 层默认 `React + Vite + TypeScript`，通过北向协议消费平台能力，不直接承载模型实现逻辑。
- 未来 `Admin Control Plane` 的身份、权限和数据库应优先对齐 `Radish`；RadishMind 作为 OIDC client 消费 `Radish` 登录态，不在当前 gateway 内提前实现完整账号系统，也不为了对齐 Radish 复制其后端语言栈。
- 当前 `P3 Local Product Shell / Ops Surface` 已在平台服务层暴露只读 `/v1/platform/overview` 与 `/v1/platform/local-smoke`，并用 TypeScript overview / local-smoke consumer contract、consumer smoke、console shell check、console behavior gate、console visual smoke record、dev entry check、console production packaging boundary gate、P3 checklist 与 `apps/radishmind-console/` 本地 console 壳固定 service status、model inventory、Provider/Profile Details、session/tooling surface、stop-line view model、Stop-line Details、Dev Diagnostics、`Local Readiness` 面板、refresh 状态、overview / local-smoke failure surface、连接失败诊断、production packaging 停止线和 P3 hardening 缺口。该本地只读产品壳已达到 `local usable / read-only close`。
- 当前 `Production Ops Hardening v1` 已把部署边界拆成独立层：`host_dev` 仍使用宿主机 wrapper，`docker_local` 使用 `deploy/docker-compose.local.yaml` 本地 build 与 mock provider，`docker_test` / `docker_prod` 共用 `deploy/docker-compose.yaml` 并通过镜像 track / tag、provider profile、secret 来源和外部反代配置区分环境。该层已有 Dockerfile、compose、`.env.example`、镜像命名治理、静态 compose 展开、container smoke runbook 和运行记录模板，但仍不声明 production ready。

### 2. `Conversation & Session`

- 负责 `conversation_id`、会话历史、恢复、压缩和审计边界。
- 当前已有首版 session record、history policy、state policy、recovery record、recovery checkpoint record/manifest/read result、northbound session metadata、metadata-only route smoke、session metadata route 和 overview 消费面；confirmation flow design、independent audit records design、result materialization policy design、executor boundary design、storage backend design、negative regression governance suite、route negative coverage matrix、short close readiness delta、readiness consistency rollup、executor/storage/confirmation enablement plan、stop-line manifest 和 foundation status summary 只代表 close candidate / governance-only，仍没有 durable session/checkpoint/audit/result store、长期记忆、真实 checkpoint storage backend、materialized result reader 或跨轮恢复执行器。

### 3. `Tooling Framework`

- 负责检索、附件解析、项目语义转换、本地候选生成、response builder 和工具策略。
- 当前已有首版 tool contract、registry、policy/audit record、session binding、metadata-only result cache、tool metadata route、blocked action route、checkpoint read `tool_audit_summary`、overview 消费面、promotion gate、负向消费 summary、route smoke coverage summary、readiness summary、implementation preconditions、negative regression skeleton、governance-only negative regression suite、deny-by-default implementation gates、negative coverage rollup、route smoke readiness rollup、readiness drift 检查和五类设计门禁；仍没有真实工具执行器、materialized tool result cache、durable tool store、durable audit/result store 或上层确认流接线。

### 4. `Model Runtime`

- 负责模型推理、候选文本生成、结构化约束、guided/runtime 协同和图片理解。
- `RadishMind-Core` 属于这一层，但不是整个平台本身。

### 5. `Evaluation & Governance`

- 负责 schema、smoke、offline eval、candidate record、review、promotion gate 与仓库级检查。
- 这一层保证平台不是“能跑一次”，而是“能长期复跑并解释为什么通过或不通过”。当前 P2 设计级门禁由 fixture + check + task card 共同固定，但它们仍属于治理边界，不是实现完成声明。

## 请求视角

当前还有一层必须继续补齐、但已经开始正式落地的协议翻译边界：

- 北向：`/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models`、`/v1/platform/overview`、`/v1/platform/local-smoke`、session/tooling metadata shell、Control Plane Read-Side read-only route 等兼容接口和产品面如何映射到 canonical request 或只读 discovery view
- 南向：`RadishMind-Core`、`HuggingFace`、`Ollama`、OpenAI-compatible、Gemini、Anthropic 等 provider 如何被统一调度

平台内部真相源仍应保持 `CopilotRequest / CopilotResponse / CopilotGatewayEnvelope`，兼容接口只做翻译层，不另起第二套真相源。

当前单次请求的主流程仍保持六层，部署边界作为包裹平台进程的运行层存在：

1. Client Adapters & Context Packers
2. Copilot Gateway / Task Router
3. Retrieval & Tool Layer
4. Model Runtime Layer
5. Rule Validation & Response Builder
6. Data / Evaluation / Training Pipeline

这个拆分用于隔离项目语义、工具编排、模型推理、安全校验和评测闭环，让 `RadishFlow`、`Radish` 与后续 `RadishCatalyst` 能通过统一协议接入，而不是各自私接模型。

部署层不改变 canonical request / response 流程。Docker compose 只负责启动 platform 和 console 容器，不成为 secret backend、process supervisor、confirmation flow、tool executor 或业务写回层。

## 请求流程

```text
外部客户端协议 / 上层项目状态 / 文档 / 附件 / 图像
        ↓
Protocol Compatibility Layer 归一到 canonical request
        ↓
Adapter 打包为 CopilotRequest
        ↓
Gateway 识别 project / task / schema_version
        ↓
Retrieval / Tools 获取证据并压缩上下文
        ↓
Model Runtime 生成解释、问题和候选意图
        ↓
Rule Validation / Response Builder 收口结构和风险
        ↓
CopilotResponse / GatewayEnvelope
        ↓
Protocol Compatibility Layer 翻译回 northbound response
```

## 分层职责

### 1. Client Adapters & Context Packers

- 将上层状态转换为统一 `CopilotRequest`。
- 裁剪、脱敏或摘要敏感字段。
- 将 `CopilotResponse` 映射回 UI、日志或候选提案。
- 前端 UI 默认使用 `React + Vite + TypeScript`，通过协议对接后端，不直接依赖模型实现语言。
- 当前 `RadishFlow` 优先维护 `export -> adapter -> request` 链路；`Radish` 优先维护文档和内容上下文；`RadishCatalyst` 暂不落真实 adapter。

### 2. Copilot Gateway / Task Router

- 统一校验请求、识别任务、选择 provider/profile，并返回 `CopilotGatewayEnvelope`。
- 当前 `SUPPORTED_ROUTES` 仍然有限，说明平台还在先做骨架而不是全量铺开任务面。
- 当前 `Go` 平台服务层已经通过 Python bridge 接到 `/v1/chat/completions`、`/v1/responses`、`/v1/messages` 与 `/v1/models` 的第一版兼容层，并额外暴露只读 `/v1/platform/overview`、`/v1/platform/local-smoke`、`/v1/session/metadata`、`/v1/tools/metadata`、blocked `/v1/tools/actions` 产品壳和七条 Control Plane Read-Side read-only route。bridge 目前仍是窄切片，先把非流式文本消息固定映射到 `radish/answer_docs_question`，并通过 SSE 做出第一版流式兼容骨架、把 `/v1/models` 从 provider 目录推进到 bridge-backed provider/profile inventory，再补上 `GET /v1/models/{id}` 的精确 lookup；当前已经把 request-side provider/profile 选择显式化，并把流式路径推进到 bridge 增量转发，同时补了 `HuggingFace` / `Ollama` 的第一版 southbound provider coverage。`/v1/platform/overview` 和 `/v1/platform/local-smoke` 只聚合服务状态、本地 readiness、model inventory、session/tooling route 和停止线；Control Plane Read-Side 当前默认通过离线 fixture / TypeScript view model 展示，dev-only live read path 也只通过 test-only fake auth context 和 in-memory fake store 验证 HTTP consumer shape。repository/read store 的 Go contract type、静态 runner、adapter plan、schema artifact manifest readiness 和 selector smoke readiness 只服务迁移前契约校验，没有真实 auth middleware、数据库 read store、repository adapter、store selector 或生产 UI。
- 服务/API smoke 当前锁定 advisory-only、schema validation、route metadata、error envelope 和 handoff 不执行这些不变量。

### 3. Retrieval & Tool Layer

- 承载文档检索、附件/Markdown/JSON 解析、项目语义转换和本地合法候选生成。
- 能规则化的逻辑优先留在工具层，例如 ghost completion 的合法候选空间和 recent-actions suppress 信号。
- 工具层只生成证据或候选动作，不直接写业务真相源。

### 4. Model Runtime Layer

- Teacher models 用于强基线、标注参考、蒸馏和复杂任务对照。
- Student models 用于本地化、小成本部署和回归实验。
- `RadishMind-Core` 负责理解、推理、结构化建议、候选排序、风险标记和可选图片输入理解。
- Image Generation Runtime 独立负责图片像素生成；主模型只输出结构化 image intent 和约束。

### 5. Rule Validation & Response Builder

- 校验响应结构、目标对象、风险等级、确认要求、citation 和 action shape。
- 对可规则化字段保持确定性，例如 `status / risk_level / requires_confirmation / proposed_actions / patch / issue code`。
- 当前 `task-scoped response builder` 仍是 `M4` 决策实验和 tooling 分工证据，不是 raw 模型晋级或 production contract。

### 6. Data / Evaluation / Training Pipeline

- 管理 eval sample、candidate record、audit、replay、offline eval、training sample manifest 和 review records。
- 真实模型输出、生成 JSONL 和大体积实验产物默认留在 `tmp/`。
- 训练、蒸馏和模型晋级必须同时看 raw 输出、后处理轨、离线评测、自然语言 audit、人工 review 和 holdout 泄漏边界。

### 7. Deployment Boundary Layer

- `host_dev` 是默认开发形态，使用宿主机 wrapper 启动 platform / console，不走 Compose。
- `docker_local` 是本地容器 smoke 形态，允许本机 build，默认 `mock` provider，只能证明本地容器编排和 health probe 可运行。
- `docker_test` 与 `docker_prod` 共用部署态 compose，只引用预构建镜像，通过 `RADISHMIND_IMAGE_TRACK` 或 `RADISHMIND_IMAGE_TAG` 区分 test / release 轨道。
- 部署态 compose 只定义 platform / console 容器、端口、healthcheck 和必要环境变量，不定义 Compose secrets，不执行本地 build，不用 `restart` 策略伪装 process supervisor。
- 真实运行证据应写到 `tmp/production-ops/container-smoke/`，按 record template 记录命令、probe、container、cleanup、result 和运行后 blocked condition；committed fixture 只描述模板，不保存实际运行结果。

## 当前架构映射

- `Frontend UI`：`React + Vite + TypeScript`
- `Runtime Service`：`scripts/run-copilot-inference.py`、`services/gateway/copilot_gateway.py`、`scripts/run-platform-bridge.py`
- `Platform Service Layer`：`services/platform/`，使用 `Go` 承载 `HTTP API`、`gateway`、鉴权落点、流式转发、长驻进程、观测和部署壳；当前已落第一版 bridge-backed northbound、session/tooling metadata shell、blocked action shell、只读 platform overview、local smoke readiness route、七条 fake-store-backed control plane read route 和显式 dev-only fake auth header gate
- `Control Plane`：长期默认使用 `Go`，可按职责拆出独立服务，覆盖 tenant、API key、quota、provider profile、OIDC client、audit 和 run records；当前 read-side 契约、Go fake-store-backed handler、TypeScript consumer contract、auth/db preconditions、formal UI boundary/readiness、dev-only live consumer、auth/store transition preconditions、repository contract smoke、repository implementation readiness、store selection readiness、schema migration readiness、repository contract types、静态 contract smoke runner、repository interface readiness、adapter implementation readiness refresh、selector enablement preconditions、schema migration implementation preconditions、repository adapter implementation plan、schema artifact manifest readiness 和 store selector smoke readiness 说明见 `docs/contracts/control-plane-read-side.md`，不默认引入 `.NET`
- `Product Surfaces`：目标形态包括 `User Workspace`、`Admin Control Plane`、`Model Gateway / API Distribution` 和 `Workflow / Agent Runtime`；当前已落地本地 ops console 壳、只读平台发现面、fake-store-backed read-side HTTP surface，以及 `apps/radishmind-web/` 的 read-only product UI shell。`apps/radishmind-web/` 默认离线，显式 opt-in 时可通过 dev-only HTTP consumer 消费 fake-store-backed handler；已覆盖 route catalog、共享状态、forbidden output guard、两个 admin 页面、五个 user workspace 只读页面、User Workspace Home、Workflow Review Workspace、Workflow Review Handoff、workflow function surface 离线产品面、Model Gateway 四个 evidence 面板和 Admin Operations Review / Readiness，但还不是完整 user workspace、production admin console、真实 workflow builder、production gateway 或真实 API consumer
- `P3 Local Product Shell / Ops Surface`：`GET /v1/platform/overview`、`GET /v1/platform/local-smoke`、`contracts/typescript/platform-overview-api.ts`、`contracts/typescript/platform-local-smoke-api.ts`、`scripts/run-platform-overview-consumer-smoke.py`、`scripts/run-platform-local-smoke.py`、`scripts/check-radishmind-console-behavior.py`、`scripts/check-radishmind-console-visual-smoke-record.py`、`scripts/check-radishmind-console-dev-entry.py`、`scripts/check-radishmind-console-production-boundary.py`、`scripts/check-p3-local-product-shell-short-close-checklist.py`、`apps/radishmind-console/`、`docs/contracts/platform-overview-ui-view.md`；当前本地只读壳已达到 `local usable / read-only close`
- `Deployment Boundary Layer`：`deploy/README.md`、`deploy/docker-compose.local.yaml`、`deploy/docker-compose.yaml`、`deploy/.env.example`、`services/platform/Dockerfile`、`apps/radishmind-console/Dockerfile`、`apps/radishmind-console/nginx.local.conf`、`scripts/check-production-ops-docker-*.py`、`scripts/check-production-ops-deployment-readiness-smoke.py`、`scripts/check-production-ops-container-smoke-*.py`；当前只固定 docker local/test/prod 边界、镜像命名、静态展开、runbook 和记录模板
- `Southbound Provider Layer`：`services/runtime/provider_registry.py`、`services/runtime/inference_provider.py`、`services/platform/internal/httpapi/northbound.go`、`scripts/checks/fixtures/provider-capability-matrix-v1.json`、`scripts/check-provider-capability-matrix.py`、`scripts/checks/fixtures/provider-health-smoke-v1.json`、`scripts/check-provider-health-smoke.py`、`scripts/checks/fixtures/provider-selection-policy-v1.json`、`scripts/check-provider-selection-policy.py`、`scripts/checks/fixtures/provider-retry-fallback-policy-v1.json`、`scripts/check-provider-retry-fallback-policy.py`、`scripts/checks/fixtures/provider-runtime-docs-refresh.json`、`scripts/check-provider-runtime-docs-refresh.py`
- `Conversation & Session`：`contracts/session-record.schema.json`、`contracts/session-recovery-checkpoint*.schema.json`、northbound session metadata、平台 checkpoint metadata-only route smoke、readiness summary、implementation preconditions、route smoke readiness rollup、short close readiness delta、stop-line manifest 和 storage / audit / result 边界 fixture
- `Tooling Framework`：`contracts/tool*.schema.json`、tool registry / audit fixture、`scripts/check-tooling-framework-contract.py`、`scripts/check-session-recovery-checkpoint-contract.py`、confirmation flow design、executor boundary design、result materialization policy design、negative regression suite、deny-by-default gates、enablement plan 和各类 deterministic builder/check
- `Evaluation & Governance`：`scripts/check-repo.py`、`scripts/check-radishflow-service-smoke-matrix.py`、offline eval、review records、promotion gates、negative consumption summary、negative coverage rollup、route negative coverage matrix、readiness consistency rollup、foundation status summary 和 P2 design gate checks
- `Model Runtime`：`services/runtime/`、`scripts/run-radishmind-core-candidate.py`

## 当前缺口

- 当前只有 first-pass `Go` platform service、bridge-backed `HTTP API` 和可检查 Docker 部署边界，还不是 production deployment
- northbound `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models`、`/v1/platform/overview`、`/v1/platform/local-smoke` 和 session/tooling metadata shell 已具备第一版兼容 / discovery 接口；当前已补第一版 SSE 流式兼容骨架、bridge-backed provider/profile inventory、request-side provider/profile selection、流式增量转发、`/v1/models` 列表 + 精确 lookup、结构化 diagnostics、discoverability 对齐、请求级观测、错误分类、overview consumer smoke、local-smoke readiness smoke、Docker 静态部署边界和一次 `docker_local` container smoke 运行记录，但测试环境 smoke、镜像发布和生产前复核仍未完成
- `HuggingFace` 与 `Ollama` 已进入 provider/profile inventory、diagnostics、provider capability matrix、provider health smoke、provider selection policy、provider retry/fallback policy 和 `provider-runtime-docs-refresh` 门禁；默认 health smoke 只覆盖 mock runtime 与 config-level inventory，selection policy 固定 no implicit fallback 与负向分类，retry/fallback policy 固定 `retry_policy=caller-managed` 与 `fallback_policy=disabled` 审计 metadata，docs refresh 固定入口文档口径，正式 secret backend、环境隔离、外部 provider live health 和 retry/fallback execution 仍未补齐
- 已有 session/tooling 首版契约、metadata-only 门禁、close-candidate status summary、negative regression governance suite、route/gate coverage rollup、readiness consistency rollup、short close delta、enablement plan、stop-line manifest、五类设计级边界门禁和只读本地 console 消费壳，但没有 durable session/checkpoint/audit/result store、长期记忆、真实 checkpoint storage backend、materialized result reader 或跨轮恢复执行器
- 已有 tool registry、tool audit、metadata-only result cache、result materialization policy design、executor boundary design 和 deny-by-default gate contract，但没有真实工具执行器、materialized result reader、durable tool store、durable result store 或上层确认流接线
- 已有 Control Plane Read-Side 七条 fake-store-backed Go route、统一 response envelope、负向 route smoke、TypeScript consumer contract、auth/db preconditions、formal UI boundary/readiness、七个只读页面、read-side UI 聚合收口、dev-only live read consumer、auth/store transition preconditions、repository contract type、静态 contract smoke runner、repository interface readiness、adapter readiness refresh、selector enablement preconditions、schema migration implementation preconditions、repository adapter implementation plan、schema artifact manifest readiness、store selector smoke readiness 和 `apps/radishmind-web/` read-only React shell，但没有真实 Radish OIDC / auth middleware、token validation、repository interface / adapter、store selector、schema artifact、read store repository implementation、数据库 query / migration、完整 user workspace 或 production admin console
- 尚未具备 production secret backend、process supervisor、正式部署环境隔离、真实镜像发布 workflow、测试环境 smoke、生产前复核记录、console runtime config 和可发布部署包；这些仍是 Production Ops 后置缺口，但没有明确测试或生产前复核窗口时不作为默认开发推进方向

这些缺口说明：`P1 Runtime Foundation` 已达到 short close，`P2 Session & Tooling Foundation` 当前是 close candidate / governance-only，`P3 Local Product Shell / Ops Surface` 的本地只读壳已达到 `local usable / read-only close`，`UI Design Topic / React 第二批` 和 P4 前置证据已进入 close / 后置专题状态，`Provider Runtime & Health v1` 也已进入 close candidate。下一步只有在明确运行窗口后才执行容器 smoke 或测试环境 smoke，并按运行记录模板沉淀证据；没有运行窗口时，默认沿 `apps/radishmind-web/` 继续推进可复用的离线产品证据组织，优先补 Admin provider/profile 与 deployment status 的只读审查路径，不继续补 P3 console 同类小展示项、provider 同层小切片、回头扩 P2 readiness、真实 executor、durable store、confirmation 接线或真实模型长跑。

## 当前进度

- `contracts/` 已具备 Copilot request / response / gateway envelope / training sample / image generation intent / backend request / artifact schema。
- `RadishFlow` 的 gateway demo、service smoke matrix、UI consumption 和 candidate edit handoff 已作为未来接入门禁保留；在上层项目尚未具备真实接入能力前，当前只收口前置条件与阻塞项，不继续细化新的接线设计或模拟接入 summary。
- `suggest_flowsheet_edits` 与 `suggest_ghost_completion` 的真实 candidate record、audit、replay 和治理链已阶段性收口；新增真实 capture 需要先说明非重复 drift 假设。
- `RadishMind-Core` 本地小模型观测显示 raw 仍 blocked；broader review 的 15/15 `reviewed_pass`、`3B/4B` guided capacity review、1.5B full-holdout-9 raw / repaired comparison 和 3B CPU 单样本 timeout 当前只保留为路线证据，在没有 GPU / 明确实验窗口 / 新能力假设前不再默认继续真实模型产出专题。
- `RadishMind-Image Adapter` 已具备 intent、backend request、artifact metadata、最小评测 manifest、`image_adapter_handshake_safety_gate_defined`、`image_artifact_return_runbook_evidence_defined`、`image_safety_runbook_evidence_defined` 与 `image_backend_adapter_readiness_defined` 证据；当前只固定 metadata reference / safety gate / 安全审查 runbook / backend adapter readiness，不调用真实生图 backend。

## 工程约束

- 层之间通过 schema、明确数据类型和稳定函数边界连接，避免隐式全局状态或字符串拼装协议。
- 代码优先使用对应语言的标准库和惯用结构；本仓库 Python 代码应保持直接、可测试、易读。
- 方法名和模块名必须说明真实职责；不要用空泛 helper、manager、processor 掩盖边界不清。
- 不为简单调用链增加多层抽象；只有当职责稳定、复用真实存在或复杂度明显下降时才抽 helper、builder 或 adapter。
- 修复结构漂移时优先修正 schema、builder 或任务边界，不用无限 fallback 包裹模型输出。
