# RadishMind Platform Service Layer

本目录承载 `Go` 平台服务层，统一管理 northbound HTTP、Gateway bridge、开发测试态认证、产品 repository 选择与服务生命周期。它是运行时实现入口，不是产品规划、历史 readiness 或完整运维手册的真相源。

## 服务职责

- 加载并校验平台配置，构造服务依赖，管理监听、优雅关闭和本地开发持久化生命周期。
- 承载 Model Gateway、User Workspace、Admin Control Plane、Workflow / Agent Runtime 的 HTTP 边界。
- 将 OpenAI-compatible、Responses 与 Messages 请求翻译到 canonical `CopilotGatewayEnvelope`，通过受控 Python bridge 调用 runtime。
- 为 RadishMind 自有运行数据提供 `memory_dev`、聚合 `sqlite_dev` 与显式 `postgres_dev_test` repository 选择，并保持 migration、作用域和 no-fallback 约束。
- 输出结构化 diagnostics、request observability、本地 overview 和 local smoke 摘要。

本服务已提供 RadishMind 自有本地账户、外部身份绑定、Web Session、角色、工作区成员关系与一次性 OIDC authorization transaction 的领域契约及 memory / SQLite / PostgreSQL dev/test repository；SQLite migration 进入 shared runtime，PostgreSQL 只通过 `radishmind-local-identity-migrate` 显式执行。显式 `local_session_dev_test` 模式已提供本地注册 / 登录、browser Authorization Code + PKCE、当前账户、当前账户 session directory、exact / bulk session revoke、credential rotation、session actor → local membership，以及 workspace member / role administration strict HTTP；身份与管理 mutation 继续要求对应的 exact scope / permission、CSRF / Origin、近期认证和显式确认，版本化 exact mutation 另要求 expected version。服务不复制 `Radish` 的身份数据库、组织成员关系或业务数据真相源，也不绕过 `contracts/` 自定义第二套协议。

## 路由分类

| 分类 | 路由范围 | 权威说明 |
| --- | --- | --- |
| 服务状态与本地运维 | `/healthz`、`/v1/platform/*` | [平台服务运行手册](../../docs/platform/platform-service-operations-runbook-v1.md) |
| Model Gateway / API Distribution | `/v1/models*`、`/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`/v1/model-gateway/requests*` | [服务 API 契约](../../docs/contracts/service-api.md)、[Gateway 功能专题](../../docs/features/gateway/README.md) |
| Session / Tooling metadata shell | `/v1/session/*`、`/v1/tools/*` | [Session 契约](../../docs/contracts/session.md)、[Tooling 契约](../../docs/contracts/tooling.md) |
| 本地身份（开发 / 测试态） | `/v1/auth/local/*`、`/v1/auth/oidc/*`、`/v1/auth/session`、`/v1/auth/account`、`/v1/auth/logout`、`/v1/auth/sessions`、`/v1/auth/sessions/*`、`/v1/auth/external-identities/*`、`/v1/admin/local-identity/*` | [本地账户与 Radish OIDC 联合登录 v1](../../docs/features/admin-control-plane/local-account-radish-oidc-federated-login-v1.md)、[本地账户凭证轮换与自助会话治理 v1](../../docs/features/admin-control-plane/local-account-credential-rotation-self-service-session-governance-dev-test-v1.md)、[本地用户、角色与工作区成员管理 v1](../../docs/features/admin-control-plane/local-user-role-workspace-membership-administration-dev-test-v1.md) |
| Admin Control Plane | `/v1/control-plane/*` | [Control Plane read-side 契约](../../docs/contracts/control-plane-read-side.md)、[Admin Control Plane 专题](../../docs/features/admin-control-plane/README.md) |
| User Workspace | `/v1/user-workspace/applications*`、`/v1/user-workspace/application-sessions*`（包含 Session-scoped 结果资产 list / read / archive / unarchive）、`/v1/user-workspace/applications/{application_id}/result-artifacts*`（application-scoped list / export）、`/v1/user-workspace/api-keys*`、`/v1/user-workspace/application-configuration-drafts*`、`/v1/user-workspace/application-publish-candidates*`、`/v1/user-workspace/prompt-application-templates*`、`/v1/user-workspace/agent-copilot-profiles*`、两类 `/v1/user-workspace/applications/{application_id}/*-runtime-assignment*` | [User Workspace 专题](../../docs/features/user-workspace/README.md)、[应用结果资产专题](../../docs/features/user-workspace/application-session-result-artifact-explicit-retention-dev-test-v1.md)、[应用结果资产库专题](../../docs/features/user-workspace/application-result-artifact-library-controlled-export-dev-test-v1.md)、[Prompt Application 使用指南](../../docs/features/user-workspace/prompt-application-dev-test-usage-guide.md)、[Agent / Copilot 使用指南](../../docs/features/user-workspace/agent-copilot-dev-test-usage-guide.md) |
| Workflow / Agent Runtime | `/v1/user-workspace/workflow-drafts*`、`/v1/user-workspace/workflow-runs*`、`/v1/user-workspace/workflow-definition-*`、`/v1/user-workspace/workflow-definitions*`、`/v1/user-workspace/workflow-template-candidates*`、`/v1/user-workspace/workflow-templates*`、`/v1/user-workspace/workflow-evaluation-*`、`/v1/user-workspace/workflow-retrieval-snapshots*`、`/v1/user-workspace/workflow-rag-*`、`/v1/application-rag/invocations`、`/v1/agent-copilot/invocations` | [Workflow 专题](../../docs/features/workflow/README.md)、[工作区 Workflow 模板目录专题](../../docs/features/workflow/workspace-workflow-template-catalog-review-controlled-derivation-dev-test-v1.md)、[Workflow RAG 开发测试态指南](../../docs/features/workflow/workflow-rag-dev-test-usage-governance-guide.md)、[Agent / Copilot 使用指南](../../docs/features/user-workspace/agent-copilot-dev-test-usage-guide.md) |

精确路由、核心配置、启动命令、smoke 和故障处理统一维护在[平台服务运行手册](../../docs/platform/platform-service-operations-runbook-v1.md)。schema、字段与失败 envelope 以 `contracts/` 和对应功能专题为准；README 不重复保存逐批 readiness 状态。Action Safety 当前没有独立 HTTP route：规则结果只在 response、candidate、assignment、Tool plan、pre-dispatch 与 Run 的服务端 checkpoint 中计算，并作为版本化、脱敏 projection 进入既有 owner；客户端不能提交 decision 或把 snapshot 当作 execution token。

## Browser OIDC 开发测试配置

browser OIDC 只在 `RADISHMIND_CONTROL_PLANE_READ_AUTH_MODE=local_session_dev_test` 与 `RADISHMIND_LOCAL_IDENTITY_DEV_HTTP=true` 已成立时允许显式开启。当前使用 public client + PKCE，不接收 client secret；若真实 Radish 要求 confidential client，必须留到批次 E 单独评审 secret ref、部署与泄漏响应。

- `RADISHMIND_LOCAL_IDENTITY_OIDC_DEV=true`
- `RADISHMIND_LOCAL_IDENTITY_OIDC_ISSUER`
- `RADISHMIND_LOCAL_IDENTITY_OIDC_DISCOVERY_URL`
- `RADISHMIND_LOCAL_IDENTITY_OIDC_CLIENT_ID`
- `RADISHMIND_LOCAL_IDENTITY_OIDC_REDIRECT_URI`，必须等于 allowed origin 下的 `/v1/auth/oidc/callback`
- `RADISHMIND_LOCAL_IDENTITY_OIDC_SCOPES`，逗号分隔且必须包含 `openid`
- `RADISHMIND_LOCAL_IDENTITY_OIDC_ALGORITHMS`，只允许显式 `RS* / ES*` allowlist
- `RADISHMIND_LOCAL_IDENTITY_OIDC_JWKS_ORIGIN`，必须等于 issuer origin
- `RADISHMIND_LOCAL_IDENTITY_OIDC_TRANSACTION_TTL`，默认 `5m`，最大 `15m`
- `RADISHMIND_LOCAL_IDENTITY_OIDC_FIRST_LOGIN=true` 仅表示显式开发测试首登准入，默认关闭

服务启动会先执行 bounded discovery / JWKS preflight；配置漂移、provider 不可用或 policy 不匹配都会阻止 OIDC client 初始化，不回退本地管理员、dev header 或旧缓存身份。

## 显式首管理员 bootstrap（仅开发 / 测试态）

`radishmind-local-identity-bootstrap` 只对已经存在且 active 的 exact `user_id` 建立首个 workspace membership 与 canonical `workspace_admin` assignment。它只接受 `sqlite_dev | postgres_dev_test`，同一 tenant / workspace 已存在 active identity administrator 时失败关闭；不会由注册或 Server 启动自动执行，也没有 HTTP route。数据库位置只从环境变量读取，不通过 argv 或 JSON 输出。

SQLite 使用共享本地产品数据库：

```bash
RADISHMIND_SQLITE_DEV_DATABASE_PATH=/absolute/path/to/radishmind.db \
go run ./cmd/radishmind-local-identity-bootstrap \
  --store sqlite_dev \
  --tenant-ref tenant_demo \
  --workspace-id workspace_demo \
  --user-id usr_0000000000000001 \
  --audit-ref audit:bootstrap-workspace-admin
```

PostgreSQL 必须先由 migration identity 显式应用当前 `0004_local_identity_self_service_sessions`，再用受限 runtime URL 执行 bootstrap：

```bash
RADISHMIND_LOCAL_IDENTITY_DEV_TEST_MIGRATION_DATABASE_URL='<migration-url>' \
go run ./cmd/radishmind-local-identity-migrate up

RADISHMIND_LOCAL_IDENTITY_DEV_TEST_DATABASE_URL='<runtime-url>' \
go run ./cmd/radishmind-local-identity-bootstrap \
  --store postgres_dev_test \
  --tenant-ref tenant_demo \
  --workspace-id workspace_demo \
  --user-id usr_0000000000000001 \
  --audit-ref audit:bootstrap-workspace-admin
```

可选 `RADISHMIND_LOCAL_IDENTITY_DATABASE_TIMEOUT` 使用 Go duration，默认 `30s`。命令只输出脱敏 JSON 中的 store、scope、stable membership / assignment id、role catalog metadata 与 audit ref；不输出数据库路径、URL、credential、session 或账户登录标识。重复执行不会返回既有记录，而是以 `local_identity_admin_bootstrap_denied` 拒绝。

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

本地产品与 API 密钥连续链使用[应用目录与 API 密钥开发测试指南](../../docs/features/user-workspace/application-catalog-api-key-dev-test-guide.md)；Application RAG、Workflow Definition v1 / v2、Application Session v1 / v4、应用结果资产库、v4 / v5 / v8 历史与运行观测使用[应用受控运行开发测试态指南](../../docs/features/user-workspace/application-controlled-runtime-dev-test-guide.md)；Prompt Template、Configuration Draft v3、Publish Candidate v3 与 Runtime Assignment 使用[Prompt Application 开发测试态使用指南](../../docs/features/user-workspace/prompt-application-dev-test-usage-guide.md)；Agent Profile、Configuration / Candidate v4、Runtime Assignment、Session v3 与 Run v7 使用[Agent / Copilot 开发测试态使用指南](../../docs/features/user-workspace/agent-copilot-dev-test-usage-guide.md)；Workflow RAG 的 snapshot、retrieval execution、evaluation dataset、promotion、binding 与发布重校验使用[Workflow RAG 开发测试态使用与资源治理指南](../../docs/features/workflow/workflow-rag-dev-test-usage-governance-guide.md)；聚合 SQLite 的启动档、migration 与双数据库门禁见[本地 SQLite 开发持久化 v1](../../docs/platform/local-sqlite-dev-persistence-v1.md)。

## 代码入口

| 职责 | 入口 |
| --- | --- |
| 命令与服务生命周期 | `cmd/radishmind-platform/` |
| 本地身份迁移与显式首管理员 bootstrap | `cmd/radishmind-local-identity-migrate/`、`cmd/radishmind-local-identity-bootstrap/` |
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
- [Workflow RAG 开发测试态使用与资源治理指南](../../docs/features/workflow/workflow-rag-dev-test-usage-governance-guide.md)
- [应用受控运行开发测试态指南](../../docs/features/user-workspace/application-controlled-runtime-dev-test-guide.md)
- [API 密钥生命周期与 Gateway 开发测试态认证 v1](../../docs/features/user-workspace/api-key-lifecycle-gateway-dev-test-auth-v1.md)
- [应用交互会话与受控运行编排（开发 / 测试态）v1](../../docs/features/user-workspace/application-interaction-session-controlled-runtime-orchestration-dev-test-v1.md)
- [应用结果资产库与受控导出（开发 / 测试态）v1](../../docs/features/user-workspace/application-result-artifact-library-controlled-export-dev-test-v1.md)
- [提示词应用模板版本审查与受控调用（开发 / 测试态）v1](../../docs/features/user-workspace/prompt-application-template-version-review-controlled-invocation-dev-test-v1.md)
- [Prompt Application 开发测试态使用指南](../../docs/features/user-workspace/prompt-application-dev-test-usage-guide.md)
- [Agent / Copilot 开发测试态使用指南](../../docs/features/user-workspace/agent-copilot-dev-test-usage-guide.md)
- [Production Ops Hardening v1 历史任务卡](../../docs/task-cards/production-ops-hardening-v1-plan.md)

## 停止线

- 当前能力属于内部开发者预览；`local-product`、`sqlite_dev`、`postgres_dev_test`、deterministic OIDC integration test 和本地 console 均不构成生产声明。
- production secret backend、process supervisor、部署环境隔离、console production packaging、生产认证、生产 API key、quota 和 billing 未满足时保持失败关闭。
- 不把 readiness 文档、fixture 或静态检查结果解释为运行能力；不从失败路径回退 `memory_dev` 或 fake store。
- 不让 API key、token、DSN、provider 原始响应或异常正文进入 argv、公开错误、日志和 committed 资产。
