# Radish OIDC Integration Test v1

更新时间：2026-07-12

状态：`radish_oidc_integration_test_runtime_v1_deterministic_complete_blocked_by_upstream_evidence`

## 功能定位

本功能承接已经完成的 [Authenticated Read Store Transition v1](authenticated-read-store-transition-v1.md) 与 [Tenant / Audit PostgreSQL Read Repository v1](tenant-audit-postgresql-read-repository-v1.md)，为 Admin tenant summary 和 audit summary 建立可复验的真实 Radish OIDC 集成测试路径。

它验证 RadishMind 作为 resource server 能否从受控 Radish 集成环境读取 discovery / JWKS、校验短期 Bearer token、投影 tenant 与 permission，并把 sanitized verified context 交给现有 auth boundary 和 PostgreSQL read repository。它不是浏览器登录、production auth、workspace membership 或公开生产 API。

## 当前事实与缺口

- `disabled / dev_headers / signed_test_token`、RS256 signature、issuer / audience / time claim、13 类负向认证和 route permission projection 已完成。
- `signed_test_token` 使用显式 test public key，不进行 discovery / JWKS 网络读取，也不代表真实 Radish token shape。
- Admin tenant / audit PostgreSQL dev/test runtime、read-only role、no-fallback 和真实浏览器路径已完成。
- `contracts/radish-oidc-token-validation.schema.json` 只描述 verified sanitized context，不描述 raw token、raw claim 或 membership record。
- deterministic runtime 已实现 exact issuer / audience、显式 claim mapping、受限 discovery / JWKS、algorithm / `kid`、single-flight、rotation overlap / hard expiry、required claim / time window和权限投影。
- 两条 Admin operation 已接入现有 auth boundary；五条 workspace operation 统一返回 `workspace_membership_unavailable`，鉴权和 membership denial 均为 repository zero-query。
- Web consumer 支持独立 OIDC integration token provider，token 只存在页面内存，不回退 signed test token 或 dev headers。
- 当前仍没有 reviewed Radish issuer、discovery document、JWKS URI、signing algorithm、resource audience 或 claim mapping evidence。
- 当前没有 workspace / application membership data source、cache policy 或 owner；真实 OIDC token 不能替代该缺口。
- 历史 upstream evidence / readiness 文档保留为归档输入，不再派生同层 checker 链；本设计以一个后续高风险任务卡承接实现和联调。

## 用户与运维流程

### Admin 读取联调

1. 操作者在受控 Radish integration 环境获得一个短期、仅面向 RadishMind integration audience 的测试 access token。
2. token 只通过本机进程内环境、受限临时文件或标准输入交给 HTTP / browser test harness，不进入命令参数、URL、仓库文件、日志或 storage。
3. Platform 以 `radish_oidc_integration_test` auth mode 启动，先校验 reviewed issuer 配置，再读取 discovery 和 JWKS。
4. 请求携带 Bearer token 访问 tenant summary 或 audit summary。
5. verifier 校验 signature、issuer、audience、algorithm、`kid`、时间窗口和 required claim，再按版本化 mapping 投影 verified identity。
6. auth boundary 校验 tenant path / query binding 和显式 permission，repository 只消费 sanitized context。
7. Web 只在显式 integration-test 模式从浏览器组件内存读取 token；页面刷新即丢失，不持久化。
8. 联调完成后关闭 Platform/Web，删除临时 token 与 JWKS evidence，不保留可重用 credential。

### 失败审查

1. discovery、JWKS、token 或 mapping 失败必须返回稳定 sanitized failure。
2. 失败时 repository query count 为 0；不得回退 `signed_test_token`、dev headers、cached fake identity 或默认 tenant。
3. operator evidence 只记录 issuer evidence ref、mapping version、`kid` reference、algorithm、decision、failure boundary、HTTP status、request/audit ref 和时延。
4. 不记录 raw token、Authorization header、raw discovery/JWKS、raw claim、subject email、角色名称或 provider error body。

## Operation 范围

| route | required upstream permission | repository | 本批结论 |
| --- | --- | --- | --- |
| `/v1/control-plane/tenants/{tenant_ref}/summary` | reviewed tenant-read permission | PostgreSQL dev/test | 开放集成测试 |
| `/v1/control-plane/audit` | reviewed audit-read permission | PostgreSQL dev/test | 开放集成测试 |
| Applications | membership + application permission | fake repository currently | integration mode fail closed |
| API keys | membership + API key permission | fake repository currently | integration mode fail closed |
| quota | membership + usage permission | fake repository currently | integration mode fail closed |
| workflow definitions | membership + application permission | fake repository currently | integration mode fail closed |
| runs | membership + run permission | fake repository currently | integration mode fail closed |

五条 workspace route 在 `radish_oidc_integration_test` 下不得读取 fake repository。即使 token 携带同名 permission，也必须返回 `workspace_membership_unavailable`，直到独立 Workspace Membership Contract & Adapter 设计完成。

## Upstream Evidence Gate

进入 runtime task card 前必须由 Radish owner 提供并审查以下 metadata-only evidence：

| evidence | 必需内容 | 禁止内容 |
| --- | --- | --- |
| issuer registration | environment ref、exact issuer、owner、reviewed_at、expiry / review window | client secret、operator credential |
| discovery | issuer、authorization metadata availability、JWKS URI origin、supported algorithm refs | raw document dump、未知 endpoint |
| JWKS policy | allowed origin、refresh / max-age、key rotation owner、`kid` policy、size limit | committed key set、private key |
| resource audience | exact audience、RadishMind integration resource ref、environment binding | wildcard audience、shared production credential |
| claim mapping | mapping version、subject claim、tenant claim、permission claim、type/cardinality | raw token sample、email/role inference |
| token policy | max lifetime、clock skew、required time claims、allowed algorithms | token header-controlled policy |
| test token issuer | issuance owner、short TTL、revocation / cleanup procedure | username/password、refresh token |

证据可以引用 `https://github.com/laugh0608/Radish` 中 reviewed contract 或由 Radish owner 提供的 integration manifest；长期文档不写本机 Radish 路径。没有 exact value 时保留 `not_reviewed`，不得用猜测值、localhost 默认值或示例 issuer 进入实现。

## Discovery 与 JWKS 边界

- issuer 必须来自显式 allowlist，并与 discovery `issuer` 完全一致；不接受 suffix、substring 或动态 tenant override。
- production-like integration issuer 默认要求 HTTPS。仅由独立测试进程拥有的 loopback issuer 可使用 HTTP，且不能作为 Radish upstream evidence。
- discovery URL 固定为 issuer 的 reviewed metadata endpoint；禁止请求参数覆盖、跨 origin redirect 或从 token claim 推导 URL。
- JWKS URI 必须匹配 reviewed scheme / host / port policy；禁止任意 URL、file URI、内网扫描或跨 origin redirect。
- response 必须限制超时、body size、key count、JSON depth 和 content type；失败不回显 body。
- startup 必须完成 issuer / discovery / JWKS / algorithm preflight 后再 listen；缺失或 mismatch 拒绝启动。
- runtime refresh 使用 single-flight，遵守 reviewed cache max-age 上限；到达 hard expiry 后不无限使用 stale key。
- unknown `kid` 只允许触发一次受控 refresh；refresh 后仍不存在则 token invalid，不循环 fetch。
- key rotation 测试必须覆盖旧 key 在 reviewed overlap window 内有效、新 key 可用、hard expiry 后旧 key 拒绝。

## Token Validation Contract

### Header

- 只接受 Bearer scheme 和三段 compact JWT；不接受 query token、cookie、form token 或多个 Authorization value。
- `alg` 必须同时在本地 allowlist、reviewed discovery/JWKS evidence 和 key type 中匹配；禁止 `none`、HMAC、algorithm confusion。
- `kid` 必填并满足 reference 长度限制；不按数组第一个 key 或单 key 隐式选择。
- 不信任 token header 中的 `jku / jwk / x5u / x5c` 网络位置。

### Claims

- `iss` exact match；`aud` 必须包含 exact RadishMind integration audience，不能只接受 `azp` 或通用 Radish API audience。
- `sub`、tenant claim、permission claim、`iat / nbf / exp / auth_time` 必填，类型与 mapping version 一致。
- `exp > nbf`，token lifetime 不超过 reviewed 上限；clock skew 使用固定小窗口，不能覆盖明显过期或未来签发。
- subject / tenant / permission 只接受稳定 reference；display name、email、role name 不进入 authorization。
- permission 通过版本化 exact allowlist 投影为 `tenant:read / audit:read`；未知值忽略，不支持 wildcard、prefix 或角色推导。
- raw claims 在 validation 后立即丢弃；下游只接收 `VerifiedControlPlaneIdentity` 和 `ControlPlaneResourceBinding`。

### Mapping 所有权

Radish owner 拥有 upstream claim semantics；RadishMind owner 拥有 mapping version、route grant 与失败语义。任何 claim path、type、permission identifier 或 audience 变化都必须更新 reviewed mapping evidence 并重新执行 negative suite，不允许运行时自动兼容多个猜测 shape。

## 配置与环境组合

建议新增：

- `RADISHMIND_CONTROL_PLANE_READ_AUTH_MODE=radish_oidc_integration_test`
- `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_ISSUER`
- `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_AUDIENCE`
- `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_MAPPING_VERSION`
- `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_DISCOVERY_TIMEOUT`
- `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_JWKS_MAX_AGE`
- `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_CLOCK_SKEW`

不新增 client secret 配置。真实 token 由 test harness 通过 stdin / permission-restricted temp file 提供，不进入 Platform config summary。

| auth mode | store mode | 结论 |
| --- | --- | --- |
| `signed_test_token` | `postgres_dev_test` | 保留 deterministic regression |
| `radish_oidc_integration_test` | `postgres_dev_test` | 本功能唯一真实 issuer 组合 |
| `radish_oidc_integration_test` | `fake_store_dev` | startup reject |
| `dev_headers` | 任意 integration claim | reject / credential conflict |
| production / unknown | 任意 | startup reject |

配置摘要只暴露 auth mode、issuer configured、audience configured、mapping version、discovery/JWKS readiness 与 sanitized evidence ref；不输出 issuer URL query、JWKS body、token、claim、network error detail 或临时文件路径。

## Failure 与 No-fallback

| boundary | failure code | HTTP | repository query |
| --- | --- | --- | --- |
| missing credential | `identity_context_missing` | `401` | `0` |
| malformed scheme / JWT / header | `auth_context_contract_mismatch` | `401` | `0` |
| signature / issuer / audience / algorithm / time / claim invalid | `auth_context_contract_mismatch` | `401` | `0` |
| unknown `kid` after one refresh | `auth_context_contract_mismatch` | `401` | `0` |
| tenant claim missing | `tenant_binding_missing` | `401` | `0` |
| path / query tenant mismatch | `tenant_binding_missing` | `403` | `0` |
| permission missing | `scope_denied` | `403` | `0` |
| workspace operation without membership | `workspace_membership_unavailable` | `503` | `0` |
| discovery / JWKS unavailable before token decision | `identity_provider_unavailable` | `503` | `0` |
| reviewed issuer / mapping mismatch | `auth_policy_mismatch` | startup blocked | `0` |
| PostgreSQL unavailable after auth success | `read_store_unavailable` | `503` | attempted once |

任何 OIDC failure 都不得尝试 signed test public key、dev headers、last successful identity、fake tenant、default permission 或 workspace fake repository。JWKS bounded cache 不是 identity cache，不能跨 token 复用 verified subject / tenant / permission。

## 浏览器与 Token 隐私

- 本功能不实现 authorization code、PKCE、BFF、login / callback / logout、cookie 或 refresh token。
- 浏览器验收只允许 operator 在组件 password input 或 Playwright init context 中提供短期 token；token 只存在组件 / page memory。
- token 不进入 `import.meta.env`、构建产物、URL、hash、query、localStorage、sessionStorage、IndexedDB、service worker、draft、history 或 screenshot。
- consumer 不展示 Authorization header、token length、claim dump、JWKS 或 provider error。
- browser close、page reload、test end 和 failure cleanup 必须清除 token；测试结束后 token 在 upstream 过期或撤销。
- Playwright / CI 输出不得回显 fill value、request headers 或 token；只记录 sanitized assertions。

## 验收方式

### Deterministic Unit / Integration Issuer

- discovery exact issuer、JWKS origin、content type、size、timeout、redirect 和 malformed response。
- RS / ES allowlist、`kid` selection、unknown `kid` single refresh、rotation overlap / hard expiry、algorithm confusion。
- issuer / audience / lifetime / skew / required claim / type / cardinality / mapping version matrix。
- exact permission projection、unknown permission ignore、tenant mismatch、workspace membership unavailable。
- zero-query denial、no auth fallback、no workspace fake fallback、sanitized diagnostics。

### Reviewed Radish Integration Environment

- operator evidence gate 已完成且未超过 review window。
- 使用 Radish owner 签发的短期 integration token完成 tenant ready、audit ready / pagination。
- invalid audience、expired token、removed permission、tenant mismatch 和 rotated key negative smoke。
- discovery / JWKS temporary unavailable 映射 `503`，恢复后可重新启动或刷新，不使用无限 stale key。
- PostgreSQL 与 OIDC failure 边界可区分，request / audit correlation 保持一致。

### Web / Browser

- 默认 offline 零请求；signed test regression 保持通过。
- integration mode token 只在内存，缺 token 明确失败，不回退 dev headers。
- Tenant Overview / Audit Log ready、denied、identity provider unavailable 与 database unavailable 状态可区分。
- console 0 React / unhandled error；URL / localStorage / sessionStorage / IndexedDB / screenshot / build artifact 无 token。

### 仓库门禁

- Go unit / race / vet。
- controlled issuer integration 与可选 reviewed Radish integration runner。
- Web unit / production build。
- secret / token negative scan、`git diff --check`、fast 与完整 `check-repo`。

本功能新增真实网络信任边界，因此需要一个统一高风险 runtime task card 和专项 integration tests；不创建只验证文件存在的 readiness checker。

## 实施拆分

### 批次 A：Evidence 与 Verifier Core

- 创建 `Radish OIDC Integration Test Runtime v1` 高风险任务卡。
- 固定 reviewed evidence manifest shape 和 mapping version，不提交真实 token / JWKS dump。
- 实现 issuer/discovery/JWKS client、bounded cache、refresh single-flight、validator 与 deterministic issuer tests。

### 批次 B：Auth Boundary 与 Operation Gate

- 把 verified identity 接入现有 auth context，不复制 handler authorization 模型。
- 只开放 tenant / audit；五条 workspace operation 在 membership 前 fail closed。
- 完成 config/startup、failure mapping、zero-query denial、no-fallback 和 telemetry。

### 批次 C：Radish / HTTP / Browser Evidence

- 由 Radish owner 提供 reviewed integration environment 和短期 token issuance procedure。
- 执行真实 issuer、rotation、negative claim、PostgreSQL、HTTP/Web 和浏览器验收。
- 清理 credential 与服务后关闭任务卡；未取得 upstream evidence 时保持 `blocked_by_upstream_evidence`，不以本地 issuer 代替完成。

## 停止线

- 不启用 production auth mode、production issuer、production client registration 或 production token。
- 不实现 OAuth login、authorization code、PKCE、BFF、callback、logout、session cookie、refresh token 或长期 browser session。
- 不读取 Radish数据库，不复制 user / tenant / role / permission 真相表。
- 不实现 workspace / application membership adapter、cache、API 或五条 workspace route 的真实 OIDC enablement。
- 不新增 northbound API、Gateway schema、repository、provider registry、tenant / audit writer 或管理写入。
- 不打开 application promotion、production key、quota enforcement、billing、cost ledger、secret runtime、Workflow tool / confirmation / writeback / replay / resume。
- 不把 controlled issuer、Radish integration environment、PostgreSQL、HTTP 或浏览器成功解释为 production auth ready 或 production SLA。

## 下一实现入口

[Radish OIDC Integration Test Runtime v1 任务卡](../../task-cards/radish-oidc-integration-test-runtime-v1.md) 已完成 deterministic verifier、auth boundary、operation gate、zero-query 与 Web 内存 token 批次。真实 Radish integration 保持 `blocked_by_upstream_evidence`；下一步由 Radish owner 提供 reviewed metadata-only evidence 与短期 token 流程，证据未到位前不派生同层 readiness 文档链，也不把 loopback 测试解释为真实联调。
