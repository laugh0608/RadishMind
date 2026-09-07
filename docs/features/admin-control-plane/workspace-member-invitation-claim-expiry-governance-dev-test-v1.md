# 工作区成员邀请、认领与到期治理（开发 / 测试态）v1

状态：`workspace_member_invitation_claim_expiry_governance_dev_test_v1_batch_d_pencil_approved_batch_e_ready`

更新时间：2026-09-05

## 设计结论

本专题是应用定时回归评测关闭后，从四个一级产品面重新评审选出的下一项长期开发目标。它补齐现有本地成员管理的真实断点：管理员目前必须先在线下获得一个已有 active 本地账户的 exact `user_id`，再分别创建 `WorkspaceMembership` 和角色 assignment；产品没有“管理员预先表达准入意图，成员登录后自行认领”的安全闭环。

项目所有者先批准该方向进入功能设计，并于 2026-09-02 进一步批准本文的一次性邀请代码、登录后预览、显式认领、到期与撤销治理、五批实施顺序和 `A / 完整 Pencil` 边界。[唯一高风险任务卡](../../task-cards/workspace-member-invitation-claim-expiry-governance-dev-test-v1-plan.md)已经建立，批次 A 至 C 已完成；批次 D 的五块正式 Pencil 代表面也已完成原生静态 QA，并于 2026-09-05 获项目所有者人工视觉与安全边界批准。它不是已关闭成员管理专题的 Batch F，也不修改该专题的完成事实。

首版最重要的边界是：邀请只保存待认领授权意图，`WorkspaceMembership` 继续是 workspace 访问的唯一 owner，`LocalRoleAssignment` 继续是角色与冻结 grants 的唯一 owner。邀请码不是 membership、Session、API Key 或可复用授权 token；只有在服务端单事务认领成功后，成员与角色权限才生效。

## 选择依据

- [本地用户、角色与工作区成员管理 v1](local-user-role-workspace-membership-administration-dev-test-v1.md)已经完成账户状态校验、成员目录、内建角色目录、membership / role mutation、CAS、三种 store、strict Admin HTTP 和 S7 Web；当前唯一入会入口仍要求 exact `user_id`。
- [本地账户与 Radish OIDC 联合登录 v1](local-account-radish-oidc-federated-login-v1.md)已经提供本地注册、登录、Web Session 和本地账户 owner。认领者可先注册或登录，再提交邀请码；本专题不创建账户，也不等待真实 Radish OIDC。
- 现有角色目录能从 `role_key + catalog_version + definition_digest` 派生并冻结 exact grants；邀请无需保存客户端 grants 或建立第二套角色策略。
- 本功能可以在 memory、SQLite、PostgreSQL、本地 Web Session 与真实浏览器中形成完整连续链，不依赖邮件服务、用户目录、production secret、真实 issuer 或外部部署资源。
- 相比 Gateway 短窗口速率限制和 RAG 版本恢复，本目标优先解决用户加入 workspace 的前置阻塞，并能复用成熟的 identity / authorization 基座而不沿刚关闭专题继续派生同层切片。

## 现有事实与缺口

| 事实 | 当前 owner | 对本专题的影响 |
| --- | --- | --- |
| 本地账户与登录 | `UserAccount`、`LocalCredential`、`WebSession` | 认领只接受已登录 active 本地账户，不通过邀请创建或恢复账户 |
| workspace 访问 | `WorkspaceMembership` | 继续是唯一访问 owner；邀请 pending / preview 不产生访问资格 |
| 角色授权 | `LocalRoleAssignment` 与内建角色目录 | 认领成功时由 exact catalog definition 派生一个 assignment，客户端不提交 grants |
| 管理员添加成员 | local identity administration service | 当前需要 exact `user_id`，且 membership 与 role 是两个独立人工步骤 |
| 目录隐私 | workspace-scoped member projection | 不存在全局账户、email、login identifier 或 OIDC directory search，本专题也不新增 |
| 仓储模式 | memory / SQLite / PostgreSQL dev/test | Invitation 必须同构持久化、并发单胜者、重启恢复且数据库失败不回退 memory |
| Web 基准面 | S7 User / Role 与 Authentication Gateway | Admin 创建 / 治理进入 S7；成员预览 / 认领进入 Authentication，不建立 S11 |

## 已批准产品决策

### 邀请对象与角色

1. v1 邀请不绑定 email、login identifier、display name、OIDC identity 或预先已知 `user_id`。邀请码的持有者在登录后成为候选认领者，服务端只接受与邀请 `tenant_ref` 相同的 active 本地账户。
2. 每份邀请只绑定一个 exact `tenant_ref + workspace_id`、一个内建 `role_key`、创建时的 `catalog_version + role_definition_digest` 和一个服务端计算的到期时间。
3. v1 只允许邀请 `workspace_reader`、`workspace_builder` 或 `workspace_reviewer`；禁止通过 bearer-like 邀请代码直接授予 `workspace_admin`。管理员角色仍只能在成员身份明确后，通过现有 exact member role assignment 流程授予。
4. 邀请不配置客户端 grants，不复制 permission matrix，也不保存账户或 Session 快照。角色目录发生任何版本或 definition digest 漂移后，旧 pending invitation 不再可认领，管理员必须撤销并创建新邀请。
5. 首版 invitation TTL 只允许服务端 allowlist：`1h`、`24h`、`72h`、`7d`。客户端提交固定枚举，服务端使用自身 UTC clock 计算 `expires_at`；不接受任意时间戳、无限有效期或自动续期。

### 认领资格

认领者必须同时满足：

- 使用显式 `local_session_dev_test` Web Session，账户仍为 active；
- Session 的 tenant 与 invitation tenant 精确相同；
- 最近认证仍在现有十分钟窗口内，并通过同源 Origin、CSRF 与显式确认；
- invitation secret 匹配、尚未认领或撤销，且服务端当前时间早于 `expires_at`；
- 创建时冻结的 role catalog version 与 definition digest 仍等于当前 canonical catalog；
- 当前用户在目标 workspace 不存在 active 或尚未显式撤销的过期 membership，也不存在违反 owner invariant 的 active workspace role assignment。

已撤销 membership 历史不会阻止重新加入；认领成功会创建新的 membership 与 assignment stable id。仍占据 active lifecycle 的过期 membership 必须先由管理员在现有成员管理流程中显式撤销，邀请不会隐式覆盖、复活或合并旧记录。

## Owner 与数据边界

| 资源 | owner | 本专题允许的新增能力 | 明确不做 |
| --- | --- | --- | --- |
| `WorkspaceInvitation` | local identity / authorization repository 内的新独立 invitation aggregate | create、workspace-scoped list、exact revoke、secret verify、单次 claim | 不成为 membership、role、Session、用户目录或审计 owner |
| `WorkspaceMembership` | 既有 RadishMind Authorization | claim 单事务创建一个新 active membership | 不由 preview、邀请码存在或前端状态隐式生效 |
| `LocalRoleAssignment` | 既有 RadishMind Authorization | claim 按冻结且仍匹配的 canonical role definition 创建一个 assignment | 不接受客户端 grants，不授予 `workspace_admin` |
| `UserAccount` | 既有 RadishMind Identity | claim 时读取当前账户 active 与 exact tenant | 不创建账户，不搜索或投影 login / email / OIDC facts |
| `WebSession` | 既有 RadishMind Session | 恢复 administrator 或 claimant actor，执行近期认证、Origin 与 CSRF 校验 | 不持久化 invitation code，不生成 invitation Session |
| 内建角色目录 | 既有 canonical policy | create / preview / claim 重读 role definition | 不建立数据库角色目录，不静默升级 invitation 或历史 assignment |
| Radish OIDC / 外部消息 | 外部集成 | 无新增能力 | 不发邮件、站内信或 webhook，不按 claim 自动创建 external identity |

### `WorkspaceInvitation` current record

计划定义 `workspace_invitation.v1`，至少包含：

- `invitation_id`、`record_version`；
- `tenant_ref`、`workspace_id`；
- `role_key`、`role_catalog_version`、`role_definition_digest`；
- `lifecycle_state=pending|claimed|revoked`；
- `expires_at`；
- `created_at`、`updated_at`、`created_by_actor_ref`；
- terminal 时的 `claimed_at + claimed_by_user_id + membership_id + assignment_id`，或 `revoked_at + revoked_by_actor_ref`；
- mutation 对应的 request / audit ref。

API projection 另返回按请求 `as_of` 计算的 `effective_state=pending|claimed|revoked|expired`。`expired` 是时间派生的终态投影，不依赖后台清理任务，也不把已过期记录原地改写为新的授权事实。

repository-only secret material 与 canonical projection 分离。持久层只保存 secret digest；API、JSON projection、cursor、audit、错误、日志、fixture 与页面列表均不得返回 digest。

## 一次性邀请码安全模型

邀请码计划采用可解析定位与高熵 secret 分离的格式：

```text
rmi_<invitation-id-fragment>.<random-secret>
```

- random secret 至少提供 `256 bit` CSPRNG entropy；服务端只保存 digest，并以 constant-time comparison 校验。
- 创建成功响应是唯一返回完整邀请码的时刻，并设置 `Cache-Control: no-store`；列表、详情、刷新与重启均不能恢复原文。
- 邀请码不放入 URL、hash、cookie、Web Storage、IndexedDB、Cache Storage、service worker、analytics、跨标签消息或可持久化前端状态。
- 管理员一次性交接和认领者输入只存在于各自 React 组件内存；scope / actor 变化、取消、成功、terminal failure、路由离开和卸载都会清理。
- 预览与认领均通过 request body 提交邀请码。非法格式、未知 id 与 secret 不匹配统一返回同一公开失败；格式合法但 locator 不存在时仍对固定 dummy digest 完成等价比较，再返回相同失败，避免形成显著的 invitation 枚举 oracle。
- 创建响应遗失后不能重新读取 secret；管理员只能撤销 metadata record 并新建邀请。
- 本专题不把开发测试态入口冒充 production credential distribution，也不以“高熵”替代未来 production rate limit、abuse detection 或消息交付安全评审。

## 生命周期与原子认领

Invitation 持久 lifecycle 只允许：

```text
pending -> claimed
pending -> revoked
```

到期由 `pending + now >= expires_at` 计算为 effective `expired`；`claimed`、`revoked` 和 `expired` 都不能恢复为 pending，不能延长、重复认领或修改角色。需要变更 scope、role 或 TTL 时只能创建新 invitation。

认领必须在单一 local identity repository transaction 中完成：

1. 按 code 中的 locator 锁定 exact invitation，并以 constant-time comparison 验证 secret；
2. 重新读取 invitation lifecycle、expiry、scope 与 frozen role definition；
3. 重新读取 claimant account active / tenant，并检查目标 workspace membership 与 role invariant；
4. 从当前且 exact-match 的 canonical role definition 派生排序后 grants；
5. 创建新的 `WorkspaceMembership` 与 `LocalRoleAssignment`；
6. 把 invitation CAS 更新为 `claimed`，固定 claimant、membership 与 assignment refs；
7. 三项写入与 invitation terminal transition 一起提交。

任一步失败都必须全部回滚。并发预览不占用邀请；并发 claim 只能有一个提交成功。成功响应必须来自已提交的 canonical membership / assignment / invitation projection，Web 不自行推算权限或版本。

## 已批准的 Admin 与认领 HTTP 边界

| 方法与路径 | 授权 | 语义 |
| --- | --- | --- |
| `GET /v1/admin/local-identity/workspaces/{workspace_id}/invitations` | local Session + `local_identity_members:read` + `local_identity_roles:read` | workspace-scoped 脱敏邀请目录 |
| `POST /v1/admin/local-identity/workspaces/{workspace_id}/invitations` | local Session + `local_identity_memberships:write` + `local_identity_roles:assign` + recent auth | 创建 invitation 并一次性返回 code |
| `POST /v1/admin/local-identity/workspaces/{workspace_id}/invitations/{invitation_id}/revoke` | 与 create 相同 + CAS | 撤销仍可认领的 pending invitation |
| `POST /v1/auth/workspace-invitations/preview` | active local Session + recent auth | 验证 code 后返回 exact tenant / workspace id、role summary 与 expiry，不产生授权写入 |
| `POST /v1/auth/workspace-invitations/claim` | active local Session + recent auth + Origin / CSRF + confirmed | 原子消费 invitation 并创建 membership + assignment |

Admin route 继续要求单值 `X-RadishMind-Active-Tenant` 与 `X-RadishMind-Active-Workspace`，并逐请求重读管理员 account、membership 和所有 required permissions。不得把两个 permission 的任一单项解释为足够授权，也不新增宽泛 `is_admin` 旁路。

认领 route 不要求 claimant 预先拥有目标 workspace membership，也不接受 active workspace header 作为授权事实；目标 scope 只来自 secret 匹配后的 invitation record，并必须与 claimant account tenant 匹配。Bearer、dev header、signed-test token 和 resource-server OIDC token 均不能成为 local Session 失败时的 fallback。

strict body 提议：

- create：`role_key + expected_catalog_version + expected_role_definition_digest + ttl_policy + confirmed`；
- revoke：`expected_record_version + confirmed`；
- preview：`invitation_code`；
- claim：`invitation_code + expected_record_version + confirmed`。

所有 body 拒绝未知字段、重复字段、客户端 tenant / workspace、`user_id`、`permission_grants`、任意 expiry 时间戳或多份 JSON 文档。preview 返回的 `record_version` 只服务 claim CAS；客户端不能据此断言 invitation 仍可认领。

## 目录、cursor 与可见性

- Admin invitation directory 只列 exact tenant / workspace，不提供跨 workspace 或全局搜索。
- filter 允许 `effective_state=pending|claimed|revoked|expired`；默认 `pending`。
- 排序固定为 `updated_at DESC, invitation_id DESC`，默认 limit `50`，最大 `100`。
- cursor 绑定 tenant、workspace、effective state、limit 与首个请求的 `as_of`；后续页沿同一 `as_of` 计算 expiry，避免 invitation 在翻页过程中跨越到期点导致重复或遗漏。
- claimed projection 可返回已经成为当前 workspace member 的 `claimed_by_user_id`、membership / assignment stable ref；pending / revoked / expired 不投影候选账户事实。
- directory、preview 与 claim response 不返回 secret digest、login identifier、email、credential、Session、cookie、issuer、subject、raw claim、完整 grants 或 audit body。

## 稳定失败分类

Admin surface 至少固定：

- `workspace_invitation_admin_unavailable`；
- `workspace_invitation_cursor_invalid`；
- `workspace_invitation_role_ineligible`；
- `workspace_invitation_role_catalog_mismatch`；
- `workspace_invitation_version_conflict`；
- `workspace_invitation_transition_invalid`；
- `local_identity_recent_authentication_required`；
- `workspace_membership_denied`；
- `workspace_permission_denied`。

Claimant surface 至少固定：

- `workspace_invitation_invalid`：非法格式、未知 locator 或 secret 不匹配统一使用；
- `workspace_invitation_not_claimable`：secret 已验证但 invitation 已认领、撤销、过期或 role catalog 漂移；
- `workspace_invitation_account_ineligible`：当前账户状态或 tenant 不满足；
- `workspace_invitation_membership_conflict`：已有未撤销 membership 或 owner invariant 冲突；
- `workspace_invitation_version_conflict`：preview 后 invitation 已变化；
- `workspace_invitation_store_unavailable`；
- `local_identity_recent_authentication_required`。

公开响应只给 stable failure code、request ref、可执行恢复方向和允许公开的 current version / effective state。invalid code 不返回 invitation 是否存在、scope、role、expiry 或其它 identity fact；repository、SQL、secret digest 和 raw auth error 始终不出边界。

## Web 与 Family UI

- S7 User / Role 增加 Invitation 任务，但继续共享同一 Admin tenant / workspace context；邀请目录是独立当前 owner，不与 member directory 合并成伪统一列表。
- Admin create 使用既有 `Ephemeral Credential Handoff` 模式：角色与 TTL review → 显式创建 → 一次性 code handoff → copy / done。离开 handoff 后不能恢复 code。
- Authentication Gateway 增加“认领工作区邀请”任务：输入 code → server preview → 审查 exact workspace / role / expiry → confirmed claim → 成功后重新加载可用 workspace，不自动导航或替用户选择 active workspace。
- 普通 invitation 状态与当前 selection 分离；只有驱动 detail / confirmation 的当前 invitation 使用选中语义，expired、revoked、claimed 使用独立文字、图标与语义状态。
- workspace、actor、Session 或 route 变化会清除 code、preview、selection、confirmation、cursor 与迟到响应；Admin create / revoke 和 claimant claim 成功后同样失效旧状态。

五维评分固定为 `1 / 2 / 2 / 2 / 1 = 8`，采用 `A / 完整 Pencil`：复用 S7 与 Authentication Gateway，不建立 S11；但一次性凭据交接、双 actor、preview → claim、terminal 状态与窄屏顺序存在不可从旧页面安全推导的新决策。2026-09-05 已完成 S7 Admin Desktop `q4s4X`、S7 one-time handoff Narrow `L2SJt`、Authentication Claim Desktop `MsAm1`、invalid / terminal Narrow `HSfOp` 与 R25 Decision `RY0Xx`。五块根画板共 `630` 个节点；Pencil 原生检查对裁切 / 越界、placeholder、节点命名、文字内容 / fill、硬编码 fill / stroke、原始 `rmi_` code 和小于 `7px` 文字均为 `0` 问题。项目所有者同日完成人工视觉与安全边界审查并批准；R25 已更新为 `OWNER APPROVED · 2026-09-05`。该批准不自动授权 React 或批次 E。

## 已批准的实施拆分

### 批次 A：canonical contract、secret policy 与 memory 原子链

- 定义 invitation canonical schema、API-safe projection、cursor、failure codes 与 code parser / digest policy。
- 在唯一 local identity repository 内建立 invitation aggregate 和 memory implementation。
- 完成 create、list、revoke、preview、claim，以及 claim 的 membership + assignment + terminal invitation 单事务语义。
- 覆盖 role allowlist、TTL、catalog drift、已撤销 membership 重新加入、未撤销过期 membership 冲突、并发单胜者、constant-time secret compare 边界和零部分写入。

停止线：不注册 HTTP，不改 config / migration / Pencil / Web，不创建第二套 membership 或 role owner。

批次 A 完成事实：`workspace_invitation.v1`、current / preview / mutation projection、四态 effective expiry、固定 `as_of` cursor 与 stable failure mapping 已落地；邀请码使用完整 locator 与 256-bit CSPRNG secret，repository 只保存 domain-separated digest，未知 locator 同样执行 dummy digest constant-time comparison。既有 memory local identity repository 在同一写锁内完成 claimant account / tenant、catalog、membership / assignment invariant 重读，再一次提交新 membership、catalog-derived assignment 与 claimed invitation。测试覆盖三种允许角色、禁止 `workspace_admin`、四档 TTL、catalog drift、expired / revoked / claimed、已撤销 membership 重新加入、未撤销过期 membership 与孤立 active assignment 冲突、cursor tamper、repository corruption、重放和 `24` 路并发单胜者；所有失败路径均保持零部分写入。批次 A 未注册 route，未修改 config、migration、Pencil、React、CSS、launcher、fixture 或专项 checker。

### 批次 B：SQLite / PostgreSQL durable owner

- 评审并新增最小 invitation table / index / migration；secret digest 使用 repository-only 列，不进入 canonical JSON projection。
- 实现两种 durable store 的 scope lock、CAS、stable cursor、`as_of` expiry、原子 claim 与重启恢复。
- 覆盖 migration / rollback / reapply、受限 runtime role、并发 claim、repository corruption、database unavailable 与 no-fallback。

停止线：不注册 HTTP，不自动迁移 production，不启动外部服务或引入新数据库 / ORM。

批次 B 完成事实：SQLite 与 PostgreSQL 在既有 `local_identity_records` migration 链追加 `0005_workspace_invitations` / `local_identity_records_store_v5`，物化 repository-only digest、scope / lifecycle / expiry 与稳定目录索引，没有建立新数据库、DSN、pool 或 ORM。两种 owner 与 memory 共用同一 canonical service：SQLite 使用 `BEGIN IMMEDIATE`，PostgreSQL 使用既有 workspace advisory lock 与行锁；claim 在同一事务内重读 invitation、claimant account、catalog 和 membership / assignment invariant，再创建既有 membership、catalog-derived assignment 并 CAS 写入 claimed terminal refs。SQLite 文件库与 PostgreSQL 17 受限 runtime 已覆盖 stable cursor `as_of`、create / list / revoke / preview / claim、`16` 路并发单胜者、v4 → v5、query plan、重启、损坏载荷、关闭数据库失败关闭、rollback / reapply 与 no-fallback；测试用回环 PostgreSQL 服务已在验证后关闭。批次 B 没有注册 HTTP、修改 config / Server startup、Pencil、React、CSS 或打开 production 自动迁移。

### 批次 C：strict HTTP 与本地 Session 安全边界

- 注册三条 Admin 与两条 claimant route，复用 local Session、recent auth、Origin、CSRF、request / audit ref 和 strict JSON。
- Admin 每次重读 exact scope 与组合 permission；claimant route 显式跳过“已有目标 workspace membership”前置，但不跳过 account / tenant / Session / secret / invitation 校验。
- 设置一次性响应 no-store、日志脱敏与稳定 recovery metadata；覆盖枚举、重放、scope、auth-mode fallback、CSRF 和隐私负向路径。

停止线：不改 Pencil / React，不增加 production rate limit、邮件、全局目录或真实 OIDC。

批次 C 完成事实：正式 Server 已装配单一 invitation service，并注册 workspace-scoped list / create / revoke 与 claimant preview / claim 五条 strict route。Admin route 要求单值 active tenant / workspace header、active membership、精确的 read 或 write 组合权限与 recent auth；create / revoke 复用同源 Origin、CSRF、确认、strict JSON 与 canonical request / audit refs。claimant route 只接受 active `local_session_dev_test`，从 invitation secret 验证后的 record 取得目标 scope，不预先要求目标 workspace membership，也不采信 active workspace header；claim 继续执行 Origin / CSRF 与原子 membership + assignment + terminal invitation。创建、preview 与 claim 成功响应均为 `no-store`，客户端 request id 在 invitation route 上不会进入 trace / log，invalid format / locator / secret 统一失败且不公开 scope。测试覆盖五 route 连续链、权限即时生效、revoke、枚举、重放、CAS、tenant、recent auth、Bearer / dev header / signed-test / resource-server OIDC fallback、Origin / CSRF、method / query / body strictness、组合权限、稳定 recovery 和敏感字段禁入；没有修改 Pencil、React、config、migration 或 production 能力。

### 批次 D：完整 Pencil 与人工批准

- 在既有 S7 / Authentication Gateway 页面族冻结 Admin directory / create handoff、claim preview / confirm、terminal state 和 Narrow 顺序。
- 完成 Pencil 原生结构、裁切、placeholder、命名、文字与语义 token 检查。
- 项目所有者人工批准设计与安全表达前，不进入 React。

Pencil 完成证据（2026-09-05）：

- S7 Desktop `q4s4X` 保持 invitation directory 为唯一主对象，当前 selection 与 pending / claimed / expired / revoked 状态分离；从属 rail 冻结 role / TTL review、一次性 code handoff、不可恢复说明与 `Done & clear`。
- S7 Narrow `L2SJt` 固定 `workspace context → Create task → role / TTL review → one-time code → clear / access boundary`，设计源不保存示例 code。
- Authentication Desktop `MsAm1` 固定 `code memory → server preview → explicit confirmation → atomic claim`，并说明 preview 不授权、CAS 不预占、失败零部分写入以及成功后只 reload 可用 workspace、不自动选择。
- Authentication Narrow `HSfOp` 以统一 `workspace_invitation_invalid` 承接非法格式、未知 locator 与 secret mismatch；只有 secret 已验证后才区分 claimed / revoked / expired / role drift，并给出创建新邀请而非恢复 secret 的恢复路径。
- R25 `RY0Xx` 固定双 actor、owner、四态 effective projection、invalid 枚举边界、scope / actor / Session / route 失效、敏感材料禁入与全部停止线。五块画板共 `630` 个节点，原生静态 QA 全部为零问题；项目所有者已于 2026-09-05 完成人工视觉与安全边界审查并批准。

### 批次 E：React strict consumer、双数据库产品链与收口

- 建立单一 invitation strict consumer 与两个既有页面族的共享状态失效 / secret cleanup 模型。
- SQLite 完成 admin create → code handoff → 第二账户登录 / preview / claim → permission 生效 → duplicate claim 拒绝 → 服务重启恢复。
- PostgreSQL configured Server 完成同构链、并发单胜者、停止 no-fallback 与 reconnect。
- 覆盖双标签、`1440×900`、`720×900`、`390×844`、console / network / URL / storage / cookie / database 隐私审计并回写真相源。

## 验收方式

- memory、SQLite、PostgreSQL 对 create / list / revoke / preview / claim、cursor `as_of`、CAS、并发和重启的语义一致。
- secret 原文只在 create response 与两个组件的短暂内存中出现；刷新、列表、日志、错误、audit、数据库 projection 和构建产物不能恢复。
- 一个 invitation 最多产生一个 membership 与一个 role assignment；并发、重放、catalog drift、expiry、revoke 或 repository failure 时不能部分写入。
- `workspace_admin`、未知 role、客户端 grants、任意 TTL、scope 覆盖和无 membership 的 Admin mutation 均失败关闭。
- claim 成功后的权限来自既有 membership / role owner，并在下一次业务请求立即生效；preview、code possession 或前端成功提示本身不构成授权。
- Web 的 selection / cursor / preview / code / confirmation / late response 失效、一次性交接、三视口和可访问状态表达通过。
- 真实产品链结束前清理本任务启动的服务、浏览器状态、临时账户、数据库和容器；命名持久 volume 仍按仓库既有策略处理。

## 停止线

- 不绑定或搜索 email、login identifier、display name、OIDC issuer / subject，不同步 Radish 或其它用户目录。
- 不发邮件、短信、站内信、webhook 或公开分享链接；邀请码不进入 URL。
- 不通过 invitation 创建账户、external identity、Session 或 API Key，不自动登录或自动选择 workspace。
- 不邀请 `workspace_admin`，不接受客户端 grants、自定义角色、角色组合、批量邀请、批量认领或自动角色推荐。
- 不把 preview、pending invitation、active workspace header 或上游 claim 当成 membership。
- 不延长、恢复、编辑或重复使用 invitation，不实现 background expiry worker、自动清理、delete 或永久 purge。
- 不引入 production auth / IAM、production invitation delivery、rate limit / abuse detection 声明、production secret 或真实 Radish 批次 E。
- 不新增第二套账户、membership、role、permission、Session 或审计 owner，不从已关闭成员管理专题派生 Batch F。

## 下一实现入口

[工作区成员邀请、认领与到期治理 v1 高风险任务卡](../../task-cards/workspace-member-invitation-claim-expiry-governance-dev-test-v1-plan.md)状态为 `workspace_member_invitation_claim_expiry_governance_dev_test_v1_batch_d_pencil_approved_batch_e_ready`。批次 D 已获项目所有者人工批准；下一步停在批次 E 独立授权线，未经再次明确授权不得修改 React、启动产品联调或实施 strict consumer 与双数据库产品验收。
