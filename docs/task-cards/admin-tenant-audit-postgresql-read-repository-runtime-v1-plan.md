# Admin Tenant / Audit PostgreSQL Read Repository Runtime v1

更新时间：2026-07-12

## 任务定位

任务 ID：`admin-tenant-audit-postgresql-read-repository-runtime-v1`

当前状态：`admin_tenant_audit_postgresql_read_repository_runtime_v1_complete`

本任务承接 [Tenant / Audit PostgreSQL Read Repository v1](../features/admin-control-plane/tenant-audit-postgresql-read-repository-v1.md)，把 Admin tenant summary 与 audit summary 接入 PostgreSQL dev/test read repository，并保留其余五条 workspace operation 的显式 fake repository binding。

本任务统一承接 schema、manual migration、read-only runtime、routed selector、typed cursor、HTTP/Web 和浏览器证据；不再拆出同层 readiness 文档。它不接真实 Radish OIDC，不建立 production repository，也不创建 tenant 或 audit 写入能力。

## 前置事实

- shared verified identity、signed test token、permission projection 和负向 auth runtime 已完成。
- 七条 route 已共同依赖 `ControlPlaneReadRepository`，handler 不直接读取 store。
- 当前 tenant / audit 仍来自 fixture-backed fake store。
- Application Draft、Application Publish、Gateway History 与 Workflow Run 已提供 manual migration、checksum、marker、pool lifecycle 和 PostgreSQL integration 范式。
- audit transport canonical filter 为 `event_kind / resource_ref / actor_subject_ref / failure_code`。

## 实现范围

### 批次 A：Schema、Migration 与只读边界

1. 新增 `control_plane_admin_read` migration component、manifest、up/down SQL 和 embedded runner。
2. 建立 `control_plane_read_schema_versions`、tenant projection 与 audit projection 表。
3. migration 使用独立 advisory lock、checksum、幂等 apply、dev/test rollback / reapply。
4. 新增 `radishmind-control-plane-read-migrate status|up`；startup 不自动 migration。
5. runtime pool 在启动时完成 connect、marker、schema version、checksum 和 SELECT preflight。
6. migration DSN 与 runtime DSN 分离；runtime 只要求 SELECT，不提供任何 writer。
7. PostgreSQL integration 证明 runtime role 可 SELECT，不能 INSERT、UPDATE、DELETE、TRUNCATE、CREATE、ALTER 或 DROP。

### 批次 B：Repository、Selector 与分页

1. 实现 tenant / audit PostgreSQL adapter，继续满足现有 `ControlPlaneReadRepository`。
2. routed repository 固定两条 Admin operation → PostgreSQL、五条 workspace operation → fake repository。
3. PostgreSQL failure 不得转读 fake row，也不得解释为 empty。
4. tenant query 只使用 verified context tenant predicate，零行映射 `tenant_not_found`。
5. audit handler 严格解析 filter、limit、sort 和 cursor；非法输入在 repository 前返回 `400 invalid_filter`。
6. cursor 使用 `control_plane_audit_cursor.v1` canonical JSON + base64url，携带 `recorded_at + audit_ref`。
7. audit query 使用 tenant-first keyset pagination、`limit + 1` 和稳定 next cursor。
8. request context cancel、configured timeout、pool close 和 restart recovery 必须可复验。

### 批次 C：HTTP、Web 与浏览器关闭

1. signed token 下覆盖 tenant ready / not found 与 audit ready / empty / filtered / multi-page。
2. 覆盖 invalid filter / cursor、store unavailable、schema mismatch 与 stored contract mismatch。
3. 五条 workspace route 必须证明仍使用 fake repository；Admin 数据库失败不改变 selector binding。
4. Web 保持 sanitized envelope，禁止展示 SQL、DSN、driver error 或 auth material。
5. 浏览器验证 ready、denied、empty、pagination、console、URL、localStorage 与 sessionStorage。

## 配置与组合门禁

- `RADISHMIND_CONTROL_PLANE_READ_STORE=fake_store_dev|postgres_dev_test`
- `RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_DATABASE_URL`
- `RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_MIGRATION_DATABASE_URL`
- `RADISHMIND_CONTROL_PLANE_READ_DATABASE_TIMEOUT`

允许组合：

| auth mode | store mode | 结果 |
| --- | --- | --- |
| `disabled` | 默认 | offline，零 read request |
| `dev_headers` | `fake_store_dev` | 本地开发兼容路径 |
| `signed_test_token` | `fake_store_dev` | auth regression |
| `signed_test_token` | `postgres_dev_test` | 本任务数据库路径 |
| 其它 | 任意 | startup fail closed |

配置摘要只能暴露 store mode、database configured、timeout 与 schema ready 状态，不能暴露 DSN、host、database、user、TLS 或错误正文。

## Schema 锚点

- component：`control_plane_admin_read`
- migration：`0001_admin_tenant_audit_read`
- store schema：`control_plane_admin_read_store_v1`
- tenant projection schema：`control_plane_tenant_summary_projection.v1`
- audit projection schema：`control_plane_audit_summary_projection.v1`
- keyset order：`recorded_at DESC, audit_ref DESC`
- limit：默认 `50`，范围 `1..100`
- cursor 最大输入：`1024` 字符

projection 只保存产品设计中列出的 sanitized scalar columns，不保存 raw JSON payload、token、claim、secret、API key、provider endpoint、raw tenant record 或 raw audit payload。

## Failure 与 No-fallback

| boundary | failure code | HTTP |
| --- | --- | --- |
| invalid filter / limit / sort / cursor | `invalid_filter` | `400` |
| tenant row absent | `tenant_not_found` | `404` |
| database connect / timeout / query | `read_store_unavailable` | `503` |
| marker absent | `schema_migration_not_applied` | startup blocked / `503` |
| marker或 checksum mismatch | `schema_version_mismatch` | startup blocked / `503` |
| stored projection invalid | `read_store_contract_mismatch` | `500` |

所有 failure 都返回 sanitized envelope 与空 items。禁止 stale snapshot、fake fallback、SQL error passthrough 和自动 schema repair。

## 测试矩阵

### Unit / Contract

- config default、env parsing、组合拒绝与 sanitized summary。
- migration manifest、checksum、SQL 必需列、marker 与 down contract。
- routed operation exact matrix、zero-query denial 和 no-fallback。
- tenant predicate、row validation、query failure mapping。
- audit filter、limit、sort、cursor codec、keyset boundary 与 next cursor。
- handler 在 invalid input 时 repository query count 为 0。

### PostgreSQL Integration

- `status / up`、idempotent apply、checksum、marker missing / mismatch、rollback / reapply。
- runtime SELECT preflight 与写入 / DDL deny。
- tenant ready / missing、audit empty / filters / multi-page、cross-tenant isolation。
- timeout / cancel、pool close、database unavailable 与 process restart recovery。

### HTTP / Web / Browser

- signed token positive / denied、稳定 HTTP status 与 sanitized failure。
- five workspace fake bindings 与 two Admin PostgreSQL bindings。
- Web unit、production build、浏览器 ready / empty / unavailable / pagination。
- offline 零请求，URL 只含 section hash，storage 为空，console 无 React / unhandled error。

### 仓库门禁

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- Web unit 与 production build
- PostgreSQL dev/test integration runner
- `git diff --check`
- `./scripts/check-repo.sh --fast`
- `./scripts/check-repo.sh`

## 提交拆分

1. `docs(admin): plan tenant audit postgres read runtime`
2. `feat(platform): add admin read postgres schema`
3. `feat(platform): route admin reads to postgres`
4. `test(platform): cover admin postgres read lifecycle`
5. `docs(admin): close tenant audit postgres read runtime`

如果某一批同时包含无法独立通过的接口与测试，可合并相邻代码 / 测试提交，但文档与代码保持分开。

## 停止线

- 不接真实 Radish OIDC、discovery、JWKS、PKCE、BFF、cookie 或 refresh token。
- 不读取或复制 Radish user / tenant / role / permission 表。
- 不创建 runtime writer、tenant mutation、audit writer、raw export、retention 或 compliance ledger。
- 不迁移 Applications、API keys、quota、workflow definitions 或 runs。
- 不实现 production repository、自动 migration、fallback、application promotion、production key、quota enforcement、billing 或 cost ledger。
- 不扩展 Gateway provider、fallback / load balancing，也不扩展 Workflow tool、confirmation、writeback、replay 或 resume。
- dev/test PostgreSQL、signed token 和浏览器成功不构成 production Admin ready 声明。

## 当前实施记录

- 批次 A 已落地 migration manifest、up/down SQL、checksum、manual `status / up` CLI、runtime pool、marker / checksum / SELECT preflight 和独立 runtime DSN。
- 批次 B 已落地 tenant / audit PostgreSQL adapter、两 Admin / 五 fake routed repository、strict audit request parsing、typed cursor、keyset pagination、request cancellation、pool close 与 no-fallback。
- audit repository contract 已回收为现行 `event_kind / resource_ref / actor_subject_ref / failure_code`，旧 `resource_kind / decision` 漂移已移除。
- Go 全量 unit、race、vet、PostgreSQL integration-tag 编译、Web 58 项单测、production build 和 fast repository baseline 已通过。
- 真实 PostgreSQL `check` 已通过 migration / rollback / reapply、restart reconstruction、tenant / audit read、cross-tenant isolation、multi-page cursor 与 runtime role `42501` write / DDL denial；统一脚本会恢复 Control Plane Admin read schema。
- signed-token HTTP 已覆盖 tenant、audit、workspace fake binding、invalid cursor、第二页与 CORS；Web 通过进程内 test token provider 请求，不使用 dev headers。
- 真实浏览器显示 `signed test token`、`PostgreSQL dev/test projection`、数据库 tenant row 与 3 条 audit row；console 0 error / warning，URL、localStorage、sessionStorage 与页面正文无 token。临时 token、key、服务和容器已清理。

## 完成判定

只有批次 A、B、C 的 unit、integration、HTTP/Web/browser 证据全部通过，且当前焦点、功能专题、路线图、能力矩阵与周志同步后，任务才可标记 `complete`。任一 production owner、真实 OIDC 或 durable audit writer 缺失继续作为明确停止线，不阻塞本 dev/test runtime 的完成。
