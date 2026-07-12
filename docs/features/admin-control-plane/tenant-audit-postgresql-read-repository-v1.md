# Admin Tenant / Audit PostgreSQL Read Repository v1

更新时间：2026-07-12

状态：`admin_tenant_audit_postgresql_read_repository_v1_complete`

## 功能定位

本功能承接 [Authenticated Read Store Transition v1](authenticated-read-store-transition-v1.md) 第二批，在已经完成 shared verified identity / negative auth runtime 的基础上，把 Admin tenant summary 与 audit summary 从 fixture-backed fake store 迁移到 PostgreSQL dev/test read repository。

它建立可重启、可分页、可识别 empty / unavailable / schema mismatch、且不会隐式回退 fake store 的 Admin 读取路径。它不创建 tenant 或 audit 写入 API，不把 dev/test fixture 升级为业务或合规审计真相源，也不接真实 Radish OIDC。

## 当前事实与待解决缺口

- 七条 route 已统一通过 `ControlPlaneReadRepository` interface 读取，handler 不直接接触 store。
- `ReadRepositoryContext` 已携带 request、tenant、subject、scope、audit、issuer 和 session；auth denial 时 repository query count 为 0。
- `disabled / dev_headers / signed_test_token`、13 类负向认证和 sanitized 非 2xx envelope 已完成。
- 当前 repository 默认仍由 fixture-backed fake store 实现；没有 Control Plane PostgreSQL config、migration、pool、adapter 或 operation selector。
- 仓库已有 Saved Draft、Application Draft、Application Publish、Gateway History 与 Workflow Run 的 PostgreSQL dev/test migration / marker / checksum / manual runner / restart 测试模式，可复用工程方法，但不能复用它们的业务表或 repository。
- audit transport 当前允许 `event_kind / resource_ref / actor_subject_ref / failure_code`；旧 repository readiness fixture 还记录 `resource_kind / decision`。实现前必须以现行 route contract 和 TypeScript consumer 为准统一，不在数据库批次扩新 filter。
- 当前 handler 尚未把 `limit` 解析进 `ReadRepositoryRequest`，也没有严格校验 cursor / sort 值。PostgreSQL 批次必须修正这一 contract gap，不能让 adapter 自行猜测分页输入。

## 用户流程

### Tenant Overview

1. 管理员使用 `signed_test_token` 进入显式 dev/test Admin read 环境。
2. auth boundary 校验 identity、tenant 和 `tenant:read`，生成 `ReadRepositoryContext`。
3. operation selector 把 `ReadTenantSummary` 路由到 PostgreSQL adapter。
4. repository 使用 context tenant 作为唯一 tenant predicate，读取一条 sanitized tenant projection。
5. Web 展示 ready、not found、denied、store unavailable 或 schema mismatch；不读取旧 fixture 作为替代。
6. Platform 重启后，同一数据库记录仍可读取，request / audit ref 使用本次请求值而不是持久化旧请求上下文。

### Audit Log

1. 管理员通过同一 verified context 和 `audit:read` 进入 Audit Log。
2. handler 校验 filter、limit、sort 与 cursor，再构造 typed repository request。
3. repository 按 tenant predicate、允许的等值 filter 和 keyset cursor 查询 sanitized audit projection。
4. Web 展示当前页、next cursor、empty 或稳定失败；翻页不得重复、遗漏或跨 tenant。
5. transient database failure 返回 `503 read_store_unavailable`，schema / stored projection contract drift 返回稳定失败；禁止回退 fake audit rows。

## 路由与 Operation 范围

| route | operation | `postgres_dev_test` source | required grant | 本批结果 |
| --- | --- | --- | --- | --- |
| `/v1/control-plane/tenants/{tenant_ref}/summary` | `ReadTenantSummary` | tenant projection table | `tenant:read` | 迁移 |
| `/v1/control-plane/audit` | `ListAuditSummaries` | audit projection table | `audit:read` | 迁移 |
| Applications | `ListApplicationSummaries` | fake store | `applications:read` | 保持显式 dev binding |
| API keys | `ListAPIKeySummaries` | fake store | `api_keys:read` | 保持显式 dev binding |
| quota | `ReadQuotaSummary` | fake store | `usage:read` | 保持显式 dev binding |
| workflow definitions | `ListWorkflowDefinitionSummaries` | fake store | `applications:read` | 保持显式 dev binding |
| runs | `ListRunRecordSummaries` | fake store | `runs:read` | 保持显式 dev binding |

后五条使用 fake store 是 selector 的显式 transition matrix，不是 PostgreSQL operation 失败后的 fallback。任一已迁移 operation 一旦选择 PostgreSQL，连接、query、scan、schema 或 cursor 失败都不得改读 fake store。

## Projection 所有权与数据来源

| 数据 | 当前 dev/test owner | 未来生产所有方 | 禁止解释 |
| --- | --- | --- | --- |
| tenant summary projection | PostgreSQL integration fixture / seed harness | 待 RadishMind control-plane projection ingestion 设计 | 不读取或复制 Radish tenant 表，不声明 tenant 真相源 |
| audit summary projection | PostgreSQL integration fixture / seed harness | 待 append-only audit writer、retention 与 redaction owner 设计 | 不声明合规审计账本，不保存 raw audit payload |
| request / audit correlation | 当前 Platform request boundary | Platform observability / audit boundary | 不持久化或复用旧请求的 Authorization、request id 或 audit context |
| identity / permission | 已完成的 verified auth boundary | reviewed Radish OIDC + permission mapping | repository 不解析 token、claim 或角色名 |

本批没有运行时 writer。集成测试只允许通过 migration / admin test connection 写入 sanitized seed rows；HTTP、Web 和 runtime repository 均无 INSERT / UPDATE / DELETE 能力。

## Schema 设计

独立 schema component：`control_plane_admin_read`。

### `control_plane_read_schema_versions`

- `component text primary key`
- `migration_id text not null`
- `store_schema_version text not null`
- `migration_checksum text not null`
- `applied_at timestamptz not null`

首版 marker 建议：

- migration：`0001_admin_tenant_audit_read`
- store schema：`control_plane_admin_read_store_v1`
- `automatic_apply=false`
- `environment=dev_test_only`

### `control_plane_tenant_summary_projections`

| column | type / constraint | 说明 |
| --- | --- | --- |
| `tenant_ref` | `text primary key` | 唯一 tenant predicate |
| `schema_version` | `text not null` | 必须为 tenant summary projection v1 |
| `projection_version` | `bigint not null check > 0` | fixture / future ingestion version |
| `tenant_display_name` | `text not null` | sanitized display value |
| `tenant_state` | `text not null` | sanitized state |
| `plan_ref` | `text not null` | reference only |
| `quota_summary_ref` | `text not null` | reference only |
| `deployment_status_ref` | `text not null` | reference only |
| `audit_summary_ref` | `text not null` | reference only，不做跨表 truth FK |
| `projected_at` | `timestamptz not null` | projection timestamp |

不保存 user、role、permission、secret、provider endpoint、raw tenant record 或 request auth material。

### `control_plane_audit_summary_projections`

| column | type / constraint | 说明 |
| --- | --- | --- |
| `tenant_ref` | `text not null` | 所有 query 的首个 predicate |
| `audit_ref` | `text not null` | tenant 内唯一 reference |
| `schema_version` | `text not null` | 必须为 audit summary projection v1 |
| `projection_version` | `bigint not null check > 0` | immutable projection version |
| `actor_subject_ref` | `text not null` | sanitized actor reference |
| `event_kind` | `text not null` | allowlisted event kind |
| `resource_ref` | `text not null` | sanitized resource reference |
| `decision` | `text not null` | read-only decision summary |
| `failure_code` | `text null` | sanitized stable code |
| `trace_id` | `text not null` | sanitized trace reference |
| `recorded_at` | `timestamptz not null` | keyset order anchor |
| `projected_at` | `timestamptz not null` | projection timestamp |

主键：`(tenant_ref, audit_ref)`。

必须建立 `(tenant_ref, recorded_at desc, audit_ref desc)` keyset index。现行四个等值 filter 可按真实 query plan 添加 tenant-first composite index；实现任务卡应至少覆盖 `event_kind` 与 `resource_ref`，其余索引以集成测试 explain / query plan 证据决定，不预先堆无使用证据的索引。

## Migration 与角色边界

- migration 目录建议为 `services/platform/migrations/control_plane_admin_read/`，包含 manifest、up/down SQL、embedded runner 和 tests。
- manual CLI 建议为 `cmd/radishmind-control-plane-read-migrate`，只支持 `status / up`；rollback 只暴露给 dev/test integration helper。
- migration URL 与 runtime URL 必须分离：migration role 可 DDL，runtime role 只可 CONNECT、USAGE、SELECT marker 和两张 projection table。
- runtime role 必须通过负向 SQL 测试证明不能 INSERT、UPDATE、DELETE、TRUNCATE、CREATE、ALTER 或 DROP。
- migration 使用独立 advisory lock、checksum、idempotent apply、rollback / reapply 和 marker mismatch 检查；不在 Platform startup 自动 apply。
- config summary 只输出 database configured、store mode、timeout 和 schema readiness，不输出 DSN、host、database、user、TLS material 或错误正文。

## Store Selector 与启动门禁

建议配置：

- `RADISHMIND_CONTROL_PLANE_READ_STORE=fake_store_dev|postgres_dev_test`
- `RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_DATABASE_URL`
- `RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_MIGRATION_DATABASE_URL`
- `RADISHMIND_CONTROL_PLANE_READ_DATABASE_TIMEOUT`

兼容矩阵：

| auth mode | store mode | 结论 |
| --- | --- | --- |
| `disabled` | 默认 / 无 consumer | offline，零 read request |
| `dev_headers` | `fake_store_dev` | 保留本地开发路径 |
| `signed_test_token` | `fake_store_dev` | 认证回归路径 |
| `signed_test_token` | `postgres_dev_test` | 本批唯一数据库组合 |
| 其它组合 | 任意 | startup fail closed |

`postgres_dev_test` selector 必须构造一个显式 routed repository：两个 Admin operation → PostgreSQL adapter，五个 workspace operation → fake repository。禁止用“先试 PostgreSQL、失败再 fake”的链式 wrapper。

启动顺序：config 校验 → database connect / ping → marker inspect → schema version / checksum → runtime SELECT preflight → routed repository 注入 → HTTP listen。任一步失败都拒绝启动，不创建半可用 PostgreSQL mode。启动后发生 transient query failure 则返回 `503`，不切换 store。

## Query、Filter 与分页契约

### Tenant summary

- 只使用 `ReadRepositoryContext.TenantRef`；path tenant 已在 auth boundary 校验，repository 不接受第二个 tenant override。
- 单 query、零行返回 `tenant_not_found`，多行或 contract drift 返回 `read_store_contract_mismatch`。
- response request id / audit ref 来自当前 context，不从 projection row 读取。

### Audit summary

- canonical filter 固定为 `event_kind / resource_ref / actor_subject_ref / failure_code`，全部为 exact match。
- `resource_kind / decision` 不进入本批；旧 readiness fixture 必须在实现 task card 中与 route contract 对齐。
- sort 只允许 `recorded_at_desc`；省略时使用该默认值。
- limit 默认 `50`，允许 `1..100`；非整数、重复值、越界值返回 `400 invalid_filter`。
- cursor 采用 `control_plane_audit_cursor.v1`，只编码上一页最后一项的 `recorded_at + audit_ref`，使用 canonical JSON + base64url；最大输入长度 `1024`，版本、字段、时间或 base64 非法均返回 `400 invalid_filter`。
- cursor 不承担授权，tenant 永远来自 verified context；query 使用 `(recorded_at, audit_ref) < ($cursor_time, $cursor_ref)` 并按两列降序。
- query 读取 `limit + 1` 判断 next page，只返回 limit 条；next cursor 由最后一条返回记录生成。
- 稳定数据集下翻页不得重复或遗漏。并发新增记录允许只出现在新的第一页，不保证旧 cursor snapshot isolation；这一语义必须写入测试说明。

## Failure 与 No-fallback

| boundary | failure code | HTTP | items |
| --- | --- | --- | --- |
| auth missing / invalid | 既有 auth failure | `401` | `[]` |
| tenant / permission denied | `tenant_binding_missing / scope_denied` | `403` | `[]` |
| filter / limit / sort / cursor invalid | `invalid_filter` | `400` | `[]` |
| tenant row absent | `tenant_not_found` | `404` | `[]` |
| audit real empty | none | `200` | `[]` |
| connect / pool / timeout / query unavailable | `read_store_unavailable` | `503` | `[]` |
| marker absent | `schema_migration_not_applied` | startup blocked；若运行后漂移则 `503` | `[]` |
| marker / schema mismatch | `schema_version_mismatch` | startup blocked；若运行后漂移则 `503` | `[]` |
| stored row / projection contract invalid | `read_store_contract_mismatch` | `500` | `[]` |

任何 database failure 都不返回 stale snapshot，不读取 fake row，不把 unavailable 写成 empty，也不把 SQL / driver / DSN 错误正文返回 Web。

## Repository 与资源生命周期

- 继续使用现有 `ControlPlaneReadRepository`；不新增第二套 handler 或 response adapter。
- PostgreSQL adapter 只实现两条 Admin operation 的 SQL；routed repository 负责其余五条显式转发 fake repository。
- 每个成功 route 的 repository query count 必须为 1；auth / filter / cursor denial 为 0。
- query context 使用 request context + configured timeout；client cancel 必须取消数据库 query。
- `Server.Close` 负责关闭 pool；重复 close 安全。启动失败必须关闭已创建 pool。
- Platform restart 使用同一数据库后仍能读取 seed projection；runtime 不执行 schema repair、seed 或 DDL。
- pool 采用有界连接数和 lifetime / idle policy；具体数值在实现 task card 按两条只读 route 的并发测试确定，不复制高写入 store 的默认值。

## 隐私与审计边界

- 表、SQL、日志、response 和浏览器 state 均不得包含 token、Authorization header、cookie、raw claim、membership record、API key value / hash、secret、DSN、raw tenant record 或 raw audit payload。
- actor、tenant、resource、trace、request 和 audit 只保存 / 返回 sanitized reference。
- filter value 不写完整日志；观测只记录 route、filter key 集合、limit、cursor present、decision、failure boundary、duration、query count 和 schema version。
- runtime read 不产生业务 audit row；当前 response `audit_ref` 是 request correlation，不代表 durable compliance audit 已写入。

## 验收方式

### Unit / Contract

- routed repository 两条 PostgreSQL、五条 fake 的 exact operation matrix。
- tenant predicate、required context、single-query、row scan / contract mismatch。
- audit filter allowlist、limit / sort / cursor codec、keyset boundary、next cursor。
- database failure mapping、zero-query denial、no fake fallback、no side effects。

### Migration / PostgreSQL Integration

- manual status / up、idempotent apply、checksum、marker missing / mismatch、rollback / reapply。
- runtime role SELECT success 与 write / DDL denial。
- tenant ready / not found、audit ready / empty / filtered / multi-page、cross-tenant isolation。
- process restart recovery、pool close、query timeout / cancel、database unavailable 和 stored contract mismatch。

### HTTP / Web / Browser

- `signed_test_token + postgres_dev_test` 下 tenant / audit ready。
- authenticated denied、invalid filter / cursor、tenant not found、audit empty、store unavailable 和 schema mismatch。
- 五条 workspace route 明确仍来自 fake store，Admin database failure不改变它们的 selector binding。
- offline 零请求；URL 只有 section hash，localStorage / sessionStorage 为空。
- console 没有 React / unhandled error；预期非 2xx resource 日志需与应用异常分开记录。

### 仓库门禁

- Go unit / integration / race / vet。
- Web unit test / production build。
- PostgreSQL dev/test integration runner。
- `git diff --check`、`./scripts/check-repo.sh --fast` 和完整 `./scripts/check-repo.sh`。

本批新增真实 schema、migration 和 selector，因此需要高风险 task card 与专项 integration tests；优先复用现有聚合门禁，不新增只验证文件存在的同层 checker。

## 实施拆分

### 批次 A：Schema / Migration / Read-only Role

- 创建一个高风险实现 task card，固定最终字段、marker、runner、role policy 和 integration environment。
- 实现 migration manifest、up/down、manual CLI、checksum、rollback / reapply 与 runtime role negative SQL。
- 只建立 dev/test fixture seed harness，不创建 runtime writer。

### 批次 B：Repository / Selector / Pagination

- 实现 tenant / audit PostgreSQL adapter、typed cursor codec、strict request parsing 和 routed repository。
- 接入 config / startup preflight / pool lifecycle；完成 no-fallback、timeout / cancel 和 restart tests。
- 同步旧 repository filter fixture，以现行 route contract 为 canonical source。

### 批次 C：HTTP / Web / Browser Close

- 完成 signed-token HTTP、Web failure state 和真实浏览器 ready / denied / empty / unavailable / pagination。
- 核对 console、URL / storage、secret scan、query count 与五条 workspace route 显式 fake binding。
- 通过后把第二批状态关闭，再决定是否进入 Radish OIDC Integration Test 设计；不会自动晋级 production。

## 停止线

- 不接真实 Radish OIDC、discovery / JWKS、PKCE、BFF、session cookie、refresh token 或 production auth。
- 不读取 Radish 数据库，不复制 user / tenant / role / permission 真相表。
- 不创建 tenant mutation、audit writer、raw audit export、retention / compliance ledger 或管理端写入。
- 不迁移 Applications、API keys、quota、workflow definitions 或 runs；不新增 workspace selector / membership API。
- 不启用 production repository，不自动 migration，不允许数据库 failure 回退 fake store。
- 不实现 application promotion、API key lifecycle、quota enforcement、billing、cost ledger、secret runtime、deployment apply、Gateway fallback / load balancing，或 Workflow tool / confirmation / writeback / replay / resume。
- 不把 schema、PostgreSQL dev/test、restart recovery、浏览器成功或 signed test token 解释为 production Admin ready。

## 下一实现入口

[Admin Tenant / Audit PostgreSQL Read Repository Runtime v1](../../task-cards/admin-tenant-audit-postgresql-read-repository-runtime-v1-plan.md) 已完成 schema/migration、manual CLI、startup preflight、adapter、routed selector、strict cursor、真实 PostgreSQL、HTTP/Web 与浏览器验收。下一产品设计可进入 Radish OIDC Integration Test v1，但不会自动打开 production auth。
