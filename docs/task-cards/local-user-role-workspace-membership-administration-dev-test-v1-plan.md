# 本地用户、角色与工作区成员管理 v1 高风险任务卡

更新时间：2026-08-23

状态：`local_user_role_workspace_membership_administration_dev_test_v1_completed`

对应功能设计：[本地用户、角色与工作区成员管理（开发 / 测试态）v1](../features/admin-control-plane/local-user-role-workspace-membership-administration-dev-test-v1.md)

## 任务目标

在既有 RadishMind 本地账户、Web Session、角色分配和 workspace membership owner 上，建立 workspace-scoped 成员目录、内建角色目录、受控 membership / role assignment mutation，以及 S7 User / Role 产品连续链。所有授权继续由本地 membership 与 frozen grants 决策，不读取 Radish 目录或 OIDC claim 形成 grant。

## 已确认前提

- 联合登录专题批次 A 至 D 已完成；本任务不等待真实 Radish 批次 E。
- `localIdentityRepository` 是唯一 identity / authorization repository；不创建平行用户目录、角色表或 membership store。
- 首版只接受 exact `user_id`、exact tenant / workspace 和服务器内建 `role_key`。
- 客户端不能提交任意 permission grants；assignment grants 由 canonical catalog 派生并冻结。
- 首个 workspace administrator 只通过显式开发测试态 one-shot CLI 建立；不按首个注册账户自动提权，不接受 dev header / OIDC claim，已有 active identity administrator 的 scope 必须拒绝重复 bootstrap。
- 账户禁用、凭证、session、external identity、自定义角色、邀请与生产 IAM 不在本任务范围。

## 批次 A：领域合同、角色目录与 memory 纵向链

实施范围：

- 定义 `local_identity_workspace_member_summary.v1`、exact member detail、role catalog 与 filter-bound cursor。
- 在单一 Go catalog 固定 `workspace_reader`、`workspace_builder`、`workspace_reviewer`、`workspace_admin` 的 exact grants、catalog version、definition digest 和 identity-management capability。
- 扩展 repository contract：workspace-scoped member list、exact detail、catalog-derived role assignment，以及原子 membership + workspace assignment revoke。
- 建立 service 层的 scope、CAS、recent-auth input、one-shot first-admin bootstrap、self-membership revoke denial 和 last-admin protection。
- memory 覆盖超过 `100` 条同时间戳分页、cursor filter 绑定、重复 membership / role、stale CAS、并发单胜者、catalog drift、bootstrap 单胜者和零副作用。

批次 A 停止线：

- 不注册 HTTP route，不修改 config，不增加数据库 migration，不改 Web 或 Pencil。
- 不把 permission matrix 散落到 handler、fixture 或前端；只有 canonical catalog 是 grants 来源。

完成证据：

- `local_identity_builtin_roles_v1` 已作为首版冻结；2026-09-01 为恢复既有 Publish Candidate / Prompt Runtime 只读 owner 的最小角色可达性，目录显式推进为 `local_identity_builtin_roles_v2`，固定 digest 为 `sha256:44c8a3a41eb90b2da25859662abf13ba91cef00505eb5191767dc5b17eb4abae`。只新增 `application_publish_candidates:read` 与 `prompt_application_runtime:read`，历史 v1 assignment 不被改写；四角色 exact grants、完整 allowlist 覆盖、唯一 identity-management capability 和 immutable copy 均有测试。
- summary / detail / filter-bound cursor、memory administration repository / service、catalog-derived assignment、原子 membership revoke、十分钟 recent-auth、CAS、self / last-admin protection 与 one-shot bootstrap 已落地。
- memory 覆盖 `121` 条同时间戳三页、cursor 篡改、catalog drift、重新加入不恢复 grants、并发单胜者、bootstrap 单胜者、repository 锁内授权重读和稳定失败码；完整 `internal/httpapi/...` 与 PostgreSQL tagged compile 已通过。
- 停止线保持：没有 HTTP / config / migration / durable administration owner / Pencil / Web 改动；通用 assignment 入口拒绝四项管理权限，避免在批次 B 前绕开 canonical durable 写入。

## 批次 B：SQLite / PostgreSQL durable owner

实施范围：

- 为现有 local identity repository 实现与 memory 一致的列表、详情和 aggregate mutation。
- 先审查 query plan；只有稳定分页确需时才增加下一顺序读取索引，不创建重复目录表或 role catalog 表。
- 增加显式开发测试态 bootstrap CLI：exact `tenant_ref + workspace_id + user_id`、active account、零 active identity administrator、单事务 membership + `workspace_admin` assignment 与 audit ref；不在 Server 启动或注册时自动执行。
- SQLite / PostgreSQL 覆盖 migration、rollback / reapply、受限 runtime role、cursor、CAS、事务原子性、重启和 no-fallback。

批次 B 停止线：

- 不自动迁移生产数据库，不引入新数据库或 ORM，不打开 production store mode。

完成证据：

- SQLite / PostgreSQL durable administration owner 已复用现有 local identity tables，实现直接 SQL 目录 / 详情、catalog metadata 持久化、CAS、scope 内授权重读和 membership + workspace assignment 单事务撤销；没有新增目录表或角色表。
- `0003_local_identity_administration` / `local_identity_records_store_v3` 只增加两列 catalog metadata 与稳定分页顺序索引。SQLite v2 → v3 重放和 PostgreSQL v1 → v3、rollback / reapply、受限 runtime role、重启与 no-fallback 已验证。
- query-plan 证据在 `121` 条同时间戳数据上成立：SQLite `EXPLAIN QUERY PLAN` 与 PostgreSQL `ANALYZE + EXPLAIN` 均使用 `local_workspace_memberships_directory_idx`。
- `radishmind-local-identity-bootstrap` 只允许 `sqlite_dev | postgres_dev_test`，从环境变量读取数据库位置，并把 exact tenant / workspace / active user / audit ref 交给同一领域 service；并发和重复 bootstrap 只有一次成功，memory、缺失数据库、已有管理员和部分写入均失败关闭。
- 批次 B 未注册 HTTP、未修改 config / Server startup、Pencil 或 Web，production store / IAM 与自动 bootstrap 继续关闭。

## 批次 C：Admin HTTP 与 local session 授权

实施范围：

- 注册功能设计固定的七条 strict Admin route。
- 增加 `local_identity_members:read`、`local_identity_memberships:write`、`local_identity_roles:read`、`local_identity_roles:assign` 四项 exact permission。
- mutation 要求 local Web Session、membership 重读、exact permission、同源 Origin、CSRF、近期认证、显式确认、expected version 与 request / audit ref。
- 非成员、跨 scope、权限拒绝、目标不可用、未知 role、客户端 grants 注入、stale CAS、self / last-admin denial 都返回稳定脱敏失败并保持零业务副作用。

批次 C 停止线：

- 不允许 `dev_headers`、signed-test token 或 resource-server Bearer token 成为 local session 失败 fallback。
- 不注册 first-admin bootstrap HTTP route；普通 Admin route 不能绕过 active membership 与 exact permission。
- 不增加账户禁用、密码重置、session 管理、external identity 或邀请 API。

完成证据：

- 七条 fixed Admin route 已由 `Server` 注册并复用同一 administration service；read 与 mutation 分别重读 exact `local_identity_members:read`、`local_identity_memberships:write`、`local_identity_roles:read`、`local_identity_roles:assign`，没有通用“管理员”布尔旁路。
- 认证只接受显式 `local_session_dev_test` Web Session。七条请求都必须各提供一份 `X-RadishMind-Active-Tenant` 与 `X-RadishMind-Active-Workspace`：tenant selection 必须与 middleware 恢复的 verified actor tenant 相等，workspace selection 必须与 path workspace 相等；Bearer、dev header、signed-test、缺失 session、scope / membership / permission / repository failure 都不能 fallback。
- mutation 复用 exact Origin、double-submit CSRF、十分钟近期认证、显式 `confirmed`、expected catalog / record version，并从 request id 生成有界摘要 audit ref；request body 拒绝未知 / 重复字段、客户端 `permission_grants`、scope 覆盖与多 JSON 文档。
- 自动化覆盖七条成功链、strict query / method、非成员、单权限隔离、跨 workspace、目标不可用、unknown role、grants 注入、missing confirmation、stale CAS、self / last-admin denial、CSRF / Origin、stale authentication、脱敏 projection 与稳定 recovery metadata；first-admin bootstrap HTTP 保持 `404`。
- 精准管理 HTTP / service 测试、完整 `internal/httpapi`、Platform `go test ./...` 与管理专项 race 已通过；本批没有修改 config、migration、Pencil 或 Web，也没有启动 PostgreSQL 容器或宣称 production IAM。

## 批次 D：Pencil 与 React strict consumer

实施范围：

- 五维评分固定为 `2 / 1 / 2 / 2 / 2 = 9`，采用 `A / 完整 Pencil`。
- 只更新既有 S7 User / Role Desktop、Narrow 与对应 Decision Record；复用 Family UI，不创建 S11。
- 建立单一 strict consumer、workspace member directory、selected member detail、role catalog 和受控 create / revoke flow。
- 覆盖 loading、empty、denied、unavailable、stale conflict、catalog drift、last-admin protection、success 与 revoked 状态。

批次 D 停止线：

- 不在 URL、Web Storage、IndexedDB、service worker、日志或截图持久化目录载荷、确认正文或身份敏感字段。

完成证据：

- 复用提交 `72205455` 已批准的 `docs/designs/radishmind-web-family-ui-v1.pen`，Desktop `fFTsy` / `jEmjK`、Narrow `Wrggq` / `LQ297`、Decision `bkvt3` 未重新设计。
- 单一 React strict consumer 已接入七条 Admin API；S7 User 提供 active / revoked 成员目录与 exact detail，S7 Role 复用同一 selected member、canonical 四角色目录及受控 membership / role assignment create / revoke。
- mutation 只在内存中建立候选并显式确认；成功、workspace / session / actor scope 变化会失效 cursor、selection、dirty form、confirmation 与迟到响应。客户端不提交 `permission_grants`，不推算 canonical grants 或 CAS version。
- 自动化覆盖七条 route 的 exact method / path / header / body、cookie session、CSRF、confirmation、CAS、scope / unknown / forbidden field 拒绝、稳定 recovery 分类和请求前输入拒绝；Web 全套测试与 production build 已通过。
- 隐私静态审查未发现 URL、Web Storage、IndexedDB 或 service worker 持久化目录载荷与确认内容；本批未执行批次 E 双数据库产品连续链、真实 Radish 或 production IAM。

## 批次 E：双数据库产品连续链与收口

实施范围：

- SQLite 产品链完成第一账户注册、显式 bootstrap CLI、管理员登录，再由页面完成第二账户注册、exact `user_id` 交接、membership create、role assign、业务 permission 生效、role / membership revoke、立即失败关闭与服务重启恢复。
- PostgreSQL 配置化 Server 完成同构连续链、关闭后 no-fallback 与重连恢复。
- 完成 `1440×900`、`720×900`、`390×844`、双标签、console / network / URL / storage / privacy 审计。
- 回写功能专题、Admin 入口、当前焦点、路线图、能力矩阵和周志后关闭专题。

完成证据：

- SQLite 页面链完成两账户注册、显式 bootstrap、exact `user_id` membership create、canonical `workspace_reader` assign / revoke、业务权限 `200 → workspace_permission_denied`、membership revoke 后 `workspace_membership_denied`、原子 assignment revoke、Server no-fallback 与同库重启恢复。
- PostgreSQL 17 通过独立临时容器、显式 `0003` migration、受限 runtime role 和 configured Server 完成同构链、关闭 no-fallback 与重连恢复；临时容器、网络和 volume 已删除。
- 实链路发现并修复 strict consumer 缺少 `X-RadishMind-Active-Tenant` 的 scope header；七条 API 现统一发送 exact tenant / workspace，精准 consumer 测试固定该合同，服务端仍保持 scope mismatch fail-closed。
- User 页面在 `1440×900`、`720×900`、`390×844`，Role 页面在 `390×844` 均无横向溢出；双标签 logout 同步失效。URL query、Web Storage、IndexedDB、Cache Storage 与 service worker 均未保存身份目录或确认内容，登录态 session credential 仅存在 HttpOnly + SameSite=Strict cookie。
- 预认证、近期认证、权限撤销与服务关闭产生的预期 `401 / 403 / unavailable` 已逐项核对；临时账号、数据库、浏览器快照 / 日志与所有本地服务均已清理。

## 必须保持的负向边界

- 无全局账户搜索、email / login identifier 查询、上游目录同步或 email 自动合并。
- 无客户端 grants、自定义角色、role hierarchy、批量授权、自动准入或自动推荐。
- 无 self membership revoke、最后管理员移除或跨 workspace mutation。
- 无 unauthenticated HTTP bootstrap、首用户自动提权、Server 启动时隐式 bootstrap 或已有管理员时重复 bootstrap。
- 无 identity / membership / permission / repository failure fallback。
- 无 production auth、production IAM、production database、production secret 或真实 Radish 声明。

## 验证入口

每批先运行精准测试，再按风险扩展：

```bash
(cd services/platform && go test ./internal/httpapi/...)
npm --prefix apps/radishmind-web test
npm --prefix apps/radishmind-web run build
./scripts/check-repo.sh --fast
```

批次 B 起涉及 repository / migration / auth / schema / 当前阶段真相，提交前补跑完整：

```bash
./scripts/check-repo.sh
```

## 当前完成条件

- [x] 新长期功能专题、owner、流程、失败语义、批次与停止线已定义。
- [x] 唯一高风险任务卡已建立，批次 A 可进入实现。
- [x] 批次 A：领域合同、canonical role catalog 与 memory 纵向链。
- [x] 批次 B：SQLite / PostgreSQL durable owner。
- [x] 批次 C：Admin HTTP 与 local session 授权。
- [x] 批次 D：Pencil 与 React strict consumer。
- [x] 批次 E：双数据库产品连续链与专题收口。
