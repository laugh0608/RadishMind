# Control Plane 鉴权与只读运行时契约

更新时间：2026-07-12

## 契约定位

本文说明七条 Control Plane / User Workspace read route 在当前开发测试运行时中的鉴权、repository 选择、失败与隐私边界。长期 read model、response envelope、历史 readiness 证据和 UI contract 仍见 [Control Plane Read-Side 契约](control-plane-read-side.md)。

该运行时只服务受控开发测试，不声明 production auth、production repository 或真实 Radish 联调完成。

## 路由与数据源

| 操作 | 路由 | `fake_store_dev` | `postgres_dev_test` + signed test | `postgres_dev_test` + OIDC integration |
| --- | --- | --- | --- | --- |
| Tenant Summary | `GET /v1/control-plane/tenants/{tenant_ref}/summary` | fake repository | PostgreSQL | PostgreSQL + tenant-read permission |
| Audit | `GET /v1/control-plane/audit` | fake repository | PostgreSQL | PostgreSQL + audit-read permission |
| Applications | `GET /v1/user-workspace/applications` | fake repository | fake binding | `workspace_membership_unavailable` |
| API Keys | `GET /v1/user-workspace/api-keys` | fake repository | fake binding | `workspace_membership_unavailable` |
| Quota | `GET /v1/user-workspace/usage/quota-summary` | fake repository | fake binding | `workspace_membership_unavailable` |
| Workflow Definitions | `GET /v1/user-workspace/workflow-definitions` | fake repository | fake binding | `workspace_membership_unavailable` |
| Runs | `GET /v1/user-workspace/runs` | fake repository | fake binding | `workspace_membership_unavailable` |

OIDC integration 只允许与 `postgres_dev_test` 组合。五条 workspace operation 没有正式 membership owner，必须在 repository 之前拒绝，不能读取 fake repository。

## Auth mode

- `disabled`：所有外部 read 请求 fail closed。
- `dev_headers`：只在显式 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 的本地开发路径接受测试 header。
- `signed_test_token`：验证 RS256 test token、exact issuer / audience、required claims、时间窗口、resource binding 和版本化 permission projection；不连接 discovery / JWKS。
- `radish_oidc_integration_test`：验证受控 discovery、JWKS 与 JWT；不回退 signed token、dev headers 或 fake identity。

所有模式都投影同一内部 `VerifiedControlPlaneIdentity`。handler 和 repository 只能读取脱敏 subject、tenant、permission、issuer / mapping reference、request id 与 audit ref，不得读取 Authorization、raw token 或 raw claims。

## OIDC discovery、JWKS 与 JWT

- issuer 与 audience 必须 exact match；issuer、audience、claim mapping、permission identifier 和 JWKS policy 必须来自 reviewed upstream evidence。
- discovery URL 必须与 issuer 同 origin；JWKS origin 必须与 reviewed policy exact match。
- redirect 禁止；HTTP client 使用固定 timeout、response size 上限、JSON depth / content type 校验和 bounded key count。
- JWT 只接受显式 algorithm allowlist，拒绝 HMAC、`none` 和 algorithm/key-type confusion；必须有 `kid`。
- JWKS refresh 使用 single-flight；unknown `kid` 每个验证请求只触发一次刷新。
- cache 必须 bounded，并执行 refresh max-age、rotation overlap 与 hard expiry；hard expiry 后不得继续使用 stale key。
- required claim、NumericDate、clock skew、最大 token lifetime、tenant binding 和版本化 permission mapping 都必须 fail closed。

deterministic loopback issuer 只用于测试上述 verifier 行为。没有 reviewed Radish evidence 和短期 token 时，不得把 loopback 结果写成真实 Radish 联调。

## Repository 与数据库

`postgres_dev_test` 只承载 Tenant Summary 与 Audit projection：

- migration 由独立 manual CLI 执行，runtime 不执行 DDL。
- startup 必须校验 marker、checksum、schema version 与 runtime `SELECT` 权限。
- runtime role 只读；projection schema 不接收 raw token、raw claims、secret 或 provider payload。
- pagination 使用 strict keyset cursor；query 继承 request context、timeout 和 cancel。
- connect、preflight、permission、query、scan 或 close 失败返回稳定脱敏错误，不回退 fake store。

## 稳定失败与零查询

当前边界至少保留以下稳定失败类别：

- `identity_context_missing`
- `identity_provider_unavailable`
- `auth_context_contract_mismatch`
- `tenant_binding_missing`
- `scope_denied`
- `workspace_membership_unavailable`
- `read_store_unavailable`
- `read_store_contract_mismatch`
- `schema_migration_not_applied`
- `schema_version_mismatch`

identity、issuer / audience / algorithm / signature / time / claim、tenant、permission、membership 和 provider denial 必须在 repository query 前结束，query count 为零。HTTP 层只返回统一 read-side envelope，不返回 provider raw error、discovery/JWKS body、SQL、DSN 或 token 片段。

## Web 消费

`apps/radishmind-web/` 默认离线。显式 dev-live consumer 根据 `VITE_RADISHMIND_READ_AUTH_MODE` 选择 dev headers、signed test token 或 OIDC integration；OIDC token 只能由受控 harness 通过页面内存 closure 提供，不得进入 URL、`.env`、Storage、仓库、日志、截图或构建产物。

consumer 会请求七条 route。OIDC integration 下 Tenant Summary / Audit 可成功，五条 workspace route 的 `workspace_membership_unavailable` 是预期状态；consumer 不尝试 fallback。

## 停止线

- 不实现 production auth、OAuth 登录、PKCE、BFF、callback、cookie、refresh token 或长期浏览器 session。
- 不读取或复制 Radish 用户、租户、角色和权限数据库。
- 不实现 workspace membership adapter，不开放五条 workspace operation。
- 不把 controlled issuer、PostgreSQL dev/test 或浏览器联调写成 production ready 或真实 SLA。
