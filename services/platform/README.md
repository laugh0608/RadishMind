# RadishMind Platform Service Layer
<!-- markdown-size-allow: 平台历史运行说明仍被多项 runbook checker 直接读取；人工按当前功能定位阅读，R6 将拆分配置、路由与历史 readiness 章节。 -->
本目录承载 `Go` 平台服务层的最小骨架。

当前职责：

- 启动最小本地 `HTTP` 服务
- 承载 northbound `API` / `gateway` 入口
- 通过 Python bridge 调用 canonical `CopilotGatewayEnvelope`
- 提供结构化诊断、观测、本地产品 overview、部署壳和后续鉴权 / 流式转发落点

当前明确不做：

- 不在这里复制第二套业务真相源
- 不在这里重写模型推理、训练、评测或 `builder`
- 不绕过 `contracts/` 自定义另一套 canonical protocol

当前 HTTP 路由：

- `GET /healthz`
- `GET /v1/platform/overview`
- `GET /v1/platform/local-smoke`
- `GET /v1/models`
- `GET /v1/models/{id}`
- `POST /v1/chat/completions`
- `POST /v1/responses`
- `POST /v1/messages`
- `GET /v1/session/metadata`
- `GET /v1/session/recovery/checkpoints/{checkpoint_id}`
- `GET /v1/tools/metadata`
- `POST /v1/tools/actions`
- `GET /v1/control-plane/tenants/{tenant_ref}/summary`
- `GET /v1/user-workspace/applications`
- `GET /v1/user-workspace/api-keys`
- `GET /v1/user-workspace/usage/quota-summary`
- `GET /v1/user-workspace/workflow-definitions`
- `GET /v1/user-workspace/runs`
- `GET /v1/control-plane/audit`
- `POST /v1/user-workspace/workflow-drafts`
- `GET /v1/user-workspace/workflow-drafts`
- `GET /v1/user-workspace/workflow-drafts/{draft_id}`
- `POST /v1/user-workspace/workflow-drafts/validate`
- `POST /v1/user-workspace/application-drafts/validate`
- `POST /v1/user-workspace/application-drafts`
- `GET /v1/user-workspace/application-drafts`
- `GET /v1/user-workspace/application-drafts/{draft_id}`
- `POST /v1/user-workspace/application-publish-candidates`
- `GET /v1/user-workspace/application-publish-candidates`
- `GET /v1/user-workspace/application-publish-candidates/{candidate_id}`
- `POST /v1/user-workspace/application-publish-candidates/{candidate_id}/reviews`
- `POST /v1/user-workspace/workflow-drafts/{draft_id}/runs`
- `GET /v1/user-workspace/workflow-runs`
- `GET /v1/user-workspace/workflow-runs/{run_id}`
- `GET /v1/user-workspace/workflow-runs/{candidate_run_id}/comparison`
- `POST /v1/user-workspace/workflow-evaluation-cases`
- `GET /v1/user-workspace/workflow-evaluation-cases`
- `GET /v1/user-workspace/workflow-evaluation-cases/{case_id}`
- `GET /v1/user-workspace/workflow-evaluation-cases/{case_id}/review`
- `POST /v1/user-workspace/workflow-evaluation-cases/{case_id}/revisions`
- `GET /v1/user-workspace/workflow-evaluation-cases/{case_id}/revisions`
- `GET /v1/user-workspace/workflow-evaluation-cases/{case_id}/revisions/{version}`
- `POST /v1/user-workspace/workflow-evaluation-suites`
- `GET /v1/user-workspace/workflow-evaluation-suites`
- `GET /v1/user-workspace/workflow-evaluation-suites/{suite_id}`
- `GET /v1/user-workspace/workflow-evaluation-suites/{suite_id}/review`
- `POST /v1/user-workspace/workflow-evaluation-suites/{suite_id}/decisions`
- `GET /v1/user-workspace/workflow-evaluation-suites/{suite_id}/decisions`
- `GET /v1/model-gateway/requests`
- `GET /v1/model-gateway/requests/{request_id}`

路由是否可用由各自的显式 dev/test gate 决定；注册到 mux 不等于默认开放。Application Draft、Application Publish、Workflow Executor / Evaluation 和 Gateway Request History 的开关、store selector、数据库连接与失败语义见本文后面的配置表。所有 `postgres_dev_test` store 都要求对应 manual migration marker / checksum preflight，连接或 preflight 失败时不得回退 `memory_dev`。

其中 `GET /v1/platform/overview` 是 `P3 Local Product Shell / Ops Surface` 的首个只读产品面入口：它汇总服务状态、可选 model/profile、session/tooling metadata route、blocked action route 和当前停止线，供 `apps/radishmind-console/` 本地控制台或上层 UI 一次读取。它不启用真实 executor、durable store、confirmation 接线、长期记忆、业务写回或 replay。

`GET /v1/platform/local-smoke` 是本地开发 readiness 摘要入口：它聚合 `/healthz`、overview contract、model inventory、session/tooling metadata、blocked action no-side-effects、local console CORS origin 和停止线状态，便于开发者或轻量脚本一次判断默认 `7000/4000` 本地 console 链路是否可读。它不启动或守护进程，不实现 production health dashboard、executor、durable store、confirmation、业务写回或 replay。

`/v1/chat/completions`、`/v1/responses` 和 `/v1/messages` 已接到最小 canonical bridge：`Go` 只负责 northbound 请求翻译、provider 选择和进程调度，真正的 canonical request / response 语义仍由 Python runtime 与 gateway 维持。

`GET /v1/session/recovery/checkpoints/{checkpoint_id}` 当前只是 session recovery checkpoint 的 metadata-only route smoke：它返回固定 fixture 边界、checkpoint refs、tool audit refs、`tool_audit_summary`、replay policy 摘要和 state summary，不读取 durable checkpoint store，不返回 materialized tool result，也不执行跨轮 replay。该 route 会拒绝 materialized result 和 replay 类查询参数，例如 `include_materialized_results=true`、`include_tool_results=true` 或 `auto_replay=true`。

`GET /v1/session/metadata`、`GET /v1/tools/metadata` 与 `POST /v1/tools/actions` 当前构成最小 session/tooling 可用外壳：前两者返回平台可消费的 session 扩展字段、history/state/recovery 边界、tool registry metadata 和 contract-only execution policy；后者对任何工具 action 请求都返回 `tool_action_blocked_response`，明确 `status=blocked`、`execution_enabled=false`、`executed=false`、`result_ref=null`、`durable_memory_written=false`、`writes_business_truth=false`。这些路由只用于上层或 UI 发现能力和展示 blocked action 状态，不启用真实 executor、durable store、confirmation 接线、长期记忆、业务写回或 replay。

`control-plane-read-fake-store-handler-plan-v1` 与 `control-plane-read-fake-store-handler-implementation-v1` 继续作为历史 contract evidence 保留；当前七条 read route 已统一经过 shared verified identity 与 route authorization。`RADISHMIND_CONTROL_PLANE_READ_STORE=fake_store_dev` 仍服务显式开发测试路径；`postgres_dev_test` 只把 Tenant Summary 与 Audit 路由到 PostgreSQL read repository，五条 workspace operation 在 signed test 模式仍使用 fake binding。`radish_oidc_integration_test` 只允许与 `postgres_dev_test` 组合：Tenant Summary / Audit 分别要求映射后的 tenant-read / audit-read permission，五条 workspace operation 因 membership owner 未成立统一返回 `workspace_membership_unavailable`，不得读取 fake repository。任何 identity、tenant、permission、membership 或 provider denial 都必须在 repository query 前结束。

`control-plane-verified-identity-negative-auth-runtime-v1` 已在同一七条 route 上新增 `disabled / dev_headers / signed_test_token` auth mode、shared verified identity / resource binding、`RS256` test-token verifier、显式 permission projection 和 sanitized `401 / 403` envelope。后续 `radish_oidc_integration_test` 已完成 deterministic discovery / JWKS / JWT verifier、两条 Admin route 和五条 workspace membership fail-closed；signed test verifier 仍不连接 discovery / JWKS，OIDC integration 也不代表真实 Radish 联调或 production auth。

`control-plane-read-auth-db-preconditions-v1` 当前只固定未来替换 fake auth / fake store 前必须满足的前置条件：read route 必须等 `future Radish OIDC / auth middleware` 注入 identity、tenant、subject、scope、audit 和 request context，真实 read store 必须先定义 `future control plane read store repository` 的窄接口、tenant predicate、sanitized projection、failure taxonomy 和 smoke transition plan。它不实现 OIDC validation、数据库 schema / migration / query、repository、API key lifecycle、quota enforcement、executor、confirmation、writeback 或 replay。

`control-plane-read-auth-store-transition-preconditions-v1` 当前只固定 dev fake auth / fixture-backed fake store 迁移到未来 auth middleware / read store repository 前的 auth/store transition preconditions：auth middleware 必须先有 issuer discovery、token validation contract、claim mapping、tenant binding、scope projection、audit context 和 fail-closed gate；read store 必须先有 repository interface、tenant predicate、sanitized summary projection、cursor/filter/sort allowlist、failure taxonomy、no database write 和 no secret material gate。它不实现 Radish OIDC、token validation、数据库 schema / migration / query、repository implementation、production API consumer、API key lifecycle、quota enforcement、executor、confirmation、writeback 或 replay。

`control-plane-read-repository-contract-preconditions-v1` 当前只固定未来 read store repository contract 的更具体前置条件：`ControlPlaneReadRepository` interface、`ReadRepositoryContext`、七条 read route 到 repository operation 的映射、tenant predicate、sanitized projection、cursor/filter/sort allowlist、failure mapping 和 contract smoke 要求。它不实现 repository interface、不写 SQL、不建 migration、不接真实数据库、不接 Radish OIDC、不启用 production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-disabled-database-guard-v1` 当前只固定 disabled database read guard：database / postgres / repository read mode 仍是 reserved disabled，未来误请求必须 fail-closed 为 `database_read_disabled`，不得静默回退到 fixture-backed fake store。它不新增正式配置入口、不实现 repository adapter、不写 SQL、不建 migration、不接真实数据库、不接 Radish OIDC、不启用 production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-contract-smoke-v1` 当前只固定未来 repository contract smoke：输入字段、repository context、request / output envelope、七条 read route 覆盖、failure mapping、no fake fallback 和 no side effects。该切片本身不声明 repository interface、不写 SQL、不建 migration、不接真实数据库、不接 Radish OIDC、不启用 token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-implementation-readiness-v1` 当前只固定 repository implementation readiness：未来文件落点、实现准入 gate、七条 route readiness matrix、dual smoke plan、failure mapping、no fake fallback、no side effects 和停止线。该切片本身不创建 Go repository 文件、不声明 repository interface、不实现 repository adapter、不写 SQL、不建 migration、不接真实数据库、不接 Radish OIDC、不启用 token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-store-selection-readiness-v1` 当前只固定 store selection readiness：默认 read source 仍为 fixture-backed fake store，未来 `RADISHMIND_CONTROL_PLANE_READ_STORE` 只作为保留配置键，database / postgres / repository mode 必须 fail-closed 为 `database_read_disabled`，未知 selector value 必须 fail-closed 为 `invalid_read_store_mode`。该切片本身不创建正式配置入口、不声明 read repository interface、不实现 store selector、repository adapter、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-schema-migration-readiness-v1` 当前只固定 schema migration readiness：schema ownership、migration layout、rollback plan、tenant index strategy、read-only role policy、migration smoke、failure mapping 和停止线。该切片本身不创建 migration 目录、不声明 read repository interface、不写 SQL、不实现 migration runner、store selector、repository adapter、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-contract-types-readiness-v1` 当前只固定 repository contract types readiness：未来 `ReadRepositoryContext`、七条 read route request / result type、failure code type、projection / filter / sort type 和 contract smoke type 输入已定义。该切片本身不创建 `control_plane_read_repository_contract.go`、不声明 read repository interface、不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-contract-types-implementation-v1` 当前只固定 repository contract types implementation：`control_plane_read_repository_contract.go` 已创建，包含 `ReadRepositoryContext`、七条 read route request / result type、failure code、projection / filter / sort type 和 route type matrix。它不声明 `ControlPlaneReadRepository` interface，不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-contract-smoke-runner-readiness-v1` 当前只固定 repository contract smoke runner readiness：未来 `ControlPlaneReadRepositoryContractSmokeRunner` 如何消费 `controlPlaneReadRepositoryRouteTypeContracts()`、既有 smoke fixture、failure mapping、no fake fallback 和 no side effects 已进入 fast baseline。该切片本身不创建 runner 文件、不声明 repository interface、不实现 runner、repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-contract-smoke-runner-implementation-v1` 当前只固定 repository contract smoke runner implementation：`control_plane_read_repository_contract_smoke_runner.go` 已创建，静态 runner 消费 `controlPlaneReadRepositoryRouteTypeContracts()`，并用 Go 测试与既有 smoke fixture 对齐七条 read route、failure mapping、no fake fallback 和 no side effects。它不声明 `ControlPlaneReadRepository` interface，不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-interface-readiness-v1` 当前只固定 repository interface readiness：未来 `ControlPlaneReadRepository` 的 method matrix 必须消费已落地的 Go contract type 和静态 runner 证据，七条 read route 的 request / result type、failure mapping、no side effects 和 adapter gate 已进入 fast baseline。它不创建 `control_plane_read_repository_interface.go`，不声明 repository interface，不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-adapter-implementation-readiness-refresh-v1` 当前只固定 repository adapter implementation readiness refresh：未来 adapter gate 必须消费 repository interface readiness、Go contract type matrix、静态 runner、schema migration readiness、store selection readiness 和 disabled database guard 证据，七条 read route 的 future adapter 检查矩阵已进入 fast baseline。它不创建 `control_plane_read_repository_interface.go`、`control_plane_read_repository_adapter.go` 或 adapter test，不声明 repository interface，不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-store-selector-enablement-preconditions-v1` 当前只固定 store selector enablement preconditions：未来 selector gate 必须保留 fixture fake store 默认路径，database / postgres / repository 继续 reserved disabled，unknown mode 必须 fail-closed，且 reserved mode 不得回退到 fixture fake store。它不创建 `control_plane_read_store_selector.go`，不新增正式 `RADISHMIND_CONTROL_PLANE_READ_STORE` 运行配置入口，不实现 selector、repository adapter、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-schema-migration-implementation-preconditions-v1` 当前只固定 schema migration implementation preconditions：未来 migration artifact manifest、DDL review、rollback fixture、schema version smoke、tenant index smoke 和 read-only role smoke 必须先于 migration 文件创建。它不创建 migration 目录、manifest、SQL 或 `ControlPlaneReadSchemaMigrationRunner`，不实现 migration runner、repository adapter、store selector、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-repository-adapter-implementation-plan-v1` 当前只固定 repository adapter implementation plan：未来 adapter 实现必须消费 repository interface readiness、Go contract type matrix、静态 runner、schema migration implementation preconditions 和 store selector enablement preconditions，并覆盖七条 read route adapter matrix、selector gate、failure mapping、no fake fallback 和 no side effects。它不创建 `control_plane_read_repository_interface.go`、`control_plane_read_repository_adapter.go`、adapter test、selector 文件、migration 目录、manifest 或 SQL，不实现 repository adapter、store selector、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-schema-artifact-manifest-readiness-v1` 当前只固定 schema artifact manifest readiness，状态为 `schema_artifact_manifest_readiness_defined`：未来 schema artifact manifest contract、DDL review evidence、rollback fixture evidence、schema version smoke、tenant index smoke、read-only role smoke 和 durable adapter smoke dependency 必须先于 schema artifact、migration 或 adapter 文件创建。它不创建 migration 目录、manifest、SQL、DDL review artifact、rollback fixture、schema smoke artifact 或 `ControlPlaneReadSchemaMigrationRunner`，不实现 migration runner、repository adapter、store selector、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-schema-artifact-evidence-v1` 当前只固定 schema artifact evidence，状态为 `schema_artifact_evidence_defined`：DDL review evidence、rollback fixture evidence、schema version evidence、tenant index evidence、read-only role evidence 和七条 read route 到未来 schema artifact / projection 的映射关系已经可检查。它不创建 migration 目录、manifest、SQL、DDL review artifact、rollback fixture、schema smoke artifact 或 `ControlPlaneReadSchemaMigrationRunner`，不实现 migration runner、repository adapter、store selector、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-store-selector-smoke-readiness-v1` 当前只固定 store selector smoke readiness，状态为 `store_selector_smoke_readiness_defined`：未来 selector smoke contract、模式矩阵、reserved / unknown mode fail-closed 和七条 route selector smoke matrix 必须先于 selector / config artifact 创建。它不创建 `control_plane_read_store_selector.go`、selector test、selector smoke fixture / checker，不新增正式 `RADISHMIND_CONTROL_PLANE_READ_STORE` 配置入口，不实现 store selector、repository adapter、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-production-auth-readiness-v1` 当前只固定 production auth readiness，状态为 `production_auth_readiness_defined`：未来 Radish OIDC issuer discovery evidence、token validation contract preconditions、claim mapping、tenant binding、scope projection、failure mapping、no fake fallback 和 no side effects 必须先于 auth middleware、token validation 或 production API consumer 创建。它不创建 `contracts/radish-oidc-token-validation.schema.json`、`control_plane_read_auth_middleware.go`、production auth smoke fixture / checker，不实现 Radish OIDC client、issuer network call、token validation、auth middleware、login / logout flow、session cookie、repository adapter、SQL、migration、真实数据库、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-adapter-smoke-readiness-v1` 当前只固定 adapter smoke readiness，状态为 `adapter_smoke_readiness_defined`：未来 durable adapter smoke 必须消费 schema artifact manifest readiness、store selector smoke readiness、production auth readiness、static runner 和 repository adapter implementation plan，并覆盖七条 read route 的 operation、request/result type、scope、failure mapping、no fake fallback 和 no side effects。它不创建 adapter smoke fixture / checker、`control_plane_read_repository_interface.go`、`control_plane_read_repository_adapter.go`、adapter test、selector、SQL、migration、真实数据库、auth middleware、token validation 或 production API consumer，不实现 API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-implementation-trigger-review-v1` 当前只固定 implementation trigger review，状态为 `implementation_trigger_review_defined`：schema artifact、store selector、production auth 和 adapter smoke 四类候选均为 `not_satisfied`，当前没有任何 read-side implementation trigger satisfied。它不创建 migration manifest、SQL、selector、auth middleware、adapter smoke fixture / checker、repository interface、repository adapter、真实数据库、token validation 或 production API consumer，不实现 API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`control-plane-read-implementation-entry-review-v1` 当前只固定 implementation entry review，状态为 `implementation_entry_review_defined`：读取 trigger review、schema artifact evidence 和 product surface recheck 后，结论仍是当前不打开数据库 / OIDC / adapter 实现入口。它不新增同层只读 UI、不启动开发服务器、不创建 migration manifest、SQL、selector、auth middleware、adapter smoke fixture / checker、repository adapter、真实数据库、token validation 或 production API consumer。

`control-plane-durable-read-foundation-v1` 当前固定为 `durable_read_foundation_implemented`：`ControlPlaneReadRepository` interface 边界已在 `services/platform/internal/httpapi/control_plane_read_repository.go` 落地，现有七条 fake-store-backed read handlers 已通过 repository interface 消费数据。当前 repository 只包裹 fixture-backed fake store，保持 response envelope、dev-only fake auth、product sample fixture 和 no-side-effects 行为稳定；它不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

`workflow-saved-draft-v1-implementation` 已同时支持默认 `memory_dev` 和显式 `postgres_dev_test`。后者落地真实 PostgreSQL migration、schema marker、连接池、repository query executor、重启恢复和原子 expected-version；`repository_disabled` / production `repository` 仍返回 `repository_store_disabled`。这项开发 / 测试能力与 Production Secret Backend / Audit Store 链解耦；其上只新增了显式 opt-in 的 Workflow Executor v0，OIDC、production secret、publish、tool、confirmation、writeback、replay 和 production executor 停止线不变。
## Saved Workflow Draft 运行层说明

维护 saved draft 平台代码时，按以下边界理解当前已落地的 durable store 相关实现：

- `workflow_saved_draft.go` 是 domain service；只有 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP=1` 时 route 才可用，保存还必须显式启用 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE=1`。
- `workflow_saved_draft_repository.go` 定义 repository contract；`workflow_saved_draft_postgres.go` 以 PostgreSQL 原子 CAS 实现 save / read / list，`workflow_saved_draft_repository_store.go` 负责 domain bridge。
- `workflow_saved_draft_repository_adapter.go` 是 adapter contract enforcement：先检查 actor context、operation scope、payload scope 和 schema preflight，再调用 injected query executor。query executor 缺失返回 `draft_store_unavailable`，schema migration 未应用返回 `draft_schema_migration_not_applied`，store schema 不一致返回 `draft_store_schema_version_mismatch`，migration 状态不可观测返回 `draft_store_migration_unavailable`。
- `migrations/workflow_saved_drafts/` 与 `cmd/radishmind-workflow-draft-migrate` 提供 reviewed up/down SQL、checksum、advisory lock 和显式 `status` / `up`；服务启动只 preflight，不自动 migration。
- `postgres_dev_test` 使用独立 migration / runtime 数据库身份，runtime 只有表级 DML。任何数据库、marker 或连接错误都失败关闭，不能回退 `memory_dev`。
- `workflow_saved_draft_production_auth_runtime.go` 仍只接受上游 verified context；production OIDC middleware、membership adapter 和 production `repository` 没有启用。

本地 PostgreSQL dev/test 依次运行 `./scripts/run-workflow-saved-draft-postgres-dev-test.sh check`、`./scripts/run-radishmind-web-dev.sh --mode dev-live --saved-draft-postgres-dev-test` 与 `./scripts/run-workflow-saved-draft-postgres-dev-test.sh down`；PowerShell 使用同名 `.ps1` 入口与 `-SavedDraftPostgresDevTest`。`check` 留下已迁移 schema 供联调，`down` 停止容器但保留命名卷。

新增 failure code 均应保持 fail closed：`draft_auth_context_contract_mismatch`、`draft_identity_context_missing`、`draft_tenant_binding_missing`、`draft_workspace_membership_denied`、`draft_application_scope_denied`、`draft_owner_scope_denied`、`draft_scope_grant_missing`、`draft_audit_context_missing`、`draft_store_contract_mismatch`、`draft_schema_migration_not_applied`、`draft_store_schema_version_mismatch` 和 `draft_store_migration_unavailable` 不得返回草案主体，也不得创建 executor、confirmation、writeback 或 replay side effect。

## Application Configuration Draft 运行层说明

- `application_configuration_draft.go` 定义独立 application draft domain、sanitized payload、validation、scope / owner 和 expected-version CAS；不复用 Workflow draft graph 或 repository。
- `application_configuration_draft_http.go` 提供 dev-only validate / save / read / list routes。只有 `RADISHMIND_APPLICATION_DRAFT_DEV_HTTP=1` 可访问，保存还要求 `RADISHMIND_APPLICATION_DRAFT_DEV_WRITE=1`、dev auth、`application_drafts:write` 和匹配 workspace / application headers。
- `application_configuration_draft_postgres.go` 与 `migrations/application_configuration_drafts/` 提供 PostgreSQL dev/test repository、0001 schema、marker、checksum、advisory lock 和显式 `status` / `up` runner；平台启动只 preflight，不自动 migration。
- `memory_dev` 与 `postgres_dev_test` 使用相同 validation、scope 和 CAS 语义；PostgreSQL store 连接、marker 或 query 失败返回 `application_draft_store_unavailable`，不得回退 memory。
- 草案不保存 API key、Authorization、internal caller header、provider credential / endpoint 或 Gateway 测试输入输出，也不创建 / 发布 / 删除正式 application。

## Application Publish Governance 运行层说明

- `application_publish_candidate.go` 定义 immutable candidate、append-only review、server-side draft reload / digest、baseline / draft drift、superseded 与 promotion eligibility；candidate body 不接受配置 snapshot。
- `application_publish_candidate_http.go` 提供 dev-only create / list / read / review routes，分别要求 `application_publish_candidates:write`、`:read` 或 `:review`，以及匹配 workspace / application header。
- `application_publish_candidate_postgres.go` 与 `migrations/application_publish_candidates/` 提供独立 PostgreSQL dev/test repository、0001 schema、marker、checksum、advisory lock和显式 `status` / `up` runner；平台启动只 preflight，不自动 migration。
- `memory_dev` 与 `postgres_dev_test` 共享 immutable create、scope、review CAS、排序和 sanitized projection 语义；数据库失败返回稳定 store failure，不回退 memory。
- eligibility 当前总是 `promotion_blocked`；正式 application repository、production auth / membership、发布 owner 与 promotion runtime 未建立，不存在 promotion route，也不修改 application read model。

## Control Plane Read-Side readiness 运行层说明

平台服务层当前支持 `fake_store_dev`、受限的 `signed_test_token + postgres_dev_test`，以及 `radish_oidc_integration_test + postgres_dev_test` Control Plane read 组合。数据库模式只把 Admin tenant summary 与 audit summary 路由到 `control_plane_admin_read` PostgreSQL projection；signed test 模式下其余五条 workspace route 保持显式 fake binding，OIDC integration 模式下则在 handler 授权边界统一返回 `workspace_membership_unavailable`，不会触达 fake repository。startup 校验 OIDC policy / discovery / JWKS、migration marker、checksum 与 runtime SELECT 权限，任何失败均不回退。真实 Radish 联调为 `real_radish_integration_deferred`；未来由 Radish 注册 RadishMind application/client 与 resource audience 后恢复。该路径不启用 production repository、运行时 writer 或 production API consumer。

Control Plane Read 在 formal UI / dev-live consumer 之后的旧 repository readiness 尾链已退出活动仓库基线，历史文件继续保留。当前 read route contract、negative contract、正式 UI 聚合检查和 Go 测试仍是活动门禁；这次放宽只服务 Saved Draft `postgres_dev_test`，不代表 Control Plane Read database store 已实现。

`control-plane-read-consumer-contract-v1` 当前固定 `contracts/typescript/control-plane-read-api.ts`、`scripts/run-control-plane-read-consumer-smoke.py` 和七条 read route 的只读 consumer view model。它让上层可以按统一 envelope、failure view、cursor、audit ref 和 forbidden output policy 消费 response fixture，但不实现正式 UI、不请求真实后端、不接数据库、OIDC、executor、confirmation、writeback 或 replay。

`control-plane-read-formal-ui-boundary-v1` 当前只固定正式 UI 边界：`Admin Control Plane` 与 `User Workspace` 的页面划分、每个页面消费的 read route、loading / empty / denied / stale / partial failure / forbidden projection 状态和敏感字段停止线。它不修改 `apps/radishmind-console/`，不实现 React UI、不请求真实后端、不接数据库、OIDC、executor、confirmation、writeback 或 replay。

`control-plane-read-formal-ui-implementation-readiness-v1` 当前只固定正式 UI 实现 readiness：未来正式产品 UI 预留落点为 `apps/radishmind-web/`，`apps/radishmind-console/` 继续只是本地 ops surface，页面实现顺序、consumer contract 复用和测试策略均由 fixture/checker 固定。该 readiness 不创建 React 页面、不创建 `apps/radishmind-web/`、不修改当前 ops console、不新增 platform route、不接数据库、OIDC、executor、confirmation、writeback 或 replay。

`control-plane-read-shared-shell-v1` 当前创建 `apps/radishmind-web/` 首个 read-only shared shell。它只复用 `contracts/typescript/control-plane-read-api.ts` 渲染离线 route catalog、共享状态和 forbidden output guard，不请求 platform live route，不新增 platform route，不接数据库、OIDC、executor、confirmation、writeback 或 replay；`apps/radishmind-console/` 仍保持本地 ops surface。

`control-plane-read-admin-tenant-overview-v1` 当前在 `apps/radishmind-web/` 的 shared shell 内新增只读 `admin-tenant-overview` 页面切片。它只消费 `tenant-summary-route` 的离线 TypeScript view model，展示租户摘要、route metadata、request / audit ref、页面状态和 forbidden output guard；不请求 platform live route，不新增 platform route，不接数据库、OIDC、API key / quota、executor、confirmation、writeback 或 replay，也不声明 production admin console ready。

`control-plane-read-workspace-applications-v1` 当前在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-applications` 页面切片。它只消费 `application-summary-list-route` 的离线 TypeScript view model，展示应用摘要列表、cursor、route metadata、request / audit ref、页面状态和 forbidden output guard；不请求 platform live route，不新增 platform route，不接数据库、OIDC、API key / quota、executor、confirmation、writeback 或 replay，也不声明 formal user workspace complete。

`control-plane-read-workspace-api-keys-v1` 当前在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-api-keys` 页面切片。它只消费 `api-key-summary-list-route` 的离线 TypeScript view model，展示 API key id、owner、scope、state、时间字段、route metadata、request / audit ref、页面状态和 forbidden output guard；不请求 platform live route，不新增 platform route，不接数据库、OIDC、API key lifecycle、quota enforcement、executor、confirmation、writeback 或 replay，不展示 key value 或 hash，也不声明 formal user workspace complete。

`control-plane-read-workspace-usage-quota-v1` 当前在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-usage-quota` 页面切片。它只消费 `quota-summary-route` 的离线 TypeScript view model，展示 quota id、period、request / token / cost limit、usage snapshot、over quota failure code、route metadata、request / audit ref、页面状态和 forbidden output guard；不请求 platform live route，不新增 platform route，不接数据库、OIDC、quota enforcement、rate limit、billing、cost ledger、executor、confirmation、writeback 或 replay，也不声明 formal user workspace complete。

`control-plane-read-workspace-workflow-definitions-v1` 当前在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-workflow-definitions` 页面切片。它只消费 `workflow-definition-summary-list-route` 的离线 TypeScript view model，展示 workflow definition id、application ref、version、definition status、node count、risk level、confirmation capability、updated timestamp、route metadata、request / audit ref、页面状态和 forbidden output guard；不请求 platform live route，不新增 platform route，不接数据库、OIDC、workflow builder、workflow lifecycle mutation、workflow executor、tool executor、confirmation、writeback 或 replay，也不声明 formal user workspace complete。

`control-plane-read-workspace-run-history-v1` 当前在 `apps/radishmind-web/` 的 shared shell 内新增只读 `workspace-run-history` 页面切片。它只消费 `run-record-summary-list-route` 的离线 TypeScript view model，展示 run id、workflow definition ref、application ref、status、failure code、cost summary、trace id、started / completed timestamp、route metadata、request / audit ref、cursor、页面状态和 forbidden output guard；不请求 platform live route，不新增 platform route，不接数据库、OIDC、workflow executor、tool executor、run replay、run resume、materialized result reader、confirmation、writeback 或 replay，也不声明 formal user workspace complete。

read-side UI 当前已经完成七个页面、formal UI readiness close、dev-only live read consumer 和 auth/store transition preconditions。后续普通只读展示页应继续通过聚合 surface matrix / checker 收口，不再默认逐页新增专项门禁。任何后续 read-side 切片都不能把当前 fake-store-backed handler、测试身份上下文或 auth/store transition gate 写成真实数据库、Radish OIDC、repository、API key lifecycle、quota enforcement、billing、workflow executor、confirmation、writeback 或 replay 能力。

当前 bridge northbound 仍以文本消息和单轮问答为主：只把最后一条文本用户消息映射到 `radish/answer_docs_question`，返回内容优先取 canonical `summary`、必要时回退首条 `answer`，并支持增量流式转发。

`GET /v1/models` 目前通过 Python provider registry 输出带 route metadata 的 model inventory，作为 northbound discoverability 的第一版收口；它当前已支持列表和 `/v1/models/{id}` 精确 lookup，并带出 provider-qualified profile inventory。profile 可选择 ID 固定为 `profile:<profile>` 或 `provider:<provider>:profile:<profile>`，并与请求侧 selection 和 `diagnostics.providers.selectable_model_ids` 共用同一套 metadata。

## Provider runtime / health boundary

`Provider Runtime & Health v1` 当前已经固定五个可复验边界：

- `provider-capability-matrix-v1`：`scripts/checks/fixtures/provider-capability-matrix-v1.json` 与 `scripts/check-provider-capability-matrix.py` 从 `services/runtime/provider_registry.py` 校验 provider capability matrix。它只说明 provider 声明、profile model id、northbound route / protocol 和 capability metadata 可检查，不说明 provider health 或 production readiness。
- `provider-health-smoke-v1`：`scripts/checks/fixtures/provider-health-smoke-v1.json` 与 `scripts/check-provider-health-smoke.py` 默认只跑 mock runtime smoke 和 config-level inventory smoke。它不联网、不要求真实 credential、不下载模型；optional live health 仍是手动未来切片，失败只能作为 provider health signal，不能写成 production outage。
- `provider-selection-policy-v1`：`scripts/checks/fixtures/provider-selection-policy-v1.json` 与 `scripts/check-provider-selection-policy.py` 固定 request-side profile / provider / concrete model selection、`/v1/models/{id}` 负向边界、credential missing、unsupported capability、timeout 分类和 no implicit fallback 口径。
- `provider-retry-fallback-policy-v1`：`scripts/checks/fixtures/provider-retry-fallback-policy-v1.json` 与 `scripts/check-provider-retry-fallback-policy.py` 固定 retry/fallback audit metadata。当前 `retry_policy=caller-managed`、`fallback_policy=disabled`，平台失败路径不自动重试、不隐式切换 provider。
- `provider-runtime-docs-refresh`：`scripts/checks/fixtures/provider-runtime-docs-refresh.json` 与 `scripts/check-provider-runtime-docs-refresh.py` 固定说明文档、任务卡和入口文档的阶段口径，避免把 capability、health smoke 或 selection policy 误写成 live health、retry/fallback 或 production ready。

调用方应把 `/v1/models` 中的 `credential_state`、`deployment_mode`、`auth_mode`、`streaming`、`northbound_routes` 和 `northbound_protocols` 视为只读发现信息。请求选中某个 profile 后，平台会在 canonical request 的 `context.northbound` 中带出同源 selection metadata，便于审计和排障。

选择失败或未知输入的边界保持显式：

- 未知 `/v1/models/{id}` 返回 `MODEL_NOT_FOUND`。
- 未知 concrete model 可以作为 `runtime_override` 进入 canonical request，但不会被解释为 inventory match。
- 显式未知 `radishmind.provider_profile` 不会被 active profile 隐式替换。
- 当前不做隐式 provider fallback，也不声明 retry/fallback execution 已实现。

这一层仍不是 production provider health system。正式 secret backend、外部 provider live health、live timeout probe、真实 retry/fallback、测试 / 生产环境 smoke 和 production readiness 仍需要独立任务、运行窗口和证据记录。

当前平台级 `ops smoke` 已由 `scripts/check-platform-ops-smoke.py` 固定为快速门禁。它不启动长期驻留服务、不访问外部 provider，只验证三类可运行边界：

- `go test ./...` 能覆盖平台服务层的 `healthz`、northbound 路由、provider/profile selection、session recovery checkpoint metadata-only read route 和 SSE 兼容行为。
- `go test ./...` 也覆盖最小 session/tooling metadata shell 和 blocked action response，确保它们不暴露 executor、materialized result、durable memory 或业务写回能力。
- `scripts/run-platform-bridge.py providers` 能从 Python registry 输出 `mock`、`openai-compatible`、`huggingface` 与 `ollama` provider 能力。
- `scripts/run-platform-bridge.py inventory` 能在受控环境变量下暴露 openai-compatible fallback chain、HuggingFace profile 和 Ollama local profile，并且只暴露 `has_api_key` / `credential_state`，不泄漏 key 原文。

配置分层门禁由 `scripts/check-platform-config.py` 固定到快速检查中。它通过同一个 `config.LoadFromEnv` 入口验证 `config-summary` 和 `config-check`，只输出脱敏字段：provider、profile、model、base_url 是否存在、`credential_state`、timeout、listen addr、Python bridge mode / worker / queue / handshake、路径与字段来源，不输出 `RADISHMIND_PLATFORM_API_KEY` 或 `base_url` 原文。

部署壳 smoke 由 `scripts/check-platform-deployment-smoke.py` 固定到快速检查中。它不启动长驻服务、不访问外部 provider，只验证本地配置文件加载、环境变量覆盖、无效配置失败和 secret 不泄漏。

结构化诊断 smoke 由 `scripts/check-platform-diagnostics.py` 固定到快速检查中。它通过一次性 `diagnostics` 命令聚合启动配置、必填字段、Python bridge provider registry 和 provider/profile inventory，不启动长驻服务、不访问外部 provider，并只输出 `credential_state`、计数和状态字段，不输出 secret、token 或 provider URL 原文。

## Production config / secret boundary

`Production Ops Hardening v1` 的第一切片是 `config-secret-boundary`。当前由 `scripts/checks/fixtures/production-ops-config-secret-boundary.json` 与 `scripts/check-production-ops-config-secret-boundary.py` 固定为 governance boundary：配置来源、密钥注入、provider/profile 覆盖和不可提交项已经有可检查口径，但 production secret backend 仍未实现，不能据此声明 production ready。

配置来源继续按 `default < config file < env` 分层。`mock` provider、`local-smoke`、本地 wrapper 默认值和 `tmp/radishmind-platform.local.json` 只代表 local / dev readiness；远端 provider 的 `RADISHMIND_PLATFORM_API_KEY`、provider profile API key、token、cookie、authorization header 和真实 provider base URL 不得写入 committed 文档、fixture、日志样例或 console 包。

生产环境仍缺少正式 secret backend、环境隔离、provider health policy、process supervisor 和 console production packaging。直到这些条件分别有任务卡、实现和门禁前，`config-summary`、`config-check`、`diagnostics` 和 provider inventory 只能输出 `credential_state`、`base_url_configured`、`secret_fields`、`field_sources` 这类脱敏字段，不输出原始 credential 或 provider URL。

## Production secret backend contract

`production-secret-backend-contract` 已由 `scripts/checks/fixtures/production-ops-secret-backend-contract.json` 与 `scripts/check-production-ops-secret-backend-contract.py` 固定为 Production Ops Hardening v1 的最小治理切片。该切片只定义未来 external secret backend adapter contract：按 `environment`、`provider`、`provider_profile` 与 `secret_ref` 识别 secret reference，并要求后续运行面只暴露 `credential_state`、`secret_backend_configured`、`secret_ref_present`、`missing_secret_refs` 和 `field_sources` 等脱敏字段。

当前明确不实现真实云 secret 服务、不写入真实 secret、不调用云 API、不声明 production ready。`RADISHMIND_PLATFORM_API_KEY` 仍只允许作为 developer env override；`RADISHMIND_SECRET_SOURCE` 只能表示部署态外部 secret 来源要求，不是 secret backend 本身。真实 production secret backend、secret rotation runtime、production secret audit store、provider health policy、environment isolation 和 process supervisor 仍为 `not_satisfied`。

## Production secret backend implementation readiness

`production-secret-backend-implementation-readiness` 已由 `docs/task-cards/production-secret-backend-implementation-v1-plan.md`、`scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json` 与 `scripts/check-production-ops-secret-backend-implementation-readiness.py` 固定为下一步实现前置条件清单。当前 secret ref schema、config 注入点、provider profile binding、脱敏审计字段、failure taxonomy、operator runbook / negative gates、rotation / audit policy、test fixture strategy / fake resolver entry review、fake resolver contract / no leakage strategy、fake resolver task card、test-only fake resolver runtime、真实 resolver preconditions / entry review / entry refresh、backend profile selection、real resolver no leakage strategy / runtime entry review、credential handle boundary / runtime entry review、operator approval evidence / runtime entry review、audit store handoff / contract / ownership / delivery-idempotency / runtime entry refresh v3、backend health boundary 和 backend health runtime entry review 已有可检查证据；test-only fake resolver runtime 默认 disabled，production resolver runtime 仍为 `not_created`，不接云、不读 secret、不接数据库。

该 readiness 不实现 production resolver runtime、backend health runtime、no secret leakage smoke runtime、云 secret service、secret 解析、credential payload、credential handle runtime、approval runtime、audit store runtime、数据库连接或 workflow saved draft repository mode，也不改变生产运行时默认状态。audit store、approval、credential handle、no leakage smoke、backend health 和真实 resolver 均已完成 runtime task card 前的静态准入复核，但结论仍是 blocked before runtime task card；后续只有在单独实现切片中补齐仍 blocked 的 runtime dependency 后，才能继续讨论真实 production secret backend。

## Production secret reference schema

`secret-ref-schema-and-fixtures` 已由 `contracts/production-secret-reference.schema.json`、`scripts/checks/fixtures/production-secret-reference-basic.json` 与 `scripts/check-production-secret-reference-contract.py` 固定为 committed secret reference contract。该 schema 只允许保存 `environment`、`provider`、`provider_profile`、`secret_ref`、`required_fields` 和 `sanitized_fields`，并要求 fixture 明确 `stores_secret_values=false`、`resolver_enabled=false`、`cloud_calls_allowed=false` 和 `production_secret_backend_ready=false`。

该 schema / fixture 只证明 secret reference 格式可检查，不实现 resolver，不接云，不包含 secret value、provider raw URL、API key、token、cookie、authorization header 或 credential 原文。当前已继续补齐 `config-secret-ref-readiness` 与 `provider-profile-secret-binding` 两个前置，让 config summary / diagnostics 和 provider/profile binding 都只报告脱敏状态。

## Production secret backend config / secret ref readiness

`config-secret-ref-readiness` 已由 `docs/platform/production-secret-backend-config-secret-ref-readiness-v1.md`、`docs/task-cards/production-secret-backend-config-secret-ref-readiness-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py` 固定为 `config_secret_ref_readiness_defined`。它只定义 future config / diagnostics 层如何报告 `secret_backend_configured`、`secret_ref_present`、`missing_secret_refs` 和 `field_sources` 等脱敏状态，并把 `secret_reference_manifest_missing`、`secret_reference_manifest_invalid`、`secret_ref_missing`、`secret_backend_disabled` 和 `resolver_invocation_forbidden` 映射到 `configuration` failure boundary。

该 readiness 不修改 `config.LoadFromEnv` runtime，不实现 resolver、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不接 database connection provider，也不启用 workflow saved draft repository mode。`production_secret_backend` 仍为 `not_satisfied`，`resolver_implementation_status` 仍为 `not_started`。

## Production secret backend provider profile secret binding readiness

`provider-profile-secret-binding` 已由 `docs/platform/production-secret-backend-provider-profile-secret-binding-readiness-v1.md`、`docs/task-cards/production-secret-backend-provider-profile-secret-binding-readiness-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py` 固定为 `provider_profile_secret_binding_readiness_defined`。它只定义 future provider/profile inventory 如何声明 `credential_requirement`、`secret_ref_status`、`secret_ref_present`、`missing_secret_refs`、`field_sources` 和环境绑定，并把 `provider_profile_binding_missing`、`provider_profile_credential_required`、`provider_profile_secret_ref_missing`、`provider_profile_environment_mismatch`、`provider_profile_secret_backend_disabled` 和 `provider_profile_resolver_forbidden` 映射到 `configuration` failure boundary。

该 readiness 不修改 provider runtime，不实现 resolver、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不访问 provider、不接 database connection provider，也不启用 workflow saved draft repository mode。`secret_ref_status=present` 不等于 credential resolved，`production_secret_backend` 仍为 `not_satisfied`，`resolver_implementation_status` 仍为 `not_started`。

## Production secret backend secret resolver interface disabled readiness

`secret-resolver-interface-disabled` 已由 `docs/platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md`、`docs/task-cards/production-secret-backend-secret-resolver-interface-disabled-readiness-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py` 固定为 `secret_resolver_interface_disabled_readiness_defined`。它只定义 future resolver interface 的 reference-only input、disabled result、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。

该 readiness 不修改 platform runtime，不实现 resolver runtime、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不访问 provider、不创建 credential handle、不接 database connection provider，也不启用 workflow saved draft repository mode。`resolver_state=disabled` 不等于 credential resolved，`production_secret_backend` 仍为 `not_satisfied`，`resolver_implementation_status` 仍为 `not_started`。operator runbook / negative gates 已由下一节固定为独立 readiness。

## Production secret backend operator runbook / negative gates readiness

`operator-runbook-and-negative-gates` 已由 `docs/platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md`、`docs/task-cards/production-secret-backend-operator-runbook-negative-gates-readiness-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py` 固定为 `operator_runbook_negative_gates_readiness_defined`。它只定义 operator runbook、test / production secret source、operator approval evidence、sanitized verification、smoke record reference、negative gates、failure mapping、no fallback、no side effects 和 artifact guard。

该 readiness 不修改 platform runtime，不实现 operator runbook executor、不实现 negative gate runtime、不实现 resolver runtime、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不访问 provider、不创建 credential handle、不接 database connection provider，也不启用 workflow saved draft repository mode。它只满足 `operator-runbook-and-negative-gates` 前置；rotation / audit policy 已由下一节固定，test fixture strategy / fake resolver implementation 和真实 production secret backend 仍为后续独立目标。

## Production secret backend rotation / audit policy readiness

`rotation-and-audit-policy` 已由 `docs/platform/production-secret-backend-rotation-audit-policy-readiness-v1.md`、`docs/task-cards/production-secret-backend-rotation-audit-policy-readiness-v1-plan.md`、`scripts/checks/fixtures/production-secret-backend-rotation-audit-policy-readiness-v1.json` 与 `scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py` 固定为 `rotation_audit_policy_readiness_defined`。它只定义 rotation trigger、approval / change window、secret ref version reference、rollback / disable policy、sanitized verification、audit event fields、failure mapping、no fallback、no side effects 和 artifact guard。该 readiness 不修改 platform runtime，不实现 rotation runtime、不写 production secret audit store、不创建 audit writer、不实现 resolver runtime、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不访问 provider、不创建 credential handle、不接 database connection provider，也不启用 workflow saved draft repository mode。它只满足 `rotation-and-audit-policy` 前置；test fixture strategy / fake resolver implementation、真实 resolver implementation、测试环境 smoke 和生产前复核仍为后续独立目标。

`test-fixture-strategy-fake-resolver-entry-review`、`fake-resolver-contract-no-secret-leakage-smoke-strategy`、`fake-resolver-implementation-task-card-entry-readiness-review`、`fake-resolver-implementation`、`fake-resolver-runtime-implementation-entry-review`、`fake-resolver-runtime-implementation`、real resolver preconditions / entry review / entry refresh、backend profile selection、real resolver no leakage strategy / runtime entry review、credential handle boundary / runtime entry review、operator approval evidence / runtime entry review、audit store handoff / contract / ownership / delivery-idempotency / runtime entry refresh v3、backend health boundary 和 backend health runtime entry review 已由对应 platform topic、task card、fixture、checker 与 Go 单测固定为证据链；它们说明 fake resolver runtime 只能在 test-only、disabled-by-default、offline no leakage smoke 边界下工作，并且只能输出 opaque test credential handle metadata 与脱敏诊断。这些证据不把 test-only fake resolver runtime 写成 production resolver、不创建 backend health runtime、不创建 no secret leakage smoke runtime、不读取环境 secret、不调用云 secret 服务、不访问 provider、不创建 credential payload、不连接数据库、不打开 driver / connection factory、不执行 SQL、不读取或写入 schema marker，也不启用 workflow saved draft repository mode 或 production API。`fake_resolver_runtime_test_only_implemented` 不表示 production secret backend ready。

## Startup / supervisor boundary

`startup-supervisor-boundary` 已由 `scripts/checks/fixtures/production-ops-startup-supervisor-boundary.json` 与 `scripts/check-production-ops-startup-supervisor-boundary.py` 固定为 governance boundary。当前支持的启动入口只有两类：`scripts/run-platform-service.ps1` / `scripts/run-platform-service.sh` 的人工 platform wrapper，以及 `scripts/run-radishmind-console-dev.ps1` / `scripts/run-radishmind-console-dev.sh` 的本地 console dev launcher。

platform wrapper 只支持 `serve`、`config-summary`、`config-check` 和 `diagnostics`，未知命令必须失败并返回退出码 `2`。console dev launcher 只负责启动或复用本地 backend / frontend，探测 `/healthz`、`/v1/platform/overview`、`/v1/platform/local-smoke`、本地 CORS preflight 和前端页面；`-ExitAfterProbe` / `--exit-after-probe` 只表示探测后清理本次创建的开发进程，不是 lifecycle management。

当前仍没有 production process supervisor、automatic restart、production service manager 或 production log retention。`tmp/radishmind-console-dev` 下的日志只用于本地排障，不能解释为生产日志路径；port reuse 只表示开发期复用本机已有进程，不能解释为服务发现。`local-smoke` 仍不是 production health。

## Environment isolation boundary

`environment-isolation` 已由 `scripts/checks/fixtures/production-ops-environment-isolation-boundary.json` 与 `scripts/check-production-ops-environment-isolation-boundary.py` 固定为 governance boundary。当前只区分三类 readiness：`local_readiness`、`dev_smoke` 和仍为 `not_satisfied` 的 `production_readiness`。

`/v1/platform/local-smoke`、`mock` provider、本地 console CORS、developer env override 和 Vite preview 都只属于 local / dev 范围。`local-smoke` 只能说明本地只读 console 所需 route、contract、CORS 和 stop-lines 可读；它不是 production health。`mock` provider 和 demo profile 只能证明协议、UI 和本地排障链路可用，不能解释为 production provider / profile。`http://127.0.0.1:4000` 与 `http://localhost:4000` 的 CORS 允许列表只服务本地 console 开发，不是 production CORS policy。

production readiness 仍必须等待 deployment environment isolation、production auth policy、production CORS policy、provider health policy、secret backend、process supervisor 和 console production packaging。当前任何 local / dev smoke 通过都不得升级为 production ready。

## Console package smoke boundary

`console-production-package-smoke` 已由 `scripts/checks/fixtures/production-ops-console-package-smoke.json` 与 `scripts/check-production-ops-console-package-smoke.py` 固定为 governance boundary。当前 console package 只允许 `dev`、`build` 和 `preview` 三个 npm script：`npm run build` 用于本地或 CI smoke，执行 `tsc --noEmit && vite build`；`npm run preview` 只允许绑定 `127.0.0.1:4000` 做本地 build preview。

这些入口不等于 production package 或 production hosting。`apps/radishmind-console/package.json` 必须保持 `private=true`，不得添加 deploy / publish / release / package / docker 类脚本，不得提交 `apps/radishmind-console/dist/` 或 `apps/radishmind-console/node_modules/`。console production packaging 仍为 `not_satisfied`，直到另有生产发布目标、静态资产策略、正式鉴权 / CORS policy 和部署门禁。

## Docker local compose boundary

`docker-local-compose` 已由 `scripts/checks/fixtures/production-ops-docker-local-compose.json` 与 `scripts/check-production-ops-docker-local-compose.py` 固定为本地容器 smoke 资产。当前只新增 `services/platform/Dockerfile`、`apps/radishmind-console/Dockerfile`、`apps/radishmind-console/nginx.local.conf` 和 `deploy/docker-compose.local.yaml`，用于本机 build 并验证 platform + console 的容器组合。

本地 compose 默认使用 `mock` provider，发布 `7000/4000` 到宿主机，console 构建期 `VITE_RADISHMIND_PLATFORM_BASE_URL` 默认指向 `http://127.0.0.1:7000`。它不包含 secret、不使用 `RADISHMIND_IMAGE_TRACK` / `RADISHMIND_IMAGE_TAG`，不定义测试 / 生产部署态，不实现 production secret backend、process supervisor、正式 auth / CORS policy、镜像发布或 production ready。

本地长驻服务入口由 `scripts/run-platform-service.sh` 与 `scripts/run-platform-service.ps1` 收口。wrapper 会固定仓库根、`services/platform` 工作目录、默认 `GOCACHE=/tmp/radishmind-go-build-cache`，在未显式设置 `RADISHMIND_PLATFORM_PYTHON_BIN` 时使用仓库 `.venv` Python；缺少 `.venv` 时要求先执行 bootstrap，不再隐式回退系统 Python。`tmp/radishmind-platform.local.json` 存在时会自动作为默认 `RADISHMIND_PLATFORM_CONFIG`。

`/v1/models` 的 profile metadata 现在必须带出稳定 discoverability 字段：`capabilities`、`northbound_protocols`、`northbound_routes`、`credential_state`、`deployment_mode`、`auth_mode` 与 `streaming`。调用方应基于这些字段判断某个 profile 能否用于 chat、是否支持流式、凭据是否已配置，以及它属于 remote API 还是 local daemon。

profile 可选择 ID 固定为 `profile:<profile>` 或 `provider:<provider>:profile:<profile>`。`/v1/models`、请求时的 provider/profile selection 和 `diagnostics.providers.selectable_model_ids` 使用同一套 ID 与 readiness metadata；当请求选中某个 profile 时，canonical request 的 `context.northbound` 会记录 `credential_state`、`deployment_mode`、`auth_mode`、`streaming`、`northbound_routes` 与 `northbound_protocols`，用于审计和排障。

请求级观测已固定为轻量平台能力：`/v1/chat/completions`、`/v1/responses` 与 `/v1/messages` 会接受或生成 `request_id`，通过 `X-Request-Id` 响应头返回，并把同一 ID 写入 canonical `CopilotRequest.request_id` 与 `context.northbound.request_id`。非流式和流式路径都会记录统一日志字段：`request_id`、route、HTTP status、latency、provider/profile、selected model、selection source，以及失败时的 error code 与 failure boundary。

错误响应采用统一 error taxonomy，不输出 provider URL 或 credential 原文。当前固定的主要失败边界包括：`northbound_request`、`canonical_request`、`provider_inventory`、`python_bridge`、`platform_response`、`southbound_provider` 与 `configuration`。

## 本地启动 runbook

该 runbook 面向开发者本机验证，不等同于 production deployment。启动服务前应先运行平台层单元测试：

```bash
cd services/platform
GOCACHE=/tmp/radishmind-go-build-cache go test ./...
```

从仓库根目录启动本地平台服务：

```bash
RADISHMIND_PLATFORM_LISTEN_ADDR=127.0.0.1:7000 \
RADISHMIND_PLATFORM_PROVIDER=mock \
RADISHMIND_PLATFORM_MODEL=radishmind-local-dev \
go run ./services/platform/cmd/radishmind-platform
```

如果已经进入 `services/platform` 目录，也可以使用：

```bash
RADISHMIND_PLATFORM_LISTEN_ADDR=127.0.0.1:7000 \
RADISHMIND_PLATFORM_PROVIDER=mock \
RADISHMIND_PLATFORM_MODEL=radishmind-local-dev \
go run ./cmd/radishmind-platform
```

启动路径仍由 `config.LoadFromEnv -> httpapi.NewServer -> ListenAndServe` 组成。默认 `stdio_pool` 在监听端口前完成四个 Python worker 的 protocol v1 handshake，`Go` 层只负责本地 HTTP 壳、northbound 请求翻译和受控 bridge 调度，不直接承载模型推理逻辑。

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `RADISHMIND_PLATFORM_CONFIG` | 空 | 可选 JSON 配置文件路径 |
| `RADISHMIND_PLATFORM_LISTEN_ADDR` | `:7000` | 本地 HTTP 监听地址 |
| `RADISHMIND_PLATFORM_READ_HEADER_TIMEOUT` | `5s` | HTTP header 读取超时 |
| `RADISHMIND_PLATFORM_WRITE_TIMEOUT` | `30s` | HTTP 写响应超时 |
| `RADISHMIND_PLATFORM_BRIDGE_TIMEOUT` | `30s` | Go 调 Python bridge 的超时 |
| `RADISHMIND_PLATFORM_BRIDGE_MODE` | `stdio_pool` | `stdio_pool` 或显式 `process_per_request` 回滚模式 |
| `RADISHMIND_PLATFORM_BRIDGE_WORKER_COUNT` | `4` | 持久 worker 数量上限 |
| `RADISHMIND_PLATFORM_BRIDGE_QUEUE_CAPACITY` | `64` | worker 忙时的有界等待容量 |
| `RADISHMIND_PLATFORM_BRIDGE_HANDSHAKE_TIMEOUT` | `5s` | worker 启动握手超时 |
| `RADISHMIND_PLATFORM_PYTHON_BIN` | 仓库 `.venv` Python | Python bridge 解释器 |
| `RADISHMIND_PLATFORM_BRIDGE_SCRIPT` | `scripts/run-platform-bridge.py` | Python bridge 脚本路径 |
| `RADISHMIND_PLATFORM_PROVIDER` | `mock` | 默认 southbound provider |
| `RADISHMIND_PLATFORM_PROVIDER_PROFILE` | 空 | 默认 provider profile |
| `RADISHMIND_PLATFORM_MODEL` | 空 | northbound 默认 model id |
| `RADISHMIND_PLATFORM_BASE_URL` | 空 | 显式 provider base URL 覆盖 |
| `RADISHMIND_PLATFORM_API_KEY` | 空 | 显式 provider API key 覆盖；不得写入文档或提交 |
| `RADISHMIND_PLATFORM_TEMPERATURE` | `0` | provider 调用温度 |
| `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH` | `false` | 显式启用 read-side、workflow saved draft、application draft 与 publish candidate 的 dev-only 测试身份 header |
| `RADISHMIND_CONTROL_PLANE_READ_AUTH_MODE` | `disabled` | `disabled`、`dev_headers`、`signed_test_token` 或受控 `radish_oidc_integration_test` |
| `RADISHMIND_CONTROL_PLANE_READ_STORE` | `fake_store_dev` | `fake_store_dev` 或 `postgres_dev_test`；OIDC integration 只允许后者 |
| `RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_DATABASE_URL` | 空 | Control Plane read PostgreSQL dev/test runtime SELECT 连接；secret |
| `RADISHMIND_CONTROL_PLANE_READ_DATABASE_TIMEOUT` | `5s` | Control Plane read connect / preflight / query timeout |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_ISSUER` | 空 | reviewed exact issuer；仅 HTTPS 或 deterministic loopback HTTP |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_DISCOVERY_URL` | 空 | reviewed exact discovery URL，必须与 issuer 同 origin |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_AUDIENCE` | 空 | reviewed exact RadishMind integration audience |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_MAPPING_VERSION` | 空 | reviewed claim / permission mapping version |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_EVIDENCE_REF` | 空 | sanitized reviewed evidence reference，不是 issuer URL dump |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_SUBJECT_CLAIM` | 空 | reviewed subject claim name |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_TENANT_CLAIM` | 空 | reviewed tenant claim name |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_PERMISSION_CLAIM` | 空 | reviewed permission array claim name |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_TENANT_PERMISSION` | 空 | reviewed tenant-read upstream permission identifier |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_AUDIT_PERMISSION` | 空 | reviewed audit-read upstream permission identifier |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_ALGORITHMS` | 空 | reviewed comma-separated `RS*` / `ES*` allowlist；不允许 HMAC / `none` |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_JWKS_ORIGIN` | 空 | reviewed exact JWKS origin |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_DISCOVERY_TIMEOUT` | `3s` | discovery / JWKS HTTP timeout |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_JWKS_MAX_AGE` | `5m` | cache refresh 上限 |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_JWKS_HARD_EXPIRY` | `15m` | key cache hard expiry，超过后不使用 stale key |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_ROTATION_OVERLAP` | `5m` | retired key 受控 overlap window |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_CLOCK_SKEW` | `30s` | token NumericDate 固定 skew |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_MAX_TOKEN_LIFETIME` | `10m` | access token 最大允许 lifetime |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_MAX_RESPONSE_BYTES` | `262144` | discovery / JWKS 最大响应字节数 |
| `RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_MAX_KEYS` | `32` | bounded JWKS cache 最大 key 数 |
| `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP` | `false` | 显式启用 saved workflow draft dev-only HTTP route |
| `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE` | `false` | 显式允许 saved workflow draft dev-only save 操作；read / validate 不用该开关 |
| `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE` | `memory_dev` | `memory_dev` 或显式 `postgres_dev_test`；production `repository`、`repository_disabled` 和 unknown mode 均 fail closed |
| `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL` | 空 | `postgres_dev_test` 平台运行连接；secret，不进入摘要或日志 |
| `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_MIGRATION_DATABASE_URL` | 空 | 仅 migration `up` 使用的一次性 DDL 连接；secret |
| `RADISHMIND_WORKFLOW_SAVED_DRAFT_DATABASE_TIMEOUT` | `5s` | PostgreSQL connect / preflight timeout |
| `RADISHMIND_APPLICATION_DRAFT_DEV_HTTP` | `false` | 显式启用 application configuration draft dev-only validate / read / list route |
| `RADISHMIND_APPLICATION_DRAFT_DEV_WRITE` | `false` | 显式允许 application configuration draft save |
| `RADISHMIND_APPLICATION_DRAFT_STORE` | `memory_dev` | `memory_dev` 或显式 `postgres_dev_test`；其他 mode fail closed |
| `RADISHMIND_APPLICATION_DRAFT_DEV_TEST_DATABASE_URL` | 空 | application draft PostgreSQL dev/test runtime DML 连接；secret |
| `RADISHMIND_APPLICATION_DRAFT_DEV_TEST_MIGRATION_DATABASE_URL` | 空 | application draft migration 一次性 DDL 连接；secret |
| `RADISHMIND_APPLICATION_DRAFT_DATABASE_TIMEOUT` | `5s` | application draft PostgreSQL connect / preflight timeout |
| `RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP` | `false` | 显式启用 application publish candidate dev-only create / read / list / review route |
| `RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE` | `false` | 显式允许 candidate create / review；不启用正式 promotion |
| `RADISHMIND_APPLICATION_PUBLISH_STORE` | `memory_dev` | `memory_dev` 或显式 `postgres_dev_test`；其他 mode fail closed |
| `RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_DATABASE_URL` | 空 | publish candidate PostgreSQL dev/test runtime DML 连接；secret |
| `RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_MIGRATION_DATABASE_URL` | 空 | publish candidate migration 一次性 DDL 连接；secret |
| `RADISHMIND_APPLICATION_PUBLISH_DATABASE_TIMEOUT` | `5s` | publish candidate PostgreSQL connect / preflight timeout |
| `RADISHMIND_APPLICATION_CATALOG_DEV_HTTP` | `false` | 显式启用 application catalog dev-only list / create / read / update / archive route |
| `RADISHMIND_APPLICATION_CATALOG_DEV_WRITE` | `false` | 显式允许 application catalog create / update / archive |
| `RADISHMIND_APPLICATION_CATALOG_STORE` | `memory_dev` | `memory_dev` 或显式 `postgres_dev_test`；其他 mode fail closed |
| `RADISHMIND_APPLICATION_CATALOG_DEV_TEST_DATABASE_URL` | 空 | application catalog PostgreSQL dev/test runtime DML 连接；secret |
| `RADISHMIND_APPLICATION_CATALOG_DEV_TEST_MIGRATION_DATABASE_URL` | 空 | application catalog migration 一次性 DDL 连接；secret |
| `RADISHMIND_APPLICATION_CATALOG_DATABASE_TIMEOUT` | `5s` | application catalog PostgreSQL connect / preflight timeout |
| `RADISHMIND_WORKFLOW_EXECUTOR_DEV` | `false` | 显式启用受控 Workflow Executor v0 dev-only POST / GET route；不启用完整生产执行器 |
| `RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV` | `false` | 与 dev auth 双 gate 显式启用 Gateway 请求历史记录和 scoped list / detail route；store 由 `RADISHMIND_GATEWAY_REQUEST_STORE` 选择 |
| `RADISHMIND_GATEWAY_REQUEST_STORE` | `memory_dev` | `memory_dev` 或显式 `postgres_dev_test`；reserved production mode 和 unknown mode fail closed |
| `RADISHMIND_GATEWAY_REQUEST_DEV_TEST_DATABASE_URL` | 空 | Gateway request `postgres_dev_test` runtime DML 连接；secret，不进入摘要或日志 |
| `RADISHMIND_GATEWAY_REQUEST_DEV_TEST_MIGRATION_DATABASE_URL` | 空 | 仅 Gateway request manual migration `up` 使用的一次性 DDL 连接；secret |
| `RADISHMIND_GATEWAY_REQUEST_DATABASE_TIMEOUT` | `5s` | Gateway request PostgreSQL connect / preflight timeout |
配置优先级固定为 `default < config file < env`。配置文件当前使用 JSON，字段名与脱敏 summary 保持一致，例如：

```json
{
  "listen_addr": "127.0.0.1:7000",
  "provider": "mock",
  "model": "radishmind-local-dev",
  "bridge_timeout": "30s",
  "workflow_saved_draft_store": "memory_dev"
}
```

可用配置文件启动本地平台服务：

```bash
RADISHMIND_PLATFORM_CONFIG=tmp/radishmind-platform.local.json \
go run ./services/platform/cmd/radishmind-platform
```

推荐通过稳定 wrapper 先跑配置检查和结构化诊断，再启动长驻服务：

```bash
./scripts/run-platform-service.sh config-check
./scripts/run-platform-service.sh diagnostics
./scripts/run-platform-service.sh serve
```

Windows / PowerShell 使用：

```powershell
pwsh ./scripts/run-platform-service.ps1 -Command config-check
pwsh ./scripts/run-platform-service.ps1 -Command diagnostics
pwsh ./scripts/run-platform-service.ps1 -Command serve
```

上层消费 smoke 可以先用离线 fixture 生成展示视图，不要求启动服务：

```bash
./scripts/run-python.sh scripts/run-platform-overview-consumer-smoke.py --check
./scripts/run-python.sh scripts/run-platform-local-smoke.py --check
./scripts/run-python.sh scripts/run-platform-session-tooling-consumer-smoke.py --check
```

服务启动后，也可以指向本地平台 API 生成同一份消费视图：

```bash
./scripts/run-python.sh scripts/run-platform-overview-consumer-smoke.py \
  --base-url http://127.0.0.1:7000 \
  --check

./scripts/run-python.sh scripts/run-platform-local-smoke.py \
  --base-url http://127.0.0.1:7000 \
  --check

./scripts/run-python.sh scripts/run-platform-session-tooling-consumer-smoke.py \
  --base-url http://127.0.0.1:7000 \
  --check
```

overview consumer smoke 只读取 `GET /v1/platform/overview`，把 service status、model inventory、session/tooling surface 和 stop-lines 投影成本地 console view model；local smoke consumer 只读取 `GET /v1/platform/local-smoke`，把本地 readiness 摘要投影为 healthz、overview、model inventory、session/tooling、CORS 和停止线检查；session/tooling consumer smoke 只读取 `session metadata`、`tools metadata` 并提交一次会被阻断的 tool action 请求，用于验证上层可展示 `blocked`、`requires_confirmation` 与 `no_side_effects`。这些 smoke 都不会启用真实 executor、durable store、confirmation、replay 或业务写回。

最小本地 console 壳位于 `apps/radishmind-console/`。它复用 `contracts/typescript/platform-overview-api.ts` 与 `contracts/typescript/platform-local-smoke-api.ts`，默认读取 `http://127.0.0.1:7000/v1/platform/overview` 和 `http://127.0.0.1:7000/v1/platform/local-smoke`：

```bash
cd apps/radishmind-console
npm install
npm run dev
```

该 console 只展示 service status、model/profile inventory、Provider/Profile Details、session/tooling blocked 状态、Blocked Action Detail、Local Readiness、stop-lines、Stop-line Details、audit boundary、Dev Diagnostics、refresh 状态和连接失败诊断，不调用 `/v1/tools/actions`，也不实现 executor、durable store、confirmation、业务写回或 replay。refresh 期间和连接失败后可以保留上一份只读 overview / local-smoke readiness，用于排障，不代表平台会自动恢复执行。

当前页面结构是浅色左侧导航栏、主工作区和右侧 readiness / stop-line 辅助栏；窄屏下改为单列信息顺序。该结构来自 `docs/designs/radishmind-console-ops-surface-v0.pen` 和 [UI 设计规范](../../docs/radishmind-ui-design-spec.md)，但仍只表示本地 ops surface，不表示 production console 或 production packaging 已完成。

平台服务当前只为 `http://127.0.0.1:4000` 与 `http://localhost:4000` 返回本地 console CORS header，并处理 `OPTIONS` preflight；该能力只服务本地 console 开发，不等同于 production CORS policy、正式鉴权或外部公开部署。

console production packaging 仍未完成：`apps/radishmind-console/package.json` 必须保持 `private=true`，不添加 deploy / publish / release 脚本，不提交 `dist/` 或 `node_modules/`。P3 short-close checklist 继续把 production secret backend、process supervisor、部署环境隔离和 console production packaging 标为 `not_satisfied`；当前只固定本地开发入口和最小 deployment smoke。

macOS / Linux / WSL 使用：

```bash
./scripts/run-radishmind-console-dev.sh
```

Windows / PowerShell 使用：

```powershell
pwsh ./scripts/run-radishmind-console-dev.ps1
```

该入口复用 `scripts/run-platform-service.ps1` / `scripts/run-platform-service.sh` 和 `apps/radishmind-console/` 的 `npm run dev`，启动或复用 `http://127.0.0.1:7000` 与 `http://127.0.0.1:4000`，并探测 `http://127.0.0.1:7000/healthz`、`http://127.0.0.1:7000/v1/platform/overview`、`http://127.0.0.1:7000/v1/platform/local-smoke`、本地 console CORS preflight 和 `http://127.0.0.1:4000`。端口冲突时先释放 `7000/4000` 或确认现有服务就是 RadishMind；CORS 失败时确认 console origin 是允许的本地 origin；浏览器 `unsafe port` / `ERR_UNSAFE_PORT` 通常表示端口被浏览器直接拦截，优先回到默认 `4000/7000`。该入口不是 production supervisor，不实现真实 executor、durable store、confirmation、业务写回或 replay。

验证脚本本身时可加 `-ExitAfterProbe` 或 `--exit-after-probe`，让它启动、探测成功后自动停止本次创建的本地进程。

可用一次性命令检查本地配置摘要，输出不会暴露 secret：

```bash
go run ./services/platform/cmd/radishmind-platform config-summary
go run ./services/platform/cmd/radishmind-platform config-check
go run ./services/platform/cmd/radishmind-platform diagnostics
```

`diagnostics` 输出固定为结构化 JSON，主要字段包括：

- `status`：`ok` 或 `error`
- `config`：复用脱敏 `config-summary`
- `checks`：`config_required_fields`、`bridge_provider_registry`、`bridge_provider_inventory` 与 `deployment_readiness`
- `bridge`：Python bridge 脚本、registry / inventory 可用性和 bridge 失败码
- `providers`：provider registry 数量、profile 数量、active profile chain、credential state 计数和 deployment mode 计数
- `providers.selectable_model_ids`：与 `/v1/models` 和请求选择一致的 profile model id 列表
- `failure_codes` / `failure`：启动前可定位的失败边界

## 本地 smoke 验证

服务启动后，在另一个终端执行：

```bash
curl -sS http://127.0.0.1:7000/healthz
curl -sS http://127.0.0.1:7000/v1/platform/overview
curl -sS http://127.0.0.1:7000/v1/platform/local-smoke
curl -sS http://127.0.0.1:7000/v1/models
curl -sS http://127.0.0.1:7000/v1/models/mock
curl -sS http://127.0.0.1:7000/v1/session/metadata
curl -sS 'http://127.0.0.1:7000/v1/session/recovery/checkpoints/session-checkpoint-0001?session_id=radishflow-session-001&turn_id=turn-0003'
curl -sS http://127.0.0.1:7000/v1/tools/metadata
curl -sS http://127.0.0.1:7000/v1/tools/actions \
  -H 'Content-Type: application/json' \
  -d '{"tool_id":"radishflow.suggest_edits.candidate_builder.v1","action":"execute","session_id":"radishflow-session-001","turn_id":"turn-0003"}'
curl -sS http://127.0.0.1:7000/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"radishmind-local-dev","messages":[{"role":"user","content":"请简要说明当前 RadishMind 平台状态。"}]}'
```

预期边界：

- `/healthz` 返回 `status=ok`、`service=radishmind-platform`。
- `/v1/platform/overview` 返回 `platform_overview`，其中 `product_surface.mode=local_read_only_product_shell`，并汇总 `/v1/models`、session metadata、tool metadata 和 blocked action route；所有 executor / durable store / confirmation / writeback / replay 停止线均为 `false`。
- `/v1/platform/local-smoke` 返回 `platform_local_smoke`，其中 `summary.local_console_ready=true` 表示本地只读 console 所需的 healthz、overview、model inventory、session/tooling metadata、blocked action no-side-effects、local CORS origin 和停止线均可读；该 route 只做摘要，不启动服务、不守护进程、不表示生产部署 ready。
- `/v1/models` 返回 OpenAI-compatible `object=list`，并包含 provider registry 与 profile inventory。
- `/v1/models/mock` 可通过精确 lookup 返回 mock provider model。
- `/v1/session/metadata` 返回 `session_metadata`，其中 durable session/checkpoint store、long-term memory、automatic replay 和 business truth write 均为 `false`。
- `/v1/session/recovery/checkpoints/session-checkpoint-0001` 返回 `session_recovery_checkpoint_read_result`，且 `access_policy.metadata_only=true`、`materialized_results_included=false`、`auto_replay_enabled=false`，`result.tool_audit_summary.execution_enabled=false`。
- `/v1/tools/metadata` 返回 `tooling_metadata`，其中 `registry_policy.execution_enabled=false`，每个工具的 execution mode 为 `contract_only`。
- 七条 control-plane read route 默认 fail closed；显式 dev headers、signed test token 或 OIDC integration 请求必须满足对应 auth / store 配置。OIDC integration 只允许 Tenant Summary / Audit 成功，workspace operation 应返回 `workspace_membership_unavailable`。
- workflow saved draft dev route 默认关闭；只有同时设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 和 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP=1` 才能 list / read / validate，保存还必须设置 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE=1`，并带上 `X-RadishMind-Dev-Workflow-Workspace` / `X-RadishMind-Dev-Workflow-Application` 与匹配 scope。`RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE` 默认 `memory_dev`；设置为 `repository_disabled` / `repository` 会返回 `repository_store_disabled`，设置未知值会返回 `invalid_draft_store_mode`，不会把失败请求回退成 sample 或 memory dev 成功。
- application configuration draft dev route 默认关闭；只有同时设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 和 `RADISHMIND_APPLICATION_DRAFT_DEV_HTTP=1` 才能 list / read / validate，保存还要求 `RADISHMIND_APPLICATION_DRAFT_DEV_WRITE=1`、`application_drafts:write` scope，以及匹配的 `X-RadishMind-Dev-Application-Workspace` / `Application` header。store 默认是 `memory_dev`；显式 `postgres_dev_test` 要求手工 migration、独立 runtime DSN 与 marker / checksum preflight，数据库失败不回退内存。
- application publish candidate dev route 默认关闭；只有同时设置 dev auth、`RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP=1` 与匹配 scope / workspace / application header 才能 list / read；create / review 还要求 `RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE=1`。`postgres_dev_test` 同时要求 application draft 也使用 PostgreSQL dev/test，使服务端 draft reload 与 candidate create 保持 durable；数据库失败不回退内存。
- application catalog dev route 默认关闭；只有同时设置 dev auth、`RADISHMIND_APPLICATION_CATALOG_DEV_HTTP=1` 和对应 read / write / archive scope 才能访问，写入还要求 `RADISHMIND_APPLICATION_CATALOG_DEV_WRITE=1`。store 默认是 `memory_dev`；显式 `postgres_dev_test` 要求手工 migration、独立 runtime DSN 与 marker / checksum preflight，数据库失败不回退内存或预置应用列表。
- Gateway request history 默认关闭；只有同时设置 `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1` 与 `RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV=1`，并带完整 `X-RadishMind-Dev-Gateway-Tenant` / `Workspace` / `Consumer` / 可选 `Application` / `Subject` / `Scopes` / `Audit` header 时，三个 northbound route 才创建 scoped sanitized record。`GET /v1/model-gateway/requests` 与 detail 还要求 `gateway_requests:read`。store 默认是 500 条进程内 `memory_dev`；显式 `postgres_dev_test` 要求 manual migration、独立 runtime DSN 和 marker / checksum preflight，数据库失败不回退内存。两种模式都不是 production audit ledger。
- Gateway request migration 使用 `go run ./cmd/radishmind-gateway-request-migrate status|up`；application draft / publish / catalog migration 分别使用 `go run ./cmd/radishmind-application-draft-migrate status|up`、`go run ./cmd/radishmind-application-publish-migrate status|up` 与 `go run ./cmd/radishmind-application-catalog-migrate status|up`。`status` 使用 runtime DSN 且只读 marker，`up` 只接受独立 migration DSN。仓库 PostgreSQL 集成入口会同时验证 Saved Draft、Workflow Run、Gateway Request、Application Draft、Application Publish 与 Application Catalog 六套相互独立的 schema。
- `/v1/tools/actions` 返回 `tool_action_blocked_response`，且不会运行工具、返回 materialized result、写 durable memory 或写业务真相源。
- `/v1/chat/completions` 在 `mock` provider 下返回 advisory 文本，不访问外部 provider，不写回任何上层项目。
## 故障边界

- 启动前先运行 `./scripts/run-platform-service.sh diagnostics`；如果 `status=error`，优先读取 `failure.code` 和 `checks[].code`，再决定是否启动长驻服务。
- 若 `failure.code=CONFIG_REQUIRED_FIELDS_MISSING`，优先检查 `config.missing_required_fields`，不要从日志或诊断输出寻找 secret 原文。
- 若 `failure.code=PROVIDER_REGISTRY_UNAVAILABLE` 或 `PROVIDER_INVENTORY_UNAVAILABLE`，优先检查 `bridge.python_binary`、`bridge.script`、当前工作目录和 Python import 路径。
- 若启动时报 `load config`，优先检查 duration / float 类环境变量格式，例如 `RADISHMIND_PLATFORM_BRIDGE_TIMEOUT=30s`、`RADISHMIND_PLATFORM_TEMPERATURE=0`。
- 若 `/v1/models` 返回 `PROVIDER_INVENTORY_UNAVAILABLE`，优先检查 `RADISHMIND_PLATFORM_PYTHON_BIN`、`RADISHMIND_PLATFORM_BRIDGE_SCRIPT` 和当前工作目录是否能访问仓库根下的 Python bridge。
- 若 northbound 路由返回 `BRIDGE_WORKER_*`，先运行 diagnostics 检查 mode、握手和 inventory；必要时可显式设置 `RADISHMIND_PLATFORM_BRIDGE_MODE=process_per_request` 复核回滚路径，但不得恢复 argv 或长期环境变量传递请求 credential。
- 若返回 `MODEL_NOT_FOUND`，优先用 `/v1/models` 或 `diagnostics.providers.selectable_model_ids` 确认可选择 ID，避免把 provider id、profile id 与真实 upstream model 混用。
- 若返回 `PLATFORM_RESPONSE_INVALID`，说明 Python bridge 已返回 envelope，但 `Go` northbound 兼容层无法翻译为目标协议响应，应优先检查 envelope 的 `response` 结构。
- 若 saved workflow draft route 返回 `repository_store_disabled` 或 `invalid_draft_store_mode`，优先检查 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE`；只有 `memory_dev` 与完整显式配置的 `postgres_dev_test` 可成功，production `repository` 仍关闭。
- 若 saved workflow draft repository adapter 返回 schema / auth 类 failure code，先确认调用方是否误把 readiness artifact 当成 runtime：`draft_schema_migration_not_applied`、`draft_store_schema_version_mismatch` 和 `draft_store_migration_unavailable` 表示 schema preflight 不满足；`draft_auth_context_contract_mismatch`、`draft_identity_context_missing`、`draft_tenant_binding_missing`、`draft_workspace_membership_denied`、`draft_application_scope_denied`、`draft_owner_scope_denied`、`draft_scope_grant_missing` 和 `draft_audit_context_missing` 表示上游 verified auth context 或 binding 不满足。
- 旧 Saved Draft repository / migration / connection readiness 专题保留为历史设计证据；当前实现真相以 `saved-workflow-draft-postgresql-dev-test-repository-v1`、真实 migration 和 Go 集成测试为准。它们只放宽 `postgres_dev_test`，不启用 production secret resolver 或 production `repository`。
- 若要接真实 provider，必须通过环境变量或本机 secret 注入 `RADISHMIND_PLATFORM_BASE_URL` / `RADISHMIND_PLATFORM_API_KEY` 或 provider profile 配置；不要把 key、token、cookie 或真实 provider raw dump 写入 committed 文档。
