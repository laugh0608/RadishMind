# Control Plane Read-Side 契约

更新时间：2026-06-02

## 契约目的

本专题说明 `Control Plane / User Workspace / Workflow v1` 的只读控制面契约层。它把用户工作区和管理端会消费的 summary、route、response、negative contract、implementation preconditions、fake-store-backed read handler plan、fake-store-backed handler implementation、auth/db preconditions、consumer contract、正式 UI 边界、正式 UI 实现 readiness、shared shell、管理端只读页面切片、用户工作区只读页面切片、formal UI readiness close、dev-only live read consumer、auth/store transition preconditions、repository contract preconditions、disabled database read guard 和 repository contract smoke 固定为可检查治理边界，避免在正式数据库、OIDC 或完整 UI 尚未准备好时，从本地 ops console 直接堆出产品功能。

当前 `control-plane-read-disabled-database-guard-v1` 已把 disabled database read guard 纳入同一 read-side 契约层；它只固定 database / postgres / repository read mode 的 reserved disabled 状态、`database_read_disabled` fail-closed guard 和无 fake fallback 口径，不实现数据库、OIDC、repository、API key / quota、workflow executor、confirmation、writeback 或 replay。

当前 `control-plane-read-repository-contract-smoke-v1` 已把未来 repository contract smoke 的输入输出、七条 read route 覆盖、failure mapping、no fake fallback、no side effects 和文档停止线纳入同一 read-side 契约层；它只定义未来 smoke 应验证什么，不实现 SQL、migration、repository adapter、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

当前 `control-plane-read-repository-implementation-readiness-v1` 已把未来 repository implementation readiness 纳入同一 read-side 契约层；它只固定未来文件落点、实现准入 gate、七条 route readiness matrix、dual smoke plan、failure mapping、no fake fallback、no side effects 和停止线，不创建 Go repository 文件、不实现 repository interface / adapter、不写 SQL、不建 migration、不接真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-store-selection-readiness-v1` 已把未来 store selection readiness 纳入同一 read-side 契约层；它只固定默认 read source、保留 read source、失败映射、七条 route selection matrix、no fake fallback、no side effects 和停止线，不创建正式配置入口、不实现 store selector、不实现 repository interface / adapter、不写 SQL、不建 migration、不接真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 read-side 已实现七条 fake-store-backed Go read route：`tenant-summary-route`、`application-summary-list-route`、`api-key-summary-list-route`、`quota-summary-route`、`workflow-definition-summary-list-route`、`run-record-summary-list-route` 与 `audit-summary-list-route`。这些 route 只使用 in-memory fixture fake store 与 test-only fake auth context，不代表完整 read-side API、数据库 query、真实 OIDC 或正式 UI 已实现。

`control-plane-read-auth-db-preconditions-v1` 进一步固定未来迁移到 `future Radish OIDC / auth middleware` 与 `future control plane read store repository` 之前的准入条件。它只定义 auth context contract、read store repository contract、route transition requirements、failure taxonomy 和 smoke transition plan，不实现真实 auth middleware、数据库 query 或 repository。

`control-plane-read-consumer-contract-v1` 固定 `contracts/typescript/control-plane-read-api.ts`、`scripts/run-control-plane-read-consumer-smoke.py` 和上层 view model 停止线。它只定义 TypeScript consumer contract 与离线 fixture 消费方式，不实现正式 user workspace UI、production admin console、真实 auth/db 或 production API。

`control-plane-read-formal-ui-boundary-v1` 固定正式 UI 边界：`Admin Control Plane` 和 `User Workspace` 的页面划分、页面到 read route 的分配、loading / empty / denied / stale / partial failure / forbidden projection 状态和敏感字段停止线。它不实现 React 页面、不修改当前 ops console、不请求真实后端，也不声明 formal UI 或 production admin console ready。

`control-plane-read-formal-ui-implementation-readiness-v1` 固定正式 UI 实现前的工程 readiness：未来正式产品 UI 预留落点为 `apps/radishmind-web/`，`apps/radishmind-console/` 继续只是本地 ops surface；页面实现顺序、consumer contract 复用、测试策略和停止线均已写入 fixture。该切片不创建 React 页面、不创建 `apps/radishmind-web/`、不修改当前 ops console、不请求真实后端，也不声明 formal UI implementation ready。

`control-plane-read-shared-shell-v1` 是正式 UI 的首个实现切片：已创建 `apps/radishmind-web/` 的 `shared-read-shell`，复用 `contracts/typescript/control-plane-read-api.ts` 渲染 route catalog、共享状态组件和 forbidden output guard。该切片仍不请求真实后端、不接数据库 / OIDC、不实现 API key / quota、workflow executor、confirmation、writeback 或 replay，也不把 `apps/radishmind-console/` 改成正式产品端。

`control-plane-read-admin-tenant-overview-v1` 是正式 UI 的首个页面切片：在 `apps/radishmind-web/` 的 shared shell 内新增只读 `admin-tenant-overview`，只消费 `tenant-summary-route` 的离线 consumer view model，展示租户摘要、route metadata、request / audit ref、页面状态和 forbidden output guard。该切片不请求 live backend、不接数据库 / OIDC、不实现 API key / quota、workflow executor、confirmation、writeback 或 replay，也不声明 production admin console ready。

`control-plane-read-workspace-applications-v1` 是正式 UI 的首个用户工作区列表页切片：在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-applications`，只消费 `application-summary-list-route` 的离线 consumer view model，展示应用摘要列表、cursor、route metadata、request / audit ref、页面状态和 forbidden output guard。该切片不请求 live backend、不接数据库 / OIDC、不实现 API key / quota、workflow executor、confirmation、writeback 或 replay，也不声明 formal user workspace complete。

`control-plane-read-workspace-api-keys-v1` 是正式 UI 的用户工作区 API key 列表页切片：在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-api-keys`，只消费 `api-key-summary-list-route` 的离线 consumer view model，展示 API key id、owner、scope、state、时间字段、route metadata、request / audit ref、页面状态和 forbidden output guard。该切片不请求 live backend、不接数据库 / OIDC、不实现 API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay，不展示 key value 或 hash，也不声明 formal user workspace complete。

`control-plane-read-workspace-usage-quota-v1` 是正式 UI 的用户工作区 usage quota 页面切片：在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-usage-quota`，只消费 `quota-summary-route` 的离线 consumer view model，展示 quota id、period、request / token / cost limit、usage snapshot、over quota failure code、route metadata、request / audit ref、页面状态和 forbidden output guard。该切片不请求 live backend、不接数据库 / OIDC、不实现 quota enforcement、rate limit、billing、cost ledger、workflow executor、confirmation、writeback 或 replay，也不声明 formal user workspace complete。

`control-plane-read-workspace-workflow-definitions-v1` 是正式 UI 的用户工作区 workflow definitions 页面切片：在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-workflow-definitions`，只消费 `workflow-definition-summary-list-route` 的离线 consumer view model，展示 workflow definition id、application ref、version、definition status、node count、risk level、requires confirmation capable、updated at、route metadata、request / audit ref、页面状态和 forbidden output guard。该切片不请求 live backend、不接数据库 / OIDC、不实现 workflow builder、workflow definition lifecycle mutation、workflow executor、tool executor、confirmation、writeback 或 replay，也不声明 formal user workspace complete。

`control-plane-read-workspace-run-history-v1` 是正式 UI 的用户工作区 run history 页面切片：在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-run-history`，只消费 `run-record-summary-list-route` 的离线 consumer view model，展示 run id、workflow definition ref、application ref、status、failure code、cost summary、trace id、started / completed timestamp、route metadata、request / audit ref、cursor、页面状态和 forbidden output guard。该切片不请求 live backend、不接数据库 / OIDC、不实现 workflow executor、tool executor、run replay、run resume、materialized result reader、confirmation、writeback 或 replay，也不声明 formal user workspace complete。

`control-plane-read-admin-audit-log-v1` 是正式 UI 的管理端 audit log 页面切片：在 `apps/radishmind-web/` 的 shared shell 内新增只读 `admin-audit-log`，只消费 `audit-summary-list-route` 的离线 consumer view model，展示 audit ref、actor、event kind、resource、decision、failure code、trace id、recorded timestamp、route metadata、request / audit ref、cursor、页面状态和 forbidden output guard。该切片不请求 live backend、不接数据库 / OIDC、不实现 durable audit store、raw payload export、audit record mutation、workflow executor、confirmation、writeback 或 replay，也不声明 production admin console ready。

`control-plane-read-formal-ui-readiness-close-v1` 是当前 read-side formal UI 页面集合的聚合收口：用 surface matrix 固定 `admin-tenant-overview`、`admin-audit-log`、`workspace-applications`、`workspace-api-keys`、`workspace-usage-quota`、`workspace-workflow-definitions` 与 `workspace-run-history` 到七条 read route、离线 consumer view model、状态预览、request / audit ref 和 forbidden output guard 的一致性，状态为 `formal_ui_readiness_closed`。该收口不新增页面功能、不请求 live backend、不接数据库 / OIDC、不实现 API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay，也不声明 production admin console 或完整 formal user workspace ready。

`control-plane-read-dev-live-consumer-v1` 是当前 read-side formal UI 的 dev-only live read consumer 切片：`apps/radishmind-web/` 默认仍使用 `offline_fixture` view model；只有显式设置 `VITE_RADISHMIND_READ_SOURCE=dev-live-http` 时才进入 `dev_live_http` 路径，通过 HTTP 请求七条 fake-store-backed read route。平台服务只有在 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 时才接受 `X-RadishMind-Dev-Read-*` 测试身份 header 并注入 test-only fake auth context。该切片不接真实数据库、不接 Radish OIDC、不实现 repository migration、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay，也不声明 production API consumer ready。

`control-plane-read-auth-store-transition-preconditions-v1` 是 dev-only live read consumer 之后的 auth/store transition preconditions 切片：它固定从当前 dev fake auth header / test-only fake auth context 与 fixture-backed fake store，迁移到未来 `future Radish OIDC / auth middleware` 与 `future control plane read store repository` 前必须满足的 auth middleware gates、read store gates、route transition matrix、dual smoke plan、failure code 和禁止项。该切片只定义迁移准入条件，不接真实数据库、不接 Radish OIDC、不实现 token validation、repository migration、repository implementation、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-contract-preconditions-v1` 是 auth/store transition preconditions 之后的 read store repository contract 前置切片：它固定未来 `ControlPlaneReadRepository` interface、`ReadRepositoryContext`、七条 route 到 repository operation 的映射、tenant predicate、sanitized projection、cursor/filter/sort allowlist、failure mapping 和 contract smoke 要求。该切片只定义 repository contract preconditions，不写 SQL、不建 migration、不实现 repository、不接真实数据库、不接 Radish OIDC、不实现 token validation、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-disabled-database-guard-v1` 是 repository contract preconditions 之后的 disabled database read guard 切片：它固定 database / postgres / repository read mode 当前仍是 reserved disabled，七条 read route 在误请求 database mode 时必须 fail-closed 为 `database_read_disabled`，不得静默回退到 fixture-backed fake store，也不得产生写入、executor、confirmation、writeback 或 replay 副作用。该切片不新增正式配置入口、不实现 repository adapter、不写 SQL、不建 migration、不接真实数据库、不接 Radish OIDC、不实现 token validation、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-contract-smoke-v1` 是 disabled database read guard 之后的 repository contract smoke 定义切片：它固定未来 `ControlPlaneReadRepositoryContractSmoke` 的输入字段、repository context、request / output envelope、七条 route smoke matrix、failure mapping、no fake fallback 和 no side effects。该切片不实现 smoke runner、不实现 repository adapter、不写 SQL、不建 migration、不接真实数据库、不接 Radish OIDC、不实现 token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-implementation-readiness-v1` 是 repository contract smoke 之后的 repository implementation readiness 切片：它固定未来 repository contract / adapter 文件落点、实现准入 gate、route readiness matrix、dual smoke plan、failure mapping 和停止线。该切片不创建 Go repository 文件、不实现 repository interface 或 adapter、不写 SQL、不建 migration、不接真实数据库、不接 Radish OIDC、不实现 token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-store-selection-readiness-v1` 是 repository implementation readiness 之后的 store selection readiness 切片：它固定未来 read source selector 的默认 fake-store source、database / postgres / repository reserved disabled source、未知 selector fail-closed、route selection matrix、failure mapping 和 no fake fallback。该切片不创建正式配置入口、不实现 store selector、不实现 repository interface 或 adapter、不写 SQL、不建 migration、不接真实数据库、不接 Radish OIDC、不实现 token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

## 阶段门禁调整

`admin-audit-log` 已完成当前 read-side UI 页面集合的最后一个优先页面。普通 read-only UI 展示页后续不再默认逐项新增独立 task card、fixture 和 checker，应通过 `control-plane-read-formal-ui-readiness-close-v1` 一类聚合收口，统一校验页面到 `CONTROL_PLANE_READ_ROUTES` / `CONTROL_PLANE_READ_ROUTE_DEFINITIONS` 的绑定、状态组件、forbidden output guard、request / audit ref 和停止线。

`control-plane-read-dev-live-consumer-v1` 已把“不请求 live backend”放宽为显式 dev-only live read path。它只能用于验证 `apps/radishmind-web/` 通过 HTTP 消费现有 read API shape，并且只能连接 fake-store-backed handler 与测试身份上下文。该放宽不允许接入真实数据库、Radish OIDC、repository migration、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-auth-store-transition-preconditions-v1` 把 dev-only live read path 之后的下一步限定为迁移前置治理：先证明 auth context、repository interface、tenant predicate、sanitized projection、dual smoke 和 fail-closed guard 都已定义，再谈真实 middleware 或 store 实现。该 gate 不允许在同一切片内加入数据库 query、OIDC validation、repository migration 或生产 consumer。

`control-plane-read-repository-contract-preconditions-v1` 将停止线前移到 repository contract 层：可以定义 interface、operation matrix、projection / filter / sort allowlist 和 failure mapping，但仍不允许在同一切片内加入 SQL、migration、repository implementation、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-disabled-database-guard-v1` 将停止线继续前移到 database mode guard 层：可以声明 database / postgres / repository read mode 是 reserved disabled，并固定 `database_read_disabled` fail-closed 行为，但仍不允许新增正式配置启用入口、SQL、migration、repository adapter、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-repository-contract-smoke-v1` 将停止线固定到 future smoke 层：可以定义未来 repository contract smoke 的输入输出、route 覆盖、failure mapping、no fake fallback 和 no side effects，但仍不允许在同一切片内加入 SQL、migration、repository adapter、真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-repository-implementation-readiness-v1` 将停止线固定到 implementation readiness 层：可以定义未来文件落点、实现准入、dual smoke 和 failure mapping，但仍不允许在同一切片内创建 Go repository 文件、加入 SQL、migration、repository adapter、真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-store-selection-readiness-v1` 将停止线固定到 store selection readiness 层：可以定义 future read source selector 的默认值、保留值、失败映射和 route matrix，但仍不允许在同一切片内创建正式配置入口、实现 selector、加入 SQL、migration、repository adapter、真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

## 分层关系

当前 read-side 契约按二十七层固定：

1. `control-plane-read-model-v1`
   - 固定 tenant、application、API key、quota、workflow definition、run record 和 audit 的只读 summary 模型。
   - 固定访问策略、脱敏策略和停止线。
2. `control-plane-read-route-contract-v1`
   - 固定七类 tenant-scoped read-only route contract。
   - 固定 `GET` 方法、scope、分页 / 过滤、失败分类和 fail-closed 访问边界。
3. `control-plane-read-response-fixtures-v1`
   - 固定统一 response envelope：`request_id`、`tenant_ref`、`items`、`next_cursor`、`failure_code`、`audit_ref`。
   - 固定成功 / 失败样例、`failure_code` 来源和敏感字段脱敏。
4. `control-plane-read-negative-contract-v1`
   - 固定缺少身份、cross-tenant read、scope denied、invalid filter、forbidden method、forbidden query、forbidden fallback 和敏感字段投影拒绝。
   - 固定 fail-closed、无副作用、无 executor、无 database write、无 confirmation decision、无 writeback 和无 replay。
5. `control-plane-read-implementation-preconditions-v1`
   - 固定未来 read route 实现前必须具备的 `handler ownership`、`fake store` strategy、`auth middleware` dependency、response fixture conformance 和 negative route smoke readiness。
   - 固定首轮实现只能进入 fixture-backed fake store 与显式 fake auth context，不声明 Go handler、route smoke、数据库、OIDC、API key lifecycle、quota enforcement、executor、confirmation、writeback 或 replay ready。
6. `control-plane-read-fake-store-handler-plan-v1`
   - 固定未来 fake-store-backed read handler plan 的 Go package 落点、future file layout、test-only fake auth context、route smoke 顺序和 no-side-effect gate。
   - 这是计划证据，本身不创建 handler；已被下一层 implementation 消费。
7. `control-plane-read-fake-store-handler-implementation-v1`
   - 固定七条 read route 的 fake-store-backed handler implementation。
   - 只使用 `services/platform/internal/httpapi` 内的 in-memory fake store 与 test-only fake auth context，不接数据库、OIDC、executor、confirmation、writeback 或 replay。
8. `control-plane-read-auth-db-preconditions-v1`
   - 固定未来替换 test-only fake auth context 与 in-memory fixture fake store 前必须满足的 auth/db preconditions。
   - 固定 `future Radish OIDC / auth middleware`、`future control plane read store repository`、route transition、failure taxonomy、smoke transition 和停止线；不接真实 OIDC、不写数据库 query、不实现 repository。
9. `control-plane-read-consumer-contract-v1`
   - 固定上层消费契约：route catalog、request / response 类型、统一 envelope、failure view、cursor view、forbidden output 检测和只读 view model。
   - 由 `contracts/typescript/control-plane-read-api.ts` 与 `scripts/run-control-plane-read-consumer-smoke.py --check` 固定；不实现正式 UI、不请求真实后端、不接数据库或 OIDC。
10. `control-plane-read-formal-ui-boundary-v1`
   - 固定正式 UI 边界：`admin-tenant-overview`、`admin-audit-log`、`workspace-applications`、`workspace-api-keys`、`workspace-usage-quota`、`workspace-workflow-definitions` 与 `workspace-run-history` 的 route 消费关系。
   - 固定只读 UI 状态和禁止命令；不实现 React UI、不升级当前 ops console、不接真实数据库或 OIDC。
11. `control-plane-read-formal-ui-implementation-readiness-v1`
   - 固定正式 UI 实现 readiness：`apps/radishmind-web/` 预留 app 落点、`apps/radishmind-console/` 本地 ops 边界、页面实现顺序、consumer contract 复用、测试策略和停止线。
   - 不创建 React 页面、不创建 product UI app、不修改当前 ops console、不请求真实后端、不接数据库、OIDC、executor、confirmation、writeback 或 replay。
12. `control-plane-read-shared-shell-v1`
   - 创建 `apps/radishmind-web/` 首个 read-only shared shell，固定 route catalog binding、共享状态组件和 forbidden output guard。
   - 不请求 live backend、不接真实 auth/db、不提供 create / edit / issue / execute / confirm / replay 控件。
13. `control-plane-read-admin-tenant-overview-v1`
   - 在 shared shell 内实现只读 `admin-tenant-overview` 页面切片，只消费 `tenant-summary-route`。
   - 固定 tenant summary 展示、route metadata、request / audit ref、页面状态预览和 forbidden output guard；不请求 live backend、不提供写入或执行控件。
14. `control-plane-read-workspace-applications-v1`
   - 在 shared shell 内实现只读 `workspace-applications` 页面切片，只消费 `application-summary-list-route`。
   - 固定 application summary 列表、cursor 状态、route metadata、request / audit ref、页面状态预览和 forbidden output guard；不请求 live backend、不提供写入或执行控件。
15. `control-plane-read-workspace-api-keys-v1`
   - 在 shared shell 内实现只读 `workspace-api-keys` 页面切片，只消费 `api-key-summary-list-route`。
   - 固定 API key summary 列表、scope、状态、时间字段、route metadata、request / audit ref、页面状态预览和 forbidden output guard；不请求 live backend、不展示 key value 或 hash、不提供 issue / rotate / revoke 控件。
16. `control-plane-read-workspace-usage-quota-v1`
   - 在 shared shell 内实现只读 `workspace-usage-quota` 页面切片，只消费 `quota-summary-route`。
   - 固定 quota id、period、request / token / cost limit、usage snapshot、over quota failure code、route metadata、request / audit ref、页面状态预览和 forbidden output guard；不请求 live backend、不提供 quota enforcement、rate limit、billing 或 cost ledger 控件。
17. `control-plane-read-workspace-workflow-definitions-v1`
   - 在 shared shell 内实现只读 `workspace-workflow-definitions` 页面切片，只消费 `workflow-definition-summary-list-route`。
   - 固定 workflow definition id、application ref、version、definition status、node count、risk level、requires confirmation capable、updated at、route metadata、request / audit ref、页面状态预览和 forbidden output guard；不请求 live backend、不提供 create / edit / run / confirm / replay 控件。
18. `control-plane-read-workspace-run-history-v1`
   - 在 shared shell 内实现只读 `workspace-run-history` 页面切片，只消费 `run-record-summary-list-route`。
   - 固定 run id、workflow definition ref、application ref、status、failure code、cost summary、trace id、started / completed timestamp、route metadata、request / audit ref、cursor、页面状态预览和 forbidden output guard；不请求 live backend、不提供 start / cancel / resume / replay / materialize result / write business truth 控件。
19. `control-plane-read-admin-audit-log-v1`
   - 在 shared shell 内实现只读 `admin-audit-log` 页面切片，只消费 `audit-summary-list-route`。
   - 固定 audit ref、actor、event kind、resource、decision、failure code、trace id、recorded timestamp、route metadata、request / audit ref、cursor、页面状态预览和 forbidden output guard；不请求 live backend、不提供 edit / delete / raw payload export / reveal secret 控件。
20. `control-plane-read-formal-ui-readiness-close-v1`
   - 用 surface matrix 聚合固定七个 read-side formal UI 页面切片和七条 read route 的绑定关系。
   - 固定状态预览、request / audit ref、forbidden output guard、`apps/radishmind-console/` 边界和停止线；不新增页面功能、不请求 live backend、不声明 production admin console 或完整 formal user workspace ready。
21. `control-plane-read-dev-live-consumer-v1`
   - 在 `apps/radishmind-web/` 内新增可选 `dev_live_http` consumer 路径，默认仍使用离线 fixture/view model。
   - 通过 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 和 `X-RadishMind-Dev-Read-*` 测试身份 header 消费 fake-store-backed read handlers；不接真实数据库、Radish OIDC、repository migration、生产 API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
22. `control-plane-read-auth-store-transition-preconditions-v1`
   - 固定 dev fake auth / fixture-backed fake store 迁移到未来 auth middleware / read store repository 前的 auth/store transition preconditions。
   - 固定 auth middleware gates、read store gates、route matrix、dual smoke plan、failure codes 和禁止项；不接真实数据库、Radish OIDC、token validation、repository migration、repository implementation、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
23. `control-plane-read-repository-contract-preconditions-v1`
   - 固定未来 read store repository contract 的 interface、context、operation matrix、tenant predicate、sanitized projection、cursor/filter/sort allowlist、failure mapping 和 contract smoke 要求。
   - 不写 SQL、不建 migration、不实现 repository、不接真实数据库、Radish OIDC、token validation、repository migration、repository implementation、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
24. `control-plane-read-disabled-database-guard-v1`
   - 固定 database / postgres / repository read mode 当前仍是 reserved disabled，误请求必须 fail-closed 为 `database_read_disabled`。
   - 不新增正式配置入口、不回退到 fake store、不写 SQL、不建 migration、不实现 repository adapter、不接真实数据库、Radish OIDC、token validation、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
25. `control-plane-read-repository-contract-smoke-v1`
   - 固定未来 repository contract smoke 的输入输出、七条 read route 覆盖、failure mapping、no fake fallback 和 no side effects。
   - 不实现 smoke runner、不写 SQL、不建 migration、不实现 repository adapter、不接真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
26. `control-plane-read-repository-implementation-readiness-v1`
   - 固定未来 repository implementation readiness 的文件落点、实现准入 gate、七条 route readiness matrix、dual smoke plan、failure mapping 和停止线。
   - 不创建 Go repository 文件、不实现 repository interface 或 adapter、不写 SQL、不建 migration、不接真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
27. `control-plane-read-store-selection-readiness-v1`
   - 固定未来 store selection readiness 的默认 source、保留 source、失败映射、七条 route selection matrix、no fake fallback 和 no side effects。
   - 不创建正式配置入口、不实现 store selector、不实现 repository interface 或 adapter、不写 SQL、不建 migration、不接真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

## 程序化证据

read-side 契约当前由以下 fixture 和 checker 固定：

- `scripts/checks/fixtures/control-plane-read-model-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-model-v1.py`
- `scripts/checks/fixtures/control-plane-read-route-contract-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-route-contract-v1.py`
- `scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-response-fixtures-v1.py`
- `scripts/checks/fixtures/control-plane-read-negative-contract-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-negative-contract-v1.py`
- `scripts/checks/fixtures/control-plane-read-implementation-preconditions-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-implementation-preconditions-v1.py`
- `scripts/checks/fixtures/control-plane-read-fake-store-handler-plan-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-fake-store-handler-plan-v1.py`
- `scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-fake-store-handler-implementation-v1.py`
- `scripts/checks/fixtures/control-plane-read-auth-db-preconditions-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-auth-db-preconditions-v1.py`
- `scripts/checks/fixtures/control-plane-read-consumer-contract-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-consumer-contract-v1.py`
- `scripts/checks/fixtures/control-plane-read-formal-ui-boundary-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-formal-ui-boundary-v1.py`
- `scripts/checks/fixtures/control-plane-read-formal-ui-implementation-readiness-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-formal-ui-implementation-readiness-v1.py`
- `scripts/checks/fixtures/control-plane-read-shared-shell-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-shared-shell-v1.py`
- `scripts/checks/fixtures/control-plane-read-admin-tenant-overview-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-admin-tenant-overview-v1.py`
- `scripts/checks/fixtures/control-plane-read-workspace-applications-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-workspace-applications-v1.py`
- `scripts/checks/fixtures/control-plane-read-workspace-api-keys-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-workspace-api-keys-v1.py`
- `scripts/checks/fixtures/control-plane-read-workspace-usage-quota-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-workspace-usage-quota-v1.py`
- `scripts/checks/fixtures/control-plane-read-workspace-workflow-definitions-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-workspace-workflow-definitions-v1.py`
- `scripts/checks/fixtures/control-plane-read-workspace-run-history-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-workspace-run-history-v1.py`
- `scripts/checks/fixtures/control-plane-read-admin-audit-log-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-admin-audit-log-v1.py`
- `scripts/checks/fixtures/control-plane-read-formal-ui-readiness-close-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-formal-ui-readiness-close-v1.py`
- `scripts/checks/fixtures/control-plane-read-dev-live-consumer-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-dev-live-consumer-v1.py`
- `scripts/checks/fixtures/control-plane-read-auth-store-transition-preconditions-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-auth-store-transition-preconditions-v1.py`
- `scripts/checks/fixtures/control-plane-read-repository-contract-preconditions-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-repository-contract-preconditions-v1.py`
- `scripts/checks/fixtures/control-plane-read-disabled-database-guard-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-disabled-database-guard-v1.py`
- `scripts/checks/fixtures/control-plane-read-repository-contract-smoke-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-repository-contract-smoke-v1.py`
- `scripts/checks/fixtures/control-plane-read-repository-implementation-readiness-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-repository-implementation-readiness-v1.py`
- `scripts/checks/fixtures/control-plane-read-store-selection-readiness-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-store-selection-readiness-v1.py`
- `contracts/typescript/control-plane-read-api.ts`
- `scripts/run-control-plane-read-consumer-smoke.py`

这些 checker 已接入 `scripts/check-repo.py --fast`。它们的作用是防止契约、样例、负向边界、实现前置条件、fake-store-backed read handler plan、handler implementation、auth/db preconditions、consumer contract、正式 UI 边界、正式 UI 实现 readiness、shared read shell、admin tenant overview、workspace applications、workspace api keys、workspace usage quota、workspace workflow definitions、workspace run history、admin audit log 页面切片、formal UI readiness close、dev-only live consumer、auth/store transition preconditions、repository contract preconditions、disabled database read guard、repository contract smoke、repository implementation readiness、store selection readiness 和文档说明互相漂移，不负责启动服务或模拟真实数据库。

## 路由范围

当前固定的七类只读 route 是：

- `GET /v1/control-plane/tenants/{tenant_ref}/summary`
- `GET /v1/user-workspace/applications`
- `GET /v1/user-workspace/api-keys`
- `GET /v1/user-workspace/usage/quota-summary`
- `GET /v1/user-workspace/workflow-definitions`
- `GET /v1/user-workspace/runs`
- `GET /v1/control-plane/audit`

这些 route 已由 `control-plane-read-fake-store-handler-implementation-v1` 注册为 fake-store-backed Go route，并由 Go 单元测试覆盖成功、missing identity、tenant binding missing、cross-tenant query denied、scope denied、invalid filter、forbidden sensitive projection、forbidden method、forbidden query 和 no-side-effects。`control-plane-read-consumer-contract-v1` 已固定 TypeScript consumer contract，`control-plane-read-formal-ui-boundary-v1` 已固定 formal UI 页面边界；这些 route 仍不是数据库 read path、真实 OIDC auth path、正式 UI implementation ready 或 production ready。

## 输出边界

read-side 输出只允许暴露已脱敏 summary、计数、状态、成本摘要、trace id、audit ref 和 redacted secret reference。以下内容不得通过 read-side 输出或投影泄漏：

- raw secret value
- API key value 或 API key hash
- authorization header、bearer token、cookie value
- raw request body dump
- raw tool payload
- business writeback payload
- full prompt dump with secret

## 停止线

- 不把 read model 写成数据库 schema。
- 不把 route contract 写成 Go handler 已完成。
- 不把 response fixture 写成真实 API 返回。
- 不把 negative contract 写成 route smoke 已完成。
- 不把 implementation preconditions 写成 Go handler、fake store、auth middleware 或 route smoke 已完成。
- 不把 fake-store-backed read handler plan 写成完整 route implementation。
- 不把 `control-plane-read-fake-store-handler-implementation-v1` 写成完整 read-side API、真实数据库 read path、OIDC auth path 或 UI consumer ready。
- 不把 `control-plane-read-auth-db-preconditions-v1` 写成 auth middleware ready、database ready、repository ready 或完整 read-side API ready。
- 不把 `future control plane read store repository` 写成当前数据库 query、migration 或 durable adapter。
- 不把 `control-plane-read-consumer-contract-v1` 写成正式 user workspace UI、production admin console、真实 API consumer 或 production ready。
- 不把 `control-plane-read-formal-ui-boundary-v1` 写成 React UI ready、formal user workspace UI ready、production admin console ready、real API consumer ready 或 production ready。
- 不把 `control-plane-read-formal-ui-implementation-readiness-v1` 写成 React UI implemented、formal user workspace UI ready、production admin console ready、real API consumer ready 或 production ready。
- 不把 `control-plane-read-shared-shell-v1` 写成完整 formal user workspace、production admin console、real API consumer、database ready、OIDC ready 或 production ready。
- 不把 `control-plane-read-admin-tenant-overview-v1` 写成完整 formal user workspace、production admin console、real API consumer、database ready、OIDC ready 或 production ready。
- 不把 `control-plane-read-workspace-applications-v1` 写成完整 formal user workspace、production admin console、real API consumer、database ready、OIDC ready 或 production ready。
- 不把 `control-plane-read-workspace-api-keys-v1` 写成完整 formal user workspace、production admin console、real API consumer、database ready、OIDC ready、API key lifecycle ready、quota enforcement ready 或 production ready。
- 不把 `control-plane-read-workspace-usage-quota-v1` 写成完整 formal user workspace、production admin console、real API consumer、database ready、OIDC ready、quota enforcement ready、rate limit ready、billing ready、cost ledger ready 或 production ready。
- 不把 `control-plane-read-workspace-workflow-definitions-v1` 写成完整 formal user workspace、production admin console、real API consumer、database ready、OIDC ready、workflow builder ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-workspace-run-history-v1` 写成完整 formal user workspace、production admin console、real API consumer、database ready、OIDC ready、workflow executor ready、run replay ready、run resume ready、materialized result reader ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-admin-audit-log-v1` 写成完整 formal user workspace、production admin console、real API consumer、database ready、OIDC ready、durable audit store ready、raw audit payload export ready、audit record mutation ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-formal-ui-readiness-close-v1` 写成完整 formal user workspace、production admin console、real API consumer、database ready、OIDC ready、API key lifecycle ready、quota enforcement ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-dev-live-consumer-v1` 写成 production API consumer、真实数据库 read path、Radish OIDC auth path、repository migration、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-auth-store-transition-preconditions-v1` 写成 Radish OIDC ready、auth middleware ready、token validation ready、database ready、read store repository ready、repository migration ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-repository-contract-preconditions-v1` 写成 repository implementation ready、database query ready、database migration ready、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-repository-contract-smoke-v1` 写成 repository implementation ready、smoke runner implemented、database query ready、database migration ready、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-repository-implementation-readiness-v1` 写成 repository implementation ready、repository adapter ready、database query ready、database migration ready、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-store-selection-readiness-v1` 写成 store selector implemented、正式配置入口 ready、repository implementation ready、repository adapter ready、database query ready、database migration ready、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `apps/radishmind-web/` 首个 shared shell 写成完整 product UI app。
- 不把 `apps/radishmind-console/` 改写成正式用户端、生产管理端或 control plane write surface。
- 不把 `scripts/run-control-plane-read-consumer-smoke.py` 写成真实 API smoke；它只消费 committed fixture。
- 不通过 read-side 契约启用 OIDC、API key lifecycle、quota enforcement、billing ledger、workflow executor、confirmation、business writeback 或 replay。
- 不把本地 ops console 提升为正式 user workspace 或 production admin console。
