# 本地账户凭证轮换与自助会话治理 v1 高风险任务卡

更新时间：2026-08-25

状态：`local_account_credential_rotation_self_service_session_governance_dev_test_v1_design_proposed_review_required`

对应功能设计：[本地账户凭证轮换与自助会话治理（开发 / 测试态）v1](../features/admin-control-plane/local-account-credential-rotation-self-service-session-governance-dev-test-v1.md)

## 任务目标

在唯一 `localIdentityRepository` 内建立当前账户的 session directory、exact session revoke、revoke others 与本地 credential rotation 原子链，再依次接入双数据库、strict HTTP、Authentication / Account Web 和真实产品连续验证。实现不得复制身份 owner，也不得把开发测试态会话治理解释为 production auth。

## 已确认前提

- 联合登录批次 A 至 D、本地成员管理批次 A 至 E 均已完成；真实 Radish 批次 E 仍为外部阻塞。
- `UserAccount`、`LocalCredential`、`WebSession` 与 `ExternalIdentityBinding` 继续由现有 local identity repository 单一拥有。
- 当前 `POST /v1/auth/sessions/{session_id}/revoke` 只允许撤销当前 session；现有 `ReplaceCredential` 与 `RevokeWebSession` 是彼此独立的 repository 原语，不能直接证明新产品语义已经成立。
- 本任务只开放当前账户 self-service，不新增管理员账户禁用、代重置、全局 session 搜索、MFA、恢复或生产能力。
- 功能专题的原子语义、路由提议、Pencil 覆盖级别和停止线必须先由项目所有者评审；评审前不得修改运行时代码。

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

- [ ] owner、schema、cursor、effective state 与 aggregate contract 已落地。
- [ ] memory 正向、负向、并发与原子性测试通过。
- [ ] 精准 Go 测试、完整 Platform 测试、race 与 fast gate 通过。

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

- [ ] SQLite / PostgreSQL contract 与 memory 语义一致。
- [ ] migration 与 query-plan 证据只包含必要变更。
- [ ] 双数据库原子性、并发、重启与 no-fallback 通过。

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

- [ ] exact method / path / query / header / body 与稳定 failure mapping 已固定。
- [ ] 成功、scope、stale、CSRF / Origin、recent-auth、password、strict JSON 和零副作用测试通过。
- [ ] HTTP 响应、错误和日志敏感字段扫描通过。

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

- [ ] Pencil 代表面和 Decision Record 通过人工评审。
- [ ] strict consumer 与状态测试通过，Web production build 成功。
- [ ] Desktop / Narrow 结构符合 Family UI 与无障碍边界。

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
- [ ] 项目所有者已评审并批准批次 A 进入代码。
- [ ] 批次 A：领域合同与 memory 原子链。
- [ ] 批次 B：SQLite / PostgreSQL durable owner。
- [ ] 批次 C：strict HTTP 与 local session 授权。
- [ ] 批次 D：Pencil 与 React strict consumer。
- [ ] 批次 E：双数据库产品连续链与专题收口。
