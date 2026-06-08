# Control Plane Read-Side 契约

更新时间：2026-06-08

## 契约目的

本专题说明 `Control Plane / User Workspace / Workflow v1` 的只读控制面契约层。它把用户工作区和管理端会消费的 summary、route、response、negative contract、implementation preconditions、fake-store-backed read handler plan、fake-store-backed handler implementation、auth/db preconditions、consumer contract、正式 UI 边界、正式 UI 实现 readiness、shared shell、管理端只读页面切片、用户工作区只读页面切片、formal UI readiness close、dev-only live read consumer、auth/store transition preconditions、repository contract preconditions、disabled database read guard、repository contract smoke、repository implementation readiness、store selection readiness、schema migration readiness、schema migration implementation preconditions、repository adapter implementation plan、schema artifact manifest readiness、store selector smoke readiness、production auth readiness、adapter smoke readiness 和 implementation trigger review 固定为可检查治理边界，避免在正式数据库、OIDC 或完整 UI 尚未准备好时，从本地 ops console 直接堆出产品功能。

当前 `control-plane-read-disabled-database-guard-v1` 已把 disabled database read guard 纳入同一 read-side 契约层；它只固定 database / postgres / repository read mode 的 reserved disabled 状态、`database_read_disabled` fail-closed guard 和无 fake fallback 口径，不实现数据库、OIDC、repository、API key / quota、workflow executor、confirmation、writeback 或 replay。

当前 `control-plane-read-repository-contract-smoke-v1` 已把未来 repository contract smoke 的输入输出、七条 read route 覆盖、failure mapping、no fake fallback、no side effects 和文档停止线纳入同一 read-side 契约层；它只定义未来 smoke 应验证什么，不实现 SQL、migration、repository adapter、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

当前 `control-plane-read-repository-implementation-readiness-v1` 已把未来 repository implementation readiness 纳入同一 read-side 契约层；它只固定未来文件落点、实现准入 gate、七条 route readiness matrix、dual smoke plan、failure mapping、no fake fallback、no side effects 和停止线，不创建 Go repository 文件、不实现 repository interface / adapter、不写 SQL、不建 migration、不接真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-store-selection-readiness-v1` 已把未来 store selection readiness 纳入同一 read-side 契约层；它只固定默认 read source、保留 read source、失败映射、七条 route selection matrix、no fake fallback、no side effects 和停止线，不创建正式配置入口、不实现 store selector、不实现 repository interface / adapter、不写 SQL、不建 migration、不接真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-schema-migration-readiness-v1` 已把未来 schema migration readiness 纳入同一 read-side 契约层；它只固定 schema ownership、migration layout、rollback plan、tenant index strategy、read-only role policy、migration smoke、failure mapping 和停止线，不创建 migration 目录、不写 SQL、不实现 migration runner、store selector、repository interface / adapter、真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-repository-contract-types-readiness-v1` 已把未来 repository contract types readiness 纳入同一 read-side 契约层；它只固定 `ReadRepositoryContext`、七条 read route 的 request / result type、failure code type、projection / filter / sort type 和 contract smoke type 输入，不创建 Go contract type 文件、不实现 repository interface / adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-repository-contract-types-implementation-v1` 已把 repository contract types implementation 纳入同一 read-side 契约层；它只创建 Go contract type 文件和单元测试，落地 `ReadRepositoryContext`、七条 read route 的 request / result type、failure code、projection / filter / sort type 和 route type matrix，不声明 `ControlPlaneReadRepository` interface，不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-repository-contract-smoke-runner-readiness-v1` 已把未来 repository contract smoke runner readiness 纳入同一 read-side 契约层；它只固定 `ControlPlaneReadRepositoryContractSmokeRunner` 未来如何消费 `controlPlaneReadRepositoryRouteTypeContracts()`、既有 smoke fixture、failure mapping、no fake fallback 和 no side effects，不创建 runner 文件、不实现 repository interface / adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-repository-contract-smoke-runner-implementation-v1` 已把 repository contract smoke runner implementation 纳入同一 read-side 契约层；它只创建静态 `ControlPlaneReadRepositoryContractSmokeRunner`，消费 `controlPlaneReadRepositoryRouteTypeContracts()` 并与既有 smoke fixture 对齐七条 read route、failure mapping、no fake fallback 和 no side effects，不声明 repository interface、不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-repository-interface-readiness-v1` 已把 repository interface readiness 纳入同一 read-side 契约层；它只固定未来 `ControlPlaneReadRepository` interface method matrix、adapter gate、production auth gate、failure mapping 和 no side effects，要求消费已落地 Go type matrix 与静态 runner 证据，不创建 interface 文件、不声明 repository interface、不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-repository-adapter-implementation-readiness-refresh-v1` 已把 repository adapter implementation readiness refresh 纳入同一 read-side 契约层；它只刷新未来 adapter 实现准入、文件落点、依赖证据、七条 read route adapter 检查矩阵、failure mapping、no fake fallback 和 no side effects，要求消费 repository interface readiness、Go contract type matrix 与静态 runner 证据，不创建 interface / adapter 文件、不声明 repository interface、不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-store-selector-enablement-preconditions-v1` 已把 store selector enablement preconditions 纳入同一 read-side 契约层；它只固定未来 selector enablement gates、配置禁用态、reserved mode、unknown mode fail-closed、failure mapping、no fake fallback 和 no side effects，不创建正式配置入口、不实现 store selector、repository adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-schema-migration-implementation-preconditions-v1` 已把 schema migration implementation preconditions 纳入同一 read-side 契约层；它只固定未来 migration artifact manifest、DDL review、rollback fixture、schema version smoke、tenant index smoke、read-only role smoke、failure mapping、no fake fallback 和 no side effects，不创建 migration 目录、manifest、SQL、migration runner、store selector、repository adapter、真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-repository-adapter-implementation-plan-v1` 已把 repository adapter implementation plan 纳入同一 read-side 契约层；它只固定未来 adapter / interface / selector / migration 文件落点、依赖证据、接口 / contract type / static runner 消费、七条 read route adapter matrix、schema migration implementation preconditions 消费、selector gate、failure mapping、no fake fallback 和 no side effects，不创建 adapter/interface 文件、不写 SQL、不创建 migration、不实现 repository adapter、store selector、真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-schema-artifact-manifest-readiness-v1` 已把 schema artifact manifest readiness 纳入同一 read-side 契约层；它只固定未来 schema artifact manifest contract、DDL review evidence gate、rollback fixture evidence gate、schema version smoke、tenant index smoke、read-only role smoke、durable adapter smoke dependency、failure mapping、no fake fallback 和 no side effects，状态为 `schema_artifact_manifest_readiness_defined`；不创建 migration 目录、manifest、SQL、DDL review artifact、rollback fixture、schema smoke artifact、migration runner、repository adapter、store selector、真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-store-selector-smoke-readiness-v1` 已把 store selector smoke readiness 纳入同一 read-side 契约层；它只固定未来 selector smoke contract、模式矩阵、七条 route selector smoke matrix、reserved mode / unknown mode fail-closed、no fake fallback 和 no side effects，状态为 `store_selector_smoke_readiness_defined`；不创建 selector 文件、selector test、selector smoke fixture / checker、正式 `RADISHMIND_CONTROL_PLANE_READ_STORE` 配置入口、repository adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

当前 `control-plane-read-production-auth-readiness-v1` 已把 production auth readiness 纳入同一 read-side 契约层；它只固定未来 Radish OIDC issuer discovery evidence、token validation contract preconditions、claim mapping、tenant binding、scope projection、failure mapping、no fake fallback 和 no side effects，状态为 `production_auth_readiness_defined`；不创建 token validation schema、auth middleware、production auth smoke fixture / checker，不实现 Radish OIDC client、issuer network call、token validation、auth middleware、production API consumer、repository adapter、SQL、migration、真实数据库、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

当前 `control-plane-read-adapter-smoke-readiness-v1` 已把 adapter smoke readiness 纳入同一 read-side 契约层；它只固定未来 durable adapter smoke 如何消费 schema artifact manifest readiness、store selector smoke readiness、production auth readiness、static runner 和 repository adapter implementation plan，状态为 `adapter_smoke_readiness_defined`；不创建 adapter smoke fixture / checker、repository interface、repository adapter、adapter test、selector、SQL、migration、真实数据库、auth middleware、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

当前 `control-plane-read-implementation-trigger-review-v1` 已把 implementation trigger review 纳入同一 read-side 契约层；它只审查 schema artifact、store selector、production auth 和 adapter smoke 四类候选是否具备进入实现的触发条件，状态为 `implementation_trigger_review_defined`，当前结论为没有任何 read-side implementation trigger satisfied；不创建 migration manifest、SQL、selector、auth middleware、adapter smoke fixture / checker、repository interface、repository adapter、真实数据库、token validation 或 production API consumer。

这组 repository/read store readiness 的设计顺序是：先定义 repository contract 与 route operation matrix，再用 disabled database guard 防止误启用数据库模式，然后定义未来 contract smoke 的输入输出，随后固定 implementation readiness、store selection readiness、schema migration readiness、repository contract types readiness、repository contract types implementation、smoke runner readiness / implementation、repository interface readiness、repository adapter implementation readiness refresh、store selector enablement preconditions、schema migration implementation preconditions、repository adapter implementation plan、schema artifact manifest readiness、store selector smoke readiness、production auth readiness、adapter smoke readiness 和 implementation trigger review。任何一步都必须保留七条 read route 的 tenant predicate、sanitized projection、failure mapping、no fake fallback 和 no side effects，不能通过 fake store fallback 掩盖真实数据库、schema、adapter 或 production auth 的未就绪状态。

## Readiness checker 读法

新增的三个 readiness checker 是同一条 read-side implementation ladder 的尾部说明，不是运行时能力：

- `control-plane-read-production-auth-readiness-v1` 读取 OIDC preconditions、auth/store transition、route contract、negative contract 和 store selector smoke readiness，只证明未来 production auth 必须先具备 issuer evidence、token validation contract、claim mapping、tenant binding 和 scope projection。
- `control-plane-read-adapter-smoke-readiness-v1` 读取 repository adapter plan、schema artifact manifest readiness、store selector smoke readiness、production auth readiness 和静态 runner，只证明未来 durable adapter smoke 应怎样组合这些证据。
- `control-plane-read-implementation-trigger-review-v1` 读取 schema artifact、selector、production auth 和 adapter smoke 四类候选，当前统一判定为 `not_satisfied`，因此后续不能直接创建 implementation task card 或实现 artifact。

如果这些 checker 因 forbidden artifact、forbidden literal 或 `does_not_claim` 漂移失败，优先按“过早实现或误声明”处理：删除越界 artifact 或恢复停止线，而不是为了让检查通过去削弱 fixture。只有当外部真实证据和对应 task card 都已补齐后，才允许把某个候选从 readiness 改成实现切片。

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

`Workflow / Agent Runtime Function Surface v1` 是 read-side summary 之上的离线产品面契约，不新增 Go route 或 read store operation。它复用 `application-summary-list-route`、`workflow-definition-summary-list-route` 与 `run-record-summary-list-route` 的 TypeScript view model，再用 committed fixture 派生 application detail、definition detail、run detail、blocked action preview、confirmation placeholder、offline draft designer、offline validation inspector、execution plan preview 和 runtime readiness inspector。该层只固定页面可展示的 identity、route / request / audit metadata、risk、node / edge、timeline、draft、validation、plan、readiness 和 blocked capability 信息，不实现 workflow executor、tool executor、confirmation decision、decision store、execution unlock、builder mutation、draft persistence、validation result persistence、execution plan persistence、runtime readiness persistence、business writeback、run replay、run resume、数据库、Radish OIDC 或 production API consumer。

Workflow Surface Overview、workflow workspace context selection、Workflow Scenario Inspector 和 Workflow Review Workspace 是同一契约层上的 UI-derived organization surface。它们不扩展 read route，不创建新的 repository operation，也不改变 `contracts/typescript/control-plane-read-api.ts`；它们只把当前选中的 application、definition、run、draft 和 scenario 与 draft validation、execution plan、runtime readiness、blocked capability、stop line、route / request / audit metadata 组织成可审查视图。context selection 只改变浏览器内当前查看状态；scenario inspector 和 review workspace 只做场景解释、关系映射、blocked rollup 与 stop-line rollup，不保存 scenario / review，不发布，不执行，不提交 confirmation decision，不写回业务数据。

`control-plane-read-repository-contract-preconditions-v1` 是 auth/store transition preconditions 之后的 read store repository contract 前置切片：它固定未来 `ControlPlaneReadRepository` interface、`ReadRepositoryContext`、七条 route 到 repository operation 的映射、tenant predicate、sanitized projection、cursor/filter/sort allowlist、failure mapping 和 contract smoke 要求。该切片只定义 repository contract preconditions，不写 SQL、不建 migration、不实现 repository、不接真实数据库、不接 Radish OIDC、不实现 token validation、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-disabled-database-guard-v1` 是 repository contract preconditions 之后的 disabled database read guard 切片：它固定 database / postgres / repository read mode 当前仍是 reserved disabled，七条 read route 在误请求 database mode 时必须 fail-closed 为 `database_read_disabled`，不得静默回退到 fixture-backed fake store，也不得产生写入、executor、confirmation、writeback 或 replay 副作用。该切片不新增正式配置入口、不实现 repository adapter、不写 SQL、不建 migration、不接真实数据库、不接 Radish OIDC、不实现 token validation、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-contract-smoke-v1` 是 disabled database read guard 之后的 repository contract smoke 定义切片：它固定未来 `ControlPlaneReadRepositoryContractSmoke` 的输入字段、repository context、request / output envelope、七条 route smoke matrix、failure mapping、no fake fallback 和 no side effects。该切片不实现 smoke runner、不实现 repository adapter、不写 SQL、不建 migration、不接真实数据库、不接 Radish OIDC、不实现 token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-implementation-readiness-v1` 是 repository contract smoke 之后的 repository implementation readiness 切片：它固定未来 repository contract / adapter 文件落点、实现准入 gate、route readiness matrix、dual smoke plan、failure mapping 和停止线。该切片不创建 Go repository 文件、不实现 repository interface 或 adapter、不写 SQL、不建 migration、不接真实数据库、不接 Radish OIDC、不实现 token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-store-selection-readiness-v1` 是 repository implementation readiness 之后的 store selection readiness 切片：它固定未来 read source selector 的默认 fake-store source、database / postgres / repository reserved disabled source、未知 selector fail-closed、route selection matrix、failure mapping 和 no fake fallback。该切片不创建正式配置入口、不实现 store selector、不实现 repository interface 或 adapter、不写 SQL、不建 migration、不接真实数据库、不接 Radish OIDC、不实现 token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-schema-migration-readiness-v1` 是 store selection readiness 之后的 schema migration readiness 切片：它固定未来 read store schema ownership、migration root / naming、rollback / backup / migration lock、tenant index strategy、read-only role policy、migration smoke 和 failure mapping。该切片不创建 migration 目录、不写 SQL、不实现 migration runner、store selector、repository interface 或 adapter、不接真实数据库、不接 Radish OIDC、不实现 token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-contract-types-readiness-v1` 是 schema migration readiness 之后的 repository contract types readiness 切片：它固定未来 Go contract type 的字段与映射边界，包括 `ReadRepositoryContext`、七条 read route request / result type、failure code type、projection / filter / sort type 和 contract smoke type 输入。该切片不创建 `control_plane_read_repository_contract.go`，不实现 repository interface / adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-contract-types-implementation-v1` 是 repository contract types readiness 之后的受控实现切片：它创建 `control_plane_read_repository_contract.go` 和单元测试，只落地 `ReadRepositoryContext`、七条 read route request / result type、failure code、projection / filter / sort type 和 route type matrix。该切片不声明 repository interface，不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-contract-smoke-runner-readiness-v1` 是 repository contract types implementation 之后的 smoke runner readiness 切片：它固定未来 `ControlPlaneReadRepositoryContractSmokeRunner` 的输入输出、future 文件落点、type catalog 消费点、七条 route runner matrix、failure mapping、no fake fallback 和 no side effects。该切片不创建 `control_plane_read_repository_contract_smoke_runner.go`，不实现 smoke runner、repository interface / adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-contract-smoke-runner-implementation-v1` 是 smoke runner readiness 之后的静态 runner 实现切片：它创建 `control_plane_read_repository_contract_smoke_runner.go` 和对应 Go 测试，runner 只消费 type catalog 与既有 smoke fixture 的静态 contract case，输出 route result、failure result、contract mismatch report、side-effect report 和 summary。该切片不声明 repository interface、不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-interface-readiness-v1` 是静态 runner 之后的 repository interface readiness 切片：它固定未来 `ControlPlaneReadRepository` 的 method matrix，要求七条 read route 的 operation、request / result type 和 summary type 与 contract type implementation / static runner implementation 对齐，并保留 adapter implementation gate 与 production auth gate 为 `not_satisfied`。该切片不创建 `control_plane_read_repository_interface.go`，不声明 repository interface、不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-schema-migration-implementation-preconditions-v1` 是 store selector enablement preconditions 之后的 schema migration implementation preconditions 切片：它固定未来 migration artifact manifest、DDL review gate、rollback fixture gate、schema version smoke、tenant index smoke、read-only role smoke、route migration implementation matrix、failure mapping、no fake fallback 和 no side effects。该切片不创建 migration 目录、manifest、SQL 文件、migration runner、repository adapter、store selector、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-adapter-implementation-plan-v1` 是 schema migration implementation preconditions 之后的 repository adapter implementation plan 切片：它固定未来 adapter implementation 必须消费的 repository interface method matrix、Go contract type matrix、静态 runner、schema migration implementation preconditions 和 selector gate，并用七条 route adapter matrix 固定 future adapter checks、failure mapping、no fake fallback 和 no side effects。该切片不创建 adapter/interface 文件、不声明 repository interface、不写 SQL、不创建 migration、不实现 repository adapter、store selector、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-schema-artifact-manifest-readiness-v1` 是 repository adapter implementation plan 之后的 schema artifact manifest readiness 切片：它固定未来 schema artifact manifest contract、DDL review evidence gate、rollback fixture evidence gate、schema version smoke、tenant index smoke、read-only role smoke、durable adapter smoke dependency、route schema artifact matrix、failure mapping、no fake fallback 和 no side effects，状态为 `schema_artifact_manifest_readiness_defined`。该切片不创建 migration 目录、manifest、SQL、DDL review artifact、rollback fixture、schema smoke artifact、migration runner、repository adapter、store selector、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-store-selector-smoke-readiness-v1` 是 schema artifact manifest readiness 之后的 store selector smoke readiness 切片：它固定未来 selector smoke contract、模式矩阵、七条 route selector smoke matrix、reserved mode fail-closed、unknown mode fail-closed、no fake fallback 和 no side effects，状态为 `store_selector_smoke_readiness_defined`。该切片不创建 `control_plane_read_store_selector.go`、selector test、selector smoke fixture / checker，不新增正式 `RADISHMIND_CONTROL_PLANE_READ_STORE` 配置入口，不实现 store selector、repository adapter、database query、SQL、migration、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-production-auth-readiness-v1` 是 store selector smoke readiness 之后的 production auth readiness 切片：它固定未来 issuer discovery evidence、token validation contract preconditions、claim mapping、tenant binding、scope projection、failure mapping、no fake fallback 和 no side effects，状态为 `production_auth_readiness_defined`。该切片不创建 token validation schema、auth middleware、production auth smoke fixture / checker，不接真实 OIDC、不执行 token validation、不启用 production API consumer、repository adapter、SQL、migration、真实数据库或任何写入 / 执行能力。

`control-plane-read-adapter-smoke-readiness-v1` 是 production auth readiness 之后的 adapter smoke readiness 切片：它固定未来 durable adapter smoke 如何消费 schema artifact manifest readiness、store selector smoke readiness、production auth readiness、static runner 和 repository adapter implementation plan，状态为 `adapter_smoke_readiness_defined`。该切片不创建 adapter smoke fixture / checker、repository interface、repository adapter、adapter test、selector、SQL、migration、真实数据库、auth middleware、token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-implementation-trigger-review-v1` 是 adapter smoke readiness 之后的 implementation trigger review 切片：它固定 schema artifact、store selector、production auth 和 adapter smoke 四类候选的触发条件审查顺序、当前 blocker 和停止线，状态为 `implementation_trigger_review_defined`。该切片不创建 implementation task card、migration manifest、SQL、selector、auth middleware、adapter smoke fixture / checker、repository interface、repository adapter、真实数据库、token validation、production API consumer、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

## 阶段门禁调整

`admin-audit-log` 已完成当前 read-side UI 页面集合的最后一个优先页面。普通 read-only UI 展示页后续不再默认逐项新增独立 task card、fixture 和 checker，应通过 `control-plane-read-formal-ui-readiness-close-v1` 一类聚合收口，统一校验页面到 `CONTROL_PLANE_READ_ROUTES` / `CONTROL_PLANE_READ_ROUTE_DEFINITIONS` 的绑定、状态组件、forbidden output guard、request / audit ref 和停止线。

`control-plane-read-dev-live-consumer-v1` 已把“不请求 live backend”放宽为显式 dev-only live read path。它只能用于验证 `apps/radishmind-web/` 通过 HTTP 消费现有 read API shape，并且只能连接 fake-store-backed handler 与测试身份上下文。该放宽不允许接入真实数据库、Radish OIDC、repository migration、API key lifecycle、quota enforcement、billing / cost ledger、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-auth-store-transition-preconditions-v1` 把 dev-only live read path 之后的下一步限定为迁移前置治理：先证明 auth context、repository interface、tenant predicate、sanitized projection、dual smoke 和 fail-closed guard 都已定义，再谈真实 middleware 或 store 实现。该 gate 不允许在同一切片内加入数据库 query、OIDC validation、repository migration 或生产 consumer。

`control-plane-read-repository-contract-preconditions-v1` 将停止线前移到 repository contract 层：可以定义 interface、operation matrix、projection / filter / sort allowlist 和 failure mapping，但仍不允许在同一切片内加入 SQL、migration、repository implementation、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-disabled-database-guard-v1` 将停止线继续前移到 database mode guard 层：可以声明 database / postgres / repository read mode 是 reserved disabled，并固定 `database_read_disabled` fail-closed 行为，但仍不允许新增正式配置启用入口、SQL、migration、repository adapter、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-repository-contract-smoke-v1` 将停止线固定到 future smoke 层：可以定义未来 repository contract smoke 的输入输出、route 覆盖、failure mapping、no fake fallback 和 no side effects，但仍不允许在同一切片内加入 SQL、migration、repository adapter、真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-repository-implementation-readiness-v1` 将停止线固定到 implementation readiness 层：可以定义未来文件落点、实现准入、dual smoke 和 failure mapping，但仍不允许在同一切片内创建 Go repository 文件、加入 SQL、migration、repository adapter、真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-store-selection-readiness-v1` 将停止线固定到 store selection readiness 层：可以定义 future read source selector 的默认值、保留值、失败映射和 route matrix，但仍不允许在同一切片内创建正式配置入口、实现 selector、加入 SQL、migration、repository adapter、真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-schema-migration-readiness-v1` 将停止线固定到 schema migration readiness 层：可以定义 schema ownership、migration layout、rollback、tenant index、read-only role、migration smoke 和 failure mapping，但仍不允许在同一切片内创建 migration 目录、写 SQL、实现 migration runner、store selector、repository adapter、真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-repository-contract-types-readiness-v1` 将停止线固定到 repository contract types readiness 层：可以定义未来 Go contract type 的 required fields、route request / result type、failure code type、projection / filter / sort type 和 contract smoke type 输入，但仍不允许在同一切片内创建 Go contract 文件、实现 repository interface、repository adapter、store selector、写 SQL、建 migration、接真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-repository-contract-types-implementation-v1` 将停止线固定到 repository contract types implementation 层：可以创建受控 Go contract type 文件和类型测试，但仍不允许在同一切片内声明 repository interface、实现 repository adapter、store selector、写 SQL、建 migration、接真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-repository-contract-smoke-runner-readiness-v1` 将停止线固定到 smoke runner readiness 层：可以定义 future runner 如何消费 type matrix 和既有 smoke fixture，但仍不允许在同一切片内创建 runner 文件、声明 repository interface、实现 repository adapter、store selector、写 SQL、建 migration、接真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-repository-contract-smoke-runner-implementation-v1` 将停止线固定到静态 runner 层：可以创建 runner 文件、静态 smoke case、Go 测试和 implementation checker，但仍不允许在同一切片内声明 repository interface、实现 repository adapter、store selector、写 SQL、建 migration、接真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-repository-interface-readiness-v1` 将停止线固定到 repository interface readiness 层：可以定义 future interface method matrix、future 文件落点和 adapter gate，但仍不允许在同一切片内创建 interface 文件、声明 repository interface、实现 repository adapter、store selector、写 SQL、建 migration、接真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-repository-adapter-implementation-readiness-refresh-v1` 将停止线固定到 repository adapter implementation readiness refresh 层：可以刷新 future adapter gate、依赖证据、文件落点和七条 route adapter 检查矩阵，但仍不允许在同一切片内创建 interface / adapter 文件、声明 repository interface、实现 repository adapter、store selector、写 SQL、建 migration、接真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-store-selector-enablement-preconditions-v1` 将停止线固定到 store selector enablement preconditions 层：可以定义 future selector enablement gates、配置禁用态、reserved mode、unknown mode fail-closed、failure mapping 和 no fake fallback，但仍不允许在同一切片内创建正式配置入口、实现 store selector、repository adapter、写 SQL、建 migration、接真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-schema-migration-implementation-preconditions-v1` 将停止线固定到 schema migration implementation preconditions 层：可以定义 future migration artifact manifest、DDL review、rollback fixture、schema version smoke、tenant index smoke、read-only role smoke 和 route migration implementation matrix，但仍不允许在同一切片内创建 migration 目录、manifest、SQL、migration runner、repository adapter、store selector、接真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-repository-adapter-implementation-plan-v1` 将停止线固定到 repository adapter implementation plan 层：可以定义 future adapter/interface/selector/migration 文件落点、依赖证据、接口 / contract type / static runner 消费、schema migration implementation preconditions 消费、selector gate 和七条 route adapter matrix，但仍不允许在同一切片内创建 adapter/interface 文件、写 SQL、建 migration、实现 repository adapter、store selector、接真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-schema-artifact-manifest-readiness-v1` 将停止线固定到 schema artifact manifest readiness 层：可以定义 future schema artifact manifest contract、DDL review evidence、rollback fixture evidence、schema version smoke、tenant index smoke、read-only role smoke 和 durable adapter smoke dependency，但仍不允许在同一切片内创建 migration 目录、manifest、SQL、DDL review artifact、rollback fixture、schema smoke artifact、migration runner、repository adapter、store selector、接真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-store-selector-smoke-readiness-v1` 将停止线固定到 store selector smoke readiness 层：可以定义 future selector smoke contract、模式矩阵、reserved / unknown mode fail-closed 和七条 route selector smoke matrix，但仍不允许在同一切片内创建 selector 文件、selector test、selector smoke fixture / checker、正式配置入口、repository adapter、SQL、migration、接真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-production-auth-readiness-v1` 将停止线固定到 production auth readiness 层：可以定义 future issuer discovery evidence、token validation contract preconditions、claim mapping、tenant binding、scope projection 和 failure mapping，但仍不允许在同一切片内创建 token validation schema、auth middleware、production auth smoke fixture / checker、接真实 OIDC、执行 token validation、启用 production API consumer、repository adapter、SQL、migration、接真实数据库或任何写入 / 执行能力。

`control-plane-read-adapter-smoke-readiness-v1` 将停止线固定到 adapter smoke readiness 层：可以定义未来 durable adapter smoke 如何消费 schema artifact manifest readiness、store selector smoke readiness、production auth readiness、static runner 和 repository adapter implementation plan，但仍不允许在同一切片内创建 adapter smoke fixture / checker、repository interface、repository adapter、adapter test、selector、SQL、migration、接真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

`control-plane-read-implementation-trigger-review-v1` 将停止线固定到 implementation trigger review 层：可以审查 schema artifact、selector、production auth 和 adapter smoke 是否具备实现触发条件，但仍不允许在同一切片内创建 implementation task card、migration manifest、SQL、selector、auth middleware、adapter smoke fixture / checker、repository interface、repository adapter、接真实数据库、OIDC validation、production API consumer 或任何写入 / 执行能力。

## 分层关系

当前 read-side 契约按四十二层固定：

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
28. `control-plane-read-schema-migration-readiness-v1`
   - 固定未来 schema migration readiness 的 schema ownership、migration layout、rollback plan、tenant index strategy、read-only role policy、migration smoke、failure mapping 和停止线。
   - 不创建 migration 目录、不写 SQL、不实现 migration runner、store selector、repository interface 或 adapter、不接真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
29. `control-plane-read-repository-contract-types-readiness-v1`
   - 固定未来 repository contract types readiness：`ReadRepositoryContext`、七条 route request / result type、failure code type、projection / filter / sort type 和 contract smoke type 输入。
   - 不创建 Go contract type 文件、不实现 repository interface、repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
30. `control-plane-read-repository-contract-types-implementation-v1`
   - 创建受控 Go contract type 文件和测试，落地 `ReadRepositoryContext`、七条 route request / result type、failure code、projection / filter / sort type 和 route type matrix。
   - 不声明 repository interface、不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
31. `control-plane-read-repository-contract-smoke-runner-readiness-v1`
   - 固定未来 smoke runner readiness：`ControlPlaneReadRepositoryContractSmokeRunner` 如何消费 type matrix、既有 smoke fixture、failure mapping、no fake fallback 和 no side effects。
   - 不创建 runner 文件、不声明 repository interface、不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
32. `control-plane-read-repository-contract-smoke-runner-implementation-v1`
   - 创建静态 repository contract smoke runner 和测试，消费 `controlPlaneReadRepositoryRouteTypeContracts()` 并与既有 smoke fixture 对齐。
   - 不声明 repository interface、不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
33. `control-plane-read-repository-interface-readiness-v1`
   - 固定未来 `ControlPlaneReadRepository` interface method matrix 和 adapter gate，消费已落地 Go type matrix 与静态 runner 证据。
   - 不创建 interface 文件、不声明 repository interface、不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
34. `control-plane-read-repository-adapter-implementation-readiness-refresh-v1`
   - 固定未来 repository adapter implementation readiness refresh，消费 repository interface readiness、Go contract type matrix、静态 runner、schema migration readiness、store selection readiness 和 disabled database guard 证据。
   - 不创建 interface / adapter 文件、不声明 repository interface、不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
35. `control-plane-read-store-selector-enablement-preconditions-v1`
   - 固定未来 store selector enablement preconditions，保留 fixture fake store 默认路径、reserved disabled modes、unknown mode fail-closed、failure mapping 和 no fake fallback。
   - 不创建正式配置入口、不实现 store selector、repository adapter、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
36. `control-plane-read-schema-migration-implementation-preconditions-v1`
   - 固定未来 schema migration implementation preconditions：migration artifact manifest、DDL review、rollback fixture、schema version smoke、tenant index smoke、read-only role smoke、route migration implementation matrix、failure mapping 和 no fake fallback。
   - 不创建 migration 目录、manifest、SQL、migration runner、repository adapter、store selector、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
37. `control-plane-read-repository-adapter-implementation-plan-v1`
   - 固定未来 repository adapter implementation plan：文件落点、依赖证据、接口 / contract type / static runner 消费、schema migration implementation preconditions 消费、selector gate、七条 read route adapter matrix、failure mapping 和 no fake fallback。
   - 不创建 adapter/interface 文件、不写 SQL、不创建 migration、不实现 repository adapter、store selector、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
38. `control-plane-read-schema-artifact-manifest-readiness-v1`
   - 固定未来 schema artifact manifest readiness：manifest contract、DDL review evidence、rollback fixture evidence、schema version smoke、tenant index smoke、read-only role smoke、durable adapter smoke dependency、七条 route schema artifact matrix、failure mapping 和 no fake fallback。
   - 不创建 migration 目录、manifest、SQL、DDL review artifact、rollback fixture、schema smoke artifact、migration runner、repository adapter、store selector、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
39. `control-plane-read-store-selector-smoke-readiness-v1`
   - 固定未来 store selector smoke readiness：selector smoke contract、模式矩阵、七条 route selector smoke matrix、reserved mode / unknown mode fail-closed、failure mapping 和 no fake fallback。
   - 不创建 selector 文件、selector test、selector smoke fixture / checker、正式配置入口、repository adapter、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
40. `control-plane-read-production-auth-readiness-v1`
   - 固定未来 production auth readiness：issuer discovery evidence、token validation contract preconditions、claim mapping、tenant binding、scope projection、failure mapping 和 no fake fallback。
   - 不创建 token validation schema、auth middleware、production auth smoke fixture / checker，不实现 Radish OIDC client、issuer network call、token validation、production API consumer、repository adapter、SQL、migration、真实数据库、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
41. `control-plane-read-adapter-smoke-readiness-v1`
   - 固定未来 adapter smoke readiness：durable adapter smoke 必须消费 schema artifact manifest readiness、store selector smoke readiness、production auth readiness、static runner 和 repository adapter implementation plan。
   - 不创建 adapter smoke fixture / checker、repository interface、repository adapter、adapter test、selector、SQL、migration、真实数据库、auth middleware、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。
42. `control-plane-read-implementation-trigger-review-v1`
   - 固定 implementation trigger review：schema artifact、store selector、production auth 和 adapter smoke 四类候选均为 `not_satisfied`。
   - 不创建 implementation task card、migration manifest、SQL、selector、auth middleware、adapter smoke fixture / checker、repository interface、repository adapter、真实数据库、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

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
- `scripts/checks/fixtures/control-plane-read-schema-migration-readiness-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-schema-migration-readiness-v1.py`
- `scripts/checks/fixtures/control-plane-read-repository-contract-types-readiness-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-repository-contract-types-readiness-v1.py`
- `scripts/checks/fixtures/control-plane-read-repository-contract-types-implementation-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-repository-contract-types-implementation-v1.py`
- `scripts/checks/fixtures/control-plane-read-repository-contract-smoke-runner-readiness-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-repository-contract-smoke-runner-readiness-v1.py`
- `scripts/checks/fixtures/control-plane-read-repository-contract-smoke-runner-implementation-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-repository-contract-smoke-runner-implementation-v1.py`
- `scripts/checks/fixtures/control-plane-read-repository-interface-readiness-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-repository-interface-readiness-v1.py`
- `scripts/checks/fixtures/control-plane-read-schema-migration-implementation-preconditions-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-schema-migration-implementation-preconditions-v1.py`
- `scripts/checks/fixtures/control-plane-read-repository-adapter-implementation-plan-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-repository-adapter-implementation-plan-v1.py`
- `scripts/checks/fixtures/control-plane-read-schema-artifact-manifest-readiness-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-schema-artifact-manifest-readiness-v1.py`
- `scripts/checks/fixtures/control-plane-read-store-selector-smoke-readiness-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-store-selector-smoke-readiness-v1.py`
- `scripts/checks/fixtures/control-plane-read-production-auth-readiness-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-production-auth-readiness-v1.py`
- `scripts/checks/fixtures/control-plane-read-adapter-smoke-readiness-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-adapter-smoke-readiness-v1.py`
- `scripts/checks/fixtures/control-plane-read-implementation-trigger-review-v1.json`
- `scripts/checks/control_plane/check-control-plane-read-implementation-trigger-review-v1.py`
- `contracts/typescript/control-plane-read-api.ts`
- `scripts/run-control-plane-read-consumer-smoke.py`

这些 checker 已接入 `scripts/check-repo.py --fast`。它们的作用是防止契约、样例、负向边界、实现前置条件、fake-store-backed read handler plan、handler implementation、auth/db preconditions、consumer contract、正式 UI 边界、正式 UI 实现 readiness、shared read shell、admin tenant overview、workspace applications、workspace api keys、workspace usage quota、workspace workflow definitions、workspace run history、admin audit log 页面切片、formal UI readiness close、dev-only live consumer、auth/store transition preconditions、repository contract preconditions、disabled database read guard、repository contract smoke、repository implementation readiness、store selection readiness、schema migration readiness、schema migration implementation preconditions、repository adapter implementation plan、schema artifact manifest readiness、store selector smoke readiness、production auth readiness、adapter smoke readiness、implementation trigger review 和文档说明互相漂移，不负责启动服务或模拟真实数据库。

Workflow function surface 的 fixture / checker 同样位于 `scripts/checks/fixtures/` 与 `scripts/checks/control_plane/`，覆盖 `workflow-function-surface-boundary-v1`、application detail、definition detail、run detail、blocked action preview、confirmation placeholder、offline draft designer 和 offline draft validation inspector。它们验证离线 view model、UI panel、blocked capability 和停止线一致，不启动 runtime、不请求真实后端、不写入 draft / validation / decision / run 状态。

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
- 不把 `control-plane-read-schema-migration-readiness-v1` 写成 database schema ready、migration files created、migration runner ready、repository adapter ready、database query ready、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-repository-contract-types-readiness-v1` 写成 Go contract type file ready、repository interface ready、repository adapter ready、store selector ready、database query ready、database migration ready、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-repository-contract-types-implementation-v1` 写成 repository interface ready、repository adapter ready、store selector ready、database query ready、database migration ready、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-repository-contract-smoke-runner-readiness-v1` 写成 smoke runner implemented、repository interface ready、repository adapter ready、store selector ready、database query ready、database migration ready、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-repository-contract-smoke-runner-implementation-v1` 写成 repository interface ready、repository adapter ready、store selector ready、database query ready、database migration ready、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-repository-interface-readiness-v1` 写成 repository interface ready、repository adapter ready、store selector ready、database query ready、database migration ready、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-schema-migration-implementation-preconditions-v1` 写成 migration files created、migration runner ready、database schema ready、database query ready、database migration ready、repository adapter ready、store selector ready、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-repository-adapter-implementation-plan-v1` 写成 repository interface ready、repository adapter ready、store selector ready、database query ready、database migration ready、migration files created、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-schema-artifact-manifest-readiness-v1` 写成 schema artifact manifest ready、DDL review ready、rollback fixture ready、schema version smoke ready、tenant index smoke ready、read-only role smoke ready、database schema ready、database migration ready、repository adapter ready、store selector ready、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-store-selector-smoke-readiness-v1` 写成 store selector smoke ready、store selector implemented、formal read store config ready、repository adapter ready、database query ready、database migration ready、Radish OIDC ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-production-auth-readiness-v1` 写成 production auth ready、Radish OIDC client ready、token validation ready、auth middleware ready、production API consumer ready、repository adapter ready、database query ready、database migration ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-adapter-smoke-readiness-v1` 写成 adapter smoke ready、repository interface ready、repository adapter ready、database query ready、database migration ready、store selector ready、production auth ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `control-plane-read-implementation-trigger-review-v1` 写成 implementation trigger satisfied、schema artifact ready、store selector ready、production auth ready、adapter smoke ready、repository interface ready、repository adapter ready、database query ready、database migration ready、token validation ready、production API consumer ready、API key lifecycle ready、quota enforcement ready、billing ready、cost ledger ready、workflow executor ready、confirmation ready、writeback ready、replay ready 或 production ready。
- 不把 `Workflow / Agent Runtime Function Surface v1` 的 application / definition / run / draft / validation / execution plan / runtime readiness / overview / scenario / review workspace 面板写成 runtime API、builder mutation、draft persistence、validation result persistence、execution plan persistence、runtime readiness persistence、scenario / review persistence、executor、confirmation decision、business writeback、run replay、run resume、真实数据库、Radish OIDC 或 production API consumer ready。
- 不把 `apps/radishmind-web/` 首个 shared shell 写成完整 product UI app。
- 不把 `apps/radishmind-console/` 改写成正式用户端、生产管理端或 control plane write surface。
- 不把 `scripts/run-control-plane-read-consumer-smoke.py` 写成真实 API smoke；它只消费 committed fixture。
- 不通过 read-side 契约启用 OIDC、API key lifecycle、quota enforcement、billing ledger、workflow executor、confirmation、business writeback 或 replay。
- 不把本地 ops console 提升为正式 user workspace 或 production admin console。
