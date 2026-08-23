# RadishMind 本地账户与 Radish OIDC 联合登录 v1

更新时间：2026-08-23

状态：`local_account_radish_oidc_federated_login_v1_batch_d_completed_batch_e_external_blocked`

## 功能定位

本功能为 RadishMind 建立自己的注册、用户、登录、会话、角色和工作区成员关系体系，同时允许 RadishMind 作为 OIDC Relying Party 接入 `Radish`。本地账户是 RadishMind 平台身份与授权的唯一内部 owner；Radish OIDC 只提供一种外部认证方式，不接管本地角色、工作区成员关系或业务资源授权。

本专题不同于已完成的 [Radish OIDC Integration Test v1](radish-oidc-integration-test-v1.md)：既有专题验证 RadishMind 作为 resource server 校验 Bearer token 的 discovery / JWKS / JWT 边界；本专题实现面向用户的本地注册、登录、Authorization Code + PKCE、callback、外部身份绑定和 RadishMind Web Session。两条路径可以复用受限网络和签名校验基础，但不能共用 token audience、session、claim authority 或 permission projection。

## 已确认的产品决策

- RadishMind 拥有本地 `UserAccount`、`LocalCredential`、`WebSession`、角色、权限与 `WorkspaceMembership`。
- RadishMind 可以注册为 Radish OIDC client；Radish 继续拥有自己的 issuer、Radish 用户与 Radish 业务授权事实。
- 外部身份使用唯一 `(issuer, subject)` 映射到本地 `user_id`；同一外部身份不能绑定多个本地账户。
- email、display name、preferred username 和上游 role 名称都不是稳定身份键，不能单独触发账户合并或本地授权。
- 本地登录与 OIDC 登录最终都只签发 RadishMind 自己的 Web Session；业务 handler 不直接消费浏览器中的 access token 或 ID token。
- OIDC 首次登录只有在本地开放注册、有效邀请或管理员准入策略允许时才能创建本地账户；否则失败关闭。
- 已存在的本地账户只能在已验证的本地登录态中显式绑定外部身份，不按相同 email 自动绑定。
- 当前范围不让 RadishMind 成为 OIDC issuer，也不读取、复制或同步 Radish 的身份数据库。

## 当前代码事实

- Platform 的 deterministic discovery / JWKS / JWT primitive 现在分别服务 Admin resource-server operation 与独立 browser OIDC Relying Party policy；两者不共用 audience、permission projection 或 session issuance。
- `POST /v1/auth/oidc/start` 与 `GET /v1/auth/oidc/callback` 已实现 Authorization Code + PKCE、一次性 `state / nonce / code verifier`、exact client policy、ID token 验证、外部身份 resolve / create / link 和 RadishMind Web Session；refresh token 仍不进入当前范围。
- Platform 已落地本地账户、密码凭证、外部身份绑定、Web Session、角色分配、workspace membership 与 OIDC authorization transaction 的领域契约，以及 memory / 聚合 SQLite / 显式 PostgreSQL dev/test repository；PostgreSQL 保持 manual migration 和受限 runtime role，v1 → v2 upgrade、重启与 rollback 已复验，三种实现都不 fallback。
- 本地注册、登录、当前 session、logout / revoke 与 browser OIDC route 均接入独立 `local_session_dev_test` 模式；middleware 只从 Web Session cookie 恢复 `user:<user_id>`，再由本地 `WorkspaceMembershipProvider` 重读 membership / role。dev header、signed-test token、resource-server Bearer token 与本地 session 互斥，不形成失败 fallback。
- `GET /v1/auth/account` 已通过唯一 repository projection 返回当前会话的本地账户、外部身份、角色分配、workspace membership 与受控 capability；memory、SQLite、PostgreSQL 均在同一读取事务内形成确定性排序，不返回登录标识、issuer、subject、credential、token、raw claim 或 audit ref。
- `POST /v1/auth/external-identities/{binding_id}/revoke` 已实现同源 Origin、CSRF、近期认证、当前账户 ownership、record version CAS 与最后登录方式保护；不存在或不属于当前账户的 binding 不泄露其事实。
- Web 仅在 `VITE_RADISHMIND_LOCAL_IDENTITY_MODE=local_identity_dev` 显式启用本地身份 gateway；consumer 使用 credentialed cookie transport、严格 exact response shape 与敏感字段拒绝，不写 Web Storage。S7 User / Role 已改为只消费当前本地账户 owner；没有 assignment / membership 时显示真实空状态，不构造目录、角色或成员 fixture。

## 批次 A 实现状态

- `local_identity_domain.go` 固定六类 owner、稳定 ID、状态机、版本、时间、失败分类、受限登录标识与 exact issuer / subject 规范化；`user_id -> user:<user_id>` 是唯一 actor projection。
- 本地密码派生使用 Go 标准库 `crypto/pbkdf2` 的 PBKDF2-HMAC-SHA-256，策略版本、迭代数、salt 和 key length 显式；salt、派生结果和 session credential digest 均不进入 JSON projection。
- `CreateAccount` 原子写入账户与首个 credential；credential replacement 使用当前 `credential_id + record_version` 双条件 CAS，账户禁用与全 session revoke、external binding、session、role 和 membership 均采用身份键、版本条件与唯一约束。
- 聚合 SQLite migration 已进入 shared local persistence runtime；PostgreSQL 提供独立 manual migration、sanitized runner、运行角色、重启和 rollback integration test，不在服务启动时自动迁移。
- 本地 membership adapter 每次重新读取 active account、membership 与 role grants；account disable、session revoke / expire、membership revoke 和 permission denial 都失败关闭，不读取 OIDC claims 作为本地 grant。

## Owner 模型

| 领域对象 | owner | 稳定身份 / 作用域 | 核心边界 |
| --- | --- | --- | --- |
| `UserAccount` | RadishMind Identity | `user_id` | 保存本地账户状态与受控 profile；不保存 OIDC token |
| `LocalCredential` | RadishMind Credential | `credential_id + user_id` | 只保存版本化密码派生结果和状态；原始密码只存在单次请求内存 |
| `ExternalIdentityBinding` | RadishMind Identity | `issuer + subject` 唯一 | 映射到一个 `user_id`；禁止 email 自动合并 |
| `WebSession` | RadishMind Session | `session_id + user_id` | 浏览器只持有不可读或随机 session credential；服务端保存失效和过期事实 |
| `Role / Permission` | RadishMind Authorization | tenant / workspace scope | 本地授权真相；上游 claims 只有经过显式 mapping 才能成为非权威参考 |
| `WorkspaceMembership` | RadishMind Authorization | `tenant_ref + workspace_id + user_id` | 唯一 workspace 访问 owner；active workspace 不是 membership proof |
| `OIDCProviderRegistration` | RadishMind Integration Config | environment + issuer + client ref | 只保存 reviewed metadata 和 secret ref；不保存 raw client secret 或 token |

`actor_ref`、`owner_subject_ref` 等既有业务引用在身份接入后统一由本地 `user_id` 派生稳定 subject reference。业务 repository 不直接以 email、OIDC `sub`、Radish role 或 raw claim 作为 owner key。

## 用户流程

### 本地注册与登录

1. 用户提交受限注册标识与密码，服务端执行格式、重复账户、准入和速率策略检查。
2. credential owner 使用固定 allowlist 的密码派生算法创建版本化结果；日志、错误和审计不记录原始密码。
3. 注册成功后创建或激活本地 `UserAccount`，再由 Session owner 签发新会话；注册和会话写入不能出现半成功状态。
4. 后续登录只根据本地稳定标识定位账户，校验凭证、账户状态和登录策略后签发新会话。
5. 登出、账户禁用、凭证重置和管理员撤销会使相关会话失效；业务请求不能回退 dev header 或旧身份缓存。

### Radish OIDC 登录

1. 浏览器请求 OIDC login start，服务端生成一次性 `state`、`nonce`、PKCE verifier / challenge 和受控 return target。
2. authorization request 只使用 reviewed issuer、client id、redirect URI、scope 和 response type；浏览器不能覆盖 issuer、client、redirect URI 或 scope。
3. callback 先消费单次 `state`，再由服务端使用 authorization code 与 PKCE verifier 换取 token；失败不得重放或回退本地管理员。
4. 服务端校验 exact issuer、client audience、签名、算法、`exp / iat`、`nonce` 和授权流绑定，再立即丢弃 raw token / claim envelope。
5. `(issuer, subject)` 已绑定时，读取对应本地账户并检查其状态；未绑定时，仅按本地注册 / 邀请 / 管理员准入策略创建新账户和 binding。
6. 认证成功后只签发 RadishMind Web Session；workspace membership 和 permission 由本地 authorization owner 重新读取。

### 显式绑定与解绑

- 绑定必须从已验证且近期认证的本地会话发起，并重新执行完整 OIDC flow。
- callback 只能绑定到发起会话中的 exact `user_id`；目标外部身份已被其它账户占用时失败关闭。
- 解绑前必须保证账户仍有至少一种可用登录方式；不能把活跃账户变成无法恢复的孤立账户。
- 绑定、冲突、解绑、失败和管理员处置只写脱敏 audit ref，不记录 raw token、code、cookie、email 或 claim dump。

## 会话与浏览器安全边界

- Web Session 使用 `HttpOnly`、`Secure`、受控 `SameSite` cookie；session credential 不进入 URL、Web Storage、IndexedDB、service worker、React state dump 或截图。
- 登录 / callback / logout mutation 必须有明确 CSRF 与 origin policy；OIDC `state` 不能替代登录后的通用 CSRF 防护。
- 会话记录至少包含本地 `user_id`、认证方式、创建 / 最近验证 / 过期时间、失效状态、策略版本和脱敏 audit ref，不保存 ID token、access token、refresh token 或 authorization code。
- refresh token 默认不进入首批；若真实 Radish contract 必须使用，需单独评审加密存储、rotation、revocation、logout 和泄漏响应，不从当前设计自动打开。
- 身份提供方不可用、会话 store 不可用或本地 membership 不可用时失败关闭，不回退到 dev header、signed-test token、静态管理员或上次成功身份。

## 授权流程

一次业务请求按以下顺序决策：

1. Session owner 校验 Web Session 并恢复本地 `user_id`。
2. Identity owner 确认账户仍为 active。
3. Authorization owner读取 tenant / workspace membership 和 exact permission。
4. 资源 handler 检查 path / query / payload 与已验证 binding 一致。
5. 业务 repository 只消费脱敏 actor context；任一失败都在业务写入、Gateway、Provider 或 Tool 副作用前结束。

Radish OIDC claims 不直接注入 `WorkspaceMembershipProvider`。未来若需要从 Radish 同步成员关系，必须建立独立同步专题，覆盖 source ownership、冲突策略、撤销时效、审计与本地覆盖规则。

## 稳定失败分类

- `registration_disabled`
- `registration_admission_denied`
- `account_identifier_conflict`
- `local_credential_invalid`
- `account_inactive`
- `session_missing`
- `session_invalid`
- `session_expired`
- `oidc_provider_unavailable`
- `oidc_authorization_state_invalid`
- `oidc_callback_contract_mismatch`
- `oidc_token_exchange_failed`
- `oidc_identity_contract_mismatch`
- `external_identity_unbound`
- `external_identity_conflict`
- `account_link_requires_recent_authentication`
- `last_login_method_removal_denied`
- `workspace_membership_denied`
- `workspace_permission_denied`

公开错误只说明稳定 failure code、可执行恢复方向和 request / audit ref；不回显账户是否存在、密码原因、raw provider error、token、claim、code、cookie 或 client secret。

## 实施拆分

### 批次 A：领域契约与开发测试仓储

- 已完成账户、凭证、外部身份、会话、角色与 membership contract。
- 已完成 repository interface、memory / SQLite / PostgreSQL 语义、migration、唯一约束和 no-fallback。
- 已完成本地 `user_id -> actor_ref` 投影及 `WorkspaceMembershipProvider` 的本地 owner adapter。
- 已完成密码派生、敏感字段禁入、并发注册、绑定唯一性、会话撤销和 repository negative tests。

### 批次 B：本地注册、登录与会话

- 已实现 `POST /v1/auth/local/register`、`POST /v1/auth/local/login`、`GET /v1/auth/session`、`POST /v1/auth/logout` 与 `POST /v1/auth/sessions/{session_id}/revoke`；注册使用 `CreateAccountAndWebSession` 在 memory / SQLite / PostgreSQL 中原子写入账户、首个 credential 与 session，避免半成功。
- session credential 只进入 `HttpOnly`、`SameSite=Strict` cookie，HTTPS 使用 `Secure + __Host-`；显式 loopback HTTP 开发测试配置使用带 `_dev` 后缀的非 Secure cookie，不能解释为生产策略。CSRF 采用同源 Origin + session 派生双提交 cookie/header；注册和登录使用同源 bootstrap header。
- return target 只允许本地绝对路径；登录未知账户、错误密码、禁用账户统一返回相同认证失败；响应、错误与 audit ref 不包含登录标识、密码或 session credential。
- `local_session_dev_test` 与 `dev_headers`、`signed_test_token`、resource-server OIDC Bearer token 互斥。session、account、membership 或 repository 失败均失败关闭；业务 handler 只得到 actor / local binding，不得到 cookie 或 credential。
- 开发测试启用必须显式配置 `RADISHMIND_LOCAL_IDENTITY_DEV_HTTP=true`、`RADISHMIND_CONTROL_PLANE_READ_AUTH_MODE=local_session_dev_test`、精确 `RADISHMIND_LOCAL_IDENTITY_ALLOWED_ORIGIN`、cookie 策略和 session TTL；store 可选择 `memory_dev`、聚合 `sqlite_dev` 或显式 `postgres_dev_test`。当前不声明生产速率限制、MFA、账户恢复或 production session store。

### 批次 C：确定性 OIDC Relying Party

- 已实现 Authorization Code + PKCE、单次消费 `state`、摘要化 `nonce`、持久化一次性 code verifier、callback 与 `(issuer, subject)` resolve / create / link；畸形 callback、过期与重放均在 token endpoint 前失败关闭。
- exact issuer / client id / redirect URI / scope / algorithm 与 discovery / JWKS origin 由显式开发测试配置固定；独立 ID token policy 校验 issuer、audience / `azp`、签名、算法、`iat / nbf / exp` 和 nonce，不消费上游 email、role、permission 形成账户合并或本地 grant。
- 首次准入允许时以 `CreateOIDCAccountAndWebSession` 原子写入账户、binding 与 session；已绑定登录只签发新的本地 session。显式 link 同时绑定发起会话的 `session_id + record_version + user_id`，开始与 callback 均要求最近 10 分钟内认证。
- memory / SQLite / PostgreSQL 均提供 authorization transaction 单胜者语义与最后登录方式解绑保护；PostgreSQL `local_identity_records_store_v2` 支持 v1 升级，集成测试覆盖受限运行角色、重启、rollback / reapply 与 no-fallback。
- 仓库自有 loopback issuer 已覆盖首登、已绑定登录、准入拒绝、显式绑定、冲突、issuer / audience / `azp` / algorithm / signature / time、nonce / PKCE、rotation、provider unavailable、code / state replay、畸形 callback、两个 callback 同时观察未绑定时的原子注册单胜者和隐私；这些证据不冒充真实 Radish 联调。

### 批次 D：Web 与管理面

- 五维评分 `2 / 1 / 2 / 2 / 2 = 9`，采用 `A / 完整 Pencil`。正确的 `docs/designs/radishmind-web-family-ui-v1.pen` 第二排新增 Desktop `scHoA`、Narrow `uR4Yd` 与 Decision `SQPBB` 三块代表板；旧节点未修改，布局问题检查通过。
- 已提供本地注册 / 登录、Radish 登录入口、显式绑定 / 解绑、当前会话、失败与恢复状态；本地密码和 OIDC 最终只恢复 RadishMind session，浏览器不保存 token 或 session credential。
- User / Role 页面已切换为当前账户的本地 identity / role owner。页面只呈现 `GET /v1/auth/account` 的真实当前账户投影；缺少 assignment / membership 时保持空状态，离线或服务失败时保持 blocked，不提供假目录或管理 mutation。
- 真实浏览器完成错误密码、正确登录、当前账户面板、S7 User / Role、本地空状态、双标签登出传播、刷新恢复、SQLite 服务重启恢复及 `1440×900`、`720×900`、`390×844` 响应式检查；三标签 console 均无 warning / error。确定性 OIDC login / link 与 revoke 继续由批次 C loopback issuer、HTTP 和 strict consumer 自动化覆盖，真实 Radish 浏览器证据仍属于批次 E。

### 批次 E：真实 Radish 集成

- 等待 Radish 提供 reviewed client registration、exact issuer、redirect URI、scope、logout 和测试账户流程。
- 只在受控 integration environment 执行真实浏览器联调；credential 和 token 不进入仓库、命令参数、日志或截图。
- 真实登录成功不自动声明 production auth ready；仍需部署、secret、session store、审计、速率限制和安全复核。

## 验收方式

- 本地注册、重复标识、密码失败、账户禁用、会话过期 / 撤销、并发注册和 credential upgrade。
- `(issuer, subject)` 唯一绑定、不同 issuer 相同 subject 隔离、email 相同不合并、绑定冲突、最后登录方式解绑拒绝。
- OIDC state / nonce / PKCE、issuer / audience / algorithm / signature / time、code replay、callback origin 和 provider unavailable。
- membership / permission 在业务 repository 前失败关闭；Gateway、Provider、Tool 和业务写入副作用为 0。
- memory / SQLite / PostgreSQL 的唯一约束、事务、重启、并发、migration / rollback、运行角色和 no-fallback。
- Web 刷新、重启、跨标签登出、返回目标、失败恢复、三视口和控制台零未处理错误。
- URL、浏览器存储、日志、错误、audit、fixture、构建产物和截图不含密码、token、code、cookie、client secret 或 raw claim。

## 停止线

- 不按 email、用户名、display name 或 role 自动合并账户或授予权限。
- 不让 OIDC token、Radish claim 或 active workspace 直接替代本地 membership decision。
- 不在浏览器持久化 ID token、access token、refresh token 或 session credential。
- 不实现任意 issuer 动态接入、用户提供 redirect URI、wildcard audience、implicit flow 或 Resource Owner Password flow。
- 不让 RadishMind 在本专题成为 OIDC issuer，不同步 Radish 身份数据库，也不覆盖 Radish 自身业务权限。
- 不把 loopback issuer、开发测试数据库或真实 Radish integration smoke 解释为 production auth、production SLA 或合规认证完成。

## 下一实现入口

[本地账户与 Radish OIDC 联合登录 v1 高风险任务卡](../../task-cards/local-account-radish-oidc-federated-login-v1-plan.md)承接批次 A 至 E。批次 A 至 D 的开发测试态身份、Web 与管理面已经完成；批次 E 仍等待 reviewed Radish client registration、issuer、redirect、scope、signing policy、secret 注入与测试账户流程。外部条件成立前不继续用 loopback 证据冒充真实联调，也不从本专题派生 production auth、refresh token、MFA、恢复或速率限制。
