# 本地账户与 Radish OIDC 联合登录 v1 高风险任务卡

更新时间：2026-08-19

状态：`local_account_radish_oidc_federated_login_v1_batch_a_completed_batch_b_ready`

对应功能设计：[RadishMind 本地账户与 Radish OIDC 联合登录 v1](../features/admin-control-plane/local-account-radish-oidc-federated-login-v1.md)

## 任务目标

在不复制 Radish 身份数据库、不信任 email 自动合并、不把 OIDC claim 当作本地授权的前提下，为 RadishMind 建立可复验的本地账户、凭证、外部身份、Web Session、角色与工作区成员关系 owner，并实现本地登录和 Radish OIDC Authorization Code + PKCE 的统一本地会话链。

## 已确认前提

- RadishMind 拥有本地注册、用户、角色、权限、session 和 workspace membership。
- RadishMind 是 Radish OIDC client / Relying Party，不在本任务中成为 issuer。
- `(issuer, subject)` 是外部身份唯一键；email 只作受控 profile / 联系字段。
- 业务 handler 只消费本地 session actor context 和本地 membership decision。
- 既有 deterministic OIDC verifier 是 resource-server 基础，不等于 browser login。

## 批次 A：领域契约与仓储

实施范围：

- 新增清晰的 identity / authentication / authorization 领域文件，不把账户逻辑塞进 `control_plane_read_auth.go`。
- 定义 `UserAccount`、`LocalCredential`、`ExternalIdentityBinding`、`WebSession`、`LocalRoleAssignment`、`WorkspaceMembership` 与 repository interfaces。
- 固定 stable id、状态机、版本、时间、唯一约束、事务边界、failure taxonomy 和 sanitized projection。
- 建立 memory、聚合 SQLite 和显式 PostgreSQL dev/test migration / repository；三种实现保持语义一致且不 fallback。
- 使用标准密码派生能力，参数与版本显式；原始密码不进入领域对象、日志、错误、audit 或 fixture。
- 建立本地 membership adapter，并验证所有既有 workspace operation 在 identity / membership failure 时保持零业务副作用。

批次 A 验收：

- 同一规范化本地登录标识不能创建两个 active 账户。
- 同一 `(issuer, subject)` 不能绑定两个本地账户；不同 issuer 下相同 subject 必须隔离。
- 账户、credential、binding、session 与 membership 的并发更新具有明确单胜者或幂等语义。
- session revoke / expire、account disable 和 membership revoke 在下一次请求立即失败关闭。
- memory / SQLite / PostgreSQL 覆盖创建、读取、版本冲突、唯一冲突、重启、migration、rollback、运行角色和 no-fallback。

## 批次 B：本地注册、登录与 Session HTTP

实施范围：

- 注册、登录、当前 session、logout、session revoke API。
- 安全 cookie、CSRF / origin policy、受限 return target、账户枚举防护、稳定错误和脱敏审计。
- 统一 middleware 从 local session 恢复 actor context，再读取 local membership；不直接把 cookie 或 credential 传给业务 handler。
- dev header / signed-test token 继续受显式 gate 约束，不能成为本地 session 失败后的 fallback。

批次 B 验收：

- 正向注册 / 登录 / 刷新 / 登出 / 重启恢复。
- 重复注册、错误凭证、禁用账户、过期 / 撤销 session、CSRF / origin / return target 负向矩阵。
- cookie、密码和 session credential 不进入 JSON 响应、URL、日志、audit 或浏览器持久存储。

## 批次 C：确定性 OIDC Client

实施范围：

- OIDC login start、Authorization Code + PKCE callback、external identity resolve / create / link。
- exact issuer / client / redirect / scope、一次性 state / nonce / code verifier、ID token 签名 / audience / time 校验。
- OIDC 网络读取复用既有 bounded HTTP / JWKS primitives；client policy、ID token validation 和本地 session issuance 保持独立职责。
- loopback deterministic issuer 覆盖 rotation、provider unavailable、code replay、state / nonce / PKCE mismatch 和 token 隐私。

批次 C 验收：

- 已绑定身份登录、准入允许时首次创建、准入拒绝、显式绑定、绑定冲突和最后登录方式解绑拒绝。
- 相同 email 不自动合并；上游 tenant / role / permission claim 不直接形成 local grant。
- callback 失败不创建半账户、半 binding、半 session，也不回退本地管理员。

## 批次 D：Web 产品面

实施范围：

- 登录 / 注册、Radish 登录、账号绑定与当前会话管理。
- User / Role 页面切换到本地真实 owner，保留 loading / empty / denied / unavailable / conflict 状态。
- 先按五维评分选择完整 Pencil、局部 Pencil 或直接实现，再完成 React strict consumer 和响应式实现。

批次 D 验收：

- 桌面与窄屏覆盖本地登录、OIDC 登录、刷新、跨标签登出、绑定 / 解绑、账户禁用和恢复指引。
- URL、Web Storage、IndexedDB、service worker、截图与构建产物不含身份 credential。
- 控制台无 React warning、未处理异常或敏感 provider error。

## 批次 E：真实 Radish 联调

进入条件：

- reviewed client registration、issuer、redirect URI、scope、signing policy 和测试账户流程可用。
- client credential 通过正式 secret ref 注入；不得提交真实 secret 或 token。
- 本地 account / binding / session / membership owner 已通过批次 A 至 D。

验收只声明受控 integration environment 登录链成立；不自动声明 production auth、生产 session store、生产审计、速率限制、MFA、账号恢复或合规完成。

## 必须保持的负向边界

- 无 email 自动绑定、无 wildcard issuer / audience、无浏览器 token persistence。
- 无 raw token / claim / cookie / password / client secret 日志或 committed fixture。
- 无 OIDC role / permission 到本地 authorization 的隐式提升。
- 无 identity、session、membership 或 repository failure fallback。
- 无 Radish 数据库读取、身份表复制或本专题外业务写回。

## 验证入口

每个批次先跑新增领域和 HTTP 精准测试，再按风险扩展：

```bash
(cd services/platform && go test ./internal/httpapi/...)
npm --prefix apps/radishmind-web test
npm --prefix apps/radishmind-web run build
./scripts/check-repo.sh --fast
```

涉及正式文档真相源、身份协议、schema / migration、生产声明或真实 Radish 联调的批次，在提交前补跑完整：

```bash
./scripts/check-repo.sh
```

## 完成条件

- [x] 批次 A：领域契约与三种开发测试仓储完成。
- [x] 批次 B：本地注册、登录和 Web Session 完成。
- [ ] 批次 C：确定性 OIDC Relying Party 完成。
- [ ] 批次 D：Web、浏览器、隐私和响应式验收完成。
- [ ] 批次 E：真实 Radish integration evidence 完成，或明确保留为外部阻塞且不影响 A 至 D 的开发测试态关闭。
- [ ] 功能专题、当前焦点、路线图、契约、周志和停止线同步。
- [ ] 精准测试、fast gate 和最终完整门禁按风险通过。
