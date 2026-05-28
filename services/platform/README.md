# RadishMind Platform Service Layer

本目录承载 `Go` 平台服务层的最小骨架。

当前职责：

- 启动最小本地 `HTTP` 服务
- 承载 northbound `API` / `gateway` 入口
- 通过 Python bridge 调用 canonical `CopilotGatewayEnvelope`
- 提供结构化诊断、观测、本地产品 overview、部署壳和后续鉴权 / 流式转发落点

当前明确不做：

- 不在这里复制第二套业务真相源
- 不在这里重写模型推理、训练、评测或 `builder`
- 不绕过 `contracts/` 自定义另一套 canonical protocol

当前最小路由：

- `GET /healthz`
- `GET /v1/platform/overview`
- `GET /v1/platform/local-smoke`
- `GET /v1/models`
- `GET /v1/models/{id}`
- `POST /v1/chat/completions`
- `POST /v1/responses`
- `POST /v1/messages`
- `GET /v1/session/metadata`
- `GET /v1/session/recovery/checkpoints/{checkpoint_id}`
- `GET /v1/tools/metadata`
- `POST /v1/tools/actions`
- `GET /v1/control-plane/tenants/{tenant_ref}/summary`
- `GET /v1/user-workspace/applications`
- `GET /v1/user-workspace/api-keys`
- `GET /v1/user-workspace/usage/quota-summary`
- `GET /v1/user-workspace/workflow-definitions`
- `GET /v1/user-workspace/runs`
- `GET /v1/control-plane/audit`

其中 `GET /v1/platform/overview` 是 `P3 Local Product Shell / Ops Surface` 的首个只读产品面入口：它汇总服务状态、可选 model/profile、session/tooling metadata route、blocked action route 和当前停止线，供 `apps/radishmind-console/` 本地控制台或上层 UI 一次读取。它不启用真实 executor、durable store、confirmation 接线、长期记忆、业务写回或 replay。

`GET /v1/platform/local-smoke` 是本地开发 readiness 摘要入口：它聚合 `/healthz`、overview contract、model inventory、session/tooling metadata、blocked action no-side-effects、local console CORS origin 和停止线状态，便于开发者或轻量脚本一次判断默认 `7000/4000` 本地 console 链路是否可读。它不启动或守护进程，不实现 production health dashboard、executor、durable store、confirmation、业务写回或 replay。

`/v1/chat/completions`、`/v1/responses` 和 `/v1/messages` 已接到最小 canonical bridge：`Go` 只负责 northbound 请求翻译、provider 选择和进程调度，真正的 canonical request / response 语义仍由 Python runtime 与 gateway 维持。

`GET /v1/session/recovery/checkpoints/{checkpoint_id}` 当前只是 session recovery checkpoint 的 metadata-only route smoke：它返回固定 fixture 边界、checkpoint refs、tool audit refs、`tool_audit_summary`、replay policy 摘要和 state summary，不读取 durable checkpoint store，不返回 materialized tool result，也不执行跨轮 replay。该 route 会拒绝 materialized result 和 replay 类查询参数，例如 `include_materialized_results=true`、`include_tool_results=true` 或 `auto_replay=true`。

`GET /v1/session/metadata`、`GET /v1/tools/metadata` 与 `POST /v1/tools/actions` 当前构成最小 session/tooling 可用外壳：前两者返回平台可消费的 session 扩展字段、history/state/recovery 边界、tool registry metadata 和 contract-only execution policy；后者对任何工具 action 请求都返回 `tool_action_blocked_response`，明确 `status=blocked`、`execution_enabled=false`、`executed=false`、`result_ref=null`、`durable_memory_written=false`、`writes_business_truth=false`。这些路由只用于上层或 UI 发现能力和展示 blocked action 状态，不启用真实 executor、durable store、confirmation 接线、长期记忆、业务写回或 replay。

`control-plane-read-fake-store-handler-plan-v1` 仍作为 fake-store-backed read handler plan 保留；`control-plane-read-fake-store-handler-implementation-v1` 当前实现七条 fake-store-backed read route：`GET /v1/control-plane/tenants/{tenant_ref}/summary`、`GET /v1/user-workspace/applications`、`GET /v1/user-workspace/api-keys`、`GET /v1/user-workspace/usage/quota-summary`、`GET /v1/user-workspace/workflow-definitions`、`GET /v1/user-workspace/runs` 与 `GET /v1/control-plane/audit`。它们使用 `services/platform/internal/httpapi` 内的 in-memory fake store 和 test-only fake auth context，Go route smoke 覆盖成功、missing identity、tenant binding mismatch、cross-tenant query denied、scope denied、invalid filter、forbidden sensitive projection、forbidden method、forbidden query 和 no-side-effects。长驻 HTTP 服务当前没有 fake-auth header 或真实 auth middleware；未注入身份上下文的外部请求应返回 fail-closed envelope，而不是成功读取 fixture 数据。该实现仍不接数据库 query、OIDC、API key lifecycle、quota enforcement、executor、confirmation、writeback 或 replay，也不声明正式 user workspace / admin control plane UI ready。

`control-plane-read-auth-db-preconditions-v1` 当前只固定未来替换 fake auth / fake store 前必须满足的前置条件：read route 必须等 `future Radish OIDC / auth middleware` 注入 identity、tenant、subject、scope、audit 和 request context，真实 read store 必须先定义 `future control plane read store repository` 的窄接口、tenant predicate、sanitized projection、failure taxonomy 和 smoke transition plan。它不实现 OIDC validation、数据库 schema / migration / query、repository、API key lifecycle、quota enforcement、executor、confirmation、writeback 或 replay。

`control-plane-read-consumer-contract-v1` 当前固定 `contracts/typescript/control-plane-read-api.ts`、`scripts/run-control-plane-read-consumer-smoke.py` 和七条 read route 的只读 consumer view model。它让上层可以按统一 envelope、failure view、cursor、audit ref 和 forbidden output policy 消费 response fixture，但不实现正式 UI、不请求真实后端、不接数据库、OIDC、executor、confirmation、writeback 或 replay。

`control-plane-read-formal-ui-boundary-v1` 当前只固定正式 UI 边界：`Admin Control Plane` 与 `User Workspace` 的页面划分、每个页面消费的 read route、loading / empty / denied / stale / partial failure / forbidden projection 状态和敏感字段停止线。它不修改 `apps/radishmind-console/`，不实现 React UI、不请求真实后端、不接数据库、OIDC、executor、confirmation、writeback 或 replay。

`control-plane-read-formal-ui-implementation-readiness-v1` 当前只固定正式 UI 实现 readiness：未来正式产品 UI 预留落点为 `apps/radishmind-web/`，`apps/radishmind-console/` 继续只是本地 ops surface，页面实现顺序、consumer contract 复用和测试策略均由 fixture/checker 固定。该 readiness 不创建 React 页面、不创建 `apps/radishmind-web/`、不修改当前 ops console、不新增 platform route、不接数据库、OIDC、executor、confirmation、writeback 或 replay。

当前第一版 bridge 仍是窄切片：

- 当前仍以文本消息和单轮问答切片为主，但已支持第一版 bridge 增量流式转发
- 当前只把最后一条文本用户消息映射到 `radish/answer_docs_question`
- 返回内容当前优先取 canonical `summary`，必要时回退首条 `answer`

`GET /v1/models` 目前通过 Python provider registry 输出带 route metadata 的 model inventory，作为 northbound discoverability 的第一版收口；它当前已支持列表和 `/v1/models/{id}` 精确 lookup，并带出 provider-qualified profile inventory。profile 可选择 ID 固定为 `profile:<profile>` 或 `provider:<provider>:profile:<profile>`，并与请求侧 selection 和 `diagnostics.providers.selectable_model_ids` 共用同一套 metadata。

## Provider runtime / health boundary

`Provider Runtime & Health v1` 当前已经固定五个可复验边界：

- `provider-capability-matrix-v1`：`scripts/checks/fixtures/provider-capability-matrix-v1.json` 与 `scripts/check-provider-capability-matrix.py` 从 `services/runtime/provider_registry.py` 校验 provider capability matrix。它只说明 provider 声明、profile model id、northbound route / protocol 和 capability metadata 可检查，不说明 provider health 或 production readiness。
- `provider-health-smoke-v1`：`scripts/checks/fixtures/provider-health-smoke-v1.json` 与 `scripts/check-provider-health-smoke.py` 默认只跑 mock runtime smoke 和 config-level inventory smoke。它不联网、不要求真实 credential、不下载模型；optional live health 仍是手动未来切片，失败只能作为 provider health signal，不能写成 production outage。
- `provider-selection-policy-v1`：`scripts/checks/fixtures/provider-selection-policy-v1.json` 与 `scripts/check-provider-selection-policy.py` 固定 request-side profile / provider / concrete model selection、`/v1/models/{id}` 负向边界、credential missing、unsupported capability、timeout 分类和 no implicit fallback 口径。
- `provider-retry-fallback-policy-v1`：`scripts/checks/fixtures/provider-retry-fallback-policy-v1.json` 与 `scripts/check-provider-retry-fallback-policy.py` 固定 retry/fallback audit metadata。当前 `retry_policy=caller-managed`、`fallback_policy=disabled`，平台失败路径不自动重试、不隐式切换 provider。
- `provider-runtime-docs-refresh`：`scripts/checks/fixtures/provider-runtime-docs-refresh.json` 与 `scripts/check-provider-runtime-docs-refresh.py` 固定说明文档、任务卡和入口文档的阶段口径，避免把 capability、health smoke 或 selection policy 误写成 live health、retry/fallback 或 production ready。

调用方应把 `/v1/models` 中的 `credential_state`、`deployment_mode`、`auth_mode`、`streaming`、`northbound_routes` 和 `northbound_protocols` 视为只读发现信息。请求选中某个 profile 后，平台会在 canonical request 的 `context.northbound` 中带出同源 selection metadata，便于审计和排障。

选择失败或未知输入的边界保持显式：

- 未知 `/v1/models/{id}` 返回 `MODEL_NOT_FOUND`。
- 未知 concrete model 可以作为 `runtime_override` 进入 canonical request，但不会被解释为 inventory match。
- 显式未知 `radishmind.provider_profile` 不会被 active profile 隐式替换。
- 当前不做隐式 provider fallback，也不声明 retry/fallback execution 已实现。

这一层仍不是 production provider health system。正式 secret backend、外部 provider live health、live timeout probe、真实 retry/fallback、测试 / 生产环境 smoke 和 production readiness 仍需要独立任务、运行窗口和证据记录。

当前平台级 `ops smoke` 已由 `scripts/check-platform-ops-smoke.py` 固定为快速门禁。它不启动长期驻留服务、不访问外部 provider，只验证三类可运行边界：

- `go test ./...` 能覆盖平台服务层的 `healthz`、northbound 路由、provider/profile selection、session recovery checkpoint metadata-only read route 和 SSE 兼容行为。
- `go test ./...` 也覆盖最小 session/tooling metadata shell 和 blocked action response，确保它们不暴露 executor、materialized result、durable memory 或业务写回能力。
- `scripts/run-platform-bridge.py providers` 能从 Python registry 输出 `mock`、`openai-compatible`、`huggingface` 与 `ollama` provider 能力。
- `scripts/run-platform-bridge.py inventory` 能在受控环境变量下暴露 openai-compatible fallback chain、HuggingFace profile 和 Ollama local profile，并且只暴露 `has_api_key` / `credential_state`，不泄漏 key 原文。

配置分层门禁由 `scripts/check-platform-config.py` 固定到快速检查中。它通过同一个 `config.LoadFromEnv` 入口验证 `config-summary` 和 `config-check`，只输出脱敏字段：provider、profile、model、base_url 是否存在、`credential_state`、timeout、listen addr、Python bridge 路径与字段来源，不输出 `RADISHMIND_PLATFORM_API_KEY` 或 `base_url` 原文。

部署壳 smoke 由 `scripts/check-platform-deployment-smoke.py` 固定到快速检查中。它不启动长驻服务、不访问外部 provider，只验证本地配置文件加载、环境变量覆盖、无效配置失败和 secret 不泄漏。

结构化诊断 smoke 由 `scripts/check-platform-diagnostics.py` 固定到快速检查中。它通过一次性 `diagnostics` 命令聚合启动配置、必填字段、Python bridge provider registry 和 provider/profile inventory，不启动长驻服务、不访问外部 provider，并只输出 `credential_state`、计数和状态字段，不输出 secret、token 或 provider URL 原文。

## Production config / secret boundary

`Production Ops Hardening v1` 的第一切片是 `config-secret-boundary`。当前由 `scripts/checks/fixtures/production-ops-config-secret-boundary.json` 与 `scripts/check-production-ops-config-secret-boundary.py` 固定为 governance boundary：配置来源、密钥注入、provider/profile 覆盖和不可提交项已经有可检查口径，但 production secret backend 仍未实现，不能据此声明 production ready。

配置来源继续按 `default < config file < env` 分层。`mock` provider、`local-smoke`、本地 wrapper 默认值和 `tmp/radishmind-platform.local.json` 只代表 local / dev readiness；远端 provider 的 `RADISHMIND_PLATFORM_API_KEY`、provider profile API key、token、cookie、authorization header 和真实 provider base URL 不得写入 committed 文档、fixture、日志样例或 console 包。

生产环境仍缺少正式 secret backend、环境隔离、provider health policy、process supervisor 和 console production packaging。直到这些条件分别有任务卡、实现和门禁前，`config-summary`、`config-check`、`diagnostics` 和 provider inventory 只能输出 `credential_state`、`base_url_configured`、`secret_fields`、`field_sources` 这类脱敏字段，不输出原始 credential 或 provider URL。

## Production secret backend contract

`production-secret-backend-contract` 已由 `scripts/checks/fixtures/production-ops-secret-backend-contract.json` 与 `scripts/check-production-ops-secret-backend-contract.py` 固定为 Production Ops Hardening v1 的最小治理切片。该切片只定义未来 external secret backend adapter contract：按 `environment`、`provider`、`provider_profile` 与 `secret_ref` 识别 secret reference，并要求后续运行面只暴露 `credential_state`、`secret_backend_configured`、`secret_ref_present`、`missing_secret_refs` 和 `field_sources` 等脱敏字段。

当前明确不实现真实云 secret 服务、不写入真实 secret、不调用云 API、不声明 production ready。`RADISHMIND_PLATFORM_API_KEY` 仍只允许作为 developer env override；`RADISHMIND_SECRET_SOURCE` 只能表示部署态外部 secret 来源要求，不是 secret backend 本身。真实 production secret backend、secret rotation policy、production secret audit store、provider health policy、environment isolation 和 process supervisor 仍为 `not_satisfied`。

## Production secret backend implementation readiness

`production-secret-backend-implementation-readiness` 已由 `docs/task-cards/production-secret-backend-implementation-v1-plan.md`、`scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json` 与 `scripts/check-production-ops-secret-backend-implementation-readiness.py` 固定为下一步实现前置条件清单。它要求先定义 secret ref schema、config 注入点、provider profile binding、脱敏审计字段、failure taxonomy、fake resolver 测试策略、operator runbook 和 rotation / audit policy。

该 readiness 不实现 resolver、不接云、不读取或写入真实 secret、不要求真实 credential，也不改变当前运行时默认状态。后续只有在单独实现切片中完成 schema、disabled resolver、负向门禁和运行手册后，才能继续讨论真实 production secret backend。

## Production secret reference schema

`secret-ref-schema-and-fixtures` 已由 `contracts/production-secret-reference.schema.json`、`scripts/checks/fixtures/production-secret-reference-basic.json` 与 `scripts/check-production-secret-reference-contract.py` 固定为 committed secret reference contract。该 schema 只允许保存 `environment`、`provider`、`provider_profile`、`secret_ref`、`required_fields` 和 `sanitized_fields`，并要求 fixture 明确 `stores_secret_values=false`、`resolver_enabled=false`、`cloud_calls_allowed=false` 和 `production_secret_backend_ready=false`。

该 schema / fixture 只证明 secret reference 格式可检查，不实现 resolver，不接云，不包含 secret value、provider raw URL、API key、token、cookie、authorization header 或 credential 原文。下一步如继续推进，应进入 `config-secret-ref-readiness`，让 config summary / diagnostics 能报告 `secret_ref_present` 和 `missing_secret_refs` 等脱敏状态。

## Startup / supervisor boundary

`startup-supervisor-boundary` 已由 `scripts/checks/fixtures/production-ops-startup-supervisor-boundary.json` 与 `scripts/check-production-ops-startup-supervisor-boundary.py` 固定为 governance boundary。当前支持的启动入口只有两类：`scripts/run-platform-service.ps1` / `scripts/run-platform-service.sh` 的人工 platform wrapper，以及 `scripts/run-radishmind-console-dev.ps1` / `scripts/run-radishmind-console-dev.sh` 的本地 console dev launcher。

platform wrapper 只支持 `serve`、`config-summary`、`config-check` 和 `diagnostics`，未知命令必须失败并返回退出码 `2`。console dev launcher 只负责启动或复用本地 backend / frontend，探测 `/healthz`、`/v1/platform/overview`、`/v1/platform/local-smoke`、本地 CORS preflight 和前端页面；`-ExitAfterProbe` / `--exit-after-probe` 只表示探测后清理本次创建的开发进程，不是 lifecycle management。

当前仍没有 production process supervisor、automatic restart、production service manager 或 production log retention。`tmp/radishmind-console-dev` 下的日志只用于本地排障，不能解释为生产日志路径；port reuse 只表示开发期复用本机已有进程，不能解释为服务发现。`local-smoke` 仍不是 production health。

## Environment isolation boundary

`environment-isolation` 已由 `scripts/checks/fixtures/production-ops-environment-isolation-boundary.json` 与 `scripts/check-production-ops-environment-isolation-boundary.py` 固定为 governance boundary。当前只区分三类 readiness：`local_readiness`、`dev_smoke` 和仍为 `not_satisfied` 的 `production_readiness`。

`/v1/platform/local-smoke`、`mock` provider、本地 console CORS、developer env override 和 Vite preview 都只属于 local / dev 范围。`local-smoke` 只能说明本地只读 console 所需 route、contract、CORS 和 stop-lines 可读；它不是 production health。`mock` provider 和 demo profile 只能证明协议、UI 和本地排障链路可用，不能解释为 production provider / profile。`http://127.0.0.1:4000` 与 `http://localhost:4000` 的 CORS 允许列表只服务本地 console 开发，不是 production CORS policy。

production readiness 仍必须等待 deployment environment isolation、production auth policy、production CORS policy、provider health policy、secret backend、process supervisor 和 console production packaging。当前任何 local / dev smoke 通过都不得升级为 production ready。

## Console package smoke boundary

`console-production-package-smoke` 已由 `scripts/checks/fixtures/production-ops-console-package-smoke.json` 与 `scripts/check-production-ops-console-package-smoke.py` 固定为 governance boundary。当前 console package 只允许 `dev`、`build` 和 `preview` 三个 npm script：`npm run build` 用于本地或 CI smoke，执行 `tsc --noEmit && vite build`；`npm run preview` 只允许绑定 `127.0.0.1:4000` 做本地 build preview。

这些入口不等于 production package 或 production hosting。`apps/radishmind-console/package.json` 必须保持 `private=true`，不得添加 deploy / publish / release / package / docker 类脚本，不得提交 `apps/radishmind-console/dist/` 或 `apps/radishmind-console/node_modules/`。console production packaging 仍为 `not_satisfied`，直到另有生产发布目标、静态资产策略、正式鉴权 / CORS policy 和部署门禁。

## Docker local compose boundary

`docker-local-compose` 已由 `scripts/checks/fixtures/production-ops-docker-local-compose.json` 与 `scripts/check-production-ops-docker-local-compose.py` 固定为本地容器 smoke 资产。当前只新增 `services/platform/Dockerfile`、`apps/radishmind-console/Dockerfile`、`apps/radishmind-console/nginx.local.conf` 和 `deploy/docker-compose.local.yaml`，用于本机 build 并验证 platform + console 的容器组合。

本地 compose 默认使用 `mock` provider，发布 `7000/4000` 到宿主机，console 构建期 `VITE_RADISHMIND_PLATFORM_BASE_URL` 默认指向 `http://127.0.0.1:7000`。它不包含 secret、不使用 `RADISHMIND_IMAGE_TRACK` / `RADISHMIND_IMAGE_TAG`，不定义测试 / 生产部署态，不实现 production secret backend、process supervisor、正式 auth / CORS policy、镜像发布或 production ready。

本地长驻服务入口由 `scripts/run-platform-service.sh` 与 `scripts/run-platform-service.ps1` 收口。wrapper 会固定仓库根、`services/platform` 工作目录、默认 `GOCACHE=/tmp/radishmind-go-build-cache`，在未显式设置 `RADISHMIND_PLATFORM_PYTHON_BIN` 时优先使用仓库 `.venv` Python，再回退系统 `python3` / `python`，并在 `tmp/radishmind-platform.local.json` 存在时自动作为默认 `RADISHMIND_PLATFORM_CONFIG`。

`/v1/models` 的 profile metadata 现在必须带出稳定 discoverability 字段：`capabilities`、`northbound_protocols`、`northbound_routes`、`credential_state`、`deployment_mode`、`auth_mode` 与 `streaming`。调用方应基于这些字段判断某个 profile 能否用于 chat、是否支持流式、凭据是否已配置，以及它属于 remote API 还是 local daemon。

profile 可选择 ID 固定为 `profile:<profile>` 或 `provider:<provider>:profile:<profile>`。`/v1/models`、请求时的 provider/profile selection 和 `diagnostics.providers.selectable_model_ids` 使用同一套 ID 与 readiness metadata；当请求选中某个 profile 时，canonical request 的 `context.northbound` 会记录 `credential_state`、`deployment_mode`、`auth_mode`、`streaming`、`northbound_routes` 与 `northbound_protocols`，用于审计和排障。

请求级观测已固定为轻量平台能力：`/v1/chat/completions`、`/v1/responses` 与 `/v1/messages` 会接受或生成 `request_id`，通过 `X-Request-Id` 响应头返回，并把同一 ID 写入 canonical `CopilotRequest.request_id` 与 `context.northbound.request_id`。非流式和流式路径都会记录统一日志字段：`request_id`、route、HTTP status、latency、provider/profile、selected model、selection source，以及失败时的 error code 与 failure boundary。

错误响应采用统一 error taxonomy，不输出 provider URL 或 credential 原文。当前固定的主要失败边界包括：`northbound_request`、`canonical_request`、`provider_inventory`、`python_bridge`、`platform_response`、`southbound_provider` 与 `configuration`。

## 本地启动 runbook

该 runbook 面向开发者本机验证，不等同于 production deployment。启动服务前应先运行平台层单元测试：

```bash
cd services/platform
GOCACHE=/tmp/radishmind-go-build-cache go test ./...
```

从仓库根目录启动本地平台服务：

```bash
RADISHMIND_PLATFORM_LISTEN_ADDR=127.0.0.1:7000 \
RADISHMIND_PLATFORM_PROVIDER=mock \
RADISHMIND_PLATFORM_MODEL=radishmind-local-dev \
go run ./services/platform/cmd/radishmind-platform
```

如果已经进入 `services/platform` 目录，也可以使用：

```bash
RADISHMIND_PLATFORM_LISTEN_ADDR=127.0.0.1:7000 \
RADISHMIND_PLATFORM_PROVIDER=mock \
RADISHMIND_PLATFORM_MODEL=radishmind-local-dev \
go run ./cmd/radishmind-platform
```

启动路径仍由 `config.LoadFromEnv -> httpapi.NewServer -> ListenAndServe` 组成。`Go` 层只负责本地 HTTP 壳、northbound 请求翻译和 Python bridge 调度，不直接承载模型推理逻辑。

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `RADISHMIND_PLATFORM_CONFIG` | 空 | 可选 JSON 配置文件路径 |
| `RADISHMIND_PLATFORM_LISTEN_ADDR` | `:7000` | 本地 HTTP 监听地址 |
| `RADISHMIND_PLATFORM_READ_HEADER_TIMEOUT` | `5s` | HTTP header 读取超时 |
| `RADISHMIND_PLATFORM_WRITE_TIMEOUT` | `30s` | HTTP 写响应超时 |
| `RADISHMIND_PLATFORM_BRIDGE_TIMEOUT` | `30s` | Go 调 Python bridge 的超时 |
| `RADISHMIND_PLATFORM_PYTHON_BIN` | `python3` | Python bridge 解释器 |
| `RADISHMIND_PLATFORM_BRIDGE_SCRIPT` | `scripts/run-platform-bridge.py` | Python bridge 脚本路径 |
| `RADISHMIND_PLATFORM_PROVIDER` | `mock` | 默认 southbound provider |
| `RADISHMIND_PLATFORM_PROVIDER_PROFILE` | 空 | 默认 provider profile |
| `RADISHMIND_PLATFORM_MODEL` | 空 | northbound 默认 model id |
| `RADISHMIND_PLATFORM_BASE_URL` | 空 | 显式 provider base URL 覆盖 |
| `RADISHMIND_PLATFORM_API_KEY` | 空 | 显式 provider API key 覆盖；不得写入文档或提交 |
| `RADISHMIND_PLATFORM_TEMPERATURE` | `0` | provider 调用温度 |

配置优先级固定为 `default < config file < env`。配置文件当前使用 JSON，字段名与脱敏 summary 保持一致，例如：

```json
{
  "listen_addr": "127.0.0.1:7000",
  "provider": "mock",
  "model": "radishmind-local-dev",
  "bridge_timeout": "30s"
}
```

可用配置文件启动本地平台服务：

```bash
RADISHMIND_PLATFORM_CONFIG=tmp/radishmind-platform.local.json \
go run ./services/platform/cmd/radishmind-platform
```

推荐通过稳定 wrapper 先跑配置检查和结构化诊断，再启动长驻服务：

```bash
./scripts/run-platform-service.sh config-check
./scripts/run-platform-service.sh diagnostics
./scripts/run-platform-service.sh serve
```

Windows / PowerShell 使用：

```powershell
pwsh ./scripts/run-platform-service.ps1 -Command config-check
pwsh ./scripts/run-platform-service.ps1 -Command diagnostics
pwsh ./scripts/run-platform-service.ps1 -Command serve
```

上层消费 smoke 可以先用离线 fixture 生成展示视图，不要求启动服务：

```bash
python scripts/run-platform-overview-consumer-smoke.py --check
python scripts/run-platform-local-smoke.py --check
python scripts/run-platform-session-tooling-consumer-smoke.py --check
```

服务启动后，也可以指向本地平台 API 生成同一份消费视图：

```bash
python scripts/run-platform-overview-consumer-smoke.py \
  --base-url http://127.0.0.1:7000 \
  --check

python scripts/run-platform-local-smoke.py \
  --base-url http://127.0.0.1:7000 \
  --check

python scripts/run-platform-session-tooling-consumer-smoke.py \
  --base-url http://127.0.0.1:7000 \
  --check
```

overview consumer smoke 只读取 `GET /v1/platform/overview`，把 service status、model inventory、session/tooling surface 和 stop-lines 投影成本地 console view model；local smoke consumer 只读取 `GET /v1/platform/local-smoke`，把本地 readiness 摘要投影为 healthz、overview、model inventory、session/tooling、CORS 和停止线检查；session/tooling consumer smoke 只读取 `session metadata`、`tools metadata` 并提交一次会被阻断的 tool action 请求，用于验证上层可展示 `blocked`、`requires_confirmation` 与 `no_side_effects`。这些 smoke 都不会启用真实 executor、durable store、confirmation、replay 或业务写回。

最小本地 console 壳位于 `apps/radishmind-console/`。它复用 `contracts/typescript/platform-overview-api.ts` 与 `contracts/typescript/platform-local-smoke-api.ts`，默认读取 `http://127.0.0.1:7000/v1/platform/overview` 和 `http://127.0.0.1:7000/v1/platform/local-smoke`：

```bash
cd apps/radishmind-console
npm install
npm run dev
```

该 console 只展示 service status、model/profile inventory、Provider/Profile Details、session/tooling blocked 状态、Blocked Action Detail、Local Readiness、stop-lines、Stop-line Details、audit boundary、Dev Diagnostics、refresh 状态和连接失败诊断，不调用 `/v1/tools/actions`，也不实现 executor、durable store、confirmation、业务写回或 replay。refresh 期间和连接失败后可以保留上一份只读 overview / local-smoke readiness，用于排障，不代表平台会自动恢复执行。

当前页面结构是浅色左侧导航栏、主工作区和右侧 readiness / stop-line 辅助栏；窄屏下改为单列信息顺序。该结构来自 `docs/designs/radishmind-console-ops-surface-v0.pen` 和 [UI 设计规范](../../docs/radishmind-ui-design-spec.md)，但仍只表示本地 ops surface，不表示 production console 或 production packaging 已完成。

平台服务当前只为 `http://127.0.0.1:4000` 与 `http://localhost:4000` 返回本地 console CORS header，并处理 `OPTIONS` preflight；该能力只服务本地 console 开发，不等同于 production CORS policy、正式鉴权或外部公开部署。

console production packaging 仍未完成：`apps/radishmind-console/package.json` 必须保持 `private=true`，不添加 deploy / publish / release 脚本，不提交 `dist/` 或 `node_modules/`。P3 short-close checklist 继续把 production secret backend、process supervisor、部署环境隔离和 console production packaging 标为 `not_satisfied`；当前只固定本地开发入口和最小 deployment smoke。

如果要同时启动并验证 platform 后端和本地 console，可从仓库根目录使用：

```powershell
pwsh ./scripts/run-radishmind-console-dev.ps1
```

Linux / WSL 使用：

```bash
./scripts/run-radishmind-console-dev.sh
```

该入口复用 `scripts/run-platform-service.ps1` / `scripts/run-platform-service.sh` 和 `apps/radishmind-console/` 的 `npm run dev`，启动或复用 `http://127.0.0.1:7000` 与 `http://127.0.0.1:4000`，并探测 `http://127.0.0.1:7000/healthz`、`http://127.0.0.1:7000/v1/platform/overview`、`http://127.0.0.1:7000/v1/platform/local-smoke`、本地 console CORS preflight 和 `http://127.0.0.1:4000`。端口冲突时先释放 `7000/4000` 或确认现有服务就是 RadishMind；CORS 失败时确认 console origin 是允许的本地 origin；浏览器 `unsafe port` / `ERR_UNSAFE_PORT` 通常表示端口被浏览器直接拦截，优先回到默认 `4000/7000`。该入口不是 production supervisor，不实现真实 executor、durable store、confirmation、业务写回或 replay。

验证脚本本身时可加 `-ExitAfterProbe` 或 `--exit-after-probe`，让它启动、探测成功后自动停止本次创建的本地进程。

可用一次性命令检查本地配置摘要，输出不会暴露 secret：

```bash
go run ./services/platform/cmd/radishmind-platform config-summary
go run ./services/platform/cmd/radishmind-platform config-check
go run ./services/platform/cmd/radishmind-platform diagnostics
```

`diagnostics` 输出固定为结构化 JSON，主要字段包括：

- `status`：`ok` 或 `error`
- `config`：复用脱敏 `config-summary`
- `checks`：`config_required_fields`、`bridge_provider_registry`、`bridge_provider_inventory` 与 `deployment_readiness`
- `bridge`：Python bridge 脚本、registry / inventory 可用性和 bridge 失败码
- `providers`：provider registry 数量、profile 数量、active profile chain、credential state 计数和 deployment mode 计数
- `providers.selectable_model_ids`：与 `/v1/models` 和请求选择一致的 profile model id 列表
- `failure_codes` / `failure`：启动前可定位的失败边界

## 本地 smoke 验证

服务启动后，在另一个终端执行：

```bash
curl -sS http://127.0.0.1:7000/healthz
curl -sS http://127.0.0.1:7000/v1/platform/overview
curl -sS http://127.0.0.1:7000/v1/platform/local-smoke
curl -sS http://127.0.0.1:7000/v1/models
curl -sS http://127.0.0.1:7000/v1/models/mock
curl -sS http://127.0.0.1:7000/v1/session/metadata
curl -sS 'http://127.0.0.1:7000/v1/session/recovery/checkpoints/session-checkpoint-0001?session_id=radishflow-session-001&turn_id=turn-0003'
curl -sS http://127.0.0.1:7000/v1/tools/metadata
curl -sS http://127.0.0.1:7000/v1/tools/actions \
  -H 'Content-Type: application/json' \
  -d '{"tool_id":"radishflow.suggest_edits.candidate_builder.v1","action":"execute","session_id":"radishflow-session-001","turn_id":"turn-0003"}'
curl -sS http://127.0.0.1:7000/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"radishmind-local-dev","messages":[{"role":"user","content":"请简要说明当前 RadishMind 平台状态。"}]}'
```

预期边界：

- `/healthz` 返回 `status=ok`、`service=radishmind-platform`。
- `/v1/platform/overview` 返回 `platform_overview`，其中 `product_surface.mode=local_read_only_product_shell`，并汇总 `/v1/models`、session metadata、tool metadata 和 blocked action route；所有 executor / durable store / confirmation / writeback / replay 停止线均为 `false`。
- `/v1/platform/local-smoke` 返回 `platform_local_smoke`，其中 `summary.local_console_ready=true` 表示本地只读 console 所需的 healthz、overview、model inventory、session/tooling metadata、blocked action no-side-effects、local CORS origin 和停止线均可读；该 route 只做摘要，不启动服务、不守护进程、不表示生产部署 ready。
- `/v1/models` 返回 OpenAI-compatible `object=list`，并包含 provider registry 与 profile inventory。
- `/v1/models/mock` 可通过精确 lookup 返回 mock provider model。
- `/v1/session/metadata` 返回 `session_metadata`，其中 durable session/checkpoint store、long-term memory、automatic replay 和 business truth write 均为 `false`。
- `/v1/session/recovery/checkpoints/session-checkpoint-0001` 返回 `session_recovery_checkpoint_read_result`，且 `access_policy.metadata_only=true`、`materialized_results_included=false`、`auto_replay_enabled=false`，`result.tool_audit_summary.execution_enabled=false`。
- `/v1/tools/metadata` 返回 `tooling_metadata`，其中 `registry_policy.execution_enabled=false`，每个工具的 execution mode 为 `contract_only`。
- 七条 control-plane read route 目前只在 Go test 中通过 test-only fake auth context 验证；直接 curl 未带未来 auth context 时应 fail closed，而不是匿名返回跨租户数据。
- `/v1/tools/actions` 返回 `tool_action_blocked_response`，且不会运行工具、返回 materialized result、写 durable memory 或写业务真相源。
- `/v1/chat/completions` 在 `mock` provider 下返回 advisory 文本，不访问外部 provider，不写回任何上层项目。

## 故障边界

- 启动前先运行 `./scripts/run-platform-service.sh diagnostics`；如果 `status=error`，优先读取 `failure.code` 和 `checks[].code`，再决定是否启动长驻服务。
- 若 `failure.code=CONFIG_REQUIRED_FIELDS_MISSING`，优先检查 `config.missing_required_fields`，不要从日志或诊断输出寻找 secret 原文。
- 若 `failure.code=PROVIDER_REGISTRY_UNAVAILABLE` 或 `PROVIDER_INVENTORY_UNAVAILABLE`，优先检查 `bridge.python_binary`、`bridge.script`、当前工作目录和 Python import 路径。
- 若启动时报 `load config`，优先检查 duration / float 类环境变量格式，例如 `RADISHMIND_PLATFORM_BRIDGE_TIMEOUT=30s`、`RADISHMIND_PLATFORM_TEMPERATURE=0`。
- 若 `/v1/models` 返回 `PROVIDER_INVENTORY_UNAVAILABLE`，优先检查 `RADISHMIND_PLATFORM_PYTHON_BIN`、`RADISHMIND_PLATFORM_BRIDGE_SCRIPT` 和当前工作目录是否能访问仓库根下的 Python bridge。
- 若 northbound 路由返回 `PLATFORM_BRIDGE_FAILED`，先用 `python3 scripts/run-platform-bridge.py providers` 与 `python3 scripts/run-platform-bridge.py inventory` 验证 Python 侧 provider registry / inventory 是否可用。
- 若返回 `MODEL_NOT_FOUND`，优先用 `/v1/models` 或 `diagnostics.providers.selectable_model_ids` 确认可选择 ID，避免把 provider id、profile id 与真实 upstream model 混用。
- 若返回 `PLATFORM_RESPONSE_INVALID`，说明 Python bridge 已返回 envelope，但 `Go` northbound 兼容层无法翻译为目标协议响应，应优先检查 envelope 的 `response` 结构。
- 若要接真实 provider，必须通过环境变量或本机 secret 注入 `RADISHMIND_PLATFORM_BASE_URL` / `RADISHMIND_PLATFORM_API_KEY` 或 provider profile 配置；不要把 key、token、cookie 或真实 provider raw dump 写入 committed 文档。
