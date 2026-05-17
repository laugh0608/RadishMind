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
- `GET /v1/models`
- `GET /v1/models/{id}`
- `POST /v1/chat/completions`
- `POST /v1/responses`
- `POST /v1/messages`
- `GET /v1/session/metadata`
- `GET /v1/session/recovery/checkpoints/{checkpoint_id}`
- `GET /v1/tools/metadata`
- `POST /v1/tools/actions`

其中 `GET /v1/platform/overview` 是 `P3 Local Product Shell` 的首个只读产品面入口：它汇总服务状态、可选 model/profile、session/tooling metadata route、blocked action route 和当前停止线，供未来本地控制台或上层 UI 一次读取。它不启用真实 executor、durable store、confirmation 接线、长期记忆、业务写回或 replay。

`/v1/chat/completions`、`/v1/responses` 和 `/v1/messages` 已接到最小 canonical bridge：`Go` 只负责 northbound 请求翻译、provider 选择和进程调度，真正的 canonical request / response 语义仍由 Python runtime 与 gateway 维持。

`GET /v1/session/recovery/checkpoints/{checkpoint_id}` 当前只是 session recovery checkpoint 的 metadata-only route smoke：它返回固定 fixture 边界、checkpoint refs、tool audit refs、`tool_audit_summary`、replay policy 摘要和 state summary，不读取 durable checkpoint store，不返回 materialized tool result，也不执行跨轮 replay。该 route 会拒绝 materialized result 和 replay 类查询参数，例如 `include_materialized_results=true`、`include_tool_results=true` 或 `auto_replay=true`。

`GET /v1/session/metadata`、`GET /v1/tools/metadata` 与 `POST /v1/tools/actions` 当前构成最小 session/tooling 可用外壳：前两者返回平台可消费的 session 扩展字段、history/state/recovery 边界、tool registry metadata 和 contract-only execution policy；后者对任何工具 action 请求都返回 `tool_action_blocked_response`，明确 `status=blocked`、`execution_enabled=false`、`executed=false`、`result_ref=null`、`durable_memory_written=false`、`writes_business_truth=false`。这些路由只用于上层或 UI 发现能力和展示 blocked action 状态，不启用真实 executor、durable store、confirmation 接线、长期记忆、业务写回或 replay。

当前第一版 bridge 仍是窄切片：

- 当前仍以文本消息和单轮问答切片为主，但已支持第一版 bridge 增量流式转发
- 当前只把最后一条文本用户消息映射到 `radish/answer_docs_question`
- 返回内容当前优先取 canonical `summary`，必要时回退首条 `answer`

`GET /v1/models` 目前通过 Python provider registry 输出带 route metadata 的 model inventory，作为 northbound discoverability 的第一版收口；它当前已支持列表和 `/v1/models/{id}` 精确 lookup，并带出第一版 provider-qualified profile inventory，但还不是完整的动态 provider/profile discovery。下一步优先补更广 provider/profile discoverability、配置分层、部署观测和 failure boundary。

当前平台级 `ops smoke` 已由 `scripts/check-platform-ops-smoke.py` 固定为快速门禁。它不启动长期驻留服务、不访问外部 provider，只验证三类可运行边界：

- `go test ./...` 能覆盖平台服务层的 `healthz`、northbound 路由、provider/profile selection、session recovery checkpoint metadata-only read route 和 SSE 兼容行为。
- `go test ./...` 也覆盖最小 session/tooling metadata shell 和 blocked action response，确保它们不暴露 executor、materialized result、durable memory 或业务写回能力。
- `scripts/run-platform-bridge.py providers` 能从 Python registry 输出 `mock`、`openai-compatible`、`huggingface` 与 `ollama` provider 能力。
- `scripts/run-platform-bridge.py inventory` 能在受控环境变量下暴露 openai-compatible fallback chain、HuggingFace profile 和 Ollama local profile，并且只暴露 `has_api_key` / `credential_state`，不泄漏 key 原文。

配置分层门禁由 `scripts/check-platform-config.py` 固定到快速检查中。它通过同一个 `config.LoadFromEnv` 入口验证 `config-summary` 和 `config-check`，只输出脱敏字段：provider、profile、model、base_url 是否存在、`credential_state`、timeout、listen addr、Python bridge 路径与字段来源，不输出 `RADISHMIND_PLATFORM_API_KEY` 或 `base_url` 原文。

部署壳 smoke 由 `scripts/check-platform-deployment-smoke.py` 固定到快速检查中。它不启动长驻服务、不访问外部 provider，只验证本地配置文件加载、环境变量覆盖、无效配置失败和 secret 不泄漏。

结构化诊断 smoke 由 `scripts/check-platform-diagnostics.py` 固定到快速检查中。它通过一次性 `diagnostics` 命令聚合启动配置、必填字段、Python bridge provider registry 和 provider/profile inventory，不启动长驻服务、不访问外部 provider，并只输出 `credential_state`、计数和状态字段，不输出 secret、token 或 provider URL 原文。

本地长驻服务入口由 `scripts/run-platform-service.sh` 与 `scripts/run-platform-service.ps1` 收口。wrapper 会固定仓库根、`services/platform` 工作目录、默认 `GOCACHE=/tmp/radishmind-go-build-cache`，并在 `tmp/radishmind-platform.local.json` 存在时自动作为默认 `RADISHMIND_PLATFORM_CONFIG`。

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
RADISHMIND_PLATFORM_LISTEN_ADDR=127.0.0.1:8080 \
RADISHMIND_PLATFORM_PROVIDER=mock \
RADISHMIND_PLATFORM_MODEL=radishmind-local-dev \
go run ./services/platform/cmd/radishmind-platform
```

如果已经进入 `services/platform` 目录，也可以使用：

```bash
RADISHMIND_PLATFORM_LISTEN_ADDR=127.0.0.1:8080 \
RADISHMIND_PLATFORM_PROVIDER=mock \
RADISHMIND_PLATFORM_MODEL=radishmind-local-dev \
go run ./cmd/radishmind-platform
```

启动路径仍由 `config.LoadFromEnv -> httpapi.NewServer -> ListenAndServe` 组成。`Go` 层只负责本地 HTTP 壳、northbound 请求翻译和 Python bridge 调度，不直接承载模型推理逻辑。

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `RADISHMIND_PLATFORM_CONFIG` | 空 | 可选 JSON 配置文件路径 |
| `RADISHMIND_PLATFORM_LISTEN_ADDR` | `:8080` | 本地 HTTP 监听地址 |
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
  "listen_addr": "127.0.0.1:8080",
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
python scripts/run-platform-session-tooling-consumer-smoke.py --check
```

服务启动后，也可以指向本地平台 API 生成同一份消费视图：

```bash
python scripts/run-platform-overview-consumer-smoke.py \
  --base-url http://127.0.0.1:8080 \
  --check

python scripts/run-platform-session-tooling-consumer-smoke.py \
  --base-url http://127.0.0.1:8080 \
  --check
```

overview consumer smoke 只读取 `GET /v1/platform/overview`，把 service status、model inventory、session/tooling surface 和 stop-lines 投影成本地 console view model；session/tooling consumer smoke 只读取 `session metadata`、`tools metadata` 并提交一次会被阻断的 tool action 请求，用于验证上层可展示 `blocked`、`requires_confirmation` 与 `no_side_effects`。二者都不会启用真实 executor、durable store、confirmation、replay 或业务写回。

生产前仍需要单独补 secret 管理、部署环境隔离和观测策略；当前只固定本地开发入口和最小 deployment smoke。

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
curl -sS http://127.0.0.1:8080/healthz
curl -sS http://127.0.0.1:8080/v1/platform/overview
curl -sS http://127.0.0.1:8080/v1/models
curl -sS http://127.0.0.1:8080/v1/models/mock
curl -sS http://127.0.0.1:8080/v1/session/metadata
curl -sS 'http://127.0.0.1:8080/v1/session/recovery/checkpoints/session-checkpoint-0001?session_id=radishflow-session-001&turn_id=turn-0003'
curl -sS http://127.0.0.1:8080/v1/tools/metadata
curl -sS http://127.0.0.1:8080/v1/tools/actions \
  -H 'Content-Type: application/json' \
  -d '{"tool_id":"radishflow.suggest_edits.candidate_builder.v1","action":"execute","session_id":"radishflow-session-001","turn_id":"turn-0003"}'
curl -sS http://127.0.0.1:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"radishmind-local-dev","messages":[{"role":"user","content":"请简要说明当前 RadishMind 平台状态。"}]}'
```

预期边界：

- `/healthz` 返回 `status=ok`、`service=radishmind-platform`。
- `/v1/platform/overview` 返回 `platform_overview`，其中 `product_surface.mode=local_read_only_product_shell`，并汇总 `/v1/models`、session metadata、tool metadata 和 blocked action route；所有 executor / durable store / confirmation / writeback / replay 停止线均为 `false`。
- `/v1/models` 返回 OpenAI-compatible `object=list`，并包含 provider registry 与 profile inventory。
- `/v1/models/mock` 可通过精确 lookup 返回 mock provider model。
- `/v1/session/metadata` 返回 `session_metadata`，其中 durable session/checkpoint store、long-term memory、automatic replay 和 business truth write 均为 `false`。
- `/v1/session/recovery/checkpoints/session-checkpoint-0001` 返回 `session_recovery_checkpoint_read_result`，且 `access_policy.metadata_only=true`、`materialized_results_included=false`、`auto_replay_enabled=false`，`result.tool_audit_summary.execution_enabled=false`。
- `/v1/tools/metadata` 返回 `tooling_metadata`，其中 `registry_policy.execution_enabled=false`，每个工具的 execution mode 为 `contract_only`。
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
