# RadishMind 平台专题入口

更新时间：2026-06-20

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
| Auth / Store Transition | readiness 已有证据 | `docs/contracts/control-plane-read-side.md`、相关 control-plane read task cards | 等真实 OIDC 或 store 迁移入口满足后再拆细专题 |
| Repository Adapter / Store Selector | control-plane readiness 已有证据；saved workflow draft selector 已按 fail-closed mode selection 落地 | `control-plane-read-*repository*`、`*store-selector*` task cards、`docs/features/workflow/saved-workflow-draft-store-selector-implementation-v1.md` | 后续 repository adapter / database 仍需独立实现专题 |
| Provider Runtime & Health | close candidate | [Provider Runtime & Health v1 任务卡](../task-cards/provider-runtime-health-v1-plan.md) | 不继续扩同层 provider 小切片 |
| Production Ops / Deployment | 静态边界已 close；secret ref config、provider profile binding、disabled resolver interface、operator runbook / negative gates、rotation / audit policy readiness、test fixture strategy / fake resolver entry review、fake resolver contract / no secret leakage smoke strategy、fake resolver implementation task card entry readiness、fake resolver implementation task card、fake resolver runtime implementation entry review、test-only fake resolver runtime、真实 resolver runtime preconditions、真实 resolver runtime implementation entry review、resolver backend profile selection readiness、真实 resolver no leakage smoke runtime strategy、credential handle runtime boundary readiness、operator approval runtime evidence readiness、audit store handoff readiness 与 resolver backend health boundary readiness 已实现 | [Production Ops Hardening v1](../task-cards/production-ops-hardening-v1-plan.md)、[Docker Deployment v1](../task-cards/production-ops-docker-deployment-v1-plan.md)、[Production Secret Backend Config / Secret Ref Readiness v1](production-secret-backend-config-secret-ref-readiness-v1.md)、[Production Secret Backend Provider Profile Secret Binding Readiness v1](production-secret-backend-provider-profile-secret-binding-readiness-v1.md)、[Production Secret Backend Secret Resolver Interface Disabled Readiness v1](production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md)、[Production Secret Backend Operator Runbook / Negative Gates Readiness v1](production-secret-backend-operator-runbook-negative-gates-readiness-v1.md)、[Production Secret Backend Rotation / Audit Policy Readiness v1](production-secret-backend-rotation-audit-policy-readiness-v1.md)、[Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review v1](production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.md)、[Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy v1](production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md)、[Production Secret Backend Fake Resolver Implementation Task Card Entry Readiness Review v1](production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.md)、[Production Secret Backend Fake Resolver Implementation v1](production-secret-backend-fake-resolver-implementation-v1.md)、[Production Secret Backend Fake Resolver Runtime Implementation Entry Review v1](production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.md)、[Production Secret Backend Fake Resolver Runtime Implementation v1](production-secret-backend-fake-resolver-runtime-implementation-v1.md)、[Production Secret Backend Real Resolver Runtime Preconditions v1](production-secret-backend-real-resolver-runtime-preconditions-v1.md)、[Production Secret Backend Real Resolver Runtime Implementation Entry Review v1](production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.md)、[Production Secret Backend Resolver Backend Profile Selection Readiness v1](production-secret-backend-resolver-backend-profile-selection-readiness-v1.md)、[Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Strategy v1](production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.md)、[Production Secret Backend Credential Handle Runtime Boundary Readiness v1](production-secret-backend-credential-handle-runtime-boundary-readiness-v1.md)、[Production Secret Backend Operator Approval Runtime Evidence Readiness v1](production-secret-backend-operator-approval-runtime-evidence-readiness-v1.md)、[Production Secret Backend Audit Store Handoff Readiness v1](production-secret-backend-audit-store-handoff-readiness-v1.md)、[Production Secret Backend Resolver Backend Health Boundary Readiness v1](production-secret-backend-resolver-backend-health-boundary-readiness-v1.md) | 下一步若继续 secret backend，应选择 backend health runtime implementation entry review 或 real resolver runtime implementation entry refresh；真实 resolver runtime 仍需独立任务 |
| Dev-only Write Path | implemented | [Dev-only Saved Draft Consumer](../features/workflow/dev-only-saved-draft-consumer.md) | 已服务 saved draft consumer integration；后续按 conflict UI / smoke / contract 固化拆批次 |

## Saved Workflow Draft Store Selector

Saved workflow draft 的平台服务配置已经新增 `workflow_saved_draft_store` / `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE`。当前唯一可成功读写的 mode 是 `memory_dev`，它只服务 dev-only saved draft route；`repository_disabled` 和 `repository` 都返回 `repository_store_disabled`，unknown mode 返回 `invalid_draft_store_mode`。这些失败必须直接暴露给 save / read / list 调用方，不得回退到 memory dev、sample 或 fixture。

`services/platform/migrations/workflow_saved_drafts/` 当前只承载 `manifest.json`、`ddl-review.md`、`rollback-evidence.json` 和 `migration-smoke.json` 四个静态 schema artifact 证据文件。它们说明 future durable store 的 logical schema、predicate、review 和 rollback 边界，不是 SQL migration，不会被 platform service 自动执行，也不表示 repository mode、真实数据库或 production auth 已可用。

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

该专题确认真实 resolver runtime implementation task card 当时仍 blocked。后续已由 resolver backend profile selection readiness、no leakage smoke runtime strategy、credential handle runtime boundary readiness、operator approval runtime evidence readiness、audit store handoff readiness 和 resolver backend health boundary readiness 固定静态前置；当前仍不在本批创建 production resolver runtime implementation task card，不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不连接数据库，不创建 credential payload，也不启用 repository mode 或 public production API。

## Production Secret Backend Resolver Backend Profile Selection Readiness

`Production Secret Backend Resolver Backend Profile Selection Readiness v1` 已固定 `resolver_backend_profile_selection_readiness_defined`，对应切片为 `production-secret-backend-resolver-backend-profile-selection-readiness-v1`。它把真实 resolver runtime entry review 中的 backend profile selection blocker 拆成独立静态前置证据，固定 backend profile shape、reserved backend kind allowlist、environment / provider profile / policy version binding、operator approval、audit / rotation dependency、backend health reference、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。

该专题不创建 backend runtime、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 credential payload 或 credential handle runtime，也不启用 repository mode 或 public production API。真实 resolver no leakage smoke runtime strategy、credential handle runtime boundary readiness、operator approval runtime evidence readiness、audit store handoff readiness 和 backend health boundary readiness 已由后续专题固定；真实 resolver runtime implementation task card 仍必须作为后续独立目标评审。

## Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Strategy

`Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Strategy v1` 已固定 `real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`，对应切片为 `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1`。它把真实 resolver runtime entry review 中的 no-secret-leakage smoke runtime strategy blocker 拆成独立静态前置证据，固定 future smoke runtime gate 的离线策略、扫描面、输入 / 输出 allowlist、负向 probe、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。

该专题不创建 no secret leakage smoke runtime、不执行 resolver smoke、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 credential payload 或 credential handle runtime，也不启用 repository mode 或 public production API。后续已固定 credential handle runtime boundary readiness、operator approval runtime evidence readiness、audit store handoff readiness 与 backend health boundary readiness；真实 resolver runtime implementation task card 仍必须作为后续独立目标评审。

## Production Secret Backend Credential Handle Runtime Boundary Readiness

`Production Secret Backend Credential Handle Runtime Boundary Readiness v1` 已固定 `credential_handle_runtime_boundary_readiness_defined`，对应切片为 `production-secret-backend-credential-handle-runtime-boundary-readiness-v1`。它把真实 resolver runtime entry review 中的 credential handle blocker 拆成独立静态前置证据，固定 future opaque credential handle 的 reference 定义、允许 metadata、禁止 payload / secret material、secret ref / provider profile / environment binding 前置、operator approval / audit / rotation 依赖、handle lifecycle、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。

该专题不创建 credential handle runtime、不创建 credential payload、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 no secret leakage smoke runtime，也不启用 repository mode 或 public production API。后续已固定 operator approval runtime evidence readiness、audit store handoff readiness 与 backend health boundary readiness；真实 resolver runtime implementation task card 仍必须作为后续独立目标评审。

## Production Secret Backend Operator Approval Runtime Evidence Readiness

`Production Secret Backend Operator Approval Runtime Evidence Readiness v1` 已固定 `operator_approval_runtime_evidence_readiness_defined`，对应切片为 `production-secret-backend-operator-approval-runtime-evidence-readiness-v1`。它把真实 resolver runtime entry review 中的 operator approval runtime evidence blocker 拆成独立静态前置证据，固定 future operator approval runtime evidence 的 evidence shape、允许 metadata、禁止 payload / secret material、approval subject binding、secret ref / provider profile / environment / backend profile binding、credential handle boundary、no leakage strategy、operator identity、dual control、ticket / change window、audit / rotation dependency、lifecycle、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。

该专题不创建 operator approval runtime、不执行 approval runtime、不创建 approval executor、不写 audit store、不创建 credential handle runtime、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不执行 no secret leakage smoke runtime，也不启用 repository mode 或 public production API。后续已固定 audit store handoff readiness 与 backend health boundary readiness；真实 resolver runtime implementation task card 仍必须作为后续独立目标评审。

## Production Secret Backend Audit Store Handoff Readiness

`Production Secret Backend Audit Store Handoff Readiness v1` 已固定 `audit_store_handoff_readiness_defined`，对应切片为 `production-secret-backend-audit-store-handoff-readiness-v1`。它把真实 resolver runtime entry review 中的 audit store handoff blocker 拆成独立静态前置证据，固定 future audit store handoff 的 reference-only handoff envelope、允许 metadata、禁止 payload / secret material、event kind allowlist、secret ref / provider profile / environment / backend profile binding、credential handle boundary、operator approval evidence、rotation / retention / redaction policy、delivery semantics、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。

该专题不创建 audit store、不创建 audit writer、不写 audit event、不连接数据库、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不创建 credential payload 或 credential handle runtime、不执行 no secret leakage smoke runtime，也不启用 repository mode 或 public production API。后续已由 backend health boundary readiness 固定 health reference、metadata 和失败语义；真实 resolver runtime implementation task card 仍必须作为后续独立目标评审。

## Production Secret Backend Resolver Backend Health Boundary Readiness

`Production Secret Backend Resolver Backend Health Boundary Readiness v1` 已固定 `resolver_backend_health_boundary_readiness_defined`，对应切片为 `production-secret-backend-resolver-backend-health-boundary-readiness-v1`。它把真实 resolver runtime entry review 中的 backend health boundary 拆成独立静态前置证据，固定 future resolver backend health boundary 的 reference-only 输入、profile binding、environment binding、允许 health metadata、禁止 secret material / credential payload / provider raw URL / DSN、health lifecycle、failure mapping、sanitized diagnostics、operator / audit / rotation / no leakage / credential handle 依赖、no fallback、no side effects、artifact guard、entry review alignment 和后续 runtime implementation 拆分。

该专题不创建 backend health runtime、不执行 backend health check、不实现 production resolver runtime、不读取真实 secret、不调用云 secret 服务、不连接数据库、不创建 credential payload、不创建 credential handle runtime、不执行 approval runtime、不创建 audit store / writer / event，也不启用 repository mode 或 public production API。下一步若继续 production secret backend，应单独推进 backend health runtime implementation entry review 或 real resolver runtime implementation entry refresh。

## 停止线

- 不用平台专题替代 task card；进入代码实现、route、schema、checker 或高风险 gate 时仍必须有实现批次记录。
- 不把 dev-only 能力写成 production ready。
- 不在没有明确触发条件时打开真实数据库、Radish OIDC、repository adapter、store selector、secret backend、executor、confirmation、writeback 或 replay。
