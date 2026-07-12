# Admin Control Plane Authenticated Read Store Transition v1

更新时间：2026-07-12

状态：`admin_control_plane_authenticated_read_store_transition_v1_in_progress`

## 功能目标

让平台管理员从 RadishMind Admin Control Plane 进入经过身份验证、tenant 绑定和权限检查的只读管理路径，读取 durable、sanitized 的 tenant 与 audit summary，并能够区分认证失败、权限不足、数据源不可用、schema 漂移和真实空结果。

本功能负责从当前 dev header / fixture-backed fake store 有序迁移到 authenticated read path。它不把 Admin 页面升级为管理写入控制台，也不因为身份或 read repository 可用就打开 application promotion、API key、quota、billing、secret、部署或 Workflow 高风险能力。

## 当前事实与设计判断

- 七条 Control Plane read route、`ControlPlaneReadRepository` interface、typed repository context、sanitized envelope、filter / cursor allowlist 和稳定 failure code 已存在。
- 当前 HTTP runtime 只有 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 与 `X-RadishMind-Dev-Read-*` header，repository 默认仍是 fixture-backed fake store；没有正式 auth mode 或 read store selector。
- `contracts/radish-oidc-token-validation.schema.json` 已定义 verified token 的脱敏输出，但当前没有 token validator、OIDC middleware、membership adapter 或 authenticated route binding。
- Workflow 的 production auth runtime bridge 是 saved draft 专用边界，可复用 verified-context 原则和失败治理，不直接充当 Admin 共享身份层。
- Application Publish Governance 已把 production auth / membership 与正式 repository 暴露为 blocker；本功能只解决 Admin read transition，不顺带启用 promotion。

因此，后续不能把历史 readiness 文档继续派生成同层静态链，也不能一次同时实现 OIDC、membership、PostgreSQL repository 和浏览器登录。实现必须按共享身份边界、Admin read repository、真实 Radish 联调三个阶段推进。

## Radish 在线事实与未决依赖

2026-07-12 设计复核读取以下在线主分支资料：

- [Radish 在线仓库](https://github.com/laugh0608/Radish) 当前包含 `Radish.Auth`、`Radish.Api`、`Radish.Gateway` 与独立 repository / database 层。
- [Radish 鉴权与授权指南](https://github.com/laugh0608/Radish/blob/master/Docs/guide/authentication.md) 说明 `Radish.Auth` 基于 OpenIddict，提供 discovery、authorize、token、userinfo、logout、revoke 和 introspect 等标准端点，并要求资源访问同时满足 `aud=radish-api` 与 `scope=radish-api`。
- [Radish 身份语义收敛设计](https://github.com/laugh0608/Radish/blob/master/Docs/architecture/identity-claim-convergence.md) 将 `sub`、`tenant_id`、role 和 scope 归一为 `CurrentUser`，禁止业务运行时分散解析 claim。

这些资料证明 Radish 可以继续作为身份、tenant 与上游权限语义 owner，但本次复核没有得到可直接承诺给 RadishMind 的 workspace membership endpoint、RadishMind client registration 或 RadishMind permission mapping。以下三项在真实联调前必须形成独立、可复验的上游证据：

1. `radishmind-admin` 或等价 OIDC client registration、redirect / logout policy、audience 和 scope。
2. issuer / discovery / JWKS、算法 allowlist、轮换和不可用策略。
3. Radish role / permission 到 RadishMind `tenant:read`、`audit:read` 的显式映射；不能仅凭角色名字猜测授权。

## 目标用户流程

完整产品流程如下：

1. 管理员从 Admin Control Plane 进入受保护管理面；未建立 authenticated session 时进入 Radish OIDC Authorization Code + PKCE 或后续审查通过的 BFF 登录流。
2. RadishMind auth boundary 校验 issuer、audience、signature、algorithm、token time、required claims 和 policy version，只产生 sanitized verified identity context。
3. tenant / permission binding 根据 verified subject、tenant 和 reviewed permission mapping 生成 Admin read grant；请求参数不能覆盖 tenant identity。
4. route authorization 针对 `tenant:read` 或 `audit:read` fail closed，并把 request、tenant、subject、issuer、session、scope 和 audit reference 投影到 repository context。
5. read store 仅按 repository context 查询 sanitized tenant / audit summary；scope predicate 必须进入 query，不允许先全量读取再在 Web 过滤。
6. Web 分别展示 authenticated、denied、expired、store unavailable、schema mismatch、empty 和 ready；失败不回退 offline fixture 或 dev fake store。
7. 管理员可以复验 request / audit ref 和数据来源状态，但不能查看 token、claim dump、membership raw record、SQL / DSN 或 secret material。
8. logout / session expiry 清除当前 authenticated UI state；不把 token、tenant 或 Admin 数据写入 URL、`localStorage` 或 `sessionStorage`。

第一阶段真实页面验收只覆盖 Admin tenant summary 与 audit summary。Applications、API keys、quota、workflow definitions 和 runs 在 workspace selection / membership contract 成立后再迁移，不把 tenant-level Admin 权限代替 workspace authorization。

## 真相源与所有权

| 数据 / 决策 | owner | RadishMind 可做 | 禁止替代 |
| --- | --- | --- | --- |
| user、subject、tenant、role、上游 permission | Radish | 校验并消费脱敏引用 | 不复制用户、密码、角色或 tenant 真相源 |
| token signature / issuer / audience / time | RadishMind auth boundary，依据 reviewed Radish evidence | 输出 verified identity context | 不把 Web claim parsing 或 dev header 当验证 |
| Radish → RadishMind permission mapping | 版本化 policy，由双方契约审查 | 投影 `tenant:read` / `audit:read` | 不按角色名称或 UI 可见性隐式授权 |
| RadishMind workspace / application membership | future RadishMind resource membership provider | 绑定自身资源 reference 和 expiry | 不创建第二套用户 / tenant / role 系统 |
| tenant / audit / application 等控制面 summary | `ControlPlaneReadRepository` 对应正式 projection owner | scoped read 与 sanitized projection | 不以 fixture、Web state 或 Gateway History 充当真相源 |
| request / audit ref | Platform request / audit boundary | 透传脱敏关联信息 | 不保存 Authorization、raw claim 或 membership record |

## 共享 Verified Identity Context

后续实现应建立 Control Plane 共享类型，不复制 Saved Draft 私有 bridge。建议职责拆为：

- `VerifiedControlPlaneIdentity`：auth source、issuer ref、subject ref、tenant ref、audience refs、upstream roles / permissions refs、scope grants、issued / expiry / auth time、policy version、session ref、request / audit ref。
- `ControlPlaneResourceBinding`：tenant verified、Admin permission grants，以及后续 workspace / application binding；绑定结果必须带 source ref、policy version 与 expiry。
- `ControlPlaneReadAuthorization`：把 route operation 映射为 required grant，生成 `ReadRepositoryContext`；未知 route / operation fail closed。

`ReadRepositoryContext` 已有 request、tenant、subject、scope、audit、issuer 和 session 字段。实现时应补齐并校验这些字段；workspace-scoped route 后续需要新增内部 `WorkspaceID`，但本功能设计不改变当前 HTTP response schema。

只有 auth boundary 可以解析标准 claim。route、repository、Web consumer 和业务代码只消费归一化语义，不理解 `sub`、`tenant_id`、role 或 scope 的 raw claim 名。

## Tenant、Permission 与 Membership 边界

### Admin tenant / audit

- `GET /v1/control-plane/tenants/{tenant_ref}/summary` 要求 verified token tenant 与 path tenant 一致，并具备 `tenant:read`。
- `GET /v1/control-plane/audit` 的 tenant 只能来自 verified context，并具备 `audit:read`。
- upstream `radish-api` audience / scope 只证明 token 可访问 Radish resource domain，不自动等于 RadishMind Admin grant。
- permission mapping 必须是显式、版本化、可测试的 allowlist；缺失 mapping 返回 denied，不回退 `System` / `Admin` 名称判断。

### 后续 workspace routes

- application、API key、quota、workflow definition 和 run route 必须额外消费 active workspace binding。
- workspace selector 只是资源选择，不是授权证明；即使未来来自 header / session，也必须经过 membership provider 校验。
- 多 workspace 且没有唯一 active binding 时必须要求显式选择，不能默认为第一个 workspace 或 `workspace_demo`。
- 本设计不新增 workspace selector API；该 contract 在 Admin 两条 route 完成后单独设计。

## Auth / Store 模式与兼容矩阵

建议配置维度分离为 `CONTROL_PLANE_READ_AUTH_MODE` 与 `CONTROL_PLANE_READ_STORE`，但启动时按组合 fail closed：

| auth mode | store mode | 环境 | 结论 |
| --- | --- | --- | --- |
| `disabled` | none | 默认 offline Web | 零请求，不启动 Platform read consumer |
| `dev_headers` | `fake_store_dev` | 显式本地开发 | 保留现有路径；只允许 loopback / dev，不声明 authenticated |
| `signed_test_token` | `fake_store_dev` | 第一实现批次 | 验证 token / permission / failure 边界，不打开数据库 |
| `signed_test_token` | `postgres_dev_test` | 第二实现批次 | 验证 authenticated context → durable Admin read query |
| `radish_oidc` | `postgres_dev_test` | 显式集成测试 | 只有 reviewed issuer / client / permission evidence 成立后允许 |
| `radish_oidc` | `repository` | production reserved | 需要正式 repository owner、migration、secret、audit 与发布复核；当前 disabled |

production 环境必须拒绝 `dev_headers`、`signed_test_token`、`fake_store_dev` 和未迁移 `postgres_dev_test`。任何未知模式、非法组合、连接失败或 schema mismatch 都不得回退 fake store。

## Route 与 Repository 迁移矩阵

| route | 当前 source | 本专题第一 read-store 范围 | 后续前置 |
| --- | --- | --- | --- |
| tenant summary | fake store | PostgreSQL dev/test tenant projection | `tenant:read` permission mapping |
| audit summary list | fake store | PostgreSQL dev/test sanitized audit projection | `audit:read` mapping、append-only writer owner 后续独立评审 |
| applications | fake store | 暂不切换 | active workspace membership contract |
| API keys | fake store | 暂不切换 | workspace binding、真实 key summary owner；不含 key material |
| quota summary | fake store | 暂不切换 | workspace binding、quota policy owner；不执行 enforcement |
| workflow definitions | fake store | 暂不切换 | workspace / application binding |
| runs | fake store | 暂不切换 | workspace / application binding 与 durable run projection |

已有七条 route path、filter / cursor allowlist 与 JSON envelope 保持兼容。实现 authenticated transport 时需要把 failure envelope 与 HTTP status 对齐，并作为专项 task card 的协议变更审查：

- `401`：missing / malformed / invalid / expired identity；
- `403`：tenant mismatch 或 scope / permission denied；
- `400`：invalid filter / cursor；
- `503`：membership、read store、migration 或 schema unavailable；
- `500`：sanitized contract mismatch；
- `200`：ready 或真实 empty。

failure body 继续使用现有 `request_id / tenant_ref / items=[] / next_cursor=null / failure_code / audit_ref`，Web consumer 必须解析非 2xx sanitized envelope，不能压成泛化网络错误。

## Durable Read Repository 边界

- repository interface 继续是唯一 handler 数据入口，不在 handler 中写 SQL。
- PostgreSQL dev/test adapter 首批只实现 tenant summary 与 audit summary；未迁移 operation 必须由 selector 明确路由到允许的 dev fake store，不能在同一 operation 内隐式 fallback。
- schema、migration、marker、checksum、runtime read-only role、timeout 与 no-fallback 复用现有 PostgreSQL dev/test 工程模式，但使用独立 control-plane read schema 和 runner。
- tenant predicate、permission grant 与 projection allowlist 必须进入 repository contract test；runtime role 不具备 DDL 或写业务表权限。
- audit summary repository 只读取 sanitized projection。正式 append-only audit writer / retention / redaction owner 未成立前，PostgreSQL 数据只属于 dev/test fixture，不声明合规审计账本。
- repository 不读取 Radish 业务数据库，也不复制 Radish user / role / tenant 表；跨项目只通过 OIDC 和审查后的 membership / permission contract 集成。

## 失败语义与 No Fallback

沿用并收紧以下稳定 failure：

- `identity_context_missing`
- `auth_context_contract_mismatch`
- `tenant_binding_missing`
- `scope_denied`
- `invalid_filter`
- `read_store_unavailable`
- `read_store_contract_mismatch`
- `database_read_disabled`
- `schema_migration_not_applied`
- `schema_version_mismatch`

token signature、issuer、audience、algorithm、time 或 required claim 失败统一落在 auth boundary；公开响应不暴露详细校验原因，内部 sanitized diagnostic 可记录 boundary、policy version 和 case id。tenant / permission denial 不查询 repository。store / schema 失败不返回旧 snapshot，不回退 fixture，不把 empty 当 unavailable。

第一实现批次固定以下 13 类负向认证基线，后续 task card 不得合并成只验证通用 `401` 的单一用例：

1. 缺少认证凭据；
2. Authorization scheme 或 token 结构非法；
3. 使用 `none` 或 allowlist 外算法；
4. signature 无效或签名 key 不受信任；
5. issuer 不匹配；
6. audience 不匹配；
7. token 已过期；
8. token 尚未生效或签发时间异常；
9. required claim 缺失或类型不符合 contract；
10. tenant binding 缺失；
11. verified tenant 与 path tenant 不一致；
12. permission mapping 缺失或 required grant 被拒绝；
13. authenticated mode 中注入 dev header，或 Bearer 与 dev header 同时出现。

前 10 类公开返回 sanitized `401`，第 11、12 类返回 `403`；第 13 类必须拒绝 dev header 参与身份选择，并按无效认证组合返回 `401`。所有 13 类都要求 repository query count 为 0。

## 隐私与安全边界

- raw access token、refresh token、authorization code、Authorization header、cookie、client secret、raw claim / JWKS dump 和 membership raw record 不得进入日志、response、repository、URL 或浏览器 storage。
- Admin Web 首选 server-managed BFF / `HttpOnly + Secure + SameSite` session；若后续选择 SPA 直持 token，必须单独完成威胁评审，本设计不默认允许 `localStorage` / `sessionStorage` token persistence。
- repository 与 response 只允许 sanitized summary；API key value / hash、secret value / ref payload、DSN、SQL、provider endpoint、raw audit payload 和业务正文禁止返回。
- dev header middleware 在 authenticated mode 中必须完全禁用；同时出现 Bearer 与 dev header 时不得选择 dev header。
- auth / membership cache 只缓存脱敏 reference、grant 和 expiry，按 tenant + subject + policy version 隔离；deny / unavailable 不长期缓存为 allow。

## 观测与审计

每次 authenticated read 至少产生：

- request id、audit ref、route id、auth mode、store mode；
- issuer ref、subject ref、tenant ref 的脱敏引用；
- token validation policy、permission mapping policy、repository schema version；
- decision：allowed / denied / unavailable；
- failure boundary、duration 和 repository query count。

禁止记录 token、raw claim、membership record、query parameter value 全量、SQL、DSN 或数据库错误正文。auth denial 的 repository query count 必须为 0；成功的单 route query 数应受 contract test 约束。

## 实施顺序

### 第一批：Verified Identity Context & Negative Auth Runtime

- 创建高风险 task card，建立共享 verified identity / resource binding / route authorization 类型。
- 使用静态签名测试 key 或 in-process test issuer 验证 schema、issuer / audience / signature / time / claim、permission mapping 和 13 类负向 auth 场景。
- 将现有七条 route 的 authorization 统一接入共享边界，但 repository 仍为 fake store；Admin tenant / audit 完成 signed test token HTTP smoke。
- 更新 Web consumer 解析 401 / 403 sanitized failure；默认 offline 和 dev header 路径继续可复验。
- 不连接 Radish、不 fetch 真实 JWKS、不创建 database adapter。

完成状态：`control_plane_verified_identity_negative_auth_runtime_v1_complete`。现有 runtime 已支持 `disabled / dev_headers / signed_test_token`、RS256 test verifier、版本化 permission projection、13 类负向认证、七条 route authorization、sanitized 非 2xx envelope 和 Web denial state；repository 继续使用 fake store。

### 第二批：Admin Tenant / Audit PostgreSQL Read Repository

- 创建独立 task card、schema / migration / marker / runner、read-only runtime role 和 selector。
- 只迁移 tenant summary 与 audit summary operation，完成 scope predicate、分页、restart、rollback / reapply、schema mismatch 和 no-fallback。
- 使用 `signed_test_token + postgres_dev_test` 完成真实浏览器 authenticated / denied / empty / store failure 路径。
- 不启用 production repository，不创建 audit writer 或管理写入。

实施状态：[`admin_tenant_audit_postgresql_read_repository_v1_complete`](tenant-audit-postgresql-read-repository-v1.md)。projection schema / ownership、manual migration、read-only role、routed selector、strict cursor / filter、failure 和 no-fallback runtime 已通过真实 PostgreSQL、HTTP/Web 与浏览器验收。

### 第三批：Radish OIDC Integration Test

- 已创建统一高风险任务卡，并完成 deterministic discovery / JWKS、JWT validation、tenant permission binding、两条 Admin route、key rotation / JWKS unavailable、跨 tenant、permission denied、zero-query 与 no dev fallback。
- 五条 workspace operation 在 membership owner 缺失时统一返回 `workspace_membership_unavailable`，不读取 fake repository。
- 当前 reviewed issuer / audience / JWKS policy / claim mapping / permission identifier 仍未落地，真实 Radish HTTP/Web/浏览器联调主动 deferred，不再阻塞当前产品开发。
- 本批不实现 authorization code、PKCE、BFF、session cookie、refresh token 或 logout；未来 production auth 必须另行设计和验收，不从集成测试自动晋级。

### 第四批：Workspace-scoped Read Transition

另行设计 active workspace selection、membership provider 和五条 User Workspace route 的 repository transition。它不属于本专题前三批的默认实施范围。

## 验收方式

每批必须分别提供单元、HTTP、repository / migration、浏览器和安全证据：

- auth：positive verified context、13 类负向 case、tenant / permission mapping、zero repository query on denial、dev header isolation、session expiry；
- store：migration / rollback / reapply、read-only role、restart recovery、scope predicate、pagination、empty / unavailable 区分、schema mismatch 和 no fallback；
- Web：offline 零请求、authenticated / expired / denied / unavailable / empty / ready、logout 清理、URL / storage / console 泄漏检查；
- integration：reviewed Radish issuer / client / permission evidence、JWKS rotation / unavailable、跨 tenant denial 和完整 request / audit correlation；
- repository：Go test / race / vet、Web test / build、PostgreSQL integration、`git diff --check`、fast 与完整 `check-repo`。

本设计本身只修改文档，复用文档语言、Markdown size、链接、fast/full 仓库门禁；不新增 checker。首个实现 task card 因涉及 token validation、HTTP status 和 auth boundary，必须补专项测试，不以历史静态 readiness checker 替代 runtime evidence。

## 停止线

- 不在本设计批次实现 OIDC client、token validator、auth middleware、membership adapter、BFF、session store、PostgreSQL adapter 或 production route。
- 不自建用户、密码、tenant、role、permission 或 OIDC 真相源，不读取 Radish 业务数据库。
- 不把 `radish-api` scope、`System` / `Admin` 角色名或前端页面可见性直接解释为 RadishMind Admin grant。
- 不在 verified identity / permission mapping 未成立时启用 repository query。
- 不允许 production 使用 dev header、signed test token、fake store 或 `postgres_dev_test`。
- 不新增管理写入、tenant / user / role mutation、raw audit export、application promotion、API key lifecycle、quota enforcement、billing、cost ledger、secret runtime 或 deployment apply。
- 不扩展 Gateway protocol、provider registry、fallback、load balancing，也不扩展 Workflow tool、confirmation、writeback、replay 或 resume。
- 不把设计完成、signed test token、PostgreSQL dev/test、真实 Radish 联调或两条 Admin route 成功解释为 production Admin ready。

## 下一实现入口

`Admin Tenant / Audit PostgreSQL Read Repository Runtime v1` 已完成真实 PostgreSQL、signed-token HTTP/Web 和浏览器验收。[Radish OIDC Integration Test v1](radish-oidc-integration-test-v1.md) 已完成 deterministic runtime、auth boundary 与 operation gate。真实 Radish 联调为 `real_radish_integration_deferred`，不再是当前下一步；未来由 Radish 注册 RadishMind application/client 与 resource audience 后恢复，不迁移五条 workspace-scoped route。
