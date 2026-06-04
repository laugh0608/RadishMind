# RadishMind 阶段路线图

更新时间：2026-06-04

## 路线图原则

路线图只记录阶段目标、当前进度、下一步和停止线。批次细节、历史失败、完整实验输出和长命令不放在本入口文档中，应进入周志、实验 manifest、run record 或任务卡。

当前长期目标更新为：`RadishMind` 是 `Radish` 体系下的 AI 工具、工作流、模型网关和 Copilot 集成平台，不是单一万能模型，也不是只服务本地 demo 的 runtime 壳。

若要理解“为什么路线这样排”，先读 [战略定义](radishmind-strategy.md)。

长期产品机会单独维护在 [产品机会池](radishmind-product-ideas.md)。该文件只记录候选方向，不代表路线图承诺；进入实施前必须再拆任务卡。

## 当前路线切换

从 2026-05-10 开始，仓库主线正式从“围绕 `M3/M4` 收口继续做局部维护”切换为“基于已收口证据继续建设平台本体”。

当前已经冻结的历史证据：

- `M3`：gateway、service smoke、UI consumption 与 candidate handoff 已收口为服务/API 门禁。
- `M4`：broader 15 样本人工复核为 15/15 `reviewed_pass`，`3B/4B` guided capacity review 已收口为正式审计记录。

这些资产继续保留，但不再等同于当前唯一主线。

## 四个产品面

长期产品路线按四个产品面展开：

1. `User Workspace`：AI 应用、Prompt 应用、Workflow、Agent / Copilot 应用、RAG / 知识问答应用、API key、调用量、运行记录和成本摘要。
2. `Admin Control Plane`：租户、用户、角色、权限、provider/profile、模型路由、API key、quota、price、secret backend、审计和部署状态；未来作为 OIDC client 接入 `Radish`。Control Plane 默认使用 Go，可独立拆服务，不因参考 Radish 而默认引入 `.NET` / ASP.NET Core。
3. `Model Gateway / API Distribution`：OpenAI-compatible / Responses / Messages / Models 等 API 分发，多 provider / profile / model 路由，后续补 quota、限流、成本、trace、fallback、load balancing 和 health。
4. `Workflow / Agent Runtime`：Prompt、LLM、HTTP tool、RAG retrieval、condition、output 和后续受控 agent loop；高风险动作默认要求确认。

当前实现仍处在平台基础能力阶段。已有本地 console 只覆盖 ops surface 和只读发现面，不代表正式用户端、生产管理端或完整 workflow builder 已完成。

2026-05-27 已新增 [Control Plane / User Workspace / Workflow v1 计划](task-cards/control-plane-user-workspace-workflow-v1-plan.md)，先固定四个产品面的服务边界、数据边界、建议切片和停止线；`product-surface-v1-boundary` 已落地为 `product-surface-v1-boundary.json` 与 `check-product-surface-v1-boundary.py`，`control-plane-data-boundary` 已落地为 `control-plane-data-boundary.json` 与 `check-control-plane-data-boundary.py`，`radish-oidc-client-preconditions` 已落地为 `radish-oidc-client-preconditions.json` 与 `check-radish-oidc-client-preconditions.py`，`gateway-api-key-quota-readiness` 已落地为 `gateway-api-key-quota-readiness.json` 与 `check-gateway-api-key-quota-readiness.py`，`workflow-definition-run-record-boundary` 已落地为 `workflow-definition-run-record-boundary.json` 与 `check-workflow-definition-run-record-boundary.py`，`control-plane-read-model-v1` 已落地为 `control-plane-read-model-v1.json` 与 `check-control-plane-read-model-v1.py`，`control-plane-read-route-contract-v1` 已落地为 `control-plane-read-route-contract-v1.json` 与 `check-control-plane-read-route-contract-v1.py`，`control-plane-read-response-fixtures-v1` 已落地为 `control-plane-read-response-fixtures-v1.json` 与 `check-control-plane-read-response-fixtures-v1.py`，`control-plane-read-negative-contract-v1` 已落地为 `control-plane-read-negative-contract-v1.json` 与 `check-control-plane-read-negative-contract-v1.py`，`control-plane-read-implementation-preconditions-v1` 已落地为 `control-plane-read-implementation-preconditions-v1.json` 与 `check-control-plane-read-implementation-preconditions-v1.py`，`control-plane-read-fake-store-handler-plan-v1` 已落地为 `control-plane-read-fake-store-handler-plan-v1.json` 与 `check-control-plane-read-fake-store-handler-plan-v1.py`，`control-plane-read-fake-store-handler-implementation-v1` 已落地为七条 fake-store-backed read handler implementation，`control-plane-read-auth-db-preconditions-v1` 已落地为 `control-plane-read-auth-db-preconditions-v1.json` 与 `check-control-plane-read-auth-db-preconditions-v1.py`，`control-plane-read-consumer-contract-v1` 已落地为 `control-plane-read-consumer-contract-v1.json` 与 `check-control-plane-read-consumer-contract-v1.py`，固定上层消费契约；`control-plane-read-formal-ui-boundary-v1` 已落地为 `control-plane-read-formal-ui-boundary-v1.json` 与 `check-control-plane-read-formal-ui-boundary-v1.py`，固定正式 UI 边界；这些切片不实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay。read-side 已完成当前页面集合、`control-plane-read-formal-ui-readiness-close-v1`、`control-plane-read-dev-live-consumer-v1` 与 `control-plane-read-auth-store-transition-preconditions-v1`。

2026-05-28 已补 `control-plane-read-formal-ui-implementation-readiness-v1`，落地为 `control-plane-read-formal-ui-implementation-readiness-v1.json` 与 `check-control-plane-read-formal-ui-implementation-readiness-v1.py`，固定正式 UI 实现前的 `apps/radishmind-web/` 预留落点、`apps/radishmind-console/` 边界、页面顺序、contract 复用和测试策略；它不创建 React 页面、不创建 product UI app、不把本地 ops console 升级为 production admin console。

2026-05-31 已完成 `control-plane-read-shared-shell-v1`，创建 `apps/radishmind-web/` 的首个 read-only shared shell，复用 `contracts/typescript/control-plane-read-api.ts` 渲染 route catalog、共享状态组件和 forbidden output guard。该切片不请求真实后端、不接数据库 / OIDC、不实现 API key / quota、workflow executor、confirmation、writeback 或 replay，也不改变 `apps/radishmind-console/` 的本地 ops surface 定位。

2026-05-31 已完成 `control-plane-read-admin-tenant-overview-v1`，在 `apps/radishmind-web/` 的 shared shell 内新增只读 `admin-tenant-overview` 页面切片，消费 `tenant-summary-route` 渲染租户摘要、route metadata、request / audit ref、页面状态和 forbidden output guard。该切片不请求真实后端、不接数据库 / OIDC、不实现 API key / quota、workflow executor、confirmation、writeback 或 replay，也不声明 production admin console ready。

2026-05-31 已完成 `control-plane-read-workspace-applications-v1`，在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-applications` 页面切片，消费 `application-summary-list-route` 渲染应用摘要列表、cursor、route metadata、request / audit ref、页面状态和 forbidden output guard。该切片不请求真实后端、不接数据库 / OIDC、不实现 API key / quota、workflow executor、confirmation、writeback 或 replay，也不声明 formal user workspace complete。

2026-05-31 已完成 `control-plane-read-workspace-api-keys-v1`，在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-api-keys` 页面切片，消费 `api-key-summary-list-route` 渲染 API key id、owner、scope、state、时间字段、route metadata、request / audit ref、页面状态和 forbidden output guard。该切片不请求真实后端、不接数据库 / OIDC、不实现 API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay，不展示 key value 或 hash，也不声明 formal user workspace complete。

2026-05-31 已完成 `control-plane-read-workspace-usage-quota-v1`，在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-usage-quota` 页面切片，消费 `quota-summary-route` 渲染 quota id、period、request / token / cost limit、usage snapshot、over quota failure code、route metadata、request / audit ref、页面状态和 forbidden output guard。该切片不请求真实后端、不接数据库 / OIDC、不实现 quota enforcement、rate limit、billing、cost ledger、workflow executor、confirmation、writeback 或 replay，也不声明 formal user workspace complete。

2026-05-31 已完成 `control-plane-read-workspace-workflow-definitions-v1`，在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-workflow-definitions` 页面切片，消费 `workflow-definition-summary-list-route` 渲染 workflow definition id、application ref、version、definition status、node count、risk level、requires confirmation capable、updated at、route metadata、request / audit ref、页面状态和 forbidden output guard。该切片不请求真实后端、不接数据库 / OIDC、不实现 workflow builder、workflow definition lifecycle mutation、workflow executor、tool executor、confirmation、writeback 或 replay，也不声明 formal user workspace complete。

2026-05-31 已完成 `control-plane-read-workspace-run-history-v1`，在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-run-history` 页面切片，消费 `run-record-summary-list-route` 渲染 run id、workflow definition ref、application ref、status、failure code、cost summary、trace id、started / completed timestamp、route metadata、request / audit ref、cursor、页面状态和 forbidden output guard。该切片不请求真实后端、不接数据库 / OIDC、不实现 workflow executor、tool executor、run replay、run resume、materialized result reader、confirmation、writeback 或 replay，也不声明 formal user workspace complete。

2026-06-01 已完成 `control-plane-read-admin-audit-log-v1`，在 `apps/radishmind-web/` 的 shared shell 内新增只读 `admin-audit-log` 页面切片，消费 `audit-summary-list-route` 渲染 audit ref、actor、event kind、resource、decision、failure code、trace id、recorded timestamp、route metadata、request / audit ref、cursor、页面状态和 forbidden output guard。该切片不请求真实后端、不接数据库 / OIDC、不实现 durable audit store、raw payload export、audit record mutation、workflow executor、confirmation、writeback 或 replay，也不声明 production admin console ready。

2026-06-01 已完成 `control-plane-read-formal-ui-readiness-close-v1` formal UI readiness close，以 `formal_ui_readiness_closed` 状态和 surface matrix 聚合校验七个只读页面的 route binding、状态预览、request / audit ref、forbidden output guard 和停止线。后续不再默认为普通只读展示页逐项新增 task card、fixture 和 checker；如果放宽到 dev-only live read path，只能消费 fake-store-backed handler 与测试身份上下文，真实数据库、Radish OIDC、API key lifecycle、quota enforcement、billing、workflow executor、confirmation、writeback 和 replay 仍保持停止线。

2026-06-01 已完成 `control-plane-read-dev-live-consumer-v1`，在 `apps/radishmind-web/` 内新增 dev-only live read consumer 路径。默认仍使用离线 fixture/view model；只有显式设置 `VITE_RADISHMIND_READ_SOURCE=dev-live-http`，且平台服务设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 时，页面才会通过 HTTP 消费现有 fake-store-backed read handlers 和测试身份上下文。该切片不接真实数据库、不接 Radish OIDC、不实现 repository migration、API key lifecycle、quota enforcement、billing、cost ledger、workflow executor、confirmation、writeback 或 replay，也不声明 production API consumer、production admin console 或完整 formal user workspace ready。

2026-06-01 已完成 `control-plane-read-auth-store-transition-preconditions-v1`，把 dev fake auth / fixture-backed fake store 迁移到未来 `future Radish OIDC / auth middleware` 与 `future control plane read store repository` 前的 auth/store transition preconditions 固定为可检查证据。该切片只定义 auth middleware gates、read store gates、route transition matrix、dual smoke plan、failure code 和禁止项，不接真实数据库、不接 Radish OIDC、不实现 token validation、repository migration、repository implementation、API key lifecycle、quota enforcement、billing、cost ledger、workflow executor、confirmation、writeback 或 replay，也不声明 production API consumer ready。
2026-06-02 已完成 `control-plane-read-repository-contract-preconditions-v1`，把未来 read store repository contract 固定为可检查证据：`ControlPlaneReadRepository` interface、`ReadRepositoryContext`、七条 read route 到 repository operation 的映射、tenant predicate、sanitized projection、cursor/filter/sort allowlist、contract smoke 和 failure mapping。该切片只推进 read store repository contract，不写 SQL、不建 migration、不实现 repository、不接真实数据库、不接 Radish OIDC、不实现 token validation、API key lifecycle、quota enforcement、billing、cost ledger、workflow executor、confirmation、writeback 或 replay，也不声明 production API consumer ready。

2026-06-02 已完成 `control-plane-read-disabled-database-guard-v1`，把 disabled database read guard 固定为可检查证据：database / postgres / repository read mode 当前仍是 reserved disabled，七条 route 误入 database mode 时必须 fail-closed 为 `database_read_disabled`，不得静默回退到 fake store，也不得产生写入或执行副作用。该切片不新增正式配置入口、不写 SQL、不建 migration、不实现 repository adapter、不接真实数据库、不接 Radish OIDC、不实现 token validation、API key lifecycle、quota enforcement、billing、cost ledger、workflow executor、confirmation、writeback 或 replay，也不声明 production API consumer ready。

2026-06-02 已完成 `control-plane-read-repository-contract-smoke-v1`，把未来 repository contract smoke 固定为可检查证据：`ControlPlaneReadRepositoryContractSmoke` 输入输出、repository context、七条 read route smoke matrix、failure mapping、no fake fallback、no side effects 和文档停止线。该切片不实现 smoke runner、不写 SQL、不建 migration、不实现 repository adapter、不接真实数据库、不接 Radish OIDC、不实现 token validation、production API consumer、API key lifecycle、quota enforcement、billing、cost ledger、workflow executor、confirmation、writeback 或 replay，也不声明 repository implementation ready。
2026-06-02 已完成 `control-plane-read-repository-implementation-readiness-v1`、`control-plane-read-store-selection-readiness-v1` 和 `control-plane-read-schema-migration-readiness-v1`：分别固定 repository implementation readiness、store selection readiness 与 schema migration readiness。三者只定义文件落点 / 选择策略 / schema ownership / migration layout / rollback / tenant index / read-only role / failure mapping / no fake fallback / no side effects，不创建正式配置入口、migration 目录或 SQL，不实现 store selector、repository interface / adapter、migration runner、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、billing、workflow executor、confirmation、writeback 或 replay。
2026-06-03 已完成 `control-plane-read-repository-contract-types-readiness-v1`，固定 repository contract types readiness：未来 `ReadRepositoryContext`、七条 read route request / result type、failure code type、projection / filter / sort type 和 contract smoke type 输入已进入可检查证据。该切片不创建 Go contract type 文件，不实现 repository interface / adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、billing、workflow executor、confirmation、writeback 或 replay。
2026-06-03 已完成 `control-plane-read-repository-contract-types-implementation-v1`、`control-plane-read-repository-contract-smoke-runner-readiness-v1` 与 `control-plane-read-repository-contract-smoke-runner-implementation-v1`：完成 repository contract types implementation 与 repository contract smoke runner implementation，创建 Go repository contract type 文件和测试，并实现静态 repository contract smoke runner；runner 消费 `controlPlaneReadRepositoryRouteTypeContracts()`，并与既有 smoke fixture 对齐七条 read route、failure mapping、no fake fallback 和 no side effects。该阶段仍不声明 repository interface，不实现 adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。
2026-06-03 已完成 `control-plane-read-repository-interface-readiness-v1`：固定 repository interface readiness，未来 `ControlPlaneReadRepository` method matrix 必须消费已落地 Go contract type 与静态 runner 证据，并继续保留 adapter implementation gate、production auth gate、failure mapping 和 no side effects。该切片不创建 interface 文件，不声明 repository interface，不实现 adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。
2026-06-04 已完成 `control-plane-read-repository-adapter-implementation-readiness-refresh-v1`、`control-plane-read-store-selector-enablement-preconditions-v1`、`control-plane-read-schema-migration-implementation-preconditions-v1`、`control-plane-read-repository-adapter-implementation-plan-v1` 与 `control-plane-read-schema-artifact-manifest-readiness-v1`：分别固定 repository adapter implementation readiness refresh、store selector enablement preconditions、schema migration implementation preconditions、repository adapter implementation plan 与 schema artifact manifest readiness，最新状态为 `schema_artifact_manifest_readiness_defined`。五者只对齐 adapter gate、selector enablement gates、migration artifact manifest、DDL review、rollback fixture、schema version / tenant index / read-only role smoke、七条 route adapter / schema artifact matrix、failure mapping、no fake fallback 和 no side effects，不创建 interface / adapter / selector / migration runner 文件，不实现 SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。
## 五条主线

### 1. `Runtime Service / Model Gateway`

目标：把现有 CLI runtime、进程内 gateway、route 识别和 smoke gate 收口为明确的 provider registry、协议兼容层、本地运行、配置、启动、部署基础和模型 API 分发底座。

状态：`scripts/run-copilot-inference.py`、`services/gateway/copilot_gateway.py`、`services/runtime/inference_provider.py`、`services/runtime/provider_registry.py`、`services/platform/`、`RadishFlow` gateway demo 与 service smoke matrix 已具备基础骨架；当前 southbound 已通过统一 registry 收口 `mock`、`openai-compatible`、`HuggingFace`、`Ollama` 主入口与 `openai-compatible chat`、`gemini-native`、`anthropic-messages` 分流，`local_transformers` 则主要存在于 candidate/runtime 实验链路中。平台表层语言分工已固定为 `UI=React + Vite + TypeScript`、`Platform Service Layer=Go`、`Model Side=Python`。当前 `Go` 层已落最小服务启动、`/healthz`、`/v1/models`、`/v1/chat/completions`、`/v1/responses` 和 `/v1/messages` bridge，并补了第一版 SSE 流式兼容骨架、bridge-backed provider/profile inventory、`GET /v1/models/{id}` 精确 lookup、request-side provider/profile 选择、流式增量转发、`HuggingFace` / `Ollama` coverage。平台级 `ops smoke` 已固定 `go test ./...`、provider registry 与受控 profile inventory 门禁；本地启动 runbook、runbook drift check、脱敏配置摘要 / config check、JSON 配置文件层级、稳定本地启动 wrapper、最小 deployment smoke、结构化 diagnostics/failure boundary、provider/profile discoverability 对齐、request-level observability 与 error taxonomy 已补齐。`P1 Runtime Foundation` 已达到 short close；第一版 northbound 仍是窄切片，但继续横向扩同层配置、别名和兜底的收益已经下降，主要实现重心切到 `P2 Session & Tooling Foundation`。

下一步：不再继续把 `P1` 做成无限硬化阶段；`P2 Session & Tooling Foundation` 和 `P3 Local Product Shell / Ops Surface` 已有 metadata / blocked / read-only 产品外壳，Production Ops 静态边界也已收口。`Provider Runtime & Health v1` 已把 provider capability、health smoke、selection policy、provider-retry-fallback-policy-v1 与 docs refresh 固定成可检查口径，进入 close candidate；后续不再默认扩 provider runtime 同层切片。模型网关继续推进时，应围绕 API key、quota、trace、cost、tenant binding、secret ref readiness 和真实 production secret backend 的前置条件拆任务。

上述拆任务应优先沿 `Control Plane / User Workspace / Workflow v1` 任务卡推进，避免把模型网关、用户端和管理端混入当前本地 ops console。

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

下一步：真实模型产出、3B/4B 长跑、训练 JSONL、蒸馏和权重相关工作转入后置专题；保留现有 P4 v1 runbook、治理复核和预检结果作为未来重开依据。当前不因 Provider Runtime & Health v1 close candidate 而重开模型长跑。

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

### `P3 后续专题`：Production Ops Hardening v1

目标：把 P3 的 production secret backend、process supervisor、deployment environment isolation 和 console production packaging 缺口拆成可验证前置条件。

状态：已新增 [Production Ops Hardening v1 任务卡](task-cards/production-ops-hardening-v1-plan.md)。当前只处理配置、密钥边界、production secret backend contract、production-secret-backend-implementation-readiness、secret-ref-schema-and-fixtures、启动 / supervisor 边界、环境隔离、console production packaging smoke 和 P3 checklist alignment；不声明 production ready。`config-secret-boundary`、`production-secret-backend-contract`、`production-secret-backend-implementation-readiness`、`secret-ref-schema-and-fixtures`、`startup-supervisor-boundary`、`environment-isolation` 与 `console-production-package-smoke` 已分别用 production ops fixture / checker 固定为 governance boundary、implementation readiness 或 secret ref schema evidence，证据包括 `scripts/checks/fixtures/production-ops-config-secret-boundary.json`、`scripts/checks/fixtures/production-ops-secret-backend-contract.json`、`scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`、`contracts/production-secret-reference.schema.json`、`scripts/checks/fixtures/production-secret-reference-basic.json`、`scripts/checks/fixtures/production-ops-startup-supervisor-boundary.json`、`scripts/checks/fixtures/production-ops-environment-isolation-boundary.json` 和 `scripts/checks/fixtures/production-ops-console-package-smoke.json`；`scripts/checks/fixtures/p3-local-product-shell-short-close-checklist.json` 已完成 `short-close-checklist-refresh`，跨读这些 boundary fixture 并确认 production secret backend 仍为 not_satisfied，process supervisor、deployment environment isolation 和 console production packaging 也仍为 `not_satisfied`。`secret-ref-schema-and-fixtures` 只固定 committed secret reference schema 和 fixture，不保存 secret value、不实现 resolver、不接云、不声明 production ready。已新增 [Production Ops Docker Deployment v1 计划](task-cards/production-ops-docker-deployment-v1-plan.md)，并用 `docker-deployment-mode-definition` 固定 Radish 风格 docker local/test/prod 部署方向；`docker-local-compose` 已用 `scripts/checks/fixtures/production-ops-docker-local-compose.json` 固定为本地容器 smoke 资产；`docker-test-prod-compose` 已用 `scripts/checks/fixtures/production-ops-docker-test-prod-compose.json` 固定测试 / 生产共用部署态 compose 边界，资产包括 `deploy/docker-compose.yaml` 和 `deploy/.env.example`；`docker-image-build-publish` 已用 `scripts/checks/fixtures/production-ops-docker-image-build-publish.json` 固定镜像命名与 tag 后缀策略，真实发布 workflow 仍未实现；`deployment-readiness-smoke` 已用 `scripts/checks/fixtures/production-ops-deployment-readiness-smoke.json` 固定 `docker compose config` / 静态展开检查；`container-smoke-runbook` 已用 `scripts/checks/fixtures/production-ops-container-smoke-runbook.json` 固定容器 smoke 命令和清理边界；`container-smoke-record-template` 已用 `scripts/checks/fixtures/production-ops-container-smoke-record-template.json` 固定运行记录字段和 `tmp/production-ops/container-smoke/` 证据根。2026-05-26 已完成一次 `docker_local` container smoke 运行记录，`container_smoke_ready` 进入可审计证据状态；测试环境 smoke 和生产前复核仍未实现。

进入条件：已满足。P3 本地只读产品壳已经可用，且 checklist 已明确 production hardening 缺口。

下一步：该专题的静态边界已 close，且已补一次本地 `docker_local` container smoke 运行记录。后续只有在明确测试或生产前复核窗口后，才执行测试环境 smoke 或 production preflight 记录；不把已固定的部署态 compose、镜像命名策略、静态展开检查、runbook 或本地 mock smoke 解释为 secret backend、process supervisor、镜像发布或 production ready。

停止线：不实现真实 secret backend、不实现 process supervisor、不新增 executor、confirmation、writeback、replay 或 materialized result reader；不把 local-smoke、mock provider、demo profile 写成 production ready。

### `Runtime 后续专题`：Provider Runtime & Health v1

目标：把 provider registry、provider/profile inventory、request-side selection、diagnostics 和 error taxonomy 收口为可解释、可检查、可继续接真实 provider 的 runtime/health 层。

状态：已新增 [Provider Runtime & Health v1 任务卡](task-cards/provider-runtime-health-v1-plan.md)。当前只做 provider capability matrix、provider health smoke、provider selection policy、provider retry/fallback policy 和相关文档刷新；不实现真实 tool executor、confirmation / writeback / replay、训练、production secret backend 或 production ready。`provider-capability-matrix-v1` 已落地为 `scripts/checks/fixtures/provider-capability-matrix-v1.json` 与 `scripts/check-provider-capability-matrix.py`；`provider-health-smoke-v1` 已落地为 `scripts/checks/fixtures/provider-health-smoke-v1.json` 与 `scripts/check-provider-health-smoke.py`；`provider-selection-policy-v1` 已落地为 `scripts/checks/fixtures/provider-selection-policy-v1.json`、`scripts/check-provider-selection-policy.py` 和 Go selection 单元测试；`provider-retry-fallback-policy-v1` 已落地为 `scripts/checks/fixtures/provider-retry-fallback-policy-v1.json` 与 `scripts/check-provider-retry-fallback-policy.py`，并用 Go 失败路径测试固定 `retry_policy=caller-managed`、`fallback_policy=disabled`；`provider-runtime-docs-refresh` 已落地为 `scripts/checks/fixtures/provider-runtime-docs-refresh.json` 与 `scripts/check-provider-runtime-docs-refresh.py`。五者均已接入快速仓库检查，Provider Runtime & Health v1 进入 close candidate。

下一步：不再默认扩 provider runtime 同层切片；后续 retry/fallback execution、optional live health、container smoke 或 production secret backend 只作为明确独立任务重开。

停止线：capability 不等于 health；health smoke 不等于 production readiness；默认检查不联网、不要求 credential、不下载模型；不隐式 fallback，不把单一 provider 写成唯一方向。

### `P4`：Model Adaptation & Training

目标：在平台边界稳定后，定义首版基座、蒸馏和训练升级计划。

状态：前置计划已形成阶段证据并转入后置专题。v1 模型能力目标、teacher/student 边界、样本分层、晋级门槛、预检 runbook、治理复核记录和 `Qwen2.5-1.5B-Instruct` full-holdout-9 预检结果已有首版记录。raw student 在 docs QA 与 ghost completion 上通过，但在 `suggest_flowsheet_edits` 上 3/3 blocked；repaired comparison 可通过机器门禁，但只作为后处理证据。`Qwen2.5-3B-Instruct` CPU 单样本 probe 300 秒 timeout，当前不继续默认长跑。

### `P5`：Real Upstream Integration

目标：在上层项目具备真实挂载点后，选择首个切片完成真正接入。

状态：当前等待上层条件成熟。

### `P6`：User Workspace & Workflow Builder

目标：形成正式用户端，让用户创建 AI 应用、Prompt 应用、Workflow、Agent / Copilot 应用和 RAG / 知识问答应用，并查看 API key、调用量、运行记录和成本。

状态：边界任务卡已开始。当前 `apps/radishmind-console/` 只是本地 ops surface，不是用户端产品；`Control Plane / User Workspace / Workflow v1` 任务卡只定义用户工作区和 workflow builder 的资源边界、运行记录和停止线，不实现正式用户端或 executor。

### `P7`：Admin Control Plane & Radish Auth Integration

目标：形成正式管理端，管理租户、用户、角色、权限、provider/profile、模型路由、quota、price、secret、审计和部署状态，并作为 OIDC client 接入 `Radish`。

状态：边界任务卡已开始。当前只保留 Radish 对齐方向，不实现账号系统、不接真实 OIDC、不声明 production ready。默认技术栈为 Go control plane + Go gateway + Python model/eval/worker + TypeScript/Vite frontend，不新增 `.NET` 作为默认后端语言；任务卡只固定 OIDC client、tenant、quota、API key、secret ref、audit 和 deployment status 的前置边界。

## 下一步

1. 保持 `Provider Runtime & Health v1` close candidate：`provider-capability-matrix-v1`、`provider-health-smoke-v1`、`provider-selection-policy-v1`、`provider-retry-fallback-policy-v1` 与 `provider-runtime-docs-refresh` 已接入 fast baseline；后续不默认新增 provider 同层小切片。
2. 以 `Control Plane / User Workspace / Workflow v1` 任务卡作为下一条平台主线边界；已完成的 read-side 契约、fake-store-backed read handler plan / implementation、真实 auth/db 前置条件、consumer contract、正式 UI 边界、shared shell、七个只读页面、formal UI close、dev-only live consumer、auth/store transition、repository contract / guard / smoke / implementation / store selection / schema migration readiness、contract types、smoke runner、interface readiness、adapter readiness refresh、selector enablement preconditions、schema migration implementation preconditions、repository adapter implementation plan 和 schema artifact manifest readiness 继续作为证据保留。后续如继续 read-side，必须先补真实 schema artifact、selector、auth 和 adapter smoke 证据；不默认进入数据库、OIDC、token validation、repository migration、quota enforcement、billing、workflow executor、confirmation、writeback 或 replay。
3. 将 `Production Ops Hardening v1` 维持为 static boundary close + docker_local smoke recorded；后续只在明确测试或生产前复核窗口后补测试环境 smoke 或 production preflight 记录。
4. 将 `P3 Local Product Shell / Ops Surface` 与 UI 第二批维持在 `local usable / read-only close candidate`；不再默认补同类只读 console 小切片，除非真实使用暴露新缺口。
5. 将真实模型产出、3B/4B 长跑、训练 JSONL、蒸馏和权重相关工作保留为后置专题；没有 GPU / 明确实验窗口 / 新能力假设前不重开。
6. UI 后续扩张必须先回到设计稿或任务卡，不直接增加 confirmation、writeback、replay 或 production packaging。
7. 只为新增 API、执行边界、生产声明、数据格式、外部 provider 风险或高风险能力新增专项门禁；普通 UI 展示改动优先复用现有 console behavior / visual smoke / fast baseline。
8. 继续维持上层项目接入前置条件总表，不提前细化不存在的真实接线。
9. 产品机会池中近期优先只保留 `Model API Routing Studio`、`Workflow Marketplace`、`Eval-as-a-Feature`、`AI Cost Ledger` 和 `Policy Sandbox` 作为候选；进入实施前必须先写任务卡。

## 停止线

- 不把 repaired、injected、guided 或 builder 轨通过写成 raw 模型能力通过。
- 不把机器指标通过写成人工可接受度通过。
- 不在没有非重复能力假设时继续扩同一批 `M4` 实验。
- 不在上层项目没有真实挂载点时继续细化假想接线设计。
- 不把 `P1` 继续扩成无止境的 provider/config/diagnostics 细化阶段。
- 不把当前本地 console 写成正式用户端、生产管理端、完整 Dify-like workflow builder 或完整模型 API 分发平台。
- 不自建与 `Radish` 冲突的身份、权限、数据库和部署真相源；Radish OIDC client 接入必须作为独立任务推进。
- 不把 Control Plane、Gateway 和 Workflow Executor 混成单体，也不把微服务拆分变成新增语言栈的理由。
- 不把 P3 继续扩成无止境的本地只读 console 小切片阶段。
- 不把 Production Ops 继续扩成无止境的静态 governance fixture 阶段。
- 不把 provider health smoke 写成 production readiness 或隐式 provider fallback。
- 不把当前本地 console 壳扩成 production console、大面积复杂交互或真实确认 / 写回 / replay UI。
- 不把真实模型产出、3B/4B 长跑、训练 JSONL、蒸馏或权重相关工作作为当前默认主线。
- 不让模型直接写上层业务真相源。
- 不用晦涩抽象、空泛 helper 或多层 fallback 掩盖代码职责不清。
