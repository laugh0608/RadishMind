# RadishMind Platform Service Layer

本目录承载 `Go` 平台服务层的最小骨架。

当前职责：

- 启动最小本地 `HTTP` 服务
- 承载 northbound `API` / `gateway` 入口
- 通过 Python bridge 调用 canonical `CopilotGatewayEnvelope`
- 提供观测、部署壳和后续鉴权 / 流式转发落点

当前明确不做：

- 不在这里复制第二套业务真相源
- 不在这里重写模型推理、训练、评测或 `builder`
- 不绕过 `contracts/` 自定义另一套 canonical protocol

当前最小路由：

- `GET /healthz`
- `GET /v1/models`
- `GET /v1/models/{id}`
- `POST /v1/chat/completions`
- `POST /v1/responses`
- `POST /v1/messages`

其中 `/v1/chat/completions`、`/v1/responses` 和 `/v1/messages` 已接到最小 canonical bridge：`Go` 只负责 northbound 请求翻译、provider 选择和进程调度，真正的 canonical request / response 语义仍由 Python runtime 与 gateway 维持。

当前第一版 bridge 仍是窄切片：

- 当前仍以文本消息和单轮问答切片为主，但已支持第一版 bridge 增量流式转发
- 当前只把最后一条文本用户消息映射到 `radish/answer_docs_question`
- 返回内容当前优先取 canonical `summary`，必要时回退首条 `answer`

`GET /v1/models` 目前通过 Python provider registry 输出带 route metadata 的 model inventory，作为 northbound discoverability 的第一版收口；它当前已支持列表和 `/v1/models/{id}` 精确 lookup，并带出第一版 provider-qualified profile inventory，但还不是完整的动态 provider/profile discovery。下一步优先补更广 provider/profile discoverability、配置分层和部署壳。

当前平台级 `ops smoke` 已由 `scripts/check-platform-ops-smoke.py` 固定为快速门禁。它不启动长期驻留服务、不访问外部 provider，只验证三类可运行边界：

- `go test ./...` 能覆盖平台服务层的 `healthz`、northbound 路由、provider/profile selection 和 SSE 兼容行为。
- `scripts/run-platform-bridge.py providers` 能从 Python registry 输出 `mock`、`openai-compatible`、`huggingface` 与 `ollama` provider 能力。
- `scripts/run-platform-bridge.py inventory` 能在受控环境变量下暴露 openai-compatible fallback chain、HuggingFace profile 和 Ollama local profile，并且只暴露 `has_api_key` / `credential_state`，不泄漏 key 原文。

配置分层门禁由 `scripts/check-platform-config.py` 固定到快速检查中。它通过同一个 `config.LoadFromEnv` 入口验证 `config-summary` 和 `config-check`，只输出脱敏字段：provider、profile、model、base_url 是否存在、`credential_state`、timeout、listen addr、Python bridge 路径与字段来源，不输出 `RADISHMIND_PLATFORM_API_KEY` 或 `base_url` 原文。

部署壳 smoke 由 `scripts/check-platform-deployment-smoke.py` 固定到快速检查中。它不启动长驻服务、不访问外部 provider，只验证本地配置文件加载、环境变量覆盖、无效配置失败和 secret 不泄漏。

本地长驻服务入口由 `scripts/run-platform-service.sh` 与 `scripts/run-platform-service.ps1` 收口。wrapper 会固定仓库根、`services/platform` 工作目录、默认 `GOCACHE=/tmp/radishmind-go-build-cache`，并在 `tmp/radishmind-platform.local.json` 存在时自动作为默认 `RADISHMIND_PLATFORM_CONFIG`。

`/v1/models` 的 profile metadata 现在必须带出稳定 discoverability 字段：`capabilities`、`northbound_protocols`、`northbound_routes`、`credential_state`、`deployment_mode`、`auth_mode` 与 `streaming`。调用方应基于这些字段判断某个 profile 能否用于 chat、是否支持流式、凭据是否已配置，以及它属于 remote API 还是 local daemon。

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

推荐通过稳定 wrapper 先跑配置检查，再启动长驻服务：

```bash
./scripts/run-platform-service.sh config-check
./scripts/run-platform-service.sh serve
```

Windows / PowerShell 使用：

```powershell
pwsh ./scripts/run-platform-service.ps1 -Command config-check
pwsh ./scripts/run-platform-service.ps1 -Command serve
```

生产前仍需要单独补 secret 管理、部署环境隔离和观测策略；当前只固定本地开发入口和最小 deployment smoke。

可用一次性命令检查本地配置摘要，输出不会暴露 secret：

```bash
go run ./services/platform/cmd/radishmind-platform config-summary
go run ./services/platform/cmd/radishmind-platform config-check
```

## 本地 smoke 验证

服务启动后，在另一个终端执行：

```bash
curl -sS http://127.0.0.1:8080/healthz
curl -sS http://127.0.0.1:8080/v1/models
curl -sS http://127.0.0.1:8080/v1/models/mock
curl -sS http://127.0.0.1:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"radishmind-local-dev","messages":[{"role":"user","content":"请简要说明当前 RadishMind 平台状态。"}]}'
```

预期边界：

- `/healthz` 返回 `status=ok`、`service=radishmind-platform`。
- `/v1/models` 返回 OpenAI-compatible `object=list`，并包含 provider registry 与 profile inventory。
- `/v1/models/mock` 可通过精确 lookup 返回 mock provider model。
- `/v1/chat/completions` 在 `mock` provider 下返回 advisory 文本，不访问外部 provider，不写回任何上层项目。

## 故障边界

- 若启动时报 `load config`，优先检查 duration / float 类环境变量格式，例如 `RADISHMIND_PLATFORM_BRIDGE_TIMEOUT=30s`、`RADISHMIND_PLATFORM_TEMPERATURE=0`。
- 若 `/v1/models` 返回 `PROVIDER_REGISTRY_UNAVAILABLE`，优先检查 `RADISHMIND_PLATFORM_PYTHON_BIN`、`RADISHMIND_PLATFORM_BRIDGE_SCRIPT` 和当前工作目录是否能访问仓库根下的 Python bridge。
- 若 chat 路由返回 bridge 失败，先用 `python3 scripts/run-platform-bridge.py providers` 与 `python3 scripts/run-platform-bridge.py inventory` 验证 Python 侧 provider registry 是否可用。
- 若要接真实 provider，必须通过环境变量或本机 secret 注入 `RADISHMIND_PLATFORM_BASE_URL` / `RADISHMIND_PLATFORM_API_KEY` 或 provider profile 配置；不要把 key、token、cookie 或真实 provider raw dump 写入 committed 文档。
