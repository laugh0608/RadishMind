# 本地用户、角色与工作区成员管理 v1 高风险任务卡

更新时间：2026-08-23

状态：`local_user_role_workspace_membership_administration_dev_test_v1_design_defined_batch_a_ready`

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

## 批次 B：SQLite / PostgreSQL durable owner

实施范围：

- 为现有 local identity repository 实现与 memory 一致的列表、详情和 aggregate mutation。
- 先审查 query plan；只有稳定分页确需时才增加下一顺序读取索引，不创建重复目录表或 role catalog 表。
- 增加显式开发测试态 bootstrap CLI：exact `tenant_ref + workspace_id + user_id`、active account、零 active identity administrator、单事务 membership + `workspace_admin` assignment 与 audit ref；不在 Server 启动或注册时自动执行。
- SQLite / PostgreSQL 覆盖 migration、rollback / reapply、受限 runtime role、cursor、CAS、事务原子性、重启和 no-fallback。

批次 B 停止线：

- 不自动迁移生产数据库，不引入新数据库或 ORM，不打开 production store mode。

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

## 批次 D：Pencil 与 React strict consumer

实施范围：

- 五维评分固定为 `2 / 1 / 2 / 2 / 2 = 9`，采用 `A / 完整 Pencil`。
- 只更新既有 S7 User / Role Desktop、Narrow 与对应 Decision Record；复用 Family UI，不创建 S11。
- 建立单一 strict consumer、workspace member directory、selected member detail、role catalog 和受控 create / revoke flow。
- 覆盖 loading、empty、denied、unavailable、stale conflict、catalog drift、last-admin protection、success 与 revoked 状态。

批次 D 停止线：

- 不在 URL、Web Storage、IndexedDB、service worker、日志或截图持久化目录载荷、确认正文或身份敏感字段。

## 批次 E：双数据库产品连续链与收口

实施范围：

- SQLite 产品链完成第一账户注册、显式 bootstrap CLI、管理员登录，再由页面完成第二账户注册、exact `user_id` 交接、membership create、role assign、业务 permission 生效、role / membership revoke、立即失败关闭与服务重启恢复。
- PostgreSQL 配置化 Server 完成同构连续链、关闭后 no-fallback 与重连恢复。
- 完成 `1440×900`、`720×900`、`390×844`、双标签、console / network / URL / storage / privacy 审计。
- 回写功能专题、Admin 入口、当前焦点、路线图、能力矩阵和周志后关闭专题。

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
- [ ] 批次 A：领域合同、canonical role catalog 与 memory 纵向链。
- [ ] 批次 B：SQLite / PostgreSQL durable owner。
- [ ] 批次 C：Admin HTTP 与 local session 授权。
- [ ] 批次 D：Pencil 与 React strict consumer。
- [ ] 批次 E：双数据库产品连续链与专题收口。
