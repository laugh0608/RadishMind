# 本地账户凭证轮换与自助会话治理 v1 高风险任务卡

更新时间：2026-08-26

状态：`local_account_credential_rotation_self_service_session_governance_dev_test_v1_batch_d_react_completed_batch_e_ready`

对应功能设计：[本地账户凭证轮换与自助会话治理（开发 / 测试态）v1](../features/admin-control-plane/local-account-credential-rotation-self-service-session-governance-dev-test-v1.md)

## 任务目标

在唯一 `localIdentityRepository` 内建立当前账户的 session directory、exact session revoke、revoke others 与本地 credential rotation 原子链，再依次接入双数据库、strict HTTP、Authentication / Account Web 和真实产品连续验证。实现不得复制身份 owner，也不得把开发测试态会话治理解释为 production auth。

## 已确认前提

- 联合登录批次 A 至 D、本地成员管理批次 A 至 E 均已完成；真实 Radish 批次 E 仍为外部阻塞。
- `UserAccount`、`LocalCredential`、`WebSession` 与 `ExternalIdentityBinding` 继续由现有 local identity repository 单一拥有。
- 批次 C 前的 `POST /v1/auth/sessions/{session_id}/revoke` 只允许撤销当前 session；任务启动时 `ReplaceCredential` 与 `RevokeWebSession` 也是彼此独立的 repository 原语，不能直接证明新产品语义已经成立。
- 本任务只开放当前账户 self-service，不新增管理员账户禁用、代重置、全局 session 搜索、MFA、恢复或生产能力。
- 项目所有者已于 2026-08-25 批准功能专题的 owner、原子语义、路由提议、Pencil 覆盖级别和停止线，并授权批次 A 进入代码；批准不提前开放批次 B 以后的 HTTP、数据库、Pencil 或 Web 实施范围。

## 批次 A：领域合同与 memory 原子链

实施范围：

- 定义 `local_identity_self_service_session_summary.v1`、page、snapshot-bound cursor、self-service actor 和 mutation result。
- 扩展同一 repository owner，提供 user-scoped session list、exact owned session revoke、revoke other sessions，以及 credential replacement + source-bound local-password session revoke aggregate。
- credential rotation 在同一锁 / transaction 中重读 active account、active credential、current-password proof 和 source-bound sessions；旧 credential supersede 与全部目标 session revoke 只有一个提交点。
- memory 覆盖 `121` 条同时间戳分页、state filter、cursor limit / owner / snapshot 绑定、expired boundary、target ownership、recent auth、stale version、bulk atomicity、password reuse、并发单胜者和零部分写入。

批次 A 停止线：

- 不注册或修改 HTTP route，不修改 config、数据库 migration、Pencil、React、CSS、fixture 或专项 checker。
- 不通过循环调用现有单项 `RevokeWebSession` 冒充原子 bulk / rotation transaction。
- 不暴露 credential id、authentication source ref、digest、cookie、IP、User-Agent、audit body 或 raw password。

完成条件：

- [x] owner、schema、cursor、effective state 与 aggregate contract 已落地。
- [x] memory 正向、负向、并发与原子性测试通过。
- [x] 精准 Go 测试、完整 Platform 测试、race 与 fast gate 通过。

完成证据（2026-08-25）：

- `localIdentitySelfServiceSecurityRepository` 作为既有 local identity owner 的 capability interface，已由 `memoryLocalIdentityRepository` 实现；通用 repository interface、SQLite / PostgreSQL adapter 和 HTTP wiring 保持不变。
- session summary / page / cursor、self-service actor、exact revoke、revoke others 与 credential rotation result 已落地；响应结构不包含 `user_id`、credential id、authentication source ref、digest、cookie、audit body、IP 或 User-Agent。
- `121` 条同时间戳、三页 keyset、snapshot 过期边界、后插入排除、active / expired / revoked / all、cursor owner / state / limit / tamper、跨账户 target、recent authentication、CAS、当前 session 与其它 session 均已覆盖。
- bulk revoke 与 credential rotation 都先构建并验证完整变更集再提交；损坏目标证明零部分写入。当前 local-password session 随旧 credential 失效，当前 OIDC session 保留；错误当前密码、密码策略、密码复用、credential unavailable 和四争用者并发单胜者已覆盖。
- 验证通过：精准新增测试；`go test -race ./internal/httpapi/...`；`go test ./...`；`go vet ./...`；`go test -tags postgres ./internal/httpapi -run '^$'`；仓库 fast / full 门禁。

## 批次 B：SQLite / PostgreSQL durable owner

实施范围：

- 以直接 SQL 实现与 memory 同构的 session page 与 aggregate mutation，复用现有 local identity tables。
- 先执行 SQLite / PostgreSQL query-plan 审查；只有稳定分页不能由现有索引满足时，才增加最小 `user_id + created_at DESC + session_id DESC` 顺序索引并推进 store marker。
- credential rotation、source-bound revoke 与 revoke others 均在单一数据库 transaction 中完成；并发只允许一个赢家。
- 覆盖 migration、rollback / reapply、受限 runtime role、同时间戳分页、cursor、CAS、重启、连接失败和 no-fallback。

批次 B 停止线：

- 不创建第二套 session / credential / audit table，不引入 ORM，不自动迁移生产数据库。
- 不启用 production repository、production secret 或 production session store。

完成条件：

- [x] SQLite / PostgreSQL contract 与 memory 语义一致。
- [x] migration 与 query-plan 证据只包含必要变更。
- [x] 双数据库原子性、并发、重启与 no-fallback 通过。

完成证据（2026-08-25）：

- SQLite / PostgreSQL durable adapter 已实现 snapshot-bound session page、exact owned revoke、revoke others 与 credential rotation aggregate；SQLite 使用 `BEGIN IMMEDIATE`，PostgreSQL 锁定 exact account、credential 与 user-scoped session rows，写入与 memory 领域判断只有一个提交点。
- 现有 active / expiry 索引无法提供 self-service 页面的稳定顺序，因此只追加 `0004_local_identity_self_service_sessions` / `local_identity_records_store_v4` 与 `(user_id, created_at DESC, session_id DESC)` ordered index。SQLite / PostgreSQL query-plan 均命中新索引；没有新表、ORM、平行 owner 或生产自动迁移。
- 双库 contract 覆盖 cursor snapshot、同时间戳分页、CAS、并发单胜者、bulk/current/expired 边界和 credential source binding。跨账户 credential id 唯一键冲突发生在 supersede SQL 之后，最终旧 credential 与 session 全部保持不变；该用例直接证明数据库事务没有部分提交。
- SQLite v3 → v4、重放、重启与关闭后 no-fallback 已通过；PostgreSQL 17 完成 v1 / v3 → v4、既有数据保留、受限 runtime role、DDL 拒绝、`ANALYZE + EXPLAIN`、rollback / reapply、pool 重连与 no-fallback。当前 checksum 为 `sha256:80439276fc49f9ca35a61aa321b81ceae404201678739d95197ce43427b1534a`。
- 精准测试、SQLite race、`go vet`、PostgreSQL tagged compile 与 PostgreSQL 聚合集成已通过；批次 B 未注册 HTTP、修改 config / Server startup / Pencil / Web 或打开 production 能力。

## 批次 C：strict HTTP 与 local session 授权

实施范围：

- 固定 `GET /v1/auth/sessions`、exact revoke、revoke others 与 local credential rotate 四条 route contract。
- 所有入口只消费当前 `local_session_dev_test` actor；mutation 要求 exact Origin、double-submit CSRF、recent authentication 或当前密码 proof、strict JSON 与显式影响确认。
- session page 递归拒绝敏感字段；exact revoke 必须重读 ownership 与 expected version；bulk 和 rotation 只调用 aggregate repository operation。
- 当前 session 被撤销时清理 cookie；其它 session mutation 不改写当前 cookie。跨账户 target 使用统一脱敏 scope failure。

批次 C 停止线：

- 不允许 dev header、signed-test token、resource-server Bearer、active workspace 或 membership grant fallback。
- 不新增账户禁用、管理员 credential reset、MFA、恢复、refresh token 或 OIDC upstream logout route。

完成条件：

- [x] exact method / path / query / header / body 与稳定 failure mapping 已固定。
- [x] 成功、scope、stale、CSRF / Origin、recent-auth、password、strict JSON 和零副作用测试通过。
- [x] HTTP 响应、错误和日志敏感字段扫描通过。

完成证据（2026-08-25）：

- 四条 route 已通过 `registerLocalIdentitySelfServiceSecurityHTTPRoutes` 接入既有 local identity HTTP gate；旧 `POST /v1/auth/logout` 仍独立撤销当前 session，exact revoke 不再复用旧的 current-only shortcut。
- session page query 只接受单值 `state / limit / cursor`；mutation 禁止 query，要求 exact Origin、double-submit CSRF 与 strict JSON。exact revoke body 固定 `expected_record_version + confirmed`，revoke others 固定 `confirmed`，credential rotate 固定 `current_password + new_password + session_impact_confirmed`。
- `Server` 显式要求 repository 提供 `localIdentitySelfServiceSecurityRepository` capability，并只从 `local_session_dev_test` middleware 的 current user / session / `last_verified_at` 构造 actor；dev header、Bearer、workspace 与 membership 均不能替代。所有 domain failure 使用专题稳定码与脱敏状态映射，contract body failure 沿用共享 `INVALID_JSON` / local payload 边界。
- 成功链覆盖 snapshot-bound page、tamper cursor、owned exact revoke、current exact revoke、revoke others、local-password current rotation 与 OIDC current rotation。只有 current session 实际失效时清 session / CSRF cookie；其它 mutation 不写当前 session cookie。
- 负向链覆盖跨账户与缺失 target 同码、stale / duplicate CAS、错误 method、非法 / 重复 query、Origin / CSRF、unknown / duplicate field、多份 JSON、未确认、stale recent-auth、错误密码、密码复用及失败零业务副作用；响应和 `logRequestTrace` 扫描未出现 password、cookie、credential id / digest、authentication source、audit ref 或 login identifier。
- 验证通过：self-service HTTP 精准测试、完整 `internal/httpapi`、定向 race、完整 Platform `go test ./...`、`go vet ./...` 与 PostgreSQL tagged compile。批次 C 未修改 Pencil、React、CSS、migration 或生产 gate。

## 批次 D：Pencil 与 React strict consumer

实施范围：

- 按五维提议 `0 / 2 / 2 / 1 / 2 = 7` 采用 `A / 完整 Pencil`，在现有 Authentication Gateway 页面族补 Desktop、Narrow 与关键危险操作状态，不建立 S11。
- session directory 保持单一主 owner；credential rotation 是从属安全动作。设计需覆盖 current / other、active / expired / revoked、single / bulk confirmation 与 forced re-login。
- React 使用单一 strict consumer；password、confirmation 与 raw mutation input 只在组件内存中存在，并在取消、成功、scope change、卸载和路由离开时清空。
- mutation 成功、session / actor 改变和跨标签 signal 都失效旧 cursor、selection、confirmation 与迟到响应。

批次 D 停止线：

- 不在 URL、Web Storage、IndexedDB、Cache Storage、service worker、日志、截图或跨标签消息中保存密码、session directory payload 或确认正文。
- 不复制 S7 User / Role，不建立 device management、全局 session console 或生产状态。

完成条件：

- [x] Pencil 代表面和 Decision Record 通过人工评审。
- [x] strict consumer 与状态测试通过，Web production build 成功。
- [x] Desktop / Narrow 结构符合 Family UI 与无障碍边界。

Pencil 完成证据（2026-08-25）：

- 正确设计源 `docs/designs/radishmind-web-family-ui-v1.pen` 已在现有 Authentication Gateway 第二排页面族后新增 Desktop `pOLcz`、Narrow `LMi7H`、credential rotation danger state `n2O8A5` 与 Decision Record `DASE0`，未修改原 `scHoA` / `uR4Yd` / `SQPBB`。
- Desktop 以 session directory 为唯一主区域，固定 current session、其它 active session、expired / revoked history、exact selected target 与从属 credential rotation；未增加 device、IP、User-Agent、全局账户或 session 搜索。
- Narrow 采用真实纵向重排：current session 固定在前、ended history 收起、credential rotation 退为 disclosure，revoke others 使用 stacked danger confirmation，并明确当前 session 保持 active。
- danger state 只显示空 password input placeholder，不放示例值或掩码值；显式展示 source-bound local-password session 影响、OIDC session 保留、原子失败不变、成功清 cookie 与 forced re-login handoff。
- R21 固定五维评分 `0 / 2 / 2 / 1 / 2 = 7`、全状态覆盖、metadata-only invalidation、组件内存清理、敏感材料禁入边界与停止线；项目所有者已于 2026-08-25 人工批准，记录标记为 `OWNER APPROVED`。
- 四张根画板均已移除 placeholder，并通过 Pencil `ctx.problems` 全树检查；没有裁切、越界或循环尺寸问题。

React 完成证据（2026-08-26）：

- 新增单一 self-service security strict consumer，严格消费 session directory、exact revoke、revoke others 和 credential rotation 四条 canonical route；未知键、递归敏感键、scope 漂移、非法 cursor / state / time 与错误 owner 边界全部失败关闭。
- 新增 scope state 与投影模块：`generation + userId + currentSessionId` 绑定请求与迟到响应，pagination merge 拒绝重复 session，投影要求唯一 current actor 并区分 current / other active / ended 与 local credential impact。
- Authentication Gateway 只接入一个 security panel，使用 metadata-only `session_changed` 跨标签信号。mutation 成功、actor / session / generation 变化与路由离开均失效旧 cursor、selection、confirmation 和迟到响应。
- exact revoke 只发送 target CAS，revoke others 和 credential rotation 只调用 aggregate route；需要 exact target set 的复核只在 directory 完整加载后开放。password / confirmation 只留在组件内存，进入复核前从可见 input state 清空，并在取消、成功、失败、scope 改变、离开与卸载时清理。
- Desktop 与 Narrow 使用 Family UI 语义 token，支持键盘可达的 button / form / dialog，在 `900px` 和 `620px` 两个边界完成单列与纵向危险确认结构；真实三视口、console / network / storage / cookie 审计仍是批次 E 的独立完成条件。
- strict consumer / state 专项与 Web 全量 `398/398` 测试通过，production build 成功；Platform `go test ./internal/httpapi/...` 与 `go test -race ./internal/httpapi/...`、仓库 fast / full 门禁均通过，批次 D 未修改 HTTP、repository、migration、config 或 production gate。

## 批次 E：双数据库产品连续链与收口

实施范围：

- SQLite 创建同账户的多个 local-password / OIDC session，连续完成 session list、exact revoke、revoke others、credential rotation、旧密码失败、新密码登录、双标签失效和服务重启恢复。
- PostgreSQL configured Server 完成同构链、受限 runtime role、migration、Server 停止 no-fallback 与同库重连恢复。
- 浏览器覆盖 `1440×900`、`720×900`、`390×844`，并审计 console、network、URL、storage、cookie 属性和临时产物。
- 回写功能专题、入口、当前焦点、路线图、能力矩阵和周志后关闭专题。

完成条件：

- [ ] 双数据库产品连续链和服务重启恢复通过。
- [ ] 旧 credential、旧 password 与 revoked session 全部失败关闭。
- [ ] 浏览器三视口、跨标签和隐私审计通过。
- [ ] 快速及完整仓库门禁通过。

## 必须保持的负向边界

- 无账户禁用 / 删除、管理员代重置、全局账户或 session 搜索、邀请或批量账户操作。
- 无 MFA、账户恢复、reset token、refresh token、密码历史服务或生产速率限制。
- 无 IP、原始 User-Agent、设备指纹、位置、风险评分或上游 token persistence / revoke。
- 无 identity、session、credential、repository failure fallback，无部分 credential replacement 或部分 session revoke。
- 无 production auth、production session store、production secret、真实 Radish、production IAM 或合规完成声明。

## 验证入口

每批先运行精准测试，再按风险扩展：

```bash
(cd services/platform && go test ./internal/httpapi/...)
(cd services/platform && go test -race ./internal/httpapi/...)
npm --prefix apps/radishmind-web test
npm --prefix apps/radishmind-web run build
./scripts/check-repo.sh --fast
```

批次 B 起涉及 repository / migration / auth / schema，提交前补跑完整：

```bash
./scripts/check-repo.sh
```

## 当前准入状态

- [x] 新长期功能目标已选择。
- [x] owner、用户流程、数据边界、原子语义、批次和停止线已形成设计提议。
- [x] 唯一高风险任务卡已建立。
- [x] 项目所有者已评审并批准批次 A 进入代码。
- [x] 批次 A：领域合同与 memory 原子链。
- [x] 批次 B：SQLite / PostgreSQL durable owner。
- [x] 批次 C：strict HTTP 与 local session 授权。
- [x] 批次 D：Pencil、React strict consumer、状态测试与 Web production build。
- [ ] 批次 E：双数据库产品连续链与专题收口。
