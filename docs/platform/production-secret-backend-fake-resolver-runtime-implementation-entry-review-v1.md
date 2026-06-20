# Production Secret Backend Fake Resolver Runtime Implementation Entry Review v1

更新时间：2026-06-20

## 文档目的

本文档评审 production secret backend 的 fake resolver implementation 静态任务卡是否已经具备进入 test-only runtime implementation 的条件。

结论：状态为 `fake_resolver_runtime_implementation_entry_review_defined`，entry decision 为 `fake_resolver_runtime_implementation_ready_for_next_task`。现有 fake resolver implementation task card 已覆盖 disabled-by-default gate、placeholder secret ref fixture、environment binding、opaque test credential handle metadata、sanitized diagnostics runtime emission、no secret leakage runtime smoke、offline fast baseline 和 artifact guard 的后续实现要求，因此下一步可以单独创建 runtime implementation 任务卡。

本批仍不创建 fake resolver runtime，不创建 no secret leakage smoke runtime，不实现 resolver runtime，不解析 secret，不读取真实环境 secret，不创建 credential handle runtime，不连接数据库，不调用云 secret 服务，不启用 workflow saved draft repository mode 或 production API。

## 输入证据

- `production-secret-backend-fake-resolver-implementation-v1` 已创建 fake resolver implementation 静态任务卡，并把 runtime implementation requirements 固定为可检查证据。
- `production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1` 已固定 fake resolver input / output allowlist、禁止字段和 no secret leakage smoke strategy。
- `production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1` 已固定 `test-fixture-strategy` 的 blocked 事实和 failure mapping。
- `production-ops-secret-backend-implementation-readiness` 显示 fake resolver runtime、resolver runtime 和 no secret leakage smoke runtime 仍为 `not_created`。
- `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 仍确认 saved draft database secret resolver implementation 不打开。

## Entry Review Decision

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| implementation task card | `satisfied_static_task_card` | 已定义 runtime scope、disabled gate、fixture shape、diagnostics boundary 和 no leakage smoke plan |
| disabled-by-default gate | `satisfied_for_runtime_entry` | runtime 后续必须默认 disabled，只能由 test-only gate 显式启用 |
| placeholder secret ref fixture | `satisfied_for_runtime_entry` | 输入只能使用 placeholder `secret_ref_key` 和 version metadata，不允许 secret value |
| environment binding | `satisfied_for_runtime_entry` | runtime smoke 必须验证 test environment binding，不能跨环境 fallback |
| opaque handle metadata | `satisfied_for_runtime_entry` | 成功结果只能返回 opaque test credential handle metadata，不返回 payload |
| sanitized diagnostics runtime emission | `satisfied_for_runtime_entry` | diagnostics 只能包含 failure code、request id、audit ref 和 policy version 等脱敏字段 |
| no secret leakage runtime smoke | `satisfied_for_runtime_entry` | 下一批必须实现离线 smoke，扫描 fixture、runtime response、diagnostics 和 smoke record |
| fake resolver runtime | `ready_for_next_task_not_created` | 可以进入下一张 runtime implementation 任务卡，但本批不创建 runtime |
| resolver runtime | `forbidden` | production resolver runtime 仍不打开 |
| database / repository mode | `blocked` | DB provider、SQL、schema marker、repository mode 和 production API 均不打开 |

## 下一张任务卡必须覆盖

下一张 fake resolver runtime implementation task card 如果创建，必须只覆盖以下 test-only runtime 边界：

- 创建默认 disabled 的 fake resolver runtime gate，且只允许测试 fixture / smoke 明确启用。
- 创建 placeholder secret ref fixture 与 environment binding 校验。
- 创建 opaque test credential handle metadata 结果类型，但不得包含 credential payload。
- 创建 sanitized diagnostics runtime emission，字段必须受 allowlist 约束。
- 创建离线 no secret leakage runtime smoke，扫描 runtime response、diagnostics、fixture 和 smoke record。
- 创建 side effect counters，并确保 provider call、cloud secret call、database connection、driver open、SQL execution 和 repository mode enablement 均为零。

下一张任务卡不得合入 production resolver runtime、云 secret service、真实 credential、database connection provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode runtime、production secret audit store 或 public production API。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `fake_resolver_runtime_entry_task_card_missing` | `implementation_entry` | fake resolver implementation 静态任务卡不存在或未被消费 |
| `fake_resolver_runtime_entry_disabled_gate_missing` | `implementation_gate` | 后续 runtime scope 未定义 disabled-by-default gate |
| `fake_resolver_runtime_entry_placeholder_fixture_missing` | `test_fixture` | 后续 runtime scope 未定义 placeholder secret ref fixture |
| `fake_resolver_runtime_entry_environment_binding_missing` | `environment_binding` | 后续 runtime scope 未定义 environment binding |
| `fake_resolver_runtime_entry_opaque_handle_metadata_missing` | `credential_boundary` | 后续 runtime scope 未定义 opaque test credential handle metadata 边界 |
| `fake_resolver_runtime_entry_diagnostics_boundary_missing` | `sanitized_diagnostics` | 后续 runtime scope 未定义 runtime diagnostics allowlist |
| `fake_resolver_runtime_entry_no_leakage_smoke_missing` | `no_secret_leakage` | 后续 runtime scope 未定义离线 no secret leakage smoke |
| `fake_resolver_runtime_entry_secret_value_detected` | `artifact_guard` | 文档、fixture、checker 或任务卡出现 secret-looking value |
| `fake_resolver_runtime_created_in_entry_review` | `artifact_guard` | 本批创建 fake resolver runtime、resolver runtime 或 smoke runner |
| `fake_resolver_runtime_entry_cloud_call_forbidden` | `no_side_effects` | checker、fixture 或任务卡要求联网、provider call 或云 secret call |
| `fake_resolver_runtime_entry_repository_mode_forbidden` | `repository_mode` | fake resolver 被用于打开 workflow saved draft repository mode 成功路径 |
| `fake_resolver_runtime_entry_scope_overreach` | `implementation_boundary` | 本批把 DB provider、SQL、schema marker、audit store 或 production API 合入 runtime entry review |

所有失败都必须 fail closed，不返回 secret value、DSN、provider raw URL、database hostname、database error detail、credential payload、完整 secret ref value、完整用户 claim 或草案主体。

## Sanitized Diagnostics

允许输出：

- `fake_resolver_runtime_entry_status`
- `fake_resolver_implementation_task_card_status`
- `fake_resolver_contract_status`
- `no_secret_leakage_smoke_strategy_status`
- `test_fixture_strategy_status`
- `resolver_state`
- `fake_resolver_runtime_status`
- `no_secret_leakage_smoke_runtime_status`
- `secret_ref_status`
- `environment_binding_status`
- `failure_code`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、full secret ref value 或完整用户 claim。本批不允许 runtime emission；所有 diagnostics 只存在于 committed fixture / checker 的静态评审语义中。

## No Fallback

- 不允许 fake resolver fallback 到 production resolver。
- 不允许 production resolver fallback 到 fake resolver。
- 不允许 test secret ref fallback 到 production secret ref。
- 不允许 resolver failure fallback 到 committed value、developer env plaintext、fixture credential、mock provider、local-smoke profile、fake query executor 或 sample。
- 不允许 runtime entry review 被解释为 credential resolved、production secret backend ready、cloud secret backend ready、repository mode ready 或 production ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema、已有 readiness fixture 和 `check-repo.py` 注册顺序，不读取真实环境变量、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 resolver、不创建 credential handle、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不调用 production API。

side effect counters 必须保持：

- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `secret_value_read_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `credential_handle_created_count=0`
- `database_connection_count=0`
- `driver_open_count=0`
- `sql_execution_count=0`
- `schema_marker_read_count=0`
- `schema_marker_write_count=0`
- `repository_mode_enablement_count=0`
- `production_api_call_count=0`

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.md`
- `docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.json`
- `scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py`

不得新增或启用以下 artifact：

- fake resolver runtime
- resolver runtime
- no secret leakage smoke runtime
- no secret leakage smoke fixture
- fake resolver runtime implementation task card
- cloud secret SDK
- production credential file
- secret value fixture
- credential handle runtime
- database connection provider
- DB driver / DSN parser
- connection factory
- SQL migration
- schema marker reader / writer
- migration runner
- workflow saved draft repository mode runtime
- production secret audit store
- public production API

## 后续推进

后续已创建并完成 `Production Secret Backend Fake Resolver Runtime Implementation v1` 的独立实现任务卡，test-only fake resolver runtime 已由 Go 包 `services/platform/internal/secretbackend` 和 Go 单测固定。任何真实 resolver runtime、backend health runtime、DB provider、cloud secret service、repository mode、audit store runtime 或 production API 都仍需独立评审。

`test-fixture-strategy` 在后续 test-only runtime 中只满足离线替身需求；它仍不满足 production resolver runtime、no secret leakage smoke runtime、backend health runtime、credential handle runtime、approval runtime、audit store runtime、DB provider 或 repository mode 的启用条件。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
