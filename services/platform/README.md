# RadishMind Platform Service Layer

本目录承载 `Go` 平台服务层，统一管理 northbound HTTP、Gateway bridge、开发测试态认证、产品 repository 选择与服务生命周期。它是运行时实现入口，不是产品规划、历史 readiness 或完整运维手册的真相源。

## 服务职责

- 加载并校验平台配置，构造服务依赖，管理监听、优雅关闭和本地开发持久化生命周期。
- 承载 Model Gateway、User Workspace、Admin Control Plane、Workflow / Agent Runtime 的 HTTP 边界。
- 将 OpenAI-compatible、Responses 与 Messages 请求翻译到 canonical `CopilotGatewayEnvelope`，通过受控 Python bridge 调用 runtime。
- 为 RadishMind 自有运行数据提供 `memory_dev`、聚合 `sqlite_dev` 与显式 `postgres_dev_test` repository 选择，并保持 migration、作用域和 no-fallback 约束。
- 输出结构化 diagnostics、request observability、本地 overview 和 local smoke 摘要。

本服务不复制 `Radish` 的身份、组织成员关系或业务数据真相源，不重写模型训练 / 推理 / 评测逻辑，也不绕过 `contracts/` 自定义第二套协议。路由注册不等于能力默认启用；各产品路由仍受显式 dev/test gate、身份、作用域和 store selector 约束。

## 路由分类

| 分类 | 路由范围 | 权威说明 |
| --- | --- | --- |
| 服务状态与本地运维 | `/healthz`、`/v1/platform/*` | [平台服务运行手册](../../docs/platform/platform-service-operations-runbook-v1.md) |
| Model Gateway / API Distribution | `/v1/models*`、`/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/model-gateway/requests*` | [服务 API 契约](../../docs/contracts/service-api.md)、[Gateway 功能专题](../../docs/features/gateway/README.md) |
| Session / Tooling metadata shell | `/v1/session/*`、`/v1/tools/*` | [Session 与 Tooling 契约](../../docs/contracts/session-and-tooling.md) |
| Admin Control Plane | `/v1/control-plane/*` | [Control Plane read-side 契约](../../docs/contracts/control-plane-read-side.md)、[Admin Control Plane 专题](../../docs/features/admin-control-plane/README.md) |
| User Workspace | `/v1/user-workspace/applications*`、`/v1/user-workspace/api-keys*`、`/v1/user-workspace/application-drafts*`、`/v1/user-workspace/application-publish-candidates*` | [User Workspace 专题](../../docs/features/user-workspace/README.md) |
| Workflow / Agent Runtime | `/v1/user-workspace/workflow-drafts*`、`/v1/user-workspace/workflow-runs*`、`/v1/user-workspace/workflow-evaluation-*` | [Workflow 专题](../../docs/features/workflow/README.md) |

精确路由、核心配置、启动命令、smoke 和故障处理统一维护在[平台服务运行手册](../../docs/platform/platform-service-operations-runbook-v1.md)。schema、字段与失败 envelope 以 `contracts/` 和对应功能专题为准；README 不重复保存逐批 readiness 状态。

## 启动入口

从仓库根目录先准备开发环境并执行配置检查：

```bash
./scripts/bootstrap-dev.sh
./scripts/run-platform-service.sh config-check
./scripts/run-platform-service.sh diagnostics
./scripts/run-platform-service.sh serve
```

Windows / PowerShell 使用：

```powershell
pwsh ./scripts/bootstrap-dev.ps1
pwsh ./scripts/run-platform-service.ps1 -Command config-check
pwsh ./scripts/run-platform-service.ps1 -Command diagnostics
pwsh ./scripts/run-platform-service.ps1 -Command serve
```

wrapper 默认使用 `local-product` 档；显式组件配置、PostgreSQL 专项验收和故障注入使用 `configured` 档。`config-summary`、`config-check` 与 `diagnostics` 不创建 SQLite 文件或执行 migration，只有 `serve` 进入 shared runtime 生命周期。长驻服务、真实 PostgreSQL、浏览器联调和真实 provider 应按运行手册由开发者在明确窗口中启动与关闭。

本地产品与 API 密钥连续链使用[应用目录与 API 密钥开发测试指南](../../docs/features/user-workspace/application-catalog-api-key-dev-test-guide.md)；聚合 SQLite 的启动档、migration 与双数据库门禁见[本地 SQLite 开发持久化 v1](../../docs/platform/local-sqlite-dev-persistence-v1.md)。

## 代码入口

| 职责 | 入口 |
| --- | --- |
| 命令与服务生命周期 | `cmd/radishmind-platform/` |
| 配置 | `internal/config/` |
| HTTP 与 northbound 协议适配 | `internal/httpapi/` |
| Python bridge 与 worker pool | `internal/bridge/` |
| diagnostics 与观测 | `internal/diagnostics/`、`internal/httpapi/observability.go` |
| SQLite 开发运行时 | `internal/sqlitedev/` |
| 产品 repository | `internal/*repository*` 与对应 migration 包 |
| production secret 抽象 | `internal/secretbackend/` |

## 权威专题

- [平台专题入口](../../docs/platform/README.md)
- [平台服务运行手册](../../docs/platform/platform-service-operations-runbook-v1.md)
- [工程健康与产品化整改专题 v1](../../docs/platform/engineering-health-productization-remediation-v1.md)
- [Gateway Python Bridge Runtime v1](../../docs/features/gateway/python-bridge-runtime-v1.md)
- [本地 SQLite 开发持久化 v1](../../docs/platform/local-sqlite-dev-persistence-v1.md)
- [Admin Control Plane Authenticated Read Store Transition v1](../../docs/features/admin-control-plane/authenticated-read-store-transition-v1.md)
- [Saved Workflow Draft v1](../../docs/features/workflow/saved-workflow-draft-v1.md)
- [API 密钥生命周期与 Gateway 开发测试态认证 v1](../../docs/features/user-workspace/api-key-lifecycle-gateway-dev-test-auth-v1.md)
- [Production Ops Hardening v1 历史任务卡](../../docs/task-cards/production-ops-hardening-v1-plan.md)

## 停止线

- 当前能力属于内部开发者预览；`local-product`、`sqlite_dev`、`postgres_dev_test`、deterministic OIDC integration test 和本地 console 均不构成生产声明。
- production secret backend、process supervisor、部署环境隔离、console production packaging、生产认证、生产 API key、quota 和 billing 未满足时保持失败关闭。
- 不把 readiness 文档、fixture 或静态检查结果解释为运行能力；不从失败路径回退 `memory_dev` 或 fake store。
- 不让 API key、token、DSN、provider 原始响应或异常正文进入 argv、公开错误、日志和 committed 资产。
