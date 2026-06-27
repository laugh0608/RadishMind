# RadishMind 平台专题入口

更新时间：2026-06-27

## 文档目的

本目录用于承接跨产品面的平台专题。平台专题不属于某一个页面，也不应塞进单个产品面大方向文档；它们负责固定 auth、store、repository、provider、deployment、dev-only write path 等长期边界和准入条件。

具体实现批次仍进入 `docs/task-cards/`。平台专题只说明为什么要做、允许打开什么、依赖什么证据、哪些能力必须作为独立目标。

## 何时放在这里

- 能力会影响多个产品面，例如 User Workspace、Admin Control Plane 和 Workflow 同时依赖的 auth / store / repository。
- 能力属于服务层或运行层，例如 provider runtime、deployment runtime config、secret backend、store selector。
- 能力需要明确 dev-only、test、production 的启用边界。
- 能力不是单个页面可独立验收的 UI 组织问题。

## 当前平台专题候选

| 专题 | 当前状态 | 现有事实源 | 下一步 |
| --- | --- | --- | --- |
| Auth / Store Transition | readiness 已有证据；Radish OIDC token / membership readiness 已固定 `radish_oidc_token_membership_readiness_defined`，implementation entry review 已固定 `radish_oidc_token_membership_implementation_entry_review_defined`，upstream evidence refresh 已固定 `radish_oidc_token_membership_upstream_evidence_refresh_defined` | `docs/contracts/control-plane-read-side.md`、相关 control-plane read task cards、[Radish OIDC Token / Membership Readiness v1](../integrations/radish-oidc-token-membership-readiness-v1.md)、[Radish OIDC Token / Membership Implementation Entry Review v1](../integrations/radish-oidc-token-membership-implementation-entry-review-v1.md)、[Radish OIDC Token / Membership Upstream Evidence Refresh v1](../integrations/radish-oidc-token-membership-upstream-evidence-refresh-v1.md) | 后续应先重新评审 schema / middleware / membership / runtime smoke 是否可进入任务卡；不直接创建 auth middleware、token validator、membership adapter、repository mode 或 production API |
| Repository Adapter / Store Selector | control-plane readiness 已有证据；saved workflow draft selector 已按 fail-closed mode selection 落地 | `control-plane-read-*repository*`、`*store-selector*` task cards、`docs/features/workflow/saved-workflow-draft-store-selector-implementation-v1.md` | 后续 repository adapter / database 仍需独立实现专题 |
| Provider Runtime & Health | close candidate | [Provider Runtime & Health v1 任务卡](../task-cards/provider-runtime-health-v1-plan.md) | 不继续扩同层 provider 小切片 |
| Production Ops / Deployment | 静态边界已 close；secret ref config、provider profile binding、disabled resolver interface、operator runbook / negative gates、rotation / audit policy readiness、test fixture strategy / fake resolver entry review、fake resolver contract / no secret leakage smoke strategy、fake resolver implementation task card entry readiness、fake resolver implementation task card、fake resolver runtime implementation entry review、test-only fake resolver runtime、真实 resolver runtime preconditions、真实 resolver runtime implementation entry review、resolver backend profile selection readiness、真实 resolver no leakage smoke runtime strategy、真实 resolver no leakage smoke runtime implementation entry review / refresh、credential handle runtime boundary readiness、operator approval runtime evidence readiness、audit store handoff readiness、resolver backend health boundary readiness、resolver backend health runtime implementation entry review、audit store runtime implementation entry review、audit store contract / event schema readiness、audit store runtime implementation entry refresh v2、audit store ownership boundary readiness、audit store delivery / idempotency runtime boundary readiness、audit store runtime implementation entry refresh v3、audit store durable backend boundary readiness（`audit_store_durable_backend_boundary_readiness_defined`）、audit store writer runtime boundary readiness（`audit_store_writer_runtime_boundary_readiness_defined`）、runtime event schema materialization readiness（`audit_store_runtime_event_schema_materialization_readiness_defined`）、delivery runtime readiness（`audit_store_delivery_runtime_readiness_defined`）、idempotency runtime readiness（`audit_store_idempotency_runtime_readiness_defined`）与 audit store runtime implementation entry refresh v4（`audit_store_runtime_implementation_entry_refresh_v4_defined`）、operator approval runtime implementation entry review、credential handle runtime implementation entry review、real resolver runtime implementation entry refresh、production resolver runtime blocker consolidation、credential handle runtime implementation entry refresh、operator approval runtime implementation entry refresh、cloud secret service selection readiness、backend health runtime implementation entry refresh 与 no leakage smoke runtime implementation entry refresh 已实现 | [Production Ops Hardening v1](../task-cards/production-ops-hardening-v1-plan.md)、[Docker Deployment v1](../task-cards/production-ops-docker-deployment-v1-plan.md)、[Production Secret Backend Config / Secret Ref Readiness v1](production-secret-backend-config-secret-ref-readiness-v1.md)、[Production Secret Backend Provider Profile Secret Binding Readiness v1](production-secret-backend-provider-profile-secret-binding-readiness-v1.md)、[Production Secret Backend Secret Resolver Interface Disabled Readiness v1](production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md)、[Production Secret Backend Real Resolver Runtime Implementation Entry Refresh v1](production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.md)、[Production Secret Backend Production Resolver Runtime Blocker Consolidation v1](production-secret-backend-production-resolver-runtime-blocker-consolidation-v1.md)、[Production Secret Backend Credential Handle Runtime Implementation Entry Refresh v1](production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.md)、[Production Secret Backend Operator Approval Runtime Implementation Entry Refresh v1](production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.md)、[Production Secret Backend Cloud Secret Service Selection Readiness v1](production-secret-backend-cloud-secret-service-selection-readiness-v1.md)、[Production Secret Backend Resolver Backend Health Runtime Implementation Entry Refresh v1](production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.md)、[Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Implementation Entry Refresh v1](production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1.md)、[Production Secret Backend Operator Approval Runtime Implementation Entry Review v1](production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.md)、[Production Secret Backend Audit Store Runtime Implementation Entry Refresh v3](production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.md)、[Production Secret Backend Audit Store Durable Backend Boundary Readiness v1](production-secret-backend-audit-store-durable-backend-boundary-readiness-v1.md)、[Production Secret Backend Audit Store Writer Runtime Boundary Readiness v1](production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.md)、[Production Secret Backend Audit Store Runtime Event Schema Materialization Readiness v1](production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.md)、[Production Secret Backend Audit Store Delivery Runtime Readiness v1](production-secret-backend-audit-store-delivery-runtime-readiness-v1.md)、[Production Secret Backend Audit Store Idempotency Runtime Readiness v1](production-secret-backend-audit-store-idempotency-runtime-readiness-v1.md)、[Production Secret Backend Audit Store Runtime Implementation Entry Refresh v4](production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.md)、[Production Secret Backend Resolver Backend Health Runtime Implementation Entry Review v1](production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.md) | 下一步若继续 secret backend，应补齐真实 runtime 依赖或重新评审单项 runtime 边界；不得直接创建 audit store runtime task card、writer runtime task card、delivery runtime task card、idempotency runtime task card、runtime event schema implementation task card、no leakage smoke runtime task card、backend health runtime task card 或 production resolver runtime task card |
| Dev-only Write Path | implemented | [Dev-only Saved Draft Consumer](../features/workflow/dev-only-saved-draft-consumer.md) | 已服务 saved draft consumer integration；后续按 conflict UI / smoke / contract 固化拆批次 |

## Saved Workflow Draft Store Selector

Saved workflow draft 的平台服务配置已经新增 `workflow_saved_draft_store` / `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE`。当前唯一可成功读写的 mode 是 `memory_dev`，它只服务 dev-only saved draft route；`repository_disabled` 和 `repository` 都返回 `repository_store_disabled`，unknown mode 返回 `invalid_draft_store_mode`。这些失败必须直接暴露给 save / read / list 调用方，不得回退到 memory dev、sample 或 fixture。

`services/platform/migrations/workflow_saved_drafts/` 当前只承载 `manifest.json`、`ddl-review.md`、`rollback-evidence.json` 和 `migration-smoke.json` 四个静态 schema artifact 证据文件。它们说明 future durable store 的 logical schema、predicate、review 和 rollback 边界，不是 SQL migration，不会被 platform service 自动执行，也不表示 repository mode、真实数据库或 production auth 已可用。

`workflow-saved-draft-repository-mode-runtime-boundary-review-v1` 已固定 `draft_repository_mode_runtime_boundary_review_defined`。该评审消费 repository adapter、adapter smoke、production auth runtime bridge、runner / connection / resolver entry review、audit store runtime entry refresh v3 和 production secret backend implementation readiness 证据，结论仍是 repository mode runtime task card blocked；不启用 repository store mode，不创建真实 query executor、schema marker runtime、OIDC / membership、production API、audit store runtime、executor、confirmation、writeback 或 replay。

`workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1` 已固定 `draft_schema_marker_migration_runner_readiness_refresh_defined`。该 refresh 只收束 applied marker、manual runner、dry-run、idempotency / lock、duplicate handling 和 rollback observability 的静态前置；不创建 schema marker implementation task card、migration runner implementation task card、SQL、schema version table、marker runtime、runner、DB provider 或 repository mode runtime。

`workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1` 已固定 `draft_database_driver_dsn_tls_policy_readiness_defined`。该专题只定义 future DB driver selection、DSN construction / redaction boundary、TLS policy、environment binding、forbidden diagnostics scan、role policy dependency 和 connection smoke 前置；不选择或导入 DB driver、不创建 DSN parser、不创建 TLS runtime、不创建 connection factory、不运行 connection smoke。

`workflow-saved-draft-database-role-policy-readiness-v1` 已固定 `draft_database_role_policy_readiness_defined`。该专题只定义 future runtime DML role、migration / marker role、least privilege review、environment binding、role claim metadata boundary 和 cross-environment denial smoke 前置；不创建 role runtime、不创建 grant、不解析 token claim、不连接数据库、不运行 SQL。

`workflow-saved-draft-database-connection-smoke-strategy-v1` 已固定 `draft_database_connection_smoke_strategy_defined`。该专题只定义 future explicit test database、metadata-only credential handoff、smoke input / output shape、role denial cases、no leakage scan 和 manual execution boundary；不创建 smoke runner、不执行 smoke、不提交 smoke 输出、不连接数据库，也不把该策略接入 fast baseline 的运行时数据库检查。

`workflow-saved-draft-database-connection-lifecycle-readiness-v1` 已固定 `draft_database_connection_lifecycle_readiness_defined`。该专题只定义 future timeout、pool、health check、close responsibility、request / audit propagation 和 sanitized diagnostics runtime 前置；不创建 lifecycle runtime、connection factory、DB provider、driver、SQL、schema marker 或 repository mode runtime。

## Production Secret Backend Config / Secret Ref Readiness

`Production Secret Backend Config / Secret Ref Readiness v1` 已固定 `config_secret_ref_readiness_defined`，对应切片为 `production-secret-backend-config-secret-ref-readiness-v1`。它只把 `production-secret-backend-implementation-readiness` 中的 `config-secret-ref-readiness` 前置条件推进到可检查状态，要求后续配置层只能处理 `secret_ref` 存在性、`secret_backend_configured`、`secret_ref_present`、`missing_secret_refs` 和 `field_sources` 等脱敏状态。

该专题不修改 `config.LoadFromEnv` runtime，不实现 resolver、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不接 database connection provider，也不启用 workflow saved draft repository mode。`production_secret_backend` 仍为 `not_satisfied`，`resolver_implementation_status` 仍为 `not_started`。

## Production Secret Backend Provider Profile Secret Binding Readiness

`Production Secret Backend Provider Profile Secret Binding Readiness v1` 已固定 `provider_profile_secret_binding_readiness_defined`，对应切片为 `production-secret-backend-provider-profile-secret-binding-readiness-v1`。它只把 `production-secret-backend-implementation-readiness` 中的 `provider-profile-secret-binding` 前置条件推进到可检查状态，要求后续 provider/profile inventory 只能声明 `credential_requirement`、`secret_ref_status`、`secret_ref_present`、`missing_secret_refs` 和 `field_sources` 等脱敏状态。

该专题不修改 provider runtime，不实现 resolver、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不把 `secret_ref_status=present` 写成 credential resolved、不接 database connection provider，也不启用 workflow saved draft repository mode。`production_secret_backend` 仍为 `not_satisfied`，`resolver_implementation_status` 仍为 `not_started`。

## Production Secret Backend Secret Resolver Interface Disabled Readiness

`Production Secret Backend Secret Resolver Interface Disabled Readiness v1` 已固定 `secret_resolver_interface_disabled_readiness_defined`，对应切片为 `production-secret-backend-secret-resolver-interface-disabled-readiness-v1`。它只把 `production-secret-backend-implementation-readiness` 中的 `secret-resolver-interface-disabled` 推进到可检查状态，定义 future resolver interface 的 reference-only 输入、disabled result、sanitized diagnostics、failure mapping、no fallback、no side effects 和 artifact guard。

该专题不实现 resolver runtime、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不访问 provider、不创建 credential handle、不接 database connection provider，也不启用 workflow saved draft repository mode。`production_secret_backend` 仍为 `not_satisfied`，`resolver_implementation_status` 仍为 `not_started`。operator runbook / negative gates 已由下一节固定为独立 readiness；它仍不代表 resolver、fake resolver、云 secret backend 或 production ready。

## Production Secret Backend Operator Runbook / Negative Gates Readiness

`Production Secret Backend Operator Runbook / Negative Gates Readiness v1` 已固定 `operator_runbook_negative_gates_readiness_defined`，对应切片为 `production-secret-backend-operator-runbook-negative-gates-readiness-v1`。它只把 `production-secret-backend-implementation-readiness` 中的 `operator-runbook-and-negative-gates` 推进到可检查状态，定义人工启用、test / production secret source、operator approval evidence、sanitized verification、smoke record reference、negative gate review、failure mapping、no fallback、no side effects 和 artifact guard。

该专题不实现 resolver runtime、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不访问 provider、不创建 credential handle、不接 database connection provider，也不启用 workflow saved draft repository mode。`production_secret_backend` 仍为 `not_satisfied`，`resolver_implementation_status` 仍为 `not_started`，rotation / audit policy 已由下一节固定为独立 readiness；test fixture strategy 仍是后续阻塞项。

## Production Secret Backend Rotation / Audit Policy Readiness

`Production Secret Backend Rotation / Audit Policy Readiness v1` 已固定 `rotation_audit_policy_readiness_defined`，对应切片为 `production-secret-backend-rotation-audit-policy-readiness-v1`。它只把 `production-secret-backend-implementation-readiness` 中的 `rotation-and-audit-policy` 推进到可检查状态，定义 rotation trigger、approval / change window、secret ref versioning、rollback / disable policy、sanitized verification、audit event fields、failure mapping、no fallback、no side effects 和 artifact guard。

该专题不实现 rotation runtime、不写 production secret audit store、不创建 audit writer、不实现 resolver runtime、不创建 fake resolver、不调用云 secret 服务、不读取 secret value、不访问 provider、不创建 credential handle、不接 database connection provider，也不启用 workflow saved draft repository mode。`production_secret_backend` 仍为 `not_satisfied`，`resolver_implementation_status` 仍为 `not_started`，test fixture strategy 仍是后续阻塞项。后续若继续 production secret backend，应重新评审 test fixture strategy / fake resolver implementation 是否打开。

## Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review

`Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review v1` 已固定 `test_fixture_strategy_fake_resolver_entry_review_defined`，对应切片为 `production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1`。它只把 `production-secret-backend-implementation-readiness` 中仍 blocked 的 `test-fixture-strategy` 与 fake resolver implementation entry 评审为可检查证据，结论是 fake resolver implementation entry 仍不打开。

该专题不实现 resolver runtime、不实现 fake resolver runtime、不创建 no secret leakage smoke runtime、不调用云 secret 服务、不读取 secret value、不访问 provider、不创建 credential handle、不接 database connection provider，也不启用 workflow saved draft repository mode。`production_secret_backend` 仍为 `not_satisfied`，`resolver_implementation_status` 仍为 `not_started`，`fake_resolver_status` 仍为 `not_created`。

## Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy

`Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy v1` 已固定 `fake_resolver_contract_no_secret_leakage_smoke_strategy_defined`，对应切片为 `production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1`。它只定义 fake resolver 的静态输入 / 输出 allowlist、禁止字段、no secret leakage smoke strategy、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。

该专题不实现 resolver runtime、不实现 fake resolver runtime、不创建 no secret leakage smoke runtime 或 fixture、不调用云 secret 服务、不读取 secret value、不访问 provider、不创建 credential handle、不接 database connection provider，也不启用 workflow saved draft repository mode。`test-fixture-strategy` 仍为 `required_before_implementation`，fake resolver implementation entry 仍未打开。

## Production Secret Backend Fake Resolver Implementation Task Card Entry Readiness Review

`Production Secret Backend Fake Resolver Implementation Task Card Entry Readiness Review v1` 已固定 `fake_resolver_implementation_task_card_entry_readiness_review_defined`，对应切片为 `production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1`。它只确认下一步可以创建 fake resolver implementation task card；后续任务卡已由 `production-secret-backend-fake-resolver-implementation-v1` 单独创建。

该专题不实现 resolver runtime、不实现 fake resolver runtime、不创建 no secret leakage smoke runtime 或 fixture、不调用云 secret 服务、不读取 secret value、不访问 provider、不创建 credential handle、不接 database connection provider，也不启用 workflow saved draft repository mode。`test-fixture-strategy` 仍为 `required_before_implementation`，production secret backend 仍为 `not_satisfied`。

## Production Secret Backend Fake Resolver Implementation

`Production Secret Backend Fake Resolver Implementation v1` 已固定 `fake_resolver_implementation_task_card_defined`，对应切片为 `production-secret-backend-fake-resolver-implementation-v1`。它只创建 fake resolver implementation 静态任务卡、fixture、checker 和入口同步，并把 `fake_resolver_implementation_task_card_status` 推进为 `created_static_task_card`。

该专题不实现 resolver runtime、不实现 fake resolver runtime、不创建 no secret leakage smoke runtime 或 fixture、不调用云 secret 服务、不读取 secret value、不访问 provider、不创建 credential handle、不接 database connection provider，也不启用 workflow saved draft repository mode。`test-fixture-strategy` 仍为 `required_before_implementation`，production secret backend 仍为 `not_satisfied`。

## Production Secret Backend Fake Resolver Runtime Implementation Entry Review

`Production Secret Backend Fake Resolver Runtime Implementation Entry Review v1` 已固定 `fake_resolver_runtime_implementation_entry_review_defined`，对应切片为 `production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1`。它只评审现有 fake resolver implementation 静态任务卡是否足以进入下一张 test-only runtime implementation 任务卡，结论为 `fake_resolver_runtime_implementation_ready_for_next_task`。

该专题不把 test-only fake resolver runtime 写成 production resolver、不创建 no secret leakage smoke runtime、不实现 resolver runtime、不调用云 secret 服务、不读取 secret value、不访问 provider、不创建 credential handle、不接 database connection provider，也不启用 workflow saved draft repository mode。`test-fixture-strategy` 仍为 `required_before_implementation`，production secret backend 仍为 `not_satisfied`。

## Production Secret Backend Fake Resolver Runtime Implementation

`Production Secret Backend Fake Resolver Runtime Implementation v1` 已固定 `fake_resolver_runtime_test_only_implemented`，对应切片为 `production-secret-backend-fake-resolver-runtime-implementation-v1`。它实现 test-only、默认 disabled 的 fake resolver runtime，并通过平台专题、task card、fixture、checker、Go runtime 与 Go 单测固定 placeholder secret ref、environment binding、opaque metadata、sanitized diagnostics、offline no leakage smoke 和 side effect counters。

该专题已新增 `services/platform/internal/secretbackend/fake_resolver.go` 与 `fake_resolver_test.go`。runtime 默认 disabled，只能由测试显式启用；不实现 production resolver runtime、不创建 no secret leakage smoke runtime、不调用云 secret 服务、不读取 secret value、不访问 provider、不创建 credential payload、不接 database connection provider，也不启用 workflow saved draft repository mode。`test-fixture-strategy` 仅满足 test-only fake resolver runtime，production secret backend 仍为 `not_satisfied`。

## Production Secret Backend Real Resolver Runtime Preconditions

`Production Secret Backend Real Resolver Runtime Preconditions v1` 已固定 `real_resolver_runtime_preconditions_defined`，对应切片为 `production-secret-backend-real-resolver-runtime-preconditions-v1`。它只定义真实 resolver runtime implementation 之前的启用条件、secret ref / provider profile / environment binding 前置、operator approval、audit / rotation dependency、failure mapping、sanitized diagnostics、no fallback、no side effects、artifact guard 和后续实现拆分。

该专题不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 credential payload、不启用 workflow saved draft repository mode，也不新增 public production API。`fake_resolver_runtime_test_only_implemented` 仍只表示离线测试替身已可用，production secret backend 仍为 `not_satisfied`。

## Production Secret Backend Real Resolver Runtime Implementation Entry Review

`Production Secret Backend Real Resolver Runtime Implementation Entry Review v1` 已固定 `real_resolver_runtime_implementation_entry_review_defined`，对应切片为 `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1`。它评审真实 resolver runtime implementation task card 是否可以创建，结论为 `real_resolver_runtime_implementation_blocked_before_task_card`。

该专题确认真实 resolver runtime implementation task card 当时仍 blocked。后续已由 resolver backend profile selection readiness、no leakage smoke runtime strategy、credential handle runtime boundary readiness、operator approval runtime evidence readiness、audit store handoff readiness、resolver backend health boundary readiness 和 resolver backend health runtime implementation entry review 固定静态前置或 blocked-before-task-card 证据；当前仍不在本批创建 production resolver runtime implementation task card，不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不连接数据库，不创建 credential payload，也不启用 repository mode 或 public production API。

## Production Secret Backend Real Resolver Runtime Implementation Entry Refresh

`Production Secret Backend Real Resolver Runtime Implementation Entry Refresh v1` 已固定 `real_resolver_runtime_implementation_entry_refresh_defined`，对应切片为 `production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1`。它消费真实 resolver runtime preconditions、原始 entry review、backend profile selection readiness、no leakage smoke runtime strategy、credential handle runtime boundary 与 implementation entry review、operator approval runtime evidence 与 implementation entry review、audit store handoff 与 runtime entry review、resolver backend health boundary 与 runtime entry review，并把 production resolver runtime implementation task card 的当前结论固定为 `real_resolver_runtime_implementation_still_blocked_before_task_card`。

该专题不创建 production resolver runtime implementation task card、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 credential payload、不创建 credential handle、不创建或执行 approval runtime、不创建 audit store runtime、不创建 backend health runtime、不创建 no secret leakage smoke runtime，也不启用 repository mode 或 public production API。后续若继续 secret backend，应先单独处理仍 blocked 的 runtime 依赖，或单独评审 production resolver runtime task card entry readiness。

## Production Secret Backend Production Resolver Runtime Blocker Consolidation

`Production Secret Backend Production Resolver Runtime Blocker Consolidation v1` 已固定 `production_resolver_runtime_blocker_consolidation_defined`，对应切片为 `production-secret-backend-production-resolver-runtime-blocker-consolidation-v1`。它消费 real resolver runtime implementation entry refresh、audit store runtime implementation entry refresh v3、credential handle / operator approval / backend health / no leakage smoke runtime entry review、Saved Workflow Draft database secret resolver runtime dependency refresh 与 negative auth smoke runtime readiness，把 production resolver runtime task card 的当前结论固定为 `production_resolver_runtime_task_card_still_blocked_after_consolidation`。

该专题只把 credential handle runtime、operator approval runtime、audit store runtime、backend health runtime、no leakage smoke runtime、cloud secret service、workflow database secret resolver、negative auth smoke runtime 和 repository mode blocker 收束成矩阵；不创建 production resolver runtime implementation task card、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 credential payload、不执行 approval / health / smoke runtime、不启用 repository mode 或 public production API。后续若继续 secret backend，应选择单个仍 blocked 的 runtime 依赖继续拆解。

## Production Secret Backend Credential Handle Runtime Implementation Entry Refresh

`Production Secret Backend Credential Handle Runtime Implementation Entry Refresh v1` 已固定 `credential_handle_runtime_implementation_entry_refresh_defined`，对应切片为 `production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1`。它消费 credential handle runtime implementation entry review、production resolver runtime blocker consolidation、operator approval / audit store / backend health / no leakage runtime entry evidence、resolver backend profile selection readiness、implementation readiness 和 secret reference 证据，把 credential handle runtime implementation task card 的当前结论固定为 `credential_handle_runtime_task_card_still_blocked_after_refresh`。

该专题只复评 credential handle runtime 的剩余依赖；不创建 credential handle runtime implementation task card、不实现 credential handle runtime、handle issuer、handle lifecycle runtime、credential handle、credential payload、production resolver runtime、approval runtime、audit store runtime、backend health runtime、no leakage smoke runtime、database runtime、repository mode runtime 或 public production API。后续已补齐 operator approval runtime refresh、cloud secret service selection readiness、backend health runtime implementation entry refresh 与 no leakage smoke runtime implementation entry refresh；若继续 secret backend，应选择 audit store 中仍 blocked 的单项依赖继续拆解。

## Production Secret Backend Operator Approval Runtime Implementation Entry Refresh

`Production Secret Backend Operator Approval Runtime Implementation Entry Refresh v1` 已固定 `operator_approval_runtime_implementation_entry_refresh_defined`，对应切片为 `production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1`。它消费 operator approval runtime implementation entry review、credential handle runtime implementation entry refresh、production resolver runtime blocker consolidation、audit store runtime entry refresh v3、backend health runtime entry review、no leakage smoke runtime entry review、resolver backend profile selection readiness、implementation readiness 和 secret reference 证据，把 operator approval runtime implementation task card 的当前结论固定为 `operator_approval_runtime_task_card_still_blocked_after_refresh`。

该专题只复评 operator approval runtime 的剩余依赖；不创建 operator approval runtime implementation task card、不实现 operator approval runtime、approval executor、operator identity provider、dual control verifier、ticket / change window verifier、policy evaluator、credential handle runtime、audit store runtime、backend health runtime、no leakage smoke runtime、production resolver runtime、database runtime、repository mode runtime 或 public production API。后续已补齐 cloud secret service selection readiness、backend health runtime implementation entry refresh 与 no leakage smoke runtime implementation entry refresh；若继续 secret backend，应选择 audit store 中仍 blocked 的单项依赖继续拆解，不选择具体云厂商或创建 cloud secret client。

## Production Secret Backend Resolver Backend Profile Selection Readiness

`Production Secret Backend Resolver Backend Profile Selection Readiness v1` 已固定 `resolver_backend_profile_selection_readiness_defined`，对应切片为 `production-secret-backend-resolver-backend-profile-selection-readiness-v1`。它把真实 resolver runtime entry review 中的 backend profile selection blocker 拆成独立静态前置证据，固定 backend profile shape、reserved backend kind allowlist、environment / provider profile / policy version binding、operator approval、audit / rotation dependency、backend health reference、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。

该专题不创建 backend runtime、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 credential payload 或 credential handle runtime，也不启用 repository mode 或 public production API。真实 resolver no leakage smoke runtime strategy、credential handle runtime boundary readiness、operator approval runtime evidence readiness、audit store handoff readiness 和 backend health boundary readiness 已由后续专题固定；真实 resolver runtime implementation task card 仍必须作为后续独立目标评审。

## Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Strategy

`Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Strategy v1` 已固定 `real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`，对应切片为 `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1`。它把真实 resolver runtime entry review 中的 no-secret-leakage smoke runtime strategy blocker 拆成独立静态前置证据，固定 future smoke runtime gate 的离线策略、扫描面、输入 / 输出 allowlist、负向 probe、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。

该专题不创建 no secret leakage smoke runtime、不执行 resolver smoke、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 credential payload 或 credential handle runtime，也不启用 repository mode 或 public production API。后续已固定 no leakage smoke runtime implementation entry review、credential handle runtime boundary readiness、operator approval runtime evidence readiness、audit store handoff readiness、backend health boundary readiness 与 backend health runtime entry review；真实 resolver runtime implementation task card 仍必须作为后续独立目标评审。

## Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Implementation Entry Review / Refresh

`Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Implementation Entry Review v1` 已固定 `real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined`；后续 `Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Implementation Entry Refresh v1` 已固定 `real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_refresh_defined`，对应切片为 `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1`。

该 refresh 消费原始 entry review、cloud secret service selection readiness、backend health refresh、production resolver blocker consolidation、credential handle / operator approval refresh 和 audit store refresh v3，entry decision 为 `real_resolver_no_secret_leakage_smoke_runtime_task_card_still_blocked_after_refresh`；不创建 no leakage smoke runtime implementation task card、不创建 smoke runtime、runner、artifact scanner 或 output fixture、不执行 smoke、不调用 fake resolver 或 production resolver、不创建 production resolver runtime、不调用云 secret 服务、不接数据库、不启用 repository mode 或 public production API。

## Production Secret Backend Credential Handle Runtime Boundary Readiness

`Production Secret Backend Credential Handle Runtime Boundary Readiness v1` 已固定 `credential_handle_runtime_boundary_readiness_defined`，对应切片为 `production-secret-backend-credential-handle-runtime-boundary-readiness-v1`。它把真实 resolver runtime entry review 中的 credential handle blocker 拆成独立静态前置证据，固定 future opaque credential handle 的 reference 定义、允许 metadata、禁止 payload / secret material、secret ref / provider profile / environment binding 前置、operator approval / audit / rotation 依赖、handle lifecycle、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。

该专题不创建 credential handle runtime、不创建 credential payload、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 no secret leakage smoke runtime，也不启用 repository mode 或 public production API。后续已固定 operator approval runtime evidence readiness、audit store handoff readiness、backend health boundary readiness 与 backend health runtime entry review；真实 resolver runtime implementation task card 仍必须作为后续独立目标评审。

## Production Secret Backend Credential Handle Runtime Implementation Entry Review

`Production Secret Backend Credential Handle Runtime Implementation Entry Review v1` 已固定 `credential_handle_runtime_implementation_entry_review_defined`，对应切片为 `production-secret-backend-credential-handle-runtime-implementation-entry-review-v1`。它消费 credential handle runtime boundary readiness，并把 credential handle runtime task card 的入口结论固定为 `credential_handle_runtime_implementation_blocked_before_task_card`。

该专题只固定 entry decision、blocked conditions、future runtime task card requirements、failure mapping、sanitized diagnostics、no fallback、no side effects、artifact guard、real resolver entry review alignment 和 implementation readiness alignment；不创建 credential handle runtime implementation task card、不创建 credential handle runtime、不创建 handle issuer、不创建 credential handle 或 payload、不执行 handle issuance、不执行 approval runtime、不创建 audit store runtime、不创建 backend health runtime、不实现 production resolver runtime、不启用 repository mode 或 public production API。

## Production Secret Backend Operator Approval Runtime Evidence Readiness

`Production Secret Backend Operator Approval Runtime Evidence Readiness v1` 已固定 `operator_approval_runtime_evidence_readiness_defined`，对应切片为 `production-secret-backend-operator-approval-runtime-evidence-readiness-v1`。它把真实 resolver runtime entry review 中的 operator approval runtime evidence blocker 拆成独立静态前置证据，固定 future operator approval runtime evidence 的 evidence shape、允许 metadata、禁止 payload / secret material、approval subject binding、secret ref / provider profile / environment / backend profile binding、credential handle boundary、no leakage strategy、operator identity、dual control、ticket / change window、audit / rotation dependency、lifecycle、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。

该专题不创建 operator approval runtime、不执行 approval runtime、不创建 approval executor、不写 audit store、不创建 credential handle runtime、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不执行 no secret leakage smoke runtime，也不启用 repository mode 或 public production API。后续已固定 audit store handoff readiness、backend health boundary readiness 与 backend health runtime entry review；真实 resolver runtime implementation task card 仍必须作为后续独立目标评审。

## Production Secret Backend Operator Approval Runtime Implementation Entry Review

`Production Secret Backend Operator Approval Runtime Implementation Entry Review v1` 已固定 `operator_approval_runtime_implementation_entry_review_defined`，对应切片为 `production-secret-backend-operator-approval-runtime-implementation-entry-review-v1`。它消费 operator approval runtime evidence readiness，并把 operator approval runtime task card 的入口结论固定为 `operator_approval_runtime_implementation_blocked_before_task_card`。

该专题只固定 entry decision、blocked conditions、future runtime task card requirements、failure mapping、sanitized diagnostics、no fallback、no side effects、artifact guard、real resolver entry review alignment 和 implementation readiness alignment；不创建 operator approval runtime implementation task card、不创建 operator approval runtime、不创建 approval executor、不连接 operator identity provider、不执行 ticket / change window verifier、不执行 approval runtime、不连接数据库、不执行 audit write、不创建 credential handle runtime、不创建 backend health runtime、不实现 production resolver runtime、不启用 repository mode 或 public production API。

## Production Secret Backend Audit Store Handoff Readiness

`Production Secret Backend Audit Store Handoff Readiness v1` 已固定 `audit_store_handoff_readiness_defined`，对应切片为 `production-secret-backend-audit-store-handoff-readiness-v1`。它把真实 resolver runtime entry review 中的 audit store handoff blocker 拆成独立静态前置证据，固定 future audit store handoff 的 reference-only handoff envelope、允许 metadata、禁止 payload / secret material、event kind allowlist、secret ref / provider profile / environment / backend profile binding、credential handle boundary、operator approval evidence、rotation / retention / redaction policy、delivery semantics、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。

该专题不创建 audit store、不创建 audit writer、不写 audit event、不连接数据库、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不创建 credential payload 或 credential handle runtime、不执行 no secret leakage smoke runtime，也不启用 repository mode 或 public production API。后续已由 backend health boundary readiness 固定 health reference、metadata 和失败语义，并由 backend health runtime entry review 固定 runtime task card 仍 blocked；真实 resolver runtime implementation task card 仍必须作为后续独立目标评审。

## Production Secret Backend Audit Store Runtime Implementation Entry Review

`Production Secret Backend Audit Store Runtime Implementation Entry Review v1` 已固定 `audit_store_runtime_implementation_entry_review_defined`，对应切片为 `production-secret-backend-audit-store-runtime-implementation-entry-review-v1`。它消费 audit store handoff readiness，并把 audit store runtime task card 的入口结论固定为 `audit_store_runtime_implementation_blocked_before_task_card`。

该专题只固定 entry decision、blocked conditions、future runtime task card requirements、failure mapping、sanitized diagnostics、no fallback、no side effects、artifact guard、real resolver entry review alignment 和 implementation readiness alignment；不创建 audit store runtime implementation task card、不创建 audit store runtime、不创建 audit writer、不写 audit event、不创建 runtime event schema、不执行 delivery、不连接数据库、不执行 approval runtime、不创建 credential handle runtime、不创建 backend health runtime、不实现 production resolver runtime、不启用 repository mode 或 public production API。

## Production Secret Backend Audit Store Contract / Event Schema Readiness

`Production Secret Backend Audit Store Contract / Event Schema Readiness v1` 已固定 `audit_store_contract_event_schema_readiness_defined`，对应切片为 `production-secret-backend-audit-store-contract-event-schema-readiness-v1`。它消费 audit store handoff readiness 与 audit store runtime implementation entry review，固定 future audit store 的 metadata-only event schema、event version、event kind allowlist、required / optional fields、reference-only writer input、idempotency key reference、delivery result envelope、retention / redaction binding、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。

该专题只补齐 audit store runtime 之前的 contract / event schema readiness；不创建 audit store runtime implementation task card、不创建 audit store runtime、不创建 audit writer、不写 audit event、不创建 runtime event schema、不连接数据库、不读取真实 secret、不调用云 secret 服务、不创建 credential payload、不执行 approval runtime、不执行 backend health check、不执行 no leakage smoke runtime、不创建 production resolver runtime、不启用 repository mode 或 public production API。audit store runtime implementation task card 仍必须后续独立评审。

## Production Secret Backend Audit Store Runtime Implementation Entry Refresh

`Production Secret Backend Audit Store Runtime Implementation Entry Refresh v2` 已固定 `audit_store_runtime_implementation_entry_refresh_v2_defined`，对应切片为 `production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2`。它消费 audit store runtime implementation entry review 与 audit store contract / event schema readiness，把已满足的 static contract prerequisites 与仍 blocked 的 store ownership、writer ownership、runtime event schema、delivery runtime、operator approval、credential handle、backend health 和 no leakage runtime 依赖分开。

该专题只刷新 audit store runtime implementation task card 的入口结论；entry decision 仍为 blocked before runtime task card。不创建 audit store runtime implementation task card、不创建 audit store runtime、不创建 audit writer、不写 audit event、不创建 runtime event schema、不连接数据库、不读取真实 secret、不调用云 secret 服务、不创建 credential payload、不执行 approval runtime、不执行 backend health check、不执行 no leakage smoke runtime、不创建 production resolver runtime、不启用 repository mode 或 public production API。

## Production Secret Backend Audit Store Ownership Boundary Readiness

`Production Secret Backend Audit Store Ownership Boundary Readiness v1` 已固定 `audit_store_ownership_boundary_readiness_defined`，对应切片为 `production-secret-backend-audit-store-ownership-boundary-readiness-v1`。它消费 audit store handoff、contract / event schema、runtime entry refresh v2 与相关 runtime entry review，把 store owner、writer owner、runtime event schema owner、delivery / idempotency owner、retention / redaction owner 和 dependency owner reference 收束为静态职责边界。

该专题只补齐 ownership boundary；不创建 audit store runtime implementation task card、不创建 audit store runtime、不创建 audit writer、不写 audit event、不创建 runtime event schema、不执行 delivery、不连接数据库、不读取真实 secret、不调用云 secret 服务、不创建 production resolver runtime、不启用 repository mode 或 public production API。delivery / idempotency runtime boundary 已由后续专题固定；后续若继续 audit store runtime 前置，应重新评审 audit store runtime implementation entry。

## Production Secret Backend Audit Store Delivery / Idempotency Runtime Boundary Readiness

`Production Secret Backend Audit Store Delivery / Idempotency Runtime Boundary Readiness v1` 已固定 `audit_store_delivery_idempotency_runtime_boundary_readiness_defined`，对应切片为 `production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1`。它消费 audit store handoff、contract / event schema、runtime entry refresh v2、ownership boundary 与相关 runtime entry review，把 delivery owner、idempotency key owner、duplicate handling、retry / failure semantics、delivery result envelope 和 metadata-only diagnostics 收束为静态运行前置边界。

该专题只补齐 delivery / idempotency runtime boundary；不创建 audit store runtime implementation task card、不创建 audit store runtime、不创建 audit writer、不写 audit event、不创建 runtime event schema、不执行 delivery、不创建 idempotency runtime、duplicate detector runtime 或 retry runtime、不连接数据库、不读取真实 secret、不调用云 secret 服务、不创建 production resolver runtime、不启用 repository mode 或 public production API。后续若继续 audit store runtime 前置，应重新评审 audit store runtime implementation entry，而不是直接创建 runtime task card。

## Production Secret Backend Audit Store Runtime Implementation Entry Refresh v3 / v4

`Production Secret Backend Audit Store Runtime Implementation Entry Refresh v3` 已固定 `audit_store_runtime_implementation_entry_refresh_v3_defined`；后续 `Production Secret Backend Audit Store Runtime Implementation Entry Refresh v4` 已固定 `audit_store_runtime_implementation_entry_refresh_v4_defined`，对应切片为 `production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4`。它消费 audit store runtime entry refresh v3、durable backend boundary、writer runtime boundary、runtime event schema materialization readiness、delivery runtime readiness、idempotency runtime readiness、credential handle / operator approval / backend health / no leakage refresh 与 implementation readiness，把已满足的静态前置与仍缺失的真实 runtime / durable backend selection / runtime schema artifact / writer / delivery / idempotency / approval / credential handle / backend health / no leakage runtime 依赖分开。

该专题只刷新 audit store runtime implementation task card 的入口结论；entry decision 仍为 still blocked before runtime task card。不创建 audit store runtime implementation task card、不创建 audit store runtime、不创建 audit writer、不写 audit event、不创建 runtime event schema artifact、不执行 delivery、不创建 idempotency runtime、不连接数据库、不读取真实 secret、不调用云 secret 服务、不创建 credential payload、不执行 approval runtime、不执行 backend health check、不执行 no leakage smoke runtime、不创建 production resolver runtime、不启用 repository mode 或 public production API。

## Production Secret Backend Resolver Backend Health Boundary Readiness

`Production Secret Backend Resolver Backend Health Boundary Readiness v1` 已固定 `resolver_backend_health_boundary_readiness_defined`，对应切片为 `production-secret-backend-resolver-backend-health-boundary-readiness-v1`。它把真实 resolver runtime entry review 中的 backend health boundary 拆成独立静态前置证据，固定 future resolver backend health boundary 的 reference-only 输入、profile binding、environment binding、允许 health metadata、禁止 secret material / credential payload / provider raw URL / DSN、health lifecycle、failure mapping、sanitized diagnostics、operator / audit / rotation / no leakage / credential handle 依赖、no fallback、no side effects、artifact guard、entry review alignment 和后续 runtime implementation 拆分。

该专题不创建 backend health runtime、不执行 backend health check、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 credential payload、不创建 credential handle runtime、不执行 approval runtime、不创建 audit store / writer / event，也不启用 repository mode 或 public production API。后续已由 backend health runtime implementation entry review 固定 blocked before task card；真实 resolver runtime implementation task card 仍必须作为后续独立目标评审。

## Production Secret Backend Resolver Backend Health Runtime Implementation Entry Review / Refresh

`Production Secret Backend Resolver Backend Health Runtime Implementation Entry Review v1` 已固定 `resolver_backend_health_runtime_implementation_entry_review_defined`；后续 `Production Secret Backend Resolver Backend Health Runtime Implementation Entry Refresh v1` 已固定 `resolver_backend_health_runtime_implementation_entry_refresh_defined`，对应切片为 `production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1`。

该 refresh 消费原始 entry review、cloud secret service selection readiness、production resolver blocker consolidation、credential handle / operator approval refresh、audit store refresh v3 和 no leakage smoke entry review，entry decision 为 `resolver_backend_health_runtime_task_card_still_blocked_after_refresh`；不创建 backend health runtime implementation task card、不创建 backend health runtime、不执行 backend health check、不访问 provider、不调用云 secret 服务、不创建 credential handle runtime、不执行 approval runtime、不创建 audit store / writer / event、不实现 production resolver runtime、不启用 repository mode 或 public production API。

## 停止线

- 不用平台专题替代 task card；进入代码实现、route、schema、checker 或高风险 gate 时仍必须有实现批次记录。
- 不把 dev-only 能力写成 production ready。
- 不在没有明确触发条件时打开真实数据库、Radish OIDC、repository adapter、store selector、secret backend、executor、confirmation、writeback 或 replay。
