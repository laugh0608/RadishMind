# 本地用户、角色与工作区成员管理（开发 / 测试态）v1

状态：`local_user_role_workspace_membership_administration_dev_test_v1_completed`

更新时间：2026-08-23

## 功能定位

本功能在 RadishMind 已有本地账户、Web Session、角色分配与 `WorkspaceMembership` owner 之上，为平台管理员提供 workspace-scoped 的本地成员目录、内建角色目录、成员关系管理和角色分配管理。

它解决的是联合登录批次 D 之后的真实产品缺口：S7 User / Role 已从仅展示当前账户升级为 exact workspace 的成员目录、详情、内建角色目录与受控 membership / role assignment create / revoke。首版只管理 RadishMind 本地授权事实，不读取 Radish 用户目录，不同步 OIDC claim，也不建立第二套身份或权限真相源。

## 选择依据

- [本地账户与 Radish OIDC 联合登录 v1](local-account-radish-oidc-federated-login-v1.md)已经完成 `UserAccount`、`LocalRoleAssignment`、`WorkspaceMembership`、memory / SQLite / PostgreSQL repository 与本地 Session 授权链。
- 当前 `localIdentityRepository` 已有单账户读取、角色分配创建 / 撤销、成员关系创建 / 撤销和 CAS；批次 A 至 D 已补齐三种 store 的 workspace-scoped 管理 repository、产品级 mutation service、正式 Admin HTTP 与 S7 strict consumer。
- S7 User / Role 只消费 repository-owned canonical 目录、详情、角色与 mutation projection，不伪造目录事实，也不依赖真实 Radish 批次 E。
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

首版已注册以下开发测试态 HTTP surface：

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

strict request body 已固定：membership create 为 `user_id + expires_at? + confirmed`；membership / assignment revoke 为 `expected_record_version + confirmed`；role assignment create 为 `user_id + role_key + expected_catalog_version + expected_role_definition_digest + expires_at? + confirmed`。客户端提交 `tenant_ref`、`workspace_id`、`permission_grants`、重复字段、未知字段或多份 JSON 文档均被拒绝。成功响应统一携带 `request_id + tenant_ref + workspace_id`，再返回当前 endpoint 的 canonical member、catalog、membership 或 assignment 管理投影；不返回领域记录中的 audit ref、登录标识、credential 或 session。

first-admin bootstrap 不注册 HTTP route。`radishmind-local-identity-bootstrap` 只在开发者显式执行时调用同一领域 service，并要求 exact 账户已存在且 active、目标 scope 没有 active identity administrator、repository mode 明确为 `sqlite_dev` 或 `postgres_dev_test`。数据库位置只从对应环境变量读取，不进入 argv 或 JSON 输出；memory、生产模式、服务启动和普通注册流程不得触发 bootstrap。

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

批次 D 复用提交 `72205455` 已批准的设计源 `docs/designs/radishmind-web-family-ui-v1.pen`：Desktop `fFTsy` / `jEmjK`、Narrow `Wrggq` / `LQ297`、Decision `bkvt3`。React 使用单一 strict consumer 接入七条 Admin API；目录保持 dominant、详情为 subordinate inspector，Role 复用同一 selected member 与服务器 canonical 四角色目录。create / revoke 均先形成内存候选再显式确认，成功后失效旧 selection、cursor、表单、确认和迟到响应；目录载荷与确认内容不写入 URL、Web Storage、IndexedDB 或 service worker。

## 实施拆分

### 批次 A：领域合同、角色目录与 memory 纵向链

- 已固定 summary / detail / role catalog / filter-bound cursor 合同、exact grant matrix、catalog version 和 digest；legacy assignment 通过 catalog metadata 比较显式投影 drift。
- 已在同一 memory identity owner 上增加独立 administration capability，完成 workspace-scoped list / exact read、catalog-derived assignment、原子 membership + workspace assignment revoke，以及重新加入不恢复旧 grants。
- 已建立管理 service、十分钟 recent-auth 输入、CAS、并发单胜者、显式 one-shot first-admin bootstrap、self / last-admin protection，并在 repository 锁内重读 actor 权限以关闭授权检查与写入之间的竞态。
- 四项管理权限已进入 canonical workspace permission allowlist，但通用 role assignment 入口明确拒绝这些 grants；在批次 B durable canonical assignment 落地前，SQLite / PostgreSQL 旧入口不能取得本地身份管理能力。
- memory 证据覆盖 `121` 条同时间戳三页稳定分页、cursor filter / limit / tamper 拒绝、脱敏详情、catalog drift、重复与 stale CAS、并发单胜者、原子 revoke、bootstrap 单胜者、权限即时失效和零副作用负向路径。
- 本批未注册 HTTP、未修改 config / migration / durable administration owner，也未修改 Pencil 或 Web。

### 批次 B：SQLite / PostgreSQL durable owner

- 已复用现有 local identity tables，并把 PostgreSQL marker 推进为 `0003_local_identity_administration` / `local_identity_records_store_v3`。迁移只为 assignment 增加 nullable catalog version / definition digest，并增加 `(tenant_ref, workspace_id, lifecycle_state, updated_at DESC, membership_id DESC)` 顺序索引；没有创建重复目录表或 role catalog 表。
- SQLite / PostgreSQL 已实现直接 SQL 稳定分页与 exact detail；写入在单一事务中载入 exact scope、复用批次 A 的同一 invariant owner，再持久化 catalog-derived assignment、CAS revoke 和 membership + workspace assignment 原子撤销。SQLite 使用 `BEGIN IMMEDIATE` 串行化写者，PostgreSQL 使用 scope-bound transaction advisory lock；通用持久化 mutation 同样不能绕过 catalog aggregate 与 scope coordination。
- `radishmind-local-identity-bootstrap` 已支持 `sqlite_dev | postgres_dev_test`，只接受命令行中的 exact tenant / workspace / user / audit ref，数据库路径或 URL 只从环境变量读取；active account、零 active identity administrator、membership + `workspace_admin` 同事务和重复拒绝均由同一 repository service 保证。Server 启动与注册流程未接入该命令。
- SQLite 覆盖 v2 → v3 升级、迁移重放、`121` 条同时间戳分页、cursor、catalog metadata、并发 CAS、原子撤销与重启；PostgreSQL 17 覆盖 v1 → v3、真实 `ANALYZE + EXPLAIN` 索引计划、受限 runtime role、显式 bootstrap、回滚后 no-fallback、重新应用与重启。测试容器和网络已关闭。
- 停止线保持：没有注册 Admin HTTP、没有修改 config / Server startup、Pencil 或 Web，也没有自动迁移生产数据库、引入新数据库 / ORM 或打开 production store mode。

### 批次 C：Admin HTTP 与 local session 授权

- 已注册七条 strict route；Server 只从现有 `local_session_dev_test` middleware 恢复 exact `user_id + tenant`，要求单值 active workspace 与 path 一致，并在解析 query / payload 与执行 mutation 前重读 active account、membership 和 endpoint 对应的 exact permission。
- mutation 已复用同源 Origin、double-submit CSRF、十分钟近期认证、显式 `confirmed`、catalog / record expected version 和摘要化 request-bound audit ref；成功只返回 canonical 管理投影，稳定失败附带脱敏 `recovery` metadata。
- HTTP 自动化已覆盖七条成功链、非成员、单权限隔离、跨 workspace、目标不可用、未知 role、客户端 grants 注入、cursor / filter、stale CAS、self / last-admin protection、缺失确认、CSRF / Origin、近期认证、错误 method 和零业务副作用。
- Bearer、dev header、signed-test 与缺失 / 无效 local session 均不能 fallback；first-admin bootstrap 没有 HTTP route，Server startup 和普通注册也未接入 bootstrap。

### 批次 D：Pencil 与 React strict consumer

- 已在既有 Family UI 中完成 S7 User / Role Desktop、Narrow 和 Decision Record；批准节点为 Desktop `fFTsy` / `jEmjK`、Narrow `Wrggq` / `LQ297`、Decision `bkvt3`，本批不重新设计。
- 已建立单一 strict consumer，接入 workspace member directory、exact detail、role catalog 和四条显式 mutation flow；客户端只提交 canonical candidate，不提交任意 grants，也不推算服务端版本或权限。
- 已覆盖 loading、empty、denied、unavailable、stale conflict、catalog drift、last-admin protection、success 与 revoked，并以 authorization key、generation 与 abort 清除 scope 变化和 mutation 成功后的旧状态 / 迟到响应。
- Web 不缓存身份目录或确认内容到 URL、Web Storage、IndexedDB 或 service worker；当前批次仅完成自动化 Web 测试与 production build，不执行批次 E 的真实双数据库页面连续链。

### 批次 E：双数据库产品连续链与专题收口

- SQLite 产品链已完成第一账户注册 → 显式 bootstrap CLI 建立首个 workspace administrator → 页面注册第二账户 → exact `user_id` 交接 → membership create → `workspace_reader` assign → 业务读取 `200` → role revoke 后 `workspace_permission_denied` → 重新分配 → membership revoke 后 `workspace_membership_denied` → 服务停止 no-fallback → 同库重启恢复 revoked membership 与 assignment 历史。
- PostgreSQL 17 已由 migration identity 显式应用 `0003_local_identity_administration`，再以受限 runtime identity 启动 configured Server 并完成同构注册、bootstrap、目录 / 详情、角色分配、业务权限、role / membership revoke、立即失败关闭、Server 停止 no-fallback 与同库重连恢复。独立临时容器、网络和 volume 已删除。
- 真实页面链发现 strict consumer 缺少 `X-RadishMind-Active-Tenant`，导致已 bootstrap 的 local session 被服务端以 `local_identity_admin_scope_mismatch` 失败关闭；七条 Admin API 现统一发送 exact tenant / workspace header，并由 consumer 自动化固定，不放宽服务端 scope 校验。
- `1440×900`、`720×900`、`390×844` 的 User 页面及 `390×844` Role 页面均无横向溢出；双标签 logout 立即回到认证入口。浏览器审计确认 location 只保留静态 owner hash，query 为空，Local / Session Storage、IndexedDB、Cache Storage 和 service worker 均为空；登录态只有 SameSite=Strict 的 CSRF cookie 与 HttpOnly session cookie，未保存目录载荷或确认正文。
- 预认证 `401`、近期认证拒绝、显式业务权限 `403` 和 Server 停止时的 service-unavailable console / network 记录均与预期 fail-closed 状态对应；测试账号、SQLite 数据库、Playwright 快照 / 日志、服务、浏览器和 PostgreSQL 临时资源已清理。

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

[本地用户、角色与工作区成员管理 v1 高风险任务卡](../../task-cards/local-user-role-workspace-membership-administration-dev-test-v1-plan.md)承接的批次 A 至 E 已全部完成，专题关闭。下一产品入口回到[功能设计文档入口](../README.md)选择新的长期功能目标；不得从本专题派生批次 F、真实 Radish、production IAM、全局账户搜索、自定义角色或客户端 grants。
