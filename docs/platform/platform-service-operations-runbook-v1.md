# 平台服务运行手册 v1

更新时间：2026-07-15

## 文档职责

本文档是 `services/platform/` 的本地开发与测试运行手册，维护核心配置、精确基础路由、启动方式、smoke 和故障处理。服务职责与代码入口见 [Platform Service README](../../services/platform/README.md)；产品 API 字段、功能 gate、repository 与权限边界继续以 `docs/contracts/` 和对应 `docs/features/` 专题为准。

本文档不描述 production deployment ready。涉及真实 PostgreSQL、浏览器 dev server、真实 provider 或长驻服务时，由开发者在明确运行窗口中启动，并在验证结束后关闭。

## 本地启动 runbook

首次拉取或缺少 `.venv` 时先运行：

```bash
./scripts/bootstrap-dev.sh
```

平台层单元测试：

```bash
cd services/platform
GOCACHE=/tmp/radishmind-go-build-cache go test ./...
```

从仓库根目录直接启动最小 mock 服务：

```bash
RADISHMIND_PLATFORM_LISTEN_ADDR=127.0.0.1:7000 \
RADISHMIND_PLATFORM_PROVIDER=mock \
RADISHMIND_PLATFORM_MODEL=radishmind-local-dev \
go run ./services/platform/cmd/radishmind-platform
```

进入 `services/platform` 后可使用：

```bash
go run ./cmd/radishmind-platform
```

稳定 wrapper 支持 `serve`、`config-summary`、`config-check` 和 `diagnostics`：

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

wrapper 默认使用 `local-product` 档：选择聚合 `sqlite_dev`、共享 SQLite 文件和七组件开发 gate。`configured` 档不注入 store、数据库连接或开发 gate，只消费调用方显式配置，供 PostgreSQL 集成和故障注入使用：

```bash
./scripts/run-platform-service.sh --profile configured config-check
./scripts/run-platform-service.sh --profile configured serve
```

配置文件启动方式：

```bash
RADISHMIND_PLATFORM_CONFIG=tmp/radishmind-platform.local.json \
go run ./services/platform/cmd/radishmind-platform
```

命令行也可直接执行：

```bash
go run ./services/platform/cmd/radishmind-platform config-summary
go run ./services/platform/cmd/radishmind-platform config-check
go run ./services/platform/cmd/radishmind-platform diagnostics
```

启动链固定为 `config.LoadFromEnv -> httpapi.NewServer -> ListenAndServe`。默认 `stdio_pool` 在监听端口前完成 Python worker protocol v1 handshake；`process_per_request` 只作为显式回滚模式。`config-summary`、`config-check` 和 `diagnostics` 不创建 SQLite 文件或执行 migration，只有 `serve` 进入 shared runtime 生命周期。

## 环境变量

配置优先级固定为 `default < config file < env`。核心平台配置如下；源码真相源为 `services/platform/internal/config/config.go`。

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `RADISHMIND_PLATFORM_CONFIG` | 空 | 可选 JSON 配置文件路径 |
| `RADISHMIND_PLATFORM_LISTEN_ADDR` | `:7000` | HTTP 监听地址 |
| `RADISHMIND_PLATFORM_READ_HEADER_TIMEOUT` | `5s` | HTTP header 读取超时 |
| `RADISHMIND_PLATFORM_WRITE_TIMEOUT` | `30s` | HTTP 写响应超时 |
| `RADISHMIND_PLATFORM_BRIDGE_TIMEOUT` | `30s` | Go 调 Python bridge 的请求超时 |
| `RADISHMIND_PLATFORM_BRIDGE_MODE` | `stdio_pool` | `stdio_pool` 或显式 `process_per_request` |
| `RADISHMIND_PLATFORM_BRIDGE_WORKER_COUNT` | `4` | worker 数量上限 |
| `RADISHMIND_PLATFORM_BRIDGE_QUEUE_CAPACITY` | `64` | worker 忙时的有界等待容量 |
| `RADISHMIND_PLATFORM_BRIDGE_HANDSHAKE_TIMEOUT` | `5s` | worker 启动握手超时 |
| `RADISHMIND_PLATFORM_PYTHON_BIN` | 仓库 `.venv` Python | bridge Python 解释器 |
| `RADISHMIND_PLATFORM_BRIDGE_SCRIPT` | `scripts/run-platform-bridge.py` | bridge 脚本路径 |
| `RADISHMIND_PLATFORM_PROVIDER` | `mock` | 默认 southbound provider |
| `RADISHMIND_PLATFORM_PROVIDER_PROFILE` | 空 | 默认 provider profile |
| `RADISHMIND_PLATFORM_MODEL` | 空 | 默认 northbound model id |
| `RADISHMIND_PLATFORM_BASE_URL` | 空 | provider base URL 显式覆盖 |
| `RADISHMIND_PLATFORM_API_KEY` | 空 | provider API key 显式覆盖；不得提交或进入摘要 |
| `RADISHMIND_PLATFORM_TEMPERATURE` | `0` | provider 调用温度 |

产品组件配置按职责维护：

- 聚合 `sqlite_dev`、共享数据库路径和七组件 store 投影见[本地 SQLite 开发持久化 v1](local-sqlite-dev-persistence-v1.md)。
- 应用目录、API 密钥生命周期和 Gateway Bearer 开发测试态认证见[应用目录与 API 密钥开发测试指南](../features/user-workspace/application-catalog-api-key-dev-test-guide.md)。
- Workflow 草案、运行、评测与执行 gate 见 [Workflow 专题](../features/workflow/README.md)。
- Application Draft / Publish gate 见 [User Workspace 专题](../features/user-workspace/README.md)。
- Control Plane auth / store / OIDC integration test 见 [Admin Control Plane 专题](../features/admin-control-plane/README.md)。

所有未知 selector、保留 production 值和不完整数据库配置都在启动前失败关闭。组件显式 `*_STORE` 与聚合 `RADISHMIND_LOCAL_PERSISTENCE_MODE=sqlite_dev` 不得同时设置；数据库、migration、marker、checksum 或查询失败不得回退 `memory_dev`。

## 基础路由清单

以下清单由 `services/platform/internal/httpapi/server.go` 直接注册；产品子路由由各自 handler 按显式 gate 管理。

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

完整路由按服务 README 的六类入口导航到对应协议或功能专题。路由注册不表示默认启用；User Workspace、Workflow、Gateway history 与 Admin 路由必须满足各自 auth、scope、dev/test gate 和 store selector。

## 本地 smoke 验证

服务启动后执行：

```bash
curl -sS http://127.0.0.1:7000/healthz
curl -sS http://127.0.0.1:7000/v1/platform/local-smoke
curl -sS http://127.0.0.1:7000/v1/models
curl -sS http://127.0.0.1:7000/v1/models/mock
curl -sS http://127.0.0.1:7000/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"radishmind-local-dev","messages":[{"role":"user","content":"请简要说明当前 RadishMind 平台状态。"}]}'
```

上层消费 smoke 默认读取离线 fixture，不要求启动服务：

```bash
./scripts/run-python.sh scripts/run-platform-overview-consumer-smoke.py --check
./scripts/run-python.sh scripts/run-platform-local-smoke.py --check
./scripts/run-python.sh scripts/run-platform-session-tooling-consumer-smoke.py --check
```

需要连接已启动服务时增加 `--base-url http://127.0.0.1:7000 --check`。这些 smoke 只验证只读 overview、local readiness、session/tooling metadata 与 blocked action no-side-effects，不启动 executor、confirmation、业务写回或 replay。

运行手册漂移由以下既有检查承担：

```bash
./scripts/run-python.sh scripts/check-platform-deployment-smoke.py
./scripts/run-python.sh scripts/check-platform-diagnostics.py
./scripts/run-python.sh scripts/check-platform-runbook.py
```

## PostgreSQL 开发测试态验证

兼容入口实际覆盖八组平台 migration 和七组件 repository：

```bash
./scripts/run-workflow-saved-draft-postgres-dev-test.sh check
./scripts/run-workflow-saved-draft-postgres-dev-test.sh status
./scripts/run-workflow-saved-draft-postgres-dev-test.sh down
```

`check` 会保留容器与已迁移 schema 供后续联调；验证结束后必须显式执行 `down`。migration URL 只供 `up` 使用，runtime URL 只授予必要 DML；两者不得互换、输出或提交。该链只证明 PostgreSQL 开发测试态能力，不启用 production repository。

## 故障边界

- 启动前先执行 `./scripts/run-platform-service.sh diagnostics`。`status=error` 时优先读取结构化 `failure.code` 和 `checks[].code`，不从日志寻找 secret 原文。
- `CONFIG_REQUIRED_FIELDS_MISSING`：检查 `config.missing_required_fields` 和字段来源。
- `PROVIDER_REGISTRY_UNAVAILABLE` / `PROVIDER_INVENTORY_UNAVAILABLE`：检查 `RADISHMIND_PLATFORM_PYTHON_BIN`、`RADISHMIND_PLATFORM_BRIDGE_SCRIPT`、工作目录和 Python import 路径。
- `BRIDGE_WORKER_*`：检查 bridge mode、handshake 和 inventory；可用 `RADISHMIND_PLATFORM_BRIDGE_MODE=process_per_request` 复核显式回滚路径，但不得恢复 argv 传递 credential。
- `MODEL_NOT_FOUND`：通过 `/v1/models` 或 diagnostics inventory 确认可选择 ID，不混用 provider id、profile id 与 upstream model。
- `PLATFORM_RESPONSE_INVALID`：检查 canonical envelope 的 `response` 与 northbound 协议转换，不把 provider 原始正文写回公开错误。
- repository / migration / marker / checksum / auth / scope 错误必须保持稳定失败语义和 no fallback；具体 failure code 以对应功能专题与实现测试为准。
- 真实 provider credential 只能通过本机环境或受控 secret 注入。API key、token、cookie、Authorization、DSN、provider raw URL / response 和异常正文不得进入 argv、公开错误、日志、fixture 或 committed 文档。

## 停止线

- 本手册只覆盖本地开发和测试运行，不是 production deployment runbook。
- 不把 local smoke、deterministic OIDC integration test、SQLite / PostgreSQL dev-test、fake / mock provider 或本地 console 写成生产就绪。
- 不在缺少 production secret backend、process supervisor、环境隔离、正式身份与发布复核时启用生产入口。
- 不用历史 readiness 文档代替当前配置、协议、功能专题和可执行测试。
