# 工作区成员邀请、认领与到期治理 v1 高风险任务卡

更新时间：2026-09-02

状态：`workspace_member_invitation_claim_expiry_governance_dev_test_v1_batch_a_ready`

对应功能设计：[工作区成员邀请、认领与到期治理（开发 / 测试态）v1](../features/admin-control-plane/workspace-member-invitation-claim-expiry-governance-dev-test-v1.md)

## 任务目标

在唯一 local identity / authorization repository 内建立一次性 `WorkspaceInvitation` aggregate，让管理员可以在不知道目标 `user_id` 的情况下预先表达 exact workspace + non-admin role 的限时准入意图；认领者必须先使用 active 本地 Web Session 登录，只有 claim 单事务提交后才创建既有 `WorkspaceMembership`、`LocalRoleAssignment` 并消费 invitation。

本任务不得创建第二套账户、membership、role、permission、Session 或审计 owner，也不得把邀请码、preview 或前端状态解释为访问资格。

## 已批准前提

- 项目所有者已于 2026-09-02 批准功能设计、owner、非定向邀请码、禁止邀请 `workspace_admin`、四档 TTL、一次性 secret、preview → claim、原子 membership + assignment、五批实施顺序与 `A / 完整 Pencil` 边界，并授权批次 A 进入代码。
- [本地用户、角色与工作区成员管理 v1](../features/admin-control-plane/local-user-role-workspace-membership-administration-dev-test-v1.md)保持完成关闭；本任务是新的独立功能，不是其 Batch F。
- `UserAccount`、`WebSession`、`WorkspaceMembership`、`LocalRoleAssignment`、内建角色目录与三种 local identity repository 继续是 canonical owner。
- invitation 不绑定 email、login identifier、display name、OIDC identity 或预先已知 `user_id`；任意 active 同 tenant 本地账户只有持有有效 code 并通过 claim 才能成为认领者。
- v1 只允许 `workspace_reader`、`workspace_builder`、`workspace_reviewer`；`workspace_admin` 仍必须在成员身份明确后通过现有 exact member role assignment 流程授予。
- TTL 只允许 `1h / 24h / 72h / 7d`。服务端使用 UTC clock 计算 `expires_at`，不接受客户端任意时间戳或无限期。
- 邀请码使用 locator + 至少 256-bit CSPRNG secret。创建响应只返回一次原文，持久层只存 digest；格式合法但 locator 不存在时仍完成固定 dummy digest 的等价比较。
- 设计批准不提前开放批次 B 以后的 migration、HTTP、Pencil、React、真实产品链或 production 能力。

## 批次 A：canonical contract、secret policy 与 memory 原子链

实施范围：

- 定义 `workspace_invitation.v1` canonical contract、API-safe current / preview / mutation projection、filter-bound cursor、effective expiry 和稳定 failure code。
- invitation current record 固定 exact tenant / workspace、role key、catalog version / definition digest、`pending|claimed|revoked` lifecycle、TTL、terminal membership / assignment refs 与 request / audit metadata；secret digest 只存在于 repository-private state。
- 建立 code generator / parser、CSPRNG secret、digest、constant-time comparison 与 unknown-locator dummy comparison；测试和错误中不输出原始 code 或 digest。
- 在唯一 memory local identity repository 上增加 invitation capability，完成 create、workspace-scoped list、exact revoke、preview 与 claim。
- claim 在同一锁 / aggregate commit 中重新读取 invitation、claimant account、tenant、role catalog、membership / assignment invariant，原子创建 membership + catalog-derived assignment 并把 invitation 变为 claimed。
- list 固定 `updated_at DESC, invitation_id DESC`，默认 `50`、最大 `100`；cursor 绑定 tenant、workspace、effective state、limit 与首个请求的 `as_of`。
- 覆盖允许 / 禁止角色、四档 TTL、catalog drift、expired / revoked / claimed、已撤销 membership 重新加入、未撤销过期 membership 冲突、并发 claim 单胜者、重放、cursor tamper、repository corruption 和零部分写入。

批次 A 停止线：

- 不注册或修改 HTTP route，不修改 config、SQLite / PostgreSQL migration、Pencil、React、CSS、launcher、fixture 或专项 checker。
- 不为 invitation 建立独立数据库或第二套 identity / authorization repository。
- 不让 preview、code validity、active workspace header、OIDC claim 或 Web state 替代 membership decision。
- 不实现 email / login / display name / OIDC directory search、消息交付、批量邀请、管理员邀请、自动登录或 production rate limit。

批次 A 完成条件：

- [ ] canonical schema、domain validation、API-safe projection、cursor 与 failure mapping 落地。
- [ ] memory create / list / revoke / preview / claim 正向、负向、并发与原子性测试通过。
- [ ] 原始 code / secret digest 未进入 canonical projection、错误、日志、audit、fixture 或快照。
- [ ] 精准 Go / contract 测试、Platform 普通测试、race 与仓库快速门禁通过。
- [ ] 回写功能专题、任务卡、当前焦点与 W36 周志后提交；批次 B 仍等待下一次明确推进。

## 批次 B：SQLite / PostgreSQL durable owner

实施范围：

- 评审并新增最小 invitation table、scope / lifecycle / expiry / stable-order index 与 schema marker；digest 使用 repository-only 列，不进入 canonical JSON projection。
- 两种 durable store 实现与 memory 一致的 create / list / revoke / preview / claim、scope lock、CAS、cursor `as_of` 和原子 membership + assignment + terminal invitation。
- 覆盖 migration / rollback / reapply、受限 runtime role、并发 claim、重启恢复、损坏载荷、database unavailable 与 no-fallback。

停止线：不自动迁移 production，不注册 HTTP，不启动 Web，不引入新数据库或 ORM。

## 批次 C：strict HTTP 与本地 Session 安全边界

实施范围：

- 注册三条 Admin route：workspace-scoped list、create、revoke；组合重读现有 read / write permissions、active membership、recent auth、Origin、CSRF 与 strict JSON。
- 注册两条 claimant route：preview 与 claim；只接受 active `local_session_dev_test`，目标 scope 来自 secret 匹配后的 invitation，不要求预先拥有目标 workspace membership。
- 创建响应设置 `Cache-Control: no-store`；invalid locator / secret 使用统一失败，Bearer、dev header、signed-test 与 resource-server OIDC 均不得 fallback。
- 覆盖 method / query / body strictness、scope、权限组合、CSRF、recent auth、枚举、重放、CAS、稳定 recovery 与敏感字段禁入。

停止线：不修改 Pencil / React，不增加邮件、全局目录、真实 Radish、production auth / IAM 或速率限制声明。

## 批次 D：完整 Pencil 与人工批准

实施范围：

- 按五维 `1 / 2 / 2 / 2 / 1 = 8` 采用 `A / 完整 Pencil`，只扩既有 S7 与 Authentication Gateway，不建立 S11。
- 冻结最小 Admin Desktop、Claim Desktop、无法直接推导的 Narrow / terminal risk state 与共享 Decision Record。
- 覆盖 invitation directory、role / TTL review、一次性 code handoff、preview / confirmed claim、expired / revoked / claimed / invalid、scope / actor 失效与窄屏顺序。
- 完成 Pencil 原生裁切、越界、placeholder、命名、文字、fill / stroke 与语义 token 检查，并由项目所有者人工批准。

停止线：Pencil 批准前不实施 React；设计不复制完整 S7 / Authentication 页面，也不把代表数据写成运行时事实。

## 批次 E：React strict consumer、双数据库产品链与收口

实施范围：

- 建立单一 invitation strict consumer，在 S7 与 Authentication Gateway 复用同一 authority / generation / secret cleanup 模型。
- Admin code handoff 与 claimant code input 只存在于组件内存；scope / actor / Session / route 变化、取消、成功、terminal failure 和卸载均清理。
- SQLite 完成 admin create → one-time handoff → 第二账户登录 / preview / claim → permission 生效 → duplicate claim 拒绝 → 服务重启恢复。
- PostgreSQL configured Server 完成同构链、并发单胜者、停止 no-fallback 与 reconnect。
- 覆盖双标签、`1440×900`、`720×900`、`390×844`、console / network / URL / storage / cookie / database 隐私审计并回写真相源。

停止线：不自动选择 workspace，不发消息，不打开 `workspace_admin` invitation、production delivery、production auth / IAM 或真实 Radish。

## 必须保持的负向边界

- 无 email、login identifier、display name、OIDC issuer / subject 搜索、绑定或自动合并。
- 无 URL invite link、邮件、短信、站内信、webhook、公开分享、批量邀请或批量认领。
- 无 `workspace_admin` invitation、客户端 grants、自定义角色、角色组合或自动角色推荐。
- 无 invitation 续期、恢复、编辑、重复使用、background expiry worker、自动清理、delete 或永久 purge。
- 无 invitation → account / external identity / Session / API Key 创建，无自动登录或 active workspace 选择。
- 无第二套账户、membership、role、permission、Session、audit 或 repository owner。
- 无 production auth / IAM、production secret、production invitation delivery、真实 Radish 批次 E 或合规完成声明。

## 验证入口

每批先执行精准测试，再按风险扩展：

```bash
(cd services/platform && go test ./internal/httpapi/...)
(cd services/platform && go test -race ./internal/httpapi/...)
npm --prefix apps/radishmind-web test
npm --prefix apps/radishmind-web run build
./scripts/check-repo.sh --fast
```

批次 B 起涉及 schema / migration / durable auth owner，提交前补跑完整：

```bash
./scripts/check-repo.sh
```

## 当前准入状态

- [x] 新长期功能目标已选择。
- [x] owner、用户流程、数据边界、原子语义、批次与停止线已形成正式功能设计。
- [x] 项目所有者已批准设计与批次 A 进入代码。
- [x] 唯一高风险任务卡已建立。
- [ ] 批次 A：canonical contract、secret policy 与 memory 原子链。
- [ ] 批次 B：SQLite / PostgreSQL durable owner。
- [ ] 批次 C：strict HTTP 与 local Session 安全边界。
- [ ] 批次 D：完整 Pencil 与人工批准。
- [ ] 批次 E：React strict consumer、双数据库产品链与专题收口。
