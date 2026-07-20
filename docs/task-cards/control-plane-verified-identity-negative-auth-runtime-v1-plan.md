# Control Plane Verified Identity Context & Negative Auth Runtime v1

更新时间：2026-07-12

## 任务定位

任务 ID：`control-plane-verified-identity-negative-auth-runtime-v1`

当前状态：`control_plane_verified_identity_negative_auth_runtime_v1_complete`

本任务承接 [Admin Control Plane Authenticated Read Store Transition v1](../features/admin-control-plane/authenticated-read-store-transition-v1.md) 第一实现批次，为现有七条 Control Plane read route 建立共享 verified identity、显式 permission projection、signed test token 验证和 sanitized HTTP failure。repository 继续使用 fixture-backed fake store，不接 PostgreSQL 或真实 Radish OIDC。

## 输入与前置条件

- 七条既有 Control Plane read route、`ControlPlaneReadRepository` 和 `ReadRepositoryContext`。
- `contracts/radish-oidc-token-validation.schema.json` 的脱敏 verified context 边界。
- 现有 `dev_headers` 本地开发路径与默认 offline Web consumer。
- 上游真实 issuer、client registration、JWKS 和 permission mapping 仍未完成联调审查，因此本批只能使用进程内生成的 RSA test key 或显式 test public key。

## 实现范围

1. 新增 `disabled`、`dev_headers`、`signed_test_token` 三种 read auth mode，并保持旧 dev auth 开关的兼容映射。
2. 建立 `VerifiedControlPlaneIdentity`、`ControlPlaneResourceBinding` 和 route authorization 结果；handler 与 repository 不解析 raw claim。
3. `signed_test_token` 只接受 `RS256`、显式 issuer / audience 和配置的 test public key；验证 signature、时间与 required claim。
4. 通过版本化 allowlist 把 upstream permission 投影为现有七条 route grant，不按角色名授权。
5. authenticated mode 完全拒绝 dev header；Bearer 与任意 dev header 同时出现时 fail closed。
6. repository context 补齐 sanitized issuer / session ref；auth denial 时 repository query count 必须为 0。
7. read envelope 在非 2xx 时仍保持既有 sanitized shape；Web consumer 必须先校验 envelope，再依据 failure code 展示失败。
8. 默认 offline Web 保持零请求，token 不进入 URL、storage、日志或响应。

## 负向认证矩阵

| case | 公开 failure | HTTP | repository query |
| --- | --- | --- | --- |
| missing credential | `identity_context_missing` | `401` | `0` |
| malformed scheme / token | `auth_context_contract_mismatch` | `401` | `0` |
| disallowed algorithm | `auth_context_contract_mismatch` | `401` | `0` |
| invalid signature / key | `auth_context_contract_mismatch` | `401` | `0` |
| issuer mismatch | `auth_context_contract_mismatch` | `401` | `0` |
| audience mismatch | `auth_context_contract_mismatch` | `401` | `0` |
| expired | `auth_context_contract_mismatch` | `401` | `0` |
| not yet valid / future issue time | `auth_context_contract_mismatch` | `401` | `0` |
| required claim invalid | `auth_context_contract_mismatch` | `401` | `0` |
| tenant binding missing | `tenant_binding_missing` | `401` | `0` |
| path tenant mismatch | `tenant_binding_missing` | `403` | `0` |
| permission mapping missing / denied | `scope_denied` | `403` | `0` |
| dev header injection / credential conflict | `auth_context_contract_mismatch` | `401` | `0` |

## Permission Projection

测试策略只允许以下显式映射：

| upstream permission | route grant |
| --- | --- |
| `radishmind.tenant.read` | `tenant:read` |
| `radishmind.applications.read` | `applications:read` |
| `radishmind.api-keys.read` | `api_keys:read` |
| `radishmind.usage.read` | `usage:read` |
| `radishmind.runs.read` | `runs:read` |
| `radishmind.audit.read` | `audit:read` |

未知 permission 不产生 grant。`radish-api` audience / scope 或角色名不产生 RadishMind 权限。

## 配置与启动约束

- `RADISHMIND_CONTROL_PLANE_READ_AUTH_MODE=disabled|dev_headers|signed_test_token`
- `RADISHMIND_CONTROL_PLANE_READ_SIGNED_TEST_ISSUER`
- `RADISHMIND_CONTROL_PLANE_READ_SIGNED_TEST_AUDIENCE`
- `RADISHMIND_CONTROL_PLANE_READ_SIGNED_TEST_PUBLIC_KEY_PEM`

`signed_test_token` 缺少任一验证材料必须拒绝启动。public key 只作为 test verifier 输入，不进入 config summary。真实 OIDC discovery、JWKS fetch、key rotation 和 client secret 不属于本批。

## 验收

- Go：positive token、全部 13 类负向认证、七条 route grant、status mapping、zero repository query、dev header compatibility / isolation、config validation。
- Web：offline 零请求、非 2xx sanitized envelope、非法非 2xx body、secret / Authorization 不进入 consumer state。
- HTTP：Admin tenant summary 与 audit summary signed-token success / denied smoke。
- 浏览器：offline 零请求；显式 dev HTTP 下验证 ready / denied 状态、console、URL 与 storage。
- 仓库：Go test / race / vet、Web unit test / production build、`git diff --check`、fast 与完整 `check-repo`。

## 停止线

- 不实现真实 Radish OIDC、discovery / JWKS 网络读取、PKCE、BFF、session cookie、refresh token 或 production auth。
- 不新增 PostgreSQL adapter、migration、read store selector、业务写入或 fallback。
- 不启用 application promotion、API key lifecycle、quota enforcement、billing、secret runtime 或部署操作。
- 不扩展 Gateway schema / provider registry，也不扩展 Workflow tool、confirmation、writeback、replay 或 resume。
- signed test token 与 fake store 成功只代表 dev/test auth boundary 可复验，不代表 production Admin ready。

## 完成记录

- 已实现 shared verified identity、resource binding、RS256 signed test token、显式 permission mapping 和七条 route authorization。
- 13 类负向认证均返回 sanitized `401 / 403`，repository query count 为 0；invalid filter、store 与 contract failure 也已按稳定 HTTP status 映射。
- Web 保留非 2xx read envelope，offline 继续零请求；真实浏览器覆盖 ready、identity denied、无应用崩溃、URL / storage 无泄漏。
- repository 仍为 fake store；PostgreSQL、真实 Radish OIDC、workspace membership 与 production enablement 未打开。
