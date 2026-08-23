# 本地用户、角色与工作区成员管理（开发 / 测试态）v1

更新时间：2026-08-23

状态：`local_user_role_workspace_membership_administration_dev_test_v1_batch_a_completed_batch_b_ready`

## 功能定位

本功能在 RadishMind 已有本地账户、Web Session、角色分配与 `WorkspaceMembership` owner 之上，为平台管理员提供 workspace-scoped 的本地成员目录、内建角色目录、成员关系管理和角色分配管理。

它解决的是联合登录批次 D 之后的真实产品缺口：S7 User / Role 当前只能展示已登录账户及其 repository-owned grants，管理员无法在同一 workspace 中查看成员、授予或撤销 membership、分配或撤销内建角色。首版只管理 RadishMind 本地授权事实，不读取 Radish 用户目录，不同步 OIDC claim，也不建立第二套身份或权限真相源。

## 选择依据

- [本地账户与 Radish OIDC 联合登录 v1](local-account-radish-oidc-federated-login-v1.md)已经完成 `UserAccount`、`LocalRoleAssignment`、`WorkspaceMembership`、memory / SQLite / PostgreSQL repository 与本地 Session 授权链。
- 当前 `localIdentityRepository` 已有单账户读取、角色分配创建 / 撤销、成员关系创建 / 撤销和 CAS，但没有 workspace-scoped 列表、正式 Admin HTTP、管理权限与产品级 mutation service。
- S7 User / Role 明确显示当前账户，不伪造目录事实；因此下一步应在现有 owner 上建立真实目录与授权管理，而不是继续补只读 evidence 页面或依赖真实 Radish 批次 E。
- 本功能可以完全在本仓库的开发测试态三种 store 中验证，不依赖 production secret、真实 issuer、外部账号资源或上层项目接线。

## 已确认产品决策

- `UserAccount` 继续是平台级本地身份；目录默认按 exact tenant / workspace 的 membership 与 role facts 投影，不开放无作用域的全局账户枚举。
- 管理员添加成员时只提交 exact `user_id`。服务端验证目标账户 active，但公开失败不区分账户不存在或不可用，也不按 login identifier、email、display name、OIDC issuer / subject 搜索或合并。
- `WorkspaceMembership` 是访问 workspace 的唯一 owner；active workspace、角色分配或上游 claim 不能替代 membership。
- v1 只允许服务器拥有的内建角色目录。客户端提交 `role_key`，服务端从 canonical catalog 派生并冻结 `permission_grants`；客户端不能提交任意 grant 数组。
- 首批内建角色为 `workspace_reader`、`workspace_builder`、`workspace_reviewer` 与 `workspace_admin`。批次 A 必须在单一 catalog 中固定 exact grant matrix、catalog version 和 definition digest；每个 grant 必须是现有 canonical 业务权限或本专题显式引入的四项管理权限，禁止未声明字符串。
- 角色 assignment 继续保存不可变的排序后 grants。catalog 后续变化不原地重写历史 assignment；页面通过当前 catalog 与冻结 grants 的比较展示 drift。
- 撤销 workspace membership 必须在同一 repository transaction 中撤销该 membership 及其 workspace-scoped active role assignments，避免重新加入 workspace 时旧 grants 静默恢复。tenant-scoped assignment 不由单一 workspace revoke 隐式改写。
- 不允许当前 actor 撤销自己的当前 workspace membership；不得移除 workspace 最后一个仍具本地身份管理能力的 active administrator。
- 第一个 workspace administrator 不由“首个注册账户”、dev header、上游 claim 或默认配置自动产生。开发测试态只允许显式 one-shot CLI 使用 exact `tenant_ref + workspace_id + user_id` 建立 bootstrap membership 与 `workspace_admin` assignment；同一 scope 已存在任何 active identity administrator 时必须拒绝，且所有写入在单一事务中完成并带 audit ref。
- 账户禁用、凭证重置、session 全量撤销、external identity 管理继续由独立身份安全目标承接，不混入本功能。

## 目标用户与核心任务

目标用户是使用 RadishMind 本地 Web Session 的开发测试态平台管理员。

1. 在 exact tenant / workspace 下查看稳定分页的本地成员目录。
2. 打开一个成员，审查脱敏账户状态、membership、内建角色 assignment、有效期和 grant 摘要。
3. 使用 exact `user_id` 把已有 active 本地账户加入当前 workspace。
4. 从服务器内建角色目录选择角色并显式确认 assignment。
5. 以 expected version 撤销角色 assignment 或 workspace membership，并看到影响摘要。
6. 在 stale CAS、目标不可用、最后管理员保护或权限拒绝时得到稳定、可恢复且不泄漏身份事实的反馈。

## Owner 与数据边界

| 资源 | 现有 owner | 本功能允许的新增能力 | 明确不做 |
| --- | --- | --- | --- |
| `UserAccount` | RadishMind Identity | exact workspace member projection、active 状态校验 | 全局搜索、login identifier / email 投影、资料编辑、账户禁用 |
| `WorkspaceMembership` | RadishMind Authorization | workspace-scoped list、create、CAS revoke | 邀请、自动加入、跨 workspace 移动、批量 mutation |
| `LocalRoleAssignment` | RadishMind Authorization | workspace-scoped list、catalog-derived create、CAS revoke | 客户端任意 grants、自定义角色、自动 claim mapping |
| 内建角色目录 | RadishMind Authorization policy | versioned static catalog 与 digest | 数据库角色表、UI 自定义、production policy service |
| `WebSession` | RadishMind Session | 恢复管理员 actor 并重新读取 membership / permission | 管理其他用户 session、token projection、dev header fallback |
| first-admin bootstrap | RadishMind Authorization operation | 显式开发测试态 one-shot CLI、exact scope / user、单事务 membership + admin assignment | HTTP bootstrap、首用户自动提权、启动时隐式写入、已有管理员时重复 bootstrap |
| Radish OIDC | Radish Integration | 无新增能力 | 用户目录同步、email 合并、role / permission claim 授权 |

列表与详情不得返回 login identifier、credential、password policy material、session、cookie、issuer、subject、raw claim、token、email、audit body 或完整错误正文。允许返回 `user_id`、display name、账户 lifecycle、membership / assignment stable id、版本、状态、有效期、role key、排序后 permission grant 名称和脱敏时间。

## 内建角色目录

角色目录是代码内 canonical policy，不是新的数据库 owner。每个定义至少包含：

- `catalog_version`
- `role_key`
- `display_name`
- `summary`
- `permission_grants`
- `definition_digest`
- `can_manage_local_identity`

所有 grants 使用已有 canonical permission 名称。Batch A 必须一次性枚举并测试初始四个角色的 exact grant matrix；不能在 handler、Web fixture 或测试中重复维护第二份 grants。`workspace_admin` 是首版唯一可包含本功能四项管理权限的角色，最后管理员保护以 `can_manage_local_identity` 与有效 assignment 决策，不以 display name 或前端标签判断。

批次 A 已将目录冻结为 `local_identity_builtin_roles_v1`，catalog digest 为 `sha256:d784ef5d5595f4fa3ed96f32c86f3fd12edbd4098da46668366f97ce42e2d4d0`。四角色按 reader → builder → reviewer → administrator 累积既有 canonical grants；`workspace_admin` 覆盖完整 workspace permission allowlist，并且只有它包含四项本地身份管理权限。catalog、角色 definition digest、allowlist 完整覆盖和不可变复制均由 Go 测试固定；后续 grant 变化必须显式评估 catalog version，而不能静默改写历史 assignment。

## 目录、详情与 cursor

成员目录只列出 exact `tenant_ref + workspace_id` 下存在 membership 历史的本地账户，不提供全局未分配账户列表。首版固定：

- 默认只返回 active membership；显式 `membership_state=revoked` 才审查历史。
- 排序为 `updated_at DESC, membership_id DESC`。
- cursor 绑定 tenant、workspace、membership state 和 limit；改变任一 filter 必须拒绝旧 cursor。
- limit 默认为 `50`，最大 `100`；同一时间戳必须由 `membership_id` 保证稳定翻页。
- exact member detail 必须重新验证目标属于当前 tenant / workspace 的 membership 历史；无权、缺失或跨 scope 使用同一脱敏失败。

目录只投影既有记录，不建立重复的 member current-state 表或 summary store。

## Admin HTTP 与权限

首版计划使用以下开发测试态 HTTP surface：

| 方法与路径 | 权限 | 语义 |
| --- | --- | --- |
| `GET /v1/admin/local-identity/workspaces/{workspace_id}/members` | `local_identity_members:read` | workspace-scoped 稳定分页目录 |
| `GET /v1/admin/local-identity/workspaces/{workspace_id}/members/{user_id}` | `local_identity_members:read` | exact member detail |
| `GET /v1/admin/local-identity/role-catalog` | `local_identity_roles:read` | 当前内建角色目录 |
| `POST /v1/admin/local-identity/workspaces/{workspace_id}/memberships` | `local_identity_memberships:write` | exact `user_id` 加入当前 workspace |
| `POST /v1/admin/local-identity/workspaces/{workspace_id}/memberships/{membership_id}/revoke` | `local_identity_memberships:write` | CAS revoke，并原子撤销 workspace-scoped assignments |
| `POST /v1/admin/local-identity/workspaces/{workspace_id}/role-assignments` | `local_identity_roles:assign` | 从 canonical catalog 创建冻结 assignment |
| `POST /v1/admin/local-identity/workspaces/{workspace_id}/role-assignments/{assignment_id}/revoke` | `local_identity_roles:assign` | assignment CAS revoke |

所有 mutation 同时要求同源 Origin、CSRF、近期认证、request / audit ref、expected version（适用时）和显式影响确认字段。tenant 来自已验证 actor context；path workspace 必须与 active workspace 和 membership decision 一致，客户端不能覆盖 tenant 或通过 payload 改 scope。

first-admin bootstrap 不注册 HTTP route。它只由批次 B 的显式开发测试态 CLI 调用批次 A 的同一领域 service，并要求 exact 账户已存在且 active、目标 scope 没有 active identity administrator、repository mode 明确为 SQLite 或 PostgreSQL dev/test。memory 只用于领域测试；生产模式、服务启动和普通注册流程不得触发 bootstrap。

## 授权与原子性

每次 Admin 请求按以下顺序失败关闭：

1. 验证本地 Web Session 并恢复 exact `user_id`。
2. 重读 actor account active 状态。
3. 重读 actor 在 exact tenant / workspace 的 active membership。
4. 重读本功能所需 exact permission。
5. 校验 path、query、payload、CSRF / Origin、近期认证和确认字段。
6. 在单一 local identity repository transaction 中读取并变更目标记录。

目标账户、membership、assignment、catalog 或 repository 任一失败不得回退 `dev_headers`、signed-test token、resource-server Bearer token、离线 fixture 或上次成功结果。mutation 成功后由后端返回最新 canonical projection，Web 不自行推算版本或 grants。

## 稳定失败分类

- `local_identity_admin_unavailable`
- `local_identity_admin_scope_mismatch`
- `local_identity_member_unavailable`
- `local_identity_member_cursor_invalid`
- `local_identity_role_catalog_mismatch`
- `local_identity_membership_conflict`
- `local_identity_role_assignment_conflict`
- `local_identity_self_membership_revoke_denied`
- `local_identity_last_admin_removal_denied`
- `local_identity_recent_authentication_required`
- `local_identity_admin_bootstrap_denied`
- `workspace_membership_denied`
- `workspace_permission_denied`

公开错误只返回 stable failure code、request ref、当前允许公开的版本 / 状态和恢复方向；不证明未授权目标是否存在，也不返回 SQL、repository、credential 或 raw auth details。

## Web 与 Family UI

- S7 User 从“当前账户 owner”升级为 workspace member directory + selected member detail；默认 dominant region 是成员列表，详情为从属 inspector。
- S7 Role 展示同一 selected member 的 assignment、内建 role catalog 与显式 create / revoke confirmation；不建立独立角色数据库页面。
- 当前账户与 external identity 管理继续留在 Authentication / Account surface，不复制到目录详情。
- Desktop 与 Narrow 必须继承现有 Family UI 的 `Inter + Geist Mono`、冷灰白工作区、深蓝主操作、语义 token、紧凑信息密度与既有 S7 shell。
- 页面覆盖 loading、empty、denied、unavailable、stale conflict、catalog drift、last-admin protection、mutation success 与 membership revoked 状态。
- 用户切换 workspace、登出、scope 变化或 mutation 成功后，旧目录 cursor、selection、dirty form、confirmation 和迟到响应全部失效。

五维初评为 `2 / 1 / 2 / 2 / 2 = 9`，采用 `A / 完整 Pencil`。设计只修改 S7 User / Role 代表面及对应 Decision Record，不建立 S11，也不改变其它页面 owner。

## 实施拆分

### 批次 A：领域合同、角色目录与 memory 纵向链

- 已固定 summary / detail / role catalog / filter-bound cursor 合同、exact grant matrix、catalog version 和 digest；legacy assignment 通过 catalog metadata 比较显式投影 drift。
- 已在同一 memory identity owner 上增加独立 administration capability，完成 workspace-scoped list / exact read、catalog-derived assignment、原子 membership + workspace assignment revoke，以及重新加入不恢复旧 grants。
- 已建立管理 service、十分钟 recent-auth 输入、CAS、并发单胜者、显式 one-shot first-admin bootstrap、self / last-admin protection，并在 repository 锁内重读 actor 权限以关闭授权检查与写入之间的竞态。
- 四项管理权限已进入 canonical workspace permission allowlist，但通用 role assignment 入口明确拒绝这些 grants；在批次 B durable canonical assignment 落地前，SQLite / PostgreSQL 旧入口不能取得本地身份管理能力。
- memory 证据覆盖 `121` 条同时间戳三页稳定分页、cursor filter / limit / tamper 拒绝、脱敏详情、catalog drift、重复与 stale CAS、并发单胜者、原子 revoke、bootstrap 单胜者、权限即时失效和零副作用负向路径。
- 本批未注册 HTTP、未修改 config / migration / durable administration owner，也未修改 Pencil 或 Web。

### 批次 B：SQLite / PostgreSQL durable owner

- 复用现有 local identity tables；按实际 query plan 增加下一顺序索引 migration，不创建重复目录表。
- 实现稳定分页、exact detail、catalog-derived assignment 与原子 membership revoke。
- 增加显式开发测试态 bootstrap CLI；只接受 exact scope / user、已有管理员时拒绝，不在 Server 启动或注册时自动执行。
- 覆盖 migration / rollback / reapply、受限 runtime role、并发、重启和 no-fallback。

### 批次 C：Admin HTTP 与 local session 授权

- 注册七条 strict route、四项独立权限、CSRF / Origin、近期认证、确认和稳定失败映射。
- 覆盖非成员、权限不足、跨 workspace、cursor 漂移、目标枚举、stale CAS、self / last-admin protection 和零业务副作用。
- 只允许显式开发测试态 local session，不增加 dev header / signed-test fallback。

### 批次 D：Pencil 与 React strict consumer

- 在既有 Family UI 中完成 S7 User / Role Desktop、Narrow 和 Decision Record。
- 建立单一 strict consumer、workspace member directory、detail、role catalog 和显式 mutation flow。
- Web 不缓存身份目录到 URL、Web Storage、IndexedDB 或 service worker。

### 批次 E：双数据库产品连续链与专题收口

- SQLite 产品链先完成第一账户注册 → 显式 bootstrap CLI 建立首个 workspace administrator，再由页面完成第二账户注册 → exact user id 交接 → membership create → role assign → 业务权限生效 → role revoke / membership revoke → 立即失败关闭 → 服务重启恢复。
- PostgreSQL 配置化 Server 完成同构 repository / HTTP 连续链、关闭后 no-fallback 与重连恢复。
- 完成三视口、双标签、隐私、console / network / storage 审计和最终仓库门禁。

## 验收方式

- memory、SQLite、PostgreSQL 的 stable cursor、同时间戳分页、filter-bound cursor、CAS、并发单胜者、原子 membership revoke 与重启恢复。
- first-admin bootstrap 只在 exact scope 无 active identity administrator 时单次成功；重复、并发、inactive / missing account、错误 store mode 或部分写入均失败关闭。
- non-member、permission denied、scope mismatch、expired membership、inactive account、catalog mismatch、stale version、self revoke 和 last-admin protection 均在 repository mutation 前或事务内失败关闭。
- assignment 只能由 canonical catalog 派生；客户端注入 grants、未知 role、重复 assignment 和过期 role definition 均被拒绝。
- membership / role 变化在下一次业务请求立即生效；无 cache、claim 或旧 session grant fallback。
- Web 的 exact selection、确认摘要、错误恢复、workspace 切换、跨标签失效、Desktop / Narrow 与无敏感字段审计通过。

## 停止线

- 不实现全局账户枚举、模糊搜索、email / login identifier 搜索或 OIDC 用户目录同步。
- 不实现邀请、自动准入、批量导入、批量授权、跨 workspace 移动或自动角色推荐。
- 不实现自定义角色、客户端任意 permission grants、role hierarchy、deny rule 或条件策略语言。
- 不实现账户禁用、删除、密码重置、MFA、恢复、session 管理或 external identity 管理。
- 不提供 unauthenticated HTTP bootstrap、不让首个注册账户自动成为管理员、不在 Server 启动时读取环境变量隐式提权。
- 不把开发测试态目录、SQLite / PostgreSQL 证据或本地角色目录解释为 production IAM、production OIDC 或合规完成。
- 不新增第二套 membership、permission、audit 或用户目录 owner，不从本专题读取 Radish 数据库或上游 role / permission claim。

## 下一实现入口

[本地用户、角色与工作区成员管理 v1 高风险任务卡](../../task-cards/local-user-role-workspace-membership-administration-dev-test-v1-plan.md)承接批次 A 至 E。批次 A 已完成，下一入口为批次 B 的 SQLite / PostgreSQL durable administration owner 与显式开发测试态 bootstrap CLI；批次 C 至 E 必须继续依次消费前一批证据，不并行扩张 HTTP 与 UI 边界。
