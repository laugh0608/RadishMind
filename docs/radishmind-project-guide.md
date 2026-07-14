# RadishMind 项目总览与使用指南

更新时间：2026-07-12

## 这份文档讲什么

这是一份面向新读者的项目说明书，回答四个问题：

- `RadishMind` 现在被定义成什么
- 仓库怎么分工
- 今天可以怎样实际运行它
- 如果你要继续推进开发，先看哪几条主线

它不替代 `docs/radishmind-current-focus.md`、`docs/devlogs/` 或任务卡，也不记录阶段推进流水。

2026-06-14 起，具体功能或长期开发目标先看 [功能设计文档入口](features/README.md)。任务卡只承载实现批次、前置条件或高风险边界，不再作为功能默认主文档。

## 项目定位

`RadishMind` 现在的正式定位是：

`Radish` 体系下的 AI 工具、工作流、模型网关和 Copilot 集成平台。

它不是上层业务真相源，不替代 `RadishFlow`、`Radish` 或 `RadishCatalyst` 的业务决策权，也不是“只放模型实验”的仓库。

当前主要职责是：

- 提供用户端 AI 应用 / 工作流 / API key / 调用记录工作台。
- 提供管理端 provider/profile / 模型路由 / 租户 / quota / secret / 审计控制面。
- 提供模型 API 分发和兼容网关。
- 接收结构化上下文
- 运行最小推理链路
- 兼容多种上游模型接入方式与多种下游协议接口
- 提供本地只读产品发现面，供 console 或上层 UI 读取平台可展示能力和停止线
- 组织局部工具、规则和响应收口
- 输出解释、诊断、结构化建议和候选动作
- 维护统一协议、评测门禁、审计记录和训练治理

## 实现分工

- `UI`：`React + Vite + TypeScript`
- `平台服务层`：`Go`，覆盖 `HTTP API`、`gateway`、control plane、鉴权、流式转发、长驻进程、观测和部署壳
- `模型侧`：`Python`，覆盖训练、评测、`prompt / builder`、离线推理和校验逻辑
- `contracts/`：唯一 canonical protocol 真相源，所有语言只能消费，不得各自重新定义业务协议
- `Radish` 集成：部署、数据库、登录 / 授权默认对齐 `Radish`；未来作为 OIDC client 接入 `Radish`；不默认引入 `.NET`

## 当前四个产品面

1. `User Workspace`：AI 应用、Workflow、Agent / Copilot、RAG、API key、调用量、成本和运行记录。
2. `Admin Control Plane`：租户、用户、角色、权限、provider/profile、模型路由、quota、secret、审计和部署状态。
3. `Model Gateway / API Distribution`：OpenAI-compatible / Responses / Messages / Models API，多 provider / profile / model 分发。
4. `Workflow / Agent Runtime`：Prompt、LLM、HTTP tool、RAG retrieval、condition、output 和受控 agent loop。

这些产品面及 Image Path 的后续开发入口已整理到 `docs/features/`。

## 当前五条主线

1. `Runtime Service`：本地启动、gateway、route、provider/profile、协议兼容、响应封装、部署基础；当前已达到 short close，request observability 和 error taxonomy 已进入平台门禁。
2. `Conversation & Session`：会话标识、历史压缩、恢复和审计边界；当前已有 session record、recovery checkpoint record/manifest/read result、northbound session metadata、metadata-only route smoke、confirmation / audit / result / executor / storage 设计门禁、short close readiness delta、stop-line manifest 和 close-candidate status summary。
3. `Tooling Framework`：检索、附件解析、候选生成、builder、tool policy 和 audit；当前已有 tool contract、registry、audit record、session binding、metadata-only result cache、result materialization policy、executor boundary、storage backend design、deny-by-default gates 和 executor/storage/confirmation enablement plan。
4. `Evaluation & Governance`：schema、smoke、offline eval、review、promotion gate、负向消费 summary、route smoke coverage summary、readiness summary、implementation preconditions、negative regression governance suite、negative coverage rollup、route negative coverage matrix 和 readiness consistency rollup。
5. `Model Adaptation`：基座选型、prompt/runtime 协同、蒸馏、训练样本治理和模型晋级。

当前可运行的开发测试产品路径已经覆盖：Gateway 三协议调用与 sanitized Request History；Application API Integration、Configuration Draft、Publish Candidate Review；Saved Workflow Draft、受控 Executor、durable Run History 与 Evaluation；Admin verified identity、Tenant / Audit PostgreSQL read repository，以及 deterministic OIDC verifier。长期契约入口分别见 [服务/API 接入契约](contracts/service-api.md)、[Control Plane Read-Side 契约](contracts/control-plane-read-side.md) 和 [Radish OIDC Token Validation 契约](contracts/radish-oidc-token-validation.md)。这些路径仍由显式 dev/test gate 保护，不是公开 production API。

`Saved Workflow Draft v1` 已实现 platform Go domain service、memory dev store、dev-only HTTP route 和 web consumer，并进一步具备 formal store selector、静态 schema artifact、repository adapter、adapter smoke execution 与 production auth runtime bridge。Draft Designer 的 `version_conflict` 读法是保留当前本地 active draft、刷新当前 application 的 sanitized saved draft list、允许用户继续本地草案或显式恢复 saved version，并把同一份 conflict review summary 交给 Review Handoff；它不自动覆盖、不自动合并，也不表示 durable persistence、publish、run 或 executor ready。database connection / schema marker preconditions、connection provider entry review / entry refresh v2、database secret resolver readiness / entry review / runtime dependency refresh、database driver / DSN / TLS policy readiness、database role policy readiness、database connection smoke strategy、connection lifecycle readiness、schema marker runtime dependency refresh、Radish OIDC upstream evidence refresh 和 token validation auth middleware runtime entry review 只说明 future durable store 接入前的阻塞条件；它们不选择或导入 DB driver、不解析 secret、不构造 DSN、不创建 TLS runtime、不创建 connection factory、不创建 role runtime、不执行 connection smoke、不执行 SQL、不启用 repository mode，也不创建 OIDC middleware、token validator、membership adapter 或 production API。

2026-07-11 覆盖：Saved Draft 已完成显式 `postgres_dev_test` repository；Workflow Executor v0 已完成 dev-only POST / GET route、服务端图准入、Gateway advisory 调用、tenant / workspace / application scoped 进程内 run record 和 Web 创建 / 保存 / 执行 / 回读。上段“不表示 durable persistence / run / executor ready”只描述早期阶段与 production / unrestricted 边界；当前仍不开放 production repository、OIDC、tool、confirmation commit、writeback、replay / resume 或公开生产 API。

2026-07-12 覆盖：Gateway Request History、Application Configuration Draft、Application Publish Candidate 和 Admin Tenant / Audit read 均支持显式 PostgreSQL dev/test repository、manual migration、marker / checksum、runtime role、no-fallback 与重启恢复。Control Plane auth 已支持 signed test token 和 `radish_oidc_integration_test`；后者只开放 Tenant Summary / Audit，workspace operation 因 membership 未成立而 fail closed。真实 Radish 联调已 deferred，不把 deterministic issuer 或本地浏览器路径解释为真实接入。

`contracts/radish-oidc-token-validation.schema.json` 固定 future workspace membership / repository actor context 的 verified token context 脱敏投影。它只允许 `issuer_ref`、`subject_ref`、`tenant_ref`、audience / scope / workspace / application refs、时间戳、policy version、request id 和 audit ref，显式拒绝 raw token / claims、cookie、JWKS dump、membership raw record 和 secret。当前 Admin OIDC runtime 使用内部 sanitized `VerifiedControlPlaneIdentity`，不会用该 schema 绕过 workspace membership；两者关系见 [Radish OIDC Token Validation 契约](contracts/radish-oidc-token-validation.md)。

`Workflow Node Designer Surface v1` 现在是 Workflow Builder 体验的 active-draft 画布专题，位置在 Draft Designer / Review Handoff 之后、publish / run / executor 之前。它已在 `apps/radishmind-web/` 接入 `@xyflow/react`，把 active draft 派生为可拖拽节点、typed handle、custom edge、inspector、validation overlay navigation、mapping summary 和 Review Handoff evidence；节点位置通过 `additional_fields.designer_layout_v1` 保存为受控 UI metadata，画布新增 / 删除连线只修改 active draft 的 `draft.edges`。这些能力仍只服务本地草案设计与审查，不把 `Preview Plan`、`valid_for_review`、节点拖拽或 edge mutation 写成可运行、可发布、durable store 或 production API 状态。

Production secret backend 当前仍只到说明、检查、metadata-only artifact 和 test-only 离线替身层：secret reference schema、config secret ref readiness、provider profile binding、disabled resolver interface、operator runbook / negative gates、rotation / audit policy、test-only fake resolver runtime、真实 resolver runtime preconditions / entry review / refresh、backend profile selection、real resolver no leakage strategy / runtime entry review / refresh、credential handle boundary / runtime entry review / refresh、operator approval evidence / runtime entry review / refresh、cloud secret service selection readiness、backend health boundary / runtime entry review / refresh、audit store handoff / contract / ownership / delivery-idempotency、runtime event schema artifact、runtime blocker matrix、durable backend selection、writer / idempotency / delivery runtime entry review、storage adapter metadata contract artifact、static product class selection、database provider policy readiness、append-only logical table schema boundary、managed product profile review、reference-only provider profile review 和 provider account / resource / endpoint review 都已可检查。`services/platform/internal/secretbackend` 只提供 test-only、默认 disabled 的 fake resolver runtime；开发者说明见 [Production Secret Backend Fake Resolver Runtime Implementation v1](platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md)。`contracts/production-secret-audit-storage-adapter.metadata-contract.json` 只描述 future storage adapter 的 metadata-only input / result envelope，不是运行时存储记录；`managed_database_append_only_table`、`managed_postgresql_compatible_audit_store_profile` 和 `managed_postgresql_compatible_provider_reference` 都只是静态 / reference-only 选择证据，不是具体数据库、vendor、cloud product、provider account、provider resource 或 database endpoint。当前仍不创建 production resolver runtime、backend health runtime、no secret leakage smoke runtime、cloud secret service、credential handle runtime、approval runtime、audit store runtime、audit writer runtime、storage adapter runtime、delivery runtime、idempotency runtime、database provider、driver、DSN parser、SQL、schema marker runtime、migration runner、repository mode 或 production API。`Provider Runtime & Health v1` 已完成 `provider-capability-matrix-v1`、`provider-health-smoke-v1`、`provider-selection-policy-v1`、`provider-retry-fallback-policy-v1` 和 `provider-runtime-docs-refresh` 五个可检查切片并进入 close candidate，不继续默认新增 provider 同层小切片。P2 停止线继续作为背景证据保留，不代表真实 executor、durable store、confirmation 接线、materialized result reader、长期记忆、业务写回或 replay 已经完成。

`Model Gateway / API Distribution` 的当前产品 UI 也已进入离线证据组织层：Model Gateway Overview、Route Evidence、Usage/Audit Evidence 和 Evidence Review / Readiness 都位于 `apps/radishmind-web/`，复用 read shell、API key summary、quota summary、run history、audit log、provider runtime 与 gateway readiness 证据。它们只解释 northbound API surface、provider/profile、route binding、selection case、key scope、quota / cost snapshot、trace / failure、audit decision、readiness rollup、evidence checklist 和 locked capability，不新增真实网关 route、production gateway、secret resolver、API key lifecycle、quota enforcement、cost record write、retry/fallback execution、数据库、Radish auth 或 repository adapter。

`Admin Operations Review / Readiness` 是同一只读产品壳中的管理端汇总面：它复用 tenant overview、audit log、Model Gateway Evidence Review 和 Production Ops 静态证据，把 readiness、evidence checklist、operational risk 和 boundary lock 放在一个审查视图里。`Admin Provider/Profile & Deployment Evidence Review / Readiness` 继续复用 Model Gateway route / review、Admin Operations、tenant overview 和 audit log，把 provider/profile readiness、model route readiness、secret / deployment evidence、operator risks 和 locked capabilities 组织成管理端阅读路径。它们都不代表 production admin console，不提供 tenant mutation、provider/profile mutation、model route change、raw audit export、deployment preflight、secret resolver、workflow executor、writeback 或 replay。

`Image Path` 当前已有 metadata-only response builder 链路：`services/runtime/image_artifact_runtime_mapper.py` 只消费 `image_generation_artifact` metadata，`services/runtime/image_artifact_response_consumer.py` 只把 mapper 成功结果合并为现有 `CopilotResponse.citations` artifact citation，`services/runtime/inference_response.py#coerce_response_document` 只从 request artifact metadata 发现并接入该链路。它用于验证 `artifact://`、sha256、mime type、dimensions、safety review、provenance 和 fail-closed 语义，不读取图片二进制、不查 artifact store、不解析 public URL、不调用真实生图 backend、不上传 artifact、不改 response schema。

完整 production 用户端 / 管理端、workspace membership、正式 application promotion、production API key / quota / billing、production secret backend 和公开生产 API 仍未实现；当前 `apps/radishmind-web/` 是离线优先、可显式连接 dev/test runtime 的产品 UI，`apps/radishmind-console/` 仍只是本地 ops surface。

## 目录速览

- `docs/`：正式文档源
- `docs/features/`：功能设计与开发文档
- `contracts/`：JSON Schema 真相源
- `scripts/`：检查、运行、转换、评测和最小运维入口
- `deploy/`：Docker local / test / prod 部署边界说明和 compose 资产
- `datasets/`：eval 样本、示例对象和 candidate record
- `training/`：训练治理、实验说明和复核记录
- `services/`：runtime 与 gateway 实现
- `adapters/`：上游项目适配层
- `tmp/`：本地临时产物，不提交

## 今天能怎么运行

### 1. 直接跑最小推理链路

当前最小运行入口是 `scripts/run-copilot-inference.py`。它不是长驻服务，而是单次 CLI runtime。

```bash
./scripts/run-python.sh scripts/run-copilot-inference.py \
  --sample datasets/eval/radishflow/suggest-flowsheet-edits-reconnect-outlet-001.json \
  --provider mock \
  --response-output tmp/rf-suggest-edit.response.json
```

如果后续要接真实 provider，再显式传 `--provider openai-compatible|huggingface|ollama`、`--provider-profile`、`--model`、`--base-url`、`--api-key`。当前这条入口已经能按 profile 分流到 `openai-compatible chat`、`gemini-native` 和 `anthropic-messages` 三类上游协议；`services/platform/` 也已把 `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models`、`/v1/platform/overview`、`/v1/platform/local-smoke`、SSE bridge、provider/profile inventory、request-side selection、`HuggingFace` / `Ollama` coverage、本地启动 wrapper、JSON 配置层级、deployment smoke、diagnostics / failure boundary 和 discoverability 对齐纳入第一版 runtime foundation。

### 2. 跑进程内 gateway demo

如果你要看 `RadishFlow export -> adapter -> request -> gateway` 整条链路，当前正式入口是：

```bash
./scripts/run-python.sh scripts/run-radishflow-gateway-demo.py \
  --check-summary scripts/checks/fixtures/radishflow-gateway-demo-summary.json
```

这条链路当前使用 mock provider，只验证 route、schema、advisory-only 和 `requires_confirmation` 等不变量。

### 3. 跑 service/API smoke matrix

当前 `RadishFlow` 的正式服务门禁入口是：

```bash
./scripts/run-python.sh scripts/check-radishflow-service-smoke-matrix.py \
  --check-summary scripts/checks/fixtures/radishflow-service-smoke-matrix-summary.json
```

这条矩阵会一起核对：

- CLI runtime
- gateway API
- gateway demo
- UI consumption
- candidate handoff

它现在是仓库里最接近“服务切片验收”的正式门禁。

### 3.5 运行 Go 平台服务层

当前 `Go` 平台服务层已落在 `services/platform/`，用于承载 `HTTP API`、`gateway`、鉴权、流式转发、观测和部署壳。日常本地运行优先使用 wrapper：

```bash
./scripts/run-platform-service.sh config-check
./scripts/run-platform-service.sh diagnostics
./scripts/run-platform-service.sh serve
```

Windows / PowerShell 使用对应的 `pwsh ./scripts/run-platform-service.ps1 config-check|diagnostics|serve`。

wrapper 默认使用 `local-product` 档，把七组本地运行数据统一写入仓库根 `var/sqlite-dev/radishmind.db`，并开启对应开发门禁；配置摘要不会输出绝对路径。需要执行 PostgreSQL 专项验收或组件故障注入时，Shell 使用 `--profile configured`，PowerShell 使用 `-Profile configured`，该档不自动注入聚合持久化配置。

当前 Platform 除 `/healthz`、overview / local-smoke、models、三协议 northbound、session/tooling 与七条 Control Plane Read-Side route 外，还注册 Workflow Draft / Run / Evaluation、Application Draft / Publish Candidate 和 Gateway Request History dev/test route。完整路由与 gate 见 [Platform README](../services/platform/README.md)；路由注册不等于默认开放。

Control Plane Read-Side 支持 `dev_headers`、`signed_test_token` 和 `radish_oidc_integration_test` 三种显式开发测试 auth mode。`postgres_dev_test` 只承载 Tenant Summary / Audit；OIDC integration 下其余 workspace operation 返回 `workspace_membership_unavailable`，不会读取 fake repository。默认 disabled、非法组合、缺少 identity / permission / tenant binding 或 identity provider 不可用都 fail closed。

read-side repository 当前由显式 selector 路由：fake store 继续承载开发测试 workspace summary，PostgreSQL repository 只承载 Tenant Summary / Audit。PostgreSQL 模式要求 manual migration、marker / checksum、read-only runtime role、strict keyset cursor 与 startup SELECT preflight；连接、权限或 marker 失败不回退 fake store。

当前 OIDC verifier 是受控 integration-test runtime，不是 production auth。它校验 exact issuer / audience、algorithm allowlist、required `kid`、discovery / JWKS origin 与响应边界、single-flight refresh、unknown-kid 单次刷新、bounded cache / rotation / hard expiry、required claims、时间窗口、tenant binding 和版本化 permission mapping；raw token / claims 不进入 handler、repository、日志或 UI envelope。

这仍然不是 production deployment：它已经能作为本地平台服务切片运行和诊断，但尚未具备生产级 secret backend、进程监管、环境隔离和正式发布包。

### 3.6 运行本地 console 产品壳

本地 console 位于 `apps/radishmind-console/`，默认前端端口为 `4000`，后端平台端口为 `7000`。最省事的开发入口是从仓库根目录运行：

macOS / Linux / WSL 使用：

```bash
./scripts/run-radishmind-console-dev.sh
```

Windows / PowerShell 使用：

```powershell
pwsh ./scripts/run-radishmind-console-dev.ps1
```

该入口会启动或复用 platform 后端和 console 前端，并探测 `/healthz`、`/v1/platform/overview`、`/v1/platform/local-smoke`、本地 CORS preflight 和前端页面。它不是 production supervisor，不负责长期守护进程，也不实现 executor、durable store、confirmation、业务写回或 replay。

如果只想验证已有 platform 服务的本地 readiness，可运行：

```bash
./scripts/run-python.sh scripts/run-platform-local-smoke.py \
  --base-url http://127.0.0.1:7000 \
  --check
```

console 页面当前直接消费 `/v1/platform/overview` 与 `/v1/platform/local-smoke`，展示 Runtime overview、Service Status、Model Inventory、Provider/Profile Details、Session And Tooling、Blocked Action Detail、Dev Diagnostics、Local Readiness、Stop Lines 和 Audit Boundary。它仍是本地只读 ops surface，不是 production console、正式用户端或生产管理端。

### 3.7 运行产品 UI shell（开发测试态）

正式产品 UI 的当前实现位于 `apps/radishmind-web/`。它默认离线，显式 dev-only 模式可分别连接 Control Plane Read、Saved Draft / Executor、Gateway Playground / History、Application Catalog、Application Configuration Draft 和 Application Publish Review。Application Catalog 已支持创建、编辑和归档；API 密钥页面仍只展示脱敏摘要，签发、Gateway Bearer 认证与吊销应按[应用目录与 API 密钥开发测试指南](features/user-workspace/application-catalog-api-key-dev-test-guide.md)通过 HTTP API 验证。RadishFlow Copilot 与 Radish Docs Assistant 的离线样例仍由统一 fixture 防止漂移；任何 dev/test live path 都不能解释为 production API consumer、正式 application 发布、生产 API key / quota、production repository 或完整 workflow runtime ready。

日常预览或前后端联调优先使用仓库根目录启动脚本，不再手动拼接环境变量：

```bash
./start.sh web-live
./start.sh web-offline
```

Windows / PowerShell 使用：

```powershell
pwsh ./start.ps1 -Command web-live
```

`web-live` 会启动或复用 Platform 与产品 UI。按使用目标显式组合 `--saved-draft-dev` / `--saved-draft-postgres-dev-test`、`--gateway-request-postgres-dev-test`、`--application-draft-dev`、`--application-publish-dev`、`--application-publish-postgres-dev-test` 或 `--application-catalog-postgres-dev-test`；launcher 会设置对应 HTTP/write gate、consumer source 和 migration status preflight。完整命令见 [Web README](../apps/radishmind-web/README.md)。它不是 production supervisor，不启用 production auth、正式 promotion、API 密钥 Web 生命周期、quota enforcement、tool、confirmation、writeback 或 replay。

如果 macOS `Control Center` / AirPlay 占用了默认 backend 端口 `7000`，改用备用本地端口启动：

```bash
./start.sh web-live --backend-url http://127.0.0.1:7100
```

PowerShell 使用：

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -BackendUrl http://127.0.0.1:7100
```

```bash
cd apps/radishmind-web
npm run build
```

如需本地查看页面，可在明确需要时运行：

```bash
cd apps/radishmind-web
npm run dev
```

当前页面包括：

- route catalog、shared states 和 forbidden output guard
- `admin-tenant-overview`
- `admin-audit-log`
- `workspace-applications`
- `workspace-api-keys`
- `workspace-usage-quota`
- `workspace-workflow-definitions`
- `workspace-run-history`
- workflow application detail
- workflow definition detail
- workflow run detail
- workflow blocked action preview
- workflow confirmation placeholder
- Workflow Draft Designer（默认离线，可显式使用 dev-only saved draft consumer）
- workflow offline draft validation inspector
- workflow execution plan preview
- workflow runtime readiness inspector
- Workflow Surface Overview
- workflow workspace context selection
- Workflow Scenario Inspector
- Workflow Review Workspace
- User Workspace Home
- Workflow Review Handoff
- Model Gateway Overview
- Model Gateway Route Evidence
- Model Gateway Usage/Audit Evidence
- Model Gateway Evidence Review / Readiness
- Model Gateway Request History
- Gateway Playground
- Application API Integration
- Application Configuration Draft
- Application Publish Review
- Admin Operations Review / Readiness
- Admin Provider/Profile & Deployment Evidence Review / Readiness

七个 read-side summary 页面展示 route metadata、request / audit ref、状态预览和脱敏 summary；默认使用离线 view model。dev-live 按 auth/store mode 读取：signed test 模式可组合 PostgreSQL Admin read 与 workspace fake binding，OIDC integration 只读取 Tenant Summary / Audit，五条 workspace operation 显示 membership unavailable。workflow function surface 继续复用这些 summary，并可显式连接 Saved Draft、Executor 与 durable Run History；这些能力仍不开放 production membership、confirmation decision、writeback、run replay 或 run resume。

Model Gateway 的读法是先看 Overview / Route / Usage-Audit / Evidence Review，再进入 Playground 发起三协议 unary / stream 请求，最后按同一 request id 打开 sanitized Request History。Application Detail 侧可先在 API Integration 选择模型和协议，再进入 Configuration Draft 保存 / 比较配置，最后在 Publish Review 创建不可变 candidate 并记录审查决定；approved candidate 仍显示 promotion blockers，不修改正式 application。

Draft Designer 的保存路径需要额外区分：默认仍展示 sample / local draft；显式 Saved Draft 开关可写入 `memory_dev` 或 `postgres_dev_test`。同一 launcher 还会设置 `RADISHMIND_WORKFLOW_EXECUTOR_DEV=1` 与 `VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE=dev-workflow-executor-http`，允许已保存、未修改且合规的 executor v0 草案运行并回读 record。该路径用于开发 / 测试，不是 production persistence、production API、publish 或 unrestricted runtime。

Saved draft 冲突处理的本地读法：当保存返回 `version_conflict` 时，Draft Designer 保留本地 active draft，刷新当前 application 的 saved draft list，并显示 saved version metadata、validation state、blocked capability count 和下一步选择。继续本地草案会进入 `conflict_local_continued`，下一次保存以当前 saved version 作为 expected version；恢复 saved version 必须由用户从 refreshed list 显式触发。Review Handoff 只把 conflict review summary 作为 advisory-only 审查证据展示，不保存 handoff、不自动合并、不覆盖本地草案、不解锁 publish / run / confirmation / writeback。

节点画布读法：先读 [Workflow Node Designer Surface v1](features/workflow/node-designer-surface-v1.md)，再按需进入 saved draft mapping、persisted layout、edge editing preconditions 和 Review Handoff 专题。当前画布可以选择节点、拖拽布局、受控新增 / 删除 active draft edge、聚焦 validation finding，并让 validation inspector / Preview Plan / Review Handoff 消费 mutation 后的 active draft；不能把这些 UI mutation 写成 executor、publish、run、durable store、repository mode 或 production API 的已实现能力。

Graph review handoff 的读法：在 Workflow Review Handoff 中先看 summary count，确认 node-targeted、edge-targeted 与 graph-level finding 的数量；再按分组查看每条 finding 的 source check、severity、target refs、summary 和 reviewer question。node-targeted finding 用于回到具体节点和 inspector；edge-targeted finding 用于核对 active draft edge 与条件摘要；graph-level finding 用于查看全局 blocked capability、runtime readiness 或停止线。这个面板只帮助 reviewer 建立审查上下文，不保存 handoff、不持久化 validation focus、不生成 publish / run / confirmation / writeback 操作。

### 3.8 使用 Image Path metadata-only runtime mapper

图片生成路径当前只允许 metadata-only runtime helper。开发者可以在离线检查或后续 response consumer 评审中导入：

```python
from services.runtime.image_artifact_runtime_mapper import (
    map_image_artifact_to_response_reference,
)
```

输入必须是已符合 `contracts/image-generation-artifact.schema.json` 的 `image_generation_artifact` metadata。成功结果只返回 artifact citation 和 metadata reference；失败结果只返回 fail-closed failure code，不返回 citation。该 helper 不打开文件、不访问网络、不读取图片二进制、不查 store、不生成 public URL、不调用 backend。完整契约见 [图片生成契约](contracts/image-generation.md)。

### 3.9 使用 Docker 部署资产

Docker 资产位于 `deploy/`，长期说明见 [部署目录说明](../deploy/README.md)。当前固定三种模式：

- `host_dev`：日常开发仍在宿主机直跑，使用 `scripts/run-platform-service.*` 和 `scripts/run-radishmind-console-dev.*`。
- `docker_local`：使用 `deploy/docker-compose.local.yaml` 本地 build platform / console 镜像，默认 `mock` provider，只用于本地容器 smoke。
- `docker_test` / `docker_prod`：共用 `deploy/docker-compose.yaml`，只拉取预构建镜像，通过 `RADISHMIND_IMAGE_TRACK=test|release` 或固定 `RADISHMIND_IMAGE_TAG` 区分环境。

本地容器 smoke 的命令边界是：

```bash
docker compose -f deploy/docker-compose.local.yaml config
docker compose -f deploy/docker-compose.local.yaml up --build -d
./scripts/run-python.sh scripts/run-platform-local-smoke.py --base-url http://127.0.0.1:7000 --check
docker compose -f deploy/docker-compose.local.yaml ps
docker compose -f deploy/docker-compose.local.yaml down
```

执行后应按 `scripts/checks/fixtures/production-ops-container-smoke-record-template.json` 把运行记录写入 `tmp/production-ops/container-smoke/`。该记录目录不提交。

部署态 compose 的静态检查和 runbook 检查已经纳入 `./scripts/check-repo-fast.sh` 与 `pwsh ./scripts/check-repo.ps1 -Fast`。这些检查默认不启动 Docker、不拉镜像、不声明 `container_smoke_ready`。

### 4. 跑本地候选模型输出

如果你要继续看 `RadishMind-Core` 本地候选输出，入口仍是：

```bash
./scripts/run-python.sh scripts/run-radishmind-core-candidate.py \
  --provider local_transformers \
  --model-dir /path/to/model \
  --output-dir tmp/radishmind-core-candidate-local \
  --allow-invalid-output \
  --validate-task
```

这仍属于模型评测 / 训练前治理链路，不等同于平台正式服务。

### 5. 当前协议兼容边界

当前平台必须区分两类兼容：

- 南向模型接入：平台去调用哪些模型和哪些 provider
- 北向协议输出：外部客户端如何调用 `RadishMind`

当前真实状态是：

- 南向已有一部分：`openai-compatible` 主入口、`HuggingFace`、`Ollama`、`gemini-native`、`anthropic-messages`，以及评测链路中的 `local_transformers`
- 北向已有第一版兼容面：`/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/models`、`/v1/models/{id}`、overview / local-smoke、session/tooling、SSE bridge、provider/profile selection metadata 和 diagnostics discoverability；本地产品 API 另有 gated Workflow、Application、Gateway History 与七条 Control Plane Read-Side route
- `P1 Runtime Foundation` 已达到 short close，当前不应继续把 provider/config/diagnostics/observability 同层细节当作主线
- 当前仍是窄切片：已经具备 Docker local/test/prod 的静态部署边界、镜像命名治理、deployment readiness 静态 smoke、container smoke runbook、运行记录模板、一次 `docker_local` container smoke 运行记录、provider capability matrix、provider health smoke、provider selection policy、provider retry/fallback policy 和 `provider-runtime-docs-refresh` 文档收口，但还缺真实镜像发布 workflow、production secret backend、部署隔离、外部 provider live health check、retry/fallback execution、console runtime config、测试环境 smoke、生产前复核记录，以及 session/tooling 的真实确认流接线、存储、执行和完整负向回归；P3 checklist 已把本地只读产品壳标为可用，并把这些生产前置条件继续保持为 `not_satisfied`，P2 现有 route / gate / negative regression 资产仍是 governance-only。没有测试或生产前复核窗口时，不把 Docker / deployment 当作默认开发主线

## 当前还不能算完成的能力

当前仓库还没有这些正式能力：

- production deployment package 与镜像发布 workflow
- production secret backend
- process supervisor 与环境隔离
- 外部 provider health check
- console production packaging / runtime config
- 完整 user workspace / production admin control plane React UI
- `apps/radishmind-web/` 的 production API consumer、Radish 登录 / 长期 session、workspace membership、production repository 和用户端 / 管理端 production 写入能力
- Control Plane Read-Side 的 production Radish OIDC、workspace membership adapter、五条 workspace PostgreSQL repository、production repository 与管理写入
- Workflow / Agent Runtime 的 production builder / repository / executor、validation result persistence、publish、confirmation decision、execution unlock、business writeback、run replay 和 run resume；当前 durable draft / run / evaluation 与 executor 都只属于受控 dev/test
- Image Path 的 artifact store、binary reader、public URL resolver、真实 backend adapter、安全 runtime、图片生成和 artifact upload
- 测试环境 smoke 和生产前复核记录
- 更完整的 route-level smoke、stream 组合和兼容性矩阵
- durable session/checkpoint/audit/result store、materialized checkpoint/result reader 和 recovery runbook
- 真实工具执行器、materialized tool result cache、上层确认流接线和完整 session/tooling 负向回归 implementation consumer

所以如果你问“现在怎么部署”，准确答案是：仓库已支持本地 CLI、Go Platform + Python bridge、Web / Console launcher、Docker 静态部署边界，以及多组显式 `memory_dev` / `postgres_dev_test` 产品 runtime。开发者可复验 Gateway 调用与 History、Application Draft / Publish Review、Saved Workflow Draft / Executor / Run / Evaluation、Admin signed-token 与 deterministic OIDC path。它们仍不是 production deployment：真实镜像发布、测试环境 smoke、production preflight、production secret backend、Radish 登录与 membership、production repository、正式 application promotion、confirmation、业务写回和 replay 均未开放。

## 读文档顺序

如果你刚接触这个仓库，建议按这个顺序读：

1. [文档入口](README.md)
2. 项目总览与使用指南
3. [当前推进焦点](radishmind-current-focus.md)
4. [产品范围](radishmind-product-scope.md)
5. [战略定义](radishmind-strategy.md)
6. [能力矩阵](radishmind-capability-matrix.md)
7. [系统架构](radishmind-architecture.md)
8. 按任务需要继续读 UI、集成契约、部署、脚本、数据集或训练专题。

## 默认不要做

- 不把 `RadishMind` 做成上层业务真相源
- 不默认把 builder/guided 结果当成 raw 晋级证据
- 不在上层项目还没具备真实挂载点时继续细化假想接线
- 不把 production deployment、session、tooling 或完整外部兼容矩阵写成“已经具备”
- 不默认下载大模型、数据集或权重
