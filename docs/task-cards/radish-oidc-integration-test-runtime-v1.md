# Radish OIDC Integration Test Runtime v1 任务卡

更新时间：2026-07-12

状态：`deterministic_runtime_complete_blocked_by_upstream_evidence`

## 任务目标

本任务实现 `radish_oidc_integration_test` 受控认证模式，使 Admin Tenant Summary 与 Audit 能通过真实 discovery、JWKS 和 JWT validation 边界接入现有 PostgreSQL dev/test read repository。当前仓库缺少完整 reviewed Radish upstream evidence，因此本批关闭 deterministic runtime、auth boundary 与 operation gate，但不声明真实 Radish 联调完成。

## 实现批次

### 批次 A：配置、Discovery 与 JWKS

- `radish_oidc_integration_test` 只允许与 `postgres_dev_test` 组合。
- 启动前要求 exact issuer、discovery URL、audience、mapping version、evidence ref、permission identifier、algorithm allowlist 和 JWKS policy 全部显式配置。
- discovery issuer 必须 exact match；JWKS URI 必须符合 reviewed scheme / host / port policy。
- HTTP client 禁止 redirect，限制 timeout、响应大小、content type、JSON shape 和 key count。
- JWKS cache 使用 single-flight refresh、bounded key set、rotation overlap 与 hard expiry；unknown `kid` 每次 token decision 只刷新一次。

### 批次 B：Token、Identity 与 Operation Gate

- 校验 compact JWT、required `kid`、algorithm allowlist、key type、signature、exact issuer / audience、`iat / nbf / exp / auth_time` 和最大 lifetime。
- claim 名称、类型、cardinality 与 upstream permission identifier 只来自显式版本化 mapping 配置。
- 只把 sanitized `VerifiedControlPlaneIdentity` 与 `ControlPlaneResourceBinding` 放入 request context；raw token、raw claim 和 provider response 在 verifier 边界内释放。
- Tenant Summary 只接受投影后的 `tenant:read`；Audit 只接受 `audit:read`。
- Applications、API Keys、Quota、Workflow Definitions、Runs 在该 auth mode 下统一返回 `workspace_membership_unavailable`，repository query count 为 0。

### 批次 C：Deterministic 与真实联调

- 使用测试进程独占的 deterministic issuer 覆盖 discovery、JWKS、rotation、算法混淆、claim/time、permission、tenant mismatch、no-fallback、zero-query 和诊断脱敏。
- 只有 upstream evidence gate 完整且仍在 review window 内，才使用短期 Radish integration token 执行真实 HTTP/Web/浏览器联调。
- 当前真实联调状态固定为 `blocked_by_upstream_evidence`；loopback issuer 只证明 verifier 行为，不作为 Radish evidence。

## 上游证据门禁

真实 Radish 联调前必须由 Radish owner review 并提供 metadata-only manifest，至少包含：

1. environment ref、exact issuer、explicit discovery URL、owner、`reviewed_at` 和 review expiry；
2. exact JWKS scheme / host / port、algorithm allowlist、refresh max-age、rotation overlap、hard expiry、key count和响应大小限制；
3. exact RadishMind integration audience；
4. mapping version、subject claim、tenant claim、permission claim及其 JSON 类型 / cardinality；
5. exact tenant-read 与 audit-read upstream permission identifier；
6. token max lifetime、clock skew、required time claims和短期 token 清理流程。

当前结论：`not_reviewed`。禁止以 localhost 配置、示例 issuer、Radish 通用 API audience、角色名、scope 猜测或本地 token 替代以上证据。

## 验收标准

- 非法 auth/store 组合和缺失 / 不一致 policy 在 listen 前拒绝启动。
- discovery / JWKS origin、redirect、timeout、大小、content type、JSON、algorithm 与 key count 均有确定性负向测试。
- unknown `kid` 单次 refresh；并发 refresh single-flight；rotation overlap 和 hard expiry 可复验。
- issuer、audience、signature、algorithm、required claim、时间窗口、tenant 和 permission 矩阵覆盖完整。
- 两条 Admin route 成功时只把 sanitized identity 传给 repository；所有鉴权、tenant、permission、membership 和 identity provider failure 均为 zero-query。
- OIDC 模式绝不尝试 signed test key、dev header、fake identity、默认 tenant、默认 permission 或 workspace fake repository。
- Go unit / integration / race / vet、Web unit / production build、`git diff --check`、fast 与完整仓库门禁通过。

## 稳定失败语义

| 边界 | failure code | HTTP | repository query |
| --- | --- | --- | --- |
| credential missing | `identity_context_missing` | `401` | `0` |
| malformed token / invalid signature / issuer / audience / algorithm / time / claim | `auth_context_contract_mismatch` | `401` | `0` |
| token tenant claim missing | `tenant_binding_missing` | `401` | `0` |
| request tenant mismatch | `tenant_binding_missing` | `403` | `0` |
| route permission denied | `scope_denied` | `403` | `0` |
| workspace membership unavailable | `workspace_membership_unavailable` | `503` | `0` |
| discovery / JWKS unavailable or hard expired | `identity_provider_unavailable` | `503` | `0` |
| reviewed policy / mapping mismatch | `auth_policy_mismatch` | startup blocked | `0` |
| PostgreSQL failure after successful auth | `read_store_unavailable` | `503` | `1` |

## 隐私要求

- token 只能存在于进程或页面内存；不得进入 URL、命令参数、env build injection、Storage、仓库、日志、截图或构建产物。
- 不记录 Authorization、raw token、raw claims、raw discovery/JWKS、provider body / error、email、角色名或临时文件路径。
- 诊断只允许 auth mode、sanitized evidence ref、mapping version、`kid` ref、algorithm、decision、failure boundary、request / audit ref、时延和 query count。
- repository、HTTP envelope 与 Web view model 不接收 token、claim map、JWK 或 provider client。

## 停止线

- 不实现 production auth、OAuth login、authorization code、PKCE、BFF、callback、cookie、refresh token或长期浏览器会话。
- 不读取或复制 Radish 用户、租户、角色、权限或 membership 数据库。
- 不实现 workspace membership adapter，不开放五类 workspace operation。
- 不新增 northbound API、Gateway schema、repository、provider registry、tenant / audit writer或管理写入。
- 不扩展 production key、quota、billing、secret runtime、Workflow writeback、replay或 resume。
- 不把 deterministic issuer、controlled issuer或 dev/test 联调成功解释为 production ready、production SLA或真实 Radish 联调完成。
